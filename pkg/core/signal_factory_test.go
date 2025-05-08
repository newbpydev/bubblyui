package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSignal(t *testing.T) {
	t.Run("Basic Signal Creation", func(t *testing.T) {
		// We'll use t.Log for better debugging
		t.Log("Creating signal with factory function...")

		// Create a signal with the factory function
		signal := CreateSignal[int](42)

		t.Log("Signal created, checking initial value...")
		// Check that it has the correct initial value
		assert.Equal(t, 42, signal.Value(), "Initial value should be 42")

		t.Log("Setting new value to signal...")
		// Check that it updates properly
		signal.Set(100)
		assert.Equal(t, 100, signal.Value(), "Updated value should be 100")

		t.Log("Checking if metadata field is initialized...")
		// Check that the metadata was initialized
		if signal.metadata == nil {
			t.Fatal("Signal metadata is nil")
		}

		t.Log("Checking for source file metadata...")
		// Verify that source file information was captured
		sourceFile, sourceExists := signal.GetMetadata("sourceFile")
		if !sourceExists {
			t.Error("Source file metadata is missing")
		} else {
			t.Logf("Source file: %v", sourceFile)
		}

		t.Log("Checking for source line metadata...")
		// Verify the source line was recorded
		sourceLine, lineExists := signal.GetMetadata("sourceLine")
		if !lineExists {
			t.Error("Source line metadata is missing")
		} else {
			t.Logf("Source line: %v", sourceLine)
		}

		// Dump all metadata for debugging
		t.Log("All metadata:")
		for k, v := range signal.metadata {
			t.Logf("  %s: %v", k, v)
		}
	})

	t.Run("Signal with Custom Equality", func(t *testing.T) {
		// Custom equality function that considers even numbers equal
		equalityFn := func(a, b int) bool {
			return a%2 == 0 && b%2 == 0
		}

		// Create a signal with custom equality
		signal := CreateSignal[int](2, SignalOptions{
			Equals: func(a, b any) bool {
				return equalityFn(a.(int), b.(int))
			},
			DebugName: "EvenNumberSignal",
		})

		// Setting to another even number should not trigger update
		signal.Set(4)
		assert.Equal(t, 2, signal.Value(), "Value should not change when setting to a 'equal' value")

		// Setting to an odd number should update
		signal.Set(3)
		assert.Equal(t, 3, signal.Value(), "Value should change when setting to a non-equal value")

		// Check debug name in metadata
		debugName, exists := signal.GetMetadata("debugName")
		assert.True(t, exists, "Debug name metadata should be present")
		assert.Equal(t, "EvenNumberSignal", debugName)
	})
}

func TestDepsEqual(t *testing.T) {
	t.Run("Equal slices same order", func(t *testing.T) {
		deps1 := []string{"a", "b", "c"}
		deps2 := []string{"a", "b", "c"}
		assert.True(t, depsEqual(deps1, deps2), "Should be equal when contents and order match")
	})
	t.Run("Equal slices different order", func(t *testing.T) {
		deps1 := []string{"x", "y", "z"}
		deps2 := []string{"z", "x", "y"}
		assert.True(t, depsEqual(deps1, deps2), "Should be equal when contents match but order differs")
	})
	t.Run("Unequal slices different lengths", func(t *testing.T) {
		deps1 := []string{"a", "b"}
		deps2 := []string{"a", "b", "c"}
		assert.False(t, depsEqual(deps1, deps2), "Should not be equal when lengths differ")
	})
	t.Run("Unequal slices same length, different contents", func(t *testing.T) {
		deps1 := []string{"a", "b", "c"}
		deps2 := []string{"a", "b", "d"}
		assert.False(t, depsEqual(deps1, deps2), "Should not be equal when contents differ")
	})
	t.Run("Both empty", func(t *testing.T) {
		deps1 := []string{}
		deps2 := []string{}
		assert.True(t, depsEqual(deps1, deps2), "Empty slices should be equal")
	})
	t.Run("Nil vs empty", func(t *testing.T) {
		var deps1 []string = nil
		deps2 := []string{}
		assert.True(t, depsEqual(deps1, deps2), "Nil and empty slices should be considered equal")
	})
}


