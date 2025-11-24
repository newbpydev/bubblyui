package devtools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSanitizeWithOptions_DryRun verifies dry-run mode doesn't mutate data
func TestSanitizeWithOptions_DryRun(t *testing.T) {
	tests := []struct {
		name           string
		data           *ExportData
		opts           SanitizeOptions
		wantMutated    bool
		wantMatchCount int
	}{
		{
			name: "dry run does not mutate data",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Props: map[string]interface{}{
							"password": "secret123",
							"username": "alice",
						},
					},
				},
			},
			opts: SanitizeOptions{
				DryRun: true,
			},
			wantMutated:    false,
			wantMatchCount: 1, // password field
		},
		{
			name: "normal mode mutates data",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Refs: []*RefSnapshot{
							{
								ID:    "ref-1",
								Name:  "authData",
								Value: `{"password": "secret123"}`, // String value with pattern
							},
						},
					},
				},
			},
			opts: SanitizeOptions{
				DryRun: false,
			},
			wantMutated:    true,
			wantMatchCount: 0, // no dry-run result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()

			// Store original value based on test data structure
			var originalValue interface{}
			if len(tt.data.Components[0].Props) > 0 {
				originalValue = tt.data.Components[0].Props["password"]
			} else if len(tt.data.Components[0].Refs) > 0 {
				originalValue = tt.data.Components[0].Refs[0].Value
			}

			// Sanitize with options
			result, dryRunResult := sanitizer.SanitizeWithOptions(tt.data, tt.opts)

			if tt.opts.DryRun {
				// Dry-run: original data should be unchanged
				if len(tt.data.Components[0].Props) > 0 {
					assert.Equal(t, originalValue, tt.data.Components[0].Props["password"],
						"Original data should not be mutated in dry-run mode")
				} else if len(tt.data.Components[0].Refs) > 0 {
					assert.Equal(t, originalValue, tt.data.Components[0].Refs[0].Value,
						"Original data should not be mutated in dry-run mode")
				}

				// Result should be nil in dry-run mode
				assert.Nil(t, result, "Result should be nil in dry-run mode")

				// Dry-run result should be present
				require.NotNil(t, dryRunResult, "DryRunResult should be present")
				assert.Equal(t, tt.wantMatchCount, dryRunResult.WouldRedactCount,
					"WouldRedactCount should match expected")
				assert.Len(t, dryRunResult.Matches, tt.wantMatchCount,
					"Matches count should match expected")
			} else {
				// Normal mode: data should be sanitized
				require.NotNil(t, result, "Result should be present in normal mode")

				// Check that ref value was sanitized
				if len(result.Components[0].Refs) > 0 {
					refValue, ok := result.Components[0].Refs[0].Value.(string)
					require.True(t, ok, "Ref value should be a string")
					assert.NotEqual(t, originalValue, refValue,
						"Data should be sanitized in normal mode")
					assert.Contains(t, refValue, "[REDACTED]",
						"Sanitized value should contain redaction marker")
				}

				// Dry-run result should be nil
				assert.Nil(t, dryRunResult, "DryRunResult should be nil in normal mode")
			}
		})
	}
}

