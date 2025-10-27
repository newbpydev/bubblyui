package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// model wraps the calculator component
type model struct {
	calculator bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.calculator.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			// Record breadcrumb for user input
			observability.RecordBreadcrumb("user", fmt.Sprintf("User pressed key: %s", msg.String()), map[string]interface{}{
				"key": msg.String(),
			})
			m.calculator.Emit("digit", msg.String())
		case "+":
			observability.RecordBreadcrumb("user", "User pressed operator: +", map[string]interface{}{
				"operator": "+",
			})
			m.calculator.Emit("operator", "+")
		case "-":
			observability.RecordBreadcrumb("user", "User pressed operator: -", map[string]interface{}{
				"operator": "-",
			})
			m.calculator.Emit("operator", "-")
		case "*":
			observability.RecordBreadcrumb("user", "User pressed operator: *", map[string]interface{}{
				"operator": "*",
			})
			m.calculator.Emit("operator", "*")
		case "/":
			observability.RecordBreadcrumb("user", "User pressed operator: /", map[string]interface{}{
				"operator": "/",
			})
			m.calculator.Emit("operator", "/")
		case "enter", "=":
			observability.RecordBreadcrumb("user", "User pressed equals", nil)
			m.calculator.Emit("equals", nil)
		case "c":
			observability.RecordBreadcrumb("user", "User cleared calculator", nil)
			m.calculator.Emit("clear", nil)
		case "p":
			// Trigger a panic to demonstrate error tracking
			observability.RecordBreadcrumb("user", "User triggered panic (for testing)", map[string]interface{}{
				"action": "panic_test",
			})
			m.calculator.Emit("panic", nil)
		}
	}

	updatedComponent, cmd := m.calculator.Update(msg)
	m.calculator = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔍 Error Tracking - Console Reporter (Development)")

	componentView := m.calculator.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"0-9: digits • +/-/*/: operators • enter/=: calculate • c: clear • p: panic (test) • q: quit",
	)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		Italic(true)

	info := infoStyle.Render(
		"💡 Breadcrumbs are being recorded. Try triggering a panic with 'p' to see error tracking in action!",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n%s\n", title, componentView, help, info)
}

