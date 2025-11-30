// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

// DefaultMaxSamples is the default maximum number of samples to retain for percentile calculation.
const DefaultMaxSamples = 10000

// TimingTracker tracks operation timing statistics with percentile calculation.
//
// It maintains statistics for multiple named operations, including count, total time,
// min/max/mean durations, and percentiles (P50, P95, P99). Memory is bounded using
// reservoir sampling when the number of samples exceeds MaxSamples.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	tt := NewTimingTracker()
//	tt.Record("render", 5*time.Millisecond)
//	tt.Record("render", 10*time.Millisecond)
//	stats := tt.GetStats("render")
//	fmt.Printf("Mean: %v, P95: %v\n", stats.Mean, stats.P95)
type TimingTracker struct {
	// operations maps operation names to their statistics
	operations map[string]*TimingStats

	// maxSamples is the maximum number of samples to retain per operation
	maxSamples int

	// rng is the random number generator for reservoir sampling
	rng *rand.Rand

	// mu protects concurrent access
	mu sync.RWMutex
}

// TimingStats holds statistics for a single operation.
//
// Statistics are updated incrementally as samples are recorded.
// Percentiles are calculated on-demand when GetStats() is called.
type TimingStats struct {
	// Count is the number of times the operation was recorded
	Count int64

	// Total is the cumulative duration of all recordings
	Total time.Duration

	// Min is the shortest duration recorded
	Min time.Duration

	// Max is the longest duration recorded
	Max time.Duration

	// Mean is the average duration (Total / Count)
	Mean time.Duration

	// P50 is the 50th percentile (median)
	P50 time.Duration

	// P95 is the 95th percentile
	P95 time.Duration

	// P99 is the 99th percentile
	P99 time.Duration

	// samples stores individual samples for percentile calculation
	// Uses reservoir sampling when Count exceeds maxSamples
	samples []time.Duration

	// percentilesCalculated indicates if percentiles have been calculated
	// for the current sample set
	percentilesCalculated bool
}

// NewTimingTracker creates a new timing tracker with default settings.
//
// Example:
//
//	tt := NewTimingTracker()
//	tt.Record("operation", duration)
func NewTimingTracker() *TimingTracker {
	return NewTimingTrackerWithMaxSamples(DefaultMaxSamples)
}

// NewTimingTrackerWithMaxSamples creates a new timing tracker with a custom max samples limit.
//
// The maxSamples parameter controls memory usage. Higher values provide more accurate
// percentiles but use more memory. The default is 10,000 samples per operation.
//
// Example:
//
//	tt := NewTimingTrackerWithMaxSamples(5000)
func NewTimingTrackerWithMaxSamples(maxSamples int) *TimingTracker {
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamples
	}
	return &TimingTracker{
		operations: make(map[string]*TimingStats),
		maxSamples: maxSamples,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec // Not used for security
	}
}

// Record records a duration for a named operation.
//
// This method updates all statistics incrementally:
// - Count is incremented
// - Total is updated
// - Min/Max are updated if necessary
// - Mean is recalculated
// - Sample is added (using reservoir sampling if at capacity)
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tt.Record("render.component", 5*time.Millisecond)
func (tt *TimingTracker) Record(name string, duration time.Duration) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	stats, ok := tt.operations[name]
	if !ok {
		stats = &TimingStats{
			Min:     duration,
			Max:     duration,
			samples: make([]time.Duration, 0, min(1000, tt.maxSamples)),
		}
		tt.operations[name] = stats
	}

	stats.Count++
	stats.Total += duration

	// Update min/max
	if duration < stats.Min {
		stats.Min = duration
	}
	if duration > stats.Max {
		stats.Max = duration
	}

	// Update mean
	stats.Mean = time.Duration(int64(stats.Total) / stats.Count)

	// Add sample using reservoir sampling
	if len(stats.samples) < tt.maxSamples {
		stats.samples = append(stats.samples, duration)
	} else {
		// Reservoir sampling: replace random element with probability maxSamples/Count
		// This ensures each sample has equal probability of being in the reservoir
		j := tt.rng.Int63n(stats.Count)
		if j < int64(tt.maxSamples) {
			stats.samples[j] = duration
		}
	}

	// Mark percentiles as needing recalculation
	stats.percentilesCalculated = false
}

