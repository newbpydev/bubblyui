# Parent-Child Communication Patterns

## Overview

This document outlines the design for parent-child communication in BubblyUI, drawing inspiration from React and Solid.js but adapted for Go's language features. Our communication model follows unidirectional data flow principles where data flows down from parents to children via props, and events flow up from children to parents via callbacks.

## Core Communication Patterns

### 1. Props Drilling (Parent → Child)

Props drilling is the primary method for passing data from parent components to their children. In BubblyUI, props are implemented as Go structs passed during component initialization.

```go
// ButtonProps defines the properties for a Button component
type ButtonProps struct {
    Label    string
    Disabled *Signal[bool]
    Style    *StyleProps
    OnClick  func()
}

// Button component receives props during initialization
type Button struct {
    props     ButtonProps
    isHovered *Signal[bool]
}

// NewButton initializes a Button with the given props
func NewButton(props ButtonProps) *Button {
    return &Button{
        props:     props,
        isHovered: NewSignal(false),
    }
}

// Usage in a parent component
func (p *Panel) Render() string {
    // Create button with props
    button := NewButton(ButtonProps{
        Label:    "Submit",
        Disabled: p.formInvalid,  // Pass a signal
        Style:    p.theme.Button, // Pass style props
        OnClick:  p.handleSubmit, // Pass callback
    })
    
    // Render the component tree
    return lipgloss.JoinVertical(
        lipgloss.Top,
        p.header.Render(),
        button.Render(),
        p.footer.Render(),
    )
}
```

**Key Features:**
- Type-safe props with explicit structures for each component
- Immutable props design (component never modifies received props)
- Props can include signals for reactive updates
- Default values for optional props

### 2. Event Callbacks (Child → Parent)

Events flow up from child to parent components through function callbacks passed down as props.

```go
// Example of a child component triggering an event
func (b *Button) handleClick() {
    if !b.props.Disabled.Value() && b.props.OnClick != nil {
        // Notify parent through callback
        b.props.OnClick()
    }
}

// Parent component handling the event
func (f *Form) handleSubmit() {
    // Process the form submission
    data := f.collectFormData()
    
    // Call the parent's callback if provided
    if f.props.OnSubmit != nil {
        f.props.OnSubmit(data)
    }
}
```

**Key Features:**
- Type-safe function signatures for events
- Optional callbacks (always check for nil)
- Events can carry payload data
- Standard naming conventions (On* for props, handle* for methods)

### 3. Context-like State Sharing

For deeply nested component trees, we'll implement a context-like system to avoid excessive props drilling:

```go
// ThemeContext provides theme settings to the component tree
type ThemeContext struct {
    Primary   *Signal[lipgloss.Color]
    Secondary *Signal[lipgloss.Color]
    TextColor *Signal[lipgloss.Color]
    FontSize  *Signal[int]
}

// App-wide context registry
var (
    contextRegistry = make(map[string]interface{})
    contextMutex    sync.RWMutex
)

// RegisterContext adds a context to the registry
func RegisterContext(key string, context interface{}) {
    contextMutex.Lock()
    defer contextMutex.Unlock()
    contextRegistry[key] = context
}

// GetContext retrieves a context from the registry
func GetContext[T any](key string) (T, bool) {
    contextMutex.RLock()
    defer contextMutex.RUnlock()
    
    ctx, ok := contextRegistry[key]
    if !ok {
        var zero T
        return zero, false
    }
    
    typed, ok := ctx.(T)
    return typed, ok
}

// Usage example in components
func (app *App) Initialize() {
    // Create theme context
    theme := &ThemeContext{
        Primary:   NewSignal(lipgloss.Color("#0366d6")),
        Secondary: NewSignal(lipgloss.Color("#6f42c1")),
        TextColor: NewSignal(lipgloss.Color("#24292e")),
        FontSize:  NewSignal(14),
    }
    
    // Register the context
    RegisterContext("theme", theme)
}

// In a deeply nested component
func (button *Button) getStyle() lipgloss.Style {
    // Get theme from context
    theme, ok := GetContext[*ThemeContext]("theme")
    if !ok {
        // Fallback to default style
        return defaultButtonStyle
    }
    
    // Use theme settings
    return lipgloss.NewStyle().
        Foreground(theme.TextColor.Value()).
        Background(theme.Primary.Value()).
        PaddingLeft(1).
        PaddingRight(1)
}
```

**Key Features:**
- Type-safe context retrieval with generics
- Context values can be reactive (signals)
- Global registry with string-based keys
- Fallback mechanisms when contexts are missing

### 4. Child Component Registration

Parent components need to manage child components, including initialization, updating, and rendering:

