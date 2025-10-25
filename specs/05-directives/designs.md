# Design Specification: Directives

## Component Hierarchy

```
Directives System
└── Template Enhancement Layer
    ├── Conditional Directives
    │   ├── If (conditional rendering)
    │   ├── ElseIf (chained conditions)
    │   ├── Else (fallback)
    │   └── Show (visibility toggle)
    ├── List Directives
    │   └── ForEach (iteration)
    ├── Binding Directives
    │   ├── Bind (two-way binding)
    │   ├── BindCheckbox (boolean binding)
    │   └── BindSelect (dropdown binding)
    └── Event Directives
        └── On (event handling)
```

---

## Architecture Overview

### System Layers

```
┌────────────────────────────────────────────────────────────┐
│                    Directives Layer                         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────┐    ┌──────────────┐   ┌─────────────┐ │
│  │  Conditional │───>│    List      │──>│  Binding    │ │
│  │  (If/Show)   │    │  (ForEach)   │   │  (Bind)     │ │
│  └──────────────┘    └──────────────┘   └─────────────┘ │
│         │                    │                   │        │
│         └────────────────────┼───────────────────┘        │
│                              ▼                             │
│                      ┌──────────────┐                     │
│                      │    Event     │                     │
│                      │    (On)      │                     │
│                      └──────────────┘                     │
│                                                            │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│              Component Template Rendering                  │
│  ┌────────────┐   ┌──────────────┐   ┌────────────────┐ │
│  │ RenderCtx  │──>│   Template   │──>│  Rendered UI   │ │
│  └────────────┘   └──────────────┘   └────────────────┘ │
└────────────────────────────────────────────────────────────┘
```

### Directive Execution Flow

```
Template function executes
    ↓
Directive created (If, ForEach, Bind, On, Show)
    ↓
Directive evaluates condition/collection/binding
    ↓
Directive calls render function(s)
    ↓
Directive composes output
    ↓
Render() returns string
    ↓
Template continues with next element
```

---

## Data Flow

### 1. If Directive Flow
```
Condition (Ref[bool] or bool)
    ↓
If directive evaluates condition
    ↓
If true:  Execute then branch
If false: Check ElseIf or Else
    ↓
Render function executes
    ↓
Return rendered string
```

### 2. ForEach Directive Flow
```
Collection ([]T)
    ↓
ForEach iterates over items
    ↓
For each item:
    ├─> Extract item and index
    ├─> Call render function with (item, index)
    ├─> Append result to output
    └─> Continue to next item
    ↓
Join all outputs
    ↓
Return complete string
```

### 3. Bind Directive Flow
```
Ref[T] value
    ↓
Bind directive creates input handler
    ↓
Render input with current value
    ↓
User changes input
    ↓
Input event triggers
    ↓
Event handler updates Ref
    ↓
Ref notifies watchers
    ↓
Component re-renders
    ↓
New value reflected in input
```

### 4. On Directive Flow
```
Event name ("click", "keypress", etc.)
    ↓
On directive registers handler
    ↓
Render element with event hook
    ↓
User triggers event
    ↓
Event captured by Bubbletea
    ↓
Message sent to Update()
    ↓
Handler matched and executed
    ↓
State may change, causing re-render
```

---

## State Management

### Directive Structures

#### If Directive
```go
type IfDirective struct {
    condition    bool
    thenBranch   func() string
    elseIfBranches []ElseIfBranch
    elseBranch   func() string
}

type ElseIfBranch struct {
    condition bool
    branch    func() string
}

func If(condition bool, then func() string) *IfDirective
func (d *IfDirective) ElseIf(condition bool, then func() string) *IfDirective
func (d *IfDirective) Else(then func() string) *IfDirective
func (d *IfDirective) Render() string
```

#### ForEach Directive
```go
type ForEachDirective[T any] struct {
    items      []T
    renderItem func(T, int) string
}

func ForEach[T any](items []T, render func(T, int) string) *ForEachDirective[T]
func (d *ForEachDirective[T]) Render() string
```

#### Bind Directive
```go
type BindDirective[T any] struct {
    ref       *Ref[T]
    inputType string
    convert   func(string) T
}

func Bind[T any](ref *Ref[T]) *BindDirective[T]
func BindCheckbox(ref *Ref[bool]) *BindDirective[bool]
func BindSelect[T any](ref *Ref[T], options []T) *SelectBindDirective[T]
func (d *BindDirective[T]) Render() string
```

#### On Directive
```go
type OnDirective struct {
    event         string
    handler       func(interface{})
    preventDefault bool
    stopPropagation bool
}

func On(event string, handler func(interface{})) *OnDirective
func (d *OnDirective) PreventDefault() *OnDirective
func (d *OnDirective) StopPropagation() *OnDirective
func (d *OnDirective) Render(content string) string
```

