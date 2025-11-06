package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCommandRef_Creation tests that CommandRef wraps Ref correctly
func TestCommandRef_Creation(t *testing.T) {
	tests := []struct {
		name         string
		initialValue interface{}
		enabled      bool
	}{
		{
			name:         "integer ref enabled",
			initialValue: 42,
			enabled:      true,
		},
		{
			name:         "string ref enabled",
			initialValue: "hello",
			enabled:      true,
		},
		{
			name:         "boolean ref disabled",
			initialValue: true,
			enabled:      false,
		},
		{
			name:         "nil value enabled",
			initialValue: nil,
			enabled:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.initialValue)
			queue := NewCommandQueue()
			gen := &DefaultCommandGenerator{}

			cmdRef := &CommandRef[interface{}]{
				Ref:         ref,
				componentID: "test-component",
				refID:       "test-ref",
				commandGen:  gen,
				queue:       queue,
				enabled:     tt.enabled,
			}

			// Verify wrapping
			assert.NotNil(t, cmdRef.Ref)
			assert.Equal(t, tt.initialValue, cmdRef.Get())
			assert.Equal(t, "test-component", cmdRef.componentID)
			assert.Equal(t, tt.enabled, cmdRef.enabled)
			assert.NotNil(t, cmdRef.commandGen)
			assert.NotNil(t, cmdRef.queue)
		})
	}
}

// TestCommandRef_SetEnabled tests Set() generates commands when enabled
func TestCommandRef_SetEnabled(t *testing.T) {
	tests := []struct {
		name     string
		oldValue interface{}
		newValue interface{}
	}{
		{
			name:     "integer change",
			oldValue: 0,
			newValue: 42,
		},
		{
			name:     "string change",
			oldValue: "old",
			newValue: "new",
		},
		{
			name:     "boolean change",
			oldValue: false,
			newValue: true,
		},
		{
			name:     "nil to value",
			oldValue: nil,
			newValue: "value",
		},
		{
			name:     "value to nil",
			oldValue: "value",
			newValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef(tt.oldValue)
			queue := NewCommandQueue()
			gen := &DefaultCommandGenerator{}

			cmdRef := &CommandRef[interface{}]{
				Ref:         ref,
				componentID: "test-component",
				refID:       "test-ref",
				commandGen:  gen,
				queue:       queue,
				enabled:     true,
			}

			// Set new value
			cmdRef.Set(tt.newValue)

			// Verify value updated
			assert.Equal(t, tt.newValue, cmdRef.Get())

			// Verify command was generated and enqueued
			assert.Equal(t, 1, queue.Len())

			// Drain and execute command
			cmds := queue.DrainAll()
			require.Len(t, cmds, 1)

			msg := cmds[0]()
			stateMsg, ok := msg.(StateChangedMsg)
			require.True(t, ok, "Expected StateChangedMsg")

			// Verify message contents
			assert.Equal(t, "test-component", stateMsg.ComponentID)
			assert.Equal(t, tt.oldValue, stateMsg.OldValue)
			assert.Equal(t, tt.newValue, stateMsg.NewValue)
			assert.False(t, stateMsg.Timestamp.IsZero())
		})
	}
}

// TestCommandRef_SetDisabled tests Set() does NOT generate commands when disabled
func TestCommandRef_SetDisabled(t *testing.T) {
	ref := bubbly.NewRef(0)
	queue := NewCommandQueue()
	gen := &DefaultCommandGenerator{}

	cmdRef := &CommandRef[int]{
		Ref:         ref,
		componentID: "test-component",
		refID:       "test-ref",
		commandGen:  gen,
		queue:       queue,
		enabled:     false, // Disabled
	}

	// Set new value
	cmdRef.Set(42)

	// Verify value updated
	assert.Equal(t, 42, cmdRef.Get())

	// Verify NO command was generated
	assert.Equal(t, 0, queue.Len())
}

