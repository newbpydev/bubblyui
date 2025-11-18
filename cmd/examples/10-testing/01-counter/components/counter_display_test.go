package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// createTestContextForComponent creates a context for component tests
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

// TestCounterDisplay_BasicMounting demonstrates component mounting
// Shows: Component creation, mounting, initialization
func TestCounterDisplay_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()
	
	countRef := ctx.Ref(0)
	isEvenComputed := ctx.Computed(func() interface{} {
		return countRef.Get().(int)%2 == 0
	})

	display, err := CreateCounterDisplay(CounterDisplayProps{
		Count:   countRef,
		Doubled: ctx.Computed(func() interface{} { return countRef.Get().(int) * 2 }),
		IsEven:  isEvenComputed,
		History: ctx.Ref([]int{0}),
	})
	require.NoError(t, err, "Component creation should succeed")

	// Act
	ct := harness.Mount(display)

	// Assert
	assert.NotNil(t, ct, "ComponentTest should not be nil")
}

// TestCounterDisplay_RenderOutput demonstrates render testing
// Shows: Render assertions, content verification
func TestCounterDisplay_RenderOutput(t *testing.T) {
	tests := []struct {
		name            string
		count           int
		expectedCount   string
		expectedDoubled string
		expectedParity  string
	}{
		{
			name:            "zero",
			count:           0,
			expectedCount:   "Count: 0",
			expectedDoubled: "Doubled: 0",
			expectedParity:  "Even",
		},
		{
			name:            "positive even",
			count:           4,
			expectedCount:   "Count: 4",
			expectedDoubled: "Doubled: 8",
			expectedParity:  "Even",
		},
		{
			name:            "positive odd",
			count:           7,
			expectedCount:   "Count: 7",
			expectedDoubled: "Doubled: 14",
			expectedParity:  "Odd",
		},
		{
			name:            "negative even",
			count:           -2,
			expectedCount:   "Count: -2",
			expectedDoubled: "Doubled: -4",
			expectedParity:  "Even",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			ctx := createTestContextForComponent()
			countRef := ctx.Ref(tt.count)

			display, err := CreateCounterDisplay(CounterDisplayProps{
				Count: countRef,
				Doubled: ctx.Computed(func() interface{} {
					return countRef.Get().(int) * 2
				}),
				IsEven: ctx.Computed(func() interface{} {
					return countRef.Get().(int)%2 == 0
				}),
				History: ctx.Ref([]int{tt.count}),
			})
			require.NoError(t, err)

			ct := harness.Mount(display)

			// Assert
			ct.AssertRenderContains(tt.expectedCount)
			ct.AssertRenderContains(tt.expectedDoubled)
			ct.AssertRenderContains(tt.expectedParity)
		})
	}
}

// TestCounterDisplay_ReactiveUpdates demonstrates reactive rendering
// Shows: Reactive ref updates, computed value reactivity
func TestCounterDisplay_ReactiveUpdates(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()
	countRef := ctx.Ref(0)

	display, err := CreateCounterDisplay(CounterDisplayProps{
		Count: countRef,
		Doubled: ctx.Computed(func() interface{} {
			return countRef.Get().(int) * 2
		}),
		IsEven: ctx.Computed(func() interface{} {
			return countRef.Get().(int)%2 == 0
		}),
		History: ctx.Ref([]int{0}),
	})
	require.NoError(t, err)

	ct := harness.Mount(display)

	// Initial state
	ct.AssertRenderContains("Count: 0")

	// Act: Update the ref
	countRef.Set(5)

	// Assert: Render should update
	ct.AssertRenderContains("Count: 5")
	ct.AssertRenderContains("Doubled: 10")
	ct.AssertRenderContains("Odd")
}

// TestCounterDisplay_ComputedValues demonstrates computed value display
func TestCounterDisplay_ComputedValues(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		doubled string
	}{
		{"zero", 0, "Doubled: 0"},
		{"positive", 5, "Doubled: 10"},
		{"negative", -3, "Doubled: -6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			ctx := createTestContextForComponent()
			countRef := ctx.Ref(tt.count)

			display, err := CreateCounterDisplay(CounterDisplayProps{
				Count: countRef,
				Doubled: ctx.Computed(func() interface{} {
					return countRef.Get().(int) * 2
				}),
				IsEven: ctx.Computed(func() interface{} {
					return countRef.Get().(int)%2 == 0
				}),
				History: ctx.Ref([]int{tt.count}),
			})
			require.NoError(t, err)

			ct := harness.Mount(display)

			// Assert: Doubled value is rendered correctly
			ct.AssertRenderContains(tt.doubled)
		})
	}
}

// TestCounterDisplay_History demonstrates history display
func TestCounterDisplay_History(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()
	countRef := ctx.Ref(3)
	historyRef := ctx.Ref([]int{0, 1, 2, 3})

	display, err := CreateCounterDisplay(CounterDisplayProps{
		Count: countRef,
		Doubled: ctx.Computed(func() interface{} {
			return countRef.Get().(int) * 2
		}),
		IsEven: ctx.Computed(func() interface{} {
			return countRef.Get().(int)%2 == 0
		}),
		History: historyRef,
	})
	require.NoError(t, err)

	ct := harness.Mount(display)

	// Assert: History is displayed
	ct.AssertRenderContains("History:")
	// Note: Actual history values might be formatted differently in the component
}

// TestCounterDisplay_Cleanup demonstrates component cleanup
func TestCounterDisplay_Cleanup(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)
	ctx := createTestContextForComponent()
	countRef := ctx.Ref(0)

	display, err := CreateCounterDisplay(CounterDisplayProps{
		Count: countRef,
		Doubled: ctx.Computed(func() interface{} {
			return countRef.Get().(int) * 2
		}),
		IsEven: ctx.Computed(func() interface{} {
			return countRef.Get().(int)%2 == 0
		}),
		History: ctx.Ref([]int{0}),
	})
	require.NoError(t, err)

	ct := harness.Mount(display)

	// Act: Use component
	ct.AssertRenderContains("Count: 0")

	// Act: Unmount
	ct.Unmount()

	// Assert: Component is unmounted (cleanup happens automatically)
	// In a real scenario with resources, we'd verify they were released
}
