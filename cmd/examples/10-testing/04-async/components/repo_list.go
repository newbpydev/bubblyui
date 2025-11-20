package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/04-async/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// RepoListProps defines props for the repository list component
type RepoListProps struct {
	Repositories *bubbly.Ref[interface{}] // []composables.Repository
	Loading      *bubbly.Ref[interface{}] // bool
	Width        int
}

// CreateRepoList creates a repository list component
// Displays GitHub repositories with loading state
func CreateRepoList(props RepoListProps) (bubbly.Component, error) {
	return bubbly.NewComponent("RepoList").
		Setup(func(ctx *bubbly.Context) {
			ctx.Expose("repositories", props.Repositories)
			ctx.Expose("loading", props.Loading)
			ctx.Expose("width", props.Width)

			// INJECT theme colors from parent
			primaryColor := lipgloss.Color("35")   // Green
			secondaryColor := lipgloss.Color("99") // Purple
			mutedColor := lipgloss.Color("240")    // Dark grey

			if injected := ctx.Inject("primaryColor", nil); injected != nil {
				primaryColor = injected.(lipgloss.Color)
			}
			if injected := ctx.Inject("secondaryColor", nil); injected != nil {
				secondaryColor = injected.(lipgloss.Color)
			}
			if injected := ctx.Inject("mutedColor", nil); injected != nil {
				mutedColor = injected.(lipgloss.Color)
			}

			ctx.Expose("primaryColor", primaryColor)
			ctx.Expose("secondaryColor", secondaryColor)
			ctx.Expose("mutedColor", mutedColor)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			reposRef := ctx.Get("repositories").(*bubbly.Ref[interface{}])
			loadingRef := ctx.Get("loading").(*bubbly.Ref[interface{}])
			width := ctx.Get("width").(int)
			primaryColor := ctx.Get("primaryColor").(lipgloss.Color)
			secondaryColor := ctx.Get("secondaryColor").(lipgloss.Color)
			mutedColor := ctx.Get("mutedColor").(lipgloss.Color)

			loading := loadingRef.Get().(bool)
			repos := reposRef.Get().([]composables.Repository)

			// Title style
			titleStyle := lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 1)

			title := titleStyle.Render("📦 Repositories")

			// Container style
			containerStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2).
				Width(width)

			// Loading state
			if loading {
				loadingStyle := lipgloss.NewStyle().
					Foreground(mutedColor).
					Italic(true)
				content := loadingStyle.Render("Loading repositories...")
				return containerStyle.Render(title + "\n\n" + content)
			}

			// Empty state
			if len(repos) == 0 {
				emptyStyle := lipgloss.NewStyle().
					Foreground(mutedColor).
					Italic(true)
				content := emptyStyle.Render("No repositories found")
				return containerStyle.Render(title + "\n\n" + content)
			}

			// Render repositories
			var repoItems []string
			for i, repo := range repos {
				// Repo name style
				nameStyle := lipgloss.NewStyle().
					Foreground(secondaryColor).
					Bold(true)

				// Language badge style
				langStyle := lipgloss.NewStyle().
					Foreground(primaryColor).
					Background(lipgloss.Color("236")).
					Padding(0, 1)

				// Stars style
				starsStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("220")) // Yellow

				// Description style
				descStyle := lipgloss.NewStyle().
					Foreground(mutedColor)

				// Build repo item
				name := nameStyle.Render(repo.Name)
				lang := langStyle.Render(repo.Language)
				stars := starsStyle.Render(fmt.Sprintf("⭐ %d", repo.Stars))
				desc := descStyle.Render(repo.Description)

				// Truncate description if too long
				maxDescLen := width - 20
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}

				repoLine := fmt.Sprintf("%s  %s  %s", name, lang, stars)
				repoItem := repoLine + "\n" + desc

				repoItems = append(repoItems, repoItem)

				// Add separator between repos (not after last one)
				if i < len(repos)-1 {
					separator := lipgloss.NewStyle().
						Foreground(mutedColor).
						Render(strings.Repeat("─", width-4))
					repoItems = append(repoItems, separator)
				}
			}

			content := strings.Join(repoItems, "\n")
			return containerStyle.Render(title + "\n\n" + content)
		}).
		Build()
}
