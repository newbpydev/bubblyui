# Implementation Tasks: Testing Utilities

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] 01-reactivity-system completed (State testing)
- [x] 02-component-model completed (Component testing)
- [x] 03-lifecycle-hooks completed (Lifecycle testing)
- [ ] testify library integrated
- [ ] Go testing conventions established

---

## Phase 1: Test Harness Foundation (4 tasks, 12 hours)

### Task 1.1: Test Harness Core
**Description**: Main test harness for component mounting

**Prerequisites**: None

**Unlocks**: Task 1.2 (Component Mounting)

**Files**:
- `pkg/bubbly/testutil/harness.go`
- `pkg/bubbly/testutil/harness_test.go`

**Type Safety**:
```go
type TestHarness struct {
    t         *testing.T
    component Component
    refs      map[string]*Ref[interface{}]
    events    *EventTracker
    cleanup   []func()
}

func NewHarness(t *testing.T, opts ...HarnessOption) *TestHarness
func (h *TestHarness) Cleanup()
```

**Tests**:
- [ ] Harness creation
- [ ] Cleanup registration
- [ ] Automatic cleanup on test end
- [ ] Options pattern works
- [ ] Thread-safe operations

**Estimated Effort**: 3 hours

---

### Task 1.2: Component Mounting
**Description**: Mount components in test environment

**Prerequisites**: Task 1.1

**Unlocks**: Task 1.3 (State Extraction)

**Files**:
- `pkg/bubbly/testutil/mount.go`
- `pkg/bubbly/testutil/mount_test.go`

**Type Safety**:
```go
type ComponentTest struct {
    harness   *TestHarness
    component Component
    state     *StateInspector
    events    *EventInspector
}

func (h *TestHarness) Mount(component Component, props ...interface{}) *ComponentTest
func (ct *ComponentTest) Unmount()
```

**Tests**:
- [ ] Components mount correctly
- [ ] Init() called automatically
- [ ] Props applied
- [ ] State accessible
- [ ] Cleanup works

**Estimated Effort**: 3 hours

---

### Task 1.3: State Extraction & Inspection
**Description**: Extract and inspect component state

**Prerequisites**: Task 1.2

**Unlocks**: Task 1.4 (Hook Installation)

**Files**:
- `pkg/bubbly/testutil/state_inspector.go`
- `pkg/bubbly/testutil/state_inspector_test.go`

**Type Safety**:
```go
type StateInspector struct {
    refs     map[string]*Ref[interface{}]
    computed map[string]*Computed[interface{}]
}

func (si *StateInspector) GetRef(name string) *Ref[interface{}]
func (si *StateInspector) GetRefValue(name string) interface{}
func (si *StateInspector) SetRefValue(name string, value interface{})
```

**Tests**:
- [ ] Refs extracted correctly
- [ ] Computed values extracted
- [ ] GetRef returns correct ref
- [ ] SetRefValue updates state
- [ ] Error on missing ref

**Estimated Effort**: 3 hours

---

### Task 1.4: Hook Installation
**Description**: Install test hooks into components

**Prerequisites**: Task 1.3

**Unlocks**: Task 2.1 (State Assertions)

**Files**:
- `pkg/bubbly/testutil/hooks.go`
- `pkg/bubbly/testutil/hooks_test.go`

**Type Safety**:
```go
type TestHooks struct {
    onStateChange func(string, interface{})
    onEvent       func(string, interface{})
    onUpdate      func()
}

func (h *TestHarness) installHooks(component Component)
func (h *TestHarness) removeHooks()
```

**Tests**:
- [ ] Hooks install correctly
- [ ] State changes tracked
- [ ] Events tracked
- [ ] Updates tracked
- [ ] Removal works

**Estimated Effort**: 3 hours

---

## Phase 2: Assertions & Matchers (5 tasks, 15 hours)

### Task 2.1: State Assertions
**Description**: Type-safe state assertion helpers

**Prerequisites**: Task 1.4

**Unlocks**: Task 2.2 (Event Assertions)

**Files**:
- `pkg/bubbly/testutil/assertions_state.go`
- `pkg/bubbly/testutil/assertions_state_test.go`

**Type Safety**:
```go
func (ct *ComponentTest) AssertRefEquals(name string, expected interface{})
func (ct *ComponentTest) AssertRefChanged(name string, initial interface{})
func (ct *ComponentTest) AssertRefType(name string, expectedType reflect.Type)
```

