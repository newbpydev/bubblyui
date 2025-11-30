package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// ControlsPanelProps defines the props for the ControlsPanel component.
type ControlsPanelProps struct {
	// State is the current profiler state
	State *bubbly.Ref[composables.CPUProfilerState]

	// Focused indicates if this panel is focused
	Focused *bubbly.Ref[bool]

	// HasResults indicates if analysis results are available
	HasResults *bubbly.Ref[bool]

	// OnStart is called when the start action is triggered
	OnStart func()

	// OnStop is called when the stop action is triggered
	OnStop func()

	// OnAnalyze is called when the analyze action is triggered
	OnAnalyze func()

	// OnReset is called when the reset action is triggered
	OnReset func()
}

// CreateControlsPanel creates a controls panel component.
// This demonstrates:
// - Using Card component for content container
// - Dynamic controls based on state
// - Focus indicator with border color
// - Callback props for actions
func CreateControlsPanel(props ControlsPanelProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("ControlsPanel")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("state", props.State)
		ctx.Expose("focused", props.Focused)
		ctx.Expose("hasResults", props.HasResults)

		// Register event handlers for actions
		ctx.On("start", func(_ interface{}) {
			if props.OnStart != nil {
				props.OnStart()
			}
		})

		ctx.On("stop", func(_ interface{}) {
			if props.OnStop != nil {
				props.OnStop()
			}
		})

		ctx.On("analyze", func(_ interface{}) {
			if props.OnAnalyze != nil {
				props.OnAnalyze()
			}
		})

		ctx.On("reset", func(_ interface{}) {
			if props.OnReset != nil {
				props.OnReset()
			}
		})

		ctx.OnMounted(func() {
			// Controls panel mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		state := ctx.Get("state").(*bubbly.Ref[composables.CPUProfilerState]).GetTyped()
		focused := ctx.Get("focused").(*bubbly.Ref[bool]).GetTyped()
		hasResults := ctx.Get("hasResults").(*bubbly.Ref[bool]).GetTyped()

		// Build controls based on state
		var content string

		// Status indicator
		var statusText string
		var statusColor lipgloss.Color

		switch state {
		case composables.StateIdle:
			statusText = "▶ Ready"
			statusColor = lipgloss.Color("35") // Green
		case composables.StateProfiling:
			statusText = "● Recording"
			statusColor = lipgloss.Color("196") // Red
		case composables.StateComplete:
			if hasResults {
				statusText = "✓ Analyzed"
				statusColor = lipgloss.Color("99") // Purple
			} else {
				statusText = "✓ Complete"
				statusColor = lipgloss.Color("220") // Yellow
			}
		}

		statusStyle := lipgloss.NewStyle().
			Foreground(statusColor).
			Bold(true)
		content += statusStyle.Render(statusText) + "\n\n"

		// Action hints based on state
		actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
		disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		switch state {
		case composables.StateIdle:
			content += actionStyle.Render("[Space] Start Profiling") + "\n"
			content += disabledStyle.Render("[a] Analyze (disabled)") + "\n"
			content += disabledStyle.Render("[r] Reset (disabled)")

		case composables.StateProfiling:
			content += actionStyle.Render("[Space] Stop Profiling") + "\n"
			content += disabledStyle.Render("[a] Analyze (disabled)") + "\n"
			content += disabledStyle.Render("[r] Reset (disabled)")

		case composables.StateComplete:
			content += disabledStyle.Render("[Space] Start (disabled)") + "\n"
			if hasResults {
				content += disabledStyle.Render("[a] Analyzed ✓") + "\n"
			} else {
				content += actionStyle.Render("[a] Analyze Results") + "\n"
			}
			content += actionStyle.Render("[r] Reset Profiler")
		}

		// Create card with content
		card := components.Card(components.CardProps{
			Title:    "⚙ Controls",
			Content:  content,
			Width:    28,
			NoBorder: true,
		})
		card.Init()

		// Apply custom border based on focus
		borderColor := lipgloss.Color("240") // Gray when not focused
		if focused {
			borderColor = lipgloss.Color("35") // Green when focused
		}

		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

		return borderStyle.Render(card.View())
	})

	return builder.Build()
}
