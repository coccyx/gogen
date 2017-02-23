package rater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestKBpsRaterEventRate(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "kbpsrater.yml"))

	c := config.NewConfig()
	r := c.FindRater("kbpsrater")
	assert.Equal(t, "kbpsrater", r.Name)
	s := c.FindSampleByName("kbpsrater")
	assert.Equal(t, "kbpsrater", s.RaterString)

	loc, _ := time.LoadLocation("Local")
	n := time.Date(2001, 10, 20, 0, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}
	ret := EventRate(s, now(), 1)
	assert.Equal(t, 1.0, ret)
}

