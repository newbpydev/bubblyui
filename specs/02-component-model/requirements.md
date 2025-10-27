# Feature Name: Component Model

## Feature ID
02-component-model

## Overview
Implement a Vue-inspired component system that wraps Bubbletea's Model-Update-View pattern with a more declarative, composable API. Components encapsulate state (using reactive primitives from feature 01), behavior (event handlers), and presentation (template functions), providing a higher-level abstraction over raw Bubbletea models.

## User Stories
- As a **developer**, I want to create reusable components so that I can build complex UIs from smaller pieces
- As a **developer**, I want to pass data to components via props so that components are configurable and flexible
- As a **developer**, I want components to emit events so that child components can communicate with parents
- As a **developer**, I want a builder pattern API so that component creation is fluent and readable
- As a **developer**, I want components to integrate seamlessly with Bubbletea so that I can use the existing ecosystem
- As a **developer**, I want type-safe component APIs so that I catch errors at compile time

## Functional Requirements

### 1. Component Interface
1.1. Define `Component` interface that wraps `tea.Model`  
1.2. Support Init, Update, View methods (Bubbletea compatibility)  
1.3. Store reactive state (Refs from feature 01)  
1.4. Manage child components  
1.5. Handle props and events  

### 2. Component Builder (Fluent API)
2.1. Create components with `NewComponent(name string)`  
2.2. Chain methods: `Setup()`, `Props()`, `Template()`, `Children()`  
2.3. Build final component with `Build()`  
2.4. Validate configuration before building  
2.5. Type-safe method signatures  

### 3. Props System
3.1. Define props as type-safe structs  
3.2. Pass props during component creation  
3.3. Access props in setup and template functions  
3.4. Validate required props  
3.5. Support default values  
3.6. Immutable props (read-only from component perspective)  

### 4. Event System
4.1. Components can emit custom events  
4.2. Parent components can listen to child events  
4.3. Type-safe event payloads  
4.4. Event bubbling from child to parent  
4.5. Event handlers registered via `On(eventName, handler)`  
4.6. Event propagation control (stop bubbling)  
4.7. Event metadata includes source component and timestamp  

### 5. Template Rendering
5.1. Template as Go function: `func(ctx RenderContext) string`  
5.2. Access props, state, and computed values in template  
5.3. Use Lipgloss for styling  
5.4. Render child components  
5.5. Support conditional rendering and loops  

### 6. State Management
6.1. Components store Refs for reactive state  
6.2. State changes trigger re-renders  
6.3. State isolated per component instance  
6.4. Context for sharing state (provide/inject pattern)  

### 7. Component Composition
7.1. Nest components within other components  
7.2. Pass props to child components  
7.3. Listen to child events  
7.4. Manage child component lifecycle  

### 8. Bubbletea Integration
8.1. Component implements `tea.Model` interface  
8.2. Map component events to Bubbletea messages  
8.3. Integrate with Bubbletea commands  
8.4. Support Bubbletea's Update-View cycle  
8.5. Handle keyboard and mouse input via messages  

## Non-Functional Requirements

### Performance
- Component creation: < 1ms
- Render (simple component): < 5ms
- Render (complex component): < 20ms
- State update propagation: < 1ms
- Memory per component: < 2KB overhead

### Accessibility
- Components can define keyboard shortcuts
- Support screen reader hints (where applicable in TUI)
- Focus management for interactive components

### Security
- Props validation prevents invalid data
- Event handlers don't expose internal state
- Sandboxed component execution

### Type Safety
- **Strict typing:** All component APIs strongly typed
- **Generic props:** `Props[T any]` for type-safe props
- **Type-safe events:** Event payloads typed
- **No `any`:** Use interfaces with constraints
- **Compile-time validation:** Catch errors before runtime

## Acceptance Criteria

### Component Interface
- [ ] Component implements `tea.Model`
- [ ] Can create component instances
- [ ] Can render component to string
- [ ] Can update component via messages
- [ ] Integrates with Bubbletea runtime

### Builder API
- [ ] Fluent chaining works
- [ ] Builder validates configuration
- [ ] Build() produces working component
- [ ] Type-safe at compile time
- [ ] Clear error messages for misconfiguration

### Props System
- [ ] Props passed to component
- [ ] Props accessible in setup and template
- [ ] Props are immutable from component
- [ ] Required props validated
- [ ] Type-safe props

### Event System
- [ ] Components can emit events
- [ ] Parent can listen to child events
- [ ] Event payloads type-safe
- [ ] Multiple listeners per event
- [ ] Event handlers execute correctly
- [ ] Events bubble from child to parent
- [ ] Event propagation can be stopped
- [ ] Event metadata includes source and timestamp

### Template Rendering
- [ ] Template function renders to string
- [ ] Can access props and state
- [ ] Can use Lipgloss styling
- [ ] Can render child components
- [ ] Re-renders on state change

### Component Composition
- [ ] Can nest components
- [ ] Props flow to children
- [ ] Events bubble to parents
- [ ] Child lifecycle managed

### General
- [ ] Test coverage > 80%
- [ ] All public APIs documented
- [ ] Examples provided
- [ ] Performance targets met

## Dependencies
- **Requires:** 01-reactivity-system (for state management)
- **Unlocks:** 03-lifecycle-hooks, 05-directives, 06-built-in-components

## Edge Cases

### 1. Circular Component References
**Scenario:** Component A contains Component B, which contains Component A  
**Handling:** Detect at build time or runtime, return error with clear message

