package testutil

import (
	"sync"
	"time"
)

// EventTracker tracks events emitted during component testing.
// It provides thread-safe event tracking with methods to query event history.
//
// EventTracker is used internally by the test harness to track all events
// emitted by components during tests, enabling assertions on event behavior.
//
// All methods are thread-safe and can be called concurrently from multiple goroutines.
type EventTracker struct {
	events []EmittedEvent
	mu     sync.RWMutex
}

// EmittedEvent represents an event that was emitted during testing.
// It captures the event name, payload, timestamp, and source component.
type EmittedEvent struct {
	Name      string
	Payload   interface{}
	Timestamp time.Time
	Source    string
}

// NewEventTracker creates a new event tracker with an empty event list.
//
// Example:
//
//	tracker := testutil.NewEventTracker()
//	tracker.Track("click", "button-data", "button-1")
func NewEventTracker() *EventTracker {
	return &EventTracker{
		events: []EmittedEvent{},
	}
}

// Track records an event emission with the given name, payload, and source.
// This method is thread-safe and can be called concurrently.
//
// The timestamp is automatically set to the current time when Track is called.
//
// Parameters:
//   - name: The name of the event (e.g., "click", "submit", "change")
//   - payload: The event payload (can be nil or any type)
//   - source: The source component that emitted the event (component ID)
//
// Example:
//
//	tracker.Track("click", map[string]int{"x": 10, "y": 20}, "button-1")
func (et *EventTracker) Track(name string, payload interface{}, source string) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.events = append(et.events, EmittedEvent{
		Name:      name,
		Payload:   payload,
		Timestamp: time.Now(),
		Source:    source,
	})
}

// GetEvents returns all events with the given name.
// Returns an empty slice if no events with that name were tracked.
// This method is thread-safe.
//
// The returned slice is a copy of the matching events, so modifying it
// will not affect the tracker's internal state.
//
// Example:
//
//	clickEvents := tracker.GetEvents("click")
//	for _, event := range clickEvents {
//	    fmt.Printf("Click at %v from %s\n", event.Timestamp, event.Source)
//	}
func (et *EventTracker) GetEvents(name string) []EmittedEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	events := []EmittedEvent{}
	for _, e := range et.events {
		if e.Name == name {
			events = append(events, e)
		}
	}

	return events
}

// WasFired returns true if at least one event with the given name was tracked.
// This method is thread-safe.
//
// This is a convenience method equivalent to checking if len(GetEvents(name)) > 0.
//
// Example:
//
//	if tracker.WasFired("submit") {
//	    fmt.Println("Form was submitted")
//	}
func (et *EventTracker) WasFired(name string) bool {
	return len(et.GetEvents(name)) > 0
}

// FiredCount returns the number of times an event with the given name was tracked.
// Returns 0 if no events with that name were tracked.
// This method is thread-safe.
//
// This is a convenience method equivalent to len(GetEvents(name)).
//
// Example:
//
//	count := tracker.FiredCount("click")
//	fmt.Printf("Button was clicked %d times\n", count)
func (et *EventTracker) FiredCount(name string) int {
	return len(et.GetEvents(name))
}

// Clear removes all tracked events from the tracker.
// This method is thread-safe and idempotent (calling it multiple times is safe).
//
// After calling Clear, all query methods (GetEvents, WasFired, FiredCount)
// will return empty results until new events are tracked.
//
// Example:
//
//	tracker.Track("click", nil, "button-1")
//	tracker.Clear()
//	// tracker.WasFired("click") now returns false
func (et *EventTracker) Clear() {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.events = []EmittedEvent{}
}
