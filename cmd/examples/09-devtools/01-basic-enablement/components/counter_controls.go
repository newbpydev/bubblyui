package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// CounterControlsProps defines the props for CounterControls component
type CounterControlsProps struct {
	OnIncrement func()
	OnDecrement func()
	OnReset     func()
}

// CreateCounterControls creates a component with action buttons
// This demonstrates:
// - Callback props for parent communication
// - Using BubblyUI Button components
// - Event handling pattern
func CreateCounterControls(props CounterControlsProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("CounterControls")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Register event handlers that call parent callbacks
		ctx.On("increment", func(_ interface{}) {
			if props.OnIncrement != nil {
				props.OnIncrement()
			}
		})

		ctx.On("decrement", func(_ interface{}) {
			if props.OnDecrement != nil {
				props.OnDecrement()
			}
		})

		ctx.On("reset", func(_ interface{}) {
			if props.OnReset != nil {
				props.OnReset()
			}
		})

		ctx.OnMounted(func() {
			// Component mounted - visible in dev tools
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Display control hints using Text components
		// The actual key bindings are handled by the parent App component
		
		helpText := components.Text(components.TextProps{
			Content: "Controls: [i] Increment  [d] Decrement  [r] Reset  [F12] Toggle DevTools  [ctrl+c] Quit",
			Color:   lipgloss.Color("240"), // Muted color
		})
		helpText.Init()

		return helpText.View()
	})

	return builder.Build()
}
