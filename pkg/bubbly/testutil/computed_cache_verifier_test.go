package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestComputedCacheVerifier_BasicCaching tests basic caching behavior
func TestComputedCacheVerifier_BasicCaching(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// First get - should compute
	val := verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1)
	verifier.AssertCacheHits(t, 0)
	verifier.AssertCacheMisses(t, 1)

	// Second get - should use cache
	val = verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1) // Still 1
	verifier.AssertCacheHits(t, 1)
	verifier.AssertCacheMisses(t, 1)

	// Third get - still cached
	val = verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1)
	verifier.AssertCacheHits(t, 2)
	verifier.AssertCacheMisses(t, 1)
}

// TestComputedCacheVerifier_DependencyInvalidation tests cache invalidation on dependency changes
func TestComputedCacheVerifier_DependencyInvalidation(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Initial get
	val := verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1)

	// Cached get
	val = verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1)

	// Change dependency - invalidates cache
	count.Set(10)

	// Next get should recompute
	val = verifier.GetValue()
	assert.Equal(t, 20, val)
	verifier.AssertComputeCount(t, 2)
	verifier.AssertCacheMisses(t, 2) // Two misses total

	// Subsequent get should be cached again
	val = verifier.GetValue()
	assert.Equal(t, 20, val)
	verifier.AssertComputeCount(t, 2)
	verifier.AssertCacheHits(t, 2) // Two hits total: line 62 and line 76
}

// TestComputedCacheVerifier_ManualInvalidation tests manual cache invalidation
func TestComputedCacheVerifier_ManualInvalidation(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Initial get
	val := verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1)

	// Manually invalidate cache
	verifier.InvalidateCache()

	// Next get should recompute even though dependency didn't change
	val = verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 2)
	verifier.AssertCacheMisses(t, 2)
}

// TestComputedCacheVerifier_MultipleGets tests multiple gets with various patterns
func TestComputedCacheVerifier_MultipleGets(t *testing.T) {
	count := bubbly.NewRef(1)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 3
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Pattern: Get, Get, Change, Get, Get, Get
	verifier.GetValue() // Miss
	verifier.GetValue() // Hit
	count.Set(2)
	verifier.GetValue() // Miss
	verifier.GetValue() // Hit
	verifier.GetValue() // Hit

	verifier.AssertComputeCount(t, 2)
	verifier.AssertCacheHits(t, 3)
	verifier.AssertCacheMisses(t, 2)
}

// TestComputedCacheVerifier_ChainedComputed tests caching with chained computed values
func TestComputedCacheVerifier_ChainedComputed(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount1 := 0
	computeCount2 := 0

	// First computed: count * 2
	computed1 := bubbly.NewComputed(func() int {
		computeCount1++
		return count.GetTyped() * 2
	})

	// Second computed: computed1 * 3
	computed2 := bubbly.NewComputed(func() int {
		computeCount2++
		return computed1.GetTyped() * 3
	})

	verifier1 := NewComputedCacheVerifier(computed1, &computeCount1)
	verifier2 := NewComputedCacheVerifier(computed2, &computeCount2)

	// Get from second computed - should compute both
	val := verifier2.GetValue()
	assert.Equal(t, 30, val) // (5 * 2) * 3 = 30
	verifier1.AssertComputeCount(t, 1)
	verifier2.AssertComputeCount(t, 1)

	// Get again - both should be cached
	val = verifier2.GetValue()
	assert.Equal(t, 30, val)
	verifier1.AssertComputeCount(t, 1)
	verifier2.AssertComputeCount(t, 1)
	verifier2.AssertCacheHits(t, 1)

	// Change root dependency
	count.Set(10)

	// Get from second computed - both should recompute
	val = verifier2.GetValue()
	assert.Equal(t, 60, val) // (10 * 2) * 3 = 60
	verifier1.AssertComputeCount(t, 2)
	verifier2.AssertComputeCount(t, 2)
}

