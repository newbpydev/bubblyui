# User Workflow: Router System

## Developer Personas

### Persona 1: Web Developer (Sarah)
- **Background**: 5 years Vue/React, building first TUI app
- **Goal**: Multi-screen dashboard with authentication
- **Familiar With**: Vue Router, React Router
- **Expects**: Similar routing concepts in TUI
- **Success**: Routes work like Vue Router

### Persona 2: Go Developer (Marcus)  
- **Background**: 3 years Go, no routing experience
- **Goal**: CLI tool with multiple commands → screens
- **Familiar With**: CLI flags, subcommands
- **Expects**: Type-safe, Go-idiomatic routing
- **Success**: Routes compile, no runtime surprises

### Persona 3: TUI Expert (Jin)
- **Background**: Built many Bubbletea apps
- **Goal**: Add navigation to existing app
- **Familiar With**: Bubbletea models, commands
- **Expects**: Seamless Bubbletea integration
- **Success**: Router fits existing patterns

---

## Primary User Journey: First Multi-Screen App

### Entry Point: Single-Screen to Multi-Screen

**Workflow: Adding Basic Routing**

#### Step 1: Install and Import
**User Action**: Add router to existing application
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"
)
```

**System Response**:
- Package imports successfully
- Router types available

**UI Feedback**:
- No visual change yet
- Code compiles

#### Step 2: Define Routes
**User Action**: Create basic routes
```go
homeComponent, _ := bubbly.NewComponent("Home").
    Template(func(ctx bubbly.RenderContext) string {
        return "Home Screen\nPress 'a' for About"
    }).
    Build()

aboutComponent, _ := bubbly.NewComponent("About").
    Template(func(ctx bubbly.RenderContext) string {
        return "About Screen\nPress 'h' for Home"
    }).
    Build()

r := router.NewRouter().
    Route("/", homeComponent).
    Route("/about", aboutComponent).
    Build()
```

**System Response**:
- Routes registered
- Router created
- Components linked to paths

**UI Feedback**:
- Still on home screen
- Ready for navigation

#### Step 3: Connect to Application
**User Action**: Integrate router with Bubbletea model
```go
type model struct {
    router *router.Router
}

func (m model) Init() tea.Cmd {
    return m.router.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "a":
            return m, m.router.Push(&router.NavigationTarget{Path: "/about"})
        case "h":
            return m, m.router.Push(&router.NavigationTarget{Path: "/"})
        }
    case router.RouteChangedMsg:
        // Route changed, trigger re-render
    }
    
    return m, nil
}

func (m model) View() string {
    return m.router.CurrentComponent().View()
}
```

**System Response**:
- Router integrated with model
- Navigation commands generated
- Route changes trigger updates

**UI Feedback**:
- Press 'a' → navigates to About screen
- Press 'h' → navigates to Home screen
- Smooth screen transitions

**Journey Branch**:
- ✅ Success → Step 4 (Add Parameters)
- ❌ Error → Check route paths, component setup

---

### Feature Journey: Dynamic Routes with Parameters

#### Step 4: Add Parameterized Route
**User Action**: Create user detail route
```go
userDetailComponent, _ := bubbly.NewComponent("UserDetail").
    Setup(func(ctx *bubbly.Context) {
        route := router.UseRoute(ctx)
        userId := ctx.Ref("")
        
        ctx.OnMounted(func() {
            userId.Set(route.Get().Params["id"])
        })
        
        ctx.Expose("userId", userId)
    }).
    Template(func(ctx bubbly.RenderContext) string {
        userId := ctx.Get("userId").(*bubbly.Ref[interface{}])
        return fmt.Sprintf("User Detail\nID: %s", userId.Get())
    }).
    Build()

r := router.NewRouter().
    Route("/", homeComponent).
    Route("/user/:id", userDetailComponent).
    Build()
```

**System Response**:
- Route with parameter registered
- Parameter extraction configured
- Component receives param

**UI Feedback**:
- Navigate to `/user/123` → shows "User Detail, ID: 123"
- Navigate to `/user/456` → shows "User Detail, ID: 456"
- Parameters extracted automatically

**Data Flow**:
```
User navigates → router.Push({Path: "/user/123"})
→ Router matches "/user/:id" pattern
→ Extracts params: {id: "123"}
→ Creates Route object with params
→ Component receives route via UseRoute
→ onMounted extracts param
→ UI displays user ID
```

---

### Feature Journey: Navigation Guards

#### Step 5: Add Authentication Guard
**User Action**: Protect routes with authentication
```go
var isAuthenticated = false

