package rater

import (
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/stretchr/testify/assert"
)

func TestRandomizeCount(t *testing.T) {
	s := &config.Sample{RandomizeCount: 0.2}
	randSource = 2
	count := EventRate(s, time.Now(), 10)
	assert.Equal(t, 11, count)
}
