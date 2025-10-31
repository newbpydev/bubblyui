# Feature Name: Composition API

## Feature ID
04-composition-api

## Overview
Implement a Vue 3-inspired Composition API that enables developers to organize component logic into reusable, composable functions. The Composition API provides better code organization, logic reuse across components, more flexible code sharing compared to mixins, and improved TypeScript integration. It builds on top of the reactivity system (feature 01), component model (feature 02), and lifecycle hooks (feature 03).

## User Stories
- As a **developer**, I want to extract reusable logic into composable functions so that I can share code between components
- As a **developer**, I want to organize related code together so that my components are easier to maintain
- As a **developer**, I want type-safe composables so that I catch errors at compile time
- As a **developer**, I want to use composables in Setup so that I can leverage reactive state and lifecycle hooks
- As a **developer**, I want dependency injection via provide/inject so that I can pass data through component trees without prop drilling
- As a **developer**, I want standard composable patterns (useState, useEffect, etc.) so that I have familiar APIs

## Functional Requirements

### 1. Composable Function Pattern
1.1. Define composables as Go functions that return Refs, computed values, and handlers  
1.2. Composables can call other composables  
1.3. Composables have access to Context for lifecycle and reactivity  
1.4. Composables are type-safe with generics  
1.5. Naming convention: `Use*` prefix (e.g., `UseCounter`, `UseForm`)  

### 2. Standard Composables
2.1. **UseState**: Simplified state management  
2.2. **UseEffect**: Side effect handling  
2.3. **UseAsync**: Async data fetching  
2.4. **UseDebounce**: Debounced values  
2.5. **UseThrottle**: Throttled functions  
2.6. **UseLocalStorage**: Persistent state  
2.7. **UseEventListener**: Event handling  

### 3. Context System
3.1. Extended Context in Setup provides composable environment  
3.2. Context gives access to component instance  
3.3. Context provides lifecycle hook registration  
3.4. Context provides reactivity primitives  
3.5. Context is passed to all composables  
3.6. **Dependency interface** allows typed refs to work with UseEffect (quality of life enhancement)

### 4. Provide/Inject Pattern
4.1. Parent components can "provide" values  
4.2. Child components can "inject" provided values  
4.3. Type-safe provide/inject with generics  
4.4. Default values for inject  
4.5. Reactive values can be provided/injected  
4.6. Works across any depth of component tree  

### 5. Composable Composition
5.1. Composables can call other composables  
5.2. Composables can share state via closure  
5.3. Composables can return other composables  
5.4. Composables can be conditionally executed  

### 6. Return Value Convention
6.1. Composables return plain objects with named exports  
6.2. Return Refs and computed values (not plain values)  
6.3. Return cleanup functions if needed  
6.4. Destructuring preserves reactivity  

### 7. Reactive Dependency Interface (Enhancement)
7.1. Define `Dependency` interface for reactive values  
7.2. Both `Ref[T]` and `Computed[T]` implement Dependency  
7.3. UseEffect accepts `...Dependency` instead of `...*Ref[any]`  
7.4. Watch can accept Dependency for monitoring computed values  
7.5. Interface provides `Get() any` for value access  
7.6. Enables typed refs (`*Ref[int]`) to work with UseEffect without conversion  
7.7. Backwards compatible with existing code through interface implementation  

## Non-Functional Requirements

### Performance
- Composable call overhead: < 100ns
- State access through composables: < 10ns
- Provide/inject lookup: < 500ns
- No memory leaks from composables

### Accessibility
- N/A (internal system)

### Security
- Composables cannot access other component's private state
- Provide/inject scoped to component tree
- Type-safe injection prevents type confusion

### Type Safety
- **Strict typing:** All composables type-safe
- **Generic composables:** `Use*[T any]` pattern
- **Return types:** Explicit return type signatures
- **No `any`:** Use interfaces with constraints
- **Compile-time validation:** Catch errors before runtime

## Acceptance Criteria

