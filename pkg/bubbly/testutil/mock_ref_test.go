package testutil

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewMockRef tests MockRef creation with various types
func TestNewMockRef(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		expected interface{}
	}{
		{"integer", 42, 42},
		{"string", "hello", "hello"},
		{"boolean", true, true},
		{"float", 3.14, 3.14},
		{"slice", []int{1, 2, 3}, []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}, map[string]int{"a": 1}},
		{"struct", struct{ Name string }{"test"}, struct{ Name string }{"test"}},
		{"nil slice", []int(nil), []int(nil)},
		{"zero value int", 0, 0},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.initial.(type) {
			case int:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			case string:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			case bool:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			case float64:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			case []int:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			case map[string]int:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			default:
				mockRef := NewMockRef(v)
				assert.Equal(t, tt.expected, mockRef.Get())
			}
		})
	}
}

// TestMockRef_Get tests Get() method and call tracking
func TestMockRef_Get(t *testing.T) {
	tests := []struct {
		name      string
		initial   int
		getCalls  int
		wantValue int
		wantCalls int
	}{
		{"single get", 42, 1, 42, 1},
		{"multiple gets", 100, 5, 100, 5},
		{"zero gets", 0, 0, 0, 0},
		{"ten gets", 999, 10, 999, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRef := NewMockRef(tt.initial)

			// Call Get() the specified number of times
			var lastValue int
			for i := 0; i < tt.getCalls; i++ {
				lastValue = mockRef.Get()
			}

			// Verify value if Get was called
			if tt.getCalls > 0 {
				assert.Equal(t, tt.wantValue, lastValue)
			}

			// Verify call count
			mockRef.AssertGetCalled(t, tt.wantCalls)
			assert.Equal(t, tt.wantCalls, mockRef.GetCallCount())
		})
	}
}

// TestMockRef_Set tests Set() method and call tracking
func TestMockRef_Set(t *testing.T) {
	tests := []struct {
		name      string
		initial   int
		sets      []int
		wantValue int
		wantCalls int
	}{
		{"single set", 0, []int{42}, 42, 1},
		{"multiple sets", 0, []int{10, 20, 30}, 30, 3},
		{"set same value", 42, []int{42, 42}, 42, 2},
		{"set zero", 100, []int{0}, 0, 1},
		{"empty sets", 42, []int{}, 42, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRef := NewMockRef(tt.initial)

			// Call Set() with each value
			for _, val := range tt.sets {
				mockRef.Set(val)
			}

			// Verify final value
			assert.Equal(t, tt.wantValue, mockRef.Get())

			// Verify call count
			mockRef.AssertSetCalled(t, tt.wantCalls)
			assert.Equal(t, tt.wantCalls, mockRef.SetCallCount())
		})
	}
}

// TestMockRef_Watch tests watcher functionality
func TestMockRef_Watch(t *testing.T) {
	tests := []struct {
		name          string
		initial       int
		setValue      int
		expectNotify  bool
		expectedValue int
	}{
		{"value changes", 0, 42, true, 42},
		{"value same", 42, 42, false, 42},
		{"zero to non-zero", 0, 100, true, 100},
		{"non-zero to zero", 100, 0, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRef := NewMockRef(tt.initial)

			notified := false
			var notifiedValue int

			mockRef.Watch(func(newVal int) {
				notified = true
				notifiedValue = newVal
			})

			mockRef.Set(tt.setValue)

			assert.Equal(t, tt.expectNotify, notified)
			if tt.expectNotify {
				assert.Equal(t, tt.expectedValue, notifiedValue)
			}
		})
	}
}

// TestMockRef_MultipleWatchers tests multiple watchers
func TestMockRef_MultipleWatchers(t *testing.T) {
	mockRef := NewMockRef(0)

	count1 := 0
	count2 := 0
	count3 := 0

	mockRef.Watch(func(val int) { count1++ })
	mockRef.Watch(func(val int) { count2++ })
	mockRef.Watch(func(val int) { count3++ })

	// Set to different value - should notify all watchers
	mockRef.Set(42)

	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)
	assert.Equal(t, 1, count3)

	// Set to same value - should not notify
	mockRef.Set(42)

	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)
	assert.Equal(t, 1, count3)

	// Set to different value again
	mockRef.Set(100)

	assert.Equal(t, 2, count1)
	assert.Equal(t, 2, count2)
	assert.Equal(t, 2, count3)
}

