package devtools

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPerformanceMonitor_NewPerformanceMonitor tests creation of performance monitor
func TestPerformanceMonitor_NewPerformanceMonitor(t *testing.T) {
	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	require.NotNil(t, pm)
	assert.NotNil(t, pm.data)
}

// TestPerformanceMonitor_RecordRender tests recording render metrics
func TestPerformanceMonitor_RecordRender(t *testing.T) {
	tests := []struct {
		name          string
		componentID   string
		componentName string
		duration      time.Duration
		wantRecorded  bool
	}{
		{
			name:          "first render",
			componentID:   "comp-1",
			componentName: "Counter",
			duration:      5 * time.Millisecond,
			wantRecorded:  true,
		},
		{
			name:          "second render same component",
			componentID:   "comp-1",
			componentName: "Counter",
			duration:      3 * time.Millisecond,
			wantRecorded:  true,
		},
		{
			name:          "different component",
			componentID:   "comp-2",
			componentName: "Button",
			duration:      2 * time.Millisecond,
			wantRecorded:  true,
		},
	}

	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm.RecordRender(tt.componentID, tt.componentName, tt.duration)

			// Verify data was recorded
			perf := data.GetComponent(tt.componentID)
			require.NotNil(t, perf)
			assert.Equal(t, tt.componentID, perf.ComponentID)
			assert.Equal(t, tt.componentName, perf.ComponentName)
			assert.Greater(t, perf.RenderCount, int64(0))
		})
	}
}

// TestPerformanceMonitor_GetSortedComponents tests sorting functionality
func TestPerformanceMonitor_GetSortedComponents(t *testing.T) {
	tests := []struct {
		name     string
		sortBy   SortBy
		setup    func(*PerformanceMonitor)
		validate func(*testing.T, []*ComponentPerformance)
	}{
		{
			name:   "sort by render count descending",
			sortBy: SortByRenderCount,
			setup: func(pm *PerformanceMonitor) {
				pm.RecordRender("comp-1", "A", 5*time.Millisecond)
				pm.RecordRender("comp-2", "B", 3*time.Millisecond)
				pm.RecordRender("comp-2", "B", 3*time.Millisecond) // 2 renders
				pm.RecordRender("comp-3", "C", 2*time.Millisecond)
				pm.RecordRender("comp-3", "C", 2*time.Millisecond)
				pm.RecordRender("comp-3", "C", 2*time.Millisecond) // 3 renders
			},
			validate: func(t *testing.T, components []*ComponentPerformance) {
				require.Len(t, components, 3)
				assert.Equal(t, int64(3), components[0].RenderCount) // C
				assert.Equal(t, int64(2), components[1].RenderCount) // B
				assert.Equal(t, int64(1), components[2].RenderCount) // A
			},
		},
		{
			name:   "sort by average time descending",
			sortBy: SortByAvgTime,
			setup: func(pm *PerformanceMonitor) {
				pm.RecordRender("comp-1", "Slow", 10*time.Millisecond)
				pm.RecordRender("comp-2", "Medium", 5*time.Millisecond)
				pm.RecordRender("comp-3", "Fast", 1*time.Millisecond)
			},
			validate: func(t *testing.T, components []*ComponentPerformance) {
				require.Len(t, components, 3)
				assert.Equal(t, "Slow", components[0].ComponentName)
				assert.Equal(t, "Medium", components[1].ComponentName)
				assert.Equal(t, "Fast", components[2].ComponentName)
			},
		},
		{
			name:   "sort by max time descending",
			sortBy: SortByMaxTime,
			setup: func(pm *PerformanceMonitor) {
				pm.RecordRender("comp-1", "A", 5*time.Millisecond)
				pm.RecordRender("comp-1", "A", 15*time.Millisecond) // max 15ms
				pm.RecordRender("comp-2", "B", 10*time.Millisecond) // max 10ms
				pm.RecordRender("comp-3", "C", 8*time.Millisecond)  // max 8ms
			},
			validate: func(t *testing.T, components []*ComponentPerformance) {
				require.Len(t, components, 3)
				assert.Equal(t, "A", components[0].ComponentName)
				assert.Equal(t, 15*time.Millisecond, components[0].MaxRenderTime)
				assert.Equal(t, "B", components[1].ComponentName)
				assert.Equal(t, "C", components[2].ComponentName)
			},
		},
		{
			name:   "empty data",
			sortBy: SortByRenderCount,
			setup:  func(pm *PerformanceMonitor) {},
			validate: func(t *testing.T, components []*ComponentPerformance) {
				assert.Empty(t, components)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewPerformanceData()
			pm := NewPerformanceMonitor(data)

			tt.setup(pm)
			components := pm.GetSortedComponents(tt.sortBy)
			tt.validate(t, components)
		})
	}
}

