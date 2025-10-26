package main

import (
	"fmt"
	"math/rand"
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

// fetchMsg is sent when data fetch completes
type fetchMsg struct {
	user User
	err  error
}

// tickMsg is sent on each frame for spinner animation
type tickMsg time.Time

// model represents the application state with async data loading
type model struct {
	// Reactive state
	loading *bubbly.Ref[bool]
	user    *bubbly.Ref[*User]
	error   *bubbly.Ref[error]

	// Computed state
	hasData *bubbly.Computed[bool]

	// Watcher cleanup
	cleanup func()
}

// Init initializes the model and starts data fetch
func (m model) Init() tea.Cmd {
	// Start loading immediately
	m.loading.Set(true)
	return tea.Batch(fetchUser(), tickCmd())
}

// tickCmd returns a command that sends a tick message for animation
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles incoming messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.cleanup != nil {
				m.cleanup()
			}
			return m, tea.Quit

		case "r":
			// Reload data
			m.loading.Set(true)
			m.error.Set(nil)
			return m, tea.Batch(fetchUser(), tickCmd())
		}

	case fetchMsg:
		// Data fetch completed
		m.loading.Set(false)
		if msg.err != nil {
			m.error.Set(msg.err)
		} else {
			m.user.Set(&msg.user)
			m.error.Set(nil)
		}
		return m, nil

	case tickMsg:
		// Continue ticking while loading
		if m.loading.Get() {
			return m, tickCmd()
		}
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	dataStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		MarginTop(1)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	// Build view
	title := titleStyle.Render("🌐 Async Data Loading with Watchers")

	var content string

	// Show different states based on reactive values
	if m.loading.Get() {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		frame := spinner[time.Now().UnixNano()/100000000%int64(len(spinner))]
		content = loadingStyle.Render(fmt.Sprintf("%s Loading user data...", frame))
	} else if m.error.Get() != nil {
		content = errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'r' to retry", m.error.Get()))
	} else if m.hasData.Get() {
		user := m.user.Get()
		content = dataStyle.Render(fmt.Sprintf(
			"User Data:\n\nID:    %d\nName:  %s\nEmail: %s",
			user.ID,
			user.Name,
			user.Email,
		))
	} else {
		content = "No data available"
	}

	help := helpStyle.Render("r reload • q quit")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n", title, content, help)
}

// fetchUser simulates an async API call
func fetchUser() tea.Cmd {
	return func() tea.Msg {
		// Simulate network delay
		time.Sleep(2 * time.Second)

		// Simulate random success/failure
		if rand.Float32() < 0.2 {
			return fetchMsg{
				err: fmt.Errorf("network error: connection timeout"),
			}
		}

		// Return mock user data
		return fetchMsg{
			user: User{
				ID:    rand.Intn(1000),
				Name:  fmt.Sprintf("User %d", rand.Intn(100)),
				Email: fmt.Sprintf("user%d@example.com", rand.Intn(100)),
			},
		}
	}
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create reactive state for async data
	loading := bubbly.NewRef(false)
	user := bubbly.NewRef[*User](nil)
	dataError := bubbly.NewRef[error](nil)

	// Computed value to check if we have data
	hasData := bubbly.NewComputed(func() bool {
		return user.Get() != nil
	})

	// Watch for loading state changes - demonstrates side effects!
	cleanup := bubbly.Watch(loading, func(newVal, oldVal bool) {
		if newVal {
			// Started loading
			fmt.Fprintln(os.Stderr, "[DEBUG] Started loading data...")
		} else {
			// Finished loading
			fmt.Fprintln(os.Stderr, "[DEBUG] Finished loading data")
		}
	})

	// Watch for data changes
	cleanup2 := bubbly.Watch(user, func(newVal, oldVal *User) {
		if newVal != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] User data received: %s\n", newVal.Name)
		}
	})

	// Combine cleanups
	combinedCleanup := func() {
		cleanup()
		cleanup2()
	}

	// Create model
	m := model{
		loading: loading,
		user:    user,
		error:   dataError,
		hasData: hasData,
		cleanup: combinedCleanup,
	}

	// Run the program with alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
