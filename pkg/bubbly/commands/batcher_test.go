package commands

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
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

// TestCommandBatcher_EnableDisableDeduplication tests deduplication toggle methods.
func TestCommandBatcher_EnableDisableDeduplication(t *testing.T) {
	tests := []struct {
		name     string
		strategy CoalescingStrategy
	}{
		{
			name:     "CoalesceAll strategy",
			strategy: CoalesceAll,
		},
		{
			name:     "CoalesceByType strategy",
			strategy: CoalesceByType,
		},
		{
			name:     "NoCoalesce strategy",
			strategy: NoCoalesce,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(tt.strategy)

			// Initially disabled
			assert.False(t, batcher.deduplicateEnabled, "deduplication should be disabled by default")

			// Enable deduplication
			batcher.EnableDeduplication()
			assert.True(t, batcher.deduplicateEnabled, "deduplication should be enabled")

			// Disable deduplication
			batcher.DisableDeduplication()
			assert.False(t, batcher.deduplicateEnabled, "deduplication should be disabled")
		})
	}
}

// TestCommandBatcher_Batch_WithDeduplication tests the Batch method with deduplication enabled.
func TestCommandBatcher_Batch_WithDeduplication(t *testing.T) {
	batcher := NewCommandBatcher(CoalesceAll)
	batcher.EnableDeduplication()

	// Test with duplicate commands
	cmd1 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref1",
			NewValue:    1,
		}
	}
	cmd2 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref1",
			NewValue:    2,
		}
	}
	cmd3 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref2",
			NewValue:    "test",
		}
	}

	commands := []tea.Cmd{cmd1, cmd2, cmd3}
	batchedCmd := batcher.Batch(commands)

	assert.NotNil(t, batchedCmd, "should return non-nil command")

	// Execute and verify deduplication worked
	msg := batchedCmd()
	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "should return StateChangedBatchMsg")

	// Should have only 2 messages (cmd1 and cmd2 deduplicated, cmd3 kept)
	assert.Equal(t, 2, batchMsg.Count, "should have 2 messages after deduplication")
}

// TestCommandBatcher_Batch_WithoutDeduplication tests the Batch method with deduplication disabled.
func TestCommandBatcher_Batch_WithoutDeduplication(t *testing.T) {
	batcher := NewCommandBatcher(CoalesceAll)
	// Deduplication disabled by default

	cmd1 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref1",
			NewValue:    1,
		}
	}
	cmd2 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref2",
			NewValue:    2,
		}
	}

	commands := []tea.Cmd{cmd1, cmd2}
	batchedCmd := batcher.Batch(commands)

	assert.NotNil(t, batchedCmd, "should return non-nil command")

	// Execute and verify all messages are included
	msg := batchedCmd()
	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "should return StateChangedBatchMsg")

	assert.Equal(t, 2, batchMsg.Count, "should have 2 messages without deduplication")
}

// TestCommandBatcher_Batch_DeduplicationAllStrategies tests deduplication with all strategies.
func TestCommandBatcher_Batch_DeduplicationAllStrategies(t *testing.T) {
	strategies := []CoalescingStrategy{CoalesceAll, CoalesceByType, NoCoalesce}

	for _, strategy := range strategies {
		t.Run(strategy.String(), func(t *testing.T) {
			batcher := NewCommandBatcher(strategy)
			batcher.EnableDeduplication()

			// Create duplicate commands
			cmd1 := func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: "comp1",
					RefID:       "ref1",
					NewValue:    1,
				}
			}
			cmd2 := func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: "comp1",
					RefID:       "ref1",
					NewValue:    2,
				}
			}

			commands := []tea.Cmd{cmd1, cmd2}
			batchedCmd := batcher.Batch(commands)

			assert.NotNil(t, batchedCmd, "should return non-nil command for strategy %v", strategy)
		})
	}
}

// TestCommandBatcher_Batch_DeduplicationToSingleCommand tests when deduplication results in single command.
func TestCommandBatcher_Batch_DeduplicationToSingleCommand(t *testing.T) {
	batcher := NewCommandBatcher(CoalesceAll)
	batcher.EnableDeduplication()

	// All commands target same ref - should deduplicate to 1
	cmd1 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref1",
			NewValue:    1,
		}
	}
	cmd2 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref1",
			NewValue:    2,
		}
	}
	cmd3 := func() tea.Msg {
		return bubbly.StateChangedMsg{
			ComponentID: "comp1",
			RefID:       "ref1",
			NewValue:    3,
		}
	}

	commands := []tea.Cmd{cmd1, cmd2, cmd3}
	batchedCmd := batcher.Batch(commands)

	assert.NotNil(t, batchedCmd, "should return non-nil command")

	// Should return the single deduplicated command directly, not wrapped
	msg := batchedCmd()
	stateMsg, ok := msg.(bubbly.StateChangedMsg)
	assert.True(t, ok, "should return StateChangedMsg directly for single command")
	assert.Equal(t, 3, stateMsg.NewValue, "should keep the last value")
}