**Tests**:
- [ ] AssertRefEquals passes/fails correctly
- [ ] AssertRefChanged detects changes
- [ ] AssertRefType validates types
- [ ] Clear error messages
- [ ] Works with testify

**Estimated Effort**: 3 hours

---

### Task 2.2: Event Assertions
**Description**: Event tracking and assertion helpers

**Prerequisites**: Task 2.1

**Unlocks**: Task 2.3 (Render Assertions)

**Files**:
- `pkg/bubbly/testutil/assertions_events.go`
- `pkg/bubbly/testutil/assertions_events_test.go`

**Type Safety**:
```go
func (ct *ComponentTest) AssertEventFired(name string)
func (ct *ComponentTest) AssertEventNotFired(name string)
func (ct *ComponentTest) AssertEventPayload(name string, expected interface{})
func (ct *ComponentTest) AssertEventCount(name string, count int)
```

**Tests**:
- [ ] AssertEventFired works
- [ ] AssertEventNotFired works
- [ ] Payload assertions work
- [ ] Count assertions work
- [ ] Multiple events tracked

**Estimated Effort**: 3 hours

---

### Task 2.3: Render Assertions
**Description**: Assert on render output

**Prerequisites**: Task 2.2

**Unlocks**: Task 2.4 (Custom Matchers)

**Files**:
- `pkg/bubbly/testutil/assertions_render.go`
- `pkg/bubbly/testutil/assertions_render_test.go`

**Type Safety**:
```go
func (ct *ComponentTest) AssertRenderContains(substring string)
func (ct *ComponentTest) AssertRenderEquals(expected string)
func (ct *ComponentTest) AssertRenderMatches(pattern *regexp.Regexp)
```

**Tests**:
- [ ] AssertRenderContains works
- [ ] AssertRenderEquals works
- [ ] Regex matching works
- [ ] Whitespace handling
- [ ] Error messages clear

**Estimated Effort**: 2 hours

---

### Task 2.4: Custom Matchers
**Description**: Framework for custom assertion matchers

**Prerequisites**: Task 2.3

**Unlocks**: Task 2.5 (Async Assertions)

**Files**:
- `pkg/bubbly/testutil/matchers.go`
- `pkg/bubbly/testutil/matchers_test.go`

**Type Safety**:
```go
type Matcher interface {
    Match(actual interface{}) (bool, error)
    FailureMessage(actual interface{}) string
}

func (ct *ComponentTest) AssertThat(actual interface{}, matcher Matcher)

// Common matchers
func BeEmpty() Matcher
func HaveLength(expected int) Matcher
func BeNil() Matcher
```

**Tests**:
- [ ] Custom matchers work
- [ ] Built-in matchers work
- [ ] Failure messages clear
- [ ] Composable matchers
- [ ] Type-safe matching

**Estimated Effort**: 4 hours

---

### Task 2.5: Async Assertions
**Description**: Wait and assert on async conditions

**Prerequisites**: Task 2.4

**Unlocks**: Task 3.1 (Event Simulator)

**Files**:
- `pkg/bubbly/testutil/async_assertions.go`
- `pkg/bubbly/testutil/async_assertions_test.go`

**Type Safety**:
```go
type WaitOptions struct {
    Timeout  time.Duration
    Interval time.Duration
    Message  string
}

func WaitFor(t *testing.T, condition func() bool, opts WaitOptions)
func (ct *ComponentTest) WaitForRef(name string, expected interface{}, timeout time.Duration)
func (ct *ComponentTest) WaitForEvent(name string, timeout time.Duration)
```

**Tests**:
- [ ] WaitFor polls correctly
- [ ] Timeout works
- [ ] WaitForRef works
- [ ] WaitForEvent works
- [ ] Error messages include state

**Estimated Effort**: 3 hours

---

## Phase 3: Event & Message Simulation (3 tasks, 9 hours)

### Task 3.1: Event Simulator
**Description**: Simulate event emission

**Prerequisites**: Task 2.5

**Unlocks**: Task 3.2 (Message Simulator)

**Files**:
- `pkg/bubbly/testutil/event_simulator.go`
- `pkg/bubbly/testutil/event_simulator_test.go`

**Type Safety**:
```go
func (ct *ComponentTest) Emit(name string, payload interface{})
func (ct *ComponentTest) EmitAndWait(name string, payload interface{}, timeout time.Duration)
func (ct *ComponentTest) EmitMultiple(events []Event)
```

