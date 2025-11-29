# Design Specification: Enhanced Composables Library

## Architecture Overview

```
pkg/bubbly/composables/
├── Existing (12):
│   ├── use_state.go
│   ├── use_async.go
│   ├── use_effect.go
│   ├── use_debounce.go
│   ├── use_throttle.go
│   ├── use_form.go
│   ├── use_local_storage.go
│   ├── use_event_listener.go
│   ├── use_text_input.go
│   └── shared.go (CreateShared)
│
├── TUI-Specific (5):
│   ├── use_window_size.go      # Terminal dimensions & breakpoints
│   ├── use_focus.go            # Multi-pane focus management
│   ├── use_scroll.go           # Viewport scrolling
│   ├── use_selection.go        # List/table selection
│   └── use_mode.go             # Navigation/input modes
│
├── State Utilities (4):
│   ├── use_toggle.go           # Boolean toggle
│   ├── use_counter.go          # Bounded counter
│   ├── use_previous.go         # Previous value tracking
│   └── use_history.go          # Undo/redo stack
│
├── Timing (3):
│   ├── use_interval.go         # Periodic execution
│   ├── use_timeout.go          # Delayed execution
│   └── use_timer.go            # Countdown timer
│
├── Collections (4):
│   ├── use_list.go             # List CRUD
│   ├── use_map.go              # Key-value state
│   ├── use_set.go              # Unique values
│   └── use_queue.go            # FIFO queue
│
└── Development (2):
    ├── use_logger.go           # Debug logging
    └── use_notification.go     # Toast notifications
```

## Type Definitions

### TUI-Specific Types

