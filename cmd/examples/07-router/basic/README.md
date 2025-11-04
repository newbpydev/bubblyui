# Basic Router Example

This example demonstrates basic navigation between routes using the BubblyUI Router system.

## Features Demonstrated

- ✅ Declarative route configuration
- ✅ Navigation between multiple screens
- ✅ Dynamic route parameters (`:id`)
- ✅ History management (back/forward)
- ✅ Current route display with Badge component
- ✅ Component rendering with Card component
- ✅ Keyboard-driven navigation

## Routes

| Path | Name | Description |
|------|------|-------------|
| `/` | home | Home screen with welcome message |
| `/about` | about | About screen with framework info |
| `/contact` | contact | Contact information screen |
| `/user/:id` | user-detail | User profile with dynamic ID parameter |

## Keyboard Controls

### Navigation
- **1** - Navigate to Home
- **2** - Navigate to About
- **3** - Navigate to Contact
- **4** - Navigate to User 123
- **5** - Navigate to User 456

### History
- **b** - Go back in history
- **f** - Go forward in history

### General
- **q** or **Ctrl+C** - Quit the application

## Running the Example

```bash
go run cmd/examples/07-router/basic/main.go
```

## Code Highlights

### Router Configuration

```go
r, err := router.NewRouterBuilder().
    RouteWithOptions("/",
        router.WithName("home"),
        router.WithComponent(createHomeComponent()),
    ).
    RouteWithOptions("/user/:id",
        router.WithName("user-detail"),
        router.WithComponent(createUserDetailComponent()),
    ).
    Build()
```

### Navigation

```go
// Navigate by path
return m, m.router.Push(&router.NavigationTarget{Path: "/"})

// Navigate with parameters
return m, m.router.Push(&router.NavigationTarget{Path: "/user/123"})

// History navigation
return m, m.router.Back()
return m, m.router.Forward()
```

### Component Rendering

```go
// Use RouterView to render the current route's component
routerView := router.NewRouterView(m.router, 0)
content := routerView.View()
```

## Components Used

- **Card** - Content containers for each screen
- **Badge** - Current route indicator
- **RouterView** - Renders the matched route's component

## What to Try

1. Navigate between different screens using number keys
2. Use the back button (b) to return to previous screens
3. Navigate forward (f) after going back
4. Switch between different users (keys 4 and 5)
5. Observe how the current route badge updates

## Next Steps

- See `../guards/` for authentication and navigation guards
- See `../nested/` for nested routes and layouts
