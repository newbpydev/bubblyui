package devtools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport_CreatesFile(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-export.json")

	// Create dev tools with store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add some test data
	dt.store.AddComponent(&ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent",
		Type:      "test",
		Timestamp: time.Now(),
	})

	// Export with all options
	opts := ExportOptions{
		IncludeComponents:  true,
		IncludeState:       true,
		IncludeEvents:      true,
		IncludePerformance: true,
	}

	err := dt.Export(filename, opts)
	require.NoError(t, err, "Export should succeed")

	// Verify file exists
	_, err = os.Stat(filename)
	require.NoError(t, err, "Export file should exist")

	// Verify file contains valid JSON
	data, err := os.ReadFile(filename)
	require.NoError(t, err, "Should read export file")

	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	require.NoError(t, err, "Export file should contain valid JSON")

	// Verify basic structure
	assert.Equal(t, "1.0", exportData.Version, "Version should be 1.0")
	assert.NotZero(t, exportData.Timestamp, "Timestamp should be set")
	assert.Len(t, exportData.Components, 1, "Should have 1 component")
	assert.Equal(t, "comp-1", exportData.Components[0].ID, "Component ID should match")
}

func TestExport_AllDataIncluded(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-all-data.json")

	// Create dev tools with store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add component
	dt.store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
	})

	// Add state change
	dt.store.stateHistory.Record(StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  41,
		NewValue:  42,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Add event
	dt.store.events.Append(EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "comp-1",
		Timestamp: time.Now(),
	})

	// Export with all options
	opts := ExportOptions{
		IncludeComponents:  true,
		IncludeState:       true,
		IncludeEvents:      true,
		IncludePerformance: true,
	}

	err := dt.Export(filename, opts)
	require.NoError(t, err)

	// Read and verify
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	require.NoError(t, err)

	// Verify all data present
	assert.Len(t, exportData.Components, 1, "Should have 1 component")
	assert.Len(t, exportData.State, 1, "Should have 1 state change")
	assert.Len(t, exportData.Events, 1, "Should have 1 event")
	assert.NotNil(t, exportData.Performance, "Performance data should be present")
}

func TestExport_SelectiveExport(t *testing.T) {
	tests := []struct {
		name    string
		opts    ExportOptions
		checkFn func(*testing.T, ExportData)
	}{
		{
			name: "components only",
			opts: ExportOptions{
				IncludeComponents: true,
			},
			checkFn: func(t *testing.T, data ExportData) {
				assert.Len(t, data.Components, 1, "Should have components")
				assert.Empty(t, data.State, "Should not have state")
				assert.Empty(t, data.Events, "Should not have events")
				assert.Nil(t, data.Performance, "Should not have performance")
			},
		},
		{
			name: "state only",
			opts: ExportOptions{
				IncludeState: true,
			},
			checkFn: func(t *testing.T, data ExportData) {
				assert.Empty(t, data.Components, "Should not have components")
				assert.Len(t, data.State, 1, "Should have state")
				assert.Empty(t, data.Events, "Should not have events")
				assert.Nil(t, data.Performance, "Should not have performance")
			},
		},
		{
			name: "events only",
			opts: ExportOptions{
				IncludeEvents: true,
			},
			checkFn: func(t *testing.T, data ExportData) {
				assert.Empty(t, data.Components, "Should not have components")
				assert.Empty(t, data.State, "Should not have state")
				assert.Len(t, data.Events, 1, "Should have events")
				assert.Nil(t, data.Performance, "Should not have performance")
			},
		},
		{
			name: "performance only",
			opts: ExportOptions{
				IncludePerformance: true,
			},
			checkFn: func(t *testing.T, data ExportData) {
				assert.Empty(t, data.Components, "Should not have components")
				assert.Empty(t, data.State, "Should not have state")
				assert.Empty(t, data.Events, "Should not have events")
				assert.NotNil(t, data.Performance, "Should have performance")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "test-selective.json")

			// Create dev tools with test data
			dt := &DevTools{
				enabled: true,
				store:   NewDevToolsStore(100, 100, 1000),
			}

			dt.store.AddComponent(&ComponentSnapshot{ID: "comp-1", Name: "Test"})
			dt.store.stateHistory.Record(StateChange{RefID: "ref-1", RefName: "test"})
			dt.store.events.Append(EventRecord{ID: "event-1", Name: "test"})

			// Export
			err := dt.Export(filename, tt.opts)
			require.NoError(t, err)

			// Read and verify
			data, err := os.ReadFile(filename)
			require.NoError(t, err)

			var exportData ExportData
			err = json.Unmarshal(data, &exportData)
			require.NoError(t, err)

			tt.checkFn(t, exportData)
		})
	}
}

