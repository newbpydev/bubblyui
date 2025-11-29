# User Workflow: Enhanced Composables Library

## Primary User Journey: Building a Responsive TUI Application

### Entry Point
Developer creates a new BubblyUI application that needs responsive layout, focus management, and state utilities.

### Step 1: Import Composables
**User Action:** Import the composables package

```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)
```

**System Response:** Composables are available for use in Setup functions.

### Step 2: Setup Responsive Layout with UseWindowSize
**User Action:** Add window size handling to the app

```go
func CreateApp() (bubbly.Component, error) {
    return bubbly.NewComponent("App").
        Setup(func(ctx *bubbly.Context) {
            // Create window size composable - THAT'S IT!
            // UseWindowSize automatically receives resize events from the framework
            // No WithMessageHandler or manual event handling needed
            windowSize := composables.UseWindowSize(ctx)
            ctx.Expose("windowSize", windowSize)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            ws := ctx.Get("windowSize").(*composables.WindowSizeReturn)
            bp := ws.Breakpoint.GetTyped()
            
            // Adapt layout based on breakpoint
            if bp == composables.BreakpointXS || bp == composables.BreakpointSM {
                return renderMobileLayout(ctx)
            }
            return renderDesktopLayout(ctx)
        }).
        Build()
}
```

**UI Feedback:** Layout automatically adapts when terminal is resized.

**Note (Phase 6 Enhancement):** The framework automatically:
1. Detects `tea.WindowSizeMsg` in component updates
2. Emits a "windowResize" event
3. UseWindowSize auto-subscribes and updates its state

Users never need to import or handle Bubbletea types for responsive layouts.

### Step 3: Add Multi-Pane Focus Management
**User Action:** Implement focus cycling between panes

```go
type FocusPane int
const (
    FocusSidebar FocusPane = iota
    FocusMain
    FocusFooter
)

Setup(func(ctx *bubbly.Context) {
    // Create focus composable
    focus := composables.UseFocus(ctx, FocusMain, []FocusPane{
        FocusSidebar, FocusMain, FocusFooter,
    })
    ctx.Expose("focus", focus)
    
    // Handle Tab key for focus cycling
    ctx.On("nextFocus", func(_ interface{}) {
        focus.Next()
    })
    
    ctx.On("prevFocus", func(_ interface{}) {
        focus.Previous()
    })
}).
WithKeyBinding("tab", "nextFocus", "Next pane").
WithKeyBinding("shift+tab", "prevFocus", "Previous pane")
```

**UI Feedback:** Tab cycles focus between panes with visual indicators.

### Step 4: Implement Scrollable Content
**User Action:** Add scroll management for a list view

```go
Setup(func(ctx *bubbly.Context) {
    items := []string{"Item 1", "Item 2", ..., "Item 100"}
    visibleCount := 10
    
    scroll := composables.UseScroll(ctx, len(items), visibleCount)
    ctx.Expose("scroll", scroll)
    
    ctx.On("scrollUp", func(_ interface{}) {
        scroll.ScrollUp()
    })
    ctx.On("scrollDown", func(_ interface{}) {
        scroll.ScrollDown()
    })
    ctx.On("pageUp", func(_ interface{}) {
        scroll.PageUp()
    })
    ctx.On("pageDown", func(_ interface{}) {
        scroll.PageDown()
    })
}).
WithMultiKeyBindings("scrollUp", "Scroll up", "up", "k").
WithMultiKeyBindings("scrollDown", "Scroll down", "down", "j")
```

**UI Feedback:** List scrolls smoothly with keyboard navigation.

### Step 5: Add Selection Management
**User Action:** Implement item selection in list

```go
Setup(func(ctx *bubbly.Context) {
    items := []Item{...}
    
    selection := composables.UseSelection(ctx, items, 
        composables.WithWrap(true))
    ctx.Expose("selection", selection)
    
    ctx.On("select", func(_ interface{}) {
        item := selection.SelectedItem.Get()
        // Handle selection
    })
}).
WithKeyBinding("enter", "select", "Select item")
```

