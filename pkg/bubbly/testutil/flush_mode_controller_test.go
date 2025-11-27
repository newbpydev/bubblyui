package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestFlushModeController_NewFlushModeController tests constructor
func TestFlushModeController_NewFlushModeController(t *testing.T) {
	fmc := NewFlushModeController()

	assert.NotNil(t, fmc)
	assert.Equal(t, "sync", fmc.mode) // Default mode should be sync
	assert.Equal(t, 0, fmc.syncCount)
	assert.Equal(t, 0, fmc.asyncCount)
}

// TestFlushModeController_SyncMode tests that sync mode flushes immediately
func TestFlushModeController_SyncMode(t *testing.T) {
	fmc := NewFlushModeController()
	fmc.SetMode("sync")

	// Create a ref and watch it with sync mode
	ref := bubbly.NewRef(0)
	syncCalls := 0

	cleanup := bubbly.Watch(ref, func(newVal, oldVal int) {
		syncCalls++
		fmc.RecordSyncFlush()
	}, bubbly.WithFlush("sync"))
	defer cleanup()

	// Trigger change
	ref.Set(1)

	// Sync mode should execute immediately
	assert.Equal(t, 1, syncCalls, "Sync callback should execute immediately")
	fmc.AssertSyncFlush(t, 1)
	fmc.AssertAsyncFlush(t, 0)
}

// TestFlushModeController_AsyncMode tests that async mode batches updates
func TestFlushModeController_AsyncMode(t *testing.T) {
	fmc := NewFlushModeController()
	fmc.SetMode("post")

	// Create a ref and watch it with post mode
	ref := bubbly.NewRef(0)

	cleanup := bubbly.Watch(ref, func(newVal, oldVal int) {
		fmc.RecordAsyncFlush()
	}, bubbly.WithFlush("post"))
	defer cleanup()

	// Trigger change - this queues the callback
	ref.Set(1)

	// Post mode queues callbacks - need to flush them
	flushedCount := bubbly.FlushWatchers()
	assert.Equal(t, 1, flushedCount, "Should flush 1 watcher")

	// Now the async flush should be recorded
	fmc.AssertAsyncFlush(t, 1)
}

// TestFlushModeController_ModeSwitching tests switching between modes
func TestFlushModeController_ModeSwitching(t *testing.T) {
	fmc := NewFlushModeController()

	// Start with sync
	fmc.SetMode("sync")
	assert.Equal(t, "sync", fmc.GetMode())

	// Switch to post
	fmc.SetMode("post")
	assert.Equal(t, "post", fmc.GetMode())

	// Switch back to sync
	fmc.SetMode("sync")
	assert.Equal(t, "sync", fmc.GetMode())
}

// TestFlushModeController_MultipleFlushes tests tracking multiple flush operations
func TestFlushModeController_MultipleFlushes(t *testing.T) {
	fmc := NewFlushModeController()

	// Record multiple sync flushes
	fmc.RecordSyncFlush()
	fmc.RecordSyncFlush()
	fmc.RecordSyncFlush()

	fmc.AssertSyncFlush(t, 3)
	fmc.AssertAsyncFlush(t, 0)

	// Record multiple async flushes
	fmc.RecordAsyncFlush()
	fmc.RecordAsyncFlush()

	fmc.AssertSyncFlush(t, 3)
	fmc.AssertAsyncFlush(t, 2)
}

// TestFlushModeController_Reset tests resetting counters
func TestFlushModeController_Reset(t *testing.T) {
	fmc := NewFlushModeController()

	// Record some flushes
	fmc.RecordSyncFlush()
	fmc.RecordSyncFlush()
	fmc.RecordAsyncFlush()

	assert.Equal(t, 2, fmc.GetSyncCount())
	assert.Equal(t, 1, fmc.GetAsyncCount())

	// Reset
	fmc.Reset()

	assert.Equal(t, 0, fmc.GetSyncCount())
	assert.Equal(t, 0, fmc.GetAsyncCount())
}

// TestFlushModeController_CombinedModes tests using both sync and async watchers
func TestFlushModeController_CombinedModes(t *testing.T) {
	fmc := NewFlushModeController()

	ref := bubbly.NewRef(0)

	// Watch with sync mode
	cleanupSync := bubbly.Watch(ref, func(newVal, oldVal int) {
		fmc.RecordSyncFlush()
	}, bubbly.WithFlush("sync"))
	defer cleanupSync()

	// Watch with post mode
	cleanupPost := bubbly.Watch(ref, func(newVal, oldVal int) {
		fmc.RecordAsyncFlush()
	}, bubbly.WithFlush("post"))
	defer cleanupPost()

	// Trigger change - sync executes immediately, post queues
	ref.Set(1)

	// Sync should have executed
	fmc.AssertSyncFlush(t, 1)
	fmc.AssertAsyncFlush(t, 0) // Post not flushed yet

	// Flush post watchers
	bubbly.FlushWatchers()
	fmc.AssertAsyncFlush(t, 1)

	// Trigger another change
	ref.Set(2)

	fmc.AssertSyncFlush(t, 2)
	fmc.AssertAsyncFlush(t, 1) // Still 1, not flushed yet

	// Flush again
	bubbly.FlushWatchers()
	fmc.AssertAsyncFlush(t, 2)
}

// TestFlushModeController_AssertionFailures tests assertion error messages
func TestFlushModeController_AssertionFailures(t *testing.T) {
	fmc := NewFlushModeController()

	// Record 2 sync flushes
	fmc.RecordSyncFlush()
	fmc.RecordSyncFlush()

	// Create mock testing.T to capture error (reuses mockTestingT from assertions_state_test.go)
	mockT := &mockTestingT{}

	// This should fail
	fmc.AssertSyncFlush(mockT, 5)

	assert.True(t, mockT.failed, "Assertion should fail")
	assert.Len(t, mockT.errors, 1, "Should have one error")
	assert.Contains(t, mockT.errors[0], "expected 5 sync flushes, got 2")
}
