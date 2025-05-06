package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestComponentDiff tests the virtual DOM-like diffing mechanism
func TestComponentDiff(t *testing.T) {
	t.Run("Basic Component Diff", func(t *testing.T) {
		// Create original component tree
		originalRoot := NewComponentManager("Root")
		originalChild1 := NewComponentManager("Child1")
		originalChild2 := NewComponentManager("Child2")
		originalRoot.AddChild(originalChild1)
		originalRoot.AddChild(originalChild2)

		// Create new component tree (identical structure)
		newRoot := NewComponentManager("Root")
		newChild1 := NewComponentManager("Child1")
		newChild2 := NewComponentManager("Child2")
		newRoot.AddChild(newChild1)
		newRoot.AddChild(newChild2)

		// Take a snapshot of the original tree
		snapshot := NewComponentSnapshot(originalRoot)

		// Create a differ
		differ := NewComponentDiffer()

		// Compare with the new tree (should be no changes)
		diff := differ.Diff(snapshot, newRoot)

		// No changes should be detected in identical trees
		assert.Equal(t, 0, len(diff.Updates), "No components should need updating")
		assert.Equal(t, 0, len(diff.Additions), "No components should be added")
		assert.Equal(t, 0, len(diff.Removals), "No components should be removed")
	})

	t.Run("Component Property Changes", func(t *testing.T) {
		// Create original component tree
		originalRoot := NewComponentManager("Root")
		originalChild := NewComponentManager("Child")
		originalRoot.AddChild(originalChild)

		// Set some initial props
		originalRoot.SetProp("color", "blue")
		originalChild.SetProp("text", "Hello")

		// Take a snapshot of the original tree
		snapshot := NewComponentSnapshot(originalRoot)

		// Create a new tree with changed properties
		newRoot := NewComponentManager("Root")
		newChild := NewComponentManager("Child")
		newRoot.AddChild(newChild)

		// Change properties
		newRoot.SetProp("color", "red")   // Changed property
		newRoot.SetProp("size", 42)       // New property
		newChild.SetProp("text", "Hello") // Same property (no change)
		newChild.SetProp("visible", true) // New property

		// Create a differ
		differ := NewComponentDiffer()

		// Compare with the new tree
		diff := differ.Diff(snapshot, newRoot)

		// Only changed components should be in the diff
		assert.Equal(t, 2, len(diff.Updates), "Two components should need updating")
		assert.Equal(t, 0, len(diff.Additions), "No components should be added")
		assert.Equal(t, 0, len(diff.Removals), "No components should be removed")

		// Check that the updates have the expected components
		findRootUpdate := false
		findChildUpdate := false
		for _, update := range diff.Updates {
			if update.Component.GetName() == "Root" {
				findRootUpdate = true
				changedProps := update.ChangedProps
				assert.Contains(t, changedProps, "color", "Root should have changed color property")
				assert.Contains(t, changedProps, "size", "Root should have new size property")
				colorValue, exists := update.Component.GetProp("color")
				assert.True(t, exists, "Color property should exist")
				assert.Equal(t, "red", colorValue, "Root color should be updated to red")
			}
			if update.Component.GetName() == "Child" {
				findChildUpdate = true
				changedProps := update.ChangedProps
				assert.Contains(t, changedProps, "visible", "Child should have new visible property")
				assert.NotContains(t, changedProps, "text", "Child should not have changed text property")
			}
		}
		assert.True(t, findRootUpdate, "Root component should be in updates")
		assert.True(t, findChildUpdate, "Child component should be in updates")
	})

	t.Run("Component Addition", func(t *testing.T) {
		// Create original component tree
		originalRoot := NewComponentManager("Root")
		originalChild := NewComponentManager("Child1")
		originalRoot.AddChild(originalChild)

		// Take a snapshot of the original tree
		snapshot := NewComponentSnapshot(originalRoot)

		// Create a new tree with additional components
		newRoot := NewComponentManager("Root")
		newChild1 := NewComponentManager("Child1")
		newChild2 := NewComponentManager("Child2") // New child
		newRoot.AddChild(newChild1)
		newRoot.AddChild(newChild2)

		// Create a differ
		differ := NewComponentDiffer()

		// Compare with the new tree
		diff := differ.Diff(snapshot, newRoot)

		// Check that the new component is detected as an addition
		assert.Equal(t, 0, len(diff.Updates), "No components should need updating")
		assert.Equal(t, 1, len(diff.Additions), "One component should be added")
		assert.Equal(t, 0, len(diff.Removals), "No components should be removed")

		// Check that the added component is the expected one
		assert.Equal(t, "Child2", diff.Additions[0].Component.GetName(), "Child2 should be detected as added")
		assert.Equal(t, "Root", diff.Additions[0].Parent.GetName(), "Parent of added component should be Root")
	})

	t.Run("Component Removal", func(t *testing.T) {
		// Create original component tree
		originalRoot := NewComponentManager("Root")
		originalChild1 := NewComponentManager("Child1")
		originalChild2 := NewComponentManager("Child2")
		originalRoot.AddChild(originalChild1)
		originalRoot.AddChild(originalChild2)

		// Take a snapshot of the original tree
		snapshot := NewComponentSnapshot(originalRoot)

		// Create a new tree with a component removed
		newRoot := NewComponentManager("Root")
		newChild1 := NewComponentManager("Child1")
		// Child2 is missing in new tree
		newRoot.AddChild(newChild1)

		// Create a differ
		differ := NewComponentDiffer()

		// Compare with the new tree
		diff := differ.Diff(snapshot, newRoot)

		// Check that the removed component is detected
		assert.Equal(t, 0, len(diff.Updates), "No components should need updating")
		assert.Equal(t, 0, len(diff.Additions), "No components should be added")
		assert.Equal(t, 1, len(diff.Removals), "One component should be removed")

		// Check that the removed component is the expected one
		assert.Equal(t, "Child2", diff.Removals[0].Component.GetName(), "Child2 should be detected as removed")
		assert.Equal(t, "Root", diff.Removals[0].Parent.GetName(), "Parent of removed component should be Root")
	})

	t.Run("Deep Component Tree Changes", func(t *testing.T) {
		// Create original deep component tree
		originalRoot := NewComponentManager("Root")
		originalLevel1 := NewComponentManager("Level1")
		originalLevel2 := NewComponentManager("Level2")
		originalLevel3A := NewComponentManager("Level3A")
		originalLevel3B := NewComponentManager("Level3B")

		originalRoot.AddChild(originalLevel1)
		originalLevel1.AddChild(originalLevel2)
		originalLevel2.AddChild(originalLevel3A)
		originalLevel2.AddChild(originalLevel3B)

		// Take a snapshot of the original tree
		snapshot := NewComponentSnapshot(originalRoot)

		// Create a new tree with changes at different levels
		newRoot := NewComponentManager("Root")
		newLevel1 := NewComponentManager("Level1")
		newLevel2 := NewComponentManager("Level2")
		newLevel3A := NewComponentManager("Level3A")
		// Level3B is missing
		newLevel3C := NewComponentManager("Level3C") // New component

		newRoot.AddChild(newLevel1)
		newLevel1.AddChild(newLevel2)
		newLevel2.AddChild(newLevel3A)
		newLevel2.AddChild(newLevel3C)

		// Create a differ
		differ := NewComponentDiffer()

		// Compare with the new tree
		diff := differ.Diff(snapshot, newRoot)

		// Check that the changes are detected correctly
		assert.Equal(t, 0, len(diff.Updates), "No components should need updating")
		assert.Equal(t, 1, len(diff.Additions), "One component should be added")
		assert.Equal(t, 1, len(diff.Removals), "One component should be removed")

		// Check that the added component is the expected one
		assert.Equal(t, "Level3C", diff.Additions[0].Component.GetName(), "Level3C should be detected as added")
		assert.Equal(t, "Level2", diff.Additions[0].Parent.GetName(), "Parent of added component should be Level2")

		// Check that the removed component is the expected one
		assert.Equal(t, "Level3B", diff.Removals[0].Component.GetName(), "Level3B should be detected as removed")
		assert.Equal(t, "Level2", diff.Removals[0].Parent.GetName(), "Parent of removed component should be Level2")
	})

	t.Run("Key-Based Reconciliation", func(t *testing.T) {
		// Create original tree with keyed components
		originalRoot := NewComponentManager("Root")
		originalChild1 := NewComponentManager("Child1")
		originalChild2 := NewComponentManager("Child2")
		originalChild3 := NewComponentManager("Child3")

		// Set keys for reconciliation
		originalChild1.SetProp("key", "A")
		originalChild2.SetProp("key", "B")
		originalChild3.SetProp("key", "C")

		originalRoot.AddChild(originalChild1)
		originalRoot.AddChild(originalChild2)
		originalRoot.AddChild(originalChild3)

		// Take a snapshot
		snapshot := NewComponentSnapshot(originalRoot)

		// Create a new tree with reordered components (same keys)
		newRoot := NewComponentManager("Root")
		newChild1 := NewComponentManager("Child1-Renamed") // Name changed but key is the same
		newChild2 := NewComponentManager("Child2")
		newChild3 := NewComponentManager("Child3")

		// Set the same keys
		newChild1.SetProp("key", "A")
		newChild2.SetProp("key", "B")
		newChild3.SetProp("key", "C")

		// Different order
		newRoot.AddChild(newChild3) // C first now
		newRoot.AddChild(newChild1) // A second now
		newRoot.AddChild(newChild2) // B third now

		// Create a differ with key reconciliation
		differ := NewComponentDiffer()
		differ.EnableKeyReconciliation("key") // Enable key-based reconciliation using "key" prop

		// Compare with the new tree
		diff := differ.Diff(snapshot, newRoot)

		// Since keys match, it should detect reordering rather than additions/removals
		assert.Equal(t, 1, len(diff.Updates), "One component should need updating (name change)")
		assert.Equal(t, 0, len(diff.Additions), "No components should be added")
		assert.Equal(t, 0, len(diff.Removals), "No components should be removed")
		assert.Equal(t, 1, len(diff.Reorders), "One reorder operation should be detected")

		// Check the reorder operation
		assert.Equal(t, 3, len(diff.Reorders[0].NewOrder), "Reorder should include all 3 children")

		// Check that the updated component has the name change
		foundUpdate := false
		for _, update := range diff.Updates {
			if update.Component.GetName() == "Child1-Renamed" {
				foundUpdate = true
				break
			}
		}
		assert.True(t, foundUpdate, "Component with changed name should be in updates")
	})
}
