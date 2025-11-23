package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// GridLayoutProps defines the properties for the GridLayout component.
// GridLayout is a template component that provides a grid-based layout system.
type GridLayoutProps struct {
	// Items are the components to arrange in the grid.
	// Required - the components to display in grid cells.
	Items []bubbly.Component

	// Columns sets the number of columns in the grid.
	// Default is 1 if not specified.
	Columns int

	// Gap sets the spacing between grid cells in characters.
	// Default is 1 if not specified.
	Gap int

	// CellWidth sets the width of each grid cell in characters.
	// Default is 20 if not specified.
	CellWidth int

	// CellHeight sets the height of each grid cell in lines.
	// Default is 0 (auto-height based on content) if not specified.
	CellHeight int

	// CommonProps for styling and identification.
	CommonProps
}

// gridLayoutApplyDefaults sets default values for GridLayoutProps.
func gridLayoutApplyDefaults(props *GridLayoutProps) {
	if props.Columns == 0 {
		props.Columns = 1
	}
	if props.Gap == 0 {
		props.Gap = 1
	}
	if props.CellWidth == 0 {
		props.CellWidth = 20
	}
}

// gridLayoutJoinWithGaps joins items with gap spacing.
func gridLayoutJoinWithGaps(items []string, gap string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		result = append(result, item)
		if i < len(items)-1 {
			result = append(result, gap)
		}
	}
	return result
}

// gridLayoutBuildRows builds all grid rows from items.
func gridLayoutBuildRows(items []bubbly.Component, columns, gap int) []string {
	if len(items) == 0 {
		return nil
	}

	gapSpacer := lipgloss.NewStyle().Width(gap).Render("")
	var rows []string
	itemIndex := 0

	for itemIndex < len(items) {
		rowCells := make([]string, 0, columns)
		for col := 0; col < columns && itemIndex < len(items); col++ {
			rowCells = append(rowCells, items[itemIndex].View())
			itemIndex++
		}

		if len(rowCells) > 0 {
			rowWithGaps := gridLayoutJoinWithGaps(rowCells, gapSpacer)
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowWithGaps...))
		}
	}

	return rows
}

// gridLayoutRender renders the complete grid layout.
func gridLayoutRender(p GridLayoutProps) string {
	rows := gridLayoutBuildRows(p.Items, p.Columns, p.Gap)
	if len(rows) == 0 {
		return ""
	}

	vGapSpacer := ""
	if p.Gap > 0 {
		vGapSpacer = strings.Repeat("\n", p.Gap)
	}

	rowsWithGaps := gridLayoutJoinWithGaps(rows, vGapSpacer)
	result := lipgloss.JoinVertical(lipgloss.Left, rowsWithGaps...)

	if p.Style != nil {
		result = lipgloss.NewStyle().Inherit(*p.Style).Render(result)
	}

	return result
}

// GridLayout creates a grid-based layout template component.
// The layout arranges items in a grid with configurable columns, gaps, and cell dimensions.
//
// Layout Structure (3 columns example):
//
//	┌─────────┬─────────┬─────────┐
//	│ Cell 1  │ Cell 2  │ Cell 3  │
//	├─────────┼─────────┼─────────┤
//	│ Cell 4  │ Cell 5  │ Cell 6  │
//	└─────────┴─────────┴─────────┘
//
// Features:
//   - Configurable number of columns
//   - Adjustable gap between cells
//   - Custom cell width and height
//   - Automatic row wrapping
//   - Theme integration for consistent styling
//   - Custom style override support
//   - Perfect for dashboards and card grids
//
// Example:
//
//	layout := GridLayout(GridLayoutProps{
//	    Items: []bubbly.Component{
//	        Card(CardProps{Title: "Card 1"}),
//	        Card(CardProps{Title: "Card 2"}),
//	        Card(CardProps{Title: "Card 3"}),
//	    },
//	    Columns:   3,
//	    Gap:       2,
//	    CellWidth: 25,
//	})
func GridLayout(props GridLayoutProps) bubbly.Component {
	gridLayoutApplyDefaults(&props)

	component, _ := bubbly.NewComponent("GridLayout").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(GridLayoutProps)
			return gridLayoutRender(p)
		}).
		Build()

	return component
}
