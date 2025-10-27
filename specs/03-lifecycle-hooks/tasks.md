# Implementation Tasks: Lifecycle Hooks

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] Feature 01: Reactivity System complete
- [x] Feature 02: Component Model complete
- [ ] Component system tests passing
- [ ] Reactivity system integrated
- [ ] Go 1.22+ installed

---

## Phase 1: Lifecycle Manager Foundation

### Task 1.1: Lifecycle Manager Structure ✅ COMPLETE
**Description:** Define LifecycleManager struct and basic initialization

**Prerequisites:** Feature 02 complete ✅

**Unlocks:** Task 1.2 (Hook registration)

**Files:**
- `pkg/bubbly/lifecycle.go` ✅
- `pkg/bubbly/lifecycle_test.go` ✅

**Type Safety:**
```go
type LifecycleManager struct {
    component      *componentImpl
    hooks          map[string][]lifecycleHook
    cleanups       []CleanupFunc
    watchers       []watcherCleanup
    mounted        bool
    unmounting     bool
    updateCount    int
}

type lifecycleHook struct {
    id           string
    callback     func()
    dependencies []*Ref[any]
    lastValues   []any
    order        int
}

type CleanupFunc func()
```

**Tests:**
- [x] LifecycleManager creation
- [x] Initial state correct
- [x] Hooks map initialized
- [x] State flags correct

**Implementation Notes:**
- Created `lifecycle.go` with LifecycleManager struct and newLifecycleManager constructor
- Created `lifecycle_test.go` with table-driven tests (4 test functions)
- All tests pass with race detector
- Coverage: 96.2% (exceeds 80% requirement)
- Linter clean (no warnings)
- Fields marked with nolint comments for future tasks
- Types: CleanupFunc (exported), lifecycleHook (unexported), watcherCleanup (unexported)
- Constructor initializes all maps/slices to prevent nil panics
- Initial state: mounted=false, unmounting=false, updateCount=0

**Estimated effort:** 2 hours ✅ (Actual: ~1.5 hours)

---

### Task 1.2: Hook Registration Methods ✅ COMPLETE
**Description:** Implement hook registration in Context

**Prerequisites:** Task 1.1 ✅

**Unlocks:** Task 2.1 (Hook execution)

**Files:**
- `pkg/bubbly/context.go` (extend) ✅
- `pkg/bubbly/lifecycle.go` (extend) ✅
- `pkg/bubbly/lifecycle_test.go` (extend) ✅
- `pkg/bubbly/component.go` (extend) ✅

**Type Safety:**
```go
func (ctx *Context) OnMounted(hook func())
func (ctx *Context) OnUpdated(hook func(), deps ...*Ref[any])
func (ctx *Context) OnUnmounted(hook func())
func (ctx *Context) OnBeforeUpdate(hook func())
func (ctx *Context) OnBeforeUnmount(hook func())
func (ctx *Context) OnCleanup(cleanup CleanupFunc)
```

**Tests:**
- [x] Hook registration works
- [x] Multiple hooks registered
- [x] Dependencies stored correctly
- [x] Order preserved
- [x] Type safety enforced

**Implementation Notes:**
- Added `lifecycle *LifecycleManager` field to componentImpl
- Added `hookIDCounter` atomic counter for unique hook IDs
- Implemented `registerHook(hookType string, hook lifecycleHook)` method on LifecycleManager
- Implemented all Context methods: OnMounted, OnUpdated, OnUnmounted, OnBeforeUpdate, OnBeforeUnmount, OnCleanup
- OnUpdated supports variadic dependencies with initial value capture
- Lifecycle manager lazy-initialized on first hook registration
- Hook order tracked automatically based on registration sequence
- Added 6 comprehensive test functions with table-driven tests
- All tests pass with race detector
- Coverage: 94.0% (exceeds 80% requirement)
- Linter clean (no warnings)
- Code formatted with gofmt

