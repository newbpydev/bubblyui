# Example 17: Enhanced Composables Demo

This example demonstrates all 18 new composables from Feature 15: Enhanced Composables Library.

## Composables Demonstrated

### TUI-Specific (5)
- **UseWindowSize** - Terminal dimensions and responsive breakpoints
- **UseFocus** - Multi-pane focus management
- **UseScroll** - Viewport scrolling
- **UseSelection** - List/table selection
- **UseMode** - Navigation/input mode management

### State Utilities (4)
- **UseToggle** - Boolean toggle
- **UseCounter** - Bounded counter with step
- **UsePrevious** - Previous value tracking
- **UseHistory** - Undo/redo state management

### Timing (3)
- **UseInterval** - Periodic execution
- **UseTimeout** - Delayed execution
- **UseTimer** - Countdown timer with progress

### Collections (4)
- **UseList** - Generic list CRUD
- **UseMap** - Key-value state
- **UseSet** - Unique value set
- **UseQueue** - FIFO queue

### Development (2)
- **UseLogger** - Component debug logging
- **UseNotification** - Toast notifications

## Running the Example

```bash
cd cmd/examples/17-enhanced-composables
go run .
```

## Key Bindings

| Key | Action |
|-----|--------|
| `tab` | Cycle focus between panes |
| `shift+tab` | Cycle focus backwards |
| `up/k` | Scroll up / Previous item |
| `down/j` | Scroll down / Next item |
| `enter` | Select item / Toggle mode |
| `space` | Toggle selection |
| `i` | Switch to input mode |
| `esc` | Switch to navigation mode |
| `+` | Increment counter |
| `-` | Decrement counter |
| `u` | Undo |
| `r` | Redo |
| `t` | Start/stop timer |
| `n` | Show notification |
| `ctrl+c` | Quit |

## Architecture

```
17-enhanced-composables/
├── main.go           # Entry point with bubbly.Run()
├── app.go            # Root component with all composables
└── README.md         # This file
```

## Features Showcased

1. **Responsive Layout** - Uses UseWindowSize for breakpoint-based layout
2. **Focus Management** - Tab cycles through sidebar, main, and footer panes
3. **Scrollable List** - UseScroll + UseSelection for navigable list
4. **Mode Switching** - UseMode for vim-like navigation/input modes
5. **State History** - UseHistory for undo/redo functionality
6. **Timer Display** - UseTimer with progress bar
7. **Notifications** - UseNotification for toast messages
8. **Debug Logging** - UseLogger for component debugging
