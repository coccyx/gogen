package rater

import (
	"math"
	"math/rand"
	"reflect"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

var randGen *rand.Rand
var randSource int64

// EventRate takes a given sample and current count and returns the rated count
func EventRate(s *config.Sample, now time.Time, count int) (ret int) {
	if s.Rater == nil {
		s.Rater = GetRater(s.RaterString)
		log.Infof("Setting rater to %s, type %s, for sample '%s'", s.RaterString, reflect.TypeOf(s.Rater), s.Name)
	}
	rate := s.Rater.EventRate(s, now, count)
	randFactor := float64(1.0)
	if s.RandomizeCount != float64(0) {
		randBound := int(math.Round(s.RandomizeCount * 1000))
		rand := randGen.Intn(randBound)
		randFactor = 1 + (-(float64(randBound/2) - float64(rand)) / float64(1000))
		rate *= randFactor
	}
	ratedCount := rate * float64(count)
	if ratedCount < 0 {
		ret = int(math.Ceil(ratedCount - 0.5))
	} else {
		ret = int(math.Floor(ratedCount + 0.5))
	}
	log.Debugf("count: %d ratedCount: %.2f origCount: %d randFactor %.2f for sample '%s'", ret, ratedCount, count, randFactor, s.Name)
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

	if randSource == 0 { // Allow tests to override
		randSource = time.Now().UnixNano()
	}
	randGen = rand.New(rand.NewSource(randSource))
	return ret
}
