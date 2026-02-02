package internal

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// UNIT TESTS - Verify FastTemplate correctness
// =============================================================================

func TestFastTemplateRawOutput(t *testing.T) {
	line := map[string]string{
		"_raw": "User logged in from 192.168.1.1",
	}
	tokens := []Token{} // No tokens

	ft := BuildFastTemplate(line, tokens, "raw")
	assert.NotNil(t, ft)

	var buf bytes.Buffer
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	err := ft.Execute(&buf, now, now, now, randgen)
	assert.NoError(t, err)
	assert.Equal(t, "User logged in from 192.168.1.1", buf.String())
}

func TestFastTemplateWithTokens(t *testing.T) {
	line := map[string]string{
		"_raw": "User $username$ logged in",
	}
	tokens := []Token{
		{
			Name:        "username",
			Format:      "template",
			Token:       "$username$",
			Type:        "static",
			Replacement: "admin",
			Field:       "_raw",
		},
	}

	ft := BuildFastTemplate(line, tokens, "raw")
	assert.NotNil(t, ft)

	var buf bytes.Buffer
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	err := ft.Execute(&buf, now, now, now, randgen)
	assert.NoError(t, err)
	assert.Equal(t, "User admin logged in", buf.String())
}

func TestFastTemplateJSON(t *testing.T) {
	line := map[string]string{
		"_raw":  "test event",
		"host":  "myhost",
		"index": "main",
	}
	tokens := []Token{}

	ft := BuildFastTemplate(line, tokens, "json")
	assert.NotNil(t, ft)

	var buf bytes.Buffer
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	err := ft.Execute(&buf, now, now, now, randgen)
	assert.NoError(t, err)

	// Parse the output to verify it's valid JSON
	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test event", result["_raw"])
	assert.Equal(t, "myhost", result["host"])
}

func TestFastTemplateSplunkHEC(t *testing.T) {
	line := map[string]string{
		"_raw":  "test event",
		"_time": "1234567890",
		"host":  "myhost",
	}
	tokens := []Token{}

	ft := BuildFastTemplate(line, tokens, "splunkhec")
	assert.NotNil(t, ft)

	var buf bytes.Buffer
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	err := ft.Execute(&buf, now, now, now, randgen)
	assert.NoError(t, err)

	// Parse the output to verify field remapping
	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test event", result["event"]) // _raw -> event
	assert.Equal(t, "1234567890", result["time"])  // _time -> time
	assert.Equal(t, "myhost", result["host"])
}

func TestFastTemplateJSONEscaping(t *testing.T) {
	line := map[string]string{
		"_raw": `Line with "quotes" and \backslash`,
	}
	tokens := []Token{}

	ft := BuildFastTemplate(line, tokens, "json")
	assert.NotNil(t, ft)

	var buf bytes.Buffer
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	err := ft.Execute(&buf, now, now, now, randgen)
	assert.NoError(t, err)

	// Parse to verify escaping worked
	var result map[string]string
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, `Line with "quotes" and \backslash`, result["_raw"])
}

func TestFastTemplateMultipleTokens(t *testing.T) {
	line := map[string]string{
		"_raw": "User $user$ from $ip$ did $action$",
	}
	tokens := []Token{
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
		{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "10.0.0.1", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "login", Field: "_raw"},
	}

	ft := BuildFastTemplate(line, tokens, "raw")
	assert.NotNil(t, ft)

	var buf bytes.Buffer
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	err := ft.Execute(&buf, now, now, now, randgen)
	assert.NoError(t, err)
	assert.Equal(t, "User admin from 10.0.0.1 did login", buf.String())
}

// =============================================================================
// BENCHMARKS - Compare FastTemplate vs Traditional approach
// =============================================================================

// Traditional approach: Copy map → Replace tokens → Marshal JSON
func benchmarkTraditional(b *testing.B, line map[string]string, tokens []Token) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		// 1. Copy the map
		event := make(map[string]string, len(line))
		for k, v := range line {
			event[k] = v
		}

		// 2. Replace tokens
		for i := range tokens {
			t := &tokens[i]
			if t.Format == "template" {
				if fieldVal, ok := event[t.Field]; ok {
					replacement, _, _ := t.GenReplacement(-1, now, now, now, randgen, event)
					// Simple string replacement
					event[t.Field] = replaceFirst(fieldVal, t.Token, replacement)
				}
			}
		}

		// 3. Marshal to JSON
		_, _ = json.Marshal(event)
	}
}

func replaceFirst(s, old, new string) string {
	i := 0
	for {
		idx := indexAt(s, old, i)
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
		i = idx + len(new)
	}
	return s
}

func indexAt(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// FastTemplate approach
func benchmarkFastTemplate(b *testing.B, line map[string]string, tokens []Token) {
	ft := BuildFastTemplate(line, tokens, "json")
	if ft == nil {
		b.Skip("FastTemplate not supported for this format")
	}

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		ft.Execute(&buf, now, now, now, randgen)
	}
}

// Benchmark: Simple event with no tokens
func BenchmarkTraditionalNoTokens(b *testing.B) {
	line := map[string]string{
		"_raw":       "Simple log event with no tokens",
		"host":       "myhost.example.com",
		"source":     "/var/log/app.log",
		"sourcetype": "app:log",
		"index":      "main",
	}
	benchmarkTraditional(b, line, nil)
}

