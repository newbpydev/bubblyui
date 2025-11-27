# Design Specification: Advanced Internal Package Automation

## Feature ID
13-adv-internal-package-automation

## Component Hierarchy

```
BubblyUI Framework (Foundation)
├── Context API (Enhanced)
│   ├── UseTheme() method
│   ├── ProvideTheme() method
│   └── Existing Provide/Inject
├── ComponentBuilder (Enhanced)
│   ├── WithKeyBindings() method
│   └── Existing WithKeyBinding()
├── Theme System (NEW)
│   ├── Theme struct
│   └── DefaultTheme constant
└── Composables Utilities (NEW)
    └── CreateShared[T]() factory

Application Layer (Usage)
├── App Component
│   ├── ProvideTheme(theme) → Descendants
│   └── WithKeyBindings("event", "desc", keys...)
└── Child Components
    ├── UseTheme(default) ← Injection
    └── Access theme.Primary, theme.Secondary, etc.
```

## Architecture Overview

### 1. Theme System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Parent Component                         │
│                                                              │
│  Setup(func(ctx *Context) {                                 │
│    theme := DefaultTheme                                    │
│    theme.Primary = lipgloss.Color("99")  // Override       │
│    ctx.ProvideTheme(theme)                                 │
│  })                                                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ ctx.Provide("theme", theme)
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                     Child Component                          │
│                                                              │
│  Setup(func(ctx *Context) {                                 │
│    theme := ctx.UseTheme(DefaultTheme)  // 1 line!         │
│    // vs 15 lines of inject+expose                         │
│                                                              │
│    // Use theme colors                                      │
│    titleStyle := lipgloss.NewStyle().                       │
│                 Foreground(theme.Primary)                   │
│  })                                                         │
└─────────────────────────────────────────────────────────────┘
```

### 2. Multi-Key Binding Architecture

```
Component Builder Pattern:

bubbly.NewComponent("Counter").
  WithKeyBindings("increment", "Increment", "up", "k", "+").
  WithKeyBindings("decrement", "Decrement", "down", "j", "-").
  Build()

Internally expands to:
  .WithKeyBinding("up", "increment", "Increment").
  .WithKeyBinding("k", "increment", "Increment").
  .WithKeyBinding("+", "increment", "Increment").
  .WithKeyBinding("down", "decrement", "Decrement").
  .WithKeyBinding("j", "decrement", "Decrement").
  .WithKeyBinding("-", "decrement", "Decrement")
```

### 3. Shared Composable Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Composable Factory Definition                   │
│                                                              │
│  var UseSharedCounter = CreateShared(                       │
│    func(ctx *Context) *CounterComposable {                 │
│      return UseCounter(ctx, 0)                             │
│    },                                                       │
│  )                                                          │
└──────────────────────┬──────────────────────────────────────┘
                       │ sync.Once initialization
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              Component A                 Component B         │
│                                                              │
│  counter := UseSharedCounter(ctx)  ←  Same instance        │
│  // counter.Count same in both components                   │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

### Theme Injection Flow
```
1. Parent: ctx.ProvideTheme(theme)
   ↓ Store in Context: ctx.Provide("theme", theme)
   
2. Child: theme := ctx.UseTheme(DefaultTheme)
   ↓ Check injection: ctx.Inject("theme", nil)
   ↓ Type assert: injected.(Theme)
   ↓ Return injected or default
   
3. Usage: theme.Primary, theme.Secondary, etc.
   ↓ Type-safe struct access
   ↓ lipgloss.NewStyle().Foreground(theme.Primary)
```

### Multi-Key Binding Flow
```
1. Builder: .WithKeyBindings("event", "desc", "key1", "key2"...)
   ↓ Loop over keys array
   
2. For each key:
   ↓ .WithKeyBinding(key, event, desc)
   ↓ Store in keyBindings map
   
3. Runtime: tea.KeyMsg received
   ↓ Look up key in keyBindings
   ↓ Emit corresponding event
   ↓ ctx.On("event") handler executes
```

### Shared Composable Flow
```
1. Define: UseSharedX = CreateShared(factory)
   ↓ Store factory function
   ↓ Create sync.Once
   
2. First call: UseSharedX(ctx)
   ↓ once.Do(func() { instance = factory(ctx) })
   ↓ Initialize and store instance
   
3. Subsequent calls: UseSharedX(ctx)
   ↓ once.Do(...) skipped (already executed)
   ↓ Return stored instance
   ↓ Same state across all components
