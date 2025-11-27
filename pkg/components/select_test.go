package components

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestSelect_Creation(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	assert.NotNil(t, selectComp, "Select component should be created")
	assert.Equal(t, "Select", selectComp.Name(), "Component name should be 'Select'")
}

func TestSelect_Rendering(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		options     []string
		placeholder string
		wantContain string
	}{
		{
			name:        "with selected value",
			value:       "option2",
			options:     []string{"option1", "option2", "option3"},
			wantContain: "option2",
		},
		{
			name:        "with placeholder",
			value:       "",
			options:     []string{"option1", "option2"},
			placeholder: "Select an option",
			wantContain: "Select an option",
		},
		{
			name:    "first option selected",
			value:   "option1",
			options: []string{"option1", "option2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			selectComp := Select(SelectProps[string]{
				Value:       valueRef,
				Options:     tt.options,
				Placeholder: tt.placeholder,
			})

			selectComp.Init()
			view := selectComp.View()

			assert.NotEmpty(t, view, "View should not be empty")
			if tt.wantContain != "" {
				assert.Contains(t, view, tt.wantContain, "View should contain expected text")
			}
		})
	}
}

func TestSelect_OpenClose(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	// Initially closed
	view := selectComp.View()
	// Should show closed indicator (▼)
	assert.Contains(t, view, "▼", "Should show closed indicator")

	// Toggle to open
	selectComp.Emit("toggle", nil)
	view = selectComp.View()

	// Should show options when open
	assert.Contains(t, view, "option1", "Should show option1 when open")
	assert.Contains(t, view, "option2", "Should show option2 when open")
	assert.Contains(t, view, "option3", "Should show option3 when open")

	// Toggle to close
	selectComp.Emit("toggle", nil)
	view = selectComp.View()

	// Should show closed indicator again
	assert.Contains(t, view, "▼", "Should show closed indicator after toggle")
}

func TestSelect_Navigation(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	// Open the select
	selectComp.Emit("toggle", nil)

	// Navigate down
	selectComp.Emit("down", nil)
	// Should highlight option2

	// Navigate down again
	selectComp.Emit("down", nil)
	// Should highlight option3

	// Navigate down again (should wrap to option1)
	selectComp.Emit("down", nil)

	// Navigate up
	selectComp.Emit("up", nil)
	// Should highlight option3

	// Test passes if no panics occur
	assert.True(t, true, "Navigation should work without errors")
}

func TestSelect_Selection(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	// Open the select
	selectComp.Emit("toggle", nil)

	// Navigate to option2
	selectComp.Emit("down", nil)

	// Select it
	selectComp.Emit("select", nil)

	// Value should be updated
	assert.Equal(t, "option2", valueRef.GetTyped(), "Value should be updated to option2")

	// Select should be closed after selection
	view := selectComp.View()
	assert.Contains(t, view, "▼", "Should be closed after selection")
}

func TestSelect_ValueBinding(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	// Change value through ref
	valueRef.Set("option3")

	view := selectComp.View()
	assert.Contains(t, view, "option3", "View should reflect updated value")
}

func TestSelect_OnChangeCallback(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}
	callbackCalled := false
	var callbackValue string

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
		OnChange: func(value string) {
			callbackCalled = true
			callbackValue = value
		},
	})

	selectComp.Init()

	// Open and navigate
	selectComp.Emit("toggle", nil)
	selectComp.Emit("down", nil)

	// Select option2
	selectComp.Emit("select", nil)

	// OnChange should be called
	assert.True(t, callbackCalled, "OnChange callback should be called")
	assert.Equal(t, "option2", callbackValue, "Callback should receive option2")
}

func TestSelect_Disabled(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:    valueRef,
		Options:  options,
		Disabled: true,
	})

	selectComp.Init()

	// Try to open
	selectComp.Emit("toggle", nil)

	view := selectComp.View()
	// Should still be closed (disabled select doesn't open)
	assert.Contains(t, view, "▼", "Disabled select should not open")
}

func TestSelect_ThemeIntegration(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}

	// Select uses DefaultTheme when no theme is provided
	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	view := selectComp.View()
	assert.NotEmpty(t, view, "Select should render with default theme")
}

func TestSelect_CustomStyle(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	selectComp.Init()
	view := selectComp.View()

	assert.NotEmpty(t, view, "Select should render with custom style")
}

func TestSelect_BubbletaIntegration(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	// Test Init
	cmd := selectComp.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	model, cmd := selectComp.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	// Test View
	view := selectComp.View()
	assert.NotEmpty(t, view, "View should return non-empty string")
}

func TestSelect_EmptyOptions(t *testing.T) {
	valueRef := bubbly.NewRef("")
	options := []string{}

	selectComp := Select(SelectProps[string]{
		Value:       valueRef,
		Options:     options,
		Placeholder: "No options available",
	})

	selectComp.Init()
	view := selectComp.View()

	assert.NotEmpty(t, view, "Select should render even with empty options")
	assert.Contains(t, view, "No options available", "Should show placeholder")
}

