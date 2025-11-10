package devtools

import (
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTemplates_Initialization(t *testing.T) {
	// Verify all expected templates exist
	expectedTemplates := []string{"pii", "pci", "hipaa", "gdpr"}

	for _, name := range expectedTemplates {
		t.Run(name, func(t *testing.T) {
			patterns, exists := DefaultTemplates[name]
			assert.True(t, exists, "Template %s should exist", name)
			assert.NotEmpty(t, patterns, "Template %s should have patterns", name)
		})
	}
}

func TestPIITemplate_Patterns(t *testing.T) {
	patterns := DefaultTemplates["pii"]
	require.NotNil(t, patterns)
	assert.Len(t, patterns, 3, "PII template should have 3 patterns")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SSN with dashes",
			input:    `{"ssn": "123-45-6789"}`,
			expected: `{"ssn": "[REDACTED_SSN]"}`,
		},
		{
			name:     "SSN without dashes",
			input:    `{"social_security": "123456789"}`,
			expected: `{"social_security": "[REDACTED_SSN]"}`,
		},
		{
			name:     "Email address",
			input:    `{"email": "user@example.com"}`,
			expected: `{"email": "[REDACTED_EMAIL]"}`,
		},
		{
			name:     "Phone number with dashes",
			input:    `{"phone": "555-123-4567"}`,
			expected: `{"phone": "[REDACTED_PHONE]"}`,
		},
		{
			name:     "Phone number with parentheses",
			input:    `{"tel": "(555) 123-4567"}`,
			expected: `{"tel": "[REDACTED_PHONE]"}`,
		},
		{
			name:     "Case insensitive SSN",
			input:    `{"SSN": "123-45-6789"}`,
			expected: `{"SSN": "[REDACTED_SSN]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			for _, pattern := range patterns {
				result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPCITemplate_Patterns(t *testing.T) {
	patterns := DefaultTemplates["pci"]
	require.NotNil(t, patterns)
	assert.Len(t, patterns, 3, "PCI template should have 3 patterns")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Credit card with spaces",
			input:    `{"card_number": "4532 1234 5678 9010"}`,
			expected: `{"card_number": "[REDACTED_CARD]"}`,
		},
		{
			name:     "Credit card with dashes",
			input:    `{"credit_card": "4532-1234-5678-9010"}`,
			expected: `{"credit_card": "[REDACTED_CARD]"}`,
		},
		{
			name:     "CVV 3 digits",
			input:    `{"cvv": "123"}`,
			expected: `{"cvv": "[REDACTED_CVV]"}`,
		},
		{
			name:     "CVV 4 digits (Amex)",
			input:    `{"cvc": "1234"}`,
			expected: `{"cvc": "[REDACTED_CVV]"}`,
		},
		{
			name:     "Expiry date MM/YY",
			input:    `{"expiry": "12/25"}`,
			expected: `{"expiry": "[REDACTED_EXPIRY]"}`,
		},
		{
			name:     "Expiry date MM/YYYY",
			input:    `{"exp_date": "12/2025"}`,
			expected: `{"exp_date": "[REDACTED_EXPIRY]"}`,
		},
		{
			name:     "Case insensitive card",
			input:    `{"CARD_NUMBER": "4532 1234 5678 9010"}`,
			expected: `{"CARD_NUMBER": "[REDACTED_CARD]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			for _, pattern := range patterns {
				result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHIPAATemplate_Patterns(t *testing.T) {
	patterns := DefaultTemplates["hipaa"]
	require.NotNil(t, patterns)
	assert.Len(t, patterns, 2, "HIPAA template should have 2 patterns")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Medical record number",
			input:    `{"mrn": "MRN-12345"}`,
			expected: `{"mrn": "[REDACTED_MRN]"}`,
		},
		{
			name:     "Patient ID",
			input:    `{"patient_id": "PT-ABC-123"}`,
			expected: `{"patient_id": "[REDACTED_MRN]"}`,
		},
		{
			name:     "Diagnosis code",
			input:    `{"diagnosis": "J45.909"}`,
			expected: `{"diagnosis": "[REDACTED_DIAGNOSIS]"}`,
		},
		{
			name:     "ICD code",
			input:    `{"icd_code": "E11.9"}`,
			expected: `{"icd_code": "[REDACTED_DIAGNOSIS]"}`,
		},
		{
			name:     "Case insensitive MRN",
			input:    `{"MRN": "MRN-12345"}`,
			expected: `{"MRN": "[REDACTED_MRN]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			for _, pattern := range patterns {
				result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGDPRTemplate_Patterns(t *testing.T) {
	patterns := DefaultTemplates["gdpr"]
	require.NotNil(t, patterns)
	assert.Len(t, patterns, 2, "GDPR template should have 2 patterns")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 address",
			input:    `{"ip_address": "192.168.1.1"}`,
			expected: `{"ip_address": "[REDACTED_IP]"}`,
		},
		{
			name:     "IP field",
			input:    `{"ip": "10.0.0.1"}`,
			expected: `{"ip": "[REDACTED_IP]"}`,
		},
		{
			name:     "MAC address with colons",
			input:    `{"mac_address": "00:1B:44:11:3A:B7"}`,
			expected: `{"mac_address": "[REDACTED_MAC]"}`,
		},
		{
			name:     "MAC address with dashes",
			input:    `{"mac": "00-1B-44-11-3A-B7"}`,
			expected: `{"mac": "[REDACTED_MAC]"}`,
		},
		{
			name:     "Case insensitive IP",
			input:    `{"IP_ADDRESS": "192.168.1.1"}`,
			expected: `{"IP_ADDRESS": "[REDACTED_IP]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			for _, pattern := range patterns {
				result = pattern.Pattern.ReplaceAllString(result, pattern.Replacement)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizer_LoadTemplate(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		wantErr      bool
		wantPatterns int // Additional patterns beyond default 4
	}{
		{
			name:         "Load PII template",
			template:     "pii",
			wantErr:      false,
			wantPatterns: 3,
		},
		{
			name:         "Load PCI template",
			template:     "pci",
			wantErr:      false,
			wantPatterns: 3,
		},
		{
			name:         "Load HIPAA template",
			template:     "hipaa",
			wantErr:      false,
			wantPatterns: 2,
		},
		{
			name:         "Load GDPR template",
			template:     "gdpr",
			wantErr:      false,
			wantPatterns: 2,
		},
		{
			name:         "Invalid template name",
			template:     "invalid",
			wantErr:      true,
			wantPatterns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()
			initialCount := sanitizer.PatternCount()

			err := sanitizer.LoadTemplate(tt.template)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "template not found")
				assert.Equal(t, initialCount, sanitizer.PatternCount())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, initialCount+tt.wantPatterns, sanitizer.PatternCount())
			}
		})
	}
}

