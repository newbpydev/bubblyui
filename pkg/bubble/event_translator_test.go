package bubble

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestMessageToEventConversion validates that Bubble Tea messages are
// correctly converted into appropriate BubblyUI events.
func TestMessageToEventConversion(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	translator := NewEventTranslator()

	t.Run("KeyMsg is converted to KeyEvent", func(t *testing.T) {
		// Create a tea.KeyMsg
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("a"),
			Alt:   false,
		}

		// Convert the message to an event
		event, err := translator.TranslateMessage(keyMsg, mockComp)
		
		// Verify conversion worked
		assert.NoError(t, err)
		assert.NotNil(t, event)
		
		// Verify it's the right type
		keyEvent, ok := event.(*KeyEvent)
		assert.True(t, ok)
		assert.Equal(t, EventTypeKey, keyEvent.BaseEvent.Type())
		assert.Equal(t, keyMsg, keyEvent.KeyMsg)
	})

	t.Run("MouseMsg is converted to MouseEvent", func(t *testing.T) {
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

		// Convert the message to an event
		event, err := translator.TranslateMessage(mouseMsg, mockComp)
		
		// Verify conversion worked
		assert.NoError(t, err)
		assert.NotNil(t, event)
		
		// Verify it's the right type
		mouseEvent, ok := event.(*MouseEvent)
		assert.True(t, ok)
		assert.Equal(t, EventTypeMouse, mouseEvent.BaseEvent.Type())
		assert.Equal(t, mouseMsg, mouseEvent.MouseMsg)
	})

	t.Run("WindowSizeMsg is converted to WindowSizeEvent", func(t *testing.T) {
		// Create a tea.WindowSizeMsg
		windowSizeMsg := tea.WindowSizeMsg{
			Width:  80,
			Height: 24,
		}

		// Convert the message to an event
		event, err := translator.TranslateMessage(windowSizeMsg, mockComp)
		
		// Verify conversion worked
		assert.NoError(t, err)
		assert.NotNil(t, event)
		
		// Verify it's the right type
		windowSizeEvent, ok := event.(*WindowSizeEvent)
		assert.True(t, ok)
		assert.Equal(t, EventTypeWindowSize, windowSizeEvent.BaseEvent.Type())
		assert.Equal(t, windowSizeMsg, windowSizeEvent.WindowSizeMsg)
	})

	t.Run("Unknown message type returns error", func(t *testing.T) {
		// Create a custom message type
		type CustomMsg struct{ Value string }
		customMsg := CustomMsg{Value: "test"}

		// Attempt to convert the message to an event
		event, err := translator.TranslateMessage(customMsg, mockComp)
		
		// Verify appropriate error is returned
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "unsupported message type")
	})
}

// TestCustomMessageMapper tests registering and using custom message mappers.
func TestCustomMessageMapper(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	translator := NewEventTranslator()

	t.Run("Register and use custom message mapper", func(t *testing.T) {
		// Define a custom message type
		type CustomMsg struct{ Value string }
		
		// Register a custom mapper for this message type
		translator.RegisterMessageMapper(
			func(msg tea.Msg, source core.Component) (Event, bool) {
				if customMsg, ok := msg.(CustomMsg); ok {
					return NewCustomEvent(source, "customEvent", EventCategorySystem, customMsg.Value), true
				}
				return nil, false
			},
		)

		// Create and translate a custom message
		customMsg := CustomMsg{Value: "test-value"}
		event, err := translator.TranslateMessage(customMsg, mockComp)
		
		// Verify conversion worked
		assert.NoError(t, err)
		assert.NotNil(t, event)
		
		// Verify it's the correct type and has the right data
		customEvent, ok := event.(*UserDefinedEvent)
		assert.True(t, ok)
		assert.Equal(t, "customEvent", string(customEvent.BaseEvent.Type()))
		assert.Equal(t, "test-value", customEvent.Data)
	})

	t.Run("Multiple mappers are tried in order", func(t *testing.T) {
		// Define two custom message types
		type CustomMsg1 struct{ Value string }
		type CustomMsg2 struct{ Value string }
		
		// Create a new translator to start fresh
		translator := NewEventTranslator()
		
		// Register first mapper
		translator.RegisterMessageMapper(
			func(msg tea.Msg, source core.Component) (Event, bool) {
				if customMsg, ok := msg.(CustomMsg1); ok {
					return NewCustomEvent(source, "customEvent1", EventCategorySystem, customMsg.Value), true
				}
				return nil, false
			},
		)

		// Register second mapper
		translator.RegisterMessageMapper(
			func(msg tea.Msg, source core.Component) (Event, bool) {
				if customMsg, ok := msg.(CustomMsg2); ok {
					return NewCustomEvent(source, "customEvent2", EventCategorySystem, customMsg.Value), true
				}
				return nil, false
			},
		)

		// Test first mapper
		msg1 := CustomMsg1{Value: "test1"}
		event1, err := translator.TranslateMessage(msg1, mockComp)
		assert.NoError(t, err)
		assert.Equal(t, "customEvent1", string(event1.(*UserDefinedEvent).BaseEvent.Type()))
		
		// Test second mapper
		msg2 := CustomMsg2{Value: "test2"}
		event2, err := translator.TranslateMessage(msg2, mockComp)
		assert.NoError(t, err)
		assert.Equal(t, "customEvent2", string(event2.(*UserDefinedEvent).BaseEvent.Type()))
	})
}

