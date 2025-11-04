package commands

import (
	tea "github.com/charmbracelet/bubbletea"
)

// StateChangedBatchMsg represents a batch of state change messages.
//
// This message is returned by the CoalesceAll strategy when multiple commands
// are batched together. Instead of executing commands individually, all commands
// are executed at once and their messages are collected into this batch message.
//
// This optimization reduces the number of Update() cycles when many state changes
// occur in a single tick.
//
// Example:
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case StateChangedBatchMsg:
//	        // Process all state changes at once
//	        for _, stateMsg := range msg.Messages {
//	            // Handle each state change
//	        }
//	    }
//	    return m, nil
//	}
type StateChangedBatchMsg struct {
	// Messages contains all the messages returned by the batched commands.
	// The messages are in the same order as the original commands.
	Messages []tea.Msg

	// Count is the number of messages in the batch.
	// This is equal to len(Messages) and provided for convenience.
	Count int
}

// batchAll executes all commands immediately and returns a single batch message.
//
// This is the most aggressive batching strategy. All commands are executed
// within a single tea.Cmd, and their messages are collected into a
// StateChangedBatchMsg. This reduces multiple Update() cycles into one.
//
// Example:
//
//	// Three commands:
//	cmd1 := func() tea.Msg { return StateChangedMsg{...} }
//	cmd2 := func() tea.Msg { return StateChangedMsg{...} }
//	cmd3 := func() tea.Msg { return StateChangedMsg{...} }
//
//	// Batched into one:
//	batchedCmd := batcher.batchAll([]tea.Cmd{cmd1, cmd2, cmd3})
//	msg := batchedCmd() // Returns StateChangedBatchMsg with 3 messages
func (cb *CommandBatcher) batchAll(commands []tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		// Execute all commands and collect their messages
		msgs := make([]tea.Msg, 0, len(commands))

		for _, cmd := range commands {
			if cmd != nil {
				msg := cmd()
				msgs = append(msgs, msg)
			}
		}

		// Return batch message containing all collected messages
		return StateChangedBatchMsg{
			Messages: msgs,
			Count:    len(msgs),
		}
	}
}

// batchByType groups commands by their message type before batching.
//
// For Task 3.2, this is a placeholder implementation that delegates to tea.Batch.
// Full type-based grouping logic will be implemented in a future task when
// the benefit is more clear and performance testing is done.
//
// The intended behavior (for future implementation):
//   - Group commands that return the same message type
//   - Batch each group separately
//   - Return a command that executes all groups
//
// Example (future):
//
//	// Commands returning same type get grouped:
//	cmd1 := func() tea.Msg { return StateChangedMsg{ComponentID: "a"} }
//	cmd2 := func() tea.Msg { return StateChangedMsg{ComponentID: "b"} }
//	cmd3 := func() tea.Msg { return OtherMsg{} }
//
//	// Result: [batch(cmd1, cmd2), cmd3]
func (cb *CommandBatcher) batchByType(commands []tea.Cmd) tea.Cmd {
	// Placeholder: Delegate to tea.Batch
	// Full type-based grouping will be implemented when needed
	// TODO: Implement type-based grouping for better optimization
	return tea.Batch(commands...)
}

// noCoalesce executes all commands individually using tea.Batch.
//
// This strategy provides no coalescing optimization. Each command executes
// separately and returns its own message. Use this when you need to preserve
// the original command behavior without any batching.
//
// This is equivalent to calling tea.Batch directly and is provided for
// consistency with other strategies.
//
// Example:
//
//	batcher := NewCommandBatcher(NoCoalesce)
//	batchedCmd := batcher.Batch([]tea.Cmd{cmd1, cmd2, cmd3})
//	// Each command executes separately and returns its own message
func (cb *CommandBatcher) noCoalesce(commands []tea.Cmd) tea.Cmd {
	// No coalescing - execute all commands individually via tea.Batch
	return tea.Batch(commands...)
}
