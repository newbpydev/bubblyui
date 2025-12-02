// Package btesting provides testing helpers for BubblyUI components and composables.
//
// This package contains utilities for testing BubblyUI components including
// mock contexts, lifecycle triggers, and composable helpers.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/testing,
// providing a cleaner import path for users.
//
// # Features
//
//   - Test context creation for isolated testing
//   - Lifecycle hook triggers (mount, update, unmount)
//   - Mock composables for unit testing
//   - Cleanup assertion helpers
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/testing/btesting"
//
//	func TestMyComponent(t *testing.T) {
//	    ctx := btesting.NewTestContext()
//
//	    // Test composable
//	    state := btesting.MockComposable(ctx, 42)
//	    assert.Equal(t, 42, state.Get())
//
//	    // Trigger lifecycle
//	    btesting.TriggerMount(ctx)
//	    btesting.TriggerUpdate(ctx)
//	    btesting.TriggerUnmount(ctx)
//	}
package btesting

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	btestingpkg "github.com/newbpydev/bubblyui/pkg/bubbly/testing"
)

// =============================================================================
// Context Creation
// =============================================================================

// NewTestContext creates a new test context for isolated testing.
var NewTestContext = btestingpkg.NewTestContext

// SetParent sets the parent context for a child context.
var SetParent = btestingpkg.SetParent

// =============================================================================
// Lifecycle Triggers
// =============================================================================

// TriggerMount triggers the mount lifecycle for a context.
var TriggerMount = btestingpkg.TriggerMount

// TriggerUpdate triggers the update lifecycle for a context.
var TriggerUpdate = btestingpkg.TriggerUpdate

// TriggerUnmount triggers the unmount lifecycle for a context.
var TriggerUnmount = btestingpkg.TriggerUnmount

// =============================================================================
// Mock Composables
// =============================================================================

// MockComposable creates a mock composable state for testing.
func MockComposable[T any](ctx *bubbly.Context, value T) composables.UseStateReturn[T] {
	return btestingpkg.MockComposable(ctx, value)
}

// =============================================================================
// Assertions
// =============================================================================

// AssertComposableCleanup asserts that a cleanup function was called.
func AssertComposableCleanup(t *testing.T, cleanup func()) {
	btestingpkg.AssertComposableCleanup(t, cleanup)
}
