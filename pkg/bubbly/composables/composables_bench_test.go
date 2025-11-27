package composables

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ============================================================================
// UseState Benchmarks
// ============================================================================

// BenchmarkUseState measures the overhead of creating a UseState composable
// Target: < 200ns per operation
func BenchmarkUseState(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state := UseState(ctx, 0)
		_ = state
	}
}

// BenchmarkUseState_Set measures the overhead of calling Set on UseState
func BenchmarkUseState_Set(b *testing.B) {
	ctx := bubbly.NewTestContext()
	state := UseState(ctx, 0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state.Set(i)
	}
}

// BenchmarkUseState_Get measures the overhead of calling Get on UseState
func BenchmarkUseState_Get(b *testing.B) {
	ctx := bubbly.NewTestContext()
	state := UseState(ctx, 42)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = state.Get()
	}
}

// ============================================================================
// UseAsync Benchmarks
// ============================================================================

// BenchmarkUseAsync measures the overhead of creating a UseAsync composable
// Target: < 1μs per operation
func BenchmarkUseAsync(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		async := UseAsync(ctx, func() (*int, error) {
			result := 42
			return &result, nil
		})
		_ = async
	}
}

// BenchmarkUseAsync_Execute measures the overhead of Execute (not the actual async work)
func BenchmarkUseAsync_Execute(b *testing.B) {
	ctx := bubbly.NewTestContext()
	executed := 0
	async := UseAsync(ctx, func() (*int, error) {
		executed++
		result := executed
		return &result, nil
	})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		async.Execute()
	}
}

// ============================================================================
// UseEffect Benchmarks
// ============================================================================

// BenchmarkUseEffect measures the overhead of registering an effect
func BenchmarkUseEffect(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		UseEffect(ctx, func() UseEffectCleanup {
			return nil
		})
	}
}

// BenchmarkUseEffect_WithDeps measures effect with dependencies
func BenchmarkUseEffect_WithDeps(b *testing.B) {
	ctx := bubbly.NewTestContext()
	dep1 := bubbly.NewRef(0)
	dep2 := bubbly.NewRef("test")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		UseEffect(ctx, func() UseEffectCleanup {
			return nil
		}, dep1, dep2)
	}
}

// ============================================================================
// UseDebounce Benchmarks
// ============================================================================

// BenchmarkUseDebounce measures debounce creation overhead
func BenchmarkUseDebounce(b *testing.B) {
	ctx := bubbly.NewTestContext()
	value := bubbly.NewRef(0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		debounced := UseDebounce(ctx, value, 100)
		_ = debounced
	}
}

// ============================================================================
// UseThrottle Benchmarks
// ============================================================================

// BenchmarkUseThrottle measures throttle creation overhead
func BenchmarkUseThrottle(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		throttled := UseThrottle(ctx, func() {}, 100)
		_ = throttled
	}
}

// ============================================================================
// UseForm Benchmarks
// ============================================================================

// BenchmarkUseForm measures form creation overhead
func BenchmarkUseForm(b *testing.B) {
	type TestForm struct {
		Name  string
		Email string
	}

	ctx := bubbly.NewTestContext()
	validator := func(f TestForm) map[string]string {
		return make(map[string]string)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		form := UseForm(ctx, TestForm{}, validator)
		_ = form
	}
}

// BenchmarkUseForm_SetField measures SetField performance
func BenchmarkUseForm_SetField(b *testing.B) {
	type TestForm struct {
		Name  string
		Email string
	}

	ctx := bubbly.NewTestContext()
	form := UseForm(ctx, TestForm{}, func(f TestForm) map[string]string {
		return make(map[string]string)
	})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		form.SetField("Name", "Test")
	}
}

// ============================================================================
// UseLocalStorage Benchmarks
// ============================================================================

// BenchmarkUseLocalStorage measures storage composable creation
func BenchmarkUseLocalStorage(b *testing.B) {
	ctx := bubbly.NewTestContext()
	storage := NewMemoryStorage() // Use in-memory storage for benchmarks
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state := UseLocalStorage(ctx, "test-key", 0, storage)
		_ = state
	}
}

// ============================================================================
// Provide/Inject Benchmarks
// ============================================================================

// BenchmarkProvideInject_Depth1 measures inject with 1-level tree
// Target: < 500ns per operation
func BenchmarkProvideInject_Depth1(b *testing.B) {
	// Create parent
	parent := bubbly.NewTestContext()
	parent.Provide("key", "value")

	// Create child
	child := bubbly.NewTestContext()
	bubbly.SetParent(child, parent)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = child.Inject("key", "default")
	}
}

