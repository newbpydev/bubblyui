# User Workflow: Component Model

## Primary User Journey

### Journey: Developer Creates First Component

1. **Entry Point**: Developer wants to create a reusable button component
   - System response: Framework provides ComponentBuilder API
   - UI update: N/A (development phase)

2. **Step 1**: Create component with name
   ```go
   button := NewComponent("Button")
   ```
   - System response: ComponentBuilder instance created
   - Developer sees: Fluent API ready for chaining
   - Ready for: Configuration

3. **Step 2**: Define component props
   ```go
   type ButtonProps struct {
       Label    string
       Disabled bool
       OnClick  func()
   }
   
   button.Props(ButtonProps{
       Label: "Submit",
       Disabled: false,
   })
   ```
   - System response: Props stored in builder
   - Developer sees: Type-safe props structure
   - Ready for: Template definition

4. **Step 3**: Define template function
   ```go
   button.Template(func(ctx RenderContext) string {
       props := ctx.Props().(ButtonProps)
       
       style := lipgloss.NewStyle().
           Background(lipgloss.Color("63")).
           Foreground(lipgloss.Color("230")).
           Padding(0, 2).
           Bold(true)
       
       if props.Disabled {
           style = style.Faint(true)
       }
       
       return style.Render(props.Label)
   })
   ```
   - System response: Template function stored
   - Developer sees: Lipgloss styling integrated
   - Ready for: Building

5. **Step 4**: Build component
   ```go
   component, err := button.Build()
   if err != nil {
       log.Fatal(err)
   }
   ```
   - System response: Validates and creates component
   - Developer sees: Working component or clear error
   - Ready for: Usage in app

6. **Step 5**: Use component in Bubbletea app
   ```go
   func main() {
       p := tea.NewProgram(component)
       if _, err := p.Run(); err != nil {
           log.Fatal(err)
       }
   }
   ```
   - System response: Component runs as Bubbletea model
   - UI update: Button renders in terminal
   - Developer sees: Working TUI application

7. **Completion**: Reusable component working
   - End state: Component can be instantiated with different props
   - Performance: Renders in < 5ms
   - Developer satisfaction: Easy to create and use

---

## Alternative Paths

### Scenario A: Component with State (Counter)

1. **Developer adds Setup function**
   ```go
   counter := NewComponent("Counter").
       Setup(func(ctx *Context) {
           count := ctx.Ref(0)
           ctx.Expose("count", count)
           
           ctx.On("increment", func(_ interface{}) {
               count.Set(count.Get() + 1)
           })
       })
   ```
   - System: Setup executes on Init()
   - State: Reactive Ref created
   - Handlers: Event listeners registered

2. **Developer defines template accessing state**
   ```go
   .Template(func(ctx RenderContext) string {
       count := ctx.Get("count").(*Ref[int])
       return fmt.Sprintf("Count: %d", count.Get())
   })
   ```
   - System: Template accesses exposed state
   - Rendering: Current count displayed
   - Reactivity: Updates automatically on state change

3. **User interacts with component**
   - Keypress triggers "increment" event
   - Handler updates Ref
   - Component re-renders
   - New count displayed

**Use Case:** Interactive components with local state

### Scenario B: Parent-Child Communication

1. **Developer creates child button**
   ```go
   button := NewComponent("Button").
       Setup(func(ctx *Context) {
           ctx.On("keypress", func(data interface{}) {
               if data.(string) == "enter" {
                   ctx.Emit("clicked", map[string]interface{}{
                       "timestamp": time.Now(),
                   })
               }
           })
       }).
       Template(func(ctx RenderContext) string {
           return "[Submit]"
       }).
       Build()
   ```

2. **Developer creates parent form**
   ```go
   form := NewComponent("Form").
       Children(button).
       Setup(func(ctx *Context) {
           children := ctx.Children()
           children[0].On("clicked", func(data interface{}) {
               // Handle submission
               fmt.Println("Form submitted!")
           })
       }).
       Template(func(ctx RenderContext) string {
           return ctx.RenderChild(ctx.Children()[0])
       }).
       Build()
   ```

