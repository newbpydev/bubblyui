package testutil

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestMockErrorReporter_Creation tests creating a new mock error reporter
func TestMockErrorReporter_Creation(t *testing.T) {
	reporter := NewMockErrorReporter()

	require.NotNil(t, reporter)
	assert.Empty(t, reporter.GetErrors())
	assert.Empty(t, reporter.GetPanics())
	assert.Empty(t, reporter.GetContexts())
}

// TestMockErrorReporter_ReportError tests error reporting
func TestMockErrorReporter_ReportError(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		contexts []*observability.ErrorContext
	}{
		{
			name:     "single error",
			errors:   []error{errors.New("test error")},
			contexts: []*observability.ErrorContext{{ComponentName: "TestComponent"}},
		},
		{
			name:     "multiple errors",
			errors:   []error{errors.New("error 1"), errors.New("error 2")},
			contexts: []*observability.ErrorContext{{ComponentName: "Comp1"}, {ComponentName: "Comp2"}},
		},
		{
			name:     "nil context",
			errors:   []error{errors.New("test error")},
			contexts: []*observability.ErrorContext{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()

			for i, err := range tt.errors {
				reporter.ReportError(err, tt.contexts[i])
			}

			assert.Len(t, reporter.GetErrors(), len(tt.errors))
			assert.Len(t, reporter.GetContexts(), len(tt.contexts))

			for i, err := range tt.errors {
				assert.Equal(t, err, reporter.GetErrors()[i])
			}
		})
	}
}