r := router.NewRouter().
    BeforeEach(func(to, from *router.Route, next router.NextFunc) {
        // Check if route requires auth
        if to.Meta["requiresAuth"] == true && !isAuthenticated {
            // Redirect to login
            next(&router.NavigationTarget{
                Path: "/login",
                Query: map[string]string{"redirect": to.FullPath},
            })
        } else {
            // Allow navigation
            next(nil)
        }
    }).
    Route("/", homeComponent).
    Route("/login", loginComponent).
    Route("/dashboard", dashboardComponent, router.Meta{
        "requiresAuth": true,
    }).
    Build()
```

**System Response**:
- Global guard registered
- Executes before every navigation
- Can cancel, redirect, or allow

**UI Feedback**:
- Navigate to `/dashboard` while not logged in → redirects to `/login`
- After login, navigate to `/dashboard` → success
- Login screen shows "redirect" query param

**Guard Execution Flow**:
```
User navigates to /dashboard
→ router.Push({Path: "/dashboard"})
→ BeforeEach guard executes
→ Checks requiresAuth === true
→ Checks isAuthenticated === false
→ Calls next({Path: "/login", Query: {redirect: "/dashboard"}})
→ Navigation redirects to /login
→ Original navigation cancelled
→ New navigation to /login starts
```

---

### Feature Journey: Nested Routes

#### Step 6: Create Dashboard with Nested Views
**User Action**: Set up dashboard with tabs
```go
dashboardLayout, _ := bubbly.NewComponent("DashboardLayout").
    Setup(func(ctx *bubbly.Context) {
        // Router view for nested routes
        ctx.Expose("routerView", router.NewRouterView(1)) // depth 1
    }).
    Template(func(ctx bubbly.RenderContext) string {
        routerView := ctx.Get("routerView").(*router.RouterView)
        
        return lipgloss.JoinVertical(
            lipgloss.Top,
            "Dashboard Header",
            "───────────────────",
            routerView.View(), // Nested route renders here
        )
    }).
    Build()

statsComponent, _ := bubbly.NewComponent("Stats").
    Template(func(ctx bubbly.RenderContext) string {
        return "Statistics View"
    }).
    Build()

settingsComponent, _ := bubbly.NewComponent("Settings").
    Template(func(ctx bubbly.RenderContext) string {
        return "Settings View"
    }).
    Build()

r := router.NewRouter().
    Route("/dashboard", dashboardLayout, 
        router.Children(
            router.Child("/stats", statsComponent),
            router.Child("/settings", settingsComponent),
        ),
    ).
    Build()
```

**System Response**:
- Parent route registered
- Child routes registered
- Nested component rendering configured

**UI Feedback**:
- Navigate to `/dashboard/stats` → Dashboard header + Stats view
- Navigate to `/dashboard/settings` → Dashboard header + Settings view
- Parent layout stays constant, child view changes

**Nested Route Rendering**:
```
Route: /dashboard/settings
    ↓
Matched routes: [dashboard, dashboard/settings]
    ↓
RouterView(depth=0) → renders DashboardLayout
    ↓
DashboardLayout includes RouterView(depth=1)
    ↓
RouterView(depth=1) → renders Settings component
    ↓
Final UI: Dashboard header + Settings view
```

---

### Feature Journey: Programmatic Navigation

#### Step 7: Navigate from Event Handlers
**User Action**: Navigate on button click or form submit
```go
searchComponent, _ := bubbly.NewComponent("Search").
    Setup(func(ctx *bubbly.Context) {
        query := ctx.Ref("")
        router := router.UseRouter(ctx)
        
        ctx.On("search", func(data interface{}) {
            searchQuery := query.Get().(string)
            
            // Navigate with query string
            router.Push(&router.NavigationTarget{
                Path: "/results",
                Query: map[string]string{
                    "q": searchQuery,
                    "page": "1",
                },
            })
        })
        
        ctx.On("view-user", func(data interface{}) {
            userId := data.(string)
            
            // Navigate by name
            router.Push(&router.NavigationTarget{
                Name: "user-detail",
                Params: map[string]string{"id": userId},
            })
        })
        
        ctx.Expose("query", query)
    }).
    Build()
