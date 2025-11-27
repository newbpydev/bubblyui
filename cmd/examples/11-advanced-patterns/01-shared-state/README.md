# Shared State Example - Advanced Pattern

This example demonstrates the **CreateShared** pattern for sharing state across multiple components without prop drilling or global variables.

## 🎯 What This Example Shows

- **Shared Composable Pattern**: Using `composables.CreateShared()` to create singleton instances
- **State Synchronization**: Two independent components sharing the same counter state
- **Zero Bubbletea**: Uses `bubbly.Run()` - no manual Bubbletea code
- **BubblyUI Components**: Uses Card component instead of raw Lipgloss
- **Comprehensive Tests**: Full test coverage using testutil harness

## 🏗️ Architecture

### File Structure
```
01-shared-state/
├── main.go                          # Entry point with bubbly.Run()
├── app.go                           # Root component
├── app_test.go                      # Comprehensive tests with testutil
├── composables/
│   ├── use_counter.go              # Counter composable
│   └── shared_counter.go           # Shared counter using CreateShared
└── components/
    ├── counter_display.go          # Display component (reads state)
    └── counter_controls.go         # Controls component (modifies state)
```

### Component Hierarchy
```
SharedStateApp
├─ CounterDisplay (uses UseSharedCounter)
│  └─ Reads: Count, Doubled, IsEven, History
└─ CounterControls (uses UseSharedCounter)
   └─ Modifies: Increment, Decrement, Reset
```

## 🔑 Key Patterns Demonstrated

### 1. CreateShared Pattern
```go
// composables/shared_counter.go
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *CounterComposable {
        return UseCounter(ctx, 0)
    },
)
```

### 2. Zero Bubbletea
```go
// main.go - NO tea.NewProgram!
func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())  // Zero boilerplate!
}
```

### 3. Component Factory Pattern
```go
// components/counter_display.go
func CreateCounterDisplay(props CounterDisplayProps) (bubbly.Component, error) {
    return bubbly.NewComponent("CounterDisplay").
        Setup(func(ctx *bubbly.Context) {
            counter := localComposables.UseSharedCounter(ctx)
            ctx.Expose("counter", counter)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            // Use Card component
            card := components.Card(components.CardProps{...})
            card.Init()
            return card.View()
        }).
        Build()
}
```

### 4. Comprehensive Testing with testutil
```go
// app_test.go
func TestApp_SharedStateSync(t *testing.T) {
    harness := testutil.NewHarness(t)
    defer harness.Cleanup()

    app, _ := CreateApp()
    ct := harness.Mount(app)
    defer ct.Unmount()

    ct.Emit("increment", nil)
    ct.AssertRenderContains("1")  // Both components show 1!
}
```

## 🚀 Running the Example

```bash
# From repository root
cd cmd/examples/11-advanced-patterns/01-shared-state
go run .
```

## 🧪 Running Tests

```bash
# Run all tests with race detector
go test -v -race .

# Run specific test
go test -v -run TestApp_SharedStateSync

# With coverage
go test -v -cover .
```

## 🎮 Controls

- **↑/k/+**: Increment counter
- **↓/j/-**: Decrement counter
- **r**: Reset counter to 0
- **q**: Quit application

## 📊 Test Coverage

- ✅ `TestApp_Creation` - Component creation
- ✅ `TestApp_SharedStateSync` - State synchronization across components
- ✅ `TestApp_HistoryTracking` - History tracking functionality
- ✅ `TestApp_ComputedValues` - Computed values (Doubled, IsEven)
- ✅ `TestApp_KeyBindings` - Multi-key bindings verification

All tests pass with race detector (`-race` flag).

## 🔍 What Makes This Different

### Before (Prop Drilling)
```go
// Parent owns state, passes to children
counter := UseCounter(ctx, 0)
display, _ := CreateCounterDisplay(CounterDisplayProps{
    Count: counter.Count,
    // ... pass all props
})
```

### After (Shared State)
```go
// Each component gets shared instance
counter := UseSharedCounter(ctx)  // Same instance everywhere!
```

## 📚 VueUse Inspiration

This pattern is inspired by VueUse's `createSharedComposable`:

**VueUse (JavaScript)**:
```javascript
const useSharedCounter = createSharedComposable(useCounter)
const counter = useSharedCounter()  // Same instance
```

**BubblyUI (Go)**:
```go
var UseSharedCounter = composables.CreateShared(UseCounter)
counter := UseSharedCounter(ctx)  // Same instance
```

## 💡 When to Use CreateShared

✅ **Good Use Cases:**
- Global application state (user session, settings)
- Shared services (logger, API client)
- Cross-component communication without prop drilling

❌ **Avoid For:**
- Component-local state (use regular composables)
- Parent-child communication (use props/events)

## 🎓 Learning Objectives

After studying this example, you should understand:

1. How to use `CreateShared` for singleton composables
2. How to avoid prop drilling with shared state
3. How to use `bubbly.Run()` instead of manual Bubbletea
4. How to test with testutil harness
5. How to use BubblyUI components (Card) instead of raw Lipgloss
6. How to structure composables/components/app pattern

## ⚠️ Important Notes

- **Shared state persists** across component instances
- **Thread-safe** via `sync.Once` initialization
- **Test isolation** - reset state between tests if needed
- **Zero Bubbletea** - framework handles all Bubbletea code

## 🔗 Related Patterns

- **Provide/Inject**: For dependency injection (theme, config)
- **Props**: For parent-child communication
- **Events**: For child-parent communication
- **Regular Composables**: For component-local state

---

**This example demonstrates BubblyUI's advanced state management capabilities, following the framework's zero-boilerplate philosophy.**
