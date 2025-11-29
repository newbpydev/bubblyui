package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTimingTracker tests tracker creation.
func TestNewTimingTracker(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates tracker with default settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTimingTracker()

			require.NotNil(t, tracker, "tracker should not be nil")
			assert.Equal(t, 0, tracker.OperationCount(), "should have no operations initially")
		})
	}
}

// TestNewTimingTrackerWithMaxSamples tests custom max samples.
func TestNewTimingTrackerWithMaxSamples(t *testing.T) {
	tests := []struct {
		name       string
		maxSamples int
		wantValid  bool
	}{
		{name: "custom max samples", maxSamples: 5000, wantValid: true},
		{name: "zero defaults to default", maxSamples: 0, wantValid: true},
		{name: "negative defaults to default", maxSamples: -1, wantValid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTimingTrackerWithMaxSamples(tt.maxSamples)

			require.NotNil(t, tracker)
			assert.Equal(t, 0, tracker.OperationCount())
		})
	}
}

// TestTimingTracker_Record tests recording durations.
func TestTimingTracker_Record(t *testing.T) {
	tests := []struct {
		name       string
		operations []struct {
			name     string
			duration time.Duration
		}
		wantCount     int64
		wantTotal     time.Duration
		wantMin       time.Duration
		wantMax       time.Duration
		wantMean      time.Duration
		operationName string
	}{
		{
			name: "single recording",
			operations: []struct {
				name     string
				duration time.Duration
			}{
				{"test.op", 10 * time.Millisecond},
			},
			operationName: "test.op",
			wantCount:     1,
			wantTotal:     10 * time.Millisecond,
			wantMin:       10 * time.Millisecond,
			wantMax:       10 * time.Millisecond,
			wantMean:      10 * time.Millisecond,
		},
		{
			name: "multiple recordings same operation",
			operations: []struct {
				name     string
				duration time.Duration
			}{
				{"test.op", 10 * time.Millisecond},
				{"test.op", 20 * time.Millisecond},
				{"test.op", 30 * time.Millisecond},
			},
			operationName: "test.op",
			wantCount:     3,
			wantTotal:     60 * time.Millisecond,
			wantMin:       10 * time.Millisecond,
			wantMax:       30 * time.Millisecond,
			wantMean:      20 * time.Millisecond,
		},
		{
			name: "min/max update correctly",
			operations: []struct {
				name     string
				duration time.Duration
			}{
				{"test.op", 50 * time.Millisecond},
				{"test.op", 10 * time.Millisecond},
				{"test.op", 100 * time.Millisecond},
				{"test.op", 25 * time.Millisecond},
			},
			operationName: "test.op",
			wantCount:     4,
			wantTotal:     185 * time.Millisecond,
			wantMin:       10 * time.Millisecond,
			wantMax:       100 * time.Millisecond,
			wantMean:      46250 * time.Microsecond, // 185/4 = 46.25ms
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTimingTracker()

			for _, op := range tt.operations {
				tracker.Record(op.name, op.duration)
			}

			stats := tracker.GetStats(tt.operationName)
			require.NotNil(t, stats)

			assert.Equal(t, tt.wantCount, stats.Count)
			assert.Equal(t, tt.wantTotal, stats.Total)
			assert.Equal(t, tt.wantMin, stats.Min)
			assert.Equal(t, tt.wantMax, stats.Max)
			assert.Equal(t, tt.wantMean, stats.Mean)
		})
	}
}

// TestTimingTracker_Percentiles tests percentile calculation.
func TestTimingTracker_Percentiles(t *testing.T) {
	tests := []struct {
		name      string
		durations []time.Duration
		wantP50   time.Duration
		wantP95   time.Duration
		wantP99   time.Duration
	}{
		{
			name:      "single sample",
			durations: []time.Duration{10 * time.Millisecond},
			wantP50:   10 * time.Millisecond,
			wantP95:   10 * time.Millisecond,
			wantP99:   10 * time.Millisecond,
		},
		{
			name: "10 samples sequential",
			durations: []time.Duration{
				1 * time.Millisecond,
				2 * time.Millisecond,
				3 * time.Millisecond,
				4 * time.Millisecond,
				5 * time.Millisecond,
				6 * time.Millisecond,
				7 * time.Millisecond,
				8 * time.Millisecond,
				9 * time.Millisecond,
				10 * time.Millisecond,
			},
			// Using nearest-rank method: index = (percentile * n) / 100
			// P50: (50*10)/100 = 5 -> sorted[5] = 6ms
			// P95: (95*10)/100 = 9 -> sorted[9] = 10ms
			// P99: (99*10)/100 = 9 -> sorted[9] = 10ms
			wantP50: 6 * time.Millisecond,
			wantP95: 10 * time.Millisecond,
			wantP99: 10 * time.Millisecond,
		},
		{
			name: "100 samples",
			durations: func() []time.Duration {
				d := make([]time.Duration, 100)
				for i := 0; i < 100; i++ {
					d[i] = time.Duration(i+1) * time.Millisecond
				}
				return d
			}(),
			// P50: (50*100)/100 = 50 -> sorted[50] = 51ms
			// P95: (95*100)/100 = 95 -> sorted[95] = 96ms
			// P99: (99*100)/100 = 99 -> sorted[99] = 100ms
			wantP50: 51 * time.Millisecond,
			wantP95: 96 * time.Millisecond,
			wantP99: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTimingTracker()

			for _, d := range tt.durations {
				tracker.Record("test.op", d)
			}

			stats := tracker.GetStats("test.op")
			require.NotNil(t, stats)

			assert.Equal(t, tt.wantP50, stats.P50, "P50 mismatch")
			assert.Equal(t, tt.wantP95, stats.P95, "P95 mismatch")
			assert.Equal(t, tt.wantP99, stats.P99, "P99 mismatch")
		})
	}
}

