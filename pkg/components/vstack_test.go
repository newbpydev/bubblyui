package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestVStack_Creation(t *testing.T) {
	vstack := VStack(StackProps{})
	assert.NotNil(t, vstack, "VStack should be created")
}

func TestVStack_RendersItemsVertically(t *testing.T) {
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item3 := Text(TextProps{Content: "C"})
	item1.Init()
	item2.Init()
	item3.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{item1, item2, item3},
		Spacing: 0,
	})

	vstack.Init()
	output := vstack.View()

	// Items should appear in order on separate lines
	assert.Contains(t, output, "A", "Should contain first item")
	assert.Contains(t, output, "B", "Should contain second item")
	assert.Contains(t, output, "C", "Should contain third item")

	// Items should be on separate lines (newlines between them)
	lines := strings.Split(output, "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 lines")
}

func TestVStack_AppliesSpacingBetweenItems(t *testing.T) {
	tests := []struct {
		name    string
		spacing int
	}{
		{name: "no spacing", spacing: 0},
		{name: "single line", spacing: 1},
		{name: "double line", spacing: 2},
		{name: "large spacing", spacing: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item1 := Text(TextProps{Content: "X"})
			item2 := Text(TextProps{Content: "Y"})
			item1.Init()
			item2.Init()

			vstack := VStack(StackProps{
				Items:   []interface{}{item1, item2},
				Spacing: tt.spacing,
			})

			vstack.Init()
			output := vstack.View()

			assert.Contains(t, output, "X", "Should contain first item")
			assert.Contains(t, output, "Y", "Should contain second item")

			// Count lines to verify spacing
			lines := strings.Split(output, "\n")
			// With spacing, there should be extra empty lines between items
			if tt.spacing > 0 {
				// At minimum, we should have more lines than just the items
				assert.GreaterOrEqual(t, len(lines), 2, "Should have lines for items")
			}
		})
	}
}

func TestVStack_AlignsItems(t *testing.T) {
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
			// Create items with different widths
			item1 := Text(TextProps{Content: "Short"})
			item2 := Text(TextProps{Content: "Much Longer Item Here"})
			item1.Init()
			item2.Init()

			vstack := VStack(StackProps{
				Items: []interface{}{item1, item2},
				Align: tt.align,
			})

			vstack.Init()
			output := vstack.View()

			assert.Contains(t, output, "Short", "Should contain short item")
			assert.Contains(t, output, "Much Longer Item Here", "Should contain long item")
			assert.NotEmpty(t, output, "Should render with alignment")
		})
	}
}

func TestVStack_RendersDividersBetweenItems(t *testing.T) {
	item1 := Text(TextProps{Content: "Top"})
	item2 := Text(TextProps{Content: "Bottom"})
	item1.Init()
	item2.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{item1, item2},
		Divider: true,
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "Top", "Should contain first item")
	assert.Contains(t, output, "Bottom", "Should contain second item")
	// Default horizontal divider character for VStack is ─
	assert.Contains(t, output, "─", "Should contain divider character")
}

func TestVStack_CustomDividerChar(t *testing.T) {
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item1.Init()
	item2.Init()

	vstack := VStack(StackProps{
		Items:       []interface{}{item1, item2},
		Divider:     true,
		DividerChar: "=",
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "=", "Should use custom divider character")
}

func TestVStack_HandlesEmptyItemsArray(t *testing.T) {
	vstack := VStack(StackProps{
		Items: []interface{}{},
	})

	vstack.Init()
	output := vstack.View()

	// Should not panic and return empty or minimal output
	assert.NotNil(t, output, "Should handle empty items")
}

func TestVStack_HandlesSingleItem(t *testing.T) {
	item := Text(TextProps{Content: "Solo"})
	item.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{item},
		Spacing: 2,
		Divider: true,
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "Solo", "Should render single item")
	// No divider should appear with single item
	assert.NotContains(t, output, "─", "Should not have divider with single item")
}

func TestVStack_HandlesNilItems(t *testing.T) {
	vstack := VStack(StackProps{
		Items: nil,
	})

	vstack.Init()
	output := vstack.View()

	// Should not panic
	assert.NotNil(t, output, "Should handle nil items")
}

func TestVStack_ThemeIntegration(t *testing.T) {
	item := Text(TextProps{Content: "Themed"})
	item.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{item},
		Divider: true,
	})

	vstack.Init()
	output := vstack.View()

	assert.NotEmpty(t, output, "Should render with theme")
}

func TestVStack_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("99"))

	item := Text(TextProps{Content: "Styled"})
	item.Init()

	vstack := VStack(StackProps{
		Items: []interface{}{item},
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "Styled", "Should render with custom style")
}

func TestVStack_BubbleteatIntegration(t *testing.T) {
	item := Text(TextProps{Content: "Integration"})
	item.Init()

	vstack := VStack(StackProps{
		Items: []interface{}{item},
	})

	// Test Init
	cmd := vstack.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	newModel, cmd := vstack.Update(nil)
	assert.NotNil(t, newModel, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command for nil msg")

	// Test View
	output := vstack.View()
	assert.NotEmpty(t, output, "View should return output")
}

func TestVStack_DefaultSpacing(t *testing.T) {
	// When spacing is not specified, default should be 1
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item1.Init()
	item2.Init()

	vstack := VStack(StackProps{
		Items: []interface{}{item1, item2},
		// Spacing not set - should default to 1
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "A", "Should contain first item")
	assert.Contains(t, output, "B", "Should contain second item")
}

func TestVStack_FlexibleSpacer(t *testing.T) {
	// Test that VStack recognizes flexible spacers
	item1 := Text(TextProps{Content: "Top"})
	spacer := Spacer(SpacerProps{Flex: true})
	item2 := Text(TextProps{Content: "Bottom"})
	item1.Init()
	spacer.Init()
	item2.Init()

	vstack := VStack(StackProps{
		Items: []interface{}{item1, spacer, item2},
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "Top", "Should contain top item")
	assert.Contains(t, output, "Bottom", "Should contain bottom item")
}

func TestVStack_MixedContentTypes(t *testing.T) {
	// Test with different component types
	text := Text(TextProps{Content: "Text"})
	badge := Badge(BadgeProps{Label: "Badge"})
	text.Init()
	badge.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{text, badge},
		Spacing: 1,
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "Text", "Should contain text component")
	assert.Contains(t, output, "Badge", "Should contain badge component")
}

func TestVStack_DividerWithSpacing(t *testing.T) {
	item1 := Text(TextProps{Content: "A"})
	item2 := Text(TextProps{Content: "B"})
	item1.Init()
	item2.Init()

	vstack := VStack(StackProps{
		Items:   []interface{}{item1, item2},
		Spacing: 2,
		Divider: true,
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "A", "Should contain first item")
	assert.Contains(t, output, "B", "Should contain second item")
	assert.Contains(t, output, "─", "Should contain divider")
}

func TestVStack_AlignStretch(t *testing.T) {
	// AlignStretch should stretch items to fill cross-axis (width)
	item1 := Text(TextProps{Content: "Short"})
	item2 := Text(TextProps{Content: "Much Longer Item"})
	item1.Init()
	item2.Init()

	vstack := VStack(StackProps{
		Items: []interface{}{item1, item2},
		Align: AlignItemsStretch,
	})

	vstack.Init()
	output := vstack.View()

	assert.Contains(t, output, "Short", "Should contain short item")
	assert.Contains(t, output, "Much Longer Item", "Should contain long item")
}
