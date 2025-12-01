// Package components provides UI components for the quickstart example.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	// Clean import paths using alias packages
	"github.com/newbpydev/bubblyui"
	"github.com/newbpydev/bubblyui/cmd/examples/00-quickstart/composables"
	"github.com/newbpydev/bubblyui/components"

	// Need pkg/bubbly for Context/RenderContext (builder callback types)
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TaskListProps defines the props for TaskList component.
type TaskListProps struct {
	Tasks         *bubblyui.Ref[[]composables.Task]
	SelectedIndex *bubblyui.Ref[int]
	Filter        *bubblyui.Ref[string]
	IsFocused     func() bool
	OnToggle      func(id int)
	OnDelete      func(id int)
	GetFiltered   func(filter string) []composables.Task
}

// CreateTaskList creates a component that displays the task list.
// This demonstrates:
// - Using the clean import path for bubblyui
// - Props-based composition
// - Using BubblyUI components (Card, Text)
// - Reactive state rendering
func CreateTaskList(props TaskListProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TaskList").
		Setup(func(ctx *bubbly.Context) {
			// Expose props for template access and DevTools visibility
			ctx.Expose("tasks", props.Tasks)
			ctx.Expose("selectedIndex", props.SelectedIndex)
			ctx.Expose("filter", props.Filter)

			// Lifecycle hook
			ctx.OnMounted(func() {
				// TaskList mounted - visible in DevTools
			})
		}).
		// Template receives RenderContext (no pointer!)
		Template(func(_ bubbly.RenderContext) string {
			// Use GetTyped() for type-safe access
			selectedIdx := props.SelectedIndex.GetTyped()
			filter := props.Filter.GetTyped()
			filteredTasks := props.GetFiltered(filter)

			// Build task lines
			var lines []string
			if len(filteredTasks) == 0 {
				emptyText := components.Text(components.TextProps{
					Content: "  No tasks yet. Press 'a' to add one!",
					Color:   lipgloss.Color("240"),
				})
				emptyText.Init()
				lines = append(lines, emptyText.View())
			} else {
				for i, task := range filteredTasks {
					// Checkbox state
					checkbox := "[ ]"
					if task.Done {
						checkbox = "[x]"
					}

					// Selection indicator
					indicator := "  "
					if i == selectedIdx && props.IsFocused() {
						indicator = "> "
					}

					// Style based on completion
					textColor := lipgloss.Color("252")
					if task.Done {
						textColor = lipgloss.Color("240")
					}

					line := fmt.Sprintf("%s%s %s", indicator, checkbox, task.Text)
					styledLine := lipgloss.NewStyle().
						Foreground(textColor).
						Render(line)

					lines = append(lines, styledLine)
				}
			}

			// Join all lines
			content := ""
			for i, line := range lines {
				if i > 0 {
					content += "\n"
				}
				content += line
			}

			// Filter indicator
			filterIndicator := fmt.Sprintf("Filter: %s", filter)
			filterStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				MarginTop(1)
			content += "\n" + filterStyle.Render(filterIndicator)

			// Wrap in a Card component (simple props - no BorderStyle/BorderColor)
			card := components.Card(components.CardProps{
				Title:   "Tasks",
				Content: content,
			})
			card.Init() // REQUIRED before View()!

			return card.View()
		}).
		Build()
}