// createCalculator creates a calculator component with error tracking
func createCalculator() (bubbly.Component, error) {
	return bubbly.NewComponent("Calculator").
		Setup(func(ctx *bubbly.Context) {
			// Reactive state
			display := ctx.Ref("0")
			currentValue := ctx.Ref(0.0)
			previousValue := ctx.Ref(0.0)
			operator := ctx.Ref("")
			newNumber := ctx.Ref(true)

			// Expose state
			ctx.Expose("display", display)
			ctx.Expose("currentValue", currentValue)
			ctx.Expose("previousValue", previousValue)
			ctx.Expose("operator", operator)
			ctx.Expose("newNumber", newNumber)

			// Record breadcrumb on component init
			observability.RecordBreadcrumb("component", "Calculator component initialized", map[string]interface{}{
				"component": "Calculator",
			})

			// Event handlers
			ctx.On("digit", func(data interface{}) {
				digit := data.(string)
				currentDisplay := display.Get().(string)
				isNew := newNumber.Get().(bool)

				if isNew || currentDisplay == "0" {
					display.Set(digit)
					newNumber.Set(false)
				} else {
					display.Set(currentDisplay + digit)
				}

				observability.RecordBreadcrumb("state", "Display updated", map[string]interface{}{
					"display": display.Get().(string),
				})
			})

			ctx.On("operator", func(data interface{}) {
				op := data.(string)
				currentOp := operator.Get().(string)

				// If there's a pending operation, calculate it first
				if currentOp != "" {
					ctx.Emit("equals", nil)
				}

				// Parse current display value
				var val float64
				fmt.Sscanf(display.Get().(string), "%f", &val)
				previousValue.Set(val)
				operator.Set(op)
				newNumber.Set(true)

				observability.RecordBreadcrumb("state", "Operator set", map[string]interface{}{
					"operator":      op,
					"previousValue": val,
				})
			})

			ctx.On("equals", func(data interface{}) {
				op := operator.Get().(string)
				if op == "" {
					return
				}

				var current, previous float64
				fmt.Sscanf(display.Get().(string), "%f", &current)
				previous = previousValue.Get().(float64)

				var result float64
				switch op {
				case "+":
					result = previous + current
				case "-":
					result = previous - current
				case "*":
					result = previous * current
				case "/":
					if current == 0 {
						// This will cause a display error but won't panic
						display.Set("Error: Div by 0")
						operator.Set("")
						newNumber.Set(true)

						observability.RecordBreadcrumb("error", "Division by zero attempted", map[string]interface{}{
							"operator": op,
							"dividend": previous,
							"divisor":  current,
						})
						return
					}
					result = previous / current
				}

				display.Set(fmt.Sprintf("%.2f", result))
				currentValue.Set(result)
				operator.Set("")
				newNumber.Set(true)

				observability.RecordBreadcrumb("state", "Calculation completed", map[string]interface{}{
					"operation": fmt.Sprintf("%.2f %s %.2f = %.2f", previous, op, current, result),
					"result":    result,
				})
			})

			ctx.On("clear", func(data interface{}) {
				display.Set("0")
				currentValue.Set(0.0)
				previousValue.Set(0.0)
				operator.Set("")
				newNumber.Set(true)

				observability.RecordBreadcrumb("state", "Calculator cleared", nil)
			})

			ctx.On("panic", func(data interface{}) {
				// Intentionally panic to demonstrate error tracking
				observability.RecordBreadcrumb("debug", "About to trigger panic", map[string]interface{}{
					"intentional": true,
				})
				panic("Intentional panic for error tracking demonstration!")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			display := ctx.Get("display").(*bubbly.Ref[interface{}])
			operator := ctx.Get("operator").(*bubbly.Ref[interface{}])

			displayVal := display.Get().(string)
			operatorVal := operator.Get().(string)

			// Display box
			displayStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(40).
				Align(lipgloss.Right)

			displayBox := displayStyle.Render(displayVal)

			// Operator indicator
			opStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(0, 2).
				Width(40)

			opText := "Ready"
			if operatorVal != "" {
				opText = fmt.Sprintf("Operator: %s", operatorVal)
			}
			opBox := opStyle.Render(opText)

			// Breadcrumbs display
			breadcrumbs := observability.GetBreadcrumbs()
			breadcrumbStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(40).
				Height(8)

			breadcrumbText := "Recent Breadcrumbs:\n"
			// Show last 5 breadcrumbs
			start := len(breadcrumbs) - 5
			if start < 0 {
				start = 0
			}
			for i := start; i < len(breadcrumbs); i++ {
				bc := breadcrumbs[i]
				breadcrumbText += fmt.Sprintf("• [%s] %s\n", bc.Category, bc.Message)
			}

			breadcrumbBox := breadcrumbStyle.Render(breadcrumbText)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				displayBox,
				opBox,
				"",
				breadcrumbBox,
			)
		}).
		Build()
}

func main() {
	// Setup console reporter for development
	reporter := observability.NewConsoleReporter(true) // verbose mode
	observability.SetErrorReporter(reporter)

	// Record initial breadcrumb
	observability.RecordBreadcrumb("navigation", "Application started", map[string]interface{}{
		"example": "console-reporter",
		"mode":    "development",
	})

	calculator, err := createCalculator()
	if err != nil {
		fmt.Printf("Error creating calculator: %v\n", err)
		os.Exit(1)
	}

	calculator.Init()

	m := model{calculator: calculator}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Flush reporter before exit
	reporter.Flush(0)

	observability.RecordBreadcrumb("navigation", "Application exited", nil)
}
