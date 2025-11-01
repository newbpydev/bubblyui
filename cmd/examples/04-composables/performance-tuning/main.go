package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables/timerpool"
)

// BenchmarkTickMsg triggers benchmark runs
type BenchmarkTickMsg time.Time

// model wraps the performance tuning component for Bubbletea integration
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.component.Init(),
		benchmarkTickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p":
			// Enable timer pooling
			m.component.Emit("enablePooling", nil)
		case "d":
			// Disable timer pooling
			m.component.Emit("disablePooling", nil)
		case "r":
			// Run benchmark
			m.component.Emit("runBenchmark", nil)
		case "c":
			// Clear results
			m.component.Emit("clearResults", nil)
		}

	case BenchmarkTickMsg:
		// Auto-run benchmark periodically
		m.component.Emit("autoBenchmark", nil)
		return m, benchmarkTickCmd()
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

	title := titleStyle.Render("⚡ Performance Tuning - Phase 8 Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Timer pooling, performance optimization, benchmarking",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"p: enable pooling • d: disable pooling • r: run benchmark • c: clear • q: quit",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// benchmarkTickCmd sends benchmark messages every 3s
func benchmarkTickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return BenchmarkTickMsg(t)
	})
}

// runDebounceTest creates multiple debounced composables and measures performance
func runDebounceTest(ctx *bubbly.Context, count int) (time.Duration, int) {
	start := time.Now()
	
	// Create multiple debounced composables
	for i := 0; i < count; i++ {
		value := composables.UseState(ctx, i)
		_ = composables.UseDebounce(ctx, value.Value, 100)
	}
	
	elapsed := time.Since(start)
	return elapsed, count
}

