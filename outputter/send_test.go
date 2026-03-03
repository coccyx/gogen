package outputter

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name         string
		outputter    string
		expectedType interface{}
	}{
		{"stdout", "stdout", &stdout{}},
		{"devnull", "devnull", &devnull{}},
		{"file", "file", &file{}},
		{"http", "http", &httpout{}},
		{"buf", "buf", &buf{}},
		{"network", "network", &network{}},
		{"kafka", "kafka", &kafkaout{}},
		{"unknown defaults to stdout", "unknowntype", &stdout{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use a unique gout slot for each test
			num := 99 // use last slot to avoid conflicts
			gout[num] = nil

			s := &config.Sample{
				Name: "test",
				Output: &config.Output{
					Outputter: tc.outputter,
				},
			}
			item := &config.OutQueueItem{S: s}
			source := rand.NewSource(0)
			gen := rand.New(source)

			result := setup(gen, item, num)
			assert.IsType(t, tc.expectedType, result)

			// Clean up
			gout[num] = nil
		})
	}
}

func TestDevnullSend(t *testing.T) {
	d := &devnull{}
	oio := config.NewOutputIO()
	item := &config.OutQueueItem{
		S:  &config.Sample{Name: "test"},
		IO: oio,
	}

	go func() {
		io.WriteString(oio.W, "test data")
		oio.W.Close()
	}()

	err := d.Send(item)
	assert.NoError(t, err)
}

func TestDevnullClose(t *testing.T) {
	d := &devnull{}
	err := d.Close()
	assert.NoError(t, err)
}

func TestBufSend(t *testing.T) {
	var b bytes.Buffer
	s := &config.Sample{
		Name: "test",
		Buf:  &b,
	}
	oio := config.NewOutputIO()
	item := &config.OutQueueItem{
		S:  s,
		IO: oio,
	}

	go func() {
		io.WriteString(oio.W, "buffered data\n")
		oio.W.Close()
	}()

	bu := &buf{}
	err := bu.Send(item)
	assert.NoError(t, err)
	assert.Equal(t, "buffered data\n", b.String())
}

func TestStdoutSend(t *testing.T) {
	// Redirect stdout to a pipe
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	oio := config.NewOutputIO()
	item := &config.OutQueueItem{
		S:  &config.Sample{Name: "test"},
		IO: oio,
	}

	go func() {
		io.WriteString(oio.W, "stdout data\n")
		oio.W.Close()
	}()

	so := &stdout{}
	err := so.Send(item)
	assert.NoError(t, err)

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = origStdout

	assert.Equal(t, "stdout data\n", buf.String())
}

func TestStdoutClose(t *testing.T) {
	so := &stdout{}
	err := so.Close()
	assert.NoError(t, err)
}

func TestFileSendAndRotation(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "testfile.log")

	s := &config.Sample{
		Name: "filesample",
		Output: &config.Output{
			FileName:    filename,
			MaxBytes:    50, // Very small to trigger rotation
			BackupFiles: 2,
		},
	}

	f := &file{}

	// Write enough data to trigger rotation
	for i := 0; i < 5; i++ {
		oio := config.NewOutputIO()
		item := &config.OutQueueItem{S: s, IO: oio}

		go func() {
			io.WriteString(oio.W, strings.Repeat("X", 30)+"\n")
			oio.W.Close()
		}()

		err := f.Send(item)
		assert.NoError(t, err)
	}

	// Check that backup files were created
	_, err := os.Stat(filename)
	assert.NoError(t, err, "main file should exist")

	_, err = os.Stat(filename + ".1")
	assert.NoError(t, err, "backup .1 should exist")

	f.Close()
}

func TestFileSendExistingFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "existing.log")

	// Pre-create the file with some content
	os.WriteFile(filename, []byte("pre-existing data\n"), 0644)

	s := &config.Sample{
		Name: "fileexisting",
		Output: &config.Output{
			FileName:    filename,
			MaxBytes:    10000000,
			BackupFiles: 2,
		},
	}

	f := &file{}

	oio := config.NewOutputIO()
	item := &config.OutQueueItem{S: s, IO: oio}

	go func() {
		io.WriteString(oio.W, "appended data\n")
		oio.W.Close()
	}()

	err := f.Send(item)
	assert.NoError(t, err)

	// Verify the file has both old and new data
	data, _ := os.ReadFile(filename)
	assert.Contains(t, string(data), "pre-existing data")
	assert.Contains(t, string(data), "appended data")

	f.Close()
}

