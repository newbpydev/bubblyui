# User Workflow: Composition API

## Primary User Journey

### Journey: Developer Creates First Composable

1. **Entry Point**: Developer wants to extract counter logic for reuse
   - System response: Composition API provides composable pattern
   - UI update: N/A (development phase)

2. **Step 1**: Create composable function
   ```go
   func UseCounter(ctx *Context, initial int) (*Ref[int], func(), func()) {
       count := ctx.Ref(initial)
       
       increment := func() {
           count.Set(count.Get() + 1)
       }
       
       decrement := func() {
           count.Set(count.Get() - 1)
       }
       
       return count, increment, decrement
   }
   ```
   - System response: Composable defined with type-safe signature
   - Developer sees: Clean,

 reusable function
   - Ready for: Usage in components

3. **Step 2**: Use composable in first component
   ```go
   NewComponent("Counter").
       Setup(func(ctx *Context) {
           count, inc, dec := UseCounter(ctx, 0)
           
           ctx.Expose("count", count)
           ctx.On("increment", func(_ interface{}) { inc() })
           ctx.On("decrement", func(_ interface{}) { dec() })
       }).
       Build()
   ```
   - System response: Composable executes, returns values
   - Component state: Counter logic encapsulated
   - Ready for: Reuse in other components

4. **Step 3**: Reuse composable in second component
   ```go
   NewComponent("ScoreBoard").
       Setup(func(ctx *Context) {
           player1Score, p1Inc, _ := UseCounter(ctx, 0)
           player2Score, p2Inc, _ := UseCounter(ctx, 0)
           
           ctx.Expose("p1", player1Score)
           ctx.Expose("p2", player2Score)
       }).
       Build()
   ```
   - System response: Same logic, multiple instances
   - Result: Clean code reuse
   - Developer sees: DRY principle achieved

5. **Completion**: Reusable logic pattern established
   - Multiple components using same composable
   - Logic tested independently
   - Easy to maintain and update

---

## Alternative Paths

### Scenario A: Using Standard Composables

1. **Developer uses UseAsync for data fetching**
   ```go
   Setup(func(ctx *Context) {
       userData := UseAsync(ctx, func() (*User, error) {
           return fetchUser()
       })
       
       ctx.OnMounted(func() {
           userData.Execute()
       })
       
       ctx.Expose("user", userData.Data)
       ctx.Expose("loading", userData.Loading)
       ctx.Expose("error", userData.Error)
   })
   ```

2. **System manages loading states automatically**
   - loading=true while fetching
   - data populated on success
   - error set on failure
   - All reactive

**Use Case:** Async operations with minimal boilerplate

### Scenario B: Provide/Inject for Dependency Injection

1. **Root component provides theme**
   ```go
   Setup(func(ctx *Context) {
       theme := ctx.Ref("dark")
       ctx.Provide("theme", theme)
       
       ctx.On("toggleTheme", func(_ interface{}) {
           current := theme.Get()
           if current == "dark" {
               theme.Set("light")
           } else {
               theme.Set("dark")
           }
       })
   })
   ```

2. **Deep child component injects theme**
   ```go
   Setup(func(ctx *Context) {
       theme := ctx.Inject("theme", ctx.Ref("light")).(*Ref[string])
       
       style := ctx.Computed(func() string {
           if theme.Get() == "dark" {
               return "background: black; color: white"
           }
           return "background: white; color: black"
       })
       
       ctx.Expose("style", style)
   })
   ```

3. **Theme changes propagate automatically**
   - Root updates theme
   - All children see new value
   - No prop drilling needed

**Use Case:** Global state without prop drilling

### Scenario C: Composable Chains

1. **Create low-level composable**
   ```go
   func UseEventListener(ctx *Context, event string, handler func()) {
       ctx.OnMounted(func() {
           // Register listener
       })
       
       ctx.OnUnmounted(func() {
           // Cleanup listener
       })
   }
   ```

2. **Create mid-level composable using low-level**
   ```go
   func UseMouse(ctx *Context) (*Ref[int], *Ref[int]) {
       x := ctx.Ref(0)
       y := ctx.Ref(0)
       
       UseEventListener(ctx, "mousemove", func() {
           // Update x, y
       })
       
       return x, y
   }
   ```

3. **Use high-level composable**
   ```go
   Setup(func(ctx *Context) {
       x, y := UseMouse(ctx)
       ctx.Expose("mouseX", x)
       ctx.Expose("mouseY", y)
   })
   ```

