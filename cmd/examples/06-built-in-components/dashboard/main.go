package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// User represents a user in the system
type User struct {
	ID       int
	Name     string
	Email    string
	Status   string
	LastSeen string
}

// Stats represents dashboard statistics
type Stats struct {
	TotalUsers   int
	ActiveUsers  int
	PendingTasks int
	Completed    int
}

// tickMsg is sent periodically to update stats
type tickMsg time.Time

// model wraps the dashboard component
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	// Also start the ticker for live updates
	return tea.Batch(
		m.component.Init(),
		tick(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			// Refresh data
			m.component.Emit("refresh", nil)
		}
	case tickMsg:
		// Update stats periodically
		m.component.Emit("tick", nil)
		return m, tick()
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

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

	title := titleStyle.Render("📊 Dashboard - Built-in Components Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Table + Card components with live data",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render("r: refresh • q: quit")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// tick returns a command that sends a tickMsg after 2 seconds
func tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// createDashboard creates the dashboard component
func createDashboard() (bubbly.Component, error) {
	return bubbly.NewComponent("Dashboard").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for child components
			ctx.Provide("theme", components.DefaultTheme)

			// Sample user data
			users := bubbly.NewRef([]User{
				{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Status: "Active", LastSeen: "2 min ago"},
				{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Status: "Active", LastSeen: "5 min ago"},
				{ID: 3, Name: "Carol White", Email: "carol@example.com", Status: "Offline", LastSeen: "2 hours ago"},
				{ID: 4, Name: "David Brown", Email: "david@example.com", Status: "Active", LastSeen: "1 min ago"},
				{ID: 5, Name: "Eve Davis", Email: "eve@example.com", Status: "Away", LastSeen: "30 min ago"},
			})

			// Statistics
			stats := bubbly.NewRef(Stats{
				TotalUsers:   5,
				ActiveUsers:  3,
				PendingTasks: 12,
				Completed:    45,
			})

			// Computed values
			activePercentage := ctx.Computed(func() interface{} {
				s := stats.Get().(Stats)
				if s.TotalUsers == 0 {
					return 0
				}
				return (s.ActiveUsers * 100) / s.TotalUsers
			})

			completionRate := ctx.Computed(func() interface{} {
				s := stats.Get().(Stats)
				total := s.PendingTasks + s.Completed
				if total == 0 {
					return 0
				}
				return (s.Completed * 100) / total
			})

			// Create table component
			table := components.Table(components.TableProps[User]{
				Data: users,
				Columns: []components.TableColumn[User]{
					{Header: "ID", Field: "ID", Width: 5},
					{Header: "Name", Field: "Name", Width: 20},
					{Header: "Email", Field: "Email", Width: 25},
					{Header: "Status", Field: "Status", Width: 10},
					{Header: "Last Seen", Field: "LastSeen", Width: 15},
				},
				Sortable: true,
				OnRowClick: func(user User, index int) {
					// In a real app, this would show user details
				},
			})

			// Create stats cards
			totalUsersCard := components.Card(components.CardProps{
				Title:   "Total Users",
				Content: "",
				Width:   20,
				Padding: 1,
			})

			activeUsersCard := components.Card(components.CardProps{
				Title:   "Active Users",
				Content: "",
				Width:   20,
				Padding: 1,
			})

			pendingTasksCard := components.Card(components.CardProps{
				Title:   "Pending Tasks",
				Content: "",
				Width:   20,
				Padding: 1,
			})

			completedCard := components.Card(components.CardProps{
				Title:   "Completed",
				Content: "",
				Width:   20,
				Padding: 1,
			})

			// Initialize child components
			table.Init()
			totalUsersCard.Init()
			activeUsersCard.Init()
			pendingTasksCard.Init()
			completedCard.Init()

			// Expose state to template
			ctx.Expose("table", table)
			ctx.Expose("totalUsersCard", totalUsersCard)
			ctx.Expose("activeUsersCard", activeUsersCard)
			ctx.Expose("pendingTasksCard", pendingTasksCard)
			ctx.Expose("completedCard", completedCard)
			ctx.Expose("users", users)
			ctx.Expose("stats", stats)
			ctx.Expose("activePercentage", activePercentage)
			ctx.Expose("completionRate", completionRate)

			// Event: Refresh data
			ctx.On("refresh", func(_ interface{}) {
				// In a real app, this would fetch fresh data from an API
				// For demo, we'll just update the stats
				currentStats := stats.Get().(Stats)
				currentStats.ActiveUsers = (currentStats.ActiveUsers + 1) % (currentStats.TotalUsers + 1)
				currentStats.Completed++
				stats.Set(currentStats)
			})

			// Event: Periodic tick
			ctx.On("tick", func(_ interface{}) {
				// Simulate live updates
				currentStats := stats.Get().(Stats)
				// Randomly update pending tasks
				if currentStats.PendingTasks > 0 && time.Now().Second()%3 == 0 {
					currentStats.PendingTasks--
					currentStats.Completed++
					stats.Set(currentStats)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			table := ctx.Get("table").(bubbly.Component)
			statsRef := ctx.Get("stats")
			activePercentageRef := ctx.Get("activePercentage")
			completionRateRef := ctx.Get("completionRate")

			// Type assert to the correct types
			var currentStats Stats
			var activePercentage, completionRate int

			if ref, ok := statsRef.(*bubbly.Ref[Stats]); ok {
				currentStats = ref.Get().(Stats)
			}
			if comp, ok := activePercentageRef.(*bubbly.Computed[interface{}]); ok {
				activePercentage = comp.Get().(int)
			}
			if comp, ok := completionRateRef.(*bubbly.Computed[interface{}]); ok {
				completionRate = comp.Get().(int)
			}

			// Render cards with updated content
			totalCard := components.Card(components.CardProps{
				Title:   "Total Users",
				Content: fmt.Sprintf("%d", currentStats.TotalUsers),
				Width:   20,
				Padding: 1,
			})
			totalCard.Init()

			activeCard := components.Card(components.CardProps{
				Title:   "Active Users",
				Content: fmt.Sprintf("%d (%d%%)", currentStats.ActiveUsers, activePercentage),
				Width:   20,
				Padding: 1,
			})
			activeCard.Init()

			pendingCard := components.Card(components.CardProps{
				Title:   "Pending Tasks",
				Content: fmt.Sprintf("%d", currentStats.PendingTasks),
				Width:   20,
				Padding: 1,
			})
			pendingCard.Init()

			doneCard := components.Card(components.CardProps{
				Title:   "Completed",
				Content: fmt.Sprintf("%d (%d%%)", currentStats.Completed, completionRate),
				Width:   20,
				Padding: 1,
			})
			doneCard.Init()

			// Stats row with proper card rendering
			statsRow := lipgloss.JoinHorizontal(
				lipgloss.Top,
				totalCard.View(),
				" ",
				activeCard.View(),
				" ",
				pendingCard.View(),
				" ",
				doneCard.View(),
			)

			// Table section - render directly without extra wrapper
			tableTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Padding(0, 1).
				Render("User Activity")

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statsRow,
				"",
				tableTitle,
				"",
				table.View(),
			)
		}).
		Build()
}

func main() {
	component, err := createDashboard()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{
		component: component,
	}

	// Use tea.WithAltScreen() for full terminal screen mode
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
