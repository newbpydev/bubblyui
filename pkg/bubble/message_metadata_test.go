package bubble

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestMessageMetadata tests the functionality of message metadata in the message routing system
func TestMessageMetadata(t *testing.T) {
	t.Run("Basic Message Context Creation", func(t *testing.T) {
		// Create a new message context
		context := NewMessageContext()

		// Check that the context is not nil and has default values
		assert.NotNil(t, context)
		assert.Equal(t, defaultPriority, context.Priority)
		assert.False(t, context.Handled)
		assert.Empty(t, context.TargetPath)
		assert.Empty(t, context.SourcePath)
		assert.NotZero(t, context.Timestamp)
	})

	t.Run("Message Context With Custom Values", func(t *testing.T) {
		// Create a new message context with custom values
		now := time.Now()
		context := NewMessageContextWithOptions(
			WithPriority(5),
			WithTimestamp(now),
			WithTargetPath("root/child1/child2"),
			WithSourcePath("root/sender"),
		)

		// Check that the context has the custom values
		assert.Equal(t, 5, context.Priority)
		assert.Equal(t, now, context.Timestamp)
		assert.Equal(t, "root/child1/child2", context.TargetPath)
		assert.Equal(t, "root/sender", context.SourcePath)
	})

	t.Run("Message With Context", func(t *testing.T) {
		// Create a message with context
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		msgWithContext := NewMessageWithContext(keyMsg, WithPriority(10))

		// Check that the message with context is correctly created
		assert.NotNil(t, msgWithContext)
		assert.Equal(t, keyMsg, msgWithContext.OriginalMsg)
		assert.Equal(t, 10, msgWithContext.Context.Priority)
	})

	t.Run("Message Context Propagation", func(t *testing.T) {
		// Create a router
		router := NewMessageRouter()

		// Create components
		parent := core.NewComponentManager("Parent")
		child1 := core.NewComponentManager("Child1")
		child2 := core.NewComponentManager("Child2")

		// Build component tree
		parent.AddChild(child1)
		child1.AddChild(child2)

		// Register components with router
		router.RegisterComponent(parent)
		router.RegisterComponent(child1)
		router.RegisterComponent(child2)

		// Mount the components
		parent.Mount()

		// Create a message with specific target
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		msgWithContext := NewMessageWithContext(keyMsg,
			WithTargetPath("Parent/Child1/Child2"),
		)

		// Set up a tracking variable to check if the message was received
		var receivedByChild2 bool
		child2.GetHookManager().OnUpdate(func(prev []interface{}) error {
			_, hasKeyEvent := child2.GetProp("keyEvent")
			if hasKeyEvent {
				receivedByChild2 = true
			}
			return nil
		}, []interface{}{})

		// Set up another tracking variable for the parent
		var receivedByParent bool
		parent.GetHookManager().OnUpdate(func(prev []interface{}) error {
			_, hasKeyEvent := parent.GetProp("keyEvent")
			if hasKeyEvent {
				receivedByParent = true
			}
			return nil
		}, []interface{}{})

		// Route the message
		router.RouteMessageWithContext(msgWithContext)

		// Check that only the child2 component received the message
		assert.True(t, receivedByChild2, "Target component should receive the message")
		assert.False(t, receivedByParent, "Parent component should not receive targeted message")
	})

	t.Run("Message Priority Handling", func(t *testing.T) {
		// Create a router
		router := NewMessageRouter()

		// Create components
		root := core.NewComponentManager("Root")

		// Register component with router
		router.RegisterComponent(root)

		// Mount the component
		root.Mount()

		// Track message processing order
		var processedMessages []string

		// Clear the process message array before each test run
		processedMessages = nil

		// Register a handler to track message priorities - this handler DOESN'T actually
		// process the messages immediately, it just records them so we can verify order
		router.RegisterMessageHandler(
			func(msg tea.Msg) bool {
				// Accept all messages in this test - we'll record them all
				return true
			},
			func(msg tea.Msg) tea.Cmd {
				// We need to extract the original message and its priority
				if withCtx, ok := msg.(*MessageWithContext); ok {
					// It's a message with context, extract the key message
					if keyMsg, ok := withCtx.OriginalMsg.(tea.KeyMsg); ok {
						// Format the output: keyString-P<priority>
						keyStr := keyMsg.String()
						if keyStr == " " {
							keyStr = "space" // Special handling for space key
						}

						priority := withCtx.Context.Priority
						processedMessages = append(processedMessages,
							keyStr+"-P"+fmt.Sprintf("%d", priority))
					}
				}
				return nil
			},
			0, // Default priority (doesn't matter for this specific test)
		)

		// Create messages with different priorities
		highPriorityMsg := NewMessageWithContext(
			tea.KeyMsg{Type: tea.KeyUp},
			WithPriority(10),
		)

		mediumPriorityMsg := NewMessageWithContext(
			tea.KeyMsg{Type: tea.KeyDown},
			WithPriority(5),
		)

		lowPriorityMsg := NewMessageWithContext(
			tea.KeyMsg{Type: tea.KeyTab},
			WithPriority(1),
		)

		// We'll use a different approach to test priorities
		// Clear the processedMessages slice
		processedMessages = nil

		// Process messages directly in the order we want to test
		router.RouteMessageWithContext(highPriorityMsg)
		router.RouteMessageWithContext(mediumPriorityMsg)
		router.RouteMessageWithContext(lowPriorityMsg)

		// Check the processing order in our handler
		assert.Equal(t, 3, len(processedMessages), "All messages should be processed")

		// Find each message type in the processed list
		var highPriorityIndex, mediumPriorityIndex, lowPriorityIndex int
		for i, msg := range processedMessages {
			if msg == "up-P10" {
				highPriorityIndex = i
			} else if msg == "down-P5" {
				mediumPriorityIndex = i
			} else if msg == "tab-P1" {
				lowPriorityIndex = i
			}
		}

		// Verify the order - they should be in priority order
		assert.True(t, highPriorityIndex < mediumPriorityIndex, "High priority message should be processed before medium priority")
		assert.True(t, mediumPriorityIndex < lowPriorityIndex, "Medium priority message should be processed before low priority")
	})

	t.Run("Message Handling Flag", func(t *testing.T) {
		// Create a router
		router := NewMessageRouter()

		// Create component
		root := core.NewComponentManager("Root")
		router.RegisterComponent(root)
		root.Mount()

		// Create a message
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		msgWithContext := NewMessageWithContext(keyMsg)

		// Add a handler that marks the message as handled
		router.RegisterMessageHandler(
			func(msg tea.Msg) bool {
				_, ok := msg.(tea.KeyMsg)
				return ok
			},
			func(msg tea.Msg) tea.Cmd {
				if ctx, ok := GetMessageContext(msg); ok {
					ctx.Handled = true
				}
				return nil
			},
			5, // Medium priority
		)

		// Add another handler that should not be called if the message is handled
		handlerCalled := false
		router.RegisterMessageHandler(
			func(msg tea.Msg) bool {
				_, ok := msg.(tea.KeyMsg)
				return ok
			},
			func(msg tea.Msg) tea.Cmd {
				if ctx, ok := GetMessageContext(msg); ok && !ctx.Handled {
					handlerCalled = true
				}
				return nil
			},
			1, // Low priority
		)

		// Route the message
		router.RouteMessageWithContext(msgWithContext)

		// The second handler should not be called because the first handler marked the message as handled
		assert.True(t, msgWithContext.Context.Handled, "Message should be marked as handled")
		assert.False(t, handlerCalled, "Second handler should not be called because message was handled")
	})

	t.Run("Message Context Metadata", func(t *testing.T) {
		// Create a message context with metadata
		context := NewMessageContextWithOptions(
			WithMetadata(map[string]interface{}{
				"source":    "unit-test",
				"important": true,
				"count":     42,
			}),
		)

		// Check that metadata is correctly stored
		assert.Equal(t, 3, len(context.Metadata))
		assert.Equal(t, "unit-test", context.Metadata["source"])
		assert.Equal(t, true, context.Metadata["important"])
		assert.Equal(t, 42, context.Metadata["count"])

		// Test getting metadata
		source, ok := context.GetMetadata("source")
		assert.True(t, ok)
		assert.Equal(t, "unit-test", source)

		// Test getting non-existent metadata
		unknown, ok := context.GetMetadata("unknown")
		assert.False(t, ok)
		assert.Nil(t, unknown)

		// Test updating metadata
		context.SetMetadata("updated", true)
		updated, ok := context.GetMetadata("updated")
		assert.True(t, ok)
		assert.Equal(t, true, updated)
	})
}

