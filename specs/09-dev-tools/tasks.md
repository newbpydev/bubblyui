# Implementation Tasks: Dev Tools

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] 01-reactivity-system completed (State inspection)
- [x] 02-component-model completed (Component inspection)
- [x] 03-lifecycle-hooks completed (Lifecycle tracking)
- [ ] Lipgloss advanced layouts understood
- [ ] Bubbletea split-pane patterns established

---

## Phase 1: Core Infrastructure (5 tasks, 15 hours)

### Task 1.1: DevTools Manager ✅ COMPLETED
**Description**: Main dev tools singleton and lifecycle management

**Prerequisites**: None

**Unlocks**: Task 1.2 (Data Collector)

**Files**:
- `pkg/bubbly/devtools/devtools.go`
- `pkg/bubbly/devtools/devtools_test.go`

**Type Safety**:
```go
type DevTools struct {
    enabled   bool
    visible   bool
    collector *DataCollector
    store     *DevToolsStore
    ui        *DevToolsUI
    config    *Config
    mu        sync.RWMutex
}

func Enable() *DevTools
func Disable()
func Toggle()
func IsEnabled() bool
```

**Tests**:
- [x] Enable/disable works
- [x] Toggle visibility
- [x] Singleton pattern
- [x] Thread-safe access
- [x] Lifecycle management

**Estimated Effort**: 3 hours

**Implementation Notes**:
- ✅ Implemented singleton pattern using `sync.Once` for thread-safe initialization
- ✅ Package-level functions (Enable, Disable, Toggle, IsEnabled) operate on global singleton
- ✅ Instance methods (SetVisible, IsVisible, ToggleVisibility) for UI control
- ✅ All methods use `sync.RWMutex` for thread-safe concurrent access
- ✅ 100% test coverage with 9 tests including concurrent access tests
- ✅ All tests pass with race detector
- ✅ Zero lint warnings
- ✅ Comprehensive godoc comments on all exported types and functions
- ✅ Follows existing BubblyUI patterns (similar to observability/reporter.go)
- ✅ Placeholder comments for collector, store, ui, config (implemented in Tasks 1.2-1.5)
- ✅ Actual time: ~2 hours (under estimate)

---

### Task 1.2: Data Collector ✅ COMPLETED
**Description**: Hook system for collecting data from application

**Prerequisites**: Task 1.1

**Unlocks**: Task 1.3 (Data Store)

**Files**:
- `pkg/bubbly/devtools/collector.go`
- `pkg/bubbly/devtools/collector_test.go`

**Type Safety**:
```go
type DataCollector struct {
    componentHooks []ComponentHook
    stateHooks     []StateHook
    eventHooks     []EventHook
    perfHooks      []PerformanceHook
}

type ComponentHook interface {
    OnComponentCreated(*ComponentSnapshot)
    OnComponentMounted(string)
    OnComponentUpdated(string)
    OnComponentUnmounted(string)
}
```

**Tests**:
- [x] Hooks register correctly
- [x] Hooks fire at right time
- [x] Data captured accurately
- [x] Hook removal works
- [x] No performance regression

**Estimated Effort**: 3 hours

**Implementation Notes**:
- ✅ Implemented observer pattern with 4 hook types (Component, State, Event, Performance)
- ✅ Snapshot types defined: ComponentSnapshot, RefSnapshot, EventRecord
- ✅ Thread-safe hook management using `sync.RWMutex`
- ✅ Copy-on-read pattern prevents holding lock during hook execution
- ✅ Panic recovery in all Fire methods with observability integration
- ✅ Hooks isolated from application - panics don't crash host app
- ✅ 12 tests for collector + 9 tests for devtools = 21 total tests passing
- ✅ 84.6% test coverage (exceeds 80% requirement)
- ✅ All tests pass with race detector
- ✅ Zero lint warnings
- ✅ Comprehensive godoc on all exported types, interfaces, and functions
- ✅ Add/Remove methods for all hook types
- ✅ Fire methods for all lifecycle events
- ✅ Mock hooks in tests are thread-safe for concurrent testing
- ✅ Actual time: ~2.5 hours (under estimate)

