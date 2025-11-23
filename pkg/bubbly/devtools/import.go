package devtools

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// readFileWithCompression reads a file, handling gzip compression if detected.
func readFileWithCompression(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	isCompressed, err := detectCompression(file)
	if err != nil {
		return nil, fmt.Errorf("failed to detect compression: %w", err)
	}

	if _, err = file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	var reader io.Reader = file
	if isCompressed {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	return io.ReadAll(reader)
}

// restoreExportData restores export data to the dev tools store.
func (dt *DevTools) restoreExportData(exportData *ExportData) {
	dt.store.mu.Lock()
	dt.store.components = make(map[string]*ComponentSnapshot)
	dt.store.mu.Unlock()

	dt.store.stateHistory.Clear()
	dt.store.events.Clear()
	dt.store.performance.Clear()

	if exportData.Components != nil {
		for _, comp := range exportData.Components {
			dt.store.AddComponent(comp)
		}
	}

	if exportData.State != nil {
		for _, state := range exportData.State {
			dt.store.stateHistory.Record(state)
		}
	}

	if exportData.Events != nil {
		for _, event := range exportData.Events {
			dt.store.events.Append(event)
		}
	}

	if exportData.Performance != nil {
		dt.restorePerformanceData(exportData.Performance)
	}
}

// restorePerformanceData restores performance metrics from export data.
func (dt *DevTools) restorePerformanceData(perf *PerformanceData) {
	allPerf := perf.GetAll()
	for _, p := range allPerf {
		for i := int64(0); i < p.RenderCount; i++ {
			dt.store.performance.RecordRender(p.ComponentID, p.ComponentName, p.AvgRenderTime)
		}
	}
}

// Import loads debug data from a JSON file and restores it to the dev tools store.
//
// This function reads the specified file, validates the data, and replaces all
// existing data in the store with the imported data. Any existing components,
// state history, events, and performance metrics will be cleared.
//
// The function automatically detects gzip compression by checking for gzip magic
// bytes (0x1f 0x8b) at the start of the file. If detected, the file is
// automatically decompressed before parsing.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses write lock on DevTools.
//
// Example:
//
//	dt := devtools.Enable()
//	err := dt.Import("debug-state.json")  // Works with .json or .json.gz
//	if err != nil {
//	    log.Printf("Import failed: %v", err)
//	}
//
// Parameters:
//   - filename: Path to the JSON file to import (compressed or uncompressed)
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) Import(filename string) error {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	// Detect if file is gzip-compressed
	isCompressed, err := detectCompression(file)
	if err != nil {
		return fmt.Errorf("failed to detect compression: %w", err)
	}

	// Seek back to start after detection
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Create appropriate reader
	var reader io.Reader = file
	if isCompressed {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Import from reader
	return dt.ImportFromReader(io.NopCloser(reader))
}

// ImportFormat loads debug data from a file using the specified format.
//
// This method supports multiple import formats (JSON, YAML, MessagePack) and
// automatically detects and handles gzip compression.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses write lock on DevTools.
//
// Example:
//
//	// Import from YAML
//	err := devtools.ImportFormat("debug.yaml", "yaml")
//
//	// Import from compressed MessagePack
//	err := devtools.ImportFormat("debug.msgpack.gz", "msgpack")
//
//	// Auto-detect format from filename
//	format, _ := DetectFormat("debug.yaml")
//	err := devtools.ImportFormat("debug.yaml", format)
//
// Parameters:
//   - filename: Path to the file to import
//   - formatName: Format name ("json", "yaml", "msgpack")
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) ImportFormat(filename, formatName string) error {
	data, err := readFileWithCompression(filename)
	if err != nil {
		return err
	}

	registry := getGlobalRegistry()
	format, err := registry.Get(formatName)
	if err != nil {
		return fmt.Errorf("failed to get format: %w", err)
	}

	var exportData ExportData
	if err = format.Unmarshal(data, &exportData); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	if err = dt.ValidateImport(&exportData); err != nil {
		return fmt.Errorf("import validation failed: %w", err)
	}

	dt.mu.Lock()
	defer dt.mu.Unlock()

	if !dt.enabled {
		return fmt.Errorf("dev tools not enabled")
	}
	if dt.store == nil {
		return fmt.Errorf("dev tools store not initialized")
	}

	dt.restoreExportData(&exportData)
	return nil
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
//
// parseAndMigrateData parses JSON data, checks version, and applies migrations.
func parseAndMigrateData(data []byte) ([]byte, error) {
	var genericData map[string]interface{}
	if err := json.Unmarshal(data, &genericData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	version, err := extractVersion(genericData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract version: %w", err)
	}

	const currentVersion = "1.0"
	if version != currentVersion {
		genericData, err = migrateVersion(genericData, version, currentVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to migrate from version %s to %s: %w", version, currentVersion, err)
		}
		data, err = json.Marshal(genericData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal migrated data: %w", err)
		}
	}

	return data, nil
}

func (dt *DevTools) ImportFromReader(reader io.Reader) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	if !dt.enabled {
		return fmt.Errorf("dev tools not enabled")
	}
	if dt.store == nil {
		return fmt.Errorf("dev tools store not initialized")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read import data: %w", err)
	}

	data, err = parseAndMigrateData(data)
	if err != nil {
		return err
	}

	var exportData ExportData
	if err = json.Unmarshal(data, &exportData); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	if err = dt.ValidateImport(&exportData); err != nil {
		return fmt.Errorf("import validation failed: %w", err)
	}

	dt.restoreExportData(&exportData)
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
//
// validateImportComponents validates the components slice in import data.
func validateImportComponents(components []*ComponentSnapshot) error {
	componentIDs := make(map[string]bool)
	for i, comp := range components {
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
	return nil
}

// validateImportStateHistory validates the state history slice in import data.
func validateImportStateHistory(state []StateChange) error {
	for i, s := range state {
		if s.RefID == "" {
			return fmt.Errorf("state change at index %d has empty RefID", i)
		}
		if s.Timestamp.IsZero() {
			return fmt.Errorf("state change at index %d has zero timestamp", i)
		}
	}
	return nil
}

// validateImportEvents validates the events slice in import data.
func validateImportEvents(events []EventRecord) error {
	for i, event := range events {
		if event.ID == "" {
			return fmt.Errorf("event at index %d has empty ID", i)
		}
		if event.Timestamp.IsZero() {
			return fmt.Errorf("event at index %d has zero timestamp", i)
		}
	}
	return nil
}

func (dt *DevTools) ValidateImport(data *ExportData) error {
	if data == nil {
		return fmt.Errorf("import data is nil")
	}
	if data.Version == "" {
		return fmt.Errorf("version field is empty")
	}
	if data.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is zero")
	}

	if data.Components != nil {
		if err := validateImportComponents(data.Components); err != nil {
			return err
		}
	}
	if data.State != nil {
		if err := validateImportStateHistory(data.State); err != nil {
			return err
		}
	}
	if data.Events != nil {
		if err := validateImportEvents(data.Events); err != nil {
			return err
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

// detectCompression checks if a file is gzip-compressed by reading the magic bytes.
//
// Gzip files start with magic bytes 0x1f 0x8b. This function reads the first
// two bytes of the file to detect compression. The file position is NOT reset
// after detection - caller must seek back to the start if needed.
//
// Thread Safety:
//
//	Safe to call concurrently on different files. Not safe on the same file.
//
// Parameters:
//   - file: The file to check for gzip compression
//
// Returns:
//   - bool: true if file is gzip-compressed, false otherwise
//   - error: nil on success, error if read fails
func detectCompression(file *os.File) (bool, error) {
	// Read first 2 bytes for gzip magic bytes check
	magicBytes := make([]byte, 2)
	n, err := file.Read(magicBytes)

	// If we can't read 2 bytes, file is too small to be gzip
	if err == io.EOF || n < 2 {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to read magic bytes: %w", err)
	}

	// Check for gzip magic bytes: 0x1f 0x8b
	isGzip := magicBytes[0] == 0x1f && magicBytes[1] == 0x8b

	return isGzip, nil
}
