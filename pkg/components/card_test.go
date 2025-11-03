package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

func TestCard_Creation(t *testing.T) {
	card := Card(CardProps{
		Title:   "Test Card",
		Content: "Card content",
	})

	assert.NotNil(t, card, "Card should be created")
}

func TestCard_Rendering(t *testing.T) {
	card := Card(CardProps{
		Title:   "My Card",
		Content: "This is the card content",
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "My Card", "Should render title")
	assert.Contains(t, output, "This is the card content", "Should render content")
}

func TestCard_WithoutTitle(t *testing.T) {
	card := Card(CardProps{
		Content: "Content only",
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "Content only", "Should render content")
	assert.NotEmpty(t, output, "Should render even without title")
}

func TestCard_WithFooter(t *testing.T) {
	card := Card(CardProps{
		Title:   "Card with Footer",
		Content: "Main content",
		Footer:  "Footer text",
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "Card with Footer", "Should render title")
	assert.Contains(t, output, "Main content", "Should render content")
	assert.Contains(t, output, "Footer text", "Should render footer")
}

func TestCard_WithChildren(t *testing.T) {
	child := Text(TextProps{
		Content: "Child component",
	})
	child.Init()

	card := Card(CardProps{
		Title:    "Card with Children",
		Children: []bubbly.Component{child},
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "Card with Children", "Should render title")
	assert.Contains(t, output, "Child component", "Should render child components")
}

func TestCard_ThemeIntegration(t *testing.T) {
	card := Card(CardProps{
		Title:   "Themed Card",
		Content: "Content",
	})

	card.Init()
	output := card.View()

	assert.NotEmpty(t, output, "Should render with theme")
}

func TestCard_CustomStyle(t *testing.T) {
	customStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("99"))

	card := Card(CardProps{
		Title:   "Custom",
		Content: "Styled content",
		CommonProps: CommonProps{
			Style: &customStyle,
		},
	})

	card.Init()
	output := card.View()

	assert.NotEmpty(t, output, "Should render with custom style")
}

func TestCard_Width(t *testing.T) {
	card := Card(CardProps{
		Title:   "Wide Card",
		Content: "This is a wider card",
		Width:   60,
	})

	card.Init()
	output := card.View()

	assert.NotEmpty(t, output, "Should render with custom width")
}

func TestCard_Height(t *testing.T) {
	card := Card(CardProps{
		Title:   "Tall Card",
		Content: "This is a taller card",
		Height:  20,
	})

	card.Init()
	output := card.View()

	assert.NotEmpty(t, output, "Should render with custom height")
}

func TestCard_BubbleteatIntegration(t *testing.T) {
	card := Card(CardProps{
		Title:   "Integration",
		Content: "Test",
	})

	// Test Init
	cmd := card.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	// Test Update
	newModel, cmd := card.Update(nil)
	assert.NotNil(t, newModel, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command for nil msg")

	// Test View
	output := card.View()
	assert.NotEmpty(t, output, "View should return output")
}

func TestCard_EmptyContent(t *testing.T) {
	card := Card(CardProps{
		Title:   "Empty",
		Content: "",
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "Empty", "Should still render title")
}

func TestCard_LongContent(t *testing.T) {
	longContent := "This is a very long content that should be properly wrapped and displayed in the card. "
	longContent += "It contains multiple sentences and should handle line breaks appropriately."

	card := Card(CardProps{
		Title:   "Long Content",
		Content: longContent,
		Width:   40,
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "Long Content", "Should render title")
	assert.NotEmpty(t, output, "Should render long content")
}

func TestCard_Padding(t *testing.T) {
	card := Card(CardProps{
		Title:   "Padded Card",
		Content: "Content with padding",
		Padding: 2,
	})

	card.Init()
	output := card.View()

	assert.NotEmpty(t, output, "Should render with padding")
}

func TestCard_NoBorder(t *testing.T) {
	card := Card(CardProps{
		Title:    "No Border",
		Content:  "Card without border",
		NoBorder: true,
	})

	card.Init()
	output := card.View()

	assert.Contains(t, output, "No Border", "Should render title")
	assert.Contains(t, output, "Card without border", "Should render content")
}
