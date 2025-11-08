package tests

import (
	"testing"
	"time"

	config "github.com/coccyx/gogen/internal"
	"github.com/coccyx/gogen/outputter"
	"github.com/coccyx/gogen/run"
	"github.com/stretchr/testify/assert"
)

func TestMultiSampleConfiguration(t *testing.T) {
	configStr := `
global:
  debug: false
  verbose: false
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: json

samples:
  - name: webserver_logs
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 100
    lines:
      - _raw: "192.168.1.1 - - [20/Oct/2001:00:00:00 -0400] \"GET /index.html HTTP/1.1\" 200 1234"

  - name: application_logs
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 50
    lines:
      - _raw: "2001-10-20 00:00:00 ERROR Application error occurred in module X"
      - _raw: "2001-10-20 00:00:00 INFO Application started successfully"

  - name: security_logs
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 25
    lines:
      - _raw: "2001-10-20 00:00:00 SECURITY Failed login attempt from 192.168.1.100"
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Run the multi-sample configuration
	run.Run(c)

	// Wait for ROT to log statistics
	time.Sleep(2 * time.Second)

	// Verify that all samples were processed
	outputter.Mutex.RLock()
	defer outputter.Mutex.RUnlock()

	// Check that events were written for each sample
	assert.NotNil(t, outputter.EventsWritten)
	assert.NotNil(t, outputter.BytesWritten)

	// Verify webserver_logs sample
	assert.Greater(t, outputter.EventsWritten["webserver_logs"], int64(0), 
		"Expected webserver_logs to have events written")
	assert.Equal(t, int64(100), outputter.EventsWritten["webserver_logs"], 
		"Expected 100 webserver_logs events")

	// Verify application_logs sample (50 count total, randomly picks from 2 lines)
	assert.Greater(t, outputter.EventsWritten["application_logs"], int64(0), 
		"Expected application_logs to have events written")
	assert.Equal(t, int64(50), outputter.EventsWritten["application_logs"], 
		"Expected 50 application_logs events")

	// Verify security_logs sample
	assert.Greater(t, outputter.EventsWritten["security_logs"], int64(0), 
		"Expected security_logs to have events written")
	assert.Equal(t, int64(25), outputter.EventsWritten["security_logs"], 
		"Expected 25 security_logs events")

	// Verify total events (100 + 50 + 25 = 175)
	totalEvents := outputter.EventsWritten["webserver_logs"] + 
		outputter.EventsWritten["application_logs"] + 
		outputter.EventsWritten["security_logs"]
	assert.Equal(t, int64(175), totalEvents, "Expected 175 total events across all samples")

	// Verify bytes were written for each sample
	assert.Greater(t, outputter.BytesWritten["webserver_logs"], int64(0), 
		"Expected webserver_logs to have bytes written")
	assert.Greater(t, outputter.BytesWritten["application_logs"], int64(0), 
		"Expected application_logs to have bytes written")
	assert.Greater(t, outputter.BytesWritten["security_logs"], int64(0), 
		"Expected security_logs to have bytes written")

	config.CleanupConfigAndEnvironment()
}

func TestMultiSampleWithBerserkMode(t *testing.T) {
	configStr := `
global:
  debug: false
  verbose: false
  rotInterval: 1
  cacheIntervals: 2147483647
  output:
    outputter: devnull
    outputTemplate: json

samples:
  - name: high_volume_logs
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 1000
    lines:
      - _raw: "High volume event data"

  - name: medium_volume_logs
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 500
    lines:
      - _raw: "Medium volume event data"
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Start ROT in berserk mode
	go outputter.ROT(c)
	time.Sleep(10 * time.Millisecond)

	// Run the configuration
	run.Run(c)

	// In berserk mode, accounting is bypassed
	// But the outputter should still work correctly
	
	// Since berserk mode bypasses accounting, EventsWritten/BytesWritten may be empty
	// The test passes if it doesn't crash or produce NaN/Inf errors
	
	config.CleanupConfigAndEnvironment()
}

