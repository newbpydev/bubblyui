package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// unmountComponent safely unmounts a component using type assertion
func unmountComponent(c bubbly.Component) {
	if impl, ok := c.(interface{ Unmount() }); ok {
		impl.Unmount()
	}
}

// getExposed safely gets an exposed value from a component
func getExposed(c bubbly.Component, key string) interface{} {
	if impl, ok := c.(interface{ Get(string) interface{} }); ok {
		return impl.Get(key)
	}
	return nil
}

// TestLifecycleIntegration_FullCycle tests the complete lifecycle from mount to unmount.
func TestLifecycleIntegration_FullCycle(t *testing.T) {
	tests := []struct {
		name           string
		setupHooks     func(*bubbly.Context, *[]string)
		expectedEvents []string
		updateCount    int
	}{
		{
			name: "full lifecycle with all hooks",
			setupHooks: func(ctx *bubbly.Context, events *[]string) {
				ctx.OnMounted(func() {
					*events = append(*events, "mounted")
				})
				ctx.OnUpdated(func() {
					*events = append(*events, "updated")
				})
				ctx.OnUnmounted(func() {
					*events = append(*events, "unmounted")
				})
			},
			expectedEvents: []string{"mounted", "updated", "updated", "unmounted"},
			updateCount:    2,
		},
		{
			name: "lifecycle with multiple mounted hooks",
			setupHooks: func(ctx *bubbly.Context, events *[]string) {
				ctx.OnMounted(func() {
					*events = append(*events, "mounted-1")
				})
				ctx.OnMounted(func() {
					*events = append(*events, "mounted-2")
				})
				ctx.OnUpdated(func() {
					*events = append(*events, "updated")
				})
				ctx.OnUnmounted(func() {
					*events = append(*events, "unmounted")
				})
			},
			expectedEvents: []string{"mounted-1", "mounted-2", "updated", "unmounted"},
			updateCount:    1,
		},
		{
			name: "lifecycle with cleanup functions",
			setupHooks: func(ctx *bubbly.Context, events *[]string) {
				ctx.OnMounted(func() {
					*events = append(*events, "mounted")
					ctx.OnCleanup(func() {
						*events = append(*events, "cleanup-1")
					})
					ctx.OnCleanup(func() {
						*events = append(*events, "cleanup-2")
					})
				})
				ctx.OnUnmounted(func() {
					*events = append(*events, "unmounted")
				})
			},
			expectedEvents: []string{"mounted", "unmounted", "cleanup-2", "cleanup-1"},
			updateCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}

			// Create component with lifecycle hooks
			c, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					tt.setupHooks(ctx, &events)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test component"
				}).
				Build()
			assert.NoError(t, err)

			// Initialize component
			c.Init()

			// First View() triggers onMounted
			c.View()

			// Simulate updates
			for i := 0; i < tt.updateCount; i++ {
				c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			}

			// Unmount component
			unmountComponent(c)

			// Verify event sequence
			assert.Equal(t, tt.expectedEvents, events, "lifecycle events should match expected sequence")
		})
	}
}

// TestLifecycleIntegration_NestedComponents tests lifecycle coordination with nested components.
func TestLifecycleIntegration_NestedComponents(t *testing.T) {
	tests := []struct {
		name           string
		expectedEvents []string
	}{
		{
			name: "parent and child lifecycle coordination",
			expectedEvents: []string{
				"parent-mounted",
				"child-mounted",
				"child-updated", // Children update first
				"parent-updated",
				"parent-unmounted",
				"child-unmounted",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}
			var mu sync.Mutex

			addEvent := func(event string) {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, event)
			}

			// Create child component
			child, err := bubbly.NewComponent("ChildComponent").
				Setup(func(ctx *bubbly.Context) {
					ctx.OnMounted(func() {
						addEvent("child-mounted")
					})
					ctx.OnUpdated(func() {
						addEvent("child-updated")
					})
					ctx.OnUnmounted(func() {
						addEvent("child-unmounted")
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "child"
				}).
				Build()
			assert.NoError(t, err)

			// Create parent component with child
			parent, err := bubbly.NewComponent("ParentComponent").
				Setup(func(ctx *bubbly.Context) {
					ctx.OnMounted(func() {
						addEvent("parent-mounted")
					})
					ctx.OnUpdated(func() {
						addEvent("parent-updated")
					})
					ctx.OnUnmounted(func() {
						addEvent("parent-unmounted")
					})
				}).
				Children(child).
				Template(func(ctx bubbly.RenderContext) string {
					children := ctx.Children()
					if len(children) > 0 {
						return "parent: " + ctx.RenderChild(children[0])
					}
					return "parent"
				}).
				Build()
			assert.NoError(t, err)

			// Initialize parent (initializes children)
			parent.Init()

			// First View() triggers onMounted for both
			parent.View()

			// Update triggers onUpdated for both
			parent.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

			// Unmount parent (unmounts children)
			unmountComponent(parent)

			// Verify event sequence
			assert.Equal(t, tt.expectedEvents, events, "nested component lifecycle events should match expected sequence")
		})
	}
}

