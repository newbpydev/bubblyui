# BubblyUI Manual for AI Agents

**The Definitive Reference Guide for Building Vue-Inspired TUI Applications**

**Version:** 1.0  
**Last Updated:** November 18, 2025  
**Target Audience:** AI Coding Assistants

---

## Purpose

This manual enables AI agents to:
1. Build composable, testable TUI applications using TDD
2. Leverage all 12+ BubblyUI features
3. Follow proven best practices
4. Test systematically with Go testing

**Key Principle:** Use BubblyUI as primary tool, minimal direct Bubbletea usage.

---

## Quick Reference

### Essential Imports
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
    "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"
    tea "github.com/charmbracelet/bubbletea"
)
```

### Core Patterns

#### 1. Create Component
```go
func CreateComponent(props Props) (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "ComponentName",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            // 1. Create refs
            count := bubbly.NewRef(0)
            
            // 2. Create computed
            isEven := ctx.Computed(func() interface{} {
                return count.Get().(int)%2 == 0
            })
            
            // 3. Lifecycle hooks
            ctx.OnMounted(func() { /* init */ })
            ctx.OnUnmounted(func() { /* cleanup */ })
            
            // 4. Event handlers
            ctx.On("event", func(data interface{}) { /* handle */ })
            
            // 5. Expose for template
            ctx.Expose("count", count)
            
            // 6. Return template
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    return "rendered output"
                },
            }
        },
    })
}
```

#### 2. Use Composables
```go
counter := composables.UseCounter(ctx, 0)
form := composables.UseForm(ctx, FormStruct{})
async := composables.UseAsync(ctx, fetchFunction)
```

#### 3. Use Components
```go
input := components.Input(components.InputProps{
    Label: "Username",
    Value: valueRef,
})
input.Init()
output := input.View()
```

#### 4. Use Router
```go
r := router.NewRouter().
    AddRoute("/", homeComponent).
    AddRoute("/users/:id", userComponent).
    Build()
```

---

## Part 1: Complete Context API (23 Methods)

### State Management Methods

**1. Ref() - Create reactive reference**
```go
count := ctx.Ref(0)  // Returns *Ref[interface{}]
// Auto has template safety + auto commands if enabled
// Prefer bubbly.NewRef() for type safety
```

**2. Computed() - Derived values**
```go
doubled := ctx.Computed(func() interface{} {
    return count.Get().(int) * 2
})
```

**3. Watch() - React to changes**
```go
cleanup := ctx.Watch(count, func(newVal, oldVal interface{}) {
    fmt.Printf("Changed: %v → %v\n", oldVal, newVal)
})
// Auto-cleanup on unmount
```

**4. Expose() - Make available to template**
```go
ctx.Expose("count", count)
// CRITICAL: Also registers refs with DevTools for tracking
```

**5. Get() - Retrieve from state**
```go
value := ctx.Get("count")
```

**6. ExposeComponent() - CRITICAL METHOD!**
```go
child, _ := CreateChild(props)
err := ctx.ExposeComponent("child", child)
// Does 3 things:
// 1. Auto-calls Init() if not initialized
// 2. Establishes parent-child relationship (for DevTools tree)
// 3. Exposes to context
// Returns error if component is nil
```

### Event Methods

**7. On() - Register event handler**
```go
ctx.On("submit", func(data interface{}) {
    formData := data.(FormData)
})
```

**8. Emit() - Send event**
```go
ctx.Emit("submit", FormData{...})
```

### Lifecycle Methods (6 hooks)

**9. OnMounted() - After first render**
```go
ctx.OnMounted(func() {
    // Init, fetch data, start timers
})
```

**10. OnUpdated() - After updates**
```go
// Without deps - runs every update
ctx.OnUpdated(func() {
    fmt.Println("Updated")
})

