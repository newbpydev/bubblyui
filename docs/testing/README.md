# Testing in BubblyUI

## Overview

BubblyUI provides a comprehensive testing framework that makes it easy to write reliable tests for your TUI components. This documentation covers everything you need to test your BubblyUI applications effectively.

## Documentation Structure

This testing documentation is organized into focused guides:

- **[Quickstart](quickstart.md)** - Get started with testing in 5 minutes
- **[Assertions](assertions.md)** - Complete assertion reference
- **[Mocking](mocking.md)** - Mocking and isolation patterns
- **[Snapshots](snapshots.md)** - Snapshot testing guide
- **[Best Practices](#best-practices)** - Testing best practices and patterns

## What is the BubblyUI Testing Framework?

The testing framework provides:

- ✅ **Test Harness** - Clean environment for component testing
- ✅ **State Inspection** - Access refs, computed values, watchers
- ✅ **Event Tracking** - Monitor and assert on events
- ✅ **Snapshot Testing** - Regression testing for rendered output
- ✅ **Mock Utilities** - Isolate components from dependencies
- ✅ **Async Testing** - Timeout protection for async operations
- ✅ **Type Safety** - Full Go generics support

## Quick Example

```go
func TestCounter(t *testing.T) {
    // Create test harness
    harness := testutil.NewHarness(t)
    
    // Mount component
    counter := harness.Mount(createCounter())
    
    // Assert initial state
    count := counter.State().GetRef("count")
    assert.Equal(t, 0, count.Get())
    
    // Simulate event
    counter.Component().Emit("increment", nil)
    
    // Assert state changed
    assert.Equal(t, 1, count.Get())
}
```

## Installation

The testing utilities are included in the BubblyUI package:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)
```

## Core Concepts

### Test Harness

The test harness provides a clean, isolated environment for testing components:

```go
harness := testutil.NewHarness(t)
component := harness.Mount(createMyComponent())
// Automatic cleanup via t.Cleanup()
```

### State Inspection

Access component state through the state inspector:

```go
state := component.State()

// Get refs
name := state.GetRef("name")

// Get computed values
fullName := state.GetComputed("fullName")

// Check watchers
hasWatcher := state.HasWatcher("dataWatcher")
```

### Event Tracking

Track and assert on emitted events:

```go
tracker := harness.TrackEvents()
component.Component().Emit("submit", data)

tracker.AssertFired(t, "submit")
tracker.AssertPayload(t, "submit", data)
```

### Snapshot Testing

Capture and compare rendered output:

```go
output := component.Component().View()
testutil.MatchSnapshot(t, "component_initial", output)
```

## Testing Patterns

### Table-Driven Tests

Test multiple scenarios efficiently:

```go
func TestScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"zero", 0, 0},
        {"positive", 5, 10},
        {"negative", -5, -10},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            component := harness.Mount(createComponent(tt.input))
            
            result := component.State().GetRef("result")
            assert.Equal(t, tt.expected, result.Get())
        })
    }
}
```

### Async Testing

Test async operations with timeout protection:

```go
func TestAsync(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAsyncComponent())
    
    component.Component().Emit("fetch", nil)
    
    testutil.WaitFor(t, func() bool {
        loading := component.State().GetRef("loading")
        return !loading.Get().(bool)
    }, testutil.WaitOptions{
        Timeout: 5 * time.Second,
        Message: "data to load",
    })
}
```

## Best Practices

### 1. Test Behavior, Not Implementation

```go
// ✅ Good: Test observable behavior
func TestGood(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createComponent())
    
    component.Component().Emit("action", nil)
    result := component.State().GetRef("result")
    assert.Equal(t, "expected", result.Get())
}

// ❌ Bad: Test internal details
func TestBad(t *testing.T) {
    component := createComponent()
    assert.Equal(t, "internal", component.internalField)
}
```

### 2. Use Descriptive Test Names

```go
// ✅ Good: Clear, descriptive names
func TestCounter_IncrementFromZero_ReturnsOne(t *testing.T) { }

// ❌ Bad: Vague names
func TestCounter1(t *testing.T) { }
```

### 3. Arrange-Act-Assert Pattern

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

### 4. Test Edge Cases

Always test:
- Nil/empty inputs
- Zero values
- Negative values
- Large values
- Boundary conditions

### 5. Keep Tests Fast

- Use mocks for external dependencies
- Avoid real network calls
- Use time simulation for delays
- Run tests in parallel when possible

## TDD Workflow

### Red-Green-Refactor

1. **Red**: Write a failing test
2. **Green**: Write minimal code to pass
3. **Refactor**: Improve code while keeping tests green

```go
// 1. Red: Write failing test
func TestCounter_Increment_IncreasesCount(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    counter.Component().Emit("increment", nil)
    
    count := counter.State().GetRef("count")
    assert.Equal(t, 1, count.Get()) // Fails initially
}

// 2. Green: Implement feature
// ... implement increment handler ...

// 3. Refactor: Improve code
// ... refactor while tests stay green ...
```

## Running Tests

### Basic Test Run

```bash
go test -v
```

### With Race Detector

```bash
go test -race -v
```

### With Coverage

```bash
go test -cover -v
```

### Update Snapshots

```bash
go test -update
```

### Parallel Tests

```bash
go test -parallel 4
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
})
```

### Ref Not Found

**Problem**: `panic: ref "name" not found`

**Solution**: Ensure ref is exposed in Setup:

```go
Setup(func(ctx *Context) {
    name := ctx.Ref("John")
    ctx.Expose("name", name) // Must expose!
})
```

### Race Conditions

**Problem**: Tests fail with `-race` flag

**Solution**: Use proper synchronization or avoid shared state

## Next Steps

- **[Quickstart Guide](quickstart.md)** - Get started in 5 minutes
- **[Assertions Reference](assertions.md)** - Complete assertion API
- **[Mocking Guide](mocking.md)** - Isolation and mocking patterns
- **[Snapshot Testing](snapshots.md)** - Regression testing guide
- **[Advanced Testing](../guides/advanced-testing.md)** - Advanced patterns
- **[API Reference](../api/testutil-reference.md)** - Complete API docs

## Additional Resources

- **[Testing Guide](../guides/testing-guide.md)** - Comprehensive testing guide
- **[Examples](../../cmd/examples/)** - Working test examples
- **[Code Conventions](../code-conventions.md)** - BubblyUI standards

## Summary

The BubblyUI testing framework makes it easy to write reliable, maintainable tests for your TUI applications. Start with the [Quickstart Guide](quickstart.md) and explore the focused documentation for specific topics.

Key features:
- ✅ Zero-boilerplate component mounting
- ✅ Type-safe state inspection
- ✅ Event tracking and assertions
- ✅ Snapshot testing with easy updates
- ✅ Async testing without flakiness
- ✅ Mock utilities for isolation
- ✅ Fast execution (< 1s per test)

Happy testing! 🧪
