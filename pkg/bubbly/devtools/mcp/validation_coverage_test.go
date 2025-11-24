package mcp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateMultiSegmentPath tests multi-segment path validation with additional edge cases.
func TestValidateMultiSegmentPath(t *testing.T) {
	tests := []struct {
		name      string
		parts     []string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid components path",
			parts:     []string{"components", "comp-123"},
			wantError: false,
		},
		{
			name:      "valid state/refs path",
			parts:     []string{"state", "refs"},
			wantError: false,
		},
		{
			name:      "valid state/history path",
			parts:     []string{"state", "history"},
			wantError: false,
		},
		{
			name:      "valid events/log path",
			parts:     []string{"events", "log"},
			wantError: false,
		},
		{
			name:      "valid performance/metrics path",
			parts:     []string{"performance", "metrics"},
			wantError: false,
		},
		{
			name:      "valid performance/flamegraph path",
			parts:     []string{"performance", "flamegraph"},
			wantError: false,
		},
		{
			name:      "valid commands/timeline path",
			parts:     []string{"commands", "timeline"},
			wantError: false,
		},
		{
			name:      "valid debug/snapshot path",
			parts:     []string{"debug", "snapshot"},
			wantError: false,
		},
		{
			name:      "valid component with ID",
			parts:     []string{"components", "my-component-123"},
			wantError: false,
		},
		{
			name:      "valid event with ID",
			parts:     []string{"events", "evt-456"},
			wantError: false,
		},
		{
			name:      "unknown base with invalid ID - special chars",
			parts:     []string{"unknown", "id;DROP TABLE"},
			wantError: true,
			errMsg:    "invalid resource ID",
		},
		{
			name:      "unknown base with invalid ID - spaces",
			parts:     []string{"unknown", "id with spaces"},
			wantError: true,
			errMsg:    "invalid resource ID",
		},
		{
			name:      "unknown base with empty ID",
			parts:     []string{"unknown", ""},
			wantError: true,
			errMsg:    "invalid resource ID",
		},
		{
			name:      "components prefix matches - returns nil before ID check",
			parts:     []string{"components", "any-id"},
			wantError: false, // components prefix match returns early
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMultiSegmentPath(tt.parts)

			if tt.wantError {
				assert.Error(t, err, "Expected error for parts: %v", tt.parts)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err, "Expected no error for parts: %v", tt.parts)
			}
		})
	}
}

// TestIsValidID tests the ID validation function.
func TestIsValidID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "valid alphanumeric ID",
			id:       "abc123",
			expected: true,
		},
		{
			name:     "valid ID with hyphens",
			id:       "comp-123-test",
			expected: true,
		},
		{
			name:     "valid ID with underscores",
			id:       "comp_123_test",
			expected: true,
		},
		{
			name:     "valid mixed ID",
			id:       "MyComponent-123_abc",
			expected: true,
		},
		{
			name:     "empty ID",
			id:       "",
			expected: false,
		},
		{
			name:     "ID with space",
			id:       "comp 123",
			expected: false,
		},
		{
			name:     "ID with special chars",
			id:       "comp;test",
			expected: false,
		},
		{
			name:     "ID with dot",
			id:       "comp.test",
			expected: false,
		},
		{
			name:     "ID with slash",
			id:       "comp/test",
			expected: false,
		},
		{
			name:     "ID with backtick",
			id:       "comp`test",
			expected: false,
		},
		{
			name:     "single character",
			id:       "a",
			expected: true,
		},
		{
			name:     "single hyphen",
			id:       "-",
			expected: true,
		},
		{
			name:     "single underscore",
			id:       "_",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidID(tt.id)
			assert.Equal(t, tt.expected, result, "isValidID(%q) should be %v", tt.id, tt.expected)
		})
	}
}

