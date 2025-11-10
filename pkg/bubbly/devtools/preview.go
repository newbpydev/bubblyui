package devtools

import (
	"fmt"
	"reflect"
)

// DryRunResult contains the results of a dry-run sanitization.
//
// This structure is returned when sanitizing with DryRun: true in
// SanitizeOptions. It shows what would be redacted without actually
// modifying the data, allowing developers to validate patterns before
// applying them to production data.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	result := sanitizer.Preview(exportData)
//	fmt.Printf("Would redact %d values\n", result.WouldRedactCount)
//	for _, match := range result.Matches {
//	    fmt.Printf("  %s: %s → %s\n", match.Path, match.Original, match.Redacted)
//	}
type DryRunResult struct {
	// Matches is the list of all locations where patterns matched
	Matches []MatchLocation

	// WouldRedactCount is the total number of values that would be redacted
	WouldRedactCount int

	// PreviewData is the original data structure (unchanged)
	PreviewData interface{}
}

// MatchLocation describes a single match found during dry-run sanitization.
//
// It includes the path to the matched value, the pattern that matched,
// the original value, and what the redacted value would be. This allows
// developers to review exactly what would be changed before applying
// sanitization.
//
// Example:
//
//	match := MatchLocation{
//	    Path:     "components[0].props.password",
//	    Pattern:  "password",
//	    Original: "secret123",
//	    Redacted: "[REDACTED]",
//	}
type MatchLocation struct {
	// Path is the location of the match in the data structure
	// Format: "components[0].props.password" or "state[1].new_value"
	Path string

	// Pattern is the name of the pattern that matched
	Pattern string

	// Original is the original value before redaction (may be truncated)
	Original string

	// Redacted is what the value would become after redaction
	Redacted string

	// Line is the line number in JSON output (0 if not applicable)
	Line int

	// Column is the column number in JSON output (0 if not applicable)
	Column int
}

// SanitizeOptions configures sanitization behavior.
//
// Use DryRun: true to preview matches without modifying data.
// Use MaxPreviewLen to truncate long values in preview output.
//
// Example:
//
//	opts := SanitizeOptions{
//	    DryRun:        true,
//	    MaxPreviewLen: 100,
//	}
//	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)
type SanitizeOptions struct {
	// DryRun enables preview mode - matches are collected but data is not modified
	DryRun bool

	// MaxPreviewLen truncates original values longer than this in match locations.
	// Set to 0 for no truncation. Default is 100 characters.
	MaxPreviewLen int
}

// SanitizeWithOptions sanitizes data with configurable options.
//
// When DryRun is true, this method collects matches without modifying the
// original data and returns a DryRunResult. When DryRun is false, it
// performs normal sanitization and returns the sanitized data.
//
// Thread Safety:
//
//	Safe to call concurrently. Does not modify input data in dry-run mode.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	opts := SanitizeOptions{
//	    DryRun:        true,
//	    MaxPreviewLen: 100,
//	}
//	_, dryRunResult := sanitizer.SanitizeWithOptions(exportData, opts)
//	fmt.Printf("Would redact %d values\n", dryRunResult.WouldRedactCount)
//
// Parameters:
//   - data: The export data to sanitize or preview
//   - opts: Options controlling sanitization behavior
//
// Returns:
//   - *ExportData: Sanitized data (nil if DryRun is true)
//   - *DryRunResult: Dry-run results (nil if DryRun is false)
func (s *Sanitizer) SanitizeWithOptions(data *ExportData, opts SanitizeOptions) (*ExportData, *DryRunResult) {
	if data == nil {
		return nil, nil
	}

	// Set default MaxPreviewLen if not specified
	if opts.MaxPreviewLen == 0 {
		opts.MaxPreviewLen = 100
	}

	if opts.DryRun {
		// Dry-run mode: collect matches without modifying data
		result := &DryRunResult{
			Matches:          make([]MatchLocation, 0),
			WouldRedactCount: 0,
			PreviewData:      data,
		}

		// Sort patterns by priority
		s.sortPatterns()

		// Traverse data structure and collect matches
		s.collectMatches(data, "", result, opts)

		return nil, result
	}

	// Normal mode: sanitize and return modified data
	sanitized := s.Sanitize(data)
	return sanitized, nil
}

