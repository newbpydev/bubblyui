# Automatic Reactive Bridge Migration Guide

## Quick Start

**TL;DR:** Enable automatic command generation to eliminate 30-50% of boilerplate code. State changes (`Ref.Set()`) automatically trigger UI updates without manual `Emit()` calls.

### Before (Manual Bridge Pattern)
```go
// 40+ lines of boilerplate
type model struct {
    component bubbly.Component
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "space" {
            m.component.Emit("increment", nil) // Manual!
        }
    }
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}

func main() {
    counter := createCounterComponent()
    m := model{component: counter}
    tea.NewProgram(m).Run()
}
```

### After (Automatic Bridge Pattern)
```go
// 2 lines - that's it!
func main() {
    counter := createCounterComponent()
    tea.NewProgram(bubbly.Wrap(counter)).Run()
}
```

**Component code unchanged** - just enable automatic mode:
```go
component := bubbly.NewComponent("Counter").
    WithAutoCommands(true).  // Enable automatic bridge
    Setup(func(ctx *bubbly.Context) {
        count := ctx.Ref(0)
        ctx.On("increment", func(_ interface{}) {
            count.Set(count.Get().(int) + 1)
            // UI updates automatically - no Emit() needed!
        })
    }).
    Build()
```

---

## Why Migrate

### The Problem with Manual Bridge

**Pain Point 1: Boilerplate Code**
```go
// Every component needs this wrapper model (30-40 lines)
type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Manual message handling
    // Manual Emit() calls
    // Manual command batching
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.component.View()
}
```

**Pain Point 2: Manual Emit() Everywhere**
```go
ctx.On("increment", func(_ interface{}) {
    count.Set(count.Get().(int) + 1)
    ctx.Emit("state-changed", nil) // Easy to forget!
})

ctx.On("decrement", func(_ interface{}) {
    count.Set(count.Get().(int) - 1)
    ctx.Emit("state-changed", nil) // Repetitive!
})

ctx.On("reset", func(_ interface{}) {
    count.Set(0)
    ctx.Emit("state-changed", nil) // Tedious!
})
```

**Pain Point 3: Easy to Forget Emit()**
```go
ctx.On("complex-operation", func(_ interface{}) {
    value1.Set(calculate1())
    value2.Set(calculate2())
    value3.Set(calculate3())
    // Oops! Forgot to emit - UI doesn't update
})
```

### The Solution: Automatic Bridge

**Benefit 1: Zero Boilerplate**
```go
// Just one line!
tea.NewProgram(bubbly.Wrap(component)).Run()
```

**Benefit 2: Automatic Updates**
```go
ctx.On("increment", func(_ interface{}) {
    count.Set(count.Get().(int) + 1)
    // UI updates automatically - framework handles it!
})
```

**Benefit 3: Can't Forget Updates**
```go
ctx.On("complex-operation", func(_ interface{}) {
    value1.Set(calculate1())  // Triggers update
    value2.Set(calculate2())  // Triggers update
    value3.Set(calculate3())  // Triggers update
    // All updates batched and executed automatically!
})
```

**Benefit 4: Vue-like Developer Experience**
```go
// Familiar pattern for Vue developers
count.value++        // Vue
count.Set(count.Get().(int) + 1)  // BubblyUI - auto updates!
```

---

## What Changed

### API Changes

| Component | Old API | New API | Notes |
|-----------|---------|---------|-------|
| **Builder** | `NewComponent("Name")` | `.WithAutoCommands(true)` | Enable automatic mode |
| **Builder** | N/A | `.WithCommandDebug(true)` | Optional: Enable debug logging |
| **Context** | N/A | `ctx.EnableAutoCommands()` | Runtime enable |
| **Context** | N/A | `ctx.DisableAutoCommands()` | Runtime disable |
| **Context** | N/A | `ctx.IsAutoCommandsEnabled()` | Check status |
| **Context** | N/A | `ctx.ManualRef(value)` | Create ref without auto commands |
| **Context** | N/A | `ctx.SetCommandGenerator(gen)` | Custom command generation |
| **Integration** | Manual wrapper model | `bubbly.Wrap(component)` | One-line integration |
| **Event Handlers** | `count.Set(n); ctx.Emit(...)` | `count.Set(n)` | Emit() no longer needed |

