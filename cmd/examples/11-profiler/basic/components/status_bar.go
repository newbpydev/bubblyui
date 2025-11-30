package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// FocusPane represents which panel is currently focused.
type FocusPane int

const (
	// FocusMetrics indicates the metrics panel is focused.
	FocusMetrics FocusPane = iota
	// FocusControls indicates the controls panel is focused.
	FocusControls
)

// String returns the string representation of the focus pane.
func (f FocusPane) String() string {
	switch f {
	case FocusMetrics:
		return "Metrics"
	case FocusControls:
		return "Controls"
	default:
		return "Unknown"
	}
}

// StatusBarProps defines the props for the StatusBar component.
type StatusBarProps struct {
	// IsRunning indicates if the profiler is running
	IsRunning *bubbly.Ref[bool]

	// Duration is the profiling duration
	Duration *bubbly.Computed[interface{}]

	// FocusedPane indicates which panel is focused
	FocusedPane *bubbly.Ref[FocusPane]

	// LastExport holds the last export filename
	LastExport *bubbly.Ref[string]
}

// CreateStatusBar creates a status bar component.
// This demonstrates:
// - Using HStack for horizontal layout
// - Badge component for status indicators
// - Computed values for derived state
func CreateStatusBar(props StatusBarProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("StatusBar")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("isRunning", props.IsRunning)
		ctx.Expose("duration", props.Duration)
		ctx.Expose("focusedPane", props.FocusedPane)
		ctx.Expose("lastExport", props.LastExport)

		ctx.OnMounted(func() {
			// Status bar mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		isRunning := ctx.Get("isRunning").(*bubbly.Ref[bool]).GetTyped()
		durationVal := ctx.Get("duration").(*bubbly.Computed[interface{}]).Get()
		focusedPane := ctx.Get("focusedPane").(*bubbly.Ref[FocusPane]).GetTyped()
		lastExport := ctx.Get("lastExport").(*bubbly.Ref[string]).GetTyped()

		// Convert duration
		duration := time.Duration(0)
		if d, ok := durationVal.(time.Duration); ok {
			duration = d
		}

		// Status badge
		var statusText string
		var statusColor lipgloss.Color
		if isRunning {
			statusText = " ▶ RUNNING "
			statusColor = lipgloss.Color("35") // Green
		} else {
			statusText = " ⏸ STOPPED "
			statusColor = lipgloss.Color("220") // Yellow
		}

		statusBadge := lipgloss.NewStyle().
			Background(statusColor).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Render(statusText)

		// Duration display
		durationText := components.Text(components.TextProps{
			Content: fmt.Sprintf("Duration: %s", formatDuration(duration)),
			Color:   lipgloss.Color("240"),
		})
		durationText.Init()

		// Focus indicator
		focusText := components.Text(components.TextProps{
			Content: fmt.Sprintf("Focus: %s", focusedPane.String()),
			Color:   lipgloss.Color("99"),
		})
		focusText.Init()

		// Export status
		var exportText string
		if lastExport != "" {
			exportText = fmt.Sprintf("✓ Exported: %s", lastExport)
		} else {
			exportText = "No export yet"
		}
		exportStatus := components.Text(components.TextProps{
			Content: exportText,
			Color:   lipgloss.Color("240"),
		})
		exportStatus.Init()

		// Help text
		helpText := components.Text(components.TextProps{
			Content: "[Tab] Switch Focus  [q] Quit",
			Color:   lipgloss.Color("240"),
		})
		helpText.Init()

		// Create spacer
		spacer := components.Spacer(components.SpacerProps{
			Width: 2,
		})
		spacer.Init()

		// Build horizontal layout using lipgloss
		// We use lipgloss.JoinHorizontal since HStack requires bubbly.Component items
		leftSection := lipgloss.JoinHorizontal(
			lipgloss.Center,
			statusBadge,
			"  ",
			durationText.View(),
		)

		centerSection := lipgloss.JoinHorizontal(
			lipgloss.Center,
			focusText.View(),
			"  │  ",
			exportStatus.View(),
		)

		rightSection := helpText.View()

		// Calculate widths for proper spacing
		leftStyle := lipgloss.NewStyle().Width(30)
		centerStyle := lipgloss.NewStyle().Width(40).Align(lipgloss.Center)
		rightStyle := lipgloss.NewStyle().Align(lipgloss.Right)

		// Combine all sections
		statusBar := lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render(leftSection),
			centerStyle.Render(centerSection),
			rightStyle.Render(rightSection),
		)

		// Add top border
		borderStyle := lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			PaddingTop(1).
			Width(100)

		return borderStyle.Render(statusBar)
	})

	return builder.Build()
}

// formatDuration formats a duration for display in the status bar.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
