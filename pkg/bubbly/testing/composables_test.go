package btesting

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/stretchr/testify/assert"
)

// TestNewTestContext_CreatesValidContext verifies that NewTestContext creates a functional context
func TestNewTestContext_CreatesValidContext(t *testing.T) {
	ctx := NewTestContext()

	// Should not be nil
	assert.NotNil(t, ctx, "context should not be nil")

	// Should be able to create Refs
	ref := ctx.Ref(42)
	assert.NotNil(t, ref, "should create ref")
	assert.Equal(t, 42, ref.Get(), "ref should have initial value")

	// Should be able to create Computed
	computed := ctx.Computed(func() interface{} {
		return ref.Get().(int) * 2
	})
	assert.NotNil(t, computed, "should create computed")
	assert.Equal(t, 84, computed.Get(), "computed should calculate correctly")

	// Should be able to Expose/Get values
	ctx.Expose("test", ref)
	retrieved := ctx.Get("test")
	assert.Equal(t, ref, retrieved, "should retrieve exposed value")
}

// TestNewTestContext_SupportsEventHandlers verifies event handling in test context
func TestNewTestContext_SupportsEventHandlers(t *testing.T) {
	ctx := NewTestContext()

	called := false
	ctx.On("test", func(data interface{}) {
		called = true
	})

	ctx.Emit("test", nil)
	assert.True(t, called, "event handler should be called")
}

// TestNewTestContext_SupportsLifecycleHooks verifies lifecycle hooks in test context
func TestNewTestContext_SupportsLifecycleHooks(t *testing.T) {
	ctx := NewTestContext()

	mountedCalled := false
	ctx.OnMounted(func() {
		mountedCalled = true
	})

	// Trigger mount by calling the test helper
	TriggerMount(ctx)
	assert.True(t, mountedCalled, "onMounted should be called")
}

// TestNewTestContext_SupportsProvideInject verifies provide/inject in test context
func TestNewTestContext_SupportsProvideInject(t *testing.T) {
	parentCtx := NewTestContext()
	childCtx := NewTestContext()

	// Set up parent-child relationship
	SetParent(childCtx, parentCtx)

	// Provide value in parent
	parentCtx.Provide("theme", "dark")

	// Inject in child
	theme := childCtx.Inject("theme", "light")
	assert.Equal(t, "dark", theme, "should inject from parent")
}

// TestNewTestContext_SupportsWatch verifies Watch functionality in test context
func TestNewTestContext_SupportsWatch(t *testing.T) {
	ctx := NewTestContext()

	ref := ctx.Ref(0)
	watchCalled := false
	var newVal, oldVal interface{}

	ctx.Watch(ref, func(nv, ov interface{}) {
		watchCalled = true
		newVal = nv
		oldVal = ov
	})

	ref.Set(42)
	time.Sleep(10 * time.Millisecond) // Give watcher time to trigger

	assert.True(t, watchCalled, "watch callback should be called")
	assert.Equal(t, 42, newVal, "new value should be correct")
	assert.Equal(t, 0, oldVal, "old value should be correct")
}

// TestMockComposable_ReturnsCorrectStructure verifies MockComposable returns valid UseStateReturn
func TestMockComposable_ReturnsCorrectStructure(t *testing.T) {
	ctx := NewTestContext()

	state := MockComposable(ctx, 42)

	assert.NotNil(t, state.Value, "Value should not be nil")
	assert.NotNil(t, state.Set, "Set should not be nil")
	assert.NotNil(t, state.Get, "Get should not be nil")
}

// TestMockComposable_ValueCanBeRead verifies MockComposable value can be accessed
func TestMockComposable_ValueCanBeRead(t *testing.T) {
	ctx := NewTestContext()

	state := MockComposable(ctx, 42)

	// Access via Value field
	assert.Equal(t, 42, state.Value.Get(), "should read value via Value field")

	// Access via Get function
	assert.Equal(t, 42, state.Get(), "should read value via Get function")
}

