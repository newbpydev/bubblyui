package core

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock effect implementation for testing
func createMockEffect(fn func()) string {
	// Track initial dependencies by running the function once with tracking
	StartTracking()
	fn()
	deps := StopTracking()

	// Create a unique ID for this effect
	effectID := fmt.Sprintf("mock_effect_%d", time.Now().UnixNano())

	// Create the effect
	effect := &Effect{
		fn:        fn,
		deps:      deps,
		debugInfo: "mock_effect",
	}

	// Register it globally
	globalMutex.Lock()
	effectsRegistry[effectID] = effect
	globalMutex.Unlock()

	// Register each dependency properly
	for _, depID := range deps {
		registerDependencyWithSignal(depID, effectID, effect)
	}

	return effectID
}

// Mock dependency functionality for testing
func createMockEffectWithDeps(fn func(), deps []string) string {
	// Create a unique ID for this effect
	effectID := fmt.Sprintf("mock_effect_deps_%d", time.Now().UnixNano())

	// Create the effect
	effect := &Effect{
		fn:        fn,
		deps:      deps,
		debugInfo: "mock_effect_with_deps",
	}

	// Register it globally
	globalMutex.Lock()
	effectsRegistry[effectID] = effect
	globalMutex.Unlock()

	// Register each dependency properly
	for _, depID := range deps {
		registerDependencyWithSignal(depID, effectID, effect)
	}

	// Run once initially
	fn()

	return effectID
}

func TestEffectDependencyTracking(t *testing.T) {
	t.Run("Automatic Dependency Detection", func(t *testing.T) {
		// Create signals that will be dependencies
		count := CreateSignal(5)
		multiplier := CreateSignal(2)

		// Track effect executions
		var effectExecutionCount atomic.Int32

		// Create and run effect with automatic dependency tracking
		effectID := createMockEffect(func() {
			// Access signals to create automatic dependencies
			_ = count.Value() * multiplier.Value()
			effectExecutionCount.Add(1)
		})

		// Should have executed once during registration
		assert.Equal(t, int32(1), effectExecutionCount.Load(), "Effect should execute once when created")

		// Modify a dependency
		count.Set(10)
		time.Sleep(10 * time.Millisecond) // Give time for effect to run
		assert.Equal(t, int32(2), effectExecutionCount.Load(), "Effect should execute when dependency changes")

		// Modify another dependency
		multiplier.Set(3)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(3), effectExecutionCount.Load(), "Effect should execute when any dependency changes")

		// Clean up
		RemoveEffect(effectID)
	})

	t.Run("Explicit Dependency List", func(t *testing.T) {
		// Create signals to be used as dependencies
		count := CreateSignal(1)
		multiplier := CreateSignal(2)
		unused := CreateSignal(3) // This signal is read but not in the explicit deps list

		// Track effect executions
		var effectExecutionCount atomic.Int32

		// Get the actual signal IDs
		countID := count.id
		multiplierID := multiplier.id

		// Create effect with explicit dependencies
		effectID := createMockEffectWithDeps(
			func() {
				// Read all signals - but only those in the deps list should trigger updates
				_ = count.Value()*multiplier.Value() + unused.Value()
				effectExecutionCount.Add(1)
			},
			[]string{countID, multiplierID}, // Explicit signal IDs
		)

		// Should have executed once during registration
		assert.Equal(t, int32(1), effectExecutionCount.Load(), "Effect should execute once when created")

		// Signal in dependency list - should trigger update
		count.Set(10)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(2), effectExecutionCount.Load(), "Effect should execute when explicit dependency changes")

		// Signal not in dependency list - should NOT trigger update
		unused.Set(30)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(2), effectExecutionCount.Load(), "Effect should NOT execute when non-listed dependency changes")

		// Clean up
		RemoveEffect(effectID)
	})

	t.Run("Dependency Change Detection", func(t *testing.T) {
		// Create signals to be tracked
		counter := CreateSignal(0)

		// Create a tracking variable
		var lastValue atomic.Int32
		var effectRunCount atomic.Int32

		// Register an effect with automatic dependency tracking
		effectID := createMockEffect(func() {
			value := counter.Value()
			lastValue.Store(int32(value))
			effectRunCount.Add(1)
		})

		// Effect should run once on creation
		assert.Equal(t, int32(1), effectRunCount.Load(), "Effect should run once on creation")
		assert.Equal(t, int32(0), lastValue.Load(), "Initial value should be 0")

		// Change value
		counter.Set(1)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(2), effectRunCount.Load(), "Effect should run again when value changes")
		assert.Equal(t, int32(1), lastValue.Load(), "Last value should be updated to 1")

		// Set to same value - should NOT trigger
		counter.Set(1) // Same value shouldn't update signal or trigger effect
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(2), effectRunCount.Load(), "Effect should not run when value does not change")

		// Clean up
		RemoveEffect(effectID)
	})
}
