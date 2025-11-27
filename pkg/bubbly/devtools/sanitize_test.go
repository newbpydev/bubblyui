package devtools

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSanitizer(t *testing.T) {
	sanitizer := NewSanitizer()

	assert.NotNil(t, sanitizer)
	assert.Equal(t, 4, sanitizer.PatternCount(), "Should have 4 default patterns")
}

func TestSanitizer_Sanitize_Passwords(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "JSON format",
			input: `{"password": "secret123"}`,
			want:  `{"password": "[REDACTED]"}`,
		},
		{
			name:  "URL param format",
			input: "password=secret123",
			want:  "password=[REDACTED]",
		},
		{
			name:  "case insensitive",
			input: "PASSWORD: secret123",
			want:  "PASSWORD: [REDACTED]",
		},
		{
			name:  "passwd variant",
			input: "passwd=mypass",
			want:  "passwd=[REDACTED]",
		},
		{
			name:  "pwd variant",
			input: "pwd: test123",
			want:  "pwd: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_Sanitize_Tokens(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bearer token",
			input: "Authorization: Bearer abc123xyz",
			want:  "Authorization: Bearer [REDACTED]",
		},
		{
			name:  "token field",
			input: `{"token": "xyz789"}`,
			want:  `{"token": "[REDACTED]"}`,
		},
		{
			name:  "case insensitive",
			input: "TOKEN=mytoken",
			want:  "TOKEN=[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_Sanitize_APIKeys(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "api_key underscore",
			input: "api_key: sk_test_123",
			want:  "api_key: [REDACTED]",
		},
		{
			name:  "api-key hyphen",
			input: "api-key=abc123",
			want:  "api-key=[REDACTED]",
		},
		{
			name:  "apikey no separator",
			input: `{"apikey": "key123"}`,
			want:  `{"apikey": "[REDACTED]"}`,
		},
		{
			name:  "case insensitive",
			input: "API_KEY: test",
			want:  "API_KEY: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_Sanitize_Secrets(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "secret field",
			input: `{"secret": "mysecret"}`,
			want:  `{"secret": "[REDACTED]"}`,
		},
		{
			name:  "private_key underscore",
			input: "private_key: -----BEGIN",
			want:  "private_key: [REDACTED]",
		},
		{
			name:  "private-key hyphen",
			input: "private-key=rsa123",
			want:  "private-key=[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_Sanitize_CustomPatterns(t *testing.T) {
	sanitizer := NewSanitizer()
	sanitizer.AddPattern(`(?i)(credit[_-]?card)(["'\s:=]+)(\d+)`, "${1}${2}[CARD_REDACTED]")
	sanitizer.AddPattern(`\b\d{3}-\d{2}-\d{4}\b`, "[SSN_REDACTED]")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "credit card",
			input: "credit_card: 1234567890",
			want:  "credit_card: [CARD_REDACTED]",
		},
		{
			name:  "SSN pattern",
			input: "SSN: 123-45-6789",
			want:  "SSN: [SSN_REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizer.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizer_SanitizeValue_NestedMaps(t *testing.T) {
	sanitizer := NewSanitizer()

	input := map[string]interface{}{
		"username": "alice",
		"password": "secret123",
		"config": map[string]interface{}{
			"api_key": "key_abc123",
			"timeout": 30,
		},
	}

	result := sanitizer.SanitizeValue(input).(map[string]interface{})

	assert.Equal(t, "alice", result["username"])
	assert.Equal(t, "secret123", result["password"]) // Value not matched, key matched in string context
	assert.Equal(t, 30, result["config"].(map[string]interface{})["timeout"])
}

func TestSanitizer_SanitizeValue_Slices(t *testing.T) {
	sanitizer := NewSanitizer()

	input := []interface{}{
		"normal value",
		"password: secret123",
		map[string]interface{}{
			"token": "abc123",
		},
	}

	result := sanitizer.SanitizeValue(input).([]interface{})

	assert.Len(t, result, 3)
	assert.Equal(t, "normal value", result[0])
	assert.Contains(t, result[1], "[REDACTED]")
}

func TestSanitizer_SanitizeValue_Primitives(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input interface{}
	}{
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeValue(tt.input)
			assert.Equal(t, tt.input, result)
		})
	}
}

