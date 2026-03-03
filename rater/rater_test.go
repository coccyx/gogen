package rater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
	"github.com/stretchr/testify/assert"
)

func TestRandomizeCount(t *testing.T) {
	s := &config.Sample{RandomizeCount: 0.2}
	randSource = 2
	count := EventRate(s, time.Now(), 10)
	assert.Equal(t, 11, count)
}

func TestTokenRateDefault(t *testing.T) {
	dr := &DefaultRater{}
	token := config.Token{Name: "test"}
	rate := dr.TokenRate(token, time.Now())
	assert.Equal(t, 1.0, rate)
}

func TestTokenRateConfig(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "rater", "configrater.yml"))

	c := config.NewConfig()
	r := c.FindRater("testconfigrater")

	cr := &ConfigRater{c: r}
	token := config.Token{Name: "test"}

	loc, _ := time.LoadLocation("Local")
	n := time.Date(2001, 10, 20, 0, 0, 0, 100000, loc)
	rate := cr.TokenRate(token, n)
	assert.Equal(t, 2.0, rate)
}

func TestTokenRateKBps(t *testing.T) {
	kr := &KBpsRater{c: &config.RaterConfig{Name: "kbps"}}
	token := config.Token{Name: "test"}
	rate := kr.TokenRate(token, time.Now())
	assert.Equal(t, 1.0, rate)
}

func TestTokenRateScript(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "rater", "luarater.yml"))

	c := config.NewConfig()
	r := c.FindRater("multiply")

	sr := &ScriptRater{c: r}
	token := config.Token{Name: "test"}
	rate := sr.TokenRate(token, time.Now())
	assert.Equal(t, 2.0, rate)
}

func TestGetRaterFallback(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join("..", "tests", "rater", "defaultrater.yml"))

	r := GetRater("nonexistentrater")
	assert.IsType(t, &DefaultRater{}, r, "unknown rater name should fall back to DefaultRater")
}

func TestEventRateNegativeResult(t *testing.T) {
	// A rater returning a negative rate should produce a negative or zero count
	s := &config.Sample{RandomizeCount: 0}
	// Use a mock rater by pre-setting it
	s.Rater = &negativeRater{}
	randSource = 2
	count := EventRate(s, time.Now(), 10)
	assert.True(t, count <= 0, "negative rate should produce non-positive count, got %d", count)
}

func TestKBpsEventRateMissingOption(t *testing.T) {
	kr := &KBpsRater{
		c: &config.RaterConfig{
			Name:    "kbps",
			Options: map[string]interface{}{},
		},
	}
	s := &config.Sample{Name: "test"}
	rate := kr.EventRate(s, time.Now(), 10)
	assert.Equal(t, 1.0, rate)
}

func TestKBpsEventRateWrongType(t *testing.T) {
	kr := &KBpsRater{
		c: &config.RaterConfig{
			Name: "kbps",
			Options: map[string]interface{}{
				"KBps": "not_a_float",
			},
		},
	}
	s := &config.Sample{Name: "test"}
	rate := kr.EventRate(s, time.Now(), 10)
	assert.Equal(t, 1.0, rate)
}

func TestKBpsEventRateMissingSample(t *testing.T) {
	kr := &KBpsRater{
		c: &config.RaterConfig{
			Name: "kbps",
			Options: map[string]interface{}{
				"KBps": 100.0,
			},
		},
	}
	s := &config.Sample{Name: "nonexistent_sample_kbps"}
	rate := kr.EventRate(s, time.Now(), 10)
	assert.Equal(t, 1.0, rate)
}

func TestKBpsEventRateWithData(t *testing.T) {
	// Pre-populate outputter stats
	outputter.Mutex.Lock()
	outputter.BytesWritten["kbps_test_sample"] = 10000
	outputter.EventsWritten["kbps_test_sample"] = 100
	outputter.Mutex.Unlock()
	defer func() {
		outputter.Mutex.Lock()
		delete(outputter.BytesWritten, "kbps_test_sample")
		delete(outputter.EventsWritten, "kbps_test_sample")
		outputter.Mutex.Unlock()
	}()

	kr := &KBpsRater{
		c: &config.RaterConfig{
			Name: "kbps",
			Options: map[string]interface{}{
				"KBps": 100.0,
			},
		},
		t: time.Now().Add(-1 * time.Second), // pretend we started 1s ago
	}
	s := &config.Sample{Name: "kbps_test_sample"}
	rate := kr.EventRate(s, time.Now(), 10)
	assert.Equal(t, 1.0, rate) // always returns 1.0 regardless
}

// negativeRater always returns a negative rate for testing
type negativeRater struct{}

func (nr *negativeRater) EventRate(s *config.Sample, now time.Time, count int) float64 {
	return -1.0
}
func (nr *negativeRater) TokenRate(t config.Token, now time.Time) float64 {
	return -1.0
}