// Preview is a convenience method for dry-run sanitization.
//
// This is equivalent to calling SanitizeWithOptions with DryRun: true.
// It returns a DryRunResult showing what would be redacted without
// actually modifying the data.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	result := sanitizer.Preview(exportData)
//	fmt.Printf("Would redact %d values\n", result.WouldRedactCount)
//	for _, match := range result.Matches {
//	    fmt.Printf("  %s: %s → %s\n", match.Path, match.Original, match.Redacted)
//	}
//
// Parameters:
//   - data: The export data to preview
//
// Returns:
//   - *DryRunResult: Preview results showing what would be redacted
func (s *Sanitizer) Preview(data *ExportData) *DryRunResult {
	_, result := s.SanitizeWithOptions(data, SanitizeOptions{
		DryRun:        true,
		MaxPreviewLen: 100,
	})
	return result
}

// collectMatches recursively traverses the data structure and collects pattern matches.
func (s *Sanitizer) collectMatches(val interface{}, path string, result *DryRunResult, opts SanitizeOptions) {
	if val == nil {
		return
	}

	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.String:
		str := v.String()
		s.collectStringMatches(str, path, result, opts)

	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			mapPath := path
			if path != "" {
				mapPath = fmt.Sprintf("%s.%s", path, keyStr)
			} else {
				mapPath = keyStr
			}

			// Get the value
			mapValue := v.MapIndex(key).Interface()

			// For string values, check the key-value pair format that patterns expect
			if strValue, ok := mapValue.(string); ok {
				// Create a string in the format patterns expect: "key": "value"
				kvPair := fmt.Sprintf(`"%s": "%s"`, keyStr, strValue)
				s.collectStringMatches(kvPair, mapPath, result, opts)
			} else {
				// For non-string values, recurse
				s.collectMatches(mapValue, mapPath, result, opts)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			slicePath := fmt.Sprintf("%s[%d]", path, i)
			s.collectMatches(v.Index(i).Interface(), slicePath, result, opts)
		}

	case reflect.Struct:
		// Handle ExportData and its nested types
		switch val := val.(type) {
		case ExportData:
			s.collectMatchesExportData(&val, path, result, opts)
		case *ExportData:
			s.collectMatchesExportData(val, path, result, opts)
		case ComponentSnapshot:
			s.collectMatchesComponent(&val, path, result, opts)
		case *ComponentSnapshot:
			s.collectMatchesComponent(val, path, result, opts)
		case StateChange:
			s.collectMatchesStateChange(&val, path, result, opts)
		case EventRecord:
			s.collectMatchesEventRecord(&val, path, result, opts)
		default:
			// Generic struct handling
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				if field.CanInterface() {
					fieldName := v.Type().Field(i).Name
					fieldPath := path
					if path != "" {
						fieldPath = fmt.Sprintf("%s.%s", path, fieldName)
					} else {
						fieldPath = fieldName
					}
					s.collectMatches(field.Interface(), fieldPath, result, opts)
				}
			}
		}

	case reflect.Ptr:
		if !v.IsNil() {
			s.collectMatches(v.Elem().Interface(), path, result, opts)
		}

	case reflect.Interface:
		if !v.IsNil() {
			s.collectMatches(v.Elem().Interface(), path, result, opts)
		}
	}
}

