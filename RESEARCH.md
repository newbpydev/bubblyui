# BubblyUI Research Document
## Vue-Inspired TUI Framework for Go

**Date:** October 25, 2025  
**Purpose:** Comprehensive research for building a Vue.js-inspired TUI framework on top of/alongside Bubbletea

---

## Executive Summary

BubblyUI aims to bring Vue.js developer experience to Go TUI development by:
- **Component-based architecture** with reusable, composable components
- **Declarative templates** with directive-like syntax
- **Reactive state management** using Go idioms
- **Composition API** pattern for logic reuse
- **Developer-friendly abstractions** while maintaining Go's philosophy

---

## 1. Bubbletea Architecture Analysis

### 1.1 The Elm Architecture (TEA)

Bubbletea implements The Elm Architecture with three core concepts:

**Model (State)**
```go
type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}
```

**Init (Initialization)**
```go
func (m model) Init() tea.Cmd {
    return nil // or return initial commands
}
```

**Update (State Transitions)**
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle messages
    }
    return m, nil
}
```

**View (Rendering)**
```go
func (m model) View() string {
    return "rendered output"
}
```

### 1.2 Key Patterns

- **Message-based updates**: All state changes via messages
- **Immutable updates**: Return new model state
- **Command pattern**: Async operations as `tea.Cmd`
- **Composition**: Models can contain other models
- **No goroutines**: Use Bubbletea's scheduler via commands

### 1.3 Challenges Identified

1. **Event propagation**: Manual delegation required
2. **Boilerplate**: Repetitive Update/View logic
3. **Component reuse**: No built-in abstraction
4. **State management**: Global state patterns unclear
5. **Testing**: Requires understanding message flow

---

## 2. Vue.js Paradigms to Adapt

### 2.1 Component System

**Single File Component Pattern (Vue)**
```vue
<template>
  <div>{{ count }}</div>
</template>

<script setup>
import { ref } from 'vue'
const count = ref(0)
</script>
```

**Go Adaptation Strategy:**
```go
type Component interface {
    tea.Model
    // Additional component lifecycle
    Props() any
    Emit(event string, data any)
}
```

### 2.2 Reactivity System

**Vue Approach:**
- `ref()`: Reactive primitives
- `reactive()`: Reactive objects
- `computed()`: Derived state
- `watch()`: Side effects on changes

**Go Adaptation:**
```go
// Using channels for reactivity
type Ref[T any] struct {
    value T
    watchers []chan T
}

func (r *Ref[T]) Set(val T) {
    r.value = val
    for _, w := range r.watchers {
        select {
        case w <- val:
        default:
        }
    }
}
```

### 2.3 Composition API

**Benefits:**
- Better code organization
- Logic reuse via composables
- Type safety
- Explicit dependencies

**Go Pattern:**
```go
// Composable function
func useCounter() (*Ref[int], func()) {
    count := NewRef(0)
    increment := func() {
        count.Set(count.Get() + 1)
    }
    return count, increment
}
```

### 2.4 Directives

**Vue directives:**
- `v-if`: Conditional rendering
- `v-for`: List rendering
- `v-model`: Two-way binding
- `v-on`: Event handling

**Go Adaptation:**
Use builder pattern:
```go
component.
    If(condition, renderFunc).
    ForEach(items, renderItem).
    On("keypress", handler)
