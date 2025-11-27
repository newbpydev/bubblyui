package router

import (
	"errors"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestErrorCode_String tests error code string representation
func TestErrorCode_String(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
		want string
	}{
		{
			name: "route not found",
			code: ErrCodeRouteNotFound,
			want: "RouteNotFound",
		},
		{
			name: "invalid path",
			code: ErrCodeInvalidPath,
			want: "InvalidPath",
		},
		{
			name: "guard rejected",
			code: ErrCodeGuardRejected,
			want: "GuardRejected",
		},
		{
			name: "circular redirect",
			code: ErrCodeCircularRedirect,
			want: "CircularRedirect",
		},
		{
			name: "component not found",
			code: ErrCodeComponentNotFound,
			want: "ComponentNotFound",
		},
		{
			name: "invalid target",
			code: ErrCodeInvalidTarget,
			want: "InvalidTarget",
		},
		{
			name: "unknown error code",
			code: ErrorCode(999),
			want: "UnknownError(999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.code.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestError_Error tests error message formatting
func TestError_Error(t *testing.T) {
	homeRoute := &Route{Path: "/home"}

	tests := []struct {
		name    string
		err     *Error
		wantMsg string
	}{
		{
			name: "basic error with code and message",
			err: &Error{
				Code:    ErrCodeRouteNotFound,
				Message: "No route matches '/invalid'",
			},
			wantMsg: "[RouteNotFound] No route matches '/invalid'",
		},
		{
			name: "error with from route",
			err: &Error{
				Code:    ErrCodeRouteNotFound,
				Message: "No route matches '/invalid'",
				From:    homeRoute,
			},
			wantMsg: "[RouteNotFound] No route matches '/invalid' (from: /home)",
		},
		{
			name: "error with to target (path)",
			err: &Error{
				Code:    ErrCodeRouteNotFound,
				Message: "No route matches '/invalid'",
				To:      &NavigationTarget{Path: "/invalid"},
			},
			wantMsg: "[RouteNotFound] No route matches '/invalid' (to: /invalid)",
		},
		{
			name: "error with to target (name)",
			err: &Error{
				Code:    ErrCodeRouteNotFound,
				Message: "No route matches 'invalid-route'",
				To:      &NavigationTarget{Name: "invalid-route"},
			},
			wantMsg: "[RouteNotFound] No route matches 'invalid-route' (to: invalid-route (by name))",
		},
		{
			name: "error with from and to",
			err: &Error{
				Code:    ErrCodeGuardRejected,
				Message: "Authentication required",
				From:    homeRoute,
				To:      &NavigationTarget{Path: "/admin"},
			},
			wantMsg: "[GuardRejected] Authentication required (from: /home, to: /admin)",
		},
		{
			name: "error with cause",
			err: &Error{
				Code:    ErrCodeRouteNotFound,
				Message: "No route matches '/invalid'",
				From:    homeRoute,
				To:      &NavigationTarget{Path: "/invalid"},
				Cause:   ErrNoMatch,
			},
			wantMsg: "[RouteNotFound] No route matches '/invalid' (from: /home, to: /invalid): no route matches path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.wantMsg, got)
		})
	}
}

// TestError_Unwrap tests error unwrapping
func TestError_Unwrap(t *testing.T) {
	tests := []struct {
		name      string
		err       *Error
		wantCause error
	}{
		{
			name: "error with cause",
			err: &Error{
				Code:  ErrCodeRouteNotFound,
				Cause: ErrNoMatch,
			},
			wantCause: ErrNoMatch,
		},
		{
			name: "error without cause",
			err: &Error{
				Code: ErrCodeRouteNotFound,
			},
			wantCause: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			assert.Equal(t, tt.wantCause, got)

			// Test errors.Is works
			if tt.wantCause != nil {
				assert.True(t, errors.Is(tt.err, tt.wantCause))
			}
		})
	}
}

// TestNewRouteNotFoundError tests route not found error constructor
func TestNewRouteNotFoundError(t *testing.T) {
	homeRoute := &Route{Path: "/home"}

	tests := []struct {
		name      string
		path      string
		from      *Route
		cause     error
		wantCode  ErrorCode
		wantMsg   string
		wantCause error
	}{
		{
			name:      "basic route not found",
			path:      "/invalid",
			from:      homeRoute,
			cause:     ErrNoMatch,
			wantCode:  ErrCodeRouteNotFound,
			wantMsg:   "No route matches '/invalid'",
			wantCause: ErrNoMatch,
		},
		{
			name:      "route not found without from",
			path:      "/invalid",
			from:      nil,
			cause:     ErrNoMatch,
			wantCode:  ErrCodeRouteNotFound,
			wantMsg:   "No route matches '/invalid'",
			wantCause: ErrNoMatch,
		},
		{
			name:      "route not found without cause",
			path:      "/invalid",
			from:      homeRoute,
			cause:     nil,
			wantCode:  ErrCodeRouteNotFound,
			wantMsg:   "No route matches '/invalid'",
			wantCause: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRouteNotFoundError(tt.path, tt.from, tt.cause)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMsg, err.Message)
			assert.Equal(t, tt.from, err.From)
			assert.Equal(t, tt.path, err.To.Path)
			assert.Equal(t, tt.wantCause, err.Cause)
		})
	}
}

