package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// AccordionItem represents a single accordion panel.
type AccordionItem struct {
	// Title is the panel header text.
	Title string

	// Content is the panel body text.
	Content string

	// Component is an optional component to render as content.
	// If provided, takes precedence over Content string.
	Component bubbly.Component
}

// AccordionProps defines the properties for the Accordion component.
// Accordion is an organism component that displays collapsible panels.
type AccordionProps struct {
	// Items is the list of accordion panels.
	// Required - should not be empty.
	Items []AccordionItem

	// ExpandedIndexes is the reactive reference to the list of expanded panel indexes.
	// Optional - if nil, no panels are expanded initially.
	ExpandedIndexes *bubbly.Ref[[]int]

	// AllowMultiple allows multiple panels to be expanded simultaneously.
	// If false, expanding a panel collapses others.
	// Default is false.
	AllowMultiple bool

	// OnToggle is called when a panel is toggled.
	// Receives the panel index and new expanded state.
	// Optional - if nil, no callback is executed.
	OnToggle func(int, bool)

	// Width sets the accordion width in characters.
	// Default is 50 if not specified.
	Width int

	// CommonProps for styling and identification.
	CommonProps
}

// Accordion creates an accordion collapsible panels component.
// The accordion displays a list of panels that can be expanded/collapsed.
//
// Features:
//   - Multiple collapsible panels
//   - Expand/collapse functionality
//   - Single or multiple expansion modes
//   - Reactive expanded state
//   - OnToggle callback
//   - String or Component content
//   - Theme integration
//   - Custom style override
//
// Example:
//
//	expanded := bubbly.NewRef([]int{0})
//	accordion := Accordion(AccordionProps{
//	    Items: []AccordionItem{
//	        {Title: "Section 1", Content: "Content for section 1"},
//	        {Title: "Section 2", Content: "Content for section 2"},
//	        {Title: "Section 3", Content: "Content for section 3"},
//	    },
//	    ExpandedIndexes: expanded,
//	    AllowMultiple:   true,
//	})
func Accordion(props AccordionProps) bubbly.Component {
	// Set defaults
	if props.Width == 0 {
		props.Width = 50
	}

	component, _ := bubbly.NewComponent("Accordion").
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

			// Handle toggle event
			ctx.On("toggle", func(data interface{}) {
				index := data.(int)
				if props.ExpandedIndexes != nil {
					expanded := props.ExpandedIndexes.GetTyped()

					// Check if index is in expanded list
					isExpanded := false
					newExpanded := []int{}
					for _, idx := range expanded {
						if idx == index {
							isExpanded = true
						} else {
							if props.AllowMultiple {
								newExpanded = append(newExpanded, idx)
							}
						}
					}

					// Toggle expansion
					if !isExpanded {
						newExpanded = append(newExpanded, index)
					}

					props.ExpandedIndexes.Set(newExpanded)

					if props.OnToggle != nil {
						props.OnToggle(index, !isExpanded)
					}
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(AccordionProps)
			theme := ctx.Get("theme").(Theme)

			if len(p.Items) == 0 {
				return ""
			}

			// Get expanded indexes
			var expandedIndexes []int
			if p.ExpandedIndexes != nil {
				expandedIndexes = p.ExpandedIndexes.GetTyped()
			}

			// Helper to check if index is expanded
			isExpanded := func(index int) bool {
				for _, idx := range expandedIndexes {
					if idx == index {
						return true
					}
				}
				return false
			}

			var content strings.Builder

			// Render each accordion item
			for i, item := range p.Items {
				expanded := isExpanded(i)

				// Render header
				headerStyle := lipgloss.NewStyle().
					Width(p.Width-4).
					Padding(0, 1).
					Bold(true).
					Foreground(theme.Primary)

				indicator := "▶"
				if expanded {
					indicator = "▼"
				}

				content.WriteString(headerStyle.Render(indicator + " " + item.Title))
				content.WriteString("\n")

				// Render content if expanded
				if expanded {
					contentStyle := lipgloss.NewStyle().
						Width(p.Width-6).
						Padding(0, 2).
						Foreground(theme.Foreground)

					var panelContent string
					if item.Component != nil {
						panelContent = item.Component.View()
					} else {
						panelContent = item.Content
					}

					content.WriteString(contentStyle.Render(panelContent))
					content.WriteString("\n")
				}

				// Add separator between items
				if i < len(p.Items)-1 {
					separatorStyle := lipgloss.NewStyle().
						Width(p.Width - 4).
						Foreground(theme.Muted)
					content.WriteString(separatorStyle.Render(strings.Repeat("─", p.Width-4)))
					content.WriteString("\n")
				}
			}

			// Create accordion container style
			accordionStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Secondary).
				Width(p.Width).
				Padding(1)

			// Apply custom style if provided
			if p.Style != nil {
				accordionStyle = accordionStyle.Inherit(*p.Style)
			}

			return accordionStyle.Render(content.String())
		}).
		Build()

	return component
}
