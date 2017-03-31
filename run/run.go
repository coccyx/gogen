package run

import (
	"time"

	"github.com/coccyx/gogen/generator"
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/outputter"
	"github.com/coccyx/gogen/timer"
	"github.com/coccyx/gogen/web"
)

// ROT reads out data every ROTInterval seconds
func ROT(c *config.Config, gq chan *config.GenQueueItem, oq chan *config.OutQueueItem, qds chan web.QueueDepthStats) {
	for {
		timer := time.NewTimer(time.Duration(c.Global.ROTInterval) * time.Second * 5)
		<-timer.C
		log.Infof("Generator Queue: %d Output Queue: %d", len(gq), len(oq))
		qds <- web.QueueDepthStats{GeneratorQueueDepth: len(gq), OutputQueueDepth: len(oq)}
	}
}

// Run runs the mainline of the program
func Run(c *config.Config) {
	log.Info("Starting ReadOutThread")
	w := web.NewWeb()
	go outputter.ROT(c, w.OutputStatsChan)
	log.Info("Starting Timers")
	timerdone := make(chan int)
	gq := make(chan *config.GenQueueItem, config.MaxGenQueueLength)
	gqs := make(chan int)
	oq := make(chan *config.OutQueueItem, config.MaxOutQueueLength)
	oqs := make(chan int)
	gens := 0
	outs := 0
	timers := 0
	for i := 0; i < len(c.Samples); i++ {
		s := c.Samples[i]
		if !s.Disabled {
			t := timer.Timer{S: s, GQ: gq, OQ: oq, Done: timerdone}
			go t.NewTimer()
			timers++
		}
	}
	log.Infof("%d Timers started", timers)

	log.Infof("Starting Generators")
	for i := 0; i < c.Global.GeneratorWorkers; i++ {
		log.Infof("Starting Generator %d", i)
		go generator.Start(gq, gqs)
		gens++
	}

	log.Infof("Starting Outputters")
	for i := 0; i < c.Global.OutputWorkers; i++ {
		log.Infof("Starting Outputter %d", i)
		go outputter.Start(oq, oqs, i)
		outs++
	}

	go ROT(c, gq, oq, w.QueueDepthStatsChan)

	// time.Sleep(1000 * time.Millisecond)

	// Check if any timers are done
Loop1:
	for {
		select {
		case <-timerdone:
			timers--
			log.Debugf("Timer done, timers left %d", timers)
			if timers == 0 {
				break Loop1
			}
		}
	}

	// Close our channels to signal to the workers to shut down when the queue is clear
	log.Infof("Timers all done, closing generating queue")
	close(gq)

	// Check for all the workers to signal back they're done
Loop2:
	for {
		select {
		case <-gqs:
			gens--
			log.Debugf("Gen done, gens left %d", gens)
			if gens == 0 {
				break Loop2
			}
		}
	}

	// Close our output channel to signal to outputters we're done
	close(oq)
Loop3:
	for {
		select {
		case <-oqs:
			outs--
			log.Debugf("Out done, outs left %d", outs)
			if outs == 0 {
				break Loop3
			}
		}
	}

	w.Shutdown()

	// for _, s := range c.Samples {
	// 	err := s.Out.Close()
	// 	if err != nil {
	// 		log.Errorf("Error closing output for sample '%s': %s", s.Name, err)
	// 	}
	// }

	// time.Sleep(100 * time.Millisecond)
}
