package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// model wraps the container component
type model struct {
	container bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.container.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1", "2", "3":
			// Send selectItem event to container with item ID
			itemID := int(msg.String()[0] - '0') // Convert '1' -> 1, '2' -> 2, etc.
			m.container.Emit("selectItem", itemID)
		case "r":
			// Reset selection
			m.container.Emit("reset", nil)
		}
	}

	updatedComponent, cmd := m.container.Update(msg)
	m.container = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔗 Nested2 - Two-Level Component Hierarchy")

	componentView := m.container.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"1/2/3: select item • r: reset • q: quit",
	)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		MarginBottom(1)

	desc := descStyle.Render("Demonstrates: Container → Item components with event bubbling")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, desc, componentView, help)
}

// ItemProps defines the props for an Item component
type ItemProps struct {
	ID    int
	Label string
}

// createItem creates an Item component
func createItem(id int, label string) (bubbly.Component, error) {
	return bubbly.NewComponent(fmt.Sprintf("Item%d", id)).
		Props(ItemProps{ID: id, Label: label}).
		Setup(func(ctx *bubbly.Context) {
			// Event handler for activation
			ctx.On("activate", func(data interface{}) {
				// When activated, emit "selected" event with our ID
				props := ctx.Props().(ItemProps)
				ctx.Emit("selected", props.ID)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ItemProps)
			// Simple rendering - parent will add selection styling
			return fmt.Sprintf("%s (ID: %d)", props.Label, props.ID)
		}).
		Build()
}

// createContainer creates the Container component with Item children
func createContainer() (bubbly.Component, error) {
	// Create child Item components first
	item1, _ := createItem(1, "First Item")
	item2, _ := createItem(2, "Second Item")
	item3, _ := createItem(3, "Third Item")

	return bubbly.NewComponent("Container").
		Children(item1, item2, item3).
		Setup(func(ctx *bubbly.Context) {
			// Reactive state for selected item ID (0 = none selected)
			selectedID := ctx.Ref(0)
			ctx.Expose("selectedID", selectedID)

			// Get children reference
			children := ctx.Children()
			ctx.Expose("children", children)

			// Event handler for selectItem (from root model)
			ctx.On("selectItem", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					if itemID, ok := event.Data.(int); ok {
						// Find the child with matching ID and activate it
						for _, child := range children {
							if props, ok := child.Props().(ItemProps); ok {
								if props.ID == itemID {
									// Emit activate event to the specific child
									child.Emit("activate", nil)
									break
								}
							}
						}
					}
				}
			})

			// Event handler for selected (from child items)
			ctx.On("selected", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					if itemID, ok := event.Data.(int); ok {
						selectedID.Set(itemID)
					}
				}
			})

			// Event handler for reset
			ctx.On("reset", func(data interface{}) {
				selectedID.Set(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get selected ID
			selectedID := ctx.Get("selectedID").(*bubbly.Ref[interface{}])
			selected := selectedID.Get().(int)

			// Get children
			children := ctx.Get("children").([]bubbly.Component)

			// Container style
			containerStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Padding(1, 2).
				Width(50)

			// Header
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				MarginBottom(1)

			header := headerStyle.Render("Item List (Container → Items)")

			// Render children with selection styling
			var itemViews []string
			for _, child := range children {
				props := child.Props().(ItemProps)
				isSelected := props.ID == selected

				// Get child's view
				childView := child.View()

				// Apply selection styling
				var itemStyle lipgloss.Style
				var prefix string

				if isSelected {
					itemStyle = lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("15")).
						Background(lipgloss.Color("63")).
						Padding(0, 2).
						Width(44)
					prefix = "▶ "
				} else {
					itemStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("250")).
						Padding(0, 2).
						Width(44)
					prefix = "  "
				}

				itemViews = append(itemViews, itemStyle.Render(prefix+childView))
			}

			// Status
			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1).
				Italic(true)

			var status string
			if selected == 0 {
				status = statusStyle.Render("No item selected")
			} else {
				status = statusStyle.Render(fmt.Sprintf("Selected: Item %d (event bubbled from child)", selected))
			}

			// Combine all parts
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				"",
				lipgloss.JoinVertical(lipgloss.Left, itemViews...),
				"",
				status,
			)

			return containerStyle.Render(content)
		}).
		Build()
}

func main() {
	container, err := createContainer()
	if err != nil {
		fmt.Printf("Error creating container: %v\n", err)
		os.Exit(1)
	}

	container.Init()

	m := model{container: container}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
