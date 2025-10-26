# Implementation Tasks: Reactivity System

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] Go 1.22+ installed
- [x] Project structure created (`pkg/bubbly/`)
- [x] Testing framework configured (testify)
- [x] Linting configured (golangci-lint)

---

## Phase 1: Core Primitives (Foundation)

### Task 1.1: Ref[T] - Basic Implementation ✅ COMPLETE
**Description:** Implement type-safe reactive reference with Get/Set operations

**Prerequisites:** None (first task)

**Unlocks:** Task 1.2 (Ref watchers), Task 2.1 (Computed)

**Files:**
- `pkg/bubbly/ref.go` ✅
- `pkg/bubbly/ref_test.go` ✅

**Type Safety:**
```go
type Ref[T any] struct {
    mu    sync.RWMutex
    value T
}

func NewRef[T any](value T) *Ref[T]
func (r *Ref[T]) Get() T
func (r *Ref[T]) Set(value T)
```

**Tests:**
- [x] NewRef creates ref with initial value
- [x] Get returns current value
- [x] Set updates value
- [x] Type safety: different types don't mix
- [x] Zero value handling

**Implementation Notes:**
- Completed with TDD approach (RED-GREEN-REFACTOR)
- 100% test coverage achieved
- All quality gates passed (test-race, lint, fmt, vet, build)
- Thread-safe implementation using sync.RWMutex
- Comprehensive godoc comments added
- Tests cover: primitives, structs, pointers, slices, maps, zero values
- No watcher support yet (deferred to Task 1.2 as per spec)
- Dependencies added: github.com/stretchr/testify v1.11.1

**Estimated effort:** 2 hours
**Actual effort:** ~1.5 hours

---

### Task 1.2: Ref[T] - Watcher Support
**Description:** Add watcher registration and notification to Ref

**Prerequisites:** Task 1.1 (basic Ref)

**Unlocks:** Task 3.1 (Watch function)

**Files:**
- `pkg/bubbly/ref.go` (extend)
- `pkg/bubbly/ref_test.go` (extend)

**Type Safety:**
```go
type watcher[T any] struct {
    callback func(newVal, oldVal T)
    options  WatchOptions
}

func (r *Ref[T]) addWatcher(w *watcher[T])
func (r *Ref[T]) removeWatcher(w *watcher[T])
func (r *Ref[T]) notifyWatchers(newVal, oldVal T)
```

**Tests:**
- [ ] Watchers receive notifications
- [ ] Multiple watchers work independently
- [ ] Watcher removal works
- [ ] Notification order is consistent
- [ ] No memory leaks

**Estimated effort:** 3 hours

---

### Task 1.3: Ref[T] - Thread Safety
**Description:** Ensure Ref operations are safe under concurrent access

**Prerequisites:** Task 1.1, Task 1.2

**Unlocks:** All tasks (foundation complete)

**Files:**
- `pkg/bubbly/ref.go` (review and harden)
- `pkg/bubbly/ref_test.go` (add race tests)

**Tests:**
- [ ] Concurrent Get operations
- [ ] Concurrent Set operations
- [ ] Concurrent Get/Set mix
- [ ] Race detector passes
- [ ] Stress test (1000+ concurrent operations)

**Benchmarks:**
```go
BenchmarkRefGet_Concurrent
BenchmarkRefSet_Concurrent
BenchmarkRefGetSet_Mixed
```

**Estimated effort:** 2 hours

---

## Phase 2: Computed Values

### Task 2.1: Computed[T] - Basic Implementation
**Description:** Implement lazy computed values with caching

**Prerequisites:** Task 1.1 (Ref basic)

**Unlocks:** Task 2.2 (Dependency tracking)

**Files:**
- `pkg/bubbly/computed.go`
- `pkg/bubbly/computed_test.go`

**Type Safety:**
```go
type Computed[T any] struct {
    mu    sync.RWMutex
    fn    func() T
    cache T
    dirty bool
}

func NewComputed[T any](fn func() T) *Computed[T]
func (c *Computed[T]) Get() T
```

**Tests:**
- [ ] Lazy evaluation (fn not called until Get)
- [ ] Caching works (fn called only once)
- [ ] Multiple Get calls return cached value
- [ ] Type safety enforced

**Estimated effort:** 3 hours

---

### Task 2.2: Dependency Tracking System
**Description:** Implement automatic dependency tracking for computed values

**Prerequisites:** Task 2.1 (Computed basic), Task 1.3 (Ref complete)

**Unlocks:** Task 2.3 (Cache invalidation)

**Files:**
- `pkg/bubbly/tracker.go`
- `pkg/bubbly/tracker_test.go`
- `pkg/bubbly/ref.go` (integrate tracking)
- `pkg/bubbly/computed.go` (integrate tracking)

**Type Safety:**
```go
type Dependency interface {
    Invalidate()
    AddDependent(dep Dependency)
}

type DepTracker struct {
    mu          sync.RWMutex
    tracking    bool
    currentDeps []Dependency
    stack       []Dependency
}

func (dt *DepTracker) BeginTracking()
func (dt *DepTracker) Track(dep Dependency)
func (dt *DepTracker) EndTracking() []Dependency
```