func TestSanitizer_LoadTemplates(t *testing.T) {
	tests := []struct {
		name         string
		templates    []string
		wantErr      bool
		wantPatterns int // Total additional patterns
	}{
		{
			name:         "Load multiple valid templates",
			templates:    []string{"pii", "pci"},
			wantErr:      false,
			wantPatterns: 6, // 3 PII + 3 PCI
		},
		{
			name:         "Load all templates",
			templates:    []string{"pii", "pci", "hipaa", "gdpr"},
			wantErr:      false,
			wantPatterns: 10, // 3 + 3 + 2 + 2
		},
		{
			name:         "Load with invalid template",
			templates:    []string{"pii", "invalid", "pci"},
			wantErr:      true,
			wantPatterns: 3, // Only PII loaded before error
		},
		{
			name:         "Empty list",
			templates:    []string{},
			wantErr:      false,
			wantPatterns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()
			initialCount := sanitizer.PatternCount()

			err := sanitizer.LoadTemplates(tt.templates...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, initialCount+tt.wantPatterns, sanitizer.PatternCount())
			}
		})
	}
}

func TestSanitizer_MergeTemplates(t *testing.T) {
	tests := []struct {
		name          string
		templates     []string
		wantErr       bool
		wantPatterns  int
		checkPriority bool
	}{
		{
			name:          "Merge PII and PCI",
			templates:     []string{"pii", "pci"},
			wantErr:       false,
			wantPatterns:  6,
			checkPriority: true,
		},
		{
			name:          "Merge all templates",
			templates:     []string{"pii", "pci", "hipaa", "gdpr"},
			wantErr:       false,
			wantPatterns:  10,
			checkPriority: true,
		},
		{
			name:          "Merge with invalid template",
			templates:     []string{"pii", "invalid"},
			wantErr:       true,
			wantPatterns:  0,
			checkPriority: false,
		},
		{
			name:          "Single template",
			templates:     []string{"pii"},
			wantErr:       false,
			wantPatterns:  3,
			checkPriority: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()

			patterns, err := sanitizer.MergeTemplates(tt.templates...)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, patterns)
			} else {
				assert.NoError(t, err)
				assert.Len(t, patterns, tt.wantPatterns)

				// Verify priority ordering
				if tt.checkPriority {
					for i := 1; i < len(patterns); i++ {
						assert.GreaterOrEqual(t, patterns[i-1].Priority, patterns[i].Priority,
							"Patterns should be sorted by priority (highest first)")
					}
				}
			}
		})
	}
}

