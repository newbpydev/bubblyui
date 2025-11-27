# BubblyUI Compact Manual for AI Agents

**Version:** 3.1 | **Updated:** November 26, 2025 | **Status:** VERIFIED  
**Philosophy:** Zero boilerplate TUI framework with Vue-inspired composables

---

## 🚨 CRITICAL RULES

1. **Use `bubbly.Run()`** - Zero boilerplate, no manual Init/Update/View
2. **Use `bubbly.NewRef[T]()`** - Type-safe refs, avoid `ctx.Ref()` (interface{})
3. **Use `ctx.ExposeComponent()`** - Auto-init + parent-child relationship
4. **Always call `.Init()`** before `.View()` on pkg/components
5. **Use components package** - Don't reinvent with raw Lipgloss

---

## Essential Imports

```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
    composables "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
    directives "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
    csrouter "github.com/newbpydev/bubblyui/pkg/bubbly/router"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)
```

---

## App Structure

```
myapp/
├── main.go           # bubbly.Run(app, bubbly.WithAltScreen())
├── app.go            # Root component with key bindings
├── composables/      # Reusable logic (UseCounter, UseAuth, etc.)
└── components/       # Component factories with typed Props
```

**main.go:**
```go
func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())  // Zero boilerplate!
}
```

---

## Component Builder (12 Methods)

```go
bubbly.NewComponent("Name").
    Props(props).                              // Set props
    Setup(func(ctx *bubbly.Context) {...}).   // REQUIRED: State, events, lifecycle
    Template(func(ctx *bubbly.RenderContext) string {...}). // REQUIRED: Render
    WithAutoCommands(true).                    // Auto UI updates on Ref.Set()
    WithKeyBinding(key, event, desc).          // Single key binding
    WithMultiKeyBindings(event, desc, keys...). // Multiple keys → same event (NEW)
    WithConditionalKeyBinding(binding).        // Mode-based keys
    WithKeyBindings(map).                      // Batch bindings
    WithMessageHandler(handler).               // Custom tea.Msg handling
    Children(child1, child2).                  // Add children
    Build()                                    // Returns (Component, error)
```

---

## Refs & Reactivity

```go
// Type-safe (PREFERRED)
count := bubbly.NewRef(0)           // *Ref[int]
count.Set(42)
current := count.Get()              // int, no assertion

// Context refs (interface{})
count := ctx.Ref(0)                 // *Ref[interface{}]
current := count.Get().(int)        // Needs assertion

// Computed values
doubled := ctx.Computed(func() interface{} {
    return count.Get() * 2
})

// Watch changes
cleanup := ctx.Watch(count, func(new, old interface{}) {
    fmt.Printf("%v → %v\n", old, new)
})
```

---

## Context API (28 Methods)

| Category | Methods |
|----------|---------|
| **State** | `Ref()`, `ManualRef()`, `Computed()`, `Watch()` |
| **Expose** | `Expose()`, `ExposeComponent()`, `Get()`, `Set()` |
| **Events** | `On()`, `Emit()` |
| **Lifecycle** | `OnMounted()`, `OnUpdated()`, `OnUnmounted()`, `OnBeforeUpdate()`, `OnBeforeUnmount()`, `OnCleanup()` |
| **DI** | `Provide()`, `Inject()`, `ProvideTheme()`, `UseTheme()` |
| **Tree** | `Props()`, `Children()`, `Parent()` |
| **Commands** | `EnableAutoCommands()`, `DisableAutoCommands()`, `IsAutoCommandsEnabled()`, `SetCommandGenerator()` |

---

## Theme System (NEW - 94% Code Reduction)

```go
// Parent provides theme
ctx.ProvideTheme(bubbly.DefaultTheme)

// Or customize:
theme := bubbly.DefaultTheme
theme.Primary = lipgloss.Color("99")
ctx.ProvideTheme(theme)

// Child uses theme
theme := ctx.UseTheme(bubbly.DefaultTheme)  // Falls back if not provided
style := lipgloss.NewStyle().Foreground(theme.Primary)

// Theme struct fields:
// Primary, Secondary, Muted, Warning, Error, Success, Background
```

---

## Lifecycle Hooks

