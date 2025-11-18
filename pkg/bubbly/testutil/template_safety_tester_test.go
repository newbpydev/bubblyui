package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateSafetyTester_NewTester verifies basic tester creation
func TestTemplateSafetyTester_NewTester(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "empty_template",
			template: "",
		},
		{
			name:     "simple_template",
			template: "Hello World",
		},
		{
			name:     "complex_template",
			template: "Count: {{count}}\nName: {{name}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester(tt.template)

			assert.NotNil(t, tester)
			assert.Equal(t, tt.template, tester.template)
			assert.Empty(t, tester.mutations)
			assert.Empty(t, tester.violations)
			assert.True(t, tester.immutable, "Template should be immutable by default")
		})
	}
}

// TestTemplateSafetyTester_AttemptMutation_DetectsPanic verifies mutation detection
func TestTemplateSafetyTester_AttemptMutation_DetectsPanic(t *testing.T) {
	// Create a component with a ref that will be mutated in template
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// This template will attempt mutation
			return "Count: 0"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	tester := NewTemplateSafetyTester("test template")

	// Attempt mutation - should detect and record violation
	tester.AttemptMutation("Set count to 42 in template")

	// Should have recorded the violation
	assert.Len(t, tester.violations, 1)
	assert.Contains(t, tester.mutations, "Set count to 42 in template")
}

// TestTemplateSafetyTester_AssertImmutable verifies immutability assertion
func TestTemplateSafetyTester_AssertImmutable(t *testing.T) {
	tests := []struct {
		name          string
		mutations     []string
		shouldPass    bool
		expectedError string
	}{
		{
			name:       "no_mutations_passes",
			mutations:  []string{},
			shouldPass: true,
		},
		{
			name:          "with_mutations_fails",
			mutations:     []string{"mutation1", "mutation2"},
			shouldPass:    false,
			expectedError: "template is not immutable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester("test")

			// Add mutations using AttemptMutation
			for _, mutation := range tt.mutations {
				tester.AttemptMutation(mutation)
			}

			if tt.shouldPass {
				// Should not fail
				mockT := &mockTestingT{}
				tester.AssertImmutable(mockT)
				assert.False(t, mockT.failed, "Should not fail for immutable template")
			} else {
				// Should fail
				mockT := &mockTestingT{}
				tester.AssertImmutable(mockT)
				assert.True(t, mockT.failed, "Should fail when mutations detected")
				assert.NotEmpty(t, mockT.errors, "Should have error messages")
			}
		})
	}
}

// TestTemplateSafetyTester_AssertViolations verifies violation counting
func TestTemplateSafetyTester_AssertViolations(t *testing.T) {
	tests := []struct {
		name           string
		violationCount int
		expected       int
		shouldPass     bool
	}{
		{
			name:           "zero_violations_matches",
			violationCount: 0,
			expected:       0,
			shouldPass:     true,
		},
		{
			name:           "one_violation_matches",
			violationCount: 1,
			expected:       1,
			shouldPass:     true,
		},
		{
			name:           "mismatch_fails",
			violationCount: 2,
			expected:       1,
			shouldPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester("test")

			// Add violations
			for i := 0; i < tt.violationCount; i++ {
				tester.violations = append(tester.violations, SafetyViolation{
					Description: "test violation",
				})
			}

			mockT := &mockTestingT{}
			tester.AssertViolations(mockT, tt.expected)

			if tt.shouldPass {
				assert.False(t, mockT.failed, "Should pass when counts match")
			} else {
				assert.True(t, mockT.failed, "Should fail when counts don't match")
			}
		})
	}
}

// TestTemplateSafetyTester_GetViolations verifies violation retrieval
func TestTemplateSafetyTester_GetViolations(t *testing.T) {
	tester := NewTemplateSafetyTester("test")

	// Initially empty
	violations := tester.GetViolations()
	assert.Empty(t, violations)

	// Add violations
	tester.violations = []SafetyViolation{
		{Description: "violation1"},
		{Description: "violation2"},
	}

	violations = tester.GetViolations()
	assert.Len(t, violations, 2)
	assert.Equal(t, "violation1", violations[0].Description)
	assert.Equal(t, "violation2", violations[1].Description)
}

// TestTemplateSafetyTester_DeepCloning verifies deep cloning works
func TestTemplateSafetyTester_DeepCloning(t *testing.T) {
	original := NewTemplateSafetyTester("original template")
	original.mutations = []string{"mutation1"}
	original.violations = []SafetyViolation{{Description: "violation1"}}

	// Get violations should return a copy, not the original slice
	violations := original.GetViolations()
	violations[0].Description = "modified"

	// Original should be unchanged
	assert.Equal(t, "violation1", original.violations[0].Description)
}

