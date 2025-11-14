package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// Todo represents a single todo item
type Todo struct {
	ID        int
	Title     string
	Completed bool
}

// CreateApp creates the root application component using composable architecture
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("MCPTodoApp").
		WithKeyBinding("ctrl+n", "addTodo", "New todo").
		WithKeyBinding(" ", "toggleTodo", "Toggle completion").
		WithKeyBinding("ctrl+d", "deleteTodo", "Delete todo").
		WithKeyBinding("up", "moveUp", "Move up").
		WithKeyBinding("down", "moveDown", "Move down").
		WithKeyBinding("ctrl+c", "quit", "Quit application")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Use todos composable for reactive state management
		todos := UseTodos(ctx)

		// Track selected todo index (context-aware for MCP tracking)
		selectedIndex := ctx.Ref(0)

		// Expose state for MCP inspection
		ctx.Expose("items", todos.Items)
		ctx.Expose("completedCount", todos.CompletedCount)
		ctx.Expose("totalCount", todos.TotalCount)
		ctx.Expose("selectedIndex", selectedIndex)

		// Register event handlers
		ctx.On("toggleTodo", func(data interface{}) {
			index := selectedIndex.Get().(int)
			todos.Toggle(index)
		})

		ctx.On("deleteTodo", func(data interface{}) {
			index := selectedIndex.Get().(int)
			todos.Delete(index)
		})

		ctx.On("addTodo", func(data interface{}) {
			title := data.(string)
			todos.Add(title)
		})

		ctx.On("moveUp", func(_ interface{}) {
			current := selectedIndex.Get().(int)
			if current > 0 {
				selectedIndex.Set(current - 1)
			}
		})

		ctx.On("moveDown", func(_ interface{}) {
			current := selectedIndex.Get().(int)
			todosList := todos.Items.Get().([]Todo)
			if current < len(todosList)-1 {
				selectedIndex.Set(current + 1)
			}
		})

		// Lifecycle hook
		ctx.OnMounted(func() {
			fmt.Fprintln(os.Stderr, "🎉 Todo app mounted - HTTP MCP server ready!")
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		items := ctx.Get("items").(*bubbly.Ref[interface{}]).Get().([]Todo)
		completed := ctx.Get("completedCount").(*bubbly.Computed[interface{}]).Get().(int)
		total := ctx.Get("totalCount").(*bubbly.Computed[interface{}]).Get().(int)
		selectedIdx := ctx.Get("selectedIndex").(*bubbly.Ref[int]).Get().(int)

		// Use BubblyUI Card for header
		headerCard := components.Card(components.CardProps{
			Title:   "MCP Todo App (HTTP Transport)",
			Content: renderStats(completed, total),
			Width:   70,
		})
		headerCard.Init()

		// Use BubblyUI List component for todos
		todoListContent := renderTodoList(items, selectedIdx)
		listCard := components.Card(components.CardProps{
			Title:   "Tasks",
			Content: todoListContent,
			Width:   70,
		})
		listCard.Init()

		// Use BubblyUI Text for help
		helpText := components.Text(components.TextProps{
			Content: "💡 AI Query: 'How many todos are completed?' or 'Show me the todo list'",
			Color:   lipgloss.Color("240"),
		})
		helpText.Init()

		return lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			headerCard.View(),
			"",
			listCard.View(),
			"",
			helpText.View(),
			"",
		)
	})

	return builder.Build()
}

// renderStats creates statistics display using BubblyUI components
func renderStats(completed, total int) string {
	// Use BubblyUI Badge for completion status
	variant := components.VariantInfo
	if completed == total && total > 0 {
		variant = components.VariantSuccess
	}

	completionBadge := components.Badge(components.BadgeProps{
		Label:   fmt.Sprintf("%d/%d Complete", completed, total),
		Variant: variant,
	})
	completionBadge.Init()

	// Use BubblyUI Text for description
	descText := components.Text(components.TextProps{
		Content: "HTTP MCP server on localhost:8765 with auth token",
		Color:   lipgloss.Color("240"),
	})
	descText.Init()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		completionBadge.View(),
		"",
		descText.View(),
		"",
	)
}

// renderTodoList creates todo list display using BubblyUI components
func renderTodoList(todos []Todo, selectedIndex int) string {
	if len(todos) == 0 {
		emptyText := components.Text(components.TextProps{
			Content: "No todos yet. Press ctrl+n to add one!",
			Color:   lipgloss.Color("240"),
		})
		emptyText.Init()
		return "\n" + emptyText.View() + "\n"
	}

	var items []string
	for i, todo := range todos {
		// Use BubblyUI Badge for completion status
		label := "○"
		variant := components.VariantSecondary
		if todo.Completed {
			label = "✓"
			variant = components.VariantSuccess
		}

		statusBadge := components.Badge(components.BadgeProps{
			Label:   label,
			Variant: variant,
		})
		statusBadge.Init()

		// Use BubblyUI Text for todo title
		color := lipgloss.Color("")
		if i == selectedIndex {
			color = lipgloss.Color("99") // Purple for selected
		} else if todo.Completed {
			color = lipgloss.Color("240") // Grey for completed
		}

		titleText := components.Text(components.TextProps{
			Content: todo.Title,
			Bold:    i == selectedIndex,
			Color:   color,
		})
		titleText.Init()

		cursor := " "
		if i == selectedIndex {
			cursor = "▶"
		}

		item := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cursor,
			" ",
			statusBadge.View(),
			" ",
			titleText.View(),
		)
		items = append(items, item)
	}

	return "\n" + lipgloss.JoinVertical(lipgloss.Left, items...) + "\n"
}