**Use Case:** Building abstractions on abstractions

### Scenario D: Using Dependency Interface with UseEffect (Quality of Life Enhancement)

1. **Developer creates typed refs naturally**
   ```go
   Setup(func(ctx *Context) {
       // Create typed refs - no need for Ref[any]
       count := ctx.Ref(0)        // *Ref[int]
       name := ctx.Ref("Alice")   // *Ref[string]
       
       // UseEffect accepts any Dependency (Ref or Computed)
       UseEffect(ctx, func() UseEffectCleanup {
           currentCount := count.Get().(int)
           currentName := name.Get().(string)
           fmt.Printf("%s: %d\n", currentName, currentCount)
           return nil
       }, count, name)  // Works with typed refs!
   })
   ```
   - System response: No type conversion needed
   - Developer sees: Clean, ergonomic API
   - Ready for: Production use

2. **UseEffect with Computed values**
   ```go
   Setup(func(ctx *Context) {
       firstName := ctx.Ref("John")
       lastName := ctx.Ref("Doe")
       
       // Computed implements Dependency interface
       fullName := ctx.Computed(func() string {
           return firstName.Get() + " " + lastName.Get()
       })
       
       // Watch computed values directly
       UseEffect(ctx, func() UseEffectCleanup {
           fmt.Printf("Name changed: %s\n", fullName.Get())
           return nil
       }, fullName)  // Computed as dependency!
   })
   ```
   - System response: Computed values are watchable
   - Aligns with: Vue 3 behavior
   - Developer sees: Consistent, flexible API

**Use Case:** Type-safe reactive dependencies without boilerplate

### Scenario E: Form Management with UseForm

1. **Create form with validation**
   ```go
   type SignupForm struct {
       Email    string
       Password string
       Name     string
   }
   
   Setup(func(ctx *Context) {
       form := UseForm(ctx, SignupForm{}, func(f SignupForm) map[string]string {
           errors := make(map[string]string)
           if !strings.Contains(f.Email, "@") {
               errors["email"] = "Invalid email"
           }
           if len(f.Password) < 8 {
               errors["password"] = "Too short"
           }
           return errors
       })
       
       ctx.On("submit", func(_ interface{}) {
           form.Submit()
       })
       
       ctx.Expose("form", form)
   })
   ```

2. **Form handles validation automatically**
   - Real-time validation on field change
   - Error messages reactive
   - Submit only when valid

**Use Case:** Complex forms with validation

---

## Error Handling Flows

### Error 1: Composable Called Outside Setup
- **Trigger**: Developer calls composable in template or Update
- **User sees**: 
  ```
  Error: composable UseCounter called outside Setup function
  Composables must be called within Setup or other composables
  ```
- **Recovery**: Move composable call to Setup function

**Example:**
```go
// ❌ Wrong
Template(func(ctx RenderContext) string {
    count, _, _ := UseCounter(???)  // No context!
    return fmt.Sprintf("%d", count.Get())
})

// ✅ Correct
Setup(func(ctx *Context) {
    count, _, _ := UseCounter(ctx, 0)
    ctx.Expose("count", count)
})
```

### Error 2: Circular Composable Dependencies
- **Trigger**: ComposableA calls ComposableB which calls ComposableA
- **User sees**:
  ```
  Error: circular composable dependency detected
  Call stack: UseA → UseB → UseA
  ```
- **Recovery**: Restructure composables to break cycle

**Example:**
```go
// ❌ Circular
func UseA(ctx *Context) {
    UseB(ctx)  // Calls B
}

func UseB(ctx *Context) {
    UseA(ctx)  // Calls A - CIRCULAR!
}

// ✅ Fixed
func UseShared(ctx *Context) {
    // Shared logic
}

func UseA(ctx *Context) {
    UseShared(ctx)
}

func UseB(ctx *Context) {
    UseShared(ctx)
}
```

### Error 3: Inject Without Provide
- **Trigger**: Child injects key that wasn't provided
- **User sees**: Returns default value (no error if default provided)
- **Recovery**: Provide value in parent or handle default case

**Example:**
```go
// ❌ Not provided
Setup(func(ctx *Context) {
    // Nothing provided
})

// Child
Setup(func(ctx *Context) {
    theme := ctx.Inject("theme", ctx.Ref("light"))
    // Gets default "light"
})

// ✅ Provided
Setup(func(ctx *Context) {
    ctx.Provide("theme", ctx.Ref("dark"))
})
```

