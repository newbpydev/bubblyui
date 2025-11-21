package testutil

import (
	"reflect"
	"sync"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// MockStorage is a thread-safe in-memory storage implementation for testing.
// It implements the composables.Storage interface and provides additional
// testing utilities for inspecting stored data.
//
// Example:
//
//	mockStorage := NewMockStorage()
//	storage := composables.UseLocalStorage(ctx, "key", "initial", mockStorage)
//
//	// Later in tests
//	data, err := mockStorage.Load("key")
//	assert.NoError(t, err)
type MockStorage struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewMockStorage creates a new MockStorage instance.
//
// Returns:
//   - *MockStorage: A new mock storage instance
//
// Example:
//
//	mockStorage := NewMockStorage()
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

// Load retrieves data for the given key.
// Returns nil if the key doesn't exist.
//
// Parameters:
//   - key: The storage key
//
// Returns:
//   - []byte: The stored data
//   - error: Always nil for mock storage
func (ms *MockStorage) Load(key string) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	data, exists := ms.data[key]
	if !exists {
		return nil, nil
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// Save stores data for the given key.
//
// Parameters:
//   - key: The storage key
//   - value: The data to store
//
// Returns:
//   - error: Always nil for mock storage
func (ms *MockStorage) Save(key string, value []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Store a copy to prevent external modification
	data := make([]byte, len(value))
	copy(data, value)
	ms.data[key] = data

	return nil
}

// Clear removes data for the given key.
// This is a testing utility not part of the Storage interface.
//
// Parameters:
//   - key: The storage key to clear
func (ms *MockStorage) Clear(key string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.data, key)
}

// ClearAll removes all stored data.
// This is a testing utility not part of the Storage interface.
func (ms *MockStorage) ClearAll() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.data = make(map[string][]byte)
}

// UseLocalStorageTester provides utilities for testing local storage persistence.
// It integrates with the UseLocalStorage composable to test value persistence,
// storage operations, and state synchronization.
//
// This tester is specifically designed for testing components that use the UseLocalStorage
// composable. It allows you to:
//   - Set and get values
//   - Verify storage persistence
//   - Inspect stored data
//   - Clear storage
//   - Test with complex types
//
// The tester automatically extracts the storage state refs from the component,
// making it easy to assert on storage behavior at any point in the test.
//
// Example:
//
//	mockStorage := NewMockStorage()
//	comp := createStorageComponent(mockStorage)
//	tester := NewUseLocalStorageTester[string](comp, mockStorage)
//
//	// Set value
//	tester.SetValue("test data")
//
//	// Verify persistence
//	data := tester.GetStoredData("my-key")
//	assert.Contains(t, string(data), "test data")
//
// Thread Safety:
//
// UseLocalStorageTester is not thread-safe. It should only be used from a single test goroutine.
// However, the underlying MockStorage is thread-safe and can be safely accessed from multiple
// goroutines.
type UseLocalStorageTester[T any] struct {
	component bubbly.Component
	valueRef  interface{} // *Ref[T]
	set       func(T)
	get       func() T
	storage   composables.Storage
}

// NewUseLocalStorageTester creates a new UseLocalStorageTester for testing storage operations.
//
// The component must expose "value", "set", and "get" in its Setup function.
// These correspond to the fields returned by UseLocalStorage composable.
//
// Parameters:
//   - comp: The component to test (must expose storage state and methods)
//   - storage: The storage implementation (typically MockStorage)
//
// Returns:
//   - *UseLocalStorageTester[T]: A new tester instance
//
// Panics:
//   - If the component doesn't expose required refs or functions
//
// Example:
//
//	mockStorage := NewMockStorage()
//	comp, err := bubbly.NewComponent("TestStorage").
//	    Setup(func(ctx *bubbly.Context) {
//	        storage := composables.UseLocalStorage(ctx, "key", "initial", mockStorage)
//	        ctx.Expose("value", storage.Value)
//	        ctx.Expose("set", storage.Set)
//	        ctx.Expose("get", storage.Get)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseLocalStorageTester[string](comp, mockStorage)
func NewUseLocalStorageTester[T any](comp bubbly.Component, storage composables.Storage) *UseLocalStorageTester[T] {
	// Extract exposed values from component using reflection

	// Get value ref
	valueRef := extractExposedValue(comp, "value")
	if valueRef == nil {
		panic("component must expose 'value' ref")
	}

	// Extract set function
	setValue := extractExposedValue(comp, "set")
	if setValue == nil {
		panic("component must expose 'set' function")
	}
	set, ok := setValue.(func(T))
	if !ok {
		panic("'set' must be a function with signature func(T)")
	}

	// Extract get function
	getValue := extractExposedValue(comp, "get")
	if getValue == nil {
		panic("component must expose 'get' function")
	}
	get, ok := getValue.(func() T)
	if !ok {
		panic("'get' must be a function with signature func() T")
	}

	return &UseLocalStorageTester[T]{
		component: comp,
		valueRef:  valueRef,
		set:       set,
		get:       get,
		storage:   storage,
	}
}

// SetValue sets a new value in the storage.
// This triggers persistence to the underlying storage.
//
// Parameters:
//   - value: The value to set
//
// Example:
//
//	tester.SetValue("new data")
//	assert.Equal(t, "new data", tester.GetValue())
func (ulst *UseLocalStorageTester[T]) SetValue(value T) {
	ulst.set(value)
}

// GetValue returns the current value from the storage.
//
// Returns:
//   - T: The current value
//
// Example:
//
//	value := tester.GetValue()
//	assert.Equal(t, "expected", value)
func (ulst *UseLocalStorageTester[T]) GetValue() T {
	return ulst.get()
}

// GetStoredData returns the raw stored data for the given key.
// This allows inspection of the actual persisted data.
//
// Parameters:
//   - key: The storage key
//
// Returns:
//   - []byte: The raw stored data, or nil if not found
//
// Example:
//
//	data := tester.GetStoredData("my-key")
//	assert.Contains(t, string(data), "expected content")
func (ulst *UseLocalStorageTester[T]) GetStoredData(key string) []byte {
	data, _ := ulst.storage.Load(key)
	return data
}

// ClearStorage removes the stored data for the given key.
// This is useful for testing initialization behavior.
//
// Parameters:
//   - key: The storage key to clear
//
// Example:
//
//	tester.ClearStorage("my-key")
//	data := tester.GetStoredData("my-key")
//	assert.Nil(t, data)
func (ulst *UseLocalStorageTester[T]) ClearStorage(key string) {
	if mockStorage, ok := ulst.storage.(*MockStorage); ok {
		mockStorage.Clear(key)
	}
}

// GetValueFromRef returns the current value by reading directly from the ref.
// This is an alternative to GetValue() that uses reflection.
//
// Returns:
//   - T: The current value
//
// Example:
//
//	value := tester.GetValueFromRef()
//	assert.Equal(t, "expected", value)
func (ulst *UseLocalStorageTester[T]) GetValueFromRef() T {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(ulst.valueRef)
	if !v.IsValid() || v.IsNil() {
		var zero T
		return zero
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		var zero T
		return zero
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		var zero T
		return zero
	}

	// Return the typed value
	return result[0].Interface().(T)
}
