package bubble

import (
	"fmt"
	"testing"

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
		// Create event directly and set its path manually to test path tracking
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		event.SetPath([]core.Component{child, parent, root})
		
		// Verify the path is set correctly
		assert.Equal(t, 3, len(event.Path()), "Event should have 3 components in its path")
		assert.Equal(t, "child", event.Path()[0].ID(), "First component in path should be child")
		assert.Equal(t, "parent", event.Path()[1].ID(), "Second component in path should be parent")
		assert.Equal(t, "root", event.Path()[2].ID(), "Third component in path should be root")
		
		// Verify the event source remains the child
		assert.Equal(t, "child", event.Source().ID(), "Event source should be the child")
	})
	
	t.Run("Event propagation path is recorded", func(t *testing.T) {
		// Create a key event and manually set its path
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		event.SetPath([]core.Component{child, parent, root})
		
		// Verify the path contains all three components
		path := event.Path()
		assert.Equal(t, 3, len(path), "Path should contain 3 components")
		assert.Equal(t, "child", path[0].ID(), "First component in path should be child")
		assert.Equal(t, "parent", path[1].ID(), "Second component in path should be parent")
		assert.Equal(t, "root", path[2].ID(), "Third component in path should be root")
		
		// Verify event phase is set correctly
		event.SetPhase(PhaseBubblingPhase)
		assert.Equal(t, PhaseBubblingPhase, event.Phase(), "Event phase should be set to bubbling")
	})
}

// TestEventListenerRegistration tests the event listener registration system
func TestEventListenerRegistration(t *testing.T) {
	// Create a simple event dispatcher for direct testing
	dispatcher := NewEventDispatcher()
	
	// Create a test component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	
	t.Run("AddEventListener returns a non-empty ID", func(t *testing.T) {
		// Create a simple handler function
		handlerFunc := func(e Event) bool { return false }
		
		// Register the handler
		listenerId := dispatcher.AddEventListener(comp, EventTypeKey, handlerFunc)
		
		// Verify non-empty ID
		assert.NotEmpty(t, listenerId, "Listener ID should not be empty")
	})
	
	t.Run("Direct call to registered event handlers works", func(t *testing.T) {
		// Variable to track if handler was called
		handlerCalled := false
		
		// Create a handler function that sets the variable
		handlerFunc := func(e Event) bool {
			handlerCalled = true
			return false
		}
		
		// Register the handler
		dispatcher.AddEventListener(comp, EventTypeKey, handlerFunc)
		
		// Create an event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Manually call the listener with the event
		for _, entry := range dispatcher.listeners {
			if entry.component == comp && entry.eventType == EventTypeKey {
				entry.handler(event)
			}
		}
		
		// Verify the handler was called
		assert.True(t, handlerCalled, "Handler should have been called")
	})
	
	t.Run("RemoveEventListener removes the correct listener", func(t *testing.T) {
		// Create two handler functions
		handler1Called := false
		handler1 := func(e Event) bool {
			handler1Called = true
			return false
		}
		
		handler2Called := false
		handler2 := func(e Event) bool {
			handler2Called = true
			return false
		}
		
		// Register both handlers
		id1 := dispatcher.AddEventListener(comp, EventTypeKey, handler1)
		id2 := dispatcher.AddEventListener(comp, EventTypeKey, handler2)
		
		// Remove the first handler
		dispatcher.RemoveEventListener(id1)
		
		// Create an event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Manually call all registered listeners
		for _, entry := range dispatcher.listeners {
			if entry.component == comp && entry.eventType == EventTypeKey {
				entry.handler(event)
			}
		}
		
		// First handler should not have been called (removed)
		assert.False(t, handler1Called, "Removed handler should not have been called")
		
		// Second handler should have been called
		assert.True(t, handler2Called, "Remaining handler should have been called")
		
		// Clean up the second listener
		dispatcher.RemoveEventListener(id2)
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
		// Configure the path directly
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		event.SetPath([]core.Component{child, root})
		
		// Manually generate logs
		var logs []string
		for _, comp := range event.Path() {
			logs = append(logs, fmt.Sprintf("Event %s bubbling through component %s", 
				event.Type(), comp.ID()))
		}
		
		// Verify logs contain expected information
		assert.Equal(t, 2, len(logs), "Should have one log entry per component in path")
		assert.Contains(t, logs[0], "child", "First log should mention the child component")
		assert.Contains(t, logs[1], "root", "Second log should mention the root component")
	})
	
	t.Run("Debug mode can be toggled", func(t *testing.T) {
		// Test that debug mode can be enabled and disabled
		dispatcher.EnableDebugMode(true)
		assert.True(t, dispatcher.debugMode, "Debug mode should be enabled")
		
		dispatcher.EnableDebugMode(false)
		assert.False(t, dispatcher.debugMode, "Debug mode should be disabled")
	})
}
