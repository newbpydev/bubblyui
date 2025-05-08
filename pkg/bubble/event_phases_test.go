package bubble_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubble"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockComponent simulates a component with parent-child relationship
type MockComponent struct {
	mock.Mock
	id       string
	parentID string
	parent   *MockComponent
	children []*MockComponent
}

func (m *MockComponent) ID() string {
	return m.id
}

func (m *MockComponent) Parent() core.Component {
	return m.parent
}

func (m *MockComponent) Children() []core.Component {
	result := make([]core.Component, len(m.children))
	for i, child := range m.children {
		result[i] = child
	}
	return result
}

// AddChild implements the core.Component interface
func (m *MockComponent) AddChild(child core.Component) {
	if mockChild, ok := child.(*MockComponent); ok {
		m.children = append(m.children, mockChild)
		mockChild.parent = m
	}
}

// Initialize implements the core.Component interface
func (m *MockComponent) Initialize() error {
	return nil
}

// Update implements the core.Component interface
func (m *MockComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	return nil, nil
}

// Render implements the core.Component interface
func (m *MockComponent) Render() string { return "" }

// Subscribe implements the core.Component interface
func (m *MockComponent) Subscribe(eventType bubble.EventType) {}

// Unsubscribe implements the core.Component interface
func (m *MockComponent) Unsubscribe(eventType bubble.EventType) {}

// SetParent implements the core.Component interface
func (m *MockComponent) SetParent(parent core.Component) {
	if mockParent, ok := parent.(*MockComponent); ok {
		m.parent = mockParent
	}
}

// RemoveChild implements the core.Component interface
func (m *MockComponent) RemoveChild(id string) bool {
	for i, child := range m.children {
		if child.ID() == id {
			// Remove the child at index i
			m.children = append(m.children[:i], m.children[i+1:]...)
			return true
		}
	}
	return false
}

// Dispose implements the core.Component interface
func (m *MockComponent) Dispose() error {
	return nil
}

func (m *MockComponent) HandleEvent(event bubble.Event) bool {
	args := m.Called(event)
	return args.Bool(0)
}

// Helper function to create a simple component tree
func createComponentTree() (*MockComponent, *MockComponent, *MockComponent) {
	// Create a tree with root -> parent -> child structure
	root := &MockComponent{id: "root"}
	parent := &MockComponent{id: "parent", parent: root, parentID: "root"}
	child := &MockComponent{id: "child", parent: parent, parentID: "parent"}
	
	// Set up children
	root.children = []*MockComponent{parent}
	parent.children = []*MockComponent{child}
	
	return root, parent, child
}

func TestEventCapturingPhase(t *testing.T) {
	root, parent, child := createComponentTree()

	// Create an event dispatcher
	dispatcher := bubble.NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	
	// Create an array to track the order of component handling
	var handlingOrder []string
	var handlingPhases []bubble.EventPhase
	
	// Set up expectations on each component to record handling order
	root.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Run(func(args mock.Arguments) {
		event := args.Get(0).(bubble.Event)
		handlingOrder = append(handlingOrder, "root")
		handlingPhases = append(handlingPhases, event.Phase())
	}).Return(false)
	
	parent.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Run(func(args mock.Arguments) {
		event := args.Get(0).(bubble.Event)
		handlingOrder = append(handlingOrder, "parent")
		handlingPhases = append(handlingPhases, event.Phase())
	}).Return(false)
	
	child.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Run(func(args mock.Arguments) {
		event := args.Get(0).(bubble.Event)
		handlingOrder = append(handlingOrder, "child")
		handlingPhases = append(handlingPhases, event.Phase())
	}).Return(false)
	
	// Create an event from the child component
	event := bubble.NewBaseEvent(
		bubble.EventTypeCustom, 
		child,
		bubble.EventCategoryUI,
		nil,
	)
	
	// Dispatch the event
	dispatcher.DispatchEvent(event)
	
	// Should traverse root->parent->child (capture) then child->parent->root (bubble)
	expectedOrder := []string{"root", "parent", "child", "child", "parent", "root"}
	expectedPhases := []bubble.EventPhase{
		bubble.PhaseCapturePhase,  // root during capture
		bubble.PhaseCapturePhase,  // parent during capture
		bubble.PhaseAtTarget,      // child at target
		bubble.PhaseAtTarget,      // child at target (duplicate because we check HandleEvent and listeners)
		bubble.PhaseBubblingPhase, // parent during bubble
		bubble.PhaseBubblingPhase, // root during bubble
	}
	
	// Verify expectations
	assert.Equal(t, expectedOrder, handlingOrder)
	assert.Equal(t, expectedPhases, handlingPhases)
	mock.AssertExpectationsForObjects(t, root, parent, child)
}

