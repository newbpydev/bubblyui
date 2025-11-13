package devtools

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDevTools_TreeExpansion_E2E tests the complete flow of:
// 1. Adding components with children
// 2. Seeing them in the tree
// 3. Expanding/collapsing with Enter key
//
// This reproduces the user's reported issue:
// - "Can't expand tree"
// - "Press Enter and nothing happens"
func TestDevTools_TreeExpansion_E2E(t *testing.T) {
	// Reset singleton
	resetSingleton()
	
	// 1. Enable DevTools
	dt := Enable()
	require.NotNil(t, dt)
	dt.SetVisible(true)
	
	// 2. Create parent component
	parentSnapshot := &ComponentSnapshot{
		ID:   "parent-1",
		Name: "ParentComponent",
		Type: "Component",
		Refs: []*RefSnapshot{
			{ID: "ref-0x1", Name: "count", Value: "0", Type: "int"},
		},
	}
	dt.store.AddComponent(parentSnapshot)
	
	// 3. Create child components
	child1 := &ComponentSnapshot{
		ID:   "child-1",
		Name: "ChildOne",
		Type: "Component",
	}
	child2 := &ComponentSnapshot{
		ID:   "child-2", 
		Name: "ChildTwo",
		Type: "Component",
	}
	dt.store.AddComponent(child1)
	dt.store.AddComponent(child2)
	
	// 4. Establish parent-child relationships
	dt.store.AddComponentChild("parent-1", "child-1")
	dt.store.AddComponentChild("parent-1", "child-2")
	
	// 5. Get root components - should have children populated
	roots := dt.store.GetRootComponents()
	require.Len(t, roots, 1, "Should have one root component")
	
	root := roots[0]
	assert.Equal(t, "parent-1", root.ID)
	assert.Len(t, root.Children, 2, "CRITICAL: Parent must have 2 children in snapshot")
	assert.Equal(t, "child-1", root.Children[0].ID)
	assert.Equal(t, "child-2", root.Children[1].ID)
	
	// 6. Verify refs are included
	assert.Len(t, root.Refs, 1, "CRITICAL: Parent must have 1 ref in snapshot")
	
	// 7. Now test the Inspector UI
	dt.ui.inspector.SetRoot(root)
	
	// 8. Verify tree has root with children
	treeRoot := dt.ui.inspector.tree.GetRoot()
	require.NotNil(t, treeRoot)
	assert.Len(t, treeRoot.Children, 2, "Tree root must have children")
	
	// 9. Initially, root should be auto-expanded
	assert.True(t, dt.ui.inspector.tree.IsExpanded("parent-1"), "Root should be auto-expanded")
	
	// 10. Collapse by pressing Enter
	cmd := dt.ui.inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd, "Should return ClearScreen cmd")
	
	// 11. Verify collapse worked
	assert.False(t, dt.ui.inspector.tree.IsExpanded("parent-1"), "CRITICAL: Root should be collapsed after Enter")
	
	// 12. Expand again by pressing Enter
	cmd = dt.ui.inspector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd, "Should return ClearScreen cmd")
	
	// 13. Verify expansion worked
	assert.True(t, dt.ui.inspector.tree.IsExpanded("parent-1"), "CRITICAL: Root should be expanded after second Enter")
	
	// 14. Verify rendered tree shows expand/collapse indicator
	rendered := dt.ui.inspector.tree.Render()
	t.Logf("Rendered tree:\n%s", rendered)
	
	// When expanded, should show ▼
	// When collapsed, should show ▶
	assert.Contains(t, rendered, "ParentComponent", "Should show component name")
	
	// 15. Verify detail panel shows refs
	detail := dt.ui.inspector.detail.GetComponent()
	require.NotNil(t, detail)
	assert.Len(t, detail.Refs, 1, "Detail panel should show refs")
}

// TestDevTools_StateTab_ShowsRefs tests that the State tab displays refs correctly
func TestDevTools_StateTab_ShowsRefs(t *testing.T) {
	// Reset singleton
	resetSingleton()
	
	// 1. Enable DevTools
	dt := Enable()
	require.NotNil(t, dt)
	dt.SetVisible(true)
	
	// 2. Create component with refs
	snapshot := &ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Refs: []*RefSnapshot{
			{ID: "ref-0x1", Name: "count", Value: "42", Type: "int"},
			{ID: "ref-0x2", Name: "isEven", Value: "true", Type: "bool"},
		},
	}
	dt.store.AddComponent(snapshot)
	
	// 3. Set as inspector root
	dt.ui.inspector.SetRoot(snapshot)
	
	// 4. Render State tab (index 0)
	dt.ui.inspector.detail.SwitchTab(0)
	stateContent := renderStateTab(snapshot)
	
	t.Logf("State tab content:\n%s", stateContent)
	
	// 5. Verify refs are shown
	assert.Contains(t, stateContent, "count", "Should show count ref")
	assert.Contains(t, stateContent, "42", "Should show count value")
	assert.Contains(t, stateContent, "isEven", "Should show isEven ref")
	assert.Contains(t, stateContent, "true", "Should show isEven value")
	assert.NotContains(t, stateContent, "No reactive state", "CRITICAL: Should NOT show 'No reactive state'")
}
