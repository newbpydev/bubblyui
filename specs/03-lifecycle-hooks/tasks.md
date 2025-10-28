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

### Task 1.3: Lifecycle State Management ✅ COMPLETE
**Description:** Implement state tracking (mounted, unmounting, etc.)

**Prerequisites:** Task 1.2 ✅

**Unlocks:** Task 2.1 (Hook execution)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend) ✅
- `pkg/bubbly/lifecycle_test.go` (extend) ✅

**Type Safety:**
```go
func (lm *LifecycleManager) IsMounted() bool
func (lm *LifecycleManager) IsUnmounting() bool
func (lm *LifecycleManager) setMounted(mounted bool)
func (lm *LifecycleManager) setUnmounting(unmounting bool)
```

**Tests:**
- [x] State transitions correct
- [x] State queries work
- [x] Thread-safe state access
- [x] State persists correctly

**Implementation Notes:**
- Added `stateMu sync.RWMutex` to LifecycleManager for thread-safe state access
- Implemented `IsMounted()` with RLock for thread-safe reads
- Implemented `IsUnmounting()` with RLock for thread-safe reads
- Implemented `setMounted(bool)` with Lock for thread-safe writes
- Implemented `setUnmounting(bool)` with Lock for thread-safe writes
- Added 5 comprehensive test functions with table-driven tests
- All tests pass with race detector
- Coverage: 94.1% (exceeds 80% requirement)
- Linter clean (no warnings)
- Code formatted with gofmt

**Key Implementation Details:**
- Uses RWMutex for read-heavy access pattern (state queries more frequent than state changes)
- State queries use RLock (multiple concurrent readers allowed)
- State setters use Lock (exclusive write access)
- Thread-safe concurrent access verified with race detector
- State persistence verified across multiple queries
- Supports full state transition lifecycle: unmounted → mounted → unmounting

**Thread-Safety Pattern:**
```go
// Read operations use RLock
func (lm *LifecycleManager) IsMounted() bool {
    lm.stateMu.RLock()
    defer lm.stateMu.RUnlock()
    return lm.mounted
}

// Write operations use Lock
func (lm *LifecycleManager) setMounted(mounted bool) {
    lm.stateMu.Lock()
    defer lm.stateMu.Unlock()
    lm.mounted = mounted
}
```

**Estimated effort:** 2 hours ✅ (Actual: ~1.5 hours)

---

## Phase 2: Hook Execution

### Task 2.1: onMounted Execution ✅ COMPLETE
**Description:** Implement onMounted hook execution

**Prerequisites:** Task 1.3 ✅

**Unlocks:** Task 2.2 (onUpdated)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend) ✅
- `pkg/bubbly/lifecycle_test.go` (extend) ✅
- `pkg/bubbly/component.go` (integrate) ✅

**Type Safety:**
```go
func (lm *LifecycleManager) executeMounted()
func (lm *LifecycleManager) executeHooks(hookType string)
func (lm *LifecycleManager) safeExecuteHook(hookType string, hook lifecycleHook)
```

**Tests:**
- [x] Hooks execute after mount
- [x] Execution order correct
- [x] Only executes once
- [x] Integration with Component.View()
- [x] Multiple hooks work
- [x] Panic recovery works

**Implementation Notes:**
- Added `executeMounted()` method to LifecycleManager
  - Checks if already mounted (early return if true)
  - Sets mounted state to true before executing hooks
  - Calls executeHooks("mounted") to run all registered hooks
- Added `executeHooks(hookType string)` helper method
  - Iterates through hooks in registration order
  - Calls safeExecuteHook for each hook
  - Handles empty/non-existent hook arrays gracefully
- Added `safeExecuteHook(hookType, hook)` with panic recovery
  - Uses defer/recover pattern to catch panics
  - Silently recovers to prevent one hook from crashing others
  - All hooks are attempted even if some panic
- Integrated with Component.View()
  - Checks if lifecycle manager exists and component not mounted
  - Calls executeMounted() on first View() call
  - Ensures hooks execute after component is ready but before template renders