3. **User submits form**
   - User presses Enter on button
   - Button emits "clicked" event
   - Event bubbles to parent
   - Parent handler executes
   - Form processes submission

**Use Case:** Complex UI with component communication

### Scenario C: Component with Multiple Children

1. **Developer creates layout component**
   ```go
   layout := NewComponent("Layout").
       Children(header, content, footer).
       Template(func(ctx RenderContext) string {
           parts := []string{}
           for _, child := range ctx.Children() {
               parts = append(parts, ctx.RenderChild(child))
           }
           return lipgloss.JoinVertical(
               lipgloss.Left,
               parts...,
           )
       }).
       Build()
   ```

2. **System renders component tree**
   - Layout renders
   - Each child renders in order
   - Results joined vertically
   - Complete UI displayed

**Use Case:** Complex layouts with multiple sections

### Scenario D: Event Bubbling Through Component Tree

1. **Developer creates nested button in form in dialog**
   ```go
   // Deep child: Submit button
   submitButton := NewComponent("SubmitButton").
       Setup(func(ctx *Context) {
           ctx.On("keypress", func(data interface{}) {
               if data.(string) == "enter" {
                   // Emit custom event that will bubble
                   ctx.Emit("formSubmit", map[string]interface{}{
                       "source": "submitButton",
                       "timestamp": time.Now(),
                   })
               }
           })
       }).
       Template(func(ctx RenderContext) string {
           return "[Submit]"
       }).
       Build()
   
   // Middle level: Form component
   form := NewComponent("Form").
       Children(submitButton).
       Setup(func(ctx *Context) {
           // Form can handle or let bubble
           ctx.On("formSubmit", func(data interface{}) {
               fmt.Println("Form validating...")
               // Optionally stop propagation here if validation fails
               // Otherwise, event continues bubbling
           })
       }).
       Template(func(ctx RenderContext) string {
           return "Form:\n" + ctx.RenderChild(ctx.Children()[0])
       }).
       Build()
   
   // Top level: Dialog component
   dialog := NewComponent("Dialog").
       Children(form).
       Setup(func(ctx *Context) {
           // Dialog handles final submission
           ctx.On("formSubmit", func(data interface{}) {
               eventData := data.(map[string]interface{})
               fmt.Printf("Dialog closing after submit from: %s\n", 
                   eventData["source"])
               // Close dialog after successful submission
           })
       }).
       Template(func(ctx RenderContext) string {
           return "Dialog:\n" + ctx.RenderChild(ctx.Children()[0])
       }).
       Build()
   ```

2. **Event bubbling flow**
   - User presses Enter on submit button
   - submitButton emits "formSubmit" event
   - Event bubbles to Form component
   - Form validates and lets event continue
   - Event bubbles to Dialog component
   - Dialog handles final submission
   - All three components can coordinate behavior

3. **Benefits demonstrated**
   - **Decoupling**: Button doesn't know about Dialog
   - **Layered Handling**: Each level can react appropriately
   - **Coordinated State**: Multiple components respond to one event
   - **Clean Architecture**: Natural parent-child communication

**Use Case:** Multi-level component trees with coordinated event handling

### Scenario E: Event Bubbling with Stop Propagation

1. **Developer creates form with cancel functionality**
   ```go
   cancelButton := NewComponent("CancelButton").
       Setup(func(ctx *Context) {
           ctx.On("click", func(data interface{}) {
               // Emit event with stop propagation intent
               ctx.Emit("actionCancelled", map[string]interface{}{
                   "stopBubbling": true,
               })
           })
       }).
       Build()
   
   form := NewComponent("Form").
       Children(cancelButton).
       Setup(func(ctx *Context) {
           ctx.On("actionCancelled", func(data interface{}) {
               eventData := data.(map[string]interface{})
               
               // Form handles cancellation
               fmt.Println("Form handling cancellation...")
               
               // Check if we should stop bubbling
               if stop, ok := eventData["stopBubbling"].(bool); ok && stop {
                   // Stop propagation - don't notify parent dialog
                   // (Implementation: handler can set Event.Stopped flag)
                   return
               }
           })
       }).
       Build()
   
   dialog := NewComponent("Dialog").
       Children(form).
       Setup(func(ctx *Context) {
           ctx.On("actionCancelled", func(data interface{}) {
               // This won't be called if form stopped propagation
               fmt.Println("Dialog closing due to cancellation")
           })
       }).
       Build()
   ```

