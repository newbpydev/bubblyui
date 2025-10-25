# User Workflow: Lifecycle Hooks

## Primary User Journey

### Journey: Developer Adds Lifecycle Hooks to Component

1. **Entry Point**: Developer has a component that needs initialization and cleanup
   - System response: Context provides lifecycle hook methods
   - UI update: N/A (development phase)

2. **Step 1**: Register onMounted hook for initialization
   ```go
   Setup(func(ctx *Context) {
       data := ctx.Ref[[]Item](nil)
       loading := ctx.Ref(true)
       
       ctx.OnMounted(func() {
           // Initialize: Load data
           items := loadItems()
           data.Set(items)
           loading.Set(false)
       })
   })
   ```
   - System response: Hook registered in lifecycle manager
   - Developer sees: Hook stored, will execute after mount
   - Ready for: Adding more hooks

3. **Step 2**: Register onUpdated hook for persistence
   ```go
   ctx.OnUpdated(func() {
       // Save data when it changes
       saveItems(data.Get())
   }, data)  // Only when data changes
   ```
   - System response: Hook registered with dependency tracking
   - Developer sees: Hook will run when `data` Ref changes
   - Ready for: Adding cleanup

4. **Step 3**: Register onUnmounted hook for cleanup
   ```go
   ctx.OnUnmounted(func() {
       // Cleanup: Cancel pending operations
       cancelPendingRequests()
       closeConnections()
   })
   ```
   - System response: Cleanup hook registered
   - Developer sees: Cleanup guaranteed on unmount
   - Ready for: Running component

5. **Step 4**: Component lifecycle executes
   - Component mounts → onMounted runs → data loads
   - Data changes → onUpdated runs → data saves
   - Component unmounts → onUnmounted runs → cleanup executes

6. **Completion**: Full lifecycle working correctly
   - End state: Component initializes, updates, and cleans up properly
   - No memory leaks
   - Predictable behavior

---

## Alternative Paths

### Scenario A: Multiple Hooks of Same Type

1. **Developer registers multiple onMounted hooks**
   ```go
   Setup(func(ctx *Context) {
       ctx.OnMounted(func() {
           fmt.Println("First: Initialize data")
           initializeData()
       })
       
       ctx.OnMounted(func() {
           fmt.Println("Second: Start timer")
           startTimer()
       })
       
       ctx.OnMounted(func() {
           fmt.Println("Third: Connect websocket")
           connectWebSocket()
       })
   })
   ```

2. **System executes hooks in registration order**
   - Output: "First: Initialize data"
   - Output: "Second: Start timer"
   - Output: "Third: Connect websocket"

3. **All hooks execute successfully**
   - Predictable order
   - Each hook independent
   - Error in one doesn't stop others

**Use Case:** Organizing initialization into logical steps

### Scenario B: Conditional Hook Execution

1. **Developer adds hook based on props**
   ```go
   Setup(func(ctx *Context) {
       props := ctx.Props().(MyProps)
       
       if props.EnableAutoSave {
           data := ctx.Ref[*Data](nil)
           
           ctx.OnUpdated(func() {
               autoSave(data.Get())
           }, data)
       }
   })
   ```

2. **Hook only registered if condition met**
   - If EnableAutoSave=true: Hook registered
   - If EnableAutoSave=false: Hook not registered

**Use Case:** Feature flags and conditional behavior

### Scenario C: Manual Cleanup Registration

1. **Developer creates resource in onMounted**
   ```go
   Setup(func(ctx *Context) {
       var subscription *Subscription
       
       ctx.OnMounted(func() {
           subscription = websocket.Subscribe("updates")
           
           // Register cleanup immediately
           ctx.OnCleanup(func() {
               subscription.Unsubscribe()
           })
       })
   })
   ```

2. **Cleanup automatically called on unmount**
   - Component unmounts
   - OnCleanup functions execute
   - Subscription cleaned up
   - No resource leaks

**Use Case:** Scoped resource management

### Scenario D: Dependency Tracking

1. **Developer creates multiple reactive dependencies**
   ```go
   Setup(func(ctx *Context) {
       user := ctx.Ref[*User](nil)
       settings := ctx.Ref[*Settings](nil)
       theme := ctx.Ref("dark")
       
       // Watch specific dependencies
       ctx.OnUpdated(func() {
           applyTheme(theme.Get())
       }, theme)  // Only when theme changes
       
       // Watch multiple dependencies
       ctx.OnUpdated(func() {
           syncToBackend(user.Get(), settings.Get())
       }, user, settings)  // When either changes
       
       // No dependencies: runs on every update
       ctx.OnUpdated(func() {
           logUpdate()
       })  // Every update
   })
   ```

2. **System tracks changes efficiently**
   - theme changes: Only applyTheme runs
   - user changes: Both syncToBackend and logUpdate run
   - settings changes: Both syncToBackend and logUpdate run
   - Other changes: Only logUpdate runs

**Use Case:** Optimized reactive updates

---

## Error Handling Flows

