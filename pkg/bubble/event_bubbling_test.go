package bubble

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestEventPropagationMechanism tests the basic propagation of events through a component hierarchy
func TestEventPropagationMechanism(t *testing.T) {
	// Create a component hierarchy for testing
	root := new(MockComponent)
	parent := new(MockComponent)
	child := new(MockComponent)
	
	// Configure mocks
	root.On("ID").Return("root")
	root.On("Children").Return([]core.Component{parent})
	root.On("AddChild", mock.Anything).Run(func(args mock.Arguments) {
		root.children = append(root.children, args.Get(0).(core.Component))
	})
	
	parent.On("ID").Return("parent")
	parent.On("Children").Return([]core.Component{child})
	parent.On("AddChild", mock.Anything).Run(func(args mock.Arguments) {
		parent.children = append(parent.children, args.Get(0).(core.Component))
	})
	parent.On("Parent").Return(root)
	parent.On("SetParent", mock.Anything).Run(func(args mock.Arguments) {
		parent.parent = args.Get(0).(core.Component)
	})
	
	child.On("ID").Return("child")
	child.On("Children").Return([]core.Component{})
	child.On("Parent").Return(parent)
	child.On("SetParent", mock.Anything).Run(func(args mock.Arguments) {
		child.parent = args.Get(0).(core.Component)
	})
	
	// Setup hierarchy
	root.AddChild(parent)
	parent.AddChild(child)
	parent.SetParent(root)
	child.SetParent(parent)
	
	// Create event dispatcher/propagator
	dispatcher := NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	
	t.Run("Events bubble up from child to parent", func(t *testing.T) {
		// Track events as they bubble
		var capturedEvents []Event
		
		// Setup event handlers on components
		root.On("HandleEvent", mock.Anything).Run(func(args mock.Arguments) {
			// Record that the event reached the root
			event := args.Get(0).(Event)
			capturedEvents = append(capturedEvents, event)
		}).Return(false) // don't stop propagation
		
		parent.On("HandleEvent", mock.Anything).Run(func(args mock.Arguments) {
			// Record that the event reached the parent
			event := args.Get(0).(Event)
			capturedEvents = append(capturedEvents, event)
		}).Return(false) // don't stop propagation
		
		child.On("HandleEvent", mock.Anything).Run(func(args mock.Arguments) {
			// Record that the event reached the child
			event := args.Get(0).(Event)
			capturedEvents = append(capturedEvents, event)
		}).Return(false) // don't stop propagation
		
		// Create a key event from the child
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Dispatch the event (should bubble up)
		dispatcher.DispatchEvent(event)
		
		// Verify the event bubbled correctly
		assert.Equal(t, 3, len(capturedEvents), "Event should be handled by 3 components")
		
		// Events should be captured in bubbling order: child -> parent -> root
		assert.Equal(t, "child", capturedEvents[0].Source().ID(), "First handler should be the source (child)")
		assert.Equal(t, "parent", capturedEvents[1].Source().ID(), "Second handler should be the parent")
		assert.Equal(t, "root", capturedEvents[2].Source().ID(), "Third handler should be the root")
	})
	
	t.Run("Event propagation path is recorded", func(t *testing.T) {
		// Create a key event from the child
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Setup an event with context that records the path
		contextualizer := NewEventContextualizer()
		contextualizer.SetRootComponent(root)
		enrichedEvent, _ := contextualizer.EnrichEventContext(event, child)
		
		// Dispatch the event
		dispatcher.DispatchEvent(enrichedEvent)
		
		// After bubbling, event should have a complete propagation path
		// Cast to access internal fields if needed
		keyEvent, ok := enrichedEvent.(*KeyEvent)
		assert.True(t, ok, "Event should be a KeyEvent")
		
		// Verify the event context contains path information
		assert.NotNil(t, keyEvent.eventContext, "Event should have context after bubbling")
		
		// The dispatcher should update the event's phase during bubbling
		assert.Equal(t, PhaseBubblingPhase, enrichedEvent.Phase(), "Event phase should be set to bubbling")
	})
}

