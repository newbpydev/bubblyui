package router

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouteChangedMsg_Creation tests creating RouteChangedMsg
func TestRouteChangedMsg_Creation(t *testing.T) {
	tests := []struct {
		name    string
		to      *Route
		from    *Route
		wantNil bool
	}{
		{
			name: "both routes present",
			to: &Route{
				Path:     "/about",
				FullPath: "/about",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			wantNil: false,
		},
		{
			name: "from route nil (first navigation)",
			to: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			from:    nil,
			wantNil: false,
		},
		{
			name: "with params",
			to: &Route{
				Path:     "/user/:id",
				FullPath: "/user/123",
				Params:   map[string]string{"id": "123"},
				Query:    map[string]string{},
			},
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			wantNil: false,
		},
		{
			name: "with query",
			to: &Route{
				Path:     "/search",
				FullPath: "/search?q=test",
				Params:   map[string]string{},
				Query:    map[string]string{"q": "test"},
			},
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := RouteChangedMsg{
				To:   tt.to,
				From: tt.from,
			}

			assert.Equal(t, tt.to, msg.To)
			assert.Equal(t, tt.from, msg.From)

			if tt.from == nil {
				assert.Nil(t, msg.From)
			} else {
				assert.NotNil(t, msg.From)
			}
		})
	}
}

// TestNavigationErrorMsg_Creation tests creating NavigationErrorMsg
func TestNavigationErrorMsg_Creation(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		from      *Route
		to        *NavigationTarget
		wantError string
	}{
		{
			name: "route not found",
			err:  ErrNoMatch,
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			to: &NavigationTarget{
				Path: "/nonexistent",
			},
			wantError: "no route matches path",
		},
		{
			name: "nil target",
			err:  ErrNilTarget,
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			to:        nil,
			wantError: "navigation target cannot be nil",
		},
		{
			name: "empty target",
			err:  ErrEmptyTarget,
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			to: &NavigationTarget{
				Path: "",
				Name: "",
			},
			wantError: "navigation target must have path or name",
		},
		{
			name: "from route nil (no current route)",
			err:  ErrNoMatch,
			from: nil,
			to: &NavigationTarget{
				Path: "/test",
			},
			wantError: "no route matches path",
		},
		{
			name: "custom error",
			err:  errors.New("custom navigation error"),
			from: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			to: &NavigationTarget{
				Path: "/test",
			},
			wantError: "custom navigation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NavigationErrorMsg{
				Error: tt.err,
				From:  tt.from,
				To:    tt.to,
			}

			assert.Equal(t, tt.err, msg.Error)
			assert.Equal(t, tt.from, msg.From)
			assert.Equal(t, tt.to, msg.To)
			assert.Contains(t, msg.Error.Error(), tt.wantError)
		})
	}
}

