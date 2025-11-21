# Framework Run API

## Overview

`bubbly.Run()` is the recommended way to launch BubblyUI applications, eliminating all Bubbletea boilerplate from your code. This single function call replaces the manual `tea.NewProgram()` setup, making your `main.go` clean and framework-focused.

## Philosophy

**BubblyUI IS the framework.** Bubbletea is an internal dependency, not something users should interact with directly. The `bubbly.Run()` API embodies this philosophy by:

- **Zero Bubbletea Imports**: Your application code only imports `bubbly`
- **Framework-Level Abstraction**: All Bubbletea details are handled internally
- **Best Practices Built-In**: Async detection, error handling, and configuration are automatic
- **Clean Separation**: Your code focuses on components, not plumbing

## The Problem

Before `bubbly.Run()`, launching a BubblyUI application required manual Bubbletea setup:

```go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"  // ❌ Bubbletea import
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    app, _ := CreateApp()
    
    // ❌ Manual Bubbletea program creation
    p := tea.NewProgram(
        bubbly.Wrap(app),
        tea.WithAltScreen(),
    )
    
    // ❌ Manual error handling
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Issues:**
- ❌ Requires importing Bubbletea (`tea` package)
- ❌ Exposes internal framework details (`Wrap()`, `tea.NewProgram()`)
- ❌ Verbose (3+ lines just to run the app)
- ❌ Async apps require 80+ lines of manual tick wrapper boilerplate

## The Solution

`bubbly.Run()` simplifies everything to a single line:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/newbpydev/bubblyui/pkg/bubbly"  // ✅ Only bubbly import
)

func main() {
    app, _ := CreateApp()
    
    // ✅ One-line framework launch
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Benefits:**
- ✅ **Zero Bubbletea imports** - framework-focused code
- ✅ **One line to run** - minimal boilerplate
- ✅ **Async auto-detection** - no manual tick wrappers needed
- ✅ **Clean error handling** - returns error directly
- ✅ **Type-safe options** - `bubbly.With*()` option builders

## Basic Usage

### Minimal Example

The simplest possible BubblyUI application:

```go
package main

import "github.com/newbpydev/bubblyui/pkg/bubbly"

func main() {
    app, _ := bubbly.NewComponent("App").
        Template(func(ctx bubbly.RenderContext) string {
            return "Hello, BubblyUI!"
        }).
        Build()
    
    bubbly.Run(app)  // That's it!
}
```

### With Alt Screen (Full Screen Mode)

Most TUI applications use the alternate screen buffer:

```go
func main() {
    app, _ := CreateApp()
    
    // Run in full-screen mode
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        log.Fatal(err)
    }
}
```

### With Error Handling

Production-ready error handling:

```go
func main() {
    app, err := CreateApp()
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
    }
    
    if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}
```

## Run Options

`bubbly.Run()` accepts a variadic list of `RunOption` functions that configure the application. All Bubbletea program options are available through `bubbly.With*()` builders.

### Display Options

#### `WithAltScreen()`

Enables the alternate screen buffer (full-screen mode). This is the most common option for TUI applications.

```go
bubbly.Run(app, bubbly.WithAltScreen())
```

**Use when:** Building full-screen TUI applications (most cases)

#### `WithFPS(fps int)`

Sets the target frames per second for rendering. Default is 60 FPS.

```go
// High-performance dashboard with 120 FPS
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithFPS(120),
)
```

**Use when:** You need smoother animations or want to reduce CPU usage

### Mouse Options

#### `WithMouseAllMotion()`

Enables mouse support with all motion events (captures all mouse movements).

```go
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithMouseAllMotion(),
)
```

**Use when:** Building interactive UIs with hover effects, drag-and-drop, etc.

#### `WithMouseCellMotion()`

Enables mouse support with cell motion events (only when mouse moves between cells).

```go
bubbly.Run(app, bubbly.WithMouseCellMotion())
```

**Use when:** You need mouse support but want less frequent events

### I/O Options

#### `WithInput(r io.Reader)`

Sets a custom input source. Default is `os.Stdin`.

```go
customInput := strings.NewReader("test input")
bubbly.Run(app, bubbly.WithInput(customInput))
```

**Use when:** Testing, reading from files, or custom input sources

#### `WithOutput(w io.Writer)`

Sets a custom output destination. Default is `os.Stdout`.

```go
var buf bytes.Buffer
bubbly.Run(app, bubbly.WithOutput(&buf))
```

**Use when:** Testing, logging, or custom output destinations

#### `WithInputTTY()`

Forces the program to use a TTY for input.

```go
bubbly.Run(app, bubbly.WithInputTTY())
```

**Use when:** Running in non-interactive environments

### Context and Lifecycle Options

#### `WithContext(ctx context.Context)`

Sets a context for the program. The program will exit when the context is canceled.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithContext(ctx),
)
```