```

**System Response**:
- Router accessible via composable
- Navigation triggered from events
- Commands generated automatically

**UI Feedback**:
- User types "golang" and clicks search
- Navigate to `/results?q=golang&page=1`
- Results page receives query params
- Smooth navigation without page reload

---

## Alternative Workflows

### Workflow A: Adding Router to Existing App

#### Entry: Existing Single-Screen Bubbletea App

1. **Install Router Package**
   ```bash
   go get github.com/newbpydev/bubblyui/pkg/bubbly/router
   ```

2. **Convert Existing Component to Route**
   ```go
   // Before: Single component
   component := myExistingComponent()
   
   // After: Same component as a route
   router := router.NewRouter().
       Route("/", myExistingComponent()).
       Build()
   ```

3. **Update Model to Use Router**
   ```go
   // Before:
   func (m model) View() string {
       return m.component.View()
   }
   
   // After:
   func (m model) View() string {
       return m.router.CurrentComponent().View()
   }
   ```

4. **Add New Routes Incrementally**
   ```go
   router.AddRoute("/settings", settingsComponent)
   router.AddRoute("/about", aboutComponent)
   ```

5. **Test Each Route**
   - Verify navigation works
   - Check state preservation
   - Ensure cleanup happens

**Time**: 30-60 minutes for basic integration

---

### Workflow B: Building Multi-Screen CLI Tool

#### Entry: CLI Tool with Subcommands → Screens

1. **Map Commands to Routes**
   ```
   CLI: tool status  → Route: /status
   CLI: tool logs    → Route: /logs
   CLI: tool config  → Route: /config
   ```

2. **Create Route for Each Screen**
   ```go
   router := router.NewRouter().
       Route("/status", statusScreen).
       Route("/logs", logsScreen).
       Route("/config", configScreen).
       Build()
   ```

3. **Parse CLI Args to Initial Route**
   ```go
   func main() {
       initialRoute := "/"
       if len(os.Args) > 1 {
           initialRoute = "/" + os.Args[1]
       }
       
       router.Push(&router.NavigationTarget{Path: initialRoute})
       tea.NewProgram(model{router: router}).Run()
   }
   ```

4. **Add Keyboard Navigation Between Screens**
   ```go
   // In Update():
   case "1": return m, m.router.Push(&router.NavigationTarget{Path: "/status"})
   case "2": return m, m.router.Push(&router.NavigationTarget{Path: "/logs"})
   case "3": return m, m.router.Push(&router.NavigationTarget{Path: "/config"})
   ```

**Time**: 1-2 hours for CLI tool conversion

---

## Error Recovery Workflows

### Error Flow 1: Route Not Found

**Trigger**: Navigate to non-existent route
```go
router.Push(&router.NavigationTarget{Path: "/doesnt-exist"})
```

**User Sees**:
```
Error: Route not found: /doesnt-exist

Available routes:
  / - Home
  /about - About
  /user/:id - User Detail
```

**Recovery**:
1. Add 404 route:
   ```go
   router.Route("/:catchAll*", notFoundComponent)
   ```
2. Handle in guard:
   ```go
   router.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
       if to == nil {
           next(&router.NavigationTarget{Path: "/"})
       } else {
           next(nil)
       }
   })
   ```

---

### Error Flow 2: Guard Rejection

**Trigger**: Guard cancels navigation
```go
router.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
    if someCondition {
        next(&router.NavigationTarget{}) // Empty = cancel
    }
})
```

**User Sees**:
- Stays on current screen
- Optional: Toast notification "Navigation cancelled"

**Recovery**:
1. Check guard logic
2. Add feedback to user
3. Provide alternative path

---

### Error Flow 3: Circular Redirect

**Trigger**: Guard redirects infinitely
```go
router.BeforeEach(func(to, from *router.Route, next router.NextFunc) {
    next(&router.NavigationTarget{Path: "/other"})
})
// /other's guard also redirects, creating loop
```

**User Sees**:
```
Error: Maximum redirects exceeded (10)
Route: /initial → /other → /another → /other → ...

Check your navigation guards for circular redirects.
```

**Recovery**:
1. Review guard logic
2. Add redirect count check
3. Fix circular dependency

---

## State Transition Diagrams

### Navigation Lifecycle
```
Idle (on current route)
    ↓
