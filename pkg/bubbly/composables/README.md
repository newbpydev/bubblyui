# BubblyUI Composables

**Reusable composition functions for Vue-inspired TUI components**

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Standard Composables](#standard-composables)
  - [UseState](#usestate)
  - [UseEffect](#useeffect)
  - [UseAsync](#useasync)
  - [UseDebounce](#usedebounce)
  - [UseThrottle](#usethrottle)
  - [UseForm](#useform)
  - [UseLocalStorage](#uselocalstorage)
  - [UseEventListener](#useeventlistener)
- [Common Patterns](#common-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)

---

## Introduction

Composables are reusable functions that encapsulate component logic in BubblyUI. Inspired by Vue 3's Composition API and React hooks, composables enable you to:

- **Extract logic** from components into reusable functions
- **Share state** and behavior across components
- **Organize code** by feature instead of lifecycle
- **Compose functionality** by combining multiple composables
- **Test independently** without mounting full components

All composables are **type-safe** (using Go generics), **thread-safe**, and integrate seamlessly with BubblyUI's reactivity system.

---

## Installation

Composables are part of the BubblyUI package:

```bash
go get github.com/newbpydev/bubblyui
```

Import in your code:

```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)
```

---

## Quick Start

Use composables in a component's `Setup` function:

```go
package main

import (
    "fmt"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

func main() {
    component, _ := bubbly.NewComponent("Counter").
        Setup(func(ctx *bubbly.Context) {
            // Use a composable
            counter := composables.UseState(ctx, 0)

            // Event handler
            ctx.On("increment", func(_ interface{}) {
                counter.Set(counter.Get() + 1)
            })

            // Expose to template
            ctx.Expose("count", counter.Value)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[int])
            return fmt.Sprintf("Count: %d", count.GetTyped())
        }).
        Build()

    component.Init()
    component.Emit("increment", nil)
    fmt.Println(component.View()) // "Count: 1"
}
```

---

## Standard Composables

### UseState

**Simple reactive state management with getter/setter API.**

#### Signature

```go
func UseState[T any](ctx *Context, initial T) UseStateReturn[T]

type UseStateReturn[T any] struct {
    Value *Ref[T]  // Reactive reference
    Set   func(T)  // Update value
    Get   func() T // Read value
}
```

#### Use Cases

- Managing simple component state
- Form field values
- Toggle states
- Counters and numeric values

#### Example

```go
Setup(func(ctx *bubbly.Context) {
    // Create state
    name := composables.UseState(ctx, "")

    // Update state
    ctx.On("nameChange", func(data interface{}) {
        name.Set(data.(string))
    })

    // Read state
    ctx.On("submit", func(_ interface{}) {
        fmt.Printf("Name: %s\n", name.Get())
    })

    // Expose to template
    ctx.Expose("name", name.Value)
})
```

---

### UseEffect

**Side effect management with dependency tracking and cleanup.**

#### Signature

```go
func UseEffect(ctx *Context, effect func() UseEffectCleanup, deps ...Dependency)

type UseEffectCleanup func()
```

#### Use Cases

- Data fetching on mount
- Setting up subscriptions
- Logging state changes
- External API integration
- Resource cleanup

#### Example

```go
Setup(func(ctx *bubbly.Context) {
    count := ctx.Ref(0)

    // Effect runs on mount and when count changes
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        fmt.Printf("Count is now: %d\n", count.GetTyped())

        // Optional cleanup
        return func() {
            fmt.Println("Cleaning up effect")
        }
    }, count)

    ctx.Expose("count", count)
})
```

#### Dependencies

- **No deps**: Effect runs on mount and every update
- **With deps**: Effect runs on mount and when dependencies change
- **Empty slice**: Runs only on mount (use explicit empty slice: `[]Dependency{}`)

---

### UseAsync

**Async data fetching with loading, error, and data states.**

#### Signature

```go
func UseAsync[T any](ctx *Context, fetcher func() (*T, error)) UseAsyncReturn[T]

type UseAsyncReturn[T any] struct {
    Data    *Ref[*T]     // Result data
    Loading *Ref[bool]   // Loading state
    Error   *Ref[error]  // Error state
    Execute func()       // Trigger fetch
    Reset   func()       // Reset all state
}
```

#### Use Cases

- API calls
- Database queries
- File loading
- Any async operation

#### Example

```go
type User struct {
    Name  string
    Email string
}

Setup(func(ctx *bubbly.Context) {
    // Create async handler
    userData := composables.UseAsync(ctx, func() (*User, error) {
        // Simulate API call
        time.Sleep(100 * time.Millisecond)
        return &User{Name: "Alice", Email: "alice@example.com"}, nil
    })

    // Execute on mount
    ctx.OnMounted(func() {
        userData.Execute()
    })

    // Expose states
    ctx.Expose("user", userData.Data)
    ctx.Expose("loading", userData.Loading)
    ctx.Expose("error", userData.Error)
})
```

#### State Management

```go
// Before Execute()
Data:    nil
Loading: false
Error:   nil

// During Execute()
Data:    nil
Loading: true
Error:   nil

// After Success
Data:    &User{...}
Loading: false
Error:   nil

// After Failure
Data:    nil
Loading: false
Error:   &SomeError{...}
```

---

### UseDebounce

**Debounced reactive values - updates only after a quiet period.**

#### Signature

```go
func UseDebounce[T any](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T]
```

#### Use Cases

- Search input (wait for user to stop typing)
- Window resize handling
- Form validation
- API rate limiting

#### Example

```go
Setup(func(ctx *bubbly.Context) {
    searchTerm := ctx.Ref("")
    debouncedSearch := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)

    // Watch debounced value for API calls
    ctx.Watch(debouncedSearch, func(newVal, oldVal string) {
        if newVal != "" {
            performSearch(newVal)
        }
    })

    ctx.On("searchInput", func(data interface{}) {
        // Updates immediately (not debounced)
        searchTerm.Set(data.(string))
    })

    ctx.Expose("searchTerm", searchTerm)
    ctx.Expose("debouncedSearch", debouncedSearch)
})
```

#### Behavior

```
User types: "h" -> "he" -> "hel" -> "hell" -> "hello"
                                                  ↓
                                        (300ms quiet period)
                                                  ↓
API called with: "hello"
```

---

### UseThrottle

**Throttled function execution - limits call rate.**

#### Signature

```go
func UseThrottle(ctx *Context, fn func(), delay time.Duration) func()
```

#### Use Cases

- Scroll event handling
- Button click protection
- Mouse tracking
- API rate limiting

#### Example

```go
Setup(func(ctx *bubbly.Context) {
    scrollCount := ctx.Ref(0)

    handleScroll := func() {
        scrollCount.Set(scrollCount.GetTyped() + 1)
        updateScrollPosition()
    }

    // Throttle to max 1 call per 100ms
    throttledScroll := composables.UseThrottle(ctx, handleScroll, 100*time.Millisecond)

    ctx.On("scroll", func(_ interface{}) {
        throttledScroll() // Executes at most once per 100ms
    })

    ctx.Expose("scrollCount", scrollCount)
})
```

#### Throttle vs Debounce

| Feature | Throttle | Debounce |
|---------|----------|----------|
| **First call** | Executes immediately | Waits for quiet period |
| **Subsequent calls** | Limited to rate | Ignored until quiet |
| **Use case** | Continuous events (scroll) | Sporadic events (typing) |
| **Example** | `_X__X__X__X` | `________X` |

---

### UseForm

**Form state management with validation and field tracking.**

#### Signature

```go
func UseForm[T any](
    ctx *Context,
    initial T,
    validate func(T) map[string]string,
) UseFormReturn[T]

type UseFormReturn[T any] struct {
    Values   *Ref[T]                    // Form values
    Errors   *Ref[map[string]string]    // Validation errors
    Touched  *Ref[map[string]bool]      // Touched fields
    IsValid  *Computed[bool]            // Form validity
    IsDirty  *Computed[bool]            // Has modifications
    Submit   func()                     // Validate and submit
    Reset    func()                     // Reset to initial
    SetField func(field string, value interface{}) // Update field
}
```

#### Use Cases

- Login forms
- Registration forms
- Settings panels
- Any user input with validation

#### Example

```go
type LoginForm struct {
    Email    string
    Password string
}

Setup(func(ctx *bubbly.Context) {
    form := composables.UseForm(ctx, LoginForm{}, func(f LoginForm) map[string]string {
        errors := make(map[string]string)

        // Validate email
        if f.Email == "" {
            errors["Email"] = "Email is required"
        } else if !strings.Contains(f.Email, "@") {
            errors["Email"] = "Invalid email format"
        }

        // Validate password
        if len(f.Password) < 8 {
            errors["Password"] = "Password must be at least 8 characters"
        }

        return errors
    })

    // Field change handlers
    ctx.On("emailChange", func(data interface{}) {
        form.SetField("Email", data.(string))
    })

    ctx.On("passwordChange", func(data interface{}) {
        form.SetField("Password", data.(string))
    })

    // Submit handler
    ctx.On("submit", func(_ interface{}) {
        form.Submit()

        if form.IsValid.GetTyped() {
            values := form.Values.GetTyped()
            submitToAPI(values)
        }
    })

    ctx.Expose("form", form)
})
```

#### Field Updates

```go
form.SetField("Email", "user@example.com")
// 1. Updates Values.Email
// 2. Marks Email as touched
// 3. Runs validation
// 4. Updates Errors map
// 5. Updates IsValid computed
```

---

### UseLocalStorage

**Persistent state with automatic JSON serialization to disk.**

#### Signature

```go
func UseLocalStorage[T any](ctx *Context, key string, initial T, storage Storage) UseStateReturn[T]

type Storage interface {
    Load(key string) ([]byte, error)
    Save(key string, data []byte) error
}
```

#### Use Cases

- Application settings
- User preferences
- Cache data
- Recent history
- Session persistence

#### Example

```go
type Settings struct {
    Theme    string
    FontSize int
    AutoSave bool
}

Setup(func(ctx *bubbly.Context) {
    // Create storage instance
    storage := composables.NewFileStorage(os.ExpandEnv("$HOME/.config/myapp"))

    // Load or initialize settings
    settings := composables.UseLocalStorage(ctx, "app-settings", Settings{
        Theme:    "dark",
        FontSize: 14,
        AutoSave: true,
    }, storage)

    // Changes automatically saved to disk
    ctx.On("changeTheme", func(data interface{}) {
        current := settings.Get()
        current.Theme = data.(string)
        settings.Set(current) // Saved to ~/.config/myapp/app-settings.json
    })

    ctx.Expose("settings", settings.Value)
})
```

#### Storage Implementations

**FileStorage** (included):

```go
storage := composables.NewFileStorage("/path/to/data")
```

**Custom Storage**:

```go
type RedisStorage struct { /* ... */ }

func (r *RedisStorage) Load(key string) ([]byte, error) {
    return r.client.Get(ctx, key).Bytes()
}

func (r *RedisStorage) Save(key string, data []byte) error {
    return r.client.Set(ctx, key, data, 0).Err()
}
```

---

### UseEventListener

**Event handling with automatic cleanup on unmount.**

#### Signature

```go
func UseEventListener(ctx *Context, event string, handler func()) func()
```

#### Use Cases

- Button clicks
- Keyboard shortcuts
- Mouse events
- Custom component events

#### Example

```go
Setup(func(ctx *bubbly.Context) {
    clickCount := ctx.Ref(0)

    // Register event listener
    cleanup := composables.UseEventListener(ctx, "click", func() {
        clickCount.Set(clickCount.GetTyped() + 1)
    })

    // Listener automatically cleaned up on unmount
    // Or manually: cleanup()

    ctx.Expose("clickCount", clickCount)
})
```

---

## Common Patterns

### Pattern 1: Authentication

```go
func UseAuth(ctx *bubbly.Context) UseAuthReturn {
    // Inject user from parent component
    userKey := bubbly.NewProvideKey[*bubbly.Ref[*User]]("currentUser")
    user := bubbly.InjectTyped(ctx, userKey, ctx.Ref[*User](nil))

    // Computed authentication state
    isAuthenticated := ctx.Computed(func() bool {
        return user.GetTyped() != nil
    })

    isAdmin := ctx.Computed(func() bool {
        u := user.GetTyped()
        return u != nil && u.Role == "admin"
    })

    login := func(email, password string) error {
        // Login logic
        loggedInUser, err := api.Login(email, password)
        if err != nil {
            return err
        }
        user.Set(loggedInUser)
        return nil
    }

    logout := func() {
        user.Set(nil)
    }

    return UseAuthReturn{
        User:            user,
        IsAuthenticated: isAuthenticated,
        IsAdmin:         isAdmin,
        Login:           login,
        Logout:          logout,
    }
}
```

### Pattern 2: Pagination

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

    canGoNext := ctx.Computed(func() bool {
        return currentPage.Get() < totalPages.GetTyped()
    })

    canGoPrev := ctx.Computed(func() bool {
        return currentPage.Get() > 1
    })

    nextPage := func() {
        if canGoNext.GetTyped() {
            currentPage.Set(currentPage.Get() + 1)
        }
    }

    prevPage := func() {
        if canGoPrev.GetTyped() {
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
        CanGoNext:   canGoNext,
        CanGoPrev:   canGoPrev,
        NextPage:    nextPage,
        PrevPage:    prevPage,
        GoToPage:    goToPage,
        SetTotal:    func(total int) { totalItems.Set(total) },
    }
}
```

### Pattern 3: Toggle State

```go
func UseToggle(ctx *bubbly.Context, initial bool) (*bubbly.Ref[bool], func()) {
    state := composables.UseState(ctx, initial)

    toggle := func() {
        state.Set(!state.Get())
    }

    return state.Value, toggle
}

// Usage
Setup(func(ctx *bubbly.Context) {
    darkMode, toggleDarkMode := UseToggle(ctx, true)

    ctx.On("toggleTheme", func(_ interface{}) {
        toggleDarkMode()
    })

    ctx.Expose("darkMode", darkMode)
})
```

---

## Best Practices

### 1. Return Named Structs

```go
// ❌ Avoid: Multiple return values are hard to track
func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func())

// ✅ Better: Named struct is self-documenting
type UseCounterReturn struct {
    Count     *Ref[int]
    Increment func()
    Decrement func()
}

func UseCounter(ctx *Context, initial int) UseCounterReturn
```

### 2. Use Type Parameters

```go
// ✅ Type-safe with generics
func UseState[T any](ctx *Context, initial T) UseStateReturn[T]

// Usage
counter := UseState(ctx, 0)      // UseStateReturn[int]
name := UseState(ctx, "")        // UseStateReturn[string]
user := UseState(ctx, User{})    // UseStateReturn[User]
```

### 3. Register Cleanup

```go
// ✅ Cleanup registered with lifecycle
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    ticker := time.NewTicker(1 * time.Second)

    go func() {
        for range ticker.C {
            doSomething()
        }
    }()

    return func() {
        ticker.Stop() // Cleanup on unmount
    }
})
```

### 4. Avoid Global State

```go
// ❌ Bad: Global state leaks between instances
var globalCount int

func UseBadCounter(ctx *Context) *Ref[int] {
    count := ctx.Ref(globalCount) // Shared state!
    return count
}

// ✅ Good: Context-based state is isolated
func UseGoodCounter(ctx *Context, initial int) *Ref[int] {
    count := ctx.Ref(initial) // Independent state
    return count
}
```

### 5. Document Contracts

```go
// UseDebounce returns a reactive ref that updates only after the specified
// delay has elapsed with no new changes to the source ref. Useful for
// rate-limiting expensive operations like API calls or search queries.
//
// Parameters:
//   - ctx: Component context for lifecycle management
//   - value: Source reactive reference to debounce
//   - delay: Duration to wait before propagating changes
//
// Returns:
//   A new reactive reference that updates after the delay period
//
// Example:
//   searchTerm := ctx.Ref("")
//   debounced := UseDebounce(ctx, searchTerm, 300*time.Millisecond)
func UseDebounce[T any](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T]
```

### 6. Compose Liberally

```go
// Build high-level composables from low-level ones
func UseAuth(ctx *Context) UseAuthReturn {
    user := UseAsync(ctx, api.FetchCurrentUser)
    token := UseLocalStorage(ctx, "auth-token", "", storage)

    isAuthenticated := ctx.Computed(func() bool {
        return user.Data.GetTyped() != nil && token.Get() != ""
    })

    return UseAuthReturn{
        User:            user,
        Token:           token,
        IsAuthenticated: isAuthenticated,
    }
}
```

---

## Troubleshooting

### Composable Not Working

**Problem**: Composable doesn't execute or returns unexpected results

**Solution**: Ensure composable is called inside `Setup` function, not in `Template` or `Update`.

```go
// ❌ Wrong: Called in template
Template(func(ctx RenderContext) string {
    state := composables.UseState(???, 0) // No context!
    return ""
})

// ✅ Correct: Called in Setup
Setup(func(ctx *Context) {
    state := composables.UseState(ctx, 0)
    ctx.Expose("state", state.Value)
})
```

### State Shared Between Instances

**Problem**: Component instances share state unexpectedly

**Solution**: Don't use global variables in composables. Use Context or closure state.

```go
// ❌ Wrong: Global state
var globalCount int

func UseBadCounter(ctx *Context) *Ref[int] {
    return ctx.Ref(globalCount) // Shared!
}

// ✅ Correct: Closure state
func UseGoodCounter(ctx *Context, initial int) *Ref[int] {
    return ctx.Ref(initial) // Independent!
}
```

### Inject Returns Default

**Problem**: `Inject` always returns default value

**Solution**: Ensure parent component provides the value before child mounts.

```go
// Parent
Setup(func(ctx *Context) {
    theme := ctx.Ref("dark")
    ctx.Provide("theme", theme) // Provide before children mount
})

// Child
Setup(func(ctx *Context) {
    theme := ctx.Inject("theme", ctx.Ref("light")) // Will get "dark"
})
```

### Type Assertion Panic

**Problem**: Panic on type assertion

**Solution**: Ensure injected type matches provided type exactly.

```go
// ❌ Wrong: Type mismatch
ctx.Provide("count", 42) // Provides int

count := ctx.Inject("count", ctx.Ref(0)).(*Ref[int]) // Expects Ref!
// Panic: interface conversion

// ✅ Correct: Matching types
ctx.Provide("count", ctx.Ref(42)) // Provides Ref[int]

count := ctx.Inject("count", ctx.Ref(0)).(*Ref[int]) // Works!
```

### Cleanup Not Running

**Problem**: Resources not cleaned up on unmount

**Solution**: Register cleanup with lifecycle hooks.

```go
// ✅ Cleanup registered
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    // Setup
    return func() {
        // This runs on unmount
    }
})
```

---

## API Reference

For detailed API documentation, see:

```bash
go doc github.com/newbpydev/bubblyui/pkg/bubbly/composables
```

Or view online at: [pkg.go.dev](https://pkg.go.dev/github.com/newbpydev/bubblyui/pkg/bubbly/composables)

### Complete Signatures

```go
// State management
func UseState[T any](ctx *Context, initial T) UseStateReturn[T]

// Side effects
func UseEffect(ctx *Context, effect func() UseEffectCleanup, deps ...Dependency)

// Async operations
func UseAsync[T any](ctx *Context, fetcher func() (*T, error)) UseAsyncReturn[T]

// Debouncing
func UseDebounce[T any](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T]

// Throttling
func UseThrottle(ctx *Context, fn func(), delay time.Duration) func()

// Form management
func UseForm[T any](ctx *Context, initial T, validate func(T) map[string]string) UseFormReturn[T]

// Persistent storage
func UseLocalStorage[T any](ctx *Context, key string, initial T, storage Storage) UseStateReturn[T]

// Event handling
func UseEventListener(ctx *Context, event string, handler func()) func()
```

---

## Contributing

Contributions are welcome! Please:

1. Follow Go idioms and existing code style
2. Add tests for new composables (>80% coverage)
3. Update documentation (godoc + README)
4. Run quality gates: `make test lint fmt build`

---

## License

See the LICENSE file in the repository root.