**Use when:** You need timeout control or graceful shutdown

#### `WithoutSignalHandler()`

Disables the default signal handler (SIGINT, SIGTERM).

```go
bubbly.Run(app, bubbly.WithoutSignalHandler())
```

**Use when:** You want to handle signals manually

### Terminal Options

#### `WithoutBracketedPaste()`

Disables bracketed paste mode.

```go
bubbly.Run(app, bubbly.WithoutBracketedPaste())
```

**Use when:** Terminal doesn't support bracketed paste

#### `WithReportFocus()`

Enables focus reporting (terminal gains/loses focus events).

```go
bubbly.Run(app, bubbly.WithReportFocus())
```

**Use when:** You need to react to terminal focus changes

#### `WithEnvironment(env []string)`

Sets custom environment variables for the program.

```go
bubbly.Run(app,
    bubbly.WithEnvironment([]string{"TERM=xterm-256color"}),
)
```

**Use when:** Controlling terminal behavior

### Debug Options

#### `WithoutCatchPanics()`

Disables panic catching. Use during development to see full stack traces.

```go
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithoutCatchPanics(),  // Debug mode
)
```

**Use when:** Debugging panics in development

### Async Options

#### `WithAsyncRefresh(interval time.Duration)`

Explicitly enables or disables async refresh with a specific interval.

```go
// High-frequency updates (20 updates/sec)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(50*time.Millisecond),
)

// Disable async (override auto-detection)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(0),
)
```

**Use when:** You need fine-grained control over async refresh rate

#### `WithoutAsyncAutoDetect()`

Disables automatic async detection. Requires explicit `WithAsyncRefresh()`.

```go
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithoutAsyncAutoDetect(),
    bubbly.WithAsyncRefresh(100*time.Millisecond),
)
```

**Use when:** You want explicit control over async behavior

## Async Auto-Detection

One of the most powerful features of `bubbly.Run()` is automatic async detection, which eliminates the need for manual tick wrapper models.

### How It Works

The framework automatically detects if your component needs async refresh by checking the `WithAutoCommands(true)` flag:

```go
// Component with async operations
app := bubbly.NewComponent("Dashboard").
    WithAutoCommands(true).  // ← Framework detects this!
    Setup(func(ctx *bubbly.Context) {
        // Async data fetching, timers, etc.
        data := composables.UseAsync(ctx, fetchData)
        ctx.OnMounted(func() {
            data.Execute()  // Runs in goroutine
        })
    }).
    Build()

// No manual tick wrapper needed!
bubbly.Run(app, bubbly.WithAltScreen())
```

**What happens internally:**
1. `bubbly.Run()` checks if component has `WithAutoCommands(true)`
2. If yes, wraps component with `asyncWrapperModel`
3. Starts a 100ms ticker to trigger UI updates
4. Your goroutines update Refs, ticker ensures UI redraws

### Before Auto-Detection (Manual Tick Wrapper)

```go
// ❌ 80+ lines of boilerplate
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
        case "ctrl+c":
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

### After Auto-Detection (bubbly.Run)

```go
// ✅ 1 line - 69% code reduction!
func main() {
    app, _ := CreateAsyncApp()
    bubbly.Run(app, bubbly.WithAltScreen())
}
```

### Custom Async Intervals

Override the default 100ms interval:

```go
// Faster updates (20 updates/sec)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(50*time.Millisecond),
)

// Slower updates (5 updates/sec)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(200*time.Millisecond),
)
```

### Disabling Async

Force disable async even if component has `WithAutoCommands(true)`:

```go
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(0),  // Explicitly disable
)
```

## Advanced Configurations

### Multiple Options

Combine multiple options for complex configurations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithMouseAllMotion(),
    bubbly.WithFPS(120),
    bubbly.WithContext(ctx),
    bubbly.WithAsyncRefresh(50*time.Millisecond),
)
```

