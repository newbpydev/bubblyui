package testutil

import (
	"sync"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// testingT is an interface that matches the methods we need from testing.T.
// This allows for mocking in tests while maintaining compatibility with real testing.T.
type testingT interface {
	Errorf(format string, args ...interface{})
	Helper()
	Logf(format string, args ...interface{})
	Cleanup(func())
}

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
	t         testingT
	component bubbly.Component
	refs      map[string]*bubbly.Ref[interface{}]
	events    *EventTracker
	hooks     *TestHooks
	cleanup   []func()
	cleanupMu sync.Mutex
}

// HarnessOption is a functional option for configuring a TestHarness.
type HarnessOption func(*TestHarness)

// EventTracker tracks events emitted during component testing.
// It provides thread-safe event tracking with methods to query event history.
//
// EventTracker is used internally by the test harness to track all events
// emitted by components during tests, enabling assertions on event behavior.
type EventTracker struct {
	events []EmittedEvent
	mu     sync.RWMutex
}

// EmittedEvent represents an event that was emitted during testing.
type EmittedEvent struct {
	Name      string
	Payload   interface{}
	Timestamp time.Time
	Source    string
}

// NewEventTracker creates a new event tracker.
func NewEventTracker() *EventTracker {
	return &EventTracker{
		events: []EmittedEvent{},
	}
}

// Track records an event emission.
// This method is thread-safe and can be called concurrently.
//
// Parameters:
//   - name: The name of the event
//   - payload: The event payload (can be nil)
//   - source: The source component that emitted the event
func (et *EventTracker) Track(name string, payload interface{}, source string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.events = append(et.events, EmittedEvent{
		Name:      name,
		Payload:   payload,
		Timestamp: time.Now(),
		Source:    source,
	})
}

// GetEvents returns all events with the given name.
// Returns an empty slice if no events with that name were tracked.
// This method is thread-safe.
func (et *EventTracker) GetEvents(name string) []EmittedEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	events := []EmittedEvent{}
	for _, e := range et.events {
		if e.Name == name {
			events = append(events, e)
		}
	}

	return events
}

// WasFired returns true if at least one event with the given name was tracked.
// This method is thread-safe.
func (et *EventTracker) WasFired(name string) bool {
	return len(et.GetEvents(name)) > 0
}

// FiredCount returns the number of times an event with the given name was tracked.
// Returns 0 if no events with that name were tracked.
// This method is thread-safe.
func (et *EventTracker) FiredCount(name string) int {
	return len(et.GetEvents(name))
}

// testHook implements bubbly.FrameworkHook to track events for testing.
type testHook struct {
	tracker *EventTracker
	harness *TestHarness
}

func (h *testHook) OnComponentMount(id, name string)                         {}
func (h *testHook) OnComponentUpdate(id string, msg interface{})             {}
func (h *testHook) OnComponentUnmount(id string)                             {}
func (h *testHook) OnRefChange(id string, oldValue, newValue interface{})    {}
func (h *testHook) OnRefExposed(componentID, refName, refID string) {
	// Refs are exposed during Init(), so we can't access them here yet
	// They will be extracted in Mount() after Init() completes
	// This hook is primarily for DevTools tracking, not for test harness
}
func (h *testHook) OnRenderComplete(componentID string, duration time.Duration) {}
func (h *testHook) OnComputedChange(id string, oldValue, newValue interface{}) {}
func (h *testHook) OnWatchCallback(watcherID string, oldValue, newValue interface{}) {}
func (h *testHook) OnEffectRun(effectID string)                              {}
func (h *testHook) OnChildAdded(parentID, childID string)                    {}
func (h *testHook) OnChildRemoved(parentID, childID string)                  {}

func (h *testHook) OnEvent(componentID, eventName string, data interface{}) {
	h.tracker.Track(eventName, data, componentID)
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

	// Install framework hook to track events
	hook := &testHook{tracker: h.events}
	bubbly.RegisterHook(hook)

	// Register cleanup to unregister hook
	h.RegisterCleanup(func() {
		bubbly.UnregisterHook()
	})

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
