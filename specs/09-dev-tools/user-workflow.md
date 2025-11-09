# User Workflow: Dev Tools

## Developer Personas

### Persona 1: Bug Hunter (Lisa)
- **Background**: 4 years Go, debugging complex state issues
- **Goal**: Find why component not updating
- **Pain Point**: Can't see internal state without logging
- **Expects**: Real-time state inspection
- **Success**: Finds bug in 5 minutes instead of 2 hours

### Persona 2: Performance Engineer (Raj)
- **Background**: 6 years performance optimization
- **Goal**: Identify rendering bottlenecks
- **Pain Point**: No visibility into component performance
- **Expects**: Performance metrics and flame graphs
- **Success**: Optimizes app to 60 FPS

### Persona 3: New Developer (Emma)
- **Background**: 1 year coding, learning BubblyUI
- **Goal**: Understand how components work
- **Pain Point**: Framework behavior unclear
- **Expects**: Visual component tree and state
- **Success**: Learns framework in days, not weeks

---

## Primary User Journey: First-Time Dev Tools Usage

### Entry Point: Debugging State Issue

**Workflow: Enabling and Using Dev Tools**

#### Step 1: Enable Dev Tools
**User Action**: Add dev tools to application

```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

func main() {
    // Enable dev tools
    devtools.Enable()
    
    // Or per-app
    app := bubbly.NewApp().
        WithDevTools(true).
        Build()
    
    tea.NewProgram(app).Run()
}
```

**System Response**:
- Dev tools initialized
- Instrumentation hooks installed
- Data collection begins
- F12 shortcut registered

**UI Feedback**:
- Small indicator shows dev tools active
- Status bar: "DevTools: Press F12 to open"

#### Step 2: Open Dev Tools
**User Action**: Press F12

**System Response**:
- Screen splits (60/40 ratio)
- Dev tools panel appears on right
- Component tree populates
- Status shows "DevTools Active"

**Visual Layout**:
```
┌──────────────────────────┬─────────────────────────────┐
│   Your App (60%)         │   Dev Tools (40%)           │
│                          │                             │
│  Counter: 42             │  ┌─ Component Tree ───────┐│
│  [+] [-] [Reset]         │  │ ⊟ App                  ││
│                          │  │   ⊞ Header             ││
│                          │  │   ⊟ Counter            ││
│                          │  │     • count: Ref<int>  ││
│                          │  │     • last: Ref<time>  ││
│                          │  └────────────────────────┘│
│                          │                             │
│                          │  [Tabs: State│Events│Perf] │
└──────────────────────────┴─────────────────────────────┘
Press F12 to toggle │ Tab to switch focus │ ? for help
```

#### Step 3: Inspect Component
**User Action**: Navigate tree, select "Counter" component

**System Response**:
- Component highlighted in tree
- Detail panel shows component info
- State tab displays all refs

**Detail Panel Shows**:
```
┌─ Counter Component ──────────────────┐
│ ID: counter-1                        │
│ Type: bubbly.Component               │
│ Mounted: 2m 34s ago                  │
│                                      │
│ ┌─ State ──────────────────────────┐│
│ │ count: Ref<int>                  ││
│ │   value: 42                      ││
│ │   watchers: 1                    ││
│ │   changed: 42 times              ││
│ │                                  ││
│ │ lastUpdate: Ref<time.Time>       ││
│ │   value: 2024-10-29 17:15:23     ││
│ │   watchers: 0                    ││
│ └──────────────────────────────────┘│
│                                      │
│ ┌─ Props ───────────────────────────┐
│ │ initial: 0                       ││
│ │ step: 1                          ││
│ └──────────────────────────────────┘│
└──────────────────────────────────────┘
```

**Journey Milestone**: ✅ Can inspect component state!

---

### Feature Journey: Debugging State Issue

#### Step 4: Track State Changes
**User Action**: Click on "count" ref, enable history

**System Response**:
- History timeline appears
- Shows all value changes
- Timestamps and sources visible

**History Display**:
```
State History: count
────────────────────────────────────────
Time         Value  Source           Δ
────────────────────────────────────────
17:15:23.001    42  increment event  +1
17:15:22.998    41  increment event  +1
17:15:22.512    40  increment event  +1
17:15:21.023    39  increment event  +1
17:15:20.001     0  onMounted hook   +0
────────────────────────────────────────
Total changes: 42
```

