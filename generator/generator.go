package generator

import (
	"math/rand"
	"sync"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

var (
	cache      map[string][]map[string]string
	cacheMutex *sync.RWMutex
)

func Start(gq chan *config.GenQueueItem, gqs chan int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	gens := make(map[string]config.Generator)
	cache = make(map[string][]map[string]string)
	cacheMutex = &sync.RWMutex{}
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	for {
		item, ok := <-gq
		if !ok {
			gqs <- 1
			break
		}
		item.Rand = generator
		// Check to see if our generator is not set
		if gens[item.S.Name] == nil {
			log.Infof("Setting sample '%s' to generator '%s'", item.S.Name, item.S.Generator)
			if item.S.Generator == "sample" || item.S.Generator == "replay" {
				s := new(sample)
				gens[item.S.Name] = s
			} else {
				s := new(luagen)
				gens[item.S.Name] = s
			}
			PrimeRater(item.S)
		}
		useCache := false
		var cachedEvents []map[string]string
		if item.Cache.UseCache {
			cacheMutex.RLock()
			cachedEvents, useCache = cache[item.S.Name]
			cacheMutex.RUnlock()
		}
		if useCache {
			sendItem(item, cachedEvents)
		} else {
			// log.Debugf("Generating item %#v", item)
			err := gens[item.S.Name].Gen(item)
			if err != nil {
				log.Errorf("Error received from generator: %s", err)
			}
		}
		// log.Debugf("Finished generating item %#v", item)
	}
}

func sendItem(item *config.GenQueueItem, events []map[string]string) {
	outitem := &config.OutQueueItem{S: item.S, Events: events, Cache: item.Cache}
	if item.Cache.SetCache {
		item.Cache.Lock()
		cache[item.S.Name] = events
		item.Cache.Unlock()
	}
	item.OQ <- outitem
}
