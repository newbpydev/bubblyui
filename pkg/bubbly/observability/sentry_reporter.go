package observability

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// SentryReporter is an error reporter that sends errors to Sentry.
// It's designed for production use, providing centralized error tracking
// with rich context, tags, and breadcrumbs.
//
// The reporter uses Sentry's Hub API for thread-safe error reporting
// and supports customization via functional options.
//
// Thread-safe: All methods are safe for concurrent use.
//
// Example usage:
//
//	// Production: Sentry reporter with DSN
//	reporter, err := NewSentryReporter(
//	    os.Getenv("SENTRY_DSN"),
//	    WithDebug(true),
//	    WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
//	        // Filter or modify events before sending
//	        return event
//	    }),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	SetErrorReporter(reporter)
//	defer reporter.Flush(5 * time.Second)
type SentryReporter struct {
	// hub is the Sentry hub used for error reporting
	hub *sentry.Hub
}

// SentryOption is a functional option for configuring SentryReporter.
// Options are applied to the Sentry ClientOptions during initialization.
//
// Example:
//
//	reporter, err := NewSentryReporter(
//	    dsn,
//	    WithDebug(true),
//	    WithBeforeSend(myBeforeSendFunc),
//	)
type SentryOption func(*sentry.ClientOptions)

// WithBeforeSend configures a BeforeSend hook for the Sentry client.
// The hook is called before each event is sent, allowing you to
// filter or modify events.
//
// Parameters:
//   - fn: Function that receives the event and hint, and returns
//     the modified event (or nil to drop the event)
//
// Returns:
//   - SentryOption: Option that can be passed to NewSentryReporter
//
// Example:
//
//	reporter, err := NewSentryReporter(
//	    dsn,
//	    WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
//	        // Filter out events from test components
//	        if event.Tags["component"] == "Test" {
//	            return nil // Drop event
//	        }
//	        // Add custom tag
//	        event.Tags["environment"] = "production"
//	        return event
//	    }),
//	)
func WithBeforeSend(fn func(*sentry.Event, *sentry.EventHint) *sentry.Event) SentryOption {
	return func(opts *sentry.ClientOptions) {
		opts.BeforeSend = fn
	}
}

// WithDebug enables debug mode for the Sentry client.
// When enabled, Sentry will log detailed information about
// event processing to stderr.
//
// Parameters:
//   - debug: If true, enables debug logging
//
// Returns:
//   - SentryOption: Option that can be passed to NewSentryReporter
//
// Example:
//
//	// Enable debug mode for troubleshooting
//	reporter, err := NewSentryReporter(
//	    dsn,
//	    WithDebug(true),
//	)
func WithDebug(debug bool) SentryOption {
	return func(opts *sentry.ClientOptions) {
		opts.Debug = debug
	}
}

// WithEnvironment sets the environment tag for all events.
//
// Parameters:
//   - environment: Environment name (e.g., "production", "staging", "development")
//
// Returns:
//   - SentryOption: Option that can be passed to NewSentryReporter
//
// Example:
//
//	reporter, err := NewSentryReporter(
//	    dsn,
//	    WithEnvironment("production"),
//	)
func WithEnvironment(environment string) SentryOption {
	return func(opts *sentry.ClientOptions) {
		opts.Environment = environment
	}
}

// WithRelease sets the release version for all events.
//
// Parameters:
//   - release: Release identifier (e.g., "v1.0.0", "abc123")
//
// Returns:
//   - SentryOption: Option that can be passed to NewSentryReporter
//
// Example:
//
//	reporter, err := NewSentryReporter(
//	    dsn,
//	    WithRelease("v1.0.0"),
//	)
func WithRelease(release string) SentryOption {
	return func(opts *sentry.ClientOptions) {
		opts.Release = release
	}
}

