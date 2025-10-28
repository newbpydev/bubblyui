# Design Specification: Component Model

## Component Hierarchy

```
Foundation Layer (Interfaces)
└── Component System
    ├── Component (Interface)
    ├── ComponentBuilder (Fluent API)
    ├── Context (Setup context)
    └── RenderContext (Template context)

Component Tree Example:
App (Root Component)
├── Header (Organism)
│   ├── Logo (Atom)
│   └── NavButton (Atom)
└── Counter (Molecule)
    ├── Button (Atom) "+"
    ├── Display (Atom)
    └── Button (Atom) "-"
```

---

## Architecture Overview

### System Layers

```
┌────────────────────────────────────────────────────────────┐
│                    BubblyUI Component Layer                │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────┐     ┌──────────────┐                   │
│  │  Component   │────>│ Reactivity   │                   │
│  │  Interface   │     │  System      │                   │
│  └──────────────┘     └──────────────┘                   │
│         │                                                  │
│         ├──> Props System                                 │
│         ├──> Event System                                 │
│         ├──> State Management                             │
│         └──> Template Engine                              │
│                                                            │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│                   Bubbletea Runtime                        │
│  ┌──────────┐   ┌──────────┐   ┌─────────┐              │
│  │  Model   │──>│  Update  │──>│  View   │              │
│  └──────────┘   └──────────┘   └─────────┘              │
└────────────────────────────────────────────────────────────┘
```

### Component Lifecycle Flow

```
Developer Code: NewComponent("name")
    ↓
ComponentBuilder created
    ↓
Developer Code: .Setup(fn).Props(p).Template(t)
    ↓
Developer Code: .Build()
    ↓
Component validation
    ↓
Component instance created
    ↓
Component.Init() called by Bubbletea
    ↓
Setup function executes (create Refs, register handlers)
    ↓
Component ready for rendering
    ↓
User interaction → Message → Update
    ↓
Event handler executes
    ↓
State changes (Refs updated)
    ↓
Re-render triggered
    ↓
Template function called
    ↓
View string generated
```

---

## Data Flow

### 1. Props Flow (Top-Down)
```
Parent Component
    ↓ (pass props)
Child Component Props struct
    ↓ (accessed in)
Setup function + Template function
    ↓ (renders)
UI Output
```

### 2. Event Flow (Bottom-Up)
```
User Interaction (keypress)
    ↓
Bubbletea Message
    ↓
Component.Update()
    ↓
Event handler matched (On("keypress"))
    ↓
Handler executes
    ↓
May emit custom event (Emit("customEvent", data))
    ↓
Parent listens (On("customEvent"))
    ↓
Parent handler executes
```

### 3. State Update Flow
```
Event handler: ref.Set(newValue)
    ↓
Ref notifies watchers (reactivity system)
    ↓
Component marked for re-render
    ↓
Bubbletea calls Component.View()
    ↓
Template function executes
    ↓
ref.Get() returns new value
    ↓
New UI string generated
```

---

## State Management

### Component Structure
```go
type Component struct {
    // Identification
    name      string
    id        string  // unique instance ID
    
    // Configuration
    props     interface{}
    children  []*Component
    
    // State
    state     map[string]interface{}  // Stores Refs, Computed
    context   *Context                 // Setup context
    
    // Behavior
    setup     SetupFunc
    template  RenderFunc
    handlers  map[string]EventHandler
    
    // Lifecycle
    mounted   bool
    parent    *Component
    
    // Bubbletea integration
    teaModel  tea.Model  // Underlying Bubbletea model
}
```

### Context (Setup Time)
```go
type Context struct {
    component *Component
    
    // State management
    Ref       func(value interface{}) *Ref[interface{}]
    Computed  func(fn func() interface{}) *Computed[interface{}]
    Watch     func(ref *Ref[interface{}], callback WatchCallback)
    
    // Component API
    Expose    func(key string, value interface{})
    Get       func(key string) interface{}
    
    // Events
    On        func(event string, handler EventHandler)
    Emit      func(event string, data interface{})
    
    // Props access
    Props     func() interface{}
    
    // Child management
    Children  func() []*Component
}
```

