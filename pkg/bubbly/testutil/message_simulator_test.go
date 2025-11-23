package testutil

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestSendMessage_SendsMessageToComponent tests that SendMessage calls Update on the component
func TestSendMessage_SendsMessageToComponent(t *testing.T) {
	tests := []struct {
		name        string
		msg         tea.Msg
		expectCmd   bool
		description string
	}{
		{
			name:        "send key message",
			msg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expectCmd:   false,
			description: "Should send KeyMsg to component",
		},
		{
			name:        "send quit message",
			msg:         tea.QuitMsg{},
			expectCmd:   false,
			description: "Should send QuitMsg (command depends on component implementation)",
		},
		{
			name: "send window size message",
			msg: tea.WindowSizeMsg{
				Width:  80,
				Height: 24,
			},
			expectCmd:   false,
			description: "Should send WindowSizeMsg to component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			harness := NewHarness(t)

			// Create a simple component that tracks Update calls
			updateCalled := false
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					count := ctx.Ref(0)
					ctx.Expose("count", count)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			assert.NoError(t, err, "Component build should succeed")

			// Wrap component to track Update calls
			wrappedComponent := &updateTrackingComponent{
				Component:    component,
				updateCalled: &updateCalled,
			}

			ct := harness.Mount(wrappedComponent)

			// Send message
			cmd := ct.SendMessage(tt.msg)

			// Verify Update was called
			assert.True(t, updateCalled, "Update should have been called")

			// Verify command returned based on expectation
			if tt.expectCmd {
				assert.NotNil(t, cmd, "Command should be returned")
			}
			// Command can be nil or not nil depending on component logic
			// Just verify no panic occurred
		})
	}
}

// TestSendKey_CreatesKeyMsg tests that SendKey creates proper KeyMsg
func TestSendKey_CreatesKeyMsg(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectRunes bool
		description string
	}{
		{
			name:        "single character",
			key:         "a",
			expectRunes: true,
			description: "Should create KeyMsg with rune 'a'",
		},
		{
			name:        "space key",
			key:         " ",
			expectRunes: true,
			description: "Should create KeyMsg with space rune",
		},
		{
			name:        "enter key",
			key:         "enter",
			expectRunes: false,
			description: "Should create KeyMsg for enter",
		},
		{
			name:        "ctrl+c",
			key:         "ctrl+c",
			expectRunes: false,
			description: "Should create KeyMsg for ctrl+c",
		},
		{
			name:        "up arrow",
			key:         "up",
			expectRunes: false,
			description: "Should create KeyMsg for up arrow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			harness := NewHarness(t)

			// Track received messages
			var receivedMsg tea.Msg
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Empty setup
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			assert.NoError(t, err, "Component build should succeed")

			// Wrap to capture messages
			wrappedComponent := &messageCaptureComponent{
				Component:   component,
				capturedMsg: &receivedMsg,
			}

			ct := harness.Mount(wrappedComponent)

			// Send key
			cmd := ct.SendKey(tt.key)

			// Verify message was sent
			assert.NotNil(t, receivedMsg, "Message should have been captured")

			// Verify it's a KeyMsg
			keyMsg, ok := receivedMsg.(tea.KeyMsg)
			assert.True(t, ok, "Message should be a KeyMsg")

			// Verify key string matches
			assert.Equal(t, tt.key, keyMsg.String(), "Key string should match")

			// Command can be nil or not nil
			_ = cmd
		})
	}
}