#### Show Directive
```go
type ShowDirective struct {
    visible    bool
    content    func() string
    transition bool
}

func Show(visible bool, content func() string) *ShowDirective
func (d *ShowDirective) WithTransition() *ShowDirective
func (d *ShowDirective) Render() string
```

---

## Type Definitions

### Core Directive Interface
```go
// Base directive interface
type Directive interface {
    Render() string
}

// Conditional directive
type ConditionalDirective interface {
    Directive
    ElseIf(condition bool, then func() string) ConditionalDirective
    Else(then func() string) ConditionalDirective
}

// Iteration directive
type IterationDirective[T any] interface {
    Directive
    Filter(predicate func(T) bool) IterationDirective[T]
    Map(transform func(T) T) IterationDirective[T]
}

// Binding directive
type BindingDirective[T any] interface {
    Directive
    WithValidator(validator func(T) bool) BindingDirective[T]
    OnChange(callback func(T)) BindingDirective[T]
}

// Event directive
type EventDirective interface {
    Directive
    PreventDefault() EventDirective
    StopPropagation() EventDirective
    Once() EventDirective
}
```

---

## API Contracts

### If Directive Contract
```go
// Simple If
If(condition, thenFunc).Render()

// If with Else
If(condition, thenFunc).Else(elseFunc).Render()

// If with ElseIf chain
If(cond1, func1).
    ElseIf(cond2, func2).
    ElseIf(cond3, func3).
    Else(func4).
    Render()

// Nested If
If(outerCondition, func() string {
    return If(innerCondition, innerThen).Else(innerElse).Render()
}).Render()
```

### ForEach Directive Contract
```go
// Basic iteration
ForEach(items, func(item T, index int) string {
    return fmt.Sprintf("%d: %v\n", index, item)
}).Render()

// With empty handling
func (ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    
    if len(items.Get()) == 0 {
        return "No items"
    }
    
    return ForEach(items.Get(), renderItem).Render()
}

// Nested ForEach
ForEach(categories, func(cat Category, i int) string {
    return ForEach(cat.Items, func(item string, j int) string {
        return fmt.Sprintf("  %s\n", item)
    }).Render()
}).Render()
```

### Bind Directive Contract
```go
// Text input
Setup(func(ctx *Context) {
    text := ctx.Ref("")
    ctx.On("input", func(value interface{}) {
        text.Set(value.(string))
    })
})

// Checkbox
Setup(func(ctx *Context) {
    checked := ctx.Ref(false)
    ctx.On("toggle", func(value interface{}) {
        checked.Set(value.(bool))
    })
})

// Select
Setup(func(ctx *Context) {
    selected := ctx.Ref("option1")
    options := []string{"option1", "option2", "option3"}
    ctx.On("select", func(value interface{}) {
        selected.Set(value.(string))
    })
})
```

### On Directive Contract
```go
// Simple event
On("click", func() {
    handleClick()
}).Render("Click Me")

// With event data
On("keypress", func(key string) {
    handleKey(key)
}).Render()

// With modifiers
On("submit", func() {
    submitForm()
}).PreventDefault().Render("Submit")

// Multiple events
content := "..."
content = On("mouseenter", handleEnter).Render(content)
content = On("mouseleave", handleLeave).Render(content)
```

### Show Directive Contract
```go
// Simple toggle
Show(visible.Get(), func() string {
    return "Conditionally visible content"
}).Render()

// With transition
Show(visible.Get(), func() string {
    return "Fades in/out"
}).WithTransition().Render()

// Nested in If
If(shouldRender.Get(), func() string {
    return Show(visible.Get(), content).Render()
}).Render()
```

---

## Implementation Details

### If Directive Implementation
```go
type IfDirective struct {
    condition      bool
    thenBranch     func() string
    elseIfBranches []ElseIfBranch
    elseBranch     func() string
}

type ElseIfBranch struct {
    condition bool
    branch    func() string
}

func If(condition bool, then func() string) *IfDirective {
    return &IfDirective{
        condition:      condition,
        thenBranch:     then,
        elseIfBranches: []ElseIfBranch{},
    }
}

func (d *IfDirective) ElseIf(condition bool, then func() string) *IfDirective {
    d.elseIfBranches = append(d.elseIfBranches, ElseIfBranch{
        condition: condition,
        branch:    then,
    })
    return d
}

func (d *IfDirective) Else(then func() string) *IfDirective {
    d.elseBranch = then
    return d
}

func (d *IfDirective) Render() string {
    // Check main condition
    if d.condition {
        return d.thenBranch()
    }
    
    // Check ElseIf branches
    for _, branch := range d.elseIfBranches {
        if branch.condition {
            return branch.branch()
        }
    }
    
    // Execute Else if present
    if d.elseBranch != nil {
        return d.elseBranch()
    }
    
    // No conditions met, return empty
    return ""
}
```

