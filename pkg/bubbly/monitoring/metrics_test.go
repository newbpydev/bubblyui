package monitoring

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoOpMetrics_ImplementsInterface tests that NoOpMetrics implements ComposableMetrics
func TestNoOpMetrics_ImplementsInterface(t *testing.T) {
	var _ ComposableMetrics = (*NoOpMetrics)(nil)
}

// TestNoOpMetrics_AllMethodsSafe tests that all NoOpMetrics methods are safe to call
func TestNoOpMetrics_AllMethodsSafe(t *testing.T) {
	noop := &NoOpMetrics{}

	// All methods should be safe (no panics)
	assert.NotPanics(t, func() {
		noop.RecordComposableCreation("UseState", 100*time.Nanosecond)
	}, "RecordComposableCreation should not panic")

	assert.NotPanics(t, func() {
		noop.RecordProvideInjectDepth(5)
	}, "RecordProvideInjectDepth should not panic")

	assert.NotPanics(t, func() {
		noop.RecordAllocationBytes("UseForm", 1024)
	}, "RecordAllocationBytes should not panic")

	assert.NotPanics(t, func() {
		noop.RecordCacheHit("reflection")
	}, "RecordCacheHit should not panic")

	assert.NotPanics(t, func() {
		noop.RecordCacheMiss("timer")
	}, "RecordCacheMiss should not panic")
}

// TestNoOpMetrics_ZeroAllocation tests that NoOpMetrics has zero allocation overhead
func TestNoOpMetrics_ZeroAllocation(t *testing.T) {
	noop := &NoOpMetrics{}

	// Run in a loop to allow allocation detection
	allocs := testing.AllocsPerRun(100, func() {
		noop.RecordComposableCreation("UseState", 100*time.Nanosecond)
		noop.RecordProvideInjectDepth(5)
		noop.RecordAllocationBytes("UseForm", 1024)
		noop.RecordCacheHit("reflection")
		noop.RecordCacheMiss("timer")
	})

	assert.Equal(t, float64(0), allocs, "NoOpMetrics should have zero allocations")
}

// TestGlobalMetrics_DefaultIsNoOp tests that global metrics defaults to NoOp
func TestGlobalMetrics_DefaultIsNoOp(t *testing.T) {
	// Reset to default
	SetGlobalMetrics(&NoOpMetrics{})

	metrics := GetGlobalMetrics()
	require.NotNil(t, metrics, "GetGlobalMetrics should never return nil")

	// Should be NoOpMetrics
	_, ok := metrics.(*NoOpMetrics)
	assert.True(t, ok, "Default metrics should be NoOpMetrics")
}

// TestGlobalMetrics_SetAndGet tests setting and getting global metrics
func TestGlobalMetrics_SetAndGet(t *testing.T) {
	// Create mock metrics
	mock := &MockMetrics{}

	// Set global metrics
	SetGlobalMetrics(mock)

	// Get and verify
	metrics := GetGlobalMetrics()
	require.NotNil(t, metrics, "GetGlobalMetrics should not return nil")

	retrieved, ok := metrics.(*MockMetrics)
	assert.True(t, ok, "Should retrieve MockMetrics")
	assert.Equal(t, mock, retrieved, "Should be the same instance")

	// Reset to NoOp
	SetGlobalMetrics(&NoOpMetrics{})
}

// TestGlobalMetrics_ThreadSafe tests that global metrics is thread-safe
func TestGlobalMetrics_ThreadSafe(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Alternate between setting and getting
			if id%2 == 0 {
				SetGlobalMetrics(&NoOpMetrics{})
			} else {
				metrics := GetGlobalMetrics()
				require.NotNil(t, metrics, "Should never get nil metrics")
			}
		}(i)
	}

	wg.Wait()

	// Verify we can still get metrics
	metrics := GetGlobalMetrics()
	assert.NotNil(t, metrics, "Should have valid metrics after concurrent access")
}

// TestGlobalMetrics_NilSafety tests that setting nil doesn't break system
func TestGlobalMetrics_NilSafety(t *testing.T) {
	// Setting nil should either panic or default to NoOp (depending on implementation)
	// For safety, we expect it to handle nil gracefully
	SetGlobalMetrics(nil)

	metrics := GetGlobalMetrics()
	assert.NotNil(t, metrics, "GetGlobalMetrics should never return nil even after setting nil")
}

