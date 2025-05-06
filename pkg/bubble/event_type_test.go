package bubble

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockComponent is a mock implementation of the core.Component interface for testing.
type MockComponent struct {
	mock.Mock
	id string
	children []core.Component
}

// NewMockComponent creates a new mock component with a given ID
func NewMockComponent(id string) *MockComponent {
	return &MockComponent{id: id, children: make([]core.Component, 0)}
}

// Initialize implements core.Component
func (m *MockComponent) Initialize() error {
	args := m.Called()
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

// Update implements core.Component
func (m *MockComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	args := m.Called(msg)
	if len(args) > 1 {
		return args.Get(0).(tea.Cmd), args.Error(1)
	}
	return nil, nil
}

// Render implements core.Component
func (m *MockComponent) Render() string {
	args := m.Called()
	if len(args) > 0 {
		return args.String(0)
	}
	return ""
}

// Dispose implements core.Component
func (m *MockComponent) Dispose() error {
	args := m.Called()
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

// ID implements core.Component
func (m *MockComponent) ID() string {
	args := m.Called()
	if len(args) > 0 {
		return args.String(0)
	}
	return m.id
}

// AddChild implements core.Component
func (m *MockComponent) AddChild(child core.Component) {
	m.Called(child)
	m.children = append(m.children, child)
}

// RemoveChild implements core.Component
func (m *MockComponent) RemoveChild(id string) bool {
	m.Called(id)
	for i, child := range m.children {
		if child.ID() == id {
			m.children = append(m.children[:i], m.children[i+1:]...)
			return true
		}
	}
	return false
}

// Children implements core.Component
func (m *MockComponent) Children() []core.Component {
	m.Called()
	return m.children
}

// TestEventInterface validates that the Event interface provides the required functionality
// for a comprehensive event system that can handle Bubble Tea messages.
func TestEventInterface(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")

	t.Run("Event interface provides type identification", func(t *testing.T) {
		// Create a base event with a specific type
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		
		// Verify the type is correctly returned
		assert.Equal(t, EventTypeKey, event.Type())
	})

	t.Run("Event interface provides access to source component", func(t *testing.T) {
		// Create a base event with our mock component
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		
		// Verify the source component is correctly returned
		assert.Equal(t, mockComp, event.Source())
	})

	t.Run("Event interface provides metadata access", func(t *testing.T) {
		// Create a base event
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		
		// Verify metadata fields are accessible
		assert.NotZero(t, event.Timestamp())
		assert.Equal(t, EventCategoryInput, event.Category())
		assert.Equal(t, PhaseAtTarget, event.Phase())
		
		// Modify and check phase
		event.SetPhase(PhaseBubblingPhase)
		assert.Equal(t, PhaseBubblingPhase, event.Phase())
	})

	t.Run("Event interface supports prevention of default behavior", func(t *testing.T) {
		// Create a base event
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		
		// Initially, default is not prevented
		assert.False(t, event.IsDefaultPrevented())
		
		// Prevent default and check
		event.PreventDefault()
		assert.True(t, event.IsDefaultPrevented())
	})

	t.Run("Event interface supports stopping propagation", func(t *testing.T) {
		// Create a base event
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		
		// Initially, propagation is not stopped
		assert.False(t, event.IsPropagationStopped())
		assert.False(t, event.IsImmediatePropagationStopped())
		
		// Stop propagation and check
		event.StopPropagation()
		assert.True(t, event.IsPropagationStopped())
		assert.False(t, event.IsImmediatePropagationStopped())
		
		// Create a new event for testing immediate stop
		event = NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		event.StopImmediatePropagation()
		assert.True(t, event.IsPropagationStopped())
		assert.True(t, event.IsImmediatePropagationStopped())
	})

	t.Run("Event interface supports component path", func(t *testing.T) {
		// Create a base event
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		
		// Default path should contain just the source component
		assert.Len(t, event.Path(), 1)
		assert.Equal(t, mockComp, event.Path()[0])
		
		// Create additional mock components for path
		parentComp := new(MockComponent)
		grandparentComp := new(MockComponent)
		
		// Set custom path and verify
		path := []core.Component{mockComp, parentComp, grandparentComp}
		event.SetPath(path)
		assert.Equal(t, path, event.Path())
		assert.Len(t, event.Path(), 3)
	})
}

// TestStandardEventTypes validates that standard event types (keyboard, mouse, etc.)
// are properly implemented and handle the corresponding Bubble Tea messages.
func TestStandardEventTypes(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")

	t.Run("KeyEvent properly wraps tea.KeyMsg", func(t *testing.T) {
		// Create a tea.KeyMsg
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("a"),
			Alt:   false,
		}
		
		// Create a KeyEvent
		keyEvent := NewKeyEvent(mockComp, keyMsg)
		
		// Verify the event properties
		assert.Equal(t, EventTypeKey, keyEvent.Type())
		assert.Equal(t, EventCategoryInput, keyEvent.Category())
		assert.Equal(t, mockComp, keyEvent.Source())
		assert.Equal(t, keyMsg, keyEvent.KeyMsg)
		assert.Equal(t, keyMsg, keyEvent.RawMessage())
		assert.Equal(t, "a", keyEvent.String())
		assert.Equal(t, []rune("a"), keyEvent.Runes())
	})

	t.Run("MouseEvent properly wraps tea.MouseMsg", func(t *testing.T) {
		// Create a tea.MouseMsg
		mouseMsg := tea.MouseMsg{
			Type:   tea.MouseLeft,
			Button: tea.MouseButtonLeft,
			X:      10,
			Y:      20,
			Alt:    true,
			Ctrl:   false,
			Shift:  true,
		}
		
		// Create a MouseEvent
		mouseEvent := NewMouseEvent(mockComp, mouseMsg)
		
		// Verify the event properties
		assert.Equal(t, EventTypeMouse, mouseEvent.Type())
		assert.Equal(t, EventCategoryInput, mouseEvent.Category())
		assert.Equal(t, mockComp, mouseEvent.Source())
		assert.Equal(t, mouseMsg, mouseEvent.MouseMsg)
		assert.Equal(t, mouseMsg, mouseEvent.RawMessage())
		assert.Equal(t, tea.MouseLeft, mouseEvent.MouseType())
		assert.Equal(t, tea.MouseButtonLeft, mouseEvent.Button())
		assert.Equal(t, 10, mouseEvent.X())
		assert.Equal(t, 20, mouseEvent.Y())
		assert.True(t, mouseEvent.Alt())
		assert.False(t, mouseEvent.Ctrl())
		assert.True(t, mouseEvent.Shift())
	})

	t.Run("WindowSizeEvent properly wraps tea.WindowSizeMsg", func(t *testing.T) {
		// Create a tea.WindowSizeMsg
		windowSizeMsg := tea.WindowSizeMsg{
			Width:  80,
			Height: 24,
		}
		
		// Create a WindowSizeEvent
		windowSizeEvent := NewWindowSizeEvent(mockComp, windowSizeMsg)
		
		// Verify the event properties
		assert.Equal(t, EventTypeWindowSize, windowSizeEvent.Type())
		assert.Equal(t, EventCategoryUI, windowSizeEvent.Category())
		assert.Equal(t, mockComp, windowSizeEvent.Source())
		assert.Equal(t, windowSizeMsg, windowSizeEvent.WindowSizeMsg)
		assert.Equal(t, windowSizeMsg, windowSizeEvent.RawMessage())
		assert.Equal(t, 80, windowSizeEvent.Width())
		assert.Equal(t, 24, windowSizeEvent.Height())
	})
}

// TestCustomEventTypes validates that custom event types can be created and used.
func TestCustomEventTypes(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	
	// Define a custom event type
	const EventTypeFormSubmit EventType = "formSubmit"

	t.Run("Custom event can be created", func(t *testing.T) {
		// Create custom event data
		type FormData struct {
			Name  string
			Email string
		}
		
		formData := FormData{
			Name:  "Test User",
			Email: "test@example.com",
		}
		
		// Create a custom event
		customEvent := NewCustomEvent(mockComp, EventTypeFormSubmit, EventCategoryUI, formData)
		
		// Verify the event properties
		assert.Equal(t, EventTypeFormSubmit, customEvent.BaseEvent.Type())
		assert.Equal(t, EventCategoryUI, customEvent.BaseEvent.Category())
		assert.Equal(t, mockComp, customEvent.BaseEvent.Source())
		
		// Verify the event data
		data, ok := customEvent.EventData().(FormData)
		assert.True(t, ok)
		assert.Equal(t, "Test User", data.Name)
		assert.Equal(t, "test@example.com", data.Email)
	})

	t.Run("Custom event is identifiable by type", func(t *testing.T) {
		// Create multiple custom events with different types
		const (
			EventTypeUserAction EventType = "userAction"
			EventTypeSystemEvent EventType = "systemEvent"
		)
		
		customEvent1 := NewCustomEvent(mockComp, EventTypeUserAction, EventCategoryUI, "user clicked save")
		customEvent2 := NewCustomEvent(mockComp, EventTypeSystemEvent, EventCategorySystem, "system notification")
		customEvent3 := NewCustomEvent(mockComp, EventTypeFormSubmit, EventCategoryUI, "form submitted")
		
		// Verify each event has the correct type
		assert.Equal(t, EventTypeUserAction, customEvent1.BaseEvent.Type())
		assert.Equal(t, EventTypeSystemEvent, customEvent2.BaseEvent.Type())
		assert.Equal(t, EventTypeFormSubmit, customEvent3.BaseEvent.Type())
		
		// Verify category assignment
		assert.Equal(t, EventCategoryUI, customEvent1.BaseEvent.Category())
		assert.Equal(t, EventCategorySystem, customEvent2.BaseEvent.Category())
		assert.Equal(t, EventCategoryUI, customEvent3.BaseEvent.Category())
	})
}

// TestEventMetadata tests the event metadata structure.
func TestEventMetadata(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	
	t.Run("Event metadata includes timestamp", func(t *testing.T) {
		// Record time before and after event creation
		before := time.Now()
		event := NewBaseEvent(EventTypeKey, mockComp, EventCategoryInput, nil)
		after := time.Now()
		
		// Verify the timestamp is between before and after
		timestamp := event.Timestamp()
		assert.True(t, timestamp.After(before) || timestamp.Equal(before))
		assert.True(t, timestamp.Before(after) || timestamp.Equal(after))
	})

	t.Run("Event metadata includes component path", func(t *testing.T) {
		// Create mock components for a component hierarchy
		child := NewMockComponent("child")
		parent := NewMockComponent("parent")
		grandparent := NewMockComponent("grandparent")
		
		// Create an event and set up a component path
		event := NewBaseEvent(EventTypeKey, child, EventCategoryInput, nil)
		
		// Initially path should just contain the source
		assert.Equal(t, []core.Component{child}, event.Path())
		
		// Set a full component path and verify
		path := []core.Component{child, parent, grandparent}
		event.SetPath(path)
		assert.Equal(t, path, event.Path())
		assert.Equal(t, 3, len(event.Path()))
		assert.Equal(t, child, event.Path()[0])
		assert.Equal(t, parent, event.Path()[1])
		assert.Equal(t, grandparent, event.Path()[2])
	})
}

// TestEventTypeHierarchy tests the categorization and hierarchy of event types.
func TestEventTypeHierarchy(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	
	t.Run("Input events are categorized properly", func(t *testing.T) {
		// Create keyboard and mouse events
		keyEvent := NewKeyEvent(mockComp, tea.KeyMsg{})
		mouseEvent := NewMouseEvent(mockComp, tea.MouseMsg{})
		
		// Verify they have the input category
		assert.Equal(t, EventCategoryInput, keyEvent.Category())
		assert.Equal(t, EventCategoryInput, mouseEvent.Category())
	})

	t.Run("UI events are categorized properly", func(t *testing.T) {
		// Create window size event
		windowEvent := NewWindowSizeEvent(mockComp, tea.WindowSizeMsg{})
		
		// Verify it has the UI category
		assert.Equal(t, EventCategoryUI, windowEvent.Category())
	})

	t.Run("System events are categorized properly", func(t *testing.T) {
		// Create a system event
		systemEvent := NewCustomEvent(mockComp, "systemError", EventCategorySystem, "error message")
		
		// Verify it has the system category
		assert.Equal(t, EventCategorySystem, systemEvent.BaseEvent.Category())
	})

	t.Run("Custom events can be categorized", func(t *testing.T) {
		// Create custom events with different categories
		customInput := NewCustomEvent(mockComp, "customInput", EventCategoryInput, "input data")
		customUI := NewCustomEvent(mockComp, "customUI", EventCategoryUI, "ui data")
		customSystem := NewCustomEvent(mockComp, "customSystem", EventCategorySystem, "system data")
		
		// Verify each has the correct category
		assert.Equal(t, EventCategoryInput, customInput.BaseEvent.Category())
		assert.Equal(t, EventCategoryUI, customUI.BaseEvent.Category())
		assert.Equal(t, EventCategorySystem, customSystem.BaseEvent.Category())
	})
}
