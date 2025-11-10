package devtools

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizationStats_RedactedCount(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		patterns      []struct{ pattern, replacement string }
		expectedCount int
	}{
		{
			name:  "single password match",
			input: `{"password": "secret123"}`,
			patterns: []struct{ pattern, replacement string }{
				{`(?i)(password)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]"},
			},
			expectedCount: 1,
		},
		{
			name:  "multiple matches",
			input: `{"password": "secret123", "token": "abc", "apikey": "xyz"}`,
			patterns: []struct{ pattern, replacement string }{
				{`(?i)(password)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]"},
				{`(?i)(token)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]"},
				{`(?i)(apikey)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]"},
			},
			expectedCount: 3,
		},
		{
			name:          "no matches",
			input:         `{"username": "alice"}`,
			patterns:      []struct{ pattern, replacement string }{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sanitizer{patterns: make([]SanitizePattern, 0)}
			for _, p := range tt.patterns {
				s.AddPattern(p.pattern, p.replacement)
			}

			_ = s.SanitizeString(tt.input)
			stats := s.GetLastStats()

			require.NotNil(t, stats, "GetLastStats should return stats")
			assert.Equal(t, tt.expectedCount, stats.RedactedCount, "RedactedCount mismatch")
		})
	}
}

func TestSanitizationStats_PatternMatches(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedMatches map[string]int
	}{
		{
			name:  "track password pattern",
			input: `{"password": "secret123", "passwd": "abc"}`,
			expectedMatches: map[string]int{
				"pattern_0": 2, // Both password and passwd match same pattern
			},
		},
		{
			name:  "track multiple patterns",
			input: `{"password": "secret", "token": "abc", "apikey": "xyz"}`,
			expectedMatches: map[string]int{
				"pattern_0": 1, // password
				"pattern_1": 1, // token
				"pattern_2": 1, // apikey
			},
		},
		{
			name:            "no matches",
			input:           `{"username": "alice"}`,
			expectedMatches: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSanitizer()
			_ = s.SanitizeString(tt.input)
			stats := s.GetLastStats()

			require.NotNil(t, stats, "GetLastStats should return stats")
			assert.Equal(t, tt.expectedMatches, stats.PatternMatches, "PatternMatches mismatch")
		})
	}
}

func TestSanitizationStats_Duration(t *testing.T) {
	s := NewSanitizer()
	input := strings.Repeat(`{"password": "secret123"}`, 100)

	_ = s.SanitizeString(input)
	stats := s.GetLastStats()

	require.NotNil(t, stats, "GetLastStats should return stats")
	assert.Greater(t, stats.Duration, time.Duration(0), "Duration should be positive")
	assert.Less(t, stats.Duration, 1*time.Second, "Duration should be reasonable")
	assert.False(t, stats.StartTime.IsZero(), "StartTime should be set")
	assert.False(t, stats.EndTime.IsZero(), "EndTime should be set")
	assert.True(t, stats.EndTime.After(stats.StartTime), "EndTime should be after StartTime")
}

func TestSanitizationStats_BytesProcessed(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedBytes int64
	}{
		{
			name:          "simple string",
			input:         `{"password": "secret"}`,
			expectedBytes: int64(len(`{"password": "secret"}`)),
		},
		{
			name:          "empty string",
			input:         "",
			expectedBytes: 0,
		},
		{
			name:          "large string",
			input:         strings.Repeat("a", 1000),
			expectedBytes: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSanitizer()
			_ = s.SanitizeString(tt.input)
			stats := s.GetLastStats()

			require.NotNil(t, stats, "GetLastStats should return stats")
			assert.Equal(t, tt.expectedBytes, stats.BytesProcessed, "BytesProcessed mismatch")
		})
	}
}