func TestExport_Sanitization(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-sanitize.json")

	// Create dev tools with sensitive data
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add component with sensitive props
	dt.store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "LoginForm",
		Props: map[string]interface{}{
			"username": "alice",
			"password": "secret123",
			"apiKey":   "key-abc-123",
		},
		Refs: []*RefSnapshot{
			{
				ID:    "ref-1",
				Name:  "userToken",
				Value: "token-xyz-789",
			},
		},
	})

	// Add state change with sensitive data
	dt.store.stateHistory.Record(StateChange{
		RefID:    "ref-1",
		RefName:  "password",
		OldValue: "oldpass",
		NewValue: "newpass",
	})

	// Export with sanitization
	opts := ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
		Sanitize:          true,
		RedactPatterns:    []string{"password", "token", "apikey"},
	}

	err := dt.Export(filename, opts)
	require.NoError(t, err)

	// Read and verify
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	require.NoError(t, err)

	// Verify sensitive data redacted
	comp := exportData.Components[0]
	assert.Equal(t, "[REDACTED]", comp.Props["password"], "Password should be redacted")
	assert.Equal(t, "[REDACTED]", comp.Props["apiKey"], "API key should be redacted")
	assert.Equal(t, "alice", comp.Props["username"], "Username should not be redacted")
	assert.Equal(t, "[REDACTED]", comp.Refs[0].Value, "Token should be redacted")

	// Verify state history redacted
	state := exportData.State[0]
	assert.Equal(t, "[REDACTED]", state.OldValue, "Old password should be redacted")
	assert.Equal(t, "[REDACTED]", state.NewValue, "New password should be redacted")
}

func TestExport_LargeExport(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-large.json")

	// Create dev tools with large dataset
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(10000, 10000, 1000),
	}

	// Add 1000 components
	for i := 0; i < 1000; i++ {
		dt.store.AddComponent(&ComponentSnapshot{
			ID:   "comp-" + string(rune(i)),
			Name: "Component" + string(rune(i)),
		})
	}

	// Add 1000 state changes
	for i := 0; i < 1000; i++ {
		dt.store.stateHistory.Record(StateChange{
			RefID:   "ref-" + string(rune(i)),
			RefName: "state" + string(rune(i)),
		})
	}

	// Export
	start := time.Now()
	opts := ExportOptions{
		IncludeComponents: true,
		IncludeState:      true,
	}

	err := dt.Export(filename, opts)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 5*time.Second, "Large export should complete in < 5 seconds")

	// Verify file size is reasonable
	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(1000), "Export file should have content")
}

func TestExport_InvalidPath(t *testing.T) {
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Try to export to invalid path
	err := dt.Export("/invalid/path/that/does/not/exist/file.json", ExportOptions{})
	assert.Error(t, err, "Should fail with invalid path")
	assert.Contains(t, err.Error(), "failed to create export file", "Error should mention file creation failure")
}

func TestExport_EmptyStore(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-empty.json")

	// Create dev tools with empty store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Export with all options (but no data)
	opts := ExportOptions{
		IncludeComponents:  true,
		IncludeState:       true,
		IncludeEvents:      true,
		IncludePerformance: true,
	}

	err := dt.Export(filename, opts)
	require.NoError(t, err, "Should succeed even with empty store")

	// Read and verify
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	require.NoError(t, err)

	// Verify structure but no data
	assert.Equal(t, "1.0", exportData.Version)
	assert.Empty(t, exportData.Components, "Should have no components")
	assert.Empty(t, exportData.State, "Should have no state")
	assert.Empty(t, exportData.Events, "Should have no events")
}

