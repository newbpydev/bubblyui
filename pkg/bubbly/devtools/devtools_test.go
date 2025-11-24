package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestEnable_CreatesInstance tests that Enable creates a DevTools instance
func TestEnable_CreatesInstance(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	dt := Enable()

	assert.NotNil(t, dt, "Enable should return non-nil DevTools instance")
	assert.True(t, dt.IsEnabled(), "DevTools should be enabled after Enable()")
}

// TestEnable_ReturnsSameInstance tests singleton pattern
func TestEnable_ReturnsSameInstance(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	dt1 := Enable()
	dt2 := Enable()

	assert.Same(t, dt1, dt2, "Multiple Enable() calls should return same instance")
}

// TestDisable_DisablesDevTools tests that Disable disables the dev tools
func TestDisable_DisablesDevTools(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	Enable()
	Disable()

	assert.False(t, IsEnabled(), "DevTools should be disabled after Disable()")
}

// TestToggle_TogglesEnabledState tests Toggle functionality
func TestToggle_TogglesEnabledState(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	// Initially disabled
	assert.False(t, IsEnabled(), "DevTools should start disabled")

	// Toggle to enabled
	Toggle()
	assert.True(t, IsEnabled(), "DevTools should be enabled after first Toggle()")

	// Toggle back to disabled
	Toggle()
	assert.False(t, IsEnabled(), "DevTools should be disabled after second Toggle()")
}

// TestIsEnabled_ReflectsState tests IsEnabled returns correct state
func TestIsEnabled_ReflectsState(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	assert.False(t, IsEnabled(), "DevTools should start disabled")

	Enable()
	assert.True(t, IsEnabled(), "DevTools should be enabled after Enable()")

	Disable()
	assert.False(t, IsEnabled(), "DevTools should be disabled after Disable()")
}

// TestSetVisible_SetsVisibility tests SetVisible method
func TestSetVisible_SetsVisibility(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	dt := Enable()

	// Initially not visible
	assert.False(t, dt.IsVisible(), "DevTools should start not visible")

	// Set visible
	dt.SetVisible(true)
	assert.True(t, dt.IsVisible(), "DevTools should be visible after SetVisible(true)")

	// Set not visible
	dt.SetVisible(false)
	assert.False(t, dt.IsVisible(), "DevTools should not be visible after SetVisible(false)")
}

// TestToggleVisibility_TogglesVisibility tests ToggleVisibility method
func TestToggleVisibility_TogglesVisibility(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	dt := Enable()

	// Initially not visible
	assert.False(t, dt.IsVisible(), "DevTools should start not visible")

	// Toggle to visible
	dt.ToggleVisibility()
	assert.True(t, dt.IsVisible(), "DevTools should be visible after first ToggleVisibility()")

	// Toggle back to not visible
	dt.ToggleVisibility()
	assert.False(t, dt.IsVisible(), "DevTools should not be visible after second ToggleVisibility()")
}

// TestConcurrentAccess_ThreadSafe tests thread-safe concurrent access
func TestConcurrentAccess_ThreadSafe(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	dt := Enable()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent Enable calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Enable()
		}()
	}

	// Concurrent IsEnabled calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			IsEnabled()
		}()
	}

	// Concurrent SetVisible calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			dt.SetVisible(i%2 == 0)
		}(i)
	}

	// Concurrent IsVisible calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dt.IsVisible()
		}()
	}

	wg.Wait()

	// Should not panic or deadlock
	assert.NotNil(t, dt, "DevTools should still be valid after concurrent access")
}

// TestToggle_ConcurrentAccess tests concurrent Toggle calls
func TestToggle_ConcurrentAccess(t *testing.T) {
	// Reset singleton for test isolation
	resetSingleton()

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Toggle()
		}()
	}

	wg.Wait()

	// Should not panic or deadlock
	// Final state is non-deterministic but that's OK
	_ = IsEnabled()
}

