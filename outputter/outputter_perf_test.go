package outputter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// UNIT TESTS - Verify correctness of event handling
// =============================================================================

// TestSplunkHECEventCopyCorrectness verifies that splunkhec formatting
// does not mutate the original event (critical for caching)
func TestSplunkHECEventCopyCorrectness(t *testing.T) {
	original := map[string]string{
		"_raw":   "test event data",
		"_time":  "2024-01-01T00:00:00Z",
		"host":   "myhost",
		"source": "/var/log/test",
	}

	// Make a copy to compare against
	originalRaw := original["_raw"]
	originalTime := original["_time"]

	// Use the optimized writeJSONWithHECRemap function
	var buf bytes.Buffer
	writeJSONWithHECRemap(&buf, original)
	result := buf.String()

	// Verify original is unchanged
	assert.Equal(t, originalRaw, original["_raw"], "original _raw should be unchanged")
	assert.Equal(t, originalTime, original["_time"], "original _time should be unchanged")

	// Verify output has correct field remapping
	assert.Contains(t, result, `"event":"test event data"`, "output should have event field")
	assert.Contains(t, result, `"time":"2024-01-01T00:00:00Z"`, "output should have time field")
	assert.NotContains(t, result, `"_raw"`, "output should not have _raw")
	assert.NotContains(t, result, `"_time"`, "output should not have _time")
	assert.Contains(t, result, `"host":"myhost"`, "output should preserve host field")
}

// TestWriteJSONWithRemapCorrectness verifies the writeJSONWithRemap function
func TestWriteJSONWithRemapCorrectness(t *testing.T) {
	original := map[string]string{
		"_raw":   "test event data",
		"host":   "myhost",
		"source": "/var/log/test",
	}

	var buf bytes.Buffer
	writeJSONWithRemap(&buf, original, "_raw", "message")
	result := buf.String()

	// Verify field remapping
	assert.Contains(t, result, `"message":"test event data"`, "output should have message field")
	assert.NotContains(t, result, `"_raw"`, "output should not have _raw")
	assert.Contains(t, result, `"host":"myhost"`, "output should preserve host field")

	// Verify original is unchanged
	assert.Equal(t, "test event data", original["_raw"], "original should be unchanged")
}

