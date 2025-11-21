package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestApp_BasicMounting demonstrates the most basic test pattern
// Shows: TestHarness creation, component mounting, render verification
func TestApp_BasicMounting(t *testing.T) {
	// Arrange: Create test harness
	harness := testutil.NewHarness(t)

	// Create and mount app
	app, err := CreateApp()
	require.NoError(t, err, "App creation should succeed")

	ct := harness.Mount(app)

	// Assert: App mounted successfully
	assert.NotNil(t, ct, "ComponentTest should not be nil")

	// Assert: Initial render shows count of 0
	ct.AssertRenderContains("Count: 0")
	ct.AssertRenderContains("Doubled: 0")
}

// TestApp_Increment demonstrates increment functionality
// Shows: Event emission, render output verification
func TestApp_Increment(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, _ := CreateApp()
	ct := harness.Mount(app)

	// Act: Emit increment event
	ct.Emit("increment", nil)

	// Assert: Render updated
	ct.AssertRenderContains("Count: 1")
	ct.AssertRenderContains("Doubled: 2")
}

// TestApp_TableDriven demonstrates table-driven test pattern
// Shows: Multiple test cases, structured testing, comprehensive coverage
func TestApp_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		events         []string
		expectedCount  string
		expectedDouble string
	}{
		{
			name:           "single increment",
			events:         []string{"increment"},
			expectedCount:  "Count: 1",
			expectedDouble: "Doubled: 2",
		},
		{
			name:           "multiple increments",
			events:         []string{"increment", "increment", "increment"},
			expectedCount:  "Count: 3",
			expectedDouble: "Doubled: 6",
		},
		{
			name:           "increment then decrement",
			events:         []string{"increment", "increment", "decrement"},
			expectedCount:  "Count: 1",
			expectedDouble: "Doubled: 2",
		},
		{
			name:           "reset after increments",
			events:         []string{"increment", "increment", "reset"},
			expectedCount:  "Count: 0",
			expectedDouble: "Doubled: 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			app, err := CreateApp()
			require.NoError(t, err)

			ct := harness.Mount(app)

			// Act: Emit events
			for _, event := range tt.events {
				ct.Emit(event, nil)
			}

			// Assert: Render shows expected values
			ct.AssertRenderContains(tt.expectedCount)
			ct.AssertRenderContains(tt.expectedDouble)
		})
	}
}

// TestApp_EventTracking demonstrates event tracking
// Shows: EventTracker, AssertEventFired, AssertEventCount
func TestApp_EventTracking(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, _ := CreateApp()
	ct := harness.Mount(app)

	// Act: Emit multiple events
	ct.Emit("increment", nil)
	ct.Emit("increment", nil)
	ct.Emit("decrement", nil)

	// Assert: Events were tracked
	ct.AssertEventFired("increment")
	ct.AssertEventCount("increment", 2)
	ct.AssertEventCount("decrement", 1)
	ct.AssertEventNotFired("reset")
}

// TestApp_RenderOutput demonstrates render assertions
// Shows: AssertRenderContains, comprehensive output verification
func TestApp_RenderOutput(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, _ := CreateApp()
	ct := harness.Mount(app)

	// Assert: Initial render contains all expected elements
	ct.AssertRenderContains("Counter App")
	ct.AssertRenderContains("Count: 0")
	ct.AssertRenderContains("Doubled: 0")
	ct.AssertRenderContains("Parity:")
	ct.AssertRenderContains("History:")

	// Act: Change state
	ct.Emit("increment", nil)

	// Assert: Render updated with new values
	ct.AssertRenderContains("Count: 1")
	ct.AssertRenderContains("Doubled: 2")
}

// TestApp_MultipleOperations demonstrates complex scenarios
// Shows: Multiple state changes, sequential operations
func TestApp_MultipleOperations(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, _ := CreateApp()
	ct := harness.Mount(app)

	// Act: Perform multiple operations
	ct.Emit("increment", nil) // 0 -> 1
	ct.Emit("increment", nil) // 1 -> 2
	ct.Emit("increment", nil) // 2 -> 3
	ct.Emit("decrement", nil) // 3 -> 2
	ct.Emit("set", 10)        // 2 -> 10
	ct.Emit("reset", nil)     // 10 -> 0

	// Assert: Final state is correct
	ct.AssertRenderContains("Count: 0")

	// Assert: All events were fired
	ct.AssertEventCount("increment", 3)
	ct.AssertEventCount("decrement", 1)
	ct.AssertEventCount("set", 1)
	ct.AssertEventCount("reset", 1)
}

// TestApp_Cleanup demonstrates cleanup verification
// Shows: Resource cleanup, unmount behavior
func TestApp_Cleanup(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	app, _ := CreateApp()
	ct := harness.Mount(app)

	// Act: Use app
	ct.Emit("increment", nil)
	ct.AssertRenderContains("Count: 1")

	// Act: Unmount (cleanup happens automatically via t.Cleanup)
	ct.Unmount()

	// Assert: Component is unmounted
	// (In real scenarios, you'd verify resources were released)
}
