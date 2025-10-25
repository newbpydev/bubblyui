# User Workflow: Directives

## Primary User Journey

### Journey: Developer Uses Directives for Cleaner Templates

1. **Entry Point**: Developer has complex template with manual loops and conditionals
   - System response: Directives provide declarative alternatives
   - UI update: N/A (development phase)

2. **Step 1**: Replace manual loop with ForEach
   ```go
   // Before: Manual loop
   Template(func(ctx RenderContext) string {
       items := ctx.Get("items").(*Ref[[]string])
       var output strings.Builder
       for i, item := range items.Get() {
           output.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
       }
       return output.String()
   })
   
   // After: ForEach directive
   Template(func(ctx RenderContext) string {
       items := ctx.Get("items").(*Ref[[]string])
       return ForEach(items.Get(), func(item string, i int) string {
           return fmt.Sprintf("%d. %s\n", i+1, item)
       }).Render()
   })
   ```
   - System response: Cleaner, more readable code
   - Developer sees: Intent clearer
   - Ready for: More directives

3. **Step 2**: Add conditional rendering with If
   ```go
   Template(func(ctx RenderContext) string {
       items := ctx.Get("items").(*Ref[[]string])
       
       return If(len(items.Get()) > 0,
           func() string {
               return ForEach(items.Get(), renderItem).Render()
           },
       ).Else(func() string {
           return "No items to display"
       }).Render()
   })
   ```
   - System response: Conditional logic declarative
   - Result: Self-documenting code
   - Ready for: Event handling

4. **Step 3**: Add event handling with On
   ```go
   Template(func(ctx RenderContext) string {
       return On("click", func() {
           // Handle click
       }).Render("Click Me")
   })
   ```
   - System response: Events attached declaratively
   - Template: Clean and expressive
   - Ready for: Complex interactions

5. **Completion**: Template transformed
   - Before: 30 lines of imperative code
   - After: 10 lines of declarative directives
   - Maintainability: Significantly improved
   - Type safety: Maintained throughout

---

## Alternative Paths

### Scenario A: Form with Bind Directive

1. **Developer creates form with two-way binding**
   ```go
   Setup(func(ctx *Context) {
       name := ctx.Ref("")
       email := ctx.Ref("")
       agreed := ctx.Ref(false)
       
       ctx.Expose("name", name)
       ctx.Expose("email", email)
       ctx.Expose("agreed", agreed)
   })
   
   Template(func(ctx RenderContext) string {
       name := ctx.Get("name").(*Ref[string])
       email := ctx.Get("email").(*Ref[string])
       agreed := ctx.Get("agreed").(*Ref[bool])
       
       return fmt.Sprintf(
           "Name: %s\nEmail: %s\nAgreed: %v\n",
           Bind(name).Render(),
           Bind(email).Render(),
           BindCheckbox(agreed).Render(),
       )
   })
   ```

2. **User types in form**
   - Input changes trigger Bind handlers
   - Refs update automatically
   - Template re-renders with new values
   - Two-way sync working

**Use Case:** Forms with minimal boilerplate

### Scenario B: Nested Lists with ForEach

1. **Developer renders categories with items**
   ```go
   Template(func(ctx RenderContext) string {
       categories := ctx.Get("categories").(*Ref[[]Category])
       
       return ForEach(categories.Get(), func(cat Category, i int) string {
           header := fmt.Sprintf("%s:\n", cat.Name)
           
           items := ForEach(cat.Items, func(item string, j int) string {
               return fmt.Sprintf("  - %s\n", item)
           }).Render()
           
           return header + items
       }).Render()
   })
   ```

2. **Nested structure renders correctly**
   - Outer loop renders categories
   - Inner loop renders items
   - Indentation preserved
   - Clean hierarchy

**Use Case:** Complex nested data structures

### Scenario C: Conditional States with If/ElseIf/Else

1. **Developer shows different UI based on status**
   ```go
   Template(func(ctx RenderContext) string {
       status := ctx.Get("status").(*Ref[string])
       
       return If(status.Get() == "loading",
           func() string { return "Loading..." },
       ).ElseIf(status.Get() == "error",
           func() string { return "Error occurred!" },
       ).ElseIf(status.Get() == "empty",
           func() string { return "No data" },
       ).Else(func() string {
           return "Data loaded successfully"
       }).Render()
   })
   ```

2. **Status changes, UI updates accordingly**
   - loading → "Loading..."
   - error → "Error occurred!"
   - empty → "No data"
   - success → "Data loaded successfully"

**Use Case:** State-based UI rendering

### Scenario D: Show/Hide with Show Directive

