package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRef verifies that NewRef creates a ref with the initial value
func TestNewRef(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		expected interface{}
	}{
		{
			name:     "integer ref",
			initial:  42,
			expected: 42,
		},
		{
			name:     "string ref",
			initial:  "hello",
			expected: "hello",
		},
		{
			name:     "zero value int",
			initial:  0,
			expected: 0,
		},
		{
			name:     "zero value string",
			initial:  "",
			expected: "",
		},
		{
			name:     "boolean ref",
			initial:  true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.initial.(type) {
			case int:
				ref := NewRef(v)
				assert.NotNil(t, ref, "NewRef should return non-nil reference")
				assert.Equal(t, tt.expected, ref.Get(), "Initial value should match")
			case string:
				ref := NewRef(v)
				assert.NotNil(t, ref, "NewRef should return non-nil reference")
				assert.Equal(t, tt.expected, ref.Get(), "Initial value should match")
			case bool:
				ref := NewRef(v)
				assert.NotNil(t, ref, "NewRef should return non-nil reference")
				assert.Equal(t, tt.expected, ref.Get(), "Initial value should match")
			}
		})
	}
}

// TestRef_Get verifies that Get returns the current value
func TestRef_Get(t *testing.T) {
	t.Run("get integer value", func(t *testing.T) {
		ref := NewRef(100)
		value := ref.Get()
		assert.Equal(t, 100, value, "Get should return current value")
	})

	t.Run("get string value", func(t *testing.T) {
		ref := NewRef("test")
		value := ref.Get()
		assert.Equal(t, "test", value, "Get should return current value")
	})

	t.Run("get struct value", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		user := User{Name: "John", Age: 30}
		ref := NewRef(user)
		value := ref.Get()
		assert.Equal(t, user, value, "Get should return current struct value")
	})

	t.Run("get pointer value", func(t *testing.T) {
		val := 42
		ref := NewRef(&val)
		value := ref.Get()
		assert.Equal(t, &val, value, "Get should return current pointer value")
		assert.Equal(t, 42, *value, "Dereferenced value should be correct")
	})
}

// TestRef_Set verifies that Set updates the value
func TestRef_Set(t *testing.T) {
	t.Run("set integer value", func(t *testing.T) {
		ref := NewRef(10)
		assert.Equal(t, 10, ref.Get(), "Initial value should be 10")

		ref.Set(20)
		assert.Equal(t, 20, ref.Get(), "Value should be updated to 20")

		ref.Set(30)
		assert.Equal(t, 30, ref.Get(), "Value should be updated to 30")
	})

	t.Run("set string value", func(t *testing.T) {
		ref := NewRef("initial")
		assert.Equal(t, "initial", ref.Get(), "Initial value should be 'initial'")

		ref.Set("updated")
		assert.Equal(t, "updated", ref.Get(), "Value should be updated to 'updated'")
	})

	t.Run("set to zero value", func(t *testing.T) {
		ref := NewRef(100)
		ref.Set(0)
		assert.Equal(t, 0, ref.Get(), "Should be able to set to zero value")
	})

	t.Run("set struct value", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		ref := NewRef(User{Name: "John", Age: 30})

		newUser := User{Name: "Jane", Age: 25}
		ref.Set(newUser)
		assert.Equal(t, newUser, ref.Get(), "Struct value should be updated")
	})
}

// TestRef_TypeSafety verifies compile-time type safety
func TestRef_TypeSafety(t *testing.T) {
	t.Run("int ref maintains type", func(t *testing.T) {
		ref := NewRef(42)
		// This should compile - same type
		ref.Set(100)
		assert.Equal(t, 100, ref.Get())
	})

	t.Run("string ref maintains type", func(t *testing.T) {
		ref := NewRef("hello")
		// This should compile - same type
		ref.Set("world")
		assert.Equal(t, "world", ref.Get())
	})

	// Note: Type mismatches are caught at compile time with generics
	// The following would not compile:
	// ref := NewRef(42)
	// ref.Set("string") // Compile error: cannot use "string" (type string) as type int
}

