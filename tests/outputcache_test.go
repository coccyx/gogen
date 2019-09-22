package tests

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestOutputCache(t *testing.T) {
	// Setup environment
	config.ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
	home := ".."
	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "tests", "outputcache", "outputcache.yml"))

	c := config.NewConfig()
	run.Run(c)

	assert.Equal(t, `2001-10-20T12:00:00
2001-10-20T12:00:00
2001-10-20T12:00:00
2001-10-20T12:00:03
`, c.Buf.String())
}
