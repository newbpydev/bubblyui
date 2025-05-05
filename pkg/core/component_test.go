package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestComponentInterface verifies that the Component interface has the expected methods
func TestComponentInterface(t *testing.T) {
	// This is a compile-time test to ensure Component has all required methods
	var _ Component = (*MockComponent)(nil)
	var _ Component = (*BaseComponent)(nil)
}

// MockComponent is a test implementation of the Component interface
type MockComponent struct {
	initialized       bool
	renderedCalled    bool
	rendered          string
	disposed          bool
	id                string
	children          []Component
	updateCallCount   int
	updateReturnCmd   tea.Cmd
	updateReturnError error
	initializeError   error
	disposeError      error
	handledMsgs       []tea.Msg
}

// Initialize implements the Component interface
func (m *MockComponent) Initialize() error {
	m.initialized = true

	// Check for error before initializing children
	if m.initializeError != nil {
		return m.initializeError
	}

	// Initialize children
	for _, child := range m.children {
		if err := child.Initialize(); err != nil {
			return fmt.Errorf("error initializing child %s: %w", child.ID(), err)
		}
	}

	return nil
}

// Update implements the Component interface
func (m *MockComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	m.updateCallCount++
	m.handledMsgs = append(m.handledMsgs, msg)

	// Propagate update to children
	cmds := make([]tea.Cmd, 0)
	for _, child := range m.children {
		cmd, err := child.Update(msg)
		if err != nil {
			return nil, fmt.Errorf("error updating child %s: %w", child.ID(), err)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Add mock's command if set
	if m.updateReturnCmd != nil {
		cmds = append(cmds, m.updateReturnCmd)
	}

	if m.updateReturnError != nil {
		return nil, m.updateReturnError
	}

	if len(cmds) == 0 {
		return nil, nil
	}
	if len(cmds) == 1 {
		return cmds[0], nil
	}
	return tea.Batch(cmds...), nil
}

// Render implements the Component interface
func (m *MockComponent) Render() string {
	m.renderedCalled = true
	m.rendered = "MockComponent:" + m.id

	// Incorporate children's rendering
	childrenOutput := make([]string, len(m.children))
	for i, child := range m.children {
		childrenOutput[i] = child.Render()
	}

	if len(childrenOutput) > 0 {
		return m.rendered + "\n" + strings.Join(childrenOutput, "\n")
	}
	return m.rendered
}

// Dispose implements the Component interface
func (m *MockComponent) Dispose() error {
	m.disposed = true

	// Dispose children first
	for _, child := range m.children {
		if err := child.Dispose(); err != nil {
			return fmt.Errorf("error disposing child %s: %w", child.ID(), err)
		}
	}

	return m.disposeError
}

// ID implements the Component interface
func (m *MockComponent) ID() string {
	return m.id
}

// AddChild implements the Component interface
func (m *MockComponent) AddChild(child Component) {
	m.children = append(m.children, child)
}

// RemoveChild implements the Component interface
func (m *MockComponent) RemoveChild(id string) bool {
	for i, child := range m.children {
		if child.ID() == id {
			// Remove the child by replacing it with the last element and truncating
			m.children[i] = m.children[len(m.children)-1]
			m.children = m.children[:len(m.children)-1]
			return true
		}
	}
	return false
}

// Children implements the Component interface
func (m *MockComponent) Children() []Component {
	return m.children
}

// TestComponentLifecycle checks that lifecycle methods work correctly
func TestComponentLifecycle(t *testing.T) {
	t.Run("Basic Lifecycle", func(t *testing.T) {
		m := &MockComponent{id: "test"}

		// Test Initialize
		err := m.Initialize()
		if err != nil {
			t.Errorf("Initialize returned an error: %v", err)
		}
		if !m.initialized {
			t.Error("Initialize did not set initialized flag")
		}

		// Test Render
		rendered := m.Render()
		if !m.renderedCalled {
			t.Error("Render did not set renderedCalled flag")
		}
		if !strings.Contains(rendered, "MockComponent:test") {
			t.Errorf("Render returned unexpected string: %s", rendered)
		}

		// Test Dispose
		err = m.Dispose()
		if err != nil {
			t.Errorf("Dispose returned an error: %v", err)
		}
		if !m.disposed {
			t.Error("Dispose did not set disposed flag")
		}
	})

	t.Run("Error Handling", func(t *testing.T) {
		// Test error propagation during Initialize
		expectedError := errors.New("initialize error")
		m := &MockComponent{id: "errorTest", initializeError: expectedError}

		err := m.Initialize()
		if err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}

		// Test error propagation during Update
		expectedError = errors.New("update error")
		m = &MockComponent{id: "errorTest", updateReturnError: expectedError}

		_, err = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}

		// Test error propagation during Dispose
		expectedError = errors.New("dispose error")
		m = &MockComponent{id: "errorTest", disposeError: expectedError}

		err = m.Dispose()
		if err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
	})

	t.Run("Command Batch", func(t *testing.T) {
		// Test command handling
		cmdExecuted := false
		testCmd := func() tea.Msg { cmdExecuted = true; return nil }
		m := &MockComponent{id: "cmdTest", updateReturnCmd: testCmd}

		cmd, err := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if err != nil {
			t.Errorf("Update returned unexpected error: %v", err)
		}
		if cmd == nil {
			t.Error("Expected command, got nil")
		} else {
			// Execute the command
			if msg := cmd(); msg != nil {
				t.Errorf("Command returned unexpected message: %v", msg)
			}
			if !cmdExecuted {
				t.Error("Command was not executed")
			}
		}
	})
}

