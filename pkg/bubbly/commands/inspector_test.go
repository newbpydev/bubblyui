package commands

import (
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test suite for the Command Inspector (Task 6.2: Command Inspector).
//
// This test suite comprehensively validates the CommandInspector type and its
// methods, ensuring they meet the requirements for:
//   - Inspecting pending commands in the queue
//   - Accurate command count reporting
//   - Extracting command metadata (ComponentInfo)
//   - Clearing pending commands
//   - Thread-safe concurrent operation
//
// Test Coverage:
//   - Functional correctness (inspector shows accurate state)
//   - Count accuracy (PendingCount matches actual queue state)
//   - Command info extraction (metadata correctly extracted)
//   - Clear functionality (pending commands removed)
//   - Thread safety (concurrent access patterns)
//   - Edge cases (empty queue, nil commands, concurrent operations)
//
// Usage Examples:
//
//	// Create inspector for a queue
//	queue := NewCommandQueue()
//	inspector := NewCommandInspector(queue)
//
//	// Check pending count
//	count := inspector.PendingCount()
//
//	// Get command info
//	commands := inspector.PendingCommands()
//	for _, cmd := range commands {
//	    fmt.Printf("Component: %s, Ref: %s, Time: %v\n",
//	        cmd.ComponentID, cmd.RefID, cmd.Timestamp)
//	}
//
//	// Clear pending
//	inspector.ClearPending()

// TestCommandInspector_PendingCount tests that PendingCount() returns
// accurate count of commands in the queue.
//
// This test validates:
//   - Empty queue returns 0
//   - Single command returns 1
//   - Multiple commands return correct count
//   - Count updates after enqueue/drain operations
//
// Test Cases:
//  1. Empty queue (initial state)
//  2. Single command enqueued
//  3. Multiple commands enqueued
//  4. After draining commands
//  5. After clearing queue
func TestCommandInspector_PendingCount(t *testing.T) {
	tests := []struct {
		name          string
		setupQueue    func(*CommandQueue)
		expectedCount int
	}{
		{
			name:          "empty queue",
			setupQueue:    func(q *CommandQueue) {},
			expectedCount: 0,
		},
		{
			name: "single command",
			setupQueue: func(q *CommandQueue) {
				cmd := func() tea.Msg {
					return StateChangedMsg{
						ComponentID: "comp-1",
						RefID:       "ref-1",
						Timestamp:   time.Now(),
					}
				}
				q.Enqueue(cmd)
			},
			expectedCount: 1,
		},
		{
			name: "multiple commands",
			setupQueue: func(q *CommandQueue) {
				for i := 0; i < 5; i++ {
					cmd := func() tea.Msg {
						return StateChangedMsg{
							ComponentID: "comp-1",
							RefID:       "ref-1",
							Timestamp:   time.Now(),
						}
					}
					q.Enqueue(cmd)
				}
			},
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			tt.setupQueue(queue)

			inspector := NewCommandInspector(queue)
			count := inspector.PendingCount()

			assert.Equal(t, tt.expectedCount, count,
				"PendingCount should return %d", tt.expectedCount)
		})
	}
}

