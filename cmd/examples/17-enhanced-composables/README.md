# Example 17: Enhanced Composables Demo

This example demonstrates all 18 new composables from Feature 15: Enhanced Composables Library.

## Structure

```
17-enhanced-composables/
├── main.go                    # Entry point with bubbly.Run()
├── app.go                     # Root component (layout orchestration)
├── composables/
│   └── use_demo_state.go      # Shared state composable (CreateShared)
└── components/
    ├── sidebar_list.go        # Composable list sidebar
    ├── counter_card.go        # Counter demonstration (UseCounter, UseHistory)
    ├── timer_card.go          # Timer demonstration (UseTimer)
    └── collections_card.go    # Collections demonstration (UseList, UseSet, UseToggle)
```

## Composables Demonstrated

### TUI-Specific
- **UseWindowSize** - Terminal dimensions handling (via resize events)
- **UseFocus** - Multi-pane focus management (sidebar/main)

### State Utilities  
- **UseToggle** - Boolean dark mode toggle
- **UseCounter** - Bounded counter with step
- **UsePrevious** - Previous value tracking
- **UseHistory** - Undo/redo state management

### Timing
- **UseTimer** - Countdown timer with progress

### Collections
- **UseList** - Task list storage
- **UseSet** - Tags (sorted to prevent iteration order issues)

### Development
- **UseNotification** - Toast notifications

### Utilities
- **CreateShared** - Singleton state across components

## Running

```bash
go run ./cmd/examples/17-enhanced-composables
```

## Key Bindings

| Key | Action |
|-----|--------|
| `tab` | Cycle focus between panes |
| `up/k` | Navigate up in sidebar |
| `down/j` | Navigate down in sidebar |
| `enter` | Select item in sidebar |
| `+` | Increment counter |
| `-` | Decrement counter |
| `u` | Undo counter change |
| `r` | Redo counter change |
| `t` | Start/stop/reset timer |
| `space` | Toggle dark mode |
| `n` | Show notification |
| `ctrl+c` | Quit |

## Key Patterns Demonstrated

1. **Proper BubblyUI Structure** - No direct Bubbletea, uses bubbly.Run()
2. **Component Separation** - Reusable components in components/
3. **Shared State** - CreateShared for singleton composable
4. **Parent Border Control** - Children render content only
5. **Stable UseSet** - Sorted slice to avoid map iteration order issues
6. **Focus Highlighting** - Visual feedback for selected pane
7. **Card Component Usage** - Using pkg/components instead of raw Lipgloss
