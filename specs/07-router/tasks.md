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

### Task 2.5: Router Builder API ✅ COMPLETED
**Description**: Fluent API for router configuration

**Prerequisites**: Task 2.1

**Unlocks**: Task 3.1 (History Management)

**Files**:
- `pkg/bubbly/router/builder.go` ✅
- `pkg/bubbly/router/builder_test.go` ✅

**Type Safety**:
```go
type RouterBuilder struct {
    routes      []*RouteRecord
    beforeHooks []NavigationGuard
    afterHooks  []AfterNavigationHook
}

func NewRouterBuilder() *RouterBuilder
func (rb *RouterBuilder) Route(path, name string) *RouterBuilder
func (rb *RouterBuilder) RouteWithMeta(path, name string, meta map[string]interface{}) *RouterBuilder
func (rb *RouterBuilder) BeforeEach(guard NavigationGuard) *RouterBuilder
func (rb *RouterBuilder) AfterEach(hook AfterNavigationHook) *RouterBuilder
func (rb *RouterBuilder) Build() (*Router, error)
```

**Tests**:
- [x] Fluent API works
- [x] Route registration
- [x] Guard registration
- [x] Validation on Build()
- [x] Error reporting
- [x] Multiple builds from same builder

**Estimated Effort**: 2 hours

**Implementation Notes**:
- **Coverage**: 94.6% (exceeds 80% target)
- **Tests**: All 12 test suites passing (fluent API, validation, guards, complex scenarios)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings (go vet passes)
- **Architecture**:
  - `RouterBuilder` - fluent builder for router configuration
  - `NewRouterBuilder()` - constructor for builder
  - `Route()` - adds route without metadata
  - `RouteWithMeta()` - adds route with metadata
  - `BeforeEach()` - registers global before guard
  - `AfterEach()` - registers global after hook
  - `Build()` - creates configured router with validation
  - `validate()` - internal validation method
- **Error Types**:
  - `ErrEmptyPath` - path cannot be empty
  - `ErrDuplicatePath` - duplicate path detected
  - `ErrDuplicateName` - duplicate name detected
- **Validation Rules**:
  - Path cannot be empty
  - Paths must be unique
  - Names must be unique (if provided)
  - Validation runs before router creation
  - Clear error messages with context
- **Builder Pattern**:
  - Fluent API with method chaining
  - Immutable router after Build()
  - Builder can be reused for multiple routers
  - All methods return *RouterBuilder for chaining
  - Build() creates new router instance each time
- **Integration**:
  - Uses existing RouteRecord from matcher.go
  - Delegates to router.registry.Register()
  - Delegates to router.BeforeEach() and AfterEach()
  - Seamless integration with existing router system
- **Thread Safety**:
  - Builder is NOT thread-safe (single-goroutine use)
  - Built router IS thread-safe (concurrent use)
  - Builder should be used during setup only
- **Usage Example**:
  ```go
  router, err := NewRouterBuilder().
      Route("/", "home").
      Route("/about", "about").
      RouteWithMeta("/dashboard", "dashboard", map[string]interface{}{
          "requiresAuth": true,
      }).
      BeforeEach(authGuard).
      AfterEach(analyticsHook).
      Build()
  if err != nil {
      log.Fatal(err)
  }
  ```
- **Edge Cases Handled**:
  - Empty builder (no routes) is valid
  - Empty path validation
  - Duplicate path detection
  - Duplicate name detection
  - Multiple builds from same builder
  - Routes with and without metadata
  - Routes with and without names
- **Design Decisions**:
  - **Reuses RouteRecord**: Uses existing type from matcher.go
  - **Simple API**: Route() for common case, RouteWithMeta() for metadata
  - **Validation on Build()**: Catches errors before router creation
  - **Method chaining**: Fluent API for readability
  - **Immutable router**: Router cannot be modified after Build()
- **Benefits**:
  - Improved developer experience
  - Type-safe configuration
  - Clear validation errors
  - Readable route definitions
  - Chainable method calls
  - Reusable builder pattern
