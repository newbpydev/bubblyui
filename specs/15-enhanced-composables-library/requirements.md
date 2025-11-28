# Feature 15: Enhanced Composables Library

## Feature ID
15-enhanced-composables-library

## Overview
A comprehensive set of production-ready composables that provide common patterns for TUI development, improving developer experience and reducing boilerplate code. This feature expands BubblyUI's composables from 12 to 30+, covering TUI-specific needs, state utilities, collections, and development tools.

## User Stories

### TUI-Specific
- As a developer, I want to use `UseWindowSize` so that I can build responsive TUI applications that adapt to terminal dimensions
- As a developer, I want to use `UseFocus` so that I can manage focus between multiple panes without manual state tracking
- As a developer, I want to use `UseScroll` so that I can implement scrollable viewports with proper boundary handling
- As a developer, I want to use `UseSelection` so that I can manage list/table selection with keyboard navigation
- As a developer, I want to use `UseMode` so that I can implement navigation/input mode patterns like vim

### State Utilities
- As a developer, I want to use `UseToggle` so that I can manage boolean states with a simple API
- As a developer, I want to use `UseCounter` so that I can implement bounded counters with step support
- As a developer, I want to use `UsePrevious` so that I can track the previous value of any state
- As a developer, I want to use `UseHistory` so that I can implement undo/redo functionality

### Timing
- As a developer, I want to use `UseInterval` so that I can schedule periodic updates using tea.Tick pattern
- As a developer, I want to use `UseTimeout` so that I can schedule delayed actions
- As a developer, I want to use `UseTimer` so that I can create countdown timers for animations

### Collections
- As a developer, I want to use `UseList` so that I can manage list state with CRUD operations
- As a developer, I want to use `UseMap` so that I can manage key-value state reactively
- As a developer, I want to use `UseSet` so that I can manage unique value sets reactively
- As a developer, I want to use `UseQueue` so that I can manage FIFO data structures

### Development
- As a developer, I want to use `UseLogger` so that I can debug component state and lifecycle
- As a developer, I want to use `UseNotification` so that I can display toast notifications

## Functional Requirements

### 1. TUI-Specific Composables

#### 1.1 UseWindowSize
- Track terminal width and height reactively
- Automatically calculate responsive breakpoints (xs, sm, md, lg, xl)
- Provide sidebar visibility based on breakpoint
- Calculate optimal grid columns for current width
- Handle minimum dimension enforcement
- Support custom breakpoint configuration
- **Zero Bubbletea Boilerplate**: Automatically receive window resize events without requiring `WithMessageHandler` or manual `tea.WindowSizeMsg` handling (Phase 6 enhancement)

#### 1.2 UseFocus
- Track current focused pane with generic type support
- Cycle focus through defined order (Next/Previous)
- Check if specific pane is focused
- Support dynamic focus order changes
- Emit focus change events

#### 1.3 UseScroll
- Track scroll offset for viewport
- ScrollUp/ScrollDown with bounds checking
- ScrollTo specific index
- ScrollToTop/ScrollToBottom convenience methods
- Track if at top/bottom
- Support page up/page down (by visible count)

#### 1.4 UseSelection
- Track selected index in list/table
- Get selected item via computed value
- SelectNext/SelectPrevious with wrap option
- Check if specific index is selected
- Support multi-selection mode (optional)

#### 1.5 UseMode
- Track current mode with generic type
- Track previous mode for transitions
- Switch to specific mode
- Toggle between two modes
- Check if in specific mode

### 2. State Utility Composables

#### 2.1 UseToggle
- Boolean value management
- Toggle() flips value
- Set(bool), On(), Off() explicit control
- Initial value support

#### 2.2 UseCounter
- Counter with increment/decrement
- Support step size (default: 1)
- Optional min/max bounds
- Reset to initial value
- IncrementBy/DecrementBy with custom amounts

#### 2.3 UsePrevious
- Track previous value of any ref
- Update on value change via Watch
- Generic type support
- Initial undefined handling

#### 2.4 UseHistory
- Push new states to history
- Undo/Redo navigation
- CanUndo/CanRedo computed values
- Configurable max history size
- Clear history
- Generic type support

### 3. Timing Composables

#### 3.1 UseInterval
- Execute callback at regular intervals
- Start/Stop/Toggle control
- Reset to restart interval
- IsRunning state tracking
- Uses tea.Tick pattern internally

#### 3.2 UseTimeout
- Execute callback after delay
- Start/Cancel/Reset control
- IsPending state tracking
- Auto-cleanup on unmount

#### 3.3 UseTimer
- Countdown timer with remaining time ref
- IsExpired computed value
- Start/Stop/Reset control
- Optional onExpire callback
- Progress percentage computed

### 4. Collection Composables

