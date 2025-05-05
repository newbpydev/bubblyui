package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorBoundary(t *testing.T) {
	t.Run("Basic Error Boundary", func(t *testing.T) {
		// Create hook managers
		parentHM := NewHookManager("parent")
		childHM := NewHookManager("child")

		// Set parent as error boundary
		parentHM.SetAsErrorBoundary(true)

		// Track error handling
		parentCaughtError := false
		capturedError := error(nil)

		// Set up parent-child relationship
		childHM.SetParentManager(parentHM)

		// Set up error handler in parent
		parentHM.SetErrorHandler(func(err error) {
			parentCaughtError = true
			capturedError = err
		})

		// Add hook that fails in child
		expectedError := errors.New("child hook error")
		childHM.OnMount(func() error {
			return expectedError
		})

		// Execute child's hooks with error handling
		err := childHM.ExecuteMountHooksWithErrorHandling()

		// Error should be caught by boundary
		assert.NoError(t, err, "Error should be caught by boundary")
		assert.True(t, parentCaughtError, "Parent error boundary should have caught error")
		assert.Equal(t, expectedError, capturedError, "Captured error should match original error")
	})

	t.Run("Nested Error Boundaries", func(t *testing.T) {
		// Create hook managers with nested boundaries
		rootHM := NewHookManager("root")
		middleHM := NewHookManager("middle")
		leafHM := NewHookManager("leaf")

		// Set up hierarchy
		middleHM.SetParentManager(rootHM)
		leafHM.SetParentManager(middleHM)

		// Set up error boundaries
		rootHM.SetAsErrorBoundary(true)
		middleHM.SetAsErrorBoundary(true)

		// Track error handling at each level
		rootCaughtError := false
		middleCaughtError := false

		// Set up error handlers
		rootHM.SetErrorHandler(func(err error) {
			rootCaughtError = true
		})

		middleHM.SetErrorHandler(func(err error) {
			middleCaughtError = true
		})

		// Add hook that fails in leaf
		leafHM.OnMount(func() error {
			return errors.New("leaf error")
		})

		// Execute leaf's hooks with error handling
		err := leafHM.ExecuteMountHooksWithErrorHandling()

		// Error should be caught by middle boundary (closest ancestor)
		assert.NoError(t, err, "Error should be caught by boundary")
		assert.True(t, middleCaughtError, "Middle error boundary should have caught error")
		assert.False(t, rootCaughtError, "Root error boundary should not have caught error")
	})
}

func TestUnmountHookTimeout(t *testing.T) {
	t.Run("Successful Execution Within Timeout", func(t *testing.T) {
		// Create hook manager
		hm := NewHookManager("test-component")

		// Add hook that completes quickly
		hm.OnUnmount(func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})

		// Execute with timeout
		err := hm.ExecuteUnmountHooksWithTimeout(100 * time.Millisecond)
		assert.NoError(t, err, "Hook execution should complete within timeout")
	})

	t.Run("Timeout Exceeded", func(t *testing.T) {
		// Create hook manager
		hm := NewHookManager("test-component")

		// Add hook that takes too long
		hm.OnUnmount(func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})

		// Execute with short timeout
		err := hm.ExecuteUnmountHooksWithTimeout(50 * time.Millisecond)

		// Should return timeout error
		assert.Error(t, err, "ExecuteUnmountHooksWithTimeout should return error when timeout exceeded")
		assert.Contains(t, err.Error(), "timeout", "Error should indicate timeout")
	})

	t.Run("Parallel Hook Execution", func(t *testing.T) {
		// Skip if testing in short mode
		if testing.Short() {
			t.Skip("skipping parallel hook test in short mode")
		}

		// Create hook manager
		hm := NewHookManager("test-component")

		// Add multiple hooks with varying execution times
		hookCount := 5
		for i := 0; i < hookCount; i++ {
			i := i // Capture loop variable
			hm.OnUnmount(func() error {
				// Stagger execution times
				time.Sleep(time.Duration(i*20) * time.Millisecond)
				return nil
			})
		}

		// Execute parallel with longer timeout to allow all to complete
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := hm.ExecuteUnmountHooksWithContextParallel(ctx)
		elapsed := time.Since(start)

		// All hooks should complete
		assert.NoError(t, err, "All hooks should complete within timeout")

		// Parallel execution should be faster than sequential (which would be ~200ms)
		assert.Less(t, elapsed, 150*time.Millisecond, "Parallel execution should be faster than sequential")
	})
}

func TestGuaranteedCleanup(t *testing.T) {
	t.Run("Cleanup After Panic", func(t *testing.T) {
		// Create hook manager
		hm := NewHookManager("test-component")

		// Track cleanup execution
		cleanupRun := false

		// Add unmount hook for cleanup
		hm.OnUnmount(func() error {
			cleanupRun = true
			return nil
		})

		// Execute in a recover block
		func() {
			defer func() {
				recover() // Catch the panic

				// Cleanup should still run despite panic
				err := hm.ExecuteUnmountHooks()
				assert.NoError(t, err, "ExecuteUnmountHooks should not error")
				assert.True(t, cleanupRun, "Cleanup hook should have run after panic")
			}()

			// Simulate code that panics
			panic("simulated panic")
		}()
	})

	t.Run("All Cleanup Errors Collected", func(t *testing.T) {
		// Create hook manager
		hm := NewHookManager("test-component")

		// Add multiple cleanup hooks that fail
		hm.OnUnmount(func() error {
			return errors.New("error 1")
		})

		hm.OnUnmount(func() error {
			return errors.New("error 2")
		})

		hm.OnUnmount(func() error {
			return errors.New("error 3")
		})

		// Execute all cleanup hooks
		err := hm.ExecuteUnmountHooks()

		// Should have error containing all error messages
		assert.Error(t, err, "ExecuteUnmountHooks should return error when hooks fail")
		errMsg := err.Error()
		assert.Contains(t, errMsg, "error 1", "Error should contain first error message")
		assert.Contains(t, errMsg, "error 2", "Error should contain second error message")
		assert.Contains(t, errMsg, "error 3", "Error should contain third error message")
	})
}
