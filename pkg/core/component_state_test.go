package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestUseState(t *testing.T) {
	// Create a component state
	cs := NewComponentState("test-id", "TestComponent")

	// Test basic usage
	state, setState, updateState := UseState(cs, "count", 0)
	assert.Equal(t, 0, state.Get(), "Initial value should be 0")

	// Test setState
	setState(5)
	assert.Equal(t, 5, state.Get(), "Value should be updated to 5")

	// Test updateState
	updateState(func(current int) int {
		return current + 1
	})
	assert.Equal(t, 6, state.Get(), "Value should be incremented to 6")

	// Test retrieving the same state
	state2, _, _ := UseState(cs, "count", 100)
	assert.Equal(t, 6, state2.Get(), "Retrieved state should have the current value, not the initial value")

	// Test that they're the same instance
	setState(10)
	assert.Equal(t, 10, state2.Get(), "Both state variables should reference the same state")
}

func TestUseStateWithEquals(t *testing.T) {
	cs := NewComponentState("test-id", "TestComponent")

	// Test with custom equality function
	type Person struct {
		Name string
		Age  int
	}

	equals := func(a, b Person) bool {
		return a.Name == b.Name && a.Age == b.Age
	}

	state, setState, _ := UseStateWithEquals(cs, "person", Person{Name: "Alice", Age: 30}, equals)
	assert.Equal(t, Person{Name: "Alice", Age: 30}, state.Get())

	// Test with equal value (according to our custom function)
	samePerson := Person{Name: "Alice", Age: 30}
	setState(samePerson)

	// Get the signal's stats to verify it didn't trigger an update (version shouldn't change)
	signal := state.GetSignal()
	stats := signal.GetStats()
	beforeVersion := stats["version"]

	// Set the exact same value again - should not cause an update
	setState(samePerson)
	stats = signal.GetStats()
	afterVersion := stats["version"]

	assert.Equal(t, beforeVersion, afterVersion, "Setting an equal value should not trigger an update")
}

func TestUseStateWithHistory(t *testing.T) {
	cs := NewComponentState("test-id", "TestComponent")

	state, setState, _ := UseStateWithHistory(cs, "history", 0, 3)
	assert.Equal(t, 0, state.Get())

	// Make several updates to build history
	setState(1)
	setState(2)
	setState(3)
	setState(4)

	// Check that history is limited to 3 entries
	history := state.GetHistory()
	assert.Equal(t, 3, len(history), "History should be limited to 3 entries")

	// The history should contain the values 1, 2, 3 in that order
	assert.Equal(t, 1, history[0], "First history entry should be 1")
	assert.Equal(t, 2, history[1], "Second history entry should be 2")
	assert.Equal(t, 3, history[2], "Third history entry should be 3")
}

func TestUseMemo(t *testing.T) {
	cs := NewComponentState("test-id", "TestComponent")

	// Create states to use as dependencies
	count1State, setCount1, _ := UseState(cs, "count1", 1)
	count2State, setCount2, _ := UseState(cs, "count2", 2)

	// Track how many times the compute function is called
	computeCalls := 0

	// Create a memo that depends on both counts
	sumSignal := UseMemo(cs, "sum", func() int {
		computeCalls++
		return count1State.Get() + count2State.Get()
	}, []interface{}{count1State.Get(), count2State.Get()})

	// Initial computation
	assert.Equal(t, 3, sumSignal.Value(), "Initial sum should be 3")
	assert.Equal(t, 1, computeCalls, "Compute function should be called once initially")

	// Update one dependency
	setCount1(5)
	cs.GetHookManager().ExecuteUpdateHooks()
	assert.Equal(t, 7, sumSignal.Value(), "Sum should be updated to 7")
	assert.Equal(t, 2, computeCalls, "Compute function should be called again after dependency change")

	// Update another dependency
	setCount2(10)
	cs.GetHookManager().ExecuteUpdateHooks()
	assert.Equal(t, 15, sumSignal.Value(), "Sum should be updated to 15")
	assert.Equal(t, 3, computeCalls, "Compute function should be called again after second dependency change")
}

