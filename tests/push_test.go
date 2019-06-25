package tests

// import (
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	config "github.com/coccyx/gogen/internal"
// 	"github.com/coccyx/gogen/run"
// )

// func TestSharePush(t *testing.T) {
// 	config.ResetConfig()
// 	os.Setenv("GOGEN_HOME", "..")
// 	os.Setenv("GOGEN_ALWAYS_REFRESH", "")
// 	home := ".."
// 	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "examples", "weblog", "weblog.yml"))
// 	os.Setenv("GOGEN_EXPORT", "1")
// 	_ = config.NewConfig()
// 	var r run.Runner
// 	config.Push("weblog", r)
// 	os.Setenv("GOGEN_EXPORT", "")
// }
