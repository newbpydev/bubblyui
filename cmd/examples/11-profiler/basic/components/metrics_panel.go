// Package components provides UI components for the profiler example.
package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/basic/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// MetricsPanelProps defines the props for the MetricsPanel component.
type MetricsPanelProps struct {
	// Metrics holds the current profiler metrics
	Metrics *bubbly.Ref[*composables.ProfilerMetrics]

	// Focused indicates if this panel has focus
	Focused *bubbly.Ref[bool]

	// IsRunning indicates if the profiler is running
	IsRunning *bubbly.Ref[bool]
}

// CreateMetricsPanel creates a component that displays profiler metrics.
// This demonstrates:
// - Component factory pattern
// - Props-based composition
// - Using BubblyUI Card and Text components
// - Dynamic styling based on focus state
func CreateMetricsPanel(props MetricsPanelProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("MetricsPanel")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("metrics", props.Metrics)
		ctx.Expose("focused", props.Focused)
		ctx.Expose("isRunning", props.IsRunning)

		// Lifecycle hook
		ctx.OnMounted(func() {
			// Panel mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		metrics := ctx.Get("metrics").(*bubbly.Ref[*composables.ProfilerMetrics]).GetTyped()
		focused := ctx.Get("focused").(*bubbly.Ref[bool]).GetTyped()
		isRunning := ctx.Get("isRunning").(*bubbly.Ref[bool]).GetTyped()

		// Build metrics content using Text components
		var metricsContent string

		// FPS metric
		fpsLabel := components.Text(components.TextProps{
			Content: "FPS:",
			Color:   lipgloss.Color("240"),
		})
		fpsLabel.Init()

		fpsValue := components.Text(components.TextProps{
			Content: fmt.Sprintf("%.1f", metrics.FPS),
			Color:   getFPSColor(metrics.FPS),
		})
		fpsValue.Init()

		// Memory metric
		memLabel := components.Text(components.TextProps{
			Content: "Memory:",
			Color:   lipgloss.Color("240"),
		})
		memLabel.Init()

		memValue := components.Text(components.TextProps{
			Content: composables.FormatBytes(metrics.MemoryUsage),
			Color:   getMemoryColor(metrics.MemoryUsage),
		})
		memValue.Init()

		// Goroutines metric
		gorLabel := components.Text(components.TextProps{
			Content: "Goroutines:",
			Color:   lipgloss.Color("240"),
		})
		gorLabel.Init()

		gorValue := components.Text(components.TextProps{
			Content: fmt.Sprintf("%d", metrics.GoroutineCount),
			Color:   lipgloss.Color("39"),
		})
		gorValue.Init()

		// Renders metric
		renderLabel := components.Text(components.TextProps{
			Content: "Renders:",
			Color:   lipgloss.Color("240"),
		})
		renderLabel.Init()

		renderValue := components.Text(components.TextProps{
			Content: fmt.Sprintf("%d", metrics.RenderCount),
			Color:   lipgloss.Color("213"),
		})
		renderValue.Init()

		// Bottlenecks metric
		bottleneckLabel := components.Text(components.TextProps{
			Content: "Bottlenecks:",
			Color:   lipgloss.Color("240"),
		})
		bottleneckLabel.Init()

		bottleneckColor := lipgloss.Color("35") // Green = good
		if metrics.BottleneckCount > 0 {
			bottleneckColor = lipgloss.Color("196") // Red = issues
		}
		bottleneckValue := components.Text(components.TextProps{
			Content: fmt.Sprintf("%d", metrics.BottleneckCount),
			Color:   bottleneckColor,
		})
		bottleneckValue.Init()

		// Samples metric
		samplesLabel := components.Text(components.TextProps{
			Content: "Samples:",
			Color:   lipgloss.Color("240"),
		})
		samplesLabel.Init()

		samplesValue := components.Text(components.TextProps{
			Content: fmt.Sprintf("%d", metrics.SampleCount),
			Color:   lipgloss.Color("99"),
		})
		samplesValue.Init()

		// Build the metrics grid using lipgloss for layout
		rowStyle := lipgloss.NewStyle().Width(30)
		labelWidth := 12
		valueWidth := 15

		labelStyle := lipgloss.NewStyle().Width(labelWidth).Align(lipgloss.Right).PaddingRight(1)
		valueStyle := lipgloss.NewStyle().Width(valueWidth).Align(lipgloss.Left)

		metricsContent = lipgloss.JoinVertical(
			lipgloss.Left,
			rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(fpsLabel.View()),
				valueStyle.Render(fpsValue.View()),
			)),
			rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(memLabel.View()),
				valueStyle.Render(memValue.View()),
			)),
			rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(gorLabel.View()),
				valueStyle.Render(gorValue.View()),
			)),
			rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(renderLabel.View()),
				valueStyle.Render(renderValue.View()),
			)),
			rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(bottleneckLabel.View()),
				valueStyle.Render(bottleneckValue.View()),
			)),
			rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(samplesLabel.View()),
				valueStyle.Render(samplesValue.View()),
			)),
		)

		// Add status indicator
		statusText := "⏸ Paused"
		statusColor := lipgloss.Color("220") // Yellow
		if isRunning {
			statusText = "▶ Running"
			statusColor = lipgloss.Color("35") // Green
		}

		statusBadge := components.Text(components.TextProps{
			Content: statusText,
			Color:   statusColor,
		})
		statusBadge.Init()

		// Combine content
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			statusBadge.View(),
			"",
			metricsContent,
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
			Title:    "📊 Metrics",
			Content:  content,
			Width:    35,
			NoBorder: true, // We'll apply our own border
		})
		card.Init()

		// Wrap with our styled border
		return cardStyle.Render(card.View())
	})

	return builder.Build()
}

// getFPSColor returns a color based on FPS value.
func getFPSColor(fps float64) lipgloss.Color {
	switch {
	case fps >= 60:
		return lipgloss.Color("35") // Green - excellent
	case fps >= 30:
		return lipgloss.Color("220") // Yellow - acceptable
	default:
		return lipgloss.Color("196") // Red - poor
	}
}

// getMemoryColor returns a color based on memory usage.
func getMemoryColor(bytes uint64) lipgloss.Color {
	const (
		MB = 1024 * 1024
	)
	switch {
	case bytes < 50*MB:
		return lipgloss.Color("35") // Green - low
	case bytes < 200*MB:
		return lipgloss.Color("220") // Yellow - moderate
	default:
		return lipgloss.Color("196") // Red - high
	}
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}
