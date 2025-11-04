package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

// CoalescingStrategy determines how multiple commands are batched together.
//
// Different strategies optimize for different use cases:
//   - CoalesceAll: Batch all commands into a single command (default)
//   - CoalesceByType: Group commands by message type before batching (future)
//   - NoCoalesce: Execute all commands individually via tea.Batch
//
// Example:
//
//	batcher := NewCommandBatcher(CoalesceAll)
//	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})
type CoalescingStrategy int

const (
	// CoalesceAll batches all commands into a single command that executes them all.
	// This is the most aggressive batching strategy and produces the fewest commands.
	CoalesceAll CoalescingStrategy = iota

	// CoalesceByType groups commands by their message type before batching.
	// Commands that produce the same message type are batched together.
	// This strategy will be implemented in Task 3.2.
	CoalesceByType

	// NoCoalesce executes all commands individually using tea.Batch.
	// This preserves the original command behavior with minimal overhead.
	NoCoalesce
)

// CommandBatcher batches multiple Bubbletea commands into a single command.
//
// The batcher uses a CoalescingStrategy to determine how to combine commands.
// This optimization reduces the number of Update() cycles and improves
// performance when many state changes occur in a single tick.
//
// Thread Safety:
//
// CommandBatcher is not thread-safe. Create separate instances for concurrent use.
//
// Example:
//
//	batcher := NewCommandBatcher(CoalesceAll)
//	commands := []tea.Cmd{cmd1, cmd2, cmd3}
//	batchedCmd := batcher.Batch(commands)
//
//	// In component Update():
//	return model, batchedCmd
type CommandBatcher struct {
	strategy CoalescingStrategy
}

// NewCommandBatcher creates a new CommandBatcher with the specified strategy.
//
// Example:
//
//	batcher := NewCommandBatcher(CoalesceAll)
func NewCommandBatcher(strategy CoalescingStrategy) *CommandBatcher {
	return &CommandBatcher{
		strategy: strategy,
	}
}

// Batch combines multiple commands into a single command using the configured strategy.
//
// Returns nil if the input is empty or contains only nil commands.
// Returns a single command as-is without wrapping.
// Returns a batched command for multiple commands.
//
// Nil commands in the input are filtered out before batching.
//
// Example:
//
//	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})
//	if batchedCmd != nil {
//	    return model, batchedCmd
//	}
func (cb *CommandBatcher) Batch(commands []tea.Cmd) tea.Cmd {
	// Filter out nil commands
	filtered := make([]tea.Cmd, 0, len(commands))
	for _, cmd := range commands {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}

	// Handle edge cases
	if len(filtered) == 0 {
		return nil
	}

	if len(filtered) == 1 {
		return filtered[0]
	}

	// Batch commands based on strategy
	// For Task 3.1, all strategies use tea.Batch
	// Task 3.2 will implement actual coalescing logic
	switch cb.strategy {
	case CoalesceAll:
		// For now, delegate to tea.Batch
		// Task 3.2 will implement actual message coalescing
		return tea.Batch(filtered...)
	case CoalesceByType:
		// Placeholder for Task 3.2 implementation
		return tea.Batch(filtered...)
	case NoCoalesce:
		// Execute all commands individually via tea.Batch
		return tea.Batch(filtered...)
	default:
		// Unknown strategy, use safe default
		return tea.Batch(filtered...)
	}
}
