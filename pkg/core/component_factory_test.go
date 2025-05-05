package core

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestComponentFactory tests the component factory functions
func TestComponentFactory(t *testing.T) {
	t.Run("Create Basic Component", func(t *testing.T) {
		// Test creating a basic component with factory function
		id := "test-component"
		component := CreateComponent(id)

		assert.NotNil(t, component, "Component should not be nil")
		assert.Equal(t, id, component.ID(), "Component ID should match provided ID")
		assert.Empty(t, component.Children(), "New component should have no children")
	})

	t.Run("Create Component With Options", func(t *testing.T) {
		// Test component creation with various configuration options
		id := "test-component-with-options"

		// Create a render function for testing
		customRender := func(c Component) string {
			return "Custom Rendered " + c.ID()
		}

		// Create component with options
		component := CreateComponent(
			id,
			WithRender(customRender),
			WithInitialChildren(
				CreateComponent("child-1"),
				CreateComponent("child-2"),
			),
		)

		assert.NotNil(t, component, "Component should not be nil")
		assert.Equal(t, id, component.ID(), "Component ID should match provided ID")
		assert.Len(t, component.Children(), 2, "Component should have 2 children")
		assert.Equal(t, "Custom Rendered test-component-with-options", component.Render(), "Custom render function should be used")
	})

	t.Run("Create Stateful Component", func(t *testing.T) {
		// Test creating a stateful component
		id := "stateful-component"
		name := "Test Stateful Component"

		component := CreateStatefulComponent(id, name)

		assert.NotNil(t, component, "Stateful component should not be nil")
		assert.Equal(t, id, component.ID(), "Component ID should match provided ID")
		assert.NotNil(t, component.GetState(), "Stateful component should have a state")
		assert.False(t, component.IsMounted(), "New component should not be mounted yet")
	})

	t.Run("Create Stateful Component With Options", func(t *testing.T) {
		// Test creating a stateful component with options
		id := "stateful-component-with-options"
		name := "Test Stateful Component With Options"

		// Custom update function for testing
		customUpdate := func(c Component, msg tea.Msg) (tea.Cmd, error) {
			return tea.Batch(tea.Println("Custom update called")), nil
		}

		// Create stateful component with options
		component := CreateStatefulComponent(
			id,
			name,
			WithUpdate(customUpdate),
			WithMounted(true),
		)

		assert.NotNil(t, component, "Stateful component should not be nil")
		assert.Equal(t, id, component.ID(), "Component ID should match provided ID")
		assert.True(t, component.IsMounted(), "Component should be mounted as specified in options")

		// Test custom update function
		cmd, err := component.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.NoError(t, err, "Update should not return error")
		assert.NotNil(t, cmd, "Custom update should return a command")
	})
}

// TestComponentConfiguration tests component configuration options
func TestComponentConfiguration(t *testing.T) {
	t.Run("Apply Config Options", func(t *testing.T) {
		// Create a basic component
		component := CreateComponent("config-test")

		// Define custom lifecycle functions for testing
		initCalled := false
		CustomInit := func(c Component) error {
			initCalled = true
			return nil
		}

		disposeCalled := false
		CustomDispose := func(c Component) error {
			disposeCalled = true
			return nil
		}

		// Apply configuration options
		configurator := ComponentConfigurator{
			WithInit(CustomInit),
			WithDispose(CustomDispose),
		}
		configurator.ApplyTo(component)

		// Test that the configurations are applied
		_ = component.Initialize()
		assert.True(t, initCalled, "Custom init function should be called")

		_ = component.Dispose()
		assert.True(t, disposeCalled, "Custom dispose function should be called")
	})
}

// TestComponentTreeOperations tests component tree operations
func TestComponentTreeOperations(t *testing.T) {
	t.Run("Find Component By ID", func(t *testing.T) {
		// Create a component tree
		root := CreateComponent("root")
		child1 := CreateComponent("child-1")
		child2 := CreateComponent("child-2")
		grandchild := CreateComponent("grandchild")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild)

		// Test finding components by ID
		found := FindComponentByID(root, "child-1")
		assert.NotNil(t, found, "Should find existing component")
		assert.Equal(t, "child-1", found.ID(), "Found component should have correct ID")

		found = FindComponentByID(root, "grandchild")
		assert.NotNil(t, found, "Should find nested component")
		assert.Equal(t, "grandchild", found.ID(), "Found component should have correct ID")

		found = FindComponentByID(root, "nonexistent")
		assert.Nil(t, found, "Should return nil for nonexistent component")
	})

	t.Run("Tree Traversal", func(t *testing.T) {
		// Create a component tree
		root := CreateComponent("root")
		child1 := CreateComponent("child-1")
		child2 := CreateComponent("child-2")
		grandchild1 := CreateComponent("grandchild-1")
		grandchild2 := CreateComponent("grandchild-2")

		root.AddChild(child1)
		root.AddChild(child2)
		child1.AddChild(grandchild1)
		child2.AddChild(grandchild2)

		// Test depth-first traversal
		visited := []string{}
		TraverseComponentTree(root, func(c Component) {
			visited = append(visited, c.ID())
		})

		expected := []string{"root", "child-1", "grandchild-1", "child-2", "grandchild-2"}
		assert.Equal(t, expected, visited, "Depth-first traversal should visit components in correct order")
	})
}

// TestComponentReconciliation tests component reconciliation with keys
func TestComponentReconciliation(t *testing.T) {
	t.Run("Component Keys and Reconciliation", func(t *testing.T) {
		// Create components with keys
		c1 := CreateComponent("c1", WithKey("key1"))
		c2 := CreateComponent("c2", WithKey("key2"))
		c3 := CreateComponent("c3", WithKey("key3"))

		// Create reconciler and add components
		reconciler := NewComponentReconciler()
		reconciler.AddComponents(c1, c2, c3)

		// Test key-based lookup
		found := reconciler.FindByKey("key2")
		assert.NotNil(t, found, "Should find component by key")
		assert.Equal(t, "c2", found.ID(), "Found component should have correct ID")

		// Test reconciliation with new components
		newC1 := CreateComponent("new-c1", WithKey("key1"))
		// Create another component but don't include it in reconciliation
		newC4 := CreateComponent("new-c4", WithKey("key4"))

		newComponents := []Component{newC1, newC4, c2}
		result := reconciler.Reconcile(newComponents)

		// Check reconciliation results
		assert.Len(t, result.Reused, 2, "Should reuse 2 components")
		assert.Len(t, result.Added, 1, "Should add 1 new component")
		assert.Len(t, result.Removed, 1, "Should remove 1 old component")

		// Check that key1 was reused but has new ID
		assert.Equal(t, "new-c1", result.Reused[0].ID(), "Component with key1 should have new ID")

		// Check that key3 was removed
		assert.Equal(t, "c3", result.Removed[0].ID(), "Component with key3 should be removed")

		// Check that key4 was added
		assert.Equal(t, "new-c4", result.Added[0].ID(), "Component with key4 should be added")
	})
}
