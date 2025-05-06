package bubble

import (
	"github.com/newbpydev/bubblyui/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	*BaseEvent
	KeyMsg tea.KeyMsg
}

// NewKeyEvent creates a new keyboard event from a tea.KeyMsg.
func NewKeyEvent(source core.Component, keyMsg tea.KeyMsg) *KeyEvent {
	return &KeyEvent{
		BaseEvent: NewBaseEvent(EventTypeKey, source, EventCategoryInput, keyMsg),
		KeyMsg:    keyMsg,
	}
}

// String returns the key as a string.
func (e *KeyEvent) String() string {
	return e.KeyMsg.String()
}

// Runes returns the runes from the key press.
func (e *KeyEvent) Runes() []rune {
	return e.KeyMsg.Runes
}

// MouseEvent represents a mouse event.
type MouseEvent struct {
	*BaseEvent
	MouseMsg tea.MouseMsg
}

// NewMouseEvent creates a new mouse event from a tea.MouseMsg.
func NewMouseEvent(source core.Component, mouseMsg tea.MouseMsg) *MouseEvent {
	return &MouseEvent{
		BaseEvent: NewBaseEvent(EventTypeMouse, source, EventCategoryInput, mouseMsg),
		MouseMsg:  mouseMsg,
	}
}

// Type returns the mouse event type.
func (e *MouseEvent) MouseType() tea.MouseEventType {
	return e.MouseMsg.Type
}

// Button returns the mouse button that was pressed.
func (e *MouseEvent) Button() tea.MouseButton {
	return e.MouseMsg.Button
}

// X returns the X coordinate of the mouse event.
func (e *MouseEvent) X() int {
	return e.MouseMsg.X
}

// Y returns the Y coordinate of the mouse event.
func (e *MouseEvent) Y() int {
	return e.MouseMsg.Y
}

// Alt returns whether the Alt key was pressed during the mouse event.
func (e *MouseEvent) Alt() bool {
	return e.MouseMsg.Alt
}

// Ctrl returns whether the Ctrl key was pressed during the mouse event.
func (e *MouseEvent) Ctrl() bool {
	return e.MouseMsg.Ctrl
}

// Shift returns whether the Shift key was pressed during the mouse event.
func (e *MouseEvent) Shift() bool {
	return e.MouseMsg.Shift
}

// WindowSizeEvent represents a window size change event.
type WindowSizeEvent struct {
	*BaseEvent
	WindowSizeMsg tea.WindowSizeMsg
}

// NewWindowSizeEvent creates a new window size event from a tea.WindowSizeMsg.
func NewWindowSizeEvent(source core.Component, windowSizeMsg tea.WindowSizeMsg) *WindowSizeEvent {
	return &WindowSizeEvent{
		BaseEvent:     NewBaseEvent(EventTypeWindowSize, source, EventCategoryUI, windowSizeMsg),
		WindowSizeMsg: windowSizeMsg,
	}
}

// Width returns the width of the window.
func (e *WindowSizeEvent) Width() int {
	return e.WindowSizeMsg.Width
}

// Height returns the height of the window.
func (e *WindowSizeEvent) Height() int {
	return e.WindowSizeMsg.Height
}

// UserDefinedEvent represents a custom event type defined by the user.
type UserDefinedEvent struct {
	*BaseEvent
	Data interface{}
}

// NewCustomEvent creates a new custom event with the provided data.
func NewCustomEvent(source core.Component, eventType EventType, category EventType, data interface{}) *UserDefinedEvent {
	return &UserDefinedEvent{
		BaseEvent: NewBaseEvent(eventType, source, category, data),
		Data:      data,
	}
}

// EventData returns the data associated with this custom event.
func (e *UserDefinedEvent) EventData() interface{} {
	return e.Data
}