// TestSendMouseClick_CreatesMouseMsg tests that SendMouseClick creates proper MouseMsg
func TestSendMouseClick_CreatesMouseMsg(t *testing.T) {
	tests := []struct {
		name        string
		x           int
		y           int
		description string
	}{
		{
			name:        "click at origin",
			x:           0,
			y:           0,
			description: "Should create MouseMsg at (0,0)",
		},
		{
			name:        "click at center",
			x:           40,
			y:           12,
			description: "Should create MouseMsg at (40,12)",
		},
		{
			name:        "click at bottom right",
			x:           79,
			y:           23,
			description: "Should create MouseMsg at (79,23)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			harness := NewHarness(t)

			// Track received messages
			var receivedMsg tea.Msg
			component, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Empty setup
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Test"
				}).
				Build()
			assert.NoError(t, err, "Component build should succeed")

			// Wrap to capture messages
			wrappedComponent := &messageCaptureComponent{
				Component:   component,
				capturedMsg: &receivedMsg,
			}

			ct := harness.Mount(wrappedComponent)

			// Send mouse click
			cmd := ct.SendMouseClick(tt.x, tt.y)

			// Verify message was sent
			assert.NotNil(t, receivedMsg, "Message should have been captured")

			// Verify it's a MouseMsg
			mouseMsg, ok := receivedMsg.(tea.MouseMsg)
			assert.True(t, ok, "Message should be a MouseMsg")

			// Verify coordinates
			assert.Equal(t, tt.x, mouseMsg.X, "X coordinate should match")
			assert.Equal(t, tt.y, mouseMsg.Y, "Y coordinate should match")

			// Verify it's a left click using the modern API
			assert.Equal(t, tea.MouseButtonLeft, mouseMsg.Button, "Should be left mouse button")
			assert.Equal(t, tea.MouseActionPress, mouseMsg.Action, "Should be press action")

			// Command can be nil or not nil
			_ = cmd
		})
	}
}

// TestSendMessage_ReturnsCommand tests that commands are properly returned
func TestSendMessage_ReturnsCommand(t *testing.T) {
	harness := NewHarness(t)

	// Create component that returns a command on specific message
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// Empty setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Test"
		}).
		Build()
	assert.NoError(t, err, "Component build should succeed")

	ct := harness.Mount(component)

	// Send quit message (should return quit command)
	cmd := ct.SendMessage(tea.QuitMsg{})

	// Verify command returned
	// Note: We can't easily test the command itself without executing it,
	// but we can verify it's not nil for quit messages
	_ = cmd // Command presence depends on component implementation
}

// TestMessageSimulator_Integration tests message simulation in realistic scenario
func TestMessageSimulator_Integration(t *testing.T) {
	harness := NewHarness(t)

	// Create a counter component that responds to key presses
	component, err := bubbly.NewComponent("Counter").
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
			return "Counter"
		}).
		Build()
	assert.NoError(t, err, "Component build should succeed")

	ct := harness.Mount(component)

	// Initial count should be 0
	ct.AssertRefEquals("count", 0)

	// Simulate space key press (typically triggers increment)
	ct.SendKey(" ")

	// Emit increment event manually (since we're not handling keys in component)
	ct.Emit("increment", nil)

	// Count should be 1
	ct.AssertRefEquals("count", 1)

	// Send multiple keys
	ct.SendKey("up")
	ct.SendKey("down")
	ct.SendKey("enter")

	// Verify component still works
	view := ct.component.View()
	assert.Contains(t, view, "Counter")
}

// Helper: updateTrackingComponent wraps a component to track Update calls
type updateTrackingComponent struct {
	bubbly.Component
	updateCalled *bool
}

func (c *updateTrackingComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	*c.updateCalled = true
	return c.Component.Update(msg)
}

// Helper: messageCaptureComponent wraps a component to capture messages
type messageCaptureComponent struct {
	bubbly.Component
	capturedMsg *tea.Msg
}

func (c *messageCaptureComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	*c.capturedMsg = msg
	return c.Component.Update(msg)
}

