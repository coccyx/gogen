package internal

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
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

func TestFindRater(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "rater", "configrater.yml"))

	c := NewConfig()

	r := c.FindRater("testconfigrater")
	assert.NotNil(t, r)
	assert.Equal(t, "testconfigrater", r.Name)

	CleanupConfigAndEnvironment()
}

func TestFindRaterNotFound(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "rater", "configrater.yml"))

	c := NewConfig()

	r := c.FindRater("nonexistentrater")
	assert.Nil(t, r)

	CleanupConfigAndEnvironment()
}

func TestFindSampleByNameNotFound(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "rater", "configrater.yml"))

	c := NewConfig()

	s := c.FindSampleByName("nonexistentsample")
	assert.Nil(t, s)

	CleanupConfigAndEnvironment()
}

func TestClean(t *testing.T) {
	configStr := `
samples:
  - name: enabled-sample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test
  - name: disabled-sample
    disabled: true
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	ResetConfig()
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()

	// After Clean(), only enabled real samples should remain
	found := false
	for _, s := range c.Samples {
		if s.Name == "disabled-sample" {
			found = true
		}
	}
	assert.False(t, found, "disabled sample should be removed by Clean()")

	foundEnabled := false
	for _, s := range c.Samples {
		if s.Name == "enabled-sample" {
			foundEnabled = true
		}
	}
	assert.True(t, foundEnabled, "enabled sample should remain after Clean()")
}

func TestParseBeginEndWithEndIntervals(t *testing.T) {
	s := &Sample{
		Name:         "test",
		EndIntervals: 3,
		Interval:     5,
	}

	ParseBeginEnd(s)

	assert.Equal(t, "-15s", s.Begin)
	assert.Equal(t, "now", s.End)
	assert.False(t, s.Realtime)
	assert.False(t, s.BeginParsed.IsZero())
	assert.False(t, s.EndParsed.IsZero())
}

func TestParseBeginEndEmptyEnd(t *testing.T) {
	s := &Sample{
		Name: "test",
		End:  "",
	}

	ParseBeginEnd(s)

	// Empty end means realtime
	assert.True(t, s.Realtime)
	assert.True(t, s.EndParsed.IsZero())
}

func TestParseBeginEndBeginOverridesRealtime(t *testing.T) {
	s := &Sample{
		Name:  "test",
		Begin: "-60s",
		End:   "",
	}

	ParseBeginEnd(s)

	// Begin set without endIntervals: sets Realtime to false via parsing
	assert.False(t, s.Realtime)
	assert.False(t, s.BeginParsed.IsZero())
}

func TestSetupFromFile(t *testing.T) {
	SetupFromFile("/tmp/testfile.yml")
	defer CleanupConfigAndEnvironment()

	assert.Equal(t, "..", os.Getenv("GOGEN_HOME"))
	assert.Equal(t, "1", os.Getenv("GOGEN_ALWAYS_REFRESH"))
	assert.Equal(t, "/tmp/testfile.yml", os.Getenv("GOGEN_FULLCONFIG"))
}

func TestSetupSystemTokensSplunkHEC(t *testing.T) {
	ResetConfig()

	configStr := `
global:
  output:
    outputter: stdout
    outputTemplate: splunkhec
samples:
  - name: hectokensample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("hectokensample")
	assert.NotNil(t, s)

	// Should have a _time token added by SetupSystemTokens
	foundTime := false
	for _, tk := range s.Tokens {
		if tk.Name == "_time" {
			foundTime = true
			assert.Equal(t, "epochtimestamp", tk.Type)
		}
	}
	assert.True(t, foundTime, "splunkhec should add _time token")
}

