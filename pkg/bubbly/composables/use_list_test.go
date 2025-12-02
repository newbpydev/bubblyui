package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseList_InitialItems tests that initial items are set correctly.
func TestUseList_InitialItems(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		expected []string
	}{
		{
			name:     "empty initial",
			initial:  []string{},
			expected: []string{},
		},
		{
			name:     "nil initial",
			initial:  nil,
			expected: []string{},
		},
		{
			name:     "single item",
			initial:  []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "multiple items",
			initial:  []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			assert.Equal(t, tt.expected, list.Items.GetTyped())
			assert.Equal(t, len(tt.expected), list.Length.GetTyped())
			assert.Equal(t, len(tt.expected) == 0, list.IsEmpty.GetTyped())
		})
	}
}

// TestUseList_Push tests adding items to the end.
func TestUseList_Push(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		push     []int
		expected []int
	}{
		{
			name:     "push to empty",
			initial:  []int{},
			push:     []int{1},
			expected: []int{1},
		},
		{
			name:     "push single",
			initial:  []int{1, 2},
			push:     []int{3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "push multiple",
			initial:  []int{1},
			push:     []int{2, 3, 4},
			expected: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			list.Push(tt.push...)

			assert.Equal(t, tt.expected, list.Items.GetTyped())
			assert.Equal(t, len(tt.expected), list.Length.GetTyped())
		})
	}
}

// TestUseList_Pop tests removing and returning the last item.
func TestUseList_Pop(t *testing.T) {
	tests := []struct {
		name         string
		initial      []string
		expectedItem string
		expectedOk   bool
		expectedList []string
	}{
		{
			name:         "pop from empty",
			initial:      []string{},
			expectedItem: "",
			expectedOk:   false,
			expectedList: []string{},
		},
		{
			name:         "pop from single",
			initial:      []string{"a"},
			expectedItem: "a",
			expectedOk:   true,
			expectedList: []string{},
		},
		{
			name:         "pop from multiple",
			initial:      []string{"a", "b", "c"},
			expectedItem: "c",
			expectedOk:   true,
			expectedList: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			item, ok := list.Pop()

			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedList, list.Items.GetTyped())
		})
	}
}

// TestUseList_Shift tests removing and returning the first item.
func TestUseList_Shift(t *testing.T) {
	tests := []struct {
		name         string
		initial      []int
		expectedItem int
		expectedOk   bool
		expectedList []int
	}{
		{
			name:         "shift from empty",
			initial:      []int{},
			expectedItem: 0,
			expectedOk:   false,
			expectedList: []int{},
		},
		{
			name:         "shift from single",
			initial:      []int{1},
			expectedItem: 1,
			expectedOk:   true,
			expectedList: []int{},
		},
		{
			name:         "shift from multiple",
			initial:      []int{1, 2, 3},
			expectedItem: 1,
			expectedOk:   true,
			expectedList: []int{2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			item, ok := list.Shift()

			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedList, list.Items.GetTyped())
		})
	}
}

// TestUseList_Unshift tests adding items to the beginning.
func TestUseList_Unshift(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string
		unshift  []string
		expected []string
	}{
		{
			name:     "unshift to empty",
			initial:  []string{},
			unshift:  []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "unshift single",
			initial:  []string{"b", "c"},
			unshift:  []string{"a"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "unshift multiple",
			initial:  []string{"c"},
			unshift:  []string{"a", "b"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			list.Unshift(tt.unshift...)

			assert.Equal(t, tt.expected, list.Items.GetTyped())
		})
	}
}