### Composable Pattern
- [ ] Can define composable functions
- [ ] Composables return reactive values
- [ ] Composables can be called in Setup
- [ ] Composables can call other composables
- [ ] Type safety enforced

### Standard Composables
- [ ] UseState works correctly
- [ ] UseEffect handles side effects
- [ ] UseAsync handles async operations
- [ ] All standard composables documented
- [ ] Examples for each composable

### Context System
- [ ] Context available in Setup
- [ ] Context provides all needed APIs
- [ ] Context passed to composables
- [ ] Type-safe context

### Provide/Inject
- [ ] Can provide values
- [ ] Can inject values
- [ ] Works across component tree
- [ ] Type-safe
- [ ] Default values work
- [ ] Reactive values propagate

### Dependency Interface
- [ ] Dependency interface defined
- [ ] Ref implements Dependency
- [ ] Computed implements Dependency
- [ ] UseEffect accepts Dependency
- [ ] Watch accepts Dependency
- [ ] Backwards compatible
- [ ] Typed refs work seamlessly

### General
- [ ] Test coverage > 80%
- [ ] All composables documented
- [ ] Examples provided
- [ ] Performance acceptable

## Dependencies
- **Requires:** 
  - 01-reactivity-system (Ref, Computed, Watch)
  - 02-component-model (Component, Context)
  - 03-lifecycle-hooks (onMounted, onUpdated, onUnmounted)
- **Unlocks:** All higher-level features (directives, built-in components)

## Edge Cases

### 1. Composable Called Outside Setup
**Scenario:** Developer calls composable outside Setup function  
**Handling:** Throw error with clear message about composable scope

### 2. Circular Composable Dependencies
**Scenario:** ComposableA calls ComposableB which calls ComposableA  
**Handling:** Detect cycle, throw error with call stack

### 3. Inject Without Provide
**Scenario:** Child tries to inject value that wasn't provided  
**Handling:** Return default value or throw error if no default

### 4. Multiple Provides Same Key
**Scenario:** Parent and grandparent both provide same key  
**Handling:** Use nearest provider (parent wins)

### 5. Provide/Inject with Non-Reactive Values
**Scenario:** Provided value is not reactive  
**Handling:** Allow but document that changes won't propagate

### 6. Composable State Leaking Between Instances
**Scenario:** Composable uses global state accidentally  
**Handling:** Document that composables should use closure or Context for state

## Testing Requirements

### Unit Tests (80%+ coverage)
- Composable function execution
- Standard composables behavior
- Context integration
- Provide/inject functionality
- Composable composition
- Error handling

### Integration Tests
- Composables in components
- Cross-component provide/inject
- Composable chains
- Cleanup verification

### Example Composables
- UseCounter (simple state)
- UseForm (complex state)
- UseAsync (async operations)
- UseMouse (event handling)
- UseLocalStorage (persistence)

## Atomic Design Level
**Foundation + Utilities**

- **Foundation:** Composable pattern enables all reusable logic
- **Utilities:** Standard composables provide common functionality

## Related Components
- Uses: Reactivity (Ref, Computed, Watch)
- Uses: Component Context
- Uses: Lifecycle hooks
- Enables: All built-in components and directives

## Technical Constraints
- Composables must be called in Setup or other composables
- Cannot use composables in template functions
- Provide/inject must follow component tree hierarchy
- No global composable state (must use Context)

## API Design

### Composable Function Signature
```go
// Basic composable
func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
    count := ctx.Ref(initial)
    
    increment := func() {
        count.Set(count.Get() + 1)
    }
    
    decrement := func() {
        count.Set(count.Get() - 1)
    }
    
    return count, increment, decrement
}

// Usage in component
Setup(func(ctx *Context) {
    count, increment, decrement := UseCounter(ctx, 0)
    
    ctx.Expose("count", count)
    ctx.On("increment", func(_ interface{}) {
        increment()
    })
    ctx.On("decrement", func(_ interface{}) {
        decrement()
    })
})
```

