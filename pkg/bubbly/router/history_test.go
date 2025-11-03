package router

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHistory_Push tests the Push operation for adding history entries.
func TestHistory_Push(t *testing.T) {
	tests := []struct {
		name           string
		initialEntries []*HistoryEntry
		initialCurrent int
		pushRoute      *Route
		wantLength     int
		wantCurrent    int
		wantPath       string
	}{
		{
			name:           "push to empty history",
			initialEntries: nil,
			initialCurrent: -1,
			pushRoute:      NewRoute("/home", "home", nil, nil, "", nil, nil),
			wantLength:     1,
			wantCurrent:    0,
			wantPath:       "/home",
		},
		{
			name: "push to history with one entry",
			initialEntries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
			},
			initialCurrent: 0,
			pushRoute:      NewRoute("/about", "about", nil, nil, "", nil, nil),
			wantLength:     2,
			wantCurrent:    1,
			wantPath:       "/about",
		},
		{
			name: "push truncates forward history",
			initialEntries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
				{Route: NewRoute("/about", "about", nil, nil, "", nil, nil)},
				{Route: NewRoute("/contact", "contact", nil, nil, "", nil, nil)},
			},
			initialCurrent: 1, // At /about
			pushRoute:      NewRoute("/faq", "faq", nil, nil, "", nil, nil),
			wantLength:     3, // /home, /about, /faq (contact removed)
			wantCurrent:    2,
			wantPath:       "/faq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				entries: tt.initialEntries,
				current: tt.initialCurrent,
			}

			h.Push(tt.pushRoute)

			assert.Equal(t, tt.wantLength, len(h.entries), "history length mismatch")
			assert.Equal(t, tt.wantCurrent, h.current, "current index mismatch")
			assert.Equal(t, tt.wantPath, h.entries[h.current].Route.Path, "current route path mismatch")
		})
	}
}

// TestHistory_Replace tests the Replace operation for updating current entry.
func TestHistory_Replace(t *testing.T) {
	tests := []struct {
		name           string
		initialEntries []*HistoryEntry
		initialCurrent int
		replaceRoute   *Route
		wantLength     int
		wantCurrent    int
		wantPath       string
	}{
		{
			name:           "replace in empty history creates entry",
			initialEntries: nil,
			initialCurrent: -1,
			replaceRoute:   NewRoute("/home", "home", nil, nil, "", nil, nil),
			wantLength:     1,
			wantCurrent:    0,
			wantPath:       "/home",
		},
		{
			name: "replace current entry",
			initialEntries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
			},
			initialCurrent: 0,
			replaceRoute:   NewRoute("/dashboard", "dashboard", nil, nil, "", nil, nil),
			wantLength:     1,
			wantCurrent:    0,
			wantPath:       "/dashboard",
		},
		{
			name: "replace middle entry preserves others",
			initialEntries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
				{Route: NewRoute("/about", "about", nil, nil, "", nil, nil)},
				{Route: NewRoute("/contact", "contact", nil, nil, "", nil, nil)},
			},
			initialCurrent: 1, // At /about
			replaceRoute:   NewRoute("/team", "team", nil, nil, "", nil, nil),
			wantLength:     3,
			wantCurrent:    1,
			wantPath:       "/team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				entries: tt.initialEntries,
				current: tt.initialCurrent,
			}

			h.Replace(tt.replaceRoute)

			assert.Equal(t, tt.wantLength, len(h.entries), "history length mismatch")
			assert.Equal(t, tt.wantCurrent, h.current, "current index mismatch")
			assert.Equal(t, tt.wantPath, h.entries[h.current].Route.Path, "current route path mismatch")
		})
	}
}

// TestHistory_MaxSize tests that history respects max size limit.
func TestHistory_MaxSize(t *testing.T) {
	h := &History{
		maxSize: 3,
	}

	// Push 5 routes
	routes := []*Route{
		NewRoute("/1", "r1", nil, nil, "", nil, nil),
		NewRoute("/2", "r2", nil, nil, "", nil, nil),
		NewRoute("/3", "r3", nil, nil, "", nil, nil),
		NewRoute("/4", "r4", nil, nil, "", nil, nil),
		NewRoute("/5", "r5", nil, nil, "", nil, nil),
	}

	for _, route := range routes {
		h.Push(route)
	}

	// Should only keep last 3
	assert.Equal(t, 3, len(h.entries), "history should respect max size")
	assert.Equal(t, 2, h.current, "current should be at end")
	assert.Equal(t, "/3", h.entries[0].Route.Path, "oldest entry should be /3")
	assert.Equal(t, "/5", h.entries[2].Route.Path, "newest entry should be /5")
}

// TestHistory_ThreadSafety tests concurrent access to history.
func TestHistory_ThreadSafety(t *testing.T) {
	h := &History{}

	var wg sync.WaitGroup
	concurrency := 10

	// Concurrent pushes
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(n int) {
			defer wg.Done()
			route := NewRoute("/route", "route", nil, nil, "", nil, nil)
			h.Push(route)
		}(i)
	}

	wg.Wait()

	// Should have all entries
	assert.Equal(t, concurrency, len(h.entries), "all concurrent pushes should succeed")
}

// TestHistory_PushWithState tests pushing with state preservation.
func TestHistory_PushWithState(t *testing.T) {
	h := &History{}

	route := NewRoute("/home", "home", nil, nil, "", nil, nil)
	state := map[string]interface{}{"scrollPos": 100}

	h.PushWithState(route, state)

	assert.Equal(t, 1, len(h.entries), "should have one entry")
	assert.Equal(t, state, h.entries[0].State, "state should be preserved")
}

// TestHistory_CurrentState tests retrieving current state.
func TestHistory_CurrentState(t *testing.T) {
	tests := []struct {
		name      string
		entries   []*HistoryEntry
		current   int
		wantState interface{}
	}{
		{
			name:      "empty history returns nil",
			entries:   nil,
			current:   -1,
			wantState: nil,
		},
		{
			name: "returns current state",
			entries: []*HistoryEntry{
				{
					Route: NewRoute("/home", "home", nil, nil, "", nil, nil),
					State: map[string]interface{}{"scrollPos": 100},
				},
			},
			current:   0,
			wantState: map[string]interface{}{"scrollPos": 100},
		},
		{
			name: "returns nil for entry without state",
			entries: []*HistoryEntry{
				{Route: NewRoute("/home", "home", nil, nil, "", nil, nil)},
			},
			current:   0,
			wantState: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				entries: tt.entries,
				current: tt.current,
			}

			state := h.CurrentState()
			assert.Equal(t, tt.wantState, state)
		})
	}
}
