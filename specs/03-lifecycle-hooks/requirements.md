# Feature Name: Lifecycle Hooks

## Feature ID
03-lifecycle-hooks

## Overview
Implement Vue-inspired lifecycle hooks that allow developers to execute code at specific points in a component's lifecycle. Hooks provide a way to initialize data, perform side effects, manage resources, and cleanup when components are mounted, updated, or unmounted. This feature integrates with the component model (feature 02) and reactivity system (feature 01).

## User Stories
- As a **developer**, I want to run code when a component mounts so that I can initialize resources
- As a **developer**, I want to run code when a component updates so that I can react to changes
- As a **developer**, I want to run code when a component unmounts so that I can cleanup resources
- As a **developer**, I want to setup watchers that auto-cleanup so that I don't have memory leaks
- As a **developer**, I want hooks to execute in predictable order so that I can reason about my code
- As a **developer**, I want type-safe hooks so that I catch errors at compile time

## Functional Requirements

### 1. Lifecycle Hook Types
1.1. **onSetup**: Executes once during component initialization (before mount)  
1.2. **onMounted**: Executes after component is mounted and ready  
1.3. **onUpdated**: Executes after component re-renders due to state changes  
1.4. **onUnmounted**: Executes when component is being removed  
1.5. **onBeforeUpdate**: Executes before component updates (optional)  
1.6. **onBeforeUnmount**: Executes before component unmounts (optional)  

### 2. Hook Registration
2.1. Register hooks in Setup function  
2.2. Multiple hooks of same type allowed  
2.3. Hooks execute in registration order  
2.4. Type-safe hook signatures  
2.5. Hooks can be conditional  

### 3. Hook Execution
3.1. onSetup: Runs immediately during Init()  
3.2. onMounted: Runs after first render  
3.3. onUpdated: Runs after each state-driven re-render  
3.4. onUnmounted: Runs when component removed  
3.5. Hooks execute synchronously  
3.6. Async operations via Bubbletea commands  

### 4. Cleanup Management
4.1. onUnmounted registers cleanup functions  
4.2. Watchers auto-cleanup on unmount  
4.3. Event handlers auto-cleanup on unmount  
4.4. Manual cleanup functions supported  
4.5. Cleanup in reverse order of registration  

### 5. Error Handling
5.1. Hook errors don't crash component  
5.2. Errors logged and reported  
5.3. Component continues functioning  
5.4. Error boundaries (optional)  

### 6. Dependency Tracking
6.1. onUpdated can specify dependencies  
6.2. Only runs when dependencies change  
6.3. Empty deps: run on every update  
6.4. No deps specified: run on every update  

## Non-Functional Requirements

### Performance
- Hook registration: < 100ns per hook
- Hook execution overhead: < 500ns
- No memory leaks from forgotten cleanup
- Minimal impact on render performance

### Accessibility
- N/A (internal system)

### Security
- Hook errors contained
- No exposure of internal state
- Safe cleanup execution

### Type Safety
- **Strict typing:** All hook functions typed
- **Callback signatures:** Type-safe parameters
- **Cleanup functions:** Typed return values
- **No `any`:** Use interfaces with constraints

## Acceptance Criteria

### Hook Registration
- [ ] Can register multiple hooks
- [ ] Hooks execute in order
- [ ] Type-safe registration
- [ ] Conditional registration works
- [ ] No registration after mount

### Hook Execution
- [ ] onMounted runs after first render
- [ ] onUpdated runs on state change
- [ ] onUnmounted runs on removal
- [ ] Hooks execute synchronously
- [ ] Execution order predictable

### Cleanup
- [ ] onUnmounted cleanup works
- [ ] Watchers auto-cleanup
- [ ] Event handlers auto-cleanup
- [ ] Manual cleanup supported
- [ ] No memory leaks

### Error Handling
- [ ] Hook errors caught
- [ ] Component continues working
- [ ] Errors logged
- [ ] Clear error messages

### General
- [ ] Test coverage > 80%
- [ ] All hooks documented
- [ ] Examples provided
- [ ] Performance acceptable

## Dependencies
- **Requires:** 02-component-model (component system)
- **Uses:** 01-reactivity-system (watchers)
- **Unlocks:** 04-composition-api (composables can use hooks)

## Edge Cases

### 1. Hook Registration After Mount
**Scenario:** Developer tries to register hook after component mounted  
**Handling:** Throw error or ignore (document that hooks must be registered in Setup)

### 2. Hook Throws Error
**Scenario:** onMounted hook panics  
**Handling:** Recover, log error, continue with next hook

### 3. Circular Updates in onUpdated
**Scenario:** onUpdated modifies state, triggering another update  
**Handling:** Detect infinite loop (max iterations), break and log warning

### 4. Cleanup Function Throws
**Scenario:** onUnmounted cleanup panics  
**Handling:** Recover, log error, continue with other cleanup

### 5. onUpdated Without Dependencies
**Scenario:** onUpdated runs on every render (performance issue)  
**Handling:** Allow but document performance implications

### 6. Component Unmounts During Hook Execution
**Scenario:** Async operation in onMounted, component unmounts before completion  
**Handling:** Cancel operation, skip remaining hooks

## Testing Requirements

### Unit Tests (80%+ coverage)
- Hook registration
- Hook execution order
- Cleanup functionality
- Error handling
- Dependency tracking

### Integration Tests
- Full lifecycle (mount → update → unmount)
- Multiple components
- Nested components
- Cleanup verification

