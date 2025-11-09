package devtools

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
)

// GuardResult represents the result of a navigation guard execution.
type GuardResult int

const (
	// GuardAllow indicates the guard allowed navigation to proceed.
	GuardAllow GuardResult = iota
	// GuardCancel indicates the guard canceled navigation.
	GuardCancel
	// GuardRedirect indicates the guard redirected to a different route.
	GuardRedirect
)

// String returns the string representation of a GuardResult.
func (gr GuardResult) String() string {
	switch gr {
	case GuardAllow:
		return "Allow"
	case GuardCancel:
		return "Cancel"
	case GuardRedirect:
		return "Redirect"
	default:
		return "Unknown"
	}
}

// RouteRecord captures a navigation event between routes.
//
// RouteRecord stores information about a single navigation event, including
// the source and destination routes, timing information, and success status.
// This is used by the RouterDebugger to track navigation history.
//
// Fields:
//   - From: The route being navigated away from (nil for initial navigation)
//   - To: The route being navigated to
//   - Timestamp: When the navigation occurred
//   - Duration: How long the navigation took
//   - Success: Whether the navigation completed successfully
//
// Example:
//
//	record := RouteRecord{
//		From:      previousRoute,
//		To:        newRoute,
//		Timestamp: time.Now(),
//		Duration:  5 * time.Millisecond,
//		Success:   true,
//	}
type RouteRecord struct {
	From      *router.Route // Source route (nil for initial navigation)
	To        *router.Route // Destination route
	Timestamp time.Time     // When navigation occurred
	Duration  time.Duration // Navigation duration
	Success   bool          // Whether navigation succeeded
}

// GuardExecution captures a navigation guard execution event.
//
// GuardExecution stores information about a single guard execution, including
// the guard name, result, and timing information. This is used by the
// RouterDebugger to trace guard execution during navigation.
//
// Fields:
//   - Name: The name of the guard function
//   - Result: The guard's decision (Allow, Cancel, or Redirect)
//   - Timestamp: When the guard executed
//   - Duration: How long the guard took to execute
//
// Example:
//
//	execution := GuardExecution{
//		Name:      "authGuard",
//		Result:    GuardAllow,
//		Timestamp: time.Now(),
//		Duration:  1 * time.Millisecond,
//	}
type GuardExecution struct {
	Name      string        // Guard function name
	Result    GuardResult   // Guard decision
	Timestamp time.Time     // When guard executed
	Duration  time.Duration // Guard execution duration
}

// RouterDebugger tracks and displays router navigation state and history.
//
// RouterDebugger provides debugging capabilities for the router system,
// including current route inspection, navigation history tracking, and
// guard execution tracing. It maintains a circular buffer of navigation
// records and guard executions for debugging purposes.
//
// Thread Safety:
// All methods are thread-safe and can be called concurrently from multiple
// goroutines. Internal state is protected by a sync.RWMutex.
//
// Usage:
//
//	debugger := NewRouterDebugger(100)
//
//	// Record navigation
//	debugger.RecordNavigation(fromRoute, toRoute, 5*time.Millisecond, true)
//
//	// Record guard execution
//	debugger.RecordGuard("authGuard", GuardAllow, 1*time.Millisecond)
//
//	// Display debug information
//	output := debugger.Render()
//	fmt.Println(output)
type RouterDebugger struct {
	currentRoute *router.Route    // Current active route
	history      []RouteRecord    // Navigation history (circular buffer)
	guards       []GuardExecution // Guard execution history (circular buffer)
	maxSize      int              // Maximum history size
	mu           sync.RWMutex     // Protects all fields
}

// NewRouterDebugger creates a new RouterDebugger with the specified history size.
//
// The debugger maintains circular buffers for navigation history and guard
// executions. When the buffers reach maxSize, the oldest entries are removed
// to make room for new ones.
//
// Parameters:
//   - maxSize: Maximum number of records to keep in history buffers
//
// Returns:
//   - *RouterDebugger: A new debugger instance
//
// Example:
//
//	debugger := NewRouterDebugger(100)
func NewRouterDebugger(maxSize int) *RouterDebugger {
	return &RouterDebugger{
		currentRoute: nil,
		history:      make([]RouteRecord, 0, maxSize),
		guards:       make([]GuardExecution, 0, maxSize),
		maxSize:      maxSize,
	}
}

