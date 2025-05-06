package bubble

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
)

// EventType identifies the type of event.
type EventType string

// EventPhase represents the phase of the event propagation.
type EventPhase int

// Standard event phases, mimicking DOM event phases for familiarity.
const (
	PhaseNone EventPhase = iota
	PhaseCapturePhase   // Event moving down to target
	PhaseAtTarget       // Event at target element
	PhaseBubblingPhase  // Event bubbling up from target
)

// Standard event types
const (
	EventTypeKey        EventType = "key"
	EventTypeMouse      EventType = "mouse"
	EventTypeWindowSize EventType = "windowSize"
	EventTypeCustom     EventType = "custom"
	// Categories for grouping events
	EventCategoryInput  EventType = "input"  // Keyboard, mouse, etc.
	EventCategoryUI     EventType = "ui"     // Window, resize, etc.
	EventCategorySystem EventType = "system" // Application specific
)

// Event is the interface that represents an event in the BubblyUI framework.
// It provides access to event metadata, source component, and propagation control.
type Event interface {
	// Type returns the type of the event.
	Type() EventType
	
	// Source returns the component that originated the event.
	Source() core.Component
	
	// Timestamp returns when the event was created.
	Timestamp() time.Time
	
	// Phase returns the current phase of event propagation.
	Phase() EventPhase
	
	// SetPhase updates the current phase of event propagation.
	SetPhase(phase EventPhase)
	
	// IsPropagationStopped returns whether propagation has been stopped.
	IsPropagationStopped() bool
	
	// StopPropagation prevents the event from further propagation.
	StopPropagation()
	
	// IsImmediatePropagationStopped returns whether immediate propagation has been stopped.
	IsImmediatePropagationStopped() bool
	
	// StopImmediatePropagation prevents the event from being handled by other handlers.
	StopImmediatePropagation()
	
	// IsDefaultPrevented returns whether the default action has been prevented.
	IsDefaultPrevented() bool
	
	// PreventDefault prevents the default action associated with the event.
	PreventDefault()
	
	// Path returns the component path from the target to the root.
	Path() []core.Component
	
	// SetPath sets the component path from the target to the root.
	SetPath(path []core.Component)
	
	// Category returns the category of the event.
	Category() EventType
	
	// RawMessage returns the original Bubble Tea message.
	RawMessage() tea.Msg
	
	// Priority returns the priority of the event.
	Priority() EventPriority
	
	// SetPriority sets the priority of the event.
	SetPriority(priority EventPriority)
}

// BaseEvent provides a base implementation of the Event interface.
// It is meant to be embedded in specific event types.
type BaseEvent struct {
	eventType             EventType
	sourceComponent       core.Component
	timeStamp             time.Time
	phase                 EventPhase
	propagationStopped    bool
	immediateStopped      bool
	defaultPrevented      bool
	componentPath         []core.Component
	category              EventType
	originalMessage       tea.Msg
	eventPriority         EventPriority
	eventContext          *EventContext  // Added field for event context information
}

// NewBaseEvent creates a new base event with common properties.
func NewBaseEvent(
	eventType EventType,
	source core.Component,
	category EventType,
	msg tea.Msg,
) *BaseEvent {
	// Determine default priority based on event type
	defaultPriority := PriorityNormal
	
	// Assign higher priority to mouse events
	if eventType == EventTypeMouse {
		defaultPriority = PriorityHigh
	}
	
	// Assign lower priority to window size events
	if eventType == EventTypeWindowSize {
		defaultPriority = PriorityLow
	}
	
	return &BaseEvent{
		eventType:          eventType,
		sourceComponent:    source,
		timeStamp:          time.Now(),
		phase:              PhaseAtTarget,
		propagationStopped: false,
		immediateStopped:   false,
		defaultPrevented:   false,
		componentPath:      []core.Component{source},
		category:           category,
		originalMessage:    msg,
		eventPriority:      defaultPriority,
	}
}

// Type returns the event type.
func (e *BaseEvent) Type() EventType {
	return e.eventType
}

// Source returns the component that originated the event.
func (e *BaseEvent) Source() core.Component {
	return e.sourceComponent
}

// Timestamp returns when the event was created.
func (e *BaseEvent) Timestamp() time.Time {
	return e.timeStamp
}

// Phase returns the current phase of event propagation.
func (e *BaseEvent) Phase() EventPhase {
	return e.phase
}

// SetPhase updates the current phase of event propagation.
func (e *BaseEvent) SetPhase(phase EventPhase) {
	e.phase = phase
}

// IsPropagationStopped returns whether propagation has been stopped.
func (e *BaseEvent) IsPropagationStopped() bool {
	return e.propagationStopped
}

// StopPropagation prevents the event from further propagation.
func (e *BaseEvent) StopPropagation() {
	e.propagationStopped = true
}

// IsImmediatePropagationStopped returns whether immediate propagation has been stopped.
func (e *BaseEvent) IsImmediatePropagationStopped() bool {
	return e.immediateStopped
}

// StopImmediatePropagation prevents the event from being handled by other handlers.
func (e *BaseEvent) StopImmediatePropagation() {
	e.immediateStopped = true
	e.propagationStopped = true
}

// IsDefaultPrevented returns whether the default action has been prevented.
func (e *BaseEvent) IsDefaultPrevented() bool {
	return e.defaultPrevented
}

// PreventDefault prevents the default action associated with the event.
func (e *BaseEvent) PreventDefault() {
	e.defaultPrevented = true
}

// Path returns the component path from the target to the root.
func (e *BaseEvent) Path() []core.Component {
	return e.componentPath
}

// SetPath sets the component path from the target to the root.
func (e *BaseEvent) SetPath(path []core.Component) {
	e.componentPath = path
}

// Category returns the category of the event.
func (e *BaseEvent) Category() EventType {
	return e.category
}

// RawMessage returns the original Bubble Tea message.
func (e *BaseEvent) RawMessage() tea.Msg {
	return e.originalMessage
}