```

### 2.5 Watchers

**Purpose:** React to state changes

**Go Pattern:**
```go
func (c *Component) Watch(source *Ref[T], callback func(T, T)) {
    // Register watcher
}
```

---

## 3. Go TUI Ecosystem

### 3.1 Core Libraries

**Bubbletea**
- Elm Architecture implementation
- Message loop & rendering
- Terminal handling
- Command scheduler

**Bubbles**
- Pre-built components (textinput, list, table, viewport, spinner, etc.)
- Component pattern examples
- Good architecture reference

**Lipgloss**
- Styling library (like CSS)
- Borders, colors, padding, margins
- Layout utilities (JoinHorizontal, JoinVertical)
- Adaptive colors for light/dark themes

**tview**
- Alternative widget-based approach
- More imperative style
- Rich component library

### 3.2 Comparison

| Feature | Bubbletea | tview |
|---------|-----------|-------|
| Paradigm | Functional (Elm) | OOP/Widget |
| Learning Curve | Moderate | Lower |
| Flexibility | High | Medium |
| Styling | Lipgloss | Built-in |
| Best For | Custom UIs | Standard UIs |

---

## 4. Go Framework Design Patterns

### 4.1 Clean Architecture

**Layers:**
1. **Domain**: Business logic, pure functions
2. **Application**: Use cases, orchestration
3. **Ports**: Interfaces
4. **Adapters**: Implementations

**Benefits:**
- Testability
- Maintainability
- Flexibility
- Clear separation of concerns

**BubblyUI Application:**
```
bubblyui/
├── core/           # Domain layer
│   ├── component.go
│   ├── reactivity.go
│   └── lifecycle.go
├── runtime/        # Application layer
│   ├── renderer.go
│   ├── scheduler.go
│   └── differ.go
├── adapters/       # Adapter layer
│   └── bubbletea/
└── components/     # Built-in components
```

### 4.2 Interface Design

**Principle:** Accept interfaces, return structs

```go
// Core interface
type Component interface {
    Render(ctx Context) string
    HandleMessage(msg Msg) (Component, Cmd)
}

// Builder interface
type ComponentBuilder interface {
    Props(props any) ComponentBuilder
    Children(...Component) ComponentBuilder
    Build() Component
}
```

### 4.3 Generics Usage

**Where to use:**
- Ref[T] for type-safe reactivity
- Props[T] for type-safe component props
- Event handlers with typed payloads

```go
type Ref[T any] struct {
    value T
}

func NewRef[T any](initial T) *Ref[T] {
    return &Ref[T]{value: initial}
}
```

### 4.4 Channel Patterns

**For reactivity:**
- Pub/sub for events
- Channels for watchers
- Context for cancellation

```go
type EventBus struct {
    subscribers map[string][]chan Event
}

