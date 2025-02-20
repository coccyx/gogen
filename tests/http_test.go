package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
)

var lastRequest []byte

func setupTestHTTPServer(endpoint string) *http.Server {
	server := &http.Server{Addr: ":8088"}

	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		lastRequest = body
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Give the server a moment to start up
	time.Sleep(100 * time.Millisecond)

	return server
}

func TestHTTPOutput(t *testing.T) {
	server := setupTestHTTPServer("/http")
	defer server.Close()

	configStr := `
global:
  output:
    outputter: http
    outputTemplate: json
    endpoints:
      - http://localhost:8088/http
    headers:
      Authorization: Splunk 00112233-4455-6677-8899-AABBCCDDEEFF
samples:
  - name: outputhttpsample
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
      - sourcetype: httptest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	run.Run(c)

	// Verify the last request was received and formatted correctly
	if len(lastRequest) == 0 {
		t.Fatal("No request received")
	}

	// Parse the JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(lastRequest), &jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Validate the expected fields exist
	expectedFields := []string{"sourcetype", "source", "host", "index", "_time", "_raw"}
	for _, field := range expectedFields {
		if _, ok := jsonData[field]; !ok {
			t.Errorf("Missing expected field %s in JSON output", field)
		}
	}

	// Validate specific field values
	if jsonData["sourcetype"] != "httptest" ||
		jsonData["source"] != "gogen" ||
		jsonData["host"] != "gogen" ||
		jsonData["index"] != "main" {
		t.Error("Basic fields don't match expected values")
	}

	// Check _time is correct epoch for 2001-10-20 00:00:00
	expectedEpoch := "1003561200.000"
	if jsonData["_time"] != expectedEpoch {
		t.Errorf("Expected _time to be %s, got %v", expectedEpoch, jsonData["_time"])
	}

	// Check _raw has correct timestamp format
	expectedRaw := "20/Oct/2001 00:00:00:000"
	if jsonData["_raw"] != expectedRaw {
		t.Errorf("Expected _raw to be %s, got %v", expectedRaw, jsonData["_raw"])
	}

	config.CleanupConfigAndEnvironment()
}

func TestHTTPSplunkOutput(t *testing.T) {
	server := setupTestHTTPServer("/splunk")
	defer server.Close()

	configStr := `
global:
  output:
    outputter: http
    outputTemplate: splunkhec
    endpoints:
      - http://localhost:8088/splunk
    headers:
      Authorization: Splunk 00112233-4455-6677-8899-AABBCCDDEEFF
samples:
  - name: outputhttpsample
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
      - sourcetype: httptest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	run.Run(c)

	if len(lastRequest) == 0 {
		t.Fatal("No request received")
	}

	// Parse the JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal(lastRequest, &jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Validate expected fields for Splunk HEC format
	expectedFields := map[string]string{
		"event":      "20/Oct/2001 00:00:00:000",
		"host":       "gogen",
		"index":      "main",
		"source":     "gogen",
		"sourcetype": "httptest",
		"time":       fmt.Sprintf("%.3f", float64(time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local).Unix())),
	}

	for field, expected := range expectedFields {
		actual, ok := jsonData[field]
		if !ok {
			t.Errorf("Missing expected field %s in HEC output", field)
			continue
		}
		if actual != expected {
			t.Errorf("Field %s: expected %q, got %q", field, expected, actual)
		}
	}

	config.CleanupConfigAndEnvironment()
}
