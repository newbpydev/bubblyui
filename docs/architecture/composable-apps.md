# BubblyUI Composable App Architecture

**Vue-inspired patterns for building maintainable, reusable, and debuggable TUI applications**

## Table of Contents

1. [Introduction](#introduction)
2. [Directory Structure](#directory-structure)
3. [Component Pattern](#component-pattern)
4. [Composables](#composables)
5. [State Management](#state-management)
6. [Component Communication](#component-communication)
7. [Lifecycle Hooks](#lifecycle-hooks)
8. [DevTools Integration](#devtools-integration)
9. [Testing Strategy](#testing-strategy)
10. [Complete Example](#complete-example)

---

## Introduction

### Why Composable Architecture?

**Composable architecture** breaks applications into small, reusable, testable units. This approach:

- ✅ **Reduces complexity** - Small components are easier to understand
- ✅ **Enables reuse** - Components work across different contexts
- ✅ **Simplifies testing** - Unit test individual pieces
- ✅ **Improves debugging** - Dev tools show clear component hierarchy
- ✅ **Scales better** - Add features without breaking existing code

### BubblyUI ♥ Vue Patterns

BubblyUI draws inspiration from Vue's Composition API:

| Vue Concept | BubblyUI Equivalent |
|-------------|---------------------|
| `ref()` | `bubbly.NewRef[T]()` |
| `computed()` | `ctx.Computed()` |
| `watch()` | `bubbly.Watch()` |
| `watchEffect()` | `bubbly.WatchEffect()` |
| `onMounted()` | `ctx.OnMounted()` |
| `onUpdated()` | `ctx.OnUpdated()` |
| `onUnmounted()` | `ctx.OnUnmounted()` |
| Props | Props struct |
| `emit()` | Callback props |
| Composables | Use* functions |

---

## Directory Structure

### Recommended Structure

```
myapp/
├── main.go                 # Entry point (tea.Program setup)
├── app.go                  # Root component
├── components/             # Reusable UI components
│   ├── counter_display.go
│   ├── counter_controls.go
│   ├── todo_item.go
│   └── todo_list.go
├── composables/            # Shared reactive logic
│   ├── use_counter.go
│   ├── use_form.go
│   └── use_local_storage.go
└── utils/                  # Pure utility functions
    ├── formatters.go
    └── validators.go
```

### Small Apps (< 5 components)

```
simpleapp/
├── main.go                 # Entry + root component
└── components/
    ├── header.go
    └── footer.go
```

### Medium Apps (5-20 components)

```
mediumapp/
├── main.go
├── app.go                  # Root component
├── components/
│   ├── layout/             # Layout components
│   │   ├── header.go
│   │   └── footer.go
│   ├── forms/              # Form components
│   │   ├── input.go
│   │   └── checkbox.go
│   └── display/            # Display components
│       ├── card.go
│       └── table.go
├── composables/
│   └── use_form.go
└── utils/
    └── helpers.go
```

### Large Apps (> 20 components)

```
largeapp/
├── cmd/
│   └── myapp/
│       └── main.go
├── internal/
│   ├── app/                # Application layer
│   │   ├── app.go
│   │   └── router.go
│   ├── components/         # UI components
│   │   ├── common/
│   │   ├── todos/
│   │   └── settings/
│   ├── composables/        # Shared logic
│   │   ├── use_auth.go
│   │   └── use_api.go
│   ├── models/             # Data models
│   │   └── todo.go
│   └── services/           # Business logic
│       └── todo_service.go
└── pkg/                    # Public packages
    └── theme/
```

---

## Component Pattern

### Factory Function Pattern

**Always use factory functions to create components:**

```go
// CreateCounterDisplay creates a counter display component
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "CounterDisplay",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            // Component logic here
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    // Rendering logic
                },
            }
        },
    })
}
```

### Props Struct Pattern

**Define props as a struct (like Vue props):**

```go
type CounterDisplayProps struct {
    Count       *bubbly.Ref[int]
    OnIncrement func()
    OnDecrement func()
    Theme       string
}
```

**Benefits:**
- Type-safe props
- Clear component API
- Easy documentation
- IDE autocomplete

### Setup Function Pattern

**All reactive logic in Setup:**

```go
Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
    // 1. Get props
    count := props.Count
    
    // 2. Create local refs
    showDetails := bubbly.NewRef(false)
    
    // 3. Create computed values
    isEven := ctx.Computed(func() interface{} {
        return count.Get().(int)%2 == 0
    })
    
    // 4. Register lifecycle hooks
    ctx.OnMounted(func() {
        fmt.Println("Counter mounted!")
    })
    
    // 5. Register event handlers
    ctx.On("toggleDetails", func(_ interface{}) {
        current := showDetails.Get().(bool)
        showDetails.Set(!current)
    })
    
    // 6. Expose for template
    ctx.Expose("count", count)
    ctx.Expose("isEven", isEven)
    ctx.Expose("showDetails", showDetails)
    
    // 7. Return template
    return bubbly.SetupResult{
        Template: func(ctx bubbly.RenderContext) string {
            // Use exposed values
        },
    }
}
```

### Template Function Pattern

**Keep templates simple, use components:**

```go
Template: func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[int]).Get().(int)
    isEven := ctx.Get("isEven").(*bubbly.Computed[bool]).Get().(bool)
    
    // ✅ Use BubblyUI components
    card := components.Card(components.CardProps{
        Title:   "Counter",
        Content: fmt.Sprintf("Count: %d (%s)", count, evenOrOdd(isEven)),
    })
    card.Init()
    
    return card.View()
}
```

---

## Composables

### What are Composables?

Composables are **reusable functions** that encapsulate reactive logic. Like Vue's Composition API, they return Refs, Computed values, and methods.

### Naming Convention

Prefix composable functions with `Use`:
- `UseCounter`
- `UseForm`
- `UseTodos`
- `UseLocalStorage`
- `UseKeyboard`

### UseCounter Example

```go
package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

type CounterComposable struct {
    Count     *bubbly.Ref[int]
    Increment func()
    Decrement func()
    Reset     func()
    IsEven    *bubbly.Computed[bool]
}

func UseCounter(ctx bubbly.SetupContext, initial int) *CounterComposable {
    // Create reactive state
    count := bubbly.NewRef(initial)
    
    // Create computed value
    isEven := ctx.Computed(func() interface{} {
        return count.Get().(int)%2 == 0
    })
    
    // Define methods
    increment := func() {
        current := count.Get().(int)
        count.Set(current + 1)
    }
    
    decrement := func() {
        current := count.Get().(int)
        count.Set(current - 1)
    }
    
    reset := func() {
        count.Set(initial)
    }
    
    return &CounterComposable{
        Count:     count,
        Increment: increment,
        Decrement: decrement,
        Reset:     reset,
        IsEven:    isEven,
    }
}
```

### Using Composables in Components

```go
Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
    // Use counter composable
    counter := composables.UseCounter(ctx, 0)
    
    // Use in event handlers
    ctx.On("increment", func(_ interface{}) {
        counter.Increment()
    })
    
    // Expose for template
    ctx.Expose("counter", counter)
    
    return bubbly.SetupResult{
        Template: func(ctx bubbly.RenderContext) string {
            counter := ctx.Get("counter").(*composables.CounterComposable)
            count := counter.Count.Get().(int)
            
            return fmt.Sprintf("Count: %d", count)
        },
    }
}
```

### UseForm Example

```go
type FormComposable[T any] struct {
    Values    *bubbly.Ref[T]
    Errors    *bubbly.Ref[map[string]string]
    IsValid   *bubbly.Computed[bool]
    IsDirty   *bubbly.Ref[bool]
    Submit    func()
    Reset     func()
    SetField  func(field string, value interface{})
}

func UseForm[T any](
    ctx bubbly.SetupContext,
    initial T,
    validate func(T) map[string]string,
) *FormComposable[T] {
    values := bubbly.NewRef(initial)
    errors := bubbly.NewRef(make(map[string]string))
    isDirty := bubbly.NewRef(false)
    
    isValid := ctx.Computed(func() interface{} {
        errs := errors.Get().(map[string]string)
        return len(errs) == 0
    })
    
    submit := func() {
        data := values.Get().(T)
        errs := validate(data)
        errors.Set(errs)
    }
    
    reset := func() {
        values.Set(initial)
        errors.Set(make(map[string]string))
        isDirty.Set(false)
    }
    
    setField := func(field string, value interface{}) {
        isDirty.Set(true)
        // Update field in values using reflection
    }
    
    return &FormComposable[T]{
        Values:   values,
        Errors:   errors,
        IsValid:  isValid,
        IsDirty:  isDirty,
        Submit:   submit,
        Reset:    reset,
        SetField: setField,
    }
}
```

---

## State Management

### Ref for Simple State

```go
// ✅ CORRECT: Use NewRef with generic type
count := bubbly.NewRef(0)              // Ref[int]
username := bubbly.NewRef("")          // Ref[string]
isActive := bubbly.NewRef(true)        // Ref[bool]

// Access
value := count.Get().(int)
count.Set(value + 1)
```

### Computed for Derived State

```go
// Computed values update automatically
isEven := ctx.Computed(func() interface{} {
    return count.Get().(int)%2 == 0
})

fullName := ctx.Computed(func() interface{} {
    first := firstName.Get().(string)
    last := lastName.Get().(string)
    return first + " " + last
})

// Access (read-only)
even := isEven.Get().(bool)
```

### Watch for Side Effects

```go
// Watch a single ref
bubbly.Watch(count, func(newVal, oldVal interface{}) {
    fmt.Printf("Count changed: %v → %v\n", oldVal, newVal)
})

// Watch multiple refs
bubbly.Watch(firstName, func(newVal, oldVal interface{}) {
    fmt.Printf("First name changed\n")
})
bubbly.Watch(lastName, func(newVal, oldVal interface{}) {
    fmt.Printf("Last name changed\n")
})
```

### WatchEffect for Auto-Tracking

```go
// Automatically tracks accessed refs
bubbly.WatchEffect(func() {
    first := firstName.Get().(string)  // Tracked
    last := lastName.Get().(string)    // Tracked
    
    greeting := fmt.Sprintf("Hello, %s %s!", first, last)
    greetingRef.Set(greeting)
    
    // Re-runs when firstName OR lastName changes
})
```

---

## Component Communication

### Props Down, Events Up

**Parent → Child: Props**

```go
// Parent passes data via props
child, err := components.CreateCounterDisplay(components.CounterDisplayProps{
    Count: count,
    OnIncrement: func() {
        counter.Increment()
    },
})
```

**Child → Parent: Callback Props**

```go
// Child calls callback to notify parent
ctx.On("increment", func(_ interface{}) {
    if props.OnIncrement != nil {
        props.OnIncrement()  // Call parent's handler
    }
})
```

### Component Composition

```go
Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
    // Create child components
    display, _ := components.CreateCounterDisplay(components.CounterDisplayProps{
        Count: count,
    })
    
    controls, _ := components.CreateCounterControls(components.CounterControlsProps{
        OnIncrement: func() { counter.Increment() },
        OnDecrement: func() { counter.Decrement() },
    })
    
    // Auto-initialize and expose
    ctx.ExposeComponent("display", display)
    ctx.ExposeComponent("controls", controls)
    
    return bubbly.SetupResult{
        Template: func(ctx bubbly.RenderContext) string {
            display := ctx.Get("display").(bubbly.Component)
            controls := ctx.Get("controls").(bubbly.Component)
            
            return lipgloss.JoinVertical(
                lipgloss.Left,
                display.View(),
                controls.View(),
            )
        },
    }
}
```

---

## Lifecycle Hooks

### onMounted

```go
ctx.OnMounted(func() {
    fmt.Println("Component mounted - DOM ready")
    
    // Good for:
    // - Initialize external resources
    // - Start timers
    // - Fetch initial data
    // - Register global event listeners
})
```

### onUpdated

```go
ctx.OnUpdated(func() {
    fmt.Println("Component re-rendered")
    
    // Good for:
    // - React to prop/state changes
    // - Update external state
    // - Trigger animations
})
```

### onUnmounted

```go
ctx.OnUnmounted(func() {
    fmt.Println("Component unmounting - cleanup time")
    
    // CRITICAL for:
    // - Stop timers
    // - Close connections
    // - Unregister event listeners
    // - Free resources
    
    if timer != nil {
        timer.Stop()
    }
})
```

---

## DevTools Integration

### Best Practices for Debugging

**1. Use Descriptive Component Names**

```go
bubbly.DefineComponent(bubbly.ComponentConfig{
    Name: "TodoItem",  // Shows in component tree
    // ...
})
```

**2. Expose State for Inspection**

```go
// ✅ Exposed state visible in dev tools
ctx.Expose("todos", todos)
ctx.Expose("filter", filter)
ctx.Expose("completedCount", completedCount)

// ❌ Hidden state not visible
localVar := 42  // Won't show in dev tools
```

**3. Use Refs for All Reactive State**

```go
// ✅ Tracked by dev tools
count := bubbly.NewRef(0)

// ❌ Not tracked
var count int = 0
```

**4. Component Hierarchy Matters**

```go
// ✅ Clear hierarchy
// App
//   ├─ TodoList
//   │  ├─ TodoItem
//   │  └─ TodoItem
//   └─ Footer

// ❌ Flat, hard to navigate
// App
//   ├─ Component1
//   ├─ Component2
//   └─ Component3
```

### Enable DevTools in Examples

```go
func main() {
    // Zero-config enablement
    devtools.Enable()
    
    // Or with config
    config := devtools.DefaultConfig()
    config.LayoutMode = devtools.LayoutHorizontal
    devtools.EnableWithConfig(config)
    
    // Your app
    app, _ := CreateApp()
    p := tea.NewProgram(app, tea.WithAltScreen())
    p.Run()
}
```

---

## Testing Strategy

### Unit Test Composables

```go
func TestUseCounter(t *testing.T) {
    // Create mock setup context
    ctx := bubbly.NewMockSetupContext()
    
    // Use composable
    counter := composables.UseCounter(ctx, 0)
    
    // Test initial state
    assert.Equal(t, 0, counter.Count.Get().(int))
    
    // Test increment
    counter.Increment()
    assert.Equal(t, 1, counter.Count.Get().(int))
    
    // Test computed
    assert.False(t, counter.IsEven.Get().(bool))
}
```

### Integration Test Components

```go
func TestCounterDisplay(t *testing.T) {
    count := bubbly.NewRef(5)
    
    comp, err := components.CreateCounterDisplay(components.CounterDisplayProps{
        Count: count,
    })
    require.NoError(t, err)
    
    comp.Init()
    output := comp.View()
    
    assert.Contains(t, output, "5")
}
```

---

## Complete Example

### Counter App with Composables

**Directory Structure:**
```
counter-app/
├── main.go
├── app.go
├── components/
│   ├── counter_display.go
│   └── counter_controls.go
└── composables/
    └── use_counter.go
```

**main.go:**
```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Enable dev tools
    devtools.Enable()
    
    // Create root component
    app, err := CreateApp()
    if err != nil {
        panic(err)
    }
    
    // Run program
    p := tea.NewProgram(app, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

**app.go:**
```go
package main

import (
    "myapp/components"
    "myapp/composables"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/charmbracelet/lipgloss"
)

func CreateApp() (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "App",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            // Use counter composable
            counter := composables.UseCounter(ctx, 0)
            
            // Create child components
            display, _ := components.CreateCounterDisplay(components.CounterDisplayProps{
                Count: counter.Count,
            })
            
            controls, _ := components.CreateCounterControls(components.CounterControlsProps{
                OnIncrement: counter.Increment,
                OnDecrement: counter.Decrement,
                OnReset:     counter.Reset,
            })
            
            // Expose children
            ctx.ExposeComponent("display", display)
            ctx.ExposeComponent("controls", controls)
            
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    display := ctx.Get("display").(bubbly.Component)
                    controls := ctx.Get("controls").(bubbly.Component)
                    
                    return lipgloss.JoinVertical(
                        lipgloss.Center,
                        display.View(),
                        "",
                        controls.View(),
                    )
                },
            }
        },
    })
}
```

**composables/use_counter.go:**
```go
package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

type CounterComposable struct {
    Count     *bubbly.Ref[int]
    Increment func()
    Decrement func()
    Reset     func()
    IsEven    *bubbly.Computed[bool]
}

func UseCounter(ctx bubbly.SetupContext, initial int) *CounterComposable {
    count := bubbly.NewRef(initial)
    
    isEven := ctx.Computed(func() interface{} {
        return count.Get().(int)%2 == 0
    })
    
    return &CounterComposable{
        Count:  count,
        IsEven: isEven,
        Increment: func() {
            current := count.Get().(int)
            count.Set(current + 1)
        },
        Decrement: func() {
            current := count.Get().(int)
            count.Set(current - 1)
        },
        Reset: func() {
            count.Set(initial)
        },
    }
}
```

**components/counter_display.go:**
```go
package components

import (
    "fmt"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
)

type CounterDisplayProps struct {
    Count *bubbly.Ref[int]
}

func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "CounterDisplay",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            ctx.Expose("count", props.Count)
            
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    count := ctx.Get("count").(*bubbly.Ref[int]).Get().(int)
                    
                    card := components.Card(components.CardProps{
                        Title:   "Counter",
                        Content: fmt.Sprintf("Count: %d", count),
                    })
                    card.Init()
                    
                    return card.View()
                },
            }
        },
    })
}
```

**components/counter_controls.go:**
```go
package components

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
    "github.com/charmbracelet/lipgloss"
)

