package composables

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUseTextInput_Default tests UseTextInput with default configuration
func TestUseTextInput_Default(t *testing.T) {
	// Act
	input := UseTextInput(UseTextInputConfig{})

	// Assert
	require.NotNil(t, input)
	assert.Empty(t, input.Value(), "initial value should be empty")
	assert.NotNil(t, input.Focus)
	assert.NotNil(t, input.Blur)
	assert.NotNil(t, input.Update)
	assert.NotNil(t, input.View)
	assert.NotNil(t, input.Reset)
	assert.NotNil(t, input.CursorPosition)
	assert.NotNil(t, input.SetCursorPosition)
}

// TestUseTextInput_WithPlaceholder tests UseTextInput with placeholder
func TestUseTextInput_WithPlaceholder(t *testing.T) {
	// Arrange
	config := UseTextInputConfig{
		Placeholder: "Enter text...",
	}

	// Act
	input := UseTextInput(config)

	// Assert
	require.NotNil(t, input)
	assert.Equal(t, "Enter text...", input.Model.Placeholder)
}

// TestUseTextInput_WithCharLimit tests UseTextInput with character limit
func TestUseTextInput_WithCharLimit(t *testing.T) {
	// Arrange
	config := UseTextInputConfig{
		CharLimit: 50,
	}

	// Act
	input := UseTextInput(config)

	// Assert
	require.NotNil(t, input)
	assert.Equal(t, 50, input.Model.CharLimit)
}

// TestUseTextInput_WithWidth tests UseTextInput with width
func TestUseTextInput_WithWidth(t *testing.T) {
	// Arrange
	config := UseTextInputConfig{
		Width: 40,
	}

	// Act
	input := UseTextInput(config)

	// Assert
	require.NotNil(t, input)
	assert.Equal(t, 40, input.Model.Width)
}

// TestUseTextInput_WithInitialValue tests UseTextInput with initial value
func TestUseTextInput_WithInitialValue(t *testing.T) {
	// Arrange
	config := UseTextInputConfig{
		InitialValue: "Hello World",
	}

	// Act
	input := UseTextInput(config)

	// Assert
	require.NotNil(t, input)
	assert.Equal(t, "Hello World", input.Value())
}

// TestUseTextInput_WithEchoMode tests UseTextInput with password echo mode
func TestUseTextInput_WithEchoMode(t *testing.T) {
	// Arrange
	config := UseTextInputConfig{
		EchoMode: textinput.EchoPassword,
	}

	// Act
	input := UseTextInput(config)

	// Assert
	require.NotNil(t, input)
	assert.Equal(t, textinput.EchoPassword, input.Model.EchoMode)
}

// TestUseTextInput_SetValue tests the SetValue function
func TestUseTextInput_SetValue(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{})

	// Act
	input.SetValue("New Value")

	// Assert
	assert.Equal(t, "New Value", input.Value())
}

// TestUseTextInput_Reset tests the Reset function
func TestUseTextInput_Reset(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{
		InitialValue: "Some text",
	})
	assert.Equal(t, "Some text", input.Value())

	// Act
	input.Reset()

	// Assert
	assert.Empty(t, input.Value())
}

// TestUseTextInput_FocusBlur tests the Focus and Blur functions
func TestUseTextInput_FocusBlur(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{})

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		input.Focus()
	})

	assert.NotPanics(t, func() {
		input.Blur()
	})
}

// TestUseTextInput_CursorPosition tests cursor position functions
func TestUseTextInput_CursorPosition(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{
		InitialValue: "Hello",
	})

	// Act - cursor starts at end
	pos := input.CursorPosition()

	// Assert
	assert.Equal(t, 5, pos, "cursor should be at end of text")

	// Act - set cursor position
	input.SetCursorPosition(2)

	// Assert
	assert.Equal(t, 2, input.CursorPosition())
}

// TestUseTextInput_Update tests the Update function
func TestUseTextInput_Update(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{})
	input.Focus()

	// Act - send a key message
	cmd := input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Assert - may or may not return a command
	_ = cmd // Command handling is internal to textinput
}

// TestUseTextInput_View tests the View function
func TestUseTextInput_View(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{
		InitialValue: "Test",
	})

	// Act
	view := input.View()

	// Assert
	assert.NotEmpty(t, view)
}

// TestUseTextInput_ViewWithCursorPosition tests View with cursor position enabled
func TestUseTextInput_ViewWithCursorPosition(t *testing.T) {
	// Arrange
	input := UseTextInput(UseTextInputConfig{
		InitialValue:       "Hello",
		ShowCursorPosition: true,
	})

	// Act
	view := input.View()

	// Assert - should contain position info
	assert.Contains(t, view, "[5/5]", "should show cursor position")
}

// TestUseTextInput_FullConfig tests UseTextInput with all config options
func TestUseTextInput_FullConfig(t *testing.T) {
	// Arrange
	config := UseTextInputConfig{
		Placeholder:        "Enter here...",
		CharLimit:          100,
		Width:              50,
		InitialValue:       "Initial",
		EchoMode:           textinput.EchoNormal,
		ShowCursorPosition: true,
	}

	// Act
	input := UseTextInput(config)

	// Assert
	require.NotNil(t, input)
	assert.Equal(t, "Enter here...", input.Model.Placeholder)
	assert.Equal(t, 100, input.Model.CharLimit)
	assert.Equal(t, 50, input.Model.Width)
	assert.Equal(t, "Initial", input.Value())
	assert.Equal(t, textinput.EchoNormal, input.Model.EchoMode)
}

// TestBlinkCmd tests the BlinkCmd function
func TestBlinkCmd(t *testing.T) {
	// Act
	cmd := BlinkCmd()

	// Assert - should return a valid command
	assert.NotNil(t, cmd)
}
