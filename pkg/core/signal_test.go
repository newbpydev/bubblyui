package core

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignalBasics(t *testing.T) {
	t.Run("Create and Access", func(t *testing.T) {
		// Create a signal with an initial value
		signal := NewSignal(42)

		// Check initial value
		assert.Equal(t, 42, signal.Value())

		// Set a new value
		signal.Set(100)

		// Check updated value
		assert.Equal(t, 100, signal.Value())
	})

	t.Run("Type Safety", func(t *testing.T) {
		// Test with different types
		intSignal := NewSignal(42)
		assert.Equal(t, 42, intSignal.Value())

		strSignal := NewSignal("hello")
		assert.Equal(t, "hello", strSignal.Value())

		boolSignal := NewSignal(true)
		assert.Equal(t, true, boolSignal.Value())

		// Test with struct
		type Person struct {
			Name string
			Age  int
		}

		person := Person{Name: "Alice", Age: 30}
		personSignal := NewSignal(person)
		assert.Equal(t, person, personSignal.Value())
	})
}

func TestSignalDependencyTracking(t *testing.T) {
// debugMode = true // Ensure debug logs are printed
	t.Run("Simple Dependency", func(t *testing.T) {
		countSignal := NewSignal(0)
		doubledSignal := NewSignal(0)
		fmt.Printf("[TEST DEBUG] countSignal addr: %p, id: %s\n", countSignal, countSignal.id)
		fmt.Printf("[TEST DEBUG] doubledSignal addr: %p, id: %s\n", doubledSignal, doubledSignal.id)

		// Track updates
		updateCount := 0

		// Create a computed value that depends on countSignal
		StartTracking()
		_ = countSignal.Value() // Record dependency on countSignal
		deps := StopTracking()
		fmt.Printf("[TEST DEBUG] deps after StopTracking: %v\n", deps)

		// Register dependency
		RegisterEffectWithoutInitialRun(func() {
			doubledSignal.Set(countSignal.Value() * 2)
			updateCount++
		}, deps, "test_simple_dependency_effect")

		// Initial computed value - effect should have run once
		assert.Equal(t, 0, doubledSignal.Value())
		assert.Equal(t, 1, updateCount, "Effect should have run once during registration")

		// Update the source signal
		countSignal.Set(5)

		// The effect should have run again, updating doubledSignal
		assert.Equal(t, 10, doubledSignal.Value())
		assert.Equal(t, 2, updateCount, "Effect should have run when signal changed")

		// Another update to countSignal
		countSignal.Set(7)
		assert.Equal(t, 14, doubledSignal.Value())
		assert.Equal(t, 3, updateCount, "Effect should have run again")
	})

	t.Run("Multiple Dependencies", func(t *testing.T) {
		// Create signals
		firstSignal := NewSignal(5)
		secondSignal := NewSignal(10)
		resultSignal := NewSignal(0)

		// Track updates
		updateCount := 0

		// Create a computed value that depends on multiple signals
		StartTracking()
		// Access both signals to create dependencies
		_ = firstSignal.Value()
		_ = secondSignal.Value()
		deps := StopTracking()

		// Register effect
		RegisterEffectWithoutInitialRun(func() {
			resultSignal.Set(firstSignal.Value() + secondSignal.Value())
			updateCount++
		}, deps, "test_multiple_dependencies_effect")

		// Initial values after first run
		assert.Equal(t, 15, resultSignal.Value())
		assert.Equal(t, 1, updateCount, "Effect should have run once during registration")

		// Update first signal
		firstSignal.Set(7)
		assert.Equal(t, 17, resultSignal.Value())
		assert.Equal(t, 2, updateCount, "Effect should run when first signal changes")

		// Update second signal
		secondSignal.Set(12)
		assert.Equal(t, 19, resultSignal.Value())
		assert.Equal(t, 3, updateCount, "Effect should run when second signal changes")
	})
}

func TestSignalConcurrency(t *testing.T) {
	t.Run("Concurrent Reads", func(t *testing.T) {
		signal := NewSignal(42)

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Just read the value
				_ = signal.Value()
			}()
		}
		wg.Wait()

		// If we got here without panics, test passes
		assert.Equal(t, 42, signal.Value())
	})

	t.Run("Concurrent Writes", func(t *testing.T) {
		// Create a signal specifically for atomic integer operations
		signal := NewSignal(0)

		// Create a mutex to ensure atomic operations
		var mu sync.Mutex

		// Have multiple goroutines increment the counter
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Use mutex to ensure atomic read-modify-write
				mu.Lock()
				current := signal.Value()
				signal.Set(current + 1)
				mu.Unlock()
			}()
		}
		wg.Wait()

		// Final value should be 100
		assert.Equal(t, 100, signal.Value(), "Expected all 100 increments to be applied")
	})
}