**Tests**:
- [ ] Emit works
- [ ] EmitAndWait waits correctly
- [ ] Multiple events emit in order
- [ ] Event handlers execute
- [ ] State updates after emit

**Estimated Effort**: 3 hours

---

### Task 3.2: Message Simulator
**Description**: Simulate Bubbletea messages

**Prerequisites**: Task 3.1

**Unlocks**: Task 3.3 (Event Tracker)

**Files**:
- `pkg/bubbly/testutil/message_simulator.go`
- `pkg/bubbly/testutil/message_simulator_test.go`

**Type Safety**:
```go
func (ct *ComponentTest) SendMessage(msg tea.Msg) tea.Cmd
func (ct *ComponentTest) SendKey(key string) tea.Cmd
func (ct *ComponentTest) SendMouseClick(x, y int) tea.Cmd
```

**Tests**:
- [ ] Messages sent correctly
- [ ] Update() called
- [ ] Commands returned
- [ ] KeyMsg simulation works
- [ ] MouseMsg simulation works

**Estimated Effort**: 3 hours

---

### Task 3.3: Event Tracker
**Description**: Track emitted events for inspection

**Prerequisites**: Task 3.2

**Unlocks**: Task 4.1 (Mock Ref)

**Files**:
- `pkg/bubbly/testutil/event_tracker.go`
- `pkg/bubbly/testutil/event_tracker_test.go`

**Type Safety**:
```go
type EventTracker struct {
    events []EmittedEvent
    mu     sync.RWMutex
}

type EmittedEvent struct {
    Name      string
    Payload   interface{}
    Timestamp time.Time
    Source    string
}

func (et *EventTracker) Track(name, payload, source)
func (et *EventTracker) GetEvents(name string) []EmittedEvent
func (et *EventTracker) WasFired(name string) bool
```

**Tests**:
- [ ] Events tracked correctly
- [ ] Retrieval works
- [ ] WasFired works
- [ ] Thread-safe
- [ ] Clear() works

**Estimated Effort**: 3 hours

---

## Phase 4: Mock System (5 tasks, 15 hours)

### Task 4.1: Mock Ref
**Description**: Mock ref implementation for testing

**Prerequisites**: Task 3.3

**Unlocks**: Task 4.2 (Mock Component)

**Files**:
- `pkg/bubbly/testutil/mock_ref.go`
- `pkg/bubbly/testutil/mock_ref_test.go`

**Type Safety**:
```go
type MockRef[T any] struct {
    value    T
    getCalls int
    setCalls int
    watchers []func(T)
}

func NewMockRef[T any](initial T) *MockRef[T]
func (mr *MockRef[T]) AssertGetCalled(t *testing.T, times int)
func (mr *MockRef[T]) AssertSetCalled(t *testing.T, times int)
```

**Tests**:
- [ ] MockRef implements Ref interface
- [ ] Get/Set tracking works
- [ ] Watchers work
- [ ] Assertions work
- [ ] Type-safe operations

**Estimated Effort**: 3 hours

---

### Task 4.2: Mock Component
**Description**: Mock component for testing

**Prerequisites**: Task 4.1

**Unlocks**: Task 4.3 (Mock Factory)

**Files**:
- `pkg/bubbly/testutil/mock_component.go`
- `pkg/bubbly/testutil/mock_component_test.go`

**Type Safety**:
```go
type MockComponent struct {
    name          string
    initCalled    bool
    updateCalls   int
    viewCalls     int
    unmountCalled bool
    viewOutput    string
}

func NewMockComponent(name string) *MockComponent
func (mc *MockComponent) AssertInitCalled(t *testing.T)
func (mc *MockComponent) AssertUpdateCalled(t *testing.T, times int)
```

**Tests**:
- [ ] Mock implements Component interface
- [ ] Method call tracking works
- [ ] Assertions work
- [ ] Configurable output
- [ ] Props support

**Estimated Effort**: 3 hours

---

### Task 4.3: Mock Factory
**Description**: Factory for creating mocks

**Prerequisites**: Task 4.2

**Unlocks**: Task 4.4 (Mock Router)

**Files**:
- `pkg/bubbly/testutil/mock_factory.go`
- `pkg/bubbly/testutil/mock_factory_test.go`

**Type Safety**:
```go
type MockFactory struct {
    mocks map[string]interface{}
}

func NewMockFactory() *MockFactory
func (mf *MockFactory) CreateMockRef[T any](name string, initial T) *MockRef[T]
func (mf *MockFactory) CreateMockComponent(name string) *MockComponent
```