**User Observes**: Count increments every time, normal behavior

#### Step 5: Check Event Flow
**User Action**: Switch to Events tab

**System Response**:
- Event list displays
- Real-time updates as events fire
- Filter and search available

**Event List**:
```
┌─ Recent Events ──────────────────────┐
│ [17:15:23.001] increment → Counter  ││
│   source: KeyboardEvent            ││
│   handler: incrementHandler         ││
│   duration: 0.2ms                   ││
│                                      │
│ [17:15:22.998] increment → Counter  ││
│   source: KeyboardEvent            ││
│   handler: incrementHandler         ││
│   duration: 0.2ms                   ││
│                                      │
│ [17:15:21.023] increment → Counter  ││
│   source: KeyboardEvent            ││
│   handler: incrementHandler         ││
│   duration: 0.3ms                   ││
└──────────────────────────────────────┘
Filter: [          ] 📊 42 events total
```

**User Realizes**: All events processing correctly

#### Step 6: Find the Bug
**User Action**: Notice component not re-rendering despite state changes

**Action**: Check Performance tab

**Performance View Shows**:
```
Component Performance
─────────────────────────────────────────
Component    Renders  Avg Time  Max Time
─────────────────────────────────────────
App              100    1.2ms     3.4ms
Header            50    0.8ms     1.2ms
Counter            5    1.1ms     2.1ms  ⚠️
─────────────────────────────────────────

⚠️ WARNING: Counter component has only
   rendered 5 times despite 42 state
   changes. Check if Update() returning
   commands correctly.
```

**User Found Bug**: Component Update() not batching commands properly!

**Journey Milestone**: ✅ Bug identified using dev tools!

---

### Feature Journey: Performance Optimization

#### Step 7: Identify Slow Components
**User Action**: Open Performance Monitor, sort by avg time

**System Response**:
- Components sorted by render time
- Slow components highlighted
- Flame graph available

**Performance Data**:
```
┌─ Performance Monitor ────────────────┐
│                                      │
│ Slowest Components:                  │
│                                      │
│ 1. DataTable         8.5ms avg  🔴   │
│    └─ Rendered 150 times            │
│    └─ 1.275s total                  │
│                                      │
│ 2. SearchBar         2.1ms avg  🟡   │
│    └─ Rendered 300 times            │
│    └─ 0.630s total                  │
│                                      │
│ 3. SidebarMenu       0.8ms avg  🟢   │
│    └─ Rendered 50 times             │
│    └─ 0.040s total                  │
│                                      │
│ [View Flame Graph]                   │
└──────────────────────────────────────┘
```

**User Identifies**: DataTable is bottleneck

#### Step 8: Analyze with Flame Graph
**User Action**: Click "View Flame Graph"

**Flame Graph Shows**:
```
DataTable ████████████████████████████████ 100% (8.5ms)
├─ TableHeader █ 10% (0.85ms)
├─ TableBody ███████████████████████ 80% (6.8ms)
│  ├─ TableRow (repeated) ████████████ 60% (5.1ms)
│  │  ├─ Cell █ 10% (0.85ms)
│  │  ├─ Cell █ 10% (0.85ms)
│  │  └─ Cell █ 10% (0.85ms)
│  └─ VirtualScroll ███ 20% (1.7ms)
└─ TableFooter █ 10% (0.85ms)
```

**User Realizes**: TableRow rendering is slow, need virtualization

**Journey Milestone**: ✅ Optimization target identified!

---

## Alternative Workflows

### Workflow A: Remote Debugging

#### Entry: Team Member Needs Help

**Scenario**: Teammate has bug, needs your help

1. **Export Debug State**
```go
devtools.Export("bug-state.json", devtools.ExportOptions{
    IncludeState: true,
    IncludeEvents: true,
    IncludeTimeline: true,
    Sanitize: true,  // Remove sensitive data
})
```

2. **Share File**
- Send `bug-state.json` via Slack/email
- File contains full debug context

3. **Import on Your Machine**
```go
devtools.Import("bug-state.json")
```

4. **Inspect Remotely**
- See exact state at bug time
- Replay events
- Analyze timeline

**Result**: Debug without running teammate's environment

---

### Workflow B: Learning Framework

#### Entry: New to BubblyUI

**Scenario**: Learning how reactivity works

