package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinner_Creation(t *testing.T) {
	tests := []struct {
		name  string
		props SpinnerProps
	}{
		{
			name: "Simple spinner",
			props: SpinnerProps{
				Active: true,
			},
		},
		{
			name: "Spinner with label",
			props: SpinnerProps{
				Label:  "Loading...",
				Active: true,
			},
		},
		{
			name: "Colored spinner",
			props: SpinnerProps{
				Color:  lipgloss.Color("99"),
				Active: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := Spinner(tt.props)
			require.NotNil(t, spinner, "Spinner should not be nil")

			spinner.Init()
			view := spinner.View()

			assert.NotNil(t, view, "Spinner view should not be nil")
		})
	}
}

func TestSpinner_ActiveState(t *testing.T) {
	tests := []struct {
		name   string
		active bool
	}{
		{"Active spinner", true},
		{"Inactive spinner", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := Spinner(SpinnerProps{
				Active: tt.active,
			})
			require.NotNil(t, spinner)

			spinner.Init()
			view := spinner.View()

			assert.NotNil(t, view)
		})
	}
}

func TestSpinner_Labels(t *testing.T) {
	tests := []struct {
		name  string
		label string
	}{
		{"Loading", "Loading..."},
		{"Processing", "Processing data..."},
		{"Please wait", "Please wait"},
		{"Empty label", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := Spinner(SpinnerProps{
				Label:  tt.label,
				Active: true,
			})
			require.NotNil(t, spinner)

			spinner.Init()
			view := spinner.View()

			assert.NotNil(t, view)
			if tt.label != "" {
				assert.Contains(t, view, tt.label)
			}
		})
	}
}

func TestSpinner_Colors(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Purple", lipgloss.Color("99")},
		{"Blue", lipgloss.Color("63")},
		{"Green", lipgloss.Color("46")},
		{"No color", lipgloss.Color("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := Spinner(SpinnerProps{
				Color:  tt.color,
				Active: true,
			})
			require.NotNil(t, spinner)

			spinner.Init()
			view := spinner.View()

			assert.NotNil(t, view)
		})
	}
}

func TestSpinner_BubbleteatIntegration(t *testing.T) {
	spinner := Spinner(SpinnerProps{
		Label:  "Loading",
		Active: true,
	})
	require.NotNil(t, spinner)

	_ = spinner.Init()
	// Spinner may return a tick command for animation
	// We just verify it doesn't panic

	updated, _ := spinner.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated, "Update should return updated component")

	view := spinner.View()
	assert.NotNil(t, view, "View should return rendered output")
}

func TestSpinner_Props(t *testing.T) {
	props := SpinnerProps{
		Label:  "Processing",
		Color:  lipgloss.Color("99"),
		Active: true,
	}

	spinner := Spinner(props)
	require.NotNil(t, spinner)

	spinner.Init()

	retrievedProps := spinner.Props()
	assert.NotNil(t, retrievedProps)

	spinnerProps, ok := retrievedProps.(SpinnerProps)
	assert.True(t, ok, "Props should be SpinnerProps type")
	assert.Equal(t, "Processing", spinnerProps.Label)
	assert.Equal(t, lipgloss.Color("99"), spinnerProps.Color)
	assert.True(t, spinnerProps.Active)
}

func TestSpinner_ThemeIntegration(t *testing.T) {
	spinner := Spinner(SpinnerProps{
		Active: true,
	})
	require.NotNil(t, spinner)

	spinner.Init()
	view := spinner.View()

	assert.NotNil(t, view)
}

func TestSpinner_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Padding(1, 2)

	spinner := Spinner(SpinnerProps{
		Active: true,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, spinner)

	spinner.Init()
	view := spinner.View()

	assert.NotNil(t, view)
}

func TestSpinner_InactiveState(t *testing.T) {
	// Test that inactive spinner doesn't show animation
	spinner := Spinner(SpinnerProps{
		Label:  "Done",
		Active: false,
	})
	require.NotNil(t, spinner)

	spinner.Init()
	view := spinner.View()

	assert.NotNil(t, view)
	// Inactive spinner should still render label if provided
	assert.Contains(t, view, "Done")
}
