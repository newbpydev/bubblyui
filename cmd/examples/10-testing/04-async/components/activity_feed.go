package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ActivityFeedProps defines the properties for the ActivityFeed component
type ActivityFeedProps struct {
	Activity *bubbly.Ref[interface{}] // []composables.Activity
	Loading  *bubbly.Ref[interface{}] // bool
	Width    int
}

// CreateActivityFeed creates an activity feed component
// Displays GitHub activity with loading state
func CreateActivityFeed(props ActivityFeedProps) (bubbly.Component, error) {
	return bubbly.NewComponent("ActivityFeed").
		Setup(func(ctx *bubbly.Context) {
			ctx.Expose("activity", props.Activity)
			ctx.Expose("loading", props.Loading)
			ctx.Expose("width", props.Width)

			// USE theme from parent (UseTheme pattern!)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			activityRef := ctx.Get("activity").(*bubbly.Ref[interface{}])
			loadingRef := ctx.Get("loading").(*bubbly.Ref[interface{}])
			width := ctx.Get("width").(int)
			theme := ctx.Get("theme").(bubbly.Theme)

			loading := loadingRef.Get().(bool)
			activity := activityRef.Get().([]composables.Activity)

			// Title style
			titleStyle := lipgloss.NewStyle().
				Foreground(theme.Primary).
				Bold(true).
				Padding(0, 1)

			title := titleStyle.Render("📊 Recent Activity")

			// Container style
			containerStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Primary).
				Padding(1, 2).
				Width(width)

			// Loading state
			if loading {
				loadingStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true)
				content := loadingStyle.Render("Loading activity...")
				return containerStyle.Render(title + "\n\n" + content)
			}

			// Empty state
			if len(activity) == 0 {
				emptyStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true)
				content := emptyStyle.Render("No recent activity")
				return containerStyle.Render(title + "\n\n" + content)
			}

			// Render activities
			var activityItems []string
			for i, act := range activity {
				// Activity type icon and color
				var icon string
				var typeColor lipgloss.Color
				switch act.Type {
				case "push":
					icon = "📝"
					typeColor = theme.Primary
				case "pr":
					icon = "🔀"
					typeColor = theme.Secondary
				case "issue":
					icon = "🐛"
					typeColor = theme.Warning
				case "star":
					icon = "⭐"
					typeColor = lipgloss.Color("220")
				default:
					icon = "•"
					typeColor = theme.Muted
				}

				// Type style
				typeStyle := lipgloss.NewStyle().
					Foreground(typeColor).
					Bold(true)

				// Repo style
				repoStyle := lipgloss.NewStyle().
					Foreground(theme.Secondary)

				// Message style
				msgStyle := lipgloss.NewStyle().
					Foreground(theme.Muted)

				// Timestamp style
				timeStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true)

				// Build activity item
				typeText := typeStyle.Render(fmt.Sprintf("%s %s", icon, strings.ToUpper(act.Type)))
				repo := repoStyle.Render(act.Repo)
				msg := msgStyle.Render(act.Message)
				timestamp := timeStyle.Render(act.Timestamp)

				// Truncate message if too long
				maxMsgLen := width - 20
				if len(msg) > maxMsgLen {
					msg = msg[:maxMsgLen-3] + "..."
				}

				activityLine := fmt.Sprintf("%s  %s", typeText, repo)
				activityItem := activityLine + "\n" + msg + "\n" + timestamp

				activityItems = append(activityItems, activityItem)

				// Add separator between activities (not after last one)
				if i < len(activity)-1 {
					separator := lipgloss.NewStyle().
						Foreground(theme.Muted).
						Render(strings.Repeat("─", width-4))
					activityItems = append(activityItems, separator)
				}
			}

			content := strings.Join(activityItems, "\n")
			return containerStyle.Render(title + "\n\n" + content)
		}).
		Build()
}
