# User Workflow: Testing Utilities

## Developer Personas

### Persona 1: TDD Practitioner (David)
- **Background**: 5 years TDD, writes tests first
- **Goal**: Test-drive component development
- **Pain Point**: Too much test boilerplate
- **Expects**: Fast test setup, clear assertions
- **Success**: Writes failing test → implements → passes

### Persona 2: Quality Engineer (Priya)
- **Background**: 8 years testing, quality advocate
- **Goal**: Comprehensive test coverage
- **Pain Point**: Hard to test edge cases
- **Expects**: Easy mocking, good assertions
- **Success**: 100% coverage, all edge cases tested

### Persona 3: Junior Developer (Sam)
- **Background**: 6 months coding, learning testing
- **Goal**: Write first component test
- **Pain Point**: Testing seems complex
- **Expects**: Clear examples, helpful errors
- **Success**: Writes and runs first test successfully

---

## Primary User Journey: First Component Test

### Entry Point: Writing First Test

**Workflow: Creating and Running First Test**

#### Step 1: Create Test File
**User Action**: Create test file next to component

```go
// counter_test.go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)
```

**System Response**:
- Test file recognized by Go
- Imports resolve correctly
- Test package ready

#### Step 2: Write First Test
**User Action**: Write basic component test

```go
func TestCounterInitialState(t *testing.T) {
    // Create test harness
    harness := testutil.NewHarness(t)
    
    // Mount component
    counter := harness.Mount(createCounter())
    
    // Assert initial state
    count := counter.GetRef("count")
    assert.Equal(t, 0, count.Get())
}
```

**System Response**:
- Code compiles
- Autocomplete suggests methods
- Type checking passes

**UI Feedback**:
- IDE shows no errors
- Test runner recognizes test
- Ready to run

#### Step 3: Run Test
**User Action**: Run test with `go test`

```bash
$ go test -v
=== RUN   TestCounterInitialState
--- PASS: TestCounterInitialState (0.00s)
PASS
ok      counter 0.001s
```

**System Response**:
- Test executes successfully
- Component mounts correctly
- Assertion passes
- Cleanup automatic

**Journey Milestone**: ✅ First test passes!

---

### Feature Journey: Testing State Changes

#### Step 4: Test State Mutation
**User Action**: Test incrementing counter

```go
func TestCounterIncrement(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Get initial value
    count := counter.GetRef("count")
    assert.Equal(t, 0, count.Get())
    
    // Emit increment event
    counter.Emit("increment", nil)
    
    // Assert state changed
    assert.Equal(t, 1, count.Get())
}
```

**System Response**:
- Event emitted successfully
- State updates
- Assertion passes

**Test Output**:
```
=== RUN   TestCounterIncrement
--- PASS: TestCounterIncrement (0.00s)
```

**Journey Milestone**: ✅ State changes testable!

---

### Feature Journey: Table-Driven Tests

#### Step 5: Test Multiple Scenarios
**User Action**: Use table-driven test pattern

```go
func TestCounterMultipleIncrements(t *testing.T) {
    tests := []struct {
        name     string
        clicks   int
        expected int
    }{
        {"single click", 1, 1},
        {"five clicks", 5, 5},
        {"ten clicks", 10, 10},
        {"hundred clicks", 100, 100},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            counter := harness.Mount(createCounter())
            
            // Click multiple times
            for i := 0; i < tt.clicks; i++ {
                counter.Emit("increment", nil)
            }
            
            // Assert final count
            count := counter.GetRef("count")
            assert.Equal(t, tt.expected, count.Get())
        })
    }
}
```

**System Response**:
- All sub-tests execute
- Each test isolated
- Clear output per scenario