### Production Configuration

Recommended setup for production applications:

```go
func main() {
    app, err := CreateApp()
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
    }
    
    // Production-ready configuration
    if err := bubbly.Run(app,
        bubbly.WithAltScreen(),           // Full-screen mode
        bubbly.WithMouseAllMotion(),      // Mouse support
        bubbly.WithReportFocus(),         // Focus events
        // WithoutCatchPanics() removed for production
    ); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}
```

### Development Configuration

Recommended setup for development:

```go
func main() {
    app, err := CreateApp()
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
    }
    
    // Development configuration
    if err := bubbly.Run(app,
        bubbly.WithAltScreen(),
        bubbly.WithMouseAllMotion(),
        bubbly.WithoutCatchPanics(),  // See full stack traces
    ); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}
```

### Testing Configuration

Configuration for automated tests:

```go
func TestApp(t *testing.T) {
    app, _ := CreateApp()
    
    var buf bytes.Buffer
    input := strings.NewReader("q")  // Quit immediately
    
    err := bubbly.Run(app,
        bubbly.WithInput(input),
        bubbly.WithOutput(&buf),
        // No WithAltScreen() for tests
    )
    
    assert.NoError(t, err)
    assert.Contains(t, buf.String(), "expected output")
}
```

## FAQ

### When to use Run() vs Wrap()?

**Use `bubbly.Run()`:**
- ✅ New applications (recommended)
- ✅ When you want clean, framework-focused code
- ✅ When you want async auto-detection
- ✅ When you don't need direct Bubbletea access

**Use `bubbly.Wrap()` + `tea.NewProgram()`:**
- ⚠️ Legacy applications (gradual migration)
- ⚠️ When you need direct Bubbletea program access
- ⚠️ When you're integrating with existing Bubbletea code

**Recommendation:** Always use `bubbly.Run()` for new code. It's simpler, cleaner, and more maintainable.

### Can I still use Bubbletea directly?

Yes! `bubbly.Run()` is built on top of Bubbletea, and `bubbly.Wrap()` still works:

```go
// Old pattern still works
p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
p.Run()
```

However, we **strongly recommend** using `bubbly.Run()` for new code. It provides:
- Better abstraction
- Cleaner code
- Async auto-detection
- Framework-level error handling

### How does async detection work?

The framework checks if your component was created with `WithAutoCommands(true)`:

```go
// Component with async flag
app := bubbly.NewComponent("App").
    WithAutoCommands(true).  // ← This flag
    Setup(func(ctx *bubbly.Context) {
        // Async operations here
    }).
    Build()

// Framework detects the flag and enables async automatically
bubbly.Run(app, bubbly.WithAltScreen())
```

**Detection logic:**
1. Check if `WithAutoCommands(true)` is set
2. If yes, wrap with `asyncWrapperModel` (internal)
3. Start 100ms ticker for UI updates
4. Forward all messages to component

### Can I override auto-detection?

Yes! Use `WithAsyncRefresh()`:

```go
// Force enable with custom interval
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(50*time.Millisecond),
)

// Force disable (even if WithAutoCommands is true)
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithAsyncRefresh(0),
)

// Disable auto-detection entirely
bubbly.Run(app,
    bubbly.WithAltScreen(),
    bubbly.WithoutAsyncAutoDetect(),
)
```

### What about error handling?

`bubbly.Run()` returns an error directly, making error handling clean:

```go
if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
    log.Fatalf("Application error: %v", err)
}
```

Compare to manual setup:

```go
p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
if _, err := p.Run(); err != nil {  // Ignoring *tea.Model return value
    log.Fatalf("Application error: %v", err)
}
```

### Is there any performance overhead?

No! `bubbly.Run()` is just a thin wrapper around `tea.NewProgram()`. The only overhead is:
- One function call (negligible)
- Async detection check (happens once at startup)

For async apps, the ticker runs at the same frequency (100ms) whether you use manual tick wrapper or `bubbly.Run()`.

### Can I access the tea.Program?

