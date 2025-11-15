package testutil

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// StateInspector provides access to component state for testing.
// It allows tests to inspect and manipulate reactive refs, computed values,
// and watchers within a component, enabling verification of state changes and behavior.
//
// StateInspector is created by TestHarness.Mount() and provides methods to:
//   - Get refs by name
//   - Read ref values
//   - Update ref values
//   - Get computed values by name
//   - Read computed values
//   - Access watchers
//
// Example usage:
//
//	harness := testutil.NewHarness(t)
//	ct := harness.Mount(createMyComponent())
//
//	// Get a ref
//	countRef := ct.state.GetRef("count")
//	assert.NotNil(t, countRef)
//
//	// Get ref value
//	value := ct.state.GetRefValue("count")
//	assert.Equal(t, 0, value)
//
//	// Set ref value
//	ct.state.SetRefValue("count", 42)
//	assert.Equal(t, 42, ct.state.GetRefValue("count"))
//
//	// Get computed value
//	doubled := ct.state.GetComputedValue("doubled")
//	assert.Equal(t, 84, doubled)
type StateInspector struct {
	refs     map[string]*bubbly.Ref[interface{}]
	computed map[string]*bubbly.Computed[interface{}]
	watchers map[string]bubbly.WatchCleanup
}

// NewStateInspector creates a new state inspector with the given refs, computed values, and watchers.
// These maps should contain all exposed state from the component's Setup function.
//
// Parameters:
//   - refs: Map of ref names to Ref instances (can be nil or empty)
//   - computed: Map of computed names to Computed instances (can be nil or empty)
//   - watchers: Map of watcher names to WatchCleanup functions (can be nil or empty)
//
// Returns:
//   - *StateInspector: A new state inspector instance
//
// Example:
//
//	refs := map[string]*bubbly.Ref[interface{}]{
//	    "count": bubbly.NewRef[interface{}](0),
//	}
//	computed := map[string]*bubbly.Computed[interface{}]{
//	    "doubled": bubbly.NewComputed(func() interface{} { return count.Get().(int) * 2 }),
//	}
//	inspector := NewStateInspector(refs, computed, nil)
func NewStateInspector(
	refs map[string]*bubbly.Ref[interface{}],
	computed map[string]*bubbly.Computed[interface{}],
	watchers map[string]bubbly.WatchCleanup,
) *StateInspector {
	return &StateInspector{
		refs:     refs,
		computed: computed,
		watchers: watchers,
	}
}

// GetRef retrieves a ref by name.
// Returns nil if the ref doesn't exist.
//
// This method is useful when you need direct access to the Ref instance,
// for example to check if it exists or to pass it to other functions.
//
// Parameters:
//   - name: The name of the ref to retrieve
//
// Returns:
//   - *bubbly.Ref[interface{}]: The ref if found, nil otherwise
//
// Example:
//
//	ref := inspector.GetRef("count")
//	if ref != nil {
//	    value := ref.Get()
//	    fmt.Printf("Count: %v\n", value)
//	}
func (si *StateInspector) GetRef(name string) *bubbly.Ref[interface{}] {
	if si.refs == nil {
		return nil
	}
	return si.refs[name]
}

// GetRefValue retrieves the current value of a ref by name.
// Panics if the ref doesn't exist.
//
// This is a convenience method that combines GetRef and Get() into a single call.
// Use this when you know the ref exists and want to access its value directly.
//
// Parameters:
//   - name: The name of the ref to retrieve
//
// Returns:
//   - interface{}: The current value of the ref
//
// Panics:
//   - If the ref with the given name doesn't exist
//
// Example:
//
//	value := inspector.GetRefValue("count")
//	assert.Equal(t, 42, value)
func (si *StateInspector) GetRefValue(name string) interface{} {
	ref := si.GetRef(name)
	if ref == nil {
		panic(fmt.Sprintf("ref %q not found", name))
	}
	return ref.Get()
}

