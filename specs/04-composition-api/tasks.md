# Implementation Tasks: Composition API

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] Feature 01: Reactivity System complete
- [x] Feature 02: Component Model complete
- [x] Feature 03: Lifecycle Hooks complete
- [ ] All previous features tested and working
- [ ] Context system available
- [ ] Go 1.22+ installed

---

## Phase 1: Context Extension

### Task 1.1: Extended Context Structure ✅ COMPLETE
**Description:** Extend Context to support composable APIs (Provide/Inject)

**Prerequisites:** Features 01, 02, 03 complete

**Unlocks:** Task 1.2 (Provide/Inject implementation)

**Files:**
- `pkg/bubbly/context.go` (extend) ✅
- `pkg/bubbly/context_test.go` (extend) ✅
- `pkg/bubbly/component.go` (extend) ✅

**Type Safety:**
```go
type Context struct {
    component *componentImpl
    
    // Existing APIs
    Ref       func(value interface{}) *Ref[interface{}]
    Computed  func(fn func() interface{}) *Computed[interface{}]
    Watch     func(ref *Ref[interface{}], callback WatchCallback)
    
    // Composition API additions
    Provide   func(key string, value interface{})
    Inject    func(key string, defaultValue interface{}) interface{}
    
    // Internal
    provides  map[string]interface{}
}
```

**Tests:**
- [x] Context structure updated
- [x] Provide/Inject methods added
- [x] Provides map initialized
- [x] Type safety maintained

**Implementation Notes:**
- Added `Provide(key, value)` method to Context - stores values in component's provides map
- Added `Inject(key, defaultValue)` method to Context - retrieves from ancestor tree
- Extended componentImpl with `provides map[string]interface{}` and `providesMu sync.RWMutex`
- Changed parent field from `*Component` to `*componentImpl` for tree traversal
- Implemented `inject(key, defaultValue)` helper method with recursive tree walking
- Thread-safe with RWMutex protecting provides map
- Comprehensive test coverage: 9 test cases covering all scenarios
- All tests pass with race detector
- Coverage: 96.0% (exceeds 80% requirement)

**Estimated effort:** 2 hours (actual: ~2 hours)

---

### Task 1.2: Provide/Inject Implementation ✅ COMPLETE
**Description:** Implement provide/inject functionality with tree traversal

**Prerequisites:** Task 1.1

**Unlocks:** Task 2.1 (Standard composables)

**Files:**
- `pkg/bubbly/context.go` (extend) ✅
- `pkg/bubbly/component.go` (extend) ✅
- `pkg/bubbly/context_test.go` (tests added) ✅

**Type Safety:**
```go
func (c *componentImpl) inject(key string, defaultValue interface{}) interface{}
```

**Tests:**
- [x] Provide stores value
- [x] Inject retrieves from parent
- [x] Inject walks up tree
- [x] Default value returned if not found
- [x] Nearest provider wins
- [x] Reactive values propagate

**Implementation Notes:**
- Task 1.2 was completed together with Task 1.1 as a single cohesive implementation
- The `inject()` method implements recursive tree traversal with early return optimization
- Tests cover all scenarios: self-injection, parent injection, deep tree (4 levels), nearest wins, multiple keys, reactive values
- Thread-safe with RWMutex protecting the provides map during tree traversal
- Performance: O(depth) time complexity for inject lookups

**Estimated effort:** 4 hours (actual: included in Task 1.1, ~2 hours total for both)

---

### Task 1.3: Provide/Inject Type Safety Helpers
**Description:** Create type-safe provide/inject helpers using generics

**Prerequisites:** Task 1.2

**Unlocks:** Task 2.1 (Standard composables)

**Files:**
- `pkg/bubbly/provide_inject.go`
- `pkg/bubbly/provide_inject_test.go` (extend)

**Type Safety:**
```go
type ProvideKey[T any] struct {
    key string
}

func NewProvideKey[T any](key string) ProvideKey[T]
func ProvideTyped[T any](ctx *Context, key ProvideKey[T], value T)
func InjectTyped[T any](ctx *Context, key ProvideKey[T], defaultValue T) T
```

**Tests:**
- [ ] Type-safe provide
- [ ] Type-safe inject
- [ ] Compile-time type checking
- [ ] Key generation works
- [ ] Type mismatch caught

**Estimated effort:** 3 hours

