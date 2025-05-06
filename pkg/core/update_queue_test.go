package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUpdateQueue tests the update queue functionality
func TestUpdateQueue(t *testing.T) {
	t.Run("Basic Queue Operations", func(t *testing.T) {
		// Create a new update queue
		queue := NewUpdateQueue()

		// Create test components with different priorities
		compHigh := NewComponentManager("HighPriority")
		compMedium := NewComponentManager("MediumPriority")
		compLow := NewComponentManager("LowPriority")

		// Enqueue components with different priorities
		queue.Enqueue(compHigh, UpdatePriorityHigh)
		queue.Enqueue(compMedium, UpdatePriorityMedium)
		queue.Enqueue(compLow, UpdatePriorityLow)

		// Check queue length
		assert.Equal(t, 3, queue.Length(), "Queue should have 3 components")

		// Dequeue should return components in priority order
		comp1 := queue.Dequeue()
		assert.Equal(t, "HighPriority", comp1.GetName(), "High priority component should be dequeued first")

		comp2 := queue.Dequeue()
		assert.Equal(t, "MediumPriority", comp2.GetName(), "Medium priority component should be dequeued second")

		comp3 := queue.Dequeue()
		assert.Equal(t, "LowPriority", comp3.GetName(), "Low priority component should be dequeued third")

		// Queue should be empty now
		assert.Equal(t, 0, queue.Length(), "Queue should be empty")
		assert.Nil(t, queue.Dequeue(), "Dequeue on empty queue should return nil")
	})

	t.Run("Component Uniqueness", func(t *testing.T) {
		// Create a new update queue
		queue := NewUpdateQueue()

		// Create a test component
		comp := NewComponentManager("TestComponent")

		// Enqueue the same component multiple times
		queue.Enqueue(comp, UpdatePriorityMedium)
		queue.Enqueue(comp, UpdatePriorityLow)  // Lower priority should be ignored
		queue.Enqueue(comp, UpdatePriorityHigh) // Higher priority should replace medium

		// Check queue length - should only have the component once
		assert.Equal(t, 1, queue.Length(), "Queue should have only 1 component despite multiple enqueues")

		// Component should be at the highest priority it was enqueued with
		dequeuedComp := queue.Dequeue()
		assert.Equal(t, comp, dequeuedComp, "Dequeued component should be the same as enqueued")
		assert.Equal(t, 0, queue.Length(), "Queue should be empty after dequeue")
	})

	t.Run("Component Priority Change", func(t *testing.T) {
		// Create a new update queue
		queue := NewUpdateQueue()

		// Create test components
		comp1 := NewComponentManager("Component1")
		comp2 := NewComponentManager("Component2")
		comp3 := NewComponentManager("Component3")

		// Enqueue with initial priorities
		queue.Enqueue(comp1, UpdatePriorityLow)
		queue.Enqueue(comp2, UpdatePriorityMedium)
		queue.Enqueue(comp3, UpdatePriorityHigh)

		// Change priority of comp1 to highest
		queue.UpdatePriority(comp1, UpdatePriorityUrgent)

		// Dequeue should now return comp1 first
		dequeuedComp := queue.Dequeue()
		assert.Equal(t, comp1, dequeuedComp, "Component with updated priority should be dequeued first")
	})

	t.Run("Batched Updates", func(t *testing.T) {
		// Create a new update queue
		queue := NewUpdateQueue()

		// Create test components for batched updates
		parent := NewComponentManager("ParentComponent")
		child1 := NewComponentManager("ChildComponent1")
		child2 := NewComponentManager("ChildComponent2")

		// Build component tree
		parent.AddChild(child1)
		parent.AddChild(child2)

		// Set a shared batch ID for related components
		batchID := "test-batch-1"

		// Enqueue components with the same batch ID
		queue.EnqueueBatched(parent, UpdatePriorityMedium, batchID)
		queue.EnqueueBatched(child1, UpdatePriorityLow, batchID)  // Despite lower priority, should be grouped
		queue.EnqueueBatched(child2, UpdatePriorityHigh, batchID) // Despite higher priority, should be grouped

		// Dequeue the batch
		batch := queue.DequeueBatch(batchID)

		// Should get all components in the batch
		assert.Equal(t, 3, len(batch), "Batch should contain all 3 components")

		// Queue should be empty
		assert.Equal(t, 0, queue.Length(), "Queue should be empty after batch dequeue")
	})

	t.Run("Update Strategy Selection", func(t *testing.T) {
		// Create component tree
		root := NewComponentManager("Root")
		child1 := NewComponentManager("Child1")
		child2 := NewComponentManager("Child2")
		grandchild1 := NewComponentManager("Grandchild1")
		grandchild2 := NewComponentManager("Grandchild2")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild1)
		child2.AddChild(grandchild2)

		// Test breadth-first update order
		queue := NewUpdateQueue()
		queue.EnqueueTree(root, UpdateStrategyBreadthFirst)

		// Expected order: root, child1, child2, grandchild1, grandchild2
		expectedBFSOrder := []string{"Root", "Child1", "Child2", "Grandchild1", "Grandchild2"}
		for i, expected := range expectedBFSOrder {
			comp := queue.Dequeue()
			assert.Equal(t, expected, comp.GetName(), "Component %d in BFS order should be %s", i, expected)
		}

		// Test depth-first update order
		queue = NewUpdateQueue()
		queue.EnqueueTree(root, UpdateStrategyDepthFirst)

		// Expected order: root, child1, grandchild1, child2, grandchild2
		expectedDFSOrder := []string{"Root", "Child1", "Grandchild1", "Child2", "Grandchild2"}
		for i, expected := range expectedDFSOrder {
			comp := queue.Dequeue()
			assert.Equal(t, expected, comp.GetName(), "Component %d in DFS order should be %s", i, expected)
		}
	})

	t.Run("Update Deadlines", func(t *testing.T) {
		// Create a new update queue with deadlines
		queue := NewUpdateQueue()

		// Create components
		comp1 := NewComponentManager("Component1")
		comp2 := NewComponentManager("Component2")

		// Set different deadlines
		now := time.Now()
		queue.EnqueueWithDeadline(comp1, UpdatePriorityMedium, now.Add(100*time.Millisecond))
		queue.EnqueueWithDeadline(comp2, UpdatePriorityMedium, now.Add(10*time.Millisecond))

		// Despite same priority, earlier deadline should be first
		dequeuedComp := queue.Dequeue()
		assert.Equal(t, comp2, dequeuedComp, "Component with earlier deadline should be dequeued first")
	})

	t.Run("Parent-Child Coordination", func(t *testing.T) {
		// Create component tree
		parent := NewComponentManager("Parent")
		child1 := NewComponentManager("Child1")
		child2 := NewComponentManager("Child2")

		parent.AddChild(child1)
		parent.AddChild(child2)

		// Create queue with parent-first coordination
		queue := NewUpdateQueue()
		queue.SetParentChildCoordination(UpdateCoordinationParentFirst)

		// Enqueue child first, but parent should still come first in the queue
		queue.Enqueue(child1, UpdatePriorityHigh)
		queue.Enqueue(parent, UpdatePriorityMedium)

		// Despite lower priority, parent should come first
		dequeuedComp := queue.Dequeue()
		assert.Equal(t, parent, dequeuedComp, "Parent should be dequeued before children with parent-first coordination")

		// Reset and test child-first coordination
		queue = NewUpdateQueue()
		queue.SetParentChildCoordination(UpdateCoordinationChildrenFirst)

		queue.Enqueue(parent, UpdatePriorityHigh)
		queue.Enqueue(child1, UpdatePriorityMedium)

		// Despite lower priority, child should come first
		dequeuedComp = queue.Dequeue()
		assert.Equal(t, child1, dequeuedComp, "Child should be dequeued before parent with children-first coordination")
	})
}
