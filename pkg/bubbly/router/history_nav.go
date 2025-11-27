package router

import (
	tea "github.com/charmbracelet/bubbletea"
)

// CanGoBack returns true if there are entries before the current position.
//
// This helper method checks if backward navigation is possible in the
// history stack. It's useful for UI elements like back buttons to
// determine if they should be enabled.
//
// Returns:
//   - bool: true if current > 0, false otherwise
//
// Thread Safety:
// This method acquires the mutex and is safe for concurrent use.
//
// Example:
//
//	if router.history.CanGoBack() {
//		// Show enabled back button
//		backButton.SetEnabled(true)
//	} else {
//		// Show disabled back button
//		backButton.SetEnabled(false)
//	}
func (h *History) CanGoBack() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.current > 0
}

// CanGoForward returns true if there are entries after the current position.
//
// This helper method checks if forward navigation is possible in the
// history stack. It's useful for UI elements like forward buttons to
// determine if they should be enabled.
//
// Returns:
//   - bool: true if current < len(entries)-1, false otherwise
//
// Thread Safety:
// This method acquires the mutex and is safe for concurrent use.
//
// Example:
//
//	if router.history.CanGoForward() {
//		// Show enabled forward button
//		forwardButton.SetEnabled(true)
//	} else {
//		// Show disabled forward button
//		forwardButton.SetEnabled(false)
//	}
func (h *History) CanGoForward() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return len(h.entries) > 0 && h.current < len(h.entries)-1
}

// navigateHistory is a helper function that performs history navigation.
// It updates the history index by delta and returns a RouteChangedMsg.
func (r *Router) navigateHistory(delta int) tea.Cmd {
	return func() tea.Msg {
		r.mu.Lock()
		defer r.mu.Unlock()

		// Save current route for "from" in message
		from := r.currentRoute

		// Move in history by delta
		r.history.mu.Lock()
		r.history.current += delta
		newRoute := r.history.entries[r.history.current].Route
		r.history.mu.Unlock()

		// Update current route
		r.currentRoute = newRoute

		return RouteChangedMsg{
			To:   newRoute,
			From: from,
		}
	}
}

// Back navigates to the previous entry in history.
//
// Back generates a Bubbletea command that moves to the previous history
// entry and updates the current route. If there is no previous entry
// (already at the first entry or empty history), it returns nil (no-op).
//
// Returns:
//   - tea.Cmd: Command that navigates back, or nil if cannot go back
//
// Navigation Flow:
//  1. Check if can go back (current > 0)
//  2. If not, return nil (no-op)
//  3. Move current index backward
//  4. Get route from history entry
//  5. Update current route
//  6. Return RouteChangedMsg
//
// Thread Safety:
// The command executes asynchronously but updates router state with
// proper locking. Multiple Back() calls are serialized by Bubbletea.
//
// Example:
//
//	// In component event handler
//	ctx.On("backButton", func(data interface{}) {
//		cmd := router.Back()
//		if cmd != nil {
//			// Navigation will happen
//		}
//	})
//
//	// In Update() method
//	case tea.KeyMsg:
//		if msg.String() == "esc" {
//			return m, router.Back()
//		}
//
// Use Cases:
//   - Back button in navigation bar
//   - Keyboard shortcut (ESC, Backspace)
//   - Cancel/undo navigation
//   - Return to previous screen
func (r *Router) Back() tea.Cmd {
	// Check if we can go back
	if !r.history.CanGoBack() {
		return nil // No-op
	}

	return r.navigateHistory(-1)
}

// Forward navigates to the next entry in history.
//
// Forward generates a Bubbletea command that moves to the next history
// entry and updates the current route. If there is no next entry
// (already at the last entry or empty history), it returns nil (no-op).
//
// Returns:
//   - tea.Cmd: Command that navigates forward, or nil if cannot go forward
//
// Navigation Flow:
//  1. Check if can go forward (current < len-1)
//  2. If not, return nil (no-op)
//  3. Move current index forward
//  4. Get route from history entry
//  5. Update current route
//  6. Return RouteChangedMsg
//
// Thread Safety:
// The command executes asynchronously but updates router state with
// proper locking. Multiple Forward() calls are serialized by Bubbletea.
//
// Example:
//
//	// In component event handler
//	ctx.On("forwardButton", func(data interface{}) {
//		cmd := router.Forward()
//		if cmd != nil {
//			// Navigation will happen
//		}
//	})
//
//	// In Update() method
//	case tea.KeyMsg:
//		if msg.String() == "ctrl+]" {
//			return m, router.Forward()
//		}
//
// Use Cases:
//   - Forward button in navigation bar
//   - Keyboard shortcut (Ctrl+], Alt+Right)
//   - Redo navigation
//   - Return to next screen after going back
func (r *Router) Forward() tea.Cmd {
	// Check if we can go forward
	if !r.history.CanGoForward() {
		return nil // No-op
	}

	return r.navigateHistory(1)
}

// Go navigates n steps in history (negative for back, positive for forward).
//
// Go generates a Bubbletea command that moves n steps in the history stack.
// Negative values move backward, positive values move forward. If n is 0,
// it returns nil (no-op). If n would move beyond the bounds, it clamps to
// the first or last entry.
//
// Parameters:
//   - n: Number of steps to move (negative=back, positive=forward, 0=no-op)
//
// Returns:
//   - tea.Cmd: Command that navigates n steps, or nil if n==0
//
// Navigation Flow:
//  1. Check if n == 0 (no-op)
//  2. Calculate new index (current + n)
//  3. Clamp to bounds [0, len-1]
//  4. If same as current, return nil (no-op)
//  5. Update current index
//  6. Get route from history entry
//  7. Update current route
//  8. Return RouteChangedMsg
//
// Bounds Clamping:
//   - new_index = max(0, min(current + n, len-1))
//   - Ensures index stays within valid range
//   - Prevents index out of bounds errors
//
// Thread Safety:
// The command executes asynchronously but updates router state with
// proper locking. Multiple Go() calls are serialized by Bubbletea.
//
// Example:
//
//	// Go back 2 steps
//	cmd := router.Go(-2)
//
//	// Go forward 3 steps
//	cmd := router.Go(3)
//
//	// Go to specific position (if you know the offset)
//	offset := targetIndex - currentIndex
//	cmd := router.Go(offset)
//
// Use Cases:
//   - Keyboard shortcuts with step count
//   - History navigation with specific offsets
//   - Jump to specific history position
//   - Batch navigation (skip multiple entries)
func (r *Router) Go(n int) tea.Cmd {
	// No-op if n is 0
	if n == 0 {
		return nil
	}

	// Check if history is empty
	r.history.mu.Lock()
	if len(r.history.entries) == 0 {
		r.history.mu.Unlock()
		return nil
	}

	// Calculate new index with bounds clamping
	newIndex := r.history.current + n
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(r.history.entries) {
		newIndex = len(r.history.entries) - 1
	}

	// No-op if we're already at the target index
	if newIndex == r.history.current {
		r.history.mu.Unlock()
		return nil
	}
	r.history.mu.Unlock()

	return func() tea.Msg {
		r.mu.Lock()
		defer r.mu.Unlock()

		// Save current route for "from" in message
		from := r.currentRoute

		// Move to new position in history
		r.history.mu.Lock()
		r.history.current = newIndex
		newRoute := r.history.entries[r.history.current].Route
		r.history.mu.Unlock()

		// Update current route
		r.currentRoute = newRoute

		return RouteChangedMsg{
			To:   newRoute,
			From: from,
		}
	}
}