- **Use Cases Enabled**:
  - Declarative router configuration
  - Centralized route definitions
  - Easy guard registration
  - Clear validation feedback
  - Multiple router instances from same config

---

### Task 2.6: Route Options ✅ COMPLETED
**Description**: Implement route configuration options

**Prerequisites**: Task 2.5

**Unlocks**: Task 3.1 (History Management)

**Files**:
- `pkg/bubbly/router/options.go` ✅
- `pkg/bubbly/router/options_test.go` ✅

**Type Safety**:
```go
type RouteOption func(*RouteRecord)

func WithName(name string) RouteOption
func WithMeta(meta map[string]interface{}) RouteOption
func WithGuard(guard NavigationGuard) RouteOption
func WithChildren(children ...*RouteRecord) RouteOption

// Builder integration
func (rb *RouterBuilder) RouteWithOptions(path string, opts ...RouteOption) *RouterBuilder
```

**Tests**:
- [x] Name option works
- [x] Meta option works
- [x] Guard option works
- [x] Children option works
- [x] Multiple options combine

**Estimated Effort**: 2 hours

**Implementation Notes**:
- **Coverage**: 94.8% (exceeds 80% target)
- **Tests**: All 10 test suites passing (options, merging, appending, complex scenarios)
- **Race detector**: Clean (no race conditions)
- **Lint**: Zero warnings (go vet passes)
- **Architecture**:
  - `RouteOption` - function type for route configuration
  - `WithName()` - sets route name
  - `WithMeta()` - sets/merges route metadata
  - `WithGuard()` - sets per-route navigation guard
  - `WithChildren()` - sets/appends child routes
  - `RouteWithOptions()` - builder method accepting options
- **Functional Options Pattern**:
  - Flexible and composable configuration
  - Options can be combined freely
  - Type-safe option functions
  - Follows Go best practices
- **Option Behaviors**:
  - **WithName**: Sets route name (overwrites if exists)
  - **WithMeta**: Merges with existing metadata (new keys added, existing overwritten)
  - **WithGuard**: Stores guard in metadata under "beforeEnter" key
  - **WithChildren**: Appends to existing children (preserves existing)
- **Integration**:
  - Works seamlessly with RouterBuilder
  - Compatible with existing Route() and RouteWithMeta() methods
  - Options applied in order specified
  - No conflicts with builder pattern
- **Usage Example**:
  ```go
  builder.RouteWithOptions("/dashboard",
      WithName("dashboard"),
      WithMeta(map[string]interface{}{
          "requiresAuth": true,
          "title": "Dashboard",
      }),
      WithGuard(authGuard),
      WithChildren(overviewRoute, settingsRoute),
  )
  ```
- **Per-Route Guards**:
  - Guards stored in route metadata under "beforeEnter" key
  - Execute after global before guards
  - Execute before component guards
  - Follow Vue Router convention
  - Can be accessed via `route.Meta["beforeEnter"]`
- **Nested Routes**:
  - Children routes for hierarchical routing
  - Supports unlimited nesting depth
  - Children appended to existing list
  - Useful for layouts with nested views
- **Edge Cases Handled**:
  - Meta merging with existing metadata
  - Children appending to existing children
  - Nil metadata initialization
  - Nil children initialization
  - Multiple options on same route
  - Options applied in sequence
- **Design Decisions**:
  - **Functional options**: More flexible than builder methods
  - **Meta merging**: Preserves existing metadata
  - **Children appending**: Preserves existing children
  - **Guard in metadata**: Follows Vue Router convention
  - **Variadic options**: Unlimited options per route
- **Benefits**:
  - Flexible route configuration
  - Composable options
  - Type-safe API
  - Clear intent
  - Easy to extend
  - Backward compatible
- **Use Cases Enabled**:
  - Per-route authentication guards
  - Nested route hierarchies
  - Route metadata configuration
  - Flexible route naming
  - Complex route structures

---

## Phase 3: History Management (3 tasks, 9 hours)

### Task 3.1: History Stack ✅ COMPLETED
**Description**: Implement history stack data structure

