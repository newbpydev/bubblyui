package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// IconProps defines the configuration properties for an Icon component.
//
// Example usage:
//
//	icon := components.Icon(components.IconProps{
//	    Symbol: "✓",
//	    Color:  lipgloss.Color("46"),
//	    Size:   components.SizeMedium,
//	})
type IconProps struct {
	// Symbol is the icon character or glyph to display.
	// Can be any Unicode character, emoji, or symbol.
	// Required - the main icon to render.
	Symbol string

	// Color sets the foreground color of the icon.
	// Supports ANSI colors, 256-color palette, and true color.
	// Optional - if not specified, uses theme foreground color.
	Color lipgloss.Color

	// Size determines the visual size of the icon.
	// Valid values: SizeSmall, SizeMedium, SizeLarge.
	// Optional - if not specified, uses natural size.
	Size Size

	// Common props for all components
	CommonProps
}

// Icon creates a new Icon atom component.
//
// Icon is a fundamental display element for rendering symbolic glyphs and indicators
// in the terminal. It supports Unicode characters, emojis, and special symbols with
// customizable colors and sizes.
//
// The icon component automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	icon := components.Icon(components.IconProps{
//	    Symbol: "⚠",
//	    Color:  lipgloss.Color("226"),
//	    Size:   components.SizeLarge,
//	})
//
//	// Initialize and use with Bubbletea
//	icon.Init()
//	view := icon.View()
//
// Common icon symbols:
//   - Checkmark: ✓
//   - Cross: ✗
//   - Warning: ⚠
//   - Info: ℹ
//   - Star: ★
//   - Heart: ♥
//   - Arrows: → ← ↑ ↓
//   - Shapes: ● ■ ◆
//
// Accessibility:
//   - Clear visual distinction with colors
//   - High contrast colors available via theme
//   - Supports all terminal color profiles
func Icon(props IconProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Icon").
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
			props := ctx.Props().(IconProps)
			theme := ctx.Get("theme").(Theme)

			// Build icon style
			iconStyle := lipgloss.NewStyle()

			// Apply color
			if props.Color != "" {
				iconStyle = iconStyle.Foreground(props.Color)
			} else {
				// Use theme foreground if no color specified
				iconStyle = iconStyle.Foreground(theme.Foreground)
			}

			// Apply size-based styling
			switch props.Size {
			case SizeSmall:
				// Small icons might have reduced padding
				iconStyle = iconStyle.Padding(0)
			case SizeLarge:
				// Large icons might have increased padding
				iconStyle = iconStyle.Padding(0, 1)
			default:
				// Medium or unspecified - default padding
				iconStyle = iconStyle.Padding(0)
			}

			// Apply custom style if provided
			if props.Style != nil {
				// Custom style overrides
				iconStyle = iconStyle.Inherit(*props.Style)
			}

			// Render icon with style
			return iconStyle.Render(props.Symbol)
		}).
		Build()

	return component
}
