package router

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// mockComponent is a simple component for testing
type mockComponent struct {
	name    string
	content string
}

func (m *mockComponent) Init() tea.Cmd {
	return nil
}

func (m *mockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *mockComponent) View() string {
	return m.content
}

func (m *mockComponent) Name() string {
	return m.name
}

func (m *mockComponent) ID() string {
	return m.name + "-id"
}

func (m *mockComponent) Props() interface{} {
	return nil
}

func (m *mockComponent) Emit(event string, data interface{}) {
}

func (m *mockComponent) On(event string, handler bubbly.EventHandler) {
}

func (m *mockComponent) KeyBindings() map[string][]bubbly.KeyBinding {
	return nil
}

func (m *mockComponent) HelpText() string {
	return ""
}

func (m *mockComponent) IsInitialized() bool {
	return true
}

// Ensure mockComponent implements bubbly.Component
var _ bubbly.Component = (*mockComponent)(nil)

// TestNewRouterView tests RouterView creation
func TestNewRouterView(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name      string
		router    *Router
		depth     int
		wantDepth int
	}{
		{
			name:      "depth 0 (root)",
			router:    router,
			depth:     0,
			wantDepth: 0,
		},
		{
			name:      "depth 1 (nested)",
			router:    router,
			depth:     1,
			wantDepth: 1,
		},
		{
			name:      "depth 2 (deeply nested)",
			router:    router,
			depth:     2,
			wantDepth: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := NewRouterView(tt.router, tt.depth)

			assert.NotNil(t, rv)
			assert.Equal(t, tt.router, rv.router)
			assert.Equal(t, tt.wantDepth, rv.depth)
		})
	}
}

// TestRouterView_RendersCurrentComponent tests that RouterView renders the matched component
func TestRouterView_RendersCurrentComponent(t *testing.T) {
	router := NewRouter()

	// Create a simple component
	homeComponent := &mockComponent{
		name:    "Home",
		content: "Home Page",
	}

	// Register route with component
	route := &RouteRecord{
		Path:      "/",
		Name:      "home",
		Component: homeComponent,
	}

	err := router.matcher.AddRouteRecord(route)
	require.NoError(t, err)

	// Navigate to the route
	match, err := router.matcher.Match("/")
	require.NoError(t, err)

	// Set current route
	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/",
		Name:    "home",
		Params:  match.Params,
		Query:   make(map[string]string),
		Hash:    "",
		Meta:    route.Meta,
		Matched: match.Matched,
	}
	router.mu.Unlock()

	// Create RouterView at depth 0
	rv := NewRouterView(router, 0)

	// Render should return the component's view
	output := rv.View()
	assert.Equal(t, "Home Page", output)
}

// TestRouterView_HandlesDepthForNesting tests depth-based rendering
func TestRouterView_HandlesDepthForNesting(t *testing.T) {
	router := NewRouter()

	// Create parent and child components
	parentComponent := &mockComponent{
		name:    "Dashboard",
		content: "Dashboard Layout",
	}

	childComponent := &mockComponent{
		name:    "Settings",
		content: "Settings Page",
	}

	// Create nested routes
	childRoute := Child("/settings",
		WithName("dashboard-settings"),
		WithComponent(childComponent),
	)

	parentRoute := &RouteRecord{
		Path:      "/dashboard",
		Name:      "dashboard",
		Component: parentComponent,
		Children:  []*RouteRecord{childRoute},
	}

	err := router.matcher.AddRouteRecord(parentRoute)
	require.NoError(t, err)

	// Navigate to child route
	match, err := router.matcher.Match("/dashboard/settings")
	require.NoError(t, err)

	// Set current route
	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/dashboard/settings",
		Name:    "dashboard-settings",
		Params:  match.Params,
		Query:   make(map[string]string),
		Hash:    "",
		Meta:    childRoute.Meta,
		Matched: match.Matched,
	}
	router.mu.Unlock()

	// RouterView at depth 0 should render parent
	rv0 := NewRouterView(router, 0)
	output0 := rv0.View()
	assert.Equal(t, "Dashboard Layout", output0)

	// RouterView at depth 1 should render child
	rv1 := NewRouterView(router, 1)
	output1 := rv1.View()
	assert.Equal(t, "Settings Page", output1)
}

// TestRouterView_HandlesNoMatch tests behavior when no route is matched
func TestRouterView_HandlesNoMatch(t *testing.T) {
	router := NewRouter()

	// No current route
	rv := NewRouterView(router, 0)

	// Should return empty string
	output := rv.View()
	assert.Equal(t, "", output)
}

// TestRouterView_HandlesDepthOutOfBounds tests behavior when depth exceeds matched routes
func TestRouterView_HandlesDepthOutOfBounds(t *testing.T) {
	router := NewRouter()

	// Create a simple route
	homeComponent := &mockComponent{
		name:    "Home",
		content: "Home Page",
	}

	route := &RouteRecord{
		Path:      "/",
		Name:      "home",
		Component: homeComponent,
	}

	err := router.matcher.AddRouteRecord(route)
	require.NoError(t, err)

	// Navigate to the route
	match, err := router.matcher.Match("/")
	require.NoError(t, err)

	// Set current route
	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/",
		Name:    "home",
		Params:  match.Params,
		Query:   make(map[string]string),
		Hash:    "",
		Meta:    route.Meta,
		Matched: match.Matched,
	}
	router.mu.Unlock()

	// RouterView at depth 0 should work
	rv0 := NewRouterView(router, 0)
	output0 := rv0.View()
	assert.Equal(t, "Home Page", output0)

	// RouterView at depth 1 (out of bounds) should return empty
	rv1 := NewRouterView(router, 1)
	output1 := rv1.View()
	assert.Equal(t, "", output1)
}

