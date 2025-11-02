package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPanelLayout_Creation tests that PanelLayout component can be created.
func TestPanelLayout_Creation(t *testing.T) {
	left := Text(TextProps{Content: "Left Panel"})
	right := Text(TextProps{Content: "Right Panel"})

	layout := PanelLayout(PanelLayoutProps{
		Left:  left,
		Right: right,
	})

	assert.NotNil(t, layout, "PanelLayout should be created")
}

// TestPanelLayout_HorizontalSplit tests horizontal split (left/right).
func TestPanelLayout_HorizontalSplit(t *testing.T) {
	left := Text(TextProps{Content: "Master List"})
	left.Init()

	right := Text(TextProps{Content: "Detail View"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:      left,
		Right:     right,
		Direction: "horizontal",
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Master List", "Should render left panel")
	assert.Contains(t, output, "Detail View", "Should render right panel")
}

// TestPanelLayout_VerticalSplit tests vertical split (top/bottom).
func TestPanelLayout_VerticalSplit(t *testing.T) {
	top := Text(TextProps{Content: "Top Section"})
	top.Init()

	bottom := Text(TextProps{Content: "Bottom Section"})
	bottom.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:      top,    // Using Left for Top
		Right:     bottom, // Using Right for Bottom
		Direction: "vertical",
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Top Section", "Should render top panel")
	assert.Contains(t, output, "Bottom Section", "Should render bottom panel")

	// Top should come before bottom
	topIdx := strings.Index(output, "Top Section")
	bottomIdx := strings.Index(output, "Bottom Section")
	assert.True(t, topIdx < bottomIdx, "Top should come before bottom")
}

// TestPanelLayout_CustomSplit tests custom split ratio.
func TestPanelLayout_CustomSplit(t *testing.T) {
	left := Text(TextProps{Content: "Narrow"})
	left.Init()

	right := Text(TextProps{Content: "Wide"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:       left,
		Right:      right,
		SplitRatio: 0.3, // 30% left, 70% right
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Narrow", "Should render left panel")
	assert.Contains(t, output, "Wide", "Should render right panel")
}

// TestPanelLayout_WithOnlyLeft tests panel with only left/top section.
func TestPanelLayout_WithOnlyLeft(t *testing.T) {
	left := Text(TextProps{Content: "Only Left"})
	left.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left: left,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Only Left", "Should render left panel")
	assert.NotEmpty(t, output, "Should produce output")
}

// TestPanelLayout_WithOnlyRight tests panel with only right/bottom section.
func TestPanelLayout_WithOnlyRight(t *testing.T) {
	right := Text(TextProps{Content: "Only Right"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Right: right,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Only Right", "Should render right panel")
	assert.NotEmpty(t, output, "Should produce output")
}

// TestPanelLayout_WithBorder tests panel with border enabled.
func TestPanelLayout_WithBorder(t *testing.T) {
	left := Text(TextProps{Content: "Left"})
	left.Init()

	right := Text(TextProps{Content: "Right"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:       left,
		Right:      right,
		ShowBorder: true,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Left", "Should render left panel")
	assert.Contains(t, output, "Right", "Should render right panel")
	assert.NotEmpty(t, output, "Should produce bordered output")
}

// TestPanelLayout_CustomDimensions tests custom width and height.
func TestPanelLayout_CustomDimensions(t *testing.T) {
	left := Text(TextProps{Content: "Left"})
	left.Init()

	right := Text(TextProps{Content: "Right"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:   left,
		Right:  right,
		Width:  100,
		Height: 30,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Left", "Should render left panel")
	assert.Contains(t, output, "Right", "Should render right panel")
}

// TestPanelLayout_MasterDetailPattern tests master-detail use case.
func TestPanelLayout_MasterDetailPattern(t *testing.T) {
	// Master list - using Card instead of List for simplicity
	master := Card(CardProps{
		Title:   "Items",
		Content: "Item 1\nItem 2\nItem 3",
	})
	master.Init()

	// Detail view
	detail := Card(CardProps{
		Title:   "Item Details",
		Content: "Selected item information",
	})
	detail.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:       master,
		Right:      detail,
		SplitRatio: 0.3,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Items", "Should render master list")
	assert.Contains(t, output, "Item Details", "Should render detail view")
}

// TestPanelLayout_ThemeIntegration tests theme integration.
func TestPanelLayout_ThemeIntegration(t *testing.T) {
	left := Text(TextProps{Content: "Themed Left"})
	left.Init()

	right := Text(TextProps{Content: "Themed Right"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:  left,
		Right: right,
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Themed Left", "Should render with theme")
	assert.Contains(t, output, "Themed Right", "Should render with theme")
}

// TestPanelLayout_BubbleteatIntegration tests Bubbletea integration.
func TestPanelLayout_BubbleteatIntegration(t *testing.T) {
	left := Text(TextProps{Content: "TUI Left"})
	left.Init()

	right := Text(TextProps{Content: "TUI Right"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:  left,
		Right: right,
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
	assert.Contains(t, output, "TUI Left", "Should render left panel")
	assert.Contains(t, output, "TUI Right", "Should render right panel")
}

// TestPanelLayout_EmptyLayout tests panel with no sections.
func TestPanelLayout_EmptyLayout(t *testing.T) {
	layout := PanelLayout(PanelLayoutProps{})

	layout.Init()
	output := layout.View()

	assert.NotNil(t, output, "Should produce output even when empty")
}

// TestPanelLayout_DefaultDirection tests default horizontal direction.
func TestPanelLayout_DefaultDirection(t *testing.T) {
	left := Text(TextProps{Content: "Default Left"})
	left.Init()

	right := Text(TextProps{Content: "Default Right"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:  left,
		Right: right,
		// Direction not specified, should default to horizontal
	})

	layout.Init()
	output := layout.View()

	assert.Contains(t, output, "Default Left", "Should render left panel")
	assert.Contains(t, output, "Default Right", "Should render right panel")
}
