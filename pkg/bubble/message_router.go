package bubble

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
)

// MessageRouter handles routing messages between Bubble Tea and BubblyUI components.
// It's responsible for:
// 1. Routing Bubble Tea messages to appropriate components
// 2. Bubbling component events up to the Bubble Tea model
// 3. Registering custom message types and handlers
// 4. Providing message metadata and context
type MessageRouter struct {
	// Components registered with this router
	components map[string]*core.ComponentManager

	// Message type detectors map message type names to detection functions
	messageTypeDetectors map[string]MessageTypeDetector

	// Cache for message type detection results
	messageTypeCache map[string]string

	// Message handlers are prioritized for processing messages
	messageHandlers []messageHandlerEntry

	// Custom handlers for specific message paths
	customHandlers map[string]HandleFunc

	// Component handlers map component paths to message handlers
	componentHandlers map[string]map[string]ComponentHandleFunc

	// Message dispatcher for middleware and async processing
	dispatcher *MessageDispatcher

	// Event handler for component events
	eventHandler EventHandler

	// Associated bubble model (if any)
	bubbleModel *BubbleModel

	// Thread safety
	mutex sync.RWMutex
}

// MessageTypeDetector is a function that determines if a message is of a specific type
type MessageTypeDetector func(msg tea.Msg) bool

// MessageHandler processes a message and returns a command
type MessageHandler func(msg tea.Msg) tea.Cmd

// HandleFunc processes a message with context and returns a command
type HandleFunc func(msg *MessageWithContext) tea.Cmd

// ComponentHandleFunc processes a message for a specific component and returns a command
type ComponentHandleFunc func(component *core.ComponentManager, msg tea.Msg) tea.Cmd

// BubbleEvent is a struct for event bubbling from a component to its parent
type BubbleEvent struct {
	ComponentName string
	Value         interface{}
}

// CustomEvent is an interface for events that need to track their source component
type CustomEvent interface {
	SetSourceComponent(name string)
	GetSourceComponent() string
}

// EventHandler receives component events and returns a command
type EventHandler func(msg tea.Msg) tea.Cmd

// ComponentEvent represents an event from a component with its path
type ComponentEvent struct {
	ComponentPath string
	Event         interface{}
}

// messageHandlerEntry combines a detector and handler with a priority
type messageHandlerEntry struct {
	detector MessageTypeDetector
	handler  MessageHandler
	priority int
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	router := &MessageRouter{
		components:           make(map[string]*core.ComponentManager),
		messageTypeDetectors: make(map[string]MessageTypeDetector),
		messageTypeCache:     make(map[string]string),
		messageHandlers:      make([]messageHandlerEntry, 0),
		customHandlers:       make(map[string]HandleFunc),
		componentHandlers:    make(map[string]map[string]ComponentHandleFunc),
		eventHandler:         nil,
		bubbleModel:          nil,
		dispatcher:           NewMessageDispatcher(),
	}

	// Register built-in message types
	router.registerBuiltInMessageTypes()

	// Set up message dispatcher middleware
	router.setupDispatcherMiddleware()

	return router
}

// registerBuiltInMessageTypes sets up handlers for standard Bubble Tea messages
func (mr *MessageRouter) registerBuiltInMessageTypes() {
	// KeyMsg detector
	mr.RegisterMessageType("key", func(msg tea.Msg) bool {
		_, ok := msg.(tea.KeyMsg)
		return ok
	})

	// WindowSizeMsg detector
	mr.RegisterMessageType("windowSize", func(msg tea.Msg) bool {
		_, ok := msg.(tea.WindowSizeMsg)
		return ok
	})

	// MouseMsg detector
	mr.RegisterMessageType("mouse", func(msg tea.Msg) bool {
		_, ok := msg.(tea.MouseMsg)
		return ok
	})

	// QuitMsg detector
	mr.RegisterMessageType("quit", func(msg tea.Msg) bool {
		_, ok := msg.(tea.QuitMsg)
		return ok
	})
}

// RegisterComponent adds a component to the router
func (mr *MessageRouter) RegisterComponent(component *core.ComponentManager) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	mr.components[component.GetName()] = component
}

// UnregisterComponent removes a component from the router
func (mr *MessageRouter) UnregisterComponent(componentID string) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	delete(mr.components, componentID)
}

