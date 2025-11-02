package components

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestInput_Creation(t *testing.T) {
	valueRef := bubbly.NewRef("")

	input := Input(InputProps{
		Value:       valueRef,
		Placeholder: "Enter text",
		Type:        InputText,
	})

	assert.NotNil(t, input, "Input component should be created")
	assert.Equal(t, "Input", input.Name(), "Component name should be 'Input'")
}

func TestInput_Rendering(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		placeholder string
		wantContain string
	}{
		{
			name:        "empty with placeholder",
			value:       "",
			placeholder: "Enter name",
			wantContain: "Enter name",
		},
		{
			name:        "with value",
			value:       "John Doe",
			placeholder: "Enter name",
			wantContain: "John Doe",
		},
		{
			name:        "empty no placeholder",
			value:       "",
			placeholder: "",
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			input := Input(InputProps{
				Value:       valueRef,
				Placeholder: tt.placeholder,
				Type:        InputText,
			})

			input.Init()
			view := input.View()

			if tt.wantContain != "" {
				assert.Contains(t, view, tt.wantContain, "View should contain expected text")
			}
		})
	}
}

func TestInput_ValueBinding(t *testing.T) {
	valueRef := bubbly.NewRef("initial")

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
	})

	input.Init()

	// Change value through ref
	valueRef.Set("updated")

	view := input.View()
	assert.Contains(t, view, "updated", "View should reflect updated value")
}

func TestInput_Validation(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		newValue      string
		validate      func(string) error
		wantError     bool
		wantErrorText string
	}{
		{
			name:         "valid input",
			initialValue: "",
			newValue:     "valid@email.com",
			validate: func(s string) error {
				if s != "" && !strings.Contains(s, "@") {
					return errors.New("invalid email")
				}
				return nil
			},
			wantError: false,
		},
		{
			name:         "invalid input",
			initialValue: "",
			newValue:     "invalid",
			validate: func(s string) error {
				if s != "" && !strings.Contains(s, "@") {
					return errors.New("invalid email")
				}
				return nil
			},
			wantError:     true,
			wantErrorText: "invalid email",
		},
		{
			name:         "no validation",
			initialValue: "",
			newValue:     "anything",
			validate:     nil,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.initialValue)

			input := Input(InputProps{
				Value:    valueRef,
				Type:     InputText,
				Validate: tt.validate,
			})

			input.Init()

			// Change value to trigger validation
			valueRef.Set(tt.newValue)

			view := input.View()

			if tt.wantError {
				assert.Contains(t, view, tt.wantErrorText, "View should show error message")
			} else if tt.wantErrorText != "" {
				assert.NotContains(t, view, tt.wantErrorText, "View should not show error message")
			}
		})
	}
}

func TestInput_FocusState(t *testing.T) {
	valueRef := bubbly.NewRef("")

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
	})

	input.Init()

	// Emit focus event
	input.Emit("focus", nil)

	// View should reflect focused state (different border color)
	view := input.View()
	assert.NotEmpty(t, view, "View should render when focused")

	// Emit blur event
	input.Emit("blur", nil)

	view = input.View()
	assert.NotEmpty(t, view, "View should render when blurred")
}

func TestInput_PasswordMasking(t *testing.T) {
	tests := []struct {
		name      string
		inputType InputType
		value     string
		wantShow  bool
	}{
		{
			name:      "text type shows value",
			inputType: InputText,
			value:     "password123",
			wantShow:  true,
		},
		{
			name:      "password type masks value",
			inputType: InputPassword,
			value:     "password123",
			wantShow:  false,
		},
		{
			name:      "email type shows value",
			inputType: InputEmail,
			value:     "test@example.com",
			wantShow:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			input := Input(InputProps{
				Value: valueRef,
				Type:  tt.inputType,
			})

			input.Init()
			view := input.View()

			if tt.wantShow {
				assert.Contains(t, view, tt.value, "Value should be visible")
			} else {
				assert.NotContains(t, view, tt.value, "Value should be masked")
				// Should contain asterisks instead
				assert.Contains(t, view, "*", "Should show asterisks for password")
			}
		})
	}
}

func TestInput_OnChangeCallback(t *testing.T) {
	valueRef := bubbly.NewRef("")
	callbackCalled := false
	var callbackValue string

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
		OnChange: func(value string) {
			callbackCalled = true
			callbackValue = value
		},
	})

	input.Init()

	// Change value
	valueRef.Set("new value")

	// OnChange should be called (via Watch)
	// Note: Watch is synchronous in tests
	assert.True(t, callbackCalled, "OnChange callback should be called")
	assert.Equal(t, "new value", callbackValue, "Callback should receive new value")
}

