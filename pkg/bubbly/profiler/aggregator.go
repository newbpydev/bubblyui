// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"runtime"
	"sync"
	"time"
)

// DataAggregator aggregates profiling data from various sources for reporting.
//
// It collects data from MetricCollector, ComponentTracker, and RenderProfiler,
// combining them into a unified AggregatedData structure suitable for report
// generation and analysis.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	da := NewDataAggregator()
//	data := da.Aggregate(collector)
//	summary := da.CalculateSummary(data)
type DataAggregator struct {
	// mu protects concurrent access
	mu sync.RWMutex
}

// AggregatedData contains all aggregated profiling data.
//
// This struct combines data from multiple profiler components into a single
// structure for analysis and report generation.
type AggregatedData struct {
	// Timings contains aggregated timing statistics by operation name
	Timings map[string]*AggregatedTiming

	// Counters contains aggregated counter values by name
	Counters map[string]*AggregatedCounter

	// Allocations contains aggregated memory allocation statistics by location
	Allocations map[string]*AggregatedAllocation

	// Components contains per-component performance metrics
	Components []*ComponentMetrics

	// FrameCount is the total number of frames recorded
	FrameCount int64

	// AverageFPS is the average frames per second
	AverageFPS float64

	// DroppedFramePercent is the percentage of dropped frames
	DroppedFramePercent float64

	// MemorySnapshot contains the latest memory statistics
	MemorySnapshot *runtime.MemStats

	// GoroutineCount is the number of goroutines at aggregation time
	GoroutineCount int

	// Timestamp is when the data was aggregated
	Timestamp time.Time
}