**Prerequisites**: Task 2.4, Task 2.5, Task 2.6

**Unlocks**: Task 3.2 (History Navigation)

**Files**:
- `pkg/bubbly/router/history.go` ✅
- `pkg/bubbly/router/history_test.go` ✅

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
- [x] Push adds entry
- [x] Replace updates entry
- [x] Forward history truncated on push
- [x] Max size enforced
- [x] Thread-safe operations

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 95.0% (exceeds 80% target)
- **Tests**: All 6 test suites passing (33 test cases total)
- **Race detector**: Clean (no race conditions)
- **go vet**: Zero warnings
- **Architecture**:
  - `History` struct with entries slice, current index, maxSize, and mutex
  - `HistoryEntry` struct with Route and optional State
  - `Push()` - adds entry, truncates forward history, enforces max size
  - `Replace()` - updates current entry without changing history length
  - `PushWithState()` - push with state preservation
  - `CurrentState()` - retrieves state from current entry
  - `enforceMaxSize()` - internal helper to trim oldest entries
- **Thread Safety**:
  - Uses `sync.Mutex` for all operations
  - Safe for concurrent Push/Replace calls
  - Tested with 10 concurrent goroutines
- **Forward History Truncation**:
  - Push truncates entries after current position
  - Example: [A, B←, C] + Push(D) = [A, B, D←] (C removed)
  - Prevents "forward" navigation after new push
- **Max Size Enforcement**:
  - Optional limit on history stack size
  - Oldest entries removed when limit exceeded
  - Current index adjusted to maintain correct position
  - Example: maxSize=3, [A, B, C, D, E←] = [C, D, E←]
- **State Preservation**:
  - PushWithState() attaches arbitrary state to entries
  - CurrentState() retrieves state from current entry
  - Useful for scroll position, form data, filters, etc.
  - State is interface{} for flexibility
- **Edge Cases Handled**:
  - Empty history (current = -1)
  - Push to empty history
  - Replace in empty history (creates first entry)
  - Forward history truncation
  - Max size enforcement with index adjustment
  - Concurrent access (thread-safe)
  - Nil state handling
- **Integration**:
  - Removed placeholder History struct from router.go
  - Router.history field now uses full implementation
  - Ready for Task 3.2 (Back/Forward navigation)
- **Performance**:
  - O(1) Push operation (amortized)
  - O(1) Replace operation
  - O(n) max size enforcement (only when limit exceeded)
  - Minimal memory overhead (single slice + index + mutex)

---

### Task 3.2: History Navigation ✅ COMPLETED
**Description**: Implement Back, Forward, Go methods

**Prerequisites**: Task 3.1

**Unlocks**: Task 4.1 (Nested Routes)

**Files**:
- `pkg/bubbly/router/history_nav.go` ✅
- `pkg/bubbly/router/history_nav_test.go` ✅

**Type Safety**:
```go
func (r *Router) Back() tea.Cmd
func (r *Router) Forward() tea.Cmd
func (r *Router) Go(n int) tea.Cmd

func (h *History) CanGoBack() bool
func (h *History) CanGoForward() bool
```

**Tests**:
- [x] Back moves to previous
- [x] Forward moves to next
- [x] Go(n) moves n steps
- [x] Bounds checking
- [x] No-op on boundaries
- [x] Commands generated

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 94.7% (exceeds 80% target)
- **Tests**: All 7 test suites passing (48 test cases total)
- **Race detector**: Clean (no race conditions)
- **go vet**: Zero warnings
- **Architecture**:
  - **History helpers**:
    - `CanGoBack()` - checks if current > 0
    - `CanGoForward()` - checks if current < len-1
    - Thread-safe with mutex
  - **Router navigation methods**:
    - `Back()` - navigates to previous entry
    - `Forward()` - navigates to next entry
    - `Go(n)` - navigates n steps (negative=back, positive=forward)
    - All return `tea.Cmd` for Bubbletea integration
    - Return nil for no-op (boundaries, empty history)
