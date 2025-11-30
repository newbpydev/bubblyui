package profiler

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricCollector_Creation tests collector creation with all trackers.
func TestMetricCollector_Creation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates collector with default state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()

			require.NotNil(t, mc, "collector should not be nil")
			assert.NotNil(t, mc.timings, "timings tracker should not be nil")
			assert.NotNil(t, mc.memory, "memory tracker should not be nil")
			assert.NotNil(t, mc.counters, "counters tracker should not be nil")
			assert.False(t, mc.IsEnabled(), "should be disabled by default")
		})
	}
}

// TestMetricCollector_EnableDisable tests toggling enabled state.
func TestMetricCollector_EnableDisable(t *testing.T) {
	tests := []struct {
		name        string
		actions     []string
		wantEnabled bool
	}{
		{
			name:        "starts disabled",
			actions:     []string{},
			wantEnabled: false,
		},
		{
			name:        "enable turns on collection",
			actions:     []string{"enable"},
			wantEnabled: true,
		},
		{
			name:        "disable turns off collection",
			actions:     []string{"enable", "disable"},
			wantEnabled: false,
		},
		{
			name:        "multiple enables",
			actions:     []string{"enable", "enable", "enable"},
			wantEnabled: true,
		},
		{
			name:        "toggle sequence",
			actions:     []string{"enable", "disable", "enable", "disable", "enable"},
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()

			for _, action := range tt.actions {
				switch action {
				case "enable":
					mc.Enable()
				case "disable":
					mc.Disable()
				}
			}

			assert.Equal(t, tt.wantEnabled, mc.IsEnabled())
		})
	}
}

// TestMetricCollector_Measure tests timing a function.
func TestMetricCollector_Measure(t *testing.T) {
	tests := []struct {
		name            string
		enabled         bool
		sleepDuration   time.Duration
		wantExecuted    bool
		wantMinDuration time.Duration
	}{
		{
			name:            "measures when enabled",
			enabled:         true,
			sleepDuration:   10 * time.Millisecond,
			wantExecuted:    true,
			wantMinDuration: 5 * time.Millisecond,
		},
		{
			name:            "executes but skips timing when disabled",
			enabled:         false,
			sleepDuration:   10 * time.Millisecond,
			wantExecuted:    true,
			wantMinDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()
			if tt.enabled {
				mc.Enable()
			}

			executed := false
			mc.Measure("test.operation", func() {
				executed = true
				time.Sleep(tt.sleepDuration)
			})

			assert.Equal(t, tt.wantExecuted, executed, "function should be executed")

			// Verify timing was recorded when enabled
			if tt.enabled {
				stats := mc.timings.GetStats("test.operation")
				require.NotNil(t, stats, "stats should be recorded when enabled")
				assert.GreaterOrEqual(t, stats.Total, tt.wantMinDuration)
				assert.Equal(t, int64(1), stats.Count)
			}
		})
	}
}

// TestMetricCollector_StartTiming tests the start/stop timing pattern.
func TestMetricCollector_StartTiming(t *testing.T) {
	tests := []struct {
		name            string
		enabled         bool
		sleepDuration   time.Duration
		wantMinDuration time.Duration
	}{
		{
			name:            "records timing when enabled",
			enabled:         true,
			sleepDuration:   10 * time.Millisecond,
			wantMinDuration: 5 * time.Millisecond,
		},
		{
			name:            "no-op closure when disabled",
			enabled:         false,
			sleepDuration:   10 * time.Millisecond,
			wantMinDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()
			if tt.enabled {
				mc.Enable()
			}

			stop := mc.StartTiming("test.timing")
			time.Sleep(tt.sleepDuration)
			stop()

			if tt.enabled {
				stats := mc.timings.GetStats("test.timing")
				require.NotNil(t, stats)
				assert.GreaterOrEqual(t, stats.Total, tt.wantMinDuration)
			}
		})
	}
}

// TestMetricCollector_RecordMetric tests generic metric recording.
func TestMetricCollector_RecordMetric(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		metricName   string
		value        float64
		wantRecorded bool
	}{
		{
			name:         "records metric when enabled",
			enabled:      true,
			metricName:   "test.metric",
			value:        42.5,
			wantRecorded: true,
		},
		{
			name:         "skips recording when disabled",
			enabled:      false,
			metricName:   "test.metric",
			value:        42.5,
			wantRecorded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()
			if tt.enabled {
				mc.Enable()
			}

			mc.RecordMetric(tt.metricName, tt.value)

			counter := mc.counters.GetCounter(tt.metricName)
			if tt.wantRecorded {
				require.NotNil(t, counter)
				assert.Equal(t, tt.value, counter.Value)
			} else {
				// When disabled, counter might be nil or not updated
				if counter != nil {
					assert.Equal(t, float64(0), counter.Value)
				}
			}
		})
	}
}

