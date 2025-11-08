package devtools

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// FlameGraphRenderer renders performance data as an ASCII flame graph.
//
// Flame graphs visualize hierarchical performance data, showing which components
// consume the most time. Each bar's width is proportional to the time spent.
//
// Example output:
//
//	Application ████████████████████████████████ 100% (50ms)
//	├─ Counter ████████████ 40% (20ms)
//	├─ Header ████████ 30% (15ms)
//	└─ Footer ████ 20% (10ms)
//
// Thread Safety:
//
//	FlameGraphRenderer is stateless and safe to use concurrently.
type FlameGraphRenderer struct {
	// width is the total width for rendering bars
	width int

	// height is the maximum depth to render (currently unused, for future expansion)
	height int
}

// FlameNode represents a node in the flame graph tree.
//
// Each node has a name, time duration, percentage of total time, and children.
type FlameNode struct {
	// Name is the component or section name
	Name string

	// Time is the duration spent in this node
	Time time.Duration

	// TimePercentage is the percentage of parent's time (0-100)
	TimePercentage float64

	// Children are the child nodes
	Children []*FlameNode
}

// NewFlameGraphRenderer creates a new flame graph renderer.
//
// Parameters:
//   - width: Total width for rendering bars (typically 80-120)
//   - height: Maximum depth to render (reserved for future use)
//
// Returns:
//   - *FlameGraphRenderer: A new renderer instance
//
// Example:
//
//	renderer := NewFlameGraphRenderer(80, 10)
//	output := renderer.Render(performanceData)
func NewFlameGraphRenderer(width, height int) *FlameGraphRenderer {
	return &FlameGraphRenderer{
		width:  width,
		height: height,
	}
}

// Render generates an ASCII flame graph from performance data.
//
// The output shows a hierarchical view of component performance with:
//   - Root node showing total application time
//   - Child nodes for each component
//   - Bars proportional to time spent
//   - Colors indicating performance (green=fast, yellow=medium, red=slow)
//   - Percentages and absolute times
//
// Parameters:
//   - data: The performance data to visualize
//
// Returns:
//   - string: The rendered flame graph
//
// Example:
//
//	data := NewPerformanceData()
//	data.RecordRender("comp-1", "Counter", 20*time.Millisecond)
//	renderer := NewFlameGraphRenderer(80, 10)
//	fmt.Println(renderer.Render(data))
func (fgr *FlameGraphRenderer) Render(data *PerformanceData) string {
	// Build flame tree from performance data
	root := fgr.buildFlameTree(data)

	// Handle empty data
	if root.Time == 0 {
		return fgr.renderEmpty()
	}

	// Render the tree
	var output strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Padding(0, 1)

	output.WriteString(titleStyle.Render("Flame Graph"))
	output.WriteString("\n\n")

	// Render root node
	output.WriteString(fgr.renderNode(root, 0))

	// Render children
	for i, child := range root.Children {
		isLast := i == len(root.Children)-1
		output.WriteString(fgr.renderNode(child, 1, isLast))
	}

	return output.String()
}

// buildFlameTree constructs a flame graph tree from performance data.
//
// The tree has a root node representing the total application time,
// with children for each component. Percentages are calculated relative
// to the total time.
//
// Parameters:
//   - data: The performance data
//
// Returns:
//   - *FlameNode: The root node of the flame tree
func (fgr *FlameGraphRenderer) buildFlameTree(data *PerformanceData) *FlameNode {
	// Get all components
	components := data.GetAll()

	// Calculate total time
	var totalTime time.Duration
	for _, comp := range components {
		totalTime += comp.TotalRenderTime
	}

	// Create root node
	root := &FlameNode{
		Name:           "Application",
		Time:           totalTime,
		TimePercentage: 100.0,
		Children:       make([]*FlameNode, 0, len(components)),
	}

	// Add children (sorted by time, descending)
	type componentTime struct {
		name string
		time time.Duration
	}

	compTimes := make([]componentTime, 0, len(components))
	for _, comp := range components {
		compTimes = append(compTimes, componentTime{
			name: comp.ComponentName,
			time: comp.TotalRenderTime,
		})
	}

	// Sort by time (descending)
	sort.Slice(compTimes, func(i, j int) bool {
		return compTimes[i].time > compTimes[j].time
	})

	// Create child nodes
	for _, ct := range compTimes {
		percentage := 0.0
		if totalTime > 0 {
			percentage = float64(ct.time) / float64(totalTime) * 100.0
		}

		child := &FlameNode{
			Name:           ct.name,
			Time:           ct.time,
			TimePercentage: percentage,
			Children:       nil,
		}
		root.Children = append(root.Children, child)
	}

	return root
}

