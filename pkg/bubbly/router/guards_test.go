package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouter_BeforeEach verifies global before guard registration
func TestRouter_BeforeEach(t *testing.T) {
	router := NewRouter()

	guardCalled := false
	guard := func(to, from *Route, next NextFunc) {
		guardCalled = true
		next(nil)
	}

	// Register guard
	router.BeforeEach(guard)

	// Verify guard was registered
	assert.Len(t, router.beforeHooks, 1, "Should have one before guard")

	// Setup route and navigate
	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	msg := cmd()

	// Verify guard was called
	assert.True(t, guardCalled, "Guard should have been called")
	assert.IsType(t, RouteChangedMsg{}, msg, "Navigation should succeed")
}

// TestRouter_AfterEach verifies global after hook registration
func TestRouter_AfterEach(t *testing.T) {
	router := NewRouter()

	hookCalled := false
	hook := func(to, from *Route) {
		hookCalled = true
	}

	// Register hook
	router.AfterEach(hook)

	// Verify hook was registered
	assert.Len(t, router.afterHooks, 1, "Should have one after hook")

	// Setup route and navigate
	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	msg := cmd()

	// Verify hook was called
	assert.True(t, hookCalled, "Hook should have been called")
	assert.IsType(t, RouteChangedMsg{}, msg, "Navigation should succeed")
}

// TestRouter_Guards_ExecutionOrder verifies guards execute in correct order
func TestRouter_Guards_ExecutionOrder(t *testing.T) {
	router := NewRouter()

	var executionOrder []string

	// Register multiple guards
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		executionOrder = append(executionOrder, "guard1")
		next(nil)
	})

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		executionOrder = append(executionOrder, "guard2")
		next(nil)
	})

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		executionOrder = append(executionOrder, "guard3")
		next(nil)
	})

	// Setup route and navigate
	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	cmd()

	// Verify execution order
	require.Len(t, executionOrder, 3, "All guards should execute")
	assert.Equal(t, "guard1", executionOrder[0])
	assert.Equal(t, "guard2", executionOrder[1])
	assert.Equal(t, "guard3", executionOrder[2])
}

// TestRouter_Guards_AllowNavigation verifies next(nil) allows navigation
func TestRouter_Guards_AllowNavigation(t *testing.T) {
	router := NewRouter()

	guardCalled := false
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		guardCalled = true
		next(nil) // Allow navigation
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	msg := cmd()

	assert.True(t, guardCalled, "Guard should be called")
	assert.IsType(t, RouteChangedMsg{}, msg, "Navigation should succeed")

	// Verify route was changed
	currentRoute := router.CurrentRoute()
	require.NotNil(t, currentRoute)
	assert.Equal(t, "/test", currentRoute.Path)
}

// TestRouter_Guards_CancelNavigation verifies next() with empty target cancels
func TestRouter_Guards_CancelNavigation(t *testing.T) {
	router := NewRouter()

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		// Cancel navigation with empty target
		next(&NavigationTarget{})
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	msg := cmd()

	// Should return error message
	assert.IsType(t, NavigationErrorMsg{}, msg, "Navigation should be cancelled")

	// Verify route was NOT changed
	assert.Nil(t, router.CurrentRoute(), "Route should not change")
}

// TestRouter_Guards_RedirectNavigation verifies next() with path redirects
func TestRouter_Guards_RedirectNavigation(t *testing.T) {
	router := NewRouter()

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		if to.Path == "/protected" {
			// Redirect to login
			next(&NavigationTarget{Path: "/login"})
		} else {
			next(nil)
		}
	})

	router.registry.Register("/protected", "protected", nil)
	router.registry.Register("/login", "login", nil)

	cmd := router.Push(&NavigationTarget{Path: "/protected"})
	msg := cmd()

	// Should succeed with redirect
	assert.IsType(t, RouteChangedMsg{}, msg, "Navigation should succeed")

	// Verify we ended up at /login, not /protected
	currentRoute := router.CurrentRoute()
	require.NotNil(t, currentRoute)
	assert.Equal(t, "/login", currentRoute.Path, "Should redirect to login")
}