---

### Task 1.3: Data Store ✅ COMPLETED
**Description**: In-memory storage for collected debug data

**Prerequisites**: Task 1.2

**Unlocks**: Task 1.4 (Instrumentation)

**Files**:
- `pkg/bubbly/devtools/store.go`
- `pkg/bubbly/devtools/store_test.go`

**Type Safety**:
```go
type DevToolsStore struct {
    components   map[string]*ComponentSnapshot
    stateHistory *StateHistory
    events       *EventLog
    performance  *PerformanceData
    commands     *CommandTimeline
    mu           sync.RWMutex
}

func (s *DevToolsStore) AddComponent(*ComponentSnapshot)
func (s *DevToolsStore) GetComponent(id string) *ComponentSnapshot
func (s *DevToolsStore) GetAllComponents() []*ComponentSnapshot
```

**Tests**:
- [x] Add/get components
- [x] State history tracking
- [x] Event logging
- [x] Performance data
- [x] Thread-safe operations
- [x] Memory limits enforced

**Estimated Effort**: 4 hours

**Implementation Notes**:
- ✅ Implemented four core data structures: StateHistory, EventLog, PerformanceData, DevToolsStore
- ✅ All structures use `sync.RWMutex` for thread-safe concurrent access
- ✅ Circular buffer pattern for StateHistory and EventLog with configurable max sizes
- ✅ Copy-on-read pattern prevents external modification of internal state
- ✅ PerformanceData tracks min/max/avg render times with automatic calculation
- ✅ DevToolsStore provides unified access to all data subsystems
- ✅ 27 comprehensive tests covering all operations and edge cases
- ✅ 91.0% test coverage (exceeds 80% requirement)
- ✅ All tests pass with race detector (`-race` flag)
- ✅ Zero lint warnings (go vet clean)
- ✅ Code formatted with gofmt
- ✅ Builds successfully
- ✅ Comprehensive godoc comments on all exported types and functions
- ✅ Table-driven tests for all major functionality
- ✅ Concurrent access tests verify thread safety under load
- ✅ Memory limits enforced via circular buffer trimming
- ✅ Clear separation of concerns: StateHistory, EventLog, PerformanceData, DevToolsStore
- ✅ Actual time: ~3 hours (under estimate)

---

### Task 1.4: Instrumentation System
**Description**: Install hooks into application code

**Prerequisites**: Task 1.3

**Unlocks**: Task 1.5 (Configuration)

**Files**:
- `pkg/bubbly/devtools/instrumentation.go`
- `pkg/bubbly/devtools/instrumentation_test.go`

**Type Safety**:
```go
type Instrumentor struct {
    hooks *HookRegistry
}

func (i *Instrumentor) InstrumentComponent(*componentImpl)
func (i *Instrumentor) InstrumentRef(*Ref[T])
func (i *Instrumentor) InstrumentRouter(*Router)
```

**Tests**:
- [ ] Components instrumented
- [ ] Refs instrumented
- [ ] Router instrumented
- [ ] Hooks don't break app
- [ ] Overhead < 5%

**Estimated Effort**: 3 hours

---

### Task 1.5: Configuration System
**Description**: Dev tools configuration and options

**Prerequisites**: Task 1.4

**Unlocks**: Task 2.1 (Component Inspector)

**Files**:
- `pkg/bubbly/devtools/config.go`
- `pkg/bubbly/devtools/config_test.go`

**Type Safety**:
```go
type Config struct {
    Enabled        bool
    LayoutMode     LayoutMode
    SplitRatio     float64
    MaxComponents  int
    MaxEvents      int
    MaxStateHistory int
    SamplingRate   float64
}

func DefaultConfig() *Config
func LoadConfig(path string) (*Config, error)
```

**Tests**:
- [ ] Default config works
- [ ] Config loading
- [ ] Validation
- [ ] Override options
- [ ] Environment variables

**Estimated Effort**: 2 hours

---

## Phase 2: Component Inspector (6 tasks, 18 hours)

