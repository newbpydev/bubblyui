package testutil

import (
	"fmt"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// MockErrorReporter is a mock implementation of observability.ErrorReporter for testing.
// It captures all reported errors, panics, and contexts for later inspection and assertion.
//
// Thread-safe: All methods are safe for concurrent use.
//
// Example usage:
//
//	reporter := NewMockErrorReporter()
//	observability.SetErrorReporter(reporter)
//	defer observability.SetErrorReporter(nil)
//
//	// Run code that may report errors
//	component.Emit("event", data)
//
//	// Assert errors were reported
//	reporter.AssertErrorReported(t, expectedErr)
//	assert.Len(t, reporter.GetErrors(), 1)
type MockErrorReporter struct {
	errors   []error
	panics   []*observability.HandlerPanicError
	contexts []*observability.ErrorContext
	mu       sync.RWMutex
}

// NewMockErrorReporter creates a new mock error reporter.
// The reporter starts with empty error, panic, and context slices.
//
// Returns:
//   - *MockErrorReporter: A new mock reporter ready for use
//
// Example:
//
//	reporter := NewMockErrorReporter()
//	observability.SetErrorReporter(reporter)
func NewMockErrorReporter() *MockErrorReporter {
	return &MockErrorReporter{
		errors:   make([]error, 0),
		panics:   make([]*observability.HandlerPanicError, 0),
		contexts: make([]*observability.ErrorContext, 0),
	}
}

// ReportError records an error and its context.
// This method is called by the component system when errors occur.
//
// Parameters:
//   - err: The error to report
//   - ctx: Rich context about where and when the error occurred (can be nil)
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors = append(m.errors, err)
	m.contexts = append(m.contexts, ctx)
}

// ReportPanic records a panic and its context.
// This method is called by the component system when event handlers panic.
//
// Parameters:
//   - err: The HandlerPanicError containing panic details
//   - ctx: Rich context about where and when the panic occurred (can be nil)
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.panics = append(m.panics, err)
	m.contexts = append(m.contexts, ctx)
}

// Flush is a no-op for the mock reporter.
// Always returns nil since there's nothing to flush in a mock.
//
// Parameters:
//   - timeout: Ignored for mock reporter
//
// Returns:
//   - error: Always nil
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

// GetErrors returns a copy of all reported errors.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - []error: Copy of all reported errors
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) GetErrors() []error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]error, len(m.errors))
	copy(result, m.errors)
	return result
}

// GetPanics returns a copy of all reported panics.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - []*observability.HandlerPanicError: Copy of all reported panics
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) GetPanics() []*observability.HandlerPanicError {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*observability.HandlerPanicError, len(m.panics))
	copy(result, m.panics)
	return result
}

// GetContexts returns a copy of all error contexts.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - []*observability.ErrorContext: Copy of all error contexts
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) GetContexts() []*observability.ErrorContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*observability.ErrorContext, len(m.contexts))
	copy(result, m.contexts)
	return result
}

// GetBreadcrumbs returns all breadcrumbs from all error contexts.
// Breadcrumbs are collected from all contexts in chronological order.
//
// Returns:
//   - []observability.Breadcrumb: All breadcrumbs from all contexts
//
// Thread-safe: Safe to call concurrently.
func (m *MockErrorReporter) GetBreadcrumbs() []observability.Breadcrumb {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var breadcrumbs []observability.Breadcrumb
	for _, ctx := range m.contexts {
		if ctx != nil && len(ctx.Breadcrumbs) > 0 {
			breadcrumbs = append(breadcrumbs, ctx.Breadcrumbs...)
		}
	}
	return breadcrumbs
}

// AssertErrorReported asserts that a specific error was reported.
// Fails the test if the error was not found in the reported errors.
//
// Parameters:
//   - t: The testing.T instance
//   - expectedErr: The error to look for
//
// Example:
//
//	reporter.AssertErrorReported(t, errors.New("expected error"))
func (m *MockErrorReporter) AssertErrorReported(t testingT, expectedErr error) {
	t.Helper()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, err := range m.errors {
		if err.Error() == expectedErr.Error() {
			return
		}
	}

	t.Errorf("expected error %q not found in reported errors", expectedErr.Error())
}

// AssertPanicReported asserts that a specific panic was reported.
// Fails the test if the panic was not found in the reported panics.
//
// Parameters:
//   - t: The testing.T instance
//   - expectedPanic: The panic to look for
//
// Example:
//
//	reporter.AssertPanicReported(t, &observability.HandlerPanicError{
//	    ComponentName: "TestComp",
//	    EventName:     "click",
//	})
func (m *MockErrorReporter) AssertPanicReported(t testingT, expectedPanic *observability.HandlerPanicError) {
	t.Helper()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, panic := range m.panics {
		if panic.ComponentName == expectedPanic.ComponentName &&
			panic.EventName == expectedPanic.EventName {
			return
		}
	}

	t.Errorf("expected panic (component=%q, event=%q) not found in reported panics",
		expectedPanic.ComponentName, expectedPanic.EventName)
}

// Reset clears all recorded errors, panics, and contexts.
// Useful for reusing the same reporter across multiple test cases.
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	reporter.Reset()
//	// Reporter is now empty and ready for new test
func (m *MockErrorReporter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors = make([]error, 0)
	m.panics = make([]*observability.HandlerPanicError, 0)
	m.contexts = make([]*observability.ErrorContext, 0)
}

// String returns a human-readable summary of the mock reporter state.
// Useful for debugging test failures.
//
// Returns:
//   - string: Summary of errors, panics, and contexts
//
// Example:
//
//	fmt.Println(reporter.String())
//	// Output: MockErrorReporter: 2 errors, 1 panic, 3 contexts
func (m *MockErrorReporter) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("MockErrorReporter: %d errors, %d panics, %d contexts",
		len(m.errors), len(m.panics), len(m.contexts))
}
