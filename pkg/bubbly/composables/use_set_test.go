package composables

import (
	"sort"
	"sync"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UseSet Tests - Task 4.3
// =============================================================================

func TestUseSet_InitialValuesSetCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		expected map[string]struct{}
	}{
		{
			name:     "empty slice",
			initial:  []string{},
			expected: map[string]struct{}{},
		},
		{
			name:     "nil slice becomes empty",
			initial:  nil,
			expected: map[string]struct{}{},
		},
		{
			name:     "slice with values",
			initial:  []string{"a", "b", "c"},
			expected: map[string]struct{}{"a": {}, "b": {}, "c": {}},
		},
		{
			name:     "single value",
			initial:  []string{"only"},
			expected: map[string]struct{}{"only": {}},
		},
		{
			name:     "duplicates in initial ignored",
			initial:  []string{"a", "b", "a", "c", "b"},
			expected: map[string]struct{}{"a": {}, "b": {}, "c": {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			require.NotNil(t, s)
			require.NotNil(t, s.Values)
			assert.Equal(t, tt.expected, s.Values.GetTyped())
		})
	}
}

func TestUseSet_Add_AddsValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		value    string
		expected map[string]struct{}
	}{
		{
			name:     "add to empty set",
			initial:  []string{},
			value:    "a",
			expected: map[string]struct{}{"a": {}},
		},
		{
			name:     "add new value",
			initial:  []string{"a"},
			value:    "b",
			expected: map[string]struct{}{"a": {}, "b": {}},
		},
		{
			name:     "add existing value (no-op)",
			initial:  []string{"a", "b"},
			value:    "a",
			expected: map[string]struct{}{"a": {}, "b": {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			s.Add(tt.value)

			assert.Equal(t, tt.expected, s.Values.GetTyped())
		})
	}
}

func TestUseSet_Delete_RemovesValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		value    string
		wantOk   bool
		expected map[string]struct{}
	}{
		{
			name:     "delete existing value",
			initial:  []string{"a", "b", "c"},
			value:    "b",
			wantOk:   true,
			expected: map[string]struct{}{"a": {}, "c": {}},
		},
		{
			name:     "delete non-existing value",
			initial:  []string{"a", "b"},
			value:    "c",
			wantOk:   false,
			expected: map[string]struct{}{"a": {}, "b": {}},
		},
		{
			name:     "delete from empty set",
			initial:  []string{},
			value:    "any",
			wantOk:   false,
			expected: map[string]struct{}{},
		},
		{
			name:     "delete last value",
			initial:  []string{"only"},
			value:    "only",
			wantOk:   true,
			expected: map[string]struct{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			ok := s.Delete(tt.value)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.expected, s.Values.GetTyped())
		})
	}
}