### Error 1: Hook Panics During Execution
- **Trigger**: onMounted hook contains code that panics
- **User sees**: Error logged, component continues
- **Recovery**: Automatic (panic recovered)

**Example:**
```go
ctx.OnMounted(func() {
    // This will panic
    var data *Data
    data.Process()  // nil pointer dereference
})

// System behavior:
// 1. Catch panic in safeExecuteHook
// 2. Log error with stack trace
// 3. Continue with next hook (if any)
// 4. Component remains functional
```

### Error 2: Infinite Update Loop
- **Trigger**: onUpdated modifies state that triggers onUpdated
- **User sees**: Warning after 100 iterations
- **Recovery**: Loop broken, error logged

**Example:**
```go
ctx.OnUpdated(func() {
    count := count.Get()
    count.Set(count.Get() + 1)  // Creates infinite loop!
})

// System behavior:
// 1. Detect updateCount > 100
// 2. Log ErrMaxUpdateDepth
// 3. Stop executing onUpdated hooks
// 4. Prevent crash
```

### Error 3: Cleanup Function Throws
- **Trigger**: onUnmounted cleanup panics
- **User sees**: Error logged, other cleanups continue
- **Recovery**: Automatic (cleanup continues)

**Example:**
```go
ctx.OnUnmounted(func() {
    // First cleanup - will panic
    panic("cleanup error")
})

ctx.OnUnmounted(func() {
    // Second cleanup - still executes
    fmt.Println("Second cleanup runs")
})

// System behavior:
// 1. First cleanup panics
// 2. Panic caught and logged
// 3. Second cleanup still executes
// 4. All cleanups attempted
```

### Error 4: Hook Registered After Mount
- **Trigger**: Developer tries to register hook after component mounted
- **User sees**: Error or immediate execution (depending on hook type)
- **Recovery**: Hook may execute immediately or be rejected

**Example:**
```go
Setup(func(ctx *Context) {
    ctx.OnMounted(func() {
        // Component is now mounted
        
        // Trying to register onMounted after mount
        ctx.OnMounted(func() {
            // This might execute immediately or be rejected
        })
    })
})

// System behavior:
// Option A: Execute immediately (component already mounted)
// Option B: Log warning and ignore
// Document: Hooks should be registered in Setup
```

### Error 5: Missing Cleanup
- **Trigger**: Developer forgets to cleanup resource
- **User sees**: Resource leak (memory, connections, etc.)
- **Recovery**: Manual (developer must add cleanup)

**Example:**
```go
// ❌ Missing cleanup
ctx.OnMounted(func() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            // Do something
        }
    }()
    // MISSING: cleanup for ticker!
})

// ✅ Proper cleanup
ctx.OnMounted(func() {
    ticker := time.NewTicker(time.Second)
    
    ctx.OnCleanup(func() {
        ticker.Stop()
    })
    
    go func() {
        for range ticker.C {
            // Do something
        }
    }()
})
```

---

## State Transitions

### Component Lifecycle States
```
Created
    ↓
Initializing (Init called, Setup executing)
    ↓
Initialized (Hooks registered)
    ↓
Mounting (First View call)
    ↓
Mounted (onMounted executed)
    ↓
Active (Rendering, updating)
    ↓
Unmounting (Unmount called)
    ↓
Unmounted (onUnmounted executed, cleanup done)
    ↓
Destroyed
```

### Hook Execution States
```
Registered
    ↓
Pending (Waiting for lifecycle event)
    ↓
Executing (Callback running)
    ↓
Executed (Callback finished)
    ↓
(For onUpdated) Back to Pending
    ↓
(For onUnmounted) Cleaned up
```

### Update Cycle with Hooks
```
State Change
    ↓
onBeforeUpdate hooks execute (optional)
    ↓
Update() called
    ↓
View() called (re-render)
    ↓
onUpdated hooks execute
    ├─> Check dependencies
    ├─> Execute if changed
    └─> Update lastValues
    ↓
Component ready for next update
```

---

## Integration Points

### Connected to: Component System (Feature 02)
- **Uses:** Component lifecycle (Init, Update, View, Unmount)
- **Flow:** Hooks execute at component lifecycle milestones
- **Data:** LifecycleManager stored in component

### Connected to: Reactivity System (Feature 01)
- **Uses:** Watch function for auto-cleanup
- **Flow:** Watchers registered in hooks, cleaned up on unmount
- **Data:** Cleanup functions tracked

### Connected to: Composition API (Feature 04, Future)
- **Uses:** Hooks in composable functions
- **Flow:** Composables register hooks, return to component
- **Data:** Shared lifecycle logic

---

## Performance Considerations

### Hook Execution Performance
**Target:** < 500ns per hook execution

**Measurement:**
```go
func BenchmarkHookExecute(b *testing.B) {
    ctx := createTestContext()
    hookCount := 0
    
    ctx.OnMounted(func() {
        hookCount++
    })
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ctx.component.lifecycle.executeMounted()
    }
}
```

### Dependency Checking Performance
**Scenario:** onUpdated with 10 dependencies

