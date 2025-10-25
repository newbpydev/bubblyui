# Design Specification: Composition API

## Component Hierarchy

```
Composition API System
└── Composable Framework
    ├── Context (Extended for composables)
    ├── Composable Functions (Use* pattern)
    ├── Provide/Inject System
    └── Standard Composables Library

Composable Layers:
├── Foundation (UseState, UseEffect)
├── Async (UseAsync, UseDebounce, UseThrottle)
├── DOM (UseMouse, UseKeyboard, UseEventListener)
├── Storage (UseLocalStorage, UseSessionStorage)
└── Application (UseAuth, UseForm, UseValidation)
```

---

## Architecture Overview

### System Layers

```
┌────────────────────────────────────────────────────────────┐
│                   Composition API Layer                     │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────────┐      ┌────────────────────┐        │
│  │   Composable     │─────>│  Context System    │        │
│  │   Functions      │      │  (Extended)        │        │
│  └──────────────────┘      └────────────────────┘        │
│         │                           │                      │
│         ├──> Uses Reactivity        │                      │
│         ├──> Uses Lifecycle         │                      │
│         └──> Uses Provide/Inject <──┘                      │
│                                                            │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│              Component + Reactivity + Lifecycle            │
│  ┌──────────┐   ┌──────────┐   ┌─────────────┐          │
│  │  Ref[T]  │   │Component │   │  Lifecycle  │          │
│  │Computed  │   │ Context  │   │   Hooks     │          │
│  └──────────┘   └──────────┘   └─────────────┘          │
└────────────────────────────────────────────────────────────┘
```

### Composable Execution Flow

```
Component Setup() executes
    ↓
Context created and available
    ↓
Developer calls Composable (e.g., UseCounter)
    ↓
Composable receives Context
    ↓
Composable creates Refs (ctx.Ref)
    ↓
Composable registers hooks (ctx.OnMounted)
    ↓
Composable returns reactive values and functions
    ↓
Component uses returned values
    ↓
Component exposes to template
```

### Provide/Inject Flow

```
Parent Component
    ↓
Setup() calls ctx.Provide("key", value)
    ↓
Value stored in component's provide map
    ↓
Child Component (any depth)
    ↓
Setup() calls ctx.Inject("key", default)
    ↓
Walk up component tree looking for "key"
    ↓
Found: Return provided value
Not Found: Return default value
```

---

## Data Flow

### 1. Composable Call Flow
```
Setup function
    ↓
Call UseCounter(ctx, 0)
    ↓
UseCounter creates count Ref
    ↓
UseCounter creates increment/decrement functions
    ↓
UseCounter returns (count, increment, decrement)
    ↓
Setup receives returned values
    ↓
Setup exposes to template
```

### 2. Composable Chain Flow
```
High-level composable: UseUserData(ctx)
    ↓
Calls mid-level: UseResource(ctx, "/api/user")
    ↓
Calls low-level: UseAsync(ctx, fetchFunc)
    ↓
UseAsync creates Refs (data, loading, error)
    ↓
Values propagate back up chain
    ↓
High-level composable returns final interface
```

### 3. Provide/Inject Data Flow
```
Root Component provides "theme" → Ref("dark")
    ↓
Component tree established
    ↓
Deep child injects "theme"
    ↓
System walks up tree: child → parent → grandparent → root
    ↓
Found at root: return theme Ref
    ↓
Child watches theme Ref
    ↓
Root changes theme.Set("light")
    ↓
Child automatically sees new value
```

---

## State Management

### Extended Context Structure
```go
type Context struct {
    component *componentImpl
    
    // Existing APIs
    Ref       func(value interface{}) *Ref[interface{}]
    Computed  func(fn func() interface{}) *Computed[interface{}]
    Watch     func(ref *Ref[interface{}], callback WatchCallback)
    OnMounted func(hook func())
    OnUpdated func(hook func(), deps ...*Ref[any])
    OnUnmounted func(hook func())
    
    // Composition API additions
    Provide   func(key string, value interface{})
    Inject    func(key string, defaultValue interface{}) interface{}
    
    // Internal
    provides  map[string]interface{}
}
```

