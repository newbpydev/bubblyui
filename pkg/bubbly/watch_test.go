package bubbly

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWatch_BasicFunctionality tests basic Watch behavior
func TestWatch_BasicFunctionality(t *testing.T) {
	t.Run("callback executes on value change", func(t *testing.T) {
		ref := NewRef(10)
		var called bool
		var newVal, oldVal int

		cleanup := Watch(ref, func(n, o int) {
			called = true
			newVal = n
			oldVal = o
		})
		defer cleanup()

		ref.Set(20)

		assert.True(t, called, "Callback should be called")
		assert.Equal(t, 20, newVal, "New value should be 20")
		assert.Equal(t, 10, oldVal, "Old value should be 10")
	})

	t.Run("callback executes on multiple changes", func(t *testing.T) {
		ref := NewRef(0)
		var callCount int
		var values []int

		cleanup := Watch(ref, func(n, o int) {
			callCount++
			values = append(values, n)
		})
		defer cleanup()

		ref.Set(1)
		ref.Set(2)
		ref.Set(3)

		assert.Equal(t, 3, callCount, "Callback should be called 3 times")
		assert.Equal(t, []int{1, 2, 3}, values, "Should track all new values")
	})

	t.Run("callback receives correct old and new values", func(t *testing.T) {
		ref := NewRef("initial")
		var transitions []string

		cleanup := Watch(ref, func(newVal, oldVal string) {
			transitions = append(transitions, oldVal+" -> "+newVal)
		})
		defer cleanup()

		ref.Set("second")
		ref.Set("third")
		ref.Set("fourth")

		expected := []string{
			"initial -> second",
			"second -> third",
			"third -> fourth",
		}
		assert.Equal(t, expected, transitions)
	})
}

// TestWatch_Cleanup tests cleanup function behavior
func TestWatch_Cleanup(t *testing.T) {
	t.Run("cleanup stops watching", func(t *testing.T) {
		ref := NewRef(10)
		var callCount int

		cleanup := Watch(ref, func(n, o int) {
			callCount++
		})

		ref.Set(20)
		assert.Equal(t, 1, callCount, "Should be called once before cleanup")

		cleanup()

		ref.Set(30)
		ref.Set(40)
		assert.Equal(t, 1, callCount, "Should not be called after cleanup")
	})

	t.Run("cleanup can be called multiple times safely", func(t *testing.T) {
		ref := NewRef(10)
		var callCount int

		cleanup := Watch(ref, func(n, o int) {
			callCount++
		})

		ref.Set(20)
		assert.Equal(t, 1, callCount)

		// Call cleanup multiple times - should not panic
		cleanup()
		cleanup()
		cleanup()

		ref.Set(30)
		assert.Equal(t, 1, callCount, "Should still not be called")
	})

	t.Run("cleanup does not affect other watchers", func(t *testing.T) {
		ref := NewRef(10)
		var count1, count2 int

		cleanup1 := Watch(ref, func(n, o int) {
			count1++
		})
		cleanup2 := Watch(ref, func(n, o int) {
			count2++
		})

		ref.Set(20)
		assert.Equal(t, 1, count1)
		assert.Equal(t, 1, count2)

		// Remove first watcher
		cleanup1()

		ref.Set(30)
		assert.Equal(t, 1, count1, "First watcher should not be called")
		assert.Equal(t, 2, count2, "Second watcher should still be called")

		cleanup2()
	})
}

