package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// Metric represents a dashboard metric
type Metric struct {
	Name  string
	Value int
	Unit  string
	Trend string // "up", "down", "stable"
}

// Event represents a recent event
type Event struct {
	Time    string
	Type    string
	Message string
}

// Server represents server data for the table
type Server struct {
	Name   string
	Status string
	CPU    int
	Memory int
	Uptime string
}

// tickMsg for real-time updates
type tickMsg time.Time

// model wraps the dashboard component
type model struct {
	component      bubbly.Component
	navigationMode bool // true = navigation, false = table selection
}

func (m model) Init() tea.Cmd {
	// Start the ticker for real-time updates
	return tea.Batch(
		m.component.Init(),
		tick(),
	)
}

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Update dashboard data
		m.component.Emit("updateData", nil)
		// Continue ticking
		cmds = append(cmds, tick())
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			// Switch between tabs
			m.component.Emit("nextTab", nil)
			return m, nil
		case "shift+tab":
			m.component.Emit("prevTab", nil)
			return m, nil
		case "n":
			// Toggle navigation mode
			m.navigationMode = !m.navigationMode
			m.component.Emit("setNavigationMode", m.navigationMode)
			return m, nil
		case "up", "down":
			// Navigate in table/list
			if !m.navigationMode {
				m.component.Emit("navigate", msg.String())
			}
			return m, nil
		case "r":
			// Manual refresh
			m.component.Emit("updateData", nil)
			return m, nil
		case "enter":
			// Select current item
			m.component.Emit("selectItem", nil)
			return m, nil
		}
	}

	// Update component
	updated, cmd := m.component.Update(msg)
	m.component = updated.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	modeStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)

	var modeIndicator string
	if m.navigationMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("33"))
		modeIndicator = modeStyle.Render("🧭 NAV MODE")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		modeIndicator = modeStyle.Render("📊 SELECT MODE")
	}

	title := titleStyle.Render("📊 BubblyUI Dashboard")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render("Real-time monitoring with data display components")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render(
		"tab/shift+tab: switch tabs • n: toggle mode • up/down: navigate • r: refresh • q: quit",
	)

	return fmt.Sprintf("%s  %s\n%s\n\n%s\n\n%s",
		title,
		modeIndicator,
		subtitle,
		m.component.View(),
		help,
	)
}

// generateRandomData generates random dashboard data
func generateRandomData() ([]Metric, []Event, []Server) {
	// Generate metrics
	metrics := []Metric{
		{
			Name:  "Active Users",
			Value: 1200 + rand.Intn(200),
			Unit:  "users",
			Trend: []string{"up", "down", "stable"}[rand.Intn(3)],
		},
		{
			Name:  "Response Time",
			Value: 80 + rand.Intn(40),
			Unit:  "ms",
			Trend: []string{"up", "down", "stable"}[rand.Intn(3)],
		},
		{
			Name:  "Requests/sec",
			Value: 3000 + rand.Intn(1000),
			Unit:  "req/s",
			Trend: []string{"up", "down", "stable"}[rand.Intn(3)],
		},
		{
			Name:  "Error Rate",
			Value: rand.Intn(5),
			Unit:  "%",
			Trend: []string{"up", "down", "stable"}[rand.Intn(3)],
		},
	}

	// Generate events
	eventTypes := []string{"INFO", "WARNING", "ERROR", "SUCCESS"}
	eventMessages := []string{
		"System health check completed",
		"Database backup started",
		"User authentication spike detected",
		"Cache cleared successfully",
		"API rate limit approaching",
		"Deployment completed",
	}

	events := make([]Event, 5)
	for i := 0; i < 5; i++ {
		events[i] = Event{
			Time:    time.Now().Add(-time.Duration(i*5) * time.Minute).Format("15:04:05"),
			Type:    eventTypes[rand.Intn(len(eventTypes))],
			Message: eventMessages[rand.Intn(len(eventMessages))],
		}
	}

	// Generate server data
	servers := []Server{
		{
			Name:   "web-01",
			Status: []string{"Online", "Online", "Maintenance"}[rand.Intn(3)],
			CPU:    20 + rand.Intn(60),
			Memory: 30 + rand.Intn(50),
			Uptime: fmt.Sprintf("%dd %dh", rand.Intn(30), rand.Intn(24)),
		},
		{
			Name:   "api-01",
			Status: []string{"Online", "Online", "Offline"}[rand.Intn(3)],
			CPU:    10 + rand.Intn(70),
			Memory: 20 + rand.Intn(60),
			Uptime: fmt.Sprintf("%dd %dh", rand.Intn(30), rand.Intn(24)),
		},
		{
			Name:   "db-01",
			Status: "Online",
			CPU:    30 + rand.Intn(40),
			Memory: 60 + rand.Intn(30),
			Uptime: fmt.Sprintf("%dd %dh", rand.Intn(90), rand.Intn(24)),
		},
		{
			Name:   "cache-01",
			Status: "Online",
			CPU:    5 + rand.Intn(20),
			Memory: 40 + rand.Intn(30),
			Uptime: fmt.Sprintf("%dd %dh", rand.Intn(60), rand.Intn(24)),
		},
	}

	return metrics, events, servers
}

