package generator

import (
	"bytes"
	"sync"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
)

// Pool for output buffers used in fast path
var fastOutputPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate 64KB buffer - typical batch might be 10-100 events * 500 bytes
		return bytes.NewBuffer(make([]byte, 0, 65536))
	},
}

// fastgen implements the fast path generator using pre-compiled templates
type fastgen struct{}

// Gen generates events using pre-compiled FastTemplates
// This bypasses all map operations for maximum performance
func (fg fastgen) Gen(item *config.GenQueueItem) error {
	s := item.S

	if !s.UseFastPath || len(s.FastTemplates) == 0 {
		// Fall back to traditional generator
		log.Debugf("FastPath not available for sample '%s', falling back to traditional", s.Name)
		traditional := sample{}
		return traditional.Gen(item)
	}

	if item.Count == -1 {
		item.Count = len(s.Lines)
	}

	// Get buffer from pool
	buf := fastOutputPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Estimate size and pre-grow buffer
	avgTemplateSize := 0
	for _, ft := range s.FastTemplates {
		avgTemplateSize += ft.TotalStatic + 256 // static + estimated token output
	}
	avgTemplateSize /= len(s.FastTemplates)
	buf.Grow(avgTemplateSize * item.Count)

	slen := len(s.FastTemplates)
	eventsGenerated := 0

	if s.RandomizeEvents {
		// Random event selection
		for i := 0; i < item.Count; i++ {
			if i > 0 {
				buf.WriteByte('\n')
			}
			idx := item.Rand.Intn(slen)
			if err := s.FastTemplates[idx].Execute(buf, item.Earliest, item.Latest, item.Now, item.Rand); err != nil {
				log.Errorf("FastPath error for sample '%s': %s", s.Name, err)
				fastOutputPool.Put(buf)
				// Fall back to traditional
				traditional := sample{}
				return traditional.Gen(item)
			}
			eventsGenerated++
		}
	} else {
		// Sequential event generation
		eventIdx := 0
		for i := 0; i < item.Count; i++ {
			if i > 0 {
				buf.WriteByte('\n')
			}
			if err := s.FastTemplates[eventIdx].Execute(buf, item.Earliest, item.Latest, item.Now, item.Rand); err != nil {
				log.Errorf("FastPath error for sample '%s': %s", s.Name, err)
				fastOutputPool.Put(buf)
				// Fall back to traditional
				traditional := sample{}
				return traditional.Gen(item)
			}
			eventsGenerated++
			eventIdx = (eventIdx + 1) % slen
		}
	}

	// Send the pre-formatted output
	sendFastItem(item, buf.Bytes(), eventsGenerated)

	// Return buffer to pool
	fastOutputPool.Put(buf)

	return nil
}

// sendFastItem sends a pre-formatted output batch to the output queue
func sendFastItem(item *config.GenQueueItem, output []byte, eventCount int) {
	// Make a copy of the bytes since we're returning the buffer to the pool
	outputCopy := make([]byte, len(output))
	copy(outputCopy, output)

	outitem := &config.OutQueueItem{
		S:          item.S,
		Events:     nil, // Not used in fast path
		Cache:      item.Cache,
		FastOutput: outputCopy,
		EventCount: eventCount,
	}

	if item.Cache.SetCache {
		item.Cache.Lock()
		cache[item.S.Name] = nil // Fast path doesn't use event cache
		item.Cache.Unlock()
	}

	item.OQ <- outitem
}

// CanUseFastPath checks if a sample can use the fast path generator
func CanUseFastPath(s *config.Sample) bool {
	return s.UseFastPath && len(s.FastTemplates) > 0
}
