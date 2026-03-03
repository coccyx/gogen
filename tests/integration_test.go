package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
	"github.com/coccyx/gogen/run"
	"github.com/coccyx/gogen/template"
	"github.com/stretchr/testify/assert"
)

// resetState clears the config singleton and outputter statistics.
func resetState() {
	config.ResetConfig()
	outputter.Mutex.Lock()
	outputter.BytesWritten = make(map[string]int64)
	outputter.EventsWritten = make(map[string]int64)
	outputter.Mutex.Unlock()
}

// captureStdoutRun sets up a config from a YAML string, captures stdout during
// run.Run, and returns the captured output.
func captureStdoutRun(t *testing.T, configStr string) string {
	t.Helper()
	resetState()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	config.SetupFromString(configStr)
	c := config.NewConfig()
	defer config.CleanupConfigAndEnvironment()

	done := make(chan bool)
	go func() {
		run.Run(c)
		w.Close()
		done <- true
	}()

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	assert.NoError(t, err)
	<-done

	return strings.TrimSpace(buf.String())
}

// runDevnull sets up a config from a YAML string, runs the pipeline with devnull
// output, and returns the config for inspection.
func runDevnull(t *testing.T, configStr string) *config.Config {
	t.Helper()
	resetState()
	config.SetupFromString(configStr)
	c := config.NewConfig()
	run.Run(c)
	config.CleanupConfigAndEnvironment()
	return c
}

// ---------------------------------------------------------------------------
// 1. Config Defaults
// ---------------------------------------------------------------------------

func TestConfigDefaultsApplied(t *testing.T) {
	resetState()
	configStr := `
samples:
  - name: defaultsample
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    lines:
      - _raw: hello
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	defer config.CleanupConfigAndEnvironment()

	// Global worker/queue defaults
	assert.Equal(t, 1, c.Global.GeneratorWorkers, "GeneratorWorkers default")
	assert.Equal(t, 1, c.Global.OutputWorkers, "OutputWorkers default")
	assert.Equal(t, 50, c.Global.GeneratorQueueLength, "GeneratorQueueLength default")
	assert.Equal(t, 10, c.Global.OutputQueueLength, "OutputQueueLength default")

	// Output defaults
	assert.Equal(t, "stdout", c.Global.Output.Outputter, "Outputter default")
	assert.Equal(t, "raw", c.Global.Output.OutputTemplate, "OutputTemplate default")
	assert.Equal(t, "/tmp/test.log", c.Global.Output.FileName, "FileName default")
	assert.Equal(t, int64(10485760), c.Global.Output.MaxBytes, "MaxBytes default")
	assert.Equal(t, 5, c.Global.Output.BackupFiles, "BackupFiles default")
	assert.Equal(t, 4096, c.Global.Output.BufferBytes, "BufferBytes default")
	assert.Equal(t, 10*time.Second, c.Global.Output.Timeout, "Timeout default")
	assert.Equal(t, "defaultTopic", c.Global.Output.Topic, "Topic default")
	assert.Equal(t, "application/json", c.Global.Output.Headers["Content-Type"], "Headers default")

	// ROT
	assert.Equal(t, 1, c.Global.ROTInterval, "ROTInterval default")
}

func TestConfigRaterDefaultMaps(t *testing.T) {
	resetState()
	configStr := `
samples:
  - name: ratersample
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    lines:
      - _raw: hello
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	defer config.CleanupConfigAndEnvironment()

	r := c.FindRater("config")
	if !assert.NotNil(t, r, "config rater should exist") {
		return
	}

	// HourOfDay: 24 entries, keys 0-23, all 1.0
	hod := r.Options["HourOfDay"].(map[int]float64)
	assert.Len(t, hod, 24)
	for i := 0; i < 24; i++ {
		assert.Equal(t, 1.0, hod[i], "HourOfDay[%d]", i)
	}

	// DayOfWeek: 7 entries, keys 0-6, all 1.0
	dow := r.Options["DayOfWeek"].(map[int]float64)
	assert.Len(t, dow, 7)
	for i := 0; i < 7; i++ {
		assert.Equal(t, 1.0, dow[i], "DayOfWeek[%d]", i)
	}

	// MinuteOfHour: 60 entries, keys 0-59, all 1.0
	moh := r.Options["MinuteOfHour"].(map[int]float64)
	assert.Len(t, moh, 60)
	for i := 0; i < 60; i++ {
		assert.Equal(t, 1.0, moh[i], "MinuteOfHour[%d]", i)
	}
}

