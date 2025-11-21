# Migration Guide: Manual Setup to bubbly.Run()

## Quick Start

Replace 3 lines of manual Bubbletea setup with 1 line of `bubbly.Run()`:

```go
// Before
p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
if _, err := p.Run(); err != nil { ... }

// After
if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil { ... }
```

**That's it!** For most applications, this is the only change needed.

## Benefits

Migrating to `bubbly.Run()` provides significant improvements:

### Code Reduction
- **Sync apps**: 30-50% less code (3-5 lines → 1 line)
- **Async apps**: 69-82% less code (80-100 lines → 1 line)
- **Example**: `04-async` went from 101 lines to 31 lines (69% reduction)

### Cleaner Code
- ✅ **Zero Bubbletea imports** - only import `bubbly`
- ✅ **Framework-focused** - no internal plumbing exposed
- ✅ **Simpler main()** - clean entry point
- ✅ **Better abstraction** - BubblyUI IS the framework

### Automatic Features
- ✅ **Async auto-detection** - no manual tick wrappers
- ✅ **Error handling** - returns error directly
- ✅ **Type-safe options** - `bubbly.With*()` builders
- ✅ **Best practices** - built-in by default

### Maintainability
- ✅ **Less boilerplate** - focus on business logic
- ✅ **Fewer bugs** - less code = fewer places for errors
- ✅ **Easier onboarding** - simpler patterns for new developers
- ✅ **Future-proof** - framework handles internal changes

## Step-by-Step Migration

### Step 1: Remove Bubbletea Import

**Before:**
```go
import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"  // ❌ Remove this
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)
```

**After:**
```go
import (
    "fmt"
    "os"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly"  // ✅ Only this
)
```

**Note:** Keep the `tea` import only if you have custom `tea.Msg` types. Otherwise, remove it completely.

### Step 2: Replace tea.NewProgram with bubbly.Run

