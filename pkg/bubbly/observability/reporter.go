package observability

import (
	"fmt"
	"sync"
	"time"
)

// HandlerPanicError wraps a panic that occurred in an event handler.
// This allows the application to continue running even if a handler panics.
//
// This type is defined here to avoid import cycles between bubbly and observability packages.
type HandlerPanicError struct {
	// ComponentName is the name of the component where the panic occurred
	ComponentName string
	// EventName is the name of the event being handled
	EventName string
	// PanicValue is the value passed to panic()
	PanicValue interface{}
}

// Error implements the error interface for HandlerPanicError.
func (e *HandlerPanicError) Error() string {
	return fmt.Sprintf("panic in event handler: component '%s', event '%s', panic: %v",
		e.ComponentName, e.EventName, e.PanicValue)
}

// ErrorReporter is a pluggable interface for error tracking backends.
// Implementations can send errors to services like Sentry, Rollbar, or custom backends.
//
// The interface is optional - if no reporter is configured via SetErrorReporter,
// errors are silently ignored with zero overhead (just a nil check).
//
// Thread-safe: All methods must be safe for concurrent use by multiple goroutines.
//
// Example usage:
//
//	// Development: Console reporter
//	reporter := NewConsoleReporter(true)
//	SetErrorReporter(reporter)
//
//	// Production: Sentry reporter
//	reporter, err := NewSentryReporter(os.Getenv("SENTRY_DSN"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	SetErrorReporter(reporter)
//	defer reporter.Flush(5 * time.Second)
//
// Integration with component error handling:
//
//	if reporter := GetErrorReporter(); reporter != nil {
//	    reporter.ReportPanic(panicErr, &ErrorContext{
//	        ComponentName: c.name,
//	        ComponentID:   c.id,
//	        EventName:     event.Name,
//	        Timestamp:     time.Now(),
//	        StackTrace:    debug.Stack(),
//	    })
//	}
type ErrorReporter interface {
	// ReportPanic reports a panic that occurred in an event handler.
	// This is called automatically by the component system when a handler panics.
	//
	// Parameters:
	//   - err: The HandlerPanicError containing panic details
	//   - ctx: Rich context about where and when the panic occurred
	//
	// Thread-safe: Must be safe to call concurrently.
	ReportPanic(err *HandlerPanicError, ctx *ErrorContext)

	// ReportError reports a general error.
	// This can be called manually to report validation errors, business logic errors, etc.
	//
	// Parameters:
	//   - err: The error to report
	//   - ctx: Rich context about where and when the error occurred
	//
	// Thread-safe: Must be safe to call concurrently.
	ReportError(err error, ctx *ErrorContext)

	// Flush ensures all pending errors are sent before shutdown.
	// This should be called before the application exits to ensure no errors are lost.
	//
	// Parameters:
	//   - timeout: Maximum time to wait for pending errors to be sent
	//
	// Returns:
	//   - error: Non-nil if flush failed or timed out
	//
	// Thread-safe: Must be safe to call concurrently.
	//
	// Example:
	//   defer reporter.Flush(5 * time.Second)
	Flush(timeout time.Duration) error
}

// ErrorContext provides rich context about where and when an error occurred.
// This information helps with debugging and understanding error patterns in production.
//
// All fields are optional, but providing more context leads to better error reports.
//
// Example:
//
//	ctx := &ErrorContext{
//	    ComponentName: "LoginForm",
//	    ComponentID:   "form-123",
//	    EventName:     "submit",
//	    Timestamp:     time.Now(),
//	    Tags: map[string]string{
//	        "environment": "production",
//	        "user_type":   "premium",
//	    },
//	    Extra: map[string]interface{}{
//	        "form_data": formData,
//	        "validation_errors": errors,
//	    },
//	    Breadcrumbs: []Breadcrumb{
//	        {Type: "navigation", Message: "User navigated to login page"},
//	        {Type: "user", Message: "User entered username"},
//	        {Type: "user", Message: "User clicked submit"},
//	    },
//	    StackTrace: debug.Stack(),
//	}
type ErrorContext struct {
	// ComponentName is the name of the component where the error occurred.
	// Example: "Button", "Form", "LoginDialog"
	ComponentName string

	// ComponentID is the unique instance identifier of the component.
	// This helps distinguish between multiple instances of the same component.
	// Example: "btn-submit-123", "form-login-456"
	ComponentID string

	// EventName is the name of the event being handled when the error occurred.
	// Example: "click", "submit", "change", "keypress"
	EventName string

	// Timestamp is when the error occurred.
	// Set to time.Now() when creating the context.
	Timestamp time.Time

	// Tags are key-value pairs for filtering and grouping errors.
	// Tags should be low-cardinality values (not unique per error).
	//
	// Good tags:
	//   - "environment": "production"
	//   - "component_type": "form"
	//   - "user_role": "admin"
	//
	// Bad tags (too high cardinality):
	//   - "user_id": "12345" (use Extra instead)
	//   - "timestamp": "2024-01-01..." (already in Timestamp)
	Tags map[string]string

	// Extra contains arbitrary additional data about the error.
	// This can include high-cardinality data, complex objects, etc.
	//
	// Examples:
	//   - "user_id": "12345"
	//   - "form_data": map[string]interface{}{...}
	//   - "request_id": "req-abc-123"
	Extra map[string]interface{}

	// Breadcrumbs is a trail of actions leading up to the error.
	// This helps understand the sequence of events that caused the error.
	//
	// Breadcrumbs should be added chronologically as actions occur.
	// Most recent breadcrumb should be last in the slice.
	//
	// Example:
	//   []Breadcrumb{
	//       {Type: "navigation", Message: "User opened form"},
	//       {Type: "user", Message: "User entered email"},
	//       {Type: "user", Message: "User clicked submit"},
	//   }
	Breadcrumbs []Breadcrumb

	// StackTrace is the stack trace from where the error occurred.
	// Use debug.Stack() to capture the current stack trace.
	//
	// Example:
	//   import "runtime/debug"
	//   ctx.StackTrace = debug.Stack()
	StackTrace []byte
}

