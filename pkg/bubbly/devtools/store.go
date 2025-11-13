package devtools

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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

	// nextID is the auto-incrementing ID counter for state changes
	nextID int64

	// mu protects concurrent access to changes
	mu sync.RWMutex
}

// StateChange represents a single state mutation.
type StateChange struct {
	// ID is the auto-incrementing unique identifier for this state change
	ID int64 `json:"id"`

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
// to make room for the new one. An auto-incrementing ID is assigned to
// the state change for incremental export tracking.
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

	// Assign auto-incrementing ID
	change.ID = atomic.AddInt64(&sh.nextID, 1)

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

// GetMaxID returns the highest ID assigned to any state change.
//
// This is used for creating checkpoints in incremental exports.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int64: The highest ID, or 0 if no state changes exist
func (sh *StateHistory) GetMaxID() int64 {
	return atomic.LoadInt64(&sh.nextID)
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

	// nextID is the auto-incrementing ID counter for events
	nextID int64

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
// An auto-incrementing sequence ID is assigned to the event for
// incremental export tracking.
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

	// Assign auto-incrementing sequence ID
	event.SeqID = atomic.AddInt64(&el.nextID, 1)

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

// GetMaxID returns the highest sequence ID assigned to any event.
//
// This is used for creating checkpoints in incremental exports.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - int64: The highest sequence ID, or 0 if no events exist
func (el *EventLog) GetMaxID() int64 {
	return atomic.LoadInt64(&el.nextID)
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
// events, performance metrics, and command timeline. The store acts as the
// central data repository for the dev tools system.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	store := NewDevToolsStore(1000, 5000, 1000)
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

	// componentRefs maps component ID to its owned ref IDs
	componentRefs map[string][]string

	// refOwners maps ref ID to owning component ID
	refOwners map[string]string

	// componentTree maps parent component ID to child IDs
	componentTree map[string][]string

	// componentParent maps child component ID to parent ID
	componentParent map[string]string

	// stateHistory tracks state changes over time
	stateHistory *StateHistory

	// events logs application events
	events *EventLog

	// performance tracks component performance metrics
	performance *PerformanceData

	// commands tracks command execution timeline
	commands *CommandTimeline

	// mu protects concurrent access to all maps
	mu sync.RWMutex
}

// NewDevToolsStore creates a new dev tools store with the specified limits.
//
// Parameters:
//   - maxStateHistory: Maximum number of state changes to keep
//   - maxEvents: Maximum number of events to keep
//   - maxCommands: Maximum number of commands to keep
//
// Returns:
//   - *DevToolsStore: A new store instance
func NewDevToolsStore(maxStateHistory, maxEvents, maxCommands int) *DevToolsStore {
	return &DevToolsStore{
		components:      make(map[string]*ComponentSnapshot),
		componentRefs:   make(map[string][]string),
		refOwners:       make(map[string]string),
		componentTree:   make(map[string][]string),
		componentParent: make(map[string]string),
		stateHistory:    NewStateHistory(maxStateHistory),
		events:          NewEventLog(maxEvents),
		performance:     NewPerformanceData(),
		commands:        NewCommandTimeline(maxCommands),
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
	
	// Clean up ownership tracking
	delete(s.componentRefs, id)
	delete(s.componentParent, id)
	delete(s.componentTree, id)
}

// RegisterRefOwner registers that a specific component owns a specific ref.
// This enables proper tracking of which refs belong to which components.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - componentID: The component that owns the ref
//   - refID: The ref ID being owned
func (s *DevToolsStore) RegisterRefOwner(componentID, refID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Add ref to component's ref list
	if s.componentRefs[componentID] == nil {
		s.componentRefs[componentID] = make([]string, 0)
	}
	
	// Check if already registered
	for _, existingRef := range s.componentRefs[componentID] {
		if existingRef == refID {
			return
		}
	}
	
	s.componentRefs[componentID] = append(s.componentRefs[componentID], refID)
	s.refOwners[refID] = componentID
	
	// CRITICAL FIX: Add ref to component snapshot immediately
	// Previously, refs only appeared after value changed (OnRefChange hook)
	// This ensures refs are visible as soon as they're exposed
	comp, exists := s.components[componentID]
	if exists {
		// Check if ref already in snapshot
		for _, ref := range comp.Refs {
			if ref.ID == refID {
				return
			}
		}
		
		// Add ref to snapshot with initial value (will be updated on first change)
		refName := extractRefName(refID)
		comp.Refs = append(comp.Refs, &RefSnapshot{
			ID:    refID,
			Name:  refName,
			Value: nil,  // Initial value unknown until first ref.Set()
			Type:  "unknown",
		})
	}
}

// UpdateRefValue updates a ref value only for the component that owns it.
// Returns the owning component ID, or empty string if ref has no owner.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - refID: The ref to update
//   - newValue: The new value
//
// Returns:
//   - string: The owning component ID
//   - bool: Whether the update was applied
func (s *DevToolsStore) UpdateRefValue(refID string, newValue interface{}) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Find the component that owns this ref
	ownerID, exists := s.refOwners[refID]
	if !exists {
		return "", false
	}
	
	// Update the component's ref value
	comp, exists := s.components[ownerID]
	if !exists {
		return ownerID, false
	}
	
	// Update existing ref or add new one
	refUpdated := false
	for i, ref := range comp.Refs {
		if ref.ID == refID {
			comp.Refs[i].Value = newValue
			comp.Refs[i].Type = fmt.Sprintf("%T", newValue)  // Update type too
			refUpdated = true
			break
		}
	}
	
	if !refUpdated {
		// Add new ref
		refName := extractRefName(refID)
		comp.Refs = append(comp.Refs, &RefSnapshot{
			ID:    refID,
			Name:  refName,
			Value: newValue,
			Type:  fmt.Sprintf("%T", newValue),
		})
	}
	
	// Also update State map
	if comp.State == nil {
		comp.State = make(map[string]interface{})
	}
	comp.State[refID] = newValue
	comp.Timestamp = time.Now()
	
	return ownerID, true
}

// AddComponentChild adds a child-parent relationship in the component tree.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - parentID: The parent component ID
//   - childID: The child component ID
func (s *DevToolsStore) AddComponentChild(parentID, childID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Initialize parent's children list if needed
	if s.componentTree[parentID] == nil {
		s.componentTree[parentID] = make([]string, 0)
	}
	
	// Check if already added
	for _, existing := range s.componentTree[parentID] {
		if existing == childID {
			return
		}
	}
	
	s.componentTree[parentID] = append(s.componentTree[parentID], childID)
	s.componentParent[childID] = parentID
	
	// Update component snapshot's Children field
	if parent, exists := s.components[parentID]; exists {
		if child, exists := s.components[childID]; exists {
			if parent.Children == nil {
				parent.Children = make([]*ComponentSnapshot, 0)
			}
			parent.Children = append(parent.Children, child)
		}
	}
}

// RemoveComponentChild removes a child-parent relationship from the component tree.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - parentID: The parent component ID
//   - childID: The child component ID to remove
func (s *DevToolsStore) RemoveComponentChild(parentID, childID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Remove from parent's children list
	if children, exists := s.componentTree[parentID]; exists {
		for i, id := range children {
			if id == childID {
				s.componentTree[parentID] = append(children[:i], children[i+1:]...)
				break
			}
		}
	}
	
	// Remove parent reference
	delete(s.componentParent, childID)
	
	// Update component snapshot's Children field
	if parent, exists := s.components[parentID]; exists {
		if parent.Children != nil {
			for i, child := range parent.Children {
				if child.ID == childID {
					parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
					break
				}
			}
		}
	}
}

// GetComponentChildren returns the IDs of a component's direct children.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Parameters:
//   - componentID: The parent component ID
//
// Returns:
//   - []string: List of child component IDs
func (s *DevToolsStore) GetComponentChildren(componentID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	children := s.componentTree[componentID]
	if children == nil {
		return []string{}
	}
	
	// Return a copy
	result := make([]string, len(children))
	copy(result, children)
	return result
}

// GetRootComponents returns components that have no parent.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Returns:
//   - []*ComponentSnapshot: List of root components
func (s *DevToolsStore) GetRootComponents() []*ComponentSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	roots := make([]*ComponentSnapshot, 0)
	for id, comp := range s.components {
		// Check if this component has no parent
		if _, hasParent := s.componentParent[id]; !hasParent {
			roots = append(roots, comp)
		}
	}
	return roots
}

// extractRefName extracts a simple name from a ref ID.
// Example: "ref-0x123abc" -> "ref", "count-ref-0x456" -> "count"
func extractRefName(refID string) string {
	// Split on "-ref-" or "-0x" to get the prefix
	if idx := strings.Index(refID, "-ref-"); idx >= 0 {
		return refID[:idx]
	}
	if idx := strings.Index(refID, "-0x"); idx >= 0 {
		return refID[:idx]
	}
	// If no pattern matches, return the full ID
	return refID
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

// GetSince returns incremental data since the given checkpoint.
//
// This method filters events, state changes, and commands to include only
// those with IDs greater than the checkpoint's last IDs. This enables
// incremental exports that contain only new data.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses read locks on store.
//
// Example:
//
//	checkpoint := &ExportCheckpoint{
//	    LastEventID:   100,
//	    LastStateID:   50,
//	    LastCommandID: 25,
//	}
//	delta, err := store.GetSince(checkpoint)
//	if err != nil {
//	    log.Printf("Failed to get delta: %v", err)
//	}
//
// Parameters:
//   - checkpoint: The checkpoint to filter from
//
// Returns:
//   - *IncrementalExportData: The filtered incremental data
//   - error: nil on success, error describing the failure otherwise
func (s *DevToolsStore) GetSince(checkpoint *ExportCheckpoint) (*IncrementalExportData, error) {
	if checkpoint == nil {
		return nil, fmt.Errorf("checkpoint is nil")
	}

	delta := &IncrementalExportData{
		Checkpoint: *checkpoint,
	}

	// Filter events by sequence ID
	allEvents := s.events.GetRecent(s.events.Len())
	for _, event := range allEvents {
		if event.SeqID > checkpoint.LastEventID {
			delta.NewEvents = append(delta.NewEvents, event)
		}
	}

	// Filter state changes by ID
	allState := s.stateHistory.GetAll()
	for _, state := range allState {
		if state.ID > checkpoint.LastStateID {
			delta.NewState = append(delta.NewState, state)
		}
	}

	// Filter commands by sequence ID
	allCommands := s.commands.GetAll()
	for _, cmd := range allCommands {
		if cmd.SeqID > checkpoint.LastCommandID {
			delta.NewCommands = append(delta.NewCommands, cmd)
		}
	}

	return delta, nil
}
