package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// TestNewUseEffectTester tests the constructor
func TestNewUseEffectTester(t *testing.T) {
	effectCalled := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return nil
			})
			ctx.Expose("effectCalled", &effectCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEffectTester(comp)
	assert.NotNil(t, tester, "tester should not be nil")
	assert.NotNil(t, tester.component, "tester component should not be nil")
}

// TestUseEffectTester_EffectRunsOnMount verifies effect runs when component mounts
func TestUseEffectTester_EffectRunsOnMount(t *testing.T) {
	effectCalled := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return nil
			})
			ctx.Expose("effectCalled", &effectCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Trigger mount

	tester := NewUseEffectTester(comp)

	// Effect should have run once on mount
	count := tester.GetEffectCallCount("effectCalled")
	assert.Equal(t, 1, count, "effect should run once on mount")
}

// TestUseEffectTester_TriggerUpdate tests TriggerUpdate method
func TestUseEffectTester_TriggerUpdate(t *testing.T) {
	effectCalled := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			// Effect with no deps - runs on every update
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return nil
			})
			ctx.Expose("effectCalled", &effectCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Trigger mount - effect runs once

	tester := NewUseEffectTester(comp)
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))

	// Trigger update
	tester.TriggerUpdate()
	assert.Equal(t, 2, tester.GetEffectCallCount("effectCalled"),
		"effect should run again on update")

	// Trigger multiple updates
	tester.TriggerUpdate()
	tester.TriggerUpdate()
	assert.Equal(t, 4, tester.GetEffectCallCount("effectCalled"),
		"effect should run on each update")
}

// TestUseEffectTester_CleanupOnRerun tests cleanup execution before re-running effect
func TestUseEffectTester_CleanupOnRerun(t *testing.T) {
	effectCalled := 0
	cleanupCalled := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return func() {
					cleanupCalled++
				}
			})
			ctx.Expose("effectCalled", &effectCalled)
			ctx.Expose("cleanupCalled", &cleanupCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Mount

	tester := NewUseEffectTester(comp)
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))
	assert.Equal(t, 0, tester.GetCleanupCallCount("cleanupCalled"))

	// Trigger update - should run cleanup then effect
	tester.TriggerUpdate()
	assert.Equal(t, 2, tester.GetEffectCallCount("effectCalled"))
	assert.Equal(t, 1, tester.GetCleanupCallCount("cleanupCalled"),
		"cleanup should run before re-running effect")

	// Another update
	tester.TriggerUpdate()
	assert.Equal(t, 3, tester.GetEffectCallCount("effectCalled"))
	assert.Equal(t, 2, tester.GetCleanupCallCount("cleanupCalled"))
}

// TestUseEffectTester_TriggerUnmount tests TriggerUnmount method
func TestUseEffectTester_TriggerUnmount(t *testing.T) {
	effectCalled := 0
	cleanupCalled := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return func() {
					cleanupCalled++
				}
			})
			ctx.Expose("effectCalled", &effectCalled)
			ctx.Expose("cleanupCalled", &cleanupCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Mount

	tester := NewUseEffectTester(comp)
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))
	assert.Equal(t, 0, tester.GetCleanupCallCount("cleanupCalled"))

	// Trigger unmount - should run final cleanup
	tester.TriggerUnmount()
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"),
		"effect should not run again on unmount")
	assert.Equal(t, 1, tester.GetCleanupCallCount("cleanupCalled"),
		"cleanup should run on unmount")
}

// TestUseEffectTester_SetRefValue tests SetRefValue method
func TestUseEffectTester_SetRefValue(t *testing.T) {
	effectCalled := 0
	var countRef *bubbly.Ref[int]

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			countRef = bubbly.NewRef(0)

			// Effect depends on count
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return nil
			}, countRef)

			ctx.Expose("count", countRef)
			ctx.Expose("effectCalled", &effectCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Mount - effect runs once

	tester := NewUseEffectTester(comp)
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))

	// Change ref value
	tester.SetRefValue("count", 5)
	assert.Equal(t, 5, countRef.Get(), "ref value should be updated")

	// Trigger update - effect should re-run since dependency changed
	tester.TriggerUpdate()
	assert.Equal(t, 2, tester.GetEffectCallCount("effectCalled"),
		"effect should re-run when dependency changes")
}

