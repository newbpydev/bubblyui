# Router Todo Example

This example demonstrates a complete Todo application using the BubblyUI router system with **zero Bubbletea boilerplate**.

## Features

- **Zero-boilerplate** with `bubbly.Run()` - no manual `tea.NewProgram`
- **Router with multiple pages** - home, add, detail
- **Route parameters** - `/todo/:id` for viewing individual todos
- **Shared state via composables** - `UseTodos()` singleton pattern
- **Mode-based input handling** - navigation vs input mode
- **Built-in components** - Card, Badge for consistent UI
- **Conditional key bindings** - different keys active in different modes
- **WithMessageHandler** - for text input capture

## Structure

```
todo/
├── main.go           # Entry point with bubbly.Run()
├── app.go            # Root component with router setup
├── composables/
│   └── use_todos.go  # Shared todo state management
└── pages/
    ├── home.go       # Home page (todo list)
    ├── add.go        # Add todo page
    └── detail.go     # Todo detail page
```

## Running

```bash
cd cmd/examples/07-router/todo
go run .
```

## Keyboard Shortcuts

### Navigation Mode (default)

| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| Space | Toggle todo completion |
| Enter | View todo detail |
| a | Add new todo |
| d | Delete selected todo |
| b/Esc | Go back |
| q | Quit |

### Input Mode (on Add page)

| Key | Action |
|-----|--------|
| Tab | Next field |
| Shift+Tab | Previous field |
| p | Cycle priority (low → medium → high) |
| Enter | Submit form |
| Esc | Cancel and go back |

## Key Patterns Demonstrated

### 1. Zero-Boilerplate Entry Point

```go
func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())  // That's it!
}
```

### 2. Router Setup with ProvideRouter

```go
r, _ := router.NewRouterBuilder().
    RouteWithOptions("/", router.WithComponent(homePage)).
    RouteWithOptions("/add", router.WithComponent(addPage)).
    RouteWithOptions("/todo/:id", router.WithComponent(detailPage)).
    Build()

// In Setup:
router.ProvideRouter(ctx, r)
```

### 3. Accessing Router in Child Components

```go
// In any child component's Setup:
r := router.UseRouter(ctx)
r.Push(&router.NavigationTarget{Path: "/add"})
```

### 4. Reactive Route Access

```go
route := router.UseRoute(ctx)
// route is *bubbly.Ref[*router.Route] - automatically updates on navigation
```

### 5. Shared State Composable

```go
// Singleton pattern for shared state
var UseTodos = composables.CreateShared(func() *TodosReturn {
    // ... initialize state
})

// Use in any component:
todoManager := composables.UseTodos()
todoManager.AddTodo("New task", "Description", "high")
```

### 6. Mode-Based Key Bindings

```go
.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key: "a", Event: "add", Description: "Add todo",
    Condition: func() bool { return mode.GetTyped() == ModeNavigation },
})
```

### 7. Text Input with MessageHandler

```go
.WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
    if mode.GetTyped() != ModeInput {
        return nil
    }
    // Handle key runes for text input
})
```

## Routes

| Path | Page | Description |
|------|------|-------------|
| `/` | Home | Todo list with selection |
| `/add` | Add | Form to create new todo |
| `/todo/:id` | Detail | View/edit individual todo |

## Components Used

- `components.Card` - For content containers
- `components.Badge` - For status indicators (stats, route, mode)
- `router.View` - For rendering current route's component

## Composables Used

- `router.UseRouter()` - Access router for navigation
- `router.UseRoute()` - Reactive access to current route
- `UseTodos()` - Shared todo state management (custom)
