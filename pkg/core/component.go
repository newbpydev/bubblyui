package core

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Component is the base interface for all UI components in BubblyUI.
// It defines the core lifecycle methods and component management functionality
// inspired by modern web component frameworks like React and Solid.js.
type Component interface {
	// Initialize sets up the component and is called once when the component is added to the UI tree.
	// This is where you would set up initial state, register event handlers, or set up subscriptions.
	Initialize() error

	// Update handles messages and returns commands to be executed.
	// This is called whenever a message is received (e.g., user input, timer events).
	// It returns a tea.Cmd that will be executed by the Bubble Tea runtime.
	Update(msg tea.Msg) (tea.Cmd, error)

	// Render returns the visual representation of the component as a string.
	// This is called whenever the UI needs to be redrawn.
	Render() string

	// Dispose cleans up resources when the component is removed from the UI tree.
	// This is where you would clean up subscriptions, close channels, etc.
	Dispose() error

	// ID returns the unique identifier for this component.
	// This is used for component reconciliation and finding components in the UI tree.
	ID() string

	// AddChild adds a child component to this component.
	// This establishes the parent-child relationship for component composition.
	AddChild(child Component)

	// RemoveChild removes a child component with the given ID.
	// Returns true if a child was removed, false otherwise.
	RemoveChild(id string) bool

	// Children returns all child components of this component.
	// This allows traversal of the component tree.
	Children() []Component
}

// BaseComponent provides a default implementation of the Component interface.
// Components can embed this struct to get default implementations of common methods.
type BaseComponent struct {
	id       string
	children []Component
}

// NewBaseComponent creates a new BaseComponent with the given ID.
func NewBaseComponent(id string) *BaseComponent {
	return &BaseComponent{
		id:       id,
		children: make([]Component, 0),
	}
}

// Initialize provides a default implementation that initializes all children.
func (b *BaseComponent) Initialize() error {
	for _, child := range b.children {
		if err := child.Initialize(); err != nil {
			return err
		}
	}
	return nil
}

// Update provides a default implementation that updates all children.
func (b *BaseComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	cmds := make([]tea.Cmd, 0)
	for _, child := range b.children {
		cmd, err := child.Update(msg)
		if err != nil {
			return nil, err
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) == 0 {
		return nil, nil
	}
	if len(cmds) == 1 {
		return cmds[0], nil
	}
	return tea.Batch(cmds...), nil
}

// Render provides a default implementation that returns an empty string.
// Components should override this to provide their own rendering logic.
func (b *BaseComponent) Render() string {
	return ""
}

// Dispose provides a default implementation that disposes all children.
func (b *BaseComponent) Dispose() error {
	for _, child := range b.children {
		if err := child.Dispose(); err != nil {
			return err
		}
	}
	return nil
}

// ID returns the component's unique identifier.
func (b *BaseComponent) ID() string {
	return b.id
}

// AddChild adds a child component to this component.
func (b *BaseComponent) AddChild(child Component) {
	b.children = append(b.children, child)
}

// RemoveChild removes a child component with the given ID.
func (b *BaseComponent) RemoveChild(id string) bool {
	for i, child := range b.children {
		if child.ID() == id {
			// Remove the child by replacing it with the last element and truncating
			b.children[i] = b.children[len(b.children)-1]
			b.children = b.children[:len(b.children)-1]
			return true
		}
	}
	return false
}

// Children returns all child components of this component.
func (b *BaseComponent) Children() []Component {
	return b.children
}
