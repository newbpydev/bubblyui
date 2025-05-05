/*
Package core provides the foundational interfaces and types for the BubblyUI framework.

BubblyUI is a component-based reactive TUI framework built on top of Bubble Tea
and Lip Gloss, designed to transform Bubble Tea's Elm-inspired architecture into a
React/Solid-like component system with fine-grained reactivity, independent component
state, and elegant composition patterns.

# Component Interface

The Component interface is the foundation of BubblyUI's component model. It defines
the core lifecycle methods and component management functionality that all UI components
must implement.

## Lifecycle Methods

Components in BubblyUI follow a clear lifecycle:

1. Initialization: The component is created and initialized.
2. Updates: The component receives messages and updates its state.
3. Rendering: The component produces its visual representation.
4. Disposal: The component is removed and cleans up resources.

These lifecycle stages are managed through the following methods:

- Initialize(): Called once when the component is added to the UI tree.
- Update(msg): Called whenever a message is received.
- Render(): Called whenever the UI needs to be redrawn.
- Dispose(): Called when the component is removed from the UI tree.

## Component Composition

Components can be composed into a tree structure using the child management methods:

- AddChild(child): Adds a child component to this component.
- RemoveChild(id): Removes a child component with the given ID.
- Children(): Returns all child components of this component.

Each component has a unique ID, which is used for reconciliation and finding components
in the UI tree:

- ID(): Returns the unique identifier for this component.

# Usage Examples

## Basic Component Implementation

Here's a simple example of implementing a Component:

	type Button struct {
		*core.BaseComponent
		label string
		onClick func()
	}

	func NewButton(id string, label string, onClick func()) *Button {
		return &Button{
			BaseComponent: core.NewBaseComponent(id),
			label: label,
			onClick: onClick,
		}
	}

	func (b *Button) Render() string {
		return fmt.Sprintf("[ %s ]", b.label)
	}

	func (b *Button) Update(msg tea.Msg) (tea.Cmd, error) {
		switch msg := msg.(type) {
		case tea.MouseMsg:
			if msg.Type == tea.MouseLeft {
				// Check if click is within button bounds
				// ...
				if b.onClick != nil {
					b.onClick()
				}
			}
		}
		return nil, nil
	}

## Component Composition Example

Here's how to compose components together:

	// Create a form with input and button
	form := NewForm("myForm")
	input := NewInput("nameInput", "Enter your name")
	button := NewButton("submitBtn", "Submit", func() {
		fmt.Println("Name submitted:", input.Value())
	})

	// Add components to form
	form.AddChild(input)
	form.AddChild(button)

	// Initialize all components
	form.Initialize()

# Best Practices

## Component Design

 1. **Single Responsibility**: Each component should have a single responsibility.
    For example, a Button component should only handle button-related logic and rendering.

 2. **Reusability**: Design components to be reusable by accepting props that
    customize their behavior and appearance.

 3. **Testability**: Components should be easy to test in isolation. Avoid
    direct dependencies on global state.

## Lifecycle Management

 1. **Proper Initialization**: Set up all resources during Initialize(). Don't assume
    resources are available before this method is called.

 2. **Clean Disposal**: Always clean up resources in Dispose() to prevent memory leaks.
    This includes unregistering event handlers and closing channels.

 3. **Propagate Lifecycle Events**: Parent components should propagate lifecycle
    events to their children.

## Update Handling

 1. **Message Filtering**: Only handle messages that are relevant to your component.
    Pass other messages to children.

 2. **Batched Commands**: When multiple commands need to be returned, use tea.Batch
    to combine them.

 3. **Error Handling**: Always check for errors when updating children and propagate
    them up the component tree.

## Rendering

 1. **Composition**: Compose component output using Lip Gloss's join functions
    (JoinHorizontal, JoinVertical).

 2. **Positioning**: Handle positioning of child components within the parent's
    layout.

 3. **Caching**: Cache rendered output when possible to avoid unnecessary string
    operations.

For more detailed examples and documentation, see the examples directory.
*/
package core
