package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// Model holds the application state
type model struct {
	router *router.Router
	width  int
	height int
}

func main() {
	// Create nested routes: Dashboard with child routes
	var r *router.Router

	r, err := router.NewRouterBuilder().
		RouteWithOptions("/",
			router.WithName("home"),
			router.WithComponent(createHomeComponent()),
		).
		RouteWithOptions("/dashboard",
			router.WithName("dashboard"),
			router.WithComponent(createDashboardLayoutFactory(&r)),
			router.WithChildren(
				&router.RouteRecord{
					Path:      "stats",
					Name:      "dashboard-stats",
					Component: createStatsComponent(),
				},
				&router.RouteRecord{
					Path:      "settings",
					Name:      "dashboard-settings",
					Component: createSettingsComponent(),
				},
				&router.RouteRecord{
					Path:      "profile",
					Name:      "dashboard-profile",
					Component: createProfileComponent(),
				},
			),
		).
		Build()

	if err != nil {
		fmt.Printf("Error creating router: %v\n", err)
		os.Exit(1)
	}

	m := model{
		router: r,
		width:  80,
		height: 24,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	// Navigate to home on start
	return m.router.Push(&router.NavigationTarget{Path: "/"})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// Navigation shortcuts
		case "1":
			return m, m.router.Push(&router.NavigationTarget{Path: "/"})
		case "2":
			return m, m.router.Push(&router.NavigationTarget{Path: "/dashboard/stats"})
		case "3":
			return m, m.router.Push(&router.NavigationTarget{Path: "/dashboard/settings"})
		case "4":
			return m, m.router.Push(&router.NavigationTarget{Path: "/dashboard/profile"})

		case "b":
			return m, m.router.Back()
		case "f":
			return m, m.router.Forward()
		}

	case router.RouteChangedMsg:
		return m, nil

	case router.NavigationErrorMsg:
		fmt.Printf("Navigation error: %v\n", msg.Error)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	route := m.router.CurrentRoute()
	if route == nil {
		return "Loading..."
	}

	// Create View for root level
	routerView := router.NewRouterView(m.router, 0)

	// Render header
	header := renderHeader(route)

	// Render navigation menu
	nav := renderNavigation()

	// Render component content
	content := routerView.View()
	if content == "" {
		content = "No component for route: " + route.Path
	}

	// Render footer
	footer := renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		nav,
		"",
		content,
		"",
		footer,
	)
}

func renderHeader(route *router.Route) string {
	// Show breadcrumbs from matched routes
	breadcrumbs := ""
	for i, matched := range route.Matched {
		if i > 0 {
			breadcrumbs += " > "
		}
		breadcrumbs += matched.Name
	}

	badge := components.Badge(components.BadgeProps{
		Label:   breadcrumbs,
		Variant: components.VariantPrimary,
	})
	badge.Init()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Padding(1, 2)

	title := titleStyle.Render("🏗️  BubblyUI Router - Nested Routes Example")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		badge.View(),
	)
}

func renderNavigation() string {
	navStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 2)

	navText := "Navigation: [1] Home  [2] Dashboard/Stats  [3] Dashboard/Settings  [4] Dashboard/Profile"

	return navStyle.Render(navText)
}

func renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 2)

	return footerStyle.Render("Press 'q' or Ctrl+C to quit  •  [b] Back  [f] Forward")
}

// Component creators

func createHomeComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Home").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := "Welcome to the Nested Routes Example!\n\nThis demonstrates:\n  • Parent-child route relationships\n  • Nested component rendering\n  • Breadcrumb navigation\n  • Layout composition\n\nNavigate to the Dashboard (keys 2-4) to see nested routes in action.\nThe Dashboard layout stays constant while child views change."

			card := components.Card(components.CardProps{
				Title:   "🏠 Home",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createDashboardLayoutFactory(routerPtr **router.Router) bubbly.Component {
	comp, _ := bubbly.NewComponent("DashboardLayout").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("35")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("35"))

			header := headerStyle.Render("📊 Dashboard Layout")

			sidebarStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(20)

			sidebar := sidebarStyle.Render("Sidebar\n\n• Stats\n• Settings\n• Profile")

			contentStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(50)

			// Render child route at depth 1
			var childContent string
			if *routerPtr != nil {
				childRouter := router.NewRouterView(*routerPtr, 1)
				childContent = childRouter.View()
			}

			if childContent == "" {
				childContent = "No child route selected.\nNavigate to Stats, Settings, or Profile."
			}

			content := contentStyle.Render(childContent)

			main := lipgloss.JoinHorizontal(
				lipgloss.Top,
				sidebar,
				"  ",
				content,
			)

			return lipgloss.JoinVertical(
				lipgloss.Top,
				header,
				"",
				main,
			)
		}).
		Build()

	return comp
}

func createStatsComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Stats").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := "Statistics Dashboard\n\nUsers: 1,234\nActive Sessions: 56\nRequests/sec: 89\n\nThis is a child route of /dashboard.\nThe parent layout (Dashboard) stays constant\nwhile this child view is rendered within it."

			card := components.Card(components.CardProps{
				Title:   "📈 Statistics",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createSettingsComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Settings").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := "Application Settings\n\n• Theme: Dark\n• Language: English\n• Notifications: Enabled\n• Auto-save: On\n\nThis is another child route of /dashboard.\nNotice how the breadcrumb shows the full path."

			card := components.Card(components.CardProps{
				Title:   "⚙️  Settings",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createProfileComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Profile").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := "User Profile\n\nName: John Doe\nEmail: john@example.com\nRole: Administrator\nMember since: 2024\n\nThis is the third child route of /dashboard.\nAll three children share the same parent layout."

			card := components.Card(components.CardProps{
				Title:   "👤 Profile",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}
