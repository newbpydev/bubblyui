package bubbly

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewLifecycleManager tests the creation of a new LifecycleManager.
func TestNewLifecycleManager(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates lifecycle manager successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component for testing
			c := newComponentImpl("TestComponent")

			// Create lifecycle manager
			lm := newLifecycleManager(c)

			// Verify lifecycle manager is not nil
			assert.NotNil(t, lm, "lifecycle manager should not be nil")

			// Verify component reference is set
			assert.Equal(t, c, lm.component, "component reference should be set")

			// Verify hooks map is initialized
			assert.NotNil(t, lm.hooks, "hooks map should be initialized")

			// Verify cleanups slice is initialized
			assert.NotNil(t, lm.cleanups, "cleanups slice should be initialized")

			// Verify watchers slice is initialized
			assert.NotNil(t, lm.watchers, "watchers slice should be initialized")
		})
	}
}

// TestLifecycleManager_InitialState tests the initial state of a LifecycleManager.
func TestLifecycleManager_InitialState(t *testing.T) {
	tests := []struct {
		name                string
		expectedMounted     bool
		expectedUnmounting  bool
		expectedUpdateCount int
	}{
		{
			name:                "initial state is correct",
			expectedMounted:     false,
			expectedUnmounting:  false,
			expectedUpdateCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component for testing
			c := newComponentImpl("TestComponent")

			// Create lifecycle manager
			lm := newLifecycleManager(c)

			// Verify initial state flags
			assert.Equal(t, tt.expectedMounted, lm.mounted, "mounted should be false initially")
			assert.Equal(t, tt.expectedUnmounting, lm.unmounting, "unmounting should be false initially")
			assert.Equal(t, tt.expectedUpdateCount, lm.updateCount, "updateCount should be 0 initially")
		})
	}
}

// TestLifecycleManager_HooksMapInitialized tests that the hooks map is properly initialized.
func TestLifecycleManager_HooksMapInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "hooks map is initialized and empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component for testing
			c := newComponentImpl("TestComponent")

			// Create lifecycle manager
			lm := newLifecycleManager(c)

			// Verify hooks map is not nil
			assert.NotNil(t, lm.hooks, "hooks map should not be nil")

			// Verify hooks map is empty
			assert.Empty(t, lm.hooks, "hooks map should be empty initially")
		})
	}
}

// TestLifecycleManager_StateFlags tests the state flag fields.
func TestLifecycleManager_StateFlags(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "state flags are correct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component for testing
			c := newComponentImpl("TestComponent")

			// Create lifecycle manager
			lm := newLifecycleManager(c)

			// Verify mounted flag
			assert.False(t, lm.mounted, "mounted should be false")

			// Verify unmounting flag
			assert.False(t, lm.unmounting, "unmounting should be false")

			// Verify updateCount
			assert.Equal(t, 0, lm.updateCount, "updateCount should be 0")
		})
	}
}

// TestLifecycleManager_RegisterHook tests hook registration.
func TestLifecycleManager_RegisterHook(t *testing.T) {
	tests := []struct {
		name     string
		hookType string
		hookID   string
	}{
		{
			name:     "register mounted hook",
			hookType: "mounted",
			hookID:   "hook-1",
		},
		{
			name:     "register updated hook",
			hookType: "updated",
			hookID:   "hook-2",
		},
		{
			name:     "register unmounted hook",
			hookType: "unmounted",
			hookID:   "hook-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Create a test hook
			called := false
			hook := lifecycleHook{
				id: tt.hookID,
				callback: func() {
					called = true
				},
				order: 0,
			}

			// Register the hook
			lm.registerHook(tt.hookType, hook)

			// Verify hook was registered
			hooks := lm.hooks[tt.hookType]
			assert.Len(t, hooks, 1, "should have one hook registered")
			assert.Equal(t, tt.hookID, hooks[0].id, "hook ID should match")
			assert.Equal(t, 0, hooks[0].order, "hook order should be 0")

			// Verify callback is stored
			hooks[0].callback()
			assert.True(t, called, "callback should be callable")
		})
	}
}

// TestContext_OnMounted tests OnMounted hook registration.
func TestContext_OnMounted(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "register onMounted hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component with lifecycle
			c := newComponentImpl("TestComponent")
			c.lifecycle = newLifecycleManager(c)
			ctx := &Context{component: c}

			// Register hook
			called := false
			ctx.OnMounted(func() {
				called = true
			})

			// Verify hook was registered
			hooks := c.lifecycle.hooks["mounted"]
			assert.Len(t, hooks, 1, "should have one mounted hook")
			assert.NotEmpty(t, hooks[0].id, "hook should have an ID")

			// Verify callback works
			hooks[0].callback()
			assert.True(t, called, "callback should execute")
		})
	}
}

// TestContext_OnUpdated tests OnUpdated hook registration.
func TestContext_OnUpdated(t *testing.T) {
	tests := []struct {
		name     string
		withDeps bool
		depCount int
	}{
		{
			name:     "register onUpdated without dependencies",
			withDeps: false,
			depCount: 0,
		},
		{
			name:     "register onUpdated with one dependency",
			withDeps: true,
			depCount: 1,
		},
		{
			name:     "register onUpdated with multiple dependencies",
			withDeps: true,
			depCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component with lifecycle
			c := newComponentImpl("TestComponent")
			c.lifecycle = newLifecycleManager(c)
			ctx := &Context{component: c}

			// Create dependencies if needed
			var deps []*Ref[any]
			if tt.withDeps {
				for i := 0; i < tt.depCount; i++ {
					ref := NewRef[any](i)
					deps = append(deps, ref)
				}
			}

			// Register hook
			called := false
			ctx.OnUpdated(func() {
				called = true
			}, deps...)

			// Verify hook was registered
			hooks := c.lifecycle.hooks["updated"]
			assert.Len(t, hooks, 1, "should have one updated hook")
			assert.NotEmpty(t, hooks[0].id, "hook should have an ID")

			// Verify dependencies
			if tt.withDeps {
				assert.Len(t, hooks[0].dependencies, tt.depCount, "should have correct number of dependencies")
				assert.Len(t, hooks[0].lastValues, tt.depCount, "should have captured initial values")
			} else {
				assert.Empty(t, hooks[0].dependencies, "should have no dependencies")
			}

			// Verify callback works
			hooks[0].callback()
			assert.True(t, called, "callback should execute")
		})
	}
}