```go
// =====================================
// UseWindowSize
// =====================================

// Breakpoint represents responsive breakpoints for terminal layouts.
type Breakpoint string

const (
    BreakpointXS Breakpoint = "xs"  // <60 cols
    BreakpointSM Breakpoint = "sm"  // 60-79 cols
    BreakpointMD Breakpoint = "md"  // 80-119 cols
    BreakpointLG Breakpoint = "lg"  // 120-159 cols
    BreakpointXL Breakpoint = "xl"  // 160+ cols
)

// BreakpointConfig allows custom breakpoint thresholds.
type BreakpointConfig struct {
    XS int // Default: 0
    SM int // Default: 60
    MD int // Default: 80
    LG int // Default: 120
    XL int // Default: 160
}

// WindowSizeReturn is the return value of UseWindowSize.
type WindowSizeReturn struct {
    // Width is the current terminal width in columns.
    Width *bubbly.Ref[int]
    
    // Height is the current terminal height in rows.
    Height *bubbly.Ref[int]
    
    // Breakpoint is the current responsive breakpoint.
    Breakpoint *bubbly.Ref[Breakpoint]
    
    // SidebarVisible indicates if sidebar should be visible.
    SidebarVisible *bubbly.Ref[bool]
    
    // GridColumns is the recommended number of grid columns.
    GridColumns *bubbly.Ref[int]
}

// SetSize updates the window dimensions and recalculates derived values.
func (w *WindowSizeReturn) SetSize(width, height int)

// GetContentWidth returns available content width (accounting for sidebar).
func (w *WindowSizeReturn) GetContentWidth() int

// GetCardWidth returns optimal card width for current grid.
func (w *WindowSizeReturn) GetCardWidth() int

// WindowSizeOption configures UseWindowSize.
type WindowSizeOption func(*windowSizeConfig)

// WithBreakpoints sets custom breakpoint thresholds.
func WithBreakpoints(config BreakpointConfig) WindowSizeOption

// WithMinDimensions sets minimum width and height.
func WithMinDimensions(minWidth, minHeight int) WindowSizeOption

// WithSidebarWidth sets sidebar width for content calculation.
func WithSidebarWidth(width int) WindowSizeOption

// UseWindowSize creates a window size composable for responsive layouts.
// 
// IMPORTANT (Phase 6 Enhancement): UseWindowSize automatically subscribes to
// the framework's "windowResize" event. Users do NOT need to:
// - Use WithMessageHandler to catch tea.WindowSizeMsg
// - Manually emit or handle resize events
// - Call SetSize() manually on window resize
//
// The framework automatically:
// 1. Detects tea.WindowSizeMsg in componentImpl.Update()
// 2. Emits a "windowResize" event with width/height
// 3. UseWindowSize receives this event and updates automatically
func UseWindowSize(ctx *bubbly.Context, opts ...WindowSizeOption) *WindowSizeReturn

// =====================================
// UseFocus
// =====================================

// FocusReturn is the return value of UseFocus.
type FocusReturn[T comparable] struct {
    // Current is the currently focused pane.
    Current *bubbly.Ref[T]
    
    // order is the focus cycle order.
    order []T
}

// IsFocused returns true if the given pane is currently focused.
func (f *FocusReturn[T]) IsFocused(pane T) bool

// Focus sets focus to the specified pane.
func (f *FocusReturn[T]) Focus(pane T)

// Next moves focus to the next pane in order.
func (f *FocusReturn[T]) Next()

// Previous moves focus to the previous pane in order.
func (f *FocusReturn[T]) Previous()

// UseFocus creates a focus management composable.
func UseFocus[T comparable](ctx *bubbly.Context, initial T, order []T) *FocusReturn[T]

// =====================================
// UseScroll
// =====================================

// ScrollReturn is the return value of UseScroll.
type ScrollReturn struct {
    // Offset is the current scroll position (0-indexed).
    Offset *bubbly.Ref[int]
    
    // MaxOffset is the maximum valid scroll offset.
    MaxOffset *bubbly.Ref[int]
    
    // VisibleCount is the number of visible items.
    VisibleCount *bubbly.Ref[int]
    
    // TotalItems is the total number of items.
    TotalItems *bubbly.Ref[int]
}

// ScrollUp moves the scroll position up by one.
func (s *ScrollReturn) ScrollUp()

// ScrollDown moves the scroll position down by one.
func (s *ScrollReturn) ScrollDown()

// ScrollTo moves to a specific offset (clamped to valid range).
func (s *ScrollReturn) ScrollTo(offset int)

// ScrollToTop scrolls to the beginning.
func (s *ScrollReturn) ScrollToTop()

// ScrollToBottom scrolls to the end.
func (s *ScrollReturn) ScrollToBottom()

// PageUp scrolls up by visible count.
func (s *ScrollReturn) PageUp()

// PageDown scrolls down by visible count.
func (s *ScrollReturn) PageDown()

// IsAtTop returns true if scrolled to top.
func (s *ScrollReturn) IsAtTop() bool

// IsAtBottom returns true if scrolled to bottom.
func (s *ScrollReturn) IsAtBottom() bool

// SetTotalItems updates the total item count and recalculates max offset.
func (s *ScrollReturn) SetTotalItems(count int)

// SetVisibleCount updates visible count and recalculates max offset.
func (s *ScrollReturn) SetVisibleCount(count int)

// UseScroll creates a scroll management composable.
func UseScroll(ctx *bubbly.Context, totalItems, visibleCount int) *ScrollReturn

// =====================================
// UseSelection
// =====================================

// SelectionOption configures UseSelection.
type SelectionOption func(*selectionConfig)

// WithWrap enables wrapping at list boundaries.
func WithWrap(wrap bool) SelectionOption

// WithMultiSelect enables multi-selection mode.
func WithMultiSelect(multi bool) SelectionOption

// SelectionReturn is the return value of UseSelection.
type SelectionReturn[T any] struct {
    // SelectedIndex is the currently selected index.
    SelectedIndex *bubbly.Ref[int]
    
    // SelectedItem is the currently selected item (computed).
    SelectedItem *bubbly.Computed[T]
    
    // SelectedIndices is for multi-select mode.
    SelectedIndices *bubbly.Ref[[]int]
    
    // Items is the list of selectable items.
    Items *bubbly.Ref[[]T]
}

// Select sets the selection to a specific index.
func (s *SelectionReturn[T]) Select(index int)

// SelectNext moves selection to the next item.
func (s *SelectionReturn[T]) SelectNext()

// SelectPrevious moves selection to the previous item.
func (s *SelectionReturn[T]) SelectPrevious()

// IsSelected returns true if the index is selected.
func (s *SelectionReturn[T]) IsSelected(index int) bool

// ToggleSelection toggles selection at index (multi-select mode).
func (s *SelectionReturn[T]) ToggleSelection(index int)

// ClearSelection clears all selections.
func (s *SelectionReturn[T]) ClearSelection()

// SetItems updates the items list and adjusts selection.
func (s *SelectionReturn[T]) SetItems(items []T)

// UseSelection creates a selection management composable.
func UseSelection[T any](ctx *bubbly.Context, items []T, opts ...SelectionOption) *SelectionReturn[T]

// =====================================
// UseMode
// =====================================

// ModeReturn is the return value of UseMode.
type ModeReturn[T comparable] struct {
    // Current is the current mode.
    Current *bubbly.Ref[T]
    
    // Previous is the previous mode (for transitions).
    Previous *bubbly.Ref[T]
}

// IsMode returns true if currently in the specified mode.
func (m *ModeReturn[T]) IsMode(mode T) bool

// Switch changes to a new mode.
func (m *ModeReturn[T]) Switch(mode T)

// Toggle switches between two modes.
func (m *ModeReturn[T]) Toggle(a, b T)

// UseMode creates a mode management composable.
func UseMode[T comparable](ctx *bubbly.Context, initial T) *ModeReturn[T]
```

