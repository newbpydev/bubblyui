# Design Specification: Router System

## Component Hierarchy

```
Enabler System (Foundation)
└── Router
    ├── Route Matcher (core algorithm)
    ├── History Manager (stack management)
    ├── Guard Executor (navigation control)
    ├── Route Object (current state)
    └── Navigation Composer (Bubbletea integration)
```

This is a foundational system that enables multi-screen navigation, not a visual component itself.

---

## Architecture Overview

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     Application Layer                         │
│  (Components use router for navigation)                      │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                       Router System                           │
├──────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐    ┌────────────────┐ │
│  │    Router    │───→│     Route    │←───│  Navigation    │ │
│  │  (singleton) │    │   (current)  │    │    Guards      │ │
│  └──────┬───────┘    └──────────────┘    └────────────────┘ │
│         │                                                     │
│         ├──→ ┌──────────────┐  ┌─────────────┐             │
│         │    │ Route Matcher│  │   History   │             │
│         │    │              │  │   Manager   │             │
│         │    └──────────────┘  └─────────────┘             │
│         │                                                     │
│         └──→ ┌──────────────────────────────────┐           │
│              │   Navigation Command Generator    │           │
│              └──────────────────────────────────┘           │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                      Bubbletea Framework                      │
│  (Receives navigation commands, triggers updates)            │
└──────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Navigation Flow

```
User Action (key press, event)
    ↓
Event Handler calls router.Push()
    ↓
Router validates navigation target
    ↓
Execute Global Before Guards
    ├─ next() called → continue
    ├─ next(false) → cancel navigation
    ├─ next(path) → redirect
    └─ error → abort with error
    ↓
Execute Route-specific beforeEnter Guard
    ↓
Resolve Target Component
    ↓
Execute Component BeforeRouteLeave (old component)
    ↓
Execute Component BeforeRouteEnter (new component)
    ↓
Update History Stack
    ↓
Update Current Route Object
    ↓
Execute Global After Guards
    ↓
Generate Navigation Command (tea.Cmd)
    ↓
Return command to Bubbletea
    ↓
Bubbletea processes command
    ↓
Component Update() receives RouteChangedMsg
    ↓
Component renders new route
    ↓
View() displays new screen
```

### Route Matching Flow

```
Navigation Target Path
    ↓
Normalize Path (/foo//bar → /foo/bar)
    ↓
Split into segments
    ↓
For each registered route:
    ├─ Match static segments
    ├─ Extract dynamic params (:id)
    ├─ Match optional params (:id?)
    ├─ Match wildcards (:path*)
    └─ Calculate match score
    ↓
Sort matches by specificity
    ↓
Return best match or 404
    ↓
Parse query string
    ↓
Parse hash fragment
    ↓
Create Route object
```

---

## Type Definitions

### Core Types

```go
// Router is the singleton router instance
type Router struct {
    routes         []*RouteRecord
    history        *History
    currentRoute   *Route
    beforeEachHooks []NavigationGuard
    afterEachHooks  []AfterNavigationHook
    mu             sync.RWMutex
}

// RouteRecord defines a route configuration
type RouteRecord struct {
    Path      string
    Name      string
    Component bubbly.Component
    Children  []*RouteRecord
    BeforeEnter NavigationGuard
    Meta      map[string]interface{}
    
    // Compiled pattern for matching
    pattern   *routePattern
}

// Route represents the current active route
type Route struct {
    Path     string
    Name     string
    Params   map[string]string
    Query    map[string]string
    Hash     string
    Meta     map[string]interface{}
    Matched  []*RouteRecord  // For nested routes
    FullPath string
}

// NavigationTarget specifies where to navigate
type NavigationTarget struct {
    Path   string              // e.g., "/user/123"
    Name   string              // e.g., "user-detail"
    Params map[string]string   // e.g., {"id": "123"}
    Query  map[string]string   // e.g., {"tab": "settings"}
    Hash   string              // e.g., "#section"
}

// NavigationGuard is a before guard function
type NavigationGuard func(to, from *Route, next NextFunc)

// NextFunc controls navigation flow
type NextFunc func(target *NavigationTarget)

// AfterNavigationHook executes after navigation
type AfterNavigationHook func(to, from *Route)

// History manages the navigation stack
type History struct {
    entries []*HistoryEntry
    current int
    mu      sync.Mutex
}

// HistoryEntry represents a history stack entry
type HistoryEntry struct {
    Route *Route
    State interface{}  // Optional state
}

// NavigationCommand is a Bubbletea command
type NavigationCommand func() tea.Msg

// RouteChangedMsg signals a route change
type RouteChangedMsg struct {
    To   *Route
    From *Route
}
```

