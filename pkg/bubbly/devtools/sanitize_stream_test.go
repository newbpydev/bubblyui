package devtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStreamSanitizer tests the constructor
func TestNewStreamSanitizer(t *testing.T) {
	base := NewSanitizer()

	stream := NewStreamSanitizer(base, 64*1024)
	assert.NotNil(t, stream)
	assert.Equal(t, 64*1024, stream.bufferSize)
}

// TestNewStreamSanitizer_DefaultBufferSize tests default buffer size
func TestNewStreamSanitizer_DefaultBufferSize(t *testing.T) {
	base := NewSanitizer()

	stream := NewStreamSanitizer(base, 0)
	assert.NotNil(t, stream)
	assert.Equal(t, 64*1024, stream.bufferSize) // Should default to 64KB
}

// TestStreamSanitizer_SanitizeStream_Basic tests basic streaming sanitization
func TestStreamSanitizer_SanitizeStream_Basic(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 1024)

	input := `{"password": "secret123", "username": "alice"}`
	reader := strings.NewReader(input)
	var output bytes.Buffer

	err := stream.SanitizeStream(reader, &output, nil)
	require.NoError(t, err)

	result := output.String()
	assert.Contains(t, result, "[REDACTED]")
	assert.Contains(t, result, "alice")
	assert.NotContains(t, result, "secret123")
}

// TestStreamSanitizer_SanitizeStream_ProgressCallback tests progress reporting
func TestStreamSanitizer_SanitizeStream_ProgressCallback(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 1024)

	// Create larger input to trigger multiple progress callbacks
	input := strings.Repeat(`{"password": "secret", "data": "value"}`, 100)
	reader := strings.NewReader(input)
	var output bytes.Buffer

	var progressCalls []int64
	var mu sync.Mutex
	progress := func(bytesProcessed int64) {
		mu.Lock()
		progressCalls = append(progressCalls, bytesProcessed)
		mu.Unlock()
	}

	err := stream.SanitizeStream(reader, &output, progress)
	require.NoError(t, err)

	mu.Lock()
	assert.Greater(t, len(progressCalls), 0, "Progress callback should be invoked")
	mu.Unlock()
}

// TestStreamSanitizer_SanitizeStream_LargeData tests handling of large data
func TestStreamSanitizer_SanitizeStream_LargeData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large data test in short mode")
	}

	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 64*1024)

	// Create ~10MB of data
	var inputBuilder strings.Builder
	inputBuilder.WriteString(`{"components":[`)
	for i := 0; i < 10000; i++ {
		if i > 0 {
			inputBuilder.WriteString(",")
		}
		inputBuilder.WriteString(fmt.Sprintf(`{"id":"comp-%d","password":"secret%d","data":"value%d"}`, i, i, i))
	}
	inputBuilder.WriteString(`]}`)

	reader := strings.NewReader(inputBuilder.String())
	var output bytes.Buffer

	// Force GC before measuring
	runtime.GC()

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	err := stream.SanitizeStream(reader, &output, nil)
	require.NoError(t, err)

	// Force GC after processing
	runtime.GC()

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Memory increase should be bounded
	// Note: We check HeapAlloc instead of Alloc to avoid overflow issues
	memIncrease := int64(memAfter.HeapAlloc) - int64(memBefore.HeapAlloc)
	if memIncrease < 0 {
		// Negative means GC ran, which is fine
		memIncrease = 0
	}
	assert.Less(t, memIncrease, int64(100*1024*1024), "Memory usage should stay under 100MB")

	// Verify sanitization worked
	result := output.String()
	assert.Contains(t, result, "[REDACTED]")
	assert.NotContains(t, result, "secret0")
}

