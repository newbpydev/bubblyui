package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryParser_Parse tests query string parsing
func TestQueryParser_Parse(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name     string
		query    string
		expected map[string]string
	}{
		{
			name:  "simple query",
			query: "key=value",
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name:  "multiple params",
			query: "a=1&b=2&c=3",
			expected: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
		},
		{
			name:  "empty value",
			query: "key=",
			expected: map[string]string{
				"key": "",
			},
		},
		{
			name:  "no value",
			query: "key",
			expected: map[string]string{
				"key": "",
			},
		},
		{
			name:     "empty query",
			query:    "",
			expected: map[string]string{},
		},
		{
			name:  "url encoded characters",
			query: "name=John+Doe&email=test%40example.com",
			expected: map[string]string{
				"name":  "John Doe",
				"email": "test@example.com",
			},
		},
		{
			name:  "special characters",
			query: "path=%2Fhome%2Fuser&query=hello%20world",
			expected: map[string]string{
				"path":  "/home/user",
				"query": "hello world",
			},
		},
		{
			name:  "with question mark prefix",
			query: "?key=value&foo=bar",
			expected: map[string]string{
				"key": "value",
				"foo": "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestQueryParser_Build tests query string building
func TestQueryParser_Build(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name     string
		params   map[string]string
		expected string
	}{
		{
			name: "simple params",
			params: map[string]string{
				"key": "value",
			},
			expected: "key=value",
		},
		{
			name: "multiple params",
			params: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
			// Note: Go's url.Values.Encode() sorts keys alphabetically
			expected: "a=1&b=2&c=3",
		},
		{
			name: "empty value",
			params: map[string]string{
				"key": "",
			},
			expected: "key=",
		},
		{
			name:     "empty params",
			params:   map[string]string{},
			expected: "",
		},
		{
			name: "special characters",
			params: map[string]string{
				"name":  "John Doe",
				"email": "test@example.com",
			},
			expected: "email=test%40example.com&name=John+Doe",
		},
		{
			name: "path with slashes",
			params: map[string]string{
				"path":  "/home/user",
				"query": "hello world",
			},
			expected: "path=%2Fhome%2Fuser&query=hello+world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Build(tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestQueryParser_RoundTrip tests parse-build consistency
func TestQueryParser_RoundTrip(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name   string
		params map[string]string
	}{
		{
			name: "simple params",
			params: map[string]string{
				"key": "value",
			},
		},
		{
			name: "multiple params",
			params: map[string]string{
				"a": "1",
				"b": "2",
				"c": "3",
			},
		},
		{
			name: "special characters",
			params: map[string]string{
				"name":  "John Doe",
				"email": "test@example.com",
				"path":  "/home/user",
			},
		},
		{
			name: "empty value",
			params: map[string]string{
				"key": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build query string from params
			queryString := parser.Build(tt.params)

			// Parse it back
			parsed := parser.Parse(queryString)

			// Should match original params
			assert.Equal(t, tt.params, parsed)
		})
	}
}

// TestQueryParser_EdgeCases tests edge cases
func TestQueryParser_EdgeCases(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name     string
		query    string
		expected map[string]string
	}{
		{
			name:  "multiple ampersands",
			query: "a=1&&&&b=2",
			expected: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
		{
			name:  "trailing ampersand",
			query: "a=1&b=2&",
			expected: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
		{
			name:  "leading ampersand",
			query: "&a=1&b=2",
			expected: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
		{
			name:  "equals in value",
			query: "key=value=with=equals",
			expected: map[string]string{
				"key": "value=with=equals",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestQueryParser_DuplicateKeys tests handling of duplicate keys
func TestQueryParser_DuplicateKeys(t *testing.T) {
	parser := NewQueryParser()

	// When there are duplicate keys, only the last value is kept
	// This is the expected behavior for our simple string map
	query := "key=first&key=second&key=third"
	result := parser.Parse(query)

	// Should keep the last value
	assert.Equal(t, map[string]string{
		"key": "third",
	}, result)
}
