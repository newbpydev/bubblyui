package devtools

import (
	"sync"
	"time"
)

// StateHistory tracks changes to reactive state over time.
//
// It maintains a circular buffer of state changes with a configurable maximum size.
// When the buffer is full, the oldest changes are discarded to make room for new ones.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	history := NewStateHistory(1000)
//	history.Record(StateChange{
//	    RefID:     "ref-1",
//	    RefName:   "counter",
//	    OldValue:  41,
//	    NewValue:  42,
//	    Timestamp: time.Now(),
//	    Source:    "increment",
//	})
type StateHistory struct {
	// changes is the circular buffer of state changes
	changes []StateChange

	// maxSize is the maximum number of changes to keep
	maxSize int

	// mu protects concurrent access to changes
	mu sync.RWMutex
}

// StateChange represents a single state mutation.
type StateChange struct {
	// RefID is the unique identifier of the ref that changed
	RefID string

	// RefName is the variable name of the ref
	RefName string

	// OldValue is the value before the change
	OldValue interface{}

	// NewValue is the value after the change
	NewValue interface{}

	// Timestamp is when the change occurred
	Timestamp time.Time

	// Source identifies what caused the change (component ID, function name, etc.)
	Source string
}

// NewStateHistory creates a new state history with the specified maximum size.
//
// The maxSize parameter determines how many state changes to keep in memory.
// When the limit is reached, the oldest changes are discarded.
//
// Example:
//
//	history := NewStateHistory(1000) // Keep last 1000 changes
//
// Parameters:
//   - maxSize: Maximum number of state changes to keep
//
// Returns:
//   - *StateHistory: A new state history instance
func NewStateHistory(maxSize int) *StateHistory {
	return &StateHistory{
		changes: make([]StateChange, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record adds a state change to the history.
//
// If the history is at maximum capacity, the oldest change is removed
// to make room for the new one.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	history.Record(StateChange{
//	    RefID:     "ref-1",
//	    RefName:   "counter",
//	    OldValue:  41,
//	    NewValue:  42,
//	    Timestamp: time.Now(),
//	    Source:    "increment",
//	})
//
// Parameters:
//   - change: The state change to record
func (sh *StateHistory) Record(change StateChange) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.changes = append(sh.changes, change)

	// Keep only last maxSize changes
	if len(sh.changes) > sh.maxSize {
		sh.changes = sh.changes[len(sh.changes)-sh.maxSize:]
	}
}

// GetHistory returns all state changes for a specific ref.
//
// The returned slice is a copy and safe to modify without affecting
// the internal state.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	changes := history.GetHistory("ref-1")
//	for _, change := range changes {
//	    fmt.Printf("%v -> %v\n", change.OldValue, change.NewValue)
//	}
//
// Parameters:
//   - refID: The ref ID to get history for
//
// Returns:
//   - []StateChange: All changes for the specified ref
func (sh *StateHistory) GetHistory(refID string) []StateChange {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	result := make([]StateChange, 0)
	for _, change := range sh.changes {
		if change.RefID == refID {
			result = append(result, change)
		}
	}

	return result
}

// GetAll returns all state changes in the history.
//
// The returned slice is a copy and safe to modify without affecting
// the internal state.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - []StateChange: All state changes
func (sh *StateHistory) GetAll() []StateChange {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	result := make([]StateChange, len(sh.changes))
	copy(result, sh.changes)
	return result
}

// Clear removes all state changes from the history.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (sh *StateHistory) Clear() {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.changes = make([]StateChange, 0, sh.maxSize)
}

// EventLog maintains a log of events that occurred in the application.
//
// It maintains a circular buffer of events with a configurable maximum size.
// When the buffer is full, the oldest events are discarded.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
type EventLog struct {
	// records is the circular buffer of event records
	records []EventRecord

	// maxSize is the maximum number of events to keep
	maxSize int

	// mu protects concurrent access to records
	mu sync.RWMutex
}

// NewEventLog creates a new event log with the specified maximum size.
//
// Parameters:
//   - maxSize: Maximum number of events to keep
//
// Returns:
//   - *EventLog: A new event log instance
func NewEventLog(maxSize int) *EventLog {
	return &EventLog{
		records: make([]EventRecord, 0, maxSize),
		maxSize: maxSize,
	}
}

// Append adds an event to the log.
//
// If the log is at maximum capacity, the oldest event is removed.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - event: The event to append
func (el *EventLog) Append(event EventRecord) {
	el.mu.Lock()
	defer el.mu.Unlock()

	el.records = append(el.records, event)

	// Keep only last maxSize events
	if len(el.records) > el.maxSize {
		el.records = el.records[len(el.records)-el.maxSize:]
	}
}

// GetRecent returns the N most recent events.
//
// If n is greater than the number of events, all events are returned.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - n: Number of recent events to return
//
// Returns:
//   - []EventRecord: The N most recent events
func (el *EventLog) GetRecent(n int) []EventRecord {
	el.mu.RLock()
	defer el.mu.RUnlock()

	if n > len(el.records) {
		n = len(el.records)
	}

	start := len(el.records) - n
	result := make([]EventRecord, n)
	copy(result, el.records[start:])
	return result
}

// Len returns the number of events in the log.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int: Number of events
func (el *EventLog) Len() int {
	el.mu.RLock()
	defer el.mu.RUnlock()
	return len(el.records)
}

// Clear removes all events from the log.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (el *EventLog) Clear() {
	el.mu.Lock()
	defer el.mu.Unlock()
	el.records = make([]EventRecord, 0, el.maxSize)
}

