package testutil

import (
	"reflect"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// MessageHandlerTester provides utilities for testing Bubbletea message handling and routing.
//
// It wraps a component and tracks all messages sent to it, identifies message types,
// counts handler invocations, tracks unhandled messages, and captures returned commands.
// This is useful for verifying that components correctly process different message types
// and that message handlers are invoked as expected.
//
// Type Safety:
//   - Thread-safe message tracking with mutex protection
//   - Reflection-based message type identification
//   - Clear assertion methods for handler verification
//
// Example:
//
//	func TestMessageHandling(t *testing.T) {
//		comp, _ := bubbly.NewComponent("TestComponent").
//			WithKeyBinding(" ", "increment", "Increment").
//			Setup(func(ctx *bubbly.Context) {
//				count := ctx.Ref(0)
//				ctx.On("increment", func(_ interface{}) {
//					count.Set(count.Get().(int) + 1)
//				})
//			}).
//			Template(func(ctx bubbly.RenderContext) string {
//				return "Counter"
//			}).
//			Build()
//
//		tester := testutil.NewMessageHandlerTester(comp)
//
//		// Send message
//		cmd := tester.SendMessage(tea.KeyMsg{Type: tea.KeySpace})
//
//		// Verify message was handled
//		tester.AssertMessageHandled(t, "tea.KeyMsg", 1)
//
//		// Verify command returned
//		assert.NotNil(t, cmd)
//	}
type MessageHandlerTester struct {
	// component is the component being tested
	component bubbly.Component

	// messages tracks all messages sent to the component
	messages []tea.Msg

	// handled maps message type names to invocation counts
	handled map[string]int

	// unhandled tracks messages that didn't trigger any handler
	unhandled []tea.Msg

	// mu protects concurrent access to tracking fields
	mu sync.RWMutex
}

// NewMessageHandlerTester creates a new MessageHandlerTester for testing message handling.
//
// Parameters:
//   - comp: The component to test
//
// Returns:
//   - *MessageHandlerTester: A new tester instance
//
// Panics:
//   - If comp is nil
//
// Example:
//
//	comp, _ := bubbly.NewComponent("TestComponent").
//		Setup(func(ctx *bubbly.Context) {}).
//		Template(func(ctx bubbly.RenderContext) string { return "Test" }).
//		Build()
//
//	tester := testutil.NewMessageHandlerTester(comp)
func NewMessageHandlerTester(comp bubbly.Component) *MessageHandlerTester {
	if comp == nil {
		panic("MessageHandlerTester: component cannot be nil")
	}

	return &MessageHandlerTester{
		component: comp,
		messages:  []tea.Msg{},
		handled:   make(map[string]int),
		unhandled: []tea.Msg{},
	}
}

// SendMessage sends a message to the component and tracks handling.
//
// This method calls the component's Update method with the provided message,
// tracks the message in the messages slice, determines if it was handled
// (by checking if a command was returned), increments the handler count
// for the message type, and captures any returned command.
//
// A message is considered "handled" if the component returns a non-nil command
// or if the component's internal state changes (which we can't directly detect,
// so we rely on command return as the primary indicator).
//
// Parameters:
//   - msg: The message to send to the component
//
// Returns:
//   - tea.Cmd: Any command returned by the component's Update method (may be nil)
//
// Thread Safety:
//   - This method is thread-safe and can be called concurrently
//
// Example:
//
//	tester := testutil.NewMessageHandlerTester(comp)
//	cmd := tester.SendMessage(tea.KeyMsg{Type: tea.KeySpace})
//	assert.NotNil(t, cmd)
func (mht *MessageHandlerTester) SendMessage(msg tea.Msg) tea.Cmd {
	mht.mu.Lock()
	defer mht.mu.Unlock()

	// Track the message
	mht.messages = append(mht.messages, msg)

	// Get message type name using reflection
	msgType := reflect.TypeOf(msg).String()

	// Send message to component
	_, cmd := mht.component.Update(msg)

	// Determine if message was handled
	// A message is considered handled if:
	// 1. A command was returned (indicates the component processed it)
	// 2. OR it's a standard Bubbletea message type that components process
	//
	// We consider all standard Bubbletea messages as "handled" because
	// components receive them through Update() even if they don't explicitly
	// process them. This matches the Bubbletea model where all messages
	// flow through Update().
	handled := cmd != nil

	// Standard Bubbletea message types are always considered handled
	// even if no command is returned
	switch msg.(type) {
	case tea.KeyMsg, tea.WindowSizeMsg, tea.MouseMsg, tea.QuitMsg:
		handled = true
	}

	if handled {
		// Increment handler count for this message type
		mht.handled[msgType]++
	} else {
		// Track as unhandled (typically custom message types)
		mht.unhandled = append(mht.unhandled, msg)
	}

	return cmd
}

// AssertMessageHandled asserts that a specific message type was handled the expected number of times.
//
// This method checks the handler invocation count for the specified message type
// and reports an error if it doesn't match the expected count.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - msgType: The message type name (e.g., "tea.KeyMsg", "tea.WindowSizeMsg")
//   - times: The expected number of times the message type was handled
//
// Example:
//
//	tester.SendMessage(tea.KeyMsg{Type: tea.KeySpace})
//	tester.SendMessage(tea.KeyMsg{Type: tea.KeyEnter})
//	tester.AssertMessageHandled(t, "tea.KeyMsg", 2)
func (mht *MessageHandlerTester) AssertMessageHandled(t testingT, msgType string, times int) {
	t.Helper()

	mht.mu.RLock()
	actual := mht.handled[msgType]
	mht.mu.RUnlock()

	if actual != times {
		t.Errorf("expected message type %q to be handled %d times, but was handled %d times",
			msgType, times, actual)
	}
}

// AssertUnhandledMessages asserts that the expected number of messages were unhandled.
//
// This method checks the count of unhandled messages and reports an error
// if it doesn't match the expected count. Unhandled messages are those that
// didn't return a command and aren't known message types that components
// typically process silently.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - count: The expected number of unhandled messages
//
// Example:
//
//	// Component doesn't handle custom messages
//	tester.SendMessage(customMsg{data: "test"})
//	tester.AssertUnhandledMessages(t, 1)
func (mht *MessageHandlerTester) AssertUnhandledMessages(t testingT, count int) {
	t.Helper()

	mht.mu.RLock()
	actual := len(mht.unhandled)
	mht.mu.RUnlock()

	if actual != count {
		t.Errorf("expected %d unhandled messages, but got %d", count, actual)
	}
}

// GetHandledMessages returns all messages that were sent to the component.
//
// This method returns a copy of the messages slice to prevent external
// modification of the internal state.
//
// Returns:
//   - []tea.Msg: A copy of all messages sent to the component
//
// Thread Safety:
//   - This method is thread-safe
//
// Example:
//
//	tester.SendMessage(tea.KeyMsg{Type: tea.KeySpace})
//	tester.SendMessage(tea.WindowSizeMsg{Width: 80, Height: 24})
//
//	messages := tester.GetHandledMessages()
//	assert.Len(t, messages, 2)
func (mht *MessageHandlerTester) GetHandledMessages() []tea.Msg {
	mht.mu.RLock()
	defer mht.mu.RUnlock()

	// Return a copy to prevent external modification
	messagesCopy := make([]tea.Msg, len(mht.messages))
	copy(messagesCopy, mht.messages)

	return messagesCopy
}
