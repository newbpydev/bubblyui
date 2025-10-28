package bubbly

import (
	"errors"
	"fmt"
)

// Component-level error types for validation and runtime errors.
// These errors are returned when component operations fail due to
// invalid state, configuration, or usage patterns.
var (
	// ErrCircularRef is returned when a circular component reference is detected.
	// This occurs when attempting to add a component as a child that would create
	// a cycle in the component tree (e.g., A -> B -> A or A -> A).
	//
	// Circular references are prevented to avoid:
	//   - Infinite loops during rendering
	//   - Stack overflow during tree traversal
	//   - Memory leaks from reference cycles
	//
	// Example scenario:
	//   compA.AddChild(compB)  // OK
	//   compB.AddChild(compA)  // Returns ErrCircularRef
	ErrCircularRef = errors.New("circular component reference detected")

	// ErrMaxDepth is returned when the component tree exceeds the maximum allowed depth.
	// The maximum depth is defined by MaxComponentDepth constant.
	//
	// Deep component trees can cause:
	//   - Performance degradation during rendering
	//   - Stack overflow in recursive operations
	//   - Difficult debugging and maintenance
	//
	// Recommended maximum depth is ~10 levels. Trees deeper than MaxComponentDepth
	// (50 levels) are rejected to prevent performance issues.
	//
	// Example scenario:
	//   // Creating a 51-level deep tree
	//   for i := 0; i < 51; i++ {
	//       parent.AddChild(child)  // Returns ErrMaxDepth at depth 51
	//   }
	ErrMaxDepth = errors.New("maximum component depth exceeded")
)

// MaxComponentDepth defines the maximum allowed depth of the component tree.
// This limit prevents performance issues and stack overflow from deeply nested components.
//
// The depth is measured from the root component (depth 0) to the deepest leaf.
// For example:
//
//	Root (depth 0)
//	└── Child (depth 1)
//	    └── Grandchild (depth 2)
//
// Recommended depth: ~10 levels for optimal performance
// Maximum allowed depth: 50 levels (enforced by AddChild)
const MaxComponentDepth = 50

// CircularRefError provides detailed information about a circular reference.
// It includes the component names involved in the cycle for easier debugging.
type CircularRefError struct {
	// ParentName is the name of the component attempting to add a child
	ParentName string
	// ChildName is the name of the component being added
	ChildName string
	// Message provides additional context about the circular reference
	Message string
}

// Error implements the error interface for CircularRefError.
func (e *CircularRefError) Error() string {
	return fmt.Sprintf("circular component reference: %s -> %s (%s)",
		e.ParentName, e.ChildName, e.Message)
}

// Unwrap returns the underlying ErrCircularRef for error comparison.
func (e *CircularRefError) Unwrap() error {
	return ErrCircularRef
}

// MaxDepthError provides detailed information about a max depth violation.
// It includes the current depth and the component name for debugging.
type MaxDepthError struct {
	// ComponentName is the name of the component that would exceed max depth
	ComponentName string
	// CurrentDepth is the depth at which the error occurred
	CurrentDepth int
	// MaxDepth is the maximum allowed depth
	MaxDepth int
}

// Error implements the error interface for MaxDepthError.
func (e *MaxDepthError) Error() string {
	return fmt.Sprintf("maximum component depth exceeded: component '%s' at depth %d (max: %d)",
		e.ComponentName, e.CurrentDepth, e.MaxDepth)
}

// Unwrap returns the underlying ErrMaxDepth for error comparison.
func (e *MaxDepthError) Unwrap() error {
	return ErrMaxDepth
}

// HandlerPanicError wraps a panic that occurred in an event handler.
// This allows the application to continue running even if a handler panics.
//
// Note: This type is also available in the observability package to avoid
// import cycles when integrating with error reporting.
type HandlerPanicError struct {
	// ComponentName is the name of the component where the panic occurred
	ComponentName string
	// EventName is the name of the event being handled
	EventName string
	// PanicValue is the value passed to panic()
	PanicValue interface{}
}

// Error implements the error interface for HandlerPanicError.
func (e *HandlerPanicError) Error() string {
	return fmt.Sprintf("panic in event handler: component '%s', event '%s', panic: %v",
		e.ComponentName, e.EventName, e.PanicValue)
}

// calculateComponentDepth calculates the depth of a component in the tree.
// Returns the maximum depth from this component to any leaf node.
// A component with no children has depth 0.
func calculateComponentDepth(c *componentImpl) int {
	if len(c.children) == 0 {
		return 0
	}

	maxChildDepth := 0
	for _, child := range c.children {
		if impl, ok := child.(*componentImpl); ok {
			childDepth := calculateComponentDepth(impl)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}
	}

	return maxChildDepth + 1
}

// hasAncestor checks if the given component is an ancestor of the current component.
// This is used to detect circular references in the component tree.
// Returns true if ancestor is found in the parent chain.
func (c *componentImpl) hasAncestor(ancestor Component) bool {
	if c.parent == nil {
		return false
	}

	// Check if immediate parent is the ancestor
	if *c.parent == ancestor {
		return true
	}

	// Recursively check parent's ancestors
	if parentImpl, ok := (*c.parent).(*componentImpl); ok {
		return parentImpl.hasAncestor(ancestor)
	}

	return false
}
