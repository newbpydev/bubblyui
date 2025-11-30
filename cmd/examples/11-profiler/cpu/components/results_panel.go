package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// ResultsPanelProps defines the props for the ResultsPanel component.
type ResultsPanelProps struct {
	// State is the current profiler state
	State *bubbly.Ref[composables.CPUProfilerState]

	// HotFunctions contains the analyzed hot functions
	HotFunctions *bubbly.Ref[[]composables.HotFunctionInfo]

	// Filename is the profile filename (for pprof hint)
	Filename *bubbly.Ref[string]

	// Focused indicates if this panel is focused
	Focused *bubbly.Ref[bool]
}

// CreateResultsPanel creates a results panel component.
// This demonstrates:
// - Using Card component for content container
// - Dynamic content based on state and results
// - Focus indicator with border color
// - List rendering with ForEach-like pattern
func CreateResultsPanel(props ResultsPanelProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("ResultsPanel")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("state", props.State)
		ctx.Expose("hotFunctions", props.HotFunctions)
		ctx.Expose("filename", props.Filename)
		ctx.Expose("focused", props.Focused)

		ctx.OnMounted(func() {
			// Results panel mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		state := ctx.Get("state").(*bubbly.Ref[composables.CPUProfilerState]).GetTyped()
		hotFunctions := ctx.Get("hotFunctions").(*bubbly.Ref[[]composables.HotFunctionInfo]).GetTyped()
		filename := ctx.Get("filename").(*bubbly.Ref[string]).GetTyped()
		focused := ctx.Get("focused").(*bubbly.Ref[bool]).GetTyped()

		// Build content based on state
		var content string

		switch state {
		case composables.StateIdle:
			content = "No results yet.\n\n"
			content += "Start a CPU profile to\n"
			content += "see hot functions here."

		case composables.StateProfiling:
			content = "Profiling in progress...\n\n"
			content += "Stop profiling to see\n"
			content += "analysis results."

		case composables.StateComplete:
			if len(hotFunctions) == 0 {
				content = "Profile complete.\n\n"
				content += "Press [a] to analyze\n"
				content += "and see hot functions."
			} else {
				// Show hot functions list
				content = renderHotFunctions(hotFunctions)

				// Add pprof hint
				if filename != "" {
					hintStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("240")).
						Italic(true)
					content += "\n" + hintStyle.Render("go tool pprof "+filename)
				}
			}
		}

		// Create card with content
		card := components.Card(components.CardProps{
			Title:    "📊 Hot Functions",
			Content:  content,
			Width:    40,
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

// renderHotFunctions renders the hot functions list.
func renderHotFunctions(functions []composables.HotFunctionInfo) string {
	if len(functions) == 0 {
		return "No hot functions found."
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	percentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("35"))

	samplesStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Header
	result := headerStyle.Render("Top CPU Consumers:") + "\n\n"

	// Show top 5 functions (or fewer if less available)
	maxFuncs := 5
	if len(functions) < maxFuncs {
		maxFuncs = len(functions)
	}

	for i := 0; i < maxFuncs; i++ {
		fn := functions[i]

		// Truncate long function names
		name := fn.Name
		if len(name) > 30 {
			name = "..." + name[len(name)-27:]
		}

		// Format: name (percent%) - samples
		line := fmt.Sprintf("%d. %s\n   %s - %s\n",
			i+1,
			nameStyle.Render(name),
			percentStyle.Render(fmt.Sprintf("%.1f%%", fn.Percent)),
			samplesStyle.Render(fmt.Sprintf("%d samples", fn.Samples)),
		)
		result += line
	}

	if len(functions) > maxFuncs {
		result += samplesStyle.Render(fmt.Sprintf("\n... and %d more", len(functions)-maxFuncs))
	}

	return result
}
