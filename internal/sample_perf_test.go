package internal

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// setupPerfTest sets up the environment for performance tests
func setupPerfTest(configFile string) (*Config, *rand.Rand, time.Time) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", configFile))

	source := rand.NewSource(0)
	randgen := rand.New(source)
	loc, _ := time.LoadLocation("UTC")
	now := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)

	c := NewConfig()
	return c, randgen, now
}

// =============================================================================
// UNIT TESTS - Verify correctness before optimizing
// =============================================================================

func TestGoRandIPv4Correctness(t *testing.T) {
	c, randgen, now := setupPerfTest("tokens.yml")
	s := c.FindSampleByName("tokens")
	fullevent := make(map[string]string)

	// Token index 9 is random_ipv4
	token := s.Tokens[9]
	replacement, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)

	assert.NoError(t, err)
	assert.Regexp(t, `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, replacement)
}

func TestGoRandIPv6Correctness(t *testing.T) {
	c, randgen, now := setupPerfTest("tokens.yml")
	s := c.FindSampleByName("tokens")
	fullevent := make(map[string]string)

	// Token index 10 is random_ipv6
	token := s.Tokens[10]
	replacement, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)

	assert.NoError(t, err)
	assert.Regexp(t, `^[0-9a-f]+:[0-9a-f]+:[0-9a-f]+:[0-9a-f]+:[0-9a-f]+:[0-9a-f]+:[0-9a-f]+:[0-9a-f]+$`, replacement)
}

func TestMultiTokenReplacementCorrectness(t *testing.T) {
	c, randgen, now := setupPerfTest("tokens-multi.yml")
	s := c.FindSampleByName("tokens-multi")
	fullevent := make(map[string]string)

	event := s.Lines[0]["_raw"]
	originalEvent := event

	// Replace all tokens in sequence
	for _, token := range s.Tokens {
		_, err := token.Replace(&event, -1, now, now, now, randgen, fullevent)
		assert.NoError(t, err)
	}

	// Verify all tokens were replaced
	assert.NotEqual(t, originalEvent, event)
	assert.Contains(t, event, "ONE")
	assert.Contains(t, event, "TWO")
	assert.Contains(t, event, "THREE")
	assert.Contains(t, event, "FOUR")
	assert.Contains(t, event, "FIVE")
	assert.NotContains(t, event, "$token")

	expected := "Event with ONE and TWO and THREE and FOUR and FIVE tokens"
	assert.Equal(t, expected, event)
}

func TestRegexReplacementCorrectness(t *testing.T) {
	c, randgen, now := setupPerfTest("tokens-regex-perf.yml")
	s := c.FindSampleByName("tokens-regex-perf")
	fullevent := make(map[string]string)

	event := s.Lines[0]["_raw"]
	token := s.Tokens[0]

	_, err := token.Replace(&event, -1, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "Line with REPLACED in it", event)
}

func TestLongRandomStringCorrectness(t *testing.T) {
	c, randgen, now := setupPerfTest("tokens-perf.yml")
	s := c.FindSampleByName("tokens-perf")
	fullevent := make(map[string]string)

	// Token index 3 is random_string_long (length 100)
	token := s.Tokens[3]
	replacement, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)

	assert.NoError(t, err)
	assert.Len(t, replacement, 100)
	// Verify it only contains alphanumeric characters
	assert.Regexp(t, `^[a-zA-Z0-9]+$`, replacement)
}

func TestScriptTokenCorrectness(t *testing.T) {
	c, randgen, now := setupPerfTest("tokens-perf.yml")
	s := c.FindSampleByName("tokens-perf")
	fullevent := make(map[string]string)

	// Token index 2 is script_static
	token := s.Tokens[2]
	replacement, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)

	assert.NoError(t, err)
	assert.Equal(t, "scripted", replacement)
}

// =============================================================================
// BENCHMARKS - Measure performance for optimization targets
// =============================================================================

// BenchmarkGoRandIPv4 benchmarks IPv4 generation (string concat with dots)
func BenchmarkGoRandIPv4(b *testing.B) {
	c, randgen, now := setupPerfTest("tokens.yml")
	s := c.FindSampleByName("tokens")
	fullevent := make(map[string]string)
	token := s.Tokens[9] // random_ipv4

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _, _ = token.GenReplacement(-1, now, now, now, randgen, fullevent)
	}
}

// BenchmarkGoRandIPv6 benchmarks IPv6 generation (string concat with colons)
func BenchmarkGoRandIPv6(b *testing.B) {
	c, randgen, now := setupPerfTest("tokens.yml")
	s := c.FindSampleByName("tokens")
	fullevent := make(map[string]string)
	token := s.Tokens[10] // random_ipv6

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _, _ = token.GenReplacement(-1, now, now, now, randgen, fullevent)
	}
}

// BenchmarkGoRandStringLong benchmarks long random string generation (100 chars)
func BenchmarkGoRandStringLong(b *testing.B) {
	c, randgen, now := setupPerfTest("tokens-perf.yml")
	s := c.FindSampleByName("tokens-perf")
	fullevent := make(map[string]string)
	token := s.Tokens[3] // random_string_long

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _, _ = token.GenReplacement(-1, now, now, now, randgen, fullevent)
	}
}

// BenchmarkScriptToken benchmarks Lua script token (VM creation overhead)
func BenchmarkScriptToken(b *testing.B) {
	c, randgen, now := setupPerfTest("tokens-perf.yml")
	s := c.FindSampleByName("tokens-perf")
	fullevent := make(map[string]string)
	token := s.Tokens[2] // script_static

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _, _ = token.GenReplacement(-1, now, now, now, randgen, fullevent)
	}
}

// BenchmarkRegexReplacement benchmarks regex token replacement (regex compilation)
func BenchmarkRegexReplacement(b *testing.B) {
	c, randgen, now := setupPerfTest("tokens-regex-perf.yml")
	s := c.FindSampleByName("tokens-regex-perf")
	fullevent := make(map[string]string)
	token := s.Tokens[0]
	originalEvent := s.Lines[0]["_raw"]

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		event := originalEvent
		_, _ = token.Replace(&event, -1, now, now, now, randgen, fullevent)
	}
}

// BenchmarkMultiTokenReplacement benchmarks replacing 5 tokens in one event
// This is the critical test for string concatenation in Replace()
func BenchmarkMultiTokenReplacement(b *testing.B) {
	c, randgen, now := setupPerfTest("tokens-multi.yml")
	s := c.FindSampleByName("tokens-multi")
	fullevent := make(map[string]string)
	originalEvent := s.Lines[0]["_raw"]

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		event := originalEvent
		for _, token := range s.Tokens {
			_, _ = token.Replace(&event, -1, now, now, now, randgen, fullevent)
		}
	}
}

// BenchmarkSingleTokenReplacement benchmarks replacing 1 token (baseline)
func BenchmarkSingleTokenReplacement(b *testing.B) {
	c, randgen, now := setupPerfTest("token-static.yml")
	s := c.FindSampleByName("token-static")
	fullevent := make(map[string]string)
	token := s.Tokens[0]

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		event := "$static$"
		_, _ = token.Replace(&event, -1, now, now, now, randgen, fullevent)
	}
}

// BenchmarkGetReplacementOffsetsTemplate benchmarks template token offset finding
func BenchmarkGetReplacementOffsetsTemplate(b *testing.B) {
	c, _, _ := setupPerfTest("tokens-multi.yml")
	s := c.FindSampleByName("tokens-multi")
	token := s.Tokens[0]
	event := s.Lines[0]["_raw"]

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _ = token.GetReplacementOffsets(event)
	}
}

// BenchmarkGetReplacementOffsetsRegex benchmarks regex token offset finding
func BenchmarkGetReplacementOffsetsRegex(b *testing.B) {
	c, _, _ := setupPerfTest("tokens-regex-perf.yml")
	s := c.FindSampleByName("tokens-regex-perf")
	token := s.Tokens[0]
	event := s.Lines[0]["_raw"]

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _ = token.GetReplacementOffsets(event)
	}
}
