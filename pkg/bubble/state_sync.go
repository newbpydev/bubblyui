package bubble

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/newbpydev/bubblyui/pkg/core"
)

// StateSynchronizer manages state synchronization between Bubble Tea model and components.
// It provides:
// 1. Bidirectional state sharing between model and components
// 2. State persistence with snapshot creation and restoration
// 3. State versioning and migration strategies
// 4. Consistent state across component tree changes
type StateSynchronizer struct {
	model                *BubbleModel
	registeredComponents map[string]*core.ComponentManager
	sharedStates         map[string]*sharedState
	persistentStates     map[string]bool
	stateVersions        map[string]int
	migrations           map[string]map[int]migrationFunc
	snapshotScheduler    *time.Ticker
	lastSnapshot         time.Time
	mutex                sync.RWMutex
}

// sharedState represents a state that is shared between the model and components
type sharedState struct {
	name         string
	value        interface{}
	valueType    string
	persistent   bool
	version      int
	lastModified time.Time
	mutex        sync.RWMutex
}

// StateEntry represents a single state entry in a state snapshot
type StateEntry struct {
	Type    string          `json:"type"`
	Version int             `json:"version"`
	Data    json.RawMessage `json:"data"`
}

// StateSnapshot represents a complete snapshot of the application state
type StateSnapshot struct {
	Version     int                   `json:"version"`
	Timestamp   time.Time             `json:"timestamp"`
	States      map[string]StateEntry `json:"states"`
	Application string                `json:"application"`
}

// migrationFunc is a function that can migrate state from one version to another
type migrationFunc func(oldData json.RawMessage) (json.RawMessage, error)

// NewStateSynchronizer creates a new state synchronizer for the given model
func NewStateSynchronizer(model *BubbleModel) *StateSynchronizer {
	return &StateSynchronizer{
		model:                model,
		registeredComponents: make(map[string]*core.ComponentManager),
		sharedStates:         make(map[string]*sharedState),
		persistentStates:     make(map[string]bool),
		stateVersions:        make(map[string]int),
		migrations:           make(map[string]map[int]migrationFunc),
		lastSnapshot:         time.Now(),
	}
}

// RegisterComponent registers a component with the synchronizer
func (s *StateSynchronizer) RegisterComponent(component *core.ComponentManager) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.registeredComponents[component.GetName()] = component
}

// UnregisterComponent removes a component from the synchronizer
func (s *StateSynchronizer) UnregisterComponent(componentID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.registeredComponents, componentID)
}

// CreateSharedState is a package-level generic function that creates a new shared state
func CreateSharedState[T any](s *StateSynchronizer, name string, initialValue T) (*core.Signal[T], error) {
	// Create the state without type parameters in the method
	if _, err := s.CreateSharedStateAny(name, initialValue); err != nil {
		return nil, err
	}

	// Create and return a signal of the correct type
	signal := core.NewSignal(initialValue)
	return signal, nil
}

// CreateSharedStateAny creates a new shared state with any value type
func (s *StateSynchronizer) CreateSharedStateAny(name string, initialValue interface{}) (interface{}, error) {
	// First check if the state already exists
	s.mutex.Lock()
	if _, exists := s.sharedStates[name]; exists {
		s.mutex.Unlock()
		return nil, fmt.Errorf("shared state '%s' already exists", name)
	}

	// Create a new shared state
	shState := &sharedState{
		name:         name,
		value:        initialValue,
		valueType:    fmt.Sprintf("%T", initialValue),
		persistent:   false,
		version:      1,
		lastModified: time.Now(),
	}

	s.sharedStates[name] = shState
	s.stateVersions[name] = 1

	// Release the lock before propagating
	s.mutex.Unlock()

	// Sync the state to all components
	err := s.propagateStateToComponents(name)
	return initialValue, err
}