### Error 4: Type Mismatch in Inject
- **Trigger**: Injected value type doesn't match expected
- **User sees**: Panic on type assertion
- **Recovery**: Fix type or use type switch

**Example:**
```go
// ❌ Type mismatch
ctx.Provide("count", 42)  // Provided int

// Consumer
count := ctx.Inject("count", ctx.Ref(0)).(*Ref[int])  // Expected Ref!
// Panic: interface conversion

// ✅ Correct types
ctx.Provide("count", ctx.Ref(42))  // Provide Ref
count := ctx.Inject("count", ctx.Ref(0)).(*Ref[int])  // Works!
```

### Error 5: Composable State Leaking
- **Trigger**: Composable uses global variable instead of Context
- **User sees**: State shared between component instances
- **Recovery**: Use Context or closure for state

**Example:**
```go
// ❌ Global state (leaks between instances)
var globalCount int

func UseBadCounter(ctx *Context) (*Ref[int], func()) {
    count := ctx.Ref(globalCount)  // BAD!
    increment := func() {
        globalCount++
        count.Set(globalCount)
    }
    return count, increment
}

// ✅ Context/closure state (isolated)
func UseGoodCounter(ctx *Context, initial int) (*Ref[int], func()) {
    count := ctx.Ref(initial)
    increment := func() {
        count.Set(count.Get() + 1)
    }
    return count, increment
}
```

---

## State Transitions

### Composable Lifecycle
```
Composable defined
    ↓
Component Setup() called
    ↓
Composable called with Context
    ↓
Composable creates Refs/Computed
    ↓
Composable registers hooks
    ↓
Composable returns values
    ↓
Component uses values
    ↓
Component lifecycle runs (mounted, updated)
    ↓
Composable hooks execute
    ↓
Component unmounts
    ↓
Composable cleanup executes
```

### Provide/Inject State Flow
```
Provider Setup
    ↓
ctx.Provide(key, value) called
    ↓
Value stored in component.provides
    ↓
Consumer Setup
    ↓
ctx.Inject(key, default) called
    ↓
Walk up component tree
    ↓
Found: return provided value
Not Found: return default
    ↓
Consumer watches value
    ↓
Provider changes value
    ↓
Consumer sees new value (reactive)
```

---

## Integration Points

### Connected to: Reactivity System (Feature 01)
- **Uses:** Ref, Computed, Watch
- **Flow:** Composables create and return reactive values
- **Data:** All state management through reactivity

### Connected to: Component System (Feature 02)
- **Uses:** Context, Setup function
- **Flow:** Composables called in Setup, values exposed to template
- **Data:** Component state populated by composables

### Connected to: Lifecycle Hooks (Feature 03)
- **Uses:** onMounted, onUpdated, onUnmounted
- **Flow:** Composables register hooks for initialization and cleanup
- **Data:** Lifecycle callbacks managed by composables

---

## Performance Considerations

### Composable Call Overhead
**Target:** < 100ns per composable call

**Optimization:**
```go
// Lightweight composable
func UseState[T any](ctx *Context, initial T) UseStateReturn[T] {
    value := ctx.Ref(initial)  // Fast ref creation
    
    return UseStateReturn[T]{  // Struct return (no heap allocation)
        Value: value,
        Set:   func(v T) { value.Set(v) },
    }
}
```

### Provide/Inject Lookup
**Target:** < 500ns for inject

**Optimization:**
- Cache inject results
- Limit tree traversal depth
- Use map for O(1) lookup

---

## Common Patterns

### Pattern 1: Shared Auth State
```go
func UseAuth(ctx *Context) UseAuthReturn {
    user := ctx.Inject("currentUser", ctx.Ref[*User](nil)).(*Ref[*User])
    
    isAuthenticated := ctx.Computed(func() bool {
        return user.Get() != nil
    })
    
    login := func(email, password string) error {
        // Login logic
        user.Set(loggedInUser)
        return nil
    }
    
    logout := func() {
        user.Set(nil)
    }
    
    return UseAuthReturn{
        User:            user,
        IsAuthenticated: isAuthenticated,
        Login:           login,
        Logout:          logout,
    }
}
```

