package run

import (
	"bytes"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
	"github.com/stretchr/testify/assert"
)

// resetRunState resets config and outputter stats for a clean test.
func resetRunState() {
	config.ResetConfig()
	outputter.Mutex.Lock()
	outputter.BytesWritten = make(map[string]int64)
	outputter.EventsWritten = make(map[string]int64)
	outputter.Mutex.Unlock()
}

func TestRunCompletesWithEndIntervals(t *testing.T) {
	resetRunState()

	configStr := `
global:
  utc: true
  output:
    outputter: devnull
    outputTemplate: raw
  rotInterval: 1
samples:
  - name: runtest
    description: "Run completion test"
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: run test event
`
	config.SetupFromString(configStr)
	defer config.CleanupConfigAndEnvironment()

	c := config.NewConfig()
	assert.NotEmpty(t, c.Samples)

	done := make(chan struct{})
	go func() {
		Run(c)
		close(done)
	}()

	select {
	case <-done:
		// Run completed successfully
	case <-time.After(10 * time.Second):
		t.Fatal("Run() did not complete within timeout")
	}
}

func TestRunMultipleSamples(t *testing.T) {
	resetRunState()

	configStr := `
global:
  utc: true
  output:
    outputter: devnull
    outputTemplate: raw
  rotInterval: 1
samples:
  - name: multi1
    description: "Multi sample 1"
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: event from multi1
  - name: multi2
    description: "Multi sample 2"
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: event from multi2
`
	config.SetupFromString(configStr)
	defer config.CleanupConfigAndEnvironment()

	c := config.NewConfig()
	assert.Len(t, c.Samples, 2)

	done := make(chan struct{})
	go func() {
		Run(c)
		close(done)
	}()

	select {
	case <-done:
		outputter.Mutex.RLock()
		totalEvents := int64(0)
		for _, v := range outputter.EventsWritten {
			totalEvents += v
		}
		outputter.Mutex.RUnlock()
		assert.Greater(t, totalEvents, int64(0), "should have generated events")
	case <-time.After(10 * time.Second):
		t.Fatal("Run() did not complete within timeout")
	}
}

func TestOnceMethod(t *testing.T) {
	resetRunState()

	configStr := `
global:
  utc: true
  output:
    outputter: buf
    outputTemplate: json
  rotInterval: 1
samples:
  - name: oncemethodtest
    description: "Once method test"
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: once method event
`
	config.SetupFromString(configStr)
	defer config.CleanupConfigAndEnvironment()

	r := Runner{}

	done := make(chan struct{})
	go func() {
		r.Once("oncemethodtest")
		close(done)
	}()

	select {
	case <-done:
		// Once completed without error
	case <-time.After(10 * time.Second):
		t.Fatal("Once() did not complete within timeout")
	}
}

func TestOncePublic(t *testing.T) {
	resetRunState()

	configStr := `
global:
  utc: true
  output:
    outputter: buf
    outputTemplate: json
  rotInterval: 1
samples:
  - name: oncetest
    description: "Once test sample"
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: once event data
`
	config.SetupFromString(configStr)
	defer config.CleanupConfigAndEnvironment()

	c := config.NewConfig()
	assert.NotEmpty(t, c.Samples)

	// Set up a buffer for the sample
	var buf bytes.Buffer
	s := c.FindSampleByName("oncetest")
	s.Buf = &buf

	r := Runner{}

	// Start ROT before onceWithConfig (Once() normally does this)
	go outputter.ROT(c)
	// Give ROT goroutine time to create the new rotchan
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		r.onceWithConfig("oncetest", c)
		close(done)
	}()

	select {
	case <-done:
		assert.Contains(t, buf.String(), "once event data")
	case <-time.After(10 * time.Second):
		t.Fatal("Once() did not complete within timeout")
	}
}
