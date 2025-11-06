package bubbly

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// customTestMsg is a custom message type for testing message handler
type customTestMsg struct {
	value string
}

// TestMessageHandler_CalledWithComponentAndMessage verifies that the message handler
// is called with the correct component reference and message.
func TestMessageHandler_CalledWithComponentAndMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     tea.Msg
		wantCalled  bool
		wantMsgType string
	}{
		{
			name:        "handler receives KeyMsg",
			message:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantCalled:  true,
			wantMsgType: "tea.KeyMsg",
		},
		{
			name:        "handler receives custom message",
			message:     customTestMsg{value: "test"},
			wantCalled:  true,
			wantMsgType: "bubbly.customTestMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			var receivedComp Component
			var receivedMsg tea.Msg

			component, err := NewComponent("TestComponent").
				Template(func(ctx RenderContext) string {
					return "Test"
				}).
				WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
					handlerCalled = true
					receivedComp = comp
					receivedMsg = msg
					return nil
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)

			// Call Update with message
			_, _ = component.Update(tt.message)

			// Verify handler was called
			assert.Equal(t, tt.wantCalled, handlerCalled, "handler should be called")
			if tt.wantCalled {
				assert.NotNil(t, receivedComp, "handler should receive component")
				assert.Equal(t, component, receivedComp, "handler should receive correct component")
				assert.NotNil(t, receivedMsg, "handler should receive message")
			}
		})
	}
}

// TestMessageHandler_ReturnNil verifies that handler can return nil (no command).
func TestMessageHandler_ReturnNil(t *testing.T) {
	component, err := NewComponent("TestComponent").
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
			// Return nil - no command
			return nil
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// Call Update
	_, cmd := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Cmd should be nil (no commands from handler or other sources)
	assert.Nil(t, cmd, "handler returning nil should result in no command")
}

// TestMessageHandler_ReturnCommand verifies that handler can return a command.
func TestMessageHandler_ReturnCommand(t *testing.T) {
	commandExecuted := false

	component, err := NewComponent("TestComponent").
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
			// Return a command
			return func() tea.Msg {
				commandExecuted = true
				return customTestMsg{value: "from handler"}
			}
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// Call Update
	_, cmd := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Command should be returned
	assert.NotNil(t, cmd, "handler returning command should result in command")

	// Execute the command
	msg := cmd()
	assert.True(t, commandExecuted, "command from handler should execute")
	assert.IsType(t, customTestMsg{}, msg, "command should return correct message type")
}

// TestMessageHandler_CommandsBatchedWithOthers verifies that handler commands
// are batched with commands from other sources (like auto-commands).
func TestMessageHandler_CommandsBatchedWithOthers(t *testing.T) {
	handlerCmdExecuted := false

	component, err := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			// This would generate auto-commands when state changes
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
			// Handler returns a command
			return func() tea.Msg {
				handlerCmdExecuted = true
				return customTestMsg{value: "handler cmd"}
			}
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// Initialize component
	_ = component.Init()

	// Call Update - handler should return command
	_, cmd := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Command should be returned (batched if multiple sources)
	assert.NotNil(t, cmd, "commands should be batched")

	// Execute the command (in real Bubbletea, this would be done by runtime)
	if cmd != nil {
		msg := cmd()
		// At least handler command should execute
		if msg != nil {
			assert.True(t, handlerCmdExecuted, "handler command should be in batch")
		}
	}
}

// TestMessageHandler_CanEmitEvents verifies that handler can emit events to component.
func TestMessageHandler_CanEmitEvents(t *testing.T) {
	eventReceived := false
	var receivedData interface{}

	component, err := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			ctx.On("testEvent", func(data interface{}) {
				eventReceived = true
				receivedData = data
			})
		}).
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
			// Handler emits event
			if _, ok := msg.(customTestMsg); ok {
				comp.Emit("testEvent", "data from handler")
			}
			return nil
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// Initialize component
	_ = component.Init()

	// Call Update with custom message
	_, _ = component.Update(customTestMsg{value: "test"})

	// Verify event was emitted and received
	assert.True(t, eventReceived, "handler should emit event")
	assert.Equal(t, "data from handler", receivedData, "event should have correct data")
}

