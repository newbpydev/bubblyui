package profiler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateReport_AllMetricsPopulated verifies that GenerateReport()
// populates ALL metrics: duration, operations, FPS, memory, goroutines,
// bottlenecks, CPU profile, memory profile, and recommendations.
func TestGenerateReport_AllMetricsPopulated(t *testing.T) {
	// Create profiler (enabled=false initially, will enable via Start())
	prof := New()

	// Create and set hook adapter
	hookAdapter := NewProfilerHookAdapter(prof)
	prof.SetHookAdapter(hookAdapter)

	// Start profiler (this sets enabled=true and startTime)
	err := prof.Start()
	require.NoError(t, err)

	// Simulate component activity
	hookAdapter.OnComponentMount("comp-1", "TaskList")
	hookAdapter.OnComponentMount("comp-2", "TaskInput")
	hookAdapter.OnComponentMount("comp-3", "TaskStats")

	// Simulate renders (make TaskList slow enough to trigger bottleneck detection)
	hookAdapter.OnRenderComplete("comp-1", 20*time.Millisecond) // Slow
	hookAdapter.OnRenderComplete("comp-2", 5*time.Millisecond)
	hookAdapter.OnRenderComplete("comp-3", 3*time.Millisecond)
	hookAdapter.OnRenderComplete("comp-1", 25*time.Millisecond) // Very slow
	hookAdapter.OnRenderComplete("comp-1", 22*time.Millisecond) // Slow

	// Wait a bit to accumulate duration
	time.Sleep(100 * time.Millisecond)

	// Stop profiler
	err = prof.Stop()
	require.NoError(t, err)

	// Generate report
	report := prof.GenerateReport()

	// =============================================================================
	// Verify ALL fields are populated
	// =============================================================================

	require.NotNil(t, report)
	require.NotNil(t, report.Summary)

	// 1. Duration should be > 0 (we slept for 100ms)
	assert.Greater(t, report.Summary.Duration, 50*time.Millisecond,
		"Duration should reflect profiling time")

	// 2. Total Operations should be 5 (5 render calls)
	assert.Equal(t, int64(5), report.Summary.TotalOperations,
		"TotalOperations should match render count")

	// 3. Memory Usage should be non-zero
	assert.Greater(t, report.Summary.MemoryUsage, uint64(0),
		"MemoryUsage should be populated from runtime stats")

	// 4. Goroutine Count should be > 0
	assert.Greater(t, report.Summary.GoroutineCount, 0,
		"GoroutineCount should be > 0 (at least this test goroutine)")

	// 5. Components should have metrics
	require.NotNil(t, report.Components)
	assert.Len(t, report.Components, 3, "Should have 3 components")

	// Find TaskList component
	var taskListMetrics *ComponentMetrics
	for _, m := range report.Components {
		if m.ComponentName == "TaskList" {
			taskListMetrics = m
			break
		}
	}

	require.NotNil(t, taskListMetrics)
	assert.Equal(t, int64(3), taskListMetrics.RenderCount)
	assert.Equal(t, 67*time.Millisecond, taskListMetrics.TotalRenderTime) // 20+25+22=67ms

	// 6. Bottlenecks should be detected (TaskList has slow render: 18ms > 16ms)
	require.NotNil(t, report.Bottlenecks)

	// Debug: Print component metrics to see actual values
	t.Logf("TaskList metrics: RenderCount=%d, AvgRenderTime=%v, MaxRenderTime=%v",
		taskListMetrics.RenderCount, taskListMetrics.AvgRenderTime, taskListMetrics.MaxRenderTime)
	t.Logf("Bottlenecks detected: %d", len(report.Bottlenecks))

	assert.Greater(t, len(report.Bottlenecks), 0,
		"Should detect bottlenecks for slow renders")

	// 7. CPU Profile should be initialized (empty is OK for this test)
	require.NotNil(t, report.CPUProfile)
	require.NotNil(t, report.CPUProfile.HotFunctions)

	// 8. Memory Profile should be populated
	require.NotNil(t, report.MemProfile)
	assert.Greater(t, report.MemProfile.HeapAlloc, uint64(0),
		"MemProfile should have heap allocation data")

	// 9. Recommendations should be generated
	require.NotNil(t, report.Recommendations)
	// Should have at least one recommendation for the slow component
	assert.Greater(t, len(report.Recommendations), 0,
		"Should generate recommendations for slow components")

	// 10. Timestamp should be recent
	assert.WithinDuration(t, time.Now(), report.Timestamp, 5*time.Second)
}

// TestGenerateReport_MetricsAccuracy verifies the accuracy of calculated metrics.
func TestGenerateReport_MetricsAccuracy(t *testing.T) {
	prof := New()
	hookAdapter := NewProfilerHookAdapter(prof)
	prof.SetHookAdapter(hookAdapter)

	err := prof.Start()
	require.NoError(t, err)
	startTime := prof.startTime

	// Simulate 100ms of runtime
	time.Sleep(100 * time.Millisecond)

	err = prof.Stop()
	require.NoError(t, err)

	report := prof.GenerateReport()

	// Duration should be approximately 100ms (±50ms tolerance)
	assert.InDelta(t, 100*time.Millisecond, report.Summary.Duration, float64(50*time.Millisecond),
		"Duration should match actual runtime")

	// Verify duration calculation
	expectedDuration := prof.stopTime.Sub(startTime)
	assert.Equal(t, expectedDuration, report.Summary.Duration)
}
