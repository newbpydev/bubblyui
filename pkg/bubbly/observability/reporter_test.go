package observability

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReporter is a test implementation of ErrorReporter
type mockReporter struct {
	panicCalls []mockPanicCall
	errorCalls []mockErrorCall
	flushCalls int
	flushError error
	mu         sync.Mutex
}

type mockPanicCall struct {
	err *HandlerPanicError
	ctx *ErrorContext
}

type mockErrorCall struct {
	err error
	ctx *ErrorContext
}

func (m *mockReporter) ReportPanic(err *HandlerPanicError, ctx *ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panicCalls = append(m.panicCalls, mockPanicCall{err: err, ctx: ctx})
}

func (m *mockReporter) ReportError(err error, ctx *ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalls = append(m.errorCalls, mockErrorCall{err: err, ctx: ctx})
}

func (m *mockReporter) Flush(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushCalls++
	return m.flushError
}

func (m *mockReporter) getPanicCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.panicCalls)
}

func (m *mockReporter) getErrorCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.errorCalls)
}

func (m *mockReporter) getFlushCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.flushCalls
}

// TestErrorReporter_Interface verifies the ErrorReporter interface is defined correctly
func TestErrorReporter_Interface(t *testing.T) {
	tests := []struct {
		name     string
		reporter ErrorReporter
	}{
		{
			name:     "mock reporter implements interface",
			reporter: &mockReporter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify interface methods exist and can be called
			require.NotNil(t, tt.reporter)

			// Test ReportPanic
			panicErr := &HandlerPanicError{
				ComponentName: "TestComponent",
				EventName:     "testEvent",
				PanicValue:    "test panic",
			}
			ctx := &ErrorContext{
				ComponentName: "TestComponent",
				ComponentID:   "test-id",
				EventName:     "testEvent",
				Timestamp:     time.Now(),
			}
			tt.reporter.ReportPanic(panicErr, ctx)

			// Test ReportError
			tt.reporter.ReportError(assert.AnError, ctx)

			// Test Flush
			err := tt.reporter.Flush(5 * time.Second)
			assert.NoError(t, err)
		})
	}
}

// TestSetErrorReporter tests setting the global error reporter
func TestSetErrorReporter(t *testing.T) {
	tests := []struct {
		name     string
		reporter ErrorReporter
		wantNil  bool
	}{
		{
			name:     "set non-nil reporter",
			reporter: &mockReporter{},
			wantNil:  false,
		},
		{
			name:     "set nil reporter",
			reporter: nil,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set reporter
			SetErrorReporter(tt.reporter)

			// Verify it was set
			got := GetErrorReporter()
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.reporter, got)
			}

			// Cleanup
			SetErrorReporter(nil)
		})
	}
}

// TestGetErrorReporter tests retrieving the global error reporter
func TestGetErrorReporter(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantNil bool
	}{
		{
			name: "get when reporter is set",
			setup: func() {
				SetErrorReporter(&mockReporter{})
			},
			wantNil: false,
		},
		{
			name: "get when reporter is nil",
			setup: func() {
				SetErrorReporter(nil)
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setup()

			// Get reporter
			got := GetErrorReporter()

			// Verify
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
			}

			// Cleanup
			SetErrorReporter(nil)
		})
	}
}

// TestErrorReporter_NilHandling tests that nil reporter is handled gracefully
func TestErrorReporter_NilHandling(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "nil reporter does not panic on get",
			test: func(t *testing.T) {
				SetErrorReporter(nil)
				assert.NotPanics(t, func() {
					reporter := GetErrorReporter()
					assert.Nil(t, reporter)
				})
			},
		},
		{
			name: "setting nil reporter multiple times is safe",
			test: func(t *testing.T) {
				assert.NotPanics(t, func() {
					SetErrorReporter(nil)
					SetErrorReporter(nil)
					SetErrorReporter(nil)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
			// Cleanup
			SetErrorReporter(nil)
		})
	}
}