func TestUseSet_Has_ChecksExistence(t *testing.T) {
	tests := []struct {
		name    string
		initial []string
		value   string
		want    bool
	}{
		{
			name:    "value exists",
			initial: []string{"a", "b", "c"},
			value:   "b",
			want:    true,
		},
		{
			name:    "value does not exist",
			initial: []string{"a", "b"},
			value:   "c",
			want:    false,
		},
		{
			name:    "empty set",
			initial: []string{},
			value:   "any",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			got := s.Has(tt.value)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUseSet_Toggle_AddsOrRemoves(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		value    string
		expected map[string]struct{}
	}{
		{
			name:     "toggle adds when not present",
			initial:  []string{"a", "b"},
			value:    "c",
			expected: map[string]struct{}{"a": {}, "b": {}, "c": {}},
		},
		{
			name:     "toggle removes when present",
			initial:  []string{"a", "b", "c"},
			value:    "b",
			expected: map[string]struct{}{"a": {}, "c": {}},
		},
		{
			name:     "toggle on empty set adds",
			initial:  []string{},
			value:    "new",
			expected: map[string]struct{}{"new": {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			s.Toggle(tt.value)

			assert.Equal(t, tt.expected, s.Values.GetTyped())
		})
	}
}

func TestUseSet_Clear_EmptiesSet(t *testing.T) {
	tests := []struct {
		name    string
		initial []string
	}{
		{
			name:    "clear non-empty set",
			initial: []string{"a", "b", "c"},
		},
		{
			name:    "clear empty set",
			initial: []string{},
		},
		{
			name:    "clear single value set",
			initial: []string{"only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			s.Clear()

			assert.Equal(t, map[string]struct{}{}, s.Values.GetTyped())
			assert.Equal(t, 0, s.Size.GetTyped())
			assert.True(t, s.IsEmpty.GetTyped())
		})
	}
}

func TestUseSet_ToSlice_ReturnsValues(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		expected []string
	}{
		{
			name:     "empty set",
			initial:  []string{},
			expected: []string{},
		},
		{
			name:     "single value",
			initial:  []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "multiple values",
			initial:  []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			slice := s.ToSlice()

			// Sort both for comparison (map iteration order is random)
			sort.Strings(slice)
			sort.Strings(tt.expected)
			assert.Equal(t, tt.expected, slice)
		})
	}
}

func TestUseSet_Size_ComputedCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		expected int
	}{
		{
			name:     "empty set",
			initial:  []string{},
			expected: 0,
		},
		{
			name:     "single value",
			initial:  []string{"a"},
			expected: 1,
		},
		{
			name:     "multiple values",
			initial:  []string{"a", "b", "c"},
			expected: 3,
		},
		{
			name:     "duplicates counted once",
			initial:  []string{"a", "a", "b", "b", "c"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			assert.Equal(t, tt.expected, s.Size.GetTyped())
		})
	}
}

func TestUseSet_IsEmpty_ComputedCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		expected bool
	}{
		{
			name:     "empty set",
			initial:  []string{},
			expected: true,
		},
		{
			name:     "nil slice",
			initial:  nil,
			expected: true,
		},
		{
			name:     "non-empty set",
			initial:  []string{"a"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UseSet[string](nil, tt.initial)

			assert.Equal(t, tt.expected, s.IsEmpty.GetTyped())
		})
	}
}

func TestUseSet_SizeUpdatesAfterOperations(t *testing.T) {
	s := UseSet[string](nil, []string{"a"})

	assert.Equal(t, 1, s.Size.GetTyped())

	s.Add("b")
	assert.Equal(t, 2, s.Size.GetTyped())

	s.Add("c")
	assert.Equal(t, 3, s.Size.GetTyped())

	s.Delete("a")
	assert.Equal(t, 2, s.Size.GetTyped())

	s.Clear()
	assert.Equal(t, 0, s.Size.GetTyped())
}

func TestUseSet_IsEmptyUpdatesAfterOperations(t *testing.T) {
	s := UseSet[string](nil, nil)

	assert.True(t, s.IsEmpty.GetTyped())

	s.Add("a")
	assert.False(t, s.IsEmpty.GetTyped())

	s.Delete("a")
	assert.True(t, s.IsEmpty.GetTyped())

	s.Add("b")
	s.Add("c")
	assert.False(t, s.IsEmpty.GetTyped())

	s.Clear()
	assert.True(t, s.IsEmpty.GetTyped())
}

func TestUseSet_GenericTypes(t *testing.T) {
	t.Run("int values", func(t *testing.T) {
		s := UseSet[int](nil, []int{1, 2, 3})

		assert.True(t, s.Has(1))
		assert.True(t, s.Has(2))
		assert.False(t, s.Has(4))

		s.Add(4)
		assert.True(t, s.Has(4))
		assert.Equal(t, 4, s.Size.GetTyped())
	})

	t.Run("custom comparable type", func(t *testing.T) {
		type Status string
		const (
			Active   Status = "active"
			Inactive Status = "inactive"
			Pending  Status = "pending"
		)

		s := UseSet[Status](nil, []Status{Active, Inactive})

		assert.True(t, s.Has(Active))
		assert.False(t, s.Has(Pending))

		s.Add(Pending)
		assert.True(t, s.Has(Pending))
	})
}

