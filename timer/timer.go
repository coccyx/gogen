package timer

import (
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/rater"
)

// Timer will put work into the generator queue on an interval specified by the Sample.
// One instance is created per sample.
type Timer struct {
	S    *config.Sample
	cur  int
	GQ   chan *config.GenQueueItem
	OQ   chan *config.OutQueueItem
	Done chan int
}

// NewTimer creates a new Timer for a sample which will put work into the generator queue on each interval
func (t *Timer) NewTimer() {
	s := t.S
	// If we're not realtime, then we should be backfilling
	if !s.Realtime {
		// Set the end time based on configuration, either now or a specified time in the config
		var endtime time.Time
		n := time.Now()
		if s.EndParsed.Before(n) && !s.EndParsed.IsZero() {
			endtime = s.EndParsed
		} else {
			endtime = n
		}
		// Run through as many intervals until we're at endtime
		for s.Current.Before(endtime) {
			// log.Debugf("Backfilling, at %s, ending at %s", t.S.Current, endtime)
			t.genWork()
			t.inc()
		}
		// If we had no endtime set, then keep going in realtime mode
		if s.EndParsed.IsZero() {
			for s.Current.Before(time.Now()) {
				t.genWork()
				t.inc()
			}
			s.Realtime = true
		}
	}
	// Endtime can be greater than now, so continue until we've reached the end time... Realtime won't get set, so we'll end after this
	if !t.S.Realtime {
		for s.Current.Before(s.EndParsed) {
			t.genWork()
			t.inc()
		}
	}
	// In realtime mode, continue until we get an interrupt
	if s.Realtime {
		for {
			if s.Generator == "replay" {
				time.Sleep(s.ReplayOffsets[t.cur])
				t.genWork()
				t.cur++
				if t.cur >= len(s.ReplayOffsets) {
					t.cur = 0
				}
			} else {
				timer := time.NewTimer(time.Duration(s.Interval) * time.Second)
				<-timer.C
				t.genWork()
			}
		}
	} else {
		t.Done <- 1
	}
}

func (t *Timer) genWork() {
	s := t.S
	now := s.Now()
	var item *config.GenQueueItem
	if s.Generator == "replay" {
		earliest := now
		latest := now
		count := 1
		item = &config.GenQueueItem{S: s, Count: count, Event: t.cur, Earliest: earliest, Latest: latest, Now: now, OQ: t.OQ}
	} else {
		earliest := now.Add(s.EarliestParsed)
		latest := now.Add(s.LatestParsed)
		count := int(rater.EventRate(s, now, s.Count))
		item = &config.GenQueueItem{S: s, Count: count, Event: -1, Earliest: earliest, Latest: latest, Now: now, OQ: t.OQ}
	}
	// log.Debugf("Placing item in queue for sample '%s': %#v", t.S.Name, item)
	t.GQ <- item
}

func (t *Timer) inc() {
	s := t.S
	if s.Generator == "replay" {
		s.Current = s.Current.Add(s.ReplayOffsets[t.cur])
		t.cur++
		if t.cur >= len(s.ReplayOffsets) {
			t.cur = 0
		}
	} else {
		s.Current = s.Current.Add(time.Duration(s.Interval) * time.Second)
	}
}