### State Utility Types

```go
// =====================================
// UseToggle
// =====================================

// ToggleReturn is the return value of UseToggle.
type ToggleReturn struct {
    // Value is the current boolean value.
    Value *bubbly.Ref[bool]
}

// Toggle flips the value.
func (t *ToggleReturn) Toggle()

// Set sets the value explicitly.
func (t *ToggleReturn) Set(val bool)

// On sets value to true.
func (t *ToggleReturn) On()

// Off sets value to false.
func (t *ToggleReturn) Off()

// UseToggle creates a boolean toggle composable.
func UseToggle(ctx *bubbly.Context, initial bool) *ToggleReturn

// =====================================
// UseCounter
// =====================================

// CounterOption configures UseCounter.
type CounterOption func(*counterConfig)

// WithMin sets minimum counter value.
func WithMin(min int) CounterOption

// WithMax sets maximum counter value.
func WithMax(max int) CounterOption

// WithStep sets increment/decrement step size.
func WithStep(step int) CounterOption

// CounterReturn is the return value of UseCounter.
type CounterReturn struct {
    // Count is the current counter value.
    Count *bubbly.Ref[int]
    
    // config holds min/max/step settings.
    config counterConfig
    
    // initial is the starting value for reset.
    initial int
}

// Increment increases count by step (respects max).
func (c *CounterReturn) Increment()

// Decrement decreases count by step (respects min).
func (c *CounterReturn) Decrement()

// IncrementBy increases count by n (respects max).
func (c *CounterReturn) IncrementBy(n int)

// DecrementBy decreases count by n (respects min).
func (c *CounterReturn) DecrementBy(n int)

// Set sets the count to a specific value (clamped to bounds).
func (c *CounterReturn) Set(n int)

// Reset resets to initial value.
func (c *CounterReturn) Reset()

// UseCounter creates a bounded counter composable.
func UseCounter(ctx *bubbly.Context, initial int, opts ...CounterOption) *CounterReturn

// =====================================
// UsePrevious
// =====================================

// PreviousReturn is the return value of UsePrevious.
type PreviousReturn[T any] struct {
    // Value is the previous value (nil if no previous).
    Value *bubbly.Ref[*T]
}

// Get returns the previous value (nil if none).
func (p *PreviousReturn[T]) Get() *T

// UsePrevious tracks the previous value of a ref.
func UsePrevious[T any](ctx *bubbly.Context, ref *bubbly.Ref[T]) *PreviousReturn[T]

// =====================================
// UseHistory
// =====================================

// HistoryReturn is the return value of UseHistory.
type HistoryReturn[T any] struct {
    // Current is the current state value.
    Current *bubbly.Ref[T]
    
    // CanUndo indicates if undo is available.
    CanUndo *bubbly.Computed[bool]
    
    // CanRedo indicates if redo is available.
    CanRedo *bubbly.Computed[bool]
    
    // history stack and position
    past   []T
    future []T
    maxSize int
}

// Push adds a new state to history (clears redo stack).
func (h *HistoryReturn[T]) Push(value T)

// Undo reverts to previous state.
func (h *HistoryReturn[T]) Undo()

// Redo restores next state.
func (h *HistoryReturn[T]) Redo()

// Clear clears all history.
func (h *HistoryReturn[T]) Clear()

// UseHistory creates an undo/redo history composable.
func UseHistory[T any](ctx *bubbly.Context, initial T, maxSize int) *HistoryReturn[T]
```

