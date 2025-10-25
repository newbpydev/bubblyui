# Design Specification: Lifecycle Hooks

## Component Hierarchy

```
Component Lifecycle System
└── Lifecycle Manager
    ├── Hook Registry
    ├── Hook Executor
    ├── Cleanup Manager
    └── Dependency Tracker

Hook Types:
├── onSetup (initialization)
├── onMounted (after first render)
├── onBeforeUpdate (before re-render)
├── onUpdated (after re-render)
├── onBeforeUnmount (before removal)
└── onUnmounted (cleanup)
```

---

## Architecture Overview

### System Layers

```
┌────────────────────────────────────────────────────────────┐
│                   Lifecycle Hook System                     │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌────────────────┐      ┌──────────────────┐           │
│  │ Hook Registry  │─────>│  Hook Executor   │           │
│  │  (register)    │      │  (call hooks)    │           │
│  └────────────────┘      └──────────────────┘           │
│         │                         │                       │
│         │                         ▼                       │
│         │                ┌──────────────────┐            │
│         └───────────────>│ Cleanup Manager  │            │
│                          │  (auto-cleanup)  │            │
│                          └──────────────────┘            │
│                                                            │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│                    Component System                        │
│  Init() → onSetup → onMounted → Update → onUpdated       │
│                                    ↓                        │
│                            onUnmounted (cleanup)           │
└────────────────────────────────────────────────────────────┘
```

### Hook Lifecycle Flow

```
Component.Init() called
    ↓
Setup() function executes
    ↓
Hooks registered in registry
    ├─> onMounted hooks
    ├─> onUpdated hooks
    └─> onUnmounted hooks
    ↓
Component marked as "initialized"
    ↓
First View() call
    ↓
onMounted hooks execute (all in order)
    ↓
Component "mounted" state = true
    ↓
State changes → Update() → View()
    ↓
onBeforeUpdate hooks execute (optional)
    ↓
Re-render happens
    ↓
onUpdated hooks execute (all in order)
    ├─> Check dependencies
    └─> Execute if dependencies changed
    ↓
... (more updates) ...
    ↓
Component.Unmount() called
    ↓
onBeforeUnmount hooks execute (optional)
    ↓
onUnmounted hooks execute (all in order)
    ↓
Auto-cleanup executes
    ├─> Cleanup watchers
    ├─> Cleanup event handlers
    └─> Call registered cleanup functions
    ↓
Component destroyed
```

---

## Data Flow

### 1. Hook Registration Flow
```
Setup() function executing
    ↓
Developer calls ctx.OnMounted(fn)
    ↓
Hook stored in component.hooks.mounted[]
    ↓
Hook includes:
    - Callback function
    - Registration order
    - Dependencies (if any)
    ↓
Hook ready for execution
```

### 2. Hook Execution Flow
```
Lifecycle event triggered (e.g., component mounted)
    ↓
Lifecycle manager checks if hooks registered
    ↓
For each hook in registration order:
    ├─> Check if should execute
    ├─> Execute hook callback
    ├─> Catch any errors
    ├─> Log errors (don't crash)
    └─> Continue to next hook
    ↓
All hooks executed
```

### 3. Cleanup Flow
```
Component unmounting
    ↓
onUnmounted hooks execute first
    ↓
Auto-cleanup phase begins
    ├─> Cleanup watchers (from reactivity system)
    │   └─> Call watcher.cleanup()
    ├─> Cleanup event handlers
    │   └─> Remove all registered handlers
    └─> Call manual cleanup functions
        └─> Execute in reverse order
    ↓
Component fully cleaned up
```

---

## State Management

### Lifecycle Manager Structure
```go
type LifecycleManager struct {
    component *componentImpl
    
    // Hook storage
    mounted        []LifecycleHook
    beforeUpdate   []LifecycleHook
    updated        []LifecycleHook
    beforeUnmount  []LifecycleHook
    unmounted      []LifecycleHook
    
    // Cleanup storage
    cleanups       []CleanupFunc
    watchers       []*Watcher
    eventHandlers  []EventHandler
    
    // State
    isMounted      bool
    isUnmounting   bool
    updateCount    int
    
    // Error handling
    errorHandler   func(error)
}
```

### Lifecycle Hook Structure
```go
type LifecycleHook struct {
    id           string
    callback     func()
    dependencies []*Ref[any]
    lastValues   []any
    order        int
    executed     bool
}
```