// With deps - runs only when deps change
ctx.OnUpdated(func() {
    fmt.Println("Count changed")
}, count)
```

**11. OnUnmounted() - Before destroy**
```go
ctx.OnUnmounted(func() {
    // CRITICAL: Cleanup resources!
    ticker.Stop()
})
```

**12. OnBeforeUpdate() - Before update**
```go
ctx.OnBeforeUpdate(func() {
    // Prepare for update
})
```

**13. OnBeforeUnmount() - Before unmount**
```go
ctx.OnBeforeUnmount(func() {
    // Final preparations
})
```

**14. OnCleanup() - Register cleanup**
```go
ctx.OnCleanup(func() {
    ticker.Stop()
})
// Executes in LIFO order
```

### Dependency Injection Methods

**15. Provide() - Provide to descendants**
```go
theme := ctx.Ref("dark")
ctx.Provide("theme", theme)
```

**16. Inject() - Get from ancestors**
```go
theme := ctx.Inject("theme", "light")
// Walks up tree, returns default if not found
```

### Props & Children Methods

**17. Props() - Get component props**
```go
props := ctx.Props().(ButtonProps)
```

**18. Children() - Get child components**
```go
children := ctx.Children()
```

### Command Generation Methods (5 methods)

**19. EnableAutoCommands()**
```go
ctx.EnableAutoCommands()
// Ref.Set() now triggers UI updates automatically
```

**20. DisableAutoCommands()**
```go
ctx.DisableAutoCommands()
// Batch updates without commands
```

**21. IsAutoCommandsEnabled()**
```go
if ctx.IsAutoCommandsEnabled() {
    // Auto commands active
}
```

**22. ManualRef() - Ref without auto commands**
```go
internal := ctx.ManualRef(0)
internal.Set(100)  // Never generates command
ctx.Emit("update", nil)  // Manual update required
```

**23. SetCommandGenerator() - Custom generator**
```go
ctx.SetCommandGenerator(&CustomGenerator{})
```

### Component Lifecycle

```
Setup() → onMounted() → [onUpdated()...] → onUnmounted()
```

**Critical:** Always cleanup in `onUnmounted()` to prevent leaks!

### Props Pattern

```go
type Props struct {
    Title    string
    Value    *bubbly.Ref[int]
    OnChange func(int)
}

