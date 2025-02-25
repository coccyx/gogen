package run

import (
	"math/rand"
	"time"

	"github.com/coccyx/gogen/generator"
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/outputter"
)

// Runner is a naked struct allowing Once to be an interface
type Runner struct{}

// Once runs a given sample a single time and outputs to a byte Buffer
func (r Runner) Once(name string) {
	c := config.NewConfig()
	go outputter.ROT(c)
	r.onceWithConfig(name, c)
}

// onceWithConfig runs a sample once using the provided config
func (r Runner) onceWithConfig(name string, c *config.Config) {
	s := c.FindSampleByName(name)

	source := rand.NewSource(time.Now().UnixNano())
	randgen := rand.New(source)
	// Generate one event for our named sample
	if s.Description == "" {
		log.Fatalf("Description not set for sample '%s'", s.Name)
	}

	log.Debugf("Generating for sample '%s'", s.Name)
	origOutputter := s.Output.Outputter
	origOutputTemplate := s.Output.OutputTemplate
	s.Output.Outputter = "buf"
	s.Output.OutputTemplate = "json"
	gq := make(chan *config.GenQueueItem)
	gqs := make(chan int)
	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	// Start outputter first so it's ready to receive
	go outputter.Start(oq, oqs, 1)
	// Then start generator
	go generator.Start(gq, gqs)

	// Get current time for event generation
	now := time.Now()
	if c.Global.UTC {
		now = now.UTC()
	}

	// Send generation request
	gqi := &config.GenQueueItem{
		Count:    1,
		Earliest: now,
		Latest:   now,
		S:        s,
		OQ:       oq,
		Rand:     randgen,
		Event:    -1,
		Cache: &config.CacheItem{
			UseCache: false,
			SetCache: false,
		},
	}
	gq <- gqi

	// Close generator and wait for it to finish
	log.Debugf("Closing generator channel")
	close(gq)

Loop1:
	for {
		select {
		case <-gqs:
			log.Debugf("Generator closed")
			break Loop1
		case <-time.After(2 * time.Second):
			log.Debugf("Generator timeout waiting for close signal")
			break Loop1
		}
	}

	// Give outputter time to process any remaining items
	time.Sleep(100 * time.Millisecond)

	// Now close outputter and wait for it to finish
	log.Debugf("Closing outputter channel")
	close(oq)

Loop2:
	for {
		select {
		case <-oqs:
			log.Debugf("Outputter closed")
			break Loop2
		case <-time.After(2 * time.Second):
			log.Debugf("Outputter timeout waiting for close signal")
			break Loop2
		}
	}

	s.Output.Outputter = origOutputter
	s.Output.OutputTemplate = origOutputTemplate

	log.Debugf("Buffer contents: %s", s.Buf.String())
}
