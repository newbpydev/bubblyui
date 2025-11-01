# BubblyUI Directives Guide

Complete guide to using directives for declarative TUI development with type-safe templates.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [If Directive](#if-directive)
- [Show Directive](#show-directive)
- [ForEach Directive](#foreach-directive)
- [Bind Directive](#bind-directive)
- [On Directive](#on-directive)
- [Performance Guide](#performance-guide)
- [Troubleshooting](#troubleshooting)

---

## Introduction

### What are Directives?

Directives are **declarative functions** that enhance template rendering in BubblyUI. They provide clean, readable ways to handle common UI patterns like conditionals, loops, data binding, and events.

### Why Use Directives?

**Before (Imperative):**
```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    var output strings.Builder
    if len(items.Get()) > 0 {
        for i, item := range items.Get() {
            output.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
        }
    } else {
        output.WriteString("No items")
    }
    return output.String()
})
```

**After (Declarative):**
```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    return If(len(items.Get()) > 0,
        func() string {
            return ForEach(items.Get(), func(item string, i int) string {
                return fmt.Sprintf("%d. %s\n", i+1, item)
            }).Render()
        },
    ).Else(func() string {
        return "No items"
    }).Render()
})
```

### Key Benefits

✅ **Type-Safe**: Go generics provide compile-time type checking  
✅ **Declarative**: Intent is clear and self-documenting  
✅ **Composable**: Mix and nest directives naturally  
✅ **Performant**: Optimized for terminal rendering (see benchmarks)  
✅ **TUI-Specific**: Designed for terminal output, not web

---

## Quick Start

### Installation

Directives are part of the core BubblyUI package:

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
```

### Your First Directive

```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

func main() {
    app := bubbly.NewComponent(
        bubbly.Setup(func(ctx *bubbly.Context) {
            showMessage := ctx.Ref(true)
            ctx.Expose("showMessage", showMessage)
        }),
        bubbly.Template(func(ctx bubbly.RenderContext) string {
            show := ctx.Get("showMessage").(*bubbly.Ref[bool])
            
            return directives.If(show.Get(),
                func() string { return "Hello, BubblyUI!" },
            ).Render()
        }),
    )
    
    // Integrate with Bubbletea...
}
```

---

## If Directive

Conditional rendering that removes content from output when condition is false.

### Basic Usage

#### Example 1: Simple Condition

```go
directives.If(isLoggedIn,
    func() string { return "Welcome back!" },
).Render()
```

**Output when true:** `"Welcome back!"`  
**Output when false:** `""` (empty string)

#### Example 2: If-Else

```go
directives.If(hasItems,
    func() string { return "Show items list" },
).Else(func() string {
    return "No items found"
}).Render()
```

### Advanced Usage

#### Example 3: ElseIf Chains

```go
directives.If(score >= 90,
    func() string { return "Grade: A" },
).ElseIf(score >= 80, func() string {
    return "Grade: B"
}).ElseIf(score >= 70, func() string {
    return "Grade: C"
}).Else(func() string {
    return "Grade: F"
}).Render()
```

#### Example 4: Nested Conditions

```go
directives.If(user != nil,
    func() string {
        return directives.If(user.IsAdmin,
            func() string { return "[ADMIN] " + user.Name },
        ).Else(func() string {
            return user.Name
        }).Render()
    },
).Render()
```

#### Example 5: With Reactive State

```go
Setup(func(ctx *Context) {
    isLoading := ctx.Ref(false)
    ctx.Expose("isLoading", isLoading)
})

Template(func(ctx RenderContext) string {
    loading := ctx.Get("isLoading").(*Ref[bool])
    
    return directives.If(loading.Get(),
        func() string { return "⏳ Loading..." },
    ).Else(func() string {
        return "✓ Ready"
    }).Render()
})
```

### Performance

- **Simple:** 2.8ns, 0 allocs (when false)
- **True branch:** 8.2ns, 0 allocs
- **ElseIf chain:** 223ns, 3 allocs (complexity-dependent)

---

## Show Directive

Visibility toggle that keeps content in output with `[Hidden]` marker (unlike If).

### Basic Usage

#### Example 6: Basic Visibility

```go
directives.Show(isVisible,
    func() string { return "Visible content" },
).Render()
```

**When true:** `"Visible content"`  
**When false:** `""` (empty by default)

#### Example 7: With Transition

```go
directives.Show(isVisible,
    func() string { return "Fading content" },
).WithTransition().Render()
```

**When true:** `"Fading content"`  
**When false:** `"[Hidden: Fading content]"`

### Advanced Usage

#### Example 8: Conditional Panels

```go
Setup(func(ctx *Context) {
    activePanel := ctx.Ref("main") // "main", "settings", "help"
    ctx.Expose("panel", activePanel)
})

Template(func(ctx RenderContext) string {
    panel := ctx.Get("panel").(*Ref[string])
    
    main := directives.Show(panel.Get() == "main", func() string {
        return "Main Panel Content"
    }).Render()
    
    settings := directives.Show(panel.Get() == "settings", func() string {
        return "Settings Panel"
    }).Render()
    
    help := directives.Show(panel.Get() == "help", func() string {
        return "Help Content"
    }).Render()
    
    return main + settings + help
})
```

### Performance

- **Visible:** 7.1ns, 0 allocs
- **Hidden (no transition):** 2.2ns, 0 allocs
- **Hidden (with transition):** 148ns, 2 allocs (string formatting)

---

## ForEach Directive

Type-safe list iteration with Go generics.

### Basic Usage

#### Example 9: Simple List

```go
items := []string{"Apple", "Banana", "Cherry"}

directives.ForEach(items, func(item string, index int) string {
    return fmt.Sprintf("%d. %s\n", index+1, item)
}).Render()
```

**Output:**
```
1. Apple
2. Banana
3. Cherry
```

#### Example 10: Struct Iteration

```go
type Task struct {
    Title     string
    Completed bool
}

tasks := []Task{
    {Title: "Write docs", Completed: true},
    {Title: "Add tests", Completed: false},
}

directives.ForEach(tasks, func(task Task, i int) string {
    status := "[ ]"
    if task.Completed {
        status = "[X]"
    }
    return fmt.Sprintf("%s %s\n", status, task.Title)
}).Render()
```

### Advanced Usage

#### Example 11: Nested ForEach

```go
type Category struct {
    Name  string
    Items []string
}

categories := []Category{
    {Name: "Fruits", Items: []string{"Apple", "Banana"}},
    {Name: "Vegetables", Items: []string{"Carrot", "Broccoli"}},
}

directives.ForEach(categories, func(cat Category, i int) string {
    header := fmt.Sprintf("## %s\n", cat.Name)
    items := directives.ForEach(cat.Items, func(item string, j int) string {
        return fmt.Sprintf("  - %s\n", item)
    }).Render()
    return header + items
}).Render()
```

#### Example 12: Filtering with ForEach

```go
allTasks := []Task{/* ... */}

activeTasks := []Task{}
for _, task := range allTasks {
    if !task.Completed {
        activeTasks = append(activeTasks, task)
    }
}

directives.ForEach(activeTasks, func(task Task, i int) string {
    return renderTask(task)
}).Render()
```

#### Example 13: Empty List Handling

```go
directives.If(len(items) > 0,
    func() string {
        return directives.ForEach(items, renderItem).Render()
    },
).Else(func() string {
    return "No items to display"
}).Render()
```

### Performance

- **10 items:** 1.6μs, 12 allocs
- **100 items:** 17μs, 102 allocs
- **1000 items:** 189μs, 2490 allocs
- **Scales linearly** with item count

---

## Bind Directive

Type-safe two-way data binding for input elements.

### Basic Usage

#### Example 14: String Binding

```go
Setup(func(ctx *Context) {
    username := ctx.Ref("")
    ctx.Expose("username", username)
    ctx.Expose("usernameInput", directives.Bind(username))
})

Template(func(ctx RenderContext) string {
    input := ctx.Get("usernameInput").(*directives.BindDirective[string])
    username := ctx.Get("username").(*Ref[string])
    
    return fmt.Sprintf("Username: %s\nCurrent: %s",
        input.Render(),
        username.Get())
})
```

#### Example 15: Integer Binding

```go
Setup(func(ctx *Context) {
    age := ctx.Ref(25)
    ctx.Expose("age", age)
    ctx.Expose("ageInput", directives.Bind(age))
})
```

#### Example 16: Boolean Checkbox

```go
Setup(func(ctx *Context) {
    agreed := ctx.Ref(false)
    ctx.Expose("agreed", agreed)
    ctx.Expose("agreeCheckbox", directives.BindCheckbox(agreed))
})

Template(func(ctx RenderContext) string {
    checkbox := ctx.Get("agreeCheckbox").(*directives.BindDirective[bool])
    return checkbox.Render() + " I agree to terms"
})
```

**Output when true:** `"[Checkbox: [X]] I agree to terms"`  
**Output when false:** `"[Checkbox: [ ]] I agree to terms"`

### Advanced Usage

#### Example 17: Select Dropdown

```go
Setup(func(ctx *Context) {
    size := ctx.Ref("Medium")
    sizes := []string{"Small", "Medium", "Large", "XL"}
    ctx.Expose("size", size)
    ctx.Expose("sizeSelect", directives.BindSelect(size, sizes))
})

Template(func(ctx RenderContext) string {
    select := ctx.Get("sizeSelect").(*directives.SelectBindDirective[string])
    return "Select size:\n" + select.Render()
})
```

**Output:**
```
Select size:
[Select:
  Small
> Medium
  Large
  XL
]
```

#### Example 18: Custom Type Binding

```go
type Priority int

const (
    Low Priority = iota
    Medium
    High
)

Setup(func(ctx *Context) {
    priority := ctx.Ref(Medium)
    priorities := []Priority{Low, Medium, High}
    ctx.Expose("priority", priority)
    ctx.Expose("prioritySelect", directives.BindSelect(priority, priorities))
})
```

### Performance

- **BindCheckbox:** 15.7ns, **0 allocs** ✅
- **String:** 190ns, 3 allocs
- **Int:** 135ns, 1 alloc
- **Select (3 options):** 572ns, 9 allocs
- **Select (50 options):** 5.2μs, 42 allocs

---

## On Directive

Declarative event handling with modifiers.

### Basic Usage

#### Example 19: Simple Event

```go
directives.On("click", func(data interface{}) {
    fmt.Println("Button clicked!")
}).Render("Click Me")
```

**Output:** `"[Event:click]Click Me"`

#### Example 20: Keyboard Event

```go
directives.On("keypress", func(data interface{}) {
    if key, ok := data.(string); ok {
        handleKeyPress(key)
    }
}).Render(content)
```

### Event Modifiers

#### Example 21: Prevent Default

```go
directives.On("submit", handleSubmit).
    PreventDefault().
    Render("Submit Form")
```

**Output:** `"[Event:submit:prevent]Submit Form"`

#### Example 22: Stop Propagation

```go
directives.On("click", handleClick).
    StopPropagation().
    Render("Inner Button")
```

**Output:** `"[Event:click:stop]Inner Button"`

#### Example 23: All Modifiers

```go
directives.On("submit", handleSubmit).
    PreventDefault().
    StopPropagation().
    Once().
    Render("One-Time Submit")
```

**Output:** `"[Event:submit:prevent:stop:once]One-Time Submit"`

### Advanced Usage

#### Example 24: Multiple Events

```go
content := "Interactive Element"
content = directives.On("click", handleClick).Render(content)
content = directives.On("hover", handleHover).Render(content)
```

**Output:** `"[Event:hover][Event:click]Interactive Element"`

#### Example 25: Event with ForEach

```go
directives.ForEach(buttons, func(btn Button, i int) string {
    return directives.On("click", func(data interface{}) {
        handleButtonClick(i)
    }).Render(btn.Label)
}).Render()
```

### Performance

- **Simple:** 48.7ns, 1 alloc (meets <80ns target) ✅
- **With modifiers:** 52-60ns, 1 alloc
- **Multiple events:** 175ns, 3 allocs

---

## Performance Guide

### Benchmark Results

All directives are optimized for terminal rendering:

| Directive | Performance | Status |
|-----------|-------------|--------|
| If (simple) | 2-8ns, 0 allocs | ✅ Exceeds target by 20x |
| Show (visible) | 7ns, 0 allocs | ✅ Exceeds target by 7x |
| ForEach (100 items) | 17μs, 102 allocs | ✅ Exceeds target by 50x |
| On (simple) | 48.7ns, 1 alloc | ✅ Meets <80ns target |
| BindCheckbox | 15.7ns, 0 allocs | ✅ Zero allocations |

### Optimization Tips

#### 1. Use BindCheckbox for Booleans

```go
// ✅ Optimized (15.7ns, 0 allocs)
directives.BindCheckbox(boolRef)

// ❌ Slower (122ns, 1 alloc)
directives.Bind(boolRef)
```

#### 2. Avoid fmt.Sprintf in Hot Paths

The directive implementations use `strings.Builder` internally for optimal performance. If you're building custom rendering:

```go
// ✅ Fast (uses strings.Builder)
var builder strings.Builder
builder.Grow(estimatedSize)
builder.WriteString("prefix")
builder.WriteString(value)
return builder.String()

// ❌ Slow (multiple allocations)
return fmt.Sprintf("prefix%s", value)
```

#### 3. Pre-filter Large Lists

```go
// ✅ Filter once, iterate once
filtered := filterTasks(allTasks)
directives.ForEach(filtered, renderTask).Render()

// ❌ Filter on every render
directives.ForEach(allTasks, func(task Task, i int) string {
    if shouldShow(task) {
        return renderTask(task)
    }
    return ""
}).Render()
```

---

## Troubleshooting

### Common Issues

#### Issue 1: Type Mismatch with Generics

**Problem:**
```go
// ❌ Compiler error: type mismatch
items := ctx.Get("items").(*Ref[string]) // Wrong: items is []string
```

**Solution:**
```go
// ✅ Correct type assertion
items := ctx.Get("items").(*Ref[[]string])
```

#### Issue 2: Nil Pointer Panic

**Problem:**
```go
// ❌ Panics if "user" not exposed
user := ctx.Get("user").(*Ref[*User])
name := user.Get().Name // Panic if user is nil
```

**Solution:**
```go
// ✅ Check for nil
if user.Get() != nil {
    name := user.Get().Name
}

// Or use If directive
directives.If(user.Get() != nil, func() string {
    return user.Get().Name
}).Render()
```

#### Issue 3: ForEach with Empty Slice

**Problem:** Need to show message when list is empty

**Solution:**
```go
// ✅ Use If directive
directives.If(len(items) > 0,
    func() string {
        return directives.ForEach(items, render).Render()
    },
).Else(func() string {
    return "No items"
}).Render()
```

#### Issue 4: Performance Issues with Large Lists

**Problem:** Slow rendering with 10,000+ items

**Solution:**
```go
// ✅ Implement pagination
const pageSize = 100
start := currentPage * pageSize
end := start + pageSize
if end > len(allItems) {
    end = len(allItems)
}
visibleItems := allItems[start:end]

directives.ForEach(visibleItems, render).Render()
```

#### Issue 5: Event Handler Not Firing

**Problem:** On directive renders but handler doesn't execute

**Solution:** Ensure component event system is integrated properly. The directive only adds event markers; the component must process them:

```go
Setup(func(ctx *Context) {
    ctx.On("click", func(data interface{}) {
        // Handler registered with component
    })
})
```

---

## Next Steps

- **[Directive Patterns](directive-patterns.md)**: Best practices and advanced patterns
- **[Performance Optimization](performance-optimization.md)**: Deep dive into optimization techniques
- **[Component Guide](../components.md)**: Integrate directives with components
- **[Reactive System](reactive-dependencies.md)**: Use with Ref[T] and Computed[T]

---

## Related Documentation

- **API Reference**: See godoc comments in `pkg/bubbly/directives/*.go`
- **Specs**: `specs/05-directives/` for complete feature specification
- **Examples**: `cmd/examples/` for working applications

---

**Note:** BubblyUI is a TUI (Terminal User Interface) framework for Go, not a web framework. All directives render to terminal output using ANSI escape codes and Lipgloss styling, not HTML/CSS.
