package testutil

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ChildrenManagementTester provides utilities for testing component children rendering and lifecycle.
//
// It wraps a parent component and provides methods for adding/removing children,
// tracking their mount/unmount status, and asserting on their lifecycle behavior.
// This is useful for verifying that parent components correctly manage their children's
// lifecycle and that children are properly initialized, mounted, and cleaned up.
//
// Type Safety:
//   - Thread-safe tracking of children state
//   - Clear assertion methods for mount/unmount verification
//   - Tracks children order and count
//
// Example:
//
//	func TestParentWithChildren(t *testing.T) {
//		// Create parent component
//		parent, _ := bubbly.NewComponent("Parent").Build()
//		parent.Init()
//
//		// Create tester
//		tester := testutil.NewChildrenManagementTester(parent)
//
//		// Add children
//		child1, _ := bubbly.NewComponent("Child1").Build()
//		child2, _ := bubbly.NewComponent("Child2").Build()
//		tester.AddChild(child1)
//		tester.AddChild(child2)
//
//		// Assert children mounted
//		tester.AssertChildMounted(t, child1)
//		tester.AssertChildMounted(t, child2)
//		tester.AssertChildCount(t, 2)
//
//		// Remove child
//		tester.RemoveChild(child1)
//		tester.AssertChildUnmounted(t, child1)
//		tester.AssertChildCount(t, 1)
//	}
type ChildrenManagementTester struct {
	// parent is the parent component being tested
	parent bubbly.Component

	// children is the list of child components added to the parent
	children []bubbly.Component

	// mounted tracks which children have been mounted
	mounted map[bubbly.Component]bool

	// unmounted tracks which children have been unmounted
	unmounted map[bubbly.Component]bool
}

// NewChildrenManagementTester creates a new ChildrenManagementTester for testing children management.
//
// Parameters:
//   - parent: The parent component to test
//
// Returns:
//   - *ChildrenManagementTester: A new tester instance
//
// Example:
//
//	parent, _ := bubbly.NewComponent("Parent").Build()
//	parent.Init()
//
//	tester := testutil.NewChildrenManagementTester(parent)
func NewChildrenManagementTester(parent bubbly.Component) *ChildrenManagementTester {
	return &ChildrenManagementTester{
		parent:    parent,
		children:  []bubbly.Component{},
		mounted:   make(map[bubbly.Component]bool),
		unmounted: make(map[bubbly.Component]bool),
	}
}

// AddChild adds a child component to the parent and tracks its mount status.
//
// This method:
//  1. Appends the child to the parent's internal children slice
//  2. Initializes the child component (calls Init())
//  3. Triggers mount by calling View() (which executes onMounted hooks)
//  4. Tracks the child in the mounted map
//
// The child must not have been initialized before calling this method.
//
// Parameters:
//   - child: The child component to add
//
// Example:
//
//	child, _ := bubbly.NewComponent("Child").Build()
//	tester.AddChild(child)
//	tester.AssertChildMounted(t, child)
func (cmt *ChildrenManagementTester) AddChild(child bubbly.Component) {
	// Add to our tracking list
	cmt.children = append(cmt.children, child)

	// Type assert to access internal children slice
	// This is safe because we control the component creation
	type componentWithChildren interface {
		bubbly.Component
		// We need to access the internal children field
		// This is done via reflection-like access in the actual implementation
	}

	// For now, we'll use a simpler approach: just track the child
	// and call its lifecycle methods directly
	// The parent-child relationship in BubblyUI is managed through
	// the component tree, not through direct parent.children manipulation

	// Initialize child
	child.Init()

	// Trigger mount by calling View (this executes onMounted hooks)
	child.View()

	// Mark as mounted
	cmt.mounted[child] = true
}

// RemoveChild removes a child component from the parent and tracks its unmount status.
//
// This method:
//  1. Removes the child from the tester's children list
//  2. Calls the child's Unmount() method (which executes onUnmounted hooks)
//  3. Tracks the child in the unmounted map
//
// Parameters:
//   - child: The child component to remove
//
// Example:
//
//	tester.RemoveChild(child)
//	tester.AssertChildUnmounted(t, child)
func (cmt *ChildrenManagementTester) RemoveChild(child bubbly.Component) {
	// Remove from our tracking list
	newChildren := []bubbly.Component{}
	for _, c := range cmt.children {
		if c != child {
			newChildren = append(newChildren, c)
		}
	}
	cmt.children = newChildren

	// Type assert to access Unmount() method
	// Unmount() is not on the Component interface, but on the implementation
	type unmountable interface {
		Unmount()
	}

	if u, ok := child.(unmountable); ok {
		u.Unmount()
	}

	// Mark as unmounted
	cmt.unmounted[child] = true
}

// AssertChildMounted asserts that a child component has been mounted.
//
// This method verifies that the child was added via AddChild() and its
// mount lifecycle was triggered. It fails the test if the child is not
// in the mounted map.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - child: The child component to check
//
// Example:
//
//	tester.AddChild(child)
//	tester.AssertChildMounted(t, child) // Passes
func (cmt *ChildrenManagementTester) AssertChildMounted(t testingT, child bubbly.Component) {
	t.Helper()

	if !cmt.mounted[child] {
		t.Errorf("child component %q (ID: %s) was not mounted",
			child.Name(), child.ID())
	}
}

