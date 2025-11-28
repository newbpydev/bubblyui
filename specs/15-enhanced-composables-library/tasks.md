# Implementation Tasks: Enhanced Composables Library

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] 01-reactivity-system completed (Ref, Computed, Watch)
- [x] 04-composition-api completed (composables foundation)
- [x] 14-advanced-layout-system completed (responsive examples)
- [ ] Review existing composables API for consistency

---

## Phase 1: TUI-Specific Core Composables

### Task 1.1: UseWindowSize ✅ COMPLETED
- **Description**: Terminal dimensions and responsive breakpoints composable
- **Prerequisites**: None
- **Unlocks**: Task 1.2, Phase 2 tasks
- **Files**:
  - `pkg/bubbly/composables/use_window_size.go`
  - `pkg/bubbly/composables/use_window_size_test.go`
- **Type Safety**:
  ```go
  type Breakpoint string
  type WindowSizeReturn struct {
      Width, Height *bubbly.Ref[int]
      Breakpoint *bubbly.Ref[Breakpoint]
      SidebarVisible *bubbly.Ref[bool]
      GridColumns *bubbly.Ref[int]
  }
  func UseWindowSize(ctx *bubbly.Context, opts ...WindowSizeOption) *WindowSizeReturn
  ```
- **Tests**:
  - [x] Breakpoint calculation for all ranges (xs, sm, md, lg, xl)
  - [x] SetSize updates all derived values
  - [x] Min dimension enforcement
  - [x] Custom breakpoint configuration
  - [x] GetContentWidth with/without sidebar
  - [x] GetCardWidth calculation
- **Estimated effort**: 4 hours
- **Implementation Notes**:
  - 13 test functions with 30+ sub-tests covering all breakpoint ranges
  - 100% coverage on all functions except GetCardWidth (80% - defensive edge case)
  - Includes: WithBreakpoints, WithMinDimensions, WithSidebarWidth options
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Default 80x24 terminal size with MD breakpoint

### Task 1.2: UseFocus ✅ COMPLETED
- **Description**: Multi-pane focus management with generic type support
- **Prerequisites**: Task 1.1
- **Unlocks**: Task 1.3
- **Files**:
  - `pkg/bubbly/composables/use_focus.go`
  - `pkg/bubbly/composables/use_focus_test.go`
- **Type Safety**:
  ```go
  type FocusReturn[T comparable] struct {
      Current *bubbly.Ref[T]
  }
  func UseFocus[T comparable](ctx *bubbly.Context, initial T, order []T) *FocusReturn[T]
  ```
- **Tests**:
  - [x] Initial focus set correctly
  - [x] Next() cycles through order
  - [x] Previous() cycles backward
  - [x] Focus() sets specific pane
  - [x] IsFocused() returns correct value
  - [x] Empty order panics
  - [x] Single item order stays focused
- **Estimated effort**: 3 hours
- **Implementation Notes**:
  - 13 test functions covering all requirements and edge cases
  - 88.9% coverage on composables package (above 80% requirement)
  - Generic type support verified with int (FocusPane) and string types
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Focus on non-existent pane is no-op (safe behavior)
  - Initial value not in order defaults to first item in order
  - Comprehensive godoc comments with examples

### Task 1.3: UseScroll ✅ COMPLETED
- **Description**: Viewport scrolling management
- **Prerequisites**: Task 1.2
- **Unlocks**: Task 1.4
- **Files**:
  - `pkg/bubbly/composables/use_scroll.go`
  - `pkg/bubbly/composables/use_scroll_test.go`
- **Type Safety**:
  ```go
  type ScrollReturn struct {
      Offset, MaxOffset, VisibleCount, TotalItems *bubbly.Ref[int]
  }
  func UseScroll(ctx *bubbly.Context, totalItems, visibleCount int) *ScrollReturn
  ```
- **Tests**:
  - [x] Initial offset is 0
  - [x] ScrollUp/ScrollDown respect bounds
  - [x] ScrollTo clamps to valid range
  - [x] ScrollToTop/ScrollToBottom work correctly
  - [x] PageUp/PageDown move by visible count
  - [x] IsAtTop/IsAtBottom return correct values
  - [x] SetTotalItems recalculates max offset
  - [x] Empty list (0 items) handled
