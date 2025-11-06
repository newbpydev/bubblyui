# Todo App Versions - Detailed Comparison

## 🎯 Overview

We have **three production-ready implementations** of the same todo application, each showcasing different architectural approaches with BubblyUI:

| Version | Architecture | Lines | Components | Composables | Best For |
|---------|--------------|-------|------------|-------------|----------|
| **02-todo** | Monolithic | 583 | 1 | Yes (UseForm) | Production apps |
| **02-todo-bubbletea** | Pure Bubbletea | 451 | 1 | No | Learning Bubbletea |
| **02-todo-components** | Component-based | ~740 | 4 | Yes (UseForm) | Modular apps |

---

## 📊 Detailed Feature Comparison

### Component Structure

#### 02-todo (Monolithic)
```
TodoApp (single component)
├── Setup() - All state and logic
├── Template() - All rendering
└── Key bindings - Declarative with conditions
```

**Pros:**
- ✅ Simple mental model
- ✅ All code in one place
- ✅ Easy to understand flow
- ✅ Auto-initialized by wrapper

**Cons:**
- ❌ Can grow large (583 lines)
- ❌ Hard to reuse parts
- ❌ All state in one scope

---

#### 02-todo-bubbletea (Pure Bubbletea)
```
model struct
├── Update() - Handle all messages
├── View() - Render everything
└── Manual state management
```

**Pros:**
- ✅ No framework overhead
- ✅ Direct Bubbletea patterns
- ✅ Full control
- ✅ Smallest codebase (451 lines)

**Cons:**
- ❌ No declarative patterns
- ❌ No reactive state
- ❌ Manual input mode tracking
- ❌ No automatic help text

---

#### 02-todo-components (Component-based)
```
TodoApp (parent)
├── TodoForm (child component)
│   ├── Uses UseForm composable
│   ├── Props: InputMode, FocusedField
│   └── Events: OnSubmit, OnCancel
├── TodoList (child component)
│   ├── Props: Todos, SelectedIndex, InputMode
│   └── Events: OnToggle, OnSelect
└── TodoStats (child component)
    ├── Props: Todos
    └── Computed: Total, Completed, Pending
```

**Pros:**
- ✅ Highly reusable components
- ✅ Clear separation of concerns
- ✅ Easier testing (isolated components)
- ✅ Scales to large apps

**Cons:**
- ❌ More boilerplate
- ❌ Manual child initialization required
- ❌ More files to manage
- ❌ Learning curve for patterns

---

## 🔧 Composables Usage

### 02-todo (Monolithic)
```go
// Uses UseForm composable for form state
form := composables.UseForm(ctx, TodoFormData{
    Title:       "",
    Description: "",
    Priority:    "medium",
}, validateFn)

// Direct access to form state
form.SetField("Title", "New Task")
form.Submit()
if form.IsValid.GetTyped() {
    // Handle submission
}
```

**Composables Used:**
- ✅ `UseForm` - Form state management with validation

---

### 02-todo-bubbletea (Pure Bubbletea)
```go
// Manual state management - no composables
type model struct {
    title       string
    description string
    priority    string
    // ... manual validation
}
```

**Composables Used:**
- ❌ None - Pure Bubbletea approach

---

### 02-todo-components (Component-based)
```go
// TodoForm component uses UseForm composable
form := composables.UseForm(ctx, TodoFormData{...}, validateFn)

// TodoList and TodoStats use computed values
totalCount := ctx.Computed(func() interface{} {
    return len(todos.Get().([]Todo))
})

// Could add more composables:
// - UseList for todo list management
// - UseSelection for selection state
// - UseFilter for filtering todos
```

**Composables Used:**
- ✅ `UseForm` - In TodoForm component
- ✅ `ctx.Computed` - In TodoStats for derived state
- 🔶 Could use more (UseList, UseSelection, etc.)

**Opportunity for Enhancement:**
We could create custom composables like:
- `UseTodoList(ctx, initialTodos)` - Encapsulate todo CRUD operations
- `UseSelection(ctx, items)` - Reusable selection logic
- `UseFilter(ctx, items, filterFn)` - Filter todos by status

---

## 🔄 Component Initialization

### The `.Init()` Requirement

**Question:** With automatic-bridge, should we initialize components manually?

**Answer:** It depends on the component hierarchy:

#### Parent Component (Auto-initialized)
```go
// Parent component wrapped with RunAuto()
func main() {
    comp := createTodoApp()
    bubbly.RunAuto(comp) // ✅ Auto-initialized by wrapper
}
```
The automatic bridge wrapper calls `.Init()` on the root component automatically.

#### Child Components (Manual initialization required)
```go
// In parent's Setup()
todoForm, _ := components.CreateTodoForm(props)
todoList, _ := components.CreateTodoList(props)
todoStats, _ := components.CreateTodoStats(props)

// ⚠️ REQUIRED: Initialize children manually
todoForm.Init()
todoList.Init()
todoStats.Init()
```

**Why Manual Init is Needed:**
1. Child components are created **dynamically** in parent's Setup()
2. Framework doesn't know about them (not part of component tree)
3. Their Setup() won't run until Init() is called
4. Without Init(), computed values and state don't exist → panic!

