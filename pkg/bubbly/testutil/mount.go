package testutil

import (
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// EventInspector provides access to component events for testing.
// This is a minimal stub - full implementation will be provided in Task 3.3.
type EventInspector struct {
	tracker *EventTracker
}

// NewEventInspector creates a new event inspector.
func NewEventInspector(tracker *EventTracker) *EventInspector {
	return &EventInspector{
		tracker: tracker,
	}
}

// ComponentTest wraps a mounted component for testing.
// It provides access to the component, state inspector, and event inspector.
//
// ComponentTest is created by TestHarness.Mount() and provides a convenient
// interface for testing component behavior, state changes, and events.
//
// Example usage:
//
//	harness := testutil.NewHarness(t)
//	ct := harness.Mount(createMyComponent())
//
//	// Access component
//	view := ct.component.View()
//
//	// Access state (when StateInspector is fully implemented)
//	// value := ct.state.GetRef("myRef")
//
//	// Unmount when done (or let harness cleanup handle it)
//	ct.Unmount()
type ComponentTest struct {
	harness   *TestHarness
	component bubbly.Component
	state     *StateInspector
	events    *EventInspector

	// onUnmount is called when Unmount() is called (for testing)
	onUnmount func()
	unmounted bool
}

// Mount mounts a component in the test environment.
// It initializes the component by calling Init(), creates state and event
// inspectors, and registers cleanup to unmount the component when the test completes.
//
// The component should be created with all necessary props before mounting,
// as props are immutable after component creation.
//
// Example:
//
//	harness := testutil.NewHarness(t)
//
//	component := bubbly.NewComponent("Counter").
//	    Setup(func(ctx *bubbly.Context) {
//	        count := ctx.Ref(0)
//	        ctx.Expose("count", count)
//	    }).
//	    Template(func(ctx bubbly.RenderContext) string {
//	        return "Counter"
//	    }).
//	    Build()
//
//	ct := harness.Mount(component)
//	// Component is now initialized and ready for testing
//
// The props parameter is reserved for future use and currently ignored.
func (h *TestHarness) Mount(component bubbly.Component, props ...interface{}) *ComponentTest {
	// Store component in harness
	h.component = component

	// Initialize component (calls Setup function)
	component.Init()

	// Extract refs from component state using reflection
	// This is necessary because the component's state map is private
	extractRefsFromComponent(component, h.refs)

	// Create state inspector with harness refs
	// TODO: Extract computed values and watchers from component in future tasks
	stateInspector := NewStateInspector(h.refs, nil, nil)

	// Create event inspector with harness event tracker
	eventInspector := NewEventInspector(h.events)

	// Create component test wrapper
	ct := &ComponentTest{
		harness:   h,
		component: component,
		state:     stateInspector,
		events:    eventInspector,
		unmounted: false,
	}

	// Register unmount cleanup with harness
	h.RegisterCleanup(func() {
		ct.Unmount()
	})

	return ct
}

// extractRefsFromComponent uses reflection to extract refs from the component's state.
// This is a workaround since the component's state map is private.
func extractRefsFromComponent(component bubbly.Component, refs map[string]*bubbly.Ref[interface{}]) {
	// Use reflection to access the private state field
	// This is safe for testing purposes
	v := reflect.ValueOf(component)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Find the state field
	stateField := v.FieldByName("state")
	if !stateField.IsValid() || stateField.IsNil() {
		return
	}

	// Make the field accessible (it's unexported)
	stateField = reflect.NewAt(stateField.Type(), stateField.Addr().UnsafePointer()).Elem()

	// Access the state map
	if stateField.Kind() == reflect.Map {
		// Iterate over the map
		iter := stateField.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			value := iter.Value().Interface()

			// Check if the value is a Ref
			if ref, ok := value.(*bubbly.Ref[interface{}]); ok {
				refs[key] = ref
			}
		}
	}
}

// Unmount unmounts the component and performs cleanup.
// This method is idempotent - calling it multiple times will only
// execute cleanup once.
//
// Unmount is automatically called by the test harness cleanup when
// the test completes, but can also be called manually if needed.
//
// Example:
//
//	ct := harness.Mount(component)
//	// ... test code ...
//	ct.Unmount() // Manual unmount
func (ct *ComponentTest) Unmount() {
	// Idempotent - only unmount once
	if ct.unmounted {
		return
	}

	ct.unmounted = true

	// Call onUnmount callback if set (for testing)
	if ct.onUnmount != nil {
		ct.onUnmount()
	}

	// Future: Call component unmount/cleanup if Component interface exposes it
	// For now, cleanup is handled by harness.Cleanup()
}
