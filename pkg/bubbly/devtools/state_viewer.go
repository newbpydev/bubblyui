package devtools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// StateViewer displays all reactive state in the application.
//
// It provides a view of all refs across all components with filtering,
// selection, and value editing capabilities. The viewer integrates with
// the DevToolsStore to access component snapshots and their refs.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	store := NewDevToolsStore(1000, 5000, 1000)
//	viewer := NewStateViewer(store)
//	viewer.SetFilter("count")
//	viewer.SelectRef("ref-1")
//	output := viewer.Render()
type StateViewer struct {
	// store is the dev tools data store
	store *DevToolsStore

	// selected is the currently selected ref
	selected *RefSnapshot

	// filter is the current filter string (case-insensitive)
	filter string

	// mu protects concurrent access to selected and filter
	mu sync.RWMutex
}

// NewStateViewer creates a new state viewer for the given store.
//
// The viewer starts with no selection and no filter applied.
//
// Example:
//
//	store := NewDevToolsStore(1000, 5000, 1000)
//	viewer := NewStateViewer(store)
//
// Parameters:
//   - store: The dev tools store to read state from
//
// Returns:
//   - *StateViewer: A new state viewer instance
func NewStateViewer(store *DevToolsStore) *StateViewer {
	return &StateViewer{
		store:    store,
		selected: nil,
		filter:   "",
	}
}

// Render generates the visual output of the state viewer.
//
// The output includes:
//   - A header with the title "Reactive State"
//   - Component sections with their refs
//   - Selection indicator (►) for the selected ref
//   - Filtered refs based on the current filter
//   - Styled with Lipgloss for terminal display
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - string: The rendered output
func (sv *StateViewer) Render() string {
	sv.mu.RLock()
	selectedID := ""
	if sv.selected != nil {
		selectedID = sv.selected.ID
	}
	filter := sv.filter
	sv.mu.RUnlock()

	// Get all components from store
	components := sv.store.GetAllComponents()

	if len(components) == 0 {
		return sv.renderEmpty()
	}

	// Build output
	var lines []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")). // Purple
		Padding(0, 1)
	lines = append(lines, headerStyle.Render("Reactive State:"))
	lines = append(lines, "")

	// Render each component
	for _, comp := range components {
		if len(comp.Refs) == 0 {
			// Component with no refs
			lines = append(lines, fmt.Sprintf("┌─ %s", comp.Name))
			noRefsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)
			lines = append(lines, fmt.Sprintf("│  %s", noRefsStyle.Render("(no refs)")))
			lines = append(lines, "└─")
			lines = append(lines, "")
			continue
		}

		// Component header
		compStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("35")) // Green
		lines = append(lines, fmt.Sprintf("┌─ %s", compStyle.Render(comp.Name)))

		// Render refs
		for _, ref := range comp.Refs {
			// Apply filter
			if filter != "" && !sv.matchesFilter(ref.Name, filter) {
				continue
			}

			// Selection indicator
			indicator := " "
			refStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")) // Light grey
			if ref.ID == selectedID {
				indicator = selectionIndicator
				refStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("99")) // Purple
			}

			// Format value
			valueStr := sv.formatValue(ref.Value)

			// Type info
			typeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")) // Dark grey
			typeInfo := typeStyle.Render(fmt.Sprintf("(%s)", ref.Type))

			// Watchers info
			watcherInfo := ""
			if ref.Watchers > 0 {
				watcherStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")) // Orange
				watcherInfo = watcherStyle.Render(fmt.Sprintf(" [%d watchers]", ref.Watchers))
			}

			line := fmt.Sprintf("│ %s %s: %s %s%s",
				indicator,
				refStyle.Render(ref.Name),
				valueStr,
				typeInfo,
				watcherInfo)

			lines = append(lines, line)
		}

		lines = append(lines, "└─")
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderEmpty renders the empty state message.
func (sv *StateViewer) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Dark grey
		Italic(true).
		Padding(1, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")). // Purple
		Padding(0, 1)

	return headerStyle.Render("Reactive State:") + "\n\n" +
		emptyStyle.Render("No components with reactive state")
}

// formatValue formats a value for display.
func (sv *StateViewer) formatValue(value interface{}) string {
	if value == nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("<nil>")
	}

	// Format based on type
	str := fmt.Sprintf("%v", value)

	// Truncate long values
	maxLen := 50
	if len(str) > maxLen {
		str = str[:maxLen] + "..."
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")). // Yellow
		Render(str)
}

// matchesFilter checks if a ref name matches the filter (case-insensitive).
func (sv *StateViewer) matchesFilter(name, filter string) bool {
	return strings.Contains(strings.ToLower(name), strings.ToLower(filter))
}

// SelectRef selects a ref by ID.
//
// If the ref exists in any component, it is selected and the method returns true.
// If the ref does not exist, the selection is cleared and the method returns false.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	if viewer.SelectRef("ref-1") {
//	    fmt.Println("Ref selected")
//	}
//
// Parameters:
//   - id: The ref ID to select
//
// Returns:
//   - bool: True if the ref was found and selected, false otherwise
func (sv *StateViewer) SelectRef(id string) bool {
	// Find the ref in all components
	components := sv.store.GetAllComponents()
	for _, comp := range components {
		for _, ref := range comp.Refs {
			if ref.ID == id {
				sv.mu.Lock()
				sv.selected = ref
				sv.mu.Unlock()
				return true
			}
		}
	}

	// Ref not found, clear selection
	sv.mu.Lock()
	sv.selected = nil
	sv.mu.Unlock()
	return false
}

// GetSelected returns the currently selected ref.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - *RefSnapshot: The selected ref, or nil if no ref is selected
func (sv *StateViewer) GetSelected() *RefSnapshot {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.selected
}

// ClearSelection clears the current ref selection.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (sv *StateViewer) ClearSelection() {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.selected = nil
}

// SetFilter sets the filter string for ref names.
//
// The filter is case-insensitive and matches substrings.
// An empty string clears the filter.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	viewer.SetFilter("count") // Show only refs with "count" in name
//	viewer.SetFilter("")      // Show all refs
//
// Parameters:
//   - filter: The filter string
func (sv *StateViewer) SetFilter(filter string) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.filter = filter
}

// GetFilter returns the current filter string.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - string: The current filter
func (sv *StateViewer) GetFilter() string {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.filter
}

// EditValue edits the value of a ref.
//
// This method updates the ref value in the store. The ref must exist
// in one of the components, otherwise an error is returned.
//
// Note: This is a development tool feature. In a real application,
// editing values directly may have side effects and should be used
// with caution.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	err := viewer.EditValue("ref-1", 100)
//	if err != nil {
//	    fmt.Printf("Failed to edit: %v\n", err)
//	}
//
// Parameters:
//   - id: The ref ID to edit
//   - value: The new value
//
// Returns:
//   - error: An error if the ref was not found
func (sv *StateViewer) EditValue(id string, value interface{}) error {
	// Find the ref in all components
	components := sv.store.GetAllComponents()
	for _, comp := range components {
		for _, ref := range comp.Refs {
			if ref.ID == id {
				// Update the value
				ref.Value = value

				// Update in store (re-add the component)
				sv.store.AddComponent(comp)

				// Update selected ref if it's the one being edited
				sv.mu.Lock()
				if sv.selected != nil && sv.selected.ID == id {
					sv.selected.Value = value
				}
				sv.mu.Unlock()

				return nil
			}
		}
	}

	return fmt.Errorf("ref not found: %s", id)
}
