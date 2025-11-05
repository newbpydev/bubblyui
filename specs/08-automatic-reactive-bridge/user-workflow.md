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

## Workflow 4: Declarative Key Bindings for Zero-Boilerplate

### Persona: Developer Converting from Manual Keyboard Handling

**Background**: Developer has been writing manual keyboard routing code and discovers the declarative key binding system.

**Goal**: Eliminate all keyboard handling boilerplate using declarative key bindings.

### Journey

#### Step 1: Identify Boilerplate Pattern

Developer realizes they're writing the same pattern repeatedly:

```go
// Manual pattern - 40 lines per component
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "space": m.component.Emit("toggle", nil)
        case "enter": m.component.Emit("submit", nil)
        case "esc": m.component.Emit("cancel", nil)
        case "up": m.component.Emit("selectPrevious", nil)
        case "down": m.component.Emit("selectNext", nil)
        // ... 15+ more cases
        }
    }
    // ... rest of Update
}
```

**Pain Point**: 40 lines of repetitive boilerplate per component

#### Step 2: Discover Key Binding System

Reads documentation about `WithKeyBinding()`:

```go
component := bubbly.NewComponent("TodoApp").
    WithAutoCommands(true).
    WithKeyBinding("space", "toggle", "Toggle completion").
    WithKeyBinding("enter", "submit", "Submit form").
    WithKeyBinding("esc", "cancel", "Cancel").
    WithKeyBinding("up", "selectPrevious", "Move up").
    WithKeyBinding("down", "selectNext", "Move down").
    Setup(func(ctx *Context) {
        ctx.On("toggle", func(_ interface{}) {
            // Just handle the event - no keyboard routing!
        })
    }).
    Build()
```

**Realization**: 40 lines → 5 lines of declarations!

#### Step 3: Convert Component

**Before (Manual):**
```go
type model struct {
    component bubbly.Component
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "space":
            m.component.Emit("toggle", nil)
        case "enter":
            m.component.Emit("submit", nil)
        // ... 15+ more cases
        }
    }
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}

func main() {
    m := model{component: createComponent()}
    tea.NewProgram(m).Run()
}
```

**After (Declarative):**
```go
func main() {
    component := bubbly.NewComponent("App").
        WithAutoCommands(true).
        WithKeyBinding("ctrl+c", "quit", "Quit").
        WithKeyBinding("space", "toggle", "Toggle").
        WithKeyBinding("enter", "submit", "Submit").
        // ... more bindings (1 line each)
        Setup(func(ctx *Context) {
            // Just event handlers - no keyboard logic!
            ctx.On("toggle", func(_ interface{}) {
                // State changes auto-update UI
            })
        }).
        Build()
    
    tea.NewProgram(bubbly.Wrap(component)).Run()
}
```

**Result**: Deleted 40-line wrapper model entirely!

#### Step 4: Add Auto-Generated Help

Developer discovers help text is automatic:

```go
.Template(func(ctx RenderContext) string {
    // Get component reference
    comp := ctx.GetComponent()
    
    // Help text auto-generated from bindings!
    help := comp.HelpText()
    // Returns: "ctrl+c: Quit • space: Toggle • enter: Submit • up: Move up • down: Move down"
    
    return mainContent + "\n\n" + help
})
```

**Benefit**: Self-documenting code, help stays in sync automatically

#### Step 5: Implement Mode-Based Keys

Developer needs space key to toggle in navigation mode OR type space in input mode:

```go
component := bubbly.NewComponent("TodoApp").
    WithAutoCommands(true).
    Setup(func(ctx *Context) {
        inputMode := ctx.Ref(false)
        ctx.Expose("inputMode", inputMode)
    }).
    // Conditional bindings for modes
    WithConditionalKeyBinding(KeyBinding{
        Key:   "space",
        Event: "toggle",
        Description: "Toggle completion",
        Condition: func() bool {
            mode := component.Get("inputMode").(*Ref[interface{}])
            return !mode.Get().(bool) // Only in navigation mode
        },
    }).
    WithConditionalKeyBinding(KeyBinding{
        Key:   "space",
        Event: "addChar",
        Data:  " ",
        Description: "Add space",
        Condition: func() bool {
            mode := component.Get("inputMode").(*Ref[interface{}])
            return mode.Get().(bool) // Only in input mode
        },
    }).
    Build()
```

**Result**: Mode-based input handling without boilerplate!

### Metrics

**Code Reduction**:
- Manual wrapper model: 40 lines → 0 lines (100% reduction)
- Keyboard routing: 20 lines → 5 declarative bindings (75% reduction)
- **Total saved per component: 55+ lines**

**Development Time**:
- Manual pattern: 30 minutes (write wrapper, test routing)
- Declarative pattern: 5 minutes (declare bindings)
- **Time saved: 83% faster**

**Maintainability**:
- Manual: Keys scattered in Update(), hard to see all bindings
- Declarative: All keys visible at component definition
- **Improvement: 10x better at-a-glance understanding**

---

## Workflow 5: Complex Application with Tree Structure

### Persona: Advanced Developer Building Production TUI

**Background**: Developer building a complex TUI application (file manager, dashboard, etc.) with nested components and custom message types.

**Goal**: Build Vue-like component tree with layout components and mixed message handling patterns.

### Journey

#### Step 1: Design Component Tree

Developer plans application structure (Vue-like):

```
AppComponent (Root)
├── HeaderComponent (logo, navigation)
├── ContentComponent (PageLayout)
│   ├── SidebarComponent (filters, actions)
│   │   ├── FilterMenuComponent
│   │   └── ActionsListComponent
│   └── MainComponent (data display)
│       ├── DataTableComponent
│       └── PaginationComponent
└── FooterComponent (status, help)
```