func TestSetupSystemTokensElasticsearch(t *testing.T) {
	ResetConfig()

	configStr := `
global:
  output:
    outputter: stdout
    outputTemplate: elasticsearch
samples:
  - name: estokensample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("estokensample")
	assert.NotNil(t, s)

	foundTimestamp := false
	for _, tk := range s.Tokens {
		if tk.Name == "@timestamp" {
			foundTimestamp = true
			assert.Equal(t, "gotimestamp", tk.Type)
		}
	}
	assert.True(t, foundTimestamp, "elasticsearch should add @timestamp token")
}

func TestSetupSystemTokensRFC3164(t *testing.T) {
	ResetConfig()

	configStr := `
global:
  output:
    outputter: stdout
    outputTemplate: rfc3164
samples:
  - name: rfc3164sample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: syslog event
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("rfc3164sample")
	assert.NotNil(t, s)

	foundTime := false
	foundPriority := false
	foundHost := false
	foundTag := false
	foundPid := false
	for _, tk := range s.Tokens {
		if tk.Name == "_time" {
			foundTime = true
			assert.Equal(t, "gotimestamp", tk.Type)
		}
	}
	assert.True(t, foundTime, "rfc3164 should add _time token")

	// Check that syslog fields were added to lines
	if len(s.Lines) > 0 {
		if _, ok := s.Lines[0]["priority"]; ok {
			foundPriority = true
		}
		if _, ok := s.Lines[0]["host"]; ok {
			foundHost = true
		}
		if _, ok := s.Lines[0]["tag"]; ok {
			foundTag = true
		}
		if _, ok := s.Lines[0]["pid"]; ok {
			foundPid = true
		}
	}
	assert.True(t, foundPriority, "rfc3164 should add priority field")
	assert.True(t, foundHost, "rfc3164 should add host field")
	assert.True(t, foundTag, "rfc3164 should add tag field")
	assert.True(t, foundPid, "rfc3164 should add pid field")
}

func TestSetupSystemTokensRFC5424(t *testing.T) {
	ResetConfig()

	configStr := `
global:
  output:
    outputter: stdout
    outputTemplate: rfc5424
samples:
  - name: rfc5424sample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: syslog5424 event
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("rfc5424sample")
	assert.NotNil(t, s)

	foundTime := false
	for _, tk := range s.Tokens {
		if tk.Name == "_time" {
			foundTime = true
			assert.Equal(t, "gotimestamp", tk.Type)
		}
	}
	assert.True(t, foundTime, "rfc5424 should add _time token")

	if len(s.Lines) > 0 {
		_, hasAppName := s.Lines[0]["appName"]
		assert.True(t, hasAppName, "rfc5424 should add appName field")
	}
}

func TestBuildConfigDefaults(t *testing.T) {
	ResetConfig()

	configStr := `
samples:
  - name: defaultsample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()

	// Check defaults were applied
	assert.Equal(t, 1, c.Global.GeneratorWorkers)
	assert.Equal(t, 1, c.Global.OutputWorkers)
	assert.Equal(t, "stdout", c.Global.Output.Outputter)
	assert.Equal(t, "raw", c.Global.Output.OutputTemplate)
	assert.Equal(t, 5, c.Global.Output.BackupFiles)
	assert.NotZero(t, c.Global.Output.MaxBytes)
	assert.NotZero(t, c.Global.Output.BufferBytes)
	assert.NotZero(t, c.Global.Output.Timeout)
}

func TestValidateDisabledNoLines(t *testing.T) {
	ResetConfig()

	configStr := `
samples:
  - name: nolines
    interval: 1
    count: 1
    endIntervals: 1
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	// Sample with no lines should be disabled and cleaned away
	s := c.FindSampleByName("nolines")
	assert.Nil(t, s, "sample with no lines should be removed")
}

func TestConvertUTC(t *testing.T) {
	ResetConfig()

	configStr := `
global:
  utc: true
samples:
  - name: utctest
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	_ = NewConfig()

	now := time.Now()
	utcTime := convertUTC(now)
	assert.Equal(t, now.UTC(), utcTime)
}

func TestSampleNow(t *testing.T) {
	s := &Sample{
		Realtime: true,
	}
	beforeCall := time.Now()
	result := s.Now()
	afterCall := time.Now()
	assert.True(t, !result.Before(beforeCall) && !result.After(afterCall),
		"Realtime Now() should return current time")

	fixedTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Realtime = false
	s.Current = fixedTime
	result = s.Now()
	assert.Equal(t, fixedTime, result)
}

