package devtools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// SearchWidget provides fuzzy search functionality for component snapshots.
//
// The search widget supports:
// - Case-insensitive substring matching on component names and types
// - Result navigation with keyboard controls
// - Thread-safe concurrent access
// - Performance optimized for large component trees
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	sw := devtools.NewSearchWidget(components)
//	sw.Search("button")
//	selected := sw.GetSelected()
//	output := sw.Render()
type SearchWidget struct {
	mu         sync.RWMutex
	components []*ComponentSnapshot
	query      string
	results    []*ComponentSnapshot
	cursor     int
}

// NewSearchWidget creates a new search widget with the given components.
//
// The components parameter can be nil or empty, in which case searches
// will return no results until components are set via SetComponents.
func NewSearchWidget(components []*ComponentSnapshot) *SearchWidget {
	return &SearchWidget{
		components: components,
		query:      "",
		results:    []*ComponentSnapshot{},
		cursor:     0,
	}
}

// Search performs a fuzzy search on component names and types.
//
// The search is case-insensitive and matches substrings. An empty query
// returns all components. The cursor is reset to 0 after each search.
//
// Matching criteria:
// - Component name contains query (case-insensitive)
// - Component type contains query (case-insensitive)
func (sw *SearchWidget) Search(query string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.query = query
	sw.cursor = 0
	sw.results = sw.performSearch(query)
}

// performSearch executes the search algorithm (must be called with lock held).
func (sw *SearchWidget) performSearch(query string) []*ComponentSnapshot {
	// Empty query returns all components
	if query == "" {
		return sw.components
	}

	queryLower := strings.ToLower(query)
	var results []*ComponentSnapshot

	for _, component := range sw.components {
		if sw.matchesQuery(component, queryLower) {
			results = append(results, component)
		}
	}

	return results
}

// matchesQuery checks if a component matches the search query.
func (sw *SearchWidget) matchesQuery(component *ComponentSnapshot, queryLower string) bool {
	if component == nil {
		return false
	}

	// Match against name
	nameLower := strings.ToLower(component.Name)
	if strings.Contains(nameLower, queryLower) {
		return true
	}

	// Match against type
	typeLower := strings.ToLower(component.Type)
	if strings.Contains(typeLower, queryLower) {
		return true
	}

	return false
}

// NextResult moves the cursor to the next result, wrapping around to
// the first result if at the end.
func (sw *SearchWidget) NextResult() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if len(sw.results) == 0 {
		sw.cursor = 0
		return
	}

	sw.cursor = (sw.cursor + 1) % len(sw.results)
}

// PrevResult moves the cursor to the previous result, wrapping around to
// the last result if at the beginning.
func (sw *SearchWidget) PrevResult() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if len(sw.results) == 0 {
		sw.cursor = 0
		return
	}

	sw.cursor = (sw.cursor - 1 + len(sw.results)) % len(sw.results)
}

// GetSelected returns the currently selected result, or nil if there
// are no results.
func (sw *SearchWidget) GetSelected() *ComponentSnapshot {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	if len(sw.results) == 0 {
		return nil
	}

	if sw.cursor < 0 || sw.cursor >= len(sw.results) {
		return nil
	}

	return sw.results[sw.cursor]
}

// Clear resets the search query, results, and cursor.
func (sw *SearchWidget) Clear() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.query = ""
	sw.results = []*ComponentSnapshot{}
	sw.cursor = 0
}

// SetComponents updates the component search space.
//
// This does not automatically re-run the current search. Call Search()
// again to search the new components.
func (sw *SearchWidget) SetComponents(components []*ComponentSnapshot) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.components = components
}

// GetQuery returns the current search query.
func (sw *SearchWidget) GetQuery() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.query
}

// GetResults returns a copy of the current search results.
func (sw *SearchWidget) GetResults() []*ComponentSnapshot {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	// Return a copy to prevent external modification
	results := make([]*ComponentSnapshot, len(sw.results))
	copy(results, sw.results)
	return results
}

// GetCursor returns the current cursor position.
func (sw *SearchWidget) GetCursor() int {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.cursor
}

// GetResultCount returns the number of search results.
func (sw *SearchWidget) GetResultCount() int {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return len(sw.results)
}

// Render generates the visual representation of the search widget.
//
// The output includes:
// - Search query input
// - Result count and current position
// - List of results with cursor indicator
// - Selected result highlighting
//
// Returns a styled message if there are no results.
func (sw *SearchWidget) Render() string {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	var sections []string

	// Render search input
	sections = append(sections, sw.renderSearchInput())

	// Render results
	if len(sw.results) == 0 {
		sections = append(sections, sw.renderNoResults())
	} else {
		sections = append(sections, sw.renderResults())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderSearchInput renders the search input field with query.
func (sw *SearchWidget) renderSearchInput() string {
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	queryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("35"))

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	label := labelStyle.Render("Search: ")
	query := queryStyle.Render(sw.query)

	var count string
	if len(sw.results) > 0 {
		count = countStyle.Render(fmt.Sprintf(" (%d/%d)", sw.cursor+1, len(sw.results)))
	} else if sw.query != "" {
		count = countStyle.Render(" (0 results)")
	}

	return label + query + count
}

// renderNoResults returns a message when no results are found.
func (sw *SearchWidget) renderNoResults() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(1, 0)

	return style.Render("No results found")
}

// renderResults renders the list of search results.
func (sw *SearchWidget) renderResults() string {
	var lines []string

	// Show up to 10 results at a time
	maxDisplay := 10
	start := 0
	end := len(sw.results)

	// If more than maxDisplay results, show window around cursor
	if len(sw.results) > maxDisplay {
		start = sw.cursor - maxDisplay/2
		if start < 0 {
			start = 0
		}
		end = start + maxDisplay
		if end > len(sw.results) {
			end = len(sw.results)
			start = end - maxDisplay
			if start < 0 {
				start = 0
			}
		}
	}

	// Add ellipsis if not showing all results
	if start > 0 {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("  ..."))
	}

	// Render visible results
	for i := start; i < end; i++ {
		result := sw.results[i]
		line := sw.renderResult(result, i == sw.cursor)
		lines = append(lines, line)
	}

	// Add ellipsis if not showing all results
	if end < len(sw.results) {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("  ..."))
	}

	return strings.Join(lines, "\n")
}

// renderResult renders a single search result.
func (sw *SearchWidget) renderResult(result *ComponentSnapshot, isSelected bool) string {
	var style lipgloss.Style
	prefix := "  "

	if isSelected {
		prefix = "► "
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)
	} else {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	}

	typeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	name := result.Name
	if result.Type != "" {
		name = fmt.Sprintf("%s %s", name, typeStyle.Render(fmt.Sprintf("(%s)", result.Type)))
	}

	return prefix + style.Render(name)
}
