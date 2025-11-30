// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDataAggregator tests DataAggregator creation.
func TestNewDataAggregator(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates new aggregator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := NewDataAggregator()
			require.NotNil(t, da)
		})
	}
}

// TestDataAggregator_Aggregate tests data aggregation from MetricCollector.
func TestDataAggregator_Aggregate(t *testing.T) {
	tests := []struct {
		name            string
		setupCollector  func() *MetricCollector
		wantTimings     int
		wantCounters    int
		wantAllocations int
	}{
		{
			name: "nil collector returns empty data",
			setupCollector: func() *MetricCollector {
				return nil
			},
			wantTimings:     0,
			wantCounters:    0,
			wantAllocations: 0,
		},
		{
			name: "empty collector returns empty data",
			setupCollector: func() *MetricCollector {
				mc := NewMetricCollector()
				mc.Enable()
				return mc
			},
			wantTimings:     0,
			wantCounters:    0,
			wantAllocations: 0,
		},
		{
			name: "collector with timings aggregates correctly",
			setupCollector: func() *MetricCollector {
				mc := NewMetricCollector()
				mc.Enable()
				mc.Measure("render", func() {
					time.Sleep(1 * time.Millisecond)
				})
				mc.Measure("render", func() {
					time.Sleep(2 * time.Millisecond)
				})
				mc.Measure("update", func() {
					time.Sleep(1 * time.Millisecond)
				})
				return mc
			},
			wantTimings:     2, // render and update
			wantCounters:    0,
			wantAllocations: 0,
		},
		{
			name: "collector with counters aggregates correctly",
			setupCollector: func() *MetricCollector {
				mc := NewMetricCollector()
				mc.Enable()
				mc.IncrementCounter("events")
				mc.IncrementCounter("events")
				mc.IncrementCounter("clicks")
				mc.RecordMetric("fps", 60.0)
				return mc
			},
			wantTimings:     0,
			wantCounters:    3, // events, clicks, fps
			wantAllocations: 0,
		},
		{
			name: "collector with memory allocations aggregates correctly",
			setupCollector: func() *MetricCollector {
				mc := NewMetricCollector()
				mc.Enable()
				mc.RecordMemory("component.state", 1024)
				mc.RecordMemory("component.state", 2048)
				mc.RecordMemory("buffer", 512)
				return mc
			},
			wantTimings:     0,
			wantCounters:    0,
			wantAllocations: 2, // component.state and buffer
		},
		{
			name: "collector with all metrics aggregates correctly",
			setupCollector: func() *MetricCollector {
				mc := NewMetricCollector()
				mc.Enable()
				mc.Measure("render", func() {})
				mc.IncrementCounter("events")
				mc.RecordMemory("state", 1024)
				return mc
			},
			wantTimings:     1,
			wantCounters:    1,
			wantAllocations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := NewDataAggregator()
			collector := tt.setupCollector()

			data := da.Aggregate(collector)

			require.NotNil(t, data)
			assert.Len(t, data.Timings, tt.wantTimings)
			assert.Len(t, data.Counters, tt.wantCounters)
			assert.Len(t, data.Allocations, tt.wantAllocations)
		})
	}
}

