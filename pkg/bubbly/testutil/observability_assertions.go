package testutil

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// ObservabilityAssertions provides high-level assertions for testing observability hooks
// and telemetry data collection. It wraps a MockErrorReporter and provides convenient
// methods for asserting that errors, panics, contexts, and breadcrumbs were recorded correctly.
//
// This utility helps test that your components properly integrate with the observability
// system, ensuring errors are reported with rich context for debugging.
//
// Example usage:
//
//	reporter := NewMockErrorReporter()
//	observability.SetErrorReporter(reporter)
//	defer observability.SetErrorReporter(nil)
//
//	oa := NewObservabilityAssertions(reporter)
//
//	// Run code that may report errors
//	component.Emit("event", data)
//
//	// Assert errors were reported correctly
//	oa.AssertErrorReported(t, expectedErr)
//	oa.AssertContextHasTag(t, "environment", "test")
//	oa.AssertBreadcrumbRecorded(t, "user", "User clicked button")
type ObservabilityAssertions struct {
	reporter *MockErrorReporter
}

// NewObservabilityAssertions creates a new observability assertions helper.
// The reporter should be the same MockErrorReporter that was configured via
// observability.SetErrorReporter().
//
// Parameters:
//   - reporter: The mock error reporter to assert against
//
// Returns:
//   - *ObservabilityAssertions: A new assertions helper
//
// Example:
//
//	reporter := NewMockErrorReporter()
//	observability.SetErrorReporter(reporter)
//	oa := NewObservabilityAssertions(reporter)
func NewObservabilityAssertions(reporter *MockErrorReporter) *ObservabilityAssertions {
	return &ObservabilityAssertions{
		reporter: reporter,
	}
}

// AssertErrorReported asserts that a specific error was reported to the observability system.
// Fails the test if the error was not found in the reported errors.
//
// The error is matched by its error message string.
//
// Parameters:
//   - t: The testing.T instance
//   - expectedErr: The error to look for
//
// Example:
//
//	oa.AssertErrorReported(t, errors.New("validation failed"))
func (oa *ObservabilityAssertions) AssertErrorReported(t testingT, expectedErr error) {
	t.Helper()

	errors := oa.reporter.GetErrors()
	for _, err := range errors {
		if err.Error() == expectedErr.Error() {
			return
		}
	}

	t.Errorf("expected error %q not found in reported errors (found %d errors)",
		expectedErr.Error(), len(errors))
}

// AssertPanicReported asserts that a panic was reported for a specific component and event.
// Fails the test if no matching panic was found.
//
// Parameters:
//   - t: The testing.T instance
//   - componentName: The name of the component that panicked
//   - eventName: The name of the event being handled when the panic occurred
//
// Example:
//
//	oa.AssertPanicReported(t, "LoginForm", "submit")
func (oa *ObservabilityAssertions) AssertPanicReported(t testingT, componentName, eventName string) {
	t.Helper()

	panics := oa.reporter.GetPanics()
	for _, panic := range panics {
		if panic.ComponentName == componentName && panic.EventName == eventName {
			return
		}
	}

	t.Errorf("expected panic (component=%q, event=%q) not found in reported panics (found %d panics)",
		componentName, eventName, len(panics))
}

// AssertContextHasTag asserts that at least one error context has a specific tag with the expected value.
// Fails the test if no context has the tag or if the tag has a different value.
//
// Tags are used for filtering and grouping errors in observability systems.
//
// Parameters:
//   - t: The testing.T instance
//   - key: The tag key to look for
//   - expectedValue: The expected tag value
//
// Example:
//
//	oa.AssertContextHasTag(t, "environment", "production")
//	oa.AssertContextHasTag(t, "user_role", "admin")
func (oa *ObservabilityAssertions) AssertContextHasTag(t testingT, key, expectedValue string) {
	t.Helper()

	contexts := oa.reporter.GetContexts()
	for _, ctx := range contexts {
		if ctx == nil || ctx.Tags == nil {
			continue
		}

		if value, ok := ctx.Tags[key]; ok {
			if value == expectedValue {
				return
			}
			t.Errorf("tag %q found but has value %q, expected %q", key, value, expectedValue)
			return
		}
	}

	t.Errorf("tag %q with value %q not found in any error context (checked %d contexts)",
		key, expectedValue, len(contexts))
}

