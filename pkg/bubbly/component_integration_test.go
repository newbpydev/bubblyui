package bubbly

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdate_StateChangedMsg_HandlesCorrectly tests that Update() properly handles StateChangedMsg
func TestUpdate_StateChangedMsg_HandlesCorrectly(t *testing.T) {
	tests := []struct {
		name              string
		componentID       string
		msgComponentID    string
		hasLifecycle      bool
		expectHookExecute bool
	}{
		{
			name:              "matching component ID executes hooks",
			componentID:       "test-comp-1",
			msgComponentID:    "test-comp-1",
			hasLifecycle:      true,
			expectHookExecute: true,
		},
		{
			name:              "non-matching component ID skips hooks",
			componentID:       "test-comp-1",
			msgComponentID:    "test-comp-2",
			hasLifecycle:      true,
			expectHookExecute: false,
		},
		{
			name:              "no lifecycle manager skips hooks",
			componentID:       "test-comp-1",
			msgComponentID:    "test-comp-1",
			hasLifecycle:      false,
			expectHookExecute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component
			comp := &componentImpl{
				id:           tt.componentID,
				commandQueue: NewCommandQueue(),
				commandGen:   &defaultCommandGenerator{},
			}

			// Add lifecycle if needed
			hookExecuted := false
			if tt.hasLifecycle {
				comp.lifecycle = newLifecycleManager(comp)
				// Mark as mounted so hooks can execute
				comp.lifecycle.executeMounted()
				// Register onUpdated hook using Context API
				ctx := &Context{component: comp}
				ctx.OnUpdated(func() {
					hookExecuted = true
				})
			}

			// Create StateChangedMsg
			msg := StateChangedMsg{
				ComponentID: tt.msgComponentID,
				RefID:       "ref-1",
				OldValue:    0,
				NewValue:    1,
				Timestamp:   time.Now(),
			}

			// Call Update
			updated, cmd := comp.Update(msg)

			// Verify component returned
			assert.NotNil(t, updated)
			assert.Equal(t, comp, updated)

			// Verify hook execution
			assert.Equal(t, tt.expectHookExecute, hookExecuted)

			// Verify command returned (should be nil or valid)
			_ = cmd
		})
	}
}

// TestUpdate_CommandQueueDraining tests that Update() drains and returns queued commands
func TestUpdate_CommandQueueDraining(t *testing.T) {
	tests := []struct {
		name        string
		queuedCmds  int
		expectCmds  bool
		hasChildren bool
		childCmds   int
	}{
		{
			name:        "no queued commands returns nil",
			queuedCmds:  0,
			expectCmds:  false,
			hasChildren: false,
		},
		{
			name:        "single queued command returned",
			queuedCmds:  1,
			expectCmds:  true,
			hasChildren: false,
		},
		{
			name:        "multiple queued commands batched",
			queuedCmds:  3,
			expectCmds:  true,
			hasChildren: false,
		},
		{
			name:        "queued commands batch with child commands",
			queuedCmds:  2,
			expectCmds:  true,
			hasChildren: true,
			childCmds:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component
			comp := &componentImpl{
				id:           "test-comp",
				commandQueue: NewCommandQueue(),
				commandGen:   &defaultCommandGenerator{},
			}

			// Enqueue commands
			for i := 0; i < tt.queuedCmds; i++ {
				cmd := func() tea.Msg {
					return StateChangedMsg{
						ComponentID: "test-comp",
						RefID:       "ref-1",
						Timestamp:   time.Now(),
					}
				}
				comp.commandQueue.Enqueue(cmd)
			}

			// Add children if needed
			if tt.hasChildren {
				child := &componentImpl{
					id:           "child-comp",
					commandQueue: NewCommandQueue(),
					commandGen:   &defaultCommandGenerator{},
				}
				// Enqueue child commands
				for i := 0; i < tt.childCmds; i++ {
					cmd := func() tea.Msg {
						return StateChangedMsg{
							ComponentID: "child-comp",
							RefID:       "ref-2",
							Timestamp:   time.Now(),
						}
					}
					child.commandQueue.Enqueue(cmd)
				}
				comp.children = []Component{child}
			}

			// Verify queue has commands before Update
			assert.Equal(t, tt.queuedCmds, comp.commandQueue.Len())

			// Call Update with a dummy message
			updated, cmd := comp.Update(tea.KeyMsg{})

			// Verify component returned
			assert.NotNil(t, updated)

			// Verify queue is drained
			assert.Equal(t, 0, comp.commandQueue.Len())

			// Verify command returned
			if tt.expectCmds || tt.hasChildren {
				assert.NotNil(t, cmd, "expected command to be returned")
			}

			// Execute command to verify it works
			if cmd != nil {
				msg := cmd()
				assert.NotNil(t, msg, "command should return a message")
			}
		})
	}
}

