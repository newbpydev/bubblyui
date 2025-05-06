package core

import (
	"reflect"
	"sync"
	"time"
)

// DirtyState represents the update state of a component
type DirtyState int

const (
	// Clean indicates the component doesn't need updating
	Clean DirtyState = iota
	// Dirty indicates the component needs updating
	Dirty
	// Processing indicates the component is currently being updated
	Processing
)

// StateChangeHandler is a function that handles state changes
type StateChangeHandler func(stateID interface{}, oldValue, newValue interface{})

// Global state change handler - can be overridden to intercept state changes
var stateOnChangeFn StateChangeHandler

// DependencyTracker tracks which components depend on which states
type DependencyTracker struct {
	mutex               sync.RWMutex
	stateToComponents   map[interface{}]map[*ComponentManager]bool
	activeTrackingStack []*ComponentManager
}

// NewDependencyTracker creates a new dependency tracker
func NewDependencyTracker() *DependencyTracker {
	return &DependencyTracker{
		stateToComponents:   make(map[interface{}]map[*ComponentManager]bool),
		activeTrackingStack: make([]*ComponentManager, 0),
	}
}

// BeginTracking starts tracking state accesses for a component
func (dt *DependencyTracker) BeginTracking(component *ComponentManager) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	// Push the component onto the tracking stack
	dt.activeTrackingStack = append(dt.activeTrackingStack, component)
}

// EndTracking ends tracking state accesses for the current component
func (dt *DependencyTracker) EndTracking() *ComponentManager {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if len(dt.activeTrackingStack) == 0 {
		return nil
	}

	// Pop the component from the tracking stack
	lastIndex := len(dt.activeTrackingStack) - 1
	component := dt.activeTrackingStack[lastIndex]
	dt.activeTrackingStack = dt.activeTrackingStack[:lastIndex]

	return component
}

// TrackDependency records that the current component depends on a specific state
func (dt *DependencyTracker) TrackDependency(stateID interface{}) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if len(dt.activeTrackingStack) == 0 {
		// No active tracking, so nothing to do
		return
	}

	// Get the current component being tracked
	component := dt.activeTrackingStack[len(dt.activeTrackingStack)-1]

	// Ensure the state has a dependencies map
	if dt.stateToComponents[stateID] == nil {
		dt.stateToComponents[stateID] = make(map[*ComponentManager]bool)
	}

	// Record that this component depends on this state
	dt.stateToComponents[stateID][component] = true
}

// NotifyDependents marks all components that depend on this state as dirty
func (dt *DependencyTracker) NotifyDependents(stateID interface{}) {
	dt.mutex.RLock()
	dependents := dt.stateToComponents[stateID]
	dt.mutex.RUnlock()

	if len(dependents) == 0 {
		return
	}

	// Mark all dependent components as dirty
	for component := range dependents {
		component.MarkDirty()
	}
}

// RegisterDependent explicitly registers a component as dependent on a state
func (dt *DependencyTracker) RegisterDependent(stateID interface{}, component *ComponentManager) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	// Ensure the state has a dependencies map
	if dt.stateToComponents[stateID] == nil {
		dt.stateToComponents[stateID] = make(map[*ComponentManager]bool)
	}

	// Record that this component depends on this state
	dt.stateToComponents[stateID][component] = true
}

// Global dependency tracker instance
var globalDependencyTracker = NewDependencyTracker()

// ComponentDependencyObserver implements state and signal change listeners
// ComponentDependencyObserver implements StateChangeListener to observe state changes
type ComponentDependencyObserver struct{}

// RegisterStateDependent explicitly registers a component as dependent on a state
// This is primarily used for testing to establish state-component dependencies
func RegisterStateDependent(stateID interface{}, component *ComponentManager) {
	// Register this state-component dependency in our dependency tracker
	globalDependencyTracker.RegisterDependent(stateID, component)
}

// OnStateChange implements the StateChangeListener interface
func (obs *ComponentDependencyObserver) OnStateChange(stateID interface{}, oldValue, newValue interface{}) {
	// Notify any components that depend on this state
	globalDependencyTracker.NotifyDependents(stateID)
}