// TestDataAggregator_CalculateSummary tests summary calculation.
func TestDataAggregator_CalculateSummary(t *testing.T) {
	tests := []struct {
		name            string
		setupData       func() *AggregatedData
		wantOperations  int64
		wantMemoryUsage bool // just check if > 0
		wantGoroutines  bool // just check if > 0
	}{
		{
			name: "nil data returns empty summary",
			setupData: func() *AggregatedData {
				return nil
			},
			wantOperations:  0,
			wantMemoryUsage: false,
			wantGoroutines:  false,
		},
		{
			name: "empty data returns summary with runtime info",
			setupData: func() *AggregatedData {
				return &AggregatedData{
					Timings:     make(map[string]*AggregatedTiming),
					Counters:    make(map[string]*AggregatedCounter),
					Allocations: make(map[string]*AggregatedAllocation),
				}
			},
			wantOperations:  0,
			wantMemoryUsage: true, // runtime.MemStats should have data
			wantGoroutines:  true, // runtime.NumGoroutine() should be > 0
		},
		{
			name: "data with timings calculates total operations",
			setupData: func() *AggregatedData {
				return &AggregatedData{
					Timings: map[string]*AggregatedTiming{
						"render": {Count: 100, Total: 500 * time.Millisecond},
						"update": {Count: 50, Total: 100 * time.Millisecond},
					},
					Counters:    make(map[string]*AggregatedCounter),
					Allocations: make(map[string]*AggregatedAllocation),
				}
			},
			wantOperations:  150, // 100 + 50
			wantMemoryUsage: true,
			wantGoroutines:  true,
		},
		{
			name: "data with FPS counter includes average FPS",
			setupData: func() *AggregatedData {
				return &AggregatedData{
					Timings: make(map[string]*AggregatedTiming),
					Counters: map[string]*AggregatedCounter{
						"fps": {Count: 60, Value: 58.5},
					},
					Allocations: make(map[string]*AggregatedAllocation),
				}
			},
			wantOperations:  0,
			wantMemoryUsage: true,
			wantGoroutines:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := NewDataAggregator()
			data := tt.setupData()

			summary := da.CalculateSummary(data)

			require.NotNil(t, summary)
			assert.Equal(t, tt.wantOperations, summary.TotalOperations)
			if tt.wantMemoryUsage {
				assert.Greater(t, summary.MemoryUsage, uint64(0))
			}
			if tt.wantGoroutines {
				assert.Greater(t, summary.GoroutineCount, 0)
			}
		})
	}
}

// TestDataAggregator_AggregateTimings tests timing aggregation details.
func TestDataAggregator_AggregateTimings(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	mc.Enable()

	// Record multiple timings
	for i := 0; i < 10; i++ {
		mc.Measure("render", func() {
			time.Sleep(time.Duration(i) * time.Millisecond)
		})
	}

	data := da.Aggregate(mc)

	require.NotNil(t, data)
	require.Contains(t, data.Timings, "render")

	timing := data.Timings["render"]
	assert.Equal(t, int64(10), timing.Count)
	assert.Greater(t, timing.Total, time.Duration(0))
	assert.GreaterOrEqual(t, timing.Max, timing.Min)
	assert.Greater(t, timing.Mean, time.Duration(0))
}

// TestDataAggregator_AggregateCounters tests counter aggregation details.
func TestDataAggregator_AggregateCounters(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	mc.Enable()

	// Record multiple counter increments
	for i := 0; i < 5; i++ {
		mc.IncrementCounter("events")
	}
	mc.RecordMetric("fps", 60.0)

	data := da.Aggregate(mc)

	require.NotNil(t, data)
	require.Contains(t, data.Counters, "events")
	require.Contains(t, data.Counters, "fps")

	events := data.Counters["events"]
	assert.Equal(t, int64(5), events.Count)

	fps := data.Counters["fps"]
	assert.Equal(t, 60.0, fps.Value)
}

// TestDataAggregator_AggregateAllocations tests allocation aggregation details.
func TestDataAggregator_AggregateAllocations(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	mc.Enable()

	// Record multiple allocations
	mc.RecordMemory("component.state", 1024)
	mc.RecordMemory("component.state", 2048)
	mc.RecordMemory("buffer", 512)

	data := da.Aggregate(mc)

	require.NotNil(t, data)
	require.Contains(t, data.Allocations, "component.state")
	require.Contains(t, data.Allocations, "buffer")

	state := data.Allocations["component.state"]
	assert.Equal(t, int64(2), state.Count)
	assert.Equal(t, int64(3072), state.TotalSize) // 1024 + 2048
	assert.Equal(t, int64(1536), state.AvgSize)   // 3072 / 2

	buffer := data.Allocations["buffer"]
	assert.Equal(t, int64(1), buffer.Count)
	assert.Equal(t, int64(512), buffer.TotalSize)
}

// TestDataAggregator_ThreadSafety tests concurrent access.
func TestDataAggregator_ThreadSafety(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	mc.Enable()

	// Pre-populate some data
	mc.Measure("render", func() {})
	mc.IncrementCounter("events")
	mc.RecordMemory("state", 1024)

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent aggregation
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := da.Aggregate(mc)
			_ = da.CalculateSummary(data)
		}()
	}

	wg.Wait()
	// Test passes if no race conditions detected
}

