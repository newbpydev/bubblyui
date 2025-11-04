package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHistory_CanGoBack tests the CanGoBack helper method.
func TestHistory_CanGoBack(t *testing.T) {
	tests := []struct {
		name    string
		entries []*HistoryEntry
		current int
		want    bool
	}{
		{
			name:    "empty history cannot go back",
			entries: nil,
			current: -1,
			want:    false,
		},
		{
			name: "first entry cannot go back",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
			},
			current: 0,
			want:    false,
		},
		{
			name: "second entry can go back",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
				{Route: NewRoute("/about", "about", nil, nil, "", nil, nil)},
			},
			current: 1,
			want:    true,
		},
		{
			name: "middle entry can go back",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
				{Route: NewRoute("/about", "about", nil, nil, "", nil, nil)},
				{Route: NewRoute("/contact", "contact", nil, nil, "", nil, nil)},
			},
			current: 1,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				entries: tt.entries,
				current: tt.current,
			}

			got := h.CanGoBack()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestHistory_CanGoForward tests the CanGoForward helper method.
func TestHistory_CanGoForward(t *testing.T) {
	tests := []struct {
		name    string
		entries []*HistoryEntry
		current int
		want    bool
	}{
		{
			name:    "empty history cannot go forward",
			entries: nil,
			current: -1,
			want:    false,
		},
		{
			name: "last entry cannot go forward",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
			},
			current: 0,
			want:    false,
		},
		{
			name: "first entry can go forward",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
				{Route: NewRoute("/about", "about", nil, nil, "", nil, nil)},
			},
			current: 0,
			want:    true,
		},
		{
			name: "middle entry can go forward",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
				{Route: NewRoute("/about", "about", nil, nil, "", nil, nil)},
				{Route: NewRoute("/contact", "contact", nil, nil, "", nil, nil)},
			},
			current: 1,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				entries: tt.entries,
				current: tt.current,
			}

			got := h.CanGoForward()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRouter_Back tests the Back navigation method.
func TestRouter_Back(t *testing.T) {
	tests := []struct {
		name         string
		setupHistory func(*Router)
		wantPath     string
		wantMsgType  string
		wantNoOp     bool
	}{
		{
			name: "back from second entry",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[1].Route
			},
			wantPath:    "/home",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "back from third entry",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/contact", "contact", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[2].Route
			},
			wantPath:    "/about",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "back on first entry is no-op",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[0].Route
			},
			wantNoOp: true,
		},
		{
			name: "back on empty history is no-op",
			setupHistory: func(r *Router) {
				// No history
			},
			wantNoOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			tt.setupHistory(router)

			cmd := router.Back()
			if tt.wantNoOp {
				assert.Nil(t, cmd, "expected nil command for no-op")
				return
			}

			assert.NotNil(t, cmd, "expected non-nil command")

			// Execute command
			msg := cmd()

			// Check message type
			switch msg := msg.(type) {
			case RouteChangedMsg:
				assert.Equal(t, "RouteChangedMsg", tt.wantMsgType)
				assert.Equal(t, tt.wantPath, msg.To.Path)
			default:
				t.Errorf("unexpected message type: %T", msg)
			}
		})
	}
}

// TestRouter_Forward tests the Forward navigation method.
func TestRouter_Forward(t *testing.T) {
	tests := []struct {
		name         string
		setupHistory func(*Router)
		wantPath     string
		wantMsgType  string
		wantNoOp     bool
	}{
		{
			name: "forward from first entry",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.current = 0 // Move back to first
				r.currentRoute = r.history.entries[0].Route
			},
			wantPath:    "/about",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "forward from middle entry",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/contact", "contact", nil, nil, "", nil, nil))
				r.history.current = 1 // At /about
				r.currentRoute = r.history.entries[1].Route
			},
			wantPath:    "/contact",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "forward on last entry is no-op",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[1].Route
			},
			wantNoOp: true,
		},
		{
			name: "forward on empty history is no-op",
			setupHistory: func(r *Router) {
				// No history
			},
			wantNoOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			tt.setupHistory(router)

			cmd := router.Forward()
			if tt.wantNoOp {
				assert.Nil(t, cmd, "expected nil command for no-op")
				return
			}

			assert.NotNil(t, cmd, "expected non-nil command")

			// Execute command
			msg := cmd()

			// Check message type
			switch msg := msg.(type) {
			case RouteChangedMsg:
				assert.Equal(t, "RouteChangedMsg", tt.wantMsgType)
				assert.Equal(t, tt.wantPath, msg.To.Path)
			default:
				t.Errorf("unexpected message type: %T", msg)
			}
		})
	}
}