Navigation Requested (router.Push())
    ↓
Guards Executing (BeforeEach, BeforeEnter)
    ├─ Allowed → Continue
    ├─ Cancelled → Return to Idle
    └─ Redirected → Start new navigation
    ↓
Route Matched
    ↓
Component Transition
    ├─ Old component BeforeRouteLeave
    ├─ New component BeforeRouteEnter
    └─ Or component BeforeRouteUpdate
    ↓
History Updated
    ↓
Route Changed (RouteChangedMsg)
    ↓
After Guards Execute (AfterEach)
    ↓
Idle (on new route)
```

### Guard Decision Flow
```
                ┌─────────────┐
                │  next(nil)  │
                └──────┬──────┘
                       ↓
                ┌─────────────────┐
                │ Allow Navigation│
                └─────────────────┘

                ┌─────────────┐
                │  next({})   │
                └──────┬──────┘
                       ↓
                ┌─────────────────┐
                │Cancel Navigation│
                └─────────────────┘

                ┌─────────────┐
                │next({Path}) │
                └──────┬──────┘
                       ↓
                ┌─────────────────┐
                │Redirect to Path │
                └─────────────────┘
```

---

## Integration Points Map

### Feature Cross-Reference
```
07-router
    ← Uses: 01-reactivity-system (route state)
    ← Uses: 02-component-model (route components)
    ← Uses: 03-lifecycle-hooks (component guards)
    ← Uses: 04-composition-api (useRouter, useRoute)
    → Used by: 06-built-in-components (RouterView, RouterLink)
    → Used by: 09-dev-tools (route debugging)
    → Used by: Applications (multi-screen apps)
```

---

## User Success Paths

### Path 1: Quick Win (< 1 hour)
```
Single screen → Add router → Define 2 routes → Navigate → Success! 🎉
Features used: 00, 01, 02, 07 (basic)
```

### Path 2: Intermediate (< 4 hours)
```
Basic routing → Add params → Add guards → Nested routes → Success! 🎉
Features used: 00-07 (all routing features)
```

### Path 3: Production App (< 2 days)
```
Full app → Auth guards → Dashboard → Nested views → History → Deploy → Success! 🎉
Features used: ALL + real-world patterns
```

---

## Common Patterns

### Pattern 1: Authentication Flow
```go
// 1. Global auth guard
router.BeforeEach(authGuard)

// 2. Login success → redirect to intended page
func handleLogin(router *router.Router, redirectPath string) {
    isAuthenticated = true
    if redirectPath != "" {
        router.Push(&router.NavigationTarget{Path: redirectPath})
    } else {
        router.Push(&router.NavigationTarget{Path: "/"})
    }
}

// 3. Logout → go to login
func handleLogout(router *router.Router) {
    isAuthenticated = false
    router.Push(&router.NavigationTarget{Path: "/login"})
}
```

### Pattern 2: Route-based Data Fetching
```go
component.Setup(func(ctx *bubbly.Context) {
    data := ctx.Ref[*Data](nil)
    route := router.UseRoute(ctx)
    
    // Fetch data when route params change
    ctx.OnUpdated(func() {
        userId := route.Get().Params["id"]
        // Emit fetch event with userId
        ctx.Emit("fetch-user", userId)
    }, route)
})
```

### Pattern 3: Breadcrumb Navigation
```go
func buildBreadcrumbs(route *router.Route) []Breadcrumb {
    breadcrumbs := []Breadcrumb{}
    
    for _, matched := range route.Matched {
        breadcrumbs = append(breadcrumbs, Breadcrumb{
            Name: matched.Meta["title"].(string),
            Path: matched.Path,
        })
    }
    
    return breadcrumbs
}
```

---

## Summary

The Router system enables multi-screen TUI applications with familiar Vue Router patterns adapted for terminal interfaces. Developers can define routes declaratively, use navigation guards for access control, handle dynamic parameters, and build nested route hierarchies. The system integrates seamlessly with Bubbletea's command pattern and BubblyUI's reactive components, providing a smooth development experience from simple two-screen apps to complex multi-view dashboards.

**Key Success Factors**:
- ✅ Familiar API for web developers
- ✅ Type-safe for Go developers
- ✅ Bubbletea-native for TUI experts
- ✅ Progressive complexity (start simple, add features)
- ✅ Clear error messages and recovery paths
- ✅ Production-ready patterns
