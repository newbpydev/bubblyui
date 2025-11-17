package testutil

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// DependencyNode represents a node in the dependency graph.
type DependencyNode struct {
	Name string
}

// DependencyEdge represents a directed edge in the dependency graph.
type DependencyEdge struct {
	From string // Source node (dependent)
	To   string // Target node (dependency)
}

// DependencyGraph represents the complete dependency graph structure.
type DependencyGraph struct {
	Nodes []DependencyNode
	Edges []DependencyEdge
}

// DependencyTrackingInspector provides utilities for testing dependency graph tracking and visualization.
// It tracks dependencies between reactive values (refs and computed values) and provides methods
// to verify the dependency graph structure, detect circular dependencies, and visualize the graph.
//
// Key Features:
//   - Track source -> target dependency relationships
//   - Verify specific dependencies exist
//   - Get complete dependency graph structure
//   - Visualize dependencies as text
//   - Detect circular dependencies
//   - Find orphaned dependencies
//
// Use Cases:
//   - Testing computed value dependency tracking
//   - Verifying watch dependencies
//   - Detecting circular dependency bugs
//   - Visualizing complex dependency chains
//   - Performance testing with large graphs
//
// Example:
//
//	inspector := NewDependencyTrackingInspector()
//
//	// Track dependencies
//	inspector.TrackDependency("computed1", "ref1")
//	inspector.TrackDependency("computed2", "computed1")
//
//	// Verify dependencies
//	inspector.AssertDependency(t, "computed1", "ref1")
//
//	// Visualize graph
//	fmt.Println(inspector.VisualizeDependencies())
//
//	// Detect circular dependencies
//	circular := inspector.DetectCircularDependencies()
//	if len(circular) > 0 {
//	    t.Errorf("Circular dependency detected: %v", circular)
//	}
//
// Thread Safety:
//
// DependencyTrackingInspector is not thread-safe. It should only be used from a single test goroutine.
type DependencyTrackingInspector struct {
	// tracked maps source node to list of target nodes
	tracked map[string][]string
	// nodes stores all unique nodes in the graph
	nodes map[string]bool
}

// NewDependencyTrackingInspector creates a new DependencyTrackingInspector for testing dependency graphs.
//
// Returns:
//   - *DependencyTrackingInspector: A new inspector instance
//
// Example:
//
//	inspector := NewDependencyTrackingInspector()
//	inspector.TrackDependency("computed1", "ref1")
func NewDependencyTrackingInspector() *DependencyTrackingInspector {
	return &DependencyTrackingInspector{
		tracked: make(map[string][]string),
		nodes:   make(map[string]bool),
	}
}

// TrackDependency records a dependency relationship from source to target.
//
// This method tracks that 'source' depends on 'target'. For example, if a computed
// value 'computed1' depends on a ref 'ref1', you would call:
//
//	inspector.TrackDependency("computed1", "ref1")
//
// Duplicate dependencies are automatically deduplicated.
//
// Parameters:
//   - source: The dependent node (e.g., computed value)
//   - target: The dependency node (e.g., ref)
//
// Example:
//
//	// Computed value depends on two refs
//	inspector.TrackDependency("computed1", "ref1")
//	inspector.TrackDependency("computed1", "ref2")
//
//	// Chained dependencies
//	inspector.TrackDependency("computed2", "computed1")
func (dti *DependencyTrackingInspector) TrackDependency(source, target string) {
	// Add nodes to node set
	dti.nodes[source] = true
	dti.nodes[target] = true

	// Check if dependency already exists (deduplicate)
	if targets, exists := dti.tracked[source]; exists {
		for _, t := range targets {
			if t == target {
				return // Already tracked
			}
		}
	}

	// Add dependency
	dti.tracked[source] = append(dti.tracked[source], target)
}

// AssertDependency asserts that a dependency from source to target exists.
//
// This method verifies that TrackDependency was called with the given source and target.
// If the dependency does not exist, the test fails with a descriptive error message.
//
// Parameters:
//   - t: The testing.T instance
//   - source: The expected source node
//   - target: The expected target node
//
// Example:
//
//	inspector.TrackDependency("computed1", "ref1")
//	inspector.AssertDependency(t, "computed1", "ref1") // Passes
//	inspector.AssertDependency(t, "computed1", "ref2") // Fails
func (dti *DependencyTrackingInspector) AssertDependency(t *testing.T, source, target string) {
	t.Helper()

	targets, exists := dti.tracked[source]
	if !exists {
		t.Errorf("No dependencies tracked for source '%s'", source)
		return
	}

	for _, t := range targets {
		if t == target {
			return // Found
		}
	}

	t.Errorf("Dependency '%s' -> '%s' not found. Available targets: %v",
		source, target, targets)
}

// GetDependencyGraph returns the complete dependency graph structure.
//
// The graph includes all nodes and edges that have been tracked. Nodes are sorted
// alphabetically for consistent output.
//
// Returns:
//   - *DependencyGraph: The complete graph with nodes and edges
//
// Example:
//
//	graph := inspector.GetDependencyGraph()
//	fmt.Printf("Nodes: %d, Edges: %d\n", len(graph.Nodes), len(graph.Edges))
func (dti *DependencyTrackingInspector) GetDependencyGraph() *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make([]DependencyNode, 0, len(dti.nodes)),
		Edges: make([]DependencyEdge, 0),
	}

	// Add all nodes (sorted for consistent output)
	nodeNames := make([]string, 0, len(dti.nodes))
	for name := range dti.nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	for _, name := range nodeNames {
		graph.Nodes = append(graph.Nodes, DependencyNode{Name: name})
	}

	// Add all edges
	for source, targets := range dti.tracked {
		for _, target := range targets {
			graph.Edges = append(graph.Edges, DependencyEdge{
				From: source,
				To:   target,
			})
		}
	}

	return graph
}

