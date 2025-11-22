package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
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

// TestChildrenManagementTester_GetChildren tests retrieving all children
func TestChildrenManagementTester_GetChildren(t *testing.T) {
	tests := []struct {
		name          string
		childrenCount int
		verify        func(*testing.T, *ChildrenManagementTester)
	}{
		{
			name:          "no_children",
			childrenCount: 0,
			verify: func(t *testing.T, tester *ChildrenManagementTester) {
				children := tester.GetChildren()
				assert.Empty(t, children)
			},
		},
		{
			name:          "single_child",
			childrenCount: 1,
			verify: func(t *testing.T, tester *ChildrenManagementTester) {
				children := tester.GetChildren()
				assert.Len(t, children, 1)
			},
		},
		{
			name:          "multiple_children",
			childrenCount: 5,
			verify: func(t *testing.T, tester *ChildrenManagementTester) {
				children := tester.GetChildren()
				assert.Len(t, children, 5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent, err := createChildTestComponent("Parent")
			assert.NoError(t, err)
			parent.Init()

			tester := NewChildrenManagementTester(parent)

			// Add children
			for i := 0; i < tt.childrenCount; i++ {
				child, err := createChildTestComponent("Child")
				assert.NoError(t, err)
				tester.AddChild(child)
			}

			// Verify
			tt.verify(t, tester)
		})
	}
}

// TestChildrenManagementTester_GetChildren_AfterRemoval tests GetChildren after removing children
func TestChildrenManagementTester_GetChildren_AfterRemoval(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	// Add 3 children
	child1, _ := createChildTestComponent("Child1")
	child2, _ := createChildTestComponent("Child2")
	child3, _ := createChildTestComponent("Child3")

	tester.AddChild(child1)
	tester.AddChild(child2)
	tester.AddChild(child3)

	// Verify initial state
	children := tester.GetChildren()
	assert.Len(t, children, 3)

	// Remove middle child
	tester.RemoveChild(child2)

	// Verify updated state
	children = tester.GetChildren()
	assert.Len(t, children, 2)
	assert.Equal(t, child1, children[0])
	assert.Equal(t, child3, children[1])
}

// TestChildrenManagementTester_GetMountedChildren tests retrieving only mounted children
func TestChildrenManagementTester_GetMountedChildren(t *testing.T) {
	tests := []struct {
		name                 string
		setup                func(*ChildrenManagementTester) []bubbly.Component
		expectedMountedCount int
	}{
		{
			name: "no_children",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				return nil
			},
			expectedMountedCount: 0,
		},
		{
			name: "all_mounted",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")
				tester.AddChild(child1)
				tester.AddChild(child2)
				return []bubbly.Component{child1, child2}
			},
			expectedMountedCount: 2,
		},
		{
			name: "only_added_children",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")
				child3, _ := createChildTestComponent("Child3")

				tester.AddChild(child1)
				tester.AddChild(child2)
				tester.AddChild(child3)

				// All three should be in mounted map
				return []bubbly.Component{child1, child2, child3}
			},
			expectedMountedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent, err := createChildTestComponent("Parent")
			assert.NoError(t, err)
			parent.Init()

			tester := NewChildrenManagementTester(parent)

			// Setup
			expectedMounted := tt.setup(tester)

			// Get mounted children
			mounted := tester.GetMountedChildren()

			// Verify count
			assert.Len(t, mounted, tt.expectedMountedCount)

			// Verify correct children are marked as mounted
			for _, expected := range expectedMounted {
				found := false
				for _, actual := range mounted {
					if expected == actual {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected mounted child not found")
			}
		})
	}
}

// TestChildrenManagementTester_GetMountedChildren_MapIteration tests map-based iteration
func TestChildrenManagementTester_GetMountedChildren_MapIteration(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	// Add children in specific order
	child1, _ := createChildTestComponent("Child1")
	child2, _ := createChildTestComponent("Child2")
	child3, _ := createChildTestComponent("Child3")

	tester.AddChild(child1)
	tester.AddChild(child2)
	tester.AddChild(child3)

	// Get mounted
	mounted := tester.GetMountedChildren()

	// All three should be in mounted map
	assert.Len(t, mounted, 3)

	// Order might not be preserved (it's from a map)
	// Just verify all three are there
	mountedSet := make(map[bubbly.Component]bool)
	for _, c := range mounted {
		mountedSet[c] = true
	}

	assert.True(t, mountedSet[child1], "child1 should be mounted")
	assert.True(t, mountedSet[child2], "child2 should be mounted")
	assert.True(t, mountedSet[child3], "child3 should be mounted")
}

