# Design Specification: Dev Tools

## Component Hierarchy

```
Developer Tools System
└── DevTools Manager
    ├── Component Inspector
    │   ├── Tree View
    │   ├── Detail Panel
    │   └── Search Widget
    ├── State Viewer
    │   ├── State Tree
    │   ├── Value Editor
    │   └── History Timeline
    ├── Event Tracker
    │   ├── Event List
    │   ├── Event Detail
    │   └── Event Filter
    ├── Router Debugger
    │   ├── Route Info
    │   ├── History Stack
    │   └── Guard Trace
    ├── Performance Monitor
    │   ├── Metrics Display
    │   ├── Flame Graph
    │   └── Timeline View
    ├── Command Timeline
    │   ├── Command List
    │   ├── Batch Visualization
    │   └── Replay Controls
    └── Layout Manager
        ├── Split Pane
        ├── Tab Controller
        └── Keyboard Handler
```

---

## Architecture Overview

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                   Application Process                         │
├──────────────────────────────────────────────────────────────┤
│  ┌────────────────────┐        ┌──────────────────────────┐  │
│  │   User App         │───────→│   Dev Tools Collector    │  │
│  │   (BubblyUI)       │        │   (Instrumentation)      │  │
│  └────────────────────┘        └──────────┬───────────────┘  │
│                                            │                  │
│                                            ↓                  │
│                                 ┌──────────────────────────┐ │
│                                 │   Dev Tools Store        │ │
│                                 │   (In-memory DB)         │ │
│                                 └──────────┬───────────────┘ │
│                                            │                  │
│                                            ↓                  │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Dev Tools UI (Split Pane)                   ││
│  │  ┌────────────────────┬─────────────────────────────┐   ││
│  │  │  Inspector Panels  │   Layout Manager            │   ││
│  │  └────────────────────┴─────────────────────────────┘   ││
│  └─────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────┘
                              │
                              ↓
                    ┌─────────────────────┐
                    │  Terminal Display   │
                    └─────────────────────┘
```

---

## Data Flow

### Instrumentation Flow

```
Application Event (State Change, Component Update, etc.)
    ↓
DevTools Hook Intercepts
    ↓
Data Collected
    ↓
Stored in DevTools Store
    ↓
UI Notified of Update
    ↓
Inspector Panels Refresh
    ↓
Developer Sees Change
```

### Inspection Flow

```
Developer Selects Component
    ↓
Inspector Queries Store
    ↓
Component Data Retrieved
    ↓
Detail Panel Displays Info
    ↓
Developer Can:
    - View State
    - Edit Values
    - View Events
    - See Performance
```

---

## Type Definitions

### Core Types

```go
// DevTools is the main dev tools instance
type DevTools struct {
    enabled      bool
    visible      bool
    collector    *DataCollector
    store        *DevToolsStore
    ui           *DevToolsUI
    config       *DevToolsConfig
    mu           sync.RWMutex
}

// DataCollector hooks into application
type DataCollector struct {
    componentHooks []ComponentHook
    stateHooks     []StateHook
    eventHooks     []EventHook
    routeHooks     []RouteHook
    perfHooks      []PerformanceHook
}

// DevToolsStore holds collected data
type DevToolsStore struct {
    components      map[string]*ComponentSnapshot
    stateHistory    *StateHistory
    events          *EventLog
    routes          *RouteHistory
    performance     *PerformanceData
    commands        *CommandTimeline
    mu              sync.RWMutex
}

// ComponentSnapshot captures component state at a point in time
type ComponentSnapshot struct {
    ID           string
    Name         string
    Type         string
    Parent       *ComponentSnapshot
    Children     []*ComponentSnapshot
    State        map[string]interface{}
    Props        map[string]interface{}
    Refs         []*RefSnapshot
    Computed     []*ComputedSnapshot
    Watchers     []*WatcherSnapshot
    Lifecycle    *LifecycleStatus
    Performance  *ComponentPerformance
    Timestamp    time.Time
}

// RefSnapshot captures ref state
type RefSnapshot struct {
    ID          string
    Name        string
    Type        string
    Value       interface{}
    Watchers    int
    History     []ValueChange
}

// EventRecord captures an event
type EventRecord struct {
    ID          string
    Name        string
    SourceID    string
    TargetID    string
    Payload     interface{}
    Timestamp   time.Time
    Duration    time.Duration
    BubblePath  []string
    Handlers    []HandlerExecution
}