2. **Controlled propagation flow**
   - Cancel button emits event
   - Form handles locally
   - Form stops propagation
   - Dialog never receives event
   - Each component has control over bubbling

**Use Case:** Fine-grained control over event propagation

### Scenario F: Dynamic Component Creation

1. **Developer creates component factory**
   ```go
   func CreateListItem(text string, index int) *Component {
       return NewComponent(fmt.Sprintf("Item-%d", index)).
           Props(ListItemProps{Text: text, Index: index}).
           Template(func(ctx RenderContext) string {
               props := ctx.Props().(ListItemProps)
               return fmt.Sprintf("%d. %s", props.Index, props.Text)
           }).
           Build()
   }
   ```

2. **Developer creates dynamic list**
   ```go
   items := []string{"Apple", "Banana", "Cherry"}
   children := make([]*Component, len(items))
   
   for i, item := range items {
       children[i] = CreateListItem(item, i+1)
   }
   
   list := NewComponent("List").
       Children(children...).
       Template(func(ctx RenderContext) string {
           outputs := []string{}
           for _, child := range ctx.Children() {
               outputs = append(outputs, ctx.RenderChild(child))
           }
           return strings.Join(outputs, "\n")
       }).
       Build()
   ```

**Use Case:** Dynamic lists based on data

---

## Error Handling Flows

### Error 1: Missing Required Template
- **Trigger**: Build() called without Template()
- **User sees**: 
  ```
  Error: component validation failed: template is required
  ```
- **Recovery**:
  1. Add Template() call
  2. Define render function
  3. Call Build() again

**Example:**
```go
// ❌ Error
component, err := NewComponent("Test").Build()
// err: "template is required"

// ✅ Fixed
component, err := NewComponent("Test").
    Template(func(ctx RenderContext) string {
        return "Hello"
    }).
    Build()
```

### Error 2: Props Type Mismatch
- **Trigger**: Template tries to cast props to wrong type
- **User sees**: Panic with type assertion failure
- **Recovery**:
  1. Check props type definition
  2. Ensure template casts to correct type
  3. Use type switch for flexibility

**Example:**
```go
// ❌ Wrong type
.Template(func(ctx RenderContext) string {
    props := ctx.Props().(WrongType)  // Panic!
    return props.Label
})

// ✅ Correct type
.Template(func(ctx RenderContext) string {
    props := ctx.Props().(ButtonProps)
    return props.Label
})

// ✅ Safe with type switch
.Template(func(ctx RenderContext) string {
    switch p := ctx.Props().(type) {
    case ButtonProps:
        return p.Label
    default:
        return "Unknown"
    }
})
```

### Error 3: State Access Before Expose
- **Trigger**: Template calls ctx.Get() for non-existent key
- **User sees**: Returns nil, potential panic
- **Recovery**:
  1. Ensure Setup() calls ctx.Expose()
  2. Check key spelling
  3. Add nil check in template

**Example:**
```go
// ❌ Not exposed
.Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    // Forgot to expose!
}).
Template(func(ctx RenderContext) string {
    count := ctx.Get("count")  // Returns nil!
    return fmt.Sprintf("%d", count)  // Panic!
})

// ✅ Properly exposed
.Setup(func(ctx *Context) {
    count := ctx.Ref(0)
    ctx.Expose("count", count)  // Expose it!
}).
Template(func(ctx RenderContext) string {
    count := ctx.Get("count").(*Ref[int])
    return fmt.Sprintf("%d", count.Get())
})
```

### Error 4: Event Handler Panic
- **Trigger**: Handler code panics during execution
- **User sees**: Component recovers, logs error, continues
- **Recovery**: Automatic (component continues working)
- **Developer action**: Fix handler code, redeploy

