package outputter

import (
	"io"
	"net"
	"time"

	config "github.com/coccyx/gogen/internal"
)

type network struct {
	conn        net.Conn
	initialized bool
	closed      bool
}

func (n *network) Send(item *config.OutQueueItem) error {
	if n.initialized == false {
		timeout, err := time.ParseDuration(item.S.Output.NetworkTimeout)
		if err != nil {
			timeout, _ = time.ParseDuration("10s")
		}

		conn, err := net.DialTimeout(item.S.Output.Protocol, item.S.Output.Server, timeout)
		if err != nil {
			return err
		}
		n.conn = conn
		n.initialized = true
	}
	_, err := io.Copy(n.conn, item.IO.R)
	return err
}

func (n *network) Close() error {
	n.closed = true
	if n.conn != nil {
		return n.conn.Close()
	}
	return nil
}
