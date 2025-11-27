package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestBox_Creation(t *testing.T) {
	box := Box(BoxProps{
		Content: "Test content",
	})

	assert.NotNil(t, box, "Box should be created")
}

func TestBox_RenderContentWithPadding(t *testing.T) {
	tests := []struct {
		name     string
		padding  int
		paddingX int
		paddingY int
		content  string
	}{
		{
			name:    "uniform padding",
			padding: 1,
			content: "Padded content",
		},
		{
			name:     "horizontal padding only",
			paddingX: 2,
			content:  "Horizontal padding",
		},
		{
			name:     "vertical padding only",
			paddingY: 1,
			content:  "Vertical padding",
		},
		{
			name:     "mixed padding",
			padding:  1,
			paddingX: 3,
			paddingY: 2,
			content:  "Mixed padding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := Box(BoxProps{
				Content:  tt.content,
				Padding:  tt.padding,
				PaddingX: tt.paddingX,
				PaddingY: tt.paddingY,
			})

			box.Init()
			output := box.View()

			assert.Contains(t, output, tt.content, "Should render content")
			assert.NotEmpty(t, output, "Should have output")
		})
	}
}

func TestBox_RenderBorderWhenEnabled(t *testing.T) {
	tests := []struct {
		name       string
		border     bool
		wantBorder bool
	}{
		{
			name:       "border enabled",
			border:     true,
			wantBorder: true,
		},
		{
			name:       "border disabled",
			border:     false,
			wantBorder: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := Box(BoxProps{
				Content: "Border test",
				Border:  tt.border,
			})

			box.Init()
			output := box.View()

			// Check for border characters (normal border uses │ and ─)
			hasBorder := strings.Contains(output, "│") || strings.Contains(output, "─") ||
				strings.Contains(output, "┌") || strings.Contains(output, "└")

			if tt.wantBorder {
				assert.True(t, hasBorder, "Should have border characters")
			} else {
				assert.False(t, hasBorder, "Should not have border characters")
			}
		})
	}
}

func TestBox_RenderTitleOnBorder(t *testing.T) {
	box := Box(BoxProps{
		Content: "Content",
		Border:  true,
		Title:   "My Title",
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "My Title", "Should render title")
	assert.Contains(t, output, "Content", "Should render content")
}

func TestBox_NilChildUsesContent(t *testing.T) {
	box := Box(BoxProps{
		Child:   nil,
		Content: "Fallback content",
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Fallback content", "Should use Content when Child is nil")
}

func TestBox_ChildComponentRendered(t *testing.T) {
	child := Text(TextProps{
		Content: "Child text component",
	})
	child.Init()

	box := Box(BoxProps{
		Child: child,
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Child text component", "Should render child component")
}

func TestBox_ThemeIntegration(t *testing.T) {
	box := Box(BoxProps{
		Content: "Themed content",
		Border:  true,
	})

	box.Init()
	output := box.View()

	assert.NotEmpty(t, output, "Should render with theme")
}

func TestBox_CustomBorderStyle(t *testing.T) {
	box := Box(BoxProps{
		Content:     "Custom border",
		Border:      true,
		BorderStyle: lipgloss.RoundedBorder(),
	})

	box.Init()
	output := box.View()

	// Rounded border uses ╭ ╮ ╯ ╰
	hasRoundedBorder := strings.Contains(output, "╭") || strings.Contains(output, "╮")
	assert.True(t, hasRoundedBorder, "Should use rounded border style")
}

func TestBox_PaddingXYOverridesPadding(t *testing.T) {
	// PaddingX and PaddingY should take precedence over Padding
	box := Box(BoxProps{
		Content:  "Override test",
		Padding:  1,
		PaddingX: 4,
		PaddingY: 2,
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Override test", "Should render content")
	assert.NotEmpty(t, output, "Should have output with overridden padding")
}

func TestBox_WidthHeight(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:  "fixed width",
			width: 30,
		},
		{
			name:   "fixed height",
			height: 10,
		},
		{
			name:   "fixed both",
			width:  40,
			height: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := Box(BoxProps{
				Content: "Sized box",
				Width:   tt.width,
				Height:  tt.height,
			})

			box.Init()
			output := box.View()

			assert.NotEmpty(t, output, "Should render with dimensions")
		})
	}
}

func TestBox_Background(t *testing.T) {
	box := Box(BoxProps{
		Content:    "Background test",
		Background: lipgloss.Color("99"),
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Background test", "Should render content with background")
}

func TestBox_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	box := Box(BoxProps{
		Content: "Custom styled",
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Custom styled", "Should render with custom style")
}

func TestBox_BubbleteatIntegration(t *testing.T) {
	box := Box(BoxProps{
		Content: "Integration test",
	})

	// Test Init
	cmd := box.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	newModel, cmd := box.Update(nil)
	assert.NotNil(t, newModel, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command for nil msg")

	// Test View
	output := box.View()
	assert.NotEmpty(t, output, "View should return output")
}

func TestBox_EmptyContent(t *testing.T) {
	box := Box(BoxProps{
		Content: "",
		Border:  true,
	})

	box.Init()
	output := box.View()

	// Should still render the border even with empty content
	hasBorder := strings.Contains(output, "│") || strings.Contains(output, "─")
	assert.True(t, hasBorder, "Should render border even with empty content")
}

func TestBox_ChildWithContent(t *testing.T) {
	// When both Child and Content are provided, Child takes precedence
	child := Text(TextProps{
		Content: "Child wins",
	})
	child.Init()

	box := Box(BoxProps{
		Child:   child,
		Content: "Content loses",
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Child wins", "Child should take precedence")
	assert.NotContains(t, output, "Content loses", "Content should not be rendered when Child exists")
}

func TestBox_TitleWithoutBorder(t *testing.T) {
	// Title should still render even without border
	box := Box(BoxProps{
		Content: "Content",
		Title:   "Title without border",
		Border:  false,
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Title without border", "Title should render even without border")
	assert.Contains(t, output, "Content", "Content should render")
}

func TestBox_AllPropsCombo(t *testing.T) {
	// Test with all props set
	child := Text(TextProps{
		Content: "Full combo child",
	})
	child.Init()

	customStyle := lipgloss.NewStyle().Bold(true)

	box := Box(BoxProps{
		Child:       child,
		Content:     "Ignored content",
		Padding:     1,
		PaddingX:    2,
		PaddingY:    1,
		Border:      true,
		BorderStyle: lipgloss.DoubleBorder(),
		Title:       "Full Combo",
		Width:       50,
		Height:      10,
		Background:  lipgloss.Color("236"),
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	box.Init()
	output := box.View()

	assert.Contains(t, output, "Full combo child", "Should render child")
	assert.Contains(t, output, "Full Combo", "Should render title")
	// Double border uses ║ and ═
	hasDoubleBorder := strings.Contains(output, "║") || strings.Contains(output, "═")
	assert.True(t, hasDoubleBorder, "Should use double border")
}
