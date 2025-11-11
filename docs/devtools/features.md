# Features Overview

**Complete tour of BubblyUI Dev Tools capabilities**

This guide provides an in-depth look at every feature, with examples and use cases.

## Table of Contents

1. [Component Inspector](#component-inspector)
2. [State Viewer](#state-viewer)
3. [Event Tracker](#event-tracker)
4. [Performance Monitor](#performance-monitor)
5. [Framework Hooks](#framework-hooks)
6. [Export & Import System](#export--import-system)
7. [Data Sanitization](#data-sanitization)
8. [Responsive UI](#responsive-ui)
9. [Configuration](#configuration)

---

## Component Inspector

Visualize and inspect your component hierarchy in real-time.

### Component Tree View

Hierarchical representation of your application's component structure:

```
Component Tree
├─ App
│  ├─ Header
│  ├─ TodoList (*)
│  │  ├─ TodoItem #1
│  │  ├─ TodoItem #2
│  │  └─ TodoItem #3
│  └─ Footer
```

**Navigation:**
- `↑`/`↓` - Select component
- `→` - Expand node
- `←` - Collapse node
- `Space` - Toggle expansion
- `Enter` - View details

### Component Details Panel

When a component is selected, view comprehensive details:

#### State Tab
```
State: TodoList
───────────────────────────
Refs:
• items: [3 items] (Array)
• filter: "all" (String)

Computed:
• visibleItems: [3 items] (computed from: items, filter)
• completedCount: 2 (computed from: items)

Watchers:
• watch-0x1a2b3c (watching: items)
• watch-0x4d5e6f (watching: filter)
```

#### Props Tab
```
Props: TodoItem
───────────────────────────
• id: "todo-123"
• text: "Buy groceries"
• done: false
• onToggle: function
```

#### Events Tab
```
Recent Events: TodoItem
───────────────────────────
[11:30:45] click → TodoItem#todo-123
[11:30:45] toggle → TodoList
```

### Search and Filter

Find components quickly:

```
Press '/' to search
> todo

Results (3):
├─ TodoList
├─ TodoItem #1
└─ TodoItem #2
```

**Search features:**
- Fuzzy matching
- Search by component name or ID
- Real-time results
- Navigate results with `↑`/`↓`

---

## State Viewer

Track all reactive state in your application with complete history.

### All Refs Display

```
State Viewer
───────────────────────────────────────
Ref Name           Value      Watchers
─────────────────────────────────────── 
count              5          1
username           "john"     2
isValid            true       1
items              [...]      3
```

### State History

View complete timeline of state changes:

```
State History: count
───────────────────────────────────────
Time         Old → New    Source
───────────────────────────────────────
11:30:45     0 → 1        Button#increment
11:30:47     1 → 2        Button#increment
11:30:50     2 → 0        Button#reset
```

**Features:**
- Circular buffer (configurable size)
- Timestamp precision (milliseconds)
- Source component tracking
- Filter by ref name

### Time-Travel Debugging

Restore previous state for testing:

```
1. Select ref in State Viewer
2. Press 'h' for history
3. Select historical value
4. Press 'r' to restore
```

### Edit State Values

Test edge cases by editing state directly:

```
1. Select ref
2. Press 'e' to edit
3. Enter new value
4. Press Enter to apply
```

**Supported types:**
- Strings: `"new value"`
- Numbers: `42`, `3.14`
- Booleans: `true`, `false`
- Arrays: `[1, 2, 3]`
- Objects: `{"key": "value"}`

---

## Event Tracker

Capture, filter, and replay component events.

### Event Log

Real-time event capture with full details:

```
Event Log (showing 50 of 125)
─────────────────────────────────────────────────────────
Time      Event Name    Source           Target      Payload
───────────────────────────────────────────────────────── 
11:30:45  click         Button#btn-1     -           {}
11:30:45  increment     Counter          -           {amount: 1}
11:30:47  input         TextInput        -           {value: "hello"}
11:30:48  submit        Form             -           {data: {...}}
```

### Event Filtering

Focus on specific events:

```
Filter by:
• Event name: "click", "input", "submit"
• Source component: "Button*", "Form"
• Time range: last 10s, last 1m
• Payload content: search within event data
```

**Filter syntax:**
- Exact: `click`
- Wildcard: `*click*`, `Button*`
- Multiple: `click,input,submit`

### Event Statistics

Analyze event patterns:

```
Event Statistics
───────────────────────────────
Event Name     Count    Freq/sec
───────────────────────────────
click          45       0.75
input          12       0.20
submit         3        0.05
```

### Pause/Resume Capture

Control when events are captured:

```
Press 'p' - Pause capture
Press 'r' - Resume capture

Status: ⏸ PAUSED (125 events in buffer)
```

---

## Performance Monitor

Profile component render performance and detect bottlenecks.

### Render Timing

Per-component performance metrics:

```
Performance Monitor
─────────────────────────────────────────────────────
Component    Renders  Avg    Min    Max    Total
─────────────────────────────────────────────────────
TodoList     15       3.2ms  1.1ms  8.5ms  48.0ms  ⚠
TodoItem     45       0.8ms  0.5ms  2.1ms  36.0ms  ✓
Header       5        0.4ms  0.3ms  0.6ms  2.0ms   ✓
```

**Performance zones:**
- ✓ Green (< 10ms): Optimal
- ⚠ Yellow (10-50ms): Acceptable
- ❌ Red (> 50ms): Needs optimization

### Flame Graph

Visualize render call stack:

```
Flame Graph (11:30:45.123 - 11:30:45.135)
═══════════════════════════════════════════════════
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ App.View() 12ms
  ▓▓▓▓▓▓▓▓▓▓ TodoList.View() 8ms
    ▓▓ TodoItem.View() 2ms
    ▓▓ TodoItem.View() 2ms
    ▓▓ TodoItem.View() 2ms
  ▓▓ Header.View() 2ms
```

**Usage:**
1. Select time range (drag on timeline)
2. View flame graph for that period
3. Click on bars to drill down
4. Identify slowest operations

### Timeline View

Chronological render sequence:

```
Timeline (last 5 seconds)
═══════════════════════════════════════════════════
11:30:45  App         ▓▓▓▓▓▓▓▓▓▓▓▓ 12ms
11:30:46  TodoList    ▓▓▓▓▓▓▓ 8ms
11:30:47  TodoItem    ▓▓ 2ms
11:30:48  TodoItem    ▓▓ 2ms
```

### Slow Operation Detection

Automatic alerts for slow renders:

```
⚠ SLOW RENDER DETECTED
Component: TodoList
Duration: 52.3ms (threshold: 50ms)
Timestamp: 11:30:45.123
Suggestion: Move expensive operations to Update()
```

---

## Framework Hooks

Visualize reactive cascades and component lifecycle events.

See the dedicated [Framework Hooks Guide](./hooks.md) for comprehensive coverage.

### Reactive Cascade Visualization

Track data flow through your application:

```
Reactive Cascade
═══════════════════════════════════════════════════
1. Ref Change: count (0 → 1)
   ↓
2. Computed Re-eval: isEven (true → false)
   ↓
3. Watch Callback: watch-0x1a2b (triggered)
   ↓
4. Effect Run: effect-0x3c4d (re-run)
   ↓
5. Component Update: Counter (re-render)
```

### Hook Events

Track all framework lifecycle events:

```
Hook Events
═══════════════════════════════════════════════════
[11:30:45.001] OnComponentMount(Counter, counter-1)
[11:30:45.002] OnRefChange(count, 0, 1)
[11:30:45.003] OnComputedChange(isEven, true, false)
[11:30:45.004] OnWatchCallback(watch-0x1a2b, 1, 0)
[11:30:45.005] OnEffectRun(effect-0x3c4d)
[11:30:45.006] OnRenderComplete(Counter, 2.1ms)
```

### Component Tree Mutations

Track parent-child relationships:

```
Tree Mutations
═══════════════════════════════════════════════════
[11:30:45] OnChildAdded(TodoList, TodoItem#3)
[11:30:50] OnChildRemoved(TodoList, TodoItem#1)
```

---

## Export & Import System

Save and share debug sessions with multiple formats and compression.

See the dedicated [Export & Import Guide](./export-import.md) for complete details.

### Quick Export

```go
devtools.Export("session.json", devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
    IncludePerf:   true,
})
```

### Compression

Reduce file size by 60-70%:

```go
devtools.Export("session.json.gz", devtools.ExportOptions{
    Compress:         true,
    CompressionLevel: gzip.BestCompression,  // 70% reduction
})
```

**Compression levels:**
- `gzip.BestSpeed` - 50% reduction, fast
- `gzip.DefaultCompression` - 60% reduction, balanced
- `gzip.BestCompression` - 70% reduction, maximum

### Multiple Formats

Choose the best format for your use case:

```go
// JSON - Universal, human-readable
devtools.Export("session.json", opts)

// YAML - Most readable, configuration tools
devtools.Export("session.yaml", opts)

// MessagePack - Smallest, fastest (binary)
devtools.Export("session.msgpack", opts)
```

**Format comparison:**
- JSON: 100% size, universal compatibility
- YAML: 110% size, human-friendly
- MessagePack: 60% size, binary format

### Auto-Import with Format Detection

```go
// Automatically detects format and compression
devtools.Import("session.json.gz")  // JSON + gzip
devtools.Import("session.yaml")     // YAML
devtools.Import("session.msgpack")  // MessagePack
```

---

## Data Sanitization

Remove sensitive data before sharing exports.

### Built-in Compliance Templates

```go
sanitizer := devtools.NewSanitizer()

// Load compliance templates
sanitizer.LoadTemplates("pii", "pci", "hipaa", "gdpr")
```

**Templates:**
- `pii` - Personal Identifiable Information (SSN, email, phone)
- `pci` - Payment Card Industry (credit cards, CVV)
- `hipaa` - Health Insurance Portability (medical records)
- `gdpr` - General Data Protection Regulation (IP, MAC addresses)

### Custom Patterns

Add custom sanitization rules:

```go
// Redact API keys with priority
sanitizer.AddPatternWithPriority(
    `(?i)(api[_-]?key)(["'\s:=]+)([^\s"']+)`,
    "${1}${2}[REDACTED]",
    80,  // High priority (0-100)
    "api_key",
)
```

### Preview (Dry-Run)

See what would be redacted before applying:

```go
result := sanitizer.Preview(exportData)
fmt.Printf("Would redact %d values:\n", result.WouldRedactCount)
for _, match := range result.Matches {
    fmt.Printf("  - %s: %s\n", match.Pattern, match.Location)
}
```

### Sanitization Metrics

Track sanitization effectiveness:

```go
stats := sanitizer.GetLastStats()
fmt.Printf("Patterns applied: %d\n", stats.PatternsApplied)
fmt.Printf("Values redacted: %d\n", stats.ValuesRedacted)
fmt.Printf("Processing time: %v\n", stats.Duration)
```

---

## Responsive UI

Adaptive layout that responds to terminal size changes.

### Automatic Layout Modes

```go
config := devtools.DefaultConfig()
config.LayoutMode = devtools.LayoutHorizontal  // Side-by-side (default)
```

**Available modes:**
- `LayoutHorizontal` - Side-by-side (app | tools)
- `LayoutVertical` - Stacked (app / tools)
- `LayoutOverlay` - Full-screen toggle (F12 switches views)
- `LayoutHidden` - Dev tools disabled

### Terminal Resize Handling

Dev tools automatically adapt to terminal size:

```
Small terminal (80x24):
├─ Overlay mode (one at a time)

Medium terminal (120x40):
├─ Vertical split

Large terminal (200x60):
├─ Horizontal split (60/40)
```

### Manual Layout Control

Override automatic layout:

```go
dt := devtools.Get()
dt.SetManualLayoutMode(devtools.LayoutVertical)
dt.EnableAutoLayout(false)  // Disable automatic adaptation
```

### Configurable Split Ratio

Adjust space allocation:

```go
config.SplitRatio = 0.70  // 70% app, 30% dev tools
```

---

## Configuration

Customize dev tools behavior via code or environment variables.

### Code-Based Configuration

```go
config := devtools.DefaultConfig()

// Layout
config.LayoutMode = devtools.LayoutHorizontal
config.SplitRatio = 0.60

// Limits
config.MaxComponents = 10000
config.MaxEvents = 5000
config.MaxStateHistory = 1000

// Performance
config.SamplingRate = 1.0  // 100% of events (0.5 = 50%)

// Features
config.EnableHooks = true
config.EnablePerformanceMonitor = true

devtools.EnableWithConfig(config)
```

### Environment Variables

```bash
# Enable/disable
export BUBBLY_DEVTOOLS_ENABLED=true

# Layout
export BUBBLY_DEVTOOLS_LAYOUT_MODE=horizontal
export BUBBLY_DEVTOOLS_SPLIT_RATIO=0.60

# Limits
export BUBBLY_DEVTOOLS_MAX_COMPONENTS=10000
export BUBBLY_DEVTOOLS_MAX_EVENTS=5000
export BUBBLY_DEVTOOLS_MAX_STATE_HISTORY=1000

# Performance
export BUBBLY_DEVTOOLS_SAMPLING_RATE=1.0
```

### Load from File

```go
config, err := devtools.LoadConfig("devtools.yaml")
if err != nil {
    log.Fatal(err)
}
devtools.EnableWithConfig(config)
```

**Example `devtools.yaml`:**
```yaml
enabled: true
layoutMode: horizontal
splitRatio: 0.60
maxComponents: 10000
maxEvents: 5000
samplingRate: 1.0
```

---

## Next Steps

- **[Framework Hooks](./hooks.md)** - Deep dive into reactive cascade tracking
- **[Export & Import](./export-import.md)** - Complete export/import workflows
- **[Best Practices](./best-practices.md)** - Performance optimization tips
- **[Reference](./reference.md)** - Complete keyboard shortcuts and API

---

**Found an issue?** See [Troubleshooting](./troubleshooting.md) or report on GitHub.
