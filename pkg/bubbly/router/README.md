# BubblyUI Router

**Package Path:** `github.com/newbpydev/bubblyui/pkg/bubbly/router`  
**Version:** 3.0  
**Purpose:** Vue Router-inspired routing with type-safe params, guards, and nested routes

## Overview

Router provides declarative routing for BubblyUI with dynamic parameters, query strings, route guards, named routes, nested routes, and full navigation control.

## Quick Start

```go
import csrouter "github.com/newbpydev/bubblyui/pkg/bubbly/router"

// Create router
router := csrouter.NewRouter().
    AddRoute("/", homeComponent).
    AddRoute("/users", userListComponent).
    AddRoute("/users/:id", userDetailComponent).
    WithNotFound(notFoundComponent).
    Build()

// Navigate
router.Navigate("/users/123")
```

## Features

### 1. Route Parameters

```go
// Pattern: /users/:id
router.AddRoute("/users/:id", func() (bubbly.Component, error) {
    return NewComponent("User").
        Setup(func(ctx *Context) {
            route := router.CurrentRoute()
            userId := route.Params["id"]
        }).
        Build()
})

// Multiple params: /users/:userId/posts/:postId
// Constraints: /users/:id(\\d+)
```

### 2. Query Parameters

```go
// URL: /search?q=hello&page=2
route := router.CurrentRoute()
query := route.Query

q := query.Get("q")                    // "hello"
page := query.GetDefault("page", "1")  // "2"
tags := query.GetAll("tag")            // ["go", "tui"]
```

### 3. Route Guards

```go
// Auth guard
authGuard := func(ctx *csrouter.GuardContext) bool {
    if !ctx.Inject("isAuthenticated", false).(bool) {
        ctx.Navigate("/login")
        return false
    }
    return true
}

router := NewRouter().
    WithGuard(authGuard).  // Global
    AddRoute("/admin", adminComponent, adminGuard).  // Route-specific
    Build()
```

### 4. Named Routes

```go
router.AddNamedRoute("userProfile", "/users/:id", userComponent)
router.NavigateTo("userProfile", map[string]string{"id": "123"})
```

### 5. Navigation

```go
router.Navigate("/users/123")       // Push to history
router.GoBack()                      // Back button
router.GoForward()                   // Forward button
router.Replace("/login")             // Replace current
router.CurrentRoute()                // Current route
```

### 6. Nested Routes

```go
// /dashboard/* routes
dashboard := NewRouter().
    AddRoute("/", dashboardHome).
    AddRoute("/settings", settings).
    AddRoute("/users", users).
    Build()

router.AddRoute("/dashboard", dashboard)
```

### 7. Router Composables

```go
// Use in Setup()
router := composables.UseRouter(ctx)
params := composables.UseParams(ctx)
route := composables.UseRoute(ctx)
nav := composables.UseNavigation(ctx)
```

## Example: Protected Dashboard

```go
func createRouter() *csrouter.Router {
    authGuard := func(ctx *csrouter.GuardContext) bool {
        auth := ctx.Inject("auth", Auth{}).(Auth)
        if !auth.IsAuthenticated {
            ctx.Navigate("/login")
            return false
        }
        return true
    }
    
    adminGuard := func(ctx *csrouter.GuardContext) bool {
        role := ctx.Inject("userRole", "").(string)
        if role != "admin" {
            ctx.Navigate("/unauthorized")
            return false
        }
        return true
    }
    
    return csrouter.NewRouter().
        WithGuard(authGuard).
        AddRoute("/login", loginComponent).
        AddRoute("/dashboard", dashboardComponent).
        AddRoute("/admin", adminComponent, adminGuard).
        AddRoute("/users/:id", userDetailComponent).
        WithNotFound(notFoundComponent).
        Build()
}
```

## Performance

```
Route matching:      < 1μs per lookup
Param extraction:    < 500ns per route
Navigation:          < 10μs per transition
Memory:              O(depth) where depth < 10
```

**Package:** 14,415 LOC | 30+ files | Complete routing system

## API Reference

See [Full Router API](docs/api/router.md) for complete documentation.