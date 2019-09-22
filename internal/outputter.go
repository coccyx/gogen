package internal

import (
	"io"
	"math/rand"
)

// OutQueueItem represents one batch of events to output
type OutQueueItem struct {
	S      *Sample
	Events []map[string]string
	Rand   *rand.Rand
	IO     *OutputIO
	OS     chan *OutputStats
	Cache  *CacheItem
}

// OutputStats are sent by each outputter to the ReadOutThread for accounting
type OutputStats struct {
	EventsWritten int64
	BytesWritten  int64
	SampleName    string
}

// OutputIO contains our Readers and Writers
type OutputIO struct {
	R io.Reader
	W io.WriteCloser
}

// NewOutputIO returns a freshly initialized pipe and TeeReader
func NewOutputIO() *OutputIO {
	o := new(OutputIO)
	o.R, o.W = io.Pipe()
	return o
}

// Outputter will do the work of actually sending events
type Outputter interface {
	Send(item *OutQueueItem) error
	Close() error
}
