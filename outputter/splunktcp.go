package outputter

import (
	"time"

	"github.com/coccyx/go-s2s/s2s"
	config "github.com/coccyx/gogen/internal"
)

type splunktcp struct {
	initialized bool
	done        chan int
	s2s         *s2s.S2S
}

func (st *splunktcp) Send(item *config.OutQueueItem) error {
	var err error
	if st.initialized == false {
		st.s2s, err = s2s.NewS2S(item.S.Output.Endpoints, item.S.Output.BufferBytes)
		if err != nil {
			return err
		}
		st.initialized = true
	}
	_, err = st.s2s.Copy(item.IO.R)
	if err != nil {
		return err
	}
	return nil
}

func (st *splunktcp) Close() error {
	if st.initialized {
		time.Sleep(500 * time.Millisecond) // Hack for Cribl flush bug
		if st.s2s != nil {
			st.s2s.Close()
			st.s2s = nil
		}
		st.initialized = false
	}
	return nil
}
