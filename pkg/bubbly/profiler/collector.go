// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricCollector collects performance metrics from the application.
//
// It coordinates collection of timing, memory, and counter metrics through
// specialized trackers. The collector can be enabled/disabled at runtime
// with minimal overhead when disabled.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	mc := NewMetricCollector()
//	mc.Enable()
//
//	// Measure function execution time
//	mc.Measure("render.component", func() {
//	    component.Render()
//	})
//
//	// Or use StartTiming for more control
//	stop := mc.StartTiming("update.cycle")
//	performUpdate()
//	stop()
type MetricCollector struct {
	// timings tracks operation timing statistics
	timings *TimingTracker

	// memory tracks memory allocation statistics
	memory *MemoryTracker

	// counters tracks generic counter metrics
	counters *CounterTracker

	// enabled indicates whether collection is active
	// Using atomic for fast read path
	enabled atomic.Bool

	// mu protects access to trackers during reset
	mu sync.RWMutex
}

// TimingTracker and TimingStats are defined in timing.go (Task 1.3)

// MemoryTracker tracks memory allocation statistics.
//
// This is the stub implementation for Task 1.2.
// Full implementation will be in Task 1.4.
type MemoryTracker struct {
	allocations map[string]*AllocationStats
	mu          sync.RWMutex
}

// AllocationStats holds statistics for a memory allocation location.
type AllocationStats struct {
	// Count is the number of allocations
	Count int64

	// TotalSize is the cumulative bytes allocated
	TotalSize int64

	// AvgSize is the average allocation size
	AvgSize int64
}

// CounterTracker tracks generic counter metrics.
//
// This is the stub implementation for Task 1.2.
type CounterTracker struct {
	counters map[string]*CounterStats
	mu       sync.RWMutex
}

// CounterStats holds statistics for a counter.
type CounterStats struct {
	// Count is the number of increments
	Count int64

	// Value is the current value (for gauge-style metrics)
	Value float64
}

// NewMetricCollector creates a new metric collector.
//
// The collector is created in a disabled state. Call Enable() to begin collection.
//
// Example:
//
//	mc := NewMetricCollector()
//	mc.Enable()
//	defer mc.Disable()
func NewMetricCollector() *MetricCollector {
	return &MetricCollector{
		timings:  newTimingTracker(),
		memory:   newMemoryTracker(),
		counters: newCounterTracker(),
	}
}

// newTimingTracker creates a new timing tracker.
func newTimingTracker() *TimingTracker {
	return NewTimingTracker()
}

// newMemoryTracker creates a new memory tracker.
func newMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		allocations: make(map[string]*AllocationStats),
	}
}

// newCounterTracker creates a new counter tracker.
func newCounterTracker() *CounterTracker {
	return &CounterTracker{
		counters: make(map[string]*CounterStats),
	}
}

// Enable activates metric collection.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mc *MetricCollector) Enable() {
	mc.enabled.Store(true)
}

// Disable deactivates metric collection.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mc *MetricCollector) Disable() {
	mc.enabled.Store(false)
}

// IsEnabled returns whether collection is active.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mc *MetricCollector) IsEnabled() bool {
	return mc.enabled.Load()
}

// Measure times the execution of a function and records the duration.
//
// The function is always executed, even when collection is disabled.
// Only the timing recording is skipped when disabled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mc.Measure("component.render", func() {
//	    component.Render()
//	})
func (mc *MetricCollector) Measure(name string, fn func()) {
	if !mc.enabled.Load() {
		fn()
		return
	}

	start := time.Now()
	fn()
	duration := time.Since(start)

	mc.timings.Record(name, duration)
}

// StartTiming begins timing an operation and returns a function to stop timing.
//
// This is useful when you need more control over when timing starts and stops.
// Returns a no-op function when collection is disabled.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	stop := mc.StartTiming("update.cycle")
//	performUpdate()
//	stop()
func (mc *MetricCollector) StartTiming(name string) func() {
	if !mc.enabled.Load() {
		return func() {}
	}

	start := time.Now()
	return func() {
		duration := time.Since(start)
		mc.timings.Record(name, duration)
	}
}