// TestNestedComponentLifecycle tests lifecycle methods propagation through the component tree
func TestNestedComponentLifecycle(t *testing.T) {
	// Create a tree of components
	parent := &MockComponent{id: "parent"}
	child1 := &MockComponent{id: "child1"}
	child2 := &MockComponent{id: "child2"}
	grandchild1 := &MockComponent{id: "grandchild1"}
	child1.AddChild(grandchild1)
	parent.AddChild(child1)
	parent.AddChild(child2)

	// Test initialization propagation
	err := parent.Initialize()
	if err != nil {
		t.Errorf("Initialize returned unexpected error: %v", err)
	}
	if !parent.initialized || !child1.initialized || !child2.initialized || !grandchild1.initialized {
		t.Error("Initialize did not propagate to all components in the tree")
	}

	// Test update propagation
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, err = parent.Update(msg)
	if err != nil {
		t.Errorf("Update returned unexpected error: %v", err)
	}
	if parent.updateCallCount != 1 || child1.updateCallCount != 1 || child2.updateCallCount != 1 || grandchild1.updateCallCount != 1 {
		t.Error("Update did not propagate to all components in the tree")
	}

	// Check if all components received the same message
	components := []*MockComponent{parent, child1, child2, grandchild1}
	for _, c := range components {
		if len(c.handledMsgs) != 1 {
			t.Errorf("Component %s did not receive expected number of messages", c.id)
			continue
		}
		receivedMsg, ok := c.handledMsgs[0].(tea.KeyMsg)
		if !ok {
			t.Errorf("Component %s received wrong message type", c.id)
			continue
		}
		if receivedMsg.Type != tea.KeyEnter {
			t.Errorf("Component %s received wrong message: %v", c.id, receivedMsg)
		}
	}

	// Test render propagation and composition
	output := parent.Render()
	// Each component should be in the output
	for _, c := range components {
		if !strings.Contains(output, c.id) {
			t.Errorf("Render output does not contain component %s: %s", c.id, output)
		}
	}

	// Test disposal propagation
	err = parent.Dispose()
	if err != nil {
		t.Errorf("Dispose returned unexpected error: %v", err)
	}
	if !parent.disposed || !child1.disposed || !child2.disposed || !grandchild1.disposed {
		t.Error("Dispose did not propagate to all components in the tree")
	}
}

