package bubble

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestEventSourceIdentification tests that events properly identify their source component.
func TestEventSourceIdentification(t *testing.T) {
	// Create a component hierarchy
	root := NewMockComponent("root")
	parent := NewMockComponent("parent")
	child := NewMockComponent("child")
	
	// Configure mocks to handle method calls
	root.On("ID").Return("root")
	root.On("Children").Return([]core.Component{parent})
	root.On("Called").Return()
	
	parent.On("ID").Return("parent")
	parent.On("Children").Return([]core.Component{child})
	parent.On("Called").Return()
	
	child.On("ID").Return("child")
	child.On("Children").Return([]core.Component{})
	child.On("Called").Return()
	
	// Add components to hierarchy
	root.On("AddChild", mock.Anything).Return()
	parent.On("AddChild", mock.Anything).Return()
	root.AddChild(parent)
	parent.AddChild(child)
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Event maintains original source", func(t *testing.T) {
		// Create a basic event
		event := NewKeyEvent(child, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("a"),
		})
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, child)
		assert.NoError(t, err)
		
		// Verify the source is maintained
		assert.Equal(t, child, enriched.Source())
		assert.Equal(t, child.ID(), enriched.Source().ID())
	})
	
	t.Run("Event source is verifiable against component tree", func(t *testing.T) {
		// Create a basic event
		event := NewKeyEvent(child, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("a"),
		})
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, child)
		assert.NoError(t, err)
		
		// Verify the source exists in the component tree
		found := false
		for _, c := range parent.Children() {
			if c.ID() == enriched.Source().ID() {
				found = true
				break
			}
		}
		assert.True(t, found, "Event source should be findable in component tree")
	})
}

// TestEventTimestampAndSequence tests timestamp and sequence tracking in events.
func TestEventTimestampAndSequence(t *testing.T) {
	// Create a component
	comp := NewMockComponent("test-component")
	
	// Configure mock component
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	comp.On("Called").Return()
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Events have accurate timestamps", func(t *testing.T) {
		// Record time before creating event
		beforeTime := time.Now()
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		
		// Create a basic event
		event := NewKeyEvent(comp, tea.KeyMsg{})
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Record time after creating event
		time.Sleep(1 * time.Millisecond) // Ensure time difference
		afterTime := time.Now()
		
		// Verify the timestamp is between before and after times
		assert.True(t, enriched.Timestamp().After(beforeTime))
		assert.True(t, enriched.Timestamp().Before(afterTime))
	})
	
	t.Run("Events are tracked in sequence", func(t *testing.T) {
		// Create multiple events and get their sequence numbers
		event1 := NewKeyEvent(comp, tea.KeyMsg{})
		enriched1, _ := contextualizer.EnrichEventContext(event1, comp)
		
		event2 := NewKeyEvent(comp, tea.KeyMsg{})
		enriched2, _ := contextualizer.EnrichEventContext(event2, comp)
		
		event3 := NewKeyEvent(comp, tea.KeyMsg{})
		enriched3, _ := contextualizer.EnrichEventContext(event3, comp)
		
		// Get sequence numbers through the context
		seq1 := enriched1.(*KeyEvent).BaseEvent.eventContext.SequenceNumber
		seq2 := enriched2.(*KeyEvent).BaseEvent.eventContext.SequenceNumber
		seq3 := enriched3.(*KeyEvent).BaseEvent.eventContext.SequenceNumber
		
		// Verify sequence increases
		assert.Less(t, seq1, seq2)
		assert.Less(t, seq2, seq3)
	})
}

// TestComponentPathInEventData tests that events include the component path.
func TestComponentPathInEventData(t *testing.T) {
	// Create a component hierarchy
	root := NewMockComponent("root")
	parent := NewMockComponent("parent")
	child := NewMockComponent("child")
	
	// Configure mock components
	root.On("ID").Return("root")
	root.On("Children").Return([]core.Component{parent})
	root.On("Called").Return()
	
	parent.On("ID").Return("parent")
	parent.On("Children").Return([]core.Component{child})
	parent.On("Called").Return()
	
	child.On("ID").Return("child")
	child.On("Children").Return([]core.Component{})
	child.On("Called").Return()
	
	// Setup add child expectations
	root.On("AddChild", mock.Anything).Return()
	parent.On("AddChild", mock.Anything).Return()
	child.On("AddChild", mock.Anything).Return()
	
	// Add components to hierarchy
	root.AddChild(parent)
	parent.AddChild(child)
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Event includes complete component path", func(t *testing.T) {
		// Create a basic event
		event := NewKeyEvent(child, tea.KeyMsg{})
		
		// Set up context with hierarchy
		contextualizer.SetRootComponent(root)
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, child)
		assert.NoError(t, err)
		
		// Verify the path is complete
		path := enriched.Path()
		assert.Len(t, path, 3)
		assert.Equal(t, child.ID(), path[0].ID())
		assert.Equal(t, parent.ID(), path[1].ID())
		assert.Equal(t, root.ID(), path[2].ID())
	})
	
	t.Run("Path is built even for deeply nested components", func(t *testing.T) {
		// Create a deeper hierarchy
		deep1 := NewMockComponent("deep1")
		deep2 := NewMockComponent("deep2")
		deep3 := NewMockComponent("deep3")
		
		// Configure deeper mock components
		deep1.On("ID").Return("deep1")
		deep1.On("Children").Return([]core.Component{deep2})
		deep1.On("Called").Return()
		deep1.On("AddChild", mock.Anything).Return()
		
		deep2.On("ID").Return("deep2")
		deep2.On("Children").Return([]core.Component{deep3})
		deep2.On("Called").Return()
		deep2.On("AddChild", mock.Anything).Return()
		
		deep3.On("ID").Return("deep3")
		deep3.On("Children").Return([]core.Component{})
		deep3.On("Called").Return()
		
		// Update child's children list
		child.On("Children").Return([]core.Component{deep1}).Once()
		
		child.AddChild(deep1)
		deep1.AddChild(deep2)
		deep2.AddChild(deep3)
		
		// Create a basic event
		event := NewKeyEvent(deep3, tea.KeyMsg{})
		
		// Set up context with hierarchy
		contextualizer.SetRootComponent(root)
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, deep3)
		assert.NoError(t, err)
		
		// Verify the path is complete
		path := enriched.Path()
		assert.Len(t, path, 6)
		assert.Equal(t, deep3.ID(), path[0].ID())
		assert.Equal(t, deep2.ID(), path[1].ID())
		assert.Equal(t, deep1.ID(), path[2].ID())
		assert.Equal(t, child.ID(), path[3].ID())
		assert.Equal(t, parent.ID(), path[4].ID())
		assert.Equal(t, root.ID(), path[5].ID())
	})
}

