package testutil

import (
	"fmt"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a test component with a template for children tests
func createChildTestComponent(name string) (bubbly.Component, error) {
	return bubbly.NewComponent(name).
		Template(func(ctx bubbly.RenderContext) string {
			return fmt.Sprintf("%s view", name)
		}).
		Build()
}

// TestNewChildrenManagementTester tests creating a new ChildrenManagementTester
func TestNewChildrenManagementTester(t *testing.T) {
	// Create a parent component
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)

	// Create tester
	tester := NewChildrenManagementTester(parent)

	// Assert tester created
	assert.NotNil(t, tester)
	assert.NotNil(t, tester.parent)
	assert.NotNil(t, tester.children)
	assert.NotNil(t, tester.mounted)
	assert.NotNil(t, tester.unmounted)
	assert.Equal(t, parent, tester.parent)
	assert.Empty(t, tester.children)
}

// TestChildrenManagementTester_AddChild tests adding a single child
func TestChildrenManagementTester_AddChild(t *testing.T) {
	// Create parent
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	// Create child
	child, err := createChildTestComponent("Child")
	assert.NoError(t, err)

	// Create tester
	tester := NewChildrenManagementTester(parent)

	// Add child
	tester.AddChild(child)

	// Assert child added
	assert.Len(t, tester.children, 1)
	assert.Equal(t, child, tester.children[0])
	assert.True(t, tester.mounted[child], "Child should be marked as mounted")
}

// TestChildrenManagementTester_AddMultipleChildren tests adding multiple children
func TestChildrenManagementTester_AddMultipleChildren(t *testing.T) {
	// Create parent
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	// Create children
	child1, err := createChildTestComponent("Child1")
	assert.NoError(t, err)
	child2, err := createChildTestComponent("Child2")
	assert.NoError(t, err)
	child3, err := createChildTestComponent("Child3")
	assert.NoError(t, err)

	// Create tester
	tester := NewChildrenManagementTester(parent)

	// Add children
	tester.AddChild(child1)
	tester.AddChild(child2)
	tester.AddChild(child3)

	// Assert all children added
	assert.Len(t, tester.children, 3)
	assert.True(t, tester.mounted[child1])
	assert.True(t, tester.mounted[child2])
	assert.True(t, tester.mounted[child3])
}

// TestChildrenManagementTester_RemoveChild tests removing a child
func TestChildrenManagementTester_RemoveChild(t *testing.T) {
	// Create parent
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	// Create children
	child1, err := createChildTestComponent("Child1")
	assert.NoError(t, err)
	child2, err := createChildTestComponent("Child2")
	assert.NoError(t, err)

	// Create tester and add children
	tester := NewChildrenManagementTester(parent)
	tester.AddChild(child1)
	tester.AddChild(child2)

	// Remove child1
	tester.RemoveChild(child1)

	// Assert child1 removed and unmounted
	assert.Len(t, tester.children, 1)
	assert.Equal(t, child2, tester.children[0])
	assert.True(t, tester.unmounted[child1], "Child1 should be marked as unmounted")
	assert.False(t, tester.unmounted[child2], "Child2 should not be unmounted")
}

// TestChildrenManagementTester_AssertChildMounted tests mount assertion
func TestChildrenManagementTester_AssertChildMounted(t *testing.T) {
	// Create parent and child
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	child, err := createChildTestComponent("Child")
	assert.NoError(t, err)

	// Create tester and add child
	tester := NewChildrenManagementTester(parent)
	tester.AddChild(child)

	// Assert child mounted (should pass)
	tester.AssertChildMounted(t, child)
}

// TestChildrenManagementTester_AssertChildMounted_Failure tests mount assertion failure
func TestChildrenManagementTester_AssertChildMounted_Failure(t *testing.T) {
	// Create parent and child
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)

	child, err := createChildTestComponent("Child")
	assert.NoError(t, err)

	// Create tester (don't add child)
	tester := NewChildrenManagementTester(parent)

	// Create mock testing.T
	mockT := &mockTestingT{}

	// Assert child mounted (should fail)
	tester.AssertChildMounted(mockT, child)

	// Verify error was called
	assert.True(t, len(mockT.errors) > 0, "Expected error to be recorded")
	assert.Contains(t, mockT.errors[0], "not mounted")
}

// TestChildrenManagementTester_AssertChildUnmounted tests unmount assertion
func TestChildrenManagementTester_AssertChildUnmounted(t *testing.T) {
	// Create parent and child
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	child, err := createChildTestComponent("Child")
	assert.NoError(t, err)

	// Create tester, add and remove child
	tester := NewChildrenManagementTester(parent)
	tester.AddChild(child)
	tester.RemoveChild(child)

	// Assert child unmounted (should pass)
	tester.AssertChildUnmounted(t, child)
}

