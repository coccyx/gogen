package tests

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
)

var lastNetworkData []byte

func setupTestTCPServer(port string) (*net.TCPListener, chan bool) {
	addr, err := net.ResolveTCPAddr("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		lastNetworkData = buf[:n]
		done <- true
	}()

	return listener, done
}

func setupTestUDPServer(port string) (*net.UDPConn, chan bool) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	go func() {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		lastNetworkData = buf[:n]
		done <- true
	}()

	return conn, done
}

func TestTCPOutput(t *testing.T) {
	listener, done := setupTestTCPServer("8089")
	defer listener.Close()

	configStr := `
global:
  output:
    outputter: network
    protocol: tcp
    outputTemplate: json
    endpoints:
      - localhost:8089
samples:
  - name: outputnetworksample
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
      - sourcetype: networktest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	run.Run(c)

	// Wait for the server to receive data
	<-done

	// Verify the data was received
	if len(lastNetworkData) == 0 {
		t.Fatal("No data received over TCP")
	}

	// Parse the JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal(lastNetworkData, &jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Validate the expected fields exist and have correct values
	expectedFields := map[string]string{
		"sourcetype": "networktest",
		"source":     "gogen",
		"host":       "gogen",
		"index":      "main",
	}

	for field, expected := range expectedFields {
		actual, ok := jsonData[field]
		if !ok {
			t.Errorf("Missing expected field %s in TCP output", field)
			continue
		}
		if actual != expected {
			t.Errorf("Field %s: expected %q, got %q", field, expected, actual)
		}
	}

	// Check _time is correct epoch for 2001-10-20 00:00:00
	expectedEpoch := fmt.Sprintf("%.3f", float64(time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local).Unix()))
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

func TestUDPOutput(t *testing.T) {
	conn, done := setupTestUDPServer("8090")
	defer conn.Close()

	configStr := `
global:
  output:
    outputter: network
    protocol: udp
    outputTemplate: json
    endpoints:
      - localhost:8090
samples:
  - name: outputnetworksample
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
      - sourcetype: networktest
        source: gogen
        host: gogen
        index: main
        _time: $epochts$
        _raw: $ts$
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	run.Run(c)

	// Wait for the server to receive data
	<-done

	// Verify the data was received
	if len(lastNetworkData) == 0 {
		t.Fatal("No data received over UDP")
	}

	// Parse the JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal(lastNetworkData, &jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Validate the expected fields exist and have correct values
	expectedFields := map[string]string{
		"sourcetype": "networktest",
		"source":     "gogen",
		"host":       "gogen",
		"index":      "main",
	}

	for field, expected := range expectedFields {
		actual, ok := jsonData[field]
		if !ok {
			t.Errorf("Missing expected field %s in UDP output", field)
			continue
		}
		if actual != expected {
			t.Errorf("Field %s: expected %q, got %q", field, expected, actual)
		}
	}

	// Check _time is correct epoch for 2001-10-20 00:00:00
	expectedEpoch := fmt.Sprintf("%.3f", float64(time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local).Unix()))
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