// resetSingleton resets the global singleton for test isolation
// This is a test helper function
func resetSingleton() {
	globalDevToolsMu.Lock()
	defer globalDevToolsMu.Unlock()
	globalDevTools = nil
	// CRITICAL FIX: Must reset sync.Once so hook gets registered in next Enable()
	globalDevToolsOnce = sync.Once{}

	// Also unregister the hook from bubbly package
	_ = bubbly.UnregisterHook()
}

// TestDevTools_GetStore tests GetStore method
func TestDevTools_GetStore(t *testing.T) {
	resetSingleton()

	dt := Enable()

	store := dt.GetStore()

	assert.NotNil(t, store, "GetStore should return non-nil Store")
}

// TestDevTools_SetMCPServer tests SetMCPServer method
func TestDevTools_SetMCPServer(t *testing.T) {
	resetSingleton()

	dt := Enable()

	// Initially no MCP server
	assert.Nil(t, dt.GetMCPServer(), "MCP server should be nil initially")
	assert.False(t, dt.MCPEnabled(), "MCPEnabled should be false initially")

	// Set MCP server
	mockServer := struct{}{}
	dt.SetMCPServer(mockServer)

	assert.NotNil(t, dt.GetMCPServer(), "GetMCPServer should return the server")
	assert.True(t, dt.MCPEnabled(), "MCPEnabled should be true after setting server")
}

// TestDevTools_GetMCPServer tests GetMCPServer method
func TestDevTools_GetMCPServer(t *testing.T) {
	resetSingleton()

	dt := Enable()

	// Initially nil
	assert.Nil(t, dt.GetMCPServer(), "GetMCPServer should return nil initially")

	// Set a mock server
	mockServer := "mock-server"
	dt.SetMCPServer(mockServer)

	// Should return the server
	retrieved := dt.GetMCPServer()
	assert.Equal(t, mockServer, retrieved, "GetMCPServer should return the set server")
}

// TestDevTools_MCPEnabled tests MCPEnabled method
func TestDevTools_MCPEnabled(t *testing.T) {
	tests := []struct {
		name       string
		setServer  interface{}
		wantResult bool
	}{
		{
			name:       "no server set",
			setServer:  nil,
			wantResult: false,
		},
		{
			name:       "server set",
			setServer:  "mock-server",
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSingleton()
			dt := Enable()

			if tt.setServer != nil {
				dt.SetMCPServer(tt.setServer)
			}

			assert.Equal(t, tt.wantResult, dt.MCPEnabled())
		})
	}
}

// TestRenderView_WhenDisabled tests RenderView when DevTools is disabled
func TestRenderView_WhenDisabled(t *testing.T) {
	resetSingleton()

	appView := "Test Application View"

	// DevTools not enabled - should return app view unchanged
	result := RenderView(appView)

	assert.Equal(t, appView, result, "RenderView should return app view unchanged when disabled")
}

// TestRenderView_WhenNotVisible tests RenderView when DevTools is enabled but not visible
func TestRenderView_WhenNotVisible(t *testing.T) {
	resetSingleton()

	dt := Enable()
	dt.SetVisible(false)

	appView := "Test Application View"

	result := RenderView(appView)

	assert.Equal(t, appView, result, "RenderView should return app view unchanged when not visible")
}

// TestHandleUpdate_WhenDisabled tests HandleUpdate when DevTools is disabled
func TestHandleUpdate_WhenDisabled(t *testing.T) {
	resetSingleton()

	// DevTools not enabled
	cmd := HandleUpdate(nil)

	assert.Nil(t, cmd, "HandleUpdate should return nil when disabled")
}

// TestHandleUpdate_WhenNotVisible tests HandleUpdate when DevTools is enabled but not visible
func TestHandleUpdate_WhenNotVisible(t *testing.T) {
	resetSingleton()

	dt := Enable()
	dt.SetVisible(false)

	cmd := HandleUpdate(nil)

	assert.Nil(t, cmd, "HandleUpdate should return nil when not visible")
}

// TestFrameworkHookAdapter_OnComponentMount tests framework hook adapter
func TestFrameworkHookAdapter_OnComponentMount(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Mount a user component (not framework internal)
	adapter.OnComponentMount("user-comp-1", "MyComponent")

	comp := store.GetComponent("user-comp-1")
	assert.NotNil(t, comp, "User component should be added to store")
	assert.Equal(t, "MyComponent", comp.Name)
	assert.Equal(t, "mounted", comp.Status)
}

