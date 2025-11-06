# Todo App - Pure Bubbletea Version

**Comparison**: This is the pure Bubbletea implementation for comparison with the BubblyUI version.

## Overview

This is a **100% functionally identical** todo application to the BubblyUI version, but implemented using only pure Bubbletea. It demonstrates what you need to write manually when not using BubblyUI's automatic bridge and declarative patterns.

## Features

✅ **Full CRUD operations** - Create, Read, Update, Delete todos  
✅ **Mode-based input** - Navigation mode vs Input mode  
✅ **Form validation** - Title must be at least 3 characters  
✅ **Priority indicators** - 🔴 High, 🟡 Medium, 🟢 Low  
✅ **Statistics** - Total, Completed, Pending counts  
✅ **Visual feedback** - Dynamic border colors based on mode  
✅ **10+ key bindings** - All CRUD operations  

## Running the Example

```bash
# From this directory
go run main.go

# Or build and run
go build -o todo-app main.go
./todo-app
```

## Key Bindings

### Navigation Mode (Default)
- `↑/↓`: Select todo
- `Space`: Toggle completion
- `Ctrl+N`: Create new todo
- `Ctrl+E`: Edit selected todo
- `Ctrl+D`: Delete selected todo
- `Enter`: Enter input mode to add new todo
- `ESC`: Toggle to input mode

### Input Mode
- Type any character to add to current field
- `Tab`: Cycle through fields (Title → Description → Priority)
- `Enter`: Submit form (add or update todo)
- `Backspace`: Delete character from current field
- `ESC`: Return to navigation mode

### Always Active
- `Ctrl+C`: Quit application

## Code Structure

### Model (14 lines)
```go
type model struct {
    // Todo list state
    todos         []Todo
    nextID        int
    selectedIndex int
    
    // Input mode and form state
    inputMode    bool
    editMode     bool
    focusedField string
    
    // Form fields
    titleInput    string
    descInput     string
    priorityInput string
}
```

**Note**: Must manually track ALL state in model struct.

### Init Method (3 lines)
```go
func (m model) Init() tea.Cmd {
    return nil
}
```

**Note**: Required by Bubbletea interface.

### Update Method (~100 lines)
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "esc":
            // Toggle mode logic
        case "ctrl+n":
            // New todo logic
        case "ctrl+e":
            // Edit logic
        case "ctrl+d":
            // Delete logic
        case "up":
            // Selection up logic
        case "down":
            // Selection down logic
        case "enter":
            // Context-dependent enter logic (15+ lines)
        case "tab":
            // Field navigation logic
        case "backspace":
            // Delete character logic (10+ lines)
        default:
            // Space key handling (mode-dependent, 20+ lines)
            // Character input handling (10+ lines)
        }
    }
    return m, nil
}
```

**Note**: ALL keyboard logic lives here. Gets unwieldy with many keys.

### View Method (~150 lines)
```go
func (m model) View() string {
    // Calculate statistics manually
    totalCount := len(m.todos)
    completedCount := 0
    for _, todo := range m.todos {
        if todo.Completed {
            completedCount++
        }
    }
    pendingCount := totalCount - completedCount
    
    // Render UI with Lipgloss
    // ... styling and layout code
    
    // Manual help text (must keep in sync with Update!)
    var help string
    if m.inputMode {
        help = "tab: next field • enter: save • esc: cancel • ctrl+c: quit"
    } else {
        help = "↑/↓: select • space: toggle • ctrl+e: edit • ctrl+d: delete • ctrl+n: new • enter: add • ctrl+c: quit"
    }
    
    return lipgloss.JoinVertical(...)
}
```

**Note**: Must manually sync help text with Update() method.

### Main Function (19 lines)
```go
func main() {
    m := model{
        todos:         []Todo{},
        nextID:        1,
        selectedIndex: 0,
        inputMode:     false,
        editMode:      false,
        focusedField:  "Title",
        titleInput:    "",
        descInput:     "",
        priorityInput: "medium",
    }
    
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Note**: Must manually initialize all state fields.

## What This Demonstrates

### Pure Bubbletea Approach

**Pros**:
- ✅ Simple mental model (just functions)
- ✅ Full control over message flow
- ✅ No framework magic
- ✅ Easy to debug (everything is explicit)
- ✅ Direct state mutation (no Ref.Get/Set)

**Cons**:
- ❌ Massive Update() method (100+ lines)
- ❌ Manual key handling gets messy
- ❌ Help text easily gets out of sync
- ❌ Hard to extract reusable patterns
- ❌ Mode checking scattered everywhere
- ❌ Character input handling is verbose
- ❌ Must manually track all state

### Common Pain Points

1. **Update() Method Grows**: With 10+ key bindings, Update() becomes 100+ lines of nested switch statements.

2. **Help Text Sync**: Easy to forget updating help text when adding/changing keys. No compile-time safety.

3. **Mode Checking Everywhere**: Every key handler needs `if !m.inputMode && !m.editMode` checks.

4. **Character Input**: Handling space key differently based on mode requires nested conditionals.

5. **State Initialization**: Must remember to initialize every field in main().

6. **Manual Statistics**: Must recalculate counts on every render.

## Comparison with BubblyUI

See [../COMPARISON.md](../COMPARISON.md) for a comprehensive comparison.

**Quick Summary**:
- **Pure Bubbletea**: 451 lines, 100-line Update() method, manual help sync
- **BubblyUI**: 583 lines, clean separation, auto help, zero boilerplate

**Trade-off**: BubblyUI has 132 more lines but eliminates 117 lines of boilerplate and provides better architecture.

## When to Use Pure Bubbletea

Use this approach when:
- Building simple apps with few keys
- Learning TUI fundamentals
- Need maximum control over message flow
- Want minimal framework overhead
- Debugging is top priority

## When to Use BubblyUI

Use BubblyUI when:
- Building complex apps with 10+ key bindings
- Need auto-generated help text
- Want mode-based input patterns
- Value maintainability over simplicity
- Need composable, reusable logic

## Related Examples

- **BubblyUI Version**: [../02-todo/](../02-todo/) - Same app with declarative patterns
- **Comparison**: [../COMPARISON.md](../COMPARISON.md) - Detailed side-by-side analysis
- **Simple Counter**: [../../01-counters/](../../01-counters/) - Simpler comparison

---

**Key Takeaway**: Pure Bubbletea gives you full control but requires manual management of everything. For complex apps, the boilerplate adds up quickly. Choose based on your app's complexity and your priorities (control vs productivity). 🎯
