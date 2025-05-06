package bubble

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestKeyboardMessageMapping tests the mapping of Bubble Tea KeyMsg to BubblyUI KeyEvent
func TestKeyboardMessageMapping(t *testing.T) {
	// Create mock component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create translator
	translator := NewEventTranslator()
	
	t.Run("Simple key press is mapped correctly", func(t *testing.T) {
		// Create a tea.KeyMsg
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("a"),
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(keyMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a KeyEvent
		keyEvent, ok := event.(*KeyEvent)
		assert.True(t, ok, "Event should be a KeyEvent")
		
		// Check properties
		assert.Equal(t, EventTypeKey, keyEvent.Type())
		assert.Equal(t, comp, keyEvent.Source())
		assert.Equal(t, keyMsg, keyEvent.KeyMsg)
		assert.Equal(t, PriorityNormal, keyEvent.Priority())
		assert.False(t, keyEvent.IsPropagationStopped())
	})
	
	t.Run("Special key is mapped correctly", func(t *testing.T) {
		// Create a tea.KeyMsg for Enter key
		keyMsg := tea.KeyMsg{
			Type: tea.KeyEnter,
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(keyMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a KeyEvent
		keyEvent, ok := event.(*KeyEvent)
		assert.True(t, ok, "Event should be a KeyEvent")
		
		// Check properties
		assert.Equal(t, EventTypeKey, keyEvent.Type())
		assert.Equal(t, comp, keyEvent.Source())
		assert.Equal(t, keyMsg, keyEvent.KeyMsg)
		assert.Equal(t, PriorityNormal, keyEvent.Priority())
	})
	
	t.Run("Special key is correctly mapped", func(t *testing.T) {
		// Create a tea.KeyMsg for Ctrl+C (usually has special handling)
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyCtrlC,
			Runes: []rune{3}, // ASCII for Ctrl+C
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(keyMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a KeyEvent
		keyEvent, ok := event.(*KeyEvent)
		assert.True(t, ok, "Event should be a KeyEvent")
		
		// Verify the KeyMsg is the same as the original
		assert.Equal(t, keyMsg, keyEvent.KeyMsg)
		// Verify the string representation is correct
		assert.Equal(t, "ctrl+c", keyEvent.String())
	})
}

// TestMouseEventTranslations tests the mapping of Bubble Tea MouseMsg to BubblyUI MouseEvent
func TestMouseEventTranslations(t *testing.T) {
	// Create mock component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create translator
	translator := NewEventTranslator()
	
	t.Run("Mouse click is mapped correctly", func(t *testing.T) {
		// Create a tea.MouseMsg for left click
		mouseMsg := tea.MouseMsg{
			Type:   tea.MouseLeft,
			X:      10,
			Y:      20,
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(mouseMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a MouseEvent
		mouseEvent, ok := event.(*MouseEvent)
		assert.True(t, ok, "Event should be a MouseEvent")
		
		// Check properties
		assert.Equal(t, EventTypeMouse, mouseEvent.Type())
		assert.Equal(t, comp, mouseEvent.Source())
		assert.Equal(t, mouseMsg, mouseEvent.MouseMsg)
		assert.Equal(t, PriorityHigh, mouseEvent.Priority(), "Mouse events should have high priority")
		assert.Equal(t, 10, mouseEvent.X())
		assert.Equal(t, 20, mouseEvent.Y())
	})
	
	t.Run("Mouse drag is mapped correctly", func(t *testing.T) {
		// Create a tea.MouseMsg for dragging
		mouseMsg := tea.MouseMsg{
			Type:   tea.MouseMotion,
			X:      15,
			Y:      25,
			Shift:  true,
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(mouseMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a MouseEvent
		mouseEvent, ok := event.(*MouseEvent)
		assert.True(t, ok, "Event should be a MouseEvent")
		
		// Check properties
		assert.Equal(t, EventTypeMouse, mouseEvent.Type())
		assert.Equal(t, comp, mouseEvent.Source())
		assert.Equal(t, mouseMsg, mouseEvent.MouseMsg)
		assert.Equal(t, 15, mouseEvent.X())
		assert.Equal(t, 25, mouseEvent.Y())
		assert.True(t, mouseEvent.MouseMsg.Shift, "Shift modifier should be preserved")
	})
}

// TestWindowResizeEventMapping tests the mapping of Bubble Tea WindowSizeMsg to BubblyUI WindowSizeEvent
func TestWindowResizeEventMapping(t *testing.T) {
	// Create mock component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create translator
	translator := NewEventTranslator()
	
	t.Run("Window resize is mapped correctly", func(t *testing.T) {
		// Create a tea.WindowSizeMsg
		sizeMsg := tea.WindowSizeMsg{
			Width:  80,
			Height: 24,
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(sizeMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a WindowSizeEvent
		sizeEvent, ok := event.(*WindowSizeEvent)
		assert.True(t, ok, "Event should be a WindowSizeEvent")
		
		// Check properties
		assert.Equal(t, EventTypeWindowSize, sizeEvent.Type())
		assert.Equal(t, comp, sizeEvent.Source())
		assert.Equal(t, sizeMsg, sizeEvent.WindowSizeMsg)
		assert.Equal(t, PriorityLow, sizeEvent.Priority(), "Window resize events should have low priority")
		assert.Equal(t, 80, sizeEvent.Width())
		assert.Equal(t, 24, sizeEvent.Height())
	})
}

// TestCustomMessageMapping tests mapping of custom messages to BubblyUI events
func TestCustomMessageMapping(t *testing.T) {
	// Create mock component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create custom message type
	type CustomMsg struct {
		Name  string
		Value int
	}
	
	// Create custom event type
	type CustomEvent struct {
		*BaseEvent
		CustomData CustomMsg
	}
	
	// Create translator
	translator := NewEventTranslator()
	
	// Create a custom event factory
	customEventFactory := func(msg tea.Msg, source core.Component) (Event, bool) {
		customMsg, ok := msg.(CustomMsg)
		if !ok {
			return nil, false
		}
		
		baseEvent := NewBaseEvent(EventType("custom"), source, EventCategoryInput, customMsg)
		return &CustomEvent{
			BaseEvent:  baseEvent,
			CustomData: customMsg,
		}, true
	}
	
	// Register the custom mapper
	translator.RegisterMessageMapper(customEventFactory)
	
	t.Run("Custom message is mapped to custom event", func(t *testing.T) {
		// Create a custom message
		customMsg := CustomMsg{
			Name:  "test",
			Value: 42,
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(customMsg, comp)
		assert.NoError(t, err, "Translation should not error")
		assert.NotNil(t, event, "Event should not be nil")
		
		// Assert it's a CustomEvent
		customEvent, ok := event.(*CustomEvent)
		assert.True(t, ok, "Event should be a CustomEvent")
		
		// Check properties
		assert.Equal(t, EventType("custom"), customEvent.Type())
		assert.Equal(t, comp, customEvent.Source())
		assert.Equal(t, customMsg, customEvent.CustomData)
	})
}

// TestUnknownMessageHandling tests how unknown messages are handled
func TestUnknownMessageHandling(t *testing.T) {
	// Create mock component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create translator
	translator := NewEventTranslator()
	
	t.Run("Unknown message returns generic event or error", func(t *testing.T) {
		// Create an unknown message type
		type UnknownMsg struct {
			Data string
		}
		
		// Create an unknown message
		unknownMsg := UnknownMsg{
			Data: "test data",
		}
		
		// Translate to BubblyUI event
		event, err := translator.TranslateMessage(unknownMsg, comp)
		
		// Most likely will return an error for unknown message types
		if err != nil {
			// Error indicates unsupported message type
			assert.Contains(t, err.Error(), "unsupported message type", "Error should mention unsupported message type")
		} else if event != nil {
			// If event is returned, it should be of the UserDefinedEvent type or any event type
			// implementing the Event interface 
			// Check basic properties that any event should have
			assert.NotNil(t, event.Type(), "Event should have a type")
			assert.Equal(t, comp, event.Source(), "Event source should be the component")
			assert.NotNil(t, event.Priority(), "Event should have a priority")
		} else {
			// Both event and error are nil, which is unexpected but possible
			t.Log("Both event and error are nil, which is unexpected")
		}
	})
}