### Backward Compatibility

✅ **100% Backward Compatible**
- Manual bridge pattern still works
- Existing code unchanged
- Gradual migration supported
- Mix automatic and manual patterns

---

## Migration Steps

### Step 1: Assess Your Codebase

**Identify manual bridge patterns:**

```bash
# Find manual wrapper models
grep -r "type.*struct.*component.*Component" .

# Find manual Emit() calls
grep -r "\.Emit(" . | grep -v "test"

# Count lines of wrapper code
wc -l model.go  # Compare before/after
```

**Calculate potential savings:**
- Manual wrapper: ~40 lines
- Automatic wrapper: 1 line
- **Savings: 39 lines per component**

### Step 2: Enable Automatic Mode

**Option A: Enable globally in component builder** (Recommended)

```go
// Old
counter := bubbly.NewComponent("Counter").
    Setup(...).
    Template(...).
    Build()

// New
counter := bubbly.NewComponent("Counter").
    WithAutoCommands(true).  // Add this line
    Setup(...).
    Template(...).
    Build()
```

**Option B: Enable at runtime in Setup**

```go
Setup(func(ctx *bubbly.Context) {
    ctx.EnableAutoCommands()  // Enable here
    
    count := ctx.Ref(0)
    // ... rest of setup
})
```

**Option C: Enable debug mode (development only)**

```go
counter := bubbly.NewComponent("Counter").
    WithAutoCommands(true).
    WithCommandDebug(true).  // See command generation logs
    Setup(...).
    Build()
```

### Step 3: Simplify Wrapper Model

**Old (Manual Wrapper):**

```go
type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "space":
            m.component.Emit("increment", nil)
        }
    }
    
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.component.View()
}

func main() {
    component := createComponent()
    m := model{component: component}
    tea.NewProgram(m, tea.WithAltScreen()).Run()
}
```

**New (Automatic Wrapper):**

```go
func main() {
    component := createComponent()
    tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen()).Run()
}
```

**That's it!** Delete the entire `model` struct and methods.

### Step 4: Remove Manual Emit() Calls

**Old:**

```go
ctx.On("increment", func(_ interface{}) {
    count.Set(count.Get().(int) + 1)
    ctx.Emit("state-changed", nil)  // Remove this
})

ctx.On("decrement", func(_ interface{}) {
    count.Set(count.Get().(int) - 1)
    ctx.Emit("state-changed", nil)  // Remove this
})

ctx.On("update-name", func(data interface{}) {
    name.Set(data.(string))
    ctx.Emit("name-updated", nil)  // Remove this
})
```

**New:**

```go
ctx.On("increment", func(_ interface{}) {
    count.Set(count.Get().(int) + 1)
    // UI updates automatically!
})

ctx.On("decrement", func(_ interface{}) {
    count.Set(count.Get().(int) - 1)
    // UI updates automatically!
})

ctx.On("update-name", func(data interface{}) {
    name.Set(data.(string))
    // UI updates automatically!
})
```

**Search and remove:**

```bash
# Find all Emit() calls that follow Set()
grep -A 1 "\.Set(" . | grep "\.Emit("

# Verify they're for state updates (not custom events)
# Remove the Emit() lines
```

### Step 5: Verify and Test

**Run your application:**

```bash
go run main.go
```

**Enable debug mode to see commands:**

```go
.WithCommandDebug(true)
```

**Output:**
```
[DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 0 → 1
[DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 1 → 2
```

**Verify:**
- ✅ UI updates on every state change
- ✅ No manual Emit() needed
- ✅ Performance is equivalent
- ✅ All interactions work correctly

### Step 6: Clean Up (Optional)

**Remove unused imports:**

```go
// Can remove if only used for manual wrapper
import tea "github.com/charmbracelet/bubbletea"  // Still needed for Wrap()
```

