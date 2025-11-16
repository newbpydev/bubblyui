package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/stretchr/testify/assert"
)

// TestNewQueryParamsTester tests the constructor.
func TestNewQueryParamsTester(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates tester successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Create tester
			tester := NewQueryParamsTester(r)

			// Verify tester created
			assert.NotNil(t, tester)
			assert.NotNil(t, tester.router)
		})
	}
}

// TestQueryParamsTester_ParseFromURL tests parsing query params from URL.
func TestQueryParamsTester_ParseFromURL(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		query    map[string]string
		expected map[string]string
	}{
		{
			name:     "single param",
			path:     "/search",
			query:    map[string]string{"q": "golang"},
			expected: map[string]string{"q": "golang"},
		},
		{
			name:     "multiple params",
			path:     "/search",
			query:    map[string]string{"q": "golang", "page": "1", "sort": "date"},
			expected: map[string]string{"q": "golang", "page": "1", "sort": "date"},
		},
		{
			name:     "encoded params",
			path:     "/search",
			query:    map[string]string{"q": "hello world", "email": "test@example.com"},
			expected: map[string]string{"q": "hello world", "email": "test@example.com"},
		},
		{
			name:     "no params",
			path:     "/search",
			query:    map[string]string{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Navigate to path with query params
			cmd := r.Push(&router.NavigationTarget{
				Path:  tt.path,
				Query: tt.query,
			})
			msg := cmd()

			// Verify navigation succeeded
			_, ok := msg.(router.RouteChangedMsg)
			assert.True(t, ok, "Expected RouteChangedMsg")

			// Create tester
			tester := NewQueryParamsTester(r)

			// Assert query params
			tester.AssertQueryParams(t, tt.expected)
		})
	}
}

// TestQueryParamsTester_SetQueryParam tests setting individual query params.
func TestQueryParamsTester_SetQueryParam(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "set single param",
			key:      "q",
			value:    "golang",
			expected: "golang",
		},
		{
			name:     "set param with spaces",
			key:      "query",
			value:    "hello world",
			expected: "hello world",
		},
		{
			name:     "set param with special chars",
			key:      "email",
			value:    "test@example.com",
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Navigate to route
			cmd := r.Push(&router.NavigationTarget{Path: "/search"})
			cmd()

			// Create tester
			tester := NewQueryParamsTester(r)

			// Set query param
			tester.SetQueryParam(tt.key, tt.value)

			// Assert param was set
			tester.AssertQueryParam(t, tt.key, tt.expected)
		})
	}
}

// TestQueryParamsTester_AssertQueryParam tests asserting individual params.
func TestQueryParamsTester_AssertQueryParam(t *testing.T) {
	tests := []struct {
		name       string
		setupQuery map[string]string
		key        string
		expected   string
		shouldErr  bool
	}{
		{
			name:       "param exists and matches",
			setupQuery: map[string]string{"q": "golang"},
			key:        "q",
			expected:   "golang",
			shouldErr:  false,
		},
		{
			name:       "param missing",
			setupQuery: map[string]string{},
			key:        "q",
			expected:   "golang",
			shouldErr:  true,
		},
		{
			name:       "param value mismatch",
			setupQuery: map[string]string{"q": "rust"},
			key:        "q",
			expected:   "golang",
			shouldErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Navigate to path with query
			cmd := r.Push(&router.NavigationTarget{
				Path:  "/search",
				Query: tt.setupQuery,
			})
			cmd()

			// Create tester
			tester := NewQueryParamsTester(r)

			// Create mock testing.T to capture errors
			mockT := &mockTestingT{}

			// Assert query param
			tester.AssertQueryParam(mockT, tt.key, tt.expected)

			// Verify error expectation
			if tt.shouldErr {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			}
		})
	}
}

