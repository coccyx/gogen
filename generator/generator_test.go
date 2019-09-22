package generator

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/tests"
	"github.com/stretchr/testify/assert"
)

func TestGenerator(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := filepath.Join("..", "tests", "tokens")
	os.Setenv("GOGEN_SAMPLES_DIR", home)
	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	// gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)
	s := tests.FindSampleInFile(home, "token-static")
	if s == nil {
		t.Fatalf("Sample token-static not found in file: %s", home)
	}
	gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{UseCache: false, SetCache: false}}
	gq := make(chan *config.GenQueueItem)
	gqs := make(chan int)
	go Start(gq, gqs)
	gq <- gqi
	close(gq)
	oqi := <-oq
	assert.Equal(t, "foo", oqi.Events[0]["_raw"])
}

func TestGeneratorCache(t *testing.T) {
	// Setup environment
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	home := filepath.Join("..", "tests", "tokens")
	os.Setenv("GOGEN_SAMPLES_DIR", home)
	loc, _ := time.LoadLocation("Local")
	source := rand.NewSource(0)
	randgen := rand.New(source)

	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time {
		return n
	}

	// gq := make(chan *config.GenQueueItem)
	oq := make(chan *config.OutQueueItem)
	s := tests.FindSampleInFile(home, "token-static")
	if s == nil {
		t.Fatalf("Sample token-static not found in file: %s", home)
	}
	gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{UseCache: false, SetCache: true}}
	gq := make(chan *config.GenQueueItem)
	gqs := make(chan int)
	go Start(gq, gqs)
	gq <- gqi
	oqi := <-oq
	assert.Equal(t, "foo", oqi.Events[0]["_raw"])

	// Change token replacement, validate it's different without cache
	s.Tokens[0].Replacement = "foo2"
	gqi = &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{UseCache: false, SetCache: false}}
	gq <- gqi
	oqi = <-oq
	assert.Equal(t, "foo2", oqi.Events[0]["_raw"])

	// Now use cache, should be same as the old
	gqi = &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{UseCache: true, SetCache: false}}
	gq <- gqi
	close(gq)
	oqi = <-oq
	assert.Equal(t, "foo", oqi.Events[0]["_raw"])
}
