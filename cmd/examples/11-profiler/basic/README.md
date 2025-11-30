# Basic Profiler Example

A real-time performance monitoring dashboard demonstrating the BubblyUI profiler package.

## Features

- **Live Metrics Display**: FPS, memory usage, goroutine count, render count, bottlenecks
- **Multi-Pane Focus Management**: Tab between Metrics and Controls panels
- **Dynamic Key Bindings**: Context-aware controls based on focused panel
- **Real-Time Updates**: UseInterval composable for periodic metric refresh
- **Export Reports**: Generate HTML performance reports

## Architecture

```
basic/
├── main.go                    # Entry point with bubbly.Run() - ZERO BOILERPLATE!
├── app.go                     # Root component with multi-pane layout
├── composables/
│   ├── use_profiler.go        # Wraps profiler.Profiler with reactive state
│   └── use_profiler_test.go   # 96.8% coverage
└── components/
    ├── metrics_panel.go       # Shows live metrics (FPS, memory, etc.)
    ├── metrics_panel_test.go  # 99.5% coverage
    ├── controls_panel.go      # Start/stop/reset controls
    ├── controls_panel_test.go
    ├── status_bar.go          # Shows profiler status and help
    └── status_bar_test.go
```

## BubblyUI Patterns Demonstrated

### Zero Boilerplate Entry Point
```go
func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())  // That's it!
}
```

### Composables
- **UseProfiler**: Custom composable encapsulating profiler logic
- **UseInterval**: Built-in composable for periodic updates
- **Typed Refs**: `bubbly.NewRef[T]()` for type-safe reactive state
- **Computed Values**: Derived state (duration from start time)

### Components
- **Card**: Content containers with dynamic border colors
- **Text**: Styled text labels with color coding
- **Spacer**: Layout spacing

### Key Bindings
- **Global**: Tab, q, ctrl+c
- **Context-Aware**: Space (toggle) only works when Controls focused

### Focus Management
- Multi-pane navigation with Tab
- Visual feedback (green border = focused)
- Status bar shows current focus

## Running

```bash
go run ./cmd/examples/11-profiler/basic
```

## Controls

| Key | Action | Context |
|-----|--------|---------|
| Tab | Switch focus between panels | Global |
| Space | Start/Stop profiler | Controls focused |
| r | Reset metrics | Any |
| e | Export report to HTML | Any |
| q | Quit | Global |
| ctrl+c | Quit | Global |

## User Workflow

1. **Start**: App opens with Controls panel focused, profiler stopped
2. **Start Profiling**: Press Space to start
3. **Monitor**: Watch live metrics update every 100ms
4. **Navigate**: Press Tab to switch to Metrics panel
5. **Stop**: Press Tab back to Controls, then Space to stop
6. **Export**: Press 'e' to save HTML report
7. **Reset**: Press 'r' to clear all metrics
8. **Quit**: Press 'q' to exit

## Test Coverage

```bash
go test -race -cover ./cmd/examples/11-profiler/basic/...
```

- **composables**: 96.8% coverage
- **components**: 99.5% coverage

## Key Implementation Details

### UseProfiler Composable

```go
type ProfilerComposable struct {
    Profiler       *profiler.Profiler
    IsRunning      *bubbly.Ref[bool]
    Metrics        *bubbly.Ref[*ProfilerMetrics]
    StartTime      *bubbly.Ref[time.Time]
    Duration       *bubbly.Computed[interface{}]
    LastExport     *bubbly.Ref[string]
    
    Start          func()
    Stop           func()
    Toggle         func()
    Reset          func()
    ExportReport   func(filename string) error
    RefreshMetrics func()
}
```

### Focus State Management

```go
// Create focus state refs
focusedPane := bubbly.NewRef(components.FocusControls)
metricsFocused := bubbly.NewRef(false)
controlsFocused := bubbly.NewRef(true)

// Switch focus on Tab
ctx.On("switchFocus", func(_ interface{}) {
    if focusedPane.GetTyped() == components.FocusMetrics {
        focusedPane.Set(components.FocusControls)
        metricsFocused.Set(false)
        controlsFocused.Set(true)
    } else {
        focusedPane.Set(components.FocusMetrics)
        metricsFocused.Set(true)
        controlsFocused.Set(false)
    }
})
```

### Live Updates with UseInterval

```go
interval := bubblyComposables.UseInterval(ctx, func() {
    if profiler.IsRunning.GetTyped() {
        profiler.RefreshMetrics()
    }
}, 100*time.Millisecond)

interval.Start()
```

## Related Examples

- `11-profiler/cpu/` - CPU profiling workflow
- `11-profiler/memory/` - Memory leak detection
- `11-profiler/benchmark/` - Benchmark integration