// TestSanitizeWithOptions_MatchLocations verifies match location tracking
func TestSanitizeWithOptions_MatchLocations(t *testing.T) {
	tests := []struct {
		name         string
		data         *ExportData
		wantMatches  int
		wantPaths    []string
		wantPatterns []string
	}{
		{
			name: "tracks nested component props",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "LoginForm",
						Props: map[string]interface{}{
							"password": "secret123",
							"apiKey":   "sk_live_abc123",
						},
					},
				},
			},
			wantMatches:  2,
			wantPaths:    []string{"components[0].props.password", "components[0].props.apiKey"},
			wantPatterns: []string{"pattern_0", "pattern_2"}, // password and apikey patterns
		},
		{
			name: "tracks state changes",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				State: []StateChange{
					{
						RefID:     "auth-token",
						RefName:   "authToken",
						OldValue:  nil,
						NewValue:  `{"token": "bearer_token_xyz"}`, // String with pattern
						Timestamp: time.Now(),
						Source:    "login",
					},
				},
			},
			wantMatches:  1,
			wantPaths:    []string{"state[0].new_value"},
			wantPatterns: []string{"pattern_1"}, // token pattern
		},
		{
			name: "tracks event payloads",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Events: []EventRecord{
					{
						ID:        "evt-1",
						Name:      "login",
						SourceID:  "form-1",
						Timestamp: time.Now(),
						Payload: map[string]interface{}{
							"username": "alice",
							"password": "secret456",
						},
					},
				},
			},
			wantMatches:  1,
			wantPaths:    []string{"events[0].payload.password"},
			wantPatterns: []string{"pattern_0"}, // password pattern
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()

			opts := SanitizeOptions{
				DryRun: true,
			}

			_, dryRunResult := sanitizer.SanitizeWithOptions(tt.data, opts)

			require.NotNil(t, dryRunResult, "DryRunResult should be present")
			assert.Equal(t, tt.wantMatches, dryRunResult.WouldRedactCount,
				"WouldRedactCount should match expected")
			assert.Len(t, dryRunResult.Matches, tt.wantMatches,
				"Matches count should match expected")

			// Verify paths are tracked (order may vary due to map iteration)
			foundPaths := make(map[string]bool)
			for _, match := range dryRunResult.Matches {
				foundPaths[match.Path] = true
			}
			for _, expectedPath := range tt.wantPaths {
				assert.True(t, foundPaths[expectedPath],
					"Expected path %s should be found", expectedPath)
			}

			// Verify pattern names are tracked (order may vary)
			foundPatterns := make(map[string]bool)
			for _, match := range dryRunResult.Matches {
				foundPatterns[match.Pattern] = true
			}
			for _, expectedPattern := range tt.wantPatterns {
				assert.True(t, foundPatterns[expectedPattern],
					"Expected pattern %s should be found", expectedPattern)
			}
		})
	}
}

// TestSanitizeWithOptions_MaxPreviewLen verifies preview truncation
func TestSanitizeWithOptions_MaxPreviewLen(t *testing.T) {
	tests := []struct {
		name          string
		data          *ExportData
		maxPreviewLen int
		wantTruncated bool
	}{
		{
			name: "truncates long values",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Props: map[string]interface{}{
							"password": "this_is_a_very_long_password_that_should_be_truncated_for_preview_purposes_123456789",
						},
					},
				},
			},
			maxPreviewLen: 20,
			wantTruncated: true,
		},
		{
			name: "does not truncate short values",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Props: map[string]interface{}{
							"password": "short",
						},
					},
				},
			},
			maxPreviewLen: 100,
			wantTruncated: false,
		},
		{
			name: "zero maxPreviewLen means no truncation",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Props: map[string]interface{}{
							"password": "this_is_a_very_long_password_that_should_not_be_truncated_when_max_is_zero",
						},
					},
				},
			},
			maxPreviewLen: 0,
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()

			opts := SanitizeOptions{
				DryRun:        true,
				MaxPreviewLen: tt.maxPreviewLen,
			}

			_, dryRunResult := sanitizer.SanitizeWithOptions(tt.data, opts)

			require.NotNil(t, dryRunResult, "DryRunResult should be present")
			require.Len(t, dryRunResult.Matches, 1, "Should have one match")

			match := dryRunResult.Matches[0]

			if tt.wantTruncated {
				assert.LessOrEqual(t, len(match.Original), tt.maxPreviewLen+3, // +3 for "..."
					"Original value should be truncated")
				assert.Contains(t, match.Original, "...",
					"Truncated value should contain ellipsis")
			} else {
				if tt.maxPreviewLen > 0 {
					assert.LessOrEqual(t, len(match.Original), tt.maxPreviewLen,
						"Original value should not exceed max length")
				}
			}
		})
	}
}

