package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSharePull(t *testing.T) {
	// Mock the Get function to return a GogenInfo with Config
	originalGet := Get
	defer func() { Get = originalGet }()

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:       "testuser/testconfig",
			Name:        "testconfig",
			Description: "Test configuration",
			Notes:       "Test notes",
			Owner:       "testuser",
			Version:     1,
			Config:      "sample: test\nname: testconfig",
		}, nil
	}

	os.Setenv("GOGEN_HOME", "..")
	_ = os.Mkdir("testout", 0777)
	defer os.RemoveAll("testout")

	Pull("testuser/testconfig", "testout", false)
	_, err := os.Stat("testout/testconfig.yml")
	assert.NoError(t, err, "Couldn't find file testconfig.yml")
}

func TestSharePullWithDeconstruct(t *testing.T) {
	// Mock the Get function to return a GogenInfo with Config
	originalGet := Get
	defer func() { Get = originalGet }()

	configYaml := `
global:
  debug: false
  verbose: false
  generatorWorkers: 1
  outputWorkers: 1
  rotInterval: 1
  output:
    outputter: file
    fileName: /tmp/testconfig.log
    maxBytes: 102400
    outputTemplate: json   
samples:
  - name: testconfig
    description: Test configuration
    notes: Test notes
    endIntervals: 100
    interval: 1
    count: 100
    tokens:
      - name: host
        format: template
        type: choice
        field: host
        sample: hosts.sample
        choice:
          - host1
          - host2
      - name: useragent
        format: template
        type: choice
        field: useragent
        choice:
          - "Mozilla/5.0"
          - "Chrome/51.0"
        sample: useragents.sample
      - name: value
        format: template
        type: random
        replacement: float
        precision: 3
        lower: 0
        upper: 10
    lines:
      - _raw: host=$host$ useragent="$useragent$" value=$value$
`

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:       "testuser/testconfig",
			Name:        "testconfig",
			Description: "Test configuration",
			Notes:       "Test notes",
			Owner:       "testuser",
			Version:     1,
			Config:      configYaml,
		}, nil
	}

	os.Setenv("GOGEN_HOME", "..")
	_ = os.Mkdir("testout", 0777)
	defer os.RemoveAll("testout")

	Pull("testuser/testconfig", "testout", true)
	_, err := os.Stat("testout/samples/testconfig.yml")
	assert.NoError(t, err, "Couldn't find file samples/testconfig.yml")
	_, err = os.Stat("testout/samples/hosts.sample")
	assert.NoError(t, err, "Couldn't find file samples/hosts.sample")
	_, err = os.Stat("testout/samples/useragents.sample")
	assert.NoError(t, err, "Couldn't find file samples/useragents.sample")
}

func TestSharePullFile(t *testing.T) {
	// Mock the Get function to return a GogenInfo with Config
	originalGet := Get
	defer func() { Get = originalGet }()

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:       "testuser/testconfig",
			Name:        "testconfig",
			Description: "Test configuration",
			Notes:       "Test notes",
			Owner:       "testuser",
			Version:     1,
			Config:      "sample: test\nname: testconfig",
		}, nil
	}

	os.Setenv("GOGEN_TMPDIR", "..")
	os.Setenv("GOGEN_HOME", "..")
	os.Remove("../.versioncache_testuser%2Ftestconfig")
	os.Remove("../.configcache_testuser%2Ftestconfig")
	defer func() {
		os.Remove(".test.yml")
		os.Remove("../.versioncache_testuser%2Ftestconfig")
		os.Remove("../.configcache_testuser%2Ftestconfig")
	}()

	PullFile("testuser/testconfig", ".test.yml")
	_, err := os.Stat(".test.yml")
	assert.NoError(t, err, "Couldn't find .test.yml")
	_, err = os.Stat(filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".versioncache_testuser%2Ftestconfig"))
	assert.NoError(t, err, "Couldn't find version cache file")
	_, err = os.Stat(filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".configcache_testuser%2Ftestconfig"))
	assert.NoError(t, err, "Couldn't find cache file")
}
