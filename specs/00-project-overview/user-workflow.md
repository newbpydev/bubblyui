# User Workflow: BubblyUI Complete Journey

## Developer Personas

### Persona 1: Vue Developer (Sara)
- **Background**: 5 years Vue.js experience, learning Go
- **Goal**: Build CLI tool with familiar patterns
- **Pain Points**: Unfamiliar with TUI development, Go syntax learning curve
- **Success Criteria**: Build functional app in < 1 day

### Persona 2: Go Developer (Mike)
- **Background**: 3 years Go, no frontend experience
- **Goal**: Add interactive UI to existing CLI tool
- **Pain Points**: UI/UX patterns unfamiliar, want type safety
- **Success Criteria**: Maintain Go idioms, compile-time safety

### Persona 3: DevOps Engineer (Alex)
- **Background**: Mixed stack, builds internal tools
- **Goal**: Create monitoring dashboard quickly
- **Pain Points**: Limited time, needs pre-built components
- **Success Criteria**: Production-ready in < 1 week

## Primary User Journey: First Application

### Entry Point: New Project Setup

**Workflow: Getting Started (00-project-setup)**

#### Step 1: Installation
**User Action**: Install BubblyUI
```bash
go get github.com/newbpydev/bubblyui/pkg/bubbly
```

**System Response**:
- Downloads package and dependencies
- Go modules resolves versions

**UI Feedback**:
- Terminal shows download progress
- Success message on completion

#### Step 2: Create First Component
**User Action**: Write hello world component
```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    component, _ := bubbly.NewComponent("HelloWorld").
        Template(func(ctx bubbly.RenderContext) string {
            return "Hello, BubblyUI! 👋"
        }).
        Build()
    
    m := model{component: component}
    tea.NewProgram(m).Run()
}

type model struct {
    component bubbly.Component
}

func (m model) Init() tea.Cmd { return m.component.Init() }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if _, ok := msg.(tea.KeyMsg); ok {
        return m, tea.Quit
    }
    updated, cmd := m.component.Update(msg)
    m.component = updated.(bubbly.Component)
    return m, cmd
}
func (m model) View() string { return m.component.View() }
```

**System Response**:
- Code compiles successfully
- Component renders

**UI Feedback**:
- "Hello, BubblyUI! 👋" appears in terminal
- Press any key to exit

**Journey Branch**: 
- ✅ Success → Step 3 (Add Reactivity)
- ❌ Error → Troubleshooting guide

---

### Feature Journey: Adding Reactivity (01-reactivity-system)

#### Step 3: Add Interactive State
**User Action**: Add counter with reactive state
```go
bubbly.NewComponent("Counter").
    Setup(func(ctx *bubbly.Context) {
        count := ctx.Ref(0)
        ctx.Expose("count", count)
        
        ctx.On("increment", func(_ interface{}) {
            current := count.Get().(int)
            count.Set(current + 1)
        })
    }).
    Template(func(ctx bubbly.RenderContext) string {
        count := ctx.Get("count").(*bubbly.Ref[interface{}])
        return fmt.Sprintf("Count: %d", count.Get().(int))
    }).
    Build()
```

**System Response**:
- Ref[T] created with type safety
- State management initialized

**UI Feedback**:
- Counter displays current value
- Updates on increment event

**Data Flow**:
```
User presses 'i' → tea.KeyMsg 
→ model.Update() 
→ component.Emit("increment") 
→ Event handler 
→ count.Set(+1) 
→ State updated 
→ View() re-renders 
→ UI shows new count
```

---

### Feature Journey: Component Composition (02-component-model)

#### Step 4: Extract Reusable Components
**User Action**: Create button component and use it
```go
// Create button component
button := bubbly.NewComponent("Button").
    Props(ButtonProps{Label: "Click Me"}).
    Template(func(ctx bubbly.RenderContext) string {
        props := ctx.Props().(ButtonProps)
        return fmt.Sprintf("[%s]", props.Label)
    }).
    Build()

// Use in parent component
parent := bubbly.NewComponent("App").
    Setup(func(ctx *bubbly.Context) {
        clicks := ctx.Ref(0)
        
        ctx.On("button-click", func(_ interface{}) {
            current := clicks.Get().(int)
            clicks.Set(current + 1)
        })
    }).
    Children(button).
    Build()
```

