package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// model wraps the parent component
type model struct {
	parent bubbly.Component
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.parent.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			// Toggle theme
			m.parent.Emit("toggleTheme", nil)
		case "s":
			// Toggle size
			m.parent.Emit("toggleSize", nil)
		}
	}

	updatedParent, cmd := m.parent.Update(msg)
	m.parent = updatedParent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔗 Composables - Provide/Inject Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Dependency injection across component tree with provide/inject",
	)

	componentView := m.parent.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"t: toggle theme • s: toggle size • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// createChildComponent creates a child component that injects values from parent
func createChildComponent(name string) (bubbly.Component, error) {
	return bubbly.NewComponent(name).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme from parent (with default fallback)
			// This demonstrates dependency injection via inject
			theme := ctx.Inject("theme", "default")
			size := ctx.Inject("size", "medium")

			// Expose injected values to template
			ctx.Expose("theme", theme)
			ctx.Expose("size", size)
			ctx.Expose("name", name)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get injected values
			theme := ctx.Get("theme")
			size := ctx.Get("size")
			name := ctx.Get("name").(string)

			themeVal := theme.(string)
			sizeVal := size.(string)

			// Style based on injected theme
			var bgColor, fgColor, borderColor lipgloss.Color
			switch themeVal {
			case "dark":
				bgColor = lipgloss.Color("235")
				fgColor = lipgloss.Color("15")
				borderColor = lipgloss.Color("99")
			case "light":
				bgColor = lipgloss.Color("255")
				fgColor = lipgloss.Color("0")
				borderColor = lipgloss.Color("240")
			case "blue":
				bgColor = lipgloss.Color("63")
				fgColor = lipgloss.Color("15")
				borderColor = lipgloss.Color("99")
			default:
				bgColor = lipgloss.Color("240")
				fgColor = lipgloss.Color("15")
				borderColor = lipgloss.Color("241")
			}

			// Size based on injected size
			var width, padding int
			switch sizeVal {
			case "small":
				width = 30
				padding = 1
			case "medium":
				width = 40
				padding = 2
			case "large":
				width = 50
				padding = 3
			default:
				width = 40
				padding = 2
			}

			childStyle := lipgloss.NewStyle().
				Foreground(fgColor).
				Background(bgColor).
				Padding(padding, padding*2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Width(width)

			return childStyle.Render(fmt.Sprintf(
				"Child: %s\n\n"+
					"Theme: %s\n"+
					"Size:  %s\n\n"+
					"(Injected from parent)",
				name,
				themeVal,
				sizeVal,
			))
		}).
		Build()
}

// createParentComponent creates a parent component that provides values to children
func createParentComponent() (bubbly.Component, error) {
	// Create child components
	child1, err := createChildComponent("Component A")
	if err != nil {
		return nil, err
	}

	child2, err := createChildComponent("Component B")
	if err != nil {
		return nil, err
	}

	child3, err := createChildComponent("Component C")
	if err != nil {
		return nil, err
	}

	return bubbly.NewComponent("ParentComponent").
		Setup(func(ctx *bubbly.Context) {
			// Create reactive state for theme and size
			theme := ctx.Ref("dark")
			size := ctx.Ref("medium")

			// Provide values to child components
			// This makes them available to all descendants via inject
			ctx.Provide("theme", theme)
			ctx.Provide("size", size)

			// Expose state to template
			ctx.Expose("theme", theme)
			ctx.Expose("size", size)

			// Event handler for theme toggle
			ctx.On("toggleTheme", func(_ interface{}) {
				current := theme.GetTyped().(string)
				switch current {
				case "dark":
					theme.Set("light")
				case "light":
					theme.Set("blue")
				case "blue":
					theme.Set("dark")
				}
			})

			// Event handler for size toggle
			ctx.On("toggleSize", func(_ interface{}) {
				current := size.GetTyped().(string)
				switch current {
				case "small":
					size.Set("medium")
				case "medium":
					size.Set("large")
				case "large":
					size.Set("small")
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			theme := ctx.Get("theme").(*bubbly.Ref[interface{}])
			size := ctx.Get("size").(*bubbly.Ref[interface{}])

			themeVal := theme.GetTyped().(string)
			sizeVal := size.GetTyped().(string)

			// Parent box
			parentStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("170")).
				Padding(2, 4).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(60).
				Align(lipgloss.Center)

			parentBox := parentStyle.Render(fmt.Sprintf(
				"Parent Component\n\n"+
					"Providing:\n"+
					"  theme: %s\n"+
					"  size:  %s",
				themeVal,
				sizeVal,
			))

			// Get children views
			children := ctx.Children()
			var childViews []string
			for _, child := range children {
				childViews = append(childViews, child.View())
			}

			// Info box
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(60)

			infoBox := infoStyle.Render(
				"Provide/Inject Pattern:\n\n" +
					"• Parent provides theme and size\n" +
					"• Children inject values from tree\n" +
					"• Changes propagate automatically\n" +
					"• Type-safe dependency injection\n" +
					"• Works across any depth",
			)

			// Join all views
			result := lipgloss.JoinVertical(
				lipgloss.Left,
				parentBox,
				"",
			)

			// Add children in a horizontal layout
			if len(childViews) > 0 {
				childrenRow := lipgloss.JoinHorizontal(
					lipgloss.Top,
					childViews[0],
					" ",
					childViews[1],
					" ",
					childViews[2],
				)
				result = lipgloss.JoinVertical(
					lipgloss.Left,
					result,
					childrenRow,
					"",
				)
			}

			result = lipgloss.JoinVertical(
				lipgloss.Left,
				result,
				infoBox,
			)

			return result
		}).
		Children(child1, child2, child3).
		Build()
}

func main() {
	parent, err := createParentComponent()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{parent: parent}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
