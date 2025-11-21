package testutil

import (
	"reflect"
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseAsyncTester provides utilities for testing async operations without complex timing logic.
// It integrates with the UseAsync composable to test loading states, error states, and data
// population in a deterministic way.
//
// This tester is specifically designed for testing components that use the UseAsync
// composable. It allows you to:
//   - Trigger async operations
//   - Wait for completion with timeout
//   - Check loading state
//   - Verify data and error states
//   - Test multiple executions
//
// The tester automatically extracts the async state refs from the component,
// making it easy to assert on async behavior at any point in the test.
//
// Example:
//
//	comp := createAsyncComponent() // Component using UseAsync
//	tester := NewUseAsyncTester(comp)
//
//	// Trigger async operation
//	tester.TriggerAsync()
//
//	// Wait for completion
//	tester.WaitForCompletion(t, 100*time.Millisecond)
//
//	// Verify results
//	assert.False(t, tester.IsLoading())
//	assert.NotNil(t, tester.GetData())
//	assert.Nil(t, tester.GetError())
//
// Thread Safety:
//
// UseAsyncTester is not thread-safe. It should only be used from a single test goroutine.
type UseAsyncTester struct {
	component  bubbly.Component
	dataRef    interface{} // Actual typed ref (e.g., *Ref[*string])
	loadingRef interface{} // Actual typed ref (*Ref[bool])
	errorRef   interface{} // Actual typed ref (*Ref[error])
	execute    func()
}

// NewUseAsyncTester creates a new UseAsyncTester for testing async operations.
//
// The component must expose "data", "loading", "error", and "execute" in its Setup function.
// These correspond to the fields returned by UseAsync composable.
//
// Parameters:
//   - comp: The component to test (must expose async state refs and execute function)
//
// Returns:
//   - *UseAsyncTester: A new tester instance
//
// Panics:
//   - If the component doesn't expose required refs or execute function
//
// Example:
//
//	comp, err := bubbly.NewComponent("TestAsync").
//	    Setup(func(ctx *bubbly.Context) {
//	        async := composables.UseAsync(ctx, func() (*User, error) {
//	            return fetchUser()
//	        })
//	        ctx.Expose("data", async.Data)
//	        ctx.Expose("loading", async.Loading)
//	        ctx.Expose("error", async.Error)
//	        ctx.Expose("execute", async.Execute)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseAsyncTester(comp)
func NewUseAsyncTester(comp bubbly.Component) *UseAsyncTester {
	// Extract exposed values from component using reflection
	// UseAsync exposes typed refs, not interface{} refs, so we use extractExposedValue

	// Get data ref (any type)
	dataRef := extractExposedValue(comp, "data")
	if dataRef == nil {
		panic("component must expose 'data' ref")
	}

	// Get loading ref (bool)
	loadingRef := extractExposedValue(comp, "loading")
	if loadingRef == nil {
		panic("component must expose 'loading' ref")
	}

	// Get error ref (error)
	errorRef := extractExposedValue(comp, "error")
	if errorRef == nil {
		panic("component must expose 'error' ref")
	}

	// Extract execute function
	execute, ok := extractFunctionFromComponent(comp, "execute")
	if !ok {
		panic("component must expose 'execute' function")
	}

	return &UseAsyncTester{
		component:  comp,
		dataRef:    dataRef,
		loadingRef: loadingRef,
		errorRef:   errorRef,
		execute:    execute,
	}
}

// TriggerAsync triggers the async operation by calling the execute function.
// The operation runs in a goroutine, so you should use WaitForCompletion()
// to wait for it to finish before asserting on results.
//
// Example:
//
//	tester.TriggerAsync()
//	tester.WaitForCompletion(t, 100*time.Millisecond)
//	assert.NotNil(t, tester.GetData())
func (uat *UseAsyncTester) TriggerAsync() {
	uat.execute()
}

// WaitForCompletion waits for the async operation to complete (loading becomes false).
// This method polls the loading state until it becomes false or the timeout is reached.
//
// Parameters:
//   - t: The testing.T instance for reporting timeout errors
//   - timeout: Maximum time to wait for completion
//
// Example:
//
//	tester.TriggerAsync()
//	tester.WaitForCompletion(t, 100*time.Millisecond)
//	// Now safe to assert on data/error
func (uat *UseAsyncTester) WaitForCompletion(t testing.TB, timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if !uat.IsLoading() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Errorf("async operation did not complete within %v", timeout)
}

// IsLoading returns whether an async operation is currently in progress.
//
// Returns:
//   - bool: True if loading, false otherwise
//
// Example:
//
//	assert.False(t, tester.IsLoading()) // Before trigger
//	tester.TriggerAsync()
//	time.Sleep(10 * time.Millisecond)
//	assert.True(t, tester.IsLoading()) // During execution
func (uat *UseAsyncTester) IsLoading() bool {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(uat.loadingRef)
	if !v.IsValid() || v.IsNil() {
		return false
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return false
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return false
	}

	// Convert to bool - result is interface{} containing bool
	value := result[0].Interface()
	if value == nil {
		return false
	}
	return value.(bool)
}

// GetData returns the current data value from the async operation.
// Returns nil if no data has been fetched yet or if the last operation failed.
//
// Returns:
//   - interface{}: The fetched data (type depends on async operation)
//
// Example:
//
//	tester.TriggerAsync()
//	tester.WaitForCompletion(t, 100*time.Millisecond)
//	data := tester.GetData()
//	if data != nil {
//	    user := data.(*User)
//	    assert.Equal(t, "Alice", user.Name)
//	}
func (uat *UseAsyncTester) GetData() interface{} {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(uat.dataRef)
	if !v.IsValid() || v.IsNil() {
		return nil
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return nil
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return nil
	}

	// Return the interface value
	return result[0].Interface()
}

// GetError returns the current error from the async operation.
// Returns nil if no error occurred or if the operation hasn't been executed yet.
//
// Returns:
//   - error: The error from the last async operation
//
// Example:
//
//	tester.TriggerAsync()
//	tester.WaitForCompletion(t, 100*time.Millisecond)
//	err := tester.GetError()
//	if err != nil {
//	    assert.Contains(t, err.Error(), "fetch failed")
//	}
func (uat *UseAsyncTester) GetError() error {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(uat.errorRef)
	if !v.IsValid() || v.IsNil() {
		return nil
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return nil
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return nil
	}

	// Convert to error (may be nil)
	errValue := result[0].Interface()
	if errValue == nil {
		return nil
	}

	return errValue.(error)
}
