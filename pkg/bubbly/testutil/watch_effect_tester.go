package testutil

import (
	"reflect"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// WatchEffectTester provides utilities for testing automatic dependency tracking with WatchEffect.
// It allows you to verify that effects execute automatically when their tracked dependencies change,
// without needing to manually specify which dependencies to watch.
//
// WatchEffect automatically discovers dependencies by tracking which Refs/Computed values are accessed
// during effect execution. This tester helps verify that behavior.
//
// Key Features:
//   - Track effect execution count
//   - Trigger dependency changes
//   - Verify automatic re-execution
//   - Test conditional dependencies
//   - Verify cleanup behavior
//
// Example:
//
//	count := bubbly.NewRef(0)
//	execCount := 0
//
//	cleanup := bubbly.WatchEffect(func() {
//	    execCount++
//	    _ = count.Get() // Automatically tracks count as dependency
//	})
//	defer cleanup()
//
//	tester := NewWatchEffectTester(&execCount)
//
//	// Verify initial execution
//	tester.AssertExecuted(t, 1)
//
//	// Change dependency
//	tester.TriggerDependency(count, 5)
//
//	// Verify automatic re-execution
//	tester.AssertExecuted(t, 2)
//
// Thread Safety:
//
// WatchEffectTester is not thread-safe. It should only be used from a single test goroutine.
type WatchEffectTester struct {
	execCounter *int
	cleanup     bubbly.WatchCleanup
}

// NewWatchEffectTester creates a new WatchEffectTester for testing automatic dependency tracking.
//
// The tester requires a pointer to an execution counter that is incremented inside the
// WatchEffect function. This allows the tester to verify how many times the effect has executed.
//
// Parameters:
//   - execCounter: Pointer to an int that tracks execution count
//
// Returns:
//   - *WatchEffectTester: A new tester instance
//
// Example:
//
//	count := bubbly.NewRef(0)
//	execCount := 0
//
//	cleanup := bubbly.WatchEffect(func() {
//	    execCount++
//	    _ = count.Get()
//	})
//
//	tester := NewWatchEffectTester(&execCount)
//	tester.SetCleanup(cleanup) // Optional: allows tester to cleanup on test end
func NewWatchEffectTester(execCounter *int) *WatchEffectTester {
	return &WatchEffectTester{
		execCounter: execCounter,
	}
}

// SetCleanup sets the cleanup function returned by WatchEffect.
// This is optional but recommended to ensure proper cleanup.
//
// Parameters:
//   - cleanup: The cleanup function returned by WatchEffect
//
// Example:
//
//	cleanup := bubbly.WatchEffect(func() { ... })
//	tester.SetCleanup(cleanup)
func (wet *WatchEffectTester) SetCleanup(cleanup bubbly.WatchCleanup) {
	wet.cleanup = cleanup
}

// Cleanup calls the WatchEffect cleanup function if it was set.
// This should be called at the end of the test to stop the effect.
//
// Example:
//
//	tester := NewWatchEffectTester(&execCount)
//	defer tester.Cleanup()
func (wet *WatchEffectTester) Cleanup() {
	if wet.cleanup != nil {
		wet.cleanup()
	}
}

// TriggerDependency changes a dependency value to trigger the effect.
// This works with any Ref[T] type.
//
// Parameters:
//   - dep: The dependency to change (must be a *Ref[T])
//   - value: The new value to set
//
// Example:
//
//	count := bubbly.NewRef(0)
//	tester.TriggerDependency(count, 5)
//	tester.AssertExecuted(t, 2) // Effect should re-run
func (wet *WatchEffectTester) TriggerDependency(dep interface{}, value interface{}) {
	// Use reflection to call Set() on the ref
	v := reflect.ValueOf(dep)
	if !v.IsValid() {
		return
	}

	// Check if it's a pointer and if it's nil
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}

	// Call Set() method
	setMethod := v.MethodByName("Set")
	if !setMethod.IsValid() {
		return
	}

	// Call Set with the value
	setMethod.Call([]reflect.Value{reflect.ValueOf(value)})
}

// AssertExecuted asserts that the effect has executed the expected number of times.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: The expected execution count
//
// Example:
//
//	tester.AssertExecuted(t, 1) // Initial execution
//	tester.TriggerDependency(count, 5)
//	tester.AssertExecuted(t, 2) // Re-executed after dependency change
func (wet *WatchEffectTester) AssertExecuted(t testing.TB, expected int) {
	t.Helper()
	if wet.execCounter == nil {
		t.Fatal("execution counter is nil")
		return
	}

	actual := *wet.execCounter
	if actual != expected {
		t.Errorf("expected effect to execute %d times, but executed %d times", expected, actual)
	}
}

// GetExecutionCount returns the current execution count.
// This is useful for custom assertions or debugging.
//
// Returns:
//   - int: The current execution count
//
// Example:
//
//	count := tester.GetExecutionCount()
//	assert.Greater(t, count, 0)
func (wet *WatchEffectTester) GetExecutionCount() int {
	if wet.execCounter == nil {
		return 0
	}
	return *wet.execCounter
}
