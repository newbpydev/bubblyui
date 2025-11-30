// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
)

// MemoryProfiler handles memory profiling with pprof integration.
//
// It provides methods to take memory snapshots, write heap profiles,
// and detect memory growth. The profiler integrates with Go's pprof tools
// for detailed heap analysis.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	mp := NewMemoryProfiler()
//
//	// Take snapshots over time
//	mp.TakeSnapshot()
//	// ... run workload ...
//	mp.TakeSnapshot()
//
//	// Check for memory growth
//	growth := mp.GetMemoryGrowth()
//	if growth > 1024*1024 {
//	    fmt.Println("Warning: >1MB memory growth detected")
//	}
//
//	// Write heap profile for detailed analysis
//	mp.WriteHeapProfile("heap.prof")
//	// Analyze with: go tool pprof heap.prof
type MemoryProfiler struct {
	// baseline is the initial memory snapshot taken at creation
	baseline *runtime.MemStats

	// snapshots stores memory snapshots taken over time
	snapshots []*runtime.MemStats

	// mu protects concurrent access to profiler state
	mu sync.RWMutex
}

// NewMemoryProfiler creates a new memory profiler instance.
//
// The profiler captures a baseline memory snapshot at creation time.
// Use TakeSnapshot() to capture additional snapshots for comparison.
//
// Example:
//
//	mp := NewMemoryProfiler()
//	baseline := mp.GetBaseline()
//	fmt.Printf("Initial heap: %d bytes\n", baseline.HeapAlloc)
func NewMemoryProfiler() *MemoryProfiler {
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)

	return &MemoryProfiler{
		baseline:  &baseline,
		snapshots: make([]*runtime.MemStats, 0),
	}
}

// TakeSnapshot captures current runtime memory statistics.
//
// The snapshot is appended to the internal history and returned.
// Use GetMemoryGrowth() to compare snapshots over time.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshot := mp.TakeSnapshot()
//	fmt.Printf("HeapAlloc: %d bytes\n", snapshot.HeapAlloc)
//	fmt.Printf("HeapObjects: %d\n", snapshot.HeapObjects)
func (mp *MemoryProfiler) TakeSnapshot() *runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.snapshots = append(mp.snapshots, &stats)
	return &stats
}

// WriteHeapProfile writes a heap profile to the specified file.
//
// The profile is written in pprof format and can be analyzed with
// Go's pprof tools: go tool pprof <filename>
//
// Returns an error if the file cannot be created or written.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := mp.WriteHeapProfile("heap.prof")
//	if err != nil {
//	    log.Fatal("could not write heap profile:", err)
//	}
//	// Analyze with: go tool pprof heap.prof
func (mp *MemoryProfiler) WriteHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write heap profile using runtime/pprof
	if err := pprof.WriteHeapProfile(f); err != nil {
		return err
	}

	return nil
}

// GetMemoryGrowth returns the heap allocation growth between baseline and latest snapshot.
//
// Returns 0 if no snapshots have been taken.
// A positive value indicates memory growth, negative indicates shrinkage.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mp.TakeSnapshot()
//	// ... allocate memory ...
//	mp.TakeSnapshot()
//	growth := mp.GetMemoryGrowth()
//	if growth > 1024*1024 {
//	    fmt.Println("Warning: >1MB memory growth detected")
//	}
func (mp *MemoryProfiler) GetMemoryGrowth() int64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) < 2 {
		return 0
	}

	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]

	return int64(last.HeapAlloc) - int64(first.HeapAlloc)
}

// GetBaseline returns the baseline memory snapshot taken at creation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	baseline := mp.GetBaseline()
//	fmt.Printf("Initial heap: %d bytes\n", baseline.HeapAlloc)
func (mp *MemoryProfiler) GetBaseline() *runtime.MemStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.baseline
}

// GetSnapshots returns a copy of all captured memory snapshots.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshots := mp.GetSnapshots()
//	for i, s := range snapshots {
//	    fmt.Printf("Snapshot %d: HeapAlloc=%d\n", i, s.HeapAlloc)
//	}
func (mp *MemoryProfiler) GetSnapshots() []*runtime.MemStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	result := make([]*runtime.MemStats, len(mp.snapshots))
	copy(result, mp.snapshots)
	return result
}

// GetLatestSnapshot returns the most recent memory snapshot.
//
// Returns nil if no snapshots have been taken.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if latest := mp.GetLatestSnapshot(); latest != nil {
//	    fmt.Printf("Current heap: %d bytes\n", latest.HeapAlloc)
//	}
func (mp *MemoryProfiler) GetLatestSnapshot() *runtime.MemStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if len(mp.snapshots) == 0 {
		return nil
	}
	return mp.snapshots[len(mp.snapshots)-1]
}

// SnapshotCount returns the number of snapshots taken.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := mp.SnapshotCount()
//	fmt.Printf("Taken %d snapshots\n", count)
func (mp *MemoryProfiler) SnapshotCount() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.snapshots)
}

// Reset clears all snapshots and refreshes the baseline.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mp.Reset()
//	// Start fresh with new baseline
func (mp *MemoryProfiler) Reset() {
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)

	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.baseline = &baseline
	mp.snapshots = make([]*runtime.MemStats, 0)
}