func TestSanitizer_GetLastStats(t *testing.T) {
	s := NewSanitizer()

	// Before any sanitization
	stats := s.GetLastStats()
	assert.Nil(t, stats, "GetLastStats should return nil before first sanitization")

	// After sanitization
	_ = s.SanitizeString(`{"password": "secret"}`)
	stats = s.GetLastStats()
	require.NotNil(t, stats, "GetLastStats should return stats after sanitization")
	assert.Greater(t, stats.RedactedCount, 0, "Should have redacted values")

	// Second sanitization updates stats
	_ = s.SanitizeString(`{"token": "abc", "apikey": "xyz"}`)
	stats2 := s.GetLastStats()
	require.NotNil(t, stats2, "GetLastStats should return new stats")
	assert.NotEqual(t, stats.RedactedCount, stats2.RedactedCount, "Stats should be updated")
}

func TestSanitizer_ResetStats(t *testing.T) {
	s := NewSanitizer()

	// Sanitize and verify stats exist
	_ = s.SanitizeString(`{"password": "secret"}`)
	stats := s.GetLastStats()
	require.NotNil(t, stats, "Stats should exist after sanitization")

	// Reset stats
	s.ResetStats()

	// Verify stats are cleared
	stats = s.GetLastStats()
	assert.Nil(t, stats, "Stats should be nil after reset")
}

func TestSanitizationStats_ThreadSafety(t *testing.T) {
	s := NewSanitizer()
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent sanitizations
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			input := strings.Repeat(`{"password": "secret"}`, n%10+1)
			_ = s.SanitizeString(input)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.GetLastStats()
		}()
	}

	// Concurrent resets
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.ResetStats()
		}()
	}

	wg.Wait()

	// Should not panic and final stats should be valid
	stats := s.GetLastStats()
	if stats != nil {
		assert.GreaterOrEqual(t, stats.RedactedCount, 0, "RedactedCount should be non-negative")
		assert.GreaterOrEqual(t, stats.BytesProcessed, int64(0), "BytesProcessed should be non-negative")
	}
}

func TestSanitizationStats_String(t *testing.T) {
	tests := []struct {
		name     string
		stats    *SanitizationStats
		contains []string
	}{
		{
			name: "with matches",
			stats: &SanitizationStats{
				RedactedCount: 47,
				PatternMatches: map[string]int{
					"password": 23,
					"token":    15,
					"apikey":   9,
				},
				Duration: 142 * time.Millisecond,
			},
			contains: []string{"47", "password=23", "token=15", "apikey=9", "142ms"},
		},
		{
			name: "no matches",
			stats: &SanitizationStats{
				RedactedCount:  0,
				PatternMatches: map[string]int{},
				Duration:       10 * time.Millisecond,
			},
			contains: []string{"0", "10ms"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.String()
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr, "String output should contain %q", substr)
			}
		})
	}
}

func TestSanitizationStats_JSON(t *testing.T) {
	stats := &SanitizationStats{
		RedactedCount: 10,
		PatternMatches: map[string]int{
			"password": 5,
			"token":    5,
		},
		Duration:       100 * time.Millisecond,
		BytesProcessed: 1024,
		StartTime:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2024, 1, 1, 0, 0, 0, 100000000, time.UTC),
	}

	data, err := stats.JSON()
	require.NoError(t, err, "JSON() should not error")
	assert.NotEmpty(t, data, "JSON data should not be empty")

	// Verify it's valid JSON
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err, "Should be valid JSON")

	// Verify fields are present
	assert.Contains(t, decoded, "redacted_count")
	assert.Contains(t, decoded, "pattern_matches")
	assert.Contains(t, decoded, "duration_ms")
	assert.Contains(t, decoded, "bytes_processed")
}

func TestSanitize_WithStats(t *testing.T) {
	s := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Props: map[string]interface{}{
					"config": `{"password": "secret123", "token": "abc"}`,
					"auth":   "Bearer token123",
				},
			},
		},
	}

	result := s.Sanitize(data)
	stats := s.GetLastStats()

	require.NotNil(t, result, "Sanitize should return result")
	require.NotNil(t, stats, "Stats should be available after Sanitize")
	assert.Greater(t, stats.RedactedCount, 0, "Should have redacted values")
	assert.Greater(t, stats.BytesProcessed, int64(0), "Should have processed bytes")
	assert.NotEmpty(t, stats.PatternMatches, "Should have pattern matches")
}
