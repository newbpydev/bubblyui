# Component Architecture Proof-of-Concept Sketches

This document provides conceptual sketches of the BubblyUI component architecture to illustrate how the various pieces fit together.

## Component Hierarchy Sketch

The following ASCII diagram shows a simple application built with the BubblyUI component architecture:

```
+--------------------------------------- App Component ----------------------------------+
| [Signal: appState]                                                                     |
|                                                                                        |
| +---------------------------------- Header Component --------------------------------+ |
| | Props: {title: "BubblyUI Demo"}                                                    | |
| | +----------------------------------------------------------------------------+     | |
| | |                             BubblyUI Demo App                              |     | |
| | +----------------------------------------------------------------------------+     | |
| +------------------------------------------------------------------------------------+ |
|                                                                                        |
| +--------------------------------- Sidebar Component --------------------------------+ |
| | Props: {selected: Signal[string], onSelect: func()}    [Signal: isHovered]         | |
| |                                                                                    | |
| | +------------------------+                                                         | |
| | | > Dashboard           |                                                          | |
| | | > Settings            |                                                          | |
| | | > Profile             |                                                          | |
| | | > Help                |                                                          | |
| | +------------------------+                                                         | |
| +------------------------------------------------------------------------------------+ |
|                                                                                        |
| +--------------------------------- Content Component --------------------------------+ |
| | Props: {view: Signal[string], data: Signal[AppData]}    [Signal: contentState]     | |
| |                                                                                    | |
| | +---------------------------- Dashboard View ---------------------------+          | |
| | |                                                                       |          | |
| | | +----------------------+  +----------------------+                    |          | |
| | | |    Active Tasks      |  |    Pending Tasks     |                    |          | |
| | | | [TaskList Component] |  | [TaskList Component] |                    |          | |
| | | +----------------------+  +----------------------+                    |          | |
| | |                                                                       |          | |
| | | +---------------------------------------------------------------+     |          | |
| | | |                     [Chart Component]                         |     |          | |
| | | +---------------------------------------------------------------+     |          | |
| | +-----------------------------------------------------------------------+          | |
| +------------------------------------------------------------------------------------+ |
|                                                                                        |
| +---------------------------------- Footer Component --------------------------------+ |
| | Props: {copyright: string, version: string}                                        | |
| | +----------------------------------------------------------------------------+     | |
| | |                      © 2025 BubblyUI - Version 1.0.0                       |     | |
| | +----------------------------------------------------------------------------+     | |
| +------------------------------------------------------------------------------------+ |
+----------------------------------------------------------------------------------------+
```

## Component Code Structure Sketch

### Component Definition

```go
// Component interface that all components must implement
type Component interface {
    // Initialize sets up the component
    Initialize() error
    
    // Update handles messages and state changes
    Update(msg tea.Msg) (tea.Cmd, error)
    
    // Render returns the visual representation of the component
    Render() string
    
    // Dispose cleans up resources when component is removed
    Dispose() error
}
```

### Functional Component Example

```go
// Example of a functional button component
type Button struct {
    props ButtonProps
    
    // Internal signals
    isHovered *Signal[bool]
    isFocused *Signal[bool]
    
    // Computed values
    style *Computed[lipgloss.Style]
}

type ButtonProps struct {
    Label    string
    Disabled *Signal[bool]
    OnClick  func()
    Style    *StyleProps
}

func NewButton(props ButtonProps) *Button {
    button := &Button{
        props:     props,
        isHovered: NewSignal(false),
        isFocused: NewSignal(false),
    }
    
    // Initialize computed style based on component state
    button.style = NewComputed(func() lipgloss.Style {
        baseStyle := DefaultButtonStyle
        
        if props.Style != nil {
            baseStyle = props.Style.Apply(baseStyle)
        }
        
        // Apply state-based styling
        if button.isHovered.Value() {
            baseStyle = baseStyle.Background(HoverColor)
        }
        
        if button.isFocused.Value() {
            baseStyle = baseStyle.BorderForeground(FocusColor)
        }
        
        if props.Disabled != nil && props.Disabled.Value() {
            baseStyle = baseStyle.Foreground(DisabledColor)
        }
        
        return baseStyle
    })
    
    return button
}

func (b *Button) Initialize() error {
    // Set up any effects or subscriptions
    NewEffect(func() {
        // This effect runs when props.Disabled changes
        if b.props.Disabled != nil && b.props.Disabled.Value() {
            b.isHovered.SetValue(false)
        }
    })
    
    return nil
}

func (b *Button) Update(msg tea.Msg) (tea.Cmd, error) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // Handle mouse events
        switch msg.Type {
        case tea.MouseMotion:
            // Check if mouse is over button
            if containsPoint(b.bounds, msg.X, msg.Y) {
                b.isHovered.SetValue(true)
            } else {
                b.isHovered.SetValue(false)
            }
            
        case tea.MouseLeft:
            // Handle click if not disabled
            if b.isHovered.Value() && 
               (b.props.Disabled == nil || !b.props.Disabled.Value()) {
                if b.props.OnClick != nil {
                    b.props.OnClick()
                }
            }
        }
        
    case tea.KeyMsg:
        // Handle keyboard events
        if b.isFocused.Value() && msg.String() == "enter" &&
           (b.props.Disabled == nil || !b.props.Disabled.Value()) {
            if b.props.OnClick != nil {
                b.props.OnClick()
            }
        }
    }
    
    return nil, nil
}

func (b *Button) Render() string {
    // Create styled button using computed style
    styledButton := b.style.Value().Render(b.props.Label)
    return styledButton
}

func (b *Button) Dispose() error {
    // Clean up any resources
    return nil
}
```