1. **Developer toggles element visibility**
   ```go
   Setup(func(ctx *Context) {
       visible := ctx.Ref(true)
       
       ctx.On("toggle", func(_ interface{}) {
           visible.Set(!visible.Get())
       })
       
       ctx.Expose("visible", visible)
   })
   
   Template(func(ctx RenderContext) string {
       visible := ctx.Get("visible").(*Ref[bool])
       
       return Show(visible.Get(), func() string {
           return "This content can be hidden"
       }).Render()
   })
   ```

2. **Toggle triggered**
   - visible changes from true to false
   - Content hidden (stays in DOM)
   - Fast toggle performance

**Use Case:** Collapsible sections, dropdowns

---

## Error Handling Flows

### Error 1: Bind on Non-Input Element
- **Trigger**: Bind used on element that doesn't accept input
- **User sees**:
  ```
  Error: Bind directive requires input element
  Cannot use Bind on: div, span, etc.
  ```
- **Recovery**: Use Bind only on input-capable elements

**Example:**
```go
// ❌ Wrong
Bind(value).Render() // On a div

// ✅ Correct
// Bind only works in Setup with proper input handling
Setup(func(ctx *Context) {
    ctx.On("input", func(val interface{}) {
        ref.Set(val.(string))
    })
})
```

### Error 2: ForEach with Nil Slice
- **Trigger**: ForEach receives nil instead of empty slice
- **User sees**: Returns empty string, no error
- **Recovery**: Check for nil before ForEach or use defensive coding

**Example:**
```go
// ❌ Can be nil
items := ctx.Get("items").(*Ref[[]string])
ForEach(items.Get(), renderItem).Render() // nil panic

// ✅ Safe
items := ctx.Get("items").(*Ref[[]string])
if items.Get() != nil {
    return ForEach(items.Get(), renderItem).Render()
}
return "No items"
```

### Error 3: Render Function Panics
- **Trigger**: Directive render function panics
- **User sees**: Error logged, empty string returned
- **Recovery**: Fix render function, add error handling

**Example:**
```go
// ❌ Panics
ForEach(items.Get(), func(item string, i int) string {
    var ptr *string
    return *ptr // nil pointer dereference!
}).Render()

// ✅ Safe
ForEach(items.Get(), func(item string, i int) string {
    if item == "" {
        return "Empty"
    }
    return item
}).Render()
```

### Error 4: Circular If Nesting
- **Trigger**: If conditions reference each other
- **User sees**: Stack overflow or timeout
- **Recovery**: Restructure conditions

**Example:**
```go
// ❌ Circular
If(a.Get(), func() string {
    return If(b.Get(), func() string {
        return If(a.Get(), ...) // Circular!
    }).Render()
}).Render()

// ✅ Proper structure
If(a.Get() && b.Get(), func() string {
    return "Both true"
}).Render()
```

### Error 5: Event Handler Not Registered
- **Trigger**: On directive used but handler not in Setup
- **User sees**: Events don't trigger
- **Recovery**: Register handler in Setup

**Example:**
```go
// ❌ No handler registered
Template(func(ctx RenderContext) string {
    return On("click", func() {
        // This won't work if not registered in Setup!
    }).Render("Click")
})

// ✅ Proper registration
Setup(func(ctx *Context) {
    ctx.On("click", func(_ interface{}) {
        handleClick()
    })
})
```

---

## State Transitions

### If Directive States
```
Condition evaluated
    ↓
If true:  Then branch executes
If false: Check ElseIf conditions
    ↓
All ElseIf false: Execute Else
    ↓
No Else: Return empty string
```

### ForEach Directive States
```
Collection received
    ↓
If empty: Return empty string
If non-empty:
    ├─> For each item: Call render function
    ├─> Collect outputs
    └─> Join and return
```

### Bind Directive States
```
Initial render
    ↓
Input displayed with current Ref value
    ↓
User types
    ↓
Input change event
    ↓
Bind handler updates Ref
    ↓
Ref notifies watchers
    ↓
Component re-renders
    ↓
Input updated with new value
```

---

## Integration Points

### Connected to: Component System (Feature 02)
- **Uses:** Template function, RenderContext
- **Flow:** Directives execute within template render
- **Data:** Component state accessed via context

### Connected to: Reactivity System (Feature 01)
- **Uses:** Ref values, Computed values
- **Flow:** Directives react to state changes
- **Data:** Reactive values drive directive behavior

### Connected to: Lifecycle Hooks (Feature 03)
- **Uses:** onMounted for event registration
- **Flow:** Event handlers cleanup on unmount
- **Data:** Handlers registered during lifecycle

