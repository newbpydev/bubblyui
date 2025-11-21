# User Workflow: Advanced Internal Package Automation

## Feature ID
13-adv-internal-package-automation

## Primary User Journey: Theme System Adoption

### Entry Point
Developer creating a new BubblyUI application with multiple components that need consistent colors.

### Workflow Steps

#### Step 1: Define Theme in Parent Component
**User Action**: Create app component and provide theme to descendants

```go
func CreateApp() (bubbly.Component, error) {
    return bubbly.NewComponent("App").
        Setup(func(ctx *bubbly.Context) {
            // Define custom theme
            theme := bubbly.DefaultTheme
            theme.Primary = lipgloss.Color("99")    // Purple brand color
            theme.Secondary = lipgloss.Color("120") // Custom accent
            
            // ONE LINE: Provide to all descendants
            ctx.ProvideTheme(theme)
            
            // Create child components...
        }).
        Build()
}
```

**System Response**: Theme stored in Context's dependency injection system

**UI Update**: No immediate visual change (setup phase)

**Developer Experience**: 
- Clear, semantic API
- One line replaces 5+ individual Provide calls
- Type-safe theme struct

#### Step 2: Consume Theme in Child Component
**User Action**: Access theme colors in child component

```go
// OLD WAY (15 lines):
func CreateCard(props CardProps) (bubbly.Component, error) {
    return bubbly.NewComponent("Card").
        Setup(func(ctx *bubbly.Context) {
            // Inject with defaults
            primaryColor := lipgloss.Color("35")
            if injected := ctx.Inject("primaryColor", nil); injected != nil {
                primaryColor = injected.(lipgloss.Color)
            }
            
            secondaryColor := lipgloss.Color("99")
            if injected := ctx.Inject("secondaryColor", nil); injected != nil {
                secondaryColor = injected.(lipgloss.Color)
            }
            
            mutedColor := lipgloss.Color("240")
            if injected := ctx.Inject("mutedColor", nil); injected != nil {
                mutedColor = injected.(lipgloss.Color)
            }
            
            // Expose for template
            ctx.Expose("primaryColor", primaryColor)
            ctx.Expose("secondaryColor", secondaryColor)
            ctx.Expose("mutedColor", mutedColor)
        }).
        Build()
}

// NEW WAY (1 line):
func CreateCard(props CardProps) (bubbly.Component, error) {
    return bubbly.NewComponent("Card").
        Setup(func(ctx *bubbly.Context) {
            // ONE LINE: Get theme or use default
            theme := ctx.UseTheme(bubbly.DefaultTheme)
            
            // Use directly in template
            ctx.Expose("theme", theme)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            theme := ctx.Get("theme").(bubbly.Theme)
            
            // Type-safe color access
            titleStyle := lipgloss.NewStyle().
                Foreground(theme.Primary).
                Bold(true)
            
            borderStyle := lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(theme.Secondary)
            
            // ...
        }).
        Build()
}
```

**System Response**: 
- Check parent for theme injection
- Return injected theme or default
- Type-safe struct access

**UI Update**: Component renders with correct colors

**Developer Experience**:
- **94% code reduction** (15 lines → 1 line)
- **Type safety**: No type assertions in template
- **Clear intent**: "Use theme" vs manual inject/expose
- **Fallback**: Default theme if parent doesn't provide

#### Step 3: Override Theme Locally (Optional)
**User Action**: Component wants to modify theme for its subtree

```go
func CreateModal(props ModalProps) (bubbly.Component, error) {
    return bubbly.NewComponent("Modal").
        Setup(func(ctx *bubbly.Context) {
            // Get parent theme
            parentTheme := ctx.UseTheme(bubbly.DefaultTheme)
            
            // Override for modal (darker colors)
            modalTheme := parentTheme
            modalTheme.Background = lipgloss.Color("232") // Darker
            modalTheme.Muted = lipgloss.Color("238")
            
            // Provide modified theme to modal's children
            ctx.ProvideTheme(modalTheme)
            
            // Create child components...
        }).
        Build()
}
```

**System Response**: Modified theme provided to descendants only

**UI Update**: Modal and its children use darker colors

**Developer Experience**:
- Easy theme customization
- Scoped to component subtree
- Maintains parent theme elsewhere

### Completion
Developer has consistent, maintainable theming across entire application with minimal boilerplate.

---

## Alternative Journey: Multi-Key Binding Workflow

