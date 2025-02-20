package tests

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestFileOutput(t *testing.T) {
	config.SetupFromFile(filepath.Join("..", "tests", "fileoutput", "fileoutput.yml"))
	c := config.NewConfig()
	// s := c.FindSampleByName("backfill")
	run.Run(c)

	info, err := os.Stat(c.Global.Output.FileName)
	assert.NoError(t, err)
	assert.Condition(t, func() bool {
		return info.Size() < c.Global.Output.MaxBytes
	}, "Rotation failing, main file size of %d greater than MaxBytes %d", info.Size(), c.Global.Output.MaxBytes)
	for i := 1; i <= c.Global.Output.BackupFiles; i++ {
		info, err = os.Stat(c.Global.Output.FileName + "." + strconv.Itoa(i))
		assert.NoError(t, err)
		assert.Condition(t, func() bool {
			return info.Size() > c.Global.Output.MaxBytes
		}, "Rotation failing, file %d less than MaxBytes", i)
	}

}
