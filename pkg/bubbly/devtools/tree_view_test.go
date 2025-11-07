package devtools

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTreeView tests tree view creation
func TestNewTreeView(t *testing.T) {
	tests := []struct {
		name     string
		root     *ComponentSnapshot
		wantRoot bool
	}{
		{
			name: "with root component",
			root: &ComponentSnapshot{
				ID:   "root-1",
				Name: "App",
			},
			wantRoot: true,
		},
		{
			name:     "with nil root",
			root:     nil,
			wantRoot: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv := NewTreeView(tt.root)

			require.NotNil(t, tv)
			if tt.wantRoot {
				assert.Equal(t, tt.root, tv.GetRoot())
			} else {
				assert.Nil(t, tv.GetRoot())
			}
		})
	}
}

// TestTreeView_Render tests basic tree rendering
func TestTreeView_Render(t *testing.T) {
	tests := []struct {
		name          string
		root          *ComponentSnapshot
		expectedLines []string
	}{
		{
			name: "single component",
			root: &ComponentSnapshot{
				ID:   "root-1",
				Name: "App",
				Refs: []*RefSnapshot{},
			},
			expectedLines: []string{
				"App (0 refs)",
			},
		},
		{
			name: "component with refs",
			root: &ComponentSnapshot{
				ID:   "root-1",
				Name: "Counter",
				Refs: []*RefSnapshot{
					{Name: "count"},
					{Name: "step"},
				},
			},
			expectedLines: []string{
				"Counter (2 refs)",
			},
		},
		{
			name: "component with children (collapsed)",
			root: &ComponentSnapshot{
				ID:   "root-1",
				Name: "App",
				Children: []*ComponentSnapshot{
					{ID: "child-1", Name: "Header"},
					{ID: "child-2", Name: "Content"},
				},
			},
			expectedLines: []string{
				"▶ App (0 refs)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv := NewTreeView(tt.root)
			output := tv.Render()

			for _, expected := range tt.expectedLines {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// TestTreeView_RenderExpanded tests rendering with expanded nodes
func TestTreeView_RenderExpanded(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{
				ID:   "child-1",
				Name: "Header",
				Refs: []*RefSnapshot{{Name: "title"}},
			},
			{
				ID:   "child-2",
				Name: "Content",
				Children: []*ComponentSnapshot{
					{ID: "grandchild-1", Name: "Article"},
				},
			},
		},
	}

	tv := NewTreeView(root)
	tv.Expand("root-1")

	output := tv.Render()

	// Should show expanded icon and children
	assert.Contains(t, output, "▼ App")
	assert.Contains(t, output, "Header (1 refs)")
	assert.Contains(t, output, "Content (0 refs)")

	// Grandchild should not be visible (Content is collapsed)
	assert.NotContains(t, output, "Article")
}

// TestTreeView_RenderNested tests deeply nested tree rendering
func TestTreeView_RenderNested(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{
				ID:   "child-1",
				Name: "Layout",
				Children: []*ComponentSnapshot{
					{
						ID:   "grandchild-1",
						Name: "Sidebar",
						Children: []*ComponentSnapshot{
							{ID: "great-grandchild-1", Name: "Menu"},
						},
					},
				},
			},
		},
	}

	tv := NewTreeView(root)
	tv.Expand("root-1")
	tv.Expand("child-1")
	tv.Expand("grandchild-1")

	output := tv.Render()

	// All levels should be visible
	assert.Contains(t, output, "App")
	assert.Contains(t, output, "Layout")
	assert.Contains(t, output, "Sidebar")
	assert.Contains(t, output, "Menu")

	// Check indentation (approximate - actual rendering may vary)
	lines := strings.Split(output, "\n")
	assert.True(t, len(lines) >= 4, "Should have at least 4 lines for 4 components")
}

// TestTreeView_Select tests component selection
func TestTreeView_Select(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{ID: "child-1", Name: "Header"},
			{ID: "child-2", Name: "Content"},
		},
	}

	tv := NewTreeView(root)

	// Initially no selection
	assert.Nil(t, tv.GetSelected())

	// Select root
	tv.Select("root-1")
	selected := tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "root-1", selected.ID)

	// Select child
	tv.Select("child-1")
	selected = tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "child-1", selected.ID)

	// Select non-existent component
	tv.Select("non-existent")
	// Should keep previous selection
	selected = tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "child-1", selected.ID)
}

// TestTreeView_SelectRendering tests selection indicator in rendering
func TestTreeView_SelectRendering(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{ID: "child-1", Name: "Header"},
		},
	}

	tv := NewTreeView(root)
	tv.Expand("root-1")
	tv.Select("child-1")

	output := tv.Render()

	// Selected component should have indicator
	lines := strings.Split(output, "\n")
	var foundSelected bool
	for _, line := range lines {
		if strings.Contains(line, "Header") {
			assert.Contains(t, line, "►", "Selected component should have ► indicator")
			foundSelected = true
			break
		}
	}
	assert.True(t, foundSelected, "Should find selected component in output")
}

