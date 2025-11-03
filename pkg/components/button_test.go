package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestButton_Creation(t *testing.T) {
	tests := []struct {
		name  string
		props ButtonProps
	}{
		{
			name: "Primary button",
			props: ButtonProps{
				Label:   "Submit",
				Variant: ButtonPrimary,
			},
		},
		{
			name: "Secondary button",
			props: ButtonProps{
				Label:   "Cancel",
				Variant: ButtonSecondary,
			},
		},
		{
			name: "Danger button",
			props: ButtonProps{
				Label:   "Delete",
				Variant: ButtonDanger,
			},
		},
		{
			name: "Disabled button",
			props: ButtonProps{
				Label:    "Disabled",
				Variant:  ButtonPrimary,
				Disabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := Button(tt.props)
			require.NotNil(t, button, "Button should not be nil")

			// Initialize component
			button.Init()

			// Verify component can render
			view := button.View()
			assert.NotEmpty(t, view, "Button view should not be empty")
			assert.Contains(t, view, tt.props.Label, "Button should contain label")
		})
	}
}

func TestButton_Variants(t *testing.T) {
	tests := []struct {
		name    string
		variant ButtonVariant
	}{
		{"Primary variant", ButtonPrimary},
		{"Secondary variant", ButtonSecondary},
		{"Danger variant", ButtonDanger},
		{"Success variant", ButtonSuccess},
		{"Warning variant", ButtonWarning},
		{"Info variant", ButtonInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := Button(ButtonProps{
				Label:   "Test",
				Variant: tt.variant,
			})
			require.NotNil(t, button)

			button.Init()
			view := button.View()

			// Verify button renders with variant
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestButton_DisabledState(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		wantText string
	}{
		{
			name:     "Enabled button",
			disabled: false,
			wantText: "Click Me",
		},
		{
			name:     "Disabled button",
			disabled: true,
			wantText: "Click Me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := Button(ButtonProps{
				Label:    tt.wantText,
				Variant:  ButtonPrimary,
				Disabled: tt.disabled,
			})
			require.NotNil(t, button)

			button.Init()
			view := button.View()

			assert.Contains(t, view, tt.wantText)
			// Disabled buttons should still render but with different styling
			assert.NotEmpty(t, view)
		})
	}
}

func TestButton_ClickEvent(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		wantCall bool
	}{
		{
			name:     "Enabled button fires click",
			disabled: false,
			wantCall: true,
		},
		{
			name:     "Disabled button does not fire click",
			disabled: true,
			wantCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clicked := false
			button := Button(ButtonProps{
				Label:    "Click",
				Variant:  ButtonPrimary,
				Disabled: tt.disabled,
				OnClick: func() {
					clicked = true
				},
			})
			require.NotNil(t, button)

			button.Init()

			// Simulate click event
			button.Emit("click", nil)

			if tt.wantCall {
				assert.True(t, clicked, "OnClick should be called for enabled button")
			} else {
				assert.False(t, clicked, "OnClick should not be called for disabled button")
			}
		})
	}
}

func TestButton_OnClickNil(t *testing.T) {
	// Test that button doesn't panic when OnClick is nil
	button := Button(ButtonProps{
		Label:   "Test",
		Variant: ButtonPrimary,
		OnClick: nil, // No handler
	})
	require.NotNil(t, button)

	button.Init()

	// Should not panic
	assert.NotPanics(t, func() {
		button.Emit("click", nil)
	})
}

func TestButton_Rendering(t *testing.T) {
	tests := []struct {
		name     string
		props    ButtonProps
		contains []string
	}{
		{
			name: "Primary button renders with label",
			props: ButtonProps{
				Label:   "Submit Form",
				Variant: ButtonPrimary,
			},
			contains: []string{"Submit Form"},
		},
		{
			name: "Button with special characters",
			props: ButtonProps{
				Label:   "Save & Exit",
				Variant: ButtonSecondary,
			},
			contains: []string{"Save & Exit"},
		},
		{
			name: "Button with emoji",
			props: ButtonProps{
				Label:   "✓ Confirm",
				Variant: ButtonSuccess,
			},
			contains: []string{"✓ Confirm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			button := Button(tt.props)
			require.NotNil(t, button)

			button.Init()
			view := button.View()

			for _, text := range tt.contains {
				assert.Contains(t, view, text)
			}
		})
	}
}

func TestButton_BubbleteatIntegration(t *testing.T) {
	// Test that button works with Bubbletea Update cycle
	button := Button(ButtonProps{
		Label:   "Test",
		Variant: ButtonPrimary,
	})
	require.NotNil(t, button)

	// Init
	cmd := button.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")

	// Update with key message
	updated, cmd := button.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated, "Update should return updated component")
	assert.Nil(t, cmd, "Update should return nil cmd for key messages")

	// View
	view := button.View()
	assert.NotEmpty(t, view, "View should return rendered output")
}

func TestButton_Props(t *testing.T) {
	props := ButtonProps{
		Label:    "Test Button",
		Variant:  ButtonDanger,
		Disabled: true,
		OnClick: func() {
			// Test handler
		},
	}

	button := Button(props)
	require.NotNil(t, button)

	button.Init()

	// Verify props are accessible
	retrievedProps := button.Props()
	assert.NotNil(t, retrievedProps)

	buttonProps, ok := retrievedProps.(ButtonProps)
	assert.True(t, ok, "Props should be ButtonProps type")
	assert.Equal(t, "Test Button", buttonProps.Label)
	assert.Equal(t, ButtonDanger, buttonProps.Variant)
	assert.True(t, buttonProps.Disabled)
}

func TestButton_MultipleClicks(t *testing.T) {
	clickCount := 0
	button := Button(ButtonProps{
		Label:   "Counter",
		Variant: ButtonPrimary,
		OnClick: func() {
			clickCount++
		},
	})
	require.NotNil(t, button)

	button.Init()

	// Click multiple times
	for i := 0; i < 5; i++ {
		button.Emit("click", nil)
	}

	assert.Equal(t, 5, clickCount, "Should handle multiple clicks")
}

func TestButton_DefaultVariant(t *testing.T) {
	// Test button with empty variant defaults to primary
	button := Button(ButtonProps{
		Label:   "Default",
		Variant: "", // Empty variant
	})
	require.NotNil(t, button)

	button.Init()
	view := button.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Default")
}

func TestButton_LongLabel(t *testing.T) {
	longLabel := strings.Repeat("Very Long Button Label ", 10)
	button := Button(ButtonProps{
		Label:   longLabel,
		Variant: ButtonPrimary,
	})
	require.NotNil(t, button)

	button.Init()
	view := button.View()

	assert.NotEmpty(t, view)
	// Should handle long labels without panic
}

func TestButton_EmptyLabel(t *testing.T) {
	button := Button(ButtonProps{
		Label:   "",
		Variant: ButtonPrimary,
	})
	require.NotNil(t, button)

	button.Init()
	view := button.View()

	// Should render even with empty label
	assert.NotEmpty(t, view)
}