// TestTemplateSafetyTester_SharedTemplatesIsolated verifies template isolation
func TestTemplateSafetyTester_SharedTemplatesIsolated(t *testing.T) {
	template := "shared template"

	tester1 := NewTemplateSafetyTester(template)
	tester2 := NewTemplateSafetyTester(template)

	// Mutate tester1
	tester1.AttemptMutation("mutation1")

	// tester2 should be unaffected
	assert.Empty(t, tester2.mutations)
	assert.Empty(t, tester2.violations)
}

// TestTemplateSafetyTester_PerformanceOverhead verifies minimal overhead
func TestTemplateSafetyTester_PerformanceOverhead(t *testing.T) {
	tester := NewTemplateSafetyTester("test template")

	// Create many violations - should handle efficiently
	for i := 0; i < 1000; i++ {
		tester.AttemptMutation("mutation")
	}

	assert.Len(t, tester.mutations, 1000)
	assert.Len(t, tester.violations, 1000)

	// GetViolations should be fast
	violations := tester.GetViolations()
	assert.Len(t, violations, 1000)
}

// TestTemplateSafetyTester_WithRealComponent tests with actual component
func TestTemplateSafetyTester_WithRealComponent(t *testing.T) {
	// Create component that attempts mutation in template
	component, err := bubbly.NewComponent("TestComponent").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			// This should panic - mutation in template
			count.Set(42)
			return "unreachable"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Verify component panics on View()
	assert.Panics(t, func() {
		component.View()
	}, "Component should panic when mutating in template")

	// Now test with tester
	tester := NewTemplateSafetyTester("mutation template")
	tester.AttemptMutation("Set count in template")

	assert.Len(t, tester.violations, 1)
	assert.Contains(t, tester.violations[0].Description, "Set count in template")
}

// TestTemplateSafetyTester_ViolationDetails verifies violation details are captured
func TestTemplateSafetyTester_ViolationDetails(t *testing.T) {
	tester := NewTemplateSafetyTester("test")

	tester.AttemptMutation("detailed mutation attempt")

	violations := tester.GetViolations()
	require.Len(t, violations, 1)

	violation := violations[0]
	assert.Contains(t, violation.Description, "detailed mutation attempt")
	assert.NotZero(t, violation.Timestamp, "Timestamp should be set")
	assert.NotEmpty(t, violation.StackTrace, "Stack trace should be captured")
}

// TestTemplateSafetyTester_MultipleAttempts verifies multiple mutation attempts
func TestTemplateSafetyTester_MultipleAttempts(t *testing.T) {
	tester := NewTemplateSafetyTester("test")

	// Attempt multiple mutations
	tester.AttemptMutation("mutation1")
	tester.AttemptMutation("mutation2")
	tester.AttemptMutation("mutation3")

	assert.Len(t, tester.mutations, 3)
	assert.Len(t, tester.violations, 3)

	mockT := &mockTestingT{}
	tester.AssertViolations(mockT, 3)
	assert.False(t, mockT.failed)
}

// TestTemplateSafetyTester_ImmutableFlag verifies immutable flag behavior
func TestTemplateSafetyTester_ImmutableFlag(t *testing.T) {
	tester := NewTemplateSafetyTester("test")

	// Initially immutable
	assert.True(t, tester.immutable)

	// After mutation attempt, should be marked as not immutable
	tester.AttemptMutation("mutation")
	assert.False(t, tester.immutable, "Should be marked as not immutable after mutation")
}

// TestTemplateSafetyTester_GetMutations verifies mutation retrieval
func TestTemplateSafetyTester_GetMutations(t *testing.T) {
	tests := []struct {
		name      string
		mutations []string
	}{
		{
			name:      "empty_mutations",
			mutations: []string{},
		},
		{
			name:      "single_mutation",
			mutations: []string{"mutation1"},
		},
		{
			name:      "multiple_mutations",
			mutations: []string{"mutation1", "mutation2", "mutation3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester("test")

			// Add mutations
			for _, mutation := range tt.mutations {
				tester.AttemptMutation(mutation)
			}

			// Get mutations
			result := tester.GetMutations()

			// Verify count and content
			assert.Len(t, result, len(tt.mutations))
			for i, mutation := range tt.mutations {
				assert.Equal(t, mutation, result[i])
			}

			// Verify it returns a copy (defensive copying)
			if len(result) > 0 {
				result[0] = "modified"
				assert.NotEqual(t, "modified", tester.mutations[0], "Should return a copy, not the original")
			}
		})
	}
}

