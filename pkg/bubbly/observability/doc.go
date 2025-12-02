// Package observability provides error tracking, breadcrumbs, and monitoring for BubblyUI applications.
//
// # Overview
//
// The observability package enables comprehensive error tracking and debugging capabilities
// for BubblyUI applications. It provides a pluggable error reporting system, breadcrumb trails
// for debugging, and integration with popular error tracking services like Sentry.
//
// # Error Reporting
//
// The package supports multiple error reporting backends through the ErrorReporter interface:
//
//   - ConsoleReporter: Logs errors to stdout (development)
//   - SentryReporter: Sends errors to Sentry (production)
//   - Custom implementations: Implement ErrorReporter for other services
//
// Basic setup:
//
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"
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
//
// # Breadcrumbs
//
// Breadcrumbs provide a trail of events leading up to an error, making debugging easier.
// They are automatically included in error reports when using Sentry or custom reporters.
//
//	// Record breadcrumbs during application execution
//	observability.RecordBreadcrumb("navigation", "User navigated to /users", map[string]interface{}{
//	    "path": "/users",
//	    "method": "push",
//	})
//
//	observability.RecordBreadcrumb("ui", "Button clicked", map[string]interface{}{
//	    "component": "SubmitButton",
//	    "action": "submit",
//	})
//
//	// Get all breadcrumbs
//	crumbs := observability.GetBreadcrumbs()
//
//	// Clear breadcrumbs after error is reported
//	observability.ClearBreadcrumbs()
//
// # Error Types
//
// The package defines specialized error types for BubblyUI:
//
//   - HandlerPanicError: Wraps panics in event handlers
//   - CommandGenerationError: Wraps panics during automatic command generation
//
// These types provide rich context for debugging:
//
//	err := &observability.HandlerPanicError{
//	    ComponentName: "Counter",
//	    EventName:     "increment",
//	    PanicValue:    "nil pointer dereference",
//	}
//	fmt.Println(err.Error())
//	// Output: panic in event handler: component 'Counter', event 'increment', panic: nil pointer dereference
//
// # Error Context
//
// When reporting errors, include rich context for easier debugging:
//
//	reporter.ReportPanic(err, &observability.ErrorContext{
//	    ComponentName: "Counter",
//	    ComponentID:   "counter-123",
//	    EventName:     "increment",
//	    Timestamp:     time.Now(),
//	    StackTrace:    debug.Stack(),
//	})
//
// # Thread Safety
//
// All functions and types in this package are thread-safe:
//
//   - SetErrorReporter/GetErrorReporter use atomic operations
//   - Breadcrumb recording is protected by sync.RWMutex
//   - All reporter implementations must be concurrent-safe
//
// # Integration with BubblyUI
//
// The observability package integrates automatically with BubblyUI components:
//
//   - Event handler panics are captured and reported
//   - Command generation failures are tracked
//   - Component lifecycle errors include full context
//   - Reactive cascade errors include dependency information
//
// # Performance
//
// The package is designed for minimal overhead:
//
//   - No-op when no reporter is configured (single nil check)
//   - Breadcrumb recording: < 100ns per breadcrumb
//   - Error reporting: Async by default (Sentry)
//   - Circular buffer for breadcrumbs (configurable max)
//
// # Best Practices
//
//  1. Always configure an error reporter in production
//  2. Use breadcrumbs liberally for debugging context
//  3. Include component and event information in error context
//  4. Flush the reporter before application exit
//  5. Use environment-specific reporters (console for dev, Sentry for prod)
package observability
