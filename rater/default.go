package rater

import (
	"time"

	config "github.com/coccyx/gogen/internal"
)

// DefaultRater simply returns the passed count
type DefaultRater struct {
	c *config.RaterConfig
}

// EventRate takes a given sample and current count and returns the rated count
func (dr *DefaultRater) EventRate(s *config.Sample, now time.Time, count int) float64 {
	return 1.0
}

// TokenRate takes a token and returns the rated value
func (dr *DefaultRater) TokenRate(t config.Token, now time.Time) float64 {
	return 1.0
}