func TestInput_OnBlurCallback(t *testing.T) {
	valueRef := bubbly.NewRef("")
	callbackCalled := false

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
		OnBlur: func() {
			callbackCalled = true
		},
	})

	input.Init()

	// Emit blur event
	input.Emit("blur", nil)

	assert.True(t, callbackCalled, "OnBlur callback should be called")
}

func TestInput_ThemeIntegration(t *testing.T) {
	valueRef := bubbly.NewRef("test")

	// Input uses DefaultTheme when no theme is provided
	// Theme injection happens through parent components in real usage
	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
	})

	input.Init()

	view := input.View()
	assert.NotEmpty(t, view, "Input should render with default theme")
}

func TestInput_CustomStyle(t *testing.T) {
	valueRef := bubbly.NewRef("test")
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	input.Init()
	view := input.View()

	assert.NotEmpty(t, view, "Input should render with custom style")
}

func TestInput_Width(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "default width",
			width: 0, // Should use default 30
		},
		{
			name:  "custom width 50",
			width: 50,
		},
		{
			name:  "small width 10",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef("test")

			input := Input(InputProps{
				Value: valueRef,
				Type:  InputText,
				Width: tt.width,
			})

			input.Init()
			view := input.View()

			assert.NotEmpty(t, view, "Input should render with specified width")
		})
	}
}

func TestInput_BubbletaIntegration(t *testing.T) {
	valueRef := bubbly.NewRef("")

	input := Input(InputProps{
		Value:       valueRef,
		Placeholder: "Enter text",
		Type:        InputText,
	})

	// Test Init
	cmd := input.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update (no-op for Input, handled by events)
	model, cmd := input.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	// Test View
	view := input.View()
	assert.NotEmpty(t, view, "View should return non-empty string")
}

func TestInput_ErrorDisplay(t *testing.T) {
	valueRef := bubbly.NewRef("")

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
		Validate: func(s string) error {
			if s != "" {
				return errors.New("validation failed")
			}
			return nil
		},
	})

	input.Init()

	// Set invalid value to trigger validation
	valueRef.Set("invalid")

	view := input.View()

	// Error message should be displayed
	assert.Contains(t, view, "validation failed", "Error message should be displayed")
}

func TestInput_InputEvent(t *testing.T) {
	valueRef := bubbly.NewRef("")

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
	})

	input.Init()

	// Emit input event with new value
	input.Emit("input", "new text")

	// Value should be updated
	assert.Equal(t, "new text", valueRef.Get(), "Value should be updated via input event")
}

func TestInput_DefaultType(t *testing.T) {
	valueRef := bubbly.NewRef("")

	// Create input without specifying type
	input := Input(InputProps{
		Value: valueRef,
	})

	input.Init()
	view := input.View()

	assert.NotEmpty(t, view, "Input should render with default type")
}

func TestInput_EmptyValue(t *testing.T) {
	valueRef := bubbly.NewRef("")

	input := Input(InputProps{
		Value:       valueRef,
		Placeholder: "Type here",
		Type:        InputText,
	})

	input.Init()
	view := input.View()

	// Should show placeholder when empty
	assert.Contains(t, view, "Type here", "Should show placeholder when value is empty")
}

func TestInput_LongValue(t *testing.T) {
	longValue := strings.Repeat("a", 100)
	valueRef := bubbly.NewRef(longValue)

	input := Input(InputProps{
		Value: valueRef,
		Type:  InputText,
		Width: 30,
	})

	input.Init()
	view := input.View()

	assert.NotEmpty(t, view, "Input should handle long values")
}

func TestInput_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "unicode characters",
			value: "Hello 世界 🌍",
		},
		{
			name:  "special symbols",
			value: "!@#$%^&*()",
		},
		{
			name:  "newlines and tabs",
			value: "line1\nline2\ttab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			input := Input(InputProps{
				Value: valueRef,
				Type:  InputText,
			})

			input.Init()
			view := input.View()

			assert.NotEmpty(t, view, "Input should handle special characters")
		})
	}
}

func TestInput_Props(t *testing.T) {
	valueRef := bubbly.NewRef("test")
	props := InputProps{
		Value:       valueRef,
		Placeholder: "Enter text",
		Type:        InputEmail,
		Width:       40,
	}

	input := Input(props)
	input.Init()

	// Props should be accessible
	retrievedProps := input.Props()
	assert.NotNil(t, retrievedProps, "Props should be accessible")

	// Type assertion should work
	inputProps, ok := retrievedProps.(InputProps)
	assert.True(t, ok, "Props should be of type InputProps")
	assert.Equal(t, "Enter text", inputProps.Placeholder, "Placeholder should match")
	assert.Equal(t, InputEmail, inputProps.Type, "Type should match")
	assert.Equal(t, 40, inputProps.Width, "Width should match")
}