1. **Open Example App with Dev Tools**
```bash
BUBBLY_DEV_TOOLS=1 go run examples/counter/main.go
```

2. **Observe Component Tree**
- See parent-child relationships
- Understand component structure

3. **Watch State Changes Live**
- Click button
- See ref value change in real-time
- Observe watcher triggers

4. **Track Event Flow**
- See event emission
- Watch bubbling path
- Understand handler execution

5. **Study Performance**
- See render timing
- Understand update cycles
- Learn optimization patterns

**Result**: Learn framework internals by observation

---

### Workflow C: Using Pattern Templates for Compliance

#### Entry: Need PCI-DSS Compliance for Export

**Scenario**: Exporting debug data that may contain payment information

1. **Create Sanitizer with PCI Template**
```go
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplate("pci")  // Loads credit card, CVV, expiry patterns
```

2. **Add Custom Patterns**
```go
// Add organization-specific patterns
sanitizer.AddPatternWithPriority(
    `(?i)(merchant[_-]?id)(["'\s:=]+)([A-Z0-9]+)`,
    "${1}${2}[REDACTED_MERCHANT]",
    80,  // High priority
    "merchant_id",
)
```

3. **Review Pattern Coverage**
```go
patterns := sanitizer.GetPatterns()
fmt.Printf("Loaded %d patterns\n", len(patterns))
// Output: Loaded 4 patterns (3 PCI + 1 custom)
```

4. **Export with Sanitization**
```go
devtools.Export("debug-data.json", devtools.ExportOptions{
    Sanitize:  sanitizer,
    Templates: []string{"pci"},
})
```

5. **Validate Compliance**
- Review exported file
- Verify no card numbers visible
- Check audit logs

**Result**: PCI-compliant debug exports ready for sharing

**Available Templates**:
- `"pii"` - SSN, email, phone (GDPR, CCPA)
- `"pci"` - Card numbers, CVV, expiry dates
- `"hipaa"` - Medical records, diagnoses
- `"gdpr"` - IP addresses, MAC addresses

---

### Workflow D: Preview Sanitization with Dry-Run

#### Entry: Unsure What Will Be Redacted

**Scenario**: Testing new sanitization patterns before applying

1. **Enable Dry-Run Mode**
```go
sanitizer := devtools.NewSanitizer()
sanitizer.AddPatternWithPriority(
    `(?i)(internal[_-]?id)(["'\s:=]+)([A-Z0-9-]+)`,
    "${1}${2}[REDACTED_ID]",
    50,
    "internal_id",
)

result := sanitizer.Preview(exportData)
```

2. **Review Matches**
```go
fmt.Printf("Would redact %d values\n", result.WouldRedactCount)

for _, match := range result.Matches {
    fmt.Printf("  %s: %s → %s\n", 
        match.Path,
        match.Original,
        match.Redacted,
    )
}
```

**Output**:
```
Would redact 3 values
  components[0].props.password: secret123 → [REDACTED]
  components[1].state.apiKey: sk_live_abc... → [REDACTED]
  state[0].new.internal_id: INT-12345 → [REDACTED_ID]
```

3. **Adjust Patterns if Needed**
```go
// Pattern too broad? Refine it
sanitizer.AddPattern(
    `(?i)(internal[_-]?id)(["'\s:=]+)(INT-[A-Z0-9]+)`,  // More specific
    "${1}${2}[REDACTED_ID]",
)
```

4. **Run Again to Verify**
```go
result = sanitizer.Preview(exportData)
// Check matches are now correct
```

5. **Apply for Real**
```go
cleanData := sanitizer.Sanitize(exportData)
```

**Result**: Validated patterns before applying to production data

---

### Workflow E: Large Export with Streaming

#### Entry: Need to Export 500MB Debug Data

**Scenario**: Application has been running for days, huge export

1. **Check Data Size**
```go
dataSize := devtools.EstimateExportSize()
fmt.Printf("Export will be approximately %d MB\n", dataSize/1024/1024)
// Output: Export will be approximately 512 MB
```

2. **Use Streaming API**
```go
err := devtools.ExportStream("large-export.json", devtools.ExportOptions{
    IncludeState:      true,
    IncludeEvents:     true,
    UseStreaming:      true,
    ProgressCallback: func(bytes int64) {
        mb := bytes / 1024 / 1024
        fmt.Printf("\rProcessed: %d MB", mb)
    },
})
```

**Console Output**:
```
Processed: 0 MB
Processed: 64 MB
Processed: 128 MB
...
Processed: 512 MB
✓ Export complete (2m 15s)
```

3. **Verify Memory Usage**
```bash
# Memory stays under 100MB even for 512MB file
ps aux | grep myapp
# myapp  87.5 MB
```

4. **Import for Analysis**
```go
// Also uses streaming for import
devtools.ImportStream("large-export.json")
```

**Result**: Handled large exports without OOM, constant memory usage

**When to Use Streaming**:
- File size >100MB
- Long-running applications
- Memory-constrained environments
- CI/CD debug logs

---

### Workflow F: Audit Sanitization with Metrics

#### Entry: Need to Verify Data Protection

**Scenario**: Security audit requires proof of sanitization

1. **Sanitize with Metrics Enabled**
```go
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "pci")

