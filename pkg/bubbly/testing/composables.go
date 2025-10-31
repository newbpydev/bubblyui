// Package btesting provides testing utilities for BubblyUI composables.
// It includes helpers for creating test contexts, mocking composables,
// and asserting cleanup behavior.
//
// Example usage:
//
//	func TestMyComposable(t *testing.T) {
//	    ctx := btesting.NewTestContext()
//
//	    // Test your composable
//	    state := MyComposable(ctx, "initial")
//	    assert.Equal(t, "initial", state.Get())
//
//	    // Test cleanup
//	    cleanup := MyComposableWithCleanup(ctx)
//	    btesting.AssertComposableCleanup(t, cleanup)
//	}
package btesting

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// NewTestContext creates a minimal Context suitable for testing composables.
// The returned context supports all standard Context operations including:
//   - Reactive primitives (Ref, Computed, Watch)
//   - Event handling (On, Emit)
//   - State management (Expose, Get)
//   - Lifecycle hooks (OnMounted, OnUpdated, OnUnmounted)
//   - Dependency injection (Provide, Inject)
//
// Unlike a production component, the test context:
//   - Has no Bubbletea integration
//   - Has no template or rendering
//   - Has no parent by default (use SetParent to establish relationships)
//
// Example:
//
//	ctx := btesting.NewTestContext()
//	count := ctx.Ref(0)
//	ctx.OnMounted(func() {
//	    count.Set(10)
//	})
//	btesting.TriggerMount(ctx)
//	// count.Get() == 10
func NewTestContext() *bubbly.Context {
	return bubbly.NewTestContext()
}

// MockComposable creates a mock composable state for testing.
// It returns a UseStateReturn[T] that can be used in tests without
// calling the real UseState implementation.
//
// This is useful for:
//   - Testing composables that depend on other composables
//   - Isolating the code under test
//   - Providing controlled test data
//
// Example:
//
//	ctx := btesting.NewTestContext()
//	mockUser := btesting.MockComposable(ctx, User{Name: "Alice"})
//
//	// Use mockUser in tests
//	assert.Equal(t, "Alice", mockUser.Get().Name)
//	mockUser.Set(User{Name: "Bob"})
//	assert.Equal(t, "Bob", mockUser.Get().Name)
func MockComposable[T any](ctx *bubbly.Context, value T) composables.UseStateReturn[T] {
	ref := bubbly.NewRef(value)

	return composables.UseStateReturn[T]{
		Value: ref,
		Set:   func(v T) { ref.Set(v) },
		Get:   func() T { return ref.GetTyped() },
	}
}

// AssertComposableCleanup verifies that a cleanup function executes without panicking.
// It's a test helper that marks itself with t.Helper() for better error reporting.
//
// The function:
//   - Handles nil cleanup gracefully (no-op)
//   - Catches panics and reports them as test errors
//   - Marks itself as a helper for cleaner test output
//
// Note: This helper only verifies that cleanup doesn't panic. Verifying that
// cleanup actually cleans up resources is test-specific and should be done
// in the individual test.
//
// Example:
//
//	cleanup := UseEventListener(ctx, "click", handler)
//	btesting.AssertComposableCleanup(t, cleanup)
//
//	// Then verify handler no longer executes...
func AssertComposableCleanup(t *testing.T, cleanup func()) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("cleanup panicked: %v", r)
		}
	}()

	if cleanup != nil {
		cleanup()
	}
}

// TriggerMount simulates the component mount lifecycle event for testing.
// This executes any onMounted hooks registered on the context.
//
// Example:
//
//	ctx := btesting.NewTestContext()
//	mounted := false
//	ctx.OnMounted(func() { mounted = true })
//	btesting.TriggerMount(ctx)
//	// mounted == true
func TriggerMount(ctx *bubbly.Context) {
	bubbly.TriggerMount(ctx)
}

// TriggerUpdate simulates the component update lifecycle event for testing.
// This executes any onUpdated hooks registered on the context.
//
// Example:
//
//	ctx := btesting.NewTestContext()
//	updated := false
//	ctx.OnUpdated(func() { updated = true })
//	btesting.TriggerUpdate(ctx)
//	// updated == true
func TriggerUpdate(ctx *bubbly.Context) {
	bubbly.TriggerUpdate(ctx)
}

// TriggerUnmount simulates the component unmount lifecycle event for testing.
// This executes any onUnmounted hooks registered on the context.
//
// Example:
//
//	ctx := btesting.NewTestContext()
//	unmounted := false
//	ctx.OnUnmounted(func() { unmounted = true })
//	btesting.TriggerUnmount(ctx)
//	// unmounted == true
func TriggerUnmount(ctx *bubbly.Context) {
	bubbly.TriggerUnmount(ctx)
}

// SetParent establishes a parent-child relationship between two test contexts.
// This is needed for testing provide/inject functionality across component trees.
//
// Example:
//
//	parent := btesting.NewTestContext()
//	child := btesting.NewTestContext()
//	btesting.SetParent(child, parent)
//
//	parent.Provide("theme", "dark")
//	theme := child.Inject("theme", "light")
//	// theme == "dark"
func SetParent(child, parent *bubbly.Context) {
	bubbly.SetParent(child, parent)
}
