# Feature Name: Directives

## Feature ID
05-directives

## Overview
Implement Vue-inspired directives that provide declarative ways to manipulate the rendered output within component templates. Directives are special functions that enhance template rendering with common patterns like conditional rendering (If), list rendering (ForEach), two-way binding (Bind), and event handling (On). This feature builds on the component model (feature 02) and enables cleaner, more expressive templates.

## User Stories
- As a **developer**, I want to conditionally render elements so that I can show/hide UI based on state
- As a **developer**, I want to render lists of items so that I can display dynamic data collections
- As a **developer**, I want two-way binding for inputs so that form handling is simpler
- As a **developer**, I want declarative event handling so that my templates are more readable
- As a **developer**, I want to control element visibility so that I can toggle UI without removing from DOM
- As a **developer**, I want type-safe directives so that I catch errors at compile time

## Functional Requirements

### 1. If Directive (Conditional Rendering)
1.1. Render element only if condition is truthy  
1.2. Remove element from render output if condition is falsy  
1.3. Support Else and ElseIf for chained conditions  
1.4. Work with Ref[bool] and computed boolean values  
1.5. Efficient re-rendering on condition changes  

### 2. ForEach Directive (List Rendering)
2.1. Iterate over slices and arrays  
2.2. Provide item, index, and key to render function  
2.3. Efficient updates (minimal re-renders)  
2.4. Support nested ForEach  
2.5. Handle empty collections gracefully  

### 3. Bind Directive (Two-Way Binding)
3.1. Synchronize Ref value with input element  
3.2. Update Ref when input changes  
3.3. Update input when Ref changes  
3.4. Support text, number, checkbox, select inputs  
3.5. Type-safe binding with generics  

### 4. On Directive (Event Handling)
4.1. Attach event handlers declaratively  
4.2. Support keyboard, mouse, and custom events  
4.3. Event modifiers (prevent default, stop propagation)  
4.4. Type-safe event handlers  
4.5. Auto-cleanup on component unmount  

### 5. Show Directive (Visibility Toggle)
5.1. Toggle element visibility via CSS display property  
5.2. Keep element in DOM (unlike If)  
5.3. Fast toggle performance  
5.4. Work with Ref[bool] and computed values  

### 6. Directive Composition
6.1. Multiple directives on same element  
6.2. Directive execution order defined  
6.3. Directives can access component context  
6.4. Directives can access other directive results  

## Non-Functional Requirements

### Performance
- Directive execution: < 100ns overhead per directive
- ForEach rendering: < 1ms for 100 items
- If/Show toggle: < 50ns
- Bind synchronization: < 100ns
- Minimal re-renders on updates

### Accessibility
- Directives maintain semantic HTML
- Keyboard navigation preserved
- Screen reader compatibility
- Focus management

### Security
- Input sanitization in Bind
- XSS prevention in templates
- Event handler validation

### Type Safety
- **Strict typing:** All directives type-safe
- **Generic directives:** ForEach[T], Bind[T]
- **Type-safe callbacks:** Render functions typed
- **No `any`:** Use interfaces with constraints
- **Compile-time validation:** Catch errors before runtime

## Acceptance Criteria

### If Directive
- [ ] Conditionally renders elements
- [ ] Supports Else and ElseIf
- [ ] Works with Ref and Computed
- [ ] Efficient re-rendering
- [ ] Type-safe conditions

### ForEach Directive
- [ ] Iterates over collections
- [ ] Provides item and index
- [ ] Efficient diff algorithm
- [ ] Handles empty collections
- [ ] Nested ForEach works
- [ ] Type-safe iterations

### Bind Directive
- [ ] Two-way synchronization
- [ ] Works with text inputs
- [ ] Works with checkboxes
- [ ] Works with select elements
- [ ] Type-safe binding
- [ ] Updates propagate correctly

### On Directive
- [ ] Attaches event handlers
- [ ] Supports common events
- [ ] Event modifiers work
- [ ] Type-safe handlers
- [ ] Auto-cleanup on unmount

