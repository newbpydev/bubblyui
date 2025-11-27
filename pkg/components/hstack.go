// Package components provides layout components for the BubblyUI framework.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// DefaultStackSpacing is the default spacing between items in a stack.
const DefaultStackSpacing = 1

// DefaultHStackDividerChar is the default divider character for HStack.
// Uses vertical line since HStack renders horizontally.
const DefaultHStackDividerChar = "│"

// StackProps defines the properties for HStack and VStack components.
// Stack components provide simplified stacking layouts with spacing and alignment.
type StackProps struct {
	// Items are the child components to stack.
	// Each item can be any bubbly.Component.
	Items []interface{}

	// Spacing between items in characters (HStack) or lines (VStack).
	// Default: 1
	Spacing int

	// Align controls cross-axis alignment.
	// For HStack: vertical alignment (top/center/bottom)
	// For VStack: horizontal alignment (left/center/right)
	// Default: AlignItemsStart
	Align AlignItems

	// Divider optionally renders a divider between items.
	// Default: false
	Divider bool

	// DividerChar is the character for dividers.
	// Default: "│" for HStack, "─" for VStack
	DividerChar string

	// CommonProps for styling and identification.
	CommonProps
}

// hstackApplyDefaults sets default values for StackProps when used with HStack.
func hstackApplyDefaults(props *StackProps) {
	// Spacing defaults to 1 if not explicitly set to 0 or negative
	// We check if it's the zero value and apply default
	// Note: 0 spacing is valid, so we only default when props is freshly created
	// The design spec says default is 1

	// DividerChar defaults to vertical line for HStack
	if props.DividerChar == "" {
		props.DividerChar = DefaultHStackDividerChar
	}

	// Align defaults to start
	if props.Align == "" {
		props.Align = AlignItemsStart
	}
}

// hstackRenderItems renders all items and returns their string representations.
func hstackRenderItems(items []interface{}) []string {
	rendered := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if comp, ok := item.(bubbly.Component); ok {
			rendered = append(rendered, comp.View())
		}
	}
	return rendered
}

// hstackCalculateMaxHeight finds the maximum height among rendered items.
func hstackCalculateMaxHeight(items []string) int {
	maxHeight := 0
	for _, item := range items {
		height := lipgloss.Height(item)
		if height > maxHeight {
			maxHeight = height
		}
	}
	return maxHeight
}

// hstackAlignItem aligns a single item vertically based on AlignItems.
func hstackAlignItem(item string, maxHeight int, align AlignItems) string {
	itemHeight := lipgloss.Height(item)
	itemWidth := lipgloss.Width(item)

	if itemHeight >= maxHeight {
		return item
	}

	// Calculate padding needed
	diff := maxHeight - itemHeight

	switch align {
	case AlignItemsCenter:
		topPad := diff / 2
		bottomPad := diff - topPad
		return lipgloss.NewStyle().
			PaddingTop(topPad).
			PaddingBottom(bottomPad).
			Width(itemWidth).
			Render(item)
	case AlignItemsEnd:
		return lipgloss.NewStyle().
			PaddingTop(diff).
			Width(itemWidth).
			Render(item)
	case AlignItemsStretch:
		// Stretch fills the height
		return lipgloss.NewStyle().
			Height(maxHeight).
			Width(itemWidth).
			Render(item)
	default: // AlignItemsStart
		return lipgloss.NewStyle().
			PaddingBottom(diff).
			Width(itemWidth).
			Render(item)
	}
}

// hstackCreateDivider creates a vertical divider for HStack.
func hstackCreateDivider(char string, height int, theme Theme) string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = char
	}
	dividerContent := strings.Join(lines, "\n")

	style := lipgloss.NewStyle().
		Foreground(theme.Muted)

	return style.Render(dividerContent)
}

// hstackCreateSpacer creates horizontal spacing.
func hstackCreateSpacer(width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat(" ", width)
}

// hstackJoinWithSpacing joins items horizontally with spacing and optional dividers.
func hstackJoinWithSpacing(items []string, spacing int, divider bool, dividerChar string, maxHeight int, theme Theme) string {
	if len(items) == 0 {
		return ""
	}

	if len(items) == 1 {
		return items[0]
	}

	var parts []string
	for i, item := range items {
		parts = append(parts, item)

		// Add spacing/divider between items (not after last)
		if i < len(items)-1 {
			if divider {
				// Add spacing, divider, spacing
				if spacing > 0 {
					parts = append(parts, hstackCreateSpacer(spacing/2))
				}
				parts = append(parts, hstackCreateDivider(dividerChar, maxHeight, theme))
				if spacing > 0 {
					parts = append(parts, hstackCreateSpacer(spacing-spacing/2))
				}
			} else if spacing > 0 {
				parts = append(parts, hstackCreateSpacer(spacing))
			}
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// HStack creates a horizontal stack layout component.
// Items are arranged horizontally (left to right) with configurable spacing
// and cross-axis (vertical) alignment.
//
// Features:
//   - Horizontal arrangement of child components
//   - Configurable spacing between items
//   - Cross-axis alignment (start/center/end/stretch)
//   - Optional dividers between items
//   - Theme integration for divider styling
//   - Custom style override support
//
// Example:
//
//	// Simple horizontal stack
//	hstack := HStack(StackProps{
//	    Items: []interface{}{button1, button2, button3},
//	    Spacing: 2,
//	})
//
//	// With alignment and dividers
//	hstack := HStack(StackProps{
//	    Items:   []interface{}{logo, spacer, menuItems},
//	    Align:   AlignItemsCenter,
//	    Divider: true,
//	})
//
//	// Toolbar pattern with flexible spacer
//	hstack := HStack(StackProps{
//	    Items: []interface{}{
//	        Text(TextProps{Content: "Title"}),
//	        Spacer(SpacerProps{Flex: true}),
//	        Button(ButtonProps{Label: "Action"}),
//	    },
//	})
//
//nolint:dupl // Component creation pattern is intentionally similar across all components
func HStack(props StackProps) bubbly.Component {
	hstackApplyDefaults(&props)

	component, _ := bubbly.NewComponent("HStack").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(StackProps)
			theme := ctx.Get("theme").(Theme)

			// Handle empty or nil items
			if len(p.Items) == 0 {
				return ""
			}

			// Render all items
			rendered := hstackRenderItems(p.Items)
			if len(rendered) == 0 {
				return ""
			}

			// Calculate max height for alignment
			maxHeight := hstackCalculateMaxHeight(rendered)

			// Apply alignment to each item
			aligned := make([]string, len(rendered))
			for i, item := range rendered {
				aligned[i] = hstackAlignItem(item, maxHeight, p.Align)
			}

			// Join with spacing and dividers
			result := hstackJoinWithSpacing(aligned, p.Spacing, p.Divider, p.DividerChar, maxHeight, theme)

			// Apply custom style if provided
			if p.Style != nil {
				style := lipgloss.NewStyle().Inherit(*p.Style)
				return style.Render(result)
			}

			return result
		}).
		Build()

	return component
}
