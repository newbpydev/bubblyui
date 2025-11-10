package devtools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONFormat tests the JSON format implementation
func TestJSONFormat(t *testing.T) {
	format := &JSONFormat{}

	t.Run("metadata", func(t *testing.T) {
		assert.Equal(t, "json", format.Name())
		assert.Equal(t, ".json", format.Extension())
		assert.Equal(t, "application/json", format.ContentType())
	})

	t.Run("marshal_unmarshal", func(t *testing.T) {
		// Create test data
		original := &ExportData{
			Version:   "1.0",
			Timestamp: time.Now().Truncate(time.Second), // Truncate for comparison
			Components: []*ComponentSnapshot{
				{
					ID:   "comp-1",
					Name: "TestComponent",
					Props: map[string]interface{}{
						"key": "value",
					},
				},
			},
		}

		// Marshal
		bytes, err := format.Marshal(original)
		require.NoError(t, err)
		assert.NotEmpty(t, bytes)

		// Unmarshal
		var restored ExportData
		err = format.Unmarshal(bytes, &restored)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, original.Version, restored.Version)
		assert.Equal(t, original.Timestamp.Unix(), restored.Timestamp.Unix())
		assert.Len(t, restored.Components, 1)
		assert.Equal(t, "comp-1", restored.Components[0].ID)
		assert.Equal(t, "TestComponent", restored.Components[0].Name)
	})

	t.Run("empty_data", func(t *testing.T) {
		data := &ExportData{
			Version:   "1.0",
			Timestamp: time.Now(),
		}

		bytes, err := format.Marshal(data)
		require.NoError(t, err)

		var restored ExportData
		err = format.Unmarshal(bytes, &restored)
		require.NoError(t, err)

		assert.Equal(t, data.Version, restored.Version)
	})
}

// TestYAMLFormat tests the YAML format implementation
func TestYAMLFormat(t *testing.T) {
	format := &YAMLFormat{}

	t.Run("metadata", func(t *testing.T) {
		assert.Equal(t, "yaml", format.Name())
		assert.Equal(t, ".yaml", format.Extension())
		assert.Equal(t, "application/x-yaml", format.ContentType())
	})

	t.Run("marshal_unmarshal", func(t *testing.T) {
		// Create test data
		original := &ExportData{
			Version:   "1.0",
			Timestamp: time.Now().Truncate(time.Second),
			State: []StateChange{
				{
					RefID:     "ref-1",
					RefName:   "counter",
					OldValue:  0,
					NewValue:  1,
					Timestamp: time.Now().Truncate(time.Second),
				},
			},
		}

		// Marshal
		bytes, err := format.Marshal(original)
		require.NoError(t, err)
		assert.NotEmpty(t, bytes)

		// Unmarshal
		var restored ExportData
		err = format.Unmarshal(bytes, &restored)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, original.Version, restored.Version)
		assert.Len(t, restored.State, 1)
		assert.Equal(t, "ref-1", restored.State[0].RefID)
		assert.Equal(t, "counter", restored.State[0].RefName)
	})

	t.Run("produces_valid_yaml", func(t *testing.T) {
		data := &ExportData{
			Version:   "1.0",
			Timestamp: time.Now(),
		}

		bytes, err := format.Marshal(data)
		require.NoError(t, err)

		// YAML should contain version field
		assert.Contains(t, string(bytes), "version:")
		assert.Contains(t, string(bytes), "timestamp:")
	})
}

// TestMessagePackFormat tests the MessagePack format implementation
func TestMessagePackFormat(t *testing.T) {
	format := &MessagePackFormat{}

	t.Run("metadata", func(t *testing.T) {
		assert.Equal(t, "msgpack", format.Name())
		assert.Equal(t, ".msgpack", format.Extension())
		assert.Equal(t, "application/msgpack", format.ContentType())
	})

	t.Run("marshal_unmarshal", func(t *testing.T) {
		// Create test data
		original := &ExportData{
			Version:   "1.0",
			Timestamp: time.Now().Truncate(time.Second),
			Events: []EventRecord{
				{
					ID:        "event-1",
					Name:      "click",
					SourceID:  "button",
					Timestamp: time.Now().Truncate(time.Second),
				},
			},
		}

		// Marshal
		bytes, err := format.Marshal(original)
		require.NoError(t, err)
		assert.NotEmpty(t, bytes)

		// Unmarshal
		var restored ExportData
		err = format.Unmarshal(bytes, &restored)
		require.NoError(t, err)

		// Verify
		assert.Equal(t, original.Version, restored.Version)
		assert.Len(t, restored.Events, 1)
		assert.Equal(t, "event-1", restored.Events[0].ID)
		assert.Equal(t, "click", restored.Events[0].Name)
	})

	t.Run("binary_format", func(t *testing.T) {
		data := &ExportData{
			Version:   "1.0",
			Timestamp: time.Now(),
		}

		bytes, err := format.Marshal(data)
		require.NoError(t, err)

		// MessagePack is binary, should not be valid UTF-8 text
		// (though it might contain some readable parts)
		assert.NotEmpty(t, bytes)
	})
}

