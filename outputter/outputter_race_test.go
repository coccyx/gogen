package outputter

import (
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
)

func TestAccountWaitsForROTInitialization(t *testing.T) {
	rotchan = nil
	rotwg = sync.WaitGroup{}

	done := make(chan struct{})
	go func() {
		Account(1, 1, "race-sample")
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Account returned before ROT initialization; expected it to wait")
	case <-time.After(20 * time.Millisecond):
	}

	rotInterval = 1
	rotwg.Add(1)
	rotchan = make(chan *config.OutputStats)
	go readStats()

	t.Cleanup(func() {
		if rotchan != nil {
			close(rotchan)
			rotwg.Wait()
			rotchan = nil
		}
	})

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Account did not unblock after ROT initialization")
	}
}
