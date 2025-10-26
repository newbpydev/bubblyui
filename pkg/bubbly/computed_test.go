package bubbly

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewComputed verifies that NewComputed creates a computed value with the given function
func TestNewComputed(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() int
		expected int
	}{
		{
			name:     "simple constant function",
			fn:       func() int { return 42 },
			expected: 42,
		},
		{
			name:     "zero value function",
			fn:       func() int { return 0 },
			expected: 0,
		},
		{
			name:     "computation function",
			fn:       func() int { return 10 * 5 },
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			computed := NewComputed(tt.fn)
			assert.NotNil(t, computed, "NewComputed should return non-nil computed value")
			// Note: Function should not be called yet (lazy evaluation)
		})
	}
}

// TestComputed_LazyEvaluation verifies that the function is not called until Get is invoked
func TestComputed_LazyEvaluation(t *testing.T) {
	t.Run("function not called on creation", func(t *testing.T) {
		var called bool
		fn := func() int {
			called = true
			return 42
		}

		computed := NewComputed(fn)
		assert.NotNil(t, computed)
		assert.False(t, called, "Function should not be called on creation")
	})

	t.Run("function called on first Get", func(t *testing.T) {
		var callCount int
		fn := func() int {
			callCount++
			return 42
		}

		computed := NewComputed(fn)
		assert.Equal(t, 0, callCount, "Function should not be called yet")

		value := computed.Get()
		assert.Equal(t, 42, value, "Should return computed value")
		assert.Equal(t, 1, callCount, "Function should be called exactly once")
	})
}

// TestComputed_Caching verifies that the function is only called once and result is cached
func TestComputed_Caching(t *testing.T) {
	t.Run("function called only once", func(t *testing.T) {
		var callCount int
		fn := func() int {
			callCount++
			return 100
		}

		computed := NewComputed(fn)

		// Call Get multiple times
		value1 := computed.Get()
		value2 := computed.Get()
		value3 := computed.Get()

		assert.Equal(t, 100, value1)
		assert.Equal(t, 100, value2)
		assert.Equal(t, 100, value3)
		assert.Equal(t, 1, callCount, "Function should be called exactly once")
	})

	t.Run("cached value is consistent", func(t *testing.T) {
		counter := 0
		fn := func() int {
			counter++
			return counter * 10
		}

		computed := NewComputed(fn)

		// First call should compute and cache
		first := computed.Get()
		assert.Equal(t, 10, first, "First call should compute value")

		// Subsequent calls should return cached value
		for i := 0; i < 10; i++ {
			value := computed.Get()
			assert.Equal(t, 10, value, "Should always return cached value")
		}
	})
}

// TestComputed_TypeSafety verifies compile-time type safety with different types
func TestComputed_TypeSafety(t *testing.T) {
	t.Run("int computed", func(t *testing.T) {
		computed := NewComputed(func() int { return 42 })
		value := computed.Get()
		assert.Equal(t, 42, value)
	})

	t.Run("string computed", func(t *testing.T) {
		computed := NewComputed(func() string { return "hello" })
		value := computed.Get()
		assert.Equal(t, "hello", value)
	})

	t.Run("bool computed", func(t *testing.T) {
		computed := NewComputed(func() bool { return true })
		value := computed.Get()
		assert.True(t, value)
	})

	t.Run("struct computed", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		computed := NewComputed(func() User {
			return User{Name: "John", Age: 30}
		})
		value := computed.Get()
		assert.Equal(t, "John", value.Name)
		assert.Equal(t, 30, value.Age)
	})

	t.Run("slice computed", func(t *testing.T) {
		computed := NewComputed(func() []int {
			return []int{1, 2, 3}
		})
		value := computed.Get()
		assert.Equal(t, []int{1, 2, 3}, value)
	})
}

// TestComputed_WithRef verifies computed values can use Ref values
func TestComputed_WithRef(t *testing.T) {
	t.Run("computed depends on ref", func(t *testing.T) {
		count := NewRef(5)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		value := doubled.Get()
		assert.Equal(t, 10, value, "Should compute based on ref value")
	})

	t.Run("computed with multiple refs", func(t *testing.T) {
		a := NewRef(10)
		b := NewRef(20)
		sum := NewComputed(func() int {
			return a.Get() + b.Get()
		})

		value := sum.Get()
		assert.Equal(t, 30, value, "Should compute sum of refs")
	})
}

