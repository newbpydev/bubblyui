// Package pages provides page components for the todo router example.
package pages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// AddPageState holds the form state for the add page.
// Exported so parent can access and manipulate it.
type AddPageState struct {
	Title        *bubbly.Ref[string]
	Description  *bubbly.Ref[string]
	Priority     *bubbly.Ref[string]
	FocusedField *bubbly.Ref[int]
	ErrorMsg     *bubbly.Ref[string]
}

// NewAddPageState creates a new add page state.
func NewAddPageState() *AddPageState {
	return &AddPageState{
		Title:        bubbly.NewRef(""),
		Description:  bubbly.NewRef(""),
		Priority:     bubbly.NewRef("medium"),
		FocusedField: bubbly.NewRef(0),
		ErrorMsg:     bubbly.NewRef(""),
	}
}

// CreateAddPage creates the add todo page component.
// This is a pure view component - form logic is handled by parent app.go.
func CreateAddPage(state *AddPageState) (bubbly.Component, error) {
	return bubbly.NewComponent("AddPage").
		Setup(func(ctx *bubbly.Context) {
			// Pure view component - no setup needed
		}).
		Template(func(ctx bubbly.RenderContext) string {
			titleVal := state.Title.GetTyped()
			descVal := state.Description.GetTyped()
			priorityVal := state.Priority.GetTyped()
			focused := state.FocusedField.GetTyped()
			errMsg := state.ErrorMsg.GetTyped()

			// Field styles
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true)

			valueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

			focusedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("35")).
				Bold(true)

			cursorStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("35"))

			// Build form fields
			var fields []string

			// Title field
			titleLabel := "Title:"
			titleValue := titleVal
			if titleValue == "" {
				titleValue = "(type here)"
			}
			if focused == 0 {
				titleLabel = focusedStyle.Render("▶ " + titleLabel)
				titleValue = valueStyle.Render(titleVal) + cursorStyle.Render("▌")
			} else {
				titleLabel = labelStyle.Render("  " + titleLabel)
				titleValue = valueStyle.Render(titleValue)
			}
			fields = append(fields, titleLabel+" "+titleValue)

			// Description field
			descLabel := "Description:"
			descValue := descVal
			if descValue == "" {
				descValue = "(optional)"
			}
			if focused == 1 {
				descLabel = focusedStyle.Render("▶ " + descLabel)
				descValue = valueStyle.Render(descVal) + cursorStyle.Render("▌")
			} else {
				descLabel = labelStyle.Render("  " + descLabel)
				descValue = valueStyle.Render(descValue)
			}
			fields = append(fields, descLabel+" "+descValue)

			// Priority field
			priorityLabel := "Priority:"
			priorityDisplay := priorityVal
			priorityIcon := "🟡"
			switch priorityVal {
			case "high":
				priorityIcon = "🔴"
			case "low":
				priorityIcon = "🟢"
			}
			if focused == 2 {
				priorityLabel = focusedStyle.Render("▶ " + priorityLabel)
				priorityDisplay = valueStyle.Render(priorityVal) + cursorStyle.Render("▌")
			} else {
				priorityLabel = labelStyle.Render("  " + priorityLabel)
				priorityDisplay = valueStyle.Render(priorityVal)
			}
			fields = append(fields, priorityLabel+" "+priorityIcon+" "+priorityDisplay)

			// Error message
			if errMsg != "" {
				errorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true)
				fields = append(fields, "", errorStyle.Render("❌ "+errMsg))
			}

			// Create card
			card := components.Card(components.CardProps{
				Title:   "➕ Add New Todo",
				Content: strings.Join(fields, "\n"),
			})
			card.Init()

			// Help text
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)
			help := helpStyle.Render("tab: next field • shift+tab: prev field • p: cycle priority • enter: save • esc: cancel")

			return lipgloss.JoinVertical(
				lipgloss.Left,
				card.View(),
				help,
			)
		}).
		Build()
}
