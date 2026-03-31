package rater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)


func TestScriptRaterEventRate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "luarater.yml"))

	c := config.NewConfig()
	r := c.FindRater("multiply")
	assert.Equal(t, "multiply", r.Name)
	s := c.FindSampleByName("double")
	assert.Equal(t, "multiply", s.RaterString)
	ret := EventRate(s, time.Now(), 1)
	assert.IsType(t, &ScriptRater{}, s.Rater)
	assert.True(t, assert.ObjectsAreEqual(r, s.Rater.(*ScriptRater).c))
	assert.Equal(t, 2, ret)
}

func TestScriptRaterNowExposed(t *testing.T) {
	config.ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "luarater_time.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("time_aware")
	assert.Equal(t, "time_rater", s.RaterString)
	// The script returns 3.0 if now > 0, else 1.0
	ret := EventRate(s, time.Now(), 1)
	assert.Equal(t, 3, ret)
}

func TestScriptRaterBeginTimeExposed(t *testing.T) {
	config.ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "luarater_time.yml"))

	c := config.NewConfig()
	s := c.FindSampleByName("time_aware")
	// Set BeginParsed so beginTime is exposed
	s.BeginParsed = time.Date(2001, 10, 20, 12, 0, 0, 0, time.UTC)

	// Override rater with one that checks beginTime
	r := &ScriptRater{
		c: &config.RaterConfig{
			Name:   "begin_check",
			Type:   "script",
			Script: "if beginTime ~= nil and beginTime > 0 then return 5.0 else return 1.0 end",
		},
	}
	s.Rater = r
	ret := r.EventRate(s, time.Now(), 1)
	assert.Equal(t, float64(5.0), ret)
}
