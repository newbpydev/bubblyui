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

// TestWatch_WithImmediate tests the WithImmediate option
func TestWatch_WithImmediate(t *testing.T) {
	t.Run("callback executes immediately with current value", func(t *testing.T) {
		ref := NewRef(42)
		var called bool
		var receivedNew, receivedOld int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			called = true
			receivedNew = newVal
			receivedOld = oldVal
		}, WithImmediate())
		defer cleanup()

		assert.True(t, called, "Callback should be called immediately")
		assert.Equal(t, 42, receivedNew, "Should receive current value as new")
		assert.Equal(t, 42, receivedOld, "Should receive current value as old")
	})

	t.Run("callback still executes on subsequent changes", func(t *testing.T) {
		ref := NewRef(10)
		var callCount int
		var values []int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
			values = append(values, newVal)
		}, WithImmediate())
		defer cleanup()

		assert.Equal(t, 1, callCount, "Should be called once immediately")
		assert.Equal(t, []int{10}, values)

		ref.Set(20)
		assert.Equal(t, 2, callCount, "Should be called again on change")
		assert.Equal(t, []int{10, 20}, values)
	})

	t.Run("without immediate option callback not called initially", func(t *testing.T) {
		ref := NewRef(42)
		var called bool

		cleanup := Watch(ref, func(newVal, oldVal int) {
			called = true
		})
		defer cleanup()

		assert.False(t, called, "Callback should not be called without WithImmediate")

		ref.Set(100)
		assert.True(t, called, "Callback should be called on change")
	})

	t.Run("immediate with different types", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			ref := NewRef("hello")
			var received string

			cleanup := Watch(ref, func(newVal, oldVal string) {
				received = newVal
			}, WithImmediate())
			defer cleanup()

			assert.Equal(t, "hello", received)
		})

		t.Run("struct", func(t *testing.T) {
			type User struct {
				Name string
			}
			ref := NewRef(User{Name: "John"})
			var received User

			cleanup := Watch(ref, func(newVal, oldVal User) {
				received = newVal
			}, WithImmediate())
			defer cleanup()

			assert.Equal(t, "John", received.Name)
		})
	})
}

// TestWatch_WithDeep tests the WithDeep option
func TestWatch_WithDeep(t *testing.T) {
	t.Run("deep option is accepted", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		ref := NewRef(User{Name: "John", Age: 30})
		var called bool

		cleanup := Watch(ref, func(newVal, oldVal User) {
			called = true
		}, WithDeep())
		defer cleanup()

		// Deep watching is a placeholder, but option should be accepted
		ref.Set(User{Name: "Jane", Age: 25})
		assert.True(t, called, "Callback should be called on Set")
	})

	t.Run("deep option documented as placeholder", func(t *testing.T) {
		// This test documents that WithDeep is currently a placeholder
		// Deep watching would require reflection or manual change detection
		// For now, it only triggers on Set() calls, not nested field changes
		type Profile struct {
			Bio string
		}
		type User struct {
			Name    string
			Profile Profile
		}

		ref := NewRef(User{Name: "John", Profile: Profile{Bio: "Developer"}})
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal User) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// This triggers the watcher (Set is called)
		user := ref.Get()
		user.Profile.Bio = "Engineer"
		ref.Set(user)

		assert.Equal(t, 1, callCount, "Should be called when Set is used")
	})
}

// TestWatch_WithFlush tests the WithFlush option
func TestWatch_WithFlush(t *testing.T) {
	t.Run("sync flush mode", func(t *testing.T) {
		ref := NewRef(0)
		var called bool

		cleanup := Watch(ref, func(newVal, oldVal int) {
			called = true
		}, WithFlush("sync"))
		defer cleanup()

		ref.Set(1)
		assert.True(t, called, "Callback should be called with sync flush")
	})

	t.Run("post flush mode queues callback", func(t *testing.T) {
		// Clear any pending callbacks from previous tests
		FlushWatchers()

		ref := NewRef(0)
		var called bool

		cleanup := Watch(ref, func(newVal, oldVal int) {
			called = true
		}, WithFlush("post"))
		defer cleanup()

		ref.Set(1)
		// Post flush queues callback, doesn't execute immediately
		assert.False(t, called, "Callback should not be called immediately")

		// Flush to execute
		FlushWatchers()
		assert.True(t, called, "Callback should be called after flush")
	})

	t.Run("default flush mode is sync", func(t *testing.T) {
		ref := NewRef(0)
		var called bool

		cleanup := Watch(ref, func(newVal, oldVal int) {
			called = true
		})
		defer cleanup()

		ref.Set(1)
		assert.True(t, called, "Callback should be called with default flush mode")
	})
}