### RenderContext (Template Time)
```go
type RenderContext struct {
    component *Component
    
    // Data access
    Get      func(key string) interface{}
    Props    func() interface{}
    Children func() []*Component
    
    // Rendering
    RenderChild func(child *Component) string
    
    // Styling (Lipgloss integration)
    Style    *lipgloss.Style
}
```

---

## Type Definitions

### Core Types
```go
// Component interface (implements tea.Model)
type Component interface {
    tea.Model
    
    // Identity
    Name() string
    ID() string
    
    // Props
    Props() interface{}
    SetProps(props interface{}) error
    
    // Events
    Emit(event string, data interface{})
    On(event string, handler EventHandler)
    
    // Children
    Children() []*Component
    AddChild(child *Component)
    
    // State
    Get(key string) interface{}
    Set(key string, value interface{})
}

// ComponentBuilder (fluent API)
type ComponentBuilder struct {
    component *componentImpl
    errors    []error
}

func NewComponent(name string) *ComponentBuilder
func (b *ComponentBuilder) Props(props interface{}) *ComponentBuilder
func (b *ComponentBuilder) Setup(fn SetupFunc) *ComponentBuilder
func (b *ComponentBuilder) Template(fn RenderFunc) *ComponentBuilder
func (b *ComponentBuilder) Children(children ...*Component) *ComponentBuilder
func (b *ComponentBuilder) Build() (*Component, error)

// Function types
type SetupFunc func(ctx *Context)
type RenderFunc func(ctx RenderContext) string
type EventHandler func(data interface{})

// Event system
type Event struct {
    Name      string
    Source    *Component
    Data      interface{}
    Timestamp time.Time
}
```

### Generic Props Pattern
```go
// Type-safe props
type Props[T any] struct {
    Value T
}

func NewPropsComponent[T any](name string, props T) *ComponentBuilder {
    return NewComponent(name).Props(Props[T]{Value: props})
}

// Usage
type ButtonProps struct {
    Label string
    Disabled bool
}

button := NewPropsComponent("Button", ButtonProps{
    Label: "Click me",
    Disabled: false,
})
```

---

## API Contracts

### ComponentBuilder API
```go
// Create component
builder := NewComponent("MyComponent")

// Configure (all optional except Build)
builder.
    Props(myProps).                    // Set props
    Setup(func(ctx *Context) {         // Initialize state
        count := ctx.Ref(0)
        ctx.Expose("count", count)
    }).
    Template(func(ctx RenderContext) string {  // Define view
        return "Hello"
    }).
    Children(child1, child2).          // Add children
    Build()                            // Create component

// Build returns (*Component, error)
component, err := builder.Build()
if err != nil {
    // Handle validation errors
}
```

### Setup Function API
```go
Setup(func(ctx *Context) {
    // Create reactive state
    count := ctx.Ref(0)
    doubled := ctx.Computed(func() int {
        return count.Get() * 2
    })
    
    // Expose to template
    ctx.Expose("count", count)
    ctx.Expose("doubled", doubled)
    
    // Register event handlers
    ctx.On("increment", func(data interface{}) {
        count.Set(count.Get() + 1)
    })
    
    // Watch for changes
    ctx.Watch(count, func(newVal, oldVal int) {
        log.Printf("Count: %d -> %d", oldVal, newVal)
    })
    
    // Access props
    props := ctx.Props().(MyProps)
    
    // Access children
    children := ctx.Children()
})
```

### Template Function API
```go
Template(func(ctx RenderContext) string {
    // Access state
    count := ctx.Get("count").(*Ref[int])
    
    // Access props
    props := ctx.Props().(ButtonProps)
    
    // Use Lipgloss
    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color("63")).
        Bold(true)
    
    // Render
    return style.Render(
        fmt.Sprintf("%s: %d", props.Label, count.Get()),
    )
    
    // Render children
    childOutputs := []string{}
    for _, child := range ctx.Children() {
        childOutputs = append(childOutputs, ctx.RenderChild(child))
    }
    return strings.Join(childOutputs, "\n")
})
```

### Event System API
```go
// In child component setup
ctx.On("click", func(data interface{}) {
    // Handle internal click
})
ctx.Emit("buttonClicked", map[string]interface{}{
    "timestamp": time.Now(),
})

// In parent component setup
child.On("buttonClicked", func(data interface{}) {
    // Parent handles child's event
})
```