### Entry Point
Developer wants users to increment/decrement with multiple key options (vim keys, arrows, numpad).

### Workflow Steps

#### Step 1: Define Multi-Key Bindings
**User Action**: Register multiple keys for same action

```go
// OLD WAY (6 lines):
func CreateCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        WithKeyBinding("up", "increment", "Increment counter").
        WithKeyBinding("k", "increment", "Increment counter").
        WithKeyBinding("+", "increment", "Increment counter").
        WithKeyBinding("down", "decrement", "Decrement counter").
        WithKeyBinding("j", "decrement", "Decrement counter").
        WithKeyBinding("-", "decrement", "Decrement counter").
        Setup(func(ctx *bubbly.Context) {
            // ...
        }).
        Build()
}

// NEW WAY (2 lines):
func CreateCounter() (bubbly.Component, error) {
    return bubbly.NewComponent("Counter").
        WithKeyBindings("increment", "Increment counter", "up", "k", "+").
        WithKeyBindings("decrement", "Decrement counter", "down", "j", "-").
        Setup(func(ctx *bubbly.Context) {
            // ...
        }).
        Build()
}
```

**System Response**: 
- Loop over keys
- Register each key with same event and description
- Store in keyBindings map

**UI Update**: Help text shows all keys

**Developer Experience**:
- **67% code reduction** (6 lines → 2 lines)
- **Clear intent**: "These keys do the same thing"
- **Maintainability**: Easy to add/remove keys
- **Consistency**: Same description for all keys

#### Step 2: Handle Events (Unchanged)
**User Action**: Event handler already handles the event

```go
Setup(func(ctx *bubbly.Context) {
    counter := composables.UseCounter(ctx, 0)
    
    // Same event handler for all keys
    ctx.On("increment", func(data interface{}) {
        counter.Increment()
    })
    
    ctx.On("decrement", func(data interface{}) {
        counter.Decrement()
    })
})
```

**System Response**: Any bound key triggers the handler

**UI Update**: Counter increments/decrements

**Developer Experience**:
- No changes to event handling
- Same API, less boilerplate
- Works with existing composables

### Completion
Developer has flexible key bindings with minimal code and clear intent.

---

## Alternative Journey: Shared Composable Workflow

### Entry Point
Developer wants to share counter state across multiple components without prop drilling.

### Workflow Steps

#### Step 1: Define Shared Composable
**User Action**: Create shared composable factory

```go
// composables/shared_counter.go
package composables

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// UseSharedCounter returns a singleton counter instance
// shared across all components that call it
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *CounterComposable {
        return UseCounter(ctx, 0) // Initial value: 0
    },
)
```

**System Response**: 
- Factory wrapped with sync.Once
- Instance created on first call
- Subsequent calls return same instance

**Developer Experience**:
- VueUse-inspired pattern (familiar to Vue developers)
- Type-safe with generics
- Clear naming convention (UseShared prefix)

#### Step 2: Use in Multiple Components
**User Action**: Call shared composable from different components

```go
// Component A - Display
func CreateCounterDisplay() (bubbly.Component, error) {
    return bubbly.NewComponent("CounterDisplay").
        Setup(func(ctx *bubbly.Context) {
            // Get shared instance
            counter := composables.UseSharedCounter(ctx)
            ctx.Expose("counter", counter)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            counter := ctx.Get("counter").(*composables.CounterComposable)
            return fmt.Sprintf("Count: %d", counter.Count.Get().(int))
        }).
        Build()
}

// Component B - Controls (same instance!)
func CreateCounterControls() (bubbly.Component, error) {
    return bubbly.NewComponent("CounterControls").
        WithKeyBindings("increment", "Increment", "up", "+").
        WithKeyBindings("decrement", "Decrement", "down", "-").
        Setup(func(ctx *bubbly.Context) {
            // Same shared instance as Component A
            counter := composables.UseSharedCounter(ctx)
            
            ctx.On("increment", func(data interface{}) {
                counter.Increment() // Updates display in Component A!
            })
            
            ctx.On("decrement", func(data interface{}) {
                counter.Decrement()
            })
        }).
        Build()
}
```

**System Response**: 
- Both components receive same counter instance
- State changes in one visible in other
- Automatic reactivity (BubblyUI reactive system)

**UI Update**: 
- Display component shows current count
- Control component increments/decrements
- Display updates automatically