### Show Directive
- [ ] Toggles visibility
- [ ] Keeps element in DOM
- [ ] Fast toggle performance
- [ ] Works with reactive values

### General
- [ ] Test coverage > 80%
- [ ] All directives documented
- [ ] Examples provided
- [ ] Performance acceptable

## Dependencies
- **Requires:** 02-component-model (templates, render context)
- **Uses:** 01-reactivity-system (Ref, Computed for values)
- **Enables:** Cleaner templates, less boilerplate

## Edge Cases

### 1. If with Null/Undefined Values
**Scenario:** If directive receives null or undefined  
**Handling:** Treat as falsy, don't render

### 2. ForEach with Changing Keys
**Scenario:** Item keys change between renders  
**Handling:** Re-render affected items, maintain others

### 3. ForEach with Empty Collection
**Scenario:** Slice has zero items  
**Handling:** Render nothing, no errors

### 4. Bind with Non-Input Elements
**Scenario:** Bind used on non-input element  
**Handling:** Error with clear message

### 5. Multiple On Directives Same Event
**Scenario:** Two On("click") on same element  
**Handling:** Both handlers execute in order

### 6. Nested If Directives
**Scenario:** If inside ForEach inside If  
**Handling:** All directives evaluate correctly

### 7. Bind Type Mismatch
**Scenario:** Bind[int] on text input  
**Handling:** Parse/convert or error with message

## Testing Requirements

### Unit Tests (80%+ coverage)
- Each directive independently
- Edge cases for each directive
- Type safety validation
- Performance benchmarks

### Integration Tests
- Directives in templates
- Multiple directives together
- Directive with reactivity
- Directive with lifecycle

### Example Usage
- Form with Bind
- List with ForEach
- Conditional UI with If
- Interactive events with On

## Atomic Design Level
**Template Enhancement** (Works with all component levels)

Directives enhance templates at any level:
- Atoms: If/Show on buttons, Bind on inputs
- Molecules: ForEach for lists, On for interactions
- Organisms: Combined directives for complex UI

## Related Components
- Uses: Component templates (RenderContext)
- Uses: Reactivity (Ref, Computed)
- Enables: Cleaner template code
- Simplifies: Common template patterns

## Technical Constraints
- Directives execute within template render
- Cannot create new reactive state
- Must be synchronous
- Limited to template scope
- No side effects in directive evaluation

## API Design

### If Directive
```go
// In template function
Template(func(ctx RenderContext) string {
    return bubbly.If(condition.Get(),
        func() string {
            return "Condition is true"
        },
    ).Else(func() string {
        return "Condition is false"
    }).Render()
})

// With ElseIf
Template(func(ctx RenderContext) string {
    status := ctx.Get("status").(*Ref[string])
    
    return bubbly.If(status.Get() == "loading",
        func() string { return "Loading..." },
    ).ElseIf(status.Get() == "error",
        func() string { return "Error occurred" },
    ).Else(func() string {
        return "Success"
    }).Render()
})
```

### ForEach Directive
```go
// Basic ForEach
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    
    return bubbly.ForEach(items.Get(), func(item string, index int) string {
        return fmt.Sprintf("%d. %s\n", index+1, item)
    }).Render()
})

// ForEach with struct
Template(func(ctx RenderContext) string {
    users := ctx.Get("users").(*Ref[[]User])
    
    return bubbly.ForEach(users.Get(), func(user User, index int) string {
        return fmt.Sprintf("User: %s (%s)\n", user.Name, user.Email)
    }).Render()
})

// Nested ForEach
Template(func(ctx RenderContext) string {
    categories := ctx.Get("categories").(*Ref[[]Category])
    
    return bubbly.ForEach(categories.Get(), func(cat Category, i int) string {
        items := bubbly.ForEach(cat.Items, func(item string, j int) string {
            return fmt.Sprintf("  - %s\n", item)
        }).Render()
        
        return fmt.Sprintf("%s:\n%s", cat.Name, items)
    }).Render()
})
```