func TestReadSamplesDir(t *testing.T) {
	ResetConfig()

	// Use the existing test samples directory
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join("..", "tests", "tokens"))
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	// Should have loaded samples from the tokens test directory
	assert.NotEmpty(t, c.Samples, "should load samples from samples dir")
}

func TestParseFileConfigJSON(t *testing.T) {
	ResetConfig()

	// Create a JSON config file with the Config struct format
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "test.json")
	jsonContent := `{"samples": [{"name": "jsonsample", "interval": 1, "count": 1, "endIntervals": 1, "lines": [{"_raw": "json test"}]}]}`
	os.WriteFile(jsonFile, []byte(jsonContent), 0644)

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", jsonFile)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	assert.NotEmpty(t, c.Samples, "should load samples from JSON config")
}

func TestNegativeCacheIntervals(t *testing.T) {
	ResetConfig()

	configStr := `
global:
  cacheIntervals: -5
samples:
  - name: cachesample
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	assert.Equal(t, 0, c.Global.CacheIntervals, "negative cacheIntervals should be clamped to 0")
}

func TestReadSamplesDirSampleFile(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()

	// Create a .sample file
	err := os.WriteFile(filepath.Join(dir, "test.sample"), []byte("line one\nline two\nline three\n"), 0644)
	assert.NoError(t, err)

	c := &Config{cc: ConfigConfig{}}
	c.readSamplesDir(dir)

	// Should have loaded one sample with 3 lines
	found := false
	for _, s := range c.Samples {
		if s.Name == "test.sample" {
			found = true
			assert.True(t, s.Disabled, ".sample files should be disabled by default")
			assert.Equal(t, 3, len(s.Lines))
			assert.Equal(t, "line one", s.Lines[0]["_raw"])
			assert.Equal(t, "line two", s.Lines[1]["_raw"])
			assert.Equal(t, "line three", s.Lines[2]["_raw"])
		}
	}
	assert.True(t, found, "should find test.sample")
}

func TestReadSamplesDirCSVFile(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()

	// Create a .csv file with header
	csvContent := "name,city,state\nalice,NYC,NY\nbob,LA,CA\n"
	err := os.WriteFile(filepath.Join(dir, "test.csv"), []byte(csvContent), 0644)
	assert.NoError(t, err)

	c := &Config{cc: ConfigConfig{}}
	c.readSamplesDir(dir)

	found := false
	for _, s := range c.Samples {
		if s.Name == "test.csv" {
			found = true
			assert.True(t, s.Disabled, ".csv files should be disabled by default")
			assert.Equal(t, 2, len(s.Lines))
			assert.Equal(t, "alice", s.Lines[0]["name"])
			assert.Equal(t, "NYC", s.Lines[0]["city"])
			assert.Equal(t, "NY", s.Lines[0]["state"])
			assert.Equal(t, "bob", s.Lines[1]["name"])
		}
	}
	assert.True(t, found, "should find test.csv")
}

func TestReadSamplesDirYAMLFile(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()

	yamlContent := `name: yamlsample
interval: 1
count: 1
lines:
  - _raw: yaml test line
`
	err := os.WriteFile(filepath.Join(dir, "yamltest.yml"), []byte(yamlContent), 0644)
	assert.NoError(t, err)

	c := &Config{cc: ConfigConfig{}}
	c.readSamplesDir(dir)

	found := false
	for _, s := range c.Samples {
		if s.Name == "yamlsample" {
			found = true
			assert.True(t, s.realSample)
		}
	}
	assert.True(t, found, "should find yamlsample")
}

func TestReadSamplesDirEmptyDir(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()

	c := &Config{cc: ConfigConfig{}}
	c.readSamplesDir(dir)

	// Should not crash and should have no samples
	assert.Empty(t, c.Samples)
}

func TestReadGeneratorFallbackPath(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()
	genDir := filepath.Join(dir, "generators")
	os.MkdirAll(genDir, 0755)

	// Create generator script in the fallback directory
	scriptContent := `-- test generator\nsetToken("test", "value")\n`
	err := os.WriteFile(filepath.Join(genDir, "testgen.lua"), []byte(scriptContent), 0644)
	assert.NoError(t, err)

	c := &Config{cc: ConfigConfig{ConfigDir: dir}}
	g := &GeneratorConfig{Name: "testgen", FileName: "testgen.lua"}

	err = c.readGenerator(dir, g)
	assert.NoError(t, err)
	assert.Contains(t, g.Script, "test generator")
}

func TestReadGeneratorNotFound(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()
	c := &Config{cc: ConfigConfig{ConfigDir: dir}}
	g := &GeneratorConfig{Name: "missing", FileName: "nonexistent.lua"}

	err := c.readGenerator(dir, g)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot find generator file")
}

func TestValidateTokenRandomString(t *testing.T) {
	ResetConfig()
	configStr := `
samples:
  - name: randstring
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: rs
        format: template
        token: $rs$
        type: random
        replacement: string
        length: 10
    lines:
      - _raw: $rs$
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("randstring")
	assert.NotNil(t, s, "sample with valid random string token should not be disabled")
}

