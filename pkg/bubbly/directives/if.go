package directives

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// IfDirective implements conditional rendering with ElseIf and Else support.
//
// The If directive provides a declarative way to conditionally render content
// based on boolean conditions. It supports chaining multiple conditions via
// ElseIf and provides a fallback via Else.
//
// # Basic Usage
//
//	If(condition, func() string {
//	    return "Condition is true"
//	}).Render()
//
// # With Else
//
//	If(condition, func() string {
//	    return "True branch"
//	}).Else(func() string {
//	    return "False branch"
//	}).Render()
//
// # With ElseIf Chain
//
//	If(status == "loading",
//	    func() string { return "Loading..." },
//	).ElseIf(status == "error",
//	    func() string { return "Error occurred" },
//	).ElseIf(status == "empty",
//	    func() string { return "No data" },
//	).Else(func() string {
//	    return "Data loaded"
//	}).Render()
//
// # Nested If
//
//	If(outerCondition, func() string {
//	    return If(innerCondition, func() string {
//	        return "Both true"
//	    }).Else(func() string {
//	        return "Outer true, inner false"
//	    }).Render()
//	}).Render()
//
// # Type Safety
//
// All branch functions must return strings. The directive is type-safe and
// will catch type mismatches at compile time.
//
// # Performance
//
// The directive evaluates conditions lazily - only the branch that matches
// is executed. This makes it efficient even with expensive render functions.
//
// # Purity
//
// The directive is pure - it has no side effects and always produces the same
// output for the same input. Branch functions should also be pure for predictable
// behavior.
type IfDirective struct {
	condition      bool
	thenBranch     func() string
	elseIfBranches []ElseIfBranch
	elseBranch     func() string
}

// ElseIfBranch represents a single ElseIf condition and its associated branch.
//
// This type is used internally by IfDirective to store chained ElseIf conditions.
// Each ElseIfBranch contains a boolean condition and a function to execute if
// that condition is true (and all previous conditions were false).
type ElseIfBranch struct {
	condition bool
	branch    func() string
}

// If creates a new conditional directive with the given condition and then branch.
//
// The If function is the entry point for conditional rendering. It evaluates the
// condition and, if true, executes the then function. If false, it checks any
// ElseIf branches or the Else branch.
//
// Parameters:
//   - condition: Boolean expression to evaluate
//   - then: Function to execute if condition is true
//
// Returns:
//   - *IfDirective: A new If directive that can be chained with ElseIf/Else
//
// Example:
//
//	If(user.IsAdmin(), func() string {
//	    return "Admin Panel"
//	}).Else(func() string {
//	    return "User Panel"
//	}).Render()
//
// The returned directive implements ConditionalDirective, allowing method chaining
// for ElseIf and Else branches.
func If(condition bool, then func() string) *IfDirective {
	return &IfDirective{
		condition:      condition,
		thenBranch:     then,
		elseIfBranches: []ElseIfBranch{},
		elseBranch:     nil,
	}
}

// ElseIf adds an additional conditional branch to the directive chain.
//
// This method allows chaining multiple conditions, where each condition is
// evaluated in order until one is true. If this ElseIf's condition is true
// and all previous conditions were false, the provided then function is executed.
//
// Parameters:
//   - condition: Boolean expression to evaluate
//   - then: Function to execute if condition is true and all previous conditions were false
//
// Returns:
//   - ConditionalDirective: Self reference for method chaining
//
// Example:
//
//	If(score >= 90, func() string { return "A" }).
//	    ElseIf(score >= 80, func() string { return "B" }).
//	    ElseIf(score >= 70, func() string { return "C" }).
//	    Else(func() string { return "F" }).
//	    Render()
//
// ElseIf branches are evaluated in the order they are added. The first matching
// condition's branch is executed, and subsequent branches are skipped.
func (d *IfDirective) ElseIf(condition bool, then func() string) ConditionalDirective {
	d.elseIfBranches = append(d.elseIfBranches, ElseIfBranch{
		condition: condition,
		branch:    then,
	})
	return d
}

