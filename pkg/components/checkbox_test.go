package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestCheckbox_Creation(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	checkbox := Checkbox(CheckboxProps{
		Label:   "Accept terms",
		Checked: checkedRef,
	})

	assert.NotNil(t, checkbox, "Checkbox component should be created")
	assert.Equal(t, "Checkbox", checkbox.Name(), "Component name should be 'Checkbox'")
}

func TestCheckbox_Rendering(t *testing.T) {
	tests := []struct {
		name        string
		label       string
		checked     bool
		wantContain string
	}{
		{
			name:        "unchecked with label",
			label:       "Enable feature",
			checked:     false,
			wantContain: "Enable feature",
		},
		{
			name:        "checked with label",
			label:       "Enable feature",
			checked:     true,
			wantContain: "Enable feature",
		},
		{
			name:    "empty label unchecked",
			label:   "",
			checked: false,
		},
		{
			name:    "empty label checked",
			label:   "",
			checked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkedRef := bubbly.NewRef(tt.checked)

			checkbox := Checkbox(CheckboxProps{
				Label:   tt.label,
				Checked: checkedRef,
			})

			checkbox.Init()
			view := checkbox.View()

			assert.NotEmpty(t, view, "View should not be empty")
			if tt.wantContain != "" {
				assert.Contains(t, view, tt.wantContain, "View should contain label")
			}

			// Check for checkbox indicators
			if tt.checked {
				// Should contain checked indicator
				if tt.label == "" {
					assert.True(t,
						view == "☑" || view == "[x]" || view == "[X]",
						"View should show checked indicator without label")
				} else {
					assert.True(t,
						view == "☑ "+tt.label || view == "[x] "+tt.label || view == "[X] "+tt.label,
						"View should show checked indicator with label")
				}
			} else {
				// Should contain unchecked indicator
				if tt.label == "" {
					assert.True(t,
						view == "☐" || view == "[ ]",
						"View should show unchecked indicator without label")
				} else {
					assert.True(t,
						view == "☐ "+tt.label || view == "[ ] "+tt.label,
						"View should show unchecked indicator with label")
				}
			}
		})
	}
}

func TestCheckbox_Toggle(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	checkbox := Checkbox(CheckboxProps{
		Label:   "Toggle me",
		Checked: checkedRef,
	})

	checkbox.Init()

	// Initial state: unchecked
	assert.False(t, checkedRef.GetTyped(), "Should start unchecked")

	// Emit toggle event
	checkbox.Emit("toggle", nil)

	// Should now be checked
	assert.True(t, checkedRef.GetTyped(), "Should be checked after toggle")

	// Toggle again
	checkbox.Emit("toggle", nil)

	// Should be unchecked again
	assert.False(t, checkedRef.GetTyped(), "Should be unchecked after second toggle")
}

func TestCheckbox_ValueBinding(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	checkbox := Checkbox(CheckboxProps{
		Label:   "Bound checkbox",
		Checked: checkedRef,
	})

	checkbox.Init()

	// Change value through ref
	checkedRef.Set(true)

	view := checkbox.View()
	// Should reflect checked state
	assert.True(t,
		view == "☑ Bound checkbox" || view == "[x] Bound checkbox" || view == "[X] Bound checkbox",
		"View should reflect checked state")

	// Change back to unchecked
	checkedRef.Set(false)

	view = checkbox.View()
	// Should reflect unchecked state
	assert.True(t,
		view == "☐ Bound checkbox" || view == "[ ] Bound checkbox",
		"View should reflect unchecked state")
}

func TestCheckbox_OnChangeCallback(t *testing.T) {
	checkedRef := bubbly.NewRef(false)
	callbackCalled := false
	var callbackValue bool

	checkbox := Checkbox(CheckboxProps{
		Label:   "Callback test",
		Checked: checkedRef,
		OnChange: func(checked bool) {
			callbackCalled = true
			callbackValue = checked
		},
	})

	checkbox.Init()

	// Toggle checkbox
	checkbox.Emit("toggle", nil)

	// OnChange should be called
	assert.True(t, callbackCalled, "OnChange callback should be called")
	assert.True(t, callbackValue, "Callback should receive true")

	// Reset and toggle again
	callbackCalled = false
	checkbox.Emit("toggle", nil)

	assert.True(t, callbackCalled, "OnChange callback should be called again")
	assert.False(t, callbackValue, "Callback should receive false")
}

func TestCheckbox_Disabled(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	checkbox := Checkbox(CheckboxProps{
		Label:    "Disabled checkbox",
		Checked:  checkedRef,
		Disabled: true,
	})

	checkbox.Init()

	// Try to toggle
	checkbox.Emit("toggle", nil)

	// Should still be unchecked (toggle ignored)
	assert.False(t, checkedRef.GetTyped(), "Disabled checkbox should not toggle")
}

