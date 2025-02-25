package internal

import (
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProfileOff(t *testing.T) {
	assert.Equal(t, false, ProfileOn)
}

func TestSingleton(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	c := NewConfig()
	c2 := NewConfig()
	assert.Equal(t, c, c2)
}

func TestGlobal(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	c := NewConfig()
	output := Output{
		Outputter:      "stdout",
		OutputTemplate: "raw",
		FileName:       "/tmp/test.log",
		Topic:          "defaultTopic",
		MaxBytes:       10485760,
		BackupFiles:    5,
		BufferBytes:    4096,
		Endpoints:      []string(nil),
		Headers:        map[string]string{"Content-Type": "application/json"},
		Timeout:        time.Duration(10 * time.Second),
		channelIdx:     0,
		channelMap:     map[string]int{},
	}
	global := Global{
		Debug:                false,
		Verbose:              false,
		GeneratorWorkers:     1,
		OutputWorkers:        1,
		GeneratorQueueLength: 50,
		OutputQueueLength:    10,
		ROTInterval:          1,
		Output:               output,
		SamplesDir:           []string(nil),
		CacheIntervals:       0,
	}
	assert.Equal(t, global, c.Global)
}

func TestFileOutput(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "fileoutput", "fileoutput.yml"))
	// os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "config", "tests", "fileoutput.yml"))
	c := NewConfig()

	// Test flatten
	assert.Equal(t, "/tmp/fileoutput.log", c.Global.Output.FileName)
	assert.Equal(t, int64(102400), c.Global.Output.MaxBytes)
	assert.Equal(t, "file", c.Global.Output.Outputter)
	assert.Equal(t, "json", c.Global.Output.OutputTemplate)
}

func TestHTTPOutput(t *testing.T) {
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

	SetupFromString(configStr)

	c := NewConfig()

	headers := map[string]string{"Authorization": "Splunk 00112233-4455-6677-8899-AABBCCDDEEFF"}
	endpoints := []string{"http://localhost:8088/http"}
	de := reflect.DeepEqual(headers, c.Global.Output.Headers)
	assert.True(t, de, "Headers do not match: %#v vs %#v", headers, c.Global.Output.Headers)
	de = reflect.DeepEqual(endpoints, c.Global.Output.Endpoints)
	assert.True(t, de, "Endpoints do not match: %#v vs %#v", endpoints, c.Global.Output.Endpoints)

	CleanupConfigAndEnvironment()
}

func TestFlatten(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := filepath.Join("..", "tests", "flatten")
	// os.Setenv("GOGEN_GLOBAL", filepath.Join(home, "config", "global.yml"))
	rand.Seed(0)

	var s *Sample
	s = FindSampleInFile(home, "flatten")
	assert.Equal(t, false, s.Disabled)
	assert.Equal(t, "sample", s.Generator)
	assert.Equal(t, "stdout", s.Output.Outputter)
	assert.Equal(t, "raw", s.Output.OutputTemplate)
	// assert.Equal(t, "config", s.Rater)
	assert.Equal(t, 0, s.Interval)
	assert.Equal(t, 0, s.Count)
	assert.Equal(t, "now", s.Earliest)
	assert.Equal(t, "now", s.Latest)
	// if diff := math.Abs(float64(0.2 - s.RandomizeCount)); diff > 0.000001 {
	// 	t.Fatalf("RandomizeCount not equal")
	// }
	assert.Equal(t, false, s.RandomizeEvents)
	assert.Equal(t, 1, s.EndIntervals)
}

func TestValidate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "validation")
	rand.Seed(0)
	// loc, _ := time.LoadLocation("Local")
	// n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	// now := func() time.Time {
	// 	return n
	// }

	var s *Sample
	checks := []string{
		"validate-lower-upper",
		"validate-upper",
		"validate-string-length",
		"validate-choice-items",
		"validate-weightedchoice-items",
		"validate-fieldchoice-items",
		"validate-fieldchoice-items",
		"validate-fieldchoice-badfield",
		"validate-badrandom",
		"validate-earliest-latest",
		"validate-nolines",
	}
	for _, v := range checks {
		s = FindSampleInFile(home, v)
		assert.Nil(t, s, "%s not nil", v)
	}
}

func TestSinglePass(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "singlepass")
	rand.Seed(0)

	checks := []string{
		"missed-regex",
		"overlapping-regex",
		"wide-regex",
	}
	for _, v := range checks {
		s := FindSampleInFile(home, v)
		assert.False(t, s.SinglePass)
	}

	s := FindSampleInFile(home, "test1")
	assert.True(t, s.SinglePass)
	assert.Len(t, s.BrokenLines[0]["otherfield"], 1)
	assert.Len(t, s.BrokenLines[0]["_raw"], 6)
	assert.Len(t, s.BrokenLines[1]["transtype"], 2)
	assert.Len(t, s.BrokenLines[1]["_raw"], 6)
}

