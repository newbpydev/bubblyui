package testutil

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHarness_Creation tests that a harness can be created successfully
func TestNewHarness_Creation(t *testing.T) {
	harness := NewHarness(t)

	require.NotNil(t, harness, "harness should not be nil")
	assert.NotNil(t, harness.t, "harness.t should not be nil")
	assert.NotNil(t, harness.refs, "harness.refs should be initialized")
	assert.NotNil(t, harness.events, "harness.events should be initialized")
	assert.NotNil(t, harness.cleanup, "harness.cleanup should be initialized")
}

// TestNewHarness_WithOptions tests that options are applied correctly
func TestNewHarness_WithOptions(t *testing.T) {
	optionCalled := false
	testOption := func(h *TestHarness) {
		optionCalled = true
	}

	harness := NewHarness(t, testOption)

	require.NotNil(t, harness)
	assert.True(t, optionCalled, "option function should have been called")
}

// TestHarness_CleanupExecution tests that cleanup functions execute in LIFO order
func TestHarness_CleanupExecution(t *testing.T) {
	harness := NewHarness(t)

	// Track execution order
	var executionOrder []int
	var mu sync.Mutex

	// Register cleanup functions
	harness.RegisterCleanup(func() {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
	})

	harness.RegisterCleanup(func() {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
	})

	harness.RegisterCleanup(func() {
		mu.Lock()
		executionOrder = append(executionOrder, 3)
		mu.Unlock()
	})

	// Execute cleanup
	harness.Cleanup()

	// Verify LIFO order (last registered runs first)
	assert.Equal(t, []int{3, 2, 1}, executionOrder, "cleanup should execute in LIFO order")
}

// TestHarness_CleanupClearsSlice tests that cleanup slice is cleared after execution
func TestHarness_CleanupClearsSlice(t *testing.T) {
	harness := NewHarness(t)

	// Register some cleanup functions (harness already has 1 for hook cleanup)
	harness.RegisterCleanup(func() {})
	harness.RegisterCleanup(func() {})

	// Verify cleanup functions registered (3 total: 1 for hook + 2 registered)
	assert.Len(t, harness.cleanup, 3, "should have 3 cleanup functions")

	// Execute cleanup
	harness.Cleanup()

	// Verify cleanup slice is cleared
	assert.Len(t, harness.cleanup, 0, "cleanup slice should be empty after execution")
}

// TestHarness_CleanupThreadSafe tests concurrent cleanup registration
func TestHarness_CleanupThreadSafe(t *testing.T) {
	harness := NewHarness(t)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Register cleanup functions concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			harness.RegisterCleanup(func() {})
		}()
	}

	wg.Wait()

	// Verify all cleanup functions were registered (10 + 1 for hook cleanup)
	assert.Len(t, harness.cleanup, numGoroutines+1, "all cleanup functions should be registered")
}

// TestHarness_CleanupIdempotent tests that calling Cleanup multiple times is safe
func TestHarness_CleanupIdempotent(t *testing.T) {
	harness := NewHarness(t)

	executionCount := 0
	harness.RegisterCleanup(func() {
		executionCount++
	})

	// Call cleanup multiple times
	harness.Cleanup()
	harness.Cleanup()
	harness.Cleanup()

	// Verify cleanup only executed once
	assert.Equal(t, 1, executionCount, "cleanup should only execute once")
}

// TestHarness_CleanupWithPanic tests that cleanup continues even if a function panics
func TestHarness_CleanupWithPanic(t *testing.T) {
	harness := NewHarness(t)

	executed := []int{}
	var mu sync.Mutex

	// Register cleanup functions, one will panic
	harness.RegisterCleanup(func() {
		mu.Lock()
		executed = append(executed, 1)
		mu.Unlock()
	})

	harness.RegisterCleanup(func() {
		panic("test panic")
	})

	harness.RegisterCleanup(func() {
		mu.Lock()
		executed = append(executed, 3)
		mu.Unlock()
	})

	// Cleanup should not panic
	assert.NotPanics(t, func() {
		harness.Cleanup()
	}, "cleanup should handle panics gracefully")

	// Verify other cleanup functions still executed
	// Note: Order might vary due to panic recovery, but both should execute
	assert.Contains(t, executed, 1, "cleanup 1 should execute")
	assert.Contains(t, executed, 3, "cleanup 3 should execute")
}

// TestHarness_AutomaticCleanup tests that t.Cleanup is registered
func TestHarness_AutomaticCleanup(t *testing.T) {
	// This test verifies that NewHarness registers cleanup with t.Cleanup()
	// We can't directly test t.Cleanup() registration, but we can verify
	// the harness is set up correctly for automatic cleanup

	harness := NewHarness(t)

	cleanupCalled := false
	harness.RegisterCleanup(func() {
		cleanupCalled = true
	})

	// Manually call cleanup to simulate what t.Cleanup would do
	harness.Cleanup()

	assert.True(t, cleanupCalled, "cleanup should be called")
}

// TestTestHook_LifecycleMethods tests that all lifecycle methods can be called without panicking
func TestTestHook_LifecycleMethods(t *testing.T) {
	harness := NewHarness(t)
	
	// Create a test hook to verify all lifecycle methods work
	hook := &testHook{
		tracker: harness.events,
		harness: harness,
	}
	
	// Test all lifecycle methods - they should not panic
	assert.NotPanics(t, func() {
		hook.OnComponentMount("comp1", "TestComponent")
	}, "OnComponentMount should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnComponentUpdate("comp1", "test message")
	}, "OnComponentUpdate should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnComponentUnmount("comp1")
	}, "OnComponentUnmount should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnRefChange("comp1", "old", "new")
	}, "OnRefChange should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnRefExposed("comp1", "ref1", "ref1")
	}, "OnRefExposed should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnRenderComplete("comp1", 100*time.Millisecond)
	}, "OnRenderComplete should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnComputedChange("comp1", "old", "new")
	}, "OnComputedChange should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnWatchCallback("watch1", "old", "new")
	}, "OnWatchCallback should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnEffectRun("effect1")
	}, "OnEffectRun should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnChildAdded("parent1", "child1")
	}, "OnChildAdded should not panic")
	
	assert.NotPanics(t, func() {
		hook.OnChildRemoved("parent1", "child1")
	}, "OnChildRemoved should not panic")
}

// TestTestHook_OnEvent tests that OnEvent properly tracks events
func TestTestHook_OnEvent(t *testing.T) {
	harness := NewHarness(t)
	
	hook := &testHook{
		tracker: harness.events,
		harness: harness,
	}
	
	// Call OnEvent with test data
	hook.OnEvent("comp1", "testEvent", "testData")
	
	// Verify event was tracked
	events := harness.events.GetEvents("testEvent")
	assert.Len(t, events, 1, "should have tracked one event")
	
	event := events[0]
	assert.Equal(t, "testEvent", event.Name, "event name should match")
	assert.Equal(t, "comp1", event.Source, "component source should match")
	assert.Equal(t, "testData", event.Payload, "event payload should match")
}
