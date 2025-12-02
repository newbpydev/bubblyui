// Package observability provides error tracking, breadcrumbs, and monitoring for BubblyUI.
//
// The observability package enables comprehensive error tracking and debugging
// capabilities for BubblyUI applications. It provides a pluggable error reporting
// system, breadcrumb trails for debugging, and integration with popular error
// tracking services like Sentry.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/observability,
// providing a cleaner import path for users.
//
// # Error Reporting
//
//   - ConsoleReporter: Logs errors to stdout (development)
//   - SentryReporter: Sends errors to Sentry (production)
//   - Custom implementations: Implement ErrorReporter for other services
//
// # Breadcrumbs
//
// Breadcrumbs provide a trail of events leading up to an error:
//
//	observability.RecordBreadcrumb("navigation", "User navigated to /users", map[string]interface{}{
//	    "path": "/users",
//	})
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/observability"
//
//	// Development: Use console reporter
//	reporter := observability.NewConsoleReporter(true)
//	observability.SetErrorReporter(reporter)
//
//	// Production: Use Sentry
//	reporter, err := observability.NewSentryReporter(os.Getenv("SENTRY_DSN"),
//	    observability.WithEnvironment("production"),
//	    observability.WithRelease("v1.0.0"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	observability.SetErrorReporter(reporter)
//	defer reporter.Flush(5 * time.Second)
package observability

import (
	"github.com/getsentry/sentry-go"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// =============================================================================
// Constants
// =============================================================================

// MaxBreadcrumbs is the maximum number of breadcrumbs stored.
const MaxBreadcrumbs = observability.MaxBreadcrumbs

// =============================================================================
// Error Reporting
// =============================================================================

// ErrorReporter defines the interface for error reporting implementations.
type ErrorReporter = observability.ErrorReporter

// GetErrorReporter returns the current global error reporter.
var GetErrorReporter = observability.GetErrorReporter

// SetErrorReporter sets the global error reporter.
var SetErrorReporter = observability.SetErrorReporter

// ErrorContext provides contextual information for error reports.
type ErrorContext = observability.ErrorContext

// =============================================================================
// Console Reporter
// =============================================================================

// ConsoleReporter logs errors to stdout for development.
type ConsoleReporter = observability.ConsoleReporter

// NewConsoleReporter creates a new console reporter.
// Set verbose to true for detailed output.
var NewConsoleReporter = observability.NewConsoleReporter

// =============================================================================
// Sentry Reporter
// =============================================================================

// SentryReporter sends errors to Sentry for production monitoring.
type SentryReporter = observability.SentryReporter

// NewSentryReporter creates a new Sentry reporter with the given DSN.
func NewSentryReporter(dsn string, opts ...SentryOption) (*SentryReporter, error) {
	return observability.NewSentryReporter(dsn, opts...)
}

// SentryOption configures the Sentry reporter.
type SentryOption = observability.SentryOption

// WithEnvironment sets the Sentry environment tag.
var WithEnvironment = observability.WithEnvironment

// WithRelease sets the Sentry release version.
var WithRelease = observability.WithRelease

// WithDebug enables Sentry debug mode.
var WithDebug = observability.WithDebug

// WithBeforeSend sets a callback to modify events before sending.
func WithBeforeSend(fn func(*sentry.Event, *sentry.EventHint) *sentry.Event) SentryOption {
	return observability.WithBeforeSend(fn)
}

// =============================================================================
// Breadcrumbs
// =============================================================================

// Breadcrumb represents a single breadcrumb trail entry.
type Breadcrumb = observability.Breadcrumb

// RecordBreadcrumb adds a breadcrumb to the trail.
var RecordBreadcrumb = observability.RecordBreadcrumb

// GetBreadcrumbs returns all recorded breadcrumbs.
var GetBreadcrumbs = observability.GetBreadcrumbs

// ClearBreadcrumbs removes all recorded breadcrumbs.
var ClearBreadcrumbs = observability.ClearBreadcrumbs

// =============================================================================
// Error Types
// =============================================================================

// HandlerPanicError wraps panics that occur in event handlers.
type HandlerPanicError = observability.HandlerPanicError

// CommandGenerationError wraps panics during automatic command generation.
type CommandGenerationError = observability.CommandGenerationError
