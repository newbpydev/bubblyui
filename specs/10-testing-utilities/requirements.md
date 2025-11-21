# Feature Name: Testing Utilities

## Feature ID
10-testing-utilities

## Overview
Implement a comprehensive testing framework for BubblyUI components, providing utilities for unit testing, integration testing, snapshot testing, and end-to-end testing. The framework includes test harness creation, mock utilities, assertion helpers, state inspection, event simulation, and component mounting in test environments. It integrates with Go's built-in testing package and testify, enabling developers to write reliable tests with minimal boilerplate while maintaining type safety and following Go testing conventions.

## User Stories
- As a **developer**, I want to test components in isolation so that I can verify behavior without dependencies
- As a **developer**, I want to simulate events so that I can test event handlers
- As a **developer**, I want to assert on component state so that I can verify state changes
- As a **developer**, I want snapshot testing so that I can detect unintended UI changes
- As a **developer**, I want mock utilities so that I can isolate components from dependencies
- As a **developer**, I want test coverage reporting so that I can ensure code quality
- As a **TDD practitioner**, I want red-green-refactor workflow so that I can write tests first
- As a **team lead**, I want consistent test patterns so that tests are maintainable

## Functional Requirements

### 1. Component Test Harness
1.1. Mount components in test environment  
1.2. Initialize component with props  
1.3. Access component state  
1.4. Trigger lifecycle hooks  
1.5. Simulate messages  
1.6. Query component tree  
1.7. Cleanup after tests  
1.8. Isolation from other tests  

### 2. State Testing
2.1. Assert on ref values  
2.2. Assert on computed values  
2.3. Verify state changes  
2.4. Track state history  
2.5. Wait for state changes  
2.6. Async state testing  
2.7. State snapshots  

### 3. Event Testing
3.1. Simulate event emission  
3.2. Assert event handlers called  
3.3. Verify event payloads  
3.4. Track event bubbling  
3.5. Mock event handlers  
3.6. Test event order  
3.7. Async event handling  

### 4. Lifecycle Testing
4.1. Verify mount behavior  
4.2. Verify update behavior  
4.3. Verify unmount behavior  
4.4. Test hook execution order  
4.5. Test cleanup functions  
4.6. Test hook dependencies  
4.7. Infinite loop detection  

### 5. Rendering Testing
5.1. Assert on rendered output  
5.2. Snapshot testing  
5.3. Diff snapshots  
5.4. Update snapshots  
5.5. Partial matching  
5.6. Ignore dynamic content  
5.7. Visual regression detection  

### 6. Mock Utilities
6.1. Mock refs  
6.2. Mock computed values  
6.3. Mock watchers  
6.4. Mock components  
6.5. Mock props  
6.6. Mock event handlers  
6.7. Mock commands  
6.8. Mock router  

### 7. Assertion Helpers
7.1. Type-safe assertions  
7.2. Component assertions  
7.3. State assertions  
7.4. Event assertions  
7.5. Render assertions  
7.6. Custom matchers  
7.7. Async assertions  
7.8. Negation support  

### 8. Test Utilities
8.1. Test component builder  
8.2. Prop factories  
8.3. State factories  
8.4. Event factories  
8.5. Mock data generators  
8.6. Test fixtures  
8.7. Setup/teardown helpers  
8.8. Test isolation  

### 9. Integration Testing
9.1. Multi-component testing  
9.2. Parent-child interaction  
9.3. Event flow testing  
9.4. Router integration  
9.5. Full app testing  
9.6. E2E workflows  
9.7. Performance testing

### 10. Command System Testing
10.1. Command queue inspection  
10.2. Command batching verification  
10.3. Command deduplication testing  
10.4. Mock command generators  
10.5. Loop detection verification  
10.6. Auto-command testing  
10.7. Command strategy testing  
10.8. Command timeline inspection

### 11. Composables Testing
11.1. useAsync testing (promises, loading, errors)  
11.2. useDebounce testing (timing simulation)  
11.3. useThrottle testing (timing simulation)  
11.4. useForm testing (validation, errors)  
11.5. useLocalStorage testing (mock storage)  
11.6. useEffect testing (side effects)  
11.7. useEventListener testing (subscriptions)  
11.8. useState testing (simplified state)  
11.9. useTextInput testing (input handling)

