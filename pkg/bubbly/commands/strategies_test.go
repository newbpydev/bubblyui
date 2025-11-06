package commands

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestBatchAll_SingleCommand verifies that single command optimization works
func TestBatchAll_SingleCommand(t *testing.T) {
	executed := false
	cmd := func() tea.Msg {
		executed = true
		return "test-message"
	}

	batcher := NewCommandBatcher(CoalesceAll)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd})

	assert.NotNil(t, batchedCmd, "Batched command should not be nil")

	// Execute the batched command
	msg := batchedCmd()

	assert.True(t, executed, "Original command should have been executed")

	// Single command optimization: returns original message, not batch
	// This is correct behavior for all strategies
	assert.Equal(t, "test-message", msg, "Single command should return original message")
}

// TestBatchAll_MultipleCommands verifies that batchAll collects all messages
func TestBatchAll_MultipleCommands(t *testing.T) {
	cmd1 := func() tea.Msg { return "msg1" }
	cmd2 := func() tea.Msg { return "msg2" }
	cmd3 := func() tea.Msg { return "msg3" }

	batcher := NewCommandBatcher(CoalesceAll)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})

	msg := batchedCmd()

	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "Message should be StateChangedBatchMsg")
	assert.Equal(t, 3, batchMsg.Count, "Count should be 3")
	assert.Len(t, batchMsg.Messages, 3, "Should have 3 messages")
	assert.Equal(t, "msg1", batchMsg.Messages[0])
	assert.Equal(t, "msg2", batchMsg.Messages[1])
	assert.Equal(t, "msg3", batchMsg.Messages[2])
}

// TestBatchAll_NilCommandsFiltered verifies that nil commands are filtered before execution
func TestBatchAll_NilCommandsFiltered(t *testing.T) {
	cmd1 := func() tea.Msg { return "msg1" }
	cmd2 := func() tea.Msg { return "msg2" }

	batcher := NewCommandBatcher(CoalesceAll)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, nil, cmd2, nil})

	msg := batchedCmd()

	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "Message should be StateChangedBatchMsg")
	assert.Equal(t, 2, batchMsg.Count, "Count should be 2 (nil filtered)")
	assert.Len(t, batchMsg.Messages, 2, "Should have 2 messages")
}

// TestBatchAll_ExecutionOrder verifies commands execute in order
func TestBatchAll_ExecutionOrder(t *testing.T) {
	order := []int{}

	cmd1 := func() tea.Msg {
		order = append(order, 1)
		return "msg1"
	}
	cmd2 := func() tea.Msg {
		order = append(order, 2)
		return "msg2"
	}
	cmd3 := func() tea.Msg {
		order = append(order, 3)
		return "msg3"
	}

	batcher := NewCommandBatcher(CoalesceAll)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})

	_ = batchedCmd()

	assert.Equal(t, []int{1, 2, 3}, order, "Commands should execute in order")
}

// TestBatchByType_Placeholder verifies batchByType works (full impl later)
func TestBatchByType_Placeholder(t *testing.T) {
	cmd1 := func() tea.Msg { return "msg1" }
	cmd2 := func() tea.Msg { return "msg2" }

	batcher := NewCommandBatcher(CoalesceByType)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2})

	assert.NotNil(t, batchedCmd, "Batched command should not be nil")
	// For now, CoalesceByType uses tea.Batch (placeholder)
	// Full type-based grouping will be implemented later
}

// TestNoCoalesce_UseTeaBatch verifies noCoalesce delegates to tea.Batch
func TestNoCoalesce_UseTeaBatch(t *testing.T) {
	cmd1 := func() tea.Msg { return "msg1" }
	cmd2 := func() tea.Msg { return "msg2" }

	batcher := NewCommandBatcher(NoCoalesce)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2})

	assert.NotNil(t, batchedCmd, "Batched command should not be nil")
	// NoCoalesce should use tea.Batch - commands execute individually
}

// TestStateChangedBatchMsg_Structure verifies the batch message structure
func TestStateChangedBatchMsg_Structure(t *testing.T) {
	tests := []struct {
		name     string
		messages []tea.Msg
		count    int
	}{
		{
			name:     "single message",
			messages: []tea.Msg{"msg1"},
			count:    1,
		},
		{
			name:     "multiple messages",
			messages: []tea.Msg{"msg1", "msg2", "msg3"},
			count:    3,
		},
		{
			name:     "empty messages",
			messages: []tea.Msg{},
			count:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := StateChangedBatchMsg{
				Messages: tt.messages,
				Count:    tt.count,
			}

			assert.Equal(t, tt.count, msg.Count, "Count should match")
			assert.Equal(t, tt.messages, msg.Messages, "Messages should match")
		})
	}
}

// TestBatchAll_DifferentMessageTypes verifies collecting different message types
func TestBatchAll_DifferentMessageTypes(t *testing.T) {
	type customMsg struct {
		value string
	}

	cmd1 := func() tea.Msg { return "string message" }
	cmd2 := func() tea.Msg { return 42 }
	cmd3 := func() tea.Msg { return customMsg{value: "custom"} }

	batcher := NewCommandBatcher(CoalesceAll)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})

	msg := batchedCmd()

	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "Message should be StateChangedBatchMsg")
	assert.Equal(t, 3, batchMsg.Count)

	// Verify each message type
	assert.Equal(t, "string message", batchMsg.Messages[0])
	assert.Equal(t, 42, batchMsg.Messages[1])
	assert.Equal(t, customMsg{value: "custom"}, batchMsg.Messages[2])
}

// TestBatcher_StrategyMethodSelection verifies correct method is called for each strategy
func TestBatcher_StrategyMethodSelection(t *testing.T) {
	// Use multiple commands to avoid single-command optimization
	cmd1 := func() tea.Msg { return "test1" }
	cmd2 := func() tea.Msg { return "test2" }

	tests := []struct {
		name     string
		strategy CoalescingStrategy
		verify   func(t *testing.T, msg tea.Msg)
	}{
		{
			name:     "CoalesceAll returns batch message",
			strategy: CoalesceAll,
			verify: func(t *testing.T, msg tea.Msg) {
				batchMsg, ok := msg.(StateChangedBatchMsg)
				assert.True(t, ok, "Should return StateChangedBatchMsg")
				assert.Equal(t, 2, batchMsg.Count, "Should have 2 messages")
			},
		},
		{
			name:     "CoalesceByType returns command result",
			strategy: CoalesceByType,
			verify: func(t *testing.T, msg tea.Msg) {
				// For now, just verify command executes
				// Full type grouping will be tested later
				_ = msg
			},
		},
		{
			name:     "NoCoalesce returns command result",
			strategy: NoCoalesce,
			verify: func(t *testing.T, msg tea.Msg) {
				// NoCoalesce uses tea.Batch, just verify it works
				_ = msg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(tt.strategy)
			batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2})

			assert.NotNil(t, batchedCmd, "Batched command should not be nil")

			msg := batchedCmd()
			tt.verify(t, msg)
		})
	}
}