// TestFormatRegistry tests the format registry
func TestFormatRegistry(t *testing.T) {
	t.Run("register_and_get", func(t *testing.T) {
		registry := NewFormatRegistry()
		format := &JSONFormat{}

		registry.Register(format)

		retrieved, err := registry.Get("json")
		require.NoError(t, err)
		assert.Equal(t, format, retrieved)
	})

	t.Run("get_nonexistent", func(t *testing.T) {
		registry := NewFormatRegistry()

		_, err := registry.Get("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "format not found")
	})

	t.Run("get_all", func(t *testing.T) {
		registry := NewFormatRegistry()
		registry.Register(&JSONFormat{})
		registry.Register(&YAMLFormat{})

		formats := registry.GetAll()
		assert.Len(t, formats, 2)
		assert.Contains(t, formats, "json")
		assert.Contains(t, formats, "yaml")
	})

	t.Run("replace_format", func(t *testing.T) {
		registry := NewFormatRegistry()
		format1 := &JSONFormat{}
		format2 := &JSONFormat{}

		registry.Register(format1)
		registry.Register(format2)

		retrieved, err := registry.Get("json")
		require.NoError(t, err)
		assert.Equal(t, format2, retrieved)
	})
}

// TestGlobalRegistry tests the global registry functions
func TestGlobalRegistry(t *testing.T) {
	t.Run("get_supported_formats", func(t *testing.T) {
		formats := GetSupportedFormats()
		assert.GreaterOrEqual(t, len(formats), 3) // At least JSON, YAML, MessagePack
		assert.Contains(t, formats, "json")
		assert.Contains(t, formats, "yaml")
		assert.Contains(t, formats, "msgpack")
	})

	t.Run("register_custom_format", func(t *testing.T) {
		// Verify the registration mechanism works
		// by checking that we can register the built-in formats
		err := RegisterFormat(&JSONFormat{})
		require.NoError(t, err)

		// Verify built-in formats are registered
		formats := GetSupportedFormats()
		assert.Contains(t, formats, "json")
		assert.Contains(t, formats, "yaml")
		assert.Contains(t, formats, "msgpack")
	})
}

// TestDetectFormat tests format detection
func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
		wantErr  bool
	}{
		{
			name:     "json_extension",
			filename: "debug.json",
			expected: "json",
			wantErr:  false,
		},
		{
			name:     "yaml_extension",
			filename: "debug.yaml",
			expected: "yaml",
			wantErr:  false,
		},
		{
			name:     "yml_extension",
			filename: "debug.yml",
			expected: "yaml",
			wantErr:  false,
		},
		{
			name:     "msgpack_extension",
			filename: "debug.msgpack",
			expected: "msgpack",
			wantErr:  false,
		},
		{
			name:     "mp_extension",
			filename: "debug.mp",
			expected: "msgpack",
			wantErr:  false,
		},
		{
			name:     "json_with_gzip",
			filename: "debug.json.gz",
			expected: "json",
			wantErr:  false,
		},
		{
			name:     "yaml_with_gzip",
			filename: "debug.yaml.gz",
			expected: "yaml",
			wantErr:  false,
		},
		{
			name:     "msgpack_with_gzip",
			filename: "debug.msgpack.gz",
			expected: "msgpack",
			wantErr:  false,
		},
		{
			name:     "unknown_extension",
			filename: "debug.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "no_extension",
			filename: "debug",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "path_with_json",
			filename: "/path/to/debug.json",
			expected: "json",
			wantErr:  false,
		},
		{
			name:     "uppercase_extension",
			filename: "debug.JSON",
			expected: "json",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := DetectFormat(tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, format)
			}
		})
	}
}

