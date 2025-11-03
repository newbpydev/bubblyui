package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Tab represents a single tab with label and content.
type Tab struct {
	// Label is the tab button text.
	Label string

	// Content is the content to display when tab is active.
	// Can be a string or use a Component for dynamic content.
	Content string

	// Component is an optional component to render as content.
	// If provided, takes precedence over Content string.
	Component bubbly.Component
}

// TabsProps defines the properties for the Tabs component.
// Tabs is an organism component that displays tabbed interface.
type TabsProps struct {
	// Tabs is the list of tabs to display.
	// Required - should not be empty.
	Tabs []Tab

	// ActiveIndex is the reactive reference to the currently active tab index.
	// Optional - defaults to 0 (first tab).
	ActiveIndex *bubbly.Ref[int]

	// OnTabChange is called when the active tab changes.
	// Receives the new tab index.
	// Optional - if nil, no callback is executed.
	OnTabChange func(int)

	// Width sets the tabs container width in characters.
	// Default is 60 if not specified.
	Width int

	// CommonProps for styling and identification.
	CommonProps
}

// Tabs creates a tabbed interface component.
// The tabs component displays multiple content panels with tab buttons for switching.
//
// Features:
//   - Multiple tabs with labels
//   - Active tab highlighting
//   - Reactive active index
//   - OnTabChange callback
//   - String or Component content
//   - Theme integration
//   - Custom style override
//
// Example:
//
//	activeIndex := bubbly.NewRef(0)
//	tabs := Tabs(TabsProps{
//	    Tabs: []Tab{
//	        {Label: "Profile", Content: "User profile content"},
//	        {Label: "Settings", Content: "Settings content"},
//	        {Label: "Security", Content: "Security content"},
//	    },
//	    ActiveIndex: activeIndex,
//	    OnTabChange: func(index int) {
//	        loadTabContent(index)
//	    },
//	})
func Tabs(props TabsProps) bubbly.Component {
	// Set defaults
	if props.Width == 0 {
		props.Width = 60
	}

	component, _ := bubbly.NewComponent("Tabs").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme using helper
			setupTheme(ctx)

			// Handle tab change event
			ctx.On("changeTab", func(data interface{}) {
				index := data.(int)
				if props.ActiveIndex != nil {
					props.ActiveIndex.Set(index)
				}
				if props.OnTabChange != nil {
					props.OnTabChange(index)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(TabsProps)
			theme := ctx.Get("theme").(Theme)

			if len(p.Tabs) == 0 {
				return ""
			}

			// Get active index
			activeIndex := 0
			if p.ActiveIndex != nil {
				activeIndex = p.ActiveIndex.GetTyped()
			}
			// Bounds check
			if activeIndex < 0 || activeIndex >= len(p.Tabs) {
				activeIndex = 0
			}

			var content strings.Builder

			// Render tab buttons
			var tabButtons []string
			for i, tab := range p.Tabs {
				isActive := i == activeIndex

				buttonStyle := lipgloss.NewStyle().
					Padding(0, 2)

				if isActive {
					// Active tab
					buttonStyle = buttonStyle.
						Foreground(lipgloss.Color("230")).
						Background(theme.Primary).
						Bold(true)
				} else {
					// Inactive tab
					buttonStyle = buttonStyle.
						Foreground(theme.Foreground).
						Background(theme.Muted)
				}

				tabButtons = append(tabButtons, buttonStyle.Render(tab.Label))
			}

			// Join tab buttons horizontally
			tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabButtons...)
			content.WriteString(tabBar)
			content.WriteString("\n\n")

			// Render active tab content
			activeTab := p.Tabs[activeIndex]
			contentStyle := lipgloss.NewStyle().
				Width(p.Width-4).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Secondary)

			var tabContent string
			if activeTab.Component != nil {
				tabContent = activeTab.Component.View()
			} else {
				tabContent = activeTab.Content
			}

			content.WriteString(contentStyle.Render(tabContent))

			// Apply custom style if provided
			if p.Style != nil {
				// Custom style for container
				containerStyle := lipgloss.NewStyle().Inherit(*p.Style)
				return containerStyle.Render(content.String())
			}

			return content.String()
		}).
		Build()

	return component
}
