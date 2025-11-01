package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

// model wraps the component demonstrating Bind directives
type model struct {
	component    bubbly.Component
	inputMode    bool
	focusedField string // "name", "email", "age", "agreed", "country"
	currentInput string
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle space key first (using msg.Type)
		if msg.Type == tea.KeySpace {
			if !m.inputMode {
				// Navigation mode: toggle checkbox if on agreed field
				if m.focusedField == "agreed" {
					m.component.Emit("toggleAgreed", nil)
				}
			} else {
				// Input mode: add space character
				m.currentInput += " "
				m.component.Emit("updateInput", m.currentInput)
			}
		} else {
			// Handle other keys using msg.String()
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				// ESC toggles input mode
				m.inputMode = !m.inputMode
				m.component.Emit("setInputMode", m.inputMode)
				if !m.inputMode {
					// Clear input when exiting input mode
					m.currentInput = ""
					m.component.Emit("updateInput", m.currentInput)
				}
			case "tab":
				// Cycle through fields
				fields := []string{"name", "email", "age", "agreed", "country"}
				for i, field := range fields {
					if field == m.focusedField {
						m.focusedField = fields[(i+1)%len(fields)]
						m.component.Emit("setFocusedField", m.focusedField)
						break
					}
				}
			case "enter":
				if m.inputMode {
					// Save current input to focused field
					if m.currentInput != "" {
						m.component.Emit("saveField", map[string]string{
							"field": m.focusedField,
							"value": m.currentInput,
						})
						m.currentInput = ""
						m.component.Emit("updateInput", m.currentInput)
					}
					// Stay in input mode for next field
				} else {
					// Enter input mode
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "backspace":
				if m.inputMode && len(m.currentInput) > 0 {
					m.currentInput = m.currentInput[:len(m.currentInput)-1]
					m.component.Emit("updateInput", m.currentInput)
				}
			default:
				// Handle text input - only in input mode
				if m.inputMode {
					switch msg.Type {
					case tea.KeyRunes:
						m.currentInput += string(msg.Runes)
						m.component.Emit("updateInput", m.currentInput)
					}
				}
			}
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

	title := titleStyle.Render("📝 Bind Directives Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Two-way binding with Bind, BindCheckbox, and BindSelect directives",
	)

	componentView := m.component.View()

	// Mode indicator
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginTop(1)

	var modeIndicator string
	if m.inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		modeIndicator = modeStyle.Render(fmt.Sprintf("✍️  INPUT MODE - Editing: %s", m.focusedField))
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99"))
		modeIndicator = modeStyle.Render(fmt.Sprintf("🧭 NAVIGATION MODE - Focused: %s", m.focusedField))
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render(
			"enter: save • esc: cancel • backspace: delete char • ctrl+c: quit",
		)
	} else {
		help = helpStyle.Render(
			"tab: next field • enter: edit field • space: toggle checkbox • q: quit",
		)
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n", title, subtitle, componentView, modeIndicator, help)
}

// createFormComponent creates the component demonstrating Bind directives
func createFormComponent() (bubbly.Component, error) {
	return bubbly.NewComponent("FormDemo").
		Setup(func(ctx *bubbly.Context) {
			// Form field values
			name := bubbly.NewRef("")
			email := bubbly.NewRef("")
			age := bubbly.NewRef(0)
			agreed := bubbly.NewRef(false)
			country := bubbly.NewRef("USA")

			// UI state
			focusedField := bubbly.NewRef("name")
			currentInput := bubbly.NewRef("")
			inputMode := bubbly.NewRef(false)

			// Country options for select
			countries := []string{"USA", "Canada", "UK", "Germany", "France", "Japan"}

			// Expose state to template
			ctx.Expose("name", name)
			ctx.Expose("email", email)
			ctx.Expose("age", age)
			ctx.Expose("agreed", agreed)
			ctx.Expose("country", country)
			ctx.Expose("countries", bubbly.NewRef(countries))
			ctx.Expose("focusedField", focusedField)
			ctx.Expose("currentInput", currentInput)
			ctx.Expose("inputMode", inputMode)

			// Event: Set input mode
			ctx.On("setInputMode", func(data interface{}) {
				mode := data.(bool)
				inputMode.Set(mode)
			})

			// Event: Update input
			ctx.On("updateInput", func(data interface{}) {
				input := data.(string)
				currentInput.Set(input)
			})

			// Event: Set focused field
			ctx.On("setFocusedField", func(data interface{}) {
				field := data.(string)
				focusedField.Set(field)
			})

			// Event: Save field value
			ctx.On("saveField", func(data interface{}) {
				fieldData := data.(map[string]string)
				field := fieldData["field"]
				value := fieldData["value"]

				switch field {
				case "name":
					name.Set(value)
				case "email":
					email.Set(value)
				case "age":
					// Parse age as integer
					var ageVal int
					fmt.Sscanf(value, "%d", &ageVal)
					age.Set(ageVal)
				case "country":
					country.Set(value)
				}
			})

			// Event: Toggle agreed checkbox
			ctx.On("toggleAgreed", func(_ interface{}) {
				agreed.Set(!agreed.GetTyped())
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			name := ctx.Get("name").(*bubbly.Ref[string])
			email := ctx.Get("email").(*bubbly.Ref[string])
			age := ctx.Get("age").(*bubbly.Ref[int])
			agreed := ctx.Get("agreed").(*bubbly.Ref[bool])
			country := ctx.Get("country").(*bubbly.Ref[string])
			countries := ctx.Get("countries").(*bubbly.Ref[[]string])
			focusedField := ctx.Get("focusedField").(*bubbly.Ref[string])
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[bool])

			// Form box style
			formBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1, 2).
				Width(70)

			// Conditional border color based on input mode
			if inputMode.GetTyped() {
				formBoxStyle = formBoxStyle.BorderForeground(lipgloss.Color("35"))
			} else {
				formBoxStyle = formBoxStyle.BorderForeground(lipgloss.Color("99"))
			}

			formHeader := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Render("📋 Registration Form")

			// Helper function to render field with focus indicator
			renderField := func(label, fieldName, value string) string {
				focused := focusedField.GetTyped() == fieldName
				labelStyle := lipgloss.NewStyle().
					Width(15).
					Bold(focused)

				if focused {
					labelStyle = labelStyle.Foreground(lipgloss.Color("35"))
				} else {
					labelStyle = labelStyle.Foreground(lipgloss.Color("252"))
				}

				cursor := "  "
				if focused {
					cursor = "▶ "
				}

				return fmt.Sprintf("%s%s %s\n", cursor, labelStyle.Render(label+":"), value)
			}

			// Use Bind directive for text inputs
			nameBinding := directives.Bind(name).Render()
			emailBinding := directives.Bind(email).Render()
			ageBinding := directives.Bind(age).Render()

			// Use BindCheckbox for boolean
			agreedBinding := directives.BindCheckbox(agreed).Render()

			// Use BindSelect for dropdown
			selectBinding := directives.BindSelect(country, countries.GetTyped()).Render()

			// Build form content
			formContent := formHeader + "\n\n" +
				renderField("Name", "name", nameBinding) +
				renderField("Email", "email", emailBinding) +
				renderField("Age", "age", ageBinding) +
				renderField("Terms", "agreed", agreedBinding) +
				renderField("Country", "country", selectBinding)

			formBox := formBoxStyle.Render(formContent)

			// Summary box showing current values
			summaryBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1, 2).
				Width(70).
				MarginTop(1)

			summaryHeader := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Render("📊 Current Values (Reactive)")

			// Use If directive to show validation status
			validationStatus := directives.If(name.GetTyped() != "" && email.GetTyped() != "" && agreed.GetTyped(),
				func() string {
					validStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("35")).
						Bold(true)
					return validStyle.Render("\n✅ Form is valid and ready to submit!")
				},
			).Else(func() string {
				invalidStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("208")).
					Italic(true)
				return invalidStyle.Render("\n⚠️  Please fill required fields (name, email) and agree to terms.")
			}).Render()

			summaryContent := fmt.Sprintf(
				"%s\n\nName: %s\nEmail: %s\nAge: %d\nAgreed to Terms: %v\nCountry: %s%s",
				summaryHeader,
				name.GetTyped(),
				email.GetTyped(),
				age.GetTyped(),
				agreed.GetTyped(),
				country.GetTyped(),
				validationStatus,
			)

			summaryBox := summaryBoxStyle.Render(summaryContent)

			// Note about Bind directives
			noteStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true).
				MarginTop(1)

			note := noteStyle.Render(
				"Note: Bind directives show placeholder representations. Full interactive binding\n" +
					"will be implemented in future tasks with proper TUI input handling.",
			)

			return formBox + "\n" + summaryBox + "\n" + note
		}).
		Build()
}

func main() {
	component, err := createFormComponent()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	m := model{
		component:    component,
		inputMode:    false,
		focusedField: "name",
		currentInput: "",
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