// TestQueryParamsTester_AssertQueryParams tests asserting multiple params.
func TestQueryParamsTester_AssertQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		setupQuery map[string]string
		expected   map[string]string
		shouldErr  bool
	}{
		{
			name:       "all params match",
			setupQuery: map[string]string{"q": "golang", "page": "1"},
			expected:   map[string]string{"q": "golang", "page": "1"},
			shouldErr:  false,
		},
		{
			name:       "missing param",
			setupQuery: map[string]string{"q": "golang"},
			expected:   map[string]string{"q": "golang", "page": "1"},
			shouldErr:  true,
		},
		{
			name:       "extra param in route",
			setupQuery: map[string]string{"q": "golang", "page": "1", "sort": "date"},
			expected:   map[string]string{"q": "golang", "page": "1"},
			shouldErr:  true,
		},
		{
			name:       "empty params",
			setupQuery: map[string]string{},
			expected:   map[string]string{},
			shouldErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Navigate to path with query
			cmd := r.Push(&router.NavigationTarget{
				Path:  "/search",
				Query: tt.setupQuery,
			})
			cmd()

			// Create tester
			tester := NewQueryParamsTester(r)

			// Create mock testing.T
			mockT := &mockTestingT{}

			// Assert query params
			tester.AssertQueryParams(mockT, tt.expected)

			// Verify error expectation
			if tt.shouldErr {
				assert.True(t, mockT.failed, "Expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "Expected assertion to pass")
			}
		})
	}
}

// TestQueryParamsTester_ClearQueryParams tests clearing all query params.
func TestQueryParamsTester_ClearQueryParams(t *testing.T) {
	tests := []struct {
		name       string
		setupQuery map[string]string
	}{
		{
			name:       "clear params from route with params",
			setupQuery: map[string]string{"q": "golang", "page": "1"},
		},
		{
			name:       "clear params from route without params",
			setupQuery: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Navigate to path with params
			cmd := r.Push(&router.NavigationTarget{
				Path:  "/search",
				Query: tt.setupQuery,
			})
			cmd()

			// Create tester
			tester := NewQueryParamsTester(r)

			// Clear params
			tester.ClearQueryParams()

			// Assert no params remain
			tester.AssertQueryParams(t, map[string]string{})
		})
	}
}

// TestQueryParamsTester_NavigationPreservesParams tests that navigation preserves params.
func TestQueryParamsTester_NavigationPreservesParams(t *testing.T) {
	tests := []struct {
		name          string
		initialPath   string
		initialParams map[string]string
		navigatePath  string
		navigateQuery map[string]string
		expectedQuery map[string]string
	}{
		{
			name:          "navigate with new params",
			initialPath:   "/search",
			initialParams: map[string]string{},
			navigatePath:  "/search",
			navigateQuery: map[string]string{"q": "golang"},
			expectedQuery: map[string]string{"q": "golang"},
		},
		{
			name:          "navigate updates existing params",
			initialPath:   "/search?q=rust",
			initialParams: map[string]string{"q": "rust"},
			navigatePath:  "/search",
			navigateQuery: map[string]string{"q": "golang"},
			expectedQuery: map[string]string{"q": "golang"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Initial navigation
			cmd := r.Push(&router.NavigationTarget{Path: tt.initialPath})
			cmd()

			// Create tester
			tester := NewQueryParamsTester(r)

			// Navigate with new params
			cmd = r.Push(&router.NavigationTarget{
				Path:  tt.navigatePath,
				Query: tt.navigateQuery,
			})
			cmd()

			// Assert params updated
			tester.AssertQueryParams(t, tt.expectedQuery)
		})
	}
}

// TestQueryParamsTester_ParamRemoval tests removing individual params.
func TestQueryParamsTester_ParamRemoval(t *testing.T) {
	tests := []struct {
		name          string
		setupQuery    map[string]string
		removeKey     string
		expectedQuery map[string]string
	}{
		{
			name:          "remove single param",
			setupQuery:    map[string]string{"q": "golang", "page": "1"},
			removeKey:     "page",
			expectedQuery: map[string]string{"q": "golang"},
		},
		{
			name:          "remove last param",
			setupQuery:    map[string]string{"q": "golang"},
			removeKey:     "q",
			expectedQuery: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			r, err := router.NewRouterBuilder().
				Route("/search", "search").
				Build()
			assert.NoError(t, err)

			// Navigate to path with params
			cmd := r.Push(&router.NavigationTarget{
				Path:  "/search",
				Query: tt.setupQuery,
			})
			cmd()

			// Create tester
			tester := NewQueryParamsTester(r)

			// Get current route and rebuild query without the key to remove
			route := r.CurrentRoute()
			newQuery := make(map[string]string)
			for k, v := range route.Query {
				if k != tt.removeKey {
					newQuery[k] = v
				}
			}

			// Navigate with updated query
			cmd = r.Push(&router.NavigationTarget{
				Path:  route.Path,
				Query: newQuery,
			})
			cmd()

			// Assert param removed
			tester.AssertQueryParams(t, tt.expectedQuery)
		})
	}
}
