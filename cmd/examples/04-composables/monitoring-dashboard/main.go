package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
)

// TickMsg is sent by the ticker to trigger updates
type TickMsg time.Time

// MetricsTickMsg triggers metrics refresh
type MetricsTickMsg time.Time

// model wraps the dashboard component for Bubbletea integration
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.component.Init(),
		tickCmd(),
		metricsTickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			// Trigger random activity
			m.component.Emit("randomActivity", nil)
		case "c":
			// Create composables
			m.component.Emit("createComposables", nil)
		case "m":
			// Toggle metrics display
			m.component.Emit("toggleMetrics", nil)
		}

	case TickMsg:
		// Simulate activity
		m.component.Emit("tick", nil)
		return m, tickCmd()

	case MetricsTickMsg:
		// Update metrics display
		m.component.Emit("metricsUpdate", nil)
		return m, metricsTickCmd()
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

	title := titleStyle.Render("📊 Monitoring Dashboard - Phase 8 Example")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Metrics collection, Prometheus integration, real-time monitoring",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"r: random activity • c: create composables • m: toggle metrics • q: quit\n" +
			"Prometheus metrics available at: http://localhost:9090/metrics",
	)

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// tickCmd sends tick messages every 500ms
func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// metricsTickCmd sends metrics update messages every 1s
func metricsTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return MetricsTickMsg(t)
	})
}

