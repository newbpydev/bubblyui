package testutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDependencyTrackingInspector_TrackDependency tests basic dependency tracking
func TestDependencyTrackingInspector_TrackDependency(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Track simple dependency
	inspector.TrackDependency("computed1", "ref1")

	// Verify dependency was tracked
	inspector.AssertDependency(t, "computed1", "ref1")
}

// TestDependencyTrackingInspector_MultipleDependencies tests tracking multiple dependencies
func TestDependencyTrackingInspector_MultipleDependencies(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Track multiple dependencies for same source
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed1", "ref2")
	inspector.TrackDependency("computed1", "ref3")

	// Verify all dependencies tracked
	inspector.AssertDependency(t, "computed1", "ref1")
	inspector.AssertDependency(t, "computed1", "ref2")
	inspector.AssertDependency(t, "computed1", "ref3")
}

// TestDependencyTrackingInspector_DependencyChain tests chained dependencies
func TestDependencyTrackingInspector_DependencyChain(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Create dependency chain: ref1 -> computed1 -> computed2
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed2", "computed1")

	// Verify chain
	inspector.AssertDependency(t, "computed1", "ref1")
	inspector.AssertDependency(t, "computed2", "computed1")
}

// TestDependencyTrackingInspector_GetDependencyGraph tests graph retrieval
func TestDependencyTrackingInspector_GetDependencyGraph(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Build graph
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed1", "ref2")
	inspector.TrackDependency("computed2", "ref1")

	// Get graph
	graph := inspector.GetDependencyGraph()

	// Verify graph structure
	assert.NotNil(t, graph)
	assert.Len(t, graph.Nodes, 4) // ref1, ref2, computed1, computed2
	assert.True(t, len(graph.Edges) >= 3)
}

// TestDependencyTrackingInspector_VisualizeDependencies tests graph visualization
func TestDependencyTrackingInspector_VisualizeDependencies(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Build simple graph
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed2", "computed1")

	// Visualize
	output := inspector.VisualizeDependencies()

	// Verify output contains nodes and edges
	assert.Contains(t, output, "ref1")
	assert.Contains(t, output, "computed1")
	assert.Contains(t, output, "computed2")
	assert.Contains(t, output, "->") // Arrow indicating dependency
}

// TestDependencyTrackingInspector_CircularDependency tests circular dependency detection
func TestDependencyTrackingInspector_CircularDependency(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Create circular dependency: A -> B -> C -> A
	inspector.TrackDependency("A", "B")
	inspector.TrackDependency("B", "C")
	inspector.TrackDependency("C", "A")

	// Detect circular dependencies
	circular := inspector.DetectCircularDependencies()

	// Verify circular dependency detected
	assert.True(t, len(circular) > 0, "Should detect circular dependency")
	assert.Contains(t, circular[0], "A")
	assert.Contains(t, circular[0], "B")
	assert.Contains(t, circular[0], "C")
}

// TestDependencyTrackingInspector_OrphanedDependencies tests orphaned node detection
func TestDependencyTrackingInspector_OrphanedDependencies(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Create graph with orphaned nodes
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed2", "ref2")
	// ref3 and ref4 are orphaned (no dependencies)

	// Find orphaned dependencies
	orphaned := inspector.FindOrphanedDependencies()

	// In this case, all leaf nodes (ref1, ref2) could be considered orphaned
	// if they have no outgoing edges
	assert.NotNil(t, orphaned)
}

// TestDependencyTrackingInspector_ComplexGraph tests complex dependency graph
func TestDependencyTrackingInspector_ComplexGraph(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Build complex graph
	// ref1, ref2, ref3 (base refs)
	// computed1 depends on ref1, ref2
	// computed2 depends on ref2, ref3
	// computed3 depends on computed1, computed2
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed1", "ref2")
	inspector.TrackDependency("computed2", "ref2")
	inspector.TrackDependency("computed2", "ref3")
	inspector.TrackDependency("computed3", "computed1")
	inspector.TrackDependency("computed3", "computed2")

	// Get graph
	graph := inspector.GetDependencyGraph()

	// Verify graph structure
	assert.NotNil(t, graph)
	assert.Len(t, graph.Nodes, 6) // 3 refs + 3 computed

	// Verify visualization
	output := inspector.VisualizeDependencies()
	assert.Contains(t, output, "computed3")
	assert.Contains(t, output, "computed1")
	assert.Contains(t, output, "ref1")
}

// TestDependencyTrackingInspector_PerformanceWithManyDeps tests performance with many dependencies
func TestDependencyTrackingInspector_PerformanceWithManyDeps(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Create many dependencies
	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			inspector.TrackDependency(
				"computed"+string(rune(i)),
				"ref"+string(rune(j)),
			)
		}
	}

	// Get graph (should not hang or crash)
	graph := inspector.GetDependencyGraph()
	assert.NotNil(t, graph)

	// Visualize (should complete in reasonable time)
	output := inspector.VisualizeDependencies()
	assert.NotEmpty(t, output)
}

