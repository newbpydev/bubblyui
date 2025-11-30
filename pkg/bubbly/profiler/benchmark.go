// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

// Baseline represents a performance baseline for regression detection.
//
// Baselines capture performance metrics at a point in time and can be
// saved/loaded for comparison in CI/CD pipelines.
//
// Example:
//
//	baseline := &Baseline{
//	    Name:      "BenchmarkRender",
//	    NsPerOp:   1500,
//	    AllocBytes: 256,
//	}
//	bp.AssertNoRegression(baseline, 0.10) // 10% threshold
type Baseline struct {
	// Name is the benchmark name
	Name string `json:"name"`

	// NsPerOp is the nanoseconds per operation
	NsPerOp int64 `json:"ns_per_op"`

	// AllocBytes is the bytes allocated per operation
	AllocBytes int64 `json:"alloc_bytes"`

	// AllocsPerOp is the number of allocations per operation
	AllocsPerOp int64 `json:"allocs_per_op"`

	// Iterations is the number of benchmark iterations
	Iterations int `json:"iterations"`

	// Timestamp is when the baseline was created
	Timestamp time.Time `json:"timestamp"`

	// GoVersion is the Go version used
	GoVersion string `json:"go_version"`

	// GOOS is the operating system
	GOOS string `json:"goos"`

	// GOARCH is the architecture
	GOARCH string `json:"goarch"`

	// Metadata contains additional key-value pairs
	Metadata map[string]string `json:"metadata,omitempty"`
}

// BenchmarkStats contains statistical information about benchmark results.
//
// Used for analyzing benchmark performance and detecting regressions.
type BenchmarkStats struct {
	// Name is the benchmark name
	Name string

	// Iterations is the number of measurements
	Iterations int

	// NsPerOp is the average nanoseconds per operation
	NsPerOp int64

	// AllocBytes is the bytes allocated per operation
	AllocBytes int64

	// AllocsPerOp is the number of allocations per operation
	AllocsPerOp int64

	// Min is the minimum duration
	Min time.Duration

	// Max is the maximum duration
	Max time.Duration

	// Mean is the average duration
	Mean time.Duration

	// P50 is the 50th percentile (median)
	P50 time.Duration

	// P95 is the 95th percentile
	P95 time.Duration

	// P99 is the 99th percentile
	P99 time.Duration
}

// RegressionInfo contains details about performance regression.
//
// Used to provide detailed information about what regressed and by how much.
type RegressionInfo struct {
	// HasRegression indicates if any regression was detected
	HasRegression bool

	// TimeRegression is the percentage change in execution time (positive = slower)
	TimeRegression float64

	// MemoryRegression is the percentage change in memory usage (positive = more)
	MemoryRegression float64

	// AllocRegression is the percentage change in allocations (positive = more)
	AllocRegression float64

	// Details provides a human-readable description
	Details string

	// Baseline is the baseline used for comparison
	Baseline *Baseline

	// Current is the current stats
	Current *BenchmarkStats
}

// BenchmarkProfiler integrates with Go's testing.B for benchmark profiling.
//
// It provides measurement, baseline comparison, and regression detection
// capabilities for use in benchmark tests.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	func BenchmarkComponent(b *testing.B) {
//	    bp := NewBenchmarkProfiler(b)
//
//	    b.ResetTimer()
//	    for i := 0; i < b.N; i++ {
//	        bp.Measure(func() {
//	            component.Render()
//	        })
//	    }
//
//	    // Check for regression
//	    baseline := LoadBaseline("baseline.json")
//	    bp.AssertNoRegression(baseline, 0.10)
//	}
type BenchmarkProfiler struct {
	// b is the testing.B instance
	b *testing.B

	// metrics is the metric collector for timing
	metrics *MetricCollector

	// baseline is the current baseline for comparison
	baseline *Baseline

	// measurements stores individual measurement durations
	measurements []time.Duration

	// startTime is when profiling started
	startTime time.Time

	// name is the benchmark name
	name string

	// allocBytes tracks memory allocations
	allocBytes int64

	// allocsPerOp tracks allocation count
	allocsPerOp int64

	// mu protects concurrent access
	mu sync.RWMutex
}

// Common errors
var (
	// ErrNilBenchmark is returned when testing.B is nil
	ErrNilBenchmark = errors.New("testing.B cannot be nil")

	// ErrNilBaseline is returned when baseline is nil
	ErrNilBaseline = errors.New("baseline cannot be nil")

	// ErrInvalidThreshold is returned when threshold is invalid
	ErrInvalidThreshold = errors.New("threshold must be between 0.0 and 1.0")

	// ErrNoMeasurements is returned when no measurements exist
	ErrNoMeasurements = errors.New("no measurements recorded")

	// ErrRegressionDetected is returned when performance regression is detected
	ErrRegressionDetected = errors.New("performance regression detected")
)