// CreatePersistentState is a package-level generic function that creates a persistent shared state
func CreatePersistentState[T any](s *StateSynchronizer, name string, initialValue T) (*core.Signal[T], error) {
	// Create shared state first
	signal, err := CreateSharedState(s, name, initialValue)
	if err != nil {
		return nil, err
	}

	// Make it persistent
	if err := s.MakePersistent(name); err != nil {
		return nil, err
	}

	return signal, nil
}

// MakePersistent marks an existing state as persistent
func (s *StateSynchronizer) MakePersistent(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, exists := s.sharedStates[name]
	if !exists {
		return fmt.Errorf("shared state '%s' does not exist", name)
	}

	state.persistent = true
	s.persistentStates[name] = true

	return nil
}

// CreatePersistentStateWithVersion is a package-level generic function that creates a persistent state with version
func CreatePersistentStateWithVersion[T any](s *StateSynchronizer, name string, initialValue T, version int) (*core.Signal[T], error) {
	// Create persistent state first
	signal, err := CreatePersistentState(s, name, initialValue)
	if err != nil {
		return nil, err
	}

	// Set the version
	if err := s.SetStateVersion(name, version); err != nil {
		return nil, err
	}

	return signal, nil
}

// SetStateVersion sets the version of an existing state
func (s *StateSynchronizer) SetStateVersion(name string, version int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, exists := s.sharedStates[name]
	if !exists {
		return fmt.Errorf("shared state '%s' does not exist", name)
	}

	state.version = version
	s.stateVersions[name] = version

	return nil
}

// GetSharedState is a package-level generic function that retrieves a shared state value
func GetSharedState[T any](s *StateSynchronizer, name string) (T, error) {
	var zero T

	// Get the state as interface{}
	value, err := s.GetSharedStateAny(name)
	if err != nil {
		return zero, err
	}

	// Type assertion with panic recovery
	defer func() {
		if r := recover(); r != nil {
			// Type assertion failed
		}
	}()

	// Try to convert to the requested type
	typedValue, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("type mismatch for state '%s': expected %T", name, zero)
	}

	return typedValue, nil
}

// GetSharedStateAny retrieves a shared state value as interface{}
func (s *StateSynchronizer) GetSharedStateAny(name string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, exists := s.sharedStates[name]
	if !exists {
		return nil, fmt.Errorf("shared state '%s' does not exist", name)
	}

	state.mutex.RLock()
	defer state.mutex.RUnlock()

	return state.value, nil
}

// SetSharedState updates the value of a shared state
func (s *StateSynchronizer) SetSharedState(name string, value interface{}) error {
	// First get the state under read lock
	s.mutex.RLock()
	state, exists := s.sharedStates[name]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("shared state '%s' does not exist", name)
	}

	// Update the state value with its own lock
	state.mutex.Lock()
	state.value = value
	state.lastModified = time.Now()
	state.mutex.Unlock()

	// Propagate the change to all components
	return s.propagateStateToComponents(name)
}

// GetComponentState is a package-level generic function that retrieves a state value from a component
func GetComponentState[T any](s *StateSynchronizer, component *core.ComponentManager, name string) (T, error) {
	var zero T

	// Get the state as interface{}
	value, err := s.GetComponentStateAny(component, name)
	if err != nil {
		return zero, err
	}

	// Type assertion with panic recovery
	defer func() {
		if r := recover(); r != nil {
			// Type assertion failed
		}
	}()

	// Try to convert to the requested type
	typedValue, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("type mismatch for state '%s' in component '%s': expected %T",
			name, component.GetName(), zero)
	}

	return typedValue, nil
}

// GetComponentStateAny retrieves a shared state value from a component as interface{}
func (s *StateSynchronizer) GetComponentStateAny(component *core.ComponentManager, name string) (interface{}, error) {
	// First check if we have the shared state
	s.mutex.RLock()
	_, exists := s.sharedStates[name]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("shared state '%s' does not exist", name)
	}

	// Then check if the component has the state
	if propValue, hasProp := component.GetProp(name); hasProp {
		return propValue, nil
	}

	// If the component doesn't have the state yet, get it from the shared state
	// and update the component
	value, err := s.GetSharedStateAny(name)
	if err != nil {
		return nil, err
	}

	// Update the component prop
	component.SetProp(name, value)

	return value, nil
}

