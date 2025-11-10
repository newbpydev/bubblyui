package devtools

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectCompression tests the detectCompression function for gzip magic bytes.
func TestDetectCompression(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantGzip   bool
		wantErr    bool
		errMessage string
	}{
		{
			name:     "gzip magic bytes",
			data:     []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantGzip: true,
			wantErr:  false,
		},
		{
			name:     "json data (not compressed)",
			data:     []byte(`{"version":"1.0"}`),
			wantGzip: false,
			wantErr:  false,
		},
		{
			name:     "empty file",
			data:     []byte{},
			wantGzip: false,
			wantErr:  false,
		},
		{
			name:     "single byte",
			data:     []byte{0x1f},
			wantGzip: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test data
			tmpFile, err := os.CreateTemp("", "test-detect-*.dat")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			_, err = tmpFile.Write(tt.data)
			require.NoError(t, err)

			// Seek back to start for reading
			_, err = tmpFile.Seek(0, 0)
			require.NoError(t, err)

			// Test detection
			isGzip, err := detectCompression(tmpFile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantGzip, isGzip)
			}
		})
	}
}

// TestExportCompressed tests exporting with gzip compression.
func TestExportCompressed(t *testing.T) {
	tests := []struct {
		name              string
		compressionLevel  int
		includeComponents bool
		includeState      bool
		wantErr           bool
		errMessage        string
	}{
		{
			name:              "default compression",
			compressionLevel:  gzip.DefaultCompression,
			includeComponents: true,
			includeState:      true,
			wantErr:           false,
		},
		{
			name:              "best speed compression",
			compressionLevel:  gzip.BestSpeed,
			includeComponents: true,
			includeState:      false,
			wantErr:           false,
		},
		{
			name:              "best compression",
			compressionLevel:  gzip.BestCompression,
			includeComponents: false,
			includeState:      true,
			wantErr:           false,
		},
		{
			name:              "no compression (level 0)",
			compressionLevel:  gzip.NoCompression,
			includeComponents: true,
			includeState:      true,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup dev tools with test data
			dt := setupTestDevTools(t)

			// Create temp file
			tmpFile := filepath.Join(t.TempDir(), "export-compressed.json.gz")

			// Export with compression
			opts := ExportOptions{
				IncludeComponents: tt.includeComponents,
				IncludeState:      tt.includeState,
				Compress:          true,
				CompressionLevel:  tt.compressionLevel,
			}

			err := dt.Export(tmpFile, opts)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
				return
			}

			require.NoError(t, err)

			// Verify file exists
			_, err = os.Stat(tmpFile)
			require.NoError(t, err)

			// Verify file has gzip magic bytes
			file, err := os.Open(tmpFile)
			require.NoError(t, err)
			defer file.Close()

			isGzip, err := detectCompression(file)
			require.NoError(t, err)
			assert.True(t, isGzip, "exported file should have gzip magic bytes")

			// Verify we can decompress and read the data
			_, err = file.Seek(0, 0)
			require.NoError(t, err)

			gzReader, err := gzip.NewReader(file)
			require.NoError(t, err)
			defer gzReader.Close()

			var exportData ExportData
			decoder := json.NewDecoder(gzReader)
			err = decoder.Decode(&exportData)
			require.NoError(t, err)

			// Verify data structure
			assert.Equal(t, "1.0", exportData.Version)
			assert.False(t, exportData.Timestamp.IsZero())

			if tt.includeComponents {
				assert.NotNil(t, exportData.Components)
			}
			if tt.includeState {
				assert.NotNil(t, exportData.State)
			}
		})
	}
}

