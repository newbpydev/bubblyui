package testutil

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// ErrorTesting provides utilities for testing error handling and recovery.
// It tracks errors, panics, and recovery strategies to verify that components
// handle errors gracefully and don't crash the application.
//
// This utility helps test:
//   - Error catching and handling
//   - Panic recovery
//   - Error boundaries
//   - Stack trace capture
//   - Recovery strategies
//   - Cascading error prevention
//
// Thread-safe: All methods are safe for concurrent use.
//
// Example usage:
//
//	et := NewErrorTesting()
//
//	// Trigger an error
//	et.TriggerError(fmt.Errorf("validation failed"))
//
//	// Trigger a panic with recovery
//	et.TriggerPanic("unexpected nil pointer")
//
//	// Assert error was handled
//	et.AssertErrorHandled(t, fmt.Errorf("validation failed"))
//
//	// Assert panic was recovered
//	et.AssertPanicRecovered(t)
//
//	// Assert error count
//	et.AssertErrorCount(t, 1)
type ErrorTesting struct {
	// errors tracks all errors that have been triggered
	errors []error

	// recovered tracks all panics that have been recovered
	recovered []interface{}

	// errorHandlers maps error types to handler functions
	errorHandlers map[string]func(error)

	// panicHandlers maps panic types to handler functions
	panicHandlers map[string]func(interface{})

	// stackTraces stores stack traces for each error/panic
	stackTraces [][]byte

	// timestamps stores when each error/panic occurred
	timestamps []time.Time

	// mu protects concurrent access to all fields
	mu sync.RWMutex
}

// NewErrorTesting creates a new error testing utility.
// The utility starts with empty error and panic tracking.
//
// Returns:
//   - *ErrorTesting: A new error testing instance ready for use
//
// Example:
//
//	et := NewErrorTesting()
//	et.TriggerError(fmt.Errorf("test error"))
func NewErrorTesting() *ErrorTesting {
	return &ErrorTesting{
		errors:        make([]error, 0),
		recovered:     make([]interface{}, 0),
		errorHandlers: make(map[string]func(error)),
		panicHandlers: make(map[string]func(interface{})),
		stackTraces:   make([][]byte, 0),
		timestamps:    make([]time.Time, 0),
	}
}

// TriggerError simulates an error being triggered and optionally handled.
// The error is recorded along with its stack trace and timestamp.
//
// If an error handler is registered for this error type, it will be called.
//
// Parameters:
//   - err: The error to trigger
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerError(fmt.Errorf("validation failed"))
//	et.TriggerError(&observability.HandlerPanicError{
//	    ComponentName: "Button",
//	    EventName:     "click",
//	    PanicValue:    "nil pointer",
//	})
func (et *ErrorTesting) TriggerError(err error) {
	et.mu.Lock()
	defer et.mu.Unlock()

	// Record the error
	et.errors = append(et.errors, err)
	et.stackTraces = append(et.stackTraces, debug.Stack())
	et.timestamps = append(et.timestamps, time.Now())

	// Call registered handler if exists
	errorType := fmt.Sprintf("%T", err)
	if handler, ok := et.errorHandlers[errorType]; ok {
		handler(err)
	}
}

// TriggerPanic simulates a panic being triggered and recovered.
// The panic value is recorded along with its stack trace and timestamp.
//
// If a panic handler is registered for this panic type, it will be called.
//
// Parameters:
//   - panicValue: The value passed to panic()
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerPanic("unexpected nil pointer")
//	et.TriggerPanic(fmt.Errorf("critical error"))
//	et.TriggerPanic(42) // Any type can be panic value
func (et *ErrorTesting) TriggerPanic(panicValue interface{}) {
	et.mu.Lock()
	defer et.mu.Unlock()

	// Record the panic
	et.recovered = append(et.recovered, panicValue)
	et.stackTraces = append(et.stackTraces, debug.Stack())
	et.timestamps = append(et.timestamps, time.Now())

	// Call registered handler if exists
	panicType := fmt.Sprintf("%T", panicValue)
	if handler, ok := et.panicHandlers[panicType]; ok {
		handler(panicValue)
	}
}

