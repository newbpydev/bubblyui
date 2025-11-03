package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// UserRegistration represents our form data structure
type UserRegistration struct {
	Username        string
	Email           string
	Password        string
	ConfirmPassword string
	FullName        string
	Bio             string
	Newsletter      bool
	Terms           bool
	Theme           string
	Notifications   bool
}

// tickMsg for form reset countdown
type tickMsg time.Time

// model wraps the form builder component
type model struct {
	component       bubbly.Component
	inputMode       bool // Track if we're in input mode vs navigation mode
	currentField    int  // Track which field is currently focused
	countdownActive bool
	countdownSecs   int
}

func (m model) Init() tea.Cmd {
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.countdownActive {
			m.countdownSecs--
			if m.countdownSecs <= 0 {
				m.countdownActive = false
				m.component.Emit("resetForm", nil)
			} else {
				cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				}))
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.inputMode {
				return m, tea.Quit
			}
		case "esc":
			// Toggle between input and navigation modes
			m.inputMode = !m.inputMode
			m.component.Emit("setInputMode", m.inputMode)
			// Cancel countdown if active
			if m.countdownActive {
				m.countdownActive = false
			}
			return m, nil
		case "tab":
			// Navigate to next field
			m.component.Emit("nextField", nil)
			m.currentField = (m.currentField + 1) % 11 // 11 total fields
			// Cancel countdown if user is editing
			if m.countdownActive {
				m.countdownActive = false
			}
			return m, nil
		case "shift+tab":
			// Navigate to previous field
			m.component.Emit("prevField", nil)
			if m.currentField == 0 {
				m.currentField = 10
			} else {
				m.currentField--
			}
			// Cancel countdown if user is editing
			if m.countdownActive {
				m.countdownActive = false
			}
			return m, nil
		case "enter":
			if m.inputMode {
				// Submit form or toggle boolean fields
				if m.currentField == 6 || m.currentField == 7 || m.currentField == 9 {
					// Toggle checkboxes/toggles
					m.component.Emit("toggleField", m.currentField)
				} else {
					// Submit form
					m.component.Emit("submitForm", nil)
					// Check if form was valid and start countdown
					m.component.Emit("checkSubmitSuccess", nil)
				}
			} else {
				// Enter input mode
				m.inputMode = true
				m.component.Emit("setInputMode", true)
			}
			return m, nil
		case "space":
			if m.inputMode {
				// Space toggles checkboxes or adds space to text fields
				if m.currentField == 6 || m.currentField == 7 || m.currentField == 9 {
					m.component.Emit("toggleField", m.currentField)
				} else {
					m.component.Emit("handleInput", msg)
				}
			}
			// Cancel countdown if user is editing
			if m.countdownActive {
				m.countdownActive = false
			}
			return m, nil
		case "ctrl+r":
			// Manual reset
			m.countdownActive = false
			m.component.Emit("resetForm", nil)
			return m, nil
		default:
			if m.inputMode {
				// Forward to input handler
				m.component.Emit("handleInput", msg)
				// Cancel countdown if user is editing
				if m.countdownActive {
					m.countdownActive = false
				}
			}
		}
	}

	// Update component
	updated, cmd := m.component.Update(msg)
	m.component = updated.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Check if we should start countdown after successful submission
	if !m.countdownActive {
		// This is a simple way to check - in production you'd use proper message passing
		m.component.Emit("getSubmitStatus", func(success bool) {
			if success && !m.countdownActive {
				m.countdownActive = true
				m.countdownSecs = 3
				cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				}))
			}
		})
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

	modeStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)

	var modeIndicator string
	if m.inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		modeIndicator = modeStyle.Render("✍️  INPUT MODE")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("33"))
		modeIndicator = modeStyle.Render("🧭 NAVIGATION MODE")
	}

	title := titleStyle.Render("📝 Advanced Form Builder Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Complex form composition with validation, using all form components",
	)

	componentView := m.component.View()

	// Add countdown overlay if active
	if m.countdownActive && m.countdownSecs > 0 {
		countdownStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35")).
			Padding(1, 3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Width(70).
			Align(lipgloss.Center).
			MarginTop(1)

		countdownMsg := countdownStyle.Render(fmt.Sprintf(
			"✅ Form submitted successfully! Resetting in %d...",
			m.countdownSecs))

		componentView = componentView + "\n\n" + countdownMsg
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render(
		"esc: toggle mode • tab/shift+tab: navigate • enter: submit/toggle • space: toggle/type • ctrl+r: reset • q: quit (nav mode)",
	)

	return fmt.Sprintf("%s  %s\n%s\n\n%s\n\n%s",
		title,
		modeIndicator,
		subtitle,
		componentView,
		help,
	)
}

// validateUserRegistration validates the entire form
func validateUserRegistration(form UserRegistration) map[string]string {
	errors := make(map[string]string)

	// Username validation
	if len(form.Username) == 0 {
		errors["Username"] = "Username is required"
	} else if len(form.Username) < 3 {
		errors["Username"] = "Username must be at least 3 characters"
	} else if strings.Contains(form.Username, " ") {
		errors["Username"] = "Username cannot contain spaces"
	}

	// Email validation
	if len(form.Email) == 0 {
		errors["Email"] = "Email is required"
	} else if !strings.Contains(form.Email, "@") || !strings.Contains(form.Email, ".") {
		errors["Email"] = "Please enter a valid email address"
	}

	// Password validation
	if len(form.Password) == 0 {
		errors["Password"] = "Password is required"
	} else if len(form.Password) < 8 {
		errors["Password"] = "Password must be at least 8 characters"
	}

	// Confirm password validation
	if form.ConfirmPassword != form.Password {
		errors["ConfirmPassword"] = "Passwords do not match"
	}

	// Full name validation
	if len(form.FullName) == 0 {
		errors["FullName"] = "Full name is required"
	}

	// Terms acceptance
	if !form.Terms {
		errors["Terms"] = "You must accept the terms and conditions"
	}

	return errors
}

// createFormBuilder creates the comprehensive form builder component
func createFormBuilder() (bubbly.Component, error) {
	return bubbly.NewComponent("FormBuilder").
		Setup(func(ctx *bubbly.Context) {
			// Provide custom theme
			theme := components.Theme{
				Primary:    lipgloss.Color("205"),
				Secondary:  lipgloss.Color("33"),
				Success:    lipgloss.Color("35"),
				Danger:     lipgloss.Color("196"),
				Warning:    lipgloss.Color("220"),
				Foreground: lipgloss.Color("15"),
				Muted:      lipgloss.Color("241"),
				Background: lipgloss.Color("0"),
			}
			ctx.Provide("theme", theme)

			// Form state using typed refs
			username := bubbly.NewRef("")
			email := bubbly.NewRef("")
			password := bubbly.NewRef("")
			confirmPassword := bubbly.NewRef("")
			fullName := bubbly.NewRef("")
			bio := bubbly.NewRef("")
			newsletter := bubbly.NewRef(false)
			terms := bubbly.NewRef(false)
			selectedTheme := bubbly.NewRef("Light")
			notifications := bubbly.NewRef(true)

			// UI state
			currentField := bubbly.NewRef(0)
			inputMode := bubbly.NewRef(false)
			formErrors := bubbly.NewRef(make(map[string]string))
			submitAttempts := bubbly.NewRef(0)
			lastSubmitSuccess := bubbly.NewRef(false)

			// Expose all state
			ctx.Expose("username", username)
			ctx.Expose("email", email)
			ctx.Expose("password", password)
			ctx.Expose("confirmPassword", confirmPassword)
			ctx.Expose("fullName", fullName)
			ctx.Expose("bio", bio)
			ctx.Expose("newsletter", newsletter)
			ctx.Expose("terms", terms)
			ctx.Expose("selectedTheme", selectedTheme)
			ctx.Expose("notifications", notifications)
			ctx.Expose("currentField", currentField)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("formErrors", formErrors)
			ctx.Expose("submitAttempts", submitAttempts)
			ctx.Expose("lastSubmitSuccess", lastSubmitSuccess)

			// Create all form components and expose them
			// We'll create them in the template for proper reactivity

			// Event handlers
			ctx.On("setInputMode", func(data interface{}) {
				inputMode.Set(data.(bool))
			})

			ctx.On("nextField", func(_ interface{}) {
				current := currentField.Get().(int)
				currentField.Set((current + 1) % 11) // 11 fields total
			})

			ctx.On("prevField", func(_ interface{}) {
				current := currentField.Get().(int)
				if current == 0 {
					currentField.Set(10)
				} else {
					currentField.Set(current - 1)
				}
			})

			ctx.On("handleInput", func(data interface{}) {
				if msg, ok := data.(tea.KeyMsg); ok {
					field := currentField.Get().(int)

					switch field {
					case 0: // Username
						current := username.Get().(string)
						switch msg.Type {
						case tea.KeyRunes:
							username.Set(current + string(msg.Runes))
						case tea.KeyBackspace:
							if len(current) > 0 {
								username.Set(current[:len(current)-1])
							}
						case tea.KeySpace:
							// Don't allow spaces in username
						}
					case 1: // Email
						current := email.Get().(string)
						switch msg.Type {
						case tea.KeyRunes:
							email.Set(current + string(msg.Runes))
						case tea.KeyBackspace:
							if len(current) > 0 {
								email.Set(current[:len(current)-1])
							}
						case tea.KeySpace:
							// Don't allow spaces in email
						}
					case 2: // Password
						current := password.Get().(string)
						switch msg.Type {
						case tea.KeyRunes:
							password.Set(current + string(msg.Runes))
						case tea.KeyBackspace:
							if len(current) > 0 {
								password.Set(current[:len(current)-1])
							}
						case tea.KeySpace:
							password.Set(current + " ")
						}
					case 3: // Confirm Password
						current := confirmPassword.Get().(string)
						switch msg.Type {
						case tea.KeyRunes:
							confirmPassword.Set(current + string(msg.Runes))
						case tea.KeyBackspace:
							if len(current) > 0 {
								confirmPassword.Set(current[:len(current)-1])
							}
						case tea.KeySpace:
							confirmPassword.Set(current + " ")
						}
					case 4: // Full Name
						current := fullName.Get().(string)
						switch msg.Type {
						case tea.KeyRunes:
							fullName.Set(current + string(msg.Runes))
						case tea.KeyBackspace:
							if len(current) > 0 {
								fullName.Set(current[:len(current)-1])
							}
						case tea.KeySpace:
							fullName.Set(current + " ")
						}
					case 5: // Bio
						current := bio.Get().(string)
						switch msg.Type {
						case tea.KeyRunes:
							bio.Set(current + string(msg.Runes))
						case tea.KeyBackspace:
							if len(current) > 0 {
								bio.Set(current[:len(current)-1])
							}
						case tea.KeySpace:
							bio.Set(current + " ")
						case tea.KeyEnter:
							bio.Set(current + "\n")
						}
					}
				}
			})

			ctx.On("toggleField", func(data interface{}) {
				field := data.(int)
				switch field {
				case 6: // Newsletter checkbox
					newsletter.Set(!newsletter.Get().(bool))
				case 7: // Terms checkbox
					terms.Set(!terms.Get().(bool))
				case 9: // Notifications toggle
					notifications.Set(!notifications.Get().(bool))
				}
			})

			ctx.On("submitForm", func(_ interface{}) {
				// Validate form
				form := UserRegistration{
					Username:        username.Get().(string),
					Email:           email.Get().(string),
					Password:        password.Get().(string),
					ConfirmPassword: confirmPassword.Get().(string),
					FullName:        fullName.Get().(string),
					Bio:             bio.Get().(string),
					Newsletter:      newsletter.Get().(bool),
					Terms:           terms.Get().(bool),
					Theme:           selectedTheme.Get().(string),
					Notifications:   notifications.Get().(bool),
				}

				errors := validateUserRegistration(form)
				formErrors.Set(errors)

				if len(errors) == 0 {
					// Form is valid
					attempts := submitAttempts.Get().(int)
					submitAttempts.Set(attempts + 1)
					lastSubmitSuccess.Set(true)
				} else {
					// Form has errors
					lastSubmitSuccess.Set(false)
				}
			})

			ctx.On("resetForm", func(_ interface{}) {
				// Reset all fields
				username.Set("")
				email.Set("")
				password.Set("")
				confirmPassword.Set("")
				fullName.Set("")
				bio.Set("")
				newsletter.Set(false)
				terms.Set(false)
				selectedTheme.Set("Light")
				notifications.Set(true)
				formErrors.Set(make(map[string]string))
				submitAttempts.Set(0)
				lastSubmitSuccess.Set(false)
				currentField.Set(0)
			})

			ctx.On("getSubmitStatus", func(callback interface{}) {
				if cb, ok := callback.(func(bool)); ok {
					cb(lastSubmitSuccess.Get().(bool) && submitAttempts.Get().(int) > 0)
				}
			})

			ctx.On("checkSubmitSuccess", func(_ interface{}) {
				// This would trigger countdown check in Update
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get all state
			username := ctx.Get("username").(*bubbly.Ref[string])
			email := ctx.Get("email").(*bubbly.Ref[string])
			password := ctx.Get("password").(*bubbly.Ref[string])
			confirmPassword := ctx.Get("confirmPassword").(*bubbly.Ref[string])
			fullName := ctx.Get("fullName").(*bubbly.Ref[string])
			bio := ctx.Get("bio").(*bubbly.Ref[string])
			newsletter := ctx.Get("newsletter").(*bubbly.Ref[bool])
			terms := ctx.Get("terms").(*bubbly.Ref[bool])
			selectedTheme := ctx.Get("selectedTheme").(*bubbly.Ref[string])
			notifications := ctx.Get("notifications").(*bubbly.Ref[bool])
			currentField := ctx.Get("currentField").(*bubbly.Ref[int]).Get().(int)
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[bool]).Get().(bool)
			formErrors := ctx.Get("formErrors").(*bubbly.Ref[map[string]string]).Get().(map[string]string)
			submitAttempts := ctx.Get("submitAttempts").(*bubbly.Ref[int]).Get().(int)

			// Create form layout using Form component
			form := components.Form(components.FormProps[UserRegistration]{
				Initial: UserRegistration{
					Username:        username.Get().(string),
					Email:           email.Get().(string),
					Password:        password.Get().(string),
					ConfirmPassword: confirmPassword.Get().(string),
					FullName:        fullName.Get().(string),
					Bio:             bio.Get().(string),
					Newsletter:      newsletter.Get().(bool),
					Terms:           terms.Get().(bool),
					Theme:           selectedTheme.Get().(string),
					Notifications:   notifications.Get().(bool),
				},
				Validate: validateUserRegistration,
				OnSubmit: func(data UserRegistration) {
					// Handle submission
				},
			})
			form.Init()

			// Build the form UI manually for better control
			formStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(func() lipgloss.Color {
					if inputMode {
						return lipgloss.Color("35")
					}
					return lipgloss.Color("240")
				}()).
				Padding(1).
				Width(70)

			fieldStyle := lipgloss.NewStyle().
				PaddingLeft(2)

			activeFieldStyle := fieldStyle.Copy().
				Foreground(lipgloss.Color("205")).
				Bold(true)

			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				PaddingLeft(4)

			var fields []string

			// Helper to render field with focus indicator
			renderField := func(index int, label, value string, err string) string {
				style := fieldStyle
				prefix := "  "
				if currentField == index {
					style = activeFieldStyle
					prefix = "▶ "
				}

				field := prefix + label + ": " + value
				if err != "" {
					field += "\n" + errorStyle.Render("❌ "+err)
				}
				return style.Render(field)
			}

			// Username field (0)
			usernameVal := username.Get().(string)
			if usernameVal == "" {
				usernameVal = "(empty)"
			}
			fields = append(fields, renderField(0, "Username", usernameVal, formErrors["Username"]))

			// Email field (1)
			emailVal := email.Get().(string)
			if emailVal == "" {
				emailVal = "(empty)"
			}
			fields = append(fields, renderField(1, "Email", emailVal, formErrors["Email"]))

			// Password field (2)
			passwordVal := strings.Repeat("*", len(password.Get().(string)))
			if passwordVal == "" {
				passwordVal = "(empty)"
			}
			fields = append(fields, renderField(2, "Password", passwordVal, formErrors["Password"]))

			// Confirm Password field (3)
			confirmVal := strings.Repeat("*", len(confirmPassword.Get().(string)))
			if confirmVal == "" {
				confirmVal = "(empty)"
			}
			fields = append(fields, renderField(3, "Confirm Password", confirmVal, formErrors["ConfirmPassword"]))

			// Full Name field (4)
			fullNameVal := fullName.Get().(string)
			if fullNameVal == "" {
				fullNameVal = "(empty)"
			}
			fields = append(fields, renderField(4, "Full Name", fullNameVal, formErrors["FullName"]))

			// Bio field (5)
			bioVal := bio.Get().(string)
			if bioVal == "" {
				bioVal = "(empty)"
			} else if len(bioVal) > 50 {
				bioVal = bioVal[:50] + "..."
			}
			fields = append(fields, renderField(5, "Bio", bioVal, ""))

			// Newsletter checkbox (6)
			newsletterVal := "☐ Subscribe to newsletter"
			if newsletter.Get().(bool) {
				newsletterVal = "☑ Subscribe to newsletter"
			}
			fields = append(fields, renderField(6, "", newsletterVal, ""))

			// Terms checkbox (7)
			termsVal := "☐ I accept the terms and conditions"
			if terms.Get().(bool) {
				termsVal = "☑ I accept the terms and conditions"
			}
			fields = append(fields, renderField(7, "", termsVal, formErrors["Terms"]))

			// Theme select (8)
			themeVal := selectedTheme.Get().(string)
			fields = append(fields, renderField(8, "Theme", themeVal, ""))

			// Notifications toggle (9)
			notifVal := "OFF"
			if notifications.Get().(bool) {
				notifVal = "ON"
			}
			fields = append(fields, renderField(9, "Notifications", notifVal, ""))

			// Submit button (10)
			submitLabel := "[Submit Registration]"
			if currentField == 10 {
				submitLabel = "▶ " + submitLabel
			}
			fields = append(fields, "", renderField(10, "", submitLabel, ""))

			// Status bar
			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Width(70).
				Align(lipgloss.Center)

			var statusBar string
			if submitAttempts > 0 && len(formErrors) == 0 {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35"))
				statusBar = statusStyle.Render(fmt.Sprintf("✅ Registration successful! (Submissions: %d)", submitAttempts))
			} else if len(formErrors) > 0 {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("196"))
				statusBar = statusStyle.Render(fmt.Sprintf("❌ %d validation error(s)", len(formErrors)))
			} else {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("241"))
				statusBar = statusStyle.Render("Complete all fields and submit")
			}

			formContent := strings.Join(fields, "\n")

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statusBar,
				"",
				formStyle.Render(formContent),
			)
		}).
		Build()
}

func main() {
	// Create the form builder component
	formBuilder, err := createFormBuilder()
	if err != nil {
		fmt.Printf("Error creating form builder: %v\n", err)
		os.Exit(1)
	}

	// Create model
	m := model{
		component:    formBuilder,
		inputMode:    false,
		currentField: 0,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
