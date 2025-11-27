package devtools

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// TestRouterDebugger_NewRouterDebugger tests constructor
func TestRouterDebugger_NewRouterDebugger(t *testing.T) {
	rd := NewRouterDebugger(100)

	assert.NotNil(t, rd)
	assert.Nil(t, rd.GetCurrentRoute())
	assert.Equal(t, 0, rd.GetHistoryCount())
	assert.Equal(t, 100, rd.maxSize)
}

// TestRouterDebugger_RecordNavigation tests navigation recording
func TestRouterDebugger_RecordNavigation(t *testing.T) {
	tests := []struct {
		name     string
		from     *router.Route
		to       *router.Route
		duration time.Duration
		success  bool
	}{
		{
			name:     "successful navigation",
			from:     router.NewRoute("/home", "home", nil, nil, "", nil, nil),
			to:       router.NewRoute("/about", "about", nil, nil, "", nil, nil),
			duration: 5 * time.Millisecond,
			success:  true,
		},
		{
			name:     "failed navigation",
			from:     router.NewRoute("/home", "home", nil, nil, "", nil, nil),
			to:       router.NewRoute("/admin", "admin", nil, nil, "", nil, nil),
			duration: 2 * time.Millisecond,
			success:  false,
		},
		{
			name:     "initial navigation (no from)",
			from:     nil,
			to:       router.NewRoute("/home", "home", nil, nil, "", nil, nil),
			duration: 1 * time.Millisecond,
			success:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := NewRouterDebugger(10)

			rd.RecordNavigation(tt.from, tt.to, tt.duration, tt.success)

			history := rd.GetHistory()
			require.Len(t, history, 1)

			record := history[0]
			assert.Equal(t, tt.from, record.From)
			assert.Equal(t, tt.to, record.To)
			assert.Equal(t, tt.duration, record.Duration)
			assert.Equal(t, tt.success, record.Success)
			assert.False(t, record.Timestamp.IsZero())

			// Current route should be updated
			assert.Equal(t, tt.to, rd.GetCurrentRoute())
		})
	}
}

// TestRouterDebugger_RecordGuard tests guard execution recording
func TestRouterDebugger_RecordGuard(t *testing.T) {
	tests := []struct {
		name      string
		guardName string
		result    GuardResult
		duration  time.Duration
	}{
		{
			name:      "guard allows",
			guardName: "authGuard",
			result:    GuardAllow,
			duration:  1 * time.Millisecond,
		},
		{
			name:      "guard cancels",
			guardName: "permissionGuard",
			result:    GuardCancel,
			duration:  2 * time.Millisecond,
		},
		{
			name:      "guard redirects",
			guardName: "loginGuard",
			result:    GuardRedirect,
			duration:  3 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := NewRouterDebugger(10)

			rd.RecordGuard(tt.guardName, tt.result, tt.duration)

			guards := rd.GetGuards()
			require.Len(t, guards, 1)

			guard := guards[0]
			assert.Equal(t, tt.guardName, guard.Name)
			assert.Equal(t, tt.result, guard.Result)
			assert.Equal(t, tt.duration, guard.Duration)
			assert.False(t, guard.Timestamp.IsZero())
		})
	}
}

// TestRouterDebugger_MaxSize tests circular buffer enforcement
func TestRouterDebugger_MaxSize(t *testing.T) {
	rd := NewRouterDebugger(3)

	// Add 5 navigation records
	for i := 0; i < 5; i++ {
		route := router.NewRoute("/page", "page", map[string]string{"id": string(rune('0' + i))}, nil, "", nil, nil)
		rd.RecordNavigation(nil, route, time.Millisecond, true)
	}

	// Should only keep last 3
	history := rd.GetHistory()
	assert.Len(t, history, 3)

	// Verify it's the last 3 (ids 2, 3, 4)
	assert.Equal(t, "2", history[0].To.Params["id"])
	assert.Equal(t, "3", history[1].To.Params["id"])
	assert.Equal(t, "4", history[2].To.Params["id"])
}

// TestRouterDebugger_Clear tests clearing all data
func TestRouterDebugger_Clear(t *testing.T) {
	rd := NewRouterDebugger(10)

	// Add some data
	route := router.NewRoute("/home", "home", nil, nil, "", nil, nil)
	rd.RecordNavigation(nil, route, time.Millisecond, true)
	rd.RecordGuard("authGuard", GuardAllow, time.Millisecond)

	// Verify data exists
	assert.NotNil(t, rd.GetCurrentRoute())
	assert.Equal(t, 1, rd.GetHistoryCount())
	assert.Len(t, rd.GetGuards(), 1)

	// Clear
	rd.Clear()

	// Verify cleared
	assert.Nil(t, rd.GetCurrentRoute())
	assert.Equal(t, 0, rd.GetHistoryCount())
	assert.Len(t, rd.GetGuards(), 0)
}

// TestRouterDebugger_GetHistory tests history retrieval
func TestRouterDebugger_GetHistory(t *testing.T) {
	rd := NewRouterDebugger(10)

	// Add multiple records
	for i := 0; i < 3; i++ {
		route := router.NewRoute("/page", "page", map[string]string{"id": string(rune('0' + i))}, nil, "", nil, nil)
		rd.RecordNavigation(nil, route, time.Millisecond, true)
	}

	history := rd.GetHistory()
	assert.Len(t, history, 3)

	// Verify copy (modifying returned slice doesn't affect internal state)
	history[0].Success = false
	assert.True(t, rd.GetHistory()[0].Success)
}

