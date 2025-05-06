package bubble

import (
	"time"
	"sync"
)

// EventPriority defines the priority level of an event.
// Higher values indicate higher priority.
type EventPriority int

// Priority levels for events
const (
	PriorityLowest  EventPriority = 0  // Background processing events
	PriorityLow     EventPriority = 25 // Low priority events like window resize
	PriorityNormal  EventPriority = 50 // Default priority for most events
	PriorityHigh    EventPriority = 75 // High priority events like mouse clicks
	PriorityUrgent  EventPriority = 100 // Urgent events like critical errors or Ctrl+C
)

// Priority returns the priority of the event.
func (e *BaseEvent) Priority() EventPriority {
	return e.eventPriority
}

// SetPriority sets the priority of the event.
func (e *BaseEvent) SetPriority(priority EventPriority) {
	e.eventPriority = priority
}

// EventPriorityHandler manages event priorities and priority inheritance.
type EventPriorityHandler struct {
	// User interaction event types that get priority boost
	userInteractionEvents map[EventType]bool
}

// NewEventPriorityHandler creates a new event priority handler.
func NewEventPriorityHandler() *EventPriorityHandler {
	handler := &EventPriorityHandler{
		userInteractionEvents: make(map[EventType]bool),
	}

	// Register default user interaction events
	handler.userInteractionEvents[EventTypeKey] = true
	handler.userInteractionEvents[EventTypeMouse] = true

	return handler
}

// BoostUserInitiatedEvent increases the priority of user-initiated events.
func (h *EventPriorityHandler) BoostUserInitiatedEvent(event Event) Event {
	if h.userInteractionEvents[event.Type()] {
		// Only boost if it's not already high or urgent
		if event.Priority() < PriorityHigh {
			event.SetPriority(PriorityHigh)
		}
	}
	return event
}

// ApplyPriorityInheritance applies priority inheritance from parent to child event.
func (h *EventPriorityHandler) ApplyPriorityInheritance(childEvent Event, parentEvent Event) Event {
	// Child inherits parent priority only if parent priority is higher
	if parentEvent.Priority() > childEvent.Priority() {
		childEvent.SetPriority(parentEvent.Priority())
	}
	return childEvent
}

// PriorityEventQueue is a priority-based event queue that processes
// higher priority events before lower priority ones.
type PriorityEventQueue struct {
	mutex sync.Mutex
	// Queue of events grouped by priority
	queues map[EventPriority][]Event
	// Ordered list of priorities for efficient iteration
	priorities []EventPriority
}

// NewPriorityEventQueue creates a new priority-based event queue.
func NewPriorityEventQueue() *PriorityEventQueue {
	return &PriorityEventQueue{
		queues:     make(map[EventPriority][]Event),
		priorities: make([]EventPriority, 0),
	}
}

// Enqueue adds an event to the queue based on its priority.
func (q *PriorityEventQueue) Enqueue(event Event) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	priority := event.Priority()
	
	// Initialize priority queue if it doesn't exist
	if _, ok := q.queues[priority]; !ok {
		q.queues[priority] = make([]Event, 0)
		
		// Add priority to ordered list and sort
		q.priorities = append(q.priorities, priority)
		q.sortPriorities()
	}
	
	// Add event to appropriate priority queue
	q.queues[priority] = append(q.queues[priority], event)
}

// Dequeue removes and returns the highest priority event from the queue.
// Returns nil if the queue is empty.
func (q *PriorityEventQueue) Dequeue() Event {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.priorities) == 0 {
		return nil
	}

	// Get highest priority
	highestPriority := q.priorities[len(q.priorities)-1]
	events := q.queues[highestPriority]
	
	if len(events) == 0 {
		// This shouldn't happen, but handle it anyway
		// Remove this priority since it has no events
		q.priorities = q.priorities[:len(q.priorities)-1]
		delete(q.queues, highestPriority)
		
		// Try again with next highest priority
		if len(q.priorities) > 0 {
			return q.Dequeue()
		}
		return nil
	}

	// Get first event in the queue (FIFO for same priority)
	event := events[0]
	
	// Update queue
	q.queues[highestPriority] = events[1:]
	
	// If queue is now empty, remove it from priorities
	if len(q.queues[highestPriority]) == 0 {
		delete(q.queues, highestPriority)
		q.priorities = q.priorities[:len(q.priorities)-1]
	}
	
	return event
}

// IsEmpty returns true if the queue is empty.
func (q *PriorityEventQueue) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	return len(q.priorities) == 0
}

// sortPriorities sorts priorities in ascending order (highest at the end).
func (q *PriorityEventQueue) sortPriorities() {
	// Simple insertion sort since we typically have few priority levels
	for i := 1; i < len(q.priorities); i++ {
		key := q.priorities[i]
		j := i - 1
		
		for j >= 0 && q.priorities[j] > key {
			q.priorities[j+1] = q.priorities[j]
			j--
		}
		
		q.priorities[j+1] = key
	}
}

// EventThrottler provides throttling for low-priority events.
type EventThrottler struct {
	mutex                sync.Mutex
	lastProcessedTime    map[EventType]time.Time
	throttleDurations    map[EventType]time.Duration
	defaultThrottleTime  time.Duration
}

// NewEventThrottler creates a new event throttler with default settings.
func NewEventThrottler() *EventThrottler {
	return &EventThrottler{
		lastProcessedTime:   make(map[EventType]time.Time),
		throttleDurations:   make(map[EventType]time.Duration),
		defaultThrottleTime: 100 * time.Millisecond, // Default throttle duration
	}
}

// ConfigureEventTypeThrottling sets a custom throttle duration for a specific event type.
func (t *EventThrottler) ConfigureEventTypeThrottling(eventType EventType, duration time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.throttleDurations[eventType] = duration
}

// ShouldProcess determines if an event should be processed based on throttling rules.
func (t *EventThrottler) ShouldProcess(event Event) bool {
	// High priority events always bypass throttling
	if event.Priority() >= PriorityHigh {
		return true
	}
	
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	eventType := event.Type()
	now := time.Now()
	
	// Get the appropriate throttle duration for this event type
	duration, ok := t.throttleDurations[eventType]
	if !ok {
		duration = t.defaultThrottleTime
	}
	
	// Check if enough time has passed since last processed event of this type
	lastTime, exists := t.lastProcessedTime[eventType]
	if !exists || now.Sub(lastTime) >= duration {
		// Update last processed time
		t.lastProcessedTime[eventType] = now
		return true
	}
	
	// Not enough time has passed, throttle the event
	return false
}
