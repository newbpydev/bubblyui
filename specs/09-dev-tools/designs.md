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

## Advanced Sanitization Architecture

### Pattern Priority System

**Type Definitions:**
```go
type SanitizePattern struct {
    Pattern     *regexp.Regexp
    Replacement string
    Priority    int    // Higher values apply first (0 = default)
    Name        string // For metrics and audit trails
}

func (s *Sanitizer) AddPatternWithPriority(pattern, replacement string, priority int, name string) error {
    re, err := regexp.Compile(pattern)
    if err != nil {
        return fmt.Errorf("invalid pattern: %w", err)
    }
    
    s.patterns = append(s.patterns, SanitizePattern{
        Pattern:     re,
        Replacement: replacement,
        Priority:    priority,
        Name:        name,
    })
    
    return nil
}

func (s *Sanitizer) sortPatterns() {
    // Sort by priority (descending), then by insertion order (stable)
    sort.SliceStable(s.patterns, func(i, j int) bool {
        return s.patterns[i].Priority > s.patterns[j].Priority
    })
}
```

**Priority Algorithm:**
1. Sort patterns by Priority field (higher = first)
2. For equal priorities, maintain insertion order (stable sort)
3. Apply patterns sequentially
4. Already-redacted text won't match subsequent patterns

**Priority Ranges:**
- **100+**: Critical compliance patterns (always first)
- **50-99**: Organization-specific patterns
- **10-49**: Custom business rules
- **0-9**: Default/low-priority patterns
- **Negative**: Cleanup/fallback patterns (last)

### Pattern Templates

**Template Registry:**
```go
type TemplateRegistry map[string][]SanitizePattern

var DefaultTemplates = TemplateRegistry{
    "pii": {
        // Personal Identifiable Information
        {Pattern: regexp.MustCompile(`(?i)(ssn|social[-_]?security)(["'\s:=]+)(\d{3}-?\d{2}-?\d{4})`), Replacement: "${1}${2}[REDACTED_SSN]", Priority: 100, Name: "ssn"},
        {Pattern: regexp.MustCompile(`(?i)(email)(["'\s:=]+)([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`), Replacement: "${1}${2}[REDACTED_EMAIL]", Priority: 90, Name: "email"},
        {Pattern: regexp.MustCompile(`(?i)(phone|tel)(["'\s:=]+)(\+?1?\s?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4})`), Replacement: "${1}${2}[REDACTED_PHONE]", Priority: 90, Name: "phone"},
    },
    "pci": {
        // Payment Card Industry
        {Pattern: regexp.MustCompile(`(?i)(card|cc|credit[-_]?card)(["'\s:=]+)(\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4})`), Replacement: "${1}${2}[REDACTED_CARD]", Priority: 100, Name: "credit_card"},
        {Pattern: regexp.MustCompile(`(?i)(cvv|cvc|security[-_]?code)(["'\s:=]+)(\d{3,4})`), Replacement: "${1}${2}[REDACTED_CVV]", Priority: 100, Name: "cvv"},
        {Pattern: regexp.MustCompile(`(?i)(exp|expiry|expiration)(["'\s:=]+)(\d{2}/\d{2,4})`), Replacement: "${1}${2}[REDACTED_EXP]", Priority: 90, Name: "expiry"},
    },
    "hipaa": {
        // Healthcare
        {Pattern: regexp.MustCompile(`(?i)(mrn|medical[-_]?record)(["'\s:=]+)([A-Z0-9-]+)`), Replacement: "${1}${2}[REDACTED_MRN]", Priority: 100, Name: "medical_record"},
        {Pattern: regexp.MustCompile(`(?i)(diagnosis|condition)(["'\s:=]+)([^"',}\]]+)`), Replacement: "${1}${2}[REDACTED_DIAGNOSIS]", Priority: 90, Name: "diagnosis"},
    },
    "gdpr": {
        // GDPR compliance
        {Pattern: regexp.MustCompile(`(?i)(ip[-_]?address|ip)(["'\s:=]+)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`), Replacement: "${1}${2}[REDACTED_IP]", Priority: 90, Name: "ip_address"},
        {Pattern: regexp.MustCompile(`(?i)(mac[-_]?address|mac)(["'\s:=]+)([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`), Replacement: "${1}${2}[REDACTED_MAC]", Priority: 90, Name: "mac_address"},
    },
}

