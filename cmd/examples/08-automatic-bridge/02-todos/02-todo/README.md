# Todo App with Declarative Key Bindings

**Feature**: Automatic Reactive Bridge (Task 9.2)  
**Demonstrates**: Full CRUD todo app with mode-based input using declarative key bindings

## Overview

This example showcases the power of BubblyUI's **declarative key binding system** combined with the **automatic reactive bridge**. It's a full-featured todo application with:

- ✅ **10+ key bindings** for CRUD operations
- ✅ **Conditional bindings** for navigation vs input modes
- ✅ **Auto-generated help text** from binding descriptions
- ✅ **List rendering** with selection and priority indicators
- ✅ **Form input handling** with validation
- ✅ **Mode indicators** with dynamic visual feedback

## Key Features

### 1. Zero-Boilerplate Integration
```go
// That's it! No manual wrapper model needed
tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen()).Run()
```

### 2. Declarative Key Bindings
```go
builder := bubbly.NewComponent("TodoApp").
    WithAutoCommands(true).
    WithKeyBinding("ctrl+c", "quit", "Quit application").
    WithKeyBinding("ctrl+n", "newTodo", "New todo").
    WithKeyBinding("ctrl+e", "editTodo", "Edit selected").
    WithKeyBinding("ctrl+d", "deleteTodo", "Delete selected").
    WithKeyBinding("up", "selectPrevious", "Previous todo").
    WithKeyBinding("down", "selectNext", "Next todo").
    // ... more bindings
```

### 2.5. Message Handler for Character Input
```go
// Message handler captures all character input (a-z, A-Z, 0-9, etc.)
builder.WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        switch keyMsg.Type {
        case tea.KeyRunes:
            // Forward to addChar event (which checks input mode)
            comp.Emit("addChar", string(keyMsg.Runes))
        }
    }
    return nil
})
```

**Why Needed**: Declarative key bindings are perfect for specific keys (`ctrl+c`, `enter`, `tab`), but you need a message handler to capture **all character input** for typing in forms.

### 4. Conditional Key Bindings (Mode-Based)
```go
// Space key behaves differently based on mode
builder.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key:         " ", // CRITICAL: Use space character, not "space" string
    Event:       "toggleTodo",
    Description: "Toggle completion",
    Condition: func() bool {
        return !inputModeRef.Get().(bool) // Only in navigation mode
    },
}).WithConditionalKeyBinding(bubbly.KeyBinding{
    Key:         " ",
    Event:       "addChar",
    Data:        " ",
    Description: "Add space",
    Condition: func() bool {
        return inputModeRef.Get().(bool) // Only in input mode
    },
})
```

### 5. Automatic UI Updates
```go
ctx.On("toggleTodo", func(_ interface{}) {
    todoList := todos.Get().([]Todo)
    selected := selectedIndex.Get().(int)
    
    if len(todoList) > 0 && selected >= 0 && selected < len(todoList) {
        todoList[selected].Completed = !todoList[selected].Completed
        todos.Set(todoList) // UI updates automatically!
    }
})
```

## Before vs After Comparison

### Before: Manual Keyboard Routing (677 lines)

```go
// Manual wrapper model with 40+ lines of keyboard routing
type model struct {
    component    bubbly.Component
    selectedTodo int
    editMode     bool
    inputMode    bool
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        // 40+ lines of manual switch/case for every key
        if msg.Type == tea.KeySpace {
            if !m.inputMode && !m.editMode {
                m.component.Emit("toggleTodo", m.selectedTodo)
            } else if m.inputMode {
                m.component.Emit("addChar", " ")
            }
        } else {
            switch msg.String() {
            case "ctrl+c":
                return m, tea.Quit
            case "esc":
                m.inputMode = !m.inputMode
                m.component.Emit("setInputMode", m.inputMode)
                // ... more cases
            case "ctrl+n":
                m.editMode = false
                m.inputMode = true
                m.component.Emit("setInputMode", m.inputMode)
                m.component.Emit("clearForm", nil)
            // ... 15+ more cases
            }
        }
    }

    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    // ... more boilerplate
}
```

### After: Declarative Key Bindings (567 lines)