// TestRouter_Go tests the Go(n) navigation method.
func TestRouter_Go(t *testing.T) {
	tests := []struct {
		name         string
		setupHistory func(*Router)
		steps        int
		wantPath     string
		wantMsgType  string
		wantNoOp     bool
	}{
		{
			name: "go back 1 step",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[1].Route
			},
			steps:       -1,
			wantPath:    "/home",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "go back 2 steps",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/contact", "contact", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[2].Route
			},
			steps:       -2,
			wantPath:    "/home",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "go forward 1 step",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.current = 0 // Move back
				r.currentRoute = r.history.entries[0].Route
			},
			steps:       1,
			wantPath:    "/about",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "go forward 2 steps",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/contact", "contact", nil, nil, "", nil, nil))
				r.history.current = 0 // At first
				r.currentRoute = r.history.entries[0].Route
			},
			steps:       2,
			wantPath:    "/contact",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "go 0 steps is no-op",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[0].Route
			},
			steps:    0,
			wantNoOp: true,
		},
		{
			name: "go beyond bounds is clamped to first",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.currentRoute = r.history.entries[1].Route
			},
			steps:       -5,
			wantPath:    "/home",
			wantMsgType: "RouteChangedMsg",
		},
		{
			name: "go beyond bounds is clamped to last",
			setupHistory: func(r *Router) {
				r.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
				r.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
				r.history.current = 0
				r.currentRoute = r.history.entries[0].Route
			},
			steps:       5,
			wantPath:    "/about",
			wantMsgType: "RouteChangedMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			tt.setupHistory(router)

			cmd := router.Go(tt.steps)
			if tt.wantNoOp {
				assert.Nil(t, cmd, "expected nil command for no-op")
				return
			}

			assert.NotNil(t, cmd, "expected non-nil command")

			// Execute command
			msg := cmd()

			// Check message type
			switch msg := msg.(type) {
			case RouteChangedMsg:
				assert.Equal(t, "RouteChangedMsg", tt.wantMsgType)
				assert.Equal(t, tt.wantPath, msg.To.Path)
			default:
				t.Errorf("unexpected message type: %T", msg)
			}
		})
	}
}

// TestRouter_BackForward_Integration tests back/forward navigation flow.
func TestRouter_BackForward_Integration(t *testing.T) {
	router := NewRouter()

	// Build history: /home -> /about -> /contact
	router.history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
	router.history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
	router.history.Push(NewRoute("/contact", "contact", nil, nil, "", nil, nil))
	router.currentRoute = router.history.entries[2].Route

	// Back to /about
	cmd := router.Back()
	assert.NotNil(t, cmd)
	msg := cmd().(RouteChangedMsg)
	assert.Equal(t, "/about", msg.To.Path)
	assert.Equal(t, "/contact", msg.From.Path)

	// Back to /home
	cmd = router.Back()
	assert.NotNil(t, cmd)
	msg = cmd().(RouteChangedMsg)
	assert.Equal(t, "/home", msg.To.Path)
	assert.Equal(t, "/about", msg.From.Path)

	// Can't go back further
	cmd = router.Back()
	assert.Nil(t, cmd)

	// Forward to /about
	cmd = router.Forward()
	assert.NotNil(t, cmd)
	msg = cmd().(RouteChangedMsg)
	assert.Equal(t, "/about", msg.To.Path)
	assert.Equal(t, "/home", msg.From.Path)

	// Forward to /contact
	cmd = router.Forward()
	assert.NotNil(t, cmd)
	msg = cmd().(RouteChangedMsg)
	assert.Equal(t, "/contact", msg.To.Path)
	assert.Equal(t, "/about", msg.From.Path)

	// Can't go forward further
	cmd = router.Forward()
	assert.Nil(t, cmd)
}

// TestRouter_Go_BoundsChecking tests Go(n) with boundary conditions.
func TestRouter_Go_BoundsChecking(t *testing.T) {
	router := NewRouter()

	// Build history
	router.history.Push(NewRoute("/1", "r1", nil, nil, "", nil, nil))
	router.history.Push(NewRoute("/2", "r2", nil, nil, "", nil, nil))
	router.history.Push(NewRoute("/3", "r3", nil, nil, "", nil, nil))
	router.history.Push(NewRoute("/4", "r4", nil, nil, "", nil, nil))
	router.history.Push(NewRoute("/5", "r5", nil, nil, "", nil, nil))
	router.history.current = 2 // At /3
	router.currentRoute = router.history.entries[2].Route

	// Go back beyond start (should clamp to 0)
	cmd := router.Go(-10)
	assert.NotNil(t, cmd)
	msg := cmd().(RouteChangedMsg)
	assert.Equal(t, "/1", msg.To.Path)

	// Reset to middle
	router.history.current = 2
	router.currentRoute = router.history.entries[2].Route

	// Go forward beyond end (should clamp to last)
	cmd = router.Go(10)
	assert.NotNil(t, cmd)
	msg = cmd().(RouteChangedMsg)
	assert.Equal(t, "/5", msg.To.Path)
}
