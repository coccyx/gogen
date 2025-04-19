package outputter

import (
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	// TODO: Replace with the actual import path for your config package
)

// Example test case for how to reproduce ROT race condition and exiting before all the stats are collected:
// gogen -v -c examples/tutorial/tutorial1.yml --os 4 -o file --filename /dev/null  gen -c 1 -i 1 -ei 1
// This case gogen will exit before all the stats are collected.

// TestReadFinalSynchronization checks if ReadFinal correctly waits for all stats
// to be processed before returning. This specifically tests the race condition
// fixed by the introduction of the sync.WaitGroup.
func TestReadFinalSynchronization(t *testing.T) {
	// --- Test Setup ---
	// Reset package-level state variables before each test run.
	// This is crucial because tests might run concurrently or affect each other.
	Mutex.Lock()
	BytesWritten = make(map[string]int64)
	EventsWritten = make(map[string]int64)
	// Explicitly reset the WaitGroup to ensure a clean state for the test.
	// This prevents issues if ROT doesn't reset it or if state persists.
	rotwg = sync.WaitGroup{}
	// Ensure rotchan is nil so ROT creates a new one, avoiding issues with closed channels.
	// Note: This assumes ROT checks for nil or always makes a new channel. If ROT
	// has complex logic around existing channels, this might need adjustment.
	// Closing the old channel here could panic if already closed. Setting to nil is safer.
	rotchan = nil
	Mutex.Unlock()

	// Create a minimal configuration required by ROT
	// Adjust this based on the actual fields your ROT function needs.
	dummyConfig := &config.Config{
		Global: config.Global{
			ROTInterval: 1, // Value likely doesn't matter for this specific test
		},
		// Add other necessary dummy fields for config if ROT accesses them
	}

	// Initialize the outputter system (starts readStats goroutine)
	// Run ROT in a goroutine as it contains an infinite loop for periodic stats
	go ROT(dummyConfig)
	// Give the ROT goroutine a moment to start up and initialize rotchan
	// Adjust duration if needed, but keep it short for test speed.
	time.Sleep(10 * time.Millisecond)

	// --- Action ---
	testSampleName := "test_sync_sample"
	expectedBytes := int64(123)
	expectedEvents := int64(1)

	statToSend := &config.OutputStats{
		SampleName:    testSampleName,
		BytesWritten:  expectedBytes,
		EventsWritten: expectedEvents,
	}

	// Send a single stat message
	rotchan <- statToSend

	// --- Trigger Potential Race ---
	// Immediately call ReadFinal.
	// OLD Code: Might return before readStats processes the stat.
	// NEW Code: Will block until readStats processes the stat (due to rotwg.Wait).
	ReadFinal() // This also closes rotchan in the new code

	// --- Verification ---
	// Access global counters directly after ReadFinal has returned.
	// We need the lock to safely read the maps.
	Mutex.RLock()
	finalBytes := BytesWritten[testSampleName]
	finalEvents := EventsWritten[testSampleName]
	Mutex.RUnlock()

	// --- Assertions ---
	if finalBytes != expectedBytes {
		t.Errorf("Synchronization error: Expected final bytes %d, but got %d. ReadFinal might have returned too early.", expectedBytes, finalBytes)
	}
	if finalEvents != expectedEvents {
		t.Errorf("Synchronization error: Expected final events %d, but got %d. ReadFinal might have returned too early.", expectedEvents, finalEvents)
	}

	// Note: In the OLD code, this test might still pass intermittently if the
	// scheduler runs readStats very quickly. However, it's designed to reliably
	// FAIL if the race condition occurs (ReadFinal returning before update).
	// In the NEW code (with WaitGroup), this test should reliably PASS.
}
