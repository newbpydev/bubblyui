package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestText_Creation(t *testing.T) {
	tests := []struct {
		name  string
		props TextProps
	}{
		{
			name: "Simple text",
			props: TextProps{
				Content: "Hello, World!",
			},
		},
		{
			name: "Bold text",
			props: TextProps{
				Content: "Bold Text",
				Bold:    true,
			},
		},
		{
			name: "Italic text",
			props: TextProps{
				Content: "Italic Text",
				Italic:  true,
			},
		},
		{
			name: "Underline text",
			props: TextProps{
				Content:   "Underlined Text",
				Underline: true,
			},
		},
		{
			name: "Colored text",
			props: TextProps{
				Content: "Colored Text",
				Color:   lipgloss.Color("63"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(tt.props)
			require.NotNil(t, text, "Text should not be nil")

			// Initialize component
			text.Init()

			// Verify component can render
			view := text.View()
			assert.NotEmpty(t, view, "Text view should not be empty")
			assert.Contains(t, view, tt.props.Content, "Text should contain content")
		})
	}
}

func TestText_BoldFormatting(t *testing.T) {
	tests := []struct {
		name string
		bold bool
	}{
		{"Bold enabled", true},
		{"Bold disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content: "Test",
				Bold:    tt.bold,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestText_ItalicFormatting(t *testing.T) {
	tests := []struct {
		name   string
		italic bool
	}{
		{"Italic enabled", true},
		{"Italic disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content: "Test",
				Italic:  tt.italic,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestText_UnderlineFormatting(t *testing.T) {
	tests := []struct {
		name      string
		underline bool
	}{
		{"Underline enabled", true},
		{"Underline disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content:   "Test",
				Underline: tt.underline,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestText_StrikethroughFormatting(t *testing.T) {
	tests := []struct {
		name          string
		strikethrough bool
	}{
		{"Strikethrough enabled", true},
		{"Strikethrough disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content:       "Test",
				Strikethrough: tt.strikethrough,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestText_ColorFormatting(t *testing.T) {
	tests := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Red color", lipgloss.Color("196")},
		{"Blue color", lipgloss.Color("63")},
		{"Green color", lipgloss.Color("46")},
		{"No color", lipgloss.Color("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content: "Colored",
				Color:   tt.color,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Colored")
		})
	}
}

func TestText_CombinedFormatting(t *testing.T) {
	tests := []struct {
		name  string
		props TextProps
	}{
		{
			name: "Bold and Italic",
			props: TextProps{
				Content: "Bold Italic",
				Bold:    true,
				Italic:  true,
			},
		},
		{
			name: "Bold, Italic, and Underline",
			props: TextProps{
				Content:   "All Three",
				Bold:      true,
				Italic:    true,
				Underline: true,
			},
		},
		{
			name: "All formatting options",
			props: TextProps{
				Content:       "Everything",
				Bold:          true,
				Italic:        true,
				Underline:     true,
				Strikethrough: true,
				Color:         lipgloss.Color("99"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(tt.props)
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, tt.props.Content)
		})
	}
}

func TestText_Alignment(t *testing.T) {
	tests := []struct {
		name      string
		alignment Alignment
		width     int
	}{
		{"Left aligned", AlignLeft, 20},
		{"Center aligned", AlignCenter, 20},
		{"Right aligned", AlignRight, 20},
		{"No alignment", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content:   "Test",
				Alignment: tt.alignment,
				Width:     tt.width,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Test")
		})
	}
}

func TestText_EmptyContent(t *testing.T) {
	text := Text(TextProps{
		Content: "",
	})
	require.NotNil(t, text)

	text.Init()
	view := text.View()

	// Should render even with empty content
	assert.NotNil(t, view)
}

func TestText_LongContent(t *testing.T) {
	longContent := strings.Repeat("Very long text content ", 50)
	text := Text(TextProps{
		Content: longContent,
	})
	require.NotNil(t, text)

	text.Init()
	view := text.View()

	assert.NotEmpty(t, view)
	// Should handle long content without panic
}

func TestText_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		contains string // What to check for in output (may differ from input due to rendering)
	}{
		{"Unicode characters", "你好世界", "你好世界"},
		{"Emoji", "🎉 🚀 ✨", "🎉 🚀 ✨"},
		{"Special symbols", "© ® ™ € £ ¥", "© ® ™ € £ ¥"},
		{"Newlines", "Line 1\nLine 2\nLine 3", "Line 1"},
		{"Tabs", "Col1\tCol2\tCol3", "Col1"}, // Tabs are rendered as spaces in terminal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content: tt.content,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
			assert.Contains(t, view, tt.contains)
		})
	}
}

func TestText_BubbleteatIntegration(t *testing.T) {
	// Test that text works with Bubbletea Update cycle
	text := Text(TextProps{
		Content: "Test",
	})
	require.NotNil(t, text)

	// Init
	cmd := text.Init()
	assert.Nil(t, cmd, "Init should return nil cmd")

	// Update with key message
	updated, cmd := text.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated, "Update should return updated component")
	assert.Nil(t, cmd, "Update should return nil cmd for key messages")

	// View
	view := text.View()
	assert.NotEmpty(t, view, "View should return rendered output")
}

func TestText_Props(t *testing.T) {
	props := TextProps{
		Content:   "Test Content",
		Bold:      true,
		Italic:    true,
		Underline: true,
		Color:     lipgloss.Color("99"),
	}

	text := Text(props)
	require.NotNil(t, text)

	text.Init()

	// Verify props are accessible
	retrievedProps := text.Props()
	assert.NotNil(t, retrievedProps)

	textProps, ok := retrievedProps.(TextProps)
	assert.True(t, ok, "Props should be TextProps type")
	assert.Equal(t, "Test Content", textProps.Content)
	assert.True(t, textProps.Bold)
	assert.True(t, textProps.Italic)
	assert.True(t, textProps.Underline)
	assert.Equal(t, lipgloss.Color("99"), textProps.Color)
}

func TestText_ThemeIntegration(t *testing.T) {
	// Test that text integrates with theme system
	text := Text(TextProps{
		Content: "Themed Text",
		Color:   lipgloss.Color(""), // Empty color should use theme
	})
	require.NotNil(t, text)

	text.Init()
	view := text.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Themed Text")
}

func TestText_CustomStyle(t *testing.T) {
	// Test custom style override
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Padding(1, 2)

	text := Text(TextProps{
		Content: "Custom Styled",
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})
	require.NotNil(t, text)

	text.Init()
	view := text.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Custom Styled")
}

func TestText_Width(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"No width", 0},
		{"Small width", 10},
		{"Medium width", 40},
		{"Large width", 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content: "Test content for width",
				Width:   tt.width,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
		})
	}
}

func TestText_Height(t *testing.T) {
	tests := []struct {
		name   string
		height int
	}{
		{"No height", 0},
		{"Small height", 3},
		{"Medium height", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Text(TextProps{
				Content: "Test content\nwith multiple\nlines",
				Height:  tt.height,
			})
			require.NotNil(t, text)

			text.Init()
			view := text.View()

			assert.NotEmpty(t, view)
		})
	}
}
