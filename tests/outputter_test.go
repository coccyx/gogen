package tests

import (
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
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
	outputter.Mutex.Lock()
	outputter.BytesWritten = make(map[string]int64)
	outputter.EventsWritten = make(map[string]int64)
	// Note: We can't directly access unexported variables like rotwg and rotchan
	// from outside the package, so we'll rely on the outputter package's own
	// initialization and cleanup mechanisms.
	outputter.Mutex.Unlock()

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
	go outputter.ROT(dummyConfig)
	// Give the ROT goroutine a moment to start up and initialize rotchan
	// Adjust duration if needed, but keep it short for test speed.
	time.Sleep(10 * time.Millisecond)

	// --- Action ---
	testSampleName := "test_sync_sample"
	expectedBytes := int64(123)
	expectedEvents := int64(1)

	// Send a single stat message using the thread-safe Account function
	outputter.Account(expectedEvents, expectedBytes, testSampleName)

	// --- Trigger Potential Race ---
	// Immediately call ReadFinal.
	// OLD Code: Might return before readStats processes the stat.
	// NEW Code: Will block until readStats processes the stat (due to rotwg.Wait).
	outputter.ReadFinal() // This also closes rotchan in the new code

	// --- Verification ---
	// Access global counters directly after ReadFinal has returned.
	// We need the lock to safely read the maps.
	outputter.Mutex.RLock()
	finalBytes := outputter.BytesWritten[testSampleName]
	finalEvents := outputter.EventsWritten[testSampleName]
	outputter.Mutex.RUnlock()

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
