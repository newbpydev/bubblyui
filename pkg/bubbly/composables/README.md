# BubblyUI Composables

**Reusable composition functions for Vue-inspired TUI components**

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Composables Overview (30 Total)](#composables-overview-30-total)
- [Standard Composables (8)](#standard-composables-8)
  - [UseState](#usestate)
  - [UseEffect](#useeffect)
  - [UseAsync](#useasync)
  - [UseDebounce](#usedebounce)
  - [UseThrottle](#usethrottle)
  - [UseForm](#useform)
  - [UseLocalStorage](#uselocalstorage)
  - [UseEventListener](#useeventlistener)
- [TUI-Specific Composables (5)](#tui-specific-composables-5)
  - [UseWindowSize](#usewindowsize)
  - [UseFocus](#usefocus)
  - [UseScroll](#usescroll)
  - [UseSelection](#useselection)
  - [UseMode](#usemode)
- [State Utility Composables (4)](#state-utility-composables-4)
  - [UseToggle](#usetoggle)
  - [UseCounter](#usecounter)
  - [UsePrevious](#useprevious)
  - [UseHistory](#usehistory)
- [Timing Composables (3)](#timing-composables-3)
  - [UseInterval](#useinterval)
  - [UseTimeout](#usetimeout)
  - [UseTimer](#usetimer)
- [Collection Composables (4)](#collection-composables-4)
  - [UseList](#uselist)
  - [UseMap](#usemap)
  - [UseSet](#useset)
  - [UseQueue](#usequeue)
- [Development Composables (2)](#development-composables-2)
  - [UseLogger](#uselogger)
  - [UseNotification](#usenotification)
- [Utility Composables (4)](#utility-composables-4)
  - [UseTextInput](#usetextinput)
  - [UseDoubleCounter](#usedoublecounter)
  - [CreateShared](#createshared)
  - [CreateSharedWithReset](#createsharedwithreset)
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

## Composables Overview (30 Total)

BubblyUI provides 30 composables organized into 6 categories:

| Category | Count | Composables |
|----------|-------|-------------|
| **Standard** | 8 | UseState, UseAsync, UseEffect, UseDebounce, UseThrottle, UseForm, UseLocalStorage, UseEventListener |
| **TUI-Specific** | 5 | UseWindowSize, UseFocus, UseScroll, UseSelection, UseMode |
| **State Utilities** | 4 | UseToggle, UseCounter, UsePrevious, UseHistory |
| **Timing** | 3 | UseInterval, UseTimeout, UseTimer |
| **Collections** | 4 | UseList, UseMap, UseSet, UseQueue |
| **Development** | 2 | UseLogger, UseNotification |
| **Utilities** | 4 | UseTextInput, UseDoubleCounter, CreateShared, CreateSharedWithReset |

---

## TUI-Specific Composables (5)

### UseWindowSize

**Terminal dimensions and responsive breakpoints.**

```go
windowSize := composables.UseWindowSize(ctx,
    composables.WithBreakpoints(composables.BreakpointConfig{XS: 0, SM: 60, MD: 80, LG: 120, XL: 160}),
    composables.WithMinDimensions(40, 10),
    composables.WithSidebarWidth(25),
)

// Reactive values
width := windowSize.Width.Get()           // int
height := windowSize.Height.Get()         // int
bp := windowSize.Breakpoint.Get()         // Breakpoint (xs, sm, md, lg, xl)
showSidebar := windowSize.SidebarVisible.Get()  // bool
cols := windowSize.GridColumns.Get()      // int

// Helper methods
contentWidth := windowSize.GetContentWidth()
cardWidth := windowSize.GetCardWidth()
windowSize.SetSize(width, height)  // Update on resize
```

### UseFocus

**Multi-pane focus management with generic type support.**

```go
type FocusPane int
const (FocusSidebar FocusPane = iota; FocusMain; FocusFooter)

focus := composables.UseFocus(ctx, FocusMain, []FocusPane{FocusSidebar, FocusMain, FocusFooter})

focus.Next()                    // Cycle to next
focus.Previous()                // Cycle to previous
focus.Focus(FocusSidebar)       // Focus specific pane
isFocused := focus.IsFocused(FocusMain)
current := focus.Current.Get()  // FocusPane
```

### UseScroll

**Viewport scrolling management.**

```go
scroll := composables.UseScroll(ctx, 100, 10)  // 100 items, 10 visible

scroll.ScrollUp()               // Move up by 1
scroll.ScrollDown()             // Move down by 1
scroll.ScrollTo(50)             // Jump to offset
scroll.ScrollToTop()            // Jump to start
scroll.ScrollToBottom()         // Jump to end
scroll.PageUp()                 // Move up by visible count
scroll.PageDown()               // Move down by visible count
scroll.SetTotalItems(200)       // Update total
scroll.SetVisibleCount(15)      // Update visible

isTop := scroll.IsAtTop()       // bool
isBottom := scroll.IsAtBottom() // bool
offset := scroll.Offset.Get()   // int
```

### UseSelection

**List/table selection with multi-select support.**

```go
selection := composables.UseSelection(ctx, items,
    composables.WithWrap(true),
    composables.WithMultiSelect(false),
)

selection.Select(1)             // Select index
selection.SelectNext()          // Move to next
selection.SelectPrevious()      // Move to previous
selection.ToggleSelection(2)    // Toggle at index (multi-select)
selection.ClearSelection()      // Clear all
selection.SetItems(newItems)    // Update items

idx := selection.SelectedIndex.Get()      // int
item := selection.SelectedItem.Get()      // T (computed)
indices := selection.SelectedIndices.Get() // []int (multi-select)
```

### UseMode

**Navigation/input mode management.**

```go
type Mode string
const (ModeNavigation Mode = "navigation"; ModeInput Mode = "input")

mode := composables.UseMode(ctx, ModeNavigation)

mode.Switch(ModeInput)                    // Change mode
mode.Toggle(ModeNavigation, ModeInput)    // Toggle between two
isNav := mode.IsMode(ModeNavigation)      // Check current

current := mode.Current.Get()   // Mode
previous := mode.Previous.Get() // Mode
```

---

## State Utility Composables (4)

### UseToggle

**Boolean toggle state management.**

```go
toggle := composables.UseToggle(ctx, false)

toggle.Toggle()     // Flip value
toggle.Set(true)    // Set explicit value
toggle.On()         // Set to true
toggle.Off()        // Set to false

isOn := toggle.Value.Get()  // bool
```

### UseCounter

**Bounded counter with step support.**

```go
counter := composables.UseCounter(ctx, 50,
    composables.WithMin(0),
    composables.WithMax(100),
    composables.WithStep(5),
)

counter.Increment()       // +step (respects max)
counter.Decrement()       // -step (respects min)
counter.IncrementBy(10)   // +10 (respects max)
counter.DecrementBy(10)   // -10 (respects min)
counter.Set(75)           // Set value (clamped)
counter.Reset()           // Reset to initial

count := counter.Count.Get()  // int
```

### UsePrevious

**Previous value tracking.**

```go
count := bubbly.NewRef(0)
previous := composables.UsePrevious(ctx, count)

count.Set(5)
count.Set(10)

prev := previous.Get()  // *T (nil if no previous, pointer to 5 after second Set)
```

### UseHistory

**Undo/redo state management.**

```go
history := composables.UseHistory(ctx, "initial", 50)  // Max 50 undo steps

history.Push("state 1")   // Add to history (clears redo)
history.Push("state 2")
history.Undo()            // Revert to previous
history.Redo()            // Restore next
history.Clear()           // Clear all history

current := history.Current.Get()    // T
canUndo := history.CanUndo.Get()    // bool (computed)
canRedo := history.CanRedo.Get()    // bool (computed)
```

---

## Timing Composables (3)

### UseInterval

**Periodic execution with start/stop control.**

```go
interval := composables.UseInterval(ctx, func() {
    refreshData()
}, 5*time.Second)

interval.Start()    // Begin interval
interval.Stop()     // Pause interval
interval.Toggle()   // Start if stopped, stop if running
interval.Reset()    // Stop and restart

isRunning := interval.IsRunning.Get()  // bool
// Auto-cleanup on unmount
```

### UseTimeout

**Delayed execution with cancel support.**

```go
timeout := composables.UseTimeout(ctx, func() {
    showNotification()
}, 3*time.Second)

timeout.Start()     // Begin timeout
timeout.Cancel()    // Cancel pending timeout
timeout.Reset()     // Cancel and restart

isPending := timeout.IsPending.Get()  // bool
isExpired := timeout.IsExpired.Get()  // bool
// Auto-cleanup on unmount
```

### UseTimer

**Countdown timer with progress tracking.**

```go
timer := composables.UseTimer(ctx, 60*time.Second,
    composables.WithOnExpire(func() { playAlarm() }),
    composables.WithTickInterval(100*time.Millisecond),
)

timer.Start()       // Begin countdown
timer.Stop()        // Pause countdown
timer.Reset()       // Reset to initial duration

remaining := timer.Remaining.Get()  // time.Duration
isRunning := timer.IsRunning.Get()  // bool
isExpired := timer.IsExpired.Get()  // bool (computed)
progress := timer.Progress.Get()    // float64 (0.0 to 1.0, computed)
// Auto-cleanup on unmount
```

---

## Collection Composables (4)

### UseList

**Generic list CRUD operations.**

```go
list := composables.UseList(ctx, []string{"a", "b", "c"})

list.Push("d", "e")           // Add to end
item, ok := list.Pop()        // Remove from end
item, ok := list.Shift()      // Remove from start
list.Unshift("z")             // Add to start
list.Insert(1, "x")           // Insert at index
item, ok := list.RemoveAt(2)  // Remove at index
list.UpdateAt(0, "new")       // Update at index
list.Clear()                  // Remove all
item, ok := list.Get(0)       // Get at index
list.Set(newItems)            // Replace all

items := list.Items.Get()     // []T
length := list.Length.Get()   // int (computed)
isEmpty := list.IsEmpty.Get() // bool (computed)
```

### UseMap

**Generic key-value state management.**

```go
m := composables.UseMap(ctx, map[string]int{"a": 1, "b": 2})

m.Set("c", 3)                 // Add/update key
val, ok := m.Get("a")         // Get value
deleted := m.Delete("b")      // Remove key
exists := m.Has("c")          // Check existence
keys := m.Keys()              // Get all keys
values := m.Values()          // Get all values
m.Clear()                     // Remove all

data := m.Data.Get()          // map[K]V
size := m.Size.Get()          // int (computed)
isEmpty := m.IsEmpty.Get()    // bool (computed)
```

### UseSet

**Unique value set management.**

```go
set := composables.UseSet(ctx, []string{"a", "b", "c"})

set.Add("d")                  // Add value
deleted := set.Delete("a")    // Remove value
exists := set.Has("b")        // Check existence
set.Toggle("c")               // Add if absent, remove if present
set.Clear()                   // Remove all
slice := set.ToSlice()        // Convert to slice

values := set.Values.Get()    // map[T]struct{}
size := set.Size.Get()        // int (computed)
isEmpty := set.IsEmpty.Get()  // bool (computed)
```

### UseQueue

**FIFO queue operations.**

```go
queue := composables.UseQueue(ctx, []int{1, 2, 3})

queue.Enqueue(4)              // Add to back
item, ok := queue.Dequeue()   // Remove from front
item, ok := queue.Peek()      // View front without removing
queue.Clear()                 // Remove all

items := queue.Items.Get()    // []T
size := queue.Size.Get()      // int (computed)
isEmpty := queue.IsEmpty.Get() // bool (computed)
front := queue.Front.Get()    // *T (computed, nil if empty)
```

---

## Development Composables (2)

### UseLogger

**Component debug logging with levels.**

```go
logger := composables.UseLogger(ctx, "MyComponent")

logger.Debug("Debug message", extraData)
logger.Info("Info message")
logger.Warn("Warning message")
logger.Error("Error message", err)
logger.Clear()  // Clear log history

logger.Level.Set(composables.LogLevelWarn)  // Only warn and error

logs := logger.Logs.Get()     // []LogEntry
level := logger.Level.Get()   // LogLevel
```

### UseNotification

**Toast notification system.**

```go
notifications := composables.UseNotification(ctx,
    composables.WithDefaultDuration(3*time.Second),
    composables.WithMaxNotifications(5),
)

notifications.Show(composables.NotificationSuccess, "Title", "Message", 5*time.Second)
notifications.Info("Info", "Information message")
notifications.Success("Success", "Operation completed")
notifications.Warning("Warning", "Be careful")
notifications.Error("Error", "Something went wrong")
notifications.Dismiss(id)     // Dismiss by ID
notifications.DismissAll()    // Dismiss all

notifs := notifications.Notifications.Get()  // []Notification
// Auto-dismiss after duration, auto-cleanup on unmount
```

---

## Utility Composables (4)

### UseTextInput

See [UseTextInput](#usetextinput-1) in Standard Composables section.

### UseDoubleCounter

```go
count, increment, decrement := composables.UseDoubleCounter(ctx, 0)
increment()  // +2
decrement()  // -2
```

### CreateShared

**Singleton composables for cross-component state.**

```go
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *CounterReturn {
        return composables.UseCounter(ctx, 0)
    },
)

// In Component A:
counter := UseSharedCounter(ctx)  // Creates instance
counter.Increment()

// In Component B (same instance!):
counter := UseSharedCounter(ctx)  // Returns existing instance
```

### CreateSharedWithReset

**Resettable singleton composables.**

```go
shared := composables.CreateSharedWithReset(
    func(ctx *bubbly.Context) *CounterReturn {
        return composables.UseCounter(ctx, 0)
    },
)

counter := shared.Use(ctx)  // Get or create instance
shared.Reset()              // Reset to allow new instance
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
