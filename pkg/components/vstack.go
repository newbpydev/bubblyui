// Package components provides layout components for the BubblyUI framework.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// DefaultVStackDividerChar is the default divider character for VStack.
// Uses horizontal line since VStack renders vertically.
const DefaultVStackDividerChar = "─"

// vstackApplyDefaults sets default values for StackProps when used with VStack.
func vstackApplyDefaults(props *StackProps) {
	// DividerChar defaults to horizontal line for VStack
	if props.DividerChar == "" {
		props.DividerChar = DefaultVStackDividerChar
	}

	// Align defaults to start
	if props.Align == "" {
		props.Align = AlignItemsStart
	}
}

// vstackRenderItems renders all items and returns their string representations.
// This is the same as hstackRenderItems but kept separate for clarity.
func vstackRenderItems(items []interface{}) []string {
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

// vstackCalculateMaxWidth finds the maximum width among rendered items.
func vstackCalculateMaxWidth(items []string) int {
	maxWidth := 0
	for _, item := range items {
		width := lipgloss.Width(item)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

// vstackAlignItem aligns a single item horizontally based on AlignItems.
//
//nolint:dupl // Similar alignment logic is intentionally duplicated for clarity across layout components
func vstackAlignItem(item string, maxWidth int, align AlignItems) string {
	itemWidth := lipgloss.Width(item)

	if itemWidth >= maxWidth {
		return item
	}

	switch align {
	case AlignItemsCenter:
		return lipgloss.NewStyle().
			Width(maxWidth).
			Align(lipgloss.Center).
			Render(item)
	case AlignItemsEnd:
		return lipgloss.NewStyle().
			Width(maxWidth).
			Align(lipgloss.Right).
			Render(item)
	case AlignItemsStretch:
		// Stretch fills the width
		return lipgloss.NewStyle().
			Width(maxWidth).
			Render(item)
	default: // AlignItemsStart
		return lipgloss.NewStyle().
			Width(maxWidth).
			Align(lipgloss.Left).
			Render(item)
	}
}

// vstackCreateDivider creates a horizontal divider for VStack.
func vstackCreateDivider(char string, width int, theme Theme) string {
	dividerContent := strings.Repeat(char, width)

	style := lipgloss.NewStyle().
		Foreground(theme.Muted)

	return style.Render(dividerContent)
}

// vstackCreateSpacer creates vertical spacing (empty lines).
func vstackCreateSpacer(lines int) string {
	if lines <= 0 {
		return ""
	}
	// Create empty lines for spacing
	emptyLines := make([]string, lines)
	for i := range emptyLines {
		emptyLines[i] = ""
	}
	return strings.Join(emptyLines, "\n")
}

// vstackJoinWithSpacing joins items vertically with spacing and optional dividers.
func vstackJoinWithSpacing(items []string, spacing int, divider bool, dividerChar string, maxWidth int, theme Theme) string {
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
					spacer := vstackCreateSpacer(spacing / 2)
					if spacer != "" {
						parts = append(parts, spacer)
					}
				}
				parts = append(parts, vstackCreateDivider(dividerChar, maxWidth, theme))
				if spacing > 0 {
					spacer := vstackCreateSpacer(spacing - spacing/2)
					if spacer != "" {
						parts = append(parts, spacer)
					}
				}
			} else if spacing > 0 {
				spacer := vstackCreateSpacer(spacing)
				if spacer != "" {
					parts = append(parts, spacer)
				}
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// VStack creates a vertical stack layout component.
// Items are arranged vertically (top to bottom) with configurable spacing
// and cross-axis (horizontal) alignment.
//
// Features:
//   - Vertical arrangement of child components
//   - Configurable spacing between items (in lines)
//   - Cross-axis alignment (start/center/end/stretch)
//   - Optional dividers between items
//   - Theme integration for divider styling
//   - Custom style override support
//
// Example:
//
//	// Simple vertical stack
//	vstack := VStack(StackProps{
//	    Items: []interface{}{header, content, footer},
//	    Spacing: 1,
//	})
//
//	// With alignment and dividers
//	vstack := VStack(StackProps{
//	    Items:   []interface{}{title, description, actions},
//	    Align:   AlignItemsCenter,
//	    Divider: true,
//	})
//
//	// Form layout pattern
//	vstack := VStack(StackProps{
//	    Items: []interface{}{
//	        HStack(StackProps{Items: []interface{}{label1, input1}}),
//	        HStack(StackProps{Items: []interface{}{label2, input2}}),
//	        HStack(StackProps{Items: []interface{}{cancelBtn, submitBtn}}),
//	    },
//	    Spacing: 1,
//	})
//
//nolint:dupl // Component creation pattern is intentionally similar across all components
func VStack(props StackProps) bubbly.Component {
	vstackApplyDefaults(&props)

	component, _ := bubbly.NewComponent("VStack").
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
			rendered := vstackRenderItems(p.Items)
			if len(rendered) == 0 {
				return ""
			}

			// Calculate max width for alignment
			maxWidth := vstackCalculateMaxWidth(rendered)

			// Apply alignment to each item
			aligned := make([]string, len(rendered))
			for i, item := range rendered {
				aligned[i] = vstackAlignItem(item, maxWidth, p.Align)
			}

			// Join with spacing and dividers
			result := vstackJoinWithSpacing(aligned, p.Spacing, p.Divider, p.DividerChar, maxWidth, theme)

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