**Tests**:
- [ ] Factory creates mocks
- [ ] Mock registration works
- [ ] Retrieval works
- [ ] Cleanup works
- [ ] Type-safe creation

**Estimated Effort**: 2 hours

---

### Task 4.4: Mock Router
**Description**: Mock router for route testing

**Prerequisites**: Task 4.3

**Unlocks**: Task 4.5 (Mock Commands)

**Files**:
- `pkg/bubbly/testutil/mock_router.go`
- `pkg/bubbly/testutil/mock_router_test.go`

**Type Safety**:
```go
type MockRouter struct {
    currentRoute *Route
    pushCalls    []NavigationTarget
    backCalls    int
}

func NewMockRouter() *MockRouter
func (mr *MockRouter) AssertPushed(t *testing.T, path string)
func (mr *MockRouter) AssertBackCalled(t *testing.T)
```

**Tests**:
- [ ] Mock implements Router interface
- [ ] Navigation tracking works
- [ ] Current route settable
- [ ] Assertions work
- [ ] Integration with components

**Estimated Effort**: 3 hours

---

### Task 4.5: Mock Commands
**Description**: Mock Bubbletea commands

**Prerequisites**: Task 4.4

**Unlocks**: Task 5.1 (Snapshot Manager)

**Files**:
- `pkg/bubbly/testutil/mock_commands.go`
- `pkg/bubbly/testutil/mock_commands_test.go`

**Type Safety**:
```go
type MockCommand struct {
    executed bool
    message  tea.Msg
    error    error
}

func NewMockCommand(msg tea.Msg) tea.Cmd
func NewMockCommandWithError(err error) tea.Cmd
func AssertCommandExecuted(t *testing.T, cmd tea.Cmd)
```

**Tests**:
- [ ] Mock commands work
- [ ] Execution tracked
- [ ] Messages returned
- [ ] Errors handled
- [ ] Assertions work

**Estimated Effort**: 4 hours

---

## Phase 5: Snapshot Testing (4 tasks, 12 hours)

### Task 5.1: Snapshot Manager
**Description**: Core snapshot testing functionality

**Prerequisites**: Task 4.5

**Unlocks**: Task 5.2 (Snapshot Diff)

**Files**:
- `pkg/bubbly/testutil/snapshot.go`
- `pkg/bubbly/testutil/snapshot_test.go`

**Type Safety**:
```go
type SnapshotManager struct {
    dir    string
    update bool
    mu     sync.Mutex
}

func NewSnapshotManager(testDir string, update bool) *SnapshotManager
func (sm *SnapshotManager) Match(t *testing.T, name, actual string)
```

**Tests**:
- [ ] Snapshot creation works
- [ ] Snapshot comparison works
- [ ] Update mode works
- [ ] File format correct
- [ ] Thread-safe operations

**Estimated Effort**: 4 hours

---

### Task 5.2: Snapshot Diff Generation
**Description**: Generate diffs for snapshot mismatches

**Prerequisites**: Task 5.1

**Unlocks**: Task 5.3 (Snapshot Helpers)

**Files**:
- `pkg/bubbly/testutil/snapshot_diff.go`
- `pkg/bubbly/testutil/snapshot_diff_test.go`

**Type Safety**:
```go
func generateDiff(expected, actual string) string
func highlightDiff(diff string) string
func formatForTerminal(diff string) string
```

**Tests**:
- [ ] Diff generation works
- [ ] Highlighting works
- [ ] Terminal formatting works
- [ ] Large diffs handled
- [ ] Readable output

**Estimated Effort**: 3 hours

---

### Task 5.3: Snapshot Helpers
**Description**: Convenience helpers for snapshot testing

**Prerequisites**: Task 5.2

**Unlocks**: Task 5.4 (Snapshot Normalization)

**Files**:
- `pkg/bubbly/testutil/snapshot_helpers.go`
- `pkg/bubbly/testutil/snapshot_helpers_test.go`

**Type Safety**:
```go
func MatchSnapshot(t *testing.T, actual string)
func MatchNamedSnapshot(t *testing.T, name, actual string)
func MatchComponentSnapshot(t *testing.T, component Component)
func UpdateSnapshots(t *testing.T) bool
```

**Tests**:
- [ ] MatchSnapshot works
- [ ] Named snapshots work
- [ ] Component snapshots work
- [ ] Update flag detection works
- [ ] Default naming works