**Before:**
```go
func main() {
    app, _ := CreateApp()
    
    p := tea.NewProgram(
        bubbly.Wrap(app),
        tea.WithAltScreen(),
    )
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**After:**
```go
func main() {
    app, _ := CreateApp()
    
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Changes:**
1. Remove `tea.NewProgram()` call
2. Remove `bubbly.Wrap()` call
3. Replace with `bubbly.Run(app, ...)`
4. Change `tea.WithAltScreen()` to `bubbly.WithAltScreen()`

### Step 3: Convert Options

All Bubbletea options have equivalent `bubbly.With*()` builders:

| Bubbletea Option | BubblyUI Option |
|-----------------|-----------------|
| `tea.WithAltScreen()` | `bubbly.WithAltScreen()` |
| `tea.WithMouseAllMotion()` | `bubbly.WithMouseAllMotion()` |
| `tea.WithMouseCellMotion()` | `bubbly.WithMouseCellMotion()` |
| `tea.WithFPS(60)` | `bubbly.WithFPS(60)` |
| `tea.WithInput(r)` | `bubbly.WithInput(r)` |
| `tea.WithOutput(w)` | `bubbly.WithOutput(w)` |
| `tea.WithContext(ctx)` | `bubbly.WithContext(ctx)` |
| `tea.WithoutBracketedPaste()` | `bubbly.WithoutBracketedPaste()` |
| `tea.WithoutSignalHandler()` | `bubbly.WithoutSignalHandler()` |
| `tea.WithoutCatchPanics()` | `bubbly.WithoutCatchPanics()` |
| `tea.WithReportFocus()` | `bubbly.WithReportFocus()` |
| `tea.WithInputTTY()` | `bubbly.WithInputTTY()` |
| `tea.WithEnvironment(env)` | `bubbly.WithEnvironment(env)` |

**Example:**

**Before:**
```go
p := tea.NewProgram(
    bubbly.Wrap(app),
    tea.WithAltScreen(),
    tea.WithMouseAllMotion(),
    tea.WithFPS(120),
)
```

**After:**
```go
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithMouseAllMotion(),
    bubbly.WithFPS(120),
)
```

### Step 4: Remove Tick Wrapper (Async Apps)

This is the biggest win! If your app uses async operations with a manual tick wrapper, you can remove 80+ lines of boilerplate.

**Before (101 lines):**
```go
package main

import (
    "fmt"
    "os"
    "time"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Manual tick message
type tickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

// Manual wrapper model
type model struct {
    component bubbly.Component
    loading   bool
}

func (m model) Init() tea.Cmd {
    return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "r":
            m.component.Emit("refresh", nil)
            m.loading = true
            cmds = append(cmds, tickCmd())
        }
    case tickMsg:
        if m.loading {
            cmds = append(cmds, tickCmd())
        }
    }
    
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

func main() {
    app, _ := CreateAsyncApp()
    
    m := model{
        component: app,
        loading:   true,
    }
    
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**After (31 lines - 69% reduction!):**
```go
package main

import (
    "fmt"
    "os"
    "time"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    app, _ := CreateAsyncApp()
    
    // Async auto-detected - no tick wrapper needed!
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Requirements:**
Your component must use `WithAutoCommands(true)`:

```go
func CreateAsyncApp() (bubbly.Component, error) {
    return bubbly.NewComponent("AsyncApp").
        WithAutoCommands(true).  // ← This enables auto-detection
        Setup(func(ctx *bubbly.Context) {
            // Async operations here
            data := composables.UseAsync(ctx, fetchData)
            ctx.OnMounted(func() {
                data.Execute()
            })
        }).
        Build()
}
```

### Step 5: Test

After migration, test your application:

1. **Build**: `go build`
2. **Run**: Execute your application
3. **Verify**: All functionality works as before
4. **Check**: No Bubbletea imports (except custom messages)

## Migration Strategies

### Immediate Migration (Recommended)

For simple applications, migrate everything at once:

1. Remove `tea` import
2. Replace `tea.NewProgram` with `bubbly.Run`
3. Convert all options
4. Remove tick wrapper (if async)
5. Test

**Best for:**
- New applications
- Simple applications
- Applications with good test coverage

### Gradual Migration

For complex applications, migrate incrementally:

**Phase 1: New Code**
- Use `bubbly.Run()` for all new components
- Keep existing code with manual setup

**Phase 2: Simple Components**
- Migrate simple sync components first
- Test each migration

**Phase 3: Async Components**
- Migrate async components (biggest wins)
- Remove tick wrappers

**Phase 4: Cleanup**
- Remove unused `tea` imports
- Update documentation

**Best for:**
- Large applications
- Production applications
- Applications with limited test coverage

### Coexistence Strategy

You can mix both patterns during migration:

```go
// Old component (manual setup)
oldApp := CreateOldApp()
p := tea.NewProgram(bubbly.Wrap(oldApp), tea.WithAltScreen())

// New component (bubbly.Run)
newApp := CreateNewApp()
bubbly.Run(newApp, bubbly.WithAltScreen())
```

This allows you to:
- Migrate at your own pace
- Test each component individually
- Maintain backward compatibility

## Backward Compatibility

`bubbly.Run()` is **100% backward compatible**. The old pattern still works:

```go
// Old pattern - still works!
p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

**No breaking changes:**
- `bubbly.Wrap()` still exists
- All existing code continues to work
- Migration is optional (but recommended)

## Common Patterns

### Pattern 1: Simple Sync App

**Before:**
```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    app, _ := CreateApp()
    p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

**After:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly"

func main() {
    app, _ := CreateApp()
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        log.Fatal(err)
    }
}
```

**Changes:**
- Removed `tea` import
- Replaced 3 lines with 1 line
- Same functionality

### Pattern 2: Async App with Tick Wrapper

**Before:**
```go
import (
    "time"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

type model struct {
    component bubbly.Component
    loading   bool
}

func (m model) Init() tea.Cmd {
    return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "r":
            m.component.Emit("refresh", nil)
            m.loading = true
            cmds = append(cmds, tickCmd())
        }
    case tickMsg:
        if m.loading {
            cmds = append(cmds, tickCmd())
        }
    }
    
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

func main() {
    app, _ := CreateAsyncApp()
    m := model{component: app, loading: true}
    p := tea.NewProgram(m, tea.WithAltScreen())
    p.Run()
}
```

**After:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly"

func main() {
    app, _ := CreateAsyncApp()
    bubbly.Run(app, bubbly.WithAltScreen())
}
```

**Changes:**
- Removed entire tick wrapper (60+ lines)
- Removed `tea` import
- Async auto-detected from `WithAutoCommands(true)`

### Pattern 3: Manual Key Routing

**Before:**
```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

type model struct {
    counter bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.counter.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up":
            m.counter.Emit("increment", nil)
        case "down":
            m.counter.Emit("decrement", nil)
        case "r":
            m.counter.Emit("reset", nil)
        }
    }
    
    updatedComponent, cmd := m.counter.Update(msg)
    m.counter = updatedComponent.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.counter.View()
}

func main() {
    counter, _ := CreateCounter()
    m := model{counter: counter}
    p := tea.NewProgram(m, tea.WithAltScreen())
    p.Run()
}
```

**After:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly"

func CreateCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        WithKeyBinding("up", "increment", "Increment").
        WithKeyBinding("down", "decrement", "Decrement").
        WithKeyBinding("r", "reset", "Reset").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubbly.Context) {
            // ... setup code
        }).
        Build()
}

