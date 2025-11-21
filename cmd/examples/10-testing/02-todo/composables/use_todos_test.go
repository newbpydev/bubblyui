package composables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// NOTE: This file demonstrates TWO testing approaches for composables:
// 1. Direct unit testing (for custom composables like UseTodos)
// 2. Component-wrapper testing (using testutil patterns)

// ============================================================================
// APPROACH 1: Direct Unit Testing (for custom composables)
// ============================================================================

// createTestContext creates a minimal context for direct composable testing
func createTestContext() *bubbly.Context {
	var ctx *bubbly.Context
	component, _ := bubbly.NewComponent("Test").
		Setup(func(c *bubbly.Context) {
			ctx = c
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()
	// CRITICAL: Must call Init() to execute Setup and get the context
	component.Init()
	return ctx
}

// TestUseTodos_Initialization tests initial state
func TestUseTodos_Initialization(t *testing.T) {
	tests := []struct {
		name     string
		initial  []Todo
		expected int
	}{
		{"empty", []Todo{}, 0},
		{"nil", nil, 0},
		{"with items", []Todo{{ID: 1, Title: "Test", Completed: false}}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			todos := UseTodos(ctx, tt.initial)

			assert.Equal(t, tt.expected, todos.Total.Get())
			assert.Equal(t, 0, todos.Completed.Get())
			assert.Equal(t, tt.expected, todos.Remaining.Get())
		})
	}
}

// TestUseTodos_Add tests adding todos
func TestUseTodos_Add(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected int
	}{
		{"valid title", "Buy groceries", 1},
		{"empty title", "", 0}, // Should not add
		{"long title", "This is a very long todo title with many characters", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			todos := UseTodos(ctx, nil)

			todos.Add(tt.title)

			assert.Equal(t, tt.expected, todos.Total.Get())
			if tt.expected > 0 {
				current := todos.Todos.Get().([]Todo)
				assert.Equal(t, tt.title, current[0].Title)
				assert.False(t, current[0].Completed)
				assert.Greater(t, current[0].ID, int64(0))
			}
		})
	}
}

// TestUseTodos_Toggle tests toggling todo completion
func TestUseTodos_Toggle(t *testing.T) {
	ctx := createTestContext()
	initial := []Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: false},
	}
	todos := UseTodos(ctx, initial)

	// Toggle first item
	todos.Toggle(1)
	current := todos.Todos.Get().([]Todo)
	assert.True(t, current[0].Completed)
	assert.Equal(t, 1, todos.Completed.Get())
	assert.Equal(t, 1, todos.Remaining.Get())

	// Toggle back
	todos.Toggle(1)
	current = todos.Todos.Get().([]Todo)
	assert.False(t, current[0].Completed)
	assert.Equal(t, 0, todos.Completed.Get())
	assert.Equal(t, 2, todos.Remaining.Get())
}

// TestUseTodos_Remove tests removing todos
func TestUseTodos_Remove(t *testing.T) {
	ctx := createTestContext()
	initial := []Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: true},
		{ID: 3, Title: "Task 3", Completed: false},
	}
	todos := UseTodos(ctx, initial)

	// Remove middle item
	todos.Remove(2)
	current := todos.Todos.Get().([]Todo)
	assert.Equal(t, 2, len(current))
	assert.Equal(t, int64(1), current[0].ID)
	assert.Equal(t, int64(3), current[1].ID)

	// Remove non-existent
	todos.Remove(999)
	current = todos.Todos.Get().([]Todo)
	assert.Equal(t, 2, len(current)) // No change
}

// TestUseTodos_Clear tests clearing all todos
func TestUseTodos_Clear(t *testing.T) {
	ctx := createTestContext()
	initial := []Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: true},
	}
	todos := UseTodos(ctx, initial)

	todos.Clear()
	assert.Equal(t, 0, todos.Total.Get())
	assert.Equal(t, 0, todos.Completed.Get())
	assert.Equal(t, 0, todos.Remaining.Get())
}

// TestUseTodos_ToggleAll tests toggling all todos
func TestUseTodos_ToggleAll(t *testing.T) {
	tests := []struct {
		name     string
		initial  []Todo
		expected bool // Expected completion state after toggle
	}{
		{
			name: "all incomplete to complete",
			initial: []Todo{
				{ID: 1, Title: "Task 1", Completed: false},
				{ID: 2, Title: "Task 2", Completed: false},
			},
			expected: true,
		},
		{
			name: "all complete to incomplete",
			initial: []Todo{
				{ID: 1, Title: "Task 1", Completed: true},
				{ID: 2, Title: "Task 2", Completed: true},
			},
			expected: false,
		},
		{
			name: "mixed to all complete",
			initial: []Todo{
				{ID: 1, Title: "Task 1", Completed: true},
				{ID: 2, Title: "Task 2", Completed: false},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			todos := UseTodos(ctx, tt.initial)

			todos.ToggleAll()

			current := todos.Todos.Get().([]Todo)
			for _, todo := range current {
				assert.Equal(t, tt.expected, todo.Completed)
			}
		})
	}
}