#### 4.1 UseList
- Generic list management
- Push/Pop/Shift/Unshift operations
- Insert at index
- Remove by index/value
- Clear all items
- Update at index
- Find/Filter/Map utilities

#### 4.2 UseMap
- Key-value state management
- Get/Set/Delete operations
- Has key check
- Keys/Values/Entries iteration
- Clear all entries
- Size tracking
- Generic key/value types

#### 4.3 UseSet
- Unique value management
- Add/Delete/Has operations
- Clear all values
- Toggle (add if not present, remove if present)
- Size tracking
- Generic value type

#### 4.4 UseQueue
- FIFO queue operations
- Enqueue/Dequeue
- Peek front item
- Size/IsEmpty tracking
- Clear queue
- Generic item type

### 5. Development Composables

#### 5.1 UseLogger
- Log component lifecycle events
- Log state changes with context
- Configurable log levels
- Component name in log prefix
- Integrate with devtools

#### 5.2 UseNotification
- Show toast notifications
- Support types: info, success, warning, error
- Configurable duration
- Dismiss programmatically
- Stack multiple notifications
- Position configuration

## Non-Functional Requirements

### Performance
- All composables must have <100ns initialization time
- Memory-efficient (no goroutine leaks)
- Proper cleanup on unmount

### Type Safety
- All composables must use Go generics where applicable
- No type assertions in public APIs
- Strict TypeScript-style typing patterns

### Testing
- Minimum 80% test coverage per composable
- Table-driven tests for all functions
- Race condition testing with -race flag
- Integration with testutil harness

### Documentation
- Complete godoc comments on all exports
- Usage examples in doc comments
- README with composable overview

## Acceptance Criteria

### Phase 1: TUI-Specific
- [ ] UseWindowSize implemented with full breakpoint system
- [ ] UseFocus implemented with generic pane type
- [ ] UseScroll implemented with all scroll methods
- [ ] UseSelection implemented with multi-select option
- [ ] UseMode implemented with toggle support

### Phase 2: State Utilities
- [ ] UseToggle implemented with full API
- [ ] UseCounter implemented with bounds and step
- [ ] UsePrevious implemented with generic support
- [ ] UseHistory implemented with undo/redo

### Phase 3: Timing
- [ ] UseInterval implemented with tea.Tick pattern
- [ ] UseTimeout implemented with cleanup
- [ ] UseTimer implemented with countdown

### Phase 4: Collections
- [ ] UseList implemented with full CRUD
- [ ] UseMap implemented with all operations
- [ ] UseSet implemented with toggle
- [ ] UseQueue implemented with FIFO

### Phase 5: Development
- [ ] UseLogger implemented with levels
- [ ] UseNotification implemented with stacking
- [ ] All composables documented in manuals
- [ ] Example app demonstrating all composables

### Phase 6: Framework-Level Window Resize (Zero Bubbletea)
- [ ] Framework automatically emits "windowResize" events on tea.WindowSizeMsg
- [ ] UseWindowSize auto-subscribes to "windowResize" events
- [ ] No manual WithMessageHandler required for window resize
- [ ] No Bubbletea types (tea.Msg, tea.WindowSizeMsg) exposed to users
- [ ] Backward compatible with existing resize handling code

## Dependencies
- **Requires**: 04-composition-api (composables foundation)
- **Requires**: 01-reactivity-system (Ref, Computed, Watch)
- **Requires**: 14-advanced-layout-system (for responsive examples)
- **Unlocks**: Improved DX for all BubblyUI applications

## Edge Cases

### UseWindowSize
1. Terminal resize to 0x0 - enforce minimum dimensions
2. Rapid resize events - debounce calculations
3. SSR/testing without terminal - provide defaults

### UseFocus
1. Empty focus order - return error
2. Focus on non-existent pane - no-op
3. Single item in order - stay focused

### UseScroll
1. Empty list - offset stays at 0
2. Scroll beyond bounds - clamp to valid range
3. Negative visible count - error

### UseSelection
1. Empty items list - no selection
2. Selection beyond bounds - clamp
3. Items change while selected - adjust index

### UseInterval/UseTimeout
1. Negative duration - error
2. Callback panics - recover and continue
3. Component unmount during callback - cancel safely

### UseHistory
1. Undo when at start - no-op
2. Redo when at end - no-op
3. Push clears redo stack
4. Max size exceeded - drop oldest

## Testing Requirements
- Unit test coverage: 80%+
- Integration tests: All composables work together
- Performance benchmarks: <100ns init
- Race detector: All tests pass with -race

## Atomic Design Level
- **Composables** (reusable logic hooks)
- Used by: Atoms, Molecules, Organisms, Templates

## Related Components
- All BubblyUI components can use these composables
- Example apps 14-16 demonstrate usage patterns
- DevTools integration for debugging
