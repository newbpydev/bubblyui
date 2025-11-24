package router

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ===========================================================================
// Tests for guards.go - getPathOrEmpty
// ===========================================================================

func TestGetPathOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		route    *Route
		expected string
	}{
		{
			name:     "nil route returns empty string",
			route:    nil,
			expected: "",
		},
		{
			name: "route with path returns path",
			route: &Route{
				Path: "/test/path",
			},
			expected: "/test/path",
		},
		{
			name: "route with empty path returns empty string",
			route: &Route{
				Path: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPathOrEmpty(tt.route)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ===========================================================================
// Tests for guards.go - executeGuardSafe (panic recovery)
// ===========================================================================

func TestExecuteGuardSafe_PanicRecovery(t *testing.T) {
	tests := []struct {
		name           string
		guard          NavigationGuard
		expectedAction guardAction
	}{
		{
			name: "guard that panics cancels navigation",
			guard: func(to, from *Route, next NextFunc) {
				panic("test panic")
			},
			expectedAction: guardCancel,
		},
		{
			name: "guard that calls next(nil) continues",
			guard: func(to, from *Route, next NextFunc) {
				next(nil)
			},
			expectedAction: guardContinue,
		},
		{
			name: "guard that cancels",
			guard: func(to, from *Route, next NextFunc) {
				next(&NavigationTarget{})
			},
			expectedAction: guardCancel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &guardResult{action: guardContinue}
			next := createNextCallback(result)
			ctx := guardContext{
				guardType: "test_guard",
				index:     0,
				routeName: "test-route",
				from:      &Route{Path: "/from"},
				to:        &Route{Path: "/to"},
			}

			to := &Route{Path: "/to"}
			from := &Route{Path: "/from"}

			// This should not panic - recovery is built in
			executeGuardSafe(tt.guard, to, from, next, result, ctx)

			assert.Equal(t, tt.expectedAction, result.action)
		})
	}
}

func TestExecuteGuardSafe_WithNegativeIndex(t *testing.T) {
	// Test with negative index (used for route-specific guards)
	result := &guardResult{action: guardContinue}
	next := createNextCallback(result)
	ctx := guardContext{
		guardType: "route_before_enter",
		index:     -1, // Negative index for route-specific guards
		routeName: "test-route",
		from:      nil,
		to:        &Route{Path: "/to"},
	}

	guard := func(to, from *Route, next NextFunc) {
		panic("test panic with negative index")
	}

	to := &Route{Path: "/to"}

	executeGuardSafe(guard, to, nil, next, result, ctx)
	assert.Equal(t, guardCancel, result.action)
}

// ===========================================================================
// Tests for guards.go - executeBeforeGuards with route-specific beforeEnter
// ===========================================================================

func TestExecuteBeforeGuards_WithBeforeEnterGuard(t *testing.T) {
	router := NewRouter()

	beforeEnterCalled := false
	beforeEnterGuard := NavigationGuard(func(to, from *Route, next NextFunc) {
		beforeEnterCalled = true
		next(nil)
	})

	// Create route with beforeEnter in Meta
	to := &Route{
		Path: "/test",
		Name: "test",
		Meta: map[string]interface{}{
			"beforeEnter": beforeEnterGuard,
		},
	}

	result := router.executeBeforeGuards(to, nil)

	assert.True(t, beforeEnterCalled, "beforeEnter guard should be called")
	assert.Equal(t, guardContinue, result.action)
}

func TestExecuteBeforeGuards_BeforeEnterCancels(t *testing.T) {
	router := NewRouter()

	beforeEnterGuard := NavigationGuard(func(to, from *Route, next NextFunc) {
		next(&NavigationTarget{}) // Cancel
	})

	to := &Route{
		Path: "/test",
		Name: "test",
		Meta: map[string]interface{}{
			"beforeEnter": beforeEnterGuard,
		},
	}

	result := router.executeBeforeGuards(to, nil)

	assert.Equal(t, guardCancel, result.action)
}

func TestExecuteBeforeGuards_BeforeEnterRedirects(t *testing.T) {
	router := NewRouter()

	beforeEnterGuard := NavigationGuard(func(to, from *Route, next NextFunc) {
		next(&NavigationTarget{Path: "/redirect"}) // Redirect
	})

	to := &Route{
		Path: "/test",
		Name: "test",
		Meta: map[string]interface{}{
			"beforeEnter": beforeEnterGuard,
		},
	}

	result := router.executeBeforeGuards(to, nil)

	assert.Equal(t, guardRedirect, result.action)
	require.NotNil(t, result.target)
	assert.Equal(t, "/redirect", result.target.Path)
}

// ===========================================================================
// Tests for router.go - GetHistoryEntries
// ===========================================================================

func TestRouter_GetHistoryEntries(t *testing.T) {
	router := NewRouter()

	t.Run("empty history", func(t *testing.T) {
		entries := router.GetHistoryEntries()
		assert.Empty(t, entries)
	})

	t.Run("with navigation history", func(t *testing.T) {
		// Register routes
		_ = router.registry.Register("/home", "home", nil)
		_ = router.registry.Register("/about", "about", nil)

		// Navigate
		cmd := router.Push(&NavigationTarget{Path: "/home"})
		cmd()

		cmd = router.Push(&NavigationTarget{Path: "/about"})
		cmd()

		entries := router.GetHistoryEntries()
		require.Len(t, entries, 2)
		assert.Equal(t, "/home", entries[0].Route.Path)
		assert.Equal(t, "/about", entries[1].Route.Path)
	})

	t.Run("returns defensive copy", func(t *testing.T) {
		entries := router.GetHistoryEntries()
		originalLen := len(entries)

		// Modifying the returned slice should not affect internal state
		if len(entries) > 0 {
			entries[0] = nil
		}

		newEntries := router.GetHistoryEntries()
		assert.Equal(t, originalLen, len(newEntries))
		if originalLen > 0 {
			assert.NotNil(t, newEntries[0])
		}
	})
}

// ===========================================================================
// Tests for router_view.go - Emit, On, KeyBindings, HelpText, IsInitialized
// ===========================================================================

func TestRouterView_ComponentInterfaceMethods(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	t.Run("Emit does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			rv.Emit("test-event", map[string]string{"key": "value"})
		})
	})

	t.Run("On does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			rv.On("test-event", func(data interface{}) {})
		})
	})

	t.Run("KeyBindings returns nil", func(t *testing.T) {
		bindings := rv.KeyBindings()
		assert.Nil(t, bindings)
	})

	t.Run("HelpText returns empty string", func(t *testing.T) {
		help := rv.HelpText()
		assert.Equal(t, "", help)
	})

	t.Run("IsInitialized returns true", func(t *testing.T) {
		initialized := rv.IsInitialized()
		assert.True(t, initialized)
	})
}

