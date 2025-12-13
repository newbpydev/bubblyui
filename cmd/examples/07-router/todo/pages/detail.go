// Package pages provides page components for the todo router example.
package pages

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/07-router/todo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CreateDetailPage creates the todo detail page component.
// Takes todoID as parameter since parent handles route parsing.
func CreateDetailPage(todoIDRef *bubbly.Ref[int]) (bubbly.Component, error) {
	return bubbly.NewComponent("DetailPage").
		Setup(func(ctx *bubbly.Context) {
			// Pure view component - no setup needed
		}).
		Template(func(ctx bubbly.RenderContext) string {
			todoID := todoIDRef.GetTyped()

			// Get the todo
			todo := composables.UseTodos().GetTodo(todoID)

			if todo == nil {
				// Not found card
				card := components.Card(components.CardProps{
					Title:   "❌ Todo Not Found",
					Content: fmt.Sprintf("No todo found with ID: %d\n\nPress 'b' or 'esc' to go back.", todoID),
				})
				card.Init()

				helpStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					MarginTop(1)
				help := helpStyle.Render("b/esc: go back")

				return lipgloss.JoinVertical(
					lipgloss.Left,
					card.View(),
					help,
				)
			}

			// Build detail content
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true)

			valueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

			// Status badge
			var statusBadge bubbly.Component
			if todo.Completed {
				statusBadge = components.Badge(components.BadgeProps{
					Label:   "✓ Completed",
					Variant: components.VariantSuccess,
				})
			} else {
				statusBadge = components.Badge(components.BadgeProps{
					Label:   "○ Pending",
					Variant: components.VariantWarning,
				})
			}
			statusBadge.Init()

			// Priority badge
			var priorityVariant components.Variant
			priorityIcon := "🟡"
			switch todo.Priority {
			case "high":
				priorityVariant = components.VariantDanger
				priorityIcon = "🔴"
			case "medium":
				priorityVariant = components.VariantWarning
				priorityIcon = "🟡"
			case "low":
				priorityVariant = components.VariantSuccess
				priorityIcon = "🟢"
			}
			priorityBadge := components.Badge(components.BadgeProps{
				Label:   priorityIcon + " " + todo.Priority,
				Variant: priorityVariant,
			})
			priorityBadge.Init()

			// Build content
			content := fmt.Sprintf(
				"%s %s\n\n%s %s\n\n%s\n%s\n\n%s\n%s",
				labelStyle.Render("Status:"),
				statusBadge.View(),
				labelStyle.Render("Priority:"),
				priorityBadge.View(),
				labelStyle.Render("Title:"),
				valueStyle.Render(todo.Title),
				labelStyle.Render("Description:"),
				valueStyle.Render(func() string {
					if todo.Description == "" {
						return "(no description)"
					}
					return todo.Description
				}()),
			)

			// Create card
			card := components.Card(components.CardProps{
				Title:   fmt.Sprintf("📝 Todo #%d", todo.ID),
				Content: content,
			})
			card.Init()

			// Help text
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)
			help := helpStyle.Render("space: toggle status • d: delete • b/esc: go back")

			return lipgloss.JoinVertical(
				lipgloss.Left,
				card.View(),
				help,
			)
		}).
		Build()
}