**Tests:**
- [ ] Track Ref access during computed evaluation
- [ ] Multiple dependencies tracked
- [ ] Nested tracking works (computed → computed)
- [ ] Circular dependency detection
- [ ] Thread safety

**Estimated effort:** 4 hours

---

### Task 2.3: Cache Invalidation
**Description:** Invalidate computed cache when dependencies change

**Prerequisites:** Task 2.2 (Dependency tracking)

**Unlocks:** Full reactive system

**Files:**
- `pkg/bubbly/computed.go` (extend)
- `pkg/bubbly/computed_test.go` (extend)

**Type Safety:**
```go
func (c *Computed[T]) Invalidate()
func (c *Computed[T]) isDirty() bool
```

**Tests:**
- [ ] Cache invalidates on dependency change
- [ ] Recomputation happens on next Get
- [ ] Chain invalidation (A → B → C)
- [ ] Minimal recomputation (only when needed)
- [ ] No redundant evaluations

**Estimated effort:** 2 hours

---

## Phase 3: Watcher System

### Task 3.1: Watch Function
**Description:** Implement Watch function for creating watchers

**Prerequisites:** Task 1.2 (Ref watchers)

**Unlocks:** Task 3.2 (Watch options)

**Files:**
- `pkg/bubbly/watch.go`
- `pkg/bubbly/watch_test.go`

**Type Safety:**
```go
type WatchCallback[T any] func(newVal, oldVal T)
type WatchCleanup func()

func Watch[T any](
    source *Ref[T],
    callback WatchCallback[T],
    options ...WatchOption,
) WatchCleanup
```

**Tests:**
- [ ] Callback executes on value change
- [ ] Cleanup function stops watching
- [ ] Multiple watches on same Ref
- [ ] Type-safe callback parameter
- [ ] No panic on cleanup

**Estimated effort:** 3 hours

---

### Task 3.2: Watch Options
**Description:** Implement watch options (immediate, deep, flush)

**Prerequisites:** Task 3.1 (basic Watch)

**Unlocks:** Complete watcher system

**Files:**
- `pkg/bubbly/watch.go` (extend)
- `pkg/bubbly/watch_test.go` (extend)

**Type Safety:**
```go
type WatchOptions struct {
    Deep      bool
    Immediate bool
    Flush     string
}

type WatchOption func(*WatchOptions)

func WithDeep() WatchOption
func WithImmediate() WatchOption
func WithFlush(mode string) WatchOption
```

**Tests:**
- [ ] Immediate: callback runs immediately
- [ ] Deep: nested changes detected
- [ ] Flush: timing control works
- [ ] Option composition works
- [ ] Default options correct

**Estimated effort:** 2 hours

---

## Phase 4: Integration & Polish

### Task 4.1: Error Handling
**Description:** Add comprehensive error handling and validation

**Prerequisites:** All previous tasks

**Unlocks:** Production readiness

**Files:**
- `pkg/bubbly/errors.go`
- All implementation files (add error checks)

**Type Safety:**
```go
var (
    ErrCircularDependency = errors.New("circular dependency detected")
    ErrMaxDepth          = errors.New("max dependency depth exceeded")
    ErrNilCallback       = errors.New("callback cannot be nil")
)
```

**Tests:**
- [ ] Circular dependency detected
- [ ] Max depth enforced (100)
- [ ] Nil checks prevent panics
- [ ] Error messages are clear
- [ ] Errors are documented

**Estimated effort:** 2 hours

---

### Task 4.2: Performance Optimization
**Description:** Profile and optimize hot paths

**Prerequisites:** Task 4.1 (error handling)

**Unlocks:** None (optimization)

**Files:**
- All implementation files (optimize)
- Benchmarks (add/improve)

**Optimizations:**
- [ ] Reduce lock contention
- [ ] Minimize allocations
- [ ] Pool watcher objects
- [ ] Optimize notification loops
- [ ] Cache optimization

**Benchmarks:**
```go
BenchmarkRefGet      1000000000   1.2 ns/op
BenchmarkRefSet        10000000  90.5 ns/op
BenchmarkComputed       5000000 250  ns/op
BenchmarkWatch         10000000 105  ns/op
```

**Estimated effort:** 3 hours

---

### Task 4.3: Documentation
**Description:** Complete API documentation and examples

**Prerequisites:** All implementation complete

**Unlocks:** Public API ready

**Files:**
- `pkg/bubbly/doc.go` (package docs)
- All public APIs (godoc comments)
- `pkg/bubbly/example_test.go` (examples)

**Documentation:**
- [ ] Package overview
- [ ] Ref API documented
- [ ] Computed API documented
- [ ] Watch API documented
- [ ] 5+ runnable examples
- [ ] Migration guide from manual state

**Examples:**
```go
func ExampleNewRef()
func ExampleRef_Get()
func ExampleRef_Set()
func ExampleNewComputed()
func ExampleWatch()
func ExampleWatch_withOptions()
```