// TestUpdate_IntegrationFlow tests the complete automatic update flow
func TestUpdate_IntegrationFlow(t *testing.T) {
	t.Run("complete automatic update cycle", func(t *testing.T) {
		// Track hook execution
		hookExecuted := false

		// Create component with lifecycle
		comp := &componentImpl{
			id:           "counter",
			commandQueue: NewCommandQueue(),
			commandGen:   &defaultCommandGenerator{},
		}
		comp.lifecycle = newLifecycleManager(comp)
		// Mark as mounted so hooks can execute
		comp.lifecycle.executeMounted()

		// Add onUpdated hook using Context API
		ctx := &Context{component: comp}
		ctx.OnUpdated(func() {
			hookExecuted = true
		})

		// Simulate state change by enqueueing command
		cmd := comp.commandGen.Generate("counter", "count-ref", 0, 1)
		comp.commandQueue.Enqueue(cmd)

		// Verify command is queued
		assert.Equal(t, 1, comp.commandQueue.Len())

		// Call Update (simulating Bubbletea runtime)
		updated, returnedCmd := comp.Update(tea.KeyMsg{})

		// Verify queue is drained
		assert.Equal(t, 0, comp.commandQueue.Len())

		// Verify command returned
		require.NotNil(t, returnedCmd, "Update should return batched command")

		// Execute the returned command (simulating Bubbletea runtime)
		msg := returnedCmd()
		require.NotNil(t, msg, "command should return message")

		// Verify it's a StateChangedMsg
		stateMsg, ok := msg.(StateChangedMsg)
		require.True(t, ok, "message should be StateChangedMsg")
		assert.Equal(t, "counter", stateMsg.ComponentID)
		assert.Equal(t, "count-ref", stateMsg.RefID)

		// Call Update again with the StateChangedMsg (simulating Bubbletea runtime)
		updated, _ = updated.(Component).Update(stateMsg)

		// Verify hook was executed
		assert.True(t, hookExecuted, "onUpdated hook should have been executed")

		// Verify component returned
		assert.NotNil(t, updated)
	})

	t.Run("multiple state changes batch correctly", func(t *testing.T) {
		// Create component
		comp := &componentImpl{
			id:           "multi-counter",
			commandQueue: NewCommandQueue(),
			commandGen:   &defaultCommandGenerator{},
		}

		// Enqueue multiple commands (simulating multiple Ref.Set() calls)
		cmd1 := comp.commandGen.Generate("multi-counter", "ref-1", 0, 1)
		cmd2 := comp.commandGen.Generate("multi-counter", "ref-2", "a", "b")
		cmd3 := comp.commandGen.Generate("multi-counter", "ref-3", false, true)

		comp.commandQueue.Enqueue(cmd1)
		comp.commandQueue.Enqueue(cmd2)
		comp.commandQueue.Enqueue(cmd3)

		// Verify all commands queued
		assert.Equal(t, 3, comp.commandQueue.Len())

		// Call Update
		_, returnedCmd := comp.Update(tea.KeyMsg{})

		// Verify queue drained
		assert.Equal(t, 0, comp.commandQueue.Len())

		// Verify command returned
		require.NotNil(t, returnedCmd, "Update should return batched command")

		// Execute command - should execute all batched commands
		msg := returnedCmd()
		assert.NotNil(t, msg, "batched command should return message")
	})
}

// TestUpdate_WithChildren tests Update() with child components
func TestUpdate_WithChildren(t *testing.T) {
	t.Run("parent and child commands batch together", func(t *testing.T) {
		// Create parent component
		parent := &componentImpl{
			id:           "parent",
			commandQueue: NewCommandQueue(),
			commandGen:   &defaultCommandGenerator{},
		}

		// Create child component
		child := &componentImpl{
			id:           "child",
			commandQueue: NewCommandQueue(),
			commandGen:   &defaultCommandGenerator{},
		}

		// Add child to parent
		parent.children = []Component{child}

		// Enqueue commands in both
		parentCmd := parent.commandGen.Generate("parent", "ref-1", 0, 1)
		childCmd := child.commandGen.Generate("child", "ref-2", "a", "b")

		parent.commandQueue.Enqueue(parentCmd)
		child.commandQueue.Enqueue(childCmd)

		// Verify both have commands
		assert.Equal(t, 1, parent.commandQueue.Len())
		assert.Equal(t, 1, child.commandQueue.Len())

		// Call Update on parent
		_, returnedCmd := parent.Update(tea.KeyMsg{})

		// Verify both queues drained
		assert.Equal(t, 0, parent.commandQueue.Len())
		assert.Equal(t, 0, child.commandQueue.Len())

		// Verify command returned
		require.NotNil(t, returnedCmd, "Update should return batched commands")

		// Execute command
		msg := returnedCmd()
		assert.NotNil(t, msg, "batched command should return message")
	})
}

// TestUpdate_BackwardCompatibility tests that existing behavior still works
func TestUpdate_BackwardCompatibility(t *testing.T) {
	t.Run("component without command queue works", func(t *testing.T) {
		// Create component without initializing command queue (simulating old code)
		comp := &componentImpl{
			id: "legacy-comp",
			// commandQueue intentionally nil
		}

		// This should not panic
		updated, cmd := comp.Update(tea.KeyMsg{})

		// Verify component returned
		assert.NotNil(t, updated)

		// Command may be nil (no commands to return)
		_ = cmd
	})

	t.Run("component with children but no command queue works", func(t *testing.T) {
		// Create parent without command queue
		parent := &componentImpl{
			id: "legacy-parent",
			// commandQueue intentionally nil
		}

		// Create child with command queue
		child := &componentImpl{
			id:           "modern-child",
			commandQueue: NewCommandQueue(),
			commandGen:   &defaultCommandGenerator{},
		}

		parent.children = []Component{child}

		// This should not panic
		updated, cmd := parent.Update(tea.KeyMsg{})

		// Verify parent returned
		assert.NotNil(t, updated)

		// Command may be nil or valid
		_ = cmd
	})
}
