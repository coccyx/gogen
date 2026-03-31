package generator

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestLuaGen(t *testing.T) {
	config.ResetConfig()
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "realGenerator.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("generator")

	gen := new(luagen)
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 walked in")
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 sat down")
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 on the group w bench")
	testLuaGen(t, s, gen, "20/Oct/2001 12:00:00 walked in")
}

func TestSetToken(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("setToken")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	for _, t := range gen.tokens {
		if t.Name == "test" {
			found = true
		}
	}
	assert.True(t, found, "Couldn't find token 'test' in sample setToken")
}

func TestGetChoice(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getChoice")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "getChoice" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'getChoice' in sample getChoice")
	assert.Equal(t, "foo", token.Replacement)
}

func TestGetChoiceItem(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getChoiceItem")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "getChoiceItem" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'getChoiceItem' in sample getChoiceItem")
	assert.Equal(t, "bar", token.Replacement)
}

func TestGetFieldChoice(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getFieldChoice")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "getFieldChoice" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'getFieldChoice' in sample getFieldChoice")
	assert.Equal(t, "foo", token.Replacement)
}

func TestGetFieldChoiceItem(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getFieldChoiceItem")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "getFieldChoiceItem" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'getFieldChoiceItem' in sample getFieldChoiceItem")
	assert.Equal(t, "bar", token.Replacement)
}

func TestGetWeightedChoiceItem(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getWeightedChoiceItem")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "getWeightedChoiceItem" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'getWeightedChoiceItem' in sample getWeightedChoiceItem")
	assert.Equal(t, "foo", token.Replacement)
}

func TestGetGroupIdx(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getGroupIdx")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "getGroupIdx" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'getGroupIdx' in sample getGroupIdx")
	assert.Equal(t, "0", token.Replacement)
}

func TestGetLine(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getLine")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "line" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'line' in sample getLine")
	assert.Equal(t, "foo", token.Replacement)
}

func TestGetLines(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("getLines")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, t := range gen.tokens {
		if t.Name == "line1" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'line1' in sample getLines")
	assert.Equal(t, "foo", token.Replacement)
	found = false
	for _, t := range gen.tokens {
		if t.Name == "line2" {
			found = true
			token = t
		}
	}
	assert.True(t, found, "Couldn't find token 'line2' in sample getLines")
	assert.Equal(t, "bar", token.Replacement)
}

func TestReplaceTokens(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("replaceTokens")
	gen := new(luagen)
	testLuaGen(t, s, gen, "foo")
}

func TestSetTime(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("setTime")
	gen := new(luagen)
	testLuaGen(t, s, gen, "2001-10-20 11:59:59.000100")
}

func TestLuaRound(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi2.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("roundTest")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, tk := range gen.tokens {
		if tk.Name == "rounded" {
			found = true
			token = tk
		}
	}
	assert.True(t, found, "Couldn't find token 'rounded' in sample roundTest")
	assert.Equal(t, "3.14", token.Replacement)
}

func TestLuaLogInfo(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi2.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("logInfoTest")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)
	found := false
	var token config.Token
	for _, tk := range gen.tokens {
		if tk.Name == "logged" {
			found = true
			token = tk
		}
	}
	assert.True(t, found, "Couldn't find token 'logged' in sample logInfoTest")
	assert.Equal(t, "ok", token.Replacement)
}

func TestRemoveToken(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi2.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("removeTokenTest")
	gen := new(luagen)
	runLuaGen(t, s, gen)
	time.Sleep(100 * time.Millisecond)

	foundKeeper := false
	foundRemover := false
	for _, tk := range gen.tokens {
		if tk.Name == "keeper" {
			foundKeeper = true
		}
		if tk.Name == "remover" {
			foundRemover = true
		}
	}
	assert.True(t, foundKeeper, "Token 'keeper' should still be present")
	assert.False(t, foundRemover, "Token 'remover' should have been removed")
}

func TestSendEvent(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi2.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("sendEventTest")
	gen := new(luagen)
	testLuaGen(t, s, gen, "sent via sendEvent")
}

func testLuaGen(t *testing.T, s *config.Sample, gen *luagen, expected string) {
	oq, err := runLuaGen(t, s, gen)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	var good bool
	good = false
	select {
	case oqi := <-oq:
		assert.Equal(t, expected, oqi.Events[0]["_raw"])
		good = true
	case <-timeout:
		if !good {
			t.Fatalf("Timed out, err: %s", err)
		}
	}
}

func runLuaGen(t *testing.T, s *config.Sample, gen *luagen) (chan *config.OutQueueItem, error) {
	// gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)
	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{UseCache: false, SetCache: false}}
	var err error
	go func() {
		err = gen.Gen(gqi)
	}()
	return oq, err
}

func TestBeginEndTimeExposed(t *testing.T) {
	config.ResetConfig()

	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "generator", "luaapi_time.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("beginEndTime")

	// Set BeginParsed and EndParsed on the sample
	loc, _ := time.LoadLocation("Local")
	s.BeginParsed = time.Date(2001, 10, 20, 11, 0, 0, 0, loc)
	s.EndParsed = time.Date(2001, 10, 20, 13, 0, 0, 0, loc)

	beginEpoch := s.BeginParsed.Unix()
	endEpoch := s.EndParsed.Unix()
	expected := fmt.Sprintf("%d-%d", beginEpoch, endEpoch)

	gen := new(luagen)
	testLuaGen(t, s, gen, expected)
}
