# BubblyUI Dev Tools API Reference

**Version:** 1.0  
**Package:** `github.com/newbpydev/bubblyui/pkg/bubbly/devtools`

## Table of Contents

- [Core API](#core-api)
- [Component Inspector](#component-inspector)
- [State Viewer](#state-viewer)
- [Event Tracker](#event-tracker)
- [Performance Monitor](#performance-monitor)
- [Export System](#export-system)
- [Import System](#import-system)
- [Sanitization](#sanitization)
- [Framework Hooks](#framework-hooks)
- [Configuration](#configuration)
- [Data Types](#data-types)

---

## Core API

### Enable

```go
func Enable() *DevTools
```

Enables dev tools and returns the singleton instance. Safe to call multiple times (returns existing instance).

**Returns:** `*DevTools` - The global dev tools instance

**Example:**
```go
dt := devtools.Enable()
dt.SetVisible(true)
```

---

### Disable

```go
func Disable()
```

Disables dev tools, stops data collection, and cleans up resources.

---

### Toggle

```go
func Toggle()
```

Toggles dev tools enabled/disabled state.

---

### IsEnabled

```go
func IsEnabled() bool
```

Returns true if dev tools are currently enabled.

**Example:**
```go
if devtools.IsEnabled() {
    fmt.Println("Dev tools active")
}
```

---

### Instance Methods

#### SetVisible

```go
func (dt *DevTools) SetVisible(visible bool)
```

Shows or hides the dev tools UI. Data collection continues even when hidden.

**Parameters:**
- `visible` - true to show, false to hide

---

#### ToggleVisibility

```go
func (dt *DevTools) ToggleVisibility()
```

Toggles UI visibility.

---

#### IsVisible

```go
func (dt *DevTools) IsVisible() bool
```

Returns true if UI is currently visible.

---

## Component Inspector

### ComponentSnapshot

```go
type ComponentSnapshot struct {
    ID        string
    Name      string
    Type      string
    Status    string
    Parent    *ComponentSnapshot
    Children  []*ComponentSnapshot
    State     map[string]interface{}
    Props     map[string]interface{}
    Refs      []*RefSnapshot
    Timestamp time.Time
}
```

Captures component state at a point in time.

---

### CaptureComponent

```go
func CaptureComponent(comp ComponentInterface) *ComponentSnapshot
```

Creates a snapshot of a component's current state.

---

### TreeView

```go
type TreeView struct { ... }

func NewTreeView(root *ComponentSnapshot) *TreeView
func (tv *TreeView) Render() string
func (tv *TreeView) Select(id string)
func (tv *TreeView) Expand(id string)
func (tv *TreeView) Collapse(id string)
func (tv *TreeView) Toggle(id string)
func (tv *TreeView) SelectNext()
func (tv *TreeView) SelectPrevious()
func (tv *TreeView) GetSelected() *ComponentSnapshot
```

Hierarchical component tree visualization with keyboard navigation.

**Example:**
```go
tree := devtools.NewTreeView(rootSnapshot)
tree.Expand("comp-123")
tree.SelectNext()
output := tree.Render()
```

---

### DetailPanel

```go
type DetailPanel struct { ... }

func NewDetailPanel() *DetailPanel
func (dp *DetailPanel) SetComponent(comp *ComponentSnapshot)
func (dp *DetailPanel) Render() string
func (dp *DetailPanel) SwitchTab(index int)
func (dp *DetailPanel) NextTab()
func (dp *DetailPanel) PreviousTab()
```

Component detail view with State/Props/Events tabs.

---

### ComponentInspector

```go
type ComponentInspector struct { ... }

func NewComponentInspector(root *ComponentSnapshot) *ComponentInspector
func (ci *ComponentInspector) Update(msg tea.Msg) tea.Cmd
func (ci *ComponentInspector) View() string
func (ci *ComponentInspector) SetRoot(root *ComponentSnapshot)
```

Integrates tree view, detail panel, and search for complete component inspection.

---

## State Viewer

### RefSnapshot

```go
type RefSnapshot struct {
    ID       string
    Name     string
    Type     string
    Value    interface{}
    Watchers int
}
```

Captures ref state at a point in time.

---

### StateViewer

```go
type StateViewer struct { ... }

func NewStateViewer(store *DevToolsStore) *StateViewer
func (sv *StateViewer) Render() string
func (sv *StateViewer) SelectRef(id string) bool
func (sv *StateViewer) GetSelected() *RefSnapshot
func (sv *StateViewer) ClearSelection()
func (sv *StateViewer) SetFilter(filter string)
func (sv *StateViewer) EditValue(id string, value interface{}) error
```

Displays all reactive state with history tracking.

**Example:**
```go
viewer := devtools.NewStateViewer(store)
viewer.SelectRef("count")
viewer.EditValue("count", 42)
```

---

### StateHistory

```go
type StateHistory struct { ... }

type StateChange struct {
    RefID     string
    RefName   string
    OldValue  interface{}
    NewValue  interface{}
    Timestamp time.Time
    Source    string
}

func NewStateHistory(maxSize int) *StateHistory
func (sh *StateHistory) Record(change StateChange)
func (sh *StateHistory) GetHistory(refID string) []StateChange
func (sh *StateHistory) GetAll() []StateChange
func (sh *StateHistory) Clear()
```

Tracks state changes over time with circular buffer.

---

## Event Tracker

### EventRecord

```go
type EventRecord struct {
    ID          string
    Name        string
    SourceID    string
    Payload     interface{}
    Timestamp   time.Time
}
```

Captures an event emission.

---

### EventTracker

```go
type EventTracker struct { ... }

func NewEventTracker(maxEvents int) *EventTracker
func (et *EventTracker) CaptureEvent(event EventRecord)
func (et *EventTracker) Pause()
func (et *EventTracker) Resume()
func (et *EventTracker) IsPaused() bool
func (et *EventTracker) GetEventCount() int
func (et *EventTracker) GetRecent(n int) []EventRecord
func (et *EventTracker) Clear()
func (et *EventTracker) GetStatistics() *EventStatistics
```

Captures and displays events with filtering.

**Example:**
```go
tracker := devtools.NewEventTracker(5000)
tracker.CaptureEvent(EventRecord{
    Name: "click",
    SourceID: "button-1",
})
stats := tracker.GetStatistics()
```

---

## Performance Monitor

### PerformanceData

```go
type PerformanceData struct {
    renderTimes map[string][]time.Duration
    avgTimes    map[string]time.Duration
    minTimes    map[string]time.Duration
    maxTimes    map[string]time.Duration
}

func NewPerformanceData() *PerformanceData
func (pd *PerformanceData) RecordRender(componentID string, duration time.Duration)
func (pd *PerformanceData) GetComponentStats(componentID string) ComponentPerformance
func (pd *PerformanceData) GetAllStats() map[string]ComponentPerformance
func (pd *PerformanceData) Clear()
```

Tracks component render performance.

**Example:**
```go
perfData := devtools.NewPerformanceData()
perfData.RecordRender("Counter", 2*time.Millisecond)
stats := perfData.GetComponentStats("Counter")
fmt.Printf("Avg: %v\n", stats.AvgRenderTime)
```

---

## Export System

### ExportOptions

```go
type ExportOptions struct {
    IncludeComponents  bool
    IncludeState       bool
    IncludeEvents      bool
    IncludePerformance bool
    IncludeTimestamps  bool
    Compress           bool
    CompressionLevel   int
    Format             ExportFormat
    Sanitize           *Sanitizer
    UseStreaming       bool
    ProgressCallback   func(bytesProcessed int64)
}
```

Options for controlling export behavior.

---

### Export

```go
func (dt *DevTools) Export(filename string, opts ExportOptions) error
```

Exports debug session to file with optional compression and sanitization.

**Example:**
```go
devtools.Enable().Export("debug.json.gz", devtools.ExportOptions{
    Compress:         true,
    CompressionLevel: gzip.BestCompression,
    IncludeState:     true,
    IncludeEvents:    true,
})
```

---

### ExportStream

```go
func (dt *DevTools) ExportStream(filename string, opts ExportOptions) error
```

Exports using streaming mode for large datasets (>100MB). Constant memory usage.

---

### ExportFormats

```go
type JSONFormat struct {}
type YAMLFormat struct {}
type MessagePackFormat struct {}
```

Supported export formats. MessagePack is smallest, YAML most readable.

---

### ExportFull / ExportIncremental

```go
func (dt *DevTools) ExportFull(filename string, opts ExportOptions) (*ExportCheckpoint, error)
func (dt *DevTools) ExportIncremental(filename string, checkpoint *ExportCheckpoint) (*ExportCheckpoint, error)
```

Incremental exports for long-running applications (saves 90%+ storage).

---

## Import System

### Import

```go
func (dt *DevTools) Import(filename string) error
```

Imports debug session. Auto-detects format and compression.

---

### ImportFromReader

```go
func (dt *DevTools) ImportFromReader(reader io.Reader) error
```

Imports from any io.Reader (network, memory buffer, etc.).

---

### ValidateImport

```go
func (dt *DevTools) ValidateImport(data *ExportData) error
```

Validates export data before importing.

---

## Sanitization

### Sanitizer

```go
type Sanitizer struct { ... }

func NewSanitizer() *Sanitizer
func (s *Sanitizer) AddPattern(pattern, replacement string)
func (s *Sanitizer) AddPatternWithPriority(pattern, replacement string, priority int, name string) error
func (s *Sanitizer) LoadTemplate(name string) error
func (s *Sanitizer) LoadTemplates(names ...string) error
func (s *Sanitizer) Sanitize(data *ExportData) *ExportData
func (s *Sanitizer) SanitizeValue(val interface{}) interface{}
func (s *Sanitizer) Preview(data *ExportData) *DryRunResult
func (s *Sanitizer) GetPatterns() []SanitizePattern
func (s *Sanitizer) GetLastStats() *SanitizationStats
```

Pattern-based sanitization with built-in compliance templates.

**Built-in Templates:**
- `pii` - SSN, email, phone
- `pci` - Credit cards, CVV
- `hipaa` - Medical records
- `gdpr` - IP/MAC addresses

**Example:**
```go
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "pci")
sanitizer.AddPatternWithPriority(
    `(?i)(api[_-]?key)(["'\s:=]+)([^\s"']+)`,
    "${1}${2}[REDACTED]",
    80,  // High priority
    "api_key",
)
cleanData := sanitizer.Sanitize(exportData)
```

---

### StreamSanitizer

```go
type StreamSanitizer struct { ... }

func NewStreamSanitizer(base *Sanitizer, bufferSize int) *StreamSanitizer
func (s *StreamSanitizer) SanitizeStream(reader io.Reader, writer io.Writer, progress func(int64)) error
```

Streaming sanitization for large files. Memory-bounded.

---

## Framework Hooks

### FrameworkHook Interface

```go
type FrameworkHook interface {
    OnComponentMount(id, name string)
    OnComponentUpdate(id string, msg interface{})
    OnComponentUnmount(id string)
    OnRefChange(id string, oldValue, newValue interface{})
    OnEvent(componentID, eventName string, data interface{})
    OnRenderComplete(componentID string, duration time.Duration)
    OnComputedChange(id string, oldValue, newValue interface{})
    OnWatchCallback(watcherID string, newValue, oldValue interface{})
    OnEffectRun(effectID string)
    OnChildAdded(parentID, childID string)
    OnChildRemoved(parentID, childID string)
}
```

Observe complete reactive cascade from Ref → Computed → Watchers → Effects.

**Example:**
```go
type MyHook struct{}

func (h *MyHook) OnRefChange(id string, oldVal, newVal interface{}) {
    fmt.Printf("Ref %s: %v → %v\n", id, oldVal, newVal)
}

// Implement all 11 methods...

devtools.RegisterHook(&MyHook{})
```

---

### RegisterHook

```go
func RegisterHook(hook FrameworkHook) error
```

Registers a framework hook (only one at a time).

---

### UnregisterHook

```go
func UnregisterHook() error
```

Removes the registered hook.

---

### IsHookRegistered

```go
func IsHookRegistered() bool
```

Returns true if a hook is registered.

---

## Configuration

### Config

```go
type Config struct {
    Enabled          bool
    LayoutMode       LayoutMode
    SplitRatio       float64
    MaxComponents    int
    MaxEvents        int
    MaxStateHistory  int
    SamplingRate     float64
}

func DefaultConfig() *Config
func (c *Config) Validate() error
func LoadConfig(path string) (*Config, error)
func (c *Config) ApplyEnvOverrides()
```

Configuration with file/environment variable support.

**Environment Variables:**
- `BUBBLY_DEVTOOLS_ENABLED`
- `BUBBLY_DEVTOOLS_LAYOUT_MODE`
- `BUBBLY_DEVTOOLS_SPLIT_RATIO`
- `BUBBLY_DEVTOOLS_MAX_COMPONENTS`
- `BUBBLY_DEVTOOLS_MAX_EVENTS`
- `BUBBLY_DEVTOOLS_MAX_STATE_HISTORY`
- `BUBBLY_DEVTOOLS_SAMPLING_RATE`

---

### LayoutMode

```go
const (
    LayoutHorizontal LayoutMode = iota  // Side by side
    LayoutVertical                      // Stacked
    LayoutOverlay                       // Tools on top
    LayoutHidden                        // Hidden
)
```

---

## Data Types

### DevToolsStore

```go
type DevToolsStore struct { ... }

func NewDevToolsStore() *DevToolsStore
func (s *DevToolsStore) AddComponent(snapshot *ComponentSnapshot)
func (s *DevToolsStore) GetComponent(id string) *ComponentSnapshot
func (s *DevToolsStore) GetAllComponents() []*ComponentSnapshot
func (s *DevToolsStore) RecordStateChange(change StateChange)
func (s *DevToolsStore) GetStateHistory(refID string) []StateChange
func (s *DevToolsStore) RecordEvent(event EventRecord)
func (s *DevToolsStore) GetEvents() []EventRecord
func (s *DevToolsStore) Clear()
```

In-memory storage for all collected debug data.

---

### DataCollector

```go
type DataCollector struct { ... }

type ComponentHook interface {
    OnComponentCreated(*ComponentSnapshot)
    OnComponentMounted(string)
    OnComponentUpdated(string)
    OnComponentUnmounted(string)
}

func NewDataCollector() *DataCollector
func (dc *DataCollector) AddComponentHook(hook ComponentHook)
func (dc *DataCollector) RemoveComponentHook(hook ComponentHook)
func (dc *DataCollector) FireComponentMounted(id string)
```

Hooks into application for data collection.

---

## Performance Characteristics

| Operation | Target | Actual |
|-----------|--------|--------|
| Enable/Disable | < 100ms | ~50ms |
| Inspector Render | < 50ms | ~30ms |
| State Update | < 10ms | ~5ms |
| Search | < 100ms | ~50ms |
| Overhead (enabled) | < 5% | ~3% |
| Overhead (disabled) | 0% | 0% |

---

## Thread Safety

All types and functions are thread-safe:

- Singleton initialization: `sync.Once`
- Mutations: `sync.RWMutex` protected
- Copy-on-read: Prevents data races
- Hook execution: Isolated from application

---

## Error Handling

Dev tools never crash the host application:

- Panics in hooks: Recovered and reported
- Collection errors: Logged, not fatal
- Memory limits: Graceful degradation
- Validation errors: Clear messages

---

## See Also

- **Package Documentation:** `godoc github.com/newbpydev/bubblyui/pkg/bubbly/devtools`
- **User Guide:** `docs/devtools/README.md`
- **Examples:** `cmd/examples/09-devtools/`
- **Spec:** `specs/09-dev-tools/`
