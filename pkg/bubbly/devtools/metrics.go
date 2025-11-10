package devtools

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// SanitizationStats tracks statistics from a sanitization operation.
//
// It records the number of values redacted, which patterns matched,
// how long the operation took, and how many bytes were processed.
// This information is useful for auditing, performance monitoring,
// and validating sanitization effectiveness.
//
// Example:
//
//	sanitizer := NewSanitizer()
//	_ = sanitizer.SanitizeString(`{"password": "secret"}`)
//	stats := sanitizer.GetLastStats()
//	fmt.Println(stats.String())
//	// Output: Redacted 1 values: pattern_0=1 (5ms)
type SanitizationStats struct {
	// RedactedCount is the total number of values redacted
	RedactedCount int `json:"redacted_count"`

	// PatternMatches maps pattern names to their match counts
	PatternMatches map[string]int `json:"pattern_matches"`

	// Duration is how long the sanitization took
	Duration time.Duration `json:"duration_ms"`

	// BytesProcessed is the total number of bytes processed
	BytesProcessed int64 `json:"bytes_processed"`

	// StartTime is when sanitization started
	StartTime time.Time `json:"start_time"`

	// EndTime is when sanitization completed
	EndTime time.Time `json:"end_time"`
}

// String returns a human-readable representation of the stats.
//
// Format: "Redacted N values: pattern1=X, pattern2=Y (duration)"
//
// Example:
//
//	stats := &SanitizationStats{
//	    RedactedCount: 47,
//	    PatternMatches: map[string]int{"password": 23, "token": 15, "apikey": 9},
//	    Duration: 142 * time.Millisecond,
//	}
//	fmt.Println(stats.String())
//	// Output: Redacted 47 values: apikey=9, password=23, token=15 (142ms)
//
// Returns:
//   - string: Human-readable stats summary
func (s *SanitizationStats) String() string {
	if s == nil {
		return "No sanitization stats available"
	}

	// Build pattern matches string (sorted by pattern name for consistency)
	var patterns []string
	if len(s.PatternMatches) > 0 {
		// Get sorted pattern names
		names := make([]string, 0, len(s.PatternMatches))
		for name := range s.PatternMatches {
			names = append(names, name)
		}
		sort.Strings(names)

		// Build "pattern=count" strings
		for _, name := range names {
			count := s.PatternMatches[name]
			patterns = append(patterns, fmt.Sprintf("%s=%d", name, count))
		}
	}

	// Format duration in milliseconds
	durationMs := s.Duration.Milliseconds()

	// Build final string
	if len(patterns) > 0 {
		return fmt.Sprintf("Redacted %d values: %s (%dms)",
			s.RedactedCount,
			strings.Join(patterns, ", "),
			durationMs)
	}

	return fmt.Sprintf("Redacted %d values (%dms)", s.RedactedCount, durationMs)
}

// JSON returns a JSON representation of the stats.
//
// The JSON includes all fields with appropriate types:
//   - redacted_count: number
//   - pattern_matches: object mapping pattern names to counts
//   - duration_ms: number (milliseconds)
//   - bytes_processed: number
//   - start_time: ISO 8601 timestamp
//   - end_time: ISO 8601 timestamp
//
// Example:
//
//	stats := &SanitizationStats{
//	    RedactedCount: 10,
//	    PatternMatches: map[string]int{"password": 5, "token": 5},
//	    Duration: 100 * time.Millisecond,
//	    BytesProcessed: 1024,
//	}
//	data, _ := stats.JSON()
//	fmt.Println(string(data))
//	// Output: {"redacted_count":10,"pattern_matches":{"password":5,"token":5},...}
//
// Returns:
//   - []byte: JSON-encoded stats
//   - error: Error if JSON encoding fails
func (s *SanitizationStats) JSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(map[string]interface{}{
			"redacted_count":  0,
			"pattern_matches": map[string]int{},
			"duration_ms":     0,
			"bytes_processed": 0,
		})
	}

	// Create a struct with duration in milliseconds for JSON
	type jsonStats struct {
		RedactedCount  int            `json:"redacted_count"`
		PatternMatches map[string]int `json:"pattern_matches"`
		DurationMs     int64          `json:"duration_ms"`
		BytesProcessed int64          `json:"bytes_processed"`
		StartTime      time.Time      `json:"start_time"`
		EndTime        time.Time      `json:"end_time"`
	}

	js := jsonStats{
		RedactedCount:  s.RedactedCount,
		PatternMatches: s.PatternMatches,
		DurationMs:     s.Duration.Milliseconds(),
		BytesProcessed: s.BytesProcessed,
		StartTime:      s.StartTime,
		EndTime:        s.EndTime,
	}

	return json.Marshal(js)
}
