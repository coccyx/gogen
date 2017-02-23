package rater

import (
	"math"
	"reflect"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

// EventRate takes a given sample and current count and returns the rated count
func EventRate(s *config.Sample, now time.Time, count int) (ret int) {
	if s.Rater == nil {
		s.Rater = GetRater(s.RaterString)
		log.Infof("Setting rater to %s, type %s, for sample '%s'", s.RaterString, reflect.TypeOf(s.Rater), s.Name)
	}
	rate := s.Rater.EventRate(s, now, count)
	ratedCount := rate * float64(count)
	if ratedCount < 0 {
		ret = int(math.Ceil(ratedCount - 0.5))
	} else {
		ret = int(math.Floor(ratedCount + 0.5))
	}
	return ret
}

// GetRater returns a rater interface
func GetRater(name string) (ret config.Rater) {
	c := config.NewConfig()
	r := c.FindRater(name)
	if r == nil {
		r := c.FindRater("default")
		ret = &DefaultRater{c: r}
	} else if r.Name == "default" {
		r := c.FindRater("default")
		ret = &DefaultRater{c: r}
	} else if r.Type == "config" {
		ret = &ConfigRater{c: r}
	} else if r.Type == "kbps" {
		ret = &KBpsRater{c: r}
	} else {
		ret = &ScriptRater{c: r}
	}
	return ret
}
