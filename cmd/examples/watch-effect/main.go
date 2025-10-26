// Package main demonstrates WatchEffect with automatic dependency tracking.
// This example shows a real-time analytics dashboard that automatically
// reacts to data changes without manually specifying dependencies (Task 6.3).
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Model represents the application state
type model struct {
	// Reactive state
	visitors    *bubbly.Ref[int]
	pageViews   *bubbly.Ref[int]
	revenue     *bubbly.Ref[float64]
	errors      *bubbly.Ref[int]
	showDetails *bubbly.Ref[bool]

	// Computed values
	conversionRate *bubbly.Computed[float64]
	revenuePerUser *bubbly.Computed[float64]
	errorRate      *bubbly.Computed[float64]
	healthStatus   *bubbly.Computed[string]

	// UI state
	logs     []string
	quitting bool

	// Cleanup functions
	cleanups []bubbly.WatchCleanup
}

// keyMap defines keyboard shortcuts
type keyMap struct {
	Visitor  key.Binding
	PageView key.Binding
	Revenue  key.Binding
	Error    key.Binding
	Toggle   key.Binding
	Auto     key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Visitor: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "add visitor"),
	),
	PageView: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "add pageview"),
	),
	Revenue: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "add revenue"),
	),
	Error: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "add error"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle details"),
	),
	Auto: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "auto-generate"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	metricStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))

	goodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	badStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

func initialModel() model {
	// Create reactive state
	visitors := bubbly.NewRef(100)
	pageViews := bubbly.NewRef(250)
	revenue := bubbly.NewRef(1250.50)
	errors := bubbly.NewRef(5)
	showDetails := bubbly.NewRef(true)

	// Create computed values
	conversionRate := bubbly.NewComputed(func() float64 {
		v := visitors.Get()
		if v == 0 {
			return 0
		}
		return (float64(pageViews.Get()) / float64(v)) * 100
	})

	revenuePerUser := bubbly.NewComputed(func() float64 {
		v := visitors.Get()
		if v == 0 {
			return 0
		}
		return revenue.Get() / float64(v)
	})

	errorRate := bubbly.NewComputed(func() float64 {
		pv := pageViews.Get()
		if pv == 0 {
			return 0
		}
		return (float64(errors.Get()) / float64(pv)) * 100
	})

	healthStatus := bubbly.NewComputed(func() string {
		rate := errorRate.Get()
		if rate < 1.0 {
			return "Healthy"
		} else if rate < 5.0 {
			return "Warning"
		}
		return "Critical"
	})

	m := model{
		visitors:       visitors,
		pageViews:      pageViews,
		revenue:        revenue,
		errors:         errors,
		showDetails:    showDetails,
		conversionRate: conversionRate,
		revenuePerUser: revenuePerUser,
		errorRate:      errorRate,
		healthStatus:   healthStatus,
		logs:           make([]string, 0, 20),
	}

	// Task 6.3: Use WatchEffect for automatic dependency tracking!
	// These effects automatically discover and track their dependencies
	m.setupWatchEffects()

	return m
}

func (m *model) setupWatchEffects() {
	// Effect 1: Log conversion rate changes
	// Automatically tracks: visitors, pageViews (via conversionRate)
	cleanup1 := bubbly.WatchEffect(func() {
		rate := m.conversionRate.Get()
		m.addLog(fmt.Sprintf("📊 Conversion rate: %.2f%%", rate))
	})
	m.cleanups = append(m.cleanups, cleanup1)

	// Effect 2: Log revenue per user
	// Automatically tracks: visitors, revenue (via revenuePerUser)
	cleanup2 := bubbly.WatchEffect(func() {
		rpu := m.revenuePerUser.Get()
		m.addLog(fmt.Sprintf("💰 Revenue per user: $%.2f", rpu))
	})
	m.cleanups = append(m.cleanups, cleanup2)

	// Effect 3: Alert on error rate
	// Automatically tracks: errors, pageViews (via errorRate)
	cleanup3 := bubbly.WatchEffect(func() {
		rate := m.errorRate.Get()
		if rate > 5.0 {
			m.addLog(fmt.Sprintf("🚨 HIGH ERROR RATE: %.2f%%", rate))
		} else if rate > 1.0 {
			m.addLog(fmt.Sprintf("⚠️  Error rate elevated: %.2f%%", rate))
		}
	})
	m.cleanups = append(m.cleanups, cleanup3)

	// Effect 4: Conditional logging based on showDetails
	// Automatically tracks: showDetails, and conditionally tracks other values
	cleanup4 := bubbly.WatchEffect(func() {
		if m.showDetails.Get() {
			// When details are shown, track these values
			v := m.visitors.Get()
			pv := m.pageViews.Get()
			m.addLog(fmt.Sprintf("📈 Details: %d visitors, %d pageviews", v, pv))
		} else {
			// When details are hidden, only track showDetails
			m.addLog("🔒 Details hidden")
		}
	})
	m.cleanups = append(m.cleanups, cleanup4)

	// Effect 5: Health status monitoring
	// Automatically tracks: errorRate (via healthStatus)
	cleanup5 := bubbly.WatchEffect(func() {
		status := m.healthStatus.Get()
		switch status {
		case "Healthy":
			m.addLog("✅ System status: " + status)
		case "Warning":
			m.addLog("⚠️  System status: " + status)
		case "Critical":
			m.addLog("🔴 System status: " + status)
		}
	})
	m.cleanups = append(m.cleanups, cleanup5)
}