func TestSanitizer_AddPattern(t *testing.T) {
	sanitizer := NewSanitizer()
	initialCount := sanitizer.PatternCount()

	sanitizer.AddPattern(`test_pattern`, "[TEST]")

	assert.Equal(t, initialCount+1, sanitizer.PatternCount())

	// Test the added pattern works
	result := sanitizer.SanitizeString("test_pattern value")
	assert.Contains(t, result, "[TEST]")
}

func TestSanitizer_EmptyPatterns(t *testing.T) {
	// Create sanitizer without default patterns
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	input := "password: secret123"
	result := sanitizer.SanitizeString(input)

	// Should return unchanged
	assert.Equal(t, input, result)
}

func TestSanitizer_Integration_ExportData(t *testing.T) {
	sanitizer := NewSanitizer()

	// Create export data with sensitive information
	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "LoginForm",
				Props: map[string]interface{}{
					"username": "alice",
					"password": "secret123",
				},
				State: map[string]interface{}{
					"token": "bearer abc123xyz",
				},
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "apiKey",
						Value: "api_key: sk_test_123",
					},
				},
			},
		},
		State: []StateChange{
			{
				RefID:     "ref-1",
				RefName:   "password",
				OldValue:  "password: oldpass",
				NewValue:  "password: newpass",
				Timestamp: time.Now(),
			},
		},
		Events: []EventRecord{
			{
				ID:        "event-1",
				Name:      "login",
				Timestamp: time.Now(),
				Payload:   "token: xyz789",
			},
		},
	}

	// Sanitize
	result := sanitizer.Sanitize(data)

	// Verify structure preserved
	assert.Equal(t, data.Version, result.Version)
	assert.Len(t, result.Components, 1)
	assert.Len(t, result.State, 1)
	assert.Len(t, result.Events, 1)

	// Verify sensitive data redacted
	comp := result.Components[0]
	assert.Equal(t, "alice", comp.Props["username"])     // Not sensitive
	assert.Equal(t, "secret123", comp.Props["password"]) // Value itself not matched

	// Check refs
	assert.Contains(t, comp.Refs[0].Value, "[REDACTED]")

	// Check state
	assert.Contains(t, result.State[0].OldValue, "[REDACTED]")
	assert.Contains(t, result.State[0].NewValue, "[REDACTED]")

	// Check events
	assert.Contains(t, result.Events[0].Payload, "[REDACTED]")
}

func TestSanitizer_Sanitize_NilData(t *testing.T) {
	sanitizer := NewSanitizer()

	result := sanitizer.Sanitize(nil)
	assert.Nil(t, result)
}

func TestSanitizer_Sanitize_EmptyData(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
	}

	result := sanitizer.Sanitize(data)

	assert.NotNil(t, result)
	assert.Equal(t, data.Version, result.Version)
	assert.Nil(t, result.Components)
	assert.Nil(t, result.State)
	assert.Nil(t, result.Events)
}

func TestSanitizer_SanitizeValue_DeepNesting(t *testing.T) {
	sanitizer := NewSanitizer()

	input := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"secret": "password: deep_secret",
				},
			},
		},
	}

	result := sanitizer.SanitizeValue(input).(map[string]interface{})

	level3 := result["level1"].(map[string]interface{})["level2"].(map[string]interface{})["level3"].(map[string]interface{})
	assert.Contains(t, level3["secret"], "[REDACTED]")
}

func TestSanitizer_SanitizeValue_MixedTypes(t *testing.T) {
	sanitizer := NewSanitizer()

	input := map[string]interface{}{
		"string":  "token: abc123",
		"number":  42,
		"boolean": true,
		"slice":   []interface{}{"password: test", 123},
		"map": map[string]interface{}{
			"nested": "api_key: xyz",
		},
	}

	result := sanitizer.SanitizeValue(input).(map[string]interface{})

	assert.Contains(t, result["string"], "[REDACTED]")
	assert.Equal(t, 42, result["number"])
	assert.Equal(t, true, result["boolean"])
	assert.Len(t, result["slice"], 2)
	assert.Contains(t, result["map"].(map[string]interface{})["nested"], "[REDACTED]")
}

