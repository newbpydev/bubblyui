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
//   - adding child would create a circular reference (ErrCircularRef)
//   - adding child would exceed maximum depth (ErrMaxDepth)
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

	// Check for self-reference (A -> A)
	if child == Component(c) {
		return &CircularRefError{
			ParentName: c.name,
			ChildName:  child.Name(),
			Message:    "component cannot be a child of itself",
		}
	}

	// Check for circular reference (child is an ancestor of this component)
	if c.hasAncestor(child) {
		return &CircularRefError{
			ParentName: c.name,
			ChildName:  child.Name(),
			Message:    "child is an ancestor of parent",
		}
	}

	// Check if child has this component as an ancestor (reverse check)
	if childImpl, ok := child.(*componentImpl); ok {
		if childImpl.hasAncestor(Component(c)) {
			return &CircularRefError{
				ParentName: c.name,
				ChildName:  child.Name(),
				Message:    "parent is an ancestor of child",
			}
		}

		// Calculate depth after adding child
		// The child will be at: parent's depth + 1
		parentDepth := calculateDepthToRoot(c)
		childDepth := parentDepth + 1

		// The deepest point in the tree after adding this child would be:
		// child's depth + child's subtree depth
		childSubtreeDepth := calculateComponentDepth(childImpl)
		maxDepthAfterAdd := childDepth + childSubtreeDepth

		if maxDepthAfterAdd > MaxComponentDepth {
			return &MaxDepthError{
				ComponentName: child.Name(),
				CurrentDepth:  maxDepthAfterAdd,
				MaxDepth:      MaxComponentDepth,
			}
		}
	}

	c.childrenMu.Lock()
	defer c.childrenMu.Unlock()

	// Add child to slice
	c.children = append(c.children, child)

	// Set parent reference
	if childImpl, ok := child.(*componentImpl); ok {
		childImpl.parent = c
	}

	// Notify hook after successful add
	notifyHookChildAdded(c.id, child.ID())

	return nil
}

// calculateDepthToRoot calculates the depth from this component to the root.
// Returns 0 if this component is the root (no parent).
func calculateDepthToRoot(c *componentImpl) int {
	if c.parent == nil {
		return 0
	}

	return calculateDepthToRoot(c.parent) + 1
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

	// Notify hook after successful remove
	notifyHookChildRemoved(c.id, childID)

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
