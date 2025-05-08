package core

import (
	"fmt"
	"reflect"
	"sync"
)

// ComponentState manages all state related to a single component instance
type ComponentState struct {
	componentID   string
	componentName string
	stateRegistry map[string]interface{}
	signalsByName map[string]interface{}
	mutex         sync.RWMutex
	hookManager   *HookManager
}

// NewComponentState creates a new ComponentState manager for a component
func NewComponentState(componentID, componentName string) *ComponentState {
	return &ComponentState{
		componentID:   componentID,
		componentName: componentName,
		stateRegistry: make(map[string]interface{}),
		signalsByName: make(map[string]interface{}),
		hookManager:   NewHookManager(componentName),
	}
}

// UseState creates or retrieves a state instance for the component.
// This is modeled after React's useState hook. On first call, it creates a new
// state with the initial value. On subsequent calls with the same name, it
// returns the existing state instance.
func UseState[T any](cs *ComponentState, name string, initialValue T) (*State[T], func(T), func(StateUpdate[T])) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Check if state already exists for this name
	if existingState, ok := cs.stateRegistry[name]; ok {
		// Type assertion to ensure it's the correct type
		state, ok := existingState.(*State[T])
		if !ok {
			panic(fmt.Sprintf("Type mismatch for state '%s' in component '%s'",
				name, cs.componentName))
		}

		// Return the existing state and updater functions
		setValue := func(value T) {
			state.Set(value)
		}
		updateValue := func(fn StateUpdate[T]) {
			state.Update(fn)
		}
		return state, setValue, updateValue
	}

	// Create a new state
	state := NewState(initialValue)
	cs.stateRegistry[name] = state

	// Store signal name mapping for debugging
	cs.signalsByName[name] = state.GetSignal()

	// Return the state and updater functions
	setValue := func(value T) {
		state.Set(value)
	}
	updateValue := func(fn StateUpdate[T]) {
		state.Update(fn)
	}
	return state, setValue, updateValue
}

// UseStateWithEquals is like UseState but with a custom equality function.
func UseStateWithEquals[T any](
	cs *ComponentState,
	name string,
	initialValue T,
	equals func(a, b T) bool,
) (*State[T], func(T), func(StateUpdate[T])) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Check if state already exists for this name
	if existingState, ok := cs.stateRegistry[name]; ok {
		// Type assertion to ensure it's the correct type
		state, ok := existingState.(*State[T])
		if !ok {
			panic(fmt.Sprintf("Type mismatch for state '%s' in component '%s'",
				name, cs.componentName))
		}

		// Return the existing state and updater functions
		setValue := func(value T) {
			state.Set(value)
		}
		updateValue := func(fn StateUpdate[T]) {
			state.Update(fn)
		}
		return state, setValue, updateValue
	}

	// Create a new state with custom equality
	state := NewStateWithEquals(initialValue, equals)
	cs.stateRegistry[name] = state

	// Store signal name mapping for debugging
	cs.signalsByName[name] = state.GetSignal()

	// Return the state and updater functions
	setValue := func(value T) {
		state.Set(value)
	}
	updateValue := func(fn StateUpdate[T]) {
		state.Update(fn)
	}
	return state, setValue, updateValue
}

// UseStateWithHistory is like UseState but with a specific history size.
func UseStateWithHistory[T any](
	cs *ComponentState,
	name string,
	initialValue T,
	historySize int,
) (*State[T], func(T), func(StateUpdate[T])) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Check if state already exists for this name
	if existingState, ok := cs.stateRegistry[name]; ok {
		// Type assertion to ensure it's the correct type
		state, ok := existingState.(*State[T])
		if !ok {
			panic(fmt.Sprintf("Type mismatch for state '%s' in component '%s'",
				name, cs.componentName))
		}

		// Return the existing state and updater functions
		setValue := func(value T) {
			state.Set(value)
		}
		updateValue := func(fn StateUpdate[T]) {
			state.Update(fn)
		}
		return state, setValue, updateValue
	}

	// Create a new state with specified history size
	state := NewStateWithHistory(initialValue, historySize)
	cs.stateRegistry[name] = state

	// Store signal name mapping for debugging
	cs.signalsByName[name] = state.GetSignal()

	// Return the state and updater functions
	setValue := func(value T) {
		state.Set(value)
	}
	updateValue := func(fn StateUpdate[T]) {
		state.Update(fn)
	}
	return state, setValue, updateValue
}

// UseMemo creates or retrieves a computed value for the component.
// Similar to React's useMemo, it computes a value based on dependencies,
// and only recomputes when dependencies change.
func UseMemo[T any](
	cs *ComponentState,
	name string,
	computeFn func() T,
	deps []interface{},
) *Signal[T] {
	// First check if signal exists without holding the mutex
	var signal *Signal[T]
	{
		cs.mutex.Lock()
		if existingSignal, ok := cs.signalsByName[name]; ok {
			// Type assertion to ensure it's the correct type
			signal, ok := existingSignal.(*Signal[T])
			if !ok {
				panic(fmt.Sprintf("Type mismatch for memo '%s' in component '%s'",
					name, cs.componentName))
			}
			cs.mutex.Unlock()
			return signal
		}
		cs.mutex.Unlock()
	}

	// Create and register the computed signal outside of the mutex lock
	signal = NewSignal(computeFn())
	{
		cs.mutex.Lock()
		cs.signalsByName[name] = signal
		cs.mutex.Unlock()
	}

	// Register an effect to update the signal when dependencies change
	cs.hookManager.OnUpdate(func(prevDeps []interface{}) error {
		// Recompute the value when dependencies change
		newValue := computeFn()
		signal.Set(newValue)
		return nil
	}, deps)

	return signal
}

