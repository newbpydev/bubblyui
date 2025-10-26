package bubbly

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDependency is a test implementation of the Dependency interface.
type mockDependency struct {
	id          string
	invalidated bool
	dependents  []Dependency
	mu          sync.Mutex
}

func newMockDependency(id string) *mockDependency {
	return &mockDependency{id: id}
}

func (m *mockDependency) Invalidate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invalidated = true
}

func (m *mockDependency) AddDependent(dep Dependency) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dependents = append(m.dependents, dep)
}

func (m *mockDependency) IsInvalidated() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.invalidated
}

func (m *mockDependency) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invalidated = false
}

// TestDepTracker_BasicTracking tests basic dependency tracking functionality.
func TestDepTracker_BasicTracking(t *testing.T) {
	tests := []struct {
		name          string
		setupFn       func(*DepTracker) (Dependency, []Dependency)
		expectedCount int
	}{
		{
			name: "track single dependency",
			setupFn: func(dt *DepTracker) (Dependency, []Dependency) {
				computed := newMockDependency("computed")
				ref := newMockDependency("ref")

				err := dt.BeginTracking(computed)
				require.NoError(t, err)
				dt.Track(ref)

				return computed, []Dependency{ref}
			},
			expectedCount: 1,
		},
		{
			name: "track multiple dependencies",
			setupFn: func(dt *DepTracker) (Dependency, []Dependency) {
				computed := newMockDependency("computed")
				ref1 := newMockDependency("ref1")
				ref2 := newMockDependency("ref2")
				ref3 := newMockDependency("ref3")

				err := dt.BeginTracking(computed)
				require.NoError(t, err)
				dt.Track(ref1)
				dt.Track(ref2)
				dt.Track(ref3)

				return computed, []Dependency{ref1, ref2, ref3}
			},
			expectedCount: 3,
		},
		{
			name: "track with duplicates",
			setupFn: func(dt *DepTracker) (Dependency, []Dependency) {
				computed := newMockDependency("computed")
				ref := newMockDependency("ref")

				err := dt.BeginTracking(computed)
				require.NoError(t, err)
				dt.Track(ref)
				dt.Track(ref) // Duplicate
				dt.Track(ref) // Duplicate

				return computed, []Dependency{ref}
			},
			expectedCount: 1,
		},
		{
			name: "no tracking when not started",
			setupFn: func(dt *DepTracker) (Dependency, []Dependency) {
				ref := newMockDependency("ref")
				dt.Track(ref) // Should be ignored

				return nil, []Dependency{}
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := &DepTracker{}
			_, expectedDeps := tt.setupFn(dt)

			deps := dt.EndTracking()

			assert.Equal(t, tt.expectedCount, len(deps), "unexpected number of tracked dependencies")

			if tt.expectedCount > 0 {
				for i, dep := range expectedDeps {
					assert.Equal(t, dep, deps[i], "dependency mismatch at index %d", i)
				}
			}
		})
	}
}

// TestDepTracker_IsTracking tests the IsTracking method.
func TestDepTracker_IsTracking(t *testing.T) {
	dt := &DepTracker{}
	computed := newMockDependency("computed")

	// Initially not tracking
	assert.False(t, dt.IsTracking(), "should not be tracking initially")

	// Start tracking
	err := dt.BeginTracking(computed)
	require.NoError(t, err)
	assert.True(t, dt.IsTracking(), "should be tracking after BeginTracking")

	// End tracking
	dt.EndTracking()
	assert.False(t, dt.IsTracking(), "should not be tracking after EndTracking")
}