---

## Implementation Details

### Component Creation
```go
type componentImpl struct {
    name      string
    id        string
    props     interface{}
    state     map[string]interface{}
    setup     SetupFunc
    template  RenderFunc
    children  []*Component
    handlers  map[string][]EventHandler
    mounted   bool
}

func NewComponent(name string) *ComponentBuilder {
    id := generateID()  // UUID or sequential
    return &ComponentBuilder{
        component: &componentImpl{
            name:     name,
            id:       id,
            state:    make(map[string]interface{}),
            handlers: make(map[string][]EventHandler),
            children: []*Component{},
        },
        errors: []error{},
    }
}
```

### Bubbletea Integration
```go
// Component implements tea.Model
func (c *componentImpl) Init() tea.Cmd {
    // Run setup if not already done
    if !c.mounted && c.setup != nil {
        ctx := c.createContext()
        c.setup(ctx)
        c.mounted = true
    }
    
    // Initialize children
    var cmds []tea.Cmd
    for _, child := range c.children {
        cmds = append(cmds, child.Init())
    }
    
    return tea.Batch(cmds...)
}

func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Dispatch to handlers
        if handlers, ok := c.handlers["keypress"]; ok {
            for _, handler := range handlers {
                handler(msg.String())
            }
        }
        
    case CustomEventMsg:
        // Custom event from child
        if handlers, ok := c.handlers[msg.Name]; ok {
            for _, handler := range handlers {
                handler(msg.Data)
            }
        }
    }
    
    // Update children
    var cmds []tea.Cmd
    for i, child := range c.children {
        updatedChild, cmd := child.Update(msg)
        c.children[i] = updatedChild.(*componentImpl)
        cmds = append(cmds, cmd)
    }
    
    return c, tea.Batch(cmds...)
}

func (c *componentImpl) View() string {
    if c.template == nil {
        return ""
    }
    
    ctx := c.createRenderContext()
    return c.template(ctx)
}
```

### Props Validation
```go
func (b *ComponentBuilder) Build() (*Component, error) {
    // Validate configuration
    if b.component.template == nil {
        b.errors = append(b.errors, errors.New("template is required"))
    }
    
    // Check for errors
    if len(b.errors) > 0 {
        return nil, fmt.Errorf("component validation failed: %v", b.errors)
    }
    
    return b.component, nil
}
```

### Event Emission
```go
func (c *componentImpl) Emit(event string, data interface{}) {
    // Create event message
    msg := CustomEventMsg{
        Name:   event,
        Source: c,
        Data:   data,
    }
    
    // Send to Bubbletea (via parent or program)
    if c.parent != nil {
        // Bubble up
        c.parent.Update(msg)
    } else {
        // Send to program
        c.sendMsg(msg)
    }
}
```

### Event Bubbling Architecture

**Purpose**: Events automatically propagate from child components up through parent components until handled or reaching the root.

**Design Pattern**: Follows Vue.js and DOM event bubbling model for familiar, predictable behavior.

#### Bubbling Flow
```
Child Component (emits "submit")
    ↓
Parent Component (listens for "submit" or bubbles up)
    ↓
Grandparent Component (listens or bubbles up)
    ↓
Root Component (final opportunity to handle)
```

#### Implementation Design

```go
type Event struct {
    Name      string
    Source    Component    // Original emitter
    Data      interface{}
    Timestamp time.Time
    Stopped   bool         // Flag to stop propagation
}

// Emit with automatic bubbling
func (c *componentImpl) Emit(event string, data interface{}) {
    evt := Event{
        Name:      event,
        Source:    c,
        Data:      data,
        Timestamp: time.Now(),
        Stopped:   false,
    }
    
    c.bubbleEvent(evt)
}

// bubbleEvent propagates event up the component tree
func (c *componentImpl) bubbleEvent(evt Event) {
    // Skip if event propagation was stopped
    if evt.Stopped {
        return
    }
    
    // Execute local handlers first
    c.handlersMu.RLock()
    handlers, ok := c.handlers[evt.Name]
    c.handlersMu.RUnlock()
    
    if ok {
        for _, handler := range handlers {
            handler(evt.Data)
            // Note: handlers can call StopPropagation()
            // to set evt.Stopped = true
        }
    }
    
    // Bubble to parent if not stopped and parent exists
    if !evt.Stopped && c.parent != nil {
        c.parent.(*componentImpl).bubbleEvent(evt)
    }
}

// StopPropagation prevents event from bubbling further
func (e *Event) StopPropagation() {
    e.Stopped = true
}
```