// SetRefValue updates the value of a ref by name.
// Panics if the ref doesn't exist.
//
// This is a convenience method that combines GetRef and Set() into a single call.
// The ref's watchers will be notified of the change, and any reactive dependencies
// will be updated.
//
// Parameters:
//   - name: The name of the ref to update
//   - value: The new value to set
//
// Panics:
//   - If the ref with the given name doesn't exist
//
// Example:
//
//	inspector.SetRefValue("count", 100)
//	assert.Equal(t, 100, inspector.GetRefValue("count"))
func (si *StateInspector) SetRefValue(name string, value interface{}) {
	ref := si.GetRef(name)
	if ref == nil {
		panic(fmt.Sprintf("ref %q not found", name))
	}
	ref.Set(value)
}

// GetComputed retrieves a computed value by name.
// Returns nil if the computed doesn't exist.
//
// This method is useful when you need direct access to the Computed instance,
// for example to check if it exists or to pass it to other functions.
//
// Parameters:
//   - name: The name of the computed to retrieve
//
// Returns:
//   - *bubbly.Computed[interface{}]: The computed if found, nil otherwise
//
// Example:
//
//	computed := inspector.GetComputed("doubled")
//	if computed != nil {
//	    value := computed.Get()
//	    fmt.Printf("Doubled: %v\n", value)
//	}
func (si *StateInspector) GetComputed(name string) *bubbly.Computed[interface{}] {
	if si.computed == nil {
		return nil
	}
	return si.computed[name]
}

// GetComputedValue retrieves the current value of a computed by name.
// Panics if the computed doesn't exist.
//
// This is a convenience method that combines GetComputed and Get() into a single call.
// Use this when you know the computed exists and want to access its value directly.
//
// Parameters:
//   - name: The name of the computed to retrieve
//
// Returns:
//   - interface{}: The current value of the computed
//
// Panics:
//   - If the computed with the given name doesn't exist
//
// Example:
//
//	value := inspector.GetComputedValue("doubled")
//	assert.Equal(t, 84, value)
func (si *StateInspector) GetComputedValue(name string) interface{} {
	computed := si.GetComputed(name)
	if computed == nil {
		panic(fmt.Sprintf("computed %q not found", name))
	}
	return computed.Get()
}

// GetWatcher retrieves a watcher cleanup function by name.
// Returns nil if the watcher doesn't exist.
//
// This method is useful for managing watchers during tests, such as
// cleaning up watchers early or verifying watcher registration.
//
// Parameters:
//   - name: The name of the watcher to retrieve
//
// Returns:
//   - bubbly.WatchCleanup: The cleanup function if found, nil otherwise
//
// Example:
//
//	cleanup := inspector.GetWatcher("countWatcher")
//	if cleanup != nil {
//	    cleanup() // Stop watching
//	}
func (si *StateInspector) GetWatcher(name string) bubbly.WatchCleanup {
	if si.watchers == nil {
		return nil
	}
	return si.watchers[name]
}

// HasRef checks if a ref with the given name exists.
//
// Parameters:
//   - name: The name of the ref to check
//
// Returns:
//   - bool: true if the ref exists, false otherwise
//
// Example:
//
//	if inspector.HasRef("count") {
//	    value := inspector.GetRefValue("count")
//	}
func (si *StateInspector) HasRef(name string) bool {
	return si.GetRef(name) != nil
}

// HasComputed checks if a computed with the given name exists.
//
// Parameters:
//   - name: The name of the computed to check
//
// Returns:
//   - bool: true if the computed exists, false otherwise
//
// Example:
//
//	if inspector.HasComputed("doubled") {
//	    value := inspector.GetComputedValue("doubled")
//	}
func (si *StateInspector) HasComputed(name string) bool {
	return si.GetComputed(name) != nil
}

// HasWatcher checks if a watcher with the given name exists.
//
// Parameters:
//   - name: The name of the watcher to check
//
// Returns:
//   - bool: true if the watcher exists, false otherwise
//
// Example:
//
//	if inspector.HasWatcher("countWatcher") {
//	    cleanup := inspector.GetWatcher("countWatcher")
//	    cleanup()
//	}
func (si *StateInspector) HasWatcher(name string) bool {
	return si.GetWatcher(name) != nil
}