### Cleanup Function
```go
type CleanupFunc func()

type Cleanup struct {
    fn    CleanupFunc
    order int
    name  string  // For debugging
}
```

---

## Type Definitions

### Core Types
```go
// Lifecycle hook function
type LifecycleHookFunc func()

// Cleanup function
type CleanupFunc func()

// Hook with dependencies
type DependentHook struct {
    callback     LifecycleHookFunc
    dependencies []*Ref[any]
}

// Error handler
type ErrorHandler func(hookType string, err error)
```

### Context Methods
```go
// Lifecycle hooks
func (ctx *Context) OnMounted(hook LifecycleHookFunc)
func (ctx *Context) OnUpdated(hook LifecycleHookFunc, deps ...*Ref[any])
func (ctx *Context) OnUnmounted(hook LifecycleHookFunc)
func (ctx *Context) OnBeforeUpdate(hook LifecycleHookFunc)
func (ctx *Context) OnBeforeUnmount(hook LifecycleHookFunc)

// Manual cleanup
func (ctx *Context) OnCleanup(cleanup CleanupFunc)

// State queries
func (ctx *Context) IsMounted() bool
func (ctx *Context) IsUnmounting() bool
```

---

## API Contracts

### Hook Registration API
```go
// In Setup function
Setup(func(ctx *Context) {
    // onMounted: Runs after first render
    ctx.OnMounted(func() {
        fmt.Println("Component mounted!")
    })
    
    // onUpdated: Runs after every update
    ctx.OnUpdated(func() {
        fmt.Println("Component updated!")
    })
    
    // onUpdated with dependencies: Only runs when deps change
    count := ctx.Ref(0)
    ctx.OnUpdated(func() {
        fmt.Printf("Count changed to: %d\n", count.Get())
    }, count)
    
    // onUnmounted: Runs on cleanup
    ctx.OnUnmounted(func() {
        fmt.Println("Component unmounting!")
    })
    
    // Manual cleanup registration
    ctx.OnCleanup(func() {
        fmt.Println("Cleanup function called")
    })
})
```

### Hook Execution Order
```go
// Multiple hooks of same type execute in registration order
ctx.OnMounted(func() {
    fmt.Println("First mounted hook")   // Executes first
})

ctx.OnMounted(func() {
    fmt.Println("Second mounted hook")  // Executes second
})

ctx.OnMounted(func() {
    fmt.Println("Third mounted hook")   // Executes third
})
```

### Dependency Tracking
```go
// Track multiple dependencies
user := ctx.Ref[*User](nil)
settings := ctx.Ref[*Settings](nil)

ctx.OnUpdated(func() {
    // Runs only when user OR settings change
    saveToBackend(user.Get(), settings.Get())
}, user, settings)

// No dependencies: runs on every update
ctx.OnUpdated(func() {
    // Runs on EVERY update
    logUpdate()
})
```

---

## Implementation Details

### Lifecycle Manager Implementation
```go
type componentImpl struct {
    // ... existing fields ...
    lifecycle *LifecycleManager
}

type LifecycleManager struct {
    component      *componentImpl
    hooks          map[string][]LifecycleHook
    cleanups       []Cleanup
    mounted        bool
    unmounting     bool
    updateCount    int
}

func newLifecycleManager(c *componentImpl) *LifecycleManager {
    return &LifecycleManager{
        component: c,
        hooks:     make(map[string][]LifecycleHook),
        cleanups:  []Cleanup{},
        mounted:   false,
        unmounting: false,
    }
}
```

### Hook Registration
```go
func (ctx *Context) OnMounted(hook LifecycleHookFunc) {
    if ctx.component.lifecycle.mounted {
        // Already mounted, execute immediately
        hook()
        return
    }
    
    ctx.component.lifecycle.registerHook("mounted", LifecycleHook{
        id:       generateID(),
        callback: hook,
        order:    len(ctx.component.lifecycle.hooks["mounted"]),
    })
}

func (ctx *Context) OnUpdated(hook LifecycleHookFunc, deps ...*Ref[any]) {
    h := LifecycleHook{
        id:           generateID(),
        callback:     hook,
        dependencies: deps,
        order:        len(ctx.component.lifecycle.hooks["updated"]),
    }
    
    if len(deps) > 0 {
        // Capture initial values
        h.lastValues = make([]any, len(deps))
        for i, dep := range deps {
            h.lastValues[i] = dep.Get()
        }
    }
    
    ctx.component.lifecycle.registerHook("updated", h)
}
```