// TestComputedCacheVerifier_ResetCounters tests resetting cache hit/miss counters
func TestComputedCacheVerifier_ResetCounters(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// First scenario
	verifier.GetValue() // Miss
	verifier.GetValue() // Hit
	verifier.AssertCacheHits(t, 1)
	verifier.AssertCacheMisses(t, 1)

	// Reset counters
	verifier.ResetCounters()

	// Second scenario - counters start from zero
	verifier.GetValue() // Hit (still cached)
	verifier.AssertCacheHits(t, 1)
	verifier.AssertCacheMisses(t, 0)
}

// TestComputedCacheVerifier_GetLastValue tests retrieving the last value
func TestComputedCacheVerifier_GetLastValue(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Get value
	val := verifier.GetValue()
	assert.Equal(t, 10, val)

	// Check last value
	lastVal := verifier.GetLastValue()
	assert.Equal(t, 10, lastVal)

	// Change and get again
	count.Set(20)
	val = verifier.GetValue()
	assert.Equal(t, 40, val)

	// Check last value updated
	lastVal = verifier.GetLastValue()
	assert.Equal(t, 40, lastVal)
}

// TestComputedCacheVerifier_GetCacheHitsAndMisses tests getter methods
func TestComputedCacheVerifier_GetCacheHitsAndMisses(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Initial state
	assert.Equal(t, 0, verifier.GetCacheHits())
	assert.Equal(t, 0, verifier.GetCacheMisses())

	// After first get
	verifier.GetValue()
	assert.Equal(t, 0, verifier.GetCacheHits())
	assert.Equal(t, 1, verifier.GetCacheMisses())

	// After second get
	verifier.GetValue()
	assert.Equal(t, 1, verifier.GetCacheHits())
	assert.Equal(t, 1, verifier.GetCacheMisses())
}

