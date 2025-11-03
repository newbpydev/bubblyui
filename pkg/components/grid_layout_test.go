package components

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// TestGridLayout_Creation tests that GridLayout component can be created.
func TestGridLayout_Creation(t *testing.T) {
	cell1 := Text(TextProps{Content: "Cell 1"})
	cell2 := Text(TextProps{Content: "Cell 2"})

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell1, cell2},
		Columns: 2,
	})

	assert.NotNil(t, layout, "GridLayout should be created")
}

// TestGridLayout_TwoColumns tests 2-column grid.
func TestGridLayout_TwoColumns(t *testing.T) {
	cell1 := Text(TextProps{Content: "Cell 1"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "Cell 2"})
	cell2.Init()

	cell3 := Text(TextProps{Content: "Cell 3"})
	cell3.Init()

	cell4 := Text(TextProps{Content: "Cell 4"})
	cell4.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell1, cell2, cell3, cell4},
		Columns: 2,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Cell 1", "Should render cell 1")
	assert.Contains(t, output, "Cell 2", "Should render cell 2")
	assert.Contains(t, output, "Cell 3", "Should render cell 3")
	assert.Contains(t, output, "Cell 4", "Should render cell 4")
}

// TestGridLayout_ThreeColumns tests 3-column grid.
func TestGridLayout_ThreeColumns(t *testing.T) {
	items := make([]bubbly.Component, 6)
	for i := 0; i < 6; i++ {
		item := Text(TextProps{Content: "Item"})
		item.Init()
		items[i] = item
	}

	layout := GridLayout(GridLayoutProps{
		Items:   items,
		Columns: 3,
	})

	layout.Init()
	output := layout.View()

	assert.NotEmpty(t, output, "Should produce output")
}

// TestGridLayout_SingleColumn tests 1-column grid (vertical list).
func TestGridLayout_SingleColumn(t *testing.T) {
	cell1 := Text(TextProps{Content: "First"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "Second"})
	cell2.Init()

	cell3 := Text(TextProps{Content: "Third"})
	cell3.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell1, cell2, cell3},
		Columns: 1,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "First", "Should render first item")
	assert.Contains(t, output, "Second", "Should render second item")
	assert.Contains(t, output, "Third", "Should render third item")
}

// TestGridLayout_WithGap tests grid with custom gap.
func TestGridLayout_WithGap(t *testing.T) {
	cell1 := Text(TextProps{Content: "A"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "B"})
	cell2.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell1, cell2},
		Columns: 2,
		Gap:     2,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "A", "Should render cell A")
	assert.Contains(t, output, "B", "Should render cell B")
}

// TestGridLayout_WithCellWidth tests grid with custom cell width.
func TestGridLayout_WithCellWidth(t *testing.T) {
	cell1 := Text(TextProps{Content: "Wide Cell"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "Another"})
	cell2.Init()

	layout := GridLayout(GridLayoutProps{
		Items:     []bubbly.Component{cell1, cell2},
		Columns:   2,
		CellWidth: 30,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Wide Cell", "Should render cell 1")
	assert.Contains(t, output, "Another", "Should render cell 2")
}

// TestGridLayout_WithCellHeight tests grid with custom cell height.
func TestGridLayout_WithCellHeight(t *testing.T) {
	cell1 := Text(TextProps{Content: "Tall"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "Cell"})
	cell2.Init()

	layout := GridLayout(GridLayoutProps{
		Items:      []bubbly.Component{cell1, cell2},
		Columns:    2,
		CellHeight: 5,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Tall", "Should render cell 1")
	assert.Contains(t, output, "Cell", "Should render cell 2")
}

// TestGridLayout_EmptyItems tests grid with no items.
func TestGridLayout_EmptyItems(t *testing.T) {
	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{},
		Columns: 2,
	})

	layout.Init()
	output := layout.View()

	assert.NotNil(t, output, "Should produce output even when empty")
}

// TestGridLayout_CardGrid tests grid of cards.
func TestGridLayout_CardGrid(t *testing.T) {
	card1 := Card(CardProps{
		Title:   "Card 1",
		Content: "Content 1",
	})
	card1.Init()

	card2 := Card(CardProps{
		Title:   "Card 2",
		Content: "Content 2",
	})
	card2.Init()

	card3 := Card(CardProps{
		Title:   "Card 3",
		Content: "Content 3",
	})
	card3.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{card1, card2, card3},
		Columns: 3,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Card 1", "Should render card 1")
	assert.Contains(t, output, "Card 2", "Should render card 2")
	assert.Contains(t, output, "Card 3", "Should render card 3")
}

// TestGridLayout_ThemeIntegration tests theme integration.
func TestGridLayout_ThemeIntegration(t *testing.T) {
	cell := Text(TextProps{Content: "Themed"})
	cell.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell},
		Columns: 1,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Themed", "Should render with theme")
	assert.NotEmpty(t, output, "Should produce themed output")
}

// TestGridLayout_BubbleteatIntegration tests Bubbletea integration.
func TestGridLayout_BubbleteatIntegration(t *testing.T) {
	cell := Text(TextProps{Content: "TUI Grid"})
	cell.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell},
		Columns: 1,
	})

	// Test Init
	cmd := layout.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	model, cmd := layout.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	// Test View
	output := layout.View()
	assert.NotEmpty(t, output, "View should produce output")
	assert.Contains(t, output, "TUI Grid", "Should render cell")
}

// TestGridLayout_DefaultColumns tests default column count.
func TestGridLayout_DefaultColumns(t *testing.T) {
	cell1 := Text(TextProps{Content: "A"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "B"})
	cell2.Init()

	layout := GridLayout(GridLayoutProps{
		Items: []bubbly.Component{cell1, cell2},
		// Columns not specified, should default to 1
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "A", "Should render cell A")
	assert.Contains(t, output, "B", "Should render cell B")
}

// TestGridLayout_UnevenItems tests grid with uneven number of items.
func TestGridLayout_UnevenItems(t *testing.T) {
	cell1 := Text(TextProps{Content: "1"})
	cell1.Init()

	cell2 := Text(TextProps{Content: "2"})
	cell2.Init()

	cell3 := Text(TextProps{Content: "3"})
	cell3.Init()

	layout := GridLayout(GridLayoutProps{
		Items:   []bubbly.Component{cell1, cell2, cell3},
		Columns: 2, // 3 items in 2 columns = 2 rows (last row has 1 item)
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "1", "Should render cell 1")
	assert.Contains(t, output, "2", "Should render cell 2")
	assert.Contains(t, output, "3", "Should render cell 3")
}