**Test Output**:
```
=== RUN   TestCounterMultipleIncrements
=== RUN   TestCounterMultipleIncrements/single_click
=== RUN   TestCounterMultipleIncrements/five_clicks
=== RUN   TestCounterMultipleIncrements/ten_clicks
=== RUN   TestCounterMultipleIncrements/hundred_clicks
--- PASS: TestCounterMultipleIncrements (0.00s)
    --- PASS: TestCounterMultipleIncrements/single_click (0.00s)
    --- PASS: TestCounterMultipleIncrements/five_clicks (0.00s)
    --- PASS: TestCounterMultipleIncrements/ten_clicks (0.00s)
    --- PASS: TestCounterMultipleIncrements/hundred_clicks (0.00s)
```

**Journey Milestone**: ✅ Multiple scenarios tested efficiently!

---

### Feature Journey: Snapshot Testing

#### Step 6: Test Render Output
**User Action**: Create snapshot test

```go
func TestCounterRender(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Take snapshot of initial render
    output := counter.View()
    testutil.MatchSnapshot(t, output)
}
```

**First Run**:
```
=== RUN   TestCounterRender
    testutil.go:123: Created snapshot: __snapshots__/TestCounterRender_default.snap
--- PASS: TestCounterRender (0.00s)
```

**Snapshot File Created**:
```
__snapshots__/TestCounterRender_default.snap
```

**Content**:
```
Counter: 0
[+] Increment
[-] Decrement
```

**Second Run** (no changes):
```
=== RUN   TestCounterRender
--- PASS: TestCounterRender (0.00s)
```

**After Changing Output**:
```
=== RUN   TestCounterRender
    testutil.go:145: Snapshot mismatch for "TestCounterRender_default":
    
    - Counter: 0
    + Count: 0
    
    Run with -update flag to update snapshots
--- FAIL: TestCounterRender (0.00s)
```

**Update Snapshot**:
```bash
$ go test -update
=== RUN   TestCounterRender
    testutil.go:158: Updated snapshot: __snapshots__/TestCounterRender_default.snap
--- PASS: TestCounterRender (0.00s)
```

**Journey Milestone**: ✅ Render output verified with snapshots!

---

### Feature Journey: Async Testing

#### Step 7: Test Async State Changes
**User Action**: Test component with async data loading

```go
func TestAsyncDataLoad(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createDataComponent())
    
    // Initially loading
    loading := component.GetRef("loading")
    assert.Equal(t, false, loading.Get())
    
    // Trigger data fetch
    component.Emit("fetch-data", nil)
    
    // Should be loading now
    assert.Equal(t, true, loading.Get())
    
    // Wait for data to load
    testutil.WaitFor(t, func() bool {
        return loading.Get().(bool) == false
    }, testutil.WaitOptions{
        Timeout: 5 * time.Second,
        Message: "data to finish loading",
    })
    
    // Assert data loaded
    data := component.GetRef("data")
    assert.NotNil(t, data.Get())
}
```

**System Response**:
- WaitFor polls condition
- Timeout prevents hanging
- Clear error if timeout

**Successful Run**:
```
=== RUN   TestAsyncDataLoad
--- PASS: TestAsyncDataLoad (0.15s)
```

**If Timeout**:
```
=== RUN   TestAsyncDataLoad
    testutil.go:201: data to finish loading
--- FAIL: TestAsyncDataLoad (5.01s)
```

**Journey Milestone**: ✅ Async operations testable!

---

## Alternative Workflows

### Workflow A: Mocking Dependencies

#### Entry: Component Has External Dependencies

**Scenario**: Testing component that uses external service

```go
func TestUserProfileWithMock(t *testing.T) {
    // Create mock service
    mockService := testutil.NewMockService()
    mockService.SetResponse("user", &User{
        Name: "Test User",
        Email: "test@example.com",
    })
    
    // Create component with mock
    harness := testutil.NewHarness(t)
    profile := harness.Mount(createProfileComponent(mockService))
    
    // Trigger fetch
    profile.Emit("load-user", "123")
    
    // Wait for data
    testutil.WaitForRef(t, profile, "user", &User{
        Name: "Test User",
        Email: "test@example.com",
    }, 1*time.Second)
    
    // Assert service was called correctly
    mockService.AssertCalled(t, "user", "123")
}
```

