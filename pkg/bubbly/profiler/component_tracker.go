// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sort"
	"sync"
	"time"
)

// ComponentSortField defines how to sort components when retrieving top performers.
type ComponentSortField int

const (
	// SortByTotalRenderTime sorts components by total cumulative render time (descending).
	SortByTotalRenderTime ComponentSortField = iota

	// SortByRenderCount sorts components by number of renders (descending).
	SortByRenderCount

	// SortByAvgRenderTime sorts components by average render time (descending).
	SortByAvgRenderTime

	// SortByMaxRenderTime sorts components by maximum render time (descending).
	SortByMaxRenderTime
)

// ComponentTracker tracks per-component performance metrics for BubblyUI applications.
//
// It records render timing information for individual components, calculates
// statistics (count, total, average, min, max), and provides methods to
// retrieve and analyze component performance data.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	ct := NewComponentTracker()
//
//	// In your component's View() method
//	start := time.Now()
//	// ... render component ...
//	ct.RecordRender(componentID, componentName, time.Since(start))
//
//	// Get performance metrics
//	metrics := ct.GetMetrics(componentID)
//	fmt.Printf("Component %s: %d renders, avg %v\n",
//	    metrics.ComponentName, metrics.RenderCount, metrics.AvgRenderTime)
type ComponentTracker struct {
	// components maps component IDs to their metrics
	components map[string]*ComponentMetrics

	// mu protects concurrent access to tracker state
	mu sync.RWMutex
}

// NewComponentTracker creates a new component tracker.
//
// Example:
//
//	ct := NewComponentTracker()
//	ct.RecordRender("comp-1", "Counter", 10*time.Millisecond)
func NewComponentTracker() *ComponentTracker {
	return &ComponentTracker{
		components: make(map[string]*ComponentMetrics),
	}
}

// RecordRender records a render duration for a component.
//
// This method updates all statistics incrementally:
//   - RenderCount is incremented
//   - TotalRenderTime is updated
//   - AvgRenderTime is recalculated
//   - MinRenderTime/MaxRenderTime are updated if necessary
//
// If this is the first render for the component, a new metrics entry is created.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	start := time.Now()
//	// ... render component ...
//	ct.RecordRender("comp-1", "Counter", time.Since(start))
func (ct *ComponentTracker) RecordRender(id, name string, duration time.Duration) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	metrics, ok := ct.components[id]
	if !ok {
		metrics = &ComponentMetrics{
			ComponentID:   id,
			ComponentName: name,
			MinRenderTime: duration,
			MaxRenderTime: duration,
		}
		ct.components[id] = metrics
	}

	metrics.RenderCount++
	metrics.TotalRenderTime += duration

	// Update min/max
	if duration < metrics.MinRenderTime {
		metrics.MinRenderTime = duration
	}
	if duration > metrics.MaxRenderTime {
		metrics.MaxRenderTime = duration
	}

	// Calculate average
	metrics.AvgRenderTime = time.Duration(int64(metrics.TotalRenderTime) / metrics.RenderCount)
}

// GetMetrics returns metrics for a specific component.
//
// Returns nil if the component has not been tracked.
//
// Note: The returned pointer points to the internal metrics struct.
// For a safe copy that won't be affected by future updates, use GetMetricsSnapshot.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	metrics := ct.GetMetrics("comp-1")
//	if metrics != nil {
//	    fmt.Printf("Renders: %d, Avg: %v\n", metrics.RenderCount, metrics.AvgRenderTime)
//	}
func (ct *ComponentTracker) GetMetrics(id string) *ComponentMetrics {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	return ct.components[id]
}

// GetMetricsSnapshot returns a copy of metrics for a specific component.
//
// Unlike GetMetrics, this returns a copy that is safe to use without
// being affected by future updates to the component's metrics.
//
// Returns nil if the component has not been tracked.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshot := ct.GetMetricsSnapshot("comp-1")
//	// snapshot won't change even if more renders are recorded
func (ct *ComponentTracker) GetMetricsSnapshot(id string) *ComponentMetrics {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	metrics, ok := ct.components[id]
	if !ok {
		return nil
	}

	// Return a copy
	return &ComponentMetrics{
		ComponentID:     metrics.ComponentID,
		ComponentName:   metrics.ComponentName,
		RenderCount:     metrics.RenderCount,
		TotalRenderTime: metrics.TotalRenderTime,
		AvgRenderTime:   metrics.AvgRenderTime,
		MaxRenderTime:   metrics.MaxRenderTime,
		MinRenderTime:   metrics.MinRenderTime,
		MemoryUsage:     metrics.MemoryUsage,
	}
}