// TestRouterView_HandlesNoComponent tests behavior when route has no component
func TestRouterView_HandlesNoComponent(t *testing.T) {
	router := NewRouter()

	// Create route without component
	route := &RouteRecord{
		Path: "/",
		Name: "home",
		// No Component field set
	}

	err := router.matcher.AddRouteRecord(route)
	require.NoError(t, err)

	// Navigate to the route
	match, err := router.matcher.Match("/")
	require.NoError(t, err)

	// Set current route
	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/",
		Name:    "home",
		Params:  match.Params,
		Query:   make(map[string]string),
		Hash:    "",
		Meta:    route.Meta,
		Matched: match.Matched,
	}
	router.mu.Unlock()

	// RouterView should return empty string
	rv := NewRouterView(router, 0)
	output := rv.View()
	assert.Equal(t, "", output)
}

// TestRouterView_UpdatesOnRouteChange tests that RouterView responds to route changes
func TestRouterView_UpdatesOnRouteChange(t *testing.T) {
	router := NewRouter()

	// Create two components
	homeComponent := &mockComponent{
		name:    "Home",
		content: "Home Page",
	}

	aboutComponent := &mockComponent{
		name:    "About",
		content: "About Page",
	}

	// Register routes
	homeRoute := &RouteRecord{
		Path:      "/",
		Name:      "home",
		Component: homeComponent,
	}

	aboutRoute := &RouteRecord{
		Path:      "/about",
		Name:      "about",
		Component: aboutComponent,
	}

	err := router.matcher.AddRouteRecord(homeRoute)
	require.NoError(t, err)

	err = router.matcher.AddRouteRecord(aboutRoute)
	require.NoError(t, err)

	// Navigate to home
	match, err := router.matcher.Match("/")
	require.NoError(t, err)

	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/",
		Name:    "home",
		Params:  match.Params,
		Query:   make(map[string]string),
		Hash:    "",
		Meta:    homeRoute.Meta,
		Matched: match.Matched,
	}
	router.mu.Unlock()

	// Create RouterView
	rv := NewRouterView(router, 0)

	// Should render home
	output := rv.View()
	assert.Equal(t, "Home Page", output)

	// Navigate to about
	match, err = router.matcher.Match("/about")
	require.NoError(t, err)

	router.mu.Lock()
	router.currentRoute = &Route{
		Path:    "/about",
		Name:    "about",
		Params:  match.Params,
		Query:   make(map[string]string),
		Hash:    "",
		Meta:    aboutRoute.Meta,
		Matched: match.Matched,
	}
	router.mu.Unlock()

	// Should now render about
	output = rv.View()
	assert.Equal(t, "About Page", output)
}

// TestWithComponent tests the WithComponent option
func TestWithComponent(t *testing.T) {
	component := &mockComponent{
		name:    "Test",
		content: "Test Content",
	}

	route := &RouteRecord{
		Path: "/test",
		Name: "test",
	}

	// Apply WithComponent option
	opt := WithComponent(component)
	opt(route)

	assert.Equal(t, component, route.Component)
}

// TestRouterView_Init tests the Init method
func TestRouterView_Init(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	cmd := rv.Init()
	assert.Nil(t, cmd, "Init should return nil")
}

// TestRouterView_Update tests the Update method
func TestRouterView_Update(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "key message",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
		},
		{
			name: "route changed message",
			msg: RouteChangedMsg{
				To:   &Route{Path: "/test", FullPath: "/test"},
				From: nil,
			},
		},
		{
			name: "nil message",
			msg:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, cmd := rv.Update(tt.msg)
			assert.Equal(t, rv, model, "Update should return same RouterView")
			assert.Nil(t, cmd, "Update should return nil command")
		})
	}
}

// TestRouterView_Name tests the Name method
func TestRouterView_Name(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	name := rv.Name()
	assert.Equal(t, "RouterView", name)
}

// TestRouterView_ID tests the ID method
func TestRouterView_ID(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name     string
		depth    int
		expected string
	}{
		{
			name:     "depth 0",
			depth:    0,
			expected: "router-view-0",
		},
		{
			name:     "depth 1",
			depth:    1,
			expected: "router-view-1",
		},
		{
			name:     "depth 2",
			depth:    2,
			expected: "router-view-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := NewRouterView(router, tt.depth)
			id := rv.ID()
			assert.Equal(t, tt.expected, id)
		})
	}
}

// TestRouterView_Props tests the Props method
func TestRouterView_Props(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	props := rv.Props()
	assert.Nil(t, props, "Props should return nil")
}

// TestRouterView_Emit tests the Emit method
func TestRouterView_Emit(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	// Should not panic
	rv.Emit("test-event", "test-data")
	rv.Emit("", nil)
}

// TestRouterView_On tests the On method
func TestRouterView_On(t *testing.T) {
	router := NewRouter()
	rv := NewRouterView(router, 0)

	handlerCalled := false
	handler := func(data interface{}) {
		handlerCalled = true
	}

	// Should not panic
	rv.On("test-event", handler)
	rv.On("", nil)

	// Handler should never be called since RouterView doesn't handle events
	assert.False(t, handlerCalled, "Handler should not be called")
}