// TestMockRef_WatcherReceivesCorrectValue tests watcher receives new value
func TestMockRef_WatcherReceivesCorrectValue(t *testing.T) {
	tests := []struct {
		name           string
		initial        int
		values         []int
		expectedValues []int
	}{
		{"single value", 0, []int{42}, []int{42}},
		{"multiple values", 0, []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{"zero value", 10, []int{0}, []int{0}}, // Changed initial to 10 so 0 is a change
		{"negative values", 0, []int{-1, -10, -100}, []int{-1, -10, -100}},
		{"same value not notified", 42, []int{42}, []int{}}, // Setting same value should not notify
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRef := NewMockRef(tt.initial)

			receivedValues := []int{}
			mockRef.Watch(func(newVal int) {
				receivedValues = append(receivedValues, newVal)
			})

			for _, val := range tt.values {
				mockRef.Set(val)
			}

			assert.Equal(t, tt.expectedValues, receivedValues)
		})
	}
}

// TestMockRef_AssertGetCalled tests AssertGetCalled assertion
func TestMockRef_AssertGetCalled(t *testing.T) {
	tests := []struct {
		name       string
		getCalls   int
		assertCall int
		shouldFail bool
	}{
		{"correct count", 3, 3, false},
		{"zero calls", 0, 0, false},
		{"wrong count", 5, 3, true},
		{"expected more", 2, 5, true},
		{"expected less", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRef := NewMockRef(42)

			// Call Get() the specified number of times
			for i := 0; i < tt.getCalls; i++ {
				mockRef.Get()
			}

			// Use mock testing.T to capture failures
			mockT := &testing.T{}
			mockRef.AssertGetCalled(mockT, tt.assertCall)

			if tt.shouldFail {
				assert.True(t, mockT.Failed(), "Expected assertion to fail")
			} else {
				assert.False(t, mockT.Failed(), "Expected assertion to pass")
			}
		})
	}
}

// TestMockRef_AssertSetCalled tests AssertSetCalled assertion
func TestMockRef_AssertSetCalled(t *testing.T) {
	tests := []struct {
		name       string
		setCalls   int
		assertCall int
		shouldFail bool
	}{
		{"correct count", 3, 3, false},
		{"zero calls", 0, 0, false},
		{"wrong count", 5, 3, true},
		{"expected more", 2, 5, true},
		{"expected less", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRef := NewMockRef(0)

			// Call Set() the specified number of times
			for i := 0; i < tt.setCalls; i++ {
				mockRef.Set(i)
			}

			// Use mock testing.T to capture failures
			mockT := &testing.T{}
			mockRef.AssertSetCalled(mockT, tt.assertCall)

			if tt.shouldFail {
				assert.True(t, mockT.Failed(), "Expected assertion to fail")
			} else {
				assert.False(t, mockT.Failed(), "Expected assertion to pass")
			}
		})
	}
}

// TestMockRef_Reset tests Reset() method
func TestMockRef_Reset(t *testing.T) {
	mockRef := NewMockRef(0)

	// Make some calls
	mockRef.Get()
	mockRef.Get()
	mockRef.Set(42)
	mockRef.Set(100)

	// Verify calls were tracked
	assert.Equal(t, 2, mockRef.GetCallCount())
	assert.Equal(t, 2, mockRef.SetCallCount())

	// Reset
	mockRef.Reset()

	// Verify counters reset
	assert.Equal(t, 0, mockRef.GetCallCount())
	assert.Equal(t, 0, mockRef.SetCallCount())
	mockRef.AssertGetCalled(t, 0)
	mockRef.AssertSetCalled(t, 0)

	// Value should remain unchanged
	assert.Equal(t, 100, mockRef.Get())
}