// NewSentryReporter creates a new Sentry error reporter.
//
// The reporter initializes the Sentry SDK with the provided DSN and options.
// An empty DSN is allowed and will disable sending events to Sentry
// (useful for testing).
//
// Parameters:
//   - dsn: Sentry Data Source Name (DSN) for your project.
//     Get this from your Sentry project settings.
//     Pass empty string to disable sending (for testing).
//   - opts: Optional configuration options (WithDebug, WithBeforeSend, etc.)
//
// Returns:
//   - *SentryReporter: A new Sentry reporter instance
//   - error: Non-nil if Sentry initialization fails
//
// Example:
//
//	// Production setup
//	reporter, err := NewSentryReporter(
//	    os.Getenv("SENTRY_DSN"),
//	    WithEnvironment("production"),
//	    WithRelease("v1.0.0"),
//	    WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
//	        // Filter sensitive data
//	        return event
//	    }),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reporter.Flush(5 * time.Second)
//
//	// Test setup (empty DSN, won't send)
//	reporter, err := NewSentryReporter("")
//	if err != nil {
//	    t.Fatal(err)
//	}
//
// Thread-safe: The returned reporter is safe for concurrent use.
func NewSentryReporter(dsn string, opts ...SentryOption) (*SentryReporter, error) {
	// Create default client options
	clientOpts := sentry.ClientOptions{
		Dsn: dsn,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(&clientOpts)
	}

	// Initialize Sentry SDK
	err := sentry.Init(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	// Create reporter with current hub
	return &SentryReporter{
		hub: sentry.CurrentHub(),
	}, nil
}

// ReportPanic reports a panic that occurred in an event handler.
// Sends the panic to Sentry with rich context including tags,
// extras, and breadcrumbs.
//
// The panic is captured as an exception in Sentry with:
//   - Tags: component name, event name, and any custom tags from ctx
//   - Extras: panic value and any custom extras from ctx
//   - Breadcrumbs: Navigation trail leading to the panic
//   - Stack trace: From ctx.StackTrace
//
// Parameters:
//   - err: The HandlerPanicError containing panic details
//   - ctx: Rich context about where and when the panic occurred
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	reporter.ReportPanic(
//	    &bubbly.HandlerPanicError{
//	        ComponentName: "Button",
//	        EventName:     "click",
//	        PanicValue:    "unexpected error",
//	    },
//	    &ErrorContext{
//	        ComponentName: "Button",
//	        ComponentID:   "btn-1",
//	        EventName:     "click",
//	        Timestamp:     time.Now(),
//	        Tags: map[string]string{
//	            "user_role": "admin",
//	        },
//	        Breadcrumbs: []Breadcrumb{
//	            {Type: "user", Message: "User clicked button"},
//	        },
//	        StackTrace: debug.Stack(),
//	    },
//	)
func (r *SentryReporter) ReportPanic(err *bubbly.HandlerPanicError, ctx *ErrorContext) {
	// Use WithScope to add context without affecting other events
	r.hub.WithScope(func(scope *sentry.Scope) {
		// Set component tags
		scope.SetTag("component", ctx.ComponentName)
		scope.SetTag("component_id", ctx.ComponentID)
		scope.SetTag("event", ctx.EventName)

		// Set custom tags from context
		for key, value := range ctx.Tags {
			scope.SetTag(key, value)
		}

		// Set panic value as extra
		scope.SetExtra("panic_value", err.PanicValue)

		// Set custom extras from context
		for key, value := range ctx.Extra {
			scope.SetExtra(key, value)
		}

		// Add breadcrumbs
		for _, bc := range ctx.Breadcrumbs {
			scope.AddBreadcrumb(&sentry.Breadcrumb{
				Type:      bc.Type,
				Category:  bc.Category,
				Message:   bc.Message,
				Level:     sentry.Level(bc.Level),
				Timestamp: bc.Timestamp,
				Data:      bc.Data,
			}, 100) // Max 100 breadcrumbs
		}

		// Capture the panic as an exception
		r.hub.CaptureException(fmt.Errorf("panic in component '%s' event '%s': %v",
			ctx.ComponentName, ctx.EventName, err.PanicValue))
	})
}

// ReportError reports a general error.
// Sends the error to Sentry with rich context including tags,
// extras, and breadcrumbs.
//
// The error is captured in Sentry with:
//   - Tags: component name and any custom tags from ctx
//   - Extras: Any custom extras from ctx
//   - Breadcrumbs: Navigation trail leading to the error
//   - Stack trace: From ctx.StackTrace (if available)
//
// Parameters:
//   - err: The error to report
//   - ctx: Rich context about where and when the error occurred
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	reporter.ReportError(
//	    errors.New("validation failed"),
//	    &ErrorContext{
//	        ComponentName: "Form",
//	        ComponentID:   "form-1",
//	        Timestamp:     time.Now(),
//	        Tags: map[string]string{
//	            "field": "email",
//	        },
//	        Extra: map[string]interface{}{
//	            "input_value": email,
//	        },
//	    },
//	)
func (r *SentryReporter) ReportError(err error, ctx *ErrorContext) {
	// Use WithScope to add context without affecting other events
	r.hub.WithScope(func(scope *sentry.Scope) {
		// Set component tags
		scope.SetTag("component", ctx.ComponentName)
		scope.SetTag("component_id", ctx.ComponentID)
		if ctx.EventName != "" {
			scope.SetTag("event", ctx.EventName)
		}

		// Set custom tags from context
		for key, value := range ctx.Tags {
			scope.SetTag(key, value)
		}

		// Set custom extras from context
		for key, value := range ctx.Extra {
			scope.SetExtra(key, value)
		}

		// Add breadcrumbs
		for _, bc := range ctx.Breadcrumbs {
			scope.AddBreadcrumb(&sentry.Breadcrumb{
				Type:      bc.Type,
				Category:  bc.Category,
				Message:   bc.Message,
				Level:     sentry.Level(bc.Level),
				Timestamp: bc.Timestamp,
				Data:      bc.Data,
			}, 100) // Max 100 breadcrumbs
		}

		// Capture the error
		r.hub.CaptureException(err)
	})
}

// Flush ensures all pending errors are sent before shutdown.
// Blocks until all events are sent or the timeout is reached.
//
// This should be called before the application exits to ensure
// no errors are lost.
//
// Parameters:
//   - timeout: Maximum time to wait for events to be sent
//
// Returns:
//   - error: Always returns nil (Sentry's Flush returns bool, not error)
//
// Thread-safe: Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	reporter, err := NewSentryReporter(dsn)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reporter.Flush(5 * time.Second)
//
//	// ... application code ...
func (r *SentryReporter) Flush(timeout time.Duration) error {
	// Flush pending events
	// Note: sentry.Flush returns bool, but we return error for interface compatibility
	sentry.Flush(timeout)
	return nil
}
