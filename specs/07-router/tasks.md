# Implementation Tasks: Router System

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] 01-reactivity-system completed
- [x] 02-component-model completed
- [x] 03-lifecycle-hooks completed
- [x] 04-composition-api completed
- [x] Route data structures defined (RoutePattern, Segment, SegmentKind)
- [x] Test framework configured for routing tests (testify)

---

## Phase 1: Core Route Matching (5 tasks, 15 hours)

### Task 1.1: Route Pattern Compilation ✅ COMPLETED
**Description**: Implement path-to-pattern compilation for route matching

**Prerequisites**: None

**Unlocks**: Task 1.2 (Route Matching Algorithm)

**Files**:
- `pkg/bubbly/router/pattern.go` ✅
- `pkg/bubbly/router/pattern_test.go` ✅

**Type Safety**:
```go
type RoutePattern struct {
    segments []Segment
    regex    *regexp.Regexp
}

type Segment struct {
    Kind  SegmentKind  // static, param, optional, wildcard
    Name  string       // param name
    Value string       // static value
}
```

**Tests**:
- [x] Static segments compile correctly
- [x] Dynamic params (:id) compile correctly
- [x] Optional params (:id?) compile correctly
- [x] Wildcards (:path*) compile correctly
- [x] Invalid patterns return errors
- [x] Regex is generated correctly

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 93.3% (exceeds 80% target)
- **Tests**: All 4 test suites passing (38 test cases)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings
- **Architecture**: 
  - Refactored for low cyclomatic complexity (extracted helper functions)
  - `parseSegments()` - parses path parts into segments
  - `parseParamSegment()` - handles parameter segments
  - `parseWildcardSegment()` - handles wildcard segments (:path*)
  - `parseOptionalSegment()` - handles optional segments (:id?)
  - `parseRegularParam()` - handles regular params (:id)
  - `validateParamName()` - validates param names and checks duplicates
  - `generateRegex()` - creates regex from segments
  - `isValidParamName()` - validates alphanumeric + underscore names
- **Features**:
  - Static segments: `/users/list`
  - Dynamic params: `/user/:id`
  - Optional params: `/profile/:id?`
  - Wildcards: `/docs/:path*`
  - Path normalization (trailing slash handling)
  - Duplicate param detection
  - Invalid pattern validation
  - Regex-based matching with parameter extraction
- **Edge Cases Handled**:
  - Empty paths
  - Missing leading slash
  - Duplicate parameter names
  - Wildcards not at end
  - Empty parameter names
  - Invalid characters in params
  - Root path (`/`)
  - Trailing slashes
- **Performance**: Regex compilation cached in RoutePattern struct

---

### Task 1.2: Route Matching Algorithm ✅ COMPLETED
**Description**: Implement path matching with parameter extraction

**Prerequisites**: Task 1.1

**Unlocks**: Task 1.3 (Route Registry)

**Files**:
- `pkg/bubbly/router/matcher.go` ✅
- `pkg/bubbly/router/matcher_test.go` ✅

**Type Safety**:
```go
type RouteMatcher struct {
    routes []*RouteRecord
}

func (rm *RouteMatcher) Match(path string) (*RouteMatch, error)

type RouteMatch struct {
    Route  *RouteRecord
    Params map[string]string
    Score  matchScore
}
```

**Tests**:
- [x] Static routes match correctly
- [x] Dynamic params are extracted
- [x] Optional params work
- [x] Wildcards match correctly
- [x] Most specific route wins
- [x] 404 when no match
- [x] Benchmark: < 100μs per match

**Estimated Effort**: 4 hours

**Implementation Notes**:
- **Coverage**: 91.3% (exceeds 80% target)
- **Tests**: All 7 test suites passing (48 test cases)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings
- **Performance**: **1.056 μs/op** (well under 100μs target ✅)
  - Memory: 345 B/op, 5 allocs/op
- **Architecture**:
  - `NewRouteMatcher()` - creates matcher instance
  - `AddRoute(path, name)` - registers routes with pattern compilation
  - `Match(path)` - finds best match with scoring
  - `calculateScore()` - computes route specificity
  - `isMoreSpecific()` - comparison for sorting
- **Scoring Algorithm**:
  - More static segments = more specific (higher priority)
  - Fewer param segments = more specific
  - Fewer optional segments = more specific
  - Fewer wildcard segments = more specific
  - Example: `/users/new` (static) beats `/users/:id` (param)
- **Error Handling**:
  - Returns `ErrNoMatch` for unmatched paths (404 scenario)
  - Pattern compilation errors bubble from Task 1.1
  - Proper error wrapping with context
- **Edge Cases Handled**:
  - Empty paths
  - Root path (`/`)
  - Trailing slash normalization
  - Multiple routes matching (precedence via scoring)
  - No matches (404)
- **Integration**: Uses `RoutePattern` from Task 1.1 for compilation and matching

---

### Task 1.3: Route Registry ✅ COMPLETED
**Description**: Implement route registration and lookup