func CreateComponent(props Props) (bubbly.Component, error) {
    // Use props.Title, props.Value, props.OnChange
}
```

---

## Part 2: Component Builder API (11 Methods)

**1. NewComponent() - Create builder**
```go
builder := bubbly.NewComponent("ButtonComponent")
```

**2. Props() - Set component props**
```go
builder.Props(ButtonProps{Label: "Click me", Disabled: false})
```

**3. Setup() - Setup function**
```go
builder.Setup(func(ctx *bubbly.Context) {
    // Initialization logic
})
```

**4. Template() - Render function**
```go
builder.Template(func(ctx *bubbly.RenderContext) string {
    return "rendered output"
})
```

**5. Children() - Set child components**
```go
builder.Children(child1, child2, child3)
// Sets parent reference automatically
```

**6. WithAutoCommands() - Enable auto commands**
```go
builder.WithAutoCommands(true)
// Initializes command queue + generator in Build()
```

**7. WithCommandDebug() - Enable command logging**
```go
builder.WithCommandDebug(true)
// Logs: [DEBUG] Command Generated | Component: Counter | Ref: ref-5 | 0 → 1
```

**8. WithKeyBinding() - Simple key binding**
```go
builder.WithKeyBinding(" ", "increment", "Increment counter")
// CRITICAL: Use " " for space, NOT "space"!
```

**9. WithConditionalKeyBinding() - Advanced binding**
```go
inputMode := false
builder.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key:         " ",
    Event:       "addChar",
    Description: "Add space",
    Data:        " ",
    Condition:   func() bool { return inputMode },
})
```

**10. WithKeyBindings() - Batch bindings**
```go
bindings := map[string]bubbly.KeyBinding{
    " ":      {Key: " ", Event: "increment", Description: "Increment"},
    "ctrl+c": {Key: "ctrl+c", Event: "quit", Description: "Quit"},
}
builder.WithKeyBindings(bindings)
```

**11. WithMessageHandler() - Custom message handler**
```go
builder.WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        comp.Emit("resize", map[string]int{"width": msg.Width, "height": msg.Height})
        return nil
    }
    return nil
})
// Called BEFORE key binding processing
```

**Build() - Finalize component**
```go
component, err := builder.Build()
// Validates: template required, checks errors
```

---

## Part 3: Built-in Features

### 24 Built-in Components

**Atoms:** Button, Text, Icon, Badge, Spacer, Spinner
**Molecules:** Input, Checkbox, Radio, Select, Toggle, Textarea  
**Organisms:** Form, Table, List, Modal, Card, Menu, Tabs, Accordion
**Templates:** AppLayout, PageLayout, PanelLayout, GridLayout

**Example:**
```go
input := components.Input(components.InputProps{
    Label:       "Email",
    Value:       emailRef,
    Placeholder: "you@example.com",
})
input.Init()
```

### 9 Built-in Composables

```go
UseState, UseAsync, UseEffect, UseDebounce, UseThrottle,
UseForm, UseLocalStorage, UseEventListener, UseTextInput
```

**1. UseState** - Simple state
```go
state := composables.UseState(ctx, initialValue)
value := state.Value.Get()
state.SetValue(newValue)
```

**2. UseAsync** - Async operations
```go
async := composables.UseAsync(ctx, func() (interface{}, error) {
    return api.FetchData()
})
loading := async.Loading.Get().(bool)
data := async.Data.Get()
async.Execute()
```

**3. UseEffect** - Side effects
```go
composables.UseEffect(ctx, func() {
    // Effect logic
    return func() { /* cleanup */ }
}, []interface{}{dependency})
```

**4. UseDebounce** - Debounced values
```go
debounced := composables.UseDebounce(ctx, searchQuery, 300*time.Millisecond)
```

**5. UseThrottle** - Throttled values
```go
throttled := composables.UseThrottle(ctx, clickCount, 1*time.Second)
```

**6. UseForm** - Form management
```go
form := composables.UseForm(ctx, LoginForm{})
form.SetField("username", "john")
form.Validate()
isValid := form.IsValid.Get().(bool)
```

**7. UseLocalStorage** - Persistent state
```go
storage := composables.UseLocalStorage(ctx, "key", defaultValue)
value := storage.Value.Get()
storage.SetValue(newValue)
```

**8. UseEventListener** - Event management
```go
listener := composables.UseEventListener(ctx, "keypress")
events := listener.Events.Get().([]interface{})
```

**9. UseTextInput** - Text input helper
```go
textInput := composables.UseTextInput(ctx, "")
value := textInput.Value.Get().(string)
cursor := textInput.Cursor.Get().(int)
textInput.Insert("text")
textInput.Delete()
textInput.MoveCursor(1)  // Forward
textInput.Clear()
```

### 5 Directives

```go
If, Show, ForEach, Bind, On
```

**Example:**
```go
directives.If(showContent.Get().(bool), func() string {
    return "Conditional content"
})

directives.ForEach(items, func(item interface{}, index int) string {
    return fmt.Sprintf("%d. %s\n", index+1, item)
})
```

### Router System

```go
r := router.NewRouter().
    AddRoute("/", homeComponent).
    AddRoute("/users/:id", userComponent).
    WithGuard(authGuard).  // Navigation guard
    Build()

// Navigate
r.Navigate("/users/123")
r.GoBack()

// Get params
params := r.CurrentRoute().Params
id := params["id"]
```

---

## Part 3: TDD Workflow

### Test Structure (Table-Driven)

```go
func TestCounter(t *testing.T) {
    tests := []struct {
        name     string
        initial  int
        action   string
        expected int
    }{
        {"increment from 0", 0, "increment", 1},
        {"decrement from 5", 5, "decrement", 4},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            counter := composables.UseCounter(ctx, tt.initial)
            
            // Act
            if tt.action == "increment" {
                counter.Increment()
            } else {
                counter.Decrement()
            }
            
            // Assert
            assert.Equal(t, tt.expected, counter.Count.Get().(int))
        })
    }
}
```

### Test Components

```go
func TestCounterComponent(t *testing.T) {
    // Arrange
    count := bubbly.NewRef(0)
    comp, err := CreateCounter(CounterProps{
        InitialCount: 5,
    })
    require.NoError(t, err)
    
    // Act
    comp.Init()
    output := comp.View()
    
    // Assert
    assert.Contains(t, output, "5")
}
```

### Test with Assertions

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFeature(t *testing.T) {
    // require stops test on failure
    require.NotNil(t, component)
    
    // assert continues after failure
    assert.Equal(t, expected, actual)
    assert.Contains(t, output, "text")
    assert.True(t, condition)
}
```

