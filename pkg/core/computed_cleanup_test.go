package core

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComputedCleanup(t *testing.T) {
	t.Run("Cleanup Execution Timing", func(t *testing.T) {
		// Create a counter to track cleanup execution
		var cleanupExecuted atomic.Int32

		// Create a source signal that will be used in the computed
		source := CreateSignal(10)

		// Add cleanup function that will increment the counter when executed
		computed := CreateComputedWithCleanup(
			func() int {
				return source.Value() * 2
			},
			func() error {
				cleanupExecuted.Add(1)
				return nil
			},
		)

		// Initial value should be computed correctly
		assert.Equal(t, 20, computed.Value(), "Computed should have correct initial value")

		// Cleanup should not have run yet
		assert.Equal(t, int32(0), cleanupExecuted.Load(), "Cleanup should not run until signal is disposed")

		// Simulate signal disposal by removing the effect
		effectID, ok := computed.GetMetadata("effectID")
		if !ok || effectID == nil {
			t.Fatalf("effectID metadata not set on computed signal: got %#v", effectID)
		}
		idStr, ok := effectID.(string)
		if !ok || idStr == "" {
			t.Fatalf("effectID is not a valid string: %#v", effectID)
		}
		RemoveEffect(idStr)

		// Give async operations time to complete
		time.Sleep(10 * time.Millisecond)

		// Cleanup should have executed exactly once
		assert.Equal(t, int32(1), cleanupExecuted.Load(), "Cleanup should execute exactly once on disposal")
	})

	t.Run("Graceful Error Handling During Cleanup", func(t *testing.T) {
		// Create a tracker for error handler execution
		var errorHandled atomic.Bool
		var handledError error
		var errorHandlerMu sync.Mutex

		// Set up error handler
		SetErrorHandler(func(err error) {
			errorHandlerMu.Lock()
			defer errorHandlerMu.Unlock()
			errorHandled.Store(true)
			handledError = err
		})

		// Create a source signal
		source := CreateSignal(5)

		// Create computed with a cleanup function that returns an error
		expectedError := errors.New("expected cleanup error")
		computed := CreateComputedWithCleanup(
			func() int {
				return source.Value() * 3
			},
			func() error {
				return expectedError
			},
		)

		// Initial computation should work
		assert.Equal(t, 15, computed.Value(), "Computed should have correct initial value")

		// Verify no error was handled yet
		assert.False(t, errorHandled.Load(), "Error handler should not be triggered before cleanup")

		// Simulate signal disposal
		effectID, ok := computed.GetMetadata("effectID")
		if !ok || effectID == nil {
			t.Fatalf("effectID metadata not set on computed signal: got %#v", effectID)
		}
		idStr, ok := effectID.(string)
		if !ok || idStr == "" {
			t.Fatalf("effectID is not a valid string: %#v", effectID)
		}
		RemoveEffect(idStr)

		// Give time for async operations
		time.Sleep(10 * time.Millisecond)

		// Verify error was handled properly
		assert.True(t, errorHandled.Load(), "Error handler should be triggered for cleanup errors")

		errorHandlerMu.Lock()
		assert.Equal(t, expectedError, handledError, "The exact error should be passed to the handler")
		errorHandlerMu.Unlock()

		// Reset error handler
		SetErrorHandler(nil)
	})

	t.Run("Complex Resource Cleanup", func(t *testing.T) {
		// Create tracking for multiple cleanup executions
		var parentCleanupCount, childCleanupCount atomic.Int32

		// Create source signals
		source1 := CreateSignal(10)
		source2 := CreateSignal(5)

		// Create a child computed signal
		child := CreateComputedWithCleanup(
			func() int {
				return source1.Value() + source2.Value()
			},
			func() error {
				childCleanupCount.Add(1)
				return nil
			},
		)

		// Create a parent computed signal that depends on the child
		parent := CreateComputedWithCleanup(
			func() int {
				return child.Value() * 2
			},
			func() error {
				parentCleanupCount.Add(1)
				return nil
			},
		)

		// Initial values should be computed correctly
		assert.Equal(t, 15, child.Value(), "Child computed should have correct initial value")
		assert.Equal(t, 30, parent.Value(), "Parent computed should have correct initial value")

		// No cleanups should have run yet
		assert.Equal(t, int32(0), childCleanupCount.Load(), "Child cleanup should not run until disposal")
		assert.Equal(t, int32(0), parentCleanupCount.Load(), "Parent cleanup should not run until disposal")

		// Update a source signal and verify the computed values update
		source1.Set(20)
		assert.Equal(t, 25, child.Value(), "Child computed should update when source changes")
		assert.Equal(t, 50, parent.Value(), "Parent computed should update when child changes")

		// Cleanups still should not have run
		assert.Equal(t, int32(0), childCleanupCount.Load(), "Child cleanup should not run on updates")
		assert.Equal(t, int32(0), parentCleanupCount.Load(), "Parent cleanup should not run on updates")

		// Dispose the parent effect first
		parentEffectID, ok := parent.GetMetadata("effectID")
		if !ok || parentEffectID == nil {
			t.Fatalf("parent effectID metadata not set on computed signal: got %#v", parentEffectID)
		}
		parentIDStr, ok := parentEffectID.(string)
		if !ok || parentIDStr == "" {
			t.Fatalf("parent effectID is not a valid string: %#v", parentEffectID)
		}
		RemoveEffect(parentIDStr)

		// Give time for async operations
		time.Sleep(10 * time.Millisecond)

		// Only parent cleanup should have run
		assert.Equal(t, int32(0), childCleanupCount.Load(), "Child cleanup should not run when only parent is disposed")
		assert.Equal(t, int32(1), parentCleanupCount.Load(), "Parent cleanup should run when parent is disposed")

		// Dispose the child effect
		childEffectID, ok := child.GetMetadata("effectID")
		if !ok || childEffectID == nil {
			t.Fatalf("child effectID metadata not set on computed signal: got %#v", childEffectID)
		}
		childIDStr, ok := childEffectID.(string)
		if !ok || childIDStr == "" {
			t.Fatalf("child effectID is not a valid string: %#v", childEffectID)
		}
		RemoveEffect(childIDStr)

		// Give time for async operations
		time.Sleep(10 * time.Millisecond)

		// Both cleanups should have run
		assert.Equal(t, int32(1), childCleanupCount.Load(), "Child cleanup should run when child is disposed")
		assert.Equal(t, int32(1), parentCleanupCount.Load(), "Parent cleanup should still have run only once")
	})
}