// TestBidirectionalMapping tests converting events back to messages.
func TestBidirectionalMapping(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	translator := NewEventTranslator()

	t.Run("KeyEvent converts back to KeyMsg", func(t *testing.T) {
		// Create a tea.KeyMsg
		originalMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("a"),
			Alt:   false,
		}

		// Convert message to event
		event, err := translator.TranslateMessage(originalMsg, mockComp)
		assert.NoError(t, err)
		
		// Convert event back to message
		msg, err := translator.TranslateEvent(event)
		assert.NoError(t, err)
		
		// Verify it's the same message
		keyMsg, ok := msg.(tea.KeyMsg)
		assert.True(t, ok)
		assert.Equal(t, originalMsg, keyMsg)
	})

	t.Run("MouseEvent converts back to MouseMsg", func(t *testing.T) {
		// Create a tea.MouseMsg
		originalMsg := tea.MouseMsg{
			Type:   tea.MouseLeft,
			Button: tea.MouseButtonLeft,
			X:      10,
			Y:      20,
			Alt:    true,
			Ctrl:   false,
			Shift:  true,
		}

		// Convert message to event
		event, err := translator.TranslateMessage(originalMsg, mockComp)
		assert.NoError(t, err)
		
		// Convert event back to message
		msg, err := translator.TranslateEvent(event)
		assert.NoError(t, err)
		
		// Verify it's the same message
		mouseMsg, ok := msg.(tea.MouseMsg)
		assert.True(t, ok)
		assert.Equal(t, originalMsg, mouseMsg)
	})

	t.Run("Unknown event type returns error", func(t *testing.T) {
		// Create a mock event that doesn't have a standard conversion
		mockEvent := struct{ Event }{}
		
		// Attempt to convert the event to a message
		msg, err := translator.TranslateEvent(mockEvent)
		
		// Verify appropriate error is returned
		assert.Error(t, err)
		assert.Nil(t, msg)
		assert.Contains(t, err.Error(), "unsupported event type")
	})
}

// TestEventNormalization tests that events are normalized during conversion.
func TestEventNormalization(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	translator := NewEventTranslator()

	t.Run("Key events are normalized", func(t *testing.T) {
		// Register a normalizer for key events
		translator.RegisterEventNormalizer(func(event Event) Event {
			if keyEvent, ok := event.(*KeyEvent); ok {
				// Make all key events uppercase for testing
				if len(keyEvent.KeyMsg.Runes) > 0 {
					upper := []rune(string(keyEvent.KeyMsg.Runes))
					for i := range upper {
						if upper[i] >= 'a' && upper[i] <= 'z' {
							upper[i] = upper[i] - 32 // Convert to uppercase
						}
					}
					keyEvent.KeyMsg.Runes = upper
				}
				return keyEvent
			}
			return event
		})

		// Create a lowercase key message
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("hello"),
			Alt:   false,
		}

		// Convert and normalize
		event, err := translator.TranslateMessage(keyMsg, mockComp)
		assert.NoError(t, err)
		
		// Verify normalization worked
		keyEvent, ok := event.(*KeyEvent)
		assert.True(t, ok)
		assert.Equal(t, "HELLO", string(keyEvent.KeyMsg.Runes))
	})
}

// TestDefaultMappings tests that the default mappings work correctly.
func TestDefaultMappings(t *testing.T) {
	// Create a mock component for testing
	mockComp := NewMockComponent("test-component")
	translator := NewEventTranslator()

	// Test that all standard Bubble Tea messages have default mappings
	t.Run("All standard messages have default mappings", func(t *testing.T) {
		msgs := []tea.Msg{
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
			tea.MouseMsg{Type: tea.MouseLeft, X: 10, Y: 20},
			tea.WindowSizeMsg{Width: 80, Height: 24},
		}

		for _, msg := range msgs {
			event, err := translator.TranslateMessage(msg, mockComp)
			assert.NoError(t, err)
			assert.NotNil(t, event)
		}
	})
}
