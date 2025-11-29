// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComponentTracker(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates empty tracker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewComponentTracker()

			require.NotNil(t, ct)
			assert.Equal(t, 0, ct.ComponentCount())
		})
	}
}

func TestComponentTracker_RecordRender(t *testing.T) {
	tests := []struct {
		name          string
		componentID   string
		componentName string
		duration      time.Duration
		expectedCount int64
		expectedTotal time.Duration
		expectedAvg   time.Duration
		expectedMax   time.Duration
		expectedMin   time.Duration
	}{
		{
			name:          "records single render",
			componentID:   "comp-1",
			componentName: "Counter",
			duration:      10 * time.Millisecond,
			expectedCount: 1,
			expectedTotal: 10 * time.Millisecond,
			expectedAvg:   10 * time.Millisecond,
			expectedMax:   10 * time.Millisecond,
			expectedMin:   10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewComponentTracker()

			ct.RecordRender(tt.componentID, tt.componentName, tt.duration)

			metrics := ct.GetMetrics(tt.componentID)
			require.NotNil(t, metrics)
			assert.Equal(t, tt.componentID, metrics.ComponentID)
			assert.Equal(t, tt.componentName, metrics.ComponentName)
			assert.Equal(t, tt.expectedCount, metrics.RenderCount)
			assert.Equal(t, tt.expectedTotal, metrics.TotalRenderTime)
			assert.Equal(t, tt.expectedAvg, metrics.AvgRenderTime)
			assert.Equal(t, tt.expectedMax, metrics.MaxRenderTime)
			assert.Equal(t, tt.expectedMin, metrics.MinRenderTime)
		})
	}
}

func TestComponentTracker_RecordRender_MultipleRenders(t *testing.T) {
	tests := []struct {
		name          string
		componentID   string
		componentName string
		durations     []time.Duration
		expectedCount int64
		expectedTotal time.Duration
		expectedAvg   time.Duration
		expectedMax   time.Duration
		expectedMin   time.Duration
	}{
		{
			name:          "aggregates multiple renders",
			componentID:   "comp-1",
			componentName: "Counter",
			durations: []time.Duration{
				10 * time.Millisecond,
				20 * time.Millisecond,
				15 * time.Millisecond,
			},
			expectedCount: 3,
			expectedTotal: 45 * time.Millisecond,
			expectedAvg:   15 * time.Millisecond,
			expectedMax:   20 * time.Millisecond,
			expectedMin:   10 * time.Millisecond,
		},
		{
			name:          "handles same duration renders",
			componentID:   "comp-2",
			componentName: "Button",
			durations: []time.Duration{
				5 * time.Millisecond,
				5 * time.Millisecond,
				5 * time.Millisecond,
			},
			expectedCount: 3,
			expectedTotal: 15 * time.Millisecond,
			expectedAvg:   5 * time.Millisecond,
			expectedMax:   5 * time.Millisecond,
			expectedMin:   5 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewComponentTracker()

			for _, d := range tt.durations {
				ct.RecordRender(tt.componentID, tt.componentName, d)
			}

			metrics := ct.GetMetrics(tt.componentID)
			require.NotNil(t, metrics)
			assert.Equal(t, tt.expectedCount, metrics.RenderCount)
			assert.Equal(t, tt.expectedTotal, metrics.TotalRenderTime)
			assert.Equal(t, tt.expectedAvg, metrics.AvgRenderTime)
			assert.Equal(t, tt.expectedMax, metrics.MaxRenderTime)
			assert.Equal(t, tt.expectedMin, metrics.MinRenderTime)
		})
	}
}

func TestComponentTracker_RecordRender_MultipleComponents(t *testing.T) {
	ct := NewComponentTracker()

	// Record renders for multiple components
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)
	ct.RecordRender("comp-1", "Counter", 15*time.Millisecond)
	ct.RecordRender("comp-3", "Input", 8*time.Millisecond)

	assert.Equal(t, 3, ct.ComponentCount())

	// Check comp-1
	m1 := ct.GetMetrics("comp-1")
	require.NotNil(t, m1)
	assert.Equal(t, int64(2), m1.RenderCount)
	assert.Equal(t, 25*time.Millisecond, m1.TotalRenderTime)

	// Check comp-2
	m2 := ct.GetMetrics("comp-2")
	require.NotNil(t, m2)
	assert.Equal(t, int64(1), m2.RenderCount)
	assert.Equal(t, 5*time.Millisecond, m2.TotalRenderTime)

	// Check comp-3
	m3 := ct.GetMetrics("comp-3")
	require.NotNil(t, m3)
	assert.Equal(t, int64(1), m3.RenderCount)
	assert.Equal(t, 8*time.Millisecond, m3.TotalRenderTime)
}