### 12. Directives Testing
12.1. Bind directive testing (two-way binding)  
12.2. ForEach directive testing (list rendering)  
12.3. If directive testing (conditional render)  
12.4. On directive testing (event handlers)  
12.5. Show directive testing (visibility)  
12.6. Custom directive testing  
12.7. Directive error handling

### 13. Advanced Watch Testing
13.1. WatchEffect testing  
13.2. Flush mode testing (pre, post, sync)  
13.3. Deep watch testing  
13.4. Custom comparator testing  
13.5. Immediate option testing  
13.6. Watcher cleanup verification  
13.7. Watcher invocation counting  
13.8. Dependency tracking verification

### 14. Provide/Inject Testing
14.1. Provider registration testing  
14.2. Injection retrieval testing  
14.3. Injection tree traversal  
14.4. Default value testing  
14.5. Missing injection handling  
14.6. Nested injection testing

### 15. Router Advanced Testing
15.1. Route guard testing (beforeEnter, beforeLeave)  
15.2. Component guard testing  
15.3. Guard flow testing  
15.4. Navigation simulation  
15.5. History management testing  
15.6. Nested routes testing  
15.7. Query parameter testing  
15.8. Named routes testing  
15.9. Path matching testing  
15.10. Route pattern testing

### 16. Key Bindings Testing
16.1. Key binding registration verification  
16.2. Help text generation testing  
16.3. Key conflict detection  
16.4. Conditional binding testing  
16.5. Key press simulation  
16.6. Binding priority testing

### 17. Message Handler Testing
17.1. Message handler registration  
17.2. Custom message simulation  
17.3. Window resize message testing  
17.4. Mouse message testing  
17.5. Handler command return verification  
17.6. Handler execution order testing

### 18. Children Management Testing
18.1. Child initialization testing  
18.2. Child mounting/unmounting  
18.3. Child Update() cascade  
18.4. Child rendering order  
18.5. Child isolation testing  
18.6. Child state access  
18.7. Deeply nested tree testing

### 19. Template Safety Testing
19.1. Template context detection  
19.2. Set() in template prevention  
19.3. InTemplate() flag verification  
19.4. Mutation error messages

### 20. Computed Advanced Testing
20.1. Cache verification  
20.2. Recomputation testing  
20.3. Circular dependency detection  
20.4. Max depth error testing  
20.5. Dependency chain testing  
20.6. Mock computed values

### 21. Dependency Tracking Testing
21.1. Dependency collection verification  
21.2. Dependency invalidation testing  
21.3. Tracker state inspection  
21.4. Per-component tracking

### 22. Observability Testing
22.1. Mock error reporter  
22.2. Error context verification  
22.3. Breadcrumb tracking  
22.4. Panic reporting testing  
22.5. Sentry integration mocking  
22.6. Console reporter testing

### 23. Error Handling Testing
23.1. ErrNilCallback testing  
23.2. ErrNilComputeFn testing  
23.3. ErrCircularDependency testing  
23.4. ErrMaxDepthExceeded testing  
23.5. Component error testing  
23.6. Directive error testing  
23.7. Router error testing  
23.8. Panic recovery verification  

## Non-Functional Requirements

### Performance
- Test setup: < 1ms per test
- Test execution: Fast (no artificial delays)
- Snapshot comparison: < 10ms
- Mock overhead: Negligible
- Parallel test execution: Safe

### Usability
- Go testing conventions followed
- Testify integration seamless
- Minimal boilerplate
- Clear error messages
- Intuitive API

### Reliability
- Tests deterministic
- No flaky tests
- Proper cleanup
- Thread-safe
- Isolated test environment

