# Design Specification: Testing Utilities

## Component Hierarchy

```
Testing Framework
└── Test Utilities Package
    ├── Test Harness
    │   ├── Component Mounter
    │   ├── State Inspector
    │   └── Event Simulator
    ├── Assertion Helpers
    │   ├── State Assertions
    │   ├── Event Assertions
    │   └── Render Assertions
    ├── Mock System
    │   ├── Mock Factory
    │   ├── Ref Mocks
    │   └── Component Mocks
    ├── Snapshot Manager
    │   ├── Snapshot Writer
    │   ├── Snapshot Comparer
    │   └── Snapshot Updater
    └── Test Fixtures
        ├── Fixture Builder
        ├── Data Factories
        └── Setup Helpers
```

---

## Architecture Overview

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      Test Code                                │
│  (Developer's test functions using testing utilities)        │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                  Testing Utilities Framework                  │
├──────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐    ┌────────────────┐ │
│  │ Test Harness │───→│ Assertions   │←───│  Mocks         │ │
│  │              │    │              │    │                │ │
│  └──────┬───────┘    └──────────────┘    └────────────────┘ │
│         │                                                     │
│         ↓                                                     │
│  ┌──────────────────────────────────────────────────────┐   │
│  │        Component Under Test (Isolated)                │   │
│  └──────────────────────────────────────────────────────┘   │
└───────────────────────────────┬──────────────────────────────┘
                                │
┌───────────────────────────────┴──────────────────────────────┐
│                    Go Testing Framework                       │
│  (Built-in testing package + testify)                        │
└──────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Test Execution Flow

```
Test Function Starts
    ↓
Create Test Harness
    ↓
Mount Component
    ├─ Initialize state
    ├─ Execute setup
    └─ Install hooks
    ↓
Simulate Actions
    ├─ Emit events
    ├─ Update state
    └─ Send messages
    ↓
Assert Outcomes
    ├─ Check state
    ├─ Verify events
    └─ Compare render
    ↓
Cleanup
    ├─ Unmount component
    ├─ Remove hooks
    └─ Free resources
    ↓
Test Complete
```

---

## Type Definitions

### Core Types

```go
// TestHarness provides component testing environment
type TestHarness struct {
    t         *testing.T
    component Component
    refs      map[string]*Ref[interface{}]
    events    *EventTracker
    cleanup   []func()
}

// ComponentTest wraps a mounted component for testing
type ComponentTest struct {
    harness   *TestHarness
    component Component
    state     *StateInspector
    events    *EventInspector
}

// StateInspector provides state access and assertions
type StateInspector struct {
    refs     map[string]*Ref[interface{}]
    computed map[string]*Computed[interface{}]
    watchers map[string]*Watcher
}

// EventInspector tracks emitted events
type EventInspector struct {
    events   []EmittedEvent
    handlers map[string][]HandlerCall
}

// EmittedEvent captures event emission
type EmittedEvent struct {
    Name      string
    Payload   interface{}
    Timestamp time.Time
    Source    string
}

// MockRef is a mock ref for testing
type MockRef[T any] struct {
    value     T
    getCalls  int
    setCalls  int
    watchers  []func(T)
}

// SnapshotManager handles snapshot testing
type SnapshotManager struct {
    dir     string
    update  bool
    mu      sync.Mutex
}
```

---

## Test Harness Architecture

### Harness Creation

```go
func NewHarness(t *testing.T, opts ...HarnessOption) *TestHarness {
    h := &TestHarness{
        t:       t,
        refs:    make(map[string]*Ref[interface{}]),
        events:  NewEventTracker(),
        cleanup: []func(){},
    }
    
    for _, opt := range opts {
        opt(h)
    }
    
    // Register cleanup
    t.Cleanup(func() {
        h.Cleanup()
    })
    
    return h
}

type HarnessOption func(*TestHarness)

func WithIsolation() HarnessOption {
    return func(h *TestHarness) {
        // Isolate from global state
    }
}

func WithTimeout(d time.Duration) HarnessOption {
    return func(h *TestHarness) {
        // Set test timeout
    }
}
```

### Component Mounting

```go
func (h *TestHarness) Mount(component Component, props ...interface{}) *ComponentTest {
    // Initialize component
    component.Init()
    
    // Install hooks for testing
    h.installHooks(component)
    
    // Extract refs and state
    h.extractState(component)
    
    // Create test wrapper
    test := &ComponentTest{
        harness:   h,
        component: component,
        state:     NewStateInspector(h.refs),
        events:    NewEventInspector(h.events),
    }
    
    // Register cleanup
    h.cleanup = append(h.cleanup, func() {
        component.Unmount()
    })
    
    return test
}

func (h *TestHarness) installHooks(component Component) {
    // Hook into state changes
    // Hook into event emissions
    // Hook into lifecycle
}

func (h *TestHarness) extractState(component Component) {
    // Extract refs from component
    // Extract computed values
    // Store for inspection
}
```

---

## State Testing Architecture

### State Inspector

```go
type StateInspector struct {
    refs map[string]*Ref[interface{}]
}

func (si *StateInspector) GetRef(name string) *Ref[interface{}] {
    ref, ok := si.refs[name]
    if !ok {
        panic(fmt.Sprintf("ref %q not found", name))
    }
    return ref
}

func (si *StateInspector) GetRefValue(name string) interface{} {
    return si.GetRef(name).Get()
}

func (si *StateInspector) SetRefValue(name string, value interface{}) {
    si.GetRef(name).Set(value)
}

func (si *StateInspector) WaitForValue(name string, expected interface{}, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    ref := si.GetRef(name)
    
    for time.Now().Before(deadline) {
        if reflect.DeepEqual(ref.Get(), expected) {
            return nil
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    return fmt.Errorf("timeout waiting for %q to equal %v, got %v",
        name, expected, ref.Get())
}
```

### State Assertions

```go
// AssertRefEquals asserts ref value
func (ct *ComponentTest) AssertRefEquals(name string, expected interface{}) {
    actual := ct.state.GetRefValue(name)
    
    if !reflect.DeepEqual(actual, expected) {
        ct.harness.t.Errorf("ref %q: expected %v, got %v",
            name, expected, actual)
    }
}

// AssertRefChanged asserts ref changed from initial
func (ct *ComponentTest) AssertRefChanged(name string, initial interface{}) {
    actual := ct.state.GetRefValue(name)
    
    if reflect.DeepEqual(actual, initial) {
        ct.harness.t.Errorf("ref %q: expected change from %v",
            name, initial)
    }
}

// WaitForRef waits for ref to match expected
func (ct *ComponentTest) WaitForRef(name string, expected interface{}, timeout time.Duration) {
    err := ct.state.WaitForValue(name, expected, timeout)
    if err != nil {
        ct.harness.t.Fatal(err)
    }
}
```

---

## Event Testing Architecture

### Event Tracking

```go
type EventTracker struct {
    events []EmittedEvent
    mu     sync.RWMutex
}

func (et *EventTracker) Track(name string, payload interface{}, source string) {
    et.mu.Lock()
    defer et.mu.Unlock()
    
    et.events = append(et.events, EmittedEvent{
        Name:      name,
        Payload:   payload,
        Timestamp: time.Now(),
        Source:    source,
    })
}

func (et *EventTracker) GetEvents(name string) []EmittedEvent {
    et.mu.RLock()
    defer et.mu.RUnlock()
    
    events := []EmittedEvent{}
    for _, e := range et.events {
        if e.Name == name {
            events = append(events, e)
        }
    }
    
    return events
}

func (et *EventTracker) WasFired(name string) bool {
    return len(et.GetEvents(name)) > 0
}

func (et *EventTracker) FiredCount(name string) int {
    return len(et.GetEvents(name))
}
```

### Event Simulation

```go
func (ct *ComponentTest) Emit(name string, payload interface{}) {
    ct.component.Emit(name, payload)
    
    // Wait for event to process
    time.Sleep(1 * time.Millisecond)
}

func (ct *ComponentTest) EmitAndWait(name string, payload interface{}, timeout time.Duration) error {
    ct.Emit(name, payload)
    
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if ct.events.WasFired(name) {
            return nil
        }
        time.Sleep(10 * time.Millisecond)
    }
    
    return fmt.Errorf("timeout waiting for event %q", name)
}
```

### Event Assertions

```go
func (ct *ComponentTest) AssertEventFired(name string) {
    if !ct.events.WasFired(name) {
        ct.harness.t.Errorf("event %q was not fired", name)
    }
}

func (ct *ComponentTest) AssertEventNotFired(name string) {
    if ct.events.WasFired(name) {
        ct.harness.t.Errorf("event %q should not have fired", name)
    }
}

func (ct *ComponentTest) AssertEventPayload(name string, expected interface{}) {
    events := ct.events.GetEvents(name)
    if len(events) == 0 {
        ct.harness.t.Errorf("event %q was not fired", name)
        return
    }
    
    actual := events[len(events)-1].Payload
    if !reflect.DeepEqual(actual, expected) {
        ct.harness.t.Errorf("event %q payload: expected %v, got %v",
            name, expected, actual)
    }
}
```

---

## Mock System Architecture

### Mock Ref Implementation

```go
type MockRef[T any] struct {
    value    T
    getCalls int
    setCalls int
    watchers []func(T)
}

func NewMockRef[T any](initial T) *MockRef[T] {
    return &MockRef[T]{
        value:    initial,
        watchers: []func(T){},
    }
}

func (mr *MockRef[T]) Get() T {
    mr.getCalls++
    return mr.value
}

func (mr *MockRef[T]) Set(value T) {
    mr.setCalls++
    oldValue := mr.value
    mr.value = value
    
    // Notify watchers
    if !reflect.DeepEqual(oldValue, value) {
        for _, watcher := range mr.watchers {
            watcher(value)
        }
    }
}

func (mr *MockRef[T]) Watch(fn func(T)) {
    mr.watchers = append(mr.watchers, fn)
}

// Test assertions
func (mr *MockRef[T]) AssertGetCalled(t *testing.T, times int) {
    if mr.getCalls != times {
        t.Errorf("Get() called %d times, expected %d", mr.getCalls, times)
    }
}

func (mr *MockRef[T]) AssertSetCalled(t *testing.T, times int) {
    if mr.setCalls != times {
        t.Errorf("Set() called %d times, expected %d", mr.setCalls, times)
    }
}
```

### Mock Component

```go
type MockComponent struct {
    name          string
    initCalled    bool
    updateCalls   int
    viewCalls     int
    unmountCalled bool
    props         map[string]interface{}
    viewOutput    string
}

func NewMockComponent(name string) *MockComponent {
    return &MockComponent{
        name:       name,
        props:      make(map[string]interface{}),
        viewOutput: fmt.Sprintf("Mock<%s>", name),
    }
}

func (mc *MockComponent) Init() tea.Cmd {
    mc.initCalled = true
    return nil
}

func (mc *MockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    mc.updateCalls++
    return mc, nil
}

func (mc *MockComponent) View() string {
    mc.viewCalls++
    return mc.viewOutput
}

func (mc *MockComponent) Unmount() {
    mc.unmountCalled = true
}

func (mc *MockComponent) AssertInitCalled(t *testing.T) {
    if !mc.initCalled {
        t.Error("Init() was not called")
    }
}
```

---

## Snapshot Testing Architecture

### Snapshot Manager

```go
type SnapshotManager struct {
    dir    string
    update bool
    mu     sync.Mutex
}

func NewSnapshotManager(testDir string, update bool) *SnapshotManager {
    return &SnapshotManager{
        dir:    filepath.Join(testDir, "__snapshots__"),
        update: update,
    }
}

func (sm *SnapshotManager) Match(t *testing.T, name, actual string) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    filename := sm.getSnapshotFile(t, name)
    
    // Read existing snapshot
    expected, err := os.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            // No snapshot exists, create it
            sm.createSnapshot(t, filename, actual)
            return
        }
        t.Fatalf("failed to read snapshot: %v", err)
    }
    
    // Compare
    if string(expected) != actual {
        if sm.update {
            // Update mode: overwrite snapshot
            sm.updateSnapshot(t, filename, actual)
        } else {
            // Show diff
            diff := sm.generateDiff(string(expected), actual)
            t.Errorf("Snapshot mismatch for %q:\n%s\n\nRun with -update flag to update snapshots",
                name, diff)
        }
    }
}

func (sm *SnapshotManager) createSnapshot(t *testing.T, filename, content string) {
    err := os.MkdirAll(filepath.Dir(filename), 0755)
    if err != nil {
        t.Fatalf("failed to create snapshot dir: %v", err)
    }
    
    err = os.WriteFile(filename, []byte(content), 0644)
    if err != nil {
        t.Fatalf("failed to write snapshot: %v", err)
    }
    
    t.Logf("Created snapshot: %s", filename)
}

func (sm *SnapshotManager) updateSnapshot(t *testing.T, filename, content string) {
    err := os.WriteFile(filename, []byte(content), 0644)
    if err != nil {
        t.Fatalf("failed to update snapshot: %v", err)
    }
    
    t.Logf("Updated snapshot: %s", filename)
}

func (sm *SnapshotManager) generateDiff(expected, actual string) string {
    // Generate unified diff
    // Highlight differences
    // Format for readability
    return diff.Unified("expected", "actual", expected, actual)
}
```

### Snapshot Helpers

```go
func MatchSnapshot(t *testing.T, actual string) {
    sm := GetSnapshotManager(t)
    name := fmt.Sprintf("%s_%s", t.Name(), "default")
    sm.Match(t, name, actual)
}

func MatchNamedSnapshot(t *testing.T, name, actual string) {
    sm := GetSnapshotManager(t)
    fullName := fmt.Sprintf("%s_%s", t.Name(), name)
    sm.Match(t, fullName, actual)
}

func MatchComponentSnapshot(t *testing.T, component Component) {
    output := component.View()
    MatchSnapshot(t, output)
}
```

---

## Async Testing Architecture

### Wait Helpers

```go
type WaitOptions struct {
    Timeout  time.Duration
    Interval time.Duration
    Message  string
}

func WaitFor(t *testing.T, condition func() bool, opts WaitOptions) {
    if opts.Timeout == 0 {
        opts.Timeout = 5 * time.Second
    }
    if opts.Interval == 0 {
        opts.Interval = 10 * time.Millisecond
    }
    
    deadline := time.Now().Add(opts.Timeout)
    
    for time.Now().Before(deadline) {
        if condition() {
            return
        }
        time.Sleep(opts.Interval)
    }
    
    msg := "timeout waiting for condition"
    if opts.Message != "" {
        msg = opts.Message
    }
    t.Fatal(msg)
}

func WaitForState(t *testing.T, state *StateInspector, name string, expected interface{}, timeout time.Duration) {
    WaitFor(t, func() bool {
        actual := state.GetRefValue(name)
        return reflect.DeepEqual(actual, expected)
    }, WaitOptions{
        Timeout: timeout,
        Message: fmt.Sprintf("timeout waiting for %q to equal %v", name, expected),
    })
}
```

---

## Test Fixtures Architecture

### Fixture Builder

```go
type FixtureBuilder struct {
    props  map[string]interface{}
    state  map[string]interface{}
    events map[string]interface{}
}

func NewFixture() *FixtureBuilder {
    return &FixtureBuilder{
        props:  make(map[string]interface{}),
        state:  make(map[string]interface{}),
        events: make(map[string]interface{}),
    }
}

func (fb *FixtureBuilder) WithProp(key string, value interface{}) *FixtureBuilder {
    fb.props[key] = value
    return fb
}

func (fb *FixtureBuilder) WithState(key string, value interface{}) *FixtureBuilder {
    fb.state[key] = value
    return fb
}

func (fb *FixtureBuilder) WithEvent(name string, payload interface{}) *FixtureBuilder {
    fb.events[name] = payload
    return fb
}

func (fb *FixtureBuilder) Build(t *testing.T, createFn func() Component) *ComponentTest {
    harness := NewHarness(t)
    component := createFn()
    
    // Apply fixture
    for key, value := range fb.props {
        component.WithProp(key, value)
    }
    
    test := harness.Mount(component)
    
    // Set initial state
    for key, value := range fb.state {
        test.state.SetRefValue(key, value)
    }
    
    // Emit events
    for name, payload := range fb.events {
        test.Emit(name, payload)
    }
    
    return test
}
```

---

## Known Limitations & Solutions

### Limitation 1: Async Timing
**Problem**: Hard to test async operations deterministically  
**Current Design**: Polling with timeout  
**Solution**: Event-based waiting, promise-like patterns  
**Benefits**: More reliable async tests  
**Priority**: HIGH

### Limitation 2: Mock Complexity
**Problem**: Complex components hard to mock  
**Current Design**: Manual mock creation  
**Solution**: Mock generation tool, shallow rendering  
**Benefits**: Easier testing  
**Priority**: MEDIUM

### Limitation 3: Snapshot Instability
**Problem**: Dynamic content causes false failures  
**Current Design**: Ignore patterns  
**Solution**: Smart normalization, flexible matchers  
**Benefits**: Stable snapshots  
**Priority**: MEDIUM

### Limitation 4: Test Isolation
**Problem**: Global state can leak between tests  
**Current Design**: Manual cleanup  
**Solution**: Automatic isolation, per-test globals  
**Benefits**: Reliable tests  
**Priority**: HIGH

---

## Future Enhancements

### Phase 4+
1. **Visual Regression**: Terminal screenshot comparison
2. **Performance Testing**: Render time assertions
3. **Fuzzing**: Property-based testing
4. **Test Generation**: Auto-generate tests from components
5. **Coverage Reports**: Detailed coverage analysis
6. **Mutation Testing**: Test quality assessment

---

## Summary

The Testing Utilities framework provides comprehensive component testing capabilities through a test harness that mounts components in isolated environments, assertion helpers for type-safe verification, mock utilities for dependency isolation, and snapshot testing for render output validation. The system integrates with Go's built-in testing package and testify, supports table-driven tests, provides async testing helpers, and maintains < 1ms setup overhead per test with deterministic execution and automatic cleanup.
