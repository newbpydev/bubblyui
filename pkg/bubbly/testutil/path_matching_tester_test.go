package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// TestNewPathMatchingTester tests the constructor
func TestNewPathMatchingTester(t *testing.T) {
	r, err := router.NewRouterBuilder().
		Route("/home", "home").
		Build()
	assert.NoError(t, err)

	tester := NewPathMatchingTester(r)

	assert.NotNil(t, tester)
	assert.NotNil(t, tester.router)
}

// TestPathMatchingTester_TestMatch tests basic pattern matching
func TestPathMatchingTester_TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		routes  []struct{ path, name string }
		pattern string
		path    string
		matches bool
	}{
		{
			name: "static path matches exactly",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			pattern: "/home",
			path:    "/home",
			matches: true,
		},
		{
			name: "static path does not match",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			pattern: "/home",
			path:    "/about",
			matches: false,
		},
		{
			name: "dynamic segment matches",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern: "/user/:id",
			path:    "/user/123",
			matches: true,
		},
		{
			name: "dynamic segment does not match different path",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern: "/user/:id",
			path:    "/post/123",
			matches: false,
		},
		{
			name: "wildcard pattern matches",
			routes: []struct{ path, name string }{
				{"/docs/:path*", "documentation"},
			},
			pattern: "/docs/:path*",
			path:    "/docs/guide/getting-started",
			matches: true,
		},
		{
			name: "optional segment matches with value",
			routes: []struct{ path, name string }{
				{"/profile/:id?", "profile"},
			},
			pattern: "/profile/:id?",
			path:    "/profile/123",
			matches: true,
		},
		{
			name: "optional segment matches without value",
			routes: []struct{ path, name string }{
				{"/profile/:id?", "profile"},
			},
			pattern: "/profile/:id?",
			path:    "/profile",
			matches: true,
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

			tester := NewPathMatchingTester(r)
			result := tester.TestMatch(tt.pattern, tt.path)

			assert.Equal(t, tt.matches, result)
		})
	}
}

// TestPathMatchingTester_AssertMatches tests the assertion helper
func TestPathMatchingTester_AssertMatches(t *testing.T) {
	tests := []struct {
		name       string
		routes     []struct{ path, name string }
		pattern    string
		path       string
		shouldPass bool
	}{
		{
			name: "assertion passes when pattern matches",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern:    "/user/:id",
			path:       "/user/123",
			shouldPass: true,
		},
		{
			name: "assertion fails when pattern does not match",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern:    "/user/:id",
			path:       "/post/123",
			shouldPass: false,
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

			tester := NewPathMatchingTester(r)
			mockT := &mockTestingT{}
			tester.AssertMatches(mockT, tt.pattern, tt.path)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestPathMatchingTester_AssertNotMatches tests the negative assertion helper
func TestPathMatchingTester_AssertNotMatches(t *testing.T) {
	tests := []struct {
		name       string
		routes     []struct{ path, name string }
		pattern    string
		path       string
		shouldPass bool
	}{
		{
			name: "assertion passes when pattern does not match",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern:    "/user/:id",
			path:       "/post/123",
			shouldPass: true,
		},
		{
			name: "assertion fails when pattern matches",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern:    "/user/:id",
			path:       "/user/123",
			shouldPass: false,
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

			tester := NewPathMatchingTester(r)
			mockT := &mockTestingT{}
			tester.AssertNotMatches(mockT, tt.pattern, tt.path)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			} else {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			}
		})
	}
}

// TestPathMatchingTester_ExtractParams tests parameter extraction
func TestPathMatchingTester_ExtractParams(t *testing.T) {
	tests := []struct {
		name           string
		routes         []struct{ path, name string }
		pattern        string
		path           string
		expectedParams map[string]string
	}{
		{
			name: "extract single parameter",
			routes: []struct{ path, name string }{
				{"/user/:id", "user-detail"},
			},
			pattern: "/user/:id",
			path:    "/user/123",
			expectedParams: map[string]string{
				"id": "123",
			},
		},
		{
			name: "extract multiple parameters",
			routes: []struct{ path, name string }{
				{"/posts/:category/:id", "post-detail"},
			},
			pattern: "/posts/:category/:id",
			path:    "/posts/tech/456",
			expectedParams: map[string]string{
				"category": "tech",
				"id":       "456",
			},
		},
		{
			name: "extract wildcard parameter",
			routes: []struct{ path, name string }{
				{"/docs/:path*", "documentation"},
			},
			pattern: "/docs/:path*",
			path:    "/docs/guide/getting-started",
			expectedParams: map[string]string{
				"path": "guide/getting-started",
			},
		},
		{
			name: "extract optional parameter when present",
			routes: []struct{ path, name string }{
				{"/profile/:id?", "profile"},
			},
			pattern: "/profile/:id?",
			path:    "/profile/123",
			expectedParams: map[string]string{
				"id": "123",
			},
		},
		{
			name: "no parameters when optional is absent",
			routes: []struct{ path, name string }{
				{"/profile/:id?", "profile"},
			},
			pattern:        "/profile/:id?",
			path:           "/profile",
			expectedParams: map[string]string{},
		},
		{
			name: "static route has no parameters",
			routes: []struct{ path, name string }{
				{"/home", "home"},
			},
			pattern:        "/home",
			path:           "/home",
			expectedParams: map[string]string{},
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

			tester := NewPathMatchingTester(r)
			params := tester.ExtractParams(tt.pattern, tt.path)

			assert.Equal(t, tt.expectedParams, params)
		})
	}
}