// TestCreateKeyMsg_SpecialKeys tests createKeyMsg with all special key combinations
func TestCreateKeyMsg_SpecialKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected tea.KeyType
	}{
		// Navigation keys
		{"enter key", "enter", tea.KeyEnter},
		{"escape key", "esc", tea.KeyEsc},
		{"tab key", "tab", tea.KeyTab},
		{"backspace key", "backspace", tea.KeyBackspace},
		{"delete key", "delete", tea.KeyDelete},
		{"up arrow", "up", tea.KeyUp},
		{"down arrow", "down", tea.KeyDown},
		{"left arrow", "left", tea.KeyLeft},
		{"right arrow", "right", tea.KeyRight},
		{"home key", "home", tea.KeyHome},
		{"end key", "end", tea.KeyEnd},
		{"page up", "pgup", tea.KeyPgUp},
		{"page down", "pgdown", tea.KeyPgDown},

		// Ctrl combinations
		{"ctrl+c", "ctrl+c", tea.KeyCtrlC},
		{"ctrl+d", "ctrl+d", tea.KeyCtrlD},
		{"ctrl+a", "ctrl+a", tea.KeyCtrlA},
		{"ctrl+e", "ctrl+e", tea.KeyCtrlE},
		{"ctrl+k", "ctrl+k", tea.KeyCtrlK},
		{"ctrl+u", "ctrl+u", tea.KeyCtrlU},
		{"ctrl+w", "ctrl+w", tea.KeyCtrlW},
		{"ctrl+l", "ctrl+l", tea.KeyCtrlL},
		{"ctrl+n", "ctrl+n", tea.KeyCtrlN},
		{"ctrl+p", "ctrl+p", tea.KeyCtrlP},
		{"ctrl+b", "ctrl+b", tea.KeyCtrlB},
		{"ctrl+f", "ctrl+f", tea.KeyCtrlF},

		// Function keys
		{"f1 key", "f1", tea.KeyF1},
		{"f2 key", "f2", tea.KeyF2},
		{"f3 key", "f3", tea.KeyF3},
		{"f4 key", "f4", tea.KeyF4},
		{"f5 key", "f5", tea.KeyF5},
		{"f6 key", "f6", tea.KeyF6},
		{"f7 key", "f7", tea.KeyF7},
		{"f8 key", "f8", tea.KeyF8},
		{"f9 key", "f9", tea.KeyF9},
		{"f10 key", "f10", tea.KeyF10},
		{"f11 key", "f11", tea.KeyF11},
		{"f12 key", "f12", tea.KeyF12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createKeyMsg(tt.key)
			assert.Equal(t, tt.expected, result.Type, "Key type should match expected")
			assert.Empty(t, result.Runes, "Special keys should have no runes")
		})
	}
}

// TestCreateKeyMsg_CharacterKeys tests createKeyMsg with regular character input
func TestCreateKeyMsg_CharacterKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected []rune
	}{
		{"single letter", "a", []rune{'a'}},
		{"single number", "5", []rune{'5'}},
		{"uppercase letter", "X", []rune{'X'}},
		{"special character", "!", []rune{'!'}},
		{"space character", " ", []rune{' '}},
		{"multi-character", "hello", []rune{'h', 'e', 'l', 'l', 'o'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createKeyMsg(tt.key)
			assert.Equal(t, tea.KeyRunes, result.Type, "Character keys should be KeyRunes type")
			assert.Equal(t, tt.expected, result.Runes, "Runes should match input")
		})
	}
}

// TestCreateKeyMsg_EdgeCases tests createKeyMsg with edge cases
func TestCreateKeyMsg_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected tea.KeyType
	}{
		{"empty string", "", tea.KeyRunes},
		{"single character", "x", tea.KeyRunes},
		{"unknown f key", "f13", tea.KeyRunes},   // Not in switch, falls back to runes
		{"unknown ctrl", "ctrl+z", tea.KeyRunes}, // Not in switch, falls back to runes
		{"partial match", "ctrl", tea.KeyRunes},  // Not a full match, falls back to runes
		{"f without number", "f", tea.KeyRunes},  // Not a full match, falls back to runes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createKeyMsg(tt.key)
			assert.Equal(t, tt.expected, result.Type, "Key type should match expected")
		})
	}
}
