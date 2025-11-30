package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/components"
	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	bubblyComposables "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// CreateApp creates the root application component.
// This demonstrates:
// - Composable architecture pattern with UseCPUProfiler
// - Using UseInterval for live duration updates
// - Multi-pane focus management (3 panes: Profile, Controls, Results)
// - Dynamic key bindings based on state and focus
// - State machine workflow (Idle → Profiling → Complete → Analyzed)
// - Component composition with child components
func CreateApp() (bubbly.Component, error) {
	builder := bubbly.NewComponent("CPUProfilerApp").
		WithAutoCommands(true).
		// Global key bindings
		WithKeyBinding("tab", "switchFocus", "Switch focus between panels").
		WithKeyBinding("q", "quit", "Quit application").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		// Action key bindings
		WithKeyBinding(" ", "toggle", "Start/Stop profiling").
		WithKeyBinding("a", "analyze", "Analyze results").
		WithKeyBinding("r", "reset", "Reset profiler")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Use CPU profiler composable - encapsulates all profiler logic
		cpuProfiler := composables.UseCPUProfiler(ctx)

		// Create focus state refs
		focusedPane := bubbly.NewRef(components.FocusControls)
		profileFocused := bubbly.NewRef(false)
		controlsFocused := bubbly.NewRef(true)
		resultsFocused := bubbly.NewRef(false)

		// Track if we have analysis results
		hasResults := bubbly.NewRef(false)

		// Use interval for live duration updates when profiling
		interval := bubblyComposables.UseInterval(ctx, func() {
			// Just trigger a re-render to update duration display
			// The duration is calculated fresh in the template
		}, 100*time.Millisecond)

		// Create child components with props
		profilePanel, err := components.CreateProfilePanel(components.ProfilePanelProps{
			State:     cpuProfiler.State,
			Filename:  cpuProfiler.Filename,
			StartTime: cpuProfiler.StartTime,
			FileSize:  cpuProfiler.FileSize,
			Focused:   profileFocused,
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create profile panel: %v", err))
			return
		}

		controlsPanel, err := components.CreateControlsPanel(components.ControlsPanelProps{
			State:      cpuProfiler.State,
			Focused:    controlsFocused,
			HasResults: hasResults,
			OnStart: func() {
				filename := fmt.Sprintf("cpu-%s.prof", time.Now().Format("20060102-150405"))
				if err := cpuProfiler.Start(filename); err != nil {
					// Error is stored in cpuProfiler.LastError
					return
				}
				interval.Start()
			},
			OnStop: func() {
				if err := cpuProfiler.Stop(); err != nil {
					return
				}
				interval.Stop()
			},
			OnAnalyze: func() {
				if err := cpuProfiler.Analyze(); err != nil {
					return
				}
				hasResults.Set(true)
			},
			OnReset: func() {
				cpuProfiler.Reset()
				hasResults.Set(false)
				interval.Stop()
			},
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create controls panel: %v", err))
			return
		}

		resultsPanel, err := components.CreateResultsPanel(components.ResultsPanelProps{
			State:        cpuProfiler.State,
			HotFunctions: cpuProfiler.HotFunctions,
			Filename:     cpuProfiler.Filename,
			Focused:      resultsFocused,
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create results panel: %v", err))
			return
		}

		statusBar, err := components.CreateStatusBar(components.StatusBarProps{
			State:       cpuProfiler.State,
			StartTime:   cpuProfiler.StartTime,
			Filename:    cpuProfiler.Filename,
			FocusedPane: focusedPane,
			HasResults:  hasResults,
			LastError:   cpuProfiler.LastError,
		})
		if err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to create status bar: %v", err))
			return
		}

		// Register event handlers
		ctx.On("switchFocus", func(_ interface{}) {
			current := focusedPane.GetTyped()
			// Cycle: Controls → Profile → Results → Controls
			switch current {
			case components.FocusControls:
				focusedPane.Set(components.FocusProfile)
				controlsFocused.Set(false)
				profileFocused.Set(true)
				resultsFocused.Set(false)
			case components.FocusProfile:
				focusedPane.Set(components.FocusResults)
				controlsFocused.Set(false)
				profileFocused.Set(false)
				resultsFocused.Set(true)
			case components.FocusResults:
				focusedPane.Set(components.FocusControls)
				controlsFocused.Set(true)
				profileFocused.Set(false)
				resultsFocused.Set(false)
			}
		})

		ctx.On("toggle", func(_ interface{}) {
			// Only toggle if Controls panel is focused
			if focusedPane.GetTyped() != components.FocusControls {
				return
			}

			state := cpuProfiler.State.GetTyped()
			switch state {
			case composables.StateIdle:
				// Start profiling
				filename := fmt.Sprintf("cpu-%s.prof", time.Now().Format("20060102-150405"))
				if err := cpuProfiler.Start(filename); err != nil {
					return
				}
				interval.Start()
			case composables.StateProfiling:
				// Stop profiling
				if err := cpuProfiler.Stop(); err != nil {
					return
				}
				interval.Stop()
			}
		})

		ctx.On("analyze", func(_ interface{}) {
			// Only analyze if Controls panel is focused and in complete state
			if focusedPane.GetTyped() != components.FocusControls {
				return
			}
			if cpuProfiler.State.GetTyped() != composables.StateComplete {
				return
			}
			if hasResults.GetTyped() {
				return // Already analyzed
			}

			if err := cpuProfiler.Analyze(); err != nil {
				return
			}
			hasResults.Set(true)
		})

		ctx.On("reset", func(_ interface{}) {
			cpuProfiler.Reset()
			hasResults.Set(false)
			interval.Stop()
		})

		// Expose state for template access
		ctx.Expose("cpuProfiler", cpuProfiler)
		ctx.Expose("focusedPane", focusedPane)

		// Expose child components (auto-initializes them)
		if err := ctx.ExposeComponent("profilePanel", profilePanel); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose profile panel: %v", err))
			return
		}
		if err := ctx.ExposeComponent("controlsPanel", controlsPanel); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose controls panel: %v", err))
			return
		}
		if err := ctx.ExposeComponent("resultsPanel", resultsPanel); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose results panel: %v", err))
			return
		}
		if err := ctx.ExposeComponent("statusBar", statusBar); err != nil {
			ctx.Expose("error", fmt.Sprintf("Failed to expose status bar: %v", err))
			return
		}

		// Lifecycle hooks
		ctx.OnMounted(func() {
			// App mounted - start with profiler in idle state
		})

		ctx.OnUnmounted(func() {
			// Cleanup: stop interval and profiler
			interval.Stop()
			cpuProfiler.Reset()
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
		profilePanel := ctx.Get("profilePanel").(bubbly.Component)
		controlsPanel := ctx.Get("controlsPanel").(bubbly.Component)
		resultsPanel := ctx.Get("resultsPanel").(bubbly.Component)
		statusBar := ctx.Get("statusBar").(bubbly.Component)

		// Create title
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)
		title := titleStyle.Render("🔬 BubblyUI CPU Profiler")

		// Create subtitle
		subtitleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginBottom(1)
		subtitle := subtitleStyle.Render("CPU profiling with pprof integration")

		// Layout panels horizontally
		panelsRow := lipgloss.JoinHorizontal(
			lipgloss.Top,
			profilePanel.View(),
			"  ", // Spacer
			controlsPanel.View(),
			"  ", // Spacer
			resultsPanel.View(),
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
		containerStyle := lipgloss.NewStyle().Padding(1)
		return containerStyle.Render(content)
	})

	return builder.Build()
}
