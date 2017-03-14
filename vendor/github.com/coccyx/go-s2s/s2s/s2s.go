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
	"syscall"
)

// S2S sends data to Splunk using the Splunk to Splunk protocol
type S2S struct {
	buf                *bufio.Writer
	conn               net.Conn
	initialized        bool
	endpoint           string
	endpoints          []string
	closed             bool
	sent               int64
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

	err := st.newBuf(false)
	if err != nil {
		return nil, err
	}
	err = st.sendSig()
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
	var err error
	st.conn, err = net.DialTimeout("tcp", endpoint, 2*time.Second)
	if err != nil {
		return err
	}
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

		st.mutex.Lock()
		st.conn = tls.Client(st.conn, config)
		st.mutex.Unlock()
	}
	go st.readAndDiscard()
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
func encodeString(tosend string) []byte {
	// buf := bp.Get().(*bytes.Buffer)
	// defer bp.Put(buf)
	// buf.Reset()
	buf := &bytes.Buffer{}
	l := uint32(len(tosend) + 1)
	binary.Write(buf, binary.BigEndian, l)
	binary.Write(buf, binary.BigEndian, []byte(tosend))
	binary.Write(buf, binary.BigEndian, []byte{0})
	return buf.Bytes()
}

// encodeKeyValue encodes a key/value pair to send across the wire to splunk
// A key value pair is merely a concatenated set of encoded strings.
func encodeKeyValue(key, value string) []byte {
	// buf := bp.Get().(*bytes.Buffer)
	// defer bp.Put(buf)
	// buf.Reset()
	buf := &bytes.Buffer{}
	buf.Write(encodeString(key))
	buf.Write(encodeString(value))
	return buf.Bytes()
}

// EncodeEvent encodes a full Splunk event
func EncodeEvent(line map[string]string) (buf *bytes.Buffer) {
	// buf := bp.Get().(*bytes.Buffer)
	// defer bp.Put(buf)
	// buf.Reset()
	buf = &bytes.Buffer{}

	var msgSize uint32
	msgSize = 8 // Two unsigned 32 bit integers included, the number of maps and a 0 between the end of raw the _raw trailer
	maps := make([][]byte, 0)

	var indexFields string
	for k, v := range line {
		switch k {
		case "source":
			encodedSource := encodeKeyValue("MetaData:Source", "source::"+v)
			maps = append(maps, encodedSource)
			msgSize += uint32(len(encodedSource))
		case "sourcetype":
			encodedSourcetype := encodeKeyValue("MetaData:Sourcetype", "sourcetype::"+v)
			maps = append(maps, encodedSourcetype)
			msgSize += uint32(len(encodedSourcetype))
		case "host":
			encodedHost := encodeKeyValue("MetaData:Host", "host::"+v)
			maps = append(maps, encodedHost)
			msgSize += uint32(len(encodedHost))
		case "index":
			encodedIndex := encodeKeyValue("_MetaData:Index", v)
			maps = append(maps, encodedIndex)
			msgSize += uint32(len(encodedIndex))
		case "_raw":
			break
		case "_time":
			encoded := encodeKeyValue(k, v)
			maps = append(maps, encoded)
			msgSize += uint32(len(encoded))
			if strings.ContainsRune(v, '.') {
				subsecs := "." + strings.Split(v, ".")[1]
				indexFields += "_subsecond::" + subsecs + " "
			}
		default:
			indexFields += k + "::" + v + " "
		}
	}

	if len(indexFields) > 0 {
		indexFields = strings.TrimRight(indexFields, " ")
		encoded := encodeKeyValue("_meta", indexFields)
		maps = append(maps, encoded)
		msgSize += uint32(len(encoded))
	}

	encodedRaw := encodeKeyValue("_raw", line["_raw"])
	msgSize += uint32(len(encodedRaw))
	encodedRawTrailer := encodeString("_raw")
	msgSize += uint32(len(encodedRawTrailer))
	encodedDone := encodeKeyValue("_done", "_done")
	msgSize += uint32(len(encodedDone))

	binary.Write(buf, binary.BigEndian, msgSize)
	binary.Write(buf, binary.BigEndian, uint32(len(maps)+2)) // Include extra map for _done key and one for _raw
	for _, m := range maps {
		binary.Write(buf, binary.BigEndian, m)
	}
	binary.Write(buf, binary.BigEndian, encodedDone)
	binary.Write(buf, binary.BigEndian, encodedRaw)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, encodedRawTrailer)

	return buf
}

// Send sends an event to Splunk, represented as a map[string]string containing keys of index, host, source, sourcetype, and _raw.
// It is a convenience function, wrapping EncodeEvent and Copy
func (st *S2S) Send(event map[string]string) (int64, error) {
	return st.Copy(EncodeEvent(event))
}

// Copy takes a io.Reader and copies it to Splunk, needs to be encoded by EncodeEvent
func (st *S2S) Copy(r io.Reader) (int64, error) {
	if st.closed {
		return 0, fmt.Errorf("cannot send on closed connection")
	}
	if time.Now().Sub(st.lastSendTime) > time.Duration(st.maxIdleTime)*time.Second {
		st.newBuf(true)
	}
	buf := &bytes.Buffer{}
	io.Copy(buf, r)

	bytes, err := io.Copy(st.buf, buf)
	if err != nil {
		// Catch closed pipe error, resend
		switch e := err.(type) {
		case *net.OpError:
			if e.Err == syscall.EPIPE {
				err = st.newBuf(true)
				if err != nil {
					return 0, err
				}
				bytes, err = io.Copy(st.buf, buf)
				if err != nil {
					return 0, err
				}
			}
		default:
			return 0, err
		}
	}

	st.sent += bytes
	if st.sent > int64(st.bufferBytes) {
		st.mutex.RLock()
		err := st.buf.Flush()
		st.mutex.RUnlock()
		if err != nil {
			return 0, err
		}
		st.newBuf(false)
		st.sent = 0
		st.lastSendTime = time.Now()
	}
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
	return nil
}

func (st *S2S) newBuf(force bool) error {
	if time.Now().Sub(st.lastConnectTime) > time.Duration(st.rebalanceInterval)*time.Second || force {
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
	st.buf = bufio.NewWriter(st.conn)
	st.lastConnectTime = time.Now()
	return nil
}

func (st *S2S) readAndDiscard() {
	// Attempt to read from connection to see if it's closed
	// err := st.conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	// err := st.conn.SetReadDeadline(time.Time{})
	for {
		err := st.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			st.newBuf(true)
			break
		}
		one := []byte{}
		_, err = st.conn.Read(one)
		if err != nil {
			st.newBuf(true)
			break
		}
	}
}