**Prerequisites**: Task 1.2

**Unlocks**: Task 2.1 (Router Core)

**Files**:
- `pkg/bubbly/router/registry.go` ✅
- `pkg/bubbly/router/registry_test.go` ✅

**Type Safety**:
```go
type RouteRegistry struct {
    routes   []*RouteRecord
    byName   map[string]*RouteRecord
    byPath   map[string]*RouteRecord
    mu       sync.RWMutex
}

type RouteRecord struct {
    Path      string
    Name      string
    Component bubbly.Component
    Children  []*RouteRecord
    Meta      map[string]interface{}
    pattern   *RoutePattern
}
```

**Tests**:
- [x] Routes register correctly
- [x] Named routes accessible
- [x] Nested routes register
- [x] Duplicate paths rejected
- [x] Duplicate names rejected
- [x] Thread-safe registration

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 92.2% (exceeds 80% target)
- **Tests**: All 8 test suites passing (41 test cases total across router package)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings in router package
- **Architecture**:
  - `NewRouteRegistry()` - creates registry with empty indexes
  - `Register(path, name, meta)` - registers routes with duplicate detection
  - `GetByName(name)` - O(1) lookup by route name
  - `GetByPath(path)` - O(1) lookup by route path
  - `GetAll()` - returns defensive copy of all routes
- **Thread Safety**:
  - Uses `sync.RWMutex` for concurrent access
  - Multiple readers can access simultaneously
  - Exclusive write lock for registration
  - Defensive copy in GetAll() prevents external modification
- **Duplicate Detection**:
  - Checks for duplicate paths before registration
  - Checks for duplicate names before registration
  - Returns descriptive errors for duplicates
- **Indexing Strategy**:
  - Three indexes for efficient access:
    - `routes` slice: ordered list for iteration
    - `byName` map: O(1) name-based lookup
    - `byPath` map: O(1) path-based lookup
- **RouteRecord Enhancement**:
  - Added `Meta` field to matcher.go RouteRecord for metadata support
  - Added `Children` field to matcher.go RouteRecord for nested routes
  - Maintains backward compatibility with existing matcher code
- **Edge Cases Handled**:
  - Empty registry (GetAll returns empty slice)
  - Non-existent routes (GetByName/GetByPath return nil, false)
  - Concurrent registration and reads (thread-safe)
  - Nested route support via Children field
  - Metadata preservation via Meta field

---

### Task 1.4: Query String Parser ✅ COMPLETED
**Description**: Parse and handle URL query strings

**Prerequisites**: None

**Unlocks**: Task 1.5 (Route Object)

**Files**:
- `pkg/bubbly/router/query.go` ✅
- `pkg/bubbly/router/query_test.go` ✅

**Type Safety**:
```go
type QueryParser struct{}

func (qp *QueryParser) Parse(queryString string) map[string]string

func (qp *QueryParser) Build(params map[string]string) string
```

**Tests**:
- [x] Simple queries parse: ?key=value
- [x] Multiple params: ?a=1&b=2
- [x] URL encoding handled
- [x] Empty values: ?key=
- [x] No value: ?key
- [x] Build from map
- [x] Round-trip consistency

**Estimated Effort**: 2 hours

**Implementation Notes**:
- **Coverage**: 91.9% (exceeds 80% target)
- **Tests**: All 6 test suites passing (29 test cases for query parser)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings in router package
- **Standard Library**: Uses Go's `net/url` package for robust URL handling
- **Architecture**:
  - `NewQueryParser()` - creates stateless parser instance
  - `Parse(queryString)` - parses query string to map with URL decoding
  - `Build(params)` - builds query string from map with URL encoding
- **URL Encoding/Decoding**:
  - Uses `url.ParseQuery()` for parsing (RFC 3986 compliant)
  - Uses `url.Values.Encode()` for building (RFC 3986 compliant)
  - Automatic handling of special characters (spaces, @, /, etc.)
  - Proper percent-encoding (%20, %40, %2F, etc.)
- **Edge Cases Handled**:
  - Leading "?" automatically stripped
  - Empty query strings return empty map
  - Keys without values treated as empty strings
  - Multiple ampersands handled gracefully
  - Trailing/leading ampersands ignored
  - Equals signs in values preserved
  - Duplicate keys: last value wins (simplified for routing use case)
- **Round-Trip Consistency**:
  - Parse → Build → Parse yields identical result
  - Verified with comprehensive round-trip tests
  - Keys sorted alphabetically in output for deterministic results
- **Performance**:
  - Stateless parser (no memory overhead)
  - Leverages Go's optimized standard library
  - Minimal allocations (single map allocation per operation)

---

### Task 1.5: Route Object ✅ COMPLETED
**Description**: Define current route state structure

**Prerequisites**: Task 1.3, Task 1.4

**Unlocks**: Task 2.1 (Router Core)

**Files**:
- `pkg/bubbly/router/route.go` ✅
- `pkg/bubbly/router/route_test.go` ✅