**System Response**:
- Component tree established
- Props validated and passed
- Event bubbling configured

**UI Feedback**:
- Button renders with label
- Click increments parent counter
- Parent and child update together

**Integration Points**:
- Feature 01 (Reactivity): State management
- Feature 02 (Components): Composition pattern

---

### Feature Journey: Lifecycle Management (03-lifecycle-hooks)

#### Step 5: Add Data Fetching
**User Action**: Fetch data on mount, cleanup on unmount
```go
bubbly.NewComponent("DataDisplay").
    Setup(func(ctx *bubbly.Context) {
        data := ctx.Ref[*Data](nil)
        loading := ctx.Ref(true)
        
        // Fetch on mount
        ctx.OnMounted(func() {
            // Trigger data fetch command
            ctx.Emit("fetch-data", nil)
        })
        
        // Watch for data changes
        ctx.OnUpdated(func() {
            if data.Get() != nil {
                loading.Set(false)
            }
        }, data)
        
        // Cleanup
        ctx.OnUnmounted(func() {
            // Cancel pending requests
            // Close connections
        })
    }).
    Build()
```

**System Response**:
- onMounted executes after first render
- Data fetch initiated
- Watchers registered and auto-cleanup

**UI Feedback**:
- Loading spinner while fetching
- Data displays when loaded
- Clean shutdown on quit

**Integration Points**:
- Feature 01 (Reactivity): Watchers
- Feature 02 (Components): Lifecycle integration
- Feature 03 (Lifecycle): Hook system

---

### Feature Journey: Shared Logic (04-composition-api)

#### Step 6: Extract Composable
**User Action**: Create reusable logic
```go
// Define composable
func useCounter(ctx *bubbly.Context) (*bubbly.Ref[int], func(), func()) {
    count := ctx.Ref(0)
    
    increment := func() {
        count.Set(count.Get().(int) + 1)
    }
    
    decrement := func() {
        count.Set(count.Get().(int) - 1)
    }
    
    return count, increment, decrement
}

// Use in components
bubbly.NewComponent("CounterA").
    Setup(func(ctx *bubbly.Context) {
        count, inc, dec := useCounter(ctx)
        
        ctx.On("up", func(_ interface{}) { inc() })
        ctx.On("down", func(_ interface{}) { dec() })
        ctx.Expose("count", count)
    }).
    Build()
```

**System Response**:
- Composable logic reused across components
- Type safety maintained
- Lifecycle tied to component

**UI Feedback**:
- Multiple components share same logic
- Independent state per instance

**Integration Points**:
- Feature 01 (Reactivity): Shared reactive state
- Feature 02 (Components): Multiple instances
- Feature 03 (Lifecycle): Cleanup automation

---

### Feature Journey: Declarative Templates (05-directives)

#### Step 7: Use Directives
**User Action**: Simplify template logic
```go
bubbly.NewComponent("TodoList").
    Setup(func(ctx *bubbly.Context) {
        todos := ctx.Ref([]Todo{})
        filter := ctx.Ref("all")
    }).
    Template(func(ctx bubbly.RenderContext) string {
        todos := ctx.Get("todos").(*bubbly.Ref[interface{}])
        filter := ctx.Get("filter").(*bubbly.Ref[interface{}])
        
        // Using directives
        return directives.ForEach(todos.Get().([]Todo), func(todo Todo) string {
            return directives.If(shouldShow(todo, filter.Get().(string)), func() string {
                return fmt.Sprintf("- [%s] %s", checkbox(todo), todo.Text)
            })
        })
    }).
    Build()
```

**System Response**:
- Conditional rendering handled
- List iteration simplified
- Clean template code

**UI Feedback**:
- Dynamic list rendering
- Filtered items shown/hidden
- Reactive to filter changes

**Integration Points**:
- Feature 01 (Reactivity): Reactive lists
- Feature 02 (Components): Template rendering
- Feature 05 (Directives): Declarative helpers

