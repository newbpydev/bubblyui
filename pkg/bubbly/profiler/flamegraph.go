// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Default dimensions for flame graph SVG output.
const (
	// DefaultFlameGraphWidth is the default width in pixels.
	DefaultFlameGraphWidth = 1200

	// DefaultFlameGraphHeight is the default height in pixels.
	DefaultFlameGraphHeight = 600

	// frameHeight is the height of each stack frame in pixels.
	frameHeight = 18

	// framePadding is the vertical padding between frames.
	framePadding = 1

	// minWidthForLabel is the minimum width to display a label.
	minWidthForLabel = 30

	// charWidth is the approximate width of a character in pixels.
	charWidth = 7
)

// FlameGraphGenerator generates flame graph visualizations from CPU profile data.
//
// Flame graphs are a visualization of hierarchical data, created to visualize
// stack traces of profiled software. Each box represents a function in the stack,
// with the width proportional to the time spent in that function.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	fgg := NewFlameGraphGenerator()
//	root := fgg.Generate(cpuProfile)
//	svg := fgg.GenerateSVG(cpuProfile)
type FlameGraphGenerator struct {
	// width is the SVG width in pixels
	width int

	// height is the SVG height in pixels
	height int

	// mu protects concurrent access to generator state
	mu sync.RWMutex
}

// CallNode represents a node in the call tree for flame graph generation.
//
// Each node represents a function in the call stack, with children representing
// functions called by this function.
//
// Example:
//
//	node := &CallNode{
//	    Name:     "main.render",
//	    Samples:  100,
//	    Percent:  50.0,
//	    Children: []*CallNode{...},
//	}
type CallNode struct {
	// Name is the function name
	Name string

	// Samples is the number of CPU samples in this function
	Samples int64

	// Percent is the percentage of total CPU time
	Percent float64

	// Children are functions called by this function
	Children []*CallNode
}

// NewFlameGraphGenerator creates a new FlameGraphGenerator with default dimensions.
//
// Default dimensions are 1200x600 pixels.
//
// Example:
//
//	fgg := NewFlameGraphGenerator()
//	svg := fgg.GenerateSVG(profile)
func NewFlameGraphGenerator() *FlameGraphGenerator {
	return &FlameGraphGenerator{
		width:  DefaultFlameGraphWidth,
		height: DefaultFlameGraphHeight,
	}
}

// NewFlameGraphGeneratorWithDimensions creates a FlameGraphGenerator with custom dimensions.
//
// If width or height is <= 0, the default value is used.
//
// Example:
//
//	fgg := NewFlameGraphGeneratorWithDimensions(1600, 800)
func NewFlameGraphGeneratorWithDimensions(width, height int) *FlameGraphGenerator {
	fgg := &FlameGraphGenerator{
		width:  DefaultFlameGraphWidth,
		height: DefaultFlameGraphHeight,
	}

	if width > 0 {
		fgg.width = width
	}
	if height > 0 {
		fgg.height = height
	}

	return fgg
}