// createDashboard creates the dashboard component
func createDashboard() (bubbly.Component, error) {
	return bubbly.NewComponent("Dashboard").
		Setup(func(ctx *bubbly.Context) {
			// Provide custom theme
			theme := components.Theme{
				Primary:    lipgloss.Color("205"),
				Secondary:  lipgloss.Color("33"),
				Success:    lipgloss.Color("35"),
				Danger:     lipgloss.Color("196"),
				Warning:    lipgloss.Color("220"),
				Foreground: lipgloss.Color("15"),
				Muted:      lipgloss.Color("241"),
				Background: lipgloss.Color("0"),
			}
			ctx.Provide("theme", theme)

			// Initialize data
			initialMetrics, initialEvents, initialServers := generateRandomData()

			// State management using typed refs
			activeTab := bubbly.NewRef(0)
			metrics := bubbly.NewRef(initialMetrics)
			events := bubbly.NewRef(initialEvents)
			servers := bubbly.NewRef(initialServers)
			selectedServer := bubbly.NewRef(0)
			lastUpdate := bubbly.NewRef(time.Now())
			navigationMode := bubbly.NewRef(false)

			// Expose state
			ctx.Expose("activeTab", activeTab)
			ctx.Expose("metrics", metrics)
			ctx.Expose("events", events)
			ctx.Expose("servers", servers)
			ctx.Expose("selectedServer", selectedServer)
			ctx.Expose("lastUpdate", lastUpdate)
			ctx.Expose("navigationMode", navigationMode)

			// Event handlers
			ctx.On("nextTab", func(_ interface{}) {
				current := activeTab.Get().(int)
				activeTab.Set((current + 1) % 3) // 3 tabs total
			})

			ctx.On("prevTab", func(_ interface{}) {
				current := activeTab.Get().(int)
				if current == 0 {
					activeTab.Set(2)
				} else {
					activeTab.Set(current - 1)
				}
			})

			ctx.On("updateData", func(_ interface{}) {
				// Generate new random data
				newMetrics, newEvents, newServers := generateRandomData()
				metrics.Set(newMetrics)
				events.Set(newEvents)
				servers.Set(newServers)
				lastUpdate.Set(time.Now())
			})

			ctx.On("navigate", func(data interface{}) {
				direction := data.(string)
				current := selectedServer.Get().(int)
				serverCount := len(servers.Get().([]Server))

				if direction == "down" {
					selectedServer.Set((current + 1) % serverCount)
				} else if direction == "up" {
					if current == 0 {
						selectedServer.Set(serverCount - 1)
					} else {
						selectedServer.Set(current - 1)
					}
				}
			})

			ctx.On("setNavigationMode", func(data interface{}) {
				navigationMode.Set(data.(bool))
			})

			ctx.On("selectItem", func(_ interface{}) {
				// Would handle item selection here
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			activeTab := ctx.Get("activeTab").(*bubbly.Ref[int]).Get().(int)
			metrics := ctx.Get("metrics").(*bubbly.Ref[[]Metric]).Get().([]Metric)
			events := ctx.Get("events").(*bubbly.Ref[[]Event]).Get().([]Event)
			servers := ctx.Get("servers").(*bubbly.Ref[[]Server]).Get().([]Server)
			selectedServer := ctx.Get("selectedServer").(*bubbly.Ref[int]).Get().(int)
			lastUpdate := ctx.Get("lastUpdate").(*bubbly.Ref[time.Time]).Get().(time.Time)
			navigationMode := ctx.Get("navigationMode").(*bubbly.Ref[bool]).Get().(bool)

			// Tab headers
			tabStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.NormalBorder(), true, true, false, true)

			activeTabStyle := tabStyle.Copy().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				BorderForeground(lipgloss.Color("205"))

			inactiveTabStyle := tabStyle.Copy().
				Foreground(lipgloss.Color("241")).
				BorderForeground(lipgloss.Color("240"))

			tabs := []string{"Overview", "Servers", "Events"}
			var tabHeaders []string
			for i, tab := range tabs {
				if i == activeTab {
					tabHeaders = append(tabHeaders, activeTabStyle.Render(tab))
				} else {
					tabHeaders = append(tabHeaders, inactiveTabStyle.Render(tab))
				}
			}

			tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabHeaders...)

			// Content area
			contentStyle := lipgloss.NewStyle().
				Padding(1).
				Border(lipgloss.NormalBorder()).
				BorderTop(false).
				BorderForeground(func() lipgloss.Color {
					if navigationMode {
						return lipgloss.Color("33")
					}
					return lipgloss.Color("240")
				}()).
				Width(100).
				Height(25)

			var content string

			switch activeTab {
			case 0: // Overview tab
				content = renderOverviewTab(metrics, servers, lastUpdate)
			case 1: // Servers tab
				content = renderServersTab(servers, selectedServer, navigationMode)
			case 2: // Events tab
				content = renderEventsTab(events, lastUpdate)
			}

			return lipgloss.JoinVertical(
				lipgloss.Left,
				tabBar,
				contentStyle.Render(content),
			)
		}).
		Build()
}

