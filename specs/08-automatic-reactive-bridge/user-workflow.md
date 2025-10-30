# User Workflow: Automatic Reactive Bridge

## Developer Personas

### Persona 1: Vue Developer (Maria)
- **Background**: 6 years Vue.js, transitioning to Go TUI
- **Goal**: Build TUI with familiar reactive patterns
- **Pain Point**: Manual bridge code feels like "step backward"
- **Expects**: State changes trigger updates automatically
- **Success**: Writes code that "just works" like Vue

### Persona 2: Existing BubblyUI User (Chen)
- **Background**: Has app with manual bridge pattern
- **Goal**: Reduce boilerplate, improve maintainability
- **Pain Point**: Too much wrapper model code
- **Expects**: Smooth migration path
- **Success**: Migrates incrementally without breaking changes

### Persona 3: Go Purist (Alex)
- **Background**: Prefers explicit over implicit
- **Goal**: Understand exactly what framework does
- **Pain Point**: "Magic" behavior seems risky
- **Expects**: Clear control and opt-out
- **Success**: Uses automatic mode with confidence

---

## Primary User Journey: Enabling Automatic Mode

### Entry Point: Manual Bridge Boilerplate Pain

**Workflow: Migrating to Automatic Bridge**

#### Step 1: Current State (Manual Bridge)
**User Situation**: Working app with manual bridge code

```go
type model struct {
    component bubbly.Component
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "space" {
            m.component.Emit("increment", nil) // Tedious!
        }
    }
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}
```

**Pain Points**:
- 30+ lines of boilerplate
- Manual Emit() calls everywhere
- Duplicated wrapper pattern
- Hard to maintain

#### Step 2: Learn About Automatic Mode
**User Action**: Read migration guide

**Key Concepts Learned**:
- Ref.Set() generates commands automatically
- bubbly.Wrap() eliminates wrapper model
- Backward compatible with existing code
- Can mix automatic and manual patterns

**Decision Point**:
- ✅ Try automatic mode → Step 3
- 🤔 Stay with manual → Keep current code (still works!)

#### Step 3: Enable Automatic Mode
**User Action**: Add feature flag to component

```go
// Option 1: Enable globally
component := bubbly.NewComponent("Counter").
    WithAutoCommands(true).  // Enable automatic bridge
    Setup(func(ctx *bubbly.Context) {
        count := ctx.Ref(0)
        
        ctx.On("increment", func(_ interface{}) {
            count.Set(count.Get().(int) + 1)
            // No Emit() needed - UI updates automatically! 🎉
        })
        
        ctx.Expose("count", count)
    }).
    Build()

// Option 2: Enable in context
ctx.Setup(func(ctx *bubbly.Context) {
    ctx.EnableAutoCommands() // Per-component control
    // ... rest of setup
})
```

**System Response**:
- Automatic mode enabled
- Refs generate commands on Set()
- Commands queue in component

**UI Feedback**:
- App still works exactly the same
- But internals now automatic

#### Step 4: Simplify Wrapper Model
**User Action**: Replace manual wrapper with bubbly.Wrap()

```go
// Before: 40+ lines
type model struct { component bubbly.Component }
func (m model) Init() tea.Cmd { return m.component.Init() }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... lots of code
}
func (m model) View() string { return m.component.View() }

func main() {
    m := model{component: counter}
    tea.NewProgram(m).Run()
}

// After: 2 lines!
func main() {
    tea.NewProgram(bubbly.Wrap(counter)).Run()
}
```

**System Response**:
- Wrapper handles all bridging automatically
- Commands batch correctly
- Update cycles work

**UI Feedback**:
- App works identically
- Code is 90% shorter
- Easier to understand

**Journey Milestone**: ✅ Basic automatic mode working!

---

### Feature Journey: Zero-Boilerplate Counter

#### Step 5: Write New Component from Scratch
**User Action**: Create counter with automatic mode from start

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "fmt"
)