// Breadcrumb represents a single action or event in the trail leading to an error.
// Breadcrumbs help understand the sequence of events that caused an error.
//
// Inspired by Sentry's breadcrumb system.
//
// Example:
//
//	breadcrumb := Breadcrumb{
//	    Type:      "navigation",
//	    Category:  "ui",
//	    Message:   "User navigated to /login",
//	    Level:     "info",
//	    Timestamp: time.Now(),
//	    Data: map[string]interface{}{
//	        "from": "/home",
//	        "to":   "/login",
//	    },
//	}
type Breadcrumb struct {
	// Type categorizes the breadcrumb by its nature.
	//
	// Common types:
	//   - "navigation": Page or view navigation
	//   - "user": User interaction (click, input, etc.)
	//   - "http": HTTP request/response
	//   - "error": Error or warning
	//   - "debug": Debug information
	//   - "info": General information
	Type string

	// Category is a subcategory for grouping breadcrumbs.
	// This is more specific than Type.
	//
	// Examples:
	//   - "ui" (for user interactions)
	//   - "network" (for HTTP requests)
	//   - "state" (for state changes)
	//   - "validation" (for validation events)
	Category string

	// Message is a human-readable description of the breadcrumb.
	// This should be concise but descriptive.
	//
	// Examples:
	//   - "User clicked submit button"
	//   - "Form validation failed"
	//   - "HTTP request to /api/users"
	Message string

	// Level indicates the severity or importance of the breadcrumb.
	//
	// Common levels:
	//   - "debug": Detailed debugging information
	//   - "info": General information
	//   - "warning": Warning or potential issue
	//   - "error": Error occurred
	Level string

	// Timestamp is when the breadcrumb was created.
	// Set to time.Now() when creating the breadcrumb.
	Timestamp time.Time

	// Data contains arbitrary additional data about the breadcrumb.
	// This can include details specific to the breadcrumb type.
	//
	// Examples:
	//   - For navigation: {"from": "/home", "to": "/login"}
	//   - For user action: {"button": "submit", "form": "login"}
	//   - For HTTP: {"method": "POST", "url": "/api/users", "status": 200}
	Data map[string]interface{}
}

// Global error reporter state
var (
	// globalReporterMu protects access to globalReporter
	globalReporterMu sync.RWMutex

	// globalReporter is the currently configured error reporter
	// nil means no reporter is configured (errors are silently ignored)
	globalReporter ErrorReporter
)

// SetErrorReporter configures the global error reporter.
// Pass nil to disable error reporting.
//
// This function is thread-safe and can be called concurrently.
//
// The reporter will be used by the component system to report panics
// in event handlers and other errors.
//
// Example:
//
//	// Development: Console reporter
//	SetErrorReporter(NewConsoleReporter(true))
//
//	// Production: Sentry reporter
//	reporter, err := NewSentryReporter(os.Getenv("SENTRY_DSN"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	SetErrorReporter(reporter)
//	defer reporter.Flush(5 * time.Second)
//
//	// Disable reporting
//	SetErrorReporter(nil)
//
// Parameters:
//   - reporter: The error reporter to use, or nil to disable reporting
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
func SetErrorReporter(reporter ErrorReporter) {
	globalReporterMu.Lock()
	defer globalReporterMu.Unlock()
	globalReporter = reporter
}

// GetErrorReporter returns the currently configured error reporter.
// Returns nil if no reporter is configured.
//
// This function is thread-safe and can be called concurrently.
//
// The returned reporter (if non-nil) can be used to report errors manually.
//
// Example:
//
//	if reporter := GetErrorReporter(); reporter != nil {
//	    reporter.ReportError(err, &ErrorContext{
//	        ComponentName: "MyComponent",
//	        Timestamp:     time.Now(),
//	    })
//	}
//
// Returns:
//   - ErrorReporter: The current reporter, or nil if not configured
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
func GetErrorReporter() ErrorReporter {
	globalReporterMu.RLock()
	defer globalReporterMu.RUnlock()
	return globalReporter
}