### Task 2.1: Component Snapshot
**Description**: Capture component state at a point in time

**Prerequisites**: Task 1.5

**Unlocks**: Task 2.2 (Tree View)

**Files**:
- `pkg/bubbly/devtools/snapshot.go`
- `pkg/bubbly/devtools/snapshot_test.go`

**Type Safety**:
```go
type ComponentSnapshot struct {
    ID         string
    Name       string
    Type       string
    Parent     *ComponentSnapshot
    Children   []*ComponentSnapshot
    State      map[string]interface{}
    Props      map[string]interface{}
    Refs       []*RefSnapshot
    Timestamp  time.Time
}

func CaptureComponent(*componentImpl) *ComponentSnapshot
```

**Tests**:
- [ ] Snapshot captures all data
- [ ] Nested components work
- [ ] State serialization
- [ ] Props captured
- [ ] Refs snapshot correctly

**Estimated Effort**: 3 hours

---

### Task 2.2: Component Tree View
**Description**: Hierarchical component tree display

**Prerequisites**: Task 2.1

**Unlocks**: Task 2.3 (Detail Panel)

**Files**:
- `pkg/bubbly/devtools/tree_view.go`
- `pkg/bubbly/devtools/tree_view_test.go`

**Type Safety**:
```go
type TreeView struct {
    root     *ComponentSnapshot
    selected *ComponentSnapshot
    expanded map[string]bool
    viewport *Viewport
}

func (tv *TreeView) Render() string
func (tv *TreeView) Select(id string)
func (tv *TreeView) Expand(id string)
func (tv *TreeView) Collapse(id string)
```

**Tests**:
- [ ] Tree renders correctly
- [ ] Selection works
- [ ] Expand/collapse
- [ ] Navigation (up/down)
- [ ] Large trees perform well

**Estimated Effort**: 4 hours

---

### Task 2.3: Detail Panel
**Description**: Component detail view with tabs

**Prerequisites**: Task 2.2

**Unlocks**: Task 2.4 (Search Widget)

**Files**:
- `pkg/bubbly/devtools/detail_panel.go`
- `pkg/bubbly/devtools/detail_panel_test.go`

**Type Safety**:
```go
type DetailPanel struct {
    component *ComponentSnapshot
    tabs      []Tab
    activeTab int
}

type Tab struct {
    Name    string
    Render  func(*ComponentSnapshot) string
}

func (dp *DetailPanel) Render() string
func (dp *DetailPanel) SwitchTab(index int)
```

**Tests**:
- [ ] Tabs render
- [ ] Tab switching
- [ ] State tab shows refs
- [ ] Props tab shows properties
- [ ] Events tab shows history

**Estimated Effort**: 3 hours

---

### Task 2.4: Search Widget
**Description**: Search components by name/type

**Prerequisites**: Task 2.3

**Unlocks**: Task 2.5 (State Viewer)

**Files**:
- `pkg/bubbly/devtools/search.go`
- `pkg/bubbly/devtools/search_test.go`

**Type Safety**:
```go
type SearchWidget struct {
    query   string
    results []*ComponentSnapshot
    cursor  int
}

func (sw *SearchWidget) Search(query string)
func (sw *SearchWidget) NextResult()
func (sw *SearchWidget) PrevResult()
func (sw *SearchWidget) Render() string
```

**Tests**:
- [ ] Search finds components
- [ ] Fuzzy matching
- [ ] Result navigation
- [ ] Performance (large trees)
- [ ] Clear search

**Estimated Effort**: 2 hours

---

### Task 2.5: Component Filter
**Description**: Filter components by type/status

**Prerequisites**: Task 2.4

**Unlocks**: Task 2.6 (Inspector Integration)

**Files**:
- `pkg/bubbly/devtools/filter.go`
- `pkg/bubbly/devtools/filter_test.go`

**Type Safety**:
```go
type ComponentFilter struct {
    types    []string
    statuses []ComponentStatus
    custom   FilterFunc
}

type FilterFunc func(*ComponentSnapshot) bool

func (cf *ComponentFilter) Apply([]*ComponentSnapshot) []*ComponentSnapshot
```

