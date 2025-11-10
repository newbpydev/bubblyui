package devtools

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
)

// Sanitizer provides regex-based sanitization of sensitive data in exports.
//
// It uses regular expression patterns to identify and redact sensitive
// information like passwords, tokens, API keys, and secrets. The sanitizer
// can handle nested data structures including maps, slices, and structs.
//
// Thread Safety:
//
//	Sanitizer is safe to use concurrently after creation. The patterns
//	slice should not be modified after creation.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	sanitizer.AddPattern(`(?i)password["\s:=]+\S+`, "[REDACTED]")
//	cleanData := sanitizer.Sanitize(exportData)
type Sanitizer struct {
	// patterns is the list of regex patterns to apply
	patterns []SanitizePattern
}

// SanitizePattern represents a single sanitization rule with priority ordering.
//
// It contains a compiled regular expression pattern, the replacement
// string to use when the pattern matches, a priority for ordering, and
// an optional name for tracking/debugging.
//
// Priority Ranges:
//   - 100+: Critical patterns (e.g., PCI, HIPAA compliance)
//   - 50-99: Organization-specific patterns
//   - 10-49: Custom patterns
//   - 0-9: Default patterns
//   - Negative: Cleanup patterns (apply last)
//
// Higher priority patterns are applied first. When priorities are equal,
// patterns are applied in insertion order (stable sort).
//
// Example:
//
//	pattern := SanitizePattern{
//	    Pattern:     regexp.MustCompile(`(?i)api[_-]?key["\s:=]+\S+`),
//	    Replacement: "[REDACTED]",
//	    Priority:    50,
//	    Name:        "api_key",
//	}
type SanitizePattern struct {
	// Pattern is the compiled regular expression to match
	Pattern *regexp.Regexp

	// Replacement is the string to replace matches with
	Replacement string

	// Priority determines the order in which patterns are applied.
	// Higher priority patterns are applied first. Default is 0.
	Priority int

	// Name is an optional identifier for the pattern, useful for
	// tracking, debugging, and audit trails. If empty, a name will
	// be auto-generated in the format "pattern_N".
	Name string
}