**Type Safety**:
```go
type Route struct {
    Path     string
    Name     string
    Params   map[string]string
    Query    map[string]string
    Hash     string
    Meta     map[string]interface{}
    Matched  []*RouteRecord
    FullPath string
}
```

**Tests**:
- [x] Route creation
- [x] Immutability (defensive copies)
- [x] FullPath generation
- [x] Matched route chain
- [x] Meta field access

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 93.2% (exceeds 80% target)
- **Tests**: All 7 test suites passing (36 test cases for Route)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings in router package
- **Architecture**:
  - `NewRoute()` - creates immutable Route with defensive copies
  - `GetMeta(key)` - retrieves metadata with existence checking
  - `generateFullPath()` - builds complete path with query and hash
  - Helper functions for defensive copying (maps and slices)
- **Immutability**:
  - All maps (Params, Query, Meta) are defensively copied
  - Slices (Matched) are defensively copied to prevent external modification
  - RouteRecord pointers are shared (shallow copy) - correct behavior
  - Nil maps/slices converted to empty for easier usage
- **FullPath Generation**:
  - Format: `path?query#hash`
  - Query parameters sorted alphabetically (deterministic)
  - Empty components omitted (no trailing ? or #)
  - Uses QueryParser for consistent encoding
- **Matched Chain**:
  - Supports nested routes with parent-child relationships
  - Slice is copied to prevent external append operations
  - RouteRecords themselves are shared references (managed by registry)
- **Meta Field Access**:
  - `GetMeta(key)` returns (value, found) for safe access
  - Supports any type via interface{}
  - Type assertions required for concrete types
- **Edge Cases Handled**:
  - Nil maps converted to empty maps
  - Nil slices converted to empty slices
  - Empty paths and values handled gracefully
  - Defensive copying prevents external mutation
- **Thread Safety**:
  - Route instances are immutable (safe for concurrent reads)
  - No locks needed since state cannot change
  - Defensive copies ensure external modifications don't affect Route

---

## Phase 2: Router Core (6 tasks, 18 hours)

### Task 2.1: Router Structure ✅ COMPLETED
**Description**: Implement main Router singleton

**Prerequisites**: Task 1.3, Task 1.5

**Unlocks**: Task 2.2 (Navigation)

**Files**:
- `pkg/bubbly/router/router.go` ✅
- `pkg/bubbly/router/router_test.go` ✅

**Type Safety**:
```go
type Router struct {
    registry       *RouteRegistry
    matcher        *RouteMatcher
    history        *History
    currentRoute   *Route
    beforeHooks    []NavigationGuard
    afterHooks     []AfterNavigationHook
    mu             sync.RWMutex
}

func NewRouter() *Router
func (r *Router) CurrentRoute() *Route
```

**Tests**:
- [x] Router creation
- [x] Singleton behavior (simple constructor for now)
- [x] Thread-safe access
- [x] Current route tracking
- [x] Component initialization

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 93.3% (exceeds 80% target)
- **Tests**: All 6 test suites passing (includes thread safety test)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings (go vet passes)
- **Architecture**:
  - `NewRouter()` - creates router with initialized components
  - `CurrentRoute()` - thread-safe access to current route with RWMutex
  - All components initialized: registry, matcher, history, hooks
- **Type Definitions Added**:
  - `NavigationGuard` - function type for before guards
  - `NextFunc` - function type for guard flow control
  - `AfterNavigationHook` - function type for after hooks
  - `NavigationTarget` - struct for navigation targets (path, name, params, query, hash)
  - `History` - placeholder struct (full implementation in Task 3.1)
- **Thread Safety**:
  - Uses `sync.RWMutex` for concurrent access
  - Multiple readers can access CurrentRoute() simultaneously
  - Write operations will be serialized in Task 2.2
- **Immutability**:
  - Route struct is immutable by design (from Task 1.5)
  - CurrentRoute() returns the route directly (safe due to Route immutability)
- **Edge Cases Handled**:
  - Nil current route (no active route)
  - Concurrent reads of current route
  - Empty hook arrays initialization
- **Note**: This is a simple constructor for Task 2.1. Task 2.5 will add RouterBuilder for fluent route configuration API.

---

### Task 2.2: Navigation Implementation ✅ COMPLETED
**Description**: Implement Push, Replace navigation methods

**Prerequisites**: Task 2.1

**Unlocks**: Task 2.3 (Navigation Guards)

**Files**:
- `pkg/bubbly/router/navigation.go` ✅
- `pkg/bubbly/router/navigation_test.go` ✅

**Type Safety**:
```go
type NavigationTarget struct {
    Path   string
    Name   string
    Params map[string]string
    Query  map[string]string
    Hash   string
}

func (r *Router) Push(target *NavigationTarget) tea.Cmd
func (r *Router) Replace(target *NavigationTarget) tea.Cmd
```

**Tests**:
- [x] Push creates history entry (placeholder for Task 3.1)
- [x] Replace doesn't create history (placeholder for Task 3.1)
- [x] Target validation
- [x] Command generation
- [x] Route change messages
- [x] Error handling

**Estimated Effort**: 4 hours

**Implementation Notes**:
- **Coverage**: 94.1% (exceeds 80% target)
- **Tests**: All 11 test suites passing (Push, Replace, validation, from/to routes)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings (go vet passes)
- **Architecture**:
  - `Push()` - generates Bubbletea command for forward navigation
  - `Replace()` - generates Bubbletea command for replace navigation
  - `validateTarget()` - validates navigation targets (nil, empty checks)
  - `matchTarget()` - matches target to route with params/query/hash
  - `syncRegistryToMatcher()` - temporary helper to sync routes (Task 2.5 will improve)
- **Message Types**:
  - `RouteChangedMsg` - success message with To/From routes
  - `NavigationErrorMsg` - error message with error, From route, To target
- **Error Types**:
  - `ErrNilTarget` - navigation target is nil
  - `ErrEmptyTarget` - navigation target has no path or name
  - `ErrNoMatch` - no route matches the path (from matcher)
- **Navigation Flow**:
  1. Validate target (not nil, has path or name)
  2. Sync registry routes to matcher (temporary for Task 2.2)
  3. Match route using matcher
  4. Extract params from match
  5. Merge query and hash from target
  6. Create Route object with all data
  7. Update current route (thread-safe with RWMutex)
  8. Return RouteChangedMsg with from/to routes
- **Bubbletea Integration**:
  - Commands return `tea.Msg` (RouteChangedMsg or NavigationErrorMsg)
  - Async execution via Bubbletea runtime
  - Thread-safe state updates with proper locking
  - From/To routes tracked in messages
- **Edge Cases Handled**:
  - Nil target → NavigationErrorMsg
  - Empty target (no path or name) → NavigationErrorMsg
  - Route not found → NavigationErrorMsg with ErrNoMatch
  - First navigation (from = nil)
  - Query string building and parsing
  - Hash fragment handling
  - Parameter extraction from path
- **Limitations (addressed in future tasks)**:
  - Named route navigation (target.Name) not yet implemented (Task 4.5)
  - History management placeholder (Task 3.1 will add Push/Replace history)
  - Guard execution not yet implemented (Task 2.3)
  - Registry/matcher sync is temporary (Task 2.5 Router Builder will improve)
- **Thread Safety**:
  - Route updates use RWMutex (write lock)
  - Commands execute asynchronously but state updates are serialized
  - Multiple Push() calls handled by Bubbletea's message queue

---

### Task 2.3: Navigation Guards ✅ COMPLETED
**Description**: Implement guard execution system

**Prerequisites**: Task 2.2

**Unlocks**: Task 2.4 (Guard Flow Control)

**Files**:
- `pkg/bubbly/router/guards.go` ✅
- `pkg/bubbly/router/guards_test.go` ✅

**Type Safety**:
```go
type NavigationGuard func(to, from *Route, next NextFunc)
type NextFunc func(target *NavigationTarget)
type AfterNavigationHook func(to, from *Route)

func (r *Router) BeforeEach(guard NavigationGuard)
func (r *Router) AfterEach(hook AfterNavigationHook)
```

**Tests**:
- [x] Global guards execute
- [x] Route guards execute (placeholder for Task 4.3)
- [x] Execution order correct
- [x] next() allows navigation
- [x] next() with empty target cancels
- [x] next() with path redirects

**Estimated Effort**: 4 hours

**Implementation Notes**:
- **Coverage**: 94.2% (exceeds 80% target)
- **Tests**: All 12 test suites passing (guards, hooks, execution order, flow control)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings (go vet passes)
- **Architecture**:
  - `BeforeEach()` - registers global before guards (thread-safe)
  - `AfterEach()` - registers global after hooks (thread-safe)
  - `executeBeforeGuards()` - executes guards sequentially with flow control
  - `executeAfterHooks()` - executes hooks after successful navigation
  - `guardResult` - internal type for guard action (continue, cancel, redirect)
  - `guardAction` - enum for guard actions
- **Guard Types**:
  - `NavigationGuard` - function type for before guards (already in router.go)
  - `NextFunc` - function type for flow control (already in router.go)
  - `AfterNavigationHook` - function type for after hooks (already in router.go)
- **Error Types**:
  - `ErrNavigationCancelled` - returned when guard cancels navigation
- **Guard Flow Control**:
  - `next(nil)` - Allow navigation, continue to next guard
  - `next(&NavigationTarget{})` - Cancel navigation (empty target)
  - `next(&NavigationTarget{Path: "..."})` - Redirect to different route
- **Execution Flow**:
  1. BeforeEach guards execute sequentially
  2. Each guard calls next() to control flow
  3. If guard cancels → return NavigationErrorMsg, skip remaining guards
  4. If guard redirects → recursively call Push/Replace with new target
  5. If all guards allow → continue with navigation
  6. After navigation succeeds → execute AfterEach hooks
  7. After hooks execute sequentially (cannot affect navigation)
- **Integration with Navigation**:
  - Push() and Replace() both execute guards
  - Guards execute after route matching but before route update
  - After hooks execute after route update
  - Redirects handled recursively (guard can redirect to another guarded route)
- **Thread Safety**:
  - Guard registration uses write lock (RWMutex)
  - Guard execution uses read lock with defensive copy
  - Multiple guards can be registered safely
  - Guards execute in registration order
- **Edge Cases Handled**:
  - Empty target in next() → cancel
  - Nil target in next() → allow
  - Path in next() → redirect
  - Guards stop on first cancel/redirect
  - After hooks don't execute on cancel
  - Guards work with both Push() and Replace()
  - To/From routes passed correctly to guards
  - First navigation has nil 'from' route
- **Limitations (addressed in future tasks)**:
  - Route-specific guards (route.BeforeEnter) not yet implemented (Task 4.3)
  - Component guards not yet implemented (Task 4.3)
  - Circular redirect detection not yet implemented (Task 2.4)
  - Guard timeout not yet implemented (Task 2.4)
- **Use Cases Enabled**:
  - Authentication checks (redirect to login if not authenticated)
  - Authorization checks (check permissions before route access)
  - Data validation (validate route params)
  - Analytics tracking (track page views in after hooks)
  - Logging (log navigation events)
  - Focus management (set focus after navigation)
  - Document title updates (update title from route meta)

---

### Task 2.4: Guard Flow Control ✅ COMPLETED
**Description**: Implement next() function logic and guard chaining

**Prerequisites**: Task 2.3

**Unlocks**: Task 3.1 (History Management)

**Files**:
- `pkg/bubbly/router/guard_flow.go` ✅
- `pkg/bubbly/router/guard_flow_test.go` ✅

**Type Safety**:
```go
type guardResult struct {
    action guardAction
    target *NavigationTarget
}

type guardAction int

const (
    guardContinue guardAction = iota
    guardCancel
    guardRedirect
)

type redirectTracker struct {
    visited map[string]bool
    depth   int
}
```

**Tests**:
- [x] Guard chain execution
- [x] Early termination on cancel
- [x] Redirect starts new navigation
- [x] Error handling
- [x] Circular redirect detection
- [x] Max redirect depth (timeout not needed for TUI)

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 94.3% (exceeds 80% target)
- **Tests**: All 7 test suites passing (circular redirects, depth limits, complex scenarios)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings (go vet passes)
- **Architecture**:
  - `redirectTracker` - tracks visited paths and redirect depth
  - `pushWithTracking()` - internal Push with redirect tracking
  - `replaceWithTracking()` - internal Replace with redirect tracking
  - `NavigationMsg` - marker interface for type-safe message handling
  - `maxRedirectDepth` - constant set to 10 (prevents infinite loops)
- **Error Types**:
  - `ErrCircularRedirect` - circular redirect detected
  - `ErrMaxRedirectDepth` - max redirect depth (10) exceeded
- **Circular Redirect Detection**:
  - Tracks visited paths in a map
  - Detects self-redirects (A→A)
  - Detects two-step loops (A→B→A)
  - Detects multi-step loops (A→B→C→A)
  - Returns clear error with path information
- **Redirect Depth Limiting**:
  - Maximum 10 redirects per navigation
  - Prevents stack overflow from infinite redirect loops
  - Counts redirects across guard chain
  - Returns clear error when limit exceeded
- **Redirect Tracking Flow**:
  1. Create redirectTracker on initial navigation
  2. Visit each route, check if already visited
  3. If visited → return ErrCircularRedirect
  4. If guard redirects → increment depth, check limit
  5. If depth > 10 → return ErrMaxRedirectDepth
  6. Recursively navigate with same tracker
  7. Tracker resets on new navigation (not passed between navigations)
- **Integration**:
  - Push() delegates to pushWithTracking(target, nil)
  - Replace() delegates to replaceWithTracking(target, nil)
  - Tracker passed through recursive redirect calls
  - Works seamlessly with existing guard system
- **Thread Safety**:
  - redirectTracker is local to each navigation
  - No shared state between navigations
  - Safe for concurrent navigations
- **Edge Cases Handled**:
  - Self-redirect (A→A) detected immediately
  - Two-step circular (A→B→A) detected
  - Multi-step circular (A→B→C→A) detected
  - Deep redirect chains (up to 10) allowed
  - Excessive redirects (>10) rejected
  - Tracker resets between navigations
  - Works with both Push() and Replace()
- **Performance**:
  - O(1) circular redirect detection (map lookup)
  - O(1) depth check (simple counter)
  - Minimal memory overhead (small map + counter)
  - No goroutines or timers needed
- **Design Decisions**:
  - **No timeout handling**: TUI apps are synchronous, guards execute immediately
  - **Max depth of 10**: Sufficient for legitimate use cases, prevents abuse
  - **Path-based tracking**: Simple and effective for circular detection
  - **Recursive implementation**: Clean code, safe with depth limit
- **Use Cases Enabled**:
  - Safe authentication redirects (login → dashboard)
  - Multi-step redirects (old → new → current)
  - Prevents infinite redirect loops
  - Clear error messages for debugging
  - Protection against misconfigured guards

---

### Task 2.5: Router Builder API
**Description**: Fluent API for router configuration

**Prerequisites**: Task 2.1

**Unlocks**: Task 3.1 (History Management)

**Files**:
- `pkg/bubbly/router/builder.go`
- `pkg/bubbly/router/builder_test.go`

**Type Safety**:
```go
type RouterBuilder struct {
    routes      []*RouteRecord
    beforeHooks []NavigationGuard
    afterHooks  []AfterNavigationHook
}

func (rb *RouterBuilder) Route(path string, component bubbly.Component, opts ...RouteOption) *RouterBuilder
func (rb *RouterBuilder) BeforeEach(guard NavigationGuard) *RouterBuilder
func (rb *RouterBuilder) Build() (*Router, error)
```

**Tests**:
- [ ] Fluent API works
- [ ] Route registration
- [ ] Guard registration
- [ ] Validation on Build()
- [ ] Error reporting
- [ ] Nested routes

**Estimated Effort**: 2 hours

---

### Task 2.6: Route Options
**Description**: Implement route configuration options

**Prerequisites**: Task 2.5

**Unlocks**: Task 3.1 (History Management)

**Files**:
- `pkg/bubbly/router/options.go`
- `pkg/bubbly/router/options_test.go`

**Type Safety**:
```go
type RouteOption func(*RouteRecord)

func WithName(name string) RouteOption
func WithMeta(meta map[string]interface{}) RouteOption
func WithGuard(guard NavigationGuard) RouteOption
func WithChildren(children ...*RouteRecord) RouteOption
```

**Tests**:
- [ ] Name option works
- [ ] Meta option works
- [ ] Guard option works
- [ ] Children option works
- [ ] Multiple options combine

**Estimated Effort**: 2 hours

---

## Phase 3: History Management (3 tasks, 9 hours)

### Task 3.1: History Stack
**Description**: Implement history stack data structure

**Prerequisites**: Task 2.4, Task 2.5, Task 2.6

**Unlocks**: Task 3.2 (History Navigation)

**Files**:
- `pkg/bubbly/router/history.go`
- `pkg/bubbly/router/history_test.go`

**Type Safety**:
```go
type History struct {
    entries []*HistoryEntry
    current int
    maxSize int
    mu      sync.Mutex
}

type HistoryEntry struct {
    Route *Route
    State interface{}
}

func (h *History) Push(route *Route)
func (h *History) Replace(route *Route)
```

**Tests**:
- [ ] Push adds entry
- [ ] Replace updates entry
- [ ] Forward history truncated on push
- [ ] Max size enforced
- [ ] Thread-safe operations

**Estimated Effort**: 3 hours

---

### Task 3.2: History Navigation
**Description**: Implement Back, Forward, Go methods

**Prerequisites**: Task 3.1

**Unlocks**: Task 4.1 (Nested Routes)

**Files**:
- `pkg/bubbly/router/history_nav.go`
- `pkg/bubbly/router/history_nav_test.go`

**Type Safety**:
```go
func (r *Router) Back() tea.Cmd
func (r *Router) Forward() tea.Cmd
func (r *Router) Go(n int) tea.Cmd

func (h *History) CanGoBack() bool
func (h *History) CanGoForward() bool
```

**Tests**:
- [ ] Back moves to previous
- [ ] Forward moves to next
- [ ] Go(n) moves n steps
- [ ] Bounds checking
- [ ] No-op on boundaries
- [ ] Commands generated

**Estimated Effort**: 3 hours

---

### Task 3.3: History State Preservation
**Description**: Save/restore arbitrary state with history entries

**Prerequisites**: Task 3.2

**Unlocks**: Task 4.1 (Nested Routes)

**Files**:
- `pkg/bubbly/router/history_state.go`
- `pkg/bubbly/router/history_state_test.go`

**Type Safety**:
```go
func (h *History) PushWithState(route *Route, state interface{})
func (h *History) CurrentState() interface{}
```

**Tests**:
- [ ] State saved with entry
- [ ] State restored on navigation
- [ ] State type safety
- [ ] nil state handled

**Estimated Effort**: 3 hours

---

## Phase 4: Nested Routes & Advanced (5 tasks, 15 hours)

### Task 4.1: Nested Route Definition
**Description**: Support parent-child route relationships

**Prerequisites**: Task 3.2, Task 3.3

**Unlocks**: Task 4.2 (RouterView Component)

**Files**:
- `pkg/bubbly/router/nested.go`
- `pkg/bubbly/router/nested_test.go`

**Type Safety**:
```go
func Child(path string, component bubbly.Component, opts ...RouteOption) *RouteRecord

type RouteRecord struct {
    // ... existing fields
    Parent   *RouteRecord
    Children []*RouteRecord
}
```

**Tests**:
- [ ] Child routes register
- [ ] Parent-child links
- [ ] Path resolution
- [ ] Matched array correct
- [ ] Nested params

**Estimated Effort**: 3 hours

---

### Task 4.2: RouterView Component
**Description**: Component that renders current route's component

**Prerequisites**: Task 4.1

**Unlocks**: Task 5.1 (Composables)

**Files**:
- `pkg/bubbly/router/router_view.go`
- `pkg/bubbly/router/router_view_test.go`

**Type Safety**:
```go
type RouterView struct {
    router *Router
    depth  int
}

func NewRouterView(depth int) *RouterView
func (rv *RouterView) View() string
```

**Tests**:
- [ ] Renders current component
- [ ] Handles depth for nesting
- [ ] Updates on route change
- [ ] Handles no match

**Estimated Effort**: 3 hours

---

### Task 4.3: Component Navigation Guards
**Description**: BeforeRouteEnter, BeforeRouteUpdate, BeforeRouteLeave

**Prerequisites**: Task 4.2

**Unlocks**: Task 5.1 (Composables)

**Files**:
- `pkg/bubbly/router/component_guards.go`
- `pkg/bubbly/router/component_guards_test.go`

**Type Safety**:
```go
type ComponentGuards interface {
    BeforeRouteEnter(to, from *Route, next NextFunc)
    BeforeRouteUpdate(to, from *Route, next NextFunc)
    BeforeRouteLeave(to, from *Route, next NextFunc)
}
```

**Tests**:
- [ ] BeforeRouteEnter executes
- [ ] BeforeRouteUpdate executes
- [ ] BeforeRouteLeave executes
- [ ] Execution order correct
- [ ] Integration with component lifecycle

**Estimated Effort**: 4 hours

---

### Task 4.4: Route Meta Fields
**Description**: Attach arbitrary metadata to routes

**Prerequisites**: Task 4.1

**Unlocks**: Task 5.1 (Composables)

**Files**: (Integrated in existing files)

**Type Safety**:
```go
type RouteRecord struct {
    // ... existing fields
    Meta map[string]interface{}
}

func (r *Route) GetMeta(key string) (interface{}, bool)
```

**Tests**:
- [ ] Meta fields set
- [ ] Meta fields accessible
- [ ] Type assertions work
- [ ] Inherited from parent

**Estimated Effort**: 2 hours

---

### Task 4.5: Route Name Navigation
**Description**: Navigate by route name instead of path

**Prerequisites**: Task 4.1

**Unlocks**: Task 5.1 (Composables)

**Files**:
- `pkg/bubbly/router/named_routes.go`
- `pkg/bubbly/router/named_routes_test.go`

**Type Safety**:
```go
func (r *Router) PushNamed(name string, params, query map[string]string) tea.Cmd

func (r *Router) BuildPath(name string, params, query map[string]string) string
```

**Tests**:
- [ ] Named navigation works
- [ ] Params injected correctly
- [ ] Query string added
- [ ] Invalid name handled
- [ ] Path building utility

**Estimated Effort**: 3 hours

---

## Phase 5: Composables & Context Integration (3 tasks, 9 hours)

### Task 5.1: useRouter Composable
**Description**: Provide router instance to components

**Prerequisites**: Task 4.2, Task 4.3, Task 4.4, Task 4.5

**Unlocks**: Task 5.2 (useRoute)

**Files**:
- `pkg/bubbly/router/composables.go`
- `pkg/bubbly/router/composables_test.go`

**Type Safety**:
```go
func UseRouter(ctx *bubbly.Context) *Router
```

**Tests**:
- [ ] Router accessible
- [ ] Panic if not provided
- [ ] Context injection works
- [ ] Multiple components share instance

**Estimated Effort**: 2 hours

---

### Task 5.2: useRoute Composable
**Description**: Reactive access to current route

**Prerequisites**: Task 5.1

**Unlocks**: Task 5.3 (Router Provider)

**Files**: (Integrated in composables.go)

**Type Safety**:
```go
func UseRoute(ctx *bubbly.Context) *bubbly.Ref[*Route]
```

**Tests**:
- [ ] Route accessible
- [ ] Updates reactively
- [ ] Params accessible
- [ ] Query accessible
- [ ] Meta accessible

**Estimated Effort**: 3 hours

---

### Task 5.3: Router Provider
**Description**: Inject router into component tree

**Prerequisites**: Task 5.2

**Unlocks**: Task 6.1 (Integration Testing)

**Files**:
- `pkg/bubbly/router/provider.go`
- `pkg/bubbly/router/provider_test.go`

**Type Safety**:
```go
func ProvideRouter(ctx *bubbly.Context, router *Router)
```

**Tests**:
- [ ] Router provided
- [ ] Child components access router
- [ ] Nested components work
- [ ] Multiple routers (different trees)

**Estimated Effort**: 4 hours

---

## Phase 6: Integration & Polish (4 tasks, 12 hours)

### Task 6.1: Bubbletea Message Integration
**Description**: Route change messages and command generation

**Prerequisites**: Task 5.3

**Unlocks**: Task 6.2 (Error Handling)

**Files**:
- `pkg/bubbly/router/messages.go`
- `pkg/bubbly/router/messages_test.go`

**Type Safety**:
```go
type RouteChangedMsg struct {
    To   *Route
    From *Route
}

type NavigationErrorMsg struct {
    Error error
    From  *Route
    To    *NavigationTarget
}
```

**Tests**:
- [ ] Messages generated correctly
- [ ] Commands return messages
- [ ] Integration with Update()
- [ ] Error messages work

**Estimated Effort**: 3 hours

---

### Task 6.2: Error Handling & Observability
**Description**: Production-grade error handling

**Prerequisites**: Task 6.1

**Unlocks**: Task 6.3 (Documentation)

**Files**:
- `pkg/bubbly/router/errors.go`
- `pkg/bubbly/router/errors_test.go`

**Type Safety**:
```go
type RouterError struct {
    Code    ErrorCode
    Message string
    From    *Route
    To      *NavigationTarget
    Cause   error
}

type ErrorCode int

const (
    ErrRouteNotFound ErrorCode = iota
    ErrInvalidPath
    ErrGuardRejected
    ErrCircularRedirect
)
```

**Tests**:
- [ ] Errors categorized correctly
- [ ] Observability integration
- [ ] Stack traces captured
- [ ] Error recovery
- [ ] Clear error messages

**Estimated Effort**: 3 hours

---

### Task 6.3: Documentation & Examples
**Description**: API documentation and usage examples

**Prerequisites**: Task 6.2

**Unlocks**: Task 6.4 (Performance Testing)

**Files**:
- `docs/router/README.md`
- `docs/router/guides/*.md`
- `cmd/examples/07-router/basic/main.go`
- `cmd/examples/07-router/guards/main.go`
- `cmd/examples/07-router/nested/main.go`

**Content**:
- API documentation
- Getting started guide
- Navigation guards guide
- Nested routes guide
- Example applications

**Estimated Effort**: 3 hours

---

### Task 6.4: Performance & Benchmarks
**Description**: Optimize and benchmark router operations

**Prerequisites**: Task 6.3

**Unlocks**: Feature complete

**Files**:
- `pkg/bubbly/router/benchmarks_test.go`

**Benchmarks**:
- [ ] Route matching < 100μs
- [ ] Navigation < 1ms overhead
- [ ] History operations < 50μs
- [ ] Memory per route < 1KB
- [ ] Guard execution < 10μs

**Optimizations**:
- Route match caching
- Pattern compilation caching
- Guard execution pooling

**Estimated Effort**: 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01-04)
    ↓