func main() {
    counter, _ := bubbly.NewComponent("Counter").
        WithAutoCommands(true).  // Automatic mode
        Setup(func(ctx *bubbly.Context) {
            count := ctx.Ref(0)
            
            ctx.On("increment", func(_ interface{}) {
                count.Set(count.Get().(int) + 1)
                // That's it! UI updates automatically
            })
            
            ctx.On("decrement", func(_ interface{}) {
                count.Set(count.Get().(int) - 1)
            })
            
            ctx.Expose("count", count)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[interface{}])
            return fmt.Sprintf("Count: %d\nPress +/- to change", count.Get())
        }).
        Build()
    
    // One-line integration!
    tea.NewProgram(bubbly.Wrap(counter)).Run()
}
```

**System Response**:
- Component created with auto mode
- Commands generate on every Set()
- Wrapper handles Bubbletea integration

**UI Feedback**:
- Press '+' → count increments → UI updates automatically
- Press '-' → count decrements → UI updates automatically
- Zero manual bridge code needed

**Comparison with Manual Mode**:
- **Manual**: 60 lines (component + wrapper model)
- **Automatic**: 30 lines (component only)
- **Reduction**: 50% less code!

---

### Feature Journey: Mixed Automatic and Manual

#### Step 6: Use Both Patterns in Same App
**User Action**: Some refs automatic, some manual

```go
ctx.Setup(func(ctx *bubbly.Context) {
    // Automatic ref (default)
    displayCount := ctx.Ref(0)
    
    // Manual ref (for specific control)
    internalCounter := ctx.ManualRef(0)
    
    ctx.On("display-increment", func(_ interface{}) {
        displayCount.Set(displayCount.Get().(int) + 1)
        // Auto-updates UI
    })
    
    ctx.On("internal-increment", func(_ interface{}) {
        val := internalCounter.Get().(int)
        internalCounter.Set(val + 1)
        
        // Only update display every 10 increments
        if val % 10 == 0 {
            ctx.Emit("sync-display", val)
        }
    })
    
    ctx.On("sync-display", func(data interface{}) {
        displayCount.Set(data.(int))
    })
})
```

**System Response**:
- Automatic refs generate commands
- Manual refs don't generate commands
- Both work together seamlessly

**UI Feedback**:
- Display updates automatically for automatic refs
- Manual refs need explicit Emit() (as expected)
- Clear control over update behavior

**Use Case**: Performance optimization for high-frequency updates

---

## Alternative Workflows

### Workflow A: Opt-Out for Specific Cases

#### Entry: Need Manual Control for Performance

**Scenario**: Tight loop updating many refs

```go
ctx.Setup(func(ctx *bubbly.Context) {
    data := ctx.Ref([]int{})
    
    ctx.On("process-batch", func(input interface{}) {
        items := input.([]int)
        
        // Disable auto-commands for batch processing
        ctx.DisableAutoCommands()
        
        result := []int{}
        for _, item := range items {
            result = append(result, item * 2)
        }
        
        data.Set(result)
        
        // Re-enable and trigger single update
        ctx.EnableAutoCommands()
        ctx.Emit("batch-complete", nil)
    })
})
```

**Why Opt-Out**:
- 10,000 Set() calls would generate 10,000 commands
- Only need 1 UI update at the end
- Performance optimization

**Result**: Full control when needed

---

### Workflow B: Debugging Automatic Behavior

#### Entry: UI Not Updating as Expected

**Problem**: State changes but UI doesn't update

**Debug Steps**:

1. **Enable Command Debug Mode**
```go
ctx.EnableCommandDebug()
// Logs all command generation to console
```

2. **Check Console Output**
```
[DEBUG] Component 'Counter' ref 'count': 0 → 1 (command generated)
[DEBUG] Command queued: StateChangedMsg{ComponentID: "counter-1", RefID: "ref-42"}
[DEBUG] Command batch size: 1
[DEBUG] Command executed: StateChangedMsg
```

3. **Verify Auto Mode Enabled**
```go
if !ctx.IsAutoCommandsEnabled() {
    log.Println("Auto commands disabled - enable them!")
}
```

4. **Check Wrapper Usage**
```go
// Wrong: Manual wrapper without auto-mode support
model := manualModel{component: comp}

// Right: Use automatic wrapper
model := bubbly.Wrap(comp)
```

**Resolution**: Identify and fix configuration issue

---

## Error Recovery Workflows

### Error Flow 1: Ref.Set() in Template

**Trigger**: Calling Set() inside template function

```go
.Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    count.Set(100) // ❌ ERROR: Can't mutate in template!
    return fmt.Sprintf("Count: %d", count.Get())
})
```

**User Sees**:
```
PANIC: Cannot call Ref.Set() in template context
Templates must be pure functions (no side effects)

Move state updates to event handlers or lifecycle hooks.
```

**Recovery**:
1. Move Set() to event handler
2. Use computed value if derived state needed
3. Keep templates pure (read-only)

**Fixed Code**:
```go
.Setup(func(ctx *bubbly.Context) {
    count := ctx.Ref(0)
    
    ctx.On("initialize", func(_ interface{}) {
        count.Set(100) // ✅ Correct: Set in handler
    })
})
.Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    return fmt.Sprintf("Count: %d", count.Get()) // ✅ Read only
})
```

---

### Error Flow 2: Command Queue Overflow

**Trigger**: Infinite loop generating commands

```go
ctx.OnUpdated(func() {
    count.Set(count.Get().(int) + 1)
    // This triggers onUpdated again → infinite loop!
}, count)
```

**User Sees**:
```
ERROR: Maximum update depth exceeded (100)
Component 'Counter' appears to have infinite update loop

Check onUpdated hooks for recursive state changes.
```

**Recovery**:
1. Review onUpdated hooks
2. Add condition to prevent infinite recursion
3. Use different trigger mechanism

**Fixed Code**:
```go
ctx.OnUpdated(func() {
    current := count.Get().(int)
    if current < 100 {  // ✅ Condition prevents infinite loop
        count.Set(current + 1)
    }
}, count)
```

---

## State Transition Diagrams

### Automatic Update Lifecycle
```
User Action
    ↓
Event Handler Executes
    ↓
Ref.Set(newValue) called
    ↓
