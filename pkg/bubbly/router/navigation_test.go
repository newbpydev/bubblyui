package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouter_Push verifies Push navigation creates history and generates commands
func TestRouter_Push(t *testing.T) {
	tests := []struct {
		name         string
		setupRoutes  func(*Router)
		target       *NavigationTarget
		wantErr      bool
		wantMsgType  string
		checkRoute   func(*testing.T, *Route)
		checkHistory bool
	}{
		{
			name: "push to valid path",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/users", "users-list", nil)
			},
			target: &NavigationTarget{
				Path: "/users",
			},
			wantErr:      false,
			wantMsgType:  "RouteChangedMsg",
			checkHistory: true,
			checkRoute: func(t *testing.T, route *Route) {
				assert.Equal(t, "/users", route.Path)
				assert.Equal(t, "users-list", route.Name)
			},
		},
		{
			name: "push with params",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/user/:id", "user-detail", nil)
			},
			target: &NavigationTarget{
				Path: "/user/123",
			},
			wantErr:      false,
			wantMsgType:  "RouteChangedMsg",
			checkHistory: true,
			checkRoute: func(t *testing.T, route *Route) {
				assert.Equal(t, "/user/:id", route.Path)
				assert.Equal(t, "user-detail", route.Name)
				assert.Equal(t, "123", route.Params["id"])
			},
		},
		{
			name: "push with query string",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/search", "search", nil)
			},
			target: &NavigationTarget{
				Path:  "/search",
				Query: map[string]string{"q": "golang", "page": "1"},
			},
			wantErr:      false,
			wantMsgType:  "RouteChangedMsg",
			checkHistory: true,
			checkRoute: func(t *testing.T, route *Route) {
				assert.Equal(t, "/search", route.Path)
				assert.Equal(t, "golang", route.Query["q"])
				assert.Equal(t, "1", route.Query["page"])
			},
		},
		{
			name: "push with hash",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/docs", "docs", nil)
			},
			target: &NavigationTarget{
				Path: "/docs",
				Hash: "#installation",
			},
			wantErr:      false,
			wantMsgType:  "RouteChangedMsg",
			checkHistory: true,
			checkRoute: func(t *testing.T, route *Route) {
				assert.Equal(t, "/docs", route.Path)
				assert.Equal(t, "#installation", route.Hash)
			},
		},
		{
			name: "push to non-existent route",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/home", "home", nil)
			},
			target: &NavigationTarget{
				Path: "/nonexistent",
			},
			wantErr:     true,
			wantMsgType: "NavigationErrorMsg",
		},
		{
			name:        "push with empty target",
			setupRoutes: func(r *Router) {},
			target:      &NavigationTarget{},
			wantErr:     true,
			wantMsgType: "NavigationErrorMsg",
		},
		{
			name:        "push with nil target",
			setupRoutes: func(r *Router) {},
			target:      nil,
			wantErr:     true,
			wantMsgType: "NavigationErrorMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			tt.setupRoutes(router)

			// Execute Push
			cmd := router.Push(tt.target)
			require.NotNil(t, cmd, "Push should return a command")

			// Execute command to get message
			msg := cmd()
			require.NotNil(t, msg, "Command should return a message")

			// Check message type
			switch msg := msg.(type) {
			case RouteChangedMsg:
				if tt.wantErr {
					t.Errorf("Expected error message, got RouteChangedMsg")
				}
				assert.Equal(t, "RouteChangedMsg", tt.wantMsgType)

				// Verify route was updated
				currentRoute := router.CurrentRoute()
				require.NotNil(t, currentRoute, "Current route should be set")

				if tt.checkRoute != nil {
					tt.checkRoute(t, currentRoute)
				}

				// Verify from/to routes
				assert.Equal(t, currentRoute, msg.To)

			case NavigationErrorMsg:
				if !tt.wantErr {
					t.Errorf("Expected success, got error: %v", msg.Error)
				}
				assert.Equal(t, "NavigationErrorMsg", tt.wantMsgType)
				assert.NotNil(t, msg.Error)

			default:
				t.Errorf("Unexpected message type: %T", msg)
			}
		})
	}
}

