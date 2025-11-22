package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// guardTracker tracks which guards were called and in what order
type guardTracker struct {
	calls []string
}

func (gt *guardTracker) record(name string) {
	gt.calls = append(gt.calls, name)
}

func (gt *guardTracker) reset() {
	gt.calls = nil
}

// mockComponentWithGuards is a mock component that implements ComponentGuards
type mockComponentWithGuards struct {
	mockComponent    // Embed the basic mock component
	tracker          *guardTracker
	cancelOnEnter    bool
	cancelOnUpdate   bool
	cancelOnLeave    bool
	redirectOnEnter  string
	redirectOnUpdate string
	redirectOnLeave  string
	redirectOnlyTo   string // Only redirect when navigating to this path
	// Legacy fields for backward compatibility
	shouldCancel bool
	redirectTo   string
}

func (m *mockComponentWithGuards) BeforeRouteEnter(to, from *Route, next NextFunc) {
	m.tracker.record("BeforeRouteEnter")
	// Check specific flag first, fall back to legacy flag
	if m.cancelOnEnter || (m.shouldCancel && !m.cancelOnUpdate && !m.cancelOnLeave) {
		next(&NavigationTarget{Path: ""})
	} else if m.redirectOnEnter != "" {
		next(&NavigationTarget{Path: m.redirectOnEnter})
	} else if m.redirectTo != "" && m.redirectOnUpdate == "" && m.redirectOnLeave == "" {
		next(&NavigationTarget{Path: m.redirectTo})
	} else {
		next(nil)
	}
}

func (m *mockComponentWithGuards) BeforeRouteUpdate(to, from *Route, next NextFunc) {
	m.tracker.record("BeforeRouteUpdate")
	if m.cancelOnUpdate {
		next(&NavigationTarget{Path: ""})
	} else if m.redirectOnUpdate != "" {
		next(&NavigationTarget{Path: m.redirectOnUpdate})
	} else {
		next(nil)
	}
}

func (m *mockComponentWithGuards) BeforeRouteLeave(to, from *Route, next NextFunc) {
	m.tracker.record("BeforeRouteLeave")
	if m.cancelOnLeave {
		next(&NavigationTarget{Path: ""})
	} else if m.redirectOnLeave != "" {
		// Only redirect if going to specific path (if specified)
		if m.redirectOnlyTo == "" || (to != nil && to.Path == m.redirectOnlyTo) {
			next(&NavigationTarget{Path: m.redirectOnLeave})
		} else {
			next(nil)
		}
	} else {
		next(nil)
	}
}

// Ensure mockComponentWithGuards implements ComponentGuards
var _ ComponentGuards = (*mockComponentWithGuards)(nil)

// TestHasComponentGuards tests the hasComponentGuards helper function
func TestHasComponentGuards(t *testing.T) {
	tracker := &guardTracker{}

	tests := []struct {
		name          string
		component     interface{}
		wantHasGuards bool
	}{
		{
			name:          "component with guards",
			component:     &mockComponentWithGuards{tracker: tracker},
			wantHasGuards: true,
		},
		{
			name:          "component without guards",
			component:     &mockComponent{name: "test"},
			wantHasGuards: false,
		},
		{
			name:          "nil component",
			component:     nil,
			wantHasGuards: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guards, ok := hasComponentGuards(tt.component)

			assert.Equal(t, tt.wantHasGuards, ok)
			if tt.wantHasGuards {
				assert.NotNil(t, guards)
			} else {
				assert.Nil(t, guards)
			}
		})
	}
}

// TestComponentGuards_BeforeRouteEnter tests BeforeRouteEnter execution
func TestComponentGuards_BeforeRouteEnter(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	// Create component with guards
	component := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test", content: "Test Content"},
		tracker:       tracker,
	}

	// Register route with component
	route := &RouteRecord{
		Path:      "/test",
		Name:      "test",
		Component: component,
	}

	err := router.matcher.AddRouteRecord(route)
	require.NoError(t, err)

	// Navigate to the route
	router.Push(&NavigationTarget{Path: "/test"})()

	// Execute the command to trigger navigation

	// BeforeRouteEnter should have been called
	assert.Contains(t, tracker.calls, "BeforeRouteEnter")
}

