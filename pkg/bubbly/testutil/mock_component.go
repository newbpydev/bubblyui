package testutil

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// MockComponent is a mock implementation of the bubbly.Component interface for testing.
// It tracks all method calls and provides assertion helpers to verify component behavior.
//
// MockComponent implements the full Component interface including:
//   - tea.Model methods (Init, Update, View)
//   - Component identification (Name, ID, Props)
//   - Event system (Emit, On)
//   - Key bindings (KeyBindings, HelpText)
//   - Lifecycle (IsInitialized)
//
// Call tracking:
//   - initCalled: Whether Init() was called
//   - updateCalls: Number of times Update() was called
//   - viewCalls: Number of times View() was called
//   - unmountCalled: Whether Unmount() was called
//   - emitCalls: Map of event names to number of times emitted
//   - onCalls: Map of event names to number of handlers registered
//
// Configurable behavior:
//   - viewOutput: String returned by View() (default: "Mock<name>")
//   - props: Props returned by Props() method
//   - keyBindings: Key bindings returned by KeyBindings() method
//   - helpText: Help text returned by HelpText() method
//
// Example:
//
//	mock := NewMockComponent("TestComponent")
//	mock.SetViewOutput("Custom output")
//	mock.SetProps(MyProps{Value: 42})
//
//	// Use in tests
//	mock.Init()
//	mock.Update(tea.KeyMsg{})
//	output := mock.View()
//
//	// Assert behavior
//	mock.AssertInitCalled(t)
//	mock.AssertUpdateCalled(t, 1)
//	mock.AssertViewCalled(t, 1)
type MockComponent struct {
	mu sync.RWMutex

	// Identification
	name string
	id   string

	// Configuration
	props       interface{}
	viewOutput  string
	keyBindings map[string][]bubbly.KeyBinding
	helpText    string

	// Call tracking
	initCalled    bool
	updateCalls   int
	viewCalls     int
	unmountCalled bool
	emitCalls     map[string]int
	onCalls       map[string]int

	// Event handlers (for testing event system)
	handlers map[string][]bubbly.EventHandler
}

// NewMockComponent creates a new mock component with the given name.
// The component is initialized with default values:
//   - ID: "mock-<name>"
//   - ViewOutput: "Mock<<name>>"
//   - All call counters: 0
//   - All maps: empty
//
// Example:
//
//	mock := NewMockComponent("Button")
//	fmt.Println(mock.Name())       // "Button"
//	fmt.Println(mock.ID())         // "mock-Button"
//	fmt.Println(mock.View())       // "Mock<Button>"
func NewMockComponent(name string) *MockComponent {
	return &MockComponent{
		name:        name,
		id:          fmt.Sprintf("mock-%s", name),
		viewOutput:  fmt.Sprintf("Mock<%s>", name),
		keyBindings: make(map[string][]bubbly.KeyBinding),
		emitCalls:   make(map[string]int),
		onCalls:     make(map[string]int),
		handlers:    make(map[string][]bubbly.EventHandler),
	}
}

// SetViewOutput sets the string returned by View().
// This allows tests to configure custom output for assertions.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.SetViewOutput("[ Click Me ]")
//	assert.Equal(t, "[ Click Me ]", mock.View())
func (mc *MockComponent) SetViewOutput(output string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.viewOutput = output
}

// SetProps sets the props returned by Props().
// This allows tests to configure component props.
//
// Example:
//
//	type ButtonProps struct {
//	    Label string
//	}
//	mock := NewMockComponent("Button")
//	mock.SetProps(ButtonProps{Label: "Click"})
//	props := mock.Props().(ButtonProps)
//	assert.Equal(t, "Click", props.Label)
func (mc *MockComponent) SetProps(props interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.props = props
}

// SetKeyBindings sets the key bindings returned by KeyBindings().
// This allows tests to configure component key bindings.
//
// Example:
//
//	mock := NewMockComponent("Counter")
//	bindings := map[string][]bubbly.KeyBinding{
//	    " ": {{Event: "increment", Description: "Increment"}},
//	}
//	mock.SetKeyBindings(bindings)
func (mc *MockComponent) SetKeyBindings(bindings map[string][]bubbly.KeyBinding) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.keyBindings = bindings
}

// SetHelpText sets the help text returned by HelpText().
// This allows tests to configure custom help text.
//
// Example:
//
//	mock := NewMockComponent("Counter")
//	mock.SetHelpText("space: increment • r: reset")
//	assert.Equal(t, "space: increment • r: reset", mock.HelpText())
func (mc *MockComponent) SetHelpText(text string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.helpText = text
}

// Reset clears all call tracking counters.
// This is useful when testing multiple scenarios with the same mock.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.Init()
//	mock.AssertInitCalled(t)
//
//	mock.Reset()
//	mock.AssertInitCalled(t) // Would fail - init not called after reset
func (mc *MockComponent) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.initCalled = false
	mc.updateCalls = 0
	mc.viewCalls = 0
	mc.unmountCalled = false
	mc.emitCalls = make(map[string]int)
	mc.onCalls = make(map[string]int)
	mc.handlers = make(map[string][]bubbly.EventHandler)
}

// Component interface implementation

// Name returns the component's name.
func (mc *MockComponent) Name() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.name
}

// ID returns the component's unique ID.
func (mc *MockComponent) ID() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.id
}

