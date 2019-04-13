/*
Implementation of Python's strftime in Go

Example:
    str, err := strftime.Format("%Y/%m/%d", time.Now()) // 2012/12/07

Directives:
    %a - Locale’s abbreviated weekday name
    %A - Locale’s full weekday name
    %b - Locale’s abbreviated month name
    %B - Locale’s full month name
    %c - Locale’s appropriate date and time representation
    %d - Day of the month as a decimal number [01,31]
    %H - Hour (24-hour clock) as a decimal number [00,23]
    %I - Hour (12-hour clock) as a decimal number [01,12]
    %j - Day of year
    %m - Month as a decimal number [01,12]
    %M - Minute as a decimal number [00,59]
    %p - Locale’s equivalent of either AM or PM
    %S - Second as a decimal number [00,61]
    %U - Week number of the year
    %w - Weekday as a decimal number
    %W - Week number of the year
    %x - Locale’s appropriate date representation
    %X - Locale’s appropriate time representation
    %y - Year without century as a decimal number [00,99]
    %Y - Year with century as a decimal number
    %Z - Time zone name (no characters if no time zone exists)

Note that %c returns RFC1123 which is a bit different from what Python does
*/
package strftime

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"time"
)

const (
	WEEK = time.Hour * 24 * 7
)

type FormatFunc func(t time.Time) string

func weekNumberFormatter(t time.Time) string {
	start := time.Date(t.Year(), time.January, 1, 23, 0, 0, 0, time.UTC)
	week := 0
	for start.Before(t) {
		week += 1
		start = start.Add(WEEK)
	}
	return fmt.Sprintf("%02d", week)
}

// See http://docs.python.org/2/library/time.html#time.strftime
var formats = map[string]FormatFunc{
	"%a": func(t time.Time) string { // Locale’s abbreviated weekday name
		return t.Format("Mon")
	},
	"%A": func(t time.Time) string { // Locale’s full weekday name
		return t.Format("Monday")
	},
	"%b": func(t time.Time) string { // Locale’s abbreviated month name
		return t.Format("Jan")
	},
	"%B": func(t time.Time) string { // Locale’s full month name
		return t.Format("January")
	},
	"%c": func(t time.Time) string { // Locale’s appropriate date and time representation
		return t.Format(time.RFC1123)
	},
	"%d": func(t time.Time) string { // Day of the month as a decimal number [01,31]
		return t.Format("02")
	},
	"%H": func(t time.Time) string { // Hour (24-hour clock) as a decimal number [00,23]
		return t.Format("15")
	},
	"%I": func(t time.Time) string { // Hour (12-hour clock) as a decimal number [01,12]
		return t.Format("3")
	},
	"%j": func(t time.Time) string {
		start := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		day := int(t.Sub(start).Hours()/24) + 1
		return fmt.Sprintf("%03d", day)
	},
	"%m": func(t time.Time) string { // Month as a decimal number [01,12]
		return t.Format("01")
	},
	"%M": func(t time.Time) string { // Minute as a decimal number [00,59]
		return t.Format("04")
	},
	"%p": func(t time.Time) string { // Locale’s equivalent of either AM or PM
		return t.Format("PM")
	},
	"%S": func(t time.Time) string { // Second as a decimal number [00,61]
		return t.Format("05")
	},
	"%U": weekNumberFormatter, // Week number of the year
	"%W": weekNumberFormatter, // Week number of the year
	"%w": func(t time.Time) string { // Weekday as a decimal number
		return fmt.Sprintf("%d", t.Weekday())
	},
	"%x": func(t time.Time) string { // Locale’s appropriate date representation
		return t.Format("01/02/06")
	},
	"%X": func(t time.Time) string { // Locale’s appropriate time representation
		return t.Format("15:04:05")
	},
	"%y": func(t time.Time) string { // Year without century as a decimal number [00,99]
		return t.Format("06")
	},
	"%Y": func(t time.Time) string { // Year with century as a decimal number
		return t.Format("2006")
	},
	"%Z": func(t time.Time) string { // Time zone name (no characters if no time zone exists)
		return t.Format("MST")
	},
}