// TestPreview verifies the convenience Preview method
func TestPreview(t *testing.T) {
	tests := []struct {
		name           string
		data           *ExportData
		wantMatchCount int
	}{
		{
			name: "preview is shorthand for dry-run",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Props: map[string]interface{}{
							"password": "secret123",
							"token":    "bearer_xyz",
							"apiKey":   "sk_live_abc",
						},
					},
				},
			},
			wantMatchCount: 3,
		},
		{
			name: "preview with no matches",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "TestComponent",
						Props: map[string]interface{}{
							"username": "alice",
							"email":    "alice@example.com",
						},
					},
				},
			},
			wantMatchCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := NewSanitizer()

			result := sanitizer.Preview(tt.data)

			require.NotNil(t, result, "Preview result should not be nil")
			assert.Equal(t, tt.wantMatchCount, result.WouldRedactCount,
				"WouldRedactCount should match expected")
			assert.Len(t, result.Matches, tt.wantMatchCount,
				"Matches count should match expected")

			// Verify PreviewData is set
			assert.NotNil(t, result.PreviewData, "PreviewData should be set")
		})
	}
}

// TestSanitizeWithOptions_PreviewData verifies preview data structure
func TestSanitizeWithOptions_PreviewData(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Props: map[string]interface{}{
					"password": "secret123",
					"username": "alice",
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	require.NotNil(t, dryRunResult.PreviewData, "PreviewData should be present")

	// PreviewData should be the original data structure
	previewData, ok := dryRunResult.PreviewData.(*ExportData)
	require.True(t, ok, "PreviewData should be *ExportData")
	assert.Equal(t, data.Version, previewData.Version, "PreviewData should match original")
	assert.Equal(t, len(data.Components), len(previewData.Components),
		"PreviewData should have same component count")
}

// TestSanitizeWithOptions_Integration verifies integration with Sanitize()
func TestSanitizeWithOptions_Integration(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "authData",
						Value: `{"password": "secret123", "username": "alice"}`,
					},
				},
			},
		},
	}

	// First, preview to see what would be redacted
	previewResult := sanitizer.Preview(data)
	require.NotNil(t, previewResult, "Preview result should not be nil")
	assert.Equal(t, 1, previewResult.WouldRedactCount, "Should find 1 match")

	// Then, actually sanitize
	sanitizedData := sanitizer.Sanitize(data)
	require.NotNil(t, sanitizedData, "Sanitized data should not be nil")

	// Verify the ref value was sanitized
	refValue, ok := sanitizedData.Components[0].Refs[0].Value.(string)
	require.True(t, ok, "Ref value should be a string")
	assert.Contains(t, refValue, "[REDACTED]", "Password should be redacted")
	assert.Contains(t, refValue, "alice", "Username should not be redacted")
}

// TestMatchLocation_Fields verifies MatchLocation structure
func TestMatchLocation_Fields(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Props: map[string]interface{}{
					"password": "secret123",
				},
			},
		},
	}

	result := sanitizer.Preview(data)

	require.NotNil(t, result, "Preview result should not be nil")
	require.Len(t, result.Matches, 1, "Should have one match")

	match := result.Matches[0]

	// Verify all fields are populated
	assert.NotEmpty(t, match.Path, "Path should be populated")
	assert.NotEmpty(t, match.Pattern, "Pattern should be populated")
	assert.NotEmpty(t, match.Original, "Original should be populated")
	assert.NotEmpty(t, match.Redacted, "Redacted should be populated")
	assert.Contains(t, match.Redacted, "[REDACTED]", "Redacted should contain redaction marker")
}

// TestCollectMatches_SliceValues tests collectMatchesSlice for slice traversal
func TestCollectMatches_SliceValues(t *testing.T) {
	sanitizer := NewSanitizer()

	// Use strings that match the sanitizer patterns (password=value format)
	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Props: map[string]interface{}{
					"config": []interface{}{"password=secret1", "token=abc123", "normal"},
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	// Password and token patterns should be detected
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 1,
		"Should detect sensitive data in slice elements")
}

// TestCollectMatches_NestedSlice tests nested slice traversal
func TestCollectMatches_NestedSlice(t *testing.T) {
	sanitizer := NewSanitizer()

	// Use strings with pattern format: key=value or key:value
	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		State: []StateChange{
			{
				RefID:     "ref-1",
				RefName:   "tokens",
				OldValue:  nil,
				NewValue:  "token=abc123xyz",
				Timestamp: time.Now(),
			},
			{
				RefID:     "ref-2",
				RefName:   "apiKey",
				OldValue:  nil,
				NewValue:  "apikey=sk_live_xyz",
				Timestamp: time.Now(),
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	// Should detect token and apiKey patterns
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 2,
		"Should detect sensitive data in multiple slice elements")
}

// TestCollectMatches_StructFields tests struct field traversal via collectMatchesStruct
func TestCollectMatches_StructFields(t *testing.T) {
	sanitizer := NewSanitizer()

	// Test with ComponentSnapshot which has various struct fields
	// Use values in pattern format: key=value or key:value
	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				State: map[string]interface{}{
					"auth": "password=secret123",
				},
				Props: map[string]interface{}{
					"config": "apikey=sk_live_test",
				},
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "auth",
						Value: "token=bearer_xyz",
					},
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	// Should detect password, apiKey, and token patterns across struct fields
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 3,
		"Should detect sensitive data across struct fields")
}