// TestFrameworkHookAdapter_OnComponentMount_FiltersFrameworkComponents tests filtering
func TestFrameworkHookAdapter_OnComponentMount_FiltersFrameworkComponents(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// These are framework internal components and should be filtered out
	frameworkComponents := []string{"Button", "Card", "Text", "Input", "Table", "List", "Modal"}

	for _, name := range frameworkComponents {
		adapter.OnComponentMount("comp-"+name, name)
		comp := store.GetComponent("comp-" + name)
		assert.Nil(t, comp, "Framework component %s should be filtered out", name)
	}
}

// TestFrameworkHookAdapter_OnComponentMount_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnComponentMount_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnComponentMount("comp-1", "TestComponent")
	})
}

// TestFrameworkHookAdapter_OnComponentUnmount tests unmount handling
func TestFrameworkHookAdapter_OnComponentUnmount(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// First mount a component
	adapter.OnComponentMount("comp-1", "TestComponent")

	// Then unmount it
	adapter.OnComponentUnmount("comp-1")

	// Component should be updated with unmounted status
	comp := store.GetComponent("comp-1")
	assert.NotNil(t, comp)
	assert.Equal(t, "unmounted", comp.Status)
}

// TestFrameworkHookAdapter_OnComponentUnmount_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnComponentUnmount_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnComponentUnmount("comp-1")
	})
}

// TestFrameworkHookAdapter_OnRefChange tests ref change handling
func TestFrameworkHookAdapter_OnRefChange(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Record a ref change
	adapter.OnRefChange("ref-1", "old-value", "new-value")

	// Verify state history was recorded
	history := store.stateHistory.GetHistory("ref-1")
	assert.Equal(t, 1, len(history), "State change should be recorded")
	assert.Equal(t, "old-value", history[0].OldValue)
	assert.Equal(t, "new-value", history[0].NewValue)
}

// TestFrameworkHookAdapter_OnRefChange_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnRefChange_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnRefChange("ref-1", "old", "new")
	})
}

// TestFrameworkHookAdapter_OnEvent tests event handling
func TestFrameworkHookAdapter_OnEvent(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Record an event
	adapter.OnEvent("comp-1", "click", map[string]interface{}{"x": 10, "y": 20})

	// Verify event was recorded
	events := store.events.GetRecent(1)
	assert.Equal(t, 1, len(events), "Event should be recorded")
	assert.Equal(t, "click", events[0].Name)
	assert.Equal(t, "comp-1", events[0].SourceID)
}

// TestFrameworkHookAdapter_OnEvent_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnEvent_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnEvent("comp-1", "click", nil)
	})
}

// TestFrameworkHookAdapter_OnRenderComplete tests render complete handling
func TestFrameworkHookAdapter_OnRenderComplete(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// First add a component
	store.AddComponent(&ComponentSnapshot{ID: "comp-1", Name: "TestComponent"})

	// Record render completion
	adapter.OnRenderComplete("comp-1", 5*time.Millisecond)

	// Verify performance was recorded
	perf := store.performance.GetComponent("comp-1")
	assert.NotNil(t, perf)
	assert.Equal(t, int64(1), perf.RenderCount)
}

// TestFrameworkHookAdapter_OnRenderComplete_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnRenderComplete_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnRenderComplete("comp-1", 5*time.Millisecond)
	})
}

// TestFrameworkHookAdapter_OnComputedChange tests computed change handling
func TestFrameworkHookAdapter_OnComputedChange(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Record a computed change (should delegate to OnRefChange)
	adapter.OnComputedChange("computed-1", "old", "new")

	// Verify state history was recorded
	history := store.stateHistory.GetHistory("computed-1")
	assert.Equal(t, 1, len(history), "Computed change should be recorded in state history")
}

// TestFrameworkHookAdapter_OnWatchCallback tests watch callback handling
func TestFrameworkHookAdapter_OnWatchCallback(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic (currently a no-op)
	assert.NotPanics(t, func() {
		adapter.OnWatchCallback("watcher-1", "new", "old")
	})
}

