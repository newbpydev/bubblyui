package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrorHandler is a function that handles errors from hooks
type ErrorHandler func(error)

// HookManagerExtension adds additional functionality to the HookManager
type HookManagerExtension struct {
	// Context support
	context       *HookContext
	parentManager *HookManager

	// Error boundary support
	isErrorBoundary bool
	errorHandler    ErrorHandler

	// Mutex for synchronization
	mutex sync.RWMutex
}

// NewHookManagerExtension creates a new extension for hook manager
func NewHookManagerExtension() *HookManagerExtension {
	return &HookManagerExtension{
		context:         NewHookContext(),
		parentManager:   nil,
		isErrorBoundary: false,
		errorHandler:    nil,
	}
}

// Enable hook context support in HookManager
func (hm *HookManager) SetContext(ctx *HookContext) {
	// Create extension if it doesn't exist
	if hm.extension == nil {
		hm.extension = NewHookManagerExtension()
	}

	hm.extension.mutex.Lock()
	defer hm.extension.mutex.Unlock()
	hm.extension.context = ctx
}

// GetContext returns the hook context for this hook manager
// If no context has been set, it creates a new one
func (hm *HookManager) GetContext() *HookContext {
	// Create extension if it doesn't exist
	if hm.extension == nil {
		hm.extension = NewHookManagerExtension()
		return hm.extension.context
	}

	hm.extension.mutex.RLock()
	defer hm.extension.mutex.RUnlock()

	if hm.extension.context == nil {
		// Create context with locked mutex
		hm.extension.mutex.RUnlock()
		hm.extension.mutex.Lock()
		defer hm.extension.mutex.Unlock()

		// Double check after acquiring write lock
		if hm.extension.context == nil {
			hm.extension.context = NewHookContext()
		}
		return hm.extension.context
	}

	return hm.extension.context
}

// SetParentManager establishes a parent-child relationship between hook managers
// This allows for context inheritance and error boundary propagation
func (hm *HookManager) SetParentManager(parent *HookManager) {
	// Create extension if it doesn't exist
	if hm.extension == nil {
		hm.extension = NewHookManagerExtension()
	}

	hm.extension.mutex.Lock()
	defer hm.extension.mutex.Unlock()
	hm.extension.parentManager = parent

	// If we have a context, link it to parent's context
	if hm.extension.context != nil && parent != nil {
		parentCtx := parent.GetContext()
		hm.extension.context.SetParent(parentCtx)
	}
}

// SetAsErrorBoundary configures this hook manager to act as an error boundary
// An error boundary will catch errors from hooks and handle them locally
func (hm *HookManager) SetAsErrorBoundary(isErrorBoundary bool) {
	// Create extension if it doesn't exist
	if hm.extension == nil {
		hm.extension = NewHookManagerExtension()
	}

	hm.extension.mutex.Lock()
	defer hm.extension.mutex.Unlock()
	hm.extension.isErrorBoundary = isErrorBoundary
}

// SetErrorHandler sets the function that will be called when an error occurs
// in a hook and this manager is an error boundary
func (hm *HookManager) SetErrorHandler(handler ErrorHandler) {
	// Create extension if it doesn't exist
	if hm.extension == nil {
		hm.extension = NewHookManagerExtension()
	}

	hm.extension.mutex.Lock()
	defer hm.extension.mutex.Unlock()
	hm.extension.errorHandler = handler
}

// IsErrorBoundary returns whether this hook manager is an error boundary
func (hm *HookManager) IsErrorBoundary() bool {
	if hm.extension == nil {
		return false
	}

	hm.extension.mutex.RLock()
	defer hm.extension.mutex.RUnlock()
	return hm.extension.isErrorBoundary
}

// FindErrorBoundary finds the nearest error boundary in the hook manager hierarchy
// Returns nil if no error boundary is found
func (hm *HookManager) FindErrorBoundary() *HookManager {
	// Check if this hook manager is an error boundary
	if hm.IsErrorBoundary() {
		return hm
	}

	// Check parent hierarchy
	if hm.extension != nil && hm.extension.parentManager != nil {
		return hm.extension.parentManager.FindErrorBoundary()
	}

	// No error boundary found
	return nil
}

// HandleError handles an error according to error boundary configuration
// Returns true if the error was handled, false otherwise
func (hm *HookManager) HandleError(err error) bool {
	// Find the nearest error boundary
	boundary := hm.FindErrorBoundary()
	if boundary == nil {
		// No error boundary, error is not handled
		return false
	}

	// Call the error handler if it exists
	boundary.extension.mutex.RLock()
	handler := boundary.extension.errorHandler
	boundary.extension.mutex.RUnlock()

	if handler != nil {
		handler(err)
	}

	// Error is considered handled by the boundary
	return true
}

