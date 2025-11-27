package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// TestUseStateTester_BasicOperations tests basic state operations
func TestUseStateTester_BasicOperations(t *testing.T) {
	comp, err := bubbly.NewComponent("TestState").
		Setup(func(ctx *bubbly.Context) {
			state := composables.UseState(ctx, "initial")

			ctx.Expose("value", state.Value)
			ctx.Expose("set", state.Set)
			ctx.Expose("get", state.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseStateTester[string](comp)

	// Initially has initial value
	assert.Equal(t, "initial", tester.GetValue())

	// Set new value
	tester.SetValue("new value")
	assert.Equal(t, "new value", tester.GetValue())
}

// TestUseStateTester_IntegerState tests integer state
func TestUseStateTester_IntegerState(t *testing.T) {
	comp, err := bubbly.NewComponent("TestState").
		Setup(func(ctx *bubbly.Context) {
			state := composables.UseState(ctx, 0)

			ctx.Expose("value", state.Value)
			ctx.Expose("set", state.Set)
			ctx.Expose("get", state.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseStateTester[int](comp)

	// Initially 0
	assert.Equal(t, 0, tester.GetValue())

	// Increment
	tester.SetValue(1)
	assert.Equal(t, 1, tester.GetValue())

	tester.SetValue(42)
	assert.Equal(t, 42, tester.GetValue())
}

// TestUseStateTester_StructState tests struct state
func TestUseStateTester_StructState(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	comp, err := bubbly.NewComponent("TestState").
		Setup(func(ctx *bubbly.Context) {
			state := composables.UseState(ctx, User{
				Name: "Initial",
				Age:  0,
			})

			ctx.Expose("value", state.Value)
			ctx.Expose("set", state.Set)
			ctx.Expose("get", state.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseStateTester[User](comp)

	// Initially has initial value
	user := tester.GetValue()
	assert.Equal(t, "Initial", user.Name)
	assert.Equal(t, 0, user.Age)

	// Set new value
	newUser := User{
		Name: "Alice",
		Age:  25,
	}
	tester.SetValue(newUser)

	user = tester.GetValue()
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, 25, user.Age)
}

// TestUseStateTester_GetValueFromRef tests alternative value access
func TestUseStateTester_GetValueFromRef(t *testing.T) {
	comp, err := bubbly.NewComponent("TestState").
		Setup(func(ctx *bubbly.Context) {
			state := composables.UseState(ctx, "test")

			ctx.Expose("value", state.Value)
			ctx.Expose("set", state.Set)
			ctx.Expose("get", state.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseStateTester[string](comp)

	// Both methods should return same value
	assert.Equal(t, "test", tester.GetValue())
	assert.Equal(t, "test", tester.GetValueFromRef())

	// After setting
	tester.SetValue("updated")
	assert.Equal(t, "updated", tester.GetValue())
	assert.Equal(t, "updated", tester.GetValueFromRef())
}

// TestUseStateTester_MissingRefs tests panic when required refs not exposed
func TestUseStateTester_MissingRefs(t *testing.T) {
	comp, err := bubbly.NewComponent("TestState").
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
		NewUseStateTester[string](comp)
	})
}

// TestUseStateTester_GetValueFromRef_EdgeCases tests error paths for GetValueFromRef
func TestUseStateTester_GetValueFromRef_EdgeCases(t *testing.T) {
	// Test GetValueFromRef with proper setup
	comp, err := bubbly.NewComponent("TestState").
		Setup(func(ctx *bubbly.Context) {
			state := composables.UseState(ctx, "test")
			ctx.Expose("value", state.Value)
			ctx.Expose("set", state.Set)
			ctx.Expose("get", state.Get)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseStateTester[string](comp)

	// Test normal case works
	assert.Equal(t, "test", tester.GetValueFromRef())

	// Test with empty string (edge case but valid)
	tester.SetValue("")
	assert.Equal(t, "", tester.GetValueFromRef())
}

// TestUseStateTester_GetValueFromRef_MultipleTypes tests GetValueFromRef with various types
func TestUseStateTester_GetValueFromRef_MultipleTypes(t *testing.T) {
	tests := []struct {
		name         string
		initialValue interface{}
		updatedValue interface{}
		setupComp    func(*bubbly.Context, interface{})
	}{
		{
			name:         "bool_type",
			initialValue: false,
			updatedValue: true,
			setupComp: func(ctx *bubbly.Context, val interface{}) {
				state := composables.UseState(ctx, val.(bool))
				ctx.Expose("value", state.Value)
				ctx.Expose("set", state.Set)
				ctx.Expose("get", state.Get)
			},
		},
		{
			name:         "float_type",
			initialValue: 3.14,
			updatedValue: 2.71,
			setupComp: func(ctx *bubbly.Context, val interface{}) {
				state := composables.UseState(ctx, val.(float64))
				ctx.Expose("value", state.Value)
				ctx.Expose("set", state.Set)
				ctx.Expose("get", state.Get)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := bubbly.NewComponent("TestState").
				Setup(func(ctx *bubbly.Context) {
					tt.setupComp(ctx, tt.initialValue)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)
			comp.Init()

			switch v := tt.initialValue.(type) {
			case bool:
				tester := NewUseStateTester[bool](comp)
				assert.Equal(t, v, tester.GetValueFromRef())
				tester.SetValue(tt.updatedValue.(bool))
				assert.Equal(t, tt.updatedValue.(bool), tester.GetValueFromRef())
			case float64:
				tester := NewUseStateTester[float64](comp)
				assert.Equal(t, v, tester.GetValueFromRef())
				tester.SetValue(tt.updatedValue.(float64))
				assert.Equal(t, tt.updatedValue.(float64), tester.GetValueFromRef())
			}
		})
	}
}
