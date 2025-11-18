# BubblyUI Directives

**Package Path:** `github.com/newbpydev/bubblyui/pkg/bubbly/directives`  
**Version:** 3.0  
**Purpose:** Vue-inspired template directives for conditional rendering, loops, binding, and event handling

## Overview

Directives provide declarative template enhancements for BubblyUI applications. Five core directives: If/Show, ForEach, Bind, On.

## Quick Start

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"

// Conditional rendering
directives.If(isLoggedIn, func() string {
    return "Welcome!"
}).Else(func() string {
    return "Please login"
}).Render()

// List rendering
items := []string{"Apple", "Banana", "Cherry"}
directives.ForEach(items, func(item string, i int) string {
    return fmt.Sprintf("%d. %s\n", i+1, item)
}).Render()

// Event binding
directives.On("click", handleClick).Render("Click me")
```

## Directives

### 1. If - Conditional Rendering

```go
func If(condition bool, then func() string) *IfDirective

// Basic usage
result := directives.If(showHelp, func() string {
    return "Help text"
}).Render()

// With else
directives.If(isAuthenticated, func() string {
    return "Dashboard"
}).Else(func() string {
    return "Login"
}).Render()

// With elseif
directives.If(status == "loading", func() string {
    return "Loading..."
}).ElseIf(status == "error", func() string {
    return "Error!"
}).Else(func() string {
    return "Success"
}).Render()
```

**Performance:** 2-16ns per evaluation

### 2. Show - Conditional Visibility

```go
func Show(condition bool, content string) string

// Unlike If, Show keeps element in DOM but hides it
result := directives.Show(isVisible, "Secret content")
```

**Performance:** Less than 1ns

### 3. ForEach - List Rendering

```go
func ForEach[T any](items []T, render func(T, int) string) *ForEachDirective[T]

// Render list
todos := []Todo{{Title: "Task 1"}, {Title: "Task 2"}}
output := directives.ForEach(todos, func(todo Todo, i int) string {
    status := "[ ]"
    if todo.Done {
        status = "[✓]"
    }
    return status + " " + todo.Title + "\n"
}).Render()

// With index
numbers := []int{1, 2, 3, 4, 5}
result := directives.ForEach(numbers, func(num int, idx int) string {
    return fmt.Sprintf("%d: %d\n", idx, num * 2)
}).Render()
```

**Performance:** 1.6-189μs for 10-1,000 items

### 4. Bind - Two-Way Data Binding

```go
func Bind[T any](ref *bubbly.Ref[T]) *BindDirective[T]
func BindCheckbox(ref *bubbly.Ref[bool]) *BindDirective[bool]
func BindSelect[T any](ref *bubbly.Ref[T], options []T) *SelectBindDirective[T]

// Text input
username := bubbly.NewRef("")
directives.Bind(username).Render()

// Checkbox
isChecked := bubbly.NewRef(false)
directives.BindCheckbox(isChecked).Render()

// Select
color := bubbly.NewRef("red")
colors := []string{"red", "green", "blue"}
directives.BindSelect(color, colors).Render()
```

**Performance:** 15-263ns (BindCheckbox: 0 allocations)

### 5. On - Event Handling

```go
func On(event string, handler func(interface{})) *OnDirective

// Basic event handler
directives.On("click", func(data interface{}) {
    fmt.Println("Clicked!")
}).Render("Click me")

// With prevent default
directives.On("submit", func(data interface{}) {
    // Handle submit
}).PreventDefault().Render("Submit")
```

**Performance:** 48-77ns

## Composition Example

```go
// Complex UI with multiple directives
content := directives.If(len(items) > 0,
    func() string {
        return directives.ForEach(items, func(item Item, i int) string {
            return directives.On("click", func(e interface{}) {
                selectItem(i)
            }).Render(
                directives.Show(item.Visible, func() string {
                    return fmt.Sprintf("- %s\n", item.Name)
                }).Else(func() string {
                    return "(archived)\n"
                }).Render(),
            )
        }).Render()
    },
).Else(func() string {
    return "No items found\n"
}).Render()
```

## Performance

All directives optimized for terminal rendering:

```
If:     2-16ns     (5-20x faster than target)
Show:   <1ns       (instant)
ForEach:1.6-189μs  (10-1000 items)
On:     48-77ns    (meets target)
Bind:   15-263ns   (BindCheckbox: 0 allocations)
```

## Best Practices

✓ **Use If for conditional rendering**
✓ **Use ForEach instead of manual loops**
✓ **Bind to refs for reactive updates**
✓ **Compose directives for complex UI**
✗ **Avoid nested If chains (use ElseIf)**
✗ **Don't create closures in hot paths**

## Integration

Works seamlessly with other packages:

```go
// With components
input := components.Input(components.InputProps{
    Value: directives.Bind(username),  // Two-way binding
})

// With composables
search := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)
results := directives.ForEach(searchResults, renderItem)
```

## API Reference

See [Full Directives API](docs/api/directives.md) for complete signatures and advanced usage.

**Status:** 5,805 LOC | 8 files | 5 directives