// TestTimingTracker_ReservoirSampling tests memory bounding.
func TestTimingTracker_ReservoirSampling(t *testing.T) {
	// Use small max samples for testing
	maxSamples := 100
	tracker := NewTimingTrackerWithMaxSamples(maxSamples)

	// Record more samples than maxSamples
	numSamples := 1000
	for i := 0; i < numSamples; i++ {
		tracker.Record("test.op", time.Duration(i+1)*time.Microsecond)
	}

	// Verify sample count is bounded
	sampleCount := tracker.SampleCountForOperation("test.op")
	assert.Equal(t, maxSamples, sampleCount, "samples should be bounded to maxSamples")

	// Verify count is correct
	stats := tracker.GetStats("test.op")
	require.NotNil(t, stats)
	assert.Equal(t, int64(numSamples), stats.Count, "count should reflect all recordings")
}

// TestTimingTracker_GetStats_NonExistent tests getting stats for unknown operation.
func TestTimingTracker_GetStats_NonExistent(t *testing.T) {
	tracker := NewTimingTracker()

	stats := tracker.GetStats("nonexistent")
	assert.Nil(t, stats, "should return nil for nonexistent operation")
}

// TestTimingTracker_GetStatsSnapshot tests snapshot creation.
func TestTimingTracker_GetStatsSnapshot(t *testing.T) {
	tracker := NewTimingTracker()

	tracker.Record("test.op", 10*time.Millisecond)
	tracker.Record("test.op", 20*time.Millisecond)

	snapshot := tracker.GetStatsSnapshot("test.op")
	require.NotNil(t, snapshot)

	assert.Equal(t, int64(2), snapshot.Count)
	assert.Equal(t, 30*time.Millisecond, snapshot.Total)
	assert.Nil(t, snapshot.samples, "snapshot should not include samples")
	assert.True(t, snapshot.percentilesCalculated)
}

// TestTimingTracker_GetStatsSnapshot_NonExistent tests snapshot for unknown operation.
func TestTimingTracker_GetStatsSnapshot_NonExistent(t *testing.T) {
	tracker := NewTimingTracker()

	snapshot := tracker.GetStatsSnapshot("nonexistent")
	assert.Nil(t, snapshot)
}

// TestTimingTracker_GetAllStats tests getting all stats.
func TestTimingTracker_GetAllStats(t *testing.T) {
	tracker := NewTimingTracker()

	tracker.Record("op1", 10*time.Millisecond)
	tracker.Record("op2", 20*time.Millisecond)
	tracker.Record("op3", 30*time.Millisecond)

	allStats := tracker.GetAllStats()

	assert.Len(t, allStats, 3)
	assert.Contains(t, allStats, "op1")
	assert.Contains(t, allStats, "op2")
	assert.Contains(t, allStats, "op3")
}

// TestTimingTracker_GetOperationNames tests listing operations.
func TestTimingTracker_GetOperationNames(t *testing.T) {
	tracker := NewTimingTracker()

	tracker.Record("op1", 10*time.Millisecond)
	tracker.Record("op2", 20*time.Millisecond)

	names := tracker.GetOperationNames()

	assert.Len(t, names, 2)
	assert.Contains(t, names, "op1")
	assert.Contains(t, names, "op2")
}

