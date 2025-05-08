package bubble

import (
	"sync/atomic"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
)

// KeyEvent implementation
func (e *KeyEvent) Context() *EventContext {
	// Access base event's context
	return e.BaseEvent.Context()
}

func (e *KeyEvent) SwitchToPhase(phase EventPhase) {
	e.BaseEvent.SwitchToPhase(phase)
}

// MouseEvent implementation
func (e *MouseEvent) Context() *EventContext {
	// Access base event's context
	return e.BaseEvent.Context()
}

func (e *MouseEvent) SwitchToPhase(phase EventPhase) {
	e.BaseEvent.SwitchToPhase(phase)
}

// WindowSizeEvent implementation
func (e *WindowSizeEvent) Context() *EventContext {
	// Access base event's context
	return e.BaseEvent.Context()
}

func (e *WindowSizeEvent) SwitchToPhase(phase EventPhase) {
	e.BaseEvent.SwitchToPhase(phase)
}

// UserDefinedEvent implementation
func (e *UserDefinedEvent) Context() *EventContext {
	// Access base event's context
	return e.BaseEvent.Context()
}

func (e *UserDefinedEvent) SwitchToPhase(phase EventPhase) {
	e.BaseEvent.SwitchToPhase(phase)
}

// TestCustomEvent implementation (for test files)
type TestCustomEvent struct {
	*BaseEvent
	CustomField string
}

func (e *TestCustomEvent) Context() *EventContext {
	// Access base event's context
	return e.BaseEvent.Context()
}

func (e *TestCustomEvent) SwitchToPhase(phase EventPhase) {
	e.BaseEvent.SwitchToPhase(phase)
}

// Helper function to create event context
func createEventContext(component core.Component, phase EventPhase) *EventContext {
	return &EventContext{
		Timestamp:      time.Now(),
		SequenceNumber: atomic.AddUint64(&globalEventSequence, 1),
		CurrentPhase:   phase,
		OriginalTarget: component,
		CurrentTarget:  component,
		PropagationPath: []core.Component{component},
		UserInteraction: UserInteractionContext{
			Modifiers: make(map[string]bool),
		},
		ApplicationState: make(map[string]interface{}),
	}
}