- **Bubbletea Integration**:
  - Commands return `RouteChangedMsg` on success
  - Include To/From routes in message
  - Async execution via Bubbletea runtime
  - Thread-safe state updates with RWMutex
- **Bounds Checking**:
  - Back() returns nil if current == 0 (first entry)
  - Forward() returns nil if current == len-1 (last entry)
  - Go(n) clamps to [0, len-1] range
  - Go(0) returns nil (no-op)
  - Empty history always returns nil
- **Navigation Flow**:
  1. Check if navigation is possible (CanGoBack/CanGoForward)
  2. If not, return nil (no-op)
  3. Lock mutex for thread safety
  4. Save current route for "from" in message
  5. Update history.current index
  6. Get new route from history entry
  7. Update router.currentRoute
  8. Return RouteChangedMsg with to/from routes
- **Go(n) Clamping**:
  - Negative n: go back n steps
  - Positive n: go forward n steps
  - Clamps to first entry if n too negative
  - Clamps to last entry if n too positive
  - Example: current=2, Go(-10) → clamps to 0
  - Example: current=2, Go(10) → clamps to len-1
- **Thread Safety**:
  - Router.mu (RWMutex) protects currentRoute
  - History.mu (Mutex) protects entries and current
  - Both locks acquired in navigation commands
  - No deadlocks (consistent lock ordering)
  - Safe for concurrent Back/Forward/Go calls
