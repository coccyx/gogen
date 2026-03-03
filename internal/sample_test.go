package internal

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetReplacementOffsets(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", "tokens-find.yml"))

	c := NewConfig()
	s := c.FindSampleByName("tokens-find")

	_, err := s.Tokens[0].GetReplacementOffsets("foo")
	assert.Error(t, err)

	offsets, err := s.Tokens[0].GetReplacementOffsets(s.Lines[0]["_raw"])
	assert.NoError(t, err)
	assert.Equal(t, 4, offsets[0][0])
	assert.Equal(t, 14, offsets[0][1])
	assert.Equal(t, 15, offsets[1][0])
	assert.Equal(t, 25, offsets[1][1])

	offsets, err = s.Tokens[1].GetReplacementOffsets(s.Lines[0]["_raw"])
	assert.NoError(t, err)
	assert.Equal(t, 4, offsets[0][0])
	assert.Equal(t, 14, offsets[0][1])
	assert.Equal(t, 15, offsets[1][0])
	assert.Equal(t, 25, offsets[1][1])
}

func TestReplacement(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", "tokens-find.yml"))
	loc, _ := time.LoadLocation("UTC")
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	c := NewConfig()
	s := c.FindSampleByName("tokens-find")

	event := "foo"
	_, err := s.Tokens[0].Replace(&event, -1, now(), now(), now(), randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "foo", event)

	event = s.Lines[0]["_raw"]
	_, err = s.Tokens[0].Replace(&event, -1, now(), now(), now(), randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "foo newfoo newfoo", event)

	event = s.Lines[0]["_raw"]
	_, err = s.Tokens[1].Replace(&event, -1, now(), now(), now(), randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "foo newfoo newfoo", event)

	event = s.Lines[1]["_raw"]
	_, err = s.Tokens[0].Replace(&event, -1, now(), now(), now(), randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "newfoo foo newfoo foo some other", event)

	event = s.Lines[1]["_raw"]
	_, err = s.Tokens[1].Replace(&event, -1, now(), now(), now(), randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "newfoo foo newfoo foo some other", event)
}

func TestGenReplacement(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", "tokens.yml"))
	loc, _ := time.LoadLocation("UTC")
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	c := NewConfig()
	s := c.FindSampleByName("tokens")

	testToken(0, "foo", s, t)
	testToken(1, "4", s, t)
	testToken(2, "0.274", s, t)
	testToken(3, "mUNERA9rI2", s, t)
	testToken(4, "4C345", s, t)
	testToken(7, "3", s, t)
	testToken(9, "159.144.163.226", s, t)
	testToken(10, "a8f6:236d:b3ef:c41e:4808:d6ed:ecb0:4067", s, t)
	testToken(11, "2001-10-20 12:00:00.000", s, t)
	testToken(12, "2001-10-20 12:00:00.000", s, t)
	testToken(13, "1003579200", s, t)

	token := s.Tokens[5]
	replacement, _, _ := token.GenReplacement(-1, now(), now(), now(), randgen, fullevent)
	assert.Equal(t, "a", replacement)
	replacement, _, _ = token.GenReplacement(-1, now(), now(), now(), randgen, fullevent)
	assert.Equal(t, "a", replacement)

	token = s.Tokens[6]
	choices := make(map[int]int)
	for i := 0; i < 1000; i++ {
		_, choice, _ := token.GenReplacement(-1, now(), now(), now(), randgen, fullevent)
		choices[choice] = choices[choice] + 1
	}
	if choices[0] != 315 || choices[1] != 570 || choices[2] != 115 {
		t.Fatalf("Choice distribution is off: %#v\n", choices)
	}

	token = s.Tokens[8]
	replacement, _, _ = token.GenReplacement(-1, now(), now(), now(), randgen, fullevent)
	fmt.Printf("UUID: %s\n", replacement)
}

func testToken(i int, value string, s *Sample, t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}
	token := s.Tokens[i]
	replacement, _, _ := token.GenReplacement(-1, now(), now(), now(), randgen, fullevent)
	assert.Equal(t, value, replacement)
}

func TestLuaReplacement(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", "lua.yml"))

	c := NewConfig()
	s := c.FindSampleByName("lua")

	testToken(0, "foo", s, t)
	testToken(1, "3", s, t)
	testToken(2, "0.945", s, t)
	testToken(3, "NvofsbSj4", s, t)
	testToken(4, "4C345", s, t)
}

func TestParseTimestamp(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", "tokens.yml"))
	loc, _ := time.LoadLocation("UTC")

	n := time.Date(2001, 10, 20, 12, 0, 0, 0, loc)

	c := NewConfig()
	s := c.FindSampleByName("tokens")

	token := s.Tokens[11]
	ts, _ := token.ParseTimestamp("2001-10-20 12:00:00.000")
	assert.Equal(t, n, ts)
	token = s.Tokens[12]
	ts, _ = token.ParseTimestamp("2001-10-20 12:00:00.000")
	assert.Equal(t, n, ts)
	token = s.Tokens[13]
	ts, _ = token.ParseTimestamp("1003579200")
	assert.Equal(t, n.Local(), ts)
}