func (s *Sanitizer) LoadTemplate(name string) error {
    patterns, ok := DefaultTemplates[name]
    if !ok {
        return fmt.Errorf("template not found: %s", name)
    }
    
    for _, p := range patterns {
        s.patterns = append(s.patterns, p)
    }
    
    s.sortPatterns()
    return nil
}

func (s *Sanitizer) LoadTemplates(names ...string) error {
    for _, name := range names {
        if err := s.LoadTemplate(name); err != nil {
            return err
        }
    }
    return nil
}
```

### Sanitization Metrics

**Stats Tracking:**
```go
type SanitizationStats struct {
    RedactedCount    int                // Total values redacted
    PatternMatches   map[string]int     // Count per pattern name
    Duration         time.Duration      // Time taken
    BytesProcessed   int64              // Data size processed
    StartTime        time.Time
    EndTime          time.Time
}

type Sanitizer struct {
    patterns   []SanitizePattern
    lastStats  *SanitizationStats
    mu         sync.RWMutex
}

func (s *Sanitizer) Sanitize(data *ExportData) *ExportData {
    stats := &SanitizationStats{
        PatternMatches: make(map[string]int),
        StartTime:      time.Now(),
    }
    
    result := s.sanitizeWithStats(data, stats)
    
    stats.EndTime = time.Now()
    stats.Duration = stats.EndTime.Sub(stats.StartTime)
    
    s.mu.Lock()
    s.lastStats = stats
    s.mu.Unlock()
    
    return result
}

func (s *Sanitizer) GetLastStats() *SanitizationStats {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.lastStats
}

func (s *Sanitizer) sanitizeStringWithStats(str string, stats *SanitizationStats) string {
    result := str
    for _, pattern := range s.patterns {
        if pattern.Pattern.MatchString(result) {
            result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
            stats.RedactedCount++
            if pattern.Name != "" {
                stats.PatternMatches[pattern.Name]++
            }
        }
    }
    return result
}
```

### Dry-Run Preview Mode

**Preview Types:**
```go
type DryRunResult struct {
    Matches          []MatchLocation
    WouldRedactCount int
    PreviewData      interface{} // Annotated data showing what would change
}

type MatchLocation struct {
    Path     string // e.g., "components[0].props.password"
    Pattern  string // Pattern name that matched
    Original string // Original value (truncated for display)
    Redacted string // What it would become
    Line     int    // Line number in JSON (if applicable)
    Column   int    // Column in JSON (if applicable)
}

type SanitizeOptions struct {
    DryRun          bool
    MaxPreviewLen   int  // Max length of original value in preview
}

func (s *Sanitizer) SanitizeWithOptions(data *ExportData, opts SanitizeOptions) (*ExportData, *DryRunResult) {
    if opts.DryRun {
        result := &DryRunResult{
            Matches: make([]MatchLocation, 0),
        }
        
        // Traverse data and collect matches without mutating
        s.previewSanitization(data, "", result, opts.MaxPreviewLen)
        result.WouldRedactCount = len(result.Matches)
        
        return data, result // Return original data unchanged
    }
    
    return s.Sanitize(data), nil
}

func (s *Sanitizer) previewSanitization(data interface{}, path string, result *DryRunResult, maxLen int) {
    // Use reflection to traverse and find matches
    val := reflect.ValueOf(data)
    switch val.Kind() {
    case reflect.String:
        s.previewString(val.String(), path, result, maxLen)
    case reflect.Map:
        for _, key := range val.MapKeys() {
            s.previewSanitization(val.MapIndex(key).Interface(), path+"."+key.String(), result, maxLen)
        }
    case reflect.Slice:
        for i := 0; i < val.Len(); i++ {
            s.previewSanitization(val.Index(i).Interface(), fmt.Sprintf("%s[%d]", path, i), result, maxLen)
        }
    // ... handle other types
    }
}

