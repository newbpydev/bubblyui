package devtools

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
)

// TemplateRegistry is a map of template names to their sanitization patterns.
//
// Templates provide pre-configured pattern sets for common compliance
// requirements like PII, PCI, HIPAA, and GDPR. Each template contains
// patterns with appropriate priorities for the compliance domain.
//
// Example:
//
//	registry := devtools.DefaultTemplates
//	piiPatterns := registry["pii"]
//	fmt.Printf("PII template has %d patterns\n", len(piiPatterns))
type TemplateRegistry map[string][]SanitizePattern

var (
	// DefaultTemplates contains pre-configured compliance pattern sets.
	//
	// Available templates:
	//   - "pii": Personal Identifiable Information (SSN, email, phone)
	//   - "pci": Payment Card Industry (card numbers, CVV, expiry dates)
	//   - "hipaa": Health Insurance Portability and Accountability Act (MRN, diagnoses)
	//   - "gdpr": General Data Protection Regulation (IP addresses, MAC addresses)
	//
	// All patterns use case-insensitive matching with (?i) flag and capture
	// groups to preserve keys while redacting values: (key)(sep)(value)
	//
	// Example:
	//
	//	sanitizer := devtools.NewSanitizer()
	//	sanitizer.LoadTemplate("pii")  // Loads SSN, email, phone patterns
	//	sanitizer.LoadTemplate("pci")  // Adds card, CVV, expiry patterns
	DefaultTemplates TemplateRegistry

	// templateMu protects DefaultTemplates from concurrent modification
	templateMu sync.RWMutex
)

