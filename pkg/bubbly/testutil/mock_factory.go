package testutil

import (
	"sync"
)

// MockFactory provides a centralized way to create and manage mock objects for testing.
// It maintains a registry of all created mocks and provides type-safe retrieval.
//
// MockFactory is thread-safe and can be used in concurrent tests.
//
// Example usage:
//
//	factory := testutil.NewMockFactory()
//
//	// Create mocks
//	countRef := factory.CreateMockRef("count", 0)
//	buttonComp := factory.CreateMockComponent("Button")
//
//	// Retrieve mocks
//	retrieved := factory.GetMockRef[int]("count")
//	retrieved.Set(42)
//
//	// Cleanup
//	factory.Clear()
type MockFactory struct {
	mu    sync.RWMutex
	mocks map[string]interface{}
}

// NewMockFactory creates a new MockFactory instance.
// The factory is initialized with an empty mock registry.
//
// Example:
//
//	factory := NewMockFactory()
//	defer factory.Clear()  // Cleanup after tests
func NewMockFactory() *MockFactory {
	return &MockFactory{
		mocks: make(map[string]interface{}),
	}
}

// CreateMockRef creates a new MockRef[T] with the given name and initial value.
// The mock is registered in the factory and can be retrieved later using GetMockRef.
//
// If a mock with the same name already exists, it will be overwritten.
//
// This function is thread-safe.
//
// Example:
//
//	factory := NewMockFactory()
//	countRef := CreateMockRef(factory, "count", 0)
//	nameRef := CreateMockRef(factory, "name", "John")
//	itemsRef := CreateMockRef(factory, "items", []string{"a", "b"})
func CreateMockRef[T any](mf *MockFactory, name string, initial T) *MockRef[T] {
	mockRef := NewMockRef(initial)

	mf.mu.Lock()
	mf.mocks[name] = mockRef
	mf.mu.Unlock()

	return mockRef
}

// CreateMockComponent creates a new MockComponent with the given name.
// The mock is registered in the factory and can be retrieved later using GetMockComponent.
//
// If a mock with the same name already exists, it will be overwritten.
//
// This method is thread-safe.
//
// Example:
//
//	factory := NewMockFactory()
//	button := factory.CreateMockComponent("Button")
//	input := factory.CreateMockComponent("Input")
//	form := factory.CreateMockComponent("Form")
func (mf *MockFactory) CreateMockComponent(name string) *MockComponent {
	mockComp := NewMockComponent(name)

	mf.mu.Lock()
	mf.mocks[name] = mockComp
	mf.mu.Unlock()

	return mockComp
}

// GetMockRef retrieves a previously created MockRef[T] by name.
// Returns nil if no mock with the given name exists or if the mock is not a MockRef[T].
//
// This function is thread-safe.
//
// Example:
//
//	factory := NewMockFactory()
//	CreateMockRef(factory, "count", 42)
//
//	// Later in test
//	countRef := GetMockRef[int](factory, "count")
//	if countRef != nil {
//	    countRef.Set(100)
//	}
func GetMockRef[T any](mf *MockFactory, name string) *MockRef[T] {
	mf.mu.RLock()
	mock, exists := mf.mocks[name]
	mf.mu.RUnlock()

	if !exists {
		return nil
	}

	// Type assertion to MockRef[T]
	mockRef, ok := mock.(*MockRef[T])
	if !ok {
		return nil
	}

	return mockRef
}

// GetMockComponent retrieves a previously created MockComponent by name.
// Returns nil if no mock with the given name exists or if the mock is not a MockComponent.
//
// This method is thread-safe.
//
// Example:
//
//	factory := NewMockFactory()
//	factory.CreateMockComponent("Button")
//
//	// Later in test
//	button := factory.GetMockComponent("Button")
//	if button != nil {
//	    button.SetViewOutput("Custom")
//	}
func (mf *MockFactory) GetMockComponent(name string) *MockComponent {
	mf.mu.RLock()
	mock, exists := mf.mocks[name]
	mf.mu.RUnlock()

	if !exists {
		return nil
	}

	// Type assertion to MockComponent
	mockComp, ok := mock.(*MockComponent)
	if !ok {
		return nil
	}

	return mockComp
}

// Clear removes all mocks from the factory.
// This is useful for cleanup between tests or test scenarios.
//
// This method is thread-safe and idempotent (can be called multiple times safely).
//
// Example:
//
//	func TestSomething(t *testing.T) {
//	    factory := NewMockFactory()
//	    defer factory.Clear()  // Cleanup after test
//
//	    // Create and use mocks
//	    factory.CreateMockRef("count", 0)
//	    // ... test code ...
//	}
func (mf *MockFactory) Clear() {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	// Create new map instead of clearing to ensure all references are released
	mf.mocks = make(map[string]interface{})
}
