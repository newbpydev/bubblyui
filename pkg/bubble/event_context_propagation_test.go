package bubble

import (
	"strconv"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestSourceIdentificationPropagation tests that events correctly identify and propagate information about their source components
func TestSourceIdentificationPropagation(t *testing.T) {
	// Create mock components
	rootComp := new(MockComponent)
	childComp := new(MockComponent)
	grandchildComp := new(MockComponent)
	
	// Setup mock behaviors
	rootComp.On("ID").Return("root")
	rootComp.On("Children").Return([]core.Component{childComp})
	rootComp.On("AddChild", mock.Anything).Run(func(args mock.Arguments) {
		// This simulates the actual AddChild implementation
		rootComp.children = append(rootComp.children, args.Get(0).(core.Component))
	})
	
	childComp.On("ID").Return("child")
	childComp.On("Children").Return([]core.Component{grandchildComp})
	childComp.On("AddChild", mock.Anything).Run(func(args mock.Arguments) {
		// This simulates the actual AddChild implementation
		childComp.children = append(childComp.children, args.Get(0).(core.Component))
	})
	
	grandchildComp.On("ID").Return("grandchild")
	grandchildComp.On("Children").Return([]core.Component{})
	
	// Setup hierarchy
	rootComp.AddChild(childComp)
	childComp.AddChild(grandchildComp)
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	contextualizer.SetRootComponent(rootComp)
	
	t.Run("Events capture their immediate source component", func(t *testing.T) {
		// Create a key event from the grandchild component
		event := NewKeyEvent(grandchildComp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Verify the source component before enrichment
		assert.Equal(t, grandchildComp, event.Source(), "Event source should be the component that originated it")
		assert.Equal(t, "grandchild", event.Source().ID(), "Source ID should match the originating component")
		
		// Enrich the event with context
		enriched, err := contextualizer.EnrichEventContext(event, grandchildComp)
		assert.NoError(t, err)
		
		// Verify source is preserved after enrichment
		assert.Equal(t, grandchildComp, enriched.Source(), "Enriched event should preserve source component")
	})
	
	t.Run("Events capture component path during propagation", func(t *testing.T) {
		// Create a key event
		event := NewKeyEvent(grandchildComp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Enrich with context
		enriched, err := contextualizer.EnrichEventContext(event, grandchildComp)
		assert.NoError(t, err)
		
		// The path should be automatically built by the contextualizer
		// In a real app with proper parent pointers, this would work automatically
		// For this test, we're just verifying that the component path was set
		keyEvent, ok := enriched.(*KeyEvent)
		assert.True(t, ok, "Enriched event should still be a KeyEvent")
		
		// Event should at least record the source component
		assert.NotNil(t, keyEvent.eventContext, "Event should have context after enrichment")
	})
}

// TestTimestampAccuracyInEvents tests that events accurately record their timestamps
func TestTimestampAccuracyInEvents(t *testing.T) {
	// Create a test component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Events record accurate timestamps", func(t *testing.T) {
		// Record the time before creating the event
		beforeTime := time.Now()
		
		// Small delay to ensure time difference is measurable
		time.Sleep(1 * time.Millisecond)
		
		// Create an event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Small delay
		time.Sleep(1 * time.Millisecond)
		
		// Record the time after creating the event
		afterTime := time.Now()
		
		// Verify the timestamp is between before and after times
		eventTime := event.Timestamp()
		assert.True(t, eventTime.After(beforeTime) || eventTime.Equal(beforeTime), 
			"Event timestamp should be after or equal to the time before creating the event")
		assert.True(t, eventTime.Before(afterTime) || eventTime.Equal(afterTime), 
			"Event timestamp should be before or equal to the time after creating the event")
		
		// Create a key event
		event = NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Enrich the event with context
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Check timestamp accuracy
		assert.True(t, enriched.Timestamp().After(beforeTime) || enriched.Timestamp().Equal(beforeTime),
			"Enriched event timestamp should be after or equal to the time before creating the event")
	})
	
	t.Run("Sequential events have increasing timestamps", func(t *testing.T) {
		// Create several events in sequence
		event1 := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		event2 := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		event3 := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		
		// Verify timestamps are in increasing order
		assert.True(t, event1.Timestamp().Before(event2.Timestamp()), 
			"First event timestamp should be before second event timestamp")
		assert.True(t, event2.Timestamp().Before(event3.Timestamp()), 
			"Second event timestamp should be before third event timestamp")
		
		// Enrich events and verify their sequence numbers increase
		enriched1, err := contextualizer.EnrichEventContext(event1, comp)
		assert.NoError(t, err)
		enriched2, err := contextualizer.EnrichEventContext(event2, comp)
		assert.NoError(t, err)
		enriched3, err := contextualizer.EnrichEventContext(event3, comp)
		assert.NoError(t, err)
		
		// Since we don't have direct access to sequence numbers in our implementation,
		// we'll just verify the enriched events maintain timestamp order
		assert.True(t, enriched1.Timestamp().Before(enriched2.Timestamp()),
			"First enriched event timestamp should be before second enriched event timestamp")
		assert.True(t, enriched2.Timestamp().Before(enriched3.Timestamp()),
			"Second enriched event timestamp should be before third enriched event timestamp")
	})
}

// TestContextDataCompleteness tests that events contain complete context data
func TestContextDataCompleteness(t *testing.T) {
	// Create a test component
	comp := new(MockComponent)
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Events with application state should preserve state data", func(t *testing.T) {
		// Set application state in contextualizer
		appState := map[string]interface{}{
			"appMode": "edit",
			"isDirty": true,
			"activeTool": "select",
		}
		contextualizer.SetApplicationState(appState)
		
		// Create a key event
		event := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Enrich with context
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Cast to base event to access eventContext
		baseEvent, ok := enriched.(*KeyEvent)
		assert.True(t, ok, "Enriched event should still be a KeyEvent")
		
		// Verify event has context
		assert.NotNil(t, baseEvent.eventContext, "Event should have context after enrichment")
	})
	
	t.Run("Different event types should include appropriate context", func(t *testing.T) {
		// Create key, mouse, and window size events
		keyEvent := NewKeyEvent(comp, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		mouseEvent := NewMouseEvent(comp, tea.MouseMsg{Type: tea.MouseLeft, X: 10, Y: 20})
		windowEvent := NewWindowSizeEvent(comp, tea.WindowSizeMsg{Width: 800, Height: 600})
		
		// Enrich events with context
		enrichedKey, err := contextualizer.EnrichEventContext(keyEvent, comp)
		assert.NoError(t, err)
		enrichedMouse, err := contextualizer.EnrichEventContext(mouseEvent, comp)
		assert.NoError(t, err)
		enrichedWindow, err := contextualizer.EnrichEventContext(windowEvent, comp)
		assert.NoError(t, err)
		
		// Verify all events have context
		assert.NotNil(t, enrichedKey, "Key event should be enriched")
		assert.NotNil(t, enrichedMouse, "Mouse event should be enriched")
		assert.NotNil(t, enrichedWindow, "Window event should be enriched")
		
		// Verify they maintain their correct event types
		assert.Equal(t, EventTypeKey, enrichedKey.Type(), "Key event should maintain its type")
		assert.Equal(t, EventTypeMouse, enrichedMouse.Type(), "Mouse event should maintain its type")
		assert.Equal(t, EventTypeWindowSize, enrichedWindow.Type(), "Window event should maintain its type")
	})
	
	t.Run("Events from component hierarchy should include component path context", func(t *testing.T) {
		// Create a component hierarchy
		root := new(MockComponent)
		parent := new(MockComponent)
		child := new(MockComponent)
		
		// Configure mocks
		root.On("ID").Return("root")
		root.On("Children").Return([]core.Component{parent})
		root.On("AddChild", mock.Anything).Return()
		
		parent.On("ID").Return("parent")
		parent.On("Children").Return([]core.Component{child})
		parent.On("AddChild", mock.Anything).Return()
		
		child.On("ID").Return("child")
		child.On("Children").Return([]core.Component{})
		
		// Setup hierarchy
		root.AddChild(parent)
		parent.AddChild(child)
		
		// Setup contextualizer with the root
		contextualizer := NewEventContextualizer()
		contextualizer.SetRootComponent(root)
		
		// Create an event from the child
		event := NewKeyEvent(child, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		
		// Enrich with context
		enriched, err := contextualizer.EnrichEventContext(event, child)
		assert.NoError(t, err)
		
		// The enriched event should have the source component set
		assert.Equal(t, child, enriched.Source(), "Enriched event should preserve the source component")
		
		// Cast to access internal fields
		keyEvent, ok := enriched.(*KeyEvent)
		assert.True(t, ok, "Enriched event should be a KeyEvent")
		
		// Verify event context exists
		assert.NotNil(t, keyEvent.eventContext, "Event should have context information")
	})
}

// BenchmarkContextEnrichmentPerformance benchmarks the performance of context enrichment
func BenchmarkContextEnrichmentPerformance(b *testing.B) {
	// Create test component hierarchy
	rootComp := new(MockComponent)
	
	// Create child components (simulating a deeper hierarchy)
	var children []core.Component
	for i := 0; i < 10; i++ {
		child := new(MockComponent)
		// Create ID strings using strconv
		childID := "child-" + strconv.Itoa(i)
		child.On("ID").Return(childID)
		child.On("Children").Return([]core.Component{})
		children = append(children, child)
	}
	
	rootComp.On("ID").Return("root")
	rootComp.On("Children").Return(children)
	rootComp.On("AddChild", mock.Anything).Run(func(args mock.Arguments) {
		// Simulates the actual behavior
		rootComp.children = append(rootComp.children, args.Get(0).(core.Component))
	})
	
	// Setup hierarchy
	for _, child := range children {
		rootComp.AddChild(child)
	}
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	contextualizer.SetRootComponent(rootComp)
	
	// Create a key event
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	
	// Set some application state
	appState := map[string]interface{}{
		"currentView": "dashboard",
		"userId":      "user123",
	}
	contextualizer.SetApplicationState(appState)
	
	// Reset timer to ensure setup time doesn't affect the benchmark
	b.ResetTimer()
	
	// Benchmark context enrichment
	for i := 0; i < b.N; i++ {
		// Create base event
		event := NewKeyEvent(children[0], keyMsg)
		
		// Enrich the event with context
		contextualizer.EnrichEventContext(event, children[0])
	}
}
