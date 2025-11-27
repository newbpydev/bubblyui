package testutil

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewCommandQueueInspector tests inspector creation
func TestNewCommandQueueInspector(t *testing.T) {
	tests := []struct {
		name  string
		queue *bubbly.CommandQueue
	}{
		{
			name:  "with valid queue",
			queue: bubbly.NewCommandQueue(),
		},
		{
			name:  "with nil queue",
			queue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := NewCommandQueueInspector(tt.queue)
			assert.NotNil(t, inspector)
		})
	}
}

// TestCommandQueueInspector_Len tests length reporting
func TestCommandQueueInspector_Len(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*bubbly.CommandQueue)
		expected int
	}{
		{
			name:     "empty queue",
			setup:    func(q *bubbly.CommandQueue) {},
			expected: 0,
		},
		{
			name: "single command",
			setup: func(q *bubbly.CommandQueue) {
				q.Enqueue(func() tea.Msg { return nil })
			},
			expected: 1,
		},
		{
			name: "multiple commands",
			setup: func(q *bubbly.CommandQueue) {
				q.Enqueue(func() tea.Msg { return nil })
				q.Enqueue(func() tea.Msg { return nil })
				q.Enqueue(func() tea.Msg { return nil })
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := bubbly.NewCommandQueue()
			tt.setup(queue)

			inspector := NewCommandQueueInspector(queue)
			assert.Equal(t, tt.expected, inspector.Len())
		})
	}
}

// TestCommandQueueInspector_Len_NilQueue tests nil queue handling
func TestCommandQueueInspector_Len_NilQueue(t *testing.T) {
	inspector := NewCommandQueueInspector(nil)
	assert.Equal(t, 0, inspector.Len())
}

// TestCommandQueueInspector_Peek tests peeking at next command
func TestCommandQueueInspector_Peek(t *testing.T) {
	t.Run("empty queue returns nil", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()
		inspector := NewCommandQueueInspector(queue)

		cmd := inspector.Peek()
		assert.Nil(t, cmd)
	})

	t.Run("returns first command without removing", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()

		// Enqueue commands
		cmd1 := func() tea.Msg { return "first" }
		cmd2 := func() tea.Msg { return "second" }
		queue.Enqueue(cmd1)
		queue.Enqueue(cmd2)

		inspector := NewCommandQueueInspector(queue)

		// Peek should return first command
		peeked := inspector.Peek()
		assert.NotNil(t, peeked)

		// Queue should still have 2 commands
		assert.Equal(t, 2, inspector.Len())

		// Peeking again should return same command
		peeked2 := inspector.Peek()
		assert.NotNil(t, peeked2)
	})

	t.Run("nil queue returns nil", func(t *testing.T) {
		inspector := NewCommandQueueInspector(nil)
		cmd := inspector.Peek()
		assert.Nil(t, cmd)
	})
}

// TestCommandQueueInspector_GetAll tests getting all commands
func TestCommandQueueInspector_GetAll(t *testing.T) {
	t.Run("empty queue returns nil", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()
		inspector := NewCommandQueueInspector(queue)

		cmds := inspector.GetAll()
		assert.Nil(t, cmds)
	})

	t.Run("returns all commands without removing", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()

		// Enqueue commands
		cmd1 := func() tea.Msg { return "first" }
		cmd2 := func() tea.Msg { return "second" }
		cmd3 := func() tea.Msg { return "third" }
		queue.Enqueue(cmd1)
		queue.Enqueue(cmd2)
		queue.Enqueue(cmd3)

		inspector := NewCommandQueueInspector(queue)

		// GetAll should return all commands
		cmds := inspector.GetAll()
		assert.NotNil(t, cmds)
		assert.Equal(t, 3, len(cmds))

		// Queue should still have 3 commands
		assert.Equal(t, 3, inspector.Len())
	})

	t.Run("nil queue returns nil", func(t *testing.T) {
		inspector := NewCommandQueueInspector(nil)
		cmds := inspector.GetAll()
		assert.Nil(t, cmds)
	})
}

