# Design Specification: Reactivity System

## Component Hierarchy

```
Foundation (No UI components at this level)
└── Reactivity Primitives
    ├── Ref[T]        (Atom-level primitive)
    ├── Computed[T]   (Atom-level primitive)
    └── Watcher       (Atom-level primitive)
```

This is a foundational system that components will use, not a visual component itself.

---

## Architecture Overview

### Core Abstractions

```
┌─────────────────────────────────────────────────────────────┐
│                    Reactivity System                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐      ┌──────────────┐      ┌─────────────┐  │
│  │ Ref[T]   │─────>│ Dependency   │<─────│ Computed[T] │  │
│  │          │      │ Tracker      │      │             │  │
│  └──────────┘      └──────────────┘      └─────────────┘  │
│       │                                          │         │
│       │                                          │         │
│       └──────────────┐              ┌───────────┘         │
│                      │              │                      │
│                      ▼              ▼                      │
│                  ┌────────────────────┐                   │
│                  │   Watcher System   │                   │
│                  └────────────────────┘                   │
│                           │                               │
└───────────────────────────┼───────────────────────────────┘
                            │
                            ▼
                  ┌────────────────────┐
                  │  Bubbletea Cmd     │
                  │  (triggers update) │
                  └────────────────────┘
```

---

## Data Flow

### 1. Ref Value Change Flow
```
User Code: ref.Set(newValue)
    ↓
Ref: Acquire write lock
    ↓
Ref: Update internal value
    ↓
Ref: Release lock
    ↓
Ref: Notify all watchers
    ↓
Watcher: Execute callback(newValue, oldValue)
    ↓
Callback: May trigger Bubbletea Cmd
    ↓
Bubbletea: Schedule re-render
```

### 2. Computed Value Access Flow
```
User Code: computed.Get()
    ↓
Computed: Check if cache is dirty
    ↓
If dirty:
    ├─> Enable dependency tracking
    ├─> Execute compute function
    ├─> Track accessed Refs
    ├─> Cache result
    └─> Mark as clean
    ↓
Return cached value
```

### 3. Dependency Tracking Flow
```
Computed function executes
    ↓
Function calls ref.Get()
    ↓
Global tracker records: ref → computed
    ↓
When ref changes:
    ├─> Notify computed (mark dirty)
    └─> Notify watchers
```

---

## State Management

### Ref[T] State
```go
type Ref[T any] struct {
    mu       sync.RWMutex      // Thread safety
    value    T                 // Current value
    watchers []*watcher[T]     // Registered watchers
    deps     []Dependency      // What depends on this Ref
}
```

### Computed[T] State
```go
type Computed[T any] struct {
    mu       sync.RWMutex      // Thread safety
    fn       func() T          // Computation function
    cache    T                 // Cached result
    dirty    bool              // Needs recomputation
    deps     []Dependency      // What this depends on
    watchers []*watcher[T]     // Registered watchers
}
```

### Watcher State
```go
type watcher[T any] struct {
    callback func(newVal, oldVal T)  // User callback
    options  WatchOptions              // Configuration
    cleanup  func()                    // Cleanup function
}

type WatchOptions struct {
    Deep      bool    // Watch nested changes
    Immediate bool    // Execute immediately
    Flush     string  // "sync" or "post"
}
```

### Dependency Tracker State
```go
type DepTracker struct {
    mu          sync.RWMutex
    tracking    bool           // Currently tracking
    currentDeps []Dependency   // Dependencies being tracked
    stack       []Dependency   // For nested tracking
}

var globalTracker = &DepTracker{}
```

---

## Type Definitions

### Core Types
```go
// Ref is a type-safe reactive primitive
type Ref[T any] struct {
    mu       sync.RWMutex
    value    T
    watchers []*watcher[T]
    deps     []Dependency
}

// Computed is a derived reactive value
type Computed[T any] struct {
    mu       sync.RWMutex
    fn       func() T
    cache    T
    dirty    bool
    deps     []Dependency
    watchers []*watcher[T]
}

// Dependency represents something that depends on a reactive value
type Dependency interface {
    Invalidate()  // Mark as needing recomputation
}

// WatchCallback is called when watched value changes
type WatchCallback[T any] func(newVal, oldVal T)

// WatchCleanup stops watching when called
type WatchCleanup func()
```