## Signal Flow Diagram

The following diagram illustrates how signals flow through the component tree:

```
  [User Input]
      │
      ▼
+------------+  Update  +------------+
| App (Root) |◄─────────| Bubble Tea |
+------------+          +------------+
      │
      │ Updates propagate down
      ▼
+------------------+     +------------------+
| Sidebar          |     | Content          |
| - selected       |     | - view           |
+------------------+     +------------------+
      │                        │
      │ Event callbacks        │ Props (signals)
      │ bubble up              │ flow down
      ▲                        ▼
+------------------+     +------------------+
| MenuItem         |     | Dashboard        |
| - onClick        |     | - data           |
+------------------+     +------------------+
                              │
                              ▼
                         +------------------+
                         | TaskList         |
                         | - tasks          |
                         +------------------+
                              │
                              ▼
                         +------------------+
                         | TaskItem         |
                         | - task           |
                         | - onComplete     |
                         +------------------+
```

## Reactivity Model Sketch

This diagram shows how signal dependencies form a reactive graph:

```
         +-------------------+
         | userProfile       |  Root Signal
         | (name, email)     |
         +-------------------+
                  │
                  ├─────────────────┬─────────────────┐
                  │                 │                 │
                  ▼                 ▼                 ▼
     +-------------------+ +-------------------+ +-------------------+
     | displayName       | | avatarInitials    | | profileComplete   |  Computed
     | Computed Signal   | | Computed Signal   | | Computed Signal   |  Signals
     +-------------------+ +-------------------+ +-------------------+
                  │                 │                 │
                  │                 │                 │
                  ▼                 ▼                 ▼
     +-------------------+ +-------------------+ +-------------------+
     | HeaderComponent   | | AvatarComponent   | | ProfileComponent  |  Components
     | Re-renders when   | | Re-renders when   | | Re-renders when   |  affected by
     | displayName       | | avatarInitials    | | profileComplete   |  signals
     | changes           | | changes           | | changes           |
     +-------------------+ +-------------------+ +-------------------+
```

## Event Flow Sketch

```
[Button Click Event]
       │
       ▼
+----------------+
| Button         | Component where event originated
| - onClick()    |
+----------------+
       │
       │ Call props.onClick() callback
       ▼
+----------------+
| Form           | Parent handling the event
| - handleSubmit |
+----------------+
       │
       │ Update form state and propagate event up
       ▼
+----------------+
| Dashboard      | Ancestor component
| - formSubmitted|
+----------------+
       │
       │ Update application state
       ▼
+----------------+
| App            | Root component
| - updateState  |
+----------------+
       │
       │ State changes trigger re-renders
       ▼
[Signal Updates]
       │
       ▼
[UI Re-renders]
```

## Next Steps

These proof-of-concept sketches demonstrate the fundamental component architecture, signal reactivity, and event flow in BubblyUI. The next steps are to:

1. Implement a minimal working prototype of the component system
2. Create a test harness for validating component behavior
3. Refine the API based on practical usage patterns
4. Develop a set of core components using this architecture