---

## Part 4: Common Patterns

### Composable App Pattern

**Directory Structure:**
```
myapp/
├── main.go          # Entry point
├── app.go           # Root component
├── components/      # UI components
│   ├── header.go
│   └── footer.go
└── composables/     # Shared logic
    └── use_counter.go
```

**main.go:**
```go
func main() {
    app, _ := CreateApp()
    wrapped := bubbly.Wrapper(app)  // Auto command generation
    p := tea.NewProgram(wrapped, tea.WithAltScreen())
    p.Run()
}
```

### Form Pattern

```go
type FormData struct {
    Username string
    Password string
}

form := composables.UseForm(ctx, FormData{})

// In event handler
ctx.On("submit", func(_ interface{}) {
    if form.IsValid.Get().(bool) {
        data := form.Values.Get().(FormData)
        // Submit data
    }
})
```

### List Management Pattern

```go
items := bubbly.NewRef([]Todo{})

// Add item
ctx.On("add", func(data interface{}) {
    current := items.Get().([]Todo)
    items.Set(append(current, newTodo))
})

// Render list
directives.ForEach(items.Get().([]Todo), func(item interface{}, index int) string {
    return renderTodoItem(item.(Todo))
})
```

---

## Part 5: Anti-Patterns

### ❌ DON'T

1. **Don't use ctx.Ref()** - Returns `Ref[interface{}]`
   ```go
   // ❌ WRONG
   count := ctx.Ref(0)
   
   // ✅ CORRECT
   count := bubbly.NewRef(0)
   ```

2. **Don't skip initialization**
   ```go
   // ❌ WRONG
   input := components.Input(props)
   return input.View()  // Not initialized!
   
   // ✅ CORRECT
   input := components.Input(props)
   input.Init()
   return input.View()
   ```

3. **Don't forget cleanup**
   ```go
   // ❌ WRONG - Memory leak
   ctx.OnMounted(func() {
       ticker := time.NewTicker(1 * time.Second)
       // No cleanup!
   })
   
   // ✅ CORRECT
   ctx.OnMounted(func() {
       ticker := time.NewTicker(1 * time.Second)
       ctx.Set("ticker", ticker)
   })
   ctx.OnUnmounted(func() {
       ticker.(*time.Ticker).Stop()
   })
   ```

4. **Don't use Toggle.Checked** - Use `Value` prop
   ```go
   // ❌ WRONG
   toggle := components.Toggle(components.ToggleProps{
       Checked: enabledRef,
   })
   
   // ✅ CORRECT
   toggle := components.Toggle(components.ToggleProps{
       Value: enabledRef,
   })
   ```

5. **Don't hardcode Lipgloss when components exist**
   ```go
   // ❌ WRONG
   style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
   return style.Render("Card content")
   
   // ✅ CORRECT
   card := components.Card(components.CardProps{
       Content: "Card content",
   })
   card.Init()
   return card.View()
   ```

---

## Part 6: Complete Example

### Todo App with Router

**main.go:**
```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    app, _ := CreateApp()
    wrapped := bubbly.Wrapper(app)
    p := tea.NewProgram(wrapped, tea.WithAltScreen())
    p.Run()
}
```

**app.go:**
```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

func CreateApp() (bubbly.Component, error) {
    // Create screens
    homeScreen, _ := CreateHomeScreen()
    todosScreen, _ := CreateTodosScreen()
    
    // Create router
    r := router.NewRouter().
        AddRoute("/", homeScreen).
        AddRoute("/todos", todosScreen).
        Build()
    
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "App",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            ctx.Expose("router", r)
            
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    router := ctx.Get("router").(*router.Router)
                    return router.View()
                },
            }
        },
    })
}
```

