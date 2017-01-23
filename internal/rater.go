package internal

import "time"

// RaterConfig defines how to rate an event or token
type RaterConfig struct {
	Name    string                 `json:"name" yaml:"name"`
	Type    string                 `json:"type" yaml:"type"`
	Script  string                 `json:"script,omitempty" yaml:"script,omitempty"`
	Options map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
	Init    map[string]string      `json:"init,omitempty" yaml:"init,omitempty"`
}

// Rater will rate an event according to RaterConfig
type Rater interface {
	EventRate(s *Sample, now time.Time, count int) float64
	TokenRate(t Token, now time.Time) float64
}