// TestImportCompressed tests importing gzip-compressed files with auto-detection.
func TestImportCompressed(t *testing.T) {
	tests := []struct {
		name       string
		compressed bool
		wantErr    bool
		errMessage string
	}{
		{
			name:       "import compressed file",
			compressed: true,
			wantErr:    false,
		},
		{
			name:       "import uncompressed file",
			compressed: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup dev tools with test data
			dt := setupTestDevTools(t)

			// Export first
			tmpFile := filepath.Join(t.TempDir(), "export.json")
			if tt.compressed {
				tmpFile += ".gz"
			}

			exportOpts := ExportOptions{
				IncludeComponents: true,
				IncludeState:      true,
				Compress:          tt.compressed,
				CompressionLevel:  gzip.DefaultCompression,
			}

			err := dt.Export(tmpFile, exportOpts)
			require.NoError(t, err)

			// Clear store
			dt.store.mu.Lock()
			dt.store.components = make(map[string]*ComponentSnapshot)
			dt.store.mu.Unlock()
			dt.store.stateHistory.Clear()

			// Import (should auto-detect compression)
			err = dt.Import(tmpFile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
				return
			}

			require.NoError(t, err)

			// Verify data was imported
			components := dt.store.GetAllComponents()
			assert.NotEmpty(t, components)

			stateHistory := dt.store.stateHistory.GetAll()
			assert.NotEmpty(t, stateHistory)
		})
	}
}

// TestCompressionRoundTrip tests export → import round-trip with compression.
func TestCompressionRoundTrip(t *testing.T) {
	// Setup dev tools with test data
	dt := setupTestDevTools(t)

	// Get original data
	originalComponents := dt.store.GetAllComponents()
	originalState := dt.store.stateHistory.GetAll()

	require.NotEmpty(t, originalComponents, "should have test components")
	require.NotEmpty(t, originalState, "should have test state")

	// Export with compression
	tmpFile := filepath.Join(t.TempDir(), "roundtrip.json.gz")
	exportOpts := ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
		Compress:          true,
		CompressionLevel:  gzip.DefaultCompression,
	}

	err := dt.Export(tmpFile, exportOpts)
	require.NoError(t, err)

	// Clear store
	dt.store.mu.Lock()
	dt.store.components = make(map[string]*ComponentSnapshot)
	dt.store.mu.Unlock()
	dt.store.stateHistory.Clear()

	// Verify cleared
	assert.Empty(t, dt.store.GetAllComponents())
	assert.Empty(t, dt.store.stateHistory.GetAll())

	// Import
	err = dt.Import(tmpFile)
	require.NoError(t, err)

	// Verify data matches
	importedComponents := dt.store.GetAllComponents()
	importedState := dt.store.stateHistory.GetAll()

	assert.Len(t, importedComponents, len(originalComponents))
	assert.Len(t, importedState, len(originalState))

	// Verify component IDs match
	for _, orig := range originalComponents {
		found := false
		for _, imp := range importedComponents {
			if imp.ID == orig.ID {
				found = true
				assert.Equal(t, orig.Name, imp.Name)
				break
			}
		}
		assert.True(t, found, "component %s should be imported", orig.ID)
	}
}

// TestCompressionSizeReduction tests that compression reduces file size.
func TestCompressionSizeReduction(t *testing.T) {
	// Setup dev tools with substantial test data
	dt := setupTestDevTools(t)

	// Add more test data to make compression meaningful
	for i := 0; i < 50; i++ {
		comp := &ComponentSnapshot{
			ID:        generateID(),
			Name:      "TestComponent",
			Type:      "bubbly.Component",
			Timestamp: time.Now(),
			Props: map[string]interface{}{
				"prop1": "value1",
				"prop2": "value2",
				"prop3": "value3",
			},
			State: map[string]interface{}{
				"state1": "data1",
				"state2": "data2",
			},
		}
		dt.store.AddComponent(comp)
	}

	tmpDir := t.TempDir()

	// Export uncompressed
	uncompressedFile := filepath.Join(tmpDir, "uncompressed.json")
	err := dt.Export(uncompressedFile, ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
		Compress:          false,
	})
	require.NoError(t, err)

	// Export compressed
	compressedFile := filepath.Join(tmpDir, "compressed.json.gz")
	err = dt.Export(compressedFile, ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
		Compress:          true,
		CompressionLevel:  gzip.DefaultCompression,
	})
	require.NoError(t, err)

	// Get file sizes
	uncompressedInfo, err := os.Stat(uncompressedFile)
	require.NoError(t, err)
	compressedInfo, err := os.Stat(compressedFile)
	require.NoError(t, err)

	uncompressedSize := uncompressedInfo.Size()
	compressedSize := compressedInfo.Size()

	// Verify compression reduced size
	assert.Less(t, compressedSize, uncompressedSize, "compressed file should be smaller")

	// Calculate reduction percentage
	reduction := float64(uncompressedSize-compressedSize) / float64(uncompressedSize) * 100

	// Should achieve at least 30% reduction (conservative target)
	// Spec says 50-70% typical, but we'll be conservative
	assert.GreaterOrEqual(t, reduction, 30.0, "should achieve at least 30%% compression")

	t.Logf("Compression stats: uncompressed=%d bytes, compressed=%d bytes, reduction=%.1f%%",
		uncompressedSize, compressedSize, reduction)
}