### Public API Types
```go
// NewRef creates a new reactive reference
func NewRef[T any](value T) *Ref[T]

// Get returns the current value (with dependency tracking)
func (r *Ref[T]) Get() T

// Set updates the value and triggers watchers
func (r *Ref[T]) Set(value T)

// NewComputed creates a computed value
func NewComputed[T any](fn func() T) *Computed[T]

// Get returns computed value (lazy evaluation)
func (c *Computed[T]) Get() T

// Watch creates a watcher
func Watch[T any](
    source *Ref[T],
    callback WatchCallback[T],
    options ...WatchOption,
) WatchCleanup

// WatchOption configures watcher behavior
type WatchOption func(*WatchOptions)

func WithDeep() WatchOption
func WithImmediate() WatchOption
```

---

## API Contracts

### Ref API
```go
// Constructor
ref := NewRef(42)           // Create with initial value
ref := NewRef("hello")      // Works with any type
ref := NewRef[int](0)       // Explicit type parameter

// Getters
value := ref.Get()          // Thread-safe read

// Setters
ref.Set(100)                // Thread-safe write
ref.Set(ref.Get() + 1)      // Atomic increment pattern

// Advanced
ref.Update(func(current int) int {  // Atomic update with function
    return current + 1
})
```

### Computed API
```go
// Constructor
count := NewRef(0)
doubled := NewComputed(func() int {
    return count.Get() * 2   // Auto-tracks dependency on count
})

// Getter (lazy)
value := doubled.Get()       // Computes if dirty, else returns cache

// Chaining
quadrupled := NewComputed(func() int {
    return doubled.Get() * 2  // Can depend on other computed
})
```

### Watcher API
```go
// Basic watcher
cleanup := Watch(ref, func(newVal, oldVal int) {
    fmt.Printf("Changed: %d → %d\n", oldVal, newVal)
})
defer cleanup()  // Stop watching

// Immediate execution
Watch(ref, callback, WithImmediate())

// Deep watching (for nested structs)
Watch(ref, callback, WithDeep())

// Multiple options
Watch(ref, callback, WithDeep(), WithImmediate())
```

---

## Implementation Details

### Thread Safety
```go
// Ref Get (read lock)
func (r *Ref[T]) Get() T {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    // Track dependency if currently tracking
    if globalTracker.IsTracking() {
        globalTracker.Track(r)
    }
    
    return r.value
}

// Ref Set (write lock)
func (r *Ref[T]) Set(value T) {
    r.mu.Lock()
    oldValue := r.value
    r.value = value
    watchers := r.watchers  // Copy under lock
    r.mu.Unlock()
    
    // Notify outside of lock
    r.notifyWatchers(value, oldValue, watchers)
}
```

### Dependency Tracking
```go
// Enable tracking during computed evaluation
func (c *Computed[T]) Get() T {
    c.mu.Lock()
    if !c.dirty {
        val := c.cache
        c.mu.Unlock()
        return val
    }
    c.mu.Unlock()
    
    // Start tracking
    globalTracker.BeginTracking()
    result := c.fn()  // Calls ref.Get() internally
    deps := globalTracker.EndTracking()
    
    c.mu.Lock()
    c.cache = result
    c.dirty = false
    c.deps = deps
    c.mu.Unlock()
    
    // Register with dependencies
    for _, dep := range deps {
        dep.AddDependent(c)
    }
    
    return result
}
```

### Circular Dependency Detection
```go
type DepTracker struct {
    stack []Dependency  // Currently evaluating
}

func (dt *DepTracker) BeginTracking(dep Dependency) error {
    for _, d := range dt.stack {
        if d == dep {
            return ErrCircularDependency
        }
    }
    dt.stack = append(dt.stack, dep)
    return nil
}
```

---

## Integration with Bubbletea

### Triggering Updates
```go
// In component context
type ComponentContext struct {
    sendMsg func(tea.Msg)
}

// Watcher triggers Bubbletea message
func (ctx *ComponentContext) Watch(ref *Ref[int], callback func(int, int)) {
    Watch(ref, func(newVal, oldVal int) {
        callback(newVal, oldVal)
        // Trigger re-render via Bubbletea
        ctx.sendMsg(tea.Msg(RefChangedMsg{Name: "ref"}))
    })
}
```

