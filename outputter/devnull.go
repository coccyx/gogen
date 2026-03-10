package outputter

import (
	"io"

	config "github.com/coccyx/gogen/internal"
)

type devnull struct{}

func (foo devnull) Send(item *config.OutQueueItem) error {
	_, err := io.Copy(io.Discard, item.IO.R)
	return err
}

func (foo devnull) Close() error {
	return nil
}