**Key Implementation Details:**
- Hook IDs generated using atomic counter (fmt.Sprintf("hook-%d", id))
- Dependencies captured at registration time for OnUpdated
- Cleanup functions stored in slice for LIFO execution
- Lifecycle manager created lazily to avoid overhead when not used
- All hook types supported: mounted, beforeUpdate, updated, beforeUnmount, unmounted

**Estimated effort:** 3 hours ✅ (Actual: ~2.5 hours)

---

### Task 1.3: Lifecycle State Management
**Description:** Implement state tracking (mounted, unmounting, etc.)

**Prerequisites:** Task 1.2

**Unlocks:** Task 2.1 (Hook execution)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)

**Type Safety:**
```go
func (lm *LifecycleManager) IsMounted() bool
func (lm *LifecycleManager) IsUnmounting() bool
func (lm *LifecycleManager) setMounted(mounted bool)
func (lm *LifecycleManager) setUnmounting(unmounting bool)
```

**Tests:**
- [ ] State transitions correct
- [ ] State queries work
- [ ] Thread-safe state access
- [ ] State persists correctly

**Estimated effort:** 2 hours

---

## Phase 2: Hook Execution

### Task 2.1: onMounted Execution
**Description:** Implement onMounted hook execution

**Prerequisites:** Task 1.3

**Unlocks:** Task 2.2 (onUpdated)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)
- `pkg/bubbly/component.go` (integrate)

**Type Safety:**
```go
func (lm *LifecycleManager) executeMounted()
func (lm *LifecycleManager) executeHooks(hookType string)
func (lm *LifecycleManager) safeExecuteHook(hookType string, hook LifecycleHook)
```

**Tests:**
- [ ] Hooks execute after mount
- [ ] Execution order correct
- [ ] Only executes once
- [ ] Integration with Component.View()
- [ ] Multiple hooks work

**Estimated effort:** 3 hours

---

### Task 2.2: onUpdated Execution
**Description:** Implement onUpdated hook execution with dependency tracking

**Prerequisites:** Task 2.1

**Unlocks:** Task 2.3 (onUnmounted)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)

**Type Safety:**
```go
func (lm *LifecycleManager) executeUpdated()
func (lm *LifecycleManager) checkDependencies(hook *LifecycleHook) bool
func (lm *LifecycleManager) updateLastValues(hook *LifecycleHook)
```

**Tests:**
- [ ] Hooks execute on update
- [ ] Dependencies tracked correctly
- [ ] Only runs when deps change
- [ ] No deps: runs every time
- [ ] Multiple dependencies work

**Estimated effort:** 4 hours

---

### Task 2.3: onUnmounted Execution
**Description:** Implement onUnmounted hook execution and cleanup

**Prerequisites:** Task 2.2

**Unlocks:** Task 3.1 (Error handling)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)
- `pkg/bubbly/component.go` (integrate Unmount)

**Type Safety:**
```go
func (lm *LifecycleManager) executeUnmounted()
func (lm *LifecycleManager) executeCleanups()
func (c *componentImpl) Unmount()
```

**Tests:**
- [ ] onUnmounted hooks execute
- [ ] Cleanup functions execute
- [ ] Reverse order execution
- [ ] Children unmounted first
- [ ] Cleanup guaranteed

**Estimated effort:** 3 hours

---

## Phase 3: Error Handling & Safety

### Task 3.1: Error Recovery
**Description:** Implement panic recovery and error handling

**Prerequisites:** Task 2.3

**Unlocks:** Task 3.2 (Infinite loop detection)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/errors.go` (create/extend)
- `pkg/bubbly/lifecycle_test.go` (extend)

**Type Safety:**
```go
var (
    ErrHookPanic        = errors.New("hook execution panicked")
    ErrCleanupFailed    = errors.New("cleanup function failed")
    ErrMaxUpdateDepth   = errors.New("max update depth exceeded")
)

