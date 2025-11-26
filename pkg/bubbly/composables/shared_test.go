package composables

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Test helper: simple counter composable for testing
type sharedCounterComposable struct {
	count     *atomic.Int32 // Use atomic for thread-safe access
	initCount *atomic.Int32 // Track how many times factory was called
	Increment func()
	Decrement func()
}

func newSharedCounterComposable(_ *bubbly.Context, initCount *atomic.Int32) *sharedCounterComposable {
	initCount.Add(1) // Track factory calls
	count := &atomic.Int32{}

	return &sharedCounterComposable{
		count:     count,
		initCount: initCount,
		Increment: func() {
			count.Add(1)
		},
		Decrement: func() {
			count.Add(-1)
		},
	}
}

// TestCreateShared_BasicUsage verifies basic singleton behavior
func TestCreateShared_BasicUsage(t *testing.T) {
	initCount := &atomic.Int32{}

	// Create shared factory
	UseSharedCounter := CreateShared(func(ctx *bubbly.Context) *sharedCounterComposable {
		return newSharedCounterComposable(ctx, initCount)
	})

	// Create test context
	ctx := createTestContext()

	// First call - should initialize
	counter1 := UseSharedCounter(ctx)
	assert.NotNil(t, counter1)
	assert.Equal(t, int32(1), initCount.Load(), "Factory should be called once")

	// Modify state
	counter1.Increment()
	counter1.Increment()

	// Second call - should return same instance
	counter2 := UseSharedCounter(ctx)
	assert.NotNil(t, counter2)
	assert.Equal(t, int32(1), initCount.Load(), "Factory should still be called only once")

	// Verify same instance (state persists)
	assert.Equal(t, counter1.count.Load(), counter2.count.Load(), "Should return same instance")
	assert.Equal(t, int32(2), counter2.count.Load(), "State should persist")
}

// TestCreateShared_ThreadSafe verifies concurrent access safety
func TestCreateShared_ThreadSafe(t *testing.T) {
	initCount := &atomic.Int32{}

	// Create shared factory
	UseSharedCounter := CreateShared(func(ctx *bubbly.Context) *sharedCounterComposable {
		return newSharedCounterComposable(ctx, initCount)
	})

	// Launch 100 concurrent goroutines
	const numGoroutines = 100
	var wg sync.WaitGroup
	instances := make([]*sharedCounterComposable, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ctx := createTestContext()
			instances[index] = UseSharedCounter(ctx)
		}(i)
	}

	wg.Wait()

	// Verify factory called exactly once
	assert.Equal(t, int32(1), initCount.Load(), "Factory should be called exactly once despite concurrent access")

	// Verify all instances are the same (check first few)
	for i := 1; i < 10 && i < numGoroutines; i++ {
		assert.Equal(t, instances[0].count.Load(), instances[i].count.Load(),
			"All instances should share same state")
	}
}

// TestCreateShared_DifferentTypes verifies generics work with various types
func TestCreateShared_DifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		factory  func(*bubbly.Context) interface{}
		expected interface{}
	}{
		{
			name: "int type",
			factory: func(ctx *bubbly.Context) interface{} {
				return 42
			},
			expected: 42,
		},
		{
			name: "string type",
			factory: func(ctx *bubbly.Context) interface{} {
				return "hello"
			},
			expected: "hello",
		},
		{
			name: "struct type",
			factory: func(ctx *bubbly.Context) interface{} {
				return struct{ Value int }{Value: 100}
			},
			expected: struct{ Value int }{Value: 100},
		},
		{
			name: "pointer type",
			factory: func(ctx *bubbly.Context) interface{} {
				val := 99
				return &val
			},
			expected: func() interface{} {
				val := 99
				return &val
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared := CreateShared(tt.factory)
			ctx := createTestContext()

			result1 := shared(ctx)
			result2 := shared(ctx)

			// For pointer types, verify same pointer
			if _, ok := tt.expected.(*int); ok {
				assert.Same(t, result1, result2, "Should return same pointer instance")
			} else {
				assert.Equal(t, result1, result2, "Should return same value")
			}
		})
	}
}

// TestCreateShared_NilFactory verifies nil factory panics
func TestCreateShared_NilFactory(t *testing.T) {
	// This should panic when factory is nil
	assert.Panics(t, func() {
		shared := CreateShared[int](nil)
		ctx := createTestContext()
		_ = shared(ctx)
	}, "Nil factory should panic")
}

// TestCreateShared_FactoryPanic verifies factory panic propagates
func TestCreateShared_FactoryPanic(t *testing.T) {
	shared := CreateShared(func(ctx *bubbly.Context) int {
		panic("factory panic")
	})

	ctx := createTestContext()

	assert.Panics(t, func() {
		_ = shared(ctx)
	}, "Factory panic should propagate to caller")
}

// TestCreateShared_IndependentInstances verifies multiple shared factories are independent
func TestCreateShared_IndependentInstances(t *testing.T) {
	initCount1 := &atomic.Int32{}
	initCount2 := &atomic.Int32{}

	// Create two independent shared factories
	UseSharedCounter1 := CreateShared(func(ctx *bubbly.Context) *sharedCounterComposable {
		return newSharedCounterComposable(ctx, initCount1)
	})

	UseSharedCounter2 := CreateShared(func(ctx *bubbly.Context) *sharedCounterComposable {
		return newSharedCounterComposable(ctx, initCount2)
	})

	ctx := createTestContext()

	// Call both factories
	counter1 := UseSharedCounter1(ctx)
	counter2 := UseSharedCounter2(ctx)

	// Verify both factories were called
	assert.Equal(t, int32(1), initCount1.Load(), "First factory should be called once")
	assert.Equal(t, int32(1), initCount2.Load(), "Second factory should be called once")

	// Modify first counter
	counter1.Increment()
	counter1.Increment()

	// Verify second counter is independent
	assert.Equal(t, int32(2), counter1.count.Load(), "First counter should be 2")
	assert.Equal(t, int32(0), counter2.count.Load(), "Second counter should be 0 (independent)")
}