#### Step 2: Create Root App Component

Uses key bindings for app-level keys + message handler for custom messages:

```go
app := bubbly.NewComponent("App").
    WithAutoCommands(true).
    // App-level key bindings
    WithKeyBinding("ctrl+c", "quit", "Quit application").
    WithKeyBinding("?", "toggleHelp", "Show/hide help").
    WithKeyBinding("ctrl+r", "refresh", "Refresh all data").
    // Message handler for complex cases
    WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
            // Handle window resize
            comp.Emit("resize", msg)
            return nil
            
        case tea.MouseMsg:
            // Handle mouse events
            if msg.Type == tea.MouseLeft {
                comp.Emit("click", msg)
            }
            return nil
            
        case CustomDataMsg:
            // Handle custom message from backend
            comp.Emit("dataUpdate", msg.Data)
            return nil
        }
        return nil
    }).
    Setup(func(ctx *Context) {
        // Create child components
        header := createHeaderComponent()
        sidebar := createSidebarComponent()
        main := createMainComponent()
        footer := createFooterComponent()
        
        ctx.AddChild(header)
        ctx.AddChild(sidebar)
        ctx.AddChild(main)
        ctx.AddChild(footer)
        
        ctx.Expose("header", header)
        ctx.Expose("sidebar", sidebar)
        ctx.Expose("main", main)
        ctx.Expose("footer", footer)
        
        // App-level event handlers
        ctx.On("resize", func(data interface{}) {
            // Handle resize
        })
    }).
    Template(func(ctx RenderContext) string {
        // Use PageLayout for professional structure
        layout := components.PageLayout(components.PageLayoutProps{
            Header:  ctx.Get("header").(Component),
            Sidebar: ctx.Get("sidebar").(Component),
            Main:    ctx.Get("main").(Component),
            Footer:  ctx.Get("footer").(Component),
        })
        layout.Init()
        return layout.View()
    }).
    Build()
```

**Benefits**:
- Key bindings for common keys (declarative)
- Message handler for custom types (flexible)
- Layout components for structure (professional)
- Tree structure for organization (Vue-like)

#### Step 3: Create Child Components with Independent Bindings

Each child component has its own key bindings:

```go
// DataTable component - independent key bindings
table := bubbly.NewComponent("DataTable").
    WithAutoCommands(true).
    WithKeyBinding("up", "selectPrevious", "Previous row").
    WithKeyBinding("k", "selectPrevious", "Previous row (vim)").
    WithKeyBinding("down", "selectNext", "Next row").
    WithKeyBinding("j", "selectNext", "Next row (vim)").
    WithKeyBinding("enter", "open", "Open selected").
    WithKeyBinding("d", "delete", "Delete selected").
    Setup(func(ctx *Context) {
        selected := ctx.Ref(0)
        data := ctx.Ref([]Item{})
        
        ctx.On("selectNext", func(_ interface{}) {
            selected.Set(selected.Get().(int) + 1)
            // UI auto-updates!
        })
        
        ctx.On("open", func(_ interface{}) {
            // Open selected item
        })
    }).
    Build()

// Sidebar component - different key bindings
sidebar := bubbly.NewComponent("Sidebar").
    WithAutoCommands(true).
    WithKeyBinding("f", "toggleFilters", "Toggle filters").
    WithKeyBinding("a", "showActions", "Show actions").
    Setup(func(ctx *Context) {
        filtersVisible := ctx.Ref(true)
        
        ctx.On("toggleFilters", func(_ interface{}) {
            filtersVisible.Set(!filtersVisible.Get().(bool))
            // UI auto-updates!
        })
    }).
    Build()
```

**Key Point**: Each component handles its own keys independently - no conflicts!

#### Step 4: Integrate Layout Components

Uses built-in layout components for professional structure:

```go
// Use GridLayout for dashboard cards
dashboard := bubbly.NewComponent("Dashboard").
    WithAutoCommands(true).
    Setup(func(ctx *Context) {
        // Create metric cards
        card1 := createMetricCard("CPU", "45%")
        card2 := createMetricCard("Memory", "2.1GB")
        card3 := createMetricCard("Disk", "128GB")
        card4 := createMetricCard("Network", "1.2MB/s")
        
        ctx.Expose("cards", []Component{card1, card2, card3, card4})
    }).
    Template(func(ctx RenderContext) string {
        cards := ctx.Get("cards").([]Component)
        
        // Use GridLayout for responsive grid
        grid := components.GridLayout(components.GridLayoutProps{
            Columns: 2,
            Gap:     1,
            Items:   cards,
        })
        grid.Init()
        return grid.View()
    }).
    Build()
```

#### Step 5: Run Application

Single line to run entire tree:

```go
func main() {
    app := createAppComponent() // Tree of components
    tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen()).Run()
}
```

**Result**: Complex application with zero boilerplate!

### Metrics

**Component Count**: 10+ nested components
**Lines of Code**:
- Manual approach: ~600 lines (60 per component × 10)
- Declarative approach: ~300 lines (30 per component × 10)
- **Reduction: 50% fewer lines**

**Key Bindings**: 30+ across all components
**Boilerplate Eliminated**:
- Wrapper models: 10 × 40 lines = 400 lines saved
- Keyboard routing: 10 × 20 lines = 200 lines saved
- **Total: 600 lines eliminated**

**Development Time**:
- Manual: 2-3 days (boilerplate + logic)
- Declarative: 1 day (logic only)
- **Speed: 2-3x faster**

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
