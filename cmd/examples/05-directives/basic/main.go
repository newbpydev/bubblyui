package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

// model wraps the component demonstrating If/Show directives
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.component.Emit("setStatus", "loading")
		case "2":
			m.component.Emit("setStatus", "error")
		case "3":
			m.component.Emit("setStatus", "success")
		case "4":
			m.component.Emit("setStatus", "empty")
		case " ", "space":
			// Toggle details visibility
			m.component.Emit("toggleDetails", nil)
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🎯 If/Show Directives Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Conditional rendering (If) and visibility toggle (Show)",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render(
		"1: loading • 2: error • 3: success • 4: empty • space: toggle details • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// createStatusComponent creates the component demonstrating If/Show directives
func createStatusComponent() (bubbly.Component, error) {
	return bubbly.NewComponent("StatusDemo").
		Setup(func(ctx *bubbly.Context) {
			// Application status state
			status := bubbly.NewRef("loading")
			showDetails := bubbly.NewRef(true)

			// Expose state to template
			ctx.Expose("status", status)
			ctx.Expose("showDetails", showDetails)

			// Event: Set status
			ctx.On("setStatus", func(data interface{}) {
				newStatus := data.(string)
				status.Set(newStatus)
			})

			// Event: Toggle details visibility
			ctx.On("toggleDetails", func(_ interface{}) {
				current := showDetails.GetTyped()
				showDetails.Set(!current)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			status := ctx.Get("status").(*bubbly.Ref[string])
			showDetails := ctx.Get("showDetails").(*bubbly.Ref[bool])

			// Status box style
			statusBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Padding(1, 2).
				Width(60)

			// Use If directive with ElseIf chain for status display
			statusContent := directives.If(status.GetTyped() == "loading",
				func() string {
					loadingStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("33")).
						Bold(true)
					return loadingStyle.Render("⏳ Loading...") + "\n\nPlease wait while we fetch the data."
				},
			).ElseIf(status.GetTyped() == "error",
				func() string {
					errorStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("196")).
						Bold(true)
					return errorStyle.Render("❌ Error Occurred") + "\n\nFailed to load data. Please try again."
				},
			).ElseIf(status.GetTyped() == "empty",
				func() string {
					emptyStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("241")).
						Bold(true)
					return emptyStyle.Render("📭 No Data") + "\n\nThere are no items to display."
				},
			).Else(func() string {
				successStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("35")).
					Bold(true)
				return successStyle.Render("✅ Success") + "\n\nData loaded successfully!"
			}).Render()

			statusBox := statusBoxStyle.Render(statusContent)

			// Use Show directive for collapsible details panel
			detailsContent := directives.Show(showDetails.GetTyped(), func() string {
				detailsStyle := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("240")).
					Padding(1, 2).
					Width(60).
					MarginTop(1)

				detailsHeader := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("99")).
					Render("📊 Details Panel")

				detailsBody := fmt.Sprintf(
					"\nCurrent Status: %s\nDetails Visible: %v\n\nThis panel can be toggled with the space key.",
					status.GetTyped(),
					showDetails.GetTyped(),
				)

				return detailsStyle.Render(detailsHeader + detailsBody)
			}).Render()

			// Nested If example - show additional info only for error status
			additionalInfo := directives.If(status.GetTyped() == "error",
				func() string {
					infoStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("208")).
						MarginTop(1).
						Italic(true)
					return infoStyle.Render("\n💡 Tip: Check your network connection and try again.")
				},
			).Render()

			return statusBox + "\n" + detailsContent + additionalInfo
		}).
		Build()
}

func main() {
	component, err := createStatusComponent()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	m := model{
		component: component,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
