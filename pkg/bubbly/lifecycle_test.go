package bubbly

import (
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