// TestDataAggregator_WithRenderProfiler tests aggregation with render profiler data.
func TestDataAggregator_WithRenderProfiler(t *testing.T) {
	da := NewDataAggregator()
	rp := NewRenderProfiler()

	// Record some frames
	for i := 0; i < 60; i++ {
		rp.RecordFrame(time.Duration(i) * time.Millisecond)
	}

	data := da.AggregateRenderData(rp)

	require.NotNil(t, data)
	assert.Greater(t, data.FrameCount, int64(0))
	assert.GreaterOrEqual(t, data.DroppedFramePercent, 0.0)
}

// TestDataAggregator_WithComponentTracker tests aggregation with component tracker data.
func TestDataAggregator_WithComponentTracker(t *testing.T) {
	da := NewDataAggregator()
	ct := NewComponentTracker()

	// Record some component renders
	ct.RecordRender("comp1", "Counter", 5*time.Millisecond)
	ct.RecordRender("comp1", "Counter", 10*time.Millisecond)
	ct.RecordRender("comp2", "Button", 2*time.Millisecond)

	data := da.AggregateComponentData(ct)

	require.NotNil(t, data)
	assert.Len(t, data.Components, 2)
}

// TestDataAggregator_FullIntegration tests full aggregation workflow.
func TestDataAggregator_FullIntegration(t *testing.T) {
	da := NewDataAggregator()

	// Create and populate all data sources
	mc := NewMetricCollector()
	mc.Enable()
	mc.Measure("render", func() { time.Sleep(1 * time.Millisecond) })
	mc.IncrementCounter("events")
	mc.RecordMemory("state", 1024)

	ct := NewComponentTracker()
	ct.RecordRender("comp1", "Counter", 5*time.Millisecond)

	rp := NewRenderProfiler()
	rp.RecordFrame(16 * time.Millisecond)

	// Aggregate all data
	data := da.AggregateAll(mc, ct, rp)

	require.NotNil(t, data)
	assert.Greater(t, len(data.Timings), 0)
	assert.Greater(t, len(data.Counters), 0)
	assert.Greater(t, len(data.Allocations), 0)
	assert.Greater(t, len(data.Components), 0)
	assert.Greater(t, data.FrameCount, int64(0))

	// Calculate summary
	summary := da.CalculateSummary(data)
	require.NotNil(t, summary)
	assert.Greater(t, summary.TotalOperations, int64(0))
}

// TestAggregatedData_Statistics tests statistical calculations.
func TestAggregatedData_Statistics(t *testing.T) {
	tests := []struct {
		name    string
		data    *AggregatedData
		wantOps int64
		wantMem int64
	}{
		{
			name: "calculates total operations from timings",
			data: &AggregatedData{
				Timings: map[string]*AggregatedTiming{
					"a": {Count: 10},
					"b": {Count: 20},
					"c": {Count: 30},
				},
				Counters:    make(map[string]*AggregatedCounter),
				Allocations: make(map[string]*AggregatedAllocation),
			},
			wantOps: 60,
		},
		{
			name: "calculates total allocated memory",
			data: &AggregatedData{
				Timings:  make(map[string]*AggregatedTiming),
				Counters: make(map[string]*AggregatedCounter),
				Allocations: map[string]*AggregatedAllocation{
					"a": {TotalSize: 1024},
					"b": {TotalSize: 2048},
				},
			},
			wantMem: 3072,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantOps > 0 {
				ops := tt.data.TotalOperations()
				assert.Equal(t, tt.wantOps, ops)
			}
			if tt.wantMem > 0 {
				mem := tt.data.TotalAllocatedMemory()
				assert.Equal(t, tt.wantMem, mem)
			}
		})
	}
}

// TestDataAggregator_Reset tests reset functionality.
func TestDataAggregator_Reset(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	mc.Enable()
	mc.Measure("render", func() {})

	// First aggregation
	data1 := da.Aggregate(mc)
	require.NotNil(t, data1)
	assert.Len(t, data1.Timings, 1)

	// Reset collector
	mc.Reset()

	// Second aggregation should be empty
	data2 := da.Aggregate(mc)
	require.NotNil(t, data2)
	assert.Len(t, data2.Timings, 0)
}

// TestDataAggregator_DisabledCollector tests aggregation with disabled collector.
func TestDataAggregator_DisabledCollector(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	// Collector is disabled by default

	mc.Measure("render", func() {})
	mc.IncrementCounter("events")
	mc.RecordMemory("state", 1024)

	data := da.Aggregate(mc)

	require.NotNil(t, data)
	// All should be empty since collector was disabled
	assert.Len(t, data.Timings, 0)
	assert.Len(t, data.Counters, 0)
	assert.Len(t, data.Allocations, 0)
}