**Tests**:
- [ ] Type filtering
- [ ] Status filtering
- [ ] Custom filters
- [ ] Multiple filters combine
- [ ] Performance

**Estimated Effort**: 2 hours

---

### Task 2.6: Inspector Integration
**Description**: Complete component inspector panel

**Prerequisites**: Task 2.5

**Unlocks**: Task 3.1 (State Viewer)

**Files**:
- `pkg/bubbly/devtools/inspector.go`
- `pkg/bubbly/devtools/inspector_test.go`

**Type Safety**:
```go
type ComponentInspector struct {
    tree    *TreeView
    detail  *DetailPanel
    search  *SearchWidget
    filter  *ComponentFilter
}

func (ci *ComponentInspector) Update(msg tea.Msg) tea.Cmd
func (ci *ComponentInspector) View() string
```

**Tests**:
- [ ] All parts integrate
- [ ] Keyboard navigation
- [ ] Live updates
- [ ] Performance acceptable
- [ ] E2E inspector test

**Estimated Effort**: 4 hours

---

## Phase 3: State & Event Tracking (5 tasks, 15 hours)

### Task 3.1: State Viewer
**Description**: Display all reactive state

**Prerequisites**: Task 2.6

**Unlocks**: Task 3.2 (State History)

**Files**:
- `pkg/bubbly/devtools/state_viewer.go`
- `pkg/bubbly/devtools/state_viewer_test.go`

**Type Safety**:
```go
type StateViewer struct {
    store    *DevToolsStore
    selected *RefSnapshot
    filter   string
}

func (sv *StateViewer) Render() string
func (sv *StateViewer) SelectRef(id string)
func (sv *StateViewer) EditValue(id string, value interface{})
```

**Tests**:
- [ ] All state displayed
- [ ] Selection works
- [ ] Value editing
- [ ] Type display correct
- [ ] Filtering works

**Estimated Effort**: 3 hours

---

### Task 3.2: State History Tracking
**Description**: Track state changes over time

**Prerequisites**: Task 3.1

**Unlocks**: Task 3.3 (Event Tracker)

**Files**:
- `pkg/bubbly/devtools/state_history.go`
- `pkg/bubbly/devtools/state_history_test.go`

**Type Safety**:
```go
type StateHistory struct {
    changes []StateChange
    maxSize int
    mu      sync.RWMutex
}

type StateChange struct {
    RefID     string
    OldValue  interface{}
    NewValue  interface{}
    Timestamp time.Time
    Source    string
}

func (sh *StateHistory) Record(StateChange)
func (sh *StateHistory) GetHistory(refID string) []StateChange
```

**Tests**:
- [ ] Changes recorded
- [ ] History retrieved
- [ ] Max size enforced
- [ ] Thread-safe
- [ ] Performance acceptable

**Estimated Effort**: 3 hours

---

### Task 3.3: Event Tracker
**Description**: Capture and display events

**Prerequisites**: Task 3.2

**Unlocks**: Task 3.4 (Event Filter)

**Files**:
- `pkg/bubbly/devtools/event_tracker.go`
- `pkg/bubbly/devtools/event_tracker_test.go`

**Type Safety**:
```go
type EventTracker struct {
    events    *EventLog
    filter    EventFilter
    paused    bool
    maxEvents int
}

type EventRecord struct {
    ID        string
    Name      string
    Source    string
    Target    string
    Payload   interface{}
    Timestamp time.Time
    Duration  time.Duration
}

func (et *EventTracker) CaptureEvent(EventRecord)
func (et *EventTracker) Render() string
```

**Tests**:
- [ ] Events captured
- [ ] Real-time display
- [ ] Pause/resume
- [ ] Event details shown
- [ ] Performance acceptable

**Estimated Effort**: 3 hours

---

### Task 3.4: Event Filter & Search
**Description**: Filter and search events

**Prerequisites**: Task 3.3

**Unlocks**: Task 3.5 (Event Replay)

**Files**:
- `pkg/bubbly/devtools/event_filter.go`
- `pkg/bubbly/devtools/event_filter_test.go`

