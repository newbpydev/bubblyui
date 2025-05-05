package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParentChildRelationship(t *testing.T) {
	t.Run("Add and Remove Child Components", func(t *testing.T) {
		// Create parent component
		parent := NewComponentManager("parent")

		// Create child components
		child1 := NewComponentManager("child1")
		child2 := NewComponentManager("child2")

		// Add children to parent
		parent.AddChild(child1)
		parent.AddChild(child2)

		// Verify children were added
		children := parent.GetChildren()
		assert.Equal(t, 2, len(children), "Parent should have 2 children")
		assert.Contains(t, children, child1, "child1 should be in children")
		assert.Contains(t, children, child2, "child2 should be in children")

		// Verify parent was set correctly
		assert.Equal(t, parent, child1.GetParent(), "child1's parent should be set")
		assert.Equal(t, parent, child2.GetParent(), "child2's parent should be set")

		// Remove a child
		parent.RemoveChild(child1)

		// Verify child was removed
		children = parent.GetChildren()
		assert.Equal(t, 1, len(children), "Parent should have 1 child after removal")
		assert.NotContains(t, children, child1, "child1 should no longer be in children")
		assert.Contains(t, children, child2, "child2 should still be in children")

		// Verify parent reference was cleared
		assert.Nil(t, child1.GetParent(), "child1's parent should be nil after removal")
	})

	t.Run("Parent-Child Lifecycle", func(t *testing.T) {
		// Create parent and child components
		parent := NewComponentManager("parent")
		child := NewComponentManager("child")

		// Track mount/unmount events
		parentMounted := false
		childMounted := false
		parentUnmounted := false
		childUnmounted := false

		// Set up lifecycle hooks
		parent.GetHookManager().OnMount(func() error {
			parentMounted = true
			return nil
		})

		parent.GetHookManager().OnUnmount(func() error {
			parentUnmounted = true
			return nil
		})

		child.GetHookManager().OnMount(func() error {
			childMounted = true
			return nil
		})

		child.GetHookManager().OnUnmount(func() error {
			childUnmounted = true
			return nil
		})

		// Add child to parent
		parent.AddChild(child)

		// Mount parent (should cascade to children)
		err := parent.Mount()
		assert.NoError(t, err, "Mount should succeed")

		// Verify both were mounted
		assert.True(t, parentMounted, "Parent should be mounted")
		assert.True(t, childMounted, "Child should be mounted")

		// Unmount parent (should cascade to children)
		err = parent.Unmount()
		assert.NoError(t, err, "Unmount should succeed")

		// Verify both were unmounted
		assert.True(t, parentUnmounted, "Parent should be unmounted")
		assert.True(t, childUnmounted, "Child should be unmounted")
	})
}

func TestEventPropagation(t *testing.T) {
	t.Run("Events Bubble Up", func(t *testing.T) {
		// Create component hierarchy
		root := NewComponentManager("root")
		middle := NewComponentManager("middle")
		leaf := NewComponentManager("leaf")

		// Connect the components
		root.AddChild(middle)
		middle.AddChild(leaf)

		// Track event propagation
		eventReceivedByRoot := false
		eventReceivedByMiddle := false

		// Set up event handlers
		root.HandleEvent("test-event", func(eventData interface{}) bool {
			eventReceivedByRoot = true
			return false // Don't stop propagation
		})

		middle.HandleEvent("test-event", func(eventData interface{}) bool {
			eventReceivedByMiddle = true
			return false // Don't stop propagation
		})

		// Trigger event from leaf
		leaf.EmitEvent("test-event", "event data")

		// Verify event propagated up
		assert.True(t, eventReceivedByMiddle, "Event should be received by middle component")
		assert.True(t, eventReceivedByRoot, "Event should be received by root component")
	})

	t.Run("Event Stopping", func(t *testing.T) {
		// Create component hierarchy
		root := NewComponentManager("root")
		middle := NewComponentManager("middle")
		leaf := NewComponentManager("leaf")

		// Connect the components
		root.AddChild(middle)
		middle.AddChild(leaf)

		// Track event propagation
		eventReceivedByRoot := false

		// Set up event handlers
		root.HandleEvent("test-event", func(eventData interface{}) bool {
			eventReceivedByRoot = true
			return false // Don't stop propagation
		})

		middle.HandleEvent("test-event", func(eventData interface{}) bool {
			// Stop propagation
			return true
		})

		// Trigger event from leaf
		leaf.EmitEvent("test-event", "event data")

		// Verify event did not reach root
		assert.False(t, eventReceivedByRoot, "Event should not reach root when stopped by middle")
	})
}

func TestPropInheritance(t *testing.T) {
	t.Run("Props Propagate Down", func(t *testing.T) {
		// Create component hierarchy
		parent := NewComponentManager("parent")
		child := NewComponentManager("child")

		// Connect the components
		parent.AddChild(child)

		// Set prop on parent
		parent.SetProp("theme", "dark")

		// Verify child inherits the prop
		value, exists := child.GetInheritedProp("theme")
		assert.True(t, exists, "Child should see inherited prop")
		assert.Equal(t, "dark", value, "Child should inherit parent's theme prop")

		// Change parent prop
		parent.SetProp("theme", "light")

		// Verify child sees the change
		value, exists = child.GetInheritedProp("theme")
		assert.True(t, exists, "Child should see inherited prop after change")
		assert.Equal(t, "light", value, "Child should inherit updated theme prop")
	})

	t.Run("Prop Overriding", func(t *testing.T) {
		// Create component hierarchy
		parent := NewComponentManager("parent")
		child := NewComponentManager("child")

		// Connect the components
		parent.AddChild(child)

		// Set prop on parent
		parent.SetProp("theme", "dark")

		// Override prop on child
		child.SetProp("theme", "light")

		// Get local prop (should be override)
		value, exists := child.GetProp("theme")
		assert.True(t, exists, "Child should have local prop")
		assert.Equal(t, "light", value, "Child should have its own theme value")

		// Get inherited prop (should include override)
		value, exists = child.GetInheritedProp("theme")
		assert.True(t, exists, "Child should have prop")
		assert.Equal(t, "light", value, "Child's own prop should override inherited prop")
	})
}