**Framework Enhancement Opportunity:**
We could add auto-initialization when exposing components:
```go
// Future API idea
ctx.ExposeComponent("todoForm", todoForm) // Auto-calls Init()
```

---

## 📝 Code Organization

### 02-todo (Monolithic)
```
02-todo/
└── main.go (583 lines)
    ├── TodoFormData struct
    ├── Todo struct
    ├── createTodoApp() function
    │   ├── Key bindings (declarative)
    │   ├── Setup() - All logic
    │   └── Template() - All rendering
    └── main() - RunAuto()
```

**Organization:**
- Single file
- Sections marked with comments
- All state in one Setup()

---

### 02-todo-bubbletea (Pure Bubbletea)
```
02-todo-bubbletea/
└── main.go (451 lines)
    ├── todo struct
    ├── model struct
    ├── Init() - Initialize
    ├── Update() - Handle messages
    ├── View() - Render
    └── main() - tea.NewProgram()
```

**Organization:**
- Single file
- Standard Bubbletea pattern
- No framework abstractions

---

### 02-todo-components (Component-based)
```
02-todo-components/
├── main.go (343 lines)
│   ├── createTodoApp() - Parent component
│   ├── Setup() - Compose children
│   ├── Template() - Layout
│   └── main() - RunAuto()
├── components/
│   ├── todo_form.go (234 lines)
│   │   ├── TodoFormData struct
│   │   ├── TodoFormProps struct
│   │   └── CreateTodoForm() - Component builder
│   ├── todo_list.go (120 lines)
│   │   ├── Todo struct
│   │   ├── TodoListProps struct
│   │   └── CreateTodoList() - Component builder
│   └── todo_stats.go (68 lines)
│       ├── TodoStatsProps struct
│       └── CreateTodoStats() - Component builder
├── README.md - Architecture guide
├── STATUS.md - Implementation status
└── COMPARISON.md - This file
```

**Organization:**
- Multiple files by feature
- Each component is self-contained
- Props define dependencies
- Clear component boundaries

---

## 🎨 State Management

### 02-todo (Monolithic)
```go
// All state in one Setup()
todos := ctx.Ref([]Todo{})
selectedIndex := ctx.Ref(0)
inputMode := ctx.Ref(false)
editMode := ctx.Ref(false)
form := composables.UseForm(ctx, TodoFormData{...}, validateFn)
```

**Pattern:** Flat state in single scope

---

### 02-todo-bubbletea (Pure Bubbletea)
```go
// Manual state in model struct
type model struct {
    todos         []todo
    selectedIndex int
    inputMode     bool
    editMode      bool
    title         string
    description   string
    // ... more fields
}
```

**Pattern:** Struct fields, manual updates

---

### 02-todo-components (Component-based)
```go
// Parent state
todos := ctx.Ref([]Todo{})
selectedIndex := ctx.Ref(0)
inputMode := ctx.Ref(false)

// Child state (TodoForm)
form := composables.UseForm(ctx, TodoFormData{...}, validateFn)
focusedField := ctx.Ref("Title")

// Child state (TodoStats) - Computed
totalCount := ctx.Computed(func() interface{} {
    return len(props.Todos.Get().([]Todo))
})
```

**Pattern:** Distributed state with props passing

---

## 🔌 Props and Events

### Component Communication

Only **02-todo-components** uses props and events:

```go
// Props (parent → child)
todoForm := components.CreateTodoForm(components.TodoFormProps{
    InputMode:    inputMode,      // Pass reactive state
    FocusedField: focusedField,   // Pass reactive state
    OnSubmit: func(data TodoFormData) {
        // Handle submission in parent
    },
    OnCancel: func() {
        // Handle cancel in parent
    },
})

// Events (child → parent)
// Child emits:
ctx.On("submitForm", func(_ interface{}) {
    form.Submit()
    if form.IsValid.GetTyped() {
        props.OnSubmit(form.Values.Get().(TodoFormData))
    }
})

// Parent handles via OnSubmit callback
```

**Benefits:**
- ✅ Clear data flow (props down, events up)
- ✅ Loose coupling between components
- ✅ Reusable components (different parents)

---

## 🔑 Key Bindings

### Declarative vs Imperative

#### 02-todo (Declarative)
```go
WithKeyBinding("n", "newTodo", "New todo").
WithCondition(func() bool {
    return !inputModeRef.Get().(bool)
})
```

#### 02-todo-bubbletea (Imperative)
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if !m.inputMode {
            switch msg.String() {
            case "n":
                // Handle new todo
            }
        }
    }
}
```

#### 02-todo-components (Declarative)
```go
WithKeyBinding("n", "newTodo", "New todo").
WithCondition(func() bool {
    return !inputModeRef.Get().(bool)
})
```

**Winner:** Declarative (02-todo and 02-todo-components)
- Auto-generates help text
- Cleaner conditional logic
- Less boilerplate

---

## 🧪 Testing Considerations

### Unit Testing

#### 02-todo (Monolithic)
```go
// Test entire app
func TestTodoApp(t *testing.T) {
    comp := createTodoApp()
    comp.Init()
    // Test all functionality
}
```

**Pros:** Simple setup  
**Cons:** Large test scope, hard to isolate

---

#### 02-todo-components (Component-based)
```go
// Test components in isolation
func TestTodoForm(t *testing.T) {
    form, _ := CreateTodoForm(TodoFormProps{...})
    form.Init()
    // Test just the form
}