### Pattern 2: Pagination
```go
func UsePagination(ctx *Context, itemsPerPage int) UsePaginationReturn {
    currentPage := ctx.Ref(1)
    totalItems := ctx.Ref(0)
    
    totalPages := ctx.Computed(func() int {
        return (totalItems.Get() + itemsPerPage - 1) / itemsPerPage
    })
    
    nextPage := func() {
        if currentPage.Get() < totalPages.Get() {
            currentPage.Set(currentPage.Get() + 1)
        }
    }
    
    prevPage := func() {
        if currentPage.Get() > 1 {
            currentPage.Set(currentPage.Get() - 1)
        }
    }
    
    return UsePaginationReturn{
        CurrentPage: currentPage,
        TotalPages:  totalPages,
        NextPage:    nextPage,
        PrevPage:    prevPage,
        SetTotal:    func(total int) { totalItems.Set(total) },
    }
}
```

### Pattern 3: Undo/Redo
```go
func UseHistory[T any](ctx *Context, initial T) UseHistoryReturn[T] {
    history := ctx.Ref([]T{initial})
    index := ctx.Ref(0)
    
    current := ctx.Computed(func() T {
        return history.Get()[index.Get()]
    })
    
    canUndo := ctx.Computed(func() bool {
        return index.Get() > 0
    })
    
    canRedo := ctx.Computed(func() bool {
        return index.Get() < len(history.Get())-1
    })
    
    push := func(value T) {
        h := history.Get()[:index.Get()+1]
        h = append(h, value)
        history.Set(h)
        index.Set(len(h) - 1)
    }
    
    undo := func() {
        if canUndo.Get() {
            index.Set(index.Get() - 1)
        }
    }
    
    redo := func() {
        if canRedo.Get() {
            index.Set(index.Get() + 1)
        }
    }
    
    return UseHistoryReturn[T]{
        Current: current,
        CanUndo: canUndo,
        CanRedo: canRedo,
        Push:    push,
        Undo:    undo,
        Redo:    redo,
    }
}
```

---

## Testing Workflow

### Unit Test: Composable in Isolation
```go
func TestUseCounter(t *testing.T) {
    // Arrange
    ctx := createTestContext()
    
    // Act
    count, increment, decrement := UseCounter(ctx, 0)
    
    // Assert
    assert.Equal(t, 0, count.Get())
    
    increment()
    assert.Equal(t, 1, count.Get())
    
    decrement()
    assert.Equal(t, 0, count.Get())
}
```

### Integration Test: Composable in Component
```go
func TestComponentWithComposable(t *testing.T) {
    // Arrange
    component := NewComponent("Test").
        Setup(func(ctx *Context) {
            count, inc, _ := UseCounter(ctx, 0)
            ctx.Expose("count", count)
            ctx.On("click", func(_ interface{}) {
                inc()
            })
        }).
        Build()
    
    // Act
    component.Init()
    component.Update(ClickMsg{})
    
    // Assert
    count := component.Get("count").(*Ref[int])
    assert.Equal(t, 1, count.Get())
}
```

### Test: Provide/Inject
```go
func TestProvideInject(t *testing.T) {
    // Arrange
    parent := NewComponent("Parent").
        Setup(func(ctx *Context) {
            ctx.Provide("theme", ctx.Ref("dark"))
        }).
        Children(child).
        Build()
    
    child := NewComponent("Child").
        Setup(func(ctx *Context) {
            theme := ctx.Inject("theme", ctx.Ref("light"))
            ctx.Expose("theme", theme)
        }).
        Build()
    
    // Act
    parent.Init()
    child.Init()
    
    // Assert
    theme := child.Get("theme").(*Ref[string])
    assert.Equal(t, "dark", theme.Get())
}
```

---

## Documentation for Users

### Quick Start
1. Create composable function with `Use*` prefix
2. Accept `Context` as first parameter
3. Create Refs/Computed using context
4. Return stable references and functions
5. Call in component Setup

### Best Practices
- Composables return objects with named fields
- Use generics for type safety
- Register cleanup in lifecycle hooks
- Test composables independently
- Document composable contract
- Use provide/inject for cross-tree data
- Avoid global state in composables

### Troubleshooting
- **Composable not working?** Check it's called in Setup
- **State shared between instances?** Use Context not globals
- **Inject returns default?** Check parent provides value
- **Type panic?** Verify type assertion matches provided type
- **Cleanup not running?** Register with lifecycle hooks
