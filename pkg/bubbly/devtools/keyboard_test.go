package devtools

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestNewKeyboardHandler tests creating a new keyboard handler.
func TestNewKeyboardHandler(t *testing.T) {
	kh := NewKeyboardHandler()
	assert.NotNil(t, kh)
	assert.Equal(t, FocusApp, kh.GetFocus())
}

// TestKeyboardHandler_Register tests registering keyboard shortcuts.
func TestKeyboardHandler_Register(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		handler KeyHandler
		wantNil bool
	}{
		{
			name: "register valid handler",
			key:  "f12",
			handler: func(msg tea.KeyMsg) tea.Cmd {
				return nil
			},
			wantNil: false,
		},
		{
			name:    "register nil handler",
			key:     "ctrl+c",
			handler: nil,
			wantNil: true,
		},
		{
			name: "register empty key",
			key:  "",
			handler: func(msg tea.KeyMsg) tea.Cmd {
				return nil
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kh := NewKeyboardHandler()
			kh.Register(tt.key, tt.handler)

			// Try to trigger the handler
			keyMsg := tea.KeyMsg{}
			keyMsg.Type = tea.KeyRunes
			cmd := kh.Handle(keyMsg)

			if tt.wantNil {
				// Should not have registered
				assert.Nil(t, cmd)
			}
		})
	}
}

// TestKeyboardHandler_Handle tests handling keyboard messages.
func TestKeyboardHandler_Handle(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		msgString  string
		shouldCall bool
	}{
		{
			name:       "f12 toggle",
			key:        "f12",
			msgString:  "f12",
			shouldCall: true,
		},
		{
			name:       "tab switch",
			key:        "tab",
			msgString:  "tab",
			shouldCall: true,
		},
		{
			name:       "ctrl+f search",
			key:        "ctrl+f",
			msgString:  "ctrl+f",
			shouldCall: true,
		},
		{
			name:       "unregistered key",
			key:        "f12",
			msgString:  "f11",
			shouldCall: false,
		},
		{
			name:       "question mark help",
			key:        "?",
			msgString:  "?",
			shouldCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kh := NewKeyboardHandler()
			called := false

			// Register handler
			kh.Register(tt.key, func(msg tea.KeyMsg) tea.Cmd {
				called = true
				return nil
			})

			// Create key message
			keyMsg := createKeyMsg(tt.msgString)

			// Handle message
			kh.Handle(keyMsg)

			assert.Equal(t, tt.shouldCall, called)
		})
	}
}

// TestKeyboardHandler_Focus tests focus management.
func TestKeyboardHandler_Focus(t *testing.T) {
	tests := []struct {
		name  string
		focus FocusTarget
	}{
		{
			name:  "focus app",
			focus: FocusApp,
		},
		{
			name:  "focus tools",
			focus: FocusTools,
		},
		{
			name:  "focus inspector",
			focus: FocusInspector,
		},
		{
			name:  "focus state",
			focus: FocusState,
		},
		{
			name:  "focus events",
			focus: FocusEvents,
		},
		{
			name:  "focus performance",
			focus: FocusPerformance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kh := NewKeyboardHandler()
			kh.SetFocus(tt.focus)
			assert.Equal(t, tt.focus, kh.GetFocus())
		})
	}
}

// TestKeyboardHandler_HandleWithFocus tests handling with focus-specific shortcuts.
func TestKeyboardHandler_HandleWithFocus(t *testing.T) {
	kh := NewKeyboardHandler()

	appCalled := false
	toolsCalled := false

	// Register app-specific handler
	kh.RegisterWithFocus("a", FocusApp, func(msg tea.KeyMsg) tea.Cmd {
		appCalled = true
		return nil
	})

	// Register tools-specific handler
	kh.RegisterWithFocus("a", FocusTools, func(msg tea.KeyMsg) tea.Cmd {
		toolsCalled = true
		return nil
	})

	// Test with app focus
	kh.SetFocus(FocusApp)
	keyMsg := createKeyMsg("a")
	kh.Handle(keyMsg)
	assert.True(t, appCalled)
	assert.False(t, toolsCalled)

	// Reset
	appCalled = false
	toolsCalled = false

	// Test with tools focus
	kh.SetFocus(FocusTools)
	kh.Handle(keyMsg)
	assert.False(t, appCalled)
	assert.True(t, toolsCalled)
}