// TestTreeView_ExpandCollapse tests expand/collapse functionality
func TestTreeView_ExpandCollapse(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{ID: "child-1", Name: "Header"},
		},
	}

	tv := NewTreeView(root)

	// Initially collapsed
	assert.False(t, tv.IsExpanded("root-1"))

	// Expand
	tv.Expand("root-1")
	assert.True(t, tv.IsExpanded("root-1"))

	// Collapse
	tv.Collapse("root-1")
	assert.False(t, tv.IsExpanded("root-1"))

	// Toggle
	tv.Toggle("root-1")
	assert.True(t, tv.IsExpanded("root-1"))
	tv.Toggle("root-1")
	assert.False(t, tv.IsExpanded("root-1"))
}

// TestTreeView_Navigation tests up/down navigation
func TestTreeView_Navigation(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{ID: "child-1", Name: "Header"},
			{ID: "child-2", Name: "Content"},
			{ID: "child-3", Name: "Footer"},
		},
	}

	tv := NewTreeView(root)
	tv.Expand("root-1")

	// Start at root
	tv.Select("root-1")

	// Navigate down
	tv.SelectNext()
	selected := tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "child-1", selected.ID)

	// Navigate down again
	tv.SelectNext()
	selected = tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "child-2", selected.ID)

	// Navigate up
	tv.SelectPrevious()
	selected = tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "child-1", selected.ID)

	// Navigate up to root
	tv.SelectPrevious()
	selected = tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "root-1", selected.ID)

	// Navigate up at root (should stay at root)
	tv.SelectPrevious()
	selected = tv.GetSelected()
	require.NotNil(t, selected)
	assert.Equal(t, "root-1", selected.ID)
}

// TestTreeView_NavigationWithCollapsed tests navigation skips collapsed children
func TestTreeView_NavigationWithCollapsed(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{
				ID:   "child-1",
				Name: "Header",
				Children: []*ComponentSnapshot{
					{ID: "grandchild-1", Name: "Logo"},
				},
			},
			{ID: "child-2", Name: "Footer"},
		},
	}

	tv := NewTreeView(root)
	tv.Expand("root-1")
	// child-1 is collapsed, so grandchild-1 is not visible

	tv.Select("root-1")

	// Navigate down - should go to child-1
	tv.SelectNext()
	assert.Equal(t, "child-1", tv.GetSelected().ID)

	// Navigate down - should skip grandchild-1 and go to child-2
	tv.SelectNext()
	assert.Equal(t, "child-2", tv.GetSelected().ID)
}

// TestTreeView_LargeTree tests performance with large trees
func TestTreeView_LargeTree(t *testing.T) {
	// Create a tree with 100 components
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
	}

	// Add 10 children, each with 9 grandchildren
	children := make([]*ComponentSnapshot, 10)
	for i := 0; i < 10; i++ {
		child := &ComponentSnapshot{
			ID:   "child-" + string(rune('0'+i)),
			Name: "Component" + string(rune('A'+i)),
		}

		grandchildren := make([]*ComponentSnapshot, 9)
		for j := 0; j < 9; j++ {
			grandchildren[j] = &ComponentSnapshot{
				ID:   "grandchild-" + string(rune('0'+i)) + string(rune('0'+j)),
				Name: "SubComponent" + string(rune('A'+i)) + string(rune('0'+j)),
			}
		}
		child.Children = grandchildren
		children[i] = child
	}
	root.Children = children

	tv := NewTreeView(root)

	// Expand all
	tv.Expand("root-1")
	for i := 0; i < 10; i++ {
		tv.Expand("child-" + string(rune('0'+i)))
	}

	// Measure rendering time
	start := time.Now()
	output := tv.Render()
	duration := time.Since(start)

	// Should render in reasonable time (< 50ms per requirement)
	assert.Less(t, duration.Milliseconds(), int64(50), "Rendering should be fast")
	assert.NotEmpty(t, output)

	// Should contain all components
	assert.Contains(t, output, "App")
	assert.Contains(t, output, "ComponentA")
	assert.Contains(t, output, "SubComponentA0")
}

// TestTreeView_EmptyTree tests handling of empty/nil tree
func TestTreeView_EmptyTree(t *testing.T) {
	tv := NewTreeView(nil)

	output := tv.Render()
	assert.Contains(t, output, "No components")

	// Operations should not panic
	tv.Select("any-id")
	tv.Expand("any-id")
	tv.Collapse("any-id")
	tv.SelectNext()
	tv.SelectPrevious()
}

// TestTreeView_ThreadSafety tests concurrent access
func TestTreeView_ThreadSafety(t *testing.T) {
	root := &ComponentSnapshot{
		ID:   "root-1",
		Name: "App",
		Children: []*ComponentSnapshot{
			{ID: "child-1", Name: "Header"},
			{ID: "child-2", Name: "Content"},
		},
	}

	tv := NewTreeView(root)

	// Concurrent operations
	done := make(chan bool)

	// Goroutine 1: Expand/collapse
	go func() {
		for i := 0; i < 100; i++ {
			tv.Expand("root-1")
			tv.Collapse("root-1")
		}
		done <- true
	}()

	// Goroutine 2: Select
	go func() {
		for i := 0; i < 100; i++ {
			tv.Select("child-1")
			tv.Select("child-2")
		}
		done <- true
	}()

	// Goroutine 3: Render
	go func() {
		for i := 0; i < 100; i++ {
			_ = tv.Render()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}
