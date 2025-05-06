package core

import (
	"container/heap"
	"sync"
	"time"
)

// Update priority constants
const (
	UpdatePriorityLow    = 10
	UpdatePriorityMedium = 50
	UpdatePriorityHigh   = 100
	UpdatePriorityUrgent = 150
)

// Update strategy constants
const (
	UpdateStrategyBreadthFirst = iota
	UpdateStrategyDepthFirst
)

// Update coordination type constants
const (
	UpdateCoordinationPriorityOnly  = iota // Default - only consider priority
	UpdateCoordinationParentFirst          // Parents update before children regardless of priority
	UpdateCoordinationChildrenFirst        // Children update before parents regardless of priority
)

// updateQueueItem represents a component in the update queue
type updateQueueItem struct {
	component *ComponentManager
	priority  int
	deadline  time.Time
	batchID   string
	index     int // Used by heap.Interface
}

// updatePriorityQueue implements heap.Interface for priority queue of update items
type updatePriorityQueue []*updateQueueItem

// Implement heap.Interface for updatePriorityQueue
func (pq updatePriorityQueue) Len() int { return len(pq) }

func (pq updatePriorityQueue) Less(i, j int) bool {
	// First compare deadlines if both have non-zero deadlines
	if !pq[i].deadline.IsZero() && !pq[j].deadline.IsZero() {
		if pq[i].deadline.Before(pq[j].deadline) {
			return true
		}
		if pq[j].deadline.Before(pq[i].deadline) {
			return false
		}
	}

	// If deadlines are equal or not set, compare priorities
	return pq[i].priority > pq[j].priority // Higher priority first
}

