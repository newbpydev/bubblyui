package core

import (
	"sync"
)

// SlotCondition is a function that determines if a slot should be rendered
type SlotCondition func() bool

// Slot represents a named location in a component where content can be injected
type Slot struct {
	name           string
	content        *ComponentManager
	defaultContent *ComponentManager
	condition      SlotCondition
}

// ComponentManager handles the lifecycle and relationships of a UI component
type ComponentManager struct {
	// Component identification
	name string

	// Component hierarchy
	parent   *ComponentManager
	children []*ComponentManager
	mutex    sync.RWMutex

	// Lifecycle management
	hookManager *HookManager
	mounted     bool

	// Props management
	props map[string]interface{}

	// Event handling
	eventHandlers map[string][]EventHandler

	// Slot management
	slots map[string]*Slot
}

// EventHandler is a function that handles events
// Returns true if the event propagation should stop
type EventHandler func(eventData interface{}) bool

// NewComponentManager creates a new component manager with the given name
func NewComponentManager(name string) *ComponentManager {
	cm := &ComponentManager{
		name:          name,
		children:      make([]*ComponentManager, 0),
		hookManager:   NewHookManager(name),
		mounted:       false,
		props:         make(map[string]interface{}),
		eventHandlers: make(map[string][]EventHandler),
		slots:         make(map[string]*Slot),
	}
	return cm
}

// GetName returns the component's name
func (cm *ComponentManager) GetName() string {
	return cm.name
}

// GetHookManager returns the component's hook manager
func (cm *ComponentManager) GetHookManager() *HookManager {
	return cm.hookManager
}

// AddChild adds a child component to this component
func (cm *ComponentManager) AddChild(child *ComponentManager) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// If child already has a parent, remove it first
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}

	// Add child to this component
	cm.children = append(cm.children, child)

	// Set this component as the child's parent
	child.setParent(cm)

	// If this component is already mounted, mount the child too
	if cm.mounted {
		// We don't propagate errors here, but in a real implementation
		// you might want to handle them
		_ = child.Mount()
	}
}

// RemoveChild removes a child component from this component
func (cm *ComponentManager) RemoveChild(child *ComponentManager) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// If the component is mounted, unmount it first
	if cm.mounted && child.mounted {
		// We don't propagate errors here, but in a real implementation
		// you might want to handle them
		_ = child.Unmount()
	}

	// Find and remove the child
	for i, c := range cm.children {
		if c == child {
			// Remove from children slice
			cm.children = append(cm.children[:i], cm.children[i+1:]...)

			// Clear parent reference
			child.clearParent()
			break
		}
	}
}

// setParent sets the parent of this component (internal use)
func (cm *ComponentManager) setParent(parent *ComponentManager) {
	cm.parent = parent

	// Set up hook manager parent relationship
	cm.hookManager.SetParentManager(parent.hookManager)
}

// clearParent clears the parent reference (internal use)
func (cm *ComponentManager) clearParent() {
	cm.parent = nil

	// Set up hook manager parent relationship
	cm.hookManager.SetParentManager(nil)
}

// GetParent returns the parent component
func (cm *ComponentManager) GetParent() *ComponentManager {
	return cm.parent
}

// GetChildren returns a slice of all child components
func (cm *ComponentManager) GetChildren() []*ComponentManager {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// Create a copy to avoid race conditions
	result := make([]*ComponentManager, len(cm.children))
	copy(result, cm.children)

	return result
}

// Mount mounts the component and all its children
func (cm *ComponentManager) Mount() error {
	// Execute mount hooks
	err := cm.hookManager.ExecuteMountHooksWithErrorHandling()
	if err != nil {
		return err
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Skip if already mounted
	if cm.mounted {
		return nil
	}

	// Mark as mounted
	cm.mounted = true

	// Mount all children
	for _, child := range cm.children {
		err := child.Mount()
		if err != nil {
			return err
		}
	}

	// Mount slot content if present
	for _, slot := range cm.slots {
		if slot.content != nil {
			err := slot.content.Mount()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Unmount unmounts the component and all its children
func (cm *ComponentManager) Unmount() error {
	cm.mutex.Lock()

	// Skip if not mounted
	if !cm.mounted {
		cm.mutex.Unlock()
		return nil
	}

	// Unmount all children first
	for _, child := range cm.children {
		err := child.Unmount()
		if err != nil {
			cm.mutex.Unlock()
			return err
		}
	}

	// Unmount slot content if present
	for _, slot := range cm.slots {
		if slot.content != nil {
			err := slot.content.Unmount()
			if err != nil {
				cm.mutex.Unlock()
				return err
			}
		}
	}

	cm.mounted = false
	cm.mutex.Unlock()

	// Execute unmount hooks after children are unmounted
	return cm.hookManager.ExecuteUnmountHooksWithErrorHandling()
}

// SetProp sets a prop on the component
func (cm *ComponentManager) SetProp(key string, value interface{}) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.props[key] = value
}

// GetProp gets a prop from the component
func (cm *ComponentManager) GetProp(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	value, exists := cm.props[key]
	return value, exists
}

// GetInheritedProp gets a prop, falling back to parent props if not defined locally
func (cm *ComponentManager) GetInheritedProp(key string) (interface{}, bool) {
	// Check local props first
	value, exists := cm.GetProp(key)
	if exists {
		return value, true
	}

	// If not found locally and has parent, check parent
	if cm.parent != nil {
		return cm.parent.GetInheritedProp(key)
	}

	// Not found anywhere
	return nil, false
}

// HandleEvent registers an event handler for the given event type
func (cm *ComponentManager) HandleEvent(eventType string, handler func(eventData interface{}) bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Create handler slice if it doesn't exist
	if _, exists := cm.eventHandlers[eventType]; !exists {
		cm.eventHandlers[eventType] = make([]EventHandler, 0)
	}

	// Add the handler
	cm.eventHandlers[eventType] = append(cm.eventHandlers[eventType], handler)
}

// EmitEvent emits an event that bubbles up the component tree
func (cm *ComponentManager) EmitEvent(eventType string, eventData interface{}) bool {
	// Check if current component has handlers for this event
	cm.mutex.RLock()
	handlers, exists := cm.eventHandlers[eventType]
	cm.mutex.RUnlock()

	if exists {
		// Execute handlers
		for _, handler := range handlers {
			// If handler returns true, stop propagation
			if handler(eventData) {
				return true
			}
		}
	}

	// If not stopped and has parent, bubble up
	if cm.parent != nil {
		return cm.parent.EmitEvent(eventType, eventData)
	}

	// Event wasn't handled or stopped
	return false
}
