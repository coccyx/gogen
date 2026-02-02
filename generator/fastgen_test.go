package generator

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// UNIT TESTS - Verify FastGen correctness
// =============================================================================

func TestFastGenSimple(t *testing.T) {
	// Create a simple sample
	s := &config.Sample{
		Name:      "test-fastgen",
		Generator: "sample",
		Lines: []map[string]string{
			{"_raw": "Simple test event", "host": "testhost", "source": "testsource"},
		},
		Tokens: []config.Token{},
	}

	// Initialize FastPath
	ok := s.InitFastPath("json")
	assert.True(t, ok, "FastPath should initialize")
	assert.True(t, s.UseFastPath)
	assert.Len(t, s.FastTemplates, 1)

	// Create GenQueueItem
	oq := make(chan *config.OutQueueItem, 1)
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

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

	// Generate using fast path
	fg := fastgen{}
	err := fg.Gen(gqi)
	assert.NoError(t, err)

	// Check output
	oqi := <-oq
	assert.NotNil(t, oqi.FastOutput)
	assert.Equal(t, 1, oqi.EventCount)
	assert.Contains(t, string(oqi.FastOutput), "Simple test event")
	assert.Contains(t, string(oqi.FastOutput), "testhost")
}

func TestFastGenWithTokens(t *testing.T) {
	// Create a sample with tokens
	s := &config.Sample{
		Name:      "test-fastgen-tokens",
		Generator: "sample",
		Lines: []map[string]string{
			{"_raw": "User $user$ logged in from $ip$", "host": "testhost"},
		},
		Tokens: []config.Token{
			{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
			{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "10.0.0.1", Field: "_raw"},
		},
	}

	// Initialize FastPath
	ok := s.InitFastPath("json")
	assert.True(t, ok, "FastPath should initialize")

	// Create GenQueueItem
	oq := make(chan *config.OutQueueItem, 1)
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

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

	// Generate using fast path
	fg := fastgen{}
	err := fg.Gen(gqi)
	assert.NoError(t, err)

	// Check output
	oqi := <-oq
	assert.NotNil(t, oqi.FastOutput)
	output := string(oqi.FastOutput)
	assert.Contains(t, output, "admin")
	assert.Contains(t, output, "10.0.0.1")
	assert.NotContains(t, output, "$user$")
	assert.NotContains(t, output, "$ip$")
}

func TestFastGenBatch(t *testing.T) {
	s := &config.Sample{
		Name:      "test-fastgen-batch",
		Generator: "sample",
		Lines: []map[string]string{
			{"_raw": "Event 1", "host": "host1"},
			{"_raw": "Event 2", "host": "host2"},
		},
		Tokens: []config.Token{},
	}

	ok := s.InitFastPath("raw")
	assert.True(t, ok)

	oq := make(chan *config.OutQueueItem, 1)
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

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

	fg := fastgen{}
	err := fg.Gen(gqi)
	assert.NoError(t, err)

	oqi := <-oq
	assert.NotNil(t, oqi.FastOutput)
	assert.Equal(t, 10, oqi.EventCount)

	// Count newlines to verify we have multiple events
	newlines := bytes.Count(oqi.FastOutput, []byte("\n"))
	assert.Equal(t, 9, newlines) // 10 events = 9 separating newlines
}

// =============================================================================
// BENCHMARKS - Compare FastGen vs Traditional generator
// =============================================================================

func setupBenchmarkSample(tokenCount int) *config.Sample {
	s := &config.Sample{
		Name:      "bench-sample",
		Generator: "sample",
		Lines: []map[string]string{
			{
				"_raw":       "User $user$ from $ip$ did $action$ at $time$ status=$status$",
				"host":       "myhost.example.com",
				"source":     "/var/log/app.log",
				"sourcetype": "app:log",
				"index":      "main",
			},
		},
		Tokens: []config.Token{
			{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
			{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "192.168.1.100", Field: "_raw"},
			{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "login", Field: "_raw"},
			{Name: "time", Format: "template", Token: "$time$", Type: "static", Replacement: "2024-01-01T00:00:00Z", Field: "_raw"},
			{Name: "status", Format: "template", Token: "$status$", Type: "static", Replacement: "success", Field: "_raw"},
		},
	}
	if tokenCount < len(s.Tokens) {
		s.Tokens = s.Tokens[:tokenCount]
	}
	return s
}

// drainChannel reads from a channel until closed
func drainChannel(oq chan *config.OutQueueItem, done chan struct{}) {
	for range oq {
	}
	close(done)
}

// BenchmarkTraditionalGen100 benchmarks traditional generation of 100 events
func BenchmarkTraditionalGen100(b *testing.B) {
	s := setupBenchmarkSample(5)
	s.UseFastPath = false // Force traditional path

	oq := make(chan *config.OutQueueItem, 100)
	done := make(chan struct{})
	go drainChannel(oq, done)

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	gen := sample{}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		gqi := &config.GenQueueItem{
			Count:    100,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}
		gen.Gen(gqi)
	}

	close(oq)
	<-done
}

// BenchmarkFastGen100 benchmarks fast path generation of 100 events
func BenchmarkFastGen100(b *testing.B) {
	s := setupBenchmarkSample(5)
	s.InitFastPath("json")

	oq := make(chan *config.OutQueueItem, 100)
	done := make(chan struct{})
	go drainChannel(oq, done)

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	gen := fastgen{}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		gqi := &config.GenQueueItem{
			Count:    100,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}
		gen.Gen(gqi)
	}

	close(oq)
	<-done
}

