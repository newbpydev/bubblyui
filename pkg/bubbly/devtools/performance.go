package devtools

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// SortBy defines how to sort performance data
type SortBy int

const (
	// SortByRenderCount sorts by number of renders (descending)
	SortByRenderCount SortBy = iota
	// SortByAvgTime sorts by average render time (descending)
	SortByAvgTime
	// SortByMaxTime sorts by maximum render time (descending)
	SortByMaxTime
)

// String returns the string representation of SortBy
func (s SortBy) String() string {
	switch s {
	case SortByRenderCount:
		return "RenderCount"
	case SortByAvgTime:
		return "AvgTime"
	case SortByMaxTime:
		return "MaxTime"
	default:
		return "Unknown"
	}
}

// PerformanceMonitor tracks and displays component performance metrics.
//
// It provides methods to record render timing, sort components by various
// metrics, and render a formatted table of performance data. The monitor
// integrates with PerformanceData from the Store.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	data := NewPerformanceData()
//	monitor := NewPerformanceMonitor(data)
//	monitor.RecordRender("comp-1", "Counter", 5*time.Millisecond)
//	output := monitor.Render(SortByAvgTime)
type PerformanceMonitor struct {
	// data is the performance data store
	data *PerformanceData

	// sortBy is the current sort order
	sortBy SortBy

	// mu protects concurrent access to sortBy
	mu sync.RWMutex
}

// NewPerformanceMonitor creates a new performance monitor.
//
// The monitor starts with SortByRenderCount as the default sort order.
//
// Example:
//
//	data := NewPerformanceData()
//	monitor := NewPerformanceMonitor(data)
//
// Parameters:
//   - data: The performance data store to use
//
// Returns:
//   - *PerformanceMonitor: A new performance monitor instance
func NewPerformanceMonitor(data *PerformanceData) *PerformanceMonitor {
	return &PerformanceMonitor{
		data:   data,
		sortBy: SortByRenderCount,
	}
}

// RecordRender records a component render with its duration.
//
// This is a convenience method that forwards to PerformanceData.RecordRender.
// It maintains < 2% overhead as required by the specifications.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - componentID: The component ID
//   - componentName: The component name
//   - duration: How long the render took
func (pm *PerformanceMonitor) RecordRender(componentID, componentName string, duration time.Duration) {
	pm.data.RecordRender(componentID, componentName, duration)
}

// GetSortedComponents returns all components sorted by the specified order.
//
// The sorting is performed on a copy of the data, so the original data
// is not modified. Components are sorted in descending order (highest first).
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - sortBy: How to sort the components
//
// Returns:
//   - []*ComponentPerformance: Sorted component performance metrics
func (pm *PerformanceMonitor) GetSortedComponents(sortBy SortBy) []*ComponentPerformance {
	// Get all components
	allComponents := pm.data.GetAll()

	// Convert map to slice
	components := make([]*ComponentPerformance, 0, len(allComponents))
	for _, comp := range allComponents {
		components = append(components, comp)
	}

	// Sort based on criteria
	sort.Slice(components, func(i, j int) bool {
		switch sortBy {
		case SortByRenderCount:
			return components[i].RenderCount > components[j].RenderCount
		case SortByAvgTime:
			return components[i].AvgRenderTime > components[j].AvgRenderTime
		case SortByMaxTime:
			return components[i].MaxRenderTime > components[j].MaxRenderTime
		default:
			return components[i].RenderCount > components[j].RenderCount
		}
	})

	return components
}

// SetSortBy sets the default sort order for rendering.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - sortBy: The sort order to use
func (pm *PerformanceMonitor) SetSortBy(sortBy SortBy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.sortBy = sortBy
}

// GetSortBy returns the current sort order.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - SortBy: The current sort order
func (pm *PerformanceMonitor) GetSortBy() SortBy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.sortBy
}

// Render generates a formatted table of performance metrics.
//
// The output includes:
//   - A header with the title "Component Performance"
//   - A table with columns: Component, Renders, Avg Time, Max Time
//   - Components sorted by the specified order
//   - Styled with Lipgloss for terminal display
//   - Empty message if no performance data exists
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - sortBy: How to sort the components in the output
//
// Returns:
//   - string: The rendered output
func (pm *PerformanceMonitor) Render(sortBy SortBy) string {
	components := pm.GetSortedComponents(sortBy)

	if len(components) == 0 {
		return pm.renderEmpty()
	}

	// Define colors
	purple := lipgloss.Color("99")
	gray := lipgloss.Color("245")
	lightGray := lipgloss.Color("241")

	// Header style
	headerStyle := lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		Align(lipgloss.Center)

	// Cell styles
	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	oddRowStyle := cellStyle.Foreground(gray)
	evenRowStyle := cellStyle.Foreground(lightGray)

	// Build table rows
	rows := make([][]string, 0, len(components))
	for _, comp := range components {
		row := []string{
			truncate(comp.ComponentName, 18),
			fmt.Sprintf("%d", comp.RenderCount),
			formatDuration(comp.AvgRenderTime),
			formatDuration(comp.MaxRenderTime),
		}
		rows = append(rows, row)
	}

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers("Component", "Renders", "Avg Time", "Max Time").
		Rows(rows...)

	// Build output with title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(purple).
		Padding(0, 1)

	var output strings.Builder
	output.WriteString(titleStyle.Render("Component Performance"))
	output.WriteString("\n\n")
	output.WriteString(t.String())

	return output.String()
}

// renderEmpty returns a styled message for empty performance data.
func (pm *PerformanceMonitor) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(1, 2)

	return emptyStyle.Render("No performance data available.")
}

// truncate truncates a string to the specified length.
//
// If the string is longer than maxLen, it is truncated and "..." is appended.
//
// Parameters:
//   - s: The string to truncate
//   - maxLen: Maximum length
//
// Returns:
//   - string: The truncated string
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats a duration for display.
//
// Durations are formatted as:
//   - Microseconds (µs) for durations < 1ms
//   - Milliseconds (ms) for durations < 1s
//   - Seconds (s) for durations >= 1s
//
// Parameters:
//   - d: The duration to format
//
// Returns:
//   - string: The formatted duration
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