func TestExport_NotEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Create disabled dev tools
	dt := &DevTools{
		enabled: false,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	err := dt.Export(filename, ExportOptions{})
	assert.Error(t, err, "Should fail when not enabled")
	assert.Contains(t, err.Error(), "not enabled", "Error should mention not enabled")
}

func TestExport_NoStore(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	// Create dev tools without store
	dt := &DevTools{
		enabled: true,
		store:   nil,
	}

	err := dt.Export(filename, ExportOptions{})
	assert.Error(t, err, "Should fail when store is nil")
	assert.Contains(t, err.Error(), "not initialized", "Error should mention initialization")
}

func TestSanitizeExportData_NoPatterns(t *testing.T) {
	data := ExportData{
		Components: []*ComponentSnapshot{
			{
				Props: map[string]interface{}{
					"password": "secret",
				},
			},
		},
	}

	// Sanitize with no patterns
	result := sanitizeExportData(data, []string{})

	// Should return unchanged
	assert.Equal(t, "secret", result.Components[0].Props["password"])
}

func TestSanitizeExportData_CaseInsensitive(t *testing.T) {
	data := ExportData{
		Components: []*ComponentSnapshot{
			{
				Props: map[string]interface{}{
					"PASSWORD": "secret",
					"Password": "secret2",
					"password": "secret3",
				},
			},
		},
	}

	// Sanitize with lowercase pattern
	result := sanitizeExportData(data, []string{"password"})

	// All variations should be redacted
	assert.Equal(t, "[REDACTED]", result.Components[0].Props["PASSWORD"])
	assert.Equal(t, "[REDACTED]", result.Components[0].Props["Password"])
	assert.Equal(t, "[REDACTED]", result.Components[0].Props["password"])
}

func TestSanitizeExportData_ValueMatching(t *testing.T) {
	data := ExportData{
		Components: []*ComponentSnapshot{
			{
				Props: map[string]interface{}{
					"config": "password=secret123",
					"safe":   "normal value",
				},
			},
		},
	}

	// Sanitize checking values
	result := sanitizeExportData(data, []string{"password"})

	// Value containing "password" should be redacted
	assert.Equal(t, "[REDACTED]", result.Components[0].Props["config"])
	assert.Equal(t, "normal value", result.Components[0].Props["safe"])
}

func TestExport_JSONFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-format.json")

	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	dt.store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "Test",
	})

	err := dt.Export(filename, ExportOptions{IncludeComponents: true})
	require.NoError(t, err)

	// Read file content
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	content := string(data)

	// Verify indentation (should have spaces for pretty printing)
	assert.Contains(t, content, "  ", "Should be indented")
	assert.Contains(t, content, "\n", "Should have newlines")

	// Verify structure
	assert.Contains(t, content, `"version"`, "Should have version field")
	assert.Contains(t, content, `"timestamp"`, "Should have timestamp field")
	assert.Contains(t, content, `"components"`, "Should have components field")
}

func TestExport_OmitEmptyFields(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-omit.json")

	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Export with no data included
	err := dt.Export(filename, ExportOptions{})
	require.NoError(t, err)

	// Read file content
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	content := string(data)

	// Verify omitempty works - empty fields should not be in JSON
	assert.NotContains(t, content, `"components"`, "Empty components should be omitted")
	assert.NotContains(t, content, `"state"`, "Empty state should be omitted")
	assert.NotContains(t, content, `"events"`, "Empty events should be omitted")
	assert.NotContains(t, content, `"performance"`, "Nil performance should be omitted")

	// But version and timestamp should always be present
	assert.Contains(t, content, `"version"`, "Version should always be present")
	assert.Contains(t, content, `"timestamp"`, "Timestamp should always be present")
}