func TestDefaultPatterns(t *testing.T) {
	patterns := DefaultPatterns()

	assert.Len(t, patterns, 4)
	assert.Contains(t, patterns[0], "password")
	assert.Contains(t, patterns[1], "token")
	assert.Contains(t, patterns[2], "api")
	assert.Contains(t, patterns[3], "secret")
}

func TestSanitizer_SanitizeString(t *testing.T) {
	sanitizer := NewSanitizer()

	input := `{"username": "alice", "password": "secret123", "token": "xyz789"}`
	result := sanitizer.SanitizeString(input)

	assert.Contains(t, result, "alice")
	assert.Contains(t, result, "[REDACTED]")
	assert.NotContains(t, result, "secret123")
	assert.NotContains(t, result, "xyz789")
}

func TestSanitizer_PatternCount(t *testing.T) {
	sanitizer := NewSanitizer()
	assert.Equal(t, 4, sanitizer.PatternCount())

	sanitizer.AddPattern(`test`, "[TEST]")
	assert.Equal(t, 5, sanitizer.PatternCount())
}

func TestSanitizer_AddPattern_InvalidRegex(t *testing.T) {
	sanitizer := NewSanitizer()

	// Invalid regex should panic
	assert.Panics(t, func() {
		sanitizer.AddPattern(`[invalid`, "[TEST]")
	})
}

func TestSanitizer_SanitizeComponent_WithChildren(t *testing.T) {
	sanitizer := NewSanitizer()

	comp := &ComponentSnapshot{
		ID:   "parent",
		Name: "Parent",
		Props: map[string]interface{}{
			"password": "secret",
		},
		Children: []*ComponentSnapshot{
			{
				ID:   "child",
				Name: "Child",
				Props: map[string]interface{}{
					"token": "abc123",
				},
			},
		},
	}

	result := sanitizer.sanitizeComponent(comp)

	assert.Equal(t, "parent", result.ID)
	assert.Len(t, result.Children, 1)
	assert.Equal(t, "child", result.Children[0].ID)
}

func TestSanitizer_SanitizeValue_Pointer(t *testing.T) {
	sanitizer := NewSanitizer()

	str := "password: secret"
	input := &str

	result := sanitizer.SanitizeValue(input).(*string)

	assert.Contains(t, *result, "[REDACTED]")
	assert.NotEqual(t, str, *result) // Original unchanged
}

func TestSanitizer_SanitizeValue_NilPointer(t *testing.T) {
	sanitizer := NewSanitizer()

	var input *string = nil

	result := sanitizer.SanitizeValue(input)

	assert.Nil(t, result)
}

func TestSanitizer_Sanitize_PreservesOriginal(t *testing.T) {
	sanitizer := NewSanitizer()

	original := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID: "comp-1",
				Props: map[string]interface{}{
					"data": "password: secret123",
				},
			},
		},
	}

	// Store original value
	originalData := original.Components[0].Props["data"]

	// Sanitize
	result := sanitizer.Sanitize(original)

	// Verify original unchanged
	assert.Equal(t, originalData, original.Components[0].Props["data"])

	// Verify result is different (password redacted)
	assert.Contains(t, result.Components[0].Props["data"], "[REDACTED]")
	assert.NotEqual(t, originalData, result.Components[0].Props["data"])
}

// TestSanitizer_AddPatternWithPriority tests adding patterns with priority
func TestSanitizer_AddPatternWithPriority(t *testing.T) {
	sanitizer := NewSanitizer()
	initialCount := sanitizer.PatternCount()

	err := sanitizer.AddPatternWithPriority(`test_pattern`, "[TEST]", 50, "test_pattern")
	assert.NoError(t, err)
	assert.Equal(t, initialCount+1, sanitizer.PatternCount())

	// Test the added pattern works
	result := sanitizer.SanitizeString("test_pattern value")
	assert.Contains(t, result, "[TEST]")
}

