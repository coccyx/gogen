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

func TestSharePullShortName(t *testing.T) {
	// Test Pull with a short name (no "/" in gogen string)
	originalGet := Get
	defer func() { Get = originalGet }()

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:       "shortname",
			Name:        "shortname",
			Description: "Short name test",
			Owner:       "testuser",
			Version:     1,
			Config:      "sample: test\nname: shortname",
		}, nil
	}

	os.Setenv("GOGEN_HOME", "..")
	dir := t.TempDir()

	Pull("shortname", dir, false)
	_, err := os.Stat(filepath.Join(dir, "shortname.yml"))
	assert.NoError(t, err, "Couldn't find file shortname.yml")
}

func TestSharePullFileCached(t *testing.T) {
	// Test PullFile when cache exists and version matches → uses cached content
	originalGet := Get
	defer func() { Get = originalGet }()

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:   "testuser/cached",
			Name:    "cached",
			Owner:   "testuser",
			Version: 5,
			Config:  "should not be used",
		}, nil
	}

	tmpdir := t.TempDir()
	os.Setenv("GOGEN_TMPDIR", tmpdir)
	defer os.Unsetenv("GOGEN_TMPDIR")

	// Pre-create version cache with matching version
	versionCacheFile := filepath.Join(tmpdir, ".versioncache_testuser%2Fcached")
	os.WriteFile(versionCacheFile, []byte("5"), 0644)

	// Pre-create config cache with different content
	cacheFile := filepath.Join(tmpdir, ".configcache_testuser%2Fcached")
	os.WriteFile(cacheFile, []byte("cached config content"), 0644)

	outFile := filepath.Join(tmpdir, "output.yml")
	PullFile("testuser/cached", outFile)

	// Should use cached content, not the API response
	data, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	assert.Equal(t, "cached config content", string(data))
}

func TestSharePullFileVersionMismatch(t *testing.T) {
	// Test PullFile when cache version doesn't match → uses API content and updates cache
	originalGet := Get
	defer func() { Get = originalGet }()

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:   "testuser/mismatch",
			Name:    "mismatch",
			Owner:   "testuser",
			Version: 10,
			Config:  "fresh api content",
		}, nil
	}

	tmpdir := t.TempDir()
	os.Setenv("GOGEN_TMPDIR", tmpdir)
	defer os.Unsetenv("GOGEN_TMPDIR")

	// Pre-create version cache with OLD version
	versionCacheFile := filepath.Join(tmpdir, ".versioncache_testuser%2Fmismatch")
	os.WriteFile(versionCacheFile, []byte("5"), 0644)

	// Pre-create config cache with old content
	cacheFile := filepath.Join(tmpdir, ".configcache_testuser%2Fmismatch")
	os.WriteFile(cacheFile, []byte("old cached content"), 0644)

	outFile := filepath.Join(tmpdir, "output.yml")
	PullFile("testuser/mismatch", outFile)

	// Should use API content since version doesn't match
	data, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	assert.Equal(t, "fresh api content", string(data))

	// Cache files should be updated
	versionData, _ := os.ReadFile(versionCacheFile)
	assert.Equal(t, "10", string(versionData))

	cachedData, _ := os.ReadFile(cacheFile)
	assert.Equal(t, "fresh api content", string(cachedData))
}

func TestSharePullWithDeconstructCSV(t *testing.T) {
	// Test deconstructConfig with CSV fieldChoice tokens
	originalGet := Get
	defer func() { Get = originalGet }()

	configYaml := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: csvtest
    description: CSV deconstruct test
    interval: 1
    count: 1
    endIntervals: 1
    tokens:
      - name: city
        format: template
        type: fieldChoice
        field: _raw
        srcField: city
        sample: markets.csv
        fieldChoice:
          - city: NYC
            state: NY
          - city: LA
            state: CA
    lines:
      - _raw: city=$city$
`

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:   "testuser/csvtest",
			Name:    "csvtest",
			Owner:   "testuser",
			Version: 1,
			Config:  configYaml,
		}, nil
	}

	os.Setenv("GOGEN_HOME", "..")
	dir := t.TempDir()

	Pull("testuser/csvtest", dir, true)
	_, err := os.Stat(filepath.Join(dir, "samples", "markets.csv"))
	assert.NoError(t, err, "Couldn't find samples/markets.csv")
	_, err = os.Stat(filepath.Join(dir, "samples", "csvtest.yml"))
	assert.NoError(t, err, "Couldn't find samples/csvtest.yml")
}

func TestSharePullWithDeconstructGenerator(t *testing.T) {
	// Test deconstructConfig with generator that has a fileName
	originalGet := Get
	defer func() { Get = originalGet }()

	configYaml := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: raw
samples:
  - name: gentest
    description: Generator deconstruct test
    generator: mygen
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
generators:
  - name: mygen
    fileName: /path/to/mygen.lua
    script: |
      lines = getLines()
      return send(lines)
`

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:   "testuser/gentest",
			Name:    "gentest",
			Owner:   "testuser",
			Version: 1,
			Config:  configYaml,
		}, nil
	}

	os.Setenv("GOGEN_HOME", "..")
	dir := t.TempDir()

	Pull("testuser/gentest", dir, true)
	_, err := os.Stat(filepath.Join(dir, "generators", "mygen.lua"))
	assert.NoError(t, err, "Couldn't find generators/mygen.lua")
	_, err = os.Stat(filepath.Join(dir, "generators", "mygen.yml"))
	assert.NoError(t, err, "Couldn't find generators/mygen.yml")
}

func TestSharePullWithDeconstructTemplates(t *testing.T) {
	// Test deconstructConfig with templates
	originalGet := Get
	defer func() { Get = originalGet }()

	configYaml := `
global:
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: mytemplate
samples:
  - name: tmpltest
    description: Template deconstruct test
    interval: 1
    count: 1
    endIntervals: 1
    lines:
      - _raw: test event
templates:
  - name: mytemplate
    header: "HEADER\n"
    row: "{{._raw}}\n"
    footer: "FOOTER\n"
`

	Get = func(q string) (GogenInfo, error) {
		return GogenInfo{
			Gogen:   "testuser/tmpltest",
			Name:    "tmpltest",
			Owner:   "testuser",
			Version: 1,
			Config:  configYaml,
		}, nil
	}

	os.Setenv("GOGEN_HOME", "..")
	dir := t.TempDir()

	Pull("testuser/tmpltest", dir, true)
	_, err := os.Stat(filepath.Join(dir, "templates", "mytemplate.yml"))
	assert.NoError(t, err, "Couldn't find templates/mytemplate.yml")
}
