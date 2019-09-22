package run

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coccyx/gogen/generator"
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/outputter"
	"github.com/coccyx/gogen/timer"
)

// ROT reads out data every ROTInterval seconds
func ROT(c *config.Config, gq chan *config.GenQueueItem, oq chan *config.OutQueueItem) {
	for {
		timer := time.NewTimer(time.Duration(c.Global.ROTInterval) * time.Second * 5)
		<-timer.C
		log.Infof("Generator Queue: %d Output Queue: %d", len(gq), len(oq))
		// log.Infof("Goroutines: %d", runtime.NumGoroutine())
	}
}

// Run runs the mainline of the program
func Run(c *config.Config) {
	log.Info("Starting ReadOutThread")
	go outputter.ROT(c)
	log.Info("Starting Timers")
	timerdone := make(chan int)
	gq := make(chan *config.GenQueueItem, c.Global.GeneratorQueueLength)
	gqs := make(chan int)
	oq := make(chan *config.OutQueueItem, c.Global.OutputQueueLength)
	oqs := make(chan int)
	gens := 0
	outs := 0
	timers := []*timer.Timer{}
	for i := 0; i < len(c.Samples); i++ {
		s := c.Samples[i]
		if !s.Disabled {
			t := timer.Timer{S: s, GQ: gq, OQ: oq, Done: timerdone}
			go t.NewTimer(c.Global.CacheIntervals)
			timers = append(timers, &t)
		}
	}
	log.Infof("%d Timers started", len(timers))

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

	go ROT(c, gq, oq)

	// time.Sleep(1000 * time.Millisecond)

	// Check if any timers are done
	donechan := make(chan bool)
	go func() {
		timerCount := len(timers)
		for {
			select {
			case <-timerdone:
				timerCount--
				log.Debugf("Timer done, timers left %d", timerCount)
				if timerCount == 0 {
					log.Infof("Timers all done, closing generating queue")
					donechan <- true
				}
			}
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	go func() {
		for range sigchan {
			log.Infof("Caught interrupt, shutting down")
			// Shut down timers
			for _, t := range timers {
				t.Close()
			}
			// Drain generator queue
			for range gq {
				continue
			}
		}
	}()

	// Close our channels to signal to the workers to shut down when the queue is clear
	<-donechan
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

	// for _, s := range c.Samples {
	// 	err := s.Out.Close()
	// 	if err != nil {
	// 		log.Errorf("Error closing output for sample '%s': %s", s.Name, err)
	// 	}
	// }

	// time.Sleep(100 * time.Millisecond)

	outputter.ReadFinal()
}
