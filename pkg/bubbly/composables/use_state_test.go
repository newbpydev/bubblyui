package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// createTestContext creates a test context for composable testing
func createTestContext() *bubbly.Context {
	var ctx *bubbly.Context
	// Create a minimal component to get a valid context
	component, _ := bubbly.NewComponent("Test").
		Setup(func(c *bubbly.Context) {
			ctx = c
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()
	_ = component // Silence unused warning
	return ctx
}

func TestUseState_CreatesRefWithInitialValue(t *testing.T) {
	tests := []struct {
		name    string
		initial interface{}
	}{
		{"int", 42},
		{"string", "hello"},
		{"bool", true},
		{"float", 3.14},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()

			state := UseState(ctx, tt.initial)

			assert.NotNil(t, state.Value, "Value should not be nil")
			assert.Equal(t, tt.initial, state.Value.GetTyped(), "Initial value should match")
			assert.Equal(t, tt.initial, state.Get(), "Get() should return initial value")
		})
	}
}

func TestUseState_SetUpdatesValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		newValue int
	}{
		{"zero to positive", 0, 42},
		{"positive to negative", 10, -5},
		{"same value", 7, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()

			state := UseState(ctx, tt.initial)
			state.Set(tt.newValue)

			assert.Equal(t, tt.newValue, state.Value.GetTyped(), "Value should be updated")
			assert.Equal(t, tt.newValue, state.Get(), "Get() should return updated value")
		})
	}
}

func TestUseState_GetRetrievesValue(t *testing.T) {
	ctx := createTestContext()

	state := UseState(ctx, "initial")

	// Get should return current value
	assert.Equal(t, "initial", state.Get())

	// After set, Get should return new value
	state.Set("updated")
	assert.Equal(t, "updated", state.Get())
}

func TestUseState_TypeSafety(t *testing.T) {
	ctx := createTestContext()

	// Test with different types
	intState := UseState(ctx, 42)
	stringState := UseState(ctx, "hello")
	boolState := UseState(ctx, true)

	// Type assertions should work
	assert.IsType(t, 0, intState.Get())
	assert.IsType(t, "", stringState.Get())
	assert.IsType(t, true, boolState.Get())

	// Values should be correct
	assert.Equal(t, 42, intState.Get())
	assert.Equal(t, "hello", stringState.Get())
	assert.Equal(t, true, boolState.Get())
}

func TestUseState_MultipleInstancesIndependent(t *testing.T) {
	ctx := createTestContext()

	// Create two independent state instances
	state1 := UseState(ctx, 10)
	state2 := UseState(ctx, 20)

	// Verify initial values
	assert.Equal(t, 10, state1.Get())
	assert.Equal(t, 20, state2.Get())

	// Update state1
	state1.Set(100)

	// state1 should change, state2 should not
	assert.Equal(t, 100, state1.Get())
	assert.Equal(t, 20, state2.Get(), "state2 should remain unchanged")

	// Update state2
	state2.Set(200)

	// state2 should change, state1 should not be affected
	assert.Equal(t, 100, state1.Get(), "state1 should remain unchanged")
	assert.Equal(t, 200, state2.Get())
}

func TestUseState_SetAndGetConsistency(t *testing.T) {
	ctx := createTestContext()

	state := UseState(ctx, 0)

	// Multiple sets and gets
	for i := 0; i < 10; i++ {
		state.Set(i)
		assert.Equal(t, i, state.Get(), "Get should return last Set value")
		assert.Equal(t, i, state.Value.GetTyped(), "Value.Get should match")
	}
}

func TestUseState_WithStructType(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	ctx := createTestContext()

	initialUser := User{Name: "Alice", Age: 30}
	state := UseState(ctx, initialUser)

	// Check initial value
	assert.Equal(t, initialUser, state.Get())

	// Update with new struct
	newUser := User{Name: "Bob", Age: 25}
	state.Set(newUser)

	assert.Equal(t, newUser, state.Get())
	assert.Equal(t, "Bob", state.Get().Name)
	assert.Equal(t, 25, state.Get().Age)
}

func TestUseState_WithPointerType(t *testing.T) {
	type Data struct {
		Value int
	}

	ctx := createTestContext()

	initial := &Data{Value: 42}
	state := UseState(ctx, initial)

	// Check initial value
	assert.Equal(t, initial, state.Get())
	assert.Equal(t, 42, state.Get().Value)

	// Update with new pointer
	newData := &Data{Value: 100}
	state.Set(newData)

	assert.Equal(t, newData, state.Get())
	assert.Equal(t, 100, state.Get().Value)
}
