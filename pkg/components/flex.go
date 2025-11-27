// Package components provides layout components for the BubblyUI framework.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// FlexProps defines properties for the Flex layout component.
// Flex provides flexbox-style layout with direction, justify, and align options.
//
// Example:
//
//	// Row layout with space-between
//	flex := Flex(FlexProps{
//	    Items:     []bubbly.Component{btn1, btn2, btn3},
//	    Direction: FlexRow,
//	    Justify:   JustifySpaceBetween,
//	    Gap:       2,
//	})
//
//	// Column layout centered
//	flex := Flex(FlexProps{
//	    Items:     []bubbly.Component{header, content, footer},
//	    Direction: FlexColumn,
//	    Justify:   JustifyCenter,
//	    Align:     AlignItemsCenter,
//	})
type FlexProps struct {
	// Items are the child components to arrange.
	Items []bubbly.Component

	// Direction specifies row (horizontal) or column (vertical).
	// Default: FlexRow
	Direction FlexDirection

	// Justify controls main-axis distribution.
	// Default: JustifyStart
	Justify JustifyContent

	// Align controls cross-axis alignment.
	// Default: AlignItemsStart
	Align AlignItems

	// Gap is the spacing between items in characters.
	// Default: 0
	Gap int

	// Wrap enables wrapping items to next row/column.
	// Default: false (implemented in Task 4.4)
	Wrap bool

	// Width sets fixed container width. 0 = auto.
	Width int

	// Height sets fixed container height. 0 = auto.
	Height int

	// CommonProps for styling and identification.
	CommonProps
}

// flexApplyDefaults sets default values for FlexProps.
func flexApplyDefaults(props *FlexProps) {
	if props.Direction == "" {
		props.Direction = FlexRow
	}
	if props.Justify == "" {
		props.Justify = JustifyStart
	}
	if props.Align == "" {
		props.Align = AlignItemsStart
	}
}

// flexRenderItems renders all items and returns their string representations.
func flexRenderItems(items []bubbly.Component) []string {
	rendered := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		rendered = append(rendered, item.View())
	}
	return rendered
}

// flexCalculateMaxHeight finds the maximum height among rendered items.
func flexCalculateMaxHeight(items []string) int {
	maxHeight := 0
	for _, item := range items {
		height := lipgloss.Height(item)
		if height > maxHeight {
			maxHeight = height
		}
	}
	return maxHeight
}

