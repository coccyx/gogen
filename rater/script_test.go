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
