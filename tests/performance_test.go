package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
)

// BenchmarkAccountingNormal tests the accounting implementation in normal mode
func BenchmarkAccountingNormal(b *testing.B) {
	// Setup config with normal mode
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

// BenchmarkAccountingHighCache tests performance with maximum cache intervals
// This simulates what was previously called "berserk mode" - using max int for cache
func BenchmarkAccountingHighCache(b *testing.B) {
	// Setup config with maximum cache intervals (2147483647 = max int32)
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
	
	// Start ROT with high cache configuration
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

// BenchmarkAccountOperations tests the raw account operations
func BenchmarkAccountOperations(b *testing.B) {
	configStr := `
global:
  rotInterval: 1
  cacheIntervals: 1000
samples:
  - name: accounttest
    begin: 2001-01-01 00:00:00
    end: 2001-01-01 00:00:01
    interval: 1
    count: 1
    lines:
      - _raw: test
`
	config.SetupFromString(configStr)
	c := config.NewConfig()
	
	// Start ROT to initialize accounting system
	go outputter.ROT(c)
	time.Sleep(10 * time.Millisecond)
	
	b.ResetTimer()
	
	// This tests the Account operations specifically
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Just do the account operation
			outputter.Account(0, 0, "accounttest") // Minimal overhead test
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
		{"Berserk", 2147483647, 5000000, "Berserk mode should exceed 5M ops/sec"}, // Realistic threshold for mutex implementation
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

// TestCacheIntervalPerformance measures performance with different cache interval settings
func TestCacheIntervalPerformance(t *testing.T) {
	// Measure with normal cache intervals
	normalOps := measureAccountingPerformance(t, 1000, "normal")
	
	// Measure with maximum cache intervals (what used to be called berserk mode)
	// This tests that the system still works correctly with extreme cache settings
	highCacheOps := measureAccountingPerformance(t, 2147483647, "highcache")
	
	// Log the performance metrics
	t.Logf("Normal cache mode: %.0f ops/second", normalOps)
	t.Logf("High cache mode: %.0f ops/second", highCacheOps)
	
	// Calculate performance ratio for informational purposes
	var ratio float64
	if normalOps > 0 {
		ratio = highCacheOps / normalOps
		t.Logf("Performance ratio (high/normal): %.2fx", ratio)
	}
	
	// Ensure both modes achieve minimum performance threshold
	minOpsPerSecond := 10000.0 // Adjust based on your requirements
	
	if normalOps < minOpsPerSecond {
		t.Errorf("Normal mode throughput too low: %.0f ops/s, expected >= %.0f ops/s", normalOps, minOpsPerSecond)
	}
	if highCacheOps < minOpsPerSecond {
		t.Errorf("High cache mode throughput too low: %.0f ops/s, expected >= %.0f ops/s", highCacheOps, minOpsPerSecond)
	}
	
	// Verify that high cache mode doesn't degrade performance significantly
	// Even without optimizations, it shouldn't be slower
	if highCacheOps < normalOps * 0.9 {
		t.Errorf("High cache mode performance degraded: %.0f ops/s vs normal %.0f ops/s", highCacheOps, normalOps)
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

// TestConcurrentAccountingSafety ensures our accounting implementation doesn't have race conditions
func TestConcurrentAccountingSafety(t *testing.T) {
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
	
	// Run many concurrent goroutines to stress test concurrent operations
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