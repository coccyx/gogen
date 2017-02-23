package rater

import (
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

// KBpsRater rates on KB/s
type KBpsRater struct {
	c *config.RaterConfig
	t time.Time
}

// EventRate takes a given sample and current count and returns the rated count
func (r *KBpsRater) EventRate(s *config.Sample, now time.Time, count int) float64 {

	if _, ok := r.c.Options["KBps"]; !ok {
		log.Errorf("KBpsRater: KBps must be present")
		return 1.0
	}

	KBps, ok := r.c.Options["KBps"].(float64)
	if !ok {
		log.Errorf("KBpsRater: KBps must be float64")
		return 1.0
	}

	var size int
	for _, line := range(s.Lines) {
		if raw, ok := line["_raw"]; ok {
			size = len(raw)
			break
		}
	}

	if size == 0 {
		log.Errorf("KBpsRater: only works for _raw events")
		return 1.0
	}

	expected := float64(size * count) / 1024.0 / KBps
	current := time.Now()
	actual := current.Sub(r.t).Seconds()
	r.t = current

	delta := expected - actual

	if delta > 0 {
		log.Debugf("KBpsRater: expected=%.2f actual=%.2f delta=%.2f",
			expected, actual, delta)

		time.Sleep(time.Duration(delta) * time.Second)
	}

	return 1.0
}

// TokenRate takes a token and returns the rated value
func (r *KBpsRater) TokenRate(t config.Token, now time.Time) float64 {
	return 1.0
}
