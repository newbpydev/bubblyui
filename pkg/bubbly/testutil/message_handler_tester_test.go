package testutil

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestNewMessageHandlerTester_CreatesInstance tests that constructor creates valid instance
func TestNewMessageHandlerTester_CreatesInstance(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	assert.NotNil(t, tester)
	assert.NotNil(t, tester.component)
	assert.Empty(t, tester.messages)
	assert.Empty(t, tester.handled)
	assert.Empty(t, tester.unhandled)
}

// TestNewMessageHandlerTester_PanicsOnNilComponent tests that constructor panics with nil component
func TestNewMessageHandlerTester_PanicsOnNilComponent(t *testing.T) {
	assert.Panics(t, func() {
		NewMessageHandlerTester(nil)
	}, "Should panic when component is nil")
}

// TestSendMessage_TracksMessages tests that SendMessage tracks all sent messages
func TestSendMessage_TracksMessages(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send multiple messages
	tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tester.SendMessage(tea.WindowSizeMsg{Width: 80, Height: 24})
	tester.SendMessage(tea.QuitMsg{})

	// Verify all messages tracked
	messages := tester.GetHandledMessages()
	assert.Len(t, messages, 3)
}

// TestSendMessage_IdentifiesMessageTypes tests that message types are correctly identified
func TestSendMessage_IdentifiesMessageTypes(t *testing.T) {
	tests := []struct {
		name         string
		msg          tea.Msg
		expectedType string
	}{
		{
			name:         "KeyMsg",
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expectedType: "tea.KeyMsg",
		},
		{
			name:         "WindowSizeMsg",
			msg:          tea.WindowSizeMsg{Width: 80, Height: 24},
			expectedType: "tea.WindowSizeMsg",
		},
		{
			name:         "QuitMsg",
			msg:          tea.QuitMsg{},
			expectedType: "tea.QuitMsg",
		},
		{
			name:         "MouseMsg",
			msg:          tea.MouseMsg{X: 10, Y: 5, Type: tea.MouseLeft},
			expectedType: "tea.MouseMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Empty setup
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			assert.NoError(t, err)

			tester := NewMessageHandlerTester(component)
			tester.SendMessage(tt.msg)

			// Verify message type was identified
			tester.AssertMessageHandled(t, tt.expectedType, 1)
		})
	}
}

// TestSendMessage_CapturesCommands tests that returned commands are captured
func TestSendMessage_CapturesCommands(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding("q", "quit", "Quit").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send quit key (should return tea.Quit command)
	cmd := tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Verify command was captured
	assert.NotNil(t, cmd)
}

// TestAssertMessageHandled_CountsCorrectly tests handler invocation counting
func TestAssertMessageHandled_CountsCorrectly(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send same message type multiple times
	tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	// Verify count
	tester.AssertMessageHandled(t, "tea.KeyMsg", 3)
}

// TestAssertMessageHandled_FailsOnMismatch tests assertion failure detection
func TestAssertMessageHandled_FailsOnMismatch(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send one message
	tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Create mock testing.T to capture error
	mockT := &mockTestingT{}

	// Assert wrong count (should fail)
	tester.AssertMessageHandled(mockT, "tea.KeyMsg", 5)

	// Verify error was reported
	assert.True(t, mockT.failed, "Assertion should have failed")
}

// TestAssertUnhandledMessages_TracksUnhandled tests unhandled message tracking
func TestAssertUnhandledMessages_TracksUnhandled(t *testing.T) {
	// Define custom message type that won't be handled
	type customUnhandledMsg struct {
		data string
	}

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Component doesn't handle custom messages
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send custom messages that won't be handled
	// Standard Bubbletea messages (KeyMsg, WindowSizeMsg, etc.) are always considered handled
	tester.SendMessage(customUnhandledMsg{data: "test1"})
	tester.SendMessage(customUnhandledMsg{data: "test2"})

	// Verify unhandled count
	tester.AssertUnhandledMessages(t, 2)
}

// TestGetHandledMessages_ReturnsAllMessages tests message retrieval
func TestGetHandledMessages_ReturnsAllMessages(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send various messages
	msg1 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	msg2 := tea.WindowSizeMsg{Width: 80, Height: 24}
	msg3 := tea.QuitMsg{}

	tester.SendMessage(msg1)
	tester.SendMessage(msg2)
	tester.SendMessage(msg3)

	// Get all messages
	messages := tester.GetHandledMessages()

	// Verify all messages returned
	assert.Len(t, messages, 3)
	assert.Equal(t, msg1, messages[0])
	assert.Equal(t, msg2, messages[1])
	assert.Equal(t, msg3, messages[2])
}

// TestMessageHandlerTester_ThreadSafety tests concurrent message sending
func TestMessageHandlerTester_ThreadSafety(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send messages concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			tester.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all messages tracked
	messages := tester.GetHandledMessages()
	assert.Len(t, messages, 10)
}

// TestMessageHandlerTester_CustomMessages tests custom message types
func TestMessageHandlerTester_CustomMessages(t *testing.T) {
	// Define custom message type
	type customMsg struct {
		data string
	}

	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send custom message
	tester.SendMessage(customMsg{data: "test"})

	// Verify custom message was sent (tracked in messages)
	messages := tester.GetHandledMessages()
	assert.Len(t, messages, 1)

	// Custom messages without handlers are tracked as unhandled
	tester.AssertUnhandledMessages(t, 1)
}

// TestMessageHandlerTester_MessageBatching tests sending multiple messages in sequence
func TestMessageHandlerTester_MessageBatching(t *testing.T) {
	component, err := bubbly.NewComponent("TestComponent").
		WithKeyBinding(" ", "increment", "Increment").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Counter"
		}).
		Build()
	assert.NoError(t, err)

	tester := NewMessageHandlerTester(component)

	// Send batch of messages
	for i := 0; i < 5; i++ {
		tester.SendMessage(tea.KeyMsg{Type: tea.KeySpace})
	}

	// Verify all messages handled
	tester.AssertMessageHandled(t, "tea.KeyMsg", 5)
}

// Note: mockTestingT is defined in assertions_state_test.go and reused here
