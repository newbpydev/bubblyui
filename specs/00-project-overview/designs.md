# Design Specification: BubblyUI Project Overview

## System Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     User Application                          │
│  (Uses BubblyUI components, composables, directives)         │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                        BubblyUI Framework                      │
├────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Built-in       │  │   Directives    │  │ Composition  │ │
│  │   Components     │  │   (If, ForEach) │  │    API       │ │
│  │   (Feature 06)   │  │   (Feature 05)  │  │ (Feature 04) │ │
│  └─────────┬────────┘  └────────┬────────┘  └──────┬───────┘ │
│            └────────────────────┼──────────────────┘         │
│  ┌──────────────────────────────┴────────────────────────┐   │
│  │              Component Model (Feature 02)             │   │
│  │  (Builder API, Props, Events, Template Rendering)    │   │
│  └───────────────────────────┬──────────────────────────┘   │
│  ┌────────────────────────────┴───────────────────────────┐ │
│  │           Lifecycle Hooks (Feature 03)                 │ │
│  │  (onMounted, onUpdated, onUnmounted, Cleanup)         │ │
│  └───────────────────────────┬──────────────────────────┘   │
│  ┌────────────────────────────┴───────────────────────────┐ │
│  │         Reactivity System (Feature 01)                 │ │
│  │  (Ref[T], Computed[T], Watch, Dependency Tracking)    │ │
│  └───────────────────────────┬──────────────────────────┘   │
└────────────────────────────────┼──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                      Bubbletea Framework                      │
│  (Init, Update, View, Cmd, Msg, Message Loop)                │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                   Terminal / Console                          │
└──────────────────────────────────────────────────────────────┘
```

## Feature Integration Map

### Layer 1: Foundation
```
00-project-setup
    ↓
    Provides: Build tools, testing, linting, CI/CD
    Enables: All development and deployment
```

### Layer 2: Core Systems
```
01-reactivity-system
    ↓
    Provides: Ref[T], Computed[T], Watch, Dependency tracking
    Enables: Reactive state management for components
    Exports: NewRef[T](), NewComputed[T](), Watch()
    
02-component-model
    ↓
    Uses: Reactivity system for state management
    Provides: Component, Context, RenderContext, Builder API
    Enables: Component creation, props, events, templates
    Exports: NewComponent(), Component interface
    
03-lifecycle-hooks
    ↓
    Uses: Component model, Reactivity system
    Provides: Lifecycle management, cleanup automation
    Enables: Side effects, resource management
    Exports: OnMounted, OnUpdated, OnUnmounted, OnCleanup
```

### Layer 3: Advanced Features
```
04-composition-api
    ↓
    Uses: Reactivity, Components, Lifecycle
    Provides: Composables, provide/inject, logic reuse
    Enables: Shared logic patterns
    Exports: UseState, UseAsync, UseEffect, Provide, Inject
    
05-directives
    ↓
    Uses: Components, Reactivity
    Provides: Template helpers, conditional rendering
    Enables: Declarative templates
    Exports: If(), ForEach(), Show(), Bind(), On()
```

### Layer 4: Component Library
```
06-built-in-components
    ↓
    Uses: ALL framework features
    Provides: Production-ready components
    Enables: Rapid application development
    Exports: Button, Input, Form, Table, Layout, etc.
```

## Data Flow Architecture

### Reactive State Flow
```
┌─────────────────┐
│  User Action    │
│ (keypress, etc) │
└────────┬────────┘
         ↓
┌────────────────────┐
│  tea.Msg           │
│  (Bubbletea)       │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ model.Update()     │
│ (Wrapper Model)    │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ component.Emit()   │
│ (Event Bridge)     │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ Event Handler      │
│ (Component Logic)  │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ Ref.Set()          │
│ (State Change)     │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ Watchers Triggered │
│ (Side Effects)     │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ component.Update() │
│ (Process Message)  │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ onUpdated Hooks    │
│ (Lifecycle)        │
└────────┬───────────┘
         ↓
┌────────────────────┐
│ View() / Render    │
│ (UI Update)        │
└────────────────────┘
```

### Component Hierarchy Flow
```
Application
    ↓
AppLayout (Template)
    ↓
┌───────────────────┬───────────────────┐
│                   │                   │
Header          Content              Footer
(Organism)      (Organism)          (Organism)
    ↓               ↓                   ↓