---

## Phase 2: Standard Composables

### Task 2.1: UseState Composable
**Description:** Implement UseState for simple state management

**Prerequisites:** Task 1.3

**Unlocks:** Task 2.2 (UseEffect)

**Files:**
- `pkg/bubbly/composables/use_state.go`
- `pkg/bubbly/composables/use_state_test.go`

**Type Safety:**
```go
type UseStateReturn[T any] struct {
    Value *Ref[T]
    Set   func(T)
    Get   func() T
}

func UseState[T any](ctx *Context, initial T) UseStateReturn[T]
```

**Tests:**
- [ ] Creates ref with initial value
- [ ] Set updates value
- [ ] Get retrieves value
- [ ] Type safety enforced
- [ ] Multiple instances independent

**Estimated effort:** 2 hours

---

### Task 2.2: UseEffect Composable
**Description:** Implement UseEffect for side effect management

**Prerequisites:** Task 2.1

**Unlocks:** Task 2.3 (UseAsync)

**Files:**
- `pkg/bubbly/composables/use_effect.go`
- `pkg/bubbly/composables/use_effect_test.go`

**Type Safety:**
```go
type UseEffectCleanup func()

func UseEffect(ctx *Context, effect func() UseEffectCleanup, deps ...*Ref[any])
```

**Tests:**
- [ ] Effect runs on mount
- [ ] Effect runs on deps change
- [ ] Cleanup executes before re-run
- [ ] Cleanup executes on unmount
- [ ] No deps: runs every update
- [ ] Empty deps: runs once

**Estimated effort:** 3 hours

---

### Task 2.3: UseAsync Composable
**Description:** Implement UseAsync for async data fetching

**Prerequisites:** Task 2.2

**Unlocks:** Task 2.4 (UseDebounce)

**Files:**
- `pkg/bubbly/composables/use_async.go`
- `pkg/bubbly/composables/use_async_test.go`

**Type Safety:**
```go
type UseAsyncReturn[T any] struct {
    Data    *Ref[*T]
    Loading *Ref[bool]
    Error   *Ref[error]
    Execute func()
    Reset   func()
}

func UseAsync[T any](ctx *Context, fetcher func() (*T, error)) UseAsyncReturn[T]
```

**Tests:**
- [ ] Execute triggers fetch
- [ ] Loading state managed
- [ ] Data populated on success
- [ ] Error set on failure
- [ ] Reset clears state
- [ ] Concurrent executions handled

**Estimated effort:** 4 hours

---

### Task 2.4: UseDebounce Composable
**Description:** Implement UseDebounce for debounced values

**Prerequisites:** Task 2.3

**Unlocks:** Task 2.5 (UseThrottle)

**Files:**
- `pkg/bubbly/composables/use_debounce.go`
- `pkg/bubbly/composables/use_debounce_test.go`

**Type Safety:**
```go
func UseDebounce[T any](ctx *Context, value *Ref[T], delay time.Duration) *Ref[T]
```

**Tests:**
- [ ] Debounces value changes
- [ ] Delay respected
- [ ] Timer cleanup on unmount
- [ ] Multiple rapid changes handled
- [ ] Final value propagated

**Estimated effort:** 3 hours

---

### Task 2.5: UseThrottle Composable
**Description:** Implement UseThrottle for throttled function execution

**Prerequisites:** Task 2.4

**Unlocks:** Task 3.1 (UseForm)

**Files:**
- `pkg/bubbly/composables/use_throttle.go`
- `pkg/bubbly/composables/use_throttle_test.go`

**Type Safety:**
```go
func UseThrottle(ctx *Context, fn func(), delay time.Duration) func()
```

**Tests:**
- [ ] Throttles function calls
- [ ] Delay respected
- [ ] First call immediate
- [ ] Subsequent calls delayed
- [ ] Cleanup on unmount

**Estimated effort:** 3 hours

---

## Phase 3: Complex Composables

### Task 3.1: UseForm Composable
**Description:** Implement UseForm for form management with validation

**Prerequisites:** Task 2.5

**Unlocks:** Task 3.2 (UseLocalStorage)

**Files:**
- `pkg/bubbly/composables/use_form.go`
- `pkg/bubbly/composables/use_form_test.go`

