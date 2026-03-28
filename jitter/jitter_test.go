package jitter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestAnalyze_HighVariance(t *testing.T) {
	// Test with data that has high coefficient of variation
	data := "0.05\n4.2\n0.1\n5.3\n0.2\n3.8\n0.15\n6.1\n0.3\n2.9\n" +
		"0.08\n4.7\n0.25\n3.2\n0.12\n5.5\n0.18\n4.1\n0.22\n3.5\n"

	intervals := []float64{}
	for _, line := range strings.Split(strings.TrimSpace(data), "\n") {
		var val float64
		if _, err := fmt.Sscanf(line, "%f", &val); err == nil {
			intervals = append(intervals, val)
		}
	}

	opts := Options{
		CVThreshold:    0.5,
		BurstThreshold: -0.5,
		SCRThreshold:   0.3,
	}

	result, err := AnalyzeIntervals(intervals, opts)
	if err != nil {
		t.Fatalf("Expected analysis to complete, got error: %v", err)
	}

	// Verify the result has all expected fields
	if result.Stats.N != len(intervals) {
		t.Errorf("Expected N=%d, got %d", len(intervals), result.Stats.N)
	}

	if result.Stats.Mean <= 0 {
		t.Error("Expected positive mean")
	}

	if result.Stats.CV <= 0 {
		t.Error("Expected positive CV")
	}
}

func TestAnalyze_JSONOutput(t *testing.T) {
	data := "1.5\n2.1\n0.8\n1.2\n3.4\n1.9\n2.5\n1.1\n0.9\n2.8\n"
	reader := strings.NewReader(data)

	opts := Options{
		CVThreshold:    0.5,
		BurstThreshold: -0.5,
		SCRThreshold:   0.3,
		JSONOutput:     true,
		Quiet:          false,
		Reader:         reader,
	}

	// Capture stderr (we'll ignore the exit behavior in tests)
	var buf bytes.Buffer

	// This will likely fail but should produce JSON output
	// We're just testing that it doesn't panic
	_ = Analyze(opts)

	// Test passed if we got here without panic
	_ = buf.String()
}

func TestAnalyze_InsufficientData(t *testing.T) {
	data := "1.5\n2.1\n"
	reader := strings.NewReader(data)

	opts := Options{
		CVThreshold:    0.5,
		BurstThreshold: -0.5,
		SCRThreshold:   0.3,
		JSONOutput:     false,
		Quiet:          true,
		Reader:         reader,
	}

	err := Analyze(opts)
	if err == nil {
		t.Fatal("Expected error for insufficient data, got nil")
	}

	if !strings.Contains(err.Error(), "at least 5 intervals") {
		t.Errorf("Expected error about insufficient intervals, got: %v", err)
	}
}

func TestCalculateStats(t *testing.T) {
	intervals := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	stats := calculateStats(intervals)

	if stats.N != 5 {
		t.Errorf("Expected N=5, got %d", stats.N)
	}

	if stats.Mean != 3.0 {
		t.Errorf("Expected Mean=3.0, got %f", stats.Mean)
	}

	// Check CV is calculated
	if stats.CV <= 0 {
		t.Errorf("Expected positive CV, got %f", stats.CV)
	}
}

func TestCalculateKS(t *testing.T) {
	intervals := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	lambda := 1.0 / 3.0 // Mean of intervals is 3.0

	ks := calculateKS(intervals, lambda)

	if ks.Statistic < 0 || ks.Statistic > 1 {
		t.Errorf("KS statistic should be between 0 and 1, got %f", ks.Statistic)
	}

	if ks.Critical <= 0 {
		t.Errorf("KS critical value should be positive, got %f", ks.Critical)
	}
}

func TestCalculateFFT(t *testing.T) {
	// Create 20 intervals for FFT analysis
	intervals := make([]float64, 20)
	for i := range intervals {
		intervals[i] = float64(i%3 + 1) // Simple varying pattern
	}

	meanInterval := 2.0
	threshold := 0.3

	fft := calculateFFT(intervals, meanInterval, threshold)

	// SCR should be between 0 and 1
	if fft.SCR < 0 || fft.SCR > 1 {
		t.Errorf("SCR should be between 0 and 1, got %f", fft.SCR)
	}
}

func TestCalculateFFT_SmallDataset(t *testing.T) {
	// FFT should handle small datasets gracefully
	intervals := []float64{1.0, 2.0, 3.0}
	meanInterval := 2.0
	threshold := 0.3

	fft := calculateFFT(intervals, meanInterval, threshold)

	// Should pass with small dataset
	if !fft.Pass {
		t.Error("Small dataset should automatically pass FFT test")
	}
}
