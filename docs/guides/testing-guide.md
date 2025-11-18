# Testing Guide

## Overview

BubblyUI provides a comprehensive testing framework that makes it easy to write reliable tests for your TUI components. This guide covers everything from basic component testing to advanced patterns for testing reactive state, events, lifecycle hooks, and more.

## Table of Contents

- [Getting Started](#getting-started)
- [Basic Component Testing](#basic-component-testing)
- [State Testing](#state-testing)
- [Event Testing](#event-testing)
- [Lifecycle Testing](#lifecycle-testing)
- [Snapshot Testing](#snapshot-testing)
- [Table-Driven Tests](#table-driven-tests)
- [Async Testing](#async-testing)
- [Mocking](#mocking)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

## Getting Started

### Installation

The testing utilities are included in the BubblyUI package:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)
```

### Your First Test

Create a test file next to your component:

```go
// counter_test.go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

func TestCounterInitialState(t *testing.T) {
    // Create test harness
    harness := testutil.NewHarness(t)
    
    // Mount component
    counter := harness.Mount(createCounter())
    
    // Assert initial state
    count := counter.State().GetRef("count")
    assert.Equal(t, 0, count.Get())
}
```

Run your test:

```bash
go test -v
```

## Basic Component Testing

### Mounting Components

The test harness provides a clean environment for testing components:

```go
func TestComponent(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createMyComponent())
    
    // Component is now mounted and initialized
    // Cleanup happens automatically via t.Cleanup()
}
```

### Accessing Component State

Use the `State()` inspector to access refs, computed values, and watchers:

```go
func TestComponentState(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Get a ref
    name := component.State().GetRef("name")
    assert.Equal(t, "John", name.Get())
    
    // Get a computed value
    fullName := component.State().GetComputed("fullName")
    assert.Equal(t, "John Doe", fullName.Get())
}
```

### Simulating Events

Emit events to test event handlers:

```go
func TestEventHandling(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Emit an event
    component.Component().Emit("submit", "test data")
    
    // Assert state changed
    result := component.State().GetRef("result")
    assert.Equal(t, "test data", result.Get())
}
```

## State Testing

### Testing Refs

Test reactive references with type-safe assertions:

```go
func TestRefUpdates(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    count := component.State().GetRef("count")
    
    // Initial value
    assert.Equal(t, 0, count.Get())
    
    // Update value
    count.Set(42)
    assert.Equal(t, 42, count.Get())
}
```

### Testing Computed Values

Verify computed values update correctly:

```go
func TestComputedValues(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    count := component.State().GetRef("count")
    doubled := component.State().GetComputed("doubled")
    
    // Initial computed value
    assert.Equal(t, 0, doubled.Get())
    
    // Update dependency
    count.Set(5)
    
    // Computed value updates
    assert.Equal(t, 10, doubled.Get())
}
```

### Testing Watchers

Verify watchers execute correctly:

```go
func TestWatchers(t *testing.T) {
    harness := testutil.NewHarness(t)
    
    callCount := 0
    watcher := testutil.NewWatchTester(func() {
        callCount++
    })
    
    count := harness.Ref(0)
    watcher.Watch(count)
    
    // Change value
    count.Set(1)
    watcher.AssertCallCount(t, 1)
    
    count.Set(2)
    watcher.AssertCallCount(t, 2)
}
```

## Event Testing

### Tracking Events

Use the event tracker to verify events are emitted:

```go
func TestEventEmission(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    
    // Trigger action
    component.Component().Emit("click", nil)
    
    // Verify event was fired
    tracker.AssertFired(t, "click")
    tracker.AssertPayload(t, "click", nil)
}
```

### Event Order

Verify events fire in the correct order:

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

### Event Payloads

Test event data is passed correctly:

```go
func TestEventPayloads(t *testing.T) {
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

## Lifecycle Testing

### Testing Mount Behavior

Verify components initialize correctly:

```go
func TestOnMounted(t *testing.T) {
    harness := testutil.NewHarness(t)
    
    mounted := false
    component := bubbly.NewComponent().
        Setup(func(ctx *bubbly.Context) {
            ctx.OnMounted(func() {
                mounted = true
            })
        })
    
    harness.Mount(component)
    
    // onMounted should have executed
    assert.True(t, mounted)
}
```

### Testing Update Behavior

Verify update hooks execute:

```go
func TestOnUpdated(t *testing.T) {
    harness := testutil.NewHarness(t)
    
    updateCount := 0
    component := bubbly.NewComponent().
        Setup(func(ctx *bubbly.Context) {
            count := ctx.Ref(0)
            
            ctx.OnUpdated(func() {
                updateCount++
            })
            
            ctx.Expose("count", count)
        })
    
    mounted := harness.Mount(component)
    
    // Trigger update
    count := mounted.State().GetRef("count")
    count.Set(1)
    
    assert.Greater(t, updateCount, 0)
}
```

### Testing Cleanup

Verify cleanup functions execute on unmount:

```go
func TestCleanup(t *testing.T) {
    harness := testutil.NewHarness(t)
    
    cleaned := false
    component := bubbly.NewComponent().
        Setup(func(ctx *bubbly.Context) {
            ctx.OnUnmounted(func() {
                cleaned = true
            })
        })
    
    mounted := harness.Mount(component)
    mounted.Unmount()
    
    assert.True(t, cleaned)
}
```

## Snapshot Testing

### Creating Snapshots

Capture rendered output for regression testing:

```go
func TestCounterRender(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Take snapshot
    output := counter.Component().View()
    testutil.MatchSnapshot(t, "counter_initial", output)
}
```

### Updating Snapshots

When output intentionally changes, update snapshots:

```bash
go test -update
```

### Snapshot Best Practices

1. **Use descriptive names**: `testutil.MatchSnapshot(t, "counter_after_increment", output)`
2. **Test stable output**: Avoid timestamps or random data
3. **Review changes**: Always review snapshot diffs before updating
4. **Small snapshots**: Test specific parts, not entire screens

## Table-Driven Tests

### Basic Pattern

Test multiple scenarios efficiently:

```go
func TestCounterIncrement(t *testing.T) {
    tests := []struct {
        name     string
        initial  int
        clicks   int
        expected int
    }{
        {"zero to one", 0, 1, 1},
        {"one to five", 1, 4, 5},
        {"large increment", 0, 100, 100},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            counter := harness.Mount(createCounterWithValue(tt.initial))
            
            for i := 0; i < tt.clicks; i++ {
                counter.Component().Emit("increment", nil)
            }
            
            count := counter.State().GetRef("count")
            assert.Equal(t, tt.expected, count.Get())
        })
    }
}
```

### Parallel Execution

Run independent tests in parallel:

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        value int
    }{
        {"test 1", 1},
        {"test 2", 2},
        {"test 3", 3},
    }
    
    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run in parallel
            
            harness := testutil.NewHarness(t)
            component := harness.Mount(createComponent(tt.value))
            
            // Test implementation
        })
    }
}
```

## Async Testing

### Waiting for Conditions

Test async operations with timeout protection:

```go
func TestAsyncDataLoad(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createDataComponent())
    
    // Trigger async operation
    component.Component().Emit("fetch-data", nil)
    
    // Wait for completion
    testutil.WaitFor(t, func() bool {
        loading := component.State().GetRef("loading")
        return loading.Get().(bool) == false
    }, testutil.WaitOptions{
        Timeout: 5 * time.Second,
        Message: "data to finish loading",
    })
    
    // Assert data loaded
    data := component.State().GetRef("data")
    assert.NotNil(t, data.Get())
}
```

### Testing Debounced Operations

Use time simulation for deterministic testing:

```go
func TestDebounce(t *testing.T) {
    harness := testutil.NewHarness(t)
    timeSim := testutil.NewTimeSimulator()
    
    component := harness.MountWithTime(createDebounceComponent(), timeSim)
    
    input := component.State().GetRef("input")
    input.Set("test")
    
    // Not debounced yet
    debounced := component.State().GetRef("debounced")
    assert.Equal(t, "", debounced.Get())
    
    // Advance time
    timeSim.Advance(300 * time.Millisecond)
    
    // Now debounced
    assert.Equal(t, "test", debounced.Get())
}
```

## Mocking

### Mock Refs

Create mock refs for testing:

```go
func TestWithMockRef(t *testing.T) {
    mockRef := testutil.NewMockRef(42)
    
    // Track calls
    mockRef.Get()
    mockRef.Set(100)
    
    // Assert usage
    mockRef.AssertGetCalls(t, 1)
    mockRef.AssertSetCalls(t, 1)
    mockRef.AssertValue(t, 100)
}
```

### Mock Components

Mock child components for isolation:

```go
func TestWithMockChild(t *testing.T) {
    mockChild := testutil.NewMockComponent()
    mockChild.SetView("Mock Child View")
    
    parent := createParentWithChild(mockChild)
    
    harness := testutil.NewHarness(t)
    mounted := harness.Mount(parent)
    
    output := mounted.Component().View()
    assert.Contains(t, output, "Mock Child View")
}
```

## Best Practices

### 1. Test Behavior, Not Implementation

```go
// ❌ Bad: Testing implementation details
func TestBad(t *testing.T) {
    component := createComponent()
    // Don't test internal state directly
    assert.Equal(t, "internal", component.internalField)
}

// ✅ Good: Testing behavior
func TestGood(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Test observable behavior
    component.Component().Emit("action", nil)
    result := component.State().GetRef("result")
    assert.Equal(t, "expected", result.Get())
}
```

### 2. Use Helper Functions

```go
// Helper function reduces boilerplate
func mountCounter(t *testing.T, initial int) *testutil.ComponentTest {
    harness := testutil.NewHarness(t)
    return harness.Mount(createCounterWithValue(initial))
}

func TestWithHelper(t *testing.T) {
    counter := mountCounter(t, 0)
    // Test implementation
}
```

### 3. Clear Test Names

```go
// ✅ Good: Descriptive names
func TestCounter_IncrementFromZero_ReturnsOne(t *testing.T) { }
func TestCounter_DecrementBelowZero_StaysAtZero(t *testing.T) { }

// ❌ Bad: Vague names
func TestCounter1(t *testing.T) { }
func TestCounter2(t *testing.T) { }
```

### 4. Arrange-Act-Assert Pattern

```go
func TestPattern(t *testing.T) {
    // Arrange
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Act
    component.Component().Emit("action", "data")
    
    // Assert
    result := component.State().GetRef("result")
    assert.Equal(t, "expected", result.Get())
}
```

### 5. Test Edge Cases

```go
func TestEdgeCases(t *testing.T) {
    tests := []struct {
        name  string
        input interface{}
    }{
        {"nil input", nil},
        {"empty string", ""},
        {"zero value", 0},
        {"negative value", -1},
        {"large value", 999999},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test each edge case
        })
    }
}
```

## Common Patterns

### Pattern 1: Setup/Teardown

```go
func TestWithSetup(t *testing.T) {
    // Setup
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    // Teardown happens automatically via t.Cleanup()
    
    // Test implementation
}
```

### Pattern 2: Subtests

```go
func TestCounter(t *testing.T) {
    t.Run("initial state", func(t *testing.T) {
        // Test initial state
    })
    
    t.Run("increment", func(t *testing.T) {
        // Test increment
    })
    
    t.Run("decrement", func(t *testing.T) {
        // Test decrement
    })
}
```

### Pattern 3: Fixtures

```go
func setupComponent(t *testing.T, opts ...Option) *testutil.ComponentTest {
    harness := testutil.NewHarness(t)
    component := createComponent()
    
    for _, opt := range opts {
        opt(component)
    }
    
    return harness.Mount(component)
}

