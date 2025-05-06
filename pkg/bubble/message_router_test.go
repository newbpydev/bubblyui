package bubble

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
)

// ButtonClickEvent represents a test event for component event bubbling tests
type ButtonClickEvent struct {
	ComponentName string
	Value         string
}

// SetComponentName implements the component name setter interface
func (e *ButtonClickEvent) SetComponentName(name string) {
	e.ComponentName = name
}

// TestMessageRouting validates that messages can be properly routed between
// Bubble Tea and BubblyUI components.
func TestMessageRouting(t *testing.T) {
	t.Run("KeyEvent Routing", func(t *testing.T) {
		// Create a component tree with a parent and child component
		parent := core.NewComponentManager("ParentComponent")
		child := core.NewComponentManager("ChildComponent")
		parent.AddChild(child)

		// Create router and register components
		router := NewMessageRouter()
		router.RegisterComponent(parent)
		router.RegisterComponent(child)

		// Initialize components
		parent.Mount()

		// Route a key event
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		router.RouteMessage(keyMsg)

		// Verify components received the event through props
		_, parentHasKeyEvent := parent.GetProp("keyEvent")
		_, childHasKeyEvent := child.GetProp("keyEvent")
		lastKeyEvent, _ := parent.GetProp("lastKeyEvent")

		assert.True(t, parentHasKeyEvent, "Parent component should receive key event")
		assert.True(t, childHasKeyEvent, "Child component should receive key event")
		assert.Equal(t, "enter", lastKeyEvent, "Last key event should be 'enter'")
	})

	t.Run("Component Event Bubbling", func(t *testing.T) {
		// Create a component tree
		parent := core.NewComponentManager("ParentComponent")
		child := core.NewComponentManager("ChildComponent")
		parent.AddChild(child)

		// Create a router
		router := NewMessageRouter()
		router.RegisterComponent(parent)

		// Setup event tracking
		var receivedMsg tea.Msg
		router.SetEventHandler(func(msg tea.Msg) tea.Cmd {
			receivedMsg = msg
			return nil
		})

		// Mount components
		parent.Mount()

		// Trigger an event from the child component
		clickEvent := &ButtonClickEvent{
			Value: "clicked",
		}

		// Send the event through the router
		child.SetProp("event", clickEvent)
		router.BubbleEvent(child, clickEvent)

		// Verify the event was bubbled up to the handler
		assert.NotNil(t, receivedMsg, "Event should be bubbled up to handler")
		if bubbledEvent, ok := receivedMsg.(*ButtonClickEvent); ok {
			assert.Equal(t, child.GetName(), bubbledEvent.ComponentName)
			assert.Equal(t, "clicked", bubbledEvent.Value)
		} else {
			t.Fatal("Received message is not the expected event type")
		}
	})

	t.Run("Message Type Registration", func(t *testing.T) {
		// Create a router
		router := NewMessageRouter()

		// Define a custom message type
		type CustomMsg struct {
			Content string
		}

		// Register the message type
		router.RegisterMessageType("custom", func(msg tea.Msg) bool {
			_, ok := msg.(CustomMsg)
			return ok
		})

		// Test message type detection
		customMsg := CustomMsg{Content: "test"}
		msgType := router.GetMessageType(customMsg)

		assert.Equal(t, "custom", msgType, "Should correctly identify custom message type")

		// Test with an unregistered message type
		type UnknownMsg struct{}
		unknownType := router.GetMessageType(UnknownMsg{})

		assert.Equal(t, "unknown", unknownType, "Should return 'unknown' for unregistered types")
	})

	t.Run("Integration with BubbleModel", func(t *testing.T) {
		// Create a component
		root := core.NewComponentManager("RootComponent")

		// Create a bubble model with test mode
		model := NewBubbleModel(root, WithTestMode())

		// Create a router and connect it to the model
		router := NewMessageRouter()
		router.ConnectModel(model)

		// Initialize the model and router
		model.Init()

		// Track messages received by the component
		receivedKeyType := ""
		root.GetHookManager().OnUpdate(func(prev []interface{}) error {
			if keyEvent, ok := root.GetProp("lastKeyEvent"); ok {
				receivedKeyType = keyEvent.(string)
			}
			return nil
		}, []interface{}{})

		// Send a key message through the model's Update
		keyMsg := tea.KeyMsg{Type: tea.KeySpace}
		newModel, _ := model.Update(keyMsg)

		// Make sure the message was routed to the component
		assert.Equal(t, "space", receivedKeyType, "Component should receive the key event")

		// Ensure the model was returned correctly
		assert.Equal(t, model, newModel)
	})
}

// TestCustomMessageHandling verifies that custom message types can be handled correctly
func TestCustomMessageHandling(t *testing.T) {
	t.Run("Register Custom Handler", func(t *testing.T) {
		// Create a router
		router := NewMessageRouter()

		// Define a custom message type
		type CustomMsg struct {
			Value int
		}

		// Track handler execution
		handlerCalled := false
		handlerValue := 0

		// Register message handlers
		router.RegisterMessageHandler(
			func(msg tea.Msg) bool {
				// Check for raw CustomMsg type
				if _, ok := msg.(CustomMsg); ok {
					return true
				}
				// Check for CustomMsg wrapped in MessageWithContext
				if withCtx, ok := msg.(*MessageWithContext); ok {
					if _, ok := withCtx.OriginalMsg.(CustomMsg); ok {
						return true
					}
				}
				return false
			},
			func(msg tea.Msg) tea.Cmd {
				// Extract the CustomMsg, handling both raw and wrapped types
				var customMsg CustomMsg
				var ok bool

				if customMsg, ok = msg.(CustomMsg); !ok {
					if withCtx, ok := msg.(*MessageWithContext); ok {
						customMsg, ok = withCtx.OriginalMsg.(CustomMsg)
						if !ok {
							return nil
						}
					} else {
						return nil
					}
				}

				handlerCalled = true
				handlerValue = customMsg.Value
				return nil
			},
			10, // High priority
		)

		// Send a custom message
		router.RouteMessage(CustomMsg{Value: 42})

		// Verify handler was called
		assert.True(t, handlerCalled, "Custom handler should be called")
		assert.Equal(t, 42, handlerValue, "Handler should process message value")
	})

	t.Run("Multiple Handler Priority", func(t *testing.T) {
		// Create a router
		router := NewMessageRouter()

		// Define handlers with different priorities
		var calledHandler string

		// Low priority handler (added first)
		router.RegisterHandler(func(msg tea.Msg) bool {
			return true // Matches any message
		}, func(msg tea.Msg) tea.Cmd {
			calledHandler = "low"
			return nil
		})

		// High priority handler (added second)
		router.RegisterHandlerWithPriority(func(msg tea.Msg) bool {
			return true // Matches any message
		}, func(msg tea.Msg) tea.Cmd {
			calledHandler = "high"
			return nil
		}, 10) // Higher priority

		// Route a message
		router.RouteMessage(tea.WindowSizeMsg{})

		// High priority handler should be called
		assert.Equal(t, "high", calledHandler, "Higher priority handler should be called first")
	})
}
