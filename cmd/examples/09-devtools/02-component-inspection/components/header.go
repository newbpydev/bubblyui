package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// HeaderProps defines the props for Header component
type HeaderProps struct {
	Title string
}

// CreateHeader creates a header component
// Demonstrates:
// - Simple display component
// - Using Text component
// - Props-based configuration
func CreateHeader(props HeaderProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("Header")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		ctx.Expose("title", props.Title)
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		title := ctx.Get("title").(string)

		// Use Text component with custom styling
		titleText := components.Text(components.TextProps{
			Content: "📋 " + title,
			Bold:    true,
			Color:   lipgloss.Color("99"), // Purple
		})
		titleText.Init()

		// Add some spacing
		style := lipgloss.NewStyle().MarginBottom(1)
		return style.Render(titleText.View())
	})

	return builder.Build()
}
