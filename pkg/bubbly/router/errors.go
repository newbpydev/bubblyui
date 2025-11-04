package router

import (
	"fmt"
)

// ErrorCode categorizes router errors for better error handling.
//
// Error codes help applications distinguish between different types
// of navigation failures and respond appropriately.
//
// Example:
//
//	switch err.Code {
//	case ErrCodeRouteNotFound:
//		// Show 404 page
//	case ErrCodeGuardRejected:
//		// Show access denied message
//	case ErrCodeCircularRedirect:
//		// Log error, show error page
//	}
type ErrorCode int

const (
	// ErrCodeRouteNotFound indicates no route matched the requested path.
	// This typically results in a 404-style error page.
	ErrCodeRouteNotFound ErrorCode = iota

	// ErrCodeInvalidPath indicates the path format is invalid.
	// Examples: empty path, malformed URL, invalid characters.
	ErrCodeInvalidPath

	// ErrCodeGuardRejected indicates a navigation guard rejected the navigation.
	// This happens when a guard calls next() with an empty target or returns an error.
	ErrCodeGuardRejected

	// ErrCodeCircularRedirect indicates guards are redirecting in a loop.
	// The router detects this after a configurable number of redirects (default: 10).
	ErrCodeCircularRedirect

	// ErrCodeComponentNotFound indicates the route's component is nil or invalid.
	// This is a configuration error that should be caught during route registration.
	ErrCodeComponentNotFound

	// ErrCodeInvalidTarget indicates the navigation target is invalid.
	// Examples: nil target, target with neither path nor name.
	ErrCodeInvalidTarget
)

// String returns a human-readable name for the error code.
func (e ErrorCode) String() string {
	switch e {
	case ErrCodeRouteNotFound:
		return "RouteNotFound"
	case ErrCodeInvalidPath:
		return "InvalidPath"
	case ErrCodeGuardRejected:
		return "GuardRejected"
	case ErrCodeCircularRedirect:
		return "CircularRedirect"
	case ErrCodeComponentNotFound:
		return "ComponentNotFound"
	case ErrCodeInvalidTarget:
		return "InvalidTarget"
	default:
		return fmt.Sprintf("UnknownError(%d)", e)
	}
}

// RouterError represents a router-specific error with rich context.
//
// RouterError provides structured error information including:
//   - Error code for categorization
//   - Human-readable message
//   - Navigation context (from/to routes)
//   - Underlying cause (if any)
//
// This allows applications to handle different error types appropriately
// and provides detailed information for debugging and error reporting.
//
// Example:
//
//	err := &RouterError{
//		Code:    ErrRouteNotFound,
//		Message: "No route matches '/invalid'",
//		From:    currentRoute,
//		To:      &NavigationTarget{Path: "/invalid"},
//		Cause:   ErrNoMatch,
//	}
//
//	// Check error type
//	if routerErr, ok := err.(*RouterError); ok {
//		switch routerErr.Code {
//		case ErrRouteNotFound:
//			// Handle 404
//		}
//	}
type RouterError struct {
	// Code categorizes the error type
	Code ErrorCode

	// Message is a human-readable error description
	Message string

	// From is the route we were navigating from (nil if no current route)
	From *Route

	// To is the navigation target that caused the error
	To *NavigationTarget

	// Cause is the underlying error that caused this router error (optional)
	Cause error
}

// Error implements the error interface.
//
// Returns a formatted error message including the error code,
// message, and navigation context.
//
// Format: "[ErrorCode] message (from: /path, to: /target)"
//
// Example output:
//   - "[RouteNotFound] No route matches '/invalid' (from: /home, to: /invalid)"
//   - "[GuardRejected] Authentication required (from: /public, to: /admin)"
func (e *RouterError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)

	// Add navigation context if available
	if e.From != nil || e.To != nil {
		msg += " ("

		if e.From != nil {
			msg += fmt.Sprintf("from: %s", e.From.Path)
		}

		if e.To != nil {
			if e.From != nil {
				msg += ", "
			}
			if e.To.Path != "" {
				msg += fmt.Sprintf("to: %s", e.To.Path)
			} else if e.To.Name != "" {
				msg += fmt.Sprintf("to: %s (by name)", e.To.Name)
			}
		}

		msg += ")"
	}

	// Add cause if present
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}

	return msg
}