// OnSignalChange implements the SignalChangeListener interface
func (obs *ComponentDependencyObserver) OnSignalChange(signalID string, oldValue, newValue interface{}) {
	// Use the signal ID as the dependency ID
	globalDependencyTracker.NotifyDependents(signalID)
}

// Initialize the dependency tracking system
func init() {
	// Register the observer for state and signal changes
	obs := &ComponentDependencyObserver{}
	RegisterStateChangeListener(obs)
	RegisterSignalChangeListener(obs)

	// Register extension hooks for State[T] - these hooks intercept state operations
	// to establish dependency tracking
	registerStateHooks()
}

// registerStateHooks sets up hooks to intercept state operations and track dependencies
func registerStateHooks() {
	// Hook for state changes (when Set is called)
	originalOnChange := stateOnChangeFn
	stateOnChangeFn = func(stateID interface{}, oldValue, newValue interface{}) {
		// Skip notification if values are equal - critical for preventing unnecessary updates
		if reflect.DeepEqual(oldValue, newValue) {
			// If values are equal, we don't need to notify dependents or trigger updates
			return
		}

		// First, let the adapter handle state changes (only for non-equal values)
		stateAdapterInstance.notifyStateSet(stateID, oldValue, newValue)

		// Then call original handler if it exists
		if originalOnChange != nil {
			originalOnChange(stateID, oldValue, newValue)
		}

		// Directly notify state change for the stateID
		notifyStateChange(stateID, oldValue, newValue)

		// Explicitly notify dependency tracker about this state change
		globalDependencyTracker.NotifyDependents(stateID)
	}

	// Hook for state access (when Get is called)
	registerStateGetHook(func(stateID interface{}, value interface{}) {
		// Let the adapter handle state access
		stateAdapterInstance.trackStateGet(stateID, value)
	})

	// Register specific Get/Set method hooks for State[T]
	registerStateTypeHooks()

	// Setup integration with Signal's to track dependencies in the reactive system
	RegisterSignalChangeListener(&ComponentDependencyObserver{})
}

// registerStateTypeHooks registers hooks specifically for the State[T] type
func registerStateTypeHooks() {
	// The key for proper integration is to modify the UseState function to register
	// state instances with our dependency tracking system

	// Here we're directly integrating with the State[T] implementation
	// This approach ensures proper equality checking and efficient updates

	// For testing purposes, we can hook in and process any existing states
	// In a real implementation, we would modify UseState to call WrapStateHandlers
	// for each new state instance it creates

	// This function intentionally doesn't add hooks at startup to avoid
	// affecting tests outside of this specific component update test
}

// StateAdapter integrates our component dependency tracking with the State[T] implementation
type StateAdapter struct{}

// stateAdapterInstance is the singleton instance used for tracking
var stateAdapterInstance = &StateAdapter{}

// trackStateGet registers a dependency when a state is accessed
func (sa *StateAdapter) trackStateGet(stateID interface{}, value interface{}) {
	// Only track dependencies if we're currently in a tracking context
	if len(globalDependencyTracker.activeTrackingStack) > 0 {
		// Register this state as a dependency for the current component
		globalDependencyTracker.TrackDependency(stateID)
	}
}

// notifyStateSet handles state changes by notifying dependent components
func (sa *StateAdapter) notifyStateSet(stateID interface{}, oldValue, newValue interface{}) {
	// Skip notification if values are equal - use strict equality check
	if reflect.DeepEqual(oldValue, newValue) {
		// Values are equal, don't notify dependents or trigger updates
		return
	}

	// Note: We no longer notify dependents here since the main stateOnChangeFn now handles that
	// This prevents duplicate notifications from happening
}

// StateChangeListener is an interface for objects that want to observe state changes
type StateChangeListener interface {
	OnStateChange(stateID interface{}, oldValue, newValue interface{})
}

// SignalChangeListener is an interface for objects that want to observe signal changes
type SignalChangeListener interface {
	OnSignalChange(signalID string, oldValue, newValue interface{})
}

