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

	// Create MouseMsg for left click
	mouseMsg := tea.MouseMsg{
		X:    x,
		Y:    y,
		Type: tea.MouseLeft,
	}

	// Send the message
	return ct.SendMessage(mouseMsg)
}

// createKeyMsg creates a KeyMsg from a key string.
// This handles the conversion from string representation to Bubbletea KeyMsg.
func createKeyMsg(key string) tea.KeyMsg {
	// Handle special keys
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+a":
		return tea.KeyMsg{Type: tea.KeyCtrlA}
	case "ctrl+e":
		return tea.KeyMsg{Type: tea.KeyCtrlE}
	case "ctrl+k":
		return tea.KeyMsg{Type: tea.KeyCtrlK}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case "ctrl+w":
		return tea.KeyMsg{Type: tea.KeyCtrlW}
	case "ctrl+l":
		return tea.KeyMsg{Type: tea.KeyCtrlL}
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	}

	// Handle function keys
	if len(key) >= 2 && key[0] == 'f' {
		switch key {
		case "f1":
			return tea.KeyMsg{Type: tea.KeyF1}
		case "f2":
			return tea.KeyMsg{Type: tea.KeyF2}
		case "f3":
			return tea.KeyMsg{Type: tea.KeyF3}
		case "f4":
			return tea.KeyMsg{Type: tea.KeyF4}
		case "f5":
			return tea.KeyMsg{Type: tea.KeyF5}
		case "f6":
			return tea.KeyMsg{Type: tea.KeyF6}
		case "f7":
			return tea.KeyMsg{Type: tea.KeyF7}
		case "f8":
			return tea.KeyMsg{Type: tea.KeyF8}
		case "f9":
			return tea.KeyMsg{Type: tea.KeyF9}
		case "f10":
			return tea.KeyMsg{Type: tea.KeyF10}
		case "f11":
			return tea.KeyMsg{Type: tea.KeyF11}
		case "f12":
			return tea.KeyMsg{Type: tea.KeyF12}
		}
	}

	// Default: treat as runes (character input)
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(key),
	}
}
