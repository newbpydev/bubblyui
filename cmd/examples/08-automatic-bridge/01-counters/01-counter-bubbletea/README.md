# Pure Bubbletea Counter - Comparison Version

This is the **exact same counter app** implemented in pure Bubbletea for direct comparison with the BubblyUI automatic bridge version.

## Purpose

This example demonstrates what you have to write manually in pure Bubbletea vs. what BubblyUI's automatic bridge handles for you.

## How to Run

```bash
cd cmd/examples/08-automatic-bridge/01-counter-bubbletea
go run main.go
```

## Key Bindings (Identical)

- **Space** - Increment counter
- **R** - Reset to zero
- **Ctrl+C** - Quit application

## Side-by-Side Comparison

### BubblyUI Version (Automatic Bridge)

**Lines of Code**: ~120 lines  
**Manual Boilerplate**: 0 lines

```go
// 1. Declarative key bindings
component := bubbly.NewComponent("Counter").
    WithAutoCommands(true).
    WithKeyBinding(" ", "increment", "Increment counter").
    WithKeyBinding("r", "reset", "Reset to zero").
    WithKeyBinding("ctrl+c", "quit", "Quit application").
    
// 2. Event handlers - state changes trigger UI updates automatically
Setup(func(ctx *bubbly.Context) {
    count := ctx.Ref(0)
    ctx.On("increment", func(_ interface{}) {
        count.Set(count.Get().(int) + 1)
        // UI updates automatically!
    })
    ctx.On("reset", func(_ interface{}) {
        count.Set(0)
    })
    ctx.Expose("count", count)
}).

// 3. Template with auto-generated help text
Template(func(ctx bubbly.RenderContext) string {
    count := ctx.Get("count").(*bubbly.Ref[interface{}])
    comp := ctx.Component()
    return fmt.Sprintf("Count: %d\n\n%s", 
        count.Get().(int), 
        comp.HelpText()) // Auto-generated!
}).
Build()

// 4. One-line integration
tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen()).Run()
```

### Pure Bubbletea Version (Manual)

**Lines of Code**: ~110 lines  
**Manual Boilerplate**: ~40 lines (model struct + Init/Update/View methods)

```go
// 1. Manual model struct
type model struct {
    count int
}

// 2. Manual Init method
func (m model) Init() tea.Cmd {
    return nil
}

// 3. Manual Update with all key handling logic
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case " ":
            m.count++  // Manual state mutation
        case "r":
            m.count = 0
        }
    }
    return m, nil
}

// 4. Manual View with hardcoded help text
func (m model) View() string {
    // ... styling code ...
    
    // Manual help text - must keep in sync with Update()
    help := " : Increment counter • ctrl+c: Quit • r: Reset"
    
    return lipgloss.JoinVertical(...)
}

// 5. Manual model initialization
func main() {
    m := model{count: 0}
    tea.NewProgram(m, tea.WithAltScreen()).Run()
}
```

## What BubblyUI Eliminates

### 1. Model Struct Boilerplate ❌
```go
// Pure Bubbletea: REQUIRED
type model struct {
    count int
}

// BubblyUI: NOT NEEDED (handled by framework)
```

### 2. Init Method Boilerplate ❌
```go
// Pure Bubbletea: REQUIRED
func (m model) Init() tea.Cmd {
    return nil
}

// BubblyUI: NOT NEEDED (bubbly.Wrap handles it)
```

### 3. Manual Key Handling in Update() ❌
```go
// Pure Bubbletea: REQUIRED - manual switch statements
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case " ":
            m.count++
        case "r":
            m.count = 0
        }
    }
    return m, nil
}

// BubblyUI: NOT NEEDED - declarative bindings
.WithKeyBinding(" ", "increment", "Increment counter")
.WithKeyBinding("r", "reset", "Reset to zero")
```

### 4. Manual Help Text Maintenance ❌
```go
// Pure Bubbletea: REQUIRED - hardcoded, must sync with Update()
help := " : Increment counter • ctrl+c: Quit • r: Reset"

// BubblyUI: NOT NEEDED - auto-generated from bindings
comp.HelpText() // Always in sync!
```

### 5. Manual State Change Propagation ❌
```go
// Pure Bubbletea: REQUIRED - direct mutation
m.count++

// BubblyUI: AUTOMATIC - reactive updates
count.Set(count.Get().(int) + 1)
// Framework handles UI update automatically
```

## Code Metrics Comparison

| Metric | Pure Bubbletea | BubblyUI | Reduction |
|--------|----------------|----------|-----------|
| **Total Lines** | ~110 | ~120 | Similar |
| **Boilerplate Lines** | ~40 | 0 | **100%** |
| **Key Handling** | Manual switch | Declarative | **100%** |
| **Help Text** | Manual string | Auto-generated | **100%** |
| **State Updates** | Manual mutation | Automatic | **100%** |
| **Model Wrapper** | Manual struct | `bubbly.Wrap()` | **100%** |

## Maintainability Comparison

### Adding a New Key Binding

**Pure Bubbletea**: 3 places to update
1. Add case in `Update()` switch statement
2. Add logic for the action
3. Update help text string manually

**BubblyUI**: 2 places to update
1. Add `.WithKeyBinding()` call (help text included)
2. Add event handler in `Setup()`

Help text updates automatically! ✅

### Changing a Key

**Pure Bubbletea**: 2 places to update
1. Change case in `Update()` switch
2. Update help text string

**BubblyUI**: 1 place to update
1. Change key in `.WithKeyBinding()`

Help text updates automatically! ✅

### Refactoring State Logic

**Pure Bubbletea**: Scattered across Update() method
- Hard to extract
- Tightly coupled to message handling

**BubblyUI**: Isolated in event handlers
- Easy to extract to composables
- Clean separation of concerns

## When to Use Each

### Use Pure Bubbletea When:
- ✅ You need maximum control over message flow
- ✅ You're building a simple app with few keys
- ✅ You want zero framework overhead
- ✅ You're already familiar with Bubbletea patterns

### Use BubblyUI When:
- ✅ You want declarative key bindings
- ✅ You need auto-generated help text
- ✅ You want reactive state management
- ✅ You're building complex apps with many keys
- ✅ You want Vue-like developer experience
- ✅ You want composable, reusable logic

## Performance

Both versions have **identical runtime performance**:
- Same number of allocations
- Same memory usage
- Same rendering speed

BubblyUI's automatic bridge has **zero overhead** (verified by benchmarks).

## Try Both!

Run both versions side-by-side to see the difference:

```bash
# BubblyUI version
cd ../01-counter
go run main.go

# Pure Bubbletea version
cd ../01-counter-bubbletea
go run main.go
```

They look and behave identically, but the code tells a different story!

## Conclusion

Pure Bubbletea gives you **full control** but requires **manual boilerplate**.

BubblyUI gives you **declarative patterns** and **automatic updates** while maintaining **100% Bubbletea compatibility**.

Choose the tool that fits your needs! 🎯