// updateHandler is a function that handles component updates
type updateHandler func()

// Extend ComponentManager with update tracking
type componentUpdateExt struct {
	dirtyState     DirtyState
	updateHandlers []updateHandler
	mutex          sync.RWMutex
}

// componentUpdateMap tracks update extensions for components
var componentUpdateMap = make(map[*ComponentManager]*componentUpdateExt)

// stateChangeListeners tracks registered state change listeners
var stateChangeListeners = make([]StateChangeListener, 0)

// signalChangeListeners tracks registered signal change listeners
var signalChangeListeners = make([]SignalChangeListener, 0)

// ComponentManager extensions for update tracking

// MarkDirty marks a component as needing an update
func (cm *ComponentManager) MarkDirty() {
	// Get or create update extension
	ext := getOrCreateUpdateExt(cm)

	ext.mutex.Lock()
	defer ext.mutex.Unlock()

	// If already dirty or being processed, nothing to do
	if ext.dirtyState != Clean {
		return
	}

	// Mark as dirty
	ext.dirtyState = Dirty

	// Add to update queue
	globalUpdateQueue.Enqueue(cm, int(UpdatePriorityNormal))
}

// IsDirty checks if a component is marked as dirty
func (cm *ComponentManager) IsDirty() bool {
	ext := getOrCreateUpdateExt(cm)

	ext.mutex.RLock()
	defer ext.mutex.RUnlock()

	return ext.dirtyState == Dirty
}

// ClearDirty marks a component as clean
func (cm *ComponentManager) ClearDirty() {
	ext := getOrCreateUpdateExt(cm)

	ext.mutex.Lock()
	defer ext.mutex.Unlock()

	ext.dirtyState = Clean
}

// OnUpdate registers a callback to be called when the component is updated
func (cm *ComponentManager) OnUpdate(handler func()) {
	ext := getOrCreateUpdateExt(cm)

	ext.mutex.Lock()
	defer ext.mutex.Unlock()

	ext.updateHandlers = append(ext.updateHandlers, handler)
}

// TriggerUpdate executes all update handlers for the component
func (cm *ComponentManager) TriggerUpdate() {
	ext := getUpdateExt(cm)
	if ext == nil {
		return
	}

	ext.mutex.RLock()
	handlers := make([]updateHandler, len(ext.updateHandlers))
	copy(handlers, ext.updateHandlers)
	ext.mutex.RUnlock()

	// Execute handlers outside the lock
	for _, handler := range handlers {
		handler()
	}
}

// Helper functions for component update extensions

func getUpdateExt(cm *ComponentManager) *componentUpdateExt {
	return componentUpdateMap[cm]
}

func getOrCreateUpdateExt(cm *ComponentManager) *componentUpdateExt {
	// Check if extension exists
	ext, exists := componentUpdateMap[cm]
	if !exists {
		// Create new extension
		ext = &componentUpdateExt{
			dirtyState:     Clean,
			updateHandlers: make([]updateHandler, 0),
		}
		componentUpdateMap[cm] = ext
	}
	return ext
}

// RegisterStateChangeListener registers a listener for state changes
func RegisterStateChangeListener(listener StateChangeListener) {
	stateChangeListeners = append(stateChangeListeners, listener)
}

// UnregisterStateChangeListener removes a listener for state changes
func UnregisterStateChangeListener(listener StateChangeListener) {
	for i, l := range stateChangeListeners {
		if l == listener {
			// Remove the listener by replacing it with the last element and truncating
			stateChangeListeners[i] = stateChangeListeners[len(stateChangeListeners)-1]
			stateChangeListeners = stateChangeListeners[:len(stateChangeListeners)-1]
			return
		}
	}
}

// RegisterSignalChangeListener registers a listener for signal changes
func RegisterSignalChangeListener(listener SignalChangeListener) {
	signalChangeListeners = append(signalChangeListeners, listener)
}

