package main

import (
	"fmt"
	"os"
	"time"

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
		WithAutoCommands(true). // Enable auto commands for async support
		WithKeyBinding("r", "refetch", "Refetch data").
		WithKeyBinding("q", "quit", "Quit").
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
			// Get component for help text
			comp := ctx.Component()

			// Get state from UseAsync
			user := ctx.Get("user").(*bubbly.Ref[*User])
			loading := ctx.Get("loading").(*bubbly.Ref[bool])
			errorRef := ctx.Get("error").(*bubbly.Ref[error])
			fetchCount := ctx.Get("fetchCount").(*bubbly.Ref[interface{}])

			loadingVal := loading.GetTyped()
			errorVal := errorRef.GetTyped()
			fetchCountVal := fetchCount.GetTyped().(int)

			// Title and subtitle
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

			// Help text
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(2)

			help := helpStyle.Render(comp.HelpText())

			return lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				subtitle,
				"",
				statusBox,
				"",
				userBox,
				"",
				infoBox,
				"",
				help,
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

	// Run with bubbly.Run() - async auto-detected, no tick wrapper needed!
	// The framework automatically detects WithAutoCommands(true) and enables
	// async refresh with 100ms interval. 80+ lines of boilerplate eliminated!
	if err := bubbly.Run(component, bubbly.WithAltScreen()); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
