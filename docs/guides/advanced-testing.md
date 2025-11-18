# Advanced Testing Guide

## Overview

This guide covers advanced testing patterns for BubblyUI features including command systems, composables, directives, router, provide/inject, key bindings, and observability. These patterns build on the [basic testing guide](testing-guide.md).

## Table of Contents

- [Command System Testing](#command-system-testing)
- [Composables Testing](#composables-testing)
- [Directives Testing](#directives-testing)
- [Router Testing](#router-testing)
- [Watch System Testing](#watch-system-testing)
- [Provide/Inject Testing](#provideinject-testing)
- [Observability Testing](#observability-testing)
- [Performance Testing](#performance-testing)

## Command System Testing

### Testing Command Queue

```go
func TestCommandQueue(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAutoCommandComponent())
    
    queue := harness.GetCommandQueue()
    assert.Equal(t, 0, queue.Len())
    
    count := component.State().GetRef("count")
    count.Set(42)
    
    assert.Equal(t, 1, queue.Len())
}
```

### Testing Loop Detection

```go
func TestLoopDetection(t *testing.T) {
    harness := testutil.NewHarness(t)
    loopDetector := testutil.NewLoopDetector()
    
    component := harness.MountWithLoopDetector(createComponent(), loopDetector)
    
    count := component.State().GetRef("count")
    for i := 0; i < 150; i++ {
        count.Set(i)
    }
    
    loopDetector.AssertDetected(t)
}
```

## Composables Testing

### Testing useAsync

```go
func TestUseAsync(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAsyncComponent())
    
    loading := component.State().GetRef("loading")
    
    component.Component().Emit("fetch", nil)
    assert.True(t, loading.Get().(bool))
    
    testutil.WaitFor(t, func() bool {
        return !loading.Get().(bool)
    }, testutil.WaitOptions{
        Timeout: 5 * time.Second,
    })
}
```

### Testing useDebounce

```go
func TestUseDebounce(t *testing.T) {
    harness := testutil.NewHarness(t)
    timeSim := testutil.NewTimeSimulator()
    
    component := harness.MountWithTime(createDebounceComponent(), timeSim)
    
    input := component.State().GetRef("input")
    debounced := component.State().GetRef("debounced")
    
    input.Set("test")
    assert.Equal(t, "", debounced.Get())
    
    timeSim.Advance(300 * time.Millisecond)
    assert.Equal(t, "test", debounced.Get())
}
```

## Directives Testing

### Testing ForEach

```go
func TestForEachDirective(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createListComponent())
    
    items := component.State().GetRef("items")
    items.Set([]string{"apple", "banana"})
    
    output := component.Component().View()
    assert.Contains(t, output, "apple")
    assert.Contains(t, output, "banana")
}
```

### Testing Bind

```go
func TestBindDirective(t *testing.T) {
    harness := testutil.NewHarness(t)
    bindTester := testutil.NewBindTester()
    
    component := harness.MountWithBindings(createFormComponent(), bindTester)
    
    name := component.State().GetRef("name")
    name.Set("John")
    
    bindTester.AssertElementValue(t, "nameInput", "John")
}
```

## Router Testing

### Testing Guards

```go
func TestRouteGuard(t *testing.T) {
    harness := testutil.NewHarness(t)
    router := testutil.NewMockRouter()
    
    guardCalled := false
    router.AddRoute("/protected").WithGuard(func() bool {
        guardCalled = true
        return false
    })
    
    router.Navigate("/protected")
    
    assert.True(t, guardCalled)
    assert.NotEqual(t, "/protected", router.CurrentPath())
}
```

## Watch System Testing

### Testing WatchEffect

```go
func TestWatchEffect(t *testing.T) {
    harness := testutil.NewHarness(t)
    effectTester := testutil.NewWatchEffectTester()
    
    count := harness.Ref(0)
    doubled := harness.Ref(0)
    
    effectTester.TrackEffect(func() {
        doubled.Set(count.Get().(int) * 2)
    })
    
    count.Set(5)
    effectTester.AssertExecuted(t, 1)
    assert.Equal(t, 10, doubled.Get())
}
```

### Testing Deep Watch

```go
func TestDeepWatch(t *testing.T) {
    harness := testutil.NewHarness(t)
    deepTester := testutil.NewDeepWatchTester()
    
    obj := harness.Ref(map[string]interface{}{
        "nested": map[string]interface{}{"value": 0},
    })
    
    callCount := 0
    deepTester.WatchDeep(obj, func() {
        callCount++
    })
    
    nested := obj.Get().(map[string]interface{})["nested"].(map[string]interface{})
    nested["value"] = 42
    obj.Set(obj.Get())
    
    deepTester.AssertCallCount(t, 1)
}
```

## Provide/Inject Testing

### Testing Injection

```go
func TestProvideInject(t *testing.T) {
    harness := testutil.NewHarness(t)
    injector := testutil.NewProvideInjectTester()
    
    root := harness.Mount(createRootComponent())
    injector.Provide("theme", darkTheme)
    
    child := harness.Mount(createChildComponent())
    root.AddChild(child)
    
    theme := injector.Inject(child, "theme")
    assert.Equal(t, darkTheme, theme)
}
```

## Observability Testing

### Testing Error Reporting

```go
func TestErrorReporting(t *testing.T) {
    harness := testutil.NewHarness(t)
    mockReporter := testutil.NewMockErrorReporter()
    observability.SetErrorReporter(mockReporter)
    defer observability.SetErrorReporter(nil)
    
    component := harness.Mount(createComponent())
    component.Component().Emit("triggerPanic", nil)
    
    mockReporter.AssertPanicReported(t)
}
```

## Performance Testing

### Benchmarking Components

```go
func BenchmarkComponent(b *testing.B) {
    for i := 0; i < b.N; i++ {
        harness := testutil.NewHarness(&testing.T{})
        component := harness.Mount(createComponent())
        _ = component.Component().View()
    }
}
```

## Next Steps

- **[Testing Guide](testing-guide.md)** - Basic testing patterns
- **[API Reference](../api/testutil-reference.md)** - Complete API documentation
- **[Examples](../../cmd/examples/)** - Working examples

## Summary

Advanced testing patterns enable comprehensive testing of:
- ✅ Command systems and auto-commands
- ✅ Composables with time simulation
- ✅ Directives and two-way binding
- ✅ Router guards and navigation
- ✅ Watch system and effects
- ✅ Dependency injection
- ✅ Observability integration
- ✅ Performance benchmarks

These patterns ensure your BubblyUI applications are thoroughly tested and production-ready.