var (
	//	fmtRe      = regexp.MustCompile("%([%aAbBcdHIjmMpSUwWxXyYZ]|[1-9]n)")
	fmtRe          = initFormatRegexp()
	fmtBackquoteRe = initFormatBackquoteRegexp()
)

func initFormatRegexp() *regexp.Regexp {
	var buf bytes.Buffer
	buf.WriteString("%([%")
	for format, _ := range formats {
		buf.WriteString(regexp.QuoteMeta(format[1:]))
	}
	buf.WriteString("]|[1-9]n)")
	re := buf.String()
	return regexp.MustCompile(re)
}

func initFormatBackquoteRegexp() *regexp.Regexp {
	var buf bytes.Buffer
	buf.WriteString("%([^")
	for format, _ := range formats {
		buf.WriteString(regexp.QuoteMeta(format[1:]))
	}
	buf.WriteString("1-9]|[1-9][^n])")
	re := buf.String()
	return regexp.MustCompile(re)
}

// A load from pkg/time/format.go of golang source code.
// formatNano appends a fractional second, as nanoseconds, to b
// and returns the result.
func formatNano(nanosec uint, n int, trim bool) []byte {
	u := nanosec
	var buf [9]byte
	for start := len(buf); start > 0; {
		start--
		buf[start] = byte(u%10 + '0')
		u /= 10
	}

	if n > 9 {
		n = 9
	}
	if trim {
		for n > 0 && buf[n-1] == '0' {
			n--
		}
		if n == 0 {
			return buf[:0]
		}
	}
	return buf[:n]
}

func formatNanoForMatch(match string, t time.Time) string {
	// format nanosecond for a match format %[1-9]n
	size := int(match[1] - '0')
	return string(formatNano(uint(t.Nanosecond()), size, false))
}

// repl replaces % directives with right time
func repl(match string, t time.Time) string {
	if match == "%%" {
		return "%"
	}

	formatFunc, ok := formats[match]
	if ok {
		return formatFunc(t)
	}
	return formatNanoForMatch(match, t)
}

// Format return string with % directives expanded.
// Will return error on unknown directive.
func Format(format string, t time.Time) string {
	fn := func(match string) string {
		return repl(match, t)
	}
	return fmtRe.ReplaceAllStringFunc(format, fn)
}

func FormatTo(w io.Writer, format string, t time.Time) (n int, err error) {
	result := Format(format, t)
	return w.Write([]byte(result))
}

type Formatter struct {
	format     string
	strFormat  string
	formatFunc func(t time.Time) []interface{}
}

func NewFormatter(format string) *Formatter {
	f := func(match string) string {
		if match == "%%" {
			return match
		}
		return "%" + match
	}
	strFormat := fmtBackquoteRe.ReplaceAllStringFunc(format, f)
	size := 0
	f1 := func(match string) string {
		if match == "%%" {
			return match
		}
		size++
		return "%s"
	}
	strFormat = fmtRe.ReplaceAllStringFunc(strFormat, f1)
	funs := make([]FormatFunc, 0, size)
	f2 := func(match string) string {
		if match == "%%" {
			return match
		}
		f, ok := formats[match]
		if ok {
			funs = append(funs, f)
		} else {
			f := func(t time.Time) string {
				return formatNanoForMatch(match, t)
			}
			funs = append(funs, f)
		}
		return match
	}
	fmtRe.ReplaceAllStringFunc(format, f2)
	formatFunc := func(t time.Time) []interface{} {
		result := make([]interface{}, 0, len(funs))
		for _, f := range funs {
			result = append(result, f(t))
		}
		return result
	}
	return &Formatter{
		format:     format,
		strFormat:  strFormat,
		formatFunc: formatFunc,
	}
}

func (self *Formatter) Format(t time.Time) string {
	return fmt.Sprintf(self.strFormat, self.formatFunc(t)...)
}

func (self *Formatter) FormatTo(w io.Writer, t time.Time) (n int, err error) {
	return fmt.Fprintf(w, self.strFormat, self.formatFunc(t)...)
}
