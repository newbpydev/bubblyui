package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestToggle_Creation(t *testing.T) {
	valueRef := bubbly.NewRef(false)

	toggle := Toggle(ToggleProps{
		Label: "Enable feature",
		Value: valueRef,
	})

	assert.NotNil(t, toggle, "Toggle component should be created")
	assert.Equal(t, "Toggle", toggle.Name(), "Component name should be 'Toggle'")
}

func TestToggle_Rendering(t *testing.T) {
	tests := []struct {
		name        string
		label       string
		value       bool
		wantContain string
	}{
		{
			name:        "off state with label",
			label:       "Dark mode",
			value:       false,
			wantContain: "Dark mode",
		},
		{
			name:        "on state with label",
			label:       "Dark mode",
			value:       true,
			wantContain: "Dark mode",
		},
		{
			name:  "empty label off",
			label: "",
			value: false,
		},
		{
			name:  "empty label on",
			label: "",
			value: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			toggle := Toggle(ToggleProps{
				Label: tt.label,
				Value: valueRef,
			})

			toggle.Init()
			view := toggle.View()

			assert.NotEmpty(t, view, "View should not be empty")
			if tt.wantContain != "" {
				assert.Contains(t, view, tt.wantContain, "View should contain label")
			}

			// Check for toggle indicators
			if tt.value {
				// Should show ON state
				if tt.label == "" {
					assert.True(t,
						view == "[ON ]" || view == "[●─]" || view == "[ON]",
						"View should show ON state without label")
				} else {
					assert.True(t,
						view == "[ON ] "+tt.label || view == "[●─] "+tt.label || view == "[ON] "+tt.label,
						"View should show ON state with label")
				}
			} else {
				// Should show OFF state
				if tt.label == "" {
					assert.True(t,
						view == "[OFF]" || view == "[─●]",
						"View should show OFF state without label")
				} else {
					assert.True(t,
						view == "[OFF] "+tt.label || view == "[─●] "+tt.label,
						"View should show OFF state with label")
				}
			}
		})
	}
}

func TestToggle_Toggle(t *testing.T) {
	valueRef := bubbly.NewRef(false)

	toggle := Toggle(ToggleProps{
		Label: "Toggle me",
		Value: valueRef,
	})

	toggle.Init()

	// Initial state: off
	assert.False(t, valueRef.GetTyped(), "Should start off")

	// Emit toggle event
	toggle.Emit("toggle", nil)

	// Should now be on
	assert.True(t, valueRef.GetTyped(), "Should be on after toggle")

	// Toggle again
	toggle.Emit("toggle", nil)

	// Should be off again
	assert.False(t, valueRef.GetTyped(), "Should be off after second toggle")
}

func TestToggle_ValueBinding(t *testing.T) {
	valueRef := bubbly.NewRef(false)

	toggle := Toggle(ToggleProps{
		Label: "Bound toggle",
		Value: valueRef,
	})

	toggle.Init()

	// Change value through ref
	valueRef.Set(true)

	view := toggle.View()
	// Should reflect on state
	assert.Contains(t, view, "ON", "View should reflect ON state")

	// Change back to off
	valueRef.Set(false)

	view = toggle.View()
	// Should reflect off state
	assert.Contains(t, view, "OFF", "View should reflect OFF state")
}

func TestToggle_OnChangeCallback(t *testing.T) {
	valueRef := bubbly.NewRef(false)
	callbackCalled := false
	var callbackValue bool

	toggle := Toggle(ToggleProps{
		Label: "Callback test",
		Value: valueRef,
		OnChange: func(value bool) {
			callbackCalled = true
			callbackValue = value
		},
	})

	toggle.Init()

	// Toggle
	toggle.Emit("toggle", nil)

	// OnChange should be called
	assert.True(t, callbackCalled, "OnChange callback should be called")
	assert.True(t, callbackValue, "Callback should receive true")

	// Reset and toggle again
	callbackCalled = false
	toggle.Emit("toggle", nil)

	assert.True(t, callbackCalled, "OnChange callback should be called again")
	assert.False(t, callbackValue, "Callback should receive false")
}

func TestToggle_Disabled(t *testing.T) {
	valueRef := bubbly.NewRef(false)

	toggle := Toggle(ToggleProps{
		Label:    "Disabled toggle",
		Value:    valueRef,
		Disabled: true,
	})

	toggle.Init()

	// Try to toggle
	toggle.Emit("toggle", nil)

	// Should still be off (toggle ignored)
	assert.False(t, valueRef.GetTyped(), "Disabled toggle should not toggle")
}

func TestToggle_ThemeIntegration(t *testing.T) {
	valueRef := bubbly.NewRef(true)

	// Toggle uses DefaultTheme when no theme is provided
	toggle := Toggle(ToggleProps{
		Label: "Themed toggle",
		Value: valueRef,
	})

	toggle.Init()

	view := toggle.View()
	assert.NotEmpty(t, view, "Toggle should render with default theme")
}

func TestToggle_CustomStyle(t *testing.T) {
	valueRef := bubbly.NewRef(false)
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	toggle := Toggle(ToggleProps{
		Label: "Styled toggle",
		Value: valueRef,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	toggle.Init()
	view := toggle.View()

	assert.NotEmpty(t, view, "Toggle should render with custom style")
}

func TestToggle_BubbletaIntegration(t *testing.T) {
	valueRef := bubbly.NewRef(false)

	toggle := Toggle(ToggleProps{
		Label: "Integration test",
		Value: valueRef,
	})

	// Test Init
	cmd := toggle.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	model, cmd := toggle.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	// Test View
	view := toggle.View()
	assert.NotEmpty(t, view, "View should return non-empty string")
}

func TestToggle_NoOnChange(t *testing.T) {
	valueRef := bubbly.NewRef(false)

	// Create toggle without OnChange callback
	toggle := Toggle(ToggleProps{
		Label: "No callback",
		Value: valueRef,
	})

	toggle.Init()

	// Toggle should still work
	toggle.Emit("toggle", nil)
	assert.True(t, valueRef.GetTyped(), "Should toggle even without OnChange")
}
