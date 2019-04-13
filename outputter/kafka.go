package outputter

import (
	"context"
	"io"
	"math/rand"
	"time"

	config "github.com/coccyx/gogen/internal"
	kafka "github.com/segmentio/kafka-go"
)

type kafkaout struct {
	conn        *kafka.Conn
	initialized bool
	closed      bool
	cancel      context.CancelFunc
}

func (k *kafkaout) Send(item *config.OutQueueItem) error {
	if k.initialized == false {
		var err error
		endpoint := item.S.Output.Endpoints[rand.Intn(len(item.S.Output.Endpoints))]
		d := &kafka.Dialer{}
		timeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		k.conn, err = d.DialLeader(timeout, "tcp", endpoint, "topic", 0)
		if err != nil {
			return err
		}
		k.initialized = true
	}
	_, err := io.Copy(k.conn, item.IO.R)
	return err
}

func (k *kafkaout) Close() error {
	k.closed = true
	if k.conn != nil {
		k.conn.Close()
		k.cancel()
	}
	return nil
}