func TestPhaseSpecificEventHandlers(t *testing.T) {
	root, parent, child := createComponentTree()
	
	// Set up expectations for HandleEvent calls
	root.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	parent.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	child.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	
	// Create an event dispatcher
	dispatcher := bubble.NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	
	// Create maps to track handler calls for each phase
	captureHandlerCalled := make(map[string]bool)
	bubbleHandlerCalled := make(map[string]bool)
	
	// Register phase-specific event handlers for each component
	dispatcher.AddEventListenerWithOptions(root, bubble.EventTypeCustom, func(event bubble.Event) bool {
		captureHandlerCalled["root"] = true
		assert.Equal(t, bubble.PhaseCapturePhase, event.Phase())
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseCapturePhase})
	
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		captureHandlerCalled["parent"] = true
		assert.Equal(t, bubble.PhaseCapturePhase, event.Phase())
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseCapturePhase})
	
	dispatcher.AddEventListenerWithOptions(root, bubble.EventTypeCustom, func(event bubble.Event) bool {
		bubbleHandlerCalled["root"] = true
		assert.Equal(t, bubble.PhaseBubblingPhase, event.Phase())
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseBubblingPhase})
	
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		bubbleHandlerCalled["parent"] = true
		assert.Equal(t, bubble.PhaseBubblingPhase, event.Phase())
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseBubblingPhase})
	
	// Create an event from the child component
	event := bubble.NewBaseEvent(
		bubble.EventTypeCustom, 
		child,
		bubble.EventCategoryUI,
		nil,
	)
	
	// Dispatch the event
	dispatcher.DispatchEvent(event)
	
	// Verify that handlers were called in the correct phases
	assert.True(t, captureHandlerCalled["root"])
	assert.True(t, captureHandlerCalled["parent"])
	assert.True(t, bubbleHandlerCalled["root"])
	assert.True(t, bubbleHandlerCalled["parent"])
}

func TestExecutionOrderControl(t *testing.T) {
	root, parent, child := createComponentTree()
	
	// Set up expectations for HandleEvent calls
	root.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	parent.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	child.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	
	// Create an event dispatcher
	dispatcher := bubble.NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	
	// Track execution order
	var executionOrder []string
	
	// Register handlers with different priorities
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		executionOrder = append(executionOrder, "parent-low")
		return false
	}, bubble.EventListenerOptions{
		Phase:    bubble.PhaseBubblingPhase,
		Priority: bubble.PriorityLow,
	})
	
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		executionOrder = append(executionOrder, "parent-high")
		return false
	}, bubble.EventListenerOptions{
		Phase:    bubble.PhaseBubblingPhase,
		Priority: bubble.PriorityHigh,
	})
	
	// Create an event from the child component
	event := bubble.NewBaseEvent(
		bubble.EventTypeCustom, 
		child,
		bubble.EventCategoryUI,
		nil,
	)
	
	// Dispatch the event
	dispatcher.DispatchEvent(event)
	
	// High priority should execute before low priority
	assert.Equal(t, []string{"parent-high", "parent-low"}, executionOrder)
}