type CounterControlsProps struct {
    OnIncrement func()
    OnDecrement func()
    OnReset     func()
}

func CreateCounterControls(props CounterControlsProps) (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "CounterControls",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            // Register event handlers
            ctx.On("increment", func(_ interface{}) {
                if props.OnIncrement != nil {
                    props.OnIncrement()
                }
            })
            
            ctx.On("decrement", func(_ interface{}) {
                if props.OnDecrement != nil {
                    props.OnDecrement()
                }
            })
            
            ctx.On("reset", func(_ interface{}) {
                if props.OnReset != nil {
                    props.OnReset()
                }
            })
            
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    // Use Button components
                    incBtn := components.Button(components.ButtonProps{
                        Label:   "Increment",
                        OnPress: func() { ctx.Emit("increment", nil) },
                    })
                    incBtn.Init()
                    
                    decBtn := components.Button(components.ButtonProps{
                        Label:   "Decrement",
                        OnPress: func() { ctx.Emit("decrement", nil) },
                    })
                    decBtn.Init()
                    
                    resetBtn := components.Button(components.ButtonProps{
                        Label:   "Reset",
                        OnPress: func() { ctx.Emit("reset", nil) },
                    })
                    resetBtn.Init()
                    
                    return lipgloss.JoinHorizontal(
                        lipgloss.Center,
                        incBtn.View(),
                        "  ",
                        decBtn.View(),
                        "  ",
                        resetBtn.View(),
                    )
                },
            }
        },
    })
}
```

---

## Key Takeaways

### ✅ DO

1. **Use factory functions** - `CreateComponent(props)`
2. **Define props structs** - Type-safe, clear API
3. **Use composables** - Reuse reactive logic
4. **Expose state** - Visible in dev tools
5. **Use BubblyUI components** - Don't reinvent
6. **Clean up in onUnmounted** - Prevent leaks
7. **Test composables** - Unit test reusable logic
8. **Keep Setup focused** - One responsibility per component

### ❌ DON'T

1. **Don't use global state** - Use Ref/Computed
2. **Don't hardcode Lipgloss** - Use components
3. **Don't skip cleanup** - Always onUnmounted
4. **Don't mix concerns** - Separate UI from logic
5. **Don't ignore errors** - Handle component creation errors
6. **Don't create monoliths** - Break into small components

---

**Next Steps:**
- See [`cmd/examples/09-devtools/`](../../cmd/examples/09-devtools/) for complete examples
- Read [Component Reference](../components/README.md) for available components
- Check [DevTools Guide](../devtools/README.md) for debugging tips

**Questions?** Open an issue on GitHub or check the [FAQ](../devtools/troubleshooting.md#faq).
