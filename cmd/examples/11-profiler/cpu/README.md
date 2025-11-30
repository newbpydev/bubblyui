# CPU Profiler Example

A fully functional CPU profiling dashboard demonstrating BubblyUI's composable architecture with pprof integration.

## Features

- **CPU Profiling**: Start/stop CPU profiling with pprof integration
- **Hot Function Analysis**: View top CPU consumers after profiling
- **Multi-Pane Focus**: Navigate between Profile, Controls, and Results panels
- **State Machine Workflow**: Clear state transitions (Idle → Profiling → Complete → Analyzed)
- **Dynamic Key Bindings**: Context-aware controls based on state and focus
- **pprof Integration**: Generated profiles compatible with `go tool pprof`

## Running the Example

```bash
go run ./cmd/examples/11-profiler/cpu
```

## Architecture

### Directory Structure

```
cmd/examples/11-profiler/cpu/
├── main.go                    # Entry point with bubbly.Run()
├── app.go                     # Root component with multi-pane layout
├── composables/
│   ├── use_cpu_profiler.go    # CPU profiler composable
│   └── use_cpu_profiler_test.go
└── components/
    ├── profile_panel.go       # Profile status display
    ├── profile_panel_test.go
    ├── controls_panel.go      # Start/Stop/Analyze controls
    ├── controls_panel_test.go
    ├── results_panel.go       # Hot functions display
    ├── results_panel_test.go
    ├── status_bar.go          # Status and help text
    └── status_bar_test.go
```

### Composables Used

- **UseCPUProfiler**: Custom composable wrapping `profiler.CPUProfiler`
  - State machine: Idle → Profiling → Complete
  - Reactive refs for all state
  - Start/Stop/Analyze/Reset methods
  - Hot function analysis

- **UseInterval**: Built-in composable for live duration updates

### Components Created

1. **ProfilePanel**: Shows profiling status
   - Idle: "No profile active"
   - Profiling: Filename, live duration
   - Complete: File size, duration

2. **ControlsPanel**: Action controls
   - Start/Stop button (state-aware)
   - Analyze button (when complete)
   - Reset button
   - Focus indicator

3. **ResultsPanel**: Analysis results
   - Hot functions list (top 5)
   - Function name, percentage, samples
   - pprof command hint

4. **StatusBar**: Status and help
   - State badge (color-coded)
   - Duration display
   - Focus indicator
   - Context-aware help text

### BubblyUI Components Used

- `components.Card` - Content containers
- Lipgloss for layout composition only

## Controls

| Key | Action | Condition |
|-----|--------|-----------|
| `Tab` | Switch focus between panels | Always |
| `Space` | Start/Stop profiling | Controls focused |
| `a` | Analyze results | Complete state, Controls focused |
| `r` | Reset profiler | Always |
| `q` | Quit application | Always |

## Workflow

1. **Start**: Press `Space` to begin CPU profiling
2. **Wait**: Let it run for desired duration (watch live timer)
3. **Stop**: Press `Space` to stop profiling
4. **Analyze**: Press `a` to analyze the profile
5. **View**: See hot functions in Results panel
6. **pprof**: Use `go tool pprof <file>` for detailed analysis
7. **Reset**: Press `r` to start over

## State Machine

```
┌─────────┐  Space   ┌────────────┐  Space   ┌──────────┐
│  IDLE   │ ───────► │ PROFILING  │ ───────► │ COMPLETE │
└─────────┘          └────────────┘          └──────────┘
     ▲                                            │
     │                                            │ 'a'
     │                                            ▼
     │                                      ┌──────────┐
     └────────────── 'r' ◄────────────────  │ ANALYZED │
                                            └──────────┘
```

## Focus Management

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Profile   │ ──► │  Controls   │ ──► │   Results   │
│   Panel     │     │   Panel     │     │   Panel     │
└─────────────┘     └─────────────┘     └─────────────┘
       ▲                                       │
       └───────────────────────────────────────┘
                      Tab cycles
```

## Test Coverage

- **composables/**: 100% coverage
- **components/**: 98.7% coverage

## Zero Bubbletea

This example uses **only BubblyUI framework** - no raw `tea.Model` implementation:

```go
// Entry point - zero boilerplate
func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())
}
```

## pprof Integration

After profiling, use Go's pprof tools:

```bash
# Interactive mode
go tool pprof cpu-20241130-123456.prof

# Web UI (requires graphviz)
go tool pprof -http=:8080 cpu-20241130-123456.prof

# Top functions
go tool pprof -top cpu-20241130-123456.prof
```

## Key Patterns Demonstrated

1. **Composable Architecture**: Reusable logic encapsulated in `UseCPUProfiler`
2. **State Machine**: Clear state transitions with reactive updates
3. **Multi-Pane Focus**: Tab navigation with visual indicators
4. **Dynamic UI**: Controls and help text change based on state
5. **Callback Props**: Parent-child communication via callbacks
6. **Typed Refs**: Type-safe reactive state with `bubbly.NewRef[T]()`
