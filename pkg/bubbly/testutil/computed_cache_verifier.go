package testutil

import (
	"reflect"
	"testing"
)

// ComputedCacheVerifier provides utilities for testing computed value caching and invalidation.
// It wraps a Computed value and tracks computation counts, cache hits, and cache misses
// to verify that caching works correctly and computations are minimized.
//
// Key Features:
//   - Track compute function invocations
//   - Track cache hits (value returned from cache)
//   - Track cache misses (value recomputed)
//   - Manual cache invalidation for testing
//   - Verify caching behavior with assertions
//
// Caching Behavior:
//   - First Get() triggers computation (cache miss)
//   - Subsequent Get() calls return cached value (cache hits)
//   - Dependency changes invalidate cache
//   - Manual invalidation forces recomputation
//
// Example:
//
//	count := bubbly.NewRef(5)
//	computeCount := 0
//	computed := bubbly.NewComputed(func() int {
//	    computeCount++
//	    return count.GetTyped() * 2
//	})
//
//	verifier := NewComputedCacheVerifier(computed, &computeCount)
//
//	// First get - computes
//	val := verifier.GetValue()
//	verifier.AssertComputeCount(t, 1)
//	verifier.AssertCacheHits(t, 0)
//
//	// Second get - cached
//	val = verifier.GetValue()
//	verifier.AssertComputeCount(t, 1) // Still 1
//	verifier.AssertCacheHits(t, 1)    // Now 1
//
//	// Invalidate and get - recomputes
//	count.Set(10)
//	val = verifier.GetValue()
//	verifier.AssertComputeCount(t, 2) // Recomputed
//
// Thread Safety:
//
// ComputedCacheVerifier is not thread-safe. It should only be used from a single test goroutine.
type ComputedCacheVerifier struct {
	computed     interface{} // The computed value being tested (*Computed[T])
	computeCount *int        // Pointer to external compute counter
	cacheHits    int         // Number of cache hits
	cacheMisses  int         // Number of cache misses
	lastValue    interface{} // Last value returned
}

// NewComputedCacheVerifier creates a new ComputedCacheVerifier for testing computed value caching.
//
// The verifier requires an external counter that is incremented inside the compute function.
// This allows tracking how many times the compute function is actually called.
//
// Parameters:
//   - computed: The Computed value to test (*Computed[T])
//   - computeCount: Pointer to counter incremented in compute function
//
// Returns:
//   - *ComputedCacheVerifier: A new verifier instance
//
// Example:
//
//	count := bubbly.NewRef(5)
//	computeCount := 0
//	computed := bubbly.NewComputed(func() int {
//	    computeCount++  // Increment counter
//	    return count.GetTyped() * 2
//	})
//
//	verifier := NewComputedCacheVerifier(computed, &computeCount)
func NewComputedCacheVerifier(computed interface{}, computeCount *int) *ComputedCacheVerifier {
	return &ComputedCacheVerifier{
		computed:     computed,
		computeCount: computeCount,
		cacheHits:    0,
		cacheMisses:  0,
		lastValue:    nil,
	}
}

// GetValue retrieves the computed value and tracks whether it was a cache hit or miss.
//
// This method calls Get() on the computed value and determines if the compute function
// was called by checking if the compute counter increased.
//
// Returns:
//   - interface{}: The computed value
//
// Example:
//
//	val := verifier.GetValue()
//	assert.Equal(t, 10, val)
func (ccv *ComputedCacheVerifier) GetValue() interface{} {
	// Record compute count before Get()
	countBefore := *ccv.computeCount

	// Call Get() on the computed value using reflection
	v := reflect.ValueOf(ccv.computed)
	if !v.IsValid() || v.IsNil() {
		return nil
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return nil
	}

	results := getMethod.Call(nil)
	if len(results) == 0 {
		return nil
	}

	value := results[0].Interface()

	// Check if compute count increased (cache miss) or stayed same (cache hit)
	countAfter := *ccv.computeCount
	if countAfter > countBefore {
		ccv.cacheMisses++
	} else {
		ccv.cacheHits++
	}

	ccv.lastValue = value
	return value
}

