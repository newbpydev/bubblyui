package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	localcomponents "github.com/newbpydev/bubblyui/cmd/examples/10-testing/03-form/components"
	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/03-form/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root RegistrationApp component
// Demonstrates BubblyUI's composable component architecture with forms
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("RegistrationApp").
		WithAutoCommands(true).
		WithKeyBinding("tab", "focusNext", "Next field").
		WithKeyBinding("shift+tab", "focusPrevious", "Previous field").
		WithKeyBinding("enter", "submit", "Submit form").
		WithKeyBinding("esc", "toggleMode", "Toggle mode").
		WithKeyBinding("r", "reset", "Reset form").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		Setup(func(ctx *bubbly.Context) {
			// Initialize registration composable (testable business logic)
			registration := composables.UseRegistration(ctx)

			// Mode state: false = Navigation, true = Input
			inputMode := ctx.Ref(false)
			submitted := ctx.Ref(false)

			// PROVIDE theme colors to descendants (BubblyUI Provide/Inject pattern!)
			ctx.Provide("focusColor", lipgloss.Color("35"))     // Green
			ctx.Provide("inactiveColor", lipgloss.Color("240")) // Dark grey
			ctx.Provide("errorColor", lipgloss.Color("196"))    // Red

			// Create refs for each field value (typed refs!)
			nameValue := bubbly.NewRef("")
			emailValue := bubbly.NewRef("")
			passwordValue := bubbly.NewRef("")
			confirmValue := bubbly.NewRef("")

			// Create focused state refs for each field
			nameFocused := ctx.Ref(false)
			emailFocused := ctx.Ref(false)
			passwordFocused := ctx.Ref(false)
			confirmFocused := ctx.Ref(false)

			// Create error refs for each field
			nameError := ctx.Ref("")
			emailError := ctx.Ref("")
			passwordError := ctx.Ref("")
			confirmError := ctx.Ref("")

			// Sync field values to form (UseForm composable)
			bubbly.Watch(nameValue, func(newVal, oldVal string) {
				registration.Form.SetField("Name", newVal)
			})
			bubbly.Watch(emailValue, func(newVal, oldVal string) {
				registration.Form.SetField("Email", newVal)
			})
			bubbly.Watch(passwordValue, func(newVal, oldVal string) {
				registration.Form.SetField("Password", newVal)
			})
			bubbly.Watch(confirmValue, func(newVal string, oldVal string) {
				registration.Form.SetField("ConfirmPassword", newVal)
			})

			// Watch form errors and update field error refs
			bubbly.Watch(registration.Form.Errors, func(newVal map[string]string, oldVal map[string]string) {
				nameError.Set(newVal["Name"])
				emailError.Set(newVal["Email"])
				passwordError.Set(newVal["Password"])
				confirmError.Set(newVal["ConfirmPassword"])
			})

			// Watch focused field and update individual focused states
			ctx.Watch(registration.FocusedField, func(newVal, oldVal interface{}) {
				current := newVal.(string)
				nameFocused.Set(current == "name")
				emailFocused.Set(current == "email")
				passwordFocused.Set(current == "password")
				confirmFocused.Set(current == "confirm")

				// Emit focus/blur to field components
				if nameField := ctx.Get("nameField"); nameField != nil {
					comp := nameField.(bubbly.Component)
					if current == "name" {
						comp.Emit("setFocus", nil)
					} else {
						comp.Emit("setBlur", nil)
					}
				}
				if emailField := ctx.Get("emailField"); emailField != nil {
					comp := emailField.(bubbly.Component)
					if current == "email" {
						comp.Emit("setFocus", nil)
					} else {
						comp.Emit("setBlur", nil)
					}
				}
				if passwordField := ctx.Get("passwordField"); passwordField != nil {
					comp := passwordField.(bubbly.Component)
					if current == "password" {
						comp.Emit("setFocus", nil)
					} else {
						comp.Emit("setBlur", nil)
					}
				}
				if confirmField := ctx.Get("confirmField"); confirmField != nil {
					comp := confirmField.(bubbly.Component)
					if current == "confirm" {
						comp.Emit("setFocus", nil)
					} else {
						comp.Emit("setBlur", nil)
					}
				}
			})

			// Create FormField components
			nameField, err := localcomponents.CreateFormField(localcomponents.FormFieldProps{
				Label:       "Name",
				Value:       nameValue,
				Placeholder: "John Doe",
				Focused:     nameFocused,
				Error:       nameError,
				Width:       40,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create name field: %v", err))
				return
			}

			emailField, err := localcomponents.CreateFormField(localcomponents.FormFieldProps{
				Label:       "Email",
				Value:       emailValue,
				Placeholder: "john@example.com",
				Focused:     emailFocused,
				Error:       emailError,
				Width:       40,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create email field: %v", err))
				return
			}

			passwordField, err := localcomponents.CreateFormField(localcomponents.FormFieldProps{
				Label:       "Password",
				Value:       passwordValue,
				Placeholder: "Enter password",
				Focused:     passwordFocused,
				Error:       passwordError,
				Width:       40,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create password field: %v", err))
				return
			}

			confirmField, err := localcomponents.CreateFormField(localcomponents.FormFieldProps{
				Label:       "Confirm Password",
				Value:       confirmValue,
				Placeholder: "Confirm password",
				Focused:     confirmFocused,
				Error:       confirmError,
				Width:       40,
			})
			if err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to create confirm field: %v", err))
				return
			}

			// Expose components for template
			if err := ctx.ExposeComponent("nameField", nameField); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose name field: %v", err))
				return
			}
			if err := ctx.ExposeComponent("emailField", emailField); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose email field: %v", err))
				return
			}
			if err := ctx.ExposeComponent("passwordField", passwordField); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose password field: %v", err))
				return
			}
			if err := ctx.ExposeComponent("confirmField", confirmField); err != nil {
				ctx.Expose("error", fmt.Sprintf("Failed to expose confirm field: %v", err))
				return
			}

			// Expose state for tests
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("submitted", submitted)
			ctx.Expose("registration", registration)
			ctx.Expose("nameValue", nameValue)
			ctx.Expose("emailValue", emailValue)
			ctx.Expose("passwordValue", passwordValue)
			ctx.Expose("confirmValue", confirmValue)

			// Event: Toggle mode (ESC key)
			ctx.On("toggleMode", func(data interface{}) {
				current := inputMode.Get().(bool)
				newMode := !current
				inputMode.Set(newMode)

				// When entering input mode, focus first field
				if newMode {
					registration.FocusField("name")
				} else {
					registration.FocusField("")
				}
			})

			// Event: Focus next field (Tab key) - Input mode only
			ctx.On("focusNext", func(data interface{}) {
				if !inputMode.Get().(bool) {
					return
				}
				registration.FocusNext()
			})

			// Event: Focus previous field (Shift+Tab) - Input mode only
			ctx.On("focusPrevious", func(data interface{}) {
				if !inputMode.Get().(bool) {
					return
				}
				registration.FocusPrevious()
			})

			// Event: Submit form (Enter key)
			ctx.On("submit", func(data interface{}) {
				registration.Submit()
				if registration.Form.IsValid.Get().(bool) {
					submitted.Set(true)
				}
			})

			// Event: Reset form (r key) - Navigation mode only
			ctx.On("reset", func(data interface{}) {
				if inputMode.Get().(bool) {
					return
				}
				registration.Reset()
				submitted.Set(false)
				nameValue.Set("")
				emailValue.Set("")
				passwordValue.Set("")
				confirmValue.Set("")
			})

			// Event: Quit (ctrl+c)
			ctx.On("quit", func(data interface{}) {})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Check for errors
			if errMsg := ctx.Get("error"); errMsg != nil {
				return fmt.Sprintf("Error: %v", errMsg)
			}

			// Get components
			nameField := ctx.Get("nameField").(bubbly.Component)
			emailField := ctx.Get("emailField").(bubbly.Component)
			passwordField := ctx.Get("passwordField").(bubbly.Component)
			confirmField := ctx.Get("confirmField").(bubbly.Component)

			// Get state
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[interface{}]).Get().(bool)
			submitted := ctx.Get("submitted").(*bubbly.Ref[interface{}]).Get().(bool)
			registration := ctx.Get("registration").(*composables.RegistrationComposable)

			// Styling
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				MarginBottom(1)

			// Title
			title := titleStyle.Render("📝 User Registration Form")

			// Mode indicator badge (critical for UX!)
			modeStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)

			var modeIndicator string
			if inputMode {
				// INPUT MODE - Green background
				modeIndicator = modeStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					Render("✍️  INPUT MODE - Tab to navigate fields, Enter to submit, ESC for navigation")
			} else {
				// NAVIGATION MODE - Purple background
				modeIndicator = modeStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("99")).
					Render("🧭 NAVIGATION MODE - Press ESC to enter input mode")
			}

			// Success message
			var successMsg string
			if submitted && registration.Form.IsValid.Get().(bool) {
				successStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("35")).
					Bold(true).
					MarginBottom(1)
				successMsg = successStyle.Render("✓ Registration successful!")
			}

			// Form fields
			formContent := lipgloss.JoinVertical(
				lipgloss.Left,
				nameField.View(),
				"",
				emailField.View(),
				"",
				passwordField.View(),
				"",
				confirmField.View(),
			)

			// Help text (mode-specific)
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)

			comp := ctx.Component()
			var help string
			if inputMode {
				help = helpStyle.Render("tab: next field • shift+tab: prev field • enter: submit • esc: navigation • ctrl+c: quit")
			} else {
				help = helpStyle.Render(comp.HelpText())
			}

			// Compose layout
			content := lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				modeIndicator,
				"",
				successMsg,
				formContent,
				"",
				help,
			)

			return lipgloss.NewStyle().
				Padding(2).
				Render(content)
		})

	return builder.Build()
}