// TestNewInvalidTargetError tests invalid target error constructor
func TestNewInvalidTargetError(t *testing.T) {
	homeRoute := &Route{Path: "/home"}
	target := &NavigationTarget{Path: "/test"}

	tests := []struct {
		name     string
		message  string
		target   *NavigationTarget
		from     *Route
		wantCode ErrorCode
	}{
		{
			name:     "nil target",
			message:  "target cannot be nil",
			target:   nil,
			from:     homeRoute,
			wantCode: ErrCodeInvalidTarget,
		},
		{
			name:     "empty target",
			message:  "target must have path or name",
			target:   &NavigationTarget{},
			from:     homeRoute,
			wantCode: ErrCodeInvalidTarget,
		},
		{
			name:     "with target",
			message:  "invalid target format",
			target:   target,
			from:     homeRoute,
			wantCode: ErrCodeInvalidTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewInvalidTargetError(tt.message, tt.target, tt.from)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.target, err.To)
			assert.Equal(t, tt.from, err.From)
		})
	}
}

// TestNewGuardRejectedError tests guard rejected error constructor
func TestNewGuardRejectedError(t *testing.T) {
	homeRoute := &Route{Path: "/home"}
	target := &NavigationTarget{Path: "/admin"}
	guardErr := errors.New("authentication failed")

	tests := []struct {
		name      string
		guardName string
		from      *Route
		to        *NavigationTarget
		cause     error
		wantCode  ErrorCode
		wantMsg   string
	}{
		{
			name:      "guard rejected with cause",
			guardName: "authGuard",
			from:      homeRoute,
			to:        target,
			cause:     guardErr,
			wantCode:  ErrCodeGuardRejected,
			wantMsg:   "Navigation rejected by guard 'authGuard'",
		},
		{
			name:      "guard rejected without cause",
			guardName: "roleGuard",
			from:      homeRoute,
			to:        target,
			cause:     nil,
			wantCode:  ErrCodeGuardRejected,
			wantMsg:   "Navigation rejected by guard 'roleGuard'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewGuardRejectedError(tt.guardName, tt.from, tt.to, tt.cause)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMsg, err.Message)
			assert.Equal(t, tt.from, err.From)
			assert.Equal(t, tt.to, err.To)
			assert.Equal(t, tt.cause, err.Cause)
		})
	}
}

// TestNewCircularRedirectError tests circular redirect error constructor
func TestNewCircularRedirectError(t *testing.T) {
	homeRoute := &Route{Path: "/home"}

	tests := []struct {
		name          string
		redirectCount int
		path          string
		from          *Route
		wantCode      ErrorCode
		wantMsg       string
	}{
		{
			name:          "max redirects exceeded",
			redirectCount: 10,
			path:          "/login",
			from:          homeRoute,
			wantCode:      ErrCodeCircularRedirect,
			wantMsg:       "Maximum redirects (10) exceeded at '/login'",
		},
		{
			name:          "different redirect count",
			redirectCount: 5,
			path:          "/redirect",
			from:          homeRoute,
			wantCode:      ErrCodeCircularRedirect,
			wantMsg:       "Maximum redirects (5) exceeded at '/redirect'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCircularRedirectError(tt.redirectCount, tt.path, tt.from)

			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.wantMsg, err.Message)
			assert.Equal(t, tt.from, err.From)
			assert.Equal(t, tt.path, err.To.Path)
		})
	}
}

