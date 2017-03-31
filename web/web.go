package web

import (
	"net/http"
	"sync"

	"encoding/json"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

// Web outputs JSON statistics to a webserver listening on port 9999
type Web struct {
	OutputStatsChan     chan config.OutputStats
	QueueDepthStatsChan chan QueueDepthStats
	mutex               *sync.RWMutex
	clients             []webResponse
	shutdownChan        chan int
}

// QueueDepthStats contains the length of each queue
type QueueDepthStats struct {
	GeneratorQueueDepth int
	OutputQueueDepth    int
}

type webResponse struct {
	w http.ResponseWriter
	f http.Flusher
	r *http.Request
	e *json.Encoder
}

// NewWeb returns a WebStats struct
func NewWeb() *Web {
	ws := &Web{mutex: &sync.RWMutex{}}
	ws.OutputStatsChan = make(chan config.OutputStats)
	go ws.sendOutputStats()
	ws.QueueDepthStatsChan = make(chan QueueDepthStats)
	go ws.sendQueueDepthStats()
	ws.shutdownChan = make(chan int)

	http.HandleFunc("/stats", ws.addClient)
	log.Infof("Starting web server, listening on :9999")
	go func() {
		err := http.ListenAndServe(":9999", nil)
		if err != nil {
			log.WithError(err).Error("Error starting HTTP Stats server")
		}
	}()
	return ws
}

func (ws *Web) addClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/json")
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}
	ws.mutex.Lock()
	ws.clients = append(ws.clients, webResponse{r: r, w: w, f: flusher, e: json.NewEncoder(w)})
	ws.mutex.Unlock()
Loop:
	for {
		select {
		case <-ws.shutdownChan:
			break Loop
		}
	}
}

func (ws *Web) sendOutputStats() {
	for {
		os, ok := <-ws.OutputStatsChan
		if !ok {
			ws.Shutdown()
			break
		}
		ws.mutex.RLock()
		for _, client := range ws.clients {
			client.e.Encode(os)
			client.f.Flush()
		}
		ws.mutex.RUnlock()
	}
}

func (ws *Web) sendQueueDepthStats() {
	for {
		qd, ok := <-ws.QueueDepthStatsChan
		if !ok {
			ws.Shutdown()
			break
		}
		ws.mutex.RLock()
		for _, client := range ws.clients {
			client.e.Encode(qd)
			client.f.Flush()
		}
		ws.mutex.RUnlock()
	}
}

// Shutdown shuts down open HTTP requests
func (ws *Web) Shutdown() {
	if len(ws.clients) > 0 {
		log.Infof("Shutting down web requests for %d clients", len(ws.clients))
		ws.mutex.RLock()
		for _, client := range ws.clients {
			client.f.Flush()
			client.r.Body.Close()
		}
		ws.mutex.RUnlock()
		ws.mutex.Lock()
		ws.clients = []webResponse{}
		ws.mutex.Unlock()
		ws.shutdownChan <- 1
	}
}