func renderOverviewTab(metrics []Metric, servers []Server, lastUpdate time.Time) string {
	// Create metric cards using GridLayout
	var metricCards []bubbly.Component

	for _, metric := range metrics {
		// Determine color based on metric
		var valueColor lipgloss.Color
		if metric.Name == "Error Rate" {
			if metric.Value > 2 {
				valueColor = lipgloss.Color("196") // Red
			} else {
				valueColor = lipgloss.Color("35") // Green
			}
		} else {
			valueColor = lipgloss.Color("205") // Default
		}

		// Trend icon
		trendIcon := ""
		trendColor := lipgloss.Color("241")
		switch metric.Trend {
		case "up":
			trendIcon = "↑"
			trendColor = lipgloss.Color("35")
		case "down":
			trendIcon = "↓"
			trendColor = lipgloss.Color("196")
		case "stable":
			trendIcon = "→"
			trendColor = lipgloss.Color("241")
		}

		// Create card content
		cardContent := fmt.Sprintf("%s %s\n%s",
			lipgloss.NewStyle().
				Foreground(valueColor).
				Bold(true).
				Render(fmt.Sprintf("%d", metric.Value)),
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(metric.Unit),
			lipgloss.NewStyle().
				Foreground(trendColor).
				Render(trendIcon+" "+metric.Trend),
		)

		card := components.Card(components.CardProps{
			Title:   metric.Name,
			Content: cardContent,
			Width:   22,
		})
		card.Init()
		metricCards = append(metricCards, card)
	}

	// Create grid layout for metrics
	metricsGrid := components.GridLayout(components.GridLayoutProps{
		Columns: 4,
		Gap:     1,
		Items:   metricCards,
	})
	metricsGrid.Init()

	// Server status summary
	onlineCount := 0
	for _, server := range servers {
		if server.Status == "Online" {
			onlineCount++
		}
	}

	statusBadge := components.Badge(components.BadgeProps{
		Label: fmt.Sprintf("%d/%d Online", onlineCount, len(servers)),
		Variant: func() components.Variant {
			if onlineCount == len(servers) {
				return components.VariantSuccess
			} else if onlineCount > len(servers)/2 {
				return components.VariantWarning
			}
			return components.VariantDanger
		}(),
	})
	statusBadge.Init()

	// Last update time
	updateText := components.Text(components.TextProps{
		Content: fmt.Sprintf("Last updated: %s", lastUpdate.Format("15:04:05")),
		Color:   lipgloss.Color("241"),
	})
	updateText.Init()

	// Combine everything
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sectionTitle.Render("System Metrics"),
		metricsGrid.View(),
		"",
		sectionTitle.Render("Server Status"),
		statusBadge.View(),
		"",
		updateText.View(),
	)
}

