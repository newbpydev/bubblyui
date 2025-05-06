package bubble

import (
	"sync/atomic"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
)

// UserInteractionContext captures information about user interaction
// that led to an event.
type UserInteractionContext struct {
	// InputType indicates the type of input (keyboard, mouse, etc.)
	InputType string
	
	// CursorX and CursorY represent the current cursor position
	CursorX int
	CursorY int
	
	// KeyPressed contains the key that was pressed (for keyboard events)
	KeyPressed string
	
	// Screen dimensions
	ScreenWidth  int
	ScreenHeight int
	
	// Additional metadata about the interaction
	Modifiers map[string]bool
	
	// State of user focus
	FocusedComponentID string
}

// EventContext contains enriched metadata about an event context.
type EventContext struct {
	// When the event was created
	Timestamp time.Time
	
	// Globally incrementing sequence number
	SequenceNumber uint64
	
	// User interaction information
	UserInteraction UserInteractionContext
	
	// Application state at the time of the event
	ApplicationState map[string]interface{}
}

// Global sequence counter for events
var globalEventSequence uint64

// EventContextualizer enriches events with additional context information.
type EventContextualizer struct {
	rootComponent  core.Component
	appState       map[string]interface{}
}

// NewEventContextualizer creates a new EventContextualizer.
func NewEventContextualizer() *EventContextualizer {
	return &EventContextualizer{
		appState: make(map[string]interface{}),
	}
}

// SetRootComponent sets the root component for building component paths.
func (ec *EventContextualizer) SetRootComponent(root core.Component) {
	ec.rootComponent = root
}

// SetApplicationState sets the application state to be included in event context.
func (ec *EventContextualizer) SetApplicationState(state map[string]interface{}) {
	// Make a deep copy of the state to ensure it's not modified after being set
	ec.appState = deepCopyMap(state)
}

// EnrichEventContext adds context information to an event.
func (ec *EventContextualizer) EnrichEventContext(event Event, sourceComponent core.Component) (Event, error) {
	// We need to add context to the base event
	switch e := event.(type) {
	case *KeyEvent:
		ec.enrichBaseEvent(e.BaseEvent, sourceComponent)
		ec.enrichKeyEvent(e)
		return e, nil
		
	case *MouseEvent:
		ec.enrichBaseEvent(e.BaseEvent, sourceComponent)
		ec.enrichMouseEvent(e)
		return e, nil
		
	case *WindowSizeEvent:
		ec.enrichBaseEvent(e.BaseEvent, sourceComponent)
		ec.enrichWindowSizeEvent(e)
		return e, nil
		
	case *UserDefinedEvent:
		ec.enrichBaseEvent(e.BaseEvent, sourceComponent)
		return e, nil
		
	default:
		// For other event types, just return as is
		return event, nil
	}
}

// enrichBaseEvent adds context information to a BaseEvent.
func (ec *EventContextualizer) enrichBaseEvent(baseEvent *BaseEvent, sourceComponent core.Component) {
	// Create event context with timestamp and sequence number
	context := &EventContext{
		Timestamp:     time.Now(),
		SequenceNumber: atomic.AddUint64(&globalEventSequence, 1),
		UserInteraction: UserInteractionContext{
			Modifiers: make(map[string]bool),
		},
		ApplicationState: deepCopyMap(ec.appState),
	}
	
	// Attach the context to the base event
	baseEvent.eventContext = context
	
	// Build component path if root component is set
	if ec.rootComponent != nil {
		baseEvent.componentPath = ec.buildComponentPath(sourceComponent)
	} else if len(baseEvent.componentPath) == 0 {
		// Ensure at least the source component is in the path
		baseEvent.componentPath = []core.Component{sourceComponent}
	}
}

// enrichKeyEvent adds keyboard-specific context information.
func (ec *EventContextualizer) enrichKeyEvent(keyEvent *KeyEvent) {
	context := keyEvent.BaseEvent.eventContext
	context.UserInteraction.InputType = "keyboard"
	if len(keyEvent.KeyMsg.Runes) > 0 {
		context.UserInteraction.KeyPressed = string(keyEvent.KeyMsg.Runes)
	} else {
		context.UserInteraction.KeyPressed = keyEvent.String()
	}
	
	// Set modifier keys
	context.UserInteraction.Modifiers["alt"] = keyEvent.KeyMsg.Alt
}

// enrichMouseEvent adds mouse-specific context information.
func (ec *EventContextualizer) enrichMouseEvent(mouseEvent *MouseEvent) {
	context := mouseEvent.BaseEvent.eventContext
	context.UserInteraction.InputType = "mouse"
	context.UserInteraction.CursorX = mouseEvent.X()
	context.UserInteraction.CursorY = mouseEvent.Y()
	
	// Set modifier keys
	context.UserInteraction.Modifiers["alt"] = mouseEvent.Alt()
	context.UserInteraction.Modifiers["ctrl"] = mouseEvent.Ctrl()
	context.UserInteraction.Modifiers["shift"] = mouseEvent.Shift()
}

// enrichWindowSizeEvent adds window-specific context information.
func (ec *EventContextualizer) enrichWindowSizeEvent(windowEvent *WindowSizeEvent) {
	context := windowEvent.BaseEvent.eventContext
	context.UserInteraction.InputType = "system"
	context.UserInteraction.ScreenWidth = windowEvent.Width()
	context.UserInteraction.ScreenHeight = windowEvent.Height()
}

// buildComponentPath builds the complete path from a component to the root.
func (ec *EventContextualizer) buildComponentPath(component core.Component) []core.Component {
	if component == nil {
		return nil
	}
	
	path := []core.Component{component}
	currentComp := component
	
	// Recursively find path to root component
	// This implementation assumes a tree-walking algorithm to find parent-child relationships
	// In a real implementation, components might have direct parent references
	for currentComp != ec.rootComponent {
		parent := ec.findParent(currentComp, ec.rootComponent)
		if parent == nil {
			break
		}
		path = append(path, parent)
		currentComp = parent
	}
	
	return path
}

// findParent finds the parent of a component in the component tree.
func (ec *EventContextualizer) findParent(target core.Component, current core.Component) core.Component {
	// Check if any child of current is the target
	for _, child := range current.Children() {
		if child == target {
			return current
		}
	}
	
	// Recursively check children
	for _, child := range current.Children() {
		if parent := ec.findParent(target, child); parent != nil {
			return parent
		}
	}
	
	return nil
}

// Helper function to deep copy a map
func deepCopyMap(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return nil
	}
	
	copy := make(map[string]interface{}, len(original))
	for key, value := range original {
		// Handle nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			copy[key] = deepCopyMap(nestedMap)
		} else {
			// For other types, just copy the value
			// Note: This is a shallow copy for non-map types
			copy[key] = value
		}
	}
	
	return copy
}

// EventContext-related methods for BaseEvent can be added in the event.go file
// We don't need to redefine the struct here
