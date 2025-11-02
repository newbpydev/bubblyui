package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TextProps defines the configuration properties for a Text component.
//
// Example usage:
//
//	text := components.Text(components.TextProps{
//	    Content:   "Hello, World!",
//	    Bold:      true,
//	    Color:     lipgloss.Color("63"),
//	    Alignment: components.AlignCenter,
//	})
type TextProps struct {
	// Content is the text content to display.
	// Required - the main text to render.
	Content string

	// Bold applies bold formatting to the text.
	// Default: false.
	Bold bool

	// Italic applies italic formatting to the text.
	// Default: false.
	Italic bool

	// Underline applies underline formatting to the text.
	// Default: false.
	Underline bool

	// Strikethrough applies strikethrough formatting to the text.
	// Default: false.
	Strikethrough bool

	// Color sets the foreground color of the text.
	// Supports ANSI colors, 256-color palette, and true color.
	// Optional - if not specified, uses theme foreground color.
	Color lipgloss.Color

	// Background sets the background color of the text.
	// Optional - if not specified, uses transparent background.
	Background lipgloss.Color

	// Alignment sets the horizontal alignment of the text.
	// Valid values: AlignLeft, AlignCenter, AlignRight.
	// Only applies when Width is set.
	// Default: AlignLeft.
	Alignment Alignment

	// Width sets the width constraint for the text.
	// When set, text will be aligned according to Alignment.
	// Optional - if 0, text uses natural width.
	Width int

	// Height sets the height constraint for the text.
	// Optional - if 0, text uses natural height.
	Height int

	// Common props for all components
	CommonProps
}

// Text creates a new Text atom component.
//
// Text is a fundamental display element for rendering styled text in the terminal.
// It supports various formatting options including bold, italic, underline, strikethrough,
// colors, and alignment.
//
// The text component automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	text := components.Text(components.TextProps{
//	    Content:   "Welcome to BubblyUI!",
//	    Bold:      true,
//	    Color:     lipgloss.Color("99"),
//	    Alignment: components.AlignCenter,
//	    Width:     40,
//	})
//
//	// Initialize and use with Bubbletea
//	text.Init()
//	view := text.View()
//
// Formatting options:
//   - Bold: Makes text bold
//   - Italic: Makes text italic
//   - Underline: Underlines text
//   - Strikethrough: Strikes through text
//   - Color: Sets text color
//   - Background: Sets background color
//   - Alignment: Aligns text (requires Width)
//   - Width/Height: Constrains text dimensions
//
// Accessibility:
//   - Clear visual distinction with formatting options
//   - High contrast colors available via theme
//   - Supports all terminal color profiles
func Text(props TextProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Text").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Try to inject theme, fallback to DefaultTheme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}

			// Expose theme for template
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(TextProps)
			theme := ctx.Get("theme").(Theme)

			// Build text style based on props
			textStyle := lipgloss.NewStyle()

			// Apply formatting options
			if props.Bold {
				textStyle = textStyle.Bold(true)
			}

			if props.Italic {
				textStyle = textStyle.Italic(true)
			}

			if props.Underline {
				textStyle = textStyle.Underline(true)
			}

			if props.Strikethrough {
				textStyle = textStyle.Strikethrough(true)
			}

			// Apply colors
			if props.Color != "" {
				textStyle = textStyle.Foreground(props.Color)
			} else {
				// Use theme foreground if no color specified
				textStyle = textStyle.Foreground(theme.Foreground)
			}

			if props.Background != "" {
				textStyle = textStyle.Background(props.Background)
			}

			// Apply width and alignment
			if props.Width > 0 {
				textStyle = textStyle.Width(props.Width)

				// Apply alignment based on props
				switch props.Alignment {
				case AlignLeft:
					textStyle = textStyle.Align(lipgloss.Left)
				case AlignCenter:
					textStyle = textStyle.Align(lipgloss.Center)
				case AlignRight:
					textStyle = textStyle.Align(lipgloss.Right)
				default:
					// Default to left alignment
					textStyle = textStyle.Align(lipgloss.Left)
				}
			}

			// Apply height if specified
			if props.Height > 0 {
				textStyle = textStyle.Height(props.Height)
			}

			// Apply custom style if provided
			if props.Style != nil {
				// Custom style overrides
				textStyle = textStyle.Inherit(*props.Style)
			}

			// Render text with style
			return textStyle.Render(props.Content)
		}).
		Build()

	return component
}
