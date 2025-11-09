package devtools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ExportData represents the complete debug data export format.
//
// This structure is serialized to JSON when exporting dev tools data.
// It includes version information, timestamp, and optional sections for
// components, state history, events, and performance metrics.
//
// Thread Safety:
//
//	ExportData is a value type and safe to use concurrently after creation.
//
// Example:
//
//	data := ExportData{
//	    Version:   "1.0",
//	    Timestamp: time.Now(),
//	    Components: components,
//	    State:     stateHistory,
//	}
type ExportData struct {
	// Version is the export format version (currently "1.0")
	Version string `json:"version"`

	// Timestamp is when the export was created
	Timestamp time.Time `json:"timestamp"`

	// Components is the list of component snapshots (optional)
	Components []*ComponentSnapshot `json:"components,omitempty"`

	// State is the state change history (optional)
	State []StateChange `json:"state,omitempty"`

	// Events is the event log (optional)
	Events []EventRecord `json:"events,omitempty"`

	// Performance is the performance metrics (optional)
	Performance *PerformanceData `json:"performance,omitempty"`
}

// ExportOptions configures what data to include in the export.
//
// Use this to selectively export only the data you need, reducing
// file size and export time. Sanitization can be enabled to redact
// sensitive data before export.
//
// Example:
//
//	opts := ExportOptions{
//	    IncludeComponents: true,
//	    IncludeState:      true,
//	    Sanitize:          true,
//	    RedactPatterns:    []string{"password", "token"},
//	}
type ExportOptions struct {
	// IncludeComponents determines if component snapshots are exported
	IncludeComponents bool

	// IncludeState determines if state change history is exported
	IncludeState bool

	// IncludeEvents determines if event log is exported
	IncludeEvents bool

	// IncludePerformance determines if performance metrics are exported
	IncludePerformance bool

	// Sanitize enables redaction of sensitive data
	Sanitize bool

	// RedactPatterns is a list of case-insensitive strings to redact
	// Common patterns: "password", "token", "apikey", "secret"
	RedactPatterns []string
}

// Export writes dev tools debug data to a JSON file.
//
// The export includes version information, timestamp, and optionally
// components, state history, events, and performance metrics based on
// the provided options. If sanitization is enabled, sensitive data
// matching the redact patterns will be replaced with "[REDACTED]".
//
// Thread Safety:
//
//	Safe to call concurrently. Uses read lock on DevTools.
//
// Example:
//
//	opts := ExportOptions{
//	    IncludeComponents: true,
//	    IncludeState:      true,
//	    IncludeEvents:     true,
//	    Sanitize:          true,
//	    RedactPatterns:    []string{"password", "token"},
//	}
//	err := devtools.Export("debug-state.json", opts)
//	if err != nil {
//	    log.Printf("Export failed: %v", err)
//	}
//
// Parameters:
//   - filename: Path to the output JSON file
//   - opts: Export options controlling what data to include
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
func (dt *DevTools) Export(filename string, opts ExportOptions) error {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	// Check if dev tools is enabled
	if !dt.enabled {
		return fmt.Errorf("dev tools not enabled")
	}

	// Check if store exists
	if dt.store == nil {
		return fmt.Errorf("dev tools store not initialized")
	}

	// Create export data structure
	data := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
	}

	// Collect components if requested
	if opts.IncludeComponents {
		data.Components = dt.store.GetAllComponents()
	}

	// Collect state history if requested
	if opts.IncludeState {
		data.State = dt.store.stateHistory.GetAll()
	}

	// Collect events if requested
	if opts.IncludeEvents {
		// Get all events (use a large number to get everything)
		data.Events = dt.store.events.GetRecent(dt.store.events.Len())
	}

	// Collect performance data if requested
	if opts.IncludePerformance {
		data.Performance = dt.store.performance
	}

	// Apply sanitization if requested
	if opts.Sanitize {
		data = sanitizeExportData(data, opts.RedactPatterns)
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Write to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// sanitizeExportData redacts sensitive data from export data.
//
// This function performs basic string-based sanitization by replacing
// values that contain any of the redact patterns with "[REDACTED]".
// The search is case-insensitive.
//
// Note: This is a basic implementation. Task 6.3 will implement
// comprehensive regex-based sanitization.
//
// Parameters:
//   - data: The export data to sanitize
//   - patterns: List of case-insensitive strings to redact
//
// Returns:
//   - ExportData: Sanitized copy of the export data
func sanitizeExportData(data ExportData, patterns []string) ExportData {
	// If no patterns, return as-is
	if len(patterns) == 0 {
		return data
	}

	// Create lowercase patterns for case-insensitive matching
	lowerPatterns := make([]string, len(patterns))
	for i, p := range patterns {
		lowerPatterns[i] = strings.ToLower(p)
	}

	// Sanitize components
	if data.Components != nil {
		for _, comp := range data.Components {
			// Sanitize props
			if comp.Props != nil {
				for key, val := range comp.Props {
					if shouldRedact(key, lowerPatterns) || shouldRedactValue(val, lowerPatterns) {
						comp.Props[key] = "[REDACTED]"
					}
				}
			}

			// Sanitize state
			if comp.State != nil {
				for key, val := range comp.State {
					if shouldRedact(key, lowerPatterns) || shouldRedactValue(val, lowerPatterns) {
						comp.State[key] = "[REDACTED]"
					}
				}
			}

			// Sanitize refs
			for _, ref := range comp.Refs {
				if shouldRedact(ref.Name, lowerPatterns) || shouldRedactValue(ref.Value, lowerPatterns) {
					ref.Value = "[REDACTED]"
				}
			}
		}
	}

	// Sanitize state history
	if data.State != nil {
		for i := range data.State {
			if shouldRedact(data.State[i].RefName, lowerPatterns) ||
				shouldRedactValue(data.State[i].OldValue, lowerPatterns) ||
				shouldRedactValue(data.State[i].NewValue, lowerPatterns) {
				data.State[i].OldValue = "[REDACTED]"
				data.State[i].NewValue = "[REDACTED]"
			}
		}
	}

	// Sanitize events
	if data.Events != nil {
		for i := range data.Events {
			if shouldRedact(data.Events[i].Name, lowerPatterns) ||
				shouldRedactValue(data.Events[i].Payload, lowerPatterns) {
				data.Events[i].Payload = "[REDACTED]"
			}
		}
	}

	return data
}

// shouldRedact checks if a string key contains any redact pattern.
func shouldRedact(key string, patterns []string) bool {
	lowerKey := strings.ToLower(key)
	for _, pattern := range patterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}
	return false
}

// shouldRedactValue checks if a value (converted to string) contains any redact pattern.
func shouldRedactValue(val interface{}, patterns []string) bool {
	if val == nil {
		return false
	}

	// Convert value to string for checking
	valStr := fmt.Sprintf("%v", val)
	lowerVal := strings.ToLower(valStr)

	for _, pattern := range patterns {
		if strings.Contains(lowerVal, pattern) {
			return true
		}
	}
	return false
}