func (lm *LifecycleManager) handleError(hookType string, err error)
func (lm *LifecycleManager) recoverFromPanic() error
```

**Tests:**
- [ ] Panics caught
- [ ] Component continues working
- [ ] Errors logged
- [ ] Stack trace captured
- [ ] Other hooks continue

**Estimated effort:** 3 hours

---

### Task 3.2: Infinite Loop Detection
**Description:** Detect and prevent infinite update loops

**Prerequisites:** Task 3.1

**Unlocks:** Task 4.1 (Cleanup integration)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)

**Type Safety:**
```go
const maxUpdateDepth = 100

func (lm *LifecycleManager) checkUpdateDepth() error
func (lm *LifecycleManager) resetUpdateCount()
```

**Tests:**
- [ ] Infinite loops detected
- [ ] Max depth enforced
- [ ] Error logged
- [ ] Execution stopped
- [ ] Component recovers

**Estimated effort:** 2 hours

---

## Phase 4: Auto-Cleanup Integration

### Task 4.1: Watcher Auto-Cleanup
**Description:** Integrate watcher cleanup with lifecycle

**Prerequisites:** Task 3.2

**Unlocks:** Task 4.2 (Event handler cleanup)

**Files:**
- `pkg/bubbly/context.go` (extend Watch method)
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)

**Type Safety:**
```go
type WatcherCleanup struct {
    cleanup   func()
    ref       *Ref[any]
    registered time.Time
}

func (ctx *Context) Watch(ref *Ref[any], callback func(any, any)) WatchCleanup
func (lm *LifecycleManager) registerWatcher(cleanup WatchCleanup)
func (lm *LifecycleManager) cleanupWatchers()
```

**Tests:**
- [ ] Watchers registered
- [ ] Auto-cleanup on unmount
- [ ] Multiple watchers cleaned
- [ ] No memory leaks
- [ ] Cleanup order correct

**Estimated effort:** 3 hours

---

### Task 4.2: Event Handler Auto-Cleanup
**Description:** Auto-cleanup event handlers on unmount

**Prerequisites:** Task 4.1

**Unlocks:** Task 5.1 (Integration)

**Files:**
- `pkg/bubbly/events.go` (extend)
- `pkg/bubbly/lifecycle.go` (extend)
- `pkg/bubbly/lifecycle_test.go` (extend)

**Type Safety:**
```go
type EventHandlerCleanup struct {
    handler   EventHandler
    eventName string
    component *componentImpl
}

func (lm *LifecycleManager) registerEventHandler(cleanup EventHandlerCleanup)
func (lm *LifecycleManager) cleanupEventHandlers()
```

**Tests:**
- [ ] Handlers registered
- [ ] Auto-cleanup works
- [ ] Handlers removed correctly
- [ ] No memory leaks
- [ ] Component continues working

**Estimated effort:** 2 hours

---

## Phase 5: Integration & Optimization

### Task 5.1: Component Integration
**Description:** Integrate lifecycle manager into component system

**Prerequisites:** Task 4.2

**Unlocks:** Task 5.2 (Optimization)

**Files:**
- `pkg/bubbly/component.go` (extend)
- `pkg/bubbly/component_test.go` (extend)

**Type Safety:**
```go
func (c *componentImpl) Init() tea.Cmd {
    c.lifecycle = newLifecycleManager(c)
    // Execute setup
    // Register hooks
}

func (c *componentImpl) View() string {
    if !c.lifecycle.mounted {
        c.lifecycle.executeMounted()
    }
    // Render template
}

