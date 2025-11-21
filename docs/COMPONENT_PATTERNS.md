# BubblyUI Component Patterns Guide

**Unified Pattern for All Components**

**Version:** 3.1  
**Last Updated:** November 20, 2025  
**Status:** UNIFIED - Single Pattern for All Components  
**Target Audience:** All BubblyUI Developers

---

## 🎉 Great News: Unified Pattern!

**All BubblyUI components now use the same pattern.**

As of v3.1, the framework has been refactored to use **one universal pattern** for all components. No more confusion about "molecule" vs "composable" components!

---

## The Universal Pattern: ExposeComponent

**Use `ctx.ExposeComponent()` for ALL components - both custom and built-in.**

### Example: Custom App Component

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

### Example: Built-in Components from pkg/components

**All built-in components now work with `ExposeComponent`!**

**Example: Using Input Component**
```go
// In Setup:
inputComp := components.Input(components.InputProps{
    Value:       valueRef,
    Placeholder: "Enter text...",
    Width:       50,
    CharLimit:   100,
})

// ✅ UNIFIED PATTERN: Same as custom components!
if err := ctx.ExposeComponent("inputComp", inputComp); err != nil {
    ctx.Expose("error", err)
    return
}

// Focus/blur events (for state management)
ctx.On("setFocus", func(data interface{}) {
    input := ctx.Get("inputComp").(bubbly.Component)
    input.Emit("focus", nil)
})

// In Template:
inputComp := ctx.Get("inputComp").(bubbly.Component)
return inputComp.View()
```

**How This Works:**
- Input component uses `WithMessageHandler` internally
- Keyboard messages are processed via the handler before child Update() propagation
- Focus/blur events manage state without conflicting with Update() flow
- No more dual update paths - clean, single architecture

---

## Decision Matrix

**Simple: Always use ExposeComponent!**

| Component Type | Source | Use ExposeComponent? | Pattern |
|----------------|--------|---------------------|---------|
| Custom component with Setup/Template | Your `components/` folder | ✅ YES | `ctx.ExposeComponent("name", comp)` |
| Built-in from pkg/components | `components.Input()`, `components.Card()`, etc. | ✅ YES | `ctx.ExposeComponent("name", comp)` |
| Component that manages children | Your app | ✅ YES | `ctx.ExposeComponent("name", comp)` |
| ANY component | Anywhere | ✅ YES | `ctx.ExposeComponent("name", comp)` |
| Layout components | `components.AppLayout()`, etc. | ❌ NO | Manual `.Init()` |

---

## Benefits of Unified Pattern

### ✅ Simpler to Learn
- Only one pattern to remember
- No mental overhead deciding which pattern to use
- Consistent across entire codebase

### ✅ Cleaner Code
- No manual `.Init()` calls
- Automatic parent-child relationships
- Less boilerplate

### ✅ Better DevTools Integration
- All components visible in component tree
- Easier debugging and inspection
- Proper lifecycle tracking

### ✅ Fewer Bugs
- No dual update path conflicts
- Correct Update() propagation
- Proper cleanup and lifecycle

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

### How Input Component Was Fixed

The `Input` component now uses `WithMessageHandler` to intercept keyboard messages **before** child Update() processing:

**New Architecture:**
```go
// Input component (simplified):
bubbly.NewComponent("Input").
    Setup(func(ctx *bubbly.Context) {
        ti := textinput.New()  // bubbles/textinput
        focusedRef := bubbly.NewRef(false)
        
        // Internal event handler for keyboard processing
        ctx.On("__processKeyboard", func(data interface{}) {
            if msg, ok := data.(tea.Msg); ok {
                if focusedRef.GetTyped() {
                    ti, cmd = ti.Update(msg)  // Update textinput
                    // Sync value...
                }
            }
        })
        
        // Focus/blur events
        ctx.On("focus", func(_ interface{}) {
            focusedRef.Set(true)
            ti.Focus()
        })
    }).
    WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
        // Forward to internal handler
        comp.Emit("__processKeyboard", msg)
        return nil
    }).
    Build()
```

**Key Points:**
- `WithMessageHandler` runs **before** child Update() propagation
- Internal handler `__processKeyboard` has access to textinput state
- Only processes when focused
- Single update path - no conflicts
- Works perfectly with `ExposeComponent`!

---

## Real-World Example: TodoInput Component

**File:** `cmd/examples/10-testing/02-todo/components/todo_input.go`

**Using the Unified Pattern:**

```go
func CreateTodoInput(props TodoInputProps) (bubbly.Component, error) {
    return bubbly.NewComponent("TodoInput").
        Setup(func(ctx *bubbly.Context) {
            // Create Input component
            inputComp := components.Input(components.InputProps{
                Value:       props.Value,
                Placeholder: "What needs to be done?",
                Width:       50,
                CharLimit:   100,
                NoBorder:    true,
            })
            
            // ✅ UNIFIED PATTERN: Use ExposeComponent!
            if err := ctx.ExposeComponent("inputComp", inputComp); err != nil {
                ctx.Expose("error", fmt.Sprintf("Failed: %v", err))
                return
            }
            
            // Forward focus/blur for state management
            ctx.On("setFocus", func(data interface{}) {
                comp := ctx.Get("inputComp").(bubbly.Component)
                comp.Emit("focus", nil)
            })
            
            ctx.On("setBlur", func(data interface{}) {
                comp := ctx.Get("inputComp").(bubbly.Component)
                comp.Emit("blur", nil)
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

**Answer:** ✅ `ctx.ExposeComponent("input", inputComp)` - Same unified pattern for all!

### Quiz 3: What's the correct pattern?

```go
// In Setup:
cardComp := components.Card(components.CardProps{
    Title: "Welcome",
})
ctx.ExposeComponent("card", cardComp)  // Is this correct?
```

**Answer:** ✅ YES! All components use ExposeComponent now. This is the unified pattern!

---

## Debugging Checklist

**Component not rendering after state change?**
- [ ] Check if component was properly exposed with `ExposeComponent`
- [ ] Verify refs are passed correctly to component props
- [ ] Confirm reactive values are updating

**DevTools shows missing components in tree?**
- [ ] Verify you used `ExposeComponent` (not manual `.Init()`)
- [ ] Check parent component is properly initialized
- [ ] All components should be visible now with unified pattern

**Lifecycle hooks not firing?**
- [ ] Verify component is registered via `ExposeComponent`
- [ ] Check parent component is properly initialized
- [ ] Ensure component has lifecycle hooks defined

**Event loop issues?**
- [ ] Check for circular event emissions (e.g., "focus" → "focus")
- [ ] Use internal event names to prevent bubble-back loops
- [ ] Example: Parent emits "setFocus" → child forwards as "focus"

---

## Summary

**The Golden Rule:**

**Use `ctx.ExposeComponent()` for ALL components!**

1. ✅ **Custom components** → `ctx.ExposeComponent()`
2. ✅ **Built-in components** → `ctx.ExposeComponent()`
3. ✅ **Input component** → `ctx.ExposeComponent()`
4. ✅ **Everything** → `ctx.ExposeComponent()`

**Benefits:**
- Simpler to learn - one pattern
- Cleaner code - less boilerplate
- Better DevTools integration
- Fewer bugs - no dual update paths

**This unified pattern was achieved through careful refactoring of the Input component to use WithMessageHandler, enabling it to work seamlessly with ExposeComponent.**
