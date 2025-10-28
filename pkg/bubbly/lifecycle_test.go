package bubbly

import (
	"fmt"
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
