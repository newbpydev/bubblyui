package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/basic/components"
	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/basic/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	bubblyComposables "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// CreateApp creates the root application component.
// This demonstrates:
// - Composable architecture pattern
// - Using UseProfiler composable for profiler logic
// - Using UseInterval for live metric updates
// - Multi-pane focus management
// - Dynamic key bindings based on focus state
// - Component composition with child components
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("ProfilerApp").
		WithAutoCommands(true).
		// Global key bindings
		WithKeyBinding("tab", "switchFocus", "Switch focus between panels").
		WithKeyBinding("q", "quit", "Quit application").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		// Action key bindings (work when Controls panel is focused)
		WithKeyBinding(" ", "toggle", "Toggle profiler").
		WithKeyBinding("r", "reset", "Reset metrics").
		WithKeyBinding("e", "export", "Export report")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Use profiler composable - encapsulates all profiler logic
		profiler := composables.UseProfiler(ctx)

		// Create focus state refs
		focusedPane := bubbly.NewRef(components.FocusControls)
		metricsFocused := bubbly.NewRef(false)
		controlsFocused := bubbly.NewRef(true)

		// Use interval for live metric updates when profiler is running
		// This demonstrates UseInterval composable
		interval := bubblyComposables.UseInterval(ctx, func() {
			if profiler.IsRunning.GetTyped() {
				profiler.RefreshMetrics()
			}
		}, 100*time.Millisecond)

		// Start the interval
		interval.Start()

		// Create child components with props
		metricsPanel, err := components.CreateMetricsPanel(components.MetricsPanelProps{
			Metrics:   profiler.Metrics,
			Focused:   metricsFocused,
			IsRunning: profiler.IsRunning,
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create metrics panel: %v", err))
			return
		}

		controlsPanel, err := components.CreateControlsPanel(components.ControlsPanelProps{
			IsRunning: profiler.IsRunning,
			Focused:   controlsFocused,
			OnToggle:  profiler.Toggle,
			OnReset:   profiler.Reset,
			OnExport: func() {
				filename := fmt.Sprintf("profiler-report-%s.html", time.Now().Format("20060102-150405"))
				if err := profiler.ExportReport(filename); err != nil {
					// In a real app, we'd show an error notification
					fmt.Printf("Export error: %v\n", err)
				}
			},
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create controls panel: %v", err))
			return
		}

		statusBar, err := components.CreateStatusBar(components.StatusBarProps{
			IsRunning:   profiler.IsRunning,
			StartTime:   profiler.StartTime,
			FocusedPane: focusedPane,
			LastExport:  profiler.LastExport,
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create status bar: %v", err))
			return
		}

		// Register event handlers
		ctx.On("switchFocus", func(_ interface{}) {
			current := focusedPane.GetTyped()
			if current == components.FocusMetrics {
				focusedPane.Set(components.FocusControls)
				metricsFocused.Set(false)
				controlsFocused.Set(true)
			} else {
				focusedPane.Set(components.FocusMetrics)
				metricsFocused.Set(true)
				controlsFocused.Set(false)
			}
		})

		ctx.On("toggle", func(_ interface{}) {
			// Only toggle if Controls panel is focused
			if focusedPane.GetTyped() == components.FocusControls {
				profiler.Toggle()
			}
		})

		ctx.On("reset", func(_ interface{}) {
			profiler.Reset()
		})

		ctx.On("export", func(_ interface{}) {
			filename := fmt.Sprintf("profiler-report-%s.html", time.Now().Format("20060102-150405"))
			if err := profiler.ExportReport(filename); err != nil {
				fmt.Printf("Export error: %v\n", err)
			}
		})

		// Expose state for template access
		ctx.Expose("profiler", profiler)
		ctx.Expose("focusedPane", focusedPane)

		// Expose child components (auto-initializes them)
		if err := ctx.ExposeComponent("metricsPanel", metricsPanel); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose metrics panel: %v", err))
			return
		}
		if err := ctx.ExposeComponent("controlsPanel", controlsPanel); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose controls panel: %v", err))
			return
		}
		if err := ctx.ExposeComponent("statusBar", statusBar); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose status bar: %v", err))
			return
		}

		// Lifecycle hooks
		ctx.OnMounted(func() {
			// App mounted - start with profiler stopped
		})

		ctx.OnUnmounted(func() {
			// Cleanup: stop interval and profiler
			interval.Stop()
			profiler.Stop()
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Check for errors
		if errVal := ctx.Get("error"); errVal != nil {
			if errStr, ok := errVal.(string); ok {
				return fmt.Sprintf("Error: %s", errStr)
			}
		}

		// Get child components
		metricsPanel := ctx.Get("metricsPanel").(bubbly.Component)
		controlsPanel := ctx.Get("controlsPanel").(bubbly.Component)
		statusBar := ctx.Get("statusBar").(bubbly.Component)

		// Create title
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)
		title := titleStyle.Render("🔬 BubblyUI Performance Profiler")

		// Create subtitle
		subtitleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginBottom(1)
		subtitle := subtitleStyle.Render("Real-time performance monitoring and analysis")

		// Layout panels horizontally
		panelsRow := lipgloss.JoinHorizontal(
			lipgloss.Top,
			metricsPanel.View(),
			"  ", // Spacer
			controlsPanel.View(),
		)

		// Combine all sections vertically
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			subtitle,
			"",
			panelsRow,
			"",
			statusBar.View(),
		)

		// Add padding
		containerStyle := lipgloss.NewStyle().Padding(2)
		return containerStyle.Render(content)
	})

	return builder.Build()
}