func (s *Sanitizer) previewString(str, path string, result *DryRunResult, maxLen int) {
    for _, pattern := range s.patterns {
        if pattern.Pattern.MatchString(str) {
            redacted := pattern.Pattern.ReplaceAllString(str, pattern.Replacement)
            original := str
            if len(original) > maxLen {
                original = original[:maxLen] + "..."
            }
            
            result.Matches = append(result.Matches, MatchLocation{
                Path:     path,
                Pattern:  pattern.Name,
                Original: original,
                Redacted: redacted,
            })
        }
    }
}
```

---

## Streaming Sanitization Architecture

### Stream Processing Design

**Streaming API:**
```go
type StreamSanitizer struct {
    *Sanitizer
    bufferSize int // Default: 64KB
}

func (s *StreamSanitizer) SanitizeStream(reader io.Reader, writer io.Writer, progress func(bytesProcessed int64)) error {
    decoder := json.NewDecoder(reader)
    encoder := json.NewEncoder(writer)
    
    // Start JSON array
    writer.Write([]byte("{\n"))
    
    var bytesProcessed int64
    first := true
    
    // Stream components one by one
    for decoder.More() {
        var component ComponentSnapshot
        if err := decoder.Decode(&component); err != nil {
            return fmt.Errorf("decode error: %w", err)
        }
        
        // Sanitize in-place
        sanitized := s.sanitizeComponent(&component)
        
        // Write to output stream
        if !first {
            writer.Write([]byte(",\n"))
        }
        first = false
        
        if err := encoder.Encode(sanitized); err != nil {
            return fmt.Errorf("encode error: %w", err)
        }
        
        bytesProcessed += int64(unsafe.Sizeof(component))
        if progress != nil {
            progress(bytesProcessed)
        }
    }
    
    writer.Write([]byte("}\n"))
    return nil
}
```

**Chunked Processing:**
```go
func (dt *DevTools) ExportStream(filename string, opts ExportOptions) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    writer := bufio.NewWriterSize(file, 64*1024) // 64KB buffer
    defer writer.Flush()
    
    // Create in-memory reader for components
    var buf bytes.Buffer
    json.NewEncoder(&buf).Encode(dt.store.GetAllComponents())
    
    sanitizer := &StreamSanitizer{
        Sanitizer:  NewSanitizer(),
        bufferSize: 64 * 1024,
    }
    
    if opts.Templates != nil {
        sanitizer.LoadTemplates(opts.Templates...)
    }
    
    return sanitizer.SanitizeStream(&buf, writer, opts.ProgressCallback)
}
```

**Memory Bounds:**
- Buffer size: 64KB (configurable)
- Max single object: 10MB (configurable)
- Total memory: O(buffer size), not O(file size)
- Uses `json.Decoder` for streaming read
- Uses `bufio.Writer` for buffered write

### Performance Optimization

**Reflection Caching:**
```go
type typeCache struct {
    types sync.Map // map[reflect.Type]*cachedTypeInfo
}

type cachedTypeInfo struct {
    kind       reflect.Kind
    fields     []reflect.StructField // For structs
    elemType   reflect.Type          // For slices/arrays
    keyType    reflect.Type          // For maps
    valueType  reflect.Type          // For maps
}

var globalTypeCache = &typeCache{}

func (s *Sanitizer) SanitizeValueOptimized(val interface{}) interface{} {
    t := reflect.TypeOf(val)
    
    // Check cache first
    if cached, ok := globalTypeCache.types.Load(t); ok {
        info := cached.(*cachedTypeInfo)
        return s.sanitizeWithCachedInfo(val, info)
    }
    
    // Cache miss - compute and store
    info := &cachedTypeInfo{
        kind: t.Kind(),
    }
    
    switch t.Kind() {
    case reflect.Struct:
        info.fields = make([]reflect.StructField, t.NumField())
        for i := 0; i < t.NumField(); i++ {
            info.fields[i] = t.Field(i)
        }
    case reflect.Slice, reflect.Array:
        info.elemType = t.Elem()
    case reflect.Map:
        info.keyType = t.Key()
        info.valueType = t.Elem()
    }
    
    globalTypeCache.types.Store(t, info)
    return s.sanitizeWithCachedInfo(val, info)
}

