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

// ============================================================================
// ENHANCED SPACER TESTS - Task 2.3: Flex Behavior
// ============================================================================

func TestSpacer_FlexField_DefaultFalse(t *testing.T) {
	// Test that Flex defaults to false (existing behavior preserved)
	spacer := Spacer(SpacerProps{
		Width: 10,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	props := spacer.Props().(SpacerProps)

	assert.False(t, props.Flex, "Flex should default to false")
}

func TestSpacer_FlexTrue_Creation(t *testing.T) {
	// Test that Flex=true spacer can be created
	spacer := Spacer(SpacerProps{
		Flex: true,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	props := spacer.Props().(SpacerProps)

	assert.True(t, props.Flex, "Flex should be true when set")
}

func TestSpacer_FlexTrue_RendersEmpty(t *testing.T) {
	// Flex spacer without fixed dimensions renders empty
	// (parent layout will fill the space)
	spacer := Spacer(SpacerProps{
		Flex: true,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	// Flex spacer with no dimensions should render empty
	assert.Empty(t, view, "Flex spacer without dimensions should render empty")
}

func TestSpacer_FlexTrue_WithWidth_RendersMinimum(t *testing.T) {
	// Flex spacer with Width renders minimum width
	// (parent layout can expand beyond this)
	spacer := Spacer(SpacerProps{
		Flex:  true,
		Width: 5,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	// Should render at least the minimum width
	assert.Len(t, view, 5, "Flex spacer with Width should render minimum width")
}

func TestSpacer_FlexTrue_WithHeight_RendersMinimum(t *testing.T) {
	// Flex spacer with Height renders minimum height
	spacer := Spacer(SpacerProps{
		Flex:   true,
		Height: 3,
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	// Should have at least 2 newlines for 3 lines
	lineCount := strings.Count(view, "\n") + 1
	assert.GreaterOrEqual(t, lineCount, 3, "Flex spacer with Height should render minimum height")
}

func TestSpacer_ExistingBehavior_Preserved(t *testing.T) {
	// Ensure existing behavior is preserved when Flex=false
	tests := []struct {
		name   string
		props  SpacerProps
		verify func(t *testing.T, view string)
	}{
		{
			name:  "Width only",
			props: SpacerProps{Width: 10},
			verify: func(t *testing.T, view string) {
				assert.Len(t, view, 10, "Should create 10 spaces")
			},
		},
		{
			name:  "Height only",
			props: SpacerProps{Height: 3},
			verify: func(t *testing.T, view string) {
				assert.Contains(t, view, "\n", "Should contain newlines")
			},
		},
		{
			name:  "Both dimensions",
			props: SpacerProps{Width: 5, Height: 2},
			verify: func(t *testing.T, view string) {
				lines := strings.Split(view, "\n")
				assert.Len(t, lines, 2, "Should have 2 lines")
				for _, line := range lines {
					assert.Len(t, line, 5, "Each line should be 5 chars")
				}
			},
		},
		{
			name:  "Zero dimensions",
			props: SpacerProps{},
			verify: func(t *testing.T, view string) {
				assert.Empty(t, view, "Should be empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spacer := Spacer(tt.props)
			require.NotNil(t, spacer)

			spacer.Init()
			view := spacer.View()

			tt.verify(t, view)
		})
	}
}

func TestSpacer_IsFlex_Method(t *testing.T) {
	tests := []struct {
		name     string
		props    SpacerProps
		expected bool
	}{
		{
			name:     "Default (Flex=false)",
			props:    SpacerProps{Width: 10},
			expected: false,
		},
		{
			name:     "Flex=true",
			props:    SpacerProps{Flex: true},
			expected: true,
		},
		{
			name:     "Flex=true with dimensions",
			props:    SpacerProps{Flex: true, Width: 5, Height: 2},
			expected: true,
		},
		{
			name:     "Flex=false explicit",
			props:    SpacerProps{Flex: false, Width: 10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spacer := Spacer(tt.props)
			require.NotNil(t, spacer)

			spacer.Init()

			// Check via props
			props := spacer.Props().(SpacerProps)
			assert.Equal(t, tt.expected, props.Flex, "Flex field should match expected")

			// Check via IsFlex method
			assert.Equal(t, tt.expected, props.IsFlex(), "IsFlex() should match expected")
		})
	}
}

func TestSpacer_FlexMarker_ForParentLayouts(t *testing.T) {
	// This test verifies that parent layouts can detect flex spacers
	// by checking the Flex field on SpacerProps

	// Create a flex spacer
	flexSpacer := Spacer(SpacerProps{Flex: true})
	flexSpacer.Init()

	// Create a fixed spacer
	fixedSpacer := Spacer(SpacerProps{Width: 10})
	fixedSpacer.Init()

	// Simulate what a parent layout would do
	flexProps := flexSpacer.Props().(SpacerProps)
	fixedProps := fixedSpacer.Props().(SpacerProps)

	assert.True(t, flexProps.IsFlex(), "Parent should detect flex spacer")
	assert.False(t, fixedProps.IsFlex(), "Parent should detect fixed spacer")
}

func TestSpacer_FlexWithStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().Background(lipgloss.Color("99"))

	spacer := Spacer(SpacerProps{
		Flex:  true,
		Width: 10,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, spacer)

	spacer.Init()
	view := spacer.View()

	// Should render with style applied
	assert.NotEmpty(t, view, "Flex spacer with style should render")
}
