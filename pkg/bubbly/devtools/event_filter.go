package devtools

import (
	"strings"
	"sync"
	"time"
)

// EventFilter provides filtering capabilities for events.
//
// It supports filtering by event names, source IDs, and time ranges.
// Multiple filter criteria can be combined, and all must match for an
// event to pass the filter.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	filter := NewEventFilter().
//	    WithNames("click", "submit").
//	    WithSources("button-1").
//	    WithTimeRange(startTime, endTime)
//
//	if filter.Matches(event) {
//	    // Event passes filter
//	}
//
//	filtered := filter.Apply(events)
type EventFilter struct {
	// names is the list of event names to filter by
	names []string

	// sources is the list of source IDs to filter by
	sources []string

	// timeRange is the time range to filter by
	timeRange *TimeRange

	// mu protects concurrent access to filter criteria
	mu sync.RWMutex
}

// TimeRange represents a time range for filtering events.
type TimeRange struct {
	// Start is the start time (inclusive)
	Start time.Time

	// End is the end time (inclusive)
	End time.Time
}

// NewEventFilter creates a new event filter with no criteria.
//
// The filter starts empty and matches all events until criteria are added.
//
// Example:
//
//	filter := NewEventFilter()
//	filter.WithNames("click", "submit")
//
// Returns:
//   - *EventFilter: A new event filter instance
func NewEventFilter() *EventFilter {
	return &EventFilter{
		names:     []string{},
		sources:   []string{},
		timeRange: nil,
	}
}

// WithNames sets the event names to filter by.
//
// If multiple names are provided, an event matches if its name matches
// any of the provided names (OR logic). Matching is case-insensitive
// and supports substring matching.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	filter.WithNames("click", "submit", "change")
//
// Parameters:
//   - names: Event names to filter by
//
// Returns:
//   - *EventFilter: The filter instance for method chaining
func (ef *EventFilter) WithNames(names ...string) *EventFilter {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.names = names
	return ef
}

// WithSources sets the source IDs to filter by.
//
// If multiple sources are provided, an event matches if its source ID
// matches any of the provided sources (OR logic). Matching is
// case-insensitive and supports substring matching.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	filter.WithSources("button-1", "form-1")
//
// Parameters:
//   - sources: Source IDs to filter by
//
// Returns:
//   - *EventFilter: The filter instance for method chaining
func (ef *EventFilter) WithSources(sources ...string) *EventFilter {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.sources = sources
	return ef
}

// WithTimeRange sets the time range to filter by.
//
// Events are matched if their timestamp falls within the range
// (inclusive of both start and end).
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	start := time.Now().Add(-1 * time.Hour)
//	end := time.Now()
//	filter.WithTimeRange(start, end)
//
// Parameters:
//   - start: Start time (inclusive)
//   - end: End time (inclusive)
//
// Returns:
//   - *EventFilter: The filter instance for method chaining
func (ef *EventFilter) WithTimeRange(start, end time.Time) *EventFilter {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.timeRange = &TimeRange{Start: start, End: end}
	return ef
}

// GetNames returns the current name filter criteria.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - []string: Copy of the name filter list
func (ef *EventFilter) GetNames() []string {
	ef.mu.RLock()
	defer ef.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]string, len(ef.names))
	copy(result, ef.names)
	return result
}

// GetSources returns the current source filter criteria.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - []string: Copy of the source filter list
func (ef *EventFilter) GetSources() []string {
	ef.mu.RLock()
	defer ef.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]string, len(ef.sources))
	copy(result, ef.sources)
	return result
}

// GetTimeRange returns the current time range filter criteria.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - *TimeRange: Copy of the time range, or nil if not set
func (ef *EventFilter) GetTimeRange() *TimeRange {
	ef.mu.RLock()
	defer ef.mu.RUnlock()
	if ef.timeRange == nil {
		return nil
	}
	// Return a copy to prevent external modification
	return &TimeRange{
		Start: ef.timeRange.Start,
		End:   ef.timeRange.End,
	}
}

// Clear removes all filter criteria.
//
// After clearing, the filter matches all events.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (ef *EventFilter) Clear() {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.names = []string{}
	ef.sources = []string{}
	ef.timeRange = nil
}

// Matches checks if an event matches the filter criteria.
//
// All filter criteria must match for the event to pass (AND logic).
// Within each criterion (names, sources), any match is sufficient (OR logic).
// If a criterion is empty, it matches all events.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if filter.Matches(event) {
//	    fmt.Println("Event passed filter")
//	}
//
// Parameters:
//   - event: The event to check
//
// Returns:
//   - bool: True if the event matches all criteria, false otherwise
func (ef *EventFilter) Matches(event EventRecord) bool {
	ef.mu.RLock()
	defer ef.mu.RUnlock()

	// Check name filter
	if len(ef.names) > 0 {
		if !ef.matchesName(event.Name) {
			return false
		}
	}

	// Check source filter
	if len(ef.sources) > 0 {
		if !ef.matchesSource(event.SourceID) {
			return false
		}
	}

	// Check time range filter
	if ef.timeRange != nil {
		if !ef.timeRange.Contains(event.Timestamp) {
			return false
		}
	}

	return true
}

// matchesName checks if an event name matches any of the name filters.
// Must be called with lock held.
func (ef *EventFilter) matchesName(name string) bool {
	nameLower := strings.ToLower(name)
	for _, filterName := range ef.names {
		if strings.Contains(nameLower, strings.ToLower(filterName)) {
			return true
		}
	}
	return false
}

// matchesSource checks if a source ID matches any of the source filters.
// Must be called with lock held.
func (ef *EventFilter) matchesSource(source string) bool {
	sourceLower := strings.ToLower(source)
	for _, filterSource := range ef.sources {
		if strings.Contains(sourceLower, strings.ToLower(filterSource)) {
			return true
		}
	}
	return false
}

// Apply filters a slice of events and returns only those that match.
//
// The returned slice contains only events that pass all filter criteria.
// The original slice is not modified.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	filtered := filter.Apply(events)
//	fmt.Printf("Filtered %d events to %d\n", len(events), len(filtered))
//
// Parameters:
//   - events: The events to filter
//
// Returns:
//   - []EventRecord: Filtered events
func (ef *EventFilter) Apply(events []EventRecord) []EventRecord {
	if len(events) == 0 {
		return []EventRecord{}
	}

	result := make([]EventRecord, 0, len(events))
	for _, event := range events {
		if ef.Matches(event) {
			result = append(result, event)
		}
	}

	return result
}

// Contains checks if a time falls within the time range (inclusive).
//
// Example:
//
//	if timeRange.Contains(event.Timestamp) {
//	    fmt.Println("Event is within range")
//	}
//
// Parameters:
//   - t: The time to check
//
// Returns:
//   - bool: True if the time is within the range, false otherwise
func (tr *TimeRange) Contains(t time.Time) bool {
	return (t.Equal(tr.Start) || t.After(tr.Start)) &&
		(t.Equal(tr.End) || t.Before(tr.End))
}
