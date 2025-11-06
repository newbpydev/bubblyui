# Todo App - Component-Based Version (Vue-like)

**Status**: 🚧 Work in Progress - Demonstrating Vue-like component composition patterns

## Overview

This version demonstrates how to build a todo application using **component composition** - breaking the app into reusable, self-contained components similar to Vue.js single-file components.

## Architecture

### Component Structure

```
02-todo-components/
├── main.go                    # Main app composition
└── components/
    ├── todo_form.go          # Form component with UseForm composable
    ├── todo_list.go          # List component with selection
    ├── todo_stats.go         # Statistics component with computed values
    └── (helpers)             # Shared rendering functions
```

### Vue-like Patterns Demonstrated

#### 1. **Component Props** (like Vue props)
```go
type TodoFormProps struct {
    InputMode    *bubbly.Ref[bool]
    FocusedField *bubbly.Ref[string]
    OnSubmit     func(data TodoFormData)
    OnCancel     func()
}
```

#### 2. **Composables** (like Vue Composition API)
```go
// Using UseForm composable for form state management
form := composables.UseForm(ctx, TodoFormData{
    Title:       "",
    Description: "",
    Priority:    "medium",
}, func(data TodoFormData) map[string]string {
    // Validation logic
    errors := make(map[string]string)
    if len(data.Title) < 3 {
        errors["Title"] = "Must be at least 3 characters"
    }
    return errors
})
```

#### 3. **Computed Values** (like Vue computed)
```go
// Computed: Total count
totalCount := ctx.Computed(func() interface{} {
    todos := props.Todos.Get().([]Todo)
    return len(todos)
})
```

#### 4. **Event Emission** (like Vue $emit)
```go
// Child component emits event
ctx.On("submitForm", func(_ interface{}) {
    form.Submit()
    if form.IsValid.GetTyped() {
        data := form.Values.Get().(TodoFormData)
        if props.OnSubmit != nil {
            props.OnSubmit(data)  // Emit to parent
        }
    }
})
```

#### 5. **Component Composition** (like Vue template composition)
```go
// Parent composes child components
todoForm, err := components.CreateTodoForm(components.TodoFormProps{
    InputMode:    inputMode,
    FocusedField: focusedField,
    OnSubmit: func(data components.TodoFormData) {
        // Handle submission
    },
})

todoList, err := components.CreateTodoList(components.TodoListProps{
    Todos:         todos,
    SelectedIndex: selectedIndex,
    OnToggle: func(index int) {
        // Handle toggle
    },
})
```

## Component Breakdown

### 1. TodoForm Component (`todo_form.go`)

**Purpose**: Reusable form for creating/editing todos

**Features**:
- Uses `UseForm` composable for state management
- Automatic validation
- Field-level focus tracking
- Character input handling
- Form submission/cancellation

**Props**:
- `InputMode`: Controls whether form is active
- `FocusedField`: Tracks which field has focus
- `OnSubmit`: Callback when form is submitted
- `OnCancel`: Callback when form is cancelled

**Events**:
- `setFormData`: Load existing todo for editing
- `clearForm`: Reset form to initial state
- `addChar`: Add character to focused field
- `removeChar`: Remove character from focused field
- `submitForm`: Validate and submit
- `cancelForm`: Cancel and reset

### 2. TodoList Component (`todo_list.go`)

**Purpose**: Display and interact with todo list

**Features**:
- Render todos with priority indicators
- Selection tracking
- Completion toggle
- Dynamic styling based on mode

**Props**:
- `Todos`: List of todos
- `SelectedIndex`: Currently selected todo
- `InputMode`: Whether in input mode
- `OnToggle`: Callback when todo is toggled
- `OnSelect`: Callback when selection changes

**Events**:
- `toggleTodo`: Toggle completion status
- `selectPrevious`: Move selection up
- `selectNext`: Move selection down

### 3. TodoStats Component (`todo_stats.go`)

**Purpose**: Display computed statistics

**Features**:
- Computed total count
- Computed completed count
- Computed pending count
- Automatic updates when todos change

**Props**:
- `Todos`: List of todos (for computing stats)

**Computed Values**:
- `totalCount`: Total number of todos
- `completedCount`: Number of completed todos
- `pendingCount`: Number of pending todos

## Benefits of Component-Based Architecture

### 1. **Reusability**
Each component can be reused in different contexts:
```go
// Use TodoForm for creating
createForm := components.CreateTodoForm(...)

// Use TodoForm for editing
editForm := components.CreateTodoForm(...)
```

