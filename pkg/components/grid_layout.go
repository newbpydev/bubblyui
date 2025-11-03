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
	// Set defaults
	if props.Columns == 0 {
		props.Columns = 1
	}
	if props.Gap == 0 {
		props.Gap = 1
	}
	if props.CellWidth == 0 {
		props.CellWidth = 20
	}

	component, _ := bubbly.NewComponent("GridLayout").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(GridLayoutProps)

			if len(p.Items) == 0 {
				return ""
			}

			// Build grid row by row
			var rows []string
			itemIndex := 0

			for itemIndex < len(p.Items) {
				// Build one row
				var rowCells []string

				for col := 0; col < p.Columns && itemIndex < len(p.Items); col++ {
					item := p.Items[itemIndex]

					// Render cell content directly without width constraint
					// This allows bordered components like cards to render properly
					cellContent := item.View()
					rowCells = append(rowCells, cellContent)

					itemIndex++
				}

				// Join cells in row horizontally with gap
				if len(rowCells) > 0 {
					// Create gap spacer
					gapSpacer := lipgloss.NewStyle().
						Width(p.Gap).
						Render("")

					// Join row cells with gaps
					var rowWithGaps []string
					for i, cell := range rowCells {
						rowWithGaps = append(rowWithGaps, cell)
						if i < len(rowCells)-1 {
							rowWithGaps = append(rowWithGaps, gapSpacer)
						}
					}

					row := lipgloss.JoinHorizontal(lipgloss.Top, rowWithGaps...)
					rows = append(rows, row)
				}
			}

			// Join rows vertically with gap
			if len(rows) == 0 {
				return ""
			}

			// Create vertical gap spacer
			vGapSpacer := ""
			if p.Gap > 0 {
				vGapSpacer = strings.Repeat("\n", p.Gap)
			}

			// Join rows with vertical gaps
			var rowsWithGaps []string
			for i, row := range rows {
				rowsWithGaps = append(rowsWithGaps, row)
				if i < len(rows)-1 {
					rowsWithGaps = append(rowsWithGaps, vGapSpacer)
				}
			}

			result := lipgloss.JoinVertical(lipgloss.Left, rowsWithGaps...)

			// Apply custom style if provided
			if p.Style != nil {
				containerStyle := lipgloss.NewStyle().Inherit(*p.Style)
				result = containerStyle.Render(result)
			}

			return result
		}).
		Build()

	return component
}