// flexCalculateMaxWidth finds the maximum width among rendered items.
func flexCalculateMaxWidth(items []string) int {
	maxWidth := 0
	for _, item := range items {
		width := lipgloss.Width(item)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

// flexCalculateTotalWidth calculates total width of all items.
func flexCalculateTotalWidth(items []string) int {
	total := 0
	for _, item := range items {
		total += lipgloss.Width(item)
	}
	return total
}

// flexCalculateTotalHeight calculates total height of all items.
func flexCalculateTotalHeight(items []string) int {
	total := 0
	for _, item := range items {
		total += lipgloss.Height(item)
	}
	return total
}

// flexAlignItemRow aligns a single item vertically (cross-axis for row direction).
//
//nolint:dupl // Similar alignment logic is intentionally duplicated for clarity across layout components
func flexAlignItemRow(item string, maxHeight int, align AlignItems) string {
	itemHeight := lipgloss.Height(item)
	itemWidth := lipgloss.Width(item)

	if itemHeight >= maxHeight {
		return item
	}

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

// flexAlignItemColumn aligns a single item horizontally (cross-axis for column direction).
//
//nolint:dupl // Similar alignment logic is intentionally duplicated for clarity across layout components
func flexAlignItemColumn(item string, maxWidth int, align AlignItems) string {
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

// flexGapResult holds the calculated gap distribution.
type flexGapResult struct {
	gaps         []int
	startPadding int
	endPadding   int
}

// flexCalculateGaps calculates gap sizes between items based on justify mode.
// Returns a slice of gap sizes (one less than number of items) and start padding.
func flexCalculateGaps(itemSizes []int, containerSize int, justify JustifyContent, gap int) ([]int, int) {
	n := len(itemSizes)
	if n == 0 {
		return nil, 0
	}

	totalItemSize := 0
	for _, size := range itemSizes {
		totalItemSize += size
	}

	// Calculate remaining space after items and explicit gaps
	totalGapSpace := gap * (n - 1)
	remainingSpace := containerSize - totalItemSize - totalGapSpace
	if remainingSpace < 0 {
		remainingSpace = 0
	}

	// Initialize gaps with explicit gap value
	gaps := make([]int, n-1)
	for i := range gaps {
		gaps[i] = gap
	}

	result := flexDistributeSpace(n, remainingSpace, gaps, justify)
	return result.gaps, result.startPadding
}

// flexDistributeSpace distributes remaining space based on justify mode.
func flexDistributeSpace(n, remainingSpace int, gaps []int, justify JustifyContent) flexGapResult {
	switch justify {
	case JustifyEnd:
		return flexGapResult{gaps: gaps, startPadding: remainingSpace, endPadding: 0}
	case JustifyCenter:
		half := remainingSpace / 2
		return flexGapResult{gaps: gaps, startPadding: half, endPadding: remainingSpace - half}
	case JustifySpaceBetween:
		return flexDistributeSpaceBetween(n, remainingSpace, gaps)
	case JustifySpaceAround:
		return flexDistributeSpaceAround(n, remainingSpace, gaps)
	case JustifySpaceEvenly:
		return flexDistributeSpaceEvenly(n, remainingSpace, gaps)
	default: // JustifyStart
		return flexGapResult{gaps: gaps, startPadding: 0, endPadding: remainingSpace}
	}
}

// flexDistributeSpaceBetween distributes space between items (none on edges).
func flexDistributeSpaceBetween(n, remainingSpace int, gaps []int) flexGapResult {
	if n <= 1 {
		return flexGapResult{gaps: gaps, startPadding: 0, endPadding: remainingSpace}
	}
	extraGap := remainingSpace / (n - 1)
	remainder := remainingSpace % (n - 1)
	for i := range gaps {
		gaps[i] += extraGap
		if i < remainder {
			gaps[i]++
		}
	}
	return flexGapResult{gaps: gaps, startPadding: 0, endPadding: 0}
}

// flexDistributeSpaceAround distributes space around items (half on edges).
func flexDistributeSpaceAround(n, remainingSpace int, gaps []int) flexGapResult {
	if n == 0 {
		return flexGapResult{gaps: gaps, startPadding: 0, endPadding: 0}
	}
	unitSpace := remainingSpace / (n * 2)
	startPadding := unitSpace
	endPadding := unitSpace
	if n > 1 {
		innerSpace := remainingSpace - (2 * unitSpace)
		extraGap := innerSpace / (n - 1)
		for i := range gaps {
			gaps[i] += extraGap
		}
	}
	return flexGapResult{gaps: gaps, startPadding: startPadding, endPadding: endPadding}
}

// flexDistributeSpaceEvenly distributes space evenly (including edges).
func flexDistributeSpaceEvenly(n, remainingSpace int, gaps []int) flexGapResult {
	if n == 0 {
		return flexGapResult{gaps: gaps, startPadding: 0, endPadding: 0}
	}
	slots := n + 1
	slotSize := remainingSpace / slots
	remainder := remainingSpace % slots

	startPadding := slotSize
	if remainder > 0 {
		startPadding++
		remainder--
	}
	endPadding := slotSize
	if remainder > 0 {
		endPadding++
		remainder--
	}
	for i := range gaps {
		gaps[i] += slotSize
		if remainder > 0 {
			gaps[i]++
			remainder--
		}
	}
	return flexGapResult{gaps: gaps, startPadding: startPadding, endPadding: endPadding}
}

// flexJoinRow joins items horizontally with calculated gaps.
func flexJoinRow(items []string, gaps []int, startPadding, endPadding int) string {
	if len(items) == 0 {
		return ""
	}

	var parts []string

	// Add start padding
	if startPadding > 0 {
		parts = append(parts, strings.Repeat(" ", startPadding))
	}

	for i, item := range items {
		parts = append(parts, item)
		// Add gap after item (except last)
		if i < len(gaps) && gaps[i] > 0 {
			parts = append(parts, strings.Repeat(" ", gaps[i]))
		}
	}

	// Add end padding (not typically needed for rendering, but included for completeness)
	if endPadding > 0 {
		parts = append(parts, strings.Repeat(" ", endPadding))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// flexJoinColumn joins items vertically with calculated gaps.
func flexJoinColumn(items []string, gaps []int, startPadding, endPadding int) string {
	if len(items) == 0 {
		return ""
	}

	var parts []string

	// Add start padding (empty lines)
	if startPadding > 0 {
		emptyLines := make([]string, startPadding)
		for i := range emptyLines {
			emptyLines[i] = ""
		}
		parts = append(parts, strings.Join(emptyLines, "\n"))
	}

	for i, item := range items {
		parts = append(parts, item)
		// Add gap after item (except last)
		if i < len(gaps) && gaps[i] > 0 {
			emptyLines := make([]string, gaps[i])
			for j := range emptyLines {
				emptyLines[j] = ""
			}
			parts = append(parts, strings.Join(emptyLines, "\n"))
		}
	}

	// Add end padding
	if endPadding > 0 {
		emptyLines := make([]string, endPadding)
		for i := range emptyLines {
			emptyLines[i] = ""
		}
		parts = append(parts, strings.Join(emptyLines, "\n"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// Flex creates a flexbox-style layout component.
// Items are arranged in a row or column with configurable alignment and spacing.
//
// Features:
//   - Row or column direction
//   - Main-axis alignment (justify): start, center, end, space-between, space-around, space-evenly
//   - Cross-axis alignment (align): start, center, end, stretch
//   - Configurable gap between items
//   - Fixed or auto dimensions
//   - Theme integration
//   - Custom style override support
//
// Example:
//
//	// Toolbar with space-between
//	toolbar := Flex(FlexProps{
//	    Items: []bubbly.Component{
//	        Text(TextProps{Content: "Title"}),
//	        Spacer(SpacerProps{Flex: true}),
//	        Button(ButtonProps{Label: "Action"}),
//	    },
//	    Justify: JustifySpaceBetween,
//	    Width:   80,
//	})
//
//	// Centered card grid
//	grid := Flex(FlexProps{
//	    Items:     cards,
//	    Direction: FlexRow,
//	    Justify:   JustifyCenter,
//	    Gap:       2,
//	})
//
//	// Vertical form layout
//	form := Flex(FlexProps{
//	    Items:     []bubbly.Component{input1, input2, submitBtn},
//	    Direction: FlexColumn,
//	    Align:     AlignItemsStretch,
//	    Gap:       1,
//	})
func Flex(props FlexProps) bubbly.Component {
	flexApplyDefaults(&props)

	component, _ := bubbly.NewComponent("Flex").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(FlexProps)

			// Handle empty items
			if len(p.Items) == 0 {
				return ""
			}

			// Render all items
			rendered := flexRenderItems(p.Items)
			if len(rendered) == 0 {
				return ""
			}

			var result string

			if p.Direction == FlexColumn {
				result = flexRenderColumn(rendered, p)
			} else {
				result = flexRenderRow(rendered, p)
			}

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

// flexRenderRow renders items in row (horizontal) direction.
//
//nolint:dupl // Row and column rendering have similar structure but different axis handling
func flexRenderRow(rendered []string, p FlexProps) string {
	// Calculate cross-axis (height) for alignment
	maxHeight := flexCalculateMaxHeight(rendered)

	// Apply cross-axis alignment
	aligned := make([]string, len(rendered))
	for i, item := range rendered {
		aligned[i] = flexAlignItemRow(item, maxHeight, p.Align)
	}

	// Calculate item widths for gap distribution
	itemWidths := make([]int, len(aligned))
	for i, item := range aligned {
		itemWidths[i] = lipgloss.Width(item)
	}

	// Determine container width
	containerWidth := p.Width
	if containerWidth == 0 {
		// Auto width: sum of items + gaps
		containerWidth = flexCalculateTotalWidth(aligned) + p.Gap*(len(aligned)-1)
	}

	// Calculate gaps based on justify mode
	gaps, startPadding := flexCalculateGaps(itemWidths, containerWidth, p.Justify, p.Gap)

	// Join with calculated gaps
	return flexJoinRow(aligned, gaps, startPadding, 0)
}

// flexRenderColumn renders items in column (vertical) direction.
//
//nolint:dupl // Row and column rendering have similar structure but different axis handling
func flexRenderColumn(rendered []string, p FlexProps) string {
	// Calculate cross-axis (width) for alignment
	maxWidth := flexCalculateMaxWidth(rendered)

	// Apply cross-axis alignment
	aligned := make([]string, len(rendered))
	for i, item := range rendered {
		aligned[i] = flexAlignItemColumn(item, maxWidth, p.Align)
	}

	// Calculate item heights for gap distribution
	itemHeights := make([]int, len(aligned))
	for i, item := range aligned {
		itemHeights[i] = lipgloss.Height(item)
	}

	// Determine container height
	containerHeight := p.Height
	if containerHeight == 0 {
		// Auto height: sum of items + gaps
		containerHeight = flexCalculateTotalHeight(aligned) + p.Gap*(len(aligned)-1)
	}

	// Calculate gaps based on justify mode
	gaps, startPadding := flexCalculateGaps(itemHeights, containerHeight, p.Justify, p.Gap)

	// Join with calculated gaps
	return flexJoinColumn(aligned, gaps, startPadding, 0)
}