**Estimated Effort**: 2 hours

---

### Task 5.4: Snapshot Normalization
**Description**: Normalize dynamic content in snapshots

**Prerequisites**: Task 5.3

**Unlocks**: Task 6.1 (Fixture Builder)

**Files**:
- `pkg/bubbly/testutil/snapshot_normalize.go`
- `pkg/bubbly/testutil/snapshot_normalize_test.go`

**Type Safety**:
```go
type Normalizer struct {
    patterns []NormalizePattern
}

type NormalizePattern struct {
    Pattern     *regexp.Regexp
    Replacement string
}

func (n *Normalizer) Normalize(content string) string
func NormalizeTimestamps(content string) string
func NormalizeUUIDs(content string) string
```

**Tests**:
- [ ] Timestamp normalization works
- [ ] UUID normalization works
- [ ] Custom patterns work
- [ ] Multiple normalizations work
- [ ] Performance acceptable

**Estimated Effort**: 3 hours

---

## Phase 6: Fixtures & Utilities (4 tasks, 12 hours)

### Task 6.1: Fixture Builder
**Description**: Builder for test fixtures

**Prerequisites**: Task 5.4

**Unlocks**: Task 6.2 (Data Factories)

**Files**:
- `pkg/bubbly/testutil/fixture.go`
- `pkg/bubbly/testutil/fixture_test.go`

**Type Safety**:
```go
type FixtureBuilder struct {
    props  map[string]interface{}
    state  map[string]interface{}
    events map[string]interface{}
}

func NewFixture() *FixtureBuilder
func (fb *FixtureBuilder) WithProp(key, value) *FixtureBuilder
func (fb *FixtureBuilder) WithState(key, value) *FixtureBuilder
func (fb *FixtureBuilder) Build(t, createFn) *ComponentTest
```

**Tests**:
- [ ] Fixture building works
- [ ] Props applied
- [ ] State set correctly
- [ ] Events emitted
- [ ] Fluent API works

**Estimated Effort**: 3 hours

---

### Task 6.2: Data Factories
**Description**: Factories for generating test data

**Prerequisites**: Task 6.1

**Unlocks**: Task 6.3 (Setup Helpers)

**Files**:
- `pkg/bubbly/testutil/factories.go`
- `pkg/bubbly/testutil/factories_test.go`

**Type Safety**:
```go
type DataFactory[T any] struct {
    generator func() T
}

func NewFactory[T any](generator func() T) *DataFactory[T]
func (df *DataFactory[T]) Generate() T
func (df *DataFactory[T]) GenerateN(n int) []T

// Common factories
func IntFactory(min, max int) *DataFactory[int]
func StringFactory(length int) *DataFactory[string]
```

**Tests**:
- [ ] Factory generation works
- [ ] Type-safe factories
- [ ] Multiple generation works
- [ ] Built-in factories work
- [ ] Custom factories work

**Estimated Effort**: 3 hours

---

### Task 6.3: Setup & Teardown Helpers
**Description**: Helpers for test setup and teardown

**Prerequisites**: Task 6.2

**Unlocks**: Task 6.4 (Test Isolation)

**Files**:
- `pkg/bubbly/testutil/setup.go`
- `pkg/bubbly/testutil/setup_test.go`

**Type Safety**:
```go
type TestSetup struct {
    setupFuncs    []func(*testing.T)
    teardownFuncs []func(*testing.T)
}

func NewTestSetup() *TestSetup
func (ts *TestSetup) AddSetup(fn func(*testing.T)) *TestSetup
func (ts *TestSetup) AddTeardown(fn func(*testing.T)) *TestSetup
func (ts *TestSetup) Run(t *testing.T, testFn func(*testing.T))
```

**Tests**:
- [ ] Setup functions execute
- [ ] Teardown functions execute
- [ ] Execution order correct
- [ ] Error handling works
- [ ] Integration with t.Cleanup

**Estimated Effort**: 3 hours

---

### Task 6.4: Test Isolation
**Description**: Ensure tests are isolated from each other

**Prerequisites**: Task 6.3

**Unlocks**: Task 7.1 (Documentation)

**Files**:
- `pkg/bubbly/testutil/isolation.go`
- `pkg/bubbly/testutil/isolation_test.go`

**Type Safety**:
```go
type TestIsolation struct {
    savedGlobals map[string]interface{}
}

func NewTestIsolation() *TestIsolation
func (ti *TestIsolation) Isolate(t *testing.T)
func (ti *TestIsolation) Restore()
```