**Type Safety**:
```go
type EventFilter struct {
    names      []string
    sources    []string
    timeRange  *TimeRange
}

func (ef *EventFilter) Matches(EventRecord) bool
func (ef *EventFilter) Apply([]EventRecord) []EventRecord
```

**Tests**:
- [ ] Name filtering
- [ ] Source filtering
- [ ] Time range
- [ ] Multiple filters
- [ ] Search works

**Estimated Effort**: 2 hours

---

### Task 3.5: Event Replay
**Description**: Replay captured events

**Prerequisites**: Task 3.4

**Unlocks**: Task 4.1 (Performance Monitor)

**Files**:
- `pkg/bubbly/devtools/event_replay.go`
- `pkg/bubbly/devtools/event_replay_test.go`

**Type Safety**:
```go
type EventReplayer struct {
    events []EventRecord
    speed  float64
}

func (er *EventReplayer) Replay() tea.Cmd
func (er *EventReplayer) SetSpeed(float64)
func (er *EventReplayer) Pause()
func (er *EventReplayer) Resume()
```

**Tests**:
- [ ] Events replay correctly
- [ ] Speed control works
- [ ] Pause/resume
- [ ] Event order preserved
- [ ] Integration with app

**Estimated Effort**: 4 hours

---

## Phase 4: Performance & Router Debugging (5 tasks, 15 hours)

### Task 4.1: Performance Monitor
**Description**: Track component performance metrics

**Prerequisites**: Task 3.5

**Unlocks**: Task 4.2 (Flame Graph)

**Files**:
- `pkg/bubbly/devtools/performance.go`
- `pkg/bubbly/devtools/performance_test.go`

**Type Safety**:
```go
type PerformanceMonitor struct {
    data      *PerformanceData
    collector *MetricsCollector
}

type ComponentPerformance struct {
    ComponentID   string
    RenderCount   int64
    AvgRenderTime time.Duration
    MaxRenderTime time.Duration
    MemoryUsage   uint64
}

func (pm *PerformanceMonitor) RecordRender(string, time.Duration)
func (pm *PerformanceMonitor) Render() string
```

**Tests**:
- [ ] Metrics collected
- [ ] Averages calculated correctly
- [ ] Display formatted
- [ ] Sorting works
- [ ] Overhead < 2%

**Estimated Effort**: 3 hours

---

### Task 4.2: Flame Graph Renderer
**Description**: Visual flame graph for performance

**Prerequisites**: Task 4.1

**Unlocks**: Task 4.3 (Router Debugger)

**Files**:
- `pkg/bubbly/devtools/flame_graph.go`
- `pkg/bubbly/devtools/flame_graph_test.go`

**Type Safety**:
```go
type FlameGraphRenderer struct {
    width  int
    height int
}

type FlameNode struct {
    Name           string
    Time           time.Duration
    TimePercentage float64
    Children       []*FlameNode
}

func (fgr *FlameGraphRenderer) Render(*PerformanceData) string
```

**Tests**:
- [ ] Flame graph renders
- [ ] Percentages correct
- [ ] Colors/styling applied
- [ ] Interactive (select nodes)
- [ ] Readable at different sizes

**Estimated Effort**: 4 hours

---

### Task 4.3: Router Debugger
**Description**: Debug route navigation

**Prerequisites**: Task 4.2

**Unlocks**: Task 4.4 (Command Timeline)

**Files**:
- `pkg/bubbly/devtools/router_debugger.go`
- `pkg/bubbly/devtools/router_debugger_test.go`

**Type Safety**:
```go
type RouterDebugger struct {
    currentRoute *Route
    history      []RouteRecord
    guards       []GuardExecution
}

type RouteRecord struct {
    From      *Route
    To        *Route
    Timestamp time.Time
    Duration  time.Duration
    Success   bool
}

func (rd *RouterDebugger) Render() string
```

**Tests**:
- [ ] Current route shown
- [ ] History tracked
- [ ] Guard execution traced
- [ ] Failed navigation logged
- [ ] Integration with router

**Estimated Effort**: 3 hours

