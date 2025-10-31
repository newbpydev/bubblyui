package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// unmountComponent is a helper to unmount a component for testing
func unmountComponent(comp bubbly.Component) {
	// Use type assertion to access unexported componentImpl
	type unmounter interface {
		Unmount()
	}
	if u, ok := comp.(unmounter); ok {
		u.Unmount()
	}
}

func TestUseEffect_RunsOnMount(t *testing.T) {
	effectRan := false

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			UseEffect(ctx, func() UseEffectCleanup {
				effectRan = true
				return nil
			})
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Init and first View should trigger mount
	comp.Init()
	comp.View()

	assert.True(t, effectRan, "Effect should run on mount")
}

func TestUseEffect_RunsOnDepsChange(t *testing.T) {
	var effectCount int
	var dep *bubbly.Ref[any]

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			dep = bubbly.NewRef[any](0)

			UseEffect(ctx, func() UseEffectCleanup {
				effectCount++
				return nil
			}, dep)

			ctx.Expose("dep", dep)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()
	assert.Equal(t, 1, effectCount, "Effect should run once on mount")

	// Update with dep change
	dep.Set(1)
	comp.Update(nil)
	assert.Equal(t, 2, effectCount, "Effect should run when dep changes")

	// Update with same dep value
	dep.Set(1)
	comp.Update(nil)
	assert.Equal(t, 2, effectCount, "Effect should NOT run when dep unchanged")

	// Update with different dep value
	dep.Set(2)
	comp.Update(nil)
	assert.Equal(t, 3, effectCount, "Effect should run when dep changes again")
}

func TestUseEffect_CleanupExecutesBeforeRerun(t *testing.T) {
	var executionOrder []string
	var dep *bubbly.Ref[any]

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			dep = bubbly.NewRef[any](0)

			UseEffect(ctx, func() UseEffectCleanup {
				executionOrder = append(executionOrder, "effect")
				return func() {
					executionOrder = append(executionOrder, "cleanup")
				}
			}, dep)

			ctx.Expose("dep", dep)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()
	assert.Equal(t, []string{"effect"}, executionOrder, "First effect should run")

	// Update with dep change - should run cleanup then effect
	dep.Set(1)
	comp.Update(nil)
	assert.Equal(t, []string{"effect", "cleanup", "effect"}, executionOrder,
		"Cleanup should run before re-running effect")
}

func TestUseEffect_CleanupExecutesOnUnmount(t *testing.T) {
	cleanupRan := false

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			UseEffect(ctx, func() UseEffectCleanup {
				return func() {
					cleanupRan = true
				}
			})
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()
	assert.False(t, cleanupRan, "Cleanup should not run yet")

	// Unmount
	unmountComponent(comp)
	assert.True(t, cleanupRan, "Cleanup should run on unmount")
}

func TestUseEffect_NoDepsRunsEveryUpdate(t *testing.T) {
	var effectCount int

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// No deps provided - should run on every update
			UseEffect(ctx, func() UseEffectCleanup {
				effectCount++
				return nil
			})
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()
	assert.Equal(t, 1, effectCount, "Effect should run on mount")

	// Update 1
	comp.Update(nil)
	assert.Equal(t, 2, effectCount, "Effect should run on first update")

	// Update 2
	comp.Update(nil)
	assert.Equal(t, 3, effectCount, "Effect should run on second update")

	// Update 3
	comp.Update(nil)
	assert.Equal(t, 4, effectCount, "Effect should run on third update")
}

func TestUseEffect_NilCleanup(t *testing.T) {
	var effectCount int

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			UseEffect(ctx, func() UseEffectCleanup {
				effectCount++
				return nil // No cleanup
			})
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Should not panic with nil cleanup
	comp.Init()
	comp.View()
	assert.Equal(t, 1, effectCount)

	comp.Update(nil)
	assert.Equal(t, 2, effectCount)

	// Unmount should not panic
	assert.NotPanics(t, func() {
		unmountComponent(comp)
	})
}

func TestUseEffect_MultipleEffectsIndependent(t *testing.T) {
	var effect1Count, effect2Count int
	var cleanup1Count, cleanup2Count int

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			// First effect
			UseEffect(ctx, func() UseEffectCleanup {
				effect1Count++
				return func() {
					cleanup1Count++
				}
			})

			// Second effect
			UseEffect(ctx, func() UseEffectCleanup {
				effect2Count++
				return func() {
					cleanup2Count++
				}
			})
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()
	assert.Equal(t, 1, effect1Count, "Effect 1 should run")
	assert.Equal(t, 1, effect2Count, "Effect 2 should run")

	// Update
	comp.Update(nil)
	assert.Equal(t, 2, effect1Count, "Effect 1 should run again")
	assert.Equal(t, 2, effect2Count, "Effect 2 should run again")
	assert.Equal(t, 1, cleanup1Count, "Cleanup 1 should run before re-run")
	assert.Equal(t, 1, cleanup2Count, "Cleanup 2 should run before re-run")

	// Unmount
	unmountComponent(comp)
	assert.Equal(t, 2, cleanup1Count, "Cleanup 1 should run on unmount")
	assert.Equal(t, 2, cleanup2Count, "Cleanup 2 should run on unmount")
}

func TestUseEffect_WithMultipleDeps(t *testing.T) {
	var effectCount int
	var dep1, dep2 *bubbly.Ref[any]

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			dep1 = bubbly.NewRef[any](0)
			dep2 = bubbly.NewRef[any](0)

			UseEffect(ctx, func() UseEffectCleanup {
				effectCount++
				return nil
			}, dep1, dep2)

			ctx.Expose("dep1", dep1)
			ctx.Expose("dep2", dep2)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()
	assert.Equal(t, 1, effectCount, "Effect should run on mount")

	// Change dep1
	dep1.Set(1)
	comp.Update(nil)
	assert.Equal(t, 2, effectCount, "Effect should run when dep1 changes")

	// Change dep2
	dep2.Set(1)
	comp.Update(nil)
	assert.Equal(t, 3, effectCount, "Effect should run when dep2 changes")

	// Change both
	dep1.Set(2)
	dep2.Set(2)
	comp.Update(nil)
	assert.Equal(t, 4, effectCount, "Effect should run when either dep changes")

	// No change
	comp.Update(nil)
	assert.Equal(t, 4, effectCount, "Effect should NOT run when deps unchanged")
}

func TestUseEffect_CleanupOrder(t *testing.T) {
	var executionOrder []string
	var dep *bubbly.Ref[any]

	comp, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			dep = bubbly.NewRef[any](0)

			UseEffect(ctx, func() UseEffectCleanup {
				executionOrder = append(executionOrder, "effect-start")
				return func() {
					executionOrder = append(executionOrder, "cleanup")
				}
			}, dep)

			ctx.Expose("dep", dep)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)

	// Mount
	comp.Init()
	comp.View()

	// Trigger multiple updates
	dep.Set(1)
	comp.Update(nil)

	dep.Set(2)
	comp.Update(nil)

	// Unmount
	unmountComponent(comp)

	expected := []string{
		"effect-start",            // Mount
		"cleanup", "effect-start", // First update
		"cleanup", "effect-start", // Second update
		"cleanup", // Unmount
	}

	assert.Equal(t, expected, executionOrder, "Cleanup should always run before next effect")
}