// renderNode renders a single flame graph node.
//
// Parameters:
//   - node: The node to render
//   - depth: The depth in the tree (0 for root)
//   - isLast: Whether this is the last child (for tree connectors)
//
// Returns:
//   - string: The rendered node
func (fgr *FlameGraphRenderer) renderNode(node *FlameNode, depth int, isLast ...bool) string {
	var output strings.Builder

	// Determine tree connector
	connector := ""
	if depth > 0 {
		if len(isLast) > 0 && isLast[0] {
			connector = "└─ "
		} else {
			connector = "├─ "
		}
	}

	// Calculate bar width (leave space for label and percentage)
	// Format: "connector Name bar percentage (time)"
	// Reserve space: connector(3) + name(15) + percentage(8) + time(10) = 36 chars
	reservedSpace := 36
	if depth == 0 {
		reservedSpace = 26 // Root has no connector
	}

	availableWidth := fgr.width - reservedSpace
	if availableWidth < 10 {
		availableWidth = 10 // Minimum bar width
	}

	barWidth := fgr.calculateBarWidth(node.TimePercentage, availableWidth)

	// Truncate name to fit
	maxNameLen := 15
	name := fgr.truncateLabel(node.Name, maxNameLen)

	// Format bar
	bar := fgr.formatBar(barWidth)

	// Get color based on time
	color := fgr.getColorForTime(node.Time)
	barStyle := lipgloss.NewStyle().Foreground(color)

	// Format percentage and time
	percentageStr := fmt.Sprintf("%5.1f%%", node.TimePercentage)
	timeStr := formatDuration(node.Time)

	// Build line
	nameStyle := lipgloss.NewStyle().Bold(depth == 0)

	output.WriteString(connector)
	output.WriteString(nameStyle.Render(fmt.Sprintf("%-15s", name)))
	output.WriteString(" ")
	output.WriteString(barStyle.Render(bar))
	output.WriteString(" ")
	output.WriteString(percentageStr)
	output.WriteString(" (")
	output.WriteString(timeStr)
	output.WriteString(")")
	output.WriteString("\n")

	return output.String()
}

// calculateBarWidth calculates the width of a bar based on percentage.
//
// Parameters:
//   - percentage: The percentage (0-100)
//   - maxWidth: The maximum width available
//
// Returns:
//   - int: The bar width (minimum 1 if percentage > 0)
func (fgr *FlameGraphRenderer) calculateBarWidth(percentage float64, maxWidth int) int {
	width := int(percentage / 100.0 * float64(maxWidth))
	if width < 1 && percentage > 0 {
		width = 1 // Minimum width for visibility
	}
	if width > maxWidth {
		width = maxWidth
	}
	return width
}

// formatBar creates a bar string of the specified width.
//
// Parameters:
//   - width: The width in characters
//
// Returns:
//   - string: A string of █ characters
func (fgr *FlameGraphRenderer) formatBar(width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat("█", width)
}

// truncateLabel truncates a label to the specified maximum length.
//
// If the label is longer than maxLen, it is truncated and "..." is appended.
//
// Parameters:
//   - label: The label to truncate
//   - maxLen: Maximum length
//
// Returns:
//   - string: The truncated label
func (fgr *FlameGraphRenderer) truncateLabel(label string, maxLen int) string {
	if len(label) <= maxLen {
		return label
	}
	if maxLen <= 3 {
		return label[:maxLen]
	}
	return label[:maxLen-3] + "..."
}

// getColorForTime returns a color based on the duration.
//
// Color scheme:
//   - Green (35): Fast (< 5ms)
//   - Yellow (229): Medium (5-10ms)
//   - Red (196): Slow (> 10ms)
//
// Parameters:
//   - duration: The time duration
//
// Returns:
//   - lipgloss.Color: The color for styling
func (fgr *FlameGraphRenderer) getColorForTime(duration time.Duration) lipgloss.Color {
	if duration < 5*time.Millisecond {
		return lipgloss.Color("35") // Green - fast
	}
	if duration < 10*time.Millisecond {
		return lipgloss.Color("229") // Yellow - medium
	}
	return lipgloss.Color("196") // Red - slow
}

// renderEmpty returns a styled message for empty performance data.
func (fgr *FlameGraphRenderer) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Padding(1, 2)

	return emptyStyle.Render("No performance data available for flame graph.")
}