// TestRef_MultipleRefs verifies that multiple refs are independent
func TestRef_MultipleRefs(t *testing.T) {
	ref1 := NewRef(10)
	ref2 := NewRef(20)
	ref3 := NewRef("hello")

	assert.Equal(t, 10, ref1.Get(), "ref1 should have value 10")
	assert.Equal(t, 20, ref2.Get(), "ref2 should have value 20")
	assert.Equal(t, "hello", ref3.Get(), "ref3 should have value 'hello'")

	ref1.Set(100)
	assert.Equal(t, 100, ref1.Get(), "ref1 should be updated")
	assert.Equal(t, 20, ref2.Get(), "ref2 should remain unchanged")
	assert.Equal(t, "hello", ref3.Get(), "ref3 should remain unchanged")
}

// TestRef_ZeroValue verifies handling of zero values
func TestRef_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "zero int",
			test: func(t *testing.T) {
				ref := NewRef(0)
				assert.Equal(t, 0, ref.Get())
			},
		},
		{
			name: "zero string",
			test: func(t *testing.T) {
				ref := NewRef("")
				assert.Equal(t, "", ref.Get())
			},
		},
		{
			name: "zero bool",
			test: func(t *testing.T) {
				ref := NewRef(false)
				assert.Equal(t, false, ref.Get())
			},
		},
		{
			name: "nil pointer",
			test: func(t *testing.T) {
				var ptr *int
				ref := NewRef(ptr)
				assert.Nil(t, ref.Get())
			},
		},
		{
			name: "nil slice",
			test: func(t *testing.T) {
				var slice []int
				ref := NewRef(slice)
				assert.Nil(t, ref.Get())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestRef_ComplexTypes verifies Ref works with complex types
func TestRef_ComplexTypes(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		ref := NewRef([]int{1, 2, 3})
		assert.Equal(t, []int{1, 2, 3}, ref.Get())

		ref.Set([]int{4, 5, 6})
		assert.Equal(t, []int{4, 5, 6}, ref.Get())
	})

	t.Run("map", func(t *testing.T) {
		ref := NewRef(map[string]int{"a": 1, "b": 2})
		value := ref.Get()
		assert.Equal(t, 1, value["a"])
		assert.Equal(t, 2, value["b"])

		ref.Set(map[string]int{"c": 3})
		value = ref.Get()
		assert.Equal(t, 3, value["c"])
	})

	t.Run("nested struct", func(t *testing.T) {
		type Address struct {
			City string
		}
		type User struct {
			Name    string
			Address Address
		}

		user := User{Name: "John", Address: Address{City: "NYC"}}
		ref := NewRef(user)
		assert.Equal(t, "NYC", ref.Get().Address.City)

		newUser := User{Name: "Jane", Address: Address{City: "LA"}}
		ref.Set(newUser)
		assert.Equal(t, "LA", ref.Get().Address.City)
	})
}

// TestRef_WatcherNotification verifies that watchers receive notifications
func TestRef_WatcherNotification(t *testing.T) {
	t.Run("single watcher receives notification", func(t *testing.T) {
		ref := NewRef(10)

		var receivedNew, receivedOld int
		var callCount int

		w := &watcher[int]{
			callback: func(newVal, oldVal int) {
				receivedNew = newVal
				receivedOld = oldVal
				callCount++
			},
		}

		ref.addWatcher(w)
		ref.Set(20)

		assert.Equal(t, 1, callCount, "Callback should be called once")
		assert.Equal(t, 10, receivedOld, "Old value should be 10")
		assert.Equal(t, 20, receivedNew, "New value should be 20")
	})

	t.Run("watcher receives correct values on multiple updates", func(t *testing.T) {
		ref := NewRef("initial")

		var updates []struct {
			old string
			new string
		}

		w := &watcher[string]{
			callback: func(newVal, oldVal string) {
				updates = append(updates, struct {
					old string
					new string
				}{old: oldVal, new: newVal})
			},
		}

		ref.addWatcher(w)
		ref.Set("second")
		ref.Set("third")

		assert.Len(t, updates, 2, "Should have 2 updates")
		assert.Equal(t, "initial", updates[0].old)
		assert.Equal(t, "second", updates[0].new)
		assert.Equal(t, "second", updates[1].old)
		assert.Equal(t, "third", updates[1].new)
	})
}