**UI Feedback:** Selected item is highlighted, Enter triggers action.

### Step 6: Implement Navigation/Input Modes
**User Action:** Add vim-like mode switching

```go
type Mode string
const (
    ModeNavigation Mode = "navigation"
    ModeInput      Mode = "input"
)

Setup(func(ctx *bubbly.Context) {
    mode := composables.UseMode(ctx, ModeNavigation)
    ctx.Expose("mode", mode)
    
    ctx.On("toggleMode", func(_ interface{}) {
        mode.Toggle(ModeNavigation, ModeInput)
    })
    
    ctx.On("enterInput", func(_ interface{}) {
        mode.Switch(ModeInput)
    })
    
    ctx.On("exitInput", func(_ interface{}) {
        mode.Switch(ModeNavigation)
    })
}).
WithKeyBinding("i", "enterInput", "Enter input mode").
WithKeyBinding("esc", "exitInput", "Exit to navigation")
```

**UI Feedback:** Mode indicator updates, key bindings change based on mode.

---

## Alternative Workflows

### Workflow A: Using State Utilities

```go
Setup(func(ctx *bubbly.Context) {
    // Simple boolean toggle
    darkMode := composables.UseToggle(ctx, false)
    ctx.Expose("darkMode", darkMode)
    
    // Bounded counter
    volume := composables.UseCounter(ctx, 50,
        composables.WithMin(0),
        composables.WithMax(100),
        composables.WithStep(5))
    ctx.Expose("volume", volume)
    
    // Undo/redo history
    history := composables.UseHistory(ctx, initialState, 50)
    ctx.Expose("history", history)
    
    ctx.On("undo", func(_ interface{}) {
        history.Undo()
    })
    
    ctx.On("redo", func(_ interface{}) {
        history.Redo()
    })
})
```

### Workflow B: Using Timing Composables

```go
Setup(func(ctx *bubbly.Context) {
    // Auto-refresh every 5 seconds
    autoRefresh := composables.UseInterval(ctx, func() {
        fetchData()
    }, 5*time.Second)
    ctx.Expose("autoRefresh", autoRefresh)
    
    // Countdown timer
    timer := composables.UseTimer(ctx, 60*time.Second,
        composables.WithOnExpire(func() {
            ctx.Emit("timerExpired", nil)
        }))
    ctx.Expose("timer", timer)
    
    ctx.On("startTimer", func(_ interface{}) {
        timer.Start()
    })
})
```

### Workflow C: Using Collection Composables

```go
Setup(func(ctx *bubbly.Context) {
    // Task list with CRUD
    tasks := composables.UseList(ctx, []Task{})
    ctx.Expose("tasks", tasks)
    
    ctx.On("addTask", func(data interface{}) {
        task := data.(Task)
        tasks.Push(task)
    })
    
    ctx.On("removeTask", func(data interface{}) {
        idx := data.(int)
        tasks.RemoveAt(idx)
    })
    
    // Tag management with Set
    tags := composables.UseSet(ctx, []string{"urgent", "todo"})
    ctx.Expose("tags", tags)
    
    ctx.On("toggleTag", func(data interface{}) {
        tag := data.(string)
        tags.Toggle(tag) // Add if not present, remove if present
    })
})
```

### Workflow D: Using Development Composables

```go
Setup(func(ctx *bubbly.Context) {
    // Debug logging
    logger := composables.UseLogger(ctx, "MyComponent")
    ctx.Expose("logger", logger)
    
    logger.Debug("Component initialized")
    
    // Toast notifications
    notifications := composables.UseNotification(ctx,
        composables.WithDefaultDuration(3*time.Second),
        composables.WithMaxNotifications(5))
    ctx.Expose("notifications", notifications)
    
    ctx.On("saved", func(_ interface{}) {
        notifications.Success("Saved", "Changes saved successfully")
    })
    
    ctx.On("error", func(data interface{}) {
        err := data.(error)
        notifications.Error("Error", err.Error())
    })
})
```

