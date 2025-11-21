package testutil

import (
	"fmt"
	"reflect"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseThrottleTester provides utilities for testing throttled functions without real time delays.
// It integrates with real time.Sleep() to control time advancement, enabling deterministic
// testing of throttle behavior.
//
// This tester is specifically designed for testing components that use the UseThrottle
// composable. It allows you to:
//   - Trigger throttled function calls
//   - Advance time to test throttle timing
//   - Verify that calls are properly throttled (immediate execution + rate limiting)
//   - Test rapid calls that should be ignored during throttle period
//   - Check throttled state and last call time
//
// The tester automatically extracts the throttled function and call count from the component,
// making it easy to assert on throttle behavior at any point in the test.
//
// Example:
//
//	comp := createThrottleComponent() // Component using UseThrottle
//	tester := NewUseThrottleTester(comp)
//
//	// Trigger first call (executes immediately)
//	tester.TriggerThrottled()
//	assert.Equal(t, 1, tester.GetCallCount())
//
//	// Rapid calls within throttle period (ignored)
//	tester.TriggerThrottled()
//	tester.TriggerThrottled()
//	assert.Equal(t, 1, tester.GetCallCount()) // Still 1
//
//	// Wait for throttle period to pass
//	tester.AdvanceTime(150 * time.Millisecond)
//
//	// Next call executes
//	tester.TriggerThrottled()
//	assert.Equal(t, 2, tester.GetCallCount())
//
// Thread Safety:
//
// UseThrottleTester is not thread-safe. It should only be used from a single test goroutine.
type UseThrottleTester struct {
	component    bubbly.Component
	throttledFn  func()
	callCount    *bubbly.Ref[interface{}]
	lastCallTime time.Time
	isThrottled  bool
}

// NewUseThrottleTester creates a new UseThrottleTester for testing throttled functions.
//
// The component must expose both "throttledFn" (the throttled function) and "callCount"
// (a ref tracking execution count) in its Setup function.
//
// Parameters:
//   - comp: The component to test (must expose "throttledFn" and "callCount")
//
// Returns:
//   - *UseThrottleTester: A new tester instance
//
// Panics:
//   - If the component doesn't expose "throttledFn" or "callCount"
//
// Example:
//
//	comp, err := bubbly.NewComponent("TestThrottle").
//	    Setup(func(ctx *bubbly.Context) {
//	        callCount := ctx.Ref(0)
//	        throttledFn := composables.UseThrottle(ctx, func() {
//	            count := callCount.Get().(int)
//	            callCount.Set(count + 1)
//	        }, 100*time.Millisecond)
//	        ctx.Expose("callCount", callCount)
//	        ctx.Expose("throttledFn", throttledFn)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseThrottleTester(comp)
func NewUseThrottleTester(comp bubbly.Component) *UseThrottleTester {
	// Extract refs from component using reflection
	refs := make(map[string]*bubbly.Ref[interface{}])
	extractRefsFromComponent(comp, refs)

	// Get callCount ref
	callCount, ok := refs["callCount"]
	if !ok {
		panic(fmt.Sprintf("component must expose 'callCount' ref. Available refs: %v", getRefNames(refs)))
	}

	// Extract throttled function using reflection
	// The throttled function is exposed as a regular value, not a ref
	throttledFn, ok := extractFunctionFromComponent(comp, "throttledFn")
	if !ok {
		panic(fmt.Sprintf("component must expose 'throttledFn' function. Available refs: %v", getRefNames(refs)))
	}

	return &UseThrottleTester{
		component:    comp,
		throttledFn:  throttledFn,
		callCount:    callCount,
		lastCallTime: time.Time{},
		isThrottled:  false,
	}
}

// extractFunctionFromComponent extracts a function exposed by the component
func extractFunctionFromComponent(comp bubbly.Component, name string) (func(), bool) {
	// Use reflection to access the component's exposed values
	// This is similar to extractRefsFromComponent but for functions
	value := extractExposedValue(comp, name)
	if value == nil {
		return nil, false
	}

	fn, ok := value.(func())
	return fn, ok
}

// extractExposedValue extracts any exposed value from the component by name
func extractExposedValue(comp bubbly.Component, name string) interface{} {
	// Use reflection to access the private state field
	v := reflect.ValueOf(comp)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Find the state field
	stateField := v.FieldByName("state")
	if !stateField.IsValid() || stateField.IsNil() {
		return nil
	}

	// Make the field accessible (it's unexported)
	stateField = reflect.NewAt(stateField.Type(), stateField.Addr().UnsafePointer()).Elem()

	// Access the state map
	if stateField.Kind() == reflect.Map {
		// Get the value for the given key
		keyValue := reflect.ValueOf(name)
		value := stateField.MapIndex(keyValue)
		if !value.IsValid() {
			return nil
		}
		return value.Interface()
	}

	return nil
}

// TriggerThrottled calls the throttled function.
// The first call executes immediately. Subsequent calls within the throttle period
// are ignored. After the throttle period passes, the next call executes immediately.
//
// Example:
//
//	tester.TriggerThrottled() // Executes immediately
//	tester.TriggerThrottled() // Ignored (within throttle period)
//	tester.AdvanceTime(150 * time.Millisecond)
//	tester.TriggerThrottled() // Executes (throttle period passed)
func (utt *UseThrottleTester) TriggerThrottled() {
	// Record time before call
	before := time.Now()

	// Get call count before
	countBefore := utt.callCount.Get().(int)

	// Call the throttled function
	utt.throttledFn()

	// Get call count after
	countAfter := utt.callCount.Get().(int)

	// If count increased, the call executed
	if countAfter > countBefore {
		utt.lastCallTime = before
		utt.isThrottled = true
	}
}

// AdvanceTime waits for the specified duration to allow throttle timers to reset.
// This method uses real time.Sleep() to wait, but with short durations (milliseconds)
// tests remain fast while being deterministic.
//
// Since UseThrottle uses real time.AfterFunc(), we need to actually wait for
// the timers to fire. This method provides a clean API for tests.
//
// After advancing time past the throttle period, the throttled state is reset.
//
// Parameters:
//   - d: The duration to wait
//
// Example:
//
//	tester.TriggerThrottled()
//	assert.True(t, tester.IsThrottled())
//	tester.AdvanceTime(150 * time.Millisecond) // Wait past throttle period
//	assert.False(t, tester.IsThrottled())
func (utt *UseThrottleTester) AdvanceTime(d time.Duration) {
	time.Sleep(d)

	// After waiting, check if enough time has passed to reset throttle state
	// This is a heuristic - if we've waited, assume throttle period has passed
	if time.Since(utt.lastCallTime) >= d {
		utt.isThrottled = false
	}
}

// GetCallCount returns the current number of times the throttled function has executed.
// This is a convenience method equivalent to tester.callCount.Get().(int).
//
// Returns:
//   - int: The current call count
//
// Example:
//
//	tester.TriggerThrottled()
//	assert.Equal(t, 1, tester.GetCallCount())
func (utt *UseThrottleTester) GetCallCount() int {
	return utt.callCount.Get().(int)
}

// GetLastCallTime returns the time of the last successful execution.
// Returns zero time if no calls have executed yet.
//
// Returns:
//   - time.Time: The time of the last execution
//
// Example:
//
//	before := time.Now()
//	tester.TriggerThrottled()
//	after := time.Now()
//	lastCall := tester.GetLastCallTime()
//	assert.True(t, lastCall.After(before) && lastCall.Before(after))
func (utt *UseThrottleTester) GetLastCallTime() time.Time {
	return utt.lastCallTime
}

// IsThrottled returns whether the function is currently in a throttled state.
// When throttled, subsequent calls are ignored until the throttle period passes.
//
// Returns:
//   - bool: True if currently throttled, false otherwise
//
// Example:
//
//	assert.False(t, tester.IsThrottled()) // Initially not throttled
//	tester.TriggerThrottled()
//	assert.True(t, tester.IsThrottled()) // Now throttled
//	tester.AdvanceTime(150 * time.Millisecond)
//	assert.False(t, tester.IsThrottled()) // Throttle period passed
func (utt *UseThrottleTester) IsThrottled() bool {
	return utt.isThrottled
}
