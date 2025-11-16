package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/stretchr/testify/assert"
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