- Added 6 comprehensive test functions with table-driven tests:
  - TestLifecycleManager_ExecuteMounted (3 test cases)
  - TestLifecycleManager_ExecuteMounted_OnlyOnce (1 test case)
  - TestLifecycleManager_ExecuteMounted_Order (1 test case)
  - TestLifecycleManager_ExecuteMounted_PanicRecovery (3 test cases)
  - TestComponent_View_TriggersMounted (1 test case)
  - TestComponent_View_OnlyTriggersOnce (1 test case)
- All tests pass with race detector
- Coverage: 100% for new methods, 93.9% overall (exceeds 80% requirement)
- Linter clean (no warnings)
- Code formatted with gofmt

**Key Implementation Details:**
- executeMounted() is idempotent - safe to call multiple times
- Hooks execute in registration order (guaranteed by slice iteration)
- Panic recovery ensures component resilience
- Thread-safe state checks using existing IsMounted() method
- Integration point: Component.View() first call triggers execution
- No blocking operations - all hooks execute synchronously

**Estimated effort:** 3 hours ✅ (Actual: ~2 hours)

---

### Task 2.2: onUpdated Execution ✅ COMPLETE
**Description:** Implement onUpdated hook execution with dependency tracking

**Prerequisites:** Task 2.1 ✅

**Unlocks:** Task 2.3 (onUnmounted)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend) ✅
- `pkg/bubbly/lifecycle_test.go` (extend) ✅
- `pkg/bubbly/component.go` (integrate) ✅

**Type Safety:**
```go
func (lm *LifecycleManager) executeUpdated()
func (lm *LifecycleManager) shouldExecuteHook(hook *lifecycleHook) bool
func (lm *LifecycleManager) updateLastValues(hook *lifecycleHook)
```

**Tests:**
- [x] Hooks execute on update
- [x] Dependencies tracked correctly
- [x] Only runs when deps change
- [x] No deps: runs every time
- [x] Multiple dependencies work
- [x] Execution order preserved
- [x] Panic recovery works

**Implementation Notes:**
- Added `executeUpdated()` method to LifecycleManager
  - Checks if component is mounted (early return if not)
  - Iterates through "updated" hooks in registration order
  - Calls shouldExecuteHook() to check dependencies
  - Executes hook with safeExecuteHook() for panic recovery
  - Updates lastValues after successful execution
- Added `shouldExecuteHook(hook *lifecycleHook) bool` helper method
  - Returns true if hook has no dependencies (always execute)
  - Compares current dependency values with lastValues using deepEqual()
  - Returns true if any dependency changed
  - Returns false if all dependencies unchanged
- Added `updateLastValues(hook *lifecycleHook)` helper method
  - Updates lastValues slice with current dependency values
  - Called after hook execution for next comparison
- Integrated with Component.Update() method
  - Calls executeUpdated() after child component updates
  - Ensures state changes from children reflected before hook execution
  - Handles both components with and without children
- Added 5 comprehensive test functions with table-driven tests:
  - TestLifecycleManager_ExecuteUpdated (3 test cases)
  - TestLifecycleManager_ExecuteUpdated_WithDependencies (2 test cases)
  - TestLifecycleManager_ExecuteUpdated_MultipleDependencies (4 test cases)
  - TestLifecycleManager_ExecuteUpdated_Order (1 test case)
  - TestLifecycleManager_ExecuteUpdated_PanicRecovery (2 test cases)
- All tests pass with race detector
- Coverage: 94.2% overall, 91.7%-100% for new methods (exceeds 80% requirement)
- Linter clean (go vet passes)
- Code formatted with gofmt

**Key Implementation Details:**
- Uses existing deepEqual() function from deep.go for value comparison
- Dependency tracking uses reflect.DeepEqual internally
- Hook execution order guaranteed by slice iteration
- Panic recovery ensures component resilience
- Thread-safe state checks using existing IsMounted() method
- No blocking operations - all hooks execute synchronously
- lastValues updated only after successful execution
- Works with any type through *Ref[any] interface