---

## Error Handling Flows

### Error 1: Empty Focus Order
**Trigger:** Creating UseFocus with empty order slice

**User Sees:** Panic with clear error message

**Recovery:** 
```go
// Always provide at least one focus pane
focus := composables.UseFocus(ctx, FocusMain, []FocusPane{FocusMain})
```

### Error 2: Invalid Counter Bounds
**Trigger:** Set min greater than max

**User Sees:** Panic with bounds error

**Recovery:**
```go
// Ensure min < max
counter := composables.UseCounter(ctx, 50,
    composables.WithMin(0),   // min
    composables.WithMax(100)) // max > min
```

### Error 3: Negative Timer Duration
**Trigger:** UseTimer with negative duration

**User Sees:** Panic with duration error

**Recovery:**
```go
// Use positive duration
timer := composables.UseTimer(ctx, 60*time.Second) // positive
```

---

## State Transitions

```
Initial State
    │
    ▼
UseWindowSize.SetSize() → Breakpoint calculated → SidebarVisible updated
    │
    ▼
UseFocus.Next() → Current updated → Previous pane loses focus
    │
    ▼
UseScroll.ScrollDown() → Offset incremented (within bounds)
    │
    ▼
UseSelection.SelectNext() → SelectedIndex updated → SelectedItem computed
    │
    ▼
UseMode.Switch() → Current mode changes → Previous mode stored
    │
    ▼
UseHistory.Push() → State added → Redo stack cleared
    │
    ▼
View re-renders with new state
```

---

## Integration Points

### Connected Features
- **04-composition-api**: Base composable infrastructure
- **01-reactivity-system**: Ref, Computed, Watch
- **14-advanced-layout-system**: Responsive layout components
- **09-dev-tools**: Logger integration

### Data Shared Between Features
- Window dimensions → Layout components
- Focus state → Border colors, key binding activation
- Mode state → Key binding conditional activation
- Notification state → Toast component rendering

### Navigation
- Composables are used in Setup phase
- State accessed in Template phase
- Events trigger composable methods in Update phase

---

## Common Patterns

### Pattern 1: Shared Composable Across Components

```go
// composables/use_shared_focus.go
var UseSharedFocus = composables.CreateShared(
    func(ctx *bubbly.Context) *composables.FocusReturn[FocusPane] {
        return composables.UseFocus(ctx, FocusMain, []FocusPane{
            FocusSidebar, FocusMain, FocusFooter,
        })
    },
)

// In any component
focus := localComposables.UseSharedFocus(ctx)
```

### Pattern 2: Combining Multiple Composables

```go
Setup(func(ctx *bubbly.Context) {
    // Combine window size + focus + scroll
    ws := composables.UseWindowSize(ctx)
    focus := composables.UseFocus(ctx, FocusMain, panes)
    scroll := composables.UseScroll(ctx, 100, ws.Height.GetTyped()-5)
    
    // Update scroll when window resizes
    ctx.Watch(ws.Height, func(newVal, oldVal interface{}) {
        scroll.SetVisibleCount(newVal.(int) - 5)
    })
})
```

### Pattern 3: Mode-Dependent Key Bindings

```go
Setup(func(ctx *bubbly.Context) {
    mode := composables.UseMode(ctx, ModeNavigation)
    
    // Only active in navigation mode
    ctx.On("navigate", func(_ interface{}) {
        if mode.IsMode(ModeNavigation) {
            // Handle navigation
        }
    })
    
    // Only active in input mode
    ctx.On("input", func(data interface{}) {
        if mode.IsMode(ModeInput) {
            // Handle input
        }
    })
})
```
