package mcp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateResourceURI tests resource URI validation.
func TestValidateResourceURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr bool
		errMsg  string
	}{
		// Valid URIs
		{
			name:    "valid components URI",
			uri:     "bubblyui://components",
			wantErr: false,
		},
		{
			name:    "valid component by ID",
			uri:     "bubblyui://components/comp-123",
			wantErr: false,
		},
		{
			name:    "valid state refs",
			uri:     "bubblyui://state/refs",
			wantErr: false,
		},
		{
			name:    "valid state history",
			uri:     "bubblyui://state/history",
			wantErr: false,
		},
		{
			name:    "valid events log",
			uri:     "bubblyui://events/log",
			wantErr: false,
		},
		{
			name:    "valid event by ID",
			uri:     "bubblyui://events/event-456",
			wantErr: false,
		},
		{
			name:    "valid performance metrics",
			uri:     "bubblyui://performance/metrics",
			wantErr: false,
		},
		{
			name:    "valid debug snapshot",
			uri:     "bubblyui://debug/snapshot",
			wantErr: false,
		},

		// Path traversal attempts
		{
			name:    "path traversal with ../",
			uri:     "bubblyui://components/../../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "path traversal with encoded ../",
			uri:     "bubblyui://components/%2e%2e%2f%2e%2e%2f",
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "path traversal with backslash",
			uri:     "bubblyui://components\\..\\..\\",
			wantErr: true,
			errMsg:  "backslash",
		},
		{
			name:    "absolute path attempt",
			uri:     "bubblyui:///etc/passwd",
			wantErr: true,
			errMsg:  "absolute path",
		},

		// Invalid scheme
		{
			name:    "wrong scheme http",
			uri:     "http://components",
			wantErr: true,
			errMsg:  "invalid scheme",
		},
		{
			name:    "wrong scheme file",
			uri:     "file:///tmp/data",
			wantErr: true,
			errMsg:  "invalid scheme",
		},
		{
			name:    "no scheme",
			uri:     "components",
			wantErr: true,
			errMsg:  "invalid scheme",
		},

		// Invalid resource paths
		{
			name:    "invalid resource path",
			uri:     "bubblyui://invalid/path",
			wantErr: true,
			errMsg:  "invalid resource",
		},
		{
			name:    "empty resource",
			uri:     "bubblyui://",
			wantErr: true,
			errMsg:  "empty resource",
		},

		// Null bytes and control characters
		{
			name:    "null byte in URI",
			uri:     "bubblyui://components\x00/test",
			wantErr: true,
			errMsg:  "null byte",
		},
		{
			name:    "control character in URI",
			uri:     "bubblyui://components\n/test",
			wantErr: true,
			errMsg:  "control character",
		},

		// Length limits
		{
			name:    "URI too long",
			uri:     "bubblyui://components/" + strings.Repeat("a", 2000),
			wantErr: true,
			errMsg:  "URI too long",
		},

		// Empty/whitespace
		{
			name:    "empty URI",
			uri:     "",
			wantErr: true,
			errMsg:  "empty",
		},
		{
			name:    "whitespace only",
			uri:     "   ",
			wantErr: true,
			errMsg:  "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceURI(tt.uri)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for URI: %s", tt.uri)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg),
						"Error message should contain: %s", tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no error for URI: %s", tt.uri)
			}
		})
	}
}

// TestSanitizeInput tests input sanitization.
func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean input",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "SQL injection attempt",
			input:    "'; DROP TABLE users; --",
			expected: " DROP TABLE users --",
		},
		{
			name:     "command injection with backticks",
			input:    "test`rm -rf /`",
			expected: "testrm -rf ",
		},
		{
			name:     "command injection with $(",
			input:    "test$(whoami)",
			expected: "testwhoami",
		},
		{
			name:     "null bytes",
			input:    "test\x00data",
			expected: "testdata",
		},
		{
			name:     "control characters",
			input:    "test\n\r\tdata",
			expected: "test   data",
		},
		{
			name:     "path traversal",
			input:    "../../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "HTML/script tags",
			input:    "<script>alert('xss')</script>",
			expected: "scriptalert(xssscript",
		},
		{
			name:     "multiple dangerous chars",
			input:    "test';`$()|&<>\x00\n",
			expected: "test ",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    "hello 世界",
			expected: "hello 世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result, "Sanitization mismatch for input: %s", tt.input)
		})
	}
}

