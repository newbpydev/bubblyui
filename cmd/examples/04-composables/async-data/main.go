package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// User represents fetched user data
type User struct {
	ID    int
	Name  string
	Email string
	Role  string
}

// tickMsg is sent periodically to trigger UI updates while loading
// This is necessary because UseAsync updates Refs in a goroutine,
// but Bubbletea only redraws when Update() is called
type tickMsg time.Time

// tickCmd returns a command that ticks periodically
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// model wraps the async data component
type model struct {
	component bubbly.Component
	loading   bool // Track loading state to control ticking
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	// Start with a tick to handle initial loading
	return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			// Trigger refetch
			m.component.Emit("refetch", nil)
			m.loading = true
			cmds = append(cmds, tickCmd())
		}
	case tickMsg:
		// Continue ticking while loading
		// This ensures UI updates while the goroutine is running
		if m.loading {
			cmds = append(cmds, tickCmd())
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("📡 Composables - Async Data Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: UseAsync composable for data fetching with loading/error states",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"r: refetch data • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// fetchUser simulates an async API call
func fetchUser() (*User, error) {
	// Simulate network delay
	time.Sleep(2 * time.Second)

	// Simulate successful fetch
	// In a real app, this would be an HTTP request
	return &User{
		ID:    123,
		Name:  "Alice Johnson",
		Email: "alice@example.com",
		Role:  "Developer",
	}, nil
}

// createAsyncDataDemo creates a component demonstrating UseAsync
func createAsyncDataDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("AsyncDataDemo").
		Setup(func(ctx *bubbly.Context) {
			// Use the UseAsync composable for data fetching
			// This handles loading, error, and data state automatically
			userData := composables.UseAsync(ctx, fetchUser)

			// Track fetch count for demonstration
			fetchCount := ctx.Ref(0)

			// Fetch data when component mounts
			ctx.OnMounted(func() {
				userData.Execute()
				fetchCount.Set(fetchCount.GetTyped().(int) + 1)
			})

			// Expose state to template
			ctx.Expose("user", userData.Data)
			ctx.Expose("loading", userData.Loading)
			ctx.Expose("error", userData.Error)
			ctx.Expose("fetchCount", fetchCount)

			// Event handler for refetch
			ctx.On("refetch", func(_ interface{}) {
				userData.Execute()
				fetchCount.Set(fetchCount.GetTyped().(int) + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state from UseAsync
			user := ctx.Get("user").(*bubbly.Ref[*User])
			loading := ctx.Get("loading").(*bubbly.Ref[bool])
			errorRef := ctx.Get("error").(*bubbly.Ref[error])
			fetchCount := ctx.Get("fetchCount").(*bubbly.Ref[interface{}])

			loadingVal := loading.GetTyped()
			errorVal := errorRef.GetTyped()
			fetchCountVal := fetchCount.GetTyped().(int)

			// Status box
			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				Width(60).
				Align(lipgloss.Center)

			var statusBox string
			if loadingVal {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("214")).
					BorderForeground(lipgloss.Color("208"))
				statusBox = statusStyle.Render("⏳ Loading user data...")
			} else if errorVal != nil {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("196")).
					BorderForeground(lipgloss.Color("160"))
				statusBox = statusStyle.Render(fmt.Sprintf("❌ Error: %v", errorVal))
			} else {
				userData := user.GetTyped()
				if userData != nil {
					statusStyle = statusStyle.
						Foreground(lipgloss.Color("15")).
						Background(lipgloss.Color("35")).
						BorderForeground(lipgloss.Color("99"))
					statusBox = statusStyle.Render(fmt.Sprintf("✅ User: %s", userData.Name))
				} else {
					statusStyle = statusStyle.
						Foreground(lipgloss.Color("241")).
						BorderForeground(lipgloss.Color("240"))
					statusBox = statusStyle.Render("No data")
				}
			}

			// User data box
			userStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(60)

			var userBox string
			userData := user.GetTyped()
			if userData != nil {
				userBox = userStyle.Render(fmt.Sprintf(
					"User Details:\n\n"+
						"ID:    %d\n"+
						"Name:  %s\n"+
						"Email: %s\n"+
						"Role:  %s\n\n"+
						"Fetch Count: %d",
					userData.ID,
					userData.Name,
					userData.Email,
					userData.Role,
					fetchCountVal,
				))
			} else {
				userBox = userStyle.Render("No user data loaded yet")
			}

			// Info box explaining UseAsync
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(60)

			infoBox := infoStyle.Render(
				"UseAsync Pattern:\n\n" +
					"• Manages loading/error/data state\n" +
					"• Executes fetcher in goroutine\n" +
					"• Updates reactive refs automatically\n" +
					"• Provides Execute() and Reset() methods\n\n" +
					"Bubbletea Integration:\n" +
					"• Uses tea.Tick for UI updates\n" +
					"• Goroutine updates Refs\n" +
					"• Tick triggers redraws",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statusBox,
				"",
				userBox,
				"",
				infoBox,
			)
		}).
		Build()
}

func main() {
	component, err := createAsyncDataDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{component: component, loading: true}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
