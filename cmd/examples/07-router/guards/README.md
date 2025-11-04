# Navigation Guards Example

This example demonstrates authentication guards and protected routes using the BubblyUI Router system.

## Features Demonstrated

- ✅ Navigation guards (BeforeEach)
- ✅ Protected routes with authentication
- ✅ Login flow with redirect
- ✅ Auth state management
- ✅ Route metadata for auth requirements
- ✅ Mode-based input handling for login form
- ✅ Visual auth status with Badge component

## Routes

| Path | Name | Protected | Description |
|------|------|-----------|-------------|
| `/` | home | No | Public home screen |
| `/login` | login | No | Login form |
| `/dashboard` | dashboard | Yes | Protected dashboard (requires auth) |
| `/profile` | profile | Yes | Protected profile (requires auth) |

## Keyboard Controls

### Navigation Mode (Default)
- **1** - Navigate to Home
- **2** - Navigate to Dashboard (protected)
- **3** - Navigate to Profile (protected)
- **4** - Navigate to Login
- **5** - Logout
- **b** - Go back
- **f** - Go forward
- **Enter** - Activate login form (when on login page)
- **q** or **Ctrl+C** - Quit

### Input Mode (Login Form)
- **Type** - Enter username/password
- **Tab** - Switch between fields
- **Enter** - Submit login
- **ESC** - Cancel and return to navigation mode
- **Backspace** - Delete character

## Running the Example

```bash
go run cmd/examples/07-router/guards/main.go
```

## Authentication Flow

1. **Try accessing protected route** (key 2 or 3)
2. **Guard intercepts** - Checks if authenticated
3. **Redirect to login** - Saves intended destination in query params
4. **Enter credentials** - Press Enter to activate form
5. **Submit login** - Any non-empty username/password works
6. **Redirect to destination** - Returns to originally requested page

## Code Highlights

### Authentication Guard

```go
authGuard := func(to, from *router.Route, next router.NextFunc) {
    requiresAuth, ok := to.Meta["requiresAuth"].(bool)
    
    if ok && requiresAuth && !isAuthenticated {
        // Redirect to login with intended destination
        next(&router.NavigationTarget{
            Path: "/login",
            Query: map[string]string{
                "redirect": to.FullPath,
            },
        })
    } else {
        // Allow navigation
        next(nil)
    }
}
```

### Protected Route Configuration

```go
RouteWithOptions("/dashboard",
    router.WithName("dashboard"),
    router.WithComponent(createDashboardComponent()),
    router.WithMeta(map[string]interface{}{
        "requiresAuth": true,
    }),
)
```

### Register Global Guard

```go
r, err := router.NewRouterBuilder().
    // ... routes ...
    BeforeEach(authGuard).
    Build()
```

## What to Try

1. **Access protected route without auth** - Try key 2 or 3, observe redirect to login
2. **Login** - Press Enter on login page, fill form, submit
3. **Access protected routes** - Now keys 2 and 3 work
4. **Logout** - Press key 5, try accessing protected routes again
5. **History navigation** - Use back/forward after redirects

## Components Used

- **Card** - Content containers
- **Badge** - Auth status and current route indicators

## Next Steps

- See `../basic/` for basic navigation patterns
- See `../nested/` for nested routes and layouts