func (s *Sanitizer) sanitizeWithCachedInfo(val interface{}, info *cachedTypeInfo) interface{} {
    v := reflect.ValueOf(val)
    
    switch info.kind {
    case reflect.Struct:
        result := reflect.New(v.Type()).Elem()
        for _, field := range info.fields {
            if field.IsExported() {
                fieldVal := v.FieldByIndex(field.Index)
                sanitized := s.SanitizeValueOptimized(fieldVal.Interface())
                result.FieldByIndex(field.Index).Set(reflect.ValueOf(sanitized))
            }
        }
        return result.Interface()
    // ... other cases using cached info
    }
    
    return val
}
```

**Performance Benchmarks:**
```go
// Expected improvements:
// - Reflection caching: 30-50% faster for repeated types
// - Streaming: Constant memory vs O(n) for in-memory
// - Buffer tuning: 10-20% faster with optimal buffer size

BenchmarkSanitize/in-memory-10              1000  1234567 ns/op  2048000 B/op  5000 allocs/op
BenchmarkSanitize/streaming-10              1200  1034567 ns/op   128000 B/op  1200 allocs/op
BenchmarkSanitize/with-cache-10             1800   687654 ns/op  2048000 B/op  2500 allocs/op
BenchmarkSanitize/streaming+cache-10        2000   567890 ns/op   128000 B/op   600 allocs/op
```

---

## Export & Import Enhancements Architecture

### Compression Support

**Gzip Integration:**
```go
import (
    "compress/gzip"
    "io"
)

type ExportOptions struct {
    // ... existing fields
    Compress       bool
    CompressionLevel int  // gzip.DefaultCompression, gzip.BestSpeed, gzip.BestCompression
}

func (dt *DevTools) ExportCompressed(filename string, opts ExportOptions) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Create gzip writer
    var writer io.Writer = file
    if opts.Compress {
        level := gzip.DefaultCompression
        if opts.CompressionLevel != 0 {
            level = opts.CompressionLevel
        }
        
        gzipWriter, err := gzip.NewWriterLevel(file, level)
        if err != nil {
            return err
        }
        defer gzipWriter.Close()
        writer = gzipWriter
    }
    
    // Export to writer (supports both compressed and uncompressed)
    return dt.exportToWriter(writer, opts)
}

func (dt *DevTools) ImportCompressed(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Auto-detect gzip by reading magic bytes
    reader := io.Reader(file)
    magic := make([]byte, 2)
    if _, err := io.ReadFull(file, magic); err == nil {
        file.Seek(0, 0)  // Reset to start
        
        // Gzip magic bytes: 0x1f 0x8b
        if magic[0] == 0x1f && magic[1] == 0x8b {
            gzipReader, err := gzip.NewReader(file)
            if err != nil {
                return err
            }
            defer gzipReader.Close()
            reader = gzipReader
        }
    }
    
    return dt.importFromReader(reader)
}
```

**Compression Benefits:**
- File size reduction: 50-70% for typical JSON exports
- Bandwidth savings for network transfers
- Storage efficiency for long-term archival
- Negligible performance overhead (< 100ms for 10MB files)

### Multiple Format Support

**Format Interface:**
```go
type ExportFormat interface {
    Name() string
    Extension() string
    ContentType() string
    Marshal(data *ExportData) ([]byte, error)
    Unmarshal([]byte, *ExportData) error
}

type JSONFormat struct{}

func (f *JSONFormat) Name() string { return "json" }
func (f *JSONFormat) Extension() string { return ".json" }
func (f *JSONFormat) ContentType() string { return "application/json" }
func (f *JSONFormat) Marshal(data *ExportData) ([]byte, error) {
    return json.MarshalIndent(data, "", "  ")
}
func (f *JSONFormat) Unmarshal(b []byte, data *ExportData) error {
    return json.Unmarshal(b, data)
}

type YAMLFormat struct{}

func (f *YAMLFormat) Name() string { return "yaml" }
func (f *YAMLFormat) Extension() string { return ".yaml" }
func (f *YAMLFormat) ContentType() string { return "application/x-yaml" }
func (f *YAMLFormat) Marshal(data *ExportData) ([]byte, error) {
    return yaml.Marshal(data)
}
func (f *YAMLFormat) Unmarshal(b []byte, data *ExportData) error {
    return yaml.Unmarshal(b, data)
}

type MessagePackFormat struct{}

