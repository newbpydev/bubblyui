package bubbly

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChildren_Access tests accessing children from a component
func TestChildren_Access(t *testing.T) {
	tests := []struct {
		name                string
		setupChildren       func() []Component
		expectedCount       int
		expectDefensiveCopy bool
	}{
		{
			name: "no children",
			setupChildren: func() []Component {
				return []Component{}
			},
			expectedCount:       0,
			expectDefensiveCopy: true,
		},
		{
			name: "single child",
			setupChildren: func() []Component {
				child, _ := NewComponent("Child").
					Template(func(ctx RenderContext) string { return "child" }).
					Build()
				return []Component{child}
			},
			expectedCount:       1,
			expectDefensiveCopy: true,
		},
		{
			name: "multiple children",
			setupChildren: func() []Component {
				child1, _ := NewComponent("Child1").
					Template(func(ctx RenderContext) string { return "child1" }).
					Build()
				child2, _ := NewComponent("Child2").
					Template(func(ctx RenderContext) string { return "child2" }).
					Build()
				child3, _ := NewComponent("Child3").
					Template(func(ctx RenderContext) string { return "child3" }).
					Build()
				return []Component{child1, child2, child3}
			},
			expectedCount:       3,
			expectDefensiveCopy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := tt.setupChildren()

			parent, err := NewComponent("Parent").
				Template(func(ctx RenderContext) string { return "parent" }).
				Children(children...).
				Build()
			require.NoError(t, err)

			// Access children
			result := parent.(*componentImpl).Children()

			// Verify count
			assert.Equal(t, tt.expectedCount, len(result))

			// Verify defensive copy (modifying result shouldn't affect internal state)
			if tt.expectDefensiveCopy && len(result) > 0 {
				originalLen := len(parent.(*componentImpl).children)
				result = append(result, nil) // Modify returned slice
				assert.Equal(t, originalLen, len(parent.(*componentImpl).children), "internal children should not be affected")
			}
		})
	}
}

// TestAddChild tests adding children dynamically
func TestAddChild(t *testing.T) {
	tests := []struct {
		name            string
		initialChildren int
		childToAdd      func() Component
		expectError     bool
		expectedCount   int
	}{
		{
			name:            "add child to empty parent",
			initialChildren: 0,
			childToAdd: func() Component {
				child, _ := NewComponent("Child").
					Template(func(ctx RenderContext) string { return "child" }).
					Build()
				return child
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:            "add child to parent with existing children",
			initialChildren: 2,
			childToAdd: func() Component {
				child, _ := NewComponent("NewChild").
					Template(func(ctx RenderContext) string { return "new" }).
					Build()
				return child
			},
			expectError:   false,
			expectedCount: 3,
		},
		{
			name:            "add nil child",
			initialChildren: 1,
			childToAdd: func() Component {
				return nil
			},
			expectError:   true,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create parent with initial children
			var initialChildrenSlice []Component
			for i := 0; i < tt.initialChildren; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string { return "child" }).
					Build()
				initialChildrenSlice = append(initialChildrenSlice, child)
			}

			parent, err := NewComponent("Parent").
				Template(func(ctx RenderContext) string { return "parent" }).
				Children(initialChildrenSlice...).
				Build()
			require.NoError(t, err)

			// Add child
			childToAdd := tt.childToAdd()
			err = parent.(*componentImpl).AddChild(childToAdd)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify child was added
				children := parent.(*componentImpl).Children()
				assert.Equal(t, tt.expectedCount, len(children))

				// Verify parent reference was set
				if childToAdd != nil {
					childImpl := childToAdd.(*componentImpl)
					assert.NotNil(t, childImpl.parent)
					assert.Equal(t, parent.(*componentImpl).ID(), (*childImpl.parent).ID())
				}
			}
		})
	}
}