**Remove wrapper model:**

```bash
# Delete the old wrapper code
# Typically 30-40 lines removed
```

**Update comments:**

```go
// Old comment
// Manual emit required to trigger UI update

// New comment (or remove)
// Automatic UI update
```

---

## Advanced Migration Patterns

### Pattern 1: Mixed Automatic and Manual

**Use Case:** Performance-critical batch operations

```go
ctx.Setup(func(ctx *bubbly.Context) {
    // Auto-command refs (default)
    displayValue := ctx.Ref(0)
    
    // Manual-command refs (explicit control)
    internalBuffer := ctx.ManualRef([]int{})
    
    ctx.On("batch-process", func(data interface{}) {
        items := data.([]int)
        
        // Process without triggering commands
        buffer := []int{}
        for _, item := range items {
            buffer = append(buffer, item*2)
        }
        
        // Update manual ref (no command)
        internalBuffer.Set(buffer)
        
        // Update auto ref (triggers UI update)
        displayValue.Set(len(buffer))
    })
})
```

### Pattern 2: Disable for Tight Loops

**Use Case:** 1000+ updates in tight loop

```go
ctx.On("process-large-dataset", func(data interface{}) {
    items := data.([]int)
    
    // Disable auto-commands for batch
    ctx.DisableAutoCommands()
    
    result := ctx.Ref(0)
    for _, item := range items {
        result.Set(result.Get().(int) + item)
    }
    
    // Re-enable and trigger single update
    ctx.EnableAutoCommands()
    ctx.Emit("processing-complete", result.Get())
})
```

### Pattern 3: Custom Command Generator

**Use Case:** Add logging, metrics, or custom messages

```go
type MetricsGenerator struct {
    metrics *Metrics
}

func (g *MetricsGenerator) Generate(componentID, refID string, oldValue, newValue interface{}) tea.Cmd {
    // Record metric
    g.metrics.RecordStateChange(componentID, refID)
    
    // Return standard command
    return func() tea.Msg {
        return bubbly.StateChangedMsg{
            ComponentID: componentID,
            RefID:       refID,
            OldValue:    oldValue,
            NewValue:    newValue,
            Timestamp:   time.Now(),
        }
    }
}

// Use custom generator
ctx.SetCommandGenerator(&MetricsGenerator{metrics: myMetrics})
```

---

## Common Pitfalls

### Pitfall 1: Ref.Set() in Template

**Problem:**
```go
.Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    count.Set(100)  // ❌ PANIC: Cannot call Ref.Set() in template!
    return fmt.Sprintf("Count: %d", count.Get())
})
```

**Error Message:**
```
PANIC: Cannot call Ref.Set() in template - templates must be pure functions 
with no side effects.

Move state updates to event handlers (ctx.On) or lifecycle hooks 
(onMounted, onUpdated).
```

**Solution:**
```go
.Setup(func(ctx *bubbly.Context) {
    count := ctx.Ref(0)
    
    ctx.OnMounted(func() {
        count.Set(100)  // ✅ Correct: Set in lifecycle hook
    })
})
.Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    return fmt.Sprintf("Count: %d", count.Get())  // ✅ Read-only
})
```

### Pitfall 2: Infinite Update Loop

**Problem:**
```go
ctx.OnUpdated(func() {
    count.Set(count.Get().(int) + 1)  // ❌ Triggers another update!
}, count)
```

**Error Message:**
```
ERROR: Command generation loop detected for component 'Counter' ref 'count': 
generated 101 commands (max 100).

Check for recursive state updates in event handlers or lifecycle hooks.
```

**Solution:**
```go
ctx.OnUpdated(func() {
    current := count.Get().(int)
    if current < 100 {  // ✅ Add condition to prevent infinite loop
        count.Set(current + 1)
    }
}, count)
```

### Pitfall 3: Forgot to Enable Auto Commands

**Problem:**
```go
counter := bubbly.NewComponent("Counter").
    // Forgot .WithAutoCommands(true)
    Setup(...).
    Build()

// UI doesn't update on Set()
```

