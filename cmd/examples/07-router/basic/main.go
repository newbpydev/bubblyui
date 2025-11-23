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
	// Create router with routes and components
	// Note: We'll use a factory pattern for the user detail component
	var r *router.Router

	r, err := router.NewRouterBuilder().
		RouteWithOptions("/",
			router.WithName("home"),
			router.WithComponent(createHomeComponent()),
		).
		RouteWithOptions("/about",
			router.WithName("about"),
			router.WithComponent(createAboutComponent()),
		).
		RouteWithOptions("/contact",
			router.WithName("contact"),
			router.WithComponent(createContactComponent()),
		).
		RouteWithOptions("/user/:id",
			router.WithName("user-detail"),
			router.WithComponent(createUserDetailComponentFactory(&r)),
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
			return m, m.router.Push(&router.NavigationTarget{Path: "/about"})
		case "3":
			return m, m.router.Push(&router.NavigationTarget{Path: "/contact"})
		case "4":
			// Navigate to user detail with ID parameter
			return m, m.router.Push(&router.NavigationTarget{Path: "/user/123"})
		case "5":
			// Navigate to different user
			return m, m.router.Push(&router.NavigationTarget{Path: "/user/456"})

		case "b":
			// Go back in history
			return m, m.router.Back()
		case "f":
			// Go forward in history
			return m, m.router.Forward()
		}

	case router.RouteChangedMsg:
		// Route changed successfully - trigger re-render
		return m, nil

	case router.NavigationErrorMsg:
		// Navigation failed - could show error message
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
	// Get current route
	route := m.router.CurrentRoute()
	if route == nil {
		return "Loading..."
	}

	// Create View to render current component
	routerView := router.NewRouterView(m.router, 0)

	// Render header with current route badge
	header := renderHeader(route)

	// Render navigation menu
	nav := renderNavigation()

	// Render component content via View
	content := routerView.View()
	if content == "" {
		content = "No component for route: " + route.Path
	}

	// Render footer with help
	footer := renderFooter()

	// Compose layout
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
	// Create current route badge using Badge component
	badge := components.Badge(components.BadgeProps{
		Label:   "Current Route: " + route.Path,
		Variant: components.VariantPrimary,
	})
	badge.Init()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Padding(1, 2)

	title := titleStyle.Render("🧭 BubblyUI Router - Basic Navigation Example")

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

	navText := "Navigation: [1] Home  [2] About  [3] Contact  [4] User 123  [5] User 456  [b] Back  [f] Forward"

	return navStyle.Render(navText)
}

func renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 2)

	return footerStyle.Render("Press 'q' or Ctrl+C to quit")
}

// Component creators

func createHomeComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Home").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Create card with home content
			card := components.Card(components.CardProps{
				Title:   "🏠 Home",
				Content: "Welcome to the BubblyUI Router example!\n\nThis demonstrates basic navigation between routes.\n\nUse the number keys (1-5) to navigate between different screens.\nUse 'b' to go back and 'f' to go forward in history.",
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createAboutComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("About").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			card := components.Card(components.CardProps{
				Title:   "ℹ️  About",
				Content: "BubblyUI Router System\n\nVersion: 1.0.0\nFeatures:\n  • Declarative route configuration\n  • Path parameters and query strings\n  • Navigation guards\n  • History management\n  • Nested routes\n\nBuilt with ❤️  using Bubbletea and Lipgloss",
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createContactComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Contact").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			card := components.Card(components.CardProps{
				Title:   "📧 Contact",
				Content: "Get in Touch\n\nEmail: hello@bubblyui.dev\nGitHub: github.com/newbpydev/bubblyui\nDiscord: discord.gg/bubblyui\n\nWe'd love to hear from you!",
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createUserDetailComponentFactory(routerPtr **router.Router) bubbly.Component {
	comp, _ := bubbly.NewComponent("UserDetail").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get the current route from the router
			var userID string
			if *routerPtr != nil {
				route := (*routerPtr).CurrentRoute()
				if route != nil && route.Params != nil {
					userID = route.Params["id"]
				}
			}

			if userID == "" {
				userID = "[unknown]"
			}

			// Create user detail content with actual ID
			content := fmt.Sprintf("User Profile\n\nUser ID: %s\n\nThis demonstrates dynamic route parameters.\nThe :id parameter in the route path /user/:id\nis extracted and displayed here.\n\nTry navigating to different users using keys 4 and 5!", userID)

			card := components.Card(components.CardProps{
				Title:   "👤 User Detail",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}