// TestContext_OnUnmounted tests OnUnmounted hook registration.
func TestContext_OnUnmounted(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "register onUnmounted hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component with lifecycle
			c := newComponentImpl("TestComponent")
			c.lifecycle = newLifecycleManager(c)
			ctx := &Context{component: c}

			// Register hook
			called := false
			ctx.OnUnmounted(func() {
				called = true
			})

			// Verify hook was registered
			hooks := c.lifecycle.hooks["unmounted"]
			assert.Len(t, hooks, 1, "should have one unmounted hook")
			assert.NotEmpty(t, hooks[0].id, "hook should have an ID")

			// Verify callback works
			hooks[0].callback()
			assert.True(t, called, "callback should execute")
		})
	}
}

// TestContext_MultipleHooks tests registering multiple hooks of the same type.
func TestContext_MultipleHooks(t *testing.T) {
	tests := []struct {
		name      string
		hookCount int
	}{
		{
			name:      "register two hooks",
			hookCount: 2,
		},
		{
			name:      "register five hooks",
			hookCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component with lifecycle
			c := newComponentImpl("TestComponent")
			c.lifecycle = newLifecycleManager(c)
			ctx := &Context{component: c}

			// Register multiple hooks
			callOrder := []int{}
			for i := 0; i < tt.hookCount; i++ {
				index := i
				ctx.OnMounted(func() {
					callOrder = append(callOrder, index)
				})
			}

			// Verify all hooks were registered
			hooks := c.lifecycle.hooks["mounted"]
			assert.Len(t, hooks, tt.hookCount, "should have all hooks registered")

			// Verify order is preserved
			for i := 0; i < tt.hookCount; i++ {
				assert.Equal(t, i, hooks[i].order, "hook order should match registration order")
			}

			// Verify execution order
			for i := 0; i < tt.hookCount; i++ {
				hooks[i].callback()
			}
			assert.Equal(t, tt.hookCount, len(callOrder), "all hooks should execute")
			for i := 0; i < tt.hookCount; i++ {
				assert.Equal(t, i, callOrder[i], "hooks should execute in registration order")
			}
		})
	}
}

// TestContext_OnCleanup tests cleanup function registration.
func TestContext_OnCleanup(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "register cleanup function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a component with lifecycle
			c := newComponentImpl("TestComponent")
			c.lifecycle = newLifecycleManager(c)
			ctx := &Context{component: c}

			// Register cleanup
			called := false
			ctx.OnCleanup(func() {
				called = true
			})

			// Verify cleanup was registered
			assert.Len(t, c.lifecycle.cleanups, 1, "should have one cleanup function")

			// Verify cleanup works
			c.lifecycle.cleanups[0]()
			assert.True(t, called, "cleanup should execute")
		})
	}
}

// TestLifecycleManager_IsMounted tests the IsMounted state query.
func TestLifecycleManager_IsMounted(t *testing.T) {
	tests := []struct {
		name           string
		initialMounted bool
		expectedResult bool
	}{
		{
			name:           "initial state is not mounted",
			initialMounted: false,
			expectedResult: false,
		},
		{
			name:           "mounted state returns true",
			initialMounted: true,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Set initial state
			if tt.initialMounted {
				lm.setMounted(true)
			}

			// Query state
			result := lm.IsMounted()

			// Verify result
			assert.Equal(t, tt.expectedResult, result, "IsMounted should return correct state")
		})
	}
}

// TestLifecycleManager_IsUnmounting tests the IsUnmounting state query.
func TestLifecycleManager_IsUnmounting(t *testing.T) {
	tests := []struct {
		name              string
		initialUnmounting bool
		expectedResult    bool
	}{
		{
			name:              "initial state is not unmounting",
			initialUnmounting: false,
			expectedResult:    false,
		},
		{
			name:              "unmounting state returns true",
			initialUnmounting: true,
			expectedResult:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Set initial state
			if tt.initialUnmounting {
				lm.setUnmounting(true)
			}

			// Query state
			result := lm.IsUnmounting()

			// Verify result
			assert.Equal(t, tt.expectedResult, result, "IsUnmounting should return correct state")
		})
	}
}

// TestLifecycleManager_StateTransitions tests state transitions.
func TestLifecycleManager_StateTransitions(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "can transition from unmounted to mounted",
		},
		{
			name: "can transition to unmounting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Verify initial state
			assert.False(t, lm.IsMounted(), "should start unmounted")
			assert.False(t, lm.IsUnmounting(), "should start not unmounting")

			// Transition to mounted
			lm.setMounted(true)
			assert.True(t, lm.IsMounted(), "should be mounted after setMounted(true)")
			assert.False(t, lm.IsUnmounting(), "should still not be unmounting")

			// Transition to unmounting
			lm.setUnmounting(true)
			assert.True(t, lm.IsMounted(), "should still be mounted")
			assert.True(t, lm.IsUnmounting(), "should be unmounting after setUnmounting(true)")

			// Can transition back
			lm.setMounted(false)
			lm.setUnmounting(false)
			assert.False(t, lm.IsMounted(), "should be unmounted after setMounted(false)")
			assert.False(t, lm.IsUnmounting(), "should not be unmounting after setUnmounting(false)")
		})
	}
}