### Timing Types

```go
// =====================================
// UseInterval
// =====================================

// IntervalReturn is the return value of UseInterval.
type IntervalReturn struct {
    // IsRunning indicates if the interval is active.
    IsRunning *bubbly.Ref[bool]
    
    // callback and duration
    callback func()
    duration time.Duration
}

// Start begins the interval.
func (i *IntervalReturn) Start()

// Stop pauses the interval.
func (i *IntervalReturn) Stop()

// Toggle starts if stopped, stops if running.
func (i *IntervalReturn) Toggle()

// Reset stops and restarts the interval.
func (i *IntervalReturn) Reset()

// UseInterval creates a periodic execution composable.
// Uses tea.Tick pattern for Bubbletea integration.
func UseInterval(ctx *bubbly.Context, callback func(), duration time.Duration) *IntervalReturn

// =====================================
// UseTimeout
// =====================================

// TimeoutReturn is the return value of UseTimeout.
type TimeoutReturn struct {
    // IsPending indicates if timeout is pending.
    IsPending *bubbly.Ref[bool]
    
    // IsExpired indicates if timeout has expired.
    IsExpired *bubbly.Ref[bool]
}

// Start begins the timeout.
func (t *TimeoutReturn) Start()

// Cancel cancels the pending timeout.
func (t *TimeoutReturn) Cancel()

// Reset cancels and restarts the timeout.
func (t *TimeoutReturn) Reset()

// UseTimeout creates a delayed execution composable.
func UseTimeout(ctx *bubbly.Context, callback func(), duration time.Duration) *TimeoutReturn

// =====================================
// UseTimer
// =====================================

// TimerReturn is the return value of UseTimer.
type TimerReturn struct {
    // Remaining is the remaining time.
    Remaining *bubbly.Ref[time.Duration]
    
    // IsRunning indicates if timer is running.
    IsRunning *bubbly.Ref[bool]
    
    // IsExpired indicates if timer has expired.
    IsExpired *bubbly.Computed[bool]
    
    // Progress is the completion percentage (0.0 to 1.0).
    Progress *bubbly.Computed[float64]
}

// Start begins the countdown.
func (t *TimerReturn) Start()

// Stop pauses the countdown.
func (t *TimerReturn) Stop()

// Reset restarts from initial duration.
func (t *TimerReturn) Reset()

// TimerOption configures UseTimer.
type TimerOption func(*timerConfig)

// WithOnExpire sets callback when timer expires.
func WithOnExpire(fn func()) TimerOption

// WithTickInterval sets update frequency (default: 100ms).
func WithTickInterval(d time.Duration) TimerOption

// UseTimer creates a countdown timer composable.
func UseTimer(ctx *bubbly.Context, duration time.Duration, opts ...TimerOption) *TimerReturn
```

