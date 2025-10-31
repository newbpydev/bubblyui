package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// UseAsyncReturn is the return type for the UseAsync composable.
// It provides reactive state management for asynchronous operations with
// automatic loading and error state tracking.
//
// Fields:
//   - Data: Reactive reference to the fetched data (nil until fetch succeeds)
//   - Loading: Reactive boolean indicating if fetch is in progress
//   - Error: Reactive reference to any error that occurred during fetch
//   - Execute: Function to trigger the async operation
//   - Reset: Function to clear all state back to initial values
//
// Example:
//
//	async := UseAsync(ctx, func() (*User, error) {
//	    return fetchUser()
//	})
//
//	// Trigger fetch on mount
//	ctx.OnMounted(func() {
//	    async.Execute()
//	})
//
//	// Access reactive state
//	ctx.Expose("user", async.Data)
//	ctx.Expose("loading", async.Loading)
//	ctx.Expose("error", async.Error)
type UseAsyncReturn[T any] struct {
	// Data holds the result of the async operation.
	// It is nil initially and until the first successful fetch.
	Data *bubbly.Ref[*T]

	// Loading indicates whether an async operation is currently in progress.
	// True while fetching, false otherwise.
	Loading *bubbly.Ref[bool]

	// Error holds any error that occurred during the last fetch attempt.
	// Nil if the last fetch was successful or no fetch has been attempted.
	Error *bubbly.Ref[error]

	// Execute triggers the async operation.
	// Can be called multiple times. Each call starts a new fetch operation.
	// Sets Loading to true, clears Error, executes fetcher in a goroutine,
	// and updates Data/Error/Loading when complete.
	Execute func()

	// Reset clears all state back to initial values.
	// Sets Data to nil, Loading to false, and Error to nil.
	// Does not cancel any in-flight operations.
	Reset func()
}

// UseAsync creates a composable for managing asynchronous data fetching operations.
// It handles loading states, error states, and data population automatically,
// providing a clean API for async operations in components.
//
// The fetcher function is executed in a goroutine when Execute() is called.
// State updates (Data, Loading, Error) are performed on the reactive refs,
// triggering reactivity throughout the component tree.
//
// UseAsync is type-safe using Go generics. The type parameter T specifies
// the type of data being fetched.
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - fetcher: Async function that returns data or an error
//
// Returns:
//   - UseAsyncReturn[T]: Struct with reactive state and control functions
//
// Example - Basic Usage:
//
//	Setup(func(ctx *Context) {
//	    userData := UseAsync(ctx, func() (*User, error) {
//	        return fetchUserFromAPI()
//	    })
//
//	    // Trigger fetch when component mounts
//	    ctx.OnMounted(func() {
//	        userData.Execute()
//	    })
//
//	    // Expose state to template
//	    ctx.Expose("user", userData.Data)
//	    ctx.Expose("loading", userData.Loading)
//	    ctx.Expose("error", userData.Error)
//	})
//
// Example - With Error Handling:
//
//	Setup(func(ctx *Context) {
//	    async := UseAsync(ctx, fetchData)
//
//	    ctx.OnMounted(func() {
//	        async.Execute()
//	    })
//
//	    // Watch for errors
//	    ctx.Watch(async.Error, func(newErr, _ error) {
//	        if newErr != nil {
//	            log.Printf("Fetch failed: %v", newErr)
//	        }
//	    })
//	})
//
// Example - Manual Retry:
//
//	Setup(func(ctx *Context) {
//	    async := UseAsync(ctx, fetchData)
//
//	    ctx.On("retry", func(_ interface{}) {
//	        async.Execute() // Retry the fetch
//	    })
//
//	    ctx.On("reset", func(_ interface{}) {
//	        async.Reset() // Clear all state
//	    })
//	})
//
// Concurrency:
//
// Multiple concurrent Execute() calls are safe. Each call spawns a new goroutine
// and updates the shared reactive state. The last operation to complete will
// set the final state values.
//
// Note: UseAsync does not cancel in-flight operations. If you need cancellation,
// consider using context.Context in your fetcher function.
//
// Performance:
//
// UseAsync creates three Ref instances and two closure functions. The overhead
// is minimal (< 1μs) and well within the performance target for composables.
func UseAsync[T any](ctx *bubbly.Context, fetcher func() (*T, error)) UseAsyncReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseAsync", time.Since(start))
	}()

	// Create reactive state for data, loading, and error
	data := bubbly.NewRef[*T](nil)
	loading := bubbly.NewRef(false)
	errorRef := bubbly.NewRef[error](nil)

	// Execute function: triggers the async operation
	execute := func() {
		// Set loading state
		loading.Set(true)
		errorRef.Set(nil)

		// Execute fetcher in goroutine
		go func() {
			result, err := fetcher()

			// Update state based on result
			if err != nil {
				errorRef.Set(err)
				data.Set(nil)
			} else {
				data.Set(result)
				errorRef.Set(nil)
			}

			// Clear loading state
			loading.Set(false)
		}()
	}

	// Reset function: clears all state
	reset := func() {
		data.Set(nil)
		loading.Set(false)
		errorRef.Set(nil)
	}

	// Return the composable interface
	return UseAsyncReturn[T]{
		Data:    data,
		Loading: loading,
		Error:   errorRef,
		Execute: execute,
		Reset:   reset,
	}
}