// TestUseTodos_ComputedValues tests computed properties
func TestUseTodos_ComputedValues(t *testing.T) {
	ctx := createTestContext()
	initial := []Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: true},
		{ID: 3, Title: "Task 3", Completed: true},
		{ID: 4, Title: "Task 4", Completed: false},
	}
	todos := UseTodos(ctx, initial)

	assert.Equal(t, 4, todos.Total.Get())
	assert.Equal(t, 2, todos.Completed.Get())
	assert.Equal(t, 2, todos.Remaining.Get())
	assert.Equal(t, false, todos.AllDone.Get())

	// Complete remaining
	todos.Toggle(1)
	todos.Toggle(4)

	assert.Equal(t, 4, todos.Total.Get())
	assert.Equal(t, 4, todos.Completed.Get())
	assert.Equal(t, 0, todos.Remaining.Get())
	assert.Equal(t, true, todos.AllDone.Get())
}

// TestUseTodos_SequentialOperations tests complex workflows
func TestUseTodos_SequentialOperations(t *testing.T) {
	ctx := createTestContext()
	todos := UseTodos(ctx, nil)

	// Add multiple
	todos.Add("Task 1")
	todos.Add("Task 2")
	todos.Add("Task 3")
	assert.Equal(t, 3, todos.Total.Get())

	// Complete some
	current := todos.Todos.Get().([]Todo)
	todos.Toggle(current[0].ID)
	todos.Toggle(current[1].ID)
	assert.Equal(t, 2, todos.Completed.Get())

	// Remove one
	todos.Remove(current[1].ID)
	assert.Equal(t, 2, todos.Total.Get())
	assert.Equal(t, 1, todos.Completed.Get())

	// Clear all
	todos.Clear()
	assert.Equal(t, 0, todos.Total.Get())
}

// ============================================================================
// APPROACH 2: Component-Wrapper Testing (using testutil patterns)
// ============================================================================

// createTodosComponent wraps UseTodos in a component for testutil-based testing
func createTodosComponent(initial []Todo) (bubbly.Component, error) {
	return bubbly.NewComponent("TodosTest").
		Setup(func(ctx *bubbly.Context) {
			todos := UseTodos(ctx, initial)

			// Expose all fields for testing
			ctx.Expose("todos", todos.Todos)
			ctx.Expose("total", todos.Total)
			ctx.Expose("completed", todos.Completed)
			ctx.Expose("remaining", todos.Remaining)
			ctx.Expose("allDone", todos.AllDone)
			ctx.Expose("add", todos.Add)
			ctx.Expose("toggle", todos.Toggle)
			ctx.Expose("remove", todos.Remove)
			ctx.Expose("clear", todos.Clear)
			ctx.Expose("toggleAll", todos.ToggleAll)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
}

// TestUseTodos_WithTestutil_BasicMounting demonstrates testutil-based testing
func TestUseTodos_WithTestutil_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	initial := []Todo{
		{ID: 1, Title: "Task 1", Completed: false},
	}
	component, err := createTodosComponent(initial)
	require.NoError(t, err)

	// Act: Mount component
	ct := harness.Mount(component)

	// Assert: Verify todos ref was exposed
	todosRef := ct.State().GetRef("todos")
	current := todosRef.Get().([]Todo)
	assert.Len(t, current, 1)
	assert.Equal(t, "Task 1", current[0].Title)
}

// TestUseTodos_WithTestutil_StateInspection demonstrates state inspection
func TestUseTodos_WithTestutil_StateInspection(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	component, err := createTodosComponent(nil)
	require.NoError(t, err)

	ct := harness.Mount(component)

	// Assert initial state - todos should be empty
	todosRef := ct.State().GetRef("todos")
	current := todosRef.Get().([]Todo)
	assert.Len(t, current, 0)

	// Modify todos ref directly
	newTodos := []Todo{
		{ID: time.Now().UnixNano(), Title: "Test", Completed: false},
	}
	todosRef.Set(newTodos)

	// Verify state updated
	updated := todosRef.Get().([]Todo)
	assert.Len(t, updated, 1)
	assert.Equal(t, "Test", updated[0].Title)
}

// TestUseTodos_WithTestutil_Operations demonstrates state mutations via testutil
func TestUseTodos_WithTestutil_Operations(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	initial := []Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: false},
	}
	component, err := createTodosComponent(initial)
	require.NoError(t, err)

	ct := harness.Mount(component)

	// Get initial state
	todosRef := ct.State().GetRef("todos")
	initialTodos := todosRef.Get().([]Todo)
	assert.Len(t, initialTodos, 2)

	// Act: Modify state by adding a new todo directly
	newTodo := Todo{
		ID:        time.Now().UnixNano(),
		Title:     "Task 3",
		Completed: false,
	}
	updatedTodos := append(initialTodos, newTodo)
	todosRef.Set(updatedTodos)

	// Assert: Verify state changed
	current := todosRef.Get().([]Todo)
	assert.Len(t, current, 3)
	assert.Equal(t, "Task 3", current[2].Title)
}
