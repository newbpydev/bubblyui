package bubble

import (
	"sort"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
)

// MessageRouter handles routing messages between Bubble Tea and BubblyUI components.
// It's responsible for:
// 1. Routing Bubble Tea messages to appropriate components
// 2. Bubbling component events up to the Bubble Tea model
// 3. Registering custom message types and handlers
type MessageRouter struct {
	// Components registered with this router
	components map[string]*core.ComponentManager

	// Message type detectors map message type names to detection functions
	messageTypeDetectors map[string]MessageTypeDetector

	// Message handlers are prioritized for processing messages
	messageHandlers []messageHandlerEntry

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

// EventHandler receives component events and returns a command
type EventHandler func(msg tea.Msg) tea.Cmd

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
		messageHandlers:      make([]messageHandlerEntry, 0),
		eventHandler:         nil,
		bubbleModel:          nil,
	}

	// Register built-in message types
	router.registerBuiltInMessageTypes()

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

	// Check all registered message types
	for typeName, detector := range mr.messageTypeDetectors {
		if detector(msg) {
			return typeName
		}
	}

	// Default if no match found
	return "unknown"
}

// RegisterHandler adds a message handler with default priority (0)
func (mr *MessageRouter) RegisterHandler(detector MessageTypeDetector, handler MessageHandler) {
	mr.RegisterHandlerWithPriority(detector, handler, 0)
}

// RegisterHandlerWithPriority adds a message handler with specified priority
// Higher priority handlers are executed first
func (mr *MessageRouter) RegisterHandlerWithPriority(
	detector MessageTypeDetector,
	handler MessageHandler,
	priority int,
) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	entry := messageHandlerEntry{
		detector: detector,
		handler:  handler,
		priority: priority,
	}

	mr.messageHandlers = append(mr.messageHandlers, entry)

	// Sort handlers by priority (higher priorities first)
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
	mr.mutex.RLock()
	handlers := make([]messageHandlerEntry, len(mr.messageHandlers))
	copy(handlers, mr.messageHandlers)
	mr.mutex.RUnlock()

	// Try all registered handlers in priority order
	// Since we sort by priority when handlers are registered (higher first),
	// we just need to iterate through the handlers
	for _, entry := range handlers {
		if entry.detector(msg) {
			// If a handler returns a command, we're done
			if cmd := entry.handler(msg); cmd != nil {
				return cmd
			}
			// Even if the command is nil, we still treated this message
			// Break here to ensure only the highest priority matching handler runs
			break
		}
	}

	// Process message based on type
	msgType := mr.GetMessageType(msg)

	// Route based on message type
	var cmd tea.Cmd
	switch msgType {
	case "key":
		cmd = mr.handleKeyMessage(msg.(tea.KeyMsg))
	case "windowSize":
		cmd = mr.handleWindowSizeMessage(msg.(tea.WindowSizeMsg))
	case "mouse":
		cmd = mr.handleMouseMessage(msg.(tea.MouseMsg))
	case "quit":
		// Handle quit message (nothing special to do, just return nil)
		return nil
	default:
		// For unknown types, just propagate to all components
		cmd = mr.propagateMessageToComponents(msg)
	}

	return cmd
}

// BubbleEvent sends an event from a component up to the event handler
func (mr *MessageRouter) BubbleEvent(component *core.ComponentManager, event interface{}) tea.Cmd {
	mr.mutex.RLock()
	handler := mr.eventHandler
	mr.mutex.RUnlock()

	// If no handler is registered, do nothing
	if handler == nil {
		return nil
	}

	// Convert event to a tea.Msg
	var msg tea.Msg = event

	// Pass to handler
	return handler(msg)
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
