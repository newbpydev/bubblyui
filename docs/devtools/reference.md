# Reference

**Complete keyboard shortcuts, commands, and configuration options**

## Keyboard Shortcuts

### Global

| Key         | Action                       |
|-------------|------------------------------|
| `F12`       | Toggle dev tools visibility  |
| `Ctrl+C`    | Quit application             |

### Navigation

| Key           | Action                    |
|---------------|---------------------------|
| `Tab`         | Next panel                |
| `Shift+Tab`   | Previous panel            |
| `↑`           | Navigate up/previous item |
| `↓`           | Navigate down/next item   |
| `←`           | Collapse node / Go back   |
| `→`           | Expand node / Go forward  |
| `Home`        | Jump to first item        |
| `End`         | Jump to last item         |
| `Page Up`     | Scroll up one page        |
| `Page Down`   | Scroll down one page      |

### Component Inspector

| Key     | Action                     |
|---------|----------------------------|
| `Enter` | View component details     |
| `Space` | Toggle node expansion      |
| `/`     | Search components          |
| `Esc`   | Clear search               |
| `1`     | Switch to State tab        |
| `2`     | Switch to Props tab        |
| `3`     | Switch to Events tab       |

### State Viewer

| Key     | Action                     |
|---------|----------------------------|
| `Enter` | Select ref                 |
| `e`     | Edit selected ref value    |
| `h`     | View ref history           |
| `r`     | Restore historical value   |
| `/`     | Filter refs                |
| `Esc`   | Clear selection/filter     |

### Event Tracker

| Key     | Action                     |
|---------|----------------------------|
| `p`     | Pause event capture        |
| `r`     | Resume event capture       |
| `c`     | Clear event log            |
| `/`     | Filter events              |
| `Enter` | View event details         |
| `Esc`   | Close details              |

### Performance Monitor

| Key     | Action                     |
|---------|----------------------------|
| `s`     | Sort by column             |
| `f`     | View flame graph           |
| `t`     | View timeline              |
| `c`     | Clear performance data     |

### Export & Commands

| Key       | Action                     |
|-----------|----------------------------|
| `Ctrl+E`  | Export session             |
| `Ctrl+I`  | Import session             |
| `Ctrl+S`  | Save screenshot (ASCII)    |
| `?`       | Show help                  |

---

## Configuration Options

### Code Configuration

```go
config := devtools.Config{
    // Core
    Enabled: true,
    
    // Layout
    LayoutMode: devtools.LayoutHorizontal,
    SplitRatio: 0.60,
    
    // Limits
    MaxComponents:   10000,
    MaxEvents:       5000,
    MaxStateHistory: 1000,
    
    // Performance
    SamplingRate: 1.0,
    
    // Features
    EnableHooks:              true,
    EnablePerformanceMonitor: true,
    EnableEventTracker:       true,
    EnableStateHistory:       true,
}

devtools.EnableWithConfig(config)
```

### Configuration Options

#### Core Options

| Option    | Type   | Default | Description                     |
|-----------|--------|---------|----------------------------------|
| `Enabled` | bool   | false   | Enable/disable dev tools         |

#### Layout Options

| Option       | Type        | Default            | Description                          |
|--------------|-------------|--------------------|------------------------------------- |
| `LayoutMode` | LayoutMode  | LayoutHorizontal   | Split mode (horizontal/vertical/overlay) |
| `SplitRatio` | float64     | 0.60               | App/tools split (0.0-1.0)            |

**LayoutMode values:**
- `LayoutHorizontal` - Side-by-side (app | tools)
- `LayoutVertical` - Stacked (app / tools)
- `LayoutOverlay` - Full-screen toggle
- `LayoutHidden` - Hidden (data still collected)

#### Limits

| Option             | Type | Default | Description                          |
|--------------------|------|---------|--------------------------------------|
| `MaxComponents`    | int  | 10000   | Max components to track              |
| `MaxEvents`        | int  | 5000    | Max events in log                    |
| `MaxStateHistory`  | int  | 1000    | Max state changes per ref            |

#### Performance Options

| Option         | Type    | Default | Description                               |
|----------------|---------|---------|-------------------------------------------|
| `SamplingRate` | float64 | 1.0     | Fraction of events to capture (0.0-1.0)   |