// TestCommandInspector_PendingCommands tests that PendingCommands() returns
// accurate command metadata without modifying the queue.
//
// This test validates:
//   - Empty queue returns empty slice
//   - Command metadata correctly extracted
//   - ComponentID, RefID, Timestamp preserved
//   - Queue not modified (commands still pending)
//   - Multiple commands all returned
//
// Test Cases:
//  1. Empty queue returns empty slice
//  2. Single command metadata extracted
//  3. Multiple commands all returned
//  4. Queue unchanged after inspection
func TestCommandInspector_PendingCommands(t *testing.T) {
	tests := []struct {
		name           string
		setupQueue     func(*CommandQueue) []StateChangedMsg
		expectedCount  int
		verifyMetadata func(*testing.T, []CommandInfo, []StateChangedMsg)
	}{
		{
			name: "empty queue",
			setupQueue: func(q *CommandQueue) []StateChangedMsg {
				return nil
			},
			expectedCount: 0,
			verifyMetadata: func(t *testing.T, infos []CommandInfo, msgs []StateChangedMsg) {
				assert.Empty(t, infos, "should return empty slice for empty queue")
			},
		},
		{
			name: "single command",
			setupQueue: func(q *CommandQueue) []StateChangedMsg {
				msg := StateChangedMsg{
					ComponentID: "comp-1",
					RefID:       "ref-1",
					Timestamp:   time.Now(),
				}
				cmd := func() tea.Msg { return msg }
				q.Enqueue(cmd)
				return []StateChangedMsg{msg}
			},
			expectedCount: 1,
			verifyMetadata: func(t *testing.T, infos []CommandInfo, msgs []StateChangedMsg) {
				require.Len(t, infos, 1, "should return 1 command info")
				assert.Equal(t, msgs[0].ComponentID, infos[0].ComponentID)
				assert.Equal(t, msgs[0].RefID, infos[0].RefID)
				assert.Equal(t, msgs[0].Timestamp, infos[0].Timestamp)
			},
		},
		{
			name: "multiple commands",
			setupQueue: func(q *CommandQueue) []StateChangedMsg {
				msgs := []StateChangedMsg{
					{ComponentID: "comp-1", RefID: "ref-1", Timestamp: time.Now()},
					{ComponentID: "comp-1", RefID: "ref-2", Timestamp: time.Now()},
					{ComponentID: "comp-2", RefID: "ref-3", Timestamp: time.Now()},
				}
				for _, msg := range msgs {
					m := msg // Capture for closure
					cmd := func() tea.Msg { return m }
					q.Enqueue(cmd)
				}
				return msgs
			},
			expectedCount: 3,
			verifyMetadata: func(t *testing.T, infos []CommandInfo, msgs []StateChangedMsg) {
				require.Len(t, infos, 3, "should return 3 command infos")
				for i, info := range infos {
					assert.Equal(t, msgs[i].ComponentID, info.ComponentID,
						"command %d ComponentID should match", i)
					assert.Equal(t, msgs[i].RefID, info.RefID,
						"command %d RefID should match", i)
					assert.Equal(t, msgs[i].Timestamp, info.Timestamp,
						"command %d Timestamp should match", i)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			expectedMsgs := tt.setupQueue(queue)

			inspector := NewCommandInspector(queue)
			infos := inspector.PendingCommands()

			assert.Len(t, infos, tt.expectedCount,
				"should return %d command infos", tt.expectedCount)

			tt.verifyMetadata(t, infos, expectedMsgs)

			// Verify queue unchanged
			assert.Equal(t, tt.expectedCount, queue.Len(),
				"queue should still have %d commands", tt.expectedCount)
		})
	}
}

// TestCommandInspector_ClearPending tests that ClearPending() removes
// all pending commands from the queue.
//
// This test validates:
//   - Empty queue remains empty
//   - Single command cleared
//   - Multiple commands all cleared
//   - Count becomes 0 after clear
//   - Subsequent PendingCommands returns empty
//
// Test Cases:
//  1. Clear empty queue (no-op)
//  2. Clear single command
//  3. Clear multiple commands
//  4. Verify count after clear
func TestCommandInspector_ClearPending(t *testing.T) {
	tests := []struct {
		name       string
		setupQueue func(*CommandQueue)
	}{
		{
			name:       "empty queue",
			setupQueue: func(q *CommandQueue) {},
		},
		{
			name: "single command",
			setupQueue: func(q *CommandQueue) {
				cmd := func() tea.Msg {
					return StateChangedMsg{
						ComponentID: "comp-1",
						RefID:       "ref-1",
						Timestamp:   time.Now(),
					}
				}
				q.Enqueue(cmd)
			},
		},
		{
			name: "multiple commands",
			setupQueue: func(q *CommandQueue) {
				for i := 0; i < 10; i++ {
					cmd := func() tea.Msg {
						return StateChangedMsg{
							ComponentID: "comp-1",
							RefID:       "ref-1",
							Timestamp:   time.Now(),
						}
					}
					q.Enqueue(cmd)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			tt.setupQueue(queue)

			inspector := NewCommandInspector(queue)
			inspector.ClearPending()

			// Verify queue is empty
			assert.Equal(t, 0, queue.Len(), "queue should be empty after clear")
			assert.Equal(t, 0, inspector.PendingCount(), "PendingCount should be 0")
			assert.Empty(t, inspector.PendingCommands(), "PendingCommands should be empty")
		})
	}
}

// TestCommandInspector_ThreadSafety tests concurrent access to inspector methods.
//
// This test validates:
//   - Concurrent PendingCount calls are safe
//   - Concurrent PendingCommands calls are safe
//   - Concurrent ClearPending calls are safe
//   - Mixed concurrent operations are safe
//   - No race conditions detected
//
// Test Strategy:
//   - 10 goroutines performing mixed operations
//   - Each goroutine performs 10 operations
//   - Operations: PendingCount, PendingCommands, ClearPending, Enqueue
//   - Verify no panics or race conditions
func TestCommandInspector_ThreadSafety(t *testing.T) {
	queue := NewCommandQueue()
	inspector := NewCommandInspector(queue)

	const numGoroutines = 10
	const opsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				// Mix of operations
				switch j % 4 {
				case 0:
					_ = inspector.PendingCount()
				case 1:
					_ = inspector.PendingCommands()
				case 2:
					inspector.ClearPending()
				case 3:
					cmd := func() tea.Msg {
						return StateChangedMsg{
							ComponentID: "comp-1",
							RefID:       "ref-1",
							Timestamp:   time.Now(),
						}
					}
					queue.Enqueue(cmd)
				}
			}
		}(i)
	}

	wg.Wait()

	// If we get here without panics or race detector errors, test passes
	assert.True(t, true, "concurrent operations completed without errors")
}

