package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSharePull(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	_ = os.Mkdir("testout", 0777)
	Pull("coccyx/weblog", "testout", false)
	_, err := os.Stat("testout/weblog.yml")
	assert.NoError(t, err, "Couldn't find file weblog.yml")
	_ = os.Remove("testout/weblog.yml")

	Pull("coccyx/weblog", "testout", true)
	_, err = os.Stat("testout/samples/weblog.yml")
	assert.NoError(t, err, "Couldn't find file samples/weblog.yml")
	_, err = os.Stat("testout/samples/webhosts.csv")
	assert.NoError(t, err, "Couldn't find file samples/webhosts.csv")
	_, err = os.Stat("testout/samples/useragents.sample")
	assert.NoError(t, err, "Couldn't find file samples/useragents.sample")
	_ = os.RemoveAll("testout")

}

func TestSharePullFile(t *testing.T) {
	os.Setenv("GOGEN_TMPDIR", "..")
	os.Setenv("GOGEN_HOME", "..")
	os.Remove("../.versioncachefile_coccyx%2Fweblog")
	os.Remove("../.configcache_coccyx%2Fweblog")
	PullFile("coccyx/weblog", ".test.json")
	_, err := os.Stat(".test.json")
	assert.NoError(t, err, "Couldn't fine .test.json")
	_, err = os.Stat(filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".versioncache_coccyx%2Fweblog"))
	assert.NoError(t, err, "Couldn't fine version cache file")
	_, err = os.Stat(filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".configcache_coccyx%2Fweblog"))
	assert.NoError(t, err, "Couldn't find cache file")
	os.Remove(".test.json")
}