```go
ctx.OnMounted(func() {
    // After first render - fetch data, start timers
})

ctx.OnUpdated(func() {
    // After any update
}, deps...)  // Optional: only when deps change

ctx.OnUnmounted(func() {
    // Cleanup: stop timers, close connections
})

// Or use UseEffect composable:
cleanup := composables.UseEffect(ctx, func() composables.UseEffectCleanup {
    ticker := time.NewTicker(time.Second)
    return func() { ticker.Stop() }  // Cleanup function
}, deps...)
```

---

## Composables (12 Total)

| Composable | Signature | Purpose |
|------------|-----------|---------|
| `UseState[T]` | `(ctx, initial T)` | Simple reactive state |
| `UseAsync[T]` | `(ctx, fetcher)` | Async data with Loading/Error/Data |
| `UseEffect` | `(ctx, effect, deps...)` | Side effects with cleanup |
| `UseDebounce[T]` | `(ctx, ref, delay)` | Debounced ref updates |
| `UseThrottle` | `(ctx, fn, delay)` | Throttled function execution |
| `UseForm[T]` | `(ctx, form, validator)` | Form with validation |
| `UseLocalStorage[T]` | `(ctx, key, initial, storage)` | Persistent state |
| `UseEventListener` | `(ctx, event, handler)` | Event subscription |
| `UseTextInput` | `(config)` | Bubbles textinput wrapper |
| `UseCounter` | `(ctx, initial)` | Counter with Inc/Dec/Reset |
| `UseDoubleCounter` | `(ctx, initial)` | Counter with ±2 steps |
| `CreateShared[T]` | `(factory)` | **NEW:** Singleton composable |

**CreateShared Example:**
```go
var UseSharedCounter = composables.CreateShared(func(ctx *bubbly.Context) *CounterComposable {
    return UseCounter(ctx, 0)
})
// Same instance across all components - thread-safe via sync.Once
```

---

## Components Package (27 Components)

### Atoms
```go
components.Button(ButtonProps{Label, Variant, OnClick, Disabled})
components.Text(TextProps{Content, Color})
components.Chip(ChipProps{Label, Variant})  // Badge
components.Icon(IconProps{Icon, Color})
components.Spacer(SpacerProps{Height})
```

### Form Inputs (Require Ref)
```go
components.Input(InputProps{Value: ref, Placeholder, Type, Width, Validate, OnChange})
components.Toggle(ToggleProps{Label, Value: ref, OnChange})  // NOT Checked!
components.Checkbox(CheckboxProps{Label, Checked: ref, OnChange})
components.Radio(RadioProps{Options, Selected: ref, OnChange})
components.Select(SelectProps{Options, Selected: ref, OnChange})
components.Textarea(TextareaProps{Value: ref, Placeholder, Width, Height})
```

### Organisms
```go
components.Card(CardProps{Title, Content, BorderStyle, Width})
components.List(ListProps{Items, SelectedIndex, OnSelect})
components.Table(TableProps{Headers, Rows, SelectedRow, OnSelect})
components.Tabs(TabsProps{Titles, SelectedIndex, OnSelect})
components.Accordion(AccordionProps{Items, OpenIndex, OnToggle})
components.Modal(ModalProps{Title, Content, Visible: ref, OnConfirm, OnCancel})
components.Form(FormProps{Fields, OnSubmit, SubmitButton})
```

### Layouts
```go
components.AppLayout(AppLayoutProps{Header, Sidebar, Main, Footer})
components.PageLayout(PageLayoutProps{Title, Content, Actions})
components.GridLayout(GridLayoutProps{Columns, Rows, Cells, Border})
```

**Usage Pattern:**
```go
card := components.Card(props)
card.Init()  // REQUIRED before View()
return card.View()
```

---

## Key Bindings

```go
// Single key
.WithKeyBinding(" ", "increment", "Increment")  // Space key = " "

// Multiple keys → same event (NEW - 67% reduction)
.WithMultiKeyBindings("increment", "Increment", "up", "k", "+")
.WithMultiKeyBindings("decrement", "Decrement", "down", "j", "-")

// Conditional (mode-based)
.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key: " ", Event: "toggle", Description: "Toggle",
    Condition: func() bool { return !inputMode },
})

// Key strings: "a", " ", "enter", "ctrl+c", "alt+x", "up", "down", etc.
```