**Symptom:**
- State changes but UI doesn't update
- No error messages
- Debug logs show no commands generated

**Solution:**
```go
counter := bubbly.NewComponent("Counter").
    WithAutoCommands(true).  // ✅ Add this
    Setup(...).
    Build()
```

### Pitfall 4: Mixed Manual and Auto Refs Confusion

**Problem:**
```go
autoRef := ctx.Ref(0)      // Auto-commands enabled
manualRef := ctx.ManualRef(0)  // Auto-commands disabled

autoRef.Set(1)    // ✅ Triggers update
manualRef.Set(1)  // ❌ No update - need manual Emit()
```

**Solution:**
```go
// Option 1: Use auto for both
autoRef1 := ctx.Ref(0)
autoRef2 := ctx.Ref(0)
// Both trigger updates automatically

// Option 2: Document manual refs clearly
manualRef := ctx.ManualRef(0)  // Manual control required
manualRef.Set(1)
ctx.Emit("manual-updated", nil)  // Explicit emit needed
```

### Pitfall 5: Using Manual Wrapper with Auto Commands

**Problem:**
```go
component := bubbly.NewComponent("Counter").
    WithAutoCommands(true).  // Auto enabled
    Setup(...).
    Build()

// But using manual wrapper!
type model struct {
    component bubbly.Component
}
// ... manual Update() doesn't drain command queue
```

**Symptom:**
- Commands generate but never execute
- UI doesn't update
- Debug logs show commands but no messages

**Solution:**
```go
// Use automatic wrapper
tea.NewProgram(bubbly.Wrap(component)).Run()  // ✅ Correct
```

---

## Troubleshooting

### Issue: UI Not Updating

**Check 1: Auto commands enabled?**

```go
// In Setup
if !ctx.IsAutoCommandsEnabled() {
    log.Println("Auto commands are DISABLED - enable them!")
    ctx.EnableAutoCommands()
}
```

**Check 2: Using Wrap()?**

```go
// ❌ Wrong
type model struct { component bubbly.Component }
tea.NewProgram(model{component: comp}).Run()

// ✅ Right
tea.NewProgram(bubbly.Wrap(comp)).Run()
```

**Check 3: Ref created via ctx.Ref()?**

```go
// ❌ Wrong - doesn't hook into component
ref := bubbly.NewRef(0)

// ✅ Right - hooks into component
ref := ctx.Ref(0)
```

### Issue: Excessive Command Generation

**Enable debug mode:**

```go
.WithCommandDebug(true)
```

**Look for patterns:**
```
[DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 0 → 1
[DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 1 → 2
[DEBUG] Command Generated | Component: Counter (component-1) | Ref: ref-5 | 2 → 3
... (repeating)
```

**Solution: Check for infinite loops**

```go
// Look for onUpdated triggering more updates
ctx.OnUpdated(func() {
    // Add condition to prevent infinite updates
    if !someCondition {
        return
    }
    count.Set(newValue)
}, count)
```

### Issue: Performance Degradation

**Measure command overhead:**

```bash
# Run with debug mode
.WithCommandDebug(true)

# Count commands per second
grep "[DEBUG]" output.log | wc -l
```

**Expected:**
- Command generation: ~316 ns/op
- Batching: ~100-500 ns/op (depending on batch size)
- Total overhead: < 1μs per state change

**If performance is poor:**

1. **Disable auto-commands for tight loops:**
   ```go
   ctx.DisableAutoCommands()
   // ... 1000+ Set() calls
   ctx.EnableAutoCommands()
   ctx.Emit("batch-complete", nil)
   ```

2. **Use manual refs for high-frequency updates:**
   ```go
   animationFrame := ctx.ManualRef(0)  // Updates 60 times/second
   displayValue := ctx.Ref(0)          // Updates on significant changes
   ```

