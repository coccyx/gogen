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

	// Added web support.  For RunOnce, set Web to disabled and we won't try to send
	c.Global.Web = false
	sc := make(chan config.OutputStats)
	// The ROT here is to keep from blocking on accounting output
	go outputter.ROT(c, sc)

	s := c.FindSampleByName(name)

	source := rand.NewSource(time.Now().UnixNano())
	randgen := rand.New(source)
	// Generate one event for our named sample
	if s.Description == "" {
		log.Fatalf("Description not set for sample '%s'", s.Name)
	}

	log.Debugf("Generating for Push() sample '%s'", s.Name)
	origOutputter := s.Output.Outputter
	origOutputTemplate := s.Output.OutputTemplate
	s.Output.Outputter = "buf"
	s.Output.OutputTemplate = "json"
	gq := make(chan *config.GenQueueItem)
	gqs := make(chan int)
	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go generator.Start(gq, gqs)
	go outputter.Start(oq, oqs, 1)

	gqi := &config.GenQueueItem{Count: 1, Earliest: time.Now(), Latest: time.Now(), S: s, OQ: oq, Rand: randgen, Event: -1}
	gq <- gqi

	time.Sleep(time.Second)

	log.Debugf("Closing generator channel")
	close(gq)

Loop1:
	for {
		select {
		case <-gqs:
			log.Debugf("Generator closed")
			break Loop1
		}
	}

	log.Debugf("Closing outputter channel")
	close(oq)

Loop2:
	for {
		select {
		case <-oqs:
			log.Debugf("Outputter closed")
			break Loop2
		}
	}

	s.Output.Outputter = origOutputter
	s.Output.OutputTemplate = origOutputTemplate

	log.Debugf("Buffer: %s", c.Buf.String())
}
