// Package s2s is a client implementation of the Splunk to Splunk protocol in Golang
package s2s

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"encoding/binary"
)

var bp sync.Pool

func init() {
	bp = sync.Pool{
		New: func() interface{} {
			bb := bytes.NewBuffer([]byte{})
			return bb
		},
	}
}

// S2S sends data to Splunk using the Splunk to Splunk protocol
type S2S struct {
	buf                *bufio.Writer
	conn               net.Conn
	initialized        bool
	endpoint           string
	endpoints          []string
	closed             bool
	bufferBytes        int
	tls                bool
	cert               string
	serverName         string
	insecureSkipVerify bool
	rebalanceInterval  int
	lastConnectTime    time.Time
	maxIdleTime        int
	lastSendTime       time.Time
	mutex              *sync.RWMutex
	ignoreNextClosed   bool
}

type splunkSignature struct {
	signature  [128]byte
	serverName [256]byte
	mgmtPort   [16]byte
}

// Interface is the client interface definition
type Interface interface {
	Send(event map[string]string) (int64, error)
}

/*
NewS2S will initialize S2S

endpoints is a list of endpoint strings, which should be in the format of host:port

bufferBytes is the max size of the buffer before flushing
*/
func NewS2S(endpoints []string, bufferBytes int) (*S2S, error) {
	return NewS2STLS(endpoints, bufferBytes, false, "", "", false)
}

/*
NewS2STLS will initialize S2S for TLS

endpoints is a list of endpoint strings, which should be in the format of host:port

bufferBytes is the max size of the buffer before flushing

tls specifies whether to connect with TLS or not

cert is a valid root CA we should use for verifying the server cert

serverName is the name specified in your certificate, will default to "SplunkServerDefaultCert",

insecureSkipVerify specifies whether to skip verification of the server certificate
*/
func NewS2STLS(endpoints []string, bufferBytes int, tls bool, cert string, serverName string, insecureSkipVerify bool) (*S2S, error) {
	st := new(S2S)

	if len(endpoints) < 1 {
		return nil, fmt.Errorf("No endpoints specified")
	}

	st.mutex = &sync.RWMutex{}
	st.endpoints = endpoints
	st.bufferBytes = bufferBytes
	st.tls = tls
	st.cert = cert
	if serverName == "" {
		st.serverName = "SplunkServerDefaultCert"
	} else {
		st.serverName = serverName
	}
	st.insecureSkipVerify = insecureSkipVerify

	err := st.reconnect(true)
	if err != nil {
		return nil, err
	}
	st.rebalanceInterval = 300
	st.maxIdleTime = 15
	st.lastSendTime = time.Now()
	st.lastConnectTime = time.Now()
	st.initialized = true
	return st, nil
}

// Connect opens a connection to Splunk
// endpoint is the format of 'host:port'
func (st *S2S) connect(endpoint string) error {
	st.mutex.Lock()
	var err error
	st.conn, err = net.DialTimeout("tcp", endpoint, 2*time.Second)
	if err != nil {
		return err
	}
	st.buf = bufio.NewWriterSize(st.conn, st.bufferBytes)
	if st.tls {
		config := &tls.Config{
			InsecureSkipVerify: st.insecureSkipVerify,
			ServerName:         st.serverName,
		}
		if len(st.cert) > 0 {
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM([]byte(st.cert))
			if !ok {
				return fmt.Errorf("Failed to parse root certificate")
			}
			config.RootCAs = roots
		}

		st.conn = tls.Client(st.conn, config)
	}
	st.mutex.Unlock()
	err = st.sendSig()
	if err != nil {
		return err
	}
	go st.readAndDiscard()
	st.lastConnectTime = time.Now()
	st.lastSendTime = time.Now()
	return err
}

// SetRebalanceInterval sets the interval to reconnect to a new random endpoint
// Defaults to 30 seconds
func (st *S2S) SetRebalanceInterval(interval int) {
	st.rebalanceInterval = interval
}

// sendSig will write the signature to the connection if it has not already been written
// Create Signature element of the S2S Message.  Signature is C struct:
//
// struct S2S_Signature
// {
// 	char _signature[128];
// 	char _serverName[256];
// 	char _mgmtPort[16];
// };
func (st *S2S) sendSig() error {
	st.mutex.RLock()
	defer st.mutex.RUnlock()
	endpointParts := strings.Split(st.endpoint, ":")
	if len(endpointParts) != 2 {
		return fmt.Errorf("Endpoint malformed.  Should look like server:port")
	}
	serverName := endpointParts[0]
	mgmtPort := endpointParts[1]
	var sig splunkSignature
	copy(sig.signature[:], "--splunk-cooked-mode-v2--")
	copy(sig.serverName[:], serverName)
	copy(sig.mgmtPort[:], mgmtPort)
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, sig.signature)
	binary.Write(buf, binary.BigEndian, sig.serverName)
	binary.Write(buf, binary.BigEndian, sig.mgmtPort)
	st.buf.Write(buf.Bytes())
	return nil
}

// encodeString encodes a string to be sent across the wire to Splunk
// Wire protocol has an unsigned integer of the length of the string followed
// by a null terminated string.
func encodeString(tosend string, buf *bytes.Buffer) {
	l := uint32(len(tosend) + 1)
	binary.Write(buf, binary.BigEndian, l)
	binary.Write(buf, binary.BigEndian, []byte(tosend))
	binary.Write(buf, binary.BigEndian, []byte{0})
}

// encodeKeyValue encodes a key/value pair to send across the wire to splunk
// A key value pair is merely a concatenated set of encoded strings.
func encodeKeyValue(key, value string, buf *bytes.Buffer) {
	encodeString(key, buf)
	encodeString(value, buf)
}