func TestShouldRedact(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		patterns []string
		want     bool
	}{
		{
			name:     "exact match",
			key:      "password",
			patterns: []string{"password"},
			want:     true,
		},
		{
			name:     "substring match",
			key:      "userPassword",
			patterns: []string{"password"},
			want:     true,
		},
		{
			name:     "case insensitive",
			key:      "PASSWORD",
			patterns: []string{"password"},
			want:     true,
		},
		{
			name:     "no match",
			key:      "username",
			patterns: []string{"password"},
			want:     false,
		},
		{
			name:     "multiple patterns",
			key:      "apiKey",
			patterns: []string{"password", "token", "apikey"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert patterns to lowercase for testing
			lowerPatterns := make([]string, len(tt.patterns))
			for i, p := range tt.patterns {
				lowerPatterns[i] = strings.ToLower(p)
			}

			got := shouldRedact(tt.key, lowerPatterns)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldRedactValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		patterns []string
		want     bool
	}{
		{
			name:     "string with pattern",
			value:    "password=secret",
			patterns: []string{"password"},
			want:     true,
		},
		{
			name:     "string without pattern",
			value:    "normal value",
			patterns: []string{"password"},
			want:     false,
		},
		{
			name:     "nil value",
			value:    nil,
			patterns: []string{"password"},
			want:     false,
		},
		{
			name:     "number value",
			value:    12345,
			patterns: []string{"password"},
			want:     false,
		},
		{
			name:     "case insensitive value",
			value:    "TOKEN=abc123",
			patterns: []string{"token"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert patterns to lowercase for testing
			lowerPatterns := make([]string, len(tt.patterns))
			for i, p := range tt.patterns {
				lowerPatterns[i] = strings.ToLower(p)
			}

			got := shouldRedactValue(tt.value, lowerPatterns)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExportFormat tests exporting in different formats
func TestExportFormat(t *testing.T) {
	formats := []string{"json", "yaml", "msgpack"}

	for _, formatName := range formats {
		t.Run(formatName, func(t *testing.T) {
			// Create temp directory for test files
			tmpDir := t.TempDir()

			// Determine extension
			ext := ".json"
			if formatName == "yaml" {
				ext = ".yaml"
			} else if formatName == "msgpack" {
				ext = ".msgpack"
			}

			filename := filepath.Join(tmpDir, "test-export"+ext)

			// Create dev tools with store
			dt := &DevTools{
				enabled: true,
				store:   NewDevToolsStore(100, 100, 1000),
			}

			// Add some test data
			dt.store.AddComponent(&ComponentSnapshot{
				ID:        "comp-1",
				Name:      "TestComponent",
				Type:      "test",
				Timestamp: time.Now(),
			})

			// Export with format
			opts := ExportOptions{
				IncludeComponents: true,
			}

			err := dt.ExportFormat(filename, formatName, opts)
			require.NoError(t, err, "ExportFormat should succeed")

			// Verify file exists
			_, err = os.Stat(filename)
			require.NoError(t, err, "Export file should exist")

			// Verify file is not empty
			info, err := os.Stat(filename)
			require.NoError(t, err)
			assert.Greater(t, info.Size(), int64(0), "Export file should not be empty")
		})
	}
}

// TestExportFormat_WithCompression tests exporting with compression
func TestExportFormat_WithCompression(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-export.yaml.gz")

	// Create dev tools with store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add test data
	dt.store.AddComponent(&ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent",
		Type:      "test",
		Timestamp: time.Now(),
	})

	// Export with compression
	opts := ExportOptions{
		IncludeComponents: true,
		Compress:          true,
	}

	err := dt.ExportFormat(filename, "yaml", opts)
	require.NoError(t, err)

	// Verify file exists and is compressed
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Check for gzip magic bytes
	assert.Equal(t, byte(0x1f), data[0], "Should have gzip magic byte 1")
	assert.Equal(t, byte(0x8b), data[1], "Should have gzip magic byte 2")
}

// TestExportFormat_InvalidFormat tests error handling for invalid format
func TestExportFormat_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-export.txt")

	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	opts := ExportOptions{
		IncludeComponents: true,
	}

	err := dt.ExportFormat(filename, "invalid", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format not found")
}

// TestExportStream tests streaming export functionality
func TestExportStream(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-stream.json")

	// Create dev tools with store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add test data
	dt.store.AddComponent(&ComponentSnapshot{
		ID:        "comp-1",
		Name:      "TestComponent",
		Type:      "test",
		Timestamp: time.Now(),
	})

	// Export using streaming
	opts := ExportOptions{
		IncludeComponents: true,
		UseStreaming:      true,
	}

	err := dt.ExportStream(filename, opts)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filename)
	require.NoError(t, err)

	// Verify file contains valid JSON
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	var exportData ExportData
	err = json.Unmarshal(data, &exportData)
	require.NoError(t, err)

	// Verify data
	assert.Equal(t, "1.0", exportData.Version)
	assert.Len(t, exportData.Components, 1)
	assert.Equal(t, "comp-1", exportData.Components[0].ID)
}

// TestExportStream_WithSanitization tests streaming export with sanitization
func TestExportStream_WithSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-stream-sanitized.json")

	// Create dev tools with store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add test data with sensitive information
	dt.store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
		Props: map[string]interface{}{
			"password": "secret123",
			"username": "john",
		},
	})

	// Export with sanitization
	opts := ExportOptions{
		IncludeComponents: true,
		Sanitize:          true,
		RedactPatterns:    []string{"password"},
		UseStreaming:      true,
	}

	err := dt.ExportStream(filename, opts)
	require.NoError(t, err)

	// Read and verify sanitization
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Password should be redacted
	assert.Contains(t, string(data), "[REDACTED]")
	assert.NotContains(t, string(data), "secret123")
	// Username should not be redacted
	assert.Contains(t, string(data), "john")
}

