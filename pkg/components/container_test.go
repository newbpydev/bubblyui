package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// mockContainerChild creates a simple mock component for testing Container.
func mockContainerChild(content string) bubbly.Component {
	comp, _ := bubbly.NewComponent("MockChild").
		Template(func(ctx bubbly.RenderContext) string {
			return content
		}).
		Build()
	return comp
}

// TestContainer_PresetSizes tests that Container constrains width to preset sizes.
func TestContainer_PresetSizes(t *testing.T) {
	tests := []struct {
		name          string
		size          ContainerSize
		expectedWidth int
	}{
		{
			name:          "small container (40 chars)",
			size:          ContainerSm,
			expectedWidth: 40,
		},
		{
			name:          "medium container (60 chars)",
			size:          ContainerMd,
			expectedWidth: 60,
		},
		{
			name:          "large container (80 chars)",
			size:          ContainerLg,
			expectedWidth: 80,
		},
		{
			name:          "extra-large container (100 chars)",
			size:          ContainerXl,
			expectedWidth: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := mockContainerChild("Content")

			container := Container(ContainerProps{
				Child:       child,
				Size:        tt.size,
				Centered:    false, // Disable centering to test width constraint only
				CenteredSet: true,
			})

			container.Init()
			output := container.View()

			assert.Equal(t, tt.expectedWidth, lipgloss.Width(output),
				"container width should match preset size")
		})
	}
}

// TestContainer_FullSize tests that ContainerFull uses content width (no constraint).
func TestContainer_FullSize(t *testing.T) {
	content := "Short"
	child := mockContainerChild(content)

	container := Container(ContainerProps{
		Child:       child,
		Size:        ContainerFull,
		Centered:    false,
		CenteredSet: true,
	})

	container.Init()
	output := container.View()

	// Full size should not constrain width - content determines width
	assert.Contains(t, output, content, "should contain content")
	// Width should be content width (no padding added)
	assert.Equal(t, len(content), lipgloss.Width(output),
		"full size should use content width")
}

// TestContainer_CustomMaxWidth tests that MaxWidth overrides Size preset.
func TestContainer_CustomMaxWidth(t *testing.T) {
	tests := []struct {
		name          string
		size          ContainerSize
		maxWidth      int
		expectedWidth int
	}{
		{
			name:          "MaxWidth overrides Size",
			size:          ContainerLg, // 80
			maxWidth:      50,
			expectedWidth: 50,
		},
		{
			name:          "MaxWidth larger than preset",
			size:          ContainerSm, // 40
			maxWidth:      70,
			expectedWidth: 70,
		},
		{
			name:          "MaxWidth zero uses Size",
			size:          ContainerMd, // 60
			maxWidth:      0,
			expectedWidth: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := mockContainerChild("Content")

			container := Container(ContainerProps{
				Child:       child,
				Size:        tt.size,
				MaxWidth:    tt.maxWidth,
				Centered:    false,
				CenteredSet: true,
			})

			container.Init()
			output := container.View()

			assert.Equal(t, tt.expectedWidth, lipgloss.Width(output),
				"container width should match expected")
		})
	}
}

// TestContainer_Centered tests that Container centers content when Centered=true.
func TestContainer_Centered(t *testing.T) {
	child := mockContainerChild("X")

	// Create a container with centering enabled
	// We need a parent width to see centering effect
	container := Container(ContainerProps{
		Child:    child,
		Size:     ContainerSm, // 40 chars
		Centered: true,
	})

	container.Init()
	output := container.View()

	// Content should be centered within the container width
	assert.Equal(t, 40, lipgloss.Width(output), "width should be 40")

	// The "X" should be centered (not at position 0)
	xIndex := strings.Index(output, "X")
	assert.Greater(t, xIndex, 0, "X should be horizontally centered, not at start")
}

// TestContainer_NotCentered tests that Container does not center when Centered=false.
func TestContainer_NotCentered(t *testing.T) {
	child := mockContainerChild("X")

	container := Container(ContainerProps{
		Child:       child,
		Size:        ContainerSm, // 40 chars
		Centered:    false,
		CenteredSet: true,
	})

	container.Init()
	output := container.View()

	// Content should be at the start (left-aligned)
	assert.Equal(t, 40, lipgloss.Width(output), "width should be 40")

	// The "X" should be at the start
	xIndex := strings.Index(output, "X")
	assert.Equal(t, 0, xIndex, "X should be at start when not centered")
}

// TestContainer_DefaultSize tests that Container uses ContainerMd by default.
func TestContainer_DefaultSize(t *testing.T) {
	child := mockContainerChild("Content")

	container := Container(ContainerProps{
		Child:       child,
		Centered:    false,
		CenteredSet: true,
		// Size not specified - should default to ContainerMd (60)
	})

	container.Init()
	output := container.View()

	assert.Equal(t, 60, lipgloss.Width(output),
		"default size should be ContainerMd (60 chars)")
}