// BenchmarkProvideInject_Depth3 measures inject with 3-level tree
func BenchmarkProvideInject_Depth3(b *testing.B) {
	// Create 3-level tree: grandparent -> parent -> child
	grandparent := bubbly.NewTestContext()
	grandparent.Provide("key", "value")

	parent := bubbly.NewTestContext()
	bubbly.SetParent(parent, grandparent)

	child := bubbly.NewTestContext()
	bubbly.SetParent(child, parent)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = child.Inject("key", "default")
	}
}

// BenchmarkProvideInject_Depth5 measures inject with 5-level tree
func BenchmarkProvideInject_Depth5(b *testing.B) {
	// Create 5-level tree
	level0 := bubbly.NewTestContext()
	level0.Provide("key", "value")

	level1 := bubbly.NewTestContext()
	bubbly.SetParent(level1, level0)

	level2 := bubbly.NewTestContext()
	bubbly.SetParent(level2, level1)

	level3 := bubbly.NewTestContext()
	bubbly.SetParent(level3, level2)

	level4 := bubbly.NewTestContext()
	bubbly.SetParent(level4, level3)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = level4.Inject("key", "default")
	}
}

// BenchmarkProvideInject_Depth10 measures inject with 10-level tree
func BenchmarkProvideInject_Depth10(b *testing.B) {
	// Create 10-level tree
	contexts := make([]*bubbly.Context, 11)
	contexts[0] = bubbly.NewTestContext()
	contexts[0].Provide("key", "value")

	for i := 1; i <= 10; i++ {
		contexts[i] = bubbly.NewTestContext()
		bubbly.SetParent(contexts[i], contexts[i-1])
	}

	deepest := contexts[10]
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = deepest.Inject("key", "default")
	}
}

// BenchmarkProvideInject_CachedLookup measures repeated inject calls (for cache effectiveness)
func BenchmarkProvideInject_CachedLookup(b *testing.B) {
	// Create 5-level tree
	level0 := bubbly.NewTestContext()
	level0.Provide("key", "value")

	level1 := bubbly.NewTestContext()
	bubbly.SetParent(level1, level0)

	level2 := bubbly.NewTestContext()
	bubbly.SetParent(level2, level1)

	level3 := bubbly.NewTestContext()
	bubbly.SetParent(level3, level2)

	level4 := bubbly.NewTestContext()
	bubbly.SetParent(level4, level3)

	// Prime the cache with first call
	_ = level4.Inject("key", "default")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = level4.Inject("key", "default")
	}
}

// ============================================================================
// Composable Chain Benchmarks
// ============================================================================

// benchCounter is a simple composable for benchmarking (internal use only)
func benchCounter(ctx *bubbly.Context, initial int) (*bubbly.Ref[int], func(), func()) {
	state := UseState(ctx, initial)

	increment := func() {
		state.Set(state.Get() + 1)
	}

	decrement := func() {
		state.Set(state.Get() - 1)
	}

	return state.Value, increment, decrement
}

// benchDoubleCounter chains benchCounter
//
//nolint:unparam // third return value is used in BenchmarkComposableChain for API consistency
func benchDoubleCounter(ctx *bubbly.Context, initial int) (*bubbly.Ref[int], func(), func()) {
	count, inc, dec := benchCounter(ctx, initial)

	doubleInc := func() {
		inc()
		inc()
	}

	doubleDec := func() {
		dec()
		dec()
	}

	return count, doubleInc, doubleDec
}

// BenchmarkComposableChain measures chained composable overhead
// benchDoubleCounter -> benchCounter -> UseState
// Target: < 500ns combined
func BenchmarkComposableChain(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count, inc, dec := benchDoubleCounter(ctx, 0)
		_, _, _ = count, inc, dec
	}
}

// BenchmarkComposableChain_Execution measures executing chained composable functions
func BenchmarkComposableChain_Execution(b *testing.B) {
	ctx := bubbly.NewTestContext()
	_, inc, _ := benchDoubleCounter(ctx, 0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inc()
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

// BenchmarkMemory_UseStateAllocation measures pure allocation overhead
func BenchmarkMemory_UseStateAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := bubbly.NewTestContext()
		_ = UseState(ctx, i)
	}
}