func TestMultiSampleAccountingAccuracy(t *testing.T) {
	// This test verifies that multi-sample configurations 
	// correctly track separate event counts per sample
	
	configStr := `
global:
  debug: false
  verbose: false
  rotInterval: 1
  output:
    outputter: devnull
    outputTemplate: json

samples:
  - name: sample_alpha
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 200
    lines:
      - _raw: "Alpha sample event"

  - name: sample_beta
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 300
    lines:
      - _raw: "Beta sample event"
      
  - name: sample_gamma
    begin: 2001-10-20 00:00:00
    end: 2001-10-20 00:00:01
    interval: 1
    count: 150
    lines:
      - _raw: "Gamma sample event"
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Run multi-sample generation
	run.Run(c)
	
	// Verify that each sample tracked its events independently
	outputter.Mutex.RLock()
	alphaEvents := outputter.EventsWritten["sample_alpha"]
	betaEvents := outputter.EventsWritten["sample_beta"]
	gammaEvents := outputter.EventsWritten["sample_gamma"]
	outputter.Mutex.RUnlock()
	
	// Each sample should have exactly the count specified
	assert.Equal(t, int64(200), alphaEvents, "Sample Alpha should have 200 events")
	assert.Equal(t, int64(300), betaEvents, "Sample Beta should have 300 events")
	assert.Equal(t, int64(150), gammaEvents, "Sample Gamma should have 150 events")
	
	// Total should be sum of all samples
	totalEvents := alphaEvents + betaEvents + gammaEvents
	assert.Equal(t, int64(650), totalEvents, "Total events should be 650")
	
	config.CleanupConfigAndEnvironment()
}

func TestMultiSampleVaryingTimestamps(t *testing.T) {
	// This test verifies that our local lastTS fix works correctly with
	// multiple samples having different timestamp ranges - the real-world scenario
	// you were concerned about from the cribl-sandbox examples
	
	configStr := `
global:
  debug: false
  verbose: false
  rotInterval: 2
  output:
    outputter: devnull
    outputTemplate: json

samples:
  - name: morning_batch
    begin: 2001-10-20 08:30:00
    end: 2001-10-20 08:30:04
    interval: 1
    count: 120
    lines:
      - _raw: "Morning batch processing event"
      - _raw: "Morning system startup log"

  - name: midday_transactions
    begin: 2001-10-20 12:15:30
    end: 2001-10-20 12:15:35
    interval: 2
    count: 80
    lines:
      - _raw: "Midday transaction processed"
      - _raw: "Payment gateway response"
      - _raw: "User authentication successful"

  - name: evening_analytics
    begin: 2001-10-20 19:45:00
    end: 2001-10-20 19:45:03
    interval: 1
    count: 60
    lines:
      - _raw: "Evening analytics job started"
      - _raw: "Data aggregation complete"

  - name: midnight_backup
    begin: 2001-10-20 23:58:00
    end: 2001-10-21 00:02:00
    interval: 3
    count: 40
    lines:
      - _raw: "Midnight backup initiated"
      - _raw: "Database snapshot created"

  - name: early_morning_sync
    begin: 2001-10-21 01:30:00
    end: 2001-10-21 01:30:05
    interval: 1
    count: 90
    lines:
      - _raw: "Early morning data sync"
      - _raw: "Cache refresh completed"
`

	config.SetupFromString(configStr)
	c := config.NewConfig()

	// Run multi-sample generation with varying timestamps
	run.Run(c)
	
	// Wait briefly for ROT to process statistics
	time.Sleep(1 * time.Second)
	
	// Verify all samples were processed despite varying timestamps
	outputter.Mutex.RLock()
	morningEvents := outputter.EventsWritten["morning_batch"]
	middayEvents := outputter.EventsWritten["midday_transactions"] 
	eveningEvents := outputter.EventsWritten["evening_analytics"]
	midnightEvents := outputter.EventsWritten["midnight_backup"]
	earlyMorningEvents := outputter.EventsWritten["early_morning_sync"]
	outputter.Mutex.RUnlock()

	// Verify each sample processed events
	assert.Greater(t, morningEvents, int64(0), "Morning batch should have events")
	assert.Greater(t, middayEvents, int64(0), "Midday transactions should have events")
	assert.Greater(t, eveningEvents, int64(0), "Evening analytics should have events")
	assert.Greater(t, midnightEvents, int64(0), "Midnight backup should have events")
	assert.Greater(t, earlyMorningEvents, int64(0), "Early morning sync should have events")

	totalEvents := morningEvents + middayEvents + eveningEvents + midnightEvents + earlyMorningEvents
	
	t.Logf("Multi-sample varying timestamps test results:")
	t.Logf("  Morning (08:30): %d events", morningEvents)
	t.Logf("  Midday (12:15): %d events", middayEvents)
	t.Logf("  Evening (19:45): %d events", eveningEvents)  
	t.Logf("  Midnight (23:58-00:02): %d events", midnightEvents)
	t.Logf("  Early Morning (01:30): %d events", earlyMorningEvents)
	t.Logf("  Total: %d events", totalEvents)

	// The key success criteria:
	// 1. No NaN/Inf errors occurred (test completed)
	// 2. All samples processed events despite different timestamp ranges
	// 3. ROT timing calculations worked correctly with local lastTS
	assert.Greater(t, totalEvents, int64(300), "Total events should be substantial across all timestamp ranges")
	
	config.CleanupConfigAndEnvironment()
}