package testutil

import (
	"fmt"
	"runtime/debug"
	"time"
)

// SafetyViolation represents a detected template safety violation.
//
// It captures details about an attempted mutation during template rendering,
// including what was attempted, when it occurred, and the stack trace for debugging.
//
// Type Safety:
//   - Immutable once created
//   - Thread-safe for read operations
//   - Contains debugging information for violation analysis
//
// Example:
//
//	violation := SafetyViolation{
//	    Description: "Attempted Ref.Set() in template",
//	    Timestamp:   time.Now(),
//	    StackTrace:  string(debug.Stack()),
//	}
type SafetyViolation struct {
	// Description is a human-readable description of the violation
	Description string

	// Timestamp is when the violation occurred
	Timestamp time.Time

	// StackTrace is the stack trace at the time of violation
	StackTrace string
}

// TemplateSafetyTester provides utilities for testing template mutation prevention and safety checks.
//
// It helps verify that templates remain pure functions (read-only) and that any attempts
// to mutate state during template rendering are properly detected and prevented.
// This tester wraps template strings and tracks mutation attempts, violations, and
// immutability status.
//
// Type Safety:
//   - Thread-safe for single-goroutine test usage
//   - Tracks all mutation attempts with detailed violation records
//   - Provides assertion helpers for test verification
//
// Example:
//
//	func TestTemplateSafety(t *testing.T) {
//	    tester := testutil.NewTemplateSafetyTester("my template")
//
//	    // Attempt a mutation
//	    tester.AttemptMutation("Set count to 42")
//
//	    // Verify immutability
//	    tester.AssertImmutable(t)  // Fails if mutations detected
//
//	    // Check violation count
//	    tester.AssertViolations(t, 1)
//
//	    // Get violation details
//	    violations := tester.GetViolations()
//	    for _, v := range violations {
//	        fmt.Printf("Violation: %s at %v\n", v.Description, v.Timestamp)
//	    }
//	}
type TemplateSafetyTester struct {
	// template is the template string being tested
	template string

	// mutations is a list of mutation attempts (descriptions)
	mutations []string

	// violations is a list of detected safety violations
	violations []SafetyViolation

	// immutable tracks whether the template has remained immutable
	// (true if no mutations have been attempted)
	immutable bool
}

// NewTemplateSafetyTester creates a new TemplateSafetyTester for testing template safety.
//
// The tester starts in an immutable state and tracks all mutation attempts.
// Each mutation attempt is recorded as a violation with timestamp and stack trace.
//
// Parameters:
//   - template: The template string to test (can be empty)
//
// Returns:
//   - *TemplateSafetyTester: A new tester instance
//
// Example:
//
//	tester := testutil.NewTemplateSafetyTester("Count: {{count}}")
//	tester.AttemptMutation("Set count in template")
//	tester.AssertViolations(t, 1)
func NewTemplateSafetyTester(template string) *TemplateSafetyTester {
	return &TemplateSafetyTester{
		template:   template,
		mutations:  []string{},
		violations: []SafetyViolation{},
		immutable:  true, // Starts as immutable
	}
}

// AttemptMutation records a mutation attempt and creates a safety violation.
//
// This method simulates attempting to mutate state during template rendering.
// It records the mutation description, marks the template as not immutable,
// and creates a detailed violation record with timestamp and stack trace.
//
// In real usage, this would be called when testing components that attempt
// to call Ref.Set() or similar mutations inside template functions.
//
// Parameters:
//   - mutation: Description of the mutation attempt
//
// Example:
//
//	tester.AttemptMutation("Set count to 42 in template")
//	tester.AttemptMutation("Call ref.Set() during rendering")
//
//	violations := tester.GetViolations()
//	assert.Len(t, violations, 2)
func (tst *TemplateSafetyTester) AttemptMutation(mutation string) {
	// Record the mutation attempt
	tst.mutations = append(tst.mutations, mutation)

	// Mark as not immutable
	tst.immutable = false

	// Create a violation record with details
	violation := SafetyViolation{
		Description: mutation,
		Timestamp:   time.Now(),
		StackTrace:  string(debug.Stack()),
	}

	tst.violations = append(tst.violations, violation)
}