// PerformanceData tracks performance metrics
type PerformanceData struct {
    Components   map[string]*ComponentPerformance
    Updates      []UpdateMetrics
    Renders      []RenderMetrics
    Memory       *MemoryMetrics
    FPS          float64
}
```

---

## Component Inspector Architecture

### Tree View Implementation

```go
type ComponentTreeView struct {
    root       *ComponentSnapshot
    selected   *ComponentSnapshot
    expanded   map[string]bool
    viewport   *Viewport
    renderer   *TreeRenderer
}

func (tv *ComponentTreeView) Render() string {
    lines := []string{}
    
    tv.renderNode(tv.root, 0, &lines)
    
    return strings.Join(lines, "\n")
}

func (tv *ComponentTreeView) renderNode(node *ComponentSnapshot, depth int, lines *[]string) {
    indent := strings.Repeat("  ", depth)
    
    // Component icon
    icon := "⊞"
    if tv.isExpanded(node.ID) {
        icon = "⊟"
    }
    
    // Selection indicator
    prefix := " "
    if tv.selected != nil && tv.selected.ID == node.ID {
        prefix = "▶"
    }
    
    // Component name with state count
    line := fmt.Sprintf("%s%s %s (%d refs)",
        indent, icon, node.Name, len(node.Refs))
    
    *lines = append(*lines, prefix+line)
    
    // Render children if expanded
    if tv.isExpanded(node.ID) {
        for _, child := range node.Children {
            tv.renderNode(child, depth+1, lines)
        }
    }
}
```

### Detail Panel

```go
type DetailPanel struct {
    component *ComponentSnapshot
    tabs      []string
    activeTab int
}

