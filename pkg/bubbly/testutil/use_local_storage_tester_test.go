package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/stretchr/testify/assert"
)

// TestUseLocalStorageTester_BasicOperations tests basic storage operations
func TestUseLocalStorageTester_BasicOperations(t *testing.T) {
	mockStorage := NewMockStorage()

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "test-key", "initial", mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[string](comp, mockStorage)

	// Initially has initial value
	assert.Equal(t, "initial", tester.GetValue())

	// Set new value
	tester.SetValue("new value")
	assert.Equal(t, "new value", tester.GetValue())

	// Verify stored in mock storage
	stored, err := mockStorage.Load("test-key")
	assert.NoError(t, err)
	assert.Contains(t, string(stored), "new value")
}

// TestUseLocalStorageTester_Persistence tests value persistence
func TestUseLocalStorageTester_Persistence(t *testing.T) {
	mockStorage := NewMockStorage()

	// Pre-populate storage
	mockStorage.Save("test-key", []byte(`"persisted value"`))

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "test-key", "initial", mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[string](comp, mockStorage)

	// Should load persisted value
	assert.Equal(t, "persisted value", tester.GetValue())
}

// TestUseLocalStorageTester_ComplexTypes tests storage with complex types
func TestUseLocalStorageTester_ComplexTypes(t *testing.T) {
	type User struct {
		Name  string
		Email string
		Age   int
	}

	mockStorage := NewMockStorage()

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "user-key", User{
				Name:  "Initial",
				Email: "initial@example.com",
				Age:   0,
			}, mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[User](comp, mockStorage)

	// Set complex value
	newUser := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}
	tester.SetValue(newUser)

	// Verify value
	value := tester.GetValue()
	assert.Equal(t, "Alice", value.Name)
	assert.Equal(t, "alice@example.com", value.Email)
	assert.Equal(t, 25, value.Age)

	// Verify persisted
	stored, err := mockStorage.Load("user-key")
	assert.NoError(t, err)
	assert.Contains(t, string(stored), "Alice")
	assert.Contains(t, string(stored), "alice@example.com")
}

// TestUseLocalStorageTester_GetStoredData tests direct storage access
func TestUseLocalStorageTester_GetStoredData(t *testing.T) {
	mockStorage := NewMockStorage()

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "test-key", "initial", mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[string](comp, mockStorage)

	// Set value
	tester.SetValue("stored data")

	// Get stored data directly
	data := tester.GetStoredData("test-key")
	assert.NotNil(t, data)
	assert.Contains(t, string(data), "stored data")
}

// TestUseLocalStorageTester_ClearStorage tests storage clearing
func TestUseLocalStorageTester_ClearStorage(t *testing.T) {
	mockStorage := NewMockStorage()

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "test-key", "initial", mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[string](comp, mockStorage)

	// Set value
	tester.SetValue("data to clear")
	assert.Equal(t, "data to clear", tester.GetValue())

	// Clear storage
	tester.ClearStorage("test-key")

	// Storage should be empty
	data := tester.GetStoredData("test-key")
	assert.Nil(t, data)
}

// TestUseLocalStorageTester_MissingRefs tests panic when required refs not exposed
func TestUseLocalStorageTester_MissingRefs(t *testing.T) {
	mockStorage := NewMockStorage()

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			// Don't expose required refs
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	assert.Panics(t, func() {
		NewUseLocalStorageTester[string](comp, mockStorage)
	})
}

// TestMockStorage_ClearAll verifies ClearAll clears all keys
func TestMockStorage_ClearAll(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockStorage)
		verifyKeys []string
	}{
		{
			name: "clear_empty_storage",
			setup: func(ms *MockStorage) {
				// No setup - empty storage
			},
			verifyKeys: []string{},
		},
		{
			name: "clear_single_key",
			setup: func(ms *MockStorage) {
				ms.Save("key1", []byte("value1"))
			},
			verifyKeys: []string{"key1"},
		},
		{
			name: "clear_multiple_keys",
			setup: func(ms *MockStorage) {
				ms.Save("key1", []byte("value1"))
				ms.Save("key2", []byte("value2"))
				ms.Save("key3", []byte("value3"))
			},
			verifyKeys: []string{"key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage()

			// Setup keys
			tt.setup(mockStorage)

			// Verify keys exist before clear
			for _, key := range tt.verifyKeys {
				data, err := mockStorage.Load(key)
				assert.NoError(t, err)
				assert.NotNil(t, data, "Key should exist before ClearAll")
			}

			// Clear all
			mockStorage.ClearAll()

			// Verify all keys are gone (Load returns nil, nil for non-existent keys)
			for _, key := range tt.verifyKeys {
				data, err := mockStorage.Load(key)
				assert.NoError(t, err, "Load should not error for non-existent keys")
				assert.Nil(t, data, "Key should not exist after ClearAll")
			}
		})
	}
}