// TestCommandBatcher_Batch_DeduplicationToEmpty tests when deduplication results in empty list.
func TestCommandBatcher_Batch_DeduplicationToEmpty(t *testing.T) {
	batcher := NewCommandBatcher(CoalesceAll)
	batcher.EnableDeduplication()

	// All commands are nil
	commands := []tea.Cmd{nil, nil, nil}
	batchedCmd := batcher.Batch(commands)

	assert.Nil(t, batchedCmd, "should return nil for all nil commands")
}

// TestCoalescingStrategy_String tests the String() method for all strategies.
func TestCoalescingStrategy_String(t *testing.T) {
	tests := []struct {
		name     string
		strategy CoalescingStrategy
		expected string
	}{
		{
			name:     "Coalesce all",
			strategy: CoalesceAll,
			expected: "CoalesceAll",
		},
		{
			name:     "Coalesce by type",
			strategy: CoalesceByType,
			expected: "CoalesceByType",
		},
		{
			name:     "No coalesce",
			strategy: NoCoalesce,
			expected: "NoCoalesce",
		},
		{
			name:     "Unknown strategy",
			strategy: CoalescingStrategy(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.String()
			assert.Equal(t, tt.expected, result, "String() should return expected value")
		})
	}
}

// TestCommandBatcher_Batch_UnknownStrategy tests the default case in Batch method.
func TestCommandBatcher_Batch_UnknownStrategy(t *testing.T) {
	// Create a batcher with an unknown strategy
	batcher := &CommandBatcher{
		strategy:           CoalescingStrategy(999), // Unknown strategy
		deduplicateEnabled: false,
	}

	cmd1 := func() tea.Msg {
		return "test1"
	}
	cmd2 := func() tea.Msg {
		return "test2"
	}

	commands := []tea.Cmd{cmd1, cmd2}
	batchedCmd := batcher.Batch(commands)

	// Should handle unknown strategy gracefully by using safe default
	assert.NotNil(t, batchedCmd, "should handle unknown strategy with safe default")
}

// TestCommandBatcher_Batch_NilCommandFiltering tests nil command filtering.
func TestCommandBatcher_Batch_NilCommandFiltering(t *testing.T) {
	batcher := NewCommandBatcher(CoalesceAll)

	validCmd := func() tea.Msg {
		return "valid"
	}

	// Test with mixed nil and valid commands
	commands := []tea.Cmd{nil, validCmd, nil, nil, validCmd, nil}
	batchedCmd := batcher.Batch(commands)

	assert.NotNil(t, batchedCmd, "should filter nil commands and batch valid ones")

	// Execute and verify only valid commands were processed
	msg := batchedCmd()
	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "should return StateChangedBatchMsg")
	assert.Equal(t, 2, batchMsg.Count, "should have 2 valid messages after filtering nils")
}

// TestCommandBatcher_Batch_AllEdgeCases tests all remaining edge cases in Batch method.
func TestCommandBatcher_Batch_AllEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		strategy  CoalescingStrategy
		commands  []tea.Cmd
		expectNil bool
	}{
		{
			name:      "only nil commands",
			strategy:  CoalesceAll,
			commands:  []tea.Cmd{nil, nil, nil},
			expectNil: true,
		},
		{
			name:     "single valid command",
			strategy: CoalesceAll,
			commands: []tea.Cmd{
				func() tea.Msg { return "single" },
			},
			expectNil: false,
		},
		{
			name:      "empty command list",
			strategy:  CoalesceAll,
			commands:  []tea.Cmd{},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewCommandBatcher(tt.strategy)
			result := batcher.Batch(tt.commands)

			if tt.expectNil {
				assert.Nil(t, result, "should return nil for %s", tt.name)
			} else {
				assert.NotNil(t, result, "should return command for %s", tt.name)
			}
		})
	}
}

// TestCommandBatcher_Batch_ComprehensiveEdgeCases tests comprehensive edge cases for 100% coverage.
func TestCommandBatcher_Batch_ComprehensiveEdgeCases(t *testing.T) {
	// Test the specific case where we have nil commands that get filtered out
	// but some valid commands remain, ensuring the filtering loop is fully exercised
	batcher := NewCommandBatcher(CoalesceAll)

	// Create a scenario that exercises every branch in Batch method:
	// 1. Nil command filtering (cmd != nil check)
	// 2. Multiple valid commands that get batched
	// 3. All three strategies in different scenarios

	validCmd1 := func() tea.Msg {
		return bubbly.StateChangedMsg{ComponentID: "test", RefID: "ref1"}
	}
	validCmd2 := func() tea.Msg {
		return bubbly.StateChangedMsg{ComponentID: "test", RefID: "ref2"}
	}

	// Test with nil filtering + multiple commands
	commands := []tea.Cmd{nil, validCmd1, nil, validCmd2, nil}
	result := batcher.Batch(commands)

	assert.NotNil(t, result, "should return batched command after filtering nils")

	// Execute to verify it works
	msg := result()
	batchMsg, ok := msg.(StateChangedBatchMsg)
	assert.True(t, ok, "should return batch message")
	assert.Equal(t, 2, batchMsg.Count, "should batch 2 valid commands")
}