---

### Feature Journey: Production UI (06-built-in-components)

#### Step 8: Use Built-in Components
**User Action**: Build form with pre-built components
```go
form := components.NewForm().
    AddField(components.NewInput().
        Label("Name").
        Validate(required).
        Build()).
    AddField(components.NewSelect().
        Label("Role").
        Options([]string{"Admin", "User"}).
        Build()).
    OnSubmit(func(data FormData) {
        // Handle submission
    }).
    Build()

app := components.NewAppLayout().
    Header(header).
    Content(form).
    Footer(footer).
    Build()
```

**System Response**:
- Components initialized with validation
- Layout calculated
- Event handlers registered

**UI Feedback**:
- Professional-looking form
- Validation feedback
- Keyboard navigation works

**Integration Points**:
- All features: Built-in components use everything

---

## Complete Application Journey

### Journey: Building a Todo Application

#### Phase 1: Setup (Feature 00, 01)
1. Initialize Go project
2. Install BubblyUI
3. Create basic component structure
4. Add reactive state (todos list)

**Time**: 15 minutes

#### Phase 2: Core Features (Feature 02, 03)
1. Create TodoItem component
2. Create TodoList component
3. Add lifecycle hooks for persistence
4. Implement event handling

**Time**: 30 minutes

#### Phase 3: Enhanced UX (Feature 04, 05)
1. Extract composables (useLocalStorage, useTodoFilter)
2. Add directives for filtering
3. Implement keyboard shortcuts

**Time**: 20 minutes

#### Phase 4: Polish (Feature 06)
1. Replace custom input with built-in Input
2. Add built-in Button components
3. Use built-in Form for new todos
4. Apply AppLayout

**Time**: 15 minutes

**Total Time**: 80 minutes (< 1.5 hours)

---

## Alternative Workflows

### Workflow A: Migrating from Raw Bubbletea

#### Entry: Existing Bubbletea Application

1. **Add BubblyUI Dependency**
   - Install package
   - No changes to existing code yet

2. **Wrap One Model in Component**
   - Convert simplest model first
   - Keep other models as-is
   - Test integration

3. **Migrate State to Reactive**
   - Replace struct fields with Ref[T]
   - Add watchers for side effects
   - Remove manual update logic

4. **Add Lifecycle Hooks**
   - Extract init logic to onMounted
   - Add cleanup in onUnmounted
   - Use onUpdated for derived state

5. **Extract Composables**
   - Identify repeated logic
   - Create composables
   - Share across components

6. **Complete Migration**
   - All models → components
   - Remove old boilerplate
   - Use built-in components

**Time**: 1-2 hours per model (depending on complexity)

---

### Workflow B: Building Dashboard Application

#### Entry: Need Monitoring Dashboard

1. **Plan Layout** (Feature 06 - Templates)
   ```
   AppLayout
   ├── Header (status bar)
   ├── Sidebar (navigation)
   └── Content (data panels)
   ```

2. **Create Data Panel** (Feature 02, 03)
   - Fetch data on mount
   - Auto-refresh with lifecycle
   - Display in table

3. **Add Interactivity** (Feature 01, 05)
   - Reactive filtering
   - Sortable columns
   - Search functionality

4. **Extract Shared Logic** (Feature 04)
   - useDataFetch composable
   - useRefresh composable
   - useFilter composable

5. **Polish UI** (Feature 06)
   - Replace custom components with built-ins
   - Consistent styling
   - Keyboard shortcuts

**Time**: 2-3 hours for basic dashboard

---

## Error Recovery Workflows

### Error Flow 1: Component Build Failed

**Trigger**: Missing required template
```go
component, err := bubbly.NewComponent("Test").
    Build() // ❌ Missing Template()
```

**User Sees**:
```
Error: validation failed for component "Test"
  - Missing template function
  
Fix: Add .Template(func(ctx RenderContext) string { ... })
```

**Recovery**:
1. Read error message
2. Add Template() to builder chain
3. Rebuild
4. Success! ✅

---

### Error Flow 2: Type Assertion Failed

