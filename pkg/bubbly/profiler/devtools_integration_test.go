package profiler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestCompositeHook_DevToolsAndProfilerCoexist verifies that both
// DevTools and Profiler can receive events through the composite hook.
func TestCompositeHook_DevToolsAndProfilerCoexist(t *testing.T) {
	devtoolsCalls := 0

	devtoolsHook := &mockHook{
		onRenderComplete: func(id string, d time.Duration) {
			devtoolsCalls++
		},
		onComponentMount: func(id, name string) {
			devtoolsCalls++
		},
	}

	prof := New(WithEnabled(true))
	profilerHook := NewHookAdapter(prof)
	prof.SetHookAdapter(profilerHook)

	// Create composite
	composite := NewCompositeHook(devtoolsHook, profilerHook)

	// Trigger events
	composite.OnComponentMount("comp-1", "TestComponent")
	composite.OnRenderComplete("comp-1", 10*time.Millisecond)
	composite.OnRenderComplete("comp-1", 15*time.Millisecond)

	// DevTools should receive all 3 calls (1 mount + 2 renders)
	assert.Equal(t, 3, devtoolsCalls)

	// Profiler should have recorded 2 renders
	tracker := profilerHook.GetComponentTracker()
	metrics := tracker.GetMetrics("comp-1")
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(2), metrics.RenderCount)
}

// TestCompositeHook_WorksWithActualBubblyFrameworkHook verifies
// compatibility with the real bubbly.FrameworkHook interface.
func TestCompositeHook_WorksWithActualBubblyFrameworkHook(t *testing.T) {
	prof := New(WithEnabled(true))
	profilerHook := NewHookAdapter(prof)
	prof.SetHookAdapter(profilerHook)

	composite := NewCompositeHook(profilerHook)

	// Verify it implements bubbly.FrameworkHook
	var _ bubbly.FrameworkHook = composite

	// Register it (this is what main.go does)
	err := bubbly.RegisterHook(composite)
	assert.NoError(t, err)

	// Trigger events
	composite.OnComponentMount("comp-1", "TestComponent")
	composite.OnRenderComplete("comp-1", 20*time.Millisecond)

	// Profiler should have recorded the data
	report := prof.GenerateReport()
	assert.Len(t, report.Components, 1)
	assert.Equal(t, "TestComponent", report.Components[0].ComponentName)

	// Cleanup
	_ = bubbly.UnregisterHook()
}
