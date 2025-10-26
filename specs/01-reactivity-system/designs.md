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
    Immediate   bool        // Execute immediately (✅ Task 3.2)
    Deep        bool        // Watch nested changes (✅ Task 3.3)
    DeepCompare interface{} // Custom deep comparator (✅ Task 3.3)
    Flush       string      // "sync" or "post" (✅ Task 3.4)
}

// Implementation status:
// - Immediate: ✅ Fully implemented (Task 3.2)
// - Deep: ✅ Fully implemented (Task 3.3) - reflection-based comparison
// - DeepCompare: ✅ Fully implemented (Task 3.3) - custom comparator functions
// - Flush: ✅ Fully implemented (Task 3.4) - sync and post modes with batching
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

## Advanced Features

### Deep Watching ✅ (Task 3.3 - Complete)

**Problem:** By default, watchers only trigger when `Set()` is called on a Ref. Changes to nested fields don't trigger watchers:

```go
type User struct {
    Name string
    Profile Profile
}

user := NewRef(User{Name: "John"})
Watch(user, callback)  // Only triggers on user.Set()

// This does NOT trigger the watcher:
u := user.Get()
u.Profile.Bio = "New bio"
// user still has old value!
```

**Solution Implemented:**

1. **Reflection-based (automatic)** - ✅ Implemented:
```go
Watch(user, callback, WithDeep())
// Uses reflect.DeepEqual to detect nested changes
// Performance: ~7x slower than shallow (280ns vs 40ns)
// Only triggers if value actually changed
```

2. **Custom comparator (manual)** - ✅ Implemented:
```go
Watch(user, callback, WithDeepCompare(func(old, new User) bool {
    return old.Profile.Bio == new.Profile.Bio  // return true if equal
}))
// User controls what counts as a "change"
// Performance: ~2.5x slower than shallow (99ns vs 40ns)
```

3. **Hybrid approach** - ✅ Implemented:
```go
// Default to reflection for structs
Watch(user, callback, WithDeep())

// Override for performance-critical paths
Watch(user, callback, WithDeepCompare(customCompare))
```

**Implementation Details:**
- ✅ Store previous value in watcher struct
- ✅ On Set(), compare old vs new using deep equality
- ✅ Only trigger callback if deep comparison shows change
- ✅ Performance benchmarks provided
- ✅ Comprehensive documentation with warnings

**Edge Cases Handled:**
- ✅ Nil values and pointers
- ✅ Unexported fields (reflect.DeepEqual handles correctly)
- ✅ Empty collections (slices, maps)
- ✅ Pointer semantics (compares values, not addresses)

**Performance Benchmarks:**
```
BenchmarkWatch_Shallow-6      28958858   40.23 ns/op    0 B/op   0 allocs/op
BenchmarkWatch_DeepCompare-6  11447961   98.91 ns/op   48 B/op   1 allocs/op
BenchmarkWatch_Deep-6          3570967  280.3  ns/op  144 B/op   3 allocs/op
```

---

### Async Flush Modes ✅ (Task 3.4 - Complete)

**Problem:** Multiple rapid state changes trigger multiple watcher callbacks, causing redundant work:

```go
Watch(count, func(n, o int) {
    // Expensive operation (e.g., API call, render)
})

count.Set(1)  // Callback runs
count.Set(2)  // Callback runs again
count.Set(3)  // Callback runs again
// 3 expensive operations when we only need the final state!
```

**Solution Implemented: Post-Flush Mode**

```go
Watch(count, callback, WithFlush("post"))

count.Set(1)  // Queued
count.Set(2)  // Queued (replaces previous)
count.Set(3)  // Queued (replaces previous)

FlushWatchers()  // Callback runs once with final value (3)
```

**Implementation:**

1. **Sync mode** (default):
```go
func (r *Ref[T]) notifyWatchers(...) {
    if watcher.options.Flush == "sync" {
        watcher.callback(newVal, oldVal)  // Immediate
    }
}
```

2. **Post mode** (✅ implemented):
```go
type CallbackScheduler struct {
    mu    sync.Mutex
    queue map[interface{}]callbackFunc  // Batching via map
}

func (r *Ref[T]) notifyWatchers(...) {
    if watcher.options.Flush == "post" {
        // Queue callback (replaces previous for same watcher)
        globalScheduler.enqueue(watcher, func() {
            watcher.callback(newVal, oldVal)
        })
    }
}

// Public API to execute queued callbacks
func FlushWatchers() int {
    return globalScheduler.flush()
}
```