// TestDepTracker_NestedTracking tests nested dependency tracking (computed -> computed).
func TestDepTracker_NestedTracking(t *testing.T) {
	dt := &DepTracker{}

	computed1 := newMockDependency("computed1")
	computed2 := newMockDependency("computed2")
	ref1 := newMockDependency("ref1")
	ref2 := newMockDependency("ref2")

	// Start tracking for computed1
	err := dt.BeginTracking(computed1)
	require.NoError(t, err)

	dt.Track(ref1)

	// Start nested tracking for computed2
	err = dt.BeginTracking(computed2)
	require.NoError(t, err)

	dt.Track(ref2)

	// End nested tracking
	deps2 := dt.EndTracking()
	assert.Equal(t, 1, len(deps2), "computed2 should track ref2")
	assert.Equal(t, ref2, deps2[0])

	// Should still be tracking for computed1
	assert.True(t, dt.IsTracking(), "should still be tracking for outer computed")

	// Track computed2 as dependency of computed1
	dt.Track(computed2)

	// End outer tracking
	deps1 := dt.EndTracking()
	assert.Equal(t, 2, len(deps1), "computed1 should track ref1 and computed2")

	// Should not be tracking anymore
	assert.False(t, dt.IsTracking(), "should not be tracking after all EndTracking calls")
}

// TestDepTracker_CircularDependency tests circular dependency detection.
func TestDepTracker_CircularDependency(t *testing.T) {
	dt := &DepTracker{}

	computed1 := newMockDependency("computed1")
	computed2 := newMockDependency("computed2")

	// Start tracking for computed1
	err := dt.BeginTracking(computed1)
	require.NoError(t, err)

	// Start tracking for computed2
	err = dt.BeginTracking(computed2)
	require.NoError(t, err)

	// Try to start tracking for computed1 again (circular)
	err = dt.BeginTracking(computed1)
	assert.ErrorIs(t, err, ErrCircularDependency, "should detect circular dependency")

	// Clean up
	dt.EndTracking() // computed2
	dt.EndTracking() // computed1
}

// TestDepTracker_MaxDepth tests maximum depth enforcement.
func TestDepTracker_MaxDepth(t *testing.T) {
	dt := &DepTracker{}

	// Create a chain of dependencies up to max depth
	deps := make([]*mockDependency, MaxDependencyDepth+1)
	for i := range deps {
		deps[i] = newMockDependency("dep")
	}

	// Fill up to max depth
	for i := 0; i < MaxDependencyDepth; i++ {
		err := dt.BeginTracking(deps[i])
		require.NoError(t, err, "should allow up to max depth")
	}

	// Try to exceed max depth
	err := dt.BeginTracking(deps[MaxDependencyDepth])
	assert.ErrorIs(t, err, ErrMaxDepthExceeded, "should reject depth exceeding max")

	// Clean up
	for i := 0; i < MaxDependencyDepth; i++ {
		dt.EndTracking()
	}
}

// TestDepTracker_Concurrent tests thread safety of dependency tracking.
func TestDepTracker_Concurrent(t *testing.T) {
	dt := &DepTracker{}

	const numGoroutines = 10
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < numOps; j++ {
				computed := newMockDependency("computed")
				ref := newMockDependency("ref")

				err := dt.BeginTracking(computed)
				if err != nil {
					continue // May fail due to concurrent access
				}

				dt.Track(ref)
				deps := dt.EndTracking()

				// Verify we got our dependency back
				if len(deps) > 0 {
					assert.Equal(t, ref, deps[0])
				}
			}
		}()
	}

	wg.Wait()

	// Tracker should be clean after all goroutines finish
	assert.False(t, dt.IsTracking(), "tracker should not be tracking after all operations")
}

// TestDepTracker_TrackingIsolation tests that tracking is isolated between operations.
func TestDepTracker_TrackingIsolation(t *testing.T) {
	dt := &DepTracker{}

	computed1 := newMockDependency("computed1")
	computed2 := newMockDependency("computed2")
	ref1 := newMockDependency("ref1")
	ref2 := newMockDependency("ref2")

	// First tracking session
	err := dt.BeginTracking(computed1)
	require.NoError(t, err)
	dt.Track(ref1)
	deps1 := dt.EndTracking()

	assert.Equal(t, 1, len(deps1))
	assert.Equal(t, ref1, deps1[0])

	// Second tracking session should be isolated
	err = dt.BeginTracking(computed2)
	require.NoError(t, err)
	dt.Track(ref2)
	deps2 := dt.EndTracking()

	assert.Equal(t, 1, len(deps2))
	assert.Equal(t, ref2, deps2[0])

	// Verify isolation - deps2 should not contain ref1
	for _, dep := range deps2 {
		assert.NotEqual(t, ref1, dep, "second session should not contain dependencies from first session")
	}
}

