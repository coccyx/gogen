package timer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/tests"
	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	home := filepath.Join("..", "examples", "tutorial", "tutorial3")
	os.Setenv("GOGEN_CONFIG_DIR", home)

	c := config.NewConfig()
	s := c.FindSampleByName("tutorial3")
	s.Interval = 1
	s.Realtime = true
	gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)

	timer := &Timer{S: s, GQ: gq, OQ: oq}
	go timer.NewTimer(0)

	item := <-gq

	// Test that we get a GenQueueItem
	var gqi *config.GenQueueItem
	assert.Equal(t, reflect.TypeOf(gqi), reflect.ValueOf(item).Type())

	// Test that we're about the same interval
	n := time.Now()
	timer = &Timer{S: s, GQ: gq, OQ: oq}
	go timer.NewTimer(0)
	item = <-gq
	cur := time.Now()

	gt := cur.Sub(n) > (time.Duration(s.Interval) * time.Second)
	assert.Equal(t, true, gt)
}

func TestRealtimeReplay(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "timer")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := tests.FindSampleInFile(home, "realtimereplay")

	gq := make(chan *config.GenQueueItem, 1000)
	oq := make(chan *config.OutQueueItem)
	done := make(chan int)
	gqs := make([]*config.GenQueueItem, 0, 10)

	timer := &Timer{S: s, GQ: gq, OQ: oq, Done: done}
	go timer.NewTimer(0)

	time.Sleep(4 * time.Second)

Loop:
	for {
		select {
		case i := <-gq:
			gqs = append(gqs, i)
		default:
			break Loop
		}
	}
	// Should loop back around to the first entry and play it again
	assert.Equal(t, 4, len(gqs))
}

func TestBackfill(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "timer")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := tests.FindSampleInFile(home, "backfill")

	gq := make(chan *config.GenQueueItem, 1000)
	oq := make(chan *config.OutQueueItem)
	done := make(chan int)
	gqs := make([]*config.GenQueueItem, 0, 10)

	timer := &Timer{S: s, GQ: gq, OQ: oq, Done: done}
	go timer.NewTimer(0)
	<-done
Loop:
	for {
		select {
		case i := <-gq:
			gqs = append(gqs, i)
		default:
			break Loop
		}
	}
	assert.Equal(t, 6, len(gqs))
}

func TestBackfillRealtime(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "timer")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := tests.FindSampleInFile(home, "backfillrealtime")

	gq := make(chan *config.GenQueueItem, 1000)
	oq := make(chan *config.OutQueueItem)
	done := make(chan int)
	gqs := make([]*config.GenQueueItem, 0, 10)

	timer := &Timer{S: s, GQ: gq, OQ: oq, Done: done}
	go timer.NewTimer(0)

	time.Sleep(2 * time.Second)
Loop:
	for {
		select {
		case i := <-gq:
			gqs = append(gqs, i)
		default:
			break Loop
		}
	}
	inrange := len(gqs) >= 31 && len(gqs) <= 33
	assert.Equal(t, inrange, true)
}

func TestBackfillFutureEnd(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "timer")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := tests.FindSampleInFile(home, "backfillfutureend")

	gq := make(chan *config.GenQueueItem, 1000)
	oq := make(chan *config.OutQueueItem)
	done := make(chan int)
	gqs := make([]*config.GenQueueItem, 0, 10)

	timer := &Timer{S: s, GQ: gq, OQ: oq, Done: done}
	go timer.NewTimer(0)
	<-done
Loop:
	for {
		select {
		case i := <-gq:
			gqs = append(gqs, i)
		default:
			break Loop
		}
	}
	assert.Equal(t, 10, len(gqs))
}

func TestCacheIntervals(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "timer")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := tests.FindSampleInFile(home, "cachetimer")

	gq := make(chan *config.GenQueueItem, 1000)
	oq := make(chan *config.OutQueueItem)
	done := make(chan int)
	gqs := make([]*config.GenQueueItem, 0, 10)

	timer := &Timer{S: s, GQ: gq, OQ: oq, Done: done}
	go timer.NewTimer(2)
	<-done
Loop:
	for {
		select {
		case i := <-gq:
			gqs = append(gqs, i)
		default:
			break Loop
		}
	}
	assert.Equal(t, 30, len(gqs))
	for idx, gqitem := range gqs {
		// fmt.Printf("idx: %d, SetCache: %t, UseCache: %t\n", idx, gqitem.Cache.SetCache, gqitem.Cache.UseCache)
		if idx%3 == 0 {
			assert.True(t, gqitem.Cache.SetCache)
			assert.False(t, gqitem.Cache.UseCache)
		}
		if idx%3 == 1 || idx%3 == 2 {
			assert.True(t, gqitem.Cache.UseCache)
			assert.False(t, gqitem.Cache.SetCache)
		}
	}
}

func TestTimerClose(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "timer")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	s := tests.FindSampleInFile(home, "backfillrealtime")

	gq := make(chan *config.GenQueueItem, 1000)
	oq := make(chan *config.OutQueueItem)
	done := make(chan int)
	gqs := make([]*config.GenQueueItem, 0, 10)

	timer := &Timer{S: s, GQ: gq, OQ: oq, Done: done}
	go timer.NewTimer(2)

	// Let a few events generate
	time.Sleep(100 * time.Millisecond)

	// Close the timer
	timer.Close()
	close(done)

	// Give it a moment to shut down
	time.Sleep(50 * time.Millisecond)

Loop:
	for {
		select {
		case i := <-gq:
			gqs = append(gqs, i)
		default:
			break Loop
		}
	}

	// Verify we got some events but the timer stopped
	assert.Greater(t, len(gqs), 0)
	assert.Less(t, len(gqs), 100) // Sanity check that timer actually stopped
}