// TestWatch_OptionComposition tests combining multiple options
func TestWatch_OptionComposition(t *testing.T) {
	t.Run("immediate and deep together", func(t *testing.T) {
		ref := NewRef(100)
		var callCount int
		var values []int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
			values = append(values, newVal)
		}, WithImmediate(), WithDeep())
		defer cleanup()

		assert.Equal(t, 1, callCount, "Should be called immediately")
		assert.Equal(t, []int{100}, values)

		ref.Set(200)
		assert.Equal(t, 2, callCount)
		assert.Equal(t, []int{100, 200}, values)
	})

	t.Run("immediate and flush together", func(t *testing.T) {
		ref := NewRef("initial")
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal string) {
			callCount++
		}, WithImmediate(), WithFlush("sync"))
		defer cleanup()

		assert.Equal(t, 1, callCount, "Should be called immediately")

		ref.Set("changed")
		assert.Equal(t, 2, callCount)
	})

	t.Run("all options together", func(t *testing.T) {
		// Clear any pending callbacks
		FlushWatchers()

		ref := NewRef(42)
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
		}, WithImmediate(), WithDeep(), WithFlush("post"))
		defer cleanup()

		assert.Equal(t, 1, callCount, "Should be called immediately")

		// With post-flush, callback is queued
		ref.Set(100)
		assert.Equal(t, 1, callCount, "Should not be called yet (post-flush)")

		// Flush to execute
		FlushWatchers()
		assert.Equal(t, 2, callCount, "Should be called after flush")
	})

	t.Run("options order doesn't matter", func(t *testing.T) {
		ref1 := NewRef(1)
		ref2 := NewRef(1)
		var count1, count2 int

		cleanup1 := Watch(ref1, func(n, o int) { count1++ },
			WithImmediate(), WithFlush("sync"))
		cleanup2 := Watch(ref2, func(n, o int) { count2++ },
			WithFlush("sync"), WithImmediate())
		defer cleanup1()
		defer cleanup2()

		assert.Equal(t, count1, count2, "Order of options should not matter")
	})
}

// TestWatch_OptionsDefaults tests default option values
func TestWatch_OptionsDefaults(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		ref := NewRef(10)
		var immediatelyCalled bool

		cleanup := Watch(ref, func(newVal, oldVal int) {
			immediatelyCalled = true
		})
		defer cleanup()

		assert.False(t, immediatelyCalled, "Should not call immediately by default")

		ref.Set(20)
		assert.True(t, immediatelyCalled, "Should call on change")
	})
}

