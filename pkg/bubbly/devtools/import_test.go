package devtools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport_LoadsFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-import.json")

	// Create dev tools and export some data
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	// Add test data
	dt.store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
	})
	dt.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  41,
		NewValue:  42,
		Timestamp: time.Now(),
	})

	// Export
	err := dt.Export(filename, ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
	})
	require.NoError(t, err)

	// Clear store
	dt.store.mu.Lock()
	dt.store.components = make(map[string]*ComponentSnapshot)
	dt.store.mu.Unlock()
	dt.store.stateHistory.Clear()

	// Verify store is empty
	assert.Empty(t, dt.store.GetAllComponents())
	assert.Empty(t, dt.store.stateHistory.GetAll())

	// Import
	err = dt.Import(filename)
	require.NoError(t, err)

	// Verify data restored
	components := dt.store.GetAllComponents()
	assert.Len(t, components, 1)
	assert.Equal(t, "comp-1", components[0].ID)

	stateHistory := dt.store.stateHistory.GetAll()
	assert.Len(t, stateHistory, 1)
	assert.Equal(t, "ref-1", stateHistory[0].RefID)
}

func TestImportFromReader_Success(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	// Create JSON data
	jsonData := `{
		"version": "1.0",
		"timestamp": "2024-01-01T12:00:00Z",
		"components": [
			{
				"id": "comp-1",
				"name": "TestComponent",
				"type": "test",
				"timestamp": "2024-01-01T12:00:00Z"
			}
		],
		"state": [
			{
				"refID": "ref-1",
				"refName": "counter",
				"oldValue": 41,
				"newValue": 42,
				"timestamp": "2024-01-01T12:00:00Z",
				"source": "test"
			}
		]
	}`

	// Import from reader
	reader := strings.NewReader(jsonData)
	err := dt.ImportFromReader(reader)
	require.NoError(t, err)

	// Verify data restored
	components := dt.store.GetAllComponents()
	assert.Len(t, components, 1)
	assert.Equal(t, "comp-1", components[0].ID)

	stateHistory := dt.store.stateHistory.GetAll()
	assert.Len(t, stateHistory, 1)
	assert.Equal(t, "ref-1", stateHistory[0].RefID)
}

func TestValidateImport_ValidData(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{ID: "comp-1", Name: "Test"},
		},
		State: []StateChange{
			{RefID: "ref-1", RefName: "test", Timestamp: time.Now()},
		},
		Events: []EventRecord{
			{ID: "event-1", Name: "test", Timestamp: time.Now()},
		},
	}

	err := dt.ValidateImport(data)
	assert.NoError(t, err, "Valid data should pass validation")
}

func TestValidateImport_InvalidVersion(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	tests := []struct {
		name    string
		version string
	}{
		{"empty version", ""},
		{"unsupported version", "2.0"},
		{"invalid version", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ExportData{
				Version:   tt.version,
				Timestamp: time.Now(),
			}

			err := dt.ValidateImport(data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "version")
		})
	}
}

func TestValidateImport_ZeroTimestamp(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Time{}, // Zero value
	}

	err := dt.ValidateImport(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timestamp")
}