func (m *model) addLog(msg string) {
	m.logs = append(m.logs, msg)
	// Keep only last 15 logs
	if len(m.logs) > 15 {
		m.logs = m.logs[len(m.logs)-15:]
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

type autoGenMsg struct{}

func autoGenerate() tea.Msg {
	time.Sleep(500 * time.Millisecond)
	return autoGenMsg{}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			// Cleanup all watch effects
			for _, cleanup := range m.cleanups {
				cleanup()
			}
			return m, tea.Quit

		case key.Matches(msg, keys.Visitor):
			m.visitors.Set(m.visitors.Get() + rand.Intn(10) + 1)

		case key.Matches(msg, keys.PageView):
			m.pageViews.Set(m.pageViews.Get() + rand.Intn(20) + 1)

		case key.Matches(msg, keys.Revenue):
			m.revenue.Set(m.revenue.Get() + float64(rand.Intn(100)) + rand.Float64()*100)

		case key.Matches(msg, keys.Error):
			m.errors.Set(m.errors.Get() + rand.Intn(3) + 1)

		case key.Matches(msg, keys.Toggle):
			m.showDetails.Set(!m.showDetails.Get())

		case key.Matches(msg, keys.Auto):
			// Auto-generate some activity
			m.visitors.Set(m.visitors.Get() + rand.Intn(5) + 1)
			m.pageViews.Set(m.pageViews.Get() + rand.Intn(15) + 5)
			m.revenue.Set(m.revenue.Get() + float64(rand.Intn(50)) + rand.Float64()*50)
			if rand.Float64() < 0.3 {
				m.errors.Set(m.errors.Get() + 1)
			}
			return m, autoGenerate
		}

	case autoGenMsg:
		// Continue auto-generating
		m.visitors.Set(m.visitors.Get() + rand.Intn(3) + 1)
		m.pageViews.Set(m.pageViews.Get() + rand.Intn(10) + 3)
		m.revenue.Set(m.revenue.Get() + float64(rand.Intn(30)) + rand.Float64()*30)
		if rand.Float64() < 0.2 {
			m.errors.Set(m.errors.Get() + 1)
		}
		return m, autoGenerate
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Analytics dashboard closed. 👋\n"
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📊 Real-Time Analytics Dashboard - WatchEffect Demo"))
	b.WriteString("\n\n")

	// Metrics in a grid
	metrics := []string{
		metricStyle.Render(fmt.Sprintf("👥 Visitors\n%d", m.visitors.Get())),
		metricStyle.Render(fmt.Sprintf("📄 Page Views\n%d", m.pageViews.Get())),
		metricStyle.Render(fmt.Sprintf("💵 Revenue\n$%.2f", m.revenue.Get())),
		metricStyle.Render(fmt.Sprintf("❌ Errors\n%d", m.errors.Get())),
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, metrics...))
	b.WriteString("\n\n")

	// Computed metrics (automatically calculated)
	b.WriteString("Computed Metrics (Auto-tracked):\n")
	b.WriteString(fmt.Sprintf("  Conversion Rate: %.2f%%\n", m.conversionRate.Get()))
	b.WriteString(fmt.Sprintf("  Revenue/User: $%.2f\n", m.revenuePerUser.Get()))

	// Error rate with color coding
	rate := m.errorRate.Get()
	var rateStr string
	if rate < 1.0 {
		rateStr = goodStyle.Render(fmt.Sprintf("%.2f%%", rate))
	} else if rate < 5.0 {
		rateStr = warningStyle.Render(fmt.Sprintf("%.2f%%", rate))
	} else {
		rateStr = badStyle.Render(fmt.Sprintf("%.2f%%", rate))
	}
	b.WriteString(fmt.Sprintf("  Error Rate: %s\n", rateStr))

	// Health status with color coding
	status := m.healthStatus.Get()
	var statusStr string
	switch status {
	case "Healthy":
		statusStr = goodStyle.Render(status)
	case "Warning":
		statusStr = warningStyle.Render(status)
	default:
		statusStr = badStyle.Render(status)
	}
	b.WriteString(fmt.Sprintf("  Health: %s\n", statusStr))
	b.WriteString("\n")

	// WatchEffect activity log
	b.WriteString("WatchEffect Activity (Auto-triggered):\n")
	if len(m.logs) == 0 {
		b.WriteString(logStyle.Render("  (no activity yet)"))
		b.WriteString("\n")
	} else {
		for _, log := range m.logs {
			b.WriteString(logStyle.Render("  " + log))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Help
	help := fmt.Sprintf("%s • %s • %s • %s • %s • %s • %s",
		keys.Visitor.Help().Desc,
		keys.PageView.Help().Desc,
		keys.Revenue.Help().Desc,
		keys.Error.Help().Desc,
		keys.Toggle.Help().Desc,
		keys.Auto.Help().Desc,
		keys.Quit.Help().Desc,
	)
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