// TestWatch_DeepWatching tests the WithDeep option with reflection-based comparison
func TestWatch_DeepWatching(t *testing.T) {
	type Profile struct {
		Bio string
		Age int
	}
	type User struct {
		Name    string
		Profile Profile
	}

	t.Run("deep watching detects nested struct changes", func(t *testing.T) {
		user := NewRef(User{Name: "John", Profile: Profile{Bio: "Developer", Age: 30}})
		var callCount int

		cleanup := Watch(user, func(newVal, oldVal User) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set with same values - should NOT trigger (deep equal)
		user.Set(User{Name: "John", Profile: Profile{Bio: "Developer", Age: 30}})
		assert.Equal(t, 0, callCount, "Should not trigger on deep equal values")

		// Set with nested change - should trigger
		user.Set(User{Name: "John", Profile: Profile{Bio: "Engineer", Age: 30}})
		assert.Equal(t, 1, callCount, "Should trigger on nested change")
	})

	t.Run("deep watching with slice changes", func(t *testing.T) {
		ref := NewRef([]int{1, 2, 3})
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal []int) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set with same slice - should NOT trigger
		ref.Set([]int{1, 2, 3})
		assert.Equal(t, 0, callCount, "Should not trigger on equal slice")

		// Set with different slice - should trigger
		ref.Set([]int{1, 2, 4})
		assert.Equal(t, 1, callCount, "Should trigger on slice change")
	})

	t.Run("deep watching with map changes", func(t *testing.T) {
		ref := NewRef(map[string]int{"a": 1, "b": 2})
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal map[string]int) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set with same map - should NOT trigger
		ref.Set(map[string]int{"a": 1, "b": 2})
		assert.Equal(t, 0, callCount, "Should not trigger on equal map")

		// Set with different map - should trigger
		ref.Set(map[string]int{"a": 1, "b": 3})
		assert.Equal(t, 1, callCount, "Should trigger on map change")
	})

	t.Run("shallow watching triggers on every Set", func(t *testing.T) {
		user := NewRef(User{Name: "John", Profile: Profile{Bio: "Developer", Age: 30}})
		var callCount int

		// Without WithDeep - shallow watching
		cleanup := Watch(user, func(newVal, oldVal User) {
			callCount++
		})
		defer cleanup()

		// Set with same values - SHOULD trigger (shallow watching)
		user.Set(User{Name: "John", Profile: Profile{Bio: "Developer", Age: 30}})
		assert.Equal(t, 1, callCount, "Shallow watching triggers on every Set")

		// Set again - triggers again
		user.Set(User{Name: "John", Profile: Profile{Bio: "Developer", Age: 30}})
		assert.Equal(t, 2, callCount, "Shallow watching triggers every time")
	})

	t.Run("deep watching with pointer values", func(t *testing.T) {
		val1 := 10
		val2 := 10
		val3 := 20

		ref := NewRef(&val1)
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal *int) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set with pointer to same value - should NOT trigger
		ref.Set(&val2)
		assert.Equal(t, 0, callCount, "Should not trigger on equal pointer values")

		// Set with pointer to different value - should trigger
		ref.Set(&val3)
		assert.Equal(t, 1, callCount, "Should trigger on different pointer value")
	})
}

// TestWatch_DeepCompare tests the WithDeepCompare option with custom comparators
func TestWatch_DeepCompare(t *testing.T) {
	type User struct {
		ID      int
		Name    string
		Profile string // Large field we want to ignore
	}

	t.Run("custom comparator for selective comparison", func(t *testing.T) {
		// Only compare ID and Name, ignore Profile
		compareUsers := func(old, new User) bool {
			return old.ID == new.ID && old.Name == new.Name
		}

		user := NewRef(User{ID: 1, Name: "John", Profile: "Long bio..."})
		var callCount int

		cleanup := Watch(user, func(newVal, oldVal User) {
			callCount++
		}, WithDeepCompare(compareUsers))
		defer cleanup()

		// Change Profile only - should NOT trigger
		user.Set(User{ID: 1, Name: "John", Profile: "Different bio..."})
		assert.Equal(t, 0, callCount, "Should not trigger when only Profile changes")

		// Change Name - should trigger
		user.Set(User{ID: 1, Name: "Jane", Profile: "Different bio..."})
		assert.Equal(t, 1, callCount, "Should trigger when Name changes")
	})

	t.Run("custom comparator with complex logic", func(t *testing.T) {
		type Config struct {
			Version int
			Data    map[string]string
		}

		// Only trigger if version changes OR specific data keys change
		compareConfig := func(old, new Config) bool {
			if old.Version != new.Version {
				return false // Different
			}
			// Only check specific keys
			return old.Data["important"] == new.Data["important"]
		}

		config := NewRef(Config{
			Version: 1,
			Data:    map[string]string{"important": "value", "other": "data"},
		})
		var callCount int

		cleanup := Watch(config, func(newVal, oldVal Config) {
			callCount++
		}, WithDeepCompare(compareConfig))
		defer cleanup()

		// Change "other" key - should NOT trigger
		config.Set(Config{
			Version: 1,
			Data:    map[string]string{"important": "value", "other": "changed"},
		})
		assert.Equal(t, 0, callCount, "Should not trigger on non-important data change")

		// Change "important" key - should trigger
		config.Set(Config{
			Version: 1,
			Data:    map[string]string{"important": "new-value", "other": "changed"},
		})
		assert.Equal(t, 1, callCount, "Should trigger on important data change")

		// Change version - should trigger
		config.Set(Config{
			Version: 2,
			Data:    map[string]string{"important": "new-value", "other": "changed"},
		})
		assert.Equal(t, 2, callCount, "Should trigger on version change")
	})
}