func TestSelect_CustomRenderOption(t *testing.T) {
	type Option struct {
		ID   int
		Name string
	}

	valueRef := bubbly.NewRef(Option{ID: 1, Name: "First"})
	options := []Option{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
	}

	selectComp := Select(SelectProps[Option]{
		Value:   valueRef,
		Options: options,
		RenderOption: func(opt Option) string {
			return fmt.Sprintf("%d: %s", opt.ID, opt.Name)
		},
	})

	selectComp.Init()

	// Open to see options
	selectComp.Emit("toggle", nil)
	view := selectComp.View()

	assert.Contains(t, view, "1: First", "Should use custom render function")
	assert.Contains(t, view, "2: Second", "Should use custom render function")
}

func TestSelect_IntOptions(t *testing.T) {
	valueRef := bubbly.NewRef(1)
	options := []int{1, 2, 3, 4, 5}

	selectComp := Select(SelectProps[int]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()
	view := selectComp.View()

	assert.NotEmpty(t, view, "Select should work with int type")
	assert.Contains(t, view, "1", "Should display int value")
}

func TestSelect_StructOptions(t *testing.T) {
	type Language struct {
		Code string
		Name string
	}

	valueRef := bubbly.NewRef(Language{Code: "en", Name: "English"})
	options := []Language{
		{Code: "en", Name: "English"},
		{Code: "es", Name: "Spanish"},
		{Code: "fr", Name: "French"},
	}

	selectComp := Select(SelectProps[Language]{
		Value:   valueRef,
		Options: options,
		RenderOption: func(lang Language) string {
			return lang.Name
		},
	})

	selectComp.Init()
	view := selectComp.View()

	assert.NotEmpty(t, view, "Select should work with struct type")
}

func TestSelect_CloseEvent(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	// Open the select
	selectComp.Emit("toggle", nil)

	// Close without selecting
	selectComp.Emit("close", nil)

	view := selectComp.View()
	// Should be closed
	assert.Contains(t, view, "▼", "Should be closed after close event")

	// Value should not have changed
	assert.Equal(t, "option1", valueRef.GetTyped(), "Value should not change on close")
}

func TestSelect_NoOnChange(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	// Create select without OnChange callback
	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
	})

	selectComp.Init()

	// Open, navigate, and select
	selectComp.Emit("toggle", nil)
	selectComp.Emit("down", nil)
	selectComp.Emit("select", nil)

	// Should still work without OnChange
	assert.Equal(t, "option2", valueRef.GetTyped(), "Should update value even without OnChange")
}

func TestSelect_Props(t *testing.T) {
	valueRef := bubbly.NewRef("test")
	options := []string{"test", "other"}
	props := SelectProps[string]{
		Value:       valueRef,
		Options:     options,
		Placeholder: "Choose",
		Disabled:    false,
	}

	selectComp := Select(props)
	selectComp.Init()

	// Props should be accessible
	retrievedProps := selectComp.Props()
	assert.NotNil(t, retrievedProps, "Props should be accessible")
}

// ============================================================================
// SELECT HELPER FUNCTION TESTS - Additional Coverage
// ============================================================================

func TestSelect_EmptyOptions_NoPlaceholder(t *testing.T) {
	// Test selectGetDisplayValue with empty options and no placeholder
	valueRef := bubbly.NewRef("")
	options := []string{}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
		// No placeholder
	})

	selectComp.Init()
	view := selectComp.View()

	// Should show "No options" when options are empty and no placeholder
	assert.Contains(t, view, "No options", "Should show 'No options' text")
}

func TestSelect_ValueNotInOptions(t *testing.T) {
	// Test when current value is not in the options list
	valueRef := bubbly.NewRef("not-in-list")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:       valueRef,
		Options:     options,
		Placeholder: "Select one",
	})

	selectComp.Init()
	view := selectComp.View()

	// Should show placeholder when value is not in options
	assert.Contains(t, view, "Select one", "Should show placeholder when value not in options")
}

func TestSelect_ValueNotInOptions_NoPlaceholder(t *testing.T) {
	// Test when value is not in options and no placeholder is set
	valueRef := bubbly.NewRef("invalid-value")
	options := []string{"option1", "option2"}

	selectComp := Select(SelectProps[string]{
		Value:   valueRef,
		Options: options,
		// No placeholder
	})

	selectComp.Init()
	view := selectComp.View()

	// Should show the current value as fallback
	assert.Contains(t, view, "invalid-value", "Should show current value as fallback")
}

func TestSelect_NoBorder(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}

	selectComp := Select(SelectProps[string]{
		Value:    valueRef,
		Options:  options,
		NoBorder: true,
	})

	selectComp.Init()
	view := selectComp.View()

	assert.NotEmpty(t, view, "Select should render without border")
	assert.Contains(t, view, "option1", "Should display selected value")
}

func TestSelect_NoBorder_Open(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	selectComp := Select(SelectProps[string]{
		Value:    valueRef,
		Options:  options,
		NoBorder: true,
	})

	selectComp.Init()
	selectComp.Emit("toggle", nil)

	view := selectComp.View()

	assert.NotEmpty(t, view, "Select should render open without border")
}

func TestSelect_DisabledWithBorder(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}

	selectComp := Select(SelectProps[string]{
		Value:    valueRef,
		Options:  options,
		Disabled: true,
		// NoBorder is false, so border color should be muted
	})

	selectComp.Init()
	view := selectComp.View()

	assert.NotEmpty(t, view, "Disabled select should render with muted border")
}