**Type Safety:**
```go
type UseFormReturn[T any] struct {
    Values   *Ref[T]
    Errors   *Ref[map[string]string]
    Touched  *Ref[map[string]bool]
    IsValid  *Computed[bool]
    IsDirty  *Computed[bool]
    Submit   func()
    Reset    func()
    SetField func(field string, value interface{})
}

func UseForm[T any](
    ctx *Context,
    initial T,
    validate func(T) map[string]string,
) UseFormReturn[T]
```

**Tests:**
- [ ] Form initialization
- [ ] Field updates
- [ ] Validation triggers
- [ ] Submit validates
- [ ] Reset works
- [ ] Dirty tracking
- [ ] Touched tracking

**Estimated effort:** 5 hours

---

### Task 3.2: UseLocalStorage Composable
**Description:** Implement UseLocalStorage for persistent state

**Prerequisites:** Task 3.1

**Unlocks:** Task 3.3 (UseEventListener)

**Files:**
- `pkg/bubbly/composables/use_local_storage.go`
- `pkg/bubbly/composables/use_local_storage_test.go`

**Type Safety:**
```go
func UseLocalStorage[T any](ctx *Context, key string, initial T) UseStateReturn[T]
```

**Tests:**
- [ ] Loads from storage on mount
- [ ] Saves on change
- [ ] JSON serialization
- [ ] Deserialization
- [ ] Storage unavailable handled
- [ ] Type safety maintained

**Estimated effort:** 4 hours

---

### Task 3.3: UseEventListener Composable
**Description:** Implement UseEventListener for event handling with cleanup

**Prerequisites:** Task 3.2

**Unlocks:** Task 4.1 (Integration)

**Files:**
- `pkg/bubbly/composables/use_event_listener.go`
- `pkg/bubbly/composables/use_event_listener_test.go`

**Type Safety:**
```go
func UseEventListener(ctx *Context, event string, handler func()) func()
```

**Tests:**
- [ ] Registers event listener
- [ ] Handler executes on event
- [ ] Cleanup removes listener
- [ ] Multiple listeners work
- [ ] Auto-cleanup on unmount

**Estimated effort:** 3 hours

---

## Phase 4: Integration & Utilities

### Task 4.1: Composable Package Organization
**Description:** Organize composables into logical packages

**Prerequisites:** All composables implemented

**Unlocks:** Task 4.2 (Documentation)

**Files:**
- `pkg/bubbly/composables/doc.go`
- `pkg/bubbly/composables/README.md`

**Organization:**
```
pkg/bubbly/composables/
├── doc.go              # Package documentation
├── README.md           # User guide
├── use_state.go        # State management
├── use_effect.go       # Side effects
├── use_async.go        # Async operations
├── use_debounce.go     # Debouncing
├── use_throttle.go     # Throttling
├── use_form.go         # Forms
├── use_local_storage.go # Persistence
└── use_event_listener.go # Events
```

**Tests:**
- [ ] Package imports correctly
- [ ] No circular dependencies
- [ ] Documentation complete
- [ ] Examples provided

**Estimated effort:** 2 hours

---

### Task 4.2: Composable Testing Utilities
**Description:** Create utilities for testing composables

**Prerequisites:** Task 4.1

**Unlocks:** Task 4.3 (Examples)

**Files:**
- `pkg/bubbly/testing/composables.go`
- `pkg/bubbly/testing/composables_test.go`

**Type Safety:**
```go
func NewTestContext() *Context
func MockComposable[T any](ctx *Context, value T) UseStateReturn[T]
func AssertComposableCleanup(t *testing.T, cleanup func())
```

**Tests:**
- [ ] Test context creation
- [ ] Mock composables work
- [ ] Cleanup assertions work
- [ ] Integration test helpers

**Estimated effort:** 3 hours

---

### Task 4.3: Example Composables
**Description:** Create example composables demonstrating patterns

**Prerequisites:** Task 4.2

**Unlocks:** Task 5.1 (Performance)

**Files:**
- `cmd/examples/composables/counter.go`
- `cmd/examples/composables/async-data.go`
- `cmd/examples/composables/form.go`
- `cmd/examples/composables/provide-inject.go`

**Examples:**
- [ ] UseCounter (basic pattern)
- [ ] UseAsync (data fetching)
- [ ] UseForm (complex state)
- [ ] Provide/Inject (dependency injection)
- [ ] Composable chains

