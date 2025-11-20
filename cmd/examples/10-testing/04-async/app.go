package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	localcomponents "github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/components"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// tickMsg is sent periodically for auto-refresh
type tickMsg struct{}

// CreateApp creates the root GitHubDashboard component that handles async init
func CreateApp(api composables.GitHubAPI) (bubbly.Component, error) {
	// Create wrapper first
	wrapper := &asyncWrapper{
		api:       api,
		needsInit: true,
	}

	// Create component with wrapper's dashboard reference
	comp, err := createAppComponent(api, wrapper)
	if err != nil {
		return nil, err
	}

	wrapper.component = comp
	return wrapper, nil
}

// asyncWrapper wraps the component to handle initial async commands
type asyncWrapper struct {
	component   bubbly.Component
	api         composables.GitHubAPI
	dashboard   *composables.GitHubDashboardComposable
	autoRefresh *bubbly.Ref[interface{}]
	needsInit   bool
}

func (w *asyncWrapper) Init() tea.Cmd {
	cmd := w.component.Init()

	if w.needsInit && w.dashboard != nil {
		w.needsInit = false

		// Return initial fetch commands batched with component init
		cmds := []tea.Cmd{cmd, w.dashboard.Refresh()}
		if w.autoRefresh != nil && w.autoRefresh.Get().(bool) {
			cmds = append(cmds, tick(30*time.Second))
		}
		return tea.Batch(cmds...)
	}

	return cmd
}

func (w *asyncWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle async messages before passing to component
	switch msg := msg.(type) {
	case ReposFetchedMsg:
		w.component.Emit("reposFetched", msg)
	case ActivityFetchedMsg:
		w.component.Emit("activityFetched", msg)
	case tickMsg:
		w.component.Emit("tick", msg)
		if w.autoRefresh != nil && w.autoRefresh.Get().(bool) {
			updated, cmd := w.component.Update(msg)
			w.component = updated.(bubbly.Component)
			return w, tea.Batch(cmd, w.dashboard.Refresh(), tick(30*time.Second))
		}
	case tea.KeyMsg:
		if msg.String() == "r" && w.dashboard != nil {
			updated, cmd := w.component.Update(msg)
			w.component = updated.(bubbly.Component)
			return w, tea.Batch(cmd, w.dashboard.Refresh())
		}
	}

	updated, cmd := w.component.Update(msg)
	w.component = updated.(bubbly.Component)
	return w, cmd
}

func (w *asyncWrapper) View() string {
	return w.component.View()
}

// Implement bubbly.Component interface
func (w *asyncWrapper) Name() string                                 { return w.component.Name() }
func (w *asyncWrapper) ID() string                                   { return w.component.ID() }
func (w *asyncWrapper) Props() interface{}                           { return w.component.Props() }
func (w *asyncWrapper) Emit(event string, data interface{})          { w.component.Emit(event, data) }
func (w *asyncWrapper) On(event string, handler bubbly.EventHandler) { w.component.On(event, handler) }
func (w *asyncWrapper) KeyBindings() map[string][]bubbly.KeyBinding  { return w.component.KeyBindings() }
func (w *asyncWrapper) HelpText() string                             { return w.component.HelpText() }
func (w *asyncWrapper) IsInitialized() bool                          { return !w.needsInit }
func (w *asyncWrapper) Get(key string) interface{} {
	// Expose dashboard and autoRefresh for tests
	switch key {
	case "dashboard":
		return w.dashboard
	case "autoRefresh":
		return w.autoRefresh
	default:
		// Delegate to wrapped component
		// Note: Component interface doesn't have Get(), so we can't delegate
		// Just return nil for unknown keys
		return nil
	}
}

func createAppComponent(api composables.GitHubAPI, wrapper *asyncWrapper) (bubbly.Component, error) {
	return bubbly.NewComponent("GitHubDashboard").
		WithAutoCommands(true).
		WithKeyBinding("r", "refresh", "Refresh data").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		Setup(func(ctx *bubbly.Context) {
			dashboard := composables.UseGitHubDashboard(ctx, "newbpydev", api)
			autoRefresh := ctx.Ref(true)

			// Store references in wrapper for async operations
			wrapper.dashboard = dashboard
			wrapper.autoRefresh = autoRefresh

			ctx.Provide("primaryColor", lipgloss.Color("35"))
			ctx.Provide("secondaryColor", lipgloss.Color("99"))
			ctx.Provide("mutedColor", lipgloss.Color("240"))
			ctx.Provide("warningColor", lipgloss.Color("220"))
			ctx.Provide("errorColor", lipgloss.Color("196"))

			repoList, err := localcomponents.CreateRepoList(localcomponents.RepoListProps{
				Repositories: dashboard.Repositories,
				Loading:      dashboard.LoadingRepos,
				Width:        60,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create repo list: %v", err))
				return
			}

			activityFeed, err := localcomponents.CreateActivityFeed(localcomponents.ActivityFeedProps{
				Activity: dashboard.Activity,
				Loading:  dashboard.LoadingActivity,
				Width:    60,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create activity feed: %v", err))
				return
			}

			if err := ctx.ExposeComponent("repoList", repoList); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose repo list: %v", err))
				return
			}
			if err := ctx.ExposeComponent("activityFeed", activityFeed); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose activity feed: %v", err))
				return
			}

			ctx.Expose("dashboard", dashboard)
			ctx.Expose("autoRefresh", autoRefresh)

			ctx.On("reposFetched", func(data interface{}) {
				msg := data.(ReposFetchedMsg)
				var repos []composables.Repository
				for _, r := range msg.Repos {
					repos = append(repos, composables.Repository{
						Name:        r.Name,
						Stars:       r.Stars,
						Language:    r.Language,
						Description: r.Description,
					})
				}
				dashboard.HandleReposMsg(repos, msg.Error)
			})

			ctx.On("activityFetched", func(data interface{}) {
				msg := data.(ActivityFetchedMsg)
				var activities []composables.Activity
				for _, a := range msg.Activity {
					activities = append(activities, composables.Activity{
						Type:      a.Type,
						Repo:      a.Repo,
						Message:   a.Message,
						Timestamp: a.Timestamp.Format("2006-01-02 15:04"),
					})
				}
				dashboard.HandleActivityMsg(activities, msg.Error)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			dashboard := ctx.Get("dashboard").(*composables.GitHubDashboardComposable)
			repoList := ctx.Get("repoList").(bubbly.Component)
			activityFeed := ctx.Get("activityFeed").(bubbly.Component)

			titleStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("35")).
				Bold(true).
				Padding(1, 2).
				Background(lipgloss.Color("236"))

			title := titleStyle.Render("🚀 GitHub Dashboard - " + dashboard.Username.Get().(string))

			errorMsg := dashboard.Error.Get().(string)
			var errorView string
			if errorMsg != "" {
				errorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true).
					Padding(0, 2)
				errorView = errorStyle.Render("⚠ Error: "+errorMsg) + "\n\n"
			}

			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true).
				Padding(1, 2)

			help := helpStyle.Render("Press [r] to refresh • [ctrl+c] to quit")

			repoView := repoList.View()
			activityView := activityFeed.View()

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

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