Navigation      DataTable           StatusBar
(Molecule)      (Organism)          (Molecule)
    ↓               ↓                   ↓
NavButton       TableRow            Text
(Atom)          (Molecule)          (Atom)
                    ↓
               ┌────┴────┐
               │         │
            TableCell  Button
            (Atom)     (Atom)
```

## Type System Architecture

### Generic Type Hierarchy
```go
// Core Generic Types
Ref[T any]               // Reactive reference to value of type T
Computed[T any]          // Computed value of type T
Watch[T any]             // Watch source of type T

// Component Types
Component                // Interface for all components
Props[T any]            // Generic props of type T
EventHandler             // func(data interface{})
RenderContext            // Template rendering context

// Composable Types
Composable[T any]       // Composable returning type T
Provider[T any]         // Provides value of type T
Injector[T any]         // Injects value of type T
```

### Type Safety Flow
```
1. Component Creation (Compile Time)
   - Props type checked via generics
   - Event payload types validated
   - Template function signature verified

2. State Management (Compile Time + Runtime)
   - Ref[T] enforces type at creation
   - Computed[T] enforces return type
   - Watchers type-check callbacks

3. Runtime Assertions (Where Necessary)
   - RenderContext.Get() returns interface{}
   - Type assertions in template (caught in tests)
   - Event data type assertions (caught in tests)
```

## Integration Patterns

### Pattern 1: Bubbletea Bridge (Current)
```go
// Manual bridge between BubblyUI and Bubbletea

type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd {
    return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        m.component.Emit("keypress", msg)
    case customMsg:
        m.component.Emit("custom-event", msg)
    }
    
    updatedComponent, cmd := m.component.Update(msg)
    m.component = updatedComponent.(bubbly.Component)
    return m, cmd
}

func (m model) View() string {
    return m.component.View()
}
```

### Pattern 2: Reactive Bridge (Future - Phase 4)
```go
// Automatic bridge - state changes generate commands

// In Context.Ref():
ref := NewRef(value)
ref.onChange = func() tea.Cmd {
    return func() tea.Msg {
        return StateChangedMsg{
            ComponentID: ctx.component.id,
            RefID: ref.id,
        }
    }
}

// Component runtime handles command batching and UI updates
// NO manual Emit() calls needed!
```

### Pattern 3: Component Composition
```go
// Parent component with children
parent := NewComponent("Parent").
    Setup(func(ctx *Context) {
        // Parent state
        parentState := ctx.Ref("value")
        
        // Child component
        child := NewComponent("Child").
            Props(ChildProps{
                Data: parentState.Get(),
            }).
            Build()
        
        // Listen to child events
        ctx.On("child-event", func(data interface{}) {
            // Handle child event
            parentState.Set(data)
        })
    }).
    Children(child).
    Build()
```

### Pattern 4: Composable Logic
```go
// Shared logic via composables
func useCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
    count := ctx.Ref(initial)
    
    increment := func() {
        count.Set(count.Get().(int) + 1)
    }
    
    decrement := func() {
        count.Set(count.Get().(int) - 1)
    }
    
    return count, increment, decrement
}

// Usage in component
ctx.Setup(func(ctx *Context) {
    count, inc, dec := useCounter(ctx, 0)
    
    ctx.On("increment", func(_ interface{}) { inc() })
    ctx.On("decrement", func(_ interface{}) { dec() })
})
```

## State Management Architecture

### State Scope Levels
```
1. Local Component State
   - Ref[T] created in component's Setup()
   - Scoped to single component instance
   - Cleaned up on unmount

2. Shared State (Provide/Inject)
   - Provider component creates state
   - Child components inject state
   - Scoped to component subtree

3. Global State (Future)
   - Application-level state
   - Accessible from any component
   - Persistent across component lifecycle
```

### State Update Lifecycle
```
1. State Change Initiated
   ↓
2. Ref.Set() called
   ↓
3. Internal value updated
   ↓
4. Watchers notified (sync)
   ↓
5. onUpdated hooks queued
   ↓
6. Next Update() cycle
   ↓
7. onUpdated hooks execute
   ↓
8. View() re-renders
```

## Lifecycle Management Architecture

### Component Lifecycle Phases
```
Creation Phase:
   NewComponent() → Setup() → Build()
   