// TestMockRef_ThreadSafety tests concurrent access
func TestMockRef_ThreadSafety(t *testing.T) {
	mockRef := NewMockRef(0)

	const goroutines = 10
	const operations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // For both Get and Set operations

	// Concurrent Gets
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				mockRef.Get()
			}
		}()
	}

	// Concurrent Sets
	for i := 0; i < goroutines; i++ {
		go func(val int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				mockRef.Set(val)
			}
		}(i)
	}

	wg.Wait()

	// Verify call counts
	assert.Equal(t, goroutines*operations, mockRef.GetCallCount())
	assert.Equal(t, goroutines*operations, mockRef.SetCallCount())
}

// TestMockRef_WatcherThreadSafety tests concurrent watcher registration and notification
func TestMockRef_WatcherThreadSafety(t *testing.T) {
	mockRef := NewMockRef(0)

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent watcher registration
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			mockRef.Watch(func(val int) {
				// Watcher callback
			})
		}()
	}

	// Concurrent Sets (which trigger watchers)
	for i := 0; i < goroutines; i++ {
		go func(val int) {
			defer wg.Done()
			mockRef.Set(val)
		}(i)
	}

	wg.Wait()

	// Should not panic or deadlock
	assert.Equal(t, goroutines, mockRef.SetCallCount())
}

// TestMockRef_TypeSafety tests type safety with different types
func TestMockRef_TypeSafety(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		mockRef := NewMockRef("hello")
		assert.Equal(t, "hello", mockRef.Get())
		mockRef.Set("world")
		assert.Equal(t, "world", mockRef.Get())
	})

	t.Run("struct type", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}
		mockRef := NewMockRef(User{Name: "John", Age: 30})
		assert.Equal(t, User{Name: "John", Age: 30}, mockRef.Get())
		mockRef.Set(User{Name: "Jane", Age: 25})
		assert.Equal(t, User{Name: "Jane", Age: 25}, mockRef.Get())
	})

	t.Run("pointer type", func(t *testing.T) {
		val := 42
		mockRef := NewMockRef(&val)
		assert.Equal(t, &val, mockRef.Get())
	})

	t.Run("slice type", func(t *testing.T) {
		mockRef := NewMockRef([]string{"a", "b", "c"})
		assert.Equal(t, []string{"a", "b", "c"}, mockRef.Get())
		mockRef.Set([]string{"x", "y"})
		assert.Equal(t, []string{"x", "y"}, mockRef.Get())
	})

	t.Run("map type", func(t *testing.T) {
		mockRef := NewMockRef(map[string]int{"a": 1, "b": 2})
		assert.Equal(t, map[string]int{"a": 1, "b": 2}, mockRef.Get())
	})
}

// TestMockRef_ComplexScenario tests a realistic usage scenario
func TestMockRef_ComplexScenario(t *testing.T) {
	// Simulate testing a component that uses a ref
	mockRef := NewMockRef(0)

	// Track watcher calls
	watcherCalls := 0
	var watcherValues []int

	mockRef.Watch(func(newVal int) {
		watcherCalls++
		watcherValues = append(watcherValues, newVal)
	})

	// Simulate component initialization
	initialValue := mockRef.Get()
	assert.Equal(t, 0, initialValue)
	mockRef.AssertGetCalled(t, 1)

	// Simulate user interaction
	mockRef.Set(10)
	mockRef.Set(20)
	mockRef.Set(30)

	// Verify state
	assert.Equal(t, 30, mockRef.Get())
	mockRef.AssertGetCalled(t, 2)
	mockRef.AssertSetCalled(t, 3)

	// Verify watchers were notified
	assert.Equal(t, 3, watcherCalls)
	assert.Equal(t, []int{10, 20, 30}, watcherValues)

	// Reset for next test scenario
	mockRef.Reset()
	mockRef.AssertGetCalled(t, 0)
	mockRef.AssertSetCalled(t, 0)

	// Value should persist after reset
	assert.Equal(t, 30, mockRef.Get())
}