---

## Router (csrouter)

```go
router := csrouter.NewRouter().
    AddRoute("/", homeComponent).
    AddRoute("/users/:id", userComponent).
    WithNotFound(notFoundComponent).
    WithGuard(authGuard).
    Build()

// Navigation
router.Navigate("/users/123")
router.GoBack()
router.CurrentRoute()  // *Route with Path, Params, Query

// In component
route := ctx.Get("route").(*csrouter.Route)
userID := route.Params["id"]
query := route.Query.Get("page")
```

---

## Directives

```go
directives.If(condition, trueVal, falseVal)  // Conditional string
directives.Show(condition, content)           // Show/hide
directives.ForEach(slice, func(item, i) string {...})  // List iteration
```

---

## Events & Dependency Injection

```go
// Events (bubble up to parent)
ctx.On("submit", func(data interface{}) { ... })
ctx.Emit("submit", FormData{...})

// Provide/Inject (down the tree)
ctx.Provide("apiClient", client)
client := ctx.Inject("apiClient", nil).(*APIClient)
```

---

## Auto Commands (Batching)

```go
// Enable auto UI updates
builder.WithAutoCommands(true)
count.Set(5)  // Auto re-render

// Disable for batch operations
ctx.DisableAutoCommands()
for i := 0; i < 1000; i++ {
    count.Set(i)  // No re-renders
}
ctx.EnableAutoCommands()
ctx.Emit("batchComplete", nil)  // Single re-render
```

---

## Testing

```go
func TestCounter(t *testing.T) {
    tests := []struct {
        name     string
        initial  int
        action   string
        expected int
    }{
        {"increment", 0, "inc", 1},
        {"decrement", 5, "dec", 4},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            count := bubbly.NewRef(tt.initial)
            // ... test logic
            assert.Equal(t, tt.expected, count.Get())
        })
    }
}

// Run: go test -race ./...
// Coverage: go test -cover ./...
```

---

## ❌ Anti-Patterns

| Wrong | Right |
|-------|-------|
| `ctx.Ref(0)` everywhere | `bubbly.NewRef(0)` (type-safe) |
| `card.View()` without Init | `card.Init(); card.View()` |
| Manual `ctx.Emit("update")` | `WithAutoCommands(true)` |
| Raw Lipgloss for cards | `components.Card(props)` |
| `Toggle{Checked: ref}` | `Toggle{Value: ref}` |
| No cleanup for timers | `ctx.OnUnmounted()` to stop |
| Inline component creation | Separate files per component |
| 1000 `count.Set()` calls | Batch with `DisableAutoCommands()` |

---

## Migration Quick Reference

| Old Pattern | New Pattern | Reduction |
|-------------|-------------|-----------|
| 5× `ctx.Provide("*Color")` | `ctx.ProvideTheme(theme)` | 80% |
| 15 lines inject+expose colors | `ctx.UseTheme(default)` | 94% |
| 6× `WithKeyBinding` same event | `WithMultiKeyBindings(...)` | 67% |
| Separate composable instances | `CreateShared(factory)` | Singleton |

---

## Quick Start Template

```go
// main.go
package main

import (
    "fmt"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
)

func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())
}

// app.go
func CreateApp() (bubbly.Component, error) {
    return bubbly.NewComponent("App").
        WithAutoCommands(true).
        WithKeyBinding("ctrl+c", "quit", "Quit").
        WithMultiKeyBindings("increment", "Inc", "up", "k", "+").
        Setup(func(ctx *bubbly.Context) {
            count := bubbly.NewRef(0)
            ctx.ProvideTheme(bubbly.DefaultTheme)
            ctx.Expose("count", count)
            ctx.On("increment", func(_ interface{}) {
                count.Set(count.Get() + 1)
            })
        }).
        Template(func(ctx *bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[int]).Get()
            card := components.Card(components.CardProps{
                Title:   "Counter",
                Content: fmt.Sprintf("Count: %d", count),
            })
            card.Init()
            return card.View()
        }).
        Build()
}
```

---

**Total: ~450 lines vs 2600+ lines (83% reduction) - All essential APIs preserved**