// TestUserInteractionContext tests user interaction context in events.
func TestUserInteractionContext(t *testing.T) {
	// Create a component
	comp := NewMockComponent("test-component")
	
	// Configure mock component
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	comp.On("Called").Return()
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Mouse events capture user interaction position", func(t *testing.T) {
		// Create a mouse event
		mouseMsg := tea.MouseMsg{
			X: 10,
			Y: 20,
			Type: tea.MouseLeft,
		}
		event := NewMouseEvent(comp, mouseMsg)
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Verify interaction position
		mouseEvent := enriched.(*MouseEvent)
		assert.Equal(t, 10, mouseEvent.X())
		assert.Equal(t, 20, mouseEvent.Y())
		
		// Verify interaction context
		ctx := mouseEvent.BaseEvent.eventContext
		assert.Equal(t, 10, ctx.UserInteraction.CursorX)
		assert.Equal(t, 20, ctx.UserInteraction.CursorY)
		assert.Equal(t, "mouse", ctx.UserInteraction.InputType)
	})
	
	t.Run("Keyboard events capture user input type", func(t *testing.T) {
		// Create a key event
		keyMsg := tea.KeyMsg{
			Type: tea.KeyRunes,
			Runes: []rune("a"),
		}
		event := NewKeyEvent(comp, keyMsg)
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Verify interaction context
		keyEvent := enriched.(*KeyEvent)
		ctx := keyEvent.BaseEvent.eventContext
		assert.Equal(t, "keyboard", ctx.UserInteraction.InputType)
		assert.Equal(t, "a", ctx.UserInteraction.KeyPressed)
	})
	
	t.Run("Window events capture screen dimensions", func(t *testing.T) {
		// Create a window size event
		windowMsg := tea.WindowSizeMsg{
			Width: 80,
			Height: 24,
		}
		event := NewWindowSizeEvent(comp, windowMsg)
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Verify screen dimensions in context
		windowEvent := enriched.(*WindowSizeEvent)
		ctx := windowEvent.BaseEvent.eventContext
		assert.Equal(t, 80, ctx.UserInteraction.ScreenWidth)
		assert.Equal(t, 24, ctx.UserInteraction.ScreenHeight)
	})
}

// TestApplicationStateContext tests application state context in events.
func TestApplicationStateContext(t *testing.T) {
	// Create a component
	comp := NewMockComponent("test-component")
	
	// Configure mock component
	comp.On("ID").Return("test-component")
	comp.On("Children").Return([]core.Component{})
	comp.On("Called").Return()
	
	// Create an event contextualizer
	contextualizer := NewEventContextualizer()
	
	t.Run("Events include application state", func(t *testing.T) {
		// Set application state
		appState := map[string]interface{}{
			"currentView": "dashboard",
			"isLoggedIn": true,
			"userId": 12345,
		}
		contextualizer.SetApplicationState(appState)
		
		// Create a basic event
		event := NewKeyEvent(comp, tea.KeyMsg{})
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Verify application state in context
		keyEvent := enriched.(*KeyEvent)
		ctx := keyEvent.BaseEvent.eventContext
		
		assert.Equal(t, "dashboard", ctx.ApplicationState["currentView"])
		assert.Equal(t, true, ctx.ApplicationState["isLoggedIn"])
		assert.Equal(t, 12345, ctx.ApplicationState["userId"])
	})
	
	t.Run("Application state is deep copied", func(t *testing.T) {
		// Set application state
		nested := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		appState := map[string]interface{}{
			"nested": nested,
		}
		contextualizer.SetApplicationState(appState)
		
		// Create a basic event
		event := NewKeyEvent(comp, tea.KeyMsg{})
		
		// Contextualize event
		enriched, err := contextualizer.EnrichEventContext(event, comp)
		assert.NoError(t, err)
		
		// Modify original state
		nested["key1"] = "changed"
		nested["key2"] = 100
		
		// Verify event context has the original values
		keyEvent := enriched.(*KeyEvent)
		ctx := keyEvent.BaseEvent.eventContext
		nestedInCtx := ctx.ApplicationState["nested"].(map[string]interface{})
		
		assert.Equal(t, "value1", nestedInCtx["key1"])
		assert.Equal(t, 42, nestedInCtx["key2"])
	})
}
