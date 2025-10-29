# Bubbletea Integration Architecture

## Overview

BubblyUI is built as an enhancement layer on top of Bubbletea, not a replacement. This document explains how BubblyUI components integrate with Bubbletea's message-passing architecture, why certain patterns exist, and the roadmap for future improvements.

## Core Philosophy

**Enhance, Don't Replace**: BubblyUI embraces Bubbletea's proven architecture while adding Vue-inspired abstractions that make development more productive and enjoyable.

### Why Build on Bubbletea?

1. **Battle-tested**: Bubbletea is mature, stable, and widely adopted
2. **Ecosystem**: Leverage existing Bubbles components and community tools
3. **Architecture**: The Elm Architecture (TEA) is sound and well-understood
4. **Migration**: Gradual adoption path for existing Bubbletea applications
5. **Performance**: Proven performance characteristics and optimization

---

## The Bubbletea Model

### Standard Bubbletea Application

```go
package main

import tea "github.com/charmbracelet/bubbletea"

type model struct {
    count int
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "space" {
            m.count++
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("Count: %d", m.count)
}
```

### Bubbletea's Message Loop

```
User Input (keyboard, mouse, etc.)
    ↓
tea.Msg (message)
    ↓
Update(msg) → Returns (newModel, tea.Cmd)
    ↓
View() → Renders UI string
    ↓
Terminal displays output
    ↓
Loop continues...
```

**Key Insight**: Bubbletea controls when `Update()` is called. UI only re-renders after `Update()` returns.

---

## The BubblyUI Layer

### BubblyUI Component

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func createCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        Setup(func(ctx *bubbly.Context) {
            count := ctx.Ref(0)
            ctx.Expose("count", count)
            
            ctx.On("increment", func(_ interface{}) {
                current := count.Get().(int)
                count.Set(current + 1)
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[interface{}])
            return fmt.Sprintf("Count: %d", count.Get().(int))
        }).
        Build()
}
```

**Features Added**:
- ✅ Reactive state (`Ref`)
- ✅ Event system (`On`, `Emit`)
- ✅ Template functions
- ✅ Lifecycle hooks
- ✅ Type safety

---

## The Integration Pattern (Current)

### Manual Bridge Model

BubblyUI components must be wrapped in a Bubbletea model to connect to the message loop:

```go
type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // 1. Handle Bubbletea messages
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "space":
            // 2. Manually bridge to component
            m.component.Emit("increment", nil)
        }
    case customMsg:
        m.component.Emit("custom-event", msg)
    }
    
    // 3. Forward message to component
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    
    if cmd != nil {
        cmds = append(cmds, cmd)
    }
    
    return m, tea.Batch(cmds...)
}

func (m model) View() string {
    return m.component.View()
}
```

### Why This Pattern Exists

#### Reason 1: Bubbletea's Message Loop Control

Bubbletea controls when `Update()` is called. Components cannot directly trigger re-renders.

**Vue/React Approach** (doesn't work in TUI):
```javascript
// In web browsers, this triggers automatic re-render
const [count, setCount] = useState(0)
setCount(count + 1) // DOM updates automatically
```

**Bubbletea Constraint**:
```go
// This updates state but DOESN'T trigger re-render
count.Set(count.Get().(int) + 1)

// Re-render only happens when Update() is called
// which only happens when Bubbletea receives a message
```

#### Reason 2: Asynchronous Operations

Bubbletea uses commands (`tea.Cmd`) for async work. Direct goroutines break the message loop:

```go
// ❌ DON'T DO THIS - breaks Bubbletea
go func() {
    data := fetchData()
    component.Emit("data", data) // Won't trigger Update()
}()

// ✅ DO THIS - use tea.Cmd
func fetchDataCmd() tea.Cmd {
    return func() tea.Msg {
        data := fetchData()
        return dataMsg{data: data}
    }
}
```

#### Reason 3: Event to Message Translation

Component events must be translated to Bubbletea messages:

```
Component Event → Bridge Model → tea.Msg → Update() → Re-render
```

Without the bridge, component events have no way to trigger the Bubbletea message loop.

---

## Pattern Details

### Pattern 1: Keypress to Event

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "space":
            m.component.Emit("increment", nil)
        case "r":
            m.component.Emit("reset", nil)
        }
    }
    
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    return m, cmd
}
```

**Flow**:
1. User presses space
2. Bubbletea creates `tea.KeyMsg`
3. Model's `Update()` receives message
4. Model emits component event
5. Component event handler runs
6. `Ref.Set()` updates state
7. Component's `Update()` processes message
8. `View()` re-renders with new state

### Pattern 2: Async Data Fetching

```go
func (m model) Init() tea.Cmd {
    // Start async fetch
    return tea.Batch(m.component.Init(), fetchDataCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dataMsg:
        // Data arrived, forward to component
        m.component.Emit("data-loaded", msg.data)
    }
    
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    return m, cmd
}
```