func (f *MessagePackFormat) Name() string { return "msgpack" }
func (f *MessagePackFormat) Extension() string { return ".msgpack" }
func (f *MessagePackFormat) ContentType() string { return "application/msgpack" }
func (f *MessagePackFormat) Marshal(data *ExportData) ([]byte, error) {
    return msgpack.Marshal(data)
}
func (f *MessagePackFormat) Unmarshal(b []byte, data *ExportData) error {
    return msgpack.Unmarshal(b, data)
}

// Registry
var formats = map[string]ExportFormat{
    "json":    &JSONFormat{},
    "yaml":    &YAMLFormat{},
    "msgpack": &MessagePackFormat{},
}

func (dt *DevTools) ExportFormat(filename, format string, opts ExportOptions) error {
    fmt, ok := formats[format]
    if !ok {
        return fmt.Errorf("unknown format: %s", format)
    }
    
    data := dt.gatherExportData(opts)
    bytes, err := fmt.Marshal(data)
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, bytes, 0644)
}
```

**Format Comparison:**
| Format | Size | Speed | Readability | Use Case |
|--------|------|-------|-------------|----------|
| JSON | 100% | Fast | High | Default, human-readable |
| YAML | 95% | Medium | Very High | Config integration |
| MessagePack | 60% | Very Fast | None | Production, minimal size |

### Incremental Export (Delta)

**Delta Tracking:**
```go
type ExportCheckpoint struct {
    Timestamp     time.Time
    LastEventID   int
    LastStateID   int
    LastCommandID int
}

type IncrementalExportData struct {
    Checkpoint    ExportCheckpoint
    NewEvents     []EventRecord
    NewState      []StateChange
    NewCommands   []CommandRecord
}

func (dt *DevTools) ExportIncremental(filename string, since *ExportCheckpoint) error {
    data := &IncrementalExportData{
        Checkpoint: ExportCheckpoint{
            Timestamp: time.Now(),
        },
    }
    
    // Export only new data since checkpoint
    if since != nil {
        data.NewEvents = dt.store.events.GetSince(since.LastEventID)
        data.NewState = dt.store.stateHistory.GetSince(since.LastStateID)
        data.NewCommands = dt.store.timeline.GetSince(since.LastCommandID)
        
        // Update checkpoint IDs
        if len(data.NewEvents) > 0 {
            data.Checkpoint.LastEventID = data.NewEvents[len(data.NewEvents)-1].ID
        }
        if len(data.NewState) > 0 {
            data.Checkpoint.LastStateID = data.NewState[len(data.NewState)-1].ID
        }
        if len(data.NewCommands) > 0 {
            data.Checkpoint.LastCommandID = data.NewCommands[len(data.NewCommands)-1].ID
        }
    } else {
        // First export - full snapshot
        return dt.Export(filename, ExportOptions{
            IncludeState:  true,
            IncludeEvents: true,
            IncludeTimeline: true,
        })
    }
    
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, jsonData, 0644)
}
```

**Benefits:**
- Reduced file sizes for long-running sessions
- Faster exports (only delta processed)
- Time-series analysis capability
- Replay specific time ranges

### Version Migration

**Migration System:**
```go
type VersionMigration interface {
    From() string
    To() string
    Migrate(data map[string]interface{}) (map[string]interface{}, error)
}

type Migration_1_0_to_2_0 struct{}

func (m *Migration_1_0_to_2_0) From() string { return "1.0" }
func (m *Migration_1_0_to_2_0) To() string { return "2.0" }
func (m *Migration_1_0_to_2_0) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
    // Example: Add new fields, rename old ones, transform data
    
    // Add new metadata field
    if _, ok := data["metadata"]; !ok {
        data["metadata"] = map[string]interface{}{
            "upgraded_from": "1.0",
            "upgrade_time":  time.Now().Format(time.RFC3339),
        }
    }
    
    // Rename field if exists
    if components, ok := data["components"]; ok {
        // Transform component structure
        // ...
    }
    
    // Update version
    data["version"] = "2.0"
    
    return data, nil
}

var migrations = []VersionMigration{
    &Migration_1_0_to_2_0{},
}

