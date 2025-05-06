package bubble

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
)

// MessageMapper is a function that attempts to convert a tea.Msg to an Event.
// It returns the resulting Event and a boolean indicating whether the conversion was successful.
type MessageMapper func(msg tea.Msg, source core.Component) (Event, bool)

// EventNormalizer is a function that can modify an event for normalization.
// It returns the normalized event.
type EventNormalizer func(event Event) Event

// EventTranslator is responsible for converting Bubble Tea messages to BubblyUI events
// and vice versa. It supports custom message mappings and event normalization.
type EventTranslator struct {
	messageMappers   []MessageMapper
	eventNormalizers []EventNormalizer
}

// NewEventTranslator creates a new EventTranslator with default mappings.
func NewEventTranslator() *EventTranslator {
	translator := &EventTranslator{
		messageMappers:   make([]MessageMapper, 0),
		eventNormalizers: make([]EventNormalizer, 0),
	}
	
	// Register default message mappers
	translator.registerDefaultMappers()
	
	return translator
}

// registerDefaultMappers registers the default mappers for standard Bubble Tea messages.
func (t *EventTranslator) registerDefaultMappers() {
	// KeyMsg mapper
	t.RegisterMessageMapper(func(msg tea.Msg, source core.Component) (Event, bool) {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return NewKeyEvent(source, keyMsg), true
		}
		return nil, false
	})
	
	// MouseMsg mapper
	t.RegisterMessageMapper(func(msg tea.Msg, source core.Component) (Event, bool) {
		if mouseMsg, ok := msg.(tea.MouseMsg); ok {
			return NewMouseEvent(source, mouseMsg), true
		}
		return nil, false
	})
	
	// WindowSizeMsg mapper
	t.RegisterMessageMapper(func(msg tea.Msg, source core.Component) (Event, bool) {
		if windowSizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
			return NewWindowSizeEvent(source, windowSizeMsg), true
		}
		return nil, false
	})
}

// RegisterMessageMapper adds a custom message mapper to the translator.
// Mappers are tried in the order they are registered.
func (t *EventTranslator) RegisterMessageMapper(mapper MessageMapper) {
	t.messageMappers = append(t.messageMappers, mapper)
}

// RegisterEventNormalizer adds a custom event normalizer to the translator.
// Normalizers are applied in the order they are registered.
func (t *EventTranslator) RegisterEventNormalizer(normalizer EventNormalizer) {
	t.eventNormalizers = append(t.eventNormalizers, normalizer)
}

// TranslateMessage converts a Bubble Tea message to a BubblyUI event.
// It returns the event and any error that occurred during translation.
func (t *EventTranslator) TranslateMessage(msg tea.Msg, source core.Component) (Event, error) {
	// Try all registered mappers
	for _, mapper := range t.messageMappers {
		if event, ok := mapper(msg, source); ok {
			// Apply normalizers
			for _, normalizer := range t.eventNormalizers {
				event = normalizer(event)
			}
			return event, nil
		}
	}
	
	// No mapper found for this message type
	return nil, fmt.Errorf("unsupported message type: %T", msg)
}

// TranslateEvent converts a BubblyUI event back to a Bubble Tea message.
// It returns the message and any error that occurred during translation.
func (t *EventTranslator) TranslateEvent(event Event) (tea.Msg, error) {
	switch e := event.(type) {
	case *KeyEvent:
		return e.KeyMsg, nil
	case *MouseEvent:
		return e.MouseMsg, nil
	case *WindowSizeEvent:
		return e.WindowSizeMsg, nil
	case *UserDefinedEvent:
		// For custom events, use the raw message if available
		if msg, ok := e.BaseEvent.RawMessage().(tea.Msg); ok {
			return msg, nil
		}
		// Otherwise return the data
		return e.Data, nil
	default:
		return nil, fmt.Errorf("unsupported event type: %T", event)
	}
}

// GetMessageType returns a string representation of the message type.
// This is useful for debugging and logging.
func (t *EventTranslator) GetMessageType(msg tea.Msg) string {
	return fmt.Sprintf("%T", msg)
}
