package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// MenuItem represents a single menu item.
type MenuItem struct {
	// Label is the display text for the menu item.
	Label string

	// Value is the unique identifier for the menu item.
	Value string

	// Disabled indicates if the menu item is disabled.
	Disabled bool
}

// MenuProps defines the properties for the Menu component.
// Menu is an organism component that displays a navigable list of items.
type MenuProps struct {
	// Items is the list of menu items to display.
	// Required - should not be empty.
	Items []MenuItem

	// Selected is the reactive reference to the currently selected item value.
	// Optional - if nil, no item is selected initially.
	Selected *bubbly.Ref[string]

	// OnSelect is called when a menu item is selected.
	// Receives the selected item's value.
	// Optional - if nil, no callback is executed.
	OnSelect func(string)

	// Width sets the menu width in characters.
	// Default is 30 if not specified.
	Width int

	// CommonProps for styling and identification.
	CommonProps
}

// Menu creates a menu navigation component.
// The menu displays a list of selectable items with keyboard navigation support.
//
// Features:
//   - List of menu items with labels
//   - Selected item highlighting
//   - Disabled item support
//   - Reactive selection state
//   - OnSelect callback
//   - Theme integration
//   - Custom style override
//
// Example:
//
//	selected := bubbly.NewRef("")
//	menu := Menu(MenuProps{
//	    Items: []MenuItem{
//	        {Label: "Home", Value: "home"},
//	        {Label: "Settings", Value: "settings"},
//	        {Label: "Logout", Value: "logout"},
//	    },
//	    Selected: selected,
//	    OnSelect: func(value string) {
//	        navigate(value)
//	    },
//	})
func Menu(props MenuProps) bubbly.Component {
	// Set defaults
	if props.Width == 0 {
		props.Width = 30
	}

	component, _ := bubbly.NewComponent("Menu").
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

			// Handle select event
			ctx.On("select", func(data interface{}) {
				value := data.(string)
				if props.Selected != nil {
					props.Selected.Set(value)
				}
				if props.OnSelect != nil {
					props.OnSelect(value)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(MenuProps)
			theme := ctx.Get("theme").(Theme)

			var content strings.Builder

			// Get current selection
			var selectedValue string
			if p.Selected != nil {
				selectedValue = p.Selected.GetTyped()
			}

			// Render each menu item
			for i, item := range p.Items {
				isSelected := item.Value == selectedValue

				// Build item style
				itemStyle := lipgloss.NewStyle().
					Width(p.Width-2).
					Padding(0, 1)

				if item.Disabled {
					// Disabled state
					itemStyle = itemStyle.Foreground(theme.Muted)
				} else if isSelected {
					// Selected state
					itemStyle = itemStyle.
						Foreground(lipgloss.Color("230")).
						Background(theme.Primary).
						Bold(true)
				} else {
					// Normal state
					itemStyle = itemStyle.Foreground(theme.Foreground)
				}

				// Add selection indicator
				indicator := "  "
				if isSelected {
					indicator = "▶ "
				}

				content.WriteString(itemStyle.Render(indicator + item.Label))
				if i < len(p.Items)-1 {
					content.WriteString("\n")
				}
			}

			// Create menu container style
			menuStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Secondary).
				Width(p.Width)

			// Apply custom style if provided
			if p.Style != nil {
				menuStyle = menuStyle.Inherit(*p.Style)
			}

			return menuStyle.Render(content.String())
		}).
		Build()

	return component
}
