package testutil

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SendMessage sends a Bubbletea message to the component.
// It calls the component's Update method with the message and returns any command.
//
// This is the primary method for testing how components respond to Bubbletea messages.
// The message is sent synchronously and Update() is called immediately.
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	cmd := ct.SendMessage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
//	ct.AssertRefEquals("input", "a")
func (ct *ComponentTest) SendMessage(msg tea.Msg) tea.Cmd {
	ct.harness.t.Helper()

	// Call component's Update method with the message
	_, cmd := ct.component.Update(msg)

	// Return the command for verification
	return cmd
}

// SendKey simulates a keyboard key press.
// It creates a KeyMsg from the key string and sends it to the component.
//
// The key parameter can be:
//   - Single character: "a", "b", "1", " " (space)
//   - Special keys: "enter", "esc", "tab", "backspace"
//   - Arrow keys: "up", "down", "left", "right"
//   - Control combinations: "ctrl+c", "ctrl+d", "ctrl+a"
//   - Function keys: "f1", "f2", etc.
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.SendKey(" ")  // Space key
//	ct.SendKey("enter")  // Enter key
//	ct.SendKey("ctrl+c")  // Ctrl+C
func (ct *ComponentTest) SendKey(key string) tea.Cmd {
	ct.harness.t.Helper()

	// Create KeyMsg from key string
	keyMsg := createKeyMsg(key)

	// Send the message
	return ct.SendMessage(keyMsg)
}

// SendMouseClick simulates a mouse click at the specified coordinates.
// It creates a MouseMsg with left button press action.
//
// Coordinates are 0-indexed:
//   - x: horizontal position (0 = leftmost)
//   - y: vertical position (0 = topmost)
//
// Example:
//
//	ct := harness.Mount(createComponent())
//	ct.SendMouseClick(10, 5)  // Click at column 10, row 5
//	ct.AssertEventFired("clicked")
func (ct *ComponentTest) SendMouseClick(x, y int) tea.Cmd {
	ct.harness.t.Helper()

	// Create MouseMsg for left click using the modern API
	mouseMsg := tea.MouseMsg{
		X:      x,
		Y:      y,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}

	// Send the message
	return ct.SendMessage(mouseMsg)
}

// createKeyMsg creates a KeyMsg from a key string.
// This handles the conversion from string representation to Bubbletea KeyMsg.
// keyTypeMap maps key strings to their corresponding tea.KeyType.
// This reduces cyclomatic complexity by replacing a large switch statement with a map lookup.
var keyTypeMap = map[string]tea.KeyType{
	// Special keys
	"enter":     tea.KeyEnter,
	"esc":       tea.KeyEsc,
	"tab":       tea.KeyTab,
	"backspace": tea.KeyBackspace,
	"delete":    tea.KeyDelete,
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	// Control keys
	"ctrl+c": tea.KeyCtrlC,
	"ctrl+d": tea.KeyCtrlD,
	"ctrl+a": tea.KeyCtrlA,
	"ctrl+e": tea.KeyCtrlE,
	"ctrl+k": tea.KeyCtrlK,
	"ctrl+u": tea.KeyCtrlU,
	"ctrl+w": tea.KeyCtrlW,
	"ctrl+l": tea.KeyCtrlL,
	"ctrl+n": tea.KeyCtrlN,
	"ctrl+p": tea.KeyCtrlP,
	"ctrl+b": tea.KeyCtrlB,
	"ctrl+f": tea.KeyCtrlF,
	// Function keys
	"f1":  tea.KeyF1,
	"f2":  tea.KeyF2,
	"f3":  tea.KeyF3,
	"f4":  tea.KeyF4,
	"f5":  tea.KeyF5,
	"f6":  tea.KeyF6,
	"f7":  tea.KeyF7,
	"f8":  tea.KeyF8,
	"f9":  tea.KeyF9,
	"f10": tea.KeyF10,
	"f11": tea.KeyF11,
	"f12": tea.KeyF12,
}

func createKeyMsg(key string) tea.KeyMsg {
	// Check if key is in the map
	if keyType, ok := keyTypeMap[key]; ok {
		return tea.KeyMsg{Type: keyType}
	}

	// Default: treat as runes (character input)
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key),
	}
}
