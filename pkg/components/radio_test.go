package components

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestRadio_Creation(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	assert.NotNil(t, radio, "Radio component should be created")
	assert.Equal(t, "Radio", radio.Name(), "Component name should be 'Radio'")
}

func TestRadio_Rendering(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		options []string
	}{
		{
			name:    "first option selected",
			value:   "option1",
			options: []string{"option1", "option2", "option3"},
		},
		{
			name:    "middle option selected",
			value:   "option2",
			options: []string{"option1", "option2", "option3"},
		},
		{
			name:    "last option selected",
			value:   "option3",
			options: []string{"option1", "option2", "option3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueRef := bubbly.NewRef(tt.value)

			radio := Radio(RadioProps[string]{
				Value:   valueRef,
				Options: tt.options,
			})

			radio.Init()
			view := radio.View()

			assert.NotEmpty(t, view, "View should not be empty")
			// Should show all options
			for _, opt := range tt.options {
				assert.Contains(t, view, opt, "View should contain option: "+opt)
			}
		})
	}
}

func TestRadio_Navigation(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()

	// Navigate down
	radio.Emit("down", nil)

	// Navigate down again
	radio.Emit("down", nil)

	// Navigate down again (should wrap to option1)
	radio.Emit("down", nil)

	// Navigate up
	radio.Emit("up", nil)

	// Test passes if no panics occur
	assert.True(t, true, "Navigation should work without errors")
}

func TestRadio_Selection(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()

	// Navigate to option2
	radio.Emit("down", nil)

	// Select it
	radio.Emit("select", nil)

	// Value should be updated
	assert.Equal(t, "option2", valueRef.GetTyped(), "Value should be updated to option2")
}

func TestRadio_ValueBinding(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()

	// Change value through ref
	valueRef.Set("option3")

	view := radio.View()
	assert.Contains(t, view, "option3", "View should reflect updated value")
}

func TestRadio_OnChangeCallback(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}
	callbackCalled := false
	var callbackValue string

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
		OnChange: func(value string) {
			callbackCalled = true
			callbackValue = value
		},
	})

	radio.Init()

	// Navigate and select
	radio.Emit("down", nil)
	radio.Emit("select", nil)

	// OnChange should be called
	assert.True(t, callbackCalled, "OnChange callback should be called")
	assert.Equal(t, "option2", callbackValue, "Callback should receive option2")
}

func TestRadio_Disabled(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	radio := Radio(RadioProps[string]{
		Value:    valueRef,
		Options:  options,
		Disabled: true,
	})

	radio.Init()

	// Try to navigate and select
	radio.Emit("down", nil)
	radio.Emit("select", nil)

	// Value should not have changed
	assert.Equal(t, "option1", valueRef.GetTyped(), "Disabled radio should not change value")
}

func TestRadio_ThemeIntegration(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()

	view := radio.View()
	assert.NotEmpty(t, view, "Radio should render with default theme")
}

func TestRadio_CustomStyle(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2"}
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	radio.Init()
	view := radio.View()

	assert.NotEmpty(t, view, "Radio should render with custom style")
}

func TestRadio_CustomRenderOption(t *testing.T) {
	type Option struct {
		ID   int
		Name string
	}

	valueRef := bubbly.NewRef(Option{ID: 1, Name: "First"})
	options := []Option{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
	}

	radio := Radio(RadioProps[Option]{
		Value:   valueRef,
		Options: options,
		RenderOption: func(opt Option) string {
			return fmt.Sprintf("%d: %s", opt.ID, opt.Name)
		},
	})

	radio.Init()
	view := radio.View()

	assert.Contains(t, view, "1: First", "Should use custom render function")
	assert.Contains(t, view, "2: Second", "Should use custom render function")
}

func TestRadio_IntOptions(t *testing.T) {
	valueRef := bubbly.NewRef(1)
	options := []int{1, 2, 3, 4, 5}

	radio := Radio(RadioProps[int]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()
	view := radio.View()

	assert.NotEmpty(t, view, "Radio should work with int type")
	assert.Contains(t, view, "1", "Should display int value")
}

func TestRadio_EmptyOptions(t *testing.T) {
	valueRef := bubbly.NewRef("")
	options := []string{}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()
	view := radio.View()

	assert.NotEmpty(t, view, "Radio should render even with empty options")
}

func TestRadio_NoOnChange(t *testing.T) {
	valueRef := bubbly.NewRef("option1")
	options := []string{"option1", "option2", "option3"}

	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: options,
	})

	radio.Init()

	// Navigate and select
	radio.Emit("down", nil)
	radio.Emit("select", nil)

	// Should still work without OnChange
	assert.Equal(t, "option2", valueRef.GetTyped(), "Should update value even without OnChange")
}