// Effect tracking registries
var (
	effectCleanupRegistry = make(map[string]func())      // Tracks effect cleanup functions
	effectInfoRegistry    = make(map[string]*effectInfo) // Tracks effect state
	effectMutex           = sync.RWMutex{}
)

// effectInfo stores information about an effect's state
type effectInfo struct {
	lastRunDeps []interface{}
	isActive    bool // Track if effect is currently active
}

// UseEffect registers an effect that runs when dependencies change.
// Similar to React's useEffect, it runs the effect and clean-up function.
func UseEffect(
	cs *ComponentState,
	name string,
	effectFn func() (cleanup func(), err error),
	deps []interface{},
) {
	// Create a unique key for this effect based on component ID and effect name
	effectKey := fmt.Sprintf("%s_%s_%s", cs.componentID, cs.componentName, name)

	// Make sure we have access to global registries
	effectMutex.Lock()
	defer effectMutex.Unlock()

	// Initialize or get the effect info for this effect
	info, exists := effectInfoRegistry[effectKey]
	if !exists {
		info = &effectInfo{
			lastRunDeps: nil,
			isActive:    false,
		}
		effectInfoRegistry[effectKey] = info
	}

	// Register an update hook that runs the effect when dependencies change
	cs.hookManager.OnUpdate(func(prevDeps []interface{}) error {
		// Lock for thread safety
		effectMutex.Lock()
		defer effectMutex.Unlock()

		// Get the current effect info
		info, exists := effectInfoRegistry[effectKey]
		if !exists {
			// Shouldn't happen, but recreate if missing
			info = &effectInfo{lastRunDeps: nil, isActive: false}
			effectInfoRegistry[effectKey] = info
		}

		// Determine if dependencies have changed
		depsChanged := false

		// If it's the first run or deps length changed, always trigger the effect
		if prevDeps == nil || info.lastRunDeps == nil {
			depsChanged = true
		} else if len(prevDeps) != len(deps) {
			depsChanged = true
		} else {
			// Check each dependency for changes
			for i := range deps {
				if !reflect.DeepEqual(deps[i], prevDeps[i]) {
					depsChanged = true
					break
				}
			}
		}

		// Run cleanup if effect was active, regardless of whether deps changed
		if info.isActive {
			if cleanup, ok := effectCleanupRegistry[effectKey]; ok && cleanup != nil {
				// Execute the cleanup function
				cleanup()
				delete(effectCleanupRegistry, effectKey)
			}
			info.isActive = false
		}

		// Always run the effect on the first execution, or if dependencies changed
		if depsChanged {
			// Store the current deps
			info.lastRunDeps = make([]interface{}, len(deps))
			copy(info.lastRunDeps, deps)

			// Run the effect
			newCleanup, err := effectFn()
			if err != nil {
				return err
			}

			// Mark effect as active and store its cleanup
			info.isActive = true
			if newCleanup != nil {
				effectCleanupRegistry[effectKey] = newCleanup
			}
		}

		return nil
	}, deps)

	// Register an unmount hook to run the cleanup when component unmounts
	cs.hookManager.OnUnmount(func() error {
		// Lock for thread safety
		effectMutex.Lock()
		defer effectMutex.Unlock()

		// Get the current effect info
		info, exists := effectInfoRegistry[effectKey]
		if exists && info.isActive {
			// Execute cleanup if effect is active
			if cleanup, ok := effectCleanupRegistry[effectKey]; ok && cleanup != nil {
				cleanup()
				delete(effectCleanupRegistry, effectKey)
			}

			// Mark as inactive
			info.isActive = false
		}

		// Clean up registry entries
		delete(effectInfoRegistry, effectKey)
		return nil
	})
}

// BatchState executes multiple state updates as a single operation
// This ensures that only a single onChange notification is sent per state
func BatchState(cs *ComponentState, fn func()) {
	// Get a list of all states to batch
	var states []*State[any]

	// Lock the component state registry to safely access states
	cs.mutex.RLock()
	for _, stateObj := range cs.stateRegistry {
		if typedState, ok := stateObj.(*State[any]); ok {
			states = append(states, typedState)
		}
	}
	cs.mutex.RUnlock()

	// Create a coordinator function that will manage the batch operation
	coordinateBatch := func() {
		// Execute the function that contains state updates
		fn()
	}

	// Run everything inside a signal system batch
	Batch(coordinateBatch)
}

// GetStateByName retrieves a state by name, returns nil if not found
func (cs *ComponentState) GetStateByName(name string) interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if state, ok := cs.stateRegistry[name]; ok {
		return state
	}
	return nil
}

// GetSignalByName retrieves a signal by name, returns nil if not found
func (cs *ComponentState) GetSignalByName(name string) interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	if signal, ok := cs.signalsByName[name]; ok {
		return signal
	}
	return nil
}

// GetHookManager returns the hook manager for this component state
func (cs *ComponentState) GetHookManager() *HookManager {
	return cs.hookManager
}

// Dispose cleans up all state resources for the component
func (cs *ComponentState) Dispose() error {
	// Execute unmount hooks
	if err := cs.hookManager.ExecuteUnmountHooks(); err != nil {
		return err
	}

	// Clear all state registries
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.stateRegistry = make(map[string]interface{})
	cs.signalsByName = make(map[string]interface{})

	return nil
}