// TestEventListenerRegistration tests the event listener registration system
func TestEventListenerRegistration(t *testing.T) {
	// Create component for testing
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	
	// Create event dispatcher
	dispatcher := NewEventDispatcher()
	
	t.Run("Components can register event listeners", func(t *testing.T) {
		// Create a handler function
		var eventReceived bool
		handlerFunc := func(e Event) bool {
			eventReceived = true
			return false // don't stop propagation
		}
		
		// Register for key events
		listenerId := dispatcher.AddEventListener(comp, EventTypeKey, handlerFunc)
		assert.NotEmpty(t, listenerId, "Listener ID should not be empty")
		
		// Create and dispatch a key event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		dispatcher.DispatchEvent(event)
		
		// Verify handler was called
		assert.True(t, eventReceived, "Event handler should have been called")
	})
	
	t.Run("Listeners can be removed", func(t *testing.T) {
		// Create a handler function
		callCount := 0
		handlerFunc := func(e Event) bool {
			callCount++
			return false // don't stop propagation
		}
		
		// Register for key events
		listenerId := dispatcher.AddEventListener(comp, EventTypeKey, handlerFunc)
		
		// Create and dispatch a key event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		dispatcher.DispatchEvent(event)
		
		// Verify handler was called once
		assert.Equal(t, 1, callCount, "Event handler should have been called once")
		
		// Remove the listener
		dispatcher.RemoveEventListener(listenerId)
		
		// Dispatch another event
		dispatcher.DispatchEvent(event)
		
		// Verify handler wasn't called again
		assert.Equal(t, 1, callCount, "Event handler should not have been called after removal")
	})
	
	t.Run("Multiple listeners per event type are supported", func(t *testing.T) {
		// Track which handlers were called
		handler1Called := false
		handler2Called := false
		
		// Create handler functions
		handler1 := func(e Event) bool {
			handler1Called = true
			return false // don't stop propagation
		}
		
		handler2 := func(e Event) bool {
			handler2Called = true
			return false // don't stop propagation
		}
		
		// Register both for key events
		dispatcher.AddEventListener(comp, EventTypeKey, handler1)
		dispatcher.AddEventListener(comp, EventTypeKey, handler2)
		
		// Create and dispatch a key event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		dispatcher.DispatchEvent(event)
		
		// Verify both handlers were called
		assert.True(t, handler1Called, "First handler should have been called")
		assert.True(t, handler2Called, "Second handler should have been called")
	})
}

// TestPropagationDebugging tests the event propagation debugging tools
func TestPropagationDebugging(t *testing.T) {
	// Create component hierarchy
	root := new(MockComponent)
	child := new(MockComponent)
	
	root.On("ID").Return("root")
	root.On("Children").Return([]core.Component{child})
	root.On("AddChild", mock.Anything).Run(func(args mock.Arguments) {
		root.children = append(root.children, args.Get(0).(core.Component))
	})
	
	child.On("ID").Return("child")
	child.On("Children").Return([]core.Component{})
	child.On("Parent").Return(root)
	child.On("SetParent", mock.Anything).Run(func(args mock.Arguments) {
		child.parent = args.Get(0).(core.Component)
	})
	
	// Setup hierarchy
	root.AddChild(child)
	child.SetParent(root)
	
	// Create dispatcher with debug mode enabled
	dispatcher := NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	dispatcher.EnableDebugMode(true)
	
	t.Run("Debug logs capture propagation path", func(t *testing.T) {
		// Configure handlers
		root.On("HandleEvent", mock.Anything).Return(false)
		child.On("HandleEvent", mock.Anything).Return(false)
		
		// Create and dispatch an event
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Capture logs during dispatch
		logs := dispatcher.DispatchEventWithDebugLogs(event)
		
		// Verify logs contain propagation information
		assert.Greater(t, len(logs), 0, "Debug logs should contain entries")
		assert.Contains(t, logs[0], "child", "Logs should mention the source component")
		
		// Verify we have one log entry for each component in the propagation path
		assert.Equal(t, 2, len(logs), "Should have one log entry per component in path")
	})
	
	t.Run("Debug mode can be toggled", func(t *testing.T) {
		// Disable debug mode
		dispatcher.EnableDebugMode(false)
		
		// Create and dispatch an event
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Capture logs during dispatch
		logs := dispatcher.DispatchEventWithDebugLogs(event)
		
		// Verify no logs were captured when debug mode is off
		assert.Equal(t, 0, len(logs), "No logs should be captured when debug mode is disabled")
	})
}