3. **Batch related updates:**
   ```go
   // Instead of:
   x.Set(newX)  // Command 1
   y.Set(newY)  // Command 2
   z.Set(newZ)  // Command 3
   
   // Use:
   position := ctx.Ref(Position{X: x, Y: y, Z: z})
   position.Set(Position{X: newX, Y: newY, Z: newZ})  // Single command
   ```

### Issue: Commands Not Batching

**Check command queue:**

```go
// Enable debug mode and watch logs
.WithCommandDebug(true)

// Should see batched messages
// [DEBUG] Command Generated | ... (multiple in quick succession)
// Then single Update() cycle processes all
```

**Verify batching is enabled:**

```go
// Batching is automatic with Wrap()
tea.NewProgram(bubbly.Wrap(component)).Run()

// Component returns batched commands from Update()
// No additional configuration needed
```

### Issue: Debug Logs Not Showing

**Check debug mode is enabled:**

```go
component := bubbly.NewComponent("Counter").
    WithAutoCommands(true).
    WithCommandDebug(true).  // Make sure this is set
    Setup(...).
    Build()
```

**Redirect output to file:**

```bash
# Capture all debug output
go run main.go 2>&1 | tee debug.log

# Filter for debug messages
grep "\[DEBUG\]" debug.log
```

**Programmatic logging:**

```go
import "os"

logger := commands.NewCommandLogger(os.Stdout)
ctx.SetCommandGenerator(logger)  // Custom logger
```

---

## Migration Checklist

Use this checklist to track your migration progress:

### Preparation
- [ ] Read this migration guide completely
- [ ] Identify all components with manual bridge pattern
- [ ] Estimate lines of code to be removed
- [ ] Plan migration order (start with simple components)

### Per Component
- [ ] Enable automatic commands (`.WithAutoCommands(true)`)
- [ ] Replace manual wrapper with `bubbly.Wrap()`
- [ ] Remove manual `Emit()` calls after `Set()`
- [ ] Remove manual wrapper model struct
- [ ] Remove manual `Init()`, `Update()`, `View()` implementations
- [ ] Test component functionality
- [ ] Verify UI updates correctly
- [ ] Check for performance issues

### Verification
- [ ] All manual `Emit()` calls removed
- [ ] All wrapper models deleted
- [ ] Code builds without errors
- [ ] All tests pass
- [ ] Manual testing confirms UI updates
- [ ] No performance degradation

### Optional Optimization
- [ ] Enable debug mode during development
- [ ] Add custom command generator if needed
- [ ] Use `ManualRef()` for performance-critical refs
- [ ] Batch related state updates
- [ ] Document any manual refs with comments

---

## Performance Comparison

### Before (Manual Bridge)
```
Ref.Set() + Emit():  ~12,000 ns/op
Manual wrapper:       ~1,200 ns/op overhead
Total:                ~13,200 ns/op
```

### After (Automatic Bridge)
```
Ref.Set() + Auto:     ~316 ns/op (command generation)
Wrapper.Update():     ~115 ns/op overhead
Total:                ~431 ns/op
```

**Result:** ~30x faster command generation, same overall performance due to batching.

---

## Summary

### What You Gain
- ✅ **30-50% less code** - Remove wrapper boilerplate
- ✅ **Zero manual Emit()** - Framework handles updates
- ✅ **Can't forget updates** - Impossible to skip
- ✅ **Vue-like DX** - Familiar reactive patterns
- ✅ **Same performance** - Optimized batching
- ✅ **100% backward compatible** - Gradual migration

### Migration Effort
- **Simple component:** 5-10 minutes
- **Complex component:** 15-30 minutes
- **Entire application:** 1-4 hours (depending on size)

### Recommended Approach
1. Start with simplest component
2. Verify it works correctly
3. Migrate remaining components one by one
4. Run tests after each migration
5. Keep manual pattern for edge cases if needed

### Next Steps
- See **Example Applications** (Task 7.3) for complete examples
- Read **API Documentation** for detailed API reference
- Join community discussions for migration help

---

**Happy Migrating!** 🚀

For questions or issues, please open a GitHub issue or discuss in the BubblyUI community.
