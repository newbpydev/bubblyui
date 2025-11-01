package composables

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorsDefined verifies that all expected errors are defined
func TestErrorsDefined(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		isNil bool
	}{
		{
			name:  "ErrComposableOutsideSetup is defined",
			err:   ErrComposableOutsideSetup,
			isNil: false,
		},
		{
			name:  "ErrCircularComposable is defined",
			err:   ErrCircularComposable,
			isNil: false,
		},
		{
			name:  "ErrInjectNotFound is defined",
			err:   ErrInjectNotFound,
			isNil: false,
		},
		{
			name:  "ErrInvalidComposableState is defined",
			err:   ErrInvalidComposableState,
			isNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isNil {
				assert.Nil(t, tt.err)
			} else {
				assert.NotNil(t, tt.err)
			}
		})
	}
}

// TestErrorMessages verifies that error messages are clear and helpful
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedMessage string
	}{
		{
			name:            "ErrComposableOutsideSetup message is clear",
			err:             ErrComposableOutsideSetup,
			expectedMessage: "composable must be called within Setup function",
		},
		{
			name:            "ErrCircularComposable message is clear",
			err:             ErrCircularComposable,
			expectedMessage: "circular composable dependency detected",
		},
		{
			name:            "ErrInjectNotFound message is clear",
			err:             ErrInjectNotFound,
			expectedMessage: "inject key not found in component tree",
		},
		{
			name:            "ErrInvalidComposableState message is clear",
			err:             ErrInvalidComposableState,
			expectedMessage: "composable is in an invalid state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMessage, tt.err.Error())
		})
	}
}

// TestErrorIsChecking verifies that errors can be checked with errors.Is()
func TestErrorIsChecking(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		err      error
		expected bool
	}{
		{
			name:     "ErrComposableOutsideSetup matches itself",
			sentinel: ErrComposableOutsideSetup,
			err:      ErrComposableOutsideSetup,
			expected: true,
		},
		{
			name:     "ErrComposableOutsideSetup doesn't match ErrCircularComposable",
			sentinel: ErrComposableOutsideSetup,
			err:      ErrCircularComposable,
			expected: false,
		},
		{
			name:     "ErrCircularComposable matches itself",
			sentinel: ErrCircularComposable,
			err:      ErrCircularComposable,
			expected: true,
		},
		{
			name:     "ErrInjectNotFound matches itself",
			sentinel: ErrInjectNotFound,
			err:      ErrInjectNotFound,
			expected: true,
		},
		{
			name:     "ErrInvalidComposableState matches itself",
			sentinel: ErrInvalidComposableState,
			err:      ErrInvalidComposableState,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.sentinel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestWrappedErrors verifies that wrapped errors can still be checked
func TestWrappedErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		wrapped  error
		expected bool
	}{
		{
			name:     "Wrapped ErrComposableOutsideSetup can be detected",
			sentinel: ErrComposableOutsideSetup,
			wrapped:  fmt.Errorf("failed to execute: %w", ErrComposableOutsideSetup),
			expected: true,
		},
		{
			name:     "Wrapped ErrCircularComposable can be detected",
			sentinel: ErrCircularComposable,
			wrapped:  fmt.Errorf("composable chain error: %w", ErrCircularComposable),
			expected: true,
		},
		{
			name:     "Wrapped ErrInjectNotFound can be detected",
			sentinel: ErrInjectNotFound,
			wrapped:  fmt.Errorf("dependency injection failed: %w", ErrInjectNotFound),
			expected: true,
		},
		{
			name:     "Wrapped ErrInvalidComposableState can be detected",
			sentinel: ErrInvalidComposableState,
			wrapped:  fmt.Errorf("state validation failed: %w", ErrInvalidComposableState),
			expected: true,
		},
		{
			name:     "Wrapped error doesn't match wrong sentinel",
			sentinel: ErrComposableOutsideSetup,
			wrapped:  fmt.Errorf("some error: %w", ErrCircularComposable),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.wrapped, tt.sentinel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDoubleWrappedErrors verifies deeply nested wrapped errors work
func TestDoubleWrappedErrors(t *testing.T) {
	// Wrap error multiple times
	err1 := fmt.Errorf("level 1: %w", ErrComposableOutsideSetup)
	err2 := fmt.Errorf("level 2: %w", err1)
	err3 := fmt.Errorf("level 3: %w", err2)

	// Should still be able to detect the original sentinel error
	assert.True(t, errors.Is(err3, ErrComposableOutsideSetup))
	assert.False(t, errors.Is(err3, ErrCircularComposable))
}

// TestErrorComparison verifies errors are distinct
func TestErrorComparison(t *testing.T) {
	sentinelErrors := []error{
		ErrComposableOutsideSetup,
		ErrCircularComposable,
		ErrInjectNotFound,
		ErrInvalidComposableState,
	}

	// Each error should be unique
	for i, err1 := range sentinelErrors {
		for j, err2 := range sentinelErrors {
			if i == j {
				// Same error should equal itself
				assert.True(t, errors.Is(err1, err2),
					"Error %d should equal itself", i)
			} else {
				// Different errors should not be equal
				assert.False(t, errors.Is(err1, err2),
					"Error %d should not equal error %d", i, j)
			}
		}
	}
}

// TestErrorUsageExample demonstrates how to use these errors
func TestErrorUsageExample(t *testing.T) {
	// Example: Function that might return one of our errors
	checkComposable := func(isValid bool) error {
		if !isValid {
			return fmt.Errorf("validation failed: %w", ErrInvalidComposableState)
		}
		return nil
	}

	// Test error case
	err := checkComposable(false)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidComposableState))

	// Test success case
	err = checkComposable(true)
	assert.NoError(t, err)
}

// TestErrorSwitch demonstrates switch-based error handling
func TestErrorSwitch(t *testing.T) {
	// Simulate different error scenarios
	scenarios := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Handle outside setup error",
			err:      fmt.Errorf("call failed: %w", ErrComposableOutsideSetup),
			expected: "setup",
		},
		{
			name:     "Handle circular dependency error",
			err:      fmt.Errorf("cycle detected: %w", ErrCircularComposable),
			expected: "circular",
		},
		{
			name:     "Handle inject not found error",
			err:      fmt.Errorf("lookup failed: %w", ErrInjectNotFound),
			expected: "inject",
		},
		{
			name:     "Handle invalid state error",
			err:      fmt.Errorf("state check failed: %w", ErrInvalidComposableState),
			expected: "state",
		},
		{
			name:     "Handle unknown error",
			err:      errors.New("unknown error"),
			expected: "unknown",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			var result string

			switch {
			case errors.Is(scenario.err, ErrComposableOutsideSetup):
				result = "setup"
			case errors.Is(scenario.err, ErrCircularComposable):
				result = "circular"
			case errors.Is(scenario.err, ErrInjectNotFound):
				result = "inject"
			case errors.Is(scenario.err, ErrInvalidComposableState):
				result = "state"
			default:
				result = "unknown"
			}

			assert.Equal(t, scenario.expected, result)
		})
	}
}