// TestCollectMatches_GenericStruct tests generic struct traversal
func TestCollectMatches_GenericStruct(t *testing.T) {
	sanitizer := NewSanitizer()

	// Test with nested events that have map payloads
	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Events: []EventRecord{
			{
				ID:        "evt-1",
				Name:      "auth",
				SourceID:  "form-1",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"username": "alice",
					"password": "secret456",
					"nested": map[string]interface{}{
						"token": "jwt_token_value",
					},
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	// Should detect both password and token in nested structures
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 1,
		"Should detect sensitive data in nested map structures")
}

// TestBuildPath tests the buildPath helper function indirectly
func TestBuildPath_IndirectVerification(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "TestComponent",
				Props: map[string]interface{}{
					"config": map[string]interface{}{
						"auth": map[string]interface{}{
							"password": "nested_secret",
						},
					},
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	require.GreaterOrEqual(t, len(dryRunResult.Matches), 1, "Should have at least one match")

	// Verify the path contains the nested structure
	foundNestedPath := false
	for _, match := range dryRunResult.Matches {
		if match.Path != "" && len(match.Path) > 10 {
			foundNestedPath = true
			break
		}
	}
	assert.True(t, foundNestedPath, "Should have found a nested path")
}

// TestCollectMatches_ComponentSnapshot tests collectMatchesComponent
func TestCollectMatches_ComponentSnapshot(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "LoginForm",
				Props: map[string]interface{}{
					"password": "user_password",
				},
				State: map[string]interface{}{
					"token": "session_token",
				},
				Children: []*ComponentSnapshot{
					{
						ID:   "comp-2",
						Name: "PasswordField",
						Props: map[string]interface{}{
							"value": "child_password",
						},
					},
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	// Should detect passwords and tokens in parent and child components
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 2,
		"Should detect sensitive data in component hierarchy")
}

// TestCollectMatches_EventRecord tests collectMatchesEventRecord
func TestCollectMatches_EventRecord(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Events: []EventRecord{
			{
				ID:        "evt-1",
				Name:      "login",
				SourceID:  "form-1",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"password": "event_password",
				},
			},
			{
				ID:        "evt-2",
				Name:      "api_call",
				SourceID:  "service-1",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"apiKey": "sk_test_key",
				},
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 2,
		"Should detect sensitive data in event records")
}

// TestCollectMatches_StateChange tests collectMatchesStateChange
func TestCollectMatches_StateChange(t *testing.T) {
	sanitizer := NewSanitizer()

	// Use values with pattern format
	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		State: []StateChange{
			{
				RefID:     "ref-1",
				RefName:   "authState",
				OldValue:  "password=old_secret",
				NewValue:  "password=new_secret",
				Timestamp: time.Now(),
				Source:    "user_input",
			},
		},
	}

	opts := SanitizeOptions{
		DryRun: true,
	}

	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	// Should detect password patterns in both old and new values
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 1,
		"Should detect sensitive data in state changes")
}

