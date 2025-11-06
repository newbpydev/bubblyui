package commands

import (
	"time"
)

// CommandInspector provides debugging and inspection capabilities for pending
// commands in the Automatic Reactive Bridge system.
//
// This type allows developers to inspect the command queue for debugging purposes,
// including:
//   - Checking how many commands are pending
//   - Viewing command metadata (component ID, ref ID, timestamp)
//   - Clearing pending commands for testing or debugging
//
// Design Principles:
//   - Read-Only Inspection: PendingCommands() doesn't modify the queue
//   - Thread-Safe: All methods are safe for concurrent use
//   - Debugging Focus: Designed for development and troubleshooting
//   - Minimal Overhead: Lightweight wrapper around CommandQueue
//
// Usage Examples:
//
//	// Create inspector for a component's queue
//	queue := component.GetCommandQueue() // Hypothetical accessor
//	inspector := commands.NewCommandInspector(queue)
//
//	// Check pending count
//	if inspector.PendingCount() > 10 {
//	    log.Printf("Warning: %d commands pending", inspector.PendingCount())
//	}
//
//	// Inspect command details
//	for _, cmd := range inspector.PendingCommands() {
//	    log.Printf("Pending: Component=%s, Ref=%s, Time=%v",
//	        cmd.ComponentID, cmd.RefID, cmd.Timestamp)
//	}
//
//	// Clear for testing
//	inspector.ClearPending()
//
// Thread Safety:
//
// All methods are thread-safe and can be called concurrently from multiple
// goroutines. The underlying CommandQueue provides synchronization.
//
// Integration:
//
// This inspector is typically used in development and debugging scenarios:
//   - Unit tests: Verify command generation behavior
//   - Debug tools: Inspect pending commands during development
//   - Performance profiling: Track command queue depth
//   - Troubleshooting: Identify infinite loop patterns
type CommandInspector struct {
	// queue is the command queue to inspect
	// Can be nil for safe default behavior
	queue *CommandQueue
}

// CommandInfo contains metadata about a pending command.
//
// This struct provides debugging information extracted from StateChangedMsg
// commands in the queue. It's designed to be lightweight and focused on
// the most useful debugging information.
//
// Fields:
//   - ComponentID: Unique identifier of the component that generated the command
//   - RefID: Unique identifier of the ref that changed
//   - Timestamp: When the state change occurred
//
// Usage:
//
//	info := inspector.PendingCommands()[0]
//	fmt.Printf("Component %s changed ref %s at %v\n",
//	    info.ComponentID, info.RefID, info.Timestamp)
//
// Note:
//
// This struct only contains metadata from StateChangedMsg commands.
// Other command types in the queue are counted but not included in
// PendingCommands() results.
type CommandInfo struct {
	ComponentID string
	RefID       string
	Timestamp   time.Time
}

// NewCommandInspector creates a new command inspector for the given queue.
//
// This function creates an inspector that can be used to examine pending
// commands in the queue for debugging and troubleshooting purposes.
//
// Parameters:
//   - queue: The CommandQueue to inspect (can be nil for safe defaults)
//
// Returns:
//   - *CommandInspector: Ready to use for inspecting commands
//
// Nil Queue Handling:
//
// If queue is nil, the inspector returns safe defaults:
//   - PendingCount() returns 0
//   - PendingCommands() returns empty slice
//   - ClearPending() is a no-op
//
// Examples:
//
//	// Normal usage
//	queue := NewCommandQueue()
//	inspector := NewCommandInspector(queue)
//	count := inspector.PendingCount()
//
//	// Safe with nil queue
//	inspector := NewCommandInspector(nil)
//	count := inspector.PendingCount() // Returns 0
//
// Thread Safety:
//
// The returned inspector is thread-safe and can be used concurrently
// from multiple goroutines.
func NewCommandInspector(queue *CommandQueue) *CommandInspector {
	return &CommandInspector{
		queue: queue,
	}
}

// PendingCount returns the number of commands currently pending in the queue.
//
// This method provides a quick way to check the queue depth for debugging
// and monitoring purposes. It includes all commands in the queue, regardless
// of their message type.
//
// Returns:
//   - int: Number of pending commands (0 if queue is nil or empty)
//
// Thread Safety:
//
// This method is thread-safe and can be called concurrently from multiple
// goroutines. The underlying CommandQueue.Len() provides synchronization.
//
// Examples:
//
//	// Check if commands are pending
//	if inspector.PendingCount() > 0 {
//	    fmt.Println("Commands pending")
//	}
//
//	// Monitor queue depth
//	depth := inspector.PendingCount()
//	if depth > 100 {
//	    log.Printf("Warning: High queue depth: %d", depth)
//	}
//
// Performance:
//
//   - Time complexity: O(1)
//   - Space complexity: O(1)
//   - Overhead: Minimal (just a mutex lock and length check)
func (ci *CommandInspector) PendingCount() int {
	if ci.queue == nil {
		return 0
	}
	return ci.queue.Len()
}