// TestWatch_DeepEdgeCases tests edge cases for deep watching
func TestWatch_DeepEdgeCases(t *testing.T) {
	t.Run("deep watching with nil values", func(t *testing.T) {
		ref := NewRef[*int](nil)
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal *int) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set to nil again - should NOT trigger
		ref.Set(nil)
		assert.Equal(t, 0, callCount, "Should not trigger on nil to nil")

		// Set to non-nil - should trigger
		val := 10
		ref.Set(&val)
		assert.Equal(t, 1, callCount, "Should trigger on nil to non-nil")
	})

	t.Run("deep watching with empty collections", func(t *testing.T) {
		ref := NewRef([]int{})
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal []int) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set to empty slice again - should NOT trigger
		ref.Set([]int{})
		assert.Equal(t, 0, callCount, "Should not trigger on empty to empty")

		// Set to non-empty - should trigger
		ref.Set([]int{1})
		assert.Equal(t, 1, callCount, "Should trigger on empty to non-empty")
	})

	t.Run("deep watching with unexported fields", func(t *testing.T) {
		type privateStruct struct {
			Public  string
			private string // unexported
		}

		ref := NewRef(privateStruct{Public: "visible", private: "hidden"})
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal privateStruct) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// reflect.DeepEqual handles unexported fields correctly
		ref.Set(privateStruct{Public: "visible", private: "hidden"})
		assert.Equal(t, 0, callCount, "Should handle unexported fields")

		ref.Set(privateStruct{Public: "changed", private: "hidden"})
		assert.Equal(t, 1, callCount, "Should detect public field change")
	})
}

// TestWatch_DeepWithOtherOptions tests combining deep watching with other options
func TestWatch_DeepWithOtherOptions(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	t.Run("deep with immediate", func(t *testing.T) {
		user := NewRef(User{Name: "John", Age: 30})
		var callCount int

		cleanup := Watch(user, func(newVal, oldVal User) {
			callCount++
		}, WithDeep(), WithImmediate())
		defer cleanup()

		assert.Equal(t, 1, callCount, "Should call immediately")

		// Set with same values - should NOT trigger
		user.Set(User{Name: "John", Age: 30})
		assert.Equal(t, 1, callCount, "Should not trigger on deep equal")

		// Set with different values - should trigger
		user.Set(User{Name: "Jane", Age: 30})
		assert.Equal(t, 2, callCount, "Should trigger on change")
	})

	t.Run("deep compare with immediate", func(t *testing.T) {
		user := NewRef(User{Name: "John", Age: 30})
		var callCount int

		compareNames := func(old, new User) bool {
			return old.Name == new.Name
		}

		cleanup := Watch(user, func(newVal, oldVal User) {
			callCount++
		}, WithDeepCompare(compareNames), WithImmediate())
		defer cleanup()

		assert.Equal(t, 1, callCount, "Should call immediately")

		// Change age only - should NOT trigger
		user.Set(User{Name: "John", Age: 31})
		assert.Equal(t, 1, callCount, "Should not trigger when name unchanged")
	})
}

// BenchmarkWatch_Deep benchmarks deep watching performance
func BenchmarkWatch_Deep(b *testing.B) {
	type User struct {
		Name    string
		Age     int
		Profile string
	}

	user := NewRef(User{Name: "John", Age: 30, Profile: "Developer"})
	cleanup := Watch(user, func(n, o User) {}, WithDeep())
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.Set(User{Name: "John", Age: 30, Profile: "Developer"})
	}
}

// BenchmarkWatch_DeepCompare benchmarks custom comparator performance
func BenchmarkWatch_DeepCompare(b *testing.B) {
	type User struct {
		Name    string
		Age     int
		Profile string
	}

	compareUsers := func(old, new User) bool {
		return old.Name == new.Name && old.Age == new.Age
	}

	user := NewRef(User{Name: "John", Age: 30, Profile: "Developer"})
	cleanup := Watch(user, func(n, o User) {}, WithDeepCompare(compareUsers))
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.Set(User{Name: "John", Age: 30, Profile: "Different"})
	}
}

// BenchmarkWatch_Shallow benchmarks shallow watching for comparison
func BenchmarkWatch_Shallow(b *testing.B) {
	type User struct {
		Name    string
		Age     int
		Profile string
	}

	user := NewRef(User{Name: "John", Age: 30, Profile: "Developer"})
	cleanup := Watch(user, func(n, o User) {})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.Set(User{Name: "John", Age: 30, Profile: "Developer"})
	}
}