---

### Task 4.4: Command Timeline
**Description**: Visualize command execution

**Prerequisites**: Task 4.3

**Unlocks**: Task 4.5 (Timeline Controls)

**Files**:
- `pkg/bubbly/devtools/command_timeline.go`
- `pkg/bubbly/devtools/command_timeline_test.go`

**Type Safety**:
```go
type CommandTimeline struct {
    commands []CommandRecord
    paused   bool
    maxSize  int
}

type CommandRecord struct {
    ID        string
    Type      string
    Source    string
    Generated time.Time
    Executed  time.Time
    Duration  time.Duration
}

func (ct *CommandTimeline) RecordCommand(CommandRecord)
func (ct *CommandTimeline) Render(width int) string
```

**Tests**:
- [ ] Commands recorded
- [ ] Timeline visualization
- [ ] Batching shown
- [ ] Timing accurate
- [ ] Performance acceptable

**Estimated Effort**: 3 hours

---

### Task 4.5: Timeline Scrubbing & Replay
**Description**: Scrub timeline, replay commands

**Prerequisites**: Task 4.4

**Unlocks**: Task 5.1 (Layout Manager)

**Files**:
- `pkg/bubbly/devtools/timeline_controls.go`
- `pkg/bubbly/devtools/timeline_controls_test.go`

**Type Safety**:
```go
type TimelineControls struct {
    timeline *CommandTimeline
    position time.Time
}

func (tc *TimelineControls) Scrub(time.Time)
func (tc *TimelineControls) Replay()
func (tc *TimelineControls) Render() string
```

**Tests**:
- [ ] Scrubbing works
- [ ] Position indicator
- [ ] Replay functional
- [ ] Speed control
- [ ] Integration

**Estimated Effort**: 2 hours

---

## Phase 5: UI & Layout (4 tasks, 12 hours)

### Task 5.1: Layout Manager
**Description**: Split-pane and layout management

**Prerequisites**: Task 4.5

**Unlocks**: Task 5.2 (Tab Controller)

**Files**:
- `pkg/bubbly/devtools/layout.go`
- `pkg/bubbly/devtools/layout_test.go`

**Type Safety**:
```go
type LayoutManager struct {
    mode   LayoutMode
    ratio  float64
    width  int
    height int
}

type LayoutMode int

const (
    LayoutHorizontal LayoutMode = iota
    LayoutVertical
    LayoutOverlay
    LayoutHidden
)

func (lm *LayoutManager) Render(app, tools string) string
```

**Tests**:
- [ ] Horizontal split works
- [ ] Vertical split works
- [ ] Overlay mode works
- [ ] Ratio adjustment
- [ ] Responsive to resize

**Estimated Effort**: 3 hours

---

### Task 5.2: Tab Controller
**Description**: Tab navigation in dev tools

**Prerequisites**: Task 5.1

**Unlocks**: Task 5.3 (Keyboard Handler)

**Files**:
- `pkg/bubbly/devtools/tabs.go`
- `pkg/bubbly/devtools/tabs_test.go`

**Type Safety**:
```go
type TabController struct {
    tabs      []Tab
    activeTab int
}

type Tab struct {
    Name    string
    Content func() string
}

func (tc *TabController) Next()
func (tc *TabController) Prev()
func (tc *TabController) Select(int)
func (tc *TabController) Render() string
```

**Tests**:
- [ ] Tab switching
- [ ] Active tab highlighted
- [ ] Keyboard navigation
- [ ] Content renders
- [ ] Multiple tab groups

**Estimated Effort**: 2 hours

---

### Task 5.3: Keyboard Handler
**Description**: Dev tools keyboard shortcuts

**Prerequisites**: Task 5.2

**Unlocks**: Task 5.4 (DevTools UI)

**Files**:
- `pkg/bubbly/devtools/keyboard.go`
- `pkg/bubbly/devtools/keyboard_test.go`

