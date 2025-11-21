package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// FixtureBuilder provides a fluent API for creating test fixtures.
// It allows setting props, state, and events before mounting a component,
// reducing test boilerplate and making test setup more readable.
//
// The builder follows the builder pattern with method chaining:
//   - WithProp() sets component props (for future use)
//   - WithState() sets initial state values
//   - WithEvent() queues events to emit after mounting
//   - Build() creates the component, applies configuration, and returns ComponentTest
//
// Example usage:
//
//	fixture := testutil.NewFixture().
//	    WithState("count", 10).
//	    WithState("name", "Alice").
//	    WithEvent("init", nil).
//	    Build(t, createMyComponent)
//
//	// Component is now mounted with count=10, name="Alice", and "init" event emitted
//	assert.Equal(t, 10, fixture.state.GetRefValue("count"))
type FixtureBuilder struct {
	// props stores component props to apply.
	// Note: Props are currently not applied as components don't support
	// prop setting after creation. This is reserved for future use.
	// Components should be created with props in the createFn parameter.
	props map[string]interface{}

	// state stores initial state values to set after mounting.
	// Keys are ref names, values are the initial values to set.
	state map[string]interface{}

	// events stores events to emit after mounting.
	// Keys are event names, values are event payloads.
	events map[string]interface{}
}

// NewFixture creates a new FixtureBuilder with initialized maps.
// This is the entry point for creating test fixtures using the fluent API.
//
// Example:
//
//	fixture := testutil.NewFixture()
//	// Now chain configuration methods...
//	fixture.WithState("count", 0).WithEvent("init", nil).Build(t, createComponent)
//
// Returns:
//   - *FixtureBuilder: A new builder instance ready for configuration
func NewFixture() *FixtureBuilder {
	return &FixtureBuilder{
		props:  make(map[string]interface{}),
		state:  make(map[string]interface{}),
		events: make(map[string]interface{}),
	}
}

// WithProp adds a prop to the fixture configuration.
// Returns the builder for method chaining.
//
// Note: Props are currently not applied as components don't support
// prop setting after creation. This method is provided for future use
// and API consistency. Components should be created with props in the
// createFn parameter passed to Build().
//
// Example:
//
//	fixture := testutil.NewFixture().
//	    WithProp("title", "Test Title").
//	    WithProp("count", 42)
//
// Parameters:
//   - key: The prop name
//   - value: The prop value (any type)
//
// Returns:
//   - *FixtureBuilder: Self for method chaining
func (fb *FixtureBuilder) WithProp(key string, value interface{}) *FixtureBuilder {
	fb.props[key] = value
	return fb
}

// WithState adds a state value to the fixture configuration.
// The state will be applied after the component is mounted by calling
// StateInspector.SetRefValue() for each state entry.
//
// Returns the builder for method chaining.
//
// Example:
//
//	fixture := testutil.NewFixture().
//	    WithState("count", 10).
//	    WithState("name", "Alice")
//
// Parameters:
//   - key: The ref name to set
//   - value: The initial value (any type)
//
// Returns:
//   - *FixtureBuilder: Self for method chaining
func (fb *FixtureBuilder) WithState(key string, value interface{}) *FixtureBuilder {
	fb.state[key] = value
	return fb
}

// WithEvent adds an event to emit after the component is mounted.
// Events are emitted in the order they were added using ComponentTest.Emit().
//
// Returns the builder for method chaining.
//
// Example:
//
//	fixture := testutil.NewFixture().
//	    WithEvent("init", nil).
//	    WithEvent("ready", true)
//
// Parameters:
//   - name: The event name
//   - payload: The event payload (any type)
//
// Returns:
//   - *FixtureBuilder: Self for method chaining
func (fb *FixtureBuilder) WithEvent(name string, payload interface{}) *FixtureBuilder {
	fb.events[name] = payload
	return fb
}

// Build creates a test harness, mounts the component, applies the fixture
// configuration, and returns the ComponentTest for further testing.
//
// The build process:
//  1. Creates a new TestHarness with the provided testing.T
//  2. Calls createFn to create the component
//  3. Mounts the component with the harness
//  4. Applies state values using StateInspector.SetRefValue()
//  5. Emits events using ComponentTest.Emit()
//  6. Returns the ComponentTest for assertions and further interaction
//
// Example:
//
//	createCounter := func() bubbly.Component {
//	    comp, err := bubbly.NewComponent("Counter").
//	        Setup(func(ctx *bubbly.Context) {
//	            ctx.Expose("count", ctx.Ref(0))
//	        }).
//	        Template(func(ctx bubbly.RenderContext) string {
//	            return "Counter"
//	        }).
//	        Build()
//	    if err != nil {
//	        panic(err) // Or handle error appropriately
//	    }
//	    return comp
//	}
//
//	fixture := testutil.NewFixture().
//	    WithState("count", 10).
//	    Build(t, createCounter)
//
//	assert.Equal(t, 10, fixture.state.GetRefValue("count"))
//
// Parameters:
//   - t: The testing.T instance for test context
//   - createFn: Function that creates and returns the component to test
//
// Returns:
//   - *ComponentTest: The mounted component ready for testing
func (fb *FixtureBuilder) Build(t *testing.T, createFn func() bubbly.Component) *ComponentTest {
	// Create test harness
	harness := NewHarness(t)

	// Create component using provided function
	component := createFn()

	// Mount component
	ct := harness.Mount(component)

	// Apply state values
	for key, value := range fb.state {
		ct.state.SetRefValue(key, value)
	}

	// Emit events
	for name, payload := range fb.events {
		ct.component.Emit(name, payload)
	}

	// Note: Props are not applied as components don't currently support
	// prop setting after creation. The props map is reserved for future use.

	return ct
}
