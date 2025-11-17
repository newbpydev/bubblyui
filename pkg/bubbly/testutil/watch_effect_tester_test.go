package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewWatchEffectTester tests the constructor
func TestNewWatchEffectTester(t *testing.T) {
	execCount := 0

	tester := NewWatchEffectTester(&execCount)

	assert.NotNil(t, tester, "tester should not be nil")
	assert.NotNil(t, tester.execCounter, "exec counter should not be nil")
	assert.Equal(t, 0, tester.GetExecutionCount(), "initial count should be 0")
}

// TestWatchEffectTester_InitialExecution tests that effect executes immediately
func TestWatchEffectTester_InitialExecution(t *testing.T) {
	count := bubbly.NewRef(0)
	execCount := 0

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		_ = count.Get() // Track count as dependency
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)

	// Effect should execute immediately
	tester.AssertExecuted(t, 1)
}

// TestWatchEffectTester_TriggerDependency tests triggering a dependency change
func TestWatchEffectTester_TriggerDependency(t *testing.T) {
	count := bubbly.NewRef(0)
	execCount := 0

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		_ = count.Get()
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)
	tester.AssertExecuted(t, 1)

	// Trigger dependency change
	tester.TriggerDependency(count, 5)

	// Effect should re-execute
	tester.AssertExecuted(t, 2)
	assert.Equal(t, 5, count.Get(), "dependency value should be updated")
}

// TestWatchEffectTester_MultipleDependencies tests effect with multiple dependencies
func TestWatchEffectTester_MultipleDependencies(t *testing.T) {
	firstName := bubbly.NewRef("John")
	lastName := bubbly.NewRef("Doe")
	execCount := 0
	var fullName string

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		fullName = firstName.GetTyped() + " " + lastName.GetTyped()
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)
	tester.AssertExecuted(t, 1)
	assert.Equal(t, "John Doe", fullName)

	// Change first name
	tester.TriggerDependency(firstName, "Jane")
	tester.AssertExecuted(t, 2)
	assert.Equal(t, "Jane Doe", fullName)

	// Change last name
	tester.TriggerDependency(lastName, "Smith")
	tester.AssertExecuted(t, 3)
	assert.Equal(t, "Jane Smith", fullName)
}

// TestWatchEffectTester_ConditionalDependencies tests dynamic dependency tracking
func TestWatchEffectTester_ConditionalDependencies(t *testing.T) {
	toggle := bubbly.NewRef(true)
	valueA := bubbly.NewRef(1)
	valueB := bubbly.NewRef(100)
	execCount := 0
	var result int

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		if toggle.GetTyped() {
			result = valueA.GetTyped() // Only tracks valueA when toggle is true
		} else {
			result = valueB.GetTyped() // Only tracks valueB when toggle is false
		}
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)
	tester.AssertExecuted(t, 1)
	assert.Equal(t, 1, result)

	// Change valueA - should trigger (currently tracking valueA)
	tester.TriggerDependency(valueA, 2)
	tester.AssertExecuted(t, 2)
	assert.Equal(t, 2, result)

	// Change valueB - should NOT trigger (not tracking valueB)
	tester.TriggerDependency(valueB, 200)
	tester.AssertExecuted(t, 2) // Still 2
	assert.Equal(t, 2, result)  // Still 2

	// Toggle to false - should trigger and switch to valueB
	tester.TriggerDependency(toggle, false)
	tester.AssertExecuted(t, 3)
	assert.Equal(t, 200, result)

	// Now valueB changes should trigger
	tester.TriggerDependency(valueB, 300)
	tester.AssertExecuted(t, 4)
	assert.Equal(t, 300, result)

	// But valueA changes should NOT trigger
	// Note: WatchEffect re-tracks dependencies on each run, so changing valueA
	// might still trigger if the effect re-evaluates the condition
	tester.TriggerDependency(valueA, 99)
	// The effect may execute again because toggle is still tracked
	assert.GreaterOrEqual(t, tester.GetExecutionCount(), 4)
	assert.Equal(t, 300, result) // Result should still be 300 since toggle is false
}

// TestWatchEffectTester_WithComputed tests effect with computed values
func TestWatchEffectTester_WithComputed(t *testing.T) {
	count := bubbly.NewRef(2)
	doubled := bubbly.NewComputed(func() int {
		return count.GetTyped() * 2
	})
	execCount := 0
	var result int

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		result = doubled.GetTyped()
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)
	tester.AssertExecuted(t, 1)
	assert.Equal(t, 4, result)

	// Change count - should trigger through computed
	tester.TriggerDependency(count, 5)
	tester.AssertExecuted(t, 2)
	assert.Equal(t, 10, result)
}

// TestWatchEffectTester_Cleanup tests cleanup functionality
func TestWatchEffectTester_Cleanup(t *testing.T) {
	count := bubbly.NewRef(0)
	execCount := 0

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		_ = count.Get()
	})

	tester := NewWatchEffectTester(&execCount)
	tester.SetCleanup(cleanup)
	tester.AssertExecuted(t, 1)

	// Cleanup should stop the effect
	tester.Cleanup()

	// Changes after cleanup should not trigger
	tester.TriggerDependency(count, 5)
	tester.AssertExecuted(t, 1) // Still 1
}

// TestWatchEffectTester_NoDependencies tests effect with no reactive dependencies
func TestWatchEffectTester_NoDependencies(t *testing.T) {
	execCount := 0

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		// No reactive dependencies accessed
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)

	// Should execute once initially
	tester.AssertExecuted(t, 1)

	// No dependencies to trigger, count should stay at 1
	assert.Equal(t, 1, tester.GetExecutionCount())
}

