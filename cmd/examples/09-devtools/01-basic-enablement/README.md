# Example 01: Basic Enablement

**Zero-config dev tools getting started**

## What This Demonstrates

This example shows the absolute simplest way to enable dev tools in a BubblyUI application:

1. **Zero-Config Enablement** - Just call `devtools.Enable()`
2. **Component Architecture** - Composable components pattern
3. **Component Tree** - Hierarchical component structure
4. **State Inspection** - Reactive state visible in dev tools
5. **Composables** - Reusable reactive logic (UseCounter)

## Architecture

### Directory Structure
```
01-basic-enablement/
├── main.go                    # Entry point with devtools.Enable()
├── app.go                     # Root component (CounterApp)
├── components/                # UI components
│   ├── counter_display.go     # Display component (shows count)
│   └── counter_controls.go    # Control hints component
├── composables/               # Reusable logic
│   └── use_counter.go         # Counter composable
└── README.md                  # This file
```

### Component Hierarchy
```
CounterApp (root)
├── CounterDisplay (shows current count and parity)
└── CounterControls (shows keyboard shortcuts)
```

### State Flow
```
UseCounter (composable)
├── count (Ref[int])          ← Reactive state
├── isEven (Computed[bool])   ← Derived from count
└── methods (increment, decrement, reset)
```

## Key Features

### 1. Zero-Config Enablement

```go
// main.go
func main() {
    devtools.Enable()  // That's it!
    
    // Rest of your app...
}
```

No configuration needed. Dev tools work out of the box.

### 2. Composable Pattern

```go
// composables/use_counter.go
func UseCounter(ctx *bubbly.Context, initial int) *CounterComposable {
    count := bubbly.NewRef(initial)
    
    isEven := ctx.Computed(func() interface{} {
        return count.Get().(int)%2 == 0
    })
    
    return &CounterComposable{
        Count: count,
        IsEven: isEven,
        Increment: func() { /* ... */ },
        // ...
    }
}
```

Reusable reactive logic that can be shared across components.

### 3. Component Factory Pattern

```go
// components/counter_display.go
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
    builder := bubbly.NewComponent("CounterDisplay")
    
    builder = builder.Setup(func(ctx *bubbly.Context) {
        ctx.Expose("count", props.Count)
        ctx.Expose("isEven", props.IsEven)
    })
    
    builder = builder.Template(func(ctx bubbly.RenderContext) string {
        // Render using BubblyUI components
    })
    
    return builder.Build()
}
```

Clean, testable component creation.

### 4. State Exposure for Dev Tools

```go
// app.go
builder = builder.Setup(func(ctx *bubbly.Context) {
    counter := composables.UseCounter(ctx, 0)
    
    // Expose state for dev tools inspection
    ctx.Expose("counter", counter)
    
    // ...
})
```

All exposed state is visible in dev tools.

## Run the Example

```bash
cd 01-basic-enablement
go run main.go
```

## Using Dev Tools

### Toggle Visibility
Press `F12` to show/hide dev tools panel.

### Component Tree
- Use `↑`/`↓` to navigate components
- Use `→`/`←` to expand/collapse nodes
- Press `Enter` to view component details

### State Inspection
1. Press `Tab` to switch to State tab
2. See all reactive state:
   - `counter.Count` (Ref[int])
   - `counter.IsEven` (Computed[bool])
3. Watch values update as you interact

### Try It Out
1. Press `i` to increment (count goes up)
2. Press `d` to decrement (count goes down)
3. Press `r` to reset (count returns to 0)
4. Watch dev tools update in real-time!

## What to Notice

### Component Tree
```
CounterApp
├── CounterDisplay (*)  ← You'll see this
└── CounterControls     ← And this
```

Both components appear in the tree, with their state visible.

### Reactive State
```
counter.Count: 0  ← Changes when you press i/d/r
counter.IsEven: true  ← Auto-updates when count changes
```

Computed values update automatically!

### Lifecycle Events
Open your terminal console to see:
```
[CounterDisplay] Mounted - visible in dev tools!
```

Lifecycle hooks fire and are visible.

## Code Highlights

### Composable Reuse
The `UseCounter` composable encapsulates ALL counter logic:
- State management (Ref)
- Derived values (Computed)
- Methods (increment, decrement, reset)

This can be reused in ANY component that needs counter functionality.

### Component Composition
The app composes smaller components:
- `CounterDisplay` handles display logic
- `CounterControls` shows shortcuts
- `CounterApp` coordinates them

Each component is focused and testable.

### BubblyUI Components
We use framework components, not manual rendering:
- `components.Card()` for the counter display
- `components.Text()` for help text

No hardcoded Lipgloss styles!

## Next Steps

After understanding this example:

1. **Explore component tree** - Navigate the hierarchy with keyboard
2. **Watch state changes** - See reactive updates in real-time
3. **Read the architecture guide** - [Composable Apps Guide](../../../../docs/architecture/composable-apps.md)
4. **Try example 02** - More complex component hierarchy

## Related Documentation

- [Composable Apps Architecture](../../../../docs/architecture/composable-apps.md)
- [Dev Tools Quickstart](../../../../docs/devtools/quickstart.md)
- [Dev Tools Features](../../../../docs/devtools/features.md)
- [Component Reference](../../../../docs/components/README.md)

## Troubleshooting

**Dev tools not showing?**
- Make sure you called `devtools.Enable()` before creating components
- Press `F12` to toggle visibility

**Can't see state?**
- Check that you're exposing state with `ctx.Expose()`
- Switch to the State tab in dev tools

**Component not in tree?**
- Verify `ExposeComponent()` was called
- Check component was created successfully (no errors)

---

**Next:** [Example 02 - Component Inspection](../02-component-inspection/) →