// TestChildrenManagementTester_GetMountedChildren_EmptyCase tests edge case with no mounts
func TestChildrenManagementTester_GetMountedChildren_EmptyCase(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	// Get mounted without adding any children
	mounted := tester.GetMountedChildren()

	// Should be empty
	assert.Empty(t, mounted)
}

// TestChildrenManagementTester_GetUnmountedChildren tests retrieving unmounted children
func TestChildrenManagementTester_GetUnmountedChildren(t *testing.T) {
	tests := []struct {
		name                   string
		setup                  func(*ChildrenManagementTester) []bubbly.Component
		expectedUnmountedCount int
	}{
		{
			name: "no_children",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				return nil
			},
			expectedUnmountedCount: 0,
		},
		{
			name: "all_mounted_none_unmounted",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")
				tester.AddChild(child1)
				tester.AddChild(child2)
				return nil
			},
			expectedUnmountedCount: 0,
		},
		{
			name: "mixed_mounted_and_unmounted",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")
				child3, _ := createChildTestComponent("Child3")

				tester.AddChild(child1)
				tester.AddChild(child2)
				tester.AddChild(child3)

				// Unmount child2
				tester.RemoveChild(child2)

				return []bubbly.Component{child2}
			},
			expectedUnmountedCount: 1,
		},
		{
			name: "all_unmounted",
			setup: func(tester *ChildrenManagementTester) []bubbly.Component {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")

				tester.AddChild(child1)
				tester.AddChild(child2)

				tester.RemoveChild(child1)
				tester.RemoveChild(child2)

				return []bubbly.Component{child1, child2}
			},
			expectedUnmountedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent, err := createChildTestComponent("Parent")
			assert.NoError(t, err)
			parent.Init()

			tester := NewChildrenManagementTester(parent)
			expectedUnmounted := tt.setup(tester)

			unmounted := tester.GetUnmountedChildren()
			assert.Len(t, unmounted, tt.expectedUnmountedCount)

			for _, expected := range expectedUnmounted {
				found := false
				for _, actual := range unmounted {
					if expected == actual {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected unmounted child not found")
			}
		})
	}
}

// TestChildrenManagementTester_IsMounted tests checking mount status
func TestChildrenManagementTester_IsMounted(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	child1, _ := createChildTestComponent("Child1")
	child2, _ := createChildTestComponent("Child2")
	child3, _ := createChildTestComponent("Child3")

	// Initially, nothing is mounted
	assert.False(t, tester.IsMounted(child1))
	assert.False(t, tester.IsMounted(child2))
	assert.False(t, tester.IsMounted(child3))

	// Add child1 and child2
	tester.AddChild(child1)
	tester.AddChild(child2)

	// Now child1 and child2 are mounted
	assert.True(t, tester.IsMounted(child1))
	assert.True(t, tester.IsMounted(child2))
	assert.False(t, tester.IsMounted(child3))

	// Remove child1
	tester.RemoveChild(child1)

	// child1 still in mounted map (RemoveChild doesn't remove from mounted)
	assert.True(t, tester.IsMounted(child1))
	assert.True(t, tester.IsMounted(child2))
}

// TestChildrenManagementTester_IsUnmounted tests checking unmount status
func TestChildrenManagementTester_IsUnmounted(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	child1, _ := createChildTestComponent("Child1")
	child2, _ := createChildTestComponent("Child2")

	// Initially, nothing is unmounted
	assert.False(t, tester.IsUnmounted(child1))
	assert.False(t, tester.IsUnmounted(child2))

	// Add both
	tester.AddChild(child1)
	tester.AddChild(child2)

	// Still not unmounted
	assert.False(t, tester.IsUnmounted(child1))
	assert.False(t, tester.IsUnmounted(child2))

	// Remove child1
	tester.RemoveChild(child1)

	// Now child1 is unmounted
	assert.True(t, tester.IsUnmounted(child1))
	assert.False(t, tester.IsUnmounted(child2))
}

