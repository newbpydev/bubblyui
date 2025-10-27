package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// errorMsg is sent when an error is reported
type errorMsg struct {
	message string
}

// TUIReporter is a custom reporter that sends errors to the TUI
// It stores pending errors to be retrieved by the Update loop
type TUIReporter struct {
	mu            sync.Mutex
	pendingErrors []string
}

func (r *TUIReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	DebugLog("PANIC", "ReportPanic called: component=%s, event=%s, panic=%v", ctx.ComponentName, ctx.EventName, err.PanicValue)
	msg := fmt.Sprintf("Panic in '%s.%s': %v", ctx.ComponentName, ctx.EventName, err.PanicValue)
	
	r.mu.Lock()
	r.pendingErrors = append(r.pendingErrors, msg)
	r.mu.Unlock()
	
	DebugLog("PANIC", "Error stored in pending queue")
}

func (r *TUIReporter) ReportError(err error, ctx *observability.ErrorContext) {
	DebugLog("ERROR", "ReportError called: component=%s, error=%v", ctx.ComponentName, err)
	msg := fmt.Sprintf("Error in '%s': %v", ctx.ComponentName, err)
	
	r.mu.Lock()
	r.pendingErrors = append(r.pendingErrors, msg)
	r.mu.Unlock()
	
	DebugLog("ERROR", "Error stored in pending queue")
}

// GetPendingErrors retrieves and clears all pending errors
func (r *TUIReporter) GetPendingErrors() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if len(r.pendingErrors) == 0 {
		return nil
	}
	
	errors := make([]string, len(r.pendingErrors))
	copy(errors, r.pendingErrors)
	r.pendingErrors = nil
	
	return errors
}

func (r *TUIReporter) Flush(timeout time.Duration) error {
	return nil
}

// model wraps the calculator component
type model struct {
	calculator   bubbly.Component
	lastError    string
	errorVisible bool
	reporter     *TUIReporter
}