// TestStreamSanitizer_SanitizeStream_InvalidJSON tests handling of malformed input
func TestStreamSanitizer_SanitizeStream_InvalidJSON(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 1024)

	tests := []struct {
		name  string
		input string
	}{
		{"incomplete object", `{"password": "secret"`},
		{"invalid syntax", `{password: secret}`},
		{"truncated array", `[{"data": "value"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var output bytes.Buffer

			// String-based sanitization doesn't validate JSON structure
			// It just applies regex patterns to the text
			err := stream.SanitizeStream(reader, &output, nil)
			assert.NoError(t, err, "String-based sanitization should handle any text")

			// Verify patterns were still applied
			result := output.String()
			if strings.Contains(tt.input, "secret") {
				assert.Contains(t, result, "[REDACTED]")
			}
		})
	}
}

// TestStreamSanitizer_SanitizeStream_EmptyInput tests handling of empty input
func TestStreamSanitizer_SanitizeStream_EmptyInput(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 1024)

	reader := strings.NewReader("")
	var output bytes.Buffer

	err := stream.SanitizeStream(reader, &output, nil)
	// Should handle gracefully (either no error with empty output, or specific error)
	if err != nil {
		assert.Contains(t, err.Error(), "EOF")
	} else {
		assert.Equal(t, "", output.String())
	}
}

// TestStreamSanitizer_SanitizeStream_BufferSizeConfiguration tests different buffer sizes
func TestStreamSanitizer_SanitizeStream_BufferSizeConfiguration(t *testing.T) {
	base := NewSanitizer()

	bufferSizes := []int{1024, 4096, 64 * 1024, 1024 * 1024}
	input := strings.Repeat(`{"password": "secret", "data": "value"}`, 1000)

	for _, size := range bufferSizes {
		t.Run(fmt.Sprintf("buffer_%d", size), func(t *testing.T) {
			stream := NewStreamSanitizer(base, size)
			reader := strings.NewReader(input)
			var output bytes.Buffer

			err := stream.SanitizeStream(reader, &output, nil)
			require.NoError(t, err)

			result := output.String()
			assert.Contains(t, result, "[REDACTED]")
		})
	}
}

// TestStreamSanitizer_SanitizeStream_Concurrent tests concurrent stream operations
func TestStreamSanitizer_SanitizeStream_Concurrent(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 64*1024)

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			input := fmt.Sprintf(`{"password": "secret%d", "data": "value%d"}`, id, id)
			reader := strings.NewReader(input)
			var output bytes.Buffer

			err := stream.SanitizeStream(reader, &output, nil)
			assert.NoError(t, err)

			result := output.String()
			assert.Contains(t, result, "[REDACTED]")
			assert.NotContains(t, result, fmt.Sprintf("secret%d", id))
		}(i)
	}

	wg.Wait()
}

// TestStreamSanitizer_SanitizeStream_PreservesStructure tests JSON structure validity
func TestStreamSanitizer_SanitizeStream_PreservesStructure(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 1024)

	input := `{
		"version": "1.0",
		"components": [
			{"id": "1", "password": "secret1"},
			{"id": "2", "password": "secret2"}
		],
		"metadata": {
			"token": "bearer123"
		}
	}`

	reader := strings.NewReader(input)
	var output bytes.Buffer

	err := stream.SanitizeStream(reader, &output, nil)
	require.NoError(t, err)

	// Verify output is valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(output.Bytes(), &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify structure preserved
	assert.Equal(t, "1.0", result["version"])
	assert.NotNil(t, result["components"])
	assert.NotNil(t, result["metadata"])
}

// TestStreamSanitizer_RoundTrip tests stream export followed by import
func TestStreamSanitizer_RoundTrip(t *testing.T) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 1024)

	original := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Props: map[string]interface{}{
					"password": "secret123",
					"username": "alice",
				},
			},
		},
	}

	// Marshal original to JSON
	originalJSON, err := json.Marshal(original)
	require.NoError(t, err)

	// Stream sanitize
	reader := bytes.NewReader(originalJSON)
	var output bytes.Buffer
	err = stream.SanitizeStream(reader, &output, nil)
	require.NoError(t, err)

	// Unmarshal sanitized data
	var sanitized ExportData
	err = json.Unmarshal(output.Bytes(), &sanitized)
	require.NoError(t, err)

	// Verify sanitization
	assert.Equal(t, "1.0", sanitized.Version)
	assert.Len(t, sanitized.Components, 1)
	assert.Contains(t, fmt.Sprintf("%v", sanitized.Components[0].Props["password"]), "[REDACTED]")
	assert.Equal(t, "alice", sanitized.Components[0].Props["username"])
}

// BenchmarkStreamSanitizer_InMemory benchmarks in-memory sanitization
func BenchmarkStreamSanitizer_InMemory(b *testing.B) {
	base := NewSanitizer()

	// Create test data
	data := ExportData{
		Version:    "1.0",
		Timestamp:  time.Now(),
		Components: make([]*ComponentSnapshot, 1000),
	}
	for i := 0; i < 1000; i++ {
		data.Components[i] = &ComponentSnapshot{
			ID:   fmt.Sprintf("comp-%d", i),
			Name: "TestComponent",
			Props: map[string]interface{}{
				"password": fmt.Sprintf("secret%d", i),
				"data":     fmt.Sprintf("value%d", i),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base.Sanitize(&data)
	}
}

// BenchmarkStreamSanitizer_Streaming benchmarks streaming sanitization
func BenchmarkStreamSanitizer_Streaming(b *testing.B) {
	base := NewSanitizer()
	stream := NewStreamSanitizer(base, 64*1024)

	// Create test data
	data := ExportData{
		Version:    "1.0",
		Timestamp:  time.Now(),
		Components: make([]*ComponentSnapshot, 1000),
	}
	for i := 0; i < 1000; i++ {
		data.Components[i] = &ComponentSnapshot{
			ID:   fmt.Sprintf("comp-%d", i),
			Name: "TestComponent",
			Props: map[string]interface{}{
				"password": fmt.Sprintf("secret%d", i),
				"data":     fmt.Sprintf("value%d", i),
			},
		}
	}

	jsonData, _ := json.Marshal(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(jsonData)
		var output bytes.Buffer
		_ = stream.SanitizeStream(reader, &output, nil)
	}
}