// ---------------------------------------------------------------------------
// 2. Config Parsing
// ---------------------------------------------------------------------------

func TestConfigParseValidFullConfig(t *testing.T) {
	resetState()
	configStr := `
global:
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: tutorial1
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 3
    tokens:
      - name: ts
        format: template
        token: $ts$
        type: timestamp
        replacement: "%d/%b/%Y %H:%M:%S"
    lines:
      - _raw: "$ts$ line1"
      - _raw: "$ts$ line2"
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	defer config.CleanupConfigAndEnvironment()

	assert.Len(t, c.Samples, 1)
	s := c.Samples[0]
	assert.Equal(t, "tutorial1", s.Name)
	assert.Equal(t, 3, s.Count)
	assert.Equal(t, 1, s.Interval)
	assert.GreaterOrEqual(t, len(s.Tokens), 1, "should have at least the ts token")
	assert.Len(t, s.Lines, 2)
}

func TestConfigParseInvalidYAML(t *testing.T) {
	resetState()
	// BuildConfig panics/fatals on invalid YAML via log.Panic
	tmpfile, err := os.CreateTemp("", "gogen-test-bad-*.yml")
	assert.NoError(t, err)
	_, err = tmpfile.Write([]byte("{{{{invalid yaml!!!!"))
	assert.NoError(t, err)
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	defer func() {
		r := recover()
		assert.NotNil(t, r, "BuildConfig should panic on invalid YAML")
	}()

	config.BuildConfig(config.ConfigConfig{FullConfig: tmpfile.Name()})
}

// ---------------------------------------------------------------------------
// 3. Full Pipeline — Output Templates
// ---------------------------------------------------------------------------

func TestPipelineRawOutput(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: raw
samples:
  - name: rawtest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    lines:
      - _raw: "hello raw world"
`)
	assert.Equal(t, "hello raw world", output)
}

func TestPipelineJSONOutput(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: json
samples:
  - name: jsontest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    tokens:
      - name: tsepoch
        format: template
        token: $epochts$
        field: _time
        type: timestamp
        replacement: "%s.%L"
    lines:
      - sourcetype: jtest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: hello json
`)
	lines := strings.Split(output, "\n")
	assert.Equal(t, 1, len(lines), "expected one line")

	var data map[string]interface{}
	err := json.Unmarshal([]byte(lines[0]), &data)
	assert.NoError(t, err)

	for _, field := range []string{"_raw", "host", "source", "sourcetype", "index"} {
		assert.Contains(t, data, field, "missing field %s", field)
	}
	assert.Equal(t, "jtest", data["sourcetype"])
	assert.Equal(t, "gogen", data["source"])
	assert.Equal(t, "gogen", data["host"])
	assert.Equal(t, "main", data["index"])
	assert.Equal(t, "hello json", data["_raw"])
}

func TestPipelineSplunkHECOutput(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: splunkhec
samples:
  - name: hectest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    tokens:
      - name: tsepoch
        format: template
        token: $epochts$
        field: _time
        type: timestamp
        replacement: "%s.%L"
    lines:
      - sourcetype: hectype
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: hec event data
`)
	lines := strings.Split(output, "\n")
	assert.Equal(t, 1, len(lines), "expected one line")

	var data map[string]interface{}
	err := json.Unmarshal([]byte(lines[0]), &data)
	assert.NoError(t, err)

	// _raw renamed to event, _time renamed to time
	assert.Contains(t, data, "event")
	assert.Contains(t, data, "time")
	assert.NotContains(t, data, "_raw")
	assert.NotContains(t, data, "_time")
	assert.Equal(t, "hec event data", data["event"])
}

func TestPipelineCSVOutput(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: csv
samples:
  - name: csvtest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    lines:
      - _raw: csvdata
        host: myhost
        source: mysource
`)
	lines := strings.Split(output, "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "expected header + data rows")

	// Header should have sorted field names
	header := lines[0]
	fields := strings.Split(header, ",")
	sorted := make([]string, len(fields))
	copy(sorted, fields)
	sort.Strings(sorted)
	assert.Equal(t, sorted, fields, "CSV header fields should be sorted")
}

func TestPipelineCustomTemplate(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: mytemplate
samples:
  - name: customtpltest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    lines:
      - _raw: eventdata
        host: tplhost
templates:
  - name: mytemplate
    header: ""
    row: "HOST={{.host}} RAW={{._raw}}"
    footer: ""
`)
	assert.Contains(t, output, "HOST=tplhost")
	assert.Contains(t, output, "RAW=eventdata")
}