// TestMockErrorReporter_ReportPanic tests panic reporting
func TestMockErrorReporter_ReportPanic(t *testing.T) {
	tests := []struct {
		name     string
		panics   []*observability.HandlerPanicError
		contexts []*observability.ErrorContext
	}{
		{
			name: "single panic",
			panics: []*observability.HandlerPanicError{
				{ComponentName: "TestComp", EventName: "click", PanicValue: "panic!"},
			},
			contexts: []*observability.ErrorContext{{ComponentName: "TestComp"}},
		},
		{
			name: "multiple panics",
			panics: []*observability.HandlerPanicError{
				{ComponentName: "Comp1", EventName: "event1", PanicValue: "panic 1"},
				{ComponentName: "Comp2", EventName: "event2", PanicValue: "panic 2"},
			},
			contexts: []*observability.ErrorContext{{ComponentName: "Comp1"}, {ComponentName: "Comp2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()

			for i, panic := range tt.panics {
				reporter.ReportPanic(panic, tt.contexts[i])
			}

			assert.Len(t, reporter.GetPanics(), len(tt.panics))
			assert.Len(t, reporter.GetContexts(), len(tt.contexts))

			for i, panic := range tt.panics {
				assert.Equal(t, panic, reporter.GetPanics()[i])
			}
		})
	}
}

// TestMockErrorReporter_AssertErrorReported tests error assertion helper
func TestMockErrorReporter_AssertErrorReported(t *testing.T) {
	t.Run("error found", func(t *testing.T) {
		reporter := NewMockErrorReporter()
		expectedErr := errors.New("test error")

		reporter.ReportError(expectedErr, nil)

		// Should not fail
		reporter.AssertErrorReported(t, expectedErr)
	})

	t.Run("error not found", func(t *testing.T) {
		reporter := NewMockErrorReporter()
		expectedErr := errors.New("test error")
		otherErr := errors.New("other error")

		reporter.ReportError(otherErr, nil)

		// Use mock testing.T to capture failure
		mockT := &mockTestingT{}
		reporter.AssertErrorReported(mockT, expectedErr)

		assert.True(t, mockT.failed, "AssertErrorReported should fail when error not found")
	})
}

// TestMockErrorReporter_AssertPanicReported tests panic assertion helper
func TestMockErrorReporter_AssertPanicReported(t *testing.T) {
	t.Run("panic found", func(t *testing.T) {
		reporter := NewMockErrorReporter()
		expectedPanic := &observability.HandlerPanicError{
			ComponentName: "TestComp",
			EventName:     "click",
			PanicValue:    "panic!",
		}

		reporter.ReportPanic(expectedPanic, nil)

		// Should not fail
		reporter.AssertPanicReported(t, expectedPanic)
	})

	t.Run("panic not found", func(t *testing.T) {
		reporter := NewMockErrorReporter()
		expectedPanic := &observability.HandlerPanicError{
			ComponentName: "TestComp",
			EventName:     "click",
			PanicValue:    "panic!",
		}
		otherPanic := &observability.HandlerPanicError{
			ComponentName: "OtherComp",
			EventName:     "other",
			PanicValue:    "other panic",
		}

		reporter.ReportPanic(otherPanic, nil)

		// Use mock testing.T to capture failure
		mockT := &mockTestingT{}
		reporter.AssertPanicReported(mockT, expectedPanic)

		assert.True(t, mockT.failed, "AssertPanicReported should fail when panic not found")
	})
}

// TestMockErrorReporter_GetBreadcrumbs tests breadcrumb tracking
func TestMockErrorReporter_GetBreadcrumbs(t *testing.T) {
	t.Run("no breadcrumbs", func(t *testing.T) {
		reporter := NewMockErrorReporter()
		breadcrumbs := reporter.GetBreadcrumbs()

		assert.Empty(t, breadcrumbs)
	})

	t.Run("with breadcrumbs in context", func(t *testing.T) {
		reporter := NewMockErrorReporter()

		ctx := &observability.ErrorContext{
			ComponentName: "TestComp",
			Breadcrumbs: []observability.Breadcrumb{
				{Type: "navigation", Message: "User navigated"},
				{Type: "user", Message: "User clicked"},
			},
		}

		reporter.ReportError(errors.New("test"), ctx)

		breadcrumbs := reporter.GetBreadcrumbs()
		assert.Len(t, breadcrumbs, 2)
		assert.Equal(t, "navigation", breadcrumbs[0].Type)
		assert.Equal(t, "user", breadcrumbs[1].Type)
	})

	t.Run("multiple contexts with breadcrumbs", func(t *testing.T) {
		reporter := NewMockErrorReporter()

		ctx1 := &observability.ErrorContext{
			Breadcrumbs: []observability.Breadcrumb{
				{Type: "navigation", Message: "Nav 1"},
			},
		}
		ctx2 := &observability.ErrorContext{
			Breadcrumbs: []observability.Breadcrumb{
				{Type: "user", Message: "User 1"},
				{Type: "user", Message: "User 2"},
			},
		}

		reporter.ReportError(errors.New("error 1"), ctx1)
		reporter.ReportError(errors.New("error 2"), ctx2)

		breadcrumbs := reporter.GetBreadcrumbs()
		assert.Len(t, breadcrumbs, 3)
	})
}

// TestMockErrorReporter_Flush tests flush method
func TestMockErrorReporter_Flush(t *testing.T) {
	reporter := NewMockErrorReporter()

	// Flush should always succeed for mock
	err := reporter.Flush(5 * time.Second)
	assert.NoError(t, err)
}

// TestMockErrorReporter_Reset tests reset functionality
func TestMockErrorReporter_Reset(t *testing.T) {
	reporter := NewMockErrorReporter()

	// Add some data
	reporter.ReportError(errors.New("test error"), nil)
	reporter.ReportPanic(&observability.HandlerPanicError{}, nil)

	assert.NotEmpty(t, reporter.GetErrors())
	assert.NotEmpty(t, reporter.GetPanics())

	// Reset
	reporter.Reset()

	assert.Empty(t, reporter.GetErrors())
	assert.Empty(t, reporter.GetPanics())
	assert.Empty(t, reporter.GetContexts())
}

// TestMockErrorReporter_ThreadSafety tests concurrent access
func TestMockErrorReporter_ThreadSafety(t *testing.T) {
	reporter := NewMockErrorReporter()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(_ int) {
			reporter.ReportError(errors.New("error"), nil)
			reporter.ReportPanic(&observability.HandlerPanicError{}, nil)
			_ = reporter.GetErrors()
			_ = reporter.GetPanics()
			_ = reporter.GetContexts()
			_ = reporter.GetBreadcrumbs()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have collected all errors and panics
	assert.Len(t, reporter.GetErrors(), 10)
	assert.Len(t, reporter.GetPanics(), 10)
}

// TestMockErrorReporter_String tests string representation
func TestMockErrorReporter_String(t *testing.T) {
	reporter := NewMockErrorReporter()

	// Empty reporter
	str := reporter.String()
	assert.Contains(t, str, "MockErrorReporter")
	assert.Contains(t, str, "0 errors")
	assert.Contains(t, str, "0 panics")
	assert.Contains(t, str, "0 contexts")

	// With data
	reporter.ReportError(errors.New("test"), nil)
	reporter.ReportPanic(&observability.HandlerPanicError{}, nil)

	str = reporter.String()
	assert.Contains(t, str, "1 error")
	assert.Contains(t, str, "1 panic")
	assert.Contains(t, str, "2 contexts")
}
