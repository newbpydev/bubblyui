package devtools

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDevTools_RealWorldAppFlow tests the EXACT flow from the example app:
// 1. Enable DevTools
// 2. Create parent component (CounterApp)
// 3. In Setup(), create child components
// 4. ExposeComponent the children
// 5. Expose individual refs
// 6. Verify tree shows children and refs
//
// This reproduces the user's reported issues:
// - "Tree won't expand - shows ►CounterApp (0 refs)"
// - "State tab shows 'No reactive state'"
func TestDevTools_RealWorldAppFlow(t *testing.T) {
	// Reset singleton
	resetSingleton()
	
	// 1. Enable DevTools (like main.go does)
	dt := Enable()
	require.NotNil(t, dt)
	dt.SetVisible(true)
	
	// 2. Create parent component (like CreateApp() in app.go)
	parent, err := bubbly.NewComponent("CounterApp").
		Template(func(ctx bubbly.RenderContext) string {
			return "CounterApp"
		}).
		Build()
	require.NoError(t, err)
	
	// 3. Initialize parent (triggers OnComponentMount hook)
	parent.Init()
	
	// 4. Debug: Check if any components were added
	allComponents := dt.store.GetAllComponents()
	t.Logf("After parent.Init(), store has %d components", len(allComponents))
	for _, comp := range allComponents {
		t.Logf("  Component: %s (ID: %s, Status: %s)", comp.Name, comp.ID, comp.Status)
	}
	
	// 5. Verify parent was registered
	roots := dt.store.GetRootComponents()
	t.Logf("Root components: %d", len(roots))
	require.Len(t, roots, 1, "Parent should be registered as root")
	assert.Equal(t, "CounterApp", roots[0].Name)
	
	// 6. Now simulate Setup() creating children
	child1, err := bubbly.NewComponent("CounterDisplay").
		Template(func(ctx bubbly.RenderContext) string {
			return "CounterDisplay"
		}).
		Build()
	require.NoError(t, err)
	
	child2, err := bubbly.NewComponent("CounterControls").
		Template(func(ctx bubbly.RenderContext) string {
			return "CounterControls"
		}).
		Build()
	require.NoError(t, err)
	
	// 7. Initialize children first (triggers OnComponentMount for each)
	// This creates their snapshots in the store
	child1.Init()
	child2.Init()
	
	// 8. Now add parent-child relationships (simulates ctx.ExposeComponent())
	// This should trigger OnChildAdded hook which calls store.AddComponentChild
	dt.store.AddComponentChild(parent.ID(), child1.ID())
	dt.store.AddComponentChild(parent.ID(), child2.ID())
	
	// 9. Now check the tree - children should be visible
	roots = dt.store.GetRootComponents()
	require.Len(t, roots, 1)
	
	root := roots[0]
	t.Logf("Root component: %s", root.Name)
	t.Logf("Root children count: %d", len(root.Children))
	if len(root.Children) > 0 {
		for i, child := range root.Children {
			t.Logf("  Child %d: %s (ID: %s)", i, child.Name, child.ID)
		}
	}
	
	// CRITICAL ASSERTION: Children must be in the snapshot!
	assert.Len(t, root.Children, 2, "CRITICAL: Parent must have 2 children in snapshot")
	
	if len(root.Children) >= 2 {
		assert.Equal(t, "CounterDisplay", root.Children[0].Name)
		assert.Equal(t, "CounterControls", root.Children[1].Name)
	}
	
	// 10. Now simulate exposing refs (like app.go does)
	// Create some refs
	countRef := bubbly.NewRef(5)
	isEvenRef := bubbly.NewRef(false)
	
	// Get parent's context and expose refs
	// Note: In real app, this happens in Setup() with ctx.Expose()
	// For test, we'll manually register ref ownership
	countID := "ref-count-test"
	isEvenID := "ref-iseven-test"
	
	dt.store.RegisterRefOwner(parent.ID(), countID)
	dt.store.RegisterRefOwner(parent.ID(), isEvenID)
	
	// Update ref values in store
	dt.store.UpdateRefValue(countID, countRef.Get())
	dt.store.UpdateRefValue(isEvenID, isEvenRef.Get())
	
	// 11. Get updated root with refs
	roots = dt.store.GetRootComponents()
	require.Len(t, roots, 1)
	root = roots[0]
	
	t.Logf("Root refs count: %d", len(root.Refs))
	for i, ref := range root.Refs {
		t.Logf("  Ref %d: %s = %v (%s)", i, ref.Name, ref.Value, ref.Type)
	}
	
	// CRITICAL ASSERTION: Refs must be in the snapshot!
	assert.Len(t, root.Refs, 2, "CRITICAL: Parent must have 2 refs in snapshot")
	
	// 12. Set root in inspector UI
	dt.ui.inspector.SetRoot(root)
	
	// 13. Verify tree rendering shows children
	treeRoot := dt.ui.inspector.tree.GetRoot()
	require.NotNil(t, treeRoot)
	assert.Len(t, treeRoot.Children, 2, "Tree must show children")
	
	// 14. Render the tree
	rendered := dt.ui.inspector.tree.Render()
	t.Logf("Rendered tree:\n%s", rendered)
	
	assert.Contains(t, rendered, "CounterApp", "Tree must show parent")
	assert.Contains(t, rendered, "2 refs", "Tree must show ref count")
	// Note: Children only show when expanded
	
	// 15. Verify State tab shows refs
	stateContent := renderStateTab(root)
	t.Logf("State tab:\n%s", stateContent)
	
	assert.NotContains(t, stateContent, "No reactive state", "State tab must show refs")
	assert.Contains(t, stateContent, "count", "State tab must show count ref")
	assert.Contains(t, stateContent, "iseven", "State tab must show isEven ref")
}