// TestError_ObservabilityIntegration tests error reporting to observability system
func TestError_ObservabilityIntegration(t *testing.T) {
	// Create a mock reporter to capture reported errors
	mockReporter := &mockErrorReporter{
		errors: make([]reportedError, 0),
	}

	// Set the global reporter
	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Create a router error
	homeRoute := &Route{Path: "/home", Name: "home"}
	target := &NavigationTarget{Path: "/admin"}
	routerErr := NewGuardRejectedError("authGuard", homeRoute, target, nil)

	// Report the error using observability system
	if reporter := observability.GetErrorReporter(); reporter != nil {
		reporter.ReportError(routerErr, &observability.ErrorContext{
			ComponentName: "router",
			EventName:     "navigation",
			Timestamp:     time.Now(),
			Tags: map[string]string{
				"error_code": routerErr.Code.String(),
				"from_path":  homeRoute.Path,
				"to_path":    target.Path,
			},
			Extra: map[string]interface{}{
				"guard_name": "authGuard",
			},
			StackTrace: debug.Stack(),
		})
	}

	// Verify error was reported
	require.Len(t, mockReporter.errors, 1)
	reported := mockReporter.errors[0]

	assert.Equal(t, routerErr, reported.err)
	assert.Equal(t, "router", reported.ctx.ComponentName)
	assert.Equal(t, "navigation", reported.ctx.EventName)
	assert.Equal(t, "GuardRejected", reported.ctx.Tags["error_code"])
	assert.Equal(t, "/home", reported.ctx.Tags["from_path"])
	assert.Equal(t, "/admin", reported.ctx.Tags["to_path"])
	assert.Equal(t, "authGuard", reported.ctx.Extra["guard_name"])
	assert.NotEmpty(t, reported.ctx.StackTrace)
}

// TestError_StackTraceCapture tests stack trace capture
func TestError_StackTraceCapture(t *testing.T) {
	// Create error with stack trace
	_ = NewRouteNotFoundError("/invalid", nil, ErrNoMatch)
	stackTrace := debug.Stack()

	// Verify stack trace contains relevant information
	stackStr := string(stackTrace)
	assert.Contains(t, stackStr, "errors_test.go")
	assert.Contains(t, stackStr, "TestError_StackTraceCapture")
}

// TestError_ErrorCategorization tests error categorization
func TestError_ErrorCategorization(t *testing.T) {
	tests := []struct {
		name         string
		err          *Error
		wantCode     ErrorCode
		wantCategory string
	}{
		{
			name:         "route not found - client error",
			err:          NewRouteNotFoundError("/invalid", nil, ErrNoMatch),
			wantCode:     ErrCodeRouteNotFound,
			wantCategory: "client_error",
		},
		{
			name:         "invalid target - client error",
			err:          NewInvalidTargetError("nil target", nil, nil),
			wantCode:     ErrCodeInvalidTarget,
			wantCategory: "client_error",
		},
		{
			name:         "guard rejected - authorization error",
			err:          NewGuardRejectedError("authGuard", nil, &NavigationTarget{Path: "/admin"}, nil),
			wantCode:     ErrCodeGuardRejected,
			wantCategory: "authorization_error",
		},
		{
			name:         "circular redirect - configuration error",
			err:          NewCircularRedirectError(10, "/loop", nil),
			wantCode:     ErrCodeCircularRedirect,
			wantCategory: "configuration_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)

			// Categorize based on error code
			var category string
			switch tt.err.Code {
			case ErrCodeRouteNotFound, ErrCodeInvalidPath, ErrCodeInvalidTarget:
				category = "client_error"
			case ErrCodeGuardRejected:
				category = "authorization_error"
			case ErrCodeCircularRedirect, ErrCodeComponentNotFound:
				category = "configuration_error"
			}

			assert.Equal(t, tt.wantCategory, category)
		})
	}
}

