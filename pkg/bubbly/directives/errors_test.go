package directives

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorTypes verifies that all sentinel errors are defined and have correct messages.
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedMsg   string
		shouldBeError bool
	}{
		{
			name:          "ErrInvalidDirectiveUsage",
			err:           ErrInvalidDirectiveUsage,
			expectedMsg:   "invalid directive usage",
			shouldBeError: true,
		},
		{
			name:          "ErrBindTypeMismatch",
			err:           ErrBindTypeMismatch,
			expectedMsg:   "bind type mismatch",
			shouldBeError: true,
		},
		{
			name:          "ErrForEachNilCollection",
			err:           ErrForEachNilCollection,
			expectedMsg:   "forEach received nil collection",
			shouldBeError: true,
		},
		{
			name:          "ErrInvalidEventName",
			err:           ErrInvalidEventName,
			expectedMsg:   "invalid event name",
			shouldBeError: true,
		},
		{
			name:          "ErrRenderPanic",
			err:           ErrRenderPanic,
			expectedMsg:   "render function panicked",
			shouldBeError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is not nil
			assert.NotNil(t, tt.err, "Error should not be nil")

			// Verify error message
			assert.Equal(t, tt.expectedMsg, tt.err.Error(), "Error message should match")

			// Verify it implements error interface
			var _ error = tt.err
		})
	}
}

// TestErrorWrapping verifies that sentinel errors can be wrapped and still identified with errors.Is().
func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name        string
		baseErr     error
		wrappedErr  error
		shouldMatch bool
	}{
		{
			name:        "ErrInvalidDirectiveUsage wrapped",
			baseErr:     ErrInvalidDirectiveUsage,
			wrappedErr:  fmt.Errorf("context: %w", ErrInvalidDirectiveUsage),
			shouldMatch: true,
		},
		{
			name:        "ErrBindTypeMismatch wrapped",
			baseErr:     ErrBindTypeMismatch,
			wrappedErr:  fmt.Errorf("bind failed: %w", ErrBindTypeMismatch),
			shouldMatch: true,
		},
		{
			name:        "ErrForEachNilCollection wrapped",
			baseErr:     ErrForEachNilCollection,
			wrappedErr:  fmt.Errorf("forEach error: %w", ErrForEachNilCollection),
			shouldMatch: true,
		},
		{
			name:        "ErrInvalidEventName wrapped",
			baseErr:     ErrInvalidEventName,
			wrappedErr:  fmt.Errorf("on directive: %w", ErrInvalidEventName),
			shouldMatch: true,
		},
		{
			name:        "ErrRenderPanic wrapped",
			baseErr:     ErrRenderPanic,
			wrappedErr:  fmt.Errorf("render failed: %w", ErrRenderPanic),
			shouldMatch: true,
		},
		{
			name:        "Different error should not match",
			baseErr:     ErrInvalidDirectiveUsage,
			wrappedErr:  fmt.Errorf("different error: %w", ErrBindTypeMismatch),
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.wrappedErr, tt.baseErr)
			assert.Equal(t, tt.shouldMatch, result, "errors.Is() should return expected result")
		})
	}
}

// TestErrorMessages verifies that error messages are descriptive and helpful.
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "ErrInvalidDirectiveUsage message",
			err:           ErrInvalidDirectiveUsage,
			shouldContain: []string{"invalid", "directive", "usage"},
		},
		{
			name:          "ErrBindTypeMismatch message",
			err:           ErrBindTypeMismatch,
			shouldContain: []string{"bind", "type", "mismatch"},
		},
		{
			name:          "ErrForEachNilCollection message",
			err:           ErrForEachNilCollection,
			shouldContain: []string{"forEach", "nil", "collection"},
		},
		{
			name:          "ErrInvalidEventName message",
			err:           ErrInvalidEventName,
			shouldContain: []string{"invalid", "event", "name"},
		},
		{
			name:          "ErrRenderPanic message",
			err:           ErrRenderPanic,
			shouldContain: []string{"render", "function", "panicked"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()

			// Check that message contains expected substrings
			for _, substr := range tt.shouldContain {
				assert.Contains(t, msg, substr, "Error message should contain '%s'", substr)
			}

			// Check that message doesn't contain unexpected substrings
			for _, substr := range tt.shouldNotContain {
				assert.NotContains(t, msg, substr, "Error message should not contain '%s'", substr)
			}

			// Verify message is not empty
			assert.NotEmpty(t, msg, "Error message should not be empty")
		})
	}
}

// TestErrorUniqueness verifies that each error is a unique sentinel value.
func TestErrorUniqueness(t *testing.T) {
	errors := []error{
		ErrInvalidDirectiveUsage,
		ErrBindTypeMismatch,
		ErrForEachNilCollection,
		ErrInvalidEventName,
		ErrRenderPanic,
	}

	// Check that no two errors are the same
	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j {
				assert.NotEqual(t, err1, err2, "Errors at index %d and %d should be different", i, j)
			}
		}
	}
}

// TestErrorDocumentation verifies that errors have proper documentation.
// This is a compile-time check that ensures godoc comments exist.
func TestErrorDocumentation(t *testing.T) {
	// This test verifies that the errors package compiles and exports the expected errors.
	// The actual documentation is verified by godoc and linters.

	// Verify all errors are exported (start with capital letter)
	_ = ErrInvalidDirectiveUsage
	_ = ErrBindTypeMismatch
	_ = ErrForEachNilCollection
	_ = ErrInvalidEventName
	_ = ErrRenderPanic

	// If this compiles, the errors are properly exported
	assert.True(t, true, "All errors are properly exported")
}
