package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ButtonProps defines the configuration for a button component
type ButtonProps struct {
	Label   string
	Primary bool
}

// model wraps the button component for Bubbletea integration
type model struct {
	button      bubbly.Component
	isPrimary   bool
	clickCount  int
}

func (m model) Init() tea.Cmd {
	return m.button.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", " ":
			// Increment click counter
			m.clickCount++
			// Simulate button click
			m.button.Emit("click", nil)
		case "p":
			// Toggle primary style and recreate component while preserving state
			m.isPrimary = !m.isPrimary
			currentProps := m.button.Props().(ButtonProps)
			newButton, _ := createButton(ButtonProps{
				Label:   currentProps.Label,
				Primary: m.isPrimary,
			}, m.clickCount)
			m.button = newButton
			m.button.Init()
		}
	}

	// Forward other messages to component
	updatedComponent, cmd := m.button.Update(msg)
	m.button = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔘 Button Component Example")

	// Component view
	componentView := m.button.View()

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"enter/space: click • p: toggle primary style • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// createButton creates a button component with the given props and initial click count
func createButton(props ButtonProps, initialClicks int) (bubbly.Component, error) {
	return bubbly.NewComponent("Button").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Track click count with initial value
			clicks := ctx.Ref(initialClicks)
			ctx.Expose("clicks", clicks)

			// Handle click events
			ctx.On("click", func(data interface{}) {
				current := clicks.Get().(int)
				clicks.Set(current + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ButtonProps)
			clicks := ctx.Get("clicks").(*bubbly.Ref[interface{}])
			clickCount := clicks.Get().(int)

			// Button style
			buttonStyle := lipgloss.NewStyle().
				Padding(0, 3).
				Border(lipgloss.RoundedBorder())

			if props.Primary {
				buttonStyle = buttonStyle.
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("63")).
					BorderForeground(lipgloss.Color("99"))
			} else {
				buttonStyle = buttonStyle.
					Foreground(lipgloss.Color("250")).
					BorderForeground(lipgloss.Color("240"))
			}

			buttonText := fmt.Sprintf("[ %s ]", props.Label)
			button := buttonStyle.Render(buttonText)

			// Info style
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)

			info := infoStyle.Render(fmt.Sprintf(
				"Clicked: %d times • Primary: %v",
				clickCount,
				props.Primary,
			))

			return fmt.Sprintf("%s\n%s", button, info)
		}).
		Build()
}

func main() {
	// Create button component
	button, err := createButton(ButtonProps{
		Label:   "Click Me!",
		Primary: true,
	}, 0)
	if err != nil {
		fmt.Printf("Error creating button: %v\n", err)
		os.Exit(1)
	}

	// Initialize component
	button.Init()

	// Create model
	m := model{
		button:     button,
		isPrimary:  true,
		clickCount: 0,
	}

	// Run with alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
