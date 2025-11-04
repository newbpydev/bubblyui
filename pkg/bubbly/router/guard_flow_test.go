package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouter_CircularRedirect_Detection verifies circular redirect detection
func TestRouter_CircularRedirect_Detection(t *testing.T) {
	router := NewRouter()

	// Setup routes
	router.registry.Register("/a", "a", nil)
	router.registry.Register("/b", "b", nil)
	router.registry.Register("/c", "c", nil)

	// Setup circular redirect: /a -> /b -> /c -> /a
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		switch to.Path {
		case "/a":
			next(&NavigationTarget{Path: "/b"})
		case "/b":
			next(&NavigationTarget{Path: "/c"})
		case "/c":
			next(&NavigationTarget{Path: "/a"}) // Back to /a - circular!
		default:
			next(nil)
		}
	})

	// Attempt navigation
	cmd := router.Push(&NavigationTarget{Path: "/a"})
	msg := cmd()

	// Should detect circular redirect and return error
	errMsg, ok := msg.(NavigationErrorMsg)
	require.True(t, ok, "Should return NavigationErrorMsg")
	assert.Contains(t, errMsg.Error.Error(), "circular redirect", "Error should mention circular redirect")
}

// TestRouter_MaxRedirectDepth verifies redirect depth limit
func TestRouter_MaxRedirectDepth(t *testing.T) {
	router := NewRouter()

	// Setup many routes
	for i := 0; i < 15; i++ {
		router.registry.Register("/route"+string(rune('0'+i)), "route"+string(rune('0'+i)), nil)
	}

	// Setup deep redirect chain: /route0 -> /route1 -> /route2 -> ... -> /route14
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		if to.Path == "/route0" {
			next(&NavigationTarget{Path: "/route1"})
		} else if to.Path == "/route1" {
			next(&NavigationTarget{Path: "/route2"})
		} else if to.Path == "/route2" {
			next(&NavigationTarget{Path: "/route3"})
		} else if to.Path == "/route3" {
			next(&NavigationTarget{Path: "/route4"})
		} else if to.Path == "/route4" {
			next(&NavigationTarget{Path: "/route5"})
		} else if to.Path == "/route5" {
			next(&NavigationTarget{Path: "/route6"})
		} else if to.Path == "/route6" {
			next(&NavigationTarget{Path: "/route7"})
		} else if to.Path == "/route7" {
			next(&NavigationTarget{Path: "/route8"})
		} else if to.Path == "/route8" {
			next(&NavigationTarget{Path: "/route9"})
		} else if to.Path == "/route9" {
			next(&NavigationTarget{Path: "/route:"}) // 10th redirect
		} else if to.Path == "/route:" {
			next(&NavigationTarget{Path: "/route;"}) // 11th redirect - should fail
		} else {
			next(nil)
		}
	})

	// Attempt navigation
	cmd := router.Push(&NavigationTarget{Path: "/route0"})
	msg := cmd()

	// Should hit max redirect depth and return error
	errMsg, ok := msg.(NavigationErrorMsg)
	require.True(t, ok, "Should return NavigationErrorMsg")
	assert.Contains(t, errMsg.Error.Error(), "max redirect depth", "Error should mention max redirect depth")
}

// TestRouter_RedirectDepth_AllowedChain verifies allowed redirect chain
func TestRouter_RedirectDepth_AllowedChain(t *testing.T) {
	router := NewRouter()

	// Setup routes
	router.registry.Register("/start", "start", nil)
	router.registry.Register("/middle", "middle", nil)
	router.registry.Register("/end", "end", nil)

	// Setup short redirect chain: /start -> /middle -> /end (2 redirects, should succeed)
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		if to.Path == "/start" {
			next(&NavigationTarget{Path: "/middle"})
		} else if to.Path == "/middle" {
			next(&NavigationTarget{Path: "/end"})
		} else {
			next(nil)
		}
	})

	// Attempt navigation
	cmd := router.Push(&NavigationTarget{Path: "/start"})
	msg := cmd()

	// Should succeed
	changedMsg, ok := msg.(RouteChangedMsg)
	require.True(t, ok, "Should return RouteChangedMsg")
	assert.Equal(t, "/end", changedMsg.To.Path, "Should end at /end")
}

// TestRouter_CircularRedirect_SelfRedirect verifies self-redirect detection
func TestRouter_CircularRedirect_SelfRedirect(t *testing.T) {
	router := NewRouter()

	router.registry.Register("/loop", "loop", nil)

	// Setup self-redirect: /loop -> /loop
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		if to.Path == "/loop" {
			next(&NavigationTarget{Path: "/loop"}) // Redirect to self
		} else {
			next(nil)
		}
	})

	// Attempt navigation
	cmd := router.Push(&NavigationTarget{Path: "/loop"})
	msg := cmd()

	// Should detect circular redirect
	errMsg, ok := msg.(NavigationErrorMsg)
	require.True(t, ok, "Should return NavigationErrorMsg")
	assert.Contains(t, errMsg.Error.Error(), "circular redirect", "Error should mention circular redirect")
}