// Unwrap returns the underlying cause error.
//
// This implements the Go 1.13+ error unwrapping interface,
// allowing errors.Is() and errors.As() to work correctly.
//
// Example:
//
//	err := &RouterError{
//		Code:  ErrRouteNotFound,
//		Cause: ErrNoMatch,
//	}
//
//	if errors.Is(err, ErrNoMatch) {
//		// This works because Unwrap() returns ErrNoMatch
//	}
func (e *RouterError) Unwrap() error {
	return e.Cause
}

// NewRouteNotFoundError creates a RouterError for route not found.
//
// This is a convenience constructor for the common case of a route
// not being found during navigation.
//
// Parameters:
//   - path: The path that was not found
//   - from: The route we were navigating from (nil if none)
//   - cause: The underlying error (typically ErrNoMatch)
//
// Returns:
//   - *RouterError: A RouterError with ErrRouteNotFound code
//
// Example:
//
//	err := NewRouteNotFoundError("/invalid", currentRoute, ErrNoMatch)
//	// Error: "[RouteNotFound] No route matches '/invalid' (from: /home, to: /invalid)"
func NewRouteNotFoundError(path string, from *Route, cause error) *RouterError {
	return &RouterError{
		Code:    ErrCodeRouteNotFound,
		Message: fmt.Sprintf("No route matches '%s'", path),
		From:    from,
		To:      &NavigationTarget{Path: path},
		Cause:   cause,
	}
}

// NewInvalidTargetError creates a RouterError for invalid navigation target.
//
// This is used when the navigation target is nil or has neither path nor name.
//
// Parameters:
//   - message: Description of why the target is invalid
//   - target: The invalid navigation target (may be nil)
//   - from: The route we were navigating from (nil if none)
//
// Returns:
//   - *RouterError: A RouterError with ErrInvalidTarget code
//
// Example:
//
//	err := NewInvalidTargetError("target cannot be nil", nil, currentRoute)
//	// Error: "[InvalidTarget] target cannot be nil (from: /home)"
func NewInvalidTargetError(message string, target *NavigationTarget, from *Route) *RouterError {
	return &RouterError{
		Code:    ErrCodeInvalidTarget,
		Message: message,
		From:    from,
		To:      target,
	}
}

// NewGuardRejectedError creates a RouterError for guard rejection.
//
// This is used when a navigation guard rejects navigation by calling
// next() with an empty target or returning an error.
//
// Parameters:
//   - guardName: Name of the guard that rejected navigation
//   - from: The route we were navigating from
//   - to: The target we tried to navigate to
//   - cause: The underlying error from the guard (optional)
//
// Returns:
//   - *RouterError: A RouterError with ErrGuardRejected code
//
// Example:
//
//	err := NewGuardRejectedError("authGuard", currentRoute, target, nil)
//	// Error: "[GuardRejected] Navigation rejected by guard 'authGuard' (from: /public, to: /admin)"
func NewGuardRejectedError(guardName string, from *Route, to *NavigationTarget, cause error) *RouterError {
	return &RouterError{
		Code:    ErrCodeGuardRejected,
		Message: fmt.Sprintf("Navigation rejected by guard '%s'", guardName),
		From:    from,
		To:      to,
		Cause:   cause,
	}
}

// NewCircularRedirectError creates a RouterError for circular redirects.
//
// This is used when guards redirect in a loop, detected after a
// configurable number of redirects (default: 10).
//
// Parameters:
//   - redirectCount: Number of redirects that occurred
//   - path: The path where the loop was detected
//   - from: The route we started from
//
// Returns:
//   - *RouterError: A RouterError with ErrCircularRedirect code
//
// Example:
//
//	err := NewCircularRedirectError(10, "/login", startRoute)
//	// Error: "[CircularRedirect] Maximum redirects (10) exceeded at '/login' (from: /start, to: /login)"
func NewCircularRedirectError(redirectCount int, path string, from *Route) *RouterError {
	return &RouterError{
		Code:    ErrCodeCircularRedirect,
		Message: fmt.Sprintf("Maximum redirects (%d) exceeded at '%s'", redirectCount, path),
		From:    from,
		To:      &NavigationTarget{Path: path},
	}
}