### Component State Pattern
```go
type Component struct {
    state map[string]any  // Stores Refs
}

func (c *Component) Setup(ctx *Context) {
    // Create reactive state
    count := ctx.Ref(0)
    
    // Expose to template
    ctx.Expose("count", count)
    
    // Watch for changes
    ctx.Watch(count, func(newVal, oldVal int) {
        log.Printf("Count changed: %d → %d", oldVal, newVal)
    })
}
```

---

## Error Handling

### Error Types
```go
var (
    ErrCircularDependency = errors.New("circular dependency detected")
    ErrInvalidType       = errors.New("invalid type for operation")
    ErrNilValue          = errors.New("cannot set nil value on non-pointer ref")
)
```

### Error Scenarios
1. **Circular Dependency:** Return error from Computed.Get()
2. **Nil Values:** Check in Set() if T is not pointer type
3. **Max Depth:** Limit dependency chain depth (e.g., 100)

---

## Performance Optimizations

### 1. Lock Granularity
- RWMutex for reads (multiple concurrent readers)
- Write lock only during Set()
- Notify watchers outside lock

### 2. Lazy Evaluation
- Computed values only evaluate when accessed
- Cache results until dependencies change

### 3. Batch Updates
```go
// Batch multiple updates to avoid redundant recomputes
func Batch(fn func()) {
    globalTracker.BeginBatch()
    defer globalTracker.EndBatch()
    fn()
}
```

### 4. Memory Pool
```go
var watcherPool = sync.Pool{
    New: func() interface{} {
        return &watcher{}
    },
}
```

---

## Testing Strategy

### Unit Tests
```go
func TestRef_GetSet(t *testing.T)
func TestRef_Concurrent(t *testing.T)
func TestComputed_AutoTracking(t *testing.T)
func TestComputed_Caching(t *testing.T)
func TestWatch_Notification(t *testing.T)
func TestWatch_Cleanup(t *testing.T)
func TestDepTracker_CircularDetection(t *testing.T)
```

### Race Detection
```bash
go test -race ./pkg/bubbly/reactivity
```

### Benchmarks
```go
func BenchmarkRef_Get(b *testing.B)
func BenchmarkRef_Set(b *testing.B)
func BenchmarkComputed_Get(b *testing.B)
func BenchmarkWatch_Notify(b *testing.B)
```

---

## Example Usage

### Simple Counter
```go
count := NewRef(0)
doubled := NewComputed(func() int {
    return count.Get() * 2
})

Watch(count, func(newVal, oldVal int) {
    fmt.Printf("Count: %d, Doubled: %d\n", newVal, doubled.Get())
})

count.Set(5)   // Prints: Count: 5, Doubled: 10
count.Set(10)  // Prints: Count: 10, Doubled: 20
```

### Todo List
```go
todos := NewRef([]string{"Task 1", "Task 2"})
total := NewComputed(func() int {
    return len(todos.Get())
})

Watch(todos, func(newVal, oldVal []string) {
    fmt.Printf("Todo count: %d\n", total.Get())
}, WithImmediate())

todos.Set(append(todos.Get(), "Task 3"))
// Prints: Todo count: 3
```

---

## Migration Path

### From Manual State
```go
// Before (manual)
type model struct {
    count int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case IncrementMsg:
        m.count++  // Manual update
    }
    return m, nil
}

// After (reactive)
type model struct {
    count *Ref[int]
}

func (m model) Setup(ctx *Context) {
    m.count = ctx.Ref(0)
    ctx.On("increment", func() {
        m.count.Set(m.count.Get() + 1)  // Auto-updates UI
    })
}
```

---

## Future Enhancements

1. **Reactive Collections:** `RefArray[T]`, `RefMap[K,V]`
2. **Shallow Refs:** `ShallowRef[T]` for large objects
3. **Readonly Refs:** `Readonly[T]` for immutable exposure
4. **Effect Scheduling:** Control when effects run (sync, async, debounced)
5. **Dev Tools:** Visualize reactive graph