// TestUseList_Insert tests inserting an item at a specific index.
func TestUseList_Insert(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		index    int
		item     int
		expected []int
	}{
		{
			name:     "insert at start",
			initial:  []int{2, 3},
			index:    0,
			item:     1,
			expected: []int{1, 2, 3},
		},
		{
			name:     "insert at middle",
			initial:  []int{1, 3},
			index:    1,
			item:     2,
			expected: []int{1, 2, 3},
		},
		{
			name:     "insert at end",
			initial:  []int{1, 2},
			index:    2,
			item:     3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "insert into empty",
			initial:  []int{},
			index:    0,
			item:     1,
			expected: []int{1},
		},
		{
			name:     "insert negative index (clamp to 0)",
			initial:  []int{2, 3},
			index:    -1,
			item:     1,
			expected: []int{1, 2, 3},
		},
		{
			name:     "insert beyond length (clamp to end)",
			initial:  []int{1, 2},
			index:    10,
			item:     3,
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			list.Insert(tt.index, tt.item)

			assert.Equal(t, tt.expected, list.Items.GetTyped())
		})
	}
}

// TestUseList_RemoveAt tests removing an item at a specific index.
func TestUseList_RemoveAt(t *testing.T) {
	tests := []struct {
		name         string
		initial      []string
		index        int
		expectedItem string
		expectedOk   bool
		expectedList []string
	}{
		{
			name:         "remove from empty",
			initial:      []string{},
			index:        0,
			expectedItem: "",
			expectedOk:   false,
			expectedList: []string{},
		},
		{
			name:         "remove at start",
			initial:      []string{"a", "b", "c"},
			index:        0,
			expectedItem: "a",
			expectedOk:   true,
			expectedList: []string{"b", "c"},
		},
		{
			name:         "remove at middle",
			initial:      []string{"a", "b", "c"},
			index:        1,
			expectedItem: "b",
			expectedOk:   true,
			expectedList: []string{"a", "c"},
		},
		{
			name:         "remove at end",
			initial:      []string{"a", "b", "c"},
			index:        2,
			expectedItem: "c",
			expectedOk:   true,
			expectedList: []string{"a", "b"},
		},
		{
			name:         "remove negative index",
			initial:      []string{"a", "b"},
			index:        -1,
			expectedItem: "",
			expectedOk:   false,
			expectedList: []string{"a", "b"},
		},
		{
			name:         "remove beyond length",
			initial:      []string{"a", "b"},
			index:        5,
			expectedItem: "",
			expectedOk:   false,
			expectedList: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			item, ok := list.RemoveAt(tt.index)

			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedList, list.Items.GetTyped())
		})
	}
}

// TestUseList_UpdateAt tests updating an item at a specific index.
func TestUseList_UpdateAt(t *testing.T) {
	tests := []struct {
		name       string
		initial    []int
		index      int
		newValue   int
		expectedOk bool
		expected   []int
	}{
		{
			name:       "update at start",
			initial:    []int{1, 2, 3},
			index:      0,
			newValue:   10,
			expectedOk: true,
			expected:   []int{10, 2, 3},
		},
		{
			name:       "update at middle",
			initial:    []int{1, 2, 3},
			index:      1,
			newValue:   20,
			expectedOk: true,
			expected:   []int{1, 20, 3},
		},
		{
			name:       "update at end",
			initial:    []int{1, 2, 3},
			index:      2,
			newValue:   30,
			expectedOk: true,
			expected:   []int{1, 2, 30},
		},
		{
			name:       "update empty list",
			initial:    []int{},
			index:      0,
			newValue:   1,
			expectedOk: false,
			expected:   []int{},
		},
		{
			name:       "update negative index",
			initial:    []int{1, 2},
			index:      -1,
			newValue:   10,
			expectedOk: false,
			expected:   []int{1, 2},
		},
		{
			name:       "update beyond length",
			initial:    []int{1, 2},
			index:      5,
			newValue:   10,
			expectedOk: false,
			expected:   []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			ok := list.UpdateAt(tt.index, tt.newValue)

			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expected, list.Items.GetTyped())
		})
	}
}