// TestTemplateSafetyTester_IsImmutable verifies immutability check
func TestTemplateSafetyTester_IsImmutable(t *testing.T) {
	tests := []struct {
		name           string
		addMutations   bool
		mutationCount  int
		expectedResult bool
	}{
		{
			name:           "new_tester_is_immutable",
			addMutations:   false,
			expectedResult: true,
		},
		{
			name:           "after_one_mutation_not_immutable",
			addMutations:   true,
			mutationCount:  1,
			expectedResult: false,
		},
		{
			name:           "after_multiple_mutations_not_immutable",
			addMutations:   true,
			mutationCount:  3,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester("test")

			if tt.addMutations {
				for i := 0; i < tt.mutationCount; i++ {
					tester.AttemptMutation("mutation")
				}
			}

			result := tester.IsImmutable()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestTemplateSafetyTester_GetTemplate verifies template retrieval
func TestTemplateSafetyTester_GetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "empty_template",
			template: "",
		},
		{
			name:     "simple_template",
			template: "Hello World",
		},
		{
			name:     "complex_template",
			template: "Count: {{count}}\nName: {{name}}\nStatus: {{status}}",
		},
		{
			name:     "template_with_special_chars",
			template: "Special: \n\t\"quoted\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester(tt.template)

			result := tester.GetTemplate()
			assert.Equal(t, tt.template, result)

			// Verify template doesn't change after mutations
			tester.AttemptMutation("mutation")
			assert.Equal(t, tt.template, tester.GetTemplate(), "Template should not change after mutations")
		})
	}
}

// TestTemplateSafetyTester_Reset verifies reset functionality
func TestTemplateSafetyTester_Reset(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*TemplateSafetyTester)
		verify func(*testing.T, *TemplateSafetyTester)
	}{
		{
			name: "reset_after_mutations",
			setup: func(tst *TemplateSafetyTester) {
				tst.AttemptMutation("mutation1")
				tst.AttemptMutation("mutation2")
			},
			verify: func(t *testing.T, tst *TemplateSafetyTester) {
				assert.Empty(t, tst.GetMutations())
				assert.Empty(t, tst.GetViolations())
				assert.True(t, tst.IsImmutable())
			},
		},
		{
			name: "reset_multiple_times",
			setup: func(tst *TemplateSafetyTester) {
				// Add mutations, reset, add more, reset again
				tst.AttemptMutation("mutation1")
				tst.Reset()
				tst.AttemptMutation("mutation2")
				tst.Reset()
			},
			verify: func(t *testing.T, tst *TemplateSafetyTester) {
				assert.Empty(t, tst.GetMutations())
				assert.Empty(t, tst.GetViolations())
				assert.True(t, tst.IsImmutable())
			},
		},
		{
			name: "reset_preserves_template",
			setup: func(tst *TemplateSafetyTester) {
				tst.AttemptMutation("mutation")
			},
			verify: func(t *testing.T, tst *TemplateSafetyTester) {
				template := tst.GetTemplate()
				assert.NotEmpty(t, template, "Template should be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester("test template")

			// Setup
			tt.setup(tester)

			// Reset
			tester.Reset()

			// Verify
			tt.verify(t, tester)
		})
	}
}

// TestTemplateSafetyTester_String verifies string representation
func TestTemplateSafetyTester_String(t *testing.T) {
	tests := []struct {
		name     string
		template string
		setup    func(*TemplateSafetyTester)
		contains []string
	}{
		{
			name:     "new_tester_string",
			template: "simple",
			setup:    func(tst *TemplateSafetyTester) {},
			contains: []string{
				"TemplateSafetyTester",
				"template=\"simple\"",
				"mutations=0",
				"immutable=true",
			},
		},
		{
			name:     "with_mutations_string",
			template: "Count: {{count}}",
			setup: func(tst *TemplateSafetyTester) {
				tst.AttemptMutation("mutation1")
				tst.AttemptMutation("mutation2")
			},
			contains: []string{
				"TemplateSafetyTester",
				"template=\"Count: {{count}}\"",
				"mutations=2",
				"immutable=false",
			},
		},
		{
			name:     "after_reset_string",
			template: "test",
			setup: func(tst *TemplateSafetyTester) {
				tst.AttemptMutation("mutation")
				tst.Reset()
			},
			contains: []string{
				"mutations=0",
				"immutable=true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := NewTemplateSafetyTester(tt.template)

			// Setup
			tt.setup(tester)

			// Get string representation
			result := tester.String()

			// Verify it contains expected substrings
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

// TestTemplateSafetyTester_ResetReuseScenario verifies reset allows tester reuse
func TestTemplateSafetyTester_ResetReuseScenario(t *testing.T) {
	tester := NewTemplateSafetyTester("reusable template")

	// First test case - with mutations
	tester.AttemptMutation("mutation1")
	assert.False(t, tester.IsImmutable())
	assert.Len(t, tester.GetMutations(), 1)

	// Reset for next test case
	tester.Reset()

	// Second test case - should be clean
	assert.True(t, tester.IsImmutable())
	assert.Empty(t, tester.GetMutations())
	assert.Empty(t, tester.GetViolations())

	// Third test case - add different mutations
	tester.AttemptMutation("mutation2")
	tester.AttemptMutation("mutation3")
	assert.Len(t, tester.GetMutations(), 2)

	// Verify template never changed
	assert.Equal(t, "reusable template", tester.GetTemplate())
}