// TestRouter_RedirectTracking_ResetOnSuccess verifies redirect tracking resets
func TestRouter_RedirectTracking_ResetOnSuccess(t *testing.T) {
	router := NewRouter()

	router.registry.Register("/a", "a", nil)
	router.registry.Register("/b", "b", nil)
	router.registry.Register("/c", "c", nil)

	// First navigation: /a -> /b (1 redirect)
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		if to.Path == "/a" {
			next(&NavigationTarget{Path: "/b"})
		} else {
			next(nil)
		}
	})

	cmd := router.Push(&NavigationTarget{Path: "/a"})
	msg := cmd()
	require.IsType(t, RouteChangedMsg{}, msg, "First navigation should succeed")

	// Second navigation: /c (no redirect)
	cmd = router.Push(&NavigationTarget{Path: "/c"})
	msg = cmd()
	require.IsType(t, RouteChangedMsg{}, msg, "Second navigation should succeed")

	// Verify we're at /c
	currentRoute := router.CurrentRoute()
	require.NotNil(t, currentRoute)
	assert.Equal(t, "/c", currentRoute.Path)
}

// TestRouter_RedirectDepth_WorksWithReplace verifies redirect depth works with Replace
func TestRouter_RedirectDepth_WorksWithReplace(t *testing.T) {
	router := NewRouter()

	router.registry.Register("/a", "a", nil)
	router.registry.Register("/b", "b", nil)

	// Setup redirect
	router.BeforeEach(func(to, from *Route, next NextFunc) {
		if to.Path == "/a" {
			next(&NavigationTarget{Path: "/b"})
		} else {
			next(nil)
		}
	})

	// Use Replace instead of Push
	cmd := router.Replace(&NavigationTarget{Path: "/a"})
	msg := cmd()

	// Should succeed with redirect
	changedMsg, ok := msg.(RouteChangedMsg)
	require.True(t, ok, "Should return RouteChangedMsg")
	assert.Equal(t, "/b", changedMsg.To.Path, "Should end at /b")
}

// TestRouter_RedirectChain_ComplexScenario verifies complex redirect scenarios
func TestRouter_RedirectChain_ComplexScenario(t *testing.T) {
	tests := []struct {
		name          string
		startPath     string
		setupGuard    func(*Router)
		expectSuccess bool
		expectPath    string
		expectError   string
	}{
		{
			name:      "simple redirect",
			startPath: "/login",
			setupGuard: func(r *Router) {
				r.registry.Register("/login", "login", nil)
				r.registry.Register("/dashboard", "dashboard", nil)
				r.BeforeEach(func(to, from *Route, next NextFunc) {
					if to.Path == "/login" {
						next(&NavigationTarget{Path: "/dashboard"})
					} else {
						next(nil)
					}
				})
			},
			expectSuccess: true,
			expectPath:    "/dashboard",
		},
		{
			name:      "two-step redirect",
			startPath: "/old",
			setupGuard: func(r *Router) {
				r.registry.Register("/old", "old", nil)
				r.registry.Register("/new", "new", nil)
				r.registry.Register("/current", "current", nil)
				r.BeforeEach(func(to, from *Route, next NextFunc) {
					if to.Path == "/old" {
						next(&NavigationTarget{Path: "/new"})
					} else if to.Path == "/new" {
						next(&NavigationTarget{Path: "/current"})
					} else {
						next(nil)
					}
				})
			},
			expectSuccess: true,
			expectPath:    "/current",
		},
		{
			name:      "circular redirect A->B->A",
			startPath: "/a",
			setupGuard: func(r *Router) {
				r.registry.Register("/a", "a", nil)
				r.registry.Register("/b", "b", nil)
				r.BeforeEach(func(to, from *Route, next NextFunc) {
					if to.Path == "/a" {
						next(&NavigationTarget{Path: "/b"})
					} else if to.Path == "/b" {
						next(&NavigationTarget{Path: "/a"})
					} else {
						next(nil)
					}
				})
			},
			expectSuccess: false,
			expectError:   "circular redirect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			tt.setupGuard(router)

			cmd := router.Push(&NavigationTarget{Path: tt.startPath})
			msg := cmd()

			if tt.expectSuccess {
				changedMsg, ok := msg.(RouteChangedMsg)
				require.True(t, ok, "Should return RouteChangedMsg")
				assert.Equal(t, tt.expectPath, changedMsg.To.Path)
			} else {
				errMsg, ok := msg.(NavigationErrorMsg)
				require.True(t, ok, "Should return NavigationErrorMsg")
				assert.Contains(t, errMsg.Error.Error(), tt.expectError)
			}
		})
	}
}
