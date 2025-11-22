package integration

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestSharedComposable_StateSharing verifies state changes are visible across all components
// using the same shared composable instance
func TestSharedComposable_StateSharing(t *testing.T) {
	// Create shared counter composable
	UseSharedCounter := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
		return composables.UseState(ctx, 0)
	})

	// Component A - Can increment counter
	componentA, err := bubbly.NewComponent("ComponentA").
		Setup(func(ctx *bubbly.Context) {
			counter := UseSharedCounter(ctx)
			ctx.Expose("counter", counter.Value)

			ctx.On("increment", func(data interface{}) {
				counter.Set(counter.Get() + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*bubbly.Ref[int])
			return fmt.Sprintf("Component A: %d", counter.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Component B - Reads counter (no increment)
	componentB, err := bubbly.NewComponent("ComponentB").
		Setup(func(ctx *bubbly.Context) {
			counter := UseSharedCounter(ctx)
			ctx.Expose("counter", counter.Value)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*bubbly.Ref[int])
			return fmt.Sprintf("Component B: %d", counter.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Mount both components with testutil harness
	harness := testutil.NewHarness(t)
	ctA := harness.Mount(componentA)
	ctB := harness.Mount(componentB)

	// Verify initial state (both should show 0)
	ctA.AssertRenderContains("Component A: 0")
	ctB.AssertRenderContains("Component B: 0")

	// Increment in Component A
	ctA.Emit("increment", nil)

	// Verify Component B sees the change (shared state)
	ctA.AssertRenderContains("Component A: 1")
	ctB.AssertRenderContains("Component B: 1")

	// Increment again
	ctA.Emit("increment", nil)
	ctA.Emit("increment", nil)

	// Both should show 3
	ctA.AssertRenderContains("Component A: 3")
	ctB.AssertRenderContains("Component B: 3")
}

// TestSharedComposable_SameInstance verifies that multiple components receive
// the exact same composable instance (singleton behavior)
func TestSharedComposable_SameInstance(t *testing.T) {
	// Track how many times factory is called
	var factoryCalls atomic.Int32

	// Create shared composable with factory call tracking
	UseSharedCounter := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
		factoryCalls.Add(1)
		return composables.UseState(ctx, 42)
	})

	// Mount multiple components with testutil harness
	harness := testutil.NewHarness(t)
	componentTests := make([]*testutil.ComponentTest, 5)
	for i := 0; i < 5; i++ {
		comp, err := bubbly.NewComponent(fmt.Sprintf("Component%d", i)).
			Setup(func(ctx *bubbly.Context) {
				counter := UseSharedCounter(ctx)
				ctx.Expose("counter", counter.Value)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				counter := ctx.Get("counter").(*bubbly.Ref[int])
				return fmt.Sprintf("Count: %d", counter.GetTyped())
			}).
			Build()

		require.NoError(t, err)
		componentTests[i] = harness.Mount(comp)
	}

	// Verify factory was called exactly once
	assert.Equal(t, int32(1), factoryCalls.Load(), "Factory should be called exactly once for singleton")

	// Verify all components show same initial value
	for i, ct := range componentTests {
		ct.AssertRenderContains("Count: 42")
		_ = i // Avoid unused variable
	}
}

// TestSharedComposable_ReactivityIntegration verifies that changes to shared state
// trigger reactive updates in all components using the shared composable
func TestSharedComposable_ReactivityIntegration(t *testing.T) {
	// Create shared state composable
	UseSharedState := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[string] {
		return composables.UseState(ctx, "initial")
	})

	// Component A - Can update state
	componentA, err := bubbly.NewComponent("UpdaterComponent").
		Setup(func(ctx *bubbly.Context) {
			state := UseSharedState(ctx)
			ctx.Expose("state", state.Value)

			ctx.On("update", func(data interface{}) {
				if newVal, ok := data.(string); ok {
					state.Set(newVal)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			state := ctx.Get("state").(*bubbly.Ref[string])
			return fmt.Sprintf("Updater: %s", state.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Component B - Reads state with computed value
	componentB, err := bubbly.NewComponent("ComputedComponent").
		Setup(func(ctx *bubbly.Context) {
			state := UseSharedState(ctx)
			ctx.Expose("state", state.Value)

			// Computed value based on shared state
			uppercase := ctx.Computed(func() interface{} {
				val := state.Value.GetTyped()
				return fmt.Sprintf("UPPERCASE: %s", val)
			})
			ctx.Expose("uppercase", uppercase)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			uppercase := ctx.Get("uppercase").(*bubbly.Computed[interface{}])
			return uppercase.GetTyped().(string)
		}).
		Build()

	require.NoError(t, err)

	// Mount components with testutil harness
	harness := testutil.NewHarness(t)
	ctA := harness.Mount(componentA)
	ctB := harness.Mount(componentB)

	// Verify initial state
	ctA.AssertRenderContains("Updater: initial")
	ctB.AssertRenderContains("UPPERCASE: initial")

	// Update state in Component A
	ctA.Emit("update", "changed")

	// Verify Component B's computed value updates (reactivity working)
	ctA.AssertRenderContains("Updater: changed")
	ctB.AssertRenderContains("UPPERCASE: changed")

	// Update again
	ctA.Emit("update", "final")

	ctA.AssertRenderContains("Updater: final")
	ctB.AssertRenderContains("UPPERCASE: final")
}

// TestSharedComposable_IndependentInstances verifies that multiple shared composables
// maintain independent state and don't interfere with each other
func TestSharedComposable_IndependentInstances(t *testing.T) {
	// Create two independent shared composables
	UseSharedCounterA := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
		return composables.UseState(ctx, 0)
	})

	UseSharedCounterB := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
		return composables.UseState(ctx, 100)
	})

	// Component using shared counter A
	componentA, err := bubbly.NewComponent("ComponentA").
		Setup(func(ctx *bubbly.Context) {
			counter := UseSharedCounterA(ctx)
			ctx.Expose("counter", counter.Value)

			ctx.On("increment", func(data interface{}) {
				counter.Set(counter.Get() + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*bubbly.Ref[int])
			return fmt.Sprintf("Counter A: %d", counter.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Component using shared counter B
	componentB, err := bubbly.NewComponent("ComponentB").
		Setup(func(ctx *bubbly.Context) {
			counter := UseSharedCounterB(ctx)
			ctx.Expose("counter", counter.Value)

			ctx.On("increment", func(data interface{}) {
				counter.Set(counter.Get() + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*bubbly.Ref[int])
			return fmt.Sprintf("Counter B: %d", counter.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Mount components with testutil harness
	harness := testutil.NewHarness(t)
	ctA := harness.Mount(componentA)
	ctB := harness.Mount(componentB)

	// Verify initial states are independent
	ctA.AssertRenderContains("Counter A: 0")
	ctB.AssertRenderContains("Counter B: 100")

	// Increment counter A
	ctA.Emit("increment", nil)
	ctA.Emit("increment", nil)

	// Verify counter A changed but counter B is independent
	ctA.AssertRenderContains("Counter A: 2")
	ctB.AssertRenderContains("Counter B: 100")

	// Now increment counter B
	ctB.Emit("increment", nil)

	// Verify both counters are independent
	ctA.AssertRenderContains("Counter A: 2")
	ctB.AssertRenderContains("Counter B: 101")
}

// TestSharedComposable_ThreadSafe verifies that concurrent component creation
// with shared composables is thread-safe and factory is called only once
func TestSharedComposable_ThreadSafe(t *testing.T) {
	// Track factory calls
	var factoryCalls atomic.Int32
	var mu sync.Mutex
	var initOrder []int

	// Create shared composable
	UseSharedCounter := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
		factoryCalls.Add(1)
		return composables.UseState(ctx, 0)
	})

	// Create harness for mounting (must be in main goroutine for testing.T)
	harness := testutil.NewHarness(t)

	// Launch multiple goroutines creating components concurrently
	const numGoroutines = 50
	var wg sync.WaitGroup
	components := make([]bubbly.Component, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			comp, err := bubbly.NewComponent(fmt.Sprintf("Component%d", index)).
				Setup(func(ctx *bubbly.Context) {
					counter := UseSharedCounter(ctx)
					ctx.Expose("counter", counter.Value)

					// Track initialization order
					mu.Lock()
					initOrder = append(initOrder, index)
					mu.Unlock()
				}).
				Template(func(ctx bubbly.RenderContext) string {
					counter := ctx.Get("counter").(*bubbly.Ref[int])
					return fmt.Sprintf("Count: %d", counter.GetTyped())
				}).
				Build()

			if err == nil {
				components[index] = comp
			}
		}(i)
	}

	wg.Wait()

	// Mount all components in main goroutine (required for testing.T)
	componentTests := make([]*testutil.ComponentTest, 0, numGoroutines)
	for _, comp := range components {
		if comp != nil {
			ct := harness.Mount(comp)
			componentTests = append(componentTests, ct)
		}
	}

	// Verify factory was called exactly once despite concurrent access
	assert.Equal(t, int32(1), factoryCalls.Load(), "Factory should be called exactly once despite concurrent component creation")

	// Verify all components initialized
	mu.Lock()
	assert.Equal(t, numGoroutines, len(initOrder), "All components should have initialized")
	mu.Unlock()

	// Verify all components show same initial value
	for _, ct := range componentTests {
		ct.AssertRenderContains("Count: 0")
	}
}

// TestSharedComposable_WithMultipleTypes verifies shared composables work with different types
func TestSharedComposable_WithMultipleTypes(t *testing.T) {
	tests := []struct {
		name         string
		createShared func() interface{}
		initialValue interface{}
		newValue     interface{}
		expectedInit string
		expectedNew  string
	}{
		{
			name: "int type",
			createShared: func() interface{} {
				return composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
					return composables.UseState(ctx, 42)
				})
			},
			initialValue: 42,
			newValue:     100,
			expectedInit: "Value: 42",
			expectedNew:  "Value: 100",
		},
		{
			name: "string type",
			createShared: func() interface{} {
				return composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[string] {
					return composables.UseState(ctx, "hello")
				})
			},
			initialValue: "hello",
			newValue:     "world",
			expectedInit: "Value: hello",
			expectedNew:  "Value: world",
		},
		{
			name: "bool type",
			createShared: func() interface{} {
				return composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[bool] {
					return composables.UseState(ctx, false)
				})
			},
			initialValue: false,
			newValue:     true,
			expectedInit: "Value: false",
			expectedNew:  "Value: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create shared composable
			sharedFactory := tt.createShared()

			// Component A
			componentA, err := bubbly.NewComponent("ComponentA").
				Setup(func(ctx *bubbly.Context) {
					var state interface{}
					switch f := sharedFactory.(type) {
					case func(*bubbly.Context) composables.UseStateReturn[int]:
						state = f(ctx)
					case func(*bubbly.Context) composables.UseStateReturn[string]:
						state = f(ctx)
					case func(*bubbly.Context) composables.UseStateReturn[bool]:
						state = f(ctx)
					}
					ctx.Expose("state", state)

					ctx.On("update", func(data interface{}) {
						switch s := state.(type) {
						case composables.UseStateReturn[int]:
							if val, ok := data.(int); ok {
								s.Set(val)
							}
						case composables.UseStateReturn[string]:
							if val, ok := data.(string); ok {
								s.Set(val)
							}
						case composables.UseStateReturn[bool]:
							if val, ok := data.(bool); ok {
								s.Set(val)
							}
						}
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					state := ctx.Get("state")
					switch s := state.(type) {
					case composables.UseStateReturn[int]:
						return fmt.Sprintf("Value: %d", s.Value.GetTyped())
					case composables.UseStateReturn[string]:
						return fmt.Sprintf("Value: %s", s.Value.GetTyped())
					case composables.UseStateReturn[bool]:
						return fmt.Sprintf("Value: %t", s.Value.GetTyped())
					}
					return "Unknown"
				}).
				Build()

			require.NoError(t, err)

			// Mount with testutil harness
			harness := testutil.NewHarness(t)
			ct := harness.Mount(componentA)

			// Verify initial value
			ct.AssertRenderContains(tt.expectedInit)

			// Update value
			ct.Emit("update", tt.newValue)

			// Verify new value
			ct.AssertRenderContains(tt.expectedNew)
		})
	}
}

// TestSharedComposable_PersistsAcrossComponentLifecycle verifies that shared state
// persists even when components are unmounted and new ones are created
func TestSharedComposable_PersistsAcrossComponentLifecycle(t *testing.T) {
	// Create shared counter
	UseSharedCounter := composables.CreateShared(func(ctx *bubbly.Context) composables.UseStateReturn[int] {
		return composables.UseState(ctx, 0)
	})

	// Create and use first component
	component1, err := bubbly.NewComponent("Component1").
		Setup(func(ctx *bubbly.Context) {
			counter := UseSharedCounter(ctx)
			ctx.Expose("counter", counter.Value)

			ctx.On("increment", func(data interface{}) {
				counter.Set(counter.Get() + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*bubbly.Ref[int])
			return fmt.Sprintf("Count: %d", counter.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Mount with testutil harness
	harness := testutil.NewHarness(t)
	ct1 := harness.Mount(component1)

	// Increment counter
	ct1.Emit("increment", nil)
	ct1.Emit("increment", nil)
	ct1.Emit("increment", nil)

	ct1.AssertRenderContains("Count: 3")

	// Unmount first component
	ct1.Unmount()

	// Small delay to ensure unmount completes
	time.Sleep(10 * time.Millisecond)

	// Create new component using same shared composable
	component2, err := bubbly.NewComponent("Component2").
		Setup(func(ctx *bubbly.Context) {
			counter := UseSharedCounter(ctx)
			ctx.Expose("counter", counter.Value)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*bubbly.Ref[int])
			return fmt.Sprintf("Count: %d", counter.GetTyped())
		}).
		Build()

	require.NoError(t, err)
	ct2 := harness.Mount(component2)

	// Verify state persisted (should still be 3)
	ct2.AssertRenderContains("Count: 3")
}
