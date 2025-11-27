package devtools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// EventTracker captures and displays events in real-time.
//
// It provides event capture with pause/resume functionality, filtering,
// and real-time display with Lipgloss styling. The tracker integrates
// with the EventLog from the Store.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	tracker := NewEventTracker(5000)
//	tracker.CaptureEvent(EventRecord{
//	    ID:        "event-1",
//	    Name:      "click",
//	    SourceID:  "button-1",
//	    Timestamp: time.Now(),
//	})
//	output := tracker.Render()
type EventTracker struct {
	// events is the event log
	events *EventLog

	// filter is the current filter string (case-insensitive)
	filter string

	// paused indicates if event capture is paused
	paused bool

	// maxEvents is the maximum number of events to keep
	maxEvents int

	// mu protects concurrent access to filter and paused
	mu sync.RWMutex
}

// EventStatistics provides statistics about captured events.
type EventStatistics struct {
	// TotalEvents is the total number of events captured
	TotalEvents int

	// EventsByName maps event names to their count
	EventsByName map[string]int

	// EventsBySource maps source IDs to their count
	EventsBySource map[string]int
}

// NewEventTracker creates a new event tracker with the specified maximum size.
//
// The tracker starts unpaused with no filter applied.
//
// Example:
//
//	tracker := NewEventTracker(5000) // Keep last 5000 events
//
// Parameters:
//   - maxEvents: Maximum number of events to keep
//
// Returns:
//   - *EventTracker: A new event tracker instance
func NewEventTracker(maxEvents int) *EventTracker {
	return &EventTracker{
		events:    NewEventLog(maxEvents),
		filter:    "",
		paused:    false,
		maxEvents: maxEvents,
	}
}

// CaptureEvent captures an event if not paused.
//
// If the tracker is paused, the event is ignored. Otherwise, it is
// added to the event log.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tracker.CaptureEvent(EventRecord{
//	    ID:        "event-1",
//	    Name:      "click",
//	    SourceID:  "button-1",
//	    Timestamp: time.Now(),
//	    Duration:  time.Millisecond,
//	})
//
// Parameters:
//   - event: The event to capture
func (et *EventTracker) CaptureEvent(event EventRecord) {
	et.mu.RLock()
	paused := et.paused
	et.mu.RUnlock()

	if paused {
		return
	}

	et.events.Append(event)
}

// Pause pauses event capture.
//
// While paused, calls to CaptureEvent will be ignored.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (et *EventTracker) Pause() {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.paused = true
}

// Resume resumes event capture.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (et *EventTracker) Resume() {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.paused = false
}

// IsPaused returns whether event capture is paused.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - bool: True if paused, false otherwise
func (et *EventTracker) IsPaused() bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.paused
}

// GetEventCount returns the number of captured events.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int: Number of events
func (et *EventTracker) GetEventCount() int {
	return et.events.Len()
}

// GetRecent returns the N most recent events.
//
// If n is greater than the number of events, all events are returned.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - n: Number of recent events to return
//
// Returns:
//   - []EventRecord: The N most recent events
func (et *EventTracker) GetRecent(n int) []EventRecord {
	return et.events.GetRecent(n)
}

// Clear removes all captured events.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (et *EventTracker) Clear() {
	et.events.Clear()
}

// SetFilter sets the filter string for event names.
//
// The filter is case-insensitive and matches substrings.
// An empty string clears the filter.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tracker.SetFilter("click") // Show only events with "click" in name
//	tracker.SetFilter("")      // Show all events
//
// Parameters:
//   - filter: The filter string
func (et *EventTracker) SetFilter(filter string) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.filter = filter
}

// GetFilter returns the current filter string.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - string: The current filter
func (et *EventTracker) GetFilter() string {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return et.filter
}

// GetStatistics returns statistics about captured events.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - EventStatistics: Event statistics
func (et *EventTracker) GetStatistics() EventStatistics {
	events := et.events.GetRecent(et.maxEvents)

	stats := EventStatistics{
		TotalEvents:    len(events),
		EventsByName:   make(map[string]int),
		EventsBySource: make(map[string]int),
	}

	for _, event := range events {
		stats.EventsByName[event.Name]++
		stats.EventsBySource[event.SourceID]++
	}

	return stats
}

// Render generates the visual output of the event tracker.
//
// The output includes:
//   - A header with the title "Recent Events"
//   - Event list in reverse chronological order (newest first)
//   - Each event shows: timestamp, name, source → target, duration
//   - Filtered events based on the current filter
//   - Styled with Lipgloss for terminal display
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - string: The rendered output
func (et *EventTracker) Render() string {
	et.mu.RLock()
	filter := et.filter
	et.mu.RUnlock()

	// Get recent events (show last 50)
	events := et.events.GetRecent(50)

	if len(events) == 0 {
		return et.renderEmpty()
	}

	// Build output
	var lines []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")). // Purple
		Padding(0, 1)
	lines = append(lines, headerStyle.Render("Recent Events:"))
	lines = append(lines, "")

	// Render events in reverse order (newest first)
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]

		// Apply filter
		if filter != "" && !et.matchesFilter(event.Name, filter) {
			continue
		}

		// Format timestamp
		timeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Dark grey
		timeStr := timeStyle.Render(fmt.Sprintf("[%s]", event.Timestamp.Format("15:04:05.000")))

		// Event name
		nameStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("35")) // Green
		nameStr := nameStyle.Render(event.Name)

		// Source and target
		sourceStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")) // Purple
		sourceStr := sourceStyle.Render(event.SourceID)

		targetStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // Orange
		targetStr := ""
		if event.TargetID != "" {
			targetStr = " → " + targetStyle.Render(event.TargetID)
		}

		// Duration
		durationStr := ""
		if event.Duration > 0 {
			durationStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")) // Yellow
			durationStr = " " + durationStyle.Render(fmt.Sprintf("(%s)", event.Duration))
		}

		line := fmt.Sprintf("%s %s from %s%s%s",
			timeStr,
			nameStr,
			sourceStr,
			targetStr,
			durationStr)

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderEmpty renders the empty state message.
func (et *EventTracker) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Dark grey
		Italic(true).
		Padding(1, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")). // Purple
		Padding(0, 1)

	return headerStyle.Render("Recent Events:") + "\n\n" +
		emptyStyle.Render("No events captured")
}

// matchesFilter checks if an event name matches the filter (case-insensitive).
func (et *EventTracker) matchesFilter(name, filter string) bool {
	return strings.Contains(strings.ToLower(name), strings.ToLower(filter))
}