// TestFormatRoundTrip tests round-trip for all formats
func TestFormatRoundTrip(t *testing.T) {
	// Create comprehensive test data
	original := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now().Truncate(time.Second),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "Counter",
				Props: map[string]interface{}{
					"initial": 0,
					"step":    1,
				},
				State: map[string]interface{}{
					"count": 42,
				},
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "count",
						Value: 42,
					},
				},
			},
		},
		State: []StateChange{
			{
				RefID:     "ref-1",
				RefName:   "count",
				OldValue:  41,
				NewValue:  42,
				Timestamp: time.Now().Truncate(time.Second),
			},
		},
		Events: []EventRecord{
			{
				ID:        "event-1",
				Name:      "increment",
				SourceID:  "button",
				Timestamp: time.Now().Truncate(time.Second),
			},
		},
	}

	formats := []ExportFormat{
		&JSONFormat{},
		&YAMLFormat{},
		&MessagePackFormat{},
	}

	for _, format := range formats {
		t.Run(format.Name(), func(t *testing.T) {
			// Marshal
			bytes, err := format.Marshal(original)
			require.NoError(t, err)
			assert.NotEmpty(t, bytes)

			// Unmarshal
			var restored ExportData
			err = format.Unmarshal(bytes, &restored)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, original.Version, restored.Version)
			assert.Equal(t, original.Timestamp.Unix(), restored.Timestamp.Unix())
			assert.Len(t, restored.Components, 1)
			assert.Len(t, restored.State, 1)
			assert.Len(t, restored.Events, 1)

			// Verify component details
			assert.Equal(t, "comp-1", restored.Components[0].ID)
			assert.Equal(t, "Counter", restored.Components[0].Name)

			// Verify state details
			assert.Equal(t, "ref-1", restored.State[0].RefID)
			assert.Equal(t, "count", restored.State[0].RefName)

			// Verify event details
			assert.Equal(t, "event-1", restored.Events[0].ID)
			assert.Equal(t, "increment", restored.Events[0].Name)
		})
	}
}

// TestFormatSizeComparison tests relative sizes of different formats
func TestFormatSizeComparison(t *testing.T) {
	// Create test data with enough content to show size differences
	data := &ExportData{
		Version:    "1.0",
		Timestamp:  time.Now(),
		Components: make([]*ComponentSnapshot, 10),
		State:      make([]StateChange, 50),
		Events:     make([]EventRecord, 100),
	}

	// Populate with test data
	for i := 0; i < 10; i++ {
		data.Components[i] = &ComponentSnapshot{
			ID:   "comp-" + string(rune(i)),
			Name: "Component",
		}
	}

	formats := map[string]ExportFormat{
		"json":    &JSONFormat{},
		"yaml":    &YAMLFormat{},
		"msgpack": &MessagePackFormat{},
	}

	sizes := make(map[string]int)

	for name, format := range formats {
		bytes, err := format.Marshal(data)
		require.NoError(t, err)
		sizes[name] = len(bytes)
	}

	// JSON is baseline (100%)
	jsonSize := sizes["json"]

	// YAML should be smaller or similar to JSON (60-110%)
	// YAML is actually quite efficient, often smaller than JSON
	yamlRatio := float64(sizes["yaml"]) / float64(jsonSize)
	assert.Less(t, yamlRatio, 1.2, "YAML size should not be much larger than JSON")

	// MessagePack should be smaller (40-80%)
	msgpackRatio := float64(sizes["msgpack"]) / float64(jsonSize)
	assert.Less(t, msgpackRatio, 0.9, "MessagePack should be smaller than JSON")

	t.Logf("Size comparison (JSON=100%%): JSON=%d, YAML=%d (%.0f%%), MessagePack=%d (%.0f%%)",
		jsonSize, sizes["yaml"], yamlRatio*100, sizes["msgpack"], msgpackRatio*100)
}

// TestConcurrentFormatAccess tests thread safety of format operations
func TestConcurrentFormatAccess(t *testing.T) {
	registry := NewFormatRegistry()
	registry.Register(&JSONFormat{})
	registry.Register(&YAMLFormat{})

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// Get formats
			_, _ = registry.Get("json")
			_, _ = registry.Get("yaml")

			// Get all formats
			_ = registry.GetAll()

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