cleanData := sanitizer.Sanitize(exportData)
```

2. **Review Statistics**
```go
stats := sanitizer.GetLastStats()
fmt.Printf("Sanitization Report:\n")
fmt.Printf("  Redacted: %d values\n", stats.RedactedCount)
fmt.Printf("  Duration: %v\n", stats.Duration)
fmt.Printf("  Data Size: %d bytes\n", stats.BytesProcessed)
fmt.Println("\nPattern Breakdown:")
for name, count := range stats.PatternMatches {
    fmt.Printf("  %s: %d\n", name, count)
}
```

**Output**:
```
Sanitization Report:
  Redacted: 47 values
  Duration: 142ms
  Data Size: 2,458,624 bytes

Pattern Breakdown:
  password: 23
  token: 15
  apikey: 9
  credit_card: 0
  ssn: 0
```

3. **Export Metrics for Audit**
```go
metricsJSON, _ := stats.JSON()
os.WriteFile("sanitization-audit.json", metricsJSON, 0644)
```

4. **Document for Compliance**
```markdown
# Sanitization Audit Report

Date: 2024-01-15
Export: debug-session-20240115.json
Templates: PII, PCI

## Results
- ✅ 47 sensitive values redacted
- ✅ 23 passwords removed
- ✅ 15 tokens removed
- ✅ 9 API keys removed
- ✅ 0 credit cards found (none present)
- ✅ Processing time: 142ms

## Verification
All sensitive patterns successfully redacted.
No false negatives detected in spot check.
```

**Result**: Documented proof of data protection for security audits

---

## Error Recovery Workflows

### Error Flow 1: Dev Tools Crash

**Trigger**: Dev tools encounters error during inspection

**User Sees**:
```
┌─ Dev Tools Error ────────────────────┐
│ Dev tools encountered an error and   │
│ has been temporarily disabled.       │
│                                      │
│ Error: panic in component inspector  │
│ Location: tree_view.go:142          │
│                                      │
│ Your application continues running.  │
│ Press F12 to retry dev tools.       │
│                                      │
│ [Export Error Report] [Disable]     │
└──────────────────────────────────────┘
```

**Recovery**:
1. Application continues (dev tools isolated)
2. User can export error report
3. Retry enabling dev tools
4. Or disable for remainder of session

**Result**: Application never crashes from dev tools bugs

---

### Error Flow 2: Too Much Data

**Trigger**: 10,000 components, dev tools slow

**User Sees**:
```
⚠️ Performance Warning

Dev tools detected 10,000+ components.
Display may be slow.

Suggestions:
• Enable component filtering
• Use search instead of browsing
• Collapse unused tree sections
• Increase pagination size

[Auto-Optimize] [Continue Anyway]
```

**Auto-Optimize Does**:
- Enables virtual scrolling
- Sets tree pagination to 100
- Disables real-time updates
- Shows aggregated metrics only

**Result**: Dev tools remain usable with large apps

---

## State Transition Diagrams

### Dev Tools Lifecycle
```
Disabled
    ↓
User enables (F12 or code)
    ↓
Initializing
    ├─ Install hooks
    ├─ Create store
    └─ Build UI
    ↓
Active (Visible)
    ├─ Collecting data
    ├─ Updating display
    └─ Handling input
    ↓
User toggles (F12)
    ↓
Active (Hidden)
    ├─ Still collecting data
    └─ Not displaying
    ↓
User closes app
    ↓
Cleanup
    ├─ Export data (optional)
    ├─ Uninstall hooks
    └─ Free memory
    ↓
