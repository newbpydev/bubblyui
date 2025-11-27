package testutil

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/commands"
)

// TestNewBatcherTester verifies constructor creates valid instance
func TestNewBatcherTester(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	assert.NotNil(t, tester)
	assert.NotNil(t, tester.batcher)
	assert.Equal(t, 0, tester.GetBatchCount())
	assert.Empty(t, tester.GetBatches())
}

// TestNewBatcherTester_NilBatcher verifies nil batcher handling
func TestNewBatcherTester_NilBatcher(t *testing.T) {
	tester := NewBatcherTester(nil)

	assert.NotNil(t, tester)
	assert.Nil(t, tester.batcher)
	assert.Equal(t, 0, tester.GetBatchCount())
	assert.Empty(t, tester.GetBatches())
}

// TestBatcherTester_Batch_SingleCall verifies single batch tracking
func TestBatcherTester_Batch_SingleCall(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	cmd2 := func() tea.Msg { return "cmd2" }
	commands := []tea.Cmd{cmd1, cmd2}

	result := tester.Batch(commands)

	assert.NotNil(t, result)
	assert.Equal(t, 1, tester.GetBatchCount())
	assert.Len(t, tester.GetBatches(), 1)
	assert.Len(t, tester.GetBatches()[0], 2)
}

// TestBatcherTester_Batch_MultipleCalls verifies multiple batch tracking
func TestBatcherTester_Batch_MultipleCalls(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	// First batch
	cmd1 := func() tea.Msg { return "cmd1" }
	cmd2 := func() tea.Msg { return "cmd2" }
	tester.Batch([]tea.Cmd{cmd1, cmd2})

	// Second batch
	cmd3 := func() tea.Msg { return "cmd3" }
	tester.Batch([]tea.Cmd{cmd3})

	// Third batch
	cmd4 := func() tea.Msg { return "cmd4" }
	cmd5 := func() tea.Msg { return "cmd5" }
	cmd6 := func() tea.Msg { return "cmd6" }
	tester.Batch([]tea.Cmd{cmd4, cmd5, cmd6})

	assert.Equal(t, 3, tester.GetBatchCount())
	assert.Len(t, tester.GetBatches(), 3)
	assert.Len(t, tester.GetBatches()[0], 2)
	assert.Len(t, tester.GetBatches()[1], 1)
	assert.Len(t, tester.GetBatches()[2], 3)
}

// TestBatcherTester_Batch_EmptyCommands verifies empty command handling
func TestBatcherTester_Batch_EmptyCommands(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	result := tester.Batch([]tea.Cmd{})

	assert.Nil(t, result)
	assert.Equal(t, 1, tester.GetBatchCount()) // Still tracked
	assert.Len(t, tester.GetBatches(), 1)
	assert.Empty(t, tester.GetBatches()[0])
}

// TestBatcherTester_Batch_NilCommands verifies nil command filtering
func TestBatcherTester_Batch_NilCommands(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	result := tester.Batch([]tea.Cmd{nil, cmd1, nil})

	assert.NotNil(t, result)
	assert.Equal(t, 1, tester.GetBatchCount())
	// Original commands tracked (including nils)
	assert.Len(t, tester.GetBatches()[0], 3)
}

// TestBatcherTester_Batch_NilBatcher verifies nil batcher handling
func TestBatcherTester_Batch_NilBatcher(t *testing.T) {
	tester := NewBatcherTester(nil)

	cmd1 := func() tea.Msg { return "cmd1" }
	result := tester.Batch([]tea.Cmd{cmd1})

	assert.Nil(t, result) // No batcher, returns nil
	assert.Equal(t, 1, tester.GetBatchCount())
	assert.Len(t, tester.GetBatches(), 1)
}

// TestBatcherTester_GetBatchSize verifies batch size retrieval
func TestBatcherTester_GetBatchSize(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	cmd2 := func() tea.Msg { return "cmd2" }
	cmd3 := func() tea.Msg { return "cmd3" }

	tester.Batch([]tea.Cmd{cmd1, cmd2})
	tester.Batch([]tea.Cmd{cmd3})

	assert.Equal(t, 2, tester.GetBatchSize(0))
	assert.Equal(t, 1, tester.GetBatchSize(1))
	assert.Equal(t, 0, tester.GetBatchSize(999)) // Out of bounds
}

// TestBatcherTester_Clear verifies state reset
func TestBatcherTester_Clear(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	tester.Batch([]tea.Cmd{cmd1})
	tester.Batch([]tea.Cmd{cmd1})

	assert.Equal(t, 2, tester.GetBatchCount())

	tester.Clear()

	assert.Equal(t, 0, tester.GetBatchCount())
	assert.Empty(t, tester.GetBatches())
}