**Integration with Bubbletea:**
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // State changes queue callbacks
    m.count.Set(m.count.Get() + 1)
    
    // Execute all queued callbacks before returning
    FlushWatchers()
    
    return m, nil
}
```

**Benefits:**
- ✅ Reduce redundant callback executions (batching)
- ✅ Batch UI updates for better performance
- ✅ Prevent intermediate states from being rendered
- ✅ Simple map-based queue for O(1) batching
- ✅ Thread-safe implementation

**Implementation Details:**
- Global CallbackScheduler singleton
- Map-based queue: watcher → callback
- Batching: Same watcher = replace previous callback
- Type-erased callbacks using closures
- FlushWatchers() returns count executed
- PendingCallbacks() for debugging

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

## Known Limitations & Solutions

### 1. Global Tracker Contention (HIGH PRIORITY)

**Current Design:**
```go
// Single global tracker for ALL goroutines
var globalTracker = &DepTracker{
    mu sync.Mutex  // ALL goroutines contend for this lock
}
```

**Problem:**
- Single mutex causes severe contention with 100+ concurrent goroutines
- Deadlocks when many goroutines access computed values simultaneously
- Race detector timeouts in high-concurrency scenarios

**Solution Design:**
```go
// Per-goroutine tracker using sync.Map
type DepTracker struct {
    trackers sync.Map // map[goroutineID]*trackingState
}

type trackingState struct {
    mu    sync.Mutex
    stack []*trackingContext
}

func (dt *DepTracker) BeginTracking(dep Dependency) error {
    gid := getGoroutineID()  // Get current goroutine ID
    state := dt.getOrCreateState(gid)
    state.mu.Lock()  // Only locks THIS goroutine's state
    defer state.mu.Unlock()
    // ... tracking logic
}
```

**Benefits:**
- Zero contention between goroutines
- Scales to 1000+ concurrent goroutines
- No deadlocks
- Better performance

**Implementation Priority:** HIGH (before production use with high concurrency)

---

### 2. Watch Computed Values (MEDIUM PRIORITY)

**Current Design:**
```go
// Watch only accepts Ref[T]
func Watch[T any](
    source *Ref[T],  // ❌ Cannot pass *Computed[T]
    callback WatchCallback[T],
    options ...WatchOption,
) WatchCleanup
```

**Problem:**
- Cannot watch computed values directly (Vue 3 supports this)
- Requires awkward workarounds (watch underlying refs)

**Solution Design:**
```go
// Create Watchable interface
type Watchable[T any] interface {
    Get() T
    addWatcher(w *watcher[T])
    removeWatcher(w *watcher[T])
}

// Update Watch signature
func Watch[T any](
    source Watchable[T],  // ✅ Accepts Ref OR Computed
    callback WatchCallback[T],
    options ...WatchOption,
) WatchCleanup

// Computed implements Watchable
func (c *Computed[T]) addWatcher(w *watcher[T]) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.watchers == nil {
        c.watchers = make([]*watcher[T], 0, 4)
    }
    c.watchers = append(c.watchers, w)
}

func (c *Computed[T]) Get() T {
    // ... existing logic ...
    
    // After recomputation, notify watchers if value changed
    if c.dirty && !reflect.DeepEqual(oldValue, newValue) {
        c.notifyWatchers(newValue, oldValue)
    }
    
    return c.cache
}
```

**Use Cases:**
- Form validation (watch overall form validity)
- Derived state monitoring (watch computed totals)
- Business logic triggers (watch complex computed state)

**Implementation Priority:** MEDIUM (nice to have, workarounds exist)

---

## Future Enhancements

1. **Reactive Collections:** `RefArray[T]`, `RefMap[K,V]`
2. **Shallow Refs:** `ShallowRef[T]` for large objects
3. **Readonly Refs:** `Readonly[T]` for immutable exposure
4. **WatchEffect:** Automatic dependency tracking for watchers (LOW priority)
4. **Effect Scheduling:** Control when effects run (sync, async, debounced)
5. **Dev Tools:** Visualize reactive graph
