package testutil

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewMockCommandGenerator tests mock creation with configured command
func TestNewMockCommandGenerator(t *testing.T) {
	tests := []struct {
		name      string
		returnCmd tea.Cmd
	}{
		{
			name:      "with nil command",
			returnCmd: nil,
		},
		{
			name: "with valid command",
			returnCmd: func() tea.Msg {
				return "test message"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockCommandGenerator(tt.returnCmd)

			assert.NotNil(t, mock)
			assert.Equal(t, 0, mock.generateCalled)
			assert.Empty(t, mock.capturedArgs)
		})
	}
}

// TestMockCommandGenerator_Generate tests command generation
func TestMockCommandGenerator_Generate(t *testing.T) {
	tests := []struct {
		name        string
		returnCmd   tea.Cmd
		componentID string
		refID       string
		oldValue    interface{}
		newValue    interface{}
	}{
		{
			name: "returns configured command",
			returnCmd: func() tea.Msg {
				return "test message"
			},
			componentID: "comp-1",
			refID:       "count",
			oldValue:    0,
			newValue:    1,
		},
		{
			name:        "returns nil when configured with nil",
			returnCmd:   nil,
			componentID: "comp-2",
			refID:       "name",
			oldValue:    "old",
			newValue:    "new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockCommandGenerator(tt.returnCmd)

			// Generate command
			cmd := mock.Generate(tt.componentID, tt.refID, tt.oldValue, tt.newValue)

			// Verify command matches configured return
			if tt.returnCmd == nil {
				assert.Nil(t, cmd)
			} else {
				assert.NotNil(t, cmd)
				// Execute and verify message
				msg := cmd()
				assert.Equal(t, "test message", msg)
			}

			// Verify call count incremented
			assert.Equal(t, 1, mock.generateCalled)
		})
	}
}

// TestMockCommandGenerator_CapturesArguments tests argument capture
func TestMockCommandGenerator_CapturesArguments(t *testing.T) {
	mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })

	// First call
	mock.Generate("comp-1", "ref-1", 0, 1)

	// Second call
	mock.Generate("comp-2", "ref-2", "old", "new")

	// Third call
	mock.Generate("comp-3", "ref-3", false, true)

	// Verify all arguments captured
	assert.Len(t, mock.capturedArgs, 3)

	// Verify first call
	assert.Equal(t, "comp-1", mock.capturedArgs[0].ComponentID)
	assert.Equal(t, "ref-1", mock.capturedArgs[0].RefID)
	assert.Equal(t, 0, mock.capturedArgs[0].OldValue)
	assert.Equal(t, 1, mock.capturedArgs[0].NewValue)

	// Verify second call
	assert.Equal(t, "comp-2", mock.capturedArgs[1].ComponentID)
	assert.Equal(t, "ref-2", mock.capturedArgs[1].RefID)
	assert.Equal(t, "old", mock.capturedArgs[1].OldValue)
	assert.Equal(t, "new", mock.capturedArgs[1].NewValue)

	// Verify third call
	assert.Equal(t, "comp-3", mock.capturedArgs[2].ComponentID)
	assert.Equal(t, "ref-3", mock.capturedArgs[2].RefID)
	assert.Equal(t, false, mock.capturedArgs[2].OldValue)
	assert.Equal(t, true, mock.capturedArgs[2].NewValue)
}

// TestMockCommandGenerator_AssertCalled tests call count assertion
func TestMockCommandGenerator_AssertCalled(t *testing.T) {
	tests := []struct {
		name       string
		callCount  int
		assertWith int
		shouldFail bool
	}{
		{
			name:       "passes when count matches",
			callCount:  3,
			assertWith: 3,
			shouldFail: false,
		},
		{
			name:       "fails when count doesn't match",
			callCount:  2,
			assertWith: 5,
			shouldFail: true,
		},
		{
			name:       "passes with zero calls",
			callCount:  0,
			assertWith: 0,
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })

			// Make calls
			for i := 0; i < tt.callCount; i++ {
				mock.Generate("comp", "ref", i, i+1)
			}

			// Use mock testing.T to capture failures
			mockT := &mockTestingT{}
			mock.AssertCalled(mockT, tt.assertWith)

			if tt.shouldFail {
				assert.True(t, mockT.failed, "expected assertion to fail")
			} else {
				assert.False(t, mockT.failed, "expected assertion to pass")
			}
		})
	}
}

