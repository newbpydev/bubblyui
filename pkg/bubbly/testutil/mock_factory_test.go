package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewMockFactory tests factory creation
func TestNewMockFactory(t *testing.T) {
	factory := NewMockFactory()

	assert.NotNil(t, factory, "Factory should not be nil")
	assert.NotNil(t, factory.mocks, "Mocks map should be initialized")
}

// TestMockFactory_CreateMockRef tests creating mock refs with type safety
func TestMockFactory_CreateMockRef(t *testing.T) {
	tests := []struct {
		name     string
		mockName string
		initial  interface{}
		wantType string
	}{
		{
			name:     "create int ref",
			mockName: "counter",
			initial:  42,
			wantType: "int",
		},
		{
			name:     "create string ref",
			mockName: "message",
			initial:  "hello",
			wantType: "string",
		},
		{
			name:     "create bool ref",
			mockName: "enabled",
			initial:  true,
			wantType: "bool",
		},
		{
			name:     "create struct ref",
			mockName: "user",
			initial:  struct{ Name string }{Name: "Test"},
			wantType: "struct",
		},
		{
			name:     "create slice ref",
			mockName: "items",
			initial:  []int{1, 2, 3},
			wantType: "slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewMockFactory()

			// Create mock ref based on type
			switch v := tt.initial.(type) {
			case int:
				mockRef := CreateMockRef(factory, tt.mockName, v)
				assert.NotNil(t, mockRef, "MockRef should not be nil")
				assert.Equal(t, v, mockRef.Get(), "Initial value should match")
			case string:
				mockRef := CreateMockRef(factory, tt.mockName, v)
				assert.NotNil(t, mockRef, "MockRef should not be nil")
				assert.Equal(t, v, mockRef.Get(), "Initial value should match")
			case bool:
				mockRef := CreateMockRef(factory, tt.mockName, v)
				assert.NotNil(t, mockRef, "MockRef should not be nil")
				assert.Equal(t, v, mockRef.Get(), "Initial value should match")
			case struct{ Name string }:
				mockRef := CreateMockRef(factory, tt.mockName, v)
				assert.NotNil(t, mockRef, "MockRef should not be nil")
				assert.Equal(t, v, mockRef.Get(), "Initial value should match")
			case []int:
				mockRef := CreateMockRef(factory, tt.mockName, v)
				assert.NotNil(t, mockRef, "MockRef should not be nil")
				assert.Equal(t, v, mockRef.Get(), "Initial value should match")
			}
		})
	}
}

// TestMockFactory_CreateMockComponent tests creating mock components
func TestMockFactory_CreateMockComponent(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		wantName      string
	}{
		{
			name:          "create button component",
			componentName: "Button",
			wantName:      "Button",
		},
		{
			name:          "create input component",
			componentName: "Input",
			wantName:      "Input",
		},
		{
			name:          "create form component",
			componentName: "Form",
			wantName:      "Form",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewMockFactory()

			mockComp := factory.CreateMockComponent(tt.componentName)

			assert.NotNil(t, mockComp, "MockComponent should not be nil")
			assert.Equal(t, tt.wantName, mockComp.Name(), "Component name should match")
		})
	}
}

// TestMockFactory_GetMockRef tests retrieving created mock refs
func TestMockFactory_GetMockRef(t *testing.T) {
	factory := NewMockFactory()

	// Create a mock ref
	original := CreateMockRef(factory, "counter", 42)
	assert.NotNil(t, original)

	// Retrieve it
	retrieved := GetMockRef[int](factory, "counter")
	assert.NotNil(t, retrieved, "Should retrieve created mock ref")
	assert.Equal(t, 42, retrieved.Get(), "Retrieved ref should have same value")

	// Modify through retrieved ref
	retrieved.Set(100)
	assert.Equal(t, 100, original.Get(), "Should be same instance")
}

// TestMockFactory_GetMockRef_NotFound tests retrieving non-existent mock ref
func TestMockFactory_GetMockRef_NotFound(t *testing.T) {
	factory := NewMockFactory()

	retrieved := GetMockRef[int](factory, "nonexistent")
	assert.Nil(t, retrieved, "Should return nil for non-existent mock")
}

// TestMockFactory_GetMockComponent tests retrieving created mock components
func TestMockFactory_GetMockComponent(t *testing.T) {
	factory := NewMockFactory()

	// Create a mock component
	original := factory.CreateMockComponent("Button")
	assert.NotNil(t, original)

	// Retrieve it
	retrieved := factory.GetMockComponent("Button")
	assert.NotNil(t, retrieved, "Should retrieve created mock component")
	assert.Equal(t, "Button", retrieved.Name(), "Retrieved component should have same name")

	// Modify through retrieved component
	retrieved.SetViewOutput("Custom")
	assert.Equal(t, "Custom", original.View(), "Should be same instance")
}

// TestMockFactory_GetMockComponent_NotFound tests retrieving non-existent mock component
func TestMockFactory_GetMockComponent_NotFound(t *testing.T) {
	factory := NewMockFactory()

	retrieved := factory.GetMockComponent("nonexistent")
	assert.Nil(t, retrieved, "Should return nil for non-existent mock")
}

