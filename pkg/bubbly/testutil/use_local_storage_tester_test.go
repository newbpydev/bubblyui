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
