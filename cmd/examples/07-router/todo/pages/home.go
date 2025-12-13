// Package pages provides page components for the todo router example.
package pages

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/07-router/todo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateHomePage creates the home page component that displays the todo list.
// This is a pure view component - all navigation is handled by the parent app.
func CreateHomePage() (bubbly.Component, error) {
	return bubbly.NewComponent("HomePage").
		Setup(func(ctx *bubbly.Context) {
			// Pure view component - no setup needed
			// Navigation and state management handled by parent app.go
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get shared todos directly from composable
			todoManager := composables.UseTodos()
			todoList := todoManager.Todos.GetTyped()
			selected := todoManager.SelectedIndex.GetTyped()

			// Get stats
			total, completed, pending := todoManager.GetStats()

			// Stats badges
			totalBadge := components.Badge(components.BadgeProps{
				Label:   fmt.Sprintf("Total: %d", total),
				Variant: components.VariantSecondary,
			})
			totalBadge.Init()

			completedBadge := components.Badge(components.BadgeProps{
				Label:   fmt.Sprintf("Done: %d", completed),
				Variant: components.VariantSuccess,
			})
			completedBadge.Init()

			pendingBadge := components.Badge(components.BadgeProps{
				Label:   fmt.Sprintf("Pending: %d", pending),
				Variant: components.VariantWarning,
			})
			pendingBadge.Init()

			statsRow := lipgloss.JoinHorizontal(
				lipgloss.Center,
				totalBadge.View(),
				"  ",
				completedBadge.View(),
				"  ",
				pendingBadge.View(),
			)

			// Build todo list
			var todoItems string
			if len(todoList) == 0 {
				emptyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Italic(true)
				todoItems = emptyStyle.Render("No todos yet. Press 'a' to add one!")
			} else {
				for i, todo := range todoList {
					cursor := "  "
					if i == selected {
						cursor = "▶ "
					}

					checkbox := "☐"
					if todo.Completed {
						checkbox = "☑"
					}

					priorityIcon := ""
					switch todo.Priority {
					case "high":
						priorityIcon = "🔴"
					case "medium":
						priorityIcon = "🟡"
					case "low":
						priorityIcon = "🟢"
					}

					title := todo.Title
					titleStyle := lipgloss.NewStyle()
					if todo.Completed {
						titleStyle = titleStyle.
							Foreground(lipgloss.Color("240")).
							Strikethrough(true)
					} else if i == selected {
						titleStyle = titleStyle.
							Foreground(lipgloss.Color("99")).
							Bold(true)
					}

					line := fmt.Sprintf("%s%s %s %s", cursor, checkbox, priorityIcon, titleStyle.Render(title))
					if todoItems != "" {
						todoItems += "\n"
					}
					todoItems += line
				}
			}

			// Create card for todo list
			card := components.Card(components.CardProps{
				Title:   "📋 Todo List",
				Content: todoItems,
			})
			card.Init()

			// Help text
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)
			help := helpStyle.Render("↑/↓: navigate • space: toggle • enter: view • a: add • d: delete • q: quit")

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statsRow,
				"",
				card.View(),
				help,
			)
		}).
		Build()
}
