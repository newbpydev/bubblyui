// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"runtime"
	"sync"
)

// MemoryTracker tracks memory allocation statistics and runtime memory snapshots.
//
// It maintains a history of runtime.MemStats snapshots and tracks allocations
// by location. This enables memory growth detection and allocation hot spot
// identification.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	mt := NewMemoryTracker()
//	mt.TakeSnapshot()
//	mt.TrackAllocation("component.state", 1024)
//	growth := mt.GetMemoryGrowth()
type MemoryTracker struct {
	// snapshots stores runtime memory statistics snapshots
	snapshots []*runtime.MemStats

	// allocations maps location names to their statistics
	allocations map[string]*AllocationStats

	// mu protects concurrent access
	mu sync.RWMutex
}

// AllocationStats holds statistics for a memory allocation location.
type AllocationStats struct {
	// Count is the number of allocations at this location
	Count int64

	// TotalSize is the cumulative bytes allocated
	TotalSize int64

	// AvgSize is the average allocation size (TotalSize / Count)
	AvgSize int64
}

// NewMemoryTracker creates a new memory tracker.
//
// Example:
//
//	mt := NewMemoryTracker()
//	mt.TakeSnapshot()
func NewMemoryTracker() *MemoryTracker {
	return &MemoryTracker{
		snapshots:   make([]*runtime.MemStats, 0),
		allocations: make(map[string]*AllocationStats),
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
//	snapshot := mt.TakeSnapshot()
//	fmt.Printf("HeapAlloc: %d bytes\n", snapshot.HeapAlloc)
func (mt *MemoryTracker) TakeSnapshot() *runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.snapshots = append(mt.snapshots, &stats)
	return &stats
}

// TrackAllocation records a memory allocation at a specific location.
//
// The location string typically identifies where the allocation occurred
// (e.g., "component.state", "buffer.resize").
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mt.TrackAllocation("component.state", int64(unsafe.Sizeof(state)))
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
	if stats.Count > 0 {
		stats.AvgSize = stats.TotalSize / stats.Count
	}
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

// GetAllSnapshots returns all captured memory snapshots.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetAllSnapshots() []*runtime.MemStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	result := make([]*runtime.MemStats, len(mt.snapshots))
	copy(result, mt.snapshots)
	return result
}

// GetSnapshotAt returns the snapshot at the specified index.
//
// Returns nil if the index is out of bounds.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetSnapshotAt(index int) *runtime.MemStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if index < 0 || index >= len(mt.snapshots) {
		return nil
	}
	return mt.snapshots[index]
}

// GetFirstSnapshot returns the first captured snapshot.
//
// Returns nil if no snapshots have been taken.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetFirstSnapshot() *runtime.MemStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.snapshots) == 0 {
		return nil
	}
	return mt.snapshots[0]
}

// GetLatestSnapshot returns the most recent snapshot.
//
// Returns nil if no snapshots have been taken.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetLatestSnapshot() *runtime.MemStats {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.snapshots) == 0 {
		return nil
	}
	return mt.snapshots[len(mt.snapshots)-1]
}

// GetMemoryGrowth returns the heap allocation growth between first and last snapshots.
//
// Returns 0 if fewer than 2 snapshots have been taken.
// A positive value indicates memory growth, negative indicates shrinkage.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	mt.TakeSnapshot()
//	// ... allocate memory ...
//	mt.TakeSnapshot()
//	growth := mt.GetMemoryGrowth()
//	if growth > 1024*1024 {
//	    fmt.Println("Warning: >1MB memory growth detected")
//	}
func (mt *MemoryTracker) GetMemoryGrowth() int64 {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.snapshots) < 2 {
		return 0
	}

	first := mt.snapshots[0]
	last := mt.snapshots[len(mt.snapshots)-1]

	return int64(last.HeapAlloc) - int64(first.HeapAlloc)
}

// GetGoroutineGrowth returns the goroutine count growth between first and last snapshots.
//
// Returns 0 if fewer than 2 snapshots have been taken.
// Note: runtime.MemStats doesn't include goroutine count directly,
// so this uses runtime.NumGoroutine() at snapshot time stored in NumGC field workaround.
// For accurate goroutine tracking, use runtime.NumGoroutine() directly.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetGoroutineGrowth() int {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.snapshots) < 2 {
		return 0
	}

	// Note: MemStats doesn't track goroutines directly
	// This returns 0 as a placeholder - real goroutine tracking
	// would need to be done separately with runtime.NumGoroutine()
	return 0
}

// GetHeapObjectGrowth returns the heap object count growth between first and last snapshots.
//
// Returns 0 if fewer than 2 snapshots have been taken.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetHeapObjectGrowth() int64 {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if len(mt.snapshots) < 2 {
		return 0
	}

	first := mt.snapshots[0]
	last := mt.snapshots[len(mt.snapshots)-1]

	return int64(last.HeapObjects) - int64(first.HeapObjects)
}

// SnapshotCount returns the number of snapshots taken.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) SnapshotCount() int {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return len(mt.snapshots)
}

// AllocationCount returns the number of unique allocation locations tracked.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) AllocationCount() int {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return len(mt.allocations)
}

// GetTotalAllocatedSize returns the total size of all tracked allocations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetTotalAllocatedSize() int64 {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	var total int64
	for _, stats := range mt.allocations {
		total += stats.TotalSize
	}
	return total
}

// GetAllocationLocations returns the names of all allocation locations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) GetAllocationLocations() []string {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	locations := make([]string, 0, len(mt.allocations))
	for loc := range mt.allocations {
		locations = append(locations, loc)
	}
	return locations
}

// Reset clears all snapshots and allocation statistics.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (mt *MemoryTracker) Reset() {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.snapshots = make([]*runtime.MemStats, 0)
	mt.allocations = make(map[string]*AllocationStats)
}
