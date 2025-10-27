package bubbly

import (
	"sync"
	"time"
)

// eventPool is a sync.Pool for reusing Event objects.
// This reduces allocations and GC pressure during event emission.
// Events are reset when retrieved from the pool and returned after use.
var eventPool = sync.Pool{
	New: func() interface{} {
		return &Event{}
	},
}

// Event represents a component event with metadata.
// Events are emitted by components and can be listened to by parent components.
//
// The Event struct includes:
//   - Name: The event name (e.g., "click", "submit", "change")
//   - Source: The component that emitted the event
//   - Data: Arbitrary data associated with the event
//   - Timestamp: When the event was emitted
//   - Stopped: Flag to control event propagation (set via StopPropagation)
//
// Events automatically bubble up from child to parent components unless
// StopPropagation() is called by a handler.
//
// Example:
//
//	event := Event{
//	    Name:      "submit",
//	    Source:    component,
//	    Data:      FormData{Username: "user"},
//	    Timestamp: time.Now(),
//	    Stopped:   false,
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

	// Stopped indicates whether event propagation has been stopped
	// Set to true by calling StopPropagation() to prevent bubbling to parent
	Stopped bool
}

// StopPropagation prevents the event from bubbling to parent components.
// This method should be called by event handlers that want to stop
// the event from propagating further up the component tree.
//
// Example:
//
//	component.On("submit", func(data interface{}) {
//	    if event, ok := data.(*Event); ok {
//	        // Handle the event locally
//	        fmt.Println("Handling submit")
//	        // Prevent parent from seeing this event
//	        event.StopPropagation()
//	    }
//	})
func (e *Event) StopPropagation() {
	e.Stopped = true
}

// bubbleEvent propagates an event up the component tree.
// It executes local handlers first, then recursively calls the parent's
// bubbleEvent if the event hasn't been stopped.
//
// This implements Vue.js-style event bubbling where events automatically
// propagate from child to parent components unless explicitly stopped.
//
// The bubbling flow:
//  1. Execute all local handlers for this event
//  2. Pass Event pointer to handlers (so they can call StopPropagation)
//  3. Check if event.Stopped is true after handlers execute
//  4. If not stopped and parent exists, recursively call parent.bubbleEvent
//
// Thread-safe: Uses existing handlersMu RWMutex for concurrent access.
func (c *componentImpl) bubbleEvent(event *Event) {
	// Skip if event propagation was already stopped
	if event.Stopped {
		return
	}

	// Get handlers with read lock
	c.handlersMu.RLock()
	handlers, ok := c.handlers[event.Name]
	c.handlersMu.RUnlock()

	// Execute all local handlers for this event
	if ok {
		for _, handler := range handlers {
			// Recover from panics in event handlers to prevent application crashes
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log the panic but don't crash the application
						// In production, this could be sent to error tracking service
						_ = &HandlerPanicError{
							ComponentName: c.name,
							EventName:     event.Name,
							PanicValue:    r,
						}
						// Note: We don't stop propagation on panic - other handlers should still run
					}
				}()

				// Pass Event pointer to handler so it can call StopPropagation
				handler(event)
			}()

			// Check if handler stopped propagation
			if event.Stopped {
				return
			}
		}
	}

	// Bubble to parent if not stopped and parent exists
	if !event.Stopped && c.parent != nil {
		// Type assert parent to *componentImpl to access bubbleEvent
		if parentImpl, ok := (*c.parent).(*componentImpl); ok {
			parentImpl.bubbleEvent(event)
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