// OnMountWithContext registers a hook with access to context
func (hm *HookManager) OnMountWithContext(callback func(ctx *HookContext) error) HookID {
	// Wrap the context-aware callback in a standard callback
	return hm.OnMount(func() error {
		return callback(hm.GetContext())
	})
}

// OnUpdateWithContext registers an update hook with access to context
func (hm *HookManager) OnUpdateWithContext(
	callback func(ctx *HookContext, prevDeps []interface{}) error,
	deps []interface{}) HookID {
	// Wrap the context-aware callback in a standard callback
	return hm.OnUpdate(func(prevDeps []interface{}) error {
		return callback(hm.GetContext(), prevDeps)
	}, deps)
}

// OnUnmountWithContext registers an unmount hook with access to context
func (hm *HookManager) OnUnmountWithContext(callback func(ctx *HookContext) error) HookID {
	// Wrap the context-aware callback in a standard callback
	return hm.OnUnmount(func() error {
		return callback(hm.GetContext())
	})
}

// ExecuteUnmountHooksWithTimeout executes all unmount hooks with a timeout
// If the timeout is exceeded, an error is returned
func (hm *HookManager) ExecuteUnmountHooksWithTimeout(timeout time.Duration) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return hm.ExecuteUnmountHooksWithContext(ctx)
}

// ExecuteUnmountHooksWithContext executes all unmount hooks with context control
// The context can be used for timeout, cancellation, etc.
func (hm *HookManager) ExecuteUnmountHooksWithContext(ctx context.Context) error {
	hm.mutex.RLock()
	hooks := make([]*OnUnmountHook, 0, len(hm.unmountHooks))
	for _, hook := range hm.unmountHooks {
		hooks = append(hooks, hook)
	}
	hm.mutex.RUnlock()

	var allErrs []string

	// Execute each hook with context awareness
	for _, hook := range hooks {
		// Create a channel to signal completion
		done := make(chan error, 1)

		// Run the hook in a goroutine
		go func(h *OnUnmountHook) {
			done <- h.Execute()
		}(hook)

		// Wait for either completion or context cancellation
		select {
		case err := <-done:
			if err != nil {
				allErrs = append(allErrs, err.Error())
			}
		case <-ctx.Done():
			allErrs = append(allErrs, fmt.Sprintf("hook %s timed out: %v", hook.ID(), ctx.Err()))
			// Return immediately on context cancellation
			return fmt.Errorf("timeout: unmount hook %s execution exceeded time limit", hook.ID())
		}
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("errors during unmount hooks execution:\n%s", strings.Join(allErrs, "\n"))
	}

	return nil
}

// ExecuteUnmountHooksWithContextParallel executes all unmount hooks in parallel with context control
// This can be faster but may be less predictable than serial execution
func (hm *HookManager) ExecuteUnmountHooksWithContextParallel(ctx context.Context) error {
	hm.mutex.RLock()
	hooks := make([]*OnUnmountHook, 0, len(hm.unmountHooks))
	for _, hook := range hm.unmountHooks {
		hooks = append(hooks, hook)
	}
	hm.mutex.RUnlock()

	if len(hooks) == 0 {
		return nil
	}

	// Create error collection
	errChan := make(chan error, len(hooks))
	wg := sync.WaitGroup{}
	wg.Add(len(hooks))

	// Run all hooks in parallel
	for _, hook := range hooks {
		go func(h *OnUnmountHook) {
			defer wg.Done()

			// Create a channel for hook completion
			done := make(chan error, 1)

			// Execute the hook
			go func() {
				done <- h.Execute()
			}()

			// Wait for either completion or timeout
			select {
			case err := <-done:
				if err != nil {
					errChan <- fmt.Errorf("hook %s error: %w", h.ID(), err)
				}
			case <-ctx.Done():
				errChan <- fmt.Errorf("hook %s timed out: %w", h.ID(), ctx.Err())
			}
		}(hook)
	}

	// Wait for all hooks to complete or timeout
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var allErrs []string
	for err := range errChan {
		allErrs = append(allErrs, err.Error())
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("errors during parallel unmount hooks execution:\n%s", strings.Join(allErrs, "\n"))
	}

	return nil
}

// ExecuteMountHooksWithErrorHandling executes mount hooks with error boundary support
// If an error occurs and is caught by an error boundary, it returns nil
func (hm *HookManager) ExecuteMountHooksWithErrorHandling() error {
	// First attempt to execute hooks normally
	err := hm.ExecuteMountHooks()
	if err == nil {
		// No error occurred, nothing to handle
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

// ExecuteUpdateHooksWithErrorHandling executes update hooks with error boundary support
// If an error occurs and is caught by an error boundary, it returns nil
func (hm *HookManager) ExecuteUpdateHooksWithErrorHandling() error {
	// First attempt to execute hooks normally
	err := hm.ExecuteUpdateHooks()
	if err == nil {
		// No error occurred, nothing to handle
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