// RegisterMessageType adds a new message type detector
func (mr *MessageRouter) RegisterMessageType(typeName string, detector MessageTypeDetector) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	mr.messageTypeDetectors[typeName] = detector
}

// GetMessageType determines the type of a message
func (mr *MessageRouter) GetMessageType(msg tea.Msg) string {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	// First unwrap the MessageWithContext if present
	var originalMsg tea.Msg = msg
	if withContext, ok := msg.(*MessageWithContext); ok {
		originalMsg = withContext.OriginalMsg
	}

	// Check the cache first for performance
	if cachedType, ok := mr.messageTypeCache[fmt.Sprintf("%T", originalMsg)]; ok {
		return cachedType
	}

	// Check all registered message types
	for typeName, detector := range mr.messageTypeDetectors {
		if detector(originalMsg) {
			// Cache the result for future lookups
			mr.messageTypeCache[fmt.Sprintf("%T", originalMsg)] = typeName
			return typeName
		}
	}

	// Default if no match found
	return "unknown"
}

// RegisterHandler adds a message handler with default priority (0)
func (mr *MessageRouter) RegisterHandler(detector MessageTypeDetector, handler MessageHandler) {
	mr.RegisterMessageHandler(detector, handler, 0)
}

// RegisterHandlerWithPriority adds a message handler with specified priority (alias for RegisterMessageHandler)
func (mr *MessageRouter) RegisterHandlerWithPriority(detector MessageTypeDetector, handler MessageHandler, priority int) {
	mr.RegisterMessageHandler(detector, handler, priority)
}

// RegisterMessageHandler registers a handler for a specific message type with priority
func (mr *MessageRouter) RegisterMessageHandler(detector MessageTypeDetector, handler MessageHandler, priority int) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	// Add to handlers list
	mr.messageHandlers = append(mr.messageHandlers, messageHandlerEntry{
		detector: detector,
		handler:  handler,
		priority: priority,
	})

	// Sort the handler list by priority (highest first)
	sort.Slice(mr.messageHandlers, func(i, j int) bool {
		return mr.messageHandlers[i].priority > mr.messageHandlers[j].priority
	})
}

// SetEventHandler sets the handler for component events
func (mr *MessageRouter) SetEventHandler(handler EventHandler) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	mr.eventHandler = handler
}

// ConnectModel associates a BubbleModel with this router
func (mr *MessageRouter) ConnectModel(model *BubbleModel) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	mr.bubbleModel = model

	// If the model has a root component, register it
	if root := model.GetRootComponent(); root != nil {
		mr.components[root.GetName()] = root
	}
}

// RouteMessage routes a Bubble Tea message to registered components and handlers
func (mr *MessageRouter) RouteMessage(msg tea.Msg) tea.Cmd {
	// Create a default message context if none exists
	var msgWithContext *MessageWithContext
	if withContext, ok := msg.(*MessageWithContext); ok {
		// Message already has context
		msgWithContext = withContext
	} else {
		// Wrap the message with a new context
		msgWithContext = NewMessageWithContext(msg)
	}

	// Route through the message dispatcher
	return mr.RouteMessageWithContext(msgWithContext)
}

// RouteMessageWithContext routes a message with context through the message dispatcher
func (mr *MessageRouter) RouteMessageWithContext(msgWithContext *MessageWithContext) tea.Cmd {
	// Use the dispatcher to process the message
	return mr.dispatcher.Dispatch(msgWithContext, func(msg tea.Msg) tea.Cmd {
		// Extract the original message and context
		withContext, ok := msg.(*MessageWithContext)
		if !ok {
			// If not a MessageWithContext, create one
			withContext = NewMessageWithContext(msg)
		}

		// If the message has already been handled, skip processing
		if withContext.Context.Handled {
			return nil
		}

		// Process based on target path
		if withContext.Context.TargetPath != "" {
			return mr.routeToTarget(withContext)
		}

		// Process based on priority and handlers
		return mr.processMessageHandlers(withContext)
	})
}

// routeToTarget routes the message to the specific target path
func (mr *MessageRouter) routeToTarget(msgWithContext *MessageWithContext) tea.Cmd {
	// First check if we have a custom handler for the target path
	if handler, ok := mr.customHandlers[msgWithContext.Context.TargetPath]; ok {
		// Mark as handled and use the custom handler
		msgWithContext.Context.Handled = true
		return handler(msgWithContext)
	}

	// Otherwise, propagate to the specific target component
	return mr.propagateTargetedMessage(msgWithContext)
}

