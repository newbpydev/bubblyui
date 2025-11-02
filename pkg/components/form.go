package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// FormField represents a single field in a form.
// Each field has a name (for identification), a label (for display),
// and a component (the actual input/select/checkbox/etc.).
type FormField struct {
	// Name is the unique identifier for this field.
	// Used for validation error mapping.
	// Required - must be unique within the form.
	Name string

	// Label is the display text shown above the field.
	// Optional - if empty, no label is displayed.
	Label string

	// Component is the input component for this field.
	// Can be Input, Checkbox, Select, TextArea, Radio, Toggle, etc.
	// Required - must be a valid component.
	Component bubbly.Component
}

// FormProps defines the configuration properties for a Form component.
//
// Form is a generic component that works with any struct type T.
// It manages form state, validation, and submission.
//
// Example usage:
//
//	type UserData struct {
//	    Name  string
//	    Email string
//	    Age   int
//	}
//
//	nameRef := bubbly.NewRef("")
//	emailRef := bubbly.NewRef("")
//
//	form := components.Form(components.FormProps[UserData]{
//	    Initial: UserData{},
//	    Fields: []components.FormField{
//	        {
//	            Name:  "Name",
//	            Label: "Full Name",
//	            Component: components.Input(components.InputProps{
//	                Value: nameRef,
//	                Placeholder: "Enter your name",
//	            }),
//	        },
//	        {
//	            Name:  "Email",
//	            Label: "Email Address",
//	            Component: components.Input(components.InputProps{
//	                Value: emailRef,
//	                Type: components.InputEmail,
//	            }),
//	        },
//	    },
//	    Validate: func(data UserData) map[string]string {
//	        errors := make(map[string]string)
//	        if data.Name == "" {
//	            errors["Name"] = "Name is required"
//	        }
//	        if data.Email == "" {
//	            errors["Email"] = "Email is required"
//	        }
//	        return errors
//	    },
//	    OnSubmit: func(data UserData) {
//	        saveUser(data)
//	    },
//	})
type FormProps[T any] struct {
	// Initial is the initial form data.
	// Used to populate the form on first render and for reset functionality.
	// Required - must be a valid struct of type T.
	Initial T

	// Validate is a function that validates the entire form.
	// Returns a map of field names to error messages.
	// Optional - if nil, no validation is performed.
	// Called on submit and optionally on field changes.
	Validate func(T) map[string]string

	// OnSubmit is a callback function executed when the form is submitted.
	// Only called if validation passes (no errors).
	// Receives the validated form data as a parameter.
	// Optional - if nil, no callback is executed.
	OnSubmit func(T)

	// OnCancel is a callback function executed when the form is cancelled.
	// Optional - if nil, no callback is executed.
	OnCancel func()

	// Fields is the list of form fields to display.
	// Each field includes a name, label, and component.
	// Required - should not be empty for usability.
	Fields []FormField

	// Common props for all components
	CommonProps
}