// RecordNavigation records a navigation event.
//
// This method captures a navigation from one route to another, including
// timing and success information. The navigation is added to the history
// buffer, and the current route is updated to the destination route.
//
// Parameters:
//   - from: The source route (nil for initial navigation)
//   - to: The destination route
//   - duration: How long the navigation took
//   - success: Whether the navigation succeeded
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	debugger.RecordNavigation(homeRoute, aboutRoute, 5*time.Millisecond, true)
func (rd *RouterDebugger) RecordNavigation(from, to *router.Route, duration time.Duration, success bool) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	record := RouteRecord{
		From:      from,
		To:        to,
		Timestamp: time.Now(),
		Duration:  duration,
		Success:   success,
	}

	rd.history = append(rd.history, record)

	// Enforce max size (circular buffer)
	if len(rd.history) > rd.maxSize {
		rd.history = rd.history[len(rd.history)-rd.maxSize:]
	}

	// Update current route
	rd.currentRoute = to
}

// RecordGuard records a navigation guard execution.
//
// This method captures a guard execution event, including the guard name,
// result, and timing information. The execution is added to the guard
// history buffer.
//
// Parameters:
//   - guardName: The name of the guard function
//   - result: The guard's decision (Allow, Cancel, or Redirect)
//   - duration: How long the guard took to execute
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	debugger.RecordGuard("authGuard", GuardAllow, 1*time.Millisecond)
func (rd *RouterDebugger) RecordGuard(guardName string, result GuardResult, duration time.Duration) {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	execution := GuardExecution{
		Name:      guardName,
		Result:    result,
		Timestamp: time.Now(),
		Duration:  duration,
	}

	rd.guards = append(rd.guards, execution)

	// Enforce max size (circular buffer)
	if len(rd.guards) > rd.maxSize {
		rd.guards = rd.guards[len(rd.guards)-rd.maxSize:]
	}
}

// GetCurrentRoute returns the current active route.
//
// Returns:
//   - *router.Route: The current route, or nil if no route is active
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
func (rd *RouterDebugger) GetCurrentRoute() *router.Route {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	return rd.currentRoute
}

// GetHistory returns a copy of the navigation history.
//
// The returned slice is a defensive copy to prevent external modification
// of the internal state.
//
// Returns:
//   - []RouteRecord: Copy of navigation history (most recent last)
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
func (rd *RouterDebugger) GetHistory() []RouteRecord {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	// Return defensive copy
	history := make([]RouteRecord, len(rd.history))
	copy(history, rd.history)
	return history
}

// GetHistoryCount returns the number of navigation records in history.
//
// Returns:
//   - int: Number of navigation records
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
func (rd *RouterDebugger) GetHistoryCount() int {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	return len(rd.history)
}

// GetGuards returns a copy of the guard execution history.
//
// The returned slice is a defensive copy to prevent external modification
// of the internal state.
//
// Returns:
//   - []GuardExecution: Copy of guard execution history (most recent last)
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
func (rd *RouterDebugger) GetGuards() []GuardExecution {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	// Return defensive copy
	guards := make([]GuardExecution, len(rd.guards))
	copy(guards, rd.guards)
	return guards
}

// Clear removes all navigation history and guard executions.
//
// This method resets the debugger to its initial state, clearing the
// current route, navigation history, and guard execution history.
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
func (rd *RouterDebugger) Clear() {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	rd.currentRoute = nil
	rd.history = make([]RouteRecord, 0, rd.maxSize)
	rd.guards = make([]GuardExecution, 0, rd.maxSize)
}