// TestFrameworkHookAdapter_OnEffectRun tests effect run handling
func TestFrameworkHookAdapter_OnEffectRun(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic (currently a no-op)
	assert.NotPanics(t, func() {
		adapter.OnEffectRun("effect-1")
	})
}

// TestFrameworkHookAdapter_OnChildAdded tests child added handling
func TestFrameworkHookAdapter_OnChildAdded(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Add parent component
	store.AddComponent(&ComponentSnapshot{ID: "parent-1", Name: "Parent"})

	// Add child relationship
	adapter.OnChildAdded("parent-1", "child-1")

	// Verify hierarchy was recorded
	children := store.GetComponentChildren("parent-1")
	assert.Contains(t, children, "child-1", "Child should be added to parent")
}

// TestFrameworkHookAdapter_OnChildAdded_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnChildAdded_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnChildAdded("parent-1", "child-1")
	})
}

// TestFrameworkHookAdapter_OnChildRemoved tests child removed handling
func TestFrameworkHookAdapter_OnChildRemoved(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Add parent and child
	store.AddComponent(&ComponentSnapshot{ID: "parent-1", Name: "Parent"})
	store.AddComponentChild("parent-1", "child-1")

	// Remove child
	adapter.OnChildRemoved("parent-1", "child-1")

	// Verify child was removed
	children := store.GetComponentChildren("parent-1")
	assert.NotContains(t, children, "child-1", "Child should be removed from parent")
}

// TestFrameworkHookAdapter_OnChildRemoved_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnChildRemoved_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnChildRemoved("parent-1", "child-1")
	})
}

// TestFrameworkHookAdapter_OnRefExposed tests ref exposed handling
func TestFrameworkHookAdapter_OnRefExposed(t *testing.T) {
	resetSingleton()

	dt := Enable()
	store := dt.GetStore()

	adapter := &frameworkHookAdapter{store: store}

	// Add a component
	store.AddComponent(&ComponentSnapshot{ID: "comp-1", Name: "TestComponent"})

	// Expose a ref
	adapter.OnRefExposed("comp-1", "ref-1", "counter")

	// The ref owner should be registered (we can verify by updating the ref)
	ownerID, updated := store.UpdateRefValue("ref-1", 42)
	assert.True(t, updated || ownerID == "comp-1", "Ref owner should be registered")
}

// TestFrameworkHookAdapter_OnRefExposed_NilStore tests nil store handling
func TestFrameworkHookAdapter_OnRefExposed_NilStore(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic with nil store
	assert.NotPanics(t, func() {
		adapter.OnRefExposed("comp-1", "ref-1", "counter")
	})
}

// TestIsFrameworkInternalComponent tests the framework component filter
func TestIsFrameworkInternalComponent(t *testing.T) {
	tests := []struct {
		name           string
		componentName  string
		wantIsInternal bool
	}{
		{"Button is internal", "Button", true},
		{"Card is internal", "Card", true},
		{"Text is internal", "Text", true},
		{"Input is internal", "Input", true},
		{"Table is internal", "Table", true},
		{"List is internal", "List", true},
		{"Modal is internal", "Modal", true},
		{"Tabs is internal", "Tabs", true},
		{"Form is internal", "Form", true},
		{"Counter is user component", "Counter", false},
		{"TodoList is user component", "TodoList", false},
		{"MyWidget is user component", "MyWidget", false},
		{"AppLayout is internal", "AppLayout", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFrameworkInternalComponent(tt.componentName)
			assert.Equal(t, tt.wantIsInternal, result)
		})
	}
}

// TestFrameworkHookAdapter_OnComponentUpdate tests component update handling
func TestFrameworkHookAdapter_OnComponentUpdate(t *testing.T) {
	adapter := &frameworkHookAdapter{store: nil}

	// Should not panic (currently a no-op that just tracks messages)
	assert.NotPanics(t, func() {
		adapter.OnComponentUpdate("comp-1", "some-message")
	})
}