// TestTimingTracker_Reset tests clearing all stats.
func TestTimingTracker_Reset(t *testing.T) {
	tracker := NewTimingTracker()

	tracker.Record("op1", 10*time.Millisecond)
	tracker.Record("op2", 20*time.Millisecond)

	assert.Equal(t, 2, tracker.OperationCount())

	tracker.Reset()

	assert.Equal(t, 0, tracker.OperationCount())
	assert.Nil(t, tracker.GetStats("op1"))
	assert.Nil(t, tracker.GetStats("op2"))
}

// TestTimingTracker_ResetOperation tests clearing single operation.
func TestTimingTracker_ResetOperation(t *testing.T) {
	tracker := NewTimingTracker()

	tracker.Record("op1", 10*time.Millisecond)
	tracker.Record("op2", 20*time.Millisecond)

	tracker.ResetOperation("op1")

	assert.Equal(t, 1, tracker.OperationCount())
	assert.Nil(t, tracker.GetStats("op1"))
	assert.NotNil(t, tracker.GetStats("op2"))
}

// TestTimingTracker_OperationCount tests counting operations.
func TestTimingTracker_OperationCount(t *testing.T) {
	tracker := NewTimingTracker()

	assert.Equal(t, 0, tracker.OperationCount())

	tracker.Record("op1", 10*time.Millisecond)
	assert.Equal(t, 1, tracker.OperationCount())

	tracker.Record("op2", 20*time.Millisecond)
	assert.Equal(t, 2, tracker.OperationCount())

	// Same operation doesn't increase count
	tracker.Record("op1", 30*time.Millisecond)
	assert.Equal(t, 2, tracker.OperationCount())
}

// TestTimingTracker_SampleCount tests total sample counting.
func TestTimingTracker_SampleCount(t *testing.T) {
	tracker := NewTimingTracker()

	assert.Equal(t, int64(0), tracker.SampleCount())

	tracker.Record("op1", 10*time.Millisecond)
	tracker.Record("op1", 20*time.Millisecond)
	tracker.Record("op2", 30*time.Millisecond)

	assert.Equal(t, int64(3), tracker.SampleCount())
}

// TestTimingTracker_ThreadSafe tests concurrent access.
func TestTimingTracker_ThreadSafe(t *testing.T) {
	tracker := NewTimingTracker()

	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tracker.Record("concurrent.op", time.Duration(j+1)*time.Microsecond)
			}
		}()
	}

	wg.Wait()

	stats := tracker.GetStats("concurrent.op")
	require.NotNil(t, stats)
	assert.Equal(t, int64(goroutines*iterations), stats.Count)
}

// TestTimingTracker_Accuracy tests timing accuracy within ±1ms.
func TestTimingTracker_Accuracy(t *testing.T) {
	tracker := NewTimingTracker()

	// Record known durations
	durations := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, d := range durations {
		tracker.Record("accuracy.test", d)
	}

	stats := tracker.GetStats("accuracy.test")
	require.NotNil(t, stats)

	// Expected total: 166ms
	expectedTotal := 166 * time.Millisecond
	assert.Equal(t, expectedTotal, stats.Total, "total should be exact")

	// Expected mean: 33.2ms
	expectedMean := 33200 * time.Microsecond
	assert.Equal(t, expectedMean, stats.Mean, "mean should be exact")

	// Min/Max should be exact
	assert.Equal(t, 1*time.Millisecond, stats.Min)
	assert.Equal(t, 100*time.Millisecond, stats.Max)
}

// TestTimingTracker_EmptySamples tests percentiles with no samples.
func TestTimingTracker_EmptySamples(t *testing.T) {
	tracker := NewTimingTracker()

	// Get stats for nonexistent operation
	stats := tracker.GetStats("empty")
	assert.Nil(t, stats)
}

// TestTimingTracker_MultipleOperations tests tracking multiple operations.
func TestTimingTracker_MultipleOperations(t *testing.T) {
	tracker := NewTimingTracker()

	// Record different operations
	tracker.Record("render", 10*time.Millisecond)
	tracker.Record("render", 15*time.Millisecond)
	tracker.Record("update", 5*time.Millisecond)
	tracker.Record("update", 8*time.Millisecond)
	tracker.Record("init", 100*time.Millisecond)

	// Verify each operation has correct stats
	renderStats := tracker.GetStats("render")
	require.NotNil(t, renderStats)
	assert.Equal(t, int64(2), renderStats.Count)
	assert.Equal(t, 25*time.Millisecond, renderStats.Total)

	updateStats := tracker.GetStats("update")
	require.NotNil(t, updateStats)
	assert.Equal(t, int64(2), updateStats.Count)
	assert.Equal(t, 13*time.Millisecond, updateStats.Total)

	initStats := tracker.GetStats("init")
	require.NotNil(t, initStats)
	assert.Equal(t, int64(1), initStats.Count)
	assert.Equal(t, 100*time.Millisecond, initStats.Total)
}

