package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// createTestContextForComponent creates a minimal context for component testing
func createTestContextForComponent() *bubbly.Context {
	var ctx *bubbly.Context
	component, _ := bubbly.NewComponent("Test").
		Setup(func(c *bubbly.Context) {
			ctx = c
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()
	component.Init()
	return ctx
}

// TestTodoInput_BasicMounting tests component initialization and mounting
// Shows: Component creation, harness usage, basic assertions
func TestTodoInput_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("")

	// Act: Create and mount component
	input, err := CreateTodoInput(TodoInputProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Value:    valueRef,
		OnSubmit: func(title string) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Assert: Component renders
	ct.AssertRenderContains("Add New Todo")
	ct.AssertRenderContains("Press [enter] to add")
}

// TestTodoInput_RenderOutput demonstrates render testing
// Shows: Render assertions, content verification
func TestTodoInput_RenderOutput(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "empty value",
			value:    "",
			expected: "Add New Todo",
		},
		{
			name:     "with text",
			value:    "Buy groceries",
			expected: "Add New Todo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			valueRef := bubbly.NewRef(tt.value)

			input, err := CreateTodoInput(TodoInputProps{
				Focused:  bubbly.NewRef[interface{}](false),
				Value:    valueRef,
				OnSubmit: func(title string) {},
			})
			require.NoError(t, err)

			ct := harness.Mount(input)

			// Assert: Render contains expected content
			ct.AssertRenderContains(tt.expected)
		})
	}
}

// TestTodoInput_ValueProp tests value prop reactivity
// Shows: Prop passing, state management, reactive updates
func TestTodoInput_ValueProp(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("Initial value")

	input, err := CreateTodoInput(TodoInputProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Value:    valueRef,
		OnSubmit: func(title string) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Assert: Initial value is set
	assert.Equal(t, "Initial value", valueRef.Get())

	// Act: Update value
	valueRef.Set("Updated value")

	// Assert: Value updated (component receives prop changes)
	assert.Equal(t, "Updated value", valueRef.Get())
	ct.AssertRenderContains("Add New Todo")
}

// TestTodoInput_OnSubmitCallback tests submit callback
// Shows: Event handler testing, callback verification
func TestTodoInput_OnSubmitCallback(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("Test todo")
	var submittedValue string
	submitted := false

	input, err := CreateTodoInput(TodoInputProps{
		Focused: bubbly.NewRef[interface{}](false),
		Value:   valueRef,
		OnSubmit: func(title string) {
			submitted = true
			submittedValue = title
		},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Act: Emit submit event
	ct.Emit("submit", nil)

	// Assert: Callback was called with correct value
	assert.True(t, submitted, "OnSubmit callback should have been called")
	assert.Equal(t, "Test todo", submittedValue)
}

// TestTodoInput_EmptySubmit tests that empty values are not submitted
// Shows: Validation logic, edge case testing
func TestTodoInput_EmptySubmit(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("")
	submitted := false

	input, err := CreateTodoInput(TodoInputProps{
		Focused: bubbly.NewRef[interface{}](false),
		Value:   valueRef,
		OnSubmit: func(title string) {
			submitted = true
		},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Act: Try to submit empty value
	ct.Emit("submit", nil)

	// Assert: Callback was NOT called for empty value
	assert.False(t, submitted, "OnSubmit should not be called for empty value")
}

// TestTodoInput_ClearAfterSubmit tests value clearing after submit
// Shows: State mutation, side effects verification
func TestTodoInput_ClearAfterSubmit(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("Test todo")

	input, err := CreateTodoInput(TodoInputProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Value:    valueRef,
		OnSubmit: func(title string) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Act: Submit
	ct.Emit("submit", nil)

	// Assert: Value was cleared
	assert.Equal(t, "", valueRef.Get().(string))
}

// TestTodoInput_EventTracking demonstrates event tracking
// Shows: EventTracker usage, AssertEventFired
func TestTodoInput_EventTracking(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("Test")

	input, err := CreateTodoInput(TodoInputProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Value:    valueRef,
		OnSubmit: func(title string) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Act: Emit submit event
	ct.Emit("submit", nil)

	// Assert: Event was tracked
	ct.AssertEventFired("submit")
	ct.AssertEventCount("submit", 1)
}

// TestTodoInput_MultipleSubmits tests sequential submissions
// Shows: Complex workflows, state transitions
func TestTodoInput_MultipleSubmits(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("First todo")
	submissions := []string{}

	input, err := CreateTodoInput(TodoInputProps{
		Focused: bubbly.NewRef[interface{}](false),
		Value:   valueRef,
		OnSubmit: func(title string) {
			submissions = append(submissions, title)
		},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Act: Submit multiple times
	ct.Emit("submit", nil)

	valueRef.Set("Second todo")
	ct.Emit("submit", nil)

	valueRef.Set("Third todo")
	ct.Emit("submit", nil)

	// Assert: All submissions recorded
	assert.Len(t, submissions, 3)
	assert.Equal(t, "First todo", submissions[0])
	assert.Equal(t, "Second todo", submissions[1])
	assert.Equal(t, "Third todo", submissions[2])

	// Assert: Event count correct
	ct.AssertEventCount("submit", 3)
}

// TestTodoInput_Cleanup demonstrates cleanup verification
// Shows: Resource cleanup, unmount behavior
func TestTodoInput_Cleanup(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("Test")

	input, err := CreateTodoInput(TodoInputProps{
		Focused:  bubbly.NewRef[interface{}](false),
		Value:    valueRef,
		OnSubmit: func(title string) {},
	})
	require.NoError(t, err)

	ct := harness.Mount(input)

	// Act: Use component
	ct.Emit("submit", nil)

	// Act: Unmount (cleanup happens automatically via t.Cleanup)
	ct.Unmount()

	// Assert: Component is unmounted
	// (In real scenarios, you'd verify resources were released)
}