// TestMockFactory_Clear tests clearing all mocks
func TestMockFactory_Clear(t *testing.T) {
	factory := NewMockFactory()

	// Create some mocks
	CreateMockRef(factory, "counter", 42)
	CreateMockRef(factory, "message", "hello")
	factory.CreateMockComponent("Button")
	factory.CreateMockComponent("Input")

	// Verify they exist
	assert.NotNil(t, GetMockRef[int](factory, "counter"))
	assert.NotNil(t, GetMockRef[string](factory, "message"))
	assert.NotNil(t, factory.GetMockComponent("Button"))
	assert.NotNil(t, factory.GetMockComponent("Input"))

	// Clear all
	factory.Clear()

	// Verify they're gone
	assert.Nil(t, GetMockRef[int](factory, "counter"))
	assert.Nil(t, GetMockRef[string](factory, "message"))
	assert.Nil(t, factory.GetMockComponent("Button"))
	assert.Nil(t, factory.GetMockComponent("Input"))
}

// TestMockFactory_Clear_Idempotent tests that Clear can be called multiple times
func TestMockFactory_Clear_Idempotent(t *testing.T) {
	factory := NewMockFactory()

	CreateMockRef(factory, "counter", 42)

	// Clear multiple times should not panic
	factory.Clear()
	factory.Clear()
	factory.Clear()

	assert.Nil(t, GetMockRef[int](factory, "counter"))
}

// TestMockFactory_ThreadSafe tests concurrent access to factory
func TestMockFactory_ThreadSafe(t *testing.T) {
	factory := NewMockFactory()

	// Run concurrent operations
	done := make(chan bool)

	// Goroutine 1: Create refs
	go func() {
		for i := 0; i < 100; i++ {
			CreateMockRef(factory, "counter", i)
		}
		done <- true
	}()

	// Goroutine 2: Create components
	go func() {
		for i := 0; i < 100; i++ {
			factory.CreateMockComponent("Component")
		}
		done <- true
	}()

	// Goroutine 3: Get refs
	go func() {
		for i := 0; i < 100; i++ {
			GetMockRef[int](factory, "counter")
		}
		done <- true
	}()

	// Goroutine 4: Get components
	go func() {
		for i := 0; i < 100; i++ {
			factory.GetMockComponent("Component")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
	<-done

	// Should not panic and should have created mocks
	assert.NotNil(t, GetMockRef[int](factory, "counter"))
	assert.NotNil(t, factory.GetMockComponent("Component"))
}

// TestMockFactory_MultipleTypes tests creating refs of different types
func TestMockFactory_MultipleTypes(t *testing.T) {
	factory := NewMockFactory()

	// Create refs of different types
	intRef := CreateMockRef(factory, "int", 42)
	_ = CreateMockRef(factory, "string", "hello")
	_ = CreateMockRef(factory, "bool", true)

	// Retrieve and verify
	assert.Equal(t, 42, GetMockRef[int](factory, "int").Get())
	assert.Equal(t, "hello", GetMockRef[string](factory, "string").Get())
	assert.Equal(t, true, GetMockRef[bool](factory, "bool").Get())

	// Verify they're independent
	intRef.Set(100)
	assert.Equal(t, 100, GetMockRef[int](factory, "int").Get())
	assert.Equal(t, "hello", GetMockRef[string](factory, "string").Get())
}

// TestMockFactory_OverwriteExisting tests that creating with same name overwrites
func TestMockFactory_OverwriteExisting(t *testing.T) {
	factory := NewMockFactory()

	// Create initial mock
	first := CreateMockRef(factory, "counter", 42)
	assert.Equal(t, 42, first.Get())

	// Create with same name
	second := CreateMockRef(factory, "counter", 100)
	assert.Equal(t, 100, second.Get())

	// Retrieved should be the second one
	retrieved := GetMockRef[int](factory, "counter")
	assert.Equal(t, 100, retrieved.Get())

	// First ref should still have old value (different instance)
	assert.Equal(t, 42, first.Get())
}

// TestMockFactory_Integration tests realistic usage scenario
func TestMockFactory_Integration(t *testing.T) {
	// Setup: Create factory and mocks for a form component test
	factory := NewMockFactory()

	// Create mock refs for form state
	nameRef := CreateMockRef(factory, "name", "")
	emailRef := CreateMockRef(factory, "email", "")
	validRef := CreateMockRef(factory, "valid", false)

	// Create mock components
	_ = factory.CreateMockComponent("Input")
	_ = factory.CreateMockComponent("Button")

	// Simulate form interaction
	nameRef.Set("John Doe")
	emailRef.Set("john@example.com")
	validRef.Set(true)

	// Verify state
	assert.Equal(t, "John Doe", GetMockRef[string](factory, "name").Get())
	assert.Equal(t, "john@example.com", GetMockRef[string](factory, "email").Get())
	assert.Equal(t, true, GetMockRef[bool](factory, "valid").Get())

	// Verify components
	assert.Equal(t, "Input", factory.GetMockComponent("Input").Name())
	assert.Equal(t, "Button", factory.GetMockComponent("Button").Name())

	// Verify call tracking
	nameRef.AssertSetCalled(t, 1)
	emailRef.AssertSetCalled(t, 1)
	validRef.AssertSetCalled(t, 1)

	// Cleanup
	factory.Clear()
	assert.Nil(t, GetMockRef[string](factory, "name"))
}