### Hook Execution
```go
func (lm *LifecycleManager) executeMounted() {
    if lm.mounted {
        return  // Already mounted
    }
    
    lm.mounted = true
    lm.executeHooks("mounted")
}

func (lm *LifecycleManager) executeUpdated() {
    lm.updateCount++
    
    for _, hook := range lm.hooks["updated"] {
        shouldExecute := true
        
        // Check dependencies
        if len(hook.dependencies) > 0 {
            shouldExecute = false
            for i, dep := range hook.dependencies {
                currentValue := dep.Get()
                if !reflect.DeepEqual(currentValue, hook.lastValues[i]) {
                    shouldExecute = true
                    hook.lastValues[i] = currentValue
                }
            }
        }
        
        if shouldExecute {
            lm.safeExecuteHook("updated", hook)
        }
    }
}

func (lm *LifecycleManager) safeExecuteHook(hookType string, hook LifecycleHook) {
    defer func() {
        if r := recover(); r != nil {
            err := fmt.Errorf("hook panic: %v", r)
            lm.handleError(hookType, err)
        }
    }()
    
    hook.callback()
}

func (lm *LifecycleManager) handleError(hookType string, err error) {
    log.Printf("[BubblyUI] Error in %s hook: %v", hookType, err)
    // Don't crash the component
}
```

### Cleanup Execution
```go
func (lm *LifecycleManager) executeUnmounted() {
    if lm.unmounting {
        return  // Already unmounting
    }
    
    lm.unmounting = true
    
    // 1. Execute onUnmounted hooks
    lm.executeHooks("unmounted")
    
    // 2. Auto-cleanup watchers
    for _, watcher := range lm.watchers {
        watcher.cleanup()
    }
    
    // 3. Auto-cleanup event handlers
    for _, handler := range lm.eventHandlers {
        handler.remove()
    }
    
    // 4. Execute manual cleanup functions (reverse order)
    for i := len(lm.cleanups) - 1; i >= 0; i-- {
        cleanup := lm.cleanups[i]
        lm.safeExecuteCleanup(cleanup)
    }
    
    // 5. Clear all hooks
    lm.hooks = make(map[string][]LifecycleHook)
    lm.cleanups = []Cleanup{}
}

func (lm *LifecycleManager) safeExecuteCleanup(cleanup Cleanup) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[BubblyUI] Cleanup panic (%s): %v", cleanup.name, r)
        }
    }()
    
    cleanup.fn()
}
```

---

## Integration with Component System

### Component Init Integration
```go
func (c *componentImpl) Init() tea.Cmd {
    // Initialize lifecycle manager
    c.lifecycle = newLifecycleManager(c)
    
    // Execute Setup (registers hooks)
    if c.setup != nil {
        ctx := c.createContext()
        c.setup(ctx)
    }
    
    // Note: onMounted will execute after first View()
    
    return c.initChildren()
}
```

### Component View Integration
```go
func (c *componentImpl) View() string {
    // Execute onMounted on first render
    if !c.lifecycle.mounted {
        c.lifecycle.executeMounted()
    }
    
    // Render template
    if c.template == nil {
        return ""
    }
    
    ctx := c.createRenderContext()
    return c.template(ctx)
}
```

### Component Update Integration
```go
func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle message
    // ... existing update logic ...
    
    // After state changes, execute onUpdated
    if c.lifecycle.mounted && !c.lifecycle.unmounting {
        c.lifecycle.executeUpdated()
    }
    
    return c, cmd
}
```

### Component Unmount
```go
func (c *componentImpl) Unmount() {
    // Execute cleanup lifecycle
    c.lifecycle.executeUnmounted()
    
    // Unmount children
    for _, child := range c.children {
        child.Unmount()
    }
}
```

---

## Integration with Reactivity System

### Auto-Cleanup Watchers
```go
func (ctx *Context) Watch(ref *Ref[any], callback WatchCallback) WatchCleanup {
    // Create watcher
    cleanup := Watch(ref, callback)
    
    // Register for auto-cleanup
    ctx.component.lifecycle.registerWatcher(cleanup)
    
    return cleanup
}

func (lm *LifecycleManager) registerWatcher(cleanup WatchCleanup) {
    lm.watchers = append(lm.watchers, &Watcher{
        cleanup: cleanup,
    })
}
```

### State-Driven Updates
```go
// When Ref changes
ref.Set(newValue)
    ↓
Watchers notified
    ↓
Component re-renders (via Bubbletea Update)
    ↓
onUpdated hooks execute
    ├─> Check dependencies
    └─> Execute if changed
```