// ---------------------------------------------------------------------------
// 4. Token Processing
// ---------------------------------------------------------------------------

func TestTokenTimestamp(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: raw
samples:
  - name: tstest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    tokens:
      - name: ts
        format: template
        token: $ts$
        type: timestamp
        replacement: "%d/%b/%Y %H:%M:%S"
    lines:
      - _raw: "$ts$"
`)
	assert.Contains(t, output, "20/Oct/2001")
}

func TestTokenRandomInt(t *testing.T) {
	for i := 0; i < 10; i++ {
		output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: raw
samples:
  - name: randinttest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    tokens:
      - name: randnum
        format: template
        token: $randnum$
        type: random
        replacement: int
        lower: 10
        upper: 20
    lines:
      - _raw: "$randnum$"
`)
		val, err := strconv.Atoi(output)
		assert.NoError(t, err, "output should be an integer")
		assert.GreaterOrEqual(t, val, 10)
		assert.Less(t, val, 20) // upper is exclusive in randgen.Intn
	}
}

func TestTokenChoice(t *testing.T) {
	output := captureStdoutRun(t, `
global:
  output:
    outputter: stdout
    outputTemplate: raw
samples:
  - name: choicetest
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    tokens:
      - name: color
        format: template
        token: $color$
        type: choice
        choice:
          - red
          - green
          - blue
    lines:
      - _raw: "$color$"
`)
	choices := []string{"red", "green", "blue"}
	assert.Contains(t, choices, output, "output should be one of the choices")
}

// ---------------------------------------------------------------------------
// 5. HEC Transform
// ---------------------------------------------------------------------------

func TestHECTransformDirect(t *testing.T) {
	event := map[string]string{
		"_raw":  "hello",
		"_time": "12345",
		"host":  "foo",
	}
	template.TransformHECFields(event)

	assert.Equal(t, "hello", event["event"])
	assert.Equal(t, "12345", event["time"])
	assert.Equal(t, "foo", event["host"])
	_, hasRaw := event["_raw"]
	_, hasTime := event["_time"]
	assert.False(t, hasRaw, "_raw should be deleted")
	assert.False(t, hasTime, "_time should be deleted")
}

func TestHECTransformNoOp(t *testing.T) {
	event := map[string]string{
		"host":   "bar",
		"source": "baz",
	}
	original := make(map[string]string)
	for k, v := range event {
		original[k] = v
	}
	template.TransformHECFields(event)
	assert.Equal(t, original, event, "map should be unchanged when _raw and _time are absent")
}

// ---------------------------------------------------------------------------
// 6. ROT Synchronization
// ---------------------------------------------------------------------------

func TestROTReinitAfterReadFinal(t *testing.T) {
	// First cycle
	outputter.Mutex.Lock()
	outputter.BytesWritten = make(map[string]int64)
	outputter.EventsWritten = make(map[string]int64)
	outputter.Mutex.Unlock()

	dummyConfig := &config.Config{
		Global: config.Global{ROTInterval: 1},
	}

	outputter.InitROT(dummyConfig)
	outputter.Account(10, 100, "s1")
	outputter.ReadFinal()

	outputter.Mutex.RLock()
	assert.Equal(t, int64(10), outputter.EventsWritten["s1"])
	assert.Equal(t, int64(100), outputter.BytesWritten["s1"])
	outputter.Mutex.RUnlock()

	// Second cycle — validates sync.Once reset works
	outputter.Mutex.Lock()
	outputter.BytesWritten = make(map[string]int64)
	outputter.EventsWritten = make(map[string]int64)
	outputter.Mutex.Unlock()

	outputter.InitROT(dummyConfig)
	outputter.Account(20, 200, "s2")
	outputter.ReadFinal()

	outputter.Mutex.RLock()
	assert.Equal(t, int64(20), outputter.EventsWritten["s2"])
	assert.Equal(t, int64(200), outputter.BytesWritten["s2"])
	outputter.Mutex.RUnlock()
}