// BubbleEvent handles event bubbling from a child component to its parents
func (mr *MessageRouter) BubbleEvent(component *core.ComponentManager, event interface{}) tea.Cmd {
	mr.mutex.RLock()
	eventHandler := mr.eventHandler
	mr.mutex.RUnlock()

	// Set source component name to track origin based on the event type
	// Check for event types with specific interfaces

	// First check if event is a ButtonClickEvent or implements SetComponentName
	if eventWithComponent, ok := event.(interface{ SetComponentName(string) }); ok {
		// Set the component name
		eventWithComponent.SetComponentName(component.GetName())

		// If an event handler is registered, call it with the original event
		// This is needed for backward compatibility with existing tests
		if eventHandler != nil {
			return eventHandler(event) // Pass original event, not wrapped
		}
	} else if bubbleEvent, ok := event.(*BubbleEvent); ok {
		// It's a BubbleEvent, update it
		bubbleEvent.ComponentName = component.GetName()

		// If an event handler is registered, call it with the original event
		if eventHandler != nil {
			return eventHandler(bubbleEvent) // Pass original event, not wrapped
		}
	} else if customEvent, ok := event.(CustomEvent); ok {
		// It supports the CustomEvent interface, set component name
		customEvent.SetSourceComponent(component.GetName())

		// If an event handler is registered, call it with the original event
		if eventHandler != nil {
			return eventHandler(event) // Pass original event, not wrapped
		}
	} else {
		// For unknown event types, we wrap them in a MessageWithContext
		msgWithContext := NewMessageWithContext(event)

		// If an event handler is registered, call it with the wrapped event
		if eventHandler != nil {
			return eventHandler(msgWithContext)
		}

		// Route through message dispatcher
		return mr.RouteMessageWithContext(msgWithContext)
	}

	// If we got here, we set a component name but had no handler
	// Create a message context and route it
	msgWithContext := NewMessageWithContext(event)
	return mr.RouteMessageWithContext(msgWithContext)
}

// propagateTargetedMessage sends a message to a specific component by path
func (mr *MessageRouter) propagateTargetedMessage(msgWithContext *MessageWithContext) tea.Cmd {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	targetPath := msgWithContext.Context.TargetPath
	originalMsg := msgWithContext.OriginalMsg

	// First check if we have a direct match for the target path
	if component, exists := mr.components[targetPath]; exists {
		// Direct match found
		return mr.deliverMessageToComponent(component, originalMsg, msgWithContext)
	}

	// If no direct match, try to find the child component by parsing the path
	// Format is expected to be Parent/Child1/Child2
	// First, find the root component in the path
	pathParts := strings.Split(targetPath, "/")
	if len(pathParts) > 0 {
		rootName := pathParts[0]
		if rootComponent, exists := mr.components[rootName]; exists {
			// Found the root component, now traverse to find the target
			currentComponent := rootComponent
			currentPath := rootName

			// Keep track if we've found the exact target
			foundTarget := false
			var targetComponent *core.ComponentManager

			// Navigate through child components
			for i := 1; i < len(pathParts); i++ {
				childFound := false

				// Look for a child with matching name
				for _, child := range currentComponent.GetChildren() {
					// Each child is already a *core.ComponentManager
					if child.GetName() == pathParts[i] {
						// Found a matching child
						currentComponent = child
						currentPath = currentPath + "/" + pathParts[i]

						// Register the component if not already registered
						if _, exists := mr.components[currentPath]; !exists {
							mr.components[currentPath] = child
						}

						// If this is the last part of the path, we've found the target
						if i == len(pathParts)-1 {
							foundTarget = true
							targetComponent = child
						}

						childFound = true
						break
					}
				}

				// If no matching child found, we can't proceed further
				if !childFound {
					break
				}
			}

			// If we found the target component, deliver the message
			if foundTarget && targetComponent != nil {
				return mr.deliverMessageToComponent(targetComponent, originalMsg, msgWithContext)
			}
		}
	}

	// Target component not found
	return nil
}

