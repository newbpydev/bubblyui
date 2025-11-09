package devtools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Import loads debug data from a JSON file and restores it to the dev tools store.
//
// This function reads the specified file, validates the data, and replaces all
// existing data in the store with the imported data. Any existing components,
// state history, events, and performance metrics will be cleared.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses write lock on DevTools.
//
// Example:
//
//	dt := devtools.Enable()
//	err := dt.Import("debug-state.json")
//	if err != nil {
//	    log.Printf("Import failed: %v", err)
//	}
//
// Parameters:
//   - filename: Path to the JSON file to import
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) Import(filename string) error {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Import from bytes using a reader
	return dt.ImportFromReader(io.NopCloser(io.Reader(newBytesReader(data))))
}

// ImportFromReader loads debug data from an io.Reader and restores it to the dev tools store.
//
// This function reads from any io.Reader (file, network, memory buffer, etc.),
// validates the data, and replaces all existing data in the store with the
// imported data. This is more flexible than Import() as it works with any
// data source.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses write lock on DevTools.
//
// Example:
//
//	dt := devtools.Enable()
//	jsonData := `{"version":"1.0","timestamp":"2024-01-01T00:00:00Z"}`
//	reader := strings.NewReader(jsonData)
//	err := dt.ImportFromReader(reader)
//	if err != nil {
//	    log.Printf("Import failed: %v", err)
//	}
//
// Parameters:
//   - reader: io.Reader containing JSON export data
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) ImportFromReader(reader io.Reader) error {
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

	// Read all data from reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read import data: %w", err)
	}

	// Unmarshal JSON
	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Validate imported data
	err = dt.ValidateImport(&exportData)
	if err != nil {
		return fmt.Errorf("import validation failed: %w", err)
	}

	// Clear existing data
	dt.store.mu.Lock()
	dt.store.components = make(map[string]*ComponentSnapshot)
	dt.store.mu.Unlock()

	dt.store.stateHistory.Clear()
	dt.store.events.Clear()
	dt.store.performance.Clear()

	// Restore components
	if exportData.Components != nil {
		for _, comp := range exportData.Components {
			dt.store.AddComponent(comp)
		}
	}

	// Restore state history
	if exportData.State != nil {
		for _, state := range exportData.State {
			dt.store.stateHistory.Record(state)
		}
	}

	// Restore events
	if exportData.Events != nil {
		for _, event := range exportData.Events {
			dt.store.events.Append(event)
		}
	}

	// Restore performance data
	if exportData.Performance != nil {
		// Performance data is a pointer, so we need to restore each component
		allPerf := exportData.Performance.GetAll()
		for _, perf := range allPerf {
			// Restore performance metrics by recording renders
			// This is a simplified restoration - in reality we'd need to restore
			// the exact state, but for now we'll just record the metrics
			for i := int64(0); i < perf.RenderCount; i++ {
				// Record average render time for each count
				dt.store.performance.RecordRender(
					perf.ComponentID,
					perf.ComponentName,
					perf.AvgRenderTime,
				)
			}
		}
	}

	return nil
}

// ValidateImport validates imported data before applying it to the store.
//
// This function performs comprehensive validation to ensure the imported data
// is well-formed and compatible with the current dev tools version. It checks:
//   - Version compatibility (currently only "1.0" is supported)
//   - Timestamp is not zero
//   - Component IDs are unique and non-empty
//   - State changes have valid RefIDs and timestamps
//   - Events have valid IDs and timestamps
//
// Thread Safety:
//
//	Safe to call concurrently. This is a pure function with no side effects.
//
// Example:
//
//	dt := devtools.Enable()
//	data := &ExportData{Version: "1.0", Timestamp: time.Now()}
//	err := dt.ValidateImport(data)
//	if err != nil {
//	    log.Printf("Validation failed: %v", err)
//	}
//
// Parameters:
//   - data: The export data to validate
//
// Returns:
//   - error: nil if valid, error describing the validation failure otherwise
func (dt *DevTools) ValidateImport(data *ExportData) error {
	if data == nil {
		return fmt.Errorf("import data is nil")
	}

	// Validate version
	if data.Version != "1.0" {
		return fmt.Errorf("unsupported version: %s (only 1.0 is supported)", data.Version)
	}

	// Validate timestamp
	if data.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is zero")
	}

	// Validate components
	if data.Components != nil {
		componentIDs := make(map[string]bool)
		for i, comp := range data.Components {
			if comp == nil {
				return fmt.Errorf("component at index %d is nil", i)
			}
			if comp.ID == "" {
				return fmt.Errorf("component at index %d has empty ID", i)
			}
			if componentIDs[comp.ID] {
				return fmt.Errorf("duplicate component ID: %s", comp.ID)
			}
			componentIDs[comp.ID] = true
		}
	}

	// Validate state history
	if data.State != nil {
		for i, state := range data.State {
			if state.RefID == "" {
				return fmt.Errorf("state change at index %d has empty RefID", i)
			}
			if state.Timestamp.IsZero() {
				return fmt.Errorf("state change at index %d has zero timestamp", i)
			}
		}
	}

	// Validate events
	if data.Events != nil {
		for i, event := range data.Events {
			if event.ID == "" {
				return fmt.Errorf("event at index %d has empty ID", i)
			}
			if event.Timestamp.IsZero() {
				return fmt.Errorf("event at index %d has zero timestamp", i)
			}
		}
	}

	return nil
}

// newBytesReader creates a new bytes.Reader from a byte slice.
// This is a helper to avoid importing bytes package in the interface.
func newBytesReader(data []byte) io.Reader {
	return &bytesReader{data: data, pos: 0}
}

// bytesReader is a simple implementation of io.Reader for byte slices.
type bytesReader struct {
	data []byte
	pos  int
}

func (br *bytesReader) Read(p []byte) (n int, err error) {
	if br.pos >= len(br.data) {
		return 0, io.EOF
	}
	n = copy(p, br.data[br.pos:])
	br.pos += n
	return n, nil
}
