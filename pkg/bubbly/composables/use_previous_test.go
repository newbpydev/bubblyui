package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUsePrevious_InitialPreviousIsNil tests that initial previous is nil
func TestUsePrevious_InitialPreviousIsNil(t *testing.T) {
	tests := []struct {
		name    string
		initial int
	}{
		{"initial zero", 0},
		{"initial positive", 42},
		{"initial negative", -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ref := bubbly.NewRef(tt.initial)
			previous := UsePrevious(ctx, ref)

			assert.NotNil(t, previous, "UsePrevious should return non-nil")
			assert.NotNil(t, previous.Value, "Value should not be nil")
			assert.Nil(t, previous.Get(), "Initial previous should be nil")
		})
	}
}

// TestUsePrevious_PreviousUpdatesWhenRefChanges tests that previous updates when ref changes
func TestUsePrevious_PreviousUpdatesWhenRefChanges(t *testing.T) {
	tests := []struct {
		name     string
		initial  int
		newValue int
	}{
		{"change from zero", 0, 10},
		{"change to zero", 10, 0},
		{"positive to positive", 5, 15},
		{"negative to positive", -5, 5},
		{"positive to negative", 5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			ref := bubbly.NewRef(tt.initial)
			previous := UsePrevious(ctx, ref)

			// Initially nil
			assert.Nil(t, previous.Get(), "Initial previous should be nil")

			// Change the ref
			ref.Set(tt.newValue)

			// Previous should now be the old value
			assert.NotNil(t, previous.Get(), "Previous should not be nil after change")
			assert.Equal(t, tt.initial, *previous.Get(),
				"Previous should be %d after change", tt.initial)
		})
	}
}

// TestUsePrevious_GetReturnsCorrectPreviousValue tests that Get returns correct previous value
func TestUsePrevious_GetReturnsCorrectPreviousValue(t *testing.T) {
	ctx := createTestContext()
	ref := bubbly.NewRef(1)
	previous := UsePrevious(ctx, ref)

	// Initial: nil
	assert.Nil(t, previous.Get(), "Initial previous should be nil")

	// Change 1 -> 2
	ref.Set(2)
	assert.NotNil(t, previous.Get())
	assert.Equal(t, 1, *previous.Get(), "Previous should be 1")

	// Change 2 -> 3
	ref.Set(3)
	assert.Equal(t, 2, *previous.Get(), "Previous should be 2")

	// Change 3 -> 4
	ref.Set(4)
	assert.Equal(t, 3, *previous.Get(), "Previous should be 3")

	// Change 4 -> 100
	ref.Set(100)
	assert.Equal(t, 4, *previous.Get(), "Previous should be 4")
}

// TestUsePrevious_WorksWithWatchInternally tests that Watch is used internally
func TestUsePrevious_WorksWithWatchInternally(t *testing.T) {
	ctx := createTestContext()
	ref := bubbly.NewRef(10)
	previous := UsePrevious(ctx, ref)

	// Verify initial state
	assert.Nil(t, previous.Get())

	// Multiple rapid changes
	for i := 20; i <= 50; i += 10 {
		oldVal := ref.GetTyped()
		ref.Set(i)
		assert.Equal(t, oldVal, *previous.Get(),
			"Previous should track the old value after each change")
	}
}

// TestUsePrevious_WorksWithDifferentTypes tests generic type support
func TestUsePrevious_WorksWithDifferentTypes(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		ctx := createTestContext()
		ref := bubbly.NewRef("hello")
		previous := UsePrevious(ctx, ref)

		assert.Nil(t, previous.Get())

		ref.Set("world")
		assert.NotNil(t, previous.Get())
		assert.Equal(t, "hello", *previous.Get())

		ref.Set("foo")
		assert.Equal(t, "world", *previous.Get())
	})

	t.Run("struct type", func(t *testing.T) {
		type User struct {
			Name string
			Age  int
		}

		ctx := createTestContext()
		ref := bubbly.NewRef(User{Name: "Alice", Age: 30})
		previous := UsePrevious(ctx, ref)

		assert.Nil(t, previous.Get())

		ref.Set(User{Name: "Bob", Age: 25})
		assert.NotNil(t, previous.Get())
		assert.Equal(t, "Alice", previous.Get().Name)
		assert.Equal(t, 30, previous.Get().Age)
	})

	t.Run("slice type", func(t *testing.T) {
		ctx := createTestContext()
		ref := bubbly.NewRef([]int{1, 2, 3})
		previous := UsePrevious(ctx, ref)

		assert.Nil(t, previous.Get())

		ref.Set([]int{4, 5, 6})
		assert.NotNil(t, previous.Get())
		assert.Equal(t, []int{1, 2, 3}, *previous.Get())
	})
}