// TestExportStream_WithProgressCallback tests progress reporting
func TestExportStream_WithProgressCallback(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-stream-progress.json")

	// Create dev tools with store
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	// Add test data
	dt.store.AddComponent(&ComponentSnapshot{
		ID:   "comp-1",
		Name: "TestComponent",
	})

	// Track progress callbacks
	var progressCalls int
	var lastBytes int64

	// Export with progress callback
	opts := ExportOptions{
		IncludeComponents: true,
		UseStreaming:      true,
		ProgressCallback: func(bytes int64) {
			progressCalls++
			lastBytes = bytes
		},
	}

	err := dt.ExportStream(filename, opts)
	require.NoError(t, err)

	// Verify progress was reported
	assert.Greater(t, progressCalls, 0, "Progress callback should be called")
	assert.Greater(t, lastBytes, int64(0), "Should report bytes processed")
}

// TestExportStream_NotEnabled tests error when dev tools not enabled
func TestExportStream_NotEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-stream.json")

	dt := &DevTools{
		enabled: false,
		store:   NewDevToolsStore(100, 100, 1000),
	}

	opts := ExportOptions{
		IncludeComponents: true,
	}

	err := dt.ExportStream(filename, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

// TestExportStream_NoStore tests error when store not initialized
func TestExportStream_NoStore(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-stream.json")

	dt := &DevTools{
		enabled: true,
		store:   nil,
	}

	opts := ExportOptions{
		IncludeComponents: true,
	}

	err := dt.ExportStream(filename, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestSanitizeExportData_EmptyPatterns tests sanitization with no patterns
func TestSanitizeExportData_EmptyPatterns(t *testing.T) {
	data := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "Test",
				Props: map[string]interface{}{
					"password": "secret",
				},
			},
		},
	}

	// Sanitize with empty patterns
	result := sanitizeExportData(data, []string{})

	// Data should be unchanged
	assert.Equal(t, "secret", result.Components[0].Props["password"])
}

// TestSanitizeExportData_NilMaps tests sanitization with nil maps
func TestSanitizeExportData_NilMaps(t *testing.T) {
	data := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:    "comp-1",
				Name:  "Test",
				Props: nil, // nil props
				State: nil, // nil state
			},
		},
	}

	// Should not panic
	result := sanitizeExportData(data, []string{"password"})
	assert.NotNil(t, result)
}

// TestSanitizeExportData_StateAndEvents tests sanitizing state and events
func TestSanitizeExportData_StateAndEvents(t *testing.T) {
	data := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		State: []StateChange{
			{
				RefID:    "ref-1",
				RefName:  "password",
				OldValue: "old-secret",
				NewValue: "new-secret",
			},
		},
		Events: []EventRecord{
			{
				ID:      "event-1",
				Name:    "submit",
				Payload: "password=secret123",
			},
		},
	}

	result := sanitizeExportData(data, []string{"password", "secret"})

	// State should be redacted
	assert.Equal(t, "[REDACTED]", result.State[0].OldValue)
	assert.Equal(t, "[REDACTED]", result.State[0].NewValue)

	// Event payload should be redacted
	assert.Equal(t, "[REDACTED]", result.Events[0].Payload)
}
