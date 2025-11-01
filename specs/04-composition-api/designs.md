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

## Known Limitations & Solutions

### UseEffect Dependency Type Constraint

**Problem:**  
UseEffect requires dependencies as `*Ref[any]`, but users naturally create typed refs like `*Ref[int]`. Go's type system doesn't support variance (covariance/contravariance), so `*Ref[int]` and `*Ref[any]` are completely different, incompatible types with no subtype relationship.

**Root Cause:**  
Go generics are invariant by design. Unlike languages with variance support, Go cannot treat `*Ref[T]` as a subtype of `*Ref[any]` even though they have identical memory layouts. This is intentional to maintain type safety and avoid the complexities of variance (see [Go Issue #7512](https://github.com/golang/go/issues/7512) and [Mero's Blog on Variance](https://blog.merovius.de/posts/2018-06-03-why-doesnt-go-have-variance-in/)).

**Attempted Solutions:**
1. **Unsafe pointer conversion** - Tried converting `*Ref[int]` to `*Ref[any]` using `unsafe.Pointer`. Failed because calling `.Get()` on the converted pointer returns values wrapped in different interface types, causing `deepEqual` comparison failures in the lifecycle system.
2. **Reflection-based conversion** - Similar issues with type mismatches during value comparison.

**Evaluated Alternatives:**

| Solution | Pros | Cons | Recommendation |
|----------|------|------|----------------|
| **1. Interface-based (Dependency interface)** | • Go-idiomatic<br>• Works with any Ref type<br>• Type-safe at interface level<br>• Enables watching Computed values | • Requires new interface<br>• Slight performance overhead (interface method calls)<br>• API change | ⭐ **RECOMMENDED** for future enhancement |
| **2. Helper function (ToRefAny)** | • Explicit conversion<br>• No API changes<br>• Users understand what's happening | • Verbose<br>• Error-prone (easy to forget)<br>• Doesn't solve root issue | ❌ Not recommended |
| **3. Current approach (Ref[any] only)** | • Simple<br>• Type-safe<br>• No complexity | • Less ergonomic<br>• Users must use `NewRef[any](value)` | ✅ **CURRENT** - acceptable trade-off |
| **4. Functional (getter functions)** | • Flexible<br>• Works with any source | • Loses dependency tracking<br>• Not declarative<br>• Breaks Vue/React patterns | ❌ Not recommended |

**Current Implementation:**  
UseEffect accepts `deps ...*Ref[any]`. Users must create refs as `NewRef[any](value)` when using them with UseEffect. This follows Go's `context.Context` pattern: store as `any`, type assert on retrieval.

**Usage Pattern:**
```go
// Create ref as Ref[any] for use with UseEffect
count := bubbly.NewRef[any](0)

UseEffect(ctx, func() UseEffectCleanup {
    // Type assert when retrieving
    currentCount := count.Get().(int)
    fmt.Printf("Count: %d\n", currentCount)
    return nil
}, count)
```

**Recommended Future Solution:**  
Implement a `Dependency` interface that both `Ref[T]` and `Computed[T]` implement:

```go
// Dependency represents a reactive dependency that can be watched
type Dependency interface {
    Get() any
    // Internal methods for dependency tracking
    AddDependent(dep Dependency)
    Invalidate()
}

// Update UseEffect signature
func UseEffect(ctx *Context, effect func() UseEffectCleanup, deps ...Dependency)
```

**Benefits:**
- Works with any Ref type (`*Ref[int]`, `*Ref[string]`, etc.)
- Enables watching Computed values (currently not possible)
- Go-idiomatic interface-based design
- Aligns with Vue 3 behavior (computed values are watchable)
- Minimal performance impact (interface method calls are highly optimized in Go)

**Priority:** MEDIUM - Current solution works, but interface-based approach would significantly improve developer experience and enable new features.

**References:**
- [Go Issue #7512 - Covariance Support](https://github.com/golang/go/issues/7512)
- [Why Go Doesn't Have Variance](https://blog.merovius.de/posts/2018-06-03-why-doesnt-go-have-variance-in/)
- [Go Generics Design](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md)
- [ByteDance gg Library](https://github.com/bytedance/gg) - Generic utilities showing Go patterns

---

## Phase 8: Performance Optimization & Monitoring Architecture

### 1. Timer Pool Design

**Problem:** UseDebounce/UseThrottle create new `time.Timer` instances for each composable (865ns and 473ns overhead). While acceptable, timer creation can be optimized with pooling.

**Architecture:**
```
┌─────────────────────────────────────────┐
│         Timer Pool Manager              │
├─────────────────────────────────────────┤
│  ┌──────────────────────────────────┐  │
│  │  sync.Pool[*time.Timer]          │  │
│  │  - Get() → reuse or create       │  │
│  │  - Put(timer) → stop & return    │  │
│  └──────────────────────────────────┘  │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  Cleanup Tracking                │  │
│  │  - Component → []timer mapping   │  │
│  │  - OnUnmounted hook cleanup      │  │
│  └──────────────────────────────────┘  │
└─────────────────────────────────────────┘
         ↓                    ↓
   UseDebounce          UseThrottle
```

**Type Definitions:**
```go
// Timer pool with automatic cleanup
type TimerPool struct {
    pool     *sync.Pool
    active   map[*time.Timer]bool
    mu       sync.RWMutex
}

func NewTimerPool() *TimerPool {
    return &TimerPool{
        pool: &sync.Pool{
            New: func() interface{} {
                return time.NewTimer(0)
            },
        },
        active: make(map[*time.Timer]bool),
    }
}

func (tp *TimerPool) Acquire(d time.Duration) *time.Timer
func (tp *TimerPool) Release(timer *time.Timer)
func (tp *TimerPool) Stats() TimerPoolStats
```

**Integration:**
```go
// Global timer pool (optional opt-in)
var globalTimerPool = NewTimerPool()

func UseDebounce[T any](ctx *Context, source *Ref[T], delay time.Duration) *Ref[T] {
    // Option 1: Use global pool (automatic)
    timer := globalTimerPool.Acquire(delay)
    
    // Register cleanup
    ctx.OnUnmounted(func() {
        globalTimerPool.Release(timer)
    })
    
    // Rest of implementation...
}
```

**Benefits:**
- Reduces allocation overhead from 865ns → ~450ns (52% improvement)
- Zero allocations after pool warmup
- Automatic cleanup on unmount
- Thread-safe with RWMutex

**Priority:** Low (current performance already acceptable)

---

### 2. Reflection Cache Design

**Problem:** UseForm performs reflection field lookup on every `SetField()` call (422ns). Field indices can be cached by struct type for ~100ns reduction.

**Architecture:**
```
┌──────────────────────────────────────────────┐
│        Reflection Cache Manager               │
├──────────────────────────────────────────────┤
│  Type-safe field index cache                 │
│                                               │
│  map[reflect.Type]map[string]int             │
│  ↓                                            │
│  struct type → field name → field index      │
│                                               │
│  ┌─────────────────────────────────────┐    │
│  │  Cache Entry                         │    │
│  │  - FieldIndices: map[string]int      │    │
│  │  - FieldTypes: map[string]reflect.Type│   │
│  │  - Computed once, reused forever     │    │
│  └─────────────────────────────────────┘    │
└──────────────────────────────────────────────┘
```

**Type Definitions:**
```go
type FieldCache struct {
    cache map[reflect.Type]*FieldCacheEntry
    mu    sync.RWMutex
}

type FieldCacheEntry struct {
    Indices map[string]int           // field name → index
    Types   map[string]reflect.Type  // field name → type
}

func (fc *FieldCache) GetFieldIndex(structType reflect.Type, fieldName string) (int, bool)
func (fc *FieldCache) GetFieldType(structType reflect.Type, fieldName string) (reflect.Type, bool)
func (fc *FieldCache) CacheType(structType reflect.Type) *FieldCacheEntry
```

**Integration:**
```go
var globalFieldCache = NewFieldCache()

func (f *UseFormReturn[T]) SetField(field string, value interface{}) {
    formType := reflect.TypeOf(f.values.GetTyped())
    
    // Fast path: cache hit (~5ns)
    if idx, ok := globalFieldCache.GetFieldIndex(formType, field); ok {
        // Direct field access by index
        formValue := reflect.ValueOf(&f.values.value).Elem()
        fieldValue := formValue.Field(idx)
        fieldValue.Set(reflect.ValueOf(value))
        return
    }
    
    // Slow path: cache miss, populate cache + set field
    globalFieldCache.CacheType(formType)
    // ... existing reflection logic ...
}
```

**Benefits:**
- Reduces SetField from 422ns → ~300ns (29% improvement)
- Cache hit rate > 95% in typical usage
- One-time reflection cost per struct type
- Thread-safe with RWMutex

**Priority:** Low (current performance already acceptable)

---

### 3. Monitoring & Metrics Architecture

**Integration Pattern:**
```
┌────────────────────────────────────────────┐
│     Application Composables                 │
│  UseState, UseAsync, UseForm, etc.         │
└────────────────┬───────────────────────────┘
                 │ Usage events
                 ↓
┌────────────────────────────────────────────┐
│     Metrics Collector (Optional)            │
│  - Composable creation count                │
│  - Performance counters                     │
│  - Tree depth tracking                      │
│  - Memory allocation stats                  │
└────────────────┬───────────────────────────┘
                 │ Metrics export
                 ↓
┌────────────────────────────────────────────┐
│     Monitoring Backend (Pluggable)          │
│  - Prometheus (default)                     │
│  - StatsD                                   │
│  - Custom exporters                         │
└────────────────────────────────────────────┘
```

**Type Definitions:**
```go
// Metrics interface for monitoring
type ComposableMetrics interface {
    RecordComposableCreation(name string, duration time.Duration)
    RecordProvideInjectDepth(depth int)
    RecordAllocationBytes(composable string, bytes int64)
    RecordCacheHit(cache string)
    RecordCacheMiss(cache string)
}

// Prometheus implementation
type PrometheusMetrics struct {
    composableCreations *prometheus.CounterVec
    provideInjectDepth  prometheus.Histogram
    allocationBytes     *prometheus.HistogramVec
    cacheHits           *prometheus.CounterVec
}

// Global metrics (optional, nil by default)
var globalMetrics ComposableMetrics

func SetMetrics(m ComposableMetrics) {
    globalMetrics = m
}
```

**Integration Points:**
```go
func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    start := time.Now()
    defer func() {
        if globalMetrics != nil {
            globalMetrics.RecordComposableCreation("UseState", time.Since(start))
        }
    }()
    
    // ... existing implementation ...
}

func (ctx *Context) Inject(key string, defaultValue interface{}) interface{} {
    depth := ctx.calculateTreeDepth()
    if globalMetrics != nil {
        globalMetrics.RecordProvideInjectDepth(depth)
    }
    
    // ... existing implementation ...
}
```

**Monitoring Dashboards:**
- Composable usage patterns (histogram)
- Performance trends over time
- Tree depth distribution
- Cache hit rates
- Memory allocation patterns

**Priority:** Medium (valuable for production deployments)

---

### 4. Performance Regression Testing

**CI/CD Integration:**
```yaml
# .github/workflows/benchmark.yml
name: Performance Benchmarks
on: [pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -count=10 \
            ./pkg/bubbly/composables/ > new.txt
          
      - name: Compare with baseline
        run: |
          benchstat baseline.txt new.txt
          
      - name: Fail on regression
        run: |
          # Fail if any benchmark regresses > 10%
          benchstat -delta-test=ttest baseline.txt new.txt | \
            grep -E '\+[0-9]{2}\.' && exit 1 || exit 0
```

**Baseline Management:**
```bash
# Update baseline after approved changes
go test -bench=. -benchmem -count=10 \
  ./pkg/bubbly/composables/ > benchmarks/baseline.txt
```

**Statistical Analysis:**
```bash
# Run with -count=10 for statistical significance
go test -bench=BenchmarkUseState -benchmem -count=10

# Analyze variance with benchstat
benchstat results.txt
```

**Priority:** High (prevents performance regressions)

---

### 5. Profiling Utilities

**Production Profiling:**
```go
package monitoring

import (
    "net/http"
    _ "net/http/pprof"
)

// Enable profiling endpoint (optional)
func EnableProfiling(addr string) error {
    return http.ListenAndServe(addr, nil)
}
```

**Usage:**
```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**Custom Profiling Utilities:**
```go
// Profile composable performance in production
func ProfileComposables(duration time.Duration) *ComposableProfile {
    profile := &ComposableProfile{
        Start: time.Now(),
        Calls: make(map[string]*CallStats),
    }
    
    // Collect metrics for duration
    time.Sleep(duration)
    
    profile.End = time.Now()
    return profile
}
```

**Priority:** Medium (useful for production debugging)

---

## Future Enhancements

1. **Dependency Interface:** Implement interface-based dependency tracking (see Known Limitations) ✅ COMPLETE (Phase 7)
2. **Timer Pooling:** Reduce debounce/throttle overhead (see Phase 8)
3. **Reflection Caching:** Optimize UseForm SetField (see Phase 8)
4. **Monitoring Integration:** Production metrics and profiling (see Phase 8)
5. **Composable Registry:** Global registry for discoverability
6. **Async Composables:** Support for async/await patterns
7. **Suspense:** React-like suspense for async composables
8. **Dev Tools:** Visualize composable usage and dependencies
9. **Hot Reload:** Update composables without full reload
10. **Testing Utilities:** Helper functions for testing composables