**Trigger**: Wrong type in template
```go
.Template(func(ctx RenderContext) string {
    count := ctx.Get("count").(string) // ❌ It's an int!
    return count
})
```

**User Sees**:
```
panic: interface conversion: interface {} is int, not string
```

**Recovery**:
1. Check exposed type
2. Fix type assertion
3. Add test to catch this
4. Success! ✅

---

### Error Flow 3: Race Condition

**Trigger**: Concurrent access to Ref
```go
// Two goroutines accessing ref
go func() { count.Set(1) }()
go func() { count.Set(2) }()
```

**User Sees** (with `-race`):
```
WARNING: DATA RACE
Read at 0x... by goroutine 7
Write at 0x... by goroutine 8
```

**Recovery**:
1. Don't use goroutines directly
2. Use tea.Cmd for async
3. Run tests with `-race`
4. Success! ✅

---

## State Transition Diagrams

### Application Lifecycle
```
Created (Build)
    ↓
Initialized (Init)
    ↓
Mounted (onMounted)
    ↓
Running (Update loop)
    ├─ State Changes → onUpdated
    ├─ User Input → Events
    └─ Re-renders → View
    ↓
Unmounting (onBeforeUnmount)
    ↓
Unmounted (onUnmounted)
    ↓
Cleaned Up (cleanup functions)
```

### Component State Machine
```
                    ┌─────────┐
                    │ Created │
                    └────┬────┘
                         ↓
                    ┌─────────┐
                    │  Built  │
                    └────┬────┘
                         ↓
          ┌──────────────┴──────────────┐
          ↓                              ↓
    ┌──────────┐                  ┌──────────┐
    │ Mounted  │──────Events─────→│ Updating │
    └────┬─────┘                  └────┬─────┘
         │                              │
         └──────────Unmount─────────────┘
                    ↓
              ┌──────────┐
              │ Cleaned  │
              └──────────┘
```

---

## Integration Points Map

### Feature Cross-Reference
```
00-project-setup
    → Enables: ALL features (build infrastructure)

01-reactivity-system
    → Used by: 02, 03, 04, 05, 06
    → Provides: Ref, Computed, Watch
    → Unlocks: Reactive components

02-component-model
    → Uses: 01 (state management)
    → Used by: 03, 04, 05, 06
    → Provides: Component abstraction
    → Unlocks: Composition, lifecycle

03-lifecycle-hooks
    → Uses: 01 (watchers), 02 (components)
    → Used by: 04, 05, 06
    → Provides: Lifecycle management
    → Unlocks: Side effects, cleanup

04-composition-api
    → Uses: 01, 02, 03
    → Used by: 05, 06
    → Provides: Logic reuse
    → Unlocks: Composables, DI

05-directives
    → Uses: 01 (reactivity), 02 (components)
    → Used by: 06
    → Provides: Template helpers
    → Unlocks: Declarative templates

06-built-in-components
    → Uses: ALL features
    → Provides: Component library
    → Unlocks: Rapid development
```

---

## User Success Paths

### Path 1: Quick Win (< 30 min)
```
Install → Hello World → Add Counter → Success! 🎉
Features used: 00, 01, 02
```

### Path 2: Learning Journey (< 2 hours)
```
Hello World → Counter → Todo List → Data Fetch → Success! 🎉
Features used: 00, 01, 02, 03
```

### Path 3: Production App (< 1 week)
```
Planning → Layout → Features → Composables → Polish → Deploy → Success! 🎉
Features used: ALL
```

---

## Summary

BubblyUI provides a clear, progressive journey from simple hello world to production applications. Each feature builds on previous ones, creating a cohesive development experience. The framework guides users from basic concepts (reactive state) through advanced patterns (composables, directives) to production-ready applications (built-in components). Multiple entry points accommodate different skill levels and use cases, while consistent patterns across features ensure a smooth learning curve.

**Key Success Factors**:
- ✅ Familiar patterns for Vue developers
- ✅ Type safety for Go developers  
- ✅ Progressive complexity (start simple, grow as needed)
- ✅ Clear error messages and recovery paths
- ✅ Production-ready from day one