#### Event Handler with Stop Propagation

```go
// Child component
child := NewComponent("Button").
    Setup(func(ctx *Context) {
        ctx.On("click", func(data interface{}) {
            // Handle locally
            fmt.Println("Button clicked")
            
            // Emit custom event that will bubble
            ctx.Emit("buttonClicked", map[string]interface{}{
                "timestamp": time.Now(),
                "buttonId": "submit",
            })
        })
    }).
    Build()

// Parent component - handles bubbled events
parent := NewComponent("Form").
    Children(child).
    Setup(func(ctx *Context) {
        // Listen for child's buttonClicked event
        ctx.On("buttonClicked", func(data interface{}) {
            eventData := data.(map[string]interface{})
            fmt.Printf("Form received button click: %v\n", eventData)
            
            // Stop propagation to prevent grandparent from seeing it
            // (Implementation detail: handler can set stop flag)
        })
    }).
    Build()
```

#### Use Cases

1. **Form Submission**: Button click bubbles up to form for validation
2. **Menu Selection**: Item selection bubbles to menu container
3. **List Actions**: Item actions bubble to list for coordinated updates
4. **Error Handling**: Errors bubble up for centralized handling
5. **Analytics**: All events bubble to root for logging

#### Performance Considerations

- **Efficient Path**: O(depth) where depth is component tree depth
- **Early Exit**: Handlers can stop propagation to prevent unnecessary traversal
- **No Overhead**: If no parent, bubbling stops immediately
- **Thread-Safe**: Uses existing handlersMu for concurrent access

#### Comparison with DOM Event Bubbling

| Feature | DOM Events | BubblyUI Events |
|---------|------------|-----------------|
| Direction | Bottom-up | Bottom-up |
| Stop Propagation | `event.stopPropagation()` | Event.Stopped flag |
| Capture Phase | Yes | No (not needed in TUI) |
| Default Prevention | `event.preventDefault()` | Not applicable |
| Event Object | Full DOM Event | Event struct with metadata |

---

## Integration with Reactivity System

### State Storage
```go
// In Setup function
ctx.Ref(0)  // Creates Ref[int]
    ↓
Stored in component.state["generatedKey"] = ref
    ↓
ctx.Expose("count", ref)
    ↓
Stored in component.state["count"] = ref
    ↓
Template accesses: ctx.Get("count")
```

### Automatic Re-render
```go
// When Ref changes
ref.Set(newValue)
    ↓
Watcher notifies component (via Watch in setup)
    ↓
Component sends RefChangedMsg to Bubbletea
    ↓
Bubbletea triggers Update -> View cycle
    ↓
Template re-executes with new value
```

---

## Performance Optimizations

### 1. Lazy Component Creation
```go
// Only build when needed
builder := NewComponent("Button")  // Fast, no allocation
// ... configure ...
component := builder.Build()  // Actually creates component
```

### 2. Render Caching
```go
type componentImpl struct {
    lastRender  string
    renderCache *cache
}

func (c *componentImpl) View() string {
    // Check if state changed
    if !c.stateChanged() {
        return c.lastRender  // Return cached
    }
    
    // Re-render
    c.lastRender = c.template(c.createRenderContext())
    return c.lastRender
}
```

### 3. Virtual Children
```go
// Only render visible children (for large lists)
func (ctx RenderContext) RenderChildren() string {
    visible := filterVisible(ctx.Children())
    outputs := make([]string, len(visible))
    for i, child := range visible {
        outputs[i] = ctx.RenderChild(child)
    }
    return strings.Join(outputs, "\n")
}
```

### 4. Event Handler Pooling
```go
var handlerPool = sync.Pool{
    New: func() interface{} {
        return &eventHandler{}
    },
}
```