### 2. **Separation of Concerns**
Each component handles its own:
- State management
- Event handling
- Rendering logic
- Validation

### 3. **Testability**
Components can be tested in isolation:
```go
func TestTodoForm_Validation(t *testing.T) {
    form := components.CreateTodoForm(...)
    // Test form validation
}
```

### 4. **Maintainability**
Changes to one component don't affect others:
- Update TodoForm without touching TodoList
- Change TodoStats without affecting form logic
- Add new components without refactoring existing ones

### 5. **Composability**
Build complex UIs from simple components:
```go
// Compose multiple components
app := ComposeApp(
    TodoStats(...),
    TodoForm(...),
    TodoList(...),
)
```

## Comparison with Monolithic Version

| Aspect | Monolithic (`02-todo`) | Component-Based (`02-todo-components`) |
|--------|------------------------|----------------------------------------|
| **File Structure** | Single `main.go` (583 lines) | Multiple files (4 components) |
| **Reusability** | Low (everything coupled) | High (independent components) |
| **Testability** | Hard (test entire app) | Easy (test each component) |
| **Maintainability** | Medium (one large file) | High (focused files) |
| **Learning Curve** | Lower (everything visible) | Higher (understand composition) |
| **Scalability** | Poor (grows linearly) | Good (grows modularly) |

## Vue.js Comparison

### Vue Single-File Component
```vue
<template>
  <form @submit="handleSubmit">
    <input v-model="title" />
    <button type="submit">Submit</button>
  </form>
</template>

<script setup>
import { ref, computed } from 'vue'

const title = ref('')
const isValid = computed(() => title.value.length >= 3)

function handleSubmit() {
  if (isValid.value) {
    emit('submit', { title: title.value })
  }
}
</script>
```

### BubblyUI Component
```go
func CreateTodoForm(props TodoFormProps) (bubbly.Component, error) {
    return bubbly.NewComponent("TodoForm").
        Setup(func(ctx *bubbly.Context) {
            form := composables.UseForm(ctx, TodoFormData{
                Title: "",
            }, func(data TodoFormData) map[string]string {
                errors := make(map[string]string)
                if len(data.Title) < 3 {
                    errors["Title"] = "Too short"
                }
                return errors
            })
            
            ctx.On("submitForm", func(_ interface{}) {
                form.Submit()
                if form.IsValid.GetTyped() {
                    props.OnSubmit(form.Values.Get().(TodoFormData))
                }
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            // Render form
        }).
        Build()
}
```

**Similarities**:
- Props for component input
- Composables for reusable logic
- Computed values for derived state
- Event emission for parent communication
- Template for rendering

## When to Use Component-Based Architecture

### ✅ Use When:
- Building complex applications with many features
- Need to reuse UI patterns across different parts
- Working in a team (easier to divide work)
- Planning for long-term maintenance
- Want to test components in isolation

### ❌ Don't Use When:
- Building simple, one-off applications
- Prototyping quickly
- Learning BubblyUI basics
- Performance is absolutely critical (minimal overhead, but exists)

## Future Enhancements

### 1. **Component Library**
Build a library of reusable components:
- `FormField` - Generic form field wrapper
- `Button` - Reusable button component
- `Modal` - Reusable modal dialog
- `Dropdown` - Reusable dropdown menu

### 2. **Provide/Inject Pattern**
Share state across component tree:
```go
// Parent provides theme
ctx.Provide("theme", darkTheme)

// Child injects theme
theme := ctx.Inject("theme", defaultTheme)
```

### 3. **Slots Pattern**
Allow parent to customize child rendering:
```go
// Parent passes custom content
form := CreateTodoForm(TodoFormProps{
    HeaderSlot: func() string {
        return "Custom Header"
    },
})
```

### 4. **Component Registry**
Register components globally:
```go
registry.Register("TodoForm", CreateTodoForm)
registry.Register("TodoList", CreateTodoList)

// Use by name
form := registry.Create("TodoForm", props)
```

## Related Examples

- **Monolithic Version**: [../02-todo/](../02-todo/) - Single-file approach
- **Pure Bubbletea**: [../02-todo-bubbletea/](../02-todo-bubbletea/) - No framework
- **Comparison**: [../COMPARISON.md](../COMPARISON.md) - Detailed analysis

---

**Key Takeaway**: Component-based architecture trades initial complexity for long-term maintainability. It's the Vue.js way of building TUI applications! 🎯