// TestComponentGuards_BeforeRouteLeave tests BeforeRouteLeave execution
func TestComponentGuards_BeforeRouteLeave(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	// Create two components with guards
	component1 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test1", content: "Content1"},
		tracker:       tracker,
	}

	component2 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test2", content: "Content2"},
		tracker:       tracker,
	}

	// Register routes
	route1 := &RouteRecord{
		Path:      "/test1",
		Name:      "test1",
		Component: component1,
	}

	route2 := &RouteRecord{
		Path:      "/test2",
		Name:      "test2",
		Component: component2,
	}

	err := router.matcher.AddRouteRecord(route1)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(route2)
	require.NoError(t, err)

	// Navigate to first route
	router.Push(&NavigationTarget{Path: "/test1"})()
	require.NoError(t, err)

	tracker.reset()

	// Navigate to second route - should call BeforeRouteLeave on first component
	router.Push(&NavigationTarget{Path: "/test2"})()
	require.NoError(t, err)

	// BeforeRouteLeave should have been called on component1
	assert.Contains(t, tracker.calls, "BeforeRouteLeave")
	// BeforeRouteEnter should have been called on component2
	assert.Contains(t, tracker.calls, "BeforeRouteEnter")
}

// TestComponentGuards_BeforeRouteUpdate tests BeforeRouteUpdate execution
func TestComponentGuards_BeforeRouteUpdate(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	// Create component with guards
	component := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "User", content: "User Content"},
		tracker:       tracker,
	}

	// Register route with dynamic parameter
	route := &RouteRecord{
		Path:      "/user/:id",
		Name:      "user",
		Component: component,
	}

	err := router.matcher.AddRouteRecord(route)
	require.NoError(t, err)

	// Navigate to /user/1
	router.Push(&NavigationTarget{Path: "/user/1"})()
	require.NoError(t, err)

	tracker.reset()

	// Navigate to /user/2 - same component, different params
	// This should trigger BeforeRouteUpdate
	router.Push(&NavigationTarget{Path: "/user/2"})()
	require.NoError(t, err)

	// BeforeRouteUpdate should have been called
	assert.Contains(t, tracker.calls, "BeforeRouteUpdate")
}

// TestComponentGuards_ExecutionOrder tests that guards execute in correct order
func TestComponentGuards_ExecutionOrder(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	// Create two components with guards
	component1 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test1", content: "Content1"},
		tracker:       tracker,
	}

	component2 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test2", content: "Content2"},
		tracker:       tracker,
	}

	// Register routes
	route1 := &RouteRecord{
		Path:      "/test1",
		Name:      "test1",
		Component: component1,
	}

	route2 := &RouteRecord{
		Path:      "/test2",
		Name:      "test2",
		Component: component2,
	}

	err := router.matcher.AddRouteRecord(route1)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(route2)
	require.NoError(t, err)

	// Navigate to first route
	router.Push(&NavigationTarget{Path: "/test1"})()
	require.NoError(t, err)

	tracker.reset()

	// Navigate to second route
	router.Push(&NavigationTarget{Path: "/test2"})()
	require.NoError(t, err)

	// Verify execution order: BeforeRouteLeave (old) before BeforeRouteEnter (new)
	require.Len(t, tracker.calls, 2)
	assert.Equal(t, "BeforeRouteLeave", tracker.calls[0])
	assert.Equal(t, "BeforeRouteEnter", tracker.calls[1])
}

// TestComponentGuards_CancelNavigation tests canceling navigation from guards
func TestComponentGuards_CancelNavigation(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	// Create component that cancels navigation on leave
	component1 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test1", content: "Content1"},
		tracker:       tracker,
		cancelOnLeave: true, // Cancel navigation when leaving
	}

	component2 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test2", content: "Content2"},
		tracker:       tracker,
	}

	// Register routes
	route1 := &RouteRecord{
		Path:      "/test1",
		Name:      "test1",
		Component: component1,
	}

	route2 := &RouteRecord{
		Path:      "/test2",
		Name:      "test2",
		Component: component2,
	}

	err := router.matcher.AddRouteRecord(route1)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(route2)
	require.NoError(t, err)

	// Navigate to first route
	router.Push(&NavigationTarget{Path: "/test1"})()
	require.NoError(t, err)

	// Get current route
	currentRoute := router.CurrentRoute()
	require.NotNil(t, currentRoute)
	assert.Equal(t, "/test1", currentRoute.Path)

	tracker.reset()

	// Try to navigate to second route - should be canceled by component1's BeforeRouteLeave
	router.Push(&NavigationTarget{Path: "/test2"})()

	// Navigation should be canceled (error or still on first route)
	currentRoute = router.CurrentRoute()
	require.NotNil(t, currentRoute)
	// Should still be on first route
	assert.Equal(t, "/test1", currentRoute.Path)

	// BeforeRouteLeave should have been called
	assert.Contains(t, tracker.calls, "BeforeRouteLeave")
	// BeforeRouteEnter should NOT have been called (navigation canceled)
	assert.NotContains(t, tracker.calls, "BeforeRouteEnter")
}