// EncodeEvent encodes a full Splunk event
func EncodeEvent(line map[string]string) []byte {
	buf := bp.Get().(*bytes.Buffer)
	defer bp.Put(buf)
	buf.Reset()
	binary.Write(buf, binary.BigEndian, uint32(0)) // Message Size
	binary.Write(buf, binary.BigEndian, uint32(0)) // Map Count

	// These fields should be present in every event
	_time := line["_time"]
	host := line["host"]
	source := line["source"]
	sourcetype := line["sourcetype"]
	index := line["index"]

	// These are optional
	channel, hasChannel := line["_channel"]
	conf, hasConf := line["_conf"]
	_, hasLineBreaker := line["_linebreaker"]
	_, hasDone := line["_done"]
	var indexFields string

	// Check time for subseconds
	if strings.ContainsRune(_time, '.') {
		timeparts := strings.Split(_time, ".")
		_time = timeparts[0]
		indexFields += "_subsecond::" + timeparts[1] + " "
	}

	for k, v := range line {
		switch k {
		case "source", "sourcetype", "host", "index", "_raw", "_time", "_channel", "_conf", "_linebreaker", "_done":
			break
		default:
			indexFields += k + "::" + v + " "
		}
	}

	maps := 7
	encodeKeyValue("_raw", line["_raw"], buf)
	if len(indexFields) > 0 {
		indexFields = strings.TrimRight(indexFields, " ")
		encodeKeyValue("_meta", indexFields, buf)
		maps++
	}
	if hasDone {
		encodeKeyValue("_done", "_done", buf)
		maps++
	}
	if hasLineBreaker {
		encodeKeyValue("_linebreaker", "_linebreaker", buf)
		maps++
	}
	encodeKeyValue("_hpn", "_hpn", buf)
	encodeKeyValue("_time", _time, buf)
	if hasConf {
		encodeKeyValue("_conf", conf, buf)
		maps++
	}
	encodeKeyValue("MetaData:Source", "source::"+source, buf)
	encodeKeyValue("MetaData:Host", "host::"+host, buf)
	encodeKeyValue("MetaData:Sourcetype", "sourcetype::"+sourcetype, buf)
	if hasChannel {
		encodeKeyValue("_channel", channel, buf)
		maps++
	}
	encodeKeyValue("_MetaData:Index", index, buf)

	binary.Write(buf, binary.BigEndian, uint32(0)) // Null terminate raw
	encodeString("_raw", buf)                      // Raw trailer

	ret := buf.Bytes()

	binary.BigEndian.PutUint32(ret, uint32(len(ret)-4)) // Don't include null terminator in message size
	binary.BigEndian.PutUint32(ret[4:], uint32(maps))   // Include extra map for _done key and one for _raw

	return ret
}

// Send sends an event to Splunk, represented as a map[string]string containing keys of index, host, source, sourcetype, and _raw.
// It is a convenience function, wrapping EncodeEvent and Copy
func (st *S2S) Send(event map[string]string) (int64, error) {
	return st.Copy(bytes.NewBuffer(EncodeEvent(event)))
}

// Copy takes a io.Reader and copies it to Splunk, needs to be encoded by EncodeEvent
func (st *S2S) Copy(r io.Reader) (int64, error) {
	if st.closed {
		return 0, fmt.Errorf("cannot send on closed connection")
	}
	if time.Now().Sub(st.lastSendTime) > time.Duration(st.maxIdleTime)*time.Second {
		st.reconnect(true)
	}

	bytes, err := io.Copy(st.buf, r)
	if err != nil {
		return bytes, err
	}

	st.lastSendTime = time.Now()
	st.reconnect(false)
	return bytes, nil
}

// Close disconnects from Splunk
func (st *S2S) Close() error {
	if !st.closed {
		err := st.close()
		if err != nil {
			return err
		}
		st.closed = true
	}
	return nil
}

func (st *S2S) close() error {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	err := st.buf.Flush()
	if err != nil {
		return err
	}
	err = st.conn.Close()
	if err != nil {
		return err
	}
	st.ignoreNextClosed = true
	return nil
}

func (st *S2S) reconnect(force bool) error {
	if (len(st.endpoints) > 1 && time.Now().Sub(st.lastConnectTime) > time.Duration(st.rebalanceInterval)*time.Second) || force {
		st.endpoint = st.endpoints[rand.Intn(len(st.endpoints))]
		if st.conn != nil {
			err := st.close()
			if err != nil {
				return err
			}
		}
		err := st.connect(st.endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func (st *S2S) readAndDiscard() {
	// Attempt to read from connection to see if it's closed
	tmp := make([]byte, 0, 4096)
	for {
		err := st.conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		if err != nil {
			// TODO Maybe infinite loop here?
			st.reconnect(true)
		}
		rbytes, err := st.conn.Read(tmp)
		if err != nil {
			st.mutex.RLock()
			if !(strings.Contains(err.Error(), "use of closed network connection") && st.ignoreNextClosed) {
				st.mutex.RUnlock()
				st.mutex.Lock()
				st.ignoreNextClosed = false
				st.mutex.Unlock()
				st.reconnect(true)
			} else {
				st.mutex.RUnlock()
			}
			break
		// If we have no data, sleep for 100ms before trying again. We're not doing anything with this data, so fuck it.
		// This is a result of CPU thrashing waiting for reads. There is no non-blocking way to read data in Go using net.Conn.
		// Blocking reads thrash the CPU. See: https://github.com/golang/go/issues/27315
		} else if rbytes == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}