// TestRouter_Guards_ToFromRoutes verifies guards receive correct to/from routes
func TestRouter_Guards_ToFromRoutes(t *testing.T) {
	router := NewRouter()

	var receivedTo, receivedFrom *Route

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		receivedTo = to
		receivedFrom = from
		next(nil)
	})

	router.registry.Register("/home", "home", nil)
	router.registry.Register("/about", "about", nil)

	// First navigation (from nil)
	cmd := router.Push(&NavigationTarget{Path: "/home"})
	cmd()

	assert.NotNil(t, receivedTo, "Should receive 'to' route")
	assert.Equal(t, "/home", receivedTo.Path)
	assert.Nil(t, receivedFrom, "First navigation should have nil 'from'")

	// Second navigation (from /home to /about)
	receivedTo = nil
	receivedFrom = nil

	cmd = router.Push(&NavigationTarget{Path: "/about"})
	cmd()

	assert.NotNil(t, receivedTo, "Should receive 'to' route")
	assert.Equal(t, "/about", receivedTo.Path)
	assert.NotNil(t, receivedFrom, "Should receive 'from' route")
	assert.Equal(t, "/home", receivedFrom.Path)
}

// TestRouter_Guards_StopOnFirstCancel verifies remaining guards don't execute after cancel
func TestRouter_Guards_StopOnFirstCancel(t *testing.T) {
	router := NewRouter()

	guard1Called := false
	guard2Called := false
	guard3Called := false

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		guard1Called = true
		next(nil) // Allow
	})

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		guard2Called = true
		next(&NavigationTarget{}) // Cancel
	})

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		guard3Called = true
		next(nil)
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	cmd()

	assert.True(t, guard1Called, "Guard 1 should execute")
	assert.True(t, guard2Called, "Guard 2 should execute")
	assert.False(t, guard3Called, "Guard 3 should NOT execute after cancel")
}

// TestRouter_AfterHooks_ExecuteAfterNavigation verifies after hooks execute after navigation
func TestRouter_AfterHooks_ExecuteAfterNavigation(t *testing.T) {
	router := NewRouter()

	var receivedTo, receivedFrom *Route
	hookCalled := false

	router.AfterEach(func(to, from *Route) {
		hookCalled = true
		receivedTo = to
		receivedFrom = from
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	msg := cmd()

	assert.IsType(t, RouteChangedMsg{}, msg, "Navigation should succeed")
	assert.True(t, hookCalled, "After hook should be called")
	assert.NotNil(t, receivedTo, "Should receive 'to' route")
	assert.Equal(t, "/test", receivedTo.Path)
	assert.Nil(t, receivedFrom, "First navigation has nil 'from'")
}

// TestRouter_AfterHooks_MultipleHooks verifies multiple after hooks execute
func TestRouter_AfterHooks_MultipleHooks(t *testing.T) {
	router := NewRouter()

	var executionOrder []string

	router.AfterEach(func(to, from *Route) {
		executionOrder = append(executionOrder, "hook1")
	})

	router.AfterEach(func(to, from *Route) {
		executionOrder = append(executionOrder, "hook2")
	})

	router.AfterEach(func(to, from *Route) {
		executionOrder = append(executionOrder, "hook3")
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	cmd()

	require.Len(t, executionOrder, 3, "All hooks should execute")
	assert.Equal(t, "hook1", executionOrder[0])
	assert.Equal(t, "hook2", executionOrder[1])
	assert.Equal(t, "hook3", executionOrder[2])
}

// TestRouter_AfterHooks_NotCalledOnCancel verifies after hooks don't execute on cancel
func TestRouter_AfterHooks_NotCalledOnCancel(t *testing.T) {
	router := NewRouter()

	hookCalled := false

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		next(&NavigationTarget{}) // Cancel
	})

	router.AfterEach(func(to, from *Route) {
		hookCalled = true
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Push(&NavigationTarget{Path: "/test"})
	msg := cmd()

	assert.IsType(t, NavigationErrorMsg{}, msg, "Navigation should be cancelled")
	assert.False(t, hookCalled, "After hook should NOT be called on cancel")
}

// TestRouter_Guards_WorkWithReplace verifies guards work with Replace() too
func TestRouter_Guards_WorkWithReplace(t *testing.T) {
	router := NewRouter()

	guardCalled := false
	hookCalled := false

	router.BeforeEach(func(to, from *Route, next NextFunc) {
		guardCalled = true
		next(nil)
	})

	router.AfterEach(func(to, from *Route) {
		hookCalled = true
	})

	router.registry.Register("/test", "test", nil)
	cmd := router.Replace(&NavigationTarget{Path: "/test"})
	msg := cmd()

	assert.IsType(t, RouteChangedMsg{}, msg, "Navigation should succeed")
	assert.True(t, guardCalled, "Guard should be called for Replace")
	assert.True(t, hookCalled, "Hook should be called for Replace")
}