// deliverMessageToComponent handles delivering a message to a specific component
func (mr *MessageRouter) deliverMessageToComponent(component *core.ComponentManager, originalMsg tea.Msg, msgWithContext *MessageWithContext) tea.Cmd {
	// Process different message types without direct type assertions
	msgType := mr.GetMessageType(originalMsg)
	switch msgType {
	case "key":
		// Safely extract key message
		if keyMsg, ok := originalMsg.(tea.KeyMsg); ok {
			component.SetProp("keyEvent", keyMsg)
			keyName := keyMsg.String()
			if keyName == " " {
				keyName = "space"
			}
			component.SetProp("lastKeyEvent", keyName)
		}
	case "windowSize":
		// Safely extract window size message
		if sizeMsg, ok := originalMsg.(tea.WindowSizeMsg); ok {
			component.SetProp("windowWidth", sizeMsg.Width)
			component.SetProp("windowHeight", sizeMsg.Height)
			component.SetProp("windowSizeEvent", sizeMsg)
		}
	case "mouse":
		// Safely extract mouse message
		if mouseMsg, ok := originalMsg.(tea.MouseMsg); ok {
			component.SetProp("mouseEvent", mouseMsg)
			component.SetProp("mouseX", mouseMsg.X)
			component.SetProp("mouseY", mouseMsg.Y)

			// Convert mouse action to string
			mouseActionStr := mr.getMouseActionString(mouseMsg.Type)
			component.SetProp("mouseAction", mouseActionStr)
		}
	default:
		// For other message types, use the generic prop
		component.SetProp("bubbleTeaMessage", originalMsg)
	}

	// Execute update hooks
	component.GetHookManager().ExecuteUpdateHooks()

	// Mark as handled
	msgWithContext.Context.Handled = true
	return nil
}

// extractOriginalMessage safely extracts the original message from potentially nested MessageWithContext objects
func (mr *MessageRouter) extractOriginalMessage(msg tea.Msg) tea.Msg {
	if withContext, ok := msg.(*MessageWithContext); ok {
		return withContext.OriginalMsg
	}
	return msg
}

// processMessageHandlers processes a message through the registered message handlers
func (mr *MessageRouter) processMessageHandlers(msgWithContext *MessageWithContext) tea.Cmd {
	// Extract the original message
	originalMsg := msgWithContext.OriginalMsg

	// If the message has already been handled, skip processing
	if msgWithContext.Context.Handled {
		return nil
	}

	// Process custom handlers by priority
	mr.mutex.RLock()
	handlers := make([]messageHandlerEntry, len(mr.messageHandlers))
	copy(handlers, mr.messageHandlers)
	mr.mutex.RUnlock()

	// Sort handlers by priority (highest first)
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].priority > handlers[j].priority
	})

	// Try each handler in priority order
	for _, entry := range handlers {
		// For detection, use the original message to determine handler applicability
		unwrappedMsg := mr.extractOriginalMessage(originalMsg)
		if entry.detector(unwrappedMsg) {
			// Mark as handled
			msgWithContext.Context.Handled = true

			// Execute the handler with the original message with context to preserve metadata
			return entry.handler(msgWithContext)
		}
	}

	// If the message hasn't been handled by a custom handler, use built-in handling
	if !msgWithContext.Context.Handled {
		// Process message based on type
		msgType := mr.GetMessageType(originalMsg)

		// Route based on message type
		var cmd tea.Cmd
		switch msgType {
		case "key":
			// Safely extract the KeyMsg
			if keyMsg, ok := originalMsg.(tea.KeyMsg); ok {
				cmd = mr.handleKeyMessage(keyMsg)
			} else if withCtx, ok := originalMsg.(*MessageWithContext); ok {
				// Try to extract from nested context
				if keyMsg, ok := withCtx.OriginalMsg.(tea.KeyMsg); ok {
					cmd = mr.handleKeyMessage(keyMsg)
				}
			}
		case "windowSize":
			// Safely extract the WindowSizeMsg
			if sizeMsg, ok := originalMsg.(tea.WindowSizeMsg); ok {
				cmd = mr.handleWindowSizeMessage(sizeMsg)
			} else if withCtx, ok := originalMsg.(*MessageWithContext); ok {
				if sizeMsg, ok := withCtx.OriginalMsg.(tea.WindowSizeMsg); ok {
					cmd = mr.handleWindowSizeMessage(sizeMsg)
				}
			}
		case "mouse":
			// Safely extract the MouseMsg
			if mouseMsg, ok := originalMsg.(tea.MouseMsg); ok {
				cmd = mr.handleMouseMessage(mouseMsg)
			} else if withCtx, ok := originalMsg.(*MessageWithContext); ok {
				if mouseMsg, ok := withCtx.OriginalMsg.(tea.MouseMsg); ok {
					cmd = mr.handleMouseMessage(mouseMsg)
				}
			}
		case "quit":
			// Handle quit message (nothing special to do, just return nil)
			return nil
		default:
			// For unknown types, just propagate to all components
			cmd = mr.propagateMessageToComponents(originalMsg)
		}

		msgWithContext.Context.Handled = true
		return cmd
	}

	return nil
}

