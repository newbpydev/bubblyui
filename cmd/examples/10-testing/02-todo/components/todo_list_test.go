package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/02-todo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestTodoList_BasicMounting tests component initialization
// Shows: Component creation, mounting, basic rendering
func TestTodoList_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todosRef := ctx.Ref([]composables.Todo{})

	// Act: Create and mount component
	list, err := CreateTodoList(TodoListProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Todos:    todosRef,
		OnToggle: func(id int64) {},
		OnRemove: func(id int64) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Assert: Component renders empty state
	ct.AssertRenderContains("No todos yet")
}

// TestTodoList_RenderOutput demonstrates render testing with different states
// Shows: Table-driven tests, render assertions
func TestTodoList_RenderOutput(t *testing.T) {
	tests := []struct {
		name     string
		todos    []composables.Todo
		expected []string
	}{
		{
			name:  "empty list",
			todos: []composables.Todo{},
			expected: []string{
				"No todos yet",
			},
		},
		{
			name: "single incomplete todo",
			todos: []composables.Todo{
				{ID: 1, Title: "Buy groceries", Completed: false},
			},
			expected: []string{
				"Todo List",
				"Buy groceries",
				"○", // Incomplete icon
			},
		},
		{
			name: "single completed todo",
			todos: []composables.Todo{
				{ID: 1, Title: "Buy groceries", Completed: true},
			},
			expected: []string{
				"Todo List",
				"Buy groceries",
				"✓", // Completed icon
			},
		},
		{
			name: "multiple todos mixed",
			todos: []composables.Todo{
				{ID: 1, Title: "Task 1", Completed: false},
				{ID: 2, Title: "Task 2", Completed: true},
				{ID: 3, Title: "Task 3", Completed: false},
			},
			expected: []string{
				"Todo List",
				"Task 1",
				"Task 2",
				"Task 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			ctx := createTestContextForComponent()
			todosRef := ctx.Ref(tt.todos)

			list, err := CreateTodoList(TodoListProps{
				Focused:  bubbly.NewRef[interface{}](false),
				Todos:    todosRef,
				OnToggle: func(id int64) {},
				OnRemove: func(id int64) {},
			})
			require.NoError(t, err)

			ct := harness.Mount(list)

			// Assert: Render contains all expected strings
			for _, expected := range tt.expected {
				ct.AssertRenderContains(expected)
			}
		})
	}
}

