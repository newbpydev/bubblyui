package testutil

// FlushModeController provides utilities for testing flush timing control in reactive updates.
// It helps verify that watchers execute at the correct time based on their flush mode configuration.
//
// BubblyUI's Watch system supports different flush modes:
//   - "sync": Execute callback immediately when value changes (default)
//   - "post": Defer callback execution to next tick (placeholder - currently behaves as sync)
//
// This tester allows you to:
//   - Track sync vs async flush counts
//   - Verify flush mode behavior
//   - Test mode switching
//   - Assert on flush timing
//
// Example:
//
//	fmc := NewFlushModeController()
//	ref := bubbly.NewRef(0)
//
//	// Watch with sync mode
//	cleanup := bubbly.Watch(ref, func(newVal, oldVal int) {
//	    fmc.RecordSyncFlush()
//	}, bubbly.WithFlush("sync"))
//	defer cleanup()
//
//	ref.Set(1)
//	fmc.AssertSyncFlush(t, 1) // Verify sync flush executed
//
// Thread Safety:
//
// FlushModeController is not thread-safe. It should only be used from a single test goroutine.
type FlushModeController struct {
	mode       string // Current flush mode ("sync" or "post")
	syncCount  int    // Number of sync flushes recorded
	asyncCount int    // Number of async flushes recorded
}

// NewFlushModeController creates a new FlushModeController for testing flush timing.
//
// The controller starts in "sync" mode by default, matching the default behavior
// of the Watch system.
//
// Returns:
//   - *FlushModeController: A new controller instance
//
// Example:
//
//	fmc := NewFlushModeController()
//	fmc.SetMode("post") // Switch to async mode
func NewFlushModeController() *FlushModeController {
	return &FlushModeController{
		mode:       "sync", // Default to sync mode
		syncCount:  0,
		asyncCount: 0,
	}
}

// SetMode sets the current flush mode for testing.
//
// Valid modes:
//   - "sync": Synchronous flush (immediate execution)
//   - "post": Asynchronous flush (deferred execution)
//
// Note: This doesn't affect the actual Watch behavior, it's just for tracking
// which mode you're testing.
//
// Parameters:
//   - mode: The flush mode to set ("sync" or "post")
//
// Example:
//
//	fmc.SetMode("sync")
//	// Test sync mode behavior...
//
//	fmc.SetMode("post")
//	// Test async mode behavior...
func (fmc *FlushModeController) SetMode(mode string) {
	fmc.mode = mode
}

// GetMode returns the current flush mode.
//
// Returns:
//   - string: The current mode ("sync" or "post")
//
// Example:
//
//	mode := fmc.GetMode()
//	assert.Equal(t, "sync", mode)
func (fmc *FlushModeController) GetMode() string {
	return fmc.mode
}

// RecordSyncFlush records that a synchronous flush occurred.
// Call this from within your Watch callback when testing sync mode.
//
// Example:
//
//	cleanup := bubbly.Watch(ref, func(newVal, oldVal int) {
//	    fmc.RecordSyncFlush()
//	}, bubbly.WithFlush("sync"))
func (fmc *FlushModeController) RecordSyncFlush() {
	fmc.syncCount++
}

// RecordAsyncFlush records that an asynchronous flush occurred.
// Call this from within your Watch callback when testing post mode.
//
// Example:
//
//	cleanup := bubbly.Watch(ref, func(newVal, oldVal int) {
//	    fmc.RecordAsyncFlush()
//	}, bubbly.WithFlush("post"))
func (fmc *FlushModeController) RecordAsyncFlush() {
	fmc.asyncCount++
}

// GetSyncCount returns the number of sync flushes recorded.
//
// Returns:
//   - int: The sync flush count
//
// Example:
//
//	count := fmc.GetSyncCount()
//	assert.Equal(t, 3, count)
func (fmc *FlushModeController) GetSyncCount() int {
	return fmc.syncCount
}

// GetAsyncCount returns the number of async flushes recorded.
//
// Returns:
//   - int: The async flush count
//
// Example:
//
//	count := fmc.GetAsyncCount()
//	assert.Equal(t, 2, count)
func (fmc *FlushModeController) GetAsyncCount() int {
	return fmc.asyncCount
}

// Reset resets all flush counters to zero.
// Useful for testing multiple scenarios in the same test.
//
// Example:
//
//	fmc.RecordSyncFlush()
//	fmc.AssertSyncFlush(t, 1)
//
//	fmc.Reset()
//	fmc.AssertSyncFlush(t, 0) // Counters reset
func (fmc *FlushModeController) Reset() {
	fmc.syncCount = 0
	fmc.asyncCount = 0
}

// AssertSyncFlush asserts that the expected number of sync flushes occurred.
//
// Parameters:
//   - t: The testingT instance (compatible with *testing.T)
//   - expected: The expected number of sync flushes
//
// Example:
//
//	ref.Set(1)
//	fmc.AssertSyncFlush(t, 1) // Verify one sync flush
//
//	ref.Set(2)
//	fmc.AssertSyncFlush(t, 2) // Verify two total sync flushes
func (fmc *FlushModeController) AssertSyncFlush(t testingT, expected int) {
	t.Helper()
	if fmc.syncCount != expected {
		t.Errorf("expected %d sync flushes, got %d", expected, fmc.syncCount)
	}
}

// AssertAsyncFlush asserts that the expected number of async flushes occurred.
//
// Parameters:
//   - t: The testingT instance (compatible with *testing.T)
//   - expected: The expected number of async flushes
//
// Example:
//
//	ref.Set(1)
//	fmc.AssertAsyncFlush(t, 1) // Verify one async flush
//
//	ref.Set(2)
//	fmc.AssertAsyncFlush(t, 2) // Verify two total async flushes
func (fmc *FlushModeController) AssertAsyncFlush(t testingT, expected int) {
	t.Helper()
	if fmc.asyncCount != expected {
		t.Errorf("expected %d async flushes, got %d", expected, fmc.asyncCount)
	}
}
