package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigRater(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	os.Setenv("GOGEN_EXPORT", "")
	home := ".."

	c := NewConfig()

	var rater RaterConfig

	fullPath := filepath.Join(home, "tests", "rater", "defaultconfigrater.yml")
	if err := c.parseFileConfig(&rater, fullPath); err != nil {
		t.Fatal(err, "Couldn't open or parse tests/rater/defaultconfigrater.yml")
	}
	c.validateRater(&rater)
	dr := getRater(c, "config")
	assert.NotNil(t, dr, "config rater does not exist")
	compareConfigRater(&rater, dr, t)
}

func TestFullConfigRater(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "fullraterconfig.yml"))

	c := NewConfig()

	r1 := getRater(c, "config")
	r2 := getRater(c, "testconfigrater")
	compareConfigRater(r1, r2, t)
}

func TestLuaRaterConfig(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "rater", "luarater.yml"))

	c := NewConfig()

	multiply := getRater(c, "multiply")
	assert.Equal(t, `return options["multiplier"]`+"\n", multiply.Script)
}

func getRater(c *Config, name string) *RaterConfig {
	for _, dr := range c.Raters {
		if dr.Name == name {
			return dr
		}
	}
	return nil
}

func compareConfigRater(r1, r2 *RaterConfig, t *testing.T) {
	for _, heading := range []string{"HourOfDay", "DayOfWeek", "MinuteOfHour"} {
		for k := range r1.Options[heading].(map[int]float64) {
			raterCast := r1.Options[heading].(map[int]float64)
			drCast := r2.Options[heading].(map[int]float64)
			if raterCast[k] != drCast[k] {
				t.Fatalf("[%s][%d] does not exist or does not match, default: %f, configrater: %f", heading, k, drCast[k], raterCast[k])
			}
		}
	}
}
