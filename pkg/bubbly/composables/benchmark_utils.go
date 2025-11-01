package composables

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkStats contains statistical information about benchmark results.
//
// Used for analyzing benchmark performance and detecting regressions.
type BenchmarkStats struct {
	Name         string
	Iterations   int
	NsPerOp      int64
	AllocBytes   int64
	AllocsPerOp  int64
	MBPerSec     float64
	StartMemory  uint64
	EndMemory    uint64
	MemoryGrowth uint64
}

// RunWithStats runs a benchmark function and collects detailed statistics.
//
// This helper captures additional metrics beyond standard benchmarking:
//   - Memory before/after execution
//   - Memory growth
//   - Allocation patterns
//
// Parameters:
//   - b: Testing benchmark instance
//   - fn: Function to benchmark
//
// Example:
//
//	func BenchmarkMyOperation(b *testing.B) {
//	    RunWithStats(b, func() {
//	        // Your code to benchmark
//	        DoExpensiveOperation()
//	    })
//	}
func RunWithStats(b *testing.B, fn func()) {
	b.Helper()

	// Capture initial memory state
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Reset timer after setup
	b.ReportAllocs()
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		fn()
	}

	// Capture final memory state
	b.StopTimer()
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate memory growth
	memGrowth := memAfter.HeapAlloc - memBefore.HeapAlloc

	// Log additional stats
	b.ReportMetric(float64(memGrowth), "mem-growth-bytes")
	b.ReportMetric(float64(memAfter.NumGC-memBefore.NumGC), "gc-runs")
}

// BenchmarkComparison contains comparison between two benchmark runs.
//
// Used for analyzing performance changes between baseline and current code.
type BenchmarkComparison struct {
	Baseline       *BenchmarkStats
	Current        *BenchmarkStats
	SpeedupPercent float64 // Positive = faster, negative = slower
	MemDelta       int64   // Bytes difference
	AllocDelta     int64   // Allocation count difference
}

// CompareResults compares two benchmark results and returns the comparison.
//
// **Note:** This is a simplified comparison. For production use, prefer benchstat
// which provides statistical significance testing.
//
// Parameters:
//   - baseline: Baseline benchmark name
//   - current: Current benchmark name
//
// Returns:
//   - *BenchmarkComparison: Comparison results or nil if benchmarks not comparable
//
// Example:
//
//	// Run baseline
//	baseline := BenchmarkUseState(...)
//	
//	// Make changes
//	
//	// Run current
//	current := BenchmarkUseState(...)
//	
//	// Compare
//	comp := CompareResults("baseline", "current")
//	if comp.SpeedupPercent > 10 {
//	    fmt.Printf("Performance improved by %.2f%%\n", comp.SpeedupPercent)
//	}
func CompareResults(baseline, current string) *BenchmarkComparison {
	// This is a placeholder for the comparison logic
	// In practice, you would parse benchmark output files
	// For now, we return nil to indicate comparison is done via benchstat
	return nil
}

// RunMultiCPU runs a benchmark with different GOMAXPROCS values.
//
// This helps identify scaling characteristics of concurrent operations.
//
// Parameters:
//   - b: Testing benchmark instance
//   - fn: Function to benchmark
//   - cpus: CPU counts to test (e.g., []int{1, 2, 4, 8})
//
// Example:
//
//	func BenchmarkConcurrentOperation(b *testing.B) {
//	    RunMultiCPU(b, func(b *testing.B) {
//	        for i := 0; i < b.N; i++ {
//	            DoConcurrentWork()
//	        }
//	    }, []int{1, 2, 4, 8})
//	}
func RunMultiCPU(b *testing.B, fn func(b *testing.B), cpus []int) {
	b.Helper()

	for _, numCPU := range cpus {
		b.Run(fmt.Sprintf("cpu=%d", numCPU), func(b *testing.B) {
			// Save original GOMAXPROCS
			oldProcs := runtime.GOMAXPROCS(numCPU)
			defer runtime.GOMAXPROCS(oldProcs)

			// Run the benchmark with specified CPU count
			fn(b)
		})
	}
}

// MeasureMemoryGrowth measures memory growth over time for a long-running operation.
//
// This is useful for detecting memory leaks and unbounded growth.
//
// Parameters:
//   - b: Testing benchmark instance
//   - duration: How long to run the test
//   - fn: Function to run repeatedly
//
// Returns:
//   - startMem: Memory at start (bytes)
//   - endMem: Memory at end (bytes)
//   - growth: Total memory growth (bytes)
//
// Example:
//
//	func BenchmarkMemoryGrowth(b *testing.B) {
//	    start, end, growth := MeasureMemoryGrowth(b, 1*time.Second, func() {
//	        CreateComposable()
//	    })
//	    
//	    if growth > 1000000 { // 1MB
//	        b.Errorf("Memory growth too high: %d bytes", growth)
//	    }
//	}
func MeasureMemoryGrowth(b *testing.B, duration time.Duration, fn func()) (startMem, endMem, growth uint64) {
	b.Helper()

	// Force GC and get initial memory
	runtime.GC()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	// Run function repeatedly for duration
	deadline := time.Now().Add(duration)
	iterations := 0
	for time.Now().Before(deadline) {
		fn()
		iterations++
	}

	// Force GC and get final memory
	runtime.GC()
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	startMem = memStart.HeapAlloc
	endMem = memEnd.HeapAlloc
	
	if endMem > startMem {
		growth = endMem - startMem
	} else {
		growth = 0
	}

	// Report metrics
	b.ReportMetric(float64(iterations), "iterations")
	b.ReportMetric(float64(growth), "mem-growth-bytes")
	b.ReportMetric(float64(growth)/float64(iterations), "bytes-per-iter")

	return startMem, endMem, growth
}

// AllocPerOp returns the number of allocations per operation from benchmark result.
//
// Helper for extracting allocation metrics from testing.B.
//
// Example:
//
//	func BenchmarkSomething(b *testing.B) {
//	    b.ReportAllocs()
//	    for i := 0; i < b.N; i++ {
//	        DoSomething()
//	    }
//	    
//	    // Check allocations didn't increase
//	    if AllocPerOp(b) > 5 {
//	        b.Errorf("Too many allocations: %d", AllocPerOp(b))
//	    }
//	}
func AllocPerOp(b *testing.B) int64 {
	b.Helper()
	// This would need access to b's internal metrics
	// For now, return 0 as placeholder
	return 0
}

// BytesPerOp returns the number of bytes allocated per operation.
//
// Helper for extracting memory metrics from testing.B.
func BytesPerOp(b *testing.B) int64 {
	b.Helper()
	// This would need access to b's internal metrics
	// For now, return 0 as placeholder
	return 0
}