// TestComponentGuards_RedirectFromGuard tests redirecting from a guard
func TestComponentGuards_RedirectFromGuard(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	// Create component that redirects on leave (only when going to test2)
	component1 := &mockComponentWithGuards{
		mockComponent:   mockComponent{name: "Test1", content: "Content1"},
		tracker:         tracker,
		redirectOnLeave: "/test3", // Redirect to test3
		redirectOnlyTo:  "/test2", // Only when trying to go to test2
	}

	component2 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test2", content: "Content2"},
		tracker:       tracker,
	}

	component3 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test3", content: "Content3"},
		tracker:       tracker,
	}

	// Register routes
	route1 := &RouteRecord{
		Path:      "/test1",
		Name:      "test1",
		Component: component1,
	}

	route2 := &RouteRecord{
		Path:      "/test2",
		Name:      "test2",
		Component: component2,
	}

	route3 := &RouteRecord{
		Path:      "/test3",
		Name:      "test3",
		Component: component3,
	}

	err := router.matcher.AddRouteRecord(route1)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(route2)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(route3)
	require.NoError(t, err)

	// Navigate to first route
	router.Push(&NavigationTarget{Path: "/test1"})()
	require.NoError(t, err)

	tracker.reset()

	// Try to navigate to second route - should be redirected to test3 by component1
	router.Push(&NavigationTarget{Path: "/test2"})()
	require.NoError(t, err)

	// Should be on test3 (redirected)
	currentRoute := router.CurrentRoute()
	require.NotNil(t, currentRoute)
	assert.Equal(t, "/test3", currentRoute.Path)
}

// TestComponentGuards_NoGuards tests navigation with components that don't implement guards
func TestComponentGuards_NoGuards(t *testing.T) {
	router := NewRouter()

	// Create regular component without guards
	component1 := &mockComponent{name: "Test1", content: "Content1"}
	component2 := &mockComponent{name: "Test2", content: "Content2"}

	// Register routes
	route1 := &RouteRecord{
		Path:      "/test1",
		Name:      "test1",
		Component: component1,
	}

	route2 := &RouteRecord{
		Path:      "/test2",
		Name:      "test2",
		Component: component2,
	}

	err := router.matcher.AddRouteRecord(route1)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(route2)
	require.NoError(t, err)

	// Navigate to first route
	router.Push(&NavigationTarget{Path: "/test1"})()
	require.NoError(t, err)

	// Navigate to second route - should work without guards
	router.Push(&NavigationTarget{Path: "/test2"})()
	require.NoError(t, err)

	// Should be on second route
	currentRoute := router.CurrentRoute()
	require.NotNil(t, currentRoute)
	assert.Equal(t, "/test2", currentRoute.Path)
}