**Dependency Tracking Logic:**
- No dependencies (empty slice): hook runs on every update
- With dependencies: hook runs only when at least one dependency changes
- Uses reflect.DeepEqual for deep value comparison
- Captures initial values during hook registration (in Context.OnUpdated)
- Updates lastValues after each execution for next comparison
- Supports multiple dependencies with OR logic (any change triggers execution)

**Integration Points:**
- Component.Update() calls executeUpdated() after processing messages
- Executes after child updates to reflect state changes from children
- Only executes if component is mounted (checked via IsMounted())
- Works seamlessly with existing lifecycle system

**Estimated effort:** 4 hours ✅ (Actual: ~2.5 hours)

---

### Task 2.3: onUnmounted Execution ✅ COMPLETE
**Description:** Implement onUnmounted hook execution and cleanup

**Prerequisites:** Task 2.2 ✅

**Unlocks:** Task 3.1 (Error handling)

**Files:**
- `pkg/bubbly/lifecycle.go` (extend) ✅
- `pkg/bubbly/lifecycle_test.go` (extend) ✅
- `pkg/bubbly/component.go` (integrate Unmount) ✅

**Type Safety:**
```go
func (lm *LifecycleManager) executeUnmounted()
func (lm *LifecycleManager) executeCleanups()
func (lm *LifecycleManager) safeExecuteCleanup(cleanup CleanupFunc)
func (c *componentImpl) Unmount()
```

**Tests:**
- [x] onUnmounted hooks execute
- [x] Cleanup functions execute
- [x] Reverse order execution (LIFO)
- [x] Children unmounted recursively
- [x] Cleanup guaranteed (panic recovery)
- [x] Only executes once
- [x] Execution order preserved
- [x] Panic recovery works

**Implementation Notes:**
- Added `executeUnmounted()` method to LifecycleManager
  - Checks if already unmounting (early return if true)
  - Sets unmounting state to true before executing
  - Executes all "unmounted" hooks in registration order
  - Calls executeCleanups() to run cleanup functions
  - Uses existing executeHooks() for hook execution
- Added `executeCleanups()` method to LifecycleManager
  - Iterates through cleanups in reverse order (LIFO)
  - Executes each cleanup with safeExecuteCleanup()
  - Guarantees all cleanups are attempted even if some panic
  - LIFO ensures proper resource unwinding
- Added `safeExecuteCleanup(cleanup CleanupFunc)` helper method
  - Uses defer/recover to catch panics
  - Logs panic information (silently for tests)
  - Allows execution to continue after panic
  - Mirrors safeExecuteHook pattern
- Added `Unmount()` method to componentImpl
  - Executes lifecycle cleanup (onUnmounted + cleanups)
  - Recursively unmounts all child components
  - Type asserts children to *componentImpl for Unmount call
  - Ensures parent cleanup runs before children unmount
- Added 7 comprehensive test functions with table-driven tests:
  - TestLifecycleManager_ExecuteUnmounted (3 test cases)
  - TestLifecycleManager_ExecuteUnmounted_OnlyOnce (2 test cases)
  - TestLifecycleManager_ExecuteUnmounted_Order (1 test case)
  - TestLifecycleManager_ExecuteUnmounted_PanicRecovery (2 test cases)
  - TestLifecycleManager_ExecuteCleanups (3 test cases)
  - TestLifecycleManager_ExecuteCleanups_ReverseOrder (1 test case)
  - TestLifecycleManager_ExecuteCleanups_PanicRecovery (2 test cases)
- All tests pass with race detector
- Coverage: 93.1% overall, 100% for new methods (exceeds 80% requirement)
- Linter clean (go vet passes)
- Code formatted with gofmt

**Key Implementation Details:**
- Unmounting is idempotent - executeUnmounted() only runs once
- Uses existing IsUnmounting() and setUnmounting() for thread-safe state
- Cleanup functions execute in LIFO order (reverse registration)
- onUnmounted hooks execute before cleanup functions
- Parent components unmount before children
- Panic recovery ensures all cleanups attempt execution
- No blocking operations - all execution is synchronous
- Reuses existing safeExecuteHook pattern for consistency

