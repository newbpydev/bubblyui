# Migrating from Bubbletea to BubblyUI

**A comprehensive guide for Bubbletea developers adopting BubblyUI's Vue-inspired component model.**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Why Migrate?](#why-migrate)
3. [Before/After Comparisons](#beforeafter-comparisons)
4. [Step-by-Step Migration](#step-by-step-migration)
5. [Common Patterns](#common-patterns)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)
8. [Error Tracking & Observability](#error-tracking--observability)

---

## Introduction

BubblyUI is a **Vue-inspired TUI framework** built on top of Bubbletea. It maintains 100% compatibility with Bubbletea's Elm architecture while adding:

- **Reactive State Management** - Ref, Computed, Watch (like Vue 3)
- **Component Model** - Reusable, composable UI components
- **Builder API** - Fluent, declarative component creation
- **Event System** - Type-safe event emission and bubbling
- **Context System** - Scoped state and lifecycle management
- **Error Tracking** - Production-ready observability (Sentry, Console)

**Key Principle:** BubblyUI components **are** Bubbletea models. Every component implements `tea.Model`, so you can mix and match freely.

---

## Why Migrate?

### What You Gain

✅ **Reactive State** - Automatic UI updates when state changes  
✅ **Component Reusability** - Build once, use everywhere  
✅ **Better Organization** - Separate concerns with Setup/Template pattern  
✅ **Type Safety** - Go 1.22+ generics for compile-time safety  
✅ **Event-Driven** - Decouple components with event bubbling  
✅ **Production Ready** - Built-in error tracking and observability  

### What Stays the Same

✅ **Bubbletea Core** - Init/Update/View lifecycle unchanged  
✅ **Commands** - `tea.Cmd` and `tea.Batch` work identically  
✅ **Messages** - All Bubbletea messages supported  
✅ **Lipgloss** - Same styling library  
✅ **Performance** - Zero overhead, same speed  

---

## Before/After Comparisons

### Example 1: Simple Counter

#### Before (Pure Bubbletea)

```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    count int
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            m.count++
        case "down", "j":
            m.count--
        case "r":
            m.count = 0
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("Count: %d\n\nPress ↑/↓ to change, r to reset, q to quit", m.count)
}

func main() {
    tea.NewProgram(model{}).Run()
}
```

#### After (BubblyUI)

```go
package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

type model struct {
    counter bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.counter.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            m.counter.Emit("increment", nil)
        case "down", "j":
            m.counter.Emit("decrement", nil)
        case "r":
            m.counter.Emit("reset", nil)
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    
    updatedComponent, cmd := m.counter.Update(msg)
    m.counter = updatedComponent.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.counter.View()
}

func createCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        Setup(func(ctx *bubbly.Context) {
            // Reactive state
            count := ctx.Ref(0)
            
            // Event handlers
            ctx.On("increment", func(e *bubbly.Event) {
                count.Set(count.Get().(int) + 1)
            })
            
            ctx.On("decrement", func(e *bubbly.Event) {
                count.Set(count.Get().(int) - 1)
            })
            
            ctx.On("reset", func(e *bubbly.Event) {
                count.Set(0)
            })
            
            // Expose for template
            ctx.Expose("count", count)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[interface{}])
            return fmt.Sprintf(
                "Count: %d\n\nPress ↑/↓ to change, r to reset, q to quit",
                count.Get().(int),
            )
        }).
        Build()
}

func main() {
    counter, _ := createCounter()
    tea.NewProgram(model{counter: counter}).Run()
}
```

**Key Differences:**
- ✅ **Reactive State** - `ctx.Ref()` automatically triggers re-renders
- ✅ **Event-Driven** - `Emit()` and `On()` decouple logic
- ✅ **Reusable** - Counter component can be used in multiple apps
- ✅ **Testable** - Setup and Template are pure functions

---

### Example 2: Component Composition

#### Before (Bubbletea - Manual Composition)

```go
type model struct {
    spinner spinner.Model
    list    list.Model
    input   textinput.Model
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
        m.input.Blink,
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Update each component manually
    var cmd tea.Cmd
    m.spinner, cmd = m.spinner.Update(msg)
    cmds = append(cmds, cmd)
    
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)
    
    m.input, cmd = m.input.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m model) View() string {
    return fmt.Sprintf("%s\n%s\n%s", 
        m.spinner.View(),
        m.list.View(),
        m.input.View(),
    )
}
```

#### After (BubblyUI - Automatic Composition)

```go
func createApp() (bubbly.Component, error) {
    spinner, _ := createSpinner()
    list, _ := createList()
    input, _ := createInput()
    
    return bubbly.NewComponent("App").
        Children(spinner, list, input).  // Automatic lifecycle management
        Template(func(ctx bubbly.RenderContext) string {
            children := ctx.Children()
            return fmt.Sprintf("%s\n%s\n%s",
                children[0].View(),
                children[1].View(),
                children[2].View(),
            )
        }).
        Build()
}
```

**Benefits:**
- ✅ **Automatic Init** - Children Init() called automatically
- ✅ **Automatic Update** - Children Update() batched automatically
- ✅ **Event Bubbling** - Child events bubble to parent
- ✅ **Less Boilerplate** - No manual command batching

---

## Step-by-Step Migration

### Phase 1: Add BubblyUI Dependency

```bash
go get github.com/newbpydev/bubblyui
```

### Phase 2: Identify Components

Look for these patterns in your Bubbletea code:

1. **Repeated UI patterns** → Extract to component
2. **Nested models** → Use component composition
3. **Manual state management** → Use reactive state
4. **Message passing** → Use event system

### Phase 3: Incremental Migration

**Strategy:** Migrate one component at a time, keeping the rest as Bubbletea.

#### Step 1: Wrap Existing Model

```go
// Before: Pure Bubbletea model
type oldModel struct {
    // ... fields
}

// After: Wrap in BubblyUI component
func createOldComponent() (bubbly.Component, error) {
    return bubbly.NewComponent("OldModel").
        Setup(func(ctx *bubbly.Context) {
            // Initialize old model fields as Refs
            oldModel := ctx.Ref(&oldModel{})
            ctx.Expose("model", oldModel)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            model := ctx.Get("model").(*bubbly.Ref[interface{}])
            m := model.Get().(*oldModel)
            // Use old View() logic
            return m.View()
        }).
        Build()
}
```

#### Step 2: Extract Reusable Components

```go
// Identify repeated patterns
// Before: Inline spinner in multiple places
func (m model) View() string {
    if m.loading {
        return "⠋ Loading..."
    }
    return m.content
}

// After: Reusable spinner component
func createSpinner() (bubbly.Component, error) {
    return bubbly.NewComponent("Spinner").
        Setup(func(ctx *bubbly.Context) {
            loading := ctx.Ref(true)
            ctx.Expose("loading", loading)
            
            ctx.On("start", func(e *bubbly.Event) {
                loading.Set(true)
            })
            
            ctx.On("stop", func(e *bubbly.Event) {
                loading.Set(false)
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            loading := ctx.Get("loading").(*bubbly.Ref[interface{}])
            if loading.Get().(bool) {
                return "⠋ Loading..."
            }
            return ""
        }).
        Build()
}
```

#### Step 3: Add Reactive State

```go
// Before: Manual state updates
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dataMsg:
        m.data = msg.data
        m.loading = false
        // Manual re-render triggered
    }
    return m, nil
}

// After: Automatic reactive updates
Setup(func(ctx *bubbly.Context) {
    data := ctx.Ref([]string{})
    loading := ctx.Ref(true)
    
    ctx.On("dataLoaded", func(e *bubbly.Event) {
        data.Set(e.Data.([]string))  // Automatic re-render
        loading.Set(false)            // Automatic re-render
    })
    
    ctx.Expose("data", data)
    ctx.Expose("loading", loading)
})
```

#### Step 4: Use Event System

```go
// Before: Direct function calls
func (m *parentModel) handleChildAction() {
    m.childModel.doSomething()
    m.updateParentState()
}

// After: Event-driven communication
// Child emits event
ctx.Emit("action", actionData)

// Parent listens
ctx.On("action", func(e *bubbly.Event) {
    // Handle child action
    // Events bubble automatically
})
```

### Phase 4: Full Migration Checklist

- [ ] All models converted to components
- [ ] State management uses Ref/Computed
- [ ] Component communication uses events
- [ ] Children managed via ComponentBuilder
- [ ] Templates use RenderContext
- [ ] Tests updated for new structure
- [ ] Error tracking configured (optional)

---

## Common Patterns

### Pattern 1: Form Validation

```go
func createForm() (bubbly.Component, error) {
    return bubbly.NewComponent("Form").
        Setup(func(ctx *bubbly.Context) {
            email := ctx.Ref("")
            password := ctx.Ref("")
            
            // Computed validation
            emailValid := ctx.Computed(func() interface{} {
                e := email.Get().(string)
                return len(e) > 0 && strings.Contains(e, "@")
            })
            
            passwordValid := ctx.Computed(func() interface{} {
                p := password.Get().(string)
                return len(p) >= 8
            })
            
            formValid := ctx.Computed(func() interface{} {
                return emailValid.Get().(bool) && passwordValid.Get().(bool)
            })
            
            ctx.Expose("email", email)
            ctx.Expose("password", password)
            ctx.Expose("emailValid", emailValid)
            ctx.Expose("passwordValid", passwordValid)
            ctx.Expose("formValid", formValid)
            
            ctx.On("submit", func(e *bubbly.Event) {
                if formValid.Get().(bool) {
                    // Submit form
                }
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            // Render form with validation feedback
            emailValid := ctx.Get("emailValid").(*bubbly.Computed[interface{}])
            passwordValid := ctx.Get("passwordValid").(*bubbly.Computed[interface{}])
            
            emailIndicator := "✗"
            if emailValid.Get().(bool) {
                emailIndicator = "✓"
            }
            
            passwordIndicator := "✗"
            if passwordValid.Get().(bool) {
                passwordIndicator = "✓"
            }
            
            return fmt.Sprintf("Email: %s\nPassword: %s", 
                emailIndicator, passwordIndicator)
        }).
        Build()
}
```

### Pattern 2: List with Selection

```go
func createList() (bubbly.Component, error) {
    return bubbly.NewComponent("List").
        Setup(func(ctx *bubbly.Context) {
            items := ctx.Ref([]string{"Item 1", "Item 2", "Item 3"})
            selectedIndex := ctx.Ref(0)
            
            ctx.On("selectNext", func(e *bubbly.Event) {
                idx := selectedIndex.Get().(int)
                itemCount := len(items.Get().([]string))
                if idx < itemCount-1 {
                    selectedIndex.Set(idx + 1)
                }
            })
            
            ctx.On("selectPrev", func(e *bubbly.Event) {
                idx := selectedIndex.Get().(int)
                if idx > 0 {
                    selectedIndex.Set(idx - 1)
                }
            })
            
            ctx.Expose("items", items)
            ctx.Expose("selectedIndex", selectedIndex)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            items := ctx.Get("items").(*bubbly.Ref[interface{}])
            selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[interface{}])
            
            var output string
            for i, item := range items.Get().([]string) {
                cursor := " "
                if i == selectedIndex.Get().(int) {
                    cursor = ">"
                }
                output += fmt.Sprintf("%s %s\n", cursor, item)
            }
            return output
        }).
        Build()
}
```

### Pattern 3: Parent-Child Communication

```go
// Parent component
func createParent() (bubbly.Component, error) {
    child, _ := createChild()
    
    return bubbly.NewComponent("Parent").
        Children(child).
        Setup(func(ctx *bubbly.Context) {
            parentState := ctx.Ref("initial")
            
            // Listen to child events (automatic bubbling)
            ctx.On("childAction", func(e *bubbly.Event) {
                data := e.Data.(string)
                parentState.Set(data)
            })
            
            ctx.Expose("parentState", parentState)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            children := ctx.Children()
            parentState := ctx.Get("parentState").(*bubbly.Ref[interface{}])
            
            return fmt.Sprintf("Parent: %s\n%s",
                parentState.Get().(string),
                children[0].View(),
            )
        }).
        Build()
}

// Child component
func createChild() (bubbly.Component, error) {
    return bubbly.NewComponent("Child").
        Setup(func(ctx *bubbly.Context) {
            ctx.On("trigger", func(e *bubbly.Event) {
                // Emit event that bubbles to parent
                ctx.Emit("childAction", "data from child")
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            return "Child Component"
        }).
        Build()
}
```

---

## Troubleshooting

### Issue 1: Component Not Re-rendering

**Problem:** State changes but UI doesn't update.

**Solution:** Ensure you're using `Ref.Set()` instead of direct mutation.

```go
// ❌ Wrong - Direct mutation doesn't trigger re-render
items := ctx.Ref([]string{"a", "b"})
slice := items.Get().([]string)
slice = append(slice, "c")  // UI won't update!

// ✅ Correct - Use Set() to trigger re-render
items := ctx.Ref([]string{"a", "b"})
slice := items.Get().([]string)
newSlice := append(slice, "c")
items.Set(newSlice)  // UI updates!
```

### Issue 2: Events Not Bubbling

**Problem:** Child events not reaching parent.

**Solution:** Ensure parent reference is set via `ComponentBuilder.Children()`.

```go
// ❌ Wrong - Manual child creation doesn't set parent
child, _ := createChild()
// Parent reference not set, events won't bubble

// ✅ Correct - Use Children() method
bubbly.NewComponent("Parent").
    Children(child).  // Sets parent reference automatically
    Build()
```

### Issue 3: Type Assertions Panicking

**Problem:** `interface{}` type assertions fail.

**Solution:** Always check type assertions or use type-safe wrappers.

```go
// ❌ Risky - Can panic
count := ctx.Get("count").(*bubbly.Ref[interface{}])
value := count.Get().(int)  // Panics if not int

// ✅ Safe - Check assertion
count := ctx.Get("count").(*bubbly.Ref[interface{}])
if value, ok := count.Get().(int); ok {
    // Use value safely
}
```

### Issue 4: Memory Leaks with Watchers

**Problem:** Watchers not cleaned up.

**Solution:** Use `Watch()` sparingly and clean up when needed.

```go
Setup(func(ctx *bubbly.Context) {
    count := ctx.Ref(0)
    
    // Watcher is automatically cleaned up when component unmounts
    ctx.Watch(count, func(oldVal, newVal interface{}) {
        fmt.Printf("Count changed: %v -> %v\n", oldVal, newVal)
    })
})
```

### Issue 5: Infinite Event Loops

**Problem:** Event handlers emit events that trigger themselves.

**Solution:** Use `StopPropagation()` or avoid emitting in handlers.

```go
// ❌ Wrong - Infinite loop
ctx.On("update", func(e *bubbly.Event) {
    ctx.Emit("update", nil)  // Triggers itself!
})

// ✅ Correct - Stop propagation or use different events
ctx.On("update", func(e *bubbly.Event) {
    e.StopPropagation()  // Prevent bubbling
    // Or emit different event
    ctx.Emit("updateComplete", nil)
})
```

---

## Best Practices

### 1. Component Naming

```go
// ✅ Use PascalCase for component names
bubbly.NewComponent("UserProfile")
bubbly.NewComponent("TodoList")
bubbly.NewComponent("SearchBar")

// ❌ Avoid generic names
bubbly.NewComponent("Component")
bubbly.NewComponent("Thing")
```

### 2. State Organization

```go
// ✅ Keep state minimal and focused
Setup(func(ctx *bubbly.Context) {
    // Only essential state
    username := ctx.Ref("")
    isValid := ctx.Computed(func() interface{} {
        return len(username.Get().(string)) > 0
    })
})

// ❌ Avoid storing derived state
Setup(func(ctx *bubbly.Context) {
    username := ctx.Ref("")
    usernameLength := ctx.Ref(0)  // ❌ Derive from username instead
    usernameValid := ctx.Ref(false)  // ❌ Use Computed instead
})
```

### 3. Event Naming

```go
// ✅ Use descriptive, action-oriented names
ctx.Emit("userLoggedIn", userData)
ctx.Emit("formSubmitted", formData)
ctx.Emit("itemSelected", itemId)

// ❌ Avoid vague names
ctx.Emit("update", data)
ctx.Emit("change", data)
ctx.Emit("event", data)
```

### 4. Template Simplicity

```go
// ✅ Keep templates focused on rendering
Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    return fmt.Sprintf("Count: %d", count.Get().(int))
})

// ❌ Avoid business logic in templates
Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    // ❌ Business logic belongs in Setup
    if count.Get().(int) > 100 {
        count.Set(0)  // ❌ Don't mutate in template!
    }
    return fmt.Sprintf("Count: %d", count.Get().(int))
})
```

### 5. Component Composition

```go
// ✅ Build small, focused components
func createButton() bubbly.Component { /* ... */ }
func createInput() bubbly.Component { /* ... */ }
func createForm() bubbly.Component {
    button, _ := createButton()
    input, _ := createInput()
    return bubbly.NewComponent("Form").
        Children(button, input).
        Build()
}

// ❌ Avoid monolithic components
func createEverything() bubbly.Component {
    // ❌ 500 lines of Setup logic
    // ❌ Handles everything in one component
}
```

### 6. Error Handling

```go
// ✅ Always check Build() errors
component, err := bubbly.NewComponent("MyComponent").
    Template(func(ctx bubbly.RenderContext) string {
        return "Hello"
    }).
    Build()
if err != nil {
    log.Fatal(err)
}

// ❌ Ignoring errors
component, _ := bubbly.NewComponent("MyComponent").Build()
```

---

## Error Tracking & Observability

BubblyUI includes **Phase 8: Error Tracking & Observability** features for production applications.

### Console Reporter (Development)

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"

func main() {
    // Enable console error reporting
    reporter := observability.NewConsoleReporter(true)  // verbose mode
    observability.SetErrorReporter(reporter)
    
    // Your app
    component, _ := createApp()
    tea.NewProgram(model{app: component}).Run()
}
```

### Sentry Reporter (Production)

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"

func main() {
    // Configure Sentry
    reporter, err := observability.NewSentryReporter(
        "https://your-dsn@sentry.io/project",
        observability.WithEnvironment("production"),
        observability.WithRelease("v1.0.0"),
        observability.WithBeforeSend(func(event *sentry.Event) *sentry.Event {
            // Filter sensitive data
            return event
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    observability.SetErrorReporter(reporter)
    defer reporter.Flush(5 * time.Second)
    
    // Your app
    component, _ := createApp()
    tea.NewProgram(model{app: component}).Run()
}
```

### Breadcrumbs

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"

Setup(func(ctx *bubbly.Context) {
    ctx.On("userAction", func(e *bubbly.Event) {
        // Record breadcrumb for debugging
        observability.RecordBreadcrumb(
            "user",
            "User clicked button",
            map[string]interface{}{
                "button_id": e.Data,
            },
        )
    })
})
```

### Custom Reporter

```go
type MyReporter struct{}

func (r *MyReporter) ReportPanic(err *bubbly.HandlerPanicError, ctx *observability.ErrorContext) {
    // Send to your logging service
    log.Printf("PANIC in %s: %v", ctx.ComponentName, err.PanicValue)
}

func (r *MyReporter) ReportError(err error, ctx *observability.ErrorContext) {
    // Send to your logging service
    log.Printf("ERROR in %s: %v", ctx.ComponentName, err)
}

func (r *MyReporter) Flush(timeout time.Duration) error {
    return nil
}

// Use custom reporter
observability.SetErrorReporter(&MyReporter{})
```

---

## Next Steps

1. **Explore Examples** - Check `cmd/examples/02-component-model/` for working examples
2. **Read API Docs** - Run `godoc -http=:6060` and browse to `pkg/bubbly`
3. **Join Community** - Share your migration experience
4. **Build Components** - Start with small, reusable components
5. **Add Error Tracking** - Configure observability for production

---

## Additional Resources

- **BubblyUI Repository:** https://github.com/newbpydev/bubblyui
- **Bubbletea Docs:** https://github.com/charmbracelet/bubbletea
- **Lipgloss Styling:** https://github.com/charmbracelet/lipgloss
- **Example Apps:** `cmd/examples/02-component-model/`
- **API Reference:** `pkg/bubbly/` (godoc)

---

**Happy Migrating! 🎉**

If you encounter issues or have questions, please open an issue on GitHub.