---

## Error Handling

### Error Types
```go
var (
    ErrMissingTemplate = errors.New("component template is required")
    ErrInvalidProps    = errors.New("props validation failed")
    ErrCircularRef     = errors.New("circular component reference detected")
    ErrMaxDepth        = errors.New("max component depth exceeded")
)
```

### Error Scenarios
1. **Missing Template:** Return error from Build()
2. **Invalid Props:** Validate in SetProps()
3. **Circular Reference:** Detect during AddChild()
4. **Handler Panic:** Recover in Update(), log error

---

## Testing Strategy

### Unit Tests
```go
func TestComponent_Creation(t *testing.T)
func TestComponent_Props(t *testing.T)
func TestComponent_Events(t *testing.T)
func TestComponent_Rendering(t *testing.T)
func TestComponent_Children(t *testing.T)
func TestComponent_BubbletteaIntegration(t *testing.T)
```

### Integration Tests
```go
func TestComponentTree_ParentChild(t *testing.T)
func TestComponentTree_PropsPassing(t *testing.T)
func TestComponentTree_EventBubbling(t *testing.T)
func TestComponentTree_StateManagement(t *testing.T)
```

### Example Components
```go
func ExampleSimpleButton()
func ExampleCounterWithState()
func ExampleFormWithProps()
func ExampleNestedComponents()
```

---

## Example Usage

### Simple Button
```go
button := NewComponent("Button").
    Props(ButtonProps{Label: "Click me"}).
    Template(func(ctx RenderContext) string {
        props := ctx.Props().(ButtonProps)
        style := lipgloss.NewStyle().Bold(true)
        return style.Render(props.Label)
    }).
    Build()
```

### Counter with State
```go
counter := NewComponent("Counter").
    Setup(func(ctx *Context) {
        count := ctx.Ref(0)
        ctx.Expose("count", count)
        
        ctx.On("increment", func(_ interface{}) {
            count.Set(count.Get() + 1)
        })
        ctx.On("decrement", func(_ interface{}) {
            count.Set(count.Get() - 1)
        })
    }).
    Template(func(ctx RenderContext) string {
        count := ctx.Get("count").(*Ref[int])
        return fmt.Sprintf("Count: %d", count.Get())
    }).
    Build()
```

### Parent-Child Communication
```go
// Child button
button := NewComponent("Button").
    Setup(func(ctx *Context) {
        ctx.On("keypress", func(data interface{}) {
            if data.(string) == "enter" {
                ctx.Emit("clicked", time.Now())
            }
        })
    }).
    Template(func(ctx RenderContext) string {
        return "[Submit]"
    }).
    Build()

// Parent form
form := NewComponent("Form").
    Children(button).
    Setup(func(ctx *Context) {
        // Listen to child events
        children := ctx.Children()
        children[0].On("clicked", func(data interface{}) {
            log.Printf("Button clicked at: %v", data)
            // Handle form submission
        })
    }).
    Template(func(ctx RenderContext) string {
        return ctx.RenderChild(ctx.Children()[0])
    }).
    Build()
```

---

## Migration Path

### Wrapping Existing Bubbletea Model
```go
// Existing Bubbletea model
type legacyModel struct {
    count int
}

func (m legacyModel) Init() tea.Cmd { /* ... */ }
func (m legacyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* ... */ }
func (m legacyModel) View() string { /* ... */ }

// Wrap in BubblyUI component
wrapped := NewComponent("Legacy").
    Setup(func(ctx *Context) {
        model := legacyModel{count: 0}
        ctx.Set("model", model)
    }).
    Template(func(ctx RenderContext) string {
        model := ctx.Get("model").(legacyModel)
        return model.View()
    }).
    Build()
```

---

## Future Enhancements

1. **Async Components:** Load component code dynamically
2. **Component Slots:** Named content areas (like Vue slots)
3. **Mixins:** Share functionality across components
4. **Higher-Order Components:** Wrap components with behavior
5. **Component Registry:** Global component registration
6. **Dev Tools:** Inspect component tree at runtime
7. **Error Tracking & Observability:** Integration with error tracking services (see below)

---

## Error Tracking & Observability

