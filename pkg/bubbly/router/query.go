package router

import (
	"net/url"
	"strings"
)

// QueryParser handles parsing and building URL query strings.
//
// The parser uses Go's standard library net/url package for robust
// URL encoding/decoding, ensuring proper handling of special characters,
// spaces, and reserved characters according to RFC 3986.
//
// Features:
//   - Parse query strings into map[string]string
//   - Build query strings from map[string]string
//   - Automatic URL encoding/decoding
//   - Handle edge cases (empty values, no values, special characters)
//   - Round-trip consistency (parse → build → parse yields same result)
//
// Usage:
//
//	parser := router.NewQueryParser()
//
//	// Parse query string
//	params := parser.Parse("name=John+Doe&email=test%40example.com")
//	// params: {"name": "John Doe", "email": "test@example.com"}
//
//	// Build query string
//	query := parser.Build(map[string]string{
//		"name":  "John Doe",
//		"email": "test@example.com",
//	})
//	// query: "email=test%40example.com&name=John+Doe"
//
// Note on Duplicate Keys:
// When parsing query strings with duplicate keys (e.g., "key=1&key=2"),
// only the last value is retained. This is a simplification suitable for
// most routing use cases. If you need to preserve all values, use
// url.ParseQuery directly which returns url.Values (map[string][]string).
type QueryParser struct{}

// NewQueryParser creates a new query string parser.
//
// The parser is stateless and can be safely reused across multiple
// parse and build operations.
//
// Returns:
//   - *QueryParser: A new parser instance
//
// Example:
//
//	parser := router.NewQueryParser()
//	params := parser.Parse("key=value")
func NewQueryParser() *QueryParser {
	return &QueryParser{}
}

// Parse parses a URL query string into a map of key-value pairs.
//
// The method handles URL decoding automatically, converting encoded
// characters back to their original form. It supports both standard
// query format (key=value&key2=value2) and handles edge cases gracefully.
//
// Parameters:
//   - queryString: The query string to parse (with or without leading "?")
//
// Returns:
//   - map[string]string: Map of decoded query parameters
//
// Behavior:
//   - Leading "?" is automatically stripped if present
//   - Empty query strings return empty map
//   - Keys without values are treated as empty strings
//   - URL decoding is applied (e.g., "%20" → " ", "+" → " ")
//   - Duplicate keys: last value wins
//   - Malformed pairs are silently skipped
//
// Examples:
//
//	parser.Parse("key=value")
//	// → {"key": "value"}
//
//	parser.Parse("a=1&b=2&c=3")
//	// → {"a": "1", "b": "2", "c": "3"}
//
//	parser.Parse("name=John+Doe&email=test%40example.com")
//	// → {"name": "John Doe", "email": "test@example.com"}
//
//	parser.Parse("key=")
//	// → {"key": ""}
//
//	parser.Parse("key")
//	// → {"key": ""}
//
//	parser.Parse("")
//	// → {}
func (qp *QueryParser) Parse(queryString string) map[string]string {
	// Strip leading "?" if present
	queryString = strings.TrimPrefix(queryString, "?")

	// Handle empty query string
	if queryString == "" {
		return make(map[string]string)
	}

	// Parse using Go's standard library
	// This handles URL decoding, special characters, etc.
	values, err := url.ParseQuery(queryString)
	if err != nil {
		// If parsing fails, return empty map
		// This handles malformed query strings gracefully
		return make(map[string]string)
	}

	// Convert url.Values (map[string][]string) to map[string]string
	// For duplicate keys, we keep the last value
	result := make(map[string]string, len(values))
	for key, vals := range values {
		if len(vals) > 0 {
			// Take the last value for duplicate keys
			result[key] = vals[len(vals)-1]
		} else {
			// Key with no value
			result[key] = ""
		}
	}

	return result
}

// Build constructs a URL query string from a map of key-value pairs.
//
// The method handles URL encoding automatically, converting special
// characters to their percent-encoded form. Keys are sorted alphabetically
// for consistent output.
//
// Parameters:
//   - params: Map of query parameters to encode
//
// Returns:
//   - string: URL-encoded query string (without leading "?")
//
// Behavior:
//   - Empty map returns empty string
//   - Keys are sorted alphabetically for deterministic output
//   - Values are URL-encoded (e.g., " " → "+", "@" → "%40")
//   - Empty values are preserved as "key="
//   - Output does NOT include leading "?"
//
// Examples:
//
//	parser.Build(map[string]string{"key": "value"})
//	// → "key=value"
//
//	parser.Build(map[string]string{"a": "1", "b": "2", "c": "3"})
//	// → "a=1&b=2&c=3" (sorted alphabetically)
//
//	parser.Build(map[string]string{
//		"name":  "John Doe",
//		"email": "test@example.com",
//	})
//	// → "email=test%40example.com&name=John+Doe"
//
//	parser.Build(map[string]string{"key": ""})
//	// → "key="
//
//	parser.Build(map[string]string{})
//	// → ""
func (qp *QueryParser) Build(params map[string]string) string {
	// Handle empty params
	if len(params) == 0 {
		return ""
	}

	// Convert to url.Values for encoding
	values := make(url.Values, len(params))
	for key, value := range params {
		values.Set(key, value)
	}

	// Encode using Go's standard library
	// This handles URL encoding, sorting, etc.
	return values.Encode()
}