func TestRouterView_ViewWithNonComponentInterface(t *testing.T) {
	router := NewRouter()

	// Create route with non-bubbly.Component component
	type nonComponent struct{}
	route := &RouteRecord{
		Path:      "/test",
		Name:      "test",
		Component: &nonComponent{}, // Does not implement bubbly.Component
	}

	err := router.matcher.AddRouteRecord(route)
	require.NoError(t, err)

	match, err := router.matcher.Match("/test")
	require.NoError(t, err)

	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/test",
		Name:    "test",
		Params:  match.Params,
		Query:   make(map[string]string),
		Matched: match.Matched,
	}
	router.mu.Unlock()

	rv := NewRouterView(router, 0)
	output := rv.View()

	// Should return empty string when component doesn't implement bubbly.Component
	assert.Equal(t, "", output)
}

// ===========================================================================
// Tests for messages.go - isNavigationMsg marker methods
// ===========================================================================

func TestNavigationMsg_MarkerMethods(t *testing.T) {
	t.Run("RouteChangedMsg implements isNavigationMsg", func(t *testing.T) {
		msg := RouteChangedMsg{
			To:   &Route{Path: "/test"},
			From: nil,
		}

		// This should compile and not panic
		msg.isNavigationMsg()

		// Verify interface implementation
		var _ NavigationMsg = msg
	})

	t.Run("NavigationErrorMsg implements isNavigationMsg", func(t *testing.T) {
		msg := NavigationErrorMsg{
			Error: ErrNoMatch,
			To:    &NavigationTarget{Path: "/test"},
		}

		// This should compile and not panic
		msg.isNavigationMsg()

		// Verify interface implementation
		var _ NavigationMsg = msg
	})
}

// ===========================================================================
// Tests for pattern.go - SegmentKind.String() unknown case
// ===========================================================================

func TestSegmentKind_String_UnknownCase(t *testing.T) {
	// Test only the unknown case since other cases are tested in pattern_test.go
	unknownKind := SegmentKind(99) // Invalid/unknown kind
	result := unknownKind.String()
	assert.Equal(t, "unknown", result)
}

// ===========================================================================
// Tests for history_nav.go - Go() edge cases
// ===========================================================================

