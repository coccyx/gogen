package internal

// import (
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// var (
// 	gh *GitHub
// 	id string
// 	c  *Config
// 	tc *Config
// )

// func TestLogin(t *testing.T) {
// 	gh = NewGitHub(true)
// 	assert.NotNil(t, gh, "NewGitHub() returned nil")
// }

// func TestPush(t *testing.T) {
// 	os.Setenv("GOGEN_HOME", "..")
// 	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
// 	home := ".."
// 	os.Setenv("GOGEN_FULLCONFIG", filepath.Join(home, "examples", "weblog", "weblog.yml"))
// 	os.Setenv("GOGEN_EXPORT", "1")
// 	c = NewConfig()
// 	_ = gh.Push("test_config", c)

// 	l, _, _ := gh.client.Gists.List("", nil)
// 	inList := false
// 	for _, item := range l {
// 		if *item.Description == "test_config" {
// 			inList = true
// 			id = *item.ID
// 		}
// 	}
// 	if !inList {
// 		t.Fatal("test_config not in Gist list")
// 	}
// }

// func TestValid(t *testing.T) {
// 	g, _, err := gh.client.Gists.Get(id)
// 	assert.NoError(t, err, "Failed getting gist")
// 	// fmt.Printf("%# v", pretty.Formatter(g))
// 	_, err = gh.client.Gists.Delete(id)
// 	assert.NoError(t, err, "Failed deleting gist")
// 	content := []byte(*g.Files["test_config.yml"].Content)
// 	err = ioutil.WriteFile("test_config.yml", content, 444)
// 	assert.NoError(t, err, "Cannot write test_config.yml")

// 	os.Setenv("GOGEN_FULLCONFIG", "test_config.yml")
// 	tc = NewConfig()
// 	_ = os.Remove("test_config.yml")

// 	assert.Equal(t, c.Samples[0].Name, tc.Samples[0].Name)
// }
