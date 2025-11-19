# BubblyUI Component Patterns Guide

**Critical Reference for Understanding Component Architecture**

**Version:** 3.0  
**Last Updated:** November 19, 2025  
**Status:** VERIFIED - Based on Real Production Issues  
**Target Audience:** All BubblyUI Developers

---

## 🚨 CRITICAL: Read This First

**Failure to follow these patterns WILL cause your application to crash.**

This guide was created after discovering a critical bug where using `ExposeComponent` with molecule components from `pkg/components` caused application crashes due to conflicting update mechanisms.

---

## The Two Types of Components

### 1. Composable Components (Custom App Components)

**Definition:** Components you create for your application that compose together to form the UI structure.

**Characteristics:**
- Defined with `bubbly.NewComponent()`
- Have `Setup` and `Template` functions
- Manage their own state with refs and computed values
- May have child components
- Participate in parent-child lifecycle
- Located in your app's `components/` folder

**Pattern:** `ExposeComponent` establishes parent-child relationship

**Example:**
```go
// File: components/counter_display.go
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
    return bubbly.NewComponent("CounterDisplay").
        Setup(func(ctx *bubbly.Context) {
            ctx.Expose("count", props.Count)
            ctx.Expose("isEven", props.IsEven)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[int]).Get()
            isEven := ctx.Get("isEven").(*bubbly.Computed[interface{}]).Get().(bool)
            
            return fmt.Sprintf("Count: %d (%s)", count, 
                if isEven { "even" } else { "odd" })
        }).
        Build()
}

// Usage in parent:
display, _ := components.CreateCounterDisplay(props)

// CORRECT: Use ExposeComponent
if err := ctx.ExposeComponent("display", display); err != nil {
    ctx.Expose("error", fmt.Sprintf("Failed: %v", err))
    return
}

// In template:
display := ctx.Get("display").(bubbly.Component)
return display.View()
```

**Why ExposeComponent:**
- Establishes parent-child relationship in component tree
- Automatically calls `display.Init()` if not initialized
- Parent's `Update(msg)` automatically propagates to `display.Update(msg)`
- Required for DevTools component tree inspection
- Enables proper lifecycle management (onMounted, onUpdated, etc.)

---

### 2. Molecule Components (Built-in Rendering Helpers)

**Definition:** Pre-built components from `pkg/components` package used for inline rendering.

**Characteristics:**
- Factory functions returning `bubbly.Component`
- Use `bubbles/textinput`, `lipgloss` internally
- Update via **event-based pattern** (`Emit("textInputUpdate", msg)`)
- Do NOT override `Update()` method (use default framework behavior)
- Are **rendering helpers**, not tree nodes
- Examples: Input, Button, Text, Badge, Card, Modal, Table, etc.

**Pattern:** Manual `.Init()` + `ctx.Expose()` for storage

**Example:**
```go
// In Setup (store reference):
inputComp := components.Input(components.InputProps{
    Value:       valueRef,
    Placeholder: "Enter text...",
    Width:       50,
    CharLimit:   100,
})

// CRITICAL: Manual Init, NOT ExposeComponent!
inputComp.Init()

// Store as reference (NOT as child)
ctx.Expose("inputComp", inputComp)

// Forward events to molecule component
ctx.On("textInputUpdate", func(data interface{}) {
    inputComp.Emit("textInputUpdate", data)
})

// In Template:
inputComp := ctx.Get("inputComp").(bubbly.Component)
return inputComp.View()
```

**Alternative: Create Inline in Template:**
```go
// In Template (create + render inline):
inputComp := components.Input(components.InputProps{
    Value:       ctx.Get("value").(*bubbly.Ref[string]),
    Placeholder: "Enter text...",
    Width:       25,
})
inputComp.Init()
return inputComp.View()
```

**Why NOT ExposeComponent:**
- Molecule components are **rendering helpers**, not tree nodes
- They use **event-based updates**, not Update() override
- Making them children creates conflict: parent Update() + event updates = **CRASH**
- They're designed for inline composition, not parent-child relationships

---

## Decision Matrix

| Component Type | Source | Use ExposeComponent? | Pattern |
|----------------|--------|---------------------|---------|
| Custom component with Setup/Template | Your `components/` folder | ✅ YES | `ctx.ExposeComponent("name", comp)` |
| Built-in from pkg/components | `components.Input()`, etc. | ❌ NO | `comp.Init()` + `ctx.Expose("name", comp)` |
| Component that manages children | Your app | ✅ YES | `ctx.ExposeComponent("name", comp)` |
| Component for inline rendering | `pkg/components` | ❌ NO | Create in Template + `.Init()` |
| Layout components | `components.AppLayout()`, etc. | ❌ NO | Manual `.Init()` |

