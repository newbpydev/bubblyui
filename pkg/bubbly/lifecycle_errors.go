package bubbly

import "errors"

// Lifecycle hook error types.
// These sentinel errors are used to identify specific error conditions
// in the lifecycle system. They follow Go best practices for error handling
// using errors.New() for simple sentinel errors.
var (
	// ErrHookPanic is returned when a lifecycle hook execution panics.
	// The panic is recovered and reported to the observability system,
	// allowing other hooks to continue executing.
	//
	// Example:
	//	ctx.OnMounted(func() {
	//	    panic("something went wrong")
	//	})
	//	// Component continues working, error reported to observability
	ErrHookPanic = errors.New("hook execution panicked")

	// ErrCleanupFailed is returned when a cleanup function fails.
	// The cleanup failure is recovered and reported to the observability system,
	// allowing other cleanup functions to continue executing.
	//
	// Example:
	//	ctx.OnCleanup(func() {
	//	    panic("cleanup failed")
	//	})
	//	// Other cleanups still execute, error reported to observability
	ErrCleanupFailed = errors.New("cleanup function failed")

	// ErrMaxUpdateDepth is returned when the maximum update depth is exceeded.
	// This typically indicates an infinite loop where onUpdated hooks
	// continuously trigger more updates.
	//
	// Example:
	//	ctx.OnUpdated(func() {
	//	    count.Set(count.GetTyped() + 1) // Infinite loop!
	//	})
	//	// After 100 iterations, ErrMaxUpdateDepth is returned
	ErrMaxUpdateDepth = errors.New("max update depth exceeded")
)