func (m model) Init() tea.Cmd {
	return m.calculator.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	DebugLog("UPDATE", "Update called with msg type: %T", msg)
	
	switch msg := msg.(type) {
	case errorMsg:
		DebugLog("UPDATE", "Received errorMsg: %s", msg.message)
		m.lastError = msg.message
		m.errorVisible = true
		return m, nil
	case tea.KeyMsg:
		DebugLog("UPDATE", "Received KeyMsg: %s", msg.String())
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Clear error message
			m.errorVisible = false
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			// Record breadcrumb for user input
			observability.RecordBreadcrumb("user", fmt.Sprintf("User pressed key: %s", msg.String()), map[string]interface{}{
				"key": msg.String(),
			})
			m.calculator.Emit("digit", msg.String())
			return m, nil
		case "+":
			observability.RecordBreadcrumb("user", "User pressed operator: +", map[string]interface{}{
				"operator": "+",
			})
			m.calculator.Emit("operator", "+")
			return m, nil
		case "-":
			observability.RecordBreadcrumb("user", "User pressed operator: -", map[string]interface{}{
				"operator": "-",
			})
			m.calculator.Emit("operator", "-")
			return m, nil
		case "*":
			observability.RecordBreadcrumb("user", "User pressed operator: *", map[string]interface{}{
				"operator": "*",
			})
			m.calculator.Emit("operator", "*")
			return m, nil
		case "/":
			observability.RecordBreadcrumb("user", "User pressed operator: /", map[string]interface{}{
				"operator": "/",
			})
			m.calculator.Emit("operator", "/")
			return m, nil
		case "enter", "=":
			observability.RecordBreadcrumb("user", "User pressed equals", nil)
			m.calculator.Emit("equals", nil)
			return m, nil
		case "c":
			observability.RecordBreadcrumb("user", "User cleared calculator", nil)
			m.calculator.Emit("clear", nil)
			return m, nil
		case "p":
			// Trigger a panic to demonstrate error tracking
			DebugLog("KEY", "User pressed 'p' - about to trigger panic")
			observability.RecordBreadcrumb("user", "User triggered panic (for testing)", map[string]interface{}{
				"action": "panic_test",
			})
			DebugLog("KEY", "About to call Emit('panic', nil)")
			m.calculator.Emit("panic", nil)
			DebugLog("KEY", "Emit('panic', nil) returned successfully")
			
			// Check for pending errors after Emit
			if errors := m.reporter.GetPendingErrors(); len(errors) > 0 {
				DebugLog("KEY", "Found %d pending errors", len(errors))
				m.lastError = errors[0]
				m.errorVisible = true
			}
			return m, nil
		}
	}

	DebugLog("UPDATE", "Update returning with no action")
	
	// Always check for pending errors before returning
	if errors := m.reporter.GetPendingErrors(); len(errors) > 0 {
		DebugLog("UPDATE", "Found %d pending errors at end of Update", len(errors))
		m.lastError = errors[0]
		m.errorVisible = true
	}
	
	return m, nil
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
		"0-9: digits • +/-/*//: operators • enter/=: calculate • c: clear • p: panic (test) • esc: dismiss error • q: quit",
	)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		Italic(true)

	info := infoStyle.Render(
		"💡 Breadcrumbs are being recorded. Try triggering a panic with 'p' to see error tracking in action!",
	)

	result := fmt.Sprintf("%s\n\n%s\n%s\n%s\n", title, componentView, help, info)

	// Show error overlay if there's an error
	if m.errorVisible {
		errorStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("196")).
			Padding(1, 2).
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("196")).
			Width(60)

		errorBox := errorStyle.Render(fmt.Sprintf("🚨 ERROR CAUGHT!\n\n%s\n\nPress ESC to dismiss", m.lastError))
		result += "\n" + errorBox
	}

	return result
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
				// Extract data from Event struct
				event := data.(*bubbly.Event)
				digit := event.Data.(string)
				
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
				event := data.(*bubbly.Event)
				op := event.Data.(string)
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
				DebugLog("HANDLER", "panic event handler called")
				observability.RecordBreadcrumb("debug", "About to trigger panic", map[string]interface{}{
					"intentional": true,
				})
				DebugLog("HANDLER", "About to panic!")
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
	// Initialize debug logger
	if err := InitDebugLogger("debug.log"); err != nil {
		fmt.Printf("Error creating debug logger: %v\n", err)
		os.Exit(1)
	}
	defer CloseDebugLogger()

	// Catch any panics in main
	defer func() {
		if r := recover(); r != nil {
			DebugLogWithStack("FATAL", "Panic in main: %v", r)
			fmt.Printf("FATAL PANIC: %v\n", r)
			os.Exit(1)
		}
	}()

	DebugLog("MAIN", "Application starting")

	// Record initial breadcrumb
	observability.RecordBreadcrumb("navigation", "Application started", map[string]interface{}{
		"example": "console-reporter",
		"mode":    "development",
	})

	DebugLog("MAIN", "Creating calculator component")
	calculator, err := createCalculator()
	if err != nil {
		DebugLog("ERROR", "Error creating calculator: %v", err)
		fmt.Printf("Error creating calculator: %v\n", err)
		os.Exit(1)
	}

	DebugLog("MAIN", "Initializing calculator")
	calculator.Init()

	// Setup TUI reporter
	DebugLog("MAIN", "Setting up TUI reporter")
	reporter := &TUIReporter{}
	observability.SetErrorReporter(reporter)

	m := model{
		calculator: calculator,
		reporter:   reporter,
	}

	DebugLog("MAIN", "Creating Bubbletea program")
	p := tea.NewProgram(m, tea.WithAltScreen())

	DebugLog("MAIN", "Starting Bubbletea program")
	if _, err := p.Run(); err != nil {
		DebugLog("ERROR", "Error running program: %v", err)
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	DebugLog("MAIN", "Application exiting normally")
	observability.RecordBreadcrumb("navigation", "Application exited", nil)
}