// VisualizeDependencies returns a text-based visualization of the dependency graph.
//
// The output format shows each dependency as "source -> target" on separate lines.
// Nodes are grouped by their dependencies for readability.
//
// Returns:
//   - string: Text representation of the dependency graph
//
// Example:
//
//	output := inspector.VisualizeDependencies()
//	fmt.Println(output)
//	// Output:
//	// Dependency Graph:
//	// computed1 -> ref1
//	// computed1 -> ref2
//	// computed2 -> computed1
func (dti *DependencyTrackingInspector) VisualizeDependencies() string {
	if len(dti.nodes) == 0 {
		return "Dependency Graph: (empty)"
	}

	var sb strings.Builder
	sb.WriteString("Dependency Graph:\n")

	// Sort sources for consistent output
	sources := make([]string, 0, len(dti.tracked))
	for source := range dti.tracked {
		sources = append(sources, source)
	}
	sort.Strings(sources)

	// Write each dependency
	for _, source := range sources {
		targets := dti.tracked[source]
		// Sort targets for consistent output
		sortedTargets := make([]string, len(targets))
		copy(sortedTargets, targets)
		sort.Strings(sortedTargets)

		for _, target := range sortedTargets {
			sb.WriteString(fmt.Sprintf("  %s -> %s\n", source, target))
		}
	}

	// List isolated nodes (nodes with no dependencies)
	isolatedNodes := make([]string, 0)
	for node := range dti.nodes {
		if _, hasOutgoing := dti.tracked[node]; !hasOutgoing {
			// Check if it's a target of any dependency
			isTarget := false
			for _, targets := range dti.tracked {
				for _, target := range targets {
					if target == node {
						isTarget = true
						break
					}
				}
				if isTarget {
					break
				}
			}
			if !isTarget {
				isolatedNodes = append(isolatedNodes, node)
			}
		}
	}

	if len(isolatedNodes) > 0 {
		sort.Strings(isolatedNodes)
		sb.WriteString("\nIsolated nodes:\n")
		for _, node := range isolatedNodes {
			sb.WriteString(fmt.Sprintf("  %s\n", node))
		}
	}

	return sb.String()
}

// DetectCircularDependencies detects circular dependencies in the graph.
//
// A circular dependency occurs when a node depends on itself through a chain
// of dependencies. For example: A -> B -> C -> A
//
// Returns:
//   - [][]string: List of circular dependency chains (empty if none found)
//
// Example:
//
//	inspector.TrackDependency("A", "B")
//	inspector.TrackDependency("B", "C")
//	inspector.TrackDependency("C", "A")
//
//	circular := inspector.DetectCircularDependencies()
//	// Returns: [["A", "B", "C", "A"]]
func (dti *DependencyTrackingInspector) DetectCircularDependencies() [][]string {
	var cycles [][]string

	// Track visited nodes and current path
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	// DFS to detect cycles
	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		// Check all dependencies
		if targets, exists := dti.tracked[node]; exists {
			for _, target := range targets {
				if !visited[target] {
					if dfs(target) {
						return true
					}
				} else if recStack[target] {
					// Found cycle - extract the cycle from path
					cycleStart := -1
					for i, n := range path {
						if n == target {
							cycleStart = i
							break
						}
					}
					if cycleStart >= 0 {
						cycle := make([]string, len(path)-cycleStart+1)
						copy(cycle, path[cycleStart:])
						cycle[len(cycle)-1] = target // Close the cycle
						cycles = append(cycles, cycle)
					}
					return true
				}
			}
		}

		// Backtrack
		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	// Check all nodes
	for node := range dti.nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// FindOrphanedDependencies finds nodes that have no incoming or outgoing dependencies.
//
// An orphaned node is a node that is neither a source nor a target of any dependency.
// This can indicate unused refs or computed values.
//
// Returns:
//   - []string: List of orphaned node names (sorted alphabetically)
//
// Example:
//
//	inspector.TrackDependency("computed1", "ref1")
//	// ref2 is added but never used
//	inspector.nodes["ref2"] = true
//
//	orphaned := inspector.FindOrphanedDependencies()
//	// Returns: ["ref2"]
func (dti *DependencyTrackingInspector) FindOrphanedDependencies() []string {
	orphaned := make([]string, 0)

	for node := range dti.nodes {
		// Check if node has outgoing dependencies
		hasOutgoing := false
		if _, exists := dti.tracked[node]; exists {
			hasOutgoing = true
		}

		// Check if node is a target of any dependency
		hasIncoming := false
		for _, targets := range dti.tracked {
			for _, target := range targets {
				if target == node {
					hasIncoming = true
					break
				}
			}
			if hasIncoming {
				break
			}
		}

		// If no incoming or outgoing, it's orphaned
		if !hasOutgoing && !hasIncoming {
			orphaned = append(orphaned, node)
		}
	}

	sort.Strings(orphaned)
	return orphaned
}