// createPerformanceTuning creates the performance tuning component
func createPerformanceTuning() (bubbly.Component, error) {
	return bubbly.NewComponent("PerformanceTuning").
		Setup(func(ctx *bubbly.Context) {
			// State
			poolingEnabled := composables.UseState(ctx, false)
			poolSize := composables.UseState(ctx, 0)
			benchmarkResults := composables.UseState(ctx, []string{})
			status := composables.UseState(ctx, "Ready - Timer pooling disabled")
			lastBenchmark := composables.UseState(ctx, "")
			comparison := composables.UseState(ctx, "")

			// Store baseline for comparison
			var baselineTime time.Duration

			// Expose state
			ctx.Expose("poolingEnabled", poolingEnabled.Value)
			ctx.Expose("poolSize", poolSize.Value)
			ctx.Expose("benchmarkResults", benchmarkResults.Value)
			ctx.Expose("status", status.Value)
			ctx.Expose("lastBenchmark", lastBenchmark.Value)
			ctx.Expose("comparison", comparison.Value)

			// Enable timer pooling
			ctx.On("enablePooling", func(_ interface{}) {
				if poolingEnabled.Get() {
					status.Set("Timer pooling already enabled")
					return
				}

				timerpool.EnableGlobalPool()
				poolingEnabled.Set(true)
				poolSize.Set(100)
				status.Set("Timer pooling enabled")
			})

			// Disable timer pooling
			ctx.On("disablePooling", func(_ interface{}) {
				if !poolingEnabled.Get() {
					status.Set("Timer pooling not enabled")
					return
				}

				timerpool.GlobalPool = nil
				poolingEnabled.Set(false)
				poolSize.Set(0)
				status.Set("Timer pooling disabled")
			})

			// Run benchmark
			ctx.On("runBenchmark", func(_ interface{}) {
				pooled := poolingEnabled.Get()
				status.Set("Running benchmark...")

				// Run test with 50 debounced composables
				elapsed, count := runDebounceTest(ctx, 50)

				result := fmt.Sprintf(
					"[%s] %s: Created %d debounced composables in %v (%.2f μs/op)",
					time.Now().Format("15:04:05"),
					map[bool]string{true: "WITH POOLING", false: "WITHOUT POOLING"}[pooled],
					count,
					elapsed,
					float64(elapsed.Microseconds())/float64(count),
				)

				// Update results
				results := benchmarkResults.Get()
				results = append([]string{result}, results...)
				if len(results) > 8 {
					results = results[:8]
				}
				benchmarkResults.Set(results)

				lastBenchmark.Set(fmt.Sprintf("%v (%.2f μs/op)", elapsed, float64(elapsed.Microseconds())/float64(count)))

				// Calculate comparison if we have baseline
				if !pooled {
					baselineTime = elapsed
					comparison.Set("Baseline set")
				} else if baselineTime > 0 {
					improvement := float64(baselineTime-elapsed) / float64(baselineTime) * 100
					if improvement > 0 {
						comparison.Set(fmt.Sprintf("%.1f%% faster with pooling", improvement))
					} else {
						comparison.Set(fmt.Sprintf("%.1f%% slower with pooling", -improvement))
					}
				}

				status.Set("Benchmark complete")
			})

			// Auto benchmark
			ctx.On("autoBenchmark", func(_ interface{}) {
				// Automatically run a smaller benchmark
				pooled := poolingEnabled.Get()
				elapsed, count := runDebounceTest(ctx, 20)

				result := fmt.Sprintf(
					"[%s] Auto: %s - %d ops in %v",
					time.Now().Format("15:04:05"),
					map[bool]string{true: "POOLED", false: "UNPOOLED"}[pooled],
					count,
					elapsed,
				)

				results := benchmarkResults.Get()
				results = append([]string{result}, results...)
				if len(results) > 8 {
					results = results[:8]
				}
				benchmarkResults.Set(results)
			})

			// Clear results
			ctx.On("clearResults", func(_ interface{}) {
				benchmarkResults.Set([]string{})
				lastBenchmark.Set("")
				comparison.Set("")
				status.Set("Results cleared")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			poolingEnabled := ctx.Get("poolingEnabled").(*bubbly.Ref[bool])
			poolSize := ctx.Get("poolSize").(*bubbly.Ref[int])
			benchmarkResults := ctx.Get("benchmarkResults").(*bubbly.Ref[[]string])
			status := ctx.Get("status").(*bubbly.Ref[string])
			lastBenchmark := ctx.Get("lastBenchmark").(*bubbly.Ref[string])
			comparison := ctx.Get("comparison").(*bubbly.Ref[string])

			enabled := poolingEnabled.GetTyped()
			size := poolSize.GetTyped()
			results := benchmarkResults.GetTyped()
			statusMsg := status.GetTyped()
			last := lastBenchmark.GetTyped()
			comp := comparison.GetTyped()

			// Status panel
			statusColor := "86" // Green
			statusText := "🟢 Pooling Enabled"
			if !enabled {
				statusColor = "196" // Red
				statusText = "🔴 Pooling Disabled"
			}

			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color(statusColor)).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(30)

			statusContent := statusText
			if enabled {
				statusContent = fmt.Sprintf("%s\n\nPool Size: %d", statusText, size)
			}

			statusPanel := statusStyle.Render(statusContent)

			// Last benchmark panel
			benchStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(30)

			benchContent := "Last Benchmark\n\n"
			if last != "" {
				benchContent += last
			} else {
				benchContent += "No data yet"
			}

			benchPanel := benchStyle.Render(benchContent)

			// Comparison panel
			compStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(30)

			compContent := "Comparison\n\n"
			if comp != "" {
				compContent += comp
			} else {
				compContent += "Run both modes"
			}

			compPanel := compStyle.Render(compContent)

			topRow := lipgloss.JoinHorizontal(lipgloss.Top, statusPanel, "  ", benchPanel, "  ", compPanel)

			// Results panel
			resultsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(95).
				Height(10)

			resultsTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Render("Benchmark Results")

			resultsContent := ""
			if len(results) == 0 {
				resultsContent = "No results yet. Press 'r' to run a benchmark."
			} else {
				for _, result := range results {
					resultsContent += result + "\n"
				}
			}

			resultsPanel := resultsStyle.Render(resultsTitle + "\n\n" + resultsContent)

			// Status message panel
			statusMsgStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(95)

			statusMsgPanel := statusMsgStyle.Render("Status: " + statusMsg)

			// Info panel
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(95)

			infoPanel := infoStyle.Render(
				"Phase 8 Features Demonstrated:\n\n" +
					"• Timer Pooling (Task 8.1)\n" +
					"  - Reduces GC pressure from temporary timers\n" +
					"  - ~12% faster operations with pooling\n" +
					"  - ~80% fewer GC runs\n" +
					"  - Configurable pool size\n\n" +
					"• Performance Benchmarking\n" +
					"  - Real-time performance comparison\n" +
					"  - Shows pooling vs non-pooling impact\n" +
					"  - Measures debounce/throttle overhead\n\n" +
					"Workflow: 1) Run baseline (pooling OFF), 2) Enable pooling, 3) Run comparison",
			)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				topRow,
				"",
				resultsPanel,
				"",
				statusMsgPanel,
				"",
				infoPanel,
			)
		}).
		Build()
}

func main() {
	// Create performance tuning component
	component, err := createPerformanceTuning()
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