// TestIdentityMechanism checks component identity and keying functionality
func TestIdentityMechanism(t *testing.T) {
	t.Run("Unique ID", func(t *testing.T) {
		components := []*MockComponent{
			{id: "comp1"},
			{id: "comp2"},
			{id: "comp3"},
		}

		// Check uniqueness of IDs
		ids := make(map[string]bool)
		for _, c := range components {
			if ids[c.ID()] {
				t.Errorf("Duplicate ID found: %s", c.ID())
			}
			ids[c.ID()] = true
		}
	})

	t.Run("ID-based Lookup", func(t *testing.T) {
		// Create a component tree
		parent := &MockComponent{id: "parent"}
		for i := 1; i <= 5; i++ {
			parent.AddChild(&MockComponent{id: fmt.Sprintf("child%d", i)})
		}

		// Find a component by ID
		foundChild := findComponentByID(parent, "child3")
		if foundChild == nil {
			t.Error("Could not find component by ID")
		} else if foundChild.ID() != "child3" {
			t.Errorf("Found wrong component: %s", foundChild.ID())
		}

		// Try to find a non-existent component
		nonExistent := findComponentByID(parent, "nonexistent")
		if nonExistent != nil {
			t.Errorf("Found non-existent component: %s", nonExistent.ID())
		}
	})

	t.Run("BaseComponent ID", func(t *testing.T) {
		// Test the BaseComponent ID mechanism
		base := NewBaseComponent("baseID")
		if base.ID() != "baseID" {
			t.Errorf("BaseComponent returned wrong ID: %s", base.ID())
		}
	})
}

// Helper function to find a component by ID in a component tree
func findComponentByID(root Component, id string) Component {
	if root.ID() == id {
		return root
	}
	for _, child := range root.Children() {
		if result := findComponentByID(child, id); result != nil {
			return result
		}
	}
	return nil
}

// TestChildManagement checks child component operations
func TestChildManagement(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		parent := &MockComponent{id: "parent"}
		child1 := &MockComponent{id: "child1"}
		child2 := &MockComponent{id: "child2"}

		// Test AddChild
		parent.AddChild(child1)
		parent.AddChild(child2)

		children := parent.Children()
		if len(children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(children))
		}

		// Test RemoveChild
		removed := parent.RemoveChild("child1")
		if !removed {
			t.Error("RemoveChild returned false, expected true")
		}

		children = parent.Children()
		if len(children) != 1 {
			t.Errorf("Expected 1 child after removal, got %d", len(children))
		}

		if children[0].ID() != "child2" {
			t.Errorf("Expected remaining child to be child2, got %s", children[0].ID())
		}

		// Test removing non-existent child
		removed = parent.RemoveChild("nonexistent")
		if removed {
			t.Error("RemoveChild returned true for non-existent child, expected false")
		}
	})

	t.Run("Child Updates", func(t *testing.T) {
		// Test that parent correctly propagates updates to children
		parent := &MockComponent{id: "parent"}
		child1 := &MockComponent{id: "child1"}
		child2 := &MockComponent{id: "child2"}

		parent.AddChild(child1)
		parent.AddChild(child2)

		// Configure child2 to return a command
		cmdExecuted := false
		testCmd := func() tea.Msg { cmdExecuted = true; return nil }
		child2.updateReturnCmd = testCmd

		// Update should propagate to children and collect their commands
		cmd, err := parent.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if err != nil {
			t.Errorf("Update returned unexpected error: %v", err)
		}
		if cmd == nil {
			t.Error("Expected command from child, got nil")
		} else {
			// Execute the command
			cmd()
			if !cmdExecuted {
				t.Error("Child command was not executed")
			}
		}

		// Check that all children received the update
		if child1.updateCallCount != 1 || child2.updateCallCount != 1 {
			t.Errorf("Update did not propagate correctly: child1=%d, child2=%d",
				child1.updateCallCount, child2.updateCallCount)
		}
	})

	t.Run("Child Error Propagation", func(t *testing.T) {
		// Test that errors from children are propagated upward
		parent := &MockComponent{id: "parent"}
		child := &MockComponent{id: "child", updateReturnError: errors.New("child error")}

		parent.AddChild(child)

		_, err := parent.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if err == nil {
			t.Error("Expected error from child to propagate, got nil")
		} else if !strings.Contains(err.Error(), "child error") {
			t.Errorf("Expected error to contain 'child error', got: %v", err)
		}
	})

	t.Run("BaseComponent Child Management", func(t *testing.T) {
		// Test the BaseComponent child management
		base := NewBaseComponent("baseParent")
		child1 := NewBaseComponent("baseChild1")
		child2 := NewBaseComponent("baseChild2")

		base.AddChild(child1)
		base.AddChild(child2)

		children := base.Children()
		if len(children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(children))
		}

		removed := base.RemoveChild("baseChild1")
		if !removed {
			t.Error("RemoveChild returned false, expected true")
		}

		children = base.Children()
		if len(children) != 1 {
			t.Errorf("Expected 1 child after removal, got %d", len(children))
		}
	})
}
