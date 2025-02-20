package tests

import (
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestReplay(t *testing.T) {
	// Setup environment
	config.ResetConfig()
	config.SetupFromString(`
global:
  output:
    outputter: buf
samples:
  - name: fullreplay
    generator: replay
    begin: "2001-10-20 12:00:00"
    end: "2001-10-20 12:00:49"
    tokens:
    - name: ts1
      type: timestamp
      replacement: "%Y-%m-%dT%H:%M:%S"
      format: regex
      token: "(\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2})"
    lines:
    - "_raw": "2001-10-20T12:00:00"
    - "_raw": "2001-10-20T12:00:01"
    - "_raw": "2001-10-20T12:00:06"
    - "_raw": "2001-10-20T12:00:16"
    - "_raw": "2001-10-20T12:00:36"
`)

	c := config.NewConfig()
	run.Run(c)

	assert.Equal(t, `2001-10-20T12:00:00
2001-10-20T12:00:01
2001-10-20T12:00:06
2001-10-20T12:00:16
2001-10-20T12:00:36
`, c.Buf.String())
	config.CleanupConfigAndEnvironment()
}