func TestValidateTokenRandomStringZeroLength(t *testing.T) {
	ResetConfig()
	configStr := `
samples:
  - name: randstringbad
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: rs
        format: template
        token: $rs$
        type: random
        replacement: string
        length: 0
    lines:
      - _raw: $rs$
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("randstringbad")
	assert.Nil(t, s, "sample with zero-length random string should be disabled")
}

func TestValidateTokenReplacementTypes(t *testing.T) {
	tests := []struct {
		name        string
		replacement string
		extra       string
		valid       bool
	}{
		{"hex", "hex", "length: 5", true},
		{"guid", "guid", "", true},
		{"ipv4", "ipv4", "", true},
		{"ipv6", "ipv6", "", true},
		{"invalid", "invalid_replacement_xyz", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ResetConfig()
			extra := ""
			if tc.extra != "" {
				extra = "\n        " + tc.extra
			}
			configStr := `
samples:
  - name: ` + tc.name + `
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: tk
        format: template
        token: $tk$
        type: random
        replacement: ` + tc.replacement + extra + `
    lines:
      - _raw: $tk$
`
			SetupFromString(configStr)
			defer CleanupConfigAndEnvironment()

			c := NewConfig()
			s := c.FindSampleByName(tc.name)
			if tc.valid {
				assert.NotNil(t, s, "%s should not be disabled", tc.name)
			} else {
				assert.Nil(t, s, "%s should be disabled", tc.name)
			}
		})
	}
}

func TestValidateTokenScript(t *testing.T) {
	ResetConfig()
	configStr := `
samples:
  - name: scripttest
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: sc
        format: template
        token: $sc$
        type: script
        init:
          myvar: "42"
        scriptSrc: |
          return "hello"
    lines:
      - _raw: $sc$
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("scripttest")
	assert.NotNil(t, s, "sample with script token should not be disabled")
	// Check that script token has mutex
	for _, tk := range s.Tokens {
		if tk.Name == "sc" {
			assert.NotNil(t, tk.mutex, "script token should have mutex initialized")
		}
	}
}

func TestValidateNoInterval(t *testing.T) {
	ResetConfig()
	configStr := `
samples:
  - name: nointerval
    count: 1
    lines:
      - _raw: test
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("nointerval")
	assert.NotNil(t, s)
	assert.Equal(t, 1, s.EndIntervals, "no interval should auto-set endIntervals to 1")
}

func TestValidateEmptyName(t *testing.T) {
	ResetConfig()

	c := &Config{
		Global: Global{
			Output: Output{
				Outputter:      "stdout",
				OutputTemplate: "raw",
			},
		},
	}
	s := &Sample{
		realSample: true,
		Name:       "",
		Lines:      []map[string]string{{"_raw": "test"}},
	}
	c.validate(s)
	assert.True(t, s.Disabled, "sample with empty name should be disabled")
}

func TestValidateRaterWithIntValues(t *testing.T) {
	ResetConfig()
	configStr := `
raters:
  - name: testrater
    type: config
    options:
      HourOfDay:
        0: 1
        12: 2
      DayOfWeek:
        0: 1.5
        6: 0.5
samples:
  - name: ratertest
    interval: 1
    count: 1
    endIntervals: 1
    rater: testrater
    lines:
      - _raw: test
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	r := c.FindRater("testrater")
	assert.NotNil(t, r)
	// HourOfDay should be converted to map[int]float64
	hod, ok := r.Options["HourOfDay"].(map[int]float64)
	assert.True(t, ok, "HourOfDay should be map[int]float64")
	assert.Equal(t, 1.0, hod[0])
	assert.Equal(t, 2.0, hod[12])
}