// TestCreateShared_PersistsAcrossLifecycle verifies instance persists across component lifecycle
func TestCreateShared_PersistsAcrossLifecycle(t *testing.T) {
	initCount := &atomic.Int32{}

	UseSharedCounter := CreateShared(func(ctx *bubbly.Context) *sharedCounterComposable {
		return newSharedCounterComposable(ctx, initCount)
	})

	// Simulate first component lifecycle
	ctx1 := createTestContext()
	counter1 := UseSharedCounter(ctx1)
	counter1.Increment()
	counter1.Increment()
	counter1.Increment()

	// Simulate second component lifecycle (different context)
	ctx2 := createTestContext()
	counter2 := UseSharedCounter(ctx2)

	// Verify same instance persists
	assert.Equal(t, int32(1), initCount.Load(), "Factory should be called only once")
	assert.Equal(t, int32(3), counter2.count.Load(), "State should persist across contexts")
}

// ====================================================================
// BENCHMARKS - Shared Composable Performance Validation
// ====================================================================
//
// Performance Targets (from tasks.md Task 6.2):
// - Subsequent calls ≤ 50ns/op
// - Memory savings vs recreated composables
// - sync.Once overhead acceptable
//
// Benchmark Targets (from designs.md line 418):
// - BenchmarkSharedComposable: 50000000 ops, 30 ns/op, 0 B/op, 0 allocs/op

// benchmarkComposable is a simple composable for benchmarking
type benchmarkComposable struct {
	count *atomic.Int32
}

func newBenchmarkComposable(_ *bubbly.Context) *benchmarkComposable {
	return &benchmarkComposable{
		count: &atomic.Int32{},
	}
}

// BenchmarkSharedComposable_FirstCall benchmarks the initial creation of a shared composable.
// This measures the overhead of sync.Once initialization.
//
// Note: This benchmark creates a new shared factory for each iteration to measure
// the first-call overhead. In real usage, this happens only once per application.
func BenchmarkSharedComposable_FirstCall(b *testing.B) {
	ctx := createTestContext()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create new shared factory each iteration to measure first-call overhead
		UseShared := CreateShared(newBenchmarkComposable)
		_ = UseShared(ctx)
	}
}

// BenchmarkSharedComposable_SubsequentCalls benchmarks cached access to a shared composable.
// This measures the overhead of sync.Once on subsequent calls (should be very fast).
//
// Target: ≤50ns/op, 0 allocations
func BenchmarkSharedComposable_SubsequentCalls(b *testing.B) {
	// Setup: Create shared factory and initialize it once
	UseSharedCounter := CreateShared(newBenchmarkComposable)
	ctx := createTestContext()

	// Initialize the shared composable (first call)
	_ = UseSharedCounter(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark subsequent calls (cached access)
	for i := 0; i < b.N; i++ {
		_ = UseSharedCounter(ctx)
	}
}

// BenchmarkRecreatedComposable benchmarks creating a new composable each time.
// This is the baseline comparison for CreateShared - the old way of doing things.
//
// This benchmark measures the overhead of recreating composables without sharing,
// to compare with the shared pattern and show memory savings.
func BenchmarkRecreatedComposable(b *testing.B) {
	ctx := createTestContext()

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark recreating composable each time (non-shared pattern)
	for i := 0; i < b.N; i++ {
		_ = newBenchmarkComposable(ctx)
	}
}

// BenchmarkSharedComposable_ConcurrentAccess benchmarks concurrent access to shared composable.
// This measures the overhead of sync.Once under concurrent load.
func BenchmarkSharedComposable_ConcurrentAccess(b *testing.B) {
	// Setup: Create shared factory and initialize it once
	UseSharedCounter := CreateShared(newBenchmarkComposable)
	ctx := createTestContext()

	// Initialize the shared composable (first call)
	_ = UseSharedCounter(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark concurrent access
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = UseSharedCounter(ctx)
		}
	})
}

// BenchmarkSharedComposable_WithState benchmarks shared composable with state operations.
// This measures real-world usage with state modifications.
func BenchmarkSharedComposable_WithState(b *testing.B) {
	// Setup: Create shared factory with state
	UseSharedCounter := CreateShared(func(ctx *bubbly.Context) *benchmarkComposable {
		return &benchmarkComposable{
			count: &atomic.Int32{},
		}
	})
	ctx := createTestContext()

	// Initialize the shared composable
	counter := UseSharedCounter(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark access + state operation
	for i := 0; i < b.N; i++ {
		c := UseSharedCounter(ctx)
		c.count.Add(1)
		_ = c.count.Load()
	}

	// Prevent compiler optimization
	_ = counter
}

// BenchmarkRecreatedComposable_WithState benchmarks recreated composable with state operations.
// This is the baseline comparison for BenchmarkSharedComposable_WithState.
func BenchmarkRecreatedComposable_WithState(b *testing.B) {
	ctx := createTestContext()

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark recreating composable + state operation each time
	for i := 0; i < b.N; i++ {
		c := newBenchmarkComposable(ctx)
		c.count.Add(1)
		_ = c.count.Load()
	}
}