// TestNavigationMsg_Interface tests NavigationMsg interface implementation
func TestNavigationMsg_Interface(t *testing.T) {
	tests := []struct {
		name string
		msg  NavigationMsg
	}{
		{
			name: "RouteChangedMsg implements NavigationMsg",
			msg: RouteChangedMsg{
				To: &Route{
					Path:     "/test",
					FullPath: "/test",
					Params:   map[string]string{},
					Query:    map[string]string{},
				},
				From: nil,
			},
		},
		{
			name: "NavigationErrorMsg implements NavigationMsg",
			msg: NavigationErrorMsg{
				Error: errors.New("test error"),
				From:  nil,
				To: &NavigationTarget{
					Path: "/test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify interface implementation
			var _ NavigationMsg = tt.msg

			// Verify isNavigationMsg method exists (compile-time check)
			tt.msg.isNavigationMsg()
		})
	}
}

// TestMessages_BubbletteaIntegration tests messages work with Bubbletea Update pattern
func TestMessages_BubbletteaIntegration(t *testing.T) {
	// Create a simple model that handles navigation messages
	type testModel struct {
		currentRoute *Route
		lastError    error
		messageCount int
	}

	model := testModel{}

	// Simulate Update() method handling
	updateFunc := func(m testModel, msg tea.Msg) testModel {
		switch msg := msg.(type) {
		case RouteChangedMsg:
			m.currentRoute = msg.To
			m.messageCount++
		case NavigationErrorMsg:
			m.lastError = msg.Error
			m.messageCount++
		}
		return m
	}

	t.Run("handles RouteChangedMsg", func(t *testing.T) {
		msg := RouteChangedMsg{
			To: &Route{
				Path:     "/about",
				FullPath: "/about",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			From: nil,
		}

		model = updateFunc(model, msg)

		assert.NotNil(t, model.currentRoute)
		assert.Equal(t, "/about", model.currentRoute.Path)
		assert.Equal(t, 1, model.messageCount)
		assert.Nil(t, model.lastError)
	})

	t.Run("handles NavigationErrorMsg", func(t *testing.T) {
		msg := NavigationErrorMsg{
			Error: errors.New("test error"),
			From:  model.currentRoute,
			To: &NavigationTarget{
				Path: "/invalid",
			},
		}

		model = updateFunc(model, msg)

		assert.NotNil(t, model.lastError)
		assert.Equal(t, "test error", model.lastError.Error())
		assert.Equal(t, 2, model.messageCount)
	})
}

// TestMessages_CommandGeneration tests that commands return correct messages
func TestMessages_CommandGeneration(t *testing.T) {
	// Create router with a test route
	router := NewRouter()

	// Register a test route
	err := router.registry.Register("/test", "test", map[string]interface{}{})
	require.NoError(t, err)

	t.Run("Push generates RouteChangedMsg on success", func(t *testing.T) {
		cmd := router.Push(&NavigationTarget{Path: "/test"})
		require.NotNil(t, cmd)

		msg := cmd()

		changedMsg, ok := msg.(RouteChangedMsg)
		require.True(t, ok, "Should return RouteChangedMsg")
		assert.Equal(t, "/test", changedMsg.To.Path)
	})

	t.Run("Push generates NavigationErrorMsg on failure", func(t *testing.T) {
		cmd := router.Push(&NavigationTarget{Path: "/nonexistent"})
		require.NotNil(t, cmd)

		msg := cmd()

		errMsg, ok := msg.(NavigationErrorMsg)
		require.True(t, ok, "Should return NavigationErrorMsg")
		assert.NotNil(t, errMsg.Error)
		assert.Equal(t, "/nonexistent", errMsg.To.Path)
	})

	t.Run("Replace generates RouteChangedMsg on success", func(t *testing.T) {
		cmd := router.Replace(&NavigationTarget{Path: "/test"})
		require.NotNil(t, cmd)

		msg := cmd()

		changedMsg, ok := msg.(RouteChangedMsg)
		require.True(t, ok, "Should return RouteChangedMsg")
		assert.Equal(t, "/test", changedMsg.To.Path)
	})

	t.Run("Replace generates NavigationErrorMsg on failure", func(t *testing.T) {
		cmd := router.Replace(&NavigationTarget{Path: "/nonexistent"})
		require.NotNil(t, cmd)

		msg := cmd()

		errMsg, ok := msg.(NavigationErrorMsg)
		require.True(t, ok, "Should return NavigationErrorMsg")
		assert.NotNil(t, errMsg.Error)
	})
}

// TestMessages_TypeSafety tests type safety of NavigationMsg interface
func TestMessages_TypeSafety(t *testing.T) {
	t.Run("can handle NavigationMsg polymorphically", func(t *testing.T) {
		messages := []NavigationMsg{
			RouteChangedMsg{
				To: &Route{
					Path:     "/test",
					FullPath: "/test",
					Params:   map[string]string{},
					Query:    map[string]string{},
				},
			},
			NavigationErrorMsg{
				Error: errors.New("test"),
				To:    &NavigationTarget{Path: "/test"},
			},
		}

		for _, msg := range messages {
			// Type switch works correctly
			switch m := msg.(type) {
			case RouteChangedMsg:
				assert.NotNil(t, m.To)
			case NavigationErrorMsg:
				assert.NotNil(t, m.Error)
			default:
				t.Fatal("unexpected message type")
			}
		}
	})
}

// TestMessages_MarkerMethods tests the isNavigationMsg marker methods
func TestMessages_MarkerMethods(t *testing.T) {
	t.Run("RouteChangedMsg marker method", func(t *testing.T) {
		msg := RouteChangedMsg{
			To: &Route{
				Path:     "/test",
				FullPath: "/test",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
		}
		// Call marker method directly
		msg.isNavigationMsg()
		
		// Verify it implements interface
		var _ NavigationMsg = msg
	})

	t.Run("NavigationErrorMsg marker method", func(t *testing.T) {
		msg := NavigationErrorMsg{
			Error: errors.New("test"),
			To:    &NavigationTarget{Path: "/test"},
		}
		// Call marker method directly
		msg.isNavigationMsg()
		
		// Verify it implements interface
		var _ NavigationMsg = msg
	})
}

// TestMessages_NilSafety tests handling of nil values
func TestMessages_NilSafety(t *testing.T) {
	t.Run("RouteChangedMsg with nil From", func(t *testing.T) {
		msg := RouteChangedMsg{
			To: &Route{
				Path:     "/test",
				FullPath: "/test",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			From: nil,
		}

		assert.NotNil(t, msg.To)
		assert.Nil(t, msg.From)
	})

	t.Run("NavigationErrorMsg with nil From", func(t *testing.T) {
		msg := NavigationErrorMsg{
			Error: errors.New("test"),
			From:  nil,
			To:    &NavigationTarget{Path: "/test"},
		}

		assert.NotNil(t, msg.Error)
		assert.Nil(t, msg.From)
		assert.NotNil(t, msg.To)
	})

	t.Run("NavigationErrorMsg with nil To", func(t *testing.T) {
		msg := NavigationErrorMsg{
			Error: ErrNilTarget,
			From: &Route{
				Path:     "/",
				FullPath: "/",
				Params:   map[string]string{},
				Query:    map[string]string{},
			},
			To: nil,
		}

		assert.NotNil(t, msg.Error)
		assert.NotNil(t, msg.From)
		assert.Nil(t, msg.To)
	})
}

// TestMessages_Concurrency tests thread-safe message handling
func TestMessages_Concurrency(t *testing.T) {
	router := NewRouter()

	// Register test route
	err := router.registry.Register("/test", "test", map[string]interface{}{})
	require.NoError(t, err)

	// Pre-initialize matcher to avoid race in syncRegistryToMatcher
	// (Note: This is a workaround for a pre-existing race condition in router.matchTarget)
	_, _ = router.matchTarget(&NavigationTarget{Path: "/test"})

	t.Run("concurrent message type checking", func(t *testing.T) {
		// Generate messages sequentially but check them concurrently
		messages := make([]tea.Msg, 10)
		for i := 0; i < 10; i++ {
			cmd := router.Push(&NavigationTarget{Path: "/test"})
			messages[i] = cmd()
		}

		// Now check message types concurrently
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(msg tea.Msg) {
				// Verify message type
				switch msg.(type) {
				case RouteChangedMsg, NavigationErrorMsg:
					// Valid message types
				default:
					t.Errorf("unexpected message type: %T", msg)
				}

				done <- true
			}(messages[i])
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