// GetStats returns statistics for a named operation.
//
// Returns nil if the operation has not been recorded.
// Percentiles are calculated on-demand when this method is called.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stats := tt.GetStats("render")
//	if stats != nil {
//	    fmt.Printf("Count: %d, Mean: %v, P95: %v\n", stats.Count, stats.Mean, stats.P95)
//	}
func (tt *TimingTracker) GetStats(name string) *TimingStats {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	stats, ok := tt.operations[name]
	if !ok {
		return nil
	}

	// Calculate percentiles if not already done
	if !stats.percentilesCalculated {
		stats.calculatePercentiles()
	}

	return stats
}

// GetStatsSnapshot returns a copy of statistics for a named operation.
//
// Unlike GetStats, this returns a copy that is safe to use without holding locks.
// Returns nil if the operation has not been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) GetStatsSnapshot(name string) *TimingStats {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	stats, ok := tt.operations[name]
	if !ok {
		return nil
	}

	// Calculate percentiles if not already done
	if !stats.percentilesCalculated {
		stats.calculatePercentiles()
	}

	// Return a copy
	return &TimingStats{
		Count:                 stats.Count,
		Total:                 stats.Total,
		Min:                   stats.Min,
		Max:                   stats.Max,
		Mean:                  stats.Mean,
		P50:                   stats.P50,
		P95:                   stats.P95,
		P99:                   stats.P99,
		samples:               nil, // Don't copy samples to save memory
		percentilesCalculated: true,
	}
}

// GetAllStats returns statistics for all operations.
//
// Percentiles are calculated for all operations before returning.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) GetAllStats() map[string]*TimingStats {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	result := make(map[string]*TimingStats, len(tt.operations))
	for name, stats := range tt.operations {
		// Calculate percentiles if not already done
		if !stats.percentilesCalculated {
			stats.calculatePercentiles()
		}
		result[name] = stats
	}
	return result
}

// GetOperationNames returns the names of all recorded operations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) GetOperationNames() []string {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	names := make([]string, 0, len(tt.operations))
	for name := range tt.operations {
		names = append(names, name)
	}
	return names
}

// Reset clears all recorded statistics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) Reset() {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	tt.operations = make(map[string]*TimingStats)
}

// ResetOperation clears statistics for a specific operation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) ResetOperation(name string) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	delete(tt.operations, name)
}

// OperationCount returns the number of unique operations being tracked.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) OperationCount() int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	return len(tt.operations)
}

// SampleCount returns the total number of samples across all operations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) SampleCount() int64 {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	var total int64
	for _, stats := range tt.operations {
		total += stats.Count
	}
	return total
}

// calculatePercentiles calculates P50, P95, and P99 from the samples.
//
// This method sorts a copy of the samples to calculate percentiles.
// Must be called while holding the lock.
func (ts *TimingStats) calculatePercentiles() {
	if len(ts.samples) == 0 {
		ts.P50 = 0
		ts.P95 = 0
		ts.P99 = 0
		ts.percentilesCalculated = true
		return
	}

	// Sort a copy of samples to avoid modifying the original
	sorted := make([]time.Duration, len(ts.samples))
	copy(sorted, ts.samples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	n := len(sorted)

	// Calculate percentiles using nearest-rank method
	// P50 = value at index ceil(0.50 * n) - 1
	// P95 = value at index ceil(0.95 * n) - 1
	// P99 = value at index ceil(0.99 * n) - 1
	ts.P50 = sorted[percentileIndex(n, 50)]
	ts.P95 = sorted[percentileIndex(n, 95)]
	ts.P99 = sorted[percentileIndex(n, 99)]

	ts.percentilesCalculated = true
}

// percentileIndex calculates the index for a given percentile using nearest-rank method.
func percentileIndex(n, percentile int) int {
	if n == 0 {
		return 0
	}
	// Nearest rank method: index = ceil(percentile/100 * n) - 1
	// Simplified: index = (percentile * n - 1) / 100, but we need to handle edge cases
	idx := (percentile * n) / 100
	if idx >= n {
		idx = n - 1
	}
	if idx < 0 {
		idx = 0
	}
	return idx
}

// SampleCountForOperation returns the number of samples retained for a specific operation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tt *TimingTracker) SampleCountForOperation(name string) int {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	stats, ok := tt.operations[name]
	if !ok {
		return 0
	}
	return len(stats.samples)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