func TestCreateComputed(t *testing.T) {
	t.Run("Basic Computed Signal", func(t *testing.T) {
		// Create source signals
		count := CreateSignal[int](5)
		multiplier := CreateSignal[int](2)

		// Create a computed signal that depends on both
		computed := CreateComputed[int](func() int {
			return count.Value() * multiplier.Value()
		})

		// Check initial computed value
		assert.Equal(t, 10, computed.Value(), "Initial computed value should be 10 (5*2)")

		// Update source signals and check that computed value updates
		count.Set(10)
		assert.Equal(t, 20, computed.Value(), "Computed value should update when source changes")

		multiplier.Set(3)
		assert.Equal(t, 30, computed.Value(), "Computed value should update when any dependency changes")
	})

	t.Run("Computed with Custom Equality", func(t *testing.T) {
		// Source signals
		a := CreateSignal[int](5)
		b := CreateSignal[int](5)

		// Counter to track how many times the compute function runs
		computeCounter := 0

		// Debug: Store current counter at key test points to see when it increments
		var beforeCreate, afterCreate, afterSetSame, afterAccessValue, afterSetDifferent int
		beforeCreate = computeCounter
		t.Logf("Before create: computeCounter = %d", beforeCreate)

		// Create a computed signal that checks if values are equal
		// but only considers the result significant if it changes from true to false or vice versa
		computed := CreateComputed[bool](
			func() bool {
				t.Logf("** Running compute function, counter=%d **", computeCounter)
				computeCounter++
				return a.Value() == b.Value()
			},
			SignalOptions{
				DebugName: "EqualityComputed",
				Equals: func(a, b any) bool {
					// Only consider changes between true and false significant
					return a.(bool) == b.(bool)
				},
			},
		)

		// Initial compute ran once during creation
		afterCreate = computeCounter
		t.Logf("After create: computeCounter = %d (incremented by %d)", computeCounter, afterCreate-beforeCreate)
		assert.Equal(t, 1, computeCounter, "Compute function should run once during creation")
		assert.Equal(t, true, computed.Value(), "Initial computed value should be true (5==5)")

		// Update a to same value - this should NOT trigger a recompute since the value hasn't changed
		// The Set method of Signal correctly optimizes by not notifying dependents when value doesn't change
		t.Logf("Before first set: computeCounter = %d", computeCounter)
		a.Set(5) // Setting to same value, should be optimized away
		afterSetSame = computeCounter
		t.Logf("After setting a to 5 (same value): computeCounter = %d (change: %d)", computeCounter, afterSetSame-afterCreate)
		// Still 1 since setting the same value doesn't notify dependents
		assert.Equal(t, 1, computeCounter, "Compute function should not run when setting the same value")
		assert.Equal(t, true, computed.Value(), "Computed value should still be true (5==5)")
		afterAccessValue = computeCounter
		t.Logf("After accessing computed.Value(): computeCounter = %d (change: %d)", computeCounter, afterAccessValue-afterSetSame)

		// Now update a to a value that is actually different
		a.Set(10) // Setting to different value
		afterSetDifferent = computeCounter
		t.Logf("After setting a to 10: computeCounter = %d (change: %d)", computeCounter, afterSetDifferent-afterAccessValue)
		assert.Equal(t, 2, computeCounter, "Compute function should run when dependency actually changes")
		assert.Equal(t, false, computed.Value(), "Computed value should now be false (10!=5)")
		t.Logf("Final computeCounter = %d", computeCounter)
	})
}

func TestSignalMetadata(t *testing.T) {
	t.Run("Stats and Metadata", func(t *testing.T) {
		// Create a signal
		signal := CreateSignal(42, SignalOptions{
			DebugName: "TestSignal",
		})

		// Access the value a few times - explicitly call Value() 4 times to ensure we have enough reads
		initialValue := signal.Value()                                  // First read
		assert.Equal(t, 42, initialValue, "Initial value should be 42") // Verify the value is correct
		signal.Value()                                                  // Second read
		signal.Value()                                                  // Third read
		signal.Value()                                                  // Fourth read

		// Make a few updates
		signal.Set(100)
		signal.Set(200)

		// Read the updated value to increment access count again
		updatedValue := signal.Value() // Fifth read
		assert.Equal(t, 200, updatedValue, "Updated value should be 200")

		// Check stats
		stats := signal.GetStats()
		assert.NotNil(t, stats, "Stats should not be nil")

		// Log actual values for debugging
		t.Logf("Access count: %v", stats["accessCount"])
		t.Logf("Write count: %v", stats["writeCount"])

		assert.Equal(t, "TestSignal", stats["debugName"], "Debug name should be in stats")
		assert.Equal(t, uint64(2), stats["writeCount"], "Should have recorded 2 write operations")
		assert.GreaterOrEqual(t, stats["accessCount"].(uint64), uint64(5), "Should have recorded at least 5 read operations")

		// Add custom metadata
		signal.SetMetadata("customKey", "customValue")

		// Retrieve and check custom metadata
		value, exists := signal.GetMetadata("customKey")
		assert.True(t, exists, "Custom metadata should exist")
		assert.Equal(t, "customValue", value, "Custom metadata should have correct value")

		// Check that metadata appears in stats
		updatedStats := signal.GetStats()
		assert.Equal(t, "customValue", updatedStats["customKey"], "Custom metadata should appear in stats")
	})
}
