// Package demos provides demo views for each composable.
package demos

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/17-enhanced-composables/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateUseLoggerDemo creates the UseLogger demo view.
func CreateUseLoggerDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseLoggerDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			// Use shared logger entries
			entries := state.LoggerEntries.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `logger := composables.UseLogger(ctx, "MyComponent")
logger.Debug("Debug message")
logger.Info("Info message")
logger.Warn("Warning message")
logger.Error("Error message")
logger.SetLevel(composables.LogLevelWarn)  // Filter
logs := logger.Logs.GetTyped()
logger.Clear()`

			// Render log entries with colors
			levelColors := map[string]lipgloss.Color{
				"DEBUG": lipgloss.Color("243"),
				"INFO":  theme.Primary,
				"WARN":  theme.Warning,
				"ERROR": theme.Error,
			}

			var logContent strings.Builder
			for _, entry := range entries {
				color := levelColors[entry.Level]
				if color == "" {
					color = lipgloss.Color("252")
				}
				style := lipgloss.NewStyle().Foreground(color)
				// Use fixed-width level formatting for alignment
				logContent.WriteString(style.Render(fmt.Sprintf(
					"  [%-5s] %s\n",
					entry.Level, entry.Message,
				)))
			}

			if len(entries) == 0 {
				logContent.WriteString("  (no logs - press d/i/w/e to add)\n")
			}

			stateContent := fmt.Sprintf(
				"Log Entries (%d):\n%s\nPress d: debug | i: info | w: warn | e: error | c: clear",
				len(entries), logContent.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Logger Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseLogger provides structured logging with levels and component context. Logs are stored reactively for display in devtools or debug panels."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseLogger Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateUseNotificationDemo creates the UseNotification demo view.
func CreateUseNotificationDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("UseNotificationDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			notifications := state.Notifications.Notifications.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `notif := composables.UseNotification(ctx,
    composables.WithDefaultDuration(3*time.Second),
    composables.WithMaxNotifications(5),
)
notif.Success("Saved", "File saved successfully")
notif.Error("Error", "Failed to connect")
notif.Warning("Warning", "Low disk space")
notif.Info("Info", "New update available")
notif.Dismiss(id)  // Dismiss specific
notif.DismissAll() // Dismiss all`

			// Render notifications
			var notifContent strings.Builder
			typeColors := map[composables.NotificationType]lipgloss.Color{
				composables.NotificationSuccess: theme.Success,
				composables.NotificationError:   theme.Error,
				composables.NotificationWarning: theme.Warning,
				composables.NotificationInfo:    theme.Primary,
			}

			for _, n := range notifications {
				color := typeColors[n.Type]
				style := lipgloss.NewStyle().Foreground(color)
				icon := "●"
				switch n.Type {
				case composables.NotificationSuccess:
					icon = "✓"
				case composables.NotificationError:
					icon = "✗"
				case composables.NotificationWarning:
					icon = "⚠"
				case composables.NotificationInfo:
					icon = "ℹ"
				}
				notifContent.WriteString(style.Render(fmt.Sprintf(
					"  %s %s: %s\n", icon, n.Title, n.Message,
				)))
			}

			if len(notifications) == 0 {
				notifContent.WriteString("  (no notifications)\n")
			}

			stateContent := fmt.Sprintf(
				"Active Notifications (%d):\n%s\nPress s: success | e: error | w: warning | i: info | c: clear",
				len(notifications), notifContent.String(),
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Notification Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "UseNotification provides toast-style notifications with auto-dismiss. Supports success, error, warning, and info types."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("UseNotification Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateCreateSharedDemo creates the CreateShared demo view.
func CreateCreateSharedDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("CreateSharedDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			counterVal := state.CounterValue.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `// Define at package level
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *CounterComposable {
        return UseCounter(ctx, 0)
    },
)

// Use in any component - same instance everywhere!
func ComponentA(ctx *bubbly.Context) {
    counter := UseSharedCounter(ctx)
    counter.Increment()  // Affects all components
}

func ComponentB(ctx *bubbly.Context) {
    counter := UseSharedCounter(ctx)
    // Same counter instance as ComponentA!
}`

			stateContent := fmt.Sprintf(
				"Shared Counter Value: %d\n\nThis demo uses UseSharedDemoState\nwhich is created with CreateShared.\n\nAll components share the same state!\n\nPress +/-: change counter\n\nBenefits:\n  • Singleton pattern\n  • Cross-component state\n  • No prop drilling\n  • Thread-safe (sync.Once)",
				counterVal,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Shared State Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "CreateShared creates a singleton composable that's shared across all components. Uses sync.Once for thread-safe initialization."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("CreateShared Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}

// CreateCreateSharedWithResetDemo creates the CreateSharedWithReset demo view.
func CreateCreateSharedWithResetDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("CreateSharedWithResetDemo").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
			state := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("state", state)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			state := ctx.Get("state").(*localComposables.DemoStateComposable)

			counterVal := state.CounterValue.GetTyped()

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			codeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			usage := `// Define at package level
var UseSharedAuth = composables.CreateSharedWithReset(
    func(ctx *bubbly.Context) *AuthComposable {
        return UseAuth(ctx)
    },
)

// Use in components
auth := UseSharedAuth.Get(ctx)
auth.Login(user, pass)

// Reset on logout (creates fresh instance)
UseSharedAuth.Reset()

// Next call gets new instance
auth = UseSharedAuth.Get(ctx)  // Fresh state!`

			stateContent := fmt.Sprintf(
				"Shared Counter: %d\n\nPress +/-: change | r: reset to 50\n\nUse Cases:\n  • Auth state (reset on logout)\n  • Session data (reset on timeout)\n  • Form state (reset on submit)\n  • Cache (reset on invalidation)",
				counterVal,
			)

			stateCard := components.Card(components.CardProps{
				Title:   "Interactive Resettable State Demo",
				Content: stateContent,
				Width:   40,
			})
			stateCard.Init()

			descContent := "CreateSharedWithReset extends CreateShared with a Reset() method. Perfect for state that needs to be cleared (auth, sessions, caches)."

			descCard := components.Card(components.CardProps{
				Title:   "Description",
				Content: descContent,
				Width:   40,
			})
			descCard.Init()

			return lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("CreateSharedWithReset Demo"),
				"",
				codeStyle.Render(usage),
				"",
				stateCard.View(),
				"",
				descCard.View(),
			)
		}).
		Build()
}