func TestCheckbox_ThemeIntegration(t *testing.T) {
	checkedRef := bubbly.NewRef(true)

	// Checkbox uses DefaultTheme when no theme is provided
	checkbox := Checkbox(CheckboxProps{
		Label:   "Themed checkbox",
		Checked: checkedRef,
	})

	checkbox.Init()

	view := checkbox.View()
	assert.NotEmpty(t, view, "Checkbox should render with default theme")
}

func TestCheckbox_CustomStyle(t *testing.T) {
	checkedRef := bubbly.NewRef(false)
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	checkbox := Checkbox(CheckboxProps{
		Label:   "Styled checkbox",
		Checked: checkedRef,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	checkbox.Init()
	view := checkbox.View()

	assert.NotEmpty(t, view, "Checkbox should render with custom style")
}

func TestCheckbox_BubbletaIntegration(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	checkbox := Checkbox(CheckboxProps{
		Label:   "Integration test",
		Checked: checkedRef,
	})

	// Test Init
	cmd := checkbox.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update (no-op for Checkbox, handled by events)
	model, cmd := checkbox.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	// Test View
	view := checkbox.View()
	assert.NotEmpty(t, view, "View should return non-empty string")
}

func TestCheckbox_Props(t *testing.T) {
	checkedRef := bubbly.NewRef(true)
	props := CheckboxProps{
		Label:    "Test checkbox",
		Checked:  checkedRef,
		Disabled: false,
	}

	checkbox := Checkbox(props)
	checkbox.Init()

	// Props should be accessible
	retrievedProps := checkbox.Props()
	assert.NotNil(t, retrievedProps, "Props should be accessible")

	// Type assertion should work
	checkboxProps, ok := retrievedProps.(CheckboxProps)
	assert.True(t, ok, "Props should be of type CheckboxProps")
	assert.Equal(t, "Test checkbox", checkboxProps.Label, "Label should match")
	assert.False(t, checkboxProps.Disabled, "Disabled should match")
}

func TestCheckbox_EmptyLabel(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	checkbox := Checkbox(CheckboxProps{
		Label:   "",
		Checked: checkedRef,
	})

	checkbox.Init()
	view := checkbox.View()

	// Should render just the checkbox indicator
	assert.NotEmpty(t, view, "Checkbox should render even without label")
}

func TestCheckbox_LongLabel(t *testing.T) {
	longLabel := "This is a very long checkbox label that might wrap or extend beyond normal width"
	checkedRef := bubbly.NewRef(true)

	checkbox := Checkbox(CheckboxProps{
		Label:   longLabel,
		Checked: checkedRef,
	})

	checkbox.Init()
	view := checkbox.View()

	assert.Contains(t, view, longLabel, "Long label should be displayed")
}

func TestCheckbox_MultipleToggles(t *testing.T) {
	checkedRef := bubbly.NewRef(false)
	toggleCount := 0

	checkbox := Checkbox(CheckboxProps{
		Label:   "Multi-toggle",
		Checked: checkedRef,
		OnChange: func(checked bool) {
			toggleCount++
		},
	})

	checkbox.Init()

	// Toggle multiple times
	for i := 0; i < 5; i++ {
		checkbox.Emit("toggle", nil)
	}

	// Should have toggled 5 times
	assert.Equal(t, 5, toggleCount, "Should have called OnChange 5 times")
	// Final state should be checked (odd number of toggles)
	assert.True(t, checkedRef.GetTyped(), "Should be checked after 5 toggles")
}

func TestCheckbox_InitiallyChecked(t *testing.T) {
	checkedRef := bubbly.NewRef(true)

	checkbox := Checkbox(CheckboxProps{
		Label:   "Initially checked",
		Checked: checkedRef,
	})

	checkbox.Init()
	view := checkbox.View()

	// Should show as checked
	assert.True(t,
		view == "☑ Initially checked" || view == "[x] Initially checked" || view == "[X] Initially checked",
		"Should render as checked initially")
}

func TestCheckbox_NoOnChange(t *testing.T) {
	checkedRef := bubbly.NewRef(false)

	// Create checkbox without OnChange callback
	checkbox := Checkbox(CheckboxProps{
		Label:   "No callback",
		Checked: checkedRef,
	})

	checkbox.Init()

	// Toggle should still work
	checkbox.Emit("toggle", nil)
	assert.True(t, checkedRef.GetTyped(), "Should toggle even without OnChange")
}