// PerformanceData tracks performance metrics for components.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
type PerformanceData struct {
	// components maps component ID to performance metrics
	components map[string]*ComponentPerformance

	// mu protects concurrent access to components
	mu sync.RWMutex
}

// ComponentPerformance holds performance metrics for a single component.
type ComponentPerformance struct {
	// ComponentID is the unique identifier of the component
	ComponentID string

	// ComponentName is the human-readable name
	ComponentName string

	// RenderCount is the total number of renders
	RenderCount int64

	// TotalRenderTime is the cumulative render time
	TotalRenderTime time.Duration

	// AvgRenderTime is the average render time
	AvgRenderTime time.Duration

	// MaxRenderTime is the slowest render
	MaxRenderTime time.Duration

	// MinRenderTime is the fastest render
	MinRenderTime time.Duration

	// MemoryUsage is the estimated memory usage in bytes
	MemoryUsage uint64

	// LastUpdate is when metrics were last updated
	LastUpdate time.Time
}

// NewPerformanceData creates a new performance data tracker.
//
// Returns:
//   - *PerformanceData: A new performance data instance
func NewPerformanceData() *PerformanceData {
	return &PerformanceData{
		components: make(map[string]*ComponentPerformance),
	}
}

// RecordRender records a component render with its duration.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - componentID: The component ID
//   - componentName: The component name
//   - duration: How long the render took
func (pd *PerformanceData) RecordRender(componentID, componentName string, duration time.Duration) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	perf, exists := pd.components[componentID]
	if !exists {
		perf = &ComponentPerformance{
			ComponentID:   componentID,
			ComponentName: componentName,
			MinRenderTime: duration,
		}
		pd.components[componentID] = perf
	}

	perf.RenderCount++
	perf.TotalRenderTime += duration
	perf.AvgRenderTime = time.Duration(int64(perf.TotalRenderTime) / perf.RenderCount)

	if duration > perf.MaxRenderTime {
		perf.MaxRenderTime = duration
	}
	if duration < perf.MinRenderTime {
		perf.MinRenderTime = duration
	}

	perf.LastUpdate = time.Now()
}

// GetComponent returns performance metrics for a specific component.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - componentID: The component ID
//
// Returns:
//   - *ComponentPerformance: The performance metrics, or nil if not found
func (pd *PerformanceData) GetComponent(componentID string) *ComponentPerformance {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	if perf, exists := pd.components[componentID]; exists {
		// Return a copy to prevent external modification
		copy := *perf
		return &copy
	}
	return nil
}