func TestFileClose(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "closefile.log")

	s := &config.Sample{
		Name: "fileclose",
		Output: &config.Output{
			FileName:    filename,
			MaxBytes:    1000000,
			BackupFiles: 2,
		},
	}

	f := &file{}

	oio := config.NewOutputIO()
	item := &config.OutQueueItem{S: s, IO: oio}
	go func() {
		io.WriteString(oio.W, "data\n")
		oio.W.Close()
	}()
	f.Send(item)

	// Close should work
	err := f.Close()
	assert.NoError(t, err)

	// Close again should be idempotent
	err = f.Close()
	assert.NoError(t, err)
}

func TestHTTPSendAndFlush(t *testing.T) {
	var received bytes.Buffer
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		io.Copy(&received, r.Body)
		mu.Unlock()
		w.WriteHeader(200)
	}))
	defer ts.Close()

	s := &config.Sample{
		Name: "httpsample",
		Output: &config.Output{
			Endpoints:   []string{ts.URL},
			BufferBytes: 10, // Small buffer to trigger flush
			Headers:     map[string]string{"Content-Type": "application/json"},
			Timeout:     5 * time.Second,
		},
	}

	h := &httpout{}

	// Send enough data to exceed buffer and trigger flush
	oio := config.NewOutputIO()
	item := &config.OutQueueItem{S: s, IO: oio}
	go func() {
		io.WriteString(oio.W, strings.Repeat("D", 50)+"\n")
		oio.W.Close()
	}()

	err := h.Send(item)
	assert.NoError(t, err)

	// Verify server received data
	mu.Lock()
	data := received.String()
	mu.Unlock()
	assert.NotEmpty(t, data, "HTTP server should have received data")

	// Close flushes remaining data
	err = h.Close()
	assert.NoError(t, err)
}

func TestNetworkSend(t *testing.T) {
	// Start a TCP listener on a random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer ln.Close()

	var received bytes.Buffer
	done := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		io.Copy(&received, conn)
		conn.Close()
		close(done)
	}()

	s := &config.Sample{
		Name: "netsample",
		Output: &config.Output{
			Endpoints: []string{ln.Addr().String()},
			Protocol:  "tcp",
			Timeout:   5 * time.Second,
		},
	}

	n := &network{}

	oio := config.NewOutputIO()
	item := &config.OutQueueItem{S: s, IO: oio}
	go func() {
		io.WriteString(oio.W, "network data\n")
		oio.W.Close()
	}()

	err = n.Send(item)
	assert.NoError(t, err)

	// Close the connection so the listener goroutine can finish
	n.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for network data")
	}

	assert.Equal(t, "network data\n", received.String())
}

func TestNetworkClose(t *testing.T) {
	n := &network{}
	// Close with no connection should not error
	err := n.Close()
	assert.NoError(t, err)
	assert.False(t, n.initialized)
}

func TestBufClose(t *testing.T) {
	b := &buf{}
	err := b.Close()
	assert.NoError(t, err)
}

func TestStartDevnullWorker(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	// Reset gout slot
	gout[0] = nil

	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go Start(oq, oqs, 0)

	// Send an item through the pipeline
	s := &config.Sample{
		Name: "starttest",
		Output: &config.Output{
			Outputter:      "devnull",
			OutputTemplate: "raw",
		},
	}
	events := []map[string]string{
		{"_raw": "test event for start"},
	}
	item := &config.OutQueueItem{
		S:      s,
		Events: events,
		Cache:  &config.CacheItem{},
	}
	oq <- item

	// Close the queue and wait for worker to finish
	close(oq)
	select {
	case <-oqs:
		// Worker finished
	case <-time.After(5 * time.Second):
		t.Fatal("Start worker did not finish in time")
	}

	// Verify gout slot was cleared
	assert.Nil(t, gout[0])
}

func TestStartMultipleItems(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	gout[0] = nil

	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go Start(oq, oqs, 0)

	s := &config.Sample{
		Name: "multistart",
		Output: &config.Output{
			Outputter:      "devnull",
			OutputTemplate: "raw",
		},
	}

	for i := 0; i < 5; i++ {
		events := []map[string]string{
			{"_raw": "event number"},
		}
		item := &config.OutQueueItem{
			S:      s,
			Events: events,
			Cache:  &config.CacheItem{},
		}
		oq <- item
	}

	close(oq)
	select {
	case <-oqs:
	case <-time.After(5 * time.Second):
		t.Fatal("Start worker did not finish in time")
	}

	// Check that events were accounted for
	time.Sleep(50 * time.Millisecond)
	Mutex.RLock()
	ew := EventsWritten["multistart"]
	Mutex.RUnlock()
	assert.Equal(t, int64(5), ew)
}

