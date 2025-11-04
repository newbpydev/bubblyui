package commands

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestNewCommandQueue tests queue initialization
func TestNewCommandQueue(t *testing.T) {
	queue := NewCommandQueue()

	assert.NotNil(t, queue, "NewCommandQueue should return non-nil queue")
	assert.Equal(t, 0, queue.Len(), "New queue should be empty")
	assert.NotNil(t, queue.commands, "Commands slice should be initialized")
}

// TestCommandQueue_Enqueue tests adding commands to the queue
func TestCommandQueue_Enqueue(t *testing.T) {
	tests := []struct {
		name           string
		commands       []tea.Cmd
		expectedLen    int
		description    string
	}{
		{
			name: "single command",
			commands: []tea.Cmd{
				func() tea.Msg { return "msg1" },
			},
			expectedLen: 1,
			description: "Should enqueue single command",
		},
		{
			name: "multiple commands",
			commands: []tea.Cmd{
				func() tea.Msg { return "msg1" },
				func() tea.Msg { return "msg2" },
				func() tea.Msg { return "msg3" },
			},
			expectedLen: 3,
			description: "Should enqueue multiple commands",
		},
		{
			name: "nil command ignored",
			commands: []tea.Cmd{
				func() tea.Msg { return "msg1" },
				nil,
				func() tea.Msg { return "msg2" },
			},
			expectedLen: 2,
			description: "Should ignore nil commands",
		},
		{
			name:        "all nil commands",
			commands:    []tea.Cmd{nil, nil, nil},
			expectedLen: 0,
			description: "Should ignore all nil commands",
		},
		{
			name:        "empty list",
			commands:    []tea.Cmd{},
			expectedLen: 0,
			description: "Should handle empty command list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()

			for _, cmd := range tt.commands {
				queue.Enqueue(cmd)
			}

			assert.Equal(t, tt.expectedLen, queue.Len(), tt.description)
		})
	}
}

// TestCommandQueue_DrainAll tests draining all commands from queue
func TestCommandQueue_DrainAll(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*CommandQueue) []tea.Cmd
		expectedLen int
		description string
	}{
		{
			name: "drain single command",
			setup: func(q *CommandQueue) []tea.Cmd {
				cmd := func() tea.Msg { return "msg1" }
				q.Enqueue(cmd)
				return []tea.Cmd{cmd}
			},
			expectedLen: 1,
			description: "Should drain single command",
		},
		{
			name: "drain multiple commands",
			setup: func(q *CommandQueue) []tea.Cmd {
				cmds := []tea.Cmd{
					func() tea.Msg { return "msg1" },
					func() tea.Msg { return "msg2" },
					func() tea.Msg { return "msg3" },
				}
				for _, cmd := range cmds {
					q.Enqueue(cmd)
				}
				return cmds
			},
			expectedLen: 3,
			description: "Should drain all commands",
		},
		{
			name: "drain empty queue",
			setup: func(q *CommandQueue) []tea.Cmd {
				return nil
			},
			expectedLen: 0,
			description: "Should return nil for empty queue",
		},
		{
			name: "drain twice",
			setup: func(q *CommandQueue) []tea.Cmd {
				cmd := func() tea.Msg { return "msg1" }
				q.Enqueue(cmd)
				q.DrainAll() // First drain
				return nil   // Second drain should be empty
			},
			expectedLen: 0,
			description: "Should return nil on second drain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			tt.setup(queue)

			cmds := queue.DrainAll()

			if tt.expectedLen == 0 {
				assert.Nil(t, cmds, tt.description)
			} else {
				assert.Len(t, cmds, tt.expectedLen, tt.description)
			}

			// Queue should be empty after drain
			assert.Equal(t, 0, queue.Len(), "Queue should be empty after DrainAll")
		})
	}
}

// TestCommandQueue_DrainAll_ClearsQueue verifies queue is cleared after drain
func TestCommandQueue_DrainAll_ClearsQueue(t *testing.T) {
	queue := NewCommandQueue()

	// Add commands
	queue.Enqueue(func() tea.Msg { return "msg1" })
	queue.Enqueue(func() tea.Msg { return "msg2" })
	assert.Equal(t, 2, queue.Len(), "Should have 2 commands")

	// Drain
	cmds := queue.DrainAll()
	assert.Len(t, cmds, 2, "Should return 2 commands")
	assert.Equal(t, 0, queue.Len(), "Queue should be empty")

	// Add more commands after drain
	queue.Enqueue(func() tea.Msg { return "msg3" })
	assert.Equal(t, 1, queue.Len(), "Should have 1 command after re-adding")

	// Drain again
	cmds = queue.DrainAll()
	assert.Len(t, cmds, 1, "Should return 1 command")
	assert.Equal(t, 0, queue.Len(), "Queue should be empty again")
}