// TestComputed_ComplexComputations verifies computed works with complex logic
func TestComputed_ComplexComputations(t *testing.T) {
	t.Run("filtering logic", func(t *testing.T) {
		items := NewRef([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		evens := NewComputed(func() []int {
			result := []int{}
			for _, n := range items.Get() {
				if n%2 == 0 {
					result = append(result, n)
				}
			}
			return result
		})

		value := evens.Get()
		assert.Equal(t, []int{2, 4, 6, 8, 10}, value)
	})

	t.Run("string transformation", func(t *testing.T) {
		name := NewRef("john")
		uppercase := NewComputed(func() string {
			s := name.Get()
			return string([]byte{s[0] - 32}) + s[1:] // Simple uppercase first letter
		})

		value := uppercase.Get()
		assert.Equal(t, "John", value)
	})
}

// TestComputed_ConcurrentGet verifies concurrent Get operations are safe
func TestComputed_ConcurrentGet(t *testing.T) {
	t.Run("multiple concurrent readers", func(t *testing.T) {
		var callCount int32
		computed := NewComputed(func() int {
			atomic.AddInt32(&callCount, 1)
			return 42
		})

		const numReaders = 100
		var wg sync.WaitGroup
		wg.Add(numReaders)

		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				value := computed.Get()
				assert.Equal(t, 42, value)
			}()
		}

		wg.Wait()

		// Function should be called exactly once despite concurrent access
		count := atomic.LoadInt32(&callCount)
		assert.Equal(t, int32(1), count, "Function should be called exactly once")
	})

	t.Run("concurrent reads return consistent value", func(t *testing.T) {
		computed := NewComputed(func() string {
			return "concurrent"
		})

		const numReaders = 50
		var wg sync.WaitGroup
		wg.Add(numReaders)

		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				value := computed.Get()
				assert.Equal(t, "concurrent", value)
			}()
		}

		wg.Wait()
	})
}

// TestComputed_ChainedComputed verifies computed can depend on other computed values
func TestComputed_ChainedComputed(t *testing.T) {
	t.Run("computed depends on another computed", func(t *testing.T) {
		count := NewRef(5)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})
		quadrupled := NewComputed(func() int {
			return doubled.Get() * 2
		})

		value := quadrupled.Get()
		assert.Equal(t, 20, value, "Should compute chained value: 5 * 2 * 2 = 20")
	})

	t.Run("multiple computed dependencies", func(t *testing.T) {
		a := NewRef(10)
		b := NewRef(20)
		sum := NewComputed(func() int {
			return a.Get() + b.Get()
		})
		product := NewComputed(func() int {
			return a.Get() * b.Get()
		})
		combined := NewComputed(func() int {
			return sum.Get() + product.Get()
		})

		value := combined.Get()
		assert.Equal(t, 230, value, "Should compute: (10+20) + (10*20) = 230")
	})
}

// TestComputed_ZeroValues verifies handling of zero values
func TestComputed_ZeroValues(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "zero int",
			test: func(t *testing.T) {
				computed := NewComputed(func() int { return 0 })
				assert.Equal(t, 0, computed.Get())
			},
		},
		{
			name: "empty string",
			test: func(t *testing.T) {
				computed := NewComputed(func() string { return "" })
				assert.Equal(t, "", computed.Get())
			},
		},
		{
			name: "false bool",
			test: func(t *testing.T) {
				computed := NewComputed(func() bool { return false })
				assert.False(t, computed.Get())
			},
		},
		{
			name: "nil pointer",
			test: func(t *testing.T) {
				computed := NewComputed(func() *int { return nil })
				assert.Nil(t, computed.Get())
			},
		},
		{
			name: "nil slice",
			test: func(t *testing.T) {
				computed := NewComputed(func() []int { return nil })
				assert.Nil(t, computed.Get())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// BenchmarkComputed_Get benchmarks Get operation on cached value
func BenchmarkComputed_Get(b *testing.B) {
	computed := NewComputed(func() int { return 42 })
	// Prime the cache
	_ = computed.Get()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computed.Get()
	}
}

// BenchmarkComputed_FirstGet benchmarks first Get with computation
func BenchmarkComputed_FirstGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		computed := NewComputed(func() int { return 42 })
		_ = computed.Get()
	}
}