// TestChildrenManagementTester_AssertChildUnmounted_Failure tests unmount assertion failure
func TestChildrenManagementTester_AssertChildUnmounted_Failure(t *testing.T) {
	// Create parent and child
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	child, err := createChildTestComponent("Child")
	assert.NoError(t, err)

	// Create tester and add child (but don't remove)
	tester := NewChildrenManagementTester(parent)
	tester.AddChild(child)

	// Create mock testing.T
	mockT := &mockTestingT{}

	// Assert child unmounted (should fail)
	tester.AssertChildUnmounted(mockT, child)

	// Verify error was called
	assert.True(t, len(mockT.errors) > 0, "Expected error to be recorded")
	assert.Contains(t, mockT.errors[0], "not unmounted")
}

// TestChildrenManagementTester_AssertChildCount tests child count assertion
func TestChildrenManagementTester_AssertChildCount(t *testing.T) {
	tests := []struct {
		name          string
		childrenCount int
	}{
		{"zero children", 0},
		{"one child", 1},
		{"three children", 3},
		{"five children", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parent
			parent, err := createChildTestComponent("Parent")
			assert.NoError(t, err)
			parent.Init()

			// Create tester
			tester := NewChildrenManagementTester(parent)

			// Add children
			for i := 0; i < tt.childrenCount; i++ {
				child, err := createChildTestComponent("Child")
				assert.NoError(t, err)
				tester.AddChild(child)
			}

			// Assert count (should pass)
			tester.AssertChildCount(t, tt.childrenCount)
		})
	}
}

// TestChildrenManagementTester_AssertChildCount_Failure tests count assertion failure
func TestChildrenManagementTester_AssertChildCount_Failure(t *testing.T) {
	// Create parent
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	// Create tester with 2 children
	tester := NewChildrenManagementTester(parent)
	child1, err := createChildTestComponent("Child1")
	assert.NoError(t, err)
	child2, err := createChildTestComponent("Child2")
	assert.NoError(t, err)
	tester.AddChild(child1)
	tester.AddChild(child2)

	// Create mock testing.T
	mockT := &mockTestingT{}

	// Assert wrong count (should fail)
	tester.AssertChildCount(mockT, 5)

	// Verify error was called
	assert.True(t, len(mockT.errors) > 0, "Expected error to be recorded")
	assert.Contains(t, mockT.errors[0], "expected 5 children")
}

// TestChildrenManagementTester_ChildOrderPreserved tests that child order is maintained
func TestChildrenManagementTester_ChildOrderPreserved(t *testing.T) {
	// Create parent
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	// Create children with distinct names
	child1, err := createChildTestComponent("Child1")
	assert.NoError(t, err)
	child2, err := createChildTestComponent("Child2")
	assert.NoError(t, err)
	child3, err := createChildTestComponent("Child3")
	assert.NoError(t, err)

	// Create tester and add in specific order
	tester := NewChildrenManagementTester(parent)
	tester.AddChild(child1)
	tester.AddChild(child2)
	tester.AddChild(child3)

	// Assert order preserved
	assert.Equal(t, "Child1", tester.children[0].Name())
	assert.Equal(t, "Child2", tester.children[1].Name())
	assert.Equal(t, "Child3", tester.children[2].Name())
}

// TestChildrenManagementTester_DynamicUpdates tests dynamic add/remove operations
func TestChildrenManagementTester_DynamicUpdates(t *testing.T) {
	// Create parent
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	// Create children
	child1, err := createChildTestComponent("Child1")
	assert.NoError(t, err)
	child2, err := createChildTestComponent("Child2")
	assert.NoError(t, err)
	child3, err := createChildTestComponent("Child3")
	assert.NoError(t, err)

	// Create tester
	tester := NewChildrenManagementTester(parent)

	// Add child1
	tester.AddChild(child1)
	tester.AssertChildCount(t, 1)

	// Add child2
	tester.AddChild(child2)
	tester.AssertChildCount(t, 2)

	// Remove child1
	tester.RemoveChild(child1)
	tester.AssertChildCount(t, 1)
	assert.Equal(t, child2, tester.children[0])

	// Add child3
	tester.AddChild(child3)
	tester.AssertChildCount(t, 2)

	// Verify final state
	assert.Equal(t, child2, tester.children[0])
	assert.Equal(t, child3, tester.children[1])
	assert.True(t, tester.unmounted[child1])
	assert.True(t, tester.mounted[child2])
	assert.True(t, tester.mounted[child3])
}