// collectMatchesExportData handles ExportData structure
func (s *Sanitizer) collectMatchesExportData(data *ExportData, path string, result *DryRunResult, opts SanitizeOptions) {
	if data.Components != nil {
		for i, comp := range data.Components {
			compPath := fmt.Sprintf("components[%d]", i)
			if path != "" {
				compPath = fmt.Sprintf("%s.%s", path, compPath)
			}
			s.collectMatchesComponent(comp, compPath, result, opts)
		}
	}

	if data.State != nil {
		for i, state := range data.State {
			statePath := fmt.Sprintf("state[%d]", i)
			if path != "" {
				statePath = fmt.Sprintf("%s.%s", path, statePath)
			}
			s.collectMatchesStateChange(&state, statePath, result, opts)
		}
	}

	if data.Events != nil {
		for i, event := range data.Events {
			eventPath := fmt.Sprintf("events[%d]", i)
			if path != "" {
				eventPath = fmt.Sprintf("%s.%s", path, eventPath)
			}
			s.collectMatchesEventRecord(&event, eventPath, result, opts)
		}
	}

	if data.Performance != nil {
		perfPath := "performance"
		if path != "" {
			perfPath = fmt.Sprintf("%s.%s", path, perfPath)
		}
		s.collectMatches(data.Performance, perfPath, result, opts)
	}
}

// collectMatchesComponent handles ComponentSnapshot structure
func (s *Sanitizer) collectMatchesComponent(comp *ComponentSnapshot, path string, result *DryRunResult, opts SanitizeOptions) {
	if comp.Props != nil {
		propsPath := fmt.Sprintf("%s.props", path)
		s.collectMatches(comp.Props, propsPath, result, opts)
	}

	if comp.State != nil {
		statePath := fmt.Sprintf("%s.state", path)
		s.collectMatches(comp.State, statePath, result, opts)
	}

	if comp.Refs != nil {
		for i, ref := range comp.Refs {
			refPath := fmt.Sprintf("%s.refs[%d].value", path, i)
			s.collectMatches(ref.Value, refPath, result, opts)
		}
	}

	if comp.Children != nil {
		for i, child := range comp.Children {
			childPath := fmt.Sprintf("%s.children[%d]", path, i)
			s.collectMatchesComponent(child, childPath, result, opts)
		}
	}
}

// collectMatchesStateChange handles StateChange structure
func (s *Sanitizer) collectMatchesStateChange(state *StateChange, path string, result *DryRunResult, opts SanitizeOptions) {
	oldPath := fmt.Sprintf("%s.old_value", path)
	s.collectMatches(state.OldValue, oldPath, result, opts)

	newPath := fmt.Sprintf("%s.new_value", path)
	s.collectMatches(state.NewValue, newPath, result, opts)
}

// collectMatchesEventRecord handles EventRecord structure
func (s *Sanitizer) collectMatchesEventRecord(event *EventRecord, path string, result *DryRunResult, opts SanitizeOptions) {
	if event.Payload != nil {
		payloadPath := fmt.Sprintf("%s.payload", path)
		s.collectMatches(event.Payload, payloadPath, result, opts)
	}
}

// collectStringMatches checks a string against all patterns and collects matches
func (s *Sanitizer) collectStringMatches(str string, path string, result *DryRunResult, opts SanitizeOptions) {
	for _, pattern := range s.patterns {
		matches := pattern.Pattern.FindAllString(str, -1)
		if len(matches) > 0 {
			// Apply the pattern to get the redacted version
			redacted := pattern.Pattern.ReplaceAllString(str, pattern.Replacement)

			// Truncate original if needed
			original := str
			if opts.MaxPreviewLen > 0 && len(original) > opts.MaxPreviewLen {
				original = original[:opts.MaxPreviewLen] + "..."
			}

			// Create match location
			match := MatchLocation{
				Path:     path,
				Pattern:  pattern.Name,
				Original: original,
				Redacted: redacted,
				Line:     0, // JSON line tracking not implemented yet
				Column:   0, // JSON column tracking not implemented yet
			}

			result.Matches = append(result.Matches, match)
			result.WouldRedactCount += len(matches)

			// Only record the first match per string to avoid duplicates
			// (since patterns are applied sequentially, not all at once)
			break
		}
	}
}
