package run

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestOnceWithConfig(t *testing.T) {
	// Clean up any existing config
	config.ResetConfig()

	// Setup test configuration
	configStr := `
global:
  debug: true
  verbose: true
  utc: true
  output:
    outputter: buf
    outputTemplate: json
samples:
  - name: runoncesample
    description: "Test sample for runOnce functionality"
    begin: now
    end: now
    interval: 1
    count: 1
    tokens:
      - name: ts-dmyhmsms-template
        format: template
        token: $ts$
        type: timestamp
        replacement: "%d/%b/%Y %H:%M:%S.%L"
      - name: tsepoch
        format: template
        token: $epochts$
        field: _time
        type: timestamp
        replacement: "%s.%L"

    lines:
      - sourcetype: runonce_test
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
        field1: test1
        field2: test2
`

	// Setup config and get the instance that will be used
	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Record time before and after test to validate timestamp is within range
	beforeTest := time.Now().Truncate(time.Second)
	if c.Global.UTC {
		beforeTest = beforeTest.UTC()
	}

	// Create a new Runner and run the sample once using the private method
	runner := Runner{}
	runner.onceWithConfig("runoncesample", c)

	afterTest := time.Now().Truncate(time.Second)
	if c.Global.UTC {
		afterTest = afterTest.UTC()
	}

	// Get the sample and its buffer
	output := c.Buf.String()
	t.Logf("Buffer contents: %s", output)

	// Parse the JSON output
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(output), &jsonData)
	assert.NoError(t, err, "Failed to parse JSON output")

	// Validate the expected fields
	expectedFields := []string{"sourcetype", "source", "host", "index", "_time", "_raw", "field1", "field2"}
	for _, field := range expectedFields {
		assert.Contains(t, jsonData, field, "Missing expected field %s in JSON output", field)
	}

	// Validate specific field values
	assert.Equal(t, "runonce_test", jsonData["sourcetype"])
	assert.Equal(t, "gogen", jsonData["source"])
	assert.Equal(t, "gogen", jsonData["host"])
	assert.Equal(t, "main", jsonData["index"])
	assert.Equal(t, "test1", jsonData["field1"])
	assert.Equal(t, "test2", jsonData["field2"])

	// Check _time is within the test execution window
	eventEpoch, err := strconv.ParseFloat(jsonData["_time"].(string), 64)
	assert.NoError(t, err, "Failed to parse _time as float")
	// Truncate to seconds for comparison
	eventTime := time.Unix(int64(eventEpoch), 0).Truncate(time.Second)
	if c.Global.UTC {
		eventTime = eventTime.UTC()
	}

	// The event time should be equal to either beforeTest or afterTest
	// since we truncated all times to seconds
	assert.True(t, eventTime.Equal(beforeTest) || eventTime.Equal(afterTest),
		"Event time %v should equal either test start time %v or end time %v",
		eventTime, beforeTest, afterTest)

	// Parse and check _raw timestamp format
	rawTime, err := time.Parse("02/Jan/2006 15:04:05.000", jsonData["_raw"].(string))
	assert.NoError(t, err, "Failed to parse _raw timestamp")
	if c.Global.UTC {
		rawTime = rawTime.UTC()
	}
	rawTime = rawTime.Truncate(time.Second)
	assert.True(t, rawTime.Equal(eventTime),
		"Raw timestamp %v should match event time %v", rawTime, eventTime)

	// Clean up after test
	config.CleanupConfigAndEnvironment()
}
