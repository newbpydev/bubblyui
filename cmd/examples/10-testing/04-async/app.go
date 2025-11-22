package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localcomponents "github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/components"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root GitHubDashboard component
// Demonstrates BubblyUI's async operations with auto-refresh
func CreateApp(api composables.GitHubAPI) (bubbly.Component, error) {
	return bubbly.NewComponent("GitHubDashboard").
		WithAutoCommands(true).
		WithKeyBinding("r", "refresh", "Refresh data").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		Setup(func(ctx *bubbly.Context) {
			// Initialize dashboard composable (testable async logic)
			dashboard := composables.UseGitHubDashboard(ctx, "newbpydev", api)

			// Auto-refresh state
			autoRefresh := ctx.Ref(true)

			// PROVIDE theme to descendants (UseTheme/ProvideTheme pattern!)
			ctx.ProvideTheme(bubbly.DefaultTheme)

			// Create RepoList component
			repoList, err := localcomponents.CreateRepoList(localcomponents.RepoListProps{
				Repositories: dashboard.Repositories,
				Loading:      dashboard.LoadingRepos,
				Width:        60,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create repo list: %v", err))
				return
			}

			// Create ActivityFeed component
			activityFeed, err := localcomponents.CreateActivityFeed(localcomponents.ActivityFeedProps{
				Activity: dashboard.Activity,
				Loading:  dashboard.LoadingActivity,
				Width:    60,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create activity feed: %v", err))
				return
			}

			// Expose components for template
			if err := ctx.ExposeComponent("repoList", repoList); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose repo list: %v", err))
				return
			}
			if err := ctx.ExposeComponent("activityFeed", activityFeed); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose activity feed: %v", err))
				return
			}

			// Expose state for tests
			ctx.Expose("dashboard", dashboard)
			ctx.Expose("autoRefresh", autoRefresh)

			// Fetch data when component mounts
			ctx.OnMounted(func() {
				dashboard.Refresh()
			})

			// Event: Refresh data (R key)
			ctx.On("refresh", func(data interface{}) {
				dashboard.Refresh()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			dashboard := ctx.Get("dashboard").(*composables.GitHubDashboardComposable)
			repoList := ctx.Get("repoList").(bubbly.Component)
			activityFeed := ctx.Get("activityFeed").(bubbly.Component)

			// Title style
			titleStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("35")).
				Bold(true).
				Padding(1, 2).
				Background(lipgloss.Color("236"))

			title := titleStyle.Render("🚀 GitHub Dashboard - " + dashboard.Username.Get().(string))

			// Error display
			errorMsg := dashboard.Error.Get().(string)
			var errorView string
			if errorMsg != "" {
				errorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true).
					Padding(0, 2)
				errorView = errorStyle.Render("⚠ Error: "+errorMsg) + "\n\n"
			}

			// Help text
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true).
				Padding(1, 2)

			help := helpStyle.Render("Press [r] to refresh • [ctrl+c] to quit")

			// Render child components
			repoView := repoList.View()
			activityView := activityFeed.View()

			// Layout
			content := lipgloss.JoinHorizontal(
				lipgloss.Top,
				repoView,
				"  ",
				activityView,
			)

			return title + "\n\n" + errorView + content + "\n\n" + help
		}).
		Build()
}