```go
// Zero-boilerplate integration
func main() {
    component, err := createTodoApp()
    if err != nil {
        fmt.Printf("Error creating component: %v\n", err)
        os.Exit(1)
    }

    // One line! Automatic bridge handles everything
    tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen()).Run()
}

// Component with declarative key bindings
builder := bubbly.NewComponent("TodoApp").
    WithAutoCommands(true).
    WithKeyBinding("ctrl+c", "quit", "Quit application").
    WithKeyBinding("ctrl+n", "newTodo", "New todo").
    WithKeyBinding("ctrl+e", "editTodo", "Edit selected").
    // ... 10+ bindings declared upfront
    WithConditionalKeyBinding(bubbly.KeyBinding{
        Key:         " ",
        Event:       "toggleTodo",
        Condition: func() bool { return !inputMode.Get().(bool) },
    }).
    Setup(func(ctx *bubbly.Context) {
        // Just event handlers - no keyboard routing!
        ctx.On("toggleTodo", func(_ interface{}) {
            // State changes auto-update UI
        })
    })
```

### Code Reduction Metrics

| Metric | Manual | Declarative | Reduction |
|--------|--------|-------------|-----------|
| **Total Lines** | 677 | 567 | **16% fewer** |
| **Wrapper Model** | 40 lines | 0 lines | **100% eliminated** |
| **Keyboard Routing** | 40 lines | 10 lines | **75% reduction** |
| **Integration** | 15 lines | 1 line | **93% reduction** |

## Mode-Based Input Handling

The app has two distinct modes with clear visual feedback:

### Navigation Mode (Default)
- **Visual**: Purple border on todo list, dark grey on form
- **Indicator**: 🧭 NAVIGATION MODE (purple background)
- **Keys Active**:
  - `↑/↓`: Select todo
  - `Space`: Toggle completion
  - `Ctrl+E`: Edit selected
  - `Ctrl+D`: Delete selected
  - `Ctrl+N`: New todo
  - `Enter`: Enter input mode

### Input Mode
- **Visual**: Green border on form, dark grey on todo list
- **Indicator**: ✍️ INPUT MODE (green background)
- **Keys Active**:
  - `Tab`: Next field
  - `Enter`: Save todo
  - `ESC`: Cancel and return to navigation
  - `Backspace`: Delete character
  - `Space`: Add space character (not toggle!)
  - Any character: Type into current field

## Key Bindings Reference

### Always Active
- `Ctrl+C`: Quit application
- `ESC`: Toggle between navigation and input modes

### Navigation Mode Only
- `↑`: Select previous todo
- `↓`: Select next todo
- `Space`: Toggle todo completion
- `Ctrl+N`: Create new todo
- `Ctrl+E`: Edit selected todo
- `Ctrl+D`: Delete selected todo
- `Enter`: Enter input mode to add new todo

### Input Mode Only
- `Tab`: Cycle through form fields (Title → Description → Priority)
- `Enter`: Submit form (add or update todo)
- `Backspace`: Delete character from current field
- `Space`: Add space character to current field
- Any character: Type into current field

## Running the Example

```bash
# From project root
go run cmd/examples/08-automatic-bridge/02-todos/02-todo/main.go

# Or build and run
go build -o todo-app cmd/examples/08-automatic-bridge/02-todos/02-todo/main.go
./todo-app
```

## Code Walkthrough

### 1. Component Creation with Declarative Bindings

```go
builder := bubbly.NewComponent("TodoApp").
    WithAutoCommands(true). // Enable automatic reactive bridge
    // Declare all key bindings upfront
    WithKeyBinding("ctrl+c", "quit", "Quit application").
    WithKeyBinding("ctrl+n", "newTodo", "New todo").
    // ... more bindings
```

**Benefits**:
- All keys visible at component definition
- Self-documenting code
- Auto-generated help text
- No manual Update() logic

### 2. Message Handler for Character Input

```go
builder.WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
    // Capture all character input
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        switch keyMsg.Type {
        case tea.KeyRunes:
            // Forward to addChar event
            comp.Emit("addChar", string(keyMsg.Runes))
        }
    }
    return nil
})
```

**Why This Pattern**:
- **Declarative bindings** handle specific keys: `"ctrl+c"`, `"enter"`, `"tab"`, `"esc"`
- **Message handler** captures all character input: `a-z`, `A-Z`, `0-9`, punctuation
- **Event handler** (`addChar`) checks input mode and processes characters
- **Separation of concerns**: Message handler captures, event handler validates

### 3. Conditional Key Bindings for Modes