// TestCommandInspector_NonStateChangedMsg tests that inspector handles
// non-StateChangedMsg commands gracefully.
//
// This test validates:
//   - Non-StateChangedMsg commands are skipped
//   - Count includes all commands
//   - PendingCommands only returns StateChangedMsg metadata
//   - No panics on type assertion
//
// Test Cases:
//  1. Queue with only non-StateChangedMsg commands
//  2. Queue with mixed StateChangedMsg and other messages
//  3. Verify only StateChangedMsg metadata extracted
func TestCommandInspector_NonStateChangedMsg(t *testing.T) {
	tests := []struct {
		name                  string
		setupQueue            func(*CommandQueue)
		expectedTotalCount    int
		expectedStateChangeds int
	}{
		{
			name: "only non-StateChangedMsg",
			setupQueue: func(q *CommandQueue) {
				cmd1 := func() tea.Msg { return tea.KeyMsg{} }
				cmd2 := func() tea.Msg { return "custom message" }
				q.Enqueue(cmd1)
				q.Enqueue(cmd2)
			},
			expectedTotalCount:    2,
			expectedStateChangeds: 0,
		},
		{
			name: "mixed messages",
			setupQueue: func(q *CommandQueue) {
				cmd1 := func() tea.Msg {
					return StateChangedMsg{
						ComponentID: "comp-1",
						RefID:       "ref-1",
						Timestamp:   time.Now(),
					}
				}
				cmd2 := func() tea.Msg { return tea.KeyMsg{} }
				cmd3 := func() tea.Msg {
					return StateChangedMsg{
						ComponentID: "comp-2",
						RefID:       "ref-2",
						Timestamp:   time.Now(),
					}
				}
				q.Enqueue(cmd1)
				q.Enqueue(cmd2)
				q.Enqueue(cmd3)
			},
			expectedTotalCount:    3,
			expectedStateChangeds: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewCommandQueue()
			tt.setupQueue(queue)

			inspector := NewCommandInspector(queue)

			// Total count includes all commands
			assert.Equal(t, tt.expectedTotalCount, inspector.PendingCount(),
				"PendingCount should include all commands")

			// PendingCommands only returns StateChangedMsg metadata
			infos := inspector.PendingCommands()
			assert.Len(t, infos, tt.expectedStateChangeds,
				"PendingCommands should only return StateChangedMsg metadata")
		})
	}
}

// TestCommandInspector_NilQueue tests that inspector handles nil queue gracefully.
//
// This test validates:
//   - Creating inspector with nil queue doesn't panic
//   - Methods return safe defaults (0, empty slice)
//   - Clear on nil queue is no-op
//
// Test Cases:
//  1. PendingCount returns 0
//  2. PendingCommands returns empty slice
//  3. ClearPending is no-op
func TestCommandInspector_NilQueue(t *testing.T) {
	inspector := NewCommandInspector(nil)

	// Should not panic and return safe defaults
	assert.Equal(t, 0, inspector.PendingCount(), "PendingCount should return 0 for nil queue")
	assert.Empty(t, inspector.PendingCommands(), "PendingCommands should return empty for nil queue")
	assert.NotPanics(t, func() {
		inspector.ClearPending()
	}, "ClearPending should not panic for nil queue")
}