// TestRef_MultipleWatchers verifies that multiple watchers work independently
func TestRef_MultipleWatchers(t *testing.T) {
	t.Run("multiple watchers all receive notifications", func(t *testing.T) {
		ref := NewRef(0)

		var count1, count2, count3 int

		w1 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				count1++
			},
		}
		w2 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				count2++
			},
		}
		w3 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				count3++
			},
		}

		ref.addWatcher(w1)
		ref.addWatcher(w2)
		ref.addWatcher(w3)

		ref.Set(1)
		ref.Set(2)

		assert.Equal(t, 2, count1, "Watcher 1 should be called twice")
		assert.Equal(t, 2, count2, "Watcher 2 should be called twice")
		assert.Equal(t, 2, count3, "Watcher 3 should be called twice")
	})

	t.Run("watchers are independent", func(t *testing.T) {
		ref := NewRef(100)

		var sum1, sum2 int

		w1 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				sum1 += newVal
			},
		}
		w2 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				sum2 += newVal * 2
			},
		}

		ref.addWatcher(w1)
		ref.addWatcher(w2)

		ref.Set(10)
		ref.Set(20)

		assert.Equal(t, 30, sum1, "Watcher 1 sum should be 10+20")
		assert.Equal(t, 60, sum2, "Watcher 2 sum should be (10*2)+(20*2)")
	})
}

// TestRef_WatcherRemoval verifies that watcher removal works
func TestRef_WatcherRemoval(t *testing.T) {
	t.Run("removed watcher does not receive notifications", func(t *testing.T) {
		ref := NewRef(0)

		var count int
		w := &watcher[int]{
			callback: func(newVal, oldVal int) {
				count++
			},
		}

		ref.addWatcher(w)
		ref.Set(1)
		assert.Equal(t, 1, count, "Should be called once before removal")

		ref.removeWatcher(w)
		ref.Set(2)
		assert.Equal(t, 1, count, "Should not be called after removal")
	})

	t.Run("removing one watcher does not affect others", func(t *testing.T) {
		ref := NewRef(0)

		var count1, count2 int

		w1 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				count1++
			},
		}
		w2 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				count2++
			},
		}

		ref.addWatcher(w1)
		ref.addWatcher(w2)

		ref.Set(1)
		assert.Equal(t, 1, count1)
		assert.Equal(t, 1, count2)

		ref.removeWatcher(w1)
		ref.Set(2)

		assert.Equal(t, 1, count1, "Removed watcher should not be called")
		assert.Equal(t, 2, count2, "Other watcher should still be called")
	})

	t.Run("removing non-existent watcher is safe", func(t *testing.T) {
		ref := NewRef(0)

		w := &watcher[int]{
			callback: func(newVal, oldVal int) {},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			ref.removeWatcher(w)
		})
	})
}

// TestRef_WatcherNotificationOrder verifies notification order is consistent
func TestRef_WatcherNotificationOrder(t *testing.T) {
	t.Run("watchers notified in registration order", func(t *testing.T) {
		ref := NewRef(0)

		var order []int

		w1 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				order = append(order, 1)
			},
		}
		w2 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				order = append(order, 2)
			},
		}
		w3 := &watcher[int]{
			callback: func(newVal, oldVal int) {
				order = append(order, 3)
			},
		}

		ref.addWatcher(w1)
		ref.addWatcher(w2)
		ref.addWatcher(w3)

		ref.Set(1)

		assert.Equal(t, []int{1, 2, 3}, order, "Watchers should be notified in registration order")
	})
}

// TestRef_WatcherMemoryLeak verifies no memory leaks with watchers
func TestRef_WatcherMemoryLeak(t *testing.T) {
	t.Run("watchers can be removed to prevent leaks", func(t *testing.T) {
		ref := NewRef(0)

		// Add and remove many watchers
		watchers := make([]*watcher[int], 100)
		for i := 0; i < 100; i++ {
			w := &watcher[int]{
				callback: func(newVal, oldVal int) {},
			}
			watchers[i] = w
			ref.addWatcher(w)
		}

		// Remove all watchers
		for _, w := range watchers {
			ref.removeWatcher(w)
		}

		// Verify no watchers remain (internal check)
		ref.mu.RLock()
		watcherCount := len(ref.watchers)
		ref.mu.RUnlock()

		assert.Equal(t, 0, watcherCount, "All watchers should be removed")
	})
}
