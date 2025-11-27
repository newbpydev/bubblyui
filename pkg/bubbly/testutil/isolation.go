package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestIsolation provides test isolation by saving and restoring global state.
// This ensures tests don't interfere with each other through shared global state.
//
// BubblyUI uses some global state for framework-level features:
//   - Framework hooks (for DevTools and testing)
//   - Error reporters (for observability)
//
// TestIsolation captures the current state of these globals, clears them for
// test isolation, and restores them when the test completes.
//
// Example usage:
//
//	func TestMyComponent(t *testing.T) {
//	    isolation := testutil.NewTestIsolation()
//	    isolation.Isolate(t)  // Automatically restores on test completion
//
//	    // Test code here - isolated from global state
//	}
//
// The Isolate method automatically registers cleanup with t.Cleanup() to
// restore state when the test completes, ensuring proper cleanup even if
// the test panics or fails.
type TestIsolation struct {
	// savedGlobals stores the saved global state values.
	// Keys are descriptive names like "frameworkHook", "errorReporter".
	// Values are the actual saved objects.
	savedGlobals map[string]interface{}
}

// NewTestIsolation creates a new TestIsolation instance.
// The instance is ready to use but hasn't captured any state yet.
// Call Isolate() to capture and clear global state.
//
// Returns:
//   - *TestIsolation: A new isolation instance with empty saved state
//
// Example:
//
//	isolation := testutil.NewTestIsolation()
//	isolation.Isolate(t)
func NewTestIsolation() *TestIsolation {
	return &TestIsolation{
		savedGlobals: make(map[string]interface{}),
	}
}

// Isolate captures current global state, clears it for test isolation,
// and registers automatic restoration with t.Cleanup().
//
// This method:
//  1. Saves the current framework hook (if any)
//  2. Saves the current error reporter (if any)
//  3. Clears both globals to provide clean test environment
//  4. Registers Restore() with t.Cleanup() for automatic cleanup
//
// After calling Isolate(), the test runs in an isolated environment
// with no global state. When the test completes (pass, fail, or panic),
// t.Cleanup() automatically calls Restore() to restore the original state.
//
// Thread Safety:
//
//	This method is NOT thread-safe. It should only be called from the
//	main test goroutine, typically at the beginning of the test function.
//
// Parameters:
//   - t: The testing.T instance for registering cleanup
//
// Example:
//
//	func TestIsolated(t *testing.T) {
//	    isolation := testutil.NewTestIsolation()
//	    isolation.Isolate(t)
//
//	    // Test runs with clean global state
//	    // Cleanup happens automatically
//	}
func (ti *TestIsolation) Isolate(t *testing.T) {
	// Save framework hook if one is registered
	if hook := bubbly.GetRegisteredHook(); hook != nil {
		ti.savedGlobals["frameworkHook"] = hook
	}

	// Save error reporter if one is set
	if reporter := observability.GetErrorReporter(); reporter != nil {
		ti.savedGlobals["errorReporter"] = reporter
	}

	// Clear global state for isolation
	_ = bubbly.UnregisterHook()
	observability.SetErrorReporter(nil)

	// Register automatic restoration with t.Cleanup
	t.Cleanup(func() {
		ti.Restore()
	})
}

// Restore restores the saved global state.
//
// This method is automatically called by t.Cleanup() when Isolate() is used,
// but can also be called manually if needed.
//
// The method is idempotent - calling it multiple times is safe.
// If no state was saved, calling Restore() is a no-op.
//
// Thread Safety:
//
//	This method is NOT thread-safe. It should only be called from the
//	main test goroutine.
//
// Example:
//
//	isolation := testutil.NewTestIsolation()
//	isolation.Isolate(t)
//	// ... test code ...
//	isolation.Restore()  // Manual restore (optional, t.Cleanup does this)
func (ti *TestIsolation) Restore() {
	// Restore framework hook if one was saved
	if hook, ok := ti.savedGlobals["frameworkHook"]; ok {
		if hook != nil {
			_ = bubbly.RegisterHook(hook.(bubbly.FrameworkHook))
		}
	}

	// Restore error reporter if one was saved
	if reporter, ok := ti.savedGlobals["errorReporter"]; ok {
		if reporter != nil {
			observability.SetErrorReporter(reporter.(observability.ErrorReporter))
		}
	}
}
