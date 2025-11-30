package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// ControlsPanelProps defines the props for the ControlsPanel component.
type ControlsPanelProps struct {
	// IsRunning indicates if the profiler is running
	IsRunning *bubbly.Ref[bool]

	// Focused indicates if this panel has focus
	Focused *bubbly.Ref[bool]

	// OnToggle is called when the profiler should be toggled
	OnToggle func()

	// OnReset is called when metrics should be reset
	OnReset func()

	// OnExport is called when report should be exported
	OnExport func()
}

// CreateControlsPanel creates a component with profiler controls.
// This demonstrates:
// - Callback props for parent communication
// - Using BubblyUI Text components
// - Dynamic styling based on focus and state
func CreateControlsPanel(props ControlsPanelProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("ControlsPanel")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("isRunning", props.IsRunning)
		ctx.Expose("focused", props.Focused)

		// Register event handlers
		ctx.On("toggle", func(_ interface{}) {
			if props.OnToggle != nil {
				props.OnToggle()
			}
		})

		ctx.On("reset", func(_ interface{}) {
			if props.OnReset != nil {
				props.OnReset()
			}
		})

		ctx.On("export", func(_ interface{}) {
			if props.OnExport != nil {
				props.OnExport()
			}
		})

		ctx.OnMounted(func() {
			// Panel mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		isRunning := ctx.Get("isRunning").(*bubbly.Ref[bool]).GetTyped()
		focused := ctx.Get("focused").(*bubbly.Ref[bool]).GetTyped()

		// Build control hints
		var toggleText, toggleKey string
		var toggleColor lipgloss.Color
		if isRunning {
			toggleText = "Stop Profiler"
			toggleKey = "Space"
			toggleColor = lipgloss.Color("196") // Red for stop
		} else {
			toggleText = "Start Profiler"
			toggleKey = "Space"
			toggleColor = lipgloss.Color("35") // Green for start
		}

		// Create control items using Text components
		toggleLabel := components.Text(components.TextProps{
			Content: "[" + toggleKey + "]",
			Color:   lipgloss.Color("99"),
		})
		toggleLabel.Init()

		toggleAction := components.Text(components.TextProps{
			Content: toggleText,
			Color:   toggleColor,
		})
		toggleAction.Init()

		resetLabel := components.Text(components.TextProps{
			Content: "[r]",
			Color:   lipgloss.Color("99"),
		})
		resetLabel.Init()

		resetAction := components.Text(components.TextProps{
			Content: "Reset Metrics",
			Color:   lipgloss.Color("220"),
		})
		resetAction.Init()

		exportLabel := components.Text(components.TextProps{
			Content: "[e]",
			Color:   lipgloss.Color("99"),
		})
		exportLabel.Init()

		exportAction := components.Text(components.TextProps{
			Content: "Export Report",
			Color:   lipgloss.Color("39"),
		})
		exportAction.Init()

		// Build the controls list
		keyWidth := 8
		actionWidth := 20

		keyStyle := lipgloss.NewStyle().Width(keyWidth).Align(lipgloss.Right).PaddingRight(1)
		actionStyle := lipgloss.NewStyle().Width(actionWidth).Align(lipgloss.Left)

		controlsContent := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top,
				keyStyle.Render(toggleLabel.View()),
				actionStyle.Render(toggleAction.View()),
			),
			lipgloss.JoinHorizontal(lipgloss.Top,
				keyStyle.Render(resetLabel.View()),
				actionStyle.Render(resetAction.View()),
			),
			lipgloss.JoinHorizontal(lipgloss.Top,
				keyStyle.Render(exportLabel.View()),
				actionStyle.Render(exportAction.View()),
			),
		)

		// Add focus indicator
		focusText := ""
		if focused {
			focusText = "◆ Active"
		} else {
			focusText = "◇ Inactive"
		}

		focusIndicator := components.Text(components.TextProps{
			Content: focusText,
			Color:   lipgloss.Color("240"),
		})
		focusIndicator.Init()

		// Combine content
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			focusIndicator.View(),
			"",
			controlsContent,
		)

		// Determine border color based on focus
		borderColor := lipgloss.Color("240") // Gray when unfocused
		if focused {
			borderColor = lipgloss.Color("35") // Green when focused
		}

		// Create custom style with dynamic border color
		cardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1).
			Width(35)

		// Create card content
		card := components.Card(components.CardProps{
			Title:    "🎮 Controls",
			Content:  content,
			Width:    35,
			NoBorder: true,
		})
		card.Init()

		// Wrap with our styled border
		return cardStyle.Render(card.View())
	})

	return builder.Build()
}
