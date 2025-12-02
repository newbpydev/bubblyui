package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseQueue_InitialItems tests that initial items are set correctly.
func TestUseQueue_InitialItems(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		expected []string
	}{
		{
			name:     "empty initial",
			initial:  []string{},
			expected: []string{},
		},
		{
			name:     "nil initial",
			initial:  nil,
			expected: []string{},
		},
		{
			name:     "single item",
			initial:  []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "multiple items",
			initial:  []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			queue := UseQueue(ctx, tt.initial)

			assert.Equal(t, tt.expected, queue.Items.GetTyped())
			assert.Equal(t, len(tt.expected), queue.Size.GetTyped())
			assert.Equal(t, len(tt.expected) == 0, queue.IsEmpty.GetTyped())
		})
	}
}

// TestUseQueue_Enqueue tests adding items to the back of the queue.
func TestUseQueue_Enqueue(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		enqueue  int
		expected []int
	}{
		{
			name:     "enqueue to empty",
			initial:  []int{},
			enqueue:  1,
			expected: []int{1},
		},
		{
			name:     "enqueue to non-empty",
			initial:  []int{1, 2},
			enqueue:  3,
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			queue := UseQueue(ctx, tt.initial)

			queue.Enqueue(tt.enqueue)

			assert.Equal(t, tt.expected, queue.Items.GetTyped())
			assert.Equal(t, len(tt.expected), queue.Size.GetTyped())
		})
	}
}

// TestUseQueue_Dequeue tests removing and returning the front item.
func TestUseQueue_Dequeue(t *testing.T) {
	tests := []struct {
		name         string
		initial      []string
		expectedItem string
		expectedOk   bool
		expectedList []string
	}{
		{
			name:         "dequeue from empty",
			initial:      []string{},
			expectedItem: "",
			expectedOk:   false,
			expectedList: []string{},
		},
		{
			name:         "dequeue from single",
			initial:      []string{"a"},
			expectedItem: "a",
			expectedOk:   true,
			expectedList: []string{},
		},
		{
			name:         "dequeue from multiple",
			initial:      []string{"a", "b", "c"},
			expectedItem: "a",
			expectedOk:   true,
			expectedList: []string{"b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			queue := UseQueue(ctx, tt.initial)

			item, ok := queue.Dequeue()

			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedList, queue.Items.GetTyped())
		})
	}
}

// TestUseQueue_Peek tests returning the front item without removing.
func TestUseQueue_Peek(t *testing.T) {
	tests := []struct {
		name         string
		initial      []int
		expectedItem int
		expectedOk   bool
		expectedList []int
	}{
		{
			name:         "peek empty",
			initial:      []int{},
			expectedItem: 0,
			expectedOk:   false,
			expectedList: []int{},
		},
		{
			name:         "peek single",
			initial:      []int{1},
			expectedItem: 1,
			expectedOk:   true,
			expectedList: []int{1},
		},
		{
			name:         "peek multiple",
			initial:      []int{1, 2, 3},
			expectedItem: 1,
			expectedOk:   true,
			expectedList: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			queue := UseQueue(ctx, tt.initial)

			item, ok := queue.Peek()

			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedOk, ok)
			// Peek should NOT modify the queue
			assert.Equal(t, tt.expectedList, queue.Items.GetTyped())
		})
	}
}

// TestUseQueue_Clear tests clearing all items.
func TestUseQueue_Clear(t *testing.T) {
	tests := []struct {
		name    string
		initial []int
	}{
		{
			name:    "clear empty",
			initial: []int{},
		},
		{
			name:    "clear non-empty",
			initial: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			queue := UseQueue(ctx, tt.initial)

			queue.Clear()

			assert.Equal(t, []int{}, queue.Items.GetTyped())
			assert.Equal(t, 0, queue.Size.GetTyped())
			assert.True(t, queue.IsEmpty.GetTyped())
		})
	}
}