Initialization Phase:
   Init() → onMounted hooks → First View()
   
Update Phase (Repeated):
   Update(msg) → onUpdated hooks → View()
   
Cleanup Phase:
   onBeforeUnmount → Unmount() → onUnmounted → Cleanup functions
```

### Resource Cleanup Strategy
```
Automatic Cleanup:
   - Watchers registered via ctx.Watch()
   - Event handlers registered via ctx.On()
   - Cleanup functions registered via ctx.OnCleanup()
   
Manual Cleanup:
   - Goroutines (use tea.Cmd instead)
   - File handles (register cleanup)
   - Network connections (register cleanup)
   - Timers/tickers (register cleanup)
```

## Error Handling Architecture

### Error Reporting Levels
```
1. Development (Tests)
   - Panics surface immediately
   - Full stack traces
   - Test failures on errors

2. Production (Observability Integration)
   - Panic recovery with error reporting
   - Context-rich error logging
   - Graceful degradation

3. User-Facing
   - Friendly error messages
   - Actionable guidance
   - Error state components
```

### Error Recovery Pattern
```go
defer func() {
    if r := recover(); r != nil {
        if reporter := observability.GetErrorReporter(); reporter != nil {
            reporter.ReportPanic(&observability.HandlerPanicError{
                ComponentName: componentName,
                EventName:     eventName,
                PanicValue:    r,
            }, &observability.ErrorContext{
                ComponentName: componentName,
                ComponentID:   componentID,
                EventName:     eventName,
                Timestamp:     time.Now(),
                StackTrace:    debug.Stack(),
            })
        }
    }
}()
```

## Performance Architecture

### Optimization Strategies
```
1. Reactive System
   - Lazy computed evaluation
   - Dependency caching
   - Minimal watcher overhead

2. Component Rendering
   - Template function caching (future)
   - Conditional re-renders
   - Child update batching

3. Memory Management
   - Weak references for cleanup
   - GC-friendly component trees
   - Resource pooling (future)

4. Concurrency
   - Per-goroutine dependency tracking
   - Lock-free reads where possible
   - Batched updates
```

### Performance Benchmarks
```go
// Target Performance (Framework Overhead)
BenchmarkRefGet              1000000000   1.2 ns/op
BenchmarkRefSet                10000000  90.5 ns/op
BenchmarkComputed               5000000  250  ns/op
BenchmarkComponentRender        1000000  4500 ns/op
BenchmarkFullUpdate             500000   8000 ns/op

// vs Raw Bubbletea
Raw Bubbletea Update:           7200 ns/op
BubblyUI Update:                8000 ns/op
Overhead:                       ~11% ✅ (< 15% target)
```

## Testing Architecture

### Test Pyramid
```
┌─────────────────┐
│   E2E Tests     │  Example applications
│   (Manual/Auto) │  Real-world scenarios
├─────────────────┤
│ Integration     │  Feature-to-feature
│    Tests        │  Component composition
├─────────────────┤
│   Unit Tests    │  Individual functions
│  (>80% coverage)│  Table-driven tests
└─────────────────┘
```

### Test Categories
```
1. Unit Tests
   - Ref, Computed, Watch behavior
   - Component creation and props
   - Event emission and handling
   - Lifecycle hook execution

2. Integration Tests
   - Reactivity + Components
   - Lifecycle + State management
   - Full feature integration

3. Concurrency Tests
   - Race detector enabled
   - High concurrency scenarios
   - Memory leak detection

4. E2E Tests
   - Example applications
   - User workflows
   - Performance under load
```

## Module Organization

### Package Structure
```
github.com/newbpydev/bubblyui/
├── pkg/
│   └── bubbly/           # Core framework
│       ├── ref.go        # Reactive primitives
│       ├── computed.go   # Computed values
│       ├── watch.go      # Watchers
│       ├── component.go  # Component implementation
│       ├── builder.go    # Builder API
│       ├── lifecycle.go  # Lifecycle management
│       ├── events.go     # Event system
│       ├── context.go    # Setup/Render context
│       ├── composables/  # Built-in composables
│       ├── directives/   # Built-in directives
│       ├── components/   # Built-in components
│       └── observability/# Error reporting
├── cmd/
│   └── examples/         # Example applications
│       ├── 01-reactivity-system/
│       ├── 02-component-model/
│       ├── 03-lifecycle-hooks/
│       └── ...
├── specs/                # Feature specifications
├── docs/                 # Documentation
└── internal/             # Internal utilities
```

### Import Organization
```go
// External dependencies
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Internal packages
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/components"
)