```go
// Container manages a collection of child components
type Container struct {
    props    ContainerProps
    children []Component
    layout   *Signal[LayoutType]
}

// AddChild registers a child component with the container
func (c *Container) AddChild(child Component) {
    c.children = append(c.children, child)
    
    // Initialize the child
    if initializer, ok := child.(Initializable); ok {
        initializer.Initialize()
    }
}

// Update propagates updates to all children
func (c *Container) Update(msg tea.Msg) (tea.Cmd, error) {
    var cmds []tea.Cmd
    
    // Update all children
    for _, child := range c.children {
        cmd, err := child.Update(msg)
        if err != nil {
            return nil, err
        }
        if cmd != nil {
            cmds = append(cmds, cmd)
        }
    }
    
    return tea.Batch(cmds...), nil
}

// Render composes the output of all children
func (c *Container) Render() string {
    childViews := make([]string, len(c.children))
    
    // Collect rendered output from all children
    for i, child := range c.children {
        childViews[i] = child.Render()
    }
    
    // Apply layout based on layout type
    switch c.layout.Value() {
    case LayoutVertical:
        return lipgloss.JoinVertical(lipgloss.Left, childViews...)
    case LayoutHorizontal:
        return lipgloss.JoinHorizontal(lipgloss.Top, childViews...)
    case LayoutGrid:
        // Implement grid layout
        return renderGrid(childViews, c.props.Columns)
    default:
        return lipgloss.JoinVertical(lipgloss.Left, childViews...)
    }
}
```

### 5. Event Bubbling

For certain events, we'll implement an event bubbling system to allow events to propagate up the component tree:

```go
// Event represents a UI event that can bubble up the component tree
type Event struct {
    Type       EventType
    Target     Component
    Data       interface{}
    Bubbles    bool
    Cancelable bool
    canceled   bool
}

// StopPropagation prevents the event from bubbling further
func (e *Event) StopPropagation() {
    e.Bubbles = false
}

// PreventDefault cancels the default action
func (e *Event) PreventDefault() {
    e.canceled = true
}

// Component receives events and can choose to handle or bubble them
func (c *Container) HandleEvent(event *Event) bool {
    // Try to handle the event directly
    if c.props.OnEvent != nil {
        handled := c.props.OnEvent(event)
        if handled || event.canceled {
            return true
        }
    }
    
    // If not handled and allowed to bubble, send to parent
    if event.Bubbles && c.parent != nil {
        return c.parent.HandleEvent(event)
    }
    
    return false
}
```

## Communication Flow Diagrams

### Props Flow Down
```
   App (state=X)
   │
   ├─► Header (props={title: X.title})
   │
   └─► Content (props={items: X.items})
       │
       └─► Item (props={data: items[0]})
           │
           └─► Button (props={label: "View", onClick: fn})
```

### Events Flow Up
```
   Button ("Click")
   │
   └─► Item (handles "Click", calls props.onSelect)
       │
       └─► Content (handles "onSelect", updates selectedItem)
           │
           └─► App (receives notification of selection change)
```

### Context Access (Cross-cutting)
```
   App (provides ThemeContext, AuthContext)
   ┌─────────────────┐
   │                 │
   ▼                 ▼
 Header    Content (consumes AuthContext)
   │        ┌─────────────┐
   │        │             │
   │        ▼             ▼
   │      Item          Item
   │        │             │
   │        ▼             ▼
   └───► Button        Button (consumes ThemeContext)
```

## Implementation Considerations

### 1. Go-Friendly API Design

Our communication patterns must feel natural to Go developers:

- Use strong typing and avoid interface{} where possible
- Prefer explicit over implicit behavior
- Follow Go naming conventions
- Leverage Go's type system and generics
- Minimize reflection use for better performance and safety

### 2. Testing Strategy

All communication patterns need comprehensive tests:

- Verify props are correctly passed to child components
- Ensure events bubble properly through the component hierarchy
- Test context registration and retrieval
- Verify signal propagation across component boundaries
- Test error conditions (nil handlers, type mismatches)

### 3. Performance Optimization

Communication should be optimized for terminal UI performance:

- Minimize allocations in hot paths
- Use efficient synchronization patterns
- Batch updates to reduce rendering frequency
- Implement smart reconciliation to minimize string operations

### 4. Memory Safety

Prevent memory leaks in the communication system:

- Properly clean up event listeners
- Handle circular references between parent and child components
- Ensure context consumers don't retain references longer than needed
- Implement Dispose pattern for component cleanup

## Next Steps

1. Create proof-of-concept implementations of each pattern
2. Design test cases that verify correct communication flow
3. Benchmark communication performance in complex component trees
4. Create diagrams to visualize the communication flows
