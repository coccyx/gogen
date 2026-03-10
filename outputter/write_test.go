package outputter

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/template"
	"github.com/stretchr/testify/assert"
)

// initROT initializes the rotchan and readStats goroutine needed by write().
// Returns a cleanup function to call via defer.
func initROT() func() {
	Mutex.Lock()
	BytesWritten = make(map[string]int64)
	EventsWritten = make(map[string]int64)
	rotwg = sync.WaitGroup{}
	rotchan = make(chan *config.OutputStats)
	Mutex.Unlock()
	rotwg.Add(1)
	go readStats()
	return func() {
		close(rotchan)
		rotwg.Wait()
	}
}

func makeOutQueueItem(sampleName, outputTemplate, outputter string, events []map[string]string) *config.OutQueueItem {
	s := &config.Sample{
		Name: sampleName,
		Output: &config.Output{
			Outputter:      outputter,
			OutputTemplate: outputTemplate,
		},
	}
	oio := config.NewOutputIO()
	return &config.OutQueueItem{
		S:      s,
		Events: events,
		IO:     oio,
		Cache:  &config.CacheItem{},
	}
}

func readFromPipe(item *config.OutQueueItem) string {
	var buf bytes.Buffer
	io.Copy(&buf, item.IO.R)
	return buf.String()
}

func TestWriteRaw(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "hello world"},
	}
	item := makeOutQueueItem("rawsample", "raw", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	assert.Contains(t, result, "hello world")
}

func TestWriteJSON(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "test event", "host": "myhost"},
	}
	item := makeOutQueueItem("jsonsample", "json", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	var parsed map[string]string
	lines := strings.TrimSpace(result)
	err := json.Unmarshal([]byte(strings.Split(lines, "\n")[0]), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "test event", parsed["_raw"])
	assert.Equal(t, "myhost", parsed["host"])
}

func TestWriteSplunkHEC(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "splunk event", "_time": "1234567890"},
	}
	item := makeOutQueueItem("hecsample", "splunkhec", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	var parsed map[string]string
	lines := strings.TrimSpace(result)
	err := json.Unmarshal([]byte(strings.Split(lines, "\n")[0]), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "splunk event", parsed["event"], "splunkhec should remap _raw to event")
	assert.Equal(t, "1234567890", parsed["time"], "splunkhec should remap _time to time")
	assert.Empty(t, parsed["_raw"], "_raw should be deleted")
	assert.Empty(t, parsed["_time"], "_time should be deleted")
}

func TestWriteRFC3164(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "syslog msg", "_time": "Oct 20 12:00:00", "priority": "13", "host": "myhost", "tag": "gogen", "pid": "1234"},
	}
	item := makeOutQueueItem("rfc3164sample", "rfc3164", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	assert.Contains(t, result, "<13>")
	assert.Contains(t, result, "Oct 20 12:00:00")
	assert.Contains(t, result, "myhost")
	assert.Contains(t, result, "gogen[1234]")
	assert.Contains(t, result, "syslog msg")
}

func TestWriteRFC5424(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "syslog5424 msg", "_time": "2001-10-20T12:00:00Z", "priority": "13", "host": "myhost", "appName": "gogen", "pid": "1234", "extra": "val"},
	}
	item := makeOutQueueItem("rfc5424sample", "rfc5424", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	assert.Contains(t, result, "<13>1")
	assert.Contains(t, result, "myhost")
	assert.Contains(t, result, "gogen")
	assert.Contains(t, result, "syslog5424 msg")
	assert.Contains(t, result, "[meta")
	assert.Contains(t, result, `extra="val"`)
}

func TestWriteElasticsearch(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "es event", "index": "testindex"},
	}
	item := makeOutQueueItem("essample", "elasticsearch", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	assert.Contains(t, result, `"_index": "testindex"`)
	assert.Contains(t, result, `"_type": "doc"`)
	// _raw should be remapped to message
	var parsed map[string]interface{}
	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "elasticsearch should produce index header + body")
	err := json.Unmarshal([]byte(lines[1]), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "es event", parsed["message"])
}

