package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
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
	t.Run("state changes work", func(t *testing.T) {
		component := createAutoTestComponentWithRef("count", 0)

		tester := NewAutoCommandTester(component)
		tester.EnableAutoCommands()

		// Verify initial state
		ref := tester.state.GetRef("count")
		assert.NotNil(t, ref)
		assert.Equal(t, 0, ref.Get())

		// Trigger state change
		tester.TriggerStateChange("count", 42)

		// Verify state changed
		assert.Equal(t, 42, ref.Get(), "state should be updated")
	})

	t.Run("queue inspector accessible", func(t *testing.T) {
		component := createAutoTestComponentWithRef("count", 0)
		tester := NewAutoCommandTester(component)

		// Get queue inspector
		queue := tester.GetQueueInspector()
		assert.NotNil(t, queue, "queue inspector should not be nil")
	})

	t.Run("loop detector accessible", func(t *testing.T) {
		component := createAutoTestComponentWithRef("count", 0)
		tester := NewAutoCommandTester(component)

		// Get loop detector
		detector := tester.GetLoopDetector()
		assert.NotNil(t, detector, "loop detector should not be nil")
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
	component.Init()

	return component
}
