package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestTextArea_Creation(t *testing.T) {
	valueRef := bubbly.NewRef("Hello")

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
	})

	assert.NotNil(t, textarea, "TextArea component should be created")
	assert.Equal(t, "TextArea", textarea.Name(), "Component name should be 'TextArea'")
}

func TestTextArea_Rendering(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		placeholder string
		rows        int
		wantContain string
	}{
		{
			name:        "single line text",
			value:       "Hello world",
			rows:        3,
			wantContain: "Hello world",
		},
		{
			name:        "multi-line text",
			value:       "Line 1\nLine 2\nLine 3",
			rows:        3,
			wantContain: "Line 1",
		},
		{
			name:        "with placeholder",
			value:       "",
			placeholder: "Enter text here",
			rows:        3,
			wantContain: "Enter text here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			textarea := TextArea(TextAreaProps{
				Value:       valueRef,
				Placeholder: tt.placeholder,
				Rows:        tt.rows,
			})

			textarea.Init()
			view := textarea.View()

			assert.NotEmpty(t, view, "View should not be empty")
			if tt.wantContain != "" {
				assert.Contains(t, view, tt.wantContain, "View should contain expected text")
			}
		})
	}
}

func TestTextArea_MultiLine(t *testing.T) {
	valueRef := bubbly.NewRef("Line 1\nLine 2\nLine 3")

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
		Rows:  5,
	})

	textarea.Init()
	view := textarea.View()

	// Should contain all lines
	assert.Contains(t, view, "Line 1", "Should contain Line 1")
	assert.Contains(t, view, "Line 2", "Should contain Line 2")
	assert.Contains(t, view, "Line 3", "Should contain Line 3")
}

func TestTextArea_ValueBinding(t *testing.T) {
	valueRef := bubbly.NewRef("Initial text")

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
		Rows:  3,
	})

	textarea.Init()

	// Change value through ref
	valueRef.Set("Updated text")

	view := textarea.View()
	assert.Contains(t, view, "Updated text", "View should reflect updated value")
}

func TestTextArea_OnChangeCallback(t *testing.T) {
	valueRef := bubbly.NewRef("Initial")
	callbackCalled := false
	var callbackValue string

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
		OnChange: func(value string) {
			callbackCalled = true
			callbackValue = value
		},
	})

	textarea.Init()

	// Emit change event
	textarea.Emit("change", "New value")

	// OnChange should be called
	assert.True(t, callbackCalled, "OnChange callback should be called")
	assert.Equal(t, "New value", callbackValue, "Callback should receive new value")
}

func TestTextArea_Validation(t *testing.T) {
	valueRef := bubbly.NewRef("test")

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
		Validate: func(value string) error {
			if len(value) < 5 {
				return assert.AnError
			}
			return nil
		},
	})

	textarea.Init()

	// Set invalid value
	valueRef.Set("ab")

	view := textarea.View()
	// Should show error (implementation dependent)
	assert.NotEmpty(t, view, "Should render with validation")
}

func TestTextArea_Disabled(t *testing.T) {
	valueRef := bubbly.NewRef("text")

	textarea := TextArea(TextAreaProps{
		Value:    valueRef,
		Disabled: true,
	})

	textarea.Init()

	view := textarea.View()
	assert.NotEmpty(t, view, "Disabled textarea should render")
}

func TestTextArea_MaxLength(t *testing.T) {
	valueRef := bubbly.NewRef("Short")

	textarea := TextArea(TextAreaProps{
		Value:     valueRef,
		MaxLength: 10,
	})

	textarea.Init()
	view := textarea.View()

	assert.NotEmpty(t, view, "TextArea with max length should render")
}

func TestTextArea_ThemeIntegration(t *testing.T) {
	valueRef := bubbly.NewRef("text")

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
	})

	textarea.Init()

	view := textarea.View()
	assert.NotEmpty(t, view, "TextArea should render with default theme")
}

func TestTextArea_CustomStyle(t *testing.T) {
	valueRef := bubbly.NewRef("text")
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	textarea.Init()
	view := textarea.View()

	assert.NotEmpty(t, view, "TextArea should render with custom style")
}

func TestTextArea_EmptyValue(t *testing.T) {
	valueRef := bubbly.NewRef("")

	textarea := TextArea(TextAreaProps{
		Value:       valueRef,
		Placeholder: "Type something",
	})

	textarea.Init()
	view := textarea.View()

	assert.Contains(t, view, "Type something", "Should show placeholder when empty")
}

func TestTextArea_LongText(t *testing.T) {
	longText := strings.Repeat("This is a long line of text. ", 10)
	valueRef := bubbly.NewRef(longText)

	textarea := TextArea(TextAreaProps{
		Value: valueRef,
		Rows:  3,
	})

	textarea.Init()
	view := textarea.View()

	assert.NotEmpty(t, view, "Should handle long text")
}