```

## Type Definitions

### Theme System

```go
// pkg/bubbly/theme.go

package bubbly

import "github.com/charmbracelet/lipgloss"

// Theme defines a standard color palette for BubblyUI components.
// It provides semantic color names that can be consistently used across
// the component hierarchy via Provide/Inject pattern.
type Theme struct {
	// Primary is the main accent color (e.g., brand color)
	Primary lipgloss.Color
	
	// Secondary is the secondary accent color
	Secondary lipgloss.Color
	
	// Muted is used for less prominent text and UI elements
	Muted lipgloss.Color
	
	// Warning is used for warning messages and caution states
	Warning lipgloss.Color
	
	// Error is used for error messages and danger states
	Error lipgloss.Color
	
	// Success is used for success messages and positive states
	Success lipgloss.Color
	
	// Background is the default background color
	Background lipgloss.Color
}

// DefaultTheme provides sensible defaults for the theme colors.
// Components can use this as a fallback when no parent provides a theme.
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("35"),  // Green
	Secondary:  lipgloss.Color("99"),  // Purple
	Muted:      lipgloss.Color("240"), // Dark grey
	Warning:    lipgloss.Color("220"), // Yellow
	Error:      lipgloss.Color("196"), // Red
	Success:    lipgloss.Color("35"),  // Green
	Background: lipgloss.Color("236"), // Dark background
}

// UseTheme retrieves the theme from parent via injection or returns the default.
// This eliminates the boilerplate of manual inject+expose for theme colors.
//
// Usage in child component:
//   theme := ctx.UseTheme(DefaultTheme)
//   style := lipgloss.NewStyle().Foreground(theme.Primary)
func (ctx *Context) UseTheme(defaultTheme Theme) Theme {
	if injected := ctx.Inject("theme", nil); injected != nil {
		if theme, ok := injected.(Theme); ok {
			return theme
		}
	}
	return defaultTheme
}

// ProvideTheme provides a theme to all descendant components.
// Parent components should call this in their Setup function.
//
// Usage in parent component:
//   ctx.ProvideTheme(myCustomTheme)
func (ctx *Context) ProvideTheme(theme Theme) {
	ctx.Provide("theme", theme)
}
```

### Multi-Key Binding

```go
// pkg/bubbly/component_builder.go (ADDITION)

// WithKeyBindings registers multiple keys for the same event (convenience method).
// This eliminates repetitive WithKeyBinding calls when multiple keys should
// trigger the same action (e.g., "up", "k", "+" all increment a counter).
//
// The description applies to all keys. If different descriptions are needed,
// use separate WithKeyBinding calls.
//
// Example:
//   .WithKeyBindings("increment", "Increment counter", "up", "k", "+")
//
// is equivalent to:
//   .WithKeyBinding("up", "increment", "Increment counter").
//   .WithKeyBinding("k", "increment", "Increment counter").
//   .WithKeyBinding("+", "increment", "Increment counter")
//
// Parameters:
//   - event: The event name to emit when any of the keys is pressed
//   - description: Help text shown for these keys
//   - keys: Variadic list of keys that trigger this event
func (b *ComponentBuilder) WithKeyBindings(event, description string, keys ...string) *ComponentBuilder {
	for _, key := range keys {
		b.WithKeyBinding(key, event, description)
	}
	return b
}
```

### Shared Composable

```go
// pkg/bubbly/composables/shared.go (NEW PACKAGE)

package composables

