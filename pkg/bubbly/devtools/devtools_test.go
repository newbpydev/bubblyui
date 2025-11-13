package devtools

import (
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
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
	bubbly.UnregisterHook()
}