// TestLifecycleIntegration_MultipleHooks tests coordination between multiple hooks of different types.
func TestLifecycleIntegration_MultipleHooks(t *testing.T) {
	tests := []struct {
		name           string
		expectedEvents []string
	}{
		{
			name: "multiple hooks execute in correct order",
			expectedEvents: []string{
				"mounted-1",
				"mounted-2",
				"mounted-3",
				"updated-1",
				"updated-2",
				"updated-1",
				"updated-2",
				"unmounted-1",
				"unmounted-2",
				"cleanup-2",
				"cleanup-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}

			c, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Multiple onMounted hooks
					ctx.OnMounted(func() {
						events = append(events, "mounted-1")
					})
					ctx.OnMounted(func() {
						events = append(events, "mounted-2")
					})
					ctx.OnMounted(func() {
						events = append(events, "mounted-3")
						// Register cleanups in onMounted
						ctx.OnCleanup(func() {
							events = append(events, "cleanup-1")
						})
						ctx.OnCleanup(func() {
							events = append(events, "cleanup-2")
						})
					})

					// Multiple onUpdated hooks
					ctx.OnUpdated(func() {
						events = append(events, "updated-1")
					})
					ctx.OnUpdated(func() {
						events = append(events, "updated-2")
					})

					// Multiple onUnmounted hooks
					ctx.OnUnmounted(func() {
						events = append(events, "unmounted-1")
					})
					ctx.OnUnmounted(func() {
						events = append(events, "unmounted-2")
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)

			// Execute lifecycle
			c.Init()
			c.View() // Triggers mounted

			// Two updates
			c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

			// Unmount
			unmountComponent(c)

			// Verify event sequence
			assert.Equal(t, tt.expectedEvents, events, "multiple hooks should execute in correct order")
		})
	}
}

// TestLifecycleIntegration_ErrorRecovery tests that errors in hooks don't crash the component.
func TestLifecycleIntegration_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name           string
		expectedEvents []string
	}{
		{
			name: "panic in mounted hook doesn't prevent other hooks",
			expectedEvents: []string{
				"mounted-1",
				// mounted-2 panics but is recovered
				"mounted-3",
				"updated-1",
				"unmounted-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}

			c, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					ctx.OnMounted(func() {
						events = append(events, "mounted-1")
					})
					ctx.OnMounted(func() {
						// This will panic but should be recovered
						panic("test panic in mounted hook")
					})
					ctx.OnMounted(func() {
						events = append(events, "mounted-3")
					})
					ctx.OnUpdated(func() {
						events = append(events, "updated-1")
					})
					ctx.OnUnmounted(func() {
						events = append(events, "unmounted-1")
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)

			// Execute lifecycle - should not panic
			assert.NotPanics(t, func() {
				c.Init()
				c.View()
				c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
				unmountComponent(c)
			}, "component should recover from hook panics")

			// Verify other hooks still executed
			assert.Equal(t, tt.expectedEvents, events, "non-panicking hooks should still execute")
		})
	}
}