// TestDataAggregator_MemorySnapshot tests memory snapshot aggregation.
func TestDataAggregator_MemorySnapshot(t *testing.T) {
	da := NewDataAggregator()

	snapshot := da.TakeMemorySnapshot()

	require.NotNil(t, snapshot)
	assert.Greater(t, snapshot.HeapAlloc, uint64(0))
	assert.Greater(t, snapshot.HeapObjects, uint64(0))
}

// TestDataAggregator_GoroutineCount tests goroutine count retrieval.
func TestDataAggregator_GoroutineCount(t *testing.T) {
	da := NewDataAggregator()

	count := da.GetGoroutineCount()

	assert.Greater(t, count, 0)
	assert.Equal(t, runtime.NumGoroutine(), count)
}

// TestDataAggregator_PerformanceAcceptable tests aggregation performance.
func TestDataAggregator_PerformanceAcceptable(t *testing.T) {
	da := NewDataAggregator()
	mc := NewMetricCollector()
	mc.Enable()

	// Add substantial data
	for i := 0; i < 1000; i++ {
		mc.Measure("render", func() {})
		mc.IncrementCounter("events")
		mc.RecordMemory("state", 1024)
	}

	start := time.Now()
	data := da.Aggregate(mc)
	_ = da.CalculateSummary(data)
	duration := time.Since(start)

	// Aggregation should complete in reasonable time (< 100ms)
	assert.Less(t, duration, 100*time.Millisecond)
}

// TestAggregatedData_GetTiming tests GetTiming method.
func TestAggregatedData_GetTiming(t *testing.T) {
	tests := []struct {
		name    string
		data    *AggregatedData
		opName  string
		wantNil bool
	}{
		{
			name:    "nil data returns nil",
			data:    nil,
			opName:  "render",
			wantNil: true,
		},
		{
			name: "nil timings map returns nil",
			data: &AggregatedData{
				Timings: nil,
			},
			opName:  "render",
			wantNil: true,
		},
		{
			name: "missing operation returns nil",
			data: &AggregatedData{
				Timings: map[string]*AggregatedTiming{
					"update": {Name: "update", Count: 10},
				},
			},
			opName:  "render",
			wantNil: true,
		},
		{
			name: "existing operation returns timing",
			data: &AggregatedData{
				Timings: map[string]*AggregatedTiming{
					"render": {Name: "render", Count: 100},
				},
			},
			opName:  "render",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetTiming(tt.opName)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.opName, result.Name)
			}
		})
	}
}

// TestAggregatedData_GetCounter tests GetCounter method.
func TestAggregatedData_GetCounter(t *testing.T) {
	tests := []struct {
		name        string
		data        *AggregatedData
		counterName string
		wantNil     bool
	}{
		{
			name:        "nil data returns nil",
			data:        nil,
			counterName: "events",
			wantNil:     true,
		},
		{
			name: "nil counters map returns nil",
			data: &AggregatedData{
				Counters: nil,
			},
			counterName: "events",
			wantNil:     true,
		},
		{
			name: "missing counter returns nil",
			data: &AggregatedData{
				Counters: map[string]*AggregatedCounter{
					"clicks": {Name: "clicks", Count: 10},
				},
			},
			counterName: "events",
			wantNil:     true,
		},
		{
			name: "existing counter returns value",
			data: &AggregatedData{
				Counters: map[string]*AggregatedCounter{
					"events": {Name: "events", Count: 50, Value: 60.0},
				},
			},
			counterName: "events",
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetCounter(tt.counterName)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.counterName, result.Name)
			}
		})
	}
}

// TestAggregatedData_GetAllocation tests GetAllocation method.
func TestAggregatedData_GetAllocation(t *testing.T) {
	tests := []struct {
		name     string
		data     *AggregatedData
		location string
		wantNil  bool
	}{
		{
			name:     "nil data returns nil",
			data:     nil,
			location: "state",
			wantNil:  true,
		},
		{
			name: "nil allocations map returns nil",
			data: &AggregatedData{
				Allocations: nil,
			},
			location: "state",
			wantNil:  true,
		},
		{
			name: "missing location returns nil",
			data: &AggregatedData{
				Allocations: map[string]*AggregatedAllocation{
					"buffer": {Location: "buffer", Count: 10},
				},
			},
			location: "state",
			wantNil:  true,
		},
		{
			name: "existing location returns allocation",
			data: &AggregatedData{
				Allocations: map[string]*AggregatedAllocation{
					"state": {Location: "state", Count: 5, TotalSize: 1024},
				},
			},
			location: "state",
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetAllocation(tt.location)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.location, result.Location)
			}
		})
	}
}