// BenchmarkTraditionalGen1000 benchmarks traditional generation of 1000 events
func BenchmarkTraditionalGen1000(b *testing.B) {
	s := setupBenchmarkSample(5)
	s.UseFastPath = false

	oq := make(chan *config.OutQueueItem, 100)
	done := make(chan struct{})
	go drainChannel(oq, done)

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	gen := sample{}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		gqi := &config.GenQueueItem{
			Count:    1000,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}
		gen.Gen(gqi)
	}

	close(oq)
	<-done
}

// BenchmarkFastGen1000 benchmarks fast path generation of 1000 events
func BenchmarkFastGen1000(b *testing.B) {
	s := setupBenchmarkSample(5)
	s.InitFastPath("json")

	oq := make(chan *config.OutQueueItem, 100)
	done := make(chan struct{})
	go drainChannel(oq, done)

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	gen := fastgen{}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		gqi := &config.GenQueueItem{
			Count:    1000,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}
		gen.Gen(gqi)
	}

	close(oq)
	<-done
}

// BenchmarkEndToEndTraditional benchmarks the full pipeline (gen + output formatting)
func BenchmarkEndToEndTraditional(b *testing.B) {
	s := setupBenchmarkSample(5)
	s.UseFastPath = false

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	gen := sample{}

	var totalBytes int64

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		oq := make(chan *config.OutQueueItem, 1)

		gqi := &config.GenQueueItem{
			Count:    100,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}

		go gen.Gen(gqi)
		oqi := <-oq

		// Simulate output formatting (write to discard)
		var buf bytes.Buffer
		for _, event := range oqi.Events {
			// Simulate JSON output
			buf.WriteString(`{"_raw":"`)
			buf.WriteString(event["_raw"])
			buf.WriteString(`","host":"`)
			buf.WriteString(event["host"])
			buf.WriteString(`"}`)
			buf.WriteByte('\n')
		}
		totalBytes += int64(buf.Len())
	}
	b.SetBytes(totalBytes / int64(b.N))
}

// BenchmarkEndToEndFast benchmarks the full pipeline with fast path
func BenchmarkEndToEndFast(b *testing.B) {
	s := setupBenchmarkSample(5)
	s.InitFastPath("json")

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	gen := fastgen{}

	var totalBytes int64

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		oq := make(chan *config.OutQueueItem, 1)

		gqi := &config.GenQueueItem{
			Count:    100,
			Earliest: now,
			Latest:   now,
			Now:      now,
			S:        s,
			OQ:       oq,
			Rand:     randgen,
			Cache:    &config.CacheItem{UseCache: false, SetCache: false},
		}

		go gen.Gen(gqi)
		oqi := <-oq

		// Fast path already has formatted output - just write it
		_, _ = io.Discard.Write(oqi.FastOutput)
		totalBytes += int64(len(oqi.FastOutput))
	}
	b.SetBytes(totalBytes / int64(b.N))
}
