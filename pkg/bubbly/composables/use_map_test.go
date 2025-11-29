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
// UseMap Tests - Task 4.2
// =============================================================================

func TestUseMap_InitialDataSetCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		expected map[string]int
	}{
		{
			name:     "empty map",
			initial:  map[string]int{},
			expected: map[string]int{},
		},
		{
			name:     "nil map becomes empty",
			initial:  nil,
			expected: map[string]int{},
		},
		{
			name:     "map with values",
			initial:  map[string]int{"a": 1, "b": 2, "c": 3},
			expected: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name:     "single entry",
			initial:  map[string]int{"key": 42},
			expected: map[string]int{"key": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			require.NotNil(t, m)
			require.NotNil(t, m.Data)
			assert.Equal(t, tt.expected, m.Data.GetTyped())
		})
	}
}

func TestUseMap_Get_ReturnsValue(t *testing.T) {
	tests := []struct {
		name      string
		initial   map[string]int
		key       string
		wantValue int
		wantOk    bool
	}{
		{
			name:      "existing key",
			initial:   map[string]int{"a": 1, "b": 2},
			key:       "a",
			wantValue: 1,
			wantOk:    true,
		},
		{
			name:      "non-existing key",
			initial:   map[string]int{"a": 1, "b": 2},
			key:       "c",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "empty map",
			initial:   map[string]int{},
			key:       "any",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "key with zero value",
			initial:   map[string]int{"zero": 0},
			key:       "zero",
			wantValue: 0,
			wantOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			value, ok := m.Get(tt.key)

			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestUseMap_Set_AddsOrUpdatesKey(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		key      string
		value    int
		expected map[string]int
	}{
		{
			name:     "add new key to empty map",
			initial:  map[string]int{},
			key:      "a",
			value:    1,
			expected: map[string]int{"a": 1},
		},
		{
			name:     "add new key to existing map",
			initial:  map[string]int{"a": 1},
			key:      "b",
			value:    2,
			expected: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "update existing key",
			initial:  map[string]int{"a": 1},
			key:      "a",
			value:    10,
			expected: map[string]int{"a": 10},
		},
		{
			name:     "set zero value",
			initial:  map[string]int{"a": 1},
			key:      "b",
			value:    0,
			expected: map[string]int{"a": 1, "b": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			m.Set(tt.key, tt.value)

			assert.Equal(t, tt.expected, m.Data.GetTyped())
		})
	}
}

func TestUseMap_Delete_RemovesKey(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		key      string
		wantOk   bool
		expected map[string]int
	}{
		{
			name:     "delete existing key",
			initial:  map[string]int{"a": 1, "b": 2},
			key:      "a",
			wantOk:   true,
			expected: map[string]int{"b": 2},
		},
		{
			name:     "delete non-existing key",
			initial:  map[string]int{"a": 1},
			key:      "b",
			wantOk:   false,
			expected: map[string]int{"a": 1},
		},
		{
			name:     "delete from empty map",
			initial:  map[string]int{},
			key:      "any",
			wantOk:   false,
			expected: map[string]int{},
		},
		{
			name:     "delete last key",
			initial:  map[string]int{"only": 1},
			key:      "only",
			wantOk:   true,
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			ok := m.Delete(tt.key)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.expected, m.Data.GetTyped())
		})
	}
}

func TestUseMap_Has_ChecksExistence(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]int
		key     string
		want    bool
	}{
		{
			name:    "key exists",
			initial: map[string]int{"a": 1, "b": 2},
			key:     "a",
			want:    true,
		},
		{
			name:    "key does not exist",
			initial: map[string]int{"a": 1},
			key:     "b",
			want:    false,
		},
		{
			name:    "empty map",
			initial: map[string]int{},
			key:     "any",
			want:    false,
		},
		{
			name:    "key with zero value exists",
			initial: map[string]int{"zero": 0},
			key:     "zero",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			got := m.Has(tt.key)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUseMap_Keys_ReturnsAllKeys(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		expected []string
	}{
		{
			name:     "empty map",
			initial:  map[string]int{},
			expected: []string{},
		},
		{
			name:     "single key",
			initial:  map[string]int{"a": 1},
			expected: []string{"a"},
		},
		{
			name:     "multiple keys",
			initial:  map[string]int{"a": 1, "b": 2, "c": 3},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			keys := m.Keys()

			// Sort both for comparison (map iteration order is random)
			sort.Strings(keys)
			sort.Strings(tt.expected)
			assert.Equal(t, tt.expected, keys)
		})
	}
}

func TestUseMap_Values_ReturnsAllValues(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		expected []int
	}{
		{
			name:     "empty map",
			initial:  map[string]int{},
			expected: []int{},
		},
		{
			name:     "single value",
			initial:  map[string]int{"a": 1},
			expected: []int{1},
		},
		{
			name:     "multiple values",
			initial:  map[string]int{"a": 1, "b": 2, "c": 3},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			values := m.Values()

			// Sort both for comparison (map iteration order is random)
			sort.Ints(values)
			sort.Ints(tt.expected)
			assert.Equal(t, tt.expected, values)
		})
	}
}

func TestUseMap_Clear_EmptiesMap(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]int
	}{
		{
			name:    "clear non-empty map",
			initial: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name:    "clear empty map",
			initial: map[string]int{},
		},
		{
			name:    "clear single entry map",
			initial: map[string]int{"only": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			m.Clear()

			assert.Equal(t, map[string]int{}, m.Data.GetTyped())
			assert.Equal(t, 0, m.Size.GetTyped())
			assert.True(t, m.IsEmpty.GetTyped())
		})
	}
}

