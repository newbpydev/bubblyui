# Mocking Guide

## Overview

Learn how to use mocks to isolate components from dependencies and write focused, reliable tests. This guide covers mocking refs, components, external services, and more.

## Table of Contents

- [Why Mock?](#why-mock)
- [Mock Refs](#mock-refs)
- [Mock Components](#mock-components)
- [Mock Services](#mock-services)
- [Mock Storage](#mock-storage)
- [Mock Time](#mock-time)
- [Best Practices](#best-practices)

## Why Mock?

Mocking helps you:

- **Isolate components** - Test in isolation from dependencies
- **Control behavior** - Simulate specific scenarios
- **Avoid side effects** - No real network calls or file I/O
- **Speed up tests** - Fast, deterministic execution
- **Test edge cases** - Simulate errors and unusual conditions

## Mock Refs

### MockRef

Mock implementation of `Ref[T]` that tracks usage.

#### Creating Mock Refs

```go
func TestWithMockRef(t *testing.T) {
    mockRef := testutil.NewMockRef(42)
    
    // Use like a normal ref
    value := mockRef.Get()
    assert.Equal(t, 42, value)
    
    mockRef.Set(100)
    assert.Equal(t, 100, mockRef.Get())
}
```

#### Tracking Usage

```go
func TestMockRefTracking(t *testing.T) {
    mockRef := testutil.NewMockRef("initial")
    
    // Perform operations
    mockRef.Get()
    mockRef.Get()
    mockRef.Set("updated")
    
    // Assert usage
    mockRef.AssertGetCalls(t, 2)
    mockRef.AssertSetCalls(t, 1)
    mockRef.AssertValue(t, "updated")
}
```

#### Use Cases

**Testing computed values:**

```go
func TestComputedWithMockRef(t *testing.T) {
    mockCount := testutil.NewMockRef(5)
    
    doubled := bubbly.NewComputed(func() interface{} {
        return mockCount.Get().(int) * 2
    })
    
    assert.Equal(t, 10, doubled.Get())
    
    mockCount.Set(10)
    assert.Equal(t, 20, doubled.Get())
    
    mockCount.AssertGetCalls(t, 2)
}
```

**Testing watchers:**

```go
func TestWatcherWithMockRef(t *testing.T) {
    mockRef := testutil.NewMockRef(0)
    
    callCount := 0
    bubbly.Watch(mockRef, func() {
        callCount++
    })
    
    mockRef.Set(1)
    mockRef.Set(2)
    
    assert.Equal(t, 2, callCount)
    mockRef.AssertSetCalls(t, 2)
}
```

## Mock Components

### MockComponent

Mock implementation for testing parent-child interactions.

#### Creating Mock Components

```go
func TestWithMockChild(t *testing.T) {
    mockChild := testutil.NewMockComponent()
    mockChild.SetView("Mock Child View")
    
    parent := createParentWithChild(mockChild)
    
    harness := testutil.NewHarness(t)
    mounted := harness.Mount(parent)
    
    output := mounted.Component().View()
    assert.Contains(t, output, "Mock Child View")
}
```

#### Tracking Child Interactions

```go
func TestChildInteractions(t *testing.T) {
    mockChild := testutil.NewMockComponent()
    
    parent := createParentWithChild(mockChild)
    
    harness := testutil.NewHarness(t)
    mounted := harness.Mount(parent)
    
    // Trigger parent action that affects child
    mounted.Component().Emit("updateChild", "new data")
    
    // Assert child was updated
    mockChild.AssertUpdateCalled(t, 1)
    mockChild.AssertLastMessage(t, "new data")
}
```

#### Simulating Child Events

```go
func TestChildEvents(t *testing.T) {
    mockChild := testutil.NewMockComponent()
    
    parent := createParentWithChild(mockChild)
    
    harness := testutil.NewHarness(t)
    tracker := harness.TrackEvents()
    mounted := harness.Mount(parent)
    
    // Simulate child emitting event
    mockChild.SimulateEvent("childAction", "data")
    
    // Assert parent handled event
    tracker.AssertFired(t, "childAction")
}
```

## Mock Services

### Creating Service Mocks

Mock external services to avoid real network calls:

```go
type MockUserService struct {
    users map[string]*User
    calls int
}

func NewMockUserService() *MockUserService {
    return &MockUserService{
        users: make(map[string]*User),
    }
}

func (m *MockUserService) GetUser(id string) (*User, error) {
    m.calls++
    
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    
    return nil, errors.New("user not found")
}

func (m *MockUserService) SetUser(id string, user *User) {
    m.users[id] = user
}

func (m *MockUserService) AssertCalls(t *testing.T, expected int) {
    t.Helper()
    assert.Equal(t, expected, m.calls)
}
```

### Using Service Mocks

```go
func TestUserProfile(t *testing.T) {
    // Create mock service
    mockService := NewMockUserService()
    mockService.SetUser("123", &User{
        Name:  "John Doe",
        Email: "john@example.com",
    })
    
    // Create component with mock
    harness := testutil.NewHarness(t)
    profile := harness.Mount(createProfileComponent(mockService))
    
    // Trigger fetch
    profile.Component().Emit("load-user", "123")
    
    // Wait for data
    testutil.WaitFor(t, func() bool {
        user := profile.State().GetRef("user")
        return user.Get() != nil
    }, testutil.WaitOptions{
        Timeout: 1 * time.Second,
    })
    
    // Assert service was called
    mockService.AssertCalls(t, 1)
    
    // Assert data loaded
    user := profile.State().GetRef("user")
    loadedUser := user.Get().(*User)
    assert.Equal(t, "John Doe", loadedUser.Name)
}
```

### Simulating Errors

```go
func TestUserProfileError(t *testing.T) {
    mockService := NewMockUserService()
    // Don't set user - will return error
    
    harness := testutil.NewHarness(t)
    profile := harness.Mount(createProfileComponent(mockService))
    
    profile.Component().Emit("load-user", "999")
    
    testutil.WaitFor(t, func() bool {
        err := profile.State().GetRef("error")
        return err.Get() != nil
    }, testutil.WaitOptions{
        Timeout: 1 * time.Second,
    })
    
    // Assert error handled
    err := profile.State().GetRef("error")
    assert.NotNil(t, err.Get())
}
```

## Mock Storage

### MockStorage

Mock local storage for testing persistence:

```go
func TestUseLocalStorage(t *testing.T) {
    harness := testutil.NewHarness(t)
    storage := testutil.NewMockStorage()
    
    component := harness.MountWithStorage(createStorageComponent(), storage)
    
    // Set value
    data := component.State().GetRef("data")
    data.Set("saved data")
    
    // Verify stored
    storage.AssertSetCalled(t, "data", 1)
    assert.Equal(t, "saved data", storage.Get("data"))
    
    // Clear storage
    storage.Clear()
    
    // Verify cleared
    assert.Nil(t, storage.Get("data"))
}
```

### Storage Operations

```go
type MockStorage struct {
    data     map[string]interface{}
    setCalls map[string]int
    mu       sync.RWMutex
}

func NewMockStorage() *MockStorage {
    return &MockStorage{
        data:     make(map[string]interface{}),
        setCalls: make(map[string]int),
    }
}

func (ms *MockStorage) Get(key string) interface{} {
    ms.mu.RLock()
    defer ms.mu.RUnlock()
    return ms.data[key]
}

func (ms *MockStorage) Set(key string, value interface{}) {
    ms.mu.Lock()
    defer ms.mu.Unlock()
    
    ms.data[key] = value
    ms.setCalls[key]++
}

func (ms *MockStorage) Clear() {
    ms.mu.Lock()
    defer ms.mu.Unlock()
    
    ms.data = make(map[string]interface{})
}

func (ms *MockStorage) AssertSetCalled(t *testing.T, key string, times int) {
    t.Helper()
    
    ms.mu.RLock()
    defer ms.mu.RUnlock()
    
    actual := ms.setCalls[key]
    assert.Equal(t, times, actual, "Set(%s) called wrong number of times", key)
}
```

## Mock Time

### TimeSimulator

Simulate time passage for testing time-dependent code:

```go
func TestDebounce(t *testing.T) {
    harness := testutil.NewHarness(t)
    timeSim := testutil.NewTimeSimulator()
    
    component := harness.MountWithTime(createDebounceComponent(), timeSim)
    
    input := component.State().GetRef("input")
    debounced := component.State().GetRef("debounced")
    
    // Set value
    input.Set("test")
    
    // Not debounced yet
    assert.Equal(t, "", debounced.Get())
    
    // Advance time (not enough)
    timeSim.Advance(100 * time.Millisecond)
    assert.Equal(t, "", debounced.Get())
    
    // Advance past delay
    timeSim.Advance(300 * time.Millisecond)
    assert.Equal(t, "test", debounced.Get())
}
```

### Testing Throttle

```go
func TestThrottle(t *testing.T) {
    harness := testutil.NewHarness(t)
    timeSim := testutil.NewTimeSimulator()
    
    component := harness.MountWithTime(createThrottleComponent(), timeSim)
    
    input := component.State().GetRef("input")
    throttled := component.State().GetRef("throttled")
    
    // First value passes immediately
    input.Set("value1")
    assert.Equal(t, "value1", throttled.Get())
    
    // Second value throttled
    input.Set("value2")
    assert.Equal(t, "value1", throttled.Get())
    
    // Advance time
    timeSim.Advance(200 * time.Millisecond)
    input.Set("value3")
    assert.Equal(t, "value3", throttled.Get())
}
```

## Best Practices

### 1. Mock at the Right Level

```go
// ✅ Good: Mock external dependencies
mockService := NewMockUserService()
component := createComponent(mockService)

// ❌ Bad: Mock internal implementation
mockInternalFunction := NewMockInternal()
```

### 2. Use Dependency Injection

```go
// ✅ Good: Accept service as parameter
func createComponent(service UserService) bubbly.Component {
    return bubbly.NewComponent().Setup(func(ctx *bubbly.Context) {
        // Use service
    })
}

// ❌ Bad: Hard-coded dependency
func createComponent() bubbly.Component {
    service := NewRealUserService() // Can't mock
    // ...
}
```

### 3. Keep Mocks Simple

```go
// ✅ Good: Simple, focused mock
type MockService struct {
    response *User
    err      error
}

func (m *MockService) GetUser(id string) (*User, error) {
    return m.response, m.err
}

// ❌ Bad: Overly complex mock
type MockService struct {
    responses map[string]*User
    delays    map[string]time.Duration
    errors    map[string]error
    // Too much complexity
}
```

### 4. Verify Mock Interactions

```go
// ✅ Good: Verify mock was called correctly
mockService.AssertCalls(t, 1)
mockService.AssertCalledWith(t, "123")

// ❌ Bad: Don't verify interactions
// Mock called but not verified
```

### 5. Reset Mocks Between Tests

```go
func TestWithMockReset(t *testing.T) {
    mock := NewMockService()
    
    t.Run("test 1", func(t *testing.T) {
        // Use mock
        mock.Reset() // ✅ Reset after use
    })
    
    t.Run("test 2", func(t *testing.T) {
        // Clean state
    })
}
```

## Common Patterns

### Pattern 1: Mock Factory

```go
func createMockComponent(options ...MockOption) *MockComponent {
    mock := &MockComponent{
        view: "Default View",
    }
    
    for _, opt := range options {
        opt(mock)
    }
    
    return mock
}

type MockOption func(*MockComponent)

func WithView(view string) MockOption {
    return func(m *MockComponent) {
        m.view = view
    }
}

// Usage
mock := createMockComponent(
    WithView("Custom View"),
)
```

### Pattern 2: Spy Pattern

```go
type SpyService struct {
    calls []Call
}

type Call struct {
    Method string
    Args   []interface{}
}

func (s *SpyService) GetUser(id string) (*User, error) {
    s.calls = append(s.calls, Call{
        Method: "GetUser",
        Args:   []interface{}{id},
    })
    
    return &User{}, nil
}

func (s *SpyService) AssertCalled(t *testing.T, method string, times int) {
    t.Helper()
    
    count := 0
    for _, call := range s.calls {
        if call.Method == method {
            count++
        }
    }
    
    assert.Equal(t, times, count)
}
```

### Pattern 3: Stub Pattern

```go
type StubService struct {
    GetUserFunc func(id string) (*User, error)
}

func (s *StubService) GetUser(id string) (*User, error) {
    if s.GetUserFunc != nil {
        return s.GetUserFunc(id)
    }
    return nil, errors.New("not implemented")
}

// Usage
stub := &StubService{
    GetUserFunc: func(id string) (*User, error) {
        return &User{Name: "Test"}, nil
    },
}
```

## Summary

Mocking enables:

- ✅ **Component isolation** - Test in isolation
- ✅ **Controlled behavior** - Simulate scenarios
- ✅ **Fast tests** - No real I/O
- ✅ **Edge case testing** - Errors and unusual conditions
- ✅ **Deterministic tests** - Reliable, repeatable

Use mocks to write focused, reliable tests that verify component behavior without external dependencies.

## See Also

- **[Quickstart](quickstart.md)** - Get started quickly
- **[Assertions](assertions.md)** - Assertion reference
- **[Snapshot Testing](snapshots.md)** - Regression testing
- **[Advanced Testing](../guides/advanced-testing.md)** - Advanced patterns