// AssertContextHasExtra asserts that at least one error context has a specific extra data key.
// Fails the test if no context has the extra data key.
//
// Extra data contains arbitrary additional information about errors.
//
// Parameters:
//   - t: The testing.T instance
//   - key: The extra data key to look for
//
// Example:
//
//	oa.AssertContextHasExtra(t, "user_id")
//	oa.AssertContextHasExtra(t, "form_data")
func (oa *ObservabilityAssertions) AssertContextHasExtra(t testingT, key string) {
	t.Helper()

	contexts := oa.reporter.GetContexts()
	for _, ctx := range contexts {
		if ctx == nil || ctx.Extra == nil {
			continue
		}

		if _, ok := ctx.Extra[key]; ok {
			return
		}
	}

	t.Errorf("extra data key %q not found in any error context (checked %d contexts)",
		key, len(contexts))
}

// AssertBreadcrumbRecorded asserts that a breadcrumb with the specified category and message
// was recorded in the global breadcrumb buffer.
//
// Breadcrumbs provide a trail of actions leading up to an error.
//
// Note: Breadcrumbs are stored globally, so you may want to call observability.ClearBreadcrumbs()
// at the start of your test to ensure isolation.
//
// Parameters:
//   - t: The testing.T instance
//   - category: The breadcrumb category to look for
//   - message: The breadcrumb message to look for
//
// Example:
//
//	observability.ClearBreadcrumbs()
//	// ... run test code ...
//	oa.AssertBreadcrumbRecorded(t, "user", "User clicked submit button")
func (oa *ObservabilityAssertions) AssertBreadcrumbRecorded(t testingT, category, message string) {
	t.Helper()

	breadcrumbs := observability.GetBreadcrumbs()
	for _, bc := range breadcrumbs {
		if bc.Category == category && bc.Message == message {
			return
		}
	}

	t.Errorf("breadcrumb (category=%q, message=%q) not found in recorded breadcrumbs (found %d breadcrumbs)",
		category, message, len(breadcrumbs))
}

// AssertErrorCount asserts that exactly the expected number of errors were reported.
// Fails the test if the count doesn't match.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected number of errors
//
// Example:
//
//	oa.AssertErrorCount(t, 3)
func (oa *ObservabilityAssertions) AssertErrorCount(t testingT, expected int) {
	t.Helper()

	errors := oa.reporter.GetErrors()
	actual := len(errors)

	if actual != expected {
		t.Errorf("expected %d errors, got %d", expected, actual)
	}
}

// AssertPanicCount asserts that exactly the expected number of panics were reported.
// Fails the test if the count doesn't match.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected number of panics
//
// Example:
//
//	oa.AssertPanicCount(t, 1)
func (oa *ObservabilityAssertions) AssertPanicCount(t testingT, expected int) {
	t.Helper()

	panics := oa.reporter.GetPanics()
	actual := len(panics)

	if actual != expected {
		t.Errorf("expected %d panics, got %d", expected, actual)
	}
}

// GetAllContexts returns all error contexts that were reported.
// This is useful for custom assertions or detailed inspection.
//
// Returns:
//   - []*observability.ErrorContext: All error contexts
//
// Example:
//
//	contexts := oa.GetAllContexts()
//	for _, ctx := range contexts {
//	    assert.NotEmpty(t, ctx.ComponentName)
//	    assert.NotZero(t, ctx.Timestamp)
//	}
func (oa *ObservabilityAssertions) GetAllContexts() []*observability.ErrorContext {
	return oa.reporter.GetContexts()
}

// String returns a human-readable summary of the observability state.
// Useful for debugging test failures.
//
// Returns:
//   - string: Summary of errors, panics, contexts, and breadcrumbs
//
// Example:
//
//	fmt.Println(oa.String())
//	// Output: ObservabilityAssertions: 2 errors, 1 panic, 3 contexts, 5 breadcrumbs
func (oa *ObservabilityAssertions) String() string {
	errors := oa.reporter.GetErrors()
	panics := oa.reporter.GetPanics()
	contexts := oa.reporter.GetContexts()
	breadcrumbs := observability.GetBreadcrumbs()

	return fmt.Sprintf("ObservabilityAssertions: %d errors, %d panics, %d contexts, %d breadcrumbs",
		len(errors), len(panics), len(contexts), len(breadcrumbs))
}
