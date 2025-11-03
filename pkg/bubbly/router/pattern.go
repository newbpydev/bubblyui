package router

import (
	"fmt"
	"regexp"
	"strings"
)

// SegmentKind represents the type of a route segment
type SegmentKind int

const (
	// SegmentStatic represents a static path segment (e.g., "users")
	SegmentStatic SegmentKind = iota
	// SegmentParam represents a dynamic parameter (e.g., ":id")
	SegmentParam
	// SegmentOptional represents an optional parameter (e.g., ":id?")
	SegmentOptional
	// SegmentWildcard represents a wildcard parameter (e.g., ":path*")
	SegmentWildcard
)

// String returns the string representation of a SegmentKind
func (sk SegmentKind) String() string {
	switch sk {
	case SegmentStatic:
		return "static"
	case SegmentParam:
		return "param"
	case SegmentOptional:
		return "optional"
	case SegmentWildcard:
		return "wildcard"
	default:
		return "unknown"
	}
}

// Segment represents a single segment in a route pattern
type Segment struct {
	Kind  SegmentKind // Type of segment
	Name  string      // Parameter name (for dynamic segments)
	Value string      // Static value (for static segments)
}

// RoutePattern represents a compiled route pattern
type RoutePattern struct {
	segments []Segment
	regex    *regexp.Regexp
}

// CompilePattern compiles a path string into a RoutePattern
func CompilePattern(path string) (*RoutePattern, error) {
	// Validate path
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must start with /")
	}

	// Handle root path
	if path == "/" {
		pattern := &RoutePattern{
			segments: []Segment{},
		}
		pattern.regex = regexp.MustCompile("^/$")
		return pattern, nil
	}

	// Normalize path (remove trailing slash)
	path = strings.TrimSuffix(path, "/")

	// Parse segments
	parts := strings.Split(path, "/")[1:] // Skip empty first element
	segments, err := parseSegments(parts)
	if err != nil {
		return nil, err
	}

	// Generate regex
	pattern := &RoutePattern{
		segments: segments,
	}

	regexStr := generateRegex(segments)
	pattern.regex = regexp.MustCompile(regexStr)

	return pattern, nil
}

// Match attempts to match a path against this pattern
// Returns the extracted parameters and whether the match succeeded
func (rp *RoutePattern) Match(path string) (map[string]string, bool) {
	// Normalize path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	// Try regex match
	matches := rp.regex.FindStringSubmatch(path)
	if matches == nil {
		return nil, false
	}

	// Extract parameters
	params := make(map[string]string)
	matchIndex := 1 // Skip full match

	for _, seg := range rp.segments {
		switch seg.Kind {
		case SegmentParam:
			if matchIndex < len(matches) {
				params[seg.Name] = matches[matchIndex]
				matchIndex++
			}
		case SegmentOptional:
			if matchIndex < len(matches) && matches[matchIndex] != "" {
				params[seg.Name] = matches[matchIndex]
			}
			matchIndex++
		case SegmentWildcard:
			if matchIndex < len(matches) {
				params[seg.Name] = matches[matchIndex]
			}
			matchIndex++
		}
	}

	return params, true
}

// parseSegments parses path parts into segments
func parseSegments(parts []string) ([]Segment, error) {
	segments := make([]Segment, 0, len(parts))
	paramNames := make(map[string]bool)
	wildcardFound := false

	for i, part := range parts {
		if part == "" {
			continue
		}

		// Check if this is a parameter
		if strings.HasPrefix(part, ":") {
			seg, err := parseParamSegment(part, i, len(parts), &wildcardFound, paramNames)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
		} else {
			// Static segment
			segments = append(segments, Segment{
				Kind:  SegmentStatic,
				Value: part,
			})
		}
	}

	return segments, nil
}

// parseParamSegment parses a parameter segment (:id, :id?, :path*)
func parseParamSegment(part string, index, totalParts int, wildcardFound *bool, paramNames map[string]bool) (Segment, error) {
	if *wildcardFound {
		return Segment{}, fmt.Errorf("wildcard must be the last segment")
	}

	paramName := part[1:] // Remove ':'

	// Check for wildcard
	if strings.HasSuffix(paramName, "*") {
		return parseWildcardSegment(paramName, index, totalParts, wildcardFound, paramNames)
	}

	// Check for optional
	if strings.HasSuffix(paramName, "?") {
		return parseOptionalSegment(paramName, paramNames)
	}

	// Regular param
	return parseRegularParam(paramName, paramNames)
}

// parseWildcardSegment parses a wildcard segment (:path*)
func parseWildcardSegment(paramName string, index, totalParts int, wildcardFound *bool, paramNames map[string]bool) (Segment, error) {
	paramName = strings.TrimSuffix(paramName, "*")
	if err := validateParamName(paramName, paramNames); err != nil {
		return Segment{}, err
	}

	*wildcardFound = true

	// Wildcard must be last
	if index != totalParts-1 {
		return Segment{}, fmt.Errorf("wildcard must be the last segment")
	}

	return Segment{
		Kind: SegmentWildcard,
		Name: paramName,
	}, nil
}

// parseOptionalSegment parses an optional segment (:id?)
func parseOptionalSegment(paramName string, paramNames map[string]bool) (Segment, error) {
	paramName = strings.TrimSuffix(paramName, "?")
	if err := validateParamName(paramName, paramNames); err != nil {
		return Segment{}, err
	}

	return Segment{
		Kind: SegmentOptional,
		Name: paramName,
	}, nil
}

// parseRegularParam parses a regular parameter segment (:id)
func parseRegularParam(paramName string, paramNames map[string]bool) (Segment, error) {
	if err := validateParamName(paramName, paramNames); err != nil {
		return Segment{}, err
	}

	return Segment{
		Kind: SegmentParam,
		Name: paramName,
	}, nil
}

// validateParamName validates a parameter name and checks for duplicates
func validateParamName(paramName string, paramNames map[string]bool) error {
	if paramName == "" {
		return fmt.Errorf("parameter name cannot be empty")
	}
	if !isValidParamName(paramName) {
		return fmt.Errorf("invalid parameter name: %s", paramName)
	}
	if paramNames[paramName] {
		return fmt.Errorf("duplicate parameter name: %s", paramName)
	}
	paramNames[paramName] = true
	return nil
}

// generateRegex creates a regex pattern from segments
func generateRegex(segments []Segment) string {
	if len(segments) == 0 {
		return "^/$"
	}

	var parts []string
	parts = append(parts, "^")

	for _, seg := range segments {
		switch seg.Kind {
		case SegmentStatic:
			parts = append(parts, "/"+regexp.QuoteMeta(seg.Value))
		case SegmentParam:
			// Match any non-slash characters
			parts = append(parts, "/([^/]+)")
		case SegmentOptional:
			// Match optional segment
			parts = append(parts, "(?:/([^/]+))?")
		case SegmentWildcard:
			// Match everything remaining (including slashes)
			parts = append(parts, "(?:/(.*))?")
		}
	}

	parts = append(parts, "/?$") // Allow optional trailing slash

	return strings.Join(parts, "")
}

// isValidParamName checks if a parameter name is valid
// Valid names contain only alphanumeric characters and underscores
func isValidParamName(name string) bool {
	if name == "" {
		return false
	}

	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_') {
			return false
		}
	}

	return true
}
