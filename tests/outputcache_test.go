package tests

import (
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestOutputCache(t *testing.T) {
	configStr := `
global:
  output:
    outputter: buf
  cacheIntervals: 2
samples:
  - name: outputcache
    begin: "2001-10-20 12:00:00"
    end: "2001-10-20 12:00:04"
    interval: 1
    count: 1
    tokens:
    - name: ts1
      type: timestamp
      replacement: "%Y-%m-%dT%H:%M:%S"
      token: $ts1$
      format: template
    lines:
    - "_raw": "$ts1$"
`

	config.SetupFromString(configStr)

	c := config.NewConfig()
	run.Run(c)

	assert.Equal(t, `2001-10-20T12:00:00
2001-10-20T12:00:00
2001-10-20T12:00:00
2001-10-20T12:00:03
`, c.Buf.String())

	config.CleanupConfigAndEnvironment()
}
