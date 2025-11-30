// Package components provides focused UI components for the CPU profiler example.
package components

import (
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/11-profiler/cpu/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// FocusPane represents which pane is currently focused.
type FocusPane int

const (
	// FocusProfile indicates the profile panel is focused.
	FocusProfile FocusPane = iota

	// FocusControls indicates the controls panel is focused.
	FocusControls

	// FocusResults indicates the results panel is focused.
	FocusResults
)

// String returns a human-readable string for the focus pane.
func (f FocusPane) String() string {
	switch f {
	case FocusProfile:
		return "Profile"
	case FocusControls:
		return "Controls"
	case FocusResults:
		return "Results"
	default:
		return "Unknown"
	}
}

// ProfilePanelProps defines the props for the ProfilePanel component.
type ProfilePanelProps struct {
	// State is the current profiler state
	State *bubbly.Ref[composables.CPUProfilerState]

	// Filename is the current/last profile filename
	Filename *bubbly.Ref[string]

	// StartTime is when profiling started
	StartTime *bubbly.Ref[time.Time]

	// FileSize is the size of the profile file
	FileSize *bubbly.Ref[int64]

	// Focused indicates if this panel is focused
	Focused *bubbly.Ref[bool]
}

// CreateProfilePanel creates a profile panel component.
// This demonstrates:
// - Using Card component for content container
// - Dynamic content based on state
// - Focus indicator with border color
// - Live duration updates
func CreateProfilePanel(props ProfilePanelProps) (bubbly.Component, error) {
	builder := bubbly.NewComponent("ProfilePanel")

	builder = builder.Setup(func(ctx *bubbly.Context) {
		// Expose props for template access
		ctx.Expose("state", props.State)
		ctx.Expose("filename", props.Filename)
		ctx.Expose("startTime", props.StartTime)
		ctx.Expose("fileSize", props.FileSize)
		ctx.Expose("focused", props.Focused)

		ctx.OnMounted(func() {
			// Profile panel mounted
		})
	})

	builder = builder.Template(func(ctx bubbly.RenderContext) string {
		// Get current values from reactive state
		state := ctx.Get("state").(*bubbly.Ref[composables.CPUProfilerState]).GetTyped()
		filename := ctx.Get("filename").(*bubbly.Ref[string]).GetTyped()
		startTime := ctx.Get("startTime").(*bubbly.Ref[time.Time]).GetTyped()
		fileSize := ctx.Get("fileSize").(*bubbly.Ref[int64]).GetTyped()
		focused := ctx.Get("focused").(*bubbly.Ref[bool]).GetTyped()

		// Build content based on state
		var content string
		var icon string

		switch state {
		case composables.StateIdle:
			icon = "⏸"
			content = "No profile active\n\nPress [Space] to start profiling"

		case composables.StateProfiling:
			icon = "📊"
			duration := time.Since(startTime)
			content = "Profiling to: " + filename + "\n\n"
			content += "Duration: " + composables.FormatDuration(duration) + "\n"
			content += "Status: Recording CPU samples..."

		case composables.StateComplete:
			icon = "✅"
			content = "Profile saved: " + filename + "\n\n"
			content += "File size: " + composables.FormatBytes(fileSize) + "\n"
			if !startTime.IsZero() {
				content += "Duration: " + composables.FormatDuration(time.Since(startTime))
			}
		}

		// Create title with icon
		title := icon + " CPU Profile"

		// Create card with content
		card := components.Card(components.CardProps{
			Title:    title,
			Content:  content,
			Width:    35,
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