**Result**: Component tested without real service

---

### Workflow B: Testing Event Flow

#### Entry: Testing Parent-Child Event Bubbling

**Scenario**: Child emits event, parent handles

```go
func TestEventBubbling(t *testing.T) {
    harness := testutil.NewHarness(t)
    
    // Track events
    tracker := harness.TrackEvents()
    
    // Mount parent with child
    parent := harness.Mount(createParentWithChild())
    
    // Get child component
    child := parent.GetChild("child")
    
    // Emit event from child
    child.Emit("action", "test-data")
    
    // Assert parent received event
    assert.True(t, tracker.WasFired("action"))
    assert.Equal(t, "test-data", tracker.GetPayload("action"))
    
    // Assert parent handled event
    result := parent.GetRef("handledAction")
    assert.Equal(t, "test-data", result.Get())
}
```

**Result**: Event flow verified

---

## Error Recovery Workflows

### Error Flow 1: Test Timeout

**Trigger**: WaitFor times out

**User Sees**:
```
=== RUN   TestDataLoad
    testutil.go:201: timeout waiting for condition: data to finish loading
    Current state: loading=true, data=nil
--- FAIL: TestDataLoad (5.01s)
```

**Recovery**:
1. Check async operation is actually completing
2. Increase timeout if necessary
3. Verify condition function correct
4. Add debug logging

**Fixed Code**:
```go
// Add logging
component.Emit("fetch-data", nil)
t.Log("Fetch triggered")

testutil.WaitFor(t, func() bool {
    loading := loading.Get().(bool)
    t.Logf("Polling: loading=%v", loading)
    return !loading
}, ...)
```

---

### Error Flow 2: Snapshot Mismatch

**Trigger**: Render output changed

**User Sees**:
```
=== RUN   TestCounterRender
    testutil.go:145: Snapshot mismatch:
    
    Expected:
    Counter: 0
    [+] Increment
    
    Actual:
    Count: 0
    [+] Increment
    
    Run with -update to accept changes
--- FAIL: TestCounterRender (0.00s)
```

**Recovery Options**:

1. **Intentional Change** (update snapshot):
```bash
$ go test -update
```

2. **Unintentional Change** (fix code):
```go
// Revert change to component
```

3. **Review Diff**:
```bash
$ git diff __snapshots__/
```

---

### Error Flow 3: Ref Not Found

**Trigger**: Trying to access non-existent ref

**User Sees**:
```
=== RUN   TestCounter
panic: ref "cout" not found [recovered]
    panic: ref "cout" not found

Available refs: count, lastUpdate
--- FAIL: TestCounter (0.00s)
```

**Recovery**:
1. Check ref name spelling
2. Verify ref exposed in Setup
3. Check component initialization

**Fixed Code**:
```go
// Was: counter.GetRef("cout")  // Typo!
counter.GetRef("count")  // Correct
```

---

## State Transition Diagrams

### Test Execution Lifecycle
```
Test Function Starts
    ↓
Create Harness
    ├─ Initialize test environment
    ├─ Register cleanup handlers
    └─ Set up isolation
    ↓
Mount Component
    ├─ Initialize component
    ├─ Execute Init()
    ├─ Install test hooks
    └─ Extract state
    ↓
Execute Test Actions
    ├─ Emit events
    ├─ Update state
    ├─ Simulate messages
    └─ Wait for conditions
    ↓
Make Assertions
    ├─ State assertions
    ├─ Event assertions
    ├─ Render assertions
    └─ Custom assertions
    ↓
Test Completes
    ├─ Pass → Cleanup
    └─ Fail → Cleanup + Error
    ↓
Cleanup Executes
    ├─ Unmount component
    ├─ Remove hooks
    ├─ Clear state
    └─ Free resources
    ↓
Next Test or Exit
```

