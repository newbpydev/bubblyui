package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// BadgeProps defines the configuration properties for a Badge component.
//
// Example usage:
//
//	badge := components.Badge(components.BadgeProps{
//	    Label:   "New",
//	    Variant: components.VariantSuccess,
//	})
type BadgeProps struct {
	// Label is the text content to display in the badge.
	// Required - the main text to show.
	Label string

	// Variant determines the visual style of the badge.
	// Valid values: VariantPrimary, VariantSecondary, VariantSuccess, VariantWarning, VariantDanger, VariantInfo.
	// Optional - defaults to VariantPrimary if not specified.
	Variant Variant

	// Color sets a custom foreground color for the badge.
	// Overrides the variant color if specified.
	// Optional - if not specified, uses variant color from theme.
	Color lipgloss.Color

	// Common props for all components
	CommonProps
}

// Badge creates a new Badge atom component.
//
// Badge is a small status indicator component that displays short text labels
// with colored backgrounds. Commonly used for status indicators, counts, labels,
// and notifications.
//
// The badge component automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	badge := components.Badge(components.BadgeProps{
//	    Label:   "Active",
//	    Variant: components.VariantSuccess,
//	})
//
//	// Initialize and use with Bubbletea
//	badge.Init()
//	view := badge.View()
//
// Common use cases:
//   - Status indicators (Active, Inactive, Pending)
//   - Notification counts (5 new messages)
//   - Category labels (Bug, Feature, Documentation)
//   - Priority markers (High, Medium, Low)
//
// Accessibility:
//   - Clear visual distinction with variant colors
//   - High contrast for readability
//   - Compact design for inline use
func Badge(props BadgeProps) bubbly.Component {
	// Default to primary variant if not specified
	if props.Variant == "" {
		props.Variant = VariantPrimary
	}

	component, _ := bubbly.NewComponent("Badge").
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
			props := ctx.Props().(BadgeProps)
			theme := ctx.Get("theme").(Theme)

			// Build badge style
			badgeStyle := lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true)

			// Determine color to use
			var bgColor lipgloss.Color
			if props.Color != "" {
				// Use custom color if provided
				bgColor = props.Color
			} else {
				// Use variant color from theme
				bgColor = theme.GetVariantColor(props.Variant)
			}

			// Apply colors
			badgeStyle = badgeStyle.
				Background(bgColor).
				Foreground(lipgloss.Color("230")) // Light text for contrast

			// Apply custom style if provided
			if props.Style != nil {
				// Custom style overrides
				badgeStyle = badgeStyle.Inherit(*props.Style)
			}

			// Render badge with label
			return badgeStyle.Render(props.Label)
		}).
		Build()

	return component
}
