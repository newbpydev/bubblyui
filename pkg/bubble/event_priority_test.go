package bubble

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestEventPriorityLevels tests the creation and usage of event priority levels.
func TestEventPriorityLevels(t *testing.T) {
	// Create mock component
	comp := NewMockComponent("test-component")
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	comp.On("Called").Return()
	comp.On("AddChild", mock.Anything).Return()
	
	t.Run("Events have default priority", func(t *testing.T) {
		// Create standard events and check their default priorities
		keyEvent := NewKeyEvent(comp, tea.KeyMsg{})
		mouseEvent := NewMouseEvent(comp, tea.MouseMsg{})
		windowEvent := NewWindowSizeEvent(comp, tea.WindowSizeMsg{})
		
		// Verify all events have appropriate default priorities
		assert.Equal(t, PriorityNormal, keyEvent.Priority())
		assert.Equal(t, PriorityHigh, mouseEvent.Priority())   // Mouse events are high priority by default
		assert.Equal(t, PriorityLow, windowEvent.Priority())   // Window events are low priority by default
	})
	
	t.Run("Event priority can be changed", func(t *testing.T) {
		// Create event with default priority
		event := NewKeyEvent(comp, tea.KeyMsg{})
		assert.Equal(t, PriorityNormal, event.Priority())
		
		// Change priority
		event.SetPriority(PriorityHigh)
		assert.Equal(t, PriorityHigh, event.Priority())
		
		// Change to custom priority
		event.SetPriority(10) // Custom priority level
		assert.Equal(t, EventPriority(10), event.Priority())
	})
	
	t.Run("User-initiated events get priority boost", func(t *testing.T) {
		// Create event handler with priority boost
		handler := NewEventPriorityHandler()
		
		// Create a key event that represents user interaction
		// And set it to LowPriority initially to demonstrate the boost
		keyEvent := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		keyEvent.SetPriority(PriorityLow) // Explicitly set to low priority for the test
		
		// Before boost
		assert.Equal(t, PriorityLow, keyEvent.Priority())
		
		// Store the original event priority
		originalPriority := keyEvent.Priority()
		
		// Apply user interaction boost
		boostEvent := handler.BoostUserInitiatedEvent(keyEvent)
		
		// After boost - should be higher
		assert.Greater(t, boostEvent.Priority(), originalPriority)
		assert.Equal(t, PriorityHigh, boostEvent.Priority())
	})
}

// TestPriorityBasedEventQueue tests the event queue with priorities.
func TestPriorityBasedEventQueue(t *testing.T) {
	// Create mock component
	comp := NewMockComponent("test-component")
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	comp.On("Called").Return()
	comp.On("AddChild", mock.Anything).Return()
	
	t.Run("Events are processed in priority order", func(t *testing.T) {
		// Create event queue
		queue := NewPriorityEventQueue()
		
		// Create events with different priorities
		lowEvent := NewWindowSizeEvent(comp, tea.WindowSizeMsg{})
		lowEvent.SetPriority(PriorityLow)
		
		normalEvent := NewKeyEvent(comp, tea.KeyMsg{})
		normalEvent.SetPriority(PriorityNormal)
		
		highEvent := NewMouseEvent(comp, tea.MouseMsg{})
		highEvent.SetPriority(PriorityHigh)
		
		urgentEvent := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyCtrlC})
		urgentEvent.SetPriority(PriorityUrgent)
		
		// Add events to queue in random order
		queue.Enqueue(lowEvent)
		queue.Enqueue(normalEvent)
		queue.Enqueue(highEvent)
		queue.Enqueue(urgentEvent)
		
		// Verify events are processed in priority order
		assert.Equal(t, urgentEvent, queue.Dequeue())
		assert.Equal(t, highEvent, queue.Dequeue())
		assert.Equal(t, normalEvent, queue.Dequeue())
		assert.Equal(t, lowEvent, queue.Dequeue())
		
		// Queue should be empty now
		assert.True(t, queue.IsEmpty())
	})
	
	t.Run("Same priority events are processed in FIFO order", func(t *testing.T) {
		// Create event queue
		queue := NewPriorityEventQueue()
		
		// Create events with same priority
		event1 := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
		event1.SetPriority(PriorityNormal)
		
		event2 := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
		event2.SetPriority(PriorityNormal)
		
		event3 := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
		event3.SetPriority(PriorityNormal)
		
		// Add events to queue in order
		queue.Enqueue(event1)
		queue.Enqueue(event2)
		queue.Enqueue(event3)
		
		// Verify events are processed in FIFO order
		assert.Equal(t, event1, queue.Dequeue())
		assert.Equal(t, event2, queue.Dequeue())
		assert.Equal(t, event3, queue.Dequeue())
	})
}