**Developer Experience**:
- No prop drilling through component hierarchy
- No global variables or singletons to manage
- Type-safe shared state
- Works with BubblyUI reactivity

#### Step 3: Verify Shared State
**User Action**: Test that state is actually shared

```go
func TestSharedCounter(t *testing.T) {
    harness := testutil.NewHarness(t)
    defer harness.Cleanup()
    
    // Mount both components
    display := harness.Mount(CreateCounterDisplay())
    controls := harness.Mount(CreateCounterControls())
    
    // Increment in controls
    controls.SendEvent("increment", nil)
    
    // Verify display updated
    display.AssertRenderContains("Count: 1")
}
```

**System Response**: Test passes (shared state verified)

**Developer Experience**:
- Easy to test shared behavior
- Clear mental model
- Predictable state sharing

### Completion
Developer has elegant shared state solution without complex state management libraries.

---

## Migration Workflow: Existing Code → New Patterns

### Entry Point
Developer has existing BubblyUI app and wants to adopt new automation patterns.

### Migration Steps

#### Step 1: Identify Opportunities
**User Action**: Scan codebase for patterns to automate

```bash
# Find theme inject/expose patterns
grep -r "ctx.Inject.*Color" --include="*.go"

# Find multiple key bindings
grep -r "WithKeyBinding.*increment" --include="*.go"

# Find potential shared composables
grep -r "UseCounter\|UseForm\|UseState" --include="*.go"
```

**System Response**: List of files with automatable patterns

**Developer Experience**: 
- Quick identification of opportunities
- Clear ROI (lines saved per file)

#### Step 2: Migrate Theme System (High Priority)
**User Action**: Replace inject/expose with UseTheme

**Before** (15 lines in every child component):
```go
primaryColor := lipgloss.Color("35")
if injected := ctx.Inject("primaryColor", nil); injected != nil {
    primaryColor = injected.(lipgloss.Color)
}
ctx.Expose("primaryColor", primaryColor)

secondaryColor := lipgloss.Color("99")
if injected := ctx.Inject("secondaryColor", nil); injected != nil {
    secondaryColor = injected.(lipgloss.Color)
}
ctx.Expose("secondaryColor", secondaryColor)

mutedColor := lipgloss.Color("240")
if injected := ctx.Inject("mutedColor", nil); injected != nil {
    mutedColor = injected.(lipgloss.Color)
}
ctx.Expose("mutedColor", mutedColor)
```

**After** (1 line):
```go
theme := ctx.UseTheme(bubbly.DefaultTheme)
ctx.Expose("theme", theme)
```

**System Response**: Code compiles, tests pass, output identical

**Developer Experience**:
- Immediate code reduction
- Improved readability
- No behavior change

#### Step 3: Migrate Key Bindings (Medium Priority)
**User Action**: Consolidate multiple WithKeyBinding calls

**Before** (6 lines):
```go
.WithKeyBinding("up", "increment", "Increment").
.WithKeyBinding("k", "increment", "Increment").
.WithKeyBinding("+", "increment", "Increment").
.WithKeyBinding("down", "decrement", "Decrement").
.WithKeyBinding("j", "decrement", "Decrement").
.WithKeyBinding("-", "decrement", "Decrement")
```

**After** (2 lines):
```go
.WithKeyBindings("increment", "Increment", "up", "k", "+").
.WithKeyBindings("decrement", "Decrement", "down", "j", "-")
```

**System Response**: Key bindings work identically

**Developer Experience**:
- Cleaner builder pattern
- Easier to maintain
- Clear grouping of related keys

#### Step 4: Identify Shared State Opportunities (Optional)
**User Action**: Find composables used in multiple components

```go
// Before: Each component creates own instance
// Component A
counter1 := composables.UseCounter(ctx, 0)

// Component B
counter2 := composables.UseCounter(ctx, 0) // Different instance!
```

**After**: Create shared version if state should be shared
```go
// Define once
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *composables.CounterComposable {
        return composables.UseCounter(ctx, 0)
    },
)

// Component A and B both use
counter := UseSharedCounter(ctx) // Same instance!
```

**System Response**: Components share state

**Developer Experience**:
- Opt-in architectural pattern
- Clear when state is shared vs isolated
- No prop drilling needed

#### Step 5: Run Tests and Verify
**User Action**: Ensure migration didn't break anything

```bash
# Run all tests
make test-race

# Run linter
make lint

# Check code formatting
make fmt

# Verify examples work
cd cmd/examples/10-testing/01-counter && go run .
```

