# Composables Examples

This directory contains examples demonstrating the Composition API patterns in BubblyUI. These examples show how to create and use composables for reusable, type-safe component logic.

## Examples Overview

### 1. Counter (`counter/`)
**Demonstrates:** Custom composables and composable chains

Shows how to:
- Create custom composable functions (`UseCounter`)
- Build composable chains (`UseDoubleCounter` → `UseCounter` → `UseState` → `Ref`)
- Return multiple functions from composables
- Compose composables together for reusable logic

**Key Pattern:**
```go
func UseCounter(ctx *bubbly.Context, initial int) UseCounterReturn {
    state := composables.UseState(ctx, initial)
    return UseCounterReturn{
        Count: state.Value,
        Increment: func() { state.Set(state.Get() + 1) },
        // ...
    }
}
```

**Run:**
```bash
go run cmd/examples/04-composables/counter/main.go
```

### 2. Async Data (`async-data/`)
**Demonstrates:** Async data fetching with UseAsync

Shows how to:
- Use the `UseAsync` composable for data fetching
- Handle loading, error, and data states
- Integrate goroutines with Bubbletea using `tea.Tick`
- Trigger async operations on mount
- Implement refetch functionality

**Key Pattern:**
```go
userData := composables.UseAsync(ctx, func() (*User, error) {
    return fetchUserFromAPI()
})

ctx.OnMounted(func() {
    userData.Execute()
})
```

**Bubbletea Integration:**
- Uses `tea.Tick` to periodically trigger UI updates while loading
- Goroutine updates reactive refs
- Tick ensures Bubbletea redraws the UI

**Run:**
```bash
go run cmd/examples/04-composables/async-data/main.go
```

### 3. Form (`form/`)
**Demonstrates:** Complex state management with UseForm

Shows how to:
- Use the `UseForm` composable for form state
- Implement field validation
- Track dirty and valid states
- Handle field updates with type safety
- Display validation errors
- Submit and reset forms

**Key Pattern:**
```go
form := composables.UseForm(ctx, LoginForm{}, func(f LoginForm) map[string]string {
    errors := make(map[string]string)
    if len(f.Username) < 3 {
        errors["Username"] = "Too short"
    }
    return errors
})
```

**Run:**
```bash
go run cmd/examples/04-composables/form/main.go
```

### 4. Provide/Inject (`provide-inject/`)
**Demonstrates:** Dependency injection across component tree

Shows how to:
- Use `ctx.Provide()` to share values with descendants
- Use `ctx.Inject()` to access provided values
- Propagate changes automatically through the tree
- Set up parent-child component relationships
- Provide default fallback values

**Key Pattern:**
```go
// Parent component
ctx.Provide("theme", themeRef)

// Child component (any depth)
theme := ctx.Inject("theme", "default")
```

**Run:**
```bash
go run cmd/examples/04-composables/provide-inject/main.go
```

## Common Patterns

### Composable Function Signature
```go
func UseSomething[T any](ctx *bubbly.Context, initial T) UseSomethingReturn[T]
```

### Composable Chains
Composables can call other composables:
```go
UseDoubleCounter → UseCounter → UseState → Ref
```

### Return Types
Composables typically return structs with:
- Reactive refs (`*bubbly.Ref[T]`)
- Functions for operations
- Cleanup functions (if needed)

### Bubbletea Integration
- **Synchronous operations:** Direct state updates work immediately
- **Asynchronous operations:** Use `tea.Tick` to trigger periodic redraws
- **Event handling:** Use `component.Emit()` and `ctx.On()`
- **Lifecycle:** Use `ctx.OnMounted()`, `ctx.OnUpdated()`, `ctx.OnUnmounted()`

## Best Practices

1. **Always use Context:** All composables must receive `*bubbly.Context` as first parameter
2. **Type safety:** Use generics for type-safe composables
3. **Naming convention:** Prefix composable functions with `Use*`
4. **Return structs:** Return named structs, not tuples
5. **Register cleanup:** Use lifecycle hooks or return cleanup functions
6. **Composable composition:** Build complex composables from simple ones
7. **Don't call in Template:** Only call composables in Setup function

## Testing Composables

Use the testing utilities from `pkg/bubbly/testing`:

```go
import btesting "github.com/newbpydev/bubblyui/pkg/bubbly/testing"

func TestMyComposable(t *testing.T) {
    ctx := btesting.NewTestContext()
    
    result := MyComposable(ctx, "initial")
    
    assert.Equal(t, "initial", result.Get())
}
```

## Further Reading

- [Composables Package Documentation](../../../pkg/bubbly/composables/README.md)
- [Composition API Specification](../../../specs/04-composition-api/)
- [Component Model Examples](../02-component-model/)
- [Lifecycle Hooks Examples](../03-lifecycle-hooks/)

## Running All Examples

```bash
# Counter
go run cmd/examples/04-composables/counter/main.go

# Async Data
go run cmd/examples/04-composables/async-data/main.go

# Form
go run cmd/examples/04-composables/form/main.go

# Provide/Inject
go run cmd/examples/04-composables/provide-inject/main.go
```

## Building Examples

```bash
# Build all examples
go build ./cmd/examples/04-composables/...

# Build specific example
go build ./cmd/examples/04-composables/counter
```