func TestUseMap_Size_ComputedCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		expected int
	}{
		{
			name:     "empty map",
			initial:  map[string]int{},
			expected: 0,
		},
		{
			name:     "single entry",
			initial:  map[string]int{"a": 1},
			expected: 1,
		},
		{
			name:     "multiple entries",
			initial:  map[string]int{"a": 1, "b": 2, "c": 3},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			assert.Equal(t, tt.expected, m.Size.GetTyped())
		})
	}
}

func TestUseMap_IsEmpty_ComputedCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int
		expected bool
	}{
		{
			name:     "empty map",
			initial:  map[string]int{},
			expected: true,
		},
		{
			name:     "nil map",
			initial:  nil,
			expected: true,
		},
		{
			name:     "non-empty map",
			initial:  map[string]int{"a": 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UseMap[string, int](nil, tt.initial)

			assert.Equal(t, tt.expected, m.IsEmpty.GetTyped())
		})
	}
}

func TestUseMap_SizeUpdatesAfterOperations(t *testing.T) {
	m := UseMap[string, int](nil, map[string]int{"a": 1})

	assert.Equal(t, 1, m.Size.GetTyped())

	m.Set("b", 2)
	assert.Equal(t, 2, m.Size.GetTyped())

	m.Set("c", 3)
	assert.Equal(t, 3, m.Size.GetTyped())

	m.Delete("a")
	assert.Equal(t, 2, m.Size.GetTyped())

	m.Clear()
	assert.Equal(t, 0, m.Size.GetTyped())
}

func TestUseMap_IsEmptyUpdatesAfterOperations(t *testing.T) {
	m := UseMap[string, int](nil, nil)

	assert.True(t, m.IsEmpty.GetTyped())

	m.Set("a", 1)
	assert.False(t, m.IsEmpty.GetTyped())

	m.Delete("a")
	assert.True(t, m.IsEmpty.GetTyped())

	m.Set("b", 2)
	m.Set("c", 3)
	assert.False(t, m.IsEmpty.GetTyped())

	m.Clear()
	assert.True(t, m.IsEmpty.GetTyped())
}

func TestUseMap_GenericTypes(t *testing.T) {
	t.Run("int keys", func(t *testing.T) {
		m := UseMap[int, string](nil, map[int]string{1: "one", 2: "two"})

		val, ok := m.Get(1)
		assert.True(t, ok)
		assert.Equal(t, "one", val)

		m.Set(3, "three")
		assert.True(t, m.Has(3))
	})

	t.Run("struct values", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		m := UseMap[string, Person](nil, map[string]Person{
			"alice": {Name: "Alice", Age: 30},
		})

		val, ok := m.Get("alice")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val.Name)
		assert.Equal(t, 30, val.Age)

		m.Set("bob", Person{Name: "Bob", Age: 25})
		assert.Equal(t, 2, m.Size.GetTyped())
	})
}

func TestUseMap_ThreadSafety(t *testing.T) {
	m := UseMap[int, int](nil, nil)
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			m.Set(val, val*10)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			_, _ = m.Get(key)
		}(i)
	}

	// Concurrent Has checks
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			_ = m.Has(key)
		}(i)
	}

	wg.Wait()

	// All keys should be present
	assert.Equal(t, iterations, m.Size.GetTyped())
}

func TestUseMap_WorksWithContext(t *testing.T) {
	ctx := createTestContext()

	m := UseMap[string, int](ctx, map[string]int{"test": 42})

	require.NotNil(t, m)
	assert.Equal(t, 42, m.Data.GetTyped()["test"])
}

func TestUseMap_DataRefIsReactive(t *testing.T) {
	m := UseMap[string, int](nil, map[string]int{"a": 1})

	// Track changes via Watch
	changeCount := 0
	bubbly.Watch(m.Data, func(newVal, oldVal map[string]int) {
		changeCount++
	})

	m.Set("b", 2)
	assert.Equal(t, 1, changeCount)

	m.Delete("a")
	assert.Equal(t, 2, changeCount)

	m.Clear()
	assert.Equal(t, 3, changeCount)
}

func TestUseMap_SetDoesNotMutateOriginal(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}
	m := UseMap[string, int](nil, original)

	m.Set("c", 3)

	// Original should not be modified
	_, exists := original["c"]
	assert.False(t, exists, "original map should not be modified")
}

func TestUseMap_DeleteDoesNotMutateOriginal(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}
	m := UseMap[string, int](nil, original)

	m.Delete("a")

	// Original should not be modified
	_, exists := original["a"]
	assert.True(t, exists, "original map should not be modified")
}

func TestUseMap_UpdateExistingKey(t *testing.T) {
	m := UseMap[string, int](nil, map[string]int{"a": 1})

	// Update same key multiple times
	m.Set("a", 10)
	assert.Equal(t, 10, m.Data.GetTyped()["a"])

	m.Set("a", 100)
	assert.Equal(t, 100, m.Data.GetTyped()["a"])

	// Size should remain 1
	assert.Equal(t, 1, m.Size.GetTyped())
}

func TestUseMap_EmptyStringKey(t *testing.T) {
	m := UseMap[string, int](nil, nil)

	m.Set("", 42)

	val, ok := m.Get("")
	assert.True(t, ok)
	assert.Equal(t, 42, val)
	assert.True(t, m.Has(""))
}

func TestUseMap_NilValueType(t *testing.T) {
	m := UseMap[string, *int](nil, nil)

	// Set nil value
	m.Set("nilKey", nil)

	val, ok := m.Get("nilKey")
	assert.True(t, ok)
	assert.Nil(t, val)
	assert.True(t, m.Has("nilKey"))

	// Set non-nil value
	num := 42
	m.Set("numKey", &num)

	val, ok = m.Get("numKey")
	assert.True(t, ok)
	assert.Equal(t, 42, *val)
}