// TestKeyboardHandler_GlobalShortcuts tests global shortcuts that work regardless of focus.
func TestKeyboardHandler_GlobalShortcuts(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		focus FocusTarget
	}{
		{
			name:  "f12 with app focus",
			key:   "f12",
			focus: FocusApp,
		},
		{
			name:  "f12 with tools focus",
			key:   "f12",
			focus: FocusTools,
		},
		{
			name:  "ctrl+c with inspector focus",
			key:   "ctrl+c",
			focus: FocusInspector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kh := NewKeyboardHandler()
			called := false

			// Register global handler
			kh.RegisterGlobal(tt.key, func(msg tea.KeyMsg) tea.Cmd {
				called = true
				return nil
			})

			// Set focus
			kh.SetFocus(tt.focus)

			// Handle key
			keyMsg := createKeyMsg(tt.key)
			kh.Handle(keyMsg)

			assert.True(t, called, "Global shortcut should work with any focus")
		})
	}
}

// TestKeyboardHandler_Concurrent tests thread-safe concurrent access.
func TestKeyboardHandler_Concurrent(t *testing.T) {
	kh := NewKeyboardHandler()

	// Register handlers
	for i := 0; i < 10; i++ {
		key := string(rune('a' + i))
		kh.Register(key, func(msg tea.KeyMsg) tea.Cmd {
			return nil
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Concurrent operations
			key := string(rune('a' + (idx % 10)))
			keyMsg := createKeyMsg(key)
			kh.Handle(keyMsg)

			// Concurrent focus changes
			focus := FocusTarget(idx % 6)
			kh.SetFocus(focus)
			_ = kh.GetFocus()
		}(i)
	}

	wg.Wait()
}

// TestKeyboardHandler_UnregisterShortcut tests removing shortcuts.
func TestKeyboardHandler_UnregisterShortcut(t *testing.T) {
	kh := NewKeyboardHandler()
	called := false

	// Register handler
	kh.Register("f12", func(msg tea.KeyMsg) tea.Cmd {
		called = true
		return nil
	})

	// Verify it works
	keyMsg := createKeyMsg("f12")
	kh.Handle(keyMsg)
	assert.True(t, called)

	// Unregister
	called = false
	kh.Unregister("f12")

	// Verify it no longer works
	kh.Handle(keyMsg)
	assert.False(t, called)
}

// TestKeyboardHandler_CommandReturned tests that commands are properly returned.
func TestKeyboardHandler_CommandReturned(t *testing.T) {
	kh := NewKeyboardHandler()

	expectedCmd := func() tea.Msg {
		return tea.QuitMsg{}
	}

	// Register handler that returns a command
	kh.Register("q", func(msg tea.KeyMsg) tea.Cmd {
		return expectedCmd
	})

	keyMsg := createKeyMsg("q")
	cmd := kh.Handle(keyMsg)

	assert.NotNil(t, cmd)
	// Execute command and verify it returns QuitMsg
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok)
}

// Helper function to create KeyMsg from string.
func createKeyMsg(s string) tea.KeyMsg {
	keyMsg := tea.KeyMsg{}

	switch s {
	case "f12":
		keyMsg.Type = tea.KeyF12
	case "tab":
		keyMsg.Type = tea.KeyTab
	case "ctrl+f":
		keyMsg.Type = tea.KeyCtrlF
	case "ctrl+c":
		keyMsg.Type = tea.KeyCtrlC
	case "?":
		keyMsg.Type = tea.KeyRunes
		keyMsg.Runes = []rune{'?'}
	default:
		keyMsg.Type = tea.KeyRunes
		if len(s) > 0 {
			keyMsg.Runes = []rune{rune(s[0])}
		}
	}

	return keyMsg
}
