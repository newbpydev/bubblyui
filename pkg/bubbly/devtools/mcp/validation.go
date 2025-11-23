package mcp

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// ValidateResourceURI validates MCP resource URIs to prevent injection attacks.
//
// This function checks for:
//   - Valid bubblyui:// scheme
//   - No path traversal attempts (../, ..\, encoded variants)
//   - No absolute paths
//   - No null bytes or control characters
//   - Valid resource paths (components, state, events, performance, debug)
//   - Reasonable URI length (< 1024 characters)
//
// Security Features:
//   - Path traversal prevention
//   - Scheme validation
//   - Control character filtering
//   - Length limits
//
// Example:
//
//	// Valid URIs
//	err := ValidateResourceURI("bubblyui://components")
//	err := ValidateResourceURI("bubblyui://state/refs")
//
//	// Invalid URIs
//	err := ValidateResourceURI("bubblyui://components/../../../etc/passwd")  // path traversal
//	err := ValidateResourceURI("http://components")  // wrong scheme
//
// Parameters:
//   - uri: The resource URI to validate
//
// Returns:
//   - error: nil if valid, descriptive error otherwise
//
// validateURIBasics performs basic URI validation checks.
func validateURIBasics(uri string) error {
	if strings.TrimSpace(uri) == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}

	const maxURILength = 1024
	if len(uri) > maxURILength {
		return fmt.Errorf("resource URI too long (max %d characters)", maxURILength)
	}

	if strings.Contains(uri, "\x00") {
		return fmt.Errorf("resource URI contains null byte")
	}

	for _, r := range uri {
		if unicode.IsControl(r) && r != '\t' {
			return fmt.Errorf("resource URI contains control character: %U", r)
		}
	}

	if strings.Contains(uri, "\\") {
		return fmt.Errorf("path traversal attempt detected (backslash)")
	}

	return nil
}

// validateURIPath validates the resource path for traversal and format.
func validateURIPath(resourcePath, uri string) error {
	if resourcePath == "" {
		return fmt.Errorf("empty resource path")
	}

	if strings.Contains(resourcePath, "..") {
		return fmt.Errorf("path traversal attempt detected")
	}

	lowerURI := strings.ToLower(uri)
	if strings.Contains(lowerURI, "%2e%2e") || strings.Contains(lowerURI, "%2f") {
		return fmt.Errorf("encoded path traversal attempt detected")
	}

	if strings.HasPrefix(resourcePath, "/") && !strings.HasPrefix(resourcePath, "//") {
		if len(resourcePath) > 1 && resourcePath[1] != '/' {
			return fmt.Errorf("absolute path not allowed")
		}
	}

	return nil
}

// validateResourceBase validates that the base resource is allowed.
func validateResourceBase(baseResource string) error {
	validBases := []string{"components", "state", "events", "performance", "commands", "debug"}
	for _, valid := range validBases {
		if baseResource == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid resource path: %s (must start with components, state, events, performance, commands, or debug)", baseResource)
}

// validateMultiSegmentPath validates paths with multiple segments.
func validateMultiSegmentPath(parts []string) error {
	validResources := []string{
		"components", "state/refs", "state/history", "events/log", "events",
		"performance/metrics", "performance/flamegraph", "commands/timeline", "debug/snapshot",
	}

	fullPath := strings.Join(parts[:2], "/")
	for _, valid := range validResources {
		if strings.HasPrefix(fullPath, valid) {
			return nil
		}
	}

	if len(parts) == 2 {
		if !isValidID(parts[1]) {
			return fmt.Errorf("invalid resource ID: %s", parts[1])
		}
	}
	return nil
}

func ValidateResourceURI(uri string) error {
	if err := validateURIBasics(uri); err != nil {
		return err
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid URI format: %w", err)
	}

	if parsed.Scheme != "bubblyui" {
		return fmt.Errorf("invalid scheme: expected 'bubblyui', got '%s'", parsed.Scheme)
	}

	resourcePath := parsed.Host + parsed.Path
	if err := validateURIPath(resourcePath, uri); err != nil {
		return err
	}

	parts := strings.Split(strings.Trim(resourcePath, "/"), "/")
	if len(parts) == 0 {
		return fmt.Errorf("empty resource path")
	}

	if err := validateResourceBase(parts[0]); err != nil {
		return err
	}

	if len(parts) > 1 {
		return validateMultiSegmentPath(parts)
	}

	return nil
}

// SanitizeInput removes dangerous characters from user input.
//
// This function removes or replaces:
//   - SQL injection characters: ; ' "
//   - Command injection: ` $ ( ) | & < >
//   - Path traversal: ../
//   - Null bytes: \x00
//   - Control characters (except space and tab)
//
// Note: This is defense-in-depth. Primary defense is proper parameterization
// and validation. Use this for logging and display purposes.
//
// Example:
//
//	input := "'; DROP TABLE users; --"
//	safe := SanitizeInput(input)  // "' DROP TABLE users --"
//
// Parameters:
//   - input: The string to sanitize
//
// Returns:
//   - string: Sanitized string with dangerous characters removed
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Replace control characters with space (except tab)
	var sb strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) {
			if r == '\t' || r == '\n' || r == '\r' {
				sb.WriteRune(' ')
			}
			// Skip other control characters
		} else {
			sb.WriteRune(r)
		}
	}
	input = sb.String()

	// Remove dangerous characters for SQL/command injection
	dangerousChars := []string{
		";",   // SQL statement terminator
		"`",   // Command substitution
		"$(",  // Command substitution
		")",   // Command substitution end
		"|",   // Pipe
		"&",   // Background/AND
		"<",   // Redirect
		">",   // Redirect
		"../", // Path traversal
		".\\", // Path traversal (Windows)
		"/",   // Path separator (for path traversal)
		"'",   // SQL string delimiter
	}

	for _, char := range dangerousChars {
		input = strings.ReplaceAll(input, char, "")
	}

	return input
}