// TestMockCommandGenerator_GetCapturedArgs tests captured args retrieval
func TestMockCommandGenerator_GetCapturedArgs(t *testing.T) {
	mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })

	// No calls yet
	assert.Empty(t, mock.GetCapturedArgs())

	// Make some calls
	mock.Generate("comp-1", "ref-1", 0, 1)
	mock.Generate("comp-2", "ref-2", "a", "b")

	// Verify captured args
	args := mock.GetCapturedArgs()
	assert.Len(t, args, 2)
	assert.Equal(t, "comp-1", args[0].ComponentID)
	assert.Equal(t, "comp-2", args[1].ComponentID)
}

// TestMockCommandGenerator_Clear tests clearing state
func TestMockCommandGenerator_Clear(t *testing.T) {
	mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })

	// Make some calls
	mock.Generate("comp-1", "ref-1", 0, 1)
	mock.Generate("comp-2", "ref-2", "a", "b")
	assert.Equal(t, 2, mock.generateCalled)
	assert.Len(t, mock.capturedArgs, 2)

	// Clear state
	mock.Clear()

	// Verify state cleared
	assert.Equal(t, 0, mock.generateCalled)
	assert.Empty(t, mock.capturedArgs)

	// Verify can be used again
	mock.Generate("comp-3", "ref-3", 1, 2)
	assert.Equal(t, 1, mock.generateCalled)
	assert.Len(t, mock.capturedArgs, 1)
}

// TestMockCommandGenerator_ThreadSafe tests concurrent access
func TestMockCommandGenerator_ThreadSafe(t *testing.T) {
	mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })

	// Run concurrent Generate calls
	const goroutines = 10
	const callsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				mock.Generate("comp", "ref", j, j+1)
			}
		}(i)
	}

	wg.Wait()

	// Verify all calls recorded
	expectedCalls := goroutines * callsPerGoroutine
	assert.Equal(t, expectedCalls, mock.generateCalled)
	assert.Len(t, mock.capturedArgs, expectedCalls)
}

// TestMockCommandGenerator_IntegrationWithCommandGenerator tests interface compliance
func TestMockCommandGenerator_IntegrationWithCommandGenerator(t *testing.T) {
	// Verify MockCommandGenerator implements CommandGenerator interface
	var _ bubbly.CommandGenerator = (*MockCommandGenerator)(nil)

	mock := NewMockCommandGenerator(func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "test-comp",
			RefID:       "test-ref",
		}
	})

	// Use as CommandGenerator
	var gen bubbly.CommandGenerator = mock
	cmd := gen.Generate("comp-1", "ref-1", 0, 1)

	// Verify command works
	assert.NotNil(t, cmd)
	msg := cmd()
	stateMsg, ok := msg.(bubbly.StateChangedMsg)
	assert.True(t, ok)
	assert.Equal(t, "test-comp", stateMsg.ComponentID)
	assert.Equal(t, "test-ref", stateMsg.RefID)

	// Verify mock tracked the call
	assert.Equal(t, 1, mock.generateCalled)
}

// TestMockCommandGenerator_NilReturnCmd tests behavior with nil return command
func TestMockCommandGenerator_NilReturnCmd(t *testing.T) {
	mock := NewMockCommandGenerator(nil)

	// Generate should return nil
	cmd := mock.Generate("comp", "ref", 0, 1)
	assert.Nil(t, cmd)

	// But should still track the call
	assert.Equal(t, 1, mock.generateCalled)
	assert.Len(t, mock.capturedArgs, 1)
}

// TestMockCommandGenerator_IdempotentOperations tests idempotent methods
func TestMockCommandGenerator_IdempotentOperations(t *testing.T) {
	mock := NewMockCommandGenerator(func() tea.Msg { return "msg" })

	// Clear on empty mock
	mock.Clear()
	assert.Equal(t, 0, mock.generateCalled)
	assert.Empty(t, mock.capturedArgs)

	// Multiple clears
	mock.Generate("comp", "ref", 0, 1)
	mock.Clear()
	mock.Clear()
	mock.Clear()
	assert.Equal(t, 0, mock.generateCalled)
	assert.Empty(t, mock.capturedArgs)
}