**Type Safety**:
```go
type KeyboardHandler struct {
    shortcuts map[string]KeyHandler
    focus     FocusTarget
}

type KeyHandler func(tea.KeyMsg) tea.Cmd

func (kh *KeyboardHandler) Handle(tea.KeyMsg) tea.Cmd
func (kh *KeyboardHandler) Register(string, KeyHandler)
```

**Tests**:
- [ ] F12 toggle works
- [ ] Tab switching
- [ ] Navigation keys
- [ ] Search shortcut
- [ ] Help dialog (?)

**Estimated Effort**: 3 hours

---

### Task 5.4: DevTools UI Integration
**Description**: Complete UI assembly

**Prerequisites**: Task 5.3

**Unlocks**: Task 6.1 (Export System)

**Files**:
- `pkg/bubbly/devtools/ui.go`
- `pkg/bubbly/devtools/ui_test.go`

**Type Safety**:
```go
type DevToolsUI struct {
    layout     *LayoutManager
    inspector  *ComponentInspector
    state      *StateViewer
    events     *EventTracker
    perf       *PerformanceMonitor
    router     *RouterDebugger
    timeline   *CommandTimeline
    keyboard   *KeyboardHandler
    activePanel int
}

func (ui *DevToolsUI) Update(tea.Msg) tea.Cmd
func (ui *DevToolsUI) View() string
```

**Tests**:
- [ ] All panels integrate
- [ ] Panel switching
- [ ] Layout changes
- [ ] Keyboard shortcuts
- [ ] E2E UI test

**Estimated Effort**: 4 hours

---

## Phase 6: Data Management (3 tasks, 9 hours)

### Task 6.1: Export System
**Description**: Export debug data to JSON

**Prerequisites**: Task 5.4

**Unlocks**: Task 6.2 (Import System)

**Files**:
- `pkg/bubbly/devtools/export.go`
- `pkg/bubbly/devtools/export_test.go`

**Type Safety**:
```go
type ExportData struct {
    Version     string
    Timestamp   time.Time
    Components  []*ComponentSnapshot
    State       []StateChange
    Events      []EventRecord
    Performance *PerformanceData
}

type ExportOptions struct {
    IncludeComponents bool
    IncludeState      bool
    IncludeEvents     bool
    Sanitize          bool
    RedactPatterns    []string
}

func (dt *DevTools) Export(filename string, opts ExportOptions) error
```

**Tests**:
- [ ] Export creates file
- [ ] JSON valid
- [ ] All data included
- [ ] Sanitization works
- [ ] Large exports handle

**Estimated Effort**: 3 hours

---

### Task 6.2: Import System
**Description**: Import debug data from JSON

**Prerequisites**: Task 6.1

**Unlocks**: Task 6.3 (Data Sanitization)

**Files**:
- `pkg/bubbly/devtools/import.go`
- `pkg/bubbly/devtools/import_test.go`

**Type Safety**:
```go
func (dt *DevTools) Import(filename string) error
func (dt *DevTools) ImportFromReader(io.Reader) error
func (dt *DevTools) ValidateImport(data *ExportData) error
```

**Tests**:
- [ ] Import loads file
- [ ] Data restored
- [ ] Validation works
- [ ] Invalid data handled
- [ ] Round-trip test

**Estimated Effort**: 3 hours

---

### Task 6.3: Data Sanitization
**Description**: Remove sensitive data from exports

**Prerequisites**: Task 6.2

**Unlocks**: Task 7.1 (Documentation)

**Files**:
- `pkg/bubbly/devtools/sanitize.go`
- `pkg/bubbly/devtools/sanitize_test.go`

**Type Safety**:
```go
type Sanitizer struct {
    patterns []SanitizePattern
}

type SanitizePattern struct {
    Pattern     *regexp.Regexp
    Replacement string
}

func (s *Sanitizer) Sanitize(data *ExportData) *ExportData
func (s *Sanitizer) SanitizeValue(interface{}) interface{}
```

**Tests**:
- [ ] Passwords redacted
- [ ] Tokens redacted
- [ ] API keys redacted
- [ ] Custom patterns work
- [ ] Nested data handled

**Estimated Effort**: 3 hours

---

## Phase 7: Documentation & Polish (3 tasks, 9 hours)