func TestTodoList(t *testing.T) {
    list, _ := CreateTodoList(TodoListProps{...})
    list.Init()
    // Test just the list
}
```

**Pros:** Isolated, focused tests  
**Cons:** Need to mock props/events

---

## 📈 Scalability

### Adding New Features

**Example:** Add "Due Date" field to todos

#### 02-todo (Monolithic)
1. Add `DueDate` to `TodoFormData` and `Todo` structs
2. Update form validation in `Setup()`
3. Update form rendering in `Template()`
4. Update edit/add logic in event handlers

**Impact:** 4-5 locations in one file

---

#### 02-todo-components (Component-based)
1. Add `DueDate` to `TodoFormData` and `Todo` structs
2. Update `todo_form.go` validation
3. Update `todo_form.go` template
4. Update `todo_list.go` template (display)

**Impact:** 2 files (form and list components)

**Benefit:** Changes are localized, other components unaffected

---

## 🎓 Learning Curve

### Beginner → Advanced

```
02-todo-bubbletea (Easiest)
├── Learn pure Bubbletea patterns
├── Understand Model/Update/View
└── No framework abstractions

↓

02-todo (Intermediate)
├── Learn BubblyUI declarative patterns
├── Understand reactive state (Ref, Computed)
├── Learn composables (UseForm)
└── Understand automatic bridge

↓

02-todo-components (Advanced)
├── All of the above, plus:
├── Component composition
├── Props and events pattern
├── Child initialization lifecycle
└── Component reusability patterns
```

---

## 🚀 Performance

### Bundle Size

| Version | Lines | Components | Files |
|---------|-------|------------|-------|
| 02-todo-bubbletea | 451 | 1 | 1 |
| 02-todo | 583 | 1 | 1 |
| 02-todo-components | ~740 | 4 | 5 |

### Runtime Performance

All three versions have **similar runtime performance**:
- Same reactive system (Refs, Computed)
- Same rendering (Lipgloss)
- Same event system

Component overhead is **negligible** - just function calls.

---

## 💡 When to Use Each Version

### Use **02-todo-bubbletea** when:
- ✅ Learning Bubbletea fundamentals
- ✅ Building simple CLI tools
- ✅ Want minimal dependencies
- ✅ Need maximum control
- ✅ Don't need reactivity

### Use **02-todo** when:
- ✅ Building production TUI apps
- ✅ Want declarative patterns
- ✅ Need reactive state management
- ✅ Want composables (UseForm)
- ✅ Prefer simpler architecture
- ✅ **RECOMMENDED for most apps**

### Use **02-todo-components** when:
- ✅ Building large, complex TUIs
- ✅ Need reusable components
- ✅ Multiple developers/teams
- ✅ Want Vue-like patterns
- ✅ Planning to scale the app
- ✅ Component library development

---

## 🔮 Future Enhancements

### Potential Improvements

#### Auto-Initialize Child Components
```go
// Current (manual)
todoForm.Init()
todoList.Init()

// Future (automatic)
ctx.ExposeComponent("todoForm", todoForm) // Auto-inits
```

#### More Composables
```go
// UseTodoList composable
todoList := composables.UseTodoList(ctx, []Todo{})
todoList.Add(todo)
todoList.Remove(index)
todoList.Toggle(index)

// UseSelection composable
selection := composables.UseSelection(ctx, items)
selection.Next()
selection.Previous()
selection.Selected() // Returns current item
```

#### Component Registry
```go
// Register reusable components
bubbly.RegisterComponent("TodoForm", CreateTodoForm)
bubbly.RegisterComponent("TodoList", CreateTodoList)

// Use anywhere
form := ctx.Component("TodoForm", props)
```

---

## 📊 Final Recommendation

### For Most Applications: **02-todo** (Monolithic)

**Why:**
1. ✅ Best balance of simplicity and power
2. ✅ Declarative patterns without complexity
3. ✅ Composables (UseForm) for complex state
4. ✅ Auto-initialized by wrapper
5. ✅ Easy to understand and maintain
6. ✅ Scales well up to ~1000 lines

### For Large Applications: **02-todo-components**

**When app grows beyond:**
- Multiple screens/views
- Reusable UI components
- Team collaboration
- Component library needs

### For Learning: **02-todo-bubbletea**

**Best for:**
- Understanding Bubbletea fundamentals
- Learning Model/Update/View pattern
- No framework abstractions

---

## 🎯 Summary

All three versions are **production-ready** and demonstrate different architectural approaches:

- **02-todo-bubbletea**: Pure Bubbletea, maximum control
- **02-todo**: BubblyUI declarative patterns, best balance ⭐
- **02-todo-components**: Component-based, maximum modularity

Choose based on your app's complexity and team's preferences! 🚀