func main() {
    counter, _ := CreateCounter()
    bubbly.Run(counter, bubbly.WithAltScreen())
}
```

**Changes:**
- Removed manual wrapper model (40+ lines)
- Added declarative key bindings
- Auto-generated help text
- Removed `tea` import

### Pattern 4: Multiple Options

**Before:**
```go
import (
    "context"
    "time"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    app, _ := CreateApp()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    p := tea.NewProgram(
        bubbly.Wrap(app),
        tea.WithAltScreen(),
        tea.WithMouseAllMotion(),
        tea.WithFPS(120),
        tea.WithContext(ctx),
    )
    
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

**After:**
```go
import (
    "context"
    "time"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    app, _ := CreateApp()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    if err := bubbly.Run(app,
        bubbly.WithAltScreen(),
        bubbly.WithMouseAllMotion(),
        bubbly.WithFPS(120),
        bubbly.WithContext(ctx),
    ); err != nil {
        log.Fatal(err)
    }
}
```

**Changes:**
- Removed `tea` import
- Changed `tea.With*()` to `bubbly.With*()`
- Cleaner option passing

## Troubleshooting

### "undefined: tea" errors

**Problem:** After removing `tea` import, you get compilation errors.

**Solution:** Check for remaining `tea.*` references:
- `tea.KeyMsg` → Keep `tea` import if you have custom messages
- `tea.Msg` → Keep `tea` import if you have custom messages
- `tea.Cmd` → Should not be in your code (use component events)
- `tea.Model` → Should not be in your code (use components)

### Async app not updating

**Problem:** After migration, async operations don't update the UI.

**Solution:** Ensure your component has `WithAutoCommands(true)`:

```go
app := bubbly.NewComponent("App").
    WithAutoCommands(true).  // ← Add this!
    Setup(func(ctx *bubbly.Context) {
        // Async operations
    }).
    Build()
```

### Custom tick interval needed

**Problem:** Default 100ms interval is too slow/fast.

**Solution:** Use `WithAsyncRefresh()`:

```go
// Faster updates (50ms = 20 updates/sec)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(50*time.Millisecond),
)

// Slower updates (200ms = 5 updates/sec)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(200*time.Millisecond),
)
```

### Need direct tea.Program access

**Problem:** You need to call methods on `tea.Program`.

**Solution:** Use the manual pattern for this specific case:

```go
p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
p.Send(customMsg)  // Direct program access
p.Run()
```

However, consider if you really need this. Most use cases are covered by:
- Component events (`component.Emit()`)
- Lifecycle hooks (`ctx.OnMounted()`)
- Composables (`UseAsync`, `UseInterval`)

## Examples

### Example 1: Todo App Migration

**Before (45 lines):**
```go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

type model struct {
    app bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.app.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "a":
            m.app.Emit("add", nil)
        case "d":
            m.app.Emit("delete", nil)
        }
    }
    
    updated, cmd := m.app.Update(msg)
    m.app = updated.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.app.View()
}

func main() {
    app, _ := CreateTodoApp()
    m := model{app: app}
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**After (20 lines - 56% reduction):**
```go
package main

import (
    "fmt"
    "os"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func CreateTodoApp() (bubbly.Component, error) {
    return bubbly.NewComponent("TodoApp").
        WithKeyBinding("a", "add", "Add todo").
        WithKeyBinding("d", "delete", "Delete todo").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubbly.Context) {
            // ... setup code
        }).
        Build()
}