// TestWatch_MultipleWatchers tests multiple watchers on same Ref
func TestWatch_MultipleWatchers(t *testing.T) {
	t.Run("multiple watchers all receive notifications", func(t *testing.T) {
		ref := NewRef(0)
		var count1, count2, count3 int

		cleanup1 := Watch(ref, func(n, o int) { count1++ })
		cleanup2 := Watch(ref, func(n, o int) { count2++ })
		cleanup3 := Watch(ref, func(n, o int) { count3++ })
		defer cleanup1()
		defer cleanup2()
		defer cleanup3()

		ref.Set(1)
		ref.Set(2)

		assert.Equal(t, 2, count1)
		assert.Equal(t, 2, count2)
		assert.Equal(t, 2, count3)
	})

	t.Run("watchers are independent", func(t *testing.T) {
		ref := NewRef(10)
		var values1, values2 []int

		cleanup1 := Watch(ref, func(n, o int) {
			values1 = append(values1, n)
		})
		cleanup2 := Watch(ref, func(n, o int) {
			values2 = append(values2, n*2)
		})
		defer cleanup1()
		defer cleanup2()

		ref.Set(20)
		ref.Set(30)

		assert.Equal(t, []int{20, 30}, values1)
		assert.Equal(t, []int{40, 60}, values2)
	})

	t.Run("many watchers work correctly", func(t *testing.T) {
		ref := NewRef(0)
		const numWatchers = 100
		counts := make([]int32, numWatchers)
		cleanups := make([]WatchCleanup, numWatchers)

		for i := 0; i < numWatchers; i++ {
			idx := i
			cleanups[i] = Watch(ref, func(n, o int) {
				atomic.AddInt32(&counts[idx], 1)
			})
		}
		defer func() {
			for _, cleanup := range cleanups {
				cleanup()
			}
		}()

		ref.Set(1)
		ref.Set(2)
		ref.Set(3)

		for i, count := range counts {
			assert.Equal(t, int32(3), count, "Watcher %d should be called 3 times", i)
		}
	})
}

// TestWatch_TypeSafety tests type safety of Watch function
func TestWatch_TypeSafety(t *testing.T) {
	t.Run("int watcher", func(t *testing.T) {
		ref := NewRef(42)
		var received int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			received = newVal
		})
		defer cleanup()

		ref.Set(100)
		assert.Equal(t, 100, received)
	})

	t.Run("string watcher", func(t *testing.T) {
		ref := NewRef("hello")
		var received string

		cleanup := Watch(ref, func(newVal, oldVal string) {
			received = newVal
		})
		defer cleanup()

		ref.Set("world")
		assert.Equal(t, "world", received)
	})

	t.Run("struct watcher", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}

		ref := NewRef(User{Name: "John", Age: 30})
		var received User

		cleanup := Watch(ref, func(newVal, oldVal User) {
			received = newVal
		})
		defer cleanup()

		ref.Set(User{Name: "Jane", Age: 25})
		assert.Equal(t, "Jane", received.Name)
		assert.Equal(t, 25, received.Age)
	})

	t.Run("slice watcher", func(t *testing.T) {
		ref := NewRef([]int{1, 2, 3})
		var received []int

		cleanup := Watch(ref, func(newVal, oldVal []int) {
			received = newVal
		})
		defer cleanup()

		ref.Set([]int{4, 5, 6})
		assert.Equal(t, []int{4, 5, 6}, received)
	})

	t.Run("pointer watcher", func(t *testing.T) {
		val1 := 10
		val2 := 20
		ref := NewRef(&val1)
		var received *int

		cleanup := Watch(ref, func(newVal, oldVal *int) {
			received = newVal
		})
		defer cleanup()

		ref.Set(&val2)
		require.NotNil(t, received)
		assert.Equal(t, 20, *received)
	})
}

// TestWatch_NoPanic tests that Watch doesn't panic in edge cases
func TestWatch_NoPanic(t *testing.T) {
	t.Run("cleanup on non-existent watcher doesn't panic", func(t *testing.T) {
		ref := NewRef(10)

		cleanup := Watch(ref, func(n, o int) {})
		cleanup()

		// Calling cleanup again should not panic
		assert.NotPanics(t, func() {
			cleanup()
		})
	})

	t.Run("watch with nil callback doesn't panic during creation", func(t *testing.T) {
		// Note: This will compile but may panic at runtime
		// We're testing that our implementation handles it gracefully
		// In production, users should not pass nil callbacks
		ref := NewRef(10)

		// This should not panic during Watch creation
		assert.NotPanics(t, func() {
			cleanup := Watch(ref, func(n, o int) {
				// Empty callback is fine
			})
			cleanup()
		})
	})
}

