package router

import (
	"sync"
)

// History manages the navigation history stack with forward/back support.
//
// The history stack maintains a list of visited routes and tracks the current
// position within that list. It supports:
//   - Push: Add new entry and truncate forward history
//   - Replace: Update current entry without changing history length
//   - State preservation: Attach arbitrary state to history entries
//   - Max size enforcement: Limit history stack size (optional)
//   - Thread safety: All operations are protected by mutex
//
// Architecture:
//   - entries: Slice of history entries (routes + optional state)
//   - current: Index of current entry in the slice (-1 if empty)
//   - maxSize: Optional maximum number of entries (0 = unlimited)
//   - mu: Mutex for thread-safe concurrent access
//
// History Behavior:
//   - Push advances forward and truncates any forward history
//   - Replace updates current entry without affecting history length
//   - When maxSize is set, oldest entries are removed when limit exceeded
//
// Thread Safety:
// All public methods acquire the mutex and are safe for concurrent use
// across multiple goroutines.
//
// Usage:
//
//	history := &History{maxSize: 50}
//
//	// Push new route
//	history.Push(route)
//
//	// Replace current route
//	history.Replace(newRoute)
//
//	// Push with state
//	history.PushWithState(route, map[string]interface{}{"scrollPos": 100})
//
//	// Get current state
//	state := history.CurrentState()
//
// Example History Flow:
//
//	Initial: []
//	Push(/home): [/home] ← current
//	Push(/about): [/home, /about] ← current
//	Push(/contact): [/home, /about, /contact] ← current
//	Back: [/home, /about ← current, /contact]
//	Push(/faq): [/home, /about, /faq] ← current (contact removed)
type History struct {
	entries []*HistoryEntry // History stack entries
	current int             // Current position in history (-1 if empty)
	maxSize int             // Maximum history size (0 = unlimited)
	mu      sync.Mutex      // Protects all fields
}

// HistoryEntry represents a single entry in the navigation history.
//
// Each entry contains the route that was navigated to and optional
// state data that can be used to restore application state when
// navigating back to this entry.
//
// Fields:
//   - Route: The route for this history entry
//   - State: Optional arbitrary state data (can be nil)
//
// State Preservation:
// The State field can hold any data needed to restore the application
// state when navigating back to this entry. Common use cases:
//   - Scroll position
//   - Form data
//   - Filter settings
//   - Expanded/collapsed sections
//
// Example:
//
//	entry := &HistoryEntry{
//		Route: route,
//		State: map[string]interface{}{
//			"scrollPos": 100,
//			"filter": "active",
//		},
//	}
type HistoryEntry struct {
	Route *Route      // The route for this entry
	State interface{} // Optional state data
}

// Push adds a new route to the history stack.
//
// This operation:
//  1. Truncates any forward history (entries after current)
//  2. Appends the new route as a history entry
//  3. Advances current to the new entry
//  4. Enforces maxSize limit if set (removes oldest entries)
//
// Parameters:
//   - route: The route to add to history
//
// Thread Safety:
// This method acquires the mutex and is safe for concurrent use.
//
// Example:
//
//	history.Push(NewRoute("/home", "home", nil, nil, "", nil, nil))
//	history.Push(NewRoute("/about", "about", nil, nil, "", nil, nil))
//	// History: [/home, /about] ← current
//
// Forward History Truncation:
//
//	// Start: [/home, /about ← current, /contact]
//	history.Push(NewRoute("/faq", "faq", nil, nil, "", nil, nil))
//	// Result: [/home, /about, /faq] ← current (contact removed)
func (h *History) Push(route *Route) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Truncate forward history (remove entries after current)
	if h.current >= 0 && h.current < len(h.entries)-1 {
		h.entries = h.entries[:h.current+1]
	}

	// Append new entry
	entry := &HistoryEntry{Route: route}
	h.entries = append(h.entries, entry)
	h.current = len(h.entries) - 1

	// Enforce max size if set
	h.enforceMaxSize()
}