**Flow**:
1. `Init()` starts async command
2. Data fetches in background
3. Command returns `dataMsg`
4. `Update()` receives message
5. Emits event to component
6. Component updates state
7. `View()` re-renders

### Pattern 3: Timer Tick

```go
func (m model) Init() tea.Cmd {
    return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tickMsg:
        m.component.Emit("tick", time.Time(msg))
        // Return next tick to keep timer running
        cmds = append(cmds, tickCmd())
    }
    
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    
    if cmd != nil {
        cmds = append(cmds, cmd)
    }
    
    return m, tea.Batch(cmds...)
}

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

**Key Point**: Each tick generates a message that flows through `Update()`, triggering re-render.

---

## The Trade-off

### Current Pattern
**Pros**:
- ✅ Explicit and clear
- ✅ Full control over message flow
- ✅ Easy to debug
- ✅ Follows Go philosophy (explicit over implicit)
- ✅ Works with all Bubbletea features
- ✅ Compatible with existing Bubbletea code

**Cons**:
- ❌ Boilerplate wrapper model required
- ❌ Manual `Emit()` calls needed
- ❌ Two places to handle logic (model + component)
- ❌ Not as DX-friendly as Vue/React

### What Vue/React Do

```javascript
// Automatic re-render
setState(newValue) // Done! UI updates automatically
```

BubblyUI **cannot** do this because:
1. TUI has no DOM to diff and update
2. Bubbletea controls the render loop
3. No browser event loop to hook into

---

## Future Solution: Automatic Reactive Bridge (Phase 4)

### Goal: Eliminate Manual Bridge

**Vision**: State changes automatically generate Bubbletea commands

### Design Concept

#### Step 1: Command-Generating Refs

```go
// In Context.Ref() implementation:
func (ctx *Context) Ref(value interface{}) *Ref[interface{}] {
    ref := NewRef(value)
    
    // Auto-wire: state changes → tea.Cmd
    ref.onChange = func() tea.Cmd {
        return func() tea.Msg {
            return StateChangedMsg{
                ComponentID: ctx.component.id,
                RefID: ref.id,
            }
        }
    }
    
    return ref
}
```

#### Step 2: Component Runtime

```go
// Component collects commands from state changes
type componentImpl struct {
    pendingCommands []tea.Cmd
}

func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle state change messages
    if stateMsg, ok := msg.(StateChangedMsg); ok {
        // State already updated, just re-render
        return c, nil
    }
    
    // Execute hooks, collect commands
    c.lifecycle.executeUpdated()
    
    // Batch all pending commands
    cmds := c.pendingCommands
    c.pendingCommands = nil
    
    return c, tea.Batch(cmds...)
}
```

#### Step 3: Simplified User Code

```go
// BEFORE (current):
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "space" {
            m.component.Emit("increment", nil) // Manual!
        }
    }
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    return m, cmd
}

// AFTER (Phase 4):
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Component handles everything automatically!
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    return m, cmd
}

// Or even simpler with helper:
func main() {
    component, _ := createCounter()
    tea.NewProgram(bubbly.Wrap(component)).Run() // One-liner!
}
```

### Benefits

1. **Vue-like DX**: State changes trigger re-renders automatically
2. **Less Boilerplate**: No manual `Emit()` calls
3. **Cleaner Code**: Logic stays in component
4. **Easier Learning**: Familiar to web developers
5. **Backward Compatible**: Old pattern still works

### Implementation Timeline

**Phase 4 Enhancement** (Post v1.0):
- Design automatic bridge system
- Implement command generation
- Add `bubbly.Wrap()` helper
- Benchmark performance impact
- Migration guide for existing code
- Maintain backward compatibility

---

## Current Best Practices

Until automatic bridge is implemented, follow these patterns:

### 1. Wrapper Model Template

```go
type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Handle your app's messages
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Map keys to events
    case yourCustomMsg:
        // Forward to component
    }
    
    // Let component process
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    
    if cmd != nil {
        cmds = append(cmds, cmd)
    }
    
    return m, tea.Batch(cmds...)
}

func (m model) View() string {
    return m.component.View()
}
```

### 2. Event Naming Convention

```go
// Use verb-noun pattern
ctx.On("increment-counter", ...)
ctx.On("fetch-data", ...)
ctx.On("submit-form", ...)

// Not noun-only
ctx.On("counter", ...) // ❌ Unclear
ctx.On("data", ...)    // ❌ Unclear
```

### 3. Command Batching

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Collect all commands
    if keyCmd := m.handleKeypress(msg); keyCmd != nil {
        cmds = append(cmds, keyCmd)
    }
    
    updated, componentCmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    
    if componentCmd != nil {
        cmds = append(cmds, componentCmd)
    }
    
    // Batch all at once
    return m, tea.Batch(cmds...)
}
```