**Examples:**
- `1.0` - Capture all events (100%)
- `0.5` - Capture half of events (50%)
- `0.1` - Capture 10% of events

#### Feature Toggles

| Option                       | Type | Default | Description                      |
|------------------------------|------|---------|----------------------------------|
| `EnableHooks`                | bool | true    | Enable framework hooks           |
| `EnablePerformanceMonitor`   | bool | true    | Enable performance tracking      |
| `EnableEventTracker`         | bool | true    | Enable event capture             |
| `EnableStateHistory`         | bool | true    | Enable state history tracking    |

---

## Environment Variables

All configuration can be set via environment variables.

### Core

| Variable                     | Type   | Example              |
|------------------------------|--------|----------------------|
| `BUBBLY_DEVTOOLS_ENABLED`    | bool   | `true`               |

### Layout

| Variable                        | Type    | Example        |
|---------------------------------|---------|----------------|
| `BUBBLY_DEVTOOLS_LAYOUT_MODE`   | string  | `horizontal`   |
| `BUBBLY_DEVTOOLS_SPLIT_RATIO`   | float   | `0.60`         |

**LAYOUT_MODE values:**
- `horizontal`
- `vertical`
- `overlay`
- `hidden`

### Limits

| Variable                              | Type | Example  |
|---------------------------------------|------|----------|
| `BUBBLY_DEVTOOLS_MAX_COMPONENTS`      | int  | `10000`  |
| `BUBBLY_DEVTOOLS_MAX_EVENTS`          | int  | `5000`   |
| `BUBBLY_DEVTOOLS_MAX_STATE_HISTORY`   | int  | `1000`   |

### Performance

| Variable                            | Type    | Example |
|-------------------------------------|---------|---------|
| `BUBBLY_DEVTOOLS_SAMPLING_RATE`     | float   | `1.0`   |

### Feature Toggles

| Variable                                     | Type | Example |
|----------------------------------------------|------|---------|
| `BUBBLY_DEVTOOLS_ENABLE_HOOKS`               | bool | `true`  |
| `BUBBLY_DEVTOOLS_ENABLE_PERFORMANCE_MONITOR` | bool | `true`  |
| `BUBBLY_DEVTOOLS_ENABLE_EVENT_TRACKER`       | bool | `true`  |
| `BUBBLY_DEVTOOLS_ENABLE_STATE_HISTORY`       | bool | `true`  |

### Example .env File

```bash
# Enable dev tools
BUBBLY_DEVTOOLS_ENABLED=true

# Layout
BUBBLY_DEVTOOLS_LAYOUT_MODE=horizontal
BUBBLY_DEVTOOLS_SPLIT_RATIO=0.60

# Limits
BUBBLY_DEVTOOLS_MAX_COMPONENTS=10000
BUBBLY_DEVTOOLS_MAX_EVENTS=5000
BUBBLY_DEVTOOLS_MAX_STATE_HISTORY=1000

# Performance
BUBBLY_DEVTOOLS_SAMPLING_RATE=1.0

# Features
BUBBLY_DEVTOOLS_ENABLE_HOOKS=true
BUBBLY_DEVTOOLS_ENABLE_PERFORMANCE_MONITOR=true
BUBBLY_DEVTOOLS_ENABLE_EVENT_TRACKER=true
BUBBLY_DEVTOOLS_ENABLE_STATE_HISTORY=true
```

---

## Command-Line Flags

For applications that accept flags:

```go
package main

import (
    "flag"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

func main() {
    // Define flags
    enableDevTools := flag.Bool("devtools", false, "Enable dev tools")
    layoutMode := flag.String("devtools-layout", "horizontal", "Dev tools layout mode")
    splitRatio := flag.Float64("devtools-split", 0.60, "Dev tools split ratio")
    
    flag.Parse()
    
    // Configure dev tools
    if *enableDevTools {
        config := devtools.DefaultConfig()
        config.LayoutMode = parseLayoutMode(*layoutMode)
        config.SplitRatio = *splitRatio
        
        devtools.EnableWithConfig(config)
    }
    
    // Your app...
}
```