func TestTCPRFC3164Output(t *testing.T) {
	listener, done := setupTestTCPServer("8091")
	defer listener.Close()

	configStr := `
global:
  output:
    outputter: network
    protocol: tcp
    outputTemplate: rfc3164
    endpoints:
      - localhost:8091
samples:
  - name: outputnetworksample
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

    lines:
      - sourcetype: networktest
        source: gogen
        host: gogen
        index: main
        priority: "14"
        tag: "gogen"
        pid: "12345"
        _raw: test message
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	run.Run(c)

	// Wait for the server to receive data
	<-done

	// Verify the data was received
	if len(lastNetworkData) == 0 {
		t.Fatal("No data received over TCP")
	}

	// Get expected timestamp in local time
	expectedTime := time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local)
	expectedTimeStr := expectedTime.Format("Jan 2 15:04:05")

	// RFC3164 format: <priority>timestamp hostname tag[pid]: message
	// Actual format from output: <14>Oct 20 00:00:00 gogen gogen[12345]: test message
	rfc3164Regex := regexp.MustCompile(`^<14>(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+gogen\s+gogen\[12345\]:\s+test message$`)

	if !rfc3164Regex.Match(lastNetworkData) {
		t.Errorf("RFC3164 format mismatch. Got: %s", string(lastNetworkData))
	}

	// Extract and validate the timestamp
	matches := rfc3164Regex.FindSubmatch(lastNetworkData)
	if len(matches) != 2 {
		t.Errorf("Failed to extract timestamp from message: %s", string(lastNetworkData))
	} else {
		gotTimeStr := string(matches[1])
		if gotTimeStr != expectedTimeStr {
			t.Errorf("Timestamp mismatch. Expected: %s, Got: %s", expectedTimeStr, gotTimeStr)
		}
	}

	config.CleanupConfigAndEnvironment()
}

func TestTCPRFC5424Output(t *testing.T) {
	listener, done := setupTestTCPServer("8092")
	defer listener.Close()

	configStr := `
global:
  output:
    outputter: network
    protocol: tcp
    outputTemplate: rfc5424
    endpoints:
      - localhost:8092
samples:
  - name: outputnetworksample
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

    lines:
      - sourcetype: networktest
        source: gogen
        host: gogen
        index: main
        priority: "14"
        appName: "gogen"
        pid: "12345"
        tag: "gogen"
        custom_field1: "value1"
        custom_field2: "value2"
        _raw: test message
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	run.Run(c)

	// Wait for the server to receive data
	<-done

	// Verify the data was received
	if len(lastNetworkData) == 0 {
		t.Fatal("No data received over TCP")
	}

	// Get expected timestamp in local time with offset
	expectedTime := time.Date(2001, 10, 20, 0, 0, 0, 0, time.Local)
	_, offset := expectedTime.Zone()
	offsetHours := offset / 3600
	offsetMinutes := (offset % 3600) / 60
	var expectedTimeStr string
	if offset == 0 {
		// For UTC timezone
		expectedTimeStr = expectedTime.Format("2006-01-02T15:04:05") + "Z"
	} else if offsetHours >= 0 {
		expectedTimeStr = fmt.Sprintf("%s+%02d:%02d", expectedTime.Format("2006-01-02T15:04:05"), offsetHours, offsetMinutes)
	} else {
		expectedTimeStr = fmt.Sprintf("%s-%02d:%02d", expectedTime.Format("2006-01-02T15:04:05"), -offsetHours, offsetMinutes)
	}

	// RFC5424 format: <priority>1 timestamp hostname appname pid - [meta key="value"...] message
	// Extract timestamp with regex - now supporting both offset format and Z format
	timestampRegex := regexp.MustCompile(`^<14>1\s+(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:[-+]\d{2}:\d{2}|Z))\s+`)
	matches := timestampRegex.FindSubmatch(lastNetworkData)
	if len(matches) != 2 {
		t.Errorf("Failed to extract timestamp from message: %s", string(lastNetworkData))
	} else {
		gotTimeStr := string(matches[1])
		if gotTimeStr != expectedTimeStr {
			t.Errorf("Timestamp mismatch. Expected: %s, Got: %s", expectedTimeStr, gotTimeStr)
		}
	}

	// Full message format validation
	rfc5424Regex := regexp.MustCompile(`^<14>1\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:[-+]\d{2}:\d{2}|Z)\s+gogen\s+gogen\s+12345\s+-\s+\[meta\s+(?:[a-zA-Z0-9_]+="[^"]*"\s*)+\]\s+test message$`)

	if !rfc5424Regex.Match(lastNetworkData) {
		t.Errorf("RFC5424 format mismatch. Got: %s", string(lastNetworkData))
	}

	// Validate meta fields
	metaStr := string(lastNetworkData)
	expectedMetaFields := []string{
		`custom_field1="value1"`,
		`custom_field2="value2"`,
		`sourcetype="networktest"`,
		`source="gogen"`,
		`index="main"`,
	}

	for _, field := range expectedMetaFields {
		if !strings.Contains(metaStr, field) {
			t.Errorf("Missing expected meta field: %s", field)
		}
	}

	config.CleanupConfigAndEnvironment()
}