### Standard Composable: UseState
```go
type UseStateReturn[T any] struct {
    Value *Ref[T]
    Set   func(T)
}

func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    value := ctx.Ref(initial)
    
    return UseStateReturn[T]{
        Value: value,
        Set:   func(newVal T) { value.Set(newVal) },
    }
}

// Usage
Setup(func(ctx *Context) {
    name := UseState(ctx, "")
    age := UseState(ctx, 0)
    
    ctx.Expose("name", name.Value)
    ctx.Expose("age", age.Value)
})
```

### Standard Composable: UseAsync
```go
type UseAsyncReturn[T any] struct {
    Data    *Ref[*T]
    Loading *Ref[bool]
    Error   *Ref[error]
    Execute func()
}

func UseAsync[T any](ctx *Context, fetcher func() (*T, error)) UseAsyncReturn[T] {
    data := ctx.Ref[*T](nil)
    loading := ctx.Ref(false)
    error := ctx.Ref[error](nil)
    
    execute := func() {
        loading.Set(true)
        error.Set(nil)
        
        go func() {
            result, err := fetcher()
            if err != nil {
                error.Set(err)
            } else {
                data.Set(result)
            }
            loading.Set(false)
        }()
    }
    
    return UseAsyncReturn[T]{
        Data:    data,
        Loading: loading,
        Error:   error,
        Execute: execute,
    }
}

// Usage
Setup(func(ctx *Context) {
    userData := UseAsync(ctx, func() (*User, error) {
        return fetchUser()
    })
    
    ctx.OnMounted(func() {
        userData.Execute()
    })
    
    ctx.Expose("user", userData.Data)
    ctx.Expose("loading", userData.Loading)
})
```

### Provide/Inject API
```go
// Provider component
Setup(func(ctx *Context) {
    theme := ctx.Ref("dark")
    user := ctx.Ref[*User](nil)
    
    // Provide values to children
    ctx.Provide("theme", theme)
    ctx.Provide("currentUser", user)
})

// Consumer component (child/grandchild)
Setup(func(ctx *Context) {
    // Inject provided values
    theme := ctx.Inject[*Ref[string]]("theme", nil)
    user := ctx.Inject[*Ref[*User]]("currentUser", nil)
    
    if theme != nil {
        ctx.Expose("theme", theme)
    }
    if user != nil {
        ctx.Expose("user", user)
    }
})

// With default value
Setup(func(ctx *Context) {
    defaultTheme := ctx.Ref("light")
    theme := ctx.Inject("theme", defaultTheme)
    
    ctx.Expose("theme", theme)
})
```

### Composable Composition
```go
// Low-level composable
func UseEventListener(ctx *Context, event string, handler func()) func() {
    ctx.OnMounted(func() {
        // Register event listener
    })
    
    cleanup := func() {
        // Remove event listener
    }
    
    ctx.OnUnmounted(cleanup)
    
    return cleanup
}

// High-level composable using low-level one
func UseMouse(ctx *Context) (*Ref[int], *Ref[int]) {
    x := ctx.Ref(0)
    y := ctx.Ref(0)
    
    UseEventListener(ctx, "mousemove", func() {
        // Update x, y from mouse position
    })
    
    return x, y
}

// Usage
Setup(func(ctx *Context) {
    x, y := UseMouse(ctx)
    
    ctx.Expose("mouseX", x)
    ctx.Expose("mouseY", y)
})
```

## Performance Benchmarks
```go
BenchmarkComposableCall     10000000   100 ns/op   32 B/op   1 allocs/op
BenchmarkUseState           5000000    200 ns/op   64 B/op   2 allocs/op
BenchmarkUseAsync           2000000    800 ns/op   256 B/op  8 allocs/op
BenchmarkProvideInject      1000000    500 ns/op   128 B/op  4 allocs/op
BenchmarkComposableChain    5000000    300 ns/op   96 B/op   3 allocs/op
```

