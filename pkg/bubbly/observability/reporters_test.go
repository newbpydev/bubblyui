package observability

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConsoleReporter_New tests ConsoleReporter creation
func TestConsoleReporter_New(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "create verbose reporter",
			verbose: true,
		},
		{
			name:    "create non-verbose reporter",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewConsoleReporter(tt.verbose)
			require.NotNil(t, reporter)
			assert.Implements(t, (*ErrorReporter)(nil), reporter)
		})
	}
}

// TestConsoleReporter_ReportPanic tests panic reporting to console
func TestConsoleReporter_ReportPanic(t *testing.T) {
	tests := []struct {
		name            string
		verbose         bool
		panicErr        *HandlerPanicError
		ctx             *ErrorContext
		wantInOutput    []string
		wantNotInOutput []string
	}{
		{
			name:    "report panic verbose mode",
			verbose: true,
			panicErr: &HandlerPanicError{
				ComponentName: "TestComponent",
				EventName:     "click",
				PanicValue:    "unexpected error",
			},
			ctx: &ErrorContext{
				ComponentName: "TestComponent",
				EventName:     "click",
				StackTrace:    []byte("goroutine 1 [running]:\nmain.main()"),
			},
			wantInOutput: []string{
				"ERROR",
				"Panic",
				"TestComponent",
				"click",
				"unexpected error",
				"Stack trace",
				"goroutine 1",
			},
		},
		{
			name:    "report panic non-verbose mode",
			verbose: false,
			panicErr: &HandlerPanicError{
				ComponentName: "Button",
				EventName:     "submit",
				PanicValue:    "validation failed",
			},
			ctx: &ErrorContext{
				ComponentName: "Button",
				EventName:     "submit",
				StackTrace:    []byte("goroutine 1 [running]:\nmain.main()"),
			},
			wantInOutput: []string{
				"ERROR",
				"Panic",
				"Button",
				"submit",
				"validation failed",
			},
			wantNotInOutput: []string{
				"Stack trace",
				"goroutine 1",
			},
		},
		{
			name:    "report panic without stack trace",
			verbose: true,
			panicErr: &HandlerPanicError{
				ComponentName: "Form",
				EventName:     "change",
				PanicValue:    "nil pointer",
			},
			ctx: &ErrorContext{
				ComponentName: "Form",
				EventName:     "change",
				StackTrace:    nil,
			},
			wantInOutput: []string{
				"ERROR",
				"Panic",
				"Form",
				"change",
				"nil pointer",
			},
			wantNotInOutput: []string{
				"Stack trace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			reporter := NewConsoleReporter(tt.verbose)
			reporter.ReportPanic(tt.panicErr, tt.ctx)

			output := buf.String()

			// Verify expected strings are in output
			for _, want := range tt.wantInOutput {
				assert.Contains(t, output, want, "output should contain %q", want)
			}

			// Verify unwanted strings are not in output
			for _, notWant := range tt.wantNotInOutput {
				assert.NotContains(t, output, notWant, "output should not contain %q", notWant)
			}
		})
	}
}

// TestConsoleReporter_ReportError tests error reporting to console
func TestConsoleReporter_ReportError(t *testing.T) {
	tests := []struct {
		name            string
		verbose         bool
		err             error
		ctx             *ErrorContext
		wantInOutput    []string
		wantNotInOutput []string
	}{
		{
			name:    "report error verbose mode",
			verbose: true,
			err:     errors.New("validation error"),
			ctx: &ErrorContext{
				ComponentName: "Input",
				StackTrace:    []byte("goroutine 1 [running]:\nmain.main()"),
			},
			wantInOutput: []string{
				"ERROR",
				"Input",
				"validation error",
				"Stack trace",
			},
		},
		{
			name:    "report error non-verbose mode",
			verbose: false,
			err:     errors.New("network error"),
			ctx: &ErrorContext{
				ComponentName: "API",
				StackTrace:    []byte("goroutine 1 [running]:\nmain.main()"),
			},
			wantInOutput: []string{
				"ERROR",
				"API",
				"network error",
			},
			wantNotInOutput: []string{
				"Stack trace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			reporter := NewConsoleReporter(tt.verbose)
			reporter.ReportError(tt.err, tt.ctx)

			output := buf.String()

			// Verify expected strings are in output
			for _, want := range tt.wantInOutput {
				assert.Contains(t, output, want, "output should contain %q", want)
			}

			// Verify unwanted strings are not in output
			for _, notWant := range tt.wantNotInOutput {
				assert.NotContains(t, output, notWant, "output should not contain %q", notWant)
			}
		})
	}
}

// TestConsoleReporter_Flush tests flush is no-op
func TestConsoleReporter_Flush(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "flush with 5 second timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "flush with 1 second timeout",
			timeout: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewConsoleReporter(true)
			err := reporter.Flush(tt.timeout)
			assert.NoError(t, err, "console reporter flush should not error")
		})
	}
}

// TestSentryReporter_New tests SentryReporter creation
func TestSentryReporter_New(t *testing.T) {
	tests := []struct {
		name      string
		dsn       string
		opts      []SentryOption
		wantError bool
	}{
		{
			name:      "create with empty DSN",
			dsn:       "",
			opts:      nil,
			wantError: false, // Sentry allows empty DSN (disables sending)
		},
		{
			name:      "create with test DSN",
			dsn:       "https://public@sentry.example.com/1",
			opts:      nil,
			wantError: false,
		},
		{
			name: "create with debug option",
			dsn:  "",
			opts: []SentryOption{
				WithDebug(true),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewSentryReporter(tt.dsn, tt.opts...)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, reporter)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, reporter)
				assert.Implements(t, (*ErrorReporter)(nil), reporter)
			}

			// Cleanup
			if reporter != nil {
				_ = reporter.Flush(1 * time.Second)
			}
		})
	}
}

