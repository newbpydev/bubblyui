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

// Global auth state (in real app, this would be in a proper state manager)
var isAuthenticated = false
var username = ""

// Model holds the application state
type model struct {
	router        *router.Router
	width         int
	height        int
	inputMode     bool // Track if we're in input mode for login form
	usernameInput string
	passwordInput string
	activeField   int // 0 = username, 1 = password
}

func main() {
	// Create authentication guard
	authGuard := func(to, from *router.Route, next router.NextFunc) {
		// Check if route requires authentication
		requiresAuth, ok := to.Meta["requiresAuth"].(bool)

		if ok && requiresAuth && !isAuthenticated {
			// Redirect to login, save intended destination
			next(&router.NavigationTarget{
				Path: "/login",
				Query: map[string]string{
					"redirect": to.FullPath,
				},
			})
		} else {
			// Allow navigation
			next(nil)
		}
	}

	// Create router with routes, components, and guards
	r, err := router.NewRouterBuilder().
		RouteWithOptions("/",
			router.WithName("home"),
			router.WithComponent(createHomeComponent()),
		).
		RouteWithOptions("/login",
			router.WithName("login"),
			router.WithComponent(createLoginComponent()),
		).
		RouteWithOptions("/dashboard",
			router.WithName("dashboard"),
			router.WithComponent(createDashboardComponent()),
			router.WithMeta(map[string]interface{}{
				"requiresAuth": true,
			}),
		).
		RouteWithOptions("/profile",
			router.WithName("profile"),
			router.WithComponent(createProfileComponent()),
			router.WithMeta(map[string]interface{}{
				"requiresAuth": true,
			}),
		).
		BeforeEach(authGuard).
		Build()

	if err != nil {
		fmt.Printf("Error creating router: %v\n", err)
		os.Exit(1)
	}

	m := model{
		router:        r,
		width:         80,
		height:        24,
		inputMode:     false,
		usernameInput: "",
		passwordInput: "",
		activeField:   0,
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
		// Handle input mode for login form
		if m.inputMode {
			return m.handleInputMode(msg)
		}

		// Navigation mode
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// Navigation shortcuts
		case "1":
			return m, m.router.Push(&router.NavigationTarget{Path: "/"})
		case "2":
			// Try to access protected dashboard
			return m, m.router.Push(&router.NavigationTarget{Path: "/dashboard"})
		case "3":
			// Try to access protected profile
			return m, m.router.Push(&router.NavigationTarget{Path: "/profile"})
		case "4":
			// Go to login
			return m, m.router.Push(&router.NavigationTarget{Path: "/login"})
		case "5":
			// Logout
			isAuthenticated = false
			username = ""
			return m, m.router.Push(&router.NavigationTarget{Path: "/"})

		case "enter":
			// Enter input mode if on login page
			route := m.router.CurrentRoute()
			if route != nil && route.Path == "/login" {
				m.inputMode = true
			}
			return m, nil

		case "b":
			return m, m.router.Back()
		case "f":
			return m, m.router.Forward()
		}

	case router.RouteChangedMsg:
		// Route changed successfully
		return m, nil

	case router.NavigationErrorMsg:
		// Navigation failed
		fmt.Printf("Navigation error: %v\n", msg.Error)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit input mode
		m.inputMode = false
		return m, nil

	case "tab":
		// Switch between fields
		m.activeField = (m.activeField + 1) % 2
		return m, nil

	case "enter":
		// Submit login
		if m.usernameInput != "" && m.passwordInput != "" {
			// Simple auth: any non-empty username/password works
			isAuthenticated = true
			username = m.usernameInput

			// Clear form
			m.usernameInput = ""
			m.passwordInput = ""
			m.inputMode = false

			// Check for redirect query param
			route := m.router.CurrentRoute()
			redirectPath := "/"
			if route != nil && route.Query != nil {
				if redirect, ok := route.Query["redirect"]; ok {
					redirectPath = redirect
				}
			}

			return m, m.router.Push(&router.NavigationTarget{Path: redirectPath})
		}
		return m, nil

	case "backspace":
		// Delete character
		if m.activeField == 0 && len(m.usernameInput) > 0 {
			m.usernameInput = m.usernameInput[:len(m.usernameInput)-1]
		} else if m.activeField == 1 && len(m.passwordInput) > 0 {
			m.passwordInput = m.passwordInput[:len(m.passwordInput)-1]
		}
		return m, nil

	default:
		// Add character to active field
		if len(msg.Runes) > 0 {
			char := string(msg.Runes[0])
			if m.activeField == 0 {
				m.usernameInput += char
			} else {
				m.passwordInput += char
			}
		}
		return m, nil
	}
}