**Example:**
```go
.Setup(func(ctx *Context) {
    ctx.On("action", func(data interface{}) {
        // Handler panics
        panic("something went wrong")
    })
})

// System behavior:
// 1. Catch panic
// 2. Log error with stack trace
// 3. Continue processing other events
// 4. Component remains functional
```

### Error 5: Circular Component Reference
- **Trigger**: Component A contains B, B contains A
- **User sees**: Stack overflow or max depth error
- **Recovery**:
  1. Redesign component hierarchy
  2. Use indirection (interface or lazy loading)
  3. Flatten structure

**Example:**
```go
// ❌ Circular reference
a := NewComponent("A").Children(b).Build()
b := NewComponent("B").Children(a).Build()  // Can't reference before creation

// ✅ Proper hierarchy
parent := NewComponent("Parent").Children(childA, childB).Build()
```

---

## State Transitions

### Component Lifecycle States
```
Created (NewComponent)
    ↓
Configured (Props, Setup, Template set)
    ↓
Built (Build() called)
    ↓
Initialized (Init() called by Bubbletea)
    ↓
Mounted (Setup executed, ready for interaction)
    ↓
Active (Rendering, handling events)
    ↓ (when removed)
Unmounted (cleanup, children unmounted)
```

### State Update Cycle
```
User Input
    ↓
Bubbletea Message
    ↓
Component.Update()
    ↓
Event Handler Matched
    ↓
Handler Executes
    ↓
Ref.Set() Called
    ↓
Watchers Notified
    ↓
Re-render Scheduled
    ↓
Component.View() Called
    ↓
Template Executes
    ↓
New UI Rendered
```

### Props Update Flow
```
Parent Changes Props
    ↓
Parent Calls child.SetProps(newProps)
    ↓
Child Validates Props
    ↓
Child State Updates
    ↓
Child Re-renders
    ↓
Updated UI
```

---

## Integration Points

### Connected to: Reactivity System (Feature 01)
- **Uses:** Ref, Computed, Watch
- **Flow:** Setup creates Refs → Template accesses Refs → Watchers trigger re-renders
- **Data:** Component state stores reactive primitives

### Connected to: Lifecycle Hooks (Feature 03)
- **Uses:** onMounted, onUpdated, onUnmounted
- **Flow:** Component calls hooks at appropriate times
- **Data:** Lifecycle callbacks registered in Setup

### Connected to: Composition API (Feature 04)
- **Uses:** Composable functions
- **Flow:** Setup calls composables → Composables return Refs/handlers
- **Data:** Shared logic via composables

### Connected to: Directives (Feature 05)
- **Uses:** If(), ForEach(), Bind()
- **Flow:** Template uses directives → Directives enhance rendering
- **Data:** Directives access component context

### Connected to: Built-in Components (Feature 06)
- **Uses:** All built-in components
- **Flow:** Developers instantiate built-in components → Use in trees
- **Data:** Props passed to built-in components

---

## Performance Considerations

### Optimization: Render Caching
**Scenario:** Component renders frequently but state unchanged

**Solution:**
```go
type componentImpl struct {
    lastRender  string
    stateHash   uint64
}

func (c *componentImpl) View() string {
    currentHash := c.calculateStateHash()
    if currentHash == c.stateHash {
        return c.lastRender  // Return cached
    }
    
    c.stateHash = currentHash
    c.lastRender = c.template(c.createRenderContext())
    return c.lastRender
}
```

### Optimization: Lazy Children Rendering
**Scenario:** 1000+ items in list, only 20 visible

**Solution:**
```go
.Template(func(ctx RenderContext) string {
    visible := ctx.Children()[scrollOffset:scrollOffset+20]
    outputs := make([]string, len(visible))
    for i, child := range visible {
        outputs[i] = ctx.RenderChild(child)
    }
    return strings.Join(outputs, "\n")
})
```

### Performance Targets
- Component creation: < 1ms
- Simple render: < 5ms
- Complex render (10 children): < 20ms
- Event handling: < 1ms
- Props update: < 1ms

---

## Common Patterns

