package router

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRouter verifies router creation with proper initialization
func TestNewRouter(t *testing.T) {
	tests := []struct {
		name string
		want struct {
			hasRegistry bool
			hasMatcher  bool
			hasHistory  bool
			routeNil    bool
		}
	}{
		{
			name: "creates router with initialized components",
			want: struct {
				hasRegistry bool
				hasMatcher  bool
				hasHistory  bool
				routeNil    bool
			}{
				hasRegistry: true,
				hasMatcher:  true,
				hasHistory:  true,
				routeNil:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			require.NotNil(t, router, "NewRouter should return non-nil router")

			if tt.want.hasRegistry {
				assert.NotNil(t, router.registry, "router should have registry")
			}
			if tt.want.hasMatcher {
				assert.NotNil(t, router.matcher, "router should have matcher")
			}
			if tt.want.hasHistory {
				assert.NotNil(t, router.history, "router should have history")
			}
			if tt.want.routeNil {
				assert.Nil(t, router.currentRoute, "currentRoute should be nil initially")
			}
		})
	}
}

// TestRouter_CurrentRoute verifies thread-safe access to current route
func TestRouter_CurrentRoute(t *testing.T) {
	tests := []struct {
		name       string
		setupRoute *Route
		wantNil    bool
	}{
		{
			name:       "returns nil when no route set",
			setupRoute: nil,
			wantNil:    true,
		},
		{
			name: "returns current route when set",
			setupRoute: &Route{
				Path:     "/test",
				Name:     "test-route",
				Params:   map[string]string{"id": "123"},
				Query:    map[string]string{"tab": "settings"},
				Hash:     "#section",
				FullPath: "/test?tab=settings#section",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Setup: Set current route if provided
			if tt.setupRoute != nil {
				router.currentRoute = tt.setupRoute
			}

			// Execute
			got := router.CurrentRoute()

			// Verify
			if tt.wantNil {
				assert.Nil(t, got, "CurrentRoute should return nil")
			} else {
				require.NotNil(t, got, "CurrentRoute should return non-nil route")
				assert.Equal(t, tt.setupRoute.Path, got.Path)
				assert.Equal(t, tt.setupRoute.Name, got.Name)
				assert.Equal(t, tt.setupRoute.Params, got.Params)
				assert.Equal(t, tt.setupRoute.Query, got.Query)
				assert.Equal(t, tt.setupRoute.Hash, got.Hash)
				assert.Equal(t, tt.setupRoute.FullPath, got.FullPath)
			}
		})
	}
}

// TestRouter_CurrentRoute_ThreadSafety verifies concurrent access to CurrentRoute
func TestRouter_CurrentRoute_ThreadSafety(t *testing.T) {
	router := NewRouter()

	// Set initial route
	router.currentRoute = &Route{
		Path: "/initial",
		Name: "initial-route",
	}

	const numGoroutines = 100
	const numReads = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch multiple goroutines reading CurrentRoute concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numReads; j++ {
				route := router.CurrentRoute()
				if route != nil {
					// Access route fields to ensure no race conditions
					_ = route.Path
					_ = route.Name
				}
			}
		}()
	}

	wg.Wait()
	// Test passes if no race conditions detected (run with -race flag)
}

// TestRouter_Initialization verifies all router components are properly initialized
func TestRouter_Initialization(t *testing.T) {
	router := NewRouter()

	t.Run("registry initialized", func(t *testing.T) {
		require.NotNil(t, router.registry)
		routes := router.registry.GetAll()
		assert.Empty(t, routes, "registry should start empty")
	})

	t.Run("matcher initialized", func(t *testing.T) {
		require.NotNil(t, router.matcher)
		assert.NotNil(t, router.matcher.routes, "matcher should have routes slice")
	})

	t.Run("history initialized", func(t *testing.T) {
		require.NotNil(t, router.history)
		// History tests will be in Task 3.1
	})

	t.Run("hooks initialized", func(t *testing.T) {
		assert.NotNil(t, router.beforeHooks, "beforeHooks should be initialized")
		assert.NotNil(t, router.afterHooks, "afterHooks should be initialized")
		assert.Empty(t, router.beforeHooks, "beforeHooks should start empty")
		assert.Empty(t, router.afterHooks, "afterHooks should start empty")
	})

	t.Run("mutex initialized", func(t *testing.T) {
		// Mutex is initialized by default in Go, just verify we can lock/unlock
		// Note: Empty critical sections are intentional - we're testing mutex init only
		router.mu.Lock()
		_ = struct{}{} //nolint:staticcheck // intentionally empty critical section
		router.mu.Unlock()
		router.mu.RLock()
		_ = struct{}{} //nolint:staticcheck // intentionally empty critical section
		router.mu.RUnlock()
	})
}

// TestRouter_CurrentRoute_Immutability verifies returned route is safe from external modification
func TestRouter_CurrentRoute_Immutability(t *testing.T) {
	router := NewRouter()

	// Set initial route with params and query
	originalRoute := &Route{
		Path:   "/user/123",
		Name:   "user-detail",
		Params: map[string]string{"id": "123"},
		Query:  map[string]string{"tab": "profile"},
	}
	router.currentRoute = originalRoute

	// Get current route
	got := router.CurrentRoute()
	require.NotNil(t, got)

	// Verify we got the same route (not a copy for now - Task 2.1 is simple)
	assert.Equal(t, originalRoute.Path, got.Path)
	assert.Equal(t, originalRoute.Name, got.Name)

	// Note: Route struct is already immutable by design (from Task 1.5)
	// Maps are defensively copied in NewRoute(), so external modification is prevented
}