func TestRegisterTemplate(t *testing.T) {
	// Clean up after test
	defer func() {
		templateMu.Lock()
		delete(DefaultTemplates, "custom_test")
		delete(DefaultTemplates, "another_test")
		templateMu.Unlock()
	}()

	tests := []struct {
		name     string
		tmplName string
		patterns []SanitizePattern
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Register valid custom template",
			tmplName: "custom_test",
			patterns: []SanitizePattern{
				{
					Pattern:     regexp.MustCompile(`(?i)(custom)(["'\s:=]+)([^\s"']+)`),
					Replacement: "${1}${2}[REDACTED]",
					Priority:    50,
					Name:        "custom_pattern",
				},
			},
			wantErr: false,
		},
		{
			name:     "Empty template name",
			tmplName: "",
			patterns: []SanitizePattern{},
			wantErr:  true,
			errMsg:   "template name cannot be empty",
		},
		{
			name:     "Duplicate template name",
			tmplName: "pii", // Built-in template
			patterns: []SanitizePattern{},
			wantErr:  true,
			errMsg:   "template already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterTemplate(tt.tmplName, tt.patterns)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)

				// Verify template was registered
				templateMu.RLock()
				patterns, exists := DefaultTemplates[tt.tmplName]
				templateMu.RUnlock()

				assert.True(t, exists)
				assert.Equal(t, tt.patterns, patterns)
			}
		})
	}
}

func TestGetTemplateNames(t *testing.T) {
	// Register a custom template for testing
	customPatterns := []SanitizePattern{
		{
			Pattern:     regexp.MustCompile(`test`),
			Replacement: "[TEST]",
			Priority:    1,
			Name:        "test",
		},
	}
	err := RegisterTemplate("custom_test_names", customPatterns)
	require.NoError(t, err)

	// Clean up after test
	defer func() {
		templateMu.Lock()
		delete(DefaultTemplates, "custom_test_names")
		templateMu.Unlock()
	}()

	names := GetTemplateNames()

	// Should have 4 built-in + 1 custom
	assert.Len(t, names, 5)

	// Should be sorted alphabetically
	assert.Equal(t, []string{"custom_test_names", "gdpr", "hipaa", "pci", "pii"}, names)

	// Verify all built-in templates are present
	assert.Contains(t, names, "pii")
	assert.Contains(t, names, "pci")
	assert.Contains(t, names, "hipaa")
	assert.Contains(t, names, "gdpr")
	assert.Contains(t, names, "custom_test_names")
}