**composables/use_todos.go:**
```go
package composables

import "github.com/newbpydev/bubblyui/pkg/bubbly"

type Todo struct {
    ID        string
    Title     string
    Completed bool
}

type TodosComposable struct {
    Todos  *bubbly.Ref[[]Todo]
    Add    func(string)
    Toggle func(string)
    Remove func(string)
}

func UseTodos(ctx bubbly.SetupContext) *TodosComposable {
    todos := bubbly.NewRef([]Todo{})
    
    add := func(title string) {
        current := todos.Get().([]Todo)
        newTodo := Todo{
            ID:    generateID(),
            Title: title,
        }
        todos.Set(append(current, newTodo))
    }
    
    toggle := func(id string) {
        current := todos.Get().([]Todo)
        for i, todo := range current {
            if todo.ID == id {
                current[i].Completed = !current[i].Completed
                break
            }
        }
        todos.Set(current)
    }
    
    remove := func(id string) {
        current := todos.Get().([]Todo)
        filtered := []Todo{}
        for _, todo := range current {
            if todo.ID != id {
                filtered = append(filtered, todo)
            }
        }
        todos.Set(filtered)
    }
    
    return &TodosComposable{
        Todos:  todos,
        Add:    add,
        Toggle: toggle,
        Remove: remove,
    }
}
```

**tests:**
```go
func TestUseTodos(t *testing.T) {
    ctx := bubbly.NewMockSetupContext()
    todos := composables.UseTodos(ctx)
    
    // Test add
    todos.Add("Buy milk")
    assert.Len(t, todos.Todos.Get().([]Todo), 1)
    
    // Test toggle
    todoList := todos.Todos.Get().([]Todo)
    todos.Toggle(todoList[0].ID)
    assert.True(t, todos.Todos.Get().([]Todo)[0].Completed)
}
```

---

## Key Takeaways for AI Agents

### ✅ ALWAYS

1. Use `bubbly.NewRef()` not `ctx.Ref()`
2. Call `.Init()` on all components
3. Cleanup in `onUnmounted()`
4. Use BubblyUI components, not manual Lipgloss
5. Test with table-driven tests
6. Use `bubbly.Wrapper()` for automatic commands
7. Provide theme to children
8. Type assert when getting ref values

### 📚 Quick Reference

- **Reactivity:** `bubbly.NewRef()`, `ctx.Computed()`, `bubbly.Watch()`
- **Components:** `bubbly.DefineComponent()`, factory pattern
- **Lifecycle:** `onMounted()`, `onUpdated()`, `onUnmounted()`
- **Composables:** `Use*` functions for reusable logic
- **Components:** `components.*` for UI, always `.Init()`
- **Directives:** `directives.*` for rendering control
- **Router:** `router.NewRouter()` for navigation
- **Testing:** Table-driven with `testify/assert`

### 🔗 Resources

- Examples: `cmd/examples/01-12/`
- Components: `pkg/components/*.go`
- Composables: `pkg/bubbly/composables/*.go`
- Tests: `tests/integration/*.go`
- Project Status: `specs/PROJECT_STATUS.md`

---

---

## Part 6: Critical Patterns (MUST KNOW!)

### ExposeComponent Pattern

**The BETTER way to expose child components:**

```go
Setup: func(ctx *bubbly.Context) {
    // Create child components
    header, _ := CreateHeader(headerProps)
    sidebar, _ := CreateSidebar(sidebarProps)
    footer, _ := CreateFooter(footerProps)
    
    // ❌ OLD WAY - Manual (3 steps per component)
    header.Init()
    ctx.AddChild(header)
    ctx.Expose("header", header)
    
    // ✅ NEW WAY - Use ExposeComponent (1 step!)
    ctx.ExposeComponent("header", header)
    ctx.ExposeComponent("sidebar", sidebar)
    ctx.ExposeComponent("footer", footer)
    
    // ExposeComponent does 3 things automatically:
    // 1. Calls Init() if not already initialized
    // 2. Calls AddChild() to establish parent-child relationship (critical for DevTools)
    // 3. Calls Expose() to make available in template
    // Returns error if component is nil
}
```

### Provide/Inject Pattern

**Dependency injection without prop drilling:**