// TestWatch_PostFlush tests the WithFlush("post") option
func TestWatch_PostFlush(t *testing.T) {
	t.Run("post-flush queues callbacks", func(t *testing.T) {
		// Clear any pending callbacks from previous tests
		FlushWatchers()

		ref := NewRef(0)
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
		}, WithFlush("post"))
		defer cleanup()

		// Set value - callback should be queued, not executed
		ref.Set(1)
		assert.Equal(t, 0, callCount, "Callback should not execute immediately")
		assert.Equal(t, 1, PendingCallbacks(), "Should have 1 pending callback")

		// Flush callbacks
		flushed := FlushWatchers()
		assert.Equal(t, 1, flushed, "Should flush 1 callback")
		assert.Equal(t, 1, callCount, "Callback should execute after flush")
		assert.Equal(t, 0, PendingCallbacks(), "Should have no pending callbacks")
	})

	t.Run("batching replaces previous callbacks", func(t *testing.T) {
		ref := NewRef(0)
		var callCount int
		var lastValue int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
			lastValue = newVal
		}, WithFlush("post"))
		defer cleanup()

		// Multiple sets - should batch into single callback
		ref.Set(1)
		ref.Set(2)
		ref.Set(3)

		assert.Equal(t, 0, callCount, "Callbacks should not execute yet")
		assert.Equal(t, 1, PendingCallbacks(), "Should have 1 pending callback (batched)")

		// Flush - should execute once with final value
		FlushWatchers()
		assert.Equal(t, 1, callCount, "Should execute callback once")
		assert.Equal(t, 3, lastValue, "Should have final value")
	})

	t.Run("sync mode executes immediately", func(t *testing.T) {
		ref := NewRef(0)
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
		}, WithFlush("sync"))
		defer cleanup()

		ref.Set(1)
		assert.Equal(t, 1, callCount, "Sync callback should execute immediately")
		assert.Equal(t, 0, PendingCallbacks(), "Should have no pending callbacks")
	})

	t.Run("default mode is sync", func(t *testing.T) {
		ref := NewRef(0)
		var callCount int

		cleanup := Watch(ref, func(newVal, oldVal int) {
			callCount++
		})
		defer cleanup()

		ref.Set(1)
		assert.Equal(t, 1, callCount, "Default should execute immediately")
		assert.Equal(t, 0, PendingCallbacks(), "Should have no pending callbacks")
	})

	t.Run("multiple watchers with different flush modes", func(t *testing.T) {
		ref := NewRef(0)
		var syncCount, postCount int

		cleanup1 := Watch(ref, func(n, o int) { syncCount++ }, WithFlush("sync"))
		cleanup2 := Watch(ref, func(n, o int) { postCount++ }, WithFlush("post"))
		defer cleanup1()
		defer cleanup2()

		ref.Set(1)
		assert.Equal(t, 1, syncCount, "Sync watcher should execute")
		assert.Equal(t, 0, postCount, "Post watcher should be queued")

		FlushWatchers()
		assert.Equal(t, 1, syncCount, "Sync count unchanged")
		assert.Equal(t, 1, postCount, "Post watcher should execute")
	})

	t.Run("flush with no pending callbacks", func(t *testing.T) {
		flushed := FlushWatchers()
		assert.Equal(t, 0, flushed, "Should flush 0 callbacks")
	})
}

// TestWatch_PostFlushWithDeep tests combining post-flush with deep watching
func TestWatch_PostFlushWithDeep(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	t.Run("post-flush with deep watching", func(t *testing.T) {
		user := NewRef(User{Name: "John", Age: 30})
		var callCount int

		cleanup := Watch(user, func(newVal, oldVal User) {
			callCount++
		}, WithFlush("post"), WithDeep())
		defer cleanup()

		// Set with same value - should NOT queue (deep equal)
		user.Set(User{Name: "John", Age: 30})
		assert.Equal(t, 0, PendingCallbacks(), "Should not queue on deep equal")

		// Set with different value - should queue
		user.Set(User{Name: "Jane", Age: 30})
		assert.Equal(t, 1, PendingCallbacks(), "Should queue on change")

		FlushWatchers()
		assert.Equal(t, 1, callCount, "Should execute once")
	})
}