func TestValidateWeightedChoice(t *testing.T) {
	ResetConfig()
	configStr := `
samples:
  - name: weightsource
    disabled: true
    lines:
      - value: alpha
        _weight: "3"
      - value: beta
        _weight: "7"
  - name: weightuser
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: wt
        format: template
        token: $wt$
        type: weightedChoice
        sample: weightsource
        srcField: value
    lines:
      - _raw: $wt$
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("weightuser")
	assert.NotNil(t, s)
	for _, tk := range s.Tokens {
		if tk.Name == "wt" {
			assert.NotEmpty(t, tk.WeightedChoice, "should have weighted choices resolved")
			assert.Equal(t, 2, len(tk.WeightedChoice))
		}
	}
}

func TestValidateTokenSampleResolution(t *testing.T) {
	ResetConfig()
	configStr := `
samples:
  - name: choices
    disabled: true
    lines:
      - _raw: alpha
      - _raw: beta
      - _raw: gamma
  - name: resolver
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: pick
        format: template
        token: $pick$
        type: choice
        sample: choices
    lines:
      - _raw: $pick$
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("resolver")
	assert.NotNil(t, s)
	for _, tk := range s.Tokens {
		if tk.Name == "pick" {
			assert.Equal(t, 3, len(tk.Choice), "should resolve 3 choices from sample")
			assert.Contains(t, tk.Choice, "alpha")
			assert.Contains(t, tk.Choice, "beta")
			assert.Contains(t, tk.Choice, "gamma")
		}
	}
}

func TestValidateExportMode(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "export.yml")
	os.WriteFile(configFile, []byte(`
samples:
  - name: exportsample
    interval: 1
    count: 1
    lines:
      - _raw: test
`), 0644)

	cc := ConfigConfig{
		FullConfig: configFile,
		Export:     true,
	}
	c := BuildConfig(cc)

	// In export mode, defaults should NOT be set
	assert.Equal(t, 0, c.Global.GeneratorWorkers, "export mode should not set defaults")
	assert.Equal(t, "", c.Global.Output.Outputter, "export mode should not set output defaults")
}

func TestMergeMixConfig(t *testing.T) {
	c := &Config{}
	nc := &Config{
		Samples: []*Sample{
			{Name: "mixsample", Count: 5, Interval: 2},
		},
	}
	m := &Mix{
		Count:    10,
		Interval: 3,
		Begin:    "-60s",
		End:      "now",
	}
	c.mergeMixConfig(nc, m)

	assert.Equal(t, 1, len(c.Samples))
	assert.Equal(t, "mixsample", c.Samples[0].Name)
	assert.Equal(t, 10, c.Samples[0].Count)
	assert.Equal(t, 3, c.Samples[0].Interval)
}

func TestParseFileConfigYAMLError(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.yml")
	// Invalid YAML: tabs mixed with spaces in wrong places
	os.WriteFile(badFile, []byte("{\n  bad yaml content: [unclosed\n"), 0644)

	c := &Config{cc: ConfigConfig{}}
	s := &Sample{}
	err := c.parseFileConfig(s, badFile)
	// parseFileConfig logs errors but returns nil
	assert.NoError(t, err)
}

func TestParseFileConfigJSONError(t *testing.T) {
	ResetConfig()

	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.json")
	os.WriteFile(badFile, []byte("{invalid json"), 0644)

	c := &Config{cc: ConfigConfig{}}
	s := &Sample{}
	err := c.parseFileConfig(s, badFile)
	assert.NoError(t, err)
}