---

## Route Matching Algorithm

### Path Pattern Compilation

```go
// routePattern is the compiled pattern
type routePattern struct {
    segments []segment
    regex    *regexp.Regexp
}

type segment struct {
    kind     segmentKind  // static, param, optional, wildcard
    name     string       // param name (if dynamic)
    value    string       // static value (if static)
}

type segmentKind int

const (
    segmentStatic segmentKind = iota
    segmentParam               // :id
    segmentOptional            // :id?
    segmentWildcard            // :path*
)

// Compilation examples:
// "/users/:id"       → [static("users"), param("id")]
// "/docs/:path*"     → [static("docs"), wildcard("path")]
// "/posts/:id?"      → [static("posts"), optional("id")]
```

### Match Scoring

```go
// Routes are scored by specificity
type matchScore struct {
    staticSegments  int  // Higher = more specific
    paramSegments   int  // Lower = more specific
    wildcardSegments int  // Lowest specificity
}

// Example scoring:
// "/users/123"      → score{static: 2, param: 0, wildcard: 0}  // Best
// "/users/:id"      → score{static: 1, param: 1, wildcard: 0}  // Medium
// "/:path*"         → score{static: 0, param: 0, wildcard: 1}  // Worst
```

---

## Navigation Guard System

### Guard Execution Order

```
1. Global Before Guards (router.BeforeEach)
   ↓
2. Route-specific Before Guard (route.BeforeEnter)
   ↓
3. Component BeforeRouteLeave (old component)
   ↓
4. Component BeforeRouteEnter (new component)
   ↓
5. Component BeforeRouteUpdate (if component reused)
   ↓
6. Navigation Confirmed
   ↓
7. Global After Guards (router.AfterEach)
```

### Guard Control Flow

```go
// Guard can do 4 things:
func myGuard(to, from *Route, next NextFunc) {
    // 1. Allow navigation
    next(nil)
    
    // 2. Cancel navigation
    next(&NavigationTarget{}) // Empty target = cancel
    
    // 3. Redirect
    next(&NavigationTarget{Path: "/login"})
    
    // 4. Error (panic or return error if using error-returning variant)
    panic("navigation not allowed")
}
```

### Guard Cancellation

```go
// If navigation is cancelled or redirected during guard execution,
// remaining guards are skipped and the new navigation starts
type guardResult struct {
    action guardAction
    target *NavigationTarget
    err    error
}

type guardAction int

const (
    guardContinue guardAction = iota
    guardCancel
    guardRedirect
    guardError
)
```

---

## History Management

### History Stack

```go
type History struct {
    entries []*HistoryEntry
    current int  // Index in entries
    maxSize int  // Optional limit
    mu      sync.Mutex
}

// Operations:
// Push: Add new entry, truncate forward history
// Replace: Replace current entry
// Back: Move current--
// Forward: Move current++
// Go(n): Move current += n
```

### History Diagram

```
Initial: [Home] ← current

After Push(/about): [Home, About] ← current

After Push(/contact): [Home, About, Contact] ← current

After Back: [Home, About ← current, Contact]

After Push(/faq): [Home, About ← current, FAQ]
  // Contact is removed (forward history truncated)
```

---

## Bubbletea Integration

### Navigation Command Pattern

```go
// Router generates commands for navigation
func (r *Router) Push(target *NavigationTarget) tea.Cmd {
    return func() tea.Msg {
        // Validate and match route
        newRoute, err := r.match(target)
        if err != nil {
            return RouteErrorMsg{Error: err}
        }
        
        // Execute guards
        if !r.executeGuards(newRoute) {
            return NavigationCancelledMsg{}
        }
        
        // Update history
        r.history.Push(newRoute)
        
        // Update current route
        oldRoute := r.currentRoute
        r.currentRoute = newRoute
        
        // Return message
        return RouteChangedMsg{
            To:   newRoute,
            From: oldRoute,
        }
    }
}
```