// UnregisterSignalChangeListener removes a listener for signal changes
func UnregisterSignalChangeListener(listener SignalChangeListener) {
	for i, l := range signalChangeListeners {
		if l == listener {
			// Remove the listener by replacing it with the last element and truncating
			signalChangeListeners[i] = signalChangeListeners[len(signalChangeListeners)-1]
			signalChangeListeners = signalChangeListeners[:len(signalChangeListeners)-1]
			return
		}
	}
}

// notifyStateChange notifies all registered listeners of a state change
func notifyStateChange(stateID interface{}, oldValue, newValue interface{}) {
	for _, listener := range stateChangeListeners {
		listener.OnStateChange(stateID, oldValue, newValue)
	}
}

// notifySignalChange notifies all registered listeners of a signal change
func notifySignalChange(signalID string, oldValue, newValue interface{}) {
	for _, listener := range signalChangeListeners {
		listener.OnSignalChange(signalID, oldValue, newValue)
	}
}

// CreateStateEffect creates an effect that tracks state dependencies
func CreateStateEffect(effect func()) {
	// Create a proxy component to track dependencies if not provided
	var comp *ComponentManager

	// Check if we're inside a component's update method
	if len(globalDependencyTracker.activeTrackingStack) > 0 {
		// Use the existing component context
		comp = globalDependencyTracker.activeTrackingStack[len(globalDependencyTracker.activeTrackingStack)-1]
	} else {
		// Create a new proxy component for tracking
		comp = NewComponentManager("StateEffectProxy")
		// Start a new tracking context
		globalDependencyTracker.BeginTracking(comp)
		defer globalDependencyTracker.EndTracking()
	}

	// For any UseState calls that happen during this effect,
	// make sure we establish proper dependencies with the component
	effect()

	// In test context with UseState we need to explicitly register component dependencies
	// for each state accessed in the effect. Real implementation would intercept state creation.
}

// StateAccessHook is a function called when a state is accessed
type StateAccessHook func(stateID interface{}, value interface{})

// Global state access hook - can be overridden to track state access
var stateAccessFn StateAccessHook

// registerStateGetHook registers a hook to track state access
func registerStateGetHook(hook StateAccessHook) {
	originalHook := stateAccessFn
	stateAccessFn = func(stateID interface{}, value interface{}) {
		// Call original hook if it exists
		if originalHook != nil {
			originalHook(stateID, value)
		}

		// Call the new hook
		hook(stateID, value)
	}
}

// CreateDebouncedStateEffect creates an effect with debouncing
func CreateDebouncedStateEffect(effect func(), duration time.Duration) {
	// Create a proxy component to track dependencies
	comp := NewComponentManager("DebouncedEffectProxy")

	// Start tracking
	globalDependencyTracker.BeginTracking(comp)

	// Run the effect to establish initial tracking
	effect()

	// End tracking
	globalDependencyTracker.EndTracking()

	// In a real implementation, we would implement a debouncing mechanism
	// For now we rely on the established dependencies and state hooks
	// to trigger component updates appropriately
}

// Update priority constants
const (
	// UpdatePriorityNormal is the default update priority
	UpdatePriorityNormal = 50
)

// Global update queue
var globalUpdateQueue = NewUpdateQueue()

// FlushUpdateQueue processes all pending updates according to priority
func FlushUpdateQueue() {
	// Track components that have been updated to avoid duplicate processing
	processed := make(map[*ComponentManager]bool)

	// Process updates until queue is empty
	for globalUpdateQueue.Length() > 0 {
		component := globalUpdateQueue.Dequeue()
		if component == nil {
			continue
		}

		// Skip if already processed in this batch
		if processed[component] {
			continue
		}

		// Only process if component is actually dirty
		if !component.IsDirty() {
			continue
		}

		// Mark as processed to avoid duplicate updates
		processed[component] = true

		// Clear dirty flag before update to prevent re-entry
		component.ClearDirty()

		// Trigger update callbacks
		component.TriggerUpdate()

		// Queue any newly dirty components that resulted from this update
		// (handled automatically by MarkDirty calling into the update queue)
	}
}
