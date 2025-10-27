package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ItemProps for list items
type ItemProps struct {
	Text string
	ID   int
}

// ListProps for the list container
type ListProps struct {
	Title string
}

// model wraps the root component
type model struct {
	root bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.root.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.root.Emit("selectItem", 1)
		case "2":
			m.root.Emit("selectItem", 2)
		case "3":
			m.root.Emit("selectItem", 3)
		case "r":
			m.root.Emit("reset", nil)
		}
	}

	updatedComponent, cmd := m.root.Update(msg)
	m.root = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🪆 Nested Components - Composition")

	componentView := m.root.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render("1/2/3: select item • r: reset • q: quit")

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// createItem creates a single item component
func createItem(props ItemProps) (bubbly.Component, error) {
	return bubbly.NewComponent("Item").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			selected := ctx.Ref(false)
			ctx.Expose("selected", selected)

			ctx.On("select", func(data interface{}) {
				selected.Set(true)
			})

			ctx.On("deselect", func(data interface{}) {
				selected.Set(false)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ItemProps)
			selected := ctx.Get("selected").(*bubbly.Ref[interface{}])
			isSelected := selected.Get().(bool)

			itemStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.NormalBorder()).
				Width(40)

			if isSelected {
				itemStyle = itemStyle.
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("63")).
					BorderForeground(lipgloss.Color("99"))
			} else {
				itemStyle = itemStyle.
					Foreground(lipgloss.Color("250")).
					BorderForeground(lipgloss.Color("240"))
			}

			status := "  "
			if isSelected {
				status = "✓ "
			}

			return itemStyle.Render(fmt.Sprintf("%s%d. %s", status, props.ID, props.Text))
		}).
		Build()
}

// createList creates a list container with items
func createList(props ListProps, items []bubbly.Component) (bubbly.Component, error) {
	return bubbly.NewComponent("List").
		Props(props).
		Children(items...).
		Setup(func(ctx *bubbly.Context) {
			selectedID := ctx.Ref(0)
			ctx.Expose("selectedID", selectedID)

			ctx.On("itemSelected", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					itemID := event.Data.(int)
					selectedID.Set(itemID)

					// Select the specific item
					children := ctx.Children()
					for _, child := range children {
						childProps := child.Props().(ItemProps)
						if childProps.ID == itemID {
							child.Emit("select", nil)
						} else {
							child.Emit("deselect", nil)
						}
					}
				}
			})

			ctx.On("reset", func(data interface{}) {
				selectedID.Set(0)
				// Deselect all items
				children := ctx.Children()
				for _, child := range children {
					child.Emit("deselect", nil)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ListProps)
			selectedID := ctx.Get("selectedID").(*bubbly.Ref[interface{}])

			listStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("170")).
				Padding(1, 2).
				Width(46)

			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				MarginBottom(1)

			title := titleStyle.Render(props.Title)

			// Render children
			var itemViews []string
			for _, child := range ctx.Children() {
				itemViews = append(itemViews, ctx.RenderChild(child))
			}

			items := lipgloss.JoinVertical(lipgloss.Left, itemViews...)

			// Status
			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)

			selID := selectedID.Get().(int)
			status := "No selection"
			if selID > 0 {
				status = fmt.Sprintf("Selected: Item %d", selID)
			}
			statusText := statusStyle.Render(status)

			return listStyle.Render(lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				items,
				statusText,
			))
		}).
		Build()
}

// createContainer creates the root container
func createContainer(list bubbly.Component) (bubbly.Component, error) {
	return bubbly.NewComponent("Container").
		Children(list).
		Setup(func(ctx *bubbly.Context) {
			totalSelections := ctx.Ref(0)
			ctx.Expose("totalSelections", totalSelections)

			// Listen to list for selections
			children := ctx.Children()
			if len(children) > 0 {
				listComponent := children[0]
				listComponent.On("itemSelected", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						totalSelections.Set(totalSelections.Get().(int) + 1)
						// Forward event
						ctx.Emit("itemSelected", event.Data)
					}
				})
			}

			ctx.On("selectItem", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					itemID := event.Data.(int)
					children := ctx.Children()
					if len(children) > 0 {
						listComponent := children[0]
						// Reset all items first
						listComponent.Emit("reset", nil)
						// Send selection event to list
						listComponent.Emit("itemSelected", itemID)
					}
				}
			})

			ctx.On("reset", func(data interface{}) {
				totalSelections.Set(0)
				children := ctx.Children()
				if len(children) > 0 {
					children[0].Emit("reset", nil)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			totalSelections := ctx.Get("totalSelections").(*bubbly.Ref[interface{}])

			containerStyle := lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("205")).
				Padding(1, 2)

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

			header := headerStyle.Render("Container Component")

			// Render list child
			var listView string
			for _, child := range ctx.Children() {
				listView = ctx.RenderChild(child)
			}

			// Stats
			statsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				MarginTop(1)

			stats := statsStyle.Render(fmt.Sprintf(
				"Total selections: %d",
				totalSelections.Get().(int),
			))

			return containerStyle.Render(lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				listView,
				stats,
			))
		}).
		Build()
}

func main() {
	// Create item components
	item1, _ := createItem(ItemProps{Text: "First Item", ID: 1})
	item2, _ := createItem(ItemProps{Text: "Second Item", ID: 2})
	item3, _ := createItem(ItemProps{Text: "Third Item", ID: 3})

	// Create list with items
	list, err := createList(
		ListProps{Title: "Select an Item"},
		[]bubbly.Component{item1, item2, item3},
	)
	if err != nil {
		fmt.Printf("Error creating list: %v\n", err)
		os.Exit(1)
	}

	// Create container with list
	container, err := createContainer(list)
	if err != nil {
		fmt.Printf("Error creating container: %v\n", err)
		os.Exit(1)
	}

	container.Init()

	m := model{root: container}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