func TestPhaseSwitchingMechanism(t *testing.T) {
	root, parent, child := createComponentTree()
	
	// Set up expectations for HandleEvent calls
	root.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	parent.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	child.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	
	// Create an event dispatcher
	dispatcher := bubble.NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	
	// Track the phases of the event as it propagates
	var phases []bubble.EventPhase
	var components []string
	
	// Add a capture phase listener that forces a phase switch to bubbling
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		phases = append(phases, event.Phase())
		components = append(components, "parent-capture")
		
		// Switch to bubbling phase - this should skip the child target phase
		event.SwitchToPhase(bubble.PhaseBubblingPhase)
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseCapturePhase})
	
	// Add listeners for other phases to track execution
	dispatcher.AddEventListenerWithOptions(child, bubble.EventTypeCustom, func(event bubble.Event) bool {
		phases = append(phases, event.Phase())
		components = append(components, "child-target")
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseAtTarget})
	
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		phases = append(phases, event.Phase())
		components = append(components, "parent-bubble")
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseBubblingPhase})
	
	// Create an event from the child component
	event := bubble.NewBaseEvent(
		bubble.EventTypeCustom, 
		child,
		bubble.EventCategoryUI,
		nil,
	)
	
	// Dispatch the event
	dispatcher.DispatchEvent(event)
	
	// Should have parent-capture followed by parent-bubble, skipping child-target
	expectedPhases := []bubble.EventPhase{bubble.PhaseCapturePhase, bubble.PhaseBubblingPhase}
	expectedComponents := []string{"parent-capture", "parent-bubble"}
	
	assert.Equal(t, expectedPhases, phases)
	assert.Equal(t, expectedComponents, components)
}

func TestPhaseAwareEventContext(t *testing.T) {
	root, parent, child := createComponentTree()
	
	// Set up expectations for HandleEvent calls
	root.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	parent.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	child.On("HandleEvent", mock.AnythingOfType("*bubble.BaseEvent")).Return(false)
	
	// Create an event dispatcher
	dispatcher := bubble.NewEventDispatcher()
	dispatcher.SetRootComponent(root)
	
	// Create an event from the child component
	event := bubble.NewBaseEvent(
		bubble.EventTypeCustom, 
		child,
		bubble.EventCategoryUI,
		nil,
	)
	
	// Set up event context tracking
	var contextInfo []map[string]interface{}
	
	// Register phase-aware listeners
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		// Get the event context
		context := event.Context()
		
		// Add phase information to the context map
		contextInfo = append(contextInfo, map[string]interface{}{
			"phase":     event.Phase(),
			"component": "parent",
			"eventPath": len(event.Path()),
			"currentTarget": context.CurrentTarget.ID(),
		})
		
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseCapturePhase})
	
	dispatcher.AddEventListenerWithOptions(parent, bubble.EventTypeCustom, func(event bubble.Event) bool {
		// Get the event context
		context := event.Context()
		
		// Add phase information to the context map
		contextInfo = append(contextInfo, map[string]interface{}{
			"phase":     event.Phase(),
			"component": "parent",
			"eventPath": len(event.Path()),
			"currentTarget": context.CurrentTarget.ID(),
		})
		
		return false
	}, bubble.EventListenerOptions{Phase: bubble.PhaseBubblingPhase})
	
	// Dispatch the event
	dispatcher.DispatchEvent(event)
	
	// Verify context information
	assert.Equal(t, 2, len(contextInfo))
	assert.Equal(t, bubble.PhaseCapturePhase, contextInfo[0]["phase"])
	assert.Equal(t, bubble.PhaseBubblingPhase, contextInfo[1]["phase"])
	assert.Equal(t, "parent", contextInfo[0]["currentTarget"])
	assert.Equal(t, "parent", contextInfo[1]["currentTarget"])
}