### Connected to: Composition API (Feature 04)
- **Uses:** Composables can return directive configs
- **Flow:** Composables provide data for directives
- **Data:** Shared state accessed by directives

---

## Performance Considerations

### ForEach Performance
**Target:** < 1ms for 100 items

**Optimization:**
- Pre-allocate output slices
- Use string builders
- Cache unchanged items
- Diff algorithm for updates

### If Performance
**Target:** < 50ns per evaluation

**Optimization:**
- Short-circuit evaluation
- Pool directive objects
- Minimize allocations

---

## Common Patterns

### Pattern 1: Filtered List
```go
Template(func(ctx RenderContext) string {
    items := ctx.Get("items").(*Ref[[]Item])
    filter := ctx.Get("filter").(*Ref[string])
    
    filtered := []Item{}
    for _, item := range items.Get() {
        if strings.Contains(item.Name, filter.Get()) {
            filtered = append(filtered, item)
        }
    }
    
    return ForEach(filtered, func(item Item, i int) string {
        return fmt.Sprintf("%d. %s\n", i+1, item.Name)
    }).Render()
})
```

### Pattern 2: Table with Directives
```go
Template(func(ctx RenderContext) string {
    rows := ctx.Get("rows").(*Ref[[]Row])
    
    header := "ID | Name | Status\n"
    divider := "---|------|-------\n"
    
    body := ForEach(rows.Get(), func(row Row, i int) string {
        return fmt.Sprintf("%2d | %-4s | %s\n", 
            row.ID, row.Name, row.Status)
    }).Render()
    
    return header + divider + body
})
```

### Pattern 3: Dynamic Menu
```go
Template(func(ctx RenderContext) string {
    menuItems := ctx.Get("menuItems").(*Ref[[]MenuItem])
    selected := ctx.Get("selected").(*Ref[int])
    
    return ForEach(menuItems.Get(), func(item MenuItem, i int) string {
        isSelected := i == selected.Get()
        
        return If(isSelected,
            func() string {
                return fmt.Sprintf("> %s\n", item.Label)
            },
        ).Else(func() string {
            return fmt.Sprintf("  %s\n", item.Label)
        }).Render()
    }).Render()
})
```

---

## Testing Workflow

### Unit Test: If Directive
```go
func TestIfDirective(t *testing.T) {
    // Arrange
    thenCalled := false
    elseCalled := false
    
    // Act - condition true
    result := If(true, func() string {
        thenCalled = true
        return "then"
    }).Else(func() string {
        elseCalled = true
        return "else"
    }).Render()
    
    // Assert
    assert.True(t, thenCalled)
    assert.False(t, elseCalled)
    assert.Equal(t, "then", result)
}
```

### Unit Test: ForEach Directive
```go
func TestForEachDirective(t *testing.T) {
    // Arrange
    items := []string{"A", "B", "C"}
    
    // Act
    result := ForEach(items, func(item string, i int) string {
        return fmt.Sprintf("%d:%s,", i, item)
    }).Render()
    
    // Assert
    assert.Equal(t, "0:A,1:B,2:C,", result)
}
```

### Integration Test: Directives in Template
```go
func TestDirectivesInTemplate(t *testing.T) {
    // Arrange
    component := NewComponent("Test").
        Setup(func(ctx *Context) {
            items := ctx.Ref([]string{"X", "Y"})
            ctx.Expose("items", items)
        }).
        Template(func(ctx RenderContext) string {
            items := ctx.Get("items").(*Ref[[]string])
            return ForEach(items.Get(), func(item string, i int) string {
                return fmt.Sprintf("%s\n", item)
            }).Render()
        }).
        Build()
    
    // Act
    component.Init()
    view := component.View()
    
    // Assert
    assert.Contains(t, view, "X")
    assert.Contains(t, view, "Y")
}
```

---

## Documentation for Users

### Quick Start
1. Import directives: `import "github.com/yourusername/bubblyui/pkg/bubbly"`
2. Use in template: `If(condition, thenFunc).Render()`
3. Chain directives: `.ElseIf().Else()`
4. Compose directives: Nest for complex logic

### Best Practices
- Use If for conditional rendering
- Use Show for visibility toggle
- Use ForEach for lists
- Keep render functions pure
- Test directives independently
- Use type-safe bindings
- Compose for complex logic

### Troubleshooting
- **Directive not rendering?** Check condition/collection
- **ForEach empty?** Verify slice not nil
- **Bind not working?** Ensure handler registered
- **Events not firing?** Check event name
- **Performance slow?** Profile render functions