---

## Integration Points Map

### Feature Cross-Reference
```
10-testing-utilities
    ← Tests: 01-reactivity-system (state)
    ← Tests: 02-component-model (components)
    ← Tests: 03-lifecycle-hooks (lifecycle)
    ← Tests: 04-composition-api (composables)
    ← Tests: 05-directives (directives)
    ← Tests: 07-router (routes)
    ← Tests: 08-automatic-reactive-bridge (commands)
    → Used by: All BubblyUI projects
    → Enables: TDD workflow, quality assurance
```

---

## User Success Paths

### Path 1: Quick Test (< 5 minutes)
```
Create test file → Mount component → Assert state → Run → Pass → Success! 🎉
Tests: 1 | Passed: 1 | Failed: 0
```

### Path 2: TDD Workflow (< 30 minutes)
```
Write failing test → Implement feature → Test passes → Refactor → Tests still pass → Success! 🎉
Red → Green → Refactor cycle complete
```

### Path 3: Full Coverage (< 2 hours)
```
Test state → Test events → Test lifecycle → Test edge cases → 100% coverage → Success! 🎉
All code paths verified
```

---

## Common Patterns

### Pattern 1: Basic Component Test
```go
func TestComponent(t *testing.T) {
    harness := testutil.NewHarness(t)
    comp := harness.Mount(createComponent())
    
    // Test initial state
    state := comp.GetRef("value")
    assert.Equal(t, expectedValue, state.Get())
    
    // Test state change
    comp.Emit("change", newValue)
    assert.Equal(t, newValue, state.Get())
}
```

### Pattern 2: Table-Driven Test
```go
func TestScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected interface{}
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            comp := harness.Mount(createComponent())
            
            comp.Emit("input", tt.input)
            
            result := comp.GetRef("result")
            assert.Equal(t, tt.expected, result.Get())
        })
    }
}
```

### Pattern 3: Fixture-Based Test
```go
func setupComponent(t *testing.T) *testutil.ComponentTest {
    fixture := testutil.NewFixture().
        WithProp("title", "Test").
        WithState("count", 10).
        Build(t, createComponent)
    
    return fixture
}

func TestWithFixture(t *testing.T) {
    comp := setupComponent(t)
    // Test using pre-configured component
}
```

### Pattern 4: Snapshot Test
```go
func TestRender(t *testing.T) {
    harness := testutil.NewHarness(t)
    comp := harness.Mount(createComponent())
    
    testutil.MatchSnapshot(t, comp.View())
}
```

---

## Workflow: Testing Commands & Auto-Bridge

### Entry Point: Testing Command Queue

**User Goal**: Verify commands are enqueued correctly

#### Step 1: Mount Component with Auto-Commands
```go
func TestCommandQueue(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAutoCommandComponent())
    
    // Get command queue inspector
    queue := harness.GetCommandQueue()
    assert.Equal(t, 0, queue.Len())
}
```

#### Step 2: Trigger Command Generation
```go
    // Change state (triggers auto-command)
    count := component.GetRef("count")
    count.Set(42)
    
    // Verify command enqueued
    assert.Equal(t, 1, queue.Len())
    cmd := queue.Peek()
    assert.NotNil(t, cmd)
```

#### Step 3: Verify Loop Detection
```go
    // Simulate rapid updates
    for i := 0; i < 150; i++ {
        count.Set(i)
    }
    
    // Verify loop detected and blocked
    loopDetector := harness.GetLoopDetector()
    assert.True(t, loopDetector.WasDetected())
```

**Journey Milestone**: ✅ Command system tested!

---

## Workflow: Testing Composables

### Entry Point: Testing useDebounce with Time Simulation

**User Goal**: Test debounced values without real delays