func TestParseFileConfigNotExists(t *testing.T) {
	ResetConfig()

	c := &Config{cc: ConfigConfig{}}
	s := &Sample{}
	err := c.parseFileConfig(s, "/nonexistent/path/file.yml")
	assert.Error(t, err)
}

func TestParseWebConfigSuccess(t *testing.T) {
	ResetConfig()

	yamlContent := `
name: websample
interval: 1
count: 1
lines:
  - _raw: web test
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(yamlContent))
	}))
	defer ts.Close()

	c := &Config{cc: ConfigConfig{}}
	s := &Sample{}
	err := c.parseWebConfig(s, ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, "websample", s.Name)
}

func TestParseWebConfigJSONFallback(t *testing.T) {
	ResetConfig()

	jsonContent := `{"name": "jsonsample", "interval": 1}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(jsonContent))
	}))
	defer ts.Close()

	c := &Config{cc: ConfigConfig{}}
	s := &Sample{}
	err := c.parseWebConfig(s, ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, "jsonsample", s.Name)
}

func TestParseWebConfigBadContent(t *testing.T) {
	ResetConfig()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("<<<not yaml or json>>>"))
	}))
	defer ts.Close()

	c := &Config{cc: ConfigConfig{}}
	s := &Sample{}
	err := c.parseWebConfig(s, ts.URL)
	// JSON fallback parse returns an error for non-JSON content
	assert.Error(t, err)
	assert.Equal(t, "", s.Name, "garbage content should not set sample name")
}

func TestMergeMixConfigDuplicate(t *testing.T) {
	c := &Config{
		Samples: []*Sample{
			{Name: "existing"},
		},
	}
	nc := &Config{
		Samples: []*Sample{
			{Name: "existing", Count: 5},
		},
	}
	m := &Mix{}
	c.mergeMixConfig(nc, m)

	// Should not add duplicate
	assert.Equal(t, 1, len(c.Samples))
}

func TestGetAPIURLDefault(t *testing.T) {
	os.Unsetenv("GOGEN_APIURL")
	url := getAPIURL()
	assert.Equal(t, "https://api.gogen.io", url)
}

func TestGetAPIURLCustom(t *testing.T) {
	os.Setenv("GOGEN_APIURL", "http://localhost:4000")
	defer os.Unsetenv("GOGEN_APIURL")
	url := getAPIURL()
	assert.Equal(t, "http://localhost:4000", url)
}