// TestUseList_Remove tests removing the first occurrence of an item.
func TestUseList_Remove(t *testing.T) {
	eq := func(a, b string) bool { return a == b }

	tests := []struct {
		name         string
		initial      []string
		item         string
		expectedOk   bool
		expectedList []string
	}{
		{
			name:         "remove from empty",
			initial:      []string{},
			item:         "a",
			expectedOk:   false,
			expectedList: []string{},
		},
		{
			name:         "remove existing item",
			initial:      []string{"a", "b", "c"},
			item:         "b",
			expectedOk:   true,
			expectedList: []string{"a", "c"},
		},
		{
			name:         "remove non-existing item",
			initial:      []string{"a", "b", "c"},
			item:         "d",
			expectedOk:   false,
			expectedList: []string{"a", "b", "c"},
		},
		{
			name:         "remove first occurrence only",
			initial:      []string{"a", "b", "a", "c"},
			item:         "a",
			expectedOk:   true,
			expectedList: []string{"b", "a", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			ok := list.Remove(tt.item, eq)

			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedList, list.Items.GetTyped())
		})
	}
}

// TestUseList_Clear tests clearing all items.
func TestUseList_Clear(t *testing.T) {
	tests := []struct {
		name    string
		initial []int
	}{
		{
			name:    "clear empty",
			initial: []int{},
		},
		{
			name:    "clear non-empty",
			initial: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			list.Clear()

			assert.Equal(t, []int{}, list.Items.GetTyped())
			assert.Equal(t, 0, list.Length.GetTyped())
			assert.True(t, list.IsEmpty.GetTyped())
		})
	}
}

// TestUseList_Get tests getting an item at a specific index.
func TestUseList_Get(t *testing.T) {
	tests := []struct {
		name         string
		initial      []string
		index        int
		expectedItem string
		expectedOk   bool
	}{
		{
			name:         "get from empty",
			initial:      []string{},
			index:        0,
			expectedItem: "",
			expectedOk:   false,
		},
		{
			name:         "get at start",
			initial:      []string{"a", "b", "c"},
			index:        0,
			expectedItem: "a",
			expectedOk:   true,
		},
		{
			name:         "get at middle",
			initial:      []string{"a", "b", "c"},
			index:        1,
			expectedItem: "b",
			expectedOk:   true,
		},
		{
			name:         "get at end",
			initial:      []string{"a", "b", "c"},
			index:        2,
			expectedItem: "c",
			expectedOk:   true,
		},
		{
			name:         "get negative index",
			initial:      []string{"a", "b"},
			index:        -1,
			expectedItem: "",
			expectedOk:   false,
		},
		{
			name:         "get beyond length",
			initial:      []string{"a", "b"},
			index:        5,
			expectedItem: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			item, ok := list.Get(tt.index)

			assert.Equal(t, tt.expectedItem, item)
			assert.Equal(t, tt.expectedOk, ok)
		})
	}
}

// TestUseList_Set tests setting the entire list.
func TestUseList_Set(t *testing.T) {
	tests := []struct {
		name     string
		initial  []int
		newItems []int
		expected []int
	}{
		{
			name:     "set empty to non-empty",
			initial:  []int{},
			newItems: []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "set non-empty to empty",
			initial:  []int{1, 2, 3},
			newItems: []int{},
			expected: []int{},
		},
		{
			name:     "set to nil",
			initial:  []int{1, 2, 3},
			newItems: nil,
			expected: []int{},
		},
		{
			name:     "replace items",
			initial:  []int{1, 2, 3},
			newItems: []int{4, 5},
			expected: []int{4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			list := UseList(ctx, tt.initial)

			list.Set(tt.newItems)

			assert.Equal(t, tt.expected, list.Items.GetTyped())
			assert.Equal(t, len(tt.expected), list.Length.GetTyped())
		})
	}
}

// TestUseList_Length tests the Length computed value.
func TestUseList_Length(t *testing.T) {
	ctx := createTestContext()
	list := UseList(ctx, []int{1, 2, 3})

	assert.Equal(t, 3, list.Length.GetTyped())

	list.Push(4)
	assert.Equal(t, 4, list.Length.GetTyped())

	list.Pop()
	assert.Equal(t, 3, list.Length.GetTyped())

	list.Clear()
	assert.Equal(t, 0, list.Length.GetTyped())
}

