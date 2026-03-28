# Jitter Analysis Package

This package provides statistical analysis of inter-arrival time jitter for event timing validation. It's designed to detect whether event timing has appropriate randomness characteristics or if it exhibits unwanted patterns like periodicity or insufficient variance.

## Usage

### CLI Command

```bash
# Analyze intervals from stdin
cat intervals.txt | gogen jitter

# With custom thresholds
cat intervals.txt | gogen jitter --cv-threshold 0.6 --burst-threshold -0.4

# JSON output
cat intervals.txt | gogen jitter --json

# Quiet mode (only pass/fail)
cat intervals.txt | gogen jitter --quiet
```

### As a Library

```go
import "github.com/coccyx/gogen/jitter"

intervals := []float64{1.5, 2.1, 0.8, 1.2, 3.4, 1.9, 2.5, 1.1, 0.9, 2.8}

opts := jitter.Options{
    CVThreshold:    0.5,
    BurstThreshold: -0.5,
    SCRThreshold:   0.3,
}

result, err := jitter.AnalyzeIntervals(intervals, opts)
if err != nil {
    // Handle error
}

if result.OverallPass {
    // Intervals have good variance characteristics
}
```

## Metrics

The analysis computes several statistical measures:

### Coefficient of Variation (CV)
- **Formula**: σ/μ (standard deviation / mean)
- **Interpretation**: Measures relative variability
- **Target**: > 0.5 (for exponential-like distribution, CV ≈ 1.0)

### Burstiness
- **Formula**: (σ - μ) / (σ + μ)
- **Interpretation**: Characterizes temporal pattern
  - B > 0: Bursty (events clustered)
  - B ≈ 0: Random (exponential-like)
  - B < 0: Regular (periodic-like)
- **Target**: > -0.5 (avoid overly regular patterns)

### Kolmogorov-Smirnov (KS) Test
- **Purpose**: Tests if intervals follow exponential distribution
- **Interpretation**: Informational only (reported but not pass/fail)

### Spectral Concentration Ratio (SCR)
- **Method**: FFT-based periodicity detection
- **Formula**: peak_power / total_power
- **Interpretation**:
  - High SCR: Energy concentrated at one frequency (periodic)
  - Low SCR: Energy spread across frequencies (random)
- **Target**: < 0.3 (avoid strong periodicity)

## Input Format

The tool reads inter-arrival times (in seconds) from stdin, one value per line:

```
1.5
2.1
0.8
1.2
...
```

Minimum 5 intervals required; 16+ recommended for FFT analysis.

## Exit Codes

- 0: All tests passed
- 1: One or more tests failed
- 2: Insufficient data or parsing error

## Flags

- `--cv-threshold`: Minimum CV threshold (default: 0.5)
- `--burst-threshold`: Minimum burstiness threshold (default: -0.5)
- `--scr-threshold`: Maximum SCR threshold (default: 0.3)
- `--json`: Output results as JSON
- `--quiet`, `-q`: Only output PASS/FAIL

## Example Output

```
==========================================
Jitter Analysis
==========================================

Samples analyzed:      100
Mean interval:         0.9775 s
Std deviation:         0.9642 s
Lambda (MLE):          1.0230 events/s

CV (target > 0.5):     0.9864 ✓
Burstiness (target > -0.5): -0.0069 ✓

KS Statistic (D):      0.089234 (informational)
KS Critical (95%):     0.136000 ✓

Periodicity Analysis (FFT):
  SCR (target < 0.3):  0.0123 ✓
  Peak frequency:      0.0204 Hz
  Dominant period:     49.0123 s

RESULT: PASS - Event timing has good variance characteristics
```

## Background

This tool is useful for validating that generated events have realistic timing patterns, especially for security testing scenarios where detection systems might flag perfectly periodic or deterministic event timing as suspicious.