func (dp *DetailPanel) Render() string {
    if dp.component == nil {
        return "No component selected"
    }
    
    sections := []string{
        dp.renderHeader(),
        dp.renderTabs(),
        dp.renderActiveTab(),
    }
    
    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (dp *DetailPanel) renderActiveTab() string {
    switch dp.tabs[dp.activeTab] {
    case "State":
        return dp.renderState()
    case "Props":
        return dp.renderProps()
    case "Events":
        return dp.renderEvents()
    case "Performance":
        return dp.renderPerformance()
    default:
        return ""
    }
}

func (dp *DetailPanel) renderState() string {
    lines := []string{"State:"}
    
    for _, ref := range dp.component.Refs {
        lines = append(lines, fmt.Sprintf("  %s: %v (%s)",
            ref.Name, ref.Value, ref.Type))
    }
    
    for _, computed := range dp.component.Computed {
        lines = append(lines, fmt.Sprintf("  %s: %v (computed)",
            computed.Name, computed.Value))
    }
    
    return strings.Join(lines, "\n")
}
```

---

## State Viewer Architecture

### State Tree Visualization

```go
type StateViewer struct {
    store       *DevToolsStore
    filter      string
    showHistory bool
    selected    *RefSnapshot
}

func (sv *StateViewer) Render() string {
    components := sv.store.GetAllComponents()
    
    lines := []string{"Reactive State:"}
    
    for _, comp := range components {
        lines = append(lines, fmt.Sprintf("┌─ %s", comp.Name))
        
        for _, ref := range comp.Refs {
            if sv.matchesFilter(ref.Name) {
                indicator := " "
                if sv.selected != nil && sv.selected.ID == ref.ID {
                    indicator = "▶"
                }
                
                line := fmt.Sprintf("│ %s %s: %v",
                    indicator, ref.Name, sv.formatValue(ref.Value))
                
                if len(ref.History) > 0 {
                    line += fmt.Sprintf(" (changed %d times)", len(ref.History))
                }
                
                lines = append(lines, line)
            }
        }
        
        lines = append(lines, "└─")
    }
    
    return strings.Join(lines, "\n")
}
```

### State History

```go
type StateHistory struct {
    changes []StateChange
    maxSize int
    mu      sync.RWMutex
}

type StateChange struct {
    RefID     string
    RefName   string
    OldValue  interface{}
    NewValue  interface{}
    Timestamp time.Time
    Source    string // Which component/function changed it
}

func (sh *StateHistory) Record(change StateChange) {
    sh.mu.Lock()
    defer sh.mu.Unlock()
    
    sh.changes = append(sh.changes, change)
    
    // Keep only last N changes
    if len(sh.changes) > sh.maxSize {
        sh.changes = sh.changes[len(sh.changes)-sh.maxSize:]
    }
}

func (sh *StateHistory) GetHistory(refID string) []StateChange {
    sh.mu.RLock()
    defer sh.mu.RUnlock()
    
    history := []StateChange{}
    for _, change := range sh.changes {
        if change.RefID == refID {
            history = append(history, change)
        }
    }
    
    return history
}
```

---

## Event Tracker Architecture

### Event Capture

```go
type EventTracker struct {
    events     *EventLog
    filter     EventFilter
    paused     bool
    maxEvents  int
}

type EventLog struct {
    records []EventRecord
    mu      sync.RWMutex
}

func (et *EventTracker) CaptureEvent(event EventRecord) {
    if et.paused {
        return
    }
    
    if !et.filter.Matches(event) {
        return
    }
    
    et.events.Append(event)
    
    // Trim old events
    if et.events.Len() > et.maxEvents {
        et.events.TrimOldest()
    }
}

func (et *EventTracker) Render() string {
    events := et.events.GetRecent(50)
    
    lines := []string{"Recent Events:"}
    
    for i := len(events) - 1; i >= 0; i-- {
        event := events[i]
        
        duration := ""
        if event.Duration > 0 {
            duration = fmt.Sprintf(" (%s)", event.Duration)
        }
        
        line := fmt.Sprintf("[%s] %s → %s%s",
            event.Timestamp.Format("15:04:05.000"),
            event.Name,
            event.TargetID,
            duration)
        
        lines = append(lines, line)
    }
    
    return strings.Join(lines, "\n")
}
```

---

## Performance Monitor Architecture

### Metrics Collection

```go
type PerformanceMonitor struct {
    data      *PerformanceData
    collector *MetricsCollector
    renderer  *FlameGraphRenderer
}

type ComponentPerformance struct {
    ComponentID   string
    ComponentName string
    RenderCount   int64
    TotalRenderTime time.Duration
    AvgRenderTime   time.Duration
    MaxRenderTime   time.Duration
    MinRenderTime   time.Duration
    MemoryUsage   uint64
    LastUpdate    time.Time
}

func (pm *PerformanceMonitor) RecordRender(componentID string, duration time.Duration) {
    perf := pm.data.Components[componentID]
    if perf == nil {
        perf = &ComponentPerformance{
            ComponentID: componentID,
            MinRenderTime: duration,
        }
        pm.data.Components[componentID] = perf
    }
    
    perf.RenderCount++
    perf.TotalRenderTime += duration
    perf.AvgRenderTime = time.Duration(int64(perf.TotalRenderTime) / perf.RenderCount)
    
    if duration > perf.MaxRenderTime {
        perf.MaxRenderTime = duration
    }
    if duration < perf.MinRenderTime {
        perf.MinRenderTime = duration
    }
    
    perf.LastUpdate = time.Now()
}

func (pm *PerformanceMonitor) Render() string {
    components := pm.getSortedComponents()
    
    lines := []string{
        "Component Performance:",
        "───────────────────────────────────────────────",
        "Component          Renders  Avg Time  Max Time",
        "───────────────────────────────────────────────",
    }
    
    for _, comp := range components {
        line := fmt.Sprintf("%-18s %7d  %8s  %8s",
            truncate(comp.ComponentName, 18),
            comp.RenderCount,
            comp.AvgRenderTime,
            comp.MaxRenderTime)
        
        lines = append(lines, line)
    }
    
    return strings.Join(lines, "\n")
}
```

### Flame Graph Generation

```go
type FlameGraphRenderer struct {
    width  int
    height int
}

func (fgr *FlameGraphRenderer) Render(perfData *PerformanceData) string {
    // Build flame graph from performance data
    root := fgr.buildFlameTree(perfData)
    
    // Render as ASCII flame graph
    lines := []string{}
    fgr.renderFlameNode(root, 0, fgr.width, &lines)
    
    return strings.Join(lines, "\n")
}

func (fgr *FlameGraphRenderer) renderFlameNode(node *FlameNode, depth, width int, lines *[]string) {
    if depth >= fgr.height {
        return
    }
    
    // Calculate bar width based on time percentage
    barWidth := int(float64(width) * node.TimePercentage)
    
    if barWidth == 0 {
        return
    }
    
    // Render bar
    bar := strings.Repeat("█", barWidth)
    label := truncate(node.Name, barWidth)
    
    line := fmt.Sprintf("%s%s", strings.Repeat(" ", depth*2), bar)
    
    if len(*lines) <= depth {
        *lines = append(*lines, line)
    } else {
        (*lines)[depth] += line
    }
    
    // Render children
    childOffset := 0
    for _, child := range node.Children {
        childWidth := int(float64(width) * child.TimePercentage)
        fgr.renderFlameNode(child, depth+1, childWidth, lines)
        childOffset += childWidth
    }
}
```

---

## Command Timeline Architecture

```go
type CommandTimeline struct {
    commands  []CommandRecord
    paused    bool
    maxSize   int
    mu        sync.RWMutex
}

type CommandRecord struct {
    ID          string
    Type        string
    Source      string
    SourceID    string
    Generated   time.Time
    Executed    time.Time
    Duration    time.Duration
    Batch       int
    BatchSize   int
    Message     tea.Msg
    Error       error
}

func (ct *CommandTimeline) RecordCommand(record CommandRecord) {
    ct.mu.Lock()
    defer ct.mu.Unlock()
    
    if ct.paused {
        return
    }
    
    ct.commands = append(ct.commands, record)
    
    if len(ct.commands) > ct.maxSize {
        ct.commands = ct.commands[len(ct.commands)-ct.maxSize:]
    }
}

func (ct *CommandTimeline) Render(width int) string {
    ct.mu.RLock()
    defer ct.mu.RUnlock()
    
    // Timeline visualization
    startTime := ct.commands[0].Generated
    endTime := ct.commands[len(ct.commands)-1].Executed
    totalDuration := endTime.Sub(startTime)
    
    lines := []string{
        "Command Timeline:",
        fmt.Sprintf("Time span: %s", totalDuration),
        "",
    }
    
    // Render timeline bars
    for _, cmd := range ct.commands {
        offset := int(float64(width) * cmd.Generated.Sub(startTime).Seconds() / totalDuration.Seconds())
        duration := int(float64(width) * cmd.Duration.Seconds() / totalDuration.Seconds())
        
        if duration < 1 {
            duration = 1
        }
        
        bar := strings.Repeat(" ", offset) + strings.Repeat("▬", duration)
        label := fmt.Sprintf(" %s", truncate(cmd.Type, 20))
        
        lines = append(lines, bar+label)
    }
    
    return strings.Join(lines, "\n")
}
```

---

## Layout Manager Architecture

### Split Pane Layout

```go
type LayoutManager struct {
    mode       LayoutMode
    ratio      float64  // Split ratio (0.0 - 1.0)
    appView    string
    toolsView  string
    width      int
    height     int
}

type LayoutMode int

const (
    LayoutHorizontal LayoutMode = iota  // Side by side
    LayoutVertical                      // Stacked
    LayoutOverlay                       // Tools on top
    LayoutHidden                        // Tools hidden
)

func (lm *LayoutManager) Render(appContent, toolsContent string) string {
    switch lm.mode {
    case LayoutHorizontal:
        return lm.renderHorizontal(appContent, toolsContent)
    case LayoutVertical:
        return lm.renderVertical(appContent, toolsContent)
    case LayoutOverlay:
        return lm.renderOverlay(appContent, toolsContent)
    case LayoutHidden:
        return appContent
    default:
        return appContent
    }
}

func (lm *LayoutManager) renderHorizontal(appContent, toolsContent string) string {
    appWidth := int(float64(lm.width) * lm.ratio)
    toolsWidth := lm.width - appWidth - 1  // -1 for separator
    
    appBox := lipgloss.NewStyle().
        Width(appWidth).
        Height(lm.height).
        Border(lipgloss.NormalBorder(), false, true, false, false).
        Render(appContent)
    
    toolsBox := lipgloss.NewStyle().
        Width(toolsWidth).
        Height(lm.height).
        Render(toolsContent)
    
    return lipgloss.JoinHorizontal(lipgloss.Top, appBox, toolsBox)
}

func (lm *LayoutManager) renderVertical(appContent, toolsContent string) string {
    appHeight := int(float64(lm.height) * lm.ratio)
    toolsHeight := lm.height - appHeight - 1  // -1 for separator
    
    appBox := lipgloss.NewStyle().
        Width(lm.width).
        Height(appHeight).
        Border(lipgloss.NormalBorder(), false, false, true, false).
        Render(appContent)
    
    toolsBox := lipgloss.NewStyle().
        Width(lm.width).
        Height(toolsHeight).
        Render(toolsContent)
    
    return lipgloss.JoinVertical(lipgloss.Left, appBox, toolsBox)
}
```

---

## Instrumentation Hooks

### Component Lifecycle Hooks

```go
type ComponentHook interface {
    OnComponentCreated(snapshot *ComponentSnapshot)
    OnComponentMounted(id string)
    OnComponentUpdated(id string)
    OnComponentUnmounted(id string)
}

// Install hooks into component runtime
func (dt *DevTools) InstallComponentHooks(runtime *ComponentRuntime) {
    originalMount := runtime.onMounted
    runtime.onMounted = func() {
        originalMount()
        
        snapshot := dt.collector.CaptureComponent(runtime)
        dt.store.AddComponent(snapshot)
        
        for _, hook := range dt.collector.componentHooks {
            hook.OnComponentMounted(runtime.id)
        }
    }
    
    // Similar for other lifecycle events...
}
```

### State Change Hooks

```go
type StateHook interface {
    OnRefChanged(refID string, oldValue, newValue interface{})
    OnComputedEvaluated(computedID string, value interface{}, duration time.Duration)
    OnWatcherTriggered(watcherID string, value interface{})
}

// Install hooks into Ref
func (dt *DevTools) InstallRefHooks(ref *Ref[T]) {
    originalSet := ref.Set
    ref.Set = func(value T) {
        oldValue := ref.Get()
        originalSet(value)
        
        change := StateChange{
            RefID:     ref.id,
            RefName:   ref.name,
            OldValue:  oldValue,
            NewValue:  value,
            Timestamp: time.Now(),
        }
        
        dt.store.stateHistory.Record(change)
        
        for _, hook := range dt.collector.stateHooks {
            hook.OnRefChanged(ref.id, oldValue, value)
        }
    }
}
```

---

## Data Export/Import

### Export Format

```go
type ExportData struct {
    Version      string                 `json:"version"`
    Timestamp    time.Time              `json:"timestamp"`
    App          AppMetadata            `json:"app"`
    Components   []*ComponentSnapshot   `json:"components"`
    State        []StateChange          `json:"state"`
    Events       []EventRecord          `json:"events"`
    Performance  *PerformanceData       `json:"performance"`
    Commands     []CommandRecord        `json:"commands"`
}

func (dt *DevTools) Export(filename string, options ExportOptions) error {
    data := ExportData{
        Version:   "1.0",
        Timestamp: time.Now(),
        App: AppMetadata{
            Name:    dt.config.AppName,
            Version: dt.config.AppVersion,
        },
    }
    
    if options.IncludeComponents {
        data.Components = dt.store.GetAllComponents()
    }
    
    if options.IncludeState {
        data.State = dt.store.stateHistory.GetAll()
    }
    
    if options.Sanitize {
        data = sanitizeData(data, options.RedactPatterns)
    }
    
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, jsonData, 0644)
}
```

---

## Known Limitations & Solutions

### Limitation 1: Performance Overhead
**Problem**: Instrumentation slows down application  
**Current Design**: ~5% overhead target  
**Solution**: Sampling, lazy collection, toggleable hooks  
**Benefits**: Minimal impact, usable in production  
**Priority**: HIGH - must maintain low overhead

### Limitation 2: Large Data Sets
**Problem**: 10,000 components slow to display  
**Current Design**: Virtual scrolling, pagination  
**Solution**: Tree virtualization, lazy loading, aggregation  
**Benefits**: Handles large apps  
**Priority**: MEDIUM - important for scale

### Limitation 3: Terminal Size Limits
**Problem**: Dev tools need space, conflicts with app  
**Current Design**: Split pane reduces app space  
**Solution**: Collapsible panels, overlay mode, separate window  
**Benefits**: Flexible layouts  
**Priority**: MEDIUM - usability improvement

### Limitation 4: State Mutation Safety
**Problem**: Editing state could crash app  
**Current Design**: Read-only by default  
**Solution**: Validation, rollback, confirmation prompts  
**Benefits**: Safe experimentation  
**Priority**: HIGH - prevents crashes

---

## Future Enhancements

### Phase 4+
1. **Remote Debugging**: Connect to dev tools over network
2. **Time Travel**: Step backward/forward through state changes
3. **Automated Testing**: Record interactions, replay as tests
4. **AI Suggestions**: ML-powered debugging hints
5. **Custom Panels**: Plugin system for extensions
6. **Profiling Reports**: Generate PDF/HTML reports
7. **Multi-App Debugging**: Debug multiple TUI apps simultaneously

---

## Summary

The Dev Tools system provides comprehensive debugging and inspection capabilities for BubblyUI applications through instrumentation hooks, in-memory data collection, and a split-pane TUI interface. It captures component tree structure, reactive state changes, event flow, route navigation, performance metrics, and command timelines with < 5% overhead. The system includes component inspector, state viewer, event tracker, router debugger, performance monitor, and command timeline with keyboard-driven navigation, export capabilities, and minimal performance impact when enabled.
