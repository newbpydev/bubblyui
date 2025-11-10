package devtools

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExportFull_ReturnsCheckpoint tests that full export returns a valid checkpoint
func TestExportFull_ReturnsCheckpoint(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Add some test data
	dt.store.events.Append(EventRecord{
		ID:        "event-1",
		Name:      "click",
		Timestamp: time.Now(),
	})
	dt.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  0,
		NewValue:  1,
		Timestamp: time.Now(),
	})
	dt.store.commands.RecordCommand(CommandRecord{
		ID:        "cmd-1",
		Type:      "fetch",
		Generated: time.Now(),
	})

	// Export full
	filename := "test_full_export.json"
	defer os.Remove(filename)

	opts := ExportOptions{
		IncludeEvents: true,
		IncludeState:  true,
	}

	checkpoint, err := dt.ExportFull(filename, opts)

	require.NoError(t, err)
	require.NotNil(t, checkpoint)
	assert.Equal(t, "1.0", checkpoint.Version)
	assert.False(t, checkpoint.Timestamp.IsZero())
	assert.Greater(t, checkpoint.LastEventID, int64(0))
	assert.Greater(t, checkpoint.LastStateID, int64(0))
	assert.Greater(t, checkpoint.LastCommandID, int64(0))

	// Verify file exists
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

// TestExportIncremental_IncludesOnlyNewData tests incremental export filtering
func TestExportIncremental_IncludesOnlyNewData(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Add initial data
	dt.store.events.Append(EventRecord{
		ID:        "event-1",
		Name:      "click",
		Timestamp: time.Now(),
	})
	dt.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  0,
		NewValue:  1,
		Timestamp: time.Now(),
	})

	// Export full
	fullFile := "test_full.json"
	defer os.Remove(fullFile)

	checkpoint1, err := dt.ExportFull(fullFile, ExportOptions{
		IncludeEvents: true,
		IncludeState:  true,
	})
	require.NoError(t, err)

	// Add more data after checkpoint
	dt.store.events.Append(EventRecord{
		ID:        "event-2",
		Name:      "submit",
		Timestamp: time.Now(),
	})
	dt.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  1,
		NewValue:  2,
		Timestamp: time.Now(),
	})

	// Export incremental
	deltaFile := "test_delta.json"
	defer os.Remove(deltaFile)

	checkpoint2, err := dt.ExportIncremental(deltaFile, checkpoint1)
	require.NoError(t, err)
	require.NotNil(t, checkpoint2)

	// Read delta file and verify it only contains new data
	data, err := os.ReadFile(deltaFile)
	require.NoError(t, err)

	var delta IncrementalExportData
	err = json.Unmarshal(data, &delta)
	require.NoError(t, err)

	// Should have exactly 1 new event and 1 new state change
	assert.Len(t, delta.NewEvents, 1)
	assert.Equal(t, "event-2", delta.NewEvents[0].ID)

	assert.Len(t, delta.NewState, 1)
	assert.Equal(t, float64(2), delta.NewState[0].NewValue) // JSON unmarshals numbers as float64

	// Checkpoint should be updated
	assert.Greater(t, checkpoint2.LastEventID, checkpoint1.LastEventID)
	assert.Greater(t, checkpoint2.LastStateID, checkpoint1.LastStateID)
}

// TestExportIncremental_MultipleChain tests chaining multiple incrementals
func TestExportIncremental_MultipleChain(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Initial full export
	fullFile := "test_chain_full.json"
	defer os.Remove(fullFile)

	checkpoint1, err := dt.ExportFull(fullFile, ExportOptions{
		IncludeEvents: true,
	})
	require.NoError(t, err)

	// First incremental
	dt.store.events.Append(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})

	delta1File := "test_chain_delta1.json"
	defer os.Remove(delta1File)

	checkpoint2, err := dt.ExportIncremental(delta1File, checkpoint1)
	require.NoError(t, err)

	// Second incremental
	dt.store.events.Append(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})

	delta2File := "test_chain_delta2.json"
	defer os.Remove(delta2File)

	checkpoint3, err := dt.ExportIncremental(delta2File, checkpoint2)
	require.NoError(t, err)

	// Verify checkpoints are increasing
	assert.Less(t, checkpoint1.LastEventID, checkpoint2.LastEventID)
	assert.Less(t, checkpoint2.LastEventID, checkpoint3.LastEventID)

	// Verify each delta has exactly 1 event
	data1, _ := os.ReadFile(delta1File)
	var delta1 IncrementalExportData
	json.Unmarshal(data1, &delta1)
	assert.Len(t, delta1.NewEvents, 1)

	data2, _ := os.ReadFile(delta2File)
	var delta2 IncrementalExportData
	json.Unmarshal(data2, &delta2)
	assert.Len(t, delta2.NewEvents, 1)
}