---

## Common Mistakes & Fixes

### ❌ Mistake 1: Using ExposeComponent on Molecule Components

```go
// ❌ This will CRASH when emitting events!
inputComp := components.Input(props)
ctx.ExposeComponent("input", inputComp)  // ❌ Makes Input a child - breaks event flow!
```

**Error:** Application crashes when pressing ESC or emitting focus/blur events.

**Fix:**
```go
// ✅ CORRECT
inputComp := components.Input(props)
inputComp.Init()                         // ✅ Manual init
ctx.Expose("input", inputComp)           // ✅ Store reference, NOT as child
```

---

### ❌ Mistake 2: Manual Init on Composable Components

```go
// ❌ This bypasses parent-child relationship!
display, _ := components.CreateCounterDisplay(props)
display.Init()                           // ❌ Manual init bypasses lifecycle
ctx.Expose("display", display)           // ❌ No parent-child relationship!
```

**Error:** Component doesn't receive Update() calls, lifecycle hooks don't fire, DevTools can't see it.

**Fix:**
```go
// ✅ CORRECT
display, _ := components.CreateCounterDisplay(props)
ctx.ExposeComponent("display", display)  // ✅ Auto-init + parent-child relationship
```

---

### ❌ Mistake 3: Forgetting to Init Molecule Components

```go
// ❌ Component not initialized!
inputComp := components.Input(props)
ctx.Expose("input", inputComp)
// Missing .Init() call!

// In Template:
return inputComp.View()  // ❌ Will render blank or crash
```

**Fix:**
```go
// ✅ CORRECT
inputComp := components.Input(props)
inputComp.Init()  // ✅ Required!
ctx.Expose("input", inputComp)
```

---

## Why This Matters: Technical Deep Dive

### ExposeComponent Side Effects

When you call `ctx.ExposeComponent("name", comp)`:

1. **Auto-initialization:** Calls `comp.Init()` if not already initialized
2. **Parent-child registration:** Calls `ctx.component.AddChild(comp)`
3. **Update propagation:** Parent's `Update(msg)` automatically calls `child.Update(msg)` for ALL children

From `pkg/bubbly/context.go`:
```go
func (ctx *Context) ExposeComponent(name string, comp Component) error {
    // Auto-initialize if not already initialized
    if !comp.IsInitialized() {
        cmd := comp.Init()
        if cmd != nil && ctx.component.commandQueue != nil {
            ctx.component.commandQueue.Enqueue(cmd)
        }
    }
    
    // CRITICAL: Establish parent-child relationship
    if err := ctx.component.AddChild(comp); err != nil {
        return fmt.Errorf("failed to add child component: %w", err)
    }
    
    ctx.Expose(name, comp)
    return nil
}
```

### Molecule Component Architecture

Example: `Input` component from `pkg/components/input.go`

**Event-Based Update Pattern:**
```go
// Input component Setup:
ctx.On("textInputUpdate", func(data interface{}) {
    if msg, ok := data.(tea.Msg); ok {
        var cmd tea.Cmd
        ti, cmd = ti.Update(msg)  // Update internal textinput.Model
        
        // Sync value back to props.Value
        newValue := ti.Value()
        if newValue != props.Value.Get().(string) {
            props.Value.Set(newValue)
        }
    }
})
```

**Key Points:**
- Input does NOT override `Update()` method
- Uses default `componentImpl.Update()` from framework
- Internal `textinput.Model` ONLY updates via "textInputUpdate" event
- No automatic Update() propagation from parent needed

### The Conflict (Why It Crashes)

**Scenario: Input is made a child via ExposeComponent**

1. User presses ESC → toggleMode event fires
2. App emits `inputComp.Emit("focus", nil)`
3. TodoInput receives "focus" event
4. TodoInput forwards to Input: `inputComp.Emit("textInputUpdate", msg)`
5. **SIMULTANEOUSLY:**
   - Parent calls `child.Update(msg)` (automatic propagation)
   - Input's "textInputUpdate" handler calls `ti.Update(msg)` (event-based)
6. **Two update paths to same internal state** → race condition → **CRASH**

**Solution:**
- Keep molecule components as **references**, not children
- They receive updates ONLY via events (single update path)
- No automatic Update() propagation from parent
- Clean, predictable event flow

---

## Real-World Example: TodoInput Component

**File:** `cmd/examples/10-testing/02-todo/components/todo_input.go`