// NewBenchmarkProfiler creates a new benchmark profiler.
//
// The profiler integrates with testing.B to provide measurement and
// regression detection capabilities.
//
// Parameters:
//   - b: The testing.B instance from the benchmark function
//
// Returns:
//   - *BenchmarkProfiler: A new profiler instance
//
// Example:
//
//	func BenchmarkMyFunc(b *testing.B) {
//	    bp := NewBenchmarkProfiler(b)
//	    // Use bp for measurements
//	}
func NewBenchmarkProfiler(b *testing.B) *BenchmarkProfiler {
	name := ""
	if b != nil {
		name = b.Name()
	}

	return &BenchmarkProfiler{
		b:            b,
		metrics:      NewMetricCollector(),
		measurements: make([]time.Duration, 0, 1000),
		startTime:    time.Now(),
		name:         name,
	}
}

// Measure times the execution of a function and records the duration.
//
// This method should be called inside the benchmark loop (b.N iterations).
// The function is always executed, and its duration is recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	for i := 0; i < b.N; i++ {
//	    bp.Measure(func() {
//	        DoExpensiveOperation()
//	    })
//	}
func (bp *BenchmarkProfiler) Measure(fn func()) {
	if fn == nil {
		return
	}

	start := time.Now()
	fn()
	duration := time.Since(start)

	bp.mu.Lock()
	bp.measurements = append(bp.measurements, duration)
	bp.mu.Unlock()
}

// StartMeasurement begins timing and returns a function to stop timing.
//
// This is useful when you need more control over when timing starts and stops.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stop := bp.StartMeasurement()
//	DoExpensiveOperation()
//	stop()
func (bp *BenchmarkProfiler) StartMeasurement() func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		bp.mu.Lock()
		bp.measurements = append(bp.measurements, duration)
		bp.mu.Unlock()
	}
}

// GetStats returns statistics for all recorded measurements.
//
// Returns nil if no measurements have been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stats := bp.GetStats()
//	fmt.Printf("Mean: %v, P95: %v\n", stats.Mean, stats.P95)
func (bp *BenchmarkProfiler) GetStats() *BenchmarkStats {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	if len(bp.measurements) == 0 {
		return nil
	}

	// Calculate statistics
	var total time.Duration
	minDur := bp.measurements[0]
	maxDur := bp.measurements[0]

	for _, d := range bp.measurements {
		total += d
		if d < minDur {
			minDur = d
		}
		if d > maxDur {
			maxDur = d
		}
	}

	n := len(bp.measurements)
	mean := total / time.Duration(n)

	// Calculate percentiles
	sorted := make([]time.Duration, n)
	copy(sorted, bp.measurements)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	p50 := sorted[percentileIndex(n, 50)]
	p95 := sorted[percentileIndex(n, 95)]
	p99 := sorted[percentileIndex(n, 99)]

	return &BenchmarkStats{
		Name:        bp.name,
		Iterations:  n,
		NsPerOp:     mean.Nanoseconds(),
		AllocBytes:  bp.allocBytes,
		AllocsPerOp: bp.allocsPerOp,
		Min:         minDur,
		Max:         maxDur,
		Mean:        mean,
		P50:         p50,
		P95:         p95,
		P99:         p99,
	}
}

// GetMeasurements returns a copy of all recorded measurements.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) GetMeasurements() []time.Duration {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	result := make([]time.Duration, len(bp.measurements))
	copy(result, bp.measurements)
	return result
}

// Reset clears all recorded measurements.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) Reset() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.measurements = bp.measurements[:0]
	bp.allocBytes = 0
	bp.allocsPerOp = 0
	bp.startTime = time.Now()
}

// SetBaseline sets the baseline for regression comparison.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) SetBaseline(baseline *Baseline) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.baseline = baseline
}

// GetBaseline returns the current baseline.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) GetBaseline() *Baseline {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.baseline
}

// NewBaseline creates a new baseline from current statistics.
//
// Returns nil if no measurements have been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	baseline := bp.NewBaseline("BenchmarkRender")
//	baseline.SaveToFile("baseline.json")
func (bp *BenchmarkProfiler) NewBaseline(name string) *Baseline {
	stats := bp.GetStats()
	if stats == nil {
		return nil
	}

	return &Baseline{
		Name:        name,
		NsPerOp:     stats.NsPerOp,
		AllocBytes:  stats.AllocBytes,
		AllocsPerOp: stats.AllocsPerOp,
		Iterations:  stats.Iterations,
		Timestamp:   time.Now(),
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		Metadata:    make(map[string]string),
	}
}

