package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// WizardData represents all form data collected across steps
type WizardData struct {
	// Step 1: Personal Info
	FirstName string
	LastName  string
	Age       string

	// Step 2: Contact Info
	Email string
	Phone string
	City  string

	// Step 3: Preferences
	Theme         string
	Notifications string
	Newsletter    string
}

// model wraps the wizard component
type model struct {
	component    bubbly.Component
	focusedField string
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			// Next step or submit
			m.component.Emit("next", nil)
		case "esc":
			// Previous step
			m.component.Emit("previous", nil)
		case "tab":
			// Next field
			m.component.Emit("nextField", nil)
		case "backspace":
			// Remove character
			m.component.Emit("removeChar", nil)
		default:
			// Handle text input
			switch msg.Type {
			case tea.KeyRunes:
				m.component.Emit("addChar", string(msg.Runes))
			case tea.KeySpace:
				m.component.Emit("addChar", " ")
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

	title := titleStyle.Render("🧙 Form Wizard - Provide/Inject Pattern")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Multi-step form with shared state using provide/inject",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"tab: next field • enter: next step • esc: previous step • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// validateStep1 validates personal info step
func validateStep1(data WizardData) map[string]string {
	errors := make(map[string]string)

	if len(data.FirstName) == 0 {
		errors["FirstName"] = "First name is required"
	} else if len(data.FirstName) < 2 {
		errors["FirstName"] = "Must be at least 2 characters"
	}

	if len(data.LastName) == 0 {
		errors["LastName"] = "Last name is required"
	} else if len(data.LastName) < 2 {
		errors["LastName"] = "Must be at least 2 characters"
	}

	if len(data.Age) == 0 {
		errors["Age"] = "Age is required"
	}

	return errors
}

// validateStep2 validates contact info step
func validateStep2(data WizardData) map[string]string {
	errors := make(map[string]string)

	if len(data.Email) == 0 {
		errors["Email"] = "Email is required"
	} else if !strings.Contains(data.Email, "@") {
		errors["Email"] = "Must be a valid email"
	}

	if len(data.Phone) == 0 {
		errors["Phone"] = "Phone is required"
	} else if len(data.Phone) < 10 {
		errors["Phone"] = "Must be at least 10 digits"
	}

	if len(data.City) == 0 {
		errors["City"] = "City is required"
	}

	return errors
}

// validateStep3 validates preferences step
func validateStep3(data WizardData) map[string]string {
	errors := make(map[string]string)

	validThemes := map[string]bool{"light": true, "dark": true, "auto": true}
	if !validThemes[data.Theme] {
		errors["Theme"] = "Must be: light, dark, or auto"
	}

	validYesNo := map[string]bool{"yes": true, "no": true}
	if !validYesNo[data.Notifications] {
		errors["Notifications"] = "Must be: yes or no"
	}

	if !validYesNo[data.Newsletter] {
		errors["Newsletter"] = "Must be: yes or no"
	}

	return errors
}

// createWizard creates the wizard parent component
func createWizard() (bubbly.Component, error) {
	return bubbly.NewComponent("Wizard").
		Setup(func(ctx *bubbly.Context) {
			// Wizard state
			currentStep := ctx.Ref(1)
			totalSteps := 4 // Personal, Contact, Preferences, Review
			formData := ctx.Ref(WizardData{
				Theme:         "dark",
				Notifications: "yes",
				Newsletter:    "no",
			})
			focusedField := ctx.Ref("FirstName")
			submitted := ctx.Ref(false)

			// Validation errors
			errors := ctx.Ref(make(map[string]string))

			// Provide wizard state to child components
			ctx.Provide("currentStep", currentStep)
			ctx.Provide("totalSteps", totalSteps)
			ctx.Provide("formData", formData)
			ctx.Provide("focusedField", focusedField)
			ctx.Provide("errors", errors)
			ctx.Provide("submitted", submitted)

			// Expose state to template
			ctx.Expose("currentStep", currentStep)
			ctx.Expose("totalSteps", totalSteps)
			ctx.Expose("formData", formData)
			ctx.Expose("focusedField", focusedField)
			ctx.Expose("errors", errors)
			ctx.Expose("submitted", submitted)

			// Event: Next step
			ctx.On("next", func(_ interface{}) {
				step := currentStep.GetTyped().(int)
				data := formData.GetTyped().(WizardData)

				// Validate current step
				var stepErrors map[string]string
				switch step {
				case 1:
					stepErrors = validateStep1(data)
				case 2:
					stepErrors = validateStep2(data)
				case 3:
					stepErrors = validateStep3(data)
				case 4:
					// Review step - submit
					submitted.Set(true)
					return
				}

				if len(stepErrors) == 0 {
					// Valid - move to next step
					if step < totalSteps {
						currentStep.Set(step + 1)
						errors.Set(make(map[string]string))
						// Update focused field for next step
						switch step + 1 {
						case 2:
							focusedField.Set("Email")
						case 3:
							focusedField.Set("Theme")
						}
					}
				} else {
					// Invalid - show errors
					errors.Set(stepErrors)
				}
			})

			// Event: Previous step
			ctx.On("previous", func(_ interface{}) {
				step := currentStep.GetTyped().(int)
				if step > 1 {
					currentStep.Set(step - 1)
					errors.Set(make(map[string]string))
					// Update focused field for previous step
					switch step - 1 {
					case 1:
						focusedField.Set("FirstName")
					case 2:
						focusedField.Set("Email")
					case 3:
						focusedField.Set("Theme")
					}
				}
			})

			// Event: Next field
			ctx.On("nextField", func(_ interface{}) {
				step := currentStep.GetTyped().(int)
				field := focusedField.GetTyped().(string)

				switch step {
				case 1:
					switch field {
					case "FirstName":
						focusedField.Set("LastName")
					case "LastName":
						focusedField.Set("Age")
					case "Age":
						focusedField.Set("FirstName")
					}
				case 2:
					switch field {
					case "Email":
						focusedField.Set("Phone")
					case "Phone":
						focusedField.Set("City")
					case "City":
						focusedField.Set("Email")
					}
				case 3:
					switch field {
					case "Theme":
						focusedField.Set("Notifications")
					case "Notifications":
						focusedField.Set("Newsletter")
					case "Newsletter":
						focusedField.Set("Theme")
					}
				}
			})

			// Event: Add character
			ctx.On("addChar", func(data interface{}) {
				char := data.(string)
				field := focusedField.GetTyped().(string)
				wizardData := formData.GetTyped().(WizardData)

				switch field {
				case "FirstName":
					wizardData.FirstName += char
				case "LastName":
					wizardData.LastName += char
				case "Age":
					wizardData.Age += char
				case "Email":
					wizardData.Email += char
				case "Phone":
					wizardData.Phone += char
				case "City":
					wizardData.City += char
				case "Theme":
					wizardData.Theme += char
				case "Notifications":
					wizardData.Notifications += char
				case "Newsletter":
					wizardData.Newsletter += char
				}

				formData.Set(wizardData)
			})

			// Event: Remove character
			ctx.On("removeChar", func(_ interface{}) {
				field := focusedField.GetTyped().(string)
				wizardData := formData.GetTyped().(WizardData)

				switch field {
				case "FirstName":
					if len(wizardData.FirstName) > 0 {
						wizardData.FirstName = wizardData.FirstName[:len(wizardData.FirstName)-1]
					}
				case "LastName":
					if len(wizardData.LastName) > 0 {
						wizardData.LastName = wizardData.LastName[:len(wizardData.LastName)-1]
					}
				case "Age":
					if len(wizardData.Age) > 0 {
						wizardData.Age = wizardData.Age[:len(wizardData.Age)-1]
					}
				case "Email":
					if len(wizardData.Email) > 0 {
						wizardData.Email = wizardData.Email[:len(wizardData.Email)-1]
					}
				case "Phone":
					if len(wizardData.Phone) > 0 {
						wizardData.Phone = wizardData.Phone[:len(wizardData.Phone)-1]
					}
				case "City":
					if len(wizardData.City) > 0 {
						wizardData.City = wizardData.City[:len(wizardData.City)-1]
					}
				case "Theme":
					if len(wizardData.Theme) > 0 {
						wizardData.Theme = wizardData.Theme[:len(wizardData.Theme)-1]
					}
				case "Notifications":
					if len(wizardData.Notifications) > 0 {
						wizardData.Notifications = wizardData.Notifications[:len(wizardData.Notifications)-1]
					}
				case "Newsletter":
					if len(wizardData.Newsletter) > 0 {
						wizardData.Newsletter = wizardData.Newsletter[:len(wizardData.Newsletter)-1]
					}
				}

				formData.Set(wizardData)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			currentStep := ctx.Get("currentStep").(*bubbly.Ref[interface{}])
			totalSteps := ctx.Get("totalSteps").(int)
			formData := ctx.Get("formData").(*bubbly.Ref[interface{}])
			focusedField := ctx.Get("focusedField").(*bubbly.Ref[interface{}])
			errors := ctx.Get("errors").(*bubbly.Ref[interface{}])
			submitted := ctx.Get("submitted").(*bubbly.Ref[interface{}])

			step := currentStep.GetTyped().(int)
			data := formData.GetTyped().(WizardData)
			focused := focusedField.GetTyped().(string)
			errorMap := errors.GetTyped().(map[string]string)
			isSubmitted := submitted.GetTyped().(bool)

			// Progress bar
			progressStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(70)

			progress := ""
			for i := 1; i <= totalSteps; i++ {
				if i < step {
					progress += "✅ "
				} else if i == step {
					progress += "▶️  "
				} else {
					progress += "⭕ "
				}
			}
			progress += fmt.Sprintf(" Step %d/%d", step, totalSteps)
			progressBox := progressStyle.Render(progress)

			// If submitted, show success message
			if isSubmitted {
				successStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					Padding(2, 4).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("99")).
					Width(70).
					Align(lipgloss.Center)

				successMsg := successStyle.Render(
					"✅ Form Submitted Successfully!\n\n" +
						fmt.Sprintf("Welcome, %s %s!", data.FirstName, data.LastName),
				)

				return lipgloss.JoinVertical(
					lipgloss.Left,
					progressBox,
					"",
					successMsg,
				)
			}

			// Form style
			formStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(70)

			var formContent string

			// Render current step
			switch step {
			case 1:
				// Personal Info
				formContent = "📝 Personal Information\n\n"
				formContent += renderField("FirstName", data.FirstName, focused, errorMap)
				formContent += renderField("LastName", data.LastName, focused, errorMap)
				formContent += renderField("Age", data.Age, focused, errorMap)

			case 2:
				// Contact Info
				formContent = "📧 Contact Information\n\n"
				formContent += renderField("Email", data.Email, focused, errorMap)
				formContent += renderField("Phone", data.Phone, focused, errorMap)
				formContent += renderField("City", data.City, focused, errorMap)

			case 3:
				// Preferences
				formContent = "⚙️  Preferences\n\n"
				formContent += renderField("Theme", data.Theme, focused, errorMap)
				formContent += "  (light, dark, auto)\n"
				formContent += renderField("Notifications", data.Notifications, focused, errorMap)
				formContent += "  (yes, no)\n"
				formContent += renderField("Newsletter", data.Newsletter, focused, errorMap)
				formContent += "  (yes, no)\n"

			case 4:
				// Review
				formContent = "👀 Review Your Information\n\n"
				formContent += "Personal Info:\n"
				formContent += fmt.Sprintf("  Name: %s %s\n", data.FirstName, data.LastName)
				formContent += fmt.Sprintf("  Age:  %s\n\n", data.Age)
				formContent += "Contact Info:\n"
				formContent += fmt.Sprintf("  Email: %s\n", data.Email)
				formContent += fmt.Sprintf("  Phone: %s\n", data.Phone)
				formContent += fmt.Sprintf("  City:  %s\n\n", data.City)
				formContent += "Preferences:\n"
				formContent += fmt.Sprintf("  Theme:         %s\n", data.Theme)
				formContent += fmt.Sprintf("  Notifications: %s\n", data.Notifications)
				formContent += fmt.Sprintf("  Newsletter:    %s\n\n", data.Newsletter)
				formContent += "Press Enter to submit or Esc to go back"
			}

			// Show validation errors
			if len(errorMap) > 0 && step < 4 {
				formContent += fmt.Sprintf("\n\n❌ Please fix %d error%s",
					len(errorMap), map[bool]string{true: "", false: "s"}[len(errorMap) == 1])
			}

			formBox := formStyle.Render(formContent)

			// Info box
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(70)

			infoBox := infoStyle.Render(
				"Provide/Inject Pattern:\n\n" +
					"• Parent wizard provides shared state\n" +
					"• Each step injects wizard context\n" +
					"• State persists across navigation\n" +
					"• Validation per step\n" +
					"• Final review before submit",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				progressBox,
				"",
				formBox,
				"",
				infoBox,
			)
		}).
		Build()
}

// renderField renders a form field with focus indicator and error
func renderField(name, value, focused string, errors map[string]string) string {
	indicator := "  "
	if name == focused {
		indicator = "▶ "
	}

	displayValue := value
	if displayValue == "" {
		displayValue = "(empty)"
	}

	errorMsg := ""
	if err, ok := errors[name]; ok {
		errorMsg = " ❌ " + err
	}

	return fmt.Sprintf("%s%s: %s%s\n", indicator, name, displayValue, errorMsg)
}

func main() {
	component, err := createWizard()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{component: component, focusedField: "FirstName"}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