**Usage:**
```bash
go run main.go --devtools --devtools-layout=vertical --devtools-split=0.70
```

---

## API Quick Reference

### Core API

```go
// Lifecycle
devtools.Enable() *DevTools
devtools.Disable()
devtools.Toggle()
devtools.IsEnabled() bool
devtools.Get() *DevTools

// Instance methods
dt.SetVisible(bool)
dt.ToggleVisibility()
dt.IsVisible() bool
```

### Export/Import

```go
// Export
dt.Export(filename string, opts ExportOptions) error
dt.ExportFull(filename string, opts ExportOptions) (*ExportCheckpoint, error)
dt.ExportIncremental(filename string, checkpoint *ExportCheckpoint) (*ExportCheckpoint, error)
dt.ExportStream(filename string, opts ExportOptions) error

// Import
devtools.Import(filename string) error
devtools.ImportDelta(filename string) error
devtools.ImportFromReader(reader io.Reader) error
devtools.ValidateImport(data *ExportData) error
```

### Sanitization

```go
// Create
sanitizer := devtools.NewSanitizer()

// Configure
sanitizer.LoadTemplate(name string) error
sanitizer.LoadTemplates(names ...string) error
sanitizer.AddPattern(pattern, replacement string)
sanitizer.AddPatternWithPriority(pattern, replacement string, priority int, name string) error

// Use
cleanData := sanitizer.Sanitize(data *ExportData) *ExportData
result := sanitizer.Preview(data *ExportData) *DryRunResult
stats := sanitizer.GetLastStats() *SanitizationStats
```

### Framework Hooks

```go
// Registration
bubbly.RegisterHook(hook FrameworkHook) error
bubbly.UnregisterHook() error
bubbly.IsHookRegistered() bool
```

### Configuration

```go
// Create
config := devtools.DefaultConfig()
config, err := devtools.LoadConfig(path string)

// Apply
devtools.EnableWithConfig(config *Config)
config.ApplyEnvOverrides()
err := config.Validate()
```

---

## Data Types

### Core Types

```go
type DevTools struct { /* ... */ }
type Config struct { /* ... */ }
type ExportOptions struct { /* ... */ }
```

### Component Types

```go
type ComponentSnapshot struct { /* ... */ }
type ComponentInspector struct { /* ... */ }
type TreeView struct { /* ... */ }
type DetailPanel struct { /* ... */ }
```

### State Types

```go
type RefSnapshot struct { /* ... */ }
type StateHistory struct { /* ... */ }
type StateChange struct { /* ... */ }
type StateViewer struct { /* ... */ }
```

### Event Types

```go
type EventRecord struct { /* ... */ }
type EventTracker struct { /* ... */ }
type EventStatistics struct { /* ... */ }
```

### Performance Types

```go
type PerformanceData struct { /* ... */ }
type ComponentPerformance struct { /* ... */ }
type FlameGraph struct { /* ... */ }
```

See [API Reference](./api-reference.md) for complete type definitions.

---

## Performance Characteristics

| Operation              | Target    | Typical   |
|------------------------|-----------|-----------|
| Enable/Disable         | < 100ms   | ~50ms     |
| Toggle Visibility      | < 10ms    | ~5ms      |
| Component Inspection   | < 50ms    | ~30ms     |
| State Update           | < 10ms    | ~5ms      |
| Event Capture          | < 1ms     | ~0.5ms    |
| Search                 | < 100ms   | ~50ms     |
| Export (2MB)           | < 1s      | ~500ms    |
| Import (2MB)           | < 1s      | ~500ms    |
| Overhead (enabled)     | < 5%      | ~3%       |
| Overhead (disabled)    | 0%        | 0%        |

---

## Links

- **[API Reference](./api-reference.md)** - Complete API documentation
- **[Features Guide](./features.md)** - Detailed feature tour
- **[Framework Hooks](./hooks.md)** - Hook implementation guide
- **[Export & Import](./export-import.md)** - Data management guide
- **[Best Practices](./best-practices.md)** - Optimization tips
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions

---

**Questions?** Check [Troubleshooting](./troubleshooting.md) or open an issue on GitHub.
