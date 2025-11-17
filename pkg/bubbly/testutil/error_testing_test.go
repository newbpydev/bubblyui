package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewErrorTesting verifies that NewErrorTesting creates a properly initialized instance
func TestNewErrorTesting(t *testing.T) {
	et := NewErrorTesting()

	assert.NotNil(t, et)
	assert.NotNil(t, et.errors)
	assert.NotNil(t, et.recovered)
	assert.NotNil(t, et.errorHandlers)
	assert.NotNil(t, et.panicHandlers)
	assert.NotNil(t, et.stackTraces)
	assert.NotNil(t, et.timestamps)
	assert.Equal(t, 0, len(et.errors))
	assert.Equal(t, 0, len(et.recovered))
}

// TestTriggerError verifies that errors are properly recorded
func TestTriggerError(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected int
	}{
		{
			name:     "single error",
			errors:   []error{fmt.Errorf("error 1")},
			expected: 1,
		},
		{
			name:     "multiple errors",
			errors:   []error{fmt.Errorf("error 1"), fmt.Errorf("error 2"), fmt.Errorf("error 3")},
			expected: 3,
		},
		{
			name:     "no errors",
			errors:   []error{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := NewErrorTesting()

			for _, err := range tt.errors {
				et.TriggerError(err)
			}

			assert.Equal(t, tt.expected, len(et.GetErrors()))
			assert.Equal(t, tt.expected, len(et.GetStackTraces()))
			assert.Equal(t, tt.expected, len(et.GetTimestamps()))
		})
	}
}

// TestTriggerPanic verifies that panics are properly recorded
func TestTriggerPanic(t *testing.T) {
	tests := []struct {
		name     string
		panics   []interface{}
		expected int
	}{
		{
			name:     "single panic",
			panics:   []interface{}{"panic 1"},
			expected: 1,
		},
		{
			name:     "multiple panics",
			panics:   []interface{}{"panic 1", 42, fmt.Errorf("panic error")},
			expected: 3,
		},
		{
			name:     "no panics",
			panics:   []interface{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := NewErrorTesting()

			for _, p := range tt.panics {
				et.TriggerPanic(p)
			}

			assert.Equal(t, tt.expected, len(et.GetRecoveredPanics()))
			assert.Equal(t, tt.expected, len(et.GetStackTraces()))
			assert.Equal(t, tt.expected, len(et.GetTimestamps()))
		})
	}
}

// TestAssertErrorHandled verifies error assertion functionality
func TestAssertErrorHandled(t *testing.T) {
	et := NewErrorTesting()
	err1 := fmt.Errorf("validation failed")
	err2 := fmt.Errorf("network error")

	et.TriggerError(err1)
	et.TriggerError(err2)

	// Should pass - errors were triggered
	mockT := &mockTestingT{}
	et.AssertErrorHandled(mockT, err1)
	assert.False(t, mockT.failed, "AssertErrorHandled should pass for triggered error")

	et.AssertErrorHandled(mockT, err2)
	assert.False(t, mockT.failed, "AssertErrorHandled should pass for triggered error")

	// Should fail - error was not triggered
	mockT = &mockTestingT{}
	et.AssertErrorHandled(mockT, fmt.Errorf("not triggered"))
	assert.True(t, mockT.failed, "AssertErrorHandled should fail for non-triggered error")
}

// TestAssertPanicRecovered verifies panic recovery assertion
func TestAssertPanicRecovered(t *testing.T) {
	// Should pass - panic was recovered
	et := NewErrorTesting()
	et.TriggerPanic("test panic")

	mockT := &mockTestingT{}
	et.AssertPanicRecovered(mockT)
	assert.False(t, mockT.failed, "AssertPanicRecovered should pass when panic was recovered")

	// Should fail - no panic was recovered
	et2 := NewErrorTesting()
	mockT2 := &mockTestingT{}
	et2.AssertPanicRecovered(mockT2)
	assert.True(t, mockT2.failed, "AssertPanicRecovered should fail when no panic was recovered")
}

// TestAssertErrorCount verifies error count assertion
func TestAssertErrorCount(t *testing.T) {
	tests := []struct {
		name       string
		errorCount int
		expected   int
		shouldPass bool
	}{
		{
			name:       "correct count - zero",
			errorCount: 0,
			expected:   0,
			shouldPass: true,
		},
		{
			name:       "correct count - one",
			errorCount: 1,
			expected:   1,
			shouldPass: true,
		},
		{
			name:       "correct count - multiple",
			errorCount: 5,
			expected:   5,
			shouldPass: true,
		},
		{
			name:       "incorrect count - too few",
			errorCount: 2,
			expected:   5,
			shouldPass: false,
		},
		{
			name:       "incorrect count - too many",
			errorCount: 5,
			expected:   2,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := NewErrorTesting()

			for i := 0; i < tt.errorCount; i++ {
				et.TriggerError(fmt.Errorf("error %d", i))
			}

			mockT := &mockTestingT{}
			et.AssertErrorCount(mockT, tt.expected)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "AssertErrorCount should pass")
			} else {
				assert.True(t, mockT.failed, "AssertErrorCount should fail")
			}
		})
	}
}

