package devtools

import (
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