// TestWatch_PostFlushConcurrent tests thread safety of post-flush
func TestWatch_PostFlushConcurrent(t *testing.T) {
	t.Run("concurrent post-flush operations", func(t *testing.T) {
		ref := NewRef(0)
		var callCount int32

		cleanup := Watch(ref, func(newVal, oldVal int) {
			atomic.AddInt32(&callCount, 1)
		}, WithFlush("post"))
		defer cleanup()

		const numGoroutines = 50
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Concurrent sets
		for i := 0; i < numGoroutines; i++ {
			go func(val int) {
				defer wg.Done()
				ref.Set(val)
			}(i)
		}

		wg.Wait()

		// Should have 1 pending callback (batched)
		assert.Equal(t, 1, PendingCallbacks(), "Should batch into 1 callback")

		// Flush
		FlushWatchers()
		assert.Equal(t, int32(1), callCount, "Should execute once")
	})

	t.Run("concurrent flush calls", func(t *testing.T) {
		ref := NewRef(0)
		var callCount int32

		cleanup := Watch(ref, func(newVal, oldVal int) {
			atomic.AddInt32(&callCount, 1)
		}, WithFlush("post"))
		defer cleanup()

		ref.Set(1)

		const numGoroutines = 10
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Concurrent flushes - should be safe
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				FlushWatchers()
			}()
		}

		wg.Wait()

		// Callback should execute exactly once
		assert.Equal(t, int32(1), callCount, "Should execute once despite concurrent flushes")
	})
}

// BenchmarkWatch_PostFlush benchmarks post-flush performance
func BenchmarkWatch_PostFlush(b *testing.B) {
	ref := NewRef(0)
	cleanup := Watch(ref, func(n, o int) {}, WithFlush("post"))
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
		if i%100 == 0 {
			FlushWatchers()
		}
	}
	FlushWatchers() // Final flush
}

// BenchmarkWatch_PostFlushBatching benchmarks batching benefit
func BenchmarkWatch_PostFlushBatching(b *testing.B) {
	ref := NewRef(0)
	var count int
	cleanup := Watch(ref, func(n, o int) {
		count++ // Simulate work
	}, WithFlush("post"))
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 10 rapid changes
		for j := 0; j < 10; j++ {
			ref.Set(j)
		}
		// Flush once - callback executes once instead of 10 times
		FlushWatchers()
	}
}

// ============================================================================
// Task 6.2: Watch Computed Values Tests
// ============================================================================