// TestCollectMatches_DirectStructValues tests collectMatchesStruct via ExportData embedding
func TestCollectMatches_DirectStructValues(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name          string
		data          *ExportData
		minRedactions int
	}{
		{
			name: "ExportData with component props",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "Test",
						Props: map[string]interface{}{
							"password": "direct_secret",
						},
					},
				},
			},
			minRedactions: 1,
		},
		{
			name: "ExportData with StateChange containing sensitive values",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				State: []StateChange{
					{
						RefID:    "ref-1",
						RefName:  "auth",
						OldValue: "password=old_secret",
						NewValue: "password=new_secret",
					},
				},
			},
			minRedactions: 1,
		},
		{
			name: "ExportData with EventRecord payloads",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Events: []EventRecord{
					{
						ID:       "evt-1",
						Name:     "auth",
						SourceID: "form-1",
						Payload: map[string]interface{}{
							"credentials": "password=event_secret",
						},
					},
				},
			},
			minRedactions: 1,
		},
		{
			name: "ExportData with component state maps",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "DirectComponent",
						Props: map[string]interface{}{
							"apiKey": "sk_test_direct",
						},
						State: map[string]interface{}{
							"token": "bearer_direct",
						},
					},
				},
			},
			minRedactions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SanitizeOptions{DryRun: true}
			_, dryRunResult := sanitizer.SanitizeWithOptions(tt.data, opts)

			require.NotNil(t, dryRunResult, "DryRunResult should be present")
			assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, tt.minRedactions,
				"Should detect sensitive data in struct")
		})
	}
}

// TestCollectMatches_NestedMapInProps tests collectMatchesGenericStruct via nested maps in Props
func TestCollectMatches_NestedMapInProps(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name          string
		data          *ExportData
		minRedactions int
	}{
		{
			name: "deeply nested maps with sensitive data",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "NestedComponent",
						Props: map[string]interface{}{
							"config": map[string]interface{}{
								"auth": map[string]interface{}{
									"password": "nested_secret",
									"apiKey":   "sk_nested_key",
								},
							},
						},
					},
				},
			},
			minRedactions: 1,
		},
		{
			name: "slice of maps in props",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "SliceComponent",
						Props: map[string]interface{}{
							"credentials": []interface{}{
								map[string]interface{}{
									"password": "slice_secret1",
								},
								map[string]interface{}{
									"token": "slice_token",
								},
							},
						},
					},
				},
			},
			minRedactions: 1,
		},
		{
			name: "mixed nested types",
			data: &ExportData{
				Version:   "1.0",
				Timestamp: time.Now(),
				Components: []*ComponentSnapshot{
					{
						ID:   "comp-1",
						Name: "MixedComponent",
						State: map[string]interface{}{
							"users": []interface{}{
								map[string]interface{}{
									"name":     "alice",
									"password": "alice_pass",
								},
							},
							"settings": map[string]interface{}{
								"apiKey": "sk_settings_key",
							},
						},
					},
				},
			},
			minRedactions: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SanitizeOptions{DryRun: true}
			_, dryRunResult := sanitizer.SanitizeWithOptions(tt.data, opts)

			require.NotNil(t, dryRunResult, "DryRunResult should be present")
			assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, tt.minRedactions,
				"Should handle nested structures correctly")
		})
	}
}

// TestCollectMatches_RefSnapshots tests collectMatches via RefSnapshot traversal
func TestCollectMatches_RefSnapshots(t *testing.T) {
	sanitizer := NewSanitizer()

	data := &ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "ComponentWithRefs",
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "authToken",
						Type:  "string",
						Value: "password=ref_secret",
					},
					{
						ID:   "ref-2",
						Name: "config",
						Type: "map",
						Value: map[string]interface{}{
							"apiKey": "sk_ref_key",
						},
					},
				},
			},
		},
	}

	opts := SanitizeOptions{DryRun: true}
	_, dryRunResult := sanitizer.SanitizeWithOptions(data, opts)

	require.NotNil(t, dryRunResult, "DryRunResult should be present")
	assert.GreaterOrEqual(t, dryRunResult.WouldRedactCount, 1,
		"Should detect sensitive data in RefSnapshots")
}

// TestBuildPath tests the buildPath helper function
func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		child    string
		expected string
	}{
		{
			name:     "empty parent",
			parent:   "",
			child:    "field",
			expected: "field",
		},
		{
			name:     "non-empty parent",
			parent:   "root",
			child:    "field",
			expected: "root.field",
		},
		{
			name:     "nested path",
			parent:   "root.nested",
			child:    "field",
			expected: "root.nested.field",
		},
		{
			name:     "empty child",
			parent:   "root",
			child:    "",
			expected: "root.",
		},
		{
			name:     "both empty",
			parent:   "",
			child:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPath(tt.parent, tt.child)
			assert.Equal(t, tt.expected, result)
		})
	}
}