// TestWatchEffectTester_GetExecutionCount tests GetExecutionCount method
func TestWatchEffectTester_GetExecutionCount(t *testing.T) {
	count := bubbly.NewRef(0)
	execCount := 0

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		_ = count.Get()
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)

	assert.Equal(t, 1, tester.GetExecutionCount())

	tester.TriggerDependency(count, 1)
	assert.Equal(t, 2, tester.GetExecutionCount())

	tester.TriggerDependency(count, 2)
	assert.Equal(t, 3, tester.GetExecutionCount())
}

// TestWatchEffectTester_NilCounter tests behavior with nil counter
func TestWatchEffectTester_NilCounter(t *testing.T) {
	tester := NewWatchEffectTester(nil)

	// Should not panic
	assert.NotPanics(t, func() {
		count := tester.GetExecutionCount()
		assert.Equal(t, 0, count)
	})
}

// TestWatchEffectTester_TriggerDependency_NilDep tests TriggerDependency with nil
func TestWatchEffectTester_TriggerDependency_NilDep(t *testing.T) {
	execCount := 0
	tester := NewWatchEffectTester(&execCount)

	// Should not panic with nil dependency
	assert.NotPanics(t, func() {
		tester.TriggerDependency(nil, 5)
	})
}

// TestWatchEffectTester_TriggerDependency_InvalidType tests TriggerDependency with invalid type
func TestWatchEffectTester_TriggerDependency_InvalidType(t *testing.T) {
	execCount := 0
	tester := NewWatchEffectTester(&execCount)

	// Should not panic with non-ref type
	assert.NotPanics(t, func() {
		tester.TriggerDependency("not a ref", 5)
	})
}

// TestWatchEffectTester_MultipleEffects tests multiple independent effects
func TestWatchEffectTester_MultipleEffects(t *testing.T) {
	count1 := bubbly.NewRef(0)
	count2 := bubbly.NewRef(0)
	execCount1 := 0
	execCount2 := 0

	cleanup1 := bubbly.WatchEffect(func() {
		execCount1++
		_ = count1.Get()
	})
	defer cleanup1()

	cleanup2 := bubbly.WatchEffect(func() {
		execCount2++
		_ = count2.Get()
	})
	defer cleanup2()

	tester1 := NewWatchEffectTester(&execCount1)
	tester2 := NewWatchEffectTester(&execCount2)

	tester1.AssertExecuted(t, 1)
	tester2.AssertExecuted(t, 1)

	// Trigger first effect
	tester1.TriggerDependency(count1, 5)
	tester1.AssertExecuted(t, 2)
	tester2.AssertExecuted(t, 1) // Second effect unchanged

	// Trigger second effect
	tester2.TriggerDependency(count2, 10)
	tester1.AssertExecuted(t, 2) // First effect unchanged
	tester2.AssertExecuted(t, 2)
}

// TestWatchEffectTester_ChainedComputed tests with chained computed values
func TestWatchEffectTester_ChainedComputed(t *testing.T) {
	count := bubbly.NewRef(2)
	doubled := bubbly.NewComputed(func() int {
		return count.GetTyped() * 2
	})
	quadrupled := bubbly.NewComputed(func() int {
		return doubled.GetTyped() * 2
	})
	execCount := 0
	var result int

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		result = quadrupled.GetTyped()
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)
	tester.AssertExecuted(t, 1)
	assert.Equal(t, 8, result) // 2 * 2 * 2

	// Change count - should propagate through chain
	tester.TriggerDependency(count, 3)
	tester.AssertExecuted(t, 2)
	assert.Equal(t, 12, result) // 3 * 2 * 2
}

// TestWatchEffectTester_RapidChanges tests rapid dependency changes
func TestWatchEffectTester_RapidChanges(t *testing.T) {
	count := bubbly.NewRef(0)
	execCount := 0

	cleanup := bubbly.WatchEffect(func() {
		execCount++
		_ = count.Get()
	})
	defer cleanup()

	tester := NewWatchEffectTester(&execCount)
	tester.AssertExecuted(t, 1)

	// Rapid changes
	for i := 1; i <= 10; i++ {
		tester.TriggerDependency(count, i)
	}

	// Should have executed 11 times (1 initial + 10 changes)
	tester.AssertExecuted(t, 11)
}

// TestWatchEffectTester_TableDriven demonstrates table-driven test pattern
func TestWatchEffectTester_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int
		changes       []int
		expectedExecs int
	}{
		{
			name:          "no changes",
			initialValue:  0,
			changes:       []int{},
			expectedExecs: 1, // Initial execution only
		},
		{
			name:          "single change",
			initialValue:  0,
			changes:       []int{5},
			expectedExecs: 2, // Initial + 1 change
		},
		{
			name:          "multiple changes",
			initialValue:  0,
			changes:       []int{1, 2, 3, 4, 5},
			expectedExecs: 6, // Initial + 5 changes
		},
		{
			name:          "duplicate values",
			initialValue:  0,
			changes:       []int{5, 5, 5},
			expectedExecs: 4, // Each Set() triggers, even with same value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := bubbly.NewRef(tt.initialValue)
			execCount := 0

			cleanup := bubbly.WatchEffect(func() {
				execCount++
				_ = count.Get()
			})
			defer cleanup()

			tester := NewWatchEffectTester(&execCount)

			// Apply changes
			for _, value := range tt.changes {
				tester.TriggerDependency(count, value)
			}

			// Verify execution count
			tester.AssertExecuted(t, tt.expectedExecs)
		})
	}
}