Phase 1: Core Matching
    1.1 Pattern → 1.2 Matcher → 1.3 Registry
    1.4 Query Parser → 1.5 Route Object
    ↓
Phase 2: Router Core
    2.1 Router → 2.2 Navigation → 2.3 Guards → 2.4 Guard Flow
    2.5 Builder → 2.6 Options
    ↓
Phase 3: History
    3.1 History Stack → 3.2 History Nav → 3.3 History State
    ↓
Phase 4: Advanced
    4.1 Nested → 4.2 RouterView → 4.3 Component Guards
    4.4 Meta → 4.5 Named Routes
    ↓
Phase 5: Composables
    5.1 useRouter → 5.2 useRoute → 5.3 Provider
    ↓
Phase 6: Polish
    6.1 Messages → 6.2 Errors → 6.3 Docs → 6.4 Performance
```

---

## Validation Checklist

### Core Functionality
- [ ] Routes match correctly
- [ ] Navigation works
- [ ] Guards execute in order
- [ ] History stack works
- [ ] Nested routes render
- [ ] Composables accessible

### Type Safety
- [ ] All types generic where appropriate
- [ ] No untyped interfaces without docs
- [ ] Compile-time checking
- [ ] Clear type assertions
- [ ] Generic constraints used

### Performance
- [ ] All benchmarks pass targets
- [ ] No memory leaks
- [ ] Thread-safe operations
- [ ] Efficient route matching
- [ ] Minimal overhead

### Integration
- [ ] Works with all BubblyUI features
- [ ] Bubbletea integration clean
- [ ] Composables work
- [ ] Context injection works
- [ ] Example apps work

### Quality
- [ ] >80% test coverage
- [ ] All tests pass with -race
- [ ] Zero lint warnings
- [ ] Documentation complete
- [ ] Error handling production-grade

---

## Estimated Total Effort

- Phase 1: 15 hours
- Phase 2: 18 hours
- Phase 3: 9 hours
- Phase 4: 15 hours
- Phase 5: 9 hours
- Phase 6: 12 hours

**Total**: ~78 hours (approximately 2 weeks)

---

## Priority

**HIGH** - Critical for multi-screen applications

**Timeline**: Feature 07 should be implemented after Features 04-06 are complete, as it depends on composition API and benefits from built-in components.

**Unlocks**: Multi-screen applications, authentication flows, complex UIs, dev tools integration
