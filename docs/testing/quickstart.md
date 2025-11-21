# Testing Quickstart

## Get Started in 5 Minutes

This guide will get you writing your first BubblyUI test in just a few minutes.

## Step 1: Create a Test File

Create a test file next to your component:

```go
// counter_test.go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)
```

## Step 2: Write Your First Test

```go
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

## Step 3: Run the Test

```bash
go test -v
```

**Output:**
```
=== RUN   TestCounterInitialState
--- PASS: TestCounterInitialState (0.00s)
PASS
ok      counter 0.001s
```

## ✅ Success!

You've written and run your first BubblyUI test!

## Next: Test State Changes

```go
func TestCounterIncrement(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Get initial value
    count := counter.State().GetRef("count")
    assert.Equal(t, 0, count.Get())
    
    // Emit increment event
    counter.Component().Emit("increment", nil)
    
    // Assert state changed
    assert.Equal(t, 1, count.Get())
}
```

## Common Patterns

### Testing Multiple Scenarios

Use table-driven tests:

```go
func TestCounterMultipleIncrements(t *testing.T) {
    tests := []struct {
        name     string
        clicks   int
        expected int
    }{
        {"single click", 1, 1},
        {"five clicks", 5, 5},
        {"ten clicks", 10, 10},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            counter := harness.Mount(createCounter())
            
            for i := 0; i < tt.clicks; i++ {
                counter.Component().Emit("increment", nil)
            }
            
            count := counter.State().GetRef("count")
            assert.Equal(t, tt.expected, count.Get())
        })
    }
}
```

### Testing Events

Track emitted events:

```go
func TestEventEmission(t *testing.T) {
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    component.Component().Emit("submit", "data")
    
    tracker.AssertFired(t, "submit")
    tracker.AssertPayload(t, "submit", "data")
}
```

### Testing Rendered Output

Use snapshot testing:

```go
func TestCounterRender(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    output := counter.Component().View()
    testutil.MatchSnapshot(t, "counter_initial", output)
}
```

## Quick Reference

### Create Test Harness
```go
harness := testutil.NewHarness(t)
```

### Mount Component
```go
component := harness.Mount(createMyComponent())
```

### Access State
```go
ref := component.State().GetRef("name")
value := ref.Get()
```

### Emit Events
```go
component.Component().Emit("eventName", payload)
```

### Track Events
```go
tracker := harness.TrackEvents()
tracker.AssertFired(t, "eventName")
```

### Snapshot Testing
```go
output := component.Component().View()
testutil.MatchSnapshot(t, "snapshot_name", output)
```

## Next Steps

- **[Assertions Reference](assertions.md)** - Complete assertion API
- **[Mocking Guide](mocking.md)** - Isolate components
- **[Snapshot Testing](snapshots.md)** - Regression testing
- **[Testing Guide](../guides/testing-guide.md)** - Comprehensive guide
- **[Examples](../../cmd/examples/)** - Working examples

## Tips

1. **Use descriptive test names**: `TestCounter_IncrementFromZero_ReturnsOne`
2. **Follow Arrange-Act-Assert**: Setup, action, verification
3. **Test one thing per test**: Keep tests focused
4. **Use helper functions**: Reduce boilerplate
5. **Run with `-race`**: Catch concurrency issues

## Common Mistakes to Avoid

### ❌ Forgetting to Expose Refs

```go
// Wrong: Ref not exposed
Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    // Missing: ctx.Expose("count", count)
})
```

```go
// Correct: Ref exposed
Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    ctx.Expose("count", count) // ✅
})
```

### ❌ Not Using Test Harness

```go
// Wrong: Manual setup
func TestManual(t *testing.T) {
    component := createComponent()
    component.Init()
    // Manual cleanup needed
}
```

```go
// Correct: Use harness
func TestWithHarness(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    // Automatic cleanup
}
```

### ❌ Testing Implementation Details

```go
// Wrong: Testing internals
func TestBad(t *testing.T) {
    component := createComponent()
    assert.Equal(t, "internal", component.internalField)
}
```

```go
// Correct: Testing behavior
func TestGood(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    component.Component().Emit("action", nil)
    result := component.State().GetRef("result")
    assert.Equal(t, "expected", result.Get())
}
```

## Summary

You now know how to:
- ✅ Create a test file
- ✅ Write basic tests
- ✅ Test state changes
- ✅ Test events
- ✅ Use snapshots
- ✅ Run tests

Start testing your components and explore the other guides for more advanced patterns!
