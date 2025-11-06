package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// TodoFormData represents the form input for creating/editing todos
type TodoFormData struct {
	Title       string
	Description string
	Priority    string
}

// TodoFormProps defines the props for the TodoForm component
type TodoFormProps struct {
	InputMode    *bubbly.Ref[interface{}]
	FocusedField *bubbly.Ref[interface{}]
	OnSubmit     func(data TodoFormData)
	OnCancel     func()
}

// CreateTodoForm creates a reusable todo form component
func CreateTodoForm(props TodoFormProps) (bubbly.Component, error) {
	return bubbly.NewComponent("TodoForm").
		Setup(func(ctx *bubbly.Context) {
			// Use the UseForm composable for form state management
			form := composables.UseForm(ctx, TodoFormData{
				Title:       "",
				Description: "",
				Priority:    "medium",
			}, func(data TodoFormData) map[string]string {
				errors := make(map[string]string)
				if len(data.Title) > 0 && len(data.Title) < 3 {
					errors["Title"] = "Must be at least 3 characters"
				}
				return errors
			})

			// Expose form for external access
			ctx.Expose("form", form)

			// Event: Set form data (for editing)
			ctx.On("setFormData", func(data interface{}) {
				if formData, ok := data.(TodoFormData); ok {
					form.SetField("Title", formData.Title)
					form.SetField("Description", formData.Description)
					form.SetField("Priority", formData.Priority)
				}
			})

			// Event: Clear form
			ctx.On("clearForm", func(_ interface{}) {
				form.Reset()
			})

			// Event: Add character to focused field
			ctx.On("addChar", func(data interface{}) {
				if !props.InputMode.Get().(bool) {
					return
				}
				char := data.(string)
				field := props.FocusedField.Get().(string)

				currentData := form.Values.Get().(TodoFormData)
				switch field {
				case "Title":
					form.SetField("Title", currentData.Title+char)
				case "Description":
					form.SetField("Description", currentData.Description+char)
				case "Priority":
					form.SetField("Priority", currentData.Priority+char)
				}
			})

			// Event: Remove character from focused field
			ctx.On("removeChar", func(_ interface{}) {
				if !props.InputMode.Get().(bool) {
					return
				}
				field := props.FocusedField.Get().(string)

				currentData := form.Values.Get().(TodoFormData)
				switch field {
				case "Title":
					if len(currentData.Title) > 0 {
						form.SetField("Title", currentData.Title[:len(currentData.Title)-1])
					}
				case "Description":
					if len(currentData.Description) > 0 {
						form.SetField("Description", currentData.Description[:len(currentData.Description)-1])
					}
				case "Priority":
					if len(currentData.Priority) > 0 {
						form.SetField("Priority", currentData.Priority[:len(currentData.Priority)-1])
					}
				}
			})

			// Event: Submit form
			ctx.On("submitForm", func(_ interface{}) {
				form.Submit()
				if form.IsValid.GetTyped() {
					data := form.Values.Get().(TodoFormData)
					if props.OnSubmit != nil {
						props.OnSubmit(data)
					}
					form.Reset()
				}
			})

			// Event: Cancel form
			ctx.On("cancelForm", func(_ interface{}) {
				form.Reset()
				if props.OnCancel != nil {
					props.OnCancel()
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			form := ctx.Get("form").(composables.UseFormReturn[TodoFormData])
			data := form.Values.Get().(TodoFormData)
			errors := form.Errors.Get().(map[string]string)
			inputMode := props.InputMode.Get().(bool)
			focusedField := props.FocusedField.Get().(string)

			// Form box - dynamic border color based on mode
			formBorderColor := "240" // Dark grey (navigation mode - inactive)
			if inputMode {
				formBorderColor = "35" // Green (input mode - active)
			}
			formStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(formBorderColor)).
				Width(70)

			// Build form fields
			var formFields []string

			// Title field
			titleLabel := "Title:"
			if focusedField == "Title" {
				titleLabel = "▶ " + titleLabel
			} else {
				titleLabel = "  " + titleLabel
			}
			titleValue := data.Title
			if titleValue == "" {
				titleValue = "(empty)"
			}
			titleError := ""
			if err, ok := errors["Title"]; ok {
				titleError = " ❌ " + err
			}
			formFields = append(formFields, titleLabel+" "+titleValue+titleError)

			// Description field
			descLabel := "Description:"
			if focusedField == "Description" {
				descLabel = "▶ " + descLabel
			} else {
				descLabel = "  " + descLabel
			}
			descValue := data.Description
			if descValue == "" {
				descValue = "(empty)"
			}
			formFields = append(formFields, descLabel+" "+descValue)

			// Priority field
			priorityLabel := "Priority:"
			if focusedField == "Priority" {
				priorityLabel = "▶ " + priorityLabel
			} else {
				priorityLabel = "  " + priorityLabel
			}
			priorityValue := data.Priority
			if priorityValue == "" {
				priorityValue = "(empty)"
			}
			formFields = append(formFields, priorityLabel+" "+priorityValue)

			// Form status
			formStatus := ""
			if form.IsValid.GetTyped() && len(data.Title) >= 3 {
				formStatus = "\n✓ Ready to submit"
			} else if len(data.Title) > 0 && len(data.Title) < 3 {
				formStatus = "\n❌ Title too short"
			}

			return formStyle.Render(strings.Join(formFields, "\n") + formStatus)
		}).
		Build()
}

// RenderFormHelp renders the help text for the form
func RenderFormHelp(inputMode bool) string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	if inputMode {
		return helpStyle.Render("tab: next field • enter: save • esc: cancel • backspace: delete • ctrl+c: quit")
	}
	return helpStyle.Render("↑/↓: select • space: toggle • ctrl+e: edit • ctrl+d: delete • ctrl+n: new • enter: add • ctrl+c: quit")
}

// RenderModeIndicator renders the mode indicator badge
func RenderModeIndicator(inputMode, editMode bool) string {
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginTop(1)

	if inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		if editMode {
			return modeStyle.Render("✏️  EDIT MODE - Type to edit, ESC to cancel")
		}
		return modeStyle.Render("✍️  INPUT MODE - Type to add todo, ESC to navigate")
	}

	modeStyle = modeStyle.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("99"))
	return modeStyle.Render("🧭 NAVIGATION MODE - Use shortcuts, ENTER to add todo")
}