### Problem Statement

**Current State:**
- Component errors (panics in event handlers) are recovered but silently discarded
- `HandlerPanicError` created in `events.go:101` but assigned to `_` (unused)
- No mechanism to report errors to external services (Sentry, Rollbar, custom backends)
- Production debugging difficult without error visibility
- No context about component state when errors occur

**User Need:**
Developers need to:
- Know when production errors occur
- See error context (component name, event, panic value, stack trace)
- Track error frequency and patterns
- Debug issues users encounter
- Monitor application health

### Solution Design

**Architecture:** Hook-based error reporting system inspired by Sentry Go SDK

#### Type Definitions

```go
// ErrorReporter interface for pluggable error tracking backends
type ErrorReporter interface {
    // ReportPanic reports a panic that occurred in an event handler
    ReportPanic(err *HandlerPanicError, ctx *ErrorContext)
    
    // ReportError reports a general error
    ReportError(err error, ctx *ErrorContext)
    
    // Flush ensures all pending errors are sent before shutdown
    Flush(timeout time.Duration) error
}

// ErrorContext provides rich context about where/when error occurred
type ErrorContext struct {
    ComponentName  string
    ComponentID    string
    EventName      string
    Timestamp      time.Time
    Tags           map[string]string
    Extra          map[string]interface{}
    Breadcrumbs    []Breadcrumb
    StackTrace     []byte
}

// Breadcrumb represents a navigation trail or action leading to error
type Breadcrumb struct {
    Type      string
    Category  string
    Message   string
    Level     string
    Timestamp time.Time
    Data      map[string]interface{}
}

// Global error reporter (can be set by application)
var globalErrorReporter ErrorReporter

// SetErrorReporter configures the global error reporter
func SetErrorReporter(reporter ErrorReporter)

// GetErrorReporter returns the current error reporter (or nil)
func GetErrorReporter() ErrorReporter
```

#### Integration Points

**1. Event Handler Panic Recovery** (already in `events.go`)
```go
// Current code (events.go:96-107)
func() {
    defer func() {
        if r := recover(); r != nil {
            panicErr := &HandlerPanicError{
                ComponentName: c.name,
                EventName:     event.Name,
                PanicValue:    r,
            }
            
            // NEW: Report to error tracking service if configured
            if reporter := GetErrorReporter(); reporter != nil {
                reporter.ReportPanic(panicErr, &ErrorContext{
                    ComponentName: c.name,
                    ComponentID:   c.id,
                    EventName:     event.Name,
                    Timestamp:     time.Now(),
                    StackTrace:    debug.Stack(),
                })
            }
        }
    }()
    handler(event)
}()
```

**2. Component Builder Validation Errors**
```go
// Option: Report validation errors during development
func (b *ComponentBuilder) Build() (Component, error) {
    if len(b.errors) > 0 {
        validationErr := &ValidationError{...}
        
        // Report in development mode
        if reporter := GetErrorReporter(); reporter != nil && IsDevelopment() {
            reporter.ReportError(validationErr, &ErrorContext{...})
        }
        
        return nil, validationErr
    }
    return b.component, nil
}
```

**3. Breadcrumb Collection**
```go
// Automatic breadcrumbs for component lifecycle
func (c *componentImpl) Init() tea.Cmd {
    recordBreadcrumb("component.init", c.name)
    // ... existing code
}

func (c *componentImpl) Emit(event string, data interface{}) {
    recordBreadcrumb("component.event", event)
    // ... existing code
}
```

#### Built-in Reporters

**Console Reporter (Development)**
```go
type ConsoleReporter struct {
    verbose bool
}

func (r *ConsoleReporter) ReportPanic(err *HandlerPanicError, ctx *ErrorContext) {
    log.Printf("[ERROR] Panic in component '%s' event '%s': %v\n",
        ctx.ComponentName, ctx.EventName, err.PanicValue)
    if r.verbose {
        log.Printf("Stack trace:\n%s", ctx.StackTrace)
    }
}
```