// TestUseQueue_Front tests the Front computed value.
func TestUseQueue_Front(t *testing.T) {
	tests := []struct {
		name          string
		initial       []string
		expectedFront *string
	}{
		{
			name:          "front of empty",
			initial:       []string{},
			expectedFront: nil,
		},
		{
			name:          "front of single",
			initial:       []string{"a"},
			expectedFront: strPtr("a"),
		},
		{
			name:          "front of multiple",
			initial:       []string{"a", "b", "c"},
			expectedFront: strPtr("a"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			queue := UseQueue(ctx, tt.initial)

			front := queue.Front.GetTyped()

			if tt.expectedFront == nil {
				assert.Nil(t, front)
			} else {
				require.NotNil(t, front)
				assert.Equal(t, *tt.expectedFront, *front)
			}
		})
	}
}

// strPtr is a helper to create a pointer to a string.
func strPtr(s string) *string {
	return &s
}

// TestUseQueue_Size tests the Size computed value.
func TestUseQueue_Size(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{1, 2, 3})

	assert.Equal(t, 3, queue.Size.GetTyped())

	queue.Enqueue(4)
	assert.Equal(t, 4, queue.Size.GetTyped())

	queue.Dequeue()
	assert.Equal(t, 3, queue.Size.GetTyped())

	queue.Clear()
	assert.Equal(t, 0, queue.Size.GetTyped())
}

// TestUseQueue_IsEmpty tests the IsEmpty computed value.
func TestUseQueue_IsEmpty(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{})

	assert.True(t, queue.IsEmpty.GetTyped())

	queue.Enqueue(1)
	assert.False(t, queue.IsEmpty.GetTyped())

	queue.Dequeue()
	assert.True(t, queue.IsEmpty.GetTyped())
}

// TestUseQueue_FIFOOrder tests that queue maintains FIFO order.
func TestUseQueue_FIFOOrder(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{})

	// Enqueue items
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	// Dequeue should return in FIFO order
	item1, ok1 := queue.Dequeue()
	assert.True(t, ok1)
	assert.Equal(t, 1, item1)

	item2, ok2 := queue.Dequeue()
	assert.True(t, ok2)
	assert.Equal(t, 2, item2)

	item3, ok3 := queue.Dequeue()
	assert.True(t, ok3)
	assert.Equal(t, 3, item3)

	// Queue should be empty now
	_, ok4 := queue.Dequeue()
	assert.False(t, ok4)
}

// TestUseQueue_GenericTypes tests that UseQueue works with various types.
func TestUseQueue_GenericTypes(t *testing.T) {
	t.Run("int queue", func(t *testing.T) {
		ctx := createTestContext()
		queue := UseQueue(ctx, []int{1, 2})
		queue.Enqueue(3)
		assert.Equal(t, []int{1, 2, 3}, queue.Items.GetTyped())
	})

	t.Run("string queue", func(t *testing.T) {
		ctx := createTestContext()
		queue := UseQueue(ctx, []string{"a", "b"})
		queue.Enqueue("c")
		assert.Equal(t, []string{"a", "b", "c"}, queue.Items.GetTyped())
	})

	t.Run("struct queue", func(t *testing.T) {
		type Task struct {
			ID   int
			Name string
		}
		ctx := createTestContext()
		queue := UseQueue(ctx, []Task{{ID: 1, Name: "first"}})
		queue.Enqueue(Task{ID: 2, Name: "second"})
		assert.Equal(t, 2, queue.Size.GetTyped())

		item, ok := queue.Dequeue()
		require.True(t, ok)
		assert.Equal(t, "first", item.Name)
	})
}

// TestUseQueue_NilContext tests that UseQueue handles nil context.
func TestUseQueue_NilContext(t *testing.T) {
	// Should not panic with nil context
	queue := UseQueue[int](nil, []int{1, 2, 3})
	assert.NotNil(t, queue)
	assert.Equal(t, []int{1, 2, 3}, queue.Items.GetTyped())
}

