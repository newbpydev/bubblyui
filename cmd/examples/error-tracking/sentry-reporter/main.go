package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// model wraps the form component
type model struct {
	form bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			observability.RecordBreadcrumb("user", "User pressed tab to switch field", nil)
			m.form.Emit("next-field", nil)
		case "shift+tab":
			observability.RecordBreadcrumb("user", "User pressed shift+tab to go back", nil)
			m.form.Emit("prev-field", nil)
		case "enter":
			observability.RecordBreadcrumb("user", "User submitted form", map[string]interface{}{
				"action": "submit",
			})
			m.form.Emit("submit", nil)
		case "backspace":
			m.form.Emit("backspace", nil)
		case "e":
			// Trigger an error to demonstrate error tracking
			observability.RecordBreadcrumb("user", "User triggered error (for testing)", map[string]interface{}{
				"action": "error_test",
			})
			m.form.Emit("trigger-error", nil)
		case "p":
			// Trigger a panic to demonstrate panic tracking
			observability.RecordBreadcrumb("user", "User triggered panic (for testing)", map[string]interface{}{
				"action": "panic_test",
			})
			m.form.Emit("trigger-panic", nil)
		default:
			// Handle character input
			if len(msg.String()) == 1 {
				m.form.Emit("input", msg.String())
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

	title := titleStyle.Render("🚀 Error Tracking - Sentry Reporter (Production)")

	componentView := m.form.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"tab/shift+tab: switch fields • enter: submit • e: error test • p: panic test • q: quit",
	)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		Italic(true)

	info := infoStyle.Render(
		"💡 Using Sentry reporter with tags, extras, and breadcrumbs. Errors are sent to Sentry!",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n%s\n", title, componentView, help, info)
}

// createForm creates a form component with comprehensive error tracking
func createForm() (bubbly.Component, error) {
	return bubbly.NewComponent("RegistrationForm").
		Setup(func(ctx *bubbly.Context) {
			// Reactive state
			username := ctx.Ref("")
			email := ctx.Ref("")
			password := ctx.Ref("")
			currentField := ctx.Ref(0) // 0: username, 1: email, 2: password
			errorMessage := ctx.Ref("")
			submitCount := ctx.Ref(0)

			// Computed values
			isValid := ctx.Computed(func() interface{} {
				u := username.Get().(string)
				e := email.Get().(string)
				p := password.Get().(string)
				return len(u) >= 3 && len(e) >= 5 && len(p) >= 8
			})

			// Expose state
			ctx.Expose("username", username)
			ctx.Expose("email", email)
			ctx.Expose("password", password)
			ctx.Expose("currentField", currentField)
			ctx.Expose("errorMessage", errorMessage)
			ctx.Expose("submitCount", submitCount)
			ctx.Expose("isValid", isValid)

			// Record breadcrumb on component init
			observability.RecordBreadcrumb("component", "RegistrationForm component initialized", map[string]interface{}{
				"component": "RegistrationForm",
				"fields":    []string{"username", "email", "password"},
			})

			// Event handlers
			ctx.On("input", func(data interface{}) {
				event := data.(*bubbly.Event)
				char := event.Data.(string)
				field := currentField.Get().(int)

				switch field {
				case 0:
					username.Set(username.Get().(string) + char)
					observability.RecordBreadcrumb("state", "Username updated", map[string]interface{}{
						"field":  "username",
						"length": len(username.Get().(string)),
					})
				case 1:
					email.Set(email.Get().(string) + char)
					observability.RecordBreadcrumb("state", "Email updated", map[string]interface{}{
						"field":  "email",
						"length": len(email.Get().(string)),
					})
				case 2:
					password.Set(password.Get().(string) + char)
					observability.RecordBreadcrumb("state", "Password updated", map[string]interface{}{
						"field":  "password",
						"length": len(password.Get().(string)),
					})
				}
			})

			ctx.On("backspace", func(data interface{}) {
				field := currentField.Get().(int)

				switch field {
				case 0:
					u := username.Get().(string)
					if len(u) > 0 {
						username.Set(u[:len(u)-1])
					}
				case 1:
					e := email.Get().(string)
					if len(e) > 0 {
						email.Set(e[:len(e)-1])
					}
				case 2:
					p := password.Get().(string)
					if len(p) > 0 {
						password.Set(p[:len(p)-1])
					}
				}
			})

			ctx.On("next-field", func(data interface{}) {
				field := currentField.Get().(int)
				if field < 2 {
					currentField.Set(field + 1)
					observability.RecordBreadcrumb("navigation", "Moved to next field", map[string]interface{}{
						"from": field,
						"to":   field + 1,
					})
				}
			})

			ctx.On("prev-field", func(data interface{}) {
				field := currentField.Get().(int)
				if field > 0 {
					currentField.Set(field - 1)
					observability.RecordBreadcrumb("navigation", "Moved to previous field", map[string]interface{}{
						"from": field,
						"to":   field - 1,
					})
				}
			})

			ctx.On("submit", func(data interface{}) {
				valid := isValid.Get().(bool)
				count := submitCount.Get().(int)
				submitCount.Set(count + 1)

				observability.RecordBreadcrumb("user", "Form submission attempted", map[string]interface{}{
					"valid":        valid,
					"submitCount":  count + 1,
					"usernameLen":  len(username.Get().(string)),
					"emailLen":     len(email.Get().(string)),
					"passwordLen":  len(password.Get().(string)),
				})

				if !valid {
					errorMessage.Set("Validation failed: Check all fields")

					// Report validation error with full context
					if reporter := observability.GetErrorReporter(); reporter != nil {
						reporter.ReportError(
							fmt.Errorf("form validation failed"),
							&observability.ErrorContext{
								ComponentName: "RegistrationForm",
								ComponentID:   "form-1",
								EventName:     "submit",
								Timestamp:     time.Now(),
								Tags: map[string]string{
									"environment": "production",
									"form_type":   "registration",
									"valid":       "false",
								},
								Extra: map[string]interface{}{
									"username_length": len(username.Get().(string)),
									"email_length":    len(email.Get().(string)),
									"password_length": len(password.Get().(string)),
									"submit_count":    count + 1,
								},
								Breadcrumbs: observability.GetBreadcrumbs(),
								StackTrace:  debug.Stack(),
							},
						)
					}

					observability.RecordBreadcrumb("error", "Validation failed", map[string]interface{}{
						"reason": "invalid_fields",
					})
					return
				}

				errorMessage.Set("Success! Form submitted")
				observability.RecordBreadcrumb("state", "Form submitted successfully", map[string]interface{}{
					"username": username.Get().(string),
					"email":    email.Get().(string),
				})
			})

			ctx.On("trigger-error", func(data interface{}) {
				observability.RecordBreadcrumb("debug", "About to trigger error", map[string]interface{}{
					"intentional": true,
				})

				// Report a custom error with full context
				if reporter := observability.GetErrorReporter(); reporter != nil {
					reporter.ReportError(
						fmt.Errorf("intentional error for demonstration: invalid operation"),
						&observability.ErrorContext{
							ComponentName: "RegistrationForm",
							ComponentID:   "form-1",
							EventName:     "trigger-error",
							Timestamp:     time.Now(),
							Tags: map[string]string{
								"environment": "production",
								"error_type":  "intentional",
								"severity":    "warning",
							},
							Extra: map[string]interface{}{
								"test_mode":    true,
								"current_user": "demo_user",
								"form_state": map[string]interface{}{
									"username": username.Get().(string),
									"email":    email.Get().(string),
								},
							},
							Breadcrumbs: observability.GetBreadcrumbs(),
							StackTrace:  debug.Stack(),
						},
					)
				}

				errorMessage.Set("Error reported to Sentry!")
			})

			ctx.On("trigger-panic", func(data interface{}) {
				observability.RecordBreadcrumb("debug", "About to trigger panic", map[string]interface{}{
					"intentional": true,
					"form_state": map[string]interface{}{
						"username": username.Get().(string),
						"email":    email.Get().(string),
					},
				})

				// This will be caught by the component's panic recovery
				panic("Intentional panic for Sentry demonstration!")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			username := ctx.Get("username").(*bubbly.Ref[interface{}])
			email := ctx.Get("email").(*bubbly.Ref[interface{}])
			password := ctx.Get("password").(*bubbly.Ref[interface{}])
			currentField := ctx.Get("currentField").(*bubbly.Ref[interface{}])
			errorMessage := ctx.Get("errorMessage").(*bubbly.Ref[interface{}])
			isValid := ctx.Get("isValid").(*bubbly.Computed[interface{}])

			usernameVal := username.Get().(string)
			emailVal := email.Get().(string)
			passwordVal := password.Get().(string)
			fieldVal := currentField.Get().(int)
			errorVal := errorMessage.Get().(string)
			validVal := isValid.Get().(bool)

			// Field styles
			activeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(40)

			inactiveStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(0, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(40)

			// Username field
			usernameStyle := inactiveStyle
			if fieldVal == 0 {
				usernameStyle = activeStyle
			}
			usernameBox := usernameStyle.Render(fmt.Sprintf("Username: %s", usernameVal))

			// Email field
			emailStyle := inactiveStyle
			if fieldVal == 1 {
				emailStyle = activeStyle
			}
			emailBox := emailStyle.Render(fmt.Sprintf("Email: %s", emailVal))

			// Password field
			passwordStyle := inactiveStyle
			if fieldVal == 2 {
				passwordStyle = activeStyle
			}
			maskedPassword := ""
			for range passwordVal {
				maskedPassword += "*"
			}
			passwordBox := passwordStyle.Render(fmt.Sprintf("Password: %s", maskedPassword))

			// Validation status
			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Width(40)

			if !validVal {
				statusStyle = statusStyle.Foreground(lipgloss.Color("203"))
			}

			statusText := "✓ Valid"
			if !validVal {
				statusText = "✗ Invalid (min: user=3, email=5, pass=8)"
			}
			statusBox := statusStyle.Render(statusText)

			// Error message
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("203")).
				Padding(0, 2).
				Width(40)

			if errorVal == "Success! Form submitted" {
				errorStyle = errorStyle.Foreground(lipgloss.Color("86"))
			}

			errorBox := ""
			if errorVal != "" {
				errorBox = errorStyle.Render(errorVal)
			}

			// Breadcrumbs display
			breadcrumbs := observability.GetBreadcrumbs()
			breadcrumbStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(40).
				Height(6)

			breadcrumbText := "Recent Activity:\n"
			start := len(breadcrumbs) - 4
			if start < 0 {
				start = 0
			}
			for i := start; i < len(breadcrumbs); i++ {
				bc := breadcrumbs[i]
				breadcrumbText += fmt.Sprintf("• [%s] %s\n", bc.Category, bc.Message)
			}

			breadcrumbBox := breadcrumbStyle.Render(breadcrumbText)

			result := lipgloss.JoinVertical(
				lipgloss.Left,
				usernameBox,
				emailBox,
				passwordBox,
				"",
				statusBox,
			)

			if errorBox != "" {
				result = lipgloss.JoinVertical(lipgloss.Left, result, errorBox)
			}

			result = lipgloss.JoinVertical(lipgloss.Left, result, "", breadcrumbBox)

			return result
		}).
		Build()
}

func main() {
	// Setup Sentry reporter for production
	// In production, use: os.Getenv("SENTRY_DSN")
	// For this example, we use empty DSN (won't send to Sentry, but demonstrates the API)
	reporter, err := observability.NewSentryReporter(
		"", // Empty DSN for demo - replace with real DSN in production
		observability.WithEnvironment("production"),
		observability.WithRelease("v1.0.0"),
		observability.WithDebug(true),
	)
	if err != nil {
		fmt.Printf("Error creating Sentry reporter: %v\n", err)
		os.Exit(1)
	}

	observability.SetErrorReporter(reporter)
	defer reporter.Flush(5 * time.Second)

	// Record initial breadcrumb
	observability.RecordBreadcrumb("navigation", "Application started", map[string]interface{}{
		"example":     "sentry-reporter",
		"mode":        "production",
		"environment": "demo",
	})

	form, err := createForm()
	if err != nil {
		fmt.Printf("Error creating form: %v\n", err)
		os.Exit(1)
	}

	form.Init()

	m := model{form: form}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	observability.RecordBreadcrumb("navigation", "Application exited", nil)
}