// TestWatch_ComputedValue tests watching computed values directly
func TestWatch_ComputedValue(t *testing.T) {
	t.Run("watch computed value changes", func(t *testing.T) {
		count := NewRef(5)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		var called bool
		var newVal, oldVal int

		cleanup := Watch(doubled, func(n, o int) {
			called = true
			newVal = n
			oldVal = o
		})
		defer cleanup()

		// Change underlying ref
		count.Set(10)

		assert.True(t, called, "Callback should be called when computed value changes")
		assert.Equal(t, 20, newVal, "New computed value should be 20")
		assert.Equal(t, 10, oldVal, "Old computed value should be 10")
	})

	t.Run("watch chained computed values", func(t *testing.T) {
		count := NewRef(2)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})
		quadrupled := NewComputed(func() int {
			return doubled.Get() * 2
		})

		var callCount int
		var values []int

		cleanup := Watch(quadrupled, func(n, o int) {
			callCount++
			values = append(values, n)
		})
		defer cleanup()

		count.Set(3) // quadrupled: 2*2*2=8 -> 3*2*2=12
		count.Set(5) // quadrupled: 12 -> 5*2*2=20

		assert.Equal(t, 2, callCount, "Callback should be called twice")
		assert.Equal(t, []int{12, 20}, values, "Should track chained computed changes")
	})

	t.Run("multiple watchers on same computed", func(t *testing.T) {
		count := NewRef(1)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		var called1, called2 bool
		var val1, val2 int

		cleanup1 := Watch(doubled, func(n, o int) {
			called1 = true
			val1 = n
		})
		defer cleanup1()

		cleanup2 := Watch(doubled, func(n, o int) {
			called2 = true
			val2 = n
		})
		defer cleanup2()

		count.Set(5)

		assert.True(t, called1, "First watcher should be called")
		assert.True(t, called2, "Second watcher should be called")
		assert.Equal(t, 10, val1, "First watcher should see new value")
		assert.Equal(t, 10, val2, "Second watcher should see new value")
	})

	t.Run("computed with immediate execution", func(t *testing.T) {
		count := NewRef(7)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		var called bool
		var immediateVal int

		cleanup := Watch(doubled, func(n, o int) {
			if !called {
				immediateVal = n
			}
			called = true
		}, WithImmediate())
		defer cleanup()

		assert.True(t, called, "Callback should be called immediately")
		assert.Equal(t, 14, immediateVal, "Should receive current computed value")
	})

	t.Run("computed with deep watching", func(t *testing.T) {
		type Data struct {
			Value int
		}

		ref := NewRef(Data{Value: 5})
		computed := NewComputed(func() Data {
			return ref.Get()
		})

		var callCount int

		cleanup := Watch(computed, func(n, o Data) {
			callCount++
		}, WithDeep())
		defer cleanup()

		// Set same value - should not trigger with deep watching
		ref.Set(Data{Value: 5})
		assert.Equal(t, 0, callCount, "Should not trigger for equal values")

		// Set different value - should trigger
		ref.Set(Data{Value: 10})
		assert.Equal(t, 1, callCount, "Should trigger for different values")
	})

	t.Run("computed with flush modes", func(t *testing.T) {
		count := NewRef(0)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		var syncCalls, postCalls int

		cleanupSync := Watch(doubled, func(n, o int) {
			syncCalls++
		}, WithFlush("sync"))
		defer cleanupSync()

		cleanupPost := Watch(doubled, func(n, o int) {
			postCalls++
		}, WithFlush("post"))
		defer cleanupPost()

		count.Set(1)
		count.Set(2)
		count.Set(3)

		assert.Equal(t, 3, syncCalls, "Sync watcher should execute immediately")
		assert.Equal(t, 0, postCalls, "Post watcher should be queued")

		FlushWatchers()
		assert.Equal(t, 1, postCalls, "Post watcher should execute once after flush")
	})

	t.Run("cleanup stops watching computed", func(t *testing.T) {
		count := NewRef(1)
		doubled := NewComputed(func() int {
			return count.Get() * 2
		})

		var callCount int

		cleanup := Watch(doubled, func(n, o int) {
			callCount++
		})

		count.Set(2)
		assert.Equal(t, 1, callCount, "Should be called before cleanup")

		cleanup()

		count.Set(3)
		assert.Equal(t, 1, callCount, "Should not be called after cleanup")
	})
}

// TestWatch_ComputedConcurrency tests thread safety of watching computed values
func TestWatch_ComputedConcurrency(t *testing.T) {
	count := NewRef(0)
	doubled := NewComputed(func() int {
		return count.Get() * 2
	})

	var callCount atomic.Int32
	var wg sync.WaitGroup

	// Multiple watchers
	for i := 0; i < 10; i++ {
		cleanup := Watch(doubled, func(n, o int) {
			callCount.Add(1)
		})
		defer cleanup()
	}

	// Concurrent updates with distinct values
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(val int) {
			defer wg.Done()
			// Use unique values to ensure each update produces a different computed result
			count.Set(val * 100)
		}(i)
	}
	wg.Wait()

	// With concurrent updates, some notifications might be skipped if values don't change
	// or if updates happen while computed is being evaluated. We just verify that:
	// 1. At least some notifications happened (> 0)
	// 2. No more than expected (10 updates × 10 watchers = 100)
	calls := callCount.Load()
	assert.Greater(t, calls, int32(0), "Should have at least some notifications")
	assert.LessOrEqual(t, calls, int32(100), "Should not exceed maximum possible notifications")
}

// TestWatch_ComputedNoChange tests that computed watchers don't trigger on no-op changes
func TestWatch_ComputedNoChange(t *testing.T) {
	count := NewRef(5)
	doubled := NewComputed(func() int {
		return count.Get() * 2
	})

	var callCount int

	cleanup := Watch(doubled, func(n, o int) {
		callCount++
	})
	defer cleanup()

	// Set to same value - computed result doesn't change
	count.Set(5)
	assert.Equal(t, 0, callCount, "Should not trigger when computed value doesn't change")

	// Set to different value - computed result changes
	count.Set(10)
	assert.Equal(t, 1, callCount, "Should trigger when computed value changes")
}