### 4. Type-Safe Messages

```go
// Define custom message types
type dataFetchedMsg struct {
    data *UserData
    err  error
}

type timerTickMsg time.Time

// Use in Update()
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dataFetchedMsg:
        m.component.Emit("data-loaded", msg)
    case timerTickMsg:
        m.component.Emit("tick", time.Time(msg))
    }
    // ...
}
```

---

## Integration with Bubbles

BubblyUI components can use Bubbles components:

```go
import "github.com/charmbracelet/bubbles/textinput"

bubbly.NewComponent("InputForm").
    Setup(func(ctx *bubbly.Context) {
        // Use Bubbles textinput
        input := textinput.New()
        input.Placeholder = "Enter name"
        
        // Store in component state
        ctx.Expose("input", ctx.Ref(input))
    }).
    Template(func(ctx bubbly.RenderContext) string {
        input := ctx.Get("input").(*bubbly.Ref[interface{}])
        ti := input.Get().(textinput.Model)
        return ti.View()
    }).
    Build()
```

---

## Migration Guide

### From Raw Bubbletea to BubblyUI

#### Step 1: Keep Existing Model

```go
// Your existing Bubbletea model - keep as-is
type model struct {
    oldState string
}
```

#### Step 2: Create BubblyUI Component for One Feature

```go
// New: Create component for new feature
newFeature, _ := bubbly.NewComponent("NewFeature").
    Setup(func(ctx *bubbly.Context) {
        // Use reactive state
        data := ctx.Ref("")
        ctx.Expose("data", data)
    }).
    Template(func(ctx bubbly.RenderContext) string {
        // Render new feature
    }).
    Build()
```

#### Step 3: Integrate in Update()

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Old state handling
    // ...
    
    // New: Forward to BubblyUI component
    if msg.Type == newFeatureMsg {
        updated, cmd := newFeature.Update(msg)
        newFeature = updated.(bubbly.Component)
        return m, cmd
    }
    
    return m, nil
}
```

#### Step 4: Gradual Migration

```go
// Migrate features one by one
// Keep old code until new component is tested
// Eventually, whole model becomes BubblyUI component
```

---

## Performance Considerations

### Overhead Measurement

```
Raw Bubbletea Update cycle: 7,200 ns/op
BubblyUI Update cycle:      8,000 ns/op
Overhead:                   ~11%
```

**Acceptable**: Target was <15% overhead

### Optimization Tips

1. **Minimize State Updates**
   ```go
   // ❌ Multiple updates
   count.Set(1)
   count.Set(2)
   count.Set(3)
   
   // ✅ Single update
   count.Set(3)
   ```

2. **Batch Events**
   ```go
   // Emit multiple events in same Update() cycle
   // They'll be processed together
   ```

3. **Lazy Computed Values**
   ```go
   // Computed values only recalculate when accessed
   total := ctx.Computed(func() interface{} {
       // Expensive calculation
       return heavyComputation()
   }, dep1, dep2)
   ```

---

## Debugging Tips

### 1. Add Logging to Bridge

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    log.Printf("Update received: %T", msg)
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        log.Printf("Key pressed: %s", msg.String())
        m.component.Emit("keypress", msg)
    }
    
    updated, cmd := m.component.Update(msg)
    log.Printf("Component returned cmd: %v", cmd != nil)
    
    m.component = updated.(bubbly.Component)
    return m, cmd
}
```

### 2. Verify Event Handlers

```go
ctx.On("my-event", func(data interface{}) {
    log.Printf("Event handler called with: %+v", data)
    // Your logic
})
```

### 3. Check State Updates

```go
count := ctx.Ref(0)

ctx.On("increment", func(_ interface{}) {
    current := count.Get().(int)
    log.Printf("Before: %d", current)
    count.Set(current + 1)
    log.Printf("After: %d", count.Get().(int))
})
```

---

## Summary

The current manual bridge pattern exists because:

1. **Bubbletea controls the message loop** - Components can't trigger re-renders directly
2. **Async operations need commands** - Can't use goroutines directly
3. **Events need translation to messages** - Bridge connects component events to Bubbletea messages

This pattern is:
- ✅ **Explicit and debuggable**
- ✅ **Follows Go philosophy**
- ✅ **Compatible with Bubbletea ecosystem**
- ⚠️ **Requires boilerplate**

**Future (Phase 4)**: Automatic reactive bridge will eliminate boilerplate while maintaining all benefits.

**Current Status**: Production-ready pattern that works well. Manual bridge is a temporary trade-off for excellent type safety and explicit control.

**Recommendation**: Use current pattern confidently. It's not a hack - it's a deliberate design that respects both Bubbletea's architecture and Go's philosophy. The automatic bridge enhancement will improve DX without changing fundamentals.
