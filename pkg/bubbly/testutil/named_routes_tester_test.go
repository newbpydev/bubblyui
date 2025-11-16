package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/stretchr/testify/assert"
)

// TestNewNamedRoutesTester tests the constructor
func TestNewNamedRoutesTester(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Build()
	assert.NoError(t, err)

	tester := NewNamedRoutesTester(r)

	assert.NotNil(t, tester)
	assert.NotNil(t, tester.router)
}

// TestNamedRoutesTester_AssertRouteExists tests route existence verification
func TestNamedRoutesTester_AssertRouteExists(t *testing.T) {
	tests := []struct {
		name       string
		routes     []struct{ path, name string }
		routeName  string
		shouldPass bool
	}{
		{
			name: "route exists",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			routeName:  "home",
			shouldPass: true,
		},
		{
			name: "route does not exist",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			routeName:  "nonexistent",
			shouldPass: false,
		},
		{
			name: "multiple routes registered",
			routes: []struct{ path, name string }{
				{"/home", "home"},
				{"/about", "about"},
				{"/contact", "contact"},
			},
			routeName:  "about",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := router.NewRouterBuilder()
			for _, route := range tt.routes {
				rb.Route(route.path, route.name)
			}
			r, err := rb.Build()
			assert.NoError(t, err)

			tester := NewNamedRoutesTester(r)

			mockT := &mockTestingT{}
			tester.AssertRouteExists(mockT, tt.routeName)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestNamedRoutesTester_NavigateByName tests navigation by route name
func TestNamedRoutesTester_NavigateByName(t *testing.T) {
	tests := []struct {
		name          string
		routes        []struct{ path, name string }
		routeName     string
		params        map[string]string
		shouldSucceed bool
	}{
		{
			name: "navigate to static route",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			routeName:     "home",
			params:        nil,
			shouldSucceed: true,
		},
		{
			name: "navigate with params",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			routeName: "user-detail",
			params: map[string]string{
				"id": "123",
			},
			shouldSucceed: true,
		},
		{
			name: "navigate with multiple params",
			routes: []struct{ path, name string }{
				{"/posts/:category/:id", "post-detail"},
			},
			routeName: "post-detail",
			params: map[string]string{
				"category": "tech",
				"id":       "456",
			},
			shouldSucceed: true,
		},
		{
			name: "navigate to nonexistent route",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			routeName:     "nonexistent",
			params:        nil,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := router.NewRouterBuilder()
			for _, route := range tt.routes {
				rb.Route(route.path, route.name)
			}
			r, err := rb.Build()
			assert.NoError(t, err)

			tester := NewNamedRoutesTester(r)

			tester.NavigateByName(tt.routeName, tt.params)

			if tt.shouldSucceed {
				currentRoute := r.CurrentRoute()
				assert.NotNil(t, currentRoute)
				// Check that navigation succeeded and params were extracted
				assert.Equal(t, tt.routeName, currentRoute.Name)
				// Verify params were extracted correctly
				for key, expectedValue := range tt.params {
					assert.Equal(t, expectedValue, currentRoute.Params[key])
				}
			}
		})
	}
}

// TestNamedRoutesTester_AssertRouteName tests current route name assertion
func TestNamedRoutesTester_AssertRouteName(t *testing.T) {
	tests := []struct {
		name         string
		routes       []struct{ path, name string }
		navigateTo   string
		expectedName string
		shouldPass   bool
		skipNav      bool // Skip navigation to test no route case
	}{
		{
			name: "correct route name",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			navigateTo:   "/home",
			expectedName: "home",
			shouldPass:   true,
		},
		{
			name: "incorrect route name",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			navigateTo:   "/home",
			expectedName: "wrong-name",
			shouldPass:   false,
		},
		{
			name: "multiple routes",
			routes: []struct{ path, name string }{
				{"/home", "home"},
				{"/about", "about"},
			},
			navigateTo:   "/about",
			expectedName: "about",
			shouldPass:   true,
		},
		{
			name: "no current route",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			expectedName: "home",
			shouldPass:   false,
			skipNav:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := router.NewRouterBuilder()
			for _, route := range tt.routes {
				rb.Route(route.path, route.name)
			}
			r, err := rb.Build()
			assert.NoError(t, err)

			// Navigate to route (unless skipNav is true)
			if !tt.skipNav {
				cmd := r.Push(&router.NavigationTarget{Path: tt.navigateTo})
				cmd() // Execute the command
			}

			tester := NewNamedRoutesTester(r)
			mockT := &mockTestingT{}
			tester.AssertRouteName(mockT, tt.expectedName)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestNamedRoutesTester_GetRouteURL tests URL generation from route name
func TestNamedRoutesTester_GetRouteURL(t *testing.T) {
	tests := []struct {
		name        string
		routes      []struct{ path, name string }
		routeName   string
		params      map[string]string
		expectedURL string
		expectError bool
	}{
		{
			name: "static route",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			routeName:   "home",
			params:      nil,
			expectedURL: "/home",
			expectError: false,
		},
		{
			name: "route with single param",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			routeName: "user-detail",
			params: map[string]string{
				"id": "123",
			},
			expectedURL: "/user/123",
			expectError: false,
		},
		{
			name: "route with multiple params",
			routes: []struct{ path, name string }{
				{"/posts/:category/:id", "post-detail"},
			},
			routeName: "post-detail",
			params: map[string]string{
				"category": "tech",
				"id":       "456",
			},
			expectedURL: "/posts/tech/456",
			expectError: false,
		},
		{
			name: "nonexistent route",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			routeName:   "nonexistent",
			params:      nil,
			expectedURL: "",
			expectError: true,
		},
		{
			name: "missing required param",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			routeName:   "user-detail",
			params:      nil,
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := router.NewRouterBuilder()
			for _, route := range tt.routes {
				rb.Route(route.path, route.name)
			}
			r, err := rb.Build()
			assert.NoError(t, err)

			tester := NewNamedRoutesTester(r)

			url, err := tester.GetRouteURL(tt.routeName, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

// TestNamedRoutesTester_NameUniqueness tests that duplicate names are handled
func TestNamedRoutesTester_NameUniqueness(t *testing.T) {
	// RouterBuilder should error on duplicate names
	rb := router.NewRouterBuilder()
	rb.Route("/home", "home")
	rb.Route("/other", "home") // Duplicate name

	_, err := rb.Build()
	assert.Error(t, err, "Should error on duplicate name")
	assert.Contains(t, err.Error(), "duplicate name")
}

// TestNamedRoutesTester_AliasRoutes tests routes with aliases (same path, different names)
func TestNamedRoutesTester_AliasRoutes(t *testing.T) {
	// Note: Router doesn't support aliases (same path with different names)
	// This test documents the expected behavior
	rb := router.NewRouterBuilder()
	rb.Route("/home", "home")
	rb.Route("/home", "index") // Duplicate path

	_, err := rb.Build()
	assert.Error(t, err, "Should error on duplicate path")
	assert.Contains(t, err.Error(), "duplicate path")
}