// Else provides a fallback branch when all previous conditions are false.
//
// This method completes the conditional chain by providing a default branch
// that executes when neither the initial If condition nor any ElseIf conditions
// are true. Only one Else can be specified per conditional chain.
//
// Parameters:
//   - then: Function to execute if all previous conditions were false
//
// Returns:
//   - ConditionalDirective: Self reference for method chaining (allows Render())
//
// Example:
//
//	If(hasData, func() string {
//	    return renderData()
//	}).Else(func() string {
//	    return "No data available"
//	}).Render()
//
// If Else is not called and all conditions are false, Render() returns an empty string.
func (d *IfDirective) Else(then func() string) ConditionalDirective {
	d.elseBranch = then
	return d
}

// Render executes the directive logic and returns the resulting string output.
//
// This method evaluates the conditional chain in order:
//  1. If the main condition is true, execute the then branch
//  2. Otherwise, check each ElseIf condition in order
//  3. If an ElseIf condition is true, execute its branch
//  4. If all conditions are false, execute the Else branch (if present)
//  5. If no Else branch and all conditions false, return empty string
//
// Returns:
//   - string: The rendered output from the first matching branch, or empty string
//
// Example:
//
//	result := If(status == "loading",
//	    func() string { return "Loading..." },
//	).ElseIf(status == "error",
//	    func() string { return "Error!" },
//	).Else(func() string {
//	    return "Ready"
//	}).Render()
//
// The method is pure and idempotent - calling it multiple times with the same
// state produces the same result. Only the matching branch function is executed,
// making it efficient even with expensive render functions.
//
// # Error Handling
//
// If any branch function panics, the panic is recovered and reported to the
// observability system. The directive returns an empty string, allowing the
// application to continue running. This follows the ZERO TOLERANCE policy for
// silent error handling - all panics are reported with full context including
// stack traces.
func (d *IfDirective) Render() string {
	// Check main condition
	if d.condition {
		return d.safeExecute(d.thenBranch, "then")
	}

	// Check ElseIf branches in order
	for i, branch := range d.elseIfBranches {
		if branch.condition {
			return d.safeExecute(branch.branch, fmt.Sprintf("elseif[%d]", i))
		}
	}

	// Execute Else branch if present
	if d.elseBranch != nil {
		return d.safeExecute(d.elseBranch, "else")
	}

	// No conditions met and no Else branch
	return ""
}

// safeExecute wraps a branch function execution with panic recovery.
//
// This method ensures that panics in user-provided render functions don't crash
// the application. Instead, panics are recovered, reported to the observability
// system with full context, and an empty string is returned for graceful degradation.
//
// Parameters:
//   - fn: The branch function to execute
//   - branchName: Name of the branch for error reporting (e.g., "then", "else", "elseif[0]")
//
// Returns:
//   - string: The result of fn() if successful, or empty string if fn panics
//
// This follows the ZERO TOLERANCE policy for silent error handling. All panics
// are reported to the observability system with:
//   - Directive type ("If")
//   - Branch name (which branch panicked)
//   - Panic value (what was passed to panic())
//   - Stack trace (where the panic occurred)
//   - Timestamp (when the panic occurred)
func (d *IfDirective) safeExecute(fn func() string, branchName string) string {
	defer func() {
		if r := recover(); r != nil {
			// Report panic to observability system
			if reporter := observability.GetErrorReporter(); reporter != nil {
				// Create error with context
				err := fmt.Errorf("%w: If directive %s branch panicked: %v", ErrRenderPanic, branchName, r)

				ctx := &observability.ErrorContext{
					ComponentName: "If",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"directive_type": "If",
						"branch_name":    branchName,
						"error_type":     "render_panic",
					},
					Extra: map[string]interface{}{
						"panic_value": r,
						"branch":      branchName,
					},
				}

				reporter.ReportError(err, ctx)
			}
		}
	}()

	return fn()
}
