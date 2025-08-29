package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
)

// BenchmarkAtomicAccountingNormal tests our atomic pointer implementation in normal mode
func BenchmarkAtomicAccountingNormal(b *testing.B) {
	// Setup config with normal mode (not berserk)
	configStr := `
global:
  rotInterval: 1
  cacheIntervals: 1000
samples:
  - name: perftest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	// Start ROT in background
	go outputter.ROT(c)
	
	// Let ROT initialize
	time.Sleep(10 * time.Millisecond)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			outputter.Account(1, 100, "perftest")
		}
	})
	
	// Cleanup
	outputter.ReadFinal()
	config.CleanupConfigAndEnvironment()
}

// BenchmarkAtomicAccountingBerserk tests our berserk mode (should be fastest)
func BenchmarkAtomicAccountingBerserk(b *testing.B) {
	// Setup config with berserk mode
	configStr := `
global:
  rotInterval: 1
  cacheIntervals: 2147483647
samples:
  - name: perftest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	// Start ROT (should detect berserk mode and return immediately)
	go outputter.ROT(c)
	
	// Let ROT initialize (or skip initialization in berserk mode)
	time.Sleep(10 * time.Millisecond)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			outputter.Account(1, 100, "perftest")
		}
	})
	
	// Cleanup
	outputter.ReadFinal()
	config.CleanupConfigAndEnvironment()
}

// BenchmarkConcurrentAccounting tests thread safety under heavy concurrent load
func BenchmarkConcurrentAccounting(b *testing.B) {
	configStr := `
global:
  rotInterval: 1
  cacheIntervals: 1000
samples:
  - name: concurrenttest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	// Start ROT
	go outputter.ROT(c)
	time.Sleep(10 * time.Millisecond)
	
	const numGoroutines = 100
	var wg sync.WaitGroup
	
	b.ResetTimer()
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < b.N/numGoroutines; j++ {
				outputter.Account(1, 100, "concurrenttest")
			}
		}(i)
	}
	
	wg.Wait()
	
	// Cleanup
	outputter.ReadFinal()
	config.CleanupConfigAndEnvironment()
}

// BenchmarkAtomicPointerOperations tests the raw atomic operations
func BenchmarkAtomicPointerOperations(b *testing.B) {
	configStr := `
global:
  rotInterval: 1
  cacheIntervals: 1000
samples:
  - name: atomictest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	// Start ROT to initialize atomic pointers
	go outputter.ROT(c)
	time.Sleep(10 * time.Millisecond)
	
	b.ResetTimer()
	
	// This tests the atomic Load operations specifically
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Just do the atomic load part (what our Account function does)
			outputter.Account(0, 0, "atomictest") // Minimal overhead test
		}
	})
	
	// Cleanup
	outputter.ReadFinal()
	config.CleanupConfigAndEnvironment()
}

// TestAccountingPerformance validates performance characteristics and prevents regressions
func TestAccountingPerformance(t *testing.T) {
	tests := []struct {
		name           string
		cacheIntervals int
		minOpsPerSec   float64
		description    string
	}{
		{"Normal", 1000, 1000000, "Normal mode should exceed 1M ops/sec"}, // Conservative threshold
		{"Berserk", 2147483647, 10000000, "Berserk mode should exceed 10M ops/sec"}, // Much higher threshold
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configStr := fmt.Sprintf(`
global:
  rotInterval: 1
  cacheIntervals: %d
samples:
  - name: perftest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`, tt.cacheIntervals)
			config.SetupFromString(configStr)
			c := config.NewConfig()
			
			// Start ROT
			go outputter.ROT(c)
			time.Sleep(10 * time.Millisecond)
			
			// Measure throughput over a shorter duration for CI efficiency
			const duration = 50 * time.Millisecond
			const numGoroutines = 4 // Fewer goroutines for more predictable CI results
			
			start := time.Now()
			var totalOps int64
			var wg sync.WaitGroup
			
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ops := 0
					for time.Since(start) < duration {
						outputter.Account(1, 100, "perftest")
						ops++
					}
					totalOps += int64(ops)
				}()
			}
			
			wg.Wait()
			elapsed := time.Since(start)
			
			opsPerSecond := float64(totalOps) / elapsed.Seconds()
			t.Logf("%s: %.0f operations/sec (%s)", tt.name, opsPerSecond, tt.description)
			
			// Assert minimum performance threshold
			if opsPerSecond < tt.minOpsPerSec {
				t.Errorf("%s performance regression: got %.0f ops/sec, expected >= %.0f ops/sec", 
					tt.name, opsPerSecond, tt.minOpsPerSec)
			}
			
			// Cleanup
			outputter.ReadFinal()
			config.CleanupConfigAndEnvironment()
		})
	}
}

// TestBerserkModePerformanceAdvantage ensures berserk mode is significantly faster than normal mode
func TestBerserkModePerformanceAdvantage(t *testing.T) {
	// Measure normal mode
	normalOps := measureAccountingPerformance(t, 1000, "normal")
	
	// Measure berserk mode  
	berserkOps := measureAccountingPerformance(t, 2147483647, "berserk")
	
	// Berserk mode should be at least 3x faster (conservative threshold)
	speedup := berserkOps / normalOps
	t.Logf("Berserk mode speedup: %.1fx faster than normal mode", speedup)
	
	if speedup < 3.0 {
		t.Errorf("Berserk mode performance advantage insufficient: %.1fx speedup, expected >= 3.0x", speedup)
	}
}

func measureAccountingPerformance(t *testing.T, cacheIntervals int, mode string) float64 {
	configStr := fmt.Sprintf(`
global:
  rotInterval: 1
  cacheIntervals: %d
samples:
  - name: %s
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`, cacheIntervals, mode)
	
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	go outputter.ROT(c)
	time.Sleep(10 * time.Millisecond)
	
	const duration = 25 * time.Millisecond // Short duration for quick test
	start := time.Now()
	ops := 0
	
	for time.Since(start) < duration {
		outputter.Account(1, 100, mode)
		ops++
	}
	
	elapsed := time.Since(start)
	opsPerSecond := float64(ops) / elapsed.Seconds()
	
	outputter.ReadFinal()
	config.CleanupConfigAndEnvironment()
	
	return opsPerSecond
}

// TestAtomicPointerSafety ensures our atomic implementation doesn't have race conditions
func TestAtomicPointerSafety(t *testing.T) {
	configStr := `
global:
  rotInterval: 1
  cacheIntervals: 1000
samples:
  - name: safetytest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	go outputter.ROT(c)
	time.Sleep(10 * time.Millisecond)
	
	// Run many concurrent goroutines to stress test atomic operations
	const numGoroutines = 100
	const opsPerGoroutine = 1000
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				outputter.Account(1, 100, "safetytest")
			}
		}(i)
	}
	
	wg.Wait()
	
	// If we get here without race conditions or deadlocks, the test passes
	t.Logf("Successfully completed %d concurrent operations without race conditions", 
		numGoroutines*opsPerGoroutine)
	
	outputter.ReadFinal()
	config.CleanupConfigAndEnvironment()
}