### Composable Return Pattern
```go
// Standard return struct
type UseStateReturn[T any] struct {
    Value *Ref[T]
    Set   func(T)
}

// Async return struct
type UseAsyncReturn[T any] struct {
    Data    *Ref[*T]
    Loading *Ref[bool]
    Error   *Ref[error]
    Execute func()
    Reset   func()
}

// Form return struct
type UseFormReturn[T any] struct {
    Values  *Ref[T]
    Errors  *Ref[map[string]string]
    Submit  func()
    Reset   func()
    SetField func(field string, value interface{})
}
```

### Provide/Inject Storage
```go
type componentImpl struct {
    // Existing fields...
    
    // Provide/Inject
    provides map[string]interface{}
    parent   *componentImpl  // For inject lookup
}
```

---

## Type Definitions

### Core Types
```go
// Composable function signature
type ComposableFunc func(ctx *Context) interface{}

// Generic composable with typed return
type Composable[T any] func(ctx *Context) T

// Provide key type (for type safety)
type ProvideKey[T any] struct {
    key string
}

func NewProvideKey[T any](key string) ProvideKey[T] {
    return ProvideKey[T]{key: key}
}
```

### Standard Composable Types
```go
// UseState
type UseStateReturn[T any] struct {
    Value *Ref[T]
    Set   func(T)
    Get   func() T
}

// UseEffect (similar to React useEffect)
type UseEffectCleanup func()

// UseAsync
type UseAsyncReturn[T any] struct {
    Data    *Ref[*T]
    Loading *Ref[bool]
    Error   *Ref[error]
    Execute func()
    Reset   func()
}

// UseDebounce
type UseDeb ounceReturn[T any] struct {
    Value *Ref[T]
}

// UseForm
type UseFormReturn[T any] struct {
    Values    *Ref[T]
    Errors    *Ref[map[string]string]
    Touched   *Ref[map[string]bool]
    IsValid   *Computed[bool]
    IsDirty   *Computed[bool]
    Submit    func()
    Reset     func()
    SetField  func(field string, value interface{})
}
```

---

## API Contracts

### Composable Function Contract
```go
// All composables must:
// 1. Accept Context as first parameter
// 2. Return stable references (not change on re-call)
// 3. Clean up resources via lifecycle hooks
// 4. Be type-safe

// Example contract
func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
    // Create state
    count := ctx.Ref(initial)
    
    // Create stable functions
    increment := func() {
        count.Set(count.Get() + 1)
    }
    
    decrement := func() {
        count.Set(count.Get() - 1)
    }
    
    // Return interface
    return count, increment, decrement
}
```

### Provide/Inject API
```go
// Type-safe provide/inject using keys
var ThemeKey = NewProvideKey[*Ref[string]]("theme")

// Provider
ctx.Provide(ThemeKey.key, themeRef)

// Consumer
theme := ctx.Inject(ThemeKey.key, ctx.Ref("light")).(*Ref[string])

// Or with generic helper
func ProvideTyped[T any](ctx *Context, key ProvideKey[T], value T) {
    ctx.Provide(key.key, value)
}

func InjectTyped[T any](ctx *Context, key ProvideKey[T], defaultValue T) T {
    val := ctx.Inject(key.key, defaultValue)
    return val.(T)
}
```

### Standard Composable APIs

#### UseState
```go
state := UseState(ctx, "initial")
state.Value.Get()  // Read value
state.Set("new")   // Update value
```

#### UseAsync
```go
async := UseAsync(ctx, fetchUser)

ctx.OnMounted(func() {
    async.Execute()
})

// Access results
user := async.Data.Get()
loading := async.Loading.Get()
err := async.Error.Get()
```

#### UseForm
```go
form := UseForm(ctx, UserForm{}, validateUser)

// Set field
form.SetField("email", "user@example.com")

// Submit
form.Submit()  // Triggers validation

// Check state
isValid := form.IsValid.Get()
errors := form.Errors.Get()
```

---

## Implementation Details

### Extended Context Implementation
```go
func (c *componentImpl) createContext() *Context {
    ctx := &Context{
        component: c,
        provides:  make(map[string]interface{}),
    }
    
    // Existing methods...
    ctx.Ref = func(value interface{}) *Ref[interface{}] {
        return NewRef(value)
    }
    
    // Composition API methods
    ctx.Provide = func(key string, value interface{}) {
        ctx.provides[key] = value
    }
    
    ctx.Inject = func(key string, defaultValue interface{}) interface{} {
        return c.inject(key, defaultValue)
    }
    
    return ctx
}

func (c *componentImpl) inject(key string, defaultValue interface{}) interface{} {
    // Check current component
    if val, ok := c.provides[key]; ok {
        return val
    }
    
    // Walk up parent chain
    current := c.parent
    for current != nil {
        if val, ok := current.provides[key]; ok {
            return val
        }
        current = current.parent
    }
    
    // Not found, return default
    return defaultValue
}
```

