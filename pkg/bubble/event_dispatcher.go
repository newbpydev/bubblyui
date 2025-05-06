package bubble

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
)

// EventListenerFunc is a function that handles an event
// Returns true if the event propagation should stop
type EventListenerFunc func(event Event) bool

// listenerEntry represents a registered event listener
type listenerEntry struct {
	id         string
	component  core.Component
	eventType  EventType
	handler    EventListenerFunc
	registered time.Time
}

// EventDispatcher manages event propagation and listener registration
type EventDispatcher struct {
	rootComponent    core.Component
	listeners        map[string]listenerEntry
	debugMode        bool
	nextListenerId   int
	mu               sync.RWMutex
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners:        make(map[string]listenerEntry),
		debugMode:        false,
		nextListenerId:   1,
	}
}

// SetRootComponent sets the root component for event propagation
func (d *EventDispatcher) SetRootComponent(root core.Component) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.rootComponent = root
}

// AddEventListener registers a new event listener
func (d *EventDispatcher) AddEventListener(component core.Component, eventType EventType, handler EventListenerFunc) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Generate a unique ID for this listener
	id := strconv.Itoa(d.nextListenerId)
	d.nextListenerId++
	
	// Register the listener
	d.listeners[id] = listenerEntry{
		id:         id,
		component:  component,
		eventType:  eventType,
		handler:    handler,
		registered: time.Now(),
	}
	
	return id
}

// RemoveEventListener removes a registered event listener
func (d *EventDispatcher) RemoveEventListener(listenerId string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	delete(d.listeners, listenerId)
}

// EnableDebugMode toggles debug mode for event propagation
func (d *EventDispatcher) EnableDebugMode(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.debugMode = enabled
}

// DispatchEvent dispatches an event through the component hierarchy
func (d *EventDispatcher) DispatchEvent(event Event) {
	d.dispatchEventInternal(event, nil)
}

// DispatchEventWithDebugLogs dispatches an event and returns debug logs
func (d *EventDispatcher) DispatchEventWithDebugLogs(event Event) []string {
	logs := make([]string, 0)
	d.dispatchEventInternal(event, &logs)
	return logs
}

// dispatchEventInternal handles the actual event dispatching
func (d *EventDispatcher) dispatchEventInternal(event Event, logs *[]string) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	// Get the source component
	source := event.Source()
	if source == nil {
		return
	}
	
	// Set the event phase to bubbling
	event.SetPhase(PhaseBubblingPhase)
	
	// Build the propagation path from source to root
	var path []core.Component
	
	// For test mocks - first check if each component has a Parent() method
	current := source
	path = append(path, current) // Start with the source
	
	// Keep traversing up until we hit the root or nil
	for current != nil && current != d.rootComponent {
		var parent core.Component
		
		// Get the parent using reflection or interface assertion
		if parentGetter, ok := current.(interface{ Parent() core.Component }); ok {
			parent = parentGetter.Parent()
		}
		
		// If we found a parent, add it to the path and continue
		if parent != nil {
			path = append(path, parent)
			current = parent
		} else {
			break // No more parents
		}
	}
	
	// Ensure the root is in the path if it wasn't already added
	if d.rootComponent != nil && (len(path) == 0 || path[len(path)-1] != d.rootComponent) {
		path = append(path, d.rootComponent)
	}
	
	// Store the path in the event
	event.SetPath(path)
	
	// Generate debug logs if needed
	if d.debugMode && logs != nil {
		for _, comp := range path {
			*logs = append(*logs, fmt.Sprintf("Event %s bubbling through component %s", 
				event.Type(), comp.ID()))
		}
	}
	
	// First check for direct listeners for this event's source
	// This is critical for the event listener registration tests
	sourceComponent := event.Source()
	if sourceComponent != nil {
		for _, entry := range d.listeners {
			// Only call listeners that match this component and event type
			if entry.component == sourceComponent && entry.eventType == event.Type() {
				// Call the handler function directly
				stopPropagation := entry.handler(event)
				if stopPropagation || event.IsPropagationStopped() {
					return
				}
			}
		}
	}
	
	// Then propagate the event through each component in the path
	for _, comp := range path {
		// Call the component's HandleEvent method if available
		if handler, ok := comp.(interface{ HandleEvent(Event) bool }); ok {
			// Call the event handler directly
			if stopPropagation := handler.HandleEvent(event); stopPropagation {
				break
			}
		}
		
		// Check for listeners registered for this component
		for _, entry := range d.listeners {
			// Only call listeners that match this component and event type
			if entry.component == comp && entry.eventType == event.Type() {
				// Skip the source component's listeners as we've already processed them
				if comp == sourceComponent {
					continue
				}
				
				// Call the handler function
				stopPropagation := entry.handler(event)
				if stopPropagation || event.IsPropagationStopped() {
					break
				}
			}
		}
		
		// Check if propagation was explicitly stopped
		if event.IsPropagationStopped() {
			break
		}
	}
}

// buildPropagationPath builds the path from source to root component
func (d *EventDispatcher) buildPropagationPath(source core.Component) []core.Component {
	// Start with an empty path
	path := []core.Component{}
	
	// If no source or root, return empty path
	if source == nil || d.rootComponent == nil {
		return path
	}
	
	// Add the source component as the first in the path
	path = append(path, source)
	
	// If source is the root, we're done
	if source == d.rootComponent {
		return path
	}
	
	// Build the path from source to root
	current := source
	for current != nil && current != d.rootComponent {
		// Get parent through the Parent method
		var parent core.Component
		if parentProvider, ok := current.(interface{ Parent() core.Component }); ok {
			parent = parentProvider.Parent()
		}
		
		// If we found a parent, add it to the path and continue
		if parent != nil {
			path = append(path, parent)
			current = parent
		} else {
			// No parent found, stop here
			break
		}
	}
	
	// If we didn't reach the root component, add it explicitly
	if len(path) > 0 && path[len(path)-1] != d.rootComponent {
		path = append(path, d.rootComponent)
	}
	
	return path
}

// dispatchToListeners dispatches an event to all registered listeners for a component
func (d *EventDispatcher) dispatchToListeners(component core.Component, event Event) bool {
	for _, entry := range d.listeners {
		// Skip listeners for other components or event types
		if entry.component != component || entry.eventType != event.Type() {
			continue
		}
		
		// Call the handler
		stopPropagation := entry.handler(event)
		if stopPropagation {
			return true
		}
	}
	
	return false
}