### Collection Types

```go
// =====================================
// UseList
// =====================================

// ListReturn is the return value of UseList.
type ListReturn[T any] struct {
    // Items is the list of items.
    Items *bubbly.Ref[[]T]
    
    // Length is the item count (computed).
    Length *bubbly.Computed[int]
    
    // IsEmpty indicates if list is empty (computed).
    IsEmpty *bubbly.Computed[bool]
}

// Push adds items to the end.
func (l *ListReturn[T]) Push(items ...T)

// Pop removes and returns the last item.
func (l *ListReturn[T]) Pop() (T, bool)

// Shift removes and returns the first item.
func (l *ListReturn[T]) Shift() (T, bool)

// Unshift adds items to the beginning.
func (l *ListReturn[T]) Unshift(items ...T)

// Insert adds an item at index.
func (l *ListReturn[T]) Insert(index int, item T)

// RemoveAt removes item at index.
func (l *ListReturn[T]) RemoveAt(index int) (T, bool)

// Remove removes first occurrence of item (needs equality).
func (l *ListReturn[T]) Remove(item T, eq func(a, b T) bool) bool

// UpdateAt updates item at index.
func (l *ListReturn[T]) UpdateAt(index int, item T)

// Clear removes all items.
func (l *ListReturn[T]) Clear()

// Get returns item at index (nil if out of bounds).
func (l *ListReturn[T]) Get(index int) (T, bool)

// Set sets the entire list.
func (l *ListReturn[T]) Set(items []T)

// UseList creates a list management composable.
func UseList[T any](ctx *bubbly.Context, initial []T) *ListReturn[T]

// =====================================
// UseMap
// =====================================

// MapReturn is the return value of UseMap.
type MapReturn[K comparable, V any] struct {
    // Data is the map data.
    Data *bubbly.Ref[map[K]V]
    
    // Size is the entry count (computed).
    Size *bubbly.Computed[int]
    
    // IsEmpty indicates if map is empty (computed).
    IsEmpty *bubbly.Computed[bool]
}

// Get returns value for key.
func (m *MapReturn[K, V]) Get(key K) (V, bool)

// Set sets value for key.
func (m *MapReturn[K, V]) Set(key K, value V)

// Delete removes key.
func (m *MapReturn[K, V]) Delete(key K) bool

// Has returns true if key exists.
func (m *MapReturn[K, V]) Has(key K) bool

// Keys returns all keys.
func (m *MapReturn[K, V]) Keys() []K

// Values returns all values.
func (m *MapReturn[K, V]) Values() []V

// Clear removes all entries.
func (m *MapReturn[K, V]) Clear()

// UseMap creates a map management composable.
func UseMap[K comparable, V any](ctx *bubbly.Context, initial map[K]V) *MapReturn[K, V]

// =====================================
// UseSet
// =====================================

// SetReturn is the return value of UseSet.
type SetReturn[T comparable] struct {
    // Values is the set values.
    Values *bubbly.Ref[map[T]struct{}]
    
    // Size is the value count (computed).
    Size *bubbly.Computed[int]
    
    // IsEmpty indicates if set is empty (computed).
    IsEmpty *bubbly.Computed[bool]
}

// Add adds value to set.
func (s *SetReturn[T]) Add(value T)

// Delete removes value from set.
func (s *SetReturn[T]) Delete(value T) bool

// Has returns true if value exists.
func (s *SetReturn[T]) Has(value T) bool

// Toggle adds if not present, removes if present.
func (s *SetReturn[T]) Toggle(value T)

// Clear removes all values.
func (s *SetReturn[T]) Clear()

// ToSlice returns values as slice.
func (s *SetReturn[T]) ToSlice() []T

// UseSet creates a set management composable.
func UseSet[T comparable](ctx *bubbly.Context, initial []T) *SetReturn[T]

// =====================================
// UseQueue
// =====================================

// QueueReturn is the return value of UseQueue.
type QueueReturn[T any] struct {
    // Items is the queue items.
    Items *bubbly.Ref[[]T]
    
    // Size is the item count (computed).
    Size *bubbly.Computed[int]
    
    // IsEmpty indicates if queue is empty (computed).
    IsEmpty *bubbly.Computed[bool]
    
    // Front is the first item (computed).
    Front *bubbly.Computed[*T]
}

// Enqueue adds item to back.
func (q *QueueReturn[T]) Enqueue(item T)

// Dequeue removes and returns front item.
func (q *QueueReturn[T]) Dequeue() (T, bool)

// Peek returns front item without removing.
func (q *QueueReturn[T]) Peek() (T, bool)

// Clear removes all items.
func (q *QueueReturn[T]) Clear()

// UseQueue creates a FIFO queue composable.
func UseQueue[T any](ctx *bubbly.Context, initial []T) *QueueReturn[T]
```