// SetComponentState updates a shared state from a component
func (s *StateSynchronizer) SetComponentState(component *core.ComponentManager, name string, value interface{}) error {
	// First check if the state exists without holding any locks
	s.mutex.RLock()
	state, exists := s.sharedStates[name]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("shared state '%s' does not exist", name)
	}

	// Update the component prop without holding any locks
	component.SetProp(name, value)

	// Update the shared state value with its own lock
	state.mutex.Lock()
	state.value = value
	state.lastModified = time.Now()
	state.mutex.Unlock()

	// Propagate to other components
	return s.propagateStateToComponents(name)
}

// propagateStateToComponents updates all components with the current value of a shared state
func (s *StateSynchronizer) propagateStateToComponents(stateName string) error {
	// First get the state under a read lock
	s.mutex.RLock()
	state, exists := s.sharedStates[stateName]
	if !exists {
		s.mutex.RUnlock()
		return fmt.Errorf("shared state '%s' does not exist", stateName)
	}

	// Get a copy of all components under read lock
	components := make([]*core.ComponentManager, 0, len(s.registeredComponents))
	for _, component := range s.registeredComponents {
		components = append(components, component)
	}
	s.mutex.RUnlock()

	// Now get the state value under its own read lock
	state.mutex.RLock()
	value := state.value
	state.mutex.RUnlock()

	// Update each component without holding any locks
	for _, component := range components {
		component.SetProp(stateName, value)
	}

	return nil
}

// CreateStateSnapshot creates a snapshot of all persistent states
func (s *StateSynchronizer) CreateStateSnapshot() ([]byte, error) {
	s.mutex.Lock()
	// Create a new snapshot
	snapshot := &StateSnapshot{
		Version:     1,
		Timestamp:   time.Now(),
		States:      make(map[string]StateEntry),
		Application: "BubblyUI", // Application identifier
	}

	// Add all persistent states to the snapshot
	for name, isPersistent := range s.persistentStates {
		if !isPersistent {
			continue
		}

		state, exists := s.sharedStates[name]
		if !exists {
			continue
		}

		state.mutex.RLock()
		value := state.value
		version := state.version
		state.mutex.RUnlock()

		// Serialize the state value
		data, err := json.Marshal(value)
		if err != nil {
			s.mutex.Unlock()
			return nil, fmt.Errorf("failed to serialize state '%s': %w", name, err)
		}

		// Add to snapshot
		snapshot.States[name] = StateEntry{
			Type:    state.valueType,
			Version: version,
			Data:    data,
		}
	}

	s.lastSnapshot = time.Now()
	s.mutex.Unlock()

	// Serialize the snapshot
	return json.Marshal(snapshot)
}

