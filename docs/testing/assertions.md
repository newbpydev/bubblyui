# Assertion Reference

## Overview

Complete reference for all assertion methods available in the BubblyUI testing framework. This guide covers state assertions, event assertions, and custom matchers.

## Table of Contents

- [State Assertions](#state-assertions)
- [Event Assertions](#event-assertions)
- [Lifecycle Assertions](#lifecycle-assertions)
- [Async Assertions](#async-assertions)
- [Custom Assertions](#custom-assertions)
- [Best Practices](#best-practices)

## State Assertions

### Ref Assertions

#### GetRef

Get a ref by name:

```go
ref := component.State().GetRef("count")
```

**Returns**: `*bubbly.Ref[interface{}]` or `nil`

**Panics**: If ref doesn't exist

#### GetRefValue

Get a ref's value directly:

```go
value := component.State().GetRefValue("count")
assert.Equal(t, 42, value)
```

#### SetRefValue

Set a ref's value for testing:

```go
component.State().SetRefValue("count", 100)
```

#### HasRef

Check if a ref exists:

```go
if component.State().HasRef("count") {
    // Ref exists
}
```

**Example:**
```go
func TestRefAssertions(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Check existence
    assert.True(t, component.State().HasRef("count"))
    
    // Get and assert value
    count := component.State().GetRef("count")
    assert.Equal(t, 0, count.Get())
    
    // Set and verify
    count.Set(42)
    assert.Equal(t, 42, component.State().GetRefValue("count"))
}
```

### Computed Value Assertions

#### GetComputed

Get a computed value by name:

```go
computed := component.State().GetComputed("doubled")
```

**Returns**: `*bubbly.Computed[interface{}]` or `nil`

#### GetComputedValue

Get a computed value's current value:

```go
value := component.State().GetComputedValue("doubled")
assert.Equal(t, 20, value)
```

#### HasComputed

Check if a computed value exists:

```go
if component.State().HasComputed("doubled") {
    // Computed exists
}
```

**Example:**
```go
func TestComputedAssertions(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Check existence
    assert.True(t, component.State().HasComputed("doubled"))
    
    // Get computed value
    doubled := component.State().GetComputed("doubled")
    assert.Equal(t, 0, doubled.Get())
    
    // Change dependency
    component.State().SetRefValue("count", 10)
    
    // Verify computed updated
    assert.Equal(t, 20, component.State().GetComputedValue("doubled"))
}
```

### Watcher Assertions

#### GetWatcher

Get a watcher cleanup function:

```go
cleanup := component.State().GetWatcher("dataWatcher")
```

**Returns**: `bubbly.WatchCleanup` or `nil`

#### HasWatcher

Check if a watcher exists:

```go
if component.State().HasWatcher("dataWatcher") {
    // Watcher exists
}
```

**Example:**
```go
func TestWatcherAssertions(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Check watcher exists
    assert.True(t, component.State().HasWatcher("dataWatcher"))
    
    // Get cleanup function
    cleanup := component.State().GetWatcher("dataWatcher")
    assert.NotNil(t, cleanup)
    
    // Call cleanup
    cleanup()
    
    // Watcher should be removed
    assert.False(t, component.State().HasWatcher("dataWatcher"))
}
```

## Event Assertions

### EventTracker

Track and assert on emitted events.

#### AssertFired

Assert an event was fired:

```go
tracker.AssertFired(t, "submit")
```

**Parameters:**
- `t` - Testing context
- `eventName` - Event name to check

**Example:**
```go
func TestEventFired(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    component.Component().Emit("submit", nil)
    
    tracker.AssertFired(t, "submit")
}
```

#### AssertNotFired

Assert an event was not fired:

```go
tracker.AssertNotFired(t, "cancel")
```

**Example:**
```go
func TestEventNotFired(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    component.Component().Emit("submit", nil)
    
    tracker.AssertNotFired(t, "cancel")
}
```

#### AssertPayload

Assert event payload matches expected value:

```go
tracker.AssertPayload(t, "submit", expectedData)
```

**Parameters:**
- `t` - Testing context
- `eventName` - Event name
- `expected` - Expected payload

**Example:**
```go
func TestEventPayload(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    
    data := map[string]interface{}{
        "id":   123,
        "name": "Test",
    }
    
    component.Component().Emit("update", data)
    
    tracker.AssertPayload(t, "update", data)
}
```

#### AssertOrder

Assert events were fired in specific order:

```go
tracker.AssertOrder(t, []string{"start", "process", "complete"})
```

**Parameters:**
- `t` - Testing context
- `eventNames` - Expected event order

**Example:**
```go
func TestEventOrder(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    
    component.Component().Emit("start", nil)
    component.Component().Emit("process", nil)
    component.Component().Emit("complete", nil)
    
    tracker.AssertOrder(t, []string{"start", "process", "complete"})
}
```

#### GetEvents

Get all tracked events:

```go
events := tracker.GetEvents()
assert.Len(t, events, 3)
```

**Returns**: `[]EmittedEvent`

**Example:**
```go
func TestGetEvents(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    component.Component().Emit("event1", nil)
    component.Component().Emit("event2", nil)
    
    events := tracker.GetEvents()
    assert.Len(t, events, 2)
    assert.Equal(t, "event1", events[0].Name)
    assert.Equal(t, "event2", events[1].Name)
}
```

#### Reset

Clear all tracked events:

```go
tracker.Reset()
```

**Example:**
```go
func TestEventReset(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    component.Component().Emit("event", nil)
    
    assert.Len(t, tracker.GetEvents(), 1)
    
    tracker.Reset()
    
    assert.Len(t, tracker.GetEvents(), 0)
}
```

## Lifecycle Assertions

### Testing Mount Behavior

```go
func TestOnMounted(t *testing.T) {
    mounted := false
    
    component := bubbly.NewComponent().
        Setup(func(ctx *bubbly.Context) {
            ctx.OnMounted(func() {
                mounted = true
            })
        })
    
    harness := testutil.NewHarness(t)
    harness.Mount(component)
    
    assert.True(t, mounted)
}
```

### Testing Update Behavior

```go
func TestOnUpdated(t *testing.T) {
    updateCount := 0
    
    component := bubbly.NewComponent().
        Setup(func(ctx *bubbly.Context) {
            count := ctx.Ref(0)
            
            ctx.OnUpdated(func() {
                updateCount++
            })
            
            ctx.Expose("count", count)
        })
    
    harness := testutil.NewHarness(t)
    mounted := harness.Mount(component)
    
    // Trigger update
    mounted.State().SetRefValue("count", 1)
    
    assert.Greater(t, updateCount, 0)
}
```

### Testing Cleanup

```go
func TestOnUnmounted(t *testing.T) {
    cleaned := false
    
    component := bubbly.NewComponent().
        Setup(func(ctx *bubbly.Context) {
            ctx.OnUnmounted(func() {
                cleaned = true
            })
        })
    
    harness := testutil.NewHarness(t)
    mounted := harness.Mount(component)
    
    mounted.Unmount()
    
    assert.True(t, cleaned)
}
```

## Async Assertions

### WaitFor

Wait for a condition with timeout:

```go
testutil.WaitFor(t, func() bool {
    return condition()
}, testutil.WaitOptions{
    Timeout: 5 * time.Second,
    Message: "condition to be true",
})
```

**Parameters:**
- `t` - Testing context
- `condition` - Function returning `true` when met
- `opts` - Wait options

**Example:**
```go
func TestAsyncOperation(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAsyncComponent())
    
    component.Component().Emit("fetch", nil)
    
    testutil.WaitFor(t, func() bool {
        loading := component.State().GetRef("loading")
        return !loading.Get().(bool)
    }, testutil.WaitOptions{
        Timeout:  5 * time.Second,
        Interval: 10 * time.Millisecond,
        Message:  "data to finish loading",
    })
    
    data := component.State().GetRef("data")
    assert.NotNil(t, data.Get())
}
```

## Custom Assertions

### Creating Custom Assertions

```go
func assertCountEquals(t *testing.T, component *testutil.ComponentTest, expected int) {
    t.Helper()
    
    count := component.State().GetRef("count")
    actual := count.Get().(int)
    
    if actual != expected {
        t.Errorf("count = %d, want %d", actual, expected)
    }
}

func TestCustomAssertion(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createCounter())
    
    assertCountEquals(t, component, 0)
    
    component.Component().Emit("increment", nil)
    
    assertCountEquals(t, component, 1)
}
```

### Custom Matchers

```go
type countMatcher struct {
    expected int
}

func (m *countMatcher) Match(component *testutil.ComponentTest) bool {
    count := component.State().GetRef("count")
    return count.Get().(int) == m.expected
}

func (m *countMatcher) FailureMessage() string {
    return fmt.Sprintf("expected count to be %d", m.expected)
}

func HaveCount(expected int) *countMatcher {
    return &countMatcher{expected: expected}
}

func TestCustomMatcher(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createCounter())
    
    matcher := HaveCount(0)
    assert.True(t, matcher.Match(component), matcher.FailureMessage())
}
```

## Best Practices

### 1. Use t.Helper()

Always use `t.Helper()` in custom assertion functions:

```go
func assertValid(t *testing.T, component *testutil.ComponentTest) {
    t.Helper() // ✅ Correct line numbers in failures
    
    valid := component.State().GetRef("valid")
    assert.True(t, valid.Get().(bool))
}
```

### 2. Provide Clear Error Messages

```go
// ✅ Good: Clear message
assert.Equal(t, expected, actual, "count should be incremented")

// ❌ Bad: No message
assert.Equal(t, expected, actual)
```

### 3. Test One Thing Per Assertion

```go
// ✅ Good: Separate assertions
assert.Equal(t, "John", name.Get())
assert.Equal(t, 30, age.Get())

// ❌ Bad: Combined assertion
assert.True(t, name.Get() == "John" && age.Get() == 30)
```

### 4. Use Specific Assertions

```go
// ✅ Good: Specific assertion
assert.NotNil(t, data.Get())

// ❌ Bad: Generic assertion
assert.True(t, data.Get() != nil)
```

### 5. Assert on Behavior, Not Implementation

```go
// ✅ Good: Test behavior
component.Component().Emit("submit", data)
tracker.AssertFired(t, "submit")

// ❌ Bad: Test implementation
assert.Equal(t, 1, component.handlerCallCount)
```

## Summary

The assertion API provides:

- ✅ **State Assertions** - Refs, computed, watchers
- ✅ **Event Assertions** - Fired, payload, order
- ✅ **Lifecycle Assertions** - Mount, update, unmount
- ✅ **Async Assertions** - WaitFor with timeout
- ✅ **Custom Assertions** - Build your own

Use these assertions to write clear, maintainable tests that verify component behavior.

## See Also

- **[Quickstart](quickstart.md)** - Get started quickly
- **[Mocking Guide](mocking.md)** - Isolation patterns
- **[Snapshot Testing](snapshots.md)** - Regression testing
- **[API Reference](../api/testutil-reference.md)** - Complete API