// TestTodoList_TodosProp tests todos prop reactivity
// Shows: Prop passing, state inspection, reactive updates
func TestTodoList_TodosProp(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	initialTodos := []composables.Todo{
		{ID: 1, Title: "Initial", Completed: false},
	}
	todosRef := ctx.Ref(initialTodos)

	list, err := CreateTodoList(TodoListProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Todos:    todosRef,
		OnToggle: func(id int64) {},
		OnRemove: func(id int64) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Assert: Initial todos present
	ct.AssertRenderContains("Initial")

	// Act: Update todos
	updatedTodos := []composables.Todo{
		{ID: 1, Title: "Initial", Completed: false},
		{ID: 2, Title: "Added", Completed: false},
	}
	todosRef.Set(updatedTodos)

	// Assert: Updated todos present
	ct.AssertRenderContains("Added")
}

// TestTodoList_OnToggleCallback tests toggle callback
// Shows: Event handler testing, callback verification
func TestTodoList_OnToggleCallback(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todos := []composables.Todo{
		{ID: 123, Title: "Test", Completed: false},
	}
	todosRef := ctx.Ref(todos)

	var toggledID int64
	toggleCalled := false

	list, err := CreateTodoList(TodoListProps{
		Focused: bubbly.NewRef[interface{}](false),
		Todos:   todosRef,
		OnToggle: func(id int64) {
			toggleCalled = true
			toggledID = id
		},
		OnRemove: func(id int64) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Act: Emit toggle event
	ct.Emit("toggle", int64(123))

	// Assert: Callback was called with correct ID
	assert.True(t, toggleCalled, "OnToggle callback should have been called")
	assert.Equal(t, int64(123), toggledID)
}

// TestTodoList_OnRemoveCallback tests remove callback
// Shows: Event handler testing, callback verification
func TestTodoList_OnRemoveCallback(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todos := []composables.Todo{
		{ID: 456, Title: "Test", Completed: false},
	}
	todosRef := ctx.Ref(todos)

	var removedID int64
	removeCalled := false

	list, err := CreateTodoList(TodoListProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Todos:    todosRef,
		OnToggle: func(id int64) {},
		OnRemove: func(id int64) {
			removeCalled = true
			removedID = id
		},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Act: Emit remove event
	ct.Emit("remove", int64(456))

	// Assert: Callback was called with correct ID
	assert.True(t, removeCalled, "OnRemove callback should have been called")
	assert.Equal(t, int64(456), removedID)
}

// TestTodoList_EventTracking demonstrates event tracking
// Shows: EventTracker usage, multiple events
func TestTodoList_EventTracking(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todos := []composables.Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: false},
	}
	todosRef := ctx.Ref(todos)

	list, err := CreateTodoList(TodoListProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Todos:    todosRef,
		OnToggle: func(id int64) {},
		OnRemove: func(id int64) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Act: Emit multiple events
	ct.Emit("toggle", int64(1))
	ct.Emit("toggle", int64(2))
	ct.Emit("remove", int64(1))

	// Assert: Events were tracked
	ct.AssertEventFired("toggle")
	ct.AssertEventFired("remove")
	ct.AssertEventCount("toggle", 2)
	ct.AssertEventCount("remove", 1)
}

// TestTodoList_NilCallbacks tests component with nil callbacks
// Shows: Edge case handling, defensive programming
func TestTodoList_NilCallbacks(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todos := []composables.Todo{
		{ID: 1, Title: "Test", Completed: false},
	}
	todosRef := ctx.Ref(todos)

	// Act: Create with nil callbacks
	list, err := CreateTodoList(TodoListProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Todos:    todosRef,
		OnToggle: nil,
		OnRemove: nil,
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Act: Emit events (should not panic)
	ct.Emit("toggle", int64(1))
	ct.Emit("remove", int64(1))

	// Assert: Component still renders
	ct.AssertRenderContains("Test")
}

// TestTodoList_InvalidEventData tests handling of invalid event data
// Shows: Error handling, type safety
func TestTodoList_InvalidEventData(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todosRef := ctx.Ref([]composables.Todo{})

	toggleCalled := false
	removeCalled := false

	list, err := CreateTodoList(TodoListProps{
		Focused: bubbly.NewRef[interface{}](false),
		Todos:   todosRef,
		OnToggle: func(id int64) {
			toggleCalled = true
		},
		OnRemove: func(id int64) {
			removeCalled = true
		},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Act: Emit events with wrong data types
	ct.Emit("toggle", "not an int64")
	ct.Emit("remove", 123.45)

	// Assert: Callbacks were NOT called (type assertion failed)
	assert.False(t, toggleCalled, "OnToggle should not be called with wrong type")
	assert.False(t, removeCalled, "OnRemove should not be called with wrong type")
}

// TestTodoList_CompletedStatus tests visual distinction of completed todos
// Shows: Conditional rendering, styling verification
func TestTodoList_CompletedStatus(t *testing.T) {
	tests := []struct {
		name      string
		completed bool
		hasIcon   string
	}{
		{
			name:      "incomplete shows circle",
			completed: false,
			hasIcon:   "○",
		},
		{
			name:      "completed shows checkmark",
			completed: true,
			hasIcon:   "✓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			ctx := createTestContextForComponent()

			todos := []composables.Todo{
				{ID: 1, Title: "Test", Completed: tt.completed},
			}
			todosRef := ctx.Ref(todos)

			list, err := CreateTodoList(TodoListProps{
				Focused:  bubbly.NewRef[interface{}](false),
				Todos:    todosRef,
				OnToggle: func(id int64) {},
				OnRemove: func(id int64) {},
			})
			require.NoError(t, err)

			ct := harness.Mount(list)

			// Assert: Correct icon is shown
			ct.AssertRenderContains(tt.hasIcon)
		})
	}
}

// TestTodoList_Cleanup demonstrates cleanup verification
// Shows: Resource cleanup, unmount behavior
func TestTodoList_Cleanup(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()

	todosRef := ctx.Ref([]composables.Todo{
		{ID: 1, Title: "Test", Completed: false},
	})

	list, err := CreateTodoList(TodoListProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Todos:    todosRef,
		OnToggle: func(id int64) {},
		OnRemove: func(id int64) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(list)

	// Act: Use component
	ct.Emit("toggle", int64(1))

	// Act: Unmount (cleanup happens automatically via t.Cleanup)
	ct.Unmount()

	// Assert: Component is unmounted
	// (In real scenarios, you'd verify resources were released)
}