// TestValidateToolParams tests tool parameter validation.
func TestValidateToolParams(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		params   map[string]interface{}
		wantErr  bool
		errMsg   string
	}{
		// export_session tool
		{
			name:     "export_session valid params",
			toolName: "export_session",
			params: map[string]interface{}{
				"format":      "json",
				"compress":    true,
				"sanitize":    true,
				"include":     []string{"components", "state"},
				"destination": "/tmp/export.json",
			},
			wantErr: false,
		},
		{
			name:     "export_session invalid format",
			toolName: "export_session",
			params: map[string]interface{}{
				"format":      "xml",
				"destination": "/tmp/export.xml",
			},
			wantErr: true,
			errMsg:  "invalid format",
		},
		{
			name:     "export_session path traversal in destination",
			toolName: "export_session",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "../../../etc/passwd",
			},
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:     "export_session command injection in destination",
			toolName: "export_session",
			params: map[string]interface{}{
				"format":      "json",
				"destination": "/tmp/export.json; rm -rf /",
			},
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:     "export_session invalid include section",
			toolName: "export_session",
			params: map[string]interface{}{
				"format":      "json",
				"include":     []string{"components", "invalid_section"},
				"destination": "/tmp/export.json",
			},
			wantErr: true,
			errMsg:  "invalid section",
		},

		// search_components tool
		{
			name:     "search_components valid params",
			toolName: "search_components",
			params: map[string]interface{}{
				"query":       "counter",
				"fields":      []string{"name", "type"},
				"max_results": 50,
			},
			wantErr: false,
		},
		{
			name:     "search_components SQL injection in query",
			toolName: "search_components",
			params: map[string]interface{}{
				"query": "'; DROP TABLE components; --",
			},
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:     "search_components invalid field",
			toolName: "search_components",
			params: map[string]interface{}{
				"query":  "test",
				"fields": []string{"name", "invalid_field"},
			},
			wantErr: true,
			errMsg:  "invalid field",
		},
		{
			name:     "search_components max_results too large",
			toolName: "search_components",
			params: map[string]interface{}{
				"query":       "test",
				"max_results": 10000,
			},
			wantErr: true,
			errMsg:  "max_results",
		},

		// filter_events tool
		{
			name:     "filter_events valid params",
			toolName: "filter_events",
			params: map[string]interface{}{
				"event_names": []string{"click", "submit"},
				"source_ids":  []string{"comp-1", "comp-2"},
				"limit":       100,
			},
			wantErr: false,
		},
		{
			name:     "filter_events SQL injection in event_names",
			toolName: "filter_events",
			params: map[string]interface{}{
				"event_names": []string{"click'; DROP TABLE events; --"},
			},
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:     "filter_events path traversal in source_ids",
			toolName: "filter_events",
			params: map[string]interface{}{
				"source_ids": []string{"../../../etc/passwd"},
			},
			wantErr: true,
			errMsg:  "path traversal",
		},

		// set_ref_value tool
		{
			name:     "set_ref_value valid params",
			toolName: "set_ref_value",
			params: map[string]interface{}{
				"ref_id":    "ref-counter-123",
				"new_value": 42,
				"dry_run":   false,
			},
			wantErr: false,
		},
		{
			name:     "set_ref_value SQL injection in ref_id",
			toolName: "set_ref_value",
			params: map[string]interface{}{
				"ref_id":    "ref-123'; DROP TABLE refs; --",
				"new_value": 42,
			},
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:     "set_ref_value command injection in ref_id",
			toolName: "set_ref_value",
			params: map[string]interface{}{
				"ref_id":    "ref-123`whoami`",
				"new_value": 42,
			},
			wantErr: true,
			errMsg:  "invalid character",
		},

		// clear_state_history tool
		{
			name:     "clear_state_history valid params",
			toolName: "clear_state_history",
			params:   map[string]interface{}{},
			wantErr:  false,
		},

		// Unknown tool
		{
			name:     "unknown tool",
			toolName: "unknown_tool",
			params:   map[string]interface{}{},
			wantErr:  true,
			errMsg:   "unknown tool",
		},

		// Nil params
		{
			name:     "nil params",
			toolName: "export_session",
			params:   nil,
			wantErr:  true,
			errMsg:   "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolParams(tt.toolName, tt.params)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for tool: %s", tt.toolName)
				if tt.errMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg),
						"Error message should contain: %s", tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no error for tool: %s", tt.toolName)
			}
		})
	}
}

// TestValidateToolParams_Concurrent tests thread safety of validation.
func TestValidateToolParams_Concurrent(t *testing.T) {
	params := map[string]interface{}{
		"format":      "json",
		"destination": "/tmp/export.json",
	}

	// Run 100 concurrent validations
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			err := ValidateToolParams("export_session", params)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestValidateResourceURI_Concurrent tests thread safety of URI validation.
func TestValidateResourceURI_Concurrent(t *testing.T) {
	uri := "bubblyui://components"

	// Run 100 concurrent validations
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			err := ValidateResourceURI(uri)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