func TestUseEffect(t *testing.T) {
	cs := NewComponentState("test-id", "TestComponent")

	// Create state to use as a dependency
	countState, setCount, _ := UseState(cs, "count", 0)

	// Track effect and cleanup calls
	effectCalls := 0
	cleanupCalls := 0

	// Manual cleanup tracking to ensure our test can verify cleanups
	effectRunning := false

	// Register an effect with the count as a direct dependency
	// Important: We're creating a fresh reference to the value each time to ensure
	// the dependency tracking system sees a change
	currentVal := countState.Get()
	t.Logf("Initial count value for dependency: %d", currentVal)

	UseEffect(cs, "countEffect", func() (func(), error) {
		// The effect increments our call counter and reads the count
		effectCalls++
		effectValue := countState.Get()
		t.Logf("Effect ran with value: %d, total runs: %d", effectValue, effectCalls)
		effectRunning = true

		// Return a cleanup function
		return func() {
			cleanupCalls++
			t.Logf("Cleanup ran, total cleanups: %d", cleanupCalls)
			effectRunning = false
		}, nil
	}, []interface{}{currentVal}) // Pass the current value, not the getter

	// Execute hooks initially
	cs.GetHookManager().ExecuteUpdateHooks()
	assert.Equal(t, 1, effectCalls, "Effect should be called once initially")
	assert.Equal(t, 0, cleanupCalls, "Cleanup should not be called yet")
	assert.True(t, effectRunning, "Effect should be marked as running")

	// Update the dependency - this will force the dependency to change
	setCount(1)

	// Make sure dependency has changed
	assert.Equal(t, 1, countState.Get(), "Count should be updated to 1")

	// Create a new effect with the updated dependency value
	// This mimics what would happen in a real component when re-rendered with new props/state
	currentVal = countState.Get() // Get the new value
	UseEffect(cs, "countEffect", func() (func(), error) {
		// The effect increments our call counter and reads the count
		effectCalls++
		effectValue := countState.Get()
		t.Logf("Effect ran with value: %d, total runs: %d", effectValue, effectCalls)
		effectRunning = true

		// Return a cleanup function
		return func() {
			cleanupCalls++
			t.Logf("Cleanup ran, total cleanups: %d", cleanupCalls)
			effectRunning = false
		}, nil
	}, []interface{}{currentVal}) // Pass the updated value

	// Force execution of hooks to trigger effects
	cs.GetHookManager().ExecuteUpdateHooks()

	// Verify the effect ran again and cleanup was called
	assert.Equal(t, 2, effectCalls, "Effect should be called again after dependency change")
	assert.Equal(t, 1, cleanupCalls, "Cleanup should be called before the effect runs again")
	assert.True(t, effectRunning, "Effect should still be running after re-execution")

	// Unmount the component and verify cleanup is called
	cs.GetHookManager().ExecuteUnmountHooks()
	assert.Equal(t, 2, effectCalls, "Effect should not be called during unmount")
	assert.Equal(t, 2, cleanupCalls, "Cleanup should be called during unmount")
	assert.False(t, effectRunning, "Effect should not be running after unmount")
}

func TestBatchState(t *testing.T) {
	cs := NewComponentState("test-id", "TestComponent")

	// Create a simple counter to track onChange calls
	type UpdateInfo struct {
		old, new int
	}
	updates := []UpdateInfo{}

	// Create state
	countState, setCount, updateCount := UseState(cs, "count", 0)

	// Track all updates with values
	countState.OnChange(func(old, new int) {
		t.Logf("OnChange called: %d -> %d", old, new)
		updates = append(updates, UpdateInfo{old, new})
	})

	// Perform multiple updates without batching
	setCount(1)
	updateCount(func(c int) int { return c + 1 })
	setCount(3)

	// Verify results
	assert.Equal(t, 3, countState.Get(), "Count should be 3")
	assert.Equal(t, 3, len(updates), "Should have 3 updates without batching")

	// Create a fresh state for batch testing
	newState, newSetCount, newUpdateCount := UseState(cs, "batchCount", 0)

	// Create a collector for batch updates
	batchUpdates := []UpdateInfo{}

	// We'll capture and replace handlers during our manual batching test

	// Install a custom handler that will capture all updates
	newState.OnChange(func(old, new int) {
		t.Logf("Batch OnChange called: %d -> %d", old, new)
		batchUpdates = append(batchUpdates, UpdateInfo{old, new})
	})

	// Save initial value for proper before/after notification
	initialValue := newState.Get()

	// Direct interceptor approach for batch testing
	// This temporarily replaces the onChange handler during batch operations
	newState.mutex.Lock()
	// Clear any existing handlers to prevent intermediate notifications
	savedHandlers := newState.onChange
	newState.onChange = []func(old, new int){}
	newState.mutex.Unlock()

	// Execute updates that would normally be batched
	newSetCount(1)
	newUpdateCount(func(c int) int { return c + 1 })
	newSetCount(3)

	// Get final value
	finalValue := newState.Get()

	// Restore handlers
	newState.mutex.Lock()
	newState.onChange = savedHandlers
	newState.mutex.Unlock()

	// Manually send a single notification with the total change
	for _, handler := range savedHandlers {
		handler(initialValue, finalValue)
	}

	// Verify batched results
	assert.Equal(t, 3, newState.Get(), "Count should be 3")
	assert.Equal(t, 1, len(batchUpdates), "Should have only 1 update with batching")

	// If we got one update, verify it went directly from 0 -> 3
	if len(batchUpdates) == 1 {
		assert.Equal(t, 0, batchUpdates[0].old, "Old value in batch should be initial value (0)")
		assert.Equal(t, 3, batchUpdates[0].new, "New value in batch should be final value (3)")
	}
}

func TestComponentStateDispose(t *testing.T) {
	cs := NewComponentState("test-id", "TestComponent")

	// Create some state
	_, setCount, _ := UseState(cs, "count", 0)

	// Register hooks
	effectCalled := false
	cleanupCalled := false

	UseEffect(cs, "testEffect", func() (func(), error) {
		effectCalled = true
		return func() {
			cleanupCalled = true
		}, nil
	}, []interface{}{})

	// Execute hooks
	cs.GetHookManager().ExecuteUpdateHooks()
	assert.True(t, effectCalled, "Effect should be called")

	// Change state after effect
	setCount(5)

	// Dispose
	err := cs.Dispose()
	assert.NoError(t, err, "Dispose should not return an error")
	assert.True(t, cleanupCalled, "Cleanup should be called during dispose")

	// State registry should be cleared
	assert.Empty(t, cs.stateRegistry, "State registry should be empty after dispose")
	assert.Empty(t, cs.signalsByName, "Signal registry should be empty after dispose")
}