### ForEach Directive Implementation
```go
type ForEachDirective[T any] struct {
    items      []T
    renderItem func(T, int) string
}

func ForEach[T any](items []T, render func(T, int) string) *ForEachDirective[T] {
    return &ForEachDirective[T]{
        items:      items,
        renderItem: render,
    }
}

func (d *ForEachDirective[T]) Render() string {
    if len(d.items) == 0 {
        return ""
    }
    
    var output strings.Builder
    for i, item := range d.items {
        rendered := d.renderItem(item, i)
        output.WriteString(rendered)
    }
    
    return output.String()
}

// With optimized builder
func (d *ForEachDirective[T]) Render() string {
    if len(d.items) == 0 {
        return ""
    }
    
    // Pre-allocate buffer based on items
    output := make([]string, len(d.items))
    for i, item := range d.items {
        output[i] = d.renderItem(item, i)
    }
    
    return strings.Join(output, "")
}
```

### Bind Directive Implementation
```go
type BindDirective[T any] struct {
    ref       *Ref[T]
    inputType string
    component *componentImpl
}

func Bind[T any](ref *Ref[T]) *BindDirective[T] {
    return &BindDirective[T]{
        ref:       ref,
        inputType: "text",
    }
}

func (d *BindDirective[T]) Render() string {
    // Register input handler
    handlerID := generateID()
    d.component.registerInputHandler(handlerID, func(value string) {
        // Convert string to T and update ref
        converted := d.convert(value)
        d.ref.Set(converted)
    })
    
    // Render input with current value and handler ID
    return fmt.Sprintf("[Input: %v (id:%s)]", d.ref.Get(), handlerID)
}

// Type-specific converters
func convertString(value string) string { return value }
func convertInt(value string) int      { i, _ := strconv.Atoi(value); return i }
func convertFloat(value string) float64 { f, _ := strconv.ParseFloat(value, 64); return f }
func convertBool(value string) bool    { return value == "true" || value == "1" }
```

### On Directive Implementation
```go
type OnDirective struct {
    event           string
    handler         func(interface{})
    preventDefault  bool
    stopPropagation bool
    once            bool
    component       *componentImpl
}

func On(event string, handler func(interface{})) *OnDirective {
    return &OnDirective{
        event:   event,
        handler: handler,
    }
}

func (d *OnDirective) PreventDefault() *OnDirective {
    d.preventDefault = true
    return d
}

func (d *OnDirective) StopPropagation() *OnDirective {
    d.stopPropagation = true
    return d
}

func (d *OnDirective) Once() *OnDirective {
    d.once = true
    return d
}

func (d *OnDirective) Render(content string) string {
    // Register event handler
    handlerID := generateID()
    d.component.registerEventHandler(d.event, handlerID, d.handler, EventOptions{
        PreventDefault:  d.preventDefault,
        StopPropagation: d.stopPropagation,
        Once:            d.once,
    })
    
    // Render content with event marker
    return fmt.Sprintf("[Event:%s:%s]%s", d.event, handlerID, content)
}
```

### Show Directive Implementation
```go
type ShowDirective struct {
    visible    bool
    content    func() string
    transition bool
}

func Show(visible bool, content func() string) *ShowDirective {
    return &ShowDirective{
        visible: visible,
        content: content,
    }
}

func (d *ShowDirective) WithTransition() *ShowDirective {
    d.transition = true
    return d
}

func (d *ShowDirective) Render() string {
    if !d.visible {
        if d.transition {
            // Return content with hidden style
            return fmt.Sprintf("[Hidden]%s", d.content())
        }
        // Don't render at all
        return ""
    }
    
    return d.content()
}
```

---

## Integration with Component System

### Template Integration
```go
// Directives used in template function
NewComponent("MyComponent").
    Setup(func(ctx *Context) {
        items := ctx.Ref([]string{"A", "B", "C"})
        visible := ctx.Ref(true)
        
        ctx.Expose("items", items)
        ctx.Expose("visible", visible)
    }).
    Template(func(ctx RenderContext) string {
        items := ctx.Get("items").(*Ref[[]string])
        visible := ctx.Get("visible").(*Ref[bool])
        
        return Show(visible.Get(), func() string {
            return ForEach(items.Get(), func(item string, i int) string {
                return fmt.Sprintf("%d. %s\n", i+1, item)
            }).Render()
        }).Render()
    }).
    Build()
```

