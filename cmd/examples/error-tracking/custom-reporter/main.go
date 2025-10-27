package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// FileReporter is a custom error reporter that writes errors to a JSON file
// with privacy filtering to remove sensitive information
type FileReporter struct {
	filePath string
	mu       sync.Mutex
	errors   []ErrorReport
	// Privacy filters
	emailRegex    *regexp.Regexp
	phoneRegex    *regexp.Regexp
	ssnRegex      *regexp.Regexp
	creditCardRegex *regexp.Regexp
}

// ErrorReport represents a sanitized error report
type ErrorReport struct {
	Type          string                 `json:"type"`
	Message       string                 `json:"message"`
	ComponentName string                 `json:"component_name"`
	ComponentID   string                 `json:"component_id"`
	EventName     string                 `json:"event_name"`
	Timestamp     time.Time              `json:"timestamp"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
	Breadcrumbs   []BreadcrumbReport     `json:"breadcrumbs,omitempty"`
	StackTrace    string                 `json:"stack_trace,omitempty"`
}

// BreadcrumbReport represents a sanitized breadcrumb
type BreadcrumbReport struct {
	Type      string                 `json:"type"`
	Category  string                 `json:"category"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewFileReporter creates a new file-based error reporter with privacy filtering
func NewFileReporter(filePath string) *FileReporter {
	return &FileReporter{
		filePath:        filePath,
		errors:          make([]ErrorReport, 0),
		emailRegex:      regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		phoneRegex:      regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		ssnRegex:        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		creditCardRegex: regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
	}
}

// sanitizeString removes sensitive information from a string
func (r *FileReporter) sanitizeString(s string) string {
	// Replace email addresses
	s = r.emailRegex.ReplaceAllString(s, "[EMAIL_REDACTED]")
	// Replace phone numbers
	s = r.phoneRegex.ReplaceAllString(s, "[PHONE_REDACTED]")
	// Replace SSNs
	s = r.ssnRegex.ReplaceAllString(s, "[SSN_REDACTED]")
	// Replace credit card numbers
	s = r.creditCardRegex.ReplaceAllString(s, "[CC_REDACTED]")
	return s
}

// sanitizeData recursively sanitizes data structures
func (r *FileReporter) sanitizeData(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		return r.sanitizeString(v)
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			// Skip sensitive keys entirely
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "password") ||
				strings.Contains(lowerKey, "secret") ||
				strings.Contains(lowerKey, "token") ||
				strings.Contains(lowerKey, "api_key") {
				sanitized[key] = "[REDACTED]"
				continue
			}
			sanitized[key] = r.sanitizeData(value)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = r.sanitizeData(item)
		}
		return sanitized
	default:
		return v
	}
}

// sanitizeBreadcrumbs sanitizes breadcrumb data
func (r *FileReporter) sanitizeBreadcrumbs(breadcrumbs []observability.Breadcrumb) []BreadcrumbReport {
	sanitized := make([]BreadcrumbReport, len(breadcrumbs))
	for i, bc := range breadcrumbs {
		sanitized[i] = BreadcrumbReport{
			Type:      bc.Type,
			Category:  bc.Category,
			Message:   r.sanitizeString(bc.Message),
			Level:     bc.Level,
			Timestamp: bc.Timestamp,
			Data:      r.sanitizeData(bc.Data).(map[string]interface{}),
		}
	}
	return sanitized
}

// ReportPanic reports a panic with privacy filtering
func (r *FileReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	r.mu.Lock()
	defer r.mu.Unlock()

	report := ErrorReport{
		Type:          "panic",
		Message:       r.sanitizeString(fmt.Sprintf("%v", err.PanicValue)),
		ComponentName: ctx.ComponentName,
		ComponentID:   ctx.ComponentID,
		EventName:     ctx.EventName,
		Timestamp:     ctx.Timestamp,
		Tags:          ctx.Tags,
		Extra:         r.sanitizeData(ctx.Extra).(map[string]interface{}),
		Breadcrumbs:   r.sanitizeBreadcrumbs(ctx.Breadcrumbs),
		StackTrace:    r.sanitizeString(string(ctx.StackTrace)),
	}

	r.errors = append(r.errors, report)
	fmt.Printf("[FileReporter] Panic reported: %s in %s.%s\n", report.Message, ctx.ComponentName, ctx.EventName)
}

