package internal

import (
	"bytes"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// FastTemplate is a pre-compiled output template that eliminates map operations
// from the hot path. Instead of Copy → Replace → Marshal → Write, we do:
// Write static bytes → Generate token → Write static bytes → ...
type FastTemplate struct {
	Segments    []FastSegment
	OutputType  string // "raw", "json", "splunkhec", etc.
	TotalStatic int    // Total bytes of static content (for buffer pre-allocation)
}

// FastSegment represents either static bytes or a token to generate
type FastSegment struct {
	Static      []byte // Pre-encoded static content (nil if this is a token segment)
	Token       *Token // Token to generate (nil if this is a static segment)
	NeedsEscape bool   // Whether the token output needs JSON escaping
}

// Pool for output buffers
var fastBufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// BuildFastTemplate creates a pre-compiled template for a sample line
// This is called once at startup, not in the hot path
func BuildFastTemplate(line map[string]string, tokens []Token, outputType string) *FastTemplate {
	ft := &FastTemplate{
		Segments:   make([]FastSegment, 0, 32),
		OutputType: outputType,
	}

	switch outputType {
	case "raw":
		ft.buildRawTemplate(line, tokens)
	case "json":
		ft.buildJSONTemplate(line, tokens, false, false)
	case "splunkhec":
		ft.buildJSONTemplate(line, tokens, true, false) // remap _raw->event, _time->time
	case "elasticsearch":
		ft.buildJSONTemplate(line, tokens, false, true) // remap _raw->message
	default:
		return nil // Unsupported format
	}

	return ft
}

// buildRawTemplate builds a template for raw output (just _raw field)
func (ft *FastTemplate) buildRawTemplate(line map[string]string, tokens []Token) {
	raw, ok := line["_raw"]
	if !ok {
		return
	}

	// Find tokens that apply to the _raw field
	ft.buildFieldSegments(raw, tokens, "_raw", false)
}

// buildJSONTemplate builds a template for JSON output formats
func (ft *FastTemplate) buildJSONTemplate(line map[string]string, tokens []Token, isHEC bool, isES bool) {
	// Start JSON object
	ft.addStatic([]byte(`{`))

	first := true
	for field, value := range line {
		if !first {
			ft.addStatic([]byte(`,`))
		}
		first = false

		// Remap field names for HEC/ES
		outField := field
		if isHEC {
			if field == "_raw" {
				outField = "event"
			} else if field == "_time" {
				outField = "time"
			}
		} else if isES {
			if field == "_raw" {
				outField = "message"
			}
		}

		// Write field name
		ft.addStatic([]byte(`"` + outField + `":`))

		// Check if this field has any tokens
		hasTokens := false
		for i := range tokens {
			if tokens[i].Field == field || (tokens[i].Field == "" && field == "_raw") {
				hasTokens = true
				break
			}
		}

		if hasTokens {
			ft.addStatic([]byte(`"`))
			ft.buildFieldSegments(value, tokens, field, true)
			ft.addStatic([]byte(`"`))
		} else {
			// Static field - pre-encode the JSON value
			ft.addStatic([]byte(`"`))
			ft.addStatic(escapeJSONBytes(value))
			ft.addStatic([]byte(`"`))
		}
	}

	ft.addStatic([]byte(`}`))
}

// buildFieldSegments breaks a field value into static/token segments
func (ft *FastTemplate) buildFieldSegments(value string, tokens []Token, field string, needsEscape bool) {
	// Find all token positions in this field
	type tokenMatch struct {
		start int
		end   int
		token *Token
	}
	var matches []tokenMatch

	for i := range tokens {
		t := &tokens[i]
		if t.Disabled {
			continue
		}
		// Token must apply to this field
		if t.Field != field && !(t.Field == "" && field == "_raw") {
			continue
		}

		if t.Format == "template" {
			// Find all occurrences of the token
			pos := 0
			for {
				idx := strings.Index(value[pos:], t.Token)
				if idx == -1 {
					break
				}
				matches = append(matches, tokenMatch{
					start: pos + idx,
					end:   pos + idx + len(t.Token),
					token: t,
				})
				pos += idx + len(t.Token)
			}
		}
		// Note: regex tokens are more complex and would need special handling
	}

	// Sort matches by position
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].start < matches[i].start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Build segments
	pos := 0
	for _, m := range matches {
		if m.start > pos {
			// Static segment before this token
			staticPart := value[pos:m.start]
			if needsEscape {
				ft.addStatic(escapeJSONBytes(staticPart))
			} else {
				ft.addStatic([]byte(staticPart))
			}
		}
		// Token segment
		ft.addToken(m.token, needsEscape)
		pos = m.end
	}

	// Remaining static content after last token
	if pos < len(value) {
		staticPart := value[pos:]
		if needsEscape {
			ft.addStatic(escapeJSONBytes(staticPart))
		} else {
			ft.addStatic([]byte(staticPart))
		}
	}
}