func TestComponentTracker_GetMetrics_NonExistent(t *testing.T) {
	ct := NewComponentTracker()

	metrics := ct.GetMetrics("non-existent")

	assert.Nil(t, metrics)
}

func TestComponentTracker_GetMetricsSnapshot(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	snapshot := ct.GetMetricsSnapshot("comp-1")
	require.NotNil(t, snapshot)

	// Modify original
	ct.RecordRender("comp-1", "Counter", 20*time.Millisecond)

	// Snapshot should be unchanged
	assert.Equal(t, int64(1), snapshot.RenderCount)
	assert.Equal(t, 10*time.Millisecond, snapshot.TotalRenderTime)

	// Original should be updated
	original := ct.GetMetrics("comp-1")
	assert.Equal(t, int64(2), original.RenderCount)
}

func TestComponentTracker_GetMetricsSnapshot_NonExistent(t *testing.T) {
	ct := NewComponentTracker()

	snapshot := ct.GetMetricsSnapshot("non-existent")

	assert.Nil(t, snapshot)
}

func TestComponentTracker_GetAllMetrics(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)

	all := ct.GetAllMetrics()

	assert.Len(t, all, 2)
	assert.Contains(t, all, "comp-1")
	assert.Contains(t, all, "comp-2")
}

func TestComponentTracker_GetAllMetrics_Empty(t *testing.T) {
	ct := NewComponentTracker()

	all := ct.GetAllMetrics()

	assert.NotNil(t, all)
	assert.Len(t, all, 0)
}

func TestComponentTracker_GetComponentIDs(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)
	ct.RecordRender("comp-3", "Input", 8*time.Millisecond)

	ids := ct.GetComponentIDs()

	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "comp-1")
	assert.Contains(t, ids, "comp-2")
	assert.Contains(t, ids, "comp-3")
}

func TestComponentTracker_GetComponentIDs_Empty(t *testing.T) {
	ct := NewComponentTracker()

	ids := ct.GetComponentIDs()

	assert.NotNil(t, ids)
	assert.Len(t, ids, 0)
}

func TestComponentTracker_ComponentCount(t *testing.T) {
	ct := NewComponentTracker()

	assert.Equal(t, 0, ct.ComponentCount())

	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	assert.Equal(t, 1, ct.ComponentCount())

	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)
	assert.Equal(t, 2, ct.ComponentCount())

	// Same component, count shouldn't increase
	ct.RecordRender("comp-1", "Counter", 15*time.Millisecond)
	assert.Equal(t, 2, ct.ComponentCount())
}

func TestComponentTracker_Reset(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)

	assert.Equal(t, 2, ct.ComponentCount())

	ct.Reset()

	assert.Equal(t, 0, ct.ComponentCount())
	assert.Nil(t, ct.GetMetrics("comp-1"))
	assert.Nil(t, ct.GetMetrics("comp-2"))
}

func TestComponentTracker_ResetComponent(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)

	ct.ResetComponent("comp-1")

	assert.Equal(t, 1, ct.ComponentCount())
	assert.Nil(t, ct.GetMetrics("comp-1"))
	assert.NotNil(t, ct.GetMetrics("comp-2"))
}

func TestComponentTracker_ResetComponent_NonExistent(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	// Should not panic
	ct.ResetComponent("non-existent")

	assert.Equal(t, 1, ct.ComponentCount())
}

func TestComponentTracker_RecordMemoryUsage(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	ct.RecordMemoryUsage("comp-1", 1024)

	metrics := ct.GetMetrics("comp-1")
	require.NotNil(t, metrics)
	assert.Equal(t, uint64(1024), metrics.MemoryUsage)
}

func TestComponentTracker_RecordMemoryUsage_UpdatesExisting(t *testing.T) {
	ct := NewComponentTracker()
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)

	ct.RecordMemoryUsage("comp-1", 1024)
	ct.RecordMemoryUsage("comp-1", 2048)

	metrics := ct.GetMetrics("comp-1")
	require.NotNil(t, metrics)
	assert.Equal(t, uint64(2048), metrics.MemoryUsage)
}

func TestComponentTracker_RecordMemoryUsage_NonExistentComponent(t *testing.T) {
	ct := NewComponentTracker()

	// Should not panic, but also shouldn't create component
	ct.RecordMemoryUsage("non-existent", 1024)

	assert.Equal(t, 0, ct.ComponentCount())
}

