package devtools

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// TabItem represents a single tab in the tab controller.
//
// Each tab has a name and a content function that generates
// the content for that tab.
type TabItem struct {
	Name    string
	Content func() string
}

// TabController manages tab navigation and rendering in dev tools.
//
// The tab controller supports:
// - Multiple tabs with custom content
// - Keyboard navigation (Next, Prev, Select)
// - Thread-safe concurrent access
// - Visual highlighting of active tab
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	tabs := []TabItem{
//	    {Name: "Inspector", Content: func() string { return "Inspector content" }},
//	    {Name: "State", Content: func() string { return "State content" }},
//	    {Name: "Events", Content: func() string { return "Events content" }},
//	}
//	tc := devtools.NewTabController(tabs)
//	tc.Next() // Switch to next tab
//	output := tc.Render()
type TabController struct {
	mu        sync.RWMutex
	tabs      []TabItem
	activeTab int
}

// NewTabController creates a new tab controller with the given tabs.
//
// The first tab (index 0) is active by default.
// If no tabs are provided, the controller will render a "No tabs" message.
func NewTabController(tabs []TabItem) *TabController {
	return &TabController{
		tabs:      tabs,
		activeTab: 0,
	}
}

// Next switches to the next tab, wrapping around to the first tab
// if currently on the last tab.
//
// If there are no tabs or only one tab, this is a no-op.
func (tc *TabController) Next() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if len(tc.tabs) <= 1 {
		return
	}

	tc.activeTab = (tc.activeTab + 1) % len(tc.tabs)
}

// Prev switches to the previous tab, wrapping around to the last tab
// if currently on the first tab.
//
// If there are no tabs or only one tab, this is a no-op.
func (tc *TabController) Prev() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if len(tc.tabs) <= 1 {
		return
	}

	tc.activeTab = (tc.activeTab - 1 + len(tc.tabs)) % len(tc.tabs)
}

// Select switches to the tab at the given index.
//
// If the index is out of bounds, the active tab remains unchanged.
func (tc *TabController) Select(index int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if index >= 0 && index < len(tc.tabs) {
		tc.activeTab = index
	}
}

// GetActiveTab returns the index of the currently active tab.
func (tc *TabController) GetActiveTab() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.activeTab
}

// Render generates the visual representation of the tab controller.
//
// The output includes:
// - Tab navigation bar with active tab highlighted
// - Active tab content
//
// Returns a styled message if no tabs are configured.
func (tc *TabController) Render() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if len(tc.tabs) == 0 {
		return tc.renderNoTabs()
	}

	sections := []string{
		tc.renderTabBar(),
		tc.renderActiveContent(),
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderNoTabs returns a message when no tabs are configured.
func (tc *TabController) renderNoTabs() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(1, 2)
	return style.Render("No tabs configured")
}

// renderTabBar renders the tab navigation bar.
func (tc *TabController) renderTabBar() string {
	var tabStrings []string

	for i, tab := range tc.tabs {
		var style lipgloss.Style

		if i == tc.activeTab {
			// Active tab style - purple with bottom border
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true).
				Padding(0, 2).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(lipgloss.Color("99"))
		} else {
			// Inactive tab style - muted grey
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 2)
		}

		tabStrings = append(tabStrings, style.Render(tab.Name))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabStrings...)
}

// renderActiveContent renders the content of the active tab.
func (tc *TabController) renderActiveContent() string {
	if tc.activeTab < 0 || tc.activeTab >= len(tc.tabs) {
		return ""
	}

	tab := tc.tabs[tc.activeTab]
	content := tab.Content()

	contentStyle := lipgloss.NewStyle().
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color("240"))

	return contentStyle.Render(content)
}