func TestStartEmptyEvents(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	gout[0] = nil

	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go Start(oq, oqs, 0)

	s := &config.Sample{
		Name: "emptyevents",
		Output: &config.Output{
			Outputter:      "devnull",
			OutputTemplate: "raw",
		},
	}
	// Send item with no events - should skip the write/send
	item := &config.OutQueueItem{
		S:      s,
		Events: []map[string]string{},
		Cache:  &config.CacheItem{},
	}
	oq <- item

	close(oq)
	select {
	case <-oqs:
	case <-time.After(5 * time.Second):
		t.Fatal("Start worker did not finish in time")
	}
}

func TestStartCloseOnChannelClose(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	gout[0] = nil

	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go Start(oq, oqs, 0)

	s := &config.Sample{
		Name: "closetest",
		Output: &config.Output{
			Outputter:      "devnull",
			OutputTemplate: "raw",
		},
	}
	// Send one real item so lastS is set, then close
	events := []map[string]string{
		{"_raw": "test event"},
	}
	item := &config.OutQueueItem{
		S:      s,
		Events: events,
		Cache:  &config.CacheItem{},
	}
	oq <- item

	// Close the channel - should trigger the Close() path on the outputter
	close(oq)
	select {
	case <-oqs:
	case <-time.After(5 * time.Second):
		t.Fatal("Start worker did not finish in time")
	}

	// gout should be cleared
	assert.Nil(t, gout[0])
}

func TestStartSendError(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	// Use a network outputter pointed at a bad address to trigger Send error
	gout[0] = nil

	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go Start(oq, oqs, 0)

	s := &config.Sample{
		Name: "senderror",
		Output: &config.Output{
			Outputter:      "network",
			OutputTemplate: "raw",
			Endpoints:      []string{"127.0.0.1:1"}, // Should fail to connect
			Protocol:       "tcp",
		},
	}
	events := []map[string]string{
		{"_raw": "error event"},
	}
	item := &config.OutQueueItem{
		S:      s,
		Events: events,
		Cache:  &config.CacheItem{},
	}
	oq <- item

	close(oq)
	select {
	case <-oqs:
	case <-time.After(10 * time.Second):
		t.Fatal("Start worker did not finish in time")
	}
}

func TestHTTPFlushNon200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	s := &config.Sample{
		Name: "httpfail",
		Output: &config.Output{
			Endpoints:   []string{ts.URL},
			BufferBytes: 10,
			Headers:     map[string]string{"Content-Type": "text/plain"},
			Timeout:     5 * time.Second,
		},
	}

	h := &httpout{}

	oio := config.NewOutputIO()
	item := &config.OutQueueItem{S: s, IO: oio}
	go func() {
		io.WriteString(oio.W, strings.Repeat("X", 50)+"\n")
		oio.W.Close()
	}()

	err := h.Send(item)
	// flush should return an error due to non-200 status
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestHTTPCloseFlushError(t *testing.T) {
	// Use a server that accepts the first request (Send flush) but returns error on the second (Close flush)
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 1 {
			w.WriteHeader(500)
			w.Write([]byte("flush error"))
			return
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	s := &config.Sample{
		Name: "httpclose",
		Output: &config.Output{
			Endpoints:   []string{ts.URL},
			BufferBytes: 10, // Small buffer to trigger flush on first Send
			Headers:     map[string]string{},
			Timeout:     5 * time.Second,
		},
	}

	h := &httpout{}

	// First Send: exceeds buffer, triggers flush (call #1 → 200 OK)
	oio := config.NewOutputIO()
	item := &config.OutQueueItem{S: s, IO: oio}
	go func() {
		io.WriteString(oio.W, strings.Repeat("X", 50))
		oio.W.Close()
	}()

	err := h.Send(item)
	assert.NoError(t, err)

	// Add more data to buffer for Close to flush
	h.buf.WriteString("leftover data")

	// Close should flush remaining data and get 500 error (call #2)
	err = h.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestStartSendErrorRepeat(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	gout[0] = nil

	oq := make(chan *config.OutQueueItem)
	oqs := make(chan int)

	go Start(oq, oqs, 0)

	s := &config.Sample{
		Name: "senderrorrepeat",
		Output: &config.Output{
			Outputter:      "network",
			OutputTemplate: "raw",
			Endpoints:      []string{"127.0.0.1:1"},
			Protocol:       "tcp",
		},
	}
	// Send multiple items to trigger repeat error path (lasterr[num].count++)
	for i := 0; i < 3; i++ {
		events := []map[string]string{
			{"_raw": "error event repeat"},
		}
		item := &config.OutQueueItem{
			S:      s,
			Events: events,
			Cache:  &config.CacheItem{},
		}
		oq <- item
	}

	close(oq)
	select {
	case <-oqs:
	case <-time.After(15 * time.Second):
		t.Fatal("Start worker did not finish in time")
	}
}