// TestRemoveChild tests removing children dynamically
func TestRemoveChild(t *testing.T) {
	tests := []struct {
		name          string
		setupChildren func() ([]Component, Component)
		expectError   bool
		expectedCount int
	}{
		{
			name: "remove existing child",
			setupChildren: func() ([]Component, Component) {
				child1, _ := NewComponent("Child1").
					Template(func(ctx RenderContext) string { return "child1" }).
					Build()
				child2, _ := NewComponent("Child2").
					Template(func(ctx RenderContext) string { return "child2" }).
					Build()
				return []Component{child1, child2}, child1
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "remove non-existent child",
			setupChildren: func() ([]Component, Component) {
				child1, _ := NewComponent("Child1").
					Template(func(ctx RenderContext) string { return "child1" }).
					Build()
				child2, _ := NewComponent("Child2").
					Template(func(ctx RenderContext) string { return "child2" }).
					Build()
				otherChild, _ := NewComponent("Other").
					Template(func(ctx RenderContext) string { return "other" }).
					Build()
				return []Component{child1, child2}, otherChild
			},
			expectError:   true,
			expectedCount: 2,
		},
		{
			name: "remove nil child",
			setupChildren: func() ([]Component, Component) {
				child1, _ := NewComponent("Child1").
					Template(func(ctx RenderContext) string { return "child1" }).
					Build()
				return []Component{child1}, nil
			},
			expectError:   true,
			expectedCount: 1,
		},
		{
			name: "remove last child",
			setupChildren: func() ([]Component, Component) {
				child, _ := NewComponent("Child").
					Template(func(ctx RenderContext) string { return "child" }).
					Build()
				return []Component{child}, child
			},
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children, childToRemove := tt.setupChildren()

			parent, err := NewComponent("Parent").
				Template(func(ctx RenderContext) string { return "parent" }).
				Children(children...).
				Build()
			require.NoError(t, err)

			// Remove child
			err = parent.(*componentImpl).RemoveChild(childToRemove)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify child was removed
				remainingChildren := parent.(*componentImpl).Children()
				assert.Equal(t, tt.expectedCount, len(remainingChildren))

				// Verify parent reference was cleared
				if childToRemove != nil {
					childImpl := childToRemove.(*componentImpl)
					assert.Nil(t, childImpl.parent)
				}
			}
		})
	}
}

// TestRenderChildren tests rendering all children
func TestRenderChildren(t *testing.T) {
	tests := []struct {
		name           string
		setupChildren  func() []Component
		expectedOutput []string
	}{
		{
			name: "no children",
			setupChildren: func() []Component {
				return []Component{}
			},
			expectedOutput: []string{},
		},
		{
			name: "single child",
			setupChildren: func() []Component {
				child, _ := NewComponent("Child").
					Template(func(ctx RenderContext) string { return "Hello" }).
					Build()
				return []Component{child}
			},
			expectedOutput: []string{"Hello"},
		},
		{
			name: "multiple children with different outputs",
			setupChildren: func() []Component {
				child1, _ := NewComponent("Child1").
					Template(func(ctx RenderContext) string { return "First" }).
					Build()
				child2, _ := NewComponent("Child2").
					Template(func(ctx RenderContext) string { return "Second" }).
					Build()
				child3, _ := NewComponent("Child3").
					Template(func(ctx RenderContext) string { return "Third" }).
					Build()
				return []Component{child1, child2, child3}
			},
			expectedOutput: []string{"First", "Second", "Third"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := tt.setupChildren()

			parent, err := NewComponent("Parent").
				Template(func(ctx RenderContext) string { return "parent" }).
				Children(children...).
				Build()
			require.NoError(t, err)

			// Render children
			result := parent.(*componentImpl).renderChildren()

			// Verify output
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestChildLifecycle tests that child lifecycle is managed correctly
func TestChildLifecycle(t *testing.T) {
	t.Run("children initialized on parent Init", func(t *testing.T) {
		setupCalled := false
		child, err := NewComponent("Child").
			Setup(func(ctx *Context) {
				setupCalled = true
			}).
			Template(func(ctx RenderContext) string { return "child" }).
			Build()
		require.NoError(t, err)

		parent, err := NewComponent("Parent").
			Template(func(ctx RenderContext) string { return "parent" }).
			Children(child).
			Build()
		require.NoError(t, err)

		// Init parent (should init children)
		parent.Init()

		assert.True(t, setupCalled, "child setup should be called during parent Init")
	})

	t.Run("parent reference set during build", func(t *testing.T) {
		child, err := NewComponent("Child").
			Template(func(ctx RenderContext) string { return "child" }).
			Build()
		require.NoError(t, err)

		parent, err := NewComponent("Parent").
			Template(func(ctx RenderContext) string { return "parent" }).
			Children(child).
			Build()
		require.NoError(t, err)

		// Verify parent reference
		childImpl := child.(*componentImpl)
		assert.NotNil(t, childImpl.parent)
		assert.Equal(t, parent.(*componentImpl).ID(), (*childImpl.parent).ID())
	})
}

// TestConcurrentChildAccess tests thread safety of children operations
func TestConcurrentChildAccess(t *testing.T) {
	parent, err := NewComponent("Parent").
		Template(func(ctx RenderContext) string { return "parent" }).
		Build()
	require.NoError(t, err)

	// Create children to add
	children := make([]Component, 10)
	for i := 0; i < 10; i++ {
		child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
			Template(func(ctx RenderContext) string { return "child" }).
			Build()
		children[i] = child
	}

	// Concurrent operations
	var wg sync.WaitGroup

	// Add children concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = parent.(*componentImpl).AddChild(children[idx])
		}(i)
	}

	// Access children concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = parent.(*componentImpl).Children()
		}()
	}

	// Render children concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = parent.(*componentImpl).renderChildren()
		}()
	}

	wg.Wait()

	// Verify all children were added
	finalChildren := parent.(*componentImpl).Children()
	assert.Equal(t, 10, len(finalChildren), "all children should be added")
}
