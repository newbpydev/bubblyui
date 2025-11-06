# Counter - Automatic Reactive Bridge Example

A simple counter demonstrating the automatic reactive bridge with declarative key bindings.

## Features Demonstrated

✅ **Automatic Reactive Bridge** - State changes trigger UI updates automatically  
✅ **Declarative Key Bindings** - Register keys with `.WithKeyBinding()`  
✅ **Auto-Generated Help Text** - Help text generated from key bindings  
✅ **Zero Boilerplate** - One-line integration with `bubbly.Wrap()`  

## What This Example Shows

This is the **simplest possible** example of the automatic reactive bridge system. It demonstrates:

1. **No Manual `Emit()` Calls** - Just call `count.Set()` and the UI updates automatically
2. **Declarative Keys** - Register key bindings in the builder, not in Update()
3. **Auto Help** - Help text is generated from your key bindings
4. **One-Line Integration** - `bubbly.Wrap()` eliminates wrapper model boilerplate

## How to Run

```bash
cd cmd/examples/08-automatic-bridge/01-counter
go run main.go
```

## Key Bindings

- **Space** - Increment counter (registered as `" "` - space character)
- **R** - Reset to zero
- **Ctrl+C** - Quit application

## Code Walkthrough

### 1. Enable Automatic Commands

```go
bubbly.NewComponent("Counter").
    WithAutoCommands(true).  // Enable automatic reactive bridge
```

This enables automatic command generation from `Ref.Set()` calls.

### 2. Register Key Bindings Declaratively

```go
.WithKeyBinding(" ", "increment", "Increment counter").  // Space is " " not "space"
.WithKeyBinding("r", "reset", "Reset to zero").
.WithKeyBinding("ctrl+c", "quit", "Quit application").
```

No manual key handling in Update() - keys are mapped to events declaratively.

**Important**: The space key is registered as `" "` (a space character), not the string `"space"`. This is how Bubbletea represents the space key in `tea.KeyMsg.String()`.

### 3. Event Handlers - No Manual Emit!

```go
ctx.On("increment", func(_ interface{}) {
    current := count.Get().(int)
    count.Set(current + 1)
    // UI updates automatically - no Emit() needed!
})
```

Just call `count.Set()` and the framework handles the rest.

### 4. Auto-Generated Help Text

```go
Template(func(ctx bubbly.RenderContext) string {
    comp := ctx.Component()
    helpText := comp.HelpText()
    // Returns: " : Increment counter • ctrl+c: Quit application • r: Reset to zero"
})
```

Help text is automatically generated from your key bindings.

### 5. One-Line Integration

```go
tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen()).Run()
```

No manual model struct needed - `bubbly.Wrap()` handles everything.

## Before vs After

### Before (Manual Bridge - 40+ lines)

```go
type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "space":
            m.component.Emit("increment", nil)  // Manual!
        case "r":
            m.component.Emit("reset", nil)      // Manual!
        }
    }
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.component.View()
}

func main() {
    component, _ := createCounter()
    m := model{component: component}
    tea.NewProgram(m, tea.WithAltScreen()).Run()
}
```

### After (Automatic Bridge - 1 line!)

```go
func main() {
    component, _ := createCounter()
    tea.NewProgram(bubbly.Wrap(component), tea.WithAltScreen()).Run()
}
```

**Result**: 97% less boilerplate code!

## What You Gain

1. **Less Code** - 30-50% reduction in application code
2. **Clearer Intent** - Declarative key bindings are self-documenting
3. **Auto Help** - Help text always in sync with actual keys
4. **Fewer Bugs** - Can't forget to emit events
5. **Better DX** - Vue-like developer experience in Go TUI

## Next Steps

- See `02-todo` for a more complex example with mode-based input
- See `03-form` for multi-field forms with validation
- See `04-dashboard` for custom message handling
- Read the [Migration Guide](../../../docs/guides/automatic-bridge-migration.md)

## Related Documentation

- [Automatic Reactive Bridge Requirements](../../../specs/08-automatic-reactive-bridge/requirements.md)
- [Key Binding System Design](../../../specs/08-automatic-reactive-bridge/designs.md)
- [API Documentation](../../../pkg/bubbly/README.md)
