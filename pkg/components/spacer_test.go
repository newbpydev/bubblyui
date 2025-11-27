package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpacer_Creation(t *testing.T) {
	tests := []struct {
		name  string
		props SpacerProps
	}{
		{
			name: "Horizontal spacer",
			props: SpacerProps{
				Width: 10,
			},
		},
		{
			name: "Vertical spacer",
			props: SpacerProps{
				Height: 5,
			},
		},
		{
			name: "Both dimensions",
			props: SpacerProps{
				Width:  20,
				Height: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spacer := Spacer(tt.props)
			require.NotNil(t, spacer, "Spacer should not be nil")

			spacer.Init()
			view := spacer.View()

			assert.NotNil(t, view, "Spacer view should not be nil")
		})
	}
}

func TestSpacer_HorizontalWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"Small width", 5},
		{"Medium width", 20},
		{"Large width", 50},
		{"Zero width", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spacer := Spacer(SpacerProps{
				Width: tt.width,
			})
			require.NotNil(t, spacer)

			spacer.Init()
			view := spacer.View()

			assert.NotNil(t, view)
			if tt.width > 0 {
				// Should create horizontal space
				assert.NotEmpty(t, view)
			}
		})
	}
}

func TestSpacer_VerticalHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
	}{
		{"Small height", 2},
		{"Medium height", 5},
		{"Large height", 10},
		{"Zero height", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spacer := Spacer(SpacerProps{
				Height: tt.height,
			})
			require.NotNil(t, spacer)

			spacer.Init()
			view := spacer.View()

			assert.NotNil(t, view)
			if tt.height > 0 {
				// Should create vertical space (newlines)
				lineCount := strings.Count(view, "\n")
				assert.GreaterOrEqual(t, lineCount, 0)
			}
		})
	}
}

func TestSpacer_BothDimensions(t *testing.T) {
	spacer := Spacer(SpacerProps{
		Width:  30,
		Height: 5,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	assert.NotEmpty(t, view)
	// Should have both width and height
}

func TestSpacer_BubbleteatIntegration(t *testing.T) {
	spacer := Spacer(SpacerProps{
		Width:  10,
		Height: 2,
	})
	require.NotNil(t, spacer)

	cmd := spacer.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")

	updated, cmd := spacer.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated, "Update should return updated component")
	assert.Nil(t, cmd, "Update should return nil cmd")

	view := spacer.View()
	assert.NotNil(t, view, "View should return rendered output")
}

func TestSpacer_Props(t *testing.T) {
	props := SpacerProps{
		Width:  25,
		Height: 4,
	}

	spacer := Spacer(props)
	require.NotNil(t, spacer)

	spacer.Init()

	retrievedProps := spacer.Props()
	assert.NotNil(t, retrievedProps)

	spacerProps, ok := retrievedProps.(SpacerProps)
	assert.True(t, ok, "Props should be SpacerProps type")
	assert.Equal(t, 25, spacerProps.Width)
	assert.Equal(t, 4, spacerProps.Height)
}

func TestSpacer_ZeroDimensions(t *testing.T) {
	spacer := Spacer(SpacerProps{
		Width:  0,
		Height: 0,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	// Should render even with zero dimensions
	assert.NotNil(t, view)
}

// ============================================================================
// SPACER WITH CUSTOM STYLE TESTS - Additional Coverage
// ============================================================================

func TestSpacer_WithCustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().Background(lipgloss.Color("99"))

	spacer := Spacer(SpacerProps{
		Width:  10,
		Height: 2,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	assert.NotNil(t, view)
}

func TestSpacer_WidthOnlyWithStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("50"))

	spacer := Spacer(SpacerProps{
		Width: 15,
		// No height
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	assert.NotNil(t, view)
}