// ReportError reports an error with privacy filtering
func (r *FileReporter) ReportError(err error, ctx *observability.ErrorContext) {
	r.mu.Lock()
	defer r.mu.Unlock()

	report := ErrorReport{
		Type:          "error",
		Message:       r.sanitizeString(err.Error()),
		ComponentName: ctx.ComponentName,
		ComponentID:   ctx.ComponentID,
		EventName:     ctx.EventName,
		Timestamp:     ctx.Timestamp,
		Tags:          ctx.Tags,
		Extra:         r.sanitizeData(ctx.Extra).(map[string]interface{}),
		Breadcrumbs:   r.sanitizeBreadcrumbs(ctx.Breadcrumbs),
		StackTrace:    r.sanitizeString(string(ctx.StackTrace)),
	}

	r.errors = append(r.errors, report)
	fmt.Printf("[FileReporter] Error reported: %s in %s.%s\n", report.Message, ctx.ComponentName, ctx.EventName)
}

// Flush writes all errors to the JSON file
func (r *FileReporter) Flush(timeout time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.errors) == 0 {
		return nil
	}

	// Marshal errors to JSON
	data, err := json.MarshalIndent(r.errors, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal errors: %w", err)
	}

	// Write to file
	err = os.WriteFile(r.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write errors to file: %w", err)
	}

	fmt.Printf("[FileReporter] Flushed %d errors to %s\n", len(r.errors), r.filePath)
	return nil
}

// model wraps the payment form component
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
			m.form.Emit("next-field", nil)
		case "shift+tab":
			m.form.Emit("prev-field", nil)
		case "enter":
			m.form.Emit("submit", nil)
		case "backspace":
			m.form.Emit("backspace", nil)
		case "e":
			// Trigger an error with sensitive data
			m.form.Emit("trigger-error", nil)
		default:
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

	title := titleStyle.Render("🔒 Error Tracking - Custom Reporter with Privacy Filtering")

	componentView := m.form.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"tab/shift+tab: switch fields • enter: submit • e: error test • q: quit",
	)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		Italic(true)

	info := infoStyle.Render(
		"💡 Custom reporter filters PII (emails, phones, SSNs, credit cards, passwords). Check errors.json!",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n%s\n", title, componentView, help, info)
}

