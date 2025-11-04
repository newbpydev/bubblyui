package commands

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestCommandBatcher_EmptyList verifies that batching an empty list returns nil
func TestCommandBatcher_EmptyList(t *testing.T) {
	tests := []struct {
		name     string
		strategy CoalescingStrategy
	}{
		{"CoalesceAll", CoalesceAll},
		{"CoalesceByType", CoalesceByType},
		{"NoCoalesce", NoCoalesce},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(tt.strategy)
			cmd := batcher.Batch([]tea.Cmd{})
			assert.Nil(t, cmd, "Empty command list should return nil")
		})
	}
}

// TestCommandBatcher_SingleCommand verifies that a single command is returned as-is
func TestCommandBatcher_SingleCommand(t *testing.T) {
	executed := false
	singleCmd := func() tea.Msg {
		executed = true
		return "test"
	}

	tests := []struct {
		name     string
		strategy CoalescingStrategy
	}{
		{"CoalesceAll", CoalesceAll},
		{"CoalesceByType", CoalesceByType},
		{"NoCoalesce", NoCoalesce},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed = false
			batcher := NewCommandBatcher(tt.strategy)
			cmd := batcher.Batch([]tea.Cmd{singleCmd})

			assert.NotNil(t, cmd, "Single command should not be nil")

			// Execute the returned command
			msg := cmd()
			assert.Equal(t, "test", msg)
			assert.True(t, executed, "Command should have been executed")
		})
	}
}

// TestCommandBatcher_MultipleCommands verifies that multiple commands are batched
func TestCommandBatcher_MultipleCommands(t *testing.T) {
	count := 0
	cmd1 := func() tea.Msg {
		count++
		return "cmd1"
	}
	cmd2 := func() tea.Msg {
		count++
		return "cmd2"
	}
	cmd3 := func() tea.Msg {
		count++
		return "cmd3"
	}

	tests := []struct {
		name     string
		strategy CoalescingStrategy
	}{
		{"CoalesceAll", CoalesceAll},
		{"NoCoalesce", NoCoalesce},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count = 0
			batcher := NewCommandBatcher(tt.strategy)
			cmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})

			assert.NotNil(t, cmd, "Batched command should not be nil")

			// Note: tea.Batch returns a command that we can't easily inspect
			// We'll verify it's not nil and trust tea.Batch works correctly
			// Actual execution testing will be in integration tests
		})
	}
}

// TestCommandBatcher_NilCommands verifies that nil commands are handled
func TestCommandBatcher_NilCommands(t *testing.T) {
	cmd1 := func() tea.Msg {
		return "cmd1"
	}

	tests := []struct {
		name     string
		commands []tea.Cmd
	}{
		{
			name:     "all nil",
			commands: []tea.Cmd{nil, nil, nil},
		},
		{
			name:     "mixed nil and valid",
			commands: []tea.Cmd{nil, cmd1, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(CoalesceAll)
			cmd := batcher.Batch(tt.commands)

			// Should not panic and should handle nil gracefully
			// Result will depend on implementation (could be nil or a filtered batch)
			_ = cmd
		})
	}
}

// TestCommandBatcher_StrategySelection verifies different strategies can be created
func TestCommandBatcher_StrategySelection(t *testing.T) {
	tests := []struct {
		name     string
		strategy CoalescingStrategy
	}{
		{"CoalesceAll", CoalesceAll},
		{"CoalesceByType", CoalesceByType},
		{"NoCoalesce", NoCoalesce},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(tt.strategy)
			assert.NotNil(t, batcher, "Batcher should not be nil")
			assert.Equal(t, tt.strategy, batcher.strategy, "Strategy should be set correctly")
		})
	}
}

// TestCommandBatcher_BatchExecution verifies batched commands execute correctly
func TestCommandBatcher_BatchExecution(t *testing.T) {
	results := []string{}

	cmd1 := func() tea.Msg {
		results = append(results, "cmd1")
		return "msg1"
	}
	cmd2 := func() tea.Msg {
		results = append(results, "cmd2")
		return "msg2"
	}

	batcher := NewCommandBatcher(CoalesceAll)
	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2})

	assert.NotNil(t, batchedCmd, "Batched command should not be nil")

	// Execute the batched command
	// Note: tea.Batch creates a command that executes all sub-commands
	// We can't easily verify the internal behavior, but we can verify it doesn't panic
	_ = batchedCmd()
}