func TestComponentTracker_ThreadSafe(t *testing.T) {
	ct := NewComponentTracker()
	const numGoroutines = 50
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			componentID := "comp-" + string(rune('A'+id%5))
			componentName := "Component" + string(rune('A'+id%5))

			for j := 0; j < numOperations; j++ {
				ct.RecordRender(componentID, componentName, time.Duration(j)*time.Microsecond)
				_ = ct.GetMetrics(componentID)
				_ = ct.GetMetricsSnapshot(componentID)
				_ = ct.GetAllMetrics()
				_ = ct.GetComponentIDs()
				_ = ct.ComponentCount()
			}
		}(i)
	}

	wg.Wait()

	// Verify data integrity
	assert.LessOrEqual(t, ct.ComponentCount(), 5)
	assert.Greater(t, ct.ComponentCount(), 0)
}

func TestComponentTracker_StatisticsAccuracy(t *testing.T) {
	ct := NewComponentTracker()

	// Record specific durations
	durations := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
	}

	for _, d := range durations {
		ct.RecordRender("comp-1", "Counter", d)
	}

	metrics := ct.GetMetrics("comp-1")
	require.NotNil(t, metrics)

	// Verify statistics
	assert.Equal(t, int64(5), metrics.RenderCount)
	assert.Equal(t, 15*time.Millisecond, metrics.TotalRenderTime)
	assert.Equal(t, 3*time.Millisecond, metrics.AvgRenderTime) // 15/5 = 3
	assert.Equal(t, 5*time.Millisecond, metrics.MaxRenderTime)
	assert.Equal(t, 1*time.Millisecond, metrics.MinRenderTime)
}

func TestComponentTracker_GetTopComponents(t *testing.T) {
	ct := NewComponentTracker()

	// Create components with different render times
	ct.RecordRender("comp-slow", "SlowComponent", 100*time.Millisecond)
	ct.RecordRender("comp-fast", "FastComponent", 1*time.Millisecond)
	ct.RecordRender("comp-medium", "MediumComponent", 50*time.Millisecond)

	top := ct.GetTopComponents(2, SortByTotalRenderTime)

	require.Len(t, top, 2)
	assert.Equal(t, "comp-slow", top[0].ComponentID)
	assert.Equal(t, "comp-medium", top[1].ComponentID)
}

func TestComponentTracker_GetTopComponents_ByRenderCount(t *testing.T) {
	ct := NewComponentTracker()

	// Create components with different render counts
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 10*time.Millisecond)
	ct.RecordRender("comp-3", "Input", 10*time.Millisecond)
	ct.RecordRender("comp-3", "Input", 10*time.Millisecond)

	top := ct.GetTopComponents(2, SortByRenderCount)

	require.Len(t, top, 2)
	assert.Equal(t, "comp-1", top[0].ComponentID)
	assert.Equal(t, int64(3), top[0].RenderCount)
	assert.Equal(t, "comp-3", top[1].ComponentID)
	assert.Equal(t, int64(2), top[1].RenderCount)
}

func TestComponentTracker_GetTopComponents_ByAvgRenderTime(t *testing.T) {
	ct := NewComponentTracker()

	ct.RecordRender("comp-slow", "SlowComponent", 100*time.Millisecond)
	ct.RecordRender("comp-fast", "FastComponent", 1*time.Millisecond)
	ct.RecordRender("comp-medium", "MediumComponent", 50*time.Millisecond)

	top := ct.GetTopComponents(2, SortByAvgRenderTime)

	require.Len(t, top, 2)
	assert.Equal(t, "comp-slow", top[0].ComponentID)
	assert.Equal(t, "comp-medium", top[1].ComponentID)
}

func TestComponentTracker_GetTopComponents_LimitExceedsCount(t *testing.T) {
	ct := NewComponentTracker()

	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)

	top := ct.GetTopComponents(10, SortByTotalRenderTime)

	assert.Len(t, top, 2)
}

func TestComponentTracker_GetTopComponents_Empty(t *testing.T) {
	ct := NewComponentTracker()

	top := ct.GetTopComponents(5, SortByTotalRenderTime)

	assert.NotNil(t, top)
	assert.Len(t, top, 0)
}

func TestComponentTracker_TotalRenderCount(t *testing.T) {
	ct := NewComponentTracker()

	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp-2", "Button", 5*time.Millisecond)

	assert.Equal(t, int64(3), ct.TotalRenderCount())
}

func TestComponentTracker_TotalRenderCount_Empty(t *testing.T) {
	ct := NewComponentTracker()

	assert.Equal(t, int64(0), ct.TotalRenderCount())
}
