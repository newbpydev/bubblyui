package core

// This file contains all slot-related functionality for the ComponentManager

// RegisterSlot registers a new slot with the given name
func (cm *ComponentManager) RegisterSlot(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.slots[name] = &Slot{
		name:      name,
		condition: func() bool { return true }, // By default, slots are always rendered
	}
}

// RegisterSlotWithDefault registers a new slot with default content
func (cm *ComponentManager) RegisterSlotWithDefault(name string, defaultContent *ComponentManager) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.slots[name] = &Slot{
		name:           name,
		defaultContent: defaultContent,
		condition:      func() bool { return true }, // By default, slots are always rendered
	}
}

// RegisterSlotWithCondition registers a new slot with a rendering condition
func (cm *ComponentManager) RegisterSlotWithCondition(name string, condition SlotCondition) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.slots[name] = &Slot{
		name:      name,
		condition: condition,
	}
}

// HasSlot checks if a slot with the given name exists
func (cm *ComponentManager) HasSlot(name string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	_, exists := cm.slots[name]
	return exists
}

// GetSlotContent returns the content of a slot (either explicitly set content or default content)
func (cm *ComponentManager) GetSlotContent(name string) (*ComponentManager, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	slot, exists := cm.slots[name]
	if !exists {
		return nil, false
	}

	// If the slot has explicit content, return it
	if slot.content != nil {
		return slot.content, true
	}

	// Otherwise, return default content if present
	if slot.defaultContent != nil {
		return slot.defaultContent, true
	}

	// No content available
	return nil, false
}

// ShouldRenderSlot checks if a slot should be rendered based on its condition
func (cm *ComponentManager) ShouldRenderSlot(name string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	slot, exists := cm.slots[name]
	if !exists {
		return false
	}

	// Check if the slot has content to render
	hasContent := slot.content != nil || slot.defaultContent != nil

	// Return whether the slot has content and its condition is satisfied
	return hasContent && slot.condition()
}

// FillSlot sets the content of a named slot in a child component
func (cm *ComponentManager) FillSlot(child *ComponentManager, slotName string, content *ComponentManager) {
	child.mutex.Lock()
	defer child.mutex.Unlock()

	slot, exists := child.slots[slotName]
	if !exists {
		return // Ignore if the slot doesn't exist
	}

	// If there was previous content, unmount and remove it first
	if slot.content != nil && child.mounted {
		slot.content.Unmount()
	}

	// Set the new content
	slot.content = content

	// If the child is already mounted, mount the new content too
	if child.mounted && content != nil {
		content.Mount()
	}
}

// FillSlotWithProps sets the content of a named slot with additional props
func (cm *ComponentManager) FillSlotWithProps(child *ComponentManager, slotName string, content *ComponentManager, props map[string]interface{}) {
	// First, apply all props to the content
	for key, value := range props {
		content.SetProp(key, value)
	}

	// Then fill the slot with the content
	cm.FillSlot(child, slotName, content)
}
