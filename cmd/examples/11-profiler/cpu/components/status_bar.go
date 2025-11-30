package components

import (
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// StatusBarProps defines the props for the StatusBar component.
type StatusBarProps struct {
	// State is the current profiler state
	State *bubbly.Ref[composables.CPUProfilerState]

	// StartTime is when profiling started
	StartTime *bubbly.Ref[time.Time]

	// Filename is the profile filename
	Filename *bubbly.Ref[string]

	// FocusedPane indicates which panel is focused
	FocusedPane *bubbly.Ref[FocusPane]

	// HasResults indicates if analysis results are available
	HasResults *bubbly.Ref[bool]

	// LastError holds the last error message
	LastError *bubbly.Ref[string]
}

// CreateStatusBar creates a status bar component.
// This demonstrates:
// - Using HStack-like layout for horizontal elements
// - Dynamic content based on state
// - Context-aware help text
// - pprof command hint
func CreateStatusBar(props StatusBarProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("StatusBar")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("state", props.State)
		ctx.Expose("startTime", props.StartTime)
		ctx.Expose("filename", props.Filename)
		ctx.Expose("focusedPane", props.FocusedPane)
		ctx.Expose("hasResults", props.HasResults)
		ctx.Expose("lastError", props.LastError)

		ctx.OnMounted(func() {
			// Status bar mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		state := ctx.Get("state").(*bubbly.Ref[composables.CPUProfilerState]).GetTyped()
		startTime := ctx.Get("startTime").(*bubbly.Ref[time.Time]).GetTyped()
		filename := ctx.Get("filename").(*bubbly.Ref[string]).GetTyped()
		focusedPane := ctx.Get("focusedPane").(*bubbly.Ref[FocusPane]).GetTyped()
		hasResults := ctx.Get("hasResults").(*bubbly.Ref[bool]).GetTyped()
		lastError := ctx.Get("lastError").(*bubbly.Ref[string]).GetTyped()

		// Calculate duration if profiling
		duration := time.Duration(0)
		if !startTime.IsZero() {
			duration = time.Since(startTime)
		}

		// Status badge
		var statusText string
		var statusColor lipgloss.Color

		switch state {
		case composables.StateIdle:
			statusText = " ⏸ IDLE "
			statusColor = lipgloss.Color("240") // Gray
		case composables.StateProfiling:
			statusText = " ● PROFILING "
			statusColor = lipgloss.Color("196") // Red
		case composables.StateComplete:
			if hasResults {
				statusText = " ✓ ANALYZED "
				statusColor = lipgloss.Color("99") // Purple
			} else {
				statusText = " ✓ COMPLETE "
				statusColor = lipgloss.Color("220") // Yellow
			}
		}

		statusBadge := lipgloss.NewStyle().
			Background(statusColor).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Render(statusText)

		// Duration (only when profiling or complete)
		durationText := ""
		if state == composables.StateProfiling || state == composables.StateComplete {
			durationStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
			durationText = durationStyle.Render("Duration: " + composables.FormatDuration(duration))
		}

		// Focus indicator
		focusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))
		focusText := focusStyle.Render("Focus: " + focusedPane.String())

		// Help text (context-aware)
		helpText := getHelpText(state, focusedPane, hasResults)
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

		// pprof hint (when complete with results)
		pprofHint := ""
		if state == composables.StateComplete && hasResults && filename != "" {
			pprofStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("35")).
				Italic(true)
			pprofHint = pprofStyle.Render("go tool pprof " + filename)
		}

		// Error display
		errorText := ""
		if lastError != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))
			errorText = errorStyle.Render("Error: " + lastError)
		}

		// Build the status bar
		parts := []string{statusBadge}

		if durationText != "" {
			parts = append(parts, durationText)
		}

		parts = append(parts, focusText)

		if pprofHint != "" {
			parts = append(parts, pprofHint)
		}

		if errorText != "" {
			parts = append(parts, errorText)
		}

		// Join with separators
		separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(" │ ")

		result := ""
		for i, part := range parts {
			if i > 0 {
				result += separator
			}
			result += part
		}

		// Add help text on new line
		result += "\n" + helpStyle.Render(helpText)

		return result
	})

	return builder.Build()
}

// getHelpText returns context-aware help text.
func getHelpText(state composables.CPUProfilerState, focusedPane FocusPane, hasResults bool) string {
	switch state {
	case composables.StateIdle:
		if focusedPane == FocusControls {
			return "[Space] Start profiling • [Tab] Switch pane • [q] Quit"
		}
		return "[Tab] Switch to Controls • [q] Quit"

	case composables.StateProfiling:
		if focusedPane == FocusControls {
			return "[Space] Stop profiling • [Tab] Switch pane • [q] Quit"
		}
		return "[Tab] Switch to Controls • [q] Quit"

	case composables.StateComplete:
		if hasResults {
			return "[r] Reset • [Tab] Switch pane • [q] Quit"
		}
		if focusedPane == FocusControls {
			return "[a] Analyze • [r] Reset • [Tab] Switch pane • [q] Quit"
		}
		return "[Tab] Switch to Controls • [q] Quit"

	default:
		return "[q] Quit"
	}
}
