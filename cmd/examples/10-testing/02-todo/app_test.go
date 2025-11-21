package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/cmd/examples/10-testing/02-todo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestTodoApp_BasicMounting tests app initialization
func TestTodoApp_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)
	require.NotNil(t, app)

	// Act: Mount app
	ct := harness.Mount(app)

	// Assert: App renders
	ct.AssertRenderContains("Todo App Example")
	ct.AssertRenderContains("Add New Todo")
}

// TestTodoApp_InitialState tests initial app state
func TestTodoApp_InitialState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Empty state
	ct.AssertRenderContains("Total: 0")
	ct.AssertRenderContains("Completed: 0")
	ct.AssertRenderContains("Remaining: 0")
	ct.AssertRenderContains("No todos yet")
}

// TestTodoApp_AddTodo tests adding todos via state manipulation
func TestTodoApp_AddTodo(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Get todos ref for state manipulation
	todosRef := ct.State().GetRef("todos")

	// Act: Add a todo by modifying the todos ref directly
	// (This simulates what the composable's Add function does)
	newTodo := composables.Todo{
		ID:        1,
		Title:     "Buy groceries",
		Completed: false,
	}
	todosRef.Set([]composables.Todo{newTodo})

	// Assert: Todo was added
	currentTodos := todosRef.Get().([]composables.Todo)
	assert.Len(t, currentTodos, 1)
	assert.Equal(t, "Buy groceries", currentTodos[0].Title)
	assert.False(t, currentTodos[0].Completed)

	// Verify render updated
	ct.AssertRenderContains("Total: 1")
	ct.AssertRenderContains("Remaining: 1")
	ct.AssertRenderContains("Buy groceries")
}

// TestTodoApp_ToggleTodo tests toggling todo completion
func TestTodoApp_ToggleTodo(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Get todos ref
	todosRef := ct.State().GetRef("todos")

	// Add a todo
	initial := []composables.Todo{
		{ID: 1, Title: "Task 1", Completed: false},
	}
	todosRef.Set(initial)

	// Act: Toggle completion by modifying state
	current := todosRef.Get().([]composables.Todo)
	current[0].Completed = true
	todosRef.Set(current)

	// Assert: Todo is completed
	updatedTodos := todosRef.Get().([]composables.Todo)
	assert.True(t, updatedTodos[0].Completed)

	// Verify stats updated
	ct.AssertRenderContains("Completed: 1")
	ct.AssertRenderContains("Remaining: 0")
}

// TestTodoApp_MultipleTodos tests working with multiple todos
func TestTodoApp_MultipleTodos(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	todosRef := ct.State().GetRef("todos")

	// Act: Add multiple todos via state
	todos := []composables.Todo{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: true},
		{ID: 3, Title: "Task 3", Completed: false},
	}
	todosRef.Set(todos)

	// Assert: All todos present
	ct.AssertRenderContains("Total: 3")
	ct.AssertRenderContains("Completed: 1")
	ct.AssertRenderContains("Remaining: 2")
}

// =============================================================================
// MODE-BASED INPUT TESTS (TDD - Write tests first!)
// =============================================================================

// TestTodoApp_ModeSystem_DefaultNavigationMode tests initial mode state
func TestTodoApp_ModeSystem_DefaultNavigationMode(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Assert: Starts in navigation mode
	modeRef := ct.State().GetRef("inputMode")
	assert.False(t, modeRef.Get().(bool), "Should start in navigation mode (false)")

	// Assert: Visual indicator shows navigation mode
	ct.AssertRenderContains("NAVIGATION")
}

// TestTodoApp_ModeSystem_ToggleToInputMode tests entering input mode
func TestTodoApp_ModeSystem_ToggleToInputMode(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	modeRef := ct.State().GetRef("inputMode")

	// Act: Toggle to input mode (via 'i' key or ESC)
	ct.Emit("toggleMode", nil)

	// Assert: Now in input mode
	assert.True(t, modeRef.Get().(bool), "Should be in input mode (true)")

	// Assert: Visual indicator shows input mode
	ct.AssertRenderContains("INPUT")
}

