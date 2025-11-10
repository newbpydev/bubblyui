package devtools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ExportCheckpoint represents a point in time for incremental exports.
//
// It tracks the highest IDs exported for each data type, allowing subsequent
// incremental exports to export only data created after this checkpoint.
//
// Thread Safety:
//
//	ExportCheckpoint is a value type and safe to use concurrently after creation.
//
// Example:
//
//	checkpoint := ExportCheckpoint{
//	    Timestamp:     time.Now(),
//	    LastEventID:   100,
//	    LastStateID:   50,
//	    LastCommandID: 25,
//	    Version:       "1.0",
//	}
type ExportCheckpoint struct {
	// Timestamp is when this checkpoint was created
	Timestamp time.Time `json:"timestamp"`

	// LastEventID is the highest event ID included in this export
	LastEventID int64 `json:"last_event_id"`

	// LastStateID is the highest state change ID included in this export
	LastStateID int64 `json:"last_state_id"`

	// LastCommandID is the highest command ID included in this export
	LastCommandID int64 `json:"last_command_id"`

	// Version is the export format version
	Version string `json:"version"`
}

// IncrementalExportData represents a delta export containing only new data.
//
// This structure contains only the data that has been created since the
// previous checkpoint, making it much smaller than a full export for
// long-running sessions.
//
// Thread Safety:
//
//	IncrementalExportData is a value type and safe to use concurrently after creation.
//
// Example:
//
//	delta := IncrementalExportData{
//	    Checkpoint:  previousCheckpoint,
//	    NewEvents:   []EventRecord{...},
//	    NewState:    []StateChange{...},
//	    NewCommands: []CommandRecord{...},
//	}
type IncrementalExportData struct {
	// Checkpoint is the previous checkpoint this delta is based on
	Checkpoint ExportCheckpoint `json:"checkpoint"`

	// NewEvents contains events created since the checkpoint
	NewEvents []EventRecord `json:"new_events,omitempty"`

	// NewState contains state changes since the checkpoint
	NewState []StateChange `json:"new_state,omitempty"`

	// NewCommands contains commands since the checkpoint
	NewCommands []CommandRecord `json:"new_commands,omitempty"`
}

// ExportFull writes a complete export and returns a checkpoint for future incremental exports.
//
// This method exports all current data and creates a checkpoint marking the highest
// IDs exported. The checkpoint can be used with ExportIncremental() to export only
// changes since this point.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses read lock on DevTools.
//
// Example:
//
//	opts := ExportOptions{
//	    IncludeComponents: true,
//	    IncludeState:      true,
//	    IncludeEvents:     true,
//	}
//	checkpoint, err := devtools.ExportFull("full-export.json", opts)
//	if err != nil {
//	    log.Printf("Export failed: %v", err)
//	}
//	// Save checkpoint for later incremental exports
//	saveCheckpoint(checkpoint)
//
// Parameters:
//   - filename: Path to the output JSON file
//   - opts: Export options controlling what data to include
//
// Returns:
//   - *ExportCheckpoint: Checkpoint for future incremental exports
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) ExportFull(filename string, opts ExportOptions) (*ExportCheckpoint, error) {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// Check if dev tools is enabled
	if !dt.enabled {
		return nil, fmt.Errorf("dev tools not enabled")
	}

	// Check if store exists
	if dt.store == nil {
		return nil, fmt.Errorf("dev tools store not initialized")
	}

	// Create export data structure
	data := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
	}

	// Collect components if requested
	if opts.IncludeComponents {
		data.Components = dt.store.GetAllComponents()
	}

	// Collect state history if requested
	if opts.IncludeState {
		data.State = dt.store.stateHistory.GetAll()
	}

	// Collect events if requested
	if opts.IncludeEvents {
		data.Events = dt.store.events.GetRecent(dt.store.events.Len())
	}

	// Collect performance data if requested
	if opts.IncludePerformance {
		data.Performance = dt.store.performance
	}

	// Apply sanitization if requested
	if opts.Sanitize {
		data = sanitizeExportData(data, opts.RedactPatterns)
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Write to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write export file: %w", err)
	}

	// Create checkpoint with current max IDs
	checkpoint := &ExportCheckpoint{
		Timestamp:     time.Now(),
		LastEventID:   dt.store.events.GetMaxID(),
		LastStateID:   dt.store.stateHistory.GetMaxID(),
		LastCommandID: dt.store.commands.GetMaxID(),
		Version:       "1.0",
	}

	return checkpoint, nil
}

