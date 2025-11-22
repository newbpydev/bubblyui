# Counter App - Testing Example

A comprehensive example demonstrating BubblyUI's testing utilities, composables pattern, and the **WithMultiKeyBindings** automation feature.

## Overview

This example showcases:
- **Testing patterns** with testutil harness
- **Composables** for state management (UseCounter)
- **Component composition** (CounterDisplay component)
- **Multi-key bindings** using `WithMultiKeyBindings` helper
- **Reactive state** with Ref and Computed values
- **Event handling** with key bindings

## Architecture

```
CounterApp (Root)
├── UseCounter composable (state management)
│   ├── Count (Ref[int])
│   ├── Doubled (Computed[int])
│   ├── IsEven (Computed[bool])
│   └── History (Ref[[]int])
└── CounterDisplay component (presentation)
    └── Renders count, doubled, parity, history
```

## Key Features

### 1. Multi-Key Bindings (NEW)

This example demonstrates the **WithMultiKeyBindings** automation pattern introduced in Feature 13:

**Before (6 lines):**
```go
.WithKeyBinding("up", "increment", "Increment counter").
.WithKeyBinding("k", "increment", "Increment counter").
.WithKeyBinding("+", "increment", "Increment counter").
.WithKeyBinding("down", "decrement", "Decrement counter").
.WithKeyBinding("j", "decrement", "Decrement counter").
.WithKeyBinding("-", "decrement", "Decrement counter")
```

**After (2 lines):**
```go
.WithMultiKeyBindings("increment", "Increment counter", "up", "k", "+").
.WithMultiKeyBindings("decrement", "Decrement counter", "down", "j", "-")
```

**Benefits:**
- **67% code reduction** for multi-key bindings
- **Clearer intent** - grouped related keys
- **Easier maintenance** - add/remove keys in one place
- **Consistent pattern** - same description for all keys

### 2. Composables Pattern

The `UseCounter` composable encapsulates counter logic:
- State management with `Ref[int]`
- Computed values (doubled, parity)
- History tracking (last 5 values)
- Methods: Increment, Decrement, Reset, SetValue

### 3. Component Composition

Separates concerns:
- **App component** - handles key bindings and events
- **Display component** - renders counter state
- **Composable** - manages state logic

### 4. Comprehensive Testing

24 tests covering:
- Basic mounting and rendering
- Event emission and handling
- Table-driven test patterns
- Event tracking
- Multiple operations
- Cleanup verification

## Running the Example

```bash
# Run the application
cd cmd/examples/10-testing/01-counter
go run .

# Run tests
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

## Key Bindings

| Keys | Action | Description |
|------|--------|-------------|
| ↑, k, + | Increment | Increase counter by 1 |
| ↓, j, - | Decrement | Decrease counter by 1 |
| r | Reset | Reset counter to 0 |
| q, Ctrl+C | Quit | Exit application |

## Testing Patterns

### 1. Basic Mounting
```go
harness := testutil.NewHarness(t)
app, _ := CreateApp()
ct := harness.Mount(app)
ct.AssertRenderContains("Count: 0")
```

### 2. Event Emission
```go
ct.Emit("increment", nil)
ct.AssertRenderContains("Count: 1")
```

### 3. Table-Driven Tests
```go
tests := []struct {
    name   string
    events []string
    expect string
}{
    {"single increment", []string{"increment"}, "Count: 1"},
    {"multiple increments", []string{"increment", "increment"}, "Count: 2"},
}
```

### 4. Event Tracking
```go
ct.Emit("increment", nil)
ct.AssertEventFired("increment")
ct.AssertEventCount("increment", 1)
```

## Code Metrics

### Before Migration
- Key bindings: 10 lines (6 multi-key + 4 single-key)
- Total app.go: 91 lines

### After Migration
- Key bindings: 6 lines (2 multi-key + 4 single-key)
- Total app.go: 87 lines
- **Code reduction: 4 lines (67% for multi-key bindings)**

### Test Coverage
- `app.go`: 71%
- `components/counter_display.go`: 100%
- `composables/use_counter.go`: 100%
- **Total: 24 tests, all passing**

## File Structure

```
01-counter/
├── main.go                          # Entry point
├── app.go                           # Root component with key bindings
├── app_test.go                      # Integration tests (7 tests)
├── components/
│   ├── counter_display.go           # Display component
│   └── counter_display_test.go      # Component tests (6 tests)
├── composables/
│   ├── use_counter.go               # Counter composable
│   └── use_counter_test.go          # Composable tests (11 tests)
└── README.md                        # This file
```

## Learning Objectives

This example teaches:
1. **Testing** - How to test BubblyUI components with testutil
2. **Composables** - How to create reusable state logic
3. **Components** - How to compose components hierarchically
4. **Key Bindings** - How to use WithMultiKeyBindings for cleaner code
5. **Reactivity** - How Ref and Computed values work together
6. **Events** - How to emit and handle events

## Related Examples

- **04-async** - Demonstrates UseTheme/ProvideTheme automation
- **11-advanced-patterns** - Advanced composable patterns (CreateShared)

## Migration Notes

This example was migrated from individual `WithKeyBinding` calls to `WithMultiKeyBindings` as part of Feature 13 (Advanced Internal Package Automation). The migration demonstrates:

- **Zero breaking changes** - All tests pass without modification
- **Identical functionality** - All keys work exactly as before
- **Improved maintainability** - Easier to add/remove alternative keys
- **Clear intent** - Grouped keys show they perform the same action

## References

- **Feature 13 Spec**: `specs/13-adv-internal-package-automation/`
- **Task 5.2**: Counter example migration to WithMultiKeyBindings
- **testutil API**: `docs/api/testutil-reference.md`
- **Composables Guide**: `docs/composables.md`
