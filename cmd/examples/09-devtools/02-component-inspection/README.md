# Example 02: Component Inspection

**Multi-level component hierarchy for dev tools exploration**

## What This Demonstrates

This example showcases dev tools' component inspection capabilities with a real-world todo list:

1. **Multi-Level Hierarchy** - 4 levels of component nesting
2. **Component Tree Navigation** - Explore parent-child relationships
3. **State Across Components** - Share state via props
4. **Computed Values** - Derived state automatically updates
5. **Dynamic Children** - TodoItems created programmatically

## Architecture

### Directory Structure
```
02-component-inspection/
├── main.go                    # Entry point
├── app.go                     # Root component (TodoApp)
├── components/                # UI components
│   ├── header.go              # Header with title
│   ├── todo_item.go           # Individual todo (leaf)
│   ├── todo_list.go           # List of todos (parent)
│   └── footer.go              # Statistics footer
└── README.md                  # This file
```

### Component Hierarchy (4 levels)
```
TodoApp (root)
├── Header
├── TodoList
│   ├── TodoItem #1 "Learn BubblyUI"
│   ├── TodoItem #2 "Enable Dev Tools"
│   ├── TodoItem #3 "Inspect Component Tree"
│   ├── TodoItem #4 "View State in DevTools"
│   └── TodoItem #5 "Navigate with Keyboard"
└── Footer
```

### State Flow
```
TodoApp
├── todos: Ref[[]Todo]           ← Root state
├── selectedIndex: Ref[int]      ← Selection state
│
Flows to children:
├── TodoList receives todos + selectedIndex
│   └── TodoItem receives completed Ref for each item
└── Footer receives todos
    ├── Computes totalCount
    ├── Computes completedCount
    └── Computes pendingCount
```

## Key Features

### 1. Multi-Level Component Tree

Open dev tools (F12) and see the complete hierarchy:

```
📁 TodoApp
├── 📄 Header
├── 📁 TodoList
│   ├── 📄 TodoItem#1
│   ├── 📄 TodoItem#2
│   ├── 📄 TodoItem#3
│   ├── 📄 TodoItem#4
│   └── 📄 TodoItem#5
└── 📄 Footer
```

Navigate with `↑`/`↓`, expand with `→`, collapse with `←`.

### 2. State Inspection at Each Level

**TodoApp (root):**
```
State:
• todos: [5 items] (Ref)
• selectedIndex: 0 (Ref)
```

**TodoList:**
```
State:
• todos: [5 items] (Ref from parent)
• selectedIndex: 0 (Ref from parent)
```

**TodoItem#3:**
```
State:
• id: 3
• text: "Inspect Component Tree"
• completed: false (Ref)
• isSelected: false
```

**Footer:**
```
State:
• todos: [5 items] (Ref from parent)

Computed Values:
• totalCount: 5
• completedCount: 2
• pendingCount: 3
```

### 3. Reactive State Updates

Watch state update in real-time:

1. Press `Space` to toggle todo completion
2. See `completed` Ref change in TodoItem
3. See `completedCount` update in Footer
4. All happens automatically via reactivity!

### 4. Dynamic Child Creation

```go
// TodoList creates children dynamically
for i, todo := range todos {
    item, _ := CreateTodoItem(TodoItemProps{
        ID:        todo.ID,
        Text:      todo.Text,
        Completed: todo.Completed,
    })
    ctx.ExposeComponent(todo.Text, item)
}
```

Each TodoItem is created programmatically and registered in the component tree.

## Run the Example

```bash
cd 02-component-inspection
go run main.go
```

## Using Dev Tools

### Explore Component Tree

1. **Press F12** to open dev tools
2. **Navigate** with `↑`/`↓` keys
3. **Expand TodoList** with `→` key
4. **See all TodoItems** as children
5. **Select any component** to view its state

### State Inspection Deep Dive

**TodoApp root state:**
```
Press Enter on "TodoApp" → See State tab
• todos array with 5 items
• selectedIndex currently at 0
```

**Individual TodoItem state:**
```
Press Enter on "TodoItem#3" → See State tab
• id: 3
• text: "Inspect Component Tree"
• completed: false (this is a Ref!)
• isSelected: false
```

**Footer computed values:**
```
Press Enter on "Footer" → See State tab
• totalCount: 5 (computed)
• completedCount: 2 (computed)
• pendingCount: 3 (computed)
```

### Watch Reactive Updates