// TestLifecycleManager_ThreadSafeState tests concurrent state access.
func TestLifecycleManager_ThreadSafeState(t *testing.T) {
	tests := []struct {
		name           string
		goroutineCount int
	}{
		{
			name:           "concurrent reads and writes",
			goroutineCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Run concurrent operations
			done := make(chan bool)
			for i := 0; i < tt.goroutineCount; i++ {
				go func(id int) {
					defer func() { done <- true }()

					// Perform multiple operations
					for j := 0; j < 100; j++ {
						if id%2 == 0 {
							// Even goroutines: set and read mounted
							lm.setMounted(j%2 == 0)
							_ = lm.IsMounted()
						} else {
							// Odd goroutines: set and read unmounting
							lm.setUnmounting(j%2 == 0)
							_ = lm.IsUnmounting()
						}
					}
				}(i)
			}

			// Wait for all goroutines
			for i := 0; i < tt.goroutineCount; i++ {
				<-done
			}

			// Verify final state is valid (no panics occurred)
			_ = lm.IsMounted()
			_ = lm.IsUnmounting()
		})
	}
}

// TestLifecycleManager_StatePersistence tests that state persists correctly.
func TestLifecycleManager_StatePersistence(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "state persists across multiple queries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Set mounted state
			lm.setMounted(true)

			// Query multiple times
			for i := 0; i < 10; i++ {
				assert.True(t, lm.IsMounted(), "mounted state should persist")
			}

			// Set unmounting state
			lm.setUnmounting(true)

			// Query multiple times
			for i := 0; i < 10; i++ {
				assert.True(t, lm.IsUnmounting(), "unmounting state should persist")
			}
		})
	}
}

// ============================================================================
// Task 2.1: onMounted Execution Tests
// ============================================================================

// TestLifecycleManager_ExecuteMounted tests the executeMounted method.
func TestLifecycleManager_ExecuteMounted(t *testing.T) {
	tests := []struct {
		name           string
		hookCount      int
		expectExecuted bool
	}{
		{
			name:           "executes single onMounted hook",
			hookCount:      1,
			expectExecuted: true,
		},
		{
			name:           "executes multiple onMounted hooks",
			hookCount:      3,
			expectExecuted: true,
		},
		{
			name:           "handles no hooks gracefully",
			hookCount:      0,
			expectExecuted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Track execution
			executionCount := 0

			// Register hooks
			for i := 0; i < tt.hookCount; i++ {
				hook := lifecycleHook{
					id:       fmt.Sprintf("hook-%d", i),
					callback: func() { executionCount++ },
					order:    i,
				}
				lm.registerHook("mounted", hook)
			}

			// Execute mounted hooks
			lm.executeMounted()

			// Verify execution
			if tt.expectExecuted {
				assert.Equal(t, tt.hookCount, executionCount, "all hooks should execute")
			} else {
				assert.Equal(t, 0, executionCount, "no hooks should execute")
			}

			// Verify mounted state is set
			assert.True(t, lm.IsMounted(), "component should be marked as mounted")
		})
	}
}

// TestLifecycleManager_ExecuteMounted_OnlyOnce tests that onMounted hooks only execute once.
func TestLifecycleManager_ExecuteMounted_OnlyOnce(t *testing.T) {
	tests := []struct {
		name      string
		callCount int
	}{
		{
			name:      "executes only once when called multiple times",
			callCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Track execution
			executionCount := 0

			// Register hook
			hook := lifecycleHook{
				id:       "hook-1",
				callback: func() { executionCount++ },
				order:    0,
			}
			lm.registerHook("mounted", hook)

			// Call executeMounted multiple times
			for i := 0; i < tt.callCount; i++ {
				lm.executeMounted()
			}

			// Verify hook executed only once
			assert.Equal(t, 1, executionCount, "hook should execute only once")
		})
	}
}

// TestLifecycleManager_ExecuteMounted_Order tests that hooks execute in registration order.
func TestLifecycleManager_ExecuteMounted_Order(t *testing.T) {
	tests := []struct {
		name      string
		hookCount int
	}{
		{
			name:      "executes hooks in registration order",
			hookCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Track execution order
			var executionOrder []int

			// Register hooks
			for i := 0; i < tt.hookCount; i++ {
				index := i // Capture loop variable
				hook := lifecycleHook{
					id:       fmt.Sprintf("hook-%d", i),
					callback: func() { executionOrder = append(executionOrder, index) },
					order:    i,
				}
				lm.registerHook("mounted", hook)
			}

			// Execute mounted hooks
			lm.executeMounted()

			// Verify execution order
			assert.Len(t, executionOrder, tt.hookCount, "all hooks should execute")
			for i := 0; i < tt.hookCount; i++ {
				assert.Equal(t, i, executionOrder[i], "hooks should execute in registration order")
			}
		})
	}
}

// TestLifecycleManager_ExecuteMounted_PanicRecovery tests panic recovery in hooks.
func TestLifecycleManager_ExecuteMounted_PanicRecovery(t *testing.T) {
	tests := []struct {
		name           string
		panicHookIndex int
		totalHooks     int
	}{
		{
			name:           "recovers from panic in first hook",
			panicHookIndex: 0,
			totalHooks:     3,
		},
		{
			name:           "recovers from panic in middle hook",
			panicHookIndex: 1,
			totalHooks:     3,
		},
		{
			name:           "recovers from panic in last hook",
			panicHookIndex: 2,
			totalHooks:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Track execution
			executionCount := 0

			// Register hooks
			for i := 0; i < tt.totalHooks; i++ {
				index := i
				hook := lifecycleHook{
					id:    fmt.Sprintf("hook-%d", i),
					order: i,
				}

				if index == tt.panicHookIndex {
					// This hook will panic
					hook.callback = func() {
						executionCount++
						panic("test panic")
					}
				} else {
					// Normal hook
					hook.callback = func() {
						executionCount++
					}
				}

				lm.registerHook("mounted", hook)
			}

			// Execute mounted hooks - should not panic
			assert.NotPanics(t, func() {
				lm.executeMounted()
			}, "executeMounted should recover from panics")

			// Verify all hooks were attempted
			assert.Equal(t, tt.totalHooks, executionCount, "all hooks should be attempted despite panic")

			// Verify component is still marked as mounted
			assert.True(t, lm.IsMounted(), "component should be marked as mounted despite panic")
		})
	}
}

