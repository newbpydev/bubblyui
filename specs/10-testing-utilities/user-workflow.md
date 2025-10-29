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

**Key Success Factors**:
- ✅ Zero-boilerplate component mounting
- ✅ Type-safe assertions with clear errors
- ✅ Table-driven test support (Go idiom)
- ✅ Snapshot testing with easy updates
- ✅ Async testing without flakiness
- ✅ Mock utilities for isolation
- ✅ TDD workflow supported
- ✅ Fast execution (< 1s per test)