func TestReplay(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "replay")
	rand.Seed(0)

	checks := []string{
		"no-offsets-found",
		"bad-strptime-timestamp",
		"bad-go-timestamp",
		"bad-epoch-timestamp",
	}
	for _, v := range checks {
		s := FindSampleInFile(home, v)
		assert.Nil(t, s)
	}

	s := FindSampleInFile(home, "replay5")
	assert.Equal(t, []time.Duration{(1 * time.Second), (5 * time.Second), (10 * time.Second), (20 * time.Second), 13187500000}, s.ReplayOffsets)
}

func FindSampleInFile(home string, name string) *Sample {
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, name+".yml"))
	c := NewConfig()
	// c.Log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	return c.FindSampleByName(name)
}

func TestWriteFileFromString(t *testing.T) {
	testConfig := `name: test-config
description: test config file
disabled: false`

	filename := WriteTempConfigFileFromString(testConfig)
	defer os.Remove(filename) // Clean up after test

	// Verify file exists
	_, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Expected file %s to exist, got error: %v", filename, err)
	}

	// Read contents and verify
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Error reading file %s: %v", filename, err)
	}

	if string(contents) != testConfig {
		t.Errorf("File contents do not match. Expected:\n%s\nGot:\n%s", testConfig, string(contents))
	}
}
func TestSetupFromString(t *testing.T) {
	testConfig := `name: test-setup
description: test setup config
disabled: false`

	// Run setup
	SetupFromString(testConfig)
	defer CleanupConfigAndEnvironment() // Clean up environment after test

	// Verify environment variables were set correctly
	if os.Getenv("GOGEN_HOME") != ".." {
		t.Errorf("Expected GOGEN_HOME to be '..', got '%s'", os.Getenv("GOGEN_HOME"))
	}

	if os.Getenv("GOGEN_ALWAYS_REFRESH") != "1" {
		t.Errorf("Expected GOGEN_ALWAYS_REFRESH to be '1', got '%s'", os.Getenv("GOGEN_ALWAYS_REFRESH"))
	}

	// Verify config file was created and contains correct content
	configFile := os.Getenv("GOGEN_FULLCONFIG")
	if configFile == "" {
		t.Fatal("Expected GOGEN_FULLCONFIG to be set")
	}

	contents, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Error reading config file: %v", err)
	}

	if string(contents) != testConfig {
		t.Errorf("Config file contents do not match. Expected:\n%s\nGot:\n%s", testConfig, string(contents))
	}
}

func TestCleanup(t *testing.T) {
	// Set up test environment variables
	os.Setenv("GOGEN_HOME", "test-home")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")

	// Create a temporary config file
	configFile := WriteTempConfigFileFromString("test config content")
	os.Setenv("GOGEN_FULLCONFIG", configFile)

	// Run cleanup
	CleanupConfigAndEnvironment()

	// Verify environment variables were unset
	if val := os.Getenv("GOGEN_HOME"); val != "" {
		t.Errorf("Expected GOGEN_HOME to be unset, got '%s'", val)
	}

	if val := os.Getenv("GOGEN_ALWAYS_REFRESH"); val != "" {
		t.Errorf("Expected GOGEN_ALWAYS_REFRESH to be unset, got '%s'", val)
	}

	if val := os.Getenv("GOGEN_FULLCONFIG"); val != "" {
		t.Errorf("Expected GOGEN_FULLCONFIG to be unset, got '%s'", val)
	}

	// Verify config file was deleted
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("Expected config file to be deleted")
	}
}

func TestParseWebConfig(t *testing.T) {
	// Set up test environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "https://gist.githubusercontent.com/coccyx/98d5b83307b0b85c1c7a54a08bfec8ed/raw/1ea26d1a16ffeeb113931e696e22b17f0eb0dc81/config.yaml")
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	// Validate global settings
	assert.Equal(t, false, c.Global.Debug)
	assert.Equal(t, 1, c.Global.GeneratorWorkers)
	assert.Equal(t, 1, c.Global.OutputWorkers)
	assert.Equal(t, "stdout", c.Global.Output.Outputter)
	assert.Equal(t, "raw", c.Global.Output.OutputTemplate)

	// Validate sample configuration
	assert.Equal(t, 1, len(c.Samples))
	sample := c.Samples[0]
	assert.Equal(t, "weblog", sample.Name)
	assert.Equal(t, 10, sample.Count)
	assert.Equal(t, 1, sample.Interval)
	assert.Equal(t, "now", sample.Earliest)
	assert.Equal(t, "now", sample.Latest)
	assert.Equal(t, true, sample.RandomizeEvents)
	assert.Equal(t, "sample", sample.Generator)

	// Validate tokens
	assert.Equal(t, 8, len(sample.Tokens))
	tsToken := sample.Tokens[0]
	assert.Equal(t, "ts-dmyhmsms-template", tsToken.Name)
	assert.Equal(t, "timestamp", tsToken.Type)
	assert.Equal(t, "%d/%b/%Y %H:%M:%S:%L", tsToken.Replacement)
}
