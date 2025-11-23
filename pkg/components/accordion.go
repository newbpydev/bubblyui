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

// accordionApplyDefaults sets default values for AccordionProps.
func accordionApplyDefaults(props *AccordionProps) {
	if props.Width == 0 {
		props.Width = 50
	}
}

// accordionContainsIndex checks if an index is in the expanded list.
func accordionContainsIndex(indexes []int, target int) bool {
	for _, idx := range indexes {
		if idx == target {
			return true
		}
	}
	return false
}

// accordionToggleExpanded handles toggling an accordion item's expanded state.
func accordionToggleExpanded(props AccordionProps, index int) {
	if props.ExpandedIndexes == nil {
		return
	}

	expanded := props.ExpandedIndexes.GetTyped()
	isExpanded := accordionContainsIndex(expanded, index)

	var newExpanded []int
	if props.AllowMultiple {
		for _, idx := range expanded {
			if idx != index {
				newExpanded = append(newExpanded, idx)
			}
		}
	}

	if !isExpanded {
		newExpanded = append(newExpanded, index)
	}

	props.ExpandedIndexes.Set(newExpanded)

	if props.OnToggle != nil {
		props.OnToggle(index, !isExpanded)
	}
}

// accordionRenderItem renders a single accordion item.
func accordionRenderItem(item AccordionItem, _ int, isExpanded bool, width int, theme Theme, isLast bool) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Width(width - 4).
		Padding(0, 1).
		Bold(true).
		Foreground(theme.Primary)

	indicator := "▶"
	if isExpanded {
		indicator = "▼"
	}
	content.WriteString(headerStyle.Render(indicator + " " + item.Title))
	content.WriteString("\n")

	if isExpanded {
		contentStyle := lipgloss.NewStyle().
			Width(width - 6).
			Padding(0, 2).
			Foreground(theme.Foreground)

		panelContent := item.Content
		if item.Component != nil {
			panelContent = item.Component.View()
		}
		content.WriteString(contentStyle.Render(panelContent))
		content.WriteString("\n")
	}

	if !isLast {
		separatorStyle := lipgloss.NewStyle().
			Width(width - 4).
			Foreground(theme.Muted)
		content.WriteString(separatorStyle.Render(strings.Repeat("─", width-4)))
		content.WriteString("\n")
	}

	return content.String()
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
	accordionApplyDefaults(&props)

	component, _ := bubbly.NewComponent("Accordion").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := injectTheme(ctx)
			ctx.Expose("theme", theme)

			ctx.On("toggle", func(data interface{}) {
				accordionToggleExpanded(props, data.(int))
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(AccordionProps)
			theme := ctx.Get("theme").(Theme)

			if len(p.Items) == 0 {
				return ""
			}

			var expandedIndexes []int
			if p.ExpandedIndexes != nil {
				expandedIndexes = p.ExpandedIndexes.GetTyped()
			}

			var content strings.Builder
			for i, item := range p.Items {
				isExpanded := accordionContainsIndex(expandedIndexes, i)
				isLast := i == len(p.Items)-1
				content.WriteString(accordionRenderItem(item, i, isExpanded, p.Width, theme, isLast))
			}

			accordionStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Secondary).
				Width(p.Width).
				Padding(1)

			if p.Style != nil {
				accordionStyle = accordionStyle.Inherit(*p.Style)
			}

			return accordionStyle.Render(content.String())
		}).
		Build()

	return component
}