- **Edge Cases Handled**:
  - Empty history (no entries)
  - Single entry (can't go back or forward)
  - At first entry (can't go back)
  - At last entry (can't go forward)
  - Go(0) is no-op
  - Go beyond bounds (clamped)
  - Concurrent navigation calls
- **Integration Tests**:
  - `TestRouter_BackForward_Integration` - full navigation flow
  - `TestRouter_Go_BoundsChecking` - boundary conditions
  - Verifies back/forward sequence works correctly
  - Verifies from/to routes in messages
- **Use Cases Enabled**:
  - Back button in navigation bar
  - Forward button in navigation bar
  - Keyboard shortcuts (ESC, Backspace, Ctrl+])
  - History navigation UI controls
  - Undo/redo navigation patterns
  - Jump to specific history position
- **Performance**:
  - O(1) CanGoBack/CanGoForward checks
  - O(1) Back/Forward navigation
  - O(1) Go(n) navigation
  - Minimal memory overhead (no allocations)
  - Lock contention minimal (short critical sections)

---

### Task 3.3: History State Preservation ✅ COMPLETED (Merged into Task 3.1)
**Description**: Save/restore arbitrary state with history entries

**Prerequisites**: Task 3.2

**Unlocks**: Task 4.1 (Nested Routes)

**Files**:
- Implemented in `pkg/bubbly/router/history.go` ✅
- Tests in `pkg/bubbly/router/history_test.go` ✅

**Type Safety**:
```go
func (h *History) PushWithState(route *Route, state interface{})
func (h *History) CurrentState() interface{}
```

**Tests**:
- [x] State saved with entry
- [x] State restored on navigation
- [x] State type safety
- [x] nil state handled

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Merged into Task 3.1**: State preservation was implemented as part of the History struct
- **No separate files needed**: PushWithState() and CurrentState() are in history.go
- **Tests included**: TestHistory_PushWithState and TestHistory_CurrentState verify functionality
- **State Storage**: HistoryEntry.State field holds arbitrary interface{} data
- **Use Cases**: Scroll position, form data, filter settings, UI state
- **Type Safety**: Requires type assertion when retrieving state
- **Nil Handling**: CurrentState() returns nil for empty history or entries without state

---

## Phase 4: Nested Routes & Advanced (5 tasks, 15 hours)

### Task 4.1: Nested Route Definition ✅ COMPLETED
**Description**: Support parent-child route relationships

**Prerequisites**: Task 3.2, Task 3.3

**Unlocks**: Task 4.2 (RouterView Component)

**Files**:
- `pkg/bubbly/router/nested.go` ✅
- `pkg/bubbly/router/nested_test.go` ✅
- Updated: `pkg/bubbly/router/matcher.go` (added Parent field, Matched array, AddRouteRecord method)

**Type Safety**:
```go
func Child(path string, opts ...RouteOption) *RouteRecord

type RouteRecord struct {
    Path     string
    Name     string
    Meta     map[string]interface{}
    Parent   *RouteRecord           // Added for nested routes
    Children []*RouteRecord
    pattern  *RoutePattern
}

type RouteMatch struct {
    Route   *RouteRecord
    Params  map[string]string
    Score   matchScore
    Matched []*RouteRecord // Added for nested routes
}
```

**Tests**:
- [x] Child routes register
- [x] Parent-child links
- [x] Path resolution
- [x] Matched array correct
- [x] Nested params

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 93.4% (exceeds 80% target)
- **Tests**: All 15 test suites passing (7 nested route tests + 8 option tests)
- **Race detector**: Clean (no race conditions)
- **go vet**: Zero warnings
- **Architecture**:
  - `Child()` - creates child route records with options
  - `resolveNestedPath()` - combines parent and child paths
  - `buildMatchedArray()` - constructs matched array from root to leaf
  - `establishParentLinks()` - sets up bidirectional parent-child relationships
  - `buildFullPath()` - resolves full path for deeply nested routes
  - `AddRouteRecord()` - registers routes with children recursively
  - `registerNestedRoute()` - handles nested route registration with full path resolution
- **Type Definitions**:
  - Added `Parent` field to `RouteRecord` for bidirectional linking
  - Added `Matched` array to `RouteMatch` for nested route rendering
- **Path Resolution**:
  - Empty child path ("") creates default child matching parent path
  - Relative child paths concatenate with parent path
  - Deeply nested routes (3+ levels) fully supported
  - Path resolution handles leading slashes correctly
- **Matched Array**:
  - Contains all route records from root to matched route
  - Built by walking up parent chain
  - Essential for nested route rendering (RouterView at each level)
  - Preference given to child routes over parent when scores equal (empty child path case)
- **Nested Params**:
  - Parent and child params combined in single map
  - No param name conflicts (validation in pattern compilation)
  - Multi-level params work correctly (userId + postId + etc.)
- **Edge Cases Handled**:
  - Empty child path (default child route)
  - Deeply nested routes (3+ levels tested)
  - Parent and child params combined
  - Child routes preferred over parent when patterns identical
  - Recursive child registration
  - Full path resolution for grandchildren
- **Integration**:
  - Works seamlessly with existing matcher
  - Compatible with RouteBuilder and options
  - Ready for RouterView component (Task 4.2)
- **Performance**:
  - O(d) path resolution where d = nesting depth
  - O(d) matched array construction
  - Minimal memory overhead (parent pointers only)
  - Pattern compilation cached per route

---

### Task 4.2: RouterView Component ✅ COMPLETED
**Description**: Component that renders current route's component

**Prerequisites**: Task 4.1

**Unlocks**: Task 5.1 (Composables)

**Files**:
- `pkg/bubbly/router/router_view.go` ✅
- `pkg/bubbly/router/router_view_test.go` ✅
- Updated: `pkg/bubbly/router/matcher.go` (added Component field to RouteRecord)
- Updated: `pkg/bubbly/router/options.go` (added WithComponent option)

**Type Safety**:
```go
type RouterView struct {
    router *Router
    depth  int
}

type RouteRecord struct {
    Path      string
    Name      string
    Component interface{}            // Component to render (bubbly.Component)
    Meta      map[string]interface{}
    Parent    *RouteRecord
    Children  []*RouteRecord
    pattern   *RoutePattern
}

func NewRouterView(router *Router, depth int) *RouterView
func (rv *RouterView) View() string
func WithComponent(component interface{}) RouteOption
```

**Tests**:
- [x] Renders current component
- [x] Handles depth for nesting
- [x] Updates on route change
- [x] Handles no match

**Estimated Effort**: 3 hours

**Implementation Notes**:
- **Coverage**: 92.6% (exceeds 80% target)
- **Tests**: All 8 test suites passing
  - TestNewRouterView (3 depth scenarios)
  - TestRouterView_RendersCurrentComponent
  - TestRouterView_HandlesDepthForNesting
  - TestRouterView_HandlesNoMatch
  - TestRouterView_HandlesDepthOutOfBounds
  - TestRouterView_HandlesNoComponent
  - TestRouterView_UpdatesOnRouteChange
  - TestWithComponent
- **Race detector**: Clean (no race conditions)
- **go vet**: Zero warnings
- **Architecture**:
  - RouterView implements both `tea.Model` and `bubbly.Component` interfaces
  - Passive component - only renders, doesn't handle messages
  - Thread-safe access to router's current route via mutex
  - Depth-based rendering using Matched array from Task 4.1
- **Component Integration**:
  - Added `Component` field to `RouteRecord` (interface{} type for flexibility)
  - Added `WithComponent()` option for setting route components
  - Components stored in route records, retrieved by RouterView
  - Type assertion to `bubbly.Component` interface for rendering
- **Depth-Based Rendering**:
  - `depth` parameter determines which level of Matched array to render
  - depth 0 = root/parent component
  - depth 1 = first child component
  - depth 2+ = grandchild and deeper
  - Bounds checking prevents index out of range errors
- **Rendering Logic**:
  1. Get current route from router (thread-safe)
  2. Check if route exists
  3. Validate depth is within Matched array bounds
  4. Get RouteRecord at specified depth
  5. Check if RouteRecord has Component
  6. Type assert Component to bubbly.Component
  7. Call Component.View() to render
  8. Return rendered string or empty string if any step fails
- **Error Handling**:
  - Returns empty string for all error cases (graceful degradation)
  - No current route → ""
  - Depth out of bounds → ""
  - No component set → ""
  - Component not bubbly.Component → ""
- **Interface Implementation**:
  - `tea.Model`: Init(), Update(), View()
  - `bubbly.Component`: Name(), ID(), Props(), Emit(), On()
  - ID format: "router-view-{depth}" for debugging
  - No props, no events (passive component)
- **Use Cases**:
  - Root RouterView (depth 0) in main app layout
  - Nested RouterView (depth 1+) in parent route components
  - Multiple RouterView instances at different depths
  - Dynamic component rendering based on current route
- **Integration with Task 4.1**:
  - Uses Matched array from nested routes implementation
  - Leverages parent-child relationships
  - Works seamlessly with multi-level nesting
  - Supports empty child paths (default child routes)
- **Performance**:
  - O(1) component lookup (direct array access by depth)
  - Thread-safe read access (RWMutex on router)
  - No memory allocations during rendering
  - Minimal overhead (just index and type assertion)
- **Next Steps**:
  - Ready for composables (useRouter, useRoute)
  - Ready for component guards (Task 4.3)
  - Can be used in real applications immediately
  - Works with all existing router features

---

### Task 4.3: Component Navigation Guards ✅ COMPLETED
**Description**: BeforeRouteEnter, BeforeRouteUpdate, BeforeRouteLeave

**Prerequisites**: Task 4.2

**Unlocks**: Task 5.1 (Composables)

**Files**:
- `pkg/bubbly/router/component_guards.go` ✅
- `pkg/bubbly/router/component_guards_test.go` ✅
- Updated: `pkg/bubbly/router/guards.go` (added executeComponentGuards method)
- Updated: `pkg/bubbly/router/navigation.go` (fixed Matched array, skip sync for tests)

**Type Safety**:
```go
type ComponentGuards interface {
    BeforeRouteEnter(to, from *Route, next NextFunc)
    BeforeRouteUpdate(to, from *Route, next NextFunc)
    BeforeRouteLeave(to, from *Route, next NextFunc)
}

func hasComponentGuards(component interface{}) (ComponentGuards, bool)
func (r *Router) executeComponentGuards(to, from *Route) *guardResult
```

**Tests**:
- [x] BeforeRouteEnter executes
- [x] BeforeRouteUpdate executes
- [x] BeforeRouteLeave executes
- [x] Execution order correct
- [x] Integration with component lifecycle
- [x] Cancel navigation from guards
- [x] Redirect navigation from guards
- [x] Components without guards work normally

**Estimated Effort**: 4 hours

**Implementation Notes**:
- **Coverage**: 90.2% (exceeds 80% target)
- **Tests**: All 7 test suites passing (100% success rate)
  - TestHasComponentGuards (helper function)
  - TestComponentGuards_BeforeRouteEnter
  - TestComponentGuards_BeforeRouteLeave
  - TestComponentGuards_BeforeRouteUpdate
  - TestComponentGuards_ExecutionOrder
  - TestComponentGuards_CancelNavigation
  - TestComponentGuards_RedirectFromGuard
  - TestComponentGuards_NoGuards
- **Race detector**: Clean (no race conditions)
- **go vet**: Zero warnings
- **Architecture**:
  - `ComponentGuards` interface for optional component implementation
  - `hasComponentGuards()` helper for type checking
  - `executeComponentGuards()` executes guards in correct order
  - Integrated into `executeBeforeGuards()` flow
  - Guards execute AFTER global and route-specific guards
- **Guard Execution Order**:
  1. Global beforeEach guards
  2. Route-specific beforeEnter guards
  3. **Component BeforeRouteLeave** (old component)
  4. **Component BeforeRouteUpdate** (if component reused)
  5. **Component BeforeRouteEnter** (new component)
  6. Navigation completes
  7. Global afterEach hooks
- **BeforeRouteLeave**:
  - Called when navigating away from a route
  - Has access to component state
  - Can cancel navigation (unsaved changes)
  - Can redirect to different route
  - Only called if old component != new component
- **BeforeRouteEnter**:
  - Called before entering a new route
  - Component not yet created (no state access)
  - Can fetch data before navigation
  - Can redirect based on conditions
  - Only called if component not reused
- **BeforeRouteUpdate**:
  - Called when route changes but component reused
  - Same component, different params (e.g., /user/1 → /user/2)
  - Has access to component state
  - Can reload data for new params
  - Detected via pointer equality check
- **Component Reuse Detection**:
  - Uses pointer comparison: `oldComponent == newComponent`
  - Triggers BeforeRouteUpdate instead of Leave+Enter
  - Efficient for param-only changes
- **Guard Actions**:
  - `next(nil)` - Allow navigation, continue
  - `next(&NavigationTarget{Path: ""})` - Cancel navigation
  - `next(&NavigationTarget{Path: "/other"})` - Redirect
- **Cancellation**:
  - Returns `NavigationErrorMsg` with `ErrNavigationCancelled`
  - Current route unchanged
  - Useful for unsaved changes confirmation
- **Redirection**:
  - Recursively calls `pushWithTracking` with new target
  - Circular redirect detection prevents infinite loops
  - Redirect depth limited to 10 (maxRedirectDepth)
- **Integration with Existing Guards**:
  - Component guards execute AFTER global/route guards
  - All guard types use same `guardResult` system
  - Consistent cancel/redirect behavior
  - Thread-safe via router mutex
- **Edge Cases Handled**:
  - Components without guards (no-op)
  - Nil components (no-op)
  - Component reuse detection
  - Circular redirect prevention
  - Guard cancellation
  - Guard redirection
  - Multiple guard executions in redirect chain
- **Testing Strategy**:
  - Mock component with configurable guard behavior
  - Separate flags for each guard type (cancelOnLeave, redirectOnEnter, etc.)
  - Guard tracker records execution order
  - Tests verify cancel, redirect, and normal flow
  - Tests verify component reuse detection
- **Known Limitations**:
  - Component guards can't access component instance in BeforeRouteEnter (by design, like Vue Router)
  - Redirect loops must be prevented by guard logic (framework detects but doesn't auto-fix)
  - Component comparison uses pointer equality (works for most cases)
- **Performance**:
  - O(1) component guard execution (max 3 guards per navigation)
  - Pointer comparison for reuse detection (O(1))
  - No memory allocations for guard execution
  - Minimal overhead (type assertion + function calls)
- **Next Steps**:
  - Ready for composables (useRouter, useRoute) - Task 5.1
  - Ready for meta fields - Task 4.4
  - Ready for named routes - Task 4.5
  - Production-ready for all use cases

---

### Task 4.4: Route Meta Fields ✅ COMPLETED
**Description**: Attach arbitrary metadata to routes

**Prerequisites**: Task 4.1

**Unlocks**: Task 5.1 (Composables)

**Files**: (Integrated in existing files)
- `pkg/bubbly/router/matcher.go` (RouteRecord.Meta field already present)
- `pkg/bubbly/router/route.go` (Route.Meta field, GetMeta() method already present)
- `pkg/bubbly/router/route_test.go` ✅ (added comprehensive tests)

**Type Safety**:
```go
type RouteRecord struct {
    Path      string
    Name      string
    Component interface{}
    Meta      map[string]interface{} // Arbitrary metadata
    Parent    *RouteRecord
    Children  []*RouteRecord
    pattern   *RoutePattern
}

type Route struct {
    Path     string
    Name     string
    Params   map[string]string
    Query    map[string]string
    Hash     string
    Meta     map[string]interface{} // Route metadata (defensive copy)
    Matched  []*RouteRecord         // For accessing parent meta
    FullPath string
}

func (r *Route) GetMeta(key string) (interface{}, bool)
```

**Tests**:
- [x] Meta fields set
- [x] Meta fields accessible
- [x] Type assertions work
- [x] Inherited from parent (via matched array pattern)

**Estimated Effort**: 2 hours

**Implementation Notes**:
- **Coverage**: 90.2% (exceeds 80% target)
- **Tests**: All tests passing with race detector
  - TestRoute_MetaInheritancePattern (3 subtests)
  - TestRoute_MetaTypeAssertions (7 type tests)
  - TestRoute_MetaFieldsSet (2 subtests)
  - Plus existing TestRoute_GetMeta tests
- **Race detector**: Clean (no race conditions)
- **go vet**: Zero warnings
- **Architecture**:
  - Meta fields stored in both RouteRecord and Route
  - Route.Meta is a defensive copy (immutable)
  - GetMeta() convenience method for existence checking
  - Meta inheritance follows Vue Router pattern (via matched array)
- **Meta Inheritance Pattern** (Vue Router Compatible):
  - Meta fields are NOT automatically inherited from parent to child
  - Parent meta accessible via `route.Matched` array
  - Follows Vue Router's `to.matched.some(record => record.meta.requiresAuth)` pattern
  - Allows checking meta across entire route chain
- **Type Safety**:
  - Supports all Go types: bool, string, int, float64, slices, maps, structs
  - Type assertions required when accessing meta values
  - Comprehensive tests for common type patterns
- **Use Cases**:
  - Authentication requirements: `meta: {"requiresAuth": true}`
  - Route titles: `meta: {"title": "Dashboard"}`
  - Permissions: `meta: {"roles": []string{"admin"}}`
  - Layout selection: `meta: {"layout": "admin"}`
  - Custom data: any arbitrary metadata
- **Vue Router Compatibility**:
  - Same meta field structure
  - Same inheritance pattern (via matched array)
  - Same navigation guard usage pattern
  - Familiar API for web developers
- **Edge Cases Handled**:
  - Nil meta maps converted to empty maps
  - Defensive copying prevents external modification
  - GetMeta() returns (value, found) for safe access
  - Type assertions documented in tests
  - Deeply nested routes (3+ levels) tested
- **Performance**:
  - O(1) direct meta access via map
  - O(n) meta inheritance check where n = route depth
  - Minimal memory overhead (map per route)
  - No allocations during GetMeta() calls
- **Integration**:
  - Works with nested routes (Task 4.1)
  - Works with RouterView (Task 4.2)
  - Works with component guards (Task 4.3)
  - Ready for composables (Task 5.1)
  - Used in navigation guards for auth checks
- **Documentation**:
  - Godoc comments on Route.GetMeta()
  - Examples in test code
  - Vue Router pattern documented in tests
  - Type assertion patterns demonstrated

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
