package jitter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"

	"gonum.org/v1/gonum/dsp/fourier"
)

// Stats contains basic statistical measures for interval analysis
type Stats struct {
	N          int     `json:"n"`
	Mean       float64 `json:"mean"`
	StdDev     float64 `json:"stddev"`
	Lambda     float64 `json:"lambda"`
	CV         float64 `json:"cv"`
	Burstiness float64 `json:"burstiness"`
}

// KSResult contains Kolmogorov-Smirnov test results
type KSResult struct {
	Statistic float64 `json:"ks_stat"`
	Critical  float64 `json:"ks_critical"`
	Pass      bool    `json:"ks_pass"`
}

// FFTResult contains Fast Fourier Transform periodicity analysis results
type FFTResult struct {
	SCR      float64 `json:"scr"`
	PeakFreq float64 `json:"peak_freq"`
	Period   float64 `json:"period"`
	Pass     bool    `json:"fft_pass"`
}

// TestResult contains all analysis results
type TestResult struct {
	Stats       Stats     `json:"stats"`
	KS          KSResult  `json:"ks"`
	FFT         FFTResult `json:"fft"`
	CVPass      bool      `json:"cv_pass"`
	BurstPass   bool      `json:"burst_pass"`
	OverallPass bool      `json:"pass"`
}

// Options configures the jitter analysis
type Options struct {
	CVThreshold    float64
	BurstThreshold float64
	SCRThreshold   float64
	JSONOutput     bool
	Quiet          bool
	Reader         io.Reader
}

// Analyze performs jitter analysis on intervals read from the provided reader
// Returns an error if there's a problem reading/parsing data, or if tests fail.
func Analyze(opts Options) error {
	// Read intervals from input
	intervals, err := readIntervals(opts.Reader)
	if err != nil {
		return err
	}

	result, err := AnalyzeIntervals(intervals, opts)
	if err != nil {
		return err
	}

	// Output results
	if opts.JSONOutput {
		if err := printJSON(result); err != nil {
			return err
		}
	} else if opts.Quiet {
		if result.OverallPass {
			fmt.Println("PASS")
		} else {
			fmt.Println("FAIL")
		}
	} else {
		printResults(result, opts.CVThreshold, opts.BurstThreshold, opts.SCRThreshold)
	}

	if !result.OverallPass {
		return fmt.Errorf("jitter analysis failed")
	}
	return nil
}

// AnalyzeIntervals performs jitter analysis on the provided intervals
func AnalyzeIntervals(intervals []float64, opts Options) (TestResult, error) {
	n := len(intervals)
	if n < 5 {
		return TestResult{}, fmt.Errorf("need at least 5 intervals, got %d", n)
	}

	// Calculate basic statistics
	stats := calculateStats(intervals)

	// Calculate KS statistic
	ks := calculateKS(intervals, stats.Lambda)

	// Calculate FFT periodicity
	fft := calculateFFT(intervals, stats.Mean, opts.SCRThreshold)

	// Determine pass/fail
	cvPass := stats.CV > opts.CVThreshold
	burstPass := stats.Burstiness > opts.BurstThreshold
	overallPass := cvPass && burstPass && fft.Pass

	result := TestResult{
		Stats:       stats,
		KS:          ks,
		FFT:         fft,
		CVPass:      cvPass,
		BurstPass:   burstPass,
		OverallPass: overallPass,
	}

	return result, nil
}