// TestMetricCollector_ThreadSafe tests concurrent access.
func TestMetricCollector_ThreadSafe(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // 3 types of operations

	// Concurrent Measure calls
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				mc.Measure("concurrent.measure", func() {
					// Simulate work
					time.Sleep(time.Microsecond)
				})
			}
		}()
	}

	// Concurrent StartTiming calls
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				stop := mc.StartTiming("concurrent.timing")
				time.Sleep(time.Microsecond)
				stop()
			}
		}()
	}

	// Concurrent RecordMetric calls
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				mc.RecordMetric("concurrent.metric", float64(j))
			}
		}()
	}

	wg.Wait()

	// Verify counts are correct
	stats := mc.timings.GetStats("concurrent.measure")
	require.NotNil(t, stats)
	assert.Equal(t, int64(goroutines*iterations), stats.Count)

	stats2 := mc.timings.GetStats("concurrent.timing")
	require.NotNil(t, stats2)
	assert.Equal(t, int64(goroutines*iterations), stats2.Count)

	counter := mc.counters.GetCounter("concurrent.metric")
	require.NotNil(t, counter)
	assert.Equal(t, int64(goroutines*iterations), counter.Count)
}

// TestMetricCollector_Overhead tests that overhead is minimal.
func TestMetricCollector_Overhead(t *testing.T) {
	mc := NewMetricCollector()

	// Measure overhead when disabled
	disabledIterations := 100000
	disabledStart := time.Now()
	for i := 0; i < disabledIterations; i++ {
		mc.Measure("overhead.disabled", func() {})
	}
	disabledDuration := time.Since(disabledStart)
	disabledPerOp := float64(disabledDuration.Nanoseconds()) / float64(disabledIterations)

	// Measure overhead when enabled
	mc.Enable()
	enabledIterations := 100000
	enabledStart := time.Now()
	for i := 0; i < enabledIterations; i++ {
		mc.Measure("overhead.enabled", func() {})
	}
	enabledDuration := time.Since(enabledStart)
	enabledPerOp := float64(enabledDuration.Nanoseconds()) / float64(enabledIterations)

	// Overhead ratio should be reasonable
	// When disabled, nearly zero overhead. When enabled, some overhead is expected.
	t.Logf("Disabled: %.2f ns/op", disabledPerOp)
	t.Logf("Enabled: %.2f ns/op", enabledPerOp)

	// Verify disabled path is fast
	// Note: Race detector adds significant overhead (~4x-10x), so we use 1000ns threshold
	// In production without race detector: ~50ns
	// With race detector: ~200-900ns (varies by system load)
	assert.Less(t, disabledPerOp, float64(1000), "disabled path should be fast (race detector adds overhead)")

	// Verify enabled overhead is reasonable (< 10000ns per operation for overhead only)
	// This doesn't include actual work, just profiling overhead
	// Race detector adds ~2x overhead to mutex operations
	assert.Less(t, enabledPerOp, float64(10000), "enabled overhead should be reasonable")
}

// TestMetricCollector_MeasureReturnsClosure tests that StartTiming returns valid closure.
func TestMetricCollector_StartTiming_ReturnsClosure(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	// Get multiple closures
	stop1 := mc.StartTiming("closure.test.1")
	stop2 := mc.StartTiming("closure.test.2")

	// Sleep different amounts
	time.Sleep(5 * time.Millisecond)
	stop1()

	time.Sleep(10 * time.Millisecond)
	stop2()

	// Verify both recorded correctly
	stats1 := mc.timings.GetStats("closure.test.1")
	stats2 := mc.timings.GetStats("closure.test.2")

	require.NotNil(t, stats1)
	require.NotNil(t, stats2)

	// stats2 should have longer duration
	assert.Greater(t, stats2.Total, stats1.Total)
}

// TestMetricCollector_IncrementCounter tests counter increment functionality.
func TestMetricCollector_IncrementCounter(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		increments int
		wantCount  int64
	}{
		{
			name:       "increments when enabled",
			enabled:    true,
			increments: 10,
			wantCount:  10,
		},
		{
			name:       "skips increment when disabled",
			enabled:    false,
			increments: 10,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()
			if tt.enabled {
				mc.Enable()
			}

			for i := 0; i < tt.increments; i++ {
				mc.IncrementCounter("test.counter")
			}

			counter := mc.counters.GetCounter("test.counter")
			if tt.wantCount > 0 {
				require.NotNil(t, counter)
				assert.Equal(t, tt.wantCount, counter.Count)
			}
		})
	}
}