// BenchmarkMemory_MultipleComposables measures memory with multiple composables
func BenchmarkMemory_MultipleComposables(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := bubbly.NewTestContext()
		_ = UseState(ctx, 0)
		_ = UseState(ctx, "test")
		_ = UseState(ctx, true)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// MemoryStorage is a simple in-memory storage for benchmarks
type MemoryStorage struct {
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

func (m *MemoryStorage) Load(key string) ([]byte, error) {
	if data, ok := m.data[key]; ok {
		return data, nil
	}
	return nil, nil
}

func (m *MemoryStorage) Save(key string, data []byte) error {
	m.data[key] = data
	return nil
}

// ============================================================================
// Multi-CPU Scaling Benchmarks
// ============================================================================

// BenchmarkUseState_MultiCPU tests UseState performance with different CPU counts
// to identify scaling characteristics
func BenchmarkUseState_MultiCPU(b *testing.B) {
	RunMultiCPU(b, func(b *testing.B) {
		ctx := bubbly.NewTestContext()
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			state := UseState(ctx, i)
			_ = state
		}
	}, []int{1, 2, 4, 8})
}

// BenchmarkUseForm_MultiCPU tests UseForm performance with different CPU counts
func BenchmarkUseForm_MultiCPU(b *testing.B) {
	type TestForm struct {
		Name  string
		Email string
		Age   int
	}

	RunMultiCPU(b, func(b *testing.B) {
		ctx := bubbly.NewTestContext()
		validator := func(f TestForm) map[string]string {
			return make(map[string]string)
		}
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			form := UseForm(ctx, TestForm{}, validator)
			_ = form
		}
	}, []int{1, 2, 4, 8})
}

// BenchmarkProvideInject_MultiCPU tests provide/inject with different CPU counts
// This is particularly important for concurrent component trees
func BenchmarkProvideInject_MultiCPU(b *testing.B) {
	RunMultiCPU(b, func(b *testing.B) {
		// Create 5-level tree
		level0 := bubbly.NewTestContext()
		level0.Provide("key", "value")

		level1 := bubbly.NewTestContext()
		bubbly.SetParent(level1, level0)

		level2 := bubbly.NewTestContext()
		bubbly.SetParent(level2, level1)

		level3 := bubbly.NewTestContext()
		bubbly.SetParent(level3, level2)

		level4 := bubbly.NewTestContext()
		bubbly.SetParent(level4, level3)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = level4.Inject("key", "default")
		}
	}, []int{1, 2, 4, 8})
}

// ============================================================================
// Memory Growth Benchmarks
// ============================================================================

// BenchmarkMemoryGrowth_UseState measures memory growth for repeated UseState creation
// This detects memory leaks from composable creation
func BenchmarkMemoryGrowth_UseState(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseState(ctx, 42)
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)

	// Report as metric for tracking
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_UseForm measures memory growth for repeated UseForm creation
func BenchmarkMemoryGrowth_UseForm(b *testing.B) {
	type TestForm struct {
		Name  string
		Email string
		Age   int
	}

	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseForm(ctx, TestForm{}, func(f TestForm) map[string]string {
			return make(map[string]string)
		})
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_ManyComposables measures memory growth with multiple composables
// This simulates a real application with many components
func BenchmarkMemoryGrowth_ManyComposables(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseState(ctx, 0)
		_ = UseState(ctx, "test")
		_ = UseState(ctx, true)
		_ = UseAsync(ctx, func() (*int, error) {
			result := 42
			return &result, nil
		})
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_LongRunning measures memory growth over extended period
// This detects slow memory leaks that only appear in long-running applications
func BenchmarkMemoryGrowth_LongRunning(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping long-running memory growth test in short mode")
	}

	start, end, growth := MeasureMemoryGrowth(b, 2*time.Second, func() {
		ctx := bubbly.NewTestContext()
		state := UseState(ctx, 0)

		// Simulate typical usage: create, update, read
		state.Set(state.Get() + 1)
		_ = state.Get()
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")

	// Memory growth is expected to be minimal for well-behaved code
	// Allow up to 100KB total growth for long-running test
	if growth > 100000 {
		b.Errorf("Excessive total memory growth: %d bytes", growth)
	}
}

// BenchmarkMemoryGrowth_WithCleanup measures memory with proper cleanup
// This verifies that cleanup functions prevent memory leaks
func BenchmarkMemoryGrowth_WithCleanup(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()

		UseEffect(ctx, func() UseEffectCleanup {
			// Effect with cleanup
			return func() {
				// Cleanup runs when effect is removed
			}
		})
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// ============================================================================
// Statistical Benchmarks
// ============================================================================

// BenchmarkWithStats_UseState demonstrates using RunWithStats for detailed metrics
func BenchmarkWithStats_UseState(b *testing.B) {
	ctx := bubbly.NewTestContext()

	RunWithStats(b, func() {
		state := UseState(ctx, 42)
		state.Set(100)
		_ = state.Get()
	})
}

// BenchmarkWithStats_ComposableChain demonstrates stats for complex operations
func BenchmarkWithStats_ComposableChain(b *testing.B) {
	ctx := bubbly.NewTestContext()

	RunWithStats(b, func() {
		count, inc, _ := benchDoubleCounter(ctx, 0)
		inc()
		_ = count.Get()
	})
}
