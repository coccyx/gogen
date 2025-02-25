package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestStdoutJSONOutput(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Clean up at the end
	defer func() {
		os.Stdout = oldStdout
	}()

	configStr := `
global:
  debug: false
  verbose: false
  output:
    outputter: stdout
    outputTemplate: json
samples:
  - name: stdoutsample
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 1
    tokens:
      - name: ts-dmyhmsms-template
        format: template
        token: $ts$
        type: timestamp
        replacement: "%d/%b/%Y %H:%M:%S:%L"
      - name: tsepoch
        format: template
        token: $epochts$
        field: _time
        type: timestamp
        replacement: "%s.%L"

    lines:
      - sourcetype: stdouttest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
        field1: value1
        field2: value2
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Create a channel to signal when Run is complete
	done := make(chan bool)

	// Run in a goroutine so we can coordinate output capture
	go func() {
		run.Run(c)
		w.Close() // Close the write end after Run completes
		done <- true
	}()

	// Read the captured output
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	assert.NoError(t, err)

	// Wait for Run to complete
	<-done

	output := strings.TrimSpace(buf.String())

	// There should be exactly one line of output
	lines := strings.Split(output, "\n")
	assert.Equal(t, 1, len(lines), "Expected exactly one line of output")

	// Parse the JSON output
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(lines[0]), &jsonData)
	assert.NoError(t, err, "Failed to parse JSON output")

	// Validate the expected fields
	expectedFields := []string{"sourcetype", "source", "host", "index", "_time", "_raw", "field1", "field2"}
	for _, field := range expectedFields {
		assert.Contains(t, jsonData, field, "Missing expected field %s in JSON output", field)
	}

	// Validate specific field values
	assert.Equal(t, "stdouttest", jsonData["sourcetype"])
	assert.Equal(t, "gogen", jsonData["source"])
	assert.Equal(t, "gogen", jsonData["host"])
	assert.Equal(t, "main", jsonData["index"])
	assert.Equal(t, "value1", jsonData["field1"])
	assert.Equal(t, "value2", jsonData["field2"])

	// Check _time is correct epoch for 2001-10-20 00:00:00
	expectedEpoch := fmt.Sprintf("%.3f", float64(time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local).Unix()))
	assert.Equal(t, expectedEpoch, jsonData["_time"])

	// Check _raw has correct timestamp format
	expectedRaw := "20/Oct/2001 00:00:00:000"
	assert.Equal(t, expectedRaw, jsonData["_raw"])

	config.CleanupConfigAndEnvironment()
}

func TestStdoutCustomTemplateOutput(t *testing.T) {
	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Clean up at the end
	defer func() {
		os.Stdout = oldStdout
	}()

	configStr := `
global:
  debug: false
  verbose: false
  output:
    outputter: stdout
    outputTemplate: customtemplate
samples:
  - name: stdouttemplatesample
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 1
    tokens:
      - name: ts-dmyhmsms-template
        format: template
        token: $ts$
        type: timestamp
        replacement: "%d/%b/%Y %H:%M:%S:%L"
      - name: tsepoch
        format: template
        token: $epochts$
        field: _time
        type: timestamp
        replacement: "%s.%L"

    lines:
      - sourcetype: stdouttest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
        field1: custom1
        field2: custom2

templates:
  - name: customtemplate
    header: ""
    row: "{{.host}} [{{._time}}] {{.sourcetype}}: {{._raw}} (fields: {{.field1}}, {{.field2}})"
    footer: ""
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Create a channel to signal when Run is complete
	done := make(chan bool)

	// Run in a goroutine so we can coordinate output capture
	go func() {
		run.Run(c)
		w.Close() // Close the write end after Run completes
		done <- true
	}()

	// Read the captured output
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	assert.NoError(t, err)

	// Wait for Run to complete
	<-done

	output := strings.TrimSpace(buf.String())

	// There should be exactly one line of output
	lines := strings.Split(output, "\n")
	assert.Equal(t, 1, len(lines), "Expected exactly one line of output")

	// Build expected output string
	expectedTime := fmt.Sprintf("%.3f", float64(time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local).Unix()))
	expectedOutput := fmt.Sprintf("gogen [%s] stdouttest: 20/Oct/2001 00:00:00:000 (fields: custom1, custom2)", expectedTime)

	// Validate the output matches our expected format
	assert.Equal(t, expectedOutput, output, "Output does not match expected format")

	config.CleanupConfigAndEnvironment()
}