Not directly with `bubbly.Run()`, as it's designed to abstract away Bubbletea. If you need direct program access, use the manual pattern:

```go
p := tea.NewProgram(bubbly.Wrap(app), tea.WithAltScreen())
// Now you have access to p
p.Send(customMsg)
p.Run()
```

However, this is rarely needed. Most use cases are covered by:
- Component events (`ctx.On()`, `component.Emit()`)
- Lifecycle hooks (`ctx.OnMounted()`, etc.)
- Composables (`UseAsync`, `UseInterval`, etc.)

## Comparison with Manual Setup

| Aspect | Manual Setup | bubbly.Run() |
|--------|-------------|--------------|
| **Lines of code (sync)** | 3-5 lines | 1 line |
| **Lines of code (async)** | 80-100 lines | 1 line |
| **Bubbletea imports** | Required | Not needed |
| **Async tick wrapper** | Manual (80+ lines) | Automatic |
| **Error handling** | Manual | Built-in |
| **Type safety** | tea.ProgramOption | bubbly.RunOption |
| **Framework focus** | Low (exposes Bubbletea) | High (pure BubblyUI) |
| **Maintainability** | Lower | Higher |
| **Recommended for** | Legacy code | New code |

## Examples

### Simple Counter

```go
package main

import (
    "fmt"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    counter, _ := bubbly.NewComponent("Counter").
        WithKeyBinding("up", "increment", "Increment").
        WithKeyBinding("down", "decrement", "Decrement").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubbly.Context) {
            count := ctx.Ref(0)
            ctx.Expose("count", count)
            
            ctx.On("increment", func(interface{}) {
                count.Set(count.GetTyped().(int) + 1)
            })
            ctx.On("decrement", func(interface{}) {
                count.Set(count.GetTyped().(int) - 1)
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[interface{}])
            comp := ctx.Component()
            return fmt.Sprintf("Count: %d\n\n%s",
                count.GetTyped().(int),
                comp.HelpText())
        }).
        Build()
    
    bubbly.Run(counter, bubbly.WithAltScreen())
}
```

### Async Data Dashboard

```go
package main

import (
    "time"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

func fetchData() (string, error) {
    time.Sleep(2 * time.Second)
    return "Data loaded!", nil
}

func main() {
    dashboard, _ := bubbly.NewComponent("Dashboard").
        WithAutoCommands(true).  // Enable async auto-detection
        WithKeyBinding("r", "refresh", "Refresh data").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubbly.Context) {
            data := composables.UseAsync(ctx, fetchData)
            
            ctx.OnMounted(func() {
                data.Execute()
            })
            
            ctx.On("refresh", func(interface{}) {
                data.Execute()
            })
            
            ctx.Expose("data", data)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            data := ctx.Get("data").(*composables.AsyncState[string])
            comp := ctx.Component()
            
            status := "Ready"
            if data.Loading.GetTyped() {
                status = "Loading..."
            } else if data.Error.GetTyped() != nil {
                status = fmt.Sprintf("Error: %v", data.Error.GetTyped())
            } else if data.Data.GetTyped() != nil {
                status = *data.Data.GetTyped()
            }
            
            return fmt.Sprintf("Status: %s\n\n%s", status, comp.HelpText())
        }).
        Build()
    
    // Async auto-detected - no tick wrapper needed!
    bubbly.Run(dashboard, bubbly.WithAltScreen())
}
```

### Production Application

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    // Create application
    app, err := CreateApp()
    if err != nil {
        log.Fatalf("Failed to create app: %v", err)
    }
    
    // Set up timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()
    
    // Run with production configuration
    if err := bubbly.Run(app,
        bubbly.WithAltScreen(),
        bubbly.WithMouseAllMotion(),
        bubbly.WithReportFocus(),
        bubbly.WithContext(ctx),
        bubbly.WithFPS(60),
    ); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}
```

## See Also

- **[Migration Guide: Manual to Run API](../migration/manual-to-run-api.md)** - Step-by-step migration from manual setup
- **[Component Model](../../specs/02-component-model/)** - Understanding BubblyUI components
- **[Composition API](../../specs/04-composition-api/)** - Advanced composition patterns
- **[Async Composables](../../specs/04-composition-api/README.md#async-composables)** - Using async operations
