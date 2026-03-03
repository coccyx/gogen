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

// setupGenTest resets config, sets env vars, and returns common test fixtures.
func setupGenTest(t *testing.T, samplesDir string, seed int64) (func() time.Time, *rand.Rand) {
	t.Helper()
	config.ResetConfig()
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", "")
	os.Setenv("GOGEN_SAMPLES_DIR", samplesDir)
	loc, _ := time.LoadLocation("Local")
	randgen := rand.New(rand.NewSource(seed))
	n := time.Date(2001, 10, 20, 12, 0, 0, 100000, loc)
	now := func() time.Time { return n }
	return now, randgen
}

func TestGenerator(t *testing.T) {
	home := filepath.Join("..", "tests", "tokens")
	now, randgen := setupGenTest(t, home, 0)

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

func TestGeneratorMultiPass(t *testing.T) {
	home := filepath.Join("..", "tests", "tokens")
	now, randgen := setupGenTest(t, home, 0)

	oq := make(chan *config.OutQueueItem)
	s := tests.FindSampleInFile(home, "tokens")
	if s == nil {
		t.Fatalf("Sample tokens not found")
	}
	// Force MultiPass
	s.SinglePass = false

	// Count > lines: tests the iters > 1 path in genMultiPass
	gqi := &config.GenQueueItem{Count: len(s.Lines) + 2, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{}}
	go func() {
		err := genMultiPass(gqi)
		assert.NoError(t, err)
	}()

	oqi := <-oq
	assert.Equal(t, len(s.Lines)+2, len(oqi.Events))
}

func TestGeneratorMultiPassRandomize(t *testing.T) {
	home := filepath.Join("..", "tests", "tokens")
	now, randgen := setupGenTest(t, home, 42)

	oq := make(chan *config.OutQueueItem)
	s := tests.FindSampleInFile(home, "tokens")
	if s == nil {
		t.Fatalf("Sample tokens not found")
	}
	s.SinglePass = false
	s.RandomizeEvents = true

	gqi := &config.GenQueueItem{Count: 5, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{}}
	go func() {
		genMultiPass(gqi)
	}()

	oqi := <-oq
	assert.Equal(t, 5, len(oqi.Events))
}

func TestGeneratorSinglePassCountGtLines(t *testing.T) {
	home := filepath.Join("..", "tests", "singlepass")
	now, randgen := setupGenTest(t, filepath.Join(home, "test1.yml"), 0)

	c := config.NewConfig()
	s := c.FindSampleByName("test1")
	if s == nil {
		t.Fatalf("Sample test1 not found")
	}
	assert.True(t, s.SinglePass)

	oq := make(chan *config.OutQueueItem)
	// Count > lines: tests the iters > 1 singlepass path
	gqi := &config.GenQueueItem{Count: len(s.Lines) + 3, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{}}
	go func() {
		genSinglePass(gqi)
	}()

	oqi := <-oq
	assert.Equal(t, len(s.Lines)+3, len(oqi.Events))
}

func TestGeneratorSinglePassRandomize(t *testing.T) {
	home := filepath.Join("..", "tests", "singlepass")
	now, randgen := setupGenTest(t, filepath.Join(home, "test1.yml"), 42)

	c := config.NewConfig()
	s := c.FindSampleByName("test1")
	if s == nil {
		t.Fatalf("Sample test1 not found")
	}
	assert.True(t, s.SinglePass)
	s.RandomizeEvents = true

	oq := make(chan *config.OutQueueItem)
	gqi := &config.GenQueueItem{Count: 5, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{}}
	go func() {
		genSinglePass(gqi)
	}()

	oqi := <-oq
	assert.Equal(t, 5, len(oqi.Events))
}

func TestGeneratorStartWorker(t *testing.T) {
	home := filepath.Join("..", "tests", "tokens")
	now, randgen := setupGenTest(t, home, 0)

	oq := make(chan *config.OutQueueItem)
	s := tests.FindSampleInFile(home, "token-static")
	if s == nil {
		t.Fatalf("Sample token-static not found")
	}

	gq := make(chan *config.GenQueueItem)
	gqs := make(chan int)
	go Start(gq, gqs)

	// Send multiple items to test the "generator already set" path
	for i := 0; i < 3; i++ {
		gqi := &config.GenQueueItem{Count: 1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{}}
		gq <- gqi
		oqi := <-oq
		assert.Equal(t, "foo", oqi.Events[0]["_raw"])
	}

	close(gq)
	select {
	case <-gqs:
	case <-time.After(5 * time.Second):
		t.Fatal("Generator worker did not finish in time")
	}
}

func TestGeneratorCountMinusOne(t *testing.T) {
	home := filepath.Join("..", "tests", "tokens")
	now, randgen := setupGenTest(t, home, 0)

	oq := make(chan *config.OutQueueItem)
	s := tests.FindSampleInFile(home, "tokens")
	if s == nil {
		t.Fatalf("Sample tokens not found")
	}
	// Count=-1 means "use all lines"
	gqi := &config.GenQueueItem{Count: -1, Earliest: now(), Latest: now(), Now: now(), S: s, OQ: oq, Rand: randgen, Cache: &config.CacheItem{}}
	go func() {
		sg := sample{}
		sg.Gen(gqi)
	}()

	oqi := <-oq
	assert.Equal(t, len(s.Lines), len(oqi.Events))
}

func TestPrimeRaterSetsRater(t *testing.T) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")

	s := &config.Sample{
		Name: "primerater_test",
		Tokens: []config.Token{
			{
				Name:        "ratedtoken",
				Type:        "rated",
				RaterString: "default",
			},
			{
				Name: "normaltoken",
				Type: "choice",
			},
		},
	}

	PrimeRater(s)
	assert.NotNil(t, s.Tokens[0].Rater, "rated token should have rater set")
}

func TestGeneratorCache(t *testing.T) {
	home := filepath.Join("..", "tests", "tokens")
	now, randgen := setupGenTest(t, home, 0)

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