// TestTimingTracker_PercentilesCalculatedOnce tests lazy percentile calculation.
func TestTimingTracker_PercentilesCalculatedOnce(t *testing.T) {
	tracker := NewTimingTracker()

	for i := 0; i < 100; i++ {
		tracker.Record("test.op", time.Duration(i+1)*time.Millisecond)
	}

	// First call calculates percentiles
	stats1 := tracker.GetStats("test.op")
	require.NotNil(t, stats1)
	assert.True(t, stats1.percentilesCalculated)

	// Second call should use cached percentiles
	stats2 := tracker.GetStats("test.op")
	require.NotNil(t, stats2)
	assert.True(t, stats2.percentilesCalculated)

	// Values should be the same
	assert.Equal(t, stats1.P50, stats2.P50)
	assert.Equal(t, stats1.P95, stats2.P95)
	assert.Equal(t, stats1.P99, stats2.P99)
}

// TestTimingTracker_PercentilesRecalculatedAfterRecord tests percentile invalidation.
func TestTimingTracker_PercentilesRecalculatedAfterRecord(t *testing.T) {
	tracker := NewTimingTracker()

	// Record initial samples
	for i := 0; i < 10; i++ {
		tracker.Record("test.op", time.Duration(i+1)*time.Millisecond)
	}

	// Get stats (calculates percentiles)
	stats1 := tracker.GetStats("test.op")
	require.NotNil(t, stats1)
	p50Before := stats1.P50

	// Record more samples
	for i := 0; i < 10; i++ {
		tracker.Record("test.op", time.Duration(100+i)*time.Millisecond)
	}

	// Get stats again (should recalculate)
	stats2 := tracker.GetStats("test.op")
	require.NotNil(t, stats2)

	// P50 should have changed
	assert.NotEqual(t, p50Before, stats2.P50, "P50 should change after new recordings")
}

// TestPercentileIndex tests the percentile index calculation.
func TestPercentileIndex(t *testing.T) {
	tests := []struct {
		name       string
		n          int
		percentile int
		wantIdx    int
	}{
		// Using formula: idx = (percentile * n) / 100, clamped to [0, n-1]
		{name: "P50 of 10", n: 10, percentile: 50, wantIdx: 5},    // (50*10)/100 = 5
		{name: "P95 of 100", n: 100, percentile: 95, wantIdx: 95}, // (95*100)/100 = 95
		{name: "P99 of 100", n: 100, percentile: 99, wantIdx: 99}, // (99*100)/100 = 99
		{name: "P50 of 1", n: 1, percentile: 50, wantIdx: 0},      // (50*1)/100 = 0
		{name: "P99 of 1", n: 1, percentile: 99, wantIdx: 0},      // (99*1)/100 = 0
		{name: "empty", n: 0, percentile: 50, wantIdx: 0},         // edge case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := percentileIndex(tt.n, tt.percentile)
			assert.Equal(t, tt.wantIdx, idx)
		})
	}
}

// TestTimingTracker_SampleCountForOperation tests sample count per operation.
func TestTimingTracker_SampleCountForOperation(t *testing.T) {
	tracker := NewTimingTrackerWithMaxSamples(100)

	// Non-existent operation
	assert.Equal(t, 0, tracker.SampleCountForOperation("nonexistent"))

	// Record some samples
	for i := 0; i < 50; i++ {
		tracker.Record("test.op", time.Duration(i)*time.Microsecond)
	}
	assert.Equal(t, 50, tracker.SampleCountForOperation("test.op"))

	// Record more to hit max
	for i := 0; i < 100; i++ {
		tracker.Record("test.op", time.Duration(i)*time.Microsecond)
	}
	assert.Equal(t, 100, tracker.SampleCountForOperation("test.op"))
}

// TestMin tests the min helper function.
func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 1, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		assert.Equal(t, tt.want, got)
	}
}

// TestTimingTracker_ConcurrentReadWrite tests concurrent reads and writes.
func TestTimingTracker_ConcurrentReadWrite(t *testing.T) {
	tracker := NewTimingTracker()

	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(3)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tracker.Record("concurrent.op", time.Duration(i+1)*time.Microsecond)
		}
	}()

	// Reader goroutine 1
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.GetStats("concurrent.op")
		}
	}()

	// Reader goroutine 2
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.GetAllStats()
		}
	}()

	wg.Wait()

	// Should complete without race conditions
	stats := tracker.GetStats("concurrent.op")
	require.NotNil(t, stats)
	assert.Equal(t, int64(iterations), stats.Count)
}
