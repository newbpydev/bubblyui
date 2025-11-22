package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// deduplicateCommands removes duplicate commands from the batch while preserving order.
//
// For StateChangedMsg commands targeting the same component and ref, only the
// last state change is kept (most recent value). This optimization reduces
// redundant Update() cycles when the same ref is modified multiple times in
// a single tick.
//
// Algorithm:
//   - Iterate through commands, generating unique keys for each
//   - Track the last index where each key appears
//   - Build result slice containing only the last occurrence of each unique key
//   - Preserve relative order of first appearance
//
// Example:
//
//	commands := []tea.Cmd{
//	    cmd1, // count.Set(1)
//	    cmd2, // count.Set(2)  <- duplicate of cmd1, this one kept
//	    cmd3, // name.Set("x")
//	}
//	result := batcher.deduplicateCommands(commands) // [cmd2, cmd3]
//
// Performance: O(n) time complexity with map lookups.
//
// Thread Safety: Not thread-safe. Caller must ensure exclusive access.
func (cb *CommandBatcher) deduplicateCommands(commands []tea.Cmd) []tea.Cmd {
	// Handle edge cases
	if len(commands) == 0 {
		return nil
	}

	if len(commands) == 1 {
		// Single command optimization - no deduplication needed
		if commands[0] != nil {
			return commands
		}
		return nil
	}

	// First pass: identify the last index where each unique key appears
	lastIndex := make(map[string]int, len(commands))

	for i, cmd := range commands {
		if cmd == nil {
			continue
		}
		key := generateCommandKey(cmd)
		lastIndex[key] = i
	}

	// Second pass: build result in order, including only commands at their last occurrence
	result := make([]tea.Cmd, 0, len(lastIndex))

	for i, cmd := range commands {
		if cmd == nil {
			continue
		}

		key := generateCommandKey(cmd)

		// Only include this command if this is its last occurrence
		if lastIndex[key] == i {
			result = append(result, cmd)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// generateCommandKey generates a unique key for a command based on its message.
//
// For StateChangedMsg, the key is "componentID:refID" since only the final
// state of a ref matters when it's changed multiple times.
//
// For other message types, a generic key is generated based on the message
// type name. This ensures different message types don't interfere with each
// other during deduplication.
//
// Example:
//
//	cmd1 := func() tea.Msg {
//	    return bubbly.StateChangedMsg{
//	        ComponentID: "counter",
//	        RefID:       "count",
//	    }
//	}
//	key := generateCommandKey(cmd1) // "counter:count"
//
// Performance: O(1) for StateChangedMsg, executes command once.
//
// Thread Safety: Safe for concurrent calls if commands are goroutine-safe.
func generateCommandKey(cmd tea.Cmd) string {
	if cmd == nil {
		return ""
	}

	// Execute the command to inspect its message
	msg := cmd()

	// For StateChangedMsg, use componentID:refID as unique key
	if stateMsg, ok := msg.(bubbly.StateChangedMsg); ok {
		return fmt.Sprintf("%s:%s", stateMsg.ComponentID, stateMsg.RefID)
	}

	// For other message types, use type name as key
	// This means different message types won't deduplicate each other
	// but multiple instances of the same custom type will
	return fmt.Sprintf("%T", msg)
}