// TestUseQueue_ChainedOperations tests multiple operations in sequence.
func TestUseQueue_ChainedOperations(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{})

	// Enqueue items
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)
	assert.Equal(t, []int{1, 2, 3}, queue.Items.GetTyped())

	// Peek should not modify
	front, ok := queue.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, front)
	assert.Equal(t, 3, queue.Size.GetTyped())

	// Dequeue first
	item, ok := queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 1, item)
	assert.Equal(t, []int{2, 3}, queue.Items.GetTyped())

	// Enqueue more
	queue.Enqueue(4)
	queue.Enqueue(5)
	assert.Equal(t, []int{2, 3, 4, 5}, queue.Items.GetTyped())

	// Clear
	queue.Clear()
	assert.True(t, queue.IsEmpty.GetTyped())
}

// TestUseQueue_Reactivity tests that Items ref triggers updates.
func TestUseQueue_Reactivity(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{1, 2, 3})

	// Track changes via Watch
	changes := 0
	bubbly.Watch(queue.Items, func(newVal, oldVal []int) {
		changes++
	})

	queue.Enqueue(4)
	queue.Dequeue()
	queue.Clear()

	// Each operation should trigger a change
	assert.Equal(t, 3, changes)
}

// TestUseQueue_FrontUpdatesAfterOperations tests Front computed updates correctly.
func TestUseQueue_FrontUpdatesAfterOperations(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []string{"a", "b", "c"})

	// Initial front
	front := queue.Front.GetTyped()
	require.NotNil(t, front)
	assert.Equal(t, "a", *front)

	// After dequeue, front should update
	queue.Dequeue()
	front = queue.Front.GetTyped()
	require.NotNil(t, front)
	assert.Equal(t, "b", *front)

	// After clear, front should be nil
	queue.Clear()
	front = queue.Front.GetTyped()
	assert.Nil(t, front)

	// After enqueue, front should update
	queue.Enqueue("x")
	front = queue.Front.GetTyped()
	require.NotNil(t, front)
	assert.Equal(t, "x", *front)
}

// TestUseQueue_CreateSharedPattern tests integration with CreateShared.
func TestUseQueue_CreateSharedPattern(t *testing.T) {
	// Create a shared queue factory
	UseSharedQueue := CreateShared(func(ctx *bubbly.Context) *QueueReturn[string] {
		return UseQueue(ctx, []string{"shared"})
	})

	ctx := createTestContext()
	queue1 := UseSharedQueue(ctx)
	queue2 := UseSharedQueue(ctx)

	// Both should reference the same instance
	assert.Same(t, queue1, queue2)

	// Modifications through one should be visible through the other
	queue1.Enqueue("item")
	assert.Equal(t, []string{"shared", "item"}, queue2.Items.GetTyped())
}

// TestUseQueue_MultipleDequeues tests dequeuing until empty.
func TestUseQueue_MultipleDequeues(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{1, 2, 3})

	// Dequeue all items
	for i := 1; i <= 3; i++ {
		item, ok := queue.Dequeue()
		assert.True(t, ok)
		assert.Equal(t, i, item)
	}

	// Queue should be empty
	assert.True(t, queue.IsEmpty.GetTyped())

	// Further dequeue should return false
	_, ok := queue.Dequeue()
	assert.False(t, ok)
}

// TestUseQueue_PeekDoesNotModify tests that Peek is idempotent.
func TestUseQueue_PeekDoesNotModify(t *testing.T) {
	ctx := createTestContext()
	queue := UseQueue(ctx, []int{1, 2, 3})

	// Multiple peeks should return the same value
	for i := 0; i < 5; i++ {
		item, ok := queue.Peek()
		assert.True(t, ok)
		assert.Equal(t, 1, item)
	}

	// Queue should still have all items
	assert.Equal(t, []int{1, 2, 3}, queue.Items.GetTyped())
	assert.Equal(t, 3, queue.Size.GetTyped())
}
