package rater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRaterEventRate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "defaultrater.yml"))

	c := config.NewConfig()
	r := c.FindRater("default")
	assert.Equal(t, "default", r.Name)
	s := c.FindSampleByName("defaultrater")
	assert.Equal(t, "default", s.RaterString)
	ret := EventRate(s, time.Now(), 1)
	assert.IsType(t, &DefaultRater{}, s.Rater)
	assert.True(t, assert.ObjectsAreEqual(r, s.Rater.(*DefaultRater).c))
	assert.Equal(t, 1, ret)
}