// TestSanitizer_AddPatternWithPriority_InvalidRegex tests error handling
func TestSanitizer_AddPatternWithPriority_InvalidRegex(t *testing.T) {
	sanitizer := NewSanitizer()

	err := sanitizer.AddPatternWithPriority(`[invalid`, "[TEST]", 50, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern")
}

// TestSanitizer_PriorityOrdering tests that higher priority patterns apply first
func TestSanitizer_PriorityOrdering(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	// Add patterns in reverse priority order
	_ = sanitizer.AddPatternWithPriority(`data`, "[LOW]", 10, "low_priority")
	_ = sanitizer.AddPatternWithPriority(`data`, "[HIGH]", 100, "high_priority")
	_ = sanitizer.AddPatternWithPriority(`data`, "[MEDIUM]", 50, "medium_priority")

	// Higher priority should apply first
	result := sanitizer.SanitizeString("data value")
	assert.Equal(t, "[HIGH] value", result, "Higher priority pattern should apply first")
}

// TestSanitizer_EqualPriority_InsertionOrder tests stable sort with equal priorities
func TestSanitizer_EqualPriority_InsertionOrder(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	// Add patterns with same priority
	_ = sanitizer.AddPatternWithPriority(`data`, "[FIRST]", 50, "first")
	_ = sanitizer.AddPatternWithPriority(`data`, "[SECOND]", 50, "second")
	_ = sanitizer.AddPatternWithPriority(`data`, "[THIRD]", 50, "third")

	// First added should apply first (stable sort)
	result := sanitizer.SanitizeString("data value")
	assert.Equal(t, "[FIRST] value", result, "Equal priority should use insertion order")
}

// TestSanitizer_DefaultPriority tests that priority 0 is default behavior
func TestSanitizer_DefaultPriority(t *testing.T) {
	sanitizer := NewSanitizer()

	// Add pattern with priority 0 (default)
	_ = sanitizer.AddPatternWithPriority(`custom`, "[CUSTOM]", 0, "custom")

	// Should work normally
	result := sanitizer.SanitizeString("custom value")
	assert.Contains(t, result, "[CUSTOM]")
}

// TestSanitizer_NegativePriority tests that negative priorities apply last
func TestSanitizer_NegativePriority(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	// Add patterns with different priorities including negative
	_ = sanitizer.AddPatternWithPriority(`data`, "[POSITIVE]", 10, "positive")
	_ = sanitizer.AddPatternWithPriority(`data`, "[NEGATIVE]", -10, "negative")

	// Positive priority should apply first
	result := sanitizer.SanitizeString("data value")
	assert.Equal(t, "[POSITIVE] value", result, "Positive priority should apply before negative")
}

// TestSanitizer_OverlappingPatterns tests priority resolution for overlapping patterns
func TestSanitizer_OverlappingPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []struct {
			pattern     string
			replacement string
			priority    int
			name        string
		}
		input string
		want  string
	}{
		{
			name: "specific pattern applies first, then generic (both apply)",
			patterns: []struct {
				pattern     string
				replacement string
				priority    int
				name        string
			}{
				{`(?i)(password)(["'\s:=]+)([^\s"']+)`, "${1}${2}[GENERIC]", 10, "generic"},
				{`(?i)(password)(["'\s:=]+)(admin[^\s"']*)`, "${1}${2}[ADMIN]", 100, "admin_specific"},
			},
			input: "password: admin123",
			// Higher priority applies first: "password: [ADMIN]"
			// Then lower priority also applies: "password: [GENERIC]"
			want: "password: [GENERIC]",
		},
		{
			name: "credit card before generic number",
			patterns: []struct {
				pattern     string
				replacement string
				priority    int
				name        string
			}{
				{`\d+`, "[NUMBER]", 10, "generic_number"},
				{`\d{4}-\d{4}-\d{4}-\d{4}`, "[CARD]", 100, "credit_card"},
			},
			input: "card: 1234-5678-9012-3456",
			// Higher priority applies first: "card: [CARD]"
			// Lower priority doesn't match "[CARD]" so result stays
			want: "card: [CARD]",
		},
		{
			name: "all patterns apply in priority order",
			patterns: []struct {
				pattern     string
				replacement string
				priority    int
				name        string
			}{
				{`secret`, "[REDACTED]", 100, "high_priority"},
				{`\[REDACTED\]`, "[DOUBLE_REDACTED]", 10, "low_priority"},
			},
			input: "secret value",
			// High priority applies first: "[REDACTED] value"
			// Low priority also applies: "[DOUBLE_REDACTED] value"
			want: "[DOUBLE_REDACTED] value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := &Sanitizer{
				patterns: []SanitizePattern{},
			}

			for _, p := range tt.patterns {
				_ = sanitizer.AddPatternWithPriority(p.pattern, p.replacement, p.priority, p.name)
			}

			got := sanitizer.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestSanitizer_GetPatterns tests retrieving patterns in sorted order
func TestSanitizer_GetPatterns(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	// Add patterns in random order
	_ = sanitizer.AddPatternWithPriority(`low`, "[LOW]", 10, "low")
	_ = sanitizer.AddPatternWithPriority(`high`, "[HIGH]", 100, "high")
	_ = sanitizer.AddPatternWithPriority(`medium`, "[MEDIUM]", 50, "medium")

	patterns := sanitizer.GetPatterns()

	// Should be sorted by priority (descending)
	assert.Len(t, patterns, 3)
	assert.Equal(t, "high", patterns[0].Name)
	assert.Equal(t, 100, patterns[0].Priority)
	assert.Equal(t, "medium", patterns[1].Name)
	assert.Equal(t, 50, patterns[1].Priority)
	assert.Equal(t, "low", patterns[2].Name)
	assert.Equal(t, 10, patterns[2].Priority)
}

// TestSanitizer_PatternNames tests that pattern names are tracked correctly
func TestSanitizer_PatternNames(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	_ = sanitizer.AddPatternWithPriority(`test`, "[TEST]", 50, "my_test_pattern")

	patterns := sanitizer.GetPatterns()
	assert.Len(t, patterns, 1)
	assert.Equal(t, "my_test_pattern", patterns[0].Name)
}

// TestSanitizer_AutoGeneratedNames tests auto-generated pattern names
func TestSanitizer_AutoGeneratedNames(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	// Add pattern with empty name
	_ = sanitizer.AddPatternWithPriority(`test`, "[TEST]", 50, "")

	patterns := sanitizer.GetPatterns()
	assert.Len(t, patterns, 1)
	assert.Contains(t, patterns[0].Name, "pattern_")
}

// TestSanitizer_SortStability tests that sort is stable with many patterns
func TestSanitizer_SortStability(t *testing.T) {
	sanitizer := &Sanitizer{
		patterns: []SanitizePattern{},
	}

	// Add many patterns with same priority
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("pattern_%d", i)
		_ = sanitizer.AddPatternWithPriority(`test`, "[TEST]", 50, name)
	}

	patterns := sanitizer.GetPatterns()
	assert.Len(t, patterns, 10)

	// Verify insertion order preserved
	for i := 0; i < 10; i++ {
		expected := fmt.Sprintf("pattern_%d", i)
		assert.Equal(t, expected, patterns[i].Name)
	}
}