```go
func CreateTodoInput(props TodoInputProps) (bubbly.Component, error) {
    return bubbly.NewComponent("TodoInput").
        Setup(func(ctx *bubbly.Context) {
            // CRITICAL: Input is a molecule component
            // DO NOT use ExposeComponent!
            inputComp := components.Input(components.InputProps{
                Value:       props.Value,
                Placeholder: "What needs to be done?",
                Width:       50,
                CharLimit:   100,
                NoBorder:    true,
            })
            
            // Manual Init (proven pattern)
            inputComp.Init()
            
            // Store as reference (NOT as child)
            ctx.Expose("inputComp", inputComp)
            
            // Forward events to Input component
            ctx.On("textInputUpdate", func(data interface{}) {
                inputComp.Emit("textInputUpdate", data)
            })
            
            ctx.On("focus", func(data interface{}) {
                inputComp.Emit("focus", nil)
            })
            
            ctx.On("blur", func(data interface{}) {
                inputComp.Emit("blur", nil)
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            inputComp := ctx.Get("inputComp").(bubbly.Component)
            return inputComp.View()
        }).
        Build()
}
```

**In parent app:**
```go
// TodoInput is a composable component - use ExposeComponent
input, _ := localcomponents.CreateTodoInput(props)
if err := ctx.ExposeComponent("input", input); err != nil {
    // Handle error
}
```

---

## Reference: All pkg/components are Molecules

**Complete list of molecule components (use manual .Init()):**

### Atoms
- `components.Input()`
- `components.Button()`
- `components.Text()`
- `components.Badge()`
- `components.Icon()`
- `components.Spacer()`
- `components.Spinner()`

### Form Components
- `components.Checkbox()`
- `components.Radio()`
- `components.Toggle()`
- `components.Select()`
- `components.Textarea()`
- `components.Form()`

### Data Display
- `components.Table()`
- `components.List()`
- `components.Card()`
- `components.Modal()`

### Navigation
- `components.Tabs()`
- `components.Menu()`
- `components.Accordion()`

### Layout
- `components.AppLayout()`
- `components.PageLayout()`
- `components.GridLayout()`
- `components.PanelLayout()`

**ALL of these require manual `.Init()`, NOT `ExposeComponent`.**

---

## Testing Your Understanding

### Quiz 1: Which pattern to use?

```go
// Scenario: Creating a custom UserCard component
func CreateUserCard(props UserCardProps) (bubbly.Component, error) {
    return bubbly.NewComponent("UserCard").
        Setup(func(ctx *bubbly.Context) {
            // ... setup
        }).
        Template(func(ctx bubbly.RenderContext) string {
            // ... template
        }).
        Build()
}

// In parent:
userCard, _ := CreateUserCard(props)
// Which pattern?
```

**Answer:** ✅ `ctx.ExposeComponent("userCard", userCard)` - It's a custom composable component.

### Quiz 2: Which pattern to use?

```go
// Scenario: Using Input component
inputComp := components.Input(components.InputProps{
    Value: nameRef,
})
// Which pattern?
```

**Answer:** ✅ `inputComp.Init()` + `ctx.Expose("input", inputComp)` - It's a molecule from pkg/components.

### Quiz 3: What's wrong here?

```go
// In Setup:
cardComp := components.Card(components.CardProps{
    Title: "Welcome",
})
ctx.ExposeComponent("card", cardComp)  // Is this correct?
```

**Answer:** ❌ NO! Card is a molecule component. Should be:
```go
cardComp := components.Card(components.CardProps{
    Title: "Welcome",
})
cardComp.Init()
ctx.Expose("card", cardComp)
```

---

## Debugging Checklist

**App crashes when emitting events to components?**
- [ ] Check if you used `ExposeComponent` on a `pkg/components` molecule
- [ ] Verify you're using manual `.Init()` for all `pkg/components` components
- [ ] Confirm event forwarding is one-way (no circular emissions)

**Component not rendering after state change?**
- [ ] Check if component is created in Setup vs Template
- [ ] Verify refs are passed correctly to component props
- [ ] Confirm component is re-created in Template if values change

**DevTools shows missing components in tree?**
- [ ] Check if you used manual `.Init()` on a custom composable component
- [ ] Should be using `ExposeComponent` for custom components
- [ ] Molecule components won't appear in tree (this is correct)

**Lifecycle hooks not firing?**
- [ ] Verify component is registered via `ExposeComponent`
- [ ] Confirm it's a composable component, not a molecule
- [ ] Check parent component is properly initialized

---

## Summary

**Golden Rules:**

1. **All `pkg/components` are molecules** → Manual `.Init()` + `ctx.Expose()`
2. **All custom `components/` are composables** → `ctx.ExposeComponent()`
3. **ExposeComponent = parent-child relationship** → Use for app structure
4. **Manual .Init() = inline rendering** → Use for visual helpers
5. **When in doubt, check working examples** → `09-devtools/`, `10-testing/`

**This guide was created from real production issues. Following these patterns is NOT optional.**