// TestErrorReporter_Concurrent tests thread-safety of global reporter management
func TestErrorReporter_Concurrent(t *testing.T) {
	tests := []struct {
		name       string
		goroutines int
		operations int
	}{
		{
			name:       "10 goroutines, 100 operations each",
			goroutines: 10,
			operations: 100,
		},
		{
			name:       "50 goroutines, 50 operations each",
			goroutines: 50,
			operations: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			reporter := &mockReporter{}

			// Concurrent set/get operations
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.operations; j++ {
						SetErrorReporter(reporter)
						got := GetErrorReporter()
						assert.NotNil(t, got)
					}
				}()
			}

			wg.Wait()

			// Verify final state
			got := GetErrorReporter()
			assert.NotNil(t, got)

			// Cleanup
			SetErrorReporter(nil)
		})
	}
}

// TestErrorContext_Fields verifies ErrorContext has all required fields
func TestErrorContext_Fields(t *testing.T) {
	tests := []struct {
		name string
		ctx  ErrorContext
	}{
		{
			name: "all fields present",
			ctx: ErrorContext{
				ComponentName: "TestComponent",
				ComponentID:   "test-123",
				EventName:     "click",
				Timestamp:     time.Now(),
				Tags:          map[string]string{"env": "test"},
				Extra:         map[string]interface{}{"key": "value"},
				Breadcrumbs:   []Breadcrumb{{Type: "navigation"}},
				StackTrace:    []byte("stack trace"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible
			assert.Equal(t, "TestComponent", tt.ctx.ComponentName)
			assert.Equal(t, "test-123", tt.ctx.ComponentID)
			assert.Equal(t, "click", tt.ctx.EventName)
			assert.NotZero(t, tt.ctx.Timestamp)
			assert.NotNil(t, tt.ctx.Tags)
			assert.NotNil(t, tt.ctx.Extra)
			assert.NotNil(t, tt.ctx.Breadcrumbs)
			assert.NotNil(t, tt.ctx.StackTrace)
		})
	}
}

// TestBreadcrumb_Fields verifies Breadcrumb has all required fields
func TestBreadcrumb_Fields(t *testing.T) {
	tests := []struct {
		name       string
		breadcrumb Breadcrumb
	}{
		{
			name: "all fields present",
			breadcrumb: Breadcrumb{
				Type:      "navigation",
				Category:  "ui",
				Message:   "User clicked button",
				Level:     "info",
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"button": "submit"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible
			assert.Equal(t, "navigation", tt.breadcrumb.Type)
			assert.Equal(t, "ui", tt.breadcrumb.Category)
			assert.Equal(t, "User clicked button", tt.breadcrumb.Message)
			assert.Equal(t, "info", tt.breadcrumb.Level)
			assert.NotZero(t, tt.breadcrumb.Timestamp)
			assert.NotNil(t, tt.breadcrumb.Data)
		})
	}
}

// TestErrorReporter_ReportPanic tests ReportPanic functionality
func TestErrorReporter_ReportPanic(t *testing.T) {
	tests := []struct {
		name      string
		panicErr  *HandlerPanicError
		ctx       *ErrorContext
		wantCalls int
	}{
		{
			name: "report single panic",
			panicErr: &HandlerPanicError{
				ComponentName: "Button",
				EventName:     "click",
				PanicValue:    "unexpected error",
			},
			ctx: &ErrorContext{
				ComponentName: "Button",
				ComponentID:   "btn-1",
				EventName:     "click",
				Timestamp:     time.Now(),
			},
			wantCalls: 1,
		},
		{
			name: "report panic with stack trace",
			panicErr: &HandlerPanicError{
				ComponentName: "Form",
				EventName:     "submit",
				PanicValue:    "validation failed",
			},
			ctx: &ErrorContext{
				ComponentName: "Form",
				ComponentID:   "form-1",
				EventName:     "submit",
				Timestamp:     time.Now(),
				StackTrace:    []byte("goroutine 1 [running]:\n..."),
			},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &mockReporter{}
			SetErrorReporter(reporter)
			defer SetErrorReporter(nil)

			// Report panic
			reporter.ReportPanic(tt.panicErr, tt.ctx)

			// Verify
			assert.Equal(t, tt.wantCalls, reporter.getPanicCallCount())
		})
	}
}

