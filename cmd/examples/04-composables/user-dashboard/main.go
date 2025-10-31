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

// UserProfile represents user profile data
type UserProfile struct {
	ID       int
	Name     string
	Email    string
	Role     string
	JoinDate string
}

// Activity represents a user activity entry
type Activity struct {
	ID        int
	Action    string
	Timestamp string
	Details   string
}

// Statistics represents dashboard statistics
type Statistics struct {
	TotalLogins    int
	ActiveProjects int
	TasksCompleted int
	Uptime         string
}

// tickMsg is sent periodically to trigger UI updates while loading
type tickMsg time.Time

// tickCmd returns a command that ticks periodically
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// model wraps the dashboard component
type model struct {
	component     bubbly.Component
	loadingStates map[string]bool // Track which sections are loading
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
		case "1":
			// Refresh user profile
			m.component.Emit("refreshProfile", nil)
			m.loadingStates["profile"] = true
			cmds = append(cmds, tickCmd())
		case "2":
			// Refresh activity
			m.component.Emit("refreshActivity", nil)
			m.loadingStates["activity"] = true
			cmds = append(cmds, tickCmd())
		case "3":
			// Refresh statistics
			m.component.Emit("refreshStats", nil)
			m.loadingStates["stats"] = true
			cmds = append(cmds, tickCmd())
		case "r":
			// Refresh all
			m.component.Emit("refreshAll", nil)
			m.loadingStates["profile"] = true
			m.loadingStates["activity"] = true
			m.loadingStates["stats"] = true
			cmds = append(cmds, tickCmd())
		}
	case tickMsg:
		// Continue ticking while any section is loading
		if m.loadingStates["profile"] || m.loadingStates["activity"] || m.loadingStates["stats"] {
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

	title := titleStyle.Render("📊 User Dashboard - UseAsync Composable")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Multiple UseAsync instances with independent loading/error states",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"1: refresh profile • 2: refresh activity • 3: refresh stats • r: refresh all • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// fetchUserProfile simulates fetching user profile from API
func fetchUserProfile() (*UserProfile, error) {
	// Simulate network delay
	time.Sleep(1500 * time.Millisecond)

	// Simulate successful fetch
	return &UserProfile{
		ID:       42,
		Name:     "Alice Johnson",
		Email:    "alice@example.com",
		Role:     "Senior Developer",
		JoinDate: "2023-01-15",
	}, nil
}

// fetchRecentActivity simulates fetching recent activity from API
func fetchRecentActivity() (*[]Activity, error) {
	// Simulate network delay
	time.Sleep(2000 * time.Millisecond)

	// Simulate successful fetch
	activities := []Activity{
		{ID: 1, Action: "Deployed to production", Timestamp: "2 hours ago", Details: "v2.1.0"},
		{ID: 2, Action: "Code review completed", Timestamp: "4 hours ago", Details: "PR #234"},
		{ID: 3, Action: "Created new branch", Timestamp: "6 hours ago", Details: "feature/auth"},
		{ID: 4, Action: "Fixed critical bug", Timestamp: "1 day ago", Details: "Issue #567"},
		{ID: 5, Action: "Updated documentation", Timestamp: "2 days ago", Details: "API docs"},
	}

	return &activities, nil
}

// fetchStatistics simulates fetching statistics from API
func fetchStatistics() (*Statistics, error) {
	// Simulate network delay
	time.Sleep(1000 * time.Millisecond)

	// Simulate successful fetch
	return &Statistics{
		TotalLogins:    1247,
		ActiveProjects: 8,
		TasksCompleted: 156,
		Uptime:         "99.9%",
	}, nil
}

// createDashboard creates the dashboard component
func createDashboard() (bubbly.Component, error) {
	return bubbly.NewComponent("Dashboard").
		Setup(func(ctx *bubbly.Context) {
			// Use UseAsync for each data source
			profileData := composables.UseAsync(ctx, fetchUserProfile)
			activityData := composables.UseAsync(ctx, fetchRecentActivity)
			statsData := composables.UseAsync(ctx, fetchStatistics)

			// Track fetch counts for demonstration
			profileFetchCount := ctx.Ref(0)
			activityFetchCount := ctx.Ref(0)
			statsFetchCount := ctx.Ref(0)

			// Fetch all data when component mounts
			ctx.OnMounted(func() {
				profileData.Execute()
				profileFetchCount.Set(1)

				activityData.Execute()
				activityFetchCount.Set(1)

				statsData.Execute()
				statsFetchCount.Set(1)
			})

			// Expose state to template
			ctx.Expose("profile", profileData.Data)
			ctx.Expose("profileLoading", profileData.Loading)
			ctx.Expose("profileError", profileData.Error)
			ctx.Expose("profileFetchCount", profileFetchCount)

			ctx.Expose("activity", activityData.Data)
			ctx.Expose("activityLoading", activityData.Loading)
			ctx.Expose("activityError", activityData.Error)
			ctx.Expose("activityFetchCount", activityFetchCount)

			ctx.Expose("stats", statsData.Data)
			ctx.Expose("statsLoading", statsData.Loading)
			ctx.Expose("statsError", statsData.Error)
			ctx.Expose("statsFetchCount", statsFetchCount)

			// Event: Refresh profile
			ctx.On("refreshProfile", func(_ interface{}) {
				profileData.Execute()
				count := profileFetchCount.GetTyped().(int)
				profileFetchCount.Set(count + 1)
			})

			// Event: Refresh activity
			ctx.On("refreshActivity", func(_ interface{}) {
				activityData.Execute()
				count := activityFetchCount.GetTyped().(int)
				activityFetchCount.Set(count + 1)
			})

			// Event: Refresh statistics
			ctx.On("refreshStats", func(_ interface{}) {
				statsData.Execute()
				count := statsFetchCount.GetTyped().(int)
				statsFetchCount.Set(count + 1)
			})

			// Event: Refresh all
			ctx.On("refreshAll", func(_ interface{}) {
				profileData.Execute()
				count1 := profileFetchCount.GetTyped().(int)
				profileFetchCount.Set(count1 + 1)

				activityData.Execute()
				count2 := activityFetchCount.GetTyped().(int)
				activityFetchCount.Set(count2 + 1)

				statsData.Execute()
				count3 := statsFetchCount.GetTyped().(int)
				statsFetchCount.Set(count3 + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get profile state
			profile := ctx.Get("profile").(*bubbly.Ref[*UserProfile])
			profileLoading := ctx.Get("profileLoading").(*bubbly.Ref[bool])
			profileError := ctx.Get("profileError").(*bubbly.Ref[error])
			profileFetchCount := ctx.Get("profileFetchCount").(*bubbly.Ref[interface{}])

			// Get activity state
			activity := ctx.Get("activity").(*bubbly.Ref[*[]Activity])
			activityLoading := ctx.Get("activityLoading").(*bubbly.Ref[bool])
			activityError := ctx.Get("activityError").(*bubbly.Ref[error])
			activityFetchCount := ctx.Get("activityFetchCount").(*bubbly.Ref[interface{}])

			// Get stats state
			stats := ctx.Get("stats").(*bubbly.Ref[*Statistics])
			statsLoading := ctx.Get("statsLoading").(*bubbly.Ref[bool])
			statsError := ctx.Get("statsError").(*bubbly.Ref[error])
			statsFetchCount := ctx.Get("statsFetchCount").(*bubbly.Ref[interface{}])

			// Profile panel
			profileStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(35).
				Height(10)

			var profileContent string
			if profileLoading.GetTyped() {
				profileContent = "⏳ Loading profile..."
			} else if profileError.GetTyped() != nil {
				profileContent = fmt.Sprintf("❌ Error: %v", profileError.GetTyped())
			} else {
				profileData := profile.GetTyped()
				if profileData != nil {
					profileContent = fmt.Sprintf(
						"👤 User Profile\n\n"+
							"Name:  %s\n"+
							"Email: %s\n"+
							"Role:  %s\n"+
							"Since: %s\n\n"+
							"Fetches: %d",
						profileData.Name,
						profileData.Email,
						profileData.Role,
						profileData.JoinDate,
						profileFetchCount.GetTyped().(int),
					)
				} else {
					profileContent = "No profile data"
				}
			}
			profileBox := profileStyle.Render(profileContent)

			// Statistics panel
			statsStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(35).
				Height(10)

			var statsContent string
			if statsLoading.GetTyped() {
				statsContent = "⏳ Loading statistics..."
			} else if statsError.GetTyped() != nil {
				statsContent = fmt.Sprintf("❌ Error: %v", statsError.GetTyped())
			} else {
				statsData := stats.GetTyped()
				if statsData != nil {
					statsContent = fmt.Sprintf(
						"📈 Statistics\n\n"+
							"Logins:    %d\n"+
							"Projects:  %d\n"+
							"Tasks:     %d\n"+
							"Uptime:    %s\n\n"+
							"Fetches: %d",
						statsData.TotalLogins,
						statsData.ActiveProjects,
						statsData.TasksCompleted,
						statsData.Uptime,
						statsFetchCount.GetTyped().(int),
					)
				} else {
					statsContent = "No statistics data"
				}
			}
			statsBox := statsStyle.Render(statsContent)

			// Activity panel
			activityStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("208")).
				Width(74).
				Height(12)

			var activityContent string
			if activityLoading.GetTyped() {
				activityContent = "⏳ Loading recent activity..."
			} else if activityError.GetTyped() != nil {
				activityContent = fmt.Sprintf("❌ Error: %v", activityError.GetTyped())
			} else {
				activityData := activity.GetTyped()
				if activityData != nil && len(*activityData) > 0 {
					activityContent = fmt.Sprintf("📋 Recent Activity (Fetches: %d)\n\n",
						activityFetchCount.GetTyped().(int))
					for _, act := range *activityData {
						activityContent += fmt.Sprintf(
							"• %s - %s\n  %s\n",
							act.Action,
							act.Timestamp,
							act.Details,
						)
					}
				} else {
					activityContent = "No activity data"
				}
			}
			activityBox := activityStyle.Render(activityContent)

			// Layout: Top row (profile + stats), bottom row (activity)
			topRow := lipgloss.JoinHorizontal(
				lipgloss.Top,
				profileBox,
				"  ",
				statsBox,
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				topRow,
				"",
				activityBox,
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
		loadingStates: map[string]bool{
			"profile":  true,
			"activity": true,
			"stats":    true,
		},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