// SaveBaseline saves the current statistics as a baseline to a file.
//
// The baseline is saved in JSON format.
//
// Parameters:
//   - filename: Path to the output file
//
// Returns:
//   - error: nil on success, error on failure
//
// Example:
//
//	err := bp.SaveBaseline("baseline.json")
func (bp *BenchmarkProfiler) SaveBaseline(filename string) error {
	baseline := bp.NewBaseline(bp.name)
	if baseline == nil {
		return ErrNoMeasurements
	}

	return baseline.SaveToFile(filename)
}

// SaveToFile saves the baseline to a JSON file.
//
// Parameters:
//   - filename: Path to the output file
//
// Returns:
//   - error: nil on success, error on failure
func (b *Baseline) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write baseline file: %w", err)
	}

	return nil
}

// LoadBaseline loads a baseline from a JSON file.
//
// Parameters:
//   - filename: Path to the baseline file
//
// Returns:
//   - *Baseline: The loaded baseline
//   - error: nil on success, error on failure
//
// Example:
//
//	baseline, err := LoadBaseline("baseline.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	bp.AssertNoRegression(baseline, 0.10)
func LoadBaseline(filename string) (*Baseline, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseline Baseline
	err = json.Unmarshal(data, &baseline)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal baseline: %w", err)
	}

	return &baseline, nil
}

// HasRegression checks if current performance has regressed from baseline.
//
// A regression is detected if any metric exceeds the threshold percentage.
// For example, a threshold of 0.10 means 10% regression is allowed.
//
// Parameters:
//   - baseline: The baseline to compare against
//   - threshold: Maximum allowed regression (0.0 to 1.0)
//
// Returns:
//   - bool: true if regression detected, false otherwise
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if bp.HasRegression(baseline, 0.10) {
//	    b.Error("Performance regression detected")
//	}
func (bp *BenchmarkProfiler) HasRegression(baseline *Baseline, threshold float64) bool {
	if baseline == nil {
		return false
	}

	info := bp.GetRegressionInfo(baseline)
	if info == nil {
		return false
	}

	// Check if any regression exceeds threshold
	return info.TimeRegression > threshold ||
		info.MemoryRegression > threshold ||
		info.AllocRegression > threshold
}

// AssertNoRegression fails the benchmark if regression exceeds threshold.
//
// This method is designed for use in CI/CD pipelines to catch performance
// regressions automatically.
//
// Parameters:
//   - baseline: The baseline to compare against
//   - threshold: Maximum allowed regression (0.0 to 1.0)
//
// Returns:
//   - error: nil if no regression, ErrRegressionDetected if regression found
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if err := bp.AssertNoRegression(baseline, 0.10); err != nil {
//	    b.Fatal(err)
//	}
func (bp *BenchmarkProfiler) AssertNoRegression(baseline *Baseline, threshold float64) error {
	if baseline == nil {
		return ErrNilBaseline
	}

	if threshold < 0.0 || threshold > 1.0 {
		return ErrInvalidThreshold
	}

	info := bp.GetRegressionInfo(baseline)
	if info == nil {
		return ErrNoMeasurements
	}

	if info.HasRegression && (info.TimeRegression > threshold ||
		info.MemoryRegression > threshold ||
		info.AllocRegression > threshold) {
		// Report to testing.B if available
		if bp.b != nil {
			bp.b.Errorf("Performance regression detected: %s", info.Details)
		}
		return fmt.Errorf("%w: %s", ErrRegressionDetected, info.Details)
	}

	return nil
}

// GetRegressionInfo returns detailed regression information.
//
// Returns nil if no measurements have been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	info := bp.GetRegressionInfo(baseline)
//	if info.HasRegression {
//	    fmt.Printf("Time regression: %.2f%%\n", info.TimeRegression*100)
//	}
func (bp *BenchmarkProfiler) GetRegressionInfo(baseline *Baseline) *RegressionInfo {
	stats := bp.GetStats()
	if stats == nil {
		return nil
	}

	if baseline == nil {
		return &RegressionInfo{
			HasRegression: false,
			Current:       stats,
			Details:       "no baseline provided",
		}
	}

	// Calculate regression percentages
	var timeReg, memReg, allocReg float64

	if baseline.NsPerOp > 0 {
		timeReg = float64(stats.NsPerOp-baseline.NsPerOp) / float64(baseline.NsPerOp)
	}

	if baseline.AllocBytes > 0 {
		memReg = float64(stats.AllocBytes-baseline.AllocBytes) / float64(baseline.AllocBytes)
	}

	if baseline.AllocsPerOp > 0 {
		allocReg = float64(stats.AllocsPerOp-baseline.AllocsPerOp) / float64(baseline.AllocsPerOp)
	}

	// Determine if any regression occurred
	hasRegression := timeReg > 0 || memReg > 0 || allocReg > 0

	// Build details string
	var details string
	if hasRegression {
		details = fmt.Sprintf("time: %+.2f%%, memory: %+.2f%%, allocs: %+.2f%%",
			timeReg*100, memReg*100, allocReg*100)
	} else {
		details = "no regression detected"
	}

	return &RegressionInfo{
		HasRegression:    hasRegression,
		TimeRegression:   timeReg,
		MemoryRegression: memReg,
		AllocRegression:  allocReg,
		Details:          details,
		Baseline:         baseline,
		Current:          stats,
	}
}

// ReportMetrics reports custom metrics to testing.B.
//
// This method reports the profiler's statistics as custom metrics
// that appear in benchmark output.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	bp.ReportMetrics()
//	// Output includes: p50-ns/op, p95-ns/op, p99-ns/op
func (bp *BenchmarkProfiler) ReportMetrics() {
	if bp.b == nil {
		return
	}

	stats := bp.GetStats()
	if stats == nil {
		return
	}

	bp.b.ReportMetric(float64(stats.P50.Nanoseconds()), "p50-ns/op")
	bp.b.ReportMetric(float64(stats.P95.Nanoseconds()), "p95-ns/op")
	bp.b.ReportMetric(float64(stats.P99.Nanoseconds()), "p99-ns/op")
	bp.b.ReportMetric(float64(stats.Min.Nanoseconds()), "min-ns/op")
	bp.b.ReportMetric(float64(stats.Max.Nanoseconds()), "max-ns/op")
}

// String returns a human-readable summary of the profiler statistics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) String() string {
	stats := bp.GetStats()
	if stats == nil {
		return "BenchmarkProfiler: no measurements"
	}

	return fmt.Sprintf("BenchmarkProfiler[%s]: %d iterations, mean=%v, p50=%v, p95=%v, p99=%v, min=%v, max=%v",
		stats.Name, stats.Iterations, stats.Mean, stats.P50, stats.P95, stats.P99, stats.Min, stats.Max)
}

// MeasureCount returns the number of measurements recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) MeasureCount() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return len(bp.measurements)
}