### Development Types

```go
// =====================================
// UseLogger
// =====================================

// LogLevel defines logging levels.
type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelWarn
    LogLevelError
)

// LoggerReturn is the return value of UseLogger.
type LoggerReturn struct {
    // Level is the current log level.
    Level *bubbly.Ref[LogLevel]
    
    // Logs is the log history.
    Logs *bubbly.Ref[[]LogEntry]
    
    componentName string
}

// LogEntry represents a log entry.
type LogEntry struct {
    Time      time.Time
    Level     LogLevel
    Component string
    Message   string
    Data      interface{}
}

// Debug logs at debug level.
func (l *LoggerReturn) Debug(msg string, data ...interface{})

// Info logs at info level.
func (l *LoggerReturn) Info(msg string, data ...interface{})

// Warn logs at warn level.
func (l *LoggerReturn) Warn(msg string, data ...interface{})

// Error logs at error level.
func (l *LoggerReturn) Error(msg string, data ...interface{})

// Clear clears log history.
func (l *LoggerReturn) Clear()

// UseLogger creates a logging composable.
func UseLogger(ctx *bubbly.Context, componentName string) *LoggerReturn

// =====================================
// UseNotification
// =====================================

// NotificationType defines notification types.
type NotificationType string

const (
    NotificationInfo    NotificationType = "info"
    NotificationSuccess NotificationType = "success"
    NotificationWarning NotificationType = "warning"
    NotificationError   NotificationType = "error"
)

// Notification represents a toast notification.
type Notification struct {
    ID        int
    Type      NotificationType
    Title     string
    Message   string
    Duration  time.Duration
    CreatedAt time.Time
}

// NotificationReturn is the return value of UseNotification.
type NotificationReturn struct {
    // Notifications is the active notification stack.
    Notifications *bubbly.Ref[[]Notification]
    
    nextID int
}

// Show displays a notification.
func (n *NotificationReturn) Show(ntype NotificationType, title, message string, duration time.Duration)

// Info shows an info notification.
func (n *NotificationReturn) Info(title, message string)

// Success shows a success notification.
func (n *NotificationReturn) Success(title, message string)

// Warning shows a warning notification.
func (n *NotificationReturn) Warning(title, message string)

// Error shows an error notification.
func (n *NotificationReturn) Error(title, message string)

// Dismiss removes a specific notification.
func (n *NotificationReturn) Dismiss(id int)

// DismissAll removes all notifications.
func (n *NotificationReturn) DismissAll()

// NotificationOption configures UseNotification.
type NotificationOption func(*notificationConfig)

// WithDefaultDuration sets default notification duration.
func WithDefaultDuration(d time.Duration) NotificationOption

// WithMaxNotifications sets maximum stack size.
func WithMaxNotifications(max int) NotificationOption

// UseNotification creates a notification composable.
func UseNotification(ctx *bubbly.Context, opts ...NotificationOption) *NotificationReturn
```

