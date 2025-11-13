package devtools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// TreeView displays a hierarchical tree of component snapshots.
//
// The tree view supports:
// - Expanding/collapsing nodes to show/hide children
// - Selecting components for detailed inspection
// - Keyboard navigation (up/down)
// - Visual indicators for selection and expansion state
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	tv := devtools.NewTreeView(rootSnapshot)
//	tv.Expand("component-1")
//	tv.Select("component-2")
//	output := tv.Render()
type TreeView struct {
	mu       sync.RWMutex
	root     *ComponentSnapshot
	selected *ComponentSnapshot
	expanded map[string]bool
}

// NewTreeView creates a new tree view with the given root component.
//
// The root component can be nil, in which case the tree view will display
// an empty state message.
func NewTreeView(root *ComponentSnapshot) *TreeView {
	return &TreeView{
		root:     root,
		expanded: make(map[string]bool),
	}
}

// GetRoot returns the root component snapshot.
func (tv *TreeView) GetRoot() *ComponentSnapshot {
	tv.mu.RLock()
	defer tv.mu.RUnlock()
	return tv.root
}

// GetSelected returns the currently selected component snapshot.
func (tv *TreeView) GetSelected() *ComponentSnapshot {
	tv.mu.RLock()
	defer tv.mu.RUnlock()
	return tv.selected
}

// Render generates the visual representation of the component tree.
//
// The output includes:
// - Tree structure with indentation
// - Expand/collapse indicators (▶/▼)
// - Selection indicator (►)
// - Component names and ref counts
//
// Returns an empty message if the tree has no root component.
func (tv *TreeView) Render() string {
	tv.mu.RLock()
	defer tv.mu.RUnlock()

	if tv.root == nil {
		return tv.renderEmpty()
	}

	var lines []string
	tv.renderNode(tv.root, 0, &lines)

	return strings.Join(lines, "\n")
}

// renderEmpty returns a message for empty trees.
func (tv *TreeView) renderEmpty() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	return style.Render("No components")
}

// renderNode recursively renders a component node and its children.
func (tv *TreeView) renderNode(node *ComponentSnapshot, depth int, lines *[]string) {
	if node == nil {
		return
	}

	// Build the line for this node
	line := tv.buildNodeLine(node, depth)
	*lines = append(*lines, line)

	// Render children if expanded
	if tv.isExpandedUnsafe(node.ID) && len(node.Children) > 0 {
		for _, child := range node.Children {
			tv.renderNode(child, depth+1, lines)
		}
	}
}

// buildNodeLine constructs the display line for a single node.
func (tv *TreeView) buildNodeLine(node *ComponentSnapshot, depth int) string {
	// Indentation
	indent := strings.Repeat("  ", depth)

	// Expand/collapse indicator
	expandIcon := ""
	if len(node.Children) > 0 {
		if tv.isExpandedUnsafe(node.ID) {
			expandIcon = "▼ "
		} else {
			expandIcon = "▶ "
		}
	}

	// Selection indicator
	selectionPrefix := " "
	if tv.selected != nil && tv.selected.ID == node.ID {
		selectionPrefix = "►"
	}

	// Component info
	refCount := len(node.Refs)
	componentInfo := fmt.Sprintf("%s (%d refs)", node.Name, refCount)

	// Combine all parts
	line := fmt.Sprintf("%s%s%s%s", selectionPrefix, indent, expandIcon, componentInfo)

	// Apply styling - CRITICAL: Background highlight for selection visibility
	style := lipgloss.NewStyle()
	if tv.selected != nil && tv.selected.ID == node.ID {
		// Selected item: bright background + contrasting text for clear visibility
		style = style.
			Background(lipgloss.Color("99")).  // Purple/blue background
			Foreground(lipgloss.Color("15")).  // White text
			Bold(true)
	} else {
		// Normal item: standard text color
		style = style.Foreground(lipgloss.Color("252"))  // Light gray
	}

	return style.Render(line)
}

// Select sets the selected component by ID.
//
// If the component with the given ID is not found in the tree,
// the selection remains unchanged.
func (tv *TreeView) Select(id string) {
	tv.mu.Lock()
	defer tv.mu.Unlock()

	// Find the component in the tree
	component := tv.findComponentUnsafe(tv.root, id)
	if component != nil {
		tv.selected = component
	}
}

// Expand expands a component node to show its children.
func (tv *TreeView) Expand(id string) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	tv.expanded[id] = true
}