// Generate builds a call tree from CPU profile data.
//
// Returns the root node of the call tree, or nil if the profile is empty.
// The root node represents the total CPU time with children being top-level functions.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	root := fgg.Generate(cpuProfile)
//	if root != nil {
//	    fmt.Printf("Total samples: %d\n", root.Samples)
//	}
func (fgg *FlameGraphGenerator) Generate(profile *CPUProfileData) *CallNode {
	if profile == nil || len(profile.HotFunctions) == 0 || profile.TotalSamples == 0 {
		return nil
	}

	// Create root node with total samples
	root := &CallNode{
		Name:     "root",
		Samples:  profile.TotalSamples,
		Percent:  100.0,
		Children: make([]*CallNode, 0),
	}

	// Build function lookup map
	funcMap := make(map[string]*HotFunction)
	for _, hf := range profile.HotFunctions {
		funcMap[hf.Name] = hf
	}

	// Find root functions (functions that are not called by others)
	calledFuncs := make(map[string]bool)
	for _, callees := range profile.CallGraph {
		for _, callee := range callees {
			calledFuncs[callee] = true
		}
	}

	// Add top-level functions as children of root
	for _, hf := range profile.HotFunctions {
		// If function is not called by anyone, it's a root function
		if !calledFuncs[hf.Name] {
			child := fgg.buildCallTree(hf.Name, funcMap, profile.CallGraph, make(map[string]bool))
			if child != nil {
				root.Children = append(root.Children, child)
			}
		}
	}

	// If no root functions found (circular calls), add all as children
	if len(root.Children) == 0 {
		for _, hf := range profile.HotFunctions {
			child := &CallNode{
				Name:     hf.Name,
				Samples:  hf.Samples,
				Percent:  hf.Percent,
				Children: make([]*CallNode, 0),
			}
			root.Children = append(root.Children, child)
		}
	}

	// Sort children by samples descending
	sort.Slice(root.Children, func(i, j int) bool {
		return root.Children[i].Samples > root.Children[j].Samples
	})

	return root
}

// buildCallTree recursively builds the call tree for a function.
func (fgg *FlameGraphGenerator) buildCallTree(funcName string, funcMap map[string]*HotFunction, callGraph map[string][]string, visited map[string]bool) *CallNode {
	// Prevent infinite recursion
	if visited[funcName] {
		return nil
	}
	visited[funcName] = true

	hf, ok := funcMap[funcName]
	if !ok {
		return nil
	}

	node := &CallNode{
		Name:     hf.Name,
		Samples:  hf.Samples,
		Percent:  hf.Percent,
		Children: make([]*CallNode, 0),
	}

	// Add children from call graph
	callees := callGraph[funcName]
	for _, callee := range callees {
		child := fgg.buildCallTree(callee, funcMap, callGraph, visited)
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	// Sort children by samples descending
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Samples > node.Children[j].Samples
	})

	return node
}

// GenerateSVG generates an SVG flame graph from CPU profile data.
//
// Returns an SVG string that can be embedded in HTML or saved to a file.
// If the profile is nil or empty, returns an SVG with a "No profile data" message.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	svg := fgg.GenerateSVG(cpuProfile)
//	os.WriteFile("flamegraph.svg", []byte(svg), 0644)
func (fgg *FlameGraphGenerator) GenerateSVG(profile *CPUProfileData) string {
	fgg.mu.RLock()
	width := fgg.width
	height := fgg.height
	fgg.mu.RUnlock()

	var svg strings.Builder

	// SVG header
	svg.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	svg.WriteString("\n")

	// Add styles
	svg.WriteString(`<style>
    .frame { stroke: #333; stroke-width: 0.5; }
    .frame:hover { stroke: #000; stroke-width: 1; cursor: pointer; }
    .label { font-family: monospace; font-size: 12px; fill: #000; pointer-events: none; }
    .title { font-family: sans-serif; font-size: 16px; font-weight: bold; fill: #333; }
    .empty { font-family: sans-serif; font-size: 14px; fill: #666; }
</style>`)
	svg.WriteString("\n")

	// Title
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="20" class="title">Flame Graph</text>`, width/2-50))
	svg.WriteString("\n")

	// Generate call tree
	root := fgg.Generate(profile)
	if root == nil {
		// Empty profile message
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="empty">No profile data available</text>`, width/2-80, height/2))
		svg.WriteString("\n")
		svg.WriteString("</svg>")
		return svg.String()
	}

	// Render the flame graph starting from root
	// Start at y=30 to leave room for title
	fgg.renderNode(&svg, root, 0, width, 30, 0, root.Samples)

	svg.WriteString("</svg>")
	return svg.String()
}

