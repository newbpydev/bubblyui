package composables

import (
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// ModeReturn is the return value of UseMode.
// It provides reactive mode state management for TUI applications,
// enabling vim-like navigation/input mode patterns.
type ModeReturn[T comparable] struct {
	// Current is the current mode.
	Current *bubbly.Ref[T]

	// Previous is the previous mode (for transitions).
	// Initially set to the same value as Current.
	Previous *bubbly.Ref[T]
}

// IsMode returns true if currently in the specified mode.
//
// Example:
//
//	if mode.IsMode(ModeNavigation) {
//	    // Handle navigation keys
//	}
func (m *ModeReturn[T]) IsMode(mode T) bool {
	return m.Current.GetTyped() == mode
}

// Switch changes to a new mode.
// Updates Previous to the old mode before switching.
// If already in the specified mode, this is a no-op.
//
// Example:
//
//	ctx.On("enterInput", func(_ interface{}) {
//	    mode.Switch(ModeInput)
//	})
func (m *ModeReturn[T]) Switch(mode T) {
	current := m.Current.GetTyped()
	if current == mode {
		return // No-op if already in this mode
	}

	// Update previous before switching
	m.Previous.Set(current)
	m.Current.Set(mode)
}

// Toggle switches between two modes.
// If currently in mode a, switches to mode b.
// If currently in mode b, switches to mode a.
// If currently in neither, switches to mode a.
//
// Example:
//
//	ctx.On("toggleMode", func(_ interface{}) {
//	    mode.Toggle(ModeNavigation, ModeInput)
//	})
func (m *ModeReturn[T]) Toggle(a, b T) {
	current := m.Current.GetTyped()

	var newMode T
	if current == a {
		newMode = b
	} else if current == b {
		newMode = a
	} else {
		// Not in either mode, switch to first option
		newMode = a
	}

	// Use Switch to properly update Previous
	m.Switch(newMode)
}

// UseMode creates a mode management composable for TUI applications.
// It tracks the current mode and previous mode, enabling vim-like
// navigation/input mode patterns.
//
// This composable is essential for building TUI applications that need
// different key binding behaviors based on the current mode (e.g., vim-like
// navigation mode vs input mode).
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initial mode
//
// Returns:
//   - *ModeReturn[T]: A struct containing reactive mode state and methods
//
// Example:
//
//	type Mode string
//	const (
//	    ModeNavigation Mode = "navigation"
//	    ModeInput      Mode = "input"
//	)
//
//	Setup(func(ctx *bubbly.Context) {
//	    mode := composables.UseMode(ctx, ModeNavigation)
//	    ctx.Expose("mode", mode)
//
//	    ctx.On("toggleMode", func(_ interface{}) {
//	        mode.Toggle(ModeNavigation, ModeInput)
//	    })
//
//	    ctx.On("enterInput", func(_ interface{}) {
//	        mode.Switch(ModeInput)
//	    })
//
//	    ctx.On("exitInput", func(_ interface{}) {
//	        mode.Switch(ModeNavigation)
//	    })
//	}).
//	WithKeyBinding("i", "enterInput", "Enter input mode").
//	WithKeyBinding("esc", "exitInput", "Exit to navigation")
//
// Mode-dependent key handling in Template:
//
//	Template(func(ctx bubbly.RenderContext) string {
//	    mode := ctx.Get("mode").(*composables.ModeReturn[Mode])
//
//	    if mode.IsMode(ModeNavigation) {
//	        return renderNavigationMode(ctx)
//	    }
//	    return renderInputMode(ctx)
//	})
//
// Integration with CreateShared:
//
//	var UseSharedMode = composables.CreateShared(
//	    func(ctx *bubbly.Context) *composables.ModeReturn[Mode] {
//	        return composables.UseMode(ctx, ModeNavigation)
//	    },
//	)
func UseMode[T comparable](ctx *bubbly.Context, initial T) *ModeReturn[T] {
	// Record metrics if monitoring is enabled
	start := time.Now()
	defer func() {
		monitoring.GetGlobalMetrics().RecordComposableCreation("UseMode", time.Since(start))
	}()

	// Create reactive refs with initial value
	// Previous starts as initial since no switch has occurred yet
	current := bubbly.NewRef(initial)
	previous := bubbly.NewRef(initial)

	return &ModeReturn[T]{
		Current:  current,
		Previous: previous,
	}
}