---

## Error Handling

### Error Types
```go
var (
    ErrHookPanic          = errors.New("hook execution panicked")
    ErrCleanupFailed      = errors.New("cleanup function failed")
    ErrMaxUpdateDepth     = errors.New("max update depth exceeded (infinite loop?)")
    ErrHookAfterMount     = errors.New("cannot register hooks after mount")
)
```

### Error Recovery
```go
func (lm *LifecycleManager) safeExecuteHook(hookType string, hook LifecycleHook) {
    defer func() {
        if r := recover(); r != nil {
            // Recover from panic
            err := fmt.Errorf("%w: %v\nStack: %s", 
                ErrHookPanic, r, debug.Stack())
            
            // Log error
            lm.handleError(hookType, err)
            
            // Component continues working
        }
    }()
    
    hook.callback()
}
```

### Infinite Loop Detection
```go
func (lm *LifecycleManager) executeUpdated() {
    const maxUpdateDepth = 100
    
    if lm.updateCount > maxUpdateDepth {
        lm.handleError("updated", ErrMaxUpdateDepth)
        return
    }
    
    lm.updateCount++
    // ... execute hooks ...
}
```

---

## Performance Optimizations

### 1. Dependency Comparison
```go
// Use reflect.DeepEqual with caching
func (lm *LifecycleManager) hasChanged(hook *LifecycleHook) bool {
    for i, dep := range hook.dependencies {
        current := dep.Get()
        if !reflect.DeepEqual(current, hook.lastValues[i]) {
            return true
        }
    }
    return false
}
```

### 2. Hook Pooling
```go
var hookPool = sync.Pool{
    New: func() interface{} {
        return &LifecycleHook{}
    },
}

func getHook() *LifecycleHook {
    return hookPool.Get().(*LifecycleHook)
}

func putHook(h *LifecycleHook) {
    h.reset()
    hookPool.Put(h)
}
```

### 3. Lazy Cleanup
```go
// Batch cleanup operations
func (lm *LifecycleManager) executeCleanup() {
    // Use goroutine for non-critical cleanup
    go func() {
        for _, cleanup := range lm.cleanups {
            cleanup.fn()
        }
    }()
}
```

---

## Testing Strategy

### Unit Tests
```go
func TestLifecycle_OnMounted(t *testing.T)
func TestLifecycle_OnUpdated(t *testing.T)
func TestLifecycle_OnUpdated_WithDependencies(t *testing.T)
func TestLifecycle_OnUnmounted(t *testing.T)
func TestLifecycle_ExecutionOrder(t *testing.T)
func TestLifecycle_ErrorRecovery(t *testing.T)
func TestLifecycle_Cleanup(t *testing.T)
func TestLifecycle_InfiniteLoopDetection(t *testing.T)
```

### Integration Tests
```go
func TestLifecycle_FullCycle(t *testing.T)
func TestLifecycle_NestedComponents(t *testing.T)
func TestLifecycle_WithReactivity(t *testing.T)
```

---

## Example Usage

### Data Fetching
```go
Setup(func(ctx *Context) {
    data := ctx.Ref[*User](nil)
    loading := ctx.Ref(true)
    
    ctx.OnMounted(func() {
        // Fetch data on mount
        go func() {
            user := fetchUser()
            data.Set(user)
            loading.Set(false)
        }()
    })
    
    ctx.OnUnmounted(func() {
        // Cancel pending requests
        cancelFetch()
    })
})
```

### Interval Timer
```go
Setup(func(ctx *Context) {
    ticker := time.NewTicker(time.Second)
    
    ctx.OnMounted(func() {
        go func() {
            for range ticker.C {
                ctx.Emit("tick", time.Now())
            }
        }()
    })
    
    ctx.OnUnmounted(func() {
        ticker.Stop()
    })
})
```

### Conditional Update
```go
Setup(func(ctx *Context) {
    user := ctx.Ref[*User](nil)
    lastSaved := ctx.Ref(time.Time{})
    
    // Save only when user changes
    ctx.OnUpdated(func() {
        saveUser(user.Get())
        lastSaved.Set(time.Now())
    }, user)
})
```

---

## Future Enhancements

1. **Async Hooks:** Support for async/await patterns
2. **Hook Middleware:** Intercept hook execution
3. **Hook Debugging:** Dev tools visualization
4. **Performance Monitoring:** Track hook execution time
5. **Error Boundaries:** Component-level error handling
