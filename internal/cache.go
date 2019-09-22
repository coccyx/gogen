package internal

import (
	"sync"
)

// CacheItem represents whether to cache a given item, created in the original timer which generates
// a GenQueueItem. It's also sent along with the OutQueueItem.
type CacheItem struct {
	sync.RWMutex
	UseCache bool
	SetCache bool
}
