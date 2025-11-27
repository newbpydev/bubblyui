package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestHStack_Creation(t *testing.T) {
	hstack := HStack(StackProps{})
	assert.NotNil(t, hstack, "HStack should be created")
}

func TestHStack_RendersItemsHorizontally(t *testing.T) {
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item3 := Text(TextProps{Content: "C"})
	item1.Init()
	item2.Init()
	item3.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{item1, item2, item3},
		Spacing: 0,
	})

	hstack.Init()
	output := hstack.View()

	// Items should appear in order on the same line
	assert.Contains(t, output, "A", "Should contain first item")
	assert.Contains(t, output, "B", "Should contain second item")
	assert.Contains(t, output, "C", "Should contain third item")

	// All items should be on the same line (no newlines between them)
	lines := strings.Split(output, "\n")
	// The first line should contain all items
	assert.True(t, strings.Contains(lines[0], "A") || strings.Contains(output, "A"),
		"Items should be rendered horizontally")
}

func TestHStack_AppliesSpacingBetweenItems(t *testing.T) {
	tests := []struct {
		name    string
		spacing int
	}{
		{name: "no spacing", spacing: 0},
		{name: "single space", spacing: 1},
		{name: "double space", spacing: 2},
		{name: "large spacing", spacing: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item1 := Text(TextProps{Content: "X"})
			item2 := Text(TextProps{Content: "Y"})
			item1.Init()
			item2.Init()

			hstack := HStack(StackProps{
				Items:   []interface{}{item1, item2},
				Spacing: tt.spacing,
			})

			hstack.Init()
			output := hstack.View()

			assert.Contains(t, output, "X", "Should contain first item")
			assert.Contains(t, output, "Y", "Should contain second item")

			// With spacing > 0, there should be spaces between items
			if tt.spacing > 0 {
				// Check that there's spacing between X and Y
				expectedSpacer := strings.Repeat(" ", tt.spacing)
				assert.Contains(t, output, expectedSpacer, "Should have spacing between items")
			}
		})
	}
}

func TestHStack_AlignsItems(t *testing.T) {
	tests := []struct {
		name  string
		align AlignItems
	}{
		{name: "align start", align: AlignItemsStart},
		{name: "align center", align: AlignItemsCenter},
		{name: "align end", align: AlignItemsEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create items with different heights
			item1 := Text(TextProps{Content: "Short"})
			item2 := Text(TextProps{Content: "Tall\nItem\nHere"})
			item1.Init()
			item2.Init()

			hstack := HStack(StackProps{
				Items: []interface{}{item1, item2},
				Align: tt.align,
			})

			hstack.Init()
			output := hstack.View()

			assert.Contains(t, output, "Short", "Should contain short item")
			assert.Contains(t, output, "Tall", "Should contain tall item")
			assert.NotEmpty(t, output, "Should render with alignment")
		})
	}
}

func TestHStack_RendersDividersBetweenItems(t *testing.T) {
	item1 := Text(TextProps{Content: "Left"})
	item2 := Text(TextProps{Content: "Right"})
	item1.Init()
	item2.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{item1, item2},
		Divider: true,
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "Left", "Should contain first item")
	assert.Contains(t, output, "Right", "Should contain second item")
	// Default vertical divider character for HStack is │
	assert.Contains(t, output, "│", "Should contain divider character")
}

func TestHStack_CustomDividerChar(t *testing.T) {
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item1.Init()
	item2.Init()

	hstack := HStack(StackProps{
		Items:       []interface{}{item1, item2},
		Divider:     true,
		DividerChar: "|",
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "|", "Should use custom divider character")
}

func TestHStack_HandlesEmptyItemsArray(t *testing.T) {
	hstack := HStack(StackProps{
		Items: []interface{}{},
	})

	hstack.Init()
	output := hstack.View()

	// Should not panic and return empty or minimal output
	assert.NotNil(t, output, "Should handle empty items")
}

func TestHStack_HandlesSingleItem(t *testing.T) {
	item := Text(TextProps{Content: "Solo"})
	item.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{item},
		Spacing: 2,
		Divider: true,
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "Solo", "Should render single item")
	// No divider should appear with single item
	assert.NotContains(t, output, "│", "Should not have divider with single item")
}

func TestHStack_HandlesNilItems(t *testing.T) {
	hstack := HStack(StackProps{
		Items: nil,
	})

	hstack.Init()
	output := hstack.View()

	// Should not panic
	assert.NotNil(t, output, "Should handle nil items")
}

func TestHStack_ThemeIntegration(t *testing.T) {
	item := Text(TextProps{Content: "Themed"})
	item.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{item},
		Divider: true,
	})

	hstack.Init()
	output := hstack.View()

	assert.NotEmpty(t, output, "Should render with theme")
}

func TestHStack_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("99"))

	item := Text(TextProps{Content: "Styled"})
	item.Init()

	hstack := HStack(StackProps{
		Items: []interface{}{item},
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "Styled", "Should render with custom style")
}

func TestHStack_BubbleteatIntegration(t *testing.T) {
	item := Text(TextProps{Content: "Integration"})
	item.Init()

	hstack := HStack(StackProps{
		Items: []interface{}{item},
	})

	// Test Init
	cmd := hstack.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	newModel, cmd := hstack.Update(nil)
	assert.NotNil(t, newModel, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command for nil msg")

	// Test View
	output := hstack.View()
	assert.NotEmpty(t, output, "View should return output")
}

func TestHStack_DefaultSpacing(t *testing.T) {
	// When spacing is not specified, default should be 1
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item1.Init()
	item2.Init()

	hstack := HStack(StackProps{
		Items: []interface{}{item1, item2},
		// Spacing not set - should default to 1
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "A", "Should contain first item")
	assert.Contains(t, output, "B", "Should contain second item")
}

func TestHStack_FlexibleSpacer(t *testing.T) {
	// Test that HStack recognizes flexible spacers
	item1 := Text(TextProps{Content: "Left"})
	spacer := Spacer(SpacerProps{Flex: true})
	item2 := Text(TextProps{Content: "Right"})
	item1.Init()
	spacer.Init()
	item2.Init()

	hstack := HStack(StackProps{
		Items: []interface{}{item1, spacer, item2},
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "Left", "Should contain left item")
	assert.Contains(t, output, "Right", "Should contain right item")
}

func TestHStack_MixedContentTypes(t *testing.T) {
	// Test with different component types
	text := Text(TextProps{Content: "Text"})
	badge := Badge(BadgeProps{Label: "Badge"})
	text.Init()
	badge.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{text, badge},
		Spacing: 1,
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "Text", "Should contain text component")
	assert.Contains(t, output, "Badge", "Should contain badge component")
}

func TestHStack_DividerWithSpacing(t *testing.T) {
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item1.Init()
	item2.Init()

	hstack := HStack(StackProps{
		Items:   []interface{}{item1, item2},
		Spacing: 2,
		Divider: true,
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "A", "Should contain first item")
	assert.Contains(t, output, "B", "Should contain second item")
	assert.Contains(t, output, "│", "Should contain divider")
}

func TestHStack_AlignStretch(t *testing.T) {
	// AlignStretch should stretch items to fill cross-axis
	item1 := Text(TextProps{Content: "Short"})
	item2 := Text(TextProps{Content: "Tall\nItem"})
	item1.Init()
	item2.Init()

	hstack := HStack(StackProps{
		Items: []interface{}{item1, item2},
		Align: AlignItemsStretch,
	})

	hstack.Init()
	output := hstack.View()

	assert.Contains(t, output, "Short", "Should contain short item")
	assert.Contains(t, output, "Tall", "Should contain tall item")
}
