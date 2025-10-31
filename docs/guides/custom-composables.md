# Custom Composables Guide

## Overview

This guide teaches you how to create your own composables, from simple reusable functions to advanced patterns with lifecycle management, type safety, and testing.

## Table of Contents

- [Composable Anatomy](#composable-anatomy)
- [Creating Your First Composable](#creating-your-first-composable)
- [Naming Conventions](#naming-conventions)
- [Return Value Patterns](#return-value-patterns)
- [Type Safety with Generics](#type-safety-with-generics)
- [Composable Composition](#composable-composition)
- [Lifecycle Integration](#lifecycle-integration)
- [Advanced Patterns](#advanced-patterns)
- [Testing Composables](#testing-composables)
- [Best Practices](#best-practices)
- [Common Pitfalls](#common-pitfalls)
- [Real-World Examples](#real-world-examples)

## Composable Anatomy

A composable is a Go function with these characteristics:

```go
func Use<Name>(ctx *bubbly.Context, ...params) <ReturnType> {
    // 1. Create reactive state
    state := ctx.Ref(initialValue)
    
    // 2. Create computed values
    derived := ctx.Computed(func() T {
        return transform(state.GetTyped())
    })
    
    // 3. Define helper functions
    helper := func() {
        state.Set(newValue)
    }
    
    // 4. Register lifecycle hooks (optional)
    ctx.OnMounted(func() {
        // Initialize
    })
    
    ctx.OnUnmounted(func() {
        // Cleanup
    })
    
    // 5. Return reactive values and functions
    return ReturnType{
        State:  state,
        Derived: derived,
        Helper: helper,
    }
}
```

**Key elements:**
1. **Context parameter** - Always first parameter
2. **Reactive state** - Use `ctx.Ref()` and `ctx.Computed()`
3. **Helper functions** - Closures over reactive state
4. **Lifecycle hooks** - Optional setup/cleanup
5. **Return struct** - Named fields for clarity

## Creating Your First Composable

### Step 1: Define the Interface

Start by defining what your composable should do:

```go
// Goal: Create a toggle composable
// - State: boolean value
// - Actions: toggle, setTrue, setFalse
// - Return: state ref and action functions
```

### Step 2: Create the Function

```go
package myapp

import "github.com/newbpydev/bubblyui/pkg/bubbly"

func UseToggle(ctx *bubbly.Context, initial bool) (*bubbly.Ref[bool], func(), func(), func()) {
    // Create state
    state := ctx.Ref(initial)
    
    // Define actions
    toggle := func() {
        state.Set(!state.GetTyped())
    }
    
    setTrue := func() {
        state.Set(true)
    }
    
    setFalse := func() {
        state.Set(false)
    }
    
    return state, toggle, setTrue, setFalse
}
```

### Step 3: Use in Component

```go
Setup(func(ctx *bubbly.Context) {
    isOpen, toggle, open, close := UseToggle(ctx, false)
    
    ctx.On("toggle", func(_ interface{}) {
        toggle()
    })
    
    ctx.Expose("isOpen", isOpen)
})
```

### Step 4: Improve with Struct Return

```go
type UseToggleReturn struct {
    Value    *bubbly.Ref[bool]
    Toggle   func()
    SetTrue  func()
    SetFalse func()
}

func UseToggle(ctx *bubbly.Context, initial bool) UseToggleReturn {
    state := ctx.Ref(initial)
    
    return UseToggleReturn{
        Value:    state,
        Toggle:   func() { state.Set(!state.GetTyped()) },
        SetTrue:  func() { state.Set(true) },
        SetFalse: func() { state.Set(false) },
    }
}
```

## Naming Conventions

### Use the "Use" Prefix

```go
// ✅ Good: Follows convention
func UseCounter(ctx *Context, initial int) UseCounterReturn
func UseAuth(ctx *Context) UseAuthReturn
func UseLocalStorage[T any](ctx *Context, key string) *Ref[T]

// ❌ Avoid: Non-standard names
func Counter(ctx *Context, initial int) UseCounterReturn
func GetAuth(ctx *Context) UseAuthReturn
func Storage[T any](ctx *Context, key string) *Ref[T]
```

### Use Clear, Descriptive Names

```go
// ✅ Good: Clear purpose
func UsePagination(ctx *Context, itemsPerPage int)
func UseWebSocket(ctx *Context, url string)
func UseFormValidation[T any](ctx *Context, rules ValidationRules)

// ❌ Avoid: Vague names
func UsePage(ctx *Context, items int)
func UseWS(ctx *Context, url string)
func UseValidation[T any](ctx *Context, rules ValidationRules)
```

## Return Value Patterns

### Pattern 1: Named Struct (Recommended)

```go
type UseCounterReturn struct {
    Count     *Ref[int]
    Increment func()
    Decrement func()
    Reset     func()
}

func UseCounter(ctx *Context, initial int) UseCounterReturn {
    count := ctx.Ref(initial)
    
    return UseCounterReturn{
        Count:     count,
        Increment: func() { count.Set(count.GetTyped() + 1) },
        Decrement: func() { count.Set(count.GetTyped() - 1) },
        Reset:     func() { count.Set(initial) },
    }
}

// Usage: Clear and self-documenting
counter := UseCounter(ctx, 0)
counter.Increment()
fmt.Println(counter.Count.GetTyped())
```

### Pattern 2: Multiple Returns (Simple Cases)

```go
func UseToggle(ctx *Context, initial bool) (*Ref[bool], func()) {
    state := ctx.Ref(initial)
    toggle := func() { state.Set(!state.GetTyped()) }
    return state, toggle
}

// Usage: Simple destructuring
isOpen, toggle := UseToggle(ctx, false)
```

### Pattern 3: Single Return (Minimal)

```go
func UseDebounce[T comparable](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T] {
    debounced := ctx.Ref(value.GetTyped())
    // ... debounce logic
    return debounced
}

// Usage: Direct assignment
debounced := UseDebounce(ctx, searchTerm, 300*time.Millisecond)
```

**Recommendation:** Use named structs for composables with 3+ return values.

## Type Safety with Generics

### Basic Generic Composable

```go
func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    value := ctx.Ref(initial)
    
    return UseStateReturn[T]{
        Value: value,
        Set:   func(v T) { value.Set(v) },
        Get:   func() T { return value.GetTyped() },
    }
}

// Type-safe usage
counter := UseState(ctx, 0)       // UseStateReturn[int]
name := UseState(ctx, "Alice")    // UseStateReturn[string]
user := UseState(ctx, &User{})    // UseStateReturn[*User]
```

### Constrained Generics

```go
func UseNumericState[T constraints.Integer | constraints.Float](
    ctx *Context,
    initial T,
) UseNumericStateReturn[T] {
    value := ctx.Ref(initial)
    
    increment := func(amount T) {
        value.Set(value.GetTyped() + amount)
    }
    
    return UseNumericStateReturn[T]{
        Value:     value,
        Increment: increment,
    }
}

// Only numeric types allowed
intCounter := UseNumericState(ctx, 0)        // ✅ OK
floatCounter := UseNumericState(ctx, 0.0)    // ✅ OK
stringCounter := UseNumericState(ctx, "")    // ❌ Compile error
```

### Generic with Type Inference

```go
func UseList[T any](ctx *Context) UseListReturn[T] {
    items := ctx.Ref([]T{})
    
    add := func(item T) {
        current := items.GetTyped()
        items.Set(append(current, item))
    }
    
    remove := func(index int) {
        current := items.GetTyped()
        items.Set(append(current[:index], current[index+1:]...))
    }
    
    return UseListReturn[T]{
        Items:  items,
        Add:    add,
        Remove: remove,
    }
}

// Type inferred from usage
todos := UseList[Todo](ctx)
todos.Add(Todo{Title: "Learn BubblyUI"})
```

## Composable Composition

### Calling Composables from Composables

```go
func UseAuth(ctx *Context) UseAuthReturn {
    // Use standard composables
    user := composables.UseState(ctx, (*User)(nil))
    loading := composables.UseState(ctx, false)
    
    // Use custom composable
    token := UseTokenStorage(ctx)
    
    login := func(credentials Credentials) {
        loading.Set(true)
        // ... authentication
        loading.Set(false)
    }
    
    return UseAuthReturn{
        User:    user.Value,
        Loading: loading.Value,
        Token:   token,
        Login:   login,
    }
}
```

### Sharing State Between Composables

```go
// Shared state via parameter
func UseSharedCounter(ctx *Context, sharedCount *Ref[int]) UseCounterReturn {
    increment := func() {
        sharedCount.Set(sharedCount.GetTyped() + 1)
    }
    
    return UseCounterReturn{
        Count:     sharedCount,
        Increment: increment,
    }
}

// Usage
Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    counter1 := UseSharedCounter(ctx, count)
    counter2 := UseSharedCounter(ctx, count)
    
    // Both counters share same state!
})
```

### Composable Pipelines

```go
func UseFetchAndCache[T any](ctx *Context, fetchFn func() (*T, error)) UseDataReturn[T] {
    // Step 1: Fetch data
    fetcher := composables.UseAsync(ctx, fetchFn)
    
    // Step 2: Cache in storage
    storage := composables.NewFileStorage("/cache")
    cache := composables.UseLocalStorage(ctx, "cached-data", (*T)(nil), storage)
    
    // Step 3: Sync fetch result to cache
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        if fetcher.Data.GetTyped() != nil {
            cache.Set(fetcher.Data.GetTyped())
        }
        return nil
    }, fetcher.Data)
    
    return UseDataReturn[T]{
        Data:    fetcher.Data,
        Loading: fetcher.Loading,
        Cached:  cache.Value,
    }
}
```

## Lifecycle Integration

### OnMounted Hook

```go
func UseTimer(ctx *Context) UseTimerReturn {
    count := ctx.Ref(0)
    ticker := time.NewTicker(time.Second)
    
    ctx.OnMounted(func() {
        go func() {
            for range ticker.C {
                count.Set(count.GetTyped() + 1)
            }
        }()
    })
    
    ctx.OnUnmounted(func() {
        ticker.Stop()
    })
    
    return UseTimerReturn{Count: count}
}
```

### OnUpdated Hook with Dependencies

```go
func UseAutoSave(ctx *Context, document *Ref[Document]) {
    ctx.OnUpdated(func() {
        saveDocument(document.GetTyped())
    }, document)
}
```

### Resource Cleanup

```go
func UseWebSocket(ctx *Context, url string) UseWebSocketReturn {
    messages := ctx.Ref([]string{})
    var conn *websocket.Conn
    
    ctx.OnMounted(func() {
        var err error
        conn, err = websocket.Dial(url, "", "http://localhost/")
        if err != nil {
            return
        }
        
        go func() {
            for {
                var msg string
                if err := websocket.Message.Receive(conn, &msg); err != nil {
                    break
                }
                current := messages.GetTyped()
                messages.Set(append(current, msg))
            }
        }()
    })
    
    ctx.OnUnmounted(func() {
        if conn != nil {
            conn.Close()
        }
    })
    
    return UseWebSocketReturn{Messages: messages}
}
```

## Advanced Patterns

### Pattern 1: Pagination

```go
type UsePaginationReturn struct {
    CurrentPage *Ref[int]
    TotalPages  *Computed[int]
    HasNext     *Computed[bool]
    HasPrev     *Computed[bool]
    NextPage    func()
    PrevPage    func()
    GoToPage    func(int)
}

func UsePagination(ctx *Context, itemsPerPage int) UsePaginationReturn {
    currentPage := ctx.Ref(1)
    totalItems := ctx.Ref(0)
    
    totalPages := ctx.Computed(func() int {
        items := totalItems.GetTyped()
        if items == 0 {
            return 1
        }
        return (items + itemsPerPage - 1) / itemsPerPage
    })
    
    hasNext := ctx.Computed(func() bool {
        return currentPage.GetTyped() < totalPages.GetTyped()
    })
    
    hasPrev := ctx.Computed(func() bool {
        return currentPage.GetTyped() > 1
    })
    
    nextPage := func() {
        if hasNext.GetTyped() {
            currentPage.Set(currentPage.GetTyped() + 1)
        }
    }
    
    prevPage := func() {
        if hasPrev.GetTyped() {
            currentPage.Set(currentPage.GetTyped() - 1)
        }
    }
    
    goToPage := func(page int) {
        if page >= 1 && page <= totalPages.GetTyped() {
            currentPage.Set(page)
        }
    }
    
    return UsePaginationReturn{
        CurrentPage: currentPage,
        TotalPages:  totalPages,
        HasNext:     hasNext,
        HasPrev:     hasPrev,
        NextPage:    nextPage,
        PrevPage:    prevPage,
        GoToPage:    goToPage,
    }
}
```

### Pattern 2: Undo/Redo

```go
type UseHistoryReturn[T any] struct {
    State   *Ref[T]
    CanUndo *Computed[bool]
    CanRedo *Computed[bool]
    Undo    func()
    Redo    func()
    Clear   func()
}

func UseHistory[T any](ctx *Context, initial T, limit int) UseHistoryReturn[T] {
    state := ctx.Ref(initial)
    history := ctx.Ref([]T{initial})
    currentIndex := ctx.Ref(0)
    
    canUndo := ctx.Computed(func() bool {
        return currentIndex.GetTyped() > 0
    })
    
    canRedo := ctx.Computed(func() bool {
        return currentIndex.GetTyped() < len(history.GetTyped())-1
    })
    
    // Record state changes
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        current := state.GetTyped()
        h := history.GetTyped()
        idx := currentIndex.GetTyped()
        
        // Truncate future history
        h = h[:idx+1]
        
        // Add new state
        h = append(h, current)
        
        // Limit history size
        if len(h) > limit {
            h = h[1:]
        } else {
            idx++
        }
        
        history.Set(h)
        currentIndex.Set(idx)
        
        return nil
    }, state)
    
    undo := func() {
        if canUndo.GetTyped() {
            idx := currentIndex.GetTyped() - 1
            currentIndex.Set(idx)
            state.Set(history.GetTyped()[idx])
        }
    }
    
    redo := func() {
        if canRedo.GetTyped() {
            idx := currentIndex.GetTyped() + 1
            currentIndex.Set(idx)
            state.Set(history.GetTyped()[idx])
        }
    }
    
    clear := func() {
        history.Set([]T{state.GetTyped()})
        currentIndex.Set(0)
    }
    
    return UseHistoryReturn[T]{
        State:   state,
        CanUndo: canUndo,
        CanRedo: canRedo,
        Undo:    undo,
        Redo:    redo,
        Clear:   clear,
    }
}
```

### Pattern 3: State Machine

```go
type State string

const (
    StateIdle    State = "idle"
    StateLoading State = "loading"
    StateSuccess State = "success"
    StateError   State = "error"
)

type UseStateMachineReturn struct {
    Current *Ref[State]
    Is      func(State) bool
    Can     func(State) bool
    To      func(State)
}

func UseStateMachine(ctx *Context, initial State, transitions map[State][]State) UseStateMachineReturn {
    current := ctx.Ref(initial)
    
    is := func(state State) bool {
        return current.GetTyped() == state
    }
    
    can := func(target State) bool {
        allowed, exists := transitions[current.GetTyped()]
        if !exists {
            return false
        }
        
        for _, s := range allowed {
            if s == target {
                return true
            }
        }
        return false
    }
    
    to := func(target State) {
        if can(target) {
            current.Set(target)
        }
    }
    
    return UseStateMachineReturn{
        Current: current,
        Is:      is,
        Can:     can,
        To:      to,
    }
}

// Usage
Setup(func(ctx *Context) {
    machine := UseStateMachine(ctx, StateIdle, map[State][]State{
        StateIdle:    {StateLoading},
        StateLoading: {StateSuccess, StateError},
        StateSuccess: {StateIdle},
        StateError:   {StateIdle},
    })
    
    machine.To(StateLoading) // Valid
    machine.To(StateIdle)    // Invalid, won't transition
})
```

## Testing Composables

### Unit Testing

```go
func TestUseCounter(t *testing.T) {
    ctx := bubbly.NewTestContext()
    counter := UseCounter(ctx, 0)
    
    // Test initial state
    assert.Equal(t, 0, counter.Count.GetTyped())
    
    // Test increment
    counter.Increment()
    assert.Equal(t, 1, counter.Count.GetTyped())
    
    // Test decrement
    counter.Decrement()
    assert.Equal(t, 0, counter.Count.GetTyped())
    
    // Test reset
    counter.Increment()
    counter.Increment()
    counter.Reset()
    assert.Equal(t, 0, counter.Count.GetTyped())
}
```

### Testing with Lifecycle

```go
func TestUseTimer(t *testing.T) {
    ctx := bubbly.NewTestContext()
    timer := UseTimer(ctx)
    
    // Trigger mounted
    bubbly.TriggerMount(ctx)
    
    // Wait for timer tick
    time.Sleep(1100 * time.Millisecond)
    
    // Verify count increased
    assert.GreaterOrEqual(t, timer.Count.GetTyped(), 1)
    
    // Trigger unmount
    bubbly.TriggerUnmount(ctx)
    
    // Verify timer stopped
    count := timer.Count.GetTyped()
    time.Sleep(1100 * time.Millisecond)
    assert.Equal(t, count, timer.Count.GetTyped())
}
```

### Testing Async Composables

```go
func TestUseAsync(t *testing.T) {
    ctx := bubbly.NewTestContext()
    
    fetchCalled := false
    async := composables.UseAsync(ctx, func() (*string, error) {
        fetchCalled = true
        result := "data"
        return &result, nil
    })
    
    // Initially not loading
    assert.False(t, async.Loading.GetTyped())
    assert.Nil(t, async.Data.GetTyped())
    
    // Execute fetch
    async.Execute()
    
    // Verify fetch called
    assert.True(t, fetchCalled)
    
    // Wait for async completion
    time.Sleep(100 * time.Millisecond)
    
    // Verify data loaded
    assert.NotNil(t, async.Data.GetTyped())
    assert.Equal(t, "data", *async.Data.GetTyped())
    assert.False(t, async.Loading.GetTyped())
}
```

## Best Practices

### 1. Always Accept Context as First Parameter

```go
// ✅ Good
func UseCounter(ctx *Context, initial int) UseCounterReturn

// ❌ Wrong: Context not first
func UseCounter(initial int, ctx *Context) UseCounterReturn
```

### 2. Return Reactive Values, Not Plain Values

```go
// ✅ Good: Returns Ref
func UseCounter(ctx *Context) *Ref[int] {
    return ctx.Ref(0)
}

// ❌ Wrong: Returns plain value
func UseCounter(ctx *Context) int {
    return 0
}
```

### 3. Register Cleanup for Resources

```go
// ✅ Good: Cleanup registered
func UseWebSocket(ctx *Context, url string) *Ref[[]string] {
    messages := ctx.Ref([]string{})
    var conn *websocket.Conn
    
    ctx.OnMounted(func() {
        conn, _ = websocket.Dial(url, "", "")
    })
    
    ctx.OnUnmounted(func() {
        if conn != nil {
            conn.Close()
        }
    })
    
    return messages
}
```

### 4. Use Type Parameters for Reusability

```go
// ✅ Good: Generic, reusable
func UseList[T any](ctx *Context) UseListReturn[T]

// ❌ Limited: Only works with strings
func UseStringList(ctx *Context) UseStringListReturn
```

### 5. Document Your Composables

```go
// UseCounter provides a simple counter with increment/decrement.
//
// Parameters:
//   - ctx: Component context
//   - initial: Starting count value
//
// Returns:
//   - Count: Reactive counter
//   - Increment: Increases count by 1
//   - Decrement: Decreases count by 1
//   - Reset: Resets to initial value
//
// Example:
//   counter := UseCounter(ctx, 0)
//   counter.Increment()
func UseCounter(ctx *Context, initial int) UseCounterReturn {
    // ...
}
```

## Common Pitfalls

### Pitfall 1: Not Cleaning Up Resources

```go
// ❌ Wrong: Timer never stopped
func UseInterval(ctx *Context) {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            // ...
        }
    }()
    // ticker.Stop() never called!
}
```

### Pitfall 2: Returning Plain Values

```go
// ❌ Wrong: Loses reactivity
func UseCount(ctx *Context) int {
    count := ctx.Ref(0)
    return count.GetTyped() // Not reactive!
}
```

### Pitfall 3: Global State

```go
// ❌ Wrong: Global state leaks between instances
var globalCount int

func UseCounter(ctx *Context) int {
    globalCount++
    return globalCount
}
```

### Pitfall 4: Missing Dependencies in UseEffect

```go
// ❌ Wrong: Missing dependency
func UseAutoSave(ctx *Context) {
    document := ctx.Ref(Document{})
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        save(document.GetTyped())
        return nil
    }) // Missing document as dependency!
}
```

## Real-World Examples

See **[Composition API Guide](./composition-api.md)** for complete real-world examples including:
- Search with debouncing
- Form validation
- Authentication
- Pagination
- Data fetching with caching

## Next Steps

- **[Standard Composables](./standard-composables.md)** - Learn built-in composables
- **[Composition API Guide](./composition-api.md)** - Core concepts
- **[Testing Guide](../testing/testing-guide.md)** - Test your composables
