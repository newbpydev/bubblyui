package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Todo represents a todo item
type Todo struct {
	ID    int
	Text  string
	Done  bool
}

// TodoItemProps for individual todo items
type TodoItemProps struct {
	Todo Todo
}

// model wraps the todo app
type model struct {
	app           bubbly.Component
	input         string
	selectedIndex int
}

func (m model) Init() tea.Cmd {
	return m.app.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if len(m.input) == 0 {
				return m, tea.Quit
			}
			// If typing, allow 'q' in input
			m.input += "q"
			m.app.Emit("updateInput", m.input)
		case "enter":
			if len(m.input) > 0 {
				m.app.Emit("addTodo", m.input)
				m.input = ""
			} else if m.selectedIndex >= 0 {
				m.app.Emit("toggleTodo", m.selectedIndex)
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.app.Emit("updateInput", m.input)
			}
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			m.selectedIndex++
			// Will be clamped by app logic
		case "d":
			if m.selectedIndex >= 0 && len(m.input) == 0 {
				m.app.Emit("deleteTodo", m.selectedIndex)
			} else if len(m.input) > 0 {
				m.input += "d"
				m.app.Emit("updateInput", m.input)
			}
		case "c":
			if len(m.input) == 0 {
				m.app.Emit("clearCompleted", nil)
			} else {
				m.input += "c"
				m.app.Emit("updateInput", m.input)
			}
		default:
			// Add character to input
			if len(msg.Runes) == 1 {
				m.input += msg.String()
				m.app.Emit("updateInput", m.input)
			}
		}
	}

	updatedComponent, cmd := m.app.Update(msg)
	m.app = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	return m.app.View()
}

// createTodoItem creates a single todo item component
func createTodoItem(props TodoItemProps, index int, isSelected bool) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoItem").
		Props(props).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(TodoItemProps)
			todo := props.Todo

			itemStyle := lipgloss.NewStyle().
				Padding(0, 1).
				Width(50)

			if isSelected {
				itemStyle = itemStyle.Background(lipgloss.Color("237"))
			}

			checkbox := "☐"
			textStyle := lipgloss.NewStyle()

			if todo.Done {
				checkbox = "☑"
				textStyle = textStyle.
					Foreground(lipgloss.Color("240")).
					Strikethrough(true)
			} else {
				textStyle = textStyle.Foreground(lipgloss.Color("250"))
			}

			return itemStyle.Render(fmt.Sprintf(
				"%s %s",
				checkbox,
				textStyle.Render(todo.Text),
			))
		}).
		Build()
}

// createTodoApp creates the complete todo application
func createTodoApp(selectedIndexRef *int) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoApp").
		Setup(func(ctx *bubbly.Context) {
			// State
			todos := ctx.Ref([]Todo{})
			inputText := ctx.Ref("")

			// Computed values
			totalCount := ctx.Computed(func() interface{} {
				return len(todos.Get().([]Todo))
			})

			completedCount := ctx.Computed(func() interface{} {
				count := 0
				for _, todo := range todos.Get().([]Todo) {
					if todo.Done {
						count++
					}
				}
				return count
			})

			remainingCount := ctx.Computed(func() interface{} {
				return totalCount.Get().(int) - completedCount.Get().(int)
			})

			// Expose state
			ctx.Expose("todos", todos)
			ctx.Expose("inputText", inputText)
			ctx.Expose("totalCount", totalCount)
			ctx.Expose("completedCount", completedCount)
			ctx.Expose("remainingCount", remainingCount)

			// Event handlers
			ctx.On("addTodo", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					text := event.Data.(string)
					todoList := todos.Get().([]Todo)
					newID := len(todoList) + 1
					todoList = append(todoList, Todo{
						ID:   newID,
						Text: text,
						Done: false,
					})
					todos.Set(todoList)
					inputText.Set("")
				}
			})

			ctx.On("toggleTodo", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					index := event.Data.(int)
					todoList := todos.Get().([]Todo)
					if index >= 0 && index < len(todoList) {
						todoList[index].Done = !todoList[index].Done
						todos.Set(todoList)
					}
				}
			})

			ctx.On("deleteTodo", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					index := event.Data.(int)
					todoList := todos.Get().([]Todo)
					if index >= 0 && index < len(todoList) {
						todoList = append(todoList[:index], todoList[index+1:]...)
						todos.Set(todoList)
						// Adjust selection if needed
						if *selectedIndexRef >= len(todoList) && len(todoList) > 0 {
							*selectedIndexRef = len(todoList) - 1
						}
					}
				}
			})

			ctx.On("clearCompleted", func(data interface{}) {
				todoList := todos.Get().([]Todo)
				var remaining []Todo
				for _, todo := range todoList {
					if !todo.Done {
						remaining = append(remaining, todo)
					}
				}
				todos.Set(remaining)
				*selectedIndexRef = 0
			})

			ctx.On("updateInput", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					text := event.Data.(string)
					inputText.Set(text)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			todos := ctx.Get("todos").(*bubbly.Ref[interface{}])
			inputText := ctx.Get("inputText").(*bubbly.Ref[interface{}])
			totalCount := ctx.Get("totalCount").(*bubbly.Computed[interface{}])
			completedCount := ctx.Get("completedCount").(*bubbly.Computed[interface{}])
			remainingCount := ctx.Get("remainingCount").(*bubbly.Computed[interface{}])

			todoList := todos.Get().([]Todo)
			input := inputText.Get().(string)

			// Clamp selected index
			if *selectedIndexRef >= len(todoList) && len(todoList) > 0 {
				*selectedIndexRef = len(todoList) - 1
			}
			if *selectedIndexRef < 0 {
				*selectedIndexRef = 0
			}

			// Container style
			containerStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205")).
				Padding(1, 2).
				Width(60)

			// Title
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Align(lipgloss.Center).
				Width(54)

			title := titleStyle.Render("✅ Todo List - Complete App")

			// Input box
			inputStyle := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("86")).
				Padding(0, 1).
				Width(54).
				MarginTop(1)

			inputBox := inputStyle.Render(fmt.Sprintf("Add: %s", input))

			// Todo items
			var itemsView string
			if len(todoList) == 0 {
				emptyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Italic(true).
					Align(lipgloss.Center).
					Width(54).
					MarginTop(1)
				itemsView = emptyStyle.Render("No todos yet. Start typing to add one!")
			} else {
				var items []string
				for i, todo := range todoList {
					item, _ := createTodoItem(
						TodoItemProps{Todo: todo},
						i,
						i == *selectedIndexRef,
					)
					items = append(items, item.View())
				}
				itemsView = lipgloss.JoinVertical(lipgloss.Left, items...)
			}

			// Stats
			statsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1).
				Padding(0, 1)

			stats := statsStyle.Render(fmt.Sprintf(
				"Total: %d • Completed: %d • Remaining: %d",
				totalCount.Get().(int),
				completedCount.Get().(int),
				remainingCount.Get().(int),
			))

			// Help
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1).
				Padding(0, 1)

			help := helpStyle.Render(
				"enter: add/toggle • ↑/↓: navigate • d: delete • c: clear done • q: quit",
			)

			return containerStyle.Render(lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				inputBox,
				itemsView,
				stats,
				help,
			))
		}).
		Build()
}

func main() {
	m := model{
		input:         "",
		selectedIndex: 0,
	}

	app, err := createTodoApp(&m.selectedIndex)
	if err != nil {
		fmt.Printf("Error creating todo app: %v\n", err)
		os.Exit(1)
	}

	app.Init()
	m.app = app

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