// TestChildrenManagementTester_GetChildByName tests finding children by name
func TestChildrenManagementTester_GetChildByName(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	child1, _ := createChildTestComponent("Alice")
	child2, _ := createChildTestComponent("Bob")
	child3, _ := createChildTestComponent("Charlie")

	tester.AddChild(child1)
	tester.AddChild(child2)
	tester.AddChild(child3)

	// Find existing children
	found := tester.GetChildByName("Alice")
	assert.NotNil(t, found)
	assert.Equal(t, "Alice", found.Name())

	found = tester.GetChildByName("Bob")
	assert.NotNil(t, found)
	assert.Equal(t, "Bob", found.Name())

	found = tester.GetChildByName("Charlie")
	assert.NotNil(t, found)
	assert.Equal(t, "Charlie", found.Name())

	// Find non-existent child
	found = tester.GetChildByName("David")
	assert.Nil(t, found)

	// Find after removal
	tester.RemoveChild(child2)
	found = tester.GetChildByName("Bob")
	assert.Nil(t, found)
}

// TestChildrenManagementTester_GetChildByID tests finding children by ID
func TestChildrenManagementTester_GetChildByID(t *testing.T) {
	parent, err := createChildTestComponent("Parent")
	assert.NoError(t, err)
	parent.Init()

	tester := NewChildrenManagementTester(parent)

	child1, _ := createChildTestComponent("Child1")
	child2, _ := createChildTestComponent("Child2")
	child3, _ := createChildTestComponent("Child3")

	tester.AddChild(child1)
	tester.AddChild(child2)
	tester.AddChild(child3)

	// Find by ID
	id1 := child1.ID()
	found := tester.GetChildByID(id1)
	assert.NotNil(t, found)
	assert.Equal(t, id1, found.ID())

	id2 := child2.ID()
	found = tester.GetChildByID(id2)
	assert.NotNil(t, found)
	assert.Equal(t, id2, found.ID())

	// Find non-existent ID
	found = tester.GetChildByID("non-existent-id")
	assert.Nil(t, found)

	// Find after removal
	tester.RemoveChild(child2)
	found = tester.GetChildByID(id2)
	assert.Nil(t, found)
}

// TestChildrenManagementTester_String tests string representation
func TestChildrenManagementTester_String(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*ChildrenManagementTester)
		expectedFields map[string]bool
	}{
		{
			name: "no_children",
			setup: func(tester *ChildrenManagementTester) {
				// No setup
			},
			expectedFields: map[string]bool{
				"ChildrenManagementTester": true,
				"parent=Parent":            true,
				"children=0":               true,
				"mounted=0":                true,
				"unmounted=0":              true,
			},
		},
		{
			name: "with_children",
			setup: func(tester *ChildrenManagementTester) {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")
				tester.AddChild(child1)
				tester.AddChild(child2)
			},
			expectedFields: map[string]bool{
				"ChildrenManagementTester": true,
				"parent=Parent":            true,
				"children=2":               true,
				"mounted=2":                true,
			},
		},
		{
			name: "with_unmounted",
			setup: func(tester *ChildrenManagementTester) {
				child1, _ := createChildTestComponent("Child1")
				child2, _ := createChildTestComponent("Child2")
				tester.AddChild(child1)
				tester.AddChild(child2)
				tester.RemoveChild(child1)
			},
			expectedFields: map[string]bool{
				"ChildrenManagementTester": true,
				"parent=Parent":            true,
				"children=1":               true,
				"unmounted=1":              true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent, err := createChildTestComponent("Parent")
			assert.NoError(t, err)
			parent.Init()

			tester := NewChildrenManagementTester(parent)
			tt.setup(tester)

			str := tester.String()
			assert.NotEmpty(t, str)

			// Verify expected fields are in string
			for field := range tt.expectedFields {
				assert.Contains(t, str, field)
			}
		})
	}
}