func init() {
	DefaultTemplates = make(TemplateRegistry)

	// PII Template - Personal Identifiable Information
	// Priority 100 for critical patterns, 90 for common patterns
	DefaultTemplates["pii"] = []SanitizePattern{
		{
			Pattern:     regexp.MustCompile(`(?i)(ssn|social[_-]?security)(["'\s:=]+)(\d{3}-?\d{2}-?\d{4})`),
			Replacement: "${1}${2}[REDACTED_SSN]",
			Priority:    100,
			Name:        "ssn",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(email|e[_-]?mail)(["'\s:=]+)([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`),
			Replacement: "${1}${2}[REDACTED_EMAIL]",
			Priority:    90,
			Name:        "email",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(phone|tel|mobile)(["'\s:=]+)(\+?1?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})`),
			Replacement: "${1}${2}[REDACTED_PHONE]",
			Priority:    90,
			Name:        "phone",
		},
	}

	// PCI Template - Payment Card Industry
	// Priority 100 for all payment-related patterns (critical)
	DefaultTemplates["pci"] = []SanitizePattern{
		{
			Pattern:     regexp.MustCompile(`(?i)(card[_-]?number|credit[_-]?card|cc[_-]?number)(["'\s:=]+)(\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4})`),
			Replacement: "${1}${2}[REDACTED_CARD]",
			Priority:    100,
			Name:        "card_number",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(cvv|cvc|security[_-]?code)(["'\s:=]+)(\d{3,4})`),
			Replacement: "${1}${2}[REDACTED_CVV]",
			Priority:    100,
			Name:        "cvv",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(expiry|exp[_-]?date|expiration)(["'\s:=]+)(\d{2}/\d{2,4})`),
			Replacement: "${1}${2}[REDACTED_EXPIRY]",
			Priority:    90,
			Name:        "expiry_date",
		},
	}

	// HIPAA Template - Health Insurance Portability and Accountability Act
	// Priority 100 for medical records, 90 for diagnoses
	DefaultTemplates["hipaa"] = []SanitizePattern{
		{
			Pattern:     regexp.MustCompile(`(?i)(mrn|medical[_-]?record[_-]?number|patient[_-]?id)(["'\s:=]+)([A-Z0-9-]+)`),
			Replacement: "${1}${2}[REDACTED_MRN]",
			Priority:    100,
			Name:        "medical_record_number",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(diagnosis|condition|icd[_-]?code)(["'\s:=]+)([A-Z0-9.-]+)`),
			Replacement: "${1}${2}[REDACTED_DIAGNOSIS]",
			Priority:    90,
			Name:        "diagnosis",
		},
	}

	// GDPR Template - General Data Protection Regulation
	// Priority 90 for all GDPR-related patterns
	DefaultTemplates["gdpr"] = []SanitizePattern{
		{
			Pattern:     regexp.MustCompile(`(?i)(ip[_-]?address|ip)(["'\s:=]+)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`),
			Replacement: "${1}${2}[REDACTED_IP]",
			Priority:    90,
			Name:        "ip_address",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(mac[_-]?address|mac)(["'\s:=]+)([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`),
			Replacement: "${1}${2}[REDACTED_MAC]",
			Priority:    90,
			Name:        "mac_address",
		},
	}
}

// LoadTemplate loads a pre-configured template by name and appends its patterns.
//
// This method is composable - calling it multiple times will append patterns
// from each template. Patterns are applied in priority order when sanitizing.
//
// Available templates: "pii", "pci", "hipaa", "gdpr"
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	err := sanitizer.LoadTemplate("pii")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Now sanitizer has default patterns + PII patterns
//
// Parameters:
//   - name: Template name (case-sensitive)
//
// Returns:
//   - error: Error if template name is invalid
func (s *Sanitizer) LoadTemplate(name string) error {
	templateMu.RLock()
	patterns, exists := DefaultTemplates[name]
	templateMu.RUnlock()

	if !exists {
		return fmt.Errorf("template not found: %s", name)
	}

	// Append patterns from template
	s.patterns = append(s.patterns, patterns...)

	return nil
}

// LoadTemplates loads multiple templates at once and appends all their patterns.
//
// This is a convenience method for loading multiple compliance templates.
// Patterns from all templates are combined and will be applied in priority order.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	err := sanitizer.LoadTemplates("pii", "pci", "hipaa")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Now sanitizer has patterns from all three templates
//
// Parameters:
//   - names: Template names to load (case-sensitive)
//
// Returns:
//   - error: Error if any template name is invalid
func (s *Sanitizer) LoadTemplates(names ...string) error {
	for _, name := range names {
		if err := s.LoadTemplate(name); err != nil {
			return err
		}
	}
	return nil
}

// MergeTemplates combines patterns from multiple templates without modifying the sanitizer.
//
// This method is useful for previewing what patterns would be loaded from
// multiple templates, or for creating custom template combinations.
//
// The returned patterns are sorted by priority (highest first) and maintain
// insertion order for equal priorities.
//
// Example:
//
//	sanitizer := devtools.NewSanitizer()
//	patterns, err := sanitizer.MergeTemplates("pii", "pci")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Combined templates have %d patterns\n", len(patterns))
//
// Parameters:
//   - names: Template names to merge (case-sensitive)
//
// Returns:
//   - []SanitizePattern: Combined patterns sorted by priority
//   - error: Error if any template name is invalid
func (s *Sanitizer) MergeTemplates(names ...string) ([]SanitizePattern, error) {
	templateMu.RLock()
	defer templateMu.RUnlock()

	var merged []SanitizePattern

	for _, name := range names {
		patterns, exists := DefaultTemplates[name]
		if !exists {
			return nil, fmt.Errorf("template not found: %s", name)
		}
		merged = append(merged, patterns...)
	}

	// Sort by priority (highest first), stable sort to maintain insertion order
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].Priority > merged[j].Priority
	})

	return merged, nil
}

// RegisterTemplate registers a custom template in the global registry.
//
// This allows applications to define their own compliance templates that
// can be loaded by name. Custom templates can be used alongside built-in
// templates.
//
// Thread Safety:
//
//	Safe to call concurrently. Uses mutex to protect registry.
//
// Example:
//
//	customPatterns := []devtools.SanitizePattern{
//	    {
//	        Pattern:     regexp.MustCompile(`(?i)(internal[_-]?id)(["'\s:=]+)([A-Z0-9-]+)`),
//	        Replacement: "${1}${2}[REDACTED_ID]",
//	        Priority:    80,
//	        Name:        "internal_id",
//	    },
//	}
//	err := devtools.RegisterTemplate("custom", customPatterns)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - name: Template name (case-sensitive)
//   - patterns: Sanitization patterns for this template
//
// Returns:
//   - error: Error if template name is empty or already exists
func RegisterTemplate(name string, patterns []SanitizePattern) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	templateMu.Lock()
	defer templateMu.Unlock()

	if _, exists := DefaultTemplates[name]; exists {
		return fmt.Errorf("template already exists: %s", name)
	}

	DefaultTemplates[name] = patterns
	return nil
}

// GetTemplateNames returns a sorted list of all available template names.
//
// This includes both built-in templates (pii, pci, hipaa, gdpr) and any
// custom templates registered via RegisterTemplate().
//
// Thread Safety:
//
//	Safe to call concurrently. Uses read lock on registry.
//
// Example:
//
//	names := devtools.GetTemplateNames()
//	fmt.Printf("Available templates: %v\n", names)
//	// Output: Available templates: [gdpr hipaa pci pii]
//
// Returns:
//   - []string: Sorted list of template names
func GetTemplateNames() []string {
	templateMu.RLock()
	defer templateMu.RUnlock()

	names := make([]string, 0, len(DefaultTemplates))
	for name := range DefaultTemplates {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
