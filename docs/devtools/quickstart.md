# Quickstart Guide

**Get started with BubblyUI Dev Tools in 5 minutes**

This guide walks you through your first debugging session, from enabling dev tools to inspecting your first component.

## Prerequisites

- Go 1.22 or later
- BubblyUI installed (`go get github.com/newbpydev/bubblyui`)
- Basic familiarity with BubblyUI components

## Step 1: Enable Dev Tools (30 seconds)

Add one line to your application's `main()` function:

```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // ✨ Enable dev tools - that's it!
    devtools.Enable()
    
    // Your existing app code
    counter := NewCounter()
    p := tea.NewProgram(counter, tea.WithAltScreen())
    p.Run()
}
```

**That's it!** No configuration needed.

## Step 2: Run Your Application (10 seconds)

```bash
go run main.go
```

Your application starts normally. Dev tools are enabled but hidden by default.

## Step 3: Toggle Dev Tools Visibility (5 seconds)

Press `F12` to show/hide dev tools.

```
┌────────────────────────────────────────────────────────┐
│ Your App           │ Dev Tools                         │
│                    │ ┌───────────────────────────────┐ │
│ Counter: 0         │ │ Components  State  Events  ⚡ │ │
│                    │ ├───────────────────────────────┤ │
│ [+] [-]            │ │ Component Tree                │ │
│                    │ │ ├─ App                        │ │
│                    │ │ │  ├─ Counter (selected)      │ │
│                    │ │ │  └─ Footer                  │ │
│                    │ │                               │ │
│                    │ │ State:                        │ │
│                    │ │ • count: 0 (Ref)              │ │
│                    │ │                               │ │
│ F12: Toggle Tools  │ │ Events: 0 captured            │ │
└────────────────────────────────────────────────────────┘
```

## Step 4: Explore Component Inspector (1 minute)

The **Component Inspector** shows your component hierarchy.

### Navigate the Tree

- `↑`/`↓` - Select component
- `→` - Expand node
- `←` - Collapse node
- `Enter` - View details

### Try it:

1. Press `↓` to select "Counter"
2. Press `Tab` to switch to State tab
3. See `count: 0` displayed

## Step 5: Track State Changes (1 minute)

Click the `[+]` button in your app and watch the dev tools:

```
State:
• count: 1 (Ref)  ← Changed!
  Previous: 0
  Changed: 2024-11-11 11:30:45
```

Dev tools automatically track all state changes with timestamps.

### View State History

1. Press `Tab` twice to reach State Viewer
2. Select `count` ref
3. Press `h` to view history

```
State History: count
─────────────────────────
0 → 1  (11:30:45)  Click
1 → 2  (11:30:47)  Click
2 → 3  (11:30:49)  Click
```

## Step 6: Monitor Events (1 minute)

Switch to the **Events** tab to see component events:

1. Press `Tab` to reach Event Tracker
2. Click `[+]` button in your app
3. See events appear in real-time

```
Events
─────────────────────────────────────
[11:30:45] click → Button#btn-plus
[11:30:45] increment → Counter
[11:30:47] click → Button#btn-plus
[11:30:47] increment → Counter
```

### Filter Events

- Press `/` to start filtering
- Type `"click"` to show only click events
- Press `Esc` to clear filter

## Step 7: Check Performance (1 minute)

Switch to the **Performance** tab:

```
Performance Monitor
────────────────────────────────────
Component      Renders  Avg Time  Max Time
─────────────────────────────────────────
Counter        5        2.1ms     3.2ms
Button         10       0.8ms     1.1ms
Footer         1        0.5ms     0.5ms
```

**Green zones:** < 10ms  
**Yellow zones:** 10-50ms  
**Red zones:** > 50ms (needs optimization)

## Step 8: Export Debug Session (30 seconds)

Save your debugging session for later analysis or sharing:

```go
// Add this to your app (or use dev tools UI)
dt := devtools.Get()
dt.Export("debug-session.json", devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
})
```

Or press `Ctrl+E` in dev tools UI to export.

## Common Workflows

### Debugging State Issues

**Problem:** "Why isn't my counter updating?"

1. Open dev tools (`F12`)
2. Switch to State tab
3. Watch `count` ref as you click buttons
4. Check state history (`h`) for unexpected changes

**Look for:**
- Is the ref actually changing? (should see value updates)
- Are there duplicate changes? (possible unnecessary re-renders)
- Are changes coming from expected sources?

### Finding Slow Components

**Problem:** "My app feels sluggish"

1. Open Performance tab
2. Interact with your app
3. Sort by "Max Time" column
4. Look for components > 50ms

**Action:**
- Optimize components in red zones first
- Check if expensive operations are in `View()` (should be in `Update()`)
- Consider memoization for computed values

### Tracking Event Flow

**Problem:** "Events aren't reaching parent components"

1. Open Events tab
2. Trigger the event
3. Look for event in log
4. Check source and target components

**Look for:**
- Is event being emitted? (should appear in log)
- Is event name correct? (case-sensitive)
- Is event being handled? (check handler registration)

## Keyboard Shortcuts Quick Reference

| Key         | Action                       |
|-------------|------------------------------|
| `F12`       | Toggle dev tools visibility  |
| `Tab`       | Next panel                   |
| `Shift+Tab` | Previous panel               |
| `↑`/`↓`     | Navigate items               |
| `→`/`←`     | Expand/collapse nodes        |
| `Enter`     | Select/activate              |
| `/`         | Start filter/search          |
| `Esc`       | Cancel/close                 |
| `Ctrl+E`    | Export session               |
| `Ctrl+C`    | Quit application             |

See [Reference](./reference.md) for complete shortcuts.

## Configuration Options

### Change Layout Mode

```go
config := devtools.DefaultConfig()
config.LayoutMode = devtools.LayoutHorizontal  // Side-by-side (default)
// Or: LayoutVertical (stacked), LayoutOverlay (full-screen toggle)

devtools.EnableWithConfig(config)
```

### Adjust Split Ratio

```go
config.SplitRatio = 0.70  // 70% app, 30% tools (default: 60/40)
```

### Use Environment Variables

```bash
export BUBBLY_DEVTOOLS_ENABLED=true
export BUBBLY_DEVTOOLS_LAYOUT_MODE=horizontal
export BUBBLY_DEVTOOLS_SPLIT_RATIO=0.60

go run main.go
```

## Next Steps

Now that you've completed the quickstart:

1. **Learn about all features** - [Features Overview](./features.md)
2. **Understand framework hooks** - [Framework Hooks Guide](./hooks.md)
3. **Export and share sessions** - [Export & Import Guide](./export-import.md)
4. **Optimize performance** - [Best Practices](./best-practices.md)
5. **Explore examples** - [`cmd/examples/09-devtools/`](../../cmd/examples/09-devtools/)

## Troubleshooting

### Dev Tools Not Showing?

**Check:**
- Did you call `devtools.Enable()` before creating components?
- Press `F12` to toggle visibility
- Check `devtools.IsEnabled()` returns true

### Performance Issues?

**Try:**
- Reduce `MaxComponents` and `MaxEvents` limits
- Lower `SamplingRate` to 0.5 (50% of events)
- Use `LayoutOverlay` mode instead of split-pane

### Can't See My Component?

**Verify:**
- Component implements `Component` interface correctly
- Component was created after `devtools.Enable()`
- Component was mounted (called `Init()`)

See [Troubleshooting Guide](./troubleshooting.md) for more solutions.

---

**Questions?** Check the [FAQ](./troubleshooting.md#faq) or open an issue on GitHub.