func (pq updatePriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *updatePriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*updateQueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *updatePriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// UpdateQueue manages the scheduling of component updates
type UpdateQueue struct {
	queue      updatePriorityQueue
	components map[*ComponentManager]*updateQueueItem
	batches    map[string][]*ComponentManager
	mutex      sync.RWMutex

	// Update strategy settings
	coordination int // Controls parent-child update ordering
}

// NewUpdateQueue creates a new update queue for scheduling component updates
func NewUpdateQueue() *UpdateQueue {
	uq := &UpdateQueue{
		queue:        make(updatePriorityQueue, 0),
		components:   make(map[*ComponentManager]*updateQueueItem),
		batches:      make(map[string][]*ComponentManager),
		coordination: UpdateCoordinationPriorityOnly,
	}
	heap.Init(&uq.queue)
	return uq
}

// Enqueue adds a component to the update queue with the specified priority
func (uq *UpdateQueue) Enqueue(component *ComponentManager, priority int) {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	// Check if component is already in the queue
	if item, exists := uq.components[component]; exists {
		// Only update priority if the new priority is higher
		if priority > item.priority {
			item.priority = priority
			heap.Fix(&uq.queue, item.index)
		}
		return
	}

	// Add new component to the queue
	item := &updateQueueItem{
		component: component,
		priority:  priority,
		deadline:  time.Time{}, // Zero time (no deadline)
		index:     0,
	}

	// Apply parent-child coordination rules
	if uq.coordination != UpdateCoordinationPriorityOnly {
		item.priority = uq.adjustPriorityByCoordination(component, priority)
	}

	uq.components[component] = item
	heap.Push(&uq.queue, item)
}

// EnqueueWithDeadline adds a component to the update queue with a deadline
func (uq *UpdateQueue) EnqueueWithDeadline(component *ComponentManager, priority int, deadline time.Time) {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	// Check if component is already in the queue
	if item, exists := uq.components[component]; exists {
		// Update priority if higher or deadline if earlier
		if priority > item.priority || deadline.Before(item.deadline) {
			if priority > item.priority {
				item.priority = priority
			}
			if deadline.Before(item.deadline) || item.deadline.IsZero() {
				item.deadline = deadline
			}
			heap.Fix(&uq.queue, item.index)
		}
		return
	}

	// Add new component to the queue
	item := &updateQueueItem{
		component: component,
		priority:  priority,
		deadline:  deadline,
		index:     0,
	}

	// Apply parent-child coordination rules
	if uq.coordination != UpdateCoordinationPriorityOnly {
		item.priority = uq.adjustPriorityByCoordination(component, priority)
	}

	uq.components[component] = item
	heap.Push(&uq.queue, item)
}

// EnqueueBatched adds a component to the update queue with a batch ID
func (uq *UpdateQueue) EnqueueBatched(component *ComponentManager, priority int, batchID string) {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	// Add to batch tracking
	if _, exists := uq.batches[batchID]; !exists {
		uq.batches[batchID] = make([]*ComponentManager, 0)
	}

	// Only add to batch if not already there
	found := false
	for _, c := range uq.batches[batchID] {
		if c == component {
			found = true
			break
		}
	}

	if !found {
		uq.batches[batchID] = append(uq.batches[batchID], component)
	}

	// Add to queue with batch ID
	if item, exists := uq.components[component]; exists {
		// Update only if new priority is higher
		if priority > item.priority {
			item.priority = priority
			item.batchID = batchID
			heap.Fix(&uq.queue, item.index)
		} else if item.batchID == "" {
			// Set batch ID if not previously set
			item.batchID = batchID
		}
		return
	}

	// Add new component to the queue
	item := &updateQueueItem{
		component: component,
		priority:  priority,
		deadline:  time.Time{}, // Zero time (no deadline)
		batchID:   batchID,
		index:     0,
	}

	// Apply parent-child coordination rules
	if uq.coordination != UpdateCoordinationPriorityOnly {
		item.priority = uq.adjustPriorityByCoordination(component, priority)
	}

	uq.components[component] = item
	heap.Push(&uq.queue, item)
}

// Dequeue removes and returns the highest priority component from the queue
func (uq *UpdateQueue) Dequeue() *ComponentManager {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	if len(uq.queue) == 0 {
		return nil
	}

	item := heap.Pop(&uq.queue).(*updateQueueItem)
	delete(uq.components, item.component)

	// If component was part of a batch, don't remove it from the batch map yet
	// as we might need to retrieve the entire batch later

	return item.component
}

// DequeueBatch removes and returns all components in a batch
func (uq *UpdateQueue) DequeueBatch(batchID string) []*ComponentManager {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	// Get components in the batch
	components, exists := uq.batches[batchID]
	if !exists || len(components) == 0 {
		return nil
	}

	// Create result slice
	result := make([]*ComponentManager, len(components))
	copy(result, components)

	// Remove all components in the batch from the queue
	for i, component := range components {
		if item, exists := uq.components[component]; exists {
			// Remove from priority queue if still there
			if item.index >= 0 && item.index < len(uq.queue) {
				heap.Remove(&uq.queue, item.index)
			}
			delete(uq.components, component)
		}
		result[i] = component
	}

	// Remove batch
	delete(uq.batches, batchID)

	return result
}

// UpdatePriority changes the priority of a component in the queue
func (uq *UpdateQueue) UpdatePriority(component *ComponentManager, newPriority int) {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	if item, exists := uq.components[component]; exists {
		item.priority = newPriority
		heap.Fix(&uq.queue, item.index)
	}
}

// Length returns the number of components in the queue
func (uq *UpdateQueue) Length() int {
	uq.mutex.RLock()
	defer uq.mutex.RUnlock()

	return len(uq.queue)
}

// SetParentChildCoordination configures how parents and children are coordinated in the update queue
func (uq *UpdateQueue) SetParentChildCoordination(coordinationType int) {
	uq.mutex.Lock()
	defer uq.mutex.Unlock()

	if coordinationType >= UpdateCoordinationPriorityOnly &&
		coordinationType <= UpdateCoordinationChildrenFirst {
		uq.coordination = coordinationType
	}
}

// adjustPriorityByCoordination adjusts priority based on parent-child coordination
func (uq *UpdateQueue) adjustPriorityByCoordination(component *ComponentManager, priority int) int {
	// Calculate tree depth of the component
	depth := 0
	current := component
	for current.GetParent() != nil {
		depth++
		current = current.GetParent()
	}

	// Adjust priority based on coordination type
	switch uq.coordination {
	case UpdateCoordinationParentFirst:
		// Invert depth to give higher priority to parents (lower depth)
		// Use a large multiplier to ensure depth takes precedence over base priority
		return priority + (1000 * (1000 - depth)) // Higher number for parents (lower depth)

	case UpdateCoordinationChildrenFirst:
		// Use depth directly to give higher priority to children (higher depth)
		// Use a large multiplier to ensure depth takes precedence over base priority
		return priority + (1000 * depth) // Higher number for children (higher depth)

	default:
		// Use priority as-is
		return priority
	}
}

// EnqueueTree adds an entire component tree to the queue using the specified update strategy
func (uq *UpdateQueue) EnqueueTree(root *ComponentManager, strategy int) {
	switch strategy {
	case UpdateStrategyBreadthFirst:
		uq.enqueueTreeBFS(root)
	case UpdateStrategyDepthFirst:
		uq.enqueueTreeDFS(root)
	default:
		// Default to breadth-first
		uq.enqueueTreeBFS(root)
	}
}

// enqueueTreeBFS adds a component tree in breadth-first order
func (uq *UpdateQueue) enqueueTreeBFS(root *ComponentManager) {
	// First, collect all components in breadth-first order
	components := make([]*ComponentManager, 0)
	queue := []*ComponentManager{root}

	// Create the complete traversal order first
	for len(queue) > 0 {
		// Dequeue the next component
		current := queue[0]
		queue = queue[1:]

		// Add to ordered component list
		components = append(components, current)

		// Add all children to the BFS queue
		for _, child := range current.GetChildren() {
			queue = append(queue, child)
		}
	}

	// Now enqueue them in reverse order so they pop out in the correct order
	// This ensures the priority queue maintains the traversal order
	for i := len(components) - 1; i >= 0; i-- {
		// Higher index = lower priority number to ensure proper traversal order
		additionalPriority := len(components) - 1 - i
		uq.Enqueue(components[i], UpdatePriorityMedium+additionalPriority)
	}
}

// enqueueTreeDFS adds a component tree in depth-first order
func (uq *UpdateQueue) enqueueTreeDFS(root *ComponentManager) {
	// First, collect all components in depth-first order
	components := make([]*ComponentManager, 0)

	// Helper function to perform DFS traversal
	var dfs func(*ComponentManager)
	dfs = func(node *ComponentManager) {
		components = append(components, node)
		for _, child := range node.GetChildren() {
			dfs(child)
		}
	}

	// Perform DFS traversal starting at root
	dfs(root)

	// Now enqueue them in reverse order so they pop out in the correct order
	// This ensures the priority queue maintains the traversal order
	for i := len(components) - 1; i >= 0; i-- {
		// Higher index = lower priority number to ensure proper traversal order
		additionalPriority := len(components) - 1 - i
		uq.Enqueue(components[i], UpdatePriorityMedium+additionalPriority)
	}
}