// Render generates a styled text representation of the router debug state.
//
// The output includes:
//   - Current route information (path, name, params, query, hash)
//   - Navigation history with timestamps and success indicators
//   - Guard execution trace with results and timing
//
// Returns:
//   - string: Styled debug output using Lipgloss
//
// Thread Safety:
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	output := debugger.Render()
//	fmt.Println(output)
func (rd *RouterDebugger) Render() string {
	rd.mu.RLock()
	defer rd.mu.RUnlock()

	var sections []string

	// Current Route Section
	sections = append(sections, rd.renderCurrentRoute())

	// Navigation History Section
	sections = append(sections, rd.renderNavigationHistory())

	// Guard Execution Section
	if len(rd.guards) > 0 {
		sections = append(sections, rd.renderGuardExecution())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderCurrentRoute renders the current route section.
func (rd *RouterDebugger) renderCurrentRoute() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true)

	if rd.currentRoute == nil {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		return headerStyle.Render("Current Route") + "\n" +
			emptyStyle.Render("No current route")
	}

	lines := []string{
		headerStyle.Render("Current Route"),
		"",
	}

	// Path and name
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35")).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	lines = append(lines, fmt.Sprintf("Path: %s", pathStyle.Render(rd.currentRoute.Path)))
	if rd.currentRoute.Name != "" {
		lines = append(lines, fmt.Sprintf("Name: %s", nameStyle.Render(rd.currentRoute.Name)))
	}

	// Parameters
	if len(rd.currentRoute.Params) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Parameters:")
		for key, value := range rd.currentRoute.Params {
			lines = append(lines, fmt.Sprintf("  %s: %s", key, value))
		}
	}

	// Query
	if len(rd.currentRoute.Query) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Query:")
		for key, value := range rd.currentRoute.Query {
			lines = append(lines, fmt.Sprintf("  %s: %s", key, value))
		}
	}

	// Hash
	if rd.currentRoute.Hash != "" {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Hash: #%s", rd.currentRoute.Hash))
	}

	return strings.Join(lines, "\n")
}

// renderNavigationHistory renders the navigation history section.
func (rd *RouterDebugger) renderNavigationHistory() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true)

	lines := []string{
		"",
		headerStyle.Render("Navigation History"),
		"",
	}

	if len(rd.history) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		lines = append(lines, emptyStyle.Render("No navigation history"))
		return strings.Join(lines, "\n")
	}

	timestampStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	durationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229"))

	// Display in reverse order (most recent first)
	for i := len(rd.history) - 1; i >= 0; i-- {
		record := rd.history[i]

		timestamp := timestampStyle.Render(record.Timestamp.Format("15:04:05.000"))

		fromPath := "/"
		if record.From != nil {
			fromPath = record.From.Path
		}
		toPath := record.To.Path

		indicator := "✓"
		indicatorStyle := successStyle
		if !record.Success {
			indicator = "✗"
			indicatorStyle = failStyle
		}

		duration := durationStyle.Render(formatRouterDuration(record.Duration))

		line := fmt.Sprintf("[%s] %s %s → %s (%s)",
			timestamp,
			indicatorStyle.Render(indicator),
			fromPath,
			toPath,
			duration,
		)

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderGuardExecution renders the guard execution section.
func (rd *RouterDebugger) renderGuardExecution() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true)

	lines := []string{
		"",
		headerStyle.Render("Guard Execution"),
		"",
	}

	timestampStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	guardNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35")).Bold(true)
	allowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	cancelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	redirectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	durationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229"))

	// Display in reverse order (most recent first)
	for i := len(rd.guards) - 1; i >= 0; i-- {
		guard := rd.guards[i]

		timestamp := timestampStyle.Render(guard.Timestamp.Format("15:04:05.000"))
		guardName := guardNameStyle.Render(guard.Name)

		var resultStyle lipgloss.Style
		switch guard.Result {
		case GuardAllow:
			resultStyle = allowStyle
		case GuardCancel:
			resultStyle = cancelStyle
		case GuardRedirect:
			resultStyle = redirectStyle
		}

		result := resultStyle.Render(guard.Result.String())
		duration := durationStyle.Render(formatRouterDuration(guard.Duration))

		line := fmt.Sprintf("[%s] %s → %s (%s)",
			timestamp,
			guardName,
			result,
			duration,
		)

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatRouterDuration formats a duration for display in router debugger.
// Note: Similar to formatDuration in performance.go but kept separate to avoid coupling.
func formatRouterDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
