package core

import (
	"reflect"
	"sync"
)

// StateEqualityChecker provides a generic interface for comparing state values
type StateEqualityChecker interface {
	AreEqual(oldValue, newValue interface{}) bool
}

// DefaultStateEqualityChecker uses deep equality for state value comparison
type DefaultStateEqualityChecker struct{}

// AreEqual compares two state values for equality
func (c *DefaultStateEqualityChecker) AreEqual(oldValue, newValue interface{}) bool {
	return reflect.DeepEqual(oldValue, newValue)
}

// globalEqualityChecker is the default equality checker used for state comparisons
var globalEqualityChecker StateEqualityChecker = &DefaultStateEqualityChecker{}

// SetStateEqualityChecker replaces the global equality checker
func SetStateEqualityChecker(checker StateEqualityChecker) {
	if checker != nil {
		globalEqualityChecker = checker
	}
}

// StateIntegrationManager manages the integration between states and component updates
type StateIntegrationManager struct {
	equalityCheckersByState map[interface{}]StateEqualityChecker
	valuesByState           map[interface{}]interface{}
	mutex                   sync.RWMutex
}

// NewStateIntegrationManager creates a new state integration manager
func NewStateIntegrationManager() *StateIntegrationManager {
	return &StateIntegrationManager{
		equalityCheckersByState: make(map[interface{}]StateEqualityChecker),
		valuesByState:           make(map[interface{}]interface{}),
	}
}

// RegisterState registers a state with an optional custom equality checker
func (sim *StateIntegrationManager) RegisterState(stateID interface{}, initialValue interface{}, checker StateEqualityChecker) {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	// Store the initial value
	sim.valuesByState[stateID] = initialValue

	// Store the equality checker if provided, otherwise use the global one
	if checker != nil {
		sim.equalityCheckersByState[stateID] = checker
	}
}

// ShouldNotifyChange checks if a state change should trigger component updates
func (sim *StateIntegrationManager) ShouldNotifyChange(stateID interface{}, oldValue, newValue interface{}) bool {
	sim.mutex.RLock()
	defer sim.mutex.RUnlock()

	// Get the appropriate equality checker
	checker, hasCustom := sim.equalityCheckersByState[stateID]
	if !hasCustom {
		checker = globalEqualityChecker
	}

	// Skip notification if values are equal according to the checker
	return !checker.AreEqual(oldValue, newValue)
}

// globalStateIntegration is the singleton instance used for state integration
var globalStateIntegration = NewStateIntegrationManager()

// WrapStateHandlers wraps a state's handlers to integrate with component updates
func WrapStateHandlers[T any](state *State[T], stateID interface{}) {
	// Register the state with its initial value
	initialValue := state.Get()
	var equalityChecker StateEqualityChecker = nil

	// If state has an equality function, create a custom checker for it
	if signal := state.GetSignal(); signal != nil {
		// Note: Signal interface may need to be extended to expose equality function
		// For now, we'll use the default equality checker
		equalityChecker = &DefaultStateEqualityChecker{}
	}

	globalStateIntegration.RegisterState(stateID, initialValue, equalityChecker)

	// Hook into the state's onChange events to track dependencies
	state.OnChange(func(oldValue, newValue T) {
		if globalStateIntegration.ShouldNotifyChange(stateID, oldValue, newValue) {
			// Trigger dependency notifications
			globalDependencyTracker.NotifyDependents(stateID)
		}
	})
}

// CustomEqualityFunction is a type for user-provided equality functions
type CustomEqualityFunction[T any] func(a, b T) bool

// TypedEqualityChecker provides a way to compare typed values with a custom equality function
type TypedEqualityChecker[T any] struct {
	equalsFn CustomEqualityFunction[T]
}

// NewTypedEqualityChecker creates a typed equality checker with a custom function
func NewTypedEqualityChecker[T any](equalsFn CustomEqualityFunction[T]) *TypedEqualityChecker[T] {
	return &TypedEqualityChecker[T]{
		equalsFn: equalsFn,
	}
}

// AreEqual implements StateEqualityChecker for typed values
func (c *TypedEqualityChecker[T]) AreEqual(oldValue, newValue interface{}) bool {
	oldTyped, ok1 := oldValue.(T)
	newTyped, ok2 := newValue.(T)

	if !ok1 || !ok2 {
		// If type assertion fails, fall back to deep equality
		return reflect.DeepEqual(oldValue, newValue)
	}

	// Use the custom equality function
	return c.equalsFn(oldTyped, newTyped)
}