Terminated
```

### Inspection Flow
```
Component Tree Loaded
    ↓
User selects component
    ↓
Detail Panel Updates
    ↓
User switches tab
    ├─ State → Show refs/computed
    ├─ Props → Show properties
    ├─ Events → Show event history
    └─ Performance → Show metrics
    ↓
User edits state (optional)
    ↓
Confirmation prompt
    ↓
State updated in app
    ↓
App re-renders
    ↓
Dev tools shows new state
```

---

## Integration Points Map

### Feature Cross-Reference
```
09-dev-tools
    ← Inspects: 01-reactivity-system (state)
    ← Inspects: 02-component-model (components)
    ← Inspects: 03-lifecycle-hooks (lifecycle)
    ← Inspects: 04-composition-api (composables)
    ← Inspects: 05-directives (directives)
    ← Inspects: 07-router (routes)
    ← Inspects: 08-automatic-reactive-bridge (commands)
    → Used by: Developers (debugging)
    → Used by: 11-performance-profiler (data source)
```

---

## User Success Paths

### Path 1: Quick Debug (< 10 minutes)
```
Bug occurs → Enable dev tools → Inspect state → Find issue → Fix → Success! 🎉
Time saved: 2+ hours
```

### Path 2: Performance Tuning (< 30 minutes)
```
App slow → Open perf monitor → Find bottleneck → Optimize → Verify → Success! 🎉
Performance gain: 2-10x
```

### Path 3: Learning (< 2 hours)
```
New to framework → Enable dev tools → Explore → Understand → Build confidence → Success! 🎉
Learning accelerated: 5x faster
```

---

## Common Patterns

### Pattern 1: State Debugging
```go
// 1. Enable dev tools
devtools.Enable()

// 2. Run app
tea.NewProgram(app).Run()

// 3. Press F12 to open
// 4. Navigate to component in tree
// 5. View State tab
// 6. Watch history of changes
// 7. Identify when state went wrong
```

### Pattern 2: Event Tracing
```go
// 1. Open dev tools
// 2. Switch to Events tab
// 3. Apply filter for specific event
// 4. Trigger event in app
// 5. See event details
// 6. Check handler execution
// 7. Verify payload
```

### Pattern 3: Performance Profiling
```go
// 1. Open dev tools
// 2. Switch to Performance tab
// 3. Sort by avg render time
// 4. Identify slow components
// 5. View flame graph
// 6. Find bottleneck
// 7. Optimize code
// 8. Verify improvement
```

---

## Keyboard Shortcuts

```
F12           Toggle dev tools visibility
Tab           Switch focus (app ↔ tools)
Ctrl+F        Search components
↑↓            Navigate tree/lists
Enter         Select/expand item
Space         Expand/collapse node
Ctrl+E        Export debug data
Ctrl+R        Refresh display
Ctrl+P        Pause/resume updates
?             Show help
Esc           Close panel/dialog
```

---

## Tips & Tricks

### Tip 1: Use Search, Not Browse
For large component trees, use Ctrl+F to search by name instead of browsing.

### Tip 2: Pause During Analysis
Press Ctrl+P to pause updates while analyzing data. Prevents data from changing mid-analysis.

### Tip 3: Export Before Closing
Export debug state before closing app. Useful for later analysis or sharing.

### Tip 4: Filter Events Aggressively
In event tracker, use filters to show only relevant events. Reduces noise.

### Tip 5: Collapse Unused Sections
Collapse tree sections you're not inspecting. Improves performance and visibility.

---

## Summary

The Dev Tools system provides real-time inspection and debugging capabilities through a split-pane interface activated by F12. Developers can inspect component trees, view reactive state changes with history, track event flow, debug routes, monitor performance with flame graphs, and analyze command timelines. The system maintains < 5% overhead, exports debug data for sharing, and includes safety features like application isolation and graceful error handling. Common workflows include state debugging (10 minutes), performance tuning (30 minutes), and framework learning (2 hours), with keyboard-driven navigation and comprehensive inspection capabilities.

**Key Success Factors**:
- ✅ Zero-config enablement (just add devtools.Enable())
- ✅ Real-time inspection (see state as it changes)
- ✅ Low overhead (< 5% performance impact)
- ✅ Application safety (never crashes host app)
- ✅ Export/sharing (bug reports include context)
- ✅ Learning tool (understand framework by observation)
