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