// createPaymentForm creates a payment form with sensitive data
func createPaymentForm() (bubbly.Component, error) {
	return bubbly.NewComponent("PaymentForm").
		Setup(func(ctx *bubbly.Context) {
			// Reactive state (contains sensitive data)
			cardNumber := ctx.Ref("")
			email := ctx.Ref("")
			phone := ctx.Ref("")
			currentField := ctx.Ref(0) // 0: card, 1: email, 2: phone
			statusMessage := ctx.Ref("")

			// Expose state
			ctx.Expose("cardNumber", cardNumber)
			ctx.Expose("email", email)
			ctx.Expose("phone", phone)
			ctx.Expose("currentField", currentField)
			ctx.Expose("statusMessage", statusMessage)

			observability.RecordBreadcrumb("component", "PaymentForm initialized", map[string]interface{}{
				"component": "PaymentForm",
			})

			ctx.On("input", func(data interface{}) {
				event := data.(*bubbly.Event)
				char := event.Data.(string)
				field := currentField.Get().(int)

				switch field {
				case 0:
					cardNumber.Set(cardNumber.Get().(string) + char)
				case 1:
					email.Set(email.Get().(string) + char)
				case 2:
					phone.Set(phone.Get().(string) + char)
				}
			})

			ctx.On("backspace", func(data interface{}) {
				field := currentField.Get().(int)

				switch field {
				case 0:
					c := cardNumber.Get().(string)
					if len(c) > 0 {
						cardNumber.Set(c[:len(c)-1])
					}
				case 1:
					e := email.Get().(string)
					if len(e) > 0 {
						email.Set(e[:len(e)-1])
					}
				case 2:
					p := phone.Get().(string)
					if len(p) > 0 {
						phone.Set(p[:len(p)-1])
					}
				}
			})

			ctx.On("next-field", func(data interface{}) {
				field := currentField.Get().(int)
				if field < 2 {
					currentField.Set(field + 1)
				}
			})

			ctx.On("prev-field", func(data interface{}) {
				field := currentField.Get().(int)
				if field > 0 {
					currentField.Set(field - 1)
				}
			})

			ctx.On("submit", func(data interface{}) {
				card := cardNumber.Get().(string)
				emailVal := email.Get().(string)
				phoneVal := phone.Get().(string)

				observability.RecordBreadcrumb("user", "Payment form submitted", map[string]interface{}{
					"cardNumber": card,
					"email":      emailVal,
					"phone":      phoneVal,
				})

				statusMessage.Set("Payment processed!")
			})

			ctx.On("trigger-error", func(data interface{}) {
				// Report error with sensitive data - should be filtered
				if reporter := observability.GetErrorReporter(); reporter != nil {
					reporter.ReportError(
						fmt.Errorf("payment processing failed for card %s", cardNumber.Get().(string)),
						&observability.ErrorContext{
							ComponentName: "PaymentForm",
							ComponentID:   "payment-1",
							EventName:     "trigger-error",
							Timestamp:     time.Now(),
							Tags: map[string]string{
								"environment": "production",
								"payment_type": "credit_card",
							},
							Extra: map[string]interface{}{
								"card_number":  cardNumber.Get().(string),
								"email":        email.Get().(string),
								"phone":        phone.Get().(string),
								"password":     "super_secret_123", // Should be redacted
								"api_key":      "sk_live_abc123",   // Should be redacted
								"user_message": "My email is test@example.com and phone is 555-123-4567",
								"ssn":          "123-45-6789", // Should be redacted
							},
							Breadcrumbs: observability.GetBreadcrumbs(),
						},
					)
				}

				statusMessage.Set("Error reported (check errors.json for filtered data)")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			cardNumber := ctx.Get("cardNumber").(*bubbly.Ref[interface{}])
			email := ctx.Get("email").(*bubbly.Ref[interface{}])
			phone := ctx.Get("phone").(*bubbly.Ref[interface{}])
			currentField := ctx.Get("currentField").(*bubbly.Ref[interface{}])
			statusMessage := ctx.Get("statusMessage").(*bubbly.Ref[interface{}])

			cardVal := cardNumber.Get().(string)
			emailVal := email.Get().(string)
			phoneVal := phone.Get().(string)
			fieldVal := currentField.Get().(int)
			statusVal := statusMessage.Get().(string)

			activeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(50)

			inactiveStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(0, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(50)

			// Card field
			cardStyle := inactiveStyle
			if fieldVal == 0 {
				cardStyle = activeStyle
			}
			maskedCard := strings.Repeat("*", len(cardVal))
			if len(cardVal) > 4 {
				maskedCard = strings.Repeat("*", len(cardVal)-4) + cardVal[len(cardVal)-4:]
			}
			cardBox := cardStyle.Render(fmt.Sprintf("Card Number: %s", maskedCard))

			// Email field
			emailStyle := inactiveStyle
			if fieldVal == 1 {
				emailStyle = activeStyle
			}
			emailBox := emailStyle.Render(fmt.Sprintf("Email: %s", emailVal))

			// Phone field
			phoneStyle := inactiveStyle
			if fieldVal == 2 {
				phoneStyle = activeStyle
			}
			phoneBox := phoneStyle.Render(fmt.Sprintf("Phone: %s", phoneVal))

			// Status
			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Width(50)

			statusBox := ""
			if statusVal != "" {
				statusBox = statusStyle.Render(statusVal)
			}

			// Privacy info
			privacyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(50)

			privacyText := "🔒 Privacy Filters Active:\n"
			privacyText += "• Emails → [EMAIL_REDACTED]\n"
			privacyText += "• Phones → [PHONE_REDACTED]\n"
			privacyText += "• SSNs → [SSN_REDACTED]\n"
			privacyText += "• Credit Cards → [CC_REDACTED]\n"
			privacyText += "• Passwords/Secrets → [REDACTED]"

			privacyBox := privacyStyle.Render(privacyText)

			result := lipgloss.JoinVertical(
				lipgloss.Left,
				cardBox,
				emailBox,
				phoneBox,
			)

			if statusBox != "" {
				result = lipgloss.JoinVertical(lipgloss.Left, result, "", statusBox)
			}

			result = lipgloss.JoinVertical(lipgloss.Left, result, "", privacyBox)

			return result
		}).
		Build()
}

func main() {
	// Setup custom file reporter with privacy filtering
	reporter := NewFileReporter("errors.json")
	observability.SetErrorReporter(reporter)
	defer reporter.Flush(5 * time.Second)

	observability.RecordBreadcrumb("navigation", "Application started", map[string]interface{}{
		"example": "custom-reporter",
		"privacy": "enabled",
	})

	form, err := createPaymentForm()
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