// setupDispatcherMiddleware configures the default middleware for the message dispatcher
func (mr *MessageRouter) setupDispatcherMiddleware() {
	// Add middleware for logging messages (if needed)
	mr.dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
		// Could add logging or telemetry here if needed
		return next(msg)
	})

	// Add middleware for transforming messages (if needed)
	mr.dispatcher.Use(func(msg tea.Msg, next DispatcherFunc) tea.Cmd {
		// No transformations needed by default, but could be added here
		return next(msg)
	})
}

// getMouseActionString converts a mouse event type to a string representation
func (mr *MessageRouter) getMouseActionString(eventType tea.MouseEventType) string {
	switch eventType {
	case tea.MouseLeft:
		return "left"
	case tea.MouseRight:
		return "right"
	case tea.MouseMiddle:
		return "middle"
	case tea.MouseRelease:
		return "release"
	case tea.MouseWheelUp:
		return "wheelup"
	case tea.MouseWheelDown:
		return "wheeldown"
	default:
		return "unknown"
	}
}

// handleKeyMessage routes key messages to components
func (mr *MessageRouter) handleKeyMessage(msg tea.KeyMsg) tea.Cmd {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	// Get a named representation of the key
	keyName := msg.String()

	// Handle special keys that might not have proper string representations
	if keyName == " " {
		keyName = "space"
	}

	// Process key events for all components
	for _, component := range mr.components {
		// Set the key event as a prop on the component
		component.SetProp("lastKeyEvent", keyName)
		component.SetProp("keyEvent", msg)

		// Execute update hooks to notify of the change
		component.GetHookManager().ExecuteUpdateHooks()
	}

	return nil
}

// handleWindowSizeMessage routes window size messages to components
func (mr *MessageRouter) handleWindowSizeMessage(msg tea.WindowSizeMsg) tea.Cmd {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	// Update all components with new window size
	for _, component := range mr.components {
		component.SetProp("windowWidth", msg.Width)
		component.SetProp("windowHeight", msg.Height)
		component.SetProp("windowSizeEvent", msg)

		// Execute update hooks to notify of the change
		component.GetHookManager().ExecuteUpdateHooks()
	}

	return nil
}

// handleMouseMessage routes mouse messages to components
func (mr *MessageRouter) handleMouseMessage(msg tea.MouseMsg) tea.Cmd {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	// Update all components with mouse event
	for _, component := range mr.components {
		component.SetProp("mouseEvent", msg)
		component.SetProp("mouseX", msg.X)
		component.SetProp("mouseY", msg.Y)
		// Convert mouse action to string representation manually
		mouseActionStr := "unknown"
		switch msg.Type {
		case tea.MouseLeft:
			mouseActionStr = "left"
		case tea.MouseRight:
			mouseActionStr = "right"
		case tea.MouseMiddle:
			mouseActionStr = "middle"
		case tea.MouseRelease:
			mouseActionStr = "release"
		case tea.MouseWheelUp:
			mouseActionStr = "wheelup"
		case tea.MouseWheelDown:
			mouseActionStr = "wheeldown"
		}
		component.SetProp("mouseAction", mouseActionStr)

		// Execute update hooks to notify of the change
		component.GetHookManager().ExecuteUpdateHooks()
	}

	return nil
}

// No longer need handleErrorMessage since we're using the quit message instead

// propagateMessageToComponents sends a generic message to all components
func (mr *MessageRouter) propagateMessageToComponents(msg tea.Msg) tea.Cmd {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	// Update all components with the generic message
	for _, component := range mr.components {
		component.SetProp("bubbleTeaMessage", msg)

		// Execute update hooks to notify of the change
		component.GetHookManager().ExecuteUpdateHooks()
	}

	return nil
}