func TestTemplatePatternPriorities(t *testing.T) {
	tests := []struct {
		name             string
		template         string
		expectedPriority int
	}{
		{
			name:             "PII SSN has priority 100",
			template:         "pii",
			expectedPriority: 100,
		},
		{
			name:             "PCI card has priority 100",
			template:         "pci",
			expectedPriority: 100,
		},
		{
			name:             "HIPAA MRN has priority 100",
			template:         "hipaa",
			expectedPriority: 100,
		},
		{
			name:             "GDPR IP has priority 90",
			template:         "gdpr",
			expectedPriority: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := DefaultTemplates[tt.template]
			require.NotEmpty(t, patterns)

			// Check that at least one pattern has the expected priority
			foundPriority := false
			for _, p := range patterns {
				if p.Priority == tt.expectedPriority {
					foundPriority = true
					break
				}
			}
			assert.True(t, foundPriority, "Template should have at least one pattern with priority %d", tt.expectedPriority)
		})
	}
}

func TestSanitizer_LoadTemplate_Integration(t *testing.T) {
	// Integration test: Load template and verify sanitization works
	sanitizer := NewSanitizer()
	err := sanitizer.LoadTemplate("pii")
	require.NoError(t, err)

	input := `{
		"password": "secret123",
		"ssn": "123-45-6789",
		"email": "user@example.com",
		"token": "abc123xyz"
	}`

	result := sanitizer.SanitizeString(input)

	// Verify all sensitive data is redacted
	assert.Contains(t, result, "[REDACTED]")       // password (default pattern)
	assert.Contains(t, result, "[REDACTED_SSN]")   // SSN (PII template)
	assert.Contains(t, result, "[REDACTED_EMAIL]") // email (PII template)
	assert.NotContains(t, result, "secret123")
	assert.NotContains(t, result, "123-45-6789")
	assert.NotContains(t, result, "user@example.com")
}

func TestSanitizer_LoadTemplates_Integration(t *testing.T) {
	// Integration test: Load multiple templates and verify all work
	sanitizer := NewSanitizer()
	err := sanitizer.LoadTemplates("pii", "pci")
	require.NoError(t, err)

	input := `{
		"ssn": "123-45-6789",
		"card_number": "4532 1234 5678 9010",
		"cvv": "123",
		"email": "user@example.com"
	}`

	result := sanitizer.SanitizeString(input)

	// Verify all patterns from both templates work
	assert.Contains(t, result, "[REDACTED_SSN]")
	assert.Contains(t, result, "[REDACTED_CARD]")
	assert.Contains(t, result, "[REDACTED_CVV]")
	assert.Contains(t, result, "[REDACTED_EMAIL]")
	assert.NotContains(t, result, "123-45-6789")
	assert.NotContains(t, result, "4532 1234 5678 9010")
	assert.NotContains(t, result, "user@example.com")
}

func TestTemplateRegistry_ThreadSafety(t *testing.T) {
	// Test concurrent access to template registry
	const goroutines = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // readers + writers

	// Concurrent readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				names := GetTemplateNames()
				assert.NotEmpty(t, names)
			}
		}()
	}

	// Concurrent writers (register custom templates)
	for i := 0; i < goroutines; i++ {
		i := i // capture loop variable
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tmplName := fmt.Sprintf("concurrent_test_%d_%d", i, j)
				patterns := []SanitizePattern{
					{
						Pattern:     regexp.MustCompile(`test`),
						Replacement: "[TEST]",
						Priority:    1,
						Name:        "test",
					},
				}
				_ = RegisterTemplate(tmplName, patterns)
			}
		}()
	}

	wg.Wait()

	// Clean up registered templates
	templateMu.Lock()
	for name := range DefaultTemplates {
		if len(name) > 16 && name[:16] == "concurrent_test_" {
			delete(DefaultTemplates, name)
		}
	}
	templateMu.Unlock()
}

func TestTemplatePatternNames(t *testing.T) {
	// Verify all patterns have meaningful names
	for templateName, patterns := range DefaultTemplates {
		t.Run(templateName, func(t *testing.T) {
			for _, pattern := range patterns {
				assert.NotEmpty(t, pattern.Name, "Pattern in %s template should have a name", templateName)
				assert.NotContains(t, pattern.Name, "pattern_", "Pattern should have meaningful name, not auto-generated")
			}
		})
	}
}