// TestComponent_View_TriggersMounted tests that View() triggers onMounted hooks.
func TestComponent_View_TriggersMounted(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "first View() call triggers onMounted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track execution
			executed := false

			// Create component with setup and template
			c, err := NewComponent("TestComponent").
				Setup(func(ctx *Context) {
					ctx.OnMounted(func() {
						executed = true
					})
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err, "component build should not error")

			// Initialize component
			c.Init()

			// Verify hook not executed yet
			assert.False(t, executed, "hook should not execute before View()")

			// Call View() - should trigger onMounted
			c.View()

			// Verify hook executed
			assert.True(t, executed, "hook should execute on first View()")
		})
	}
}

// TestComponent_View_OnlyTriggersOnce tests that onMounted only triggers on first View().
func TestComponent_View_OnlyTriggersOnce(t *testing.T) {
	tests := []struct {
		name      string
		viewCalls int
	}{
		{
			name:      "onMounted triggers only on first View() call",
			viewCalls: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track execution count
			executionCount := 0

			// Create component with setup and template
			c, err := NewComponent("TestComponent").
				Setup(func(ctx *Context) {
					ctx.OnMounted(func() {
						executionCount++
					})
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()
			assert.NoError(t, err, "component build should not error")

			// Initialize component
			c.Init()

			// Call View() multiple times
			for i := 0; i < tt.viewCalls; i++ {
				c.View()
			}

			// Verify hook executed only once
			assert.Equal(t, 1, executionCount, "hook should execute only once")
		})
	}
}

// TestLifecycleManager_ExecuteUpdated tests the executeUpdated method.
func TestLifecycleManager_ExecuteUpdated(t *testing.T) {
	tests := []struct {
		name          string
		setupHooks    func(*LifecycleManager, *int)
		expectedCount int
		description   string
	}{
		{
			name: "no hooks registered",
			setupHooks: func(lm *LifecycleManager, count *int) {
				// No hooks registered
			},
			expectedCount: 0,
			description:   "should not execute any hooks when none registered",
		},
		{
			name: "single hook no dependencies",
			setupHooks: func(lm *LifecycleManager, count *int) {
				lm.registerHook("updated", lifecycleHook{
					id:       "hook-1",
					callback: func() { *count++ },
					order:    0,
				})
			},
			expectedCount: 1,
			description:   "should execute hook without dependencies",
		},
		{
			name: "multiple hooks no dependencies",
			setupHooks: func(lm *LifecycleManager, count *int) {
				lm.registerHook("updated", lifecycleHook{
					id:       "hook-1",
					callback: func() { *count++ },
					order:    0,
				})
				lm.registerHook("updated", lifecycleHook{
					id:       "hook-2",
					callback: func() { *count++ },
					order:    1,
				})
				lm.registerHook("updated", lifecycleHook{
					id:       "hook-3",
					callback: func() { *count++ },
					order:    2,
				})
			},
			expectedCount: 3,
			description:   "should execute all hooks without dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true) // Mark as mounted

			// Setup hooks
			executionCount := 0
			tt.setupHooks(lm, &executionCount)

			// Execute updated hooks
			lm.executeUpdated()

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteUpdated_WithDependencies tests dependency tracking.
func TestLifecycleManager_ExecuteUpdated_WithDependencies(t *testing.T) {
	tests := []struct {
		name          string
		setupHook     func(*LifecycleManager, *int, *Ref[any])
		changeValue   bool
		expectedCount int
		description   string
	}{
		{
			name: "single dependency - value changed",
			setupHook: func(lm *LifecycleManager, count *int, ref *Ref[any]) {
				// Capture initial value
				initialValue := ref.Get()
				lm.registerHook("updated", lifecycleHook{
					id:           "hook-1",
					callback:     func() { *count++ },
					dependencies: []*Ref[any]{ref},
					lastValues:   []any{initialValue},
					order:        0,
				})
			},
			changeValue:   true,
			expectedCount: 1,
			description:   "should execute when dependency changes",
		},
		{
			name: "single dependency - value unchanged",
			setupHook: func(lm *LifecycleManager, count *int, ref *Ref[any]) {
				// Capture initial value
				initialValue := ref.Get()
				lm.registerHook("updated", lifecycleHook{
					id:           "hook-1",
					callback:     func() { *count++ },
					dependencies: []*Ref[any]{ref},
					lastValues:   []any{initialValue},
					order:        0,
				})
			},
			changeValue:   false,
			expectedCount: 0,
			description:   "should not execute when dependency unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Create ref with initial value (using any type)
			ref := NewRef[any](10)

			// Setup hook
			executionCount := 0
			tt.setupHook(lm, &executionCount, ref)

			// Change value if needed
			if tt.changeValue {
				ref.Set(20)
			}

			// Execute updated hooks
			lm.executeUpdated()

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteUpdated_MultipleDependencies tests multiple dependencies.
func TestLifecycleManager_ExecuteUpdated_MultipleDependencies(t *testing.T) {
	tests := []struct {
		name          string
		changeFirst   bool
		changeSecond  bool
		expectedCount int
		description   string
	}{
		{
			name:          "both dependencies unchanged",
			changeFirst:   false,
			changeSecond:  false,
			expectedCount: 0,
			description:   "should not execute when no dependencies change",
		},
		{
			name:          "first dependency changed",
			changeFirst:   true,
			changeSecond:  false,
			expectedCount: 1,
			description:   "should execute when first dependency changes",
		},
		{
			name:          "second dependency changed",
			changeFirst:   false,
			changeSecond:  true,
			expectedCount: 1,
			description:   "should execute when second dependency changes",
		},
		{
			name:          "both dependencies changed",
			changeFirst:   true,
			changeSecond:  true,
			expectedCount: 1,
			description:   "should execute once when both dependencies change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Create refs with initial values (using any type)
			ref1 := NewRef[any](10)
			ref2 := NewRef[any]("hello")

			// Setup hook with multiple dependencies
			executionCount := 0
			lm.registerHook("updated", lifecycleHook{
				id:       "hook-1",
				callback: func() { executionCount++ },
				dependencies: []*Ref[any]{
					ref1,
					ref2,
				},
				lastValues: []any{ref1.Get(), ref2.Get()},
				order:      0,
			})

			// Change values if needed
			if tt.changeFirst {
				ref1.Set(20)
			}
			if tt.changeSecond {
				ref2.Set("world")
			}

			// Execute updated hooks
			lm.executeUpdated()

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteUpdated_Order tests execution order.
func TestLifecycleManager_ExecuteUpdated_Order(t *testing.T) {
	tests := []struct {
		name     string
		numHooks int
	}{
		{
			name:     "three hooks execute in order",
			numHooks: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Track execution order
			var executionOrder []int

			// Register hooks in order
			for i := 0; i < tt.numHooks; i++ {
				hookNum := i
				lm.registerHook("updated", lifecycleHook{
					id:       fmt.Sprintf("hook-%d", i),
					callback: func() { executionOrder = append(executionOrder, hookNum) },
					order:    i,
				})
			}

			// Execute updated hooks
			lm.executeUpdated()

			// Verify execution order
			assert.Equal(t, []int{0, 1, 2}, executionOrder, "hooks should execute in registration order")
		})
	}
}

// TestLifecycleManager_ExecuteUpdated_PanicRecovery tests panic recovery.
func TestLifecycleManager_ExecuteUpdated_PanicRecovery(t *testing.T) {
	tests := []struct {
		name          string
		panicInFirst  bool
		expectedCount int
		description   string
	}{
		{
			name:          "first hook panics, second still executes",
			panicInFirst:  true,
			expectedCount: 1,
			description:   "second hook should execute even if first panics",
		},
		{
			name:          "no panics, both execute",
			panicInFirst:  false,
			expectedCount: 2,
			description:   "both hooks should execute normally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Track execution count
			executionCount := 0

			// First hook - may panic
			lm.registerHook("updated", lifecycleHook{
				id: "hook-1",
				callback: func() {
					if tt.panicInFirst {
						panic("test panic")
					}
					executionCount++
				},
				order: 0,
			})

			// Second hook - always executes
			lm.registerHook("updated", lifecycleHook{
				id:       "hook-2",
				callback: func() { executionCount++ },
				order:    1,
			})

			// Execute updated hooks (should not panic)
			assert.NotPanics(t, func() {
				lm.executeUpdated()
			}, "executeUpdated should not panic")

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteUnmounted tests the executeUnmounted method.
func TestLifecycleManager_ExecuteUnmounted(t *testing.T) {
	tests := []struct {
		name          string
		setupHooks    func(*LifecycleManager, *int)
		expectedCount int
		description   string
	}{
		{
			name: "no hooks registered",
			setupHooks: func(lm *LifecycleManager, count *int) {
				// No hooks registered
			},
			expectedCount: 0,
			description:   "should not execute any hooks when none registered",
		},
		{
			name: "single onUnmounted hook",
			setupHooks: func(lm *LifecycleManager, count *int) {
				lm.registerHook("unmounted", lifecycleHook{
					id:       "hook-1",
					callback: func() { *count++ },
					order:    0,
				})
			},
			expectedCount: 1,
			description:   "should execute single unmounted hook",
		},
		{
			name: "multiple onUnmounted hooks",
			setupHooks: func(lm *LifecycleManager, count *int) {
				lm.registerHook("unmounted", lifecycleHook{
					id:       "hook-1",
					callback: func() { *count++ },
					order:    0,
				})
				lm.registerHook("unmounted", lifecycleHook{
					id:       "hook-2",
					callback: func() { *count++ },
					order:    1,
				})
				lm.registerHook("unmounted", lifecycleHook{
					id:       "hook-3",
					callback: func() { *count++ },
					order:    2,
				})
			},
			expectedCount: 3,
			description:   "should execute all unmounted hooks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Setup hooks
			executionCount := 0
			tt.setupHooks(lm, &executionCount)

			// Execute unmounted hooks
			lm.executeUnmounted()

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)

			// Verify unmounting flag is set
			assert.True(t, lm.IsUnmounting(), "should be marked as unmounting")
		})
	}
}

// TestLifecycleManager_ExecuteUnmounted_OnlyOnce tests that unmount only happens once.
func TestLifecycleManager_ExecuteUnmounted_OnlyOnce(t *testing.T) {
	tests := []struct {
		name          string
		callCount     int
		expectedCount int
		description   string
	}{
		{
			name:          "called once",
			callCount:     1,
			expectedCount: 1,
			description:   "should execute hooks once",
		},
		{
			name:          "called multiple times",
			callCount:     3,
			expectedCount: 1,
			description:   "should execute hooks only once even when called multiple times",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Register hook
			executionCount := 0
			lm.registerHook("unmounted", lifecycleHook{
				id:       "hook-1",
				callback: func() { executionCount++ },
				order:    0,
			})

			// Call executeUnmounted multiple times
			for i := 0; i < tt.callCount; i++ {
				lm.executeUnmounted()
			}

			// Verify hook executed only once
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteUnmounted_Order tests execution order.
func TestLifecycleManager_ExecuteUnmounted_Order(t *testing.T) {
	tests := []struct {
		name     string
		numHooks int
	}{
		{
			name:     "three hooks execute in order",
			numHooks: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Track execution order
			var executionOrder []int

			// Register hooks in order
			for i := 0; i < tt.numHooks; i++ {
				hookNum := i
				lm.registerHook("unmounted", lifecycleHook{
					id:       fmt.Sprintf("hook-%d", i),
					callback: func() { executionOrder = append(executionOrder, hookNum) },
					order:    i,
				})
			}

			// Execute unmounted hooks
			lm.executeUnmounted()

			// Verify execution order
			assert.Equal(t, []int{0, 1, 2}, executionOrder, "hooks should execute in registration order")
		})
	}
}

// TestLifecycleManager_ExecuteUnmounted_PanicRecovery tests panic recovery.
func TestLifecycleManager_ExecuteUnmounted_PanicRecovery(t *testing.T) {
	tests := []struct {
		name          string
		panicInFirst  bool
		expectedCount int
		description   string
	}{
		{
			name:          "first hook panics, second still executes",
			panicInFirst:  true,
			expectedCount: 1,
			description:   "second hook should execute even if first panics",
		},
		{
			name:          "no panics, both execute",
			panicInFirst:  false,
			expectedCount: 2,
			description:   "both hooks should execute normally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Track execution count
			executionCount := 0

			// First hook - may panic
			lm.registerHook("unmounted", lifecycleHook{
				id: "hook-1",
				callback: func() {
					if tt.panicInFirst {
						panic("test panic")
					}
					executionCount++
				},
				order: 0,
			})

			// Second hook - always executes
			lm.registerHook("unmounted", lifecycleHook{
				id:       "hook-2",
				callback: func() { executionCount++ },
				order:    1,
			})

			// Execute unmounted hooks (should not panic)
			assert.NotPanics(t, func() {
				lm.executeUnmounted()
			}, "executeUnmounted should not panic")

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteCleanups tests cleanup function execution.
func TestLifecycleManager_ExecuteCleanups(t *testing.T) {
	tests := []struct {
		name          string
		numCleanups   int
		expectedCount int
		description   string
	}{
		{
			name:          "no cleanups registered",
			numCleanups:   0,
			expectedCount: 0,
			description:   "should handle no cleanups gracefully",
		},
		{
			name:          "single cleanup",
			numCleanups:   1,
			expectedCount: 1,
			description:   "should execute single cleanup",
		},
		{
			name:          "multiple cleanups",
			numCleanups:   3,
			expectedCount: 3,
			description:   "should execute all cleanups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Register cleanups
			executionCount := 0
			for i := 0; i < tt.numCleanups; i++ {
				lm.cleanups = append(lm.cleanups, func() {
					executionCount++
				})
			}

			// Execute cleanups
			lm.executeCleanups()

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_ExecuteCleanups_ReverseOrder tests LIFO execution.
func TestLifecycleManager_ExecuteCleanups_ReverseOrder(t *testing.T) {
	tests := []struct {
		name        string
		numCleanups int
	}{
		{
			name:        "three cleanups execute in reverse order",
			numCleanups: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Track execution order
			var executionOrder []int

			// Register cleanups in order
			for i := 0; i < tt.numCleanups; i++ {
				cleanupNum := i
				lm.cleanups = append(lm.cleanups, func() {
					executionOrder = append(executionOrder, cleanupNum)
				})
			}

			// Execute cleanups
			lm.executeCleanups()

			// Verify reverse order (LIFO)
			assert.Equal(t, []int{2, 1, 0}, executionOrder, "cleanups should execute in reverse order (LIFO)")
		})
	}
}

// TestLifecycleManager_ExecuteCleanups_PanicRecovery tests panic recovery in cleanups.
func TestLifecycleManager_ExecuteCleanups_PanicRecovery(t *testing.T) {
	tests := []struct {
		name          string
		panicInFirst  bool
		expectedCount int
		description   string
	}{
		{
			name:          "first cleanup panics, second still executes",
			panicInFirst:  true,
			expectedCount: 1,
			description:   "second cleanup should execute even if first panics",
		},
		{
			name:          "no panics, both execute",
			panicInFirst:  false,
			expectedCount: 2,
			description:   "both cleanups should execute normally",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Track execution count
			executionCount := 0

			// Register cleanups (reverse order due to LIFO)
			// Second cleanup (executes first due to LIFO)
			lm.cleanups = append(lm.cleanups, func() { executionCount++ })

			// First cleanup (executes second due to LIFO) - may panic
			lm.cleanups = append(lm.cleanups, func() {
				if tt.panicInFirst {
					panic("test panic")
				}
				executionCount++
			})

			// Execute cleanups (should not panic)
			assert.NotPanics(t, func() {
				lm.executeCleanups()
			}, "executeCleanups should not panic")

			// Verify execution count
			assert.Equal(t, tt.expectedCount, executionCount, tt.description)
		})
	}
}

// TestLifecycleManager_InfiniteLoopDetection tests that infinite update loops are detected.
func TestLifecycleManager_InfiniteLoopDetection(t *testing.T) {
	tests := []struct {
		name            string
		updateCount     int
		expectError     bool
		expectExecution bool
	}{
		{
			name:            "below max depth allows execution",
			updateCount:     50,
			expectError:     false,
			expectExecution: true,
		},
		{
			name:            "at max depth allows execution",
			updateCount:     100,
			expectError:     false,
			expectExecution: true,
		},
		{
			name:            "exceeds max depth prevents execution",
			updateCount:     101,
			expectError:     true,
			expectExecution: false,
		},
		{
			name:            "far exceeds max depth prevents execution",
			updateCount:     200,
			expectError:     true,
			expectExecution: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)

			// Mark as mounted
			lm.setMounted(true)

			// Set update count to test value
			lm.updateCount = tt.updateCount

			// Register a hook
			executed := false
			lm.registerHook("updated", lifecycleHook{
				id:       "test-hook",
				callback: func() { executed = true },
				order:    0,
			})

			// Execute updated hooks
			lm.executeUpdated()

			// Verify execution matches expectation
			assert.Equal(t, tt.expectExecution, executed, "hook execution should match expectation")

			// Verify update count behavior
			if tt.expectError {
				// Should not increment when error occurs
				assert.Equal(t, tt.updateCount, lm.updateCount, "update count should not increment when max depth exceeded")
			}
		})
	}
}

// TestLifecycleManager_MaxDepthEnforced tests that the max update depth is strictly enforced.
func TestLifecycleManager_MaxDepthEnforced(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "max depth stops infinite loop",
			description: "update count exactly at 101 should prevent execution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Register hook that would create infinite loop
			hookCallCount := 0
			lm.registerHook("updated", lifecycleHook{
				id: "infinite-loop-hook",
				callback: func() {
					hookCallCount++
				},
				order: 0,
			})

			// Simulate reaching max depth
			lm.updateCount = 101

			// Try to execute - should be prevented
			lm.executeUpdated()

			// Hook should not execute
			assert.Equal(t, 0, hookCallCount, "hook should not execute when max depth exceeded")
		})
	}
}

// TestLifecycleManager_ErrorLogged tests that max depth error is properly logged.
func TestLifecycleManager_ErrorLogged(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "error logged when max depth exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Set update count above max
			lm.updateCount = 150

			// Execute updated - should log error via observability
			// Note: In production, this would report to observability system
			// In tests, GetErrorReporter() returns nil, so no actual reporting
			lm.executeUpdated()

			// Verify update count unchanged (error prevented execution)
			assert.Equal(t, 150, lm.updateCount, "update count should remain unchanged when max exceeded")
		})
	}
}

// TestLifecycleManager_ExecutionStopped tests that execution stops when max depth is exceeded.
func TestLifecycleManager_ExecutionStopped(t *testing.T) {
	tests := []struct {
		name             string
		hooksToRegister  int
		initialCount     int
		expectedExecuted int
	}{
		{
			name:             "all hooks skipped when max depth exceeded",
			hooksToRegister:  5,
			initialCount:     101,
			expectedExecuted: 0,
		},
		{
			name:             "no hooks registered still handles gracefully",
			hooksToRegister:  0,
			initialCount:     150,
			expectedExecuted: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			// Register multiple hooks
			executedCount := 0
			for i := 0; i < tt.hooksToRegister; i++ {
				lm.registerHook("updated", lifecycleHook{
					id: fmt.Sprintf("hook-%d", i),
					callback: func() {
						executedCount++
					},
					order: i,
				})
			}

			// Set update count above max
			lm.updateCount = tt.initialCount

			// Execute updated
			lm.executeUpdated()

			// Verify no hooks executed
			assert.Equal(t, tt.expectedExecuted, executedCount, "no hooks should execute when max depth exceeded")
		})
	}
}

// TestLifecycleManager_ComponentRecovers tests that component continues working after max depth error.
func TestLifecycleManager_ComponentRecovers(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "component recovers after max depth error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component and lifecycle manager
			c := newComponentImpl("TestComponent")
			lm := newLifecycleManager(c)
			lm.setMounted(true)

			executionLog := []string{}

			// Register hook
			lm.registerHook("updated", lifecycleHook{
				id: "test-hook",
				callback: func() {
					executionLog = append(executionLog, "executed")
				},
				order: 0,
			})

			// First, trigger max depth error
			lm.updateCount = 101
			lm.executeUpdated()
			assert.Empty(t, executionLog, "hook should not execute at max depth")

			// Reset update count (simulating recovery)
			lm.updateCount = 0

			// Now it should work again
			lm.executeUpdated()
			assert.Equal(t, 1, len(executionLog), "hook should execute after reset")
			assert.Equal(t, "executed", executionLog[0], "hook should have executed")
		})
	}
}

func TestLifecycleManager_ResetUpdateCount(t *testing.T) {
	tests := []struct {
		name          string
		initialCount  int
		expectedAfter int
	}{
		{
			name:          "reset from zero",
			initialCount:  0,
			expectedAfter: 0,
		},
		{
			name:          "reset from positive",
			initialCount:  50,
			expectedAfter: 0,
		},
		{
			name:          "reset from max depth",
			initialCount:  maxUpdateDepth,
			expectedAfter: 0,
		},
		{
			name:          "reset from over max depth",
			initialCount:  maxUpdateDepth + 10,
			expectedAfter: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			component := &componentImpl{
				name: "TestComponent",
				id:   "test-1",
			}
			lm := newLifecycleManager(component)

			// Set initial update count
			lm.updateCount = tt.initialCount

			// Reset
			lm.resetUpdateCount()

			// Verify count is reset
			assert.Equal(t, tt.expectedAfter, lm.updateCount)
		})
	}
}

// TestLifecycleManager_RegisterWatcher tests that watchers are registered correctly
func TestLifecycleManager_RegisterWatcher(t *testing.T) {
	tests := []struct {
		name          string
		numWatchers   int
		expectedCount int
	}{
		{
			name:          "register single watcher",
			numWatchers:   1,
			expectedCount: 1,
		},
		{
			name:          "register multiple watchers",
			numWatchers:   5,
			expectedCount: 5,
		},
		{
			name:          "register no watchers",
			numWatchers:   0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			component := &componentImpl{
				name: "TestComponent",
				id:   "test-1",
			}
			lm := newLifecycleManager(component)

			// Register watchers
			for i := 0; i < tt.numWatchers; i++ {
				cleanup := func() {}
				lm.registerWatcher(cleanup)
			}

			// Verify watcher count
			assert.Equal(t, tt.expectedCount, len(lm.watchers))
		})
	}
}

// TestLifecycleManager_CleanupWatchers tests that all watchers are cleaned up
func TestLifecycleManager_CleanupWatchers(t *testing.T) {
	tests := []struct {
		name          string
		numWatchers   int
		expectedCalls int
	}{
		{
			name:          "cleanup single watcher",
			numWatchers:   1,
			expectedCalls: 1,
		},
		{
			name:          "cleanup multiple watchers",
			numWatchers:   5,
			expectedCalls: 5,
		},
		{
			name:          "cleanup no watchers",
			numWatchers:   0,
			expectedCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			component := &componentImpl{
				name: "TestComponent",
				id:   "test-1",
			}
			lm := newLifecycleManager(component)

			// Track cleanup calls
			cleanupCalls := 0
			var mu sync.Mutex

			// Register watchers
			for i := 0; i < tt.numWatchers; i++ {
				cleanup := func() {
					mu.Lock()
					cleanupCalls++
					mu.Unlock()
				}
				lm.registerWatcher(cleanup)
			}

			// Execute cleanup
			lm.cleanupWatchers()

			// Verify all cleanups were called
			mu.Lock()
			assert.Equal(t, tt.expectedCalls, cleanupCalls)
			mu.Unlock()
		})
	}
}

// TestLifecycleManager_CleanupWatchers_PanicRecovery tests panic recovery in watcher cleanup
func TestLifecycleManager_CleanupWatchers_PanicRecovery(t *testing.T) {
	tests := []struct {
		name          string
		panicIndex    int
		totalWatchers int
		expectedCalls int
	}{
		{
			name:          "panic in first watcher",
			panicIndex:    0,
			totalWatchers: 3,
			expectedCalls: 3, // All should be attempted
		},
		{
			name:          "panic in middle watcher",
			panicIndex:    1,
			totalWatchers: 3,
			expectedCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create lifecycle manager
			component := &componentImpl{
				name: "TestComponent",
				id:   "test-1",
			}
			lm := newLifecycleManager(component)

			// Track cleanup calls
			cleanupCalls := 0
			var mu sync.Mutex

			// Register watchers
			for i := 0; i < tt.totalWatchers; i++ {
				idx := i
				cleanup := func() {
					mu.Lock()
					cleanupCalls++
					mu.Unlock()
					if idx == tt.panicIndex {
						panic("cleanup panic")
					}
				}
				lm.registerWatcher(cleanup)
			}

			// Execute cleanup (should not panic)
			assert.NotPanics(t, func() {
				lm.cleanupWatchers()
			})

			// Verify all cleanups were attempted
			mu.Lock()
			assert.Equal(t, tt.expectedCalls, cleanupCalls)
			mu.Unlock()
		})
	}
}