// TestDependencyTrackingInspector_TableDriven tests various scenarios
func TestDependencyTrackingInspector_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		dependencies map[string][]string // source -> targets
		wantCircular bool
		wantNodes    int
	}{
		{
			name: "simple chain",
			dependencies: map[string][]string{
				"computed1": {"ref1"},
				"computed2": {"computed1"},
			},
			wantCircular: false,
			wantNodes:    3,
		},
		{
			name: "diamond pattern",
			dependencies: map[string][]string{
				"computed1": {"ref1"},
				"computed2": {"ref1"},
				"computed3": {"computed1", "computed2"},
			},
			wantCircular: false,
			wantNodes:    4,
		},
		{
			name: "circular",
			dependencies: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			wantCircular: true,
			wantNodes:    3,
		},
		{
			name: "self-reference",
			dependencies: map[string][]string{
				"A": {"A"},
			},
			wantCircular: true,
			wantNodes:    1,
		},
		{
			name: "multiple sources",
			dependencies: map[string][]string{
				"computed1": {"ref1", "ref2"},
				"computed2": {"ref3", "ref4"},
				"computed3": {"computed1", "computed2"},
			},
			wantCircular: false,
			wantNodes:    7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := NewDependencyTrackingInspector()

			// Build graph
			for source, targets := range tt.dependencies {
				for _, target := range targets {
					inspector.TrackDependency(source, target)
				}
			}

			// Check circular dependencies
			circular := inspector.DetectCircularDependencies()
			hasCircular := len(circular) > 0
			assert.Equal(t, tt.wantCircular, hasCircular,
				"Circular dependency detection mismatch")

			// Check node count
			graph := inspector.GetDependencyGraph()
			assert.Equal(t, tt.wantNodes, len(graph.Nodes),
				"Node count mismatch")

			// Verify visualization works
			output := inspector.VisualizeDependencies()
			assert.NotEmpty(t, output)
		})
	}
}

// TestDependencyTrackingInspector_EmptyGraph tests empty graph handling
func TestDependencyTrackingInspector_EmptyGraph(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Get empty graph
	graph := inspector.GetDependencyGraph()
	assert.NotNil(t, graph)
	assert.Len(t, graph.Nodes, 0)
	assert.Len(t, graph.Edges, 0)

	// Visualize empty graph
	output := inspector.VisualizeDependencies()
	assert.Contains(t, output, "empty") // Should indicate empty graph

	// Detect circular in empty graph
	circular := inspector.DetectCircularDependencies()
	assert.Len(t, circular, 0)

	// Find orphaned in empty graph
	orphaned := inspector.FindOrphanedDependencies()
	assert.Len(t, orphaned, 0)
}

// TestDependencyTrackingInspector_DuplicateDependencies tests duplicate tracking
func TestDependencyTrackingInspector_DuplicateDependencies(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Track same dependency multiple times
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed1", "ref1")

	// Should only store once
	graph := inspector.GetDependencyGraph()
	assert.Len(t, graph.Nodes, 2) // computed1, ref1

	// Count edges - should be 1 (deduplicated)
	edgeCount := 0
	for _, edge := range graph.Edges {
		if edge.From == "computed1" && edge.To == "ref1" {
			edgeCount++
		}
	}
	assert.Equal(t, 1, edgeCount, "Duplicate edges should be deduplicated")
}

// TestDependencyTrackingInspector_VisualizationFormat tests visualization output format
func TestDependencyTrackingInspector_VisualizationFormat(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Build simple graph
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed1", "ref2")

	output := inspector.VisualizeDependencies()

	// Verify format
	lines := strings.Split(output, "\n")
	assert.True(t, len(lines) > 0, "Should have output lines")

	// Should contain dependency arrows
	assert.Contains(t, output, "->")

	// Should list all nodes
	assert.Contains(t, output, "computed1")
	assert.Contains(t, output, "ref1")
	assert.Contains(t, output, "ref2")
}

// TestDependencyTrackingInspector_AssertDependencyFailure tests assertion failures
func TestDependencyTrackingInspector_AssertDependencyFailure(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Track one dependency
	inspector.TrackDependency("computed1", "ref1")

	// Create a mock testing.T to capture failures
	mockT := &testing.T{}

	// This should fail - source doesn't exist
	inspector.AssertDependency(mockT, "nonexistent", "ref1")
	if !mockT.Failed() {
		t.Error("Expected assertion to fail for nonexistent source")
	}

	// Reset mock
	mockT = &testing.T{}

	// This should fail - target doesn't match
	inspector.AssertDependency(mockT, "computed1", "ref2")
	if !mockT.Failed() {
		t.Error("Expected assertion to fail for wrong target")
	}
}

// TestDependencyTrackingInspector_IsolatedNodes tests isolated node visualization
func TestDependencyTrackingInspector_IsolatedNodes(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Add isolated node (no dependencies)
	inspector.nodes["isolated1"] = true
	inspector.nodes["isolated2"] = true

	// Add connected nodes
	inspector.TrackDependency("computed1", "ref1")

	output := inspector.VisualizeDependencies()

	// Should show isolated nodes section
	assert.Contains(t, output, "Isolated nodes")
	assert.Contains(t, output, "isolated1")
	assert.Contains(t, output, "isolated2")
}

// TestDependencyTrackingInspector_LeafNodes tests leaf node handling
func TestDependencyTrackingInspector_LeafNodes(t *testing.T) {
	inspector := NewDependencyTrackingInspector()

	// Create chain where ref1 is a leaf (target only, no outgoing)
	inspector.TrackDependency("computed1", "ref1")
	inspector.TrackDependency("computed2", "ref1")

	// ref1 should NOT be orphaned (it's a target)
	orphaned := inspector.FindOrphanedDependencies()
	assert.NotContains(t, orphaned, "ref1")

	// computed1 and computed2 should NOT be orphaned (they have outgoing)
	assert.NotContains(t, orphaned, "computed1")
	assert.NotContains(t, orphaned, "computed2")
}