func TestWithFixture(t *testing.T) {
    component := setupComponent(t, WithTitle("Test"))
    // Test implementation
}
```

## Troubleshooting

### Test Timeout

**Problem**: Test hangs and times out

**Solution**: Use `WaitFor` with timeout:

```go
testutil.WaitFor(t, func() bool {
    return condition()
}, testutil.WaitOptions{
    Timeout: 5 * time.Second,
    Message: "condition to be true",
})
```

### Ref Not Found

**Problem**: `panic: ref "name" not found`

**Solution**: Verify ref is exposed in Setup:

```go
Setup(func(ctx *Context) {
    name := ctx.Ref("John")
    ctx.Expose("name", name) // Must expose!
})
```

### Snapshot Mismatch

**Problem**: Snapshot test fails after intentional change

**Solution**: Update snapshots:

```bash
go test -update
```

### Race Conditions

**Problem**: Tests fail with `-race` flag

**Solution**: Ensure proper synchronization:

```go
// Use mutex for shared state
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
// Access shared state
```

## Next Steps

- **[Advanced Testing Guide](advanced-testing.md)** - Commands, composables, directives, router
- **[API Reference](../api/testutil-reference.md)** - Complete API documentation
- **[Examples](../../cmd/examples/)** - Working examples for all features
- **[Best Practices](../code-conventions.md)** - BubblyUI coding standards

## Summary

The BubblyUI testing framework provides:

- ✅ **Zero-boilerplate** component mounting
- ✅ **Type-safe** state inspection
- ✅ **Event tracking** and assertions
- ✅ **Snapshot testing** with easy updates
- ✅ **Async testing** without flakiness
- ✅ **Mock utilities** for isolation
- ✅ **Table-driven** test support
- ✅ **Fast execution** (< 1s per test)

Start with basic component tests, then explore advanced patterns as your needs grow. The testing utilities integrate seamlessly with Go's built-in testing package and testify assertions.

Happy testing! 🧪
