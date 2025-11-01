package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// ProfileTickMsg triggers profiling updates
type ProfileTickMsg time.Time

// model wraps the profiling demo component for Bubbletea integration
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.component.Init(),
		profileTickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "s":
			// Start profiling
			m.component.Emit("startProfiling", nil)
		case "p":
			// Stop profiling
			m.component.Emit("stopProfiling", nil)
		case "c":
			// Create workload
			m.component.Emit("createWorkload", nil)
		case "r":
			// Run profile
			m.component.Emit("runProfile", nil)
		}

	case ProfileTickMsg:
		m.component.Emit("updateProfile", nil)
		return m, profileTickCmd()
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("🔍 Profiling Demo - Phase 8 Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: pprof integration, composable profiling, performance analysis",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"s: start profiling • p: stop profiling • c: create workload • r: run profile • q: quit\n" +
			"pprof endpoints available at: http://localhost:6060/debug/pprof/",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// profileTickCmd sends profile update messages every 2s
func profileTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return ProfileTickMsg(t)
	})
}

// createProfilingDemo creates the profiling demo component
func createProfilingDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("ProfilingDemo").
		Setup(func(ctx *bubbly.Context) {
			// State
			profilingEnabled := composables.UseState(ctx, false)
			profilingAddress := composables.UseState(ctx, "")
			workloadSize := composables.UseState(ctx, 0)
			profileSummary := composables.UseState(ctx, "No profile data yet")
			status := composables.UseState(ctx, "Ready")

			// Expose state
			ctx.Expose("profilingEnabled", profilingEnabled.Value)
			ctx.Expose("profilingAddress", profilingAddress.Value)
			ctx.Expose("workloadSize", workloadSize.Value)
			ctx.Expose("profileSummary", profileSummary.Value)
			ctx.Expose("status", status.Value)

			// Start profiling
			ctx.On("startProfiling", func(_ interface{}) {
				if profilingEnabled.Get() {
					status.Set("Profiling already enabled")
					return
				}

				addr := "localhost:6060"
				if err := monitoring.EnableProfiling(addr); err != nil {
					status.Set(fmt.Sprintf("Error: %v", err))
					return
				}

				profilingEnabled.Set(true)
				profilingAddress.Set(addr)
				status.Set("Profiling started on " + addr)
			})

			// Stop profiling
			ctx.On("stopProfiling", func(_ interface{}) {
				if !profilingEnabled.Get() {
					status.Set("Profiling not enabled")
					return
				}

				monitoring.StopProfiling()
				profilingEnabled.Set(false)
				profilingAddress.Set("")
				status.Set("Profiling stopped")
			})

			// Create workload
			ctx.On("createWorkload", func(_ interface{}) {
				// Create various composables to generate workload
				for i := 0; i < 100; i++ {
					_ = composables.UseState(ctx, i)
				}

				for i := 0; i < 50; i++ {
					_ = composables.UseAsync(ctx, func() (*int, error) {
						result := 42
						return &result, nil
					})
				}

				workloadSize.Set(workloadSize.Get() + 150)
				status.Set(fmt.Sprintf("Created 150 composables (total: %d)", workloadSize.Get()))
			})

			// Run profile
			ctx.On("runProfile", func(_ interface{}) {
				status.Set("Running composable profiling...")

				// Profile composables for 1 second
				profile := monitoring.ProfileComposables(1 * time.Second)

				if profile != nil {
					summary := profile.Summary()
					profileSummary.Set(summary)
					status.Set("Profile complete")
				} else {
					status.Set("Error: Failed to create profile")
				}
			})

			// Update profile display
			ctx.On("updateProfile", func(_ interface{}) {
				if monitoring.IsProfilingEnabled() {
					addr := monitoring.GetProfilingAddress()
					profilingAddress.Set(addr)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			profilingEnabled := ctx.Get("profilingEnabled").(*bubbly.Ref[bool])
			profilingAddress := ctx.Get("profilingAddress").(*bubbly.Ref[string])
			workloadSize := ctx.Get("workloadSize").(*bubbly.Ref[int])
			profileSummary := ctx.Get("profileSummary").(*bubbly.Ref[string])
			status := ctx.Get("status").(*bubbly.Ref[string])

			enabled := profilingEnabled.GetTyped()
			addr := profilingAddress.GetTyped()
			size := workloadSize.GetTyped()
			summary := profileSummary.GetTyped()
			statusMsg := status.GetTyped()

			// Status panel
			statusColor := "86" // Green
			statusText := "🟢 Profiling Enabled"
			if !enabled {
				statusColor = "241" // Gray
				statusText = "⚫ Profiling Disabled"
			}

			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color(statusColor)).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(35)

			statusPanel := statusStyle.Render(
				fmt.Sprintf("%s\n\n%s", statusText, statusMsg),
			)

			// Workload panel
			workloadStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(35)

			workloadPanel := workloadStyle.Render(
				fmt.Sprintf("Workload Size\n\n%d composables", size),
			)

			topRow := lipgloss.JoinHorizontal(lipgloss.Top, statusPanel, "  ", workloadPanel)

			// Profiling info panel
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(75).
				Height(8)

			infoContent := ""
			if enabled && addr != "" {
				infoContent = fmt.Sprintf(
					"🔍 pprof Endpoints Available:\n\n"+
						"• CPU Profile:       curl http://%s/debug/pprof/profile?seconds=10\n"+
						"• Heap Profile:      curl http://%s/debug/pprof/heap\n"+
						"• Goroutine Profile: curl http://%s/debug/pprof/goroutine\n"+
						"• All Profiles:      http://%s/debug/pprof/",
					addr, addr, addr, addr,
				)
			} else {
				infoContent = "Press 's' to start profiling and enable pprof endpoints"
			}

			infoPanel := infoStyle.Render(infoContent)

			// Profile summary panel
			summaryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(75).
				Height(10)

			summaryTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Render("Composable Profile Summary")

			summaryPanel := summaryStyle.Render(summaryTitle + "\n\n" + summary)

			// Features panel
			featuresStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(75)

			featuresPanel := featuresStyle.Render(
				"Phase 8 Features Demonstrated:\n\n" +
					"• Profiling Utilities (Task 8.7)\n" +
					"  - EnableProfiling/StopProfiling\n" +
					"  - Standard pprof endpoints\n" +
					"  - ProfileComposables for custom profiling\n" +
					"  - Thread-safe operations\n\n" +
					"• Production Profiling Best Practices\n" +
					"  - Localhost binding (security)\n" +
					"  - Opt-in activation\n" +
					"  - Graceful shutdown",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				topRow,
				"",
				infoPanel,
				"",
				summaryPanel,
				"",
				featuresPanel,
			)
		}).
		Build()
}

func main() {
	// Create profiling demo component
	component, err := createProfilingDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	m := model{component: component}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