// TestCompressionLevels tests different compression levels.
func TestCompressionLevels(t *testing.T) {
	dt := setupTestDevTools(t)

	// Add substantial test data
	for i := 0; i < 50; i++ {
		comp := &ComponentSnapshot{
			ID:        generateID(),
			Name:      "TestComponent",
			Type:      "bubbly.Component",
			Timestamp: time.Now(),
			Props: map[string]interface{}{
				"prop1": "value1",
				"prop2": "value2",
			},
		}
		dt.store.AddComponent(comp)
	}

	levels := []struct {
		name  string
		level int
	}{
		{"best_speed", gzip.BestSpeed},
		{"default", gzip.DefaultCompression},
		{"best_compression", gzip.BestCompression},
	}

	tmpDir := t.TempDir()
	sizes := make(map[string]int64)

	for _, lvl := range levels {
		t.Run(lvl.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, lvl.name+".json.gz")

			err := dt.Export(filename, ExportOptions{
				IncludeComponents: true,
				IncludeState:      true,
				Compress:          true,
				CompressionLevel:  lvl.level,
			})
			require.NoError(t, err)

			info, err := os.Stat(filename)
			require.NoError(t, err)

			sizes[lvl.name] = info.Size()
			t.Logf("%s: %d bytes", lvl.name, info.Size())
		})
	}

	// Verify BestCompression <= Default <= BestSpeed
	// (smaller or equal size with better compression)
	assert.LessOrEqual(t, sizes["best_compression"], sizes["default"],
		"best compression should produce smaller or equal file")
	assert.LessOrEqual(t, sizes["default"], sizes["best_speed"],
		"default should produce smaller or equal file than best speed")
}

// TestCorruptGzipHandling tests error handling for corrupt gzip files.
func TestCorruptGzipHandling(t *testing.T) {
	dt := Enable()
	defer Disable()

	// Initialize store
	if dt.store == nil {
		dt.store = NewDevToolsStore(1000, 1000)
	}

	// Create file with gzip magic bytes but corrupt data
	tmpFile := filepath.Join(t.TempDir(), "corrupt.json.gz")
	corruptData := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff}
	err := os.WriteFile(tmpFile, corruptData, 0644)
	require.NoError(t, err)

	// Attempt import - should fail gracefully
	err = dt.Import(tmpFile)
	assert.Error(t, err, "should error on corrupt gzip")
	// Error could be from gzip reader or JSON unmarshaling
	// Just verify it fails gracefully without panic
}

// setupTestDevTools creates a DevTools instance with test data.
func setupTestDevTools(t *testing.T) *DevTools {
	t.Helper()

	dt := Enable()
	t.Cleanup(func() { Disable() })

	// Initialize store if not already initialized
	if dt.store == nil {
		dt.store = NewDevToolsStore(1000, 1000)
	}

	// Add test components
	comp1 := &ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent1",
		Type:      "bubbly.Component",
		Timestamp: time.Now(),
		Props: map[string]interface{}{
			"prop1": "value1",
		},
		State: map[string]interface{}{
			"state1": "data1",
		},
		Refs: []*RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "count",
				Type:  "int",
				Value: 42,
			},
		},
	}

	comp2 := &ComponentSnapshot{
		ID:        "comp-2",
		Name:      "TestComponent2",
		Type:      "bubbly.Component",
		Timestamp: time.Now(),
		Props: map[string]interface{}{
			"prop2": "value2",
		},
	}

	dt.store.AddComponent(comp1)
	dt.store.AddComponent(comp2)

	// Add test state history
	dt.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "count",
		OldValue:  0,
		NewValue:  42,
		Timestamp: time.Now(),
		Source:    "test",
	})

	return dt
}

// generateID generates a unique ID for testing.
func generateID() string {
	return "test-" + time.Now().Format("20060102150405.000000")
}
