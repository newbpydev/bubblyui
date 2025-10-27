package observability

import (
	"log"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ConsoleReporter is a simple error reporter that logs errors to the console.
// It's designed for development and debugging, providing immediate feedback
// about errors without requiring external services.
//
// The reporter supports two modes:
//   - Verbose mode: Includes full stack traces in output
//   - Non-verbose mode: Only logs error messages without stack traces
//
// Thread-safe: All methods are safe for concurrent use.
//
// Example usage:
//
//	// Development: Verbose console reporter
//	reporter := NewConsoleReporter(true)
//	SetErrorReporter(reporter)
//
//	// Production: Non-verbose console reporter
//	reporter := NewConsoleReporter(false)
//	SetErrorReporter(reporter)
type ConsoleReporter struct {
	// verbose controls whether stack traces are included in output
	verbose bool

	// mu protects concurrent access to log output
	mu sync.Mutex
}

// NewConsoleReporter creates a new console error reporter.
//
// Parameters:
//   - verbose: If true, includes stack traces in error output.
//     If false, only logs error messages.
//
// Returns:
//   - *ConsoleReporter: A new console reporter instance
//
// Example:
//
//	// Verbose mode for development
//	reporter := NewConsoleReporter(true)
//
//	// Non-verbose mode for production
//	reporter := NewConsoleReporter(false)
//
// Thread-safe: The returned reporter is safe for concurrent use.
func NewConsoleReporter(verbose bool) *ConsoleReporter {
	return &ConsoleReporter{
		verbose: verbose,
	}
}

// ReportPanic reports a panic that occurred in an event handler.
// Logs the panic to stderr with component and event information.
//
// If verbose mode is enabled and a stack trace is available,
// it will be included in the output.
//
// Parameters:
//   - err: The HandlerPanicError containing panic details
//   - ctx: Rich context about where and when the panic occurred
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example output (verbose mode):
//
//	2024/01/01 12:00:00 [ERROR] Panic in component 'Button' event 'click': unexpected error
//	2024/01/01 12:00:00 Stack trace:
//	goroutine 1 [running]:
//	main.main()
//	    /path/to/main.go:42 +0x123
//
// Example output (non-verbose mode):
//
//	2024/01/01 12:00:00 [ERROR] Panic in component 'Button' event 'click': unexpected error
func (r *ConsoleReporter) ReportPanic(err *bubbly.HandlerPanicError, ctx *ErrorContext) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Log panic with component and event information
	log.Printf("[ERROR] Panic in component '%s' event '%s': %v",
		ctx.ComponentName, ctx.EventName, err.PanicValue)

	// Include stack trace if verbose mode is enabled and stack trace is available
	if r.verbose && len(ctx.StackTrace) > 0 {
		log.Printf("Stack trace:\n%s", ctx.StackTrace)
	}
}

// ReportError reports a general error.
// Logs the error to stderr with component information.
//
// If verbose mode is enabled and a stack trace is available,
// it will be included in the output.
//
// Parameters:
//   - err: The error to report
//   - ctx: Rich context about where and when the error occurred
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example output (verbose mode):
//
//	2024/01/01 12:00:00 [ERROR] Error in component 'Form': validation failed
//	2024/01/01 12:00:00 Stack trace:
//	goroutine 1 [running]:
//	main.validateForm()
//	    /path/to/form.go:123 +0x456
//
// Example output (non-verbose mode):
//
//	2024/01/01 12:00:00 [ERROR] Error in component 'Form': validation failed
func (r *ConsoleReporter) ReportError(err error, ctx *ErrorContext) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Log error with component information
	log.Printf("[ERROR] Error in component '%s': %v", ctx.ComponentName, err)

	// Include stack trace if verbose mode is enabled and stack trace is available
	if r.verbose && len(ctx.StackTrace) > 0 {
		log.Printf("Stack trace:\n%s", ctx.StackTrace)
	}
}

// Flush ensures all pending errors are sent before shutdown.
// For ConsoleReporter, this is a no-op since console output is immediate.
//
// Parameters:
//   - timeout: Maximum time to wait (ignored for console reporter)
//
// Returns:
//   - error: Always returns nil for console reporter
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	reporter := NewConsoleReporter(true)
//	defer reporter.Flush(5 * time.Second)
func (r *ConsoleReporter) Flush(timeout time.Duration) error {
	// Console output is immediate, nothing to flush
	return nil
}