// TestValidateAllowedValues tests the allowed values validation function.
func TestValidateAllowedValues(t *testing.T) {
	tests := []struct {
		name      string
		raw       interface{}
		fieldName string
		allowed   map[string]bool
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid values with []interface{}",
			raw:       []interface{}{"json", "yaml"},
			fieldName: "format",
			allowed:   map[string]bool{"json": true, "yaml": true, "msgpack": true},
			wantError: false,
		},
		{
			name:      "invalid value with []interface{}",
			raw:       []interface{}{"json", "xml"},
			fieldName: "format",
			allowed:   map[string]bool{"json": true, "yaml": true, "msgpack": true},
			wantError: true,
			errMsg:    "invalid format: xml",
		},
		{
			name:      "valid values with []string",
			raw:       []string{"components", "state"},
			fieldName: "section",
			allowed:   map[string]bool{"components": true, "state": true, "events": true},
			wantError: false,
		},
		{
			name:      "invalid value with []string",
			raw:       []string{"components", "invalid"},
			fieldName: "section",
			allowed:   map[string]bool{"components": true, "state": true, "events": true},
			wantError: true,
			errMsg:    "invalid section: invalid",
		},
		{
			name:      "empty slice",
			raw:       []interface{}{},
			fieldName: "format",
			allowed:   map[string]bool{"json": true},
			wantError: false,
		},
		{
			name:      "nil value",
			raw:       nil,
			fieldName: "format",
			allowed:   map[string]bool{"json": true},
			wantError: false,
		},
		{
			name:      "mixed types in slice (non-string ignored)",
			raw:       []interface{}{"json", 123, "yaml"},
			fieldName: "format",
			allowed:   map[string]bool{"json": true, "yaml": true},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAllowedValues(tt.raw, tt.fieldName, tt.allowed)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateMaxResults tests max_results parameter validation.
func TestValidateMaxResults(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		min       int
		max       int
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid float64 within range",
			params:    map[string]interface{}{"max_results": float64(50)},
			min:       1,
			max:       100,
			wantError: false,
		},
		{
			name:      "float64 too low",
			params:    map[string]interface{}{"max_results": float64(0)},
			min:       1,
			max:       100,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "float64 too high",
			params:    map[string]interface{}{"max_results": float64(1001)},
			min:       1,
			max:       1000,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "valid int within range",
			params:    map[string]interface{}{"max_results": 50},
			min:       1,
			max:       100,
			wantError: false,
		},
		{
			name:      "int too low",
			params:    map[string]interface{}{"max_results": 0},
			min:       1,
			max:       100,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "int too high",
			params:    map[string]interface{}{"max_results": 1001},
			min:       1,
			max:       1000,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "no max_results param",
			params:    map[string]interface{}{},
			min:       1,
			max:       100,
			wantError: false,
		},
		{
			name:      "at minimum boundary",
			params:    map[string]interface{}{"max_results": float64(1)},
			min:       1,
			max:       100,
			wantError: false,
		},
		{
			name:      "at maximum boundary",
			params:    map[string]interface{}{"max_results": float64(100)},
			min:       1,
			max:       100,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMaxResults(tt.params, tt.min, tt.max)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateStringSlice tests string slice validation with custom validators.
func TestValidateStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		raw       interface{}
		validator func(string) error
		wantError bool
		errMsg    string
	}{
		{
			name: "valid []interface{} slice",
			raw:  []interface{}{"valid1", "valid2"},
			validator: func(s string) error {
				return nil
			},
			wantError: false,
		},
		{
			name: "invalid []interface{} slice",
			raw:  []interface{}{"valid", "invalid;sql"},
			validator: func(s string) error {
				if strings.Contains(s, ";") {
					return assert.AnError
				}
				return nil
			},
			wantError: true,
		},
		{
			name: "valid []string slice",
			raw:  []string{"valid1", "valid2"},
			validator: func(s string) error {
				return nil
			},
			wantError: false,
		},
		{
			name: "invalid []string slice",
			raw:  []string{"valid", "invalid|pipe"},
			validator: func(s string) error {
				if strings.Contains(s, "|") {
					return assert.AnError
				}
				return nil
			},
			wantError: true,
		},
		{
			name: "nil slice",
			raw:  nil,
			validator: func(s string) error {
				return assert.AnError
			},
			wantError: false, // nil is handled gracefully
		},
		{
			name: "mixed types in slice (non-string skipped)",
			raw:  []interface{}{"valid", 123, "also-valid"},
			validator: func(s string) error {
				return nil
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStringSlice(tt.raw, tt.validator)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateSourceID tests source ID validation.
func TestValidateSourceID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid source ID",
			id:        "comp-123",
			wantError: false,
		},
		{
			name:      "path traversal attempt",
			id:        "../../../etc/passwd",
			wantError: true,
			errMsg:    "path traversal",
		},
		{
			name:      "SQL injection attempt",
			id:        "comp';DROP TABLE",
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "command injection attempt",
			id:        "comp`whoami`",
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "clean alphanumeric ID",
			id:        "component123",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSourceID(tt.id)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateLimit tests limit parameter validation.
func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		min       int
		max       int
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid float64 within range",
			params:    map[string]interface{}{"limit": float64(500)},
			min:       1,
			max:       10000,
			wantError: false,
		},
		{
			name:      "float64 too low",
			params:    map[string]interface{}{"limit": float64(0)},
			min:       1,
			max:       10000,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "float64 too high",
			params:    map[string]interface{}{"limit": float64(10001)},
			min:       1,
			max:       10000,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "valid int within range",
			params:    map[string]interface{}{"limit": 500},
			min:       1,
			max:       10000,
			wantError: false,
		},
		{
			name:      "int too low",
			params:    map[string]interface{}{"limit": 0},
			min:       1,
			max:       10000,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "int too high",
			params:    map[string]interface{}{"limit": 10001},
			min:       1,
			max:       10000,
			wantError: true,
			errMsg:    "must be between",
		},
		{
			name:      "no limit param",
			params:    map[string]interface{}{},
			min:       1,
			max:       10000,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLimit(tt.params, tt.min, tt.max)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateGetRefDependenciesParams tests get_ref_dependencies parameter validation.
func TestValidateGetRefDependenciesParams(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid ref_id",
			params:    map[string]interface{}{"ref_id": "ref-123"},
			wantError: false,
		},
		{
			name:      "ref_id with dangerous chars",
			params:    map[string]interface{}{"ref_id": "ref';DROP TABLE"},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "ref_id with null byte",
			params:    map[string]interface{}{"ref_id": "ref\x00test"},
			wantError: true,
			errMsg:    "invalid characters", // null byte is detected as invalid character
		},
		{
			name:      "ref_id with backtick",
			params:    map[string]interface{}{"ref_id": "ref`cmd`"},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{},
			wantError: false,
		},
		{
			name:      "ref_id not a string",
			params:    map[string]interface{}{"ref_id": 123},
			wantError: false, // Only validates if it's a string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGetRefDependenciesParams(tt.params)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateToolParams_GetRefDependencies tests ValidateToolParams with get_ref_dependencies tool.
func TestValidateToolParams_GetRefDependencies(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid get_ref_dependencies params",
			params:    map[string]interface{}{"ref_id": "ref-123"},
			wantError: false,
		},
		{
			name:      "get_ref_dependencies with dangerous ref_id",
			params:    map[string]interface{}{"ref_id": "ref';--"},
			wantError: true,
			errMsg:    "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolParams("get_ref_dependencies", tt.params)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateURIPath_EdgeCases tests URI path validation edge cases.
func TestValidateURIPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		resourcePath string
		uri          string
		wantError    bool
		errMsg       string
	}{
		{
			name:         "empty resource path",
			resourcePath: "",
			uri:          "bubblyui://",
			wantError:    true,
			errMsg:       "empty resource path",
		},
		{
			name:         "path traversal with ..",
			resourcePath: "components/../../../etc",
			uri:          "bubblyui://components/../../../etc",
			wantError:    true,
			errMsg:       "path traversal",
		},
		{
			name:         "encoded path traversal %2e%2e",
			resourcePath: "components",
			uri:          "bubblyui://components/%2e%2e/secret",
			wantError:    true,
			errMsg:       "encoded path traversal",
		},
		{
			name:         "encoded slash %2f",
			resourcePath: "components",
			uri:          "bubblyui://components%2fetc%2fpasswd",
			wantError:    true,
			errMsg:       "encoded path traversal",
		},
		{
			name:         "absolute path starting with /",
			resourcePath: "/etc/passwd",
			uri:          "bubblyui:///etc/passwd",
			wantError:    true,
			errMsg:       "absolute path",
		},
		{
			name:         "valid path",
			resourcePath: "components/comp-123",
			uri:          "bubblyui://components/comp-123",
			wantError:    false,
		},
		{
			name:         "double slash is allowed",
			resourcePath: "//components",
			uri:          "bubblyui:////components",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURIPath(tt.resourcePath, tt.uri)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateExportSessionParams_IncludeSliceTypes tests both []interface{} and []string for include.
func TestValidateExportSessionParams_IncludeSliceTypes(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "include as []interface{} valid",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "/tmp/test.json",
				"include":     []interface{}{"components", "state"},
			},
			wantError: false,
		},
		{
			name: "include as []interface{} invalid section",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "/tmp/test.json",
				"include":     []interface{}{"components", "invalid_section"},
			},
			wantError: true,
			errMsg:    "invalid section",
		},
		{
			name: "include as []string valid",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "/tmp/test.json",
				"include":     []string{"components", "events"},
			},
			wantError: false,
		},
		{
			name: "include as []string invalid section",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "/tmp/test.json",
				"include":     []string{"components", "bad_section"},
			},
			wantError: true,
			errMsg:    "invalid section",
		},
		{
			name: "destination with null byte",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "/tmp/test\x00.json",
			},
			wantError: true,
			errMsg:    "invalid characters", // null byte is detected as invalid character
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolParams("export_session", tt.params)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContainsDangerousChars tests the dangerous character detection.
func TestContainsDangerousChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "clean string", input: "hello world", expected: false},
		{name: "semicolon", input: "test;drop", expected: true},
		{name: "single quote", input: "test'or", expected: true},
		{name: "double quote", input: "test\"or", expected: true},
		{name: "backtick", input: "test`cmd`", expected: true},
		{name: "dollar paren", input: "test$(cmd)", expected: true},
		{name: "pipe", input: "test|cmd", expected: true},
		{name: "ampersand", input: "test&cmd", expected: true},
		{name: "less than", input: "test<cmd", expected: true},
		{name: "greater than", input: "test>cmd", expected: true},
		{name: "null byte", input: "test\x00cmd", expected: true},
		{name: "empty string", input: "", expected: false},
		{name: "unicode safe", input: "hello world", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDangerousChars(tt.input)
			assert.Equal(t, tt.expected, result, "containsDangerousChars(%q)", tt.input)
		})
	}
}