// TestPriorityInheritance tests the inheritance of priority from parent events.
func TestPriorityInheritance(t *testing.T) {
	// Create mock component hierarchy
	parent := NewMockComponent("parent")
	child := NewMockComponent("child")
	
	parent.On("ID").Return("parent")
	parent.On("Children").Return([]core.Component{child})
	parent.On("Called").Return()
	parent.On("AddChild", mock.Anything).Return()
	
	child.On("ID").Return("child")
	child.On("Children").Return([]core.Component{})
	child.On("Called").Return()
	child.On("AddChild", mock.Anything).Return()
	
	t.Run("Child events inherit parent priority", func(t *testing.T) {
		// Create event handler
		handler := NewEventPriorityHandler()
		
		// Create parent event with high priority
		parentEvent := NewKeyEvent(parent, tea.KeyMsg{})
		parentEvent.SetPriority(PriorityHigh)
		
		// Create child event with normal priority
		childEvent := NewKeyEvent(child, tea.KeyMsg{})
		assert.Equal(t, PriorityNormal, childEvent.Priority())
		
		// Apply inheritance
		inheritedEvent := handler.ApplyPriorityInheritance(childEvent, parentEvent)
		
		// Child event should inherit parent's priority
		assert.Equal(t, parentEvent.Priority(), inheritedEvent.Priority())
	})
	
	t.Run("Inheritance only applies if child priority is lower", func(t *testing.T) {
		// Create event handler
		handler := NewEventPriorityHandler()
		
		// Create parent event with normal priority
		parentEvent := NewKeyEvent(parent, tea.KeyMsg{})
		parentEvent.SetPriority(PriorityNormal)
		
		// Create child event with high priority
		childEvent := NewKeyEvent(child, tea.KeyMsg{})
		childEvent.SetPriority(PriorityHigh)
		
		// Apply inheritance
		inheritedEvent := handler.ApplyPriorityInheritance(childEvent, parentEvent)
		
		// Child event should keep its higher priority
		assert.Equal(t, PriorityHigh, inheritedEvent.Priority())
	})
}

// TestEventThrottling tests throttling of low-priority events.
func TestEventThrottling(t *testing.T) {
	// Create mock component
	comp := NewMockComponent("test-component")
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	comp.On("Called").Return()
	comp.On("AddChild", mock.Anything).Return()
	
	t.Run("Low priority events are throttled", func(t *testing.T) {
		// Create throttler
		throttler := NewEventThrottler()
		
		// Create a low priority event type
		resizeEvent1 := NewWindowSizeEvent(comp, tea.WindowSizeMsg{Width: 100, Height: 80})
		resizeEvent1.SetPriority(PriorityLow)
		
		// First event should pass through
		assert.True(t, throttler.ShouldProcess(resizeEvent1))
		
		// Create a second similar event immediately
		resizeEvent2 := NewWindowSizeEvent(comp, tea.WindowSizeMsg{Width: 101, Height: 80})
		resizeEvent2.SetPriority(PriorityLow)
		
		// Second event should be throttled
		assert.False(t, throttler.ShouldProcess(resizeEvent2))
		
		// Wait for throttle duration to pass
		time.Sleep(150 * time.Millisecond)
		
		// Now the event should be processed
		assert.True(t, throttler.ShouldProcess(resizeEvent2))
	})
	
	t.Run("High priority events bypass throttling", func(t *testing.T) {
		// Create throttler
		throttler := NewEventThrottler()
		
		// Create low priority event and process it
		lowEvent := NewWindowSizeEvent(comp, tea.WindowSizeMsg{})
		lowEvent.SetPriority(PriorityLow)
		throttler.ShouldProcess(lowEvent)
		
		// Create high priority event
		highEvent := NewMouseEvent(comp, tea.MouseMsg{})
		highEvent.SetPriority(PriorityHigh)
		
		// High priority event should bypass throttling
		assert.True(t, throttler.ShouldProcess(highEvent))
	})
	
	t.Run("Throttling can be configured per event type", func(t *testing.T) {
		// Create throttler with custom configuration
		throttler := NewEventThrottler()
		
		// Configure throttling for resize events (500ms)
		throttler.ConfigureEventTypeThrottling(EventTypeWindowSize, 500*time.Millisecond)
		
		// Configure throttling for key events (100ms)
		throttler.ConfigureEventTypeThrottling(EventTypeKey, 100*time.Millisecond)
		
		// Process a key event
		keyEvent1 := NewKeyEvent(comp, tea.KeyMsg{})
		keyEvent1.SetPriority(PriorityLow)
		assert.True(t, throttler.ShouldProcess(keyEvent1))
		
		// Process another key event immediately - should be throttled
		keyEvent2 := NewKeyEvent(comp, tea.KeyMsg{})
		keyEvent2.SetPriority(PriorityLow)
		assert.False(t, throttler.ShouldProcess(keyEvent2))
		
		// Wait for key throttle to expire (just over 100ms)
		time.Sleep(110 * time.Millisecond)
		
		// Now it should process
		assert.True(t, throttler.ShouldProcess(keyEvent2))
		
		// Process a resize event
		resizeEvent1 := NewWindowSizeEvent(comp, tea.WindowSizeMsg{})
		resizeEvent1.SetPriority(PriorityLow)
		assert.True(t, throttler.ShouldProcess(resizeEvent1))
		
		// Wait longer than key throttle but less than resize throttle
		time.Sleep(200 * time.Millisecond)
		
		// Resize event should still be throttled
		resizeEvent2 := NewWindowSizeEvent(comp, tea.WindowSizeMsg{})
		resizeEvent2.SetPriority(PriorityLow)
		assert.False(t, throttler.ShouldProcess(resizeEvent2))
	})
}
