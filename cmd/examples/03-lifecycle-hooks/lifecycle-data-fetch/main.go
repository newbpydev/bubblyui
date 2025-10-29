package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// User represents fetched user data
type User struct {
	ID    int
	Name  string
	Email string
}

// fetchUserMsg is sent when user data is fetched
type fetchUserMsg struct {
	user *User
	err  error
}

// model wraps the data fetch component
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.component.Emit("refetch", nil)
		}
	case fetchUserMsg:
		// Forward fetch result to component
		m.component.Emit("user-fetched", msg)
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("📡 Lifecycle Hooks - Data Fetching Example")

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"r: refetch data • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// fetchUserCmd simulates async data fetching
func fetchUserCmd() tea.Cmd {
	return func() tea.Msg {
		// Simulate network delay
		time.Sleep(1500 * time.Millisecond)

		// Simulate successful fetch
		user := &User{
			ID:    123,
			Name:  "Alice Johnson",
			Email: "alice@example.com",
		}

		return fetchUserMsg{user: user, err: nil}
	}
}

// createDataFetchDemo creates a component demonstrating data fetching with lifecycle hooks
func createDataFetchDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("DataFetchDemo").
		Setup(func(ctx *bubbly.Context) {
			// State
			user := ctx.Ref((*User)(nil))
			loading := ctx.Ref(false)
			errorRef := ctx.Ref((error)(nil))
			fetchCount := ctx.Ref(0)
			events := ctx.Ref([]string{})

			// Helper to add event
			addEvent := func(event string) {
				current := events.Get().([]string)
				if len(current) >= 8 {
					current = current[1:]
				}
				current = append(current, event)
				events.Set(current)
			}

			// onMounted: Fetch data when component mounts
			ctx.OnMounted(func() {
				addEvent("✅ onMounted: Component mounted")
				addEvent("📡 Fetching user data...")
				loading.Set(true)
				count := fetchCount.Get().(int)
				fetchCount.Set(count + 1)

				// Trigger async fetch via Bubbletea command
				// In real app, this would be sent through the Update loop
				// For demo, we'll handle it via event
				ctx.Emit("start-fetch", nil)
			})

			// onUpdated: React to loading state changes
			ctx.OnUpdated(func() {
				isLoading := loading.Get().(bool)
				if isLoading {
					addEvent("⏳ Loading state: true")
				} else {
					addEvent("✅ Loading state: false")
				}
			}, loading)

			// onUpdated: React to user data changes
			ctx.OnUpdated(func() {
				userData := user.Get().(*User)
				if userData != nil {
					addEvent(fmt.Sprintf("👤 User loaded: %s", userData.Name))
				}
			}, user)

			// onUnmounted: Cleanup
			ctx.OnUnmounted(func() {
				addEvent("🛑 onUnmounted: Cleaning up")
			})

			// Cleanup: Cancel pending requests (simulated)
			ctx.OnCleanup(func() {
				addEvent("🧹 Cleanup: Cancelled pending requests")
			})

			// Expose state
			ctx.Expose("user", user)
			ctx.Expose("loading", loading)
			ctx.Expose("error", errorRef)
			ctx.Expose("fetchCount", fetchCount)
			ctx.Expose("events", events)

			// Event handlers
			ctx.On("start-fetch", func(data interface{}) {
				// This would normally send a tea.Cmd
				// For demo purposes, we'll simulate the fetch completing
				go func() {
					time.Sleep(1500 * time.Millisecond)
					ctx.Emit("user-fetched", fetchUserMsg{
						user: &User{
							ID:    123,
							Name:  "Alice Johnson",
							Email: "alice@example.com",
						},
						err: nil,
					})
				}()
			})

			ctx.On("user-fetched", func(data interface{}) {
				msg := data.(fetchUserMsg)
				loading.Set(false)

				if msg.err != nil {
					errorRef.Set(msg.err)
					addEvent(fmt.Sprintf("❌ Error: %v", msg.err))
				} else {
					user.Set(msg.user)
					errorRef.Set(nil)
				}
			})

			ctx.On("refetch", func(data interface{}) {
				addEvent("🔄 Refetching data...")
				user.Set(nil)
				loading.Set(true)
				errorRef.Set(nil)
				count := fetchCount.Get().(int)
				fetchCount.Set(count + 1)

				ctx.Emit("start-fetch", nil)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			user := ctx.Get("user").(*bubbly.Ref[interface{}])
			loading := ctx.Get("loading").(*bubbly.Ref[interface{}])
			error := ctx.Get("error").(*bubbly.Ref[interface{}])
			fetchCount := ctx.Get("fetchCount").(*bubbly.Ref[interface{}])
			events := ctx.Get("events").(*bubbly.Ref[interface{}])

			loadingVal := loading.Get().(bool)
			errorVal := error.Get()
			fetchCountVal := fetchCount.Get().(int)
			eventsVal := events.Get().([]string)

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
				userData := user.Get()
				if userData != nil {
					userObj := userData.(*User)
					statusStyle = statusStyle.
						Foreground(lipgloss.Color("15")).
						Background(lipgloss.Color("35")).
						BorderForeground(lipgloss.Color("99"))
					statusBox = statusStyle.Render(fmt.Sprintf("✅ User: %s", userObj.Name))
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
			userData := user.Get()
			if userData != nil {
				userObj := userData.(*User)
				userBox = userStyle.Render(fmt.Sprintf(
					"User Details:\n\n"+
						"ID:    %d\n"+
						"Name:  %s\n"+
						"Email: %s\n\n"+
						"Fetch Count: %d",
					userObj.ID,
					userObj.Name,
					userObj.Email,
					fetchCountVal,
				))
			} else {
				userBox = userStyle.Render("No user data loaded yet")
			}

			// Events log box
			eventsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(60).
				Height(10)

			eventsStr := "Lifecycle Events:\n\n"
			for _, event := range eventsVal {
				eventsStr += event + "\n"
			}

			eventsBox := eventsStyle.Render(eventsStr)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statusBox,
				"",
				userBox,
				"",
				eventsBox,
			)
		}).
		Build()
}

func main() {
	component, err := createDataFetchDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	component.Init()

	m := model{component: component}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Unmount component for cleanup demonstration
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
		fmt.Println("\n✅ Component unmounted - cleanup hooks executed")
	}
}