// createMonitoringDashboard creates the dashboard component
func createMonitoringDashboard() (bubbly.Component, error) {
	return bubbly.NewComponent("MonitoringDashboard").
		Setup(func(ctx *bubbly.Context) {
			// State for tracking metrics
			totalComposables := composables.UseState(ctx, 0)
			cacheHits := composables.UseState(ctx, 0)
			cacheMisses := composables.UseState(ctx, 0)
			activityLog := composables.UseState(ctx, []string{})
			showMetrics := composables.UseState(ctx, true)

			// Expose state
			ctx.Expose("totalComposables", totalComposables.Value)
			ctx.Expose("cacheHits", cacheHits.Value)
			ctx.Expose("cacheMisses", cacheMisses.Value)
			ctx.Expose("activityLog", activityLog.Value)
			ctx.Expose("showMetrics", showMetrics.Value)

			// Random activity generator
			ctx.On("randomActivity", func(_ interface{}) {
				// Simulate various activities
				activities := []string{
					"UseState created",
					"UseForm created",
					"UseAsync executed",
					"Cache hit on GetFieldIndex",
					"Cache miss on GetFieldType",
					"Provide/Inject called",
				}

				activity := activities[rand.Intn(len(activities))]

				// Update activity log
				log := activityLog.Get()
				log = append([]string{fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), activity)}, log...)
				if len(log) > 10 {
					log = log[:10]
				}
				activityLog.Set(log)

				// Update metrics based on activity
				totalComposables.Set(totalComposables.Get() + 1)

				if rand.Float32() > 0.3 {
					cacheHits.Set(cacheHits.Get() + 1)
				} else {
					cacheMisses.Set(cacheMisses.Get() + 1)
				}
			})

			// Create composables to generate metrics
			ctx.On("createComposables", func(_ interface{}) {
				// Create various composables to generate real metrics
				_ = composables.UseState(ctx, "test")
				_ = composables.UseState(ctx, 42)
				_ = composables.UseAsync(ctx, func() (*string, error) {
					result := "async data"
					return &result, nil
				})

				totalComposables.Set(totalComposables.Get() + 3)

				log := activityLog.Get()
				log = append([]string{
					fmt.Sprintf("[%s] Created 3 composables", time.Now().Format("15:04:05")),
				}, log...)
				if len(log) > 10 {
					log = log[:10]
				}
				activityLog.Set(log)
			})

			// Toggle metrics display
			ctx.On("toggleMetrics", func(_ interface{}) {
				showMetrics.Set(!showMetrics.Get())
			})

			// Periodic tick
			ctx.On("tick", func(_ interface{}) {
				// Randomly generate activity
				if rand.Float32() > 0.7 {
					ctx.Emit("randomActivity", nil)
				}
			})

			// Metrics update
			ctx.On("metricsUpdate", func(_ interface{}) {
				// This would normally fetch real metrics from the monitoring system
				// For demo purposes, we're simulating it
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			totalComposables := ctx.Get("totalComposables").(*bubbly.Ref[int])
			cacheHits := ctx.Get("cacheHits").(*bubbly.Ref[int])
			cacheMisses := ctx.Get("cacheMisses").(*bubbly.Ref[int])
			activityLog := ctx.Get("activityLog").(*bubbly.Ref[[]string])
			showMetrics := ctx.Get("showMetrics").(*bubbly.Ref[bool])

			total := totalComposables.GetTyped()
			hits := cacheHits.GetTyped()
			misses := cacheMisses.GetTyped()
			log := activityLog.GetTyped()
			show := showMetrics.GetTyped()

			// Calculate hit rate
			hitRate := 0.0
			if hits+misses > 0 {
				hitRate = float64(hits) / float64(hits+misses) * 100
			}

			// Metrics panels
			var panels []string

			if show {
				// Composables metric
				composablesStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("63")).
					Padding(1, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("99")).
					Width(25)

				composablesPanel := composablesStyle.Render(
					fmt.Sprintf("Total Composables\n\n%d", total),
				)

				// Cache hit rate metric
				hitRateColor := "170" // Green
				if hitRate < 80 {
					hitRateColor = "214" // Orange
				}
				if hitRate < 50 {
					hitRateColor = "196" // Red
				}

				hitRateStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color(hitRateColor)).
					Padding(1, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("141")).
					Width(25)

				hitRatePanel := hitRateStyle.Render(
					fmt.Sprintf("Cache Hit Rate\n\n%.1f%%", hitRate),
				)

				// Cache stats metric
				cacheStatsStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("33")).
					Padding(1, 2).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("39")).
					Width(25)

				cacheStatsPanel := cacheStatsStyle.Render(
					fmt.Sprintf("Cache Stats\n\nHits: %d\nMisses: %d", hits, misses),
				)

				panels = []string{
					lipgloss.JoinHorizontal(lipgloss.Top, composablesPanel, "  ", hitRatePanel, "  ", cacheStatsPanel),
				}
			}

			// Activity log
			activityStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(80).
				Height(12)

			activityTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Render("Activity Log")

			activityContent := ""
			if len(log) == 0 {
				activityContent = "No activity yet. Press 'r' for random activity or 'c' to create composables."
			} else {
				for _, entry := range log {
					activityContent += entry + "\n"
				}
			}

			activityPanel := activityStyle.Render(activityTitle + "\n\n" + activityContent)

			// Info box
			infoStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(80)

			infoPanel := infoStyle.Render(
				"This dashboard demonstrates real-time metrics collection and monitoring.\n" +
					"Metrics are collected automatically and exposed via Prometheus at :9090/metrics\n\n" +
					"Phase 8 Features Demonstrated:\n" +
					"• Metrics Collection (Task 8.3)\n" +
					"• Prometheus Integration (Task 8.4)\n" +
					"• Real-time Monitoring\n" +
					"• Cache Performance Tracking",
			)

			if show {
				panels = append(panels, "", activityPanel, "", infoPanel)
			} else {
				panels = []string{activityPanel, "", infoPanel}
			}

			return lipgloss.JoinVertical(lipgloss.Left, panels...)
		}).
		Build()
}

func main() {
	// Initialize Prometheus metrics
	reg := prometheus.NewRegistry()
	metrics := monitoring.NewPrometheusMetrics(reg)
	monitoring.SetGlobalMetrics(metrics)

	// Start Prometheus HTTP server
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		fmt.Println("📊 Prometheus metrics available at http://localhost:9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			fmt.Printf("Error starting metrics server: %v\n", err)
		}
	}()

	// Create dashboard component
	component, err := createMonitoringDashboard()
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