func (ft *FastTemplate) addStatic(b []byte) {
	if len(b) == 0 {
		return
	}
	// Merge with previous static segment if possible
	if len(ft.Segments) > 0 && ft.Segments[len(ft.Segments)-1].Token == nil {
		ft.Segments[len(ft.Segments)-1].Static = append(ft.Segments[len(ft.Segments)-1].Static, b...)
	} else {
		ft.Segments = append(ft.Segments, FastSegment{Static: b})
	}
	ft.TotalStatic += len(b)
}

func (ft *FastTemplate) addToken(t *Token, needsEscape bool) {
	ft.Segments = append(ft.Segments, FastSegment{Token: t, NeedsEscape: needsEscape})
}

// Execute generates an event using the pre-compiled template
// This is the hot path - optimized for zero allocations where possible
func (ft *FastTemplate) Execute(buf *bytes.Buffer, et, lt, now time.Time, randgen *rand.Rand) error {
	buf.Grow(ft.TotalStatic + 256) // Pre-allocate based on static size + estimated token output

	choices := make(map[int]int, 4) // Small map for token group choices
	fullevent := make(map[string]string, 8) // For tokens that need field lookups

	for i := range ft.Segments {
		seg := &ft.Segments[i]
		if seg.Token == nil {
			// Static segment - just write bytes
			buf.Write(seg.Static)
		} else {
			// Token segment - generate replacement
			var choice int
			if c, ok := choices[seg.Token.Group]; ok {
				choice = c
			} else {
				choice = -1
			}

			replacement, newChoice, err := seg.Token.GenReplacement(choice, et, lt, now, randgen, fullevent)
			if err != nil {
				return err
			}

			if seg.Token.Group > 0 {
				choices[seg.Token.Group] = newChoice
			}

			if seg.NeedsEscape {
				writeEscapedJSON(buf, replacement)
			} else {
				buf.WriteString(replacement)
			}
		}
	}

	return nil
}

// ExecuteBatch generates multiple events efficiently
func (ft *FastTemplate) ExecuteBatch(w *bytes.Buffer, count int, et, lt, now time.Time, randgen *rand.Rand) error {
	// Pre-allocate for all events
	w.Grow((ft.TotalStatic + 256) * count)

	for i := 0; i < count; i++ {
		if i > 0 {
			w.WriteByte('\n')
		}
		if err := ft.Execute(w, et, lt, now, randgen); err != nil {
			return err
		}
	}
	return nil
}

// escapeJSONBytes escapes a string for JSON and returns bytes
func escapeJSONBytes(s string) []byte {
	var buf bytes.Buffer
	buf.Grow(len(s) + 10)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if c < 0x20 {
				buf.WriteString(`\u00`)
				buf.WriteByte("0123456789abcdef"[c>>4])
				buf.WriteByte("0123456789abcdef"[c&0xf])
			} else {
				buf.WriteByte(c)
			}
		}
	}
	return buf.Bytes()
}

// writeEscapedJSON writes a JSON-escaped string to the buffer
func writeEscapedJSON(buf *bytes.Buffer, s string) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if c < 0x20 {
				buf.WriteString(`\u00`)
				buf.WriteByte("0123456789abcdef"[c>>4])
				buf.WriteByte("0123456789abcdef"[c&0xf])
			} else {
				buf.WriteByte(c)
			}
		}
	}
}