**System Response**: All tests pass, no lint warnings, examples work

**Developer Experience**:
- Confidence in migration
- No regressions
- Identical user experience

### Completion
Developer has modern, automated BubblyUI codebase with significantly less boilerplate.

---

## Error Handling Flows

### Error 1: No Parent Provides Theme
**Trigger**: Child calls UseTheme() but no ancestor called ProvideTheme()

**User Sees**: Component uses default theme colors (graceful degradation)

**Recovery**: No recovery needed - this is expected behavior

**Developer Action**: 
- If custom theme desired, add ProvideTheme() to parent
- If default acceptable, no action needed

**System Behavior**:
```go
theme := ctx.UseTheme(bubbly.DefaultTheme)
// If no parent provides, returns DefaultTheme
// Component works normally with default colors
```

### Error 2: Invalid Type in Injection
**Trigger**: Theme provided as wrong type (interface{}, string, etc.)

**User Sees**: Component uses default theme (graceful degradation)

**Recovery**: Automatic - type assertion fails, default used

**Developer Action**: Fix ProvideTheme() call to use Theme struct

**System Behavior**:
```go
func (ctx *Context) UseTheme(defaultTheme Theme) Theme {
    if injected := ctx.Inject("theme", nil); injected != nil {
        if theme, ok := injected.(Theme); ok {
            return theme // Success
        }
        // Type assertion failed - use default
    }
    return defaultTheme
}
```

### Error 3: Empty Keys in WithKeyBindings
**Trigger**: Developer calls WithKeyBindings("event", "desc") with no keys

**User Sees**: No keys bound (no-op)

**Recovery**: N/A - harmless no-op

**Developer Action**: Add keys to the call or remove the call

**System Behavior**:
```go
func (b *ComponentBuilder) WithKeyBindings(event, description string, keys ...string) *ComponentBuilder {
    for _, key := range keys {
        // If keys is empty, loop never executes
        b.WithKeyBinding(key, event, description)
    }
    return b // Returns unchanged
}
```

### Error 4: Concurrent CreateShared Initialization
**Trigger**: Multiple goroutines call shared composable simultaneously

**User Sees**: Normal operation - single instance created

**Recovery**: N/A - sync.Once handles concurrency

**Developer Action**: None needed - thread-safe by design

**System Behavior**:
```go
func CreateShared[T any](factory func(*Context) T) func(*Context) T {
    var instance T
    var once sync.Once // Thread-safe initialization
    
    return func(ctx *Context) T {
        once.Do(func() {
            instance = factory(ctx) // Only first call executes
        })
        return instance
    }
}
```

---

## State Transitions

### Theme System State Machine
```
Initial State (Component created)
    ↓
Setup Phase
    ↓
ctx.UseTheme(default) called
    ↓
┌─────────────────────────┬─────────────────────────┐
│ Parent provides theme?  │                         │
├─────────────────────────┤                         │
│ YES                     │ NO                      │
│   ↓                     │   ↓                     │
│ Inject succeeds         │ Inject returns nil      │
│   ↓                     │   ↓                     │
│ Type assert to Theme    │ Use default theme       │
│   ↓                     │   ↓                     │
├─────────────────────────┴─────────────────────────┤
│ Return theme (provided or default)                │
│   ↓                                               │
│ Component uses theme in template                  │
│   ↓                                               │
│ Render with correct colors                        │
└───────────────────────────────────────────────────┘
```

### Shared Composable State Machine
```
Initial State (App starts)
    ↓
Component A calls UseSharedCounter(ctx)
    ↓
┌─────────────────────────────────────────────────┐
│ Is instance initialized?                        │
├─────────────────────────────────────────────────┤
│ NO                                              │
│   ↓                                             │
│ sync.Once.Do() → factory(ctx)                   │
│   ↓                                             │
│ Store instance in closure                       │
│   ↓                                             │
│ Mark as initialized                             │
└─────────────────────────────────────────────────┘
    ↓
Component B calls UseSharedCounter(ctx)
    ↓
┌─────────────────────────────────────────────────┐
│ Is instance initialized?                        │
├─────────────────────────────────────────────────┤
│ YES                                             │
│   ↓                                             │
│ sync.Once.Do() skipped                          │
│   ↓                                             │
│ Return existing instance                        │
└─────────────────────────────────────────────────┘
    ↓
Both components share same instance
```