// Replace updates the current history entry without changing history length.
//
// This operation:
//  1. Replaces the current entry's route with the new route
//  2. Preserves the State field from the current entry
//  3. Does not change the current index
//  4. If history is empty, creates a new entry (same as Push)
//
// Parameters:
//   - route: The route to replace the current entry with
//
// Thread Safety:
// This method acquires the mutex and is safe for concurrent use.
//
// Use Cases:
//   - URL updates without creating history entries
//   - Query parameter changes
//   - Hash fragment updates
//   - Route redirects that shouldn't create back button entries
//
// Example:
//
//	// History: [/home, /about ← current, /contact]
//	history.Replace(NewRoute("/team", "team", nil, nil, "", nil, nil))
//	// Result: [/home, /team ← current, /contact] (about replaced with team)
func (h *History) Replace(route *Route) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// If history is empty, create first entry
	if len(h.entries) == 0 {
		h.entries = []*HistoryEntry{{Route: route}}
		h.current = 0
		return
	}

	// Replace current entry's route, preserve state
	h.entries[h.current].Route = route
}

// PushWithState adds a new route to history with associated state.
//
// This is identical to Push() but also attaches state data to the
// history entry. The state can be retrieved later with CurrentState()
// when navigating back to this entry.
//
// Parameters:
//   - route: The route to add to history
//   - state: Arbitrary state data to preserve (can be nil)
//
// Thread Safety:
// This method acquires the mutex and is safe for concurrent use.
//
// Example:
//
//	state := map[string]interface{}{
//		"scrollPos": 100,
//		"filter": "active",
//		"selectedId": "123",
//	}
//	history.PushWithState(route, state)
//
//	// Later, when navigating back:
//	restoredState := history.CurrentState()
//	if scrollData, ok := restoredState.(map[string]interface{}); ok {
//		scrollPos := scrollData["scrollPos"].(int)
//		// Restore scroll position
//	}
func (h *History) PushWithState(route *Route, state interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Truncate forward history
	if h.current >= 0 && h.current < len(h.entries)-1 {
		h.entries = h.entries[:h.current+1]
	}

	// Append new entry with state
	entry := &HistoryEntry{
		Route: route,
		State: state,
	}
	h.entries = append(h.entries, entry)
	h.current = len(h.entries) - 1

	// Enforce max size if set
	h.enforceMaxSize()
}

// CurrentState returns the state associated with the current history entry.
//
// Returns:
//   - interface{}: The state data, or nil if no state or empty history
//
// Thread Safety:
// This method acquires the mutex and is safe for concurrent use.
//
// Type Assertions:
// The returned state is interface{} and requires type assertion:
//
//	state := history.CurrentState()
//	if scrollData, ok := state.(map[string]interface{}); ok {
//		scrollPos := scrollData["scrollPos"].(int)
//	}
//
// Example:
//
//	// After PushWithState with scroll position
//	state := history.CurrentState()
//	if state != nil {
//		scrollPos := state.(map[string]interface{})["scrollPos"].(int)
//		fmt.Printf("Scroll position: %d\n", scrollPos)
//	}
func (h *History) CurrentState() interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 || h.current < 0 {
		return nil
	}

	return h.entries[h.current].State
}

// enforceMaxSize removes oldest entries if history exceeds maxSize.
//
// This is an internal helper method called after Push operations.
// It ensures the history stack doesn't grow beyond the configured limit.
//
// Behavior:
//   - If maxSize is 0 or negative, no limit is enforced
//   - If history exceeds maxSize, oldest entries are removed
//   - Current index is adjusted to maintain correct position
//
// Thread Safety:
// This method is called while the mutex is already held by Push/PushWithState.
// It must NOT acquire the mutex again (would cause deadlock).
//
// Example:
//
//	// maxSize = 3
//	// Before: [/1, /2, /3, /4, /5] ← current
//	// After:  [/3, /4, /5] ← current (oldest 2 removed)
func (h *History) enforceMaxSize() {
	if h.maxSize <= 0 {
		return // No limit
	}

	if len(h.entries) > h.maxSize {
		// Remove oldest entries
		excess := len(h.entries) - h.maxSize
		h.entries = h.entries[excess:]
		h.current -= excess
	}
}
