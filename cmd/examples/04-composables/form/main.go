package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// LoginForm represents the form data structure
type LoginForm struct {
	Username string
	Email    string
	Password string
}

// resetMsg is sent after countdown to reset the form
type resetMsg struct{}

// tickMsg is sent periodically for countdown
type tickMsg time.Time

// model wraps the form component
type model struct {
	component       bubbly.Component
	focusedField    string // Track which field is being edited
	countdownActive bool   // Track if countdown is active
	countdownSecs   int    // Seconds remaining in countdown
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Handle countdown tick
		if m.countdownActive {
			m.countdownSecs--
			// Update component's countdown state
			m.component.Emit("updateCountdown", m.countdownSecs)

			if m.countdownSecs <= 0 {
				// Countdown finished - reset form
				m.countdownActive = false
				m.component.Emit("reset", nil)
				updatedComponent, cmd := m.component.Update(msg)
				m.component = updatedComponent.(bubbly.Component)
				return m, cmd
			}
			// Continue countdown
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			// Cycle through fields
			m.component.Emit("nextField", nil)
		case "enter":
			// Submit form
			m.component.Emit("submit", nil)
			// Check if submission was successful to start countdown
			// This will be handled by the component's response
		case "ctrl+r":
			// Reset form
			m.component.Emit("reset", nil)
		default:
			// Handle text input
			switch msg.Type {
			case tea.KeyRunes:
				// Regular character input
				m.component.Emit("addChar", string(msg.Runes))
			case tea.KeySpace:
				m.component.Emit("addChar", " ")
			case tea.KeyBackspace:
				m.component.Emit("removeChar", nil)
			}
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Check if we need to start countdown after Enter key (only if form was valid)
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" && !m.countdownActive {
		// The component has set justSubmitted=true if form was valid
		// We start the countdown timer here
		m.countdownActive = true
		m.countdownSecs = 3
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}))
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

	title := titleStyle.Render("📝 Composables - Form Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: UseForm composable for complex state management with validation",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"tab: next field • enter: submit • ctrl+r: reset • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// validateLoginForm validates the login form
func validateLoginForm(form LoginForm) map[string]string {
	errors := make(map[string]string)

	// Username validation
	if len(form.Username) == 0 {
		errors["Username"] = "Username is required"
	} else if len(form.Username) < 3 {
		errors["Username"] = "Username must be at least 3 characters"
	}

	// Email validation
	if len(form.Email) == 0 {
		errors["Email"] = "Email is required"
	} else if !strings.Contains(form.Email, "@") {
		errors["Email"] = "Email must contain @"
	}

	// Password validation
	if len(form.Password) == 0 {
		errors["Password"] = "Password is required"
	} else if len(form.Password) < 6 {
		errors["Password"] = "Password must be at least 6 characters"
	}

	return errors
}

// createFormDemo creates a component demonstrating UseForm
func createFormDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("FormDemo").
		Setup(func(ctx *bubbly.Context) {
			// Use the UseForm composable for form state management
			// This handles validation, dirty state, and field updates
			form := composables.UseForm(ctx, LoginForm{}, validateLoginForm)

			// Track which field is focused
			focusedField := ctx.Ref("Username")

			// Track submission attempts
			submitAttempts := ctx.Ref(0)
			lastSubmitSuccess := ctx.Ref(false)

			// Expose state to template
			ctx.Expose("form", form)
			ctx.Expose("focusedField", focusedField)
			ctx.Expose("submitAttempts", submitAttempts)
			ctx.Expose("lastSubmitSuccess", lastSubmitSuccess)

			// Event handler for field navigation
			ctx.On("nextField", func(_ interface{}) {
				current := focusedField.GetTyped().(string)
				switch current {
				case "Username":
					focusedField.Set("Email")
				case "Email":
					focusedField.Set("Password")
				case "Password":
					focusedField.Set("Username")
				}
			})

			// Event handler for adding characters
			ctx.On("addChar", func(data interface{}) {
				char := data.(string)
				field := focusedField.GetTyped().(string)
				currentForm := form.Values.GetTyped()

				// Get current field value and append character
				var newValue string
				switch field {
				case "Username":
					newValue = currentForm.Username + char
				case "Email":
					newValue = currentForm.Email + char
				case "Password":
					newValue = currentForm.Password + char
				}

				// SetField expects field name and field value (not entire form)
				form.SetField(field, newValue)
			})

			// Event handler for removing characters
			ctx.On("removeChar", func(_ interface{}) {
				field := focusedField.GetTyped().(string)
				currentForm := form.Values.GetTyped()

				// Get current field value and remove last character
				var newValue string
				switch field {
				case "Username":
					if len(currentForm.Username) > 0 {
						newValue = currentForm.Username[:len(currentForm.Username)-1]
					}
				case "Email":
					if len(currentForm.Email) > 0 {
						newValue = currentForm.Email[:len(currentForm.Email)-1]
					}
				case "Password":
					if len(currentForm.Password) > 0 {
						newValue = currentForm.Password[:len(currentForm.Password)-1]
					}
				}

				// SetField expects field name and field value (not entire form)
				form.SetField(field, newValue)
			})

			// Track countdown state
			countdownActive := ctx.Ref(false)
			countdownSecs := ctx.Ref(0)

			ctx.Expose("countdownActive", countdownActive)
			ctx.Expose("countdownSecs", countdownSecs)

			// Track if last submit was successful (for countdown trigger)
			justSubmitted := ctx.Ref(false)
			ctx.Expose("justSubmitted", justSubmitted)

			// Event handler for form submission
			ctx.On("submit", func(_ interface{}) {
				// Validate form before submission
				form.Submit()

				// Only count as submission attempt if form is valid
				if form.IsValid.GetTyped() {
					attempts := submitAttempts.GetTyped().(int)
					submitAttempts.Set(attempts + 1)
					lastSubmitSuccess.Set(true)
					justSubmitted.Set(true)

					// Start countdown for auto-reset
					countdownActive.Set(true)
					countdownSecs.Set(3)
				} else {
					// Form has errors - don't count as submission
					lastSubmitSuccess.Set(false)
					countdownActive.Set(false)
					justSubmitted.Set(false)
				}
			})

			// Event handler for updating countdown
			ctx.On("updateCountdown", func(data interface{}) {
				if secs, ok := data.(int); ok {
					countdownSecs.Set(secs)
				}
			})

			// Event handler for form reset
			ctx.On("reset", func(_ interface{}) {
				form.Reset()
				focusedField.Set("Username")
				lastSubmitSuccess.Set(false)
				submitAttempts.Set(0)
				countdownActive.Set(false)
				countdownSecs.Set(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			form := ctx.Get("form").(composables.UseFormReturn[LoginForm])
			focusedField := ctx.Get("focusedField").(*bubbly.Ref[interface{}])
			submitAttempts := ctx.Get("submitAttempts").(*bubbly.Ref[interface{}])
			countdownActive := ctx.Get("countdownActive").(*bubbly.Ref[interface{}])
			countdownSecs := ctx.Get("countdownSecs").(*bubbly.Ref[interface{}])

			currentForm := form.Values.GetTyped()
			errors := form.Errors.GetTyped()
			isDirty := form.IsDirty.GetTyped()
			isValid := form.IsValid.GetTyped()
			focused := focusedField.GetTyped().(string)
			attempts := submitAttempts.GetTyped().(int)
			countdown := countdownActive.GetTyped().(bool)
			secsLeft := countdownSecs.GetTyped().(int)

			// Form fields box
			fieldStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				Width(60)

			// Build form fields
			var fields []string

			// Username field
			usernameStyle := fieldStyle.Copy()
			if focused == "Username" {
				usernameStyle = usernameStyle.BorderForeground(lipgloss.Color("99"))
			} else {
				usernameStyle = usernameStyle.BorderForeground(lipgloss.Color("240"))
			}
			usernameLabel := "Username:"
			if focused == "Username" {
				usernameLabel = "▶ " + usernameLabel
			} else {
				usernameLabel = "  " + usernameLabel
			}
			usernameValue := currentForm.Username
			if usernameValue == "" {
				usernameValue = "(empty)"
			}
			usernameError := ""
			if err, ok := errors["Username"]; ok {
				usernameError = "\n❌ " + err
			}
			fields = append(fields, usernameStyle.Render(
				usernameLabel+"\n"+usernameValue+usernameError,
			))

			// Email field
			emailStyle := fieldStyle.Copy()
			if focused == "Email" {
				emailStyle = emailStyle.BorderForeground(lipgloss.Color("99"))
			} else {
				emailStyle = emailStyle.BorderForeground(lipgloss.Color("240"))
			}
			emailLabel := "Email:"
			if focused == "Email" {
				emailLabel = "▶ " + emailLabel
			} else {
				emailLabel = "  " + emailLabel
			}
			emailValue := currentForm.Email
			if emailValue == "" {
				emailValue = "(empty)"
			}
			emailError := ""
			if err, ok := errors["Email"]; ok {
				emailError = "\n❌ " + err
			}
			fields = append(fields, emailStyle.Render(
				emailLabel+"\n"+emailValue+emailError,
			))

			// Password field
			passwordStyle := fieldStyle.Copy()
			if focused == "Password" {
				passwordStyle = passwordStyle.BorderForeground(lipgloss.Color("99"))
			} else {
				passwordStyle = passwordStyle.BorderForeground(lipgloss.Color("240"))
			}
			passwordLabel := "Password:"
			if focused == "Password" {
				passwordLabel = "▶ " + passwordLabel
			} else {
				passwordLabel = "  " + passwordLabel
			}
			passwordValue := strings.Repeat("*", len(currentForm.Password))
			if passwordValue == "" {
				passwordValue = "(empty)"
			}
			passwordError := ""
			if err, ok := errors["Password"]; ok {
				passwordError = "\n❌ " + err
			}
			fields = append(fields, passwordStyle.Render(
				passwordLabel+"\n"+passwordValue+passwordError,
			))

			// Status box
			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				Width(60).
				Align(lipgloss.Center)

			var statusBox string
			if countdown && attempts > 0 {
				// Form was successfully submitted - show countdown
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					BorderForeground(lipgloss.Color("99"))
				statusBox = statusStyle.Render(fmt.Sprintf(
					"✅ Form submitted successfully! (Submissions: %d)\nResetting in %d seconds...",
					attempts, secsLeft))
			} else if !isValid && len(errors) > 0 {
				// Form has validation errors
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("196")).
					BorderForeground(lipgloss.Color("160"))
				statusBox = statusStyle.Render(fmt.Sprintf("❌ Cannot submit: Fix %d validation error(s)", len(errors)))
			} else if isValid && isDirty {
				// Form is valid and ready to submit
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					BorderForeground(lipgloss.Color("99"))
				statusBox = statusStyle.Render("✓ Form is valid - Press Enter to submit")
			} else {
				// Initial state - form not yet filled
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("241")).
					BorderForeground(lipgloss.Color("240"))
				statusBox = statusStyle.Render("Fill out the form and press Enter to submit")
			}

			// Info box
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(60)

			infoBox := infoStyle.Render(fmt.Sprintf(
				"UseForm State:\n\n"+
					"Dirty:   %v\n"+
					"Valid:   %v\n"+
					"Errors:  %d\n"+
					"Submits: %d\n\n"+
					"Pattern:\n"+
					"• Automatic validation\n"+
					"• Dirty state tracking\n"+
					"• Type-safe field updates",
				isDirty,
				isValid,
				len(errors),
				attempts,
			))

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statusBox,
				"",
				fields[0],
				"",
				fields[1],
				"",
				fields[2],
				"",
				infoBox,
			)
		}).
		Build()
}

func main() {
	component, err := createFormDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{component: component, focusedField: "Username"}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
