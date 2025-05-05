package core

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test struct for State
type TestState struct {
	Counter int
	Name    string
	Items   []string
	Data    map[string]int
}

func TestStateBasics(t *testing.T) {
	t.Run("Create and Access", func(t *testing.T) {
		// Create state with initial values
		state := NewState(TestState{
			Counter: 0,
			Name:    "Initial",
			Items:   []string{"one", "two"},
			Data:    map[string]int{"key": 42},
		})

		// Get current state
		value := state.Get()

		// Verify values
		assert.Equal(t, 0, value.Counter)
		assert.Equal(t, "Initial", value.Name)
		assert.Equal(t, []string{"one", "two"}, value.Items)
		assert.Equal(t, map[string]int{"key": 42}, value.Data)
	})

	t.Run("Direct Update", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 0,
			Name:    "Initial",
		})

		// Direct update
		state.Set(TestState{
			Counter: 10,
			Name:    "Updated",
			Items:   []string{"three", "four"},
		})

		// Get updated state
		value := state.Get()

		// Verify values
		assert.Equal(t, 10, value.Counter)
		assert.Equal(t, "Updated", value.Name)
		assert.Equal(t, []string{"three", "four"}, value.Items)
	})

	t.Run("Functional Update", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 5,
			Name:    "Start",
			Items:   []string{"a", "b"},
		})

		// Use Update with a function
		state.Update(func(current TestState) TestState {
			current.Counter += 1
			current.Items = append(current.Items, "c")
			return current
		})

		// Get updated state
		value := state.Get()

		// Verify values
		assert.Equal(t, 6, value.Counter)
		assert.Equal(t, "Start", value.Name)
		assert.Equal(t, []string{"a", "b", "c"}, value.Items)
	})
}

func TestStateBatching(t *testing.T) {
	t.Run("Batch Updates", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 0,
			Name:    "Start",
		})

		// Perform multiple updates in a batch
		state.Batch(func() {
			state.Update(func(s TestState) TestState {
				s.Counter = 1
				return s
			})
			state.Update(func(s TestState) TestState {
				s.Counter = 2
				return s
			})
			state.Update(func(s TestState) TestState {
				s.Counter = 3
				return s
			})
			state.Set(TestState{
				Counter: 4,
				Name:    "End",
			})
		})

		// Only the final update should be applied
		value := state.Get()
		assert.Equal(t, 4, value.Counter)
		assert.Equal(t, "End", value.Name)
	})

	t.Run("Nested Batches", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 0,
			Name:    "Start",
		})

		// Track updates
		updateCount := 0

		// Add change listener
		state.OnChange(func(old, new TestState) {
			updateCount++
		})

		// Perform nested batch updates
		state.Batch(func() {
			state.Update(func(s TestState) TestState {
				s.Counter = 1
				return s
			})

			state.Batch(func() {
				state.Update(func(s TestState) TestState {
					s.Counter = 2
					return s
				})
				state.Update(func(s TestState) TestState {
					s.Counter = 3
					return s
				})
			})

			state.Set(TestState{
				Counter: 4,
				Name:    "End",
			})
		})

		// Verify final state
		value := state.Get()
		assert.Equal(t, 4, value.Counter)
		assert.Equal(t, "End", value.Name)

		// Verify only one update notification was sent
		assert.Equal(t, 1, updateCount)
	})
}

func TestStateHistory(t *testing.T) {
	t.Run("State History Tracking", func(t *testing.T) {
		// Create state with history
		state := NewStateWithHistory(TestState{
			Counter: 0,
			Name:    "Start",
		}, 3) // Only keep the last 3 values

		// Perform multiple updates
		state.Set(TestState{Counter: 1, Name: "Update 1"})
		state.Set(TestState{Counter: 2, Name: "Update 2"})
		state.Set(TestState{Counter: 3, Name: "Update 3"})
		state.Set(TestState{Counter: 4, Name: "Update 4"})

		// Get history
		history := state.GetHistory()

		// Should only have the last 3 states (excluding current)
		assert.Equal(t, 3, len(history))
		assert.Equal(t, 1, history[0].Counter) // Oldest
		assert.Equal(t, 2, history[1].Counter)
		assert.Equal(t, 3, history[2].Counter) // Most recent previous

		// Current state is not in history
		assert.Equal(t, 4, state.Get().Counter)

		// Clear history
		state.ClearHistory()
		assert.Equal(t, 0, len(state.GetHistory()))
	})
}