// TestValidateSearchComponentsParams_FieldsSliceTypes tests both slice types for fields.
func TestValidateSearchComponentsParams_FieldsSliceTypes(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "fields as []interface{} valid",
			params: map[string]interface{}{
				"query":  "test",
				"fields": []interface{}{"name", "type"},
			},
			wantError: false,
		},
		{
			name: "fields as []interface{} invalid",
			params: map[string]interface{}{
				"query":  "test",
				"fields": []interface{}{"name", "invalid_field"},
			},
			wantError: true,
			errMsg:    "invalid field",
		},
		{
			name: "fields as []string valid",
			params: map[string]interface{}{
				"query":  "test",
				"fields": []string{"name", "id"},
			},
			wantError: false,
		},
		{
			name: "fields as []string invalid",
			params: map[string]interface{}{
				"query":  "test",
				"fields": []string{"name", "bad_field"},
			},
			wantError: true,
			errMsg:    "invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolParams("search_components", tt.params)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateFilterEventsParams_SliceTypes tests both slice types for event params.
func TestValidateFilterEventsParams_SliceTypes(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "event_names as []interface{} valid",
			params: map[string]interface{}{
				"event_names": []interface{}{"click", "submit"},
			},
			wantError: false,
		},
		{
			name: "event_names as []interface{} with injection",
			params: map[string]interface{}{
				"event_names": []interface{}{"click", "submit';DROP"},
			},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name: "event_names as []string valid",
			params: map[string]interface{}{
				"event_names": []string{"click", "submit"},
			},
			wantError: false,
		},
		{
			name: "source_ids as []interface{} valid",
			params: map[string]interface{}{
				"source_ids": []interface{}{"comp-1", "comp-2"},
			},
			wantError: false,
		},
		{
			name: "source_ids as []string valid",
			params: map[string]interface{}{
				"source_ids": []string{"comp-1", "comp-2"},
			},
			wantError: false,
		},
		{
			name: "source_ids with path traversal",
			params: map[string]interface{}{
				"source_ids": []string{"comp-1", "../../../etc"},
			},
			wantError: true,
			errMsg:    "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolParams("filter_events", tt.params)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
