package tests

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestDevNullOutput(t *testing.T) {
	// Create a temporary log file
	tmpFile, err := os.CreateTemp("", "gogen_devnull_test_*.log")
	assert.NoError(t, err)
	logFile := tmpFile.Name()
	tmpFile.Close() // Close it so logger can open it

	// Set up logging to the temp file
	log.SetOutput(logFile)
	log.EnableJSONOutput()
	log.SetInfo()

	// Clean up logging at the end
	defer func() {
		log.EnableTextOutput()
		log.SetDebug(false)
		os.Remove(logFile)
	}()

	configStr := `
global:
  debug: true
  verbose: true
  output:
    outputter: devnull
    outputTemplate: json
samples:
  - name: devnullsample
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
      - sourcetype: devnulltest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Run the sample
	run.Run(c)

	// Wait a bit for the readout thread to complete
	time.Sleep(2 * time.Second)

	// Read the log file contents
	logData, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	logOutput := string(logData)

	// Verify that log messages were written correctly
	assert.Contains(t, logOutput, "Setting outputter")
	assert.Contains(t, logOutput, "devnull")

	// Check for eventSec and kbytesSec > 0 in the JSON logs
	foundEventSec := false
	foundKbytesSec := false
	for _, line := range strings.Split(logOutput, "\n") {
		if line == "" {
			continue
		}
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(line), &logEntry)
		assert.NoError(t, err)

		if rate, ok := logEntry["eventsSec"]; ok {
			if rateFloat, ok := rate.(float64); ok && rateFloat > 0 {
				foundEventSec = true
			}
		}
		if rate, ok := logEntry["kbytesSec"]; ok {
			if rateFloat, ok := rate.(float64); ok && rateFloat > 0 {
				foundKbytesSec = true
			}
		}
		if foundEventSec && foundKbytesSec {
			break
		}
	}
	assert.True(t, foundEventSec, "Expected to find eventSec > 0 in logs")
	assert.True(t, foundKbytesSec, "Expected to find kbytesSec > 0 in logs")

	// Verify that no output was written to the buffer
	assert.Empty(t, c.Buf.String(), "Expected no output in buffer for devnull outputter")

	config.CleanupConfigAndEnvironment()
}