// RestoreFromSnapshot restores state from a snapshot
func (s *StateSynchronizer) RestoreFromSnapshot(snapshotData []byte) error {
	var snapshot StateSnapshot
	if err := json.Unmarshal(snapshotData, &snapshot); err != nil {
		return fmt.Errorf("failed to deserialize snapshot: %w", err)
	}

	// Store state names to propagate after restoration
	statesToPropagate := make([]string, 0, len(snapshot.States))

	// Process each state in the snapshot
	s.mutex.Lock()
	for name, entry := range snapshot.States {
		// Check if we have this state registered
		state, exists := s.sharedStates[name]
		if !exists {
			// Skip states that we don't have registered
			continue
		}

		// Check if we need to migrate the state
		currentVersion := s.stateVersions[name]
		if entry.Version < currentVersion {
			// We need to migrate
			migrated, err := s.migrateState(name, entry.Data, entry.Version, currentVersion)
			if err != nil {
				s.mutex.Unlock()
				return fmt.Errorf("failed to migrate state '%s' from v%d to v%d: %w",
					name, entry.Version, currentVersion, err)
			}
			entry.Data = migrated
			entry.Version = currentVersion
		}

		// Deserialize the state data
		state.mutex.Lock()

		// Create a new instance of the correct type based on the current state value
		// This ensures we use the proper concrete type for unmarshalling
		var value interface{}

		// Handle specific known types
		switch state.valueType {
		case "map[string]int":
			// For map[string]int, unmarshal into that specific type
			var typedValue map[string]int
			if err := json.Unmarshal(entry.Data, &typedValue); err != nil {
				state.mutex.Unlock()
				s.mutex.Unlock()
				return fmt.Errorf("failed to deserialize map[string]int state '%s': %w", name, err)
			}
			value = typedValue
		default:
			// For other types, try generic unmarshal
			if err := json.Unmarshal(entry.Data, &value); err != nil {
				state.mutex.Unlock()
				s.mutex.Unlock()
				return fmt.Errorf("failed to deserialize state '%s': %w", name, err)
			}
		}

		// Update the state
		state.value = value
		state.version = entry.Version
		state.lastModified = time.Now()

		state.mutex.Unlock()

		// Add to states we need to propagate
		statesToPropagate = append(statesToPropagate, name)
	}
	s.mutex.Unlock()

	// Propagate all states to components without holding the main mutex
	for _, name := range statesToPropagate {
		if err := s.propagateStateToComponents(name); err != nil {
			return fmt.Errorf("failed to propagate state '%s' to components: %w", name, err)
		}
	}

	return nil
}

// RegisterMigration registers a migration function for a state
func (s *StateSynchronizer) RegisterMigration(stateName string, fromVersion, toVersion int, migrateFunc migrationFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Initialize the migrations map for this state if needed
	if _, exists := s.migrations[stateName]; !exists {
		s.migrations[stateName] = make(map[int]migrationFunc)
	}

	// Register the migration
	s.migrations[stateName][fromVersion] = migrateFunc
}

// migrateState migrates a state from one version to another
func (s *StateSynchronizer) migrateState(stateName string, data json.RawMessage, fromVersion, toVersion int) (json.RawMessage, error) {
	// Get migrations for this state
	stateMigrations, exists := s.migrations[stateName]
	if !exists {
		return nil, fmt.Errorf("no migrations registered for state '%s'", stateName)
	}

	// Migrate step by step
	currentData := data
	currentVersion := fromVersion

	for currentVersion < toVersion {
		migrate, exists := stateMigrations[currentVersion]
		if !exists {
			return nil, fmt.Errorf("no migration registered from v%d for state '%s'",
				currentVersion, stateName)
		}

		// Apply the migration
		var err error
		currentData, err = migrate(currentData)
		if err != nil {
			return nil, fmt.Errorf("migration failed for state '%s' from v%d: %w",
				stateName, currentVersion, err)
		}

		currentVersion++
	}

	return currentData, nil
}

// EnableAutomaticSnapshots enables automatic state snapshots at the specified interval
func (s *StateSynchronizer) EnableAutomaticSnapshots(interval time.Duration, callback func([]byte)) {
	// Cancel any existing scheduler
	if s.snapshotScheduler != nil {
		s.snapshotScheduler.Stop()
	}

	// Create a new scheduler
	s.snapshotScheduler = time.NewTicker(interval)

	// Start the snapshot goroutine
	go func() {
		for range s.snapshotScheduler.C {
			snapshot, err := s.CreateStateSnapshot()
			if err != nil {
				// Log error or handle it as appropriate
				continue
			}

			// Call the callback with the snapshot
			if callback != nil {
				callback(snapshot)
			}
		}
	}()
}

// DisableAutomaticSnapshots disables automatic state snapshots
func (s *StateSynchronizer) DisableAutomaticSnapshots() {
	if s.snapshotScheduler != nil {
		s.snapshotScheduler.Stop()
		s.snapshotScheduler = nil
	}
}
