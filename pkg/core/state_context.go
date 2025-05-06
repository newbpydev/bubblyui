package core

import (
	"fmt"
	"reflect"
	"sync"
)

// ContextID is a unique identifier for a context
type ContextID string

// StateContext represents a shared state context that can be provided
// by a parent component and consumed by child components
type StateContext[T any] struct {
	id           ContextID
	name         string
	state        *State[T]
	defaultValue T
	mutex        sync.RWMutex
}

// globalContextRegistry stores all contexts by their ID
var (
	globalContextRegistry = make(map[ContextID]interface{})
	globalContextMutex    sync.RWMutex
)

// CreateStateContext creates a new state context with the given name and default value
func CreateStateContext[T any](name string, defaultValue T) *StateContext[T] {
	globalContextMutex.Lock()
	defer globalContextMutex.Unlock()

	id := ContextID(fmt.Sprintf("context_%s", name))

	// Check if context already exists
	if existingCtx, ok := globalContextRegistry[id]; ok {
		context, ok := existingCtx.(*StateContext[T])
		if !ok {
			panic(fmt.Sprintf("Type mismatch for existing context '%s'", name))
		}
		return context
	}

	// Create a new context
	context := &StateContext[T]{
		id:           id,
		name:         name,
		state:        NewState(defaultValue),
		defaultValue: defaultValue,
	}

	// Register context globally
	globalContextRegistry[id] = context

	return context
}

// ProvideContext registers a state provider for the given context in a component
func ProvideContext[T any](component StatefulComponent, context *StateContext[T], value T) {
	context.mutex.Lock()
	defer context.mutex.Unlock()

	// Set the value in the context
	context.state.Set(value)

	// Register an unmount hook to reset the value when the provider unmounts
	cs := component.GetState()
	if cs != nil {
		cs.GetHookManager().OnUnmount(func() error {
			// Reset the context value to its default
			context.state.Set(context.defaultValue)
			return nil
		})
	}
}

// ProvideContextState registers a state provider for the given context using a State
func ProvideContextState[T any](component StatefulComponent, context *StateContext[T], state *State[T]) {
	context.mutex.Lock()
	defer context.mutex.Unlock()

	// Create a subscription to the state
	callback := func(old, new T) {
		context.state.Set(new)
	}

	// Set initial value
	context.state.Set(state.Get())

	// Register the change handler
	state.OnChange(callback)

	// Register an unmount hook to clean up when component unmounts
	cs := component.GetState()
	if cs != nil {
		cs.GetHookManager().OnUnmount(func() error {
			// Remove the change handler
			state.RemoveOnChange(callback)
			// Reset the context to default
			context.state.Set(context.defaultValue)
			return nil
		})
	}
}

// UseContext returns the current value of the context
func UseContext[T any](component StatefulComponent, context *StateContext[T]) *Signal[T] {
	// Create a state to hold the returned value
	resultState := NewState(context.state.Get())

	// Subscribe to context state changes
	context.state.OnChange(func(oldVal, newVal T) {
		// When context value changes, update the result state
		resultState.Set(newVal)
	})

	// Return signal from the state
	return resultState.GetSignal()
}

// UseContextWithDefault returns the value of the context or a default value if not provided
func UseContextWithDefault[T any](component StatefulComponent, context *StateContext[T], defaultValue T) *Signal[T] {
	// Create a state to hold the computed value
	resultState := NewState(defaultValue)

	// Subscribe to context state changes
	context.state.OnChange(func(oldVal, newVal T) {
		// When context value changes, check if we should use the provided value or default
		if !reflect.DeepEqual(newVal, context.defaultValue) {
			// Context has a non-default value, use it
			resultState.Set(newVal)
		} else {
			// Context has the default value, use component's local default
			resultState.Set(defaultValue)
		}
	})

	// Check initial value
	initialValue := context.state.Get()
	if !reflect.DeepEqual(initialValue, context.defaultValue) {
		// Context already has a non-default value, use it
		resultState.Set(initialValue)
	}

	// Return signal from the state
	return resultState.GetSignal()
}

// UseContextState returns a state that reflects the context value
func UseContextState[T any](component StatefulComponent, context *StateContext[T]) *State[T] {
	// Create a state that mirrors the context
	componentState := NewState(context.state.Get())

	// Create a subscription to the context
	contextCallback := func(old, new T) {
		componentState.Set(new)
	}

	// Register the context change handler
	context.state.OnChange(contextCallback)

	// Create a subscription to the component state
	stateCallback := func(old, new T) {
		context.state.Set(new)
	}

	// Register the state change handler
	componentState.OnChange(stateCallback)

	// Register an unmount hook to clean up when component unmounts
	cs := component.GetState()
	if cs != nil {
		cs.GetHookManager().OnUnmount(func() error {
			// Remove the change handlers
			context.state.RemoveOnChange(contextCallback)
			componentState.RemoveOnChange(stateCallback)
			return nil
		})
	}

	// Return the state
	return componentState
}

// GetAllContexts returns all registered contexts
func GetAllContexts() map[ContextID]interface{} {
	globalContextMutex.RLock()
	defer globalContextMutex.RUnlock()

	// Make a copy to avoid race conditions
	contexts := make(map[ContextID]interface{})
	for id, ctx := range globalContextRegistry {
		contexts[id] = ctx
	}

	return contexts
}

// ClearAllContexts removes all registered contexts
// This is primarily used for testing
func ClearAllContexts() {
	globalContextMutex.Lock()
	defer globalContextMutex.Unlock()

	globalContextRegistry = make(map[ContextID]interface{})
}