func (dt *DevTools) Import(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }
    
    // Parse as generic map first
    var raw map[string]interface{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }
    
    // Check version and migrate if needed
    version, ok := raw["version"].(string)
    if !ok {
        return fmt.Errorf("missing version field")
    }
    
    if version != CurrentVersion {
        raw, err = dt.migrateVersion(raw, version, CurrentVersion)
        if err != nil {
            return fmt.Errorf("migration failed: %w", err)
        }
    }
    
    // Now unmarshal into typed structure
    migratedData, err := json.Marshal(raw)
    if err != nil {
        return err
    }
    
    var exportData ExportData
    if err := json.Unmarshal(migratedData, &exportData); err != nil {
        return err
    }
    
    return dt.restoreFromExportData(&exportData)
}

func (dt *DevTools) migrateVersion(data map[string]interface{}, from, to string) (map[string]interface{}, error) {
    current := from
    
    for current != to {
        migrated := false
        for _, mig := range migrations {
            if mig.From() == current {
                var err error
                data, err = mig.Migrate(data)
                if err != nil {
                    return nil, fmt.Errorf("migration %s->%s failed: %w", 
                        mig.From(), mig.To(), err)
                }
                current = mig.To()
                migrated = true
                break
            }
        }
        
        if !migrated {
            return nil, fmt.Errorf("no migration path from %s to %s", current, to)
        }
    }
    
    return data, nil
}
```

---

## UI & Integration Polish Architecture

### Responsive Terminal Sizing

**Terminal Size Detection:**
```go
import "github.com/charmbracelet/bubbletea"

type windowSizeMsg struct {
    width  int
    height int
}

func (dt *DevToolsUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Adjust layout based on new size
        dt.width = msg.Width
        dt.height = msg.Height
        
        // Recalculate split-pane widths
        dt.calculatePaneSizes()
        
        // Reflow content
        dt.reflow()
        
        return dt, nil
    }
    
    return dt, nil
}

func (dt *DevToolsUI) calculatePaneSizes() {
    // Adaptive widths based on terminal size
    if dt.width < 80 {
        // Narrow terminal: stack vertically
        dt.layout = "vertical"
        dt.treeWidth = dt.width
        dt.detailWidth = dt.width
    } else if dt.width < 120 {
        // Medium terminal: 50/50 split
        dt.layout = "horizontal"
        dt.treeWidth = dt.width / 2
        dt.detailWidth = dt.width / 2
    } else {
        // Wide terminal: 40/60 split (tree/detail)
        dt.layout = "horizontal"
        dt.treeWidth = dt.width * 40 / 100
        dt.detailWidth = dt.width * 60 / 100
    }
}
```

### Component Hierarchy Visualization

**Hierarchy Data Structure:**
```go
type ComponentNode struct {
    Component *ComponentSnapshot
    Children  []*ComponentNode
    Parent    *ComponentNode
    Depth     int
}

func (dt *DevTools) BuildComponentTree() *ComponentNode {
    // Build tree from flat component list using parent IDs
    nodes := make(map[string]*ComponentNode)
    var root *ComponentNode
    
    for _, comp := range dt.store.GetAllComponents() {
        node := &ComponentNode{
            Component: comp,
            Children:  make([]*ComponentNode, 0),
        }
        nodes[comp.ID] = node
        
        if comp.ParentID == "" {
            root = node
        }
    }
    
    // Link parent-child relationships
    for _, node := range nodes {
        if node.Component.ParentID != "" {
            if parent, ok := nodes[node.Component.ParentID]; ok {
                parent.Children = append(parent.Children, node)
                node.Parent = parent
                node.Depth = parent.Depth + 1
            }
        }
    }
    
    return root
}

func (n *ComponentNode) RenderTree(depth int) string {
    var sb strings.Builder
    
    // Indentation based on depth
    indent := strings.Repeat("  ", depth)
    
    // Tree branch characters
    branch := "├─"
    if depth == 0 {
        branch = ""
    }
    
    // Render this node
    sb.WriteString(fmt.Sprintf("%s%s %s (%dms)\n",
        indent,
        branch,
        n.Component.Name,
        n.Component.RenderTime,
    ))
    
    // Render children
    for i, child := range n.Children {
        if i == len(n.Children)-1 {
            // Last child uses └─
            sb.WriteString(child.RenderTree(depth + 1))
        } else {
            sb.WriteString(child.RenderTree(depth + 1))
        }
    }
    
    return sb.String()
}
```

### Framework Integration Hooks

**Automatic Instrumentation:**
```go
// In pkg/bubbly/component.go
func (c *componentImpl) Init() tea.Cmd {
    // Notify dev tools of component mount
    if devtools.IsEnabled() {
        devtools.NotifyComponentMounted(c.id, c.name)
    }
    
    // ... existing init logic
}