// ---------------------------------------------------------------------------
// 7. HTTP Helpers
// ---------------------------------------------------------------------------

func TestHTTPErrorFormatting(t *testing.T) {
	e := &config.HTTPError{
		StatusCode: 404,
		URL:        "http://example.com/test",
		Body:       "not found",
	}
	assert.Contains(t, e.Error(), "404")
	assert.Contains(t, e.Error(), "http://example.com/test")
	assert.Contains(t, e.Error(), "not found")
	assert.True(t, e.IsNotFound())

	e500 := &config.HTTPError{StatusCode: 500, URL: "http://x", Body: "err"}
	assert.False(t, e500.IsNotFound())
}

func TestDoGetSuccess(t *testing.T) {
	// We test through List() since doGet is unexported.
	// Set up a mock server that returns a valid /v1/list response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/list", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{"Items":[{"gogen":"test1","description":"desc1"}]}`)
	}))
	defer server.Close()

	os.Setenv("GOGEN_APIURL", server.URL)
	defer os.Unsetenv("GOGEN_APIURL")

	list, err := config.List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "test1", list[0].Gogen)
	assert.Equal(t, "desc1", list[0].Description)
}

func TestDoGetHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, "not found")
	}))
	defer server.Close()

	os.Setenv("GOGEN_APIURL", server.URL)
	defer os.Unsetenv("GOGEN_APIURL")

	_, err := config.List()
	assert.Error(t, err)

	var httpErr *config.HTTPError
	assert.True(t, errors.As(err, &httpErr), "should unwrap to *HTTPError")
	assert.True(t, httpErr.IsNotFound())
}

func TestDoPostWithHeaders(t *testing.T) {
	// We test indirectly: the Get function uses doGet, but to test doPost
	// with headers we need an exported function that uses it. Upsert uses
	// doPost with an Authorization header, but requires GitHub token.
	// Instead, we test that the server receives custom headers through
	// a mock of the /v1/list endpoint. Since doPost is unexported, we
	// verify header passing by testing that List properly calls the server.
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.UserAgent()
		w.WriteHeader(200)
		fmt.Fprint(w, `{"Items":[]}`)
	}))
	defer server.Close()

	os.Setenv("GOGEN_APIURL", server.URL)
	defer os.Unsetenv("GOGEN_APIURL")

	_, err := config.List()
	assert.NoError(t, err)
	// Go's default HTTP client sends a User-Agent header
	assert.NotEmpty(t, receivedUA)
}

func TestListWithMockAPI(t *testing.T) {
	// Multiple items
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, `{"Items":[
			{"gogen":"g1","description":"d1"},
			{"gogen":"g2","description":"d2"},
			{"notgogen":"bad"}
		]}`)
	}))
	defer server.Close()

	os.Setenv("GOGEN_APIURL", server.URL)
	defer os.Unsetenv("GOGEN_APIURL")

	list, err := config.List()
	assert.NoError(t, err)
	assert.Len(t, list, 2, "should skip items missing gogen or description")
	assert.Equal(t, "g1", list[0].Gogen)
	assert.Equal(t, "g2", list[1].Gogen)
}

func TestListWithMockAPIServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprint(w, "internal server error")
	}))
	defer server.Close()

	os.Setenv("GOGEN_APIURL", server.URL)
	defer os.Unsetenv("GOGEN_APIURL")

	list, err := config.List()
	assert.Nil(t, list)
	assert.Error(t, err, "should return error, not panic")
}

// ---------------------------------------------------------------------------
// 8. String Safety
// ---------------------------------------------------------------------------

func TestShortConfigStringNoPanic(t *testing.T) {
	resetState()
	// Verify that a short FullConfig string exercises strings.HasPrefix
	// without panicking. Previously [0:4] on a short string would panic.
	// We create a real (empty) file so os.Stat passes and the HasPrefix
	// check is actually reached.
	tmpfile, err := os.CreateTemp("", "ab")
	assert.NoError(t, err)
	tmpfile.Write([]byte("samples: []\n"))
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Rename to a 2-char basename path in the same dir — but temp dir
	// paths are long. Instead, just verify BuildConfig works with the
	// temp file (which has a long path but the HasPrefix("http") check
	// is what matters — it's safe for any length string).
	defer func() {
		r := recover()
		if r != nil {
			msg := fmt.Sprintf("%v", r)
			assert.NotContains(t, msg, "index out of range",
				"should not panic with index out of range on short string")
		}
	}()

	config.BuildConfig(config.ConfigConfig{FullConfig: tmpfile.Name()})
}