// AssertComputeCount asserts that the compute function was called exactly the expected number of times.
//
// This verifies that caching is working correctly by ensuring the compute function
// is not called more than necessary.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: Expected number of compute function calls
//
// Example:
//
//	verifier.GetValue()
//	verifier.GetValue()
//	verifier.AssertComputeCount(t, 1) // Should only compute once
func (ccv *ComputedCacheVerifier) AssertComputeCount(t *testing.T, expected int) {
	t.Helper()
	actual := *ccv.computeCount
	if actual != expected {
		t.Errorf("Compute count: expected %d, got %d", expected, actual)
	}
}

// AssertCacheHits asserts that the expected number of cache hits occurred.
//
// A cache hit means Get() returned a cached value without calling the compute function.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: Expected number of cache hits
//
// Example:
//
//	verifier.GetValue() // Miss
//	verifier.GetValue() // Hit
//	verifier.GetValue() // Hit
//	verifier.AssertCacheHits(t, 2)
func (ccv *ComputedCacheVerifier) AssertCacheHits(t *testing.T, expected int) {
	t.Helper()
	if ccv.cacheHits != expected {
		t.Errorf("Cache hits: expected %d, got %d", expected, ccv.cacheHits)
	}
}

// AssertCacheMisses asserts that the expected number of cache misses occurred.
//
// A cache miss means Get() had to call the compute function to calculate the value.
//
// Parameters:
//   - t: The testing.T instance
//   - expected: Expected number of cache misses
//
// Example:
//
//	verifier.GetValue() // Miss
//	verifier.GetValue() // Hit
//	verifier.AssertCacheMisses(t, 1)
func (ccv *ComputedCacheVerifier) AssertCacheMisses(t *testing.T, expected int) {
	t.Helper()
	if ccv.cacheMisses != expected {
		t.Errorf("Cache misses: expected %d, got %d", expected, ccv.cacheMisses)
	}
}

// InvalidateCache manually invalidates the computed value's cache.
//
// This forces the next Get() call to recompute the value, which is useful
// for testing cache invalidation behavior.
//
// Example:
//
//	verifier.GetValue()
//	verifier.InvalidateCache()
//	verifier.GetValue() // Will recompute
func (ccv *ComputedCacheVerifier) InvalidateCache() {
	// Call Invalidate() method on the computed value using reflection
	v := reflect.ValueOf(ccv.computed)
	if !v.IsValid() || v.IsNil() {
		return
	}

	invalidateMethod := v.MethodByName("Invalidate")
	if !invalidateMethod.IsValid() {
		return
	}

	invalidateMethod.Call(nil)
}

// GetCacheHits returns the number of cache hits that have occurred.
//
// Returns:
//   - int: Number of cache hits
//
// Example:
//
//	hits := verifier.GetCacheHits()
//	assert.Equal(t, 2, hits)
func (ccv *ComputedCacheVerifier) GetCacheHits() int {
	return ccv.cacheHits
}

// GetCacheMisses returns the number of cache misses that have occurred.
//
// Returns:
//   - int: Number of cache misses
//
// Example:
//
//	misses := verifier.GetCacheMisses()
//	assert.Equal(t, 1, misses)
func (ccv *ComputedCacheVerifier) GetCacheMisses() int {
	return ccv.cacheMisses
}

// GetLastValue returns the last value retrieved from the computed value.
//
// Returns:
//   - interface{}: The last value
//
// Example:
//
//	val := verifier.GetLastValue()
//	assert.Equal(t, 10, val)
func (ccv *ComputedCacheVerifier) GetLastValue() interface{} {
	return ccv.lastValue
}

// ResetCounters resets the cache hit and miss counters to zero.
//
// This is useful for testing multiple scenarios in the same test.
//
// Example:
//
//	verifier.ResetCounters()
//	verifier.GetValue()
//	verifier.AssertCacheHits(t, 0)
func (ccv *ComputedCacheVerifier) ResetCounters() {
	ccv.cacheHits = 0
	ccv.cacheMisses = 0
}
