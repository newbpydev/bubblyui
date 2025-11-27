package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// mockCenterChild creates a simple mock component for testing Center.
func mockCenterChild(content string) bubbly.Component {
	comp, _ := bubbly.NewComponent("MockChild").
		Template(func(ctx bubbly.RenderContext) string {
			return content
		}).
		Build()
	return comp
}

// TestCenter_BothDirections_Default tests that Center centers both horizontally
// and vertically by default when neither Horizontal nor Vertical flags are set.
func TestCenter_BothDirections_Default(t *testing.T) {
	child := mockCenterChild("X")

	center := Center(CenterProps{
		Child:  child,
		Width:  10,
		Height: 5,
	})

	center.Init()
	output := center.View()

	// Content should be centered in a 10x5 space
	lines := strings.Split(output, "\n")
	require.Equal(t, 5, len(lines), "should have 5 lines for height=5")

	// The "X" should be roughly in the middle
	// Middle line (index 2) should contain "X"
	middleLine := lines[2]
	assert.Contains(t, middleLine, "X", "middle line should contain the content")

	// Check horizontal centering - "X" should not be at the start
	xIndex := strings.Index(middleLine, "X")
	assert.Greater(t, xIndex, 0, "X should be horizontally centered, not at start")
}

// TestCenter_HorizontalOnly tests that Center only centers horizontally
// when Horizontal=true and Vertical=false.
func TestCenter_HorizontalOnly(t *testing.T) {
	child := mockCenterChild("X")

	center := Center(CenterProps{
		Child:      child,
		Width:      20,
		Height:     0, // Auto height
		Horizontal: true,
		Vertical:   false,
	})

	center.Init()
	output := center.View()

	// Should be horizontally centered in 20-char width
	assert.Equal(t, 20, lipgloss.Width(output), "width should be 20")

	// Content should be centered horizontally
	xIndex := strings.Index(output, "X")
	assert.Greater(t, xIndex, 0, "X should be horizontally centered")
}

// TestCenter_VerticalOnly tests that Center only centers vertically
// when Horizontal=false and Vertical=true.
func TestCenter_VerticalOnly(t *testing.T) {
	child := mockCenterChild("X")

	center := Center(CenterProps{
		Child:      child,
		Width:      0, // Auto width
		Height:     5,
		Horizontal: false,
		Vertical:   true,
	})

	center.Init()
	output := center.View()

	lines := strings.Split(output, "\n")
	require.Equal(t, 5, len(lines), "should have 5 lines for height=5")

	// The "X" should be in the middle line
	middleLine := lines[2]
	assert.Contains(t, middleLine, "X", "middle line should contain the content")
}

// TestCenter_WithDimensions tests that Center uses provided Width and Height.
func TestCenter_WithDimensions(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "10x5 container",
			width:          10,
			height:         5,
			expectedWidth:  10,
			expectedHeight: 5,
		},
		{
			name:           "20x10 container",
			width:          20,
			height:         10,
			expectedWidth:  20,
			expectedHeight: 10,
		},
		{
			name:           "large container",
			width:          40,
			height:         20,
			expectedWidth:  40,
			expectedHeight: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := mockCenterChild("Content")

			center := Center(CenterProps{
				Child:  child,
				Width:  tt.width,
				Height: tt.height,
			})

			center.Init()
			output := center.View()

			assert.Equal(t, tt.expectedWidth, lipgloss.Width(output), "width mismatch")
			assert.Equal(t, tt.expectedHeight, lipgloss.Height(output), "height mismatch")
		})
	}
}

// TestCenter_AutoSize tests that Center auto-sizes when dimensions are 0.
func TestCenter_AutoSize(t *testing.T) {
	tests := []struct {
		name    string
		width   int
		height  int
		content string
	}{
		{
			name:    "auto width",
			width:   0,
			height:  5,
			content: "Hello",
		},
		{
			name:    "auto height",
			width:   20,
			height:  0,
			content: "World",
		},
		{
			name:    "auto both",
			width:   0,
			height:  0,
			content: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := mockCenterChild(tt.content)

			center := Center(CenterProps{
				Child:  child,
				Width:  tt.width,
				Height: tt.height,
			})

			center.Init()
			output := center.View()

			// Output should contain the content
			assert.Contains(t, output, tt.content, "output should contain content")

			// When auto-sizing, dimensions should match content or specified value
			if tt.width > 0 {
				assert.Equal(t, tt.width, lipgloss.Width(output), "width should match specified")
			}
			if tt.height > 0 {
				assert.Equal(t, tt.height, lipgloss.Height(output), "height should match specified")
			}
		})
	}
}