func (c *componentImpl) Update(msg tea.Msg) (Component, tea.Cmd) {
    // Notify dev tools of update
    if devtools.IsEnabled() {
        devtools.NotifyComponentUpdate(c.id, msg)
    }
    
    // ... existing update logic
}

func (c *componentImpl) Unmount() {
    // Notify dev tools of unmount
    if devtools.IsEnabled() {
        devtools.NotifyComponentUnmounted(c.id)
    }
    
    // ... existing unmount logic
}

// In pkg/bubbly/ref.go
func (r *Ref[T]) Set(value T) {
    // Notify dev tools of state change
    if devtools.IsEnabled() {
        devtools.NotifyRefChanged(r.id, r.value, value)
    }
    
    // ... existing set logic
}

// In component event emission
func (c *componentImpl) Emit(eventName string, data interface{}) {
    // Notify dev tools of event
    if devtools.IsEnabled() {
        devtools.NotifyEvent(c.id, eventName, data)
    }
    
    // ... existing emit logic
}
```

**Hook Registration:**
```go
type FrameworkHook interface {
    OnComponentMount(id, name string)
    OnComponentUpdate(id string, msg interface{})
    OnComponentUnmount(id string)
    OnRefChange(id string, oldValue, newValue interface{})
    OnEvent(componentID, eventName string, data interface{})
}

var globalHook FrameworkHook

func RegisterHook(hook FrameworkHook) {
    globalHook = hook
}

func IsEnabled() bool {
    return globalHook != nil
}

func NotifyComponentMounted(id, name string) {
    if globalHook != nil {
        globalHook.OnComponentMount(id, name)
    }
}

// ... other notify functions
```

---

## Reactive Cascade Architecture

### Overview
The reactive cascade system tracks the complete data flow from source Ref changes through Computed values to Watcher callbacks and WatchEffect executions. This provides complete visibility into Vue-inspired reactivity propagation.

### Data Flow Diagram

```
Ref.Set(newValue)
    ↓
notifyHookRefChange(refID, oldValue, newValue)  [✅ Already hooked]
    ↓
Ref.notifyWatchers() [Internal]
    ├→ Computed.Invalidate() [Dependent computed values]
    │   ↓
    │   Computed.GetTyped() [When accessed or has watchers]
    │   ↓
    │   notifyHookComputedChange(computedID, oldValue, newValue)  [⚠️ NEW HOOK]
    │   ↓
    │   Computed.notifyWatchers() [Internal]
    │       ├→ Watch callbacks
    │       │   ↓
    │       │   notifyHookWatchCallback(watcherID, newValue, oldValue)  [⚠️ NEW HOOK]
    │       │
    │       └→ WatchEffect.run()
    │           ↓
    │           notifyHookEffectRun(effectID)  [⚠️ NEW HOOK]
    │           ↓
    │           effect() execution
    │
    └→ Direct Watch callbacks
        ↓
        notifyHookWatchCallback(watcherID, newValue, oldValue)  [⚠️ NEW HOOK]
