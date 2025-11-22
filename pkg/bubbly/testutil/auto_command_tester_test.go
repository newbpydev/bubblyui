package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestNewAutoCommandTester tests creating a new auto-command tester
func TestNewAutoCommandTester(t *testing.T) {
	tests := []struct {
		name      string
		component bubbly.Component
		wantNil   bool
	}{
		{
			name:      "valid component",
			component: createAutoTestComponent(),
			wantNil:   false,
		},
		{
			name:      "nil component",
			component: nil,
			wantNil:   false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewAutoCommandTester(tt.component)

			if tt.wantNil {
				assert.Nil(t, tester)
			} else {
				assert.NotNil(t, tester)
			}
		})
	}
}

// TestAutoCommandTester_EnableAutoCommands tests enabling auto-commands
func TestAutoCommandTester_EnableAutoCommands(t *testing.T) {
	tests := []struct {
		name      string
		component bubbly.Component
		wantErr   bool
	}{
		{
			name:      "enable on valid component",
			component: createAutoTestComponent(),
			wantErr:   false,
		},
		{
			name:      "enable on nil component",
			component: nil,
			wantErr:   false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewAutoCommandTester(tt.component)
			tester.EnableAutoCommands()

			// If component is not nil, verify auto-commands are enabled
			if tt.component != nil {
				// We'll need to verify this through the component's context
				// For now, just verify the method doesn't panic
			}
		})
	}
}

// TestAutoCommandTester_TriggerStateChange tests triggering state changes
func TestAutoCommandTester_TriggerStateChange(t *testing.T) {
	tests := []struct {
		name      string
		refName   string
		value     interface{}
		wantPanic bool
	}{
		{
			name:      "trigger with valid ref",
			refName:   "count",
			value:     42,
			wantPanic: false,
		},
		{
			name:      "trigger with string value",
			refName:   "name",
			value:     "test",
			wantPanic: false,
		},
		{
			name:      "trigger with nil value",
			refName:   "data",
			value:     nil,
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := createAutoTestComponentWithRef(tt.refName, 0)
			tester := NewAutoCommandTester(component)
			tester.EnableAutoCommands()

			if tt.wantPanic {
				assert.Panics(t, func() {
					tester.TriggerStateChange(tt.refName, tt.value)
				})
			} else {
				assert.NotPanics(t, func() {
					tester.TriggerStateChange(tt.refName, tt.value)
				})
			}
		})
	}
}

// TestAutoCommandTester_Integration tests integration with queue and detector
func TestAutoCommandTester_Integration(t *testing.T) {
	t.Run("commands enqueued on state change", func(t *testing.T) {
		component := createAutoTestComponentWithRef("count", 0)
		tester := NewAutoCommandTester(component)

		// Enable auto-commands
		tester.EnableAutoCommands()

		// Get queue inspector
		queue := tester.GetQueueInspector()
		assert.NotNil(t, queue)

		// Verify queue is initially empty
		initialLen := queue.Len()

		// Trigger state change - this should generate a command
		tester.TriggerStateChange("count", 42)

		// Verify command was enqueued
		assert.Greater(t, queue.Len(), initialLen, "command should be enqueued after state change")

		// Verify state actually changed
		ref := tester.state.GetRef("count")
		assert.Equal(t, 42, ref.Get(), "state should be updated")
	})

	t.Run("loop detector accessible and functional", func(t *testing.T) {
		component := createAutoTestComponentWithRef("count", 0)
		tester := NewAutoCommandTester(component)

		// Enable auto-commands
		tester.EnableAutoCommands()

		// Get loop detector
		detector := tester.GetLoopDetector()
		assert.NotNil(t, detector)

		// Verify detector starts with no loops detected
		assert.False(t, detector.WasDetected(), "no loops should be detected initially")

		// Note: Actual loop detection happens during recursive state changes
		// within a single update cycle, not from manual sequential changes.
		// The detector is accessible and ready to track loops when they occur.
		assert.Equal(t, 0, detector.GetLoopCount(), "loop count should be zero initially")
	})

	t.Run("queue inspector tracks multiple commands", func(t *testing.T) {
		component := createAutoTestComponentWithRef("count", 0)
		tester := NewAutoCommandTester(component)

		// Enable auto-commands
		tester.EnableAutoCommands()

		// Get queue inspector
		queue := tester.GetQueueInspector()
		initialLen := queue.Len()

		// Trigger multiple state changes
		tester.TriggerStateChange("count", 1)
		tester.TriggerStateChange("count", 2)
		tester.TriggerStateChange("count", 3)

		// Verify multiple commands were enqueued
		assert.Equal(t, initialLen+3, queue.Len(), "should have 3 commands enqueued")
	})
}

// TestAutoCommandTester_NilComponent tests behavior with nil component
func TestAutoCommandTester_NilComponent(t *testing.T) {
	tester := NewAutoCommandTester(nil)
	assert.NotNil(t, tester)

	// All methods should handle nil gracefully
	assert.NotPanics(t, func() {
		tester.EnableAutoCommands()
	})

	assert.NotPanics(t, func() {
		tester.TriggerStateChange("count", 42)
	})

	// Inspectors should return nil or empty values
	queue := tester.GetQueueInspector()
	assert.NotNil(t, queue) // Should return a valid inspector even with nil component

	detector := tester.GetLoopDetector()
	assert.NotNil(t, detector) // Should return a valid detector even with nil component
}

// Helper functions

func createAutoTestComponent() bubbly.Component {
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			ctx.Ref(0)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	if err != nil {
		panic(err)
	}

	// Initialize component so refs are created
	component.Init()

	return component
}

func createAutoTestComponentWithRef(refName string, initialValue interface{}) bubbly.Component {
	component, err := bubbly.NewComponent("TestComponent").
		WithAutoCommands(true). // Enable auto-commands BEFORE setup
		Setup(func(ctx *bubbly.Context) {
			ref := ctx.Ref(initialValue)
			ctx.Expose(refName, ref)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	if err != nil {
		panic(err)
	}

	// Initialize component so refs are created
	// Refs will have setHook attached because auto-commands were enabled in builder
	component.Init()

	return component
}
