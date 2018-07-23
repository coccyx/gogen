package outputter

import (
	"io"
	"math/rand"
	"net"

	config "github.com/coccyx/gogen/internal"
)

type network struct {
	conn        net.Conn
	initialized bool
	closed      bool
}

func (n *network) Send(item *config.OutQueueItem) error {
	if n.initialized == false {
		endpoint := item.S.Output.Endpoints[rand.Intn(len(item.S.Output.Endpoints))]
		conn, err := net.DialTimeout(item.S.Output.Protocol, endpoint, item.S.Output.Timeout)
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