// ExportIncremental writes only changes since the last checkpoint.
//
// This method exports only data created after the provided checkpoint, resulting
// in much smaller file sizes for long-running sessions. The returned checkpoint
// can be used for the next incremental export.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses read lock on DevTools.
//
// Example:
//
//	// After initial full export
//	checkpoint, _ := devtools.ExportFull("full.json", opts)
//
//	// Later, export only changes
//	newCheckpoint, err := devtools.ExportIncremental("delta1.json", checkpoint)
//	if err != nil {
//	    log.Printf("Incremental export failed: %v", err)
//	}
//
//	// Chain incrementals
//	checkpoint2, _ := devtools.ExportIncremental("delta2.json", newCheckpoint)
//
// Parameters:
//   - filename: Path to the output JSON file
//   - since: Checkpoint from previous export
//
// Returns:
//   - *ExportCheckpoint: New checkpoint for future incremental exports
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) ExportIncremental(filename string, since *ExportCheckpoint) (*ExportCheckpoint, error) {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// Check if dev tools is enabled
	if !dt.enabled {
		return nil, fmt.Errorf("dev tools not enabled")
	}

	// Check if store exists
	if dt.store == nil {
		return nil, fmt.Errorf("dev tools store not initialized")
	}

	// Validate checkpoint
	if since == nil {
		return nil, fmt.Errorf("checkpoint is nil")
	}

	// Get incremental data since checkpoint
	delta, err := dt.store.GetSince(since)
	if err != nil {
		return nil, fmt.Errorf("failed to get incremental data: %w", err)
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(delta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal incremental data: %w", err)
	}

	// Write to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write incremental file: %w", err)
	}

	// Create new checkpoint with current max IDs
	checkpoint := &ExportCheckpoint{
		Timestamp:     time.Now(),
		LastEventID:   dt.store.events.GetMaxID(),
		LastStateID:   dt.store.stateHistory.GetMaxID(),
		LastCommandID: dt.store.commands.GetMaxID(),
		Version:       "1.0",
	}

	return checkpoint, nil
}

// ImportDelta loads incremental data and merges it with existing data.
//
// Unlike Import() which replaces all data, ImportDelta() appends the incremental
// data to the existing store. This allows reconstructing state from a full export
// plus multiple delta files.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses write lock on DevTools.
//
// Example:
//
//	// Import full export first
//	dt.Import("full.json")
//
//	// Then import deltas in order
//	dt.ImportDelta("delta1.json")
//	dt.ImportDelta("delta2.json")
//
//	// State is now reconstructed from full + deltas
//
// Parameters:
//   - filename: Path to the incremental JSON file
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) ImportDelta(filename string) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	// Check if dev tools is enabled
	if !dt.enabled {
		return fmt.Errorf("dev tools not enabled")
	}

	// Check if store exists
	if dt.store == nil {
		return fmt.Errorf("dev tools store not initialized")
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read delta file: %w", err)
	}

	// Unmarshal incremental data
	var delta IncrementalExportData
	err = json.Unmarshal(data, &delta)
	if err != nil {
		return fmt.Errorf("failed to unmarshal delta data: %w", err)
	}

	// Append new events (don't clear existing)
	if delta.NewEvents != nil {
		for _, event := range delta.NewEvents {
			dt.store.events.Append(event)
		}
	}

	// Append new state changes
	if delta.NewState != nil {
		for _, state := range delta.NewState {
			dt.store.stateHistory.Record(state)
		}
	}

	// Append new commands
	if delta.NewCommands != nil {
		for _, cmd := range delta.NewCommands {
			dt.store.commands.Append(cmd)
		}
	}

	return nil
}