#### Step 1: Setup Time Simulator
```go
func TestUseDebounce(t *testing.T) {
    harness := testutil.NewHarness(t)
    timeSim := testutil.NewTimeSimulator()
    
    component := harness.MountWithTime(createDebounceComponent(), timeSim)
}
```

#### Step 2: Trigger Debounced Change
```go
    input := component.GetRef("input")
    input.Set("test value")
    
    // Verify not debounced yet
    debounced := component.GetRef("debounced")
    assert.Equal(t, "", debounced.Get())
```

#### Step 3: Fast-Forward Time
```go
    // Advance time past debounce delay
    timeSim.Advance(300 * time.Millisecond)
    
    // Now it should be debounced
    assert.Equal(t, "test value", debounced.Get())
```

**Journey Milestone**: ✅ Composable tested without delays!

### Alternative: Testing useLocalStorage

#### Step 1: Setup Mock Storage
```go
func TestUseLocalStorage(t *testing.T) {
    harness := testutil.NewHarness(t)
    storage := testutil.NewMockStorage()
    
    component := harness.MountWithStorage(createStorageComponent(), storage)
}
```

#### Step 2: Verify Storage Operations
```go
    // Trigger save
    data := component.GetRef("data")
    data.Set("saved data")
    
    // Verify stored
    storage.AssertSetCalled(t, "data", 1)
    assert.Equal(t, "saved data", storage.Get("data"))
```

---

## Workflow: Testing Directives

### Entry Point: Testing ForEach Directive

**User Goal**: Verify list rendering with ForEach

#### Step 1: Mount Component with List
```go
func TestForEachDirective(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createListComponent())
    
    items := component.GetRef("items")
    items.Set([]string{"apple", "banana", "cherry"})
}
```

#### Step 2: Verify Rendering
```go
    output := component.View()
    
    // All items rendered
    assert.Contains(t, output, "apple")
    assert.Contains(t, output, "banana")
    assert.Contains(t, output, "cherry")
```

#### Step 3: Update List and Verify
```go
    // Remove item
    items.Set([]string{"apple", "cherry"})
    output = component.View()
    
    assert.Contains(t, output, "apple")
    assert.NotContains(t, output, "banana")
    assert.Contains(t, output, "cherry")
```

**Journey Milestone**: ✅ Directive rendering verified!

### Alternative: Testing Bind Directive

#### Step 1: Setup Two-Way Binding Test
```go
func TestBindDirective(t *testing.T) {
    harness := testutil.NewHarness(t)
    bindTester := testutil.NewBindTester()
    
    component := harness.MountWithBindings(createFormComponent(), bindTester)
}
```

#### Step 2: Test Ref → Element
```go
    // Change ref value
    name := component.GetRef("name")
    name.Set("John Doe")
    
    // Verify element updated
    bindTester.AssertElementValue(t, "nameInput", "John Doe")
```

#### Step 3: Test Element → Ref
```go
    // Simulate element change
    bindTester.TriggerElementChange("nameInput", "Jane Smith")
    
    // Verify ref updated
    assert.Equal(t, "Jane Smith", name.Get())
```

---

## Workflow: Testing Router with Guards

### Entry Point: Testing Route Guards

**User Goal**: Verify navigation guards work correctly

#### Step 1: Setup Router with Guards
```go
func TestRouteGuard(t *testing.T) {
    harness := testutil.NewHarness(t)
    router := testutil.NewMockRouter()
    
    guardCalled := false
    route := router.AddRoute("/protected").WithGuard(func() bool {
        guardCalled = true
        return false  // Block navigation
    })
}
```

#### Step 2: Attempt Navigation
```go
    // Try to navigate
    router.Navigate("/protected")
    
    // Verify guard called
    assert.True(t, guardCalled)
```

#### Step 3: Verify Navigation Blocked
```go
    // Should still be at previous route
    assert.NotEqual(t, "/protected", router.CurrentPath())
    assert.Equal(t, "/", router.CurrentPath())
```