```

### Integration Points

#### 1. Computed Value Changes
**Location**: `pkg/bubbly/computed.go:178-183`  
**Trigger**: When `GetTyped()` re-evaluates and value changes  
**Hook**: `notifyHookComputedChange(id, oldValue, newValue)`

```go
// In Computed.GetTyped() after line 181
if hasWatchers && !reflect.DeepEqual(oldValue, result) {
    // NEW: Notify framework hooks of computed change
    computedID := fmt.Sprintf("computed-%p", c)
    notifyHookComputedChange(computedID, oldValue, result)
    
    // Existing: Notify watchers
    c.notifyWatchers(result, oldValue)
}
```

#### 2. Watch Callback Execution
**Location**: `pkg/bubbly/ref.go:234-237`, `pkg/bubbly/computed.go`  
**Trigger**: When watcher callback is about to execute  
**Hook**: `notifyHookWatchCallback(watcherID, newValue, oldValue)`

```go
// Create new helper function
func notifyWatcher[T any](w *watcher[T], newVal, oldVal T) {
    // NEW: Notify framework hooks before callback
    watcherID := fmt.Sprintf("watch-%p", w)
    notifyHookWatchCallback(watcherID, newVal, oldVal)
    
    // Existing: Execute callback
    w.callback(newVal, oldVal)
}
```

#### 3. WatchEffect Re-runs
**Location**: `pkg/bubbly/watch_effect.go:122`  
**Trigger**: When effect function re-executes due to dependency changes  
**Hook**: `notifyHookEffectRun(effectID)`

```go
// In watchEffect.run() before line 122
// NEW: Notify framework hooks before effect execution
effectID := fmt.Sprintf("effect-%p", e)
notifyHookEffectRun(effectID)

// Existing: Execute effect
e.effect()
```

#### 4. Component Tree Changes
**Location**: `pkg/bubbly/children.go:61, 153`  
**Trigger**: When child components are added or removed  
**Hooks**: `notifyHookChildAdded(parentID, childID)`, `notifyHookChildRemoved(parentID, childID)`

```go
// In AddChild after line 75
notifyHookChildAdded(c.id, child.ID())

// In RemoveChild after line 170
notifyHookChildRemoved(c.id, child.ID())
```

### Extended FrameworkHook Interface

```go
type FrameworkHook interface {
    // Existing methods (Task 8.6)
    OnComponentMount(id, name string)
    OnComponentUpdate(id string, msg interface{})
    OnComponentUnmount(id string)
    OnRefChange(id string, oldValue, newValue interface{})
    OnEvent(componentID, eventName string, data interface{})
    OnRenderComplete(componentID string, duration time.Duration)
    
    // NEW: Reactive cascade methods (Tasks 8.7-8.10)
    OnComputedChange(id string, oldValue, newValue interface{})
    OnWatchCallback(watcherID string, newValue, oldValue interface{})
    OnEffectRun(effectID string)
    OnChildAdded(parentID, childID string)
    OnChildRemoved(parentID, childID string)
}
```

### Type Definitions for Reactive Cascade

```go
// ComputedChangeRecord tracks computed value updates
type ComputedChangeRecord struct {
    ID         string
    OldValue   interface{}
    NewValue   interface{}
    Timestamp  time.Time
    TriggerRef string  // Which ref change caused this
}

// WatchCallbackRecord tracks watcher executions
type WatchCallbackRecord struct {
    ID         string
    WatcherID  string
    SourceType string  // "ref" or "computed"
    SourceID   string
    NewValue   interface{}
    OldValue   interface{}
    Timestamp  time.Time
    Duration   time.Duration
}

// EffectRunRecord tracks WatchEffect executions
type EffectRunRecord struct {
    ID           string
    EffectID     string
    TriggerCount int  // How many times triggered
    Dependencies []string  // Current dependencies
    Timestamp    time.Time
    Duration     time.Duration
}

// ChildMutationRecord tracks component tree changes
type ChildMutationRecord struct {
    ID        string
    ParentID  string
    ChildID   string
    Operation string  // "add" or "remove"
    Timestamp time.Time
}
```

### Performance Considerations

**Zero Overhead When Disabled:**
- All hooks use single nil check before execution
- No memory allocation when hook not registered
- Same pattern as existing Task 8.6 hooks

**Hook Registration:**
```go
// Fast path: nil check only
globalHookRegistry.mu.RLock()
hook := globalHookRegistry.hook
globalHookRegistry.mu.RUnlock()

if hook != nil {
    hook.OnComputedChange(id, oldValue, newValue)
}
```

**Memory Usage:**
- Computed IDs: `fmt.Sprintf("computed-%p", c)` - pointer address
- Watcher IDs: `fmt.Sprintf("watch-%p", w)` - pointer address  
- Effect IDs: `fmt.Sprintf("effect-%p", e)` - pointer address
- No ID tracking overhead when hooks disabled

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
