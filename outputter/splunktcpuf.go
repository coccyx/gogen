package outputter

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/coccyx/go-s2s/s2s"
	config "github.com/coccyx/gogen/internal"
)

type splunktcpuf struct {
	initialized bool
	closed      bool
	done        chan int
	s2s         *s2s.S2S
	bufs        map[string]*ufbuf
}

type ufbuf struct {
	channel        string
	buf            *bytes.Buffer
	event          map[string]string
	events         int64
	lastSampleName string
}

func (st *splunktcpuf) Send(item *config.OutQueueItem) error {
	var err error
	if st.initialized == false {
		st.s2s, err = s2s.NewS2S(item.S.Output.Endpoints, item.S.Output.BufferBytes)
		if err != nil {
			return err
		}
		st.initialized = true
		st.bufs = make(map[string]*ufbuf)
	}
	for _, event := range item.Events {
		// Get _channel from event
		channel, ok := event["_channel"]
		if !ok {
			return fmt.Errorf("missing _channel from event")
		}
		var buf *ufbuf
		// Create new buffer for the channel if it doesn't exist
		if buf, ok = st.bufs[channel]; !ok {
			buf = &ufbuf{
				channel: channel,
				buf:     &bytes.Buffer{},
				event:   map[string]string{},
				events:  0,
			}
			st.bufs[channel] = buf
			for k, v := range event {
				if k != "_raw" {
					st.bufs[channel].event[k] = v
				}
			}
		}
		// Copy event bytes to the buffer
		_, err = io.WriteString(buf.buf, event["_raw"])
		if err != nil {
			return err
		}
		_, err = io.WriteString(buf.buf, "\n")
		if err != nil {
			return err
		}
		buf.events++

		buf.lastSampleName = item.S.Name

		// log.Infof("channel: %s bufLen: %d", channel, buf.buf.Len())

		// If over buffer byte length, flush
		if buf.buf.Len() > item.S.Output.BufferBytes {
			err = st.Flush(buf)
		}
	}
	return err
}

func (st *splunktcpuf) Flush(buf *ufbuf) error {
	buf.event["_raw"] = buf.buf.String()
	bytes, err := st.s2s.Send(buf.event)
	if err != nil {
		return err
	}
	Account(buf.events, bytes, buf.lastSampleName)
	buf.event["_raw"] = ""
	buf.event["_done"] = "_done"
	_, err = st.s2s.Send(buf.event)
	if err != nil {
		return err
	}
	delete(buf.event, "_raw")
	delete(buf.event, "_done")
	buf.buf.Reset()
	buf.events = 0
	return nil
}

func (st *splunktcpuf) Close() error {
	if !st.closed && st.initialized {
		time.Sleep(500 * time.Millisecond) // Hack for Cribl flush bug
		for _, buf := range st.bufs {
			err := st.Flush(buf)
			if err != nil {
				return err
			}
		}
		err := st.s2s.Close()
		if err != nil {
			return err
		}
		st.closed = true
	}
	return nil
}
