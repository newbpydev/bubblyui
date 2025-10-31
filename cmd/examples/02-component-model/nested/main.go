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
		case "1":
			// Select first item
			m.container.Emit("select", 1)
		case "2":
			// Select second item
			m.container.Emit("select", 2)
		case "3":
			// Select third item
			m.container.Emit("select", 3)
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

	title := titleStyle.Render("🔗 Nested Components - Component Composition")

	componentView := m.container.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"1/2/3: select item • r: reset • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// ItemData represents data for a single item
type ItemData struct {
	ID    int
	Label string
}

// createContainer creates the root container component
func createContainer() (bubbly.Component, error) {
	// Define the items
	items := []ItemData{
		{ID: 1, Label: "First Item"},
		{ID: 2, Label: "Second Item"},
		{ID: 3, Label: "Third Item"},
	}

	return bubbly.NewComponent("Container").
		Setup(func(ctx *bubbly.Context) {
			// Reactive state for selected item ID (0 = none selected)
			selectedID := ctx.Ref(0)
			ctx.Expose("selectedID", selectedID)

			// Event handler for selection
			ctx.On("select", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					if id, ok := event.Data.(int); ok {
						selectedID.Set(id)
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
			selected := selectedID.GetTyped().(int)

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

			header := headerStyle.Render("Item List")

			// Render items
			var itemViews []string
			for _, item := range items {
				itemViews = append(itemViews, renderItem(item, selected))
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
				status = statusStyle.Render(fmt.Sprintf("Selected: Item %d", selected))
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

// renderItem renders a single item with selection styling
func renderItem(item ItemData, selectedID int) string {
	isSelected := item.ID == selectedID

	var style lipgloss.Style
	var prefix string

	if isSelected {
		// Selected style
		style = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Padding(0, 2).
			Width(44)
		prefix = "▶ "
	} else {
		// Normal style
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Padding(0, 2).
			Width(44)
		prefix = "  "
	}

	return style.Render(fmt.Sprintf("%s%s", prefix, item.Label))
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
