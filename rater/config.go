package rater

import (
	"time"

	config "github.com/coccyx/gogen/internal"
)

// ConfigRater simply returns the passed count
type ConfigRater struct {
	c *config.RaterConfig
}

// getRate acts as a general method for EventRate and TokenRate
func (cr *ConfigRater) getRate(now time.Time) float64 {
	rate := 1.0

	if _, ok := cr.c.Options["HourOfDay"]; ok {
		hod := cr.c.Options["HourOfDay"].(map[int]float64)
		if _, ok = hod[now.Hour()]; ok {
			rate *= hod[now.Hour()]
		}
	}
	if _, ok := cr.c.Options["DayOfWeek"]; ok {
		dow := cr.c.Options["DayOfWeek"].(map[int]float64)
		if _, ok := dow[int(now.Weekday())]; ok {
			rate *= dow[int(now.Weekday())]
		}
	}
	if _, ok := cr.c.Options["MinuteOfHour"]; ok {
		moh := cr.c.Options["MinuteOfHour"].(map[int]float64)
		if _, ok := moh[now.Minute()]; ok {
			rate *= moh[now.Minute()]
		}
	}

	return rate
}

// EventRate takes a given sample and current count and returns the rated count
func (cr *ConfigRater) EventRate(s *config.Sample, now time.Time, count int) float64 {
	return cr.getRate(now)
}

// TokenRate takes a token and returns the rated value
func (cr *ConfigRater) TokenRate(t config.Token, now time.Time) float64 {
	return cr.getRate(now)
}