### Pattern 1: Controlled Component
```go
// Props contain value and onChange
type InputProps struct {
    Value    string
    OnChange func(string)
}

input := NewComponent("Input").
    Props(InputProps{
        Value: "initial",
        OnChange: func(val string) {
            // Parent controls value
        },
    }).
    Setup(func(ctx *Context) {
        ctx.On("keypress", func(data interface{}) {
            props := ctx.Props().(InputProps)
            newValue := props.Value + data.(string)
            if props.OnChange != nil {
                props.OnChange(newValue)
            }
        })
    }).
    Template(func(ctx RenderContext) string {
        props := ctx.Props().(InputProps)
        return fmt.Sprintf("[%s]", props.Value)
    }).
    Build()
```

### Pattern 2: Uncontrolled Component
```go
// Component manages own state
input := NewComponent("Input").
    Setup(func(ctx *Context) {
        value := ctx.Ref("")
        ctx.Expose("value", value)
        
        ctx.On("keypress", func(data interface{}) {
            current := value.Get()
            value.Set(current + data.(string))
        })
    }).
    Template(func(ctx RenderContext) string {
        value := ctx.Get("value").(*Ref[string])
        return fmt.Sprintf("[%s]", value.Get())
    }).
    Build()
```

### Pattern 3: Render Props
```go
type ListProps struct {
    Items      []string
    RenderItem func(item string, index int) string
}

list := NewComponent("List").
    Props(ListProps{
        Items: []string{"A", "B", "C"},
        RenderItem: func(item string, index int) string {
            return fmt.Sprintf("%d. %s", index+1, item)
        },
    }).
    Template(func(ctx RenderContext) string {
        props := ctx.Props().(ListProps)
        outputs := make([]string, len(props.Items))
        for i, item := range props.Items {
            outputs[i] = props.RenderItem(item, i)
        }
        return strings.Join(outputs, "\n")
    }).
    Build()
```

---

## Testing Workflow

### Unit Test: Component Creation
```go
func TestComponent_Creation(t *testing.T) {
    // Arrange
    builder := NewComponent("Test")
    
    // Act
    component, err := builder.
        Template(func(ctx RenderContext) string {
            return "Hello"
        }).
        Build()
    
    // Assert
    require.NoError(t, err)
    assert.NotNil(t, component)
    assert.Equal(t, "Test", component.Name())
}
```

### Unit Test: Props Access
```go
func TestComponent_PropsAccess(t *testing.T) {
    // Arrange
    type TestProps struct {
        Label string
    }
    
    // Act
    component, _ := NewComponent("Test").
        Props(TestProps{Label: "Click"}).
        Template(func(ctx RenderContext) string {
            props := ctx.Props().(TestProps)
            return props.Label
        }).
        Build()
    
    // Assert
    view := component.View()
    assert.Equal(t, "Click", view)
}
```

### Integration Test: Parent-Child
```go
func TestComponent_ParentChild(t *testing.T) {
    // Arrange
    child := NewComponent("Child").
        Template(func(ctx RenderContext) string {
            return "Child"
        }).
        Build()
    
    parent := NewComponent("Parent").
        Children(child).
        Template(func(ctx RenderContext) string {
            return "Parent: " + ctx.RenderChild(ctx.Children()[0])
        }).
        Build()
    
    // Act
    view := parent.View()
    
    // Assert
    assert.Contains(t, view, "Parent: Child")
}
```

---

## Documentation for Users

### Quick Start
1. Import: `import "github.com/newbpydev/bubblyui/pkg/bubbly"`
2. Create: `NewComponent("name")`
3. Configure: `.Props()`, `.Setup()`, `.Template()`
4. Build: `.Build()`
5. Use: Pass to `tea.NewProgram()`

### Best Practices
- Always provide Template()
- Use Setup() for state and handlers
- Expose state needed in template
- Keep templates pure (no side effects)
- Use Props for configuration
- Use Events for communication
- Test components in isolation

### Troubleshooting
- **Component doesn't render?** Check Template() is defined
- **Props not accessible?** Verify type assertion
- **State not updating?** Ensure Expose() was called
- **Events not firing?** Check event name spelling
- **Children not rendering?** Call ctx.RenderChild()