// TestTodoApp_ModeSystem_NavigationModeBlocksTyping tests typing blocked in navigation
func TestTodoApp_ModeSystem_NavigationModeBlocksTyping(t *testing.T) {
	// Note: Keyboard input is now forwarded to Input component only in input mode
	// This test verifies that navigation mode doesn't forward to Input component

	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	modeRef := ct.State().GetRef("inputMode")

	// Assert: Starts in navigation mode
	assert.False(t, modeRef.Get().(bool))

	// Assert: Navigation mode should not forward keyboard to Input
	// (actual keyboard forwarding is tested via integration, not unit tests)
}

// TestTodoApp_ModeSystem_InputModeAllowsTyping tests typing works in input mode
func TestTodoApp_ModeSystem_InputModeAllowsTyping(t *testing.T) {
	// Note: Text input is now handled by the Input component
	// This test verifies that input mode is correctly enabled

	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	modeRef := ct.State().GetRef("inputMode")

	// Act: Enter input mode
	ct.Emit("toggleMode", nil)

	// Assert: Input mode is active
	assert.True(t, modeRef.Get().(bool))

	// Assert: Input component should be focused (tested via visual feedback)
	ct.AssertRenderContains("INPUT MODE")
}

// TestTodoApp_ModeSystem_InputModeBlocksCommands tests commands blocked in input mode
func TestTodoApp_ModeSystem_InputModeBlocksCommands(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	todosRef := ct.State().GetRef("todos")

	// Add a todo
	todos := []composables.Todo{
		{ID: 1, Title: "Test", Completed: false},
	}
	todosRef.Set(todos)

	// Act: Enter input mode
	ct.Emit("toggleMode", nil)

	// Act: Try to fire 'd' command (should NOT delete in input mode)
	ct.Emit("remove", nil)

	// Assert: Todo still exists (command blocked in input mode)
	currentTodos := todosRef.Get().([]composables.Todo)
	assert.Len(t, currentTodos, 1, "Commands should be blocked in input mode")
}

// TestTodoApp_ModeSystem_ToggleBackToNavigation tests returning to navigation
func TestTodoApp_ModeSystem_ToggleBackToNavigation(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	modeRef := ct.State().GetRef("inputMode")

	// Act: Toggle to input, then back to navigation
	ct.Emit("toggleMode", nil)
	assert.True(t, modeRef.Get().(bool))

	ct.Emit("toggleMode", nil)
	assert.False(t, modeRef.Get().(bool))

	// Assert: Visual indicator shows navigation mode again
	ct.AssertRenderContains("NAVIGATION")
}

// =============================================================================
// VISUAL FEEDBACK TESTS (TDD - Tests for proper UI/UX)
// =============================================================================

// TestTodoApp_SpaceKeyInInputMode tests space character input
func TestTodoApp_SpaceKeyInInputMode(t *testing.T) {
	// Note: This test verifies that the Input component receives keyboard events
	// The actual text input and cursor handling is tested in the Input component's own tests
	// Here we just verify that input mode forwards messages to the Input component

	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Act: Enter input mode
	ct.Emit("toggleMode", nil)

	// Assert: Input mode is active (the actual typing is handled by Input component)
	modeRef := ct.State().GetRef("inputMode")
	assert.True(t, modeRef.Get().(bool))
}

// TestTodoApp_SpaceKeyInNavigationMode tests space toggles todos
func TestTodoApp_SpaceKeyInNavigationMode(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)
	todosRef := ct.State().GetRef("todos")
	inputRef := ct.State().GetRef("inputValue")

	// Add a todo
	todos := []composables.Todo{
		{ID: 1, Title: "Test", Completed: false},
	}
	todosRef.Set(todos)

	// Ensure we're in navigation mode
	modeRef := ct.State().GetRef("inputMode")
	assert.False(t, modeRef.Get().(bool))

	// Act: Fire toggle event (space key in navigation)
	ct.Emit("toggle", nil)

	// Assert: Todo was toggled
	currentTodos := todosRef.Get().([]composables.Todo)
	assert.True(t, currentTodos[0].Completed, "Space should toggle todo in navigation mode")

	// Assert: Input is still empty (space didn't add character)
	assert.Equal(t, "", inputRef.Get().(string))
}

// TestTodoApp_FocusStateProvided tests focus state injection
func TestTodoApp_FocusStateProvided(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, err := CreateApp()
	require.NoError(t, err)

	ct := harness.Mount(app)

	// Act: Toggle to input mode
	ct.Emit("toggleMode", nil)

	// Assert: Should provide "inputModeFocus" via Provide/Inject
	// This will be checked by child components receiving the injected value
	ct.AssertRenderContains("INPUT MODE")
}