func TestBatchedUpdates(t *testing.T) {
// debugMode = true // Ensure debug logs are printed
	t.Run("Batching Prevents Intermediate Updates", func(t *testing.T) {
		signal := NewSignal(0)
		derived := NewSignal(0)
		fmt.Printf("[TEST DEBUG] signal addr: %p, id: %s\n", signal, signal.id)
		fmt.Printf("[TEST DEBUG] derived addr: %p, id: %s\n", derived, derived.id)

		updateCount := 0

		// Create effect that tracks signal changes
		StartTracking()
		_ = signal.Value() // Create dependency on signal
		deps := StopTracking()
		fmt.Printf("[TEST DEBUG] deps after StopTracking: %v\n", deps)

		// Create effect that tracks signal changes
		StartTracking()
		_ = signal.Value() // Create dependency on signal
		deps = StopTracking()

		RegisterEffect(func() {
			derived.Set(signal.Value() * 2)
			updateCount++
		}, deps)

		// Reset counter since RegisterEffect runs the effect once
		updateCount = 0

		// Perform multiple updates in a batch
		Batch(func() {
			signal.Set(1)
			signal.Set(2)
			signal.Set(3)
			signal.Set(4)
			signal.Set(5)
		})

		// Should only have one effect execution after the batch, not five
		assert.Equal(t, 1, updateCount, "Should have exactly one update after batch")
		assert.Equal(t, 10, derived.Value(), "Derived value should be 10 (5 * 2)")
	})

	t.Run("Nested Batches", func(t *testing.T) {
		signal := NewSignal(0)
		derived := NewSignal(0)
		updateCount := 0

		// Track dependency
		StartTracking()
		_ = signal.Value() // Create dependency on signal
		deps := StopTracking()

		RegisterEffect(func() {
			derived.Set(signal.Value() * 2)
			updateCount++
		}, deps)

		// Reset counter since RegisterEffect runs the effect once
		updateCount = 0

		// Nested batches
		Batch(func() {
			signal.Set(1)

			Batch(func() {
				signal.Set(2)
				signal.Set(3)
			})

			signal.Set(4)
		})

		// Only one update should occur, with the final value
		assert.Equal(t, 1, updateCount, "Should have exactly one update after nested batches")
		assert.Equal(t, 8, derived.Value(), "Derived value should be 8 (4 * 2)")
	})
}

func TestSignalEqualityComparison(t *testing.T) {
	t.Run("Custom Equality Function", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		// Create signal with custom equality function
		personSignal := NewSignalWithEquals(
			Person{Name: "Alice", Age: 30},
			func(a, b Person) bool {
				return a.Name == b.Name // Only compare names
			},
		)

		updateCount := 0
		// Track effect
		StartTracking()
		_ = personSignal.Value()
		deps := StopTracking()

		RegisterEffect(func() {
			_ = personSignal.Value()
			updateCount++
		}, deps)

		// Reset update count since RegisterEffect will run once
		updateCount = 0

		// Update age only - shouldn't trigger effect due to custom equality
		personSignal.Set(Person{Name: "Alice", Age: 31})
		assert.Equal(t, 0, updateCount, "Should not update when only age changes")

		// Update name - should trigger effect
		personSignal.Set(Person{Name: "Bob", Age: 31})
		assert.Equal(t, 1, updateCount, "Should update when name changes")
	})
}

func TestSignalAsync(t *testing.T) {
	t.Run("Async Effects", func(t *testing.T) {
		signal := NewSignal(0)
		resultSignal := NewSignal(0)

		// Create async effect
		StartTracking()
		_ = signal.Value()
		deps := StopTracking()

		done := make(chan bool)

		RegisterAsyncEffect(func() {
			// Simulate async operation
			go func() {
				time.Sleep(10 * time.Millisecond)
				resultSignal.Set(signal.Value() * 10)
				done <- true
			}()
		}, deps)

		// Trigger the effect
		signal.Set(5)

		// Wait for async effect to complete
		<-done

		// Check result
		assert.Equal(t, 50, resultSignal.Value())
	})
}

func TestSignalErrorHandling(t *testing.T) {
	t.Run("Effect Error Recovery", func(t *testing.T) {
		signal := NewSignal(0)
		errorCount := 0

		// Set up error handler
		SetErrorHandler(func(err error) {
			assert.Error(t, err)
			errorCount++
		})

		// Create effect that will fail when signal is negative
		StartTracking()
		_ = signal.Value()
		deps := StopTracking()

		RegisterEffect(func() {
			if signal.Value() < 0 {
				panic("negative value not allowed")
			}
		}, deps)

		// Trigger error by setting negative value
		signal.Set(-5)

		// Error should have been caught and handled
		assert.Equal(t, 1, errorCount)

		// Reset to valid value
		signal.Set(10)

		// Error handler shouldn't be called again
		assert.Equal(t, 1, errorCount)
	})
}