// TestMetricCollector_RecordMemory tests memory metric recording.
func TestMetricCollector_RecordMemory(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		location     string
		size         int64
		wantRecorded bool
	}{
		{
			name:         "records memory when enabled",
			enabled:      true,
			location:     "test.alloc",
			size:         1024,
			wantRecorded: true,
		},
		{
			name:         "skips recording when disabled",
			enabled:      false,
			location:     "test.alloc",
			size:         1024,
			wantRecorded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector()
			if tt.enabled {
				mc.Enable()
			}

			mc.RecordMemory(tt.location, tt.size)

			alloc := mc.memory.GetAllocation(tt.location)
			if tt.wantRecorded {
				require.NotNil(t, alloc)
				assert.Equal(t, int64(1), alloc.Count)
				assert.Equal(t, tt.size, alloc.TotalSize)
			}
		})
	}
}

// TestMetricCollector_Reset tests resetting all metrics.
func TestMetricCollector_Reset(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	// Record some data
	mc.Measure("test.measure", func() {
		time.Sleep(time.Millisecond)
	})
	mc.RecordMetric("test.metric", 42.0)
	mc.RecordMemory("test.memory", 1024)

	// Verify data was recorded
	assert.NotNil(t, mc.timings.GetStats("test.measure"))

	// Reset
	mc.Reset()

	// Verify data was cleared
	assert.Nil(t, mc.timings.GetStats("test.measure"))
	assert.Nil(t, mc.counters.GetCounter("test.metric"))
	assert.Nil(t, mc.memory.GetAllocation("test.memory"))

	// Collector should still be enabled
	assert.True(t, mc.IsEnabled())
}

// TestMetricCollector_GetTimings tests direct access to timing tracker.
func TestMetricCollector_GetTimings(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	mc.Measure("test.op", func() {
		time.Sleep(time.Millisecond)
	})

	timings := mc.GetTimings()
	require.NotNil(t, timings)

	stats := timings.GetStats("test.op")
	require.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Count)
}

// TestMetricCollector_GetMemory tests direct access to memory tracker.
func TestMetricCollector_GetMemory(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	mc.RecordMemory("test.alloc", 1024)

	memory := mc.GetMemory()
	require.NotNil(t, memory)

	alloc := memory.GetAllocation("test.alloc")
	require.NotNil(t, alloc)
	assert.Equal(t, int64(1), alloc.Count)
}

// TestMetricCollector_GetCounters tests direct access to counter tracker.
func TestMetricCollector_GetCounters(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	mc.IncrementCounter("test.counter")

	counters := mc.GetCounters()
	require.NotNil(t, counters)

	counter := counters.GetCounter("test.counter")
	require.NotNil(t, counter)
	assert.Equal(t, int64(1), counter.Count)
}

// TestMemoryTracker_GetAllAllocations tests getting all allocations.
func TestMemoryTracker_GetAllAllocations(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	mc.RecordMemory("alloc1", 1024)
	mc.RecordMemory("alloc2", 2048)

	memory := mc.GetMemory()
	allocs := memory.GetAllAllocations()

	assert.Len(t, allocs, 2)
	assert.Contains(t, allocs, "alloc1")
	assert.Contains(t, allocs, "alloc2")
}

// TestCounterTracker_GetAllCounters tests getting all counters.
func TestCounterTracker_GetAllCounters(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	mc.IncrementCounter("counter1")
	mc.IncrementCounter("counter2")

	counters := mc.GetCounters()
	allCounters := counters.GetAllCounters()

	assert.Len(t, allCounters, 2)
	assert.Contains(t, allCounters, "counter1")
	assert.Contains(t, allCounters, "counter2")
}

// TestMetricCollector_ConcurrentEnableDisable tests toggling while collecting.
func TestMetricCollector_ConcurrentEnableDisable(t *testing.T) {
	mc := NewMetricCollector()
	mc.Enable()

	var wg sync.WaitGroup
	var toggleCount int64
	var measureCount int64

	// Toggle enabled state rapidly
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if i%2 == 0 {
				mc.Enable()
			} else {
				mc.Disable()
			}
			atomic.AddInt64(&toggleCount, 1)
		}
	}()

	// Measure while toggling
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			mc.Measure("concurrent.toggle", func() {})
			atomic.AddInt64(&measureCount, 1)
		}
	}()

	wg.Wait()

	// Should complete without race conditions
	assert.Equal(t, int64(1000), toggleCount)
	assert.Equal(t, int64(1000), measureCount)
}