// TestAggregatedData_NilData tests nil data handling for statistics methods.
func TestAggregatedData_NilData(t *testing.T) {
	var data *AggregatedData = nil

	assert.Equal(t, int64(0), data.TotalOperations())
	assert.Equal(t, int64(0), data.TotalAllocatedMemory())
	assert.Nil(t, data.GetTiming("render"))
	assert.Nil(t, data.GetCounter("events"))
	assert.Nil(t, data.GetAllocation("state"))
}

// TestDataAggregator_AggregateRenderData_NilProfiler tests nil render profiler handling.
func TestDataAggregator_AggregateRenderData_NilProfiler(t *testing.T) {
	da := NewDataAggregator()

	data := da.AggregateRenderData(nil)

	require.NotNil(t, data)
	assert.Equal(t, int64(0), data.FrameCount)
	assert.Equal(t, 0.0, data.AverageFPS)
	assert.Equal(t, 0.0, data.DroppedFramePercent)
}

// TestDataAggregator_AggregateComponentData_NilTracker tests nil component tracker handling.
func TestDataAggregator_AggregateComponentData_NilTracker(t *testing.T) {
	da := NewDataAggregator()

	data := da.AggregateComponentData(nil)

	require.NotNil(t, data)
	assert.Len(t, data.Components, 0)
}

// TestDataAggregator_AggregateAll_NilInputs tests AggregateAll with nil inputs.
func TestDataAggregator_AggregateAll_NilInputs(t *testing.T) {
	da := NewDataAggregator()

	data := da.AggregateAll(nil, nil, nil)

	require.NotNil(t, data)
	assert.Len(t, data.Timings, 0)
	assert.Len(t, data.Counters, 0)
	assert.Len(t, data.Allocations, 0)
	assert.Len(t, data.Components, 0)
	assert.Greater(t, data.GoroutineCount, 0)
	assert.NotNil(t, data.MemorySnapshot)
}

// TestDataAggregator_ExtractGCPauses tests GC pause extraction.
func TestDataAggregator_ExtractGCPauses(t *testing.T) {
	// Test with zero GC cycles
	stats := &runtime.MemStats{
		NumGC: 0,
	}
	pauses := extractGCPauses(stats)
	assert.Len(t, pauses, 0)

	// Test with some GC cycles
	stats2 := &runtime.MemStats{
		NumGC: 5,
	}
	// Set some pause values
	stats2.PauseNs[0] = 1000000 // 1ms
	stats2.PauseNs[1] = 2000000 // 2ms
	stats2.PauseNs[2] = 3000000 // 3ms
	stats2.PauseNs[3] = 4000000 // 4ms
	stats2.PauseNs[4] = 5000000 // 5ms

	pauses2 := extractGCPauses(stats2)
	assert.Greater(t, len(pauses2), 0)
}

// TestDataAggregator_AggregateTimings_NilTracker tests nil timing tracker handling.
func TestDataAggregator_AggregateTimings_NilTracker(t *testing.T) {
	da := NewDataAggregator()
	data := &AggregatedData{
		Timings: make(map[string]*AggregatedTiming),
	}

	// This should not panic
	da.aggregateTimings(data, nil)

	assert.Len(t, data.Timings, 0)
}

// TestDataAggregator_AggregateCounters_NilTracker tests nil counter tracker handling.
func TestDataAggregator_AggregateCounters_NilTracker(t *testing.T) {
	da := NewDataAggregator()
	data := &AggregatedData{
		Counters: make(map[string]*AggregatedCounter),
	}

	// This should not panic
	da.aggregateCounters(data, nil)

	assert.Len(t, data.Counters, 0)
}

// TestDataAggregator_AggregateAllocations_NilTracker tests nil memory tracker handling.
func TestDataAggregator_AggregateAllocations_NilTracker(t *testing.T) {
	da := NewDataAggregator()
	data := &AggregatedData{
		Allocations: make(map[string]*AggregatedAllocation),
	}

	// This should not panic
	da.aggregateAllocations(data, nil)

	assert.Len(t, data.Allocations, 0)
}