// RecordMetric records a generic float64 metric value.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mc.RecordMetric("fps", 60.0)
//	mc.RecordMetric("memory.heap", float64(runtime.MemStats.HeapAlloc))
func (mc *MetricCollector) RecordMetric(name string, value float64) {
	if !mc.enabled.Load() {
		return
	}

	mc.counters.RecordValue(name, value)
}

// IncrementCounter increments a counter by 1.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mc.IncrementCounter("render.count")
//	mc.IncrementCounter("event.click")
func (mc *MetricCollector) IncrementCounter(name string) {
	if !mc.enabled.Load() {
		return
	}

	mc.counters.Increment(name)
}

// RecordMemory records a memory allocation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mc.RecordMemory("component.state", int64(unsafe.Sizeof(state)))
func (mc *MetricCollector) RecordMemory(location string, size int64) {
	if !mc.enabled.Load() {
		return
	}

	mc.memory.TrackAllocation(location, size)
}

// Reset clears all collected metrics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mc *MetricCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.timings = newTimingTracker()
	mc.memory = newMemoryTracker()
	mc.counters = newCounterTracker()
}

// GetTimings returns the timing tracker for direct access.
//
// Thread Safety:
//
//	Safe to call concurrently, but returned tracker should be accessed carefully.
func (mc *MetricCollector) GetTimings() *TimingTracker {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.timings
}

// GetMemory returns the memory tracker for direct access.
//
// Thread Safety:
//
//	Safe to call concurrently, but returned tracker should be accessed carefully.
func (mc *MetricCollector) GetMemory() *MemoryTracker {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.memory
}

// GetCounters returns the counter tracker for direct access.
//
// Thread Safety:
//
//	Safe to call concurrently, but returned tracker should be accessed carefully.
func (mc *MetricCollector) GetCounters() *CounterTracker {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.counters
}

// TimingTracker methods are defined in timing.go (Task 1.3)

// --- MemoryTracker methods ---

// TrackAllocation records a memory allocation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) TrackAllocation(location string, size int64) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	stats, ok := mt.allocations[location]
	if !ok {
		stats = &AllocationStats{}
		mt.allocations[location] = stats
	}

	stats.Count++
	stats.TotalSize += size
	stats.AvgSize = stats.TotalSize / stats.Count
}

// GetAllocation returns statistics for an allocation location.
//
// Returns nil if the location has not been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetAllocation(location string) *AllocationStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return mt.allocations[location]
}

// GetAllAllocations returns statistics for all allocation locations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetAllAllocations() map[string]*AllocationStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	result := make(map[string]*AllocationStats, len(mt.allocations))
	for k, v := range mt.allocations {
		result[k] = v
	}
	return result
}

// --- CounterTracker methods ---

// Increment increments a counter by 1.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ct *CounterTracker) Increment(name string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	stats, ok := ct.counters[name]
	if !ok {
		stats = &CounterStats{}
		ct.counters[name] = stats
	}

	stats.Count++
}

// RecordValue records a float64 value for a counter.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ct *CounterTracker) RecordValue(name string, value float64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	stats, ok := ct.counters[name]
	if !ok {
		stats = &CounterStats{}
		ct.counters[name] = stats
	}

	stats.Count++
	stats.Value = value
}

// GetCounter returns statistics for a counter.
//
// Returns nil if the counter has not been recorded.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ct *CounterTracker) GetCounter(name string) *CounterStats {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	return ct.counters[name]
}

// GetAllCounters returns statistics for all counters.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ct *CounterTracker) GetAllCounters() map[string]*CounterStats {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make(map[string]*CounterStats, len(ct.counters))
	for k, v := range ct.counters {
		result[k] = v
	}
	return result
}
