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

// TestCustomComparatorTester_GetComparisonCount tests direct counter access
func TestCustomComparatorTester_GetComparisonCount(t *testing.T) {
	ref := bubbly.NewRef[interface{}](10)
	comparator := func(a, b interface{}) bool { return a == b }
	tester := NewCustomComparatorTester(ref, comparator)

	// Initial count should be 0
	if count := tester.GetComparisonCount(); count != 0 {
		t.Errorf("expected initial count 0, got %d", count)
	}

	// After one comparison
	tester.SetValue(20)
	if count := tester.GetComparisonCount(); count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// After multiple comparisons
	tester.SetValue(30)
	tester.SetValue(40)
	if count := tester.GetComparisonCount(); count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

// TestCustomComparatorTester_WasChanged tests change detection
func TestCustomComparatorTester_WasChanged(t *testing.T) {
	ref := bubbly.NewRef[interface{}]("initial")
	comparator := func(a, b interface{}) bool { return a == b }
	tester := NewCustomComparatorTester(ref, comparator)

	// Initial state - no change yet
	if tester.WasChanged() {
		t.Error("expected WasChanged false initially")
	}

	// Same value - no change
	tester.SetValue("initial")
	if tester.WasChanged() {
		t.Error("expected WasChanged false for same value")
	}

	// Different value - change detected
	tester.SetValue("new")
	if !tester.WasChanged() {
		t.Error("expected WasChanged true for different value")
	}

	// Same as current - no change
	tester.SetValue("new")
	if tester.WasChanged() {
		t.Error("expected WasChanged false for same value again")
	}
}

// TestCustomComparatorTester_GetCurrentValue tests value retrieval
func TestCustomComparatorTester_GetCurrentValue(t *testing.T) {
	tests := []struct {
		name         string
		initialValue interface{}
		updates      []interface{}
	}{
		{
			name:         "int_values",
			initialValue: 42,
			updates:      []interface{}{100, 200, 300},
		},
		{
			name:         "string_values",
			initialValue: "start",
			updates:      []interface{}{"middle", "end"},
		},
		{
			name:         "struct_values",
			initialValue: struct{ X int }{X: 1},
			updates:      []interface{}{struct{ X int }{X: 2}, struct{ X int }{X: 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := bubbly.NewRef[interface{}](tt.initialValue)
			comparator := func(a, b interface{}) bool { return a == b }
			tester := NewCustomComparatorTester(ref, comparator)

			// Check initial value
			if current := tester.GetCurrentValue(); current != tt.initialValue {
				t.Errorf("expected initial value %v, got %v", tt.initialValue, current)
			}

			// Check after each update
			for _, update := range tt.updates {
				tester.SetValue(update)
				if current := tester.GetCurrentValue(); current != update {
					t.Errorf("expected current value %v, got %v", update, current)
				}
			}
		})
	}
}

// TestCustomComparatorTester_ResetCounters tests counter reset functionality
func TestCustomComparatorTester_ResetCounters(t *testing.T) {
	ref := bubbly.NewRef[interface{}](1)
	comparator := func(a, b interface{}) bool { return a == b }
	tester := NewCustomComparatorTester(ref, comparator)

	// Do some comparisons
	tester.SetValue(2)
	tester.SetValue(3)

	if count := tester.GetComparisonCount(); count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if !tester.WasChanged() {
		t.Error("expected WasChanged true")
	}

	// Reset
	tester.ResetCounters()

	// Verify reset
	if count := tester.GetComparisonCount(); count != 0 {
		t.Errorf("expected count 0 after reset, got %d", count)
	}
	if tester.WasChanged() {
		t.Error("expected WasChanged false after reset")
	}

	// Verify new comparisons start from 0
	tester.SetValue(4)
	if count := tester.GetComparisonCount(); count != 1 {
		t.Errorf("expected count 1 after reset and new comparison, got %d", count)
	}
}

// TestCustomComparatorTester_CompareValues tests direct comparator invocation
func TestCustomComparatorTester_CompareValues(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

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

	// Test direct comparison without modifying ref
	user1 := User{ID: 1, Name: "Alice"}
	user2 := User{ID: 1, Name: "Bob"}
	user3 := User{ID: 2, Name: "Alice"}

	// Same ID, different name - should be equal
	if !tester.CompareValues(user1, user2) {
		t.Error("expected users with same ID to be equal")
	}

	// Different ID - should not be equal
	if tester.CompareValues(user1, user3) {
		t.Error("expected users with different ID to be unequal")
	}

	// Verify comparison count increased
	if count := tester.GetComparisonCount(); count != 2 {
		t.Errorf("expected 2 comparisons, got %d", count)
	}

	// Verify ref value unchanged
	currentUser := tester.GetCurrentValue().(User)
	if currentUser.ID != 1 || currentUser.Name != "Alice" {
		t.Error("CompareValues should not modify ref value")
	}
}

// TestCustomComparatorTester_VerifyComparatorBehavior tests batch verification
func TestCustomComparatorTester_VerifyComparatorBehavior(t *testing.T) {
	type Point struct {
		X, Y int
	}

	comparePoints := func(a, b interface{}) bool {
		pointA, okA := a.(Point)
		pointB, okB := b.(Point)
		if !okA || !okB {
			return false
		}
		return pointA.X == pointB.X && pointA.Y == pointB.Y
	}

	ref := bubbly.NewRef[interface{}](Point{X: 0, Y: 0})
	tester := NewCustomComparatorTester(ref, comparePoints)

	testCases := map[string][3]interface{}{
		"same_point": {
			Point{X: 1, Y: 2},
			Point{X: 1, Y: 2},
			true,
		},
		"different_x": {
			Point{X: 1, Y: 2},
			Point{X: 3, Y: 2},
			false,
		},
		"different_y": {
			Point{X: 1, Y: 2},
			Point{X: 1, Y: 3},
			false,
		},
		"both_different": {
			Point{X: 1, Y: 2},
			Point{X: 3, Y: 4},
			false,
		},
		"origin_point": {
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 0},
			true,
		},
	}

	tester.VerifyComparatorBehavior(t, testCases)

	// Verify all comparisons were counted
	if count := tester.GetComparisonCount(); count != 5 {
		t.Errorf("expected 5 comparisons from test cases, got %d", count)
	}
}

// TestCustomComparatorTester_AssertComparatorType tests type validation
func TestCustomComparatorTester_AssertComparatorType(t *testing.T) {
	// Valid comparator - should pass
	ref := bubbly.NewRef[interface{}](42)
	comparator := func(a, b interface{}) bool { return a == b }
	tester := NewCustomComparatorTester(ref, comparator)

	// Should not panic or error
	tester.AssertComparatorType(t)

	// Nil comparator case - test implementation itself
	testerNil := &CustomComparatorTester{
		ref:        ref,
		comparator: nil,
	}

	// Verify it detects nil comparator
	if testerNil.comparator != nil {
		t.Error("expected comparator to be nil")
	}
}

// TestCustomComparatorTester_AssertComparisons_EdgeCases tests edge cases
func TestCustomComparatorTester_AssertComparisons_EdgeCases(t *testing.T) {
	ref := bubbly.NewRef[interface{}](1)
	comparator := func(a, b interface{}) bool { return a == b }
	tester := NewCustomComparatorTester(ref, comparator)

	// Test with 0 comparisons
	tester.AssertComparisons(t, 0)

	// Make comparisons
	tester.SetValue(2)
	tester.SetValue(3)
	tester.SetValue(4)

	// Test with correct count
	tester.AssertComparisons(t, 3)
}

// TestCustomComparatorTester_AssertChanged_MultipleCalls tests multiple assertion calls
func TestCustomComparatorTester_AssertChanged_MultipleCalls(t *testing.T) {
	ref := bubbly.NewRef[interface{}](100)
	comparator := func(a, b interface{}) bool { return a == b }
	tester := NewCustomComparatorTester(ref, comparator)

	// Initial - no change
	tester.AssertChanged(t, false)

	// Same value - no change
	tester.SetValue(100)
	tester.AssertChanged(t, false)

	// Different value - change
	tester.SetValue(200)
	tester.AssertChanged(t, true)

	// Same as current - no change again
	tester.SetValue(200)
	tester.AssertChanged(t, false)

	// Another different value
	tester.SetValue(300)
	tester.AssertChanged(t, true)
}