// TestRouterDebugger_Concurrent tests thread safety
func TestRouterDebugger_Concurrent(t *testing.T) {
	rd := NewRouterDebugger(1000)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			route := router.NewRoute("/page", "page", map[string]string{"id": string(rune('0' + id))}, nil, "", nil, nil)
			rd.RecordNavigation(nil, route, time.Millisecond, true)
			rd.RecordGuard("guard", GuardAllow, time.Millisecond)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rd.GetCurrentRoute()
			_ = rd.GetHistory()
			_ = rd.GetGuards()
			_ = rd.GetHistoryCount()
		}()
	}

	wg.Wait()

	// Verify no data corruption
	assert.NotNil(t, rd.GetCurrentRoute())
	assert.True(t, rd.GetHistoryCount() > 0)
}

// TestRouterDebugger_Render tests output rendering
func TestRouterDebugger_Render(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*RouterDebugger)
		contains []string
	}{
		{
			name: "empty state",
			setup: func(rd *RouterDebugger) {
				// No setup
			},
			contains: []string{"No current route", "No navigation history"},
		},
		{
			name: "with current route",
			setup: func(rd *RouterDebugger) {
				route := router.NewRoute("/user/:id", "user", map[string]string{"id": "123"}, map[string]string{"tab": "profile"}, "bio", nil, nil)
				rd.RecordNavigation(nil, route, 5*time.Millisecond, true)
			},
			contains: []string{"Current Route", "/user/:id", "user", "id: 123", "tab: profile", "#bio"},
		},
		{
			name: "with navigation history",
			setup: func(rd *RouterDebugger) {
				from := router.NewRoute("/home", "home", nil, nil, "", nil, nil)
				to := router.NewRoute("/about", "about", nil, nil, "", nil, nil)
				rd.RecordNavigation(from, to, 5*time.Millisecond, true)
			},
			contains: []string{"Navigation History", "/home", "/about", "ms", "✓"},
		},
		{
			name: "with failed navigation",
			setup: func(rd *RouterDebugger) {
				from := router.NewRoute("/home", "home", nil, nil, "", nil, nil)
				to := router.NewRoute("/admin", "admin", nil, nil, "", nil, nil)
				rd.RecordNavigation(from, to, 2*time.Millisecond, false)
			},
			contains: []string{"Navigation History", "/home", "/admin", "ms", "✗"},
		},
		{
			name: "with guard execution",
			setup: func(rd *RouterDebugger) {
				rd.RecordGuard("authGuard", GuardAllow, 1*time.Millisecond)
				rd.RecordGuard("permGuard", GuardCancel, 2*time.Millisecond)
				rd.RecordGuard("loginGuard", GuardRedirect, 3*time.Millisecond)
			},
			contains: []string{"Guard Execution", "authGuard", "Allow", "permGuard", "Cancel", "loginGuard", "Redirect"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := NewRouterDebugger(10)
			tt.setup(rd)

			output := rd.Render()

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestRouterDebugger_RenderFormatting tests output formatting details
func TestRouterDebugger_RenderFormatting(t *testing.T) {
	rd := NewRouterDebugger(10)

	// Add complex route with all fields
	route := router.NewRoute(
		"/user/:id/post/:postId",
		"user-post",
		map[string]string{"id": "123", "postId": "456"},
		map[string]string{"tab": "comments", "page": "2"},
		"section-1",
		nil,
		nil,
	)
	rd.RecordNavigation(nil, route, 10*time.Millisecond, true)

	output := rd.Render()

	// Verify all route components are displayed
	assert.Contains(t, output, "/user/:id/post/:postId")
	assert.Contains(t, output, "user-post")
	assert.Contains(t, output, "id: 123")
	assert.Contains(t, output, "postId: 456")
	assert.Contains(t, output, "tab: comments")
	assert.Contains(t, output, "page: 2")
	assert.Contains(t, output, "#section-1")

	// Verify sections are present
	assert.Contains(t, output, "Current Route")
	assert.Contains(t, output, "Navigation History")
}

// TestRouterDebugger_EmptyRender tests rendering with no data
func TestRouterDebugger_EmptyRender(t *testing.T) {
	rd := NewRouterDebugger(10)

	output := rd.Render()

	// Should have styled empty messages
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "No current route")
	assert.Contains(t, output, "No navigation history")
}

// TestRouterDebugger_MultipleNavigations tests multiple navigation records
func TestRouterDebugger_MultipleNavigations(t *testing.T) {
	rd := NewRouterDebugger(10)

	// Simulate navigation sequence
	routes := []*router.Route{
		router.NewRoute("/", "home", nil, nil, "", nil, nil),
		router.NewRoute("/about", "about", nil, nil, "", nil, nil),
		router.NewRoute("/contact", "contact", nil, nil, "", nil, nil),
	}

	for i := 1; i < len(routes); i++ {
		rd.RecordNavigation(routes[i-1], routes[i], time.Duration(i)*time.Millisecond, true)
	}

	history := rd.GetHistory()
	assert.Len(t, history, 2)

	// Verify order (chronological - oldest first)
	assert.Equal(t, "/", history[0].From.Path)
	assert.Equal(t, "/about", history[0].To.Path)
	assert.Equal(t, "/about", history[1].From.Path)
	assert.Equal(t, "/contact", history[1].To.Path)
}