// TestErrorReporter_ReportError tests ReportError functionality
func TestErrorReporter_ReportError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		ctx       *ErrorContext
		wantCalls int
	}{
		{
			name: "report single error",
			err:  assert.AnError,
			ctx: &ErrorContext{
				ComponentName: "Input",
				ComponentID:   "input-1",
				EventName:     "change",
				Timestamp:     time.Now(),
			},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &mockReporter{}
			SetErrorReporter(reporter)
			defer SetErrorReporter(nil)

			// Report error
			reporter.ReportError(tt.err, tt.ctx)

			// Verify
			assert.Equal(t, tt.wantCalls, reporter.getErrorCallCount())
		})
	}
}

// TestErrorReporter_Flush tests Flush functionality
func TestErrorReporter_Flush(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		wantCalls int
		wantError bool
	}{
		{
			name:      "flush with 5 second timeout",
			timeout:   5 * time.Second,
			wantCalls: 1,
			wantError: false,
		},
		{
			name:      "flush with 1 second timeout",
			timeout:   1 * time.Second,
			wantCalls: 1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &mockReporter{}
			SetErrorReporter(reporter)
			defer SetErrorReporter(nil)

			// Flush
			err := reporter.Flush(tt.timeout)

			// Verify
			assert.Equal(t, tt.wantCalls, reporter.getFlushCallCount())
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHandlerPanicError_Error tests the Error method of HandlerPanicError
func TestHandlerPanicError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *HandlerPanicError
		wantParts []string
	}{
		{
			name: "error message contains all fields",
			err: &HandlerPanicError{
				ComponentName: "TestButton",
				EventName:     "click",
				PanicValue:    "unexpected nil pointer",
			},
			wantParts: []string{
				"panic in event handler",
				"TestButton",
				"click",
				"unexpected nil pointer",
			},
		},
		{
			name: "error message with different values",
			err: &HandlerPanicError{
				ComponentName: "LoginForm",
				EventName:     "submit",
				PanicValue:    123,
			},
			wantParts: []string{
				"panic in event handler",
				"LoginForm",
				"submit",
				"123",
			},
		},
		{
			name: "error message with empty fields",
			err: &HandlerPanicError{
				ComponentName: "",
				EventName:     "",
				PanicValue:    nil,
			},
			wantParts: []string{
				"panic in event handler",
				"component ''",
				"event ''",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()

			for _, part := range tt.wantParts {
				assert.Contains(t, errMsg, part, "error message should contain %q", part)
			}
		})
	}
}

// TestCommandGenerationError_Error tests the Error method of CommandGenerationError
func TestCommandGenerationError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *CommandGenerationError
		wantParts []string
	}{
		{
			name: "error message contains all fields",
			err: &CommandGenerationError{
				ComponentID: "component-123",
				RefID:       "ref-456",
				PanicValue:  "failed to generate command",
			},
			wantParts: []string{
				"panic in command generation",
				"component-123",
				"ref-456",
				"failed to generate command",
			},
		},
		{
			name: "error message with integer panic value",
			err: &CommandGenerationError{
				ComponentID: "form-1",
				RefID:       "counter",
				PanicValue:  42,
			},
			wantParts: []string{
				"panic in command generation",
				"form-1",
				"counter",
				"42",
			},
		},
		{
			name: "error message with empty fields",
			err: &CommandGenerationError{
				ComponentID: "",
				RefID:       "",
				PanicValue:  nil,
			},
			wantParts: []string{
				"panic in command generation",
				"component ''",
				"ref ''",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()

			for _, part := range tt.wantParts {
				assert.Contains(t, errMsg, part, "error message should contain %q", part)
			}
		})
	}
}