### Bubbletea Message Integration
```go
// On directive sends messages
type InputChangeMsg struct {
    HandlerID string
    Value     string
}

type EventMsg struct {
    Event     string
    HandlerID string
    Data      interface{}
}

func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case InputChangeMsg:
        if handler, ok := c.inputHandlers[msg.HandlerID]; ok {
            handler(msg.Value)
        }
        
    case EventMsg:
        if handlers, ok := c.eventHandlers[msg.Event]; ok {
            for _, h := range handlers {
                if h.ID == msg.HandlerID {
                    h.Handler(msg.Data)
                    if h.Options.Once {
                        delete(handlers, h.ID)
                    }
                }
            }
        }
    }
    
    return c, nil
}
```

---

## Performance Optimizations

### 1. ForEach Diff Algorithm
```go
// Track previous render
type ForEachDirective[T any] struct {
    items       []T
    renderItem  func(T, int) string
    lastItems   []T
    lastOutputs []string
}

func (d *ForEachDirective[T]) Render() string {
    // Check if items changed
    if slicesEqual(d.items, d.lastItems) {
        // Return cached output
        return strings.Join(d.lastOutputs, "")
    }
    
    // Re-render only changed items
    outputs := make([]string, len(d.items))
    for i, item := range d.items {
        // Check if item at index i changed
        if i < len(d.lastItems) && itemEqual(item, d.lastItems[i]) {
            outputs[i] = d.lastOutputs[i]
        } else {
            outputs[i] = d.renderItem(item, i)
        }
    }
    
    // Cache for next render
    d.lastItems = d.items
    d.lastOutputs = outputs
    
    return strings.Join(outputs, "")
}
```

### 2. Directive Pooling
```go
var ifDirectivePool = sync.Pool{
    New: func() interface{} {
        return &IfDirective{}
    },
}

func If(condition bool, then func() string) *IfDirective {
    d := ifDirectivePool.Get().(*IfDirective)
    d.condition = condition
    d.thenBranch = then
    d.elseIfBranches = d.elseIfBranches[:0]
    d.elseBranch = nil
    return d
}

func (d *IfDirective) Render() string {
    result := d.render()
    ifDirectivePool.Put(d)
    return result
}
```

### 3. String Builder Optimization
```go
// Use sync.Pool for string builders
var builderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

func (d *ForEachDirective[T]) Render() string {
    builder := builderPool.Get().(*strings.Builder)
    builder.Reset()
    
    for i, item := range d.items {
        builder.WriteString(d.renderItem(item, i))
    }
    
    result := builder.String()
    builderPool.Put(builder)
    return result
}
```

---

## Error Handling

### Error Types
```go
var (
    ErrInvalidDirectiveUsage = errors.New("invalid directive usage")
    ErrBindTypeMismatch      = errors.New("bind type mismatch")
    ErrEmptyForEach          = errors.New("forEach requires non-empty collection")
    ErrInvalidEventName      = errors.New("invalid event name")
)
```

### Error Scenarios
1. **Bind on non-input:** Compile-time or runtime error
2. **Nested directive errors:** Propagate up with context
3. **Render function panics:** Recover and return error message
4. **Type conversion errors:** Log warning, use zero value

---

## Testing Strategy

### Unit Tests
```go
func TestIfDirective(t *testing.T)
func TestIfElseIfElse(t *testing.T)
func TestForEachDirective(t *testing.T)
func TestForEachEmpty(t *testing.T)
func TestBindDirective(t *testing.T)
func TestOnDirective(t *testing.T)
func TestShowDirective(t *testing.T)
func TestNestedDirectives(t *testing.T)
```

### Integration Tests
```go
func TestDirectivesInTemplate(t *testing.T)
func TestDirectiveWithReactivity(t *testing.T)
func TestDirectiveWithLifecycle(t *testing.T)
func TestMultipleDirectives(t *testing.T)
```

---

## Example Usage

### Complex Template with Multiple Directives
```go
Template(func(ctx RenderContext) string {
    users := ctx.Get("users").(*Ref[[]User])
    filter := ctx.Get("filter").(*Ref[string])
    showList := ctx.Get("showList").(*Ref[bool])
    
    return Show(showList.Get(), func() string {
        filtered := filterUsers(users.Get(), filter.Get())
        
        return If(len(filtered) > 0,
            func() string {
                return ForEach(filtered, func(user User, i int) string {
                    return fmt.Sprintf("%d. %s - %s\n", 
                        i+1, user.Name, user.Email)
                }).Render()
            },
        ).Else(func() string {
            return "No users match the filter"
        }).Render()
    }).Render()
})
```

---

## Future Enhancements

1. **Custom Directives:** User-defined directive system
2. **Directive Middleware:** Intercept directive execution
3. **Directive Caching:** Cache directive results
4. **Virtual DOM:** Diff-based re-rendering
5. **Transitions:** Built-in animation support
6. **Directive Composition Helpers:** Higher-order directives
