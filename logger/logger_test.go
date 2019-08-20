package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogLevel(t *testing.T) {
	if DefaultLogLevel != logrus.ErrorLevel {
		t.Fatalf("Log not set to Error")
	}
}