// TestDebug_ComponentGuards debugs component guard execution
func TestDebug_ComponentGuards(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	component := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test", content: "Test Content"},
		tracker:       tracker,
	}

	route := &RouteRecord{
		Path:      "/test",
		Name:      "test",
		Component: component,
	}

	err := router.matcher.AddRouteRecord(route)
	if err != nil {
		t.Fatalf("AddRouteRecord failed: %v", err)
	}

	// Check what's in the matcher
	match, err := router.matcher.Match("/test")
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	t.Logf("Match.Route.Component: %v", match.Route.Component)
	t.Logf("Match.Matched length: %d", len(match.Matched))
	if len(match.Matched) > 0 {
		t.Logf("Match.Matched[0].Component: %v", match.Matched[0].Component)
	}

	// Test if guard is callable directly
	guards, ok := hasComponentGuards(component)
	if !ok {
		t.Fatal("Component should implement ComponentGuards")
	}

	// Call guard directly
	guards.BeforeRouteEnter(nil, nil, func(target *NavigationTarget) {})
	t.Logf("After direct call, tracker calls: %v", tracker.calls)

	tracker.reset()

	// Now try navigation
	msg := router.Push(&NavigationTarget{Path: "/test"})()
	t.Logf("Navigation message type: %T", msg)
	if errMsg, ok := msg.(NavigationErrorMsg); ok {
		t.Logf("Navigation error: %v", errMsg.Error)
	}

	t.Logf("After navigation, tracker calls: %v", tracker.calls)

	// Check current route
	currentRoute := router.CurrentRoute()
	if currentRoute != nil {
		t.Logf("Current route Matched length: %d", len(currentRoute.Matched))
		if len(currentRoute.Matched) > 0 {
			t.Logf("Current route Matched[0].Component: %v", currentRoute.Matched[0].Component)
		}
	} else {
		t.Log("Current route is nil")
	}
}

// TestDebug_CancelFlow debugs the cancellation flow
func TestDebug_CancelFlow(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	component1 := &mockComponentWithGuards{
		mockComponent: mockComponent{name: "Test1", content: "Content1"},
		tracker:       tracker,
		shouldCancel:  true,
	}

	route1 := &RouteRecord{
		Path:      "/test1",
		Name:      "test1",
		Component: component1,
	}

	route2 := &RouteRecord{
		Path:      "/test2",
		Name:      "test2",
		Component: &mockComponent{name: "Test2", content: "Content2"},
	}

	router.matcher.AddRouteRecord(route1)
	router.matcher.AddRouteRecord(route2)

	// Navigate to first route
	msg1 := router.Push(&NavigationTarget{Path: "/test1"})()
	t.Logf("First navigation: %T", msg1)

	// Try to navigate away (should be canceled)
	msg2 := router.Push(&NavigationTarget{Path: "/test2"})()
	t.Logf("Second navigation: %T", msg2)

	if errMsg, ok := msg2.(NavigationErrorMsg); ok {
		t.Logf("Navigation error: %v", errMsg.Error)
	}

	currentRoute := router.CurrentRoute()
	if currentRoute != nil {
		t.Logf("Current route: %s", currentRoute.Path)
	}

	t.Logf("Tracker calls: %v", tracker.calls)
}

// TestDebug_RedirectFlow debugs the redirect flow
func TestDebug_RedirectFlow(t *testing.T) {
	tracker := &guardTracker{}
	router := NewRouter()

	component1 := &mockComponentWithGuards{
		mockComponent:   mockComponent{name: "Test1", content: "Content1"},
		tracker:         tracker,
		redirectOnLeave: "/test3",
	}

	route1 := &RouteRecord{Path: "/test1", Name: "test1", Component: component1}
	route2 := &RouteRecord{Path: "/test2", Name: "test2", Component: &mockComponent{name: "Test2", content: "Content2"}}
	route3 := &RouteRecord{Path: "/test3", Name: "test3", Component: &mockComponent{name: "Test3", content: "Content3"}}

	router.matcher.AddRouteRecord(route1)
	router.matcher.AddRouteRecord(route2)
	router.matcher.AddRouteRecord(route3)

	// Navigate to first route
	msg1 := router.Push(&NavigationTarget{Path: "/test1"})()
	t.Logf("First navigation: %T", msg1)
	t.Logf("Current route after first: %s", router.CurrentRoute().Path)

	tracker.reset()

	// Try to navigate to test2 (should redirect to test3)
	msg2 := router.Push(&NavigationTarget{Path: "/test2"})()
	t.Logf("Second navigation: %T", msg2)

	if errMsg, ok := msg2.(NavigationErrorMsg); ok {
		t.Logf("Navigation error: %v", errMsg.Error)
	}

	currentRoute := router.CurrentRoute()
	if currentRoute != nil {
		t.Logf("Current route after second: %s", currentRoute.Path)
	}

	t.Logf("Tracker calls: %v", tracker.calls)
}