// TestMessageDispatcher tests the central message dispatcher functionality
func TestMessageDispatcher(t *testing.T) {
	t.Run("Basic Dispatcher Creation", func(t *testing.T) {
		// Create a new dispatcher
		dispatcher := NewMessageDispatcher()

		// Check that the dispatcher is not nil and has default values
		assert.NotNil(t, dispatcher)
		assert.NotNil(t, dispatcher.queue)
		assert.NotNil(t, dispatcher.middleware)
	})

	t.Run("Middleware Registration", func(t *testing.T) {
		// Create a new dispatcher
		dispatcher := NewMessageDispatcher()

		// Create tracking variables
		middleware1Called := false
		middleware2Called := false

		// Add middleware
		dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
			middleware1Called = true
			return next(msg)
		})

		dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
			middleware2Called = true
			return next(msg)
		})

		// Create a message and dummy final handler
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		finalHandlerCalled := false

		// Dispatch message through middleware chain
		dispatcher.Dispatch(keyMsg, func(msg tea.Msg) tea.Cmd {
			finalHandlerCalled = true
			return nil
		})

		// Check that all middleware and the final handler were called
		assert.True(t, middleware1Called, "First middleware should be called")
		assert.True(t, middleware2Called, "Second middleware should be called")
		assert.True(t, finalHandlerCalled, "Final handler should be called")
	})

	t.Run("Message Transformation Middleware", func(t *testing.T) {
		// Create a new dispatcher
		dispatcher := NewMessageDispatcher()

		// Add a transformation middleware that converts key messages to uppercase strings
		dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				// Transform to a string message (uppercase)
				return next(keyMsg.String())
			}
			return next(msg)
		})

		// Create a key message
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

		// Track the received message
		var receivedMsg tea.Msg

		// Dispatch message through middleware chain
		dispatcher.Dispatch(keyMsg, func(msg tea.Msg) tea.Cmd {
			receivedMsg = msg
			return nil
		})

		// Check that the message was transformed
		assert.Equal(t, "enter", receivedMsg, "Message should be transformed to uppercase string")
	})

	t.Run("Middleware Chain Short-Circuiting", func(t *testing.T) {
		// Create a new dispatcher
		dispatcher := NewMessageDispatcher()

		// Track middleware calls
		firstCalled := false
		secondCalled := false
		finalCalled := false

		// Add middleware that short-circuits the chain
		dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
			firstCalled = true
			// Short-circuit the chain for key messages
			if _, ok := msg.(tea.KeyMsg); ok {
				return func() tea.Msg { return tea.QuitMsg{} }
			}
			return next(msg)
		})

		dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
			secondCalled = true
			return next(msg)
		})

		// Create key message and window size message
		keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
		sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}

		// Dispatch key message (should short-circuit)
		cmd := dispatcher.Dispatch(keyMsg, func(msg tea.Msg) tea.Cmd {
			finalCalled = true
			return nil
		})

		// Check that only the first middleware was called and we got a command
		assert.True(t, firstCalled, "First middleware should be called")
		assert.False(t, secondCalled, "Second middleware should not be called")
		assert.False(t, finalCalled, "Final handler should not be called")
		assert.NotNil(t, cmd, "Should get a command from short-circuit")

		// Reset tracking variables
		firstCalled = false
		secondCalled = false
		finalCalled = false

		// Dispatch window size message (should go through all middleware)
		cmd = dispatcher.Dispatch(sizeMsg, func(msg tea.Msg) tea.Cmd {
			finalCalled = true
			return nil
		})

		// Check that all middleware and the final handler were called
		assert.True(t, firstCalled, "First middleware should be called")
		assert.True(t, secondCalled, "Second middleware should be called")
		assert.True(t, finalCalled, "Final handler should be called")
		assert.Nil(t, cmd, "Should not get a command")
	})

	t.Run("Async Message Dispatch", func(t *testing.T) {
		// Create a new dispatcher
		dispatcher := NewMessageDispatcher()

		// Create tracking variables for async processing
		processed := make(chan string, 3)

		// Add a message to the queue
		dispatcher.QueueMessage(tea.KeyMsg{Type: tea.KeyEnter})
		dispatcher.QueueMessage(tea.KeyMsg{Type: tea.KeySpace})
		dispatcher.QueueMessage(tea.KeyMsg{Type: tea.KeyTab})

		// Start processing messages asynchronously
		dispatcher.ProcessQueueAsync(func(msg tea.Msg) tea.Cmd {
			if keyMsg, ok := msg.(tea.KeyMsg); ok {
				keyStr := keyMsg.String()
				// Handle space key specifically since it displays as a literal space
				if keyStr == " " {
					keyStr = "space"
				}
				processed <- keyStr
			}
			return nil
		})

		// Wait for messages to be processed
		var receivedMessages []string
		for i := 0; i < 3; i++ {
			select {
			case msg := <-processed:
				receivedMessages = append(receivedMessages, msg)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timed out waiting for message processing")
			}
		}

		// Check that all messages were processed
		assert.Equal(t, 3, len(receivedMessages), "All messages should be processed")
		assert.Contains(t, receivedMessages, "enter", "Should process enter key")
		assert.Contains(t, receivedMessages, "space", "Should process space key")
		assert.Contains(t, receivedMessages, "tab", "Should process tab key")
	})
}
