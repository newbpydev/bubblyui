package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// ToggleReturn is the return value of UseToggle.
// It provides reactive boolean state management with convenient methods
// for toggling, setting, and controlling the value.
type ToggleReturn struct {
	// Value is the current boolean value.
	Value *bubbly.Ref[bool]
}

// Toggle flips the value.
// If the current value is true, it becomes false.
// If the current value is false, it becomes true.
//
// Example:
//
//	toggle := composables.UseToggle(ctx, false)
//	toggle.Toggle() // Now true
//	toggle.Toggle() // Now false
func (t *ToggleReturn) Toggle() {
	current := t.Value.GetTyped()
	t.Value.Set(!current)
}

// Set sets the value explicitly.
// This allows setting the toggle to a specific value regardless of its current state.
//
// Example:
//
//	toggle := composables.UseToggle(ctx, false)
//	toggle.Set(true)  // Now true
//	toggle.Set(false) // Now false
func (t *ToggleReturn) Set(val bool) {
	t.Value.Set(val)
}

// On sets value to true.
// This is a convenience method equivalent to Set(true).
//
// Example:
//
//	toggle := composables.UseToggle(ctx, false)
//	toggle.On() // Now true (regardless of previous value)
func (t *ToggleReturn) On() {
	t.Value.Set(true)
}

// Off sets value to false.
// This is a convenience method equivalent to Set(false).
//
// Example:
//
//	toggle := composables.UseToggle(ctx, true)
//	toggle.Off() // Now false (regardless of previous value)
func (t *ToggleReturn) Off() {
	t.Value.Set(false)
}

// UseToggle creates a boolean toggle composable.
// It provides a simple API for managing boolean state with methods
// to toggle, set, turn on, and turn off the value.
//
// This composable is useful for:
//   - Dark mode toggles
//   - Sidebar visibility
//   - Feature flags
//   - Modal open/close state
//   - Any boolean state management
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initial boolean value
//
// Returns:
//   - *ToggleReturn: A struct containing the reactive boolean value and methods
//
// Example:
//
//	Setup(func(ctx *bubbly.Context) {
//	    // Create a toggle for dark mode
//	    darkMode := composables.UseToggle(ctx, false)
//	    ctx.Expose("darkMode", darkMode)
//
//	    // Toggle on key press
//	    ctx.On("toggleDarkMode", func(_ interface{}) {
//	        darkMode.Toggle()
//	    })
//
//	    // Explicit control
//	    ctx.On("enableDarkMode", func(_ interface{}) {
//	        darkMode.On()
//	    })
//
//	    ctx.On("disableDarkMode", func(_ interface{}) {
//	        darkMode.Off()
//	    })
//	}).
//	WithKeyBinding("d", "toggleDarkMode", "Toggle dark mode")
//
// Template usage:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    darkMode := ctx.Get("darkMode").(*composables.ToggleReturn)
//
//	    if darkMode.Value.GetTyped() {
//	        return renderDarkTheme(ctx)
//	    }
//	    return renderLightTheme(ctx)
//	})
//
// Integration with CreateShared:
//
//	var UseSharedDarkMode = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.ToggleReturn {
//	        return composables.UseToggle(ctx, false)
//	    },
//	)
func UseToggle(ctx *bubbly.Context, initial bool) *ToggleReturn {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseToggle", time.Since(start))
	}()

	// Create reactive ref with initial value
	value := bubbly.NewRef(initial)

	return &ToggleReturn{
		Value: value,
	}
}
