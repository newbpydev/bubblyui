package bubbly

import (
	"errors"
	"fmt"
)

var (
	// ErrNilChild is returned when attempting to add a nil child component
	ErrNilChild = errors.New("cannot add nil child component")

	// ErrChildNotFound is returned when attempting to remove a child that doesn't exist
	ErrChildNotFound = errors.New("child component not found")
)

// Children returns a copy of the component's children slice.
// This method is thread-safe and returns a defensive copy to prevent
// external modification of the internal children slice.
//
// Example:
//
//	children := component.Children()
//	for _, child := range children {
//	    fmt.Println(child.Name())
//	}
func (c *componentImpl) Children() []Component {
	c.childrenMu.RLock()
	defer c.childrenMu.RUnlock()

	// Return defensive copy
	if len(c.children) == 0 {
		return []Component{}
	}

	copy := make([]Component, len(c.children))
	for i, child := range c.children {
		copy[i] = child
	}
	return copy
}

// AddChild adds a child component to this component and sets the parent reference.
// This method is thread-safe and can be called during runtime to dynamically
// add children to a component.
//
// Returns an error if:
//   - child is nil (ErrNilChild)
//
// Example:
//
//	child, _ := NewComponent("Child").
//	    Template(func(ctx RenderContext) string { return "child" }).
//	    Build()
//
//	err := parent.AddChild(child)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *componentImpl) AddChild(child Component) error {
	if child == nil {
		return ErrNilChild
	}

	c.childrenMu.Lock()
	defer c.childrenMu.Unlock()

	// Add child to slice
	c.children = append(c.children, child)

	// Set parent reference
	if childImpl, ok := child.(*componentImpl); ok {
		parent := Component(c)
		childImpl.parent = &parent
	}

	return nil
}

// RemoveChild removes a child component from this component and clears its parent reference.
// This method is thread-safe and can be called during runtime to dynamically
// remove children from a component.
//
// The child is identified by its ID, not by pointer equality.
//
// Returns an error if:
//   - child is nil (ErrNilChild)
//   - child is not found in the children slice (ErrChildNotFound)
//
// Example:
//
//	err := parent.RemoveChild(child)
//	if err != nil {
//	    log.Printf("Failed to remove child: %v", err)
//	}
func (c *componentImpl) RemoveChild(child Component) error {
	if child == nil {
		return ErrNilChild
	}

	c.childrenMu.Lock()
	defer c.childrenMu.Unlock()

	// Find child by ID
	childID := child.ID()
	foundIndex := -1
	for i, c := range c.children {
		if c.ID() == childID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("%w: %s", ErrChildNotFound, childID)
	}

	// Remove child from slice
	c.children = append(c.children[:foundIndex], c.children[foundIndex+1:]...)

	// Clear parent reference
	if childImpl, ok := child.(*componentImpl); ok {
		childImpl.parent = nil
	}

	return nil
}

// renderChildren renders all child components and returns their output as a slice of strings.
// This is a helper method used internally by the component for rendering children.
// It calls View() on each child component.
//
// This method is thread-safe.
//
// Example:
//
//	outputs := component.renderChildren()
//	result := strings.Join(outputs, "\n")
func (c *componentImpl) renderChildren() []string {
	c.childrenMu.RLock()
	defer c.childrenMu.RUnlock()

	if len(c.children) == 0 {
		return []string{}
	}

	outputs := make([]string, len(c.children))
	for i, child := range c.children {
		outputs[i] = child.View()
	}

	return outputs
}