func BenchmarkGoStatic(b *testing.B)      { benchmarkToken("tokens", 0, b) }
func BenchmarkGoRandInt(b *testing.B)     { benchmarkToken("tokens", 1, b) }
func BenchmarkGoRandFloat(b *testing.B)   { benchmarkToken("tokens", 2, b) }
func BenchmarkGoRandString(b *testing.B)  { benchmarkToken("tokens", 3, b) }
func BenchmarkGoRandHex(b *testing.B)     { benchmarkToken("tokens", 4, b) }
func BenchmarkLuaStatic(b *testing.B)     { benchmarkToken("lua", 0, b) }
func BenchmarkLuaRandInt(b *testing.B)    { benchmarkToken("lua", 1, b) }
func BenchmarkLuaRandFloat(b *testing.B)  { benchmarkToken("lua", 2, b) }
func BenchmarkLuaRandString(b *testing.B) { benchmarkToken("lua", 3, b) }
func BenchmarkLuaRandHex(b *testing.B)    { benchmarkToken("lua", 4, b) }

func benchmarkToken(conf string, i int, b *testing.B) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", conf+".yml"))

	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	c := NewConfig()
	s := c.FindSampleByName(conf)

	for n := 0; n < b.N; n++ {
		token := s.Tokens[i]
		_, _, _ = token.GenReplacement(-1, now(), now(), now(), randgen, fullevent)
	}
}

// mockRater implements Rater for testing rated tokens
type mockRater struct {
	rate float64
}

func (m *mockRater) EventRate(s *Sample, now time.Time, count int) float64 { return m.rate }
func (m *mockRater) TokenRate(t Token, now time.Time) float64              { return m.rate }

func TestGenReplacementRatedInt(t *testing.T) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	now := time.Now()

	token := Token{
		Name:        "ratedint",
		Type:        "rated",
		Replacement: "int",
		Lower:       10,
		Upper:       20,
		Rater:       &mockRater{rate: 2.0},
	}
	result, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	// With rate=2.0, value should be roughly doubled
	val, _ := fmt.Sscanf(result, "%d", new(int))
	assert.Equal(t, 1, val, "should parse as an integer")
}

func TestGenReplacementRatedIntEqualBounds(t *testing.T) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	now := time.Now()

	token := Token{
		Name:        "ratedintequal",
		Type:        "rated",
		Replacement: "int",
		Lower:       5,
		Upper:       5,
		Rater:       &mockRater{rate: 1.0},
	}
	result, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "5", result)
}

func TestGenReplacementRatedFloat(t *testing.T) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	now := time.Now()

	token := Token{
		Name:        "ratedfloat",
		Type:        "rated",
		Replacement: "float",
		Lower:       1,
		Upper:       10,
		Precision:   2,
		Rater:       &mockRater{rate: 1.5},
	}
	result, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	// Should be a float string with 2 decimal places
	assert.Contains(t, result, ".")
}

func TestGenReplacementRatedFloatEqualBounds(t *testing.T) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	now := time.Now()

	token := Token{
		Name:        "ratedfloatequal",
		Type:        "rated",
		Replacement: "float",
		Lower:       5,
		Upper:       5,
		Precision:   2,
		Rater:       &mockRater{rate: 1.0},
	}
	result, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "5.00", result)
}

func TestGenReplacementFieldChoice(t *testing.T) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	now := time.Now()

	token := Token{
		Name:     "fieldchoice",
		Type:     "fieldChoice",
		SrcField: "city",
		FieldChoice: []map[string]string{
			{"city": "NYC", "state": "NY"},
			{"city": "LA", "state": "CA"},
		},
	}
	result, choice, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	assert.True(t, result == "NYC" || result == "LA")
	assert.True(t, choice >= 0)

	// Test with specific choice
	result, _, err = token.GenReplacement(0, now, now, now, randgen, fullevent)
	assert.NoError(t, err)
	assert.Equal(t, "NYC", result)
}

func TestGenReplacementInvalidType(t *testing.T) {
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)
	now := time.Now()

	token := Token{
		Name: "badtype",
		Type: "nonexistenttype",
	}
	_, _, err := token.GenReplacement(-1, now, now, now, randgen, fullevent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func BenchmarkReplacement(b *testing.B) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_SAMPLES_DIR", filepath.Join(home, "tests", "tokens", "token-static.yml"))

	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)
	fullevent := make(map[string]string)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	c := NewConfig()
	s := c.FindSampleByName("token-static")
	t := s.Tokens[0]

	event := "$static$"

	for n := 0; n < b.N; n++ {
		_, _ = t.Replace(&event, -1, now(), now(), now(), randgen, fullevent)
		event = "$static$"
	}
}