// RegisterErrorHandler registers a handler function for a specific error type.
// When TriggerError is called with an error of this type, the handler will be invoked.
//
// This is useful for testing error handling strategies and recovery mechanisms.
//
// Parameters:
//   - errorType: The type name of the error (e.g., "*errors.errorString", "*observability.HandlerPanicError")
//   - handler: The function to call when this error type is triggered
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.RegisterErrorHandler("*observability.HandlerPanicError", func(err error) {
//	    panicErr := err.(*observability.HandlerPanicError)
//	    fmt.Printf("Handler panic in %s: %v\n", panicErr.ComponentName, panicErr.PanicValue)
//	})
func (et *ErrorTesting) RegisterErrorHandler(errorType string, handler func(error)) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.errorHandlers[errorType] = handler
}

// RegisterPanicHandler registers a handler function for a specific panic type.
// When TriggerPanic is called with a value of this type, the handler will be invoked.
//
// This is useful for testing panic recovery strategies.
//
// Parameters:
//   - panicType: The type name of the panic value (e.g., "string", "*errors.errorString")
//   - handler: The function to call when this panic type is triggered
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.RegisterPanicHandler("string", func(p interface{}) {
//	    msg := p.(string)
//	    fmt.Printf("String panic: %s\n", msg)
//	})
func (et *ErrorTesting) RegisterPanicHandler(panicType string, handler func(interface{})) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.panicHandlers[panicType] = handler
}

// AssertErrorHandled asserts that a specific error was triggered.
// This verifies that the error was properly caught and recorded.
//
// Parameters:
//   - t: The testing.T instance
//   - expectedErr: The error that should have been triggered
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerError(fmt.Errorf("validation failed"))
//	et.AssertErrorHandled(t, fmt.Errorf("validation failed"))
func (et *ErrorTesting) AssertErrorHandled(t testingT, expectedErr error) {
	t.Helper()

	et.mu.RLock()
	defer et.mu.RUnlock()

	for _, err := range et.errors {
		if err.Error() == expectedErr.Error() {
			return
		}
	}

	t.Errorf("expected error %q to be handled, but it was not found in %d recorded errors",
		expectedErr.Error(), len(et.errors))
}

// AssertPanicRecovered asserts that at least one panic was recovered.
// This verifies that panic recovery is working correctly.
//
// Parameters:
//   - t: The testing.T instance
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerPanic("unexpected nil")
//	et.AssertPanicRecovered(t)
func (et *ErrorTesting) AssertPanicRecovered(t testingT) {
	t.Helper()

	et.mu.RLock()
	defer et.mu.RUnlock()

	if len(et.recovered) == 0 {
		t.Errorf("expected at least one panic to be recovered, but none were recorded")
	}
}

// AssertErrorCount asserts that exactly the expected number of errors were triggered.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected number of errors
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerError(fmt.Errorf("error 1"))
//	et.TriggerError(fmt.Errorf("error 2"))
//	et.AssertErrorCount(t, 2)
func (et *ErrorTesting) AssertErrorCount(t testingT, expected int) {
	t.Helper()

	et.mu.RLock()
	defer et.mu.RUnlock()

	if len(et.errors) != expected {
		t.Errorf("expected %d errors, but got %d", expected, len(et.errors))
	}
}

// AssertPanicCount asserts that exactly the expected number of panics were recovered.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected number of panics
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerPanic("panic 1")
//	et.TriggerPanic("panic 2")
//	et.AssertPanicCount(t, 2)
func (et *ErrorTesting) AssertPanicCount(t testingT, expected int) {
	t.Helper()

	et.mu.RLock()
	defer et.mu.RUnlock()

	if len(et.recovered) != expected {
		t.Errorf("expected %d panics, but got %d", expected, len(et.recovered))
	}
}

// GetErrors returns a copy of all triggered errors.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - []error: Slice of all triggered errors
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	errors := et.GetErrors()
//	for _, err := range errors {
//	    fmt.Printf("Error: %v\n", err)
//	}
func (et *ErrorTesting) GetErrors() []error {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Return a copy to prevent external modification
	errors := make([]error, len(et.errors))
	copy(errors, et.errors)
	return errors
}

// GetRecoveredPanics returns a copy of all recovered panic values.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - []interface{}: Slice of all recovered panic values
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	panics := et.GetRecoveredPanics()
//	for _, p := range panics {
//	    fmt.Printf("Panic: %v\n", p)
//	}
func (et *ErrorTesting) GetRecoveredPanics() []interface{} {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Return a copy to prevent external modification
	recovered := make([]interface{}, len(et.recovered))
	copy(recovered, et.recovered)
	return recovered
}

// GetStackTraces returns a copy of all captured stack traces.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - [][]byte: Slice of all stack traces
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	traces := et.GetStackTraces()
//	for i, trace := range traces {
//	    fmt.Printf("Stack trace %d:\n%s\n", i, string(trace))
//	}
func (et *ErrorTesting) GetStackTraces() [][]byte {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Return a copy to prevent external modification
	traces := make([][]byte, len(et.stackTraces))
	copy(traces, et.stackTraces)
	return traces
}

// GetTimestamps returns a copy of all error/panic timestamps.
// Returns a defensive copy to prevent external modification.
//
// Returns:
//   - []time.Time: Slice of all timestamps
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	timestamps := et.GetTimestamps()
//	for i, ts := range timestamps {
//	    fmt.Printf("Event %d occurred at: %v\n", i, ts)
//	}
func (et *ErrorTesting) GetTimestamps() []time.Time {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Return a copy to prevent external modification
	timestamps := make([]time.Time, len(et.timestamps))
	copy(timestamps, et.timestamps)
	return timestamps
}

// Reset clears all tracked errors, panics, and handlers.
// This is useful for reusing the same ErrorTesting instance across multiple tests.
//
// Thread-safe: Safe to call concurrently.
//
// Example:
//
//	et.TriggerError(fmt.Errorf("error 1"))
//	et.Reset()
//	et.AssertErrorCount(t, 0) // Passes - errors were cleared
func (et *ErrorTesting) Reset() {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.errors = make([]error, 0)
	et.recovered = make([]interface{}, 0)
	et.stackTraces = make([][]byte, 0)
	et.timestamps = make([]time.Time, 0)
	// Keep handlers registered - they're part of the test setup
}

// String returns a human-readable summary of the error testing state.
// Useful for debugging test failures.
//
// Returns:
//   - string: Summary of errors, panics, and handlers
//
// Example:
//
//	fmt.Println(et.String())
//	// Output: ErrorTesting: 2 errors, 1 panic, 3 handlers
func (et *ErrorTesting) String() string {
	et.mu.RLock()
	defer et.mu.RUnlock()

	totalHandlers := len(et.errorHandlers) + len(et.panicHandlers)
	return fmt.Sprintf("ErrorTesting: %d errors, %d panics, %d handlers",
		len(et.errors), len(et.recovered), totalHandlers)
}
