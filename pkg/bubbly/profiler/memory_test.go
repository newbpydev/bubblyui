package profiler

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMemoryTracker tests tracker creation.
func TestNewMemoryTracker(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates tracker with default settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMemoryTracker()

			require.NotNil(t, tracker, "tracker should not be nil")
			assert.Equal(t, 0, tracker.SnapshotCount(), "should have no snapshots initially")
			assert.Equal(t, 0, tracker.AllocationCount(), "should have no allocations initially")
		})
	}
}

// TestMemoryTracker_TakeSnapshot tests capturing runtime.MemStats.
func TestMemoryTracker_TakeSnapshot(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "captures runtime memory stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMemoryTracker()

			snapshot := tracker.TakeSnapshot()

			require.NotNil(t, snapshot, "snapshot should not be nil")
			assert.Greater(t, snapshot.HeapAlloc, uint64(0), "HeapAlloc should be > 0")
			assert.Equal(t, 1, tracker.SnapshotCount(), "should have 1 snapshot")
		})
	}
}

// TestMemoryTracker_TakeSnapshot_Multiple tests multiple snapshots.
func TestMemoryTracker_TakeSnapshot_Multiple(t *testing.T) {
	tracker := NewMemoryTracker()

	// Take multiple snapshots
	for i := 0; i < 5; i++ {
		snapshot := tracker.TakeSnapshot()
		require.NotNil(t, snapshot)
	}

	assert.Equal(t, 5, tracker.SnapshotCount())

	// Verify all snapshots are accessible
	snapshots := tracker.GetAllSnapshots()
	assert.Len(t, snapshots, 5)
}

// TestMemoryTracker_TrackAllocation tests single allocation tracking.
func TestMemoryTracker_TrackAllocation(t *testing.T) {
	tests := []struct {
		name          string
		location      string
		size          int64
		wantCount     int64
		wantTotalSize int64
		wantAvgSize   int64
	}{
		{
			name:          "tracks single allocation",
			location:      "test.alloc",
			size:          1024,
			wantCount:     1,
			wantTotalSize: 1024,
			wantAvgSize:   1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMemoryTracker()

			tracker.TrackAllocation(tt.location, tt.size)

			stats := tracker.GetAllocation(tt.location)
			require.NotNil(t, stats)
			assert.Equal(t, tt.wantCount, stats.Count)
			assert.Equal(t, tt.wantTotalSize, stats.TotalSize)
			assert.Equal(t, tt.wantAvgSize, stats.AvgSize)
		})
	}
}

// TestMemoryTracker_TrackAllocation_Multiple tests multiple allocations at same location.
func TestMemoryTracker_TrackAllocation_Multiple(t *testing.T) {
	tracker := NewMemoryTracker()

	// Track multiple allocations at same location
	tracker.TrackAllocation("test.alloc", 100)
	tracker.TrackAllocation("test.alloc", 200)
	tracker.TrackAllocation("test.alloc", 300)

	stats := tracker.GetAllocation("test.alloc")
	require.NotNil(t, stats)
	assert.Equal(t, int64(3), stats.Count)
	assert.Equal(t, int64(600), stats.TotalSize)
	assert.Equal(t, int64(200), stats.AvgSize) // 600 / 3 = 200
}

// TestMemoryTracker_TrackAllocation_MultipleLocations tests different locations.
func TestMemoryTracker_TrackAllocation_MultipleLocations(t *testing.T) {
	tracker := NewMemoryTracker()

	tracker.TrackAllocation("location1", 100)
	tracker.TrackAllocation("location2", 200)
	tracker.TrackAllocation("location3", 300)

	assert.Equal(t, 3, tracker.AllocationCount())

	stats1 := tracker.GetAllocation("location1")
	stats2 := tracker.GetAllocation("location2")
	stats3 := tracker.GetAllocation("location3")

	require.NotNil(t, stats1)
	require.NotNil(t, stats2)
	require.NotNil(t, stats3)

	assert.Equal(t, int64(100), stats1.TotalSize)
	assert.Equal(t, int64(200), stats2.TotalSize)
	assert.Equal(t, int64(300), stats3.TotalSize)
}

