package testutil

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitFor tests the WaitFor function with various scenarios
func TestWaitFor(t *testing.T) {
	tests := []struct {
		name          string
		condition     func() bool
		opts          WaitOptions
		shouldTimeout bool
		description   string
	}{
		{
			name: "condition becomes true immediately",
			condition: func() bool {
				return true
			},
			opts: WaitOptions{
				Timeout:  100 * time.Millisecond,
				Interval: 10 * time.Millisecond,
			},
			shouldTimeout: false,
			description:   "should return immediately when condition is true",
		},
		{
			name: "condition becomes true after delay",
			condition: func() func() bool {
				start := time.Now()
				return func() bool {
					return time.Since(start) > 50*time.Millisecond
				}
			}(),
			opts: WaitOptions{
				Timeout:  200 * time.Millisecond,
				Interval: 10 * time.Millisecond,
			},
			shouldTimeout: false,
			description:   "should wait and succeed when condition becomes true",
		},
		{
			name: "condition never becomes true - timeout",
			condition: func() bool {
				return false
			},
			opts: WaitOptions{
				Timeout:  50 * time.Millisecond,
				Interval: 10 * time.Millisecond,
				Message:  "custom timeout message",
			},
			shouldTimeout: true,
			description:   "should timeout when condition never becomes true",
		},
		{
			name: "uses default timeout",
			condition: func() bool {
				return true
			},
			opts: WaitOptions{
				// No timeout specified - should use default 5s
			},
			shouldTimeout: false,
			description:   "should use default timeout of 5 seconds",
		},
		{
			name: "uses default interval",
			condition: func() bool {
				return true
			},
			opts: WaitOptions{
				Timeout: 100 * time.Millisecond,
				// No interval specified - should use default 10ms
			},
			shouldTimeout: false,
			description:   "should use default interval of 10ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock testing.T to capture errors
			mockT := &mockTestingT{}

			// Run WaitFor
			WaitFor(mockT, tt.condition, tt.opts)

			// Check if timeout occurred as expected
			if tt.shouldTimeout {
				assert.True(t, mockT.failed, "expected timeout")
				assert.NotEmpty(t, mockT.errors, "expected error message")
			} else {
				assert.False(t, mockT.failed, "expected no timeout")
			}
		})
	}
}