func TestValidateImport_NilData(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	err := dt.ValidateImport(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestValidateImport_ComponentValidation(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	tests := []struct {
		name       string
		components []*ComponentSnapshot
		wantErr    string
	}{
		{
			name:       "nil component",
			components: []*ComponentSnapshot{nil},
			wantErr:    "nil",
		},
		{
			name:       "empty component ID",
			components: []*ComponentSnapshot{{ID: "", Name: "Test"}},
			wantErr:    "empty ID",
		},
		{
			name: "duplicate component IDs",
			components: []*ComponentSnapshot{
				{ID: "comp-1", Name: "Test1"},
				{ID: "comp-1", Name: "Test2"},
			},
			wantErr: "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ExportData{
				Version:    "1.0",
				Timestamp:  time.Now(),
				Components: tt.components,
			}

			err := dt.ValidateImport(data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidateImport_StateValidation(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	tests := []struct {
		name    string
		state   []StateChange
		wantErr string
	}{
		{
			name: "empty RefID",
			state: []StateChange{
				{RefID: "", RefName: "test", Timestamp: time.Now()},
			},
			wantErr: "empty RefID",
		},
		{
			name: "zero timestamp",
			state: []StateChange{
				{RefID: "ref-1", RefName: "test", Timestamp: time.Time{}},
			},
			wantErr: "zero timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				State:     tt.state,
			}

			err := dt.ValidateImport(data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidateImport_EventValidation(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	tests := []struct {
		name    string
		events  []EventRecord
		wantErr string
	}{
		{
			name: "empty event ID",
			events: []EventRecord{
				{ID: "", Name: "test", Timestamp: time.Now()},
			},
			wantErr: "empty ID",
		},
		{
			name: "zero timestamp",
			events: []EventRecord{
				{ID: "event-1", Name: "test", Timestamp: time.Time{}},
			},
			wantErr: "zero timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Events:    tt.events,
			}

			err := dt.ValidateImport(data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestImport_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(filename, []byte("not valid json {{{"), 0644)
	require.NoError(t, err)

	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	err = dt.Import(filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestImport_FileNotFound(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	err := dt.Import("/nonexistent/path/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open import file")
}

func TestImport_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "roundtrip.json")

	// Create dev tools with test data
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	// Add comprehensive test data
	originalComp := &ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent",
		Type:      "test",
		Timestamp: time.Now().Truncate(time.Second), // Truncate for JSON comparison
		Props: map[string]interface{}{
			"prop1": "value1",
			"prop2": 42,
		},
	}
	dt.store.AddComponent(originalComp)

	originalState := StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  41,
		NewValue:  42,
		Timestamp: time.Now().Truncate(time.Second),
		Source:    "test",
	}
	dt.store.stateHistory.Record(originalState)

	originalEvent := EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "comp-1",
		Timestamp: time.Now().Truncate(time.Second),
		Payload:   "test payload",
	}
	dt.store.events.Append(originalEvent)

	// Export
	err := dt.Export(filename, ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
		IncludeEvents:     true,
	})
	require.NoError(t, err)

	// Clear store
	dt.store.mu.Lock()
	dt.store.components = make(map[string]*ComponentSnapshot)
	dt.store.mu.Unlock()
	dt.store.stateHistory.Clear()
	dt.store.events.Clear()

	// Import
	err = dt.Import(filename)
	require.NoError(t, err)

	// Verify round-trip data matches
	components := dt.store.GetAllComponents()
	require.Len(t, components, 1)
	assert.Equal(t, originalComp.ID, components[0].ID)
	assert.Equal(t, originalComp.Name, components[0].Name)
	assert.Equal(t, originalComp.Type, components[0].Type)

	stateHistory := dt.store.stateHistory.GetAll()
	require.Len(t, stateHistory, 1)
	assert.Equal(t, originalState.RefID, stateHistory[0].RefID)
	assert.Equal(t, originalState.RefName, stateHistory[0].RefName)

	events := dt.store.events.GetRecent(10)
	require.Len(t, events, 1)
	assert.Equal(t, originalEvent.ID, events[0].ID)
	assert.Equal(t, originalEvent.Name, events[0].Name)
}

func TestImport_NotEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Create valid export file
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}
	err := dt.Export(filename, ExportOptions{})
	require.NoError(t, err)

	// Disable dev tools
	dt.enabled = false

	// Try to import
	err = dt.Import(filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestImport_NoStore(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Create valid export file first
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}
	err := dt.Export(filename, ExportOptions{})
	require.NoError(t, err)

	// Remove store
	dt.store = nil

	// Try to import
	err = dt.Import(filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestImport_ClearsExistingData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	// Add existing data
	dt.store.AddComponent(&ComponentSnapshot{ID: "old-comp", Name: "Old"})
	dt.store.stateHistory.Record(StateChange{RefID: "old-ref", RefName: "old", Timestamp: time.Now()})
	dt.store.events.Append(EventRecord{ID: "old-event", Name: "old", Timestamp: time.Now()})

	// Create export with different data
	dt2 := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}
	dt2.store.AddComponent(&ComponentSnapshot{ID: "new-comp", Name: "New"})
	err := dt2.Export(filename, ExportOptions{IncludeComponents: true})
	require.NoError(t, err)

	// Import into dt (should clear old data)
	err = dt.Import(filename)
	require.NoError(t, err)

	// Verify old data is gone
	components := dt.store.GetAllComponents()
	assert.Len(t, components, 1)
	assert.Equal(t, "new-comp", components[0].ID)

	// Old state and events should be cleared
	assert.Empty(t, dt.store.stateHistory.GetAll())
	assert.Empty(t, dt.store.events.GetRecent(10))
}

func TestImportFromReader_EmptyData(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	// Import minimal valid data
	jsonData := `{
		"version": "1.0",
		"timestamp": "2024-01-01T12:00:00Z"
	}`

	reader := strings.NewReader(jsonData)
	err := dt.ImportFromReader(reader)
	require.NoError(t, err)

	// Verify no data added (but no error)
	assert.Empty(t, dt.store.GetAllComponents())
	assert.Empty(t, dt.store.stateHistory.GetAll())
	assert.Empty(t, dt.store.events.GetRecent(10))
}

func TestImport_ValidationFailureDoesNotModifyStore(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "invalid.json")

	// Create invalid export (wrong version)
	invalidData := `{
		"version": "2.0",
		"timestamp": "2024-01-01T12:00:00Z"
	}`
	err := os.WriteFile(filename, []byte(invalidData), 0644)
	require.NoError(t, err)

	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100),
	}

	// Add existing data
	dt.store.AddComponent(&ComponentSnapshot{ID: "existing", Name: "Existing"})

	// Try to import (should fail validation)
	err = dt.Import(filename)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation")

	// Verify existing data is still there
	components := dt.store.GetAllComponents()
	assert.Len(t, components, 1)
	assert.Equal(t, "existing", components[0].ID)
}

func TestBytesReader(t *testing.T) {
	data := []byte("hello world")
	reader := newBytesReader(data)

	// Read all data
	buf := make([]byte, 20)
	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, "hello world", string(buf[:n]))

	// Read again (should get EOF)
	n, err = reader.Read(buf)
	assert.Equal(t, 0, n)
	assert.Error(t, err) // Should be io.EOF or similar
}
