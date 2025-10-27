package bubbly

import (
	"sync"
	"time"
)

// Event represents a component event with metadata.
// Events are emitted by components and can be listened to by parent components.
//
// The Event struct includes:
//   - Name: The event name (e.g., "click", "submit", "change")
//   - Source: The component that emitted the event
//   - Data: Arbitrary data associated with the event
//   - Timestamp: When the event was emitted
//
// Example:
//
//	event := Event{
//	    Name:      "submit",
//	    Source:    component,
//	    Data:      FormData{Username: "user"},
//	    Timestamp: time.Now(),
//	}
type Event struct {
	// Name is the event identifier (e.g., "click", "submit")
	Name string

	// Source is the component that emitted the event
	Source Component

	// Data is arbitrary data associated with the event
	// Handlers should type-assert this to the expected type
	Data interface{}

	// Timestamp is when the event was emitted
	Timestamp time.Time
}

// emitEvent is an internal method that creates and emits an Event.
// It handles event creation with proper metadata and calls all registered handlers.
//
// This method:
//   - Creates an Event struct with timestamp
//   - Looks up registered handlers for the event name
//   - Executes each handler with the event data
//   - Handles handler panics gracefully (future enhancement)
//
// Note: This is called by the public Emit() method on componentImpl.
// The Event struct is created for future use (e.g., event bubbling, logging).
func (c *componentImpl) emitEvent(eventName string, data interface{}) {
	// Create event with metadata
	// Note: Currently we pass data directly to handlers, but the Event struct
	// is available for future enhancements like event bubbling or logging.
	_ = Event{
		Name:      eventName,
		Source:    c,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Get handlers with read lock
	c.handlersMu.RLock()
	handlers, ok := c.handlers[eventName]
	c.handlersMu.RUnlock()

	// Execute all registered handlers for this event
	if ok {
		for _, handler := range handlers {
			// Call handler with event data
			// Note: In future, we may want to recover from panics here
			handler(data)
		}
	}
}

// registerHandler is an internal method that registers an event handler.
// It ensures thread-safe handler registration and supports multiple handlers per event.
//
// This method:
//   - Initializes the handlers map if needed
//   - Appends the handler to the list for the given event name
//   - Supports multiple handlers for the same event
//
// Note: This is called by the public On() method on componentImpl.
func (c *componentImpl) registerHandler(eventName string, handler EventHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()

	// Ensure handlers map is initialized
	if c.handlers == nil {
		c.handlers = make(map[string][]EventHandler)
	}

	// Append handler to the list for this event
	c.handlers[eventName] = append(c.handlers[eventName], handler)
}

// eventRegistry is a global registry for tracking event listeners.
// This is useful for debugging and testing event flow.
// Note: This is an optional enhancement for future use.
type eventRegistry struct {
	mu        sync.RWMutex
	listeners map[string]int // event name -> listener count
}

// Global event registry instance
var globalEventRegistry = &eventRegistry{
	listeners: make(map[string]int),
}

// trackEventListener increments the listener count for an event.
// This is useful for debugging and testing.
func (r *eventRegistry) trackEventListener(eventName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners[eventName]++
}

// getListenerCount returns the number of listeners for an event.
// This is useful for testing and debugging.
func (r *eventRegistry) getListenerCount(eventName string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.listeners[eventName]
}

// resetRegistry clears all listener counts.
// This is useful for testing.
func (r *eventRegistry) resetRegistry() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners = make(map[string]int)
}
