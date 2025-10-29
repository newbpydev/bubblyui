package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

var debugLog *log.Logger

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
	debugLog.Println("=== model.Init() called ===")
	// Initialize component and trigger initial fetch
	cmd := tea.Batch(m.component.Init(), fetchUserCmd())
	debugLog.Println("=== model.Init() returning batch command ===")
	return cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	debugLog.Printf("=== model.Update() received msg: %T %+v ===\n", msg, msg)
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		debugLog.Printf("    KeyMsg: %s\n", msg.String())
		switch msg.String() {
		case "ctrl+c", "q":
			debugLog.Println("    Quitting...")
			return m, tea.Quit
		case "r":
			debugLog.Println("    Refetch triggered")
			m.component.Emit("refetch", nil)
			// Trigger fetch command
			cmds = append(cmds, fetchUserCmd())
			debugLog.Println("    Added fetchUserCmd to commands")
		}
	case fetchUserMsg:
		debugLog.Printf("    fetchUserMsg received: user=%+v, err=%v\n", msg.user, msg.err)
		// Forward fetch result to component
		m.component.Emit("user-fetched", msg)
		debugLog.Println("    Emitted 'user-fetched' event to component")
	default:
		debugLog.Printf("    Unknown message type: %T\n", msg)
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)
	
	if cmd != nil {
		cmds = append(cmds, cmd)
		debugLog.Println("    Component returned a command")
	}
	
	debugLog.Printf("=== model.Update() returning %d commands ===\n", len(cmds))
	return m, tea.Batch(cmds...)
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
	debugLog.Println("*** fetchUserCmd() called - returning command function ***")
	return func() tea.Msg {
		debugLog.Println("*** fetchUserCmd function executing (async) ***")
		debugLog.Println("*** Sleeping for 1.5 seconds... ***")
		// Simulate network delay
		time.Sleep(1500 * time.Millisecond)

		// Simulate successful fetch
		user := &User{
			ID:    123,
			Name:  "Alice Johnson",
			Email: "alice@example.com",
		}

		debugLog.Printf("*** fetchUserCmd returning fetchUserMsg: %+v ***\n", user)
		return fetchUserMsg{user: user, err: nil}
	}
}

// createDataFetchDemo creates a component demonstrating data fetching with lifecycle hooks
func createDataFetchDemo() (bubbly.Component, error) {
	debugLog.Println(">>> createDataFetchDemo() called <<<")
	return bubbly.NewComponent("DataFetchDemo").
		Setup(func(ctx *bubbly.Context) {
			debugLog.Println(">>> SETUP FUNCTION EXECUTING <<<")
			// State
			user := ctx.Ref((*User)(nil))
			loading := ctx.Ref(false)
			errorRef := ctx.Ref((error)(nil))
			fetchCount := ctx.Ref(0)
			events := ctx.Ref([]string{})
			debugLog.Println("    State initialized")

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
				// Note: Fetch is triggered by model.Init() via Bubbletea command
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
				userData := user.Get()
				if userData != nil {
					if userObj, ok := userData.(*User); ok && userObj != nil {
						addEvent(fmt.Sprintf("👤 User loaded: %s", userObj.Name))
					}
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
			// Note: Actual fetching is handled by Bubbletea commands in the model
			// The component just manages state based on received messages
			debugLog.Println("    Registering 'user-fetched' event handler...")

			ctx.On("user-fetched", func(data interface{}) {
				debugLog.Println("!!! EVENT HANDLER 'user-fetched' CALLED !!!")
				debugLog.Printf("    data type: %T, value: %+v\n", data, data)
				
				msg, ok := data.(fetchUserMsg)
				if !ok {
					debugLog.Printf("    ERROR: Type assertion failed! Expected fetchUserMsg, got %T\n", data)
					return
				}
				debugLog.Printf("    Handler received: %+v\n", msg)
				loading.Set(false)
				debugLog.Println("    Set loading to false")

				if msg.err != nil {
					errorRef.Set(msg.err)
					addEvent(fmt.Sprintf("❌ Error: %v", msg.err))
				} else {
					user.Set(msg.user)
					debugLog.Printf("    Set user to: %+v\n", msg.user)
					errorRef.Set(nil)
				}
				debugLog.Println("!!! EVENT HANDLER 'user-fetched' COMPLETE !!!")
			})
			debugLog.Println("    'user-fetched' handler registered")

			debugLog.Println("    Registering 'refetch' event handler...")
			ctx.On("refetch", func(data interface{}) {
				addEvent("🔄 Refetching data...")
				user.Set(nil)
				loading.Set(true)
				errorRef.Set(nil)
				count := fetchCount.Get().(int)
				fetchCount.Set(count + 1)
				// Note: Fetch is triggered by model.Update() via Bubbletea command
			})
			debugLog.Println("    'refetch' handler registered")
			debugLog.Println(">>> SETUP FUNCTION COMPLETE <<<")
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
					userObj, ok := userData.(*User)
					if ok && userObj != nil {
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
				userObj, ok := userData.(*User)
				if ok && userObj != nil {
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
	// Initialize debug logger to file
	logFile, err := os.OpenFile("/tmp/lifecycle-data-fetch-debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error creating debug log: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	debugLog = log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	
	debugLog.Println("========================================")
	debugLog.Println("=== Application Starting ===")
	debugLog.Println("========================================")

	component, err := createDataFetchDemo()
	if err != nil {
		debugLog.Printf("Error creating component: %v\n", err)
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}
	debugLog.Println("Component created successfully")

	// DON'T call component.Init() manually - Bubbletea will call model.Init() which calls it
	// Calling it here would mark the component as mounted before Bubbletea starts
	
	m := model{component: component}

	debugLog.Println("Starting Bubbletea program...")
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		debugLog.Printf("Error running program: %v\n", err)
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Unmount component for cleanup demonstration
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
		debugLog.Println("Component unmounted")
		fmt.Println("\n✅ Component unmounted - cleanup hooks executed")
	}
	
	debugLog.Println("========================================")
	debugLog.Println("=== Application Exiting ===")
	debugLog.Println("========================================")
	fmt.Println("\n📝 Debug log written to: /tmp/lifecycle-data-fetch-debug.log")
}
