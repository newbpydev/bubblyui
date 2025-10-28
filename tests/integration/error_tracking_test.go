package integration

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// mockPanicReport captures a panic report for testing
type mockPanicReport struct {
	err *observability.HandlerPanicError
	ctx *observability.ErrorContext
}

// mockErrorReport captures an error report for testing
type mockErrorReport struct {
	err error
	ctx *observability.ErrorContext
}

// mockReporter is a test implementation of ErrorReporter that captures all calls
type mockReporter struct {
	mu          sync.Mutex
	panics      []mockPanicReport
	errors      []mockErrorReport
	flushCalled bool
}

func newMockReporter() *mockReporter {
	return &mockReporter{
		panics: make([]mockPanicReport, 0),
		errors: make([]mockErrorReport, 0),
	}
}

func (m *mockReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panics = append(m.panics, mockPanicReport{err: err, ctx: ctx})
}

func (m *mockReporter) ReportError(err error, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, mockErrorReport{err: err, ctx: ctx})
}

func (m *mockReporter) Flush(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushCalled = true
	return nil
}

func (m *mockReporter) getPanics() []mockPanicReport {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]mockPanicReport, len(m.panics))
	copy(result, m.panics)
	return result
}

func (m *mockReporter) getErrors() []mockErrorReport {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]mockErrorReport, len(m.errors))
	copy(result, m.errors)
	return result
}

func (m *mockReporter) wasFlushed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.flushCalled
}