// TestDepTracker_HighConcurrency tests scalability with 100+ concurrent goroutines.
// This test exposes the global tracker contention issue and should pass after
// implementing per-goroutine tracking.
func TestDepTracker_HighConcurrency(t *testing.T) {
	dt := &DepTracker{}

	const numGoroutines = 100
	const numOps = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Track successful operations
	var successCount atomic.Int32

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOps; j++ {
				computed := newMockDependency("computed")
				ref1 := newMockDependency("ref1")
				ref2 := newMockDependency("ref2")

				err := dt.BeginTracking(computed)
				if err != nil {
					// Should not fail with per-goroutine tracking
					t.Errorf("goroutine %d: BeginTracking failed: %v", id, err)
					continue
				}

				dt.Track(ref1)
				dt.Track(ref2)
				deps := dt.EndTracking()

				// Verify we got our dependencies back
				if len(deps) == 2 {
					successCount.Add(1)
				} else {
					t.Errorf("goroutine %d: expected 2 deps, got %d", id, len(deps))
				}
			}
		}(i)
	}

	wg.Wait()

	// All operations should succeed with per-goroutine tracking
	expectedSuccess := int32(numGoroutines * numOps)
	actualSuccess := successCount.Load()
	assert.Equal(t, expectedSuccess, actualSuccess,
		"expected %d successful operations, got %d", expectedSuccess, actualSuccess)

	// Tracker should be clean after all goroutines finish
	assert.False(t, dt.IsTracking(), "tracker should not be tracking after all operations")
}

// TestDepTracker_GoroutineIsolation tests that goroutines don't interfere with each other.
func TestDepTracker_GoroutineIsolation(t *testing.T) {
	dt := &DepTracker{}

	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Each goroutine tracks different dependencies
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			computed := newMockDependency("computed")
			ref := newMockDependency("ref")

			err := dt.BeginTracking(computed)
			require.NoError(t, err)

			dt.Track(ref)

			deps := dt.EndTracking()

			// Each goroutine should get exactly its own dependency
			assert.Equal(t, 1, len(deps), "goroutine %d: expected 1 dependency", id)
			if len(deps) > 0 {
				assert.Equal(t, ref, deps[0], "goroutine %d: wrong dependency", id)
			}
		}(i)
	}

	wg.Wait()
}

// Benchmarks

// BenchmarkDepTracker_Sequential benchmarks sequential tracking operations.
func BenchmarkDepTracker_Sequential(b *testing.B) {
	dt := &DepTracker{}
	computed := newMockDependency("computed")
	ref := newMockDependency("ref")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dt.BeginTracking(computed)
		dt.Track(ref)
		_ = dt.EndTracking()
	}
}

// BenchmarkDepTracker_Concurrent benchmarks concurrent tracking with multiple goroutines.
func BenchmarkDepTracker_Concurrent(b *testing.B) {
	dt := &DepTracker{}

	b.RunParallel(func(pb *testing.PB) {
		computed := newMockDependency("computed")
		ref := newMockDependency("ref")

		for pb.Next() {
			_ = dt.BeginTracking(computed)
			dt.Track(ref)
			_ = dt.EndTracking()
		}
	})
}

// BenchmarkDepTracker_HighConcurrency benchmarks with 100 goroutines.
func BenchmarkDepTracker_HighConcurrency(b *testing.B) {
	dt := &DepTracker{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		computed := newMockDependency("computed")
		ref1 := newMockDependency("ref1")
		ref2 := newMockDependency("ref2")

		for pb.Next() {
			_ = dt.BeginTracking(computed)
			dt.Track(ref1)
			dt.Track(ref2)
			_ = dt.EndTracking()
		}
	})
}

// BenchmarkGetGoroutineID benchmarks the goroutine ID extraction.
func BenchmarkGetGoroutineID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getGoroutineID()
	}
}