**Estimated effort:** 4 hours

---

## Phase 5: Performance & Polish

### Task 5.1: Performance Optimization
**Description:** Optimize composable performance

**Prerequisites:** Task 4.3

**Unlocks:** Task 5.2 (Documentation)

**Files:**
- All composable files (optimize)
- Benchmarks

**Optimizations:**
- [ ] Composable call overhead minimized
- [ ] Inject lookup caching
- [ ] Memory allocations reduced
- [ ] Ref access optimized
- [ ] Cleanup efficient

**Benchmarks:**
```go
BenchmarkUseState
BenchmarkUseAsync
BenchmarkUseEffect
BenchmarkProvideInject
BenchmarkComposableChain
```

**Targets:**
- UseState: < 200ns
- UseAsync: < 1μs
- Provide/Inject: < 500ns

**Estimated effort:** 4 hours

---

### Task 5.2: Comprehensive Documentation
**Description:** Complete documentation for Composition API

**Prerequisites:** Task 5.1

**Unlocks:** Task 6.1 (Integration tests)

**Files:**
- `pkg/bubbly/composables/doc.go`
- `docs/guides/composition-api.md`
- `docs/guides/standard-composables.md`
- `docs/guides/custom-composables.md`

**Documentation:**
- [ ] Package overview
- [ ] Each composable documented
- [ ] Composable pattern explained
- [ ] Provide/Inject guide
- [ ] 25+ examples
- [ ] Best practices
- [ ] Common patterns
- [ ] Troubleshooting

**Estimated effort:** 5 hours

---

### Task 5.3: Error Handling Enhancement
**Description:** Add comprehensive error handling and validation

**Prerequisites:** Task 5.2

**Unlocks:** Task 6.1 (Integration tests)

**Files:**
- `pkg/bubbly/composables/errors.go`
- All composable files (add error checks)

**Type Safety:**
```go
var (
    ErrComposableOutsideSetup = errors.New("composable called outside Setup")
    ErrCircularComposable     = errors.New("circular composable dependency")
    ErrInjectNotFound         = errors.New("inject key not found")
    ErrInvalidComposableState = errors.New("invalid composable state")
)
```

**Tests:**
- [ ] Errors defined
- [ ] Error messages clear
- [ ] Recovery mechanisms work
- [ ] Validation errors caught

**Estimated effort:** 3 hours

---

## Phase 6: Testing & Validation

### Task 6.1: Integration Tests
**Description:** Test composables integrated with components

**Prerequisites:** All implementation complete

**Unlocks:** Task 6.2 (E2E tests)

**Files:**
- `tests/integration/composables_test.go`

**Tests:**
- [ ] Composables in components
- [ ] Provide/Inject across tree
- [ ] Composable chains
- [ ] Lifecycle integration
- [ ] Cleanup verification
- [ ] State isolation

**Estimated effort:** 5 hours

---

### Task 6.2: End-to-End Examples
**Description:** Create complete applications using composables

**Prerequisites:** Task 6.1

**Unlocks:** Task 6.3 (Performance validation)

**Files:**
- `cmd/examples/todo-composables/main.go`
- `cmd/examples/user-dashboard/main.go`
- `cmd/examples/form-wizard/main.go`

**Examples:**
- [ ] Todo app with UseForm
- [ ] Dashboard with UseAsync
- [ ] Form wizard with provide/inject
- [ ] All composables demonstrated

**Estimated effort:** 6 hours

---

### Task 6.3: Performance Validation
**Description:** Validate performance meets targets

**Prerequisites:** Task 6.2

**Unlocks:** Production readiness

**Files:**
- Performance test suite
- Profiling reports

**Validation:**
- [ ] All benchmarks meet targets
- [ ] No memory leaks
- [ ] Reasonable overhead vs manual
- [ ] Profiling shows no hotspots

**Estimated effort:** 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01, 02, 03)
    ↓
Phase 1: Context Extension
    ├─> Task 1.1: Extended context
    ├─> Task 1.2: Provide/inject
    └─> Task 1.3: Type safety helpers
    ↓
Phase 2: Standard Composables
    ├─> Task 2.1: UseState
    ├─> Task 2.2: UseEffect
    ├─> Task 2.3: UseAsync
    ├─> Task 2.4: UseDebounce
    └─> Task 2.5: UseThrottle
    ↓
