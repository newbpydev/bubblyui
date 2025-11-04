# Nested Routes Example

This example demonstrates nested routing with parent-child route relationships using the BubblyUI Router system.

## Features Demonstrated

- ✅ Parent-child route relationships
- ✅ Nested component rendering
- ✅ Breadcrumb navigation from matched routes
- ✅ Layout composition (parent layout + child views)
- ✅ Route hierarchy visualization

## Routes

| Path | Name | Parent | Description |
|------|------|--------|-------------|
| `/` | home | - | Home screen |
| `/dashboard` | dashboard | - | Dashboard parent layout |
| `/dashboard/stats` | dashboard-stats | dashboard | Statistics child view |
| `/dashboard/settings` | dashboard-settings | dashboard | Settings child view |
| `/dashboard/profile` | dashboard-profile | dashboard | Profile child view |

## Keyboard Controls

- **1** - Navigate to Home
- **2** - Navigate to Dashboard/Stats
- **3** - Navigate to Dashboard/Settings
- **4** - Navigate to Dashboard/Profile
- **b** - Go back in history
- **f** - Go forward in history
- **q** or **Ctrl+C** - Quit

## Running the Example

```bash
go run cmd/examples/07-router/nested/main.go
```

## Nested Routing Concept

In nested routing, parent routes provide a layout/container while child routes render within that layout:

```
/dashboard (parent)
├── /dashboard/stats (child)
├── /dashboard/settings (child)
└── /dashboard/profile (child)
```

The Dashboard layout stays constant while child views change within it.

## Code Highlights

### Nested Route Configuration

```go
RouteWithOptions("/dashboard",
    router.WithName("dashboard"),
    router.WithComponent(createDashboardLayout()),
    router.WithChildren(
        &router.RouteRecord{
            Path:      "stats",
            Name:      "dashboard-stats",
            Component: createStatsComponent(),
        },
        &router.RouteRecord{
            Path:      "settings",
            Name:      "dashboard-settings",
            Component: createSettingsComponent(),
        },
    ),
)
```

### Breadcrumb from Matched Routes

```go
// Show breadcrumbs from matched routes
breadcrumbs := ""
for i, matched := range route.Matched {
    if i > 0 {
        breadcrumbs += " > "
    }
    breadcrumbs += matched.Name
}
```

### RouterView at Different Depths

```go
// Root level (depth 0) - renders parent
rootView := router.NewRouterView(router, 0)

// Child level (depth 1) - renders child within parent
childView := router.NewRouterView(router, 1)
```

## What to Try

1. **Navigate to Dashboard children** - Use keys 2-4 to switch between child views
2. **Observe breadcrumbs** - See how the matched routes build the breadcrumb trail
3. **Use history navigation** - Back/forward works across nested routes
4. **Notice layout persistence** - Dashboard layout concept stays constant

## Components Used

- **Card** - Content containers for child views
- **Badge** - Breadcrumb display

## Next Steps

- See `../basic/` for basic navigation patterns
- See `../guards/` for authentication and guards
- Combine nested routes with guards for protected sections

## Note

This example shows the nested routing concept. In a full implementation, the Dashboard layout component would use `RouterView` at depth 1 to dynamically render child components within the layout.