// Props returns the component's props.
func (mc *MockComponent) Props() interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.props
}

// Emit emits an event and tracks the call.
func (mc *MockComponent) Emit(event string, data interface{}) {
	mc.mu.Lock()
	mc.emitCalls[event]++
	handlers := mc.handlers[event]
	mc.mu.Unlock()

	// Call registered handlers (outside lock to prevent deadlocks)
	for _, handler := range handlers {
		handler(data)
	}
}

// On registers an event handler and tracks the call.
func (mc *MockComponent) On(event string, handler bubbly.EventHandler) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.onCalls[event]++
	mc.handlers[event] = append(mc.handlers[event], handler)
}

// KeyBindings returns the component's key bindings.
func (mc *MockComponent) KeyBindings() map[string][]bubbly.KeyBinding {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.keyBindings
}

// HelpText returns the component's help text.
func (mc *MockComponent) HelpText() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.helpText
}

// IsInitialized returns whether Init() was called.
func (mc *MockComponent) IsInitialized() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.initCalled
}

// tea.Model interface implementation

// Init marks the component as initialized and returns nil.
func (mc *MockComponent) Init() tea.Cmd {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.initCalled = true
	return nil
}

// Update increments the update counter and returns the model unchanged.
func (mc *MockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.updateCalls++
	return mc, nil
}

// View increments the view counter and returns the configured output.
func (mc *MockComponent) View() string {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.viewCalls++
	return mc.viewOutput
}

// Assertion helpers

// AssertInitCalled asserts that Init() was called.
// Fails the test if Init() was not called.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.Init()
//	mock.AssertInitCalled(t) // Passes
func (mc *MockComponent) AssertInitCalled(t testingT) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if !mc.initCalled {
		t.Errorf("Init() was not called")
	}
}

// AssertInitNotCalled asserts that Init() was not called.
// Fails the test if Init() was called.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.AssertInitNotCalled(t) // Passes
func (mc *MockComponent) AssertInitNotCalled(t testingT) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.initCalled {
		t.Errorf("Init() was called but should not have been")
	}
}

// AssertUpdateCalled asserts that Update() was called exactly the specified number of times.
// Fails the test if the call count doesn't match.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.Update(tea.KeyMsg{})
//	mock.Update(tea.KeyMsg{})
//	mock.AssertUpdateCalled(t, 2) // Passes
func (mc *MockComponent) AssertUpdateCalled(t testingT, times int) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.updateCalls != times {
		t.Errorf("Update() called %d times, expected %d", mc.updateCalls, times)
	}
}

// AssertViewCalled asserts that View() was called exactly the specified number of times.
// Fails the test if the call count doesn't match.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	_ = mock.View()
//	_ = mock.View()
//	mock.AssertViewCalled(t, 2) // Passes
func (mc *MockComponent) AssertViewCalled(t testingT, times int) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.viewCalls != times {
		t.Errorf("View() called %d times, expected %d", mc.viewCalls, times)
	}
}

// AssertEmitCalled asserts that Emit() was called for the specified event
// exactly the specified number of times.
// Fails the test if the call count doesn't match.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.Emit("click", nil)
//	mock.Emit("click", nil)
//	mock.AssertEmitCalled(t, "click", 2) // Passes
func (mc *MockComponent) AssertEmitCalled(t testingT, event string, times int) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	actual := mc.emitCalls[event]
	if actual != times {
		t.Errorf("Emit(%q) called %d times, expected %d", event, actual, times)
	}
}

// AssertOnCalled asserts that On() was called for the specified event
// exactly the specified number of times.
// Fails the test if the call count doesn't match.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.On("click", func(data interface{}) {})
//	mock.AssertOnCalled(t, "click", 1) // Passes
func (mc *MockComponent) AssertOnCalled(t testingT, event string, times int) {
	t.Helper()
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	actual := mc.onCalls[event]
	if actual != times {
		t.Errorf("On(%q) called %d times, expected %d", event, actual, times)
	}
}

// GetUpdateCallCount returns the number of times Update() was called.
// This is useful for custom assertions or logging.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.Update(tea.KeyMsg{})
//	count := mock.GetUpdateCallCount()
//	assert.Equal(t, 1, count)
func (mc *MockComponent) GetUpdateCallCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.updateCalls
}

// GetViewCallCount returns the number of times View() was called.
// This is useful for custom assertions or logging.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	_ = mock.View()
//	count := mock.GetViewCallCount()
//	assert.Equal(t, 1, count)
func (mc *MockComponent) GetViewCallCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.viewCalls
}

// GetEmitCallCount returns the number of times Emit() was called for the specified event.
// This is useful for custom assertions or logging.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.Emit("click", nil)
//	count := mock.GetEmitCallCount("click")
//	assert.Equal(t, 1, count)
func (mc *MockComponent) GetEmitCallCount(event string) int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.emitCalls[event]
}

// GetOnCallCount returns the number of times On() was called for the specified event.
// This is useful for custom assertions or logging.
//
// Example:
//
//	mock := NewMockComponent("Button")
//	mock.On("click", func(data interface{}) {})
//	count := mock.GetOnCallCount("click")
//	assert.Equal(t, 1, count)
func (mc *MockComponent) GetOnCallCount(event string) int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.onCalls[event]
}