// TestMemoryTracker_GetAllocation_NonExistent tests getting unknown allocation.
func TestMemoryTracker_GetAllocation_NonExistent(t *testing.T) {
	tracker := NewMemoryTracker()

	stats := tracker.GetAllocation("nonexistent")
	assert.Nil(t, stats)
}

// TestMemoryTracker_GetAllAllocations_Full tests getting all allocations.
func TestMemoryTracker_GetAllAllocations_Full(t *testing.T) {
	tracker := NewMemoryTracker()

	tracker.TrackAllocation("alloc1", 100)
	tracker.TrackAllocation("alloc2", 200)

	allocs := tracker.GetAllAllocations()

	assert.Len(t, allocs, 2)
	assert.Contains(t, allocs, "alloc1")
	assert.Contains(t, allocs, "alloc2")
}

// TestMemoryTracker_GetAllSnapshots tests getting all snapshots.
func TestMemoryTracker_GetAllSnapshots(t *testing.T) {
	tracker := NewMemoryTracker()

	// Take some snapshots
	tracker.TakeSnapshot()
	tracker.TakeSnapshot()
	tracker.TakeSnapshot()

	snapshots := tracker.GetAllSnapshots()

	assert.Len(t, snapshots, 3)
	for _, s := range snapshots {
		assert.NotNil(t, s)
		assert.Greater(t, s.HeapAlloc, uint64(0))
	}
}

// TestMemoryTracker_GetMemoryGrowth tests growth detection.
func TestMemoryTracker_GetMemoryGrowth(t *testing.T) {
	tests := []struct {
		name           string
		snapshotCount  int
		wantGrowthType string // "zero", "positive", or "any"
	}{
		{
			name:           "no snapshots returns zero",
			snapshotCount:  0,
			wantGrowthType: "zero",
		},
		{
			name:           "single snapshot returns zero",
			snapshotCount:  1,
			wantGrowthType: "zero",
		},
		{
			name:           "multiple snapshots returns growth",
			snapshotCount:  3,
			wantGrowthType: "any", // Could be positive, negative, or zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewMemoryTracker()

			for i := 0; i < tt.snapshotCount; i++ {
				tracker.TakeSnapshot()
				// Allocate some memory to potentially cause growth
				if i > 0 {
					_ = make([]byte, 1024*1024) // 1MB allocation
				}
			}

			growth := tracker.GetMemoryGrowth()

			switch tt.wantGrowthType {
			case "zero":
				assert.Equal(t, int64(0), growth)
			case "any":
				// Just verify it returns a value (could be positive, negative, or zero)
				// The actual value depends on GC timing
				_ = growth
			}
		})
	}
}

// TestMemoryTracker_GetMemoryGrowth_Calculation tests growth calculation.
func TestMemoryTracker_GetMemoryGrowth_Calculation(t *testing.T) {
	tracker := NewMemoryTracker()

	// Take first snapshot
	first := tracker.TakeSnapshot()
	require.NotNil(t, first)

	// Allocate some memory to cause growth
	data := make([]byte, 10*1024*1024) // 10MB
	_ = data                           // Keep reference to prevent GC

	// Force GC to get accurate reading
	runtime.GC()

	// Take second snapshot
	second := tracker.TakeSnapshot()
	require.NotNil(t, second)

	// Get growth
	growth := tracker.GetMemoryGrowth()

	// Growth should be calculated as last.HeapAlloc - first.HeapAlloc
	expectedGrowth := int64(second.HeapAlloc) - int64(first.HeapAlloc)
	assert.Equal(t, expectedGrowth, growth)
}

// TestMemoryTracker_Reset tests clearing all data.
func TestMemoryTracker_Reset(t *testing.T) {
	tracker := NewMemoryTracker()

	// Add some data
	tracker.TakeSnapshot()
	tracker.TakeSnapshot()
	tracker.TrackAllocation("test.alloc", 1024)

	assert.Equal(t, 2, tracker.SnapshotCount())
	assert.Equal(t, 1, tracker.AllocationCount())

	// Reset
	tracker.Reset()

	assert.Equal(t, 0, tracker.SnapshotCount())
	assert.Equal(t, 0, tracker.AllocationCount())
	assert.Nil(t, tracker.GetAllocation("test.alloc"))
}