import (
	"sync"
	
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateShared wraps a composable factory to return a singleton instance.
// Inspired by VueUse's createSharedComposable, this enables sharing state
// and logic across multiple components without prop drilling or global variables.
//
// The factory function is called exactly once (thread-safe via sync.Once),
// and all subsequent calls return the same instance.
//
// Type-safe with Go generics.
//
// Example:
//   var UseSharedCounter = CreateShared(
//     func(ctx *bubbly.Context) *CounterComposable {
//       return UseCounter(ctx, 0)
//     },
//   )
//
//   // In any component - same instance across all
//   counter := UseSharedCounter(ctx)
//
// Note: The context passed to the factory is from the first component that
// calls the shared composable. Ensure the factory doesn't rely on
// component-specific context state.
func CreateShared[T any](factory func(*bubbly.Context) T) func(*bubbly.Context) T {
	var instance T
	var once sync.Once
	
	return func(ctx *bubbly.Context) T {
		once.Do(func() {
			instance = factory(ctx)
		})
		return instance
	}
}
```

## State Management

### Theme State
- **Storage**: Theme struct in parent component's Setup
- **Propagation**: Via Context.Provide("theme", theme)
- **Consumption**: Via Context.UseTheme(default)
- **Scope**: Component tree (descendants of provider)
- **Mutability**: Immutable after provision (override by re-providing)

### Multi-Key Binding State
- **Storage**: ComponentBuilder's keyBindings map
- **Registration**: During component creation (builder pattern)
- **Resolution**: During component runtime (Update method)
- **Scope**: Per-component (each component has own bindings)
- **Mutability**: Immutable after Build()

### Shared Composable State
- **Storage**: Closure variable in CreateShared return function
- **Initialization**: First call to returned function (sync.Once)
- **Access**: All components calling the shared factory
- **Scope**: Application-wide (singleton)
- **Mutability**: Depends on composable implementation
- **Thread Safety**: Protected by sync.Once for initialization

## API Design Patterns

### 1. Progressive Enhancement
```go
// Basic: Manual inject/expose (still works)
primaryColor := lipgloss.Color("35")
if injected := ctx.Inject("primaryColor", nil); injected != nil {
    primaryColor = injected.(lipgloss.Color)
}

// Enhanced: UseTheme helper (new way)
theme := ctx.UseTheme(DefaultTheme)
// Use theme.Primary
```

### 2. Fluent Builder API
```go
// Old: Repetitive
.WithKeyBinding("up", "increment", "Increment").
.WithKeyBinding("k", "increment", "Increment").
.WithKeyBinding("+", "increment", "Increment")

// New: Concise
.WithKeyBindings("increment", "Increment", "up", "k", "+")
```

### 3. Factory Pattern
```go
// Define once, use everywhere
var UseSharedCounter = CreateShared(func(ctx *Context) *Counter {
    return UseCounter(ctx, 0)
})

// Component A
counter := UseSharedCounter(ctx)

// Component B (same instance)
counter := UseSharedCounter(ctx)
```

## Implementation Strategy

### Phase 1: Core Infrastructure (2 hours)
1. Create `theme.go` with Theme struct and helpers
2. Add UseTheme/ProvideTheme methods to Context
3. Add WithKeyBindings method to ComponentBuilder
4. Create `composables/shared.go` with CreateShared

### Phase 2: Testing (2-3 hours)
1. Unit tests for theme injection
2. Unit tests for multi-key bindings
3. Unit tests for shared composables
4. Integration tests across features
5. Thread safety tests (race detector)
6. Performance benchmarks

### Phase 3: Example Migration (2 hours)
1. Migrate 2-3 examples to new patterns
2. Verify output identical before/after
3. Measure code reduction
4. Document migration process

### Phase 4: Documentation (1-2 hours)
1. Update AI Manual with new patterns
2. Create migration guide
3. Add godoc for all new APIs
4. Update component reference guide

## Performance Considerations

### Theme System
- **Memory**: Single Theme struct (7 × lipgloss.Color = ~56 bytes)
- **CPU**: Type assertion on injection (<1μs)
- **Optimization**: No reflection, direct struct access

### Multi-Key Binding
- **Memory**: No additional overhead (same as individual bindings)
- **CPU**: O(n) loop during registration, O(1) lookup during runtime
- **Optimization**: Loop is one-time cost at component creation

### Shared Composable
- **Memory**: One instance instead of N instances (memory savings!)
- **CPU**: sync.Once overhead (~100ns first call, ~10ns subsequent)
- **Optimization**: Thread-safe without mutex on hot path

### Benchmark Targets
```go
BenchmarkThemeInjection     10000000    100 ns/op    0 B/op    0 allocs/op
BenchmarkMultiKeyBinding     1000000   1000 ns/op    0 B/op    0 allocs/op
BenchmarkSharedComposable   50000000     30 ns/op    0 B/op    0 allocs/op
```

## Error Handling

### Theme System
- **No parent theme**: Use provided default (graceful degradation)
- **Invalid type assertion**: Use provided default (ignore invalid)
- **Nil theme**: Not possible (Theme is struct, not pointer)

### Multi-Key Binding
- **Empty keys list**: No-op, return builder unchanged
- **Duplicate keys**: Last registration wins (same as WithKeyBinding)
- **Invalid key string**: Handled by existing WithKeyBinding logic

### Shared Composable
- **Concurrent initialization**: Protected by sync.Once
- **Nil factory**: Panic (developer error, caught in tests)
- **Factory panic**: Propagates to caller (expected Go behavior)

## Backward Compatibility

### Breaking Changes
**NONE** - All additions to API, no modifications to existing behavior

### Deprecated APIs
**NONE** - Old patterns remain fully supported

### Migration Path
- **Phase 1**: New APIs available, old code works
- **Phase 2**: Examples show new patterns (docs updated)
- **Phase 3**: Developers adopt new patterns gradually
- **Forever**: Both patterns coexist (developer choice)

## Visual Design Notes

### Theme Color Semantics
- **Primary**: Brand color, important actions, active elements
- **Secondary**: Alternative actions, secondary importance
- **Muted**: Disabled states, less important text, subtle UI
- **Warning**: Caution, non-blocking issues, informational alerts
- **Error**: Critical issues, blocking errors, danger actions
- **Success**: Completed actions, positive feedback, confirmations
- **Background**: Container backgrounds, card backgrounds

### Lipgloss Integration
```go
// Theme colors work seamlessly with Lipgloss
theme := ctx.UseTheme(DefaultTheme)

// Use in styles
titleStyle := lipgloss.NewStyle().
    Foreground(theme.Primary).
    Bold(true)

borderStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(theme.Secondary)

errorStyle := lipgloss.NewStyle().
    Foreground(theme.Error).
    Bold(true)
```

## Testing Strategy

### Unit Test Coverage
- Theme struct initialization
- UseTheme with provided theme
- UseTheme with default theme
- ProvideTheme propagation
- WithKeyBindings variadic handling
- CreateShared singleton behavior
- CreateShared thread safety

### Integration Test Scenarios
1. **Theme Hierarchy**: Parent → Child → Grandchild
2. **Theme Override**: Parent provides, child overrides locally
3. **Multi-Key Event**: Press different keys, same event emitted
4. **Shared State**: Multiple components, same composable instance
5. **Mixed Patterns**: Old inject/expose + new UseTheme in same app

### Edge Cases to Test
- No parent provides theme → default used
- Theme provided as interface{} (invalid) → default used
- 10+ keys bound to same event → all work
- 100 concurrent CreateShared calls → only 1 initialization
- Shared composable across 50 components → same instance

## Documentation Requirements

### Godoc Comments
- All exported types have comprehensive package docs
- All exported functions have usage examples
- All exported methods have parameter descriptions
- Performance characteristics documented where relevant

### Examples
- Before/after comparisons showing code reduction
- Real-world usage in example applications
- Common patterns and anti-patterns
- Migration guides with step-by-step instructions

### API Reference
- Complete API surface in AI Manual
- Type definitions with field descriptions
- Method signatures with return types
- Error conditions and edge cases

## Dependencies

### Internal
- `pkg/bubbly/context.go` - Provide/Inject methods
- `pkg/bubbly/component_builder.go` - Builder pattern
- `pkg/bubbly/component.go` - Component interface

### External
- `github.com/charmbracelet/lipgloss` - Color type
- `sync` - Once, Mutex (standard library)
- Go 1.22+ - Generics for CreateShared

### Build Requirements
- No new external dependencies
- No build tag requirements
- Works on all platforms (Linux, macOS, Windows)

## Security Considerations

### Theme System
- No user input processed
- No file I/O or network calls
- Type-safe struct access only
- No reflection or unsafe operations

### Multi-Key Binding
- Key strings validated by existing WithKeyBinding
- No command injection risk (Bubbletea handles keys)
- No untrusted input processing

### Shared Composable
- Thread-safe via sync.Once
- No race conditions possible
- Deterministic initialization order
- No shared mutable state (unless composable creates it)

## Accessibility Considerations

### Theme System
- Semantic color names improve understandability
- Consistent color usage across components
- Overridable for accessibility themes (high contrast, colorblind)

### Multi-Key Binding
- Alternative keys for same action (vim/arrow/numpad)
- Help text shows all available keys
- Consistent across components

## Future Extension Points

### Theme System
- **Custom theme colors**: Embed Theme in larger struct
- **Dynamic themes**: Change theme at runtime
- **Theme variants**: Light/dark mode switching
- **Theme inheritance**: Partial override of parent theme

### Multi-Key Binding
- **Conditional bindings**: Mode-based key switching
- **Key combination**: Support "ctrl+alt+k" style combos
- **Dynamic bindings**: Change bindings at runtime

### Shared Composable
- **Scoped sharing**: Share within subtree, not globally
- **Lazy initialization**: Defer creation until first access
- **Reset mechanism**: Clear singleton and reinitialize

**Note**: These extensions are out of scope for current implementation but architecturally supported.