// TestMockStorage_ClearAll_ThreadSafety verifies ClearAll is thread-safe
func TestMockStorage_ClearAll_ThreadSafety(t *testing.T) {
	mockStorage := NewMockStorage()

	// Pre-populate
	for i := 0; i < 100; i++ {
		mockStorage.Save(string(rune(i)), []byte("data"))
	}

	// Concurrent operations
	done := make(chan bool)

	go func() {
		mockStorage.ClearAll()
		done <- true
	}()

	go func() {
		mockStorage.Save("concurrent", []byte("data"))
		done <- true
	}()

	// Wait for both
	<-done
	<-done

	// No panic = success
}

// TestUseLocalStorageTester_GetValueFromRef tests reflection-based value retrieval
func TestUseLocalStorageTester_GetValueFromRef(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		setValue      string
		expectedValue string
	}{
		{
			name:          "get_initial_value",
			initialValue:  "initial",
			setValue:      "",
			expectedValue: "initial",
		},
		{
			name:          "get_updated_value",
			initialValue:  "initial",
			setValue:      "updated",
			expectedValue: "updated",
		},
		{
			name:          "get_empty_string",
			initialValue:  "",
			setValue:      "",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage()

			comp, err := bubbly.NewComponent("TestStorage").
				Setup(func(ctx *bubbly.Context) {
					storage := composables.UseLocalStorage(ctx, "test-key", tt.initialValue, mockStorage)

					ctx.Expose("value", storage.Value)
					ctx.Expose("set", storage.Set)
					ctx.Expose("get", storage.Get)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)
			comp.Init()

			tester := NewUseLocalStorageTester[string](comp, mockStorage)

			// Set value if specified
			if tt.setValue != "" {
				tester.SetValue(tt.setValue)
			}

			// Get value using reflection
			value := tester.GetValueFromRef()
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

// TestUseLocalStorageTester_GetValueFromRef_ComplexTypes tests reflection with complex types
func TestUseLocalStorageTester_GetValueFromRef_ComplexTypes(t *testing.T) {
	type Config struct {
		Theme    string
		Language string
		Enabled  bool
	}

	mockStorage := NewMockStorage()

	initialConfig := Config{
		Theme:    "dark",
		Language: "en",
		Enabled:  true,
	}

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "config-key", initialConfig, mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[Config](comp, mockStorage)

	// Get initial value via reflection
	value := tester.GetValueFromRef()
	assert.Equal(t, "dark", value.Theme)
	assert.Equal(t, "en", value.Language)
	assert.True(t, value.Enabled)

	// Update and verify
	newConfig := Config{
		Theme:    "light",
		Language: "es",
		Enabled:  false,
	}
	tester.SetValue(newConfig)

	value = tester.GetValueFromRef()
	assert.Equal(t, "light", value.Theme)
	assert.Equal(t, "es", value.Language)
	assert.False(t, value.Enabled)
}

// TestUseLocalStorageTester_GetValueFromRef_IntegerTypes tests reflection with integers
func TestUseLocalStorageTester_GetValueFromRef_IntegerTypes(t *testing.T) {
	mockStorage := NewMockStorage()

	comp, err := bubbly.NewComponent("TestStorage").
		Setup(func(ctx *bubbly.Context) {
			storage := composables.UseLocalStorage(ctx, "count-key", 42, mockStorage)

			ctx.Expose("value", storage.Value)
			ctx.Expose("set", storage.Set)
			ctx.Expose("get", storage.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseLocalStorageTester[int](comp, mockStorage)

	// Get value via reflection
	value := tester.GetValueFromRef()
	assert.Equal(t, 42, value)

	// Update
	tester.SetValue(100)
	value = tester.GetValueFromRef()
	assert.Equal(t, 100, value)
}