```go
// Space key: Toggle in navigation mode
builder.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key:         " ", // CRITICAL: Space character, not "space" string
    Event:       "toggleTodo",
    Description: "Toggle completion",
    Condition: func() bool {
        return !inputModeRef.Get().(bool) // Check mode
    },
})

// Space key: Add character in input mode
builder.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key:         " ",
    Event:       "addChar",
    Data:        " ",
    Condition: func() bool {
        return inputModeRef.Get().(bool) // Check mode
    },
})
```

**Key Points**:
- Same key, different behavior based on condition
- Conditions evaluated on every keypress
- First matching binding wins
- Clean separation of concerns

### 3. Event Handlers (Semantic Actions)

```go
ctx.On("toggleTodo", func(_ interface{}) {
    if !inputMode.Get().(bool) { // Only in navigation mode
        todoList := todos.Get().([]Todo)
        selected := selectedIndex.Get().(int)
        
        if len(todoList) > 0 && selected >= 0 && selected < len(todoList) {
            todoList[selected].Completed = !todoList[selected].Completed
            todos.Set(todoList) // UI updates automatically!
        }
    }
})
```

**Benefits**:
- Focus on business logic, not keyboard routing
- Automatic UI updates from `Ref.Set()`
- No manual `Emit()` calls needed
- Clean, testable code

### 4. Visual Mode Indicators

```go
// Dynamic border colors based on mode
formBorderColor := "240" // Dark grey (navigation mode - inactive)
if inInputMode {
    formBorderColor = "35" // Green (input mode - active)
}

todoBorderColor := "99" // Purple (navigation mode - active)
if inInputMode {
    todoBorderColor = "240" // Dark grey (input mode - inactive)
}

// Mode indicator badge
if inInputMode {
    modeIndicator = "✍️  INPUT MODE - Type to add todo, ESC to navigate"
} else {
    modeIndicator = "🧭 NAVIGATION MODE - Use shortcuts, ENTER to add todo"
}
```

**UX Benefits**:
- Clear visual feedback of current mode
- Active pane highlighted (green for input, purple for navigation)
- Inactive pane dimmed (dark grey)
- Mode-specific help text

## CRITICAL: Space Key Registration

⚠️ **IMPORTANT**: When registering the space key, you MUST use `" "` (a single space character), NOT the string `"space"`.

```go
// ✅ CORRECT
.WithKeyBinding(" ", "toggle", "Toggle")

// ❌ WRONG - will never trigger!
.WithKeyBinding("space", "toggle", "Toggle")
```

**Why**: Bubbletea's `tea.KeyMsg.String()` returns `" "` for the space key. Key bindings are matched using `keyMsg.String()`, so if you register as `"space"`, it will never match the actual space keypress.

## What This Example Demonstrates

1. **Zero Boilerplate**: One-line integration with `bubbly.Wrap()`
2. **Declarative Key Bindings**: All keys declared upfront, no manual routing
3. **Message Handler Pattern**: Escape hatch for capturing all character input
4. **Conditional Bindings**: Mode-based key behavior without boilerplate
5. **Automatic Updates**: `Ref.Set()` triggers UI updates automatically
6. **Auto-Generated Help**: Help text generated from binding descriptions
7. **Mode-Based Input**: Professional TUI pattern for navigation vs typing
8. **Visual Feedback**: Dynamic colors and indicators for current mode
9. **Full CRUD**: Create, Read, Update, Delete operations
10. **Form Validation**: Real-time validation with error messages
11. **Statistics**: Computed values for total/completed/pending counts
12. **Separation of Concerns**: Message handler captures, event handler validates

## Next Steps

- **Task 9.3**: Form with mode-based bindings (multi-field forms)
- **Task 9.4**: Dashboard with message handler (custom messages)
- **Task 9.5**: Complex app with tree structure (nested components)

## Related Documentation

- [Automatic Reactive Bridge Spec](../../../../specs/08-automatic-reactive-bridge/)
- [Key Bindings Design](../../../../specs/08-automatic-reactive-bridge/designs.md#declarative-key-binding-system)
- [User Workflow](../../../../specs/08-automatic-reactive-bridge/user-workflow.md#workflow-4-declarative-key-bindings-for-zero-boilerplate)
- [Previous Example: Simple Counter](../01-counter/)

---

**Key Takeaway**: Declarative key bindings + automatic bridge = **zero boilerplate** for keyboard-driven TUI applications. Focus on business logic, not plumbing!