**Journey Milestone**: ✅ Router guards tested!

### Alternative: Testing Navigation History

#### Step 1: Setup Navigation Simulator
```go
func TestNavigationHistory(t *testing.T) {
    harness := testutil.NewHarness(t)
    navSim := testutil.NewNavigationSimulator()
}
```

#### Step 2: Navigate Multiple Times
```go
    navSim.Navigate("/home")
    navSim.Navigate("/profile")
    navSim.Navigate("/settings")
    
    navSim.AssertCurrentPath(t, "/settings")
    navSim.AssertHistoryLength(t, 3)
```

#### Step 3: Test Back/Forward
```go
    navSim.Back()
    navSim.AssertCurrentPath(t, "/profile")
    
    navSim.Forward()
    navSim.AssertCurrentPath(t, "/settings")
```

---

## Workflow: Testing Advanced Watch Features

### Entry Point: Testing WatchEffect

**User Goal**: Test automatic dependency tracking

#### Step 1: Setup WatchEffect
```go
func TestWatchEffect(t *testing.T) {
    harness := testutil.NewHarness(t)
    effectTester := testutil.NewWatchEffectTester()
    
    count := harness.Ref(0)
    doubled := harness.Ref(0)
    
    effect := effectTester.TrackEffect(func() {
        doubled.Set(count.Get().(int) * 2)
    })
}
```

#### Step 2: Trigger Dependency
```go
    count.Set(5)
    
    // WatchEffect should auto-execute
    effectTester.AssertExecuted(t, 1)
    assert.Equal(t, 10, doubled.Get())
```

#### Step 3: Verify Multiple Triggers
```go
    count.Set(10)
    effectTester.AssertExecuted(t, 2)
    assert.Equal(t, 20, doubled.Get())
```

**Journey Milestone**: ✅ WatchEffect tested!

### Alternative: Testing Flush Modes

#### Step 1: Setup Flush Mode Controller
```go
func TestFlushMode(t *testing.T) {
    harness := testutil.NewHarness(t)
    flushCtrl := testutil.NewFlushModeController("post")
    
    count := harness.Ref(0)
    execOrder := []int{}
}
```

#### Step 2: Register Multiple Watchers
```go
    flushCtrl.AddWatcher(count, func() {
        execOrder = append(execOrder, 1)
    })
    flushCtrl.AddWatcher(count, func() {
        execOrder = append(execOrder, 2)
    })
```

#### Step 3: Trigger and Verify Order
```go
    count.Set(42)
    flushCtrl.Flush()
    
    // Verify execution order
    flushCtrl.AssertExecutionOrder(t, []int{1, 2})
```

---

## Workflow: Testing Provide/Inject

### Entry Point: Testing Injection Tree

**User Goal**: Verify dependency injection works across component tree

#### Step 1: Setup Provider
```go
func TestProvideInject(t *testing.T) {
    harness := testutil.NewHarness(t)
    injector := testutil.NewProvideInjectTester()
    
    // Root provides theme
    root := harness.Mount(createRootComponent())
    injector.Provide("theme", darkTheme)
}
```

#### Step 2: Create Child Components
```go
    // Child injects theme
    child := harness.Mount(createChildComponent())
    root.AddChild(child)
    
    // Verify injection
    theme := injector.Inject(child, "theme")
    assert.Equal(t, darkTheme, theme)
```

#### Step 3: Test Tree Traversal
```go
    // Deeply nested child
    grandchild := harness.Mount(createGrandchildComponent())
    child.AddChild(grandchild)
    
    // Should still inject from root
    theme = injector.Inject(grandchild, "theme")
    assert.Equal(t, darkTheme, theme)
    injector.AssertTreeTraversal(t, 2)  // Traversed 2 levels
```

**Journey Milestone**: ✅ Provide/Inject tested!

---

## Workflow: Testing Key Bindings

### Entry Point: Testing Key Binding Registration

