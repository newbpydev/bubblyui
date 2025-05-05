package core

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestCreateStateContext tests the creation and retrieval of a state context
func TestCreateStateContext(t *testing.T) {
	// Clear contexts before test
	ClearAllContexts()

	// Create a context
	context := CreateStateContext("testContext", "default")
	assert.NotNil(t, context, "Context should be created")
	assert.Equal(t, "default", context.state.Get(), "Context should have default value")

	// Retrieve the same context
	context2 := CreateStateContext("testContext", "unused")
	assert.Same(t, context, context2, "Should retrieve the same context instance")
	assert.Equal(t, "default", context2.state.Get(), "Should still have the original default value")
}

// TestProvideContext tests providing a value to a context
func TestProvideContext(t *testing.T) {
	// Clear contexts before test
	ClearAllContexts()

	// Create a context and a mock component
	context := CreateStateContext("testContext", 0)
	component := NewMockStatefulComponent("test-id", "TestComponent")

	// Provide a value to the context
	ProvideContext(component, context, 42)
	assert.Equal(t, 42, context.state.Get(), "Context value should be updated")

	// Test unmounting the component
	component.state.GetHookManager().ExecuteUnmountHooks()
	assert.Equal(t, 0, context.state.Get(), "Context should reset to default on provider unmount")
}

// TestProvideContextState tests providing a state to a context
func TestProvideContextState(t *testing.T) {
	// Clear contexts before test
	ClearAllContexts()

	// Create a context, component, and state
	context := CreateStateContext("stateContext", "default")
	component := NewMockStatefulComponent("test-id", "TestComponent")
	state := NewState("component state")

	// Provide the state to the context
	ProvideContextState(component, context, state)
	assert.Equal(t, "component state", context.state.Get(), "Context should have the state's value")

	// Update the state and verify the context updates
	state.Set("updated value")
	assert.Equal(t, "updated value", context.state.Get(), "Context should update when state changes")

	// Test unmounting the component
	component.state.GetHookManager().ExecuteUnmountHooks()
	assert.Equal(t, "default", context.state.Get(), "Context should reset to default on provider unmount")
}

// TestUseContext tests consuming a context value
func TestUseContext(t *testing.T) {
	// Clear contexts before test
	ClearAllContexts()

	// Create a context, provider component, and consumer component
	context := CreateStateContext("useContext", "default")
	provider := NewMockStatefulComponent("provider", "Provider")
	consumer := NewMockStatefulComponent("consumer", "Consumer")

	// Provide a value
	ProvideContext(provider, context, "provided value")

	// Consume the context
	signal := UseContext(consumer, context)
	assert.Equal(t, "provided value", signal.Value(), "Consumer should receive the provided value")

	// Update the provided value
	ProvideContext(provider, context, "updated value")
	assert.Equal(t, "updated value", signal.Value(), "Consumer should receive the updated value")
}

// TestUseContextWithDefault tests consuming a context with a local default
func TestUseContextWithDefault(t *testing.T) {
	// Clear contexts before test
	ClearAllContexts()

	// Create a context and consumer
	context := CreateStateContext("defaultContext", "global default")
	consumer := NewMockStatefulComponent("consumer", "Consumer")

	// Use the context with a local default
	signal := UseContextWithDefault(consumer, context, "local default")

	// Since no provider exists, the consumer should get the local default
	assert.Equal(t, "local default", signal.Value(), "Should use local default when no provider exists")

	// Add a provider
	provider := NewMockStatefulComponent("provider", "Provider")
	ProvideContext(provider, context, "provided value")

	// Consumer should now get the provided value
	assert.Equal(t, "provided value", signal.Value(), "Should use provided value when available")
}

// TestUseContextState tests two-way binding with a context
func TestUseContextState(t *testing.T) {
	// Clear contexts before test
	ClearAllContexts()

	// Create a context, provider, and consumer
	context := CreateStateContext("stateContext", "default")
	provider := NewMockStatefulComponent("provider", "Provider")
	consumer := NewMockStatefulComponent("consumer", "Consumer")

	// Provide a value
	ProvideContext(provider, context, "provided value")

	// Get a state that's bound to the context
	state := UseContextState(consumer, context)
	assert.Equal(t, "provided value", state.Get(), "Initial state should match context")

	// Update the local state
	state.Set("local update")
	assert.Equal(t, "local update", context.state.Get(), "Context should update when local state changes")

	// Update the context from provider
	ProvideContext(provider, context, "provider update")
	assert.Equal(t, "provider update", state.Get(), "Local state should update when context changes")
}

// NewMockStatefulComponent creates a mock StatefulComponent for testing
func NewMockStatefulComponent(id, name string) *MockStatefulComponent {
	return &MockStatefulComponent{
		state:   NewComponentState(id, name),
		mounted: true,
	}
}

// MockStatefulComponent implements StatefulComponent for testing
type MockStatefulComponent struct {
	state   *ComponentState
	mounted bool
}

func (m *MockStatefulComponent) Initialize() error                   { return nil }
func (m *MockStatefulComponent) Update(msg tea.Msg) (tea.Cmd, error) { return nil, nil }
func (m *MockStatefulComponent) Render() string                      { return "mock" }
func (m *MockStatefulComponent) Dispose() error                      { return nil }
func (m *MockStatefulComponent) ID() string                          { return "mock-id" }
func (m *MockStatefulComponent) AddChild(child Component)            {}
func (m *MockStatefulComponent) RemoveChild(id string) bool          { return false }
func (m *MockStatefulComponent) Children() []Component               { return nil }
func (m *MockStatefulComponent) GetState() *ComponentState           { return m.state }
func (m *MockStatefulComponent) ExecuteEffect() error                { return nil }
func (m *MockStatefulComponent) IsMounted() bool                     { return m.mounted }
func (m *MockStatefulComponent) SetMounted(mounted bool)             { m.mounted = mounted }