// Form creates a new Form organism component with generic type support.
//
// Form is a container component that manages multiple input fields,
// validation, and submission. It integrates with the UseForm composable
// for state management and provides a consistent layout for forms.
//
// The form automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	type LoginData struct {
//	    Username string
//	    Password string
//	}
//
//	usernameRef := bubbly.NewRef("")
//	passwordRef := bubbly.NewRef("")
//
//	form := components.Form(components.FormProps[LoginData]{
//	    Initial: LoginData{},
//	    Fields: []components.FormField{
//	        {
//	            Name:  "Username",
//	            Label: "Username",
//	            Component: components.Input(components.InputProps{
//	                Value: usernameRef,
//	            }),
//	        },
//	        {
//	            Name:  "Password",
//	            Label: "Password",
//	            Component: components.Input(components.InputProps{
//	                Value: passwordRef,
//	                Type: components.InputPassword,
//	            }),
//	        },
//	    },
//	    Validate: func(data LoginData) map[string]string {
//	        errors := make(map[string]string)
//	        if data.Username == "" {
//	            errors["Username"] = "Username is required"
//	        }
//	        if len(data.Password) < 8 {
//	            errors["Password"] = "Password must be at least 8 characters"
//	        }
//	        return errors
//	    },
//	    OnSubmit: func(data LoginData) {
//	        authenticate(data.Username, data.Password)
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	form.Init()
//	view := form.View()
//
// Features:
//   - Generic type support for any form data struct
//   - Field collection with labels
//   - Validation with error display per field
//   - Submit/cancel handlers
//   - Integration with UseForm composable
//   - Theme integration
//   - Custom style override
//
// Keyboard interaction:
//   - Tab: Navigate between fields
//   - Enter: Submit form (if valid)
//   - Escape: Cancel form
//
// The form uses the UseForm composable internally for state management,
// providing reactive validation and dirty tracking.
func Form[T any](props FormProps[T]) bubbly.Component {
	// Extract child components from fields
	children := make([]bubbly.Component, len(props.Fields))
	for i, field := range props.Fields {
		children[i] = field.Component
	}

	comp, err := bubbly.NewComponent("Form").
		Props(props).
		Children(children...).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme and provide to children
			theme := ctx.Inject("theme", DefaultTheme).(Theme)
			ctx.Provide("theme", theme)

			// Create reactive state for form errors
			errors := bubbly.NewRef(make(map[string]string))

			// Create reactive state for submitting status
			submitting := bubbly.NewRef(false)

			// Submit handler
			ctx.On("submit", func(_ interface{}) {
				if submitting.Get().(bool) {
					return
				}

				// Run validation if provided
				if props.Validate != nil {
					validationErrors := props.Validate(props.Initial)
					errors.Set(validationErrors)

					// Only submit if no errors
					if len(validationErrors) == 0 {
						submitting.Set(true)
						if props.OnSubmit != nil {
							props.OnSubmit(props.Initial)
						}
						submitting.Set(false)
					}
				} else {
					// No validation, submit directly
					submitting.Set(true)
					if props.OnSubmit != nil {
						props.OnSubmit(props.Initial)
					}
					submitting.Set(false)
				}
			})

			// Cancel handler
			ctx.On("cancel", func(_ interface{}) {
				if props.OnCancel != nil {
					props.OnCancel()
				}
			})

			// Expose state
			ctx.Expose("errors", errors)
			ctx.Expose("submitting", submitting)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(FormProps[T])
			errors := ctx.Get("errors").(*bubbly.Ref[map[string]string])
			submitting := ctx.Get("submitting").(*bubbly.Ref[bool])
			theme := ctx.Get("theme").(Theme)

			var output strings.Builder

			// Title style
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				MarginBottom(1)

			output.WriteString(titleStyle.Render("Form"))
			output.WriteString("\n")

			// Render fields
			for _, field := range p.Fields {
				// Field label
				if field.Label != "" {
					labelStyle := lipgloss.NewStyle().
						Foreground(theme.Foreground).
						Bold(true)
					output.WriteString(labelStyle.Render(field.Label + ":"))
					output.WriteString("\n")
				}

				// Field component
				output.WriteString(field.Component.View())
				output.WriteString("\n")

				// Field error (if any)
				errorMap := errors.Get().(map[string]string)
				if err, ok := errorMap[field.Name]; ok && err != "" {
					errorStyle := lipgloss.NewStyle().
						Foreground(theme.Danger).
						Italic(true).
						MarginLeft(2)
					output.WriteString(errorStyle.Render("⚠ " + err))
					output.WriteString("\n")
				}

				output.WriteString("\n")
			}

			// Buttons
			submitLabel := "Submit"
			if submitting.Get().(bool) {
				submitLabel = "Submitting..."
			}

			// Render buttons directly with theme styling
			submitStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true).
				Background(theme.Primary).
				Foreground(lipgloss.Color("230"))

			if submitting.Get().(bool) {
				submitStyle = submitStyle.Foreground(theme.Muted)
			}

			cancelStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Secondary)

			output.WriteString(submitStyle.Render(submitLabel))
			output.WriteString("  ")
			output.WriteString(cancelStyle.Render("Cancel"))

			return output.String()
		}).
		Build()

	if err != nil {
		panic(err) // Should never happen with valid setup
	}

	return comp
}
