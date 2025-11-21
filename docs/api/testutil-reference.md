# Testing Utilities API Reference

## Overview

Complete API reference for the BubblyUI testing utilities package (`pkg/bubbly/testutil`). This reference covers all types, functions, and methods available for testing BubblyUI components.

## Table of Contents

- [Test Harness](#test-harness)
- [Component Testing](#component-testing)
- [State Inspection](#state-inspection)
- [Event Testing](#event-testing)
- [Watch Testing](#watch-testing)
- [Snapshot Testing](#snapshot-testing)
- [Mock Utilities](#mock-utilities)
- [Async Testing](#async-testing)
- [Advanced Testing](#advanced-testing)

## Test Harness

### NewHarness

Creates a new test harness for component testing.

```go
func NewHarness(t *testing.T, opts ...HarnessOption) *TestHarness
```

**Parameters:**
- `t` - Testing context
- `opts` - Optional configuration options

**Returns:** `*TestHarness`

**Example:**
```go
harness := testutil.NewHarness(t)
```

### TestHarness Methods

#### Mount

Mounts a component in the test environment.

```go
func (h *TestHarness) Mount(component Component, props ...interface{}) *ComponentTest
```

**Parameters:**
- `component` - Component to mount
- `props` - Optional props (reserved for future use)

**Returns:** `*ComponentTest`

#### TrackEvents

Creates an event tracker for monitoring emitted events.

```go
func (h *TestHarness) TrackEvents() *EventTracker
```

**Returns:** `*EventTracker`

#### Cleanup

Manually triggers cleanup (usually automatic via `t.Cleanup()`).

```go
func (h *TestHarness) Cleanup()
```

#### RegisterCleanup

Registers a cleanup function to run on test completion.

```go
func (h *TestHarness) RegisterCleanup(fn func())
```

## Component Testing

### ComponentTest

Represents a mounted component in the test environment.

```go
type ComponentTest struct {
    // Unexported fields
}
```

#### Component

Returns the underlying component.

```go
func (ct *ComponentTest) Component() bubbly.Component
```

**Returns:** `bubbly.Component`

#### State

Returns the state inspector for accessing refs, computed values, and watchers.

```go
func (ct *ComponentTest) State() *StateInspector
```

**Returns:** `*StateInspector`

#### Events

Returns the event inspector for tracking events.

```go
func (ct *ComponentTest) Events() *EventInspector
```

**Returns:** `*EventInspector`

#### Unmount

Unmounts the component and triggers cleanup.

```go
func (ct *ComponentTest) Unmount()
```

## State Inspection

### StateInspector

Provides access to component state (refs, computed values, watchers).

```go
type StateInspector struct {
    // Unexported fields
}
```

#### GetRef

Retrieves a ref by name.

```go
func (si *StateInspector) GetRef(name string) *bubbly.Ref[interface{}]
```

**Parameters:**
- `name` - Ref name

**Returns:** `*bubbly.Ref[interface{}]` or `nil` if not found

**Panics:** If ref doesn't exist (use `HasRef` to check first)

#### GetRefValue

Retrieves a ref's value by name.

```go
func (si *StateInspector) GetRefValue(name string) interface{}
```

**Parameters:**
- `name` - Ref name

**Returns:** Current ref value

#### SetRefValue

Sets a ref's value by name.

```go
func (si *StateInspector) SetRefValue(name string, value interface{})
```

**Parameters:**
- `name` - Ref name
- `value` - New value

#### HasRef

Checks if a ref exists.

```go
func (si *StateInspector) HasRef(name string) bool
```

**Parameters:**
- `name` - Ref name

**Returns:** `true` if ref exists

#### GetComputed

Retrieves a computed value by name.

```go
func (si *StateInspector) GetComputed(name string) *bubbly.Computed[interface{}]
```

**Parameters:**
- `name` - Computed name

**Returns:** `*bubbly.Computed[interface{}]` or `nil` if not found

#### GetComputedValue

Retrieves a computed value's current value by name.

```go
func (si *StateInspector) GetComputedValue(name string) interface{}
```

**Parameters:**
- `name` - Computed name

**Returns:** Current computed value

#### HasComputed

Checks if a computed value exists.

```go
func (si *StateInspector) HasComputed(name string) bool
```

**Parameters:**
- `name` - Computed name

**Returns:** `true` if computed exists

#### GetWatcher

Retrieves a watcher cleanup function by name.

```go
func (si *StateInspector) GetWatcher(name string) bubbly.WatchCleanup
```

**Parameters:**
- `name` - Watcher name

**Returns:** Cleanup function or `nil` if not found

#### HasWatcher

Checks if a watcher exists.

```go
func (si *StateInspector) HasWatcher(name string) bool
```

**Parameters:**
- `name` - Watcher name

**Returns:** `true` if watcher exists

## Event Testing

### EventTracker

Tracks emitted events for testing.

```go
type EventTracker struct {
    // Unexported fields
}
```

#### AssertFired

Asserts that an event was fired.

```go
func (et *EventTracker) AssertFired(t testingT, eventName string)
```

**Parameters:**
- `t` - Testing context
- `eventName` - Event name to check

#### AssertNotFired

Asserts that an event was not fired.

```go
func (et *EventTracker) AssertNotFired(t testingT, eventName string)
```

#### AssertPayload

Asserts event payload matches expected value.

```go
func (et *EventTracker) AssertPayload(t testingT, eventName string, expected interface{})
```

**Parameters:**
- `t` - Testing context
- `eventName` - Event name
- `expected` - Expected payload

#### AssertOrder

Asserts events were fired in the specified order.

```go
func (et *EventTracker) AssertOrder(t testingT, eventNames []string)
```

**Parameters:**
- `t` - Testing context
- `eventNames` - Expected event order

#### GetEvents

Returns all tracked events.

```go
func (et *EventTracker) GetEvents() []EmittedEvent
```

**Returns:** Slice of `EmittedEvent`

#### Reset

Clears all tracked events.

```go
func (et *EventTracker) Reset()
```

### EmittedEvent

Represents a captured event emission.

```go
type EmittedEvent struct {
    Name      string
    Payload   interface{}
    Timestamp time.Time
}
```

## Watch Testing

### WatchTester

Tests watcher execution.

```go
type WatchTester struct {
    // Unexported fields
}
```

#### NewWatchTester

Creates a new watch tester.

```go
func NewWatchTester(callback func()) *WatchTester
```

**Parameters:**
- `callback` - Callback to track

**Returns:** `*WatchTester`

#### Watch

Watches a ref and tracks callback invocations.

```go
func (wt *WatchTester) Watch(ref *bubbly.Ref[interface{}])
```

**Parameters:**
- `ref` - Ref to watch

#### AssertCallCount

Asserts callback was called expected number of times.

```go
func (wt *WatchTester) AssertCallCount(t testingT, expected int)
```

**Parameters:**
- `t` - Testing context
- `expected` - Expected call count

#### GetCallCount

Returns the number of times callback was called.

```go
func (wt *WatchTester) GetCallCount() int
```

**Returns:** Call count

#### Reset

Resets the call counter.

```go
func (wt *WatchTester) Reset()
```

## Snapshot Testing

### MatchSnapshot

Matches rendered output against saved snapshot.

```go
func MatchSnapshot(t *testing.T, name string, output string)
```

**Parameters:**
- `t` - Testing context
- `name` - Snapshot name
- `output` - Rendered output to compare

**Behavior:**
- First run: Creates snapshot file
- Subsequent runs: Compares against saved snapshot
- With `-update` flag: Updates snapshot file

**Example:**
```go
output := component.View()
testutil.MatchSnapshot(t, "counter_initial", output)
```

## Mock Utilities

### MockRef

Mock implementation of `Ref[T]` for testing.

```go
type MockRef[T any] struct {
    // Unexported fields
}
```

#### NewMockRef

Creates a new mock ref.

```go
func NewMockRef[T any](initial T) *MockRef[T]
```

**Parameters:**
- `initial` - Initial value

**Returns:** `*MockRef[T]`

#### AssertGetCalls

Asserts `Get()` was called expected number of times.

```go
func (mr *MockRef[T]) AssertGetCalls(t testingT, expected int)
```

#### AssertSetCalls

Asserts `Set()` was called expected number of times.

```go
func (mr *MockRef[T]) AssertSetCalls(t testingT, expected int)
```

#### AssertValue

Asserts current value matches expected.

```go
func (mr *MockRef[T]) AssertValue(t testingT, expected T)
```

## Async Testing

### WaitFor

Waits for a condition to become true with timeout protection.

```go
func WaitFor(t *testing.T, condition func() bool, opts WaitOptions) bool
```

**Parameters:**
- `t` - Testing context
- `condition` - Function returning `true` when condition met
- `opts` - Wait options

**Returns:** `true` if condition met, `false` if timeout

**Example:**
```go
testutil.WaitFor(t, func() bool {
    return loading.Get().(bool) == false
}, testutil.WaitOptions{
    Timeout: 5 * time.Second,
    Message: "data to load",
})
```

### WaitOptions

Configuration for `WaitFor`.

```go
type WaitOptions struct {
    Timeout  time.Duration // Max wait time
    Interval time.Duration // Poll interval (default: 10ms)
    Message  string        // Error message on timeout
}
```

### TimeSimulator

Simulates time passage for testing time-dependent code.

```go
type TimeSimulator struct {
    // Unexported fields
}
```

#### NewTimeSimulator

Creates a new time simulator.

```go
func NewTimeSimulator() *TimeSimulator
```

**Returns:** `*TimeSimulator`

#### Advance

Advances simulated time by the specified duration.

```go
func (ts *TimeSimulator) Advance(d time.Duration)
```

**Parameters:**
- `d` - Duration to advance

## Advanced Testing

### DeepWatchTester

Tests deep watching of objects.

```go
type DeepWatchTester struct {
    // Unexported fields
}
```

#### NewDeepWatchTester

Creates a new deep watch tester.

```go
func NewDeepWatchTester() *DeepWatchTester
```

#### WatchDeep

Watches an object deeply.

```go
func (dwt *DeepWatchTester) WatchDeep(ref *bubbly.Ref[interface{}], callback func())
```

#### AssertCallCount

Asserts callback was called expected number of times.

```go
func (dwt *DeepWatchTester) AssertCallCount(t testingT, expected int)
```

### ComputedCacheVerifier

Verifies computed value caching behavior.

```go
type ComputedCacheVerifier struct {
    // Unexported fields
}
```

#### NewComputedCacheVerifier

Creates a new cache verifier.

```go
func NewComputedCacheVerifier(computed interface{}, computeCount *int) *ComputedCacheVerifier
```

**Parameters:**
- `computed` - Computed value to verify
- `computeCount` - Pointer to external counter

**Returns:** `*ComputedCacheVerifier`

#### GetValue

Gets the computed value (tracks cache hits/misses).

```go
func (ccv *ComputedCacheVerifier) GetValue() interface{}
```

#### AssertComputeCount

Asserts compute function was called expected number of times.

```go
func (ccv *ComputedCacheVerifier) AssertComputeCount(t *testing.T, expected int)
```

#### AssertCacheHits

Asserts number of cache hits.

```go
func (ccv *ComputedCacheVerifier) AssertCacheHits(t *testing.T, expected int)
```

#### AssertCacheMisses

Asserts number of cache misses.

```go
func (ccv *ComputedCacheVerifier) AssertCacheMisses(t *testing.T, expected int)
```

### MockErrorReporter

Mock implementation of error reporter for testing observability.

```go
type MockErrorReporter struct {
    // Unexported fields
}
```

#### NewMockErrorReporter

Creates a new mock error reporter.

```go
func NewMockErrorReporter() *MockErrorReporter
```

#### AssertPanicReported

Asserts a panic was reported.

```go
func (mer *MockErrorReporter) AssertPanicReported(t testingT)
```

#### GetContexts

Returns all error contexts.

```go
func (mer *MockErrorReporter) GetContexts() []*observability.ErrorContext
```

## Type Aliases

### testingT

Interface for testing context (compatible with `*testing.T` and `*testing.B`).

```go
type testingT interface {
    Helper()
    Errorf(format string, args ...interface{})
    FailNow()
}
```

## Best Practices

1. **Always use `t.Helper()`** in assertion functions
2. **Provide clear error messages** in assertions
3. **Use descriptive snapshot names**
4. **Clean up resources** (automatic with harness)
5. **Test behavior, not implementation**
6. **Use table-driven tests** for multiple scenarios
7. **Mock external dependencies**
8. **Use `WaitFor` for async operations**

## Migration Guide

### From Manual Testing

**Before:**
```go
func TestManual(t *testing.T) {
    component := createComponent()
    component.Init()
    // Manual state access
    // Manual cleanup
}
```

**After:**
```go
func TestWithHarness(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    // Automatic init and cleanup
}
```

## See Also

- **[Testing Guide](../guides/testing-guide.md)** - Getting started with testing
- **[Advanced Testing](../guides/advanced-testing.md)** - Advanced patterns
- **[Examples](../../cmd/examples/)** - Working examples
- **[Code Conventions](../code-conventions.md)** - BubblyUI standards

## Summary

The testing utilities API provides:

- ✅ **Test Harness** - Component mounting and lifecycle
- ✅ **State Inspection** - Access refs, computed, watchers
- ✅ **Event Tracking** - Monitor and assert events
- ✅ **Watch Testing** - Verify watcher execution
- ✅ **Snapshot Testing** - Regression testing
- ✅ **Mock Utilities** - Isolation and control
- ✅ **Async Testing** - Timeout protection
- ✅ **Advanced Testing** - Deep watch, caching, observability

All utilities follow Go testing conventions and integrate seamlessly with testify assertions.