**Execution Order:**
1. Check if already unmounting (return if true)
2. Set unmounting flag to true
3. Execute all onUnmounted hooks (registration order)
4. Execute all cleanup functions (reverse order - LIFO)
5. Recursively unmount children (if component has Unmount method)

**LIFO Cleanup Rationale:**
- Resources released in reverse acquisition order
- Dependencies cleaned up before dependents
- Proper unwinding of nested resources
- Matches common cleanup patterns (e.g., defer in Go)
- Example: If A registers cleanup, then B registers cleanup:
  - B's cleanup runs first (most recent)
  - A's cleanup runs second (oldest)

**Integration Points:**
- Component.Unmount() is the public API for cleanup
- Calls lifecycle.executeUnmounted() internally
- Recursively unmounts child components
- Can be called manually when component is removed
- Will be integrated with Bubbletea lifecycle in future tasks

**Panic Recovery:**
- Both hooks and cleanups use panic recovery
- One failing hook/cleanup doesn't prevent others
- Panics are caught and logged (silently in tests)
- Execution continues after panic
- All registered hooks/cleanups are attempted

**Thread Safety:**
- Uses existing stateMu RWMutex for unmounting flag
- IsUnmounting() and setUnmounting() are thread-safe
- Cleanup execution is synchronous (no goroutines)
- Safe to call from multiple goroutines (idempotent)

**Estimated effort:** 3 hours ✅ (Actual: ~2 hours)

---

## Phase 3: Error Handling & Safety

### Task 3.1: Error Recovery ✅ COMPLETE
**Description:** Implement panic recovery and error handling

**Prerequisites:** Task 2.3 ✅

**Unlocks:** Task 3.2 (Infinite loop detection)

**Files:**
- `pkg/bubbly/lifecycle.go` ✅ (panic recovery already implemented with observability)
- `pkg/bubbly/lifecycle_errors.go` ✅ (created with sentinel errors)
- `pkg/bubbly/lifecycle_test.go` ✅ (panic recovery tests already exist)

**Type Safety:**
```go
var (
    ErrHookPanic        = errors.New("hook execution panicked")
    ErrCleanupFailed    = errors.New("cleanup function failed")
    ErrMaxUpdateDepth   = errors.New("max update depth exceeded")
)
```

**Tests:**
- [x] Panics caught (existing tests: TestLifecycleManager_ExecuteMounted_PanicRecovery)
- [x] Component continues working (verified in panic recovery tests)
- [x] Errors reported to observability system (integrated in safeExecuteHook/safeExecuteCleanup)
- [x] Stack trace captured (debug.Stack() in observability reporting)
- [x] Other hooks continue (verified in panic recovery tests)

**Implementation Notes:**
- Created `lifecycle_errors.go` with three sentinel error types
- Panic recovery was already implemented in Tasks 2.1-2.3 with full observability integration
- `safeExecuteHook()` uses defer/recover and reports to observability.GetErrorReporter()
- `safeExecuteCleanup()` uses defer/recover and reports to observability.GetErrorReporter()
- Stack traces captured via `debug.Stack()` and included in error context
- Error context includes: component name, ID, event name, timestamp, tags, and extra data
- All tests pass with race detector
- Code formatted with gofmt
- Linter clean (go vet passes)
- Follows Go best practices for sentinel errors (errors.New)
- Follows CRITICAL RULE: Production Error Reporting (observability integration mandatory)

**Key Implementation Details:**
- Error types are sentinel errors for use with errors.Is()
- Observability integration provides pluggable error reporters (Sentry, Console, custom)
- Zero overhead when no reporter configured (GetErrorReporter() returns nil)
- Thread-safe error reporting
- Rich error context for debugging production issues

**Estimated effort:** 3 hours ✅ (Actual: ~1 hour - panic recovery already existed, only needed error type definitions)

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
