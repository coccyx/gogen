package tests

import (
	"testing"
	"time"

	"github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
)

func TestBerserkModeBypass(t *testing.T) {
	// Test that berserk mode (CacheIntervals = 2147483647) bypasses accounting
	config := &internal.Config{
		Global: internal.Global{
			CacheIntervals: 2147483647, // berserk mode value (fullRetard flag)
			ROTInterval:    1,
		},
	}

	// Initialize ROT with berserk mode config
	outputter.ROT(config)

	// Try to send accounting data - should not panic or deadlock
	done := make(chan bool)
	go func() {
		outputter.Account(100, 1024, "test-sample")
		done <- true
	}()

	select {
	case <-done:
		// Good - Account returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Account() blocked when it should have bypassed in berserk mode")
	}

	// ReadFinal should also not block
	go func() {
		outputter.ReadFinal()
		done <- true
	}()

	select {
	case <-done:
		// Good - ReadFinal returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ReadFinal() blocked when it should have bypassed in berserk mode")
	}
}

func TestNormalModeAccounting(t *testing.T) {
	// Test that normal mode (CacheIntervals != 2147483647) works normally
	config := &internal.Config{
		Global: internal.Global{
			CacheIntervals: 100, // normal value
			ROTInterval:    1,
		},
	}

	// Initialize ROT in a goroutine since it has an infinite loop
	go outputter.ROT(config)
	
	// Give ROT time to initialize
	time.Sleep(10 * time.Millisecond)

	// Account should work normally
	done := make(chan bool)
	go func() {
		outputter.Account(100, 1024, "test-sample")
		done <- true
	}()

	select {
	case <-done:
		// Good - Account returned
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Account() blocked unexpectedly in normal mode")
	}

	// Clean up
	outputter.ReadFinal()
}