// AssertImmutable asserts that the template has remained immutable (no mutations attempted).
//
// This method fails the test if any mutations have been recorded via AttemptMutation().
// It's useful for verifying that template safety mechanisms are working correctly.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//
// Example:
//
//	tester := testutil.NewTemplateSafetyTester("template")
//	// ... test code that should not mutate ...
//	tester.AssertImmutable(t)  // Passes if no mutations
//
//	tester.AttemptMutation("mutation")
//	tester.AssertImmutable(t)  // Fails with clear error
func (tst *TemplateSafetyTester) AssertImmutable(t testingT) {
	t.Helper()

	if !tst.immutable {
		t.Errorf("template is not immutable: %d mutation(s) attempted: %v",
			len(tst.mutations), tst.mutations)
	}
}

// AssertViolations asserts that the expected number of violations were detected.
//
// This method verifies that the tester recorded the correct number of safety
// violations. It's useful for testing that mutation attempts are properly detected.
//
// Parameters:
//   - t: The testing.T instance (or mock)
//   - expected: The expected number of violations
//
// Example:
//
//	tester := testutil.NewTemplateSafetyTester("template")
//	tester.AttemptMutation("mutation1")
//	tester.AttemptMutation("mutation2")
//
//	tester.AssertViolations(t, 2)  // Passes
//	tester.AssertViolations(t, 1)  // Fails
func (tst *TemplateSafetyTester) AssertViolations(t testingT, expected int) {
	t.Helper()

	actual := len(tst.violations)
	if actual != expected {
		t.Errorf("expected %d violation(s), got %d", expected, actual)
	}
}

// GetViolations returns a copy of all detected safety violations.
//
// This method returns a deep copy of the violations slice to prevent
// external modification of the tester's internal state. Each violation
// contains the description, timestamp, and stack trace.
//
// Returns:
//   - []SafetyViolation: A copy of all violations
//
// Example:
//
//	violations := tester.GetViolations()
//	for _, v := range violations {
//	    fmt.Printf("Violation: %s\n", v.Description)
//	    fmt.Printf("Time: %v\n", v.Timestamp)
//	    fmt.Printf("Stack:\n%s\n", v.StackTrace)
//	}
func (tst *TemplateSafetyTester) GetViolations() []SafetyViolation {
	// Return a copy to prevent external modification
	violations := make([]SafetyViolation, len(tst.violations))
	copy(violations, tst.violations)
	return violations
}

// GetMutations returns a copy of all mutation attempt descriptions.
//
// This method provides access to the list of mutation descriptions
// that were recorded via AttemptMutation().
//
// Returns:
//   - []string: A copy of all mutation descriptions
//
// Example:
//
//	mutations := tester.GetMutations()
//	for _, m := range mutations {
//	    fmt.Printf("Mutation: %s\n", m)
//	}
func (tst *TemplateSafetyTester) GetMutations() []string {
	// Return a copy to prevent external modification
	mutations := make([]string, len(tst.mutations))
	copy(mutations, tst.mutations)
	return mutations
}

// IsImmutable returns whether the template has remained immutable.
//
// Returns true if no mutations have been attempted, false otherwise.
//
// Returns:
//   - bool: True if immutable, false if mutations detected
//
// Example:
//
//	if tester.IsImmutable() {
//	    fmt.Println("Template is pure!")
//	} else {
//	    fmt.Println("Mutations detected!")
//	}
func (tst *TemplateSafetyTester) IsImmutable() bool {
	return tst.immutable
}

// GetTemplate returns the template string being tested.
//
// Returns:
//   - string: The template string
//
// Example:
//
//	template := tester.GetTemplate()
//	fmt.Printf("Testing template: %s\n", template)
func (tst *TemplateSafetyTester) GetTemplate() string {
	return tst.template
}

// Reset clears all mutations and violations, resetting to immutable state.
//
// This is useful for reusing the same tester instance across multiple test cases.
//
// Example:
//
//	tester := testutil.NewTemplateSafetyTester("template")
//	tester.AttemptMutation("mutation1")
//	tester.Reset()
//
//	assert.True(t, tester.IsImmutable())
//	assert.Empty(t, tester.GetViolations())
func (tst *TemplateSafetyTester) Reset() {
	tst.mutations = []string{}
	tst.violations = []SafetyViolation{}
	tst.immutable = true
}

// String returns a string representation of the tester state for debugging.
//
// Returns:
//   - string: A formatted string showing template, mutation count, and immutability
//
// Example:
//
//	fmt.Println(tester.String())
//	// Output: TemplateSafetyTester{template="Count: {{count}}", mutations=2, immutable=false}
func (tst *TemplateSafetyTester) String() string {
	return fmt.Sprintf("TemplateSafetyTester{template=%q, mutations=%d, immutable=%v}",
		tst.template, len(tst.mutations), tst.immutable)
}
