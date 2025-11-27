package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CardProps defines the properties for the Card component.
// Card is an organism component that displays a content container with optional title and footer.
type CardProps struct {
	// Title is the card header text.
	// Optional - if empty, no title is displayed.
	Title string

	// Content is the main body text of the card.
	// Optional - can be empty if using Children instead.
	Content string

	// Footer is optional footer text displayed at the bottom.
	// Optional - if empty, no footer is displayed.
	Footer string

	// Children are optional child components rendered in the card body.
	// These are rendered after Content if both are provided.
	// Optional - if empty, only Content is displayed.
	Children []bubbly.Component

	// Width sets the card width in characters.
	// Default is 40 if not specified.
	Width int

	// Height sets the card height in lines.
	// Default is auto-height based on content if not specified.
	Height int

	// Padding sets the internal padding.
	// Default is 1 if not specified.
	Padding int

	// NoBorder removes the border if true.
	// Default is false (border is shown).
	NoBorder bool

	// CommonProps for styling and identification.
	CommonProps
}

// cardApplyDefaults sets default values for CardProps.
func cardApplyDefaults(props *CardProps) {
	if props.Width == 0 {
		props.Width = 40
	}
	if props.Padding == 0 {
		props.Padding = 1
	}
}

// cardRenderContent builds the card's inner content string.
func cardRenderContent(p CardProps, theme Theme) string {
	var content strings.Builder
	contentWidth := p.Width - (p.Padding * 2)

	// Title
	if p.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			Width(contentWidth)
		content.WriteString(titleStyle.Render(p.Title))
		content.WriteString("\n\n")
	}

	// Content
	if p.Content != "" {
		contentStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Foreground(theme.Foreground)
		content.WriteString(contentStyle.Render(p.Content))
		if len(p.Children) > 0 {
			content.WriteString("\n")
		}
	}

	// Children
	for _, child := range p.Children {
		content.WriteString(child.View())
		content.WriteString("\n")
	}

	// Footer
	if p.Footer != "" {
		if p.Content != "" || len(p.Children) > 0 {
			content.WriteString("\n")
		}
		footerStyle := lipgloss.NewStyle().
			Foreground(theme.Muted).
			Italic(true).
			Width(contentWidth)
		content.WriteString(footerStyle.Render(p.Footer))
	}

	return content.String()
}

// cardCreateStyle creates the card's outer style.
func cardCreateStyle(p CardProps, theme Theme) lipgloss.Style {
	cardStyle := lipgloss.NewStyle().
		Width(p.Width).
		Padding(p.Padding)

	if !p.NoBorder {
		cardStyle = cardStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Secondary)
	}

	if p.Height > 0 {
		cardStyle = cardStyle.Height(p.Height)
	}

	if p.Style != nil {
		cardStyle = cardStyle.Inherit(*p.Style)
	}

	return cardStyle
}

// Card creates a card container component.
// The card displays a bordered box with optional title, content, footer, and child components.
//
// Features:
//   - Optional title header
//   - Content text or child components
//   - Optional footer
//   - Configurable width, height, and padding
//   - Border can be toggled on/off
//   - Theme integration for consistent styling
//   - Custom style override support
//
// Example:
//
//	card := Card(CardProps{
//	    Title:   "User Profile",
//	    Content: "Name: John Doe\nEmail: john@example.com",
//	    Footer:  "Last updated: 2024-01-01",
//	    Width:   50,
//	})
func Card(props CardProps) bubbly.Component {
	cardApplyDefaults(&props)

	component, _ := bubbly.NewComponent("Card").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(CardProps)
			theme := ctx.Get("theme").(Theme)

			content := cardRenderContent(p, theme)
			cardStyle := cardCreateStyle(p, theme)

			return cardStyle.Render(content)
		}).
		Build()

	return component
}