func TestValidateFromSample(t *testing.T) {
	ResetConfig()

	configStr := `
samples:
  - name: sourcesample
    disabled: true
    lines:
      - _raw: source line 1
      - _raw: source line 2
  - name: copiedsample
    fromSample: sourcesample
    interval: 1
    count: 1
    endIntervals: 1
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("copiedsample")
	assert.NotNil(t, s)
	assert.Len(t, s.Lines, 2, "copiedsample should have lines from sourcesample")
}

func TestNewGeneratorStateNumericInit(t *testing.T) {
	s := &Sample{
		CustomGenerator: &GeneratorConfig{
			Init: map[string]string{
				"count": "42",
				"rate":  "3.14",
				"label": "hello",
			},
		},
		Lines: []map[string]string{
			{"_raw": "line1", "host": "h1"},
			{"_raw": "line2", "host": "h2"},
		},
	}

	gs := NewGeneratorState(s)
	assert.NotNil(t, gs.LuaState)
	assert.NotNil(t, gs.LuaLines)

	// Numeric values should be stored as LNumber
	countVal := gs.LuaState.RawGetString("count")
	assert.NotNil(t, countVal)

	// String values should be stored as LString
	labelVal := gs.LuaState.RawGetString("label")
	assert.NotNil(t, labelVal)

	// Lines table should have entries
	assert.Equal(t, 2, gs.LuaLines.Len())
}

func TestNewGeneratorStateEmptyInit(t *testing.T) {
	s := &Sample{
		CustomGenerator: &GeneratorConfig{
			Init: map[string]string{},
		},
		Lines: []map[string]string{},
	}

	gs := NewGeneratorState(s)
	assert.NotNil(t, gs.LuaState)
	assert.NotNil(t, gs.LuaLines)
	assert.Equal(t, 0, gs.LuaLines.Len())
}

func TestBuildConfigExportMode(t *testing.T) {
	ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := filepath.Join("..", "tests", "tokens")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	cc := ConfigConfig{
		SamplesDir: home,
		Home:       "..",
		Export:     true,
	}
	c := BuildConfig(cc)
	assert.NotNil(t, c)
	// In export mode, samples should have lines populated inline
	for _, s := range c.Samples {
		if s.Name == "tokens" {
			assert.Greater(t, len(s.Lines), 0)
		}
	}
}

func TestBuildConfigWithGlobalFile(t *testing.T) {
	ResetConfig()
	globalFile := filepath.Join("..", "tests", "rater", "defaultrater.yml")
	cc := ConfigConfig{
		FullConfig: globalFile,
		Home:       "..",
	}
	c := BuildConfig(cc)
	assert.NotNil(t, c)
}

func TestValidateInvalidEarliestTime(t *testing.T) {
	ResetConfig()
	configStr := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: badtime
    description: "Bad earliest time"
    earliest: "not_a_valid_time_string!!!"
    latest: now
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("badtime")
	// With invalid earliest, EarliestParsed should default to 0
	assert.Equal(t, time.Duration(0), s.EarliestParsed)
}

func TestValidateInvalidLatestTime(t *testing.T) {
	ResetConfig()
	configStr := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: badlatest
    description: "Bad latest time"
    earliest: now
    latest: "not_a_valid_time_string!!!"
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("badlatest")
	// With invalid latest, LatestParsed should default to 0
	assert.Equal(t, time.Duration(0), s.LatestParsed)
}

func TestNewConfigNoGogenHome(t *testing.T) {
	ResetConfig()
	os.Unsetenv("GOGEN_HOME")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Unsetenv("GOGEN_FULLCONFIG")
	os.Unsetenv("GOGEN_CONFIG_DIR")
	os.Unsetenv("GOGEN_SAMPLES_DIR")

	c := NewConfig()
	assert.NotNil(t, c)
	// When GOGEN_HOME is not set, it should default to "."
	assert.Equal(t, ".", os.Getenv("GOGEN_HOME"))
}

func TestValidateNoLinesDisablesSample(t *testing.T) {
	ResetConfig()
	configStr := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: nolines
    description: "Sample with no lines"
    interval: 1
    count: 1
    endIntervals: 1
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	// After Clean(), disabled samples are removed
	// So FindSampleByName should return an empty sample (not in the list)
	found := false
	for _, s := range c.Samples {
		if s.Name == "nolines" {
			found = true
		}
	}
	assert.False(t, found, "disabled sample with no lines should be removed by Clean()")
}

func TestValidateRatedTokenDefaultRater(t *testing.T) {
	ResetConfig()
	configStr := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: rated_test
    description: "Rated token default rater"
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: myrated
        format: template
        type: rated
        replacement: int
        lower: 0
        upper: 100
    lines:
      - _raw: value=$myrated$
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("rated_test")
	// Rated token with no raterString should default to "default"
	for _, tok := range s.Tokens {
		if tok.Name == "myrated" {
			assert.Equal(t, "default", tok.RaterString)
		}
	}
}

func TestValidateLuaGenerator(t *testing.T) {
	ResetConfig()
	configStr := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: luagen
    description: "Lua generator sample"
    generator: mygen
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
generators:
  - name: mygen
    script: |
      lines = getLines()
      return send(lines)
`
	SetupFromString(configStr)
	defer CleanupConfigAndEnvironment()

	c := NewConfig()
	s := c.FindSampleByName("luagen")
	assert.NotNil(t, s)
	assert.Equal(t, "mygen", s.Generator)
	assert.NotNil(t, s.CustomGenerator)
}