// SetAllocStats sets the allocation statistics.
//
// This is useful when integrating with testing.B's allocation tracking.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) SetAllocStats(allocBytes, allocsPerOp int64) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.allocBytes = allocBytes
	bp.allocsPerOp = allocsPerOp
}

// GetName returns the benchmark name.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) GetName() string {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.name
}

// SetName sets the benchmark name.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (bp *BenchmarkProfiler) SetName(name string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.name = name
}

// Clone creates a copy of the baseline.
//
// Returns a deep copy of the baseline including metadata.
func (b *Baseline) Clone() *Baseline {
	if b == nil {
		return nil
	}

	clone := &Baseline{
		Name:        b.Name,
		NsPerOp:     b.NsPerOp,
		AllocBytes:  b.AllocBytes,
		AllocsPerOp: b.AllocsPerOp,
		Iterations:  b.Iterations,
		Timestamp:   b.Timestamp,
		GoVersion:   b.GoVersion,
		GOOS:        b.GOOS,
		GOARCH:      b.GOARCH,
	}

	if b.Metadata != nil {
		clone.Metadata = make(map[string]string, len(b.Metadata))
		for k, v := range b.Metadata {
			clone.Metadata[k] = v
		}
	}

	return clone
}

// String returns a human-readable representation of the baseline.
func (b *Baseline) String() string {
	if b == nil {
		return "Baseline: nil"
	}

	return fmt.Sprintf("Baseline[%s]: %d ns/op, %d B/op, %d allocs/op (%s, %s/%s)",
		b.Name, b.NsPerOp, b.AllocBytes, b.AllocsPerOp, b.GoVersion, b.GOOS, b.GOARCH)
}

// String returns a human-readable representation of the regression info.
func (ri *RegressionInfo) String() string {
	if ri == nil {
		return "RegressionInfo: nil"
	}

	if !ri.HasRegression {
		return "RegressionInfo: no regression"
	}

	return fmt.Sprintf("RegressionInfo: time=%+.2f%%, memory=%+.2f%%, allocs=%+.2f%%",
		ri.TimeRegression*100, ri.MemoryRegression*100, ri.AllocRegression*100)
}

// String returns a human-readable representation of the benchmark stats.
func (bs *BenchmarkStats) String() string {
	if bs == nil {
		return "BenchmarkStats: nil"
	}

	return fmt.Sprintf("BenchmarkStats[%s]: %d iters, %d ns/op, mean=%v, p50=%v, p95=%v, p99=%v",
		bs.Name, bs.Iterations, bs.NsPerOp, bs.Mean, bs.P50, bs.P95, bs.P99)
}