// TestSanitizer_PriorityRanges tests documented priority ranges
func TestSanitizer_PriorityRanges(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		category string
	}{
		{"critical", 150, "critical (100+)"},
		{"org_specific", 75, "org-specific (50-99)"},
		{"custom", 30, "custom (10-49)"},
		{"default", 5, "default (0-9)"},
		{"cleanup", -5, "cleanup (negative)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := &Sanitizer{
				patterns: []SanitizePattern{},
			}

			err := sanitizer.AddPatternWithPriority(`test`, "[TEST]", tt.priority, tt.name)
			assert.NoError(t, err)

			patterns := sanitizer.GetPatterns()
			assert.Len(t, patterns, 1)
			assert.Equal(t, tt.priority, patterns[0].Priority)
		})
	}
}

// TestSanitizer_SanitizePerformanceData tests sanitizePerformanceData
func TestSanitizer_SanitizePerformanceData(t *testing.T) {
	sanitizer := NewSanitizer()

	// Create performance data
	perf := NewPerformanceData()
	perf.RecordRender("comp-1", "Counter", 10*time.Millisecond)
	perf.RecordRender("comp-2", "Form", 20*time.Millisecond)

	// Create export data with performance
	data := &ExportData{
		Version:     "1.0",
		Timestamp:   time.Now(),
		Performance: perf,
	}

	// Sanitize the data
	result := sanitizer.Sanitize(data)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Performance)

	// Performance data should be preserved (copied)
	allPerf := result.Performance.GetAll()
	assert.Equal(t, 2, len(allPerf))
}

