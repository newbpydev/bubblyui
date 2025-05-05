package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlotManagement(t *testing.T) {
	t.Run("Create and Fill Named Slots", func(t *testing.T) {
		// Create a parent component
		parent := NewComponentManager("ParentComponent")

		// Create a child component that will have slots
		child := NewComponentManager("ChildWithSlots")

		// Define a named slot in the child component
		child.RegisterSlot("header")
		child.RegisterSlot("content")
		child.RegisterSlot("footer")

		// Validate slots are registered
		assert.True(t, child.HasSlot("header"))
		assert.True(t, child.HasSlot("content"))
		assert.True(t, child.HasSlot("footer"))
		assert.False(t, child.HasSlot("nonexistent"))

		// Create components to fill the slots
		headerContent := NewComponentManager("HeaderContent")
		mainContent := NewComponentManager("MainContent")
		footerContent := NewComponentManager("FooterContent")

		// Fill the slots
		parent.AddChild(child)
		parent.FillSlot(child, "header", headerContent)
		parent.FillSlot(child, "content", mainContent)
		parent.FillSlot(child, "footer", footerContent)

		// Verify slots are filled
		headerSlot, exists := child.GetSlotContent("header")
		assert.True(t, exists)
		assert.Equal(t, "HeaderContent", headerSlot.GetName())

		contentSlot, exists := child.GetSlotContent("content")
		assert.True(t, exists)
		assert.Equal(t, "MainContent", contentSlot.GetName())

		footerSlot, exists := child.GetSlotContent("footer")
		assert.True(t, exists)
		assert.Equal(t, "FooterContent", footerSlot.GetName())

		// Verify that a nonexistent slot returns false
		_, exists = child.GetSlotContent("nonexistent")
		assert.False(t, exists)
	})

	t.Run("Slot Default Content", func(t *testing.T) {
		// Create components
		parent := NewComponentManager("ParentComponent")
		child := NewComponentManager("ChildWithSlots")
		defaultContent := NewComponentManager("DefaultContent")

		// Register a slot with default content
		child.RegisterSlotWithDefault("content", defaultContent)

		// Validate the slot has default content
		content, exists := child.GetSlotContent("content")
		assert.True(t, exists)
		assert.Equal(t, "DefaultContent", content.GetName())

		// Override default content
		parent.AddChild(child)
		customContent := NewComponentManager("CustomContent")
		parent.FillSlot(child, "content", customContent)

		// Validate custom content overrides default
		content, exists = child.GetSlotContent("content")
		assert.True(t, exists)
		assert.Equal(t, "CustomContent", content.GetName())
	})

	t.Run("Slot Context and Props", func(t *testing.T) {
		// Create components
		parent := NewComponentManager("ParentComponent")
		child := NewComponentManager("ChildWithSlots")
		slotContent := NewComponentManager("SlotContent")

		// Register a slot
		child.RegisterSlot("content")

		// Add child to parent and fill slot with context
		parent.AddChild(child)
		slotProps := map[string]interface{}{
			"title":  "Hello World",
			"active": true,
			"count":  42,
		}
		parent.FillSlotWithProps(child, "content", slotContent, slotProps)

		// Verify slot content has props
		content, exists := child.GetSlotContent("content")
		assert.True(t, exists)

		// Check props were passed to slot content
		title, exists := content.GetProp("title")
		assert.True(t, exists)
		assert.Equal(t, "Hello World", title)

		active, exists := content.GetProp("active")
		assert.True(t, exists)
		assert.Equal(t, true, active)

		count, exists := content.GetProp("count")
		assert.True(t, exists)
		assert.Equal(t, 42, count)
	})

	t.Run("Conditional Slot Rendering", func(t *testing.T) {
		// Create components
		parent := NewComponentManager("ParentComponent")
		child := NewComponentManager("ChildWithSlots")
		slotContent := NewComponentManager("SlotContent")

		// Register a slot with a condition
		child.RegisterSlotWithCondition("conditional", func() bool {
			value, exists := child.GetProp("showConditional")
			return exists && value.(bool)
		})

		// Initially, the condition is false
		parent.AddChild(child)
		parent.FillSlot(child, "conditional", slotContent)

		// Verify slot exists but won't render because condition is false
		assert.True(t, child.HasSlot("conditional"))
		assert.False(t, child.ShouldRenderSlot("conditional"))

		// Set the condition to true
		child.SetProp("showConditional", true)

		// Verify slot will now render
		assert.True(t, child.ShouldRenderSlot("conditional"))
	})

	t.Run("Slot Lifecycle", func(t *testing.T) {
		// Create components
		parent := NewComponentManager("ParentComponent")
		child := NewComponentManager("ChildWithSlots")
		slotContent := NewComponentManager("SlotContent")

		// Setup tracking for mount/unmount
		mountCalled := false
		unmountCalled := false

		slotContent.GetHookManager().OnMount(func() error {
			mountCalled = true
			return nil
		})

		slotContent.GetHookManager().OnUnmount(func() error {
			unmountCalled = true
			return nil
		})

		// Register a slot and fill it
		child.RegisterSlot("content")
		parent.AddChild(child)
		parent.FillSlot(child, "content", slotContent)

		// Mount the components
		parent.Mount()

		// Verify slot content was mounted
		assert.True(t, mountCalled)
		assert.False(t, unmountCalled)

		// Unmount
		parent.Unmount()

		// Verify slot content was unmounted
		assert.True(t, unmountCalled)
	})

	t.Run("Replacing Slot Content", func(t *testing.T) {
		// Create components
		parent := NewComponentManager("ParentComponent")
		child := NewComponentManager("ChildWithSlots")
		originalContent := NewComponentManager("OriginalContent")
		replacementContent := NewComponentManager("ReplacementContent")

		// Setup tracking for mount/unmount
		originalMountCalled := false
		originalUnmountCalled := false

		originalContent.GetHookManager().OnMount(func() error {
			originalMountCalled = true
			return nil
		})

		originalContent.GetHookManager().OnUnmount(func() error {
			originalUnmountCalled = true
			return nil
		})

		// Register a slot and fill it
		child.RegisterSlot("content")
		parent.AddChild(child)
		parent.FillSlot(child, "content", originalContent)

		// Mount the components
		parent.Mount()

		// Verify original content was mounted
		assert.True(t, originalMountCalled)

		// Replace the slot content
		parent.FillSlot(child, "content", replacementContent)

		// Verify original content was unmounted
		assert.True(t, originalUnmountCalled)

		// Verify replacement is in the slot
		content, exists := child.GetSlotContent("content")
		assert.True(t, exists)
		assert.Equal(t, "ReplacementContent", content.GetName())
	})
}