// Types
import (
    // Type definitions
)
```

## Known Limitations & Solutions

### Limitation 1: Manual Bubbletea Bridge
**Problem**: Components require wrapper model with manual event emission  
**Current Design**: Model wraps component, forwards messages via Emit()  
**Solution Design**: Automatic command generation from state changes (Phase 4)  
**Benefits**: Vue-like DX, no manual bridge code  
**Priority**: HIGH - Phase 4 enhancement

### Limitation 2: Template Type Safety
**Problem**: RenderContext.Get() returns interface{}, requires type assertion  
**Current Design**: Type assertions in template functions  
**Solution Design**: Code generation or generics (Go 1.23+)  
**Benefits**: Compile-time type checking  
**Priority**: MEDIUM - acceptable with good tests

### Limitation 3: Global Dependency Tracker Contention
**Problem**: Single global tracker causes deadlocks at high concurrency  
**Current Design**: Mutex-protected global DepTracker  
**Solution Design**: Per-goroutine tracking with goroutine-local storage  
**Benefits**: No contention, scales to 1000+ goroutines  
**Priority**: HIGH - must fix before production

### Limitation 4: Cannot Watch Computed Values
**Problem**: Watch() only accepts *Ref[T], not *Computed[T]  
**Current Design**: Watch interface constraint  
**Solution Design**: Watchable[T] interface, implement for both  
**Benefits**: Vue 3 compatibility, better composability  
**Priority**: MEDIUM - workarounds exist

## Future Enhancements

### Phase 4: Ecosystem
1. **Automatic Reactive Bridge**: State changes → tea.Cmd automatically
2. **Router System**: Multi-screen navigation with history
3. **Dev Tools**: Component inspector, state viewer
4. **Theme System**: Consistent styling and customization
5. **Performance Profiler**: Rendering performance analysis

### Phase 5: Community
1. **Plugin System**: Third-party component ecosystem
2. **Component Marketplace**: Shareable components
3. **Visual Builder**: Drag-and-drop component composer
4. **Testing Utilities**: Component test helpers

## Documentation Architecture

### Documentation Levels
```
1. API Documentation (godoc)
   - Every exported type, function, method
   - Examples for complex APIs
   - Links to guides

2. Conceptual Guides
   - How reactivity works
   - Component lifecycle explained
   - Composition patterns

3. Tutorial Guides
   - Getting started
   - Building first component
   - Advanced patterns

4. Reference Guides
   - API reference
   - Built-in components
   - Composables library

5. Example Applications
   - Counter (basic)
   - Todo list (intermediate)
   - Dashboard (advanced)
```

## Migration Strategy

### From Raw Bubbletea
```
Phase 1: Add BubblyUI alongside existing code
Phase 2: Migrate state to Ref/Computed
Phase 3: Convert models to components
Phase 4: Add lifecycle hooks
Phase 5: Extract composables
Phase 6: Remove old model code
```

### Compatibility Guarantees
```
- BubblyUI components implement tea.Model
- Compatible with Bubbles components
- Can mix BubblyUI and raw Bubbletea
- Gradual migration path
- No breaking changes in minor versions
```

## Deployment Architecture

### Build Pipeline
```
1. Code → gofmt → goimports → golangci-lint
2. Tests → go test -race -cover
3. Build → go build
4. Benchmark → go test -bench
5. Release → semantic version tag
```

### CI/CD
```
- GitHub Actions for CI
- Automated tests on PR
- Coverage reports
- Benchmark comparisons
- Automated releases
```

## Summary

BubblyUI provides a comprehensive, type-safe, Vue-inspired framework for building terminal applications in Go. The architecture balances familiar web development patterns with Go's strengths, while maintaining seamless integration with the battle-tested Bubbletea framework. Through careful layering of features, clear separation of concerns, and systematic design, BubblyUI enables developers to build sophisticated TUI applications with excellent developer experience and production-ready quality.