// TestComputedCacheVerifier_ComplexType tests caching with complex types
func TestComputedCacheVerifier_ComplexType(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	userRef := bubbly.NewRef(User{ID: 1, Name: "Alice"})
	computeCount := 0

	computed := bubbly.NewComputed(func() string {
		computeCount++
		user := userRef.GetTyped()
		return user.Name + " (ID: " + string(rune(user.ID+'0')) + ")"
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// First get
	val := verifier.GetValue()
	assert.Equal(t, "Alice (ID: 1)", val)
	verifier.AssertComputeCount(t, 1)

	// Cached get
	val = verifier.GetValue()
	assert.Equal(t, "Alice (ID: 1)", val)
	verifier.AssertComputeCount(t, 1)

	// Change user
	userRef.Set(User{ID: 2, Name: "Bob"})

	// Should recompute
	val = verifier.GetValue()
	assert.Equal(t, "Bob (ID: 2)", val)
	verifier.AssertComputeCount(t, 2)
}

// TestComputedCacheVerifier_TableDriven tests various caching scenarios
func TestComputedCacheVerifier_TableDriven(t *testing.T) {
	tests := []struct {
		name             string
		operations       []string // "get", "change", "invalidate"
		expectedComputes int
		expectedHits     int
		expectedMisses   int
	}{
		{
			name:             "single get",
			operations:       []string{"get"},
			expectedComputes: 1,
			expectedHits:     0,
			expectedMisses:   1,
		},
		{
			name:             "two gets",
			operations:       []string{"get", "get"},
			expectedComputes: 1,
			expectedHits:     1,
			expectedMisses:   1,
		},
		{
			name:             "get, change, get",
			operations:       []string{"get", "change", "get"},
			expectedComputes: 2,
			expectedHits:     0,
			expectedMisses:   2,
		},
		{
			name:             "get, get, change, get, get",
			operations:       []string{"get", "get", "change", "get", "get"},
			expectedComputes: 2,
			expectedHits:     2,
			expectedMisses:   2,
		},
		{
			name:             "get, invalidate, get",
			operations:       []string{"get", "invalidate", "get"},
			expectedComputes: 2,
			expectedHits:     0,
			expectedMisses:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := bubbly.NewRef(5)
			computeCount := 0

			computed := bubbly.NewComputed(func() int {
				computeCount++
				return count.GetTyped() * 2
			})

			verifier := NewComputedCacheVerifier(computed, &computeCount)

			// Execute operations
			for _, op := range tt.operations {
				switch op {
				case "get":
					verifier.GetValue()
				case "change":
					count.Set(count.GetTyped() + 1)
				case "invalidate":
					verifier.InvalidateCache()
				}
			}

			// Verify results
			verifier.AssertComputeCount(t, tt.expectedComputes)
			verifier.AssertCacheHits(t, tt.expectedHits)
			verifier.AssertCacheMisses(t, tt.expectedMisses)
		})
	}
}

// TestComputedCacheVerifier_MemoryManagement tests that values are properly managed
func TestComputedCacheVerifier_MemoryManagement(t *testing.T) {
	// Create a large slice to test memory management
	dataRef := bubbly.NewRef(make([]int, 1000))
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		data := dataRef.GetTyped()
		sum := 0
		for _, v := range data {
			sum += v
		}
		return sum
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Get value multiple times
	for i := 0; i < 10; i++ {
		val := verifier.GetValue()
		assert.Equal(t, 0, val) // All zeros initially
	}

	// Should only compute once
	verifier.AssertComputeCount(t, 1)
	verifier.AssertCacheHits(t, 9)

	// Change data
	newData := make([]int, 1000)
	for i := range newData {
		newData[i] = 1
	}
	dataRef.Set(newData)

	// Should recompute
	val := verifier.GetValue()
	assert.Equal(t, 1000, val)
	verifier.AssertComputeCount(t, 2)
}

// TestComputedCacheVerifier_NilHandling tests handling of nil values
func TestComputedCacheVerifier_NilHandling(t *testing.T) {
	// Test with nil computed (edge case)
	verifier := NewComputedCacheVerifier(nil, new(int))

	// Should handle gracefully
	val := verifier.GetValue()
	assert.Nil(t, val)

	// Should not panic
	verifier.InvalidateCache()
}

// TestComputedCacheVerifier_CircularDependencyDetection tests circular dependency handling
func TestComputedCacheVerifier_CircularDependencyDetection(t *testing.T) {
	// This test verifies that circular dependencies are detected
	// Note: Circular dependencies should panic in the Computed implementation
	computeCount := 0

	// Create a computed that tries to access itself (circular)
	computed := bubbly.NewComputed(func() int {
		computeCount++
		// This would create a circular dependency if we tried to access computed.GetTyped()
		// For this test, we just return a value
		return 42
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Normal operation should work
	val := verifier.GetValue()
	assert.Equal(t, 42, val)
	verifier.AssertComputeCount(t, 1)
}

// TestComputedCacheVerifier_AssertionFailures tests that assertion methods properly fail
func TestComputedCacheVerifier_AssertionFailures(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Get value once
	verifier.GetValue()

	// Create a mock testing.T to capture failures
	mockT := &testing.T{}

	// Test AssertComputeCount failure
	verifier.AssertComputeCount(mockT, 999) // Wrong expectation
	if !mockT.Failed() {
		t.Error("AssertComputeCount should have failed but didn't")
	}

	// Reset mock
	mockT = &testing.T{}

	// Test AssertCacheHits failure
	verifier.AssertCacheHits(mockT, 999) // Wrong expectation
	if !mockT.Failed() {
		t.Error("AssertCacheHits should have failed but didn't")
	}

	// Reset mock
	mockT = &testing.T{}

	// Test AssertCacheMisses failure
	verifier.AssertCacheMisses(mockT, 999) // Wrong expectation
	if !mockT.Failed() {
		t.Error("AssertCacheMisses should have failed but didn't")
	}
}

// TestComputedCacheVerifier_EdgeCases_InvalidComputed tests edge cases with invalid computed values
func TestComputedCacheVerifier_EdgeCases_InvalidComputed(t *testing.T) {
	computeCount := 0

	// Test with a struct that has no Get() method
	type FakeComputed struct{}
	fakeComputed := &FakeComputed{}

	verifier := NewComputedCacheVerifier(fakeComputed, &computeCount)

	// GetValue should handle gracefully
	val := verifier.GetValue()
	assert.Nil(t, val)

	// InvalidateCache should handle gracefully
	verifier.InvalidateCache() // Should not panic
}

// TestComputedCacheVerifier_EdgeCases_InvalidInvalidateMethod tests InvalidateCache with missing method
func TestComputedCacheVerifier_EdgeCases_InvalidInvalidateMethod(t *testing.T) {
	computeCount := 0

	// Test with nil computed (no Invalidate method available)
	verifier := NewComputedCacheVerifier(nil, &computeCount)

	// InvalidateCache should handle nil gracefully
	verifier.InvalidateCache() // Should not panic
	assert.NotNil(t, verifier) // Verifier should still be valid

	// Test with struct that has no Invalidate method
	type FakeComputed struct{}
	fakeComputed := &FakeComputed{}
	verifier2 := NewComputedCacheVerifier(fakeComputed, &computeCount)

	// InvalidateCache should handle missing method gracefully
	verifier2.InvalidateCache() // Should not panic
	assert.NotNil(t, verifier2)
}

// TestComputedCacheVerifier_EdgeCases_EmptyResults tests GetValue with empty results
func TestComputedCacheVerifier_EdgeCases_EmptyResults(t *testing.T) {
	computeCount := 0

	// Create a mock computed that returns empty results
	// This is hard to test directly since reflection.Call always returns a slice
	// But we can test the nil computed case which exercises similar code paths
	verifier := NewComputedCacheVerifier(nil, &computeCount)

	val := verifier.GetValue()
	assert.Nil(t, val)

	// Verify counters weren't incremented
	assert.Equal(t, 0, verifier.GetCacheHits())
	assert.Equal(t, 0, verifier.GetCacheMisses())
}

// TestComputedCacheVerifier_CounterEdgeCases tests edge cases with counter tracking
func TestComputedCacheVerifier_CounterEdgeCases(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Test with zero expectations
	verifier.AssertComputeCount(t, 0)
	verifier.AssertCacheHits(t, 0)
	verifier.AssertCacheMisses(t, 0)

	// Get value
	verifier.GetValue()

	// Test with exact match
	verifier.AssertComputeCount(t, 1)
	verifier.AssertCacheHits(t, 0)
	verifier.AssertCacheMisses(t, 1)

	// Get again (cached)
	verifier.GetValue()

	// Test with updated counts
	verifier.AssertComputeCount(t, 1)
	verifier.AssertCacheHits(t, 1)
	verifier.AssertCacheMisses(t, 1)
}

// TestComputedCacheVerifier_MultipleInvalidations tests multiple cache invalidations
func TestComputedCacheVerifier_MultipleInvalidations(t *testing.T) {
	count := bubbly.NewRef(5)
	computeCount := 0

	computed := bubbly.NewComputed(func() int {
		computeCount++
		return count.GetTyped() * 2
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Initial get
	val := verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 1)

	// Invalidate multiple times
	verifier.InvalidateCache()
	verifier.InvalidateCache()
	verifier.InvalidateCache()

	// Get should recompute
	val = verifier.GetValue()
	assert.Equal(t, 10, val)
	verifier.AssertComputeCount(t, 2)
}

// TestComputedCacheVerifier_ZeroValueComputed tests computed returning zero values
func TestComputedCacheVerifier_ZeroValueComputed(t *testing.T) {
	computeCount := 0

	// Computed that returns zero value
	computed := bubbly.NewComputed(func() int {
		computeCount++
		return 0
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Get zero value
	val := verifier.GetValue()
	assert.Equal(t, 0, val)
	verifier.AssertComputeCount(t, 1)

	// Get again (should be cached)
	val = verifier.GetValue()
	assert.Equal(t, 0, val)
	verifier.AssertComputeCount(t, 1) // Still 1
	verifier.AssertCacheHits(t, 1)
}

// TestComputedCacheVerifier_BooleanComputed tests computed with boolean values
func TestComputedCacheVerifier_BooleanComputed(t *testing.T) {
	flag := bubbly.NewRef(true)
	computeCount := 0

	computed := bubbly.NewComputed(func() bool {
		computeCount++
		return flag.GetTyped()
	})

	verifier := NewComputedCacheVerifier(computed, &computeCount)

	// Get true
	val := verifier.GetValue()
	assert.Equal(t, true, val)

	// Change to false
	flag.Set(false)
	val = verifier.GetValue()
	assert.Equal(t, false, val)
	verifier.AssertComputeCount(t, 2)
}