**Optimization:**
```go
// Fast path: no dependencies
if len(hook.dependencies) == 0 {
    hook.callback()
    return
}

// Optimized comparison
for i, dep := range hook.dependencies {
    current := dep.Get()
    if current != hook.lastValues[i] {  // Fast pointer compare first
        shouldExecute = true
        hook.lastValues[i] = current
        break
    }
}
```

### Memory Management
**Goal:** No memory leaks from hooks

**Strategy:**
- Clear hook arrays on unmount
- Pool hook objects
- Cleanup in reverse order
- Use defer for guaranteed execution

---

## Common Patterns

### Pattern 1: Data Fetching
```go
Setup(func(ctx *Context) {
    data := ctx.Ref[*Data](nil)
    loading := ctx.Ref(true)
    error := ctx.Ref[error](nil)
    
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
})
```

### Pattern 2: Event Subscriptions
```go
Setup(func(ctx *Context) {
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

### Pattern 3: Debounced Saves
```go
Setup(func(ctx *Context) {
    data := ctx.Ref[*Data](nil)
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

### Pattern 4: Polling
```go
Setup(func(ctx *Context) {
    data := ctx.Ref[*Data](nil)
    done := make(chan bool)
    
    ctx.OnMounted(func() {
        go func() {
            ticker := time.NewTicker(5 * time.Second)
            defer ticker.Stop()
            
            for {
                select {
                case <-ticker.C:
                    data.Set(pollData())
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

---

## Testing Workflow

### Unit Test: onMounted Execution
```go
func TestLifecycle_OnMounted(t *testing.T) {
    // Arrange
    executed := false
    
    component := NewComponent("Test").
        Setup(func(ctx *Context) {
            ctx.OnMounted(func() {
                executed = true
            })
        }).
        Build()
    
    // Act
    component.Init()
    component.View()  // Triggers onMounted
    
    // Assert
    assert.True(t, executed, "onMounted should execute")
}
```

### Unit Test: onUpdated with Dependencies
```go
func TestLifecycle_OnUpdated_Dependencies(t *testing.T) {
    // Arrange
    updateCount := 0
    
    component := NewComponent("Test").
        Setup(func(ctx *Context) {
            count := ctx.Ref(0)
            name := ctx.Ref("test")
            
            ctx.OnUpdated(func() {
                updateCount++
            }, count)  // Only watch count
            
            ctx.Expose("count", count)
            ctx.Expose("name", name)
        }).
        Build()
    
    component.Init()
    component.View()
    
    // Act & Assert
    count := component.Get("count").(*Ref[int])
    name := component.Get("name").(*Ref[string])
    
    count.Set(1)
    component.View()
    assert.Equal(t, 1, updateCount)
    
    name.Set("changed")
    component.View()
    assert.Equal(t, 1, updateCount, "Should not increment for name change")
    
    count.Set(2)
    component.View()
    assert.Equal(t, 2, updateCount)
}
```

### Integration Test: Full Lifecycle
```go
func TestLifecycle_FullCycle(t *testing.T) {
    // Arrange
    var lifecycle []string
    
    component := NewComponent("Test").
        Setup(func(ctx *Context) {
            ctx.OnMounted(func() {
                lifecycle = append(lifecycle, "mounted")
            })
            
            ctx.OnUpdated(func() {
                lifecycle = append(lifecycle, "updated")
            })
            
            ctx.OnUnmounted(func() {
                lifecycle = append(lifecycle, "unmounted")
            })
        }).
        Build()
    
    // Act
    component.Init()
    lifecycle = append(lifecycle, "init")
    
    component.View()
    lifecycle = append(lifecycle, "view1")
    
    component.Update(someMsg)
    component.View()
    lifecycle = append(lifecycle, "view2")
    
    component.Unmount()
    lifecycle = append(lifecycle, "unmount")
    
    // Assert
    expected := []string{
        "init",
        "mounted",
        "view1",
        "updated",
        "view2",
        "unmounted",
        "unmount",
    }
    assert.Equal(t, expected, lifecycle)
}
```

---

## Documentation for Users

### Quick Start
1. Import Context in Setup function
2. Register hooks: `ctx.OnMounted(fn)`
3. Add dependencies: `ctx.OnUpdated(fn, ref1, ref2)`
4. Cleanup: `ctx.OnUnmounted(fn)` or `ctx.OnCleanup(fn)`

### Best Practices
- Register all hooks in Setup
- Use onMounted for initialization
- Specify dependencies for onUpdated
- Always cleanup resources in onUnmounted
- Don't create infinite loops in onUpdated
- Test hook execution order
- Handle errors gracefully

### Troubleshooting
- **Hook not firing?** Check if registered in Setup
- **onUpdated running too often?** Add dependencies
- **Memory leak?** Verify cleanup registered
- **Infinite loop?** Check onUpdated doesn't modify its dependencies
- **Hook after mount?** Hooks must be registered in Setup

---

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
Setup(func(ctx *Context) {
    data := ctx.Ref[*Data](nil)
    
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