- **Estimated effort**: 3 hours
- **Implementation Notes**:
  - 14 test functions with 20+ sub-tests covering all requirements and edge cases
  - 86.7% coverage on UseScroll function, 100% on most methods
  - Includes: ScrollUp, ScrollDown, ScrollTo, ScrollToTop, ScrollToBottom, PageUp, PageDown
  - IsAtTop/IsAtBottom for boundary detection
  - SetTotalItems/SetVisibleCount for dynamic list updates
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Handles edge cases: empty list, visible >= total, negative inputs
  - Comprehensive godoc comments with examples

### Task 1.4: UseSelection ✅ COMPLETED
- **Description**: List/table selection management with generic items
- **Prerequisites**: Task 1.3
- **Unlocks**: Task 1.5
- **Files**:
  - `pkg/bubbly/composables/use_selection.go`
  - `pkg/bubbly/composables/use_selection_test.go`
- **Type Safety**:
  ```go
  type SelectionReturn[T any] struct {
      SelectedIndex *bubbly.Ref[int]
      SelectedItem *bubbly.Computed[T]
      SelectedIndices *bubbly.Ref[[]int]  // for multi-select
      Items *bubbly.Ref[[]T]
  }
  func UseSelection[T any](ctx *bubbly.Context, items []T, opts ...SelectionOption) *SelectionReturn[T]
  ```
- **Tests**:
  - [x] Initial selection is 0
  - [x] SelectNext/SelectPrevious navigate
  - [x] Wrap option enables circular navigation
  - [x] SelectedItem computed correctly
  - [x] IsSelected returns correct value
  - [x] SetItems updates and adjusts selection
  - [x] Empty items list handled
  - [x] Multi-select mode (optional)
- **Estimated effort**: 4 hours
- **Implementation Notes**:
  - 20 test functions with 30+ sub-tests covering all requirements and edge cases
  - 90.4% coverage on composables package (above 80% requirement)
  - Generic type support verified with string and custom types
  - Includes: WithWrap, WithMultiSelect options
  - Methods: Select, SelectNext, SelectPrevious, IsSelected, ToggleSelection, ClearSelection, SetItems
  - Empty list returns -1 for SelectedIndex, zero value for SelectedItem
  - Multi-select mode uses SelectedIndices for tracking multiple selections
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 1.5: UseMode ✅ COMPLETED
- **Description**: Navigation/input mode management
- **Prerequisites**: Task 1.4
- **Unlocks**: Phase 2
- **Files**:
  - `pkg/bubbly/composables/use_mode.go`
  - `pkg/bubbly/composables/use_mode_test.go`
- **Type Safety**:
  ```go
  type ModeReturn[T comparable] struct {
      Current, Previous *bubbly.Ref[T]
  }
  func UseMode[T comparable](ctx *bubbly.Context, initial T) *ModeReturn[T]
  ```
- **Tests**:
  - [x] Initial mode set correctly
  - [x] Switch changes mode and updates previous
  - [x] Toggle alternates between two modes
  - [x] IsMode returns correct value
  - [x] Previous tracks correctly on multiple switches