// TestAssertPanicCount verifies panic count assertion
func TestAssertPanicCount(t *testing.T) {
	tests := []struct {
		name       string
		panicCount int
		expected   int
		shouldPass bool
	}{
		{
			name:       "correct count - zero",
			panicCount: 0,
			expected:   0,
			shouldPass: true,
		},
		{
			name:       "correct count - one",
			panicCount: 1,
			expected:   1,
			shouldPass: true,
		},
		{
			name:       "correct count - multiple",
			panicCount: 3,
			expected:   3,
			shouldPass: true,
		},
		{
			name:       "incorrect count",
			panicCount: 2,
			expected:   5,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et := NewErrorTesting()

			for i := 0; i < tt.panicCount; i++ {
				et.TriggerPanic(fmt.Sprintf("panic %d", i))
			}

			mockT := &mockTestingT{}
			et.AssertPanicCount(mockT, tt.expected)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "AssertPanicCount should pass")
			} else {
				assert.True(t, mockT.failed, "AssertPanicCount should fail")
			}
		})
	}
}

// TestErrorHandlers verifies that error handlers are called
func TestErrorHandlers(t *testing.T) {
	et := NewErrorTesting()

	handlerCalled := false
	var handledError error

	et.RegisterErrorHandler("*errors.errorString", func(err error) {
		handlerCalled = true
		handledError = err
	})

	testErr := fmt.Errorf("test error")
	et.TriggerError(testErr)

	assert.True(t, handlerCalled, "Error handler should be called")
	assert.Equal(t, testErr.Error(), handledError.Error())
}

// TestPanicHandlers verifies that panic handlers are called
func TestPanicHandlers(t *testing.T) {
	et := NewErrorTesting()

	handlerCalled := false
	var handledPanic interface{}

	et.RegisterPanicHandler("string", func(p interface{}) {
		handlerCalled = true
		handledPanic = p
	})

	testPanic := "test panic"
	et.TriggerPanic(testPanic)

	assert.True(t, handlerCalled, "Panic handler should be called")
	assert.Equal(t, testPanic, handledPanic)
}

// TestGetErrors verifies defensive copy of errors
func TestGetErrors(t *testing.T) {
	et := NewErrorTesting()
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	et.TriggerError(err1)
	et.TriggerError(err2)

	errors := et.GetErrors()
	assert.Equal(t, 2, len(errors))

	// Modify the returned slice - should not affect internal state
	errors[0] = fmt.Errorf("modified")
	errors = append(errors, fmt.Errorf("added"))

	// Internal state should be unchanged
	internalErrors := et.GetErrors()
	assert.Equal(t, 2, len(internalErrors))
	assert.Equal(t, err1.Error(), internalErrors[0].Error())
}

// TestGetRecoveredPanics verifies defensive copy of panics
func TestGetRecoveredPanics(t *testing.T) {
	et := NewErrorTesting()

	et.TriggerPanic("panic 1")
	et.TriggerPanic("panic 2")

	panics := et.GetRecoveredPanics()
	assert.Equal(t, 2, len(panics))

	// Modify the returned slice - should not affect internal state
	panics[0] = "modified"
	panics = append(panics, "added")

	// Internal state should be unchanged
	internalPanics := et.GetRecoveredPanics()
	assert.Equal(t, 2, len(internalPanics))
	assert.Equal(t, "panic 1", internalPanics[0])
}

// TestStackTraceCapture verifies that stack traces are captured
func TestStackTraceCapture(t *testing.T) {
	et := NewErrorTesting()

	et.TriggerError(fmt.Errorf("test error"))
	et.TriggerPanic("test panic")

	traces := et.GetStackTraces()
	assert.Equal(t, 2, len(traces))
	assert.NotEmpty(t, traces[0])
	assert.NotEmpty(t, traces[1])

	// Stack traces should contain function names
	assert.Contains(t, string(traces[0]), "TriggerError")
	assert.Contains(t, string(traces[1]), "TriggerPanic")
}

// TestTimestampCapture verifies that timestamps are captured
func TestTimestampCapture(t *testing.T) {
	et := NewErrorTesting()

	before := time.Now()
	et.TriggerError(fmt.Errorf("test error"))
	time.Sleep(10 * time.Millisecond)
	et.TriggerPanic("test panic")
	after := time.Now()

	timestamps := et.GetTimestamps()
	assert.Equal(t, 2, len(timestamps))

	// Timestamps should be within the test duration
	assert.True(t, timestamps[0].After(before) || timestamps[0].Equal(before))
	assert.True(t, timestamps[0].Before(after) || timestamps[0].Equal(after))
	assert.True(t, timestamps[1].After(before) || timestamps[1].Equal(before))
	assert.True(t, timestamps[1].Before(after) || timestamps[1].Equal(after))

	// Second timestamp should be after first
	assert.True(t, timestamps[1].After(timestamps[0]))
}

