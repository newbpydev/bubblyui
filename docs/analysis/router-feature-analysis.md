# Router System Feature Analysis for BubblyUI

**Document Type**: Critical Feature Analysis  
**Date**: 2025-11-03  
**Status**: For Review Before Task 1.2  
**Author**: AI Analysis based on specs/07-router/

---

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [TUI Pain Points Analysis](#tui-pain-points-analysis)
4. [Router Feature Scope](#router-feature-scope)
5. [Use Case Comparison: Before/After](#use-case-comparison-beforeafter)
6. [Critical Analysis](#critical-analysis)
7. [Recommendations](#recommendations)

---

## 1. Executive Summary

### The Question
Is a Vue Router-inspired routing system (with path params, guards, history, nested routes) worth building for a TUI framework based on Bubbletea?

### Quick Answer
**YES, but with significant simplification needed.**

**Why YES:**
- Addresses real pain point: complex TUI apps have no navigation structure
- Provides familiar mental model for developers coming from web
- Enables composable, multi-screen TUI applications
- Solves state management for screen transitions

**Why SIMPLIFICATION NEEDED:**
- Current spec is over-engineered for TUI use cases (9 major features, 67 requirements)
- TUI apps rarely need all web routing features (nested routes, hash fragments, etc.)
- Bubbletea already handles some concerns (command pattern, state updates)
- Risk: building web patterns for non-web medium

---

## 2. Current State Analysis

### What We've Built So Far (Task 1.1)
✅ **Route Pattern Compilation** - COMPLETED
- Static segments: `/users/list`
- Dynamic params: `/user/:id`
- Optional params: `/profile/:id?`
- Wildcards: `/docs/:path*`
- Regex-based matching
- 93.3% test coverage, zero lint warnings

### What's Planned Next (67 requirements total)

**Phase 1: Core Matching** (5 tasks, 15 hours)
- ✅ Task 1.1: Pattern Compilation (DONE)
- Task 1.2: Route Matching Algorithm
- Task 1.3: Route Registry
- Task 1.4: Parameter Extraction
- Task 1.5: 404 Handling

**Phase 2: Navigation** (4 tasks, 12 hours)
- Router API (Push, Replace, Back, Forward)
- Navigation commands for Bubbletea
- Named route navigation
- Query string support

**Phase 3: Guards** (4 tasks, 12 hours)
- Global guards
- Per-route guards
- Component guards
- Guard resolution chain

**Phase 4: History** (3 tasks, 9 hours)
- History stack
- Forward/back navigation
- State preservation

**Phase 5: Integration** (5 tasks, 15 hours)
- Component rendering
- Context injection
- Composables (useRouter, useRoute)
- Event system
- Reactive updates

**Phase 6: Advanced Features** (4 tasks, 12 hours)
- Nested routes
- Route meta fields
- Aliases
- Redirects

**Total Estimated**: 75 hours of development

---

## 3. TUI Pain Points Analysis

### Real Problems in TUI Development Today

#### Problem 1: No Standard Navigation Pattern
**Pain Point**: Complex TUI apps become unmaintainable spaghetti

```go
// BEFORE: Typical Bubbletea app with multiple screens
type model struct {
    screen       string  // "home", "settings", "user-detail", etc.
    userData     User
    settingsData Settings
    // ... dozens of screen-specific fields
    
    // Navigation state scattered everywhere
    previousScreen string
    screenHistory []string
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "u" {
            // How do we pass the user ID to the detail screen?
            m.screen = "user-detail"
            // Have to manually manage state
            m.previousScreen = "home"
        }
    }
    
    // Giant switch statement for every screen
    switch m.screen {
    case "home":
        return m.updateHome(msg)
    case "settings":
        return m.updateSettings(msg)
    case "user-detail":
        return m.updateUserDetail(msg)
    // ... 20 more cases
    }
}

func (m model) View() string {
    // Another giant switch
    switch m.screen {
    case "home":
        return m.viewHome()
    // ... repeat for every screen
    }
}
```

**Why This Hurts**:
- All screens share one giant model struct
- No clear separation of concerns
- Navigation logic mixed with business logic
- Hard to test individual screens
- Adding new screens requires touching multiple places

#### Problem 2: Passing Data Between Screens
**Pain Point**: No standard way to pass parameters

```go
// BEFORE: Messy workarounds
type model struct {
    currentUserID string  // For user detail screen
    currentPostID int     // For post detail screen
    returnToScreen string // To know where to go back
}

// Somewhere in code:
m.currentUserID = "123"
m.screen = "user-detail"

// In another place:
// Wait, what was the previous screen? Where should "back" go?
m.screen = m.returnToScreen  // Hope this is correct!
```

#### Problem 3: Complex Multi-Screen Flows
**Pain Point**: Features like wizards, authentication flows, onboarding

```go
// BEFORE: Authentication flow
type model struct {
    authStep int  // 0=login, 1=2fa, 2=success
    loginData LoginForm
    require2FA bool
}

// Navigation is brittle and error-prone
if authSuccess {
    if m.require2FA {
        m.authStep = 1
    } else {
        m.authStep = 2
        // But wait, where should we redirect after login?
        // Was user trying to access a protected resource?
        // No standard way to handle this!
    }
}
```

#### Problem 4: No Back/Forward History
**Pain Point**: Users expect to navigate back, but it's manual

```go
// BEFORE: Manual history management
type model struct {
    screenStack []string  // DIY history
}

func (m model) goBack() {
    if len(m.screenStack) > 1 {
        m.screenStack = m.screenStack[:len(m.screenStack)-1]
        m.screen = m.screenStack[len(m.screenStack)-1]
    }
}

// But what about the state of previous screens?
// All lost when we navigate away!
```

### What TUI Apps Actually Need

Based on analysis of popular TUI apps (lazygit, k9s, htop, gh, kubectl):

**Core Navigation Needs** (✅ = Router solves this):
1. ✅ **Screen switching** with clear paths
2. ✅ **Parameter passing** (e.g., show user ID 123)
3. ✅ **Back navigation** that works
4. ✅ **State isolation** between screens
5. ✅ **Auth guards** (protect admin screens)
6. ⚠️ **Nested layouts** (header/footer + content) - Partially needed
7. ❌ **Query strings** - Rarely needed in TUI
8. ❌ **Hash fragments** - Never needed in TUI
9. ❌ **Multiple route aliases** - Over-engineering

---

## 4. Router Feature Scope

### What Router System Provides (from specs)

#### Phase 1: Core Matching ✅ ESSENTIAL
- Pattern compilation (✅ DONE)
- Route matching
- Parameter extraction
- 404 handling

**Value**: Foundation for everything else

#### Phase 2: Navigation ✅ ESSENTIAL
- `router.Push("/user/:id", params)`
- `router.Back()`
- Named routes
- Query strings ⚠️

**Value**: Solves navigation pain points

#### Phase 3: Guards ✅ HIGHLY VALUABLE
- Global guards (auth checks)
- Per-route guards
- Component guards
- Navigation flow control

**Value**: Solves authentication, authorization, validation flows

#### Phase 4: History ✅ ESSENTIAL
- Back/forward navigation
- History stack
- State preservation

**Value**: Expected UX in multi-screen apps

#### Phase 5: Integration ✅ ESSENTIAL
- Component rendering
- Composables (`useRouter`, `useRoute`)
- Reactive updates

**Value**: Makes it usable in BubblyUI components

#### Phase 6: Advanced Features ⚠️ QUESTIONABLE
- Nested routes - Maybe for layouts?
- Route meta fields - Nice to have
- Aliases - Unnecessary complexity
- Redirects - Can be done in guards

**Value**: Low for typical TUI use cases

---

## 5. Use Case Comparison: Before/After

### Use Case 1: CLI Tool with Multiple Commands

**Example**: Git TUI (like lazygit)

#### BEFORE (No Router)
```go
type model struct {
    screen string // "status", "log", "branches", "commit-detail"
    
    // State for every screen mixed together
    commits []Commit
    branches []Branch
    selectedCommit string
    selectedBranch string
    commitDetailsData CommitDetail
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" && m.screen == "log" {
            // Navigate to commit detail
            m.selectedCommit = m.commits[m.cursor].SHA
            m.screen = "commit-detail"
            // Manually fetch commit details
            m.commitDetailsData = fetchCommitDetails(m.selectedCommit)
        }
    }
    
    // Handle updates for current screen
    switch m.screen {
    case "status": return m.updateStatus(msg)
    case "log": return m.updateLog(msg)
    case "branches": return m.updateBranches(msg)
    case "commit-detail": return m.updateCommitDetail(msg)
    }
}
```

**Problems**:
- 100+ line Update() function
- All screen state in one struct
- Manual data fetching per screen
- Hard to add new screens
- Can't reuse screen components

#### AFTER (With Router)
```go
// Define routes once
router := bubbly.NewRouter(bubbly.RouterConfig{
    Routes: []bubbly.Route{
        {Path: "/", Component: StatusScreen},
        {Path: "/log", Component: LogScreen},
        {Path: "/branches", Component: BranchesScreen},
        {Path: "/commit/:sha", Component: CommitDetailScreen},
    },
})

// Each screen is isolated component
func CommitDetailScreen(ctx bubbly.RenderContext) bubbly.Component {
    route := ctx.UseRoute()
    sha := route.Params["sha"]  // Type-safe parameter access
    
    // Load data for this screen only
    details := ctx.UseAsync(func() interface{} {
        return fetchCommitDetails(sha)
    })
    
    return bubbly.NewComponent(/* ... */)
}

// Navigation is clean
router.Push("/commit/" + selectedCommit.SHA)
router.Back()  // Just works!
```

**Benefits**:
- ✅ Each screen is independent component
- ✅ Type-safe parameter passing
- ✅ Clear navigation API
- ✅ Automatic back button
- ✅ Easy to add new screens

### Use Case 2: Dashboard with Protected Routes

**Example**: Kubernetes TUI (like k9s)

#### BEFORE (No Router)
```go
type model struct {
    authenticated bool
    userRole string
    screen string
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Auth logic scattered everywhere
    if msg.String() == "d" && m.screen == "pods" {
        // Want to delete pod
        if !m.authenticated {
            m.screen = "login"
            return m, nil
        }
        if m.userRole != "admin" {
            m.screen = "unauthorized"
            return m, nil
        }
        // Finally do the thing
        m.screen = "delete-confirm"
    }
}
```

#### AFTER (With Router)
```go
router := bubbly.NewRouter(bubbly.RouterConfig{
    Routes: []bubbly.Route{
        {Path: "/login", Component: LoginScreen},
        {
            Path: "/pods",
            Component: PodsScreen,
            BeforeEnter: RequireAuth,
        },
        {
            Path: "/pods/:id/delete",
            Component: DeleteConfirmScreen,
            BeforeEnter: RequireRole("admin"),
        },
    },
})

// Guard is reusable
func RequireAuth(to, from *bubbly.Route, next bubbly.GuardNext) {
    if !authService.IsAuthenticated() {
        next("/login")  // Redirect to login
        return
    }
    next()  // Continue navigation
}

// Navigation is simple, guards handle auth
router.Push("/pods/" + podID + "/delete")
// If not authenticated -> auto redirects to /login
// After login -> can redirect back to original destination
```

**Benefits**:
- ✅ Auth logic centralized in guards
- ✅ Automatic redirects
- ✅ Role-based access control
- ✅ Can return to original destination after login

### Use Case 3: Wizard/Multi-Step Flow

**Example**: Application onboarding or configuration wizard

#### BEFORE (No Router)
```go
type model struct {
    wizardStep int
    wizardData WizardData
}

func (m model) nextStep() {
    m.wizardStep++
    // What if step 3 is conditional based on step 1?
    if m.wizardData.Step1Choice == "advanced" {
        m.wizardStep = 5  // Skip to advanced config
    }
}
```

#### AFTER (With Router)
```go
router := bubbly.NewRouter(bubbly.RouterConfig{
    Routes: []bubbly.Route{
        {Path: "/setup/welcome", Component: WelcomeScreen},
        {Path: "/setup/basic", Component: BasicConfigScreen},
        {Path: "/setup/advanced", Component: AdvancedConfigScreen},
        {Path: "/setup/review", Component: ReviewScreen},
    },
})

// Conditional navigation
if setupData.Mode == "advanced" {
    router.Push("/setup/advanced")
} else {
    router.Push("/setup/basic")
}

// Back button works automatically
router.Back()  // Returns to previous step
```

**Benefits**:
- ✅ Clear flow structure
- ✅ Conditional routing is explicit
- ✅ Can go back to any previous step
- ✅ Each step is testable component

---

## 6. Critical Analysis

### What Works Well ✅

1. **Pattern Matching** (Task 1.1 - DONE)
   - Solves: Parameter extraction
   - Implementation: Clean, well-tested
   - Value: Essential foundation

2. **Core Router API** (Phase 2)
   - Solves: Navigation chaos
   - Fits TUI: Yes, familiar from web
   - Value: High

3. **Guards** (Phase 3)
   - Solves: Auth/authorization flows
   - Fits TUI: Yes, especially for admin tools
   - Value: High

4. **History** (Phase 4)
   - Solves: Back/forward navigation
   - Fits TUI: Yes, expected UX
   - Value: Essential

### What's Over-Engineered ⚠️

1. **Query Strings**
   - Web pattern: `/users?page=2&sort=name`
   - TUI reality: Rarely needed, can use route params
   - Example: `/users/2/sort-name` works fine
   - Verdict: OPTIONAL, low priority

2. **Hash Fragments**
   - Web pattern: `/docs#section-3` (for anchor links)
   - TUI reality: No concept of "anchors" in terminal
   - Verdict: REMOVE from spec

3. **Route Aliases**
   - Web pattern: Multiple paths → same component
   - TUI reality: Adds complexity, rarely needed
   - Verdict: REMOVE from spec

4. **Nested Routes** (Questionable)
   - Web pattern: Parent route with `<router-outlet>`
   - TUI reality: Could be useful for layouts (header + content)
   - Example: Admin layout wrapping multiple admin screens
   - Verdict: SIMPLIFIED version might be useful

5. **Multiple History Modes**
   - Web pattern: Hash mode, HTML5 mode, memory mode
   - TUI reality: Only memory mode makes sense
   - Verdict: SIMPLIFY to single mode

### Missing Features We Actually Need 🎯

1. **Breadcrumbs**
   - TUI apps often show: `Home > Users > User Detail`
   - Router should provide breadcrumb data
   - High value for complex apps

2. **Screen Titles**
   - TUI apps show current screen at top
   - Router should provide title from route meta
   - Easy win

3. **Keyboard Shortcuts Integration**
   - Common pattern: Press '1' → go to tab 1
   - Router could provide shortcut registry
   - Would be very TUI-specific

4. **State Persistence Per Route**
   - When navigating back, restore scroll position, selections
   - Higher value than web (terminal state is precious)
   - Not in current spec!

---

## 7. Recommendations

### Recommendation 1: Complete Core Features ✅

**Continue with current plan for**:
- ✅ Phase 1: Core Matching (Task 1.1 done, continue 1.2-1.5)
- ✅ Phase 2: Navigation (essential)
- ✅ Phase 3: Guards (high value)
- ✅ Phase 4: History (essential)
- ✅ Phase 5: Integration (required)

**Estimated effort**: 48 hours remaining (from 75 total)

### Recommendation 2: Simplify Phase 6 ⚠️

**Remove from scope**:
- ❌ Hash fragments (no TUI use case)
- ❌ Multiple route aliases (complexity > value)
- ❌ Complex nested route system (can do simpler layout pattern)
- ❌ Query string full spec (keep basic, deprioritize advanced)

**Add to scope**:
- ✅ Breadcrumb support
- ✅ Screen title management
- ✅ State persistence per route

**Estimated effort savings**: -12 hours, +6 hours = -6 hours net

### Recommendation 3: Create Example Apps 📝

Before completing all phases, build 2 example apps to validate:

**Example 1: Multi-screen CLI Tool**
- 5-6 screens
- Parameter passing
- Back navigation
- Estimated: 4 hours

**Example 2: Admin Dashboard with Auth**
- Login screen
- Protected admin area
- Role-based guards
- Estimated: 6 hours

**Why**: Will reveal if API is ergonomic before investing 40+ more hours

### Recommendation 4: Document TUI-Specific Patterns 📚

Create guide for:
- When to use router vs simple screen switching
- TUI-specific navigation patterns
- Keyboard shortcut integration
- How this differs from Vue Router

### Final Recommendation: PROCEED WITH MODIFICATIONS ✅

**YES, continue with router feature because**:
1. Solves real pain points (navigation structure, state management)
2. Task 1.1 proved feature is technically sound
3. Provides competitive advantage vs raw Bubbletea
4. Enables complex TUI applications

**BUT, make these changes**:
1. Simplify Phase 6 (remove over-engineered features)
2. Add TUI-specific features (breadcrumbs, state persistence)
3. Build validation examples after Phase 5 (before Phase 6)
4. Document TUI-specific usage patterns

**Expected timeline**:
- Core features (Phases 1-5): 48 hours
- Validation examples: 10 hours
- Simplified Phase 6: 6 hours
- **Total: 64 hours** (down from 75)

**Expected value**: HIGH
- Makes BubblyUI competitive with React/Vue for TUI
- Enables professional-grade CLI tools
- Solves pain points in existing TUI apps

---

## Appendix: Real-World TUI Apps That Would Benefit

### Would Benefit Greatly ✅
- **lazygit**: Multiple screens, complex navigation
- **k9s**: Protected resources, auth, many screens
- **gh dashboard**: Multiple views, need back button
- **gitui**: Similar to lazygit
- **kubectl TUI plugins**: Auth, navigation

### Would Benefit Somewhat ⚠️
- **htop**: Mostly single screen (settings dialog could use router)
- **ncdu**: Tree navigation (different pattern)

### Would Not Benefit ❌
- **top**: Single screen
- **cat/less**: Single purpose
- Simple CLI tools

---

## Conclusion

The Router feature is **valuable and should proceed**, but with the modifications outlined above. The pattern matching work (Task 1.1) is solid foundation. Continue with core features, build validation examples, then reassess before advanced features.

**Key Success Metrics**:
1. Can build lazygit-style app in 50% less code
2. Auth guards work elegantly
3. Back button "just works"
4. State preserved when navigating back
5. Easy to add new screens to existing app

**Risk Mitigation**:
- Build examples after Phase 5 to validate design
- Get community feedback before Phase 6
- Maintain ability to use raw Bubbletea alongside router
- Document when NOT to use router (simple apps)

---

**Status**: Ready for review and decision before continuing to Task 1.2