// TestPerformanceMonitor_Render tests rendering output
func TestPerformanceMonitor_Render(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*PerformanceMonitor)
		sortBy   SortBy
		contains []string
	}{
		{
			name: "renders table with components",
			setup: func(pm *PerformanceMonitor) {
				pm.RecordRender("comp-1", "Counter", 5*time.Millisecond)
				pm.RecordRender("comp-1", "Counter", 3*time.Millisecond)
				pm.RecordRender("comp-2", "Button", 2*time.Millisecond)
			},
			sortBy: SortByRenderCount,
			contains: []string{
				"Component Performance",
				"Counter",
				"Button",
				"Renders",
				"Avg Time",
				"Max Time",
			},
		},
		{
			name:   "empty data shows message",
			setup:  func(pm *PerformanceMonitor) {},
			sortBy: SortByRenderCount,
			contains: []string{
				"No performance data",
			},
		},
		{
			name: "shows render counts",
			setup: func(pm *PerformanceMonitor) {
				pm.RecordRender("comp-1", "Test", 1*time.Millisecond)
				pm.RecordRender("comp-1", "Test", 1*time.Millisecond)
				pm.RecordRender("comp-1", "Test", 1*time.Millisecond)
			},
			sortBy: SortByRenderCount,
			contains: []string{
				"Test",
				"3", // render count
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := NewPerformanceData()
			pm := NewPerformanceMonitor(data)

			tt.setup(pm)
			output := pm.Render(tt.sortBy)

			for _, substr := range tt.contains {
				assert.Contains(t, output, substr, "output should contain: %s", substr)
			}
		})
	}
}

// TestPerformanceMonitor_Concurrent tests thread safety
func TestPerformanceMonitor_Concurrent(t *testing.T) {
	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent RecordRender
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			componentID := "comp-1"
			pm.RecordRender(componentID, "Test", time.Duration(id)*time.Microsecond)
		}(i)
	}

	// Concurrent GetSortedComponents
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pm.GetSortedComponents(SortByRenderCount)
		}()
	}

	// Concurrent Render
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pm.Render(SortByAvgTime)
		}()
	}

	wg.Wait()

	// Verify data integrity
	perf := data.GetComponent("comp-1")
	require.NotNil(t, perf)
	assert.Equal(t, int64(numGoroutines), perf.RenderCount)
}

// TestPerformanceMonitor_SetSortBy tests changing sort order
func TestPerformanceMonitor_SetSortBy(t *testing.T) {
	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	// Setup data
	pm.RecordRender("comp-1", "Slow", 10*time.Millisecond)
	pm.RecordRender("comp-2", "Fast", 1*time.Millisecond)
	pm.RecordRender("comp-2", "Fast", 1*time.Millisecond)

	// Test different sort orders
	pm.SetSortBy(SortByRenderCount)
	assert.Equal(t, SortByRenderCount, pm.GetSortBy())

	pm.SetSortBy(SortByAvgTime)
	assert.Equal(t, SortByAvgTime, pm.GetSortBy())

	pm.SetSortBy(SortByMaxTime)
	assert.Equal(t, SortByMaxTime, pm.GetSortBy())
}

// TestPerformanceMonitor_RenderWithDifferentSorts tests rendering with different sort orders
func TestPerformanceMonitor_RenderWithDifferentSorts(t *testing.T) {
	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	// Setup varied data
	pm.RecordRender("comp-1", "A", 10*time.Millisecond)
	pm.RecordRender("comp-2", "B", 5*time.Millisecond)
	pm.RecordRender("comp-2", "B", 5*time.Millisecond)
	pm.RecordRender("comp-3", "C", 1*time.Millisecond)
	pm.RecordRender("comp-3", "C", 1*time.Millisecond)
	pm.RecordRender("comp-3", "C", 1*time.Millisecond)

	// Test each sort order produces output
	sortOrders := []SortBy{SortByRenderCount, SortByAvgTime, SortByMaxTime}
	for _, sortBy := range sortOrders {
		t.Run(sortBy.String(), func(t *testing.T) {
			output := pm.Render(sortBy)
			assert.NotEmpty(t, output)
			assert.Contains(t, output, "Component Performance")
		})
	}
}

// TestPerformanceMonitor_LongComponentNames tests truncation of long names
func TestPerformanceMonitor_LongComponentNames(t *testing.T) {
	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	longName := "VeryLongComponentNameThatShouldBeTruncated"
	pm.RecordRender("comp-1", longName, 1*time.Millisecond)

	output := pm.Render(SortByRenderCount)

	// Should contain truncated version with ellipsis (18 chars max: 15 chars + "...")
	truncated := longName[:15] + "..."
	assert.Contains(t, output, truncated, "should contain truncated name with ellipsis")

	// Should not contain full name
	assert.NotContains(t, output, longName, "should not contain full untruncated name")

	// Verify output is properly formatted (table borders add width, so check reasonable limit)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Table with borders can be up to ~200 chars, which is reasonable for terminal
		assert.LessOrEqual(t, len(line), 200, "line should not be excessively long")
	}
}

// TestPerformanceMonitor_Overhead tests performance overhead is minimal
func TestPerformanceMonitor_Overhead(t *testing.T) {
	data := NewPerformanceData()
	pm := NewPerformanceMonitor(data)

	// Measure overhead of RecordRender
	start := time.Now()
	iterations := 10000
	for i := 0; i < iterations; i++ {
		pm.RecordRender("comp-1", "Test", 1*time.Millisecond)
	}
	elapsed := time.Since(start)

	// Average overhead per call should be < 10 microseconds (< 2% of 1ms render)
	avgOverhead := elapsed / time.Duration(iterations)
	assert.Less(t, avgOverhead, 10*time.Microsecond, "overhead should be < 10µs per call")
}
