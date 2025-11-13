package devtools

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// DetailPanel displays detailed information about a selected component
// with tabbed views for different aspects (State, Props, Events).
//
// The detail panel supports:
// - Multiple tabs for organizing component information
// - Tab switching with keyboard navigation
// - Thread-safe concurrent access
// - Graceful handling of nil components
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	dp := devtools.NewDetailPanel(componentSnapshot)
//	dp.SwitchTab(1) // Switch to Props tab
//	output := dp.Render()
type DetailPanel struct {
	mu        sync.RWMutex
	component *ComponentSnapshot
	tabs      []Tab
	activeTab int
}

// Tab represents a single tab in the detail panel.
//
// Each tab has a name and a render function that generates
// the content for that tab based on the current component.
type Tab struct {
	Name   string
	Render func(*ComponentSnapshot) string
}

// NewDetailPanel creates a new detail panel for the given component.
//
// The panel is initialized with three default tabs:
// - State: Shows reactive state (refs)
// - Props: Shows component properties
// - Events: Shows event history (placeholder)
//
// The component can be nil, in which case the panel will display
// a "No component selected" message.
func NewDetailPanel(component *ComponentSnapshot) *DetailPanel {
	dp := &DetailPanel{
		component: component,
		activeTab: 0,
	}

	// Initialize default tabs
	dp.tabs = []Tab{
		{
			Name:   "State",
			Render: renderStateTab,
		},
		{
			Name:   "Props",
			Render: renderPropsTab,
		},
		{
			Name:   "Events",
			Render: renderEventsTab,
		},
	}

	return dp
}

// Render generates the visual representation of the detail panel.
//
// The output includes:
// - Component header with name and type
// - Tab navigation bar
// - Active tab content
//
// Returns a styled message if no component is selected.
func (dp *DetailPanel) Render() string {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	if dp.component == nil {
		return dp.renderNoComponent()
	}

	sections := []string{
		dp.renderHeader(),
		dp.renderTabs(),
		dp.renderActiveTabContent(),
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderNoComponent returns a message when no component is selected.
func (dp *DetailPanel) renderNoComponent() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(1, 2)
	return style.Render("No component selected")
}

// renderHeader renders the component header with name and type.
func (dp *DetailPanel) renderHeader() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Padding(0, 1)

	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	name := dp.component.Name
	if dp.component.Type != "" {
		name = fmt.Sprintf("%s %s", name, typeStyle.Render(fmt.Sprintf("(%s)", dp.component.Type)))
	}

	return headerStyle.Render(name)
}

// renderTabs renders the tab navigation bar.
func (dp *DetailPanel) renderTabs() string {
	var tabStrings []string

	for i, tab := range dp.tabs {
		var style lipgloss.Style

		if i == dp.activeTab {
			// Active tab style - CLEAR background for visibility
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("99")).  // Purple background
				Foreground(lipgloss.Color("15")).  // White text
				Bold(true).
				Padding(0, 2)
		} else {
			// Inactive tab style - subtle
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).  // Dim gray
				Padding(0, 2)
		}

		tabStrings = append(tabStrings, style.Render(tab.Name))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabStrings...)
}

// renderActiveTabContent renders the content of the active tab.
func (dp *DetailPanel) renderActiveTabContent() string {
	if dp.activeTab < 0 || dp.activeTab >= len(dp.tabs) {
		return ""
	}

	tab := dp.tabs[dp.activeTab]
	content := tab.Render(dp.component)

	// Simple top border line separator (cleaner than box border)
	contentStyle := lipgloss.NewStyle().
		Padding(1, 2)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("─────────────────────────────────────────────")

	return separator + "\n" + contentStyle.Render(content)
}

// SwitchTab switches to the tab at the given index.
//
// If the index is out of bounds, the active tab remains unchanged.
func (dp *DetailPanel) SwitchTab(index int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if index >= 0 && index < len(dp.tabs) {
		dp.activeTab = index
	}
}

// NextTab switches to the next tab, wrapping around to the first tab
// if currently on the last tab.
func (dp *DetailPanel) NextTab() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.activeTab = (dp.activeTab + 1) % len(dp.tabs)
}

// PreviousTab switches to the previous tab, wrapping around to the last tab
// if currently on the first tab.
func (dp *DetailPanel) PreviousTab() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.activeTab = (dp.activeTab - 1 + len(dp.tabs)) % len(dp.tabs)
}

// GetActiveTab returns the index of the currently active tab.
func (dp *DetailPanel) GetActiveTab() int {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.activeTab
}

// SetComponent updates the component being displayed.
//
// The component can be nil, in which case the panel will display
// a "No component selected" message.
func (dp *DetailPanel) SetComponent(component *ComponentSnapshot) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.component = component
}

// GetComponent returns the currently displayed component.
func (dp *DetailPanel) GetComponent() *ComponentSnapshot {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.component
}

// renderStateTab renders the State tab content showing reactive refs.
func renderStateTab(component *ComponentSnapshot) string {
	if component == nil {
		return "No component"
	}

	if len(component.Refs) == 0 {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		return style.Render("No reactive state")
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("Reactive State:"))
	lines = append(lines, "")

	for _, ref := range component.Refs {
		nameStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Bold(true)

		typeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

		valueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))

		line := fmt.Sprintf("  %s: %s %s",
			nameStyle.Render(ref.Name),
			valueStyle.Render(fmt.Sprintf("%v", ref.Value)),
			typeStyle.Render(fmt.Sprintf("(%s)", ref.Type)))

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderPropsTab renders the Props tab content showing component properties.
func renderPropsTab(component *ComponentSnapshot) string {
	if component == nil {
		return "No component"
	}

	if len(component.Props) == 0 {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		return style.Render("No props")
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("Props:"))
	lines = append(lines, "")

	// Sort keys for consistent output
	keys := make([]string, 0, len(component.Props))
	for key := range component.Props {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := component.Props[key]

		keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Bold(true)

		valueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))

		line := fmt.Sprintf("  %s: %s",
			keyStyle.Render(key),
			valueStyle.Render(fmt.Sprintf("%v", value)))

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderEventsTab renders the Events tab content.
//
// This is currently a placeholder. Full event tracking will be
// implemented in later tasks when EventLog integration is added.
func renderEventsTab(component *ComponentSnapshot) string {
	if component == nil {
		return "No component"
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	return style.Render("Event tracking coming soon...")
}