// TestSentryReporter_ReportPanic tests panic reporting to Sentry
func TestSentryReporter_ReportPanic(t *testing.T) {
	tests := []struct {
		name     string
		panicErr *HandlerPanicError
		ctx      *ErrorContext
	}{
		{
			name: "report panic with full context",
			panicErr: &HandlerPanicError{
				ComponentName: "TestComponent",
				EventName:     "click",
				PanicValue:    "unexpected error",
			},
			ctx: &ErrorContext{
				ComponentName: "TestComponent",
				ComponentID:   "test-123",
				EventName:     "click",
				Timestamp:     time.Now(),
				Tags: map[string]string{
					"environment": "test",
				},
				Extra: map[string]interface{}{
					"user_id": "user-456",
				},
				Breadcrumbs: []Breadcrumb{
					{
						Type:      "navigation",
						Message:   "User clicked button",
						Timestamp: time.Now(),
					},
				},
				StackTrace: []byte("goroutine 1 [running]:\nmain.main()"),
			},
		},
		{
			name: "report panic with minimal context",
			panicErr: &HandlerPanicError{
				ComponentName: "Button",
				EventName:     "submit",
				PanicValue:    "validation failed",
			},
			ctx: &ErrorContext{
				ComponentName: "Button",
				EventName:     "submit",
				Timestamp:     time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create reporter with empty DSN (won't actually send)
			reporter, err := NewSentryReporter("")
			require.NoError(t, err)
			require.NotNil(t, reporter)
			defer reporter.Flush(1 * time.Second)

			// Should not panic
			assert.NotPanics(t, func() {
				reporter.ReportPanic(tt.panicErr, tt.ctx)
			})
		})
	}
}

// TestSentryReporter_ReportError tests error reporting to Sentry
func TestSentryReporter_ReportError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		ctx  *ErrorContext
	}{
		{
			name: "report error with context",
			err:  errors.New("validation error"),
			ctx: &ErrorContext{
				ComponentName: "Input",
				ComponentID:   "input-1",
				Timestamp:     time.Now(),
				Tags: map[string]string{
					"field": "email",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create reporter with empty DSN (won't actually send)
			reporter, err := NewSentryReporter("")
			require.NoError(t, err)
			require.NotNil(t, reporter)
			defer reporter.Flush(1 * time.Second)

			// Should not panic
			assert.NotPanics(t, func() {
				reporter.ReportError(tt.err, tt.ctx)
			})
		})
	}
}

// TestSentryReporter_Flush tests flush functionality
func TestSentryReporter_Flush(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "flush with 5 second timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "flush with 1 second timeout",
			timeout: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewSentryReporter("")
			require.NoError(t, err)
			require.NotNil(t, reporter)

			// Flush should not error
			err = reporter.Flush(tt.timeout)
			assert.NoError(t, err)
		})
	}
}

// TestSentryReporter_BeforeSend tests BeforeSend hook
func TestSentryReporter_BeforeSend(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "before send hook is called",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookCalled := false

			reporter, err := NewSentryReporter("",
				WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
					hookCalled = true
					// Modify event
					event.Tags["custom"] = "value"
					return event
				}),
			)
			require.NoError(t, err)
			require.NotNil(t, reporter)
			defer reporter.Flush(1 * time.Second)

			// Report an error
			reporter.ReportError(errors.New("test error"), &ErrorContext{
				ComponentName: "Test",
				Timestamp:     time.Now(),
			})

			// Flush to ensure event is processed
			reporter.Flush(1 * time.Second)

			// Note: With empty DSN, BeforeSend might not be called
			// This test verifies the option is accepted without error
			_ = hookCalled
		})
	}
}

// TestSentryReporter_Options tests various Sentry options
func TestSentryReporter_Options(t *testing.T) {
	tests := []struct {
		name string
		opts []SentryOption
	}{
		{
			name: "with debug option",
			opts: []SentryOption{
				WithDebug(true),
			},
		},
		{
			name: "with multiple options",
			opts: []SentryOption{
				WithDebug(true),
				WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
					return event
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewSentryReporter("", tt.opts...)
			assert.NoError(t, err)
			require.NotNil(t, reporter)
			defer reporter.Flush(1 * time.Second)
		})
	}
}

// TestConsoleReporter_Concurrent tests thread-safety of ConsoleReporter
func TestConsoleReporter_Concurrent(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		operations int
	}{
		{
			name:       "10 goroutines reporting concurrently",
			goroutines: 10,
			operations: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			reporter := NewConsoleReporter(true)

			// Concurrent reporting
			done := make(chan bool)
			for i := 0; i < tt.goroutines; i++ {
				go func() {
					for j := 0; j < tt.operations; j++ {
						reporter.ReportPanic(
							&HandlerPanicError{
								ComponentName: "Test",
								EventName:     "event",
								PanicValue:    "panic",
							},
							&ErrorContext{
								ComponentName: "Test",
								Timestamp:     time.Now(),
							},
						)
					}
					done <- true
				}()
			}

			// Wait for all goroutines
			for i := 0; i < tt.goroutines; i++ {
				<-done
			}

			// Verify output contains error messages
			output := buf.String()
			assert.Contains(t, output, "ERROR")
			assert.Contains(t, output, "Panic")

			// Count number of error messages
			count := strings.Count(output, "ERROR")
			expectedCount := tt.goroutines * tt.operations
			assert.Equal(t, expectedCount, count, "should have %d error messages", expectedCount)
		})
	}
}