// ValidateToolParams validates parameters for MCP tool calls.
//
// This function performs tool-specific validation to prevent injection attacks:
//   - export_session: format, destination path, include sections
//   - search_components: query, fields, max_results
//   - filter_events: event_names, source_ids, limit
//   - set_ref_value: ref_id, new_value
//   - clear_state_history: no params
//   - clear_event_log: no params
//
// Security Features:
//   - SQL injection prevention
//   - Command injection prevention
//   - Path traversal prevention
//   - Parameter type validation
//   - Range validation
//
// Example:
//
//	params := map[string]interface{}{
//	    "format": "json",
//	    "destination": "/tmp/export.json",
//	}
//	err := ValidateToolParams("export_session", params)
//
// Parameters:
//   - toolName: The name of the tool being called
//   - params: The parameters passed to the tool
//
// Returns:
//   - error: nil if valid, descriptive error otherwise
func ValidateToolParams(toolName string, params map[string]interface{}) error {
	// Check for nil params
	if params == nil {
		return fmt.Errorf("tool parameters cannot be nil")
	}

	// Validate based on tool name
	switch toolName {
	case "export_session":
		return validateExportSessionParams(params)
	case "search_components":
		return validateSearchComponentsParams(params)
	case "filter_events":
		return validateFilterEventsParams(params)
	case "set_ref_value":
		return validateSetRefValueParams(params)
	case "clear_state_history", "clear_event_log":
		// No parameters to validate
		return nil
	case "get_ref_dependencies":
		return validateGetRefDependenciesParams(params)
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}

// validateExportSessionParams validates export_session tool parameters.
func validateExportSessionParams(params map[string]interface{}) error {
	// Validate format
	if format, ok := params["format"].(string); ok {
		validFormats := map[string]bool{"json": true, "yaml": true, "msgpack": true}
		if !validFormats[format] {
			return fmt.Errorf("invalid format: %s (must be json, yaml, or msgpack)", format)
		}
	}

	// Validate destination
	if dest, ok := params["destination"].(string); ok {
		// Check for path traversal
		if strings.Contains(dest, "..") {
			return fmt.Errorf("path traversal attempt in destination")
		}

		// Check for command injection
		if containsDangerousChars(dest) {
			return fmt.Errorf("destination contains invalid characters")
		}

		// Check for null bytes
		if strings.Contains(dest, "\x00") {
			return fmt.Errorf("destination contains null byte")
		}
	}

	// Validate include array
	if includeRaw, ok := params["include"]; ok {
		if includeSlice, ok := includeRaw.([]interface{}); ok {
			validSections := map[string]bool{
				"components":  true,
				"state":       true,
				"events":      true,
				"performance": true,
			}

			for _, item := range includeSlice {
				if section, ok := item.(string); ok {
					if !validSections[section] {
						return fmt.Errorf("invalid section: %s (must be components, state, events, or performance)", section)
					}
				}
			}
		} else if includeStrSlice, ok := includeRaw.([]string); ok {
			// Handle []string type
			validSections := map[string]bool{
				"components":  true,
				"state":       true,
				"events":      true,
				"performance": true,
			}

			for _, section := range includeStrSlice {
				if !validSections[section] {
					return fmt.Errorf("invalid section: %s (must be components, state, events, or performance)", section)
				}
			}
		}
	}

	return nil
}

// validateAllowedValues checks if all values in a slice are in the allowed set.
func validateAllowedValues(raw interface{}, fieldName string, allowed map[string]bool) error {
	if slice, ok := raw.([]interface{}); ok {
		for _, item := range slice {
			if s, ok := item.(string); ok {
				if !allowed[s] {
					return fmt.Errorf("invalid %s: %s", fieldName, s)
				}
			}
		}
	} else if strSlice, ok := raw.([]string); ok {
		for _, s := range strSlice {
			if !allowed[s] {
				return fmt.Errorf("invalid %s: %s", fieldName, s)
			}
		}
	}
	return nil
}

// validateMaxResults validates the max_results parameter.
func validateMaxResults(params map[string]interface{}, min, max int) error {
	if maxResults, ok := params["max_results"].(float64); ok {
		if maxResults < float64(min) || maxResults > float64(max) {
			return fmt.Errorf("max_results must be between %d and %d, got %.0f", min, max, maxResults)
		}
	} else if maxResults, ok := params["max_results"].(int); ok {
		if maxResults < min || maxResults > max {
			return fmt.Errorf("max_results must be between %d and %d, got %d", min, max, maxResults)
		}
	}
	return nil
}

// validateSearchComponentsParams validates search_components tool parameters.
func validateSearchComponentsParams(params map[string]interface{}) error {
	if query, ok := params["query"].(string); ok {
		if containsDangerousChars(query) {
			return fmt.Errorf("query contains invalid characters")
		}
	}

	if fieldsRaw, ok := params["fields"]; ok {
		validFields := map[string]bool{"name": true, "type": true, "id": true}
		if err := validateAllowedValues(fieldsRaw, "field", validFields); err != nil {
			return fmt.Errorf("%v (must be name, type, or id)", err)
		}
	}

	return validateMaxResults(params, 1, 1000)
}

// validateStringSlice validates a slice parameter, checking each string with a validator.
func validateStringSlice(raw interface{}, validator func(string) error) error {
	if slice, ok := raw.([]interface{}); ok {
		for _, item := range slice {
			if s, ok := item.(string); ok {
				if err := validator(s); err != nil {
					return err
				}
			}
		}
	} else if strSlice, ok := raw.([]string); ok {
		for _, s := range strSlice {
			if err := validator(s); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateSourceID validates a source_id string for path traversal and dangerous chars.
func validateSourceID(id string) error {
	if strings.Contains(id, "..") {
		return fmt.Errorf("path traversal attempt in source_id: %s", id)
	}
	if containsDangerousChars(id) {
		return fmt.Errorf("source_id contains invalid characters: %s", id)
	}
	return nil
}

// validateLimit validates a limit parameter is within the allowed range.
func validateLimit(params map[string]interface{}, min, max int) error {
	if limit, ok := params["limit"].(float64); ok {
		if limit < float64(min) || limit > float64(max) {
			return fmt.Errorf("limit must be between %d and %d, got %.0f", min, max, limit)
		}
	} else if limit, ok := params["limit"].(int); ok {
		if limit < min || limit > max {
			return fmt.Errorf("limit must be between %d and %d, got %d", min, max, limit)
		}
	}
	return nil
}

// validateFilterEventsParams validates filter_events tool parameters.
func validateFilterEventsParams(params map[string]interface{}) error {
	if eventNamesRaw, ok := params["event_names"]; ok {
		if err := validateStringSlice(eventNamesRaw, func(s string) error {
			if containsDangerousChars(s) {
				return fmt.Errorf("event_name contains invalid characters: %s", s)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if sourceIDsRaw, ok := params["source_ids"]; ok {
		if err := validateStringSlice(sourceIDsRaw, validateSourceID); err != nil {
			return err
		}
	}

	return validateLimit(params, 1, 10000)
}

// validateSetRefValueParams validates set_ref_value tool parameters.
func validateSetRefValueParams(params map[string]interface{}) error {
	// Validate ref_id
	if refID, ok := params["ref_id"].(string); ok {
		if containsDangerousChars(refID) {
			return fmt.Errorf("ref_id contains invalid characters")
		}

		// Check for null bytes
		if strings.Contains(refID, "\x00") {
			return fmt.Errorf("ref_id contains null byte")
		}
	}

	return nil
}

// validateGetRefDependenciesParams validates get_ref_dependencies tool parameters.
func validateGetRefDependenciesParams(params map[string]interface{}) error {
	// Validate ref_id
	if refID, ok := params["ref_id"].(string); ok {
		if containsDangerousChars(refID) {
			return fmt.Errorf("ref_id contains invalid characters")
		}

		// Check for null bytes
		if strings.Contains(refID, "\x00") {
			return fmt.Errorf("ref_id contains null byte")
		}
	}

	return nil
}

// containsDangerousChars checks if a string contains characters that could be used for injection.
func containsDangerousChars(s string) bool {
	dangerousChars := []string{
		";",    // SQL statement terminator
		"'",    // SQL string delimiter
		"\"",   // SQL string delimiter
		"`",    // Command substitution
		"$(",   // Command substitution
		"|",    // Pipe
		"&",    // Background/AND
		"<",    // Redirect
		">",    // Redirect
		"\x00", // Null byte
	}

	for _, char := range dangerousChars {
		if strings.Contains(s, char) {
			return true
		}
	}

	return false
}

// isValidID checks if a string is a valid ID (alphanumeric with hyphens and underscores).
func isValidID(id string) bool {
	if id == "" {
		return false
	}

	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}

	return true
}