func (eb *EventBus) Publish(event Event) {
    for _, ch := range eb.subscribers[event.Type] {
        ch <- event
    }
}
```

---

## 5. Testing Strategy (TDD)

### 5.1 Testing Levels

**Unit Tests**
- Component logic
- Reactivity system
- Helper functions

**Integration Tests**
- Component composition
- Message flow
- Event handling

**E2E Tests** (using tui-test)
- Full application flows
- Terminal interactions
- Visual regression

### 5.2 Table-Driven Tests

```go
func TestComponent_Render(t *testing.T) {
    tests := []struct {
        name     string
        props    Props
        expected string
    }{
        {
            name:     "renders with text",
            props:    Props{Text: "Hello"},
            expected: "Hello",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := NewComponent(tt.props)
            got := c.Render(Context{})
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

### 5.3 Mocking

**Use interfaces:**
```go
type Renderer interface {
    Render(component Component) string
}

// Mock in tests
type MockRenderer struct {
    RenderFunc func(Component) string
}
```

### 5.4 TUI Testing

**Microsoft tui-test**
- E2E terminal testing
- Cross-platform
- Snapshot testing
- Auto-wait capabilities

---

## 6. Component Model Design

### 6.1 Proposed Architecture

```go
// Component definition
type Component struct {
    // Static
    name  string
    props Props
    
    // Lifecycle
    setup    SetupFunc
    mounted  LifecycleHook
    updated  LifecycleHook
    unmounted LifecycleHook
    
    // Reactive
    state    map[string]*Ref[any]
    computed map[string]ComputedRef
    watchers []Watcher
    
    // Render
    template RenderFunc
    
    // Children
    children []Component
}

// Builder pattern
func NewComponent(name string) *ComponentBuilder {
    return &ComponentBuilder{
        component: &Component{name: name},
    }
}

func (cb *ComponentBuilder) Setup(fn SetupFunc) *ComponentBuilder {
    cb.component.setup = fn
    return cb
}

func (cb *ComponentBuilder) Template(fn RenderFunc) *ComponentBuilder {
    cb.component.template = fn
    return cb
}
```

### 6.2 Composition

```go
// Composable
type Composable func(*ComponentContext) any

func useCounter() Composable {
    return func(ctx *ComponentContext) any {
        count := ctx.Ref(0)
        increment := func() { count.Set(count.Get() + 1) }
        return struct {
            Count     *Ref[int]
            Increment func()
        }{count, increment}
    }
}

// Usage in component
c.Setup(func(ctx *ComponentContext) {
    counter := useCounter()(ctx)
    // Use counter.Count, counter.Increment
})
```

### 6.3 Props & Events

```go
// Type-safe props
type ButtonProps struct {
    Label    string
    OnClick  func()
    Disabled bool
}

// Type-safe events
type Component struct {
    emit func(event string, data any)
}

func (c *Component) Emit(event string, data any) {
    if c.emit != nil {
        c.emit(event, data)
    }
}
```

---

## 7. Reactivity Implementation

### 7.1 Core Primitives

```go
// Ref - reactive primitive
type Ref[T any] struct {
    mu       sync.RWMutex
    value    T
    watchers []func(T, T)
}

func (r *Ref[T]) Get() T {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.value
}

func (r *Ref[T]) Set(newVal T) {
    r.mu.Lock()
    oldVal := r.value
    r.value = newVal
    r.mu.Unlock()
    
    for _, watcher := range r.watchers {
        watcher(newVal, oldVal)
    }
}

// Computed - derived state
type Computed[T any] struct {
    fn    func() T
    cache T
    dirty bool
    deps  []Dependency
}

func (c *Computed[T]) Get() T {
    if c.dirty {
        c.cache = c.fn()
        c.dirty = false
    }
    return c.cache
}
```

### 7.2 Dependency Tracking

```go
type DepTracker struct {
    currentDeps []Dependency
    tracking    bool
}

var globalTracker = &DepTracker{}

func track(dep Dependency) {
    if globalTracker.tracking {
        globalTracker.currentDeps = append(globalTracker.currentDeps, dep)
    }
}
```

### 7.3 Watcher System

```go
func Watch[T any](source *Ref[T], callback func(T, T), options WatchOptions) func() {
    handler := func(newVal, oldVal T) {
        if options.Deep {
            // Deep comparison
        }
        callback(newVal, oldVal)
    }
    
    source.watchers = append(source.watchers, handler)
    
    if options.Immediate {
        callback(source.Get(), source.Get())
    }
    
    // Return cleanup function
    return func() {
        // Remove watcher
    }
}
```

---

## 8. Recommended Tech Stack

### 8.1 Core Dependencies

```go
// go.mod
module github.com/newbpydev/bubblyui

go 1.22

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.10.0
)

require (
    // Testing
    github.com/stretchr/testify v1.8.4
    golang.org/x/sync v0.6.0
)
```

### 8.2 Project Structure

```
bubblyui/
├── cmd/
│   └── examples/
│       ├── counter/
│       ├── todo/
│       └── dashboard/
├── pkg/
│   ├── bubbly/           # Core framework
│   │   ├── component.go
│   │   ├── reactivity.go
│   │   ├── lifecycle.go
│   │   ├── renderer.go
│   │   └── context.go
│   ├── directives/       # Built-in directives
│   │   ├── if.go
│   │   ├── for.go
│   │   └── model.go
│   ├── composables/      # Reusable logic
│   │   ├── use_state.go
│   │   ├── use_effect.go
│   │   └── use_async.go
│   └── components/       # Built-in components
│       ├── button/
│       ├── input/
│       ├── list/
│       └── layout/
├── internal/
│   ├── runtime/          # Internal runtime
│   ├── diff/             # Virtual DOM diff
│   └── scheduler/        # Update scheduler
├── examples/
├── docs/
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── go.mod
├── go.sum
├── README.md
└── RESEARCH.md
```

### 8.3 Development Tools

- **golangci-lint**: Linting
- **gotestsum**: Better test output
- **air**: Live reload
- **mockery**: Mock generation
- **godoc**: Documentation

---

## 9. Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- [ ] Core reactivity system (Ref, Computed)
- [ ] Basic component abstraction
- [ ] Lifecycle hooks
- [ ] Testing infrastructure
- [ ] Example: Simple counter

### Phase 2: Component System (Weeks 3-4)
- [ ] Props system
- [ ] Event system
- [ ] Component composition
- [ ] Slot/children support
- [ ] Example: Todo list

### Phase 3: Advanced Features (Weeks 5-6)
- [ ] Directives (v-if, v-for, v-model equivalents)
- [ ] Composables framework
- [ ] Context/provide-inject
- [ ] Router basics
- [ ] Example: Multi-page app

### Phase 4: Ecosystem (Weeks 7-8)
- [ ] Built-in components library
- [ ] Dev tools
- [ ] Documentation site
- [ ] More examples
- [ ] Performance optimization

---

## 10. Key Design Decisions

### 10.1 Enhance vs. Replace Bubbletea

**Decision: Enhance**

Reasons:
- Bubbletea is battle-tested
- Leverage existing ecosystem
- Lower barrier to adoption
- Incremental migration path

### 10.2 Syntax Style

**Decision: Builder pattern + functional options**

```go
NewComponent("Button").
    Props(ButtonProps{Label: "Click me"}).
    On("click", handleClick).
    Template(func(ctx Context) string {
        return lipgloss.NewStyle().Render(ctx.Props.Label)
    }).
    Build()
```

### 10.3 Reactivity Model

**Decision: Explicit refs with automatic tracking**

- Similar to Vue 3's Composition API
- Type-safe with generics
- Clear data flow
- Testable

### 10.4 Template System

**Decision: Go functions, not string templates**

Reasons:
- Type safety
- IDE support
- No parsing overhead
- Leverage Lipgloss fully

---

## 11. Best Practices

### 11.1 Don't Over-engineer

- Start simple
- Add complexity when needed
- Follow Go idioms
- Prioritize clarity over cleverness

### 11.2 Embrace Go's Strengths

- Interfaces for abstraction
- Struct composition
- Explicit error handling
- Channels for concurrency

### 11.3 Learn from Vue

- Component philosophy
- Reactivity patterns
- Developer experience
- API design

### 11.4 Testing First

- TDD from day one
- High test coverage
- Example-driven development
- Document through tests

---

## 12. Open Questions

1. **Virtual DOM**: Do we need diffing or just re-render?
2. **Performance**: Benchmark against pure Bubbletea
3. **SSR**: Server-side rendering for TUI?
4. **Plugins**: Plugin architecture needed?
5. **TypeScript**: Equivalent type safety?

---

## 13. Success Metrics

- **DX**: Faster component development
- **LOC**: Less boilerplate code
- **Learning**: Clear migration from Vue
- **Performance**: < 10% overhead vs Bubbletea
- **Adoption**: Community usage

---

## 14. References

### Documentation
- [Bubbletea Docs](https://github.com/charmbracelet/bubbletea)
- [Vue.js Documentation](https://vuejs.org)
- [The Elm Architecture](https://guide.elm-lang.org/architecture/)
- [Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/)

### Libraries
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [RxGo Reactivity](https://github.com/ReactiveX/RxGo)

### Testing
- [tui-test](https://github.com/microsoft/tui-test)
- [Testify](https://github.com/stretchr/testify)

---

## 15. Next Steps

1. **Prototype**: Build reactivity system
2. **Validate**: Create simple counter example
3. **Iterate**: Gather feedback
4. **Document**: Write architecture docs
5. **Release**: Alpha version with examples

---

*This research document will evolve as the project progresses.*
