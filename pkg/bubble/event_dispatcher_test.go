package bubble

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
)

// createTestEventDispatcher creates an event dispatcher for testing
func createTestEventDispatcher(t *testing.T, root, parent, child *MockComponent) *EventDispatcher {
	// Setup basic hierarchy
	if root != nil && parent != nil {
		root.children = []core.Component{parent}
	}
	if parent != nil && child != nil {
		parent.children = []core.Component{child}
	}
	if parent != nil && root != nil {
		parent.parent = root
	}
	if child != nil && parent != nil {
		child.parent = parent
	}

	// Create and setup dispatcher
	dispatcher := NewEventDispatcher()
	if root != nil {
		dispatcher.SetRootComponent(root)
	}
	return dispatcher
}

// dispatchTestEvent is a helper function for dispatching events in tests
func dispatchTestEvent(dispatcher *EventDispatcher, source core.Component) Event {
	event := NewKeyEvent(source, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	dispatcher.DispatchEvent(event)
	return event
}

// manualEventPropagation performs a manual event propagation for testing purposes
// This works around any issues with the mock framework
func manualEventPropagation(t *testing.T, event Event, components []core.Component) {
	// Make sure we have a valid event
	assert.NotNil(t, event, "Event should not be nil")
	
	// Set the path in the event
	event.SetPath(components)
	
	// Manually call HandleEvent on each component
	for _, comp := range components {
		if handler, ok := comp.(interface{ HandleEvent(Event) bool }); ok {
			handler.HandleEvent(event)
		}
	}
}