// TestContext_Watch_AutoCleanup tests that Context.Watch registers cleanup with lifecycle
func TestContext_Watch_AutoCleanup(t *testing.T) {
	// Create component with lifecycle
	component := &componentImpl{
		name:      "TestComponent",
		id:        "test-1",
		state:     make(map[string]interface{}),
		lifecycle: nil,
	}
	ctx := &Context{component: component}

	// Create a ref (interface{} type to match Context.Watch signature)
	ref := NewRef[interface{}](0)

	// Track callback calls
	callbackCalls := 0
	var mu sync.Mutex

	// Watch the ref through context
	cleanup := ctx.Watch(ref, func(newVal, oldVal interface{}) {
		mu.Lock()
		callbackCalls++
		mu.Unlock()
	})

	// Verify cleanup function returned
	assert.NotNil(t, cleanup)

	// Verify lifecycle manager was created
	assert.NotNil(t, component.lifecycle)

	// Verify watcher was registered
	assert.Equal(t, 1, len(component.lifecycle.watchers))

	// Change ref value - callback should be called
	ref.Set(1)
	mu.Lock()
	assert.Equal(t, 1, callbackCalls)
	mu.Unlock()

	// Execute unmount - should cleanup watchers
	component.lifecycle.executeUnmounted()

	// Change ref value again - callback should NOT be called (watcher cleaned up)
	ref.Set(2)
	mu.Lock()
	assert.Equal(t, 1, callbackCalls) // Still 1, not incremented
	mu.Unlock()
}

