package core

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

// StatefulComponent extends the Component interface with state management capabilities.
type StatefulComponent interface {
	Component

	// GetState returns the ComponentState for this component.
	GetState() *ComponentState

	// ExecuteEffect runs all registered effects when their dependencies change.
	ExecuteEffect() error

	// IsMounted returns whether the component is currently mounted in the UI tree.
	IsMounted() bool

	// SetMounted sets the component's mounted status.
	SetMounted(mounted bool)
}

// BaseStatefulComponent provides a default implementation of the StatefulComponent interface.
// Components can embed this struct to get state management capabilities.
type BaseStatefulComponent struct {
	*BaseComponent
	state   *ComponentState
	mounted bool
}

// NewBaseStatefulComponent creates a new BaseStatefulComponent with the given ID and name.
func NewBaseStatefulComponent(id, name string) *BaseStatefulComponent {
	return &BaseStatefulComponent{
		BaseComponent: NewBaseComponent(id),
		state:         NewComponentState(id, name),
		mounted:       false,
	}
}

// Initialize initializes the component and executes mount hooks.
func (b *BaseStatefulComponent) Initialize() error {
	// Initialize base component (which handles children)
	err := b.BaseComponent.Initialize()
	if err != nil {
		return err
	}

	// Mark component as mounted
	b.mounted = true

	// Execute mount hooks
	return b.state.GetHookManager().ExecuteMountHooks()
}

// Update handles messages, executes update hooks, and returns commands.
func (b *BaseStatefulComponent) Update(msg tea.Msg) (tea.Cmd, error) {
	// Run update hooks
	err := b.state.GetHookManager().ExecuteUpdateHooks()
	if err != nil {
		return nil, err
	}

	// Call base Update to handle children
	return b.BaseComponent.Update(msg)
}

// Dispose cleans up resources and executes unmount hooks.
func (b *BaseStatefulComponent) Dispose() error {
	// Mark component as unmounted
	b.mounted = false

	// Dispose component state first
	err := b.state.Dispose()
	if err != nil {
		return err
	}

	// Dispose base component (which handles children)
	return b.BaseComponent.Dispose()
}

// GetState returns the ComponentState for this component.
func (b *BaseStatefulComponent) GetState() *ComponentState {
	return b.state
}

// ExecuteEffect runs all registered effects when their dependencies change.
func (b *BaseStatefulComponent) ExecuteEffect() error {
	return b.state.GetHookManager().ExecuteUpdateHooks()
}

// IsMounted returns whether the component is currently mounted in the UI tree.
func (b *BaseStatefulComponent) IsMounted() bool {
	return b.mounted
}

// SetMounted sets the component's mounted status.
func (b *BaseStatefulComponent) SetMounted(mounted bool) {
	b.mounted = mounted
}

// WithState is a helper function to create a new component state in the current component
// and provide a handle to it for state management.
func WithState[T any](
	component StatefulComponent,
	name string,
	initialValue T,
) (*State[T], func(T), func(StateUpdate[T])) {
	return UseState(component.GetState(), name, initialValue)
}

// WithStateEquals is a helper function to create a new component state with custom equality.
func WithStateEquals[T any](
	component StatefulComponent,
	name string,
	initialValue T,
	equals func(a, b T) bool,
) (*State[T], func(T), func(StateUpdate[T])) {
	return UseStateWithEquals(component.GetState(), name, initialValue, equals)
}

// WithStateHistory is a helper function to create a new component state with history.
func WithStateHistory[T any](
	component StatefulComponent,
	name string,
	initialValue T,
	historySize int,
) (*State[T], func(T), func(StateUpdate[T])) {
	return UseStateWithHistory(component.GetState(), name, initialValue, historySize)
}

// WithMemo is a helper function to create a memoized value in the component.
func WithMemo[T any](
	component StatefulComponent,
	name string,
	computeFn func() T,
	deps []interface{},
) *Signal[T] {
	return UseMemo(component.GetState(), name, computeFn, deps)
}

// WithEffect is a helper function to register an effect in the component.
func WithEffect(
	component StatefulComponent,
	name string,
	effectFn func() (cleanup func(), err error),
	deps []interface{},
) {
	UseEffect(component.GetState(), name, effectFn, deps)
}

// BatchComponentState batches state updates for a component.
func BatchComponentState(component StatefulComponent, fn func()) {
	BatchState(component.GetState(), fn)
}

// WithAsyncState creates an async state updater for the given state.
func WithAsyncState[T any](state *State[T]) func(asyncFn func() (T, error)) {
	return func(asyncFn func() (T, error)) {
		// Use goroutine for async operation
		go func() {
			// Perform the async operation
			newValue, err := asyncFn()
			if err != nil {
				// Handle error - for now just log it
				// In the future, we could have an error state
				fmt.Printf("Error in async state update: %v\n", err)
				return
			}

			// Update the state with the new value
			state.Set(newValue)
		}()
	}
}
