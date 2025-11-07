package devtools

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ComponentInspector provides a complete component inspection interface
// integrating tree view, detail panel, search, and filtering.
//
// The inspector supports:
// - Hierarchical component tree navigation
// - Detailed component inspection with tabs
// - Search functionality for finding components
// - Filtering by type and status
// - Keyboard-driven navigation
// - Live updates as component tree changes
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	inspector := devtools.NewComponentInspector(rootSnapshot)
//
//	// Handle Bubbletea messages
//	cmd := inspector.Update(msg)
//
//	// Render the inspector
//	output := inspector.View()
type ComponentInspector struct {
	mu         sync.RWMutex
	tree       *TreeView
	detail     *DetailPanel
	search     *SearchWidget
	filter     *ComponentFilter
	searchMode bool
}

// NewComponentInspector creates a new component inspector with the given root component.
//
// The inspector is initialized with:
// - TreeView for hierarchical display
// - DetailPanel for component details
// - SearchWidget for finding components
// - ComponentFilter for filtering
//
// The root component can be nil, in which case the inspector will display
// an empty state until a root is set via SetRoot.
func NewComponentInspector(root *ComponentSnapshot) *ComponentInspector {
	// Collect all components for search
	components := collectAllComponents(root)

	// Create tree view and select root by default
	tree := NewTreeView(root)
	if root != nil {
		tree.Select(root.ID)
	}

	return &ComponentInspector{
		tree:       tree,
		detail:     NewDetailPanel(root),
		search:     NewSearchWidget(components),
		filter:     NewComponentFilter(),
		searchMode: false,
	}
}

// Update handles Bubbletea messages and updates the inspector state.
//
// Supported keyboard controls:
// - Up/Down: Navigate tree
// - Enter: Toggle node expansion
// - Tab: Next detail panel tab
// - Shift+Tab: Previous detail panel tab
// - Ctrl+F: Enter search mode
// - Esc: Exit search mode
//
// Returns a tea.Cmd if any async operations are needed, nil otherwise.
func (ci *ComponentInspector) Update(msg tea.Msg) tea.Cmd {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return ci.handleKeyMsg(msg)
	}

	return nil
}

// handleKeyMsg processes keyboard input (must be called with lock held).
func (ci *ComponentInspector) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	// Handle search mode separately
	if ci.searchMode {
		return ci.handleSearchMode(msg)
	}

	// Navigation mode
	switch msg.Type {
	case tea.KeyCtrlF:
		// Enter search mode
		ci.searchMode = true
		return nil

	case tea.KeyDown:
		// Select next component in tree
		ci.tree.SelectNext()
		ci.updateDetailPanel()
		return nil

	case tea.KeyUp:
		// Select previous component in tree
		ci.tree.SelectPrevious()
		ci.updateDetailPanel()
		return nil

	case tea.KeyEnter:
		// Toggle expansion of selected node
		selected := ci.tree.GetSelected()
		if selected != nil {
			ci.tree.Toggle(selected.ID)
		} else if ci.tree.GetRoot() != nil {
			// If nothing selected, select and expand root
			ci.tree.Select(ci.tree.GetRoot().ID)
			ci.tree.Toggle(ci.tree.GetRoot().ID)
			ci.updateDetailPanel()
		}
		return nil

	case tea.KeyTab:
		// Next detail panel tab
		ci.detail.NextTab()
		return nil

	case tea.KeyShiftTab:
		// Previous detail panel tab
		ci.detail.PreviousTab()
		return nil
	}

	return nil
}

// handleSearchMode processes keyboard input in search mode (must be called with lock held).
func (ci *ComponentInspector) handleSearchMode(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit search mode
		ci.searchMode = false
		ci.search.Clear()
		return nil

	case tea.KeyEnter:
		// Select current search result
		selected := ci.search.GetSelected()
		if selected != nil {
			ci.tree.Select(selected.ID)
			ci.updateDetailPanel()
		}
		ci.searchMode = false
		return nil

	case tea.KeyDown:
		// Next search result
		ci.search.NextResult()
		return nil

	case tea.KeyUp:
		// Previous search result
		ci.search.PrevResult()
		return nil

	case tea.KeyRunes:
		// Add characters to search query
		currentQuery := ci.search.GetQuery()
		newQuery := currentQuery + string(msg.Runes)
		ci.search.Search(newQuery)
		return nil

	case tea.KeyBackspace:
		// Remove last character from search query
		currentQuery := ci.search.GetQuery()
		if len(currentQuery) > 0 {
			newQuery := currentQuery[:len(currentQuery)-1]
			ci.search.Search(newQuery)
		}
		return nil
	}

	return nil
}

// updateDetailPanel updates the detail panel with the currently selected component.
// Must be called with lock held.
func (ci *ComponentInspector) updateDetailPanel() {
	selected := ci.tree.GetSelected()
	ci.detail.SetComponent(selected)
}

// View renders the complete inspector interface.
//
// The output includes:
// - Component tree view (left side)
// - Detail panel (right side)
// - Search widget (when in search mode)
//
// Layout is responsive and uses Lipgloss for styling.
func (ci *ComponentInspector) View() string {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	if ci.searchMode {
		return ci.renderSearchMode()
	}

	return ci.renderNormalMode()
}

// renderNormalMode renders the inspector in normal navigation mode.
// Must be called with read lock held.
func (ci *ComponentInspector) renderNormalMode() string {
	// Render tree view
	treeView := ci.tree.Render()
	treeStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(40)
	treeBox := treeStyle.Render(treeView)

	// Render detail panel
	detailView := ci.detail.Render()
	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(60)
	detailBox := detailStyle.Render(detailView)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, treeBox, detailBox)
}

// renderSearchMode renders the inspector in search mode.
// Must be called with read lock held.
func (ci *ComponentInspector) renderSearchMode() string {
	// Render search widget
	searchView := ci.search.Render()
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("35")).
		Padding(1, 2).
		Width(100)

	return searchStyle.Render(searchView)
}

// SetRoot updates the root component and refreshes all sub-components.
//
// This method should be called when the component tree changes to keep
// the inspector in sync with the application state.
func (ci *ComponentInspector) SetRoot(root *ComponentSnapshot) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	// Update tree view
	ci.tree = NewTreeView(root)

	// Update search widget with new components
	components := collectAllComponents(root)
	ci.search.SetComponents(components)

	// Clear detail panel if current selection is no longer valid
	ci.detail.SetComponent(nil)
}

// ApplyFilter applies the current filter to the search results.
//
// This method updates the search widget to show only components that
// pass the filter criteria.
func (ci *ComponentInspector) ApplyFilter() {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	// Get all components
	root := ci.tree.GetRoot()
	allComponents := collectAllComponents(root)

	// Apply filter
	filtered := ci.filter.Apply(allComponents)

	// Update search widget
	ci.search.SetComponents(filtered)
}

// collectAllComponents recursively collects all components in the tree.
func collectAllComponents(root *ComponentSnapshot) []*ComponentSnapshot {
	if root == nil {
		return []*ComponentSnapshot{}
	}

	components := []*ComponentSnapshot{root}

	for _, child := range root.Children {
		components = append(components, collectAllComponents(child)...)
	}

	return components
}