// GetAllMetrics returns metrics for all tracked components.
//
// Returns an empty map if no components have been tracked.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	all := ct.GetAllMetrics()
//	for id, metrics := range all {
//	    fmt.Printf("%s: %d renders\n", id, metrics.RenderCount)
//	}
func (ct *ComponentTracker) GetAllMetrics() map[string]*ComponentMetrics {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make(map[string]*ComponentMetrics, len(ct.components))
	for id, metrics := range ct.components {
		result[id] = metrics
	}
	return result
}

// GetComponentIDs returns the IDs of all tracked components.
//
// Returns an empty slice if no components have been tracked.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ids := ct.GetComponentIDs()
//	for _, id := range ids {
//	    metrics := ct.GetMetrics(id)
//	    // ...
//	}
func (ct *ComponentTracker) GetComponentIDs() []string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	ids := make([]string, 0, len(ct.components))
	for id := range ct.components {
		ids = append(ids, id)
	}
	return ids
}

// ComponentCount returns the number of tracked components.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := ct.ComponentCount()
//	fmt.Printf("Tracking %d components\n", count)
func (ct *ComponentTracker) ComponentCount() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	return len(ct.components)
}

// Reset clears all tracked component metrics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ct.Reset() // Clear all tracking data
func (ct *ComponentTracker) Reset() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.components = make(map[string]*ComponentMetrics)
}

// ResetComponent clears metrics for a specific component.
//
// Does nothing if the component is not being tracked.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ct.ResetComponent("comp-1") // Clear metrics for comp-1 only
func (ct *ComponentTracker) ResetComponent(id string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	delete(ct.components, id)
}

// RecordMemoryUsage updates the memory usage for a component.
//
// This method only updates memory usage for components that are already
// being tracked. If the component has not been tracked via RecordRender,
// this method does nothing.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ct.RecordMemoryUsage("comp-1", 1024) // 1KB
func (ct *ComponentTracker) RecordMemoryUsage(id string, bytes uint64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	metrics, ok := ct.components[id]
	if !ok {
		return
	}

	metrics.MemoryUsage = bytes
}

// GetTopComponents returns the top N components sorted by the specified field.
//
// The components are sorted in descending order (highest values first).
// If n exceeds the number of tracked components, all components are returned.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	// Get top 5 slowest components by total render time
//	top := ct.GetTopComponents(5, SortByTotalRenderTime)
//	for _, m := range top {
//	    fmt.Printf("%s: %v total\n", m.ComponentName, m.TotalRenderTime)
//	}
func (ct *ComponentTracker) GetTopComponents(n int, sortBy ComponentSortField) []*ComponentMetrics {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if len(ct.components) == 0 {
		return make([]*ComponentMetrics, 0)
	}

	// Create a slice of all metrics
	metrics := make([]*ComponentMetrics, 0, len(ct.components))
	for _, m := range ct.components {
		metrics = append(metrics, m)
	}

	// Sort by the specified field (descending)
	switch sortBy {
	case SortByTotalRenderTime:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].TotalRenderTime > metrics[j].TotalRenderTime
		})
	case SortByRenderCount:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].RenderCount > metrics[j].RenderCount
		})
	case SortByAvgRenderTime:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].AvgRenderTime > metrics[j].AvgRenderTime
		})
	case SortByMaxRenderTime:
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].MaxRenderTime > metrics[j].MaxRenderTime
		})
	}

	// Limit to n
	if n > len(metrics) {
		n = len(metrics)
	}

	return metrics[:n]
}

// TotalRenderCount returns the total number of renders across all components.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	total := ct.TotalRenderCount()
//	fmt.Printf("Total renders: %d\n", total)
func (ct *ComponentTracker) TotalRenderCount() int64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	var total int64
	for _, m := range ct.components {
		total += m.RenderCount
	}
	return total
}
