# Standard Composables Reference

## Overview

This guide provides complete reference documentation for all built-in composables in BubblyUI. Each composable includes API details, use cases, and practical examples.

## Table of Contents

- [Quick Reference](#quick-reference)
- [UseState](#usestate)
- [UseEffect](#useeffect)
- [UseAsync](#useasync)
- [UseDebounce](#usedebounce)
- [UseThrottle](#usethrottle)
- [UseForm](#useform)
- [UseLocalStorage](#uselocalstorage)
- [UseEventListener](#useeventlistener)
- [Composable Comparison](#composable-comparison)

## Quick Reference

| Composable | Purpose | Performance | Common Use Cases |
|------------|---------|-------------|------------------|
| `UseState` | Simple state management | < 50ns | Counters, toggles, simple data |
| `UseEffect` | Side effect handling | < 1.2μs | Data syncing, logging, side effects |
| `UseAsync` | Async operations | < 250ns | API calls, data fetching |
| `UseDebounce` | Debounced values | < 850ns | Search input, autosave |
| `UseThrottle` | Throttled functions | < 500ns | Scroll handlers, resize events |
| `UseForm` | Form management | < 750ns | Login, signup, settings forms |
| `UseLocalStorage` | Persistent state | < 1.2μs | User preferences, app settings |
| `UseEventListener` | Event handling | minimal | Click handlers, keyboard events |

## UseState

**Purpose:** Simplified reactive state management with getter/setter API.

**Signature:**
```go
func UseState[T any](ctx *Context, initial T) UseStateReturn[T]

type UseStateReturn[T any] struct {
    Value *Ref[T]  // Reactive reference
    Set   func(T)  // Setter function
    Get   func() T // Getter function
}
```

**Performance:** ~46ns per call (4.3x better than 200ns target)

### Use Cases

1. **Simple counters**
2. **Toggle states** (boolean flags)
3. **Form inputs**
4. **UI state** (modals, dropdowns)

### Example 1: Counter

```go
Setup(func(ctx *Context) {
    counter := composables.UseState(ctx, 0)
    
    ctx.On("increment", func(_ interface{}) {
        counter.Set(counter.Get() + 1)
    })
    
    ctx.On("decrement", func(_ interface{}) {
        counter.Set(counter.Get() - 1)
    })
    
    ctx.Expose("count", counter.Value)
})
```

### Example 2: Toggle

```go
Setup(func(ctx *Context) {
    isOpen := composables.UseState(ctx, false)
    
    toggle := func() {
        isOpen.Set(!isOpen.Get())
    }
    
    ctx.On("toggle", func(_ interface{}) {
        toggle()
    })
    
    ctx.Expose("isOpen", isOpen.Value)
})
```

### Example 3: Input Field

```go
Setup(func(ctx *Context) {
    username := composables.UseState(ctx, "")
    
    ctx.On("input", func(value interface{}) {
        username.Set(value.(string))
    })
    
    ctx.Expose("username", username.Value)
})
```

## UseEffect

**Purpose:** Execute side effects with dependency tracking and cleanup.

**Signature:**
```go
func UseEffect(ctx *Context, effect func() UseEffectCleanup, deps ...*Ref[any])

type UseEffectCleanup func()
```

**Performance:** ~1.2μs per registration

### Use Cases

1. **Data synchronization**
2. **Logging state changes**
3. **External system integration**
4. **Cleanup on unmount**

### Example 1: Log Changes

```go
Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        fmt.Printf("Count changed to: %d\n", count.GetTyped())
        return nil
    }, count)
})
```

### Example 2: Sync to Backend

```go
Setup(func(ctx *Context) {
    settings := ctx.Ref(Settings{})
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        // Save to backend when settings change
        go api.SaveSettings(settings.GetTyped())
        return nil
    }, settings)
})
```

### Example 3: Subscription with Cleanup

```go
Setup(func(ctx *Context) {
    topic := ctx.Ref("default")
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        subscription := pubsub.Subscribe(topic.GetTyped())
        
        return func() {
            subscription.Unsubscribe()
        }
    }, topic)
})
```

## UseAsync

**Purpose:** Async data fetching with automatic loading and error state management.

**Signature:**
```go
func UseAsync[T any](ctx *Context, asyncFn func() (*T, error)) UseAsyncReturn[T]

type UseAsyncReturn[T any] struct {
    Data    *Ref[*T]       // Result data
    Loading *Ref[bool]     // Loading state
    Error   *Ref[error]    // Error state
    Execute func()         // Trigger execution
}
```

**Performance:** ~251ns per call (4x better than 1μs target)

### Use Cases

1. **API data fetching**
2. **Async file operations**
3. **Database queries**
4. **External service calls**

### Example 1: Fetch User Data

```go
Setup(func(ctx *Context) {
    userData := composables.UseAsync(ctx, func() (*User, error) {
        return api.FetchUser(userID)
    })
    
    ctx.OnMounted(func() {
        userData.Execute()
    })
    
    ctx.Expose("user", userData.Data)
    ctx.Expose("loading", userData.Loading)
    ctx.Expose("error", userData.Error)
})
```

### Example 2: Load with Retry

```go
Setup(func(ctx *Context) {
    posts := composables.UseAsync(ctx, func() (*[]Post, error) {
        return api.FetchPosts()
    })
    
    retry := func() {
        posts.Execute()
    }
    
    ctx.OnMounted(func() {
        posts.Execute()
    })
    
    ctx.On("retry", func(_ interface{}) {
        retry()
    })
    
    ctx.Expose("posts", posts.Data)
    ctx.Expose("loading", posts.Loading)
})
```

### Example 3: Conditional Fetch

```go
Setup(func(ctx *Context) {
    shouldFetch := ctx.Ref(false)
    
    data := composables.UseAsync(ctx, func() (*Data, error) {
        return api.FetchData()
    })
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        if shouldFetch.GetTyped() {
            data.Execute()
        }
        return nil
    }, shouldFetch)
})
```

## UseDebounce

**Purpose:** Create a debounced version of a reactive value.

**Signature:**
```go
func UseDebounce[T comparable](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T]
```

**Performance:** ~834ns per call

### Use Cases

1. **Search input** (wait for user to stop typing)
2. **Autosave** (wait for editing to pause)
3. **API calls** (reduce request frequency)
4. **Window resize** (wait for resize to finish)

### Example 1: Search Input

```go
Setup(func(ctx *Context) {
    searchTerm := ctx.Ref("")
    debouncedSearch := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        term := debouncedSearch.GetTyped()
        if term != "" {
            performSearch(term)
        }
        return nil
    }, debouncedSearch)
    
    ctx.Expose("searchTerm", searchTerm)
})
```

### Example 2: Autosave

```go
Setup(func(ctx *Context) {
    document := ctx.Ref(Document{})
    debouncedDoc := composables.UseDebounce(ctx, document, 2*time.Second)
    
    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
        saveDocument(debouncedDoc.GetTyped())
        return nil
    }, debouncedDoc)
})
```

## UseThrottle

**Purpose:** Limit function execution frequency.

**Signature:**
```go
func UseThrottle(ctx *Context, fn func(), delay time.Duration) func()
```

**Performance:** ~486ns per call

### Use Cases

1. **Scroll handlers**
2. **Window resize handlers**
3. **Mouse move tracking**
4. **Rate limiting**

### Example 1: Scroll Handler

```go
Setup(func(ctx *Context) {
    handleScroll := func() {
        updateScrollPosition()
    }
    
    throttledScroll := composables.UseThrottle(ctx, handleScroll, 100*time.Millisecond)
    
    ctx.On("scroll", func(_ interface{}) {
        throttledScroll()
    })
})
```

### Example 2: Analytics Tracking

```go
Setup(func(ctx *Context) {
    trackEvent := func() {
        analytics.Track("user_interaction")
    }
    
    throttledTrack := composables.UseThrottle(ctx, trackEvent, 1*time.Second)
    
    ctx.On("interact", func(_ interface{}) {
        throttledTrack()
    })
})
```

## UseForm

**Purpose:** Form state management with validation.

**Signature:**
```go
func UseForm[T any](ctx *Context, initial T, validator func(T) map[string]string) UseFormReturn[T]

type UseFormReturn[T any] struct {
    Values   *Ref[T]
    Errors   *Ref[map[string]string]
    IsValid  *Ref[bool]
    SetField func(fieldName string, value interface{})
    Submit   func()
    Reset    func()
}
```

**Performance:** ~740ns per call

### Use Cases

1. **Login forms**
2. **Registration forms**
3. **Settings forms**
4. **Survey forms**

### Example 1: Login Form

```go
type LoginForm struct {
    Email    string
    Password string
}

Setup(func(ctx *Context) {
    form := composables.UseForm(ctx, LoginForm{}, func(f LoginForm) map[string]string {
        errors := make(map[string]string)
        
        if !strings.Contains(f.Email, "@") {
            errors["Email"] = "Invalid email"
        }
        
        if len(f.Password) < 8 {
            errors["Password"] = "Too short"
        }
        
        return errors
    })
    
    ctx.On("submit", func(_ interface{}) {
        form.Submit()
        if form.IsValid.GetTyped() {
            login(form.Values.GetTyped())
        }
    })
    
    ctx.Expose("form", form)
})
```

### Example 2: Settings Form

```go
type SettingsForm struct {
    Theme      string
    Language   string
    EnableNotif bool
}

Setup(func(ctx *Context) {
    form := composables.UseForm(ctx, SettingsForm{
        Theme:      "dark",
        Language:   "en",
        EnableNotif: true,
    }, func(f SettingsForm) map[string]string {
        errors := make(map[string]string)
        
        validThemes := []string{"light", "dark", "auto"}
        if !contains(validThemes, f.Theme) {
            errors["Theme"] = "Invalid theme"
        }
        
        return errors
    })
    
    ctx.Expose("settings", form)
})
```

### Example 3: Dynamic Validation

```go
Setup(func(ctx *Context) {
    validateOnChange := ctx.Ref(true)
    
    form := composables.UseForm(ctx, UserForm{}, func(f UserForm) map[string]string {
        if !validateOnChange.GetTyped() {
            return make(map[string]string)
        }
        return validateUser(f)
    })
    
    ctx.Expose("form", form)
})
```

## UseLocalStorage

**Purpose:** Persistent state with automatic JSON serialization.

**Signature:**
```go
func UseLocalStorage[T any](ctx *Context, key string, initial T, storage Storage) UseLocalStorageReturn[T]

type UseLocalStorageReturn[T any] struct {
    Value *Ref[T]
    Set   func(T)
    Get   func() T
}
```

**Performance:** ~1.2μs per call (depends on storage backend)

### Use Cases

1. **User preferences**
2. **App settings**
3. **Cache data**
4. **Session state**

### Example 1: Theme Preference

```go
Setup(func(ctx *Context) {
    storage := composables.NewFileStorage("/path/to/data")
    theme := composables.UseLocalStorage(ctx, "theme", "light", storage)
    
    ctx.On("toggleTheme", func(_ interface{}) {
        if theme.Get() == "light" {
            theme.Set("dark")
        } else {
            theme.Set("light")
        }
    })
    
    ctx.Expose("theme", theme.Value)
})
```

### Example 2: Recent Items

```go
Setup(func(ctx *Context) {
    storage := composables.NewFileStorage("/path/to/data")
    recentFiles := composables.UseLocalStorage(ctx, "recent-files", []string{}, storage)
    
    addRecent := func(file string) {
        recent := recentFiles.Get()
        recent = append([]string{file}, recent...)
        if len(recent) > 10 {
            recent = recent[:10]
        }
        recentFiles.Set(recent)
    }
    
    ctx.Expose("recentFiles", recentFiles.Value)
})
```

## UseEventListener

**Purpose:** Event handling with automatic cleanup.

**Signature:**
```go
func UseEventListener(ctx *Context, eventName string, handler func()) func()
```

**Performance:** Minimal overhead

### Use Cases

1. **Click handlers**
2. **Keyboard shortcuts**
3. **Custom events**
4. **Event delegation**

### Example 1: Click Counter

```go
Setup(func(ctx *Context) {
    clickCount := ctx.Ref(0)
    
    cleanup := composables.UseEventListener(ctx, "click", func() {
        clickCount.Set(clickCount.GetTyped() + 1)
    })
    
    ctx.Expose("clicks", clickCount)
})
```

### Example 2: Keyboard Shortcut

```go
Setup(func(ctx *Context) {
    composables.UseEventListener(ctx, "keypress", func() {
        // Handle Ctrl+S for save
        saveDocument()
    })
})
```

## Composable Comparison

### When to Use Which?

**For simple state:**
- ✅ **UseState** - Single values, toggles, counters
- ❌ **UseForm** - Overkill for simple state

**For async operations:**
- ✅ **UseAsync** - Loading states needed, error handling
- ❌ **UseEffect** - Manual async handling required

**For rate limiting:**
- ✅ **UseDebounce** - Wait for input to settle (search, autosave)
- ✅ **UseThrottle** - Limit execution frequency (scroll, resize)
- Choose based on: Debounce = wait until quiet, Throttle = execute regularly

**For persistence:**
- ✅ **UseLocalStorage** - Need data between sessions
- ❌ **UseState** - Data lost on unmount

**For forms:**
- ✅ **UseForm** - Multiple fields, validation needed
- ❌ **UseState** - Simple single field

### Performance Considerations

**Fastest (< 100ns):**
- UseState
- UseEventListener

**Fast (< 500ns):**
- UseAsync
- UseThrottle

**Moderate (< 1.5μs):**
- UseDebounce
- UseForm
- UseEffect
- UseLocalStorage

## Best Practices

### 1. Choose the Right Composable

```go
// ✅ Good: Right tool for the job
searchTerm := ctx.Ref("")
debounced := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)

// ❌ Overkill: Manual debouncing
searchTerm := ctx.Ref("")
timer := ctx.Ref[*time.Timer](nil)
// ... complex manual debounce logic
```

### 2. Cleanup Resources

```go
// ✅ Good: Cleanup registered
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    subscription := subscribe()
    return func() {
        subscription.Unsubscribe()
    }
})

// ❌ Bad: No cleanup
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    subscribe()
    return nil // Resource leak!
})
```

### 3. Specify Dependencies

```go
// ✅ Good: Dependencies specified
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    fmt.Println(count.GetTyped())
    return nil
}, count)

// ❌ Bad: No dependencies, runs on every update
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    fmt.Println(count.GetTyped())
    return nil
})
```

### 4. Handle Errors

```go
// ✅ Good: Error handling
userData := composables.UseAsync(ctx, fetchUser)
composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    if userData.Error.GetTyped() != nil {
        showError(userData.Error.GetTyped())
    }
    return nil
}, userData.Error)

// ❌ Bad: Ignoring errors
userData := composables.UseAsync(ctx, fetchUser)
// Errors silently ignored
```

## Next Steps

- **[Composition API Guide](./composition-api.md)** - Learn core concepts
- **[Custom Composables Guide](./custom-composables.md)** - Build your own
- **[Lifecycle Hooks Guide](./lifecycle-hooks.md)** - Understand lifecycle integration