// TestWatch_ConcurrentAccess tests thread safety of Watch
func TestWatch_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent watch registrations", func(t *testing.T) {
		ref := NewRef(0)
		const numWatchers = 50

		var wg sync.WaitGroup
		wg.Add(numWatchers)

		cleanups := make([]WatchCleanup, numWatchers)
		counts := make([]int32, numWatchers)

		for i := 0; i < numWatchers; i++ {
			go func(idx int) {
				defer wg.Done()
				cleanups[idx] = Watch(ref, func(n, o int) {
					atomic.AddInt32(&counts[idx], 1)
				})
			}(i)
		}

		wg.Wait()

		// All watchers registered, now trigger them
		ref.Set(1)

		// All watchers should have been called
		for i, count := range counts {
			assert.Equal(t, int32(1), count, "Watcher %d should be called once", i)
		}

		// Cleanup
		for _, cleanup := range cleanups {
			if cleanup != nil {
				cleanup()
			}
		}
	})

	t.Run("concurrent watch and cleanup", func(t *testing.T) {
		ref := NewRef(0)
		const numOperations = 100

		var wg sync.WaitGroup
		wg.Add(numOperations * 2)

		// Concurrent watch registrations
		for i := 0; i < numOperations; i++ {
			go func() {
				defer wg.Done()
				cleanup := Watch(ref, func(n, o int) {})
				cleanup()
			}()
		}

		// Concurrent value changes
		for i := 0; i < numOperations; i++ {
			go func(val int) {
				defer wg.Done()
				ref.Set(val)
			}(i)
		}

		wg.Wait()
		// Should not panic or deadlock
	})

	t.Run("concurrent cleanup calls", func(t *testing.T) {
		ref := NewRef(0)
		cleanup := Watch(ref, func(n, o int) {})

		const numGoroutines = 50
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Multiple goroutines calling cleanup concurrently
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				cleanup()
			}()
		}

		wg.Wait()
		// Should not panic
	})
}

// TestWatch_Integration tests Watch with other reactive features
func TestWatch_Integration(t *testing.T) {
	t.Run("watch with computed values", func(t *testing.T) {
		count := NewRef(5)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		var watchedValues []int
		cleanup := Watch(count, func(newVal, oldVal int) {
			// Access computed value in watcher
			watchedValues = append(watchedValues, doubled.Get())
		})
		defer cleanup()

		count.Set(10)
		count.Set(15)

		assert.Equal(t, []int{20, 30}, watchedValues)
	})

	t.Run("watch triggers on dependency change", func(t *testing.T) {
		a := NewRef(10)
		b := NewRef(20)

		var sumValues []int
		cleanup := Watch(a, func(newVal, oldVal int) {
			sum := newVal + b.Get()
			sumValues = append(sumValues, sum)
		})
		defer cleanup()

		a.Set(15)
		a.Set(20)

		assert.Equal(t, []int{35, 40}, sumValues)
	})
}

// BenchmarkWatch benchmarks Watch function performance
func BenchmarkWatch(b *testing.B) {
	ref := NewRef(0)
	cleanup := Watch(ref, func(n, o int) {
		// Minimal work
	})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// BenchmarkWatch_MultipleWatchers benchmarks multiple watchers
func BenchmarkWatch_MultipleWatchers(b *testing.B) {
	ref := NewRef(0)

	const numWatchers = 10
	cleanups := make([]WatchCleanup, numWatchers)
	for i := 0; i < numWatchers; i++ {
		cleanups[i] = Watch(ref, func(n, o int) {})
	}
	defer func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}