func TestRouter_Go_EdgeCases(t *testing.T) {
	t.Run("Go(0) returns nil", func(t *testing.T) {
		router := NewRouter()
		cmd := router.Go(0)
		assert.Nil(t, cmd)
	})

	t.Run("Go with empty history returns nil", func(t *testing.T) {
		router := NewRouter()
		cmd := router.Go(1)
		assert.Nil(t, cmd)

		cmd = router.Go(-1)
		assert.Nil(t, cmd)
	})

	t.Run("Go beyond bounds clamps to edges", func(t *testing.T) {
		router := NewRouter()

		// Register and navigate to create history
		_ = router.registry.Register("/home", "home", nil)
		_ = router.registry.Register("/about", "about", nil)
		_ = router.registry.Register("/contact", "contact", nil)

		cmd := router.Push(&NavigationTarget{Path: "/home"})
		cmd()

		cmd = router.Push(&NavigationTarget{Path: "/about"})
		cmd()

		cmd = router.Push(&NavigationTarget{Path: "/contact"})
		cmd()

		// Now at index 2, try to go +10 (should clamp to 2, no-op)
		cmd = router.Go(10)
		assert.Nil(t, cmd, "Go beyond end should be no-op when already at end")

		// Go back to beginning
		cmd = router.Go(-2)
		require.NotNil(t, cmd)
		cmd()

		// Try to go -10 (should clamp to 0, no-op)
		cmd = router.Go(-10)
		assert.Nil(t, cmd, "Go beyond start should be no-op when already at start")
	})

	t.Run("Go navigates correctly", func(t *testing.T) {
		router := NewRouter()

		_ = router.registry.Register("/page1", "page1", nil)
		_ = router.registry.Register("/page2", "page2", nil)

		cmd := router.Push(&NavigationTarget{Path: "/page1"})
		cmd()

		cmd = router.Push(&NavigationTarget{Path: "/page2"})
		cmd()

		// Go back
		cmd = router.Go(-1)
		require.NotNil(t, cmd)
		msg := cmd()

		changedMsg, ok := msg.(RouteChangedMsg)
		require.True(t, ok)
		assert.Equal(t, "/page1", changedMsg.To.Path)
	})
}

// ===========================================================================
// Tests for history.go - PushWithState truncation
// ===========================================================================

func TestHistory_PushWithState_Truncation(t *testing.T) {
	t.Run("truncates forward history", func(t *testing.T) {
		router := NewRouter()

		_ = router.registry.Register("/page1", "page1", nil)
		_ = router.registry.Register("/page2", "page2", nil)
		_ = router.registry.Register("/page3", "page3", nil)
		_ = router.registry.Register("/page4", "page4", nil)

		// Navigate to create history: page1 -> page2 -> page3
		cmd := router.Push(&NavigationTarget{Path: "/page1"})
		cmd()
		cmd = router.Push(&NavigationTarget{Path: "/page2"})
		cmd()
		cmd = router.Push(&NavigationTarget{Path: "/page3"})
		cmd()

		// Go back to page1
		cmd = router.Go(-2)
		require.NotNil(t, cmd)
		cmd()

		// Now navigate to page4 - should truncate forward history
		cmd = router.Push(&NavigationTarget{Path: "/page4"})
		cmd()

		entries := router.GetHistoryEntries()
		require.Len(t, entries, 2, "Forward history should be truncated")
		assert.Equal(t, "/page1", entries[0].Route.Path)
		assert.Equal(t, "/page4", entries[1].Route.Path)
	})
}

// ===========================================================================
// Tests for nested.go - resolveNestedPath and buildFullPath edge cases
// ===========================================================================

func TestResolveNestedPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		parentPath string
		childPath  string
		expected   string
	}{
		{
			name:       "empty child returns parent",
			parentPath: "/user/:id",
			childPath:  "",
			expected:   "/user/:id",
		},
		{
			name:       "parent with trailing slash",
			parentPath: "/dashboard/",
			childPath:  "/settings",
			expected:   "/dashboard/settings",
		},
		{
			name:       "child without leading slash",
			parentPath: "/dashboard",
			childPath:  "settings",
			expected:   "/dashboard/settings",
		},
		{
			name:       "root parent",
			parentPath: "/",
			childPath:  "/home",
			expected:   "/home",
		},
		{
			name:       "empty parent becomes root",
			parentPath: "",
			childPath:  "/test",
			expected:   "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveNestedPath(tt.parentPath, tt.childPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildMatchedArray_EdgeCases(t *testing.T) {
	t.Run("nil route returns nil", func(t *testing.T) {
		result := buildMatchedArray(nil)
		assert.Nil(t, result)
	})

	t.Run("route without parent", func(t *testing.T) {
		route := &RouteRecord{
			Path: "/test",
			Name: "test",
		}

		result := buildMatchedArray(route)
		require.Len(t, result, 1)
		assert.Equal(t, route, result[0])
	})

	t.Run("route with parent chain", func(t *testing.T) {
		grandparent := &RouteRecord{Path: "/a", Name: "a"}
		parent := &RouteRecord{Path: "/b", Name: "b", Parent: grandparent}
		child := &RouteRecord{Path: "/c", Name: "c", Parent: parent}

		result := buildMatchedArray(child)
		require.Len(t, result, 3)
		assert.Equal(t, grandparent, result[0])
		assert.Equal(t, parent, result[1])
		assert.Equal(t, child, result[2])
	})
}

func TestBuildFullPath_EdgeCases(t *testing.T) {
	t.Run("nil route returns empty", func(t *testing.T) {
		result := buildFullPath(nil)
		assert.Equal(t, "", result)
	})

	t.Run("route without parent", func(t *testing.T) {
		route := &RouteRecord{Path: "/test"}
		result := buildFullPath(route)
		assert.Equal(t, "/test", result)
	})

	t.Run("nested route builds full path", func(t *testing.T) {
		parent := &RouteRecord{Path: "/user/:id"}
		child := &RouteRecord{Path: "/profile", Parent: parent}

		result := buildFullPath(child)
		assert.Equal(t, "/user/:id/profile", result)
	})
}

// ===========================================================================
// Tests for component guards - executeComponentGuards edge cases
// ===========================================================================

// guardableComponent implements ComponentGuardable for testing
type guardableComponent struct {
	mockComponent
	beforeLeave  func(*Route, *Route, NextFunc)
	beforeEnter  func(*Route, *Route, NextFunc)
	beforeUpdate func(*Route, *Route, NextFunc)
}

func (g *guardableComponent) BeforeRouteLeave(to, from *Route, next NextFunc) {
	if g.beforeLeave != nil {
		g.beforeLeave(to, from, next)
	} else {
		next(nil)
	}
}

func (g *guardableComponent) BeforeRouteEnter(to, from *Route, next NextFunc) {
	if g.beforeEnter != nil {
		g.beforeEnter(to, from, next)
	} else {
		next(nil)
	}
}

func (g *guardableComponent) BeforeRouteUpdate(to, from *Route, next NextFunc) {
	if g.beforeUpdate != nil {
		g.beforeUpdate(to, from, next)
	} else {
		next(nil)
	}
}

var _ ComponentGuards = (*guardableComponent)(nil)

func TestExecuteComponentGuards_BeforeRouteLeave(t *testing.T) {
	router := NewRouter()

	leaveCalled := false
	oldComponent := &guardableComponent{
		mockComponent: mockComponent{name: "old", content: "old"},
		beforeLeave: func(to, from *Route, next NextFunc) {
			leaveCalled = true
			next(nil)
		},
	}

	newComponent := &mockComponent{name: "new", content: "new"}

	from := &Route{
		Path: "/old",
		Matched: []*RouteRecord{
			{Path: "/old", Component: oldComponent},
		},
	}

	to := &Route{
		Path: "/new",
		Matched: []*RouteRecord{
			{Path: "/new", Component: newComponent},
		},
	}

	result := router.executeComponentGuards(to, from)

	assert.True(t, leaveCalled, "BeforeRouteLeave should be called")
	assert.Equal(t, guardContinue, result.action)
}

func TestExecuteComponentGuards_BeforeRouteEnter(t *testing.T) {
	router := NewRouter()

	enterCalled := false
	newComponent := &guardableComponent{
		mockComponent: mockComponent{name: "new", content: "new"},
		beforeEnter: func(to, from *Route, next NextFunc) {
			enterCalled = true
			next(nil)
		},
	}

	to := &Route{
		Path: "/new",
		Matched: []*RouteRecord{
			{Path: "/new", Component: newComponent},
		},
	}

	result := router.executeComponentGuards(to, nil)

	assert.True(t, enterCalled, "BeforeRouteEnter should be called")
	assert.Equal(t, guardContinue, result.action)
}

func TestExecuteComponentGuards_BeforeRouteUpdate(t *testing.T) {
	router := NewRouter()

	updateCalled := false
	component := &guardableComponent{
		mockComponent: mockComponent{name: "same", content: "same"},
		beforeUpdate: func(to, from *Route, next NextFunc) {
			updateCalled = true
			next(nil)
		},
	}

	// Same component in both routes (component reused)
	from := &Route{
		Path: "/user/1",
		Matched: []*RouteRecord{
			{Path: "/user/:id", Component: component},
		},
	}

	to := &Route{
		Path: "/user/2",
		Matched: []*RouteRecord{
			{Path: "/user/:id", Component: component},
		},
	}

	result := router.executeComponentGuards(to, from)

	assert.True(t, updateCalled, "BeforeRouteUpdate should be called when component is reused")
	assert.Equal(t, guardContinue, result.action)
}

func TestExecuteComponentGuards_LeaveCancels(t *testing.T) {
	router := NewRouter()

	oldComponent := &guardableComponent{
		mockComponent: mockComponent{name: "old", content: "old"},
		beforeLeave: func(to, from *Route, next NextFunc) {
			next(&NavigationTarget{}) // Cancel
		},
	}

	newComponent := &mockComponent{name: "new", content: "new"}

	from := &Route{
		Path: "/old",
		Matched: []*RouteRecord{
			{Path: "/old", Component: oldComponent},
		},
	}

	to := &Route{
		Path: "/new",
		Matched: []*RouteRecord{
			{Path: "/new", Component: newComponent},
		},
	}

	result := router.executeComponentGuards(to, from)

	assert.Equal(t, guardCancel, result.action)
}

func TestExecuteComponentGuards_EnterCancels(t *testing.T) {
	router := NewRouter()

	newComponent := &guardableComponent{
		mockComponent: mockComponent{name: "new", content: "new"},
		beforeEnter: func(to, from *Route, next NextFunc) {
			next(&NavigationTarget{}) // Cancel
		},
	}

	to := &Route{
		Path: "/new",
		Matched: []*RouteRecord{
			{Path: "/new", Component: newComponent},
		},
	}

	result := router.executeComponentGuards(to, nil)

	assert.Equal(t, guardCancel, result.action)
}

func TestExecuteComponentGuards_UpdateCancels(t *testing.T) {
	router := NewRouter()

	component := &guardableComponent{
		mockComponent: mockComponent{name: "same", content: "same"},
		beforeUpdate: func(to, from *Route, next NextFunc) {
			next(&NavigationTarget{}) // Cancel
		},
	}

	from := &Route{
		Path: "/user/1",
		Matched: []*RouteRecord{
			{Path: "/user/:id", Component: component},
		},
	}

	to := &Route{
		Path: "/user/2",
		Matched: []*RouteRecord{
			{Path: "/user/:id", Component: component},
		},
	}

	result := router.executeComponentGuards(to, from)

	assert.Equal(t, guardCancel, result.action)
}

// ===========================================================================
// Tests for getLeafComponent
// ===========================================================================

func TestGetLeafComponent(t *testing.T) {
	t.Run("nil route returns nil", func(t *testing.T) {
		result := getLeafComponent(nil)
		assert.Nil(t, result)
	})

	t.Run("empty matched returns nil", func(t *testing.T) {
		route := &Route{
			Path:    "/test",
			Matched: []*RouteRecord{},
		}
		result := getLeafComponent(route)
		assert.Nil(t, result)
	})

	t.Run("returns last component", func(t *testing.T) {
		comp := &mockComponent{name: "leaf", content: "leaf"}
		route := &Route{
			Path: "/test",
			Matched: []*RouteRecord{
				{Path: "/", Component: &mockComponent{name: "root", content: "root"}},
				{Path: "/test", Component: comp},
			},
		}

		result := getLeafComponent(route)
		assert.Equal(t, comp, result)
	})
}

// ===========================================================================
// Tests for createNextCallback
// ===========================================================================

func TestCreateNextCallback(t *testing.T) {
	tests := []struct {
		name           string
		target         *NavigationTarget
		expectedAction guardAction
	}{
		{
			name:           "nil target means continue",
			target:         nil,
			expectedAction: guardContinue,
		},
		{
			name:           "empty target means cancel",
			target:         &NavigationTarget{},
			expectedAction: guardCancel,
		},
		{
			name:           "target with path means redirect",
			target:         &NavigationTarget{Path: "/redirect"},
			expectedAction: guardRedirect,
		},
		{
			name:           "target with name means redirect",
			target:         &NavigationTarget{Name: "redirect"},
			expectedAction: guardRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &guardResult{}
			next := createNextCallback(result)

			next(tt.target)

			assert.Equal(t, tt.expectedAction, result.action)
			if tt.expectedAction == guardRedirect {
				assert.Equal(t, tt.target, result.target)
			}
		})
	}
}

// ===========================================================================
// Tests for View interface verification
// ===========================================================================

func TestRouterView_ImplementsInterfaces(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	// Verify tea.Model
	var _ tea.Model = rv

	// Verify bubbly.Component
	var _ bubbly.Component = rv
}