### Bind Directive
```go
// Text input binding
Setup(func(ctx *Context) {
    name := ctx.Ref("")
    ctx.Expose("name", name)
    
    ctx.Expose("nameInput", bubbly.Bind(name))
})

Template(func(ctx RenderContext) string {
    nameInput := ctx.Get("nameInput").(bubbly.BindDirective[string])
    return nameInput.Render()
})

// Checkbox binding
Setup(func(ctx *Context) {
    agreed := ctx.Ref(false)
    ctx.Expose("agreed", agreed)
    ctx.Expose("agreeCheckbox", bubbly.BindCheckbox(agreed))
})

// Select binding
Setup(func(ctx *Context) {
    selected := ctx.Ref("option1")
    options := []string{"option1", "option2", "option3"}
    ctx.Expose("selectBox", bubbly.BindSelect(selected, options))
})
```

### On Directive
```go
// Event handling
Template(func(ctx RenderContext) string {
    return bubbly.On("click", func() {
        // Handle click
    }).Render("Click Me")
})

// With event modifiers
Template(func(ctx RenderContext) string {
    return bubbly.On("click", func() {
        // Handle click
    }).PreventDefault().StopPropagation().Render("Submit")
})

// Keyboard events
Template(func(ctx RenderContext) string {
    return bubbly.On("keypress", func(key string) {
        switch key {
        case "enter":
            submitForm()
        case "escape":
            cancel()
        }
    }).Render()
})
```

### Show Directive
```go
// Visibility toggle
Template(func(ctx RenderContext) string {
    visible := ctx.Get("visible").(*Ref[bool])
    
    return bubbly.Show(visible.Get(), func() string {
        return "This can be hidden"
    }).Render()
})

// With transition
Template(func(ctx RenderContext) string {
    visible := ctx.Get("visible").(*Ref[bool])
    
    return bubbly.Show(visible.Get(), func() string {
        return "Fades in/out"
    }).WithTransition().Render()
})
```

## Performance Benchmarks
```go
BenchmarkIfDirective       10000000   50  ns/op   16 B/op   1 allocs/op
BenchmarkForEach100Items   1000000    1000 ns/op  512 B/op  10 allocs/op
BenchmarkBindDirective     5000000    100 ns/op   32 B/op   2 allocs/op
BenchmarkOnDirective       10000000   80  ns/op   24 B/op   1 allocs/op
BenchmarkShowDirective     10000000   50  ns/op   16 B/op   1 allocs/op
```

## Documentation Requirements
- [ ] Package godoc with directives overview
- [ ] Each directive documented
- [ ] Directive composition guide
- [ ] 20+ runnable examples
- [ ] Best practices document
- [ ] Common patterns
- [ ] Performance guide
- [ ] Migration from manual rendering

## Success Metrics
- Templates more readable
- Less boilerplate code
- Common patterns simplified
- Developer productivity increased
- Type safety maintained
- Performance acceptable

## Directive Chaining Example
```go
// Multiple directives on template
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]Item])
    showList := ctx.Get("showList").(*Ref[bool])
    
    return bubbly.Show(showList.Get(), func() string {
        return bubbly.If(len(items.Get()) > 0,
            func() string {
                return bubbly.ForEach(items.Get(), func(item Item, i int) string {
                    return fmt.Sprintf("%d. %s\n", i+1, item.Name)
                }).Render()
            },
        ).Else(func() string {
            return "No items"
        }).Render()
    }).Render()
})
```

## Comparison: Before vs After Directives

### Before (Manual)
```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    
    var output strings.Builder
    for i, item := range items.Get() {
        output.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
    }
    return output.String()
})
```

### After (With ForEach)
```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]string])
    
    return bubbly.ForEach(items.Get(), func(item string, i int) string {
        return fmt.Sprintf("%d. %s\n", i+1, item)
    }).Render()
})
```

## Open Questions
1. Should directives support async operations?
2. How to handle directive errors gracefully?
3. Should we support custom user-defined directives?
4. Optimal strategy for ForEach diffing algorithm?
5. Should Show support CSS transitions/animations?
6. Directive caching strategy for performance?