### Standard Composable Implementations

#### UseState
```go
func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    value := ctx.Ref(initial)
    
    return UseStateReturn[T]{
        Value: value,
        Set:   func(v T) { value.Set(v) },
        Get:   func() T { return value.Get() },
    }
}
```

#### UseEffect
```go
func UseEffect(ctx *Context, effect func() UseEffectCleanup, deps ...*Ref[any]) {
    var cleanup UseEffectCleanup
    
    executeEffect := func() {
        if cleanup != nil {
            cleanup()
        }
        cleanup = effect()
    }
    
    if len(deps) == 0 {
        // No deps: run on every update
        ctx.OnMounted(executeEffect)
        ctx.OnUpdated(executeEffect)
    } else {
        // With deps: run when deps change
        ctx.OnMounted(executeEffect)
        ctx.OnUpdated(executeEffect, deps...)
    }
    
    ctx.OnUnmounted(func() {
        if cleanup != nil {
            cleanup()
        }
    })
}
```

#### UseAsync
```go
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
                data.Set(nil)
            } else {
                data.Set(result)
                error.Set(nil)
            }
            loading.Set(false)
        }()
    }
    
    reset := func() {
        data.Set(nil)
        loading.Set(false)
        error.Set(nil)
    }
    
    return UseAsyncReturn[T]{
        Data:    data,
        Loading: loading,
        Error:   error,
        Execute: execute,
        Reset:   reset,
    }
}
```

#### UseDebounce
```go
func UseDebounce[T any](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T] {
    debounced := ctx.Ref(value.Get())
    var timer *time.Timer
    
    ctx.Watch(value, func(newVal, _ T) {
        if timer != nil {
            timer.Stop()
        }
        
        timer = time.AfterFunc(delay, func() {
            debounced.Set(newVal)
        })
    })
    
    ctx.OnUnmounted(func() {
        if timer != nil {
            timer.Stop()
        }
    })
    
    return debounced
}
```

#### UseForm
```go
func UseForm[T any](
    ctx *Context,
    initial T,
    validate func(T) map[string]string,
) UseFormReturn[T] {
    values := ctx.Ref(initial)
    errors := ctx.Ref(make(map[string]string))
    touched := ctx.Ref(make(map[string]bool))
    
    isValid := ctx.Computed(func() bool {
        return len(errors.Get()) == 0
    })
    
    isDirty := ctx.Computed(func() bool {
        return len(touched.Get()) > 0
    })
    
    submit := func() {
        // Validate
        errs := validate(values.Get())
        errors.Set(errs)
        
        // If valid, emit submit event
        if len(errs) == 0 {
            ctx.Emit("submit", values.Get())
        }
    }
    
    reset := func() {
        values.Set(initial)
        errors.Set(make(map[string]string))
        touched.Set(make(map[string]bool))
    }
    
    setField := func(field string, value interface{}) {
        // Update touched
        t := touched.Get()
        t[field] = true
        touched.Set(t)
        
        // Update value
        current := values.Get()
        // Use reflection or struct tags to set field
        // Implementation depends on how T is structured
        values.Set(current)
        
        // Validate
        errs := validate(values.Get())
        errors.Set(errs)
    }
    
    return UseFormReturn[T]{
        Values:   values,
        Errors:   errors,
        Touched:  touched,
        IsValid:  isValid,
        IsDirty:  isDirty,
        Submit:   submit,
        Reset:    reset,
        SetField: setField,
    }
}
```

---

## Integration with Existing Features

### With Reactivity System
```go
// Composables use Refs directly
func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
    count := ctx.Ref(initial)  // Uses Feature 01
    // ...
}
```

### With Lifecycle Hooks
```go
// Composables register lifecycle hooks
func UseInterval(ctx *Context, callback func(), delay time.Duration) func() {
    var ticker *time.Ticker
    
    ctx.OnMounted(func() {  // Uses Feature 03
        ticker = time.NewTicker(delay)
        go func() {
            for range ticker.C {
                callback()
            }
        }()
    })
    
    ctx.OnUnmounted(func() {  // Uses Feature 03
        if ticker != nil {
            ticker.Stop()
        }
    })
    
    return func() {
        if ticker != nil {
            ticker.Stop()
        }
    }
}
```

