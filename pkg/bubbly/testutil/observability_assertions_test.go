package testutil

import (
	"errors"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
	"github.com/stretchr/testify/assert"
)

func TestNewObservabilityAssertions(t *testing.T) {
	reporter := NewMockErrorReporter()
	oa := NewObservabilityAssertions(reporter)

	assert.NotNil(t, oa)
	assert.Equal(t, reporter, oa.reporter)
}

func TestObservabilityAssertions_AssertErrorReported(t *testing.T) {
	tests := []struct {
		name          string
		reportedError error
		expectedError error
		shouldPass    bool
	}{
		{
			name:          "error found",
			reportedError: errors.New("test error"),
			expectedError: errors.New("test error"),
			shouldPass:    true,
		},
		{
			name:          "error not found",
			reportedError: errors.New("different error"),
			expectedError: errors.New("test error"),
			shouldPass:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Report error
			if tt.reportedError != nil {
				reporter.ReportError(tt.reportedError, nil)
			}

			// Create mock testing.T to capture failures
			mockT := &mockTestingT{}

			// Assert
			oa.AssertErrorReported(mockT, tt.expectedError)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_AssertPanicReported(t *testing.T) {
	tests := []struct {
		name          string
		reportedPanic *observability.HandlerPanicError
		componentName string
		eventName     string
		shouldPass    bool
	}{
		{
			name: "panic found",
			reportedPanic: &observability.HandlerPanicError{
				ComponentName: "TestComp",
				EventName:     "click",
				PanicValue:    "test panic",
			},
			componentName: "TestComp",
			eventName:     "click",
			shouldPass:    true,
		},
		{
			name: "panic not found - different component",
			reportedPanic: &observability.HandlerPanicError{
				ComponentName: "OtherComp",
				EventName:     "click",
				PanicValue:    "test panic",
			},
			componentName: "TestComp",
			eventName:     "click",
			shouldPass:    false,
		},
		{
			name: "panic not found - different event",
			reportedPanic: &observability.HandlerPanicError{
				ComponentName: "TestComp",
				EventName:     "submit",
				PanicValue:    "test panic",
			},
			componentName: "TestComp",
			eventName:     "click",
			shouldPass:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Report panic
			if tt.reportedPanic != nil {
				reporter.ReportPanic(tt.reportedPanic, nil)
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			oa.AssertPanicReported(mockT, tt.componentName, tt.eventName)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_AssertContextHasTag(t *testing.T) {
	tests := []struct {
		name       string
		context    *observability.ErrorContext
		tagKey     string
		tagValue   string
		shouldPass bool
	}{
		{
			name: "tag found with correct value",
			context: &observability.ErrorContext{
				Tags: map[string]string{
					"environment": "test",
					"user_role":   "admin",
				},
			},
			tagKey:     "environment",
			tagValue:   "test",
			shouldPass: true,
		},
		{
			name: "tag found with wrong value",
			context: &observability.ErrorContext{
				Tags: map[string]string{
					"environment": "production",
				},
			},
			tagKey:     "environment",
			tagValue:   "test",
			shouldPass: false,
		},
		{
			name: "tag not found",
			context: &observability.ErrorContext{
				Tags: map[string]string{
					"other": "value",
				},
			},
			tagKey:     "environment",
			tagValue:   "test",
			shouldPass: false,
		},
		{
			name:       "nil tags",
			context:    &observability.ErrorContext{},
			tagKey:     "environment",
			tagValue:   "test",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Report error with context
			reporter.ReportError(errors.New("test"), tt.context)

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			oa.AssertContextHasTag(mockT, tt.tagKey, tt.tagValue)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_AssertContextHasExtra(t *testing.T) {
	tests := []struct {
		name       string
		context    *observability.ErrorContext
		extraKey   string
		shouldPass bool
	}{
		{
			name: "extra key found",
			context: &observability.ErrorContext{
				Extra: map[string]interface{}{
					"user_id":   "12345",
					"form_data": map[string]string{"email": "test@example.com"},
				},
			},
			extraKey:   "user_id",
			shouldPass: true,
		},
		{
			name: "extra key not found",
			context: &observability.ErrorContext{
				Extra: map[string]interface{}{
					"other": "value",
				},
			},
			extraKey:   "user_id",
			shouldPass: false,
		},
		{
			name:       "nil extra",
			context:    &observability.ErrorContext{},
			extraKey:   "user_id",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Report error with context
			reporter.ReportError(errors.New("test"), tt.context)

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			oa.AssertContextHasExtra(mockT, tt.extraKey)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_AssertBreadcrumbRecorded(t *testing.T) {
	// Note: Breadcrumbs are global, so we need to clear them before each test
	tests := []struct {
		name       string
		category   string
		message    string
		shouldPass bool
	}{
		{
			name:       "breadcrumb found",
			category:   "user",
			message:    "User clicked button",
			shouldPass: true,
		},
		{
			name:       "breadcrumb not found",
			category:   "user",
			message:    "Different message",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear global breadcrumbs
			observability.ClearBreadcrumbs()

			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Record breadcrumb
			observability.RecordBreadcrumb("user", "User clicked button", nil)

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			oa.AssertBreadcrumbRecorded(mockT, tt.category, tt.message)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_AssertErrorCount(t *testing.T) {
	tests := []struct {
		name          string
		errorCount    int
		expectedCount int
		shouldPass    bool
	}{
		{
			name:          "correct count",
			errorCount:    3,
			expectedCount: 3,
			shouldPass:    true,
		},
		{
			name:          "wrong count",
			errorCount:    2,
			expectedCount: 3,
			shouldPass:    false,
		},
		{
			name:          "zero count",
			errorCount:    0,
			expectedCount: 0,
			shouldPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Report errors
			for i := 0; i < tt.errorCount; i++ {
				reporter.ReportError(errors.New("test error"), nil)
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			oa.AssertErrorCount(mockT, tt.expectedCount)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_AssertPanicCount(t *testing.T) {
	tests := []struct {
		name          string
		panicCount    int
		expectedCount int
		shouldPass    bool
	}{
		{
			name:          "correct count",
			panicCount:    2,
			expectedCount: 2,
			shouldPass:    true,
		},
		{
			name:          "wrong count",
			panicCount:    1,
			expectedCount: 2,
			shouldPass:    false,
		},
		{
			name:          "zero count",
			panicCount:    0,
			expectedCount: 0,
			shouldPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewMockErrorReporter()
			oa := NewObservabilityAssertions(reporter)

			// Report panics
			for i := 0; i < tt.panicCount; i++ {
				reporter.ReportPanic(&observability.HandlerPanicError{
					ComponentName: "TestComp",
					EventName:     "click",
					PanicValue:    "test",
				}, nil)
			}

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert
			oa.AssertPanicCount(mockT, tt.expectedCount)

			// Verify result
			if tt.shouldPass {
				assert.False(t, mockT.failed, "expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "expected assertion to fail")
			}
		})
	}
}

func TestObservabilityAssertions_GetAllContexts(t *testing.T) {
	reporter := NewMockErrorReporter()
	oa := NewObservabilityAssertions(reporter)

	// Report errors with contexts
	ctx1 := &observability.ErrorContext{
		ComponentName: "Comp1",
		Timestamp:     time.Now(),
	}
	ctx2 := &observability.ErrorContext{
		ComponentName: "Comp2",
		Timestamp:     time.Now(),
	}

	reporter.ReportError(errors.New("error1"), ctx1)
	reporter.ReportError(errors.New("error2"), ctx2)

	// Get all contexts
	contexts := oa.GetAllContexts()

	assert.Len(t, contexts, 2)
	assert.Equal(t, "Comp1", contexts[0].ComponentName)
	assert.Equal(t, "Comp2", contexts[1].ComponentName)
}

func TestObservabilityAssertions_Integration(t *testing.T) {
	// Integration test: full workflow
	reporter := NewMockErrorReporter()
	oa := NewObservabilityAssertions(reporter)

	// Clear breadcrumbs
	observability.ClearBreadcrumbs()

	// Simulate component error with full context
	ctx := &observability.ErrorContext{
		ComponentName: "LoginForm",
		ComponentID:   "form-123",
		EventName:     "submit",
		Timestamp:     time.Now(),
		Tags: map[string]string{
			"environment": "test",
			"user_role":   "admin",
		},
		Extra: map[string]interface{}{
			"user_id":   "12345",
			"form_data": map[string]string{"email": "test@example.com"},
		},
	}

	// Record breadcrumbs
	observability.RecordBreadcrumb("navigation", "User opened login form", nil)
	observability.RecordBreadcrumb("user", "User entered credentials", nil)
	observability.RecordBreadcrumb("user", "User clicked submit", nil)

	// Report error
	reporter.ReportError(errors.New("validation failed"), ctx)

	// Report panic
	panicErr := &observability.HandlerPanicError{
		ComponentName: "LoginForm",
		EventName:     "submit",
		PanicValue:    "unexpected nil",
	}
	reporter.ReportPanic(panicErr, ctx)

	// Assert everything
	oa.AssertErrorReported(t, errors.New("validation failed"))
	oa.AssertPanicReported(t, "LoginForm", "submit")
	oa.AssertContextHasTag(t, "environment", "test")
	oa.AssertContextHasTag(t, "user_role", "admin")
	oa.AssertContextHasExtra(t, "user_id")
	oa.AssertContextHasExtra(t, "form_data")
	oa.AssertBreadcrumbRecorded(t, "navigation", "User opened login form")
	oa.AssertBreadcrumbRecorded(t, "user", "User entered credentials")
	oa.AssertBreadcrumbRecorded(t, "user", "User clicked submit")
	oa.AssertErrorCount(t, 1)
	oa.AssertPanicCount(t, 1)

	// Verify contexts
	contexts := oa.GetAllContexts()
	assert.Len(t, contexts, 2)
	assert.Equal(t, "LoginForm", contexts[0].ComponentName)
	assert.Equal(t, "form-123", contexts[0].ComponentID)
}

// TestObservabilityAssertions_String tests the String method
func TestObservabilityAssertions_String(t *testing.T) {
	reporter := NewMockErrorReporter()
	oa := NewObservabilityAssertions(reporter)

	// Add some data
	reporter.ReportError(errors.New("test error"), &observability.ErrorContext{})
	reporter.ReportPanic(&observability.HandlerPanicError{}, &observability.ErrorContext{})
	observability.RecordBreadcrumb("test", "message", nil)

	// Get string representation
	str := oa.String()

	// Verify it contains expected information
	assert.Contains(t, str, "ObservabilityAssertions")
	assert.Contains(t, str, "1 errors")
	assert.Contains(t, str, "1 panics")
	assert.Contains(t, str, "contexts")
	assert.Contains(t, str, "breadcrumbs")
}

// mockTestingT is already defined in assertions_state_test.go
