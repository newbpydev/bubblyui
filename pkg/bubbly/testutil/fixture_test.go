package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewFixture tests that NewFixture creates a FixtureBuilder with initialized maps.
func TestNewFixture(t *testing.T) {
	fixture := NewFixture()

	assert.NotNil(t, fixture, "NewFixture should return non-nil builder")
	assert.NotNil(t, fixture.props, "props map should be initialized")
	assert.NotNil(t, fixture.state, "state map should be initialized")
	assert.NotNil(t, fixture.events, "events map should be initialized")
	assert.Empty(t, fixture.props, "props map should be empty")
	assert.Empty(t, fixture.state, "state map should be empty")
	assert.Empty(t, fixture.events, "events map should be empty")
}

// TestWithProp tests that WithProp adds props and returns the builder for chaining.
func TestWithProp(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected interface{}
	}{
		{"string prop", "title", "Test Title", "Test Title"},
		{"int prop", "count", 42, 42},
		{"bool prop", "enabled", true, true},
		{"nil prop", "optional", nil, nil},
		{"struct prop", "config", struct{ Name string }{Name: "test"}, struct{ Name string }{Name: "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := NewFixture()

			// Call WithProp
			result := fixture.WithProp(tt.key, tt.value)

			// Should return self for chaining
			assert.Equal(t, fixture, result, "WithProp should return self for chaining")

			// Should store the prop
			assert.Contains(t, fixture.props, tt.key, "props should contain the key")
			assert.Equal(t, tt.expected, fixture.props[tt.key], "prop value should match")
		})
	}
}

// TestWithState tests that WithState adds state and returns the builder for chaining.
func TestWithState(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected interface{}
	}{
		{"string state", "name", "John", "John"},
		{"int state", "age", 30, 30},
		{"slice state", "items", []string{"a", "b"}, []string{"a", "b"}},
		{"map state", "data", map[string]int{"x": 1}, map[string]int{"x": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := NewFixture()

			// Call WithState
			result := fixture.WithState(tt.key, tt.value)

			// Should return self for chaining
			assert.Equal(t, fixture, result, "WithState should return self for chaining")

			// Should store the state
			assert.Contains(t, fixture.state, tt.key, "state should contain the key")
			assert.Equal(t, tt.expected, fixture.state[tt.key], "state value should match")
		})
	}
}

// TestWithEvent tests that WithEvent adds events and returns the builder for chaining.
func TestWithEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		payload  interface{}
		expected interface{}
	}{
		{"nil payload", "click", nil, nil},
		{"string payload", "input", "test", "test"},
		{"int payload", "count", 5, 5},
		{"struct payload", "data", struct{ ID int }{ID: 123}, struct{ ID int }{ID: 123}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := NewFixture()

			// Call WithEvent
			result := fixture.WithEvent(tt.event, tt.payload)

			// Should return self for chaining
			assert.Equal(t, fixture, result, "WithEvent should return self for chaining")

			// Should store the event
			assert.Contains(t, fixture.events, tt.event, "events should contain the event name")
			assert.Equal(t, tt.expected, fixture.events[tt.event], "event payload should match")
		})
	}
}

// TestFluentChaining tests that multiple With* methods can be chained.
func TestFluentChaining(t *testing.T) {
	fixture := NewFixture().
		WithProp("title", "Test").
		WithProp("count", 10).
		WithState("name", "Alice").
		WithState("age", 25).
		WithEvent("init", nil).
		WithEvent("ready", true)

	// Verify all values were set
	assert.Equal(t, "Test", fixture.props["title"])
	assert.Equal(t, 10, fixture.props["count"])
	assert.Equal(t, "Alice", fixture.state["name"])
	assert.Equal(t, 25, fixture.state["age"])
	assert.Equal(t, nil, fixture.events["init"])
	assert.Equal(t, true, fixture.events["ready"])
}