### With Component System
```go
// Composables called in Setup
NewComponent("MyComponent").
    Setup(func(ctx *Context) {  // Uses Feature 02
        // Call composables
        count, inc, dec := UseCounter(ctx, 0)
        
        // Expose to template
        ctx.Expose("count", count)
        ctx.On("increment", func(_ interface{}) { inc() })
    }).
    Build()
```

---

## Error Handling

### Error Types
```go
var (
    ErrComposableOutsideSetup = errors.New("composable called outside Setup")
    ErrCircularComposable     = errors.New("circular composable dependency")
    ErrInjectNotFound         = errors.New("inject key not found")
    ErrProvideKeyConflict     = errors.New("provide key already exists")
)
```

### Error Scenarios
1. **Composable called outside Setup:** Detect and throw error
2. **Circular dependencies:** Track call stack, detect cycles
3. **Inject without provide:** Return default or error
4. **Type mismatch in inject:** Panic with clear message

---

## Performance Optimizations

### 1. Composable Memoization
```go
// Cache composable results within same Setup call
type composableCache struct {
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (ctx *Context) memoize(key string, fn func() interface{}) interface{} {
    ctx.cache.mu.RLock()
    if val, ok := ctx.cache.cache[key]; ok {
        ctx.cache.mu.RUnlock()
        return val
    }
    ctx.cache.mu.RUnlock()
    
    val := fn()
    
    ctx.cache.mu.Lock()
    ctx.cache.cache[key] = val
    ctx.cache.mu.Unlock()
    
    return val
}
```

### 2. Inject Lookup Caching
```go
type injectCache struct {
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (c *componentImpl) inject(key string, defaultValue interface{}) interface{} {
    // Check cache first
    c.injectCache.mu.RLock()
    if val, ok := c.injectCache.cache[key]; ok {
        c.injectCache.mu.RUnlock()
        return val
    }
    c.injectCache.mu.RUnlock()
    
    // Walk tree
    val := c.walkTreeForProvide(key, defaultValue)
    
    // Cache result
    c.injectCache.mu.Lock()
    c.injectCache.cache[key] = val
    c.injectCache.mu.Unlock()
    
    return val
}
```

---

## Testing Strategy

### Unit Tests
```go
func TestUseState(t *testing.T)
func TestUseAsync(t *testing.T)
func TestUseEffect(t *testing.T)
func TestProvideInject(t *testing.T)
func TestComposableChain(t *testing.T)
func TestComposableCleanup(t *testing.T)
```

### Integration Tests
```go
func TestComposableInComponent(t *testing.T)
func TestProvideInjectAcrossTree(t *testing.T)
func TestComposableWithLifecycle(t *testing.T)
```

---

## Example Usage

### Simple Composable
```go
func UseToggle(ctx *Context, initial bool) (*Ref[bool], func()) {
    value := ctx.Ref(initial)
    
    toggle := func() {
        value.Set(!value.Get())
    }
    
    return value, toggle
}
```

### Provide/Inject Example
```go
// Root component
Setup(func(ctx *Context) {
    theme := ctx.Ref("dark")
    ctx.Provide("theme", theme)
    ctx.Expose("theme", theme)
})

// Child component (any depth)
Setup(func(ctx *Context) {
    theme := ctx.Inject("theme", ctx.Ref("light")).(*Ref[string])
    ctx.Expose("theme", theme)
})
```

### Composable Chain Example
```go
func UseAuth(ctx *Context) UseAuthReturn {
    user := ctx.Inject("currentUser", ctx.Ref[*User](nil)).(*Ref[*User])
    
    isAuthenticated := ctx.Computed(func() bool {
        return user.Get() != nil
    })
    
    isAdmin := ctx.Computed(func() bool {
        u := user.Get()
        return u != nil && u.Role == "admin"
    })
    
    return UseAuthReturn{
        User:            user,
        IsAuthenticated: isAuthenticated,
        IsAdmin:         isAdmin,
    }
}
```

---

## Future Enhancements

1. **Composable Registry:** Global registry for discoverability
2. **Async Composables:** Support for async/await patterns
3. **Suspense:** React-like suspense for async composables
4. **Dev Tools:** Visualize composable usage and dependencies
5. **Hot Reload:** Update composables without full reload
6. **Testing Utilities:** Helper functions for testing composables