func TestStateNotifications(t *testing.T) {
	t.Run("Change Notifications", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 0,
			Name:    "Start",
		})

		var oldValue, newValue TestState
		notificationCount := 0

		// Add change listener
		state.OnChange(func(old, new TestState) {
			oldValue = old
			newValue = new
			notificationCount++
		})

		// Update state
		state.Set(TestState{
			Counter: 10,
			Name:    "Updated",
		})

		// Verify notification
		assert.Equal(t, 1, notificationCount)
		assert.Equal(t, 0, oldValue.Counter)
		assert.Equal(t, "Start", oldValue.Name)
		assert.Equal(t, 10, newValue.Counter)
		assert.Equal(t, "Updated", newValue.Name)

		// Remove the listener
		callback := func(old, new TestState) {
			oldValue = old
			newValue = new
			notificationCount++
		}
		state.RemoveOnChange(callback)

		// Reset notification count to verify callback is no longer called
		notificationCount = 0

		// Update again
		state.Set(TestState{
			Counter: 20,
			Name:    "Changed again",
		})

		// Verify callback was not called after removal
		assert.Equal(t, 0, notificationCount, "Callback should not be called after removal")
	})
}

func TestStateConcurrency(t *testing.T) {
	t.Run("Concurrent Updates", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 0,
		})

		// Perform concurrent updates
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				state.Update(func(s TestState) TestState {
					s.Counter++
					return s
				})
			}()
		}
		wg.Wait()

		// Verify counter was incremented exactly 100 times
		assert.Equal(t, 100, state.Get().Counter)
	})

	t.Run("Concurrent Reads", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 42,
			Name:    "Shared",
		})

		// Perform concurrent reads
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				value := state.Get()
				// Just verify we can read safely
				assert.Equal(t, 42, value.Counter)
				assert.Equal(t, "Shared", value.Name)
			}()
		}
		wg.Wait()

		// State should be unchanged
		assert.Equal(t, 42, state.Get().Counter)
		assert.Equal(t, "Shared", state.Get().Name)
	})
}

func TestStateCustomEquality(t *testing.T) {
	t.Run("Custom Equality Function", func(t *testing.T) {
		// Create state with custom equality that only considers the counter
		state := NewStateWithEquals(
			TestState{
				Counter: 10,
				Name:    "Start",
			},
			func(a, b TestState) bool {
				return a.Counter == b.Counter
			},
		)

		changeCount := 0
		state.OnChange(func(old, new TestState) {
			changeCount++
		})

		// Update with same counter but different name (should not trigger change)
		state.Set(TestState{
			Counter: 10, // Same counter
			Name:    "New Name",
		})
		assert.Equal(t, 0, changeCount)

		// Update with different counter (should trigger change)
		state.Set(TestState{
			Counter: 11, // Different counter
			Name:    "New Name",
		})
		assert.Equal(t, 1, changeCount)
	})
}

func TestStateReactivity(t *testing.T) {
	t.Run("Signal Integration", func(t *testing.T) {
		// Create state
		state := NewState(TestState{
			Counter: 0,
			Name:    "Start",
		})

		// Access the underlying signal
		signal := state.GetSignal()

		// Track signal dependencies
		updateCount := 0
		StartTracking()
		_ = signal.Value()
		deps := StopTracking()

		// Register effect that depends on the signal
		RegisterEffect(func() {
			stateValue := signal.Value()
			if stateValue.Counter > 0 {
				updateCount++
			}
		}, deps)

		// Reset counter since RegisterEffect runs once
		updateCount = 0

		// Update state should trigger signal
		state.Set(TestState{
			Counter: 5,
			Name:    "Updated",
		})

		// Wait a bit for async effects
		time.Sleep(10 * time.Millisecond)

		// Effect should have run
		assert.Equal(t, 1, updateCount)
	})
}