1. Select TodoApp in tree
2. Toggle a todo (press Space)
3. Watch the state change:
   - `todos` array updates
   - Footer's computed values recalculate
   - UI re-renders automatically

## Code Highlights

### Component Factory Pattern

Every component uses the factory pattern:

```go
func CreateTodoItem(props TodoItemProps) (bubbly.Component, error) {
    builder := bubbly.NewComponent(fmt.Sprintf("TodoItem#%d", props.ID))
    
    builder = builder.Setup(func(ctx *bubbly.Context) {
        ctx.Expose("id", props.ID)
        ctx.Expose("completed", props.Completed)
        // ...
    })
    
    builder = builder.Template(func(ctx bubbly.RenderContext) string {
        // Render using BubblyUI components
    })
    
    return builder.Build()
}
```

### Props Down, State Up

**Parent passes state down:**
```go
todoList, _ := components.CreateTodoList(components.TodoListProps{
    Todos:         todos,          // Ref passed down
    SelectedIndex: selectedIndex,  // Ref passed down
})
```

**Child exposes it for inspection:**
```go
ctx.Expose("todos", props.Todos)
ctx.Expose("selectedIndex", props.SelectedIndex)
```

### Computed Values (Footer)

```go
totalCount := ctx.Computed(func() interface{} {
    todos := props.Todos.Get().([]Todo)
    return len(todos)
})

completedCount := ctx.Computed(func() interface{} {
    todos := props.Todos.Get().([]Todo)
    count := 0
    for _, todo := range todos {
        if todo.Completed.Get().(bool) {
            count++
        }
    }
    return count
})

ctx.Expose("totalCount", totalCount)
ctx.Expose("completedCount", completedCount)
```

These automatically update when `todos` changes!

### Using BubblyUI Components

We use framework components throughout:
- `components.Text()` for styled text (Header)
- `components.Checkbox()` for checkboxes (TodoItem)
- `components.Badge()` for statistics (Footer)

No manual Lipgloss styling for components!

## Exploration Tasks

Try these to learn dev tools:

### Task 1: Navigate the Tree
1. Open dev tools (F12)
2. Navigate to TodoList
3. Expand it with `→`
4. Count the TodoItem children (should be 5)
5. Select TodoItem#3
6. View its state (press Enter)

### Task 2: Watch State Changes
1. Select TodoApp in tree
2. View State tab
3. Note `selectedIndex: 0`
4. Press `↓` in main app
5. Watch `selectedIndex` change to `1`

### Task 3: Observe Computed Values
1. Select Footer in tree
2. View State tab
3. Note `completedCount: 2`
4. Toggle a todo (Space)
5. Watch `completedCount` update to `3`

### Task 4: Explore Refs
1. Select TodoItem#1
2. View `completed` Ref (should be `true`)
3. Toggle it with Space
4. Watch the Ref update in real-time

## What Makes This Example Special

### Real Component Hierarchy
Not a flat list - actual parent-child relationships like real apps.

### Shared State Pattern
Demonstrates how state flows down through props (common pattern).

### Reactive Computed Values
Footer's statistics auto-update when todos change (no manual updates).

### Dynamic Component Creation
TodoItems created in a loop, each with unique ID and state.

### BubblyUI Components Throughout
Shows proper use of Card, Text, Checkbox, Badge components.

## Next Steps

After mastering this example:

1. **Examine each component's code** - See the patterns
2. **Modify the initial todos** - Add more items
3. **Add a new computed value** - Try `percentComplete`
4. **Read the architecture guide** - [Composable Apps](../../../../docs/architecture/composable-apps.md)
5. **Try example 03** - State debugging with history

## Related Documentation

- [Composable Apps Architecture](../../../../docs/architecture/composable-apps.md)
- [Component Inspector Guide](../../../../docs/devtools/features.md#component-inspector)
- [State Viewer Guide](../../../../docs/devtools/features.md#state-viewer)
- [Component Reference](../../../../docs/components/README.md)

## Troubleshooting

**Can't see TodoItems in tree?**
- Make sure to expand TodoList with `→` key
- Verify component creation succeeded (check for errors)

**State not updating?**
- Check that Refs are used (not plain values)
- Verify `ExposeComponent` was called
- Make sure component is recreated after state change

**Computed values not updating?**
- Ensure they depend on the changing Ref
- Check the computed function accesses the Ref with `.Get()`

---

**Next:** [Example 03 - State Debugging](../03-state-debugging/) →