- **Estimated effort**: 2 hours
- **Implementation Notes**:
  - 11 test functions with 15+ sub-tests covering all requirements and edge cases
  - 100% coverage on UseMode function and all methods
  - 90.8% coverage on composables package (above 80% requirement)
  - Generic type support verified with string (Mode) and int types
  - Methods: Switch, Toggle, IsMode
  - Switch to same mode is no-op (doesn't update Previous)
  - Toggle from different mode switches to first option (a)
  - Previous initially set to same value as initial (no previous yet)
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

---

## Phase 2: State Utility Composables

### Task 2.1: UseToggle ✅ COMPLETED
- **Description**: Boolean toggle state management
- **Prerequisites**: Phase 1 completed
- **Unlocks**: Task 2.2
- **Files**:
  - `pkg/bubbly/composables/use_toggle.go`
  - `pkg/bubbly/composables/use_toggle_test.go`
- **Type Safety**:
  ```go
  type ToggleReturn struct {
      Value *bubbly.Ref[bool]
  }
  func UseToggle(ctx *bubbly.Context, initial bool) *ToggleReturn
  ```
- **Tests**:
  - [x] Initial value set correctly
  - [x] Toggle flips value
  - [x] Set sets explicit value
  - [x] On sets to true
  - [x] Off sets to false
- **Estimated effort**: 1 hour
- **Implementation Notes**:
  - 9 test functions with 15+ sub-tests covering all requirements and edge cases
  - 100% coverage on UseToggle function and all methods
  - 90.9% coverage on composables package (above 80% requirement)
  - Methods: Toggle, Set, On, Off
  - Value is reactive (tested with Watch)
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 2.2: UseCounter ✅ COMPLETED
- **Description**: Bounded counter with step support
- **Prerequisites**: Task 2.1
- **Unlocks**: Task 2.3
- **Files**:
  - `pkg/bubbly/composables/use_counter.go`
  - `pkg/bubbly/composables/use_counter_test.go`
- **Type Safety**:
  ```go
  type CounterReturn struct {
      Count *bubbly.Ref[int]
  }
  func UseCounter(ctx *bubbly.Context, initial int, opts ...CounterOption) *CounterReturn
  ```
- **Tests**:
  - [x] Initial value set correctly
  - [x] Increment/Decrement by step
  - [x] Min bound enforced
  - [x] Max bound enforced
  - [x] IncrementBy/DecrementBy work
  - [x] Set clamps to bounds
  - [x] Reset returns to initial
  - [x] Default step is 1
- **Estimated effort**: 2 hours
- **Implementation Notes**:
  - 18 test functions with 40+ sub-tests covering all requirements and edge cases
  - 100% coverage on UseCounter function and all methods
  - 91.5% coverage on composables package (above 80% requirement)
  - Options: WithMin, WithMax, WithStep (functional options pattern)
  - Methods: Increment, Decrement, IncrementBy, DecrementBy, Set, Reset
  - Initial value clamped to bounds if min/max configured
  - Default step is 1, configurable via WithStep
  - Works with CreateShared pattern for cross-component sharing
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 2.3: UsePrevious ✅ COMPLETED
- **Description**: Previous value tracking
- **Prerequisites**: Task 2.2
- **Unlocks**: Task 2.4
- **Files**:
  - `pkg/bubbly/composables/use_previous.go`
  - `pkg/bubbly/composables/use_previous_test.go`
- **Type Safety**:
  ```go
  type PreviousReturn[T any] struct {
      Value *bubbly.Ref[*T]
  }
  func UsePrevious[T any](ctx *bubbly.Context, ref *bubbly.Ref[T]) *PreviousReturn[T]
  ```
- **Tests**:
  - [x] Initial previous is nil
  - [x] Previous updates when ref changes
  - [x] Get returns correct previous value
  - [x] Works with Watch internally
- **Estimated effort**: 2 hours
- **Implementation Notes**:
  - 11 test functions with 20+ sub-tests covering all requirements and edge cases
  - 100% coverage on UsePrevious function and Get method
  - 91.6% coverage on composables package (above 80% requirement)
  - Generic type support verified with int, string, struct, and slice types
  - Uses *T (pointer) to distinguish "no previous" (nil) from "previous is zero value"
  - Uses bubbly.Watch internally to observe source ref changes
  - Stores copy of oldVal to avoid aliasing issues
  - Works with CreateShared pattern for cross-component sharing
  - Value is reactive (can be watched for changes)
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 2.4: UseHistory ✅ COMPLETED
- **Description**: Undo/redo state management
- **Prerequisites**: Task 2.3
- **Unlocks**: Phase 3
- **Files**:
  - `pkg/bubbly/composables/use_history.go`
  - `pkg/bubbly/composables/use_history_test.go`
- **Type Safety**:
  ```go
  type HistoryReturn[T any] struct {
      Current *bubbly.Ref[T]
      CanUndo, CanRedo *bubbly.Computed[bool]
  }
  func UseHistory[T any](ctx *bubbly.Context, initial T, maxSize int) *HistoryReturn[T]
  ```
- **Tests**:
  - [x] Initial state set correctly
  - [x] Push adds to history
  - [x] Undo reverts state
  - [x] Redo restores state
  - [x] CanUndo/CanRedo computed correctly
  - [x] Push clears redo stack
  - [x] Max size enforced (drop oldest)
  - [x] Clear empties history
- **Estimated effort**: 4 hours
- **Implementation Notes**:
  - 17 test functions with 25+ sub-tests covering all requirements and edge cases
  - 92.1% coverage on composables package (above 80% requirement)
  - Generic type support verified with int, string, struct, and slice types
  - Uses two stacks: past (for undo) and future (for redo)
  - maxSize represents max entries in past stack (allowing maxSize undos)
  - Uses internal Ref[int] for pastLen/futureLen to enable reactive Computed values
  - CanUndo/CanRedo are Computed[bool] that react to history changes
  - Thread-safe with sync.Mutex protecting past/future slices
  - Works with CreateShared pattern for cross-component sharing
  - Current is reactive (can be watched for changes)
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

---

## Phase 3: Timing Composables

### Task 3.1: UseInterval ✅ COMPLETED
- **Description**: Periodic execution using tea.Tick pattern
- **Prerequisites**: Phase 2 completed
- **Unlocks**: Task 3.2
- **Files**:
  - `pkg/bubbly/composables/use_interval.go`
  - `pkg/bubbly/composables/use_interval_test.go`
- **Type Safety**:
  ```go
  type IntervalReturn struct {
      IsRunning *bubbly.Ref[bool]
  }
  func UseInterval(ctx *bubbly.Context, callback func(), duration time.Duration) *IntervalReturn
  ```
- **Tests**:
  - [x] Starts in stopped state
  - [x] Start begins interval
  - [x] Stop pauses interval
  - [x] Toggle flips state
  - [x] Reset restarts
  - [x] Callback executed on tick
  - [x] Cleanup on unmount
  - [x] Negative duration errors
- **Estimated effort**: 3 hours
- **Implementation Notes**:
  - 16 test functions covering all requirements and edge cases
  - 100% coverage on UseInterval function and all methods
  - 92.6% coverage on composables package (above 80% requirement)
  - Uses internal time.Ticker with goroutine for timing (simpler than tea.Tick integration)
  - Thread-safe with sync.Mutex protecting internal state
  - Methods: Start, Stop, Toggle, Reset
  - Start/Stop are idempotent (multiple calls are no-op)
  - Callback captured in local variable to avoid race conditions
  - Works with CreateShared pattern for cross-component sharing
  - Automatic cleanup via OnUnmounted hook
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 3.2: UseTimeout ✅ COMPLETED
- **Description**: Delayed execution with cancel support
- **Prerequisites**: Task 3.1
- **Unlocks**: Task 3.3
- **Files**:
  - `pkg/bubbly/composables/use_timeout.go`
  - `pkg/bubbly/composables/use_timeout_test.go`
- **Type Safety**:
  ```go
  type TimeoutReturn struct {
      IsPending, IsExpired *bubbly.Ref[bool]
  }
  func UseTimeout(ctx *bubbly.Context, callback func(), duration time.Duration) *TimeoutReturn
  ```
- **Tests**:
  - [x] Starts not pending
  - [x] Start begins timeout
  - [x] Cancel stops pending timeout
  - [x] Reset cancels and restarts
  - [x] Callback executed on expiry
  - [x] IsExpired set after execution
  - [x] Cleanup on unmount
- **Estimated effort**: 3 hours
- **Implementation Notes**:
  - 18 test functions covering all requirements and edge cases
  - 92.7% coverage on composables package (above 80% requirement)
  - Uses time.AfterFunc for efficient one-shot timing (simpler than time.Timer)
  - Thread-safe with sync.Mutex protecting internal state
  - Methods: Start, Cancel, Reset
  - Start/Cancel are idempotent (multiple calls are no-op)
  - Start after expiry resets IsExpired and starts new timeout
  - IsPending and IsExpired are reactive (can be watched for changes)
  - Works with CreateShared pattern for cross-component sharing
  - Automatic cleanup via OnUnmounted hook
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 3.3: UseTimer ✅ COMPLETED
- **Description**: Countdown timer with progress tracking
- **Prerequisites**: Task 3.2
- **Unlocks**: Phase 4
- **Files**:
  - `pkg/bubbly/composables/use_timer.go`
  - `pkg/bubbly/composables/use_timer_test.go`
- **Type Safety**:
  ```go
  type TimerReturn struct {
      Remaining *bubbly.Ref[time.Duration]
      IsRunning *bubbly.Ref[bool]
      IsExpired, Progress *bubbly.Computed[...]
  }
  func UseTimer(ctx *bubbly.Context, duration time.Duration, opts ...TimerOption) *TimerReturn
  ```
- **Tests**:
  - [x] Initial remaining equals duration
  - [x] Start begins countdown
  - [x] Stop pauses countdown
  - [x] Reset restarts from full duration
  - [x] IsExpired true when remaining <= 0
  - [x] Progress calculated correctly
  - [x] OnExpire callback executed
  - [x] Tick interval configurable
- **Estimated effort**: 4 hours
- **Implementation Notes**:
  - 20 test functions covering all requirements and edge cases
  - 92.7% coverage on composables package (above 80% requirement)
  - Uses time.Ticker with goroutine for countdown (similar to UseInterval)
  - Thread-safe with sync.Mutex protecting internal state
  - Options: WithOnExpire, WithTickInterval (default 100ms)
  - Methods: Start, Stop, Reset
  - IsExpired is Computed[bool] that reacts to Remaining <= 0
  - Progress is Computed[float64] calculated as 1.0 - (Remaining / InitialDuration)
  - Start/Stop are idempotent (multiple calls are no-op)
  - Timer stops automatically when expired and calls OnExpire callback
  - Reset() stops timer and restores Remaining to initial duration
  - Works with CreateShared pattern for cross-component sharing
  - Automatic cleanup via OnUnmounted hook
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

---

## Phase 4: Collection Composables

### Task 4.1: UseList ✅ COMPLETED
- **Description**: Generic list CRUD operations
- **Prerequisites**: Phase 3 completed
- **Unlocks**: Task 4.2
- **Files**:
  - `pkg/bubbly/composables/use_list.go`
  - `pkg/bubbly/composables/use_list_test.go`
- **Type Safety**:
  ```go
  type ListReturn[T any] struct {
      Items *bubbly.Ref[[]T]
      Length *bubbly.Computed[int]
      IsEmpty *bubbly.Computed[bool]
  }
  func UseList[T any](ctx *bubbly.Context, initial []T) *ListReturn[T]
  ```
- **Tests**:
  - [x] Initial items set correctly
  - [x] Push adds to end
  - [x] Pop removes from end
  - [x] Shift removes from start
  - [x] Unshift adds to start
  - [x] Insert at index
  - [x] RemoveAt removes by index
  - [x] UpdateAt updates by index
  - [x] Clear empties list
  - [x] Length/IsEmpty computed
  - [x] Out of bounds handled
- **Estimated effort**: 4 hours
- **Implementation Notes**:
  - 20 test functions with 50+ sub-tests covering all requirements and edge cases
  - 100% coverage on UseList function and all methods
  - 93.5% coverage on composables package (above 80% requirement)
  - Generic type support verified with int, string, and struct types
  - Methods: Push, Pop, Shift, Unshift, Insert, RemoveAt, UpdateAt, Remove, Clear, Get, Set
  - Thread-safe with sync.Mutex protecting internal state
  - Nil initial slice normalized to empty slice
  - Out of bounds handling: Insert clamps to valid range, Get/RemoveAt/UpdateAt return false
  - Works with CreateShared pattern for cross-component sharing
  - Items ref is reactive (can be watched for changes)
  - Length and IsEmpty are Computed values that auto-update
  - Race detector passes, no goroutine leaks
  - Comprehensive godoc comments with examples

### Task 4.2: UseMap
- **Description**: Generic key-value state management
- **Prerequisites**: Task 4.1
- **Unlocks**: Task 4.3
- **Files**:
  - `pkg/bubbly/composables/use_map.go`
  - `pkg/bubbly/composables/use_map_test.go`
- **Type Safety**:
  ```go
  type MapReturn[K comparable, V any] struct {
      Data *bubbly.Ref[map[K]V]
      Size *bubbly.Computed[int]
      IsEmpty *bubbly.Computed[bool]
  }
  func UseMap[K comparable, V any](ctx *bubbly.Context, initial map[K]V) *MapReturn[K, V]
  ```
- **Tests**:
  - [ ] Initial data set correctly
  - [ ] Get returns value
  - [ ] Set adds/updates key
  - [ ] Delete removes key
  - [ ] Has checks existence
  - [ ] Keys returns all keys
  - [ ] Values returns all values
  - [ ] Clear empties map
  - [ ] Size/IsEmpty computed
- **Estimated effort**: 3 hours

### Task 4.3: UseSet
- **Description**: Unique value set management
- **Prerequisites**: Task 4.2
- **Unlocks**: Task 4.4
- **Files**:
  - `pkg/bubbly/composables/use_set.go`
  - `pkg/bubbly/composables/use_set_test.go`
- **Type Safety**:
  ```go
  type SetReturn[T comparable] struct {
      Values *bubbly.Ref[map[T]struct{}]
      Size *bubbly.Computed[int]
      IsEmpty *bubbly.Computed[bool]
  }
  func UseSet[T comparable](ctx *bubbly.Context, initial []T) *SetReturn[T]
  ```
- **Tests**:
  - [ ] Initial values set correctly
  - [ ] Add adds value
  - [ ] Delete removes value
  - [ ] Has checks existence
  - [ ] Toggle adds/removes
  - [ ] Clear empties set
  - [ ] ToSlice returns values
  - [ ] Size/IsEmpty computed
  - [ ] Duplicates ignored in initial
- **Estimated effort**: 3 hours

### Task 4.4: UseQueue
- **Description**: FIFO queue operations
- **Prerequisites**: Task 4.3
- **Unlocks**: Phase 5
- **Files**:
  - `pkg/bubbly/composables/use_queue.go`
  - `pkg/bubbly/composables/use_queue_test.go`
- **Type Safety**:
  ```go
  type QueueReturn[T any] struct {
      Items *bubbly.Ref[[]T]
      Size *bubbly.Computed[int]
      IsEmpty *bubbly.Computed[bool]
      Front *bubbly.Computed[*T]
  }
  func UseQueue[T any](ctx *bubbly.Context, initial []T) *QueueReturn[T]
  ```
- **Tests**:
  - [ ] Initial items set correctly
  - [ ] Enqueue adds to back
  - [ ] Dequeue removes from front
  - [ ] Peek returns front without removing
  - [ ] Clear empties queue
  - [ ] Front computed correctly
  - [ ] Size/IsEmpty computed
  - [ ] Dequeue on empty returns false
- **Estimated effort**: 2 hours

---

## Phase 5: Development & Documentation

### Task 5.1: UseLogger
- **Description**: Component debug logging with levels
- **Prerequisites**: Phase 4 completed
- **Unlocks**: Task 5.2
- **Files**:
  - `pkg/bubbly/composables/use_logger.go`
  - `pkg/bubbly/composables/use_logger_test.go`
- **Type Safety**:
  ```go
  type LogLevel int
  type LoggerReturn struct {
      Level *bubbly.Ref[LogLevel]
      Logs *bubbly.Ref[[]LogEntry]
  }
  func UseLogger(ctx *bubbly.Context, componentName string) *LoggerReturn
  ```
- **Tests**:
  - [ ] Debug/Info/Warn/Error log at correct levels
  - [ ] Log entries include timestamp, component, message
  - [ ] Level filtering works
  - [ ] Clear removes all logs
  - [ ] Data attached to entries
- **Estimated effort**: 2 hours

### Task 5.2: UseNotification
- **Description**: Toast notification system
- **Prerequisites**: Task 5.1
- **Unlocks**: Task 5.3
- **Files**:
  - `pkg/bubbly/composables/use_notification.go`
  - `pkg/bubbly/composables/use_notification_test.go`
- **Type Safety**:
  ```go
  type NotificationType string
  type NotificationReturn struct {
      Notifications *bubbly.Ref[[]Notification]
  }
  func UseNotification(ctx *bubbly.Context, opts ...NotificationOption) *NotificationReturn
  ```
- **Tests**:
  - [ ] Show adds notification
  - [ ] Info/Success/Warning/Error convenience methods
  - [ ] Dismiss removes by ID
  - [ ] DismissAll clears all
  - [ ] Duration configurable
  - [ ] Max notifications enforced
  - [ ] Auto-dismiss after duration
- **Estimated effort**: 3 hours

### Task 5.3: Documentation & Examples
- **Description**: Update manuals and create example app
- **Prerequisites**: All composables completed
- **Unlocks**: Feature complete
- **Files**:
  - `docs/BUBBLY_AI_MANUAL_SYSTEMATIC.md` (update composables section)
  - `docs/BUBBLY_AI_MANUAL_COMPACT.md` (update composables section)
  - `pkg/bubbly/composables/README.md` (update with new composables)
  - `cmd/examples/17-enhanced-composables/` (example app)
- **Tasks**:
  - [ ] Add all new composables to systematic manual
  - [ ] Add all new composables to compact manual
  - [ ] Update composables README with full list
  - [ ] Create comprehensive example app demonstrating all composables
  - [ ] Add godoc examples for each composable
- **Estimated effort**: 6 hours

### Task 5.4: Integration Tests & Benchmarks
- **Description**: Integration tests and performance validation
- **Prerequisites**: Task 5.3
- **Unlocks**: Feature ready for release
- **Files**:
  - `tests/integration/composables_enhanced_test.go`
  - `pkg/bubbly/composables/enhanced_bench_test.go`
- **Tests**:
  - [ ] All composables work with testutil harness
  - [ ] All composables work with CreateShared
  - [ ] Composables compose together correctly
  - [ ] Performance: <100ns initialization
  - [ ] Memory: No goroutine leaks
  - [ ] Race detector passes
- **Estimated effort**: 4 hours

---

## Task Dependency Graph

```
Prerequisites (01, 04, 14 completed)
    │
    ▼
Phase 1: TUI-Specific
    ├─► 1.1 UseWindowSize
    │       ↓
    ├─► 1.2 UseFocus
    │       ↓
    ├─► 1.3 UseScroll
    │       ↓
    ├─► 1.4 UseSelection
    │       ↓
    └─► 1.5 UseMode
            ↓
Phase 2: State Utilities
    ├─► 2.1 UseToggle
    │       ↓
    ├─► 2.2 UseCounter
    │       ↓
    ├─► 2.3 UsePrevious
    │       ↓
    └─► 2.4 UseHistory
            ↓
Phase 3: Timing
    ├─► 3.1 UseInterval
    │       ↓
    ├─► 3.2 UseTimeout
    │       ↓
    └─► 3.3 UseTimer
            ↓
Phase 4: Collections
    ├─► 4.1 UseList
    │       ↓
    ├─► 4.2 UseMap
    │       ↓
    ├─► 4.3 UseSet
    │       ↓
    └─► 4.4 UseQueue
            ↓
Phase 5: Development & Docs
    ├─► 5.1 UseLogger
    │       ↓
    ├─► 5.2 UseNotification
    │       ↓
    ├─► 5.3 Documentation
    │       ↓
    └─► 5.4 Integration Tests
            ↓
        FEATURE COMPLETE
```

---

## Validation Checklist

- [ ] All types are strictly defined with generics
- [ ] All composables have tests (80%+ coverage)
- [ ] No orphaned composables (all documented and in example)
- [ ] TDD followed (tests written first)
- [ ] Race detector passes on all tests
- [ ] Performance benchmarks <100ns init
- [ ] Code conventions followed (golangci-lint clean)
- [ ] Godoc comments on all exports
- [ ] Manuals updated with all composables
- [ ] Example app demonstrates all composables

---

## Estimated Total Effort

| Phase | Tasks | Hours |
|-------|-------|-------|
| Phase 1 | 5 | 16 |
| Phase 2 | 4 | 9 |
| Phase 3 | 3 | 10 |
| Phase 4 | 4 | 12 |
| Phase 5 | 4 | 15 |
| **Total** | **20** | **62** |

---

## Notes

### Implementation Guidelines
1. Follow existing composables pattern in `pkg/bubbly/composables/`
2. Use functional options pattern for configuration
3. All composables must work with `CreateShared`
4. Use `bubbly.Computed` for derived values
5. Clean up resources in `OnUnmounted` if needed

### Testing Guidelines
1. Use table-driven tests
2. Test edge cases (empty, bounds, nil)
3. Test concurrent access with `-race`
4. Use testutil harness for integration tests
5. Benchmark initialization time

### Documentation Guidelines
1. Add godoc comments with examples
2. Update both systematic and compact manuals
3. Add to composables README
4. Create comprehensive example app