// TestUseEffectTester_MultipleEffects tests multiple independent effects
func TestUseEffectTester_MultipleEffects(t *testing.T) {
	effect1Called := 0
	effect2Called := 0
	cleanup1Called := 0
	cleanup2Called := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			// First effect
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effect1Called++
				return func() {
					cleanup1Called++
				}
			})

			// Second effect
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effect2Called++
				return func() {
					cleanup2Called++
				}
			})

			ctx.Expose("effect1Called", &effect1Called)
			ctx.Expose("effect2Called", &effect2Called)
			ctx.Expose("cleanup1Called", &cleanup1Called)
			ctx.Expose("cleanup2Called", &cleanup2Called)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Mount

	tester := NewUseEffectTester(comp)
	assert.Equal(t, 1, tester.GetEffectCallCount("effect1Called"))
	assert.Equal(t, 1, tester.GetEffectCallCount("effect2Called"))

	// Update - both effects should re-run
	tester.TriggerUpdate()
	assert.Equal(t, 2, tester.GetEffectCallCount("effect1Called"))
	assert.Equal(t, 2, tester.GetEffectCallCount("effect2Called"))
	assert.Equal(t, 1, tester.GetCleanupCallCount("cleanup1Called"))
	assert.Equal(t, 1, tester.GetCleanupCallCount("cleanup2Called"))

	// Unmount - final cleanup
	tester.TriggerUnmount()
	assert.Equal(t, 2, tester.GetCleanupCallCount("cleanup1Called"))
	assert.Equal(t, 2, tester.GetCleanupCallCount("cleanup2Called"))
}

// TestUseEffectTester_WithDepsNoChange tests effect with dependencies that don't change
func TestUseEffectTester_WithDepsNoChange(t *testing.T) {
	effectCalled := 0
	var countRef *bubbly.Ref[int]

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			countRef = bubbly.NewRef(0)

			// Effect with dependency that doesn't change
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return nil
			}, countRef)

			ctx.Expose("count", countRef)
			ctx.Expose("effectCalled", &effectCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Mount - effect runs once

	tester := NewUseEffectTester(comp)
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))

	// Update without changing dependency - effect should NOT re-run
	tester.TriggerUpdate()
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"),
		"effect should not re-run when dependency unchanged")

	tester.TriggerUpdate()
	assert.Equal(t, 1, tester.GetEffectCallCount("effectCalled"))
}

// TestUseEffectTester_NilCleanup tests effect with no cleanup function
func TestUseEffectTester_NilCleanup(t *testing.T) {
	effectCalled := 0

	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				effectCalled++
				return nil // No cleanup
			})
			ctx.Expose("effectCalled", &effectCalled)
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()
	comp.View() // Mount

	tester := NewUseEffectTester(comp)

	// Should not panic on update with nil cleanup
	assert.NotPanics(t, func() {
		tester.TriggerUpdate()
	})

	// Should not panic on unmount with nil cleanup
	assert.NotPanics(t, func() {
		tester.TriggerUnmount()
	})

	assert.Equal(t, 2, tester.GetEffectCallCount("effectCalled"))
}

// TestUseEffectTester_MissingCounter tests GetEffectCallCount with missing counter
func TestUseEffectTester_MissingCounter(t *testing.T) {
	comp, err := bubbly.NewComponent("TestEffect").
		Setup(func(ctx *bubbly.Context) {
			composables.UseEffect(ctx, func() composables.UseEffectCleanup {
				return nil
			})
			// No counter exposed
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()

	assert.NoError(t, err)
	comp.Init()

	tester := NewUseEffectTester(comp)

	// Should return 0 for missing counter, not panic
	count := tester.GetEffectCallCount("nonexistent")
	assert.Equal(t, 0, count, "should return 0 for missing counter")

	cleanupCount := tester.GetCleanupCallCount("nonexistent")
	assert.Equal(t, 0, cleanupCount, "should return 0 for missing cleanup counter")
}
