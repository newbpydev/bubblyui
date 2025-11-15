// Package testutil provides testing utilities for BubblyUI components.
package testutil

import (
	"reflect"
	"sync"
	"testing"
)

// MockRef is a mock implementation of Ref[T] for testing purposes.
// It tracks Get() and Set() calls and supports watchers for testing
// reactive behavior.
//
// MockRef provides assertion methods to verify how many times Get() and Set()
// were called, making it easy to test code that depends on reactive refs.
//
// Example usage:
//
//	mockRef := testutil.NewMockRef(42)
//	value := mockRef.Get()  // Returns 42
//	mockRef.AssertGetCalled(t, 1)
//
//	mockRef.Set(100)
//	mockRef.AssertSetCalled(t, 1)
type MockRef[T any] struct {
	mu       sync.RWMutex
	value    T
	getCalls int
	setCalls int
	watchers []func(T)
}

// NewMockRef creates a new MockRef with the given initial value.
// The mock ref is thread-safe and can be used in concurrent tests.
//
// Example:
//
//	intRef := NewMockRef(0)
//	stringRef := NewMockRef("hello")
//	structRef := NewMockRef(User{Name: "Test"})
func NewMockRef[T any](initial T) *MockRef[T] {
	return &MockRef[T]{
		value:    initial,
		watchers: []func(T){},
	}
}

// Get returns the current value and increments the get call counter.
// This method is thread-safe.
//
// Example:
//
//	mockRef := NewMockRef(42)
//	value := mockRef.Get()  // Returns 42
//	mockRef.AssertGetCalled(t, 1)
func (mr *MockRef[T]) Get() T {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.getCalls++
	return mr.value
}

// Set updates the value, increments the set call counter, and notifies watchers.
// Watchers are only notified if the value actually changed (using reflect.DeepEqual).
// This method is thread-safe.
//
// Example:
//
//	mockRef := NewMockRef(0)
//	mockRef.Set(42)
//	mockRef.AssertSetCalled(t, 1)
func (mr *MockRef[T]) Set(value T) {
	mr.mu.Lock()
	oldValue := mr.value
	mr.value = value
	mr.setCalls++

	// Copy watchers slice while holding lock
	watchersCopy := make([]func(T), len(mr.watchers))
	copy(watchersCopy, mr.watchers)
	mr.mu.Unlock()

	// Notify watchers outside the lock if value changed
	if !reflect.DeepEqual(oldValue, value) {
		for _, watcher := range watchersCopy {
			watcher(value)
		}
	}
}

// Watch registers a watcher function that will be called when the value changes.
// The watcher receives the new value as a parameter.
// This method is thread-safe.
//
// Example:
//
//	mockRef := NewMockRef(0)
//	called := false
//	mockRef.Watch(func(newVal int) {
//	    called = true
//	})
//	mockRef.Set(42)  // Triggers watcher
//	assert.True(t, called)
func (mr *MockRef[T]) Watch(fn func(T)) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.watchers = append(mr.watchers, fn)
}

// AssertGetCalled asserts that Get() was called exactly the specified number of times.
// If the assertion fails, the test will fail with a descriptive error message.
//
// Example:
//
//	mockRef := NewMockRef(42)
//	_ = mockRef.Get()
//	_ = mockRef.Get()
//	mockRef.AssertGetCalled(t, 2)  // Passes
//	mockRef.AssertGetCalled(t, 1)  // Fails with error
func (mr *MockRef[T]) AssertGetCalled(t *testing.T, times int) {
	t.Helper()

	mr.mu.RLock()
	actual := mr.getCalls
	mr.mu.RUnlock()

	if actual != times {
		t.Errorf("Get() called %d times, expected %d", actual, times)
	}
}

// AssertSetCalled asserts that Set() was called exactly the specified number of times.
// If the assertion fails, the test will fail with a descriptive error message.
//
// Example:
//
//	mockRef := NewMockRef(0)
//	mockRef.Set(10)
//	mockRef.Set(20)
//	mockRef.AssertSetCalled(t, 2)  // Passes
//	mockRef.AssertSetCalled(t, 1)  // Fails with error
func (mr *MockRef[T]) AssertSetCalled(t *testing.T, times int) {
	t.Helper()

	mr.mu.RLock()
	actual := mr.setCalls
	mr.mu.RUnlock()

	if actual != times {
		t.Errorf("Set() called %d times, expected %d", actual, times)
	}
}

// GetCallCount returns the number of times Get() has been called.
// This is useful for custom assertions or debugging.
// This method is thread-safe.
//
// Example:
//
//	mockRef := NewMockRef(42)
//	_ = mockRef.Get()
//	count := mockRef.GetCallCount()  // Returns 1
func (mr *MockRef[T]) GetCallCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.getCalls
}

// SetCallCount returns the number of times Set() has been called.
// This is useful for custom assertions or debugging.
// This method is thread-safe.
//
// Example:
//
//	mockRef := NewMockRef(0)
//	mockRef.Set(42)
//	count := mockRef.SetCallCount()  // Returns 1
func (mr *MockRef[T]) SetCallCount() int {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.setCalls
}

// Reset resets all call counters to zero.
// This is useful when reusing a mock ref across multiple test scenarios.
// This method is thread-safe.
//
// Example:
//
//	mockRef := NewMockRef(0)
//	mockRef.Set(42)
//	mockRef.Reset()
//	mockRef.AssertSetCalled(t, 0)  // Passes
func (mr *MockRef[T]) Reset() {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.getCalls = 0
	mr.setCalls = 0
}