func renderServersTab(servers []Server, selectedIndex int, navigationMode bool) string {
	// Create table for servers
	serversRef := bubbly.NewRef(servers)
	table := components.Table(components.TableProps[Server]{
		Data: serversRef,
		Columns: []components.TableColumn[Server]{
			{Header: "Server", Field: "Name", Width: 12},
			{Header: "Status", Field: "Status", Width: 12},
			{Header: "CPU %", Field: "CPU", Width: 8},
			{Header: "Memory %", Field: "Memory", Width: 10},
			{Header: "Uptime", Field: "Uptime", Width: 12},
		},
		Sortable: false,
		OnRowClick: func(s Server, index int) {
			// Handle row click
		},
	})
	table.Init()

	// Selected server details
	if selectedIndex < len(servers) {
		selected := servers[selectedIndex]

		// Create detail card
		detailContent := fmt.Sprintf(
			"Server: %s\nStatus: %s\nCPU Usage: %d%%\nMemory Usage: %d%%\nUptime: %s",
			selected.Name,
			selected.Status,
			selected.CPU,
			selected.Memory,
			selected.Uptime,
		)

		detailCard := components.Card(components.CardProps{
			Title:   "Selected Server Details",
			Content: detailContent,
			Width:   40,
		})
		detailCard.Init()

		// Create CPU gauge
		cpuBar := components.Text(components.TextProps{
			Content: fmt.Sprintf("CPU: %s %d%%",
				strings.Repeat("█", selected.CPU/10),
				selected.CPU),
			Color: func() lipgloss.Color {
				if selected.CPU > 80 {
					return lipgloss.Color("196")
				} else if selected.CPU > 60 {
					return lipgloss.Color("220")
				}
				return lipgloss.Color("35")
			}(),
		})
		cpuBar.Init()

		// Create Memory gauge
		memBar := components.Text(components.TextProps{
			Content: fmt.Sprintf("MEM: %s %d%%",
				strings.Repeat("█", selected.Memory/10),
				selected.Memory),
			Color: func() lipgloss.Color {
				if selected.Memory > 80 {
					return lipgloss.Color("196")
				} else if selected.Memory > 60 {
					return lipgloss.Color("220")
				}
				return lipgloss.Color("35")
			}(),
		})
		memBar.Init()

		// Navigation indicator
		navIndicator := ""
		if !navigationMode {
			navIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Render(fmt.Sprintf("▶ Selected: %s", selected.Name))
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			table.View(),
			"",
			navIndicator,
			"",
			detailCard.View(),
			"",
			cpuBar.View(),
			memBar.View(),
		)
	}

	return table.View()
}

func renderEventsTab(events []Event, lastUpdate time.Time) string {
	// Create list of events
	eventItems := make([]string, len(events))
	for i, event := range events {
		// Color code by type
		var typeColor lipgloss.Color
		var typeIcon string
		switch event.Type {
		case "ERROR":
			typeColor = lipgloss.Color("196")
			typeIcon = "❌"
		case "WARNING":
			typeColor = lipgloss.Color("220")
			typeIcon = "⚠️"
		case "SUCCESS":
			typeColor = lipgloss.Color("35")
			typeIcon = "✅"
		default:
			typeColor = lipgloss.Color("33")
			typeIcon = "ℹ️"
		}

		eventItems[i] = fmt.Sprintf("%s %s [%s] %s",
			typeIcon,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(event.Time),
			lipgloss.NewStyle().
				Foreground(typeColor).
				Bold(true).
				Render(event.Type),
			event.Message,
		)
	}

	// Create list component
	eventListRef := bubbly.NewRef(eventItems)
	eventList := components.List(components.ListProps[string]{
		Items: eventListRef,
		RenderItem: func(item string, index int) string {
			return item
		},
		Height: 10,
	})
	eventList.Init()

	// Stats card
	errorCount := 0
	warningCount := 0
	for _, event := range events {
		switch event.Type {
		case "ERROR":
			errorCount++
		case "WARNING":
			warningCount++
		}
	}

	statsContent := fmt.Sprintf(
		"Total Events: %d\nErrors: %d\nWarnings: %d\n\nLast Update: %s",
		len(events),
		errorCount,
		warningCount,
		lastUpdate.Format("15:04:05"),
	)

	statsCard := components.Card(components.CardProps{
		Title:   "Event Statistics",
		Content: statsContent,
		Width:   30,
	})
	statsCard.Init()

	// Layout using PanelLayout
	panelLayout := components.PanelLayout(components.PanelLayoutProps{
		Left:  eventList,
		Right: statsCard,
	})
	panelLayout.Init()

	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sectionTitle.Render("Recent Events"),
		"",
		panelLayout.View(),
	)
}

func main() {
	// Seed random for different data each run
	rand.Seed(time.Now().UnixNano())

	// Create the dashboard component
	dashboard, err := createDashboard()
	if err != nil {
		fmt.Printf("Error creating dashboard: %v\n", err)
		os.Exit(1)
	}

	// Create model
	m := model{
		component:      dashboard,
		navigationMode: false,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
