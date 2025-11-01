# Composition API Guide

## Overview

The Composition API is a Vue 3-inspired system for organizing component logic into reusable, composable functions. It enables better code organization, logic reuse across components, and improved type safety through Go generics.

## Table of Contents

- [What is Composition API?](#what-is-composition-api)
- [Why Composition API?](#why-composition-api)
- [Core Concepts](#core-concepts)
- [Getting Started](#getting-started)
- [Complete Examples](#complete-examples)
- [Provide/Inject Pattern](#provideinject-pattern)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Troubleshooting](#troubleshooting)
- [Performance](#performance)
- [Next Steps](#next-steps)

## What is Composition API?

The Composition API is a set of patterns and functions that allow you to:

1. **Extract reusable logic** into composable functions
2. **Organize related code** together instead of scattering it across lifecycle hooks
3. **Share logic** between components without complex inheritance
4. **Achieve better type safety** through Go generics

### Composable Functions

A composable is a Go function that:
- Accepts a `*Context` as the first parameter
- Returns reactive values (`*Ref[T]`, `*Computed[T]`) and helper functions
- Integrates with the component lifecycle
- Follows the `Use*` naming convention

```go
func UseCounter(ctx *bubbly.Context, initial int) (*bubbly.Ref[int], func(), func()) {
    count := ctx.Ref(initial)
    
    increment := func() {
        count.Set(count.GetTyped() + 1)
    }
    
    decrement := func() {
        count.Set(count.GetTyped() - 1)
    }
    
    return count, increment, decrement
}
```

## Why Composition API?

### Problem: Scattered Logic

Traditional component structure scatters related logic:

```go
// ❌ Logic scattered across lifecycle hooks
Setup(func(ctx *Context) {
    // User data initialization
    userData := ctx.Ref[*User](nil)
    
    // Product data initialization  
    productData := ctx.Ref[*Product](nil)
    
    ctx.OnMounted(func() {
        // User data fetching
        userData.Set(fetchUser())
        // Product data fetching
        productData.Set(fetchProduct())
    })
    
    ctx.OnUnmounted(func() {
        // User cleanup
        cleanupUser()
        // Product cleanup
        cleanupProduct()
    })
})
```

### Solution: Organized Logic

Composition API groups related logic:

```go
// ✅ Logic organized in composables
Setup(func(ctx *Context) {
    user := UseUser(ctx)         // All user logic together
    products := UseProducts(ctx) // All product logic together
    
    ctx.Expose("user", user.Data)
    ctx.Expose("products", products.Data)
})
```

### Benefits

1. **Better Organization**: Related code stays together
2. **Reusability**: Share logic across components
3. **Type Safety**: Compile-time type checking with generics
4. **Testability**: Test composables independently
5. **Composition**: Build complex logic from simple pieces

## Core Concepts

### 1. Context System

The `Context` provides access to component features:

```go
func UseExample(ctx *bubbly.Context) {
    // Create reactive state
    state := ctx.Ref(0)
    computed := ctx.Computed(func() int {
        return state.GetTyped() * 2
    })
    
    // Register lifecycle hooks
    ctx.OnMounted(func() {
        fmt.Println("Mounted")
    })
    
    // Provide values to children
    ctx.Provide("key", "value")
    
    // Inject from parents
    value := ctx.Inject("key", "default")
}
```

### 2. Reactive Values

Composables return reactive values, not plain values:

```go
// ✅ Good: Returns reactive Ref
func UseState[T any](ctx *Context, initial T) *Ref[T] {
    return ctx.Ref(initial)
}

// ❌ Bad: Returns plain value (not reactive)
func UseState[T any](ctx *Context, initial T) T {
    return initial
}
```

### 3. Lifecycle Integration

Composables can register lifecycle hooks:

```go
func UseTimer(ctx *bubbly.Context) *bubbly.Ref[int] {
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
    
    return count
}
```

### 4. Composable Composition

Composables can call other composables:

```go
func UseAuth(ctx *bubbly.Context) UseAuthReturn {
    // Use standard composables
    user := composables.UseState(ctx, (*User)(nil))
    loading := composables.UseState(ctx, false)
    
    login := func(credentials Credentials) {
        loading.Set(true)
        // ... authentication logic
        loading.Set(false)
    }
    
    return UseAuthReturn{
        User:    user.Value,
        Loading: loading.Value,
        Login:   login,
    }
}
```

## Getting Started

### Step 1: Create Your First Composable

Let's create a simple counter composable:

```go
package myapp

import "github.com/newbpydev/bubblyui/pkg/bubbly"

type UseCounterReturn struct {
    Count     *bubbly.Ref[int]
    Increment func()
    Decrement func()
    Reset     func()
}

func UseCounter(ctx *bubbly.Context, initial int) UseCounterReturn {
    count := ctx.Ref(initial)
    
    increment := func() {
        count.Set(count.GetTyped() + 1)
    }
    
    decrement := func() {
        count.Set(count.GetTyped() - 1)
    }
    
    reset := func() {
        count.Set(initial)
    }
    
    return UseCounterReturn{
        Count:     count,
        Increment: increment,
        Decrement: decrement,
        Reset:     reset,
    }
}
```

### Step 2: Use in a Component

```go
component := bubbly.NewComponent("Counter").
    Setup(func(ctx *bubbly.Context) {
        counter := UseCounter(ctx, 0)
        
        ctx.On("increment", func(_ interface{}) {
            counter.Increment()
        })
        
        ctx.On("decrement", func(_ interface{}) {
            counter.Decrement()
        })
        
        ctx.Expose("count", counter.Count)
    }).
    Template(func(ctx bubbly.RenderContext) string {
        count := ctx.Get("count").(*bubbly.Ref[int])
        return fmt.Sprintf("Count: %d", count.GetTyped())
    }).
    Build()
```

### Step 3: Reuse in Another Component

```go
component := bubbly.NewComponent("ScoreBoard").
    Setup(func(ctx *bubbly.Context) {
        player1 := UseCounter(ctx, 0)
        player2 := UseCounter(ctx, 0)
        
        ctx.Expose("p1Score", player1.Count)
        ctx.Expose("p2Score", player2.Count)
    }).
    Build()
```

## Complete Examples

### Example 1: Search with Debouncing

```go
func UseSearch(ctx *bubbly.Context) UseSearchReturn {
    searchTerm := ctx.Ref("")
    results := ctx.Ref([]string{})
    loading := ctx.Ref(false)
    
    // Debounce search term
    debouncedTerm := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)
    
    // Execute search when debounced term changes
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        term := debouncedTerm.GetTyped()
        if term == "" {
            results.Set([]string{})
            return nil
        }
        
        loading.Set(true)
        go func() {
            searchResults := performSearch(term)
            results.Set(searchResults)
            loading.Set(false)
        }()
        
        return nil
    }, debouncedTerm)
    
    return UseSearchReturn{
        SearchTerm: searchTerm,
        Results:    results,
        Loading:    loading,
    }
}
```

### Example 2: Async Data Fetching

```go
func UseUser(ctx *bubbly.Context, userID string) UseUserReturn {
    userData := composables.UseAsync(ctx, func() (*User, error) {
        return api.FetchUser(userID)
    })
    
    ctx.OnMounted(func() {
        userData.Execute()
    })
    
    return UseUserReturn{
        Data:    userData.Data,
        Loading: userData.Loading,
        Error:   userData.Error,
        Refetch: userData.Execute,
    }
}
```

### Example 3: Form with Validation

```go
type LoginForm struct {
    Email    string
    Password string
}

func UseLoginForm(ctx *bubbly.Context) UseLoginFormReturn {
    form := composables.UseForm(ctx, LoginForm{}, func(f LoginForm) map[string]string {
        errors := make(map[string]string)
        
        if !strings.Contains(f.Email, "@") {
            errors["Email"] = "Invalid email address"
        }
        
        if len(f.Password) < 8 {
            errors["Password"] = "Password must be at least 8 characters"
        }
        
        return errors
    })
    
    submit := func() {
        form.Submit()
        if form.IsValid.GetTyped() {
            loginUser(form.Values.GetTyped())
        }
    }
    
    return UseLoginFormReturn{
        Form:   form,
        Submit: submit,
    }
}
```

### Example 4: Pagination

```go
func UsePagination(ctx *bubbly.Context, itemsPerPage int) UsePaginationReturn {
    currentPage := composables.UseState(ctx, 1)
    totalItems := composables.UseState(ctx, 0)
    
    totalPages := ctx.Computed(func() int {
        items := totalItems.Get()
        if items == 0 {
            return 1
        }
        return (items + itemsPerPage - 1) / itemsPerPage
    })
    
    nextPage := func() {
        if currentPage.Get() < totalPages.GetTyped() {
            currentPage.Set(currentPage.Get() + 1)
        }
    }
    
    prevPage := func() {
        if currentPage.Get() > 1 {
            currentPage.Set(currentPage.Get() - 1)
        }
    }
    
    goToPage := func(page int) {
        if page >= 1 && page <= totalPages.GetTyped() {
            currentPage.Set(page)
        }
    }
    
    return UsePaginationReturn{
        CurrentPage: currentPage.Value,
        TotalPages:  totalPages,
        TotalItems:  totalItems.Value,
        NextPage:    nextPage,
        PrevPage:    prevPage,
        GoToPage:    goToPage,
    }
}
```

### Example 5: Toggle State

```go
func UseToggle(ctx *bubbly.Context, initial bool) (*bubbly.Ref[bool], func(), func(), func()) {
    state := composables.UseState(ctx, initial)
    
    toggle := func() {
        state.Set(!state.Get())
    }
    
    setTrue := func() {
        state.Set(true)
    }
    
    setFalse := func() {
        state.Set(false)
    }
    
    return state.Value, toggle, setTrue, setFalse
}
```

## Provide/Inject Pattern

Provide/Inject enables passing data through the component tree without prop drilling.

### Basic Usage

```go
// Parent component provides theme
parentSetup := func(ctx *bubbly.Context) {
    theme := ctx.Ref("dark")
    ctx.Provide("theme", theme)
    ctx.Expose("theme", theme)
}

// Child component injects theme
childSetup := func(ctx *bubbly.Context) {
    theme := ctx.Inject("theme", ctx.Ref("light"))
    ctx.Expose("theme", theme)
}
```

### Type-Safe Provide/Inject

```go
// Define typed keys
var ThemeKey = bubbly.NewProvideKey[*bubbly.Ref[string]]("theme")
var UserKey = bubbly.NewProvideKey[*bubbly.Ref[*User]]("currentUser")

// Parent provides
parentSetup := func(ctx *bubbly.Context) {
    theme := ctx.Ref("dark")
    user := ctx.Ref(&User{Name: "Alice"})
    
    bubbly.ProvideTyped(ctx, ThemeKey, theme)
    bubbly.ProvideTyped(ctx, UserKey, user)
}

// Child injects with type safety
childSetup := func(ctx *bubbly.Context) {
    theme := bubbly.InjectTyped(ctx, ThemeKey, ctx.Ref("light"))
    user := bubbly.InjectTyped(ctx, UserKey, ctx.Ref[*User](nil))
    
    // theme and user are properly typed!
}
```

### Composable with Provide/Inject

```go
func UseTheme(ctx *bubbly.Context) UseThemeReturn {
    theme := bubbly.InjectTyped(ctx, ThemeKey, ctx.Ref("light"))
    
    isDark := ctx.Computed(func() bool {
        return theme.GetTyped() == "dark"
    })
    
    toggleTheme := func() {
        if theme.GetTyped() == "dark" {
            theme.Set("light")
        } else {
            theme.Set("dark")
        }
    }
    
    return UseThemeReturn{
        Theme:       theme,
        IsDark:      isDark,
        ToggleTheme: toggleTheme,
    }
}
```

## Best Practices

### 1. Return Named Structs

```go
// ✅ Good: Clear, self-documenting
type UseCounterReturn struct {
    Count     *bubbly.Ref[int]
    Increment func()
    Decrement func()
}

func UseCounter(ctx *Context, initial int) UseCounterReturn {
    // ...
}

// ❌ Avoid: Unclear what each return value means
func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
    // ...
}
```

### 2. Use Type Parameters

```go
// ✅ Good: Type-safe generic composable
func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    value := ctx.Ref(initial)
    return UseStateReturn[T]{Value: value}
}

// ❌ Avoid: Loses type information
func UseState(ctx *Context, initial interface{}) UseStateReturn {
    // ...
}
```

### 3. Register Cleanup

```go
// ✅ Good: Cleanup registered
func UseInterval(ctx *Context, interval time.Duration, callback func()) {
    ticker := time.NewTicker(interval)
    
    ctx.OnMounted(func() {
        go func() {
            for range ticker.C {
                callback()
            }
        }()
    })
    
    ctx.OnUnmounted(func() {
        ticker.Stop()
    })
}

// ❌ Avoid: Resource leak
func UseInterval(ctx *Context, interval time.Duration, callback func()) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            callback()
        }
    }()
    // ticker never stopped!
}
```

### 4. Avoid Global State

```go
// ❌ Bad: Global state leaks between instances
var globalCount int

func UseCounter(ctx *Context) int {
    return globalCount
}

// ✅ Good: Context-based state is isolated
func UseCounter(ctx *Context) *Ref[int] {
    return ctx.Ref(0)
}
```

### 5. Document Composables

```go
// UseCounter provides a simple counter with increment/decrement operations.
//
// Parameters:
//   - ctx: Component context for reactivity and lifecycle
//   - initial: Starting count value
//
// Returns:
//   - Count: Reactive counter value
//   - Increment: Function to increase count by 1
//   - Decrement: Function to decrease count by 1
//
// Example:
//   counter := UseCounter(ctx, 0)
//   counter.Increment()
//   fmt.Println(counter.Count.GetTyped()) // Prints: 1
func UseCounter(ctx *bubbly.Context, initial int) UseCounterReturn {
    // ...
}
```

## Common Patterns

### Pattern 1: Composable Chains

```go
// High-level composable uses lower-level ones
func UseAuth(ctx *Context) UseAuthReturn {
    user := composables.UseState(ctx, (*User)(nil))
    token := composables.UseLocalStorage(ctx, "auth-token", "", storage)
    loading := composables.UseState(ctx, false)
    
    // Combine multiple composables
    return UseAuthReturn{
        User:    user.Value,
        Token:   token.Value,
        Loading: loading.Value,
    }
}
```

### Pattern 2: Conditional Composables

```go
func UseOptionalFeature(ctx *Context, enabled bool) *Ref[string] {
    data := ctx.Ref("")
    
    if enabled {
        // Only use composable when enabled
        fetcher := composables.UseAsync(ctx, fetchData)
        ctx.OnMounted(func() {
            fetcher.Execute()
        })
    }
    
    return data
}
```

### Pattern 3: Shared State

```go
// Composables share state via closure
func UseSharedCounter(ctx *Context) (UseCounterReturn, UseCounterReturn) {
    // Shared count
    count := ctx.Ref(0)
    
    // Two counters sharing same state
    counter1 := UseCounterReturn{
        Count:     count,
        Increment: func() { count.Set(count.GetTyped() + 1) },
    }
    
    counter2 := UseCounterReturn{
        Count:     count,
        Increment: func() { count.Set(count.GetTyped() + 1) },
    }
    
    return counter1, counter2
}
```

## Troubleshooting

### Issue 1: Composable Not Updating

**Problem:** Changes to composable state don't trigger updates

**Cause:** Returning plain values instead of reactive Refs

**Solution:**
```go
// ❌ Wrong: Returns plain value
func UseCount(ctx *Context) int {
    count := ctx.Ref(0)
    return count.GetTyped() // Loses reactivity!
}

// ✅ Correct: Returns reactive Ref
func UseCount(ctx *Context) *Ref[int] {
    return ctx.Ref(0)
}
```

### Issue 2: Memory Leaks

**Problem:** Resources not cleaned up, memory usage grows

**Cause:** Missing cleanup in OnUnmounted

**Solution:**
```go
// ✅ Proper cleanup
func UseWebSocket(ctx *Context, url string) *Ref[string] {
    data := ctx.Ref("")
    var conn *websocket.Conn
    
    ctx.OnMounted(func() {
        conn, _ = websocket.Dial(url, "", "http://localhost/")
    })
    
    ctx.OnUnmounted(func() {
        if conn != nil {
            conn.Close() // Critical: cleanup connection
        }
    })
    
    return data
}
```

### Issue 3: Inject Returns Default

**Problem:** Inject always returns default value

**Cause:** Provide/inject key mismatch or wrong component tree

**Solution:**
```go
// Ensure keys match exactly
const ThemeKey = "app-theme"

// Parent
ctx.Provide(ThemeKey, "dark") // Exact key

// Child (must be descendant of parent!)
theme := ctx.Inject(ThemeKey, "light") // Same key
```

### Issue 4: Infinite Loop in UseEffect

**Problem:** UseEffect runs infinitely

**Cause:** Missing or incorrect dependencies

**Solution:**
```go
// ❌ Wrong: No dependencies, runs on every update
composables.UseEffect(ctx, func() UseEffectCleanup {
    fmt.Println("Runs infinitely!")
    return nil
})

// ✅ Correct: Specify dependencies
count := ctx.Ref(0)
composables.UseEffect(ctx, func() UseEffectCleanup {
    fmt.Println("Runs only when count changes")
    return nil
}, count)
```

### Issue 5: Race Conditions in UseAsync

**Problem:** Multiple concurrent async calls cause data corruption

**Cause:** No cancellation of previous requests

**Solution:**
```go
func UseAsyncWithCancel(ctx *Context) UseAsyncReturn {
    data := ctx.Ref[*Result](nil)
    loading := ctx.Ref(false)
    currentRequest := ctx.Ref(0)
    
    execute := func() {
        requestID := currentRequest.GetTyped() + 1
        currentRequest.Set(requestID)
        
        loading.Set(true)
        go func() {
            result := fetchData()
            
            // Only update if this is still the latest request
            if currentRequest.GetTyped() == requestID {
                data.Set(result)
                loading.Set(false)
            }
        }()
    }
    
    return UseAsyncReturn{Data: data, Loading: loading, Execute: execute}
}
```

## Performance

Composables are designed for minimal overhead:

- **Call overhead:** < 100ns per composable call
- **State access:** < 10ns through reactive Refs
- **Provide/Inject:** < 12-122ns depending on tree depth (with caching)
- **Memory:** Minimal allocations, efficient cleanup

### Optimization Tips

1. **Memoize expensive computations** with `Computed`
2. **Use debounce/throttle** for high-frequency updates
3. **Lazy initialization** for expensive resources
4. **Proper cleanup** prevents memory leaks

## Next Steps

- **[Standard Composables Guide](./standard-composables.md)** - Learn about built-in composables
- **[Custom Composables Guide](./custom-composables.md)** - Create your own composables
- **[Lifecycle Hooks Guide](./lifecycle-hooks.md)** - Deep dive into lifecycle integration
- **[Reactivity Guide](../README.md)** - Understand the reactive system

## Further Reading

- Vue 3 Composition API: https://vuejs.org/guide/extras/composition-api-faq.html
- React Hooks: https://react.dev/reference/react/hooks
- BubblyUI Package Documentation: See `pkg/bubbly/composables/doc.go`