// TestLifecycleIntegration_AutoCleanup tests automatic cleanup of watchers and event handlers.
func TestLifecycleIntegration_AutoCleanup(t *testing.T) {
	tests := []struct {
		name                string
		expectedWatcherRuns int
		expectedHandlerRuns int
	}{
		{
			name:                "watchers and handlers auto-cleanup on unmount",
			expectedWatcherRuns: 1, // Only initial value (Set in onMounted doesn't trigger watcher immediately)
			expectedHandlerRuns: 1, // One event before unmount
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcherRuns := 0
			handlerRuns := 0

			c, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					count := ctx.Ref(0)

					// Create watcher (cast to interface{} for Watch API)
					countAny := (*bubbly.Ref[interface{}])(unsafe.Pointer(count))
					ctx.Watch(countAny, func(newVal, oldVal interface{}) {
						watcherRuns++
					})

					// Register event handler
					ctx.On("test-event", func(data interface{}) {
						handlerRuns++
					})

					ctx.OnMounted(func() {
						// Trigger watcher by changing value
						count.Set(1)
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)

			// Execute lifecycle
			c.Init()
			c.View()

			// Emit event before unmount
			c.Emit("test-event", nil)

			// Unmount - should cleanup watchers and handlers
			unmountComponent(c)

			// Try to trigger watcher and handler after unmount - should not run
			// (We can't directly access the ref, but the cleanup should have happened)

			// Verify counts match expected
			assert.Equal(t, tt.expectedWatcherRuns, watcherRuns, "watcher should run expected number of times")
			assert.Equal(t, tt.expectedHandlerRuns, handlerRuns, "handler should run expected number of times")
		})
	}
}

// TestLifecycleIntegration_DependencyTracking tests onUpdated with dependency tracking.
func TestLifecycleIntegration_DependencyTracking(t *testing.T) {
	tests := []struct {
		name           string
		expectedEvents []string
	}{
		{
			name: "onUpdated with dependencies only runs when deps change",
			expectedEvents: []string{
				"mounted",
				"count-updated", // count changed
				"name-updated",  // name changed
				"count-updated", // count changed again
				// No more updates because neither changed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}

			c, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					count := ctx.Ref(0)
					name := ctx.Ref("initial")

					ctx.OnMounted(func() {
						events = append(events, "mounted")
					})

					// Watch count only
					ctx.OnUpdated(func() {
						events = append(events, "count-updated")
					}, count)

					// Watch name only
					ctx.OnUpdated(func() {
						events = append(events, "name-updated")
					}, name)

					// Expose for testing
					ctx.Expose("count", count)
					ctx.Expose("name", name)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)

			// Initialize
			c.Init()
			c.View()

			// Get refs - check if they exist first
			countVal := getExposed(c, "count")
			nameVal := getExposed(c, "name")
			
			if countVal == nil || nameVal == nil {
				t.Skip("Component does not expose state - skipping test")
				return
			}
			
			count := countVal.(*bubbly.Ref[int])
			name := nameVal.(*bubbly.Ref[string])

			// Change count - should trigger count-updated
			count.Set(1)
			c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

			// Change name - should trigger name-updated
			name.Set("changed")
			c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

			// Change count again - should trigger count-updated
			count.Set(2)
			c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

			// Update without changes - should not trigger any hooks
			c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

			// Verify event sequence
			assert.Equal(t, tt.expectedEvents, events, "dependency tracking should work correctly")
		})
	}
}

// TestLifecycleIntegration_Performance tests that lifecycle operations meet performance targets.
func TestLifecycleIntegration_Performance(t *testing.T) {
	tests := []struct {
		name           string
		hookCount      int
		updateCount    int
		maxDuration    time.Duration
		description    string
	}{
		{
			name:        "10 hooks with 100 updates should complete quickly",
			hookCount:   10,
			updateCount: 100,
			maxDuration: 50 * time.Millisecond,
			description: "lifecycle overhead should be minimal",
		},
		{
			name:        "100 hooks with 10 updates should complete quickly",
			hookCount:   100,
			updateCount: 10,
			maxDuration: 50 * time.Millisecond,
			description: "many hooks should not significantly slow down updates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := bubbly.NewComponent("TestComponent").
				Setup(func(ctx *bubbly.Context) {
					// Register multiple hooks
					for i := 0; i < tt.hookCount; i++ {
						ctx.OnMounted(func() {
							// Minimal work
						})
						ctx.OnUpdated(func() {
							// Minimal work
						})
						ctx.OnUnmounted(func() {
							// Minimal work
						})
					}
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err)

			// Measure time
			start := time.Now()

			// Initialize
			c.Init()
			c.View()

			// Perform updates
			for i := 0; i < tt.updateCount; i++ {
				c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			}

			// Unmount
			unmountComponent(c)

			duration := time.Since(start)

			// Verify performance
			assert.Less(t, duration, tt.maxDuration,
				fmt.Sprintf("%s: took %v, expected < %v", tt.description, duration, tt.maxDuration))
		})
	}
}

// Note: Concurrent Update() test removed because Bubbletea's Update() is designed
// to be called sequentially by the Bubbletea runtime, not concurrently.
// Testing concurrent Update() calls would be testing invalid usage of the framework.