// TestCommandQueue_Len tests queue length tracking
func TestCommandQueue_Len(t *testing.T) {
	tests := []struct {
		name        string
		operations  func(*CommandQueue)
		expectedLen int
		description string
	}{
		{
			name:        "empty queue",
			operations:  func(q *CommandQueue) {},
			expectedLen: 0,
			description: "Empty queue should have length 0",
		},
		{
			name: "after enqueue",
			operations: func(q *CommandQueue) {
				q.Enqueue(func() tea.Msg { return "msg" })
			},
			expectedLen: 1,
			description: "Should have length 1 after enqueue",
		},
		{
			name: "after multiple enqueues",
			operations: func(q *CommandQueue) {
				for i := 0; i < 5; i++ {
					q.Enqueue(func() tea.Msg { return "msg" })
				}
			},
			expectedLen: 5,
			description: "Should have length 5 after 5 enqueues",
		},
		{
			name: "after enqueue and drain",
			operations: func(q *CommandQueue) {
				q.Enqueue(func() tea.Msg { return "msg" })
				q.DrainAll()
			},
			expectedLen: 0,
			description: "Should have length 0 after drain",
		},
		{
			name: "after clear",
			operations: func(q *CommandQueue) {
				q.Enqueue(func() tea.Msg { return "msg1" })
				q.Enqueue(func() tea.Msg { return "msg2" })
				q.Clear()
			},
			expectedLen: 0,
			description: "Should have length 0 after clear",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			tt.operations(queue)

			assert.Equal(t, tt.expectedLen, queue.Len(), tt.description)
		})
	}
}

// TestCommandQueue_Clear tests clearing the queue
func TestCommandQueue_Clear(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*CommandQueue)
		description string
	}{
		{
			name:        "clear empty queue",
			setup:       func(q *CommandQueue) {},
			description: "Should handle clearing empty queue",
		},
		{
			name: "clear single command",
			setup: func(q *CommandQueue) {
				q.Enqueue(func() tea.Msg { return "msg" })
			},
			description: "Should clear single command",
		},
		{
			name: "clear multiple commands",
			setup: func(q *CommandQueue) {
				for i := 0; i < 10; i++ {
					q.Enqueue(func() tea.Msg { return "msg" })
				}
			},
			description: "Should clear multiple commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			tt.setup(queue)

			queue.Clear()

			assert.Equal(t, 0, queue.Len(), tt.description)

			// Verify can still use queue after clear
			queue.Enqueue(func() tea.Msg { return "new" })
			assert.Equal(t, 1, queue.Len(), "Should be able to enqueue after clear")
		})
	}
}

// TestCommandQueue_ThreadSafety tests concurrent access
func TestCommandQueue_ThreadSafety(t *testing.T) {
	tests := []struct {
		name        string
		goroutines  int
		opsPerGo    int
		description string
	}{
		{
			name:        "10 goroutines, 100 ops each",
			goroutines:  10,
			opsPerGo:    100,
			description: "Should handle 10 concurrent goroutines",
		},
		{
			name:        "100 goroutines, 10 ops each",
			goroutines:  100,
			opsPerGo:    10,
			description: "Should handle 100 concurrent goroutines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			var wg sync.WaitGroup

			// Concurrent enqueues
			wg.Add(tt.goroutines)
			for i := 0; i < tt.goroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < tt.opsPerGo; j++ {
						queue.Enqueue(func() tea.Msg { return "msg" })
					}
				}()
			}
			wg.Wait()

			expectedLen := tt.goroutines * tt.opsPerGo
			assert.Equal(t, expectedLen, queue.Len(), tt.description)

			// Concurrent drains
			wg.Add(tt.goroutines)
			allCmds := make([][]tea.Cmd, tt.goroutines)
			for i := 0; i < tt.goroutines; i++ {
				go func(idx int) {
					defer wg.Done()
					allCmds[idx] = queue.DrainAll()
				}(i)
			}
			wg.Wait()

			// Verify all commands were drained
			totalDrained := 0
			for _, cmds := range allCmds {
				totalDrained += len(cmds)
			}
			assert.Equal(t, expectedLen, totalDrained, "All commands should be drained")
			assert.Equal(t, 0, queue.Len(), "Queue should be empty after concurrent drains")
		})
	}
}