**User Goal**: Verify key bindings work and help text generates

#### Step 1: Setup Component with Bindings
```go
func TestKeyBindings(t *testing.T) {
    harness := testutil.NewHarness(t)
    kbTester := testutil.NewKeyBindingsTester()
    
    component := harness.Mount(createInteractiveComponent())
    kbTester.RegisterBinding(" ", "increment", "Increment counter")
    kbTester.RegisterBinding("r", "reset", "Reset counter")
}
```

#### Step 2: Simulate Key Press
```go
    // Simulate space key
    cmd := kbTester.SimulateKeyPress(" ")
    
    // Verify event emitted
    assert.NotNil(t, cmd)
    component.AssertEventEmitted(t, "increment")
```

#### Step 3: Verify Help Text
```go
    helpText := component.HelpText()
    assert.Contains(t, helpText, "space: Increment counter")
    assert.Contains(t, helpText, "r: Reset counter")
```

**Journey Milestone**: ✅ Key bindings tested!

---

## Workflow: Testing Children Management

### Entry Point: Testing Child Cascade

**User Goal**: Verify children receive Updates correctly

#### Step 1: Setup Parent with Children
```go
func TestChildrenCascade(t *testing.T) {
    harness := testutil.NewHarness(t)
    childTester := testutil.NewChildrenTester()
    
    parent := harness.Mount(createParentComponent())
    child1 := harness.Mount(createChildComponent("child1"))
    child2 := harness.Mount(createChildComponent("child2"))
    
    childTester.AddChild(parent, child1)
    childTester.AddChild(parent, child2)
}
```

#### Step 2: Trigger Parent Update
```go
    // Send message to parent
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
    childTester.TriggerUpdate(msg)
```

#### Step 3: Verify Cascade
```go
    // Verify children received update
    childTester.AssertCascade(t)
    childTester.AssertUpdateOrder(t, []string{"child1", "child2"})
```

**Journey Milestone**: ✅ Children cascade tested!

---

## Workflow: Testing Template Safety

### Entry Point: Testing Mutation Prevention

**User Goal**: Verify templates can't mutate state

#### Step 1: Setup Safety Tester
```go
func TestTemplateSafety(t *testing.T) {
    harness := testutil.NewHarness(t)
    safetyTester := testutil.NewTemplateSafetyTester()
    
    component := harness.Mount(createComponent())
    count := component.GetRef("count")
}
```

#### Step 2: Enter Template Context
```go
    safetyTester.EnterTemplate()
    
    // Attempt mutation in template
    safetyTester.AttemptMutation(count, 42)
```

#### Step 3: Verify Panic
```go
    // Should panic with clear error
    safetyTester.AssertPanics(t)
    safetyTester.AssertErrorMessage(t, "Cannot Set() during template rendering")
```

**Journey Milestone**: ✅ Template safety verified!

---

## Workflow: Testing Observability Integration

### Entry Point: Testing Error Reporting

**User Goal**: Verify errors are reported to observability system

#### Step 1: Setup Mock Reporter
```go
func TestErrorReporting(t *testing.T) {
    harness := testutil.NewHarness(t)
    mockReporter := testutil.NewMockErrorReporter()
    observability.SetErrorReporter(mockReporter)
    defer observability.SetErrorReporter(nil)
}
```

#### Step 2: Trigger Error
```go
    component := harness.Mount(createComponent())
    
    // Trigger panic in event handler
    component.Emit("triggerPanic", nil)
```

#### Step 3: Verify Error Reported
```go
    // Verify panic reported
    mockReporter.AssertPanicReported(t)
    
    // Verify context captured
    contexts := mockReporter.GetContexts()
    assert.Len(t, contexts, 1)
    assert.Equal(t, "triggerPanic", contexts[0].EventName)
    assert.NotEmpty(t, contexts[0].StackTrace)
```

**Journey Milestone**: ✅ Observability tested!

---

## Tips & Tricks