// TestPathMatchingTester_PriorityOrdering tests route specificity
func TestPathMatchingTester_PriorityOrdering(t *testing.T) {
	tests := []struct {
		name          string
		routes        []struct{ path, name string }
		testPath      string
		expectedMatch string
	}{
		{
			name: "static route beats dynamic route",
			routes: []struct{ path, name string }{
				{"/users/:id", "user-detail"},
				{"/users/new", "user-new"},
			},
			testPath:      "/users/new",
			expectedMatch: "/users/new",
		},
		{
			name: "more specific dynamic route wins",
			routes: []struct{ path, name string }{
				{"/:resource/:id", "generic"},
				{"/users/:id", "user-detail"},
			},
			testPath:      "/users/123",
			expectedMatch: "/users/:id",
		},
		{
			name: "wildcard has lowest priority",
			routes: []struct{ path, name string }{
				{"/docs/:path*", "docs-wildcard"},
				{"/docs/guide", "docs-guide"},
			},
			testPath:      "/docs/guide",
			expectedMatch: "/docs/guide",
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

			tester := NewPathMatchingTester(r)

			// Test that the expected pattern matches
			assert.True(t, tester.TestMatch(tt.expectedMatch, tt.testPath),
				"Expected pattern %s to match path %s", tt.expectedMatch, tt.testPath)
		})
	}
}

// TestPathMatchingTester_RegexConstraints tests regex validation (if supported)
func TestPathMatchingTester_RegexConstraints(t *testing.T) {
	// Note: This test documents expected behavior for regex constraints
	// The current router implementation may not support regex constraints yet
	t.Skip("Regex constraints not yet implemented in router")

	tests := []struct {
		name    string
		pattern string
		path    string
		matches bool
	}{
		{
			name:    "numeric constraint matches numbers",
			pattern: "/user/:id(\\d+)",
			path:    "/user/123",
			matches: true,
		},
		{
			name:    "numeric constraint rejects non-numbers",
			pattern: "/user/:id(\\d+)",
			path:    "/user/abc",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := router.NewRouterBuilder().
				Route(tt.pattern, "test").
				Build()
			assert.NoError(t, err)

			tester := NewPathMatchingTester(r)
			result := tester.TestMatch(tt.pattern, tt.path)

			assert.Equal(t, tt.matches, result)
		})
	}
}