func BenchmarkFastTemplateNoTokens(b *testing.B) {
	line := map[string]string{
		"_raw":       "Simple log event with no tokens",
		"host":       "myhost.example.com",
		"source":     "/var/log/app.log",
		"sourcetype": "app:log",
		"index":      "main",
	}
	benchmarkFastTemplate(b, line, nil)
}

// Benchmark: Event with 3 tokens
func BenchmarkTraditional3Tokens(b *testing.B) {
	line := map[string]string{
		"_raw":       "User $user$ from $ip$ performed $action$",
		"host":       "myhost.example.com",
		"source":     "/var/log/app.log",
		"sourcetype": "app:log",
		"index":      "main",
	}
	tokens := []Token{
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
		{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "192.168.1.100", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "login", Field: "_raw"},
	}
	benchmarkTraditional(b, line, tokens)
}

func BenchmarkFastTemplate3Tokens(b *testing.B) {
	line := map[string]string{
		"_raw":       "User $user$ from $ip$ performed $action$",
		"host":       "myhost.example.com",
		"source":     "/var/log/app.log",
		"sourcetype": "app:log",
		"index":      "main",
	}
	tokens := []Token{
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
		{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "192.168.1.100", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "login", Field: "_raw"},
	}
	benchmarkFastTemplate(b, line, tokens)
}

// Benchmark: Large event with 8 fields and 5 tokens (realistic scenario)
func BenchmarkTraditionalLarge(b *testing.B) {
	line := map[string]string{
		"_raw":       "timestamp=$ts$ user=$user$ src=$srcip$ dst=$dstip$ action=$action$ status=success bytes=1234",
		"host":       "firewall.example.com",
		"source":     "/var/log/firewall/traffic.log",
		"sourcetype": "firewall:traffic",
		"index":      "security",
		"_time":      "1234567890.123",
		"priority":   "info",
		"facility":   "local0",
	}
	tokens := []Token{
		{Name: "ts", Format: "template", Token: "$ts$", Type: "static", Replacement: "2024-01-01T00:00:00Z", Field: "_raw"},
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "jsmith", Field: "_raw"},
		{Name: "srcip", Format: "template", Token: "$srcip$", Type: "static", Replacement: "10.0.0.50", Field: "_raw"},
		{Name: "dstip", Format: "template", Token: "$dstip$", Type: "static", Replacement: "192.168.1.1", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "allow", Field: "_raw"},
	}
	benchmarkTraditional(b, line, tokens)
}

func BenchmarkFastTemplateLarge(b *testing.B) {
	line := map[string]string{
		"_raw":       "timestamp=$ts$ user=$user$ src=$srcip$ dst=$dstip$ action=$action$ status=success bytes=1234",
		"host":       "firewall.example.com",
		"source":     "/var/log/firewall/traffic.log",
		"sourcetype": "firewall:traffic",
		"index":      "security",
		"_time":      "1234567890.123",
		"priority":   "info",
		"facility":   "local0",
	}
	tokens := []Token{
		{Name: "ts", Format: "template", Token: "$ts$", Type: "static", Replacement: "2024-01-01T00:00:00Z", Field: "_raw"},
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "jsmith", Field: "_raw"},
		{Name: "srcip", Format: "template", Token: "$srcip$", Type: "static", Replacement: "10.0.0.50", Field: "_raw"},
		{Name: "dstip", Format: "template", Token: "$dstip$", Type: "static", Replacement: "192.168.1.1", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "allow", Field: "_raw"},
	}
	benchmarkFastTemplate(b, line, tokens)
}

// Benchmark: Batch of 100 events (typical batch size)
func BenchmarkTraditionalBatch100(b *testing.B) {
	line := map[string]string{
		"_raw":       "User $user$ from $ip$ performed $action$",
		"host":       "myhost.example.com",
		"source":     "/var/log/app.log",
		"sourcetype": "app:log",
		"index":      "main",
	}
	tokens := []Token{
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
		{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "192.168.1.100", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "login", Field: "_raw"},
	}

	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		events := make([]map[string]string, 100)
		for i := 0; i < 100; i++ {
			// 1. Copy the map
			event := make(map[string]string, len(line))
			for k, v := range line {
				event[k] = v
			}

			// 2. Replace tokens
			for j := range tokens {
				t := &tokens[j]
				if t.Format == "template" {
					if fieldVal, ok := event[t.Field]; ok {
						replacement, _, _ := t.GenReplacement(-1, now, now, now, randgen, event)
						event[t.Field] = replaceFirst(fieldVal, t.Token, replacement)
					}
				}
			}
			events[i] = event
		}

		// 3. Marshal all to JSON
		for _, event := range events {
			_, _ = json.Marshal(event)
		}
	}
}

func BenchmarkFastTemplateBatch100(b *testing.B) {
	line := map[string]string{
		"_raw":       "User $user$ from $ip$ performed $action$",
		"host":       "myhost.example.com",
		"source":     "/var/log/app.log",
		"sourcetype": "app:log",
		"index":      "main",
	}
	tokens := []Token{
		{Name: "user", Format: "template", Token: "$user$", Type: "static", Replacement: "admin", Field: "_raw"},
		{Name: "ip", Format: "template", Token: "$ip$", Type: "static", Replacement: "192.168.1.100", Field: "_raw"},
		{Name: "action", Format: "template", Token: "$action$", Type: "static", Replacement: "login", Field: "_raw"},
	}

	ft := BuildFastTemplate(line, tokens, "json")
	source := rand.NewSource(0)
	randgen := rand.New(source)
	now := time.Now()
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		ft.ExecuteBatch(&buf, 100, now, now, now, randgen)
	}
}