// Collapse collapses a component node to hide its children.
func (tv *TreeView) Collapse(id string) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	delete(tv.expanded, id)
}

// Toggle toggles the expansion state of a component node.
func (tv *TreeView) Toggle(id string) {
	tv.mu.Lock()
	defer tv.mu.Unlock()

	if tv.expanded[id] {
		delete(tv.expanded, id)
	} else {
		tv.expanded[id] = true
	}
}

// IsExpanded returns whether a component node is expanded.
func (tv *TreeView) IsExpanded(id string) bool {
	tv.mu.RLock()
	defer tv.mu.RUnlock()
	return tv.expanded[id]
}

// isExpandedUnsafe checks expansion state without locking (internal use).
func (tv *TreeView) isExpandedUnsafe(id string) bool {
	return tv.expanded[id]
}

// GetExpandedIDs returns a copy of all currently expanded node IDs.
// This is used to preserve expansion state when rebuilding the tree.
func (tv *TreeView) GetExpandedIDs() map[string]bool {
	tv.mu.RLock()
	defer tv.mu.RUnlock()
	
	// Return a copy to prevent external modification
	expandedCopy := make(map[string]bool, len(tv.expanded))
	for id, expanded := range tv.expanded {
		expandedCopy[id] = expanded
	}
	return expandedCopy
}

// SetExpandedIDs sets which nodes should be expanded.
// This is used to restore expansion state after rebuilding the tree.
func (tv *TreeView) SetExpandedIDs(expanded map[string]bool) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	
	tv.expanded = make(map[string]bool, len(expanded))
	for id, exp := range expanded {
		tv.expanded[id] = exp
	}
}

// SelectNext selects the next visible component in the tree.
//
// Navigation follows depth-first traversal order, respecting
// the current expansion state of nodes.
func (tv *TreeView) SelectNext() {
	tv.mu.Lock()
	defer tv.mu.Unlock()

	if tv.root == nil {
		return
	}

	// Build flat list of visible components
	visible := tv.getVisibleComponentsUnsafe()
	if len(visible) == 0 {
		return
	}

	// If nothing selected, select first
	if tv.selected == nil {
		tv.selected = visible[0]
		return
	}

	// Find current selection in visible list
	for i, comp := range visible {
		if comp.ID == tv.selected.ID {
			// Select next if not at end
			if i < len(visible)-1 {
				tv.selected = visible[i+1]
			}
			return
		}
	}

	// Current selection not visible, select first
	tv.selected = visible[0]
}

// SelectPrevious selects the previous visible component in the tree.
func (tv *TreeView) SelectPrevious() {
	tv.mu.Lock()
	defer tv.mu.Unlock()

	if tv.root == nil {
		return
	}

	// Build flat list of visible components
	visible := tv.getVisibleComponentsUnsafe()
	if len(visible) == 0 {
		return
	}

	// If nothing selected, select first
	if tv.selected == nil {
		tv.selected = visible[0]
		return
	}

	// Find current selection in visible list
	for i, comp := range visible {
		if comp.ID == tv.selected.ID {
			// Select previous if not at start
			if i > 0 {
				tv.selected = visible[i-1]
			}
			return
		}
	}

	// Current selection not visible, select first
	tv.selected = visible[0]
}

// getVisibleComponentsUnsafe returns a flat list of all visible components.
// Must be called with lock held.
func (tv *TreeView) getVisibleComponentsUnsafe() []*ComponentSnapshot {
	var visible []*ComponentSnapshot
	tv.collectVisibleUnsafe(tv.root, &visible)
	return visible
}

// collectVisibleUnsafe recursively collects visible components.
func (tv *TreeView) collectVisibleUnsafe(node *ComponentSnapshot, visible *[]*ComponentSnapshot) {
	if node == nil {
		return
	}

	*visible = append(*visible, node)

	// Add children if expanded
	if tv.isExpandedUnsafe(node.ID) {
		for _, child := range node.Children {
			tv.collectVisibleUnsafe(child, visible)
		}
	}
}

// findComponentUnsafe searches for a component by ID in the tree.
// Must be called with lock held.
func (tv *TreeView) findComponentUnsafe(node *ComponentSnapshot, id string) *ComponentSnapshot {
	if node == nil {
		return nil
	}

	if node.ID == id {
		return node
	}

	// Search children
	for _, child := range node.Children {
		if found := tv.findComponentUnsafe(child, id); found != nil {
			return found
		}
	}

	return nil
}