// BenchmarkComputed_ConcurrentGet benchmarks concurrent Get operations
func BenchmarkComputed_ConcurrentGet(b *testing.B) {
	computed := NewComputed(func() int { return 42 })
	// Prime the cache
	_ = computed.Get()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = computed.Get()
		}
	})
}

// BenchmarkComputed_ComplexComputation benchmarks a more complex computation
func BenchmarkComputed_ComplexComputation(b *testing.B) {
	items := NewRef([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	for i := 0; i < b.N; i++ {
		computed := NewComputed(func() int {
			sum := 0
			for _, n := range items.Get() {
				sum += n * n
			}
			return sum
		})
		_ = computed.Get()
	}
}

// TestComputed_DependencyTracking verifies automatic dependency tracking
func TestComputed_DependencyTracking(t *testing.T) {
	t.Run("computed tracks ref dependency", func(t *testing.T) {
		count := NewRef(5)
		var callCount int
		doubled := NewComputed(func() int {
			callCount++
			return count.Get() * 2
		})

		// First access - should compute
		value1 := doubled.Get()
		assert.Equal(t, 10, value1)
		assert.Equal(t, 1, callCount, "Should compute on first access")

		// Second access - should use cache
		value2 := doubled.Get()
		assert.Equal(t, 10, value2)
		assert.Equal(t, 1, callCount, "Should use cache on second access")

		// Change ref value - should invalidate cache
		count.Set(10)

		// Next access - should recompute
		value3 := doubled.Get()
		assert.Equal(t, 20, value3)
		assert.Equal(t, 2, callCount, "Should recompute after dependency change")
	})

	t.Run("computed tracks multiple ref dependencies", func(t *testing.T) {
		a := NewRef(10)
		b := NewRef(20)
		var callCount int
		sum := NewComputed(func() int {
			callCount++
			return a.Get() + b.Get()
		})

		// Initial computation
		value1 := sum.Get()
		assert.Equal(t, 30, value1)
		assert.Equal(t, 1, callCount)

		// Change first ref
		a.Set(15)
		value2 := sum.Get()
		assert.Equal(t, 35, value2)
		assert.Equal(t, 2, callCount)

		// Change second ref
		b.Set(25)
		value3 := sum.Get()
		assert.Equal(t, 40, value3)
		assert.Equal(t, 3, callCount)
	})

	t.Run("chained computed values", func(t *testing.T) {
		count := NewRef(5)
		var doubledCalls, quadrupledCalls int

		doubled := NewComputed(func() int {
			doubledCalls++
			return count.Get() * 2
		})

		quadrupled := NewComputed(func() int {
			quadrupledCalls++
			return doubled.Get() * 2
		})

		// Initial computation
		value1 := quadrupled.Get()
		assert.Equal(t, 20, value1)
		assert.Equal(t, 1, doubledCalls)
		assert.Equal(t, 1, quadrupledCalls)

		// Change base ref - should invalidate entire chain
		count.Set(10)

		value2 := quadrupled.Get()
		assert.Equal(t, 40, value2)
		assert.Equal(t, 2, doubledCalls, "Doubled should recompute")
		assert.Equal(t, 2, quadrupledCalls, "Quadrupled should recompute")
	})

	t.Run("computed with no dependencies", func(t *testing.T) {
		var callCount int
		constant := NewComputed(func() int {
			callCount++
			return 42
		})

		// Should compute once and cache forever
		for i := 0; i < 10; i++ {
			value := constant.Get()
			assert.Equal(t, 42, value)
		}
		assert.Equal(t, 1, callCount, "Should compute only once")
	})
}

// TestComputed_CacheInvalidation verifies cache invalidation behavior
func TestComputed_CacheInvalidation(t *testing.T) {
	t.Run("invalidation propagates through chain", func(t *testing.T) {
		a := NewRef(1)
		b := NewComputed(func() int { return a.Get() * 2 })
		c := NewComputed(func() int { return b.Get() * 2 })
		d := NewComputed(func() int { return c.Get() * 2 })

		// Initial values: a=1, b=2, c=4, d=8
		assert.Equal(t, 8, d.Get())

		// Change a - should invalidate b, c, d
		a.Set(2)

		// New values: a=2, b=4, c=8, d=16
		assert.Equal(t, 16, d.Get())
	})

	t.Run("diamond dependency pattern", func(t *testing.T) {
		//     a
		//    / \
		//   b   c
		//    \ /
		//     d
		a := NewRef(10)
		var bCalls, cCalls, dCalls int

		b := NewComputed(func() int {
			bCalls++
			return a.Get() + 5
		})

		c := NewComputed(func() int {
			cCalls++
			return a.Get() * 2
		})

		d := NewComputed(func() int {
			dCalls++
			return b.Get() + c.Get()
		})

		// Initial: b=15, c=20, d=35
		value1 := d.Get()
		assert.Equal(t, 35, value1)
		assert.Equal(t, 1, bCalls)
		assert.Equal(t, 1, cCalls)
		assert.Equal(t, 1, dCalls)

		// Change a - should invalidate b, c, d
		a.Set(20)

		// New: b=25, c=40, d=65
		value2 := d.Get()
		assert.Equal(t, 65, value2)
		assert.Equal(t, 2, bCalls)
		assert.Equal(t, 2, cCalls)
		assert.Equal(t, 2, dCalls)
	})

	t.Run("selective invalidation", func(t *testing.T) {
		a := NewRef(10)
		b := NewRef(20)
		var sumCalls, productCalls int

		sum := NewComputed(func() int {
			sumCalls++
			return a.Get() + b.Get()
		})

		product := NewComputed(func() int {
			productCalls++
			return a.Get() * b.Get()
		})

		// Initial computation
		_ = sum.Get()
		_ = product.Get()
		assert.Equal(t, 1, sumCalls)
		assert.Equal(t, 1, productCalls)

		// Change only a - both should recompute
		a.Set(15)
		_ = sum.Get()
		_ = product.Get()
		assert.Equal(t, 2, sumCalls)
		assert.Equal(t, 2, productCalls)
	})
}

// TestComputed_CircularDependency verifies circular dependency detection
func TestComputed_CircularDependency(t *testing.T) {
	t.Run("circular dependency detected by tracker", func(t *testing.T) {
		// Test that the tracker detects circular dependencies
		tracker := &DepTracker{}

		dep1 := newMockDependency("dep1")
		dep2 := newMockDependency("dep2")

		// Start tracking dep1
		err := tracker.BeginTracking(dep1)
		assert.NoError(t, err)

		// Start tracking dep2 (nested)
		err = tracker.BeginTracking(dep2)
		assert.NoError(t, err)

		// Try to track dep1 again - should detect circular dependency
		err = tracker.BeginTracking(dep1)
		assert.ErrorIs(t, err, ErrCircularDependency, "Should detect circular dependency")

		// Clean up
		tracker.EndTracking() // dep2
		tracker.EndTracking() // dep1
	})

	t.Run("max depth prevents infinite recursion", func(t *testing.T) {
		// Verify max depth limit prevents stack overflow
		tracker := &DepTracker{}

		deps := make([]*mockDependency, MaxDependencyDepth+1)
		for i := range deps {
			deps[i] = newMockDependency("dep")
		}

		// Fill to max depth
		for i := 0; i < MaxDependencyDepth; i++ {
			err := tracker.BeginTracking(deps[i])
			assert.NoError(t, err)
		}

		// Exceeding max depth should error
		err := tracker.BeginTracking(deps[MaxDependencyDepth])
		assert.ErrorIs(t, err, ErrMaxDepthExceeded)

		// Clean up
		for i := 0; i < MaxDependencyDepth; i++ {
			tracker.EndTracking()
		}
	})
}

// TestComputed_ConcurrentInvalidation verifies thread safety during invalidation
func TestComputed_ConcurrentInvalidation(t *testing.T) {
	count := NewRef(0)
	var callCount int32
	computed := NewComputed(func() int {
		atomic.AddInt32(&callCount, 1)
		return count.Get() * 2
	})

	// Prime the cache
	_ = computed.Get()
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrently invalidate and read
	for i := 0; i < numGoroutines; i++ {
		go func(val int) {
			defer wg.Done()
			count.Set(val)
			_ = computed.Get()
		}(i)
	}

	wg.Wait()

	// Verify no race conditions occurred
	// The exact call count is non-deterministic due to concurrent access
	// but should be > 1 and <= numGoroutines + 1
	finalCount := atomic.LoadInt32(&callCount)
	assert.Greater(t, finalCount, int32(1), "Should have recomputed at least once")
	assert.LessOrEqual(t, finalCount, int32(numGoroutines+1), "Should not compute more than necessary")
}
