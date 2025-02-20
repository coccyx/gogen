package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// captureOutput is an internal test helper that temporarily captures log output
func captureOutput(f func()) string {
	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	f()
	logrus.SetOutput(original)
	return buf.String()
}

func TestLogLevel(t *testing.T) {
	// Test default log level
	assert.Equal(t, logrus.ErrorLevel, DefaultLogLevel, "Default log level should be Error")

	// Test setting different log levels and verify through log visibility
	tests := []struct {
		name          string
		setLevel      func()
		shouldLogInfo bool // Info messages should be visible at Info level and below
	}{
		{"Debug", func() { SetDebug(true) }, true},
		{"Info", func() { SetInfo() }, true},
		{"Warn", func() { SetWarn() }, false},
		{"Debug Off", func() { SetDebug(false) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setLevel()
			output := captureOutput(func() {
				Info("test message")
			})
			if tt.shouldLogInfo {
				assert.Contains(t, output, "test message")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLoggingMethods(t *testing.T) {
	SetDebug(true) // Set to debug to capture all messages
	defer SetDebug(false)

	tests := []struct {
		name     string
		logFunc  func()
		contains string
		level    string
	}{
		{
			name:     "Debug",
			logFunc:  func() { Debug("debug message") },
			contains: "debug message",
			level:    "debug",
		},
		{
			name:     "Info",
			logFunc:  func() { Info("info message") },
			contains: "info message",
			level:    "info",
		},
		{
			name:     "Warning",
			logFunc:  func() { Warning("warning message") },
			contains: "warning message",
			level:    "warning",
		},
		{
			name:     "Error",
			logFunc:  func() { Error("error message") },
			contains: "error message",
			level:    "error",
		},
		{
			name:     "Debugf",
			logFunc:  func() { Debugf("debug %s", "formatted") },
			contains: "debug formatted",
			level:    "debug",
		},
		{
			name:     "Infof",
			logFunc:  func() { Infof("info %s", "formatted") },
			contains: "info formatted",
			level:    "info",
		},
		{
			name:     "Warningf",
			logFunc:  func() { Warningf("warning %s", "formatted") },
			contains: "warning formatted",
			level:    "warning",
		},
		{
			name:     "Errorf",
			logFunc:  func() { Errorf("error %s", "formatted") },
			contains: "error formatted",
			level:    "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(tt.logFunc)
			if output == "" {
				t.Logf("Warning: Empty output received for test case %s", tt.name)
			}
			assert.Contains(t, strings.ToLower(output), tt.contains)
			assert.Contains(t, strings.ToLower(output), tt.level)
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// Test Info level filtering
	SetInfo()
	defer SetDebug(false)

	// Debug messages should not appear at Info level
	debugOutput := captureOutput(func() {
		Debug("debug message")
	})
	assert.Empty(t, debugOutput, "Debug message should not appear at Info level")

	// Info messages should appear
	infoOutput := captureOutput(func() {
		Info("info message")
	})
	assert.Contains(t, infoOutput, "info message")
}

func TestWithField(t *testing.T) {
	SetInfo()
	defer SetDebug(false)

	output := captureOutput(func() {
		WithField("single_key", "single_value").Info("single field test")
	})

	assert.Contains(t, output, "single_key")
	assert.Contains(t, output, "single_value")
}

func TestFatal(t *testing.T) {
	// Create a channel to signal if os.Exit was called
	executed := make(chan bool, 1)

	// Save original os.Exit
	originalExit := logrus.StandardLogger().ExitFunc
	defer func() {
		logrus.StandardLogger().ExitFunc = originalExit
	}()

	// Override os.Exit
	logrus.StandardLogger().ExitFunc = func(code int) {
		executed <- true
	}

	output := captureOutput(func() {
		Fatal("fatal message")
	})

	select {
	case <-executed:
		assert.Contains(t, output, "fatal message")
	default:
		t.Error("Fatal did not trigger exit")
	}

	// Test Fatalf
	output = captureOutput(func() {
		Fatalf("fatal %s", "formatted")
	})

	select {
	case <-executed:
		assert.Contains(t, output, "fatal formatted")
	default:
		t.Error("Fatalf did not trigger exit")
	}
}

func TestWithFields(t *testing.T) {
	SetInfo()
	defer SetDebug(false)

	fields := Fields{
		"key1": "value1",
		"key2": 42,
	}

	output := captureOutput(func() {
		WithFields(fields).Info("test message")
	})

	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "42")
}

func TestWithError(t *testing.T) {
	SetInfo()
	defer SetDebug(false)

	testErr := errors.New("test error")

	output := captureOutput(func() {
		WithError(testErr).Error("error occurred")
	})

	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "error occurred")
}

func TestJSONOutput(t *testing.T) {
	EnableJSONOutput()
	defer EnableTextOutput()
	SetInfo()
	defer SetDebug(false)

	output := captureOutput(func() {
		WithField("test_field", "test_value").Info("test message")
	})

	t.Logf("JSON Output received: %s", output)

	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", logEntry["test_field"])
	assert.Equal(t, "test message", logEntry["msg"])
}

func TestFileOutput(t *testing.T) {
	testFile := "test.log"
	SetOutput(testFile)
	SetInfo()
	defer SetDebug(false)

	Info("test file output")

	// Read the file content
	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "test file output")

	// Cleanup
	os.Remove(testFile)
}

func TestContextHook(t *testing.T) {
	EnableJSONOutput()
	defer EnableTextOutput()
	SetDebug(true)
	defer SetDebug(false)

	output := captureOutput(func() {
		Info("test context hook")
	})

	t.Logf("Output received: %s", output)

	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, logEntry["file"], "file field should be present")
		assert.NotEmpty(t, logEntry["func"], "func field should be present")
		assert.NotNil(t, logEntry["line"], "line field should be present")

		t.Logf("File: %v", logEntry["file"])
		t.Logf("Func: %v", logEntry["func"])
		t.Logf("Line: %v", logEntry["line"])
	}
}

func TestPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic as expected")
		}
	}()

	output := captureOutput(func() {
		Panic("test panic")
	})

	assert.Contains(t, output, "test panic")
}
