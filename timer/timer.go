package timer

import (
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/rater"
)

// Timer will put work into the generator queue on an interval specified by the Sample.
// One instance is created per sample.
type Timer struct {
	S              *config.Sample
	cur            int
	GQ             chan *config.GenQueueItem
	OQ             chan *config.OutQueueItem
	Done           chan int
	closed         bool
	cacheCounter   int // Number of intervals left to use cache
	cacheIntervals int // Number of intervals to cache for
}

// NewTimer creates a new Timer for a sample which will put work into the generator queue on each interval
func (t *Timer) NewTimer(cacheIntervals int) {
	s := t.S
	t.cacheIntervals = cacheIntervals
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
		t.backfill(endtime)
		// If we had no endtime set, then keep going in realtime mode
		if s.EndParsed.IsZero() {
			t.backfill(time.Now())
			s.Realtime = true
		}
	}
	// Endtime can be greater than now, so continue until we've reached the end time... Realtime won't get set, so we'll end after this
	if !t.S.Realtime {
		t.backfill(s.EndParsed)
	}
	// In realtime mode, continue until we get an interrupt
	if s.Realtime {
		for {
			if s.Generator == "replay" {
				t.genWork()
				time.Sleep(s.ReplayOffsets[t.cur])
				t.cur++
				if t.cur >= len(s.ReplayOffsets) {
					t.cur = 0
				}
			} else {
				if s.Interval > 5 {
					// For longer intervals, use ticker to check closed status
					mainTimer := time.NewTimer(time.Duration(s.Interval) * time.Second)
					checkTicker := time.NewTicker(1 * time.Second)
					select {
					case <-mainTimer.C:
						checkTicker.Stop()
						t.genWork()
					case <-checkTicker.C:
						if t.closed {
							mainTimer.Stop()
							checkTicker.Stop()
							break
						}
						continue
					}
				} else {
					// For short intervals, just use the timer directly
					timer := time.NewTimer(time.Duration(s.Interval) * time.Second)
					<-timer.C
					t.genWork()
				}
			}
			if t.closed {
				break
			}
		}
	}
	t.Done <- 1
}

func (t *Timer) backfill(until time.Time) {
	for t.S.Current.Before(until) {
		t.genWork()
		t.inc()
		if t.closed {
			break
		}
	}
}

func (t *Timer) genWork() {
	s := t.S
	now := s.Now()
	var item *config.GenQueueItem
	useCache := t.cacheCounter > 0
	setCache := !useCache && t.cacheIntervals > 0
	t.cacheCounter--
	if t.cacheCounter < 0 {
		t.cacheCounter = t.cacheIntervals
	}
	ci := &config.CacheItem{
		UseCache: useCache,
		SetCache: setCache,
	}
	if s.Generator == "replay" {
		earliest := now
		latest := now
		count := 1
		item = &config.GenQueueItem{S: s, Count: count, Event: t.cur, Earliest: earliest, Latest: latest, Now: now, OQ: t.OQ, Cache: ci}
	} else {
		earliest := now.Add(s.EarliestParsed)
		latest := now.Add(s.LatestParsed)
		count := rater.EventRate(s, now, s.Count)
		item = &config.GenQueueItem{S: s, Count: count, Event: -1, Earliest: earliest, Latest: latest, Now: now, OQ: t.OQ, Cache: ci}
	}
	// log.Debugf("Placing item in queue for sample '%s': %#v", t.S.Name, item)
Loop1:
	for {
		select {
		case t.GQ <- item:
			break Loop1
		case <-time.After(1 * time.Second):
			if t.closed {
				log.Debugf("Timer %s closed", t.S.Name)
				break Loop1
			}
			continue
		}
	}
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
	if s.Wait {
		timer := time.NewTimer(time.Duration(s.Interval) * time.Second)
		<-timer.C
	}
}

// Close shuts down a timer
func (t *Timer) Close() {
	log.Infof("Closing timer for sample %s", t.S.Name)
	t.closed = true
}