// TestSanitizer_SanitizePerformanceData_NilPerformance tests nil performance data
func TestSanitizer_SanitizePerformanceData_NilPerformance(t *testing.T) {
	sanitizer := NewSanitizer()

	// Create export data without performance
	data := &ExportData{
		Version:     "1.0",
		Timestamp:   time.Now(),
		Performance: nil,
	}

	// Sanitize the data
	result := sanitizer.Sanitize(data)

	assert.NotNil(t, result)
	assert.Nil(t, result.Performance)
}

// TestSanitizeValue_SliceTypes tests SanitizeValue with various slice types
func TestSanitizeValue_SliceTypes(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "string slice with sensitive data",
			input: []interface{}{"password123", "normal", "secret_key"},
		},
		{
			name:  "mixed type slice",
			input: []interface{}{42, "password", true, 3.14},
		},
		{
			name:  "empty slice",
			input: []interface{}{},
		},
		{
			name:  "nested slice",
			input: []interface{}{[]interface{}{"nested_password"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeValue(tt.input)
			assert.NotNil(t, result)
		})
	}
}

// TestSanitizeValue_MapTypes tests SanitizeValue with various map types
func TestSanitizeValue_MapTypes(t *testing.T) {
	sanitizer := NewSanitizer()

	// The sanitizer patterns expect format: password=value, token=value, etc.
	tests := []struct {
		name           string
		input          map[string]interface{}
		expectRedacted bool
	}{
		{
			name: "map with password pattern in value",
			input: map[string]interface{}{
				"config":   "password=secret123",
				"username": "alice",
			},
			expectRedacted: true,
		},
		{
			name: "map with nested sensitive data pattern",
			input: map[string]interface{}{
				"auth": map[string]interface{}{
					"credentials": "token=bearer_xyz",
				},
			},
			expectRedacted: true,
		},
		{
			name: "map with no sensitive data",
			input: map[string]interface{}{
				"name":  "alice",
				"email": "alice@example.com",
			},
			expectRedacted: false,
		},
		{
			name:           "empty map",
			input:          map[string]interface{}{},
			expectRedacted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeValue(tt.input)
			assert.NotNil(t, result)

			resultMap, ok := result.(map[string]interface{})
			assert.True(t, ok)

			if tt.expectRedacted {
				// Check that at least one value was redacted
				hasRedacted := false
				for _, v := range resultMap {
					if str, ok := v.(string); ok && strings.Contains(str, "[REDACTED]") {
						hasRedacted = true
						break
					}
					if nestedMap, ok := v.(map[string]interface{}); ok {
						for _, nv := range nestedMap {
							if str, ok := nv.(string); ok && strings.Contains(str, "[REDACTED]") {
								hasRedacted = true
								break
							}
						}
					}
				}
				assert.True(t, hasRedacted, "Expected at least one redacted value")
			}
		})
	}
}

// TestSanitizeValue_NilValue tests SanitizeValue with nil input
func TestSanitizeValue_NilValue(t *testing.T) {
	sanitizer := NewSanitizer()

	result := sanitizer.SanitizeValue(nil)
	assert.Nil(t, result)
}

// TestSanitizeValue_PrimitiveTypes tests SanitizeValue with primitive types
func TestSanitizeValue_PrimitiveTypes(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name   string
		input  interface{}
		expect interface{}
	}{
		{
			name:   "integer",
			input:  42,
			expect: 42,
		},
		{
			name:   "float",
			input:  3.14,
			expect: 3.14,
		},
		{
			name:   "bool",
			input:  true,
			expect: true,
		},
		{
			name:   "string without sensitive data",
			input:  "normal text",
			expect: "normal text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeValue(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// TestSanitizeValue_PointerAndInterface tests SanitizeValue with pointers and interfaces
func TestSanitizeValue_PointerAndInterface(t *testing.T) {
	sanitizer := NewSanitizer()

	// Test with a map pointer
	mapVal := map[string]interface{}{
		"password": "secret123",
	}

	result := sanitizer.SanitizeValue(mapVal)
	assert.NotNil(t, result)
}
