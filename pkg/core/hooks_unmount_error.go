package core

import (
	"errors"
	"fmt"
)

// ExecuteUnmountHooksWithErrorHandling executes all unmount hooks with error handling
func (hm *HookManager) ExecuteUnmountHooksWithErrorHandling() error {
	// Execute the unmount hooks
	err := hm.ExecuteUnmountHooks()
	if err == nil {
		return nil
	}

	// Try to find an error boundary to handle the error
	boundary := hm.FindErrorBoundary()
	if boundary == nil {
		// No error boundary found, propagate the error
		return err
	}

	// Get the error handler
	boundary.extension.mutex.RLock()
	handler := boundary.extension.errorHandler
	boundary.extension.mutex.RUnlock()

	// Call error handler if exists
	if handler != nil {
		// Unwrap any wrapped errors to get the original cause
		originalErr := errors.Unwrap(err)
		if originalErr != nil {
			// Use the original error
			handler(originalErr)
		} else {
			// No wrapped error, use as is
			handler(err)
		}
		// Error was handled by boundary
		return nil
	}

	// No handler but still an error boundary
	// Consider error handled but log it
	fmt.Printf("Warning: Error boundary %s has no handler for error: %v\n",
		boundary.componentName, err)
	return nil
}