// PendingCommands returns metadata about all pending StateChangedMsg commands.
//
// This method extracts debugging information from StateChangedMsg commands
// in the queue without modifying the queue. It's useful for:
//   - Debugging reactive update flows
//   - Identifying which components/refs have pending updates
//   - Tracking command timestamps
//   - Troubleshooting infinite loop patterns
//
// Returns:
//   - []CommandInfo: Metadata for each StateChangedMsg command
//   - Empty slice if queue is nil or contains no StateChangedMsg commands
//
// Behavior:
//   - Queue is NOT modified (commands remain pending)
//   - Only StateChangedMsg commands are included
//   - Other command types are skipped (counted but not returned)
//   - Order matches queue order (FIFO)
//
// Thread Safety:
//
// This method is thread-safe and can be called concurrently. However, the
// returned slice is a snapshot at the time of the call. Concurrent modifications
// to the queue may result in the snapshot being stale.
//
// Examples:
//
//	// Inspect all pending commands
//	for _, cmd := range inspector.PendingCommands() {
//	    fmt.Printf("Pending: %s.%s at %v\n",
//	        cmd.ComponentID, cmd.RefID, cmd.Timestamp)
//	}
//
//	// Check for specific component
//	for _, cmd := range inspector.PendingCommands() {
//	    if cmd.ComponentID == "counter-1" {
//	        fmt.Printf("Counter has pending update for %s\n", cmd.RefID)
//	    }
//	}
//
//	// Detect potential infinite loops
//	commands := inspector.PendingCommands()
//	if len(commands) > 100 {
//	    log.Printf("Warning: %d pending commands - possible infinite loop", len(commands))
//	}
//
// Performance:
//   - Time complexity: O(n) where n is the number of commands
//   - Space complexity: O(m) where m is the number of StateChangedMsg commands
//   - Each command is executed to extract metadata
//
// Note:
//
// This method executes each command to extract the message. For large queues,
// this may have performance implications. Use PendingCount() for a quick check
// before calling this method if performance is a concern.
func (ci *CommandInspector) PendingCommands() []CommandInfo {
	if ci.queue == nil {
		return nil
	}

	// Get snapshot of current commands using Peek()
	commands := ci.queue.Peek()
	if commands == nil {
		return nil
	}

	// Extract metadata from StateChangedMsg commands
	var infos []CommandInfo
	for _, cmd := range commands {
		if cmd == nil {
			continue
		}

		// Execute command to get message
		msg := cmd()

		// Extract metadata if it's a StateChangedMsg
		if stateMsg, ok := msg.(StateChangedMsg); ok {
			infos = append(infos, CommandInfo{
				ComponentID: stateMsg.ComponentID,
				RefID:       stateMsg.RefID,
				Timestamp:   stateMsg.Timestamp,
			})
		}
	}

	return infos
}

// ClearPending removes all pending commands from the queue.
//
// This method is primarily useful for:
//   - Testing: Reset queue state between tests
//   - Debugging: Clear problematic command accumulation
//   - Error recovery: Discard commands after detecting issues
//
// Behavior:
//   - All pending commands are removed
//   - Queue is reset to empty state
//   - No-op if queue is nil
//   - Safe to call on empty queue
//
// Thread Safety:
//
// This method is thread-safe and can be called concurrently from multiple
// goroutines. The underlying CommandQueue.Clear() provides synchronization.
//
// Examples:
//
//	// Clear for testing
//	inspector.ClearPending()
//	assert.Equal(t, 0, inspector.PendingCount())
//
//	// Clear after detecting infinite loop
//	if inspector.PendingCount() > 1000 {
//	    log.Println("Clearing excessive pending commands")
//	    inspector.ClearPending()
//	}
//
//	// Reset between test cases
//	func TestSomething(t *testing.T) {
//	    defer inspector.ClearPending() // Cleanup
//	    // ... test code
//	}
//
// Performance:
//   - Time complexity: O(1)
//   - Space complexity: O(1)
//   - Overhead: Minimal (mutex lock and slice reset)
//
// Warning:
//
// This method discards all pending commands, which may result in lost UI
// updates if called in production code. It's primarily intended for debugging
// and testing scenarios.
func (ci *CommandInspector) ClearPending() {
	if ci.queue == nil {
		return
	}
	ci.queue.Clear()
}