func (m model) View() string {
	// Get current route
	route := m.router.CurrentRoute()
	if route == nil {
		return "Loading..."
	}

	// Create RouterView to render current component
	routerView := router.NewRouterView(m.router, 0)

	// Render header with auth status
	header := renderHeader(route)

	// Render navigation menu
	nav := renderNavigation()

	// Render component content
	content := routerView.View()
	if content == "" {
		content = "No component for route: " + route.Path
	}

	// If on login page and in input mode, show login form
	if route.Path == "/login" && m.inputMode {
		content = m.renderLoginForm()
	}

	// Render footer with help
	footer := renderFooter(m.inputMode)

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
	// Create auth status badge
	authLabel := "🔒 Not Authenticated"
	authVariant := components.VariantDanger
	if isAuthenticated {
		authLabel = "🔓 Authenticated as " + username
		authVariant = components.VariantSuccess
	}

	authBadge := components.Badge(components.BadgeProps{
		Label:   authLabel,
		Variant: authVariant,
	})
	authBadge.Init()

	// Create current route badge
	routeBadge := components.Badge(components.BadgeProps{
		Label:   "Current: " + route.Path,
		Variant: components.VariantPrimary,
	})
	routeBadge.Init()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Padding(1, 2)

	title := titleStyle.Render("🛡️  BubblyUI Router - Navigation Guards Example")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		lipgloss.JoinHorizontal(lipgloss.Left, authBadge.View(), "  ", routeBadge.View()),
	)
}

func renderNavigation() string {
	navStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 2)

	navText := "Navigation: [1] Home  [2] Dashboard (protected)  [3] Profile (protected)  [4] Login  [5] Logout"

	return navStyle.Render(navText)
}

func renderFooter(inputMode bool) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 2)

	helpText := "Press 'q' or Ctrl+C to quit  •  [b] Back  [f] Forward"
	if inputMode {
		helpText = "INPUT MODE: Type to enter text  •  [Tab] Switch field  •  [Enter] Submit  •  [ESC] Cancel"
	}

	return footerStyle.Render(helpText)
}

func (m model) renderLoginForm() string {
	// Create form UI
	usernameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("35")).
		Padding(0, 1).
		Width(30)

	passwordStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(30)

	if m.activeField == 1 {
		// Password field is active
		usernameStyle = usernameStyle.BorderForeground(lipgloss.Color("240"))
		passwordStyle = passwordStyle.BorderForeground(lipgloss.Color("35"))
	}

	usernameField := usernameStyle.Render(fmt.Sprintf("Username: %s█", m.usernameInput))
	passwordField := passwordStyle.Render(fmt.Sprintf("Password: %s█", m.passwordInput))

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(2, 4)

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		"Login Form",
		"",
		usernameField,
		"",
		passwordField,
		"",
		"[Tab] Switch field  •  [Enter] Submit  •  [ESC] Cancel",
	)

	return formStyle.Render(form)
}

// Component creators

func createHomeComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Home").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := "Welcome to the Navigation Guards Example!\n\nThis demonstrates:\n  • Authentication guards\n  • Protected routes\n  • Login flow with redirect\n  • Auth state management\n\nTry accessing the Dashboard (key 2) without logging in.\nYou'll be redirected to the login page.\n\nAfter logging in, you'll be redirected back to the Dashboard."

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

func createLoginComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Login").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := "Login Required\n\nPress [Enter] to activate the login form.\n\nIn the form:\n  • Type your username and password\n  • Use [Tab] to switch between fields\n  • Press [Enter] to submit\n  • Press [ESC] to cancel\n\nNote: Any non-empty username/password will work for this demo."

			card := components.Card(components.CardProps{
				Title:   "🔐 Login",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}

func createDashboardComponent() bubbly.Component {
	comp, _ := bubbly.NewComponent("Dashboard").
		Setup(func(ctx *bubbly.Context) {
			ctx.Provide("theme", components.DefaultTheme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			content := fmt.Sprintf("Welcome, %s!\n\nYou successfully accessed a protected route.\n\nThe authentication guard checked your auth status\nand allowed navigation to this page.\n\nTry logging out (key 5) and accessing this page again.\nYou'll be redirected to the login page.", username)

			card := components.Card(components.CardProps{
				Title:   "📊 Dashboard (Protected)",
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
			content := fmt.Sprintf("User Profile\n\nUsername: %s\nStatus: Authenticated\nRole: User\n\nThis is another protected route that requires authentication.\n\nThe same guard protects both Dashboard and Profile routes.", username)

			card := components.Card(components.CardProps{
				Title:   "👤 Profile (Protected)",
				Content: content,
			})
			card.Init()

			return card.View()
		}).
		Build()

	return comp
}