// TestMultipleImplementations tests that multiple implementations can be used
func TestMultipleImplementations(t *testing.T) {
	implementations := []ComposableMetrics{
		&NoOpMetrics{},
		&MockMetrics{},
	}

	for i, impl := range implementations {
		t.Run(fmt.Sprintf("Implementation_%d", i), func(t *testing.T) {
			// Set implementation
			SetGlobalMetrics(impl)

			// Get and verify
			metrics := GetGlobalMetrics()
			require.NotNil(t, metrics, "Metrics should not be nil for implementation %d", i)

			// Should be safe to call
			assert.NotPanics(t, func() {
				metrics.RecordComposableCreation("UseState", 100*time.Nanosecond)
				metrics.RecordProvideInjectDepth(3)
				metrics.RecordAllocationBytes("UseAsync", 512)
				metrics.RecordCacheHit("timer")
				metrics.RecordCacheMiss("reflection")
			}, "Implementation %d should not panic", i)
		})
	}

	// Reset
	SetGlobalMetrics(&NoOpMetrics{})
}

// MockMetrics is a mock implementation for testing
type MockMetrics struct {
	CreationCalls       int
	DepthCalls          int
	AllocationCalls     int
	CacheHitCalls       int
	CacheMissCalls      int
	LastComposableName  string
	LastCreationTime    time.Duration
	LastDepth           int
	LastAllocComposable string
	LastAllocBytes      int64
	LastCacheHitName    string
	LastCacheMissName   string
	mu                  sync.Mutex
}

func (m *MockMetrics) RecordComposableCreation(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreationCalls++
	m.LastComposableName = name
	m.LastCreationTime = duration
}

func (m *MockMetrics) RecordProvideInjectDepth(depth int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DepthCalls++
	m.LastDepth = depth
}

func (m *MockMetrics) RecordAllocationBytes(composable string, bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AllocationCalls++
	m.LastAllocComposable = composable
	m.LastAllocBytes = bytes
}

func (m *MockMetrics) RecordCacheHit(cache string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheHitCalls++
	m.LastCacheHitName = cache
}

func (m *MockMetrics) RecordCacheMiss(cache string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheMissCalls++
	m.LastCacheMissName = cache
}

// TestMockMetrics_Records tests that MockMetrics records calls
func TestMockMetrics_Records(t *testing.T) {
	mock := &MockMetrics{}

	// Test RecordComposableCreation
	mock.RecordComposableCreation("UseState", 150*time.Nanosecond)
	assert.Equal(t, 1, mock.CreationCalls)
	assert.Equal(t, "UseState", mock.LastComposableName)
	assert.Equal(t, 150*time.Nanosecond, mock.LastCreationTime)

	// Test RecordProvideInjectDepth
	mock.RecordProvideInjectDepth(7)
	assert.Equal(t, 1, mock.DepthCalls)
	assert.Equal(t, 7, mock.LastDepth)

	// Test RecordAllocationBytes
	mock.RecordAllocationBytes("UseForm", 2048)
	assert.Equal(t, 1, mock.AllocationCalls)
	assert.Equal(t, "UseForm", mock.LastAllocComposable)
	assert.Equal(t, int64(2048), mock.LastAllocBytes)

	// Test RecordCacheHit
	mock.RecordCacheHit("reflection")
	assert.Equal(t, 1, mock.CacheHitCalls)
	assert.Equal(t, "reflection", mock.LastCacheHitName)

	// Test RecordCacheMiss
	mock.RecordCacheMiss("timer")
	assert.Equal(t, 1, mock.CacheMissCalls)
	assert.Equal(t, "timer", mock.LastCacheMissName)
}

// TestMockMetrics_Concurrent tests MockMetrics is thread-safe
func TestMockMetrics_Concurrent(t *testing.T) {
	mock := &MockMetrics{}

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mock.RecordComposableCreation("UseState", 100*time.Nanosecond)
			mock.RecordCacheHit("test")
		}()
	}

	wg.Wait()

	// Should have recorded all calls
	assert.Equal(t, numGoroutines, mock.CreationCalls, "Should record all creation calls")
	assert.Equal(t, numGoroutines, mock.CacheHitCalls, "Should record all cache hit calls")
}