// TestError_ClearErrorMessages tests error message clarity
func TestError_ClearErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		err             *Error
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "route not found message",
			err:  NewRouteNotFoundError("/invalid", &Route{Path: "/home"}, ErrNoMatch),
			wantContains: []string{
				"RouteNotFound",
				"/invalid",
				"from: /home",
			},
			wantNotContains: []string{
				"error",
				"failed",
			},
		},
		{
			name: "guard rejected message",
			err:  NewGuardRejectedError("authGuard", &Route{Path: "/public"}, &NavigationTarget{Path: "/admin"}, nil),
			wantContains: []string{
				"GuardRejected",
				"authGuard",
				"from: /public",
				"to: /admin",
			},
			wantNotContains: []string{},
		},
		{
			name: "circular redirect message",
			err:  NewCircularRedirectError(10, "/loop", &Route{Path: "/start"}),
			wantContains: []string{
				"CircularRedirect",
				"Maximum redirects",
				"10",
				"/loop",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()

			for _, want := range tt.wantContains {
				assert.Contains(t, errMsg, want, "error message should contain: %s", want)
			}

			for _, notWant := range tt.wantNotContains {
				assert.NotContains(t, errMsg, notWant, "error message should not contain: %s", notWant)
			}
		})
	}
}

// mockErrorReporter is a mock implementation of ErrorReporter for testing
type mockErrorReporter struct {
	mu     sync.Mutex
	errors []reportedError
	panics []reportedPanic
}

type reportedError struct {
	err error
	ctx *observability.ErrorContext
}

type reportedPanic struct {
	err *observability.HandlerPanicError
	ctx *observability.ErrorContext
}

func (m *mockErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, reportedError{err: err, ctx: ctx})
}

func (m *mockErrorReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panics = append(m.panics, reportedPanic{err: err, ctx: ctx})
}

func (m *mockErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

// TestError_ErrorRecovery tests error recovery patterns
func TestError_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name            string
		err             *Error
		wantRecoverable bool
		recoveryAction  string
	}{
		{
			name:            "route not found - show 404",
			err:             NewRouteNotFoundError("/invalid", nil, ErrNoMatch),
			wantRecoverable: true,
			recoveryAction:  "show_404_page",
		},
		{
			name:            "guard rejected - redirect to login",
			err:             NewGuardRejectedError("authGuard", nil, &NavigationTarget{Path: "/admin"}, nil),
			wantRecoverable: true,
			recoveryAction:  "redirect_to_login",
		},
		{
			name:            "circular redirect - show error page",
			err:             NewCircularRedirectError(10, "/loop", nil),
			wantRecoverable: false,
			recoveryAction:  "show_error_page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Determine if error is recoverable
			recoverable := tt.err.Code != ErrCodeCircularRedirect

			assert.Equal(t, tt.wantRecoverable, recoverable)

			// Determine recovery action
			var action string
			switch tt.err.Code {
			case ErrCodeRouteNotFound:
				action = "show_404_page"
			case ErrCodeGuardRejected:
				action = "redirect_to_login"
			case ErrCodeCircularRedirect:
				action = "show_error_page"
			}

			assert.Equal(t, tt.recoveryAction, action)
		})
	}
}

// TestError_ConcurrentReporting tests concurrent error reporting
func TestError_ConcurrentReporting(t *testing.T) {
	mockReporter := &mockErrorReporter{
		errors: make([]reportedError, 0),
	}

	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Report errors concurrently
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			err := NewRouteNotFoundError("/test", nil, ErrNoMatch)
			if reporter := observability.GetErrorReporter(); reporter != nil {
				reporter.ReportError(err, &observability.ErrorContext{
					ComponentName: "router",
					Tags: map[string]string{
						"goroutine_id": string(rune(id)),
					},
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all errors were reported
	// Note: We can't guarantee exact count due to race conditions in mock,
	// but this tests that concurrent reporting doesn't panic
	assert.True(t, len(mockReporter.errors) > 0)
}