// TestCommandRef_MultipleSetsCalls tests multiple Set() calls enqueue multiple commands
func TestCommandRef_MultipleSetsCalls(t *testing.T) {
	ref := bubbly.NewRef(0)
	queue := NewCommandQueue()
	gen := &DefaultCommandGenerator{}

	cmdRef := &CommandRef[int]{
		Ref:         ref,
		componentID: "test-component",
		refID:       "test-ref",
		commandGen:  gen,
		queue:       queue,
		enabled:     true,
	}

	// Multiple Set() calls
	cmdRef.Set(1)
	cmdRef.Set(2)
	cmdRef.Set(3)

	// Verify final value
	assert.Equal(t, 3, cmdRef.Get())

	// Verify 3 commands enqueued
	assert.Equal(t, 3, queue.Len())

	// Drain and verify commands
	cmds := queue.DrainAll()
	require.Len(t, cmds, 3)

	// Execute and verify each command
	expectedValues := []struct{ old, new int }{
		{0, 1},
		{1, 2},
		{2, 3},
	}

	for i, cmd := range cmds {
		msg := cmd()
		stateMsg, ok := msg.(StateChangedMsg)
		require.True(t, ok)

		assert.Equal(t, expectedValues[i].old, stateMsg.OldValue)
		assert.Equal(t, expectedValues[i].new, stateMsg.NewValue)
	}
}

// TestCommandRef_ThreadSafety tests concurrent Set() operations
func TestCommandRef_ThreadSafety(t *testing.T) {
	ref := bubbly.NewRef(0)
	queue := NewCommandQueue()
	gen := &DefaultCommandGenerator{}

	cmdRef := &CommandRef[int]{
		Ref:         ref,
		componentID: "test-component",
		refID:       "test-ref",
		commandGen:  gen,
		queue:       queue,
		enabled:     true,
	}

	// Concurrent Set() calls
	const goroutines = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(val int) {
			cmdRef.Set(val)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify commands were enqueued (exact count may vary due to race)
	// But should have at least some commands
	assert.Greater(t, queue.Len(), 0)
	assert.LessOrEqual(t, queue.Len(), goroutines)

	// Verify final value is one of the set values
	finalValue := cmdRef.Get()
	assert.GreaterOrEqual(t, finalValue, 0)
	assert.Less(t, finalValue, goroutines)
}

// TestCommandRef_ValueUpdatesSynchronous tests that value updates are synchronous
func TestCommandRef_ValueUpdatesSynchronous(t *testing.T) {
	ref := bubbly.NewRef(0)
	queue := NewCommandQueue()
	gen := &DefaultCommandGenerator{}

	cmdRef := &CommandRef[int]{
		Ref:         ref,
		componentID: "test-component",
		refID:       "test-ref",
		commandGen:  gen,
		queue:       queue,
		enabled:     true,
	}

	// Set value
	cmdRef.Set(42)

	// Value should be immediately visible (synchronous)
	assert.Equal(t, 42, cmdRef.Get())

	// Command is queued but not executed yet (asynchronous)
	assert.Equal(t, 1, queue.Len())
}

// TestCommandRef_DisabledModeBypassesQueue tests disabled mode doesn't touch queue
func TestCommandRef_DisabledModeBypassesQueue(t *testing.T) {
	ref := bubbly.NewRef(0)
	queue := NewCommandQueue()
	gen := &DefaultCommandGenerator{}

	cmdRef := &CommandRef[int]{
		Ref:         ref,
		componentID: "test-component",
		refID:       "test-ref",
		commandGen:  gen,
		queue:       queue,
		enabled:     false,
	}

	// Multiple Set() calls
	for i := 1; i <= 10; i++ {
		cmdRef.Set(i)
	}

	// Verify final value
	assert.Equal(t, 10, cmdRef.Get())

	// Verify NO commands enqueued
	assert.Equal(t, 0, queue.Len())
}
