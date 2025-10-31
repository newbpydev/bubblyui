package composables

import (
	"testing"

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

// UseCounter is a simple composable for benchmarking
func UseCounter(ctx *bubbly.Context, initial int) (*bubbly.Ref[int], func(), func()) {
	state := UseState(ctx, initial)

	increment := func() {
		state.Set(state.Get() + 1)
	}

	decrement := func() {
		state.Set(state.Get() - 1)
	}

	return state.Value, increment, decrement
}

// UseDoubleCounter chains UseCounter
func UseDoubleCounter(ctx *bubbly.Context, initial int) (*bubbly.Ref[int], func(), func()) {
	count, inc, dec := UseCounter(ctx, initial)

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
// UseDoubleCounter -> UseCounter -> UseState
// Target: < 500ns combined
func BenchmarkComposableChain(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count, inc, dec := UseDoubleCounter(ctx, 0)
		_, _, _ = count, inc, dec
	}
}

// BenchmarkComposableChain_Execution measures executing chained composable functions
func BenchmarkComposableChain_Execution(b *testing.B) {
	ctx := bubbly.NewTestContext()
	_, inc, _ := UseDoubleCounter(ctx, 0)
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