// TestImportDelta_AppendsData tests that ImportDelta appends without replacing
func TestImportDelta_AppendsData(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Add initial data
	dt.store.events.Append(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})

	initialCount := dt.store.events.Len()

	// Create a delta file manually
	delta := IncrementalExportData{
		Checkpoint: ExportCheckpoint{
			Timestamp:   time.Now(),
			LastEventID: 1,
			Version:     "1.0",
		},
		NewEvents: []EventRecord{
			{SeqID: 2, ID: "event-2", Name: "submit", Timestamp: time.Now()},
			{SeqID: 3, ID: "event-3", Name: "change", Timestamp: time.Now()},
		},
	}

	deltaFile := "test_import_delta.json"
	defer os.Remove(deltaFile)

	data, _ := json.MarshalIndent(delta, "", "  ")
	os.WriteFile(deltaFile, data, 0644)

	// Import delta
	err := dt.ImportDelta(deltaFile)
	require.NoError(t, err)

	// Verify data was appended, not replaced
	finalCount := dt.store.events.Len()
	assert.Equal(t, initialCount+2, finalCount)

	// Verify the new events are present
	allEvents := dt.store.events.GetRecent(finalCount)
	eventIDs := make([]string, len(allEvents))
	for i, e := range allEvents {
		eventIDs[i] = e.ID
	}
	assert.Contains(t, eventIDs, "event-1") // Original
	assert.Contains(t, eventIDs, "event-2") // From delta
	assert.Contains(t, eventIDs, "event-3") // From delta
}

// TestExportIncremental_EmptyDelta tests handling of no new data
func TestExportIncremental_EmptyDelta(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Add data and export
	dt.store.events.Append(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})

	fullFile := "test_empty_full.json"
	defer os.Remove(fullFile)

	checkpoint, err := dt.ExportFull(fullFile, ExportOptions{IncludeEvents: true})
	require.NoError(t, err)

	// Export incremental without adding new data
	deltaFile := "test_empty_delta.json"
	defer os.Remove(deltaFile)

	newCheckpoint, err := dt.ExportIncremental(deltaFile, checkpoint)
	require.NoError(t, err)

	// Read delta
	data, _ := os.ReadFile(deltaFile)
	var delta IncrementalExportData
	json.Unmarshal(data, &delta)

	// Should have no new data
	assert.Len(t, delta.NewEvents, 0)
	assert.Len(t, delta.NewState, 0)
	assert.Len(t, delta.NewCommands, 0)

	// Checkpoint IDs should be the same (no new data)
	assert.Equal(t, checkpoint.LastEventID, newCheckpoint.LastEventID)
}

// TestExportIncremental_NilCheckpoint tests error handling
func TestExportIncremental_NilCheckpoint(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	_, err := dt.ExportIncremental("test.json", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint is nil")
}

// TestRoundTrip_FullPlusDelta tests reconstructing state from full + deltas
func TestRoundTrip_FullPlusDelta(t *testing.T) {
	dt1 := Enable()
	defer Disable()

	// Initialize store
	dt1.store = NewDevToolsStore(1000, 1000, 1000)

	// Add initial data
	dt1.store.events.Append(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	dt1.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  0,
		NewValue:  1,
		Timestamp: time.Now(),
	})

	// Export full
	fullFile := "test_roundtrip_full.json"
	defer os.Remove(fullFile)

	checkpoint, err := dt1.ExportFull(fullFile, ExportOptions{
		IncludeEvents: true,
		IncludeState:  true,
	})
	require.NoError(t, err)

	// Add more data
	dt1.store.events.Append(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})
	dt1.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  1,
		NewValue:  2,
		Timestamp: time.Now(),
	})

	// Export delta
	deltaFile := "test_roundtrip_delta.json"
	defer os.Remove(deltaFile)

	_, err = dt1.ExportIncremental(deltaFile, checkpoint)
	require.NoError(t, err)

	// Create new DevTools instance and reconstruct
	dt2 := Enable()

	// Initialize store
	dt2.store = NewDevToolsStore(1000, 1000, 1000)

	// Import full
	err = dt2.Import(fullFile)
	require.NoError(t, err)

	// Import delta
	err = dt2.ImportDelta(deltaFile)
	require.NoError(t, err)

	// Verify reconstructed state matches original
	events := dt2.store.events.GetRecent(dt2.store.events.Len())
	assert.Len(t, events, 2)
	assert.Equal(t, "event-1", events[0].ID)
	assert.Equal(t, "event-2", events[1].ID)

	state := dt2.store.stateHistory.GetAll()
	assert.Len(t, state, 2)
	assert.Equal(t, float64(1), state[0].NewValue) // JSON unmarshals numbers as float64
	assert.Equal(t, float64(2), state[1].NewValue)
}