// TestCenter_NilChild tests that Center handles nil child gracefully.
func TestCenter_NilChild(t *testing.T) {
	center := Center(CenterProps{
		Child:  nil,
		Width:  10,
		Height: 5,
	})

	center.Init()
	output := center.View()

	// Should return empty or whitespace-only output
	trimmed := strings.TrimSpace(output)
	assert.Empty(t, trimmed, "nil child should produce empty content")
}

// TestCenter_ThemeIntegration tests that Center uses theme from context.
func TestCenter_ThemeIntegration(t *testing.T) {
	child := mockCenterChild("Themed")

	// Create a parent that provides a custom theme
	customTheme := Theme{
		Primary:    lipgloss.Color("99"),
		Secondary:  lipgloss.Color("88"),
		Background: lipgloss.Color("236"),
		Foreground: lipgloss.Color("255"),
		Muted:      lipgloss.Color("240"),
	}

	parent, _ := bubbly.NewComponent("Parent").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", customTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			center := Center(CenterProps{
				Child:  child,
				Width:  20,
				Height: 5,
			})
			center.Init()
			return center.View()
		}).
		Build()

	parent.Init()
	output := parent.View()

	// Should render without errors and contain content
	assert.Contains(t, output, "Themed", "should render child content")
}

// TestCenter_CustomStyle tests that Center applies custom style.
func TestCenter_CustomStyle(t *testing.T) {
	child := mockCenterChild("Styled")

	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236"))

	center := Center(CenterProps{
		Child:  child,
		Width:  20,
		Height: 5,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	center.Init()
	output := center.View()

	// Should render with custom style applied
	assert.Contains(t, output, "Styled", "should render child content")
	assert.Equal(t, 20, lipgloss.Width(output), "width should be 20")
}

// TestCenter_EmptyItems tests edge cases with empty content.
func TestCenter_EmptyContent(t *testing.T) {
	child := mockCenterChild("")

	center := Center(CenterProps{
		Child:  child,
		Width:  10,
		Height: 5,
	})

	center.Init()
	output := center.View()

	// Should still create the container with correct dimensions
	assert.Equal(t, 10, lipgloss.Width(output), "width should be 10")
	assert.Equal(t, 5, lipgloss.Height(output), "height should be 5")
}

// TestCenter_MultilineContent tests centering of multi-line content.
func TestCenter_MultilineContent(t *testing.T) {
	child := mockCenterChild("Line1\nLine2\nLine3")

	center := Center(CenterProps{
		Child:  child,
		Width:  20,
		Height: 10,
	})

	center.Init()
	output := center.View()

	// Should contain all lines
	assert.Contains(t, output, "Line1", "should contain Line1")
	assert.Contains(t, output, "Line2", "should contain Line2")
	assert.Contains(t, output, "Line3", "should contain Line3")

	// Should have correct dimensions
	assert.Equal(t, 20, lipgloss.Width(output), "width should be 20")
	assert.Equal(t, 10, lipgloss.Height(output), "height should be 10")
}

// TestCenter_BothFlagsTrue tests when both Horizontal and Vertical are true.
func TestCenter_BothFlagsTrue(t *testing.T) {
	child := mockCenterChild("X")

	center := Center(CenterProps{
		Child:      child,
		Width:      10,
		Height:     5,
		Horizontal: true,
		Vertical:   true,
	})

	center.Init()
	output := center.View()

	lines := strings.Split(output, "\n")
	require.Equal(t, 5, len(lines), "should have 5 lines")

	// Content should be centered both ways
	middleLine := lines[2]
	assert.Contains(t, middleLine, "X", "middle line should contain X")

	xIndex := strings.Index(middleLine, "X")
	assert.Greater(t, xIndex, 0, "X should be horizontally centered")
}

// TestCenter_LargeContent tests when content is larger than container.
func TestCenter_LargeContent(t *testing.T) {
	// Content wider than container
	child := mockCenterChild("This is a very long content string")

	center := Center(CenterProps{
		Child:  child,
		Width:  10,
		Height: 3,
	})

	center.Init()
	output := center.View()

	// Should still render (content may overflow)
	assert.NotEmpty(t, output, "should produce output")
}

// TestCenter_ComponentName tests that Center has correct component name.
func TestCenter_ComponentName(t *testing.T) {
	center := Center(CenterProps{
		Child:  mockCenterChild("Test"),
		Width:  10,
		Height: 5,
	})

	assert.Equal(t, "Center", center.Name(), "component name should be 'Center'")
}