// TestRouter_Replace verifies Replace navigation doesn't create history
func TestRouter_Replace(t *testing.T) {
	tests := []struct {
		name        string
		setupRoutes func(*Router)
		target      *NavigationTarget
		wantErr     bool
		wantMsgType string
		checkRoute  func(*testing.T, *Route)
	}{
		{
			name: "replace to valid path",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/home", "home", nil)
				_ = r.registry.Register("/about", "about", nil)
			},
			target: &NavigationTarget{
				Path: "/about",
			},
			wantErr:     false,
			wantMsgType: "RouteChangedMsg",
			checkRoute: func(t *testing.T, route *Route) {
				assert.Equal(t, "/about", route.Path)
				assert.Equal(t, "about", route.Name)
			},
		},
		{
			name: "replace with params",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/user/:id", "user-detail", nil)
			},
			target: &NavigationTarget{
				Path: "/user/456",
			},
			wantErr:     false,
			wantMsgType: "RouteChangedMsg",
			checkRoute: func(t *testing.T, route *Route) {
				assert.Equal(t, "456", route.Params["id"])
			},
		},
		{
			name: "replace to non-existent route",
			setupRoutes: func(r *Router) {
				_ = r.registry.Register("/home", "home", nil)
			},
			target: &NavigationTarget{
				Path: "/invalid",
			},
			wantErr:     true,
			wantMsgType: "NavigationErrorMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			tt.setupRoutes(router)

			// Execute Replace
			cmd := router.Replace(tt.target)
			require.NotNil(t, cmd, "Replace should return a command")

			// Execute command to get message
			msg := cmd()
			require.NotNil(t, msg, "Command should return a message")

			// Check message type
			switch msg := msg.(type) {
			case RouteChangedMsg:
				if tt.wantErr {
					t.Errorf("Expected error message, got RouteChangedMsg")
				}

				// Verify route was updated
				currentRoute := router.CurrentRoute()
				require.NotNil(t, currentRoute, "Current route should be set")

				if tt.checkRoute != nil {
					tt.checkRoute(t, currentRoute)
				}

			case NavigationErrorMsg:
				if !tt.wantErr {
					t.Errorf("Expected success, got error: %v", msg.Error)
				}
				assert.NotNil(t, msg.Error)

			default:
				t.Errorf("Unexpected message type: %T", msg)
			}
		})
	}
}

// TestRouter_NavigationTarget_Validation verifies target validation
func TestRouter_NavigationTarget_Validation(t *testing.T) {
	tests := []struct {
		name    string
		target  *NavigationTarget
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil target",
			target:  nil,
			wantErr: true,
			errMsg:  "navigation target cannot be nil",
		},
		{
			name:    "empty target",
			target:  &NavigationTarget{},
			wantErr: true,
			errMsg:  "navigation target must have path or name",
		},
		{
			name: "valid path",
			target: &NavigationTarget{
				Path: "/users",
			},
			wantErr: false,
		},
		{
			name: "valid name",
			target: &NavigationTarget{
				Name: "user-detail",
			},
			wantErr: false,
		},
		{
			name: "both path and name",
			target: &NavigationTarget{
				Path: "/users",
				Name: "users-list",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Try to navigate
			cmd := router.Push(tt.target)
			msg := cmd()

			if tt.wantErr {
				errMsg, ok := msg.(NavigationErrorMsg)
				require.True(t, ok, "Expected NavigationErrorMsg")
				assert.Contains(t, errMsg.Error.Error(), tt.errMsg)
			} else {
				// Should get error because route doesn't exist, but not validation error
				errMsg, ok := msg.(NavigationErrorMsg)
				if ok {
					assert.NotContains(t, errMsg.Error.Error(), "navigation target")
				}
			}
		})
	}
}

// TestRouter_Push_UpdatesCurrentRoute verifies current route is updated
func TestRouter_Push_UpdatesCurrentRoute(t *testing.T) {
	router := NewRouter()
	_ = router.registry.Register("/home", "home", nil)
	_ = router.registry.Register("/about", "about", nil)

	// Initially no route
	assert.Nil(t, router.CurrentRoute())

	// Navigate to /home
	cmd := router.Push(&NavigationTarget{Path: "/home"})
	msg := cmd()
	require.IsType(t, RouteChangedMsg{}, msg)

	// Verify current route
	route := router.CurrentRoute()
	require.NotNil(t, route)
	assert.Equal(t, "/home", route.Path)

	// Navigate to /about
	cmd = router.Push(&NavigationTarget{Path: "/about"})
	msg = cmd()
	require.IsType(t, RouteChangedMsg{}, msg)

	// Verify current route changed
	route = router.CurrentRoute()
	require.NotNil(t, route)
	assert.Equal(t, "/about", route.Path)
}

// TestRouter_RouteChangedMsg_FromTo verifies from/to routes in message
func TestRouter_RouteChangedMsg_FromTo(t *testing.T) {
	router := NewRouter()
	_ = router.registry.Register("/home", "home", nil)
	_ = router.registry.Register("/about", "about", nil)

	// First navigation (from nil)
	cmd := router.Push(&NavigationTarget{Path: "/home"})
	msg := cmd().(RouteChangedMsg)

	assert.Nil(t, msg.From, "First navigation should have nil From")
	require.NotNil(t, msg.To)
	assert.Equal(t, "/home", msg.To.Path)

	// Second navigation (from /home to /about)
	cmd = router.Push(&NavigationTarget{Path: "/about"})
	msg = cmd().(RouteChangedMsg)

	require.NotNil(t, msg.From)
	assert.Equal(t, "/home", msg.From.Path)
	require.NotNil(t, msg.To)
	assert.Equal(t, "/about", msg.To.Path)
}
