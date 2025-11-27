package testutil

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
)

// BatcherTester provides testing utilities for CommandBatcher, tracking batching
// operations and providing assertions for command batching and deduplication.
//
// This tester wraps a CommandBatcher and intercepts Batch() calls to track:
//   - Number of batches created
//   - Commands in each batch
//   - Batch sizes
//
// It's designed to verify that command batching and deduplication work correctly
// in the automatic reactive bridge system.
//
// Thread Safety:
//
// BatcherTester is not thread-safe. Create separate instances for concurrent tests.
//
// Example usage:
//
//	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
//	tester := testutil.NewBatcherTester(batcher)
//
//	// Trigger batching
//	cmd1 := func() tea.Msg { return "msg1" }
//	cmd2 := func() tea.Msg { return "msg2" }
//	tester.Batch([]tea.Cmd{cmd1, cmd2})
//
//	// Verify batching
//	tester.AssertBatched(t, 1)
//	tester.AssertBatchSize(t, 0, 2)
type BatcherTester struct {
	batcher    *commands.CommandBatcher
	batches    [][]tea.Cmd
	batchCount int
}

// NewBatcherTester creates a new command batcher tester.
//
// The tester wraps the given CommandBatcher and tracks all Batch() calls.
// The batcher parameter can be nil, in which case Batch() will return nil
// but tracking will still occur.
//
// Parameters:
//   - batcher: The CommandBatcher to wrap (can be nil)
//
// Returns:
//   - *BatcherTester: Ready to use for testing
//
// Example:
//
//	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
//	tester := NewBatcherTester(batcher)
//	assert.Equal(t, 0, tester.GetBatchCount())
func NewBatcherTester(batcher *commands.CommandBatcher) *BatcherTester {
	return &BatcherTester{
		batcher:    batcher,
		batches:    [][]tea.Cmd{},
		batchCount: 0,
	}
}

// Batch wraps the underlying CommandBatcher.Batch() method while tracking
// the batching operation.
//
// This method:
//  1. Stores a copy of the input commands for inspection
//  2. Increments the batch count
//  3. Calls the underlying batcher's Batch() method
//  4. Returns the result
//
// If the batcher is nil, this method returns nil but still tracks the operation.
//
// Parameters:
//   - commands: The commands to batch
//
// Returns:
//   - tea.Cmd: The batched command (or nil)
//
// Example:
//
//	cmd1 := func() tea.Msg { return "msg1" }
//	cmd2 := func() tea.Msg { return "msg2" }
//	result := tester.Batch([]tea.Cmd{cmd1, cmd2})
//	assert.NotNil(t, result)
//	assert.Equal(t, 1, tester.GetBatchCount())
func (bt *BatcherTester) Batch(commands []tea.Cmd) tea.Cmd {
	// Store a copy of the commands for inspection
	commandsCopy := make([]tea.Cmd, len(commands))
	copy(commandsCopy, commands)
	bt.batches = append(bt.batches, commandsCopy)
	bt.batchCount++

	// Call underlying batcher if it exists
	if bt.batcher == nil {
		return nil
	}

	return bt.batcher.Batch(commands)
}

// GetBatchCount returns the number of times Batch() has been called.
//
// This provides a quick way to verify how many batching operations occurred
// during a test.
//
// Returns:
//   - int: Number of Batch() calls
//
// Example:
//
//	tester.Batch([]tea.Cmd{cmd1})
//	tester.Batch([]tea.Cmd{cmd2})
//	assert.Equal(t, 2, tester.GetBatchCount())
func (bt *BatcherTester) GetBatchCount() int {
	return bt.batchCount
}

