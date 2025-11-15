package testutil

import (
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestHarness provides a testing environment for BubblyUI components.
// It manages component lifecycle, state inspection, event tracking, and cleanup.
//
// The harness automatically registers cleanup with testing.T to ensure
// proper resource cleanup when tests complete.
//
// Example usage:
//
//	func TestMyComponent(t *testing.T) {
//	    harness := testutil.NewHarness(t)
//	    // Test code here
//	    // Cleanup happens automatically
//	}
type TestHarness struct {
	t         *testing.T
	component bubbly.Component
	refs      map[string]*bubbly.Ref[interface{}]
	events    *EventTracker
	cleanup   []func()
	cleanupMu sync.Mutex
}

// HarnessOption is a functional option for configuring a TestHarness.
type HarnessOption func(*TestHarness)

// EventTracker is a minimal stub for tracking events.
// Full implementation will be provided in Task 3.3.
type EventTracker struct {
	// Placeholder for now
}

// NewEventTracker creates a new event tracker.
func NewEventTracker() *EventTracker {
	return &EventTracker{}
}

// NewHarness creates a new test harness for component testing.
// It initializes the harness with empty state and registers automatic
// cleanup with the provided testing.T.
//
// Options can be provided to customize harness behavior:
//
//	harness := testutil.NewHarness(t, WithIsolation(), WithTimeout(5*time.Second))
//
// The harness automatically cleans up resources when the test completes.
func NewHarness(t *testing.T, opts ...HarnessOption) *TestHarness {
	h := &TestHarness{
		t:       t,
		refs:    make(map[string]*bubbly.Ref[interface{}]),
		events:  NewEventTracker(),
		cleanup: []func(){},
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	// Register automatic cleanup with testing.T
	t.Cleanup(func() {
		h.Cleanup()
	})

	return h
}

// RegisterCleanup registers a cleanup function to be called when the test completes.
// Cleanup functions are executed in LIFO order (last registered runs first),
// similar to defer statements.
//
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	harness.RegisterCleanup(func() {
//	    // Cleanup resources
//	})
func (h *TestHarness) RegisterCleanup(fn func()) {
	h.cleanupMu.Lock()
	defer h.cleanupMu.Unlock()

	h.cleanup = append(h.cleanup, fn)
}

// Cleanup executes all registered cleanup functions in LIFO order
// (last registered runs first). After execution, the cleanup slice is cleared.
//
// This method is idempotent - calling it multiple times will only execute
// cleanup functions once.
//
// If a cleanup function panics, the panic is recovered and cleanup continues
// with the remaining functions. This ensures all cleanup happens even if
// one function fails.
//
// Cleanup is automatically called by testing.T.Cleanup() when the test completes,
// but can also be called manually if needed.
func (h *TestHarness) Cleanup() {
	h.cleanupMu.Lock()
	defer h.cleanupMu.Unlock()

	// If cleanup slice is empty, nothing to do (idempotent)
	if len(h.cleanup) == 0 {
		return
	}

	// Execute cleanup functions in LIFO order (reverse)
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		func() {
			// Recover from panics to ensure all cleanup functions run
			defer func() {
				if r := recover(); r != nil {
					// Log panic but continue with remaining cleanup
					h.t.Logf("cleanup function panicked: %v", r)
				}
			}()

			h.cleanup[i]()
		}()
	}

	// Clear cleanup slice after execution
	h.cleanup = []func(){}
}