### 2. Props Mutation Attempt
**Scenario:** Component tries to modify props directly  
**Handling:** Props are copies or readonly; document that state should be used for mutable data

### 3. Missing Required Props
**Scenario:** Component created without required props  
**Handling:** Build() returns error listing missing props

### 4. Event Handler Panics
**Scenario:** Event handler panics during execution  
**Handling:** Recover, log error, continue execution (don't crash app)

### 5. Deeply Nested Components
**Scenario:** Component tree 50+ levels deep  
**Handling:** May impact performance; document recommended max depth (~10)

### 6. Large Component Trees
**Scenario:** 1000+ components in tree  
**Handling:** Optimize rendering with virtual scrolling at app level

### 7. State Update During Render
**Scenario:** Template function calls ref.Set()  
**Handling:** Defer state updates to next cycle, document as anti-pattern

## Testing Requirements

### Unit Tests (80%+ coverage)
- Component creation and building
- Props validation and access
- Event emission and handling
- Template rendering
- State management integration
- Bubbletea integration

### Integration Tests
- Component composition (parent-child)
- Props flow through tree
- Event bubbling
- State updates trigger re-renders
- Full Bubbletea integration

### Component Examples
- Simple button component
- Counter component (with state)
- Form component (with props and events)
- Nested component tree

## Atomic Design Level
**Foundation + Molecules**

- **Foundation:** Component interface and builder (enables all components)
- **Molecules:** Example components that demonstrate patterns

## Related Components
- Uses: Reactivity system (Ref, Computed, Watch)
- Enables: All built-in components (atoms, molecules, organisms)
- Integrates with: Lifecycle hooks, Composition API

## Technical Constraints
- Must work with Bubbletea message loop
- Cannot use goroutines directly (use Bubbletea commands)
- Render must be synchronous
- State updates async (via Bubbletea Update)

## API Design

### Component Creation
```go
// Basic component
c := NewComponent("Button").
    Template(func(ctx RenderContext) string {
        return "Click me"
    }).
    Build()

// Component with props
type ButtonProps struct {
    Label string
    Disabled bool
}

c := NewComponent("Button").
    Props(ButtonProps{
        Label: "Submit",
        Disabled: false,
    }).
    Template(func(ctx RenderContext) string {
        props := ctx.Props().(ButtonProps)
        return props.Label
    }).
    Build()

// Component with state
c := NewComponent("Counter").
    Setup(func(ctx *Context) {
        count := ctx.Ref(0)
        ctx.Expose("count", count)
        
        ctx.On("increment", func() {
            count.Set(count.Get() + 1)
        })
    }).
    Template(func(ctx RenderContext) string {
        count := ctx.Get("count").(*Ref[int])
        return fmt.Sprintf("Count: %d", count.Get())
    }).
    Build()
```

## Performance Benchmarks
```go
BenchmarkComponentCreate     100000   10000 ns/op   1024 B/op   10 allocs/op
BenchmarkComponentRender     50000    20000 ns/op   2048 B/op   20 allocs/op
BenchmarkComponentUpdate     100000   5000  ns/op   512  B/op   5  allocs/op
BenchmarkPropsAccess         1000000  1000  ns/op   0    B/op   0  allocs/op
BenchmarkEventEmit           500000   2000  ns/op   256  B/op   2  allocs/op
```

## Documentation Requirements
- [ ] Package godoc with component overview
- [ ] Component interface documentation
- [ ] ComponentBuilder API documentation
- [ ] Props system guide
- [ ] Event system guide
- [ ] Template function guide
- [ ] 10+ runnable examples
- [ ] Integration with Bubbletea guide
- [ ] Best practices document

## Success Metrics
- Developers can create components in < 5 minutes
- Component API feels intuitive (user testing)
- Props and events work as expected (zero confusion)
- Performance acceptable for 100+ components
- Test coverage > 80%
- Zero critical bugs in production usage

## Migration from Bubbletea

### Before (Pure Bubbletea)
```go
type model struct {
    count int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "+" {
            m.count++
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("Count: %d", m.count)
}
```

### After (BubblyUI)
```go
func NewCounter() *Component {
    return NewComponent("Counter").
        Setup(func(ctx *Context) {
            count := ctx.Ref(0)
            ctx.Expose("count", count)
            ctx.On("increment", func() {
                count.Set(count.Get() + 1)
            })
        }).
        Template(func(ctx RenderContext) string {
            count := ctx.Get("count").(*Ref[int])
            return fmt.Sprintf("Count: %d", count.Get())
        }).
        Build()
}
```

## Future Requirements (Not in current scope)

### 9. Error Tracking & Observability (Optional Enhancement)

9.1. Pluggable error reporter interface  
9.2. Report panics from event handlers to external services  
9.3. Include rich error context (component, event, stack trace)  
9.4. Built-in Console reporter for development  
9.5. Built-in Sentry reporter for production  
9.6. Breadcrumb collection for debugging  
9.7. Privacy-aware (PII filtering via hooks)  
9.8. Zero overhead when not configured  
9.9. Async error reporting (non-blocking)  

**Priority:** MEDIUM - Useful for production debugging but not critical  
**Estimated Effort:** 15 hours (2 days)  
**See:** `designs.md` "Error Tracking & Observability" section for full design

## Open Questions
1. Should components support async initialization?
2. How to handle component-level error boundaries?
3. Should we support component slots (like Vue)?
4. Optimal strategy for large component trees?
5. How to debug component hierarchies?
6. Which error tracking services should have built-in support? (Sentry, Rollbar, Bugsnag?)