func main() {
    app, _ := CreateTodoApp()
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Example 2: Async Dashboard Migration

**Before (97 lines):**
```go
package main

import (
    "fmt"
    "os"
    "time"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

type model struct {
    component bubbly.Component
    loading   bool
}

func (m model) Init() tea.Cmd {
    return tea.Batch(m.component.Init(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "r":
            m.component.Emit("refresh", nil)
            m.loading = true
            cmds = append(cmds, tickCmd())
        }
    case tickMsg:
        if m.loading {
            cmds = append(cmds, tickCmd())
        }
    }
    
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

func fetchMetrics() (*Metrics, error) {
    time.Sleep(2 * time.Second)
    return &Metrics{CPU: 45, Memory: 60}, nil
}

func CreateDashboard() (bubbly.Component, error) {
    return bubbly.NewComponent("Dashboard").
        Setup(func(ctx *bubbly.Context) {
            metrics := composables.UseAsync(ctx, fetchMetrics)
            
            ctx.OnMounted(func() {
                metrics.Execute()
            })
            
            ctx.On("refresh", func(interface{}) {
                metrics.Execute()
            })
            
            ctx.Expose("metrics", metrics)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            metrics := ctx.Get("metrics").(*composables.AsyncState[*Metrics])
            // ... render logic
        }).
        Build()
}

func main() {
    dashboard, _ := CreateDashboard()
    m := model{component: dashboard, loading: true}
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**After (30 lines - 69% reduction):**
```go
package main

import (
    "fmt"
    "os"
    "time"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

func fetchMetrics() (*Metrics, error) {
    time.Sleep(2 * time.Second)
    return &Metrics{CPU: 45, Memory: 60}, nil
}

func CreateDashboard() (bubbly.Component, error) {
    return bubbly.NewComponent("Dashboard").
        WithAutoCommands(true).  // ← Enable async auto-detection
        WithKeyBinding("r", "refresh", "Refresh metrics").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubbly.Context) {
            metrics := composables.UseAsync(ctx, fetchMetrics)
            
            ctx.OnMounted(func() {
                metrics.Execute()
            })
            
            ctx.On("refresh", func(interface{}) {
                metrics.Execute()
            })
            
            ctx.Expose("metrics", metrics)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            metrics := ctx.Get("metrics").(*composables.AsyncState[*Metrics])
            comp := ctx.Component()
            // ... render logic with comp.HelpText()
        }).
        Build()
}

func main() {
    dashboard, _ := CreateDashboard()
    if err := bubbly.Run(dashboard, bubbly.WithAltScreen()); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Migration Checklist

Use this checklist to ensure complete migration:

- [ ] Remove `tea` import (keep only if custom messages)
- [ ] Replace `tea.NewProgram()` with `bubbly.Run()`
- [ ] Remove `bubbly.Wrap()` call
- [ ] Convert all `tea.With*()` to `bubbly.With*()`
- [ ] Remove tick wrapper (async apps)
- [ ] Add `WithAutoCommands(true)` (async apps)
- [ ] Replace manual key routing with `WithKeyBinding()`
- [ ] Remove manual wrapper models
- [ ] Test application functionality
- [ ] Verify no Bubbletea imports remain
- [ ] Update documentation
- [ ] Update tests

## See Also

- **[Framework Run API](../features/framework-run-api.md)** - Complete API documentation
- **[Component Model](../../specs/02-component-model/)** - Understanding components
- **[Composition API](../../specs/04-composition-api/)** - Advanced patterns
- **[Automatic Bridge](../../specs/08-automatic-reactive-bridge/)** - Auto commands and key bindings
