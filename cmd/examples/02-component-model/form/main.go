package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// FormProps defines form configuration
type FormProps struct {
	Title        string
	RequireEmail bool
	MinPassword  int
}

// model wraps form and result display
type model struct {
	form         bubbly.Component
	currentField int // 0=email, 1=password
	submitted    bool
	submitData   map[string]string
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.submitted {
				return m, tea.Quit
			}
			return m, tea.Quit
		case "tab":
			m.currentField = (m.currentField + 1) % 2
			m.form.Emit("focusField", m.currentField)
		case "enter":
			if !m.submitted {
				m.form.Emit("submit", nil)
			}
		case "esc":
			if m.submitted {
				m.submitted = false
			} else {
				m.form.Emit("reset", nil)
			}
		case "backspace":
			if m.currentField == 0 {
				m.form.Emit("deleteEmail", nil)
			} else {
				m.form.Emit("deletePassword", nil)
			}
		default:
			// Add character to current field
			if len(msg.Runes) == 1 && !m.submitted {
				char := msg.String()
				if m.currentField == 0 {
					m.form.Emit("typeEmail", char)
				} else {
					m.form.Emit("typePassword", char)
				}
			}
		}
	}

	updatedComponent, cmd := m.form.Update(msg)
	m.form = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("📋 Form Component - Props & Events")

	if m.submitted {
		successStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			Background(lipgloss.Color("22")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("46"))

		success := successStyle.Render(fmt.Sprintf(
			"✓ Form Submitted Successfully!\n\nEmail: %s\nPassword: %s",
			m.submitData["email"],
			strings.Repeat("*", len(m.submitData["password"])),
		))

		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

		help := helpStyle.Render("esc: back to form • q: quit")

		return fmt.Sprintf("%s\n\n%s\n%s\n", title, success, help)
	}

	componentView := m.form.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render(
		"tab: next field • enter: submit • esc: reset • backspace: delete • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// createForm creates a form component with validation
func createForm(props FormProps, onSubmit func(email, password string)) (bubbly.Component, error) {
	return bubbly.NewComponent("Form").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Form state
			email := ctx.Ref("")
			password := ctx.Ref("")
			focusedField := ctx.Ref(0)

			// Validation computed values
			emailValid := ctx.Computed(func() interface{} {
				e := email.GetTyped().(string)
				props := ctx.Props().(FormProps)
				if !props.RequireEmail {
					return true
				}
				return strings.Contains(e, "@") && len(e) > 3
			})

			passwordValid := ctx.Computed(func() interface{} {
				p := password.GetTyped().(string)
				props := ctx.Props().(FormProps)
				return len(p) >= props.MinPassword
			})

			formValid := ctx.Computed(func() interface{} {
				return emailValid.GetTyped().(bool) && passwordValid.GetTyped().(bool)
			})

			// Expose state
			ctx.Expose("email", email)
			ctx.Expose("password", password)
			ctx.Expose("focusedField", focusedField)
			ctx.Expose("emailValid", emailValid)
			ctx.Expose("passwordValid", passwordValid)
			ctx.Expose("formValid", formValid)

			// Event handlers
			ctx.On("typeEmail", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					char := event.Data.(string)
					email.Set(email.GetTyped().(string) + char)
				}
			})

			ctx.On("typePassword", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					char := event.Data.(string)
					password.Set(password.GetTyped().(string) + char)
				}
			})

			ctx.On("deleteEmail", func(data interface{}) {
				e := email.GetTyped().(string)
				if len(e) > 0 {
					email.Set(e[:len(e)-1])
				}
			})

			ctx.On("deletePassword", func(data interface{}) {
				p := password.GetTyped().(string)
				if len(p) > 0 {
					password.Set(p[:len(p)-1])
				}
			})

			ctx.On("focusField", func(data interface{}) {
				if event, ok := data.(*bubbly.Event); ok {
					focusedField.Set(event.Data.(int))
				}
			})

			ctx.On("reset", func(data interface{}) {
				email.Set("")
				password.Set("")
				focusedField.Set(0)
			})

			ctx.On("submit", func(data interface{}) {
				if formValid.GetTyped().(bool) {
					onSubmit(email.GetTyped().(string), password.GetTyped().(string))
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(FormProps)
			email := ctx.Get("email").(*bubbly.Ref[interface{}])
			password := ctx.Get("password").(*bubbly.Ref[interface{}])
			focusedField := ctx.Get("focusedField").(*bubbly.Ref[interface{}])
			emailValid := ctx.Get("emailValid").(*bubbly.Computed[interface{}])
			passwordValid := ctx.Get("passwordValid").(*bubbly.Computed[interface{}])
			formValid := ctx.Get("formValid").(*bubbly.Computed[interface{}])

			emailVal := email.GetTyped().(string)
			passwordVal := password.GetTyped().(string)
			focused := focusedField.GetTyped().(int)

			// Form container
			formStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1, 2).
				Width(50)

			// Calculate input field width to fill form
			// Form width (50) - left/right padding (2*2) - border (2) - field padding (2) = 42
			inputWidth := 42

			// Title
			formTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				Render(props.Title)

			// Email field
			emailStyle := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				Padding(0, 1).
				Width(inputWidth)

			if focused == 0 {
				emailStyle = emailStyle.BorderForeground(lipgloss.Color("86"))
			} else {
				emailStyle = emailStyle.BorderForeground(lipgloss.Color("240"))
			}

			emailLabel := "Email:"
			if !emailValid.GetTyped().(bool) && len(emailVal) > 0 {
				emailLabel += " ✗ Invalid"
				emailStyle = emailStyle.BorderForeground(lipgloss.Color("196"))
			} else if emailValid.GetTyped().(bool) && len(emailVal) > 0 {
				emailLabel += " ✓"
			}

			// Ensure consistent field width
			emailContent := fmt.Sprintf("%s\n%s", emailLabel, emailVal)
			if len(emailVal) == 0 {
				emailContent = fmt.Sprintf("%s\n ", emailLabel) // Add space for empty field
			}
			emailField := emailStyle.Render(emailContent)

			// Password field
			passwordStyle := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				Padding(0, 1).
				Width(inputWidth).
				MarginTop(1)

			if focused == 1 {
				passwordStyle = passwordStyle.BorderForeground(lipgloss.Color("86"))
			} else {
				passwordStyle = passwordStyle.BorderForeground(lipgloss.Color("240"))
			}

			passwordLabel := fmt.Sprintf("Password (min %d chars):", props.MinPassword)
			if !passwordValid.GetTyped().(bool) && len(passwordVal) > 0 {
				passwordLabel += " ✗ Too short"
				passwordStyle = passwordStyle.BorderForeground(lipgloss.Color("196"))
			} else if passwordValid.GetTyped().(bool) && len(passwordVal) > 0 {
				passwordLabel += " ✓"
			}

			// Ensure consistent field width
			passwordContent := fmt.Sprintf("%s\n%s", passwordLabel, strings.Repeat("*", len(passwordVal)))
			if len(passwordVal) == 0 {
				passwordContent = fmt.Sprintf("%s\n ", passwordLabel) // Add space for empty field
			}
			passwordField := passwordStyle.Render(passwordContent)

			// Submit button
			buttonStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				MarginTop(1)

			if formValid.GetTyped().(bool) {
				buttonStyle = buttonStyle.
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("28")).
					BorderForeground(lipgloss.Color("46"))
			} else {
				buttonStyle = buttonStyle.
					Foreground(lipgloss.Color("240")).
					BorderForeground(lipgloss.Color("240"))
			}

			button := buttonStyle.Render("[ Submit ]")

			return formStyle.Render(lipgloss.JoinVertical(
				lipgloss.Left,
				formTitle,
				"",
				emailField,
				passwordField,
				button,
			))
		}).
		Build()
}

func main() {
	m := model{
		currentField: 0,
		submitted:    false,
		submitData:   make(map[string]string),
	}

	form, err := createForm(
		FormProps{
			Title:        "Login Form",
			RequireEmail: true,
			MinPassword:  8,
		},
		func(email, password string) {
			m.submitted = true
			m.submitData["email"] = email
			m.submitData["password"] = password
		},
	)
	if err != nil {
		fmt.Printf("Error creating form: %v\n", err)
		os.Exit(1)
	}

	form.Init()
	m.form = form

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
