package rater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)


func TestConfigRaterEventRate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "configrater.yml"))

	c := config.NewConfig()
	r := c.FindRater("testconfigrater")
	assert.Equal(t, "testconfigrater", r.Name)
	s := c.FindSampleByName("configrater")
	assert.Equal(t, "testconfigrater", s.RaterString)

	loc, _ := time.LoadLocation("Local")
	n := time.Date(2001, 10, 20, 0, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}
	ret := EventRate(s, now(), 1)
	assert.IsType(t, &ConfigRater{}, s.Rater)
	assert.True(t, assert.ObjectsAreEqual(r, s.Rater.(*ConfigRater).c))
	assert.Equal(t, 2.0, ret)

	n = time.Date(2001, 10, 20, 1, 0, 0, 100000, loc)
	ret = EventRate(s, now(), 1)
	assert.Equal(t, 3.0, ret)

	n = time.Date(2001, 10, 22, 2, 0, 0, 100000, loc)
	ret = EventRate(s, now(), 1)
	assert.Equal(t, 4.0, ret)

	n = time.Date(2001, 10, 20, 2, 2, 0, 100000, loc)
	ret = EventRate(s, now(), 1)
	assert.Equal(t, 5.0, ret)
}