**Estimated effort:** 3 hours

---

## Phase 5: Testing & Validation

### Task 5.1: Integration Tests
**Description:** Test full reactive system integration

**Prerequisites:** All implementation tasks

**Unlocks:** None (validation)

**Files:**
- `tests/integration/reactivity_test.go`

**Tests:**
- [ ] Ref → Computed → Watcher flow
- [ ] Multiple component interaction
- [ ] Concurrent access patterns
- [ ] Long-running stability
- [ ] Memory leak detection

**Estimated effort:** 4 hours

---

### Task 5.2: Benchmarking Suite
**Description:** Comprehensive performance benchmarking

**Prerequisites:** Task 4.2 (optimizations)

**Unlocks:** Performance baseline

**Files:**
- `pkg/bubbly/bench_test.go`

**Benchmarks:**
- [ ] Single Ref operations
- [ ] Computed evaluation
- [ ] Watcher notification
- [ ] Large ref counts (1000+)
- [ ] Memory profiling

**Estimated effort:** 2 hours

---

### Task 5.3: Example Applications
**Description:** Create example apps demonstrating reactivity

**Prerequisites:** All tasks complete

**Unlocks:** Documentation and showcase

**Files:**
- `cmd/examples/reactive-counter/main.go`
- `cmd/examples/reactive-todo/main.go`

**Examples:**
- [ ] Simple counter (basic Ref usage)
- [ ] Todo list (Ref + Computed)
- [ ] Form validation (multiple Refs + Computed)
- [ ] Async data (Ref + Watch)

**Estimated effort:** 4 hours

---

## Task Dependency Graph

```
Prerequisites
    ↓
Task 1.1: Ref Basic ────────┐
    ↓                       │
Task 1.2: Ref Watchers      │
    ↓                       │
Task 1.3: Ref Thread Safety │
    ↓                       │
    ├───────────────────────┘
    ↓
Task 2.1: Computed Basic
    ↓
Task 2.2: Dependency Tracking
    ↓
Task 2.3: Cache Invalidation
    ↓
Task 3.1: Watch Function
    ↓
Task 3.2: Watch Options
    ↓
Task 4.1: Error Handling
    ↓
Task 4.2: Performance Optimization
    ↓
Task 4.3: Documentation
    ↓
Task 5.1: Integration Tests
    ↓
Task 5.2: Benchmarking
    ↓
Task 5.3: Example Apps
    ↓
Unlocks: 02-component-model
```

---

## Validation Checklist

### Code Quality
- [ ] All types strictly typed (no `any` except generic constraints)
- [ ] All public APIs have godoc comments
- [ ] All tests pass (`go test ./...`)
- [ ] Race detector passes (`go test -race ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Test coverage > 80% (`go test -cover ./...`)

### Functionality
- [ ] Ref Get/Set works
- [ ] Computed lazy evaluation works
- [ ] Computed caching works
- [ ] Computed dependency tracking works
- [ ] Watch notifications work
- [ ] Watch cleanup works
- [ ] Thread safety verified
- [ ] No memory leaks

### Performance
- [ ] Ref Get < 10ns
- [ ] Ref Set < 100ns
- [ ] Computed Get < 1μs (simple computation)
- [ ] Watch notification < 200ns
- [ ] Memory usage acceptable

### Documentation
- [ ] README.md in `pkg/bubbly/`
- [ ] All public APIs documented
- [ ] 5+ runnable examples
- [ ] Migration guide written
- [ ] Performance characteristics documented

### Integration
- [ ] Used in component system (when available)
- [ ] Works with Bubbletea message loop
- [ ] Composable functions can use it
- [ ] No conflicts with other systems

---

## Time Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Core Primitives | 3 | 7 hours |
| Phase 2: Computed Values | 3 | 9 hours |
| Phase 3: Watcher System | 2 | 5 hours |
| Phase 4: Integration & Polish | 3 | 8 hours |
| Phase 5: Testing & Validation | 3 | 10 hours |
| **Total** | **14 tasks** | **39 hours (~1 week)** |

---

## Development Order

### Day 1-2: Core Primitives
- Task 1.1, 1.2, 1.3
- Foundation complete

### Day 3: Computed Values
- Task 2.1, 2.2, 2.3
- Reactive system functional

### Day 4: Watchers
- Task 3.1, 3.2
- Side effects working

### Day 5: Polish
- Task 4.1, 4.2, 4.3
- Production ready

### Day 6-7: Validation
- Task 5.1, 5.2, 5.3
- Fully tested and documented

---

## Success Criteria

✅ **Definition of Done:**
1. All tests pass with > 80% coverage
2. Race detector shows no issues
3. Benchmarks meet performance targets
4. Documentation complete with examples
5. Integrated with next feature (component model)
6. No TODO comments (create issues instead)
7. Code reviewed and approved
8. Examples work and are documented

✅ **Ready for Next Feature:**
- Component model can use Ref for state
- Composables can return reactive values
- Lifecycle hooks can create/cleanup watchers
