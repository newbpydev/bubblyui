package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCustomComparatorTester_CustomComparatorUsed tests that custom comparator is used for equality checks
func TestCustomComparatorTester_CustomComparatorUsed(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	// Custom comparator that only compares ID
	compareByID := func(a, b interface{}) bool {
		userA, okA := a.(User)
		userB, okB := b.(User)
		if !okA || !okB {
			return false
		}
		return userA.ID == userB.ID
	}

	ref := bubbly.NewRef[interface{}](User{ID: 1, Name: "Alice"})
	tester := NewCustomComparatorTester(ref, compareByID)

	// Change name but keep ID same - comparator should say "equal"
	tester.SetValue(User{ID: 1, Name: "Bob"})

	// Should not have changed according to custom comparator
	tester.AssertChanged(t, false)
	tester.AssertComparisons(t, 1)
}

// TestCustomComparatorTester_ComparisonCountTracked tests that comparison count is tracked correctly
func TestCustomComparatorTester_ComparisonCountTracked(t *testing.T) {
	ref := bubbly.NewRef[interface{}](42)

	// Simple equality comparator
	compareInts := func(a, b interface{}) bool {
		intA, okA := a.(int)
		intB, okB := b.(int)
		if !okA || !okB {
			return false
		}
		return intA == intB
	}

	tester := NewCustomComparatorTester(ref, compareInts)

	// Multiple SetValue calls should track comparisons
	tester.SetValue(42) // Same value
	tester.AssertComparisons(t, 1)
	tester.AssertChanged(t, false)

	tester.SetValue(43) // Different value
	tester.AssertComparisons(t, 2)
	tester.AssertChanged(t, true)

	tester.SetValue(43) // Same as current
	tester.AssertComparisons(t, 3)
	tester.AssertChanged(t, false) // Reset to false since last change was equal
}

// TestCustomComparatorTester_LogicalEqualityVsIdentity tests logical equality vs identity
func TestCustomComparatorTester_LogicalEqualityVsIdentity(t *testing.T) {
	type Point struct {
		X, Y int
	}

	// Logical equality comparator
	comparePoints := func(a, b interface{}) bool {
		pointA, okA := a.(Point)
		pointB, okB := b.(Point)
		if !okA || !okB {
			return false
		}
		return pointA.X == pointB.X && pointA.Y == pointB.Y
	}

	ref := bubbly.NewRef[interface{}](Point{X: 1, Y: 2})
	tester := NewCustomComparatorTester(ref, comparePoints)

	// Different instance but logically equal
	tester.SetValue(Point{X: 1, Y: 2})
	tester.AssertChanged(t, false) // Logically equal
	tester.AssertComparisons(t, 1)

	// Different value
	tester.SetValue(Point{X: 3, Y: 4})
	tester.AssertChanged(t, true)
	tester.AssertComparisons(t, 2)
}

// TestCustomComparatorTester_StructComparators tests struct field-based comparators
func TestCustomComparatorTester_StructComparators(t *testing.T) {
	type Config struct {
		Version int
		Data    map[string]string
		Debug   bool
	}

	tests := []struct {
		name       string
		comparator func(a, b interface{}) bool
		initial    Config
		newValue   Config
		wantChange bool
	}{
		{
			name: "compare only version",
			comparator: func(a, b interface{}) bool {
				cfgA, okA := a.(Config)
				cfgB, okB := b.(Config)
				if !okA || !okB {
					return false
				}
				return cfgA.Version == cfgB.Version
			},
			initial:    Config{Version: 1, Debug: false},
			newValue:   Config{Version: 1, Debug: true}, // Debug changed but Version same
			wantChange: false,
		},
		{
			name: "compare version and debug",
			comparator: func(a, b interface{}) bool {
				cfgA, okA := a.(Config)
				cfgB, okB := b.(Config)
				if !okA || !okB {
					return false
				}
				return cfgA.Version == cfgB.Version && cfgA.Debug == cfgB.Debug
			},
			initial:    Config{Version: 1, Debug: false},
			newValue:   Config{Version: 1, Debug: true}, // Debug changed
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef[interface{}](tt.initial)
			tester := NewCustomComparatorTester(ref, tt.comparator)

			tester.SetValue(tt.newValue)

			tester.AssertChanged(t, tt.wantChange)
			tester.AssertComparisons(t, 1)
		})
	}
}

// TestCustomComparatorTester_ArrayComparators tests array/slice comparators
func TestCustomComparatorTester_ArrayComparators(t *testing.T) {
	// Comparator that only checks length
	compareLengthOnly := func(a, b interface{}) bool {
		sliceA, okA := a.([]int)
		sliceB, okB := b.([]int)
		if !okA || !okB {
			return false
		}
		return len(sliceA) == len(sliceB)
	}

	ref := bubbly.NewRef[interface{}]([]int{1, 2, 3})
	tester := NewCustomComparatorTester(ref, compareLengthOnly)

	// Different elements but same length
	tester.SetValue([]int{4, 5, 6})
	tester.AssertChanged(t, false) // Same length
	tester.AssertComparisons(t, 1)

	// Different length
	tester.SetValue([]int{1, 2})
	tester.AssertChanged(t, true)
	tester.AssertComparisons(t, 2)
}

// TestCustomComparatorTester_PerformanceOptimization tests that custom comparators can optimize performance
func TestCustomComparatorTester_PerformanceOptimization(t *testing.T) {
	type LargeStruct struct {
		ID       int
		Name     string
		Metadata map[string]interface{} // Large field we want to ignore
	}

	// Fast comparator that only checks ID
	fastCompare := func(a, b interface{}) bool {
		structA, okA := a.(LargeStruct)
		structB, okB := b.(LargeStruct)
		if !okA || !okB {
			return false
		}
		return structA.ID == structB.ID
	}

	largeMetadata := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeMetadata[string(rune(i))] = i
	}

	ref := bubbly.NewRef[interface{}](LargeStruct{
		ID:       1,
		Name:     "Test",
		Metadata: largeMetadata,
	})

	tester := NewCustomComparatorTester(ref, fastCompare)

	// Change metadata but keep ID same - should be fast
	newMetadata := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		newMetadata[string(rune(i))] = i * 2 // Different values
	}

	tester.SetValue(LargeStruct{
		ID:       1,
		Name:     "Different",
		Metadata: newMetadata,
	})

	// Should not have changed (ID is same)
	tester.AssertChanged(t, false)
	tester.AssertComparisons(t, 1)

	// Now change ID
	tester.SetValue(LargeStruct{
		ID:       2,
		Name:     "Test",
		Metadata: largeMetadata,
	})

	tester.AssertChanged(t, true)
	tester.AssertComparisons(t, 2)
}