// renderNode recursively renders a call node and its children as SVG rectangles.
func (fgg *FlameGraphGenerator) renderNode(svg *strings.Builder, node *CallNode, x, nodeWidth, y, depth int, totalSamples int64) {
	if nodeWidth < 1 || node == nil {
		return
	}

	// Calculate rectangle dimensions
	rectHeight := frameHeight

	// Get color based on depth
	color := getFlameColor(depth)

	// Render rectangle
	svg.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" class="frame" fill="%s">`, x, y, nodeWidth, rectHeight, color))
	svg.WriteString(fmt.Sprintf(`<title>%s (%.1f%%, %d samples)</title>`, escapeXML(node.Name), node.Percent, node.Samples))
	svg.WriteString("</rect>\n")

	// Render text label if there's enough space
	label := truncateLabel(node.Name, nodeWidth)
	if label != "" {
		textY := y + rectHeight - 4
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label">%s</text>`, x+2, textY, escapeXML(label)))
		svg.WriteString("\n")
	}

	// Render children
	childX := x
	childY := y + rectHeight + framePadding
	for _, child := range node.Children {
		// Calculate child width proportional to its samples
		var childWidth int
		if totalSamples > 0 {
			childWidth = int(float64(nodeWidth) * float64(child.Samples) / float64(node.Samples))
		}
		if childWidth < 1 {
			childWidth = 1
		}

		fgg.renderNode(svg, child, childX, childWidth, childY, depth+1, totalSamples)
		childX += childWidth
	}
}

// GetWidth returns the current SVG width.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fgg *FlameGraphGenerator) GetWidth() int {
	fgg.mu.RLock()
	defer fgg.mu.RUnlock()
	return fgg.width
}

// GetHeight returns the current SVG height.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fgg *FlameGraphGenerator) GetHeight() int {
	fgg.mu.RLock()
	defer fgg.mu.RUnlock()
	return fgg.height
}

// SetDimensions sets the SVG dimensions.
//
// Invalid dimensions (<= 0) are ignored.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fgg *FlameGraphGenerator) SetDimensions(width, height int) {
	fgg.mu.Lock()
	defer fgg.mu.Unlock()

	if width > 0 {
		fgg.width = width
	}
	if height > 0 {
		fgg.height = height
	}
}

// Reset resets the generator to default dimensions.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (fgg *FlameGraphGenerator) Reset() {
	fgg.mu.Lock()
	defer fgg.mu.Unlock()
	fgg.width = DefaultFlameGraphWidth
	fgg.height = DefaultFlameGraphHeight
}

// TotalSamples returns the total samples for this node.
func (cn *CallNode) TotalSamples() int64 {
	return cn.Samples
}

// AddChild adds a child node.
//
// Nil children are ignored.
func (cn *CallNode) AddChild(child *CallNode) {
	if child == nil {
		return
	}
	cn.Children = append(cn.Children, child)
}

// getFlameColor returns a flame-like color based on depth.
//
// Colors range from red/orange at the bottom to yellow at the top,
// mimicking the appearance of flames.
func getFlameColor(depth int) string {
	// Flame colors: red -> orange -> yellow
	// Use a gradient based on depth
	colors := []string{
		"rgb(255, 100, 50)",  // Red-orange (depth 0)
		"rgb(255, 120, 60)",  // Orange
		"rgb(255, 140, 70)",  // Light orange
		"rgb(255, 160, 80)",  // Yellow-orange
		"rgb(255, 180, 90)",  // Light yellow-orange
		"rgb(255, 200, 100)", // Yellow
		"rgb(255, 210, 110)", // Light yellow
		"rgb(255, 220, 120)", // Pale yellow
	}

	// Cycle through colors for deep stacks
	return colors[depth%len(colors)]
}

// truncateLabel truncates a label to fit within a given width.
//
// Returns empty string if the width is too small to display any text.
func truncateLabel(label string, maxWidth int) string {
	if label == "" || maxWidth < minWidthForLabel {
		return ""
	}

	// Calculate max characters that fit
	maxChars := maxWidth / charWidth
	if maxChars <= 3 {
		return ""
	}

	if len(label) <= maxChars {
		return label
	}

	// Truncate with ellipsis
	return label[:maxChars-3] + "..."
}

// escapeXML escapes special XML characters in a string.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
