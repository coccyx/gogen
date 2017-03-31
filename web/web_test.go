package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"sync"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

var (
	w *Web
)

func init() {
	w = NewWeb()
}

func TestWebOutputStats(t *testing.T) {
	os := config.OutputStats{
		EventsWritten: 1,
		BytesWritten:  1,
		SampleName:    "test",
	}

	rOS := config.OutputStats{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		response, err := http.Get("http://localhost:9999/stats")
		assert.NoError(t, err)
		buf, err := ioutil.ReadAll(response.Body)
		assert.NoError(t, err)
		json.Unmarshal(buf, &rOS)
	}()
	time.Sleep(50 * time.Millisecond)
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.OutputStatsChan <- os
		time.Sleep(50 * time.Millisecond)
		w.Shutdown()
	}()
	wg.Wait()
	assert.Equal(t, os, rOS)
}

func TestWebQueueDepthStats(t *testing.T) {
	qs := QueueDepthStats{
		GeneratorQueueDepth: 1,
		OutputQueueDepth:    1,
	}

	rQS := QueueDepthStats{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		response, err := http.Get("http://localhost:9999/stats")
		assert.NoError(t, err)
		buf, err := ioutil.ReadAll(response.Body)
		assert.NoError(t, err)
		json.Unmarshal(buf, &rQS)
	}()
	time.Sleep(50 * time.Millisecond)
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.QueueDepthStatsChan <- qs
		time.Sleep(50 * time.Millisecond)
		w.Shutdown()
	}()
	wg.Wait()
	assert.Equal(t, qs, rQS)
}