**Sentry Reporter (Production)**
```go
type SentryReporter struct {
    hub *sentry.Hub
}

func NewSentryReporter(dsn string) (*SentryReporter, error) {
    err := sentry.Init(sentry.ClientOptions{
        Dsn: dsn,
        BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
            // Filter or modify events before sending
            return event
        },
    })
    if err != nil {
        return nil, err
    }
    
    return &SentryReporter{
        hub: sentry.CurrentHub(),
    }, nil
}

func (r *SentryReporter) ReportPanic(err *HandlerPanicError, ctx *ErrorContext) {
    r.hub.WithScope(func(scope *sentry.Scope) {
        scope.SetTag("component", ctx.ComponentName)
        scope.SetTag("event", ctx.EventName)
        scope.SetExtra("panic_value", err.PanicValue)
        
        // Add breadcrumbs
        for _, bc := range ctx.Breadcrumbs {
            scope.AddBreadcrumb(&sentry.Breadcrumb{
                Type:      bc.Type,
                Category:  bc.Category,
                Message:   bc.Message,
                Timestamp: bc.Timestamp,
                Data:      bc.Data,
            }, 100)
        }
        
        r.hub.CaptureException(fmt.Errorf("%w: %v", ErrHandlerPanic, err.PanicValue))
    })
}
```

#### Usage Example

```go
// Application setup
func main() {
    // Development: Console reporter
    if os.Getenv("ENV") == "development" {
        bubbly.SetErrorReporter(&bubbly.ConsoleReporter{verbose: true})
    } else {
        // Production: Sentry reporter
        reporter, err := bubbly.NewSentryReporter(os.Getenv("SENTRY_DSN"))
        if err != nil {
            log.Fatal(err)
        }
        defer reporter.Flush(5 * time.Second)
        bubbly.SetErrorReporter(reporter)
    }
    
    // Create application
    app := bubbly.NewComponent("App").
        Setup(func(ctx *bubbly.Context) {
            // Component setup
            ctx.On("risky-operation", func(data interface{}) {
                // If this panics, it's automatically reported to Sentry
                processRiskyData(data)
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            return "App"
        }).
        Build()
    
    p := tea.NewProgram(app)
    p.Run()
}
```

### Benefits

1. **Production Visibility:** Know when errors occur in production
2. **Context-Rich:** Full component state, event info, breadcrumbs
3. **Pluggable:** Support any error tracking service (Sentry, Rollbar, custom)
4. **Zero Impact:** If no reporter configured, zero overhead
5. **Development Friendly:** Console reporter for local debugging
6. **Privacy Aware:** BeforeSend hooks to filter sensitive data

### Performance Considerations

- **No Reporter:** Zero overhead (nil check, no allocation)
- **With Reporter:** Minimal overhead (single interface call)
- **Async Reporting:** Errors sent in background (non-blocking)
- **Buffering:** Batch errors to reduce network calls
- **Sampling:** Support for error sampling in high-traffic apps

### Privacy & Security

1. **PII Filtering:** BeforeSend hooks to remove sensitive data
2. **Opt-in:** Error tracking disabled by default
3. **User Control:** Application controls what data is sent
4. **Stack Traces:** Optional (can disable for privacy)
5. **Data Retention:** Controlled by external service

### Known Limitations

1. **External Dependency:** Requires error tracking service
2. **Network Required:** Errors buffered if offline
3. **Configuration Required:** Application must set up reporter
4. **Not Real-time:** Errors sent asynchronously
5. **Overhead:** Small performance cost for error reporting

### Priority

**MEDIUM** - Useful for production debugging but not critical for framework functionality

### Implementation Estimate

- **Error Reporter Interface:** 2 hours
- **Console Reporter:** 1 hour  
- **Sentry Reporter:** 3 hours
- **Integration with existing error handling:** 2 hours
- **Breadcrumb system:** 2 hours
- **Tests:** 3 hours
- **Documentation:** 2 hours
- **Total:** ~15 hours (2 days)

### Future Enhancements (Error Tracking)

1. **Custom Reporters:** Easy to add support for Rollbar, Bugsnag, etc.
2. **Local Storage:** Buffer errors offline, send when connection restored
3. **Error Grouping:** Smart grouping of similar errors
4. **User Context:** Track which user encountered error
5. **Performance Monitoring:** Extend to track component render times
6. **Session Replay:** Record user interactions leading to error
