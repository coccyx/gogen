package generator

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/tests"
	"github.com/stretchr/testify/assert"
)

// setupGenPerfTest sets up the environment for generator performance tests
func setupGenPerfTest(configFile string) (*config.Sample, *rand.Rand, time.Time) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := filepath.Join("..", "tests", "tokens")
	os.Setenv("GOGEN_SAMPLES_DIR", home)

	source := rand.NewSource(0)
	randgen := rand.New(source)
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)

	s := tests.FindSampleInFile(home, configFile)
	return s, randgen, now
}

// =============================================================================
// UNIT TESTS - Verify correctness
// =============================================================================

func TestCopyEventCorrectness(t *testing.T) {
	src := map[string]string{
		"_raw":       "test event",
		"host":       "myhost",
		"source":     "/var/log/test",
		"sourcetype": "test:log",
		"field1":     "value1",
		"field2":     "value2",
	}

	dst := copyevent(src)

	// Verify all fields copied
	assert.Equal(t, len(src), len(dst))
	for k, v := range src {
		assert.Equal(t, v, dst[k])
	}

	// Verify it's a true copy (modifying dst doesn't affect src)
	dst["_raw"] = "modified"
	assert.NotEqual(t, src["_raw"], dst["_raw"])
}

func TestGenMultiPassCorrectness(t *testing.T) {
	s, randgen, now := setupGenPerfTest("tokens-multi")
	if s == nil {
		t.Skip("tokens-multi sample not found")
	}

	oq := make(chan *config.OutQueueItem, 1)
	gqi := &config.GenQueueItem{
		Count:    1,
		Earliest: now,
		Latest:   now,
		Now:      now,
		S:        s,
		OQ:       oq,
		Rand:     randgen,
		Cache:    &config.CacheItem{UseCache: false, SetCache: false},
	}

	err := genMultiPass(gqi)
	assert.NoError(t, err)

	oqi := <-oq
	assert.Len(t, oqi.Events, 1)
	assert.Contains(t, oqi.Events[0]["_raw"], "ONE")
	assert.Contains(t, oqi.Events[0]["_raw"], "FIVE")
}

func TestGenSinglePassCorrectness(t *testing.T) {
	s, randgen, now := setupGenPerfTest("token-static")
	if s == nil {
		t.Skip("token-static sample not found")
	}

	oq := make(chan *config.OutQueueItem, 1)
	gqi := &config.GenQueueItem{
		Count:    1,
		Earliest: now,
		Latest:   now,
		Now:      now,
		S:        s,
		OQ:       oq,
		Rand:     randgen,
		Cache:    &config.CacheItem{UseCache: false, SetCache: false},
	}

	err := genSinglePass(gqi)
	assert.NoError(t, err)

	oqi := <-oq
	assert.Len(t, oqi.Events, 1)
	assert.Equal(t, "foo", oqi.Events[0]["_raw"])
}

// =============================================================================
// BENCHMARKS - Measure performance
// =============================================================================

// BenchmarkCopyEvent benchmarks the map copy function
func BenchmarkCopyEvent(b *testing.B) {
	src := map[string]string{
		"_raw":       "test event with some content that is somewhat realistic in length",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/test.log",
		"sourcetype": "application:log",
		"field1":     "value1",
		"field2":     "value2",
		"field3":     "value3",
		"field4":     "value4",
		"field5":     "value5",
		"index":      "main",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_ = copyevent(src)
	}
}

// BenchmarkCopyEventSmall benchmarks copying a small event (3 fields)
func BenchmarkCopyEventSmall(b *testing.B) {
	src := map[string]string{
		"_raw":   "small event",
		"host":   "host",
		"source": "source",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_ = copyevent(src)
	}
}

// BenchmarkGenMultiPass benchmarks the multi-pass generation with token replacement
func BenchmarkGenMultiPass(b *testing.B) {
	s, randgen, now := setupGenPerfTest("tokens-multi")
	if s == nil {
		b.Skip("tokens-multi sample not found")
	}

	oq := make(chan *config.OutQueueItem, 100)

	// Drain the channel in background
	done := make(chan struct{})
	go func() {
		for range oq {
		}
		close(done)
	}()

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		gqi := &config.GenQueueItem{
			Count:    10,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}
		_ = genMultiPass(gqi)
	}

	close(oq)
	<-done
}

// BenchmarkGenSinglePass benchmarks the single-pass generation (optimized path)
func BenchmarkGenSinglePass(b *testing.B) {
	s, randgen, now := setupGenPerfTest("token-static")
	if s == nil {
		b.Skip("token-static sample not found")
	}

	oq := make(chan *config.OutQueueItem, 100)

	// Drain the channel in background
	done := make(chan struct{})
	go func() {
		for range oq {
		}
		close(done)
	}()

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		gqi := &config.GenQueueItem{
			Count:    10,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}
		_ = genSinglePass(gqi)
	}

	close(oq)
	<-done
}

// BenchmarkReplaceTokens benchmarks the replaceTokens function directly
func BenchmarkReplaceTokens(b *testing.B) {
	s, randgen, now := setupGenPerfTest("tokens-multi")
	if s == nil {
		b.Skip("tokens-multi sample not found")
	}

	// Pre-create an event to replace tokens in
	originalEvent := s.Lines[0]

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		event := copyevent(originalEvent)
		gqi := &config.GenQueueItem{
			Count:    1,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			Rand:     randgen,
		}
		replaceTokens(gqi, &event, nil, s.Tokens)
	}
}