// TestCommandQueue_ConcurrentOperations tests mixed concurrent operations
func TestCommandQueue_ConcurrentOperations(t *testing.T) {
	queue := NewCommandQueue()
	var wg sync.WaitGroup

	// Mix of operations
	operations := 100
	wg.Add(operations * 4) // enqueue, drain, len, clear

	// Concurrent enqueues
	for i := 0; i < operations; i++ {
		go func() {
			defer wg.Done()
			queue.Enqueue(func() tea.Msg { return "msg" })
		}()
	}

	// Concurrent drains
	for i := 0; i < operations; i++ {
		go func() {
			defer wg.Done()
			queue.DrainAll()
		}()
	}

	// Concurrent length checks
	for i := 0; i < operations; i++ {
		go func() {
			defer wg.Done()
			_ = queue.Len()
		}()
	}

	// Concurrent clears
	for i := 0; i < operations; i++ {
		go func() {
			defer wg.Done()
			queue.Clear()
		}()
	}

	wg.Wait()

	// Should not panic and queue should be in valid state
	assert.NotNil(t, queue, "Queue should still be valid")
	assert.True(t, queue.Len() >= 0, "Length should be non-negative")
}

// TestCommandQueue_CommandExecution verifies commands can be executed
func TestCommandQueue_CommandExecution(t *testing.T) {
	queue := NewCommandQueue()

	// Create commands that return specific messages
	msg1 := "message1"
	msg2 := "message2"
	msg3 := "message3"

	queue.Enqueue(func() tea.Msg { return msg1 })
	queue.Enqueue(func() tea.Msg { return msg2 })
	queue.Enqueue(func() tea.Msg { return msg3 })

	cmds := queue.DrainAll()
	assert.Len(t, cmds, 3, "Should have 3 commands")

	// Execute commands and verify messages
	messages := make([]string, 0, 3)
	for _, cmd := range cmds {
		msg := cmd()
		messages = append(messages, msg.(string))
	}

	assert.Contains(t, messages, msg1, "Should contain msg1")
	assert.Contains(t, messages, msg2, "Should contain msg2")
	assert.Contains(t, messages, msg3, "Should contain msg3")
}

// TestCommandQueue_NilCommandHandling tests nil command handling
func TestCommandQueue_NilCommandHandling(t *testing.T) {
	queue := NewCommandQueue()

	// Enqueue nil commands
	queue.Enqueue(nil)
	queue.Enqueue(nil)
	queue.Enqueue(nil)

	assert.Equal(t, 0, queue.Len(), "Nil commands should not be enqueued")

	cmds := queue.DrainAll()
	assert.Nil(t, cmds, "Should return nil for empty queue")
}

// TestCommandQueue_PreAllocation tests pre-allocated capacity
func TestCommandQueue_PreAllocation(t *testing.T) {
	queue := NewCommandQueue()

	// Verify initial capacity (implementation detail, but useful to know)
	assert.NotNil(t, queue.commands, "Commands slice should be initialized")

	// Add commands up to and beyond initial capacity
	for i := 0; i < 20; i++ {
		queue.Enqueue(func() tea.Msg { return "msg" })
	}

	assert.Equal(t, 20, queue.Len(), "Should handle growth beyond initial capacity")

	cmds := queue.DrainAll()
	assert.Len(t, cmds, 20, "Should drain all commands")
}

// TestCommandQueue_MultipleEnqueueDrainCycles tests repeated cycles
func TestCommandQueue_MultipleEnqueueDrainCycles(t *testing.T) {
	queue := NewCommandQueue()

	for cycle := 0; cycle < 10; cycle++ {
		// Enqueue
		for i := 0; i < 5; i++ {
			queue.Enqueue(func() tea.Msg { return "msg" })
		}
		assert.Equal(t, 5, queue.Len(), "Should have 5 commands in cycle %d", cycle)

		// Drain
		cmds := queue.DrainAll()
		assert.Len(t, cmds, 5, "Should drain 5 commands in cycle %d", cycle)
		assert.Equal(t, 0, queue.Len(), "Should be empty after drain in cycle %d", cycle)
	}
}