// TestUsePrevious_WorksWithCreateShared tests shared composable pattern
func TestUsePrevious_WorksWithCreateShared(t *testing.T) {
	// Create a shared ref first
	sharedRef := bubbly.NewRef(100)

	// Create shared previous tracker
	sharedPrevious := CreateShared(func(ctx *bubbly.Context) *PreviousReturn[int] {
		return UsePrevious(ctx, sharedRef)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	previous1 := sharedPrevious(ctx1)
	previous2 := sharedPrevious(ctx2)

	// Both should be the same instance
	assert.Nil(t, previous1.Get())
	assert.Nil(t, previous2.Get())

	// Change the shared ref
	sharedRef.Set(200)

	// Both should see the same previous value
	assert.Equal(t, 100, *previous1.Get())
	assert.Equal(t, 100, *previous2.Get())
}

// TestUsePrevious_ValueIsReactive tests that Value ref is reactive
func TestUsePrevious_ValueIsReactive(t *testing.T) {
	ctx := createTestContext()
	ref := bubbly.NewRef(10)
	previous := UsePrevious(ctx, ref)

	// Track changes to the previous value
	changeCount := 0
	bubbly.Watch(previous.Value, func(newVal, oldVal *int) {
		changeCount++
	})

	// Change ref should trigger watcher on previous.Value
	ref.Set(20)
	assert.Equal(t, 1, changeCount, "First change should trigger watcher")

	ref.Set(30)
	assert.Equal(t, 2, changeCount, "Second change should trigger watcher")

	ref.Set(40)
	assert.Equal(t, 3, changeCount, "Third change should trigger watcher")
}

// TestUsePrevious_MultipleChangesTrackCorrectly tests rapid changes
func TestUsePrevious_MultipleChangesTrackCorrectly(t *testing.T) {
	ctx := createTestContext()
	ref := bubbly.NewRef(0)
	previous := UsePrevious(ctx, ref)

	// Rapid changes
	for i := 1; i <= 100; i++ {
		expectedPrev := i - 1
		ref.Set(i)

		if i == 1 {
			// First change: previous was 0
			assert.Equal(t, 0, *previous.Get())
		} else {
			assert.Equal(t, expectedPrev, *previous.Get(),
				"After setting to %d, previous should be %d", i, expectedPrev)
		}
	}
}

// TestUsePrevious_ZeroValueVsNil tests distinction between zero value and nil
func TestUsePrevious_ZeroValueVsNil(t *testing.T) {
	ctx := createTestContext()
	ref := bubbly.NewRef(0) // Start with zero
	previous := UsePrevious(ctx, ref)

	// Initially nil (no previous value yet)
	assert.Nil(t, previous.Get(), "Should be nil before any changes")

	// Change from 0 to 1
	ref.Set(1)

	// Previous should be pointer to 0, not nil
	assert.NotNil(t, previous.Get(), "Should not be nil after change")
	assert.Equal(t, 0, *previous.Get(), "Previous should be 0 (the old value)")
}

// TestUsePrevious_EmptyStringVsNil tests distinction with strings
func TestUsePrevious_EmptyStringVsNil(t *testing.T) {
	ctx := createTestContext()
	ref := bubbly.NewRef("") // Start with empty string
	previous := UsePrevious(ctx, ref)

	// Initially nil
	assert.Nil(t, previous.Get())

	// Change from "" to "hello"
	ref.Set("hello")

	// Previous should be pointer to "", not nil
	assert.NotNil(t, previous.Get())
	assert.Equal(t, "", *previous.Get(), "Previous should be empty string")
}