### Maintainability
- DRY (Don't Repeat Yourself) helpers
- Reusable fixtures
- Clear test organization
- Easy to debug
- Good defaults

### Compatibility
- Works with Go testing
- Works with testify
- Works with gotestsum
- Works with coverage tools
- CI/CD friendly

## Acceptance Criteria

### Component Testing
- [ ] Components mount in tests
- [ ] State accessible in tests
- [ ] Events can be simulated
- [ ] Lifecycle hooks execute
- [ ] Cleanup works correctly
- [ ] Tests are isolated

### Assertions
- [ ] State assertions work
- [ ] Event assertions work
- [ ] Render assertions work
- [ ] Custom matchers work
- [ ] Error messages clear
- [ ] Type-safe assertions

### Mocking
- [ ] Refs can be mocked
- [ ] Components can be mocked
- [ ] Props can be mocked
- [ ] Commands can be mocked
- [ ] Router can be mocked
- [ ] Mocks easy to create

### Snapshots
- [ ] Snapshots generate correctly
- [ ] Diff shows changes
- [ ] Update workflow works
- [ ] Partial matching works
- [ ] Dynamic content ignored
- [ ] Git-friendly format

### Integration
- [ ] Multi-component tests work
- [ ] Event flow testable
- [ ] Router integration works
- [ ] Full app testing possible
- [ ] E2E workflows supported
- [ ] Coverage reporting works

### Developer Experience
- [ ] Minimal boilerplate
- [ ] Clear documentation
- [ ] Good error messages
- [ ] TDD workflow supported
- [ ] Fast test execution
- [ ] Easy debugging

### Commands
- [ ] Command queue inspectable
- [ ] Command batching verifiable
- [ ] Loop detection testable
- [ ] Auto-commands work correctly
- [ ] Mock generators easy to create
- [ ] Command strategies testable

### Composables
- [ ] All 9 composables testable
- [ ] Timing simulation works (debounce/throttle)
- [ ] Storage mocking works (useLocalStorage)
- [ ] Async testing reliable (useAsync)
- [ ] Form validation testable (useForm)
- [ ] Side effects trackable (useEffect)

### Directives
- [ ] All 5 directives testable
- [ ] Bind directive two-way binding works
- [ ] ForEach list rendering verifiable
- [ ] If conditional rendering testable
- [ ] Custom directives testable
- [ ] Directive errors catchable

### Advanced Watch
- [ ] WatchEffect testable
- [ ] Flush modes controllable
- [ ] Deep watching verifiable
- [ ] Custom comparators testable
- [ ] Watcher cleanup verifiable
- [ ] Invocation counting works

### Provide/Inject
- [ ] Injection testable
- [ ] Tree traversal verifiable
- [ ] Default values testable
- [ ] Missing injection handling works

### Router
- [ ] Guards testable (all types)
- [ ] Navigation simulation works
- [ ] History management testable
- [ ] Nested routes work
- [ ] Query params testable
- [ ] Named routes work

### System Features
- [ ] Key bindings fully testable
- [ ] Message handlers testable
- [ ] Children management verifiable
- [ ] Template safety enforceable
- [ ] Computed caching verifiable
- [ ] Dependency tracking inspectable
- [ ] Observability mockable
- [ ] Error handling comprehensive

## Dependencies

### Required Features
- **01-reactivity-system**: State testing (Ref, Computed, Watch, WatchEffect, Deep)
- **02-component-model**: Component testing (Setup, Template, Props, Children)
- **03-lifecycle-hooks**: Lifecycle testing (onMounted, onUpdated, onUnmounted)
- **04-composition-api**: Composable testing (useAsync, useDebounce, useForm, etc.) - CRITICAL
- **05-directives**: Directive testing (Bind, ForEach, If, On, Show) - CRITICAL
- **07-router**: Router testing (Guards, Navigation, History, Nested) - CRITICAL
- **08-automatic-reactive-bridge**: Command testing (Queue, Batcher, Auto-commands) - CRITICAL

### Additional Dependencies
- **Commands system**: Command queue, batching, loop detection
- **Observability system**: Error reporting, breadcrumbs
- **DevTools integration**: Component inspection, timeline

### External Dependencies
- **testify**: Assertion library
- **Go testing**: Built-in test framework
- **mockery** (optional): Mock generation

## Edge Cases

### 1. Async State Changes
**Challenge**: State changes asynchronously, test must wait  
**Handling**: `WaitFor()` helper with timeout, polling  

### 2. Event Handler Side Effects
**Challenge**: Handler modifies external state  
**Handling**: Mock external dependencies, assert on mocks  

### 3. Timing-Dependent Tests
**Challenge**: Tests depend on specific timing  
**Handling**: Avoid timing dependencies, use deterministic helpers  

### 4. Snapshot Instability
**Challenge**: Snapshots contain timestamps/random data  
**Handling**: Ignore patterns, normalize before comparison  

### 5. Test Interference
**Challenge**: Tests affect each other through global state  
**Handling**: Proper cleanup, isolated test environment  

### 6. Large Component Trees
**Challenge**: Testing complex nested components  
**Handling**: Shallow rendering, component mocking  

### 7. Memory Leaks in Tests
**Challenge**: Tests don't clean up properly  
**Handling**: Automatic cleanup, leak detection  

## Testing Requirements

### Unit Tests
- Test harness creation
- Assertion helpers
- Mock utilities
- Snapshot comparison
- State inspection

### Integration Tests
- Multi-component testing
- Event flow
- Lifecycle integration
- Router integration

### Meta Tests (Testing the Test Framework)
- Test framework reliability
- Mock behavior correctness
- Assertion accuracy
- Cleanup verification

## Atomic Design Level

**Tool/Utility** (Testing System)  
Not part of application code, but a separate testing framework that enables component verification.

## Related Components

### Tests
- Feature 01 (Reactivity): State testing
- Feature 02 (Components): Component testing
- Feature 03 (Lifecycle): Hook testing
- Feature 04 (Composition API): Composable testing
- Feature 07 (Router): Route testing
- Feature 08 (Bridge): Command testing

### Provides
- Test harness
- Assertion helpers
- Mock utilities
- Snapshot testing
- Test fixtures
- Coverage support

## Comparison with Vue Test Utils

### Similar Features
✅ Component mounting  
✅ State inspection  
✅ Event simulation  
✅ Snapshot testing  
✅ Mocking support  
✅ Assertion helpers  

### Go-Specific Differences
- **Table-Driven Tests**: Go idiom for parameterized tests
- **Testify Integration**: Popular Go assertion library
- **Go Testing Convention**: Use built-in testing package
- **Type Safety**: Generics for type-safe assertions
- **No DOM**: TUI rendering instead of DOM
- **Concurrent Tests**: Go's parallel test execution

### Additional Features for TUI
- Terminal output testing
- Lipgloss style testing
- Box drawing verification
- ANSI escape sequence handling
- Terminal size simulation

## Examples

### Basic Component Test
```go
func TestCounter(t *testing.T) {
    // Create test harness
    harness := testutil.NewHarness(t)
    
    // Mount component
    counter := harness.Mount(createCounter())
    
    // Assert initial state
    count := counter.GetRef("count")
    assert.Equal(t, 0, count.Get())
    
    // Simulate event
    counter.Emit("increment", nil)
    
    // Assert state changed
    assert.Equal(t, 1, count.Get())
}
```

### Snapshot Test
```go
func TestCounterRender(t *testing.T) {
    harness := testutil.NewHarness(t)
    counter := harness.Mount(createCounter())
    
    // Take snapshot
    output := counter.View()
    testutil.MatchSnapshot(t, output)
}
```

### Event Testing
```go
func TestEventFlow(t *testing.T) {
    harness := testutil.NewHarness(t)
    
    // Track events
    tracker := harness.TrackEvents()
    
    component := harness.Mount(createComponent())
    component.Emit("click", nil)
    
    // Assert event fired
    assert.True(t, tracker.WasFired("click"))
    assert.Equal(t, nil, tracker.GetPayload("click"))
}
```

## Future Considerations

### Post v1.0
- Visual regression testing (screenshot comparison)
- Performance benchmarking helpers
- Fuzzing utilities
- Property-based testing support
- Test generation from components
- Coverage visualization
- Mutation testing
- AI-powered test suggestions

### Out of Scope (v1.0)
- Browser-based testing (TUI-only)
- Screenshot testing (terminal screenshots)
- Network mocking (not applicable to TUI)
- Database mocking (app responsibility)

## Documentation Requirements

### API Documentation
- Test harness API
- Assertion helpers API
- Mock utilities API
- Snapshot testing API
- Configuration options

### Guides
- Getting started with testing
- Writing unit tests
- Writing integration tests
- Snapshot testing guide
- Mocking guide
- TDD workflow
- Best practices

### Examples
- Basic component tests
- Event testing examples
- State testing examples
- Lifecycle testing examples
- Integration test examples
- Full test suites

## Success Metrics

### Technical
- Test execution speed: < 1s per test
- Setup overhead: < 1ms
- Coverage: Possible to achieve 100%
- Flakiness: 0%
- Deterministic: 100%

### Developer Experience
- Time to first test: < 5 minutes
- Boilerplate reduction: 70%
- Test maintainability: High
- Debugging ease: High
- Satisfaction: > 90%

### Adoption
- 100% of BubblyUI tests use utilities
- 80%+ of user projects have tests
- Positive community feedback
- Featured in documentation
- Best practices established

## Integration Patterns

### Pattern 1: Table-Driven Tests
```go
func TestCounterIncrement(t *testing.T) {
    tests := []struct {
        name     string
        initial  int
        clicks   int
        expected int
    }{
        {"zero to one", 0, 1, 1},
        {"one to five", 1, 4, 5},
        {"large increment", 0, 100, 100},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            harness := testutil.NewHarness(t)
            counter := harness.Mount(createCounterWithValue(tt.initial))
            
            for i := 0; i < tt.clicks; i++ {
                counter.Emit("increment", nil)
            }
            
            count := counter.GetRef("count")
            assert.Equal(t, tt.expected, count.Get())
        })
    }
}
```

### Pattern 2: Fixture Setup
```go
func setupTestComponent(t *testing.T, props map[string]interface{}) *testutil.ComponentTest {
    harness := testutil.NewHarness(t)
    component := createComponent()
    
    for key, value := range props {
        component.WithProp(key, value)
    }
    
    return harness.Mount(component)
}

func TestWithFixture(t *testing.T) {
    comp := setupTestComponent(t, map[string]interface{}{
        "title": "Test",
        "count": 42,
    })
    
    // Test using fixture
    assert.Equal(t, "Test", comp.GetProp("title"))
}
```

### Pattern 3: Async Testing
```go
func TestAsyncUpdate(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAsyncComponent())
    
    component.Emit("fetch-data", nil)
    
    // Wait for state change
    testutil.WaitFor(t, func() bool {
        loading := component.GetRef("loading")
        return loading.Get().(bool) == false
    }, 5*time.Second, "data to load")
    
    data := component.GetRef("data")
    assert.NotNil(t, data.Get())
}
```

### Command Testing
```go
func TestCommandQueue(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createAutoCommandComponent())
    
    // Inspect command queue
    queue := harness.GetCommandQueue()
    assert.Equal(t, 0, queue.Len())
    
    // Trigger auto-command
    count := component.GetRef("count")
    count.Set(42)
    
    // Verify command enqueued
    assert.Equal(t, 1, queue.Len())
    cmd := queue.Peek()
    assert.NotNil(t, cmd)
}
```

### Composable Testing
```go
func TestUseDebounce(t *testing.T) {
    harness := testutil.NewHarness(t)
    timeSim := testutil.NewTimeSimulator()
    
    component := harness.MountWithTime(createDebounceComponent(), timeSim)
    
    input := component.GetRef("input")
    input.Set("test")
    
    // Advance time
    timeSim.Advance(100 * time.Millisecond)
    
    // Verify not debounced yet
    debounced := component.GetRef("debounced")
    assert.Equal(t, "", debounced.Get())
    
    // Advance past delay
    timeSim.Advance(300 * time.Millisecond)
    assert.Equal(t, "test", debounced.Get())
}
```

### Directive Testing
```go
func TestForEachDirective(t *testing.T) {
    harness := testutil.NewHarness(t)
    component := harness.Mount(createListComponent())
    
    items := component.GetRef("items")
    items.Set([]string{"a", "b", "c"})
    
    // Verify rendered output
    output := component.View()
    assert.Contains(t, output, "a")
    assert.Contains(t, output, "b")
    assert.Contains(t, output, "c")
    
    // Update list
    items.Set([]string{"x", "y"})
    output = component.View()
    assert.Contains(t, output, "x")
    assert.NotContains(t, output, "a")
}
```

### Router Testing
```go
func TestRouteGuard(t *testing.T) {
    harness := testutil.NewHarness(t)
    router := testutil.NewMockRouter()
    
    guardCalled := false
    route := router.AddRoute("/protected").WithGuard(func() bool {
        guardCalled = true
        return false  // Block navigation
    })
    
    // Attempt navigation
    router.Navigate("/protected")
    
    // Verify guard called and navigation blocked
    assert.True(t, guardCalled)
    assert.NotEqual(t, "/protected", router.CurrentPath())
}
```

## License
MIT License - consistent with project