// TestMessageHandler_CoexistsWithKeyBindings verifies that message handler
// works alongside key bindings.
func TestMessageHandler_CoexistsWithKeyBindings(t *testing.T) {
	handlerCalled := false
	keyBindingEventReceived := false

	component, err := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			ctx.On("keyEvent", func(_ interface{}) {
				keyBindingEventReceived = true
			})
		}).
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		WithKeyBinding(" ", "keyEvent", "Test key binding").
		WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
			handlerCalled = true
			return nil
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// Initialize component
	_ = component.Init()

	// Call Update with space key
	_, _ = component.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Both handler and key binding should execute
	assert.True(t, handlerCalled, "message handler should be called")
	assert.True(t, keyBindingEventReceived, "key binding should also work")
}

// TestMessageHandler_CalledBeforeKeyBindings verifies that the message handler
// is called BEFORE key binding processing (as per spec).
func TestMessageHandler_CalledBeforeKeyBindings(t *testing.T) {
	var executionOrder []string

	component, err := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			ctx.On("keyEvent", func(_ interface{}) {
				executionOrder = append(executionOrder, "key_binding")
			})
		}).
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		WithKeyBinding(" ", "keyEvent", "Test").
		WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
			executionOrder = append(executionOrder, "handler")
			return nil
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// Initialize component
	_ = component.Init()

	// Call Update
	_, _ = component.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Verify execution order
	assert.Len(t, executionOrder, 2, "both handler and key binding should execute")
	assert.Equal(t, "handler", executionOrder[0], "handler should execute FIRST")
	assert.Equal(t, "key_binding", executionOrder[1], "key binding should execute AFTER handler")
}

// TestMessageHandler_CustomMessageTypes verifies that handler can process
// custom message types that key bindings cannot handle.
func TestMessageHandler_CustomMessageTypes(t *testing.T) {
	tests := []struct {
		name          string
		message       tea.Msg
		wantEventName string
		wantEventData string
	}{
		{
			name:          "custom message type",
			message:       customTestMsg{value: "custom data"},
			wantEventName: "customReceived",
			wantEventData: "custom data",
		},
		{
			name:          "window size message",
			message:       tea.WindowSizeMsg{Width: 80, Height: 24},
			wantEventName: "resize",
			wantEventData: "80x24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventReceived := false
			var receivedEventName string
			var receivedEventData interface{}

			component, err := NewComponent("TestComponent").
				Setup(func(ctx *Context) {
					ctx.On("customReceived", func(data interface{}) {
						eventReceived = true
						receivedEventName = "customReceived"
						receivedEventData = data
					})
					ctx.On("resize", func(data interface{}) {
						eventReceived = true
						receivedEventName = "resize"
						receivedEventData = data
					})
				}).
				Template(func(ctx RenderContext) string {
					return "Test"
				}).
				WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
					switch msg := msg.(type) {
					case customTestMsg:
						comp.Emit("customReceived", msg.value)
					case tea.WindowSizeMsg:
						comp.Emit("resize", "80x24")
					}
					return nil
				}).
				Build()

			assert.NoError(t, err)
			assert.NotNil(t, component)

			// Initialize component
			_ = component.Init()

			// Call Update with custom message
			_, _ = component.Update(tt.message)

			// Verify custom message was handled
			assert.True(t, eventReceived, "custom message should be handled")
			assert.Equal(t, tt.wantEventName, receivedEventName, "correct event should be emitted")
			assert.Equal(t, tt.wantEventData, receivedEventData, "correct data should be passed")
		})
	}
}

// TestMessageHandler_NotSetDoesNotPanic verifies that components without
// a message handler don't panic (handler is optional).
func TestMessageHandler_NotSetDoesNotPanic(t *testing.T) {
	component, err := NewComponent("TestComponent").
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, component)

	// This should not panic even though no handler is set
	assert.NotPanics(t, func() {
		_, _ = component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	}, "Update should not panic when handler is not set")
}