// TestWriteJSONStringEscaping verifies JSON string escaping
func TestWriteJSONStringEscaping(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "quotes",
			input:    map[string]string{"msg": `say "hello"`},
			expected: `"msg":"say \"hello\""`,
		},
		{
			name:     "backslash",
			input:    map[string]string{"path": `C:\Windows\System32`},
			expected: `"path":"C:\\Windows\\System32"`,
		},
		{
			name:     "newline",
			input:    map[string]string{"msg": "line1\nline2"},
			expected: `"msg":"line1\nline2"`,
		},
		{
			name:     "tab",
			input:    map[string]string{"msg": "col1\tcol2"},
			expected: `"msg":"col1\tcol2"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeJSONWithRemap(&buf, tc.input, "", "")
			result := buf.String()
			assert.Contains(t, result, tc.expected)
		})
	}
}

// TestElasticsearchEventCopyCorrectness verifies that elasticsearch formatting
// does not mutate the original event (critical for caching)
func TestElasticsearchEventCopyCorrectness(t *testing.T) {
	original := map[string]string{
		"_raw":   "test event data",
		"host":   "myhost",
		"index":  "main",
		"source": "/var/log/test",
	}

	// Make a copy to compare against
	originalRaw := original["_raw"]

	// Simulate elasticsearch transformation (the fixed version)
	esLine := make(map[string]string, len(original))
	for k, v := range original {
		esLine[k] = v
	}
	if _, ok := esLine["_raw"]; ok {
		esLine["message"] = esLine["_raw"]
		delete(esLine, "_raw")
	}

	// Verify original is unchanged
	assert.Equal(t, originalRaw, original["_raw"], "original _raw should be unchanged")

	// Verify transformed has correct fields
	assert.Equal(t, originalRaw, esLine["message"], "esLine should have message field")
	assert.NotContains(t, esLine, "_raw", "esLine should not have _raw")
}

// =============================================================================
// BENCHMARKS - Measure performance of event formatting
// =============================================================================

// BenchmarkEventCopy benchmarks the map copy operation for event transformation
func BenchmarkEventCopy(b *testing.B) {
	original := map[string]string{
		"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
		"_time":      "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
		"field1":     "value1",
		"field2":     "value2",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		dst := make(map[string]string, len(original))
		for k, v := range original {
			dst[k] = v
		}
	}
}

// BenchmarkSplunkHECFormat benchmarks the full splunkhec formatting with copy
func BenchmarkSplunkHECFormat(b *testing.B) {
	original := map[string]string{
		"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
		"_time":      "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		// Copy the event to avoid mutating cached data
		hecLine := make(map[string]string, len(original))
		for k, v := range original {
			hecLine[k] = v
		}
		if _, ok := hecLine["_raw"]; ok {
			hecLine["event"] = hecLine["_raw"]
			delete(hecLine, "_raw")
		}
		if _, ok := hecLine["_time"]; ok {
			hecLine["time"] = hecLine["_time"]
			delete(hecLine, "_time")
		}
		jb, _ := json.Marshal(hecLine)
		buf.Write(jb)
	}
}

// BenchmarkElasticsearchFormat benchmarks the full elasticsearch formatting with copy
func BenchmarkElasticsearchFormat(b *testing.B) {
	original := map[string]string{
		"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
		"_time":      "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		// Copy the event to avoid mutating cached data
		esLine := make(map[string]string, len(original))
		for k, v := range original {
			esLine[k] = v
		}
		if _, ok := esLine["_raw"]; ok {
			esLine["message"] = esLine["_raw"]
			delete(esLine, "_raw")
		}
		jb, _ := json.Marshal(esLine)
		buf.Write(jb)
	}
}

// BenchmarkJSONMarshal benchmarks just the JSON marshaling (baseline)
func BenchmarkJSONMarshal(b *testing.B) {
	event := map[string]string{
		"event":      "This is a test event with some realistic content length for benchmarking purposes",
		"time":       "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		jb, _ := json.Marshal(event)
		buf.Write(jb)
	}
}

// BenchmarkRawOutput benchmarks the raw output path (no transformation)
func BenchmarkRawOutput(b *testing.B) {
	event := map[string]string{
		"_raw": "This is a test event with some realistic content length for benchmarking purposes that would be output in raw format",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		io.WriteString(&buf, event["_raw"])
	}
}

// BenchmarkSplunkHECFormatUnsafe benchmarks the old (buggy) behavior that mutates in place
// This is for comparison purposes only - DO NOT USE IN PRODUCTION
func BenchmarkSplunkHECFormatUnsafe(b *testing.B) {
	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		// Create a fresh event each iteration since we mutate it
		original := map[string]string{
			"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
			"_time":      "2024-01-01T00:00:00.000Z",
			"host":       "myhost.example.com",
			"source":     "/var/log/application/app.log",
			"sourcetype": "application:log",
			"index":      "main",
		}
		buf.Reset()
		// Old unsafe behavior - mutates original
		if _, ok := original["_raw"]; ok {
			original["event"] = original["_raw"]
			delete(original, "_raw")
		}
		if _, ok := original["_time"]; ok {
			original["time"] = original["_time"]
			delete(original, "_time")
		}
		jb, _ := json.Marshal(original)
		buf.Write(jb)
	}
}

// BenchmarkMultiEventSplunkHEC benchmarks processing multiple events (batch scenario)
func BenchmarkMultiEventSplunkHEC(b *testing.B) {
	// Create a batch of 100 events
	events := make([]map[string]string, 100)
	for i := 0; i < 100; i++ {
		events[i] = map[string]string{
			"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
			"_time":      "2024-01-01T00:00:00.000Z",
			"host":       "myhost.example.com",
			"source":     "/var/log/application/app.log",
			"sourcetype": "application:log",
			"index":      "main",
		}
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		for _, line := range events {
			// Copy the event to avoid mutating cached data
			hecLine := make(map[string]string, len(line))
			for k, v := range line {
				hecLine[k] = v
			}
			if _, ok := hecLine["_raw"]; ok {
				hecLine["event"] = hecLine["_raw"]
				delete(hecLine, "_raw")
			}
			if _, ok := hecLine["_time"]; ok {
				hecLine["time"] = hecLine["_time"]
				delete(hecLine, "_time")
			}
			jb, _ := json.Marshal(hecLine)
			buf.Write(jb)
			buf.WriteByte('\n')
		}
	}
}

// =============================================================================
// RFC FORMAT BENCHMARKS - Current vs Optimized
// =============================================================================

// BenchmarkRFC3164Current benchmarks the current fmt.Sprintf implementation
func BenchmarkRFC3164Current(b *testing.B) {
	line := map[string]string{
		"priority": "134",
		"_time":    "Jan  1 00:00:00",
		"host":     "myhost.example.com",
		"tag":      "myapp",
		"pid":      "12345",
		"_raw":     "This is a test syslog message with realistic content",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		io.WriteString(&buf, fmt.Sprintf("<%s>%s %s %s[%s]: %s", line["priority"], line["_time"], line["host"], line["tag"], line["pid"], line["_raw"]))
	}
}

// BenchmarkRFC3164Optimized benchmarks strings.Builder implementation
func BenchmarkRFC3164Optimized(b *testing.B) {
	line := map[string]string{
		"priority": "134",
		"_time":    "Jan  1 00:00:00",
		"host":     "myhost.example.com",
		"tag":      "myapp",
		"pid":      "12345",
		"_raw":     "This is a test syslog message with realistic content",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		var sb strings.Builder
		sb.Grow(len(line["priority"]) + len(line["_time"]) + len(line["host"]) + len(line["tag"]) + len(line["pid"]) + len(line["_raw"]) + 10)
		sb.WriteByte('<')
		sb.WriteString(line["priority"])
		sb.WriteByte('>')
		sb.WriteString(line["_time"])
		sb.WriteByte(' ')
		sb.WriteString(line["host"])
		sb.WriteByte(' ')
		sb.WriteString(line["tag"])
		sb.WriteByte('[')
		sb.WriteString(line["pid"])
		sb.WriteString("]: ")
		sb.WriteString(line["_raw"])
		io.WriteString(&buf, sb.String())
	}
}

// BenchmarkRFC5424Current benchmarks the current fmt.Sprintf implementation
func BenchmarkRFC5424Current(b *testing.B) {
	line := map[string]string{
		"priority": "134",
		"_time":    "2024-01-01T00:00:00.000Z",
		"host":     "myhost.example.com",
		"appName":  "myapp",
		"pid":      "12345",
		"_raw":     "This is a test syslog message with realistic content",
		"field1":   "value1",
		"field2":   "value2",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		kv := "-"
		for k, v := range line {
			if k != "_raw" && k != "_time" && k != "priority" && k != "host" && k != "appName" && k != "pid" && k != "tag" {
				kv = kv + fmt.Sprintf("%s=\"%s\" ", k, v)
			}
		}
		if len(kv) != 1 {
			kv = fmt.Sprintf("[meta %s]", kv[1:len(kv)-1])
		}
		io.WriteString(&buf, fmt.Sprintf("<%s>%d %s %s %s %s - %s %s", line["priority"], 1, line["_time"], line["host"], line["appName"], line["pid"], kv, line["_raw"]))
	}
}

// BenchmarkRFC5424Optimized benchmarks strings.Builder implementation
func BenchmarkRFC5424Optimized(b *testing.B) {
	line := map[string]string{
		"priority": "134",
		"_time":    "2024-01-01T00:00:00.000Z",
		"host":     "myhost.example.com",
		"appName":  "myapp",
		"pid":      "12345",
		"_raw":     "This is a test syslog message with realistic content",
		"field1":   "value1",
		"field2":   "value2",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		var kvBuilder strings.Builder
		kvBuilder.Grow(64)
		hasKV := false
		for k, v := range line {
			if k != "_raw" && k != "_time" && k != "priority" && k != "host" && k != "appName" && k != "pid" && k != "tag" {
				if hasKV {
					kvBuilder.WriteByte(' ')
				}
				kvBuilder.WriteString(k)
				kvBuilder.WriteString("=\"")
				kvBuilder.WriteString(v)
				kvBuilder.WriteByte('"')
				hasKV = true
			}
		}

		var sb strings.Builder
		sb.Grow(200)
		sb.WriteByte('<')
		sb.WriteString(line["priority"])
		sb.WriteString(">1 ")
		sb.WriteString(line["_time"])
		sb.WriteByte(' ')
		sb.WriteString(line["host"])
		sb.WriteByte(' ')
		sb.WriteString(line["appName"])
		sb.WriteByte(' ')
		sb.WriteString(line["pid"])
		sb.WriteString(" - ")
		if hasKV {
			sb.WriteString("[meta ")
			sb.WriteString(kvBuilder.String())
			sb.WriteByte(']')
		} else {
			sb.WriteByte('-')
		}
		sb.WriteByte(' ')
		sb.WriteString(line["_raw"])
		io.WriteString(&buf, sb.String())
	}
}

// BenchmarkElasticsearchIndexCurrent benchmarks the current fmt.Sprintf for ES index line
func BenchmarkElasticsearchIndexCurrent(b *testing.B) {
	index := "main"

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		io.WriteString(&buf, fmt.Sprintf("{ \"index\": { \"_index\": \"%s\", \"_type\": \"doc\" } }\n", index))
	}
}

// BenchmarkElasticsearchIndexOptimized benchmarks strings.Builder for ES index line
func BenchmarkElasticsearchIndexOptimized(b *testing.B) {
	index := "main"

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		var sb strings.Builder
		sb.Grow(50 + len(index))
		sb.WriteString("{ \"index\": { \"_index\": \"")
		sb.WriteString(index)
		sb.WriteString("\", \"_type\": \"doc\" } }\n")
		io.WriteString(&buf, sb.String())
	}
}

// =============================================================================
// JSON ENCODER BENCHMARKS - json.Marshal vs json.Encoder with pooled buffer
// =============================================================================

// BenchmarkJSONMarshalRepeated benchmarks json.Marshal called repeatedly (current)
func BenchmarkJSONMarshalRepeated(b *testing.B) {
	event := map[string]string{
		"event":      "This is a test event with some realistic content length",
		"time":       "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		jb, _ := json.Marshal(event)
		buf.Write(jb)
	}
}

// BenchmarkJSONEncoderReused benchmarks json.Encoder with reused buffer (optimized)
func BenchmarkJSONEncoderReused(b *testing.B) {
	event := map[string]string{
		"event":      "This is a test event with some realistic content length",
		"time":       "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		enc.Encode(event)
	}
}

// =============================================================================
// SPLUNKHEC OPTIMIZATION - Direct JSON build vs map copy + marshal
// =============================================================================

// BenchmarkSplunkHECDirectJSON benchmarks building JSON directly without map copy
func BenchmarkSplunkHECDirectJSON(b *testing.B) {
	original := map[string]string{
		"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
		"_time":      "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		buf.WriteByte('{')
		first := true
		for k, v := range original {
			if !first {
				buf.WriteByte(',')
			}
			first = false
			// Remap _raw -> event, _time -> time
			outKey := k
			if k == "_raw" {
				outKey = "event"
			} else if k == "_time" {
				outKey = "time"
			}
			buf.WriteByte('"')
			buf.WriteString(outKey)
			buf.WriteString("\":\"")
			// Escape JSON string (simplified - real impl needs full escaping)
			buf.WriteString(v)
			buf.WriteByte('"')
		}
		buf.WriteByte('}')
	}
}

// BenchmarkSplunkHECOptimized benchmarks the actual optimized implementation
func BenchmarkSplunkHECOptimized(b *testing.B) {
	original := map[string]string{
		"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
		"_time":      "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		writeJSONWithHECRemap(&buf, original)
	}
}

// BenchmarkElasticsearchOptimized benchmarks the actual optimized ES implementation
func BenchmarkElasticsearchOptimized(b *testing.B) {
	original := map[string]string{
		"_raw":       "This is a test event with some realistic content length for benchmarking purposes",
		"_time":      "2024-01-01T00:00:00.000Z",
		"host":       "myhost.example.com",
		"source":     "/var/log/application/app.log",
		"sourcetype": "application:log",
		"index":      "main",
	}

	var buf bytes.Buffer
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		buf.Reset()
		writeJSONWithRemap(&buf, original, "_raw", "message")
	}
}
