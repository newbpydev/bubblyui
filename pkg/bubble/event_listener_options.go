package bubble

// EventListenerOptions defines configuration options for event listeners
type EventListenerOptions struct {
	// Phase specifies when the listener should be called during event propagation
	// Default is PhaseBubblingPhase if not specified
	Phase EventPhase
	
	// Priority determines the execution order of listeners for the same phase
	// Higher priority listeners execute before lower priority listeners
	Priority EventPriority
	
	// Once indicates if the listener should be automatically removed after being called once
	Once bool
	
	// Passive indicates that the listener will never call preventDefault()
	// This can enable performance optimizations
	Passive bool
}

// DefaultEventListenerOptions returns the default options for event listeners
func DefaultEventListenerOptions() EventListenerOptions {
	return EventListenerOptions{
		Phase:    PhaseBubblingPhase,
		Priority: PriorityNormal,
		Once:     false,
		Passive:  false,
	}
}