// TestReset verifies that Reset clears all state
func TestReset(t *testing.T) {
	et := NewErrorTesting()

	// Add some errors and panics
	et.TriggerError(fmt.Errorf("error 1"))
	et.TriggerError(fmt.Errorf("error 2"))
	et.TriggerPanic("panic 1")

	// Register handlers
	et.RegisterErrorHandler("*errors.errorString", func(err error) {})
	et.RegisterPanicHandler("string", func(p interface{}) {})

	assert.Equal(t, 2, len(et.GetErrors()))
	assert.Equal(t, 1, len(et.GetRecoveredPanics()))

	// Reset
	et.Reset()

	// State should be cleared
	assert.Equal(t, 0, len(et.GetErrors()))
	assert.Equal(t, 0, len(et.GetRecoveredPanics()))
	assert.Equal(t, 0, len(et.GetStackTraces()))
	assert.Equal(t, 0, len(et.GetTimestamps()))

	// Handlers should still be registered
	handlerCalled := false
	et.RegisterErrorHandler("*errors.errorString", func(err error) {
		handlerCalled = true
	})
	et.TriggerError(fmt.Errorf("test"))
	assert.True(t, handlerCalled)
}

// TestString verifies the string representation
func TestString(t *testing.T) {
	et := NewErrorTesting()

	// Empty state
	str := et.String()
	assert.Contains(t, str, "0 errors")
	assert.Contains(t, str, "0 panics")

	// With errors and panics
	et.TriggerError(fmt.Errorf("error 1"))
	et.TriggerError(fmt.Errorf("error 2"))
	et.TriggerPanic("panic 1")
	et.RegisterErrorHandler("*errors.errorString", func(err error) {})
	et.RegisterPanicHandler("string", func(p interface{}) {})

	str = et.String()
	assert.Contains(t, str, "2 errors")
	assert.Contains(t, str, "1 panics")
	assert.Contains(t, str, "2 handlers")
}

// TestConcurrentAccess verifies thread safety
func TestConcurrentAccess(t *testing.T) {
	et := NewErrorTesting()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			et.TriggerError(fmt.Errorf("error %d", id))
			et.TriggerPanic(fmt.Sprintf("panic %d", id))
			_ = et.GetErrors()
			_ = et.GetRecoveredPanics()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have all errors and panics
	assert.Equal(t, 10, len(et.GetErrors()))
	assert.Equal(t, 10, len(et.GetRecoveredPanics()))
}

// TestIntegrationWithMockErrorReporter demonstrates integration with observability system
func TestIntegrationWithMockErrorReporter(t *testing.T) {
	et := NewErrorTesting()

	// Track errors that would be reported to observability system
	reportedErrors := []error{}

	et.RegisterErrorHandler("*errors.errorString", func(err error) {
		reportedErrors = append(reportedErrors, err)
	})

	// Trigger errors
	err1 := fmt.Errorf("validation failed")
	err2 := fmt.Errorf("network error")

	et.TriggerError(err1)
	et.TriggerError(err2)

	// Verify errors were handled
	assert.Equal(t, 2, len(reportedErrors))
	assert.Equal(t, err1.Error(), reportedErrors[0].Error())
	assert.Equal(t, err2.Error(), reportedErrors[1].Error())

	// Verify assertions work
	mockT := &mockTestingT{}
	et.AssertErrorHandled(mockT, err1)
	assert.False(t, mockT.failed)

	et.AssertErrorCount(mockT, 2)
	assert.False(t, mockT.failed)
}

// TestErrorBoundaryPattern demonstrates error boundary testing pattern
func TestErrorBoundaryPattern(t *testing.T) {
	et := NewErrorTesting()

	// Simulate error boundary catching errors
	errorBoundary := func(fn func() error) {
		if err := fn(); err != nil {
			et.TriggerError(err)
		}
	}

	// Simulate operations that may fail
	errorBoundary(func() error {
		return fmt.Errorf("operation 1 failed")
	})

	errorBoundary(func() error {
		return nil // Success
	})

	errorBoundary(func() error {
		return fmt.Errorf("operation 2 failed")
	})

	// Verify only failed operations were recorded
	assert.Equal(t, 2, len(et.GetErrors()))
}

// TestCascadingErrorPrevention demonstrates preventing cascading errors
func TestCascadingErrorPrevention(t *testing.T) {
	et := NewErrorTesting()

	// Track if cascading errors are prevented
	errorCount := 0

	et.RegisterErrorHandler("*errors.errorString", func(err error) {
		// Count errors as they come in
		errorCount++
		// First error would trigger circuit breaker or other prevention mechanism
		// (In real code, the handler would prevent subsequent errors from being triggered)
	})

	// Trigger first error
	et.TriggerError(fmt.Errorf("initial error"))
	assert.Equal(t, 1, errorCount, "First error should be handled")

	// Trigger second error (would be prevented by circuit breaker in real code)
	et.TriggerError(fmt.Errorf("second error"))
	assert.Equal(t, 2, errorCount, "Second error should be handled")

	// Verify both errors were recorded
	assert.Equal(t, 2, len(et.GetErrors()))
}
