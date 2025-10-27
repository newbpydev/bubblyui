package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// model represents a registration form with reactive validation
type model struct {
	// Reactive form fields
	email           *bubbly.Ref[string]
	password        *bubbly.Ref[string]
	confirmPassword *bubbly.Ref[string]

	// Computed validation states
	emailValid     *bubbly.Computed[bool]
	passwordValid  *bubbly.Computed[bool]
	passwordsMatch *bubbly.Computed[bool]
	formValid      *bubbly.Computed[bool]

	// UI state
	focusedField int
	submitted    bool
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 3

		case "shift+tab", "up":
			m.focusedField = (m.focusedField + 2) % 3

		case "enter":
			// Try to submit if form is valid
			if m.formValid.Get() {
				m.submitted = true
				return m, tea.Quit
			}

		case "backspace":
			m.handleBackspace()

		default:
			// Add character to focused field
			if len(msg.String()) == 1 {
				m.handleInput(msg.String())
			}
		}
	}

	return m, nil
}

// handleInput adds a character to the focused field
func (m *model) handleInput(char string) {
	switch m.focusedField {
	case 0:
		m.email.Set(m.email.Get() + char)
	case 1:
		m.password.Set(m.password.Get() + char)
	case 2:
		m.confirmPassword.Set(m.confirmPassword.Get() + char)
	}
}

// handleBackspace removes a character from the focused field
func (m *model) handleBackspace() {
	switch m.focusedField {
	case 0:
		if val := m.email.Get(); len(val) > 0 {
			m.email.Set(val[:len(val)-1])
		}
	case 1:
		if val := m.password.Get(); len(val) > 0 {
			m.password.Set(val[:len(val)-1])
		}
	case 2:
		if val := m.confirmPassword.Get(); len(val) > 0 {
			m.confirmPassword.Set(val[:len(val)-1])
		}
	}
}

// View renders the UI
func (m model) View() string {
	if m.submitted {
		successStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(1, 2)

		return successStyle.Render("✓ Registration successful!\n\nEmail: " + m.email.Get())
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(40)

	focusedStyle := inputStyle.Copy().
		BorderForeground(lipgloss.Color("86"))

	validStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	invalidStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	submitStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(0, 2).
		MarginTop(1)

	disabledSubmitStyle := submitStyle.Copy().
		Foreground(lipgloss.Color("241")).
		BorderForeground(lipgloss.Color("241"))

	// Build view
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📋 Reactive Form Validation"))
	b.WriteString("\n\n")

	// Email field
	b.WriteString(labelStyle.Render("Email:"))
	b.WriteString("\n")
	emailInput := m.email.Get()
	if m.focusedField == 0 {
		emailInput += "█"
		b.WriteString(focusedStyle.Render(emailInput))
	} else {
		b.WriteString(inputStyle.Render(emailInput))
	}
	b.WriteString(" ")
	if m.email.Get() != "" {
		if m.emailValid.Get() {
			b.WriteString(validStyle.Render("✓"))
		} else {
			b.WriteString(invalidStyle.Render("✗ Invalid email"))
		}
	}
	b.WriteString("\n\n")

	// Password field
	b.WriteString(labelStyle.Render("Password:"))
	b.WriteString("\n")
	passwordDisplay := strings.Repeat("•", len(m.password.Get()))
	if m.focusedField == 1 {
		passwordDisplay += "█"
		b.WriteString(focusedStyle.Render(passwordDisplay))
	} else {
		b.WriteString(inputStyle.Render(passwordDisplay))
	}
	b.WriteString(" ")
	if m.password.Get() != "" {
		if m.passwordValid.Get() {
			b.WriteString(validStyle.Render("✓"))
		} else {
			b.WriteString(invalidStyle.Render("✗ Min 8 characters"))
		}
	}
	b.WriteString("\n\n")

	// Confirm password field
	b.WriteString(labelStyle.Render("Confirm Password:"))
	b.WriteString("\n")
	confirmDisplay := strings.Repeat("•", len(m.confirmPassword.Get()))
	if m.focusedField == 2 {
		confirmDisplay += "█"
		b.WriteString(focusedStyle.Render(confirmDisplay))
	} else {
		b.WriteString(inputStyle.Render(confirmDisplay))
	}
	b.WriteString(" ")
	if m.confirmPassword.Get() != "" {
		if m.passwordsMatch.Get() {
			b.WriteString(validStyle.Render("✓"))
		} else {
			b.WriteString(invalidStyle.Render("✗ Passwords don't match"))
		}
	}
	b.WriteString("\n\n")

	// Submit button - demonstrates reactive form validation!
	if m.formValid.Get() {
		b.WriteString(submitStyle.Render("Submit (Enter)"))
	} else {
		b.WriteString(disabledSubmitStyle.Render("Submit (Enter)"))
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(
		"tab/↓ next field • shift+tab/↑ previous • enter submit • q quit",
	))

	return b.String()
}

func main() {
	// Create reactive form fields
	email := bubbly.NewRef("")
	password := bubbly.NewRef("")
	confirmPassword := bubbly.NewRef("")

	// Email validation - demonstrates computed values with regex
	emailValid := bubbly.NewComputed(func() bool {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		return emailRegex.MatchString(email.Get())
	})

	// Password validation - min 8 characters
	passwordValid := bubbly.NewComputed(func() bool {
		return len(password.Get()) >= 8
	})

	// Password match validation
	passwordsMatch := bubbly.NewComputed(func() bool {
		return password.Get() != "" &&
			password.Get() == confirmPassword.Get()
	})

	// Overall form validation - demonstrates chaining computed values!
	formValid := bubbly.NewComputed(func() bool {
		return emailValid.Get() &&
			passwordValid.Get() &&
			passwordsMatch.Get()
	})

	// Create model
	m := model{
		email:           email,
		password:        password,
		confirmPassword: confirmPassword,
		emailValid:      emailValid,
		passwordValid:   passwordValid,
		passwordsMatch:  passwordsMatch,
		formValid:       formValid,
		focusedField:    0,
	}

	// Run the program with alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
