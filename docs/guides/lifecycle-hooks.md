# Lifecycle Hooks Guide

## Overview

Lifecycle hooks allow you to execute code at specific points in a component's lifecycle. They provide a way to initialize data, perform side effects, manage resources, and cleanup when components are mounted, updated, or unmounted.

## Table of Contents

- [Quick Start](#quick-start)
- [Hook Types](#hook-types)
- [Hook Execution Order](#hook-execution-order)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)

## Quick Start

```go
component, _ := bubbly.NewComponent("MyComponent").
    Setup(func(ctx *bubbly.Context) {
        data := ctx.Ref(nil)
        
        // Initialize on mount
        ctx.OnMounted(func() {
            data.Set(loadData())
        })
        
        // React to changes
        ctx.OnUpdated(func() {
            saveData(data.Get())
        }, data)
        
        // Cleanup on unmount
        ctx.OnUnmounted(func() {
            cleanup()
        })
    }).
    Build()
```

## Hook Types

### onMounted

Executes after the component is first rendered and ready.

**Use cases:**
- Fetching data from APIs
- Initializing third-party libraries
- Starting timers or intervals
- Subscribing to events

**Example:**
```go
ctx.OnMounted(func() {
    fmt.Println("Component is ready!")
    // Initialize resources here
})
```

### onUpdated

Executes after the component re-renders due to state changes.

**Use cases:**
- Persisting data to backend
- Syncing state with external systems
- Logging state changes
- Triggering side effects

**Without dependencies (runs on every update):**
```go
ctx.OnUpdated(func() {
    fmt.Println("Component updated")
})
```

**With dependencies (runs only when dependencies change):**
```go
user := ctx.Ref(nil)

ctx.OnUpdated(func() {
    saveUser(user.Get())
}, user)  // Only runs when user changes
```

**Multiple dependencies:**
```go
ctx.OnUpdated(func() {
    sync(user.Get(), settings.Get())
}, user, settings)  // Runs when either changes
```

### onUnmounted

Executes when the component is being removed.

**Use cases:**
- Canceling pending requests
- Stopping timers
- Unsubscribing from events
- Releasing resources

**Example:**
```go
ctx.OnUnmounted(func() {
    fmt.Println("Cleaning up...")
    cancelRequests()
    stopTimer()
})
```

### onCleanup

Registers cleanup functions that execute during unmount.

**Use cases:**
- Scoped resource cleanup
- Cleanup tied to specific operations
- LIFO cleanup order

**Example:**
```go
ctx.OnMounted(func() {
    subscription := subscribe("events")
    
    // Register cleanup immediately
    ctx.OnCleanup(func() {
        subscription.Unsubscribe()
    })
})
```

## Hook Execution Order

### Component Lifecycle

```
Component Created
    ↓
Init() called
    ↓
Setup() executes
    ├─> Hooks registered (onMounted, onUpdated, onUnmounted)
    ├─> State created (Refs)
    └─> Watchers created
    ↓
First View() call
    ↓
onMounted hooks execute (all in registration order)
    ↓
Component Active (rendering, handling events)
    ↓
State changes → Update() → View()
    ↓
onUpdated hooks execute (after View())
    ↓
... (more updates) ...
    ↓
Component Unmounting
    ↓
onUnmounted hooks execute
    ├─> Registered cleanup functions
    ├─> Auto-cleanup watchers
    └─> Auto-cleanup event handlers
    ↓
Component Destroyed
```

### Multiple Hooks

When you register multiple hooks of the same type, they execute in **registration order**:

```go
ctx.OnMounted(func() {
    fmt.Println("First")   // Executes first
})

ctx.OnMounted(func() {
    fmt.Println("Second")  // Executes second
})

ctx.OnMounted(func() {
    fmt.Println("Third")   // Executes third
})
```

### Cleanup Order

Cleanup functions execute in **reverse order (LIFO)**:

```go
ctx.OnCleanup(func() {
    fmt.Println("First registered")   // Executes SECOND
})

ctx.OnCleanup(func() {
    fmt.Println("Second registered")  // Executes FIRST
})
```

This ensures proper resource unwinding (dependencies cleaned up before dependents).

## Common Patterns

### Pattern 1: Data Fetching

```go
Setup(func(ctx *bubbly.Context) {
    data := ctx.Ref(nil)
    loading := ctx.Ref(true)
    error := ctx.Ref(nil)
    
    ctx.OnMounted(func() {
        go func() {
            result, err := fetchData()
            if err != nil {
                error.Set(err)
            } else {
                data.Set(result)
            }
            loading.Set(false)
        }()
    })
    
    ctx.Expose("data", data)
    ctx.Expose("loading", loading)
    ctx.Expose("error", error)
})
```

### Pattern 2: Timer/Interval

```go
Setup(func(ctx *bubbly.Context) {
    done := make(chan bool)
    
    ctx.OnMounted(func() {
        go func() {
            ticker := time.NewTicker(time.Second)
            defer ticker.Stop()
            
            for {
                select {
                case <-ticker.C:
                    // Update time
                    ctx.Emit("tick", time.Now())
                case <-done:
                    return
                }
            }
        }()
    })
    
    ctx.OnUnmounted(func() {
        close(done)
    })
})
```

### Pattern 3: Event Subscriptions

```go
Setup(func(ctx *bubbly.Context) {
    var unsubscribe func()
    
    ctx.OnMounted(func() {
        unsubscribe = eventBus.Subscribe("updates", func(event Event) {
            // Handle event
        })
    })
    
    ctx.OnUnmounted(func() {
        if unsubscribe != nil {
            unsubscribe()
        }
    })
})
```

### Pattern 4: Auto-Save with Debouncing

```go
Setup(func(ctx *bubbly.Context) {
    data := ctx.Ref(nil)
    var saveTimer *time.Timer
    
    ctx.OnUpdated(func() {
        // Debounce: save 500ms after last change
        if saveTimer != nil {
            saveTimer.Stop()
        }
        
        saveTimer = time.AfterFunc(500*time.Millisecond, func() {
            saveData(data.Get())
        })
    }, data)
    
    ctx.OnUnmounted(func() {
        if saveTimer != nil {
            saveTimer.Stop()
            // Final save on unmount
            saveData(data.Get())
        }
    })
})
```

### Pattern 5: Conditional Hooks

```go
Setup(func(ctx *bubbly.Context) {
    props := ctx.Props().(MyProps)
    data := ctx.Ref(nil)
    
    // Only register hook if feature is enabled
    if props.EnableAutoSave {
        ctx.OnUpdated(func() {
            autoSave(data.Get())
        }, data)
    }
})
```

### Pattern 6: Watcher Auto-Cleanup

Watchers created in Setup are automatically cleaned up on unmount:

```go
Setup(func(ctx *bubbly.Context) {
    count := ctx.Ref(0)
    
    // Watcher automatically cleaned up on unmount
    ctx.Watch(count, func(newVal, oldVal interface{}) {
        fmt.Printf("Count: %d → %d\n", oldVal.(int), newVal.(int))
    })
})
```

## Best Practices

### 1. Register Hooks in Setup

✅ **Do:**
```go
Setup(func(ctx *bubbly.Context) {
    ctx.OnMounted(func() {
        // Initialize
    })
})
```

❌ **Don't:**
```go
// Hooks registered outside Setup may not work correctly
ctx.OnMounted(func() {
    // This may not execute
})
```

### 2. Specify Dependencies for onUpdated

✅ **Do:**
```go
ctx.OnUpdated(func() {
    saveUser(user.Get())
}, user)  // Only runs when user changes
```

❌ **Don't:**
```go
ctx.OnUpdated(func() {
    saveUser(user.Get())
})  // Runs on EVERY update (performance issue)
```

### 3. Always Cleanup Resources

✅ **Do:**
```go
ctx.OnMounted(func() {
    ticker := time.NewTicker(time.Second)
    
    ctx.OnCleanup(func() {
        ticker.Stop()  // Cleanup registered
    })
})
```

❌ **Don't:**
```go
ctx.OnMounted(func() {
    ticker := time.NewTicker(time.Second)
    // Missing cleanup - memory leak!
})
```

### 4. Avoid Infinite Loops

✅ **Do:**
```go
ctx.OnUpdated(func() {
    // Read state, don't modify it
    logState(count.Get())
}, count)
```

❌ **Don't:**
```go
ctx.OnUpdated(func() {
    count.Set(count.Get() + 1)  // Infinite loop!
}, count)
```

### 5. Use Scoped Cleanup

✅ **Do:**
```go
ctx.OnMounted(func() {
    resource := createResource()
    
    ctx.OnCleanup(func() {
        resource.Close()  // Cleanup tied to resource
    })
})
```

### 6. Handle Errors Gracefully

✅ **Do:**
```go
ctx.OnMounted(func() {
    defer func() {
        if r := recover(); r != nil {
            // Handle panic
            error.Set(fmt.Errorf("initialization failed: %v", r))
        }
    }()
    
    // Initialization code
})
```

### 7. Test Hook Execution

```go
func TestComponent_Lifecycle(t *testing.T) {
    executed := false
    
    component, _ := bubbly.NewComponent("Test").
        Setup(func(ctx *bubbly.Context) {
            ctx.OnMounted(func() {
                executed = true
            })
        }).
        Build()
    
    component.Init()
    component.View()
    
    assert.True(t, executed, "onMounted should execute")
}
```

## Troubleshooting

### Hook Not Firing

**Problem:** onMounted hook doesn't execute

**Solutions:**
1. Ensure hooks are registered in Setup
2. Call `component.Init()` before `component.View()`
3. Check that component is actually rendered

```go
component.Init()  // Initialize first
component.View()  // Then render (triggers onMounted)
```

### onUpdated Running Too Often

**Problem:** onUpdated executes on every update

**Solutions:**
1. Add dependencies to limit execution
2. Use specific Refs as dependencies

```go
// Before (runs on every update)
ctx.OnUpdated(func() {
    save(data.Get())
})

// After (runs only when data changes)
ctx.OnUpdated(func() {
    save(data.Get())
}, data)
```

### Memory Leak

**Problem:** Resources not cleaned up

**Solutions:**
1. Register cleanup in onUnmounted or onCleanup
2. Verify cleanup is actually called
3. Use defer patterns for guaranteed cleanup

```go
ctx.OnMounted(func() {
    ticker := time.NewTicker(time.Second)
    
    // Always register cleanup
    ctx.OnCleanup(func() {
        ticker.Stop()
    })
})
```

### Infinite Loop Detected

**Problem:** "max update depth exceeded" error

**Solutions:**
1. Don't modify dependencies in onUpdated
2. Use different Refs for read and write
3. Add guards to prevent circular updates

```go
// Problem
ctx.OnUpdated(func() {
    count.Set(count.Get() + 1)  // Modifies its own dependency!
}, count)

// Solution
ctx.OnUpdated(func() {
    // Only read, don't modify
    logCount(count.Get())
}, count)
```

### Hook Registered After Mount

**Problem:** Hook registered too late

**Solution:** Always register hooks in Setup, not in other hooks

```go
// ❌ Wrong
ctx.OnMounted(func() {
    ctx.OnMounted(func() {
        // Too late!
    })
})

// ✅ Correct
ctx.OnMounted(func() {
    // Initialize
})
```

## API Reference

### Context.OnMounted

```go
func (ctx *Context) OnMounted(hook func())
```

Registers a hook that executes after the component is first rendered.

**Parameters:**
- `hook`: Function to execute after mount

**Example:**
```go
ctx.OnMounted(func() {
    fmt.Println("Component mounted")
})
```

### Context.OnUpdated

```go
func (ctx *Context) OnUpdated(hook func(), deps ...*Ref[any])
```

Registers a hook that executes after component updates.

**Parameters:**
- `hook`: Function to execute after update
- `deps`: Optional dependencies (hook only runs when these change)

**Example:**
```go
// No dependencies: runs on every update
ctx.OnUpdated(func() {
    logUpdate()
})

// With dependencies: runs only when user changes
ctx.OnUpdated(func() {
    saveUser(user.Get())
}, user)
```

### Context.OnUnmounted

```go
func (ctx *Context) OnUnmounted(hook func())
```

Registers a hook that executes when the component is unmounted.

**Parameters:**
- `hook`: Function to execute on unmount

**Example:**
```go
ctx.OnUnmounted(func() {
    fmt.Println("Component unmounting")
    cleanup()
})
```

### Context.OnCleanup

```go
func (ctx *Context) OnCleanup(cleanup CleanupFunc)
```

Registers a cleanup function that executes during unmount.

**Parameters:**
- `cleanup`: Cleanup function to execute

**Example:**
```go
ctx.OnCleanup(func() {
    resource.Close()
})
```

**Note:** Cleanup functions execute in reverse order (LIFO).

### Context.IsMounted

```go
func (ctx *Context) IsMounted() bool
```

Returns whether the component is currently mounted.

**Returns:** `true` if mounted, `false` otherwise

### Context.IsUnmounting

```go
func (ctx *Context) IsUnmounting() bool
```

Returns whether the component is currently unmounting.

**Returns:** `true` if unmounting, `false` otherwise

## Performance Considerations

### Hook Registration

- **Target:** < 100ns per hook
- **Actual:** ~232ns (acceptable for one-time setup)
- **Impact:** Negligible (only during component initialization)

### Hook Execution

- **No dependencies:** ~15ns (66M ops/sec)
- **With dependencies:** ~37ns (27M ops/sec)
- **Impact:** Minimal overhead on updates

### Dependency Checking

- **Zero dependencies:** ~2ns (fast path)
- **One dependency:** ~36ns
- **Five dependencies:** ~180ns
- **Impact:** Scales linearly with dependency count

### Cleanup

- **10 cleanup functions:** ~64ns
- **Impact:** Minimal overhead on unmount

## Migration from Manual Lifecycle

### Before (Manual)

```go
type model struct {
    mounted bool
    data    *Data
}

func (m model) Init() tea.Cmd {
    return fetchDataCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case DataFetchedMsg:
        m.data = msg.Data
        if !m.mounted {
            m.mounted = true
            return m, initCmd
        }
    }
    return m, nil
}
```

### After (With Hooks)

```go
Setup(func(ctx *bubbly.Context) {
    data := ctx.Ref(nil)
    
    ctx.OnMounted(func() {
        go func() {
            result := fetchData()
            data.Set(result)
        }()
    })
    
    ctx.Expose("data", data)
})
```

**Benefits:**
- Clearer intent
- Less boilerplate
- Automatic cleanup
- Easier to test
- Better separation of concerns

## Advanced Topics

### Error Recovery

Hooks automatically recover from panics:

```go
ctx.OnMounted(func() {
    panic("error")  // Caught and logged
})

ctx.OnMounted(func() {
    fmt.Println("Still executes")  // Continues execution
})
```

### Infinite Loop Detection

The system detects infinite update loops (max 100 iterations):

```go
ctx.OnUpdated(func() {
    count.Set(count.Get() + 1)  // Triggers infinite loop detection
}, count)
```

### Auto-Cleanup

Watchers and event handlers are automatically cleaned up:

```go
// Watcher auto-cleanup
ctx.Watch(count, func(newVal, oldVal interface{}) {
    // Automatically cleaned up on unmount
})

// Event handler auto-cleanup
ctx.On("event", func(data interface{}) {
    // Automatically cleaned up on unmount
})
```

## Examples

See `pkg/bubbly/lifecycle_examples_test.go` for 15+ runnable examples covering:

- Basic hook usage
- Dependency tracking
- Data fetching patterns
- Timer management
- Event subscriptions
- Cleanup patterns
- Error recovery
- Nested components
- And more...

Run examples:
```bash
go test -run Example -v ./pkg/bubbly
```

## See Also

- [Component Model Guide](./component-model.md)
- [Reactivity System Guide](./reactivity-system.md)
- [Error Tracking Guide](./error-tracking.md)
- [API Documentation](../api/)