// TestUseList_IsEmpty tests the IsEmpty computed value.
func TestUseList_IsEmpty(t *testing.T) {
	ctx := createTestContext()
	list := UseList(ctx, []int{})

	assert.True(t, list.IsEmpty.GetTyped())

	list.Push(1)
	assert.False(t, list.IsEmpty.GetTyped())

	list.Pop()
	assert.True(t, list.IsEmpty.GetTyped())
}

// TestUseList_GenericTypes tests that UseList works with various types.
func TestUseList_GenericTypes(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		ctx := createTestContext()
		list := UseList(ctx, []int{1, 2, 3})
		list.Push(4)
		assert.Equal(t, []int{1, 2, 3, 4}, list.Items.GetTyped())
	})

	t.Run("string slice", func(t *testing.T) {
		ctx := createTestContext()
		list := UseList(ctx, []string{"a", "b"})
		list.Push("c")
		assert.Equal(t, []string{"a", "b", "c"}, list.Items.GetTyped())
	})

	t.Run("struct slice", func(t *testing.T) {
		type Item struct {
			ID   int
			Name string
		}
		ctx := createTestContext()
		list := UseList(ctx, []Item{{ID: 1, Name: "one"}})
		list.Push(Item{ID: 2, Name: "two"})
		assert.Equal(t, 2, list.Length.GetTyped())
		item, ok := list.Get(1)
		require.True(t, ok)
		assert.Equal(t, "two", item.Name)
	})
}

// TestUseList_NilContext tests that UseList handles nil context.
func TestUseList_NilContext(t *testing.T) {
	// Should not panic with nil context
	list := UseList[int](nil, []int{1, 2, 3})
	assert.NotNil(t, list)
	assert.Equal(t, []int{1, 2, 3}, list.Items.GetTyped())
}

// TestUseList_ChainedOperations tests multiple operations in sequence.
func TestUseList_ChainedOperations(t *testing.T) {
	ctx := createTestContext()
	list := UseList(ctx, []int{})

	// Build up list
	list.Push(1, 2, 3)
	list.Unshift(0)
	list.Insert(2, 10)

	assert.Equal(t, []int{0, 1, 10, 2, 3}, list.Items.GetTyped())

	// Remove some items
	list.RemoveAt(2) // Remove 10
	list.Pop()       // Remove 3
	list.Shift()     // Remove 0

	assert.Equal(t, []int{1, 2}, list.Items.GetTyped())

	// Update
	list.UpdateAt(0, 100)
	assert.Equal(t, []int{100, 2}, list.Items.GetTyped())

	// Replace all
	list.Set([]int{5, 6, 7})
	assert.Equal(t, []int{5, 6, 7}, list.Items.GetTyped())

	// Clear
	list.Clear()
	assert.True(t, list.IsEmpty.GetTyped())
}

// TestUseList_Reactivity tests that Items ref triggers updates.
func TestUseList_Reactivity(t *testing.T) {
	ctx := createTestContext()
	list := UseList(ctx, []int{1, 2, 3})

	// Track changes via Watch
	changes := 0
	bubbly.Watch(list.Items, func(newVal, oldVal []int) {
		changes++
	})

	list.Push(4)
	list.Pop()
	list.Clear()

	// Each operation should trigger a change
	assert.Equal(t, 3, changes)
}

// TestUseList_CreateSharedPattern tests integration with CreateShared.
func TestUseList_CreateSharedPattern(t *testing.T) {
	// Create a shared list factory
	UseSharedList := CreateShared(func(ctx *bubbly.Context) *ListReturn[string] {
		return UseList(ctx, []string{"shared"})
	})

	ctx := createTestContext()
	list1 := UseSharedList(ctx)
	list2 := UseSharedList(ctx)

	// Both should reference the same instance
	assert.Same(t, list1, list2)

	// Modifications through one should be visible through the other
	list1.Push("item")
	assert.Equal(t, []string{"shared", "item"}, list2.Items.GetTyped())
}