// TestMemoryTracker_SnapshotCount tests counting snapshots.
func TestMemoryTracker_SnapshotCount(t *testing.T) {
	tracker := NewMemoryTracker()

	assert.Equal(t, 0, tracker.SnapshotCount())

	tracker.TakeSnapshot()
	assert.Equal(t, 1, tracker.SnapshotCount())

	tracker.TakeSnapshot()
	assert.Equal(t, 2, tracker.SnapshotCount())

	tracker.TakeSnapshot()
	assert.Equal(t, 3, tracker.SnapshotCount())
}

// TestMemoryTracker_AllocationCount tests counting allocations.
func TestMemoryTracker_AllocationCount(t *testing.T) {
	tracker := NewMemoryTracker()

	assert.Equal(t, 0, tracker.AllocationCount())

	tracker.TrackAllocation("alloc1", 100)
	assert.Equal(t, 1, tracker.AllocationCount())

	tracker.TrackAllocation("alloc2", 200)
	assert.Equal(t, 2, tracker.AllocationCount())

	// Same location doesn't increase count
	tracker.TrackAllocation("alloc1", 300)
	assert.Equal(t, 2, tracker.AllocationCount())
}

// TestMemoryTracker_ThreadSafe tests concurrent access.
func TestMemoryTracker_ThreadSafe(t *testing.T) {
	tracker := NewMemoryTracker()

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent TakeSnapshot calls
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = tracker.TakeSnapshot()
			}
		}()
	}

	// Concurrent TrackAllocation calls
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tracker.TrackAllocation("concurrent.alloc", int64(j+1))
			}
		}()
	}

	wg.Wait()

	// Verify counts are correct
	assert.Equal(t, goroutines*iterations, tracker.SnapshotCount())

	stats := tracker.GetAllocation("concurrent.alloc")
	require.NotNil(t, stats)
	assert.Equal(t, int64(goroutines*iterations), stats.Count)
}

// TestMemoryTracker_ConcurrentReadWrite tests mixed read/write operations.
func TestMemoryTracker_ConcurrentReadWrite(t *testing.T) {
	tracker := NewMemoryTracker()

	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(4)

	// Writer 1: TakeSnapshot
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.TakeSnapshot()
		}
	}()

	// Writer 2: TrackAllocation
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tracker.TrackAllocation("concurrent.op", int64(i+1))
		}
	}()

	// Reader 1: GetAllSnapshots
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.GetAllSnapshots()
		}
	}()

	// Reader 2: GetAllAllocations
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.GetAllAllocations()
		}
	}()

	wg.Wait()

	// Should complete without race conditions
	assert.Equal(t, iterations, tracker.SnapshotCount())
	stats := tracker.GetAllocation("concurrent.op")
	require.NotNil(t, stats)
	assert.Equal(t, int64(iterations), stats.Count)
}

// TestMemoryTracker_GetSnapshotAt tests getting snapshot by index.
func TestMemoryTracker_GetSnapshotAt(t *testing.T) {
	tracker := NewMemoryTracker()

	// Take some snapshots
	s1 := tracker.TakeSnapshot()
	s2 := tracker.TakeSnapshot()
	s3 := tracker.TakeSnapshot()

	// Get by index
	assert.Equal(t, s1, tracker.GetSnapshotAt(0))
	assert.Equal(t, s2, tracker.GetSnapshotAt(1))
	assert.Equal(t, s3, tracker.GetSnapshotAt(2))

	// Out of bounds returns nil
	assert.Nil(t, tracker.GetSnapshotAt(-1))
	assert.Nil(t, tracker.GetSnapshotAt(3))
	assert.Nil(t, tracker.GetSnapshotAt(100))
}

// TestMemoryTracker_GetLatestSnapshot tests getting the most recent snapshot.
func TestMemoryTracker_GetLatestSnapshot(t *testing.T) {
	tracker := NewMemoryTracker()

	// No snapshots
	assert.Nil(t, tracker.GetLatestSnapshot())

	// Take snapshots
	tracker.TakeSnapshot()
	tracker.TakeSnapshot()
	latest := tracker.TakeSnapshot()

	// Should return the last one
	assert.Equal(t, latest, tracker.GetLatestSnapshot())
}