func TestUseSet_ThreadSafety(t *testing.T) {
	s := UseSet[int](nil, nil)
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent adds (0 to iterations-1)
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Add(val)
		}(i)
	}

	// Concurrent Has checks (read-only, no effect on size)
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			_ = s.Has(val)
		}(i)
	}

	wg.Wait()

	// All values from 0 to iterations-1 should be present
	assert.Equal(t, iterations, s.Size.GetTyped())

	// Test concurrent toggles separately to avoid race with adds
	s2 := UseSet[int](nil, nil)
	var wg2 sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg2.Add(1)
		go func(val int) {
			defer wg2.Done()
			s2.Toggle(val)
		}(i)
	}

	wg2.Wait()

	// All values should be present (each toggled once = added)
	assert.Equal(t, iterations, s2.Size.GetTyped())
}

func TestUseSet_WorksWithContext(t *testing.T) {
	ctx := createTestContext()

	s := UseSet[string](ctx, []string{"test"})

	require.NotNil(t, s)
	assert.True(t, s.Has("test"))
}

func TestUseSet_ValuesRefIsReactive(t *testing.T) {
	s := UseSet[string](nil, []string{"a"})

	// Track changes via Watch
	changeCount := 0
	bubbly.Watch(s.Values, func(newVal, oldVal map[string]struct{}) {
		changeCount++
	})

	s.Add("b")
	assert.Equal(t, 1, changeCount)

	s.Delete("a")
	assert.Equal(t, 2, changeCount)

	s.Toggle("c")
	assert.Equal(t, 3, changeCount)

	s.Clear()
	assert.Equal(t, 4, changeCount)
}

func TestUseSet_AddDoesNotMutateOriginal(t *testing.T) {
	original := []string{"a", "b"}
	s := UseSet[string](nil, original)

	s.Add("c")

	// Original slice should not be modified (though this is less of a concern
	// since we convert to map internally)
	assert.Equal(t, []string{"a", "b"}, original)
}

func TestUseSet_ToggleMultipleTimes(t *testing.T) {
	s := UseSet[string](nil, nil)

	// Toggle adds
	s.Toggle("a")
	assert.True(t, s.Has("a"))
	assert.Equal(t, 1, s.Size.GetTyped())

	// Toggle removes
	s.Toggle("a")
	assert.False(t, s.Has("a"))
	assert.Equal(t, 0, s.Size.GetTyped())

	// Toggle adds again
	s.Toggle("a")
	assert.True(t, s.Has("a"))
	assert.Equal(t, 1, s.Size.GetTyped())
}

func TestUseSet_EmptyStringValue(t *testing.T) {
	s := UseSet[string](nil, nil)

	s.Add("")

	assert.True(t, s.Has(""))
	assert.Equal(t, 1, s.Size.GetTyped())

	ok := s.Delete("")
	assert.True(t, ok)
	assert.False(t, s.Has(""))
}

func TestUseSet_ZeroValue(t *testing.T) {
	s := UseSet[int](nil, nil)

	s.Add(0)

	assert.True(t, s.Has(0))
	assert.Equal(t, 1, s.Size.GetTyped())

	s.Toggle(0)
	assert.False(t, s.Has(0))
}

func TestUseSet_AddExistingValueNoChange(t *testing.T) {
	s := UseSet[string](nil, []string{"a", "b"})

	// Track changes
	changeCount := 0
	bubbly.Watch(s.Values, func(newVal, oldVal map[string]struct{}) {
		changeCount++
	})

	// Add existing value - should still trigger change (we create new map)
	// This is consistent with UseMap behavior
	s.Add("a")

	// Size should remain the same
	assert.Equal(t, 2, s.Size.GetTyped())
}

func TestUseSet_DeleteNonExistingValueNoChange(t *testing.T) {
	s := UseSet[string](nil, []string{"a", "b"})

	// Delete non-existing should return false
	ok := s.Delete("c")
	assert.False(t, ok)

	// Set should be unchanged
	assert.Equal(t, 2, s.Size.GetTyped())
	assert.True(t, s.Has("a"))
	assert.True(t, s.Has("b"))
}