func TestWriteDevnull(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "devnull event data"},
	}
	item := makeOutQueueItem("devnullsample", "raw", "devnull", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	// devnull should not write any content through the pipe
	assert.Empty(t, result)

	// But bytes should still be accounted for
	time.Sleep(50 * time.Millisecond)
	Mutex.RLock()
	bw := BytesWritten["devnullsample"]
	ew := EventsWritten["devnullsample"]
	Mutex.RUnlock()
	assert.Greater(t, bw, int64(0), "bytes should be counted even for devnull")
	assert.Equal(t, int64(1), ew)
}

func TestWriteCacheMiss(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "cache miss event"},
	}
	item := makeOutQueueItem("cachemiss", "raw", "stdout", events)
	item.Cache.UseCache = true // UseCache=true but no cacheBuf exists => cache miss

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	// Cache miss should still write through the normal pipe
	assert.Contains(t, result, "cache miss event")
}

func TestWriteSetCache(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	// Clean the cache bufs
	cacheMutex.Lock()
	delete(cacheBufs, "setcache")
	cacheMutex.Unlock()

	events := []map[string]string{
		{"_raw": "cached event"},
	}
	item := makeOutQueueItem("setcache", "raw", "stdout", events)
	item.Cache.SetCache = true

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	// SetCache should write to both cache buffer and the pipe
	assert.Contains(t, result, "cached event")

	// Verify cache buffer was populated
	cacheMutex.RLock()
	cb, ok := cacheBufs["setcache"]
	cacheMutex.RUnlock()
	assert.True(t, ok, "cache buffer should be created")
	assert.Contains(t, cb.String(), "cached event")
}

func TestWriteUseCache(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	// Pre-populate the cache
	cacheMutex.Lock()
	cacheBufs["usecache"] = &bytes.Buffer{}
	cacheBufs["usecache"].WriteString("previously cached data\n")
	cacheMutex.Unlock()

	events := []map[string]string{
		{"_raw": "new event"},
	}
	item := makeOutQueueItem("usecache", "raw", "stdout", events)
	item.Cache.UseCache = true

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	// Should use the cached data, not the new events
	assert.Contains(t, result, "previously cached data")

	// Clean up
	cacheMutex.Lock()
	delete(cacheBufs, "usecache")
	cacheMutex.Unlock()
}

func TestWriteNonExistentTemplate(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "should not appear"},
	}
	item := makeOutQueueItem("badtemplate", "nonexistent_template_xyz", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	// Non-existent template should produce no output
	assert.Empty(t, result)
}

func TestWriteMultipleEvents(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "event1"},
		{"_raw": "event2"},
		{"_raw": "event3"},
	}
	item := makeOutQueueItem("multisample", "raw", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	assert.Contains(t, result, "event1")
	assert.Contains(t, result, "event2")
	assert.Contains(t, result, "event3")

	// Verify accounting
	time.Sleep(50 * time.Millisecond)
	Mutex.RLock()
	ew := EventsWritten["multisample"]
	Mutex.RUnlock()
	assert.Equal(t, int64(3), ew)
}

func TestWriteKafkaNoNewlines(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	events := []map[string]string{
		{"_raw": "kafka event"},
	}
	item := makeOutQueueItem("kafkasample", "raw", "kafka", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	// Kafka should not append newlines
	assert.Equal(t, "kafka event", result)
}

func TestWriteCustomTemplate(t *testing.T) {
	cleanup := initROT()
	defer cleanup()

	// Register a custom template
	_ = template.New("customtest_header", "HEADER\n")
	_ = template.New("customtest_row", "ROW:{{._raw}}\n")
	_ = template.New("customtest_footer", "FOOTER\n")

	events := []map[string]string{
		{"_raw": "custom line"},
	}
	item := makeOutQueueItem("customsample", "customtest", "stdout", events)

	var result string
	done := make(chan struct{})
	go func() {
		result = readFromPipe(item)
		close(done)
	}()

	write(item)
	<-done

	assert.Contains(t, result, "HEADER")
	assert.Contains(t, result, "ROW:custom line")
	assert.Contains(t, result, "FOOTER")
}
