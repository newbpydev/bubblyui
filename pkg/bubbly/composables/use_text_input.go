package composables

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInputResult contains the text input state and methods.
type TextInputResult struct {
	// Model is the underlying Bubbles textinput model
	Model textinput.Model

	// Value returns the current input value
	Value func() string

	// SetValue sets the input value
	SetValue func(string)

	// Focus focuses the input
	Focus func()

	// Blur removes focus from the input
	Blur func()

	// Update handles Bubbletea messages
	Update func(tea.Msg) tea.Cmd

	// View renders the input
	View func() string

	// Reset clears the input
	Reset func()

	// CursorPosition returns the cursor position
	CursorPosition func() int

	// SetCursorPosition sets the cursor position
	SetCursorPosition func(int)
}

// UseTextInputConfig contains configuration for the text input.
type UseTextInputConfig struct {
	// Placeholder text when input is empty
	Placeholder string

	// CharLimit is the maximum number of characters (0 = no limit)
	CharLimit int

	// Width is the visual width of the input
	Width int

	// InitialValue is the starting value
	InitialValue string

	// EchoMode for password fields (default: textinput.EchoNormal)
	EchoMode textinput.EchoMode

	// ShowCursorPosition displays "[pos/len]" after the input
	ShowCursorPosition bool
}

// UseTextInput creates a text input with cursor support using Bubbles textinput.
//
// This composable provides a full-featured text input with:
// - Blinking cursor
// - Left/Right arrow navigation within text
// - Home/End keys to jump to start/end
// - Ctrl+A to select all
// - Backspace/Delete for character deletion
// - Insert mode for typing at cursor position
//
// Example:
//
//	input := composables.UseTextInput(composables.UseTextInputConfig{
//	    Placeholder: "Enter todo title...",
//	    CharLimit:   100,
//	    Width:       40,
//	})
//	input.Focus()
//
//	// In Update:
//	cmd := input.Update(msg)
//
//	// In View:
//	return input.View()
func UseTextInput(config UseTextInputConfig) *TextInputResult {
	ti := textinput.New()

	// Apply configuration
	if config.Placeholder != "" {
		ti.Placeholder = config.Placeholder
	}

	if config.CharLimit > 0 {
		ti.CharLimit = config.CharLimit
	}

	if config.Width > 0 {
		ti.Width = config.Width
	}

	if config.InitialValue != "" {
		ti.SetValue(config.InitialValue)
	}

	if config.EchoMode != 0 {
		ti.EchoMode = config.EchoMode
	}

	result := &TextInputResult{
		Model: ti,

		Value: func() string {
			return ti.Value()
		},

		SetValue: func(value string) {
			ti.SetValue(value)
		},

		Focus: func() {
			ti.Focus()
		},

		Blur: func() {
			ti.Blur()
		},

		Update: func(msg tea.Msg) tea.Cmd {
			var cmd tea.Cmd
			ti, cmd = ti.Update(msg)
			return cmd
		},

		View: func() string {
			view := ti.View()

			// Optionally show cursor position
			if config.ShowCursorPosition {
				pos := ti.Position()
				length := len(ti.Value())
				view += lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Render(fmt.Sprintf(" [%d/%d]", pos, length))
			}

			return view
		},

		Reset: func() {
			ti.SetValue("")
		},

		CursorPosition: func() int {
			return ti.Position()
		},

		SetCursorPosition: func(pos int) {
			ti.SetCursor(pos)
		},
	}

	return result
}

// BlinkCmd returns the cursor blink command.
// This should be returned from your model's Init() function.
func BlinkCmd() tea.Cmd {
	return textinput.Blink
}