func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle update
    c.lifecycle.executeUpdated()
    return c, cmd
}
```

**Tests:**
- [ ] Full lifecycle works
- [ ] Init integrates correctly
- [ ] View triggers onMounted
- [ ] Update triggers onUpdated
- [ ] Unmount works
- [ ] Children lifecycle managed

**Estimated effort:** 4 hours

---

### Task 5.2: Performance Optimization
**Description:** Optimize hook execution and dependency checking

**Prerequisites:** Task 5.1

**Unlocks:** Task 5.3 (Documentation)

**Files:**
- `pkg/bubbly/lifecycle.go` (optimize)
- Benchmarks (add/improve)

**Optimizations:**
- [ ] Dependency comparison optimization
- [ ] Hook pooling
- [ ] Lazy cleanup execution
- [ ] Reduce allocations
- [ ] Fast path for common cases

**Benchmarks:**
```go
BenchmarkHookRegister
BenchmarkHookExecute
BenchmarkDependencyCheck
BenchmarkCleanup
BenchmarkFullLifecycle
```

**Targets:**
- Hook registration: < 100ns
- Hook execution: < 500ns
- Dependency check: < 200ns
- Cleanup: < 1μs

**Estimated effort:** 3 hours

---

### Task 5.3: Documentation
**Description:** Complete API documentation and examples

**Prerequisites:** Task 5.2

**Unlocks:** Public API ready

**Files:**
- `pkg/bubbly/lifecycle.go` (godoc)
- `pkg/bubbly/lifecycle_examples_test.go`
- `docs/guides/lifecycle-hooks.md`

**Documentation:**
- [ ] Package overview
- [ ] Each hook documented
- [ ] Execution order explained
- [ ] Cleanup best practices
- [ ] 15+ examples
- [ ] Common patterns
- [ ] Troubleshooting guide
- [ ] Migration guide

**Examples:**
```go
func ExampleContext_OnMounted()
func ExampleContext_OnUpdated()
func ExampleContext_OnUpdated_withDependencies()
func ExampleContext_OnUnmounted()
func ExampleContext_OnCleanup()
func ExampleLifecycle_DataFetching()
func ExampleLifecycle_EventSubscription()
func ExampleLifecycle_Timer()
func ExampleLifecycle_FullCycle()
```

**Estimated effort:** 4 hours

---

## Phase 6: Testing & Validation

### Task 6.1: Integration Tests
**Description:** Test full lifecycle integration with components

**Prerequisites:** All implementation tasks

**Unlocks:** None (validation)

**Files:**
- `tests/integration/lifecycle_test.go`

**Tests:**
- [ ] Full lifecycle (mount → update → unmount)
- [ ] Nested components
- [ ] Multiple hooks
- [ ] Error recovery
- [ ] Auto-cleanup verification
- [ ] Performance acceptable

**Estimated effort:** 4 hours

---

### Task 6.2: Example Applications
**Description:** Create example apps demonstrating lifecycle hooks

**Prerequisites:** Task 6.1

**Unlocks:** Documentation examples

**Files:**
- `cmd/examples/lifecycle-basic/main.go`
- `cmd/examples/lifecycle-data-fetch/main.go`
- `cmd/examples/lifecycle-subscription/main.go`
- `cmd/examples/lifecycle-timer/main.go`

**Examples:**
- [ ] Basic hooks (mount, update, unmount)
- [ ] Data fetching on mount
- [ ] Event subscription with cleanup
- [ ] Timer/interval management
- [ ] Conditional updates

**Estimated effort:** 4 hours

---

### Task 6.3: Memory Leak Testing
**Description:** Verify no memory leaks from lifecycle system

**Prerequisites:** Task 6.2

**Unlocks:** Production readiness

**Files:**
- `tests/leak_test.go`

**Tests:**
- [ ] Long-running component test
- [ ] Repeated mount/unmount
- [ ] Watcher cleanup verification
- [ ] Event handler cleanup verification
- [ ] Memory profiling
- [ ] No goroutine leaks

**Estimated effort:** 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01 & 02)
    ↓
Phase 1: Foundation
    ├─> Task 1.1: Lifecycle manager structure
    ├─> Task 1.2: Hook registration
    └─> Task 1.3: State management
    ↓
Phase 2: Execution
    ├─> Task 2.1: onMounted execution
    ├─> Task 2.2: onUpdated execution
    └─> Task 2.3: onUnmounted execution
    ↓
Phase 3: Safety
    ├─> Task 3.1: Error recovery
    └─> Task 3.2: Infinite loop detection
    ↓
Phase 4: Auto-Cleanup
    ├─> Task 4.1: Watcher cleanup
    └─> Task 4.2: Event handler cleanup
    ↓
Phase 5: Integration
    ├─> Task 5.1: Component integration
    ├─> Task 5.2: Performance optimization
    └─> Task 5.3: Documentation
    ↓
Phase 6: Validation
    ├─> Task 6.1: Integration tests
    ├─> Task 6.2: Example apps
    └─> Task 6.3: Memory leak testing
    ↓
Unlocks: 04-composition-api
```