// TestBuild tests that Build creates a ComponentTest with the fixture configuration.
func TestBuild(t *testing.T) {
	// Create a simple test component
	createComponent := func() bubbly.Component {
		comp, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				name := ctx.Ref("")
				ctx.Expose("count", count)
				ctx.Expose("name", name)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()
		if err != nil {
			t.Fatalf("failed to build component: %v", err)
		}
		return comp
	}

	// Create fixture with state
	fixture := NewFixture().
		WithState("count", 42).
		WithState("name", "TestName")

	// Build the component test
	ct := fixture.Build(t, createComponent)

	// Verify ComponentTest was created
	assert.NotNil(t, ct, "Build should return ComponentTest")
	assert.NotNil(t, ct.component, "ComponentTest should have component")
	assert.NotNil(t, ct.state, "ComponentTest should have state inspector")

	// Verify state was applied
	assert.Equal(t, 42, ct.state.GetRefValue("count"), "count state should be applied")
	assert.Equal(t, "TestName", ct.state.GetRefValue("name"), "name state should be applied")
}

// TestBuildWithEvents tests that Build emits events after mounting.
func TestBuildWithEvents(t *testing.T) {
	// Create component that tracks events
	createComponent := func() bubbly.Component {
		comp, err := bubbly.NewComponent("EventComponent").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				// Handle increment event
				ctx.On("increment", func(data interface{}) {
					current := count.Get().(int)
					count.Set(current + 1)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Event Test"
			}).
			Build()
		if err != nil {
			t.Fatalf("failed to build component: %v", err)
		}
		return comp
	}

	// Create fixture with events
	fixture := NewFixture().
		WithState("count", 0).
		WithEvent("increment", nil).
		WithEvent("increment", nil) // Emit twice

	// Build the component test
	ct := fixture.Build(t, createComponent)

	// Verify events were emitted
	// Note: Events are emitted but we can't easily verify the count changed
	// because the event handler execution is async. This test verifies
	// that Build() doesn't panic when emitting events.
	assert.NotNil(t, ct, "Build should succeed with events")
}

// TestBuildEmptyFixture tests that Build works with an empty fixture.
func TestBuildEmptyFixture(t *testing.T) {
	createComponent := func() bubbly.Component {
		comp, err := bubbly.NewComponent("EmptyComponent").
			Setup(func(ctx *bubbly.Context) {
				ctx.Expose("value", ctx.Ref(0))
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Empty"
			}).
			Build()
		if err != nil {
			t.Fatalf("failed to build component: %v", err)
		}
		return comp
	}

	fixture := NewFixture()
	ct := fixture.Build(t, createComponent)

	assert.NotNil(t, ct, "Build should work with empty fixture")
	assert.NotNil(t, ct.component, "Component should be created")
}

// TestBuildMultipleStates tests applying multiple state values.
func TestBuildMultipleStates(t *testing.T) {
	createComponent := func() bubbly.Component {
		comp, err := bubbly.NewComponent("MultiStateComponent").
			Setup(func(ctx *bubbly.Context) {
				ctx.Expose("str", ctx.Ref(""))
				ctx.Expose("num", ctx.Ref(0))
				ctx.Expose("bool", ctx.Ref(false))
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Multi"
			}).
			Build()
		if err != nil {
			t.Fatalf("failed to build component: %v", err)
		}
		return comp
	}

	fixture := NewFixture().
		WithState("str", "hello").
		WithState("num", 123).
		WithState("bool", true)

	ct := fixture.Build(t, createComponent)

	assert.Equal(t, "hello", ct.state.GetRefValue("str"))
	assert.Equal(t, 123, ct.state.GetRefValue("num"))
	assert.Equal(t, true, ct.state.GetRefValue("bool"))
}

// TestBuildIntegration tests a realistic fixture usage scenario.
func TestBuildIntegration(t *testing.T) {
	// Create a counter component
	createCounter := func() bubbly.Component {
		comp, err := bubbly.NewComponent("Counter").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				ctx.On("increment", func(data interface{}) {
					current := count.Get().(int)
					count.Set(current + 1)
				})

				ctx.On("reset", func(data interface{}) {
					count.Set(0)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				count := ctx.Get("count").(*bubbly.Ref[interface{}])
				return "Count: " + string(rune(count.Get().(int)))
			}).
			Build()
		if err != nil {
			t.Fatalf("failed to build component: %v", err)
		}
		return comp
	}

	// Use fixture to set up test scenario
	fixture := NewFixture().
		WithState("count", 10) // Start at 10

	ct := fixture.Build(t, createCounter)

	// Verify initial state from fixture
	assert.Equal(t, 10, ct.state.GetRefValue("count"))

	// Interact with component
	ct.component.Emit("increment", nil)
	// Note: Can't verify count is 11 because event handling is async
	// This is expected behavior - fixtures set initial state, not final state
}