// AggregatedTiming contains aggregated timing statistics for an operation.
type AggregatedTiming struct {
	// Name is the operation name
	Name string

	// Count is the number of times the operation was recorded
	Count int64

	// Total is the cumulative duration
	Total time.Duration

	// Min is the shortest duration
	Min time.Duration

	// Max is the longest duration
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

// AggregatedCounter contains aggregated counter values.
type AggregatedCounter struct {
	// Name is the counter name
	Name string

	// Count is the number of increments
	Count int64

	// Value is the current value (for gauge-style metrics)
	Value float64
}

// AggregatedAllocation contains aggregated memory allocation statistics.
type AggregatedAllocation struct {
	// Location is the allocation location identifier
	Location string

	// Count is the number of allocations
	Count int64

	// TotalSize is the cumulative bytes allocated
	TotalSize int64

	// AvgSize is the average allocation size
	AvgSize int64
}

// NewDataAggregator creates a new DataAggregator.
//
// Example:
//
//	da := NewDataAggregator()
//	data := da.Aggregate(collector)
func NewDataAggregator() *DataAggregator {
	return &DataAggregator{}
}

// Aggregate collects data from a MetricCollector and returns aggregated data.
//
// If collector is nil, returns an empty AggregatedData structure.
// The method extracts timing, counter, and allocation data from the collector.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	data := da.Aggregate(collector)
//	fmt.Printf("Timings: %d, Counters: %d\n", len(data.Timings), len(data.Counters))
func (da *DataAggregator) Aggregate(collector *MetricCollector) *AggregatedData {
	da.mu.Lock()
	defer da.mu.Unlock()

	data := &AggregatedData{
		Timings:     make(map[string]*AggregatedTiming),
		Counters:    make(map[string]*AggregatedCounter),
		Allocations: make(map[string]*AggregatedAllocation),
		Components:  make([]*ComponentMetrics, 0),
		Timestamp:   time.Now(),
	}

	if collector == nil {
		return data
	}

	// Aggregate timings
	da.aggregateTimings(data, collector.GetTimings())

	// Aggregate counters
	da.aggregateCounters(data, collector.GetCounters())

	// Aggregate memory allocations
	da.aggregateAllocations(data, collector.GetMemory())

	return data
}

// aggregateTimings extracts timing data from the TimingTracker.
func (da *DataAggregator) aggregateTimings(data *AggregatedData, tracker *TimingTracker) {
	if tracker == nil {
		return
	}

	// Get all operation names
	names := tracker.GetOperationNames()
	for _, name := range names {
		stats := tracker.GetStats(name)
		if stats == nil {
			continue
		}

		data.Timings[name] = &AggregatedTiming{
			Name:  name,
			Count: stats.Count,
			Total: stats.Total,
			Min:   stats.Min,
			Max:   stats.Max,
			Mean:  stats.Mean,
			P50:   stats.P50,
			P95:   stats.P95,
			P99:   stats.P99,
		}
	}
}

// aggregateCounters extracts counter data from the CounterTracker.
func (da *DataAggregator) aggregateCounters(data *AggregatedData, tracker *CounterTracker) {
	if tracker == nil {
		return
	}

	counters := tracker.GetAllCounters()
	for name, stats := range counters {
		data.Counters[name] = &AggregatedCounter{
			Name:  name,
			Count: stats.Count,
			Value: stats.Value,
		}
	}
}

// aggregateAllocations extracts allocation data from the MemoryTracker.
func (da *DataAggregator) aggregateAllocations(data *AggregatedData, tracker *MemoryTracker) {
	if tracker == nil {
		return
	}

	allocations := tracker.GetAllAllocations()
	for location, stats := range allocations {
		data.Allocations[location] = &AggregatedAllocation{
			Location:  location,
			Count:     stats.Count,
			TotalSize: stats.TotalSize,
			AvgSize:   stats.AvgSize,
		}
	}
}

// AggregateRenderData collects data from a RenderProfiler.
//
// Returns aggregated render performance data including FPS and dropped frames.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	renderData := da.AggregateRenderData(renderProfiler)
//	fmt.Printf("FPS: %.1f, Dropped: %.1f%%\n", renderData.AverageFPS, renderData.DroppedFramePercent)
func (da *DataAggregator) AggregateRenderData(profiler *RenderProfiler) *AggregatedData {
	da.mu.Lock()
	defer da.mu.Unlock()

	data := &AggregatedData{
		Timings:     make(map[string]*AggregatedTiming),
		Counters:    make(map[string]*AggregatedCounter),
		Allocations: make(map[string]*AggregatedAllocation),
		Components:  make([]*ComponentMetrics, 0),
		Timestamp:   time.Now(),
	}

	if profiler == nil {
		return data
	}

	data.FrameCount = int64(profiler.GetFrameCount())
	data.AverageFPS = profiler.GetFPS()
	data.DroppedFramePercent = profiler.GetDroppedFramePercent()

	return data
}

// AggregateComponentData collects data from a ComponentTracker.
//
// Returns aggregated component performance metrics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	compData := da.AggregateComponentData(componentTracker)
//	fmt.Printf("Components: %d\n", len(compData.Components))
func (da *DataAggregator) AggregateComponentData(tracker *ComponentTracker) *AggregatedData {
	da.mu.Lock()
	defer da.mu.Unlock()

	data := &AggregatedData{
		Timings:     make(map[string]*AggregatedTiming),
		Counters:    make(map[string]*AggregatedCounter),
		Allocations: make(map[string]*AggregatedAllocation),
		Components:  make([]*ComponentMetrics, 0),
		Timestamp:   time.Now(),
	}

	if tracker == nil {
		return data
	}

	metrics := tracker.GetAllMetrics()
	for _, m := range metrics {
		data.Components = append(data.Components, m)
	}

	return data
}

// AggregateAll collects data from all profiler components.
//
// This is a convenience method that combines data from MetricCollector,
// ComponentTracker, and RenderProfiler into a single AggregatedData structure.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	data := da.AggregateAll(collector, componentTracker, renderProfiler)
//	summary := da.CalculateSummary(data)
func (da *DataAggregator) AggregateAll(
	collector *MetricCollector,
	componentTracker *ComponentTracker,
	renderProfiler *RenderProfiler,
) *AggregatedData {
	da.mu.Lock()
	defer da.mu.Unlock()

	data := &AggregatedData{
		Timings:     make(map[string]*AggregatedTiming),
		Counters:    make(map[string]*AggregatedCounter),
		Allocations: make(map[string]*AggregatedAllocation),
		Components:  make([]*ComponentMetrics, 0),
		Timestamp:   time.Now(),
	}

	// Aggregate from MetricCollector
	if collector != nil {
		da.aggregateTimings(data, collector.GetTimings())
		da.aggregateCounters(data, collector.GetCounters())
		da.aggregateAllocations(data, collector.GetMemory())
	}

	// Aggregate from ComponentTracker
	if componentTracker != nil {
		metrics := componentTracker.GetAllMetrics()
		for _, m := range metrics {
			data.Components = append(data.Components, m)
		}
	}

	// Aggregate from RenderProfiler
	if renderProfiler != nil {
		data.FrameCount = int64(renderProfiler.GetFrameCount())
		data.AverageFPS = renderProfiler.GetFPS()
		data.DroppedFramePercent = renderProfiler.GetDroppedFramePercent()
	}

	// Add runtime information
	data.GoroutineCount = runtime.NumGoroutine()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	data.MemorySnapshot = &memStats

	return data
}

// CalculateSummary generates a Summary from AggregatedData.
//
// The summary includes total operations, memory usage, goroutine count,
// and average FPS. If data is nil, returns a summary with runtime info only.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	summary := da.CalculateSummary(data)
//	fmt.Printf("Operations: %d, Memory: %d bytes\n", summary.TotalOperations, summary.MemoryUsage)
func (da *DataAggregator) CalculateSummary(data *AggregatedData) *Summary {
	da.mu.RLock()
	defer da.mu.RUnlock()

	summary := &Summary{}

	// Get current runtime info
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	summary.MemoryUsage = memStats.HeapAlloc
	summary.GoroutineCount = runtime.NumGoroutine()

	if data == nil {
		return summary
	}

	// Calculate total operations from timings
	for _, timing := range data.Timings {
		summary.TotalOperations += timing.Count
	}

	// Get average FPS from counters or render data
	if fpsCounter, ok := data.Counters["fps"]; ok {
		summary.AverageFPS = fpsCounter.Value
	} else if data.AverageFPS > 0 {
		summary.AverageFPS = data.AverageFPS
	}

	// Use memory snapshot if available
	if data.MemorySnapshot != nil {
		summary.MemoryUsage = data.MemorySnapshot.HeapAlloc
	}

	// Use goroutine count from data if available
	if data.GoroutineCount > 0 {
		summary.GoroutineCount = data.GoroutineCount
	}

	return summary
}

// TakeMemorySnapshot captures current runtime memory statistics.
//
// Returns a MemProfileData with current heap allocation and object count.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshot := da.TakeMemorySnapshot()
//	fmt.Printf("Heap: %d bytes, Objects: %d\n", snapshot.HeapAlloc, snapshot.HeapObjects)
func (da *DataAggregator) TakeMemorySnapshot() *MemProfileData {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &MemProfileData{
		HeapAlloc:   memStats.HeapAlloc,
		HeapObjects: memStats.HeapObjects,
		GCPauses:    extractGCPauses(&memStats),
	}
}

// extractGCPauses extracts recent GC pause durations from MemStats.
func extractGCPauses(stats *runtime.MemStats) []time.Duration {
	pauses := make([]time.Duration, 0)
	if stats.NumGC == 0 {
		return pauses
	}

	// Get the last few GC pauses (up to 10)
	numPauses := int(stats.NumGC)
	if numPauses > 10 {
		numPauses = 10
	}

	for i := 0; i < numPauses; i++ {
		idx := (int(stats.NumGC) - 1 - i) % 256
		if stats.PauseNs[idx] > 0 {
			pauses = append(pauses, time.Duration(stats.PauseNs[idx]))
		}
	}

	return pauses
}

// GetGoroutineCount returns the current number of goroutines.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := da.GetGoroutineCount()
//	fmt.Printf("Goroutines: %d\n", count)
func (da *DataAggregator) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// TotalOperations returns the total number of operations recorded.
//
// This is a convenience method that sums all timing counts.
func (data *AggregatedData) TotalOperations() int64 {
	if data == nil {
		return 0
	}

	var total int64
	for _, timing := range data.Timings {
		total += timing.Count
	}
	return total
}

// TotalAllocatedMemory returns the total bytes allocated across all locations.
//
// This is a convenience method that sums all allocation sizes.
func (data *AggregatedData) TotalAllocatedMemory() int64 {
	if data == nil {
		return 0
	}

	var total int64
	for _, alloc := range data.Allocations {
		total += alloc.TotalSize
	}
	return total
}

// GetTiming returns aggregated timing for a specific operation.
//
// Returns nil if the operation was not recorded.
func (data *AggregatedData) GetTiming(name string) *AggregatedTiming {
	if data == nil || data.Timings == nil {
		return nil
	}
	return data.Timings[name]
}

// GetCounter returns aggregated counter for a specific name.
//
// Returns nil if the counter was not recorded.
func (data *AggregatedData) GetCounter(name string) *AggregatedCounter {
	if data == nil || data.Counters == nil {
		return nil
	}
	return data.Counters[name]
}

// GetAllocation returns aggregated allocation for a specific location.
//
// Returns nil if the location was not recorded.
func (data *AggregatedData) GetAllocation(location string) *AggregatedAllocation {
	if data == nil || data.Allocations == nil {
		return nil
	}
	return data.Allocations[location]
}
