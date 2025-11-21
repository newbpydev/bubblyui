package testutil

import (
	"fmt"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseDebounceTester provides utilities for testing debounced values without real time delays.
// It integrates with TimeSimulator to control time advancement, enabling deterministic
// testing of debounce behavior.
//
// This tester is specifically designed for testing components that use the UseDebounce
// composable. It allows you to:
//   - Trigger value changes without waiting for real delays
//   - Advance simulated time to test debounce timing
//   - Verify that debounced values update correctly
//   - Test rapid changes that should cancel previous timers
//
// The tester automatically extracts the debounced ref from the component, making it
// easy to assert on the debounced value at any point in the test.
//
// Example:
//
//	ts := NewTimeSimulator()
//	comp := createDebounceComponent() // Component using UseDebounce
//	tester := NewUseDebounceTester(comp, ts)
//
//	// Trigger change
//	tester.TriggerChange("new value")
//
//	// Advance time past debounce delay
//	tester.AdvanceTime(300 * time.Millisecond)
//
//	// Assert debounced value updated
//	assert.Equal(t, "new value", tester.debounced.Get())
//
// Thread Safety:
//
// UseDebounceTester is not thread-safe. It should only be used from a single test goroutine.
type UseDebounceTester struct {
	timeSim   *TimeSimulator
	component bubbly.Component
	debounced *bubbly.Ref[interface{}]
	source    *bubbly.Ref[interface{}]
}

// NewUseDebounceTester creates a new UseDebounceTester for testing debounced values.
//
// The component must expose both "source" and "debounced" refs in its Setup function.
// The tester will extract these refs and use them for testing.
//
// Parameters:
//   - comp: The component to test (must expose "source" and "debounced" refs)
//   - timeSim: The TimeSimulator to use for controlling time
//
// Returns:
//   - *UseDebounceTester: A new tester instance
//
// Panics:
//   - If the component doesn't expose "source" or "debounced" refs
//
// Example:
//
//	ts := NewTimeSimulator()
//	comp := bubbly.NewBuilder().
//	    WithSetup(func(ctx *bubbly.Context) {
//	        source := ctx.Ref("initial")
//	        debounced := composables.UseDebounce(ctx, source, 300*time.Millisecond)
//	        ctx.Expose("source", source)
//	        ctx.Expose("debounced", debounced)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseDebounceTester(comp, ts)
func NewUseDebounceTester(comp bubbly.Component, timeSim *TimeSimulator) *UseDebounceTester {
	// Extract refs from component using reflection
	refs := make(map[string]*bubbly.Ref[interface{}])
	extractRefsFromComponent(comp, refs)

	// Get source ref
	source, ok := refs["source"]
	if !ok {
		panic(fmt.Sprintf("component must expose 'source' ref. Available refs: %v", getRefNames(refs)))
	}

	// Get debounced ref
	debounced, ok := refs["debounced"]
	if !ok {
		panic(fmt.Sprintf("component must expose 'debounced' ref. Available refs: %v", getRefNames(refs)))
	}

	return &UseDebounceTester{
		timeSim:   timeSim,
		component: comp,
		debounced: debounced,
		source:    source,
	}
}

// getRefNames returns a list of ref names for error messages
func getRefNames(refs map[string]*bubbly.Ref[interface{}]) []string {
	names := make([]string, 0, len(refs))
	for name := range refs {
		names = append(names, name)
	}
	return names
}

// TriggerChange triggers a change to the source value.
// This simulates user input or any other event that would change the source value.
//
// The debounced value will not update immediately - you must call AdvanceTime()
// to advance the simulated time past the debounce delay.
//
// Parameters:
//   - value: The new value to set on the source ref
//
// Example:
//
//	tester.TriggerChange("new value")
//	tester.AdvanceTime(300 * time.Millisecond)
//	assert.Equal(t, "new value", tester.debounced.Get())
func (udt *UseDebounceTester) TriggerChange(value interface{}) {
	udt.source.Set(value)
}

// AdvanceTime waits for the specified duration to allow debounce timers to fire.
// This method uses real time.Sleep() to wait, but with short durations (milliseconds)
// tests remain fast while being deterministic.
//
// Since UseDebounce uses real time.AfterFunc(), we need to actually wait for
// the timers to fire. This method provides a clean API for tests.
//
// Parameters:
//   - d: The duration to wait
//
// Example:
//
//	tester.TriggerChange("value1")
//	tester.AdvanceTime(100 * time.Millisecond) // Wait 100ms - not enough
//	assert.Equal(t, "initial", tester.debounced.Get())
//
//	tester.AdvanceTime(250 * time.Millisecond) // Wait another 250ms - now past 300ms delay
//	assert.Equal(t, "value1", tester.debounced.Get())
func (udt *UseDebounceTester) AdvanceTime(d time.Duration) {
	time.Sleep(d)
}

// GetDebouncedValue returns the current value of the debounced ref.
// This is a convenience method equivalent to tester.debounced.Get().
//
// Returns:
//   - interface{}: The current debounced value
//
// Example:
//
//	value := tester.GetDebouncedValue()
//	assert.Equal(t, "expected", value.(string))
func (udt *UseDebounceTester) GetDebouncedValue() interface{} {
	return udt.debounced.Get()
}

// GetSourceValue returns the current value of the source ref.
// This is a convenience method equivalent to tester.source.Get().
//
// Returns:
//   - interface{}: The current source value
//
// Example:
//
//	value := tester.GetSourceValue()
//	assert.Equal(t, "expected", value.(string))
func (udt *UseDebounceTester) GetSourceValue() interface{} {
	return udt.source.Get()
}