Phase 3: Complex Composables
    ├─> Task 3.1: UseForm
    ├─> Task 3.2: UseLocalStorage
    └─> Task 3.3: UseEventListener
    ↓
Phase 4: Integration & Utilities
    ├─> Task 4.1: Package organization
    ├─> Task 4.2: Testing utilities
    └─> Task 4.3: Example composables
    ↓
Phase 5: Performance & Polish
    ├─> Task 5.1: Performance optimization
    ├─> Task 5.2: Documentation
    └─> Task 5.3: Error handling
    ↓
Phase 6: Testing & Validation
    ├─> Task 6.1: Integration tests
    ├─> Task 6.2: E2E examples
    └─> Task 6.3: Performance validation
    ↓
Complete: Ready for Features 05, 06
```

---

## Validation Checklist

### Code Quality
- [ ] All types strictly typed
- [ ] All composables documented
- [ ] All tests pass
- [ ] Race detector passes
- [ ] Linter passes
- [ ] Test coverage > 80%

### Functionality
- [ ] Provide/inject works
- [ ] All standard composables work
- [ ] Composable chains work
- [ ] Cleanup guaranteed
- [ ] Type safety enforced
- [ ] Integration with features 01-03

### Performance
- [ ] Composable call < 100ns
- [ ] UseState < 200ns
- [ ] Provide/inject < 500ns
- [ ] No memory leaks
- [ ] Acceptable overhead

### Documentation
- [ ] Package docs complete
- [ ] All composables documented
- [ ] 25+ examples
- [ ] Best practices documented
- [ ] Troubleshooting guide
- [ ] Migration patterns

### Integration
- [ ] Works with components
- [ ] Works with reactivity
- [ ] Works with lifecycle
- [ ] Ready for directives
- [ ] Ready for built-in components

---

## Time Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Context Extension | 3 | 9 hours |
| Phase 2: Standard Composables | 5 | 15 hours |
| Phase 3: Complex Composables | 3 | 12 hours |
| Phase 4: Integration & Utilities | 3 | 9 hours |
| Phase 5: Performance & Polish | 3 | 12 hours |
| Phase 6: Testing & Validation | 3 | 14 hours |
| **Total** | **20 tasks** | **71 hours (~1.8 weeks)** |

---

## Development Order

### Week 1: Core Composables
- Days 1-2: Phase 1 (Context extension)
- Days 3-5: Phase 2 (Standard composables)

### Week 2: Advanced & Polish
- Days 1-2: Phase 3 (Complex composables)
- Day 3: Phase 4 (Integration)
- Days 4-5: Phase 5 & 6 (Polish and validation)

---

## Success Criteria

✅ **Definition of Done:**
1. All tests pass with > 80% coverage
2. Race detector shows no issues
3. Benchmarks meet performance targets
4. Complete documentation with 25+ examples
5. Integration tests demonstrate full functionality
6. E2E examples work correctly
7. No memory leaks in long-running tests
8. Ready for features 05 and 06

✅ **Ready for Next Features:**
- Directives can use composables
- Built-in components can use composables
- Community can create composable libraries
- Developers understand composable pattern

---

## Risk Mitigation

### Risk: Performance Overhead
**Mitigation:**
- Benchmark early and often
- Optimize hot paths
- Profile regularly
- Accept reasonable overhead for DX

### Risk: Complex Type Signatures
**Mitigation:**
- Provide type helpers
- Document patterns clearly
- Use examples extensively
- Test with real use cases

### Risk: Memory Leaks
**Mitigation:**
- Comprehensive leak tests
- Auto-cleanup via lifecycle
- Clear cleanup documentation
- Memory profiling

### Risk: API Confusion
**Mitigation:**
- Clear naming conventions
- Comprehensive examples
- User testing feedback
- Compare with Vue patterns

---

## Notes

### Design Decisions
- Use* prefix for composables
- Context always first parameter
- Return structs with named fields
- Explicit cleanup via lifecycle
- Type-safe provide/inject

### Trade-offs
- **Boilerplate vs Type Safety:** More explicit types for safety
- **Performance vs DX:** Slight overhead for better experience
- **Flexibility vs Convention:** Strong conventions with escape hatches

### Future Enhancements
- Composable registry
- Async composables (suspense-like)
- Dev tools integration
- Hot reload support
- Testing utilities expansion