// TestContainer_DefaultCentered tests that Container centers by default.
func TestContainer_DefaultCentered(t *testing.T) {
	child := mockContainerChild("X")

	container := Container(ContainerProps{
		Child: child,
		Size:  ContainerSm, // 40 chars
		// Centered not specified - should default to true
	})

	container.Init()
	output := container.View()

	// Content should be centered (default behavior)
	xIndex := strings.Index(output, "X")
	assert.Greater(t, xIndex, 0, "X should be centered by default")
}

// TestContainer_NilChild tests that Container handles nil child gracefully.
func TestContainer_NilChild(t *testing.T) {
	container := Container(ContainerProps{
		Child:       nil,
		Size:        ContainerSm,
		Centered:    false,
		CenteredSet: true,
	})

	container.Init()
	output := container.View()

	// Should return empty or whitespace-only output with correct width
	assert.Equal(t, 40, lipgloss.Width(output), "width should still be 40")
	trimmed := strings.TrimSpace(output)
	assert.Empty(t, trimmed, "nil child should produce empty content")
}

// TestContainer_ThemeIntegration tests that Container uses theme from context.
func TestContainer_ThemeIntegration(t *testing.T) {
	child := mockContainerChild("Themed")

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
			container := Container(ContainerProps{
				Child:       child,
				Size:        ContainerMd,
				Centered:    false,
				CenteredSet: true,
			})
			container.Init()
			return container.View()
		}).
		Build()

	parent.Init()
	output := parent.View()

	assert.Contains(t, output, "Themed", "should render child content")
}

// TestContainer_CustomStyle tests that Container applies custom style.
func TestContainer_CustomStyle(t *testing.T) {
	child := mockContainerChild("Styled")

	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236"))

	container := Container(ContainerProps{
		Child:       child,
		Size:        ContainerMd,
		Centered:    false,
		CenteredSet: true,
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	container.Init()
	output := container.View()

	assert.Contains(t, output, "Styled", "should render child content")
	assert.Equal(t, 60, lipgloss.Width(output), "width should be 60")
}

// TestContainer_MultilineContent tests Container with multi-line content.
func TestContainer_MultilineContent(t *testing.T) {
	child := mockContainerChild("Line1\nLine2\nLine3")

	container := Container(ContainerProps{
		Child:       child,
		Size:        ContainerMd,
		Centered:    false,
		CenteredSet: true,
	})

	container.Init()
	output := container.View()

	assert.Contains(t, output, "Line1", "should contain Line1")
	assert.Contains(t, output, "Line2", "should contain Line2")
	assert.Contains(t, output, "Line3", "should contain Line3")
	assert.Equal(t, 60, lipgloss.Width(output), "width should be 60")
}

// TestContainer_LongContent tests Container with content longer than max width.
func TestContainer_LongContent(t *testing.T) {
	// Content wider than container
	longContent := strings.Repeat("X", 100)
	child := mockContainerChild(longContent)

	container := Container(ContainerProps{
		Child:       child,
		Size:        ContainerSm, // 40 chars
		Centered:    false,
		CenteredSet: true,
	})

	container.Init()
	output := container.View()

	// Container should still have the specified width
	// Content may overflow or be truncated
	require.NotEmpty(t, output, "should produce output")
}

// TestContainer_ComponentName tests that Container has correct component name.
func TestContainer_ComponentName(t *testing.T) {
	container := Container(ContainerProps{
		Child: mockContainerChild("Test"),
		Size:  ContainerMd,
	})

	assert.Equal(t, "Container", container.Name(), "component name should be 'Container'")
}

// TestContainer_EmptyContent tests Container with empty content.
func TestContainer_EmptyContent(t *testing.T) {
	child := mockContainerChild("")

	container := Container(ContainerProps{
		Child:       child,
		Size:        ContainerSm,
		Centered:    false,
		CenteredSet: true,
	})

	container.Init()
	output := container.View()

	// Should still create the container with correct width
	assert.Equal(t, 40, lipgloss.Width(output), "width should be 40")
}

// TestContainer_CenteredWithMaxWidth tests centering with custom MaxWidth.
func TestContainer_CenteredWithMaxWidth(t *testing.T) {
	child := mockContainerChild("X")

	container := Container(ContainerProps{
		Child:    child,
		MaxWidth: 30,
		Centered: true,
	})

	container.Init()
	output := container.View()

	assert.Equal(t, 30, lipgloss.Width(output), "width should be 30")

	// Content should be centered
	xIndex := strings.Index(output, "X")
	assert.Greater(t, xIndex, 0, "X should be centered")
}