// TestMemoryTracker_GetFirstSnapshot tests getting the first snapshot.
func TestMemoryTracker_GetFirstSnapshot(t *testing.T) {
	tracker := NewMemoryTracker()

	// No snapshots
	assert.Nil(t, tracker.GetFirstSnapshot())

	// Take snapshots
	first := tracker.TakeSnapshot()
	tracker.TakeSnapshot()
	tracker.TakeSnapshot()

	// Should return the first one
	assert.Equal(t, first, tracker.GetFirstSnapshot())
}

// TestMemoryTracker_GetGoroutineGrowth tests goroutine count tracking.
func TestMemoryTracker_GetGoroutineGrowth(t *testing.T) {
	tracker := NewMemoryTracker()

	// No snapshots
	assert.Equal(t, 0, tracker.GetGoroutineGrowth())

	// Take first snapshot
	tracker.TakeSnapshot()

	// Single snapshot
	assert.Equal(t, 0, tracker.GetGoroutineGrowth())

	// Take second snapshot
	tracker.TakeSnapshot()

	// Growth is calculated (could be 0 if no goroutines were created)
	growth := tracker.GetGoroutineGrowth()
	_ = growth // Just verify it doesn't panic
}

// TestMemoryTracker_GetHeapObjectGrowth tests heap object count tracking.
func TestMemoryTracker_GetHeapObjectGrowth(t *testing.T) {
	tracker := NewMemoryTracker()

	// No snapshots
	assert.Equal(t, int64(0), tracker.GetHeapObjectGrowth())

	// Take snapshots
	tracker.TakeSnapshot()
	tracker.TakeSnapshot()

	// Growth is calculated
	growth := tracker.GetHeapObjectGrowth()
	_ = growth // Just verify it doesn't panic
}

// TestAllocationStats_ZeroSize tests tracking zero-size allocations.
func TestAllocationStats_ZeroSize(t *testing.T) {
	tracker := NewMemoryTracker()

	tracker.TrackAllocation("zero.alloc", 0)

	stats := tracker.GetAllocation("zero.alloc")
	require.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Count)
	assert.Equal(t, int64(0), stats.TotalSize)
	assert.Equal(t, int64(0), stats.AvgSize)
}

// TestAllocationStats_NegativeSize tests tracking negative-size allocations.
func TestAllocationStats_NegativeSize(t *testing.T) {
	tracker := NewMemoryTracker()

	// Negative size could represent deallocation
	tracker.TrackAllocation("negative.alloc", -100)

	stats := tracker.GetAllocation("negative.alloc")
	require.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Count)
	assert.Equal(t, int64(-100), stats.TotalSize)
	assert.Equal(t, int64(-100), stats.AvgSize)
}

// TestMemoryTracker_GetTotalAllocatedSize tests total allocation size.
func TestMemoryTracker_GetTotalAllocatedSize(t *testing.T) {
	tracker := NewMemoryTracker()

	assert.Equal(t, int64(0), tracker.GetTotalAllocatedSize())

	tracker.TrackAllocation("alloc1", 100)
	tracker.TrackAllocation("alloc2", 200)
	tracker.TrackAllocation("alloc1", 50) // Same location

	// Total should be 100 + 200 + 50 = 350
	assert.Equal(t, int64(350), tracker.GetTotalAllocatedSize())
}

// TestMemoryTracker_GetAllocationLocations tests getting all location names.
func TestMemoryTracker_GetAllocationLocations(t *testing.T) {
	tracker := NewMemoryTracker()

	tracker.TrackAllocation("loc1", 100)
	tracker.TrackAllocation("loc2", 200)
	tracker.TrackAllocation("loc3", 300)

	locations := tracker.GetAllocationLocations()

	assert.Len(t, locations, 3)
	assert.Contains(t, locations, "loc1")
	assert.Contains(t, locations, "loc2")
	assert.Contains(t, locations, "loc3")
}