### Task 7.1: API Documentation
**Description**: Comprehensive godoc for dev tools

**Prerequisites**: Task 6.3

**Unlocks**: Task 7.2 (User Guide)

**Files**:
- All package files (add/update godoc)

**Documentation**:
- DevTools API
- Collector hooks
- Inspector interfaces
- Export/import format
- Configuration options

**Estimated Effort**: 2 hours

---

### Task 7.2: User Guide
**Description**: Complete user documentation

**Prerequisites**: Task 7.1

**Unlocks**: Task 7.3 (Examples)

**Files**:
- `docs/devtools/README.md`
- `docs/devtools/quickstart.md`
- `docs/devtools/reference.md`
- `docs/devtools/troubleshooting.md`

**Content**:
- Getting started
- Keyboard shortcuts
- Feature overview
- Tips and tricks
- Common issues

**Estimated Effort**: 4 hours

---

### Task 7.3: Example Integration
**Description**: Dev tools examples

**Prerequisites**: Task 7.2

**Unlocks**: Feature complete

**Files**:
- `cmd/examples/09-devtools/basic/main.go`
- `cmd/examples/09-devtools/debugging/main.go`
- `cmd/examples/09-devtools/performance/main.go`

**Examples**:
- Basic enablement
- Debugging workflow
- Performance profiling
- Export/import

**Estimated Effort**: 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01-03, 07, 08)
    ↓
Phase 1: Infrastructure
    1.1 Manager → 1.2 Collector → 1.3 Store → 1.4 Instrumentation → 1.5 Config
    ↓
Phase 2: Inspector
    2.1 Snapshot → 2.2 Tree → 2.3 Detail → 2.4 Search → 2.5 Filter → 2.6 Integration
    ↓
Phase 3: State & Events
    3.1 State Viewer → 3.2 History → 3.3 Tracker → 3.4 Filter → 3.5 Replay
    ↓
Phase 4: Performance & Router
    4.1 Perf Monitor → 4.2 Flame Graph → 4.3 Router → 4.4 Timeline → 4.5 Controls
    ↓
Phase 5: UI & Layout
    5.1 Layout → 5.2 Tabs → 5.3 Keyboard → 5.4 UI Integration
    ↓
Phase 6: Data
    6.1 Export → 6.2 Import → 6.3 Sanitization
    ↓
Phase 7: Documentation
    7.1 API Docs → 7.2 User Guide → 7.3 Examples
```

---

## Validation Checklist

### Core Functionality
- [ ] Enable/disable works
- [ ] Instrumentation captures data
- [ ] All panels render
- [ ] Navigation works
- [ ] Export/import works

### Performance
- [ ] Overhead < 5% when enabled
- [ ] No overhead when disabled
- [ ] Large apps handle well
- [ ] Responsive UI
- [ ] Memory limits enforced

### Safety
- [ ] Never crashes host app
- [ ] Isolated error handling
- [ ] Safe state editing
- [ ] Data sanitization works
- [ ] Graceful degradation

### Usability
- [ ] Intuitive navigation
- [ ] Clear visuals
- [ ] Helpful shortcuts
- [ ] Good documentation
- [ ] Discoverable features

### Integration
- [ ] Works with all features
- [ ] Compatible with Bubbletea
- [ ] Terminal size adaptive
- [ ] Keyboard-driven
- [ ] E2E workflows tested

---

## Estimated Total Effort

- Phase 1: 15 hours
- Phase 2: 18 hours
- Phase 3: 15 hours
- Phase 4: 15 hours
- Phase 5: 12 hours
- Phase 6: 9 hours
- Phase 7: 9 hours

**Total**: ~93 hours (approximately 2.5 weeks)

---

## Priority

**HIGH** - Critical for developer experience and framework adoption. Essential debugging tool.

**Timeline**: Implement after Features 01-08 complete, alongside or after Feature 07 (Router).

**Unlocks**:
- Improved debugging experience
- Faster issue resolution
- Learning tool for new users
- Framework transparency
- Community contributions
- Feature 11 (Performance Profiler) data source