## Data Flow

```
User Input (tea.KeyMsg, tea.WindowSizeMsg)
    │
    ▼
Component.Update()
    │
    ├──► Composable Methods (e.g., focus.Next(), scroll.ScrollDown())
    │        │
    │        ▼
    │    Update Refs (bubbly.Ref[T].Set())
    │        │
    │        ▼
    │    Trigger Watchers / Computed recalculation
    │
    ▼
Component.View()
    │
    ├──► Read Composable State (ref.Get(), computed.Get())
    │
    ▼
Rendered Output
```

## Integration with Existing System

### With CreateShared
All new composables support the shared composable pattern:

```go
// Package-level singleton
var UseSharedFocus = composables.CreateShared(
    func(ctx *bubbly.Context) *FocusReturn[FocusPane] {
        return UseFocus(ctx, FocusInput, []FocusPane{FocusInput, FocusMessages, FocusSidebar})
    },
)

// Usage in any component
focus := localComposables.UseSharedFocus(ctx)
```

### With testutil Harness
All composables work with the testing utilities:

```go
func TestUseFocus(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := createTestComponent()
    ct := harness.Mount(component)
    
    ct.Emit("nextFocus", nil)
    ct.AssertRenderContains("Messages focused")
    
    ct.Unmount()
}
```

## Known Limitations & Solutions

### 1. tea.Tick Integration
**Problem**: UseInterval/UseTimeout need to work with Bubbletea's message-based architecture.

**Solution**: Emit events that the component can handle to trigger tea.Tick commands:
```go
ctx.On("intervalTick", func(_ interface{}) {
    // Return tea.Tick command from component
})
```

### 2. Computed Dependency Tracking
**Problem**: Complex computed values with multiple dependencies.

**Solution**: Use the existing Computed system which auto-tracks dependencies.

### 3. Memory Management
**Problem**: Composables with internal state need cleanup.

**Solution**: Use OnUnmounted hook to clean up timers, cancel pending operations.

## Framework-Level Window Resize Integration (Phase 6)

### Overview
The framework automatically handles `tea.WindowSizeMsg` and emits events so that composables like `UseWindowSize` work without any Bubbletea boilerplate.

### Component Update Enhancement
```go
// In componentImpl.Update(), BEFORE messageHandler:
func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Auto-handle WindowSizeMsg - emit "windowResize" event (zero-boilerplate)
    if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
        c.Emit("windowResize", map[string]int{
            "width":  wsMsg.Width,
            "height": wsMsg.Height,
        })
    }
    
    // ... rest of existing Update logic
}
```

### UseWindowSize Auto-Subscribe
```go
func UseWindowSize(ctx *bubbly.Context, opts ...WindowSizeOption) *WindowSizeReturn {
    // ... existing setup code ...
    
    // Auto-subscribe to framework's "windowResize" event
    if ctx != nil {
        ctx.On("windowResize", func(data interface{}) {
            if sizeData, ok := data.(map[string]int); ok {
                ws.SetSize(sizeData["width"], sizeData["height"])
            }
        })
    }
    
    return ws
}
```

### Event Data Format
```go
// The "windowResize" event data is always:
map[string]int{
    "width":  int,  // Terminal width in columns
    "height": int,  // Terminal height in rows
}
```

### Backward Compatibility
- Event fires BEFORE messageHandler, allowing existing code to work
- Users can still use WithMessageHandler for custom handling
- No breaking changes to existing APIs

## Future Enhancements

1. **UseMedia** - Media query support for terminal capabilities
2. **UseClipboard** - Terminal clipboard integration (OS-specific)
3. **UseAnimation** - Frame-based animation utilities
4. **UseVirtualList** - Virtual scrolling for large lists
5. **UseKeyBindingState** - Dynamic key binding management