func readIntervals(r io.Reader) ([]float64, error) {
	var intervals []float64
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		val, err := strconv.ParseFloat(scanner.Text(), 64)
		if err == nil && val > 0 {
			intervals = append(intervals, val)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return intervals, nil
}

func calculateStats(intervals []float64) Stats {
	n := len(intervals)

	var sum, sumSq float64
	for _, v := range intervals {
		sum += v
		sumSq += v * v
	}

	mean := sum / float64(n)
	variance := (sumSq / float64(n)) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	stddev := math.Sqrt(variance)

	// Rate parameter lambda = 1/mean
	lambda := 0.0
	if mean > 0 {
		lambda = 1.0 / mean
	}

	// CV (coefficient of variation) = stddev / mean
	// For exponential distribution, CV should be ~1.0
	cv := 0.0
	if mean > 0 {
		cv = stddev / mean
	}

	// Burstiness B = (sigma - mu) / (sigma + mu)
	// For exponential, B ≈ 0; for bursty B > 0; for regular B < 0
	burstiness := 0.0
	if stddev+mean > 0 {
		burstiness = (stddev - mean) / (stddev + mean)
	}

	return Stats{
		N:          n,
		Mean:       mean,
		StdDev:     stddev,
		Lambda:     lambda,
		CV:         cv,
		Burstiness: burstiness,
	}
}

func calculateKS(intervals []float64, lambda float64) KSResult {
	n := len(intervals)

	// Sort intervals for empirical CDF
	sorted := make([]float64, n)
	copy(sorted, intervals)
	sort.Float64s(sorted)

	// Compute KS statistic: D = max|F_n(x) - F(x)|
	// where F(x) = 1 - e^(-lambda*x) for exponential
	maxD := 0.0
	for i, x := range sorted {
		// Empirical CDF at x
		fn := float64(i+1) / float64(n)
		fnPrev := float64(i) / float64(n)

		// Theoretical CDF for exponential
		fx := 0.0
		if lambda > 0 && x > 0 {
			fx = 1.0 - math.Exp(-lambda*x)
		}

		// Check both D+ and D-
		d1 := math.Abs(fn - fx)
		d2 := math.Abs(fnPrev - fx)
		if d1 > maxD {
			maxD = d1
		}
		if d2 > maxD {
			maxD = d2
		}
	}

	// Critical value at 95% confidence: D_crit = 1.36 / sqrt(n)
	critical := 1.36 / math.Sqrt(float64(n))

	return KSResult{
		Statistic: maxD,
		Critical:  critical,
		Pass:      maxD < critical,
	}
}

func calculateFFT(intervals []float64, meanInterval float64, threshold float64) FFTResult {
	n := len(intervals)

	if n < 16 {
		return FFTResult{SCR: 0, PeakFreq: 0, Period: 0, Pass: true}
	}

	// Compute FFT using gonum
	// For real input of length N, returns N/2+1 complex coefficients
	fft := fourier.NewFFT(n)
	coeffs := fft.Coefficients(nil, intervals)

	// Compute power spectrum (exclude DC component at index 0)
	// Power at each frequency bin = |coefficient|² = real² + imag²
	var totalPower, peakPower float64
	var peakIdx int
	for i := 1; i < len(coeffs); i++ {
		power := real(coeffs[i])*real(coeffs[i]) + imag(coeffs[i])*imag(coeffs[i])
		totalPower += power
		if power > peakPower {
			peakPower = power
			peakIdx = i
		}
	}

	// SCR (Spectral Concentration Ratio) = peak_power / total_power
	// High SCR = energy concentrated at one frequency (periodic)
	// Low SCR = energy spread across frequencies (aperiodic/random)
	scr := 0.0
	if totalPower > 0 {
		scr = peakPower / totalPower
	}

	// Convert FFT bin index to physical frequency and period
	// Frequency: f = k / (N × mean_interval) Hz
	peakFreq := float64(peakIdx) / (float64(n) * meanInterval)

	// Period = 1 / frequency (in seconds)
	period := 0.0
	if peakFreq > 0 {
		period = 1.0 / peakFreq
	}

	return FFTResult{
		SCR:      scr,
		PeakFreq: peakFreq,
		Period:   period,
		Pass:     scr < threshold,
	}
}

func printResults(r TestResult, cvThresh, burstThresh, scrThresh float64) {
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("Jitter Analysis")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Printf("Samples analyzed:      %d\n", r.Stats.N)
	fmt.Printf("Mean interval:         %.4f s\n", r.Stats.Mean)
	fmt.Printf("Std deviation:         %.4f s\n", r.Stats.StdDev)
	fmt.Printf("Lambda (MLE):          %.4f events/s\n", r.Stats.Lambda)
	fmt.Println()

	cvMark := mark(r.CVPass)
	burstMark := mark(r.BurstPass)
	ksMark := ""
	if r.KS.Pass {
		ksMark = "✓"
	}

	fmt.Printf("CV (target > %.1f):     %.4f %s\n", cvThresh, r.Stats.CV, cvMark)
	fmt.Printf("Burstiness (target > %.1f): %.4f %s\n", burstThresh, r.Stats.Burstiness, burstMark)
	fmt.Println()
	fmt.Printf("KS Statistic (D):      %.6f (informational)\n", r.KS.Statistic)
	fmt.Printf("KS Critical (95%%):     %.6f %s\n", r.KS.Critical, ksMark)
	fmt.Println()

	fmt.Println("Periodicity Analysis (FFT):")
	fftMark := mark(r.FFT.Pass)
	fmt.Printf("  SCR (target < %.1f):  %.4f %s\n", scrThresh, r.FFT.SCR, fftMark)
	fmt.Printf("  Peak frequency:      %.4f Hz\n", r.FFT.PeakFreq)
	fmt.Printf("  Dominant period:     %.4f s\n", r.FFT.Period)
	fmt.Println()

	if r.OverallPass {
		fmt.Println("RESULT: PASS - Event timing has good variance characteristics")
	} else {
		fmt.Println("RESULT: FAIL - Event timing lacks sufficient variance or has periodicity")
	}
	fmt.Println()
}

func printJSON(r TestResult) error {
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(r); err != nil {
		return err
	}
	return nil
}

func mark(pass bool) string {
	if pass {
		return "✓"
	}
	return "✗"
}