**Tests**:
- [ ] Isolation works
- [ ] Globals saved/restored
- [ ] Tests don't interfere
- [ ] Cleanup automatic
- [ ] Parallel tests safe

**Estimated Effort**: 3 hours

---

## Phase 7: Documentation & Examples (3 tasks, 9 hours)

### Task 7.1: API Documentation
**Description**: Comprehensive godoc for testing utilities

**Prerequisites**: Task 6.4

**Unlocks**: Task 7.2 (Testing Guide)

**Files**:
- All package files (add/update godoc)

**Documentation**:
- Test harness API
- Assertion helpers API
- Mock utilities API
- Snapshot testing API
- Fixture system API

**Estimated Effort**: 2 hours

---

### Task 7.2: Testing Guide
**Description**: Complete testing documentation

**Prerequisites**: Task 7.1

**Unlocks**: Task 7.3 (Examples)

**Files**:
- `docs/testing/README.md`
- `docs/testing/quickstart.md`
- `docs/testing/assertions.md`
- `docs/testing/mocking.md`
- `docs/testing/snapshots.md`

**Content**:
- Getting started guide
- Assertion reference
- Mocking guide
- Snapshot testing guide
- Best practices
- TDD workflow

**Estimated Effort**: 4 hours

---

### Task 7.3: Example Test Suites
**Description**: Complete example test suites

**Prerequisites**: Task 7.2

**Unlocks**: Feature complete

**Files**:
- `cmd/examples/10-testing/counter_test.go`
- `cmd/examples/10-testing/todo_test.go`
- `cmd/examples/10-testing/form_test.go`
- `cmd/examples/10-testing/async_test.go`

**Examples**:
- Basic component tests
- Table-driven tests
- Snapshot tests
- Async tests
- Mock usage
- Fixture usage

**Estimated Effort**: 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01-03, testify)
    ↓
Phase 1: Foundation
    1.1 Harness → 1.2 Mounting → 1.3 State → 1.4 Hooks
    ↓
Phase 2: Assertions
    2.1 State Asserts → 2.2 Event Asserts → 2.3 Render Asserts → 2.4 Matchers → 2.5 Async
    ↓
Phase 3: Simulation
    3.1 Events → 3.2 Messages → 3.3 Tracker
    ↓
Phase 4: Mocks
    4.1 Mock Ref → 4.2 Mock Component → 4.3 Factory → 4.4 Router → 4.5 Commands
    ↓
Phase 5: Snapshots
    5.1 Manager → 5.2 Diff → 5.3 Helpers → 5.4 Normalization
    ↓
Phase 6: Fixtures
    6.1 Builder → 6.2 Factories → 6.3 Setup → 6.4 Isolation
    ↓
Phase 7: Documentation
    7.1 API Docs → 7.2 Guide → 7.3 Examples
```

---

## Validation Checklist

### Core Functionality
- [ ] Components mount in tests
- [ ] State accessible
- [ ] Events simulatable
- [ ] Assertions work
- [ ] Cleanup automatic

### Assertions
- [ ] State assertions accurate
- [ ] Event assertions accurate
- [ ] Render assertions accurate
- [ ] Custom matchers work
- [ ] Async assertions reliable

### Mocks
- [ ] Ref mocks work
- [ ] Component mocks work
- [ ] Router mocks work
- [ ] Command mocks work
- [ ] Easy to create

### Snapshots
- [ ] Snapshots create correctly
- [ ] Diffs clear
- [ ] Update works
- [ ] Normalization works
- [ ] Git-friendly

### Performance
- [ ] Setup < 1ms
- [ ] Execution fast
- [ ] Cleanup complete
- [ ] No memory leaks
- [ ] Parallel tests safe

---

## Estimated Total Effort

- Phase 1: 12 hours
- Phase 2: 15 hours
- Phase 3: 9 hours
- Phase 4: 15 hours
- Phase 5: 12 hours
- Phase 6: 12 hours
- Phase 7: 9 hours

**Total**: ~84 hours (approximately 2.5 weeks)

---

## Priority

**HIGH** - Critical for code quality, TDD workflow, and project maintainability.

**Timeline**: Implement alongside or immediately after Features 01-03, as testing is foundational for all future development.

**Unlocks**:
- Test-driven development
- High code quality
- Refactoring confidence
- Bug prevention
- Documentation through tests
- Community contributions (with tests)