// NewSanitizer creates a new sanitizer with default patterns.
//
// The default patterns cover common sensitive data:
//   - Passwords (password, passwd, pwd)
//   - Tokens (token, bearer)
//   - API keys (api_key, apikey, api-key)
//   - Secrets (secret, private_key, private-key)
//
// All patterns are case-insensitive and match common formats like:
//   - JSON: "password": "secret123"
//   - URL params: password=secret123
//   - Headers: Authorization: Bearer token123
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	cleanData := sanitizer.Sanitize(exportData)
//
// Returns:
//   - *Sanitizer: A new sanitizer with default patterns
func NewSanitizer() *Sanitizer {
	s := &Sanitizer{
		patterns: make([]SanitizePattern, 0, 10),
	}

	// Add default patterns - capture key and value, replace only value
	s.AddPattern(`(?i)(password|passwd|pwd)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]")
	s.AddPattern(`(?i)(token|bearer)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]")
	s.AddPattern(`(?i)(api[_-]?key|apikey)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]")
	s.AddPattern(`(?i)(secret|private[_-]?key)(["'\s:=]+)([^\s"']+)`, "${1}${2}[REDACTED]")

	return s
}

// AddPattern adds a new sanitization pattern to the sanitizer with default priority (0).
//
// The pattern string is compiled as a regular expression. If compilation
// fails, this method panics. Use this during initialization when you want
// the program to fail fast on invalid patterns.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	sanitizer.AddPattern(`(?i)credit[_-]?card["\s:=]+\d+`, "[REDACTED]")
//
// Parameters:
//   - pattern: Regular expression pattern string
//   - replacement: String to replace matches with
func (s *Sanitizer) AddPattern(pattern, replacement string) {
	re := regexp.MustCompile(pattern)
	s.patterns = append(s.patterns, SanitizePattern{
		Pattern:     re,
		Replacement: replacement,
		Priority:    0,
		Name:        fmt.Sprintf("pattern_%d", len(s.patterns)),
	})
}

// AddPatternWithPriority adds a new sanitization pattern with explicit priority and name.
//
// The pattern string is compiled as a regular expression. If compilation
// fails, this method returns an error instead of panicking, making it
// suitable for runtime pattern addition.
//
// Priority determines the order in which patterns are applied:
//   - 100+: Critical patterns (PCI, HIPAA compliance)
//   - 50-99: Organization-specific patterns
//   - 10-49: Custom patterns
//   - 0-9: Default patterns
//   - Negative: Cleanup patterns (apply last)
//
// Higher priority patterns are applied first. When priorities are equal,
// patterns are applied in insertion order (stable sort).
//
// If name is empty, a name will be auto-generated in the format "pattern_N".
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	err := sanitizer.AddPatternWithPriority(
//	    `(?i)(merchant[_-]?id)(["'\s:=]+)([A-Z0-9]+)`,
//	    "${1}${2}[REDACTED_MERCHANT]",
//	    80,
//	    "merchant_id",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - pattern: Regular expression pattern string
//   - replacement: String to replace matches with
//   - priority: Priority for pattern ordering (higher applies first)
//   - name: Optional name for tracking/debugging (auto-generated if empty)
//
// Returns:
//   - error: Error if pattern compilation fails
func (s *Sanitizer) AddPatternWithPriority(pattern, replacement string, priority int, name string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	// Auto-generate name if empty
	if name == "" {
		name = fmt.Sprintf("pattern_%d", len(s.patterns))
	}

	s.patterns = append(s.patterns, SanitizePattern{
		Pattern:     re,
		Replacement: replacement,
		Priority:    priority,
		Name:        name,
	})

	return nil
}

// sortPatterns sorts patterns by priority (descending) using stable sort.
//
// This ensures that:
//   - Higher priority patterns are applied first
//   - Equal priority patterns maintain insertion order
//
// Uses sort.SliceStable to guarantee stable sorting behavior.
func (s *Sanitizer) sortPatterns() {
	sort.SliceStable(s.patterns, func(i, j int) bool {
		// Sort by priority descending (higher priority first)
		return s.patterns[i].Priority > s.patterns[j].Priority
	})
}

// GetPatterns returns a copy of all patterns in sorted order (by priority, descending).
//
// The returned slice is a copy, so modifications will not affect the sanitizer.
// Patterns are sorted by priority (highest first), with equal priorities
// maintaining insertion order.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	patterns := sanitizer.GetPatterns()
//	for _, p := range patterns {
//	    fmt.Printf("Pattern: %s, Priority: %d\n", p.Name, p.Priority)
//	}
//
// Returns:
//   - []SanitizePattern: Copy of patterns in priority order
func (s *Sanitizer) GetPatterns() []SanitizePattern {
	// Sort patterns first
	s.sortPatterns()

	// Return a copy to prevent external modification
	result := make([]SanitizePattern, len(s.patterns))
	copy(result, s.patterns)
	return result
}

// Sanitize creates a sanitized copy of the export data.
//
// This method applies all configured patterns to the export data,
// recursively sanitizing nested structures. The original data is
// not modified - a deep copy is created and sanitized.
//
// Thread Safety:
//
//	Safe to call concurrently. Does not modify the input data.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	cleanData := sanitizer.Sanitize(exportData)
//
// Parameters:
//   - data: The export data to sanitize
//
// Returns:
//   - *ExportData: A sanitized copy of the export data
func (s *Sanitizer) Sanitize(data *ExportData) *ExportData {
	if data == nil {
		return nil
	}

	// Create a copy of the export data
	result := &ExportData{
		Version:   data.Version,
		Timestamp: data.Timestamp,
	}

	// Sanitize components
	if data.Components != nil {
		result.Components = make([]*ComponentSnapshot, len(data.Components))
		for i, comp := range data.Components {
			result.Components[i] = s.sanitizeComponent(comp)
		}
	}

	// Sanitize state history
	if data.State != nil {
		result.State = make([]StateChange, len(data.State))
		for i, state := range data.State {
			result.State[i] = s.sanitizeStateChange(state)
		}
	}

	// Sanitize events
	if data.Events != nil {
		result.Events = make([]EventRecord, len(data.Events))
		for i, event := range data.Events {
			result.Events[i] = s.sanitizeEventRecord(event)
		}
	}

	// Sanitize performance data
	if data.Performance != nil {
		result.Performance = s.sanitizePerformanceData(data.Performance)
	}

	return result
}

// sanitizeComponent creates a sanitized copy of a component snapshot.
func (s *Sanitizer) sanitizeComponent(comp *ComponentSnapshot) *ComponentSnapshot {
	if comp == nil {
		return nil
	}

	result := &ComponentSnapshot{
		ID:        comp.ID,
		Name:      comp.Name,
		Type:      comp.Type,
		Timestamp: comp.Timestamp,
	}

	// Sanitize props
	if comp.Props != nil {
		result.Props = s.SanitizeValue(comp.Props).(map[string]interface{})
	}

	// Sanitize state
	if comp.State != nil {
		result.State = s.SanitizeValue(comp.State).(map[string]interface{})
	}

	// Sanitize refs
	if comp.Refs != nil {
		result.Refs = make([]*RefSnapshot, len(comp.Refs))
		for i, ref := range comp.Refs {
			result.Refs[i] = &RefSnapshot{
				ID:    ref.ID,
				Name:  ref.Name,
				Value: s.SanitizeValue(ref.Value),
			}
		}
	}

	// Sanitize children
	if comp.Children != nil {
		result.Children = make([]*ComponentSnapshot, len(comp.Children))
		for i, child := range comp.Children {
			result.Children[i] = s.sanitizeComponent(child)
		}
	}

	return result
}

// sanitizeStateChange creates a sanitized copy of a state change.
func (s *Sanitizer) sanitizeStateChange(state StateChange) StateChange {
	return StateChange{
		RefID:     state.RefID,
		RefName:   state.RefName,
		OldValue:  s.SanitizeValue(state.OldValue),
		NewValue:  s.SanitizeValue(state.NewValue),
		Timestamp: state.Timestamp,
		Source:    state.Source,
	}
}

// sanitizeEventRecord creates a sanitized copy of an event record.
func (s *Sanitizer) sanitizeEventRecord(event EventRecord) EventRecord {
	return EventRecord{
		ID:        event.ID,
		Name:      event.Name,
		SourceID:  event.SourceID,
		Timestamp: event.Timestamp,
		Payload:   s.SanitizeValue(event.Payload),
	}
}

// sanitizePerformanceData creates a sanitized copy of performance data.
func (s *Sanitizer) sanitizePerformanceData(perf *PerformanceData) *PerformanceData {
	if perf == nil {
		return nil
	}

	// Performance data typically doesn't contain sensitive info,
	// but we'll sanitize component names just in case
	result := NewPerformanceData()
	allPerf := perf.GetAll()

	for id, compPerf := range allPerf {
		result.components[id] = compPerf // Copy as-is for now
	}

	return result
}

// SanitizeValue recursively sanitizes a value of any type.
//
// This method handles:
//   - Strings: applies all regex patterns in priority order
//   - Maps: recursively sanitizes all values
//   - Slices: recursively sanitizes all elements
//   - Structs: recursively sanitizes all exported fields
//   - Primitives: returns as-is (numbers, bools, etc.)
//
// Patterns are applied in priority order (highest first), with equal
// priorities maintaining insertion order (stable sort).
//
// The original value is not modified - a deep copy is created.
//
// Thread Safety:
//
//	Safe to call concurrently. Does not modify the input value.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	cleanValue := sanitizer.SanitizeValue(map[string]interface{}{
//	    "username": "alice",
//	    "password": "secret123",
//	})
//	// cleanValue["password"] will be "[REDACTED]"
//
// Parameters:
//   - val: The value to sanitize
//
// Returns:
//   - interface{}: A sanitized copy of the value
func (s *Sanitizer) SanitizeValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	// Sort patterns by priority before applying
	s.sortPatterns()

	// Use reflection to handle different types
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.String:
		// Apply all patterns to the string in priority order
		str := v.String()
		for _, pattern := range s.patterns {
			str = pattern.Pattern.ReplaceAllString(str, pattern.Replacement)
		}
		return str

	case reflect.Map:
		// Create a new map and sanitize all values
		result := reflect.MakeMap(v.Type())
		for _, key := range v.MapKeys() {
			sanitizedValue := s.SanitizeValue(v.MapIndex(key).Interface())
			result.SetMapIndex(key, reflect.ValueOf(sanitizedValue))
		}
		return result.Interface()

	case reflect.Slice, reflect.Array:
		// Create a new slice and sanitize all elements
		result := reflect.MakeSlice(reflect.SliceOf(v.Type().Elem()), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			sanitizedElem := s.SanitizeValue(v.Index(i).Interface())
			result.Index(i).Set(reflect.ValueOf(sanitizedElem))
		}
		return result.Interface()

	case reflect.Struct:
		// Create a new struct and sanitize all exported fields
		result := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanInterface() { // Only exported fields
				sanitizedField := s.SanitizeValue(field.Interface())
				if result.Field(i).CanSet() {
					result.Field(i).Set(reflect.ValueOf(sanitizedField))
				}
			}
		}
		return result.Interface()

	case reflect.Ptr:
		// Handle pointers by sanitizing the pointed-to value
		if v.IsNil() {
			return nil
		}
		sanitized := s.SanitizeValue(v.Elem().Interface())
		result := reflect.New(v.Elem().Type())
		result.Elem().Set(reflect.ValueOf(sanitized))
		return result.Interface()

	case reflect.Interface:
		// Handle interface{} by sanitizing the concrete value
		if v.IsNil() {
			return nil
		}
		return s.SanitizeValue(v.Elem().Interface())

	default:
		// For primitives (int, bool, float, etc.), return as-is
		return val
	}
}

// DefaultPatterns returns the default sanitization patterns.
//
// This is useful for understanding what patterns are applied by default
// or for creating a custom sanitizer with modified default patterns.
//
// Returns:
//   - []string: List of default regex pattern strings
func DefaultPatterns() []string {
	return []string{
		`(?i)(password|passwd|pwd)(["'\s:=]+)([^\s"']+)`,
		`(?i)(token|bearer)(["'\s:=]+)([^\s"']+)`,
		`(?i)(api[_-]?key|apikey)(["'\s:=]+)([^\s"']+)`,
		`(?i)(secret|private[_-]?key)(["'\s:=]+)([^\s"']+)`,
	}
}

// SanitizeString applies all patterns to a single string in priority order.
//
// This is a convenience method for sanitizing individual strings
// without needing to sanitize an entire ExportData structure.
//
// Patterns are applied in priority order (highest first), with equal
// priorities maintaining insertion order (stable sort).
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	clean := sanitizer.SanitizeString(`{"password": "secret123"}`)
//	// clean will be `{"password": "[REDACTED]"}`
//
// Parameters:
//   - str: The string to sanitize
//
// Returns:
//   - string: The sanitized string
func (s *Sanitizer) SanitizeString(str string) string {
	// Sort patterns by priority before applying
	s.sortPatterns()

	result := str
	for _, pattern := range s.patterns {
		result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
	}
	return result
}

// PatternCount returns the number of patterns configured.
//
// Returns:
//   - int: Number of sanitization patterns
func (s *Sanitizer) PatternCount() int {
	return len(s.patterns)
}