// TestContext_Watch_MultipleWatchers tests multiple watchers auto-cleanup
func TestContext_Watch_MultipleWatchers(t *testing.T) {
	// Create component with lifecycle
	component := &componentImpl{
		name:      "TestComponent",
		id:        "test-1",
		state:     make(map[string]interface{}),
		lifecycle: nil,
	}
	ctx := &Context{component: component}

	// Create refs (interface{} type to match Context.Watch signature)
	ref1 := NewRef[interface{}](0)
	ref2 := NewRef[interface{}]("hello")

	// Track callback calls
	calls1 := 0
	calls2 := 0
	var mu sync.Mutex

	// Watch both refs
	cleanup1 := ctx.Watch(ref1, func(newVal, oldVal interface{}) {
		mu.Lock()
		calls1++
		mu.Unlock()
	})
	cleanup2 := ctx.Watch(ref2, func(newVal, oldVal interface{}) {
		mu.Lock()
		calls2++
		mu.Unlock()
	})

	// Verify cleanups returned
	assert.NotNil(t, cleanup1)
	assert.NotNil(t, cleanup2)

	// Verify both watchers registered
	assert.Equal(t, 2, len(component.lifecycle.watchers))

	// Change values - callbacks should be called
	ref1.Set(1)
	ref2.Set("world")
	mu.Lock()
	assert.Equal(t, 1, calls1)
	assert.Equal(t, 1, calls2)
	mu.Unlock()

	// Execute unmount - should cleanup all watchers
	component.lifecycle.executeUnmounted()

	// Change values again - callbacks should NOT be called
	ref1.Set(2)
	ref2.Set("!")
	mu.Lock()
	assert.Equal(t, 1, calls1) // Still 1
	assert.Equal(t, 1, calls2) // Still 1
	mu.Unlock()
}

// TestLifecycleManager_WatcherCleanupOrder tests that watchers cleanup before manual cleanups
func TestLifecycleManager_WatcherCleanupOrder(t *testing.T) {
	// Create lifecycle manager
	component := &componentImpl{
		name: "TestComponent",
		id:   "test-1",
	}
	lm := newLifecycleManager(component)

	// Track execution order
	var order []string
	var mu sync.Mutex

	// Register watcher cleanup
	lm.registerWatcher(func() {
		mu.Lock()
		order = append(order, "watcher")
		mu.Unlock()
	})

	// Register manual cleanup
	lm.cleanups = append(lm.cleanups, func() {
		mu.Lock()
		order = append(order, "manual")
		mu.Unlock()
	})

	// Execute unmount
	lm.executeUnmounted()

	// Verify order: watchers before manual cleanups
	mu.Lock()
	assert.Equal(t, []string{"watcher", "manual"}, order)
	mu.Unlock()
}
