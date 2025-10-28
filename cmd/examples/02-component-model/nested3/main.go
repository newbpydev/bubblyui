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

	title := titleStyle.Render("🔗 Nested3 - Three-Level Component Hierarchy")

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

	desc := descStyle.Render("Demonstrates: Container → List → Item components with event bubbling")

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, desc, componentView, help)
}

// ItemProps defines the props for an Item component
type ItemProps struct {
	ID    int
	Label string
}

// createItem creates an Item component (leaf level)
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

// createContainer creates the Container component (root level) with List child
func createContainer() (bubbly.Component, error) {
	// Create items first so we can track them
	item1, _ := createItem(1, "First Item")
	item2, _ := createItem(2, "Second Item")
	item3, _ := createItem(3, "Third Item")
	items := []bubbly.Component{item1, item2, item3}

	// Create List component with items
	// We'll pass a Ref for selection state so List can render with selection styling
	selectedIDRef := bubbly.NewRef[interface{}](0)

	list, _ := bubbly.NewComponent("List").
		Children(items...).
		Setup(func(ctx *bubbly.Context) {
			// Get children reference
			children := ctx.Children()
			ctx.Expose("children", children)

			// Store the selectedID ref so we can access it in template
			ctx.Expose("selectedID", selectedIDRef)

			// Event handler for selectItem (from Container)
			ctx.On("selectItem", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					// CRITICAL: Stop propagation to prevent infinite loop!
					// This event came from parent (Container), don't bubble it back up
					event.StopPropagation()

					if itemID, ok := event.Data.(int); ok {
						// Find the child with matching ID and activate it
						for _, child := range children {
							if props, ok := child.Props().(ItemProps); ok {
								if props.ID == itemID {
									child.Emit("activate", nil)
									break
								}
							}
						}
					}
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get children and selection state
			children := ctx.Get("children").([]bubbly.Component)
			selectedID := ctx.Get("selectedID").(*bubbly.Ref[interface{}])
			selected := selectedID.Get().(int)

			// List header
			headerStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("141")).
				Bold(true).
				MarginBottom(1)

			header := headerStyle.Render("  List Component:")

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
						Width(52)
					prefix = "▶ "
				} else {
					itemStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("250")).
						Padding(0, 2).
						Width(52)
					prefix = "  "
				}

				itemViews = append(itemViews, "    "+itemStyle.Render(prefix+childView))
			}

			return lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				lipgloss.JoinVertical(lipgloss.Left, itemViews...),
			)
		}).
		Build()

	return bubbly.NewComponent("Container").
		Children(list).
		Setup(func(ctx *bubbly.Context) {
			// Use the shared selectedID ref (same one used by List)
			ctx.Expose("selectedID", selectedIDRef)

			// Get list child reference
			children := ctx.Children()
			listChild := children[0] // We know there's only one child (the List)
			ctx.Expose("list", listChild)

			// Event handler for selectItem (from root model)
			ctx.On("selectItem", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					if itemID, ok := event.Data.(int); ok {
						// Forward the selectItem event to the List child
						listChild.Emit("selectItem", itemID)
					}
				}
			})

			// Event handler for selected (bubbled from Item through List)
			ctx.On("selected", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					if itemID, ok := event.Data.(int); ok {
						// Update the shared selectedID ref
						selectedIDRef.Set(itemID)
					}
				}
			})

			// Event handler for reset
			ctx.On("reset", func(data interface{}) {
				selectedIDRef.Set(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get selected ID and list child
			selectedID := ctx.Get("selectedID").(*bubbly.Ref[interface{}])
			selected := selectedID.Get().(int)
			listChild := ctx.Get("list").(bubbly.Component)

			// Container style
			containerStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Padding(1, 2).
				Width(60)

			// Header
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				MarginBottom(1)

			header := headerStyle.Render("Container Component:")

			// Render the List component (which will render its children with styling)
			listView := listChild.View()

			// Status
			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1).
				Italic(true)

			var status string
			if selected == 0 {
				status = statusStyle.Render("No item selected")
			} else {
				status = statusStyle.Render(fmt.Sprintf("Selected: Item %d (event bubbled: Item → List → Container)", selected))
			}

			// Combine all parts
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				"",
				listView,
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