### Component Integration

```go
// Components receive route changes via Update()
func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case RouteChangedMsg:
        // Update route-aware state
        if c.routeComponent != nil {
            return c.routeComponent.Update(msg)
        }
    }
    
    return c, nil
}
```

---

## Composable API

### useRouter Composable

```go
// Provides access to router instance
func useRouter(ctx *bubbly.Context) *Router {
    // Get router from context injection
    router, ok := ctx.Inject("router").(*Router)
    if !ok {
        panic("router not provided in context")
    }
    return router
}
```

### useRoute Composable

```go
// Provides reactive access to current route
func useRoute(ctx *bubbly.Context) *bubbly.Ref[*Route] {
    router := useRouter(ctx)
    
    // Create reactive ref
    routeRef := ctx.Ref(router.CurrentRoute())
    
    // Update on route changes
    ctx.OnMounted(func() {
        router.AfterEach(func(to, from *Route) {
            routeRef.Set(to)
        })
    })
    
    return routeRef
}
```

---

## Route Component Rendering

### RouterView Component

```go
// RouterView renders the current route's component
type RouterView struct {
    router *Router
    depth  int  // For nested routes
}

func (rv *RouterView) View() string {
    route := rv.router.CurrentRoute()
    
    // Get matched route for this depth
    if rv.depth >= len(route.Matched) {
        return "" // No route at this depth
    }
    
    matchedRoute := route.Matched[rv.depth]
    component := matchedRoute.Component
    
    return component.View()
}
```

### Nested Route Rendering

```go
// Parent route component includes child RouterView
parentComponent := bubbly.NewComponent("Dashboard").
    Template(func(ctx bubbly.RenderContext) string {
        return lipgloss.JoinVertical(
            lipgloss.Top,
            "Dashboard Header",
            "───────────────────",
            // Child routes render here
            ctx.Get("childRouter").(*RouterView).View(),
        )
    }).
    Build()
```

---

## URL Building

### Path Generation

```go
// Build URL from route name and params
func (r *Router) BuildPath(name string, params map[string]string, query map[string]string) string {
    route := r.findRouteByName(name)
    if route == nil {
        return ""
    }
    
    // Replace params in path
    path := route.Path
    for key, value := range params {
        path = strings.Replace(path, ":"+key, value, 1)
    }
    
    // Add query string
    if len(query) > 0 {
        path += "?" + buildQueryString(query)
    }
    
    return path
}
```

---

## Error Handling

### Navigation Errors

```go
type NavigationError struct {
    Code    ErrorCode
    Message string
    From    *Route
    To      *NavigationTarget
}

type ErrorCode int

const (
    ErrRouteNotFound ErrorCode = iota
    ErrInvalidPath
    ErrGuardRejected
    ErrCircularRedirect
    ErrComponentNotFound
)

// Error handling in guards
router.BeforeEach(func(to, from *Route, next NextFunc) {
    defer func() {
        if r := recover(); r != nil {
            // Report to observability
            if reporter := observability.GetErrorReporter(); reporter != nil {
                reporter.ReportPanic(&observability.NavigationPanicError{
                    RoutePath: to.Path,
                    GuardName: "global-before",
                    PanicValue: r,
                }, &observability.ErrorContext{
                    ComponentName: "router",
                    Timestamp: time.Now(),
                    StackTrace: debug.Stack(),
                })
            }
            // Cancel navigation on panic
            next(&NavigationTarget{})
        }
    }()
    
    // Guard logic...
})
```

---

## Performance Optimizations

### Route Matching Cache

```go
// Cache recent matches
type routeCache struct {
    cache map[string]*Route
    mu    sync.RWMutex
    maxSize int
}

func (rc *routeCache) get(path string) (*Route, bool) {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    route, ok := rc.cache[path]
    return route, ok
}
```

### Lazy Pattern Compilation

```go
// Compile route patterns on first use
func (r *RouteRecord) getPattern() *routePattern {
    if r.pattern == nil {
        r.pattern = compilePattern(r.Path)
    }
    return r.pattern
}
```

---

## Known Limitations & Solutions