---

## Validation Checklist

### Code Quality
- [ ] All types strictly typed
- [ ] All public APIs documented
- [ ] All tests pass
- [ ] Race detector passes
- [ ] Linter passes
- [ ] Test coverage > 80%

### Functionality
- [ ] Hook registration works
- [ ] Hook execution order correct
- [ ] Dependencies tracked correctly
- [ ] Cleanup guaranteed
- [ ] Error handling works
- [ ] Auto-cleanup works
- [ ] No memory leaks

### Performance
- [ ] Hook registration < 100ns
- [ ] Hook execution < 500ns
- [ ] Dependency check < 200ns
- [ ] No performance regression
- [ ] Acceptable overhead

### Documentation
- [ ] README.md complete
- [ ] All hooks documented
- [ ] 15+ examples
- [ ] Best practices guide
- [ ] Troubleshooting guide
- [ ] Migration guide

### Integration
- [ ] Works with component system
- [ ] Works with reactivity system
- [ ] Ready for composition API
- [ ] Children lifecycle managed
- [ ] Bubbletea compatible

---

## Time Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Foundation | 3 | 7 hours |
| Phase 2: Execution | 3 | 10 hours |
| Phase 3: Safety | 2 | 5 hours |
| Phase 4: Auto-Cleanup | 2 | 5 hours |
| Phase 5: Integration | 3 | 11 hours |
| Phase 6: Validation | 3 | 11 hours |
| **Total** | **16 tasks** | **49 hours (~1.2 weeks)** |

---

## Development Order

### Week 1: Core Implementation
- Days 1-2: Phase 1 & 2 (Foundation & Execution)
- Days 3-4: Phase 3 & 4 (Safety & Auto-Cleanup)
- Day 5: Phase 5 (Integration start)

### Week 2: Polish & Validation
- Day 1: Phase 5 (Integration complete)
- Days 2-3: Phase 6 (Validation)
- Day 4: Documentation polish
- Day 5: Final review and examples

---

## Success Criteria

✅ **Definition of Done:**
1. All tests pass with > 80% coverage
2. Race detector shows no issues
3. No memory leaks in long-running tests
4. Benchmarks meet performance targets
5. Complete documentation with examples
6. Integration tests demonstrate full lifecycle
7. Example applications work correctly
8. Ready for composition API integration

✅ **Ready for Next Features:**
- Composition API can use lifecycle hooks
- Composables can register hooks
- Hooks work in all component types
- Clean integration with existing features

---

## Risk Mitigation

### Risk: Memory Leaks
**Mitigation:**
- Comprehensive leak testing
- Auto-cleanup for all resources
- Defer patterns for guaranteed cleanup
- Regular profiling

### Risk: Performance Overhead
**Mitigation:**
- Benchmarking from start
- Optimize hot paths
- Pool allocations
- Fast paths for common cases

### Risk: Complex Integration
**Mitigation:**
- Incremental integration
- Test each integration point
- Clear documentation
- Examples for all patterns

### Risk: Error Handling Edge Cases
**Mitigation:**
- Comprehensive error tests
- Panic recovery everywhere
- Clear error messages
- Component resilience

---

## Notes

### Design Decisions
- Hooks registered in Setup (not in template)
- Execution order: registration order
- Dependencies: explicit, not automatic
- Cleanup: reverse order
- Errors: caught and logged, don't crash

### Trade-offs
- **Simplicity vs Features:** Start with core hooks only
- **Performance vs Safety:** Prioritize safety with recovery
- **Auto vs Manual:** Auto-cleanup where possible
- **Explicit vs Implicit:** Explicit dependencies over magic

### Future Enhancements
- Async hooks (Promise-like)
- Hook middleware/interceptors
- Dev tools integration
- Performance monitoring
- Error boundaries