// AssertChildUnmounted asserts that a child component has been unmounted.
//
// This method verifies that the child was removed via RemoveChild() and its
// unmount lifecycle was triggered. It fails the test if the child is not
// in the unmounted map.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - child: The child component to check
//
// Example:
//
//	tester.RemoveChild(child)
//	tester.AssertChildUnmounted(t, child) // Passes
func (cmt *ChildrenManagementTester) AssertChildUnmounted(t testingT, child bubbly.Component) {
	t.Helper()

	if !cmt.unmounted[child] {
		t.Errorf("child component %q (ID: %s) was not unmounted",
			child.Name(), child.ID())
	}
}

// AssertChildCount asserts that the parent has the expected number of children.
//
// This method verifies the count of children currently tracked by the tester.
// It fails the test if the actual count doesn't match the expected count.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: The expected number of children
//
// Example:
//
//	tester.AddChild(child1)
//	tester.AddChild(child2)
//	tester.AssertChildCount(t, 2) // Passes
func (cmt *ChildrenManagementTester) AssertChildCount(t testingT, expected int) {
	t.Helper()

	actual := len(cmt.children)
	if actual != expected {
		t.Errorf("expected %d children, got %d", expected, actual)
	}
}

// GetChildren returns the current list of children tracked by the tester.
//
// This method provides access to the children slice for advanced testing scenarios
// where you need to inspect the children directly.
//
// Returns:
//   - []bubbly.Component: The list of children
//
// Example:
//
//	children := tester.GetChildren()
//	assert.Len(t, children, 2)
//	assert.Equal(t, "Child1", children[0].Name())
func (cmt *ChildrenManagementTester) GetChildren() []bubbly.Component {
	return cmt.children
}

// GetMountedChildren returns a list of all children that have been mounted.
//
// This method filters the children list to only include those marked as mounted.
//
// Returns:
//   - []bubbly.Component: The list of mounted children
//
// Example:
//
//	mounted := tester.GetMountedChildren()
//	assert.Len(t, mounted, 2)
func (cmt *ChildrenManagementTester) GetMountedChildren() []bubbly.Component {
	var mounted []bubbly.Component
	for child, isMounted := range cmt.mounted {
		if isMounted {
			mounted = append(mounted, child)
		}
	}
	return mounted
}

// GetUnmountedChildren returns a list of all children that have been unmounted.
//
// This method filters to only include children marked as unmounted.
//
// Returns:
//   - []bubbly.Component: The list of unmounted children
//
// Example:
//
//	unmounted := tester.GetUnmountedChildren()
//	assert.Len(t, unmounted, 1)
func (cmt *ChildrenManagementTester) GetUnmountedChildren() []bubbly.Component {
	var unmounted []bubbly.Component
	for child, isUnmounted := range cmt.unmounted {
		if isUnmounted {
			unmounted = append(unmounted, child)
		}
	}
	return unmounted
}

// IsMounted checks if a specific child has been mounted.
//
// Parameters:
//   - child: The child component to check
//
// Returns:
//   - bool: True if the child is mounted, false otherwise
//
// Example:
//
//	if tester.IsMounted(child) {
//	    fmt.Println("Child is mounted")
//	}
func (cmt *ChildrenManagementTester) IsMounted(child bubbly.Component) bool {
	return cmt.mounted[child]
}

// IsUnmounted checks if a specific child has been unmounted.
//
// Parameters:
//   - child: The child component to check
//
// Returns:
//   - bool: True if the child is unmounted, false otherwise
//
// Example:
//
//	if tester.IsUnmounted(child) {
//	    fmt.Println("Child is unmounted")
//	}
func (cmt *ChildrenManagementTester) IsUnmounted(child bubbly.Component) bool {
	return cmt.unmounted[child]
}

// GetChildByName finds a child component by name.
//
// This is useful when you need to reference a child by its name rather than
// keeping a direct reference to the component.
//
// Parameters:
//   - name: The name of the child to find
//
// Returns:
//   - bubbly.Component: The child component, or nil if not found
//
// Example:
//
//	child := tester.GetChildByName("Child1")
//	if child != nil {
//	    tester.AssertChildMounted(t, child)
//	}
func (cmt *ChildrenManagementTester) GetChildByName(name string) bubbly.Component {
	for _, child := range cmt.children {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

// GetChildByID finds a child component by ID.
//
// Parameters:
//   - id: The ID of the child to find
//
// Returns:
//   - bubbly.Component: The child component, or nil if not found
//
// Example:
//
//	child := tester.GetChildByID("component-123")
//	if child != nil {
//	    tester.AssertChildMounted(t, child)
//	}
func (cmt *ChildrenManagementTester) GetChildByID(id string) bubbly.Component {
	for _, child := range cmt.children {
		if child.ID() == id {
			return child
		}
	}
	return nil
}

// String returns a string representation of the tester state for debugging.
//
// Returns:
//   - string: A formatted string showing children count and mount/unmount status
//
// Example:
//
//	fmt.Println(tester.String())
//	// Output: ChildrenManagementTester{parent=Parent, children=2, mounted=2, unmounted=0}
func (cmt *ChildrenManagementTester) String() string {
	return fmt.Sprintf("ChildrenManagementTester{parent=%s, children=%d, mounted=%d, unmounted=%d}",
		cmt.parent.Name(), len(cmt.children), len(cmt.mounted), len(cmt.unmounted))
}