// GetBatches returns all tracked batches.
//
// Each element in the returned slice represents one Batch() call, containing
// the commands that were passed to that call. The returned slice is a copy,
// so modifications to it do not affect the tester's internal state.
//
// Returns:
//   - [][]tea.Cmd: All tracked batches
//
// Example:
//
//	batches := tester.GetBatches()
//	assert.Len(t, batches, 2)
//	assert.Len(t, batches[0], 3) // First batch had 3 commands
//	assert.Len(t, batches[1], 1) // Second batch had 1 command
func (bt *BatcherTester) GetBatches() [][]tea.Cmd {
	// Return a copy to prevent external modification
	result := make([][]tea.Cmd, len(bt.batches))
	for i, batch := range bt.batches {
		batchCopy := make([]tea.Cmd, len(batch))
		copy(batchCopy, batch)
		result[i] = batchCopy
	}
	return result
}

// GetBatchSize returns the number of commands in a specific batch.
//
// This is a convenience method for checking batch sizes without retrieving
// all batches. Returns 0 if the batch index is out of bounds.
//
// Parameters:
//   - batchIdx: The 0-based index of the batch
//
// Returns:
//   - int: Number of commands in the batch (0 if out of bounds)
//
// Example:
//
//	tester.Batch([]tea.Cmd{cmd1, cmd2})
//	assert.Equal(t, 2, tester.GetBatchSize(0))
//	assert.Equal(t, 0, tester.GetBatchSize(999)) // Out of bounds
func (bt *BatcherTester) GetBatchSize(batchIdx int) int {
	if batchIdx < 0 || batchIdx >= len(bt.batches) {
		return 0
	}
	return len(bt.batches[batchIdx])
}

// Clear resets all tracking state.
//
// This method clears the batch history and resets the batch count to zero.
// It's useful for resetting state between test cases or cleaning up after
// testing. Safe to call multiple times.
//
// Example:
//
//	tester.Batch([]tea.Cmd{cmd1})
//	assert.Equal(t, 1, tester.GetBatchCount())
//
//	tester.Clear()
//	assert.Equal(t, 0, tester.GetBatchCount())
//	assert.Empty(t, tester.GetBatches())
func (bt *BatcherTester) Clear() {
	bt.batches = [][]tea.Cmd{}
	bt.batchCount = 0
}

// AssertBatched asserts that the expected number of batches were created.
//
// This is a convenience assertion method that checks the batch count and
// reports a clear error message if it doesn't match. It uses t.Helper()
// to ensure the error is reported at the correct line in the test.
//
// Parameters:
//   - t: The testing.T instance (or testingT interface)
//   - expectedBatches: The expected number of batches
//
// Example:
//
//	tester.Batch([]tea.Cmd{cmd1})
//	tester.Batch([]tea.Cmd{cmd2})
//	tester.AssertBatched(t, 2) // Passes
//	tester.AssertBatched(t, 3) // Fails with clear error
func (bt *BatcherTester) AssertBatched(t testingT, expectedBatches int) {
	t.Helper()

	actual := bt.GetBatchCount()
	if actual != expectedBatches {
		t.Errorf("command batcher: expected %d batches, got %d", expectedBatches, actual)
	}
}

// AssertBatchSize asserts that a specific batch has the expected size.
//
// This method verifies that the batch at the given index contains the expected
// number of commands. It reports a clear error if the size doesn't match or
// if the batch index is out of bounds.
//
// Parameters:
//   - t: The testing.T instance (or testingT interface)
//   - batchIdx: The 0-based index of the batch to check
//   - expectedSize: The expected number of commands in the batch
//
// Example:
//
//	tester.Batch([]tea.Cmd{cmd1, cmd2})
//	tester.AssertBatchSize(t, 0, 2) // Passes
//	tester.AssertBatchSize(t, 0, 3) // Fails
//	tester.AssertBatchSize(t, 999, 1) // Fails (out of bounds)
func (bt *BatcherTester) AssertBatchSize(t testingT, batchIdx, expectedSize int) {
	t.Helper()

	if batchIdx < 0 || batchIdx >= len(bt.batches) {
		t.Errorf("command batcher: batch index %d out of bounds (have %d batches)", batchIdx, len(bt.batches))
		return
	}

	actual := len(bt.batches[batchIdx])
	if actual != expectedSize {
		t.Errorf("command batcher: batch %d expected size %d, got %d", batchIdx, expectedSize, actual)
	}
}