## Documentation Requirements
- [ ] Package godoc with Composition API overview
- [ ] Each composable documented
- [ ] Composable pattern guide
- [ ] Provide/inject guide
- [ ] 20+ runnable examples
- [ ] Best practices document
- [ ] Common patterns
- [ ] Migration guide from inline logic

## Success Metrics
- Developers create reusable composables
- Code reuse increases
- Component logic more organized
- Composables tested independently
- Community creates composable libraries
- Type safety maintained

## Standard Composables Library

### UseState
```go
func UseState[T any](ctx *Context, initial T) UseStateReturn[T]
```
Simple state management with getter/setter.

### UseEffect
```go
func UseEffect(ctx *Context, effect func(), deps ...*Ref[any])
```
Side effect management with dependency tracking.

### UseAsync
```go
func UseAsync[T any](ctx *Context, fetcher func() (*T, error)) UseAsyncReturn[T]
```
Async data fetching with loading/error states.

### UseDebounce
```go
func UseDebounce[T any](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T]
```
Debounced reactive value.

### UseThrottle
```go
func UseThrottle(ctx *Context, fn func(), delay time.Duration) func()
```
Throttled function execution.

### UseLocalStorage
```go
func UseLocalStorage[T any](ctx *Context, key string, initial T) UseStateReturn[T]
```
State persisted to local storage.

### UseEventListener
```go
func UseEventListener(ctx *Context, event string, handler func()) func()
```
Event listener with auto-cleanup.

### UseInterval
```go
func UseInterval(ctx *Context, callback func(), delay time.Duration) func()
```
Interval timer with auto-cleanup.

### UseTimeout
```go
func UseTimeout(ctx *Context, callback func(), delay time.Duration) func()
```
Timeout with auto-cleanup.

### UseMouse
```go
func UseMouse(ctx *Context) (*Ref[int], *Ref[int])
```
Mouse position tracking.

### UseKeyboard
```go
func UseKeyboard(ctx *Context, key string) *Ref[bool]
```
Keyboard key state tracking.

### UseForm
```go
func UseForm[T any](ctx *Context, initial T, validate func(T) map[string]string) UseFormReturn[T]
```
Form state and validation.

## Example Usage Patterns

### Pattern 1: Extract Reusable Logic
```go
// Before: Inline logic
Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    increment := func() { count.Set(count.Get() + 1) }
    decrement := func() { count.Set(count.Get() - 1) }
    // ... use count, increment, decrement
})

// After: Composable
Setup(func(ctx *Context) {
    count, increment, decrement := UseCounter(ctx, 0)
    // ... use count, increment, decrement
})
```

### Pattern 2: Share Logic Between Components
```go
// Composable defined once
func UseAuth(ctx *Context) UseAuthReturn {
    user := ctx.Inject[*Ref[*User]]("currentUser", ctx.Ref[*User](nil))
    isAuthenticated := ctx.Computed(func() bool {
        return user.Get() != nil
    })
    
    return UseAuthReturn{
        User: user,
        IsAuthenticated: isAuthenticated,
    }
}

// Used in multiple components
Setup(func(ctx *Context) {
    auth := UseAuth(ctx)
    ctx.Expose("isAuthenticated", auth.IsAuthenticated)
})
```

### Pattern 3: Composable Chains
```go
// Low-level composable
func UseAsyncState[T any](ctx *Context, fetcher func() (*T, error)) UseAsyncReturn[T] {
    // Implementation
}

// Mid-level composable
func UseResource[T any](ctx *Context, url string) UseAsyncReturn[T] {
    return UseAsyncState(ctx, func() (*T, error) {
        return fetch[T](url)
    })
}

// High-level composable
func UseUserData(ctx *Context) UseAsyncReturn[User] {
    return UseResource[User](ctx, "/api/user")
}
```

## Open Questions
1. Should composables be able to call each other recursively?
2. How to handle async composables (Promise-like)?
3. Should provide/inject support reactive unwrapping?
4. Global composable registry for discoverability?
5. Dev tools for visualizing composable usage?
6. Should composables support suspense-like patterns?