### Tip 1: Use Subtests for Organization
```go
func TestCounter(t *testing.T) {
    t.Run("initial state", func(t *testing.T) { ... })
    t.Run("increment", func(t *testing.T) { ... })
    t.Run("decrement", func(t *testing.T) { ... })
    t.Run("reset", func(t *testing.T) { ... })
}
```

### Tip 2: Helper Functions Reduce Boilerplate
```go
func mountCounter(t *testing.T, initial int) *testutil.ComponentTest {
    harness := testutil.NewHarness(t)
    return harness.Mount(createCounterWithValue(initial))
}

func TestIncrement(t *testing.T) {
    counter := mountCounter(t, 0)  // Cleaner!
    // ...
}
```

### Tip 3: Use testify's require for Fatal Assertions
```go
func TestCritical(t *testing.T) {
    comp := mountComponent(t)
    
    // If this fails, no point continuing
    require.NotNil(t, comp)
    
    // These only run if above passed
    assert.Equal(t, expected, comp.GetRef("value").Get())
}
```

### Tip 4: Parallel Tests for Speed
```go
func TestIndependent(t *testing.T) {
    t.Parallel()  // Safe to run in parallel
    
    // Test implementation
}
```

### Tip 5: Use Build Tags for Integration Tests
```go
//go:build integration

func TestIntegration(t *testing.T) {
    // Only runs with: go test -tags=integration
}
```

---

## Summary

The Testing Utilities framework provides a comprehensive testing solution for BubblyUI components through a test harness that mounts components in isolated environments (< 1ms setup), state and event inspection with type-safe assertions, snapshot testing with diff visualization, async testing helpers with timeout protection, and mock utilities for dependency isolation. Developers write tests using familiar Go testing conventions, table-driven test patterns, and testify assertions, achieving fast test execution, clear error messages, automatic cleanup, and reliable test isolation.

**Core Testing Capabilities**:
- ✅ Zero-boilerplate component mounting
- ✅ Type-safe assertions with clear errors
- ✅ Table-driven test support (Go idiom)
- ✅ Snapshot testing with easy updates
- ✅ Async testing without flakiness
- ✅ Mock utilities for isolation
- ✅ TDD workflow supported
- ✅ Fast execution (< 1s per test)

**Advanced Testing Capabilities (New)**:
- ✅ **Commands**: Queue inspection, batching verification, loop detection, auto-command testing
- ✅ **Composables**: Time simulation (debounce/throttle), storage mocking (useLocalStorage), async testing (useAsync), form validation (useForm), all 9 composables covered
- ✅ **Directives**: ForEach list rendering, Bind two-way binding, If conditionals, On event handlers, Show visibility, custom directives
- ✅ **Router**: Guard testing (beforeEnter, beforeLeave), navigation simulation, history management, nested routes, query params, named routes
- ✅ **Advanced Watch**: WatchEffect with auto-tracking, flush mode control (pre/post/sync), deep watching, custom comparators, invocation counting
- ✅ **Provide/Inject**: Injection tree traversal, provider registration, default values, missing injection handling
- ✅ **Key Bindings**: Binding registration, help text generation, conflict detection, key press simulation
- ✅ **Message Handlers**: Custom message simulation, window resize/mouse events, command verification
- ✅ **Children Management**: Cascade verification, init order testing, tree inspection, isolation testing
- ✅ **Template Safety**: Mutation prevention, context detection, panic verification
- ✅ **Computed Advanced**: Cache verification, recomputation testing, circular dependency detection, dependency chains
- ✅ **Dependency Tracking**: Collection verification, invalidation testing, tracker inspection
- ✅ **Observability**: Mock error reporters, panic reporting, breadcrumbs, context capture, Sentry/Console integration
- ✅ **Error Handling**: All error types testable, panic recovery verification, comprehensive coverage

**Coverage**: 95%+ of bubbly package features testable with professional-grade utilities