// TestFileSize_IncrementalSmaller tests that incremental exports are smaller
func TestFileSize_IncrementalSmaller(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Add lots of initial data
	for i := 0; i < 100; i++ {
		dt.store.events.Append(EventRecord{
			ID:        "event-" + string(rune(i)),
			Name:      "click",
			Timestamp: time.Now(),
		})
	}

	// Export full
	fullFile := "test_size_full.json"
	defer os.Remove(fullFile)

	checkpoint, err := dt.ExportFull(fullFile, ExportOptions{IncludeEvents: true})
	require.NoError(t, err)

	// Add just a few more events
	for i := 0; i < 5; i++ {
		dt.store.events.Append(EventRecord{
			ID:        "event-new-" + string(rune(i)),
			Name:      "submit",
			Timestamp: time.Now(),
		})
	}

	// Export incremental
	deltaFile := "test_size_delta.json"
	defer os.Remove(deltaFile)

	_, err = dt.ExportIncremental(deltaFile, checkpoint)
	require.NoError(t, err)

	// Compare file sizes
	fullInfo, _ := os.Stat(fullFile)
	deltaInfo, _ := os.Stat(deltaFile)

	// Delta should be significantly smaller (at least 80% smaller)
	assert.Less(t, deltaInfo.Size(), fullInfo.Size()/5)
}

// TestCheckpointIDTracking tests that IDs are tracked correctly
func TestCheckpointIDTracking(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	dt.store = NewDevToolsStore(1000, 1000, 1000)

	// Add data with known IDs
	dt.store.events.Append(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	dt.store.stateHistory.Record(StateChange{RefID: "ref-1", RefName: "counter", Timestamp: time.Now()})
	dt.store.commands.RecordCommand(CommandRecord{ID: "cmd-1", Type: "fetch", Generated: time.Now()})

	eventID1 := dt.store.events.GetMaxID()
	stateID1 := dt.store.stateHistory.GetMaxID()
	cmdID1 := dt.store.commands.GetMaxID()

	// Add more data
	dt.store.events.Append(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})
	dt.store.stateHistory.Record(StateChange{RefID: "ref-2", RefName: "text", Timestamp: time.Now()})
	dt.store.commands.RecordCommand(CommandRecord{ID: "cmd-2", Type: "update", Generated: time.Now()})

	eventID2 := dt.store.events.GetMaxID()
	stateID2 := dt.store.stateHistory.GetMaxID()
	cmdID2 := dt.store.commands.GetMaxID()

	// IDs should be increasing
	assert.Greater(t, eventID2, eventID1)
	assert.Greater(t, stateID2, stateID1)
	assert.Greater(t, cmdID2, cmdID1)
}

// TestGetSince_FiltersCorrectly tests the store's GetSince method
func TestGetSince_FiltersCorrectly(t *testing.T) {
	store := NewDevToolsStore(1000, 1000, 1000)

	// Add data
	store.events.Append(EventRecord{ID: "event-1", Name: "click", Timestamp: time.Now()})
	store.events.Append(EventRecord{ID: "event-2", Name: "submit", Timestamp: time.Now()})
	store.stateHistory.Record(StateChange{RefID: "ref-1", RefName: "counter", Timestamp: time.Now()})

	// Create checkpoint after first event
	checkpoint := &ExportCheckpoint{
		LastEventID:   1,
		LastStateID:   0,
		LastCommandID: 0,
	}

	// Get data since checkpoint
	delta, err := store.GetSince(checkpoint)
	require.NoError(t, err)

	// Should only have event-2 and the state change
	assert.Len(t, delta.NewEvents, 1)
	assert.Equal(t, "event-2", delta.NewEvents[0].ID)
	assert.Len(t, delta.NewState, 1)
}

// TestGetSince_NilCheckpoint tests error handling
func TestGetSince_NilCheckpoint(t *testing.T) {
	store := NewDevToolsStore(1000, 1000, 1000)

	_, err := store.GetSince(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint is nil")
}