// TestCommandQueueInspector_Clear tests clearing the queue
func TestCommandQueueInspector_Clear(t *testing.T) {
	t.Run("clears all commands", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()

		// Enqueue commands
		queue.Enqueue(func() tea.Msg { return nil })
		queue.Enqueue(func() tea.Msg { return nil })
		queue.Enqueue(func() tea.Msg { return nil })

		inspector := NewCommandQueueInspector(queue)
		assert.Equal(t, 3, inspector.Len())

		// Clear
		inspector.Clear()
		assert.Equal(t, 0, inspector.Len())
	})

	t.Run("idempotent on empty queue", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()
		inspector := NewCommandQueueInspector(queue)

		inspector.Clear()
		assert.Equal(t, 0, inspector.Len())

		inspector.Clear()
		assert.Equal(t, 0, inspector.Len())
	})

	t.Run("nil queue is no-op", func(t *testing.T) {
		inspector := NewCommandQueueInspector(nil)
		inspector.Clear() // Should not panic
	})
}

// TestCommandQueueInspector_AssertEnqueued tests assertion helper
func TestCommandQueueInspector_AssertEnqueued(t *testing.T) {
	t.Run("passes with correct count", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()
		queue.Enqueue(func() tea.Msg { return nil })
		queue.Enqueue(func() tea.Msg { return nil })

		inspector := NewCommandQueueInspector(queue)

		// Use mock testing.T to capture failures
		mockT := &mockTestingT{}
		inspector.AssertEnqueued(mockT, 2)

		assert.False(t, mockT.failed, "assertion should pass")
	})

	t.Run("fails with wrong count", func(t *testing.T) {
		queue := bubbly.NewCommandQueue()
		queue.Enqueue(func() tea.Msg { return nil })

		inspector := NewCommandQueueInspector(queue)

		// Use mock testing.T to capture failures
		mockT := &mockTestingT{}
		inspector.AssertEnqueued(mockT, 3)

		assert.True(t, mockT.failed, "assertion should fail")
		assert.NotEmpty(t, mockT.errors, "should have error messages")
		errorMsg := mockT.errors[0]
		assert.Contains(t, errorMsg, "expected 3 commands")
		assert.Contains(t, errorMsg, "got 1")
	})

	t.Run("nil queue is treated as empty", func(t *testing.T) {
		inspector := NewCommandQueueInspector(nil)

		mockT := &mockTestingT{}
		inspector.AssertEnqueued(mockT, 0)

		assert.False(t, mockT.failed, "assertion should pass for nil queue with count 0")
	})
}

// TestCommandQueueInspector_ThreadSafety tests concurrent access
func TestCommandQueueInspector_ThreadSafety(t *testing.T) {
	queue := bubbly.NewCommandQueue()
	inspector := NewCommandQueueInspector(queue)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent enqueues
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			queue.Enqueue(func() tea.Msg { return nil })
		}
	}()

	// Concurrent reads
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = inspector.Len()
			_ = inspector.Peek()
			_ = inspector.GetAll()
		}
	}()

	// Concurrent clears
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			inspector.Clear()
		}
	}()

	wg.Wait()

	// Should not panic and queue should be in valid state
	assert.GreaterOrEqual(t, inspector.Len(), 0)
}

// TestCommandQueueInspector_Integration tests integration with TestHarness
func TestCommandQueueInspector_Integration(t *testing.T) {
	// This test verifies that CommandQueueInspector can be used
	// with a real component's command queue

	queue := bubbly.NewCommandQueue()
	inspector := NewCommandQueueInspector(queue)

	// Simulate component generating commands
	queue.Enqueue(func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "test-component",
			RefID:       "count",
		}
	})

	// Inspector should see the command
	assert.Equal(t, 1, inspector.Len())

	cmd := inspector.Peek()
	assert.NotNil(t, cmd)

	// Execute command to verify it's valid
	msg := cmd()
	stateMsg, ok := msg.(bubbly.StateChangedMsg)
	assert.True(t, ok)
	assert.Equal(t, "test-component", stateMsg.ComponentID)
	assert.Equal(t, "count", stateMsg.RefID)
}

// Note: mockTestingT is defined in assertions_state_test.go and reused here