// GetAll returns performance metrics for all components.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - map[string]*ComponentPerformance: All component performance metrics
func (pd *PerformanceData) GetAll() map[string]*ComponentPerformance {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	result := make(map[string]*ComponentPerformance, len(pd.components))
	for id, perf := range pd.components {
		copy := *perf
		result[id] = &copy
	}
	return result
}

// Clear removes all performance data.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (pd *PerformanceData) Clear() {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.components = make(map[string]*ComponentPerformance)
}

// DevToolsStore holds all collected debug data in memory.
//
// It provides thread-safe storage for component snapshots, state history,
// events, and performance metrics. The store acts as the central data
// repository for the dev tools system.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	store := NewDevToolsStore(1000, 5000)
//
//	// Add component
//	snapshot := &ComponentSnapshot{ID: "comp-1", Name: "Counter"}
//	store.AddComponent(snapshot)
//
//	// Get component
//	comp := store.GetComponent("comp-1")
type DevToolsStore struct {
	// components maps component ID to snapshot
	components map[string]*ComponentSnapshot

	// stateHistory tracks state changes over time
	stateHistory *StateHistory

	// events logs application events
	events *EventLog

	// performance tracks component performance metrics
	performance *PerformanceData

	// mu protects concurrent access to components map
	mu sync.RWMutex
}

// NewDevToolsStore creates a new dev tools store with the specified limits.
//
// Parameters:
//   - maxStateHistory: Maximum number of state changes to keep
//   - maxEvents: Maximum number of events to keep
//
// Returns:
//   - *DevToolsStore: A new store instance
func NewDevToolsStore(maxStateHistory, maxEvents int) *DevToolsStore {
	return &DevToolsStore{
		components:   make(map[string]*ComponentSnapshot),
		stateHistory: NewStateHistory(maxStateHistory),
		events:       NewEventLog(maxEvents),
		performance:  NewPerformanceData(),
	}
}

// AddComponent adds or updates a component snapshot in the store.
//
// If a component with the same ID already exists, it is replaced.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	snapshot := &ComponentSnapshot{
//	    ID:   "comp-1",
//	    Name: "Counter",
//	}
//	store.AddComponent(snapshot)
//
// Parameters:
//   - snapshot: The component snapshot to add
func (s *DevToolsStore) AddComponent(snapshot *ComponentSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.components[snapshot.ID] = snapshot
}

// GetComponent retrieves a component snapshot by ID.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - id: The component ID
//
// Returns:
//   - *ComponentSnapshot: The component snapshot, or nil if not found
func (s *DevToolsStore) GetComponent(id string) *ComponentSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.components[id]
}

// GetAllComponents returns all component snapshots.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - []*ComponentSnapshot: All component snapshots
func (s *DevToolsStore) GetAllComponents() []*ComponentSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ComponentSnapshot, 0, len(s.components))
	for _, snapshot := range s.components {
		result = append(result, snapshot)
	}
	return result
}

// RemoveComponent removes a component snapshot from the store.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - id: The component ID to remove
func (s *DevToolsStore) RemoveComponent(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.components, id)
}

// GetStateHistory returns the state history tracker.
//
// Returns:
//   - *StateHistory: The state history instance
func (s *DevToolsStore) GetStateHistory() *StateHistory {
	return s.stateHistory
}

// GetEventLog returns the event log.
//
// Returns:
//   - *EventLog: The event log instance
func (s *DevToolsStore) GetEventLog() *EventLog {
	return s.events
}

// GetPerformanceData returns the performance data tracker.
//
// Returns:
//   - *PerformanceData: The performance data instance
func (s *DevToolsStore) GetPerformanceData() *PerformanceData {
	return s.performance
}

// Clear removes all data from the store.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (s *DevToolsStore) Clear() {
	s.mu.Lock()
	s.components = make(map[string]*ComponentSnapshot)
	s.mu.Unlock()

	s.stateHistory.Clear()
	s.events.Clear()
	s.performance.Clear()
}