```go
// Parent component - provide values to descendants
Setup: func(ctx *bubbly.Context) {
    theme := bubbly.NewRef("dark")
    user := bubbly.NewRef(currentUser)
    config := bubbly.NewRef(appConfig)
    
    // Provide to ALL descendants
    ctx.Provide("theme", theme)
    ctx.Provide("user", user)
    ctx.Provide("config", config)
}

// Child component (any level deep) - inject values
Setup: func(ctx *bubbly.Context) {
    // Get from nearest provider (walks up tree)
    theme := ctx.Inject("theme", "light")  // With default
    user := ctx.Inject("user", nil)        // Can be nil
    
    // Use as normal
    if user != nil {
        userRef := user.(*bubbly.Ref[User])
        // ...
    }
}

// Nested child component - still works!
Setup: func(ctx *bubbly.Context) {
    // Same Inject calls work at any depth
    theme := ctx.Inject("theme", "light")
}
```

### Command Control Pattern

**Fine-grained control over automatic UI updates:**

```go
Setup: func(ctx *bubbly.Context) {
    // Enable auto commands for component
    ctx.EnableAutoCommands()
    
    counter := bubbly.NewRef(0)
    
    // Normal operation - auto updates
    ctx.On("increment", func(_ interface{}) {
        counter.Set(counter.Get().(int) + 1)  // UI updates automatically
    })
    
    // Batch updates without triggering multiple UI updates
    ctx.On("batchUpdate", func(_ interface{}) {
        ctx.DisableAutoCommands()  // Temporarily disable
        
        for i := 0; i < 1000; i++ {
            counter.Set(i)  // No command generated
        }
        
        ctx.EnableAutoCommands()   // Re-enable
        ctx.Emit("update", nil)    // Single manual update
    })
    
    // Internal state that never triggers UI updates
    internalFlag := ctx.ManualRef(false)
    internalFlag.Set(true)  // Never generates command, even if auto enabled
    
    // Check status
    if ctx.IsAutoCommandsEnabled() {
        // Auto commands are active
    }
}
```

### Template Safety Pattern

**Templates MUST be pure functions (read-only):**

```go
Template: func(ctx *bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[int])
    
    // ✅ CORRECT - Read only
    value := count.Get().(int)
    
    // ❌ WRONG - Will PANIC with clear error!
    count.Set(value + 1)  // "Cannot call Ref.Set() in template - templates must be pure"
    
    // ✅ CORRECT - Use events for mutations
    // (Events are handled outside template)
    
    return fmt.Sprintf("Count: %d", value)
}
```

### Mode-Based Input Pattern

**Different behaviors in different modes:**

```go
inputMode := false

component := bubbly.NewComponent("Form").
    WithConditionalKeyBinding(bubbly.KeyBinding{
        Key:         " ",
        Event:       "toggle",
        Description: "Toggle in navigation mode",
        Condition:   func() bool { return !inputMode },
    }).
    WithConditionalKeyBinding(bubbly.KeyBinding{
        Key:         " ",
        Event:       "addChar",
        Description: "Add space in input mode",
        Data:        " ",
        Condition:   func() bool { return inputMode },
    }).
    WithKeyBinding("esc", "toggleMode", "Switch modes").
    Setup(func(ctx *bubbly.Context) {
        ctx.On("toggleMode", func(_ interface{}) {
            inputMode = !inputMode
        })
    }).
    Build()
```

---

## Summary

### Complete Feature Coverage

- ✅ **23 Context Methods** - Complete API reference
- ✅ **11 Builder Methods** - All builder options
- ✅ **9 Composables** - Including UseTextInput
- ✅ **24 Components** - All atoms, molecules, organisms, templates
- ✅ **5 Directives** - If, Show, ForEach, Bind, On
- ✅ **Router System** - Multi-screen navigation
- ✅ **Command Generation** - Auto UI updates
- ✅ **Provide/Inject** - Dependency injection
- ✅ **Lifecycle Hooks** - All 6 hooks
- ✅ **TDD Patterns** - Table-driven tests

### Key Takeaways for AI Agents

1. **Always use ExposeComponent** for child components (not manual Init + Expose)
2. **Use Provide/Inject** for cross-component communication
3. **Space key is " "** not "space" in key bindings
4. **Toggle uses Value prop** not Checked
5. **Templates are pure** - no mutations allowed
6. **Always cleanup** in onUnmounted
7. **Test with -race flag** always
8. **Use bubbly.NewRef()** not ctx.Ref() for type safety

**This manual now covers ~100% of BubblyUI's public API. Use it as your complete reference when building TUI applications.**