### Limitation 1: No Browser History API
**Problem**: Cannot use pushState/replaceState/popstate  
**Current Design**: Custom history stack implementation  
**Solution**: In-memory history with forward/back support  
**Benefits**: Full control, TUI-optimized, no browser dependencies  
**Priority**: N/A - by design

### Limitation 2: Route Param Type Safety
**Problem**: Params are strings, need type conversion  
**Current Design**: String map for params  
**Solution Design**: Generic param extractors with validation  
**Benefits**: Type-safe param access  
**Priority**: MEDIUM - Phase 2 enhancement
```go
// Future API:
id := route.Param("id").Int()  // Returns int, panics if invalid
userID := route.Param("userId").UUID()  // Returns UUID
```

### Limitation 3: Route Lazy Loading
**Problem**: All routes loaded upfront  
**Current Design**: All components initialized at router creation  
**Solution**: Not needed for TUI - apps are small, startup is fast  
**Priority**: LOW - TUI constraint

### Limitation 4: Circular Redirect Detection
**Problem**: Guard can redirect infinitely  
**Current Design**: No detection  
**Solution Design**: Track redirect chain, break after N redirects  
**Benefits**: Prevents infinite loops  
**Priority**: HIGH - must have before v1.0
```go
const maxRedirects = 10

func (r *Router) navigate(...) {
    redirectCount := 0
    for {
        if redirectCount > maxRedirects {
            return ErrCircularRedirect
        }
        // Execute guards, check for redirects
        redirectCount++
    }
}
```

---

## Future Enhancements

### Phase 4: Advanced Features
1. **Route Transitions**: Animate route changes
2. **Persistent History**: Save/restore navigation history
3. **Multi-Router**: Multiple routers for tabs/splits
4. **Route Metadata Generation**: Extract from code comments
5. **Advanced Matching**: Regex patterns, custom matchers

### Phase 5: Developer Tools
1. **Route Inspector**: View route tree, current route, history
2. **Navigation Timeline**: See all navigation events
3. **Guard Profiler**: Measure guard execution time
4. **Route Validator**: Check for orphaned routes, missing components

---

## Integration Patterns

### Pattern 1: Authentication Guard

```go
router.BeforeEach(func(to, from *Route, next NextFunc) {
    requiresAuth := to.Meta["requiresAuth"]
    
    if requiresAuth == true && !auth.IsAuthenticated() {
        // Redirect to login, save intended destination
        next(&NavigationTarget{
            Path: "/login",
            Query: map[string]string{
                "redirect": to.FullPath,
            },
        })
    } else {
        next(nil)
    }
})
```

### Pattern 2: Route-based Focus Management

```go
router.AfterEach(func(to, from *Route) {
    // Set focus based on route meta
    if focusTarget, ok := to.Meta["focusTarget"].(string); ok {
        focusManager.SetFocus(focusTarget)
    }
})
```

### Pattern 3: Nested Dashboard Routes

```go
dashboardRoutes := []*RouteRecord{
    {
        Path: "/dashboard",
        Component: dashboardLayout,
        Children: []*RouteRecord{
            {Path: "overview", Component: overviewPage},
            {Path: "stats", Component: statsPage},
            {Path: "settings", Component: settingsPage},
        },
    },
}
```

### Pattern 4: Programmatic Navigation in Composable

```go
func useNavigation(ctx *bubbly.Context) NavigationHelper {
    router := useRouter(ctx)
    
    return NavigationHelper{
        goToUser: func(id string) {
            router.Push(&NavigationTarget{
                Name: "user-detail",
                Params: map[string]string{"id": id},
            })
        },
        goBack: func() {
            router.Back()
        },
    }
}
```

---

## Testing Strategy

### Unit Tests
- Route pattern matching
- Parameter extraction
- Query string parsing
- History stack operations
- Guard execution order
- Circular redirect detection

### Integration Tests
- Complete navigation flows
- Guard combinations
- Nested route navigation
- History back/forward
- Error recovery

### Performance Tests
- Route matching speed
- Navigation overhead
- Memory usage per route
- Guard execution time

---

## Summary

The Router system provides Vue Router-inspired navigation for TUI applications while respecting TUI constraints and Bubbletea's architecture. It uses custom history management, command-based navigation, and type-safe patterns to enable multi-screen applications with guards, nested routes, and programmatic control. The design prioritizes simplicity, performance, and seamless integration with existing BubblyUI features.