// TestWaitForRef tests the WaitForRef method on ComponentTest
func TestWaitForRef(t *testing.T) {
	t.Run("ref value matches immediately", func(t *testing.T) {
		harness := NewHarness(t)

		// Create a simple component
		count := bubbly.NewRef[interface{}](42)
		harness.refs["count"] = count

		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				ctx.Expose("count", count)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()
		require.NoError(t, err)

		// Mount component
		ct := harness.Mount(component)

		// Wait for ref - should succeed immediately
		ct.WaitForRef("count", 42, 100*time.Millisecond)

		// Verify no error
		assert.Equal(t, 42, ct.state.GetRefValue("count"))
	})

	t.Run("ref value changes to match", func(t *testing.T) {
		harness := NewHarness(t)

		// Create a component with async state update
		count := bubbly.NewRef[interface{}](0)
		harness.refs["count"] = count

		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				ctx.Expose("count", count)

				// Simulate async update
				ctx.On("update", func(payload interface{}) {
					go func() {
						time.Sleep(50 * time.Millisecond)
						count.Set(42)
					}()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()
		require.NoError(t, err)

		ct := harness.Mount(component)

		// Trigger async update
		ct.component.Emit("update", nil)

		// Wait for ref to change
		ct.WaitForRef("count", 42, 200*time.Millisecond)

		// Verify value changed
		assert.Equal(t, 42, ct.state.GetRefValue("count"))
	})

	t.Run("ref value never matches - timeout", func(t *testing.T) {
		harness := NewHarness(t)

		// Create a simple component
		count := bubbly.NewRef[interface{}](0)
		harness.refs["count"] = count

		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				ctx.Expose("count", count)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()
		require.NoError(t, err)

		ct := harness.Mount(component)

		// Create mock testing.T to capture timeout
		mockT := &mockTestingT{}
		ct.harness.t = mockT

		// Wait for ref - should timeout
		ct.WaitForRef("count", 42, 50*time.Millisecond)

		// Verify timeout occurred
		assert.True(t, mockT.failed, "expected timeout")
	})
}

// TestWaitForEvent tests the WaitForEvent method on ComponentTest
func TestWaitForEvent(t *testing.T) {
	t.Run("event already fired", func(t *testing.T) {
		harness := NewHarness(t)

		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				// Component setup
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()
		require.NoError(t, err)

		ct := harness.Mount(component)

		// Fire event immediately
		ct.events.tracker.Track("click", nil, "test")

		// Wait for event - should succeed immediately
		ct.WaitForEvent("click", 100*time.Millisecond)

		// Verify event was fired
		assert.True(t, ct.events.tracker.WasFired("click"))
	})

	t.Run("event fires after delay", func(t *testing.T) {
		harness := NewHarness(t)

		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				// Component setup
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()
		require.NoError(t, err)

		ct := harness.Mount(component)

		// Simulate async event emission
		go func() {
			time.Sleep(50 * time.Millisecond)
			ct.events.tracker.Track("completed", nil, "test")
		}()

		// Wait for completion event
		ct.WaitForEvent("completed", 200*time.Millisecond)

		// Verify event was fired
		assert.True(t, ct.events.tracker.WasFired("completed"))
	})

	t.Run("event never fires - timeout", func(t *testing.T) {
		harness := NewHarness(t)

		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				// Component setup
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()
		require.NoError(t, err)

		ct := harness.Mount(component)

		// Create mock testing.T to capture timeout
		mockT := &mockTestingT{}
		ct.harness.t = mockT

		// Wait for event that never fires
		ct.WaitForEvent("never-fired", 50*time.Millisecond)

		// Verify timeout occurred
		assert.True(t, mockT.failed, "expected timeout")
	})
}

// TestWaitForRef_Integration tests WaitForRef in a realistic scenario
func TestWaitForRef_Integration(t *testing.T) {
	harness := NewHarness(t)

	// Create a component with async state update
	loading := bubbly.NewRef[interface{}](true)
	data := bubbly.NewRef[interface{}]("")
	harness.refs["loading"] = loading
	harness.refs["data"] = data

	component, err := bubbly.NewComponent("AsyncComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.Expose("loading", loading)
			ctx.Expose("data", data)

			// Simulate async data fetch
			ctx.On("fetch", func(payload interface{}) {
				go func() {
					time.Sleep(50 * time.Millisecond)
					loading.Set(false)
					data.Set("fetched data")
				}()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "async component"
		}).
		Build()
	require.NoError(t, err)

	ct := harness.Mount(component)

	// Trigger async operation
	ct.component.Emit("fetch", nil)

	// Wait for loading to become false
	ct.WaitForRef("loading", false, 200*time.Millisecond)

	// Verify data was loaded
	assert.Equal(t, "fetched data", ct.state.GetRefValue("data"))
}

// TestWaitForEvent_Integration tests WaitForEvent in a realistic scenario
func TestWaitForEvent_Integration(t *testing.T) {
	harness := NewHarness(t)

	// Create a component
	component, err := bubbly.NewComponent("EventComponent").
		Setup(func(ctx *bubbly.Context) {
			// Component setup
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "event component"
		}).
		Build()
	require.NoError(t, err)

	ct := harness.Mount(component)

	// Simulate async event emission
	go func() {
		time.Sleep(50 * time.Millisecond)
		ct.events.tracker.Track("completed", "result", "test")
	}()

	// Wait for completion event
	ct.WaitForEvent("completed", 200*time.Millisecond)

	// Verify event was fired
	ct.AssertEventFired("completed")
	ct.AssertEventPayload("completed", "result")
}