// TestConsoleReporterIntegration verifies the development workflow with console reporter
func TestConsoleReporterIntegration(t *testing.T) {
	t.Run("panic reported to console", func(t *testing.T) {
		// Capture stderr output
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(nil)

		// Configure console reporter
		reporter := observability.NewConsoleReporter(true)
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		// Create component with panicking handler
		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("trigger", func(data interface{}) {
					panic("test panic")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Trigger panic
		component.Emit("trigger", nil)

		// Give time for async reporting
		time.Sleep(10 * time.Millisecond)

		// Verify error was logged to console
		output := buf.String()
		assert.Contains(t, output, "[ERROR]")
		assert.Contains(t, output, "TestComponent")
		assert.Contains(t, output, "trigger")
		assert.Contains(t, output, "test panic")
	})

	t.Run("verbose mode shows stack trace", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(nil)

		// Verbose reporter
		reporter := observability.NewConsoleReporter(true)
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		component, _ := bubbly.NewComponent("VerboseTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					panic("verbose panic")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()
		component.Emit("action", nil)
		time.Sleep(10 * time.Millisecond)

		output := buf.String()
		// Verbose mode should include stack trace
		assert.Contains(t, output, "Stack trace:")
	})

	t.Run("non-verbose mode hides stack trace", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(nil)

		// Non-verbose reporter
		reporter := observability.NewConsoleReporter(false)
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		component, _ := bubbly.NewComponent("NonVerboseTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					panic("non-verbose panic")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()
		component.Emit("action", nil)
		time.Sleep(10 * time.Millisecond)

		output := buf.String()
		// Non-verbose mode should not include stack trace
		assert.NotContains(t, output, "Stack trace:")
		assert.Contains(t, output, "[ERROR]")
	})
}

// TestSentryReporterIntegration verifies the production workflow with Sentry reporter
func TestSentryReporterIntegration(t *testing.T) {
	t.Run("panic reported to Sentry", func(t *testing.T) {
		// Use empty DSN for testing (noopTransport)
		reporter, err := observability.NewSentryReporter("")
		require.NoError(t, err)
		require.NotNil(t, reporter)

		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		// Create component with panicking handler
		component, err := bubbly.NewComponent("SentryTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("submit", func(data interface{}) {
					panic("sentry test panic")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Sentry Test"
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Trigger panic
		component.Emit("submit", nil)

		// Give time for async reporting
		time.Sleep(10 * time.Millisecond)

		// Flush to ensure all events are sent
		err = reporter.Flush(2 * time.Second)
		assert.NoError(t, err)
	})

	t.Run("Sentry with options", func(t *testing.T) {
		reporter, err := observability.NewSentryReporter("",
			observability.WithEnvironment("test"),
			observability.WithRelease("v1.0.0"),
			observability.WithDebug(false),
		)
		require.NoError(t, err)
		require.NotNil(t, reporter)

		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		component, _ := bubbly.NewComponent("SentryOptionsTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					panic("options test")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()
		component.Emit("action", nil)
		time.Sleep(10 * time.Millisecond)

		err = reporter.Flush(2 * time.Second)
		assert.NoError(t, err)
	})
}

// TestCustomReporterIntegration verifies custom reporter implementation pattern
func TestCustomReporterIntegration(t *testing.T) {
	t.Run("custom reporter receives panic", func(t *testing.T) {
		reporter := newMockReporter()
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		// Create component with panicking handler
		component, _ := bubbly.NewComponent("CustomTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("click", func(data interface{}) {
					panic("custom panic")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Custom Test"
			}).
			Build()

		component.Init()

		// Trigger panic
		component.Emit("click", map[string]interface{}{"button": "submit"})
		time.Sleep(10 * time.Millisecond)

		// Verify panic was reported
		panics := reporter.getPanics()
		require.Len(t, panics, 1)

		panicReport := panics[0]
		assert.Equal(t, "CustomTest", panicReport.err.ComponentName)
		assert.Equal(t, "click", panicReport.err.EventName)
		assert.Equal(t, "custom panic", panicReport.err.PanicValue)

		// Verify ErrorContext
		assert.Equal(t, "CustomTest", panicReport.ctx.ComponentName)
		assert.NotEmpty(t, panicReport.ctx.ComponentID)
		assert.Equal(t, "click", panicReport.ctx.EventName)
		assert.NotZero(t, panicReport.ctx.Timestamp)
		assert.NotEmpty(t, panicReport.ctx.StackTrace)
	})

	t.Run("manual error reporting", func(t *testing.T) {
		reporter := newMockReporter()
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		// Manually report an error
		testErr := fmt.Errorf("validation failed")
		ctx := &observability.ErrorContext{
			ComponentName: "FormComponent",
			ComponentID:   "form-123",
			EventName:     "validate",
			Timestamp:     time.Now(),
			Tags: map[string]string{
				"field": "email",
			},
			Extra: map[string]interface{}{
				"value": "invalid-email",
			},
		}

		if r := observability.GetErrorReporter(); r != nil {
			r.ReportError(testErr, ctx)
		}

		// Verify error was reported
		errors := reporter.getErrors()
		require.Len(t, errors, 1)

		errorReport := errors[0]
		assert.Equal(t, testErr, errorReport.err)
		assert.Equal(t, "FormComponent", errorReport.ctx.ComponentName)
		assert.Equal(t, "form-123", errorReport.ctx.ComponentID)
		assert.Equal(t, "validate", errorReport.ctx.EventName)
		assert.Equal(t, "email", errorReport.ctx.Tags["field"])
		assert.Equal(t, "invalid-email", errorReport.ctx.Extra["value"])
	})

	t.Run("flush is called", func(t *testing.T) {
		reporter := newMockReporter()
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		err := reporter.Flush(1 * time.Second)
		assert.NoError(t, err)
		assert.True(t, reporter.wasFlushed())
	})

	t.Run("nil reporter is safe", func(t *testing.T) {
		// Ensure no reporter is configured
		observability.SetErrorReporter(nil)

		// Create component with panicking handler
		component, _ := bubbly.NewComponent("NilReporterTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					panic("should not crash")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()

		// Should not crash even with nil reporter
		component.Emit("action", nil)
		time.Sleep(10 * time.Millisecond)

		// Test passes if we get here without crashing
	})
}

// TestBreadcrumbIntegration verifies breadcrumb trail collection
func TestBreadcrumbIntegration(t *testing.T) {
	t.Run("breadcrumbs included in error context", func(t *testing.T) {
		reporter := newMockReporter()
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		// Record breadcrumbs
		observability.RecordBreadcrumb("navigation", "User opened form", nil)
		observability.RecordBreadcrumb("user", "User entered email", map[string]interface{}{
			"field": "email",
		})
		observability.RecordBreadcrumb("user", "User clicked submit", nil)

		// Create component that will panic
		component, _ := bubbly.NewComponent("BreadcrumbTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("submit", func(data interface{}) {
					panic("submit failed")
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Form"
			}).
			Build()

		component.Init()
		component.Emit("submit", nil)
		time.Sleep(10 * time.Millisecond)

		// Verify breadcrumbs were included
		panics := reporter.getPanics()
		require.Len(t, panics, 1)

		breadcrumbs := panics[0].ctx.Breadcrumbs
		assert.GreaterOrEqual(t, len(breadcrumbs), 3)

		// Verify breadcrumb content
		found := false
		for _, bc := range breadcrumbs {
			if bc.Message == "User entered email" {
				found = true
				assert.Equal(t, "user", bc.Category)
				assert.Equal(t, "email", bc.Data["field"])
			}
		}
		assert.True(t, found, "Should find breadcrumb with email message")

		// Cleanup
		observability.ClearBreadcrumbs()
	})

	t.Run("breadcrumb FIFO eviction", func(t *testing.T) {
		observability.ClearBreadcrumbs()

		// Record more than max breadcrumbs (100)
		for i := 0; i < 150; i++ {
			observability.RecordBreadcrumb("test", fmt.Sprintf("Breadcrumb %d", i), nil)
		}

		breadcrumbs := observability.GetBreadcrumbs()
		assert.Len(t, breadcrumbs, 100, "Should keep only 100 breadcrumbs")

		// First breadcrumb should be #50 (oldest 50 dropped)
		assert.Contains(t, breadcrumbs[0].Message, "Breadcrumb 50")
		// Last breadcrumb should be #149
		assert.Contains(t, breadcrumbs[99].Message, "Breadcrumb 149")

		observability.ClearBreadcrumbs()
	})

	t.Run("breadcrumbs are chronological", func(t *testing.T) {
		observability.ClearBreadcrumbs()

		start := time.Now()
		observability.RecordBreadcrumb("test", "First", nil)
		time.Sleep(5 * time.Millisecond)
		observability.RecordBreadcrumb("test", "Second", nil)
		time.Sleep(5 * time.Millisecond)
		observability.RecordBreadcrumb("test", "Third", nil)

		breadcrumbs := observability.GetBreadcrumbs()
		require.Len(t, breadcrumbs, 3)

		// Verify timestamps are in order
		assert.True(t, breadcrumbs[0].Timestamp.Before(breadcrumbs[1].Timestamp))
		assert.True(t, breadcrumbs[1].Timestamp.Before(breadcrumbs[2].Timestamp))

		// All should be after start time
		for _, bc := range breadcrumbs {
			assert.True(t, bc.Timestamp.After(start) || bc.Timestamp.Equal(start))
		}

		observability.ClearBreadcrumbs()
	})
}

// TestPrivacyFiltering verifies PII sanitization in custom reporters
func TestPrivacyFiltering(t *testing.T) {
	t.Run("sensitive data filtered", func(t *testing.T) {
		// Helper to sanitize data
		sanitize := func(data map[string]interface{}) map[string]interface{} {
			sanitized := make(map[string]interface{})
			emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
			phoneRegex := regexp.MustCompile(`\d{3}-\d{3}-\d{4}`)

			for k, v := range data {
				if str, ok := v.(string); ok {
					// Filter emails
					str = emailRegex.ReplaceAllString(str, "[EMAIL_REDACTED]")
					// Filter phone numbers
					str = phoneRegex.ReplaceAllString(str, "[PHONE_REDACTED]")
					// Filter sensitive keys
					if strings.Contains(strings.ToLower(k), "password") ||
						strings.Contains(strings.ToLower(k), "secret") ||
						strings.Contains(strings.ToLower(k), "token") {
						str = "[REDACTED]"
					}
					sanitized[k] = str
				} else {
					sanitized[k] = v
				}
			}
			return sanitized
		}

		// Create custom filtering reporter
		type filteringReporter struct {
			*mockReporter
			sanitize func(map[string]interface{}) map[string]interface{}
		}

		baseReporter := newMockReporter()
		reporter := &filteringReporter{
			mockReporter: baseReporter,
			sanitize:     sanitize,
		}

		// Override ReportError to add filtering
		reportError := func(err error, ctx *observability.ErrorContext) {
			// Sanitize Extra data before reporting
			if ctx.Extra != nil {
				ctx.Extra = reporter.sanitize(ctx.Extra)
			}
			reporter.mockReporter.ReportError(err, ctx)
		}

		// Manually call reportError instead of using the interface
		observability.SetErrorReporter(baseReporter)
		defer observability.SetErrorReporter(nil)

		// Report error with sensitive data
		testErr := fmt.Errorf("validation error")
		ctx := &observability.ErrorContext{
			ComponentName: "PaymentForm",
			Extra: map[string]interface{}{
				"email":    "user@example.com",
				"phone":    "555-123-4567",
				"password": "secret123",
				"name":     "John Doe",
			},
		}

		// Use our filtering wrapper
		reportError(testErr, ctx)

		// Verify data was sanitized
		errors := baseReporter.getErrors()
		require.Len(t, errors, 1)

		extra := errors[0].ctx.Extra
		assert.Equal(t, "[EMAIL_REDACTED]", extra["email"])
		assert.Equal(t, "[PHONE_REDACTED]", extra["phone"])
		assert.Equal(t, "[REDACTED]", extra["password"])
		assert.Equal(t, "John Doe", extra["name"]) // Name not filtered
	})
}

// TestPerformanceImpact verifies minimal overhead when reporter is configured
func TestPerformanceImpact(t *testing.T) {
	t.Run("overhead with nil reporter", func(t *testing.T) {
		observability.SetErrorReporter(nil)

		component, _ := bubbly.NewComponent("PerfTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					// Normal handler, no panic
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()

		// Measure time for 1000 events
		start := time.Now()
		for i := 0; i < 1000; i++ {
			component.Emit("action", nil)
		}
		durationNil := time.Since(start)

		t.Logf("1000 events with nil reporter: %v", durationNil)
		assert.Less(t, durationNil.Milliseconds(), int64(100), "Should be fast with nil reporter")
	})

	t.Run("overhead with configured reporter", func(t *testing.T) {
		reporter := newMockReporter()
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		component, _ := bubbly.NewComponent("PerfTest2").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					// Normal handler, no panic
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()

		// Measure time for 1000 events
		start := time.Now()
		for i := 0; i < 1000; i++ {
			component.Emit("action", nil)
		}
		durationWithReporter := time.Since(start)

		t.Logf("1000 events with reporter: %v", durationWithReporter)
		// Should still be fast (minimal overhead for non-panicking handlers)
		assert.Less(t, durationWithReporter.Milliseconds(), int64(200), "Should have minimal overhead")
	})

	t.Run("concurrent event handling with reporter", func(t *testing.T) {
		reporter := newMockReporter()
		observability.SetErrorReporter(reporter)
		defer observability.SetErrorReporter(nil)

		component, _ := bubbly.NewComponent("ConcurrentPerfTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					// Normal handler
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Test"
			}).
			Build()

		component.Init()

		// Concurrent event emission
		var wg sync.WaitGroup
		const numGoroutines = 10
		const eventsPerGoroutine = 100

		start := time.Now()
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < eventsPerGoroutine; j++ {
					component.Emit("action", nil)
				}
			}()
		}
		wg.Wait()
		duration := time.Since(start)

		t.Logf("%d concurrent events: %v", numGoroutines*eventsPerGoroutine, duration)
		assert.Less(t, duration.Milliseconds(), int64(500), "Concurrent events should be fast")
	})
}