// TestMockComposable_SetFunctionWorks verifies MockComposable Set function updates value
func TestMockComposable_SetFunctionWorks(t *testing.T) {
	ctx := NewTestContext()

	state := MockComposable(ctx, 42)

	state.Set(100)

	assert.Equal(t, 100, state.Get(), "Set should update value")
	assert.Equal(t, 100, state.Value.Get(), "Value field should reflect update")
}

// TestMockComposable_TypeSafety verifies MockComposable maintains type safety
func TestMockComposable_TypeSafety(t *testing.T) {
	ctx := NewTestContext()

	// Int type
	intState := MockComposable(ctx, 42)
	assert.Equal(t, 42, intState.Get())

	// String type
	strState := MockComposable(ctx, "hello")
	assert.Equal(t, "hello", strState.Get())

	// Struct type
	type TestStruct struct {
		Name string
		Age  int
	}
	structState := MockComposable(ctx, TestStruct{Name: "Alice", Age: 30})
	assert.Equal(t, "Alice", structState.Get().Name)
}

// TestAssertComposableCleanup_NilCleanup verifies handling of nil cleanup
func TestAssertComposableCleanup_NilCleanup(t *testing.T) {
	// Should not panic with nil cleanup
	AssertComposableCleanup(t, nil)
}

// TestAssertComposableCleanup_ValidCleanup verifies normal cleanup works
func TestAssertComposableCleanup_ValidCleanup(t *testing.T) {
	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}

	AssertComposableCleanup(t, cleanup)
	assert.True(t, cleanupCalled, "cleanup should be called")
}

// TestAssertComposableCleanup_DoesNotPanic verifies that AssertComposableCleanup
// doesn't crash the test when cleanup panics (it catches and reports)
func TestAssertComposableCleanup_DoesNotPanic(t *testing.T) {
	// This test verifies that AssertComposableCleanup catches panics
	// We can't easily test that it reports errors without failing this test,
	// but we can verify it doesn't crash the test runner

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AssertComposableCleanup should not propagate panics, but got: %v", r)
		}
	}()

	// Multiple calls with various scenarios should not crash
	AssertComposableCleanup(t, nil)
	AssertComposableCleanup(t, func() {})
	// Note: Testing with panicking cleanup would cause test failure (by design)
	// The panic is caught and reported via t.Errorf, which is the correct behavior
}

// TestIntegration_UseStateWithTestContext verifies UseState works with test context
func TestIntegration_UseStateWithTestContext(t *testing.T) {
	ctx := NewTestContext()

	state := composables.UseState(ctx, 0)

	assert.Equal(t, 0, state.Get())

	state.Set(42)
	assert.Equal(t, 42, state.Get())
}

// TestIntegration_UseEffectWithTestContext verifies UseEffect works with test context
func TestIntegration_UseEffectWithTestContext(t *testing.T) {
	ctx := NewTestContext()

	effectCalled := false
	cleanupCalled := false

	count := ctx.Ref(0)

	composables.UseEffect(ctx, func() composables.UseEffectCleanup {
		effectCalled = true
		return func() {
			cleanupCalled = true
		}
	}, count)

	// Trigger mount
	TriggerMount(ctx)
	assert.True(t, effectCalled, "effect should be called on mount")

	// Trigger unmount
	TriggerUnmount(ctx)
	assert.True(t, cleanupCalled, "cleanup should be called on unmount")
}

// TestIntegration_MockComposableInComposable verifies mock can be used in composables
func TestIntegration_MockComposableInComposable(t *testing.T) {
	ctx := NewTestContext()

	// Create a composable that uses another composable
	mockState := MockComposable(ctx, 10)

	// Use the mock in a custom composable pattern
	doubled := ctx.Computed(func() interface{} {
		return mockState.Get() * 2
	})

	assert.Equal(t, 20, doubled.Get())

	mockState.Set(20)
	assert.Equal(t, 40, doubled.Get())
}