┌─────────────────────────────────┐
│ Ref Internal State Updates (sync)│
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ Command Generated Automatically  │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ Command Queued in Component      │
└─────────────────────────────────┘
    ↓
Event Handler Returns
    ↓
Bubbletea Calls Update()
    ↓
┌─────────────────────────────────┐
│ Component Returns Batched Commands│
└─────────────────────────────────┘
    ↓
Bubbletea Executes Commands
    ↓
┌─────────────────────────────────┐
│ StateChangedMsg Sent to Update() │
└─────────────────────────────────┘
    ↓
onUpdated Hooks Execute
    ↓
View() Re-renders
    ↓
User Sees Updated UI
```

### Command Batching Flow
```
Multiple Ref.Set() in Same Tick
    ↓
Ref1.Set() → Command 1 queued
Ref2.Set() → Command 2 queued
Ref3.Set() → Command 3 queued
    ↓
Update() Returns
    ↓
Batcher Receives Commands
    ↓
Strategy: CoalesceAll
    ↓
Single Batched Command Generated
    ↓
Bubbletea Executes Once
    ↓
Single StateChangedMsg
    ↓
All State Changes Visible Together
```

---

## Integration Points Map

### Feature Cross-Reference
```
08-automatic-reactive-bridge
    ← Enhances: 01-reactivity-system (Ref command generation)
    ← Enhances: 02-component-model (command queuing)
    ← Enhances: 03-lifecycle-hooks (command lifecycle)
    ← Uses: 04-composition-api (composable patterns)
    → Works with: 05-directives (directive commands)
    → Works with: 07-router (router commands)
    → Enables: Simpler application code
```

---

## User Success Paths

### Path 1: Quick Win (< 15 minutes)
```
Manual app → Enable auto mode → Use Wrap() → Remove Emit() calls → Success! 🎉
Code reduction: 30-50%
```

### Path 2: New Project (< 5 minutes)
```
Start new → Enable auto from beginning → Write component → One-line integration → Success! 🎉
Zero boilerplate from day 1
```

### Path 3: Gradual Migration (< 1 hour)
```
Large app → Enable auto → Migrate one component → Test → Migrate next → All migrated → Success! 🎉
Incremental, safe migration
```

---

## Common Patterns

### Pattern 1: Simple Counter (Automatic)
```go
counter := bubbly.NewComponent("Counter").
    WithAutoCommands(true).
    Setup(func(ctx *bubbly.Context) {
        count := ctx.Ref(0)
        ctx.On("inc", func(_ interface{}) { 
            count.Set(count.Get().(int) + 1) 
        })
        ctx.Expose("count", count)
    }).
    Template(func(ctx bubbly.RenderContext) string {
        count := ctx.Get("count").(*bubbly.Ref[interface{}])
        return fmt.Sprintf("Count: %d", count.Get())
    }).
    Build()

tea.NewProgram(bubbly.Wrap(counter)).Run()
```

### Pattern 2: Form with Validation
```go
form := bubbly.NewComponent("Form").
    WithAutoCommands(true).
    Setup(func(ctx *bubbly.Context) {
        name := ctx.Ref("")
        email := ctx.Ref("")
        valid := ctx.Computed(func() interface{} {
            n := name.Get().(string)
            e := email.Get().(string)
            return len(n) > 0 && strings.Contains(e, "@")
        }, name, email)
        
        ctx.On("name-change", func(data interface{}) {
            name.Set(data.(string))
            // UI updates automatically with validation state!
        })
        
        ctx.Expose("name", name)
        ctx.Expose("email", email)
        ctx.Expose("valid", valid)
    }).
    Build()
```

### Pattern 3: Real-time Data Updates
```go
dashboard := bubbly.NewComponent("Dashboard").
    WithAutoCommands(true).
    Setup(func(ctx *bubbly.Context) {
        metrics := ctx.Ref([]Metric{})
        
        // Automatic updates from ticker
        ctx.OnMounted(func() {
            // Ticker updates metrics every second
            // Each update automatically triggers UI refresh!
        })
    }).
    Build()
```

---

## Performance Comparison

### Manual Mode
```
User Action → Event Handler → Ref.Set() → ctx.Emit() → Update() → Re-render
Time: 12,000 ns
```

### Automatic Mode
```
User Action → Event Handler → Ref.Set() → [Command Gen] → Update() → Re-render
Time: 12,010 ns
```

**Overhead**: < 10ns (0.08%) ✅

---

## Summary

The Automatic Reactive Bridge transforms BubblyUI development by eliminating 30-50% of boilerplate code. Developers enable automatic mode, use `bubbly.Wrap()` for one-line integration, and write `Ref.Set()` calls that automatically trigger UI updates. The system maintains backward compatibility, allows mixing automatic and manual patterns, and provides clear opt-out mechanisms. Performance overhead is negligible (< 10ns), error handling is production-grade, and the migration path is smooth and incremental.

**Key Success Factors**:
- ✅ Vue-like DX (automatic updates)
- ✅ Backward compatible (existing code works)
- ✅ Opt-out when needed (manual control available)
- ✅ Zero performance penalty
- ✅ Production-ready error handling
- ✅ Clear migration guide