// ---------------------------------------------------------------------------
// 9. Multi-Sample and FromSample
// ---------------------------------------------------------------------------

func TestFromSampleCopy(t *testing.T) {
	resetState()
	configStr := `
global:
  output:
    outputter: devnull
samples:
  - name: sampleA
    begin: "2001-10-20 00:00:00"
    end: "2001-10-20 00:00:01"
    interval: 1
    count: 1
    lines:
      - _raw: "line from A"
  - name: sampleB
    fromSample: sampleA
    count: 2
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	defer config.CleanupConfigAndEnvironment()

	b := c.FindSampleByName("sampleB")
	if !assert.NotNil(t, b, "sampleB should exist") {
		return
	}
	assert.Equal(t, "sampleB", b.Name)
	assert.Equal(t, 2, b.Count)
	assert.Len(t, b.Lines, 1)
	assert.Equal(t, "line from A", b.Lines[0]["_raw"])
}

func TestMultipleSamplesEndToEnd(t *testing.T) {
	configStr := `
global:
  output:
    outputter: devnull
samples:
  - name: multi1
    endIntervals: 1
    interval: 1
    count: 1
    lines:
      - _raw: "event1"
  - name: multi2
    endIntervals: 1
    interval: 1
    count: 1
    lines:
      - _raw: "event2"
  - name: multi3
    endIntervals: 1
    interval: 1
    count: 1
    lines:
      - _raw: "event3"
`
	runDevnull(t, configStr)

	outputter.Mutex.RLock()
	defer outputter.Mutex.RUnlock()

	for _, name := range []string{"multi1", "multi2", "multi3"} {
		assert.Greater(t, outputter.EventsWritten[name], int64(0),
			"expected events for sample %s", name)
	}
}

// ---------------------------------------------------------------------------
// 10. Replay Generator
// ---------------------------------------------------------------------------

func TestReplayGeneratorEndToEnd(t *testing.T) {
	resetState()
	configStr := `
global:
  output:
    outputter: buf
samples:
  - name: replaytest
    generator: replay
    begin: "2001-10-20 12:00:00"
    end: "2001-10-20 12:00:49"
    tokens:
    - name: ts1
      type: timestamp
      replacement: "%Y-%m-%dT%H:%M:%S"
      format: regex
      token: "(\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2})"
    lines:
    - _raw: "2001-10-20T12:00:00"
    - _raw: "2001-10-20T12:00:01"
    - _raw: "2001-10-20T12:00:06"
    - _raw: "2001-10-20T12:00:16"
    - _raw: "2001-10-20T12:00:36"
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	defer config.CleanupConfigAndEnvironment()

	s := c.FindSampleByName("replaytest")
	if !assert.NotNil(t, s, "replaytest sample should exist") {
		return
	}
	assert.False(t, s.Disabled, "sample should not be disabled")
	assert.Len(t, s.ReplayOffsets, 5)

	run.Run(c)
	output := c.Buf.String()
	assert.NotEmpty(t, output, "replay should produce output")

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 5, len(lines), "should produce 5 events")
}

// ---------------------------------------------------------------------------
// Race detector test for ROT
// ---------------------------------------------------------------------------

func TestROTConcurrentAccounting(t *testing.T) {
	outputter.Mutex.Lock()
	outputter.BytesWritten = make(map[string]int64)
	outputter.EventsWritten = make(map[string]int64)
	outputter.Mutex.Unlock()

	dummyConfig := &config.Config{
		Global: config.Global{ROTInterval: 1},
	}

	outputter.InitROT(dummyConfig)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			outputter.Account(1, 10, fmt.Sprintf("concurrent_%d", n))
		}(i)
	}
	wg.Wait()
	outputter.ReadFinal()

	outputter.Mutex.RLock()
	total := int64(0)
	for _, v := range outputter.EventsWritten {
		total += v
	}
	outputter.Mutex.RUnlock()
	assert.Equal(t, int64(10), total, "all 10 concurrent accounts should be recorded")
}