### Example Usage
- Data fetching on mount
- Cleanup on unmount
- Reactive updates
- Resource management

## Atomic Design Level
**Foundation** (Lifecycle management for all components)

Enables proper initialization, update handling, and cleanup for any component level.

## Related Components
- Component (feature 02): Hosts lifecycle hooks
- Reactivity (feature 01): Watchers need cleanup
- Composition API (feature 04): Composables use hooks

## Technical Constraints
- Must integrate with component Init/Update/View cycle
- Cleanup must be guaranteed (defer patterns)
- Cannot block rendering
- Must work with Bubbletea's synchronous model

## API Design

### Hook Registration in Setup
```go
NewComponent("MyComponent").
    Setup(func(ctx *Context) {
        // State
        count := ctx.Ref(0)
        data := ctx.Ref[*Data](nil)
        
        // onMounted: Initialize
        ctx.OnMounted(func() {
            // Load data
            data.Set(loadData())
        })
        
        // onUpdated: React to changes
        ctx.OnUpdated(func() {
            saveData(data.Get())
        }, data)  // Only when data changes
        
        // onUnmounted: Cleanup
        ctx.OnUnmounted(func() {
            cleanup()
        })
        
        // Watchers auto-cleanup on unmount
        ctx.Watch(count, func(newVal, oldVal int) {
            log.Printf("Count: %d", newVal)
        })
    }).
    Build()
```

### Hook Execution Order
```
Component Created
    ↓
Init() called by Bubbletea
    ↓
Setup() executes
    ├─> Hooks registered (onMounted, onUpdated, onUnmounted)
    ├─> State created (Refs)
    └─> Watchers created
    ↓
onMounted() hooks execute (all registered onMounted hooks)
    ↓
Component Ready (rendering, handling events)
    ↓
State changes → Update() → View()
    ↓
onUpdated() hooks execute (after View())
    ↓
... (more updates) ...
    ↓
Component Unmounting
    ↓
onUnmounted() hooks execute
    ├─> Registered cleanup functions
    ├─> Auto-cleanup watchers
    └─> Auto-cleanup event handlers
    ↓
Component Destroyed
```

## Performance Benchmarks
```go
BenchmarkHookRegister     10000000   100 ns/op    32 B/op   1 allocs/op
BenchmarkHookExecute      5000000    500 ns/op    64 B/op   2 allocs/op
BenchmarkCleanup          2000000    1000 ns/op   128 B/op  4 allocs/op
```

## Documentation Requirements
- [ ] Package godoc with lifecycle overview
- [ ] Each hook documented
- [ ] Execution order diagram
- [ ] Cleanup best practices
- [ ] 10+ runnable examples
- [ ] Common patterns
- [ ] Troubleshooting guide

## Success Metrics
- Developers understand lifecycle
- No memory leaks in tests
- Cleanup always executes
- Hook order predictable
- Clear error messages on mistakes

## Hook Signatures

### Context Methods
```go
// Lifecycle hooks
func (ctx *Context) OnMounted(hook func())
func (ctx *Context) OnUpdated(hook func(), deps ...*Ref[any])
func (ctx *Context) OnUnmounted(hook func())
func (ctx *Context) OnBeforeUpdate(hook func())
func (ctx *Context) OnBeforeUnmount(hook func())

// Cleanup registration
func (ctx *Context) OnCleanup(cleanup func())

// Access lifecycle state
func (ctx *Context) IsMounted() bool
func (ctx *Context) IsUnmounting() bool
```

## Example Usage Patterns

### Pattern 1: Data Fetching
```go
Setup(func(ctx *Context) {
    data := ctx.Ref[*User](nil)
    loading := ctx.Ref(false)
    
    ctx.OnMounted(func() {
        loading.Set(true)
        // Async fetch via Bubbletea command
        ctx.SendCmd(fetchUserCmd(func(user *User) {
            data.Set(user)
            loading.Set(false)
        }))
    })
    
    ctx.Expose("data", data)
    ctx.Expose("loading", loading)
})
```

### Pattern 2: Interval/Timer
```go
Setup(func(ctx *Context) {
    ticker := time.NewTicker(time.Second)
    
    ctx.OnMounted(func() {
        go func() {
            for range ticker.C {
                // Update time
                ctx.Emit("tick", time.Now())
            }
        }()
    })
    
    ctx.OnUnmounted(func() {
        ticker.Stop()
    })
})
```

### Pattern 3: Subscriptions
```go
Setup(func(ctx *Context) {
    messages := ctx.Ref([]string{})
    
    ctx.OnMounted(func() {
        subscription := websocket.Subscribe("messages")
        
        ctx.OnCleanup(func() {
            subscription.Unsubscribe()
        })
        
        go func() {
            for msg := range subscription.Messages() {
                current := messages.Get()
                messages.Set(append(current, msg))
            }
        }()
    })
})
```

### Pattern 4: Conditional Updates
```go
Setup(func(ctx *Context) {
    user := ctx.Ref[*User](nil)
    settings := ctx.Ref[*Settings](nil)
    
    // Only run when user changes
    ctx.OnUpdated(func() {
        settings.Set(loadSettings(user.Get()))
    }, user)
})
```

## Open Questions
1. Should onBeforeUpdate be included (adds complexity)?
2. How to handle async hooks (Promise-like patterns)?
3. Should hooks support async/await equivalent?
4. Error boundary pattern for components?
5. Dev tools to visualize lifecycle?