// TestBatcherTester_Clear_NilBatcher verifies clear with nil batcher
func TestBatcherTester_Clear_NilBatcher(t *testing.T) {
	tester := NewBatcherTester(nil)

	cmd1 := func() tea.Msg { return "cmd1" }
	tester.Batch([]tea.Cmd{cmd1})

	tester.Clear()

	assert.Equal(t, 0, tester.GetBatchCount())
	assert.Empty(t, tester.GetBatches())
}

// TestBatcherTester_AssertBatched_Success verifies successful assertion
func TestBatcherTester_AssertBatched_Success(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	tester.Batch([]tea.Cmd{cmd1})
	tester.Batch([]tea.Cmd{cmd1})
	tester.Batch([]tea.Cmd{cmd1})

	// Should not fail
	mockT := &mockTestingT{}
	tester.AssertBatched(mockT, 3)

	assert.False(t, mockT.failed)
	assert.Empty(t, mockT.errors)
}

// TestBatcherTester_AssertBatched_Failure verifies failed assertion
func TestBatcherTester_AssertBatched_Failure(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	tester.Batch([]tea.Cmd{cmd1})

	mockT := &mockTestingT{}
	tester.AssertBatched(mockT, 3)

	assert.True(t, mockT.failed)
	assert.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "expected 3 batches")
	assert.Contains(t, mockT.errors[0], "got 1")
}

// TestBatcherTester_AssertBatchSize_Success verifies successful batch size assertion
func TestBatcherTester_AssertBatchSize_Success(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	cmd2 := func() tea.Msg { return "cmd2" }
	tester.Batch([]tea.Cmd{cmd1, cmd2})

	mockT := &mockTestingT{}
	tester.AssertBatchSize(mockT, 0, 2)

	assert.False(t, mockT.failed)
	assert.Empty(t, mockT.errors)
}

// TestBatcherTester_AssertBatchSize_Failure verifies failed batch size assertion
func TestBatcherTester_AssertBatchSize_Failure(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	tester.Batch([]tea.Cmd{cmd1})

	mockT := &mockTestingT{}
	tester.AssertBatchSize(mockT, 0, 5)

	assert.True(t, mockT.failed)
	assert.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "batch 0")
	assert.Contains(t, mockT.errors[0], "expected size 5")
	assert.Contains(t, mockT.errors[0], "got 1")
}

// TestBatcherTester_AssertBatchSize_OutOfBounds verifies out of bounds handling
func TestBatcherTester_AssertBatchSize_OutOfBounds(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	mockT := &mockTestingT{}
	tester.AssertBatchSize(mockT, 999, 1)

	assert.True(t, mockT.failed)
	assert.Len(t, mockT.errors, 1)
	assert.Contains(t, mockT.errors[0], "batch index 999 out of bounds")
}

// TestBatcherTester_WithDeduplication verifies deduplication tracking
func TestBatcherTester_WithDeduplication(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	batcher.EnableDeduplication()
	tester := NewBatcherTester(batcher)

	cmd1 := func() tea.Msg { return "cmd1" }
	cmd2 := func() tea.Msg { return "cmd2" }

	// Batch with duplicate commands
	tester.Batch([]tea.Cmd{cmd1, cmd2, cmd1})

	// Original commands tracked (before deduplication)
	assert.Equal(t, 1, tester.GetBatchCount())
	assert.Len(t, tester.GetBatches()[0], 3)
}

// TestBatcherTester_DifferentStrategies verifies tracking with different strategies
func TestBatcherTester_DifferentStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy commands.CoalescingStrategy
	}{
		{"CoalesceAll", commands.CoalesceAll},
		{"CoalesceByType", commands.CoalesceByType},
		{"NoCoalesce", commands.NoCoalesce},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := commands.NewCommandBatcher(tt.strategy)
			tester := NewBatcherTester(batcher)

			cmd1 := func() tea.Msg { return "cmd1" }
			cmd2 := func() tea.Msg { return "cmd2" }

			result := tester.Batch([]tea.Cmd{cmd1, cmd2})

			// All strategies should track correctly
			assert.Equal(t, 1, tester.GetBatchCount())
			assert.Len(t, tester.GetBatches(), 1)
			assert.Len(t, tester.GetBatches()[0], 2)

			// Result depends on strategy
			if tt.strategy == commands.CoalesceAll || tt.strategy == commands.NoCoalesce {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestBatcherTester_IdempotentOperations verifies operations are idempotent
func TestBatcherTester_IdempotentOperations(t *testing.T) {
	batcher := commands.NewCommandBatcher(commands.CoalesceAll)
	tester := NewBatcherTester(batcher)

	// Multiple clears should be safe
	tester.Clear()
	tester.Clear()
	tester.Clear()

	assert.Equal(t, 0, tester.GetBatchCount())

	// GetBatches on empty should be safe
	batches := tester.GetBatches()
	assert.Empty(t, batches)

	// GetBatchSize on empty should return 0
	size := tester.GetBatchSize(0)
	assert.Equal(t, 0, size)
}