---

## Integration Points

### With Existing Features

#### 1. Provide/Inject System
- **Connection**: UseTheme/ProvideTheme built on existing Context.Provide/Inject
- **Data Shared**: Theme struct flows through component tree
- **Integration**: Seamless - same underlying mechanism
- **Benefit**: Leverages proven DI system

#### 2. Key Binding System
- **Connection**: WithKeyBindings wraps existing WithKeyBinding
- **Data Shared**: Same keyBindings map, same event emission
- **Integration**: Transparent - no behavior change
- **Benefit**: Reduces boilerplate without new mechanisms

#### 3. Composables Pattern
- **Connection**: CreateShared wraps any composable factory
- **Data Shared**: Composable instance shared across components
- **Integration**: Works with all existing composables
- **Benefit**: Architectural pattern, not framework change

#### 4. Component Lifecycle
- **Connection**: Theme accessed in Setup, used in Template
- **Data Shared**: Theme available throughout component lifecycle
- **Integration**: No lifecycle hooks needed
- **Benefit**: Simple, stateless helper

#### 5. Bubbletea Integration
- **Connection**: Key bindings emit events, Bubbletea processes
- **Data Shared**: Key messages, event emission
- **Integration**: Zero changes to Bubbletea flow
- **Benefit**: Full compatibility maintained

### Navigation Between Features

```
Application Entry
    ↓
bubbly.Run(app) [08-automatic-reactive-bridge]
    ↓
App Component
    ├─ ctx.ProvideTheme(theme) [13-adv-internal-package-automation]
    ├─ WithKeyBindings(...) [13-adv-internal-package-automation]
    └─ UseSharedCounter(ctx) [13-adv-internal-package-automation]
    ↓
Child Components
    ├─ ctx.UseTheme(default) [13-adv-internal-package-automation]
    ├─ Event handlers with key bindings
    └─ Shared composable access
    ↓
Template Rendering [02-component-model]
    ├─ Use theme colors
    └─ Render with Lipgloss
    ↓
Bubbletea Update Loop
```

### Data Flow Between Components

```
Parent Component:
  theme = DefaultTheme (override Primary)
  ctx.ProvideTheme(theme)
      ↓
      │ Context.Provide("theme", theme)
      ↓
Child Component A:
  theme = ctx.UseTheme(DefaultTheme)
  // theme.Primary is parent's value
      ↓
Child Component B:
  theme = ctx.UseTheme(DefaultTheme)
  // Same theme as Component A
      ↓
Grandchild Component:
  theme = ctx.UseTheme(DefaultTheme)
  // Inherits from parent/grandparent chain
```

---

## Developer Personas

### Persona 1: Vue.js Developer (Familiar Pattern)
**Background**: Experience with Vue.js and VueUse
**Goal**: Build TUI app with familiar patterns
**Experience**:
- Recognizes createSharedComposable pattern immediately
- Understands Provide/Inject from Vue
- Comfortable with composition API style
- Adopts all patterns quickly

### Persona 2: Go Developer (Learning BubblyUI)
**Background**: Go experience, new to BubblyUI
**Goal**: Build maintainable TUI application
**Experience**:
- Appreciates simple, Go-idiomatic APIs
- UseTheme makes sense (no magic)
- CreateShared uses familiar sync.Once
- Prefers explicit over implicit

### Persona 3: Existing BubblyUI User (Migrating)
**Background**: Using BubblyUI before automation features
**Goal**: Reduce boilerplate in existing app
**Experience**:
- Sees immediate value (170+ lines saved)
- Migration path is clear
- Old code still works (backward compatible)
- Adopts patterns incrementally

### Persona 4: Component Library Author
**Background**: Creating reusable BubblyUI components
**Goal**: Consistent theming across components
**Experience**:
- UseTheme provides standard interface
- Components work with any theme
- Fallback to defaults ensures robustness
- Easy to document and maintain

---

## Success Metrics

### Quantitative
- **Code Reduction**: 170+ lines eliminated in examples
- **Adoption Rate**: 50%+ of new components use UseTheme
- **Migration Speed**: 1-2 hours to migrate medium app
- **Performance**: <5% overhead vs manual patterns

### Qualitative
- **Developer Satisfaction**: Positive feedback on clarity
- **Learning Curve**: New users understand patterns quickly
- **Maintenance**: Fewer support questions about inject/expose
- **Consistency**: More uniform color usage across